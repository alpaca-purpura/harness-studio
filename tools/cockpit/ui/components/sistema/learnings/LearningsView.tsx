'use client';

/**
 * LearningsView v2 (2026-06-11, pedido operador) — el carril L2 del CIL deja de
 * ser un cementerio:
 *   · 3 fuentes (sistema + transversal + tooling) con badge de origen
 *   · ESTADO de ciclo de vida: ⏳ pending (espera decisión) arriba en ámbar ·
 *     referencia y cerrados colapsados
 *   · detalle in-app (drawer con markdown completo, cero abrir archivos)
 *   · decisión con un click: ✓ aplicado · ⬆ promovido a regla · ✗ no aplica
 *   · tags con señal sistémica (3+ pending del mismo tema = atacar la causa)
 *
 * Doctrina: capturado NO es estado final — cada learning muere en una de 4
 * salidas (gate/hook > rule/skill > proposal core > referencia consciente).
 */

import { useCallback, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { ExternalLink, Search } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Pill } from '@/components/ui/Badge';
import { Drawer } from '@/components/ui/Drawer';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchEvents } from '@/components/providers/FileWatchProvider';
import {
  listLearnings,
  patchLearning,
  getFile,
  openInEditor,
  type LearningEntry,
  type LearningApplied,
} from '@/lib/api-client';

const TYPE_CLASSES: Record<string, string> = {
  technical: 'bg-[#1e3a8a] text-[#93c5fd]',
  business: 'bg-[#365314] text-[#a3e635]',
  process: 'bg-[#5b21b6] text-[#ddd6fe]',
  tooling: 'bg-[#713f12] text-[#fbbf24]',
};

const SOURCE_META: Record<string, { label: string; tip: string }> = {
  sistema: { label: 'sistema', tip: 'Vive en {sistema}/docs/learnings/ — aprendizaje de negocio del vertical.' },
  transversal: { label: 'transversal', tip: 'Vive en docs/learnings/ — aprendizaje técnico que aplica a ≥2 sistemas.' },
  tooling: { label: 'tooling', tip: 'Vive en docs/learnings/tooling/ — aprendizaje de workspace/harness.' },
};

const STATUS_META: Record<string, { label: string; cls: string; tip: string }> = {
  pending: { label: '⏳ espera decisión', cls: 'bg-[#1f1300] text-amber-400 border border-amber-800', tip: 'Capturado pero sin triage: nadie decidió aún si se automatiza, se vuelve regla, se promueve o se archiva.' },
  applied: { label: '✓ aplicado', cls: 'bg-[#052e16] text-[#86efac]', tip: 'La acción derivada ya se ejecutó (gate/hook/fix/cambio de proceso).' },
  promoted: { label: '⬆ promovido', cls: 'bg-[#1e3a5f] text-[#93c5fd]', tip: 'Se convirtió en regla/skill del harness o proposal a core.' },
  'wont-apply': { label: '✗ no aplica', cls: 'bg-[#27272a] text-[#a1a1aa]', tip: 'Decisión consciente de no actuar — queda como registro.' },
  reference: { label: '📚 referencia', cls: 'bg-[var(--color-panel2)] text-[var(--color-muted)]', tip: 'promotable: no — conocimiento de consulta, no espera acción.' },
};

function StatusPill({ status }: { status: string }) {
  const m = STATUS_META[status] ?? STATUS_META.reference;
  return (
    <span className={cn('text-[9px] px-1.5 py-0.5 rounded font-medium shrink-0', m.cls)} title={m.tip}>
      {m.label}
    </span>
  );
}

function LearningCard({
  l,
  onOpen,
}: {
  l: LearningEntry;
  onOpen: (l: LearningEntry) => void;
}) {
  const src = SOURCE_META[l.source] ?? SOURCE_META.sistema;
  return (
    <Card
      className="!p-3 cursor-pointer hover:border-[var(--color-accent)] transition-colors"
      onClick={() => onOpen(l)}
    >
      <div className="flex items-center gap-2 mb-1 flex-wrap">
        {l.date && (
          <span className="font-mono text-[10px] text-[var(--color-muted)]">{l.date}</span>
        )}
        {l.type && (
          <Pill className={cn(TYPE_CLASSES[l.type] ?? 'bg-[var(--color-panel2)]', 'text-[9px] py-0')}>
            {l.type}
          </Pill>
        )}
        <span
          className="text-[9px] px-1.5 py-0.5 rounded border border-[var(--color-border)] text-[var(--color-muted)]"
          title={src.tip}
        >
          {src.label}
        </span>
        {l.promotable && l.promotable !== 'no' && (
          <span
            className="text-[9px] px-1.5 py-0.5 rounded bg-[#2d1f00] text-amber-300"
            title="El autor lo marcó como candidato a convertirse en regla/patrón compartido."
          >
            promotable: {l.promotable}
          </span>
        )}
        <span className="ml-auto">
          <StatusPill status={l.status} />
        </span>
      </div>
      <div className="text-xs font-medium leading-snug">{l.title ?? l.slug}</div>
      {l.preview && (
        <p className="text-[10px] text-[var(--color-muted)] mt-1 leading-relaxed line-clamp-2">
          {l.preview}
        </p>
      )}
      {(l.tags ?? []).length > 0 && (
        <div className="flex flex-wrap gap-1 mt-1.5">
          {l.tags!.slice(0, 6).map((t) => (
            <span key={t} className="text-[9px] text-[var(--color-muted)] font-mono">
              #{t}
            </span>
          ))}
        </div>
      )}
    </Card>
  );
}

export function LearningsView() {
  const { sistema } = useSistema();
  const [items, setItems] = useState<LearningEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [tagFilter, setTagFilter] = useState<string | null>(null);
  const [selected, setSelected] = useState<LearningEntry | null>(null);
  const [content, setContent] = useState<string | null>(null);
  const [deciding, setDeciding] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    listLearnings(sistema)
      .then(setItems)
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  useFileWatchEvents((event) => {
    if (event.sistema && event.sistema !== sistema) return;
    if (event.docType === 'learning') load();
  });

  // Detalle: fetch del markdown completo al abrir
  useEffect(() => {
    if (!selected) {
      setContent(null);
      return;
    }
    getFile(selected.rel_path)
      .then(setContent)
      .catch(() => setContent('_No se pudo cargar el archivo._'));
  }, [selected]);

  const filtered = useMemo(() => {
    let out = items;
    if (tagFilter) out = out.filter((l) => (l.tags ?? []).includes(tagFilter));
    if (search) {
      const q = search.toLowerCase();
      out = out.filter((l) =>
        [l.slug, l.title ?? '', l.preview ?? '', ...(l.tags ?? []), l.type ?? '']
          .join(' ')
          .toLowerCase()
          .includes(q)
      );
    }
    return out;
  }, [items, search, tagFilter]);

  const pending = filtered.filter((l) => l.status === 'pending');
  const reference = filtered.filter((l) => l.status === 'reference');
  const closed = filtered.filter(
    (l) => l.status === 'applied' || l.status === 'promoted' || l.status === 'wont-apply'
  );

  // Tags sistémicos: 3+ learnings PENDING con el mismo tag = problema de fondo.
  const tagCounts = useMemo(() => {
    const counts = new Map<string, number>();
    for (const l of items.filter((i) => i.status === 'pending')) {
      for (const t of l.tags ?? []) counts.set(t, (counts.get(t) ?? 0) + 1);
    }
    return [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 14);
  }, [items]);

  async function decide(l: LearningEntry, applied: LearningApplied) {
    setDeciding(true);
    try {
      await patchLearning(l.rel_path, applied);
      toast.success(
        applied === 'pending' ? 'Reabierto.' : `Marcado: ${STATUS_META[applied]?.label ?? applied}.`
      );
      setSelected(null);
      load();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setDeciding(false);
    }
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center gap-2 text-sm text-[var(--color-muted)]">
        <Spinner /> Cargando learnings…
      </div>
    );
  }
  if (error) {
    return (
      <div className="p-6">
        <ErrorBanner message={error} />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <header className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-lg font-semibold">Learnings · mejora continua (L2)</h1>
          <p className="text-[11px] text-[var(--color-muted)] mt-1">
            <span
              className={pending.length > 0 ? 'text-amber-300 font-medium' : 'text-[#86efac] font-medium'}
              title="Capturados sin decisión. La meta es que este número BAJE: cada uno muere en gate/hook, regla/skill, proposal o referencia consciente."
            >
              {pending.length} esperando decisión
            </span>
            {' · '}
            {closed.length} cerrados · {reference.length} referencia · {items.length} total
            <span className="text-[var(--color-muted)]"> (sistema + transversal + tooling)</span>
          </p>
        </div>
        <div className="flex items-center gap-1.5 max-w-xs">
          <Search className="w-3.5 h-3.5 text-[var(--color-muted)]" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por título, tag, contenido…"
            className="!py-1"
          />
        </div>
      </header>

      {/* Tags de los pendientes — 🔥 = 3+ del mismo tema (señal sistémica) */}
      {tagCounts.length > 0 && (
        <div className="flex flex-wrap gap-1.5 items-center">
          <span className="text-[10px] text-[var(--color-muted)]">temas pendientes:</span>
          {tagCounts.map(([t, n]) => (
            <button
              key={t}
              type="button"
              onClick={() => setTagFilter(tagFilter === t ? null : t)}
              title={
                n >= 3
                  ? `🔥 ${n} learnings pendientes del mismo tema — señal SISTÉMICA: atacá la causa con UNA regla/gate, no parches.`
                  : `${n} pendiente${n > 1 ? 's' : ''} con este tag.`
              }
              className={cn(
                'text-[10px] px-2 py-0.5 rounded-full border font-mono transition-colors',
                tagFilter === t
                  ? 'border-[var(--color-accent)] text-[var(--color-text)] bg-[var(--color-panel2)]'
                  : 'border-[var(--color-border)] text-[var(--color-muted)] hover:text-[var(--color-text)]'
              )}
            >
              {n >= 3 ? '🔥 ' : ''}#{t} ({n})
            </button>
          ))}
          {tagFilter && (
            <button
              type="button"
              onClick={() => setTagFilter(null)}
              className="text-[10px] text-[var(--color-muted)] underline"
            >
              limpiar filtro
            </button>
          )}
        </div>
      )}

      {/* ⏳ Esperando decisión — lo único que pide tu atención */}
      <section>
        <h2 className="text-sm font-semibold mb-2 text-amber-300">
          ⏳ Esperando decisión ({pending.length})
        </h2>
        {pending.length === 0 ? (
          <Card className="!p-4 text-center text-xs text-[#86efac]">
            🎉 Cero learnings sin triage — el loop de mejora continua está al día.
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
            {pending.map((l) => (
              <LearningCard key={l.rel_path} l={l} onOpen={setSelected} />
            ))}
          </div>
        )}
      </section>

      {/* Cerrados + referencia — colapsados */}
      {closed.length > 0 && (
        <details className="border border-[var(--color-border)] rounded-lg bg-[var(--color-panel)]">
          <summary className="px-4 py-3 text-sm cursor-pointer select-none">
            ✓ Cerrados ({closed.length})
            <span className="text-[10px] text-[var(--color-muted)] italic ml-2">
              aplicados · promovidos · descartados conscientes
            </span>
          </summary>
          <div className="px-4 pb-4 grid grid-cols-1 md:grid-cols-2 gap-2">
            {closed.map((l) => (
              <LearningCard key={l.rel_path} l={l} onOpen={setSelected} />
            ))}
          </div>
        </details>
      )}
      {reference.length > 0 && (
        <details className="border border-[var(--color-border)] rounded-lg bg-[var(--color-panel)]">
          <summary className="px-4 py-3 text-sm cursor-pointer select-none">
            📚 Referencia ({reference.length})
            <span className="text-[10px] text-[var(--color-muted)] italic ml-2">
              promotable: no — consulta, no esperan acción
            </span>
          </summary>
          <div className="px-4 pb-4 grid grid-cols-1 md:grid-cols-2 gap-2">
            {reference.map((l) => (
              <LearningCard key={l.rel_path} l={l} onOpen={setSelected} />
            ))}
          </div>
        </details>
      )}

      {filtered.length === 0 && (
        <EmptyState>
          {items.length === 0 ? 'Aún no hay learnings capturados.' : 'Sin resultados para ese filtro.'}
        </EmptyState>
      )}

      {/* ── Drawer de detalle — el learning completo sin abrir archivos ── */}
      <Drawer
        open={selected !== null}
        onClose={() => setSelected(null)}
        width={820}
        title={
          selected ? (
            <div className="flex items-center gap-2 min-w-0">
              <StatusPill status={selected.status} />
              <span className="text-sm font-medium truncate">
                {selected.title ?? selected.slug}
              </span>
            </div>
          ) : null
        }
      >
        {selected && (
          <div className="space-y-3">
            {/* Decisión — el corazón del loop */}
            <Card className="!p-3">
              <div className="text-[10px] text-[var(--color-muted)] mb-2">
                ¿Qué hacemos con este aprendizaje? (jerarquía: automatizarlo &gt; volverlo
                regla/skill &gt; proposal a core &gt; referencia consciente)
              </div>
              <div className="flex flex-wrap gap-2">
                {selected.status === 'pending' || selected.status === 'reference' ? (
                  <>
                    <Button
                      size="sm"
                      disabled={deciding}
                      onClick={() => decide(selected, 'applied')}
                      title="La acción derivada ya se ejecutó (gate/hook/fix/cambio de proceso)."
                    >
                      ✓ Aplicado
                    </Button>
                    <Button
                      size="sm"
                      disabled={deciding}
                      onClick={() => decide(selected, 'promoted')}
                      title="Se convirtió en regla/skill del harness o proposal a core."
                    >
                      ⬆ Promovido a regla
                    </Button>
                    <Button
                      size="sm"
                      disabled={deciding}
                      onClick={() => decide(selected, 'wont-apply')}
                      title="Decisión consciente de NO actuar — deja de pedir atención."
                    >
                      ✗ No aplica
                    </Button>
                  </>
                ) : (
                  <Button
                    size="sm"
                    disabled={deciding}
                    onClick={() => decide(selected, 'pending')}
                    title="Volver a la cola de decisión."
                  >
                    ↺ Reabrir
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={async () => {
                    try {
                      await openInEditor(selected.rel_path);
                      toast.success('abierto en editor');
                    } catch (err) {
                      toast.error((err as Error).message);
                    }
                  }}
                >
                  <ExternalLink className="w-3 h-3" /> abrir en editor
                </Button>
              </div>
            </Card>

            {/* Markdown completo */}
            {content === null ? (
              <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
                <Spinner /> Cargando…
              </div>
            ) : (
              <article className="cockpit-md text-[13px] leading-relaxed text-[var(--color-text)] max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
                  {content}
                </ReactMarkdown>
              </article>
            )}
          </div>
        )}
      </Drawer>
    </div>
  );
}
