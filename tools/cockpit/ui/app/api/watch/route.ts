/**
 * GET /api/watch?sistema={sistema}
 *
 * Server-Sent Events stream que push notifications cuando un archivo del SSoT
 * cambia. Setup chokidar watcher via lib/chokidar-watcher.ts y emite cada
 * WatcherEvent como mensaje SSE.
 *
 * Mensaje formato:
 *   data: {"path":"...","action":"change","sistema":"{sistema}","docType":"checkpoint"}\n\n
 *
 * Heartbeat cada 30s:
 *   data: {"type":"heartbeat","ts":1716800000000}\n\n
 *
 * Cleanup automático al disconnect (req.signal AbortEvent).
 *
 * NOTA: este endpoint mantiene la conexión abierta · cada cliente abre 1
 * watcher chokidar dedicado. Para multi-cliente sería más eficiente compartir
 * un solo watcher singleton, pero en dev local (1-2 sesiones) es suficiente.
 */

import { NextRequest } from 'next/server';
import { setupWatcher, type WatcherEvent } from '@/lib/chokidar-watcher';
import { getWorkspaceRoot, getSelectableSistemas } from '@/lib/workspace';
import { errorResponse } from '../_lib/responses';

// Force Node.js runtime (chokidar requiere fs nativo · no Edge)
export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';

const HEARTBEAT_MS = 30_000;

export async function GET(req: NextRequest): Promise<Response> {
  const sistemaParam = req.nextUrl.searchParams.get('sistema');

  // Si no se especifica sistema, observamos todos los sistemas activos
  const validSistemas = getSelectableSistemas();
  let sistemasToWatch: string[];

  if (sistemaParam) {
    if (!validSistemas.includes(sistemaParam)) {
      return errorResponse(`sistema desconocido: ${sistemaParam}`, 400, {
        valid_sistemas: validSistemas,
      });
    }
    sistemasToWatch = [sistemaParam];
  } else {
    sistemasToWatch = validSistemas;
  }

  const workspaceRoot = getWorkspaceRoot();
  const encoder = new TextEncoder();

  let watcher: ReturnType<typeof setupWatcher> | null = null;
  let heartbeatId: NodeJS.Timeout | null = null;
  let closed = false;

  const stream = new ReadableStream({
    start(controller) {
      function safeEnqueue(payload: object): void {
        if (closed) return;
        try {
          controller.enqueue(
            encoder.encode(`data: ${JSON.stringify(payload)}\n\n`)
          );
        } catch {
          // Stream cerrado · cleanup
          cleanup();
        }
      }

      function cleanup(): void {
        if (closed) return;
        closed = true;
        if (heartbeatId) {
          clearInterval(heartbeatId);
          heartbeatId = null;
        }
        if (watcher) {
          watcher.close().catch(() => {
            // ignore · cleanup best-effort
          });
          watcher = null;
        }
        try {
          controller.close();
        } catch {
          // ignore double-close
        }
      }

      // Mensaje inicial · cliente confirma conexión activa
      safeEnqueue({
        type: 'ready',
        sistemas: sistemasToWatch,
        ts: Date.now(),
      });

      // Setup chokidar
      try {
        watcher = setupWatcher({
          workspaceRoot,
          sistemas: sistemasToWatch,
        });
      } catch (err) {
        safeEnqueue({
          type: 'error',
          message: (err as Error).message,
        });
        cleanup();
        return;
      }

      watcher.on('event', (event: WatcherEvent) => {
        safeEnqueue({
          type: 'change',
          ...event,
          ts: Date.now(),
        });
      });

      watcher.on('error', (err: unknown) => {
        safeEnqueue({
          type: 'error',
          message: err instanceof Error ? err.message : String(err),
        });
      });

      // Heartbeat keepalive · evita que proxies cierren la conexión idle
      heartbeatId = setInterval(() => {
        safeEnqueue({ type: 'heartbeat', ts: Date.now() });
      }, HEARTBEAT_MS);

      // Cleanup al abort (cliente cierra EventSource)
      req.signal.addEventListener('abort', cleanup);
    },
    cancel() {
      // Stream cancelado · cleanup
      closed = true;
      if (heartbeatId) {
        clearInterval(heartbeatId);
        heartbeatId = null;
      }
      if (watcher) {
        watcher.close().catch(() => {
          // ignore
        });
        watcher = null;
      }
    },
  });

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream; charset=utf-8',
      'Cache-Control': 'no-cache, no-transform',
      Connection: 'keep-alive',
      // Disable Nginx buffering si hay proxy adelante
      'X-Accel-Buffering': 'no',
    },
  });
}
