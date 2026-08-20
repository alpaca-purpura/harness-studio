'use client';

/**
 * FileWatchProvider — provee 1 sola conexión SSE compartida + bus de
 * suscripciones por filtro (`docType`).
 *
 * Por qué: cada vista (Roadmap/Board/Map/Learnings) y el Header del shell
 * necesitan reaccionar a eventos del watcher. Si cada uno abre su propio
 * EventSource, terminamos con N conexiones SSE por sistema · cada una con su
 * propio chokidar watcher en el servidor. Wasteful.
 *
 * Solución: 1 sola EventSource en el provider raíz (montado en AppShell),
 * suscriptores se registran vía `useFileWatchEvents(filter, onEvent)` y el
 * provider los routea según filtros (docType / path matcher).
 *
 * Estado `status` se expone vía `useFileWatchStatus()` para que `Header`
 * dibuje el indicador verde/rojo sin necesitar suscripción a eventos.
 */

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { useSistema } from './SistemaProvider';
import { useFileWatch, type WatchStatus } from './useFileWatch';
import type { WatcherEvent } from '@/lib/chokidar-watcher';

type Subscriber = (event: WatcherEvent) => void;

interface FileWatchContextValue {
  status: WatchStatus;
  lastEventAt: number | null;
  subscribe: (fn: Subscriber) => () => void;
}

const FileWatchContext = createContext<FileWatchContextValue | null>(null);

export function FileWatchProvider({ children }: { children: ReactNode }) {
  const { sistema } = useSistema();
  const subscribersRef = useRef<Set<Subscriber>>(new Set());

  // Re-render cuando cambie status para el indicador (sin disparar
  // re-suscripciones del watcher base)
  const [stateSnapshot, setStateSnapshot] = useState<{
    status: WatchStatus;
    lastEventAt: number | null;
  }>({
    status: 'connecting',
    lastEventAt: null,
  });

  const handleEvent = useCallback((event: WatcherEvent) => {
    for (const sub of subscribersRef.current) {
      try {
        sub(event);
      } catch (err) {
        // eslint-disable-next-line no-console
        console.error('[FileWatchProvider] subscriber error:', err);
      }
    }
  }, []);

  const { status, lastEventAt } = useFileWatch(sistema, handleEvent);

  // Sync snapshot only when status or lastEventAt change · evita re-render storm
  useEffect(() => {
    setStateSnapshot({ status, lastEventAt });
  }, [status, lastEventAt]);

  const subscribe = useCallback((fn: Subscriber) => {
    subscribersRef.current.add(fn);
    return () => {
      subscribersRef.current.delete(fn);
    };
  }, []);

  const value = useMemo<FileWatchContextValue>(
    () => ({
      status: stateSnapshot.status,
      lastEventAt: stateSnapshot.lastEventAt,
      subscribe,
    }),
    [stateSnapshot.status, stateSnapshot.lastEventAt, subscribe]
  );

  return (
    <FileWatchContext.Provider value={value}>
      {children}
    </FileWatchContext.Provider>
  );
}

/** Status + lastEventAt para indicador UI (Header). */
export function useFileWatchStatus(): {
  status: WatchStatus;
  lastEventAt: number | null;
} {
  const ctx = useContext(FileWatchContext);
  if (!ctx) throw new Error('useFileWatchStatus fuera de FileWatchProvider');
  return { status: ctx.status, lastEventAt: ctx.lastEventAt };
}

/**
 * Suscribe `onEvent` a TODOS los eventos del watcher.
 * El consumer filtra por docType / path como necesite.
 *
 * Uso:
 * ```tsx
 * useFileWatchEvents((e) => {
 *   if (e.docType === 'checkpoint') refetchStories();
 * });
 * ```
 */
export function useFileWatchEvents(onEvent: (event: WatcherEvent) => void): void {
  const ctx = useContext(FileWatchContext);
  if (!ctx) throw new Error('useFileWatchEvents fuera de FileWatchProvider');

  // Stable ref para el callback
  const ref = useRef(onEvent);
  useEffect(() => {
    ref.current = onEvent;
  }, [onEvent]);

  useEffect(() => {
    const unsub = ctx.subscribe((e) => ref.current(e));
    return unsub;
  }, [ctx]);
}
