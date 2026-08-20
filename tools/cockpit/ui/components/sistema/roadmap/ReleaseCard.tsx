'use client';

import { useDroppable } from '@dnd-kit/core';
import { Rocket, Pencil, Cloud } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Button } from '@/components/ui/Button';
import { ReleaseStatusBadge } from '@/components/ui/Badge';
import { EmptyState } from '@/components/ui/Spinner';
import { Tooltip } from '@/components/ui/Tooltip';
import { StoryChip } from './StoryChip';
import type { Release } from '@/lib/types';
import type { StoryWithArchive } from '@/lib/api-client';

interface ReleaseCardProps {
  release: Release;
  stories: StoryWithArchive[];
  onMergeRequested: (releaseId: string) => void;
  onEditRequested: (release: Release) => void;
}

const TERMINAL_STATES = new Set(['done', 'dropped']);

export function ReleaseCard({
  release,
  stories,
  onMergeRequested,
  onEditRequested,
}: ReleaseCardProps) {
  const isShipped = release.status === 'shipped';
  const { setNodeRef, isOver } = useDroppable({
    id: `release:${release.release_id}`,
    data: { releaseId: release.release_id },
  });

  const allTerminal =
    stories.length > 0 && stories.every((s) => TERMINAL_STATES.has(s.state));
  const ready =
    release.status === 'ready_to_merge' ||
    (release.status === 'in_progress' && allTerminal);

  // Counts por estado para badge
  const counts = stories.reduce<Record<string, number>>((acc, s) => {
    acc[s.state] = (acc[s.state] ?? 0) + 1;
    return acc;
  }, {});

  // Progreso del release (PM glance): % entregado.
  const total = stories.length;
  const doneCount = stories.filter((s) => s.state === 'done').length;
  const pct = total ? Math.round((doneCount / total) * 100) : 0;

  // Orden interno (PM repriorización): activas arriba, planeación, cerradas al fondo;
  // dentro de cada fase, prioridad crítica primero.
  const PHASE: Record<string, number> = {
    ready: 0, developing: 0, reviewing: 0, developed: 0,
    refined: 1, refining: 1, idea: 2, parked: 3, done: 4, dropped: 5,
  };
  const PRIO: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
  const sorted = [...stories].sort(
    (a, b) =>
      (PHASE[a.state] ?? 9) - (PHASE[b.state] ?? 9) ||
      (PRIO[(a.priority ?? '').toLowerCase()] ?? 9) -
        (PRIO[(b.priority ?? '').toLowerCase()] ?? 9)
  );

  return (
    <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg">
      <header className="flex items-start justify-between gap-3 px-4 py-3 border-b border-[var(--color-border)]">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-mono text-sm font-semibold">
              {release.release_id}
            </span>
            <ReleaseStatusBadge status={release.status} />
            <span className="text-[11px] text-[var(--color-muted)]">
              {stories.length} stories
            </span>
          </div>
          <div className="text-xs">{release.name}</div>
          {release.description && (
            <div className="text-[11px] text-[var(--color-muted)] mt-0.5 line-clamp-2">
              {release.description}
            </div>
          )}
          {/* Progreso entregado (PM glance) */}
          {total > 0 && (
            <div className="flex items-center gap-2 mt-1.5 max-w-xs">
              <div className="h-1.5 flex-1 rounded-full bg-[var(--color-panel2)] overflow-hidden">
                <div
                  className="h-full rounded-full bg-emerald-500 transition-all"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="text-[10px] text-[var(--color-muted)] font-mono shrink-0">
                {doneCount}/{total} · {pct}%
              </span>
            </div>
          )}
          {release.target_date && (
            <div className="text-[10px] text-[var(--color-muted)] mt-1 font-mono">
              🗓 target: {release.target_date}
            </div>
          )}
          {/* Sello de verificación (eje integración) en releases shipped */}
          {isShipped && release.verified_by && (
            <div className="text-[10px] text-emerald-400/90 mt-1">
              ✓ verificado por {release.verified_by}
              {release.verified_at ? ` · ${release.verified_at.substring(0, 10)}` : ''}
              {release.verification_note ? (
                <span className="text-[var(--color-muted)]"> · {release.verification_note}</span>
              ) : null}
            </div>
          )}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {/* Editar nombre/descripción · solo releases por venir (shipped es inmutable) */}
          {!isShipped && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onEditRequested(release)}
              title="Editar nombre / descripción"
            >
              <Pencil className="w-3 h-3" />
            </Button>
          )}
          {ready && !isShipped && (
            <Button
              size="sm"
              variant="primary"
              onClick={() => onMergeRequested(release.release_id)}
            >
              <Rocket className="w-3 h-3" />
              Cerrar → shipped
            </Button>
          )}
          {/* Pase a producción · FUTURO (eje despliegue). Placeholder deshabilitado. */}
          {isShipped && (
            <Tooltip content="Próximamente: merge a release/{sistema}-vX.Y.Z → deploy a producción (GH Actions). Aún no implementado.">
              <span>
                <Button size="sm" variant="ghost" disabled className="opacity-60">
                  <Cloud className="w-3 h-3" />
                  Pase a producción
                </Button>
              </span>
            </Tooltip>
          )}
        </div>
      </header>

      <div
        ref={setNodeRef}
        className={cn(
          'p-3 min-h-[100px] transition-colors rounded-b-lg',
          isOver && 'bg-[var(--color-accent)]/10 border-2 border-dashed border-[var(--color-accent)]/50'
        )}
      >
        {stories.length === 0 ? (
          <EmptyState>Sin stories asignadas todavía.</EmptyState>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
            {sorted.map((s) => (
              <StoryChip key={s.story_id} story={s} />
            ))}
          </div>
        )}

        {/* Mini state counts */}
        {Object.keys(counts).length > 0 && (
          <div className="mt-3 pt-2 border-t border-[var(--color-border)] flex flex-wrap gap-1.5">
            {Object.entries(counts).map(([state, count]) => (
              <span
                key={state}
                className="text-[10px] px-1.5 py-0.5 bg-[var(--color-panel2)] border border-[var(--color-border)] rounded font-mono"
              >
                {state} · {count}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
