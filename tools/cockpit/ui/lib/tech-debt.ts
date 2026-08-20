/**
 * Parser del Tech-Debt Register · docs/process/tech-debt.md (carril L3 del CIL)
 * ────────────────────────────────────────────────────────────────────────────
 * Deuda de código/infra pura (no proceso → L1 harness · no producto/arq → L2
 * learnings · no cap stale → L4 auto-detect). dev-team + /auditor appendean,
 * el PM de plataforma homologa en el stop de mejora continua.
 *
 * Misma forma de tabla que el harness-backlog (id/fecha/sev/texto/estado/ref) →
 * comparte la mecánica de parseo de `md-lifecycle-table.ts`. Acá solo el mapeo
 * de severidad propia (🔴 bloquea-pronto · 🟡 fricción · 🔵 mejora).
 *
 * Formato de fila esperado:
 *   | TD-1 | 2026-06-05 | 🔴 | deuda (problema → causa → refuerzo) | **applied** | abc123 |
 */

import {
  LIFECYCLE_ORDER,
  countByEstadoGeneric,
  normalizeEstado,
  stripEmphasis,
  tableRows,
  type LifecycleEstado,
} from './md-lifecycle-table';

export type TechDebtEstado = LifecycleEstado;

/** Orden de columnas del board L3 (mismo lifecycle compartido). */
export const TD_ESTADO_ORDER: Exclude<TechDebtEstado, 'otro'>[] = LIFECYCLE_ORDER;

export type TechDebtSeveridad = 'bloquea-pronto' | 'friccion' | 'mejora' | 'otro';

export interface TechDebtItem {
  /** Id literal, ej. "TD-1". */
  id: string;
  /** Número para ordenar. */
  num: number;
  /** Fecha de captura (YYYY-MM-DD). */
  fecha: string;
  /** Emoji de severidad crudo. */
  sevEmoji: string;
  /** Severidad legible derivada del emoji. */
  sevLabel: TechDebtSeveridad;
  /** Texto crudo de la deuda (puede traer markdown). */
  item: string;
  /** Texto sin negrita · para el título de la card. */
  title: string;
  /** Estado canónico. */
  estado: TechDebtEstado;
  /** Campo estado crudo (preserva matiz). */
  estadoRaw: string;
  /** Campo ref crudo (commit SHAs, notas). */
  ref: string;
}

const SEV_MAP: Record<string, TechDebtSeveridad> = {
  '🔴': 'bloquea-pronto',
  '🟡': 'friccion',
  '🔵': 'mejora',
};

/**
 * Parsea el contenido markdown del tech-debt register a TechDebtItem[].
 * Puro · sin I/O. Solo filas `TD-N` (header, separador, placeholder `—` y prosa
 * se ignoran — el registro arranca con una fila placeholder que NO matchea).
 */
export function parseTechDebt(md: string): TechDebtItem[] {
  return tableRows(md, /^TD-\d+$/).map((cells) => {
    const [id, fecha, sevEmoji] = cells;
    const ref = cells[cells.length - 1];
    const estadoRaw = cells[cells.length - 2];
    const item = cells.slice(3, cells.length - 2).join(' | ').trim();

    return {
      id,
      num: Number.parseInt(id.replace('TD-', ''), 10),
      fecha,
      sevEmoji,
      sevLabel: SEV_MAP[sevEmoji] ?? 'otro',
      item,
      title: stripEmphasis(item),
      estado: normalizeEstado(estadoRaw),
      estadoRaw,
      ref,
    };
  });
}

/** Cuenta items por estado canónico (incluye `otro` si aparece). */
export function countTdByEstado(items: TechDebtItem[]): Record<string, number> {
  return countByEstadoGeneric(items);
}
