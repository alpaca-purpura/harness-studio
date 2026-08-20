'use client';

import { useCallback, useEffect, useState } from 'react';
import { ChevronDown, ChevronRight, RefreshCw, Gauge } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import { getTorre } from '@/lib/api-client';
import type { TorreProyecto, TorreResponse, TorreSistema, TorreVeredicto } from '@/lib/types';

/**
 * Torre de Control read-only (F2 · DH-04 · SPEC specs/torre-read-only.md).
 *
 * Tabla sistemas × columnas del eje REPO; fila expandible al eje PROYECTO (RN-19).
 * Cada celda es un objeto del dominio operable (RN-18): click → el dato nativo
 * completo + acción sugerida. Veredictos = categorías FIJAS + nativo al lado (D4);
 * `no-medido` ≠ `rojo` (RN-13); la EDAD del dato siempre visible (RN-17).
 * Terminología: «gate de fábrica», jamás «gate» pelado (colisión G1-G8 del board).
 */

// categoría FIJA → estilo. Los nombres son inmutables desde v1 (regla Azure).
const CAT_CLS: Record<string, string> = {
  // sano
  verde: 'bg-[#14532d] text-[#86efac]',
  limpio: 'bg-[#14532d] text-[#86efac]',
  sincronizada: 'bg-[#14532d] text-[#86efac]',
  taggeado: 'bg-[#1e3a5f] text-[#93c5fd]',
  'con-ledger': 'bg-[#1e3a5f] text-[#93c5fd]',
  'con-board': 'bg-[#1e3a5f] text-[#93c5fd]',
  'con-arquitectura': 'bg-[#1e3a5f] text-[#93c5fd]',
  // atención
  'en-vuelo': 'bg-[#713f12] text-[#fbbf24]',
  adelante: 'bg-[#713f12] text-[#fbbf24]',
  atras: 'bg-[#713f12] text-[#fbbf24]',
  divergida: 'bg-[#7c2d12] text-[#fdba74]',
  // falló (el check corrió — RN-13)
  rojo: 'bg-[#7f1d1d] text-[#fca5a5]',
  // vacío honesto (RN-02)
  'sin-gate': 'bg-[#27272a] text-[#a1a1aa]',
  'sin-tag': 'bg-[#27272a] text-[#a1a1aa]',
  'sin-upstream': 'bg-[#27272a] text-[#a1a1aa]',
  'sin-ledger': 'bg-[#27272a] text-[#a1a1aa]',
  'sin-board': 'bg-[#27272a] text-[#a1a1aa]',
  'sin-arquitectura': 'bg-[#27272a] text-[#a1a1aa]',
  // no se pudo medir ≠ rojo (RN-13) — borde punteado lo distingue del vacío
  'no-medido': 'bg-[#1e1b2e] text-[#a5b4fc] border border-dashed border-[#4c1d95]',
};

// acción sugerida por categoría (RN-18: veredictos accionables, cero gráficas)
const ACCION: Record<string, string> = {
  rojo: 'El check declarado corrió y FALLÓ — revisa el resumen, corrige la FUENTE (no el artefacto) y vuelve a medir.',
  'en-vuelo': 'Trabajo sin commitear. Si es de una sesión paralela, no barras esos paths en tu commit (commit selectivo).',
  adelante: 'Commits locales sin push — `git push` cuando el gate de fábrica esté verde.',
  atras: 'El upstream avanzó — `git pull --ff-only` antes de seguir trabajando.',
  divergida: 'Rama divergida del upstream — reconcilia (rebase o merge) antes de empujar.',
  'sin-upstream': 'La rama activa no trackea un upstream — ahead/behind no medible.',
  'sin-gate': 'El sistema no declara `gate_check` en el registro curado — decláralo para ganar el veredicto (RN-05).',
  'sin-tag': 'El repo no tiene tags alcanzables — la primera release taggeada enciende esta columna.',
  'sin-ledger': 'La célula no tiene espejo máquina — `gen_ledger.py` lo emite desde su LEDGER.md (RN-09).',
  'sin-board': 'Sin stories ni releases SDD en este proyecto.',
  'sin-arquitectura':
    'La célula no cura su arquitectura-como-dato — crea products/<célula>/arquitectura.yaml (formato CK-07/DH-06); el gate de fábrica lo valida (RN-24).',
};

function edad(medidoEn?: string): string {
  if (!medidoEn) return '';
  const ms = Date.now() - new Date(medidoEn).getTime();
  if (Number.isNaN(ms)) return '';
  const min = Math.floor(ms / 60000);
  if (min < 1) return 'hace <1 min';
  if (min < 60) return `hace ${min} min`;
  const h = Math.floor(min / 60);
  if (h < 24) return `hace ${h} h`;
  return `hace ${Math.floor(h / 24)} d`;
}

/** Resumen corto del nativo que acompaña a la categoría en la celda. */
function nativoCorto(col: string, v: TorreVeredicto): string {
  switch (col) {
    case 'rama': {
      const partes = [v.rama as string];
      if (v.categoria === 'adelante') partes.push(`↑${v.ahead}`);
      if (v.categoria === 'atras') partes.push(`↓${v.behind}`);
      if (v.categoria === 'divergida') partes.push(`↑${v.ahead}↓${v.behind}`);
      return partes.filter(Boolean).join(' ');
    }
    case 'en_vuelo':
      return v.categoria === 'en-vuelo' ? `${v.total} archivo(s)` : '';
    case 'ultimo_tag':
      return v.tag
        ? `${v.tag}${typeof v.commits_desde === 'number' && v.commits_desde > 0 ? ` +${v.commits_desde}` : ''}`
        : '';
    case 'gate_fabrica':
      return typeof v.exit_code === 'number' ? `exit ${v.exit_code}` : '';
    case 'ledger':
      return v.ultima_ficha ? `${(v.ultima_ficha as { id?: string }).id} · ${v.ultima_fecha ?? ''}` : '';
    case 'board': {
      if (v.categoria !== 'con-board') return '';
      const rel = v.release_en_curso as { release_id?: string } | null;
      return `${v.total_stories} stories${rel?.release_id ? ` · ${rel.release_id}` : ''}`;
    }
    case 'arquitectura':
      return v.categoria === 'con-arquitectura'
        ? `${v.clase} · ${v.total_componentes} comp · ${v.total_relaciones} rel`
        : '';
    default:
      return '';
  }
}

function CeldaVeredicto({
  col,
  v,
  onClick,
  activa,
}: {
  col: string;
  v: TorreVeredicto;
  onClick: () => void;
  activa: boolean;
}) {
  const corto = nativoCorto(col, v);
  return (
    <button
      onClick={onClick}
      title={`${v.categoria}${v.medido_en ? ` · medido ${edad(v.medido_en)}` : ''}`}
      className={cn(
        'inline-flex max-w-full items-center gap-1.5 rounded px-2 py-1 text-left font-mono text-[11px] transition-shadow',
        CAT_CLS[v.categoria] ?? 'bg-[#27272a] text-[#d4d4d8]',
        activa && 'ring-1 ring-[var(--color-accent,#2dd4bf)]'
      )}
    >
      <span className="font-semibold">{v.categoria}</span>
      {corto && <span className="truncate opacity-80">{corto}</span>}
    </button>
  );
}

/** Panel del objeto del dominio (RN-18): nativo completo + edad + acción sugerida. */
function DetalleVeredicto({ titulo, v }: { titulo: string; v: TorreVeredicto }) {
  const archivos = (v.archivos as { estado: string; path: string }[] | undefined) ?? [];
  const ficha = v.ultima_ficha as
    | { id?: string; titulo?: string; estado?: string; vigencia?: string }
    | undefined;
  const stories = v.stories as Record<string, number> | undefined;
  const release = v.release_en_curso as { release_id?: string; name?: string; status?: string } | null | undefined;
  const accion = v.categoria === 'no-medido' ? (v.motivo as string) : ACCION[v.categoria];

  return (
    <div className="space-y-3 border-t border-[#27272a] bg-[#101013] px-4 py-3 text-xs">
      <div className="flex items-baseline justify-between gap-4">
        <span className="font-semibold uppercase tracking-wide text-[#a1a1aa]">{titulo}</span>
        {v.medido_en && (
          <span className="text-[#71717a]">
            medido {edad(v.medido_en)} · <code className="text-[10px]">{v.medido_en}</code>
          </span>
        )}
      </div>

      {accion && <p className="text-[#d4d4d8]">{accion}</p>}

      {typeof v.comando === 'string' && (
        <p className="text-[#71717a]">
          comando: <code className="text-[#a1a1aa]">{v.comando}</code>
        </p>
      )}
      {typeof v.resumen === 'string' && v.resumen !== '' && (
        <pre className="max-h-64 overflow-auto rounded bg-[#18181b] p-3 font-mono text-[11px] leading-relaxed text-[#d4d4d8]">
          {v.resumen}
        </pre>
      )}

      {archivos.length > 0 && (
        <ul className="max-h-64 divide-y divide-[#1f1f23] overflow-auto rounded bg-[#18181b] font-mono text-[11px]">
          {archivos.map((a) => (
            <li key={a.path} className="flex gap-3 px-3 py-1.5">
              <code className="w-6 shrink-0 text-[#fbbf24]">{a.estado}</code>
              <span className="truncate text-[#d4d4d8]">{a.path}</span>
            </li>
          ))}
        </ul>
      )}

      {ficha?.id && (
        <div className="rounded bg-[#18181b] p-3">
          <span className="font-semibold text-[#93c5fd]">{ficha.id}</span>{' '}
          <span className="text-[#d4d4d8]">{ficha.titulo}</span>
          <span className="ml-2 text-[#71717a]">
            {ficha.estado} · {ficha.vigencia}
            {typeof v.total_fichas === 'number' && ` · ${v.total_fichas} fichas en total`}
            {typeof v.ultima_fecha === 'string' && ` · última actividad ${v.ultima_fecha}`}
          </span>
          {typeof v.path === 'string' && (
            <div className="mt-1 text-[10px] text-[#71717a]">
              fuente: <code>{v.path}</code>
            </div>
          )}
        </div>
      )}

      {stories && (
        <div className="flex flex-wrap gap-2">
          {Object.entries(stories).map(([estado, n]) => (
            <span key={estado} className="rounded bg-[#18181b] px-2 py-1 font-mono text-[11px] text-[#d4d4d8]">
              {estado}: {n}
            </span>
          ))}
          {release?.release_id && (
            <span className="rounded bg-[#1e3a5f] px-2 py-1 font-mono text-[11px] text-[#93c5fd]">
              release en curso: {release.release_id} · {release.status}
            </span>
          )}
        </div>
      )}

      {typeof v.rama === 'string' && (
        <p className="font-mono text-[11px] text-[#a1a1aa]">
          rama <span className="text-[#d4d4d8]">{v.rama}</span>
          {typeof v.upstream === 'string' && (
            <>
              {' → '}
              {v.upstream} · ahead {String(v.ahead)} · behind {String(v.behind)}
            </>
          )}
        </p>
      )}
      {typeof v.tag === 'string' && (
        <p className="font-mono text-[11px] text-[#a1a1aa]">
          {v.tag}
          {typeof v.fecha === 'string' && ` · taggeado ${v.fecha}`}
          {typeof v.commits_desde === 'number' && ` · ${v.commits_desde} commit(s) desde el tag`}
        </p>
      )}
    </div>
  );
}

// ── lente arquitectura (F3 · DH-06 · SPEC specs/torre-arquitectura.md) ────────

interface ArqPlano {
  id: string;
  nombre: string;
  sub?: string;
}
interface ArqComponente {
  id: string;
  tipo: string;
  plano?: string;
  nombre: string;
  proposito?: string;
  estado?: string;
  fichas?: string[];
  ruta?: string;
}
interface ArqRelacion {
  from: string;
  to: string;
  tipo: string;
}

/**
 * El renderizador DE DEVHUB para arquitectura-como-dato (I-75: cada producto el suyo).
 * Discrimina por `clase` (RN-22): `modelo` → tabla de componentes agrupada por PLANO (RN-26);
 * clase desconocida → lista plana degradada, la fila jamás rompe.
 */
function DetalleArquitectura({ titulo, v }: { titulo: string; v: TorreVeredicto }) {
  const [compSel, setCompSel] = useState<string | null>(null);

  const clase = typeof v.clase === 'string' ? v.clase : 'modelo';
  const planos = (v.planos as ArqPlano[] | undefined) ?? [];
  const componentes = (v.componentes as ArqComponente[] | undefined) ?? [];
  const relaciones = (v.relaciones as ArqRelacion[] | undefined) ?? [];
  const conRenderPorPlano = clase === 'modelo' && planos.length > 0;

  // agrupación por plano (RN-26); componente con plano no declarado cae a '—' (degradado, no rompe)
  const grupos: { plano: ArqPlano; comps: ArqComponente[] }[] = conRenderPorPlano
    ? planos
        .map((plano) => ({ plano, comps: componentes.filter((c) => c.plano === plano.id) }))
        .filter((g) => g.comps.length > 0)
    : [{ plano: { id: '—', nombre: clase === 'modelo' ? 'Componentes' : `clase «${clase}» — render degradado (RN-22)` }, comps: componentes }];

  const sel = componentes.find((c) => c.id === compSel);
  const relacionesDe = (id: string) => relaciones.filter((r) => r.from === id || r.to === id);

  return (
    <div className="space-y-3 border-t border-[#27272a] bg-[#101013] px-4 py-3 text-xs">
      <div className="flex items-baseline justify-between gap-4">
        <span className="font-semibold uppercase tracking-wide text-[#a1a1aa]">{titulo}</span>
        {v.medido_en && (
          <span className="text-[#71717a]">
            medido {edad(v.medido_en as string)} · <code className="text-[10px]">{v.medido_en as string}</code>
          </span>
        )}
      </div>

      <div className="flex flex-wrap items-baseline gap-2">
        <span className="font-medium text-[#e4e4e7]">{(v.nombre as string) ?? (v.id as string)}</span>
        <span className="rounded bg-[#1e3a5f] px-1.5 py-0.5 font-mono text-[10px] text-[#93c5fd]">clase: {clase}</span>
        {typeof v.version !== 'undefined' && (
          <span className="font-mono text-[10px] text-[#71717a]">v{String(v.version)}</span>
        )}
        <span className="font-mono text-[10px] text-[#71717a]">
          {String(v.total_componentes)} componentes · {String(v.total_relaciones)} relaciones
        </span>
      </div>
      {typeof v.proposito === 'string' && <p className="max-w-3xl text-[#a1a1aa]">{v.proposito}</p>}

      {grupos.map(({ plano, comps }) => (
        <div key={plano.id} className="overflow-hidden rounded bg-[#18181b]">
          <div className="border-b border-[#27272a] px-3 py-1.5">
            <span className="font-semibold text-[#d4d4d8]">{plano.nombre}</span>
            {plano.sub && <span className="ml-2 text-[10px] text-[#71717a]">{plano.sub}</span>}
          </div>
          <ul className="divide-y divide-[#1f1f23]">
            {comps.map((c) => (
              <li key={c.id}>
                <button
                  onClick={() => setCompSel((s) => (s === c.id ? null : c.id))}
                  className={cn(
                    'flex w-full items-center gap-3 px-3 py-1.5 text-left hover:bg-[#1f1f23]',
                    compSel === c.id && 'bg-[#1f1f23] ring-1 ring-inset ring-[var(--color-accent,#2dd4bf)]'
                  )}
                >
                  <span className="w-40 shrink-0 truncate font-mono text-[11px] text-[#d4d4d8]">{c.nombre}</span>
                  <span className="rounded bg-[#27272a] px-1.5 py-0.5 font-mono text-[10px] text-[#a1a1aa]">{c.tipo}</span>
                  {c.estado && c.estado !== 'activo' && (
                    <span className="rounded bg-[#713f12] px-1.5 py-0.5 font-mono text-[10px] text-[#fbbf24]">{c.estado}</span>
                  )}
                  <span className="truncate text-[11px] text-[#71717a]">{c.proposito}</span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      ))}

      {sel && (
        <div className="space-y-2 rounded bg-[#18181b] p-3">
          <div className="flex flex-wrap items-baseline gap-2">
            <span className="font-semibold text-[#93c5fd]">{sel.nombre}</span>
            <code className="text-[10px] text-[#71717a]">{sel.id}</code>
            <span className="rounded bg-[#27272a] px-1.5 py-0.5 font-mono text-[10px] text-[#a1a1aa]">{sel.tipo}</span>
            {sel.estado && (
              <span className="font-mono text-[10px] text-[#71717a]">estado: {sel.estado}</span>
            )}
          </div>
          {sel.proposito && <p className="text-[#d4d4d8]">{sel.proposito}</p>}
          {(sel.fichas?.length ?? 0) > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {sel.fichas!.map((f) => (
                <span key={f} className="rounded bg-[#1e3a5f] px-1.5 py-0.5 font-mono text-[10px] text-[#93c5fd]">
                  {f}
                </span>
              ))}
            </div>
          )}
          {sel.ruta && (
            <p className="text-[10px] text-[#71717a]">
              ruta: <code>{sel.ruta}</code>
            </p>
          )}
          {relacionesDe(sel.id).length > 0 && (
            <ul className="space-y-0.5 font-mono text-[11px] text-[#a1a1aa]">
              {relacionesDe(sel.id).map((r, i) => (
                <li key={i}>
                  {r.from === sel.id ? (
                    <>
                      <span className="text-[#2dd4bf]">{r.tipo}</span> → {r.to}
                    </>
                  ) : (
                    <>
                      {r.from} → <span className="text-[#2dd4bf]">{r.tipo}</span> (este)
                    </>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {typeof v.path === 'string' && (
        <p className="text-[10px] text-[#71717a]">
          fuente: <code>{v.path}</code> — el YAML gated ES el contrato (RN-21); read-only.
        </p>
      )}
    </div>
  );
}

const COLS: { key: 'rama' | 'en_vuelo' | 'ultimo_tag' | 'gate_fabrica'; label: string }[] = [
  { key: 'gate_fabrica', label: 'gate de fábrica' },
  { key: 'en_vuelo', label: 'en vuelo' },
  { key: 'rama', label: 'rama' },
  { key: 'ultimo_tag', label: 'último tag' },
];

function FilaSistema({ s }: { s: TorreSistema }) {
  const [expandida, setExpandida] = useState(false);
  const [detalle, setDetalle] = useState<string | null>(null); // "repo:col" | "proy:slug:col"

  const toggleDetalle = (key: string) => setDetalle((d) => (d === key ? null : key));
  const detalleRepo = detalle?.startsWith('repo:')
    ? COLS.find((c) => `repo:${c.key}` === detalle)
    : null;

  return (
    <>
      <tr className="border-t border-[#27272a]">
        <td className="px-4 py-3 align-top">
          <button
            onClick={() => setExpandida((e) => !e)}
            className="flex items-start gap-2 text-left"
            title={s.workspace || 'sin workspace'}
          >
            {expandida ? (
              <ChevronDown className="mt-0.5 h-4 w-4 shrink-0 text-[#71717a]" />
            ) : (
              <ChevronRight className="mt-0.5 h-4 w-4 shrink-0 text-[#71717a]" />
            )}
            <span>
              <span className="block text-sm font-medium text-[#e4e4e7]">{s.slug}</span>
              <span className="block text-[11px] text-[#71717a]">{s.nombre}</span>
            </span>
          </button>
        </td>
        {COLS.map((c) => (
          <td key={c.key} className="px-3 py-3 align-top">
            <CeldaVeredicto
              col={c.key}
              v={s.repo[c.key]}
              activa={detalle === `repo:${c.key}`}
              onClick={() => toggleDetalle(`repo:${c.key}`)}
            />
          </td>
        ))}
        <td className="px-3 py-3 text-right align-top font-mono text-[11px] text-[#71717a]">
          {s.proyectos.length > 0 ? `${s.proyectos.length} proyecto(s)` : '—'}
        </td>
      </tr>

      {detalleRepo && (
        <tr>
          <td colSpan={COLS.length + 2} className="p-0">
            <DetalleVeredicto
              titulo={`${s.slug} · ${detalleRepo.label}`}
              v={s.repo[detalleRepo.key]}
            />
          </td>
        </tr>
      )}

      {expandida && (
        <tr>
          <td colSpan={COLS.length + 2} className="bg-[#101013] p-0">
            {s.proyectos.length === 0 ? (
              <p className="px-10 py-3 text-xs text-[#71717a]">
                Sin proyectos descubiertos — el eje PROYECTO es nullable con empty-state honesto
                (RN-02): este workspace no tiene células ni board SDD por convención.
              </p>
            ) : (
              <table className="w-full">
                <tbody>
                  {s.proyectos.map((p) => (
                    <FragmentoProyecto
                      key={p.slug}
                      p={p}
                      detalle={detalle}
                      onDetalle={toggleDetalle}
                    />
                  ))}
                </tbody>
              </table>
            )}
          </td>
        </tr>
      )}
    </>
  );
}

function FragmentoProyecto({
  p,
  detalle,
  onDetalle,
}: {
  p: TorreProyecto;
  detalle: string | null;
  onDetalle: (k: string) => void;
}) {
  const keyL = `proy:${p.slug}:ledger`;
  const keyB = `proy:${p.slug}:board`;
  const keyA = `proy:${p.slug}:arquitectura`;
  return (
    <>
      <tr className="border-t border-[#1f1f23]">
        <td className="w-64 py-2 pl-10 pr-3">
          <span className="font-mono text-xs text-[#d4d4d8]">{p.slug}</span>
          <span className="ml-2 rounded bg-[#27272a] px-1.5 py-0.5 text-[10px] text-[#a1a1aa]">
            {p.tipo}
          </span>
        </td>
        <td className="px-3 py-2">
          <CeldaVeredicto col="ledger" v={p.ledger} activa={detalle === keyL} onClick={() => onDetalle(keyL)} />
        </td>
        <td className="px-3 py-2">
          <CeldaVeredicto col="board" v={p.board} activa={detalle === keyB} onClick={() => onDetalle(keyB)} />
        </td>
        <td className="px-3 py-2">
          <CeldaVeredicto
            col="arquitectura"
            v={p.arquitectura}
            activa={detalle === keyA}
            onClick={() => onDetalle(keyA)}
          />
        </td>
        <td />
      </tr>
      {detalle === keyL && (
        <tr>
          <td colSpan={5} className="p-0">
            <DetalleVeredicto titulo={`${p.slug} · ledger`} v={p.ledger} />
          </td>
        </tr>
      )}
      {detalle === keyB && (
        <tr>
          <td colSpan={5} className="p-0">
            <DetalleVeredicto titulo={`${p.slug} · board`} v={p.board} />
          </td>
        </tr>
      )}
      {detalle === keyA && (
        <tr>
          <td colSpan={5} className="p-0">
            {p.arquitectura.categoria === 'con-arquitectura' ? (
              <DetalleArquitectura titulo={`${p.slug} · arquitectura`} v={p.arquitectura} />
            ) : (
              <DetalleVeredicto titulo={`${p.slug} · arquitectura`} v={p.arquitectura} />
            )}
          </td>
        </tr>
      )}
    </>
  );
}

export function TorreView() {
  const [data, setData] = useState<TorreResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [midiendo, setMidiendo] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback((medir?: 'gate_fabrica') => {
    if (medir) setMidiendo(true);
    else setLoading(true);
    setError(null);
    getTorre(medir)
      .then(setData)
      .catch((err) => setError((err as Error).message))
      .finally(() => {
        setLoading(false);
        setMidiendo(false);
      });
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) return <div className="p-8"><Spinner /></div>;
  if (error) return <div className="p-6"><ErrorBanner message={error} /></div>;
  if (!data || data.sistemas.length === 0) {
    return (
      <div className="p-6">
        <EmptyState>
          El registro de sistemas está vacío. Genera el registro con{' '}
          <code>products/devhub/scripts/gen_registro.py</code> (curado × config del operador —
          SPEC torre-read-only §2).
        </EmptyState>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl space-y-4 p-6">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold text-[#e4e4e7]">Torre de control</h1>
          <p className="mt-1 text-xs text-[#a1a1aa]">
            {data.sistemas.length} sistema(s) del registro · read-only · generado {edad(data.generado_en)}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => load()}
            className="inline-flex items-center gap-1.5 rounded border border-[#27272a] px-3 py-1.5 text-xs text-[#d4d4d8] hover:bg-[#27272a]"
            title="Re-lee los hechos baratos (git · YAML — cache ≤5s, RN-16)"
          >
            <RefreshCw className="h-3.5 w-3.5" /> Refrescar
          </button>
          <button
            onClick={() => load('gate_fabrica')}
            disabled={midiendo}
            className="inline-flex items-center gap-1.5 rounded border border-[#27272a] px-3 py-1.5 text-xs text-[#fbbf24] hover:bg-[#27272a] disabled:opacity-50"
            title="Re-mide el gate de fábrica de cada sistema (caro; TTL 10 min — RN-16)"
          >
            <Gauge className={cn('h-3.5 w-3.5', midiendo && 'animate-spin')} />
            {midiendo ? 'Midiendo gate de fábrica…' : 'Medir gate de fábrica'}
          </button>
        </div>
      </header>

      <Card className="overflow-x-auto p-0">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-[10px] uppercase tracking-wider text-[#71717a]">
              <th className="px-4 py-2 font-medium">Sistema</th>
              {COLS.map((c) => (
                <th key={c.key} className="px-3 py-2 font-medium">{c.label}</th>
              ))}
              <th className="px-3 py-2 text-right font-medium">Eje proyecto</th>
            </tr>
          </thead>
          <tbody>
            {data.sistemas.map((s) => (
              <FilaSistema key={s.slug} s={s} />
            ))}
          </tbody>
        </table>
      </Card>

      <p className="text-[11px] text-[#71717a]">
        Veredictos = categorías fijas + dato nativo al lado; <code>no-medido</code> ≠{' '}
        <code>rojo</code> (RN-13). Click en una celda abre el objeto del dominio (RN-18).
        La torre no escribe sobre los repos monitoreados ni hace <code>git fetch</code>.
      </p>
    </div>
  );
}
