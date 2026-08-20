'use client';

import { Card } from '@/components/ui/Card';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import type { ComputedStatusReport } from '@/lib/types';

/**
 * Salud de Producto · banner honesto en el tope del Mapa.
 *
 * Lee el `summary` de _status-computed.json (vía getCapabilityStatus) — NUNCA
 * inventa números. Muestra la distribución real de las capabilities por
 * computed_status para que el operador vea el bosque de un vistazo.
 *
 * Doctrina: lifecycle.md § 4 (salud de capability) + § 7 (cockpit = ver el bosque).
 */

type SegmentKey =
  | 'verified_live'
  | 'declared_live'
  | 'partial'
  | 'wip'
  | 'stub'
  | 'drift'
  | 'deprecated'
  | 'sunset';

interface SegmentDef {
  key: SegmentKey;
  label: string;
  emoji: string;
  /** clase para la barra (fondo) */
  bar: string;
  /** clase para el chip (fondo + texto) */
  chip: string;
  tip: string;
}

const SEGMENTS: SegmentDef[] = [
  {
    key: 'verified_live',
    label: 'verificado',
    emoji: '🟢',
    bar: 'bg-[#16a34a]',
    chip: 'bg-[#14532d] text-[#86efac]',
    tip: TOOLTIPS.verified_live,
  },
  {
    key: 'declared_live',
    label: 'declarado',
    emoji: '🔵',
    bar: 'bg-[#2563eb]',
    chip: 'bg-[#1e3a5f] text-[#93c5fd]',
    tip: TOOLTIPS.declared_live,
  },
  {
    key: 'partial',
    label: 'parcial',
    emoji: '🟠',
    bar: 'bg-[#d97706]',
    chip: 'bg-[#713f12] text-[#fbbf24]',
    tip: TOOLTIPS.partial,
  },
  {
    key: 'wip',
    label: 'wip',
    emoji: '🔵',
    bar: 'bg-[#0891b2]',
    chip: 'bg-[#164e63] text-[#67e8f9]',
    tip: TOOLTIPS.wip,
  },
  {
    key: 'stub',
    label: 'stub',
    emoji: '⚪',
    bar: 'bg-[#52525b]',
    chip: 'bg-[#3f3f46] text-[#d4d4d8]',
    tip: TOOLTIPS.stub,
  },
  {
    key: 'drift',
    label: 'drift',
    emoji: '🔴',
    bar: 'bg-[#dc2626]',
    chip: 'bg-[#450a0a] text-[#fca5a5]',
    tip: TOOLTIPS.drift,
  },
  {
    key: 'deprecated',
    label: 'deprecated',
    emoji: '⚫',
    bar: 'bg-[#3f3f46]',
    chip: 'bg-[#3f3f46] text-[#a1a1aa]',
    tip: TOOLTIPS.drift,
  },
  {
    key: 'sunset',
    label: 'sunset',
    emoji: '⚫',
    bar: 'bg-[#27272a]',
    chip: 'bg-[#27272a] text-[#a1a1aa]',
    tip: TOOLTIPS.drift,
  },
];

export function ProductHealthBanner({
  report,
}: {
  report: ComputedStatusReport | null;
}) {
  if (!report) return null;
  const s = report.summary;
  const total = s.total_caps;
  if (total <= 0) return null;

  // Solo segmentos con count > 0, en el orden de salud (verde → rojo).
  const present = SEGMENTS.map((seg) => ({ seg, count: s[seg.key] ?? 0 })).filter(
    (x) => x.count > 0
  );

  // Caption honesta: total · stub · drift (las verdades incómodas primero).
  const captionParts: string[] = [`${total} caps`];
  if (s.stub > 0) captionParts.push(`${s.stub} sin scenarios (stub)`);
  if (s.partial > 0) captionParts.push(`${s.partial} parcial`);
  if (s.drift > 0) captionParts.push(`${s.drift} en drift`);
  const live = s.verified_live + s.declared_live;
  if (live > 0) captionParts.push(`${live} live`);
  const caption = captionParts.join(' · ');

  return (
    <Card className="!p-4 mb-4">
      <header className="flex items-baseline justify-between gap-3 flex-wrap mb-2">
        <h2 className="text-sm font-semibold">
          <Tooltip content={TOOLTIPS.product_health} variant="header">
            Salud de Producto
          </Tooltip>
        </h2>
        <span className="text-[11px] text-[var(--color-muted)] font-mono">
          {caption}
        </span>
      </header>

      {/* Barra de distribución proporcional */}
      <div
        className="flex h-2.5 rounded-full overflow-hidden bg-[var(--color-panel)] border border-[var(--color-border)]"
        role="img"
        aria-label={`Distribución de capabilities: ${caption}`}
      >
        {present.map(({ seg, count }) => (
          <Tooltip key={seg.key} content={seg.tip}>
            <div
              className={`${seg.bar} h-full`}
              style={{ width: `${(count / total) * 100}%` }}
              title={`${seg.label}: ${count}`}
            />
          </Tooltip>
        ))}
      </div>

      {/* Chips con count por estado */}
      <div className="flex flex-wrap gap-1.5 mt-3 text-[11px]">
        {present.map(({ seg, count }) => (
          <Tooltip key={seg.key} content={seg.tip} variant="badge">
            <span
              className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded font-medium ${seg.chip}`}
            >
              <span aria-hidden="true">{seg.emoji}</span>
              {seg.label}
              <span className="font-mono opacity-80">{count}</span>
            </span>
          </Tooltip>
        ))}
      </div>
    </Card>
  );
}
