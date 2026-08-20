'use client';

import { useDraggable } from '@dnd-kit/core';
import { cn } from '@/lib/cn';
import { StateBadge } from '@/components/ui/Badge';
import { useDrawer } from '@/components/providers/DrawerProvider';
import type { StoryWithArchive } from '@/lib/api-client';
import { agentMetaOf, agentOf, priorityColor, typeMetaOf } from '@/lib/agent-meta';

const DRAGGABLE_STATES = new Set(['idea', 'refining', 'refined']);

/** Humaniza un story_id cuando no hay goal: quita prefijos {sistema}-/faseN- y guiones. */
function humanize(storyId: string): string {
  return storyId
    .replace(/^[a-z]+-/, '')
    .replace(/^fase\d+-/, '')
    .replace(/-/g, ' ');
}

export function StoryChip({ story }: { story: StoryWithArchive }) {
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
  const prioColor = priorityColor(story.priority);
  const tm = typeMetaOf(story.type);

  const style: React.CSSProperties = {
    borderLeftColor: accent,
    ...(transform ? { transform: `translate3d(${transform.x}px, ${transform.y}px, 0)` } : {}),
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...(draggable ? listeners : {})}
      {...attributes}
      onClick={(e) => {
        if (isDragging) return;
        e.stopPropagation();
        openStory(story.story_id);
      }}
      title={story.story_id}
      className={cn(
        'p-2 pl-2.5 rounded border border-l-[3px] text-xs bg-[var(--color-panel2)] border-[var(--color-border)]',
        'hover:border-[#3a4358] transition-all',
        draggable ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
        isDragging && 'opacity-40'
      )}
    >
      <div className="flex items-center gap-1.5 mb-1">
        <StateBadge state={story.state} />
        {agent && (
          <span
            className="inline-flex items-center gap-0.5 text-[10px] font-semibold truncate"
            style={{ color: accent }}
          >
            <span aria-hidden="true">{agent.emoji}</span>
            {agent.name}
          </span>
        )}
        <span className="flex-1" />
        {tm && (
          <span className="text-[11px] shrink-0 leading-none" title={tm.label} aria-hidden="true">
            {tm.icon}
          </span>
        )}
        {prioColor && (
          <span
            className="w-2 h-2 rounded-full shrink-0"
            style={{ backgroundColor: prioColor }}
            title={`Prioridad: ${story.priority}`}
          />
        )}
      </div>
      <div className="text-[11px] leading-snug line-clamp-2 text-[var(--color-text)]">
        {story.goal ?? humanize(story.story_id)}
      </div>
      {story.dup_collision && (
        <div
          className="mt-1 px-1.5 py-0.5 rounded bg-yellow-900/30 border border-yellow-700/50 text-yellow-300 text-[10px] leading-tight"
          title="story_id duplicado: existe en product/stories (live) y en archive (done). Resolver vía /pm-{sistema} (rename o borrar el stub)."
        >
          ⚠ id duplicado (live + archivado)
        </div>
      )}
    </div>
  );
}
