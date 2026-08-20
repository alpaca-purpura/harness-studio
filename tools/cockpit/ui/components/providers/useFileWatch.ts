'use client';

/**
 * useFileWatch — hook React que se conecta al endpoint SSE /api/watch
 * y dispara `onEvent` cuando un archivo del SSoT sistema cambia.
 *
 * Convenciones:
 *   - 1 EventSource por mount · cleanup en unmount
 *   - Debounce 200ms para evitar storms (cuando un skill escribe múltiples
 *     archivos en secuencia, se agrupan en un solo flush)
 *   - EventSource auto-reconnects al perder conexión · solo logueamos
 *   - Heartbeats se descartan (solo para keepalive servidor)
 *   - Eventos `ready`/`error` se exponen vía `status` para UI indicator
 *
 * Uso:
 * ```tsx
 * useFileWatch(sistema, (e) => {
 *   if (e.docType === 'checkpoint') refetchStories();
 * });
 * ```
 */

import { useEffect, useRef, useState } from 'react';
import type { WatcherEvent } from '@/lib/chokidar-watcher';

export type WatchStatus = 'connecting' | 'open' | 'error' | 'closed';

export interface FileWatchOptions {
  /** ms de debounce antes de flush al callback. Default 200ms. */
  debounceMs?: number;
}

interface FileWatchState {
  status: WatchStatus;
  /** Timestamp (epoch ms) del último evento útil recibido. null si nunca. */
  lastEventAt: number | null;
}

/**
 * Suscribe a SSE /api/watch?sistema=X y llama `onEvent` por cada cambio.
 * Devuelve estado para indicar status visual en UI.
 */
export function useFileWatch(
  sistema: string,
  onEvent: (event: WatcherEvent) => void,
  options: FileWatchOptions = {}
): FileWatchState {
  const { debounceMs = 200 } = options;
  const [status, setStatus] = useState<WatchStatus>('connecting');
  const [lastEventAt, setLastEventAt] = useState<number | null>(null);

  // Ref al callback para no re-suscribir cuando cambia (estable cross-render)
  const callbackRef = useRef(onEvent);
  useEffect(() => {
    callbackRef.current = onEvent;
  }, [onEvent]);

  useEffect(() => {
    if (!sistema) return;

    let pending: WatcherEvent[] = [];
    let timer: ReturnType<typeof setTimeout> | null = null;

    const url = `/api/watch?sistema=${encodeURIComponent(sistema)}`;
    const es = new EventSource(url);

    setStatus('connecting');

    es.onopen = () => {
      setStatus('open');
    };

    es.onerror = () => {
      // EventSource intenta reconectar automáticamente · solo marcamos error
      setStatus('error');
      // eslint-disable-next-line no-console
      console.warn('[useFileWatch] SSE error · reconectando…');
    };

    es.onmessage = (msg) => {
      let data: unknown;
      try {
        data = JSON.parse(msg.data);
      } catch {
        return;
      }
      if (!data || typeof data !== 'object') return;

      const payload = data as {
        type?: string;
        path?: string;
        action?: string;
        sistema?: string;
        docType?: string;
        ts?: number;
        message?: string;
      };

      // Ignoramos heartbeats y ready · solo loggeamos errors
      if (payload.type === 'heartbeat') return;
      if (payload.type === 'ready') {
        setStatus('open');
        return;
      }
      if (payload.type === 'error') {
        setStatus('error');
        // eslint-disable-next-line no-console
        console.warn('[useFileWatch] server error:', payload.message);
        return;
      }
      if (payload.type !== 'change' || !payload.path) return;

      setLastEventAt(payload.ts ?? Date.now());

      pending.push({
        path: payload.path,
        action: (payload.action as WatcherEvent['action']) ?? 'change',
        sistema: payload.sistema,
        docType: payload.docType as WatcherEvent['docType'],
      });

      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        const batch = pending;
        pending = [];
        timer = null;
        for (const ev of batch) {
          try {
            callbackRef.current(ev);
          } catch (err) {
            // eslint-disable-next-line no-console
            console.error('[useFileWatch] callback error:', err);
          }
        }
      }, debounceMs);
    };

    return () => {
      if (timer) clearTimeout(timer);
      pending = [];
      es.close();
      setStatus('closed');
    };
  }, [sistema, debounceMs]);

  return { status, lastEventAt };
}
