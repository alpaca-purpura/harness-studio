'use client';

import { useDroppable } from '@dnd-kit/core';
import { cn } from '@/lib/cn';
import { StateBadge } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { BoardCard } from './BoardCard';
import type { ActiveSession, StoryState } from '@/lib/types';
import type { StoryWithArchive } from '@/lib/api-client';
import { useProceso } from '@/components/providers/ProcesoProvider';
import { estadoDe, isOperatorAllowed } from '@/lib/proceso';

// F6/RN-51: tooltip de columna = la descripción del estado EN el descriptor;
// feedback de drop = la whitelist de operador derivada. Cero literales del ciclo.

interface BoardColumnProps {
  state: StoryState;
  stories: StoryWithArchive[];
  /** Estado origen que se está dragging (para feedback drop allowed/forbidden) */
  draggingFromState: StoryState | null;
  wipCap?: number;
  /** Build-claims vivos por story_id (ADR-009) → badge 🔨 en la card. */
  sessions?: Record<string, ActiveSession>;
}

export function BoardColumn({
  state,
  stories,
  draggingFromState,
  wipCap,
  sessions,
}: BoardColumnProps) {
  const proceso = useProceso();
  const { setNodeRef, isOver } = useDroppable({
    id: `state:${state}`,
    data: { state },
  });

  let dropFeedback: 'allowed' | 'forbidden' | null = null;
  if (draggingFromState && draggingFromState !== state) {
    const allowed =
      proceso !== null && isOperatorAllowed(proceso, draggingFromState, state) !== null;
    dropFeedback = allowed ? 'allowed' : 'forbidden';
  }

  // `done`: agrupar por release (release nuevo arriba) y dentro por entrega reciente
  // (last_modified desc). Sin release → grupo al fondo. Resto de columnas: orden natural.
  const releaseNum = (r: string | null | undefined): number => {
    const m = r?.match(/(\d+)/);
    return m ? Number.parseInt(m[1], 10) : -1;
  };
  const doneGroups =
    state === 'done'
      ? Object.values(
          stories.reduce<
            Record<string, { release: string | null; items: StoryWithArchive[] }>
          >((acc, s) => {
            const key = s.release ?? '∅';
            (acc[key] ??= { release: s.release ?? null, items: [] }).items.push(s);
            return acc;
          }, {})
        )
          .map((g) => ({
            release: g.release,
            items: [...g.items].sort((a, b) =>
              (b.last_modified ?? '').localeCompare(a.last_modified ?? '')
            ),
          }))
          .sort((a, b) => releaseNum(b.release) - releaseNum(a.release))
      : null;

  return (
    <div className="w-48 shrink-0 flex flex-col">
      <header className="flex items-center justify-between mb-2 px-1">
        <Tooltip
          content={(proceso && estadoDe(proceso, state)?.descripcion) || state}
          variant="badge"
        >
          <StateBadge state={state} />
        </Tooltip>
        <div className="text-[10px] text-[var(--color-muted)] font-mono">
          {stories.length}
          {wipCap !== undefined && (
            <Tooltip content={TOOLTIPS.wip_cap}>
              <span className={stories.length > wipCap ? 'text-orange-400' : ''}>
                /{wipCap}
              </span>
            </Tooltip>
          )}
        </div>
      </header>
      <div
        ref={setNodeRef}
        className={cn(
          'flex-1 min-h-[400px] p-1.5 rounded border transition-colors',
          'bg-[var(--color-panel)] border-[var(--color-border)]',
          isOver && dropFeedback === 'allowed' && 'bg-[var(--color-accent)]/10 border-dashed border-[var(--color-accent)]',
          isOver && dropFeedback === 'forbidden' && 'bg-red-900/10 border-dashed border-red-700'
        )}
      >
        {stories.length === 0 ? (
          <div className="text-[10px] text-[var(--color-muted)] italic text-center mt-4">
            Sin stories.
          </div>
        ) : doneGroups ? (
          <div className="space-y-3">
            {doneGroups.map((g) => (
              <div key={g.release ?? '∅'} className="space-y-1.5">
                <div className="flex items-center gap-1.5 px-0.5">
                  <span className="text-[10px] font-bold text-[var(--color-text)]">
                    {g.release ?? 'sin release'}
                  </span>
                  <span className="h-px flex-1 bg-[var(--color-border)]" />
                  <span className="text-[9px] text-[var(--color-muted)] font-mono">
                    {g.items.length}
                  </span>
                </div>
                {g.items.map((s) => (
                  <BoardCard
                    key={s.story_id}
                    story={s}
                    session={sessions?.[s.story_id]}
                  />
                ))}
              </div>
            ))}
          </div>
        ) : (
          <div className="space-y-1.5">
            {stories.map((s) => (
              <BoardCard
                key={s.story_id}
                story={s}
                session={sessions?.[s.story_id]}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
