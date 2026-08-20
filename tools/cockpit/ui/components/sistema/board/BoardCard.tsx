'use client';

import { useDraggable } from '@dnd-kit/core';
import { cn } from '@/lib/cn';
import { Pill } from '@/components/ui/Badge';
import { useDrawer } from '@/components/providers/DrawerProvider';
import type { StoryWithArchive } from '@/lib/api-client';
import type { ActiveSession } from '@/lib/types';
import { agentMetaOf, agentOf, releaseHue, priorityColor, typeMetaOf } from '@/lib/agent-meta';

// Solo idea ↔ refining son draggables (operador-allowed)
const DRAGGABLE_STATES = new Set(['idea', 'refining']);

/** Humaniza un story_id cuando no hay goal: quita prefijos {sistema}-/faseN- y guiones. */
function humanize(storyId: string): string {
  return storyId
    .replace(/^[a-z]+-/, '')
    .replace(/^fase\d+-/, '')
    .replace(/-/g, ' ');
}

export function BoardCard({
  story,
  session,
}: {
  story: StoryWithArchive;
  session?: ActiveSession;
}) {
  const { openStory } = useDrawer();
  const draggable = DRAGGABLE_STATES.has(story.state);

  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: story.story_id,
    disabled: !draggable,
    data: { story },
  });

  const agentId = agentOf(story);
  const agent = agentId ? agentMetaOf(agentId) : null;
  const accent = agent?.color ?? 'var(--color-border)';
  const hue = releaseHue(story.release);
  const prioColor = priorityColor(story.priority);
  const tm = typeMetaOf(story.type);
  const area = story.cap_target?.includes('.') ? story.cap_target.split('.')[1] : null;

  const style: React.CSSProperties = {
    borderLeftColor: accent,
    ...(transform
      ? { transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`, zIndex: 50 }
      : {}),
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...(draggable ? listeners : {})}
      {...attributes}
      onClick={() => {
        if (isDragging) return;
        openStory(story.story_id);
      }}
      className={cn(
        'p-2 pl-2.5 rounded border border-l-[3px] text-xs bg-[var(--color-panel2)] border-[var(--color-border)]',
        'hover:border-[#3a4358] transition-all',
        draggable ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
        isDragging && 'opacity-40'
      )}
    >
      {/* Header: agente (color) + release badge */}
      <div className="flex items-center gap-1.5 mb-1">
        {agent ? (
          <span
            className="inline-flex items-center gap-1 text-[11px] font-semibold truncate"
            style={{ color: accent }}
          >
            <span aria-hidden="true">{agent.emoji}</span>
            {agent.name}
          </span>
        ) : (
          <span className="text-[10px] text-[var(--color-muted)]">—</span>
        )}
        <span className="flex-1" />
        {tm && (
          <span className="text-[11px] shrink-0 leading-none" title={tm.label} aria-hidden="true">
            {tm.icon}
          </span>
        )}
        {story.release && (
          <span
            className="text-[10px] font-bold px-1.5 py-0.5 rounded shrink-0 leading-none"
            style={
              hue !== null
                ? {
                    color: `hsl(${hue} 85% 80%)`,
                    backgroundColor: `hsl(${hue} 55% 20%)`,
                    border: `1px solid hsl(${hue} 50% 40%)`,
                  }
                : undefined
            }
            title={`Release planificado: ${story.release}`}
          >
            {story.release}
          </span>
        )}
      </div>

      {/* Título legible (goal o id humanizado) · story_id solo en tooltip + drawer */}
      <div
        className="text-[11px] leading-snug line-clamp-2 text-[var(--color-text)]"
        title={story.story_id}
      >
        {story.goal ?? humanize(story.story_id)}
      </div>

      {session && (
        <div
          className="mt-1 px-1.5 py-0.5 rounded bg-amber-900/30 border border-amber-700/50 text-amber-300 text-[10px] leading-tight inline-flex items-center gap-1"
          title={`Sesión ${session.lane ?? `pid${session.pid}`} (${session.skill}) construyendo · bucket ${session.bucket}${session.startedAt ? ` · desde ${session.startedAt}` : ''}`}
        >
          🔨 {session.lane ?? `pid${session.pid}`}
        </div>
      )}
      {story.parse_error && (
        <div
          className="mt-1 px-1.5 py-1 rounded bg-red-950/50 border border-red-700 text-red-300 text-[10px] leading-tight"
          title={story.parse_error}
        >
          ⚠ checkpoint inválido — {story.parse_error}
        </div>
      )}
      {story.dup_collision && (
        <div
          className="mt-1 px-1.5 py-0.5 rounded bg-yellow-900/30 border border-yellow-700/50 text-yellow-300 text-[10px] leading-tight"
          title="story_id duplicado: existe en product/stories (live) y en archive (done). Resolver vía /pm-{sistema} (rename o borrar el stub)."
        >
          ⚠ id duplicado (live + archivado)
        </div>
      )}

      {/* Footer: prioridad (punto de color) + surfaces + área */}
      {(story.priority || story.surfaces?.length || area) && (
        <div className="flex items-center gap-1.5 mt-1.5 flex-wrap">
          {story.priority && (
            <span
              className="inline-flex items-center gap-1 text-[9px] text-[var(--color-muted)]"
              title={`Prioridad: ${story.priority}`}
            >
              <span
                className="w-2 h-2 rounded-full shrink-0"
                style={{ backgroundColor: prioColor ?? '#64748b' }}
              />
              {story.priority}
            </span>
          )}
          {story.surfaces?.map((s) => (
            <Pill
              key={s}
              className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[var(--color-muted)] text-[9px]"
            >
              {s}
            </Pill>
          ))}
          {area && (
            <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[var(--color-muted)] text-[9px] opacity-80">
              {area}
            </Pill>
          )}
        </div>
      )}
    </div>
  );
}
