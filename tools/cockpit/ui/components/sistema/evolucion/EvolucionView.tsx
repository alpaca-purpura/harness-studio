'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import { Pill } from '@/components/ui/Badge';
import { useSistema } from '@/components/providers/SistemaProvider';
import { isPlatform } from '@/lib/platform-context';
import { getLedger } from '@/lib/api-client';
import type { Ledger, LedgerFicha } from '@/lib/types';

/**
 * Lente "Evolución" (I-45) — para sistemas NO-SDD (la fábrica). Su board SDD sale
 * vacío porque no corre la máquina de 10 estados sobre sí mismo; su evolución vive
 * en el Ledger (fichas I-NN + log de decisiones) + el semver. Lee el artefacto
 * generado docs/product/ledger.yaml (← PRODUCT-VISION.md, gated). Sistemas que SÍ
 * corren SDD no tienen ledger.yaml → empty-state que apunta al Roadmap.
 */

const ESTADO_BADGE: Record<string, string> = {
  decidida: 'bg-[#14532d] text-[#86efac]',
  capturada: 'bg-[#1e3a5f] text-[#93c5fd]',
  'en-discusión': 'bg-[#713f12] text-[#fbbf24]',
  descartada: 'bg-[#3f3f46] text-[#a1a1aa]',
  parking: 'bg-[#27272a] text-[#a1a1aa]',
};

const ESTADO_ORDER = ['decidida', 'en-discusión', 'capturada', 'parking', 'descartada'];

function estadoCls(estado: string): string {
  return ESTADO_BADGE[estado] ?? 'bg-[#27272a] text-[#d4d4d8]';
}

export function EvolucionView() {
  const { sistema } = useSistema();
  const [ledger, setLedger] = useState<Ledger | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filtroEstado, setFiltroEstado] = useState<string>('all');

  const load = useCallback(() => {
    if (isPlatform(sistema)) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    getLedger(sistema)
      .then((data) => setLedger(data))
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  // conteo por estado (chips de filtro)
  const counts = useMemo(() => {
    const c: Record<string, number> = {};
    for (const f of ledger?.fichas ?? []) c[f.estado] = (c[f.estado] ?? 0) + 1;
    return c;
  }, [ledger]);

  const fichas: LedgerFicha[] = useMemo(() => {
    const all = ledger?.fichas ?? [];
    return filtroEstado === 'all' ? all : all.filter((f) => f.estado === filtroEstado);
  }, [ledger, filtroEstado]);

  // decisiones: el log §5 está en orden cronológico (viejo→nuevo); mostramos nuevo→viejo
  const decisiones = useMemo(
    () => [...(ledger?.decisiones ?? [])].reverse(),
    [ledger]
  );

  if (loading) return <div className="p-8"><Spinner /></div>;
  if (error) return <div className="p-6"><ErrorBanner message={error} /></div>;

  if (!ledger) {
    return (
      <div className="p-6">
        <EmptyState>
          Este sistema no trackea su evolución por un ledger. Si corre SDD, su
          evolución vive en el <span className="text-[#93c5fd]">Board</span> y el{' '}
          <span className="text-[#93c5fd]">Roadmap</span> (stories + releases).
        </EmptyState>
      </div>
    );
  }

  const estadosPresentes = ESTADO_ORDER.filter((e) => counts[e]);

  return (
    <div className="mx-auto max-w-5xl space-y-6 p-6">
      {/* Cabecera: sistema · versión · resumen */}
      <header className="space-y-2">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-[#e4e4e7]">{ledger.titulo}</h1>
          {ledger.version && <Pill>{ledger.version}</Pill>}
        </div>
        <p className="text-sm text-[#a1a1aa]">
          {ledger.fichas.length} decisiones de producto (fichas I-NN) ·{' '}
          {ledger.decisiones.length} entradas en el log · fuente{' '}
          <code className="text-[#71717a]">{ledger.fuente}</code>
        </p>
      </header>

      {/* Log de decisiones (timeline, nuevo→viejo) */}
      <Card className="p-0">
        <div className="border-b border-[#27272a] px-4 py-3">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-[#a1a1aa]">
            Línea de tiempo
          </h2>
        </div>
        <ol className="divide-y divide-[#27272a]">
          {decisiones.map((d, i) => (
            <li key={i} className="flex gap-4 px-4 py-3">
              <time className="w-24 shrink-0 font-mono text-xs text-[#71717a]">
                {d.fecha}
              </time>
              <div className="min-w-0 flex-1 space-y-1">
                <p className="text-sm text-[#d4d4d8]">{d.decision}</p>
                {d.ideas.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {d.ideas.map((id) => (
                      <span
                        key={id}
                        className="rounded bg-[#27272a] px-1.5 py-0.5 font-mono text-[10px] text-[#a1a1aa]"
                      >
                        {id}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </li>
          ))}
        </ol>
      </Card>

      {/* Fichas I-NN con filtro por estado */}
      <section className="space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="mr-1 text-sm font-semibold uppercase tracking-wide text-[#a1a1aa]">
            Fichas
          </h2>
          <FiltroChip
            label={`todas · ${ledger.fichas.length}`}
            active={filtroEstado === 'all'}
            onClick={() => setFiltroEstado('all')}
          />
          {estadosPresentes.map((e) => (
            <FiltroChip
              key={e}
              label={`${e} · ${counts[e]}`}
              active={filtroEstado === e}
              cls={estadoCls(e)}
              onClick={() => setFiltroEstado(e)}
            />
          ))}
        </div>

        <ul className="space-y-1.5">
          {fichas.map((f) => (
            <li
              key={f.id}
              className="flex items-start gap-3 rounded-md border border-[#27272a] bg-[#18181b] px-3 py-2"
            >
              <span className="w-12 shrink-0 font-mono text-xs text-[#71717a]">{f.id}</span>
              <span className="min-w-0 flex-1 text-sm text-[#d4d4d8]">
                {f.titulo}
                {f.nota && <span className="ml-2 text-xs text-[#71717a]">· {f.nota}</span>}
              </span>
              <span
                className={cn(
                  'shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium',
                  estadoCls(f.estado)
                )}
              >
                {f.estado}
              </span>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}

function FiltroChip({
  label,
  active,
  cls,
  onClick,
}: {
  label: string;
  active: boolean;
  cls?: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'rounded-full px-2.5 py-0.5 text-xs transition-colors',
        active
          ? cls ?? 'bg-[#3f3f46] text-[#e4e4e7]'
          : 'bg-[#18181b] text-[#a1a1aa] hover:bg-[#27272a]'
      )}
    >
      {label}
    </button>
  );
}
