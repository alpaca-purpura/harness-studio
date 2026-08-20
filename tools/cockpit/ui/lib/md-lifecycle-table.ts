/**
 * Parser genérico de tablas markdown con lifecycle de estados.
 * ────────────────────────────────────────────────────────────────────────────
 * Varios registros del CIL son la MISMA forma de tabla — id en col 0, fecha,
 * severidad emoji, texto en el medio, estado (con negrita/sufijo) en penúltima,
 * ref en última — con el mismo lifecycle de estados (reported→…→verified +
 * deferred). En vez de duplicar el parser por archivo (harness-backlog L1,
 * tech-debt L3), ambos consumen estos helpers (anti-duplication.md aplicado a
 * tooling: una sola mecánica de parseo, mapeos por-registro arriba).
 *
 * Puro · sin I/O (la lectura de archivo la hacen las API routes).
 */

/** Estados del lifecycle HLP/CIL + `deferred` (lateral). `otro` = no reconocido. */
export type LifecycleEstado =
  | 'reported'
  | 'triaged'
  | 'ratified'
  | 'applied'
  | 'verified'
  | 'deferred'
  | 'otro';

/** Orden canónico de columnas (lifecycle + deferred al final). */
export const LIFECYCLE_ORDER: Exclude<LifecycleEstado, 'otro'>[] = [
  'reported',
  'triaged',
  'ratified',
  'applied',
  'verified',
  'deferred',
];

const VALID_ESTADOS = new Set<LifecycleEstado>(LIFECYCLE_ORDER);

/** Quita marcadores markdown de énfasis (`**`, `__`, `*`, `` ` ``) de un texto. */
export function stripEmphasis(s: string): string {
  return s.replace(/\*\*|__|\*|`/g, '').trim();
}

/**
 * Deriva el estado canónico del campo "estado".
 * Ej: "**applied** (cont. 4)" → "applied" · "reported" → "reported".
 * Desconocido → "otro".
 */
export function normalizeEstado(raw: string): LifecycleEstado {
  const firstWord = stripEmphasis(raw).toLowerCase().match(/[a-záéíóúñ]+/)?.[0] ?? '';
  return VALID_ESTADOS.has(firstWord as LifecycleEstado)
    ? (firstWord as LifecycleEstado)
    : 'otro';
}

/**
 * Parte una fila de tabla markdown en celdas trimmeadas, descartando los vacíos
 * de borde. `null` si la línea no es una fila de tabla (no empieza con `|`).
 */
export function splitTableRow(line: string): string[] | null {
  const trimmed = line.trim();
  if (!trimmed.startsWith('|')) return null;
  const cells = trimmed.split('|').map((c) => c.trim());
  if (cells.length >= 2 && cells[0] === '') cells.shift();
  if (cells.length >= 1 && cells[cells.length - 1] === '') cells.pop();
  return cells;
}

/**
 * Devuelve las filas (arrays de celdas) cuyo primer campo matchea `idRe` y que
 * tienen al menos 6 columnas (id, fecha, sev, texto, estado, ref). Header,
 * separador `---` y prosa quedan fuera. Pipes accidentales dentro del texto se
 * preservan en las celdas del medio (el caller los reabsorbe con `slice(3, -2)`).
 */
export function tableRows(md: string, idRe: RegExp): string[][] {
  const rows: string[][] = [];
  if (!md) return rows;
  for (const line of md.split('\n')) {
    const cells = splitTableRow(line);
    if (!cells) continue;
    if (cells.length < 6) continue;
    if (!idRe.test(cells[0])) continue;
    rows.push(cells);
  }
  return rows;
}

/** Cuenta items por estado canónico (incluye `otro` si aparece). */
export function countByEstadoGeneric<T extends { estado: LifecycleEstado }>(
  items: T[]
): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const it of items) {
    counts[it.estado] = (counts[it.estado] ?? 0) + 1;
  }
  return counts;
}

/** Estados "abiertos" (aún requieren trabajo) = todo menos verified/deferred. */
export const OPEN_ESTADOS: ReadonlySet<LifecycleEstado> = new Set<LifecycleEstado>([
  'reported',
  'triaged',
  'ratified',
  'applied',
]);

/** Cuenta items en estado abierto (lo que el cockpit badge-ea como "pendiente"). */
export function countOpen<T extends { estado: LifecycleEstado }>(items: T[]): number {
  return items.filter((it) => OPEN_ESTADOS.has(it.estado)).length;
}
