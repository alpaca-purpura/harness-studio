/**
 * CapLevel · contenedor colapsable de un NIVEL del cap-drawer (N0-N4).
 *
 * Modelo de niveles (cockpit-capability-levels-proposal.md · ratificado 2026-06-06,
 * Diátaxis-altitud + caso-de-uso-RUP):
 *   N0 · Qué es            → user_facing_name + description (no es CapLevel: es el intro)
 *   N1 · Qué puedo hacer   → scenarios (casos de uso + badge de verdad)
 *   N2 · Bajo qué reglas   → business_rules (+ badge enforcement)
 *   N3 · Quién y por dónde → access.entry_points
 *   N4 · Dónde vive        → code files + related_capabilities + bidireccional
 *
 * Colapsable nativo (`<details>`) — cero dependencia, accesible, RSC-safe. N1/N2 abren
 * por defecto (lo que el operador más mira); N3/N4 colapsan (referencia técnica bajo demanda).
 */

import type { ReactNode } from 'react';

export function CapLevel({
  level,
  title,
  hint,
  defaultOpen = true,
  children,
}: {
  level: string;
  title: string;
  hint?: string;
  defaultOpen?: boolean;
  children: ReactNode;
}) {
  return (
    <details
      open={defaultOpen}
      className="group border border-[var(--color-border)] rounded-lg bg-[var(--color-panel)]/40 overflow-hidden"
    >
      <summary className="cursor-pointer select-none list-none [&::-webkit-details-marker]:hidden flex items-center gap-2 px-3 py-2 hover:bg-[var(--color-panel2)]/50">
        <span className="text-[9px] font-mono font-bold text-[var(--color-accent)] bg-[var(--color-panel2)] px-1.5 py-0.5 rounded">
          {level}
        </span>
        <span className="text-sm font-semibold">{title}</span>
        {hint && <span className="text-[10px] text-[var(--color-muted)] truncate">{hint}</span>}
        <span
          aria-hidden="true"
          className="ml-auto text-[var(--color-muted)] text-xs transition-transform group-open:rotate-90"
        >
          ▸
        </span>
      </summary>
      <div className="px-3 pb-3 pt-1">{children}</div>
    </details>
  );
}
