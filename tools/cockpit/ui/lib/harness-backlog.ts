/**
 * Parser del Harness Backlog · docs/process/harness-backlog.md (carril L1 del CIL)
 * ────────────────────────────────────────────────────────────────────────────
 * El harness-backlog es el issue-tracker liviano del harness (skills/rules/hooks/
 * agents/cockpit/templates). Vive como una tabla markdown — este parser la
 * convierte a HarnessItem[] para que el cockpit la visualice como board read-only.
 *
 * El `.md` SIGUE SIENDO el SSoT (captura sin fricción vía /harness-issue). El
 * cockpit NO escribe acá — solo lee. La mecánica de parseo de tabla vive en
 * `md-lifecycle-table.ts` (compartida con tech-debt L3); acá queda solo el mapeo
 * específico de HB (severidad 4-emoji + carril CIL).
 *
 * Formato de fila esperado:
 *   | HB-1 | 2026-06-01 | 🔴 | item texto (markdown) | **verified** | 17bf3c62 |
 */

import {
  LIFECYCLE_ORDER,
  countByEstadoGeneric,
  normalizeEstado,
  stripEmphasis,
  tableRows,
  type LifecycleEstado,
} from './md-lifecycle-table';

/** Estados del lifecycle HLP + `deferred`. Alias del lifecycle compartido. */
export type HarnessEstado = LifecycleEstado;

/** Orden canónico de columnas en el board (lifecycle + deferred al final). */
export const ESTADO_ORDER: Exclude<HarnessEstado, 'otro'>[] = LIFECYCLE_ORDER;

export type HarnessSeveridad = 'silent-killer' | 'quick-win' | 'decision' | 'wave' | 'otro';

/**
 * Carril del CIL (proceso v5 §5.7 · docs/process/continuous-improvement.md). Este
 * backlog ES el carril L1 (harness/proceso); un ítem puede taggear `[L2]`/`[L3]`/`[L4]`
 * en su texto si excepcionalmente pertenece a otro carril. Default L1. Dimensión
 * ORTOGONAL a la severidad (additive · no reemplaza nada — D-C).
 */
export type HarnessCarril = 'L1' | 'L2' | 'L3' | 'L4';

export const CARRIL_LABELS: Record<HarnessCarril, string> = {
  L1: 'harness',
  L2: 'producto/skills-arq',
  L3: 'deuda técnica',
  L4: 'capability-desfasada',
};

export interface HarnessItem {
  /** Id literal de la tabla, ej. "HB-1". */
  id: string;
  /** Número para ordenar/desempatar, ej. 1. */
  num: number;
  /** Fecha de captura (YYYY-MM-DD), tal cual la tabla. */
  fecha: string;
  /** Emoji de severidad crudo de la tabla. */
  sevEmoji: string;
  /** Severidad legible derivada del emoji. */
  sevLabel: HarnessSeveridad;
  /** Texto del item tal cual (puede traer markdown **negrita**). */
  item: string;
  /** Item con marcadores de negrita removidos · para el título de la card. */
  title: string;
  /** Carril del CIL (default L1 = harness; `[Ln]` tag en el texto lo overridea). */
  carril: HarnessCarril;
  /** Estado canónico (primera palabra del campo estado, sin negrita ni sufijo). */
  estado: HarnessEstado;
  /** Campo estado crudo, ej. "**applied** (cont. 4)" · preserva el matiz. */
  estadoRaw: string;
  /** Campo ref crudo (commit SHAs, notas). */
  ref: string;
}

const SEV_MAP: Record<string, HarnessSeveridad> = {
  '🔴': 'silent-killer',
  '🟡': 'quick-win',
  '🔵': 'decision',
  '🟣': 'wave',
};

/** Detecta un tag `[L2]`/`[L3]`/`[L4]` en el texto del item. Default L1 (backlog = carril L1). */
const CARRIL_RE = /\[(L[234])\]/i;

function deriveCarril(item: string): HarnessCarril {
  const m = item.match(CARRIL_RE);
  return (m?.[1]?.toUpperCase() as HarnessCarril) ?? 'L1';
}

/**
 * Parsea el contenido markdown del harness-backlog a HarnessItem[].
 * Puro · sin I/O. Filas no-HB (header, separador, prosa) se ignoran.
 */
export function parseHarnessBacklog(md: string): HarnessItem[] {
  return tableRows(md, /^HB-\d+$/).map((cells) => {
    const [id, fecha, sevEmoji] = cells;
    const ref = cells[cells.length - 1];
    const estadoRaw = cells[cells.length - 2];
    // El item es todo lo que quede en el medio (reabsorbe pipes accidentales).
    const item = cells.slice(3, cells.length - 2).join(' | ').trim();

    return {
      id,
      num: Number.parseInt(id.replace('HB-', ''), 10),
      fecha,
      sevEmoji,
      sevLabel: SEV_MAP[sevEmoji] ?? 'otro',
      item,
      title: stripEmphasis(item),
      carril: deriveCarril(item),
      estado: normalizeEstado(estadoRaw),
      estadoRaw,
      ref,
    };
  });
}

/** Cuenta items por estado canónico (incluye `otro` si aparece). */
export function countByEstado(items: HarnessItem[]): Record<string, number> {
  return countByEstadoGeneric(items);
}

/** Cuenta items por carril del CIL (L1-L4) · para el board 4-lanes del stop semanal. */
export function countByCarril(items: HarnessItem[]): Record<HarnessCarril, number> {
  const counts: Record<HarnessCarril, number> = { L1: 0, L2: 0, L3: 0, L4: 0 };
  for (const it of items) counts[it.carril] += 1;
  return counts;
}
