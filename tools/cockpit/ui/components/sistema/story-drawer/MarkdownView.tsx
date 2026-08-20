'use client';

import { useCallback, useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import {
  ExternalLink,
  Lock,
  Pencil,
  RotateCcw,
  Save,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Button } from '@/components/ui/Button';
import { getFile, openInEditor, putFile } from '@/lib/api-client';
import {
  type ArtifactKind,
  getArtifactKind,
  getEditPermission,
} from '@/lib/edit-permissions';
import { useProceso } from '@/components/providers/ProcesoProvider';
import type { StoryState } from '@/lib/types';

// MDEditor usa window/document → dynamic import sin SSR
const MDEditor = dynamic(() => import('@uiw/react-md-editor'), {
  ssr: false,
  loading: () => (
    <div className="text-xs text-[var(--color-muted)] italic p-3">
      Cargando editor…
    </div>
  ),
});

export interface MarkdownViewProps {
  /** Path relativo al workspace root (whitelist /api/file aplica). */
  relPath: string;
  /** Story state actual · para calcular permisos. Si se omite, read-only. */
  storyState?: StoryState;
  /** Si la story está en archive/, todo read-only. */
  isArchived?: boolean;
  /** Override del kind inferido desde relPath (útil para casos especiales). */
  kindOverride?: ArtifactKind;
  /** Mensaje si el archivo no existe (404) */
  missingMessage?: string;
}

/**
 * View + edit inline de un .md/.txt arbitrario del workspace.
 *
 * - View: react-markdown + remark-gfm + rehype-highlight
 * - Edit: @uiw/react-md-editor (lazy) — solo si edit-permissions lo permite
 * - Save: PUT /api/file
 * - Externo: POST /api/open (fallback chain xdg-open/code/xed/etc.)
 *
 * Si la story está en estado que ya consumió el artifact (ej. 01-spec.md cuando
 * state=developing), muestra badge "read-only" + tooltip explicativo + solo
 * permite ver / abrir externo.
 */
export function MarkdownView({
  relPath,
  storyState,
  isArchived = false,
  kindOverride,
  missingMessage,
}: MarkdownViewProps) {
  const [content, setContent] = useState<string | null>(null);
  const [draft, setDraft] = useState<string | null>(null);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [missing, setMissing] = useState(false);

  const kind = kindOverride ?? getArtifactKind(relPath);
  const proceso = useProceso();
  const perm = getEditPermission({
    kind,
    storyState: storyState ?? 'done',
    isArchived,
    proceso,
  });
  // operator-input + capability append-only → NO mostrar edit toggle inline acá
  // (usan UI dedicada en OperatorInputTab / CapDrawer)
  const showEditToggle = perm.editable && !perm.appendOnly;

  const load = useCallback(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setMissing(false);
    getFile(relPath)
      .then((c) => {
        if (cancelled) return;
        setContent(c);
        setDraft(c);
      })
      .catch((err: Error & { status?: number }) => {
        if (cancelled) return;
        if (err.status === 404) setMissing(true);
        else setError(err.message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [relPath]);

  useEffect(() => {
    const cleanup = load();
    return cleanup;
  }, [load]);

  async function handleOpenInEditor() {
    try {
      const res = await openInEditor(relPath);
      toast.success(`abierto en ${res?.editor ?? 'editor externo'}`);
    } catch (err) {
      const e = err as Error & { detail?: { hint?: string } };
      toast.error(e.detail?.hint ?? `error abriendo: ${e.message}`);
    }
  }

  async function handleSave() {
    if (draft === null) return;
    setSaving(true);
    try {
      await putFile(relPath, draft);
      setContent(draft);
      setEditing(false);
      toast.success('Guardado');
    } catch (err) {
      toast.error(`error guardando: ${(err as Error).message}`);
    } finally {
      setSaving(false);
    }
  }

  function handleDiscard() {
    setDraft(content);
    setEditing(false);
  }

  const dirty = editing && draft !== content;

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 justify-between">
        <div className="text-[11px] text-[var(--color-muted)] font-mono truncate">
          {relPath}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {editing ? (
            <>
              <Button
                size="sm"
                variant="ghost"
                onClick={handleDiscard}
                disabled={saving}
              >
                <RotateCcw className="w-3 h-3" /> Descartar
              </Button>
              <Button
                size="sm"
                variant="primary"
                onClick={handleSave}
                disabled={saving || !dirty}
              >
                <Save className="w-3 h-3" />
                {saving ? 'Guardando…' : 'Guardar'}
              </Button>
            </>
          ) : (
            <>
              {showEditToggle && (
                <Button
                  size="sm"
                  onClick={() => setEditing(true)}
                  disabled={loading || missing || !!error}
                  title="Editar inline"
                >
                  <Pencil className="w-3 h-3" /> Editar
                </Button>
              )}
              {!perm.editable && perm.reason && (
                <span
                  className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] bg-[var(--color-panel2)] border border-[var(--color-border)] text-[var(--color-muted)] cursor-help"
                  title={perm.reason}
                >
                  <Lock className="w-3 h-3" /> read-only
                </span>
              )}
              <Button
                size="sm"
                variant="ghost"
                onClick={handleOpenInEditor}
                title="Abrir en editor externo del host"
              >
                <ExternalLink className="w-3 h-3" />
              </Button>
            </>
          )}
        </div>
      </div>

      {perm.appendOnly && !editing && perm.reason && (
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1.5 bg-[var(--color-panel2)] border border-[var(--color-border)] rounded">
          ⓘ {perm.reason}
        </div>
      )}

      {loading && (
        <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
          <Spinner /> Cargando archivo…
        </div>
      )}
      {error && <ErrorBanner message={error} />}
      {missing && (
        <EmptyState>{missingMessage ?? 'archivo no existe todavía.'}</EmptyState>
      )}

      {!loading && !error && !missing && content !== null && !editing && (
        <article className="cockpit-md text-[13px] leading-relaxed text-[var(--color-text)] max-w-none">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeHighlight]}
          >
            {content}
          </ReactMarkdown>
        </article>
      )}

      {editing && draft !== null && (
        <div data-color-mode="dark">
          <MDEditor
            value={draft}
            onChange={(v) => setDraft(v ?? '')}
            preview="edit"
            height={500}
            visibleDragbar={false}
            textareaProps={{
              placeholder: 'Escribe en markdown…',
              spellCheck: false,
            }}
          />
        </div>
      )}
    </div>
  );
}
