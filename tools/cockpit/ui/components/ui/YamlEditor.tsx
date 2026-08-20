'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import dynamic from 'next/dynamic';
import { ExternalLink, Lock, Pencil, RotateCcw, Save } from 'lucide-react';
import toast from 'react-hot-toast';
import yaml from 'yaml';
import { Spinner, ErrorBanner, EmptyState } from './Spinner';
import { Button } from './Button';
import { getFile, openInEditor, putFile } from '@/lib/api-client';
import {
  type ArtifactKind,
  getArtifactKind,
  getEditPermission,
} from '@/lib/edit-permissions';
import { useProceso } from '@/components/providers/ProcesoProvider';
import type { StoryState } from '@/lib/types';

// CodeMirror usa window → dynamic import sin SSR
const CodeMirror = dynamic(() => import('@uiw/react-codemirror'), {
  ssr: false,
  loading: () => (
    <div className="text-xs text-[var(--color-muted)] italic p-3">
      Cargando editor YAML…
    </div>
  ),
});

interface CmModule {
  yaml: () => unknown;
  EditorView: { lineWrapping: unknown };
}

let _cm: Promise<CmModule> | null = null;
function loadCm(): Promise<CmModule> {
  if (!_cm) {
    _cm = Promise.all([
      import('@codemirror/lang-yaml'),
      import('@codemirror/view'),
    ]).then(([langYaml, view]) => ({
      yaml: langYaml.yaml,
      EditorView: view.EditorView,
    }));
  }
  return _cm;
}

export interface YamlEditorProps {
  relPath: string;
  storyState?: StoryState;
  isArchived?: boolean;
  kindOverride?: ArtifactKind;
  missingMessage?: string;
  /** Si true, oculta header path + botones · útil cuando se embebe (CheckpointTab). */
  embedded?: boolean;
  /** Callback opcional post-save (parent puede refrescar). */
  onSaved?: () => void;
}

interface CmAddons {
  extensions: unknown[];
}

export function YamlEditor({
  relPath,
  storyState,
  isArchived = false,
  kindOverride,
  missingMessage,
  embedded = false,
  onSaved,
}: YamlEditorProps) {
  const [content, setContent] = useState<string | null>(null);
  const [draft, setDraft] = useState<string | null>(null);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [missing, setMissing] = useState(false);
  const [parseError, setParseError] = useState<string | null>(null);
  const [cmAddons, setCmAddons] = useState<CmAddons | null>(null);

  const kind = kindOverride ?? getArtifactKind(relPath);
  const proceso = useProceso();
  const perm = getEditPermission({
    kind,
    storyState: storyState ?? 'done',
    isArchived,
    proceso,
  });
  const showEditToggle = perm.editable && !perm.appendOnly;

  useEffect(() => {
    loadCm()
      .then((mod) => {
        setCmAddons({
          extensions: [mod.yaml(), mod.EditorView.lineWrapping],
        });
      })
      .catch((err) => {
        console.error('failed to load codemirror addons', err);
      });
  }, []);

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

  function validateYaml(text: string): string | null {
    try {
      yaml.parse(text);
      return null;
    } catch (err) {
      return (err as Error).message;
    }
  }

  function handleDraftChange(v: string) {
    setDraft(v);
    setParseError(validateYaml(v));
  }

  async function handleOpenInEditor() {
    try {
      const res = await openInEditor(relPath);
      toast.success(`abierto en ${res?.editor ?? 'editor externo'}`);
    } catch (err) {
      const e = err as Error & { detail?: { hint?: string } };
      toast.error(e.detail?.hint ?? `error: ${e.message}`);
    }
  }

  async function handleSave() {
    if (draft === null) return;
    const err = validateYaml(draft);
    if (err) {
      toast.error(`YAML inválido · arreglá errores antes de guardar`);
      setParseError(err);
      return;
    }
    setSaving(true);
    try {
      await putFile(relPath, draft);
      setContent(draft);
      setEditing(false);
      setParseError(null);
      toast.success('Guardado');
      onSaved?.();
    } catch (e) {
      toast.error(`error guardando: ${(e as Error).message}`);
    } finally {
      setSaving(false);
    }
  }

  function handleDiscard() {
    setDraft(content);
    setParseError(null);
    setEditing(false);
  }

  const dirty = editing && draft !== content;

  const extensions = useMemo(() => cmAddons?.extensions ?? [], [cmAddons]);

  return (
    <div className="space-y-3">
      {!embedded && (
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
                  disabled={saving || !dirty || !!parseError}
                  title={parseError ?? undefined}
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
                    title="Editar YAML inline"
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
      )}

      {embedded && editing && (
        <div className="flex items-center gap-1.5 justify-end">
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
            disabled={saving || !dirty || !!parseError}
            title={parseError ?? undefined}
          >
            <Save className="w-3 h-3" />
            {saving ? 'Guardando…' : 'Guardar'}
          </Button>
        </div>
      )}

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

      {!loading && !error && !missing && content !== null && (
        <>
          {editing && draft !== null ? (
            <div className="border border-[var(--color-border)] rounded overflow-hidden">
              <CodeMirror
                value={draft}
                onChange={handleDraftChange}
                height="450px"
                extensions={extensions as never}
                basicSetup={{
                  lineNumbers: true,
                  highlightActiveLine: true,
                  foldGutter: true,
                  bracketMatching: true,
                  closeBrackets: true,
                  autocompletion: false,
                  indentOnInput: true,
                  tabSize: 2,
                }}
                theme="dark"
              />
              {parseError && (
                <div className="text-[11px] text-red-300 bg-red-950/30 border-t border-red-900 px-3 py-2 font-mono whitespace-pre-wrap">
                  YAML error: {parseError}
                </div>
              )}
            </div>
          ) : (
            <pre className="bg-[var(--color-panel2)] border border-[var(--color-border)] rounded p-3 text-xs whitespace-pre-wrap font-mono leading-relaxed text-[var(--color-text)] overflow-x-auto max-h-[60vh]">
              {content}
            </pre>
          )}
        </>
      )}
    </div>
  );
}
