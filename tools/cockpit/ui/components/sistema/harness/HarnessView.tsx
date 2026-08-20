'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { Wrench, RefreshCw, FileText, Search, BookOpen, AlertTriangle, ArrowUpRight } from 'lucide-react';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { listCilBoard, openInEditor, type CilBoard } from '@/lib/api-client';
import { ESTADO_ORDER, type HarnessEstado, type HarnessItem } from '@/lib/harness-backlog';
import type { TechDebtItem } from '@/lib/tech-debt';

const L1_SOURCE = 'docs/process/harness-backlog.md';
const L3_SOURCE = 'docs/process/tech-debt.md';

/** Qué significa cada estado + qué acción/decisión tomar (ver lib/tooltips.ts). */
const ESTADO_TOOLTIPS: Record<HarnessEstado, string> = {
  reported: TOOLTIPS.harness_reported,
  triaged: TOOLTIPS.harness_triaged,
  ratified: TOOLTIPS.harness_ratified,
  applied: TOOLTIPS.harness_applied,
  verified: TOOLTIPS.harness_verified,
  deferred: TOOLTIPS.harness_deferred,
  otro: TOOLTIPS.harness_otro,
};

/** Color por columna de estado · ámbar = avance del harness (distinto del púrpura de marca). */
const ESTADO_CLASSES: Record<HarnessEstado, string> = {
  reported: 'bg-[#1f2937] text-[#9ca3af]',
  triaged: 'bg-[#1e3a8a] text-[#93c5fd]',
  ratified: 'bg-[#5b21b6] text-[#ddd6fe]',
  applied: 'bg-[#713f12] text-[#fbbf24]',
  verified: 'bg-[#14532d] text-[#86efac]',
  deferred: 'bg-[#3f3f46] text-[#d4d4d8]',
  otro: 'bg-[#450a0a] text-[#fca5a5]',
};

const ESTADO_LABEL: Record<HarnessEstado, string> = {
  reported: 'reported',
  triaged: 'triaged',
  ratified: 'ratified',
  applied: 'applied',
  verified: 'verified',
  deferred: 'deferred',
  otro: 'otro',
};

/** Card de un item de lifecycle (HB de L1 o TD de L3) — misma forma de tabla. */
function LifecycleCard({
  id,
  title,
  sevEmoji,
  sevLabel,
  estado,
  estadoRaw,
  ref,
}: {
  id: string;
  title: string;
  sevEmoji: string;
  sevLabel: string;
  estado: HarnessEstado;
  estadoRaw: string;
  ref: string;
}) {
  const estadoNote = estadoRaw
    .replace(/\*\*|__|`/g, '')
    .replace(new RegExp(`^\\s*${estado}\\s*`, 'i'), '')
    .trim();

  return (
    <div className="bg-[var(--color-panel)] border border-[var(--color-border)] border-l-2 border-l-amber-500/70 rounded p-2.5 text-xs">
      <div className="flex items-center justify-between gap-2 mb-1">
        <span className="font-mono text-[11px] text-amber-400 font-semibold">{id}</span>
        <span title={sevLabel} aria-label={sevLabel}>
          {sevEmoji}
        </span>
      </div>
      <p className="text-[var(--color-text)] leading-snug line-clamp-4">{title}</p>
      {estadoNote && <p className="text-[10px] text-[var(--color-muted)] italic mt-1">{estadoNote}</p>}
      {ref && (
        <p className="text-[10px] text-[var(--color-muted)] font-mono mt-1 truncate" title={ref}>
          {ref}
        </p>
      )}
    </div>
  );
}

/** Board kanban por estado para una lista de items lifecycle (L1 o L3). */
function LifecycleBoard({
  items,
}: {
  items: Array<HarnessItem | TechDebtItem>;
}) {
  const columns = useMemo<HarnessEstado[]>(() => {
    const hasOtro = items.some((it) => it.estado === 'otro');
    return hasOtro ? [...ESTADO_ORDER, 'otro'] : [...ESTADO_ORDER];
  }, [items]);

  return (
    <div className="flex gap-3 overflow-x-auto pb-4">
      {columns.map((estado) => {
        const colItems = items
          .filter((it) => it.estado === estado)
          .sort((a, b) => a.num - b.num);
        return (
          <div key={estado} className="shrink-0 w-60 flex flex-col">
            <div className="flex items-center justify-between mb-2 px-1">
              <Tooltip content={ESTADO_TOOLTIPS[estado]} variant="badge">
                <Badge className={ESTADO_CLASSES[estado]}>{ESTADO_LABEL[estado]}</Badge>
              </Tooltip>
              <span className="text-[11px] text-[var(--color-muted)]">{colItems.length}</span>
            </div>
            <div className="flex flex-col gap-2">
              {colItems.length === 0 ? (
                <p className="text-[10px] text-[var(--color-muted)] italic px-1 py-2">—</p>
              ) : (
                colItems.map((it) => (
                  <LifecycleCard
                    key={it.id}
                    id={it.id}
                    title={it.title}
                    sevEmoji={it.sevEmoji}
                    sevLabel={it.sevLabel}
                    estado={it.estado}
                    estadoRaw={it.estadoRaw}
                    ref={it.ref}
                  />
                ))
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

/** Chip de un carril en la franja superior del CIL. */
function CarrilChip({
  code,
  label,
  count,
  tone,
  href,
}: {
  code: string;
  label: string;
  count: number | string;
  tone: 'amber' | 'sky' | 'rose' | 'violet';
  href?: string;
}) {
  const toneClass = {
    amber: 'border-amber-700/50 text-amber-300',
    sky: 'border-sky-700/50 text-sky-300',
    rose: 'border-rose-700/50 text-rose-300',
    violet: 'border-violet-700/50 text-violet-300',
  }[tone];

  const inner = (
    <div
      className={`flex items-center gap-2 px-3 py-2 rounded border bg-[var(--color-panel)] ${toneClass} ${
        href ? 'hover:bg-[var(--color-panel2)] transition-colors' : ''
      }`}
    >
      <span className="font-mono text-[11px] font-semibold">{code}</span>
      <span className="text-[11px] text-[var(--color-muted)]">{label}</span>
      <span className="ml-auto text-sm font-semibold">{count}</span>
      {href && <ArrowUpRight className="w-3 h-3 opacity-60" />}
    </div>
  );

  return href ? (
    <Link href={href} className="flex-1 min-w-[160px]">
      {inner}
    </Link>
  ) : (
    <div className="flex-1 min-w-[160px]">{inner}</div>
  );
}

export function HarnessView() {
  const [board, setBoard] = useState<CilBoard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    listCilBoard()
      .then(setBoard)
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const l1Filtered = useMemo<HarnessItem[]>(() => {
    const items = board?.l1.items ?? [];
    if (!search) return items;
    const q = search.toLowerCase();
    return items.filter((it) =>
      [it.id, it.item, it.estadoRaw, it.ref, it.sevLabel].join(' ').toLowerCase().includes(q)
    );
  }, [board, search]);

  const l3Filtered = useMemo<TechDebtItem[]>(() => {
    const items = board?.l3.items ?? [];
    if (!search) return items;
    const q = search.toLowerCase();
    return items.filter((it) =>
      [it.id, it.item, it.estadoRaw, it.ref, it.sevLabel].join(' ').toLowerCase().includes(q)
    );
  }, [board, search]);

  async function handleOpen(source: string) {
    try {
      await openInEditor(source);
      toast.success(`Abriendo ${source}…`);
    } catch (err) {
      toast.error(`No se pudo abrir: ${(err as Error).message}`);
    }
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center gap-2 text-sm text-[var(--color-muted)]">
        <Spinner /> Cargando CIL…
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
  if (!board) {
    return (
      <div className="p-6">
        <EmptyState>Sin datos del CIL en este worktree.</EmptyState>
      </div>
    );
  }

  return (
    <div className="p-6">
      <header className="mb-3">
        <h1 className="text-lg font-semibold flex items-center gap-2">
          <Wrench className="w-4 h-4 text-amber-400" /> Harness · CIL
          <Badge className="bg-amber-900/30 text-amber-300 border border-amber-700/50">
            4 carriles · read-only
          </Badge>
        </h1>
        <p className="text-[11px] text-[var(--color-muted)] italic mt-1">
          Continuous Improvement Loop — nada queda sin vigilar. La captura/transición la
          maneja el HLP (<span className="font-mono">/harness-issue</span> + dev-team/auditor);
          el cockpit solo lo visualiza. El stop semanal{' '}
          <span className="font-mono">/harnesses-improvement</span> remedia los 4 carriles.
        </p>
      </header>

      {/* Franja 4-carriles — el panorama del CIL de un vistazo. */}
      <div className="flex flex-wrap gap-2 mb-4">
        <CarrilChip code="L1" label="harness · abiertos" count={board.l1.open} tone="amber" />
        <CarrilChip
          code="L2"
          label="learnings"
          count={board.l2.count}
          tone="sky"
          href={board.l2.link}
        />
        <CarrilChip code="L3" label="deuda técnica · abiertos" count={board.l3.open} tone="rose" />
        <CarrilChip code="L4" label="caps desfasadas → Drift" count="ver" tone="violet" href={board.l4.link} />
      </div>

      <div className="flex items-center gap-2 mb-3 flex-wrap">
        <button
          onClick={() => handleOpen(L1_SOURCE)}
          className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded border border-[var(--color-border)] bg-[var(--color-panel2)] text-[var(--color-text)] hover:border-amber-600/60 transition-colors"
        >
          <FileText className="w-3.5 h-3.5" /> harness-backlog
        </button>
        <button
          onClick={() => handleOpen(L3_SOURCE)}
          className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded border border-[var(--color-border)] bg-[var(--color-panel2)] text-[var(--color-text)] hover:border-rose-600/60 transition-colors"
        >
          <FileText className="w-3.5 h-3.5" /> tech-debt
        </button>
        <button
          onClick={load}
          className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded border border-[var(--color-border)] bg-[var(--color-panel2)] text-[var(--color-text)] hover:border-amber-600/60 transition-colors"
        >
          <RefreshCw className="w-3.5 h-3.5" /> Recargar
        </button>
        <div className="flex items-center gap-1.5 flex-1 max-w-xs">
          <Search className="w-3.5 h-3.5 text-[var(--color-muted)]" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar id, texto, ref…"
            className="!py-1"
          />
        </div>
      </div>

      {/* L1 · harness-backlog */}
      <section className="mb-6">
        <h2 className="text-sm font-semibold flex items-center gap-2 mb-2">
          <Wrench className="w-3.5 h-3.5 text-amber-400" /> L1 · Harness backlog
          <Badge className="bg-[var(--color-panel2)] text-[var(--color-muted)]">
            {l1Filtered.length} items
          </Badge>
          <span className="text-[10px] text-[var(--color-muted)] font-mono">{L1_SOURCE}</span>
        </h2>
        {l1Filtered.length === 0 ? (
          <EmptyState>Sin items en el harness backlog (o el archivo no existe en este worktree).</EmptyState>
        ) : (
          <LifecycleBoard items={l1Filtered} />
        )}
      </section>

      {/* L3 · tech-debt */}
      <section className="mb-2">
        <h2 className="text-sm font-semibold flex items-center gap-2 mb-2">
          <AlertTriangle className="w-3.5 h-3.5 text-rose-400" /> L3 · Deuda técnica
          <Badge className="bg-[var(--color-panel2)] text-[var(--color-muted)]">
            {l3Filtered.length} items
          </Badge>
          <span className="text-[10px] text-[var(--color-muted)] font-mono">{L3_SOURCE}</span>
        </h2>
        {l3Filtered.length === 0 ? (
          <EmptyState>
            Sin deuda L3 registrada todavía — dev-team y <span className="font-mono">/auditor</span>{' '}
            appendean a <span className="font-mono">{L3_SOURCE}</span> al cerrar story.
          </EmptyState>
        ) : (
          <LifecycleBoard items={l3Filtered} />
        )}
      </section>

      {/* L2 + L4 · viven en sus vistas propias (link arriba). */}
      <p className="text-[10px] text-[var(--color-muted)] italic mt-3 flex items-center gap-1.5 flex-wrap">
        <BookOpen className="w-3 h-3" /> L2 (learnings) →{' '}
        <Link href={board.l2.link} className="text-sky-400 hover:underline">
          /learnings
        </Link>{' '}
        · L4 (capabilities desfasadas, por sistema) →{' '}
        <Link href={board.l4.link} className="text-violet-400 hover:underline">
          /drift
        </Link>
      </p>
    </div>
  );
}
