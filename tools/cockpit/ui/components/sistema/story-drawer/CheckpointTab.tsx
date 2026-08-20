'use client';

import { useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { Pencil, RotateCcw, ExternalLink } from 'lucide-react';
import type { StoryWithArchive } from '@/lib/api-client';
import { openInEditor } from '@/lib/api-client';
import { Button } from '@/components/ui/Button';
import { StateBadge } from '@/components/ui/Badge';
import { Card } from '@/components/ui/Card';
import { Tooltip } from '@/components/ui/Tooltip';
import { FIELD_TOOLTIPS } from '@/lib/field-tooltips';
import { YamlEditor } from '@/components/ui/YamlEditor';
import { MarkdownView } from './MarkdownView';
import { storyArtifactRel } from '@/lib/story-paths';

interface CheckpointTabProps {
  story: StoryWithArchive;
  onUpdated: () => void;
}

/**
 * Campos que se renderizan con formato "rico" (Badge, mono, fmt especial).
 * Cualquier otro key del frontmatter se renderiza genérico (label/value).
 */
const RICH_FIELDS = new Set([
  'story_id',
  'sistema',
  'state',
  'release',
  'cap_target',
  'cap_change_type',
  'parent_story',
  'owner',
  'type',
  'module',
  'priority',
]);

/** Campos que se muestran en el header del drawer o son ruido. Se ocultan acá. */
const SKIP_FIELDS = new Set([
  'path',
  'is_archived',
  'body',
  'next_action', // ya lo mostramos abajo con énfasis
  'goal',
  'anti',
  'reuse',
  'parked_reason',
  'dropped_reason',
]);

export function CheckpointTab({ story, onUpdated }: CheckpointTabProps) {
  const [editingRaw, setEditingRaw] = useState(false);

  const checkpointRel = useMemo(
    () => storyArtifactRel(story.path, 'checkpoint.md', story.sistema),
    [story.path, story.sistema]
  );

  async function handleOpenExternal() {
    if (!checkpointRel) return;
    try {
      const r = await openInEditor(checkpointRel);
      toast.success(`abierto en ${r?.editor ?? 'editor externo'}`);
    } catch (err) {
      const e = err as Error & { detail?: { hint?: string } };
      toast.error(e.detail?.hint ?? `error: ${e.message}`);
    }
  }

  const isDone = story.state === 'done';

  // Frontmatter completo: iterar TODOS los keys del story object.
  // Separar entre RICH (renderizado en cards específicos arriba) y GENERIC.
  const allEntries = Object.entries(story).filter(
    ([k, v]) => !SKIP_FIELDS.has(k) && v !== null && v !== undefined && v !== ''
  );
  const richEntries = allEntries.filter(([k]) => RICH_FIELDS.has(k));
  const genericEntries = allEntries.filter(([k]) => !RICH_FIELDS.has(k));

  return (
    <div className="space-y-4">
      {/* Estado actual + frontmatter rich */}
      <Card>
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold">Estado actual</h3>
          <div className="flex items-center gap-2">
            <StateBadge state={story.state} />
            {checkpointRel && (
              <Button
                size="sm"
                variant="ghost"
                onClick={handleOpenExternal}
                title="Abrir checkpoint.md en editor externo"
              >
                <ExternalLink className="w-3 h-3" />
              </Button>
            )}
          </div>
        </div>

        <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-xs">
          {richEntries.map(([k, v]) => (
            <FmField
              key={k}
              fieldKey={k}
              label={prettyLabel(k)}
              value={formatValue(v)}
              mono={['story_id', 'cap_target', 'parent_story'].includes(k)}
            />
          ))}
        </dl>

        {story.next_action && (
          <div className="mt-3 pt-3 border-t border-[var(--color-border)] text-xs">
            <div className="text-[var(--color-muted)] mb-1">Next action:</div>
            <div className="font-mono text-[11px] bg-[var(--color-panel)] px-2 py-1 rounded">
              {story.next_action}
            </div>
          </div>
        )}

        {story.goal && (
          <div className="mt-3 pt-3 border-t border-[var(--color-border)] text-xs">
            <div className="text-[var(--color-muted)] mb-1">Goal:</div>
            <div className="whitespace-pre-wrap">{story.goal}</div>
          </div>
        )}
        {story.anti && (
          <div className="mt-2 text-xs">
            <div className="text-[var(--color-muted)] mb-1">Anti-goal:</div>
            <div className="whitespace-pre-wrap italic">{story.anti}</div>
          </div>
        )}
        {story.reuse && (
          <div className="mt-2 text-xs">
            <div className="text-[var(--color-muted)] mb-1">Reuse:</div>
            <div className="whitespace-pre-wrap italic">{story.reuse}</div>
          </div>
        )}

        {story.parked_reason && (
          <div className="mt-3 pt-3 border-t border-[var(--color-border)] text-xs">
            <div className="text-[var(--color-muted)] mb-1">Razón parked:</div>
            <div className="italic">{story.parked_reason}</div>
          </div>
        )}
        {story.dropped_reason && (
          <div className="mt-3 pt-3 border-t border-[var(--color-border)] text-xs">
            <div className="text-[var(--color-muted)] mb-1">Razón dropped:</div>
            <div className="italic">{story.dropped_reason}</div>
          </div>
        )}
      </Card>

      {/* Frontmatter completo · campos genéricos (todo lo demás del YAML) */}
      {genericEntries.length > 0 && (
        <Card>
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-semibold">Metadata adicional</h3>
            <span className="text-[10px] text-[var(--color-muted)]">
              {genericEntries.length} campos · todo el frontmatter visible
            </span>
          </div>
          <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1.5 text-[11px]">
            {genericEntries.map(([k, v]) => (
              <FmField key={k} fieldKey={k} label={prettyLabel(k)} value={formatValue(v)} />
            ))}
          </dl>
        </Card>
      )}

      {/* Editor raw del checkpoint.md completo (frontmatter YAML + body MD) */}
      {checkpointRel && (
        <Card>
          <div className="flex items-center justify-between mb-3">
            <div>
              <h3 className="text-sm font-semibold">checkpoint.md raw</h3>
              <p className="text-[10px] text-[var(--color-muted)] mt-0.5">
                Frontmatter YAML + body markdown · siempre editable (escape hatch
                principal para campos no-estándar).
              </p>
            </div>
            <Button
              size="sm"
              variant={editingRaw ? 'default' : 'primary'}
              onClick={() => setEditingRaw((s) => !s)}
            >
              {editingRaw ? (
                <>
                  <RotateCcw className="w-3 h-3" /> Cerrar editor
                </>
              ) : (
                <>
                  <Pencil className="w-3 h-3" /> Editar raw
                </>
              )}
            </Button>
          </div>
          {editingRaw ? (
            <YamlEditor
              relPath={checkpointRel}
              storyState={story.state}
              isArchived={story.is_archived ?? false}
              kindOverride="checkpoint"
              embedded
              onSaved={() => {
                setEditingRaw(false);
                onUpdated();
              }}
            />
          ) : (
            <MarkdownView
              relPath={checkpointRel}
              storyState={story.state}
              isArchived={story.is_archived ?? false}
              kindOverride="checkpoint"
            />
          )}
        </Card>
      )}

      {/* Banner done */}
      {isDone && (
        <div className="px-3 py-2 bg-[#14532d]/30 border border-[#14532d] rounded text-xs text-[#86efac]">
          ✓ Esta story está cementada. Para extender el resultado, crea una story
          hija basada en esta (from-done) desde la pestaña Técnico → Files.
        </div>
      )}

    </div>
  );
}

function FmField({
  label,
  value,
  mono = false,
  fieldKey,
}: {
  label: string;
  value: string;
  mono?: boolean;
  fieldKey?: string;
}) {
  const tip = fieldKey ? FIELD_TOOLTIPS[fieldKey] : undefined;
  return (
    <>
      <dt className="text-[var(--color-muted)]">
        {tip ? (
          <Tooltip content={tip} variant="badge">
            <span className="cursor-help border-b border-dotted border-[var(--color-border)]">{label}</span>
          </Tooltip>
        ) : (
          label
        )}
      </dt>
      <dd className={mono ? 'font-mono text-[11px] break-all' : 'break-words'}>
        {value || <span className="text-[var(--color-muted)]">—</span>}
      </dd>
    </>
  );
}

function prettyLabel(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatValue(v: unknown): string {
  if (v === null || v === undefined) return '—';
  if (typeof v === 'boolean') return v ? '✓ true' : '✗ false';
  if (typeof v === 'number') return String(v);
  if (typeof v === 'string') return v;
  if (Array.isArray(v)) return v.map(String).join(', ');
  try {
    return JSON.stringify(v, null, 2);
  } catch {
    return String(v);
  }
}
