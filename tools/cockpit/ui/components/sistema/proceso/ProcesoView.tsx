'use client';

import { useEffect, useMemo, useState } from 'react';
import { Workflow, ArrowRight, ShieldCheck, Play, Square } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import { getProceso } from '@/lib/api-client';
import type { ProcesoBinding, ProcesoEstado, ProcesoResponse } from '@/lib/types';

/**
 * Vista Proceso (F4 · I-77 · SPEC specs/proceso-descriptor.md RN-33).
 *
 * Rinde el descriptor de proceso que gobierna el motor — TODO desde `/api/proceso`,
 * CERO literales de estado en este archivo: si el kit shipea otro descriptor, esta
 * vista pinta otro proceso sin tocar código. Los únicos nombres fijos son las 5
 * categorías del contrato L0 (inmutables desde v1 — regla Azure/D4); el id nativo
 * del estado se muestra SIEMPRE al lado (DevLake status + original_status).
 */

// categoría FIJA del contrato → estilo (misma paleta que la torre).
const CAT_CLS: Record<string, string> = {
  propuesto: 'bg-[#1e3a5f] text-[#93c5fd]',
  'en-progreso': 'bg-[#713f12] text-[#fbbf24]',
  completado: 'bg-[#14532d] text-[#86efac]',
  descartado: 'bg-[#27272a] text-[#a1a1aa]',
  pausado: 'bg-[#1e1b2e] text-[#a5b4fc] border border-dashed border-[#4c1d95]',
};

function CategoriaBadge({ categoria }: { categoria: string }) {
  return (
    <span
      className={cn(
        'inline-block rounded px-1.5 py-0.5 text-[10px] font-medium leading-none',
        CAT_CLS[categoria] ?? 'bg-[#27272a] text-[#a1a1aa]'
      )}
    >
      {categoria}
    </span>
  );
}

function BindingChip({ b }: { b: ProcesoBinding }) {
  return (
    <span className="inline-flex items-baseline gap-1 rounded bg-[var(--color-panel)] border border-[var(--color-border)] px-1.5 py-0.5 text-[11px]">
      <code className="text-[var(--color-accent)]">{b.arnes ?? b.rol}</code>
      <span className="text-[var(--color-text-dim)]">· {b.rol}</span>
      {b.nota ? <span className="text-[var(--color-text-dim)]">({b.nota})</span> : null}
    </span>
  );
}

export function ProcesoView() {
  const [data, setData] = useState<ProcesoResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sel, setSel] = useState<string | null>(null);

  useEffect(() => {
    getProceso()
      .then(setData)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : String(e)));
  }, []);

  const seleccionado: ProcesoEstado | null = useMemo(
    () => data?.estados.find((e) => e.id === sel) ?? null,
    [data, sel]
  );

  if (error) return <ErrorBanner message={error} />;
  if (!data) return <Spinner />;

  const transDesde = (id: string) => data.transiciones.filter((t) => t.de === id);
  const transHacia = (id: string) => data.transiciones.filter((t) => t.a === id);
  const gatesDe = (id: string) =>
    data.gates.filter((g) =>
      g.momento.some((m) => m === id || m.startsWith(`${id}→`) || m.endsWith(`→${id}`))
    );
  const porEjecutor = (ej: 'operador' | 'rol') => data.transiciones.filter((t) => t.ejecutor === ej);

  return (
    <div className="p-4 space-y-4 max-w-5xl">
      {/* Cabecera: la identidad del descriptor — el proceso es DATO, no código */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-sm font-semibold flex items-center gap-2">
            <Workflow size={16} className="text-[var(--color-accent)]" />
            {data.descriptor.nombre}
            <span className="text-[var(--color-text-dim)] font-normal">
              <code>{data.descriptor.id}</code> · v{data.descriptor.version}
            </span>
          </h1>
          {data.descriptor.descripcion ? (
            <p className="mt-1 text-xs text-[var(--color-text-dim)] max-w-2xl">
              {data.descriptor.descripcion}
            </p>
          ) : null}
        </div>
        <div className="text-right text-[10px] text-[var(--color-text-dim)] shrink-0">
          <div>fuente: {data.fuente}</div>
          <div className="mt-1 flex gap-1 justify-end">
            {data.categorias.map((c) => (
              <CategoriaBadge key={c} categoria={c} />
            ))}
          </div>
        </div>
      </div>

      {/* Ciclo: estados en su orden canónico — id NATIVO grande + categoría al lado */}
      <Card>
        <div className="text-[10px] uppercase tracking-wide text-[var(--color-text-dim)] mb-2">
          Ciclo · {data.estados.length} estados
        </div>
        <div className="flex flex-wrap gap-2">
          {data.estados.map((e) => (
            <button
              key={e.id}
              onClick={() => setSel(sel === e.id ? null : e.id)}
              className={cn(
                'flex flex-col items-start gap-1 rounded border px-2.5 py-1.5 text-left transition-colors',
                sel === e.id
                  ? 'border-[var(--color-accent)] bg-[var(--color-panel)]'
                  : 'border-[var(--color-border)] hover:border-[var(--color-text-dim)]'
              )}
            >
              <span className="text-xs font-medium flex items-center gap-1.5">
                {e.inicial ? <Play size={10} className="text-[#93c5fd]" /> : null}
                {e.terminal ? <Square size={10} className="text-[#86efac]" /> : null}
                <code>{e.id}</code>
              </span>
              <CategoriaBadge categoria={e.categoria} />
            </button>
          ))}
        </div>

        {seleccionado ? (
          <div className="mt-3 border-t border-[var(--color-border)] pt-3 space-y-2 text-xs">
            <div className="flex items-center gap-2">
              <code className="font-semibold">{seleccionado.id}</code>
              <CategoriaBadge categoria={seleccionado.categoria} />
              {seleccionado.inicial ? <span className="text-[10px] text-[#93c5fd]">estado inicial</span> : null}
              {seleccionado.terminal ? <span className="text-[10px] text-[#86efac]">terminal (por categoría)</span> : null}
              {seleccionado.wip ? (
                <span className="text-[10px] text-[#fbbf24]">WIP ≤ {seleccionado.wip}</span>
              ) : null}
            </div>
            {seleccionado.descripcion ? (
              <p className="text-[var(--color-text-dim)]">{seleccionado.descripcion}</p>
            ) : null}
            {(data.duenos[seleccionado.id] ?? []).length > 0 ? (
              <div className="flex flex-wrap items-center gap-1.5">
                <span className="text-[var(--color-text-dim)]">llega por:</span>
                {(data.duenos[seleccionado.id] ?? []).map((b, i) => (
                  <BindingChip key={i} b={b} />
                ))}
              </div>
            ) : null}
            <div className="flex flex-wrap gap-x-4 gap-y-1 text-[var(--color-text-dim)]">
              {transHacia(seleccionado.id).length > 0 ? (
                <span>
                  desde:{' '}
                  {transHacia(seleccionado.id).map((t) => (
                    <code key={`${t.de}${t.a}`} className="mr-1">{t.de}</code>
                  ))}
                </span>
              ) : null}
              {transDesde(seleccionado.id).length > 0 ? (
                <span>
                  hacia:{' '}
                  {transDesde(seleccionado.id).map((t) => (
                    <code key={`${t.de}${t.a}`} className="mr-1">{t.a}</code>
                  ))}
                </span>
              ) : null}
            </div>
            {gatesDe(seleccionado.id).length > 0 ? (
              <div className="flex flex-wrap items-center gap-1.5">
                <span className="text-[var(--color-text-dim)]">gates:</span>
                {gatesDe(seleccionado.id).map((g) => (
                  <span key={g.id} className="text-[11px]">
                    <ShieldCheck size={11} className="inline mr-0.5 text-[var(--color-accent)]" />
                    {g.nombre}
                  </span>
                ))}
              </div>
            ) : null}
          </div>
        ) : null}
      </Card>

      {/* Transiciones por ejecutor */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {(['operador', 'rol'] as const).map((ej) => (
          <Card key={ej}>
            <div className="text-[10px] uppercase tracking-wide text-[var(--color-text-dim)] mb-2">
              Transiciones · {ej === 'operador' ? 'operador (el motor las evalúa)' : 'rol (las ejecuta el skill dueño)'}
            </div>
            <div className="space-y-1">
              {porEjecutor(ej).map((t) => (
                <div key={`${t.de}→${t.a}`} className="flex items-center gap-2 text-xs flex-wrap">
                  <code>{t.de}</code>
                  <ArrowRight size={11} className="text-[var(--color-text-dim)]" />
                  <code>{t.a}</code>
                  {t.verbo ? (
                    <span
                      className="rounded bg-[var(--color-panel)] border border-[var(--color-border)] px-1.5 py-0.5 text-[10px]"
                      title="El evento que la transición emite (CDEvents <sujeto>.<predicado> — RN-45)"
                    >
                      <code className="text-[var(--color-accent)]">{t.verbo}</code>
                    </span>
                  ) : null}
                  {t.nombre ? (
                    <span className="text-[10px] text-[var(--color-text-dim)]">«{t.nombre}»</span>
                  ) : null}
                  {t.requiere_razon ? (
                    <span className="text-[10px] text-[#fbbf24]">
                      ⚡ razón ≥ {data.parametros.razon_minima} caracteres
                    </span>
                  ) : null}
                </div>
              ))}
              {porEjecutor(ej).length === 0 ? (
                <div className="text-xs text-[var(--color-text-dim)]">— ninguna —</div>
              ) : null}
            </div>
          </Card>
        ))}
      </div>

      {/* Gates: el gate ES su checklist (Essence) */}
      <Card>
        <div className="text-[10px] uppercase tracking-wide text-[var(--color-text-dim)] mb-2">
          Gates · checklists como dato
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {data.gates.map((g) => (
            <div key={g.id} className="rounded border border-[var(--color-border)] p-2.5 space-y-1.5">
              <div className="text-xs font-medium flex items-center gap-1.5">
                <ShieldCheck size={12} className="text-[var(--color-accent)]" />
                {g.nombre}
              </div>
              <div className="text-[10px] text-[var(--color-text-dim)]">
                momento: {g.momento.map((m) => <code key={m} className="mr-1">{m}</code>)}
              </div>
              <div className="text-[10px] text-[var(--color-text-dim)] flex items-center gap-1">
                autoridad: <BindingChip b={g.autoridad} />
              </div>
              <ul className="text-[11px] list-disc list-inside space-y-0.5">
                {g.checklist.map((item, i) => (
                  <li key={i}>{item}</li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
