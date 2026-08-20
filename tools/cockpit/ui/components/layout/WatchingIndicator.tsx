'use client';

/**
 * WatchingIndicator — dot pulsante que muestra status del watcher SSE.
 *
 *   verde pulsante → conexión activa, watcher escuchando
 *   amarillo       → reconectando
 *   rojo           → error / disconnected
 *
 * Tooltip muestra timestamp del último evento útil.
 */

import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchStatus } from '@/components/providers/FileWatchProvider';
import { cn } from '@/lib/cn';

function formatRelative(ts: number | null): string {
  if (ts === null) return 'sin eventos aún';
  const delta = Date.now() - ts;
  if (delta < 1000) return 'ahora mismo';
  if (delta < 60_000) return `hace ${Math.floor(delta / 1000)}s`;
  if (delta < 3_600_000) return `hace ${Math.floor(delta / 60_000)}m`;
  return new Date(ts).toLocaleTimeString();
}

export function WatchingIndicator() {
  const { sistema } = useSistema();
  const { status, lastEventAt } = useFileWatchStatus();

  const dotClass = (() => {
    switch (status) {
      case 'open':
        return 'bg-green-400 animate-pulse';
      case 'connecting':
        return 'bg-yellow-400 animate-pulse';
      case 'error':
        return 'bg-red-500';
      case 'closed':
        return 'bg-zinc-500';
      default:
        return 'bg-zinc-500';
    }
  })();

  const label = (() => {
    switch (status) {
      case 'open':
        return `watching ${sistema}`;
      case 'connecting':
        return 'conectando…';
      case 'error':
        return 'desconectado';
      case 'closed':
        return 'cerrado';
      default:
        return '—';
    }
  })();

  const title = `Estado watcher: ${status} · último evento: ${formatRelative(
    lastEventAt
  )}`;

  return (
    <span className="flex items-center gap-1.5" title={title}>
      <span
        aria-hidden="true"
        className={cn('w-2 h-2 rounded-full inline-block', dotClass)}
      />
      <span>{label}</span>
    </span>
  );
}
