/**
 * Helpers para la vista Drift — filtrado, ordenamiento y next-action.
 *
 * Extraído de DriftView para facilitar tests unitarios puros (sin DOM).
 */

import type { ComputedStatus, ComputedStatusReport, CapStatusComputed } from './types';

// ────────────────────────────────────────────────────────────────────────────
// Tipos
// ────────────────────────────────────────────────────────────────────────────

export interface DriftEntry {
  slug: string;
  /** Área funcional declarada en el YAML (ej: "lisa.servicios") */
  functional_area: string | null;
  computed_status: ComputedStatus;
  scenarios_total: number;
  drift_reasons: string[];
  /** Texto one-liner para mostrar en la fila de resumen */
  summary: string;
  /** Acción sugerida según el computed_status */
  next_action: string;
}

// ────────────────────────────────────────────────────────────────────────────
// Severidad para ordenamiento
// ────────────────────────────────────────────────────────────────────────────

const SEVERITY_ORDER: Record<ComputedStatus, number> = {
  'drift': 0,
  'partial': 1,
  'declared-live': 2,
  'wip': 3,
  'stub': 4,
  'deprecated': 5,
  'sunset': 6,
  'verified-live': 99, // no debería llegar aquí (filtrado antes)
};

export function getSeverityOrder(s: ComputedStatus): number {
  return SEVERITY_ORDER[s] ?? 99;
}

// ────────────────────────────────────────────────────────────────────────────
// Next-action por status
// ────────────────────────────────────────────────────────────────────────────

export function getNextAction(s: ComputedStatus): string {
  switch (s) {
    case 'stub':
      return 'Poblar scenarios[] vía Fase F.3 (cap_change_type=extend) o documentar como legacy bootstrap';
    case 'declared-live':
      return 'Agregar e2e_test a los scenarios para promover a verified-live';
    case 'partial':
      return 'Completar scenarios verificados (algunos siguen wip)';
    case 'wip':
      return 'Cap en desarrollo, espera implementación o transition status:beta→live';
    case 'drift':
      return 'URGENTE: scenario declarado live pero su e2e_test no existe. Ejecutar validate_code_cap_bidirectional.py para detalles, o restaurar el path declarado';
    case 'deprecated':
      return 'Cap end-of-life — confirmar removal scheduled';
    case 'sunset':
      return 'Cap end-of-life — confirmar removal scheduled';
    default:
      return 'Revisar estado manualmente';
  }
}

// ────────────────────────────────────────────────────────────────────────────
// Summary one-liner derivado de drift_reasons
// ────────────────────────────────────────────────────────────────────────────

export function buildSummary(cap: CapStatusComputed): string {
  if (cap.scenarios_total === 0) {
    return 'Sin scenarios declarados';
  }
  if (cap.drift_reasons.length > 0) {
    return cap.drift_reasons[0];
  }
  if (cap.computed_status === 'partial') {
    return `${cap.scenarios_verified}/${cap.scenarios_total} scenarios verificados`;
  }
  if (cap.computed_status === 'declared-live') {
    return `${cap.scenarios_total} scenarios sin verificación declarada`;
  }
  if (cap.computed_status === 'wip') {
    return `${cap.scenarios_total} scenarios sin verificación pasando`;
  }
  return `computed: ${cap.computed_status}`;
}

// ────────────────────────────────────────────────────────────────────────────
// Filter: excluir verified-live
// ────────────────────────────────────────────────────────────────────────────

export function filterDriftEntries(
  report: ComputedStatusReport
): DriftEntry[] {
  const entries: DriftEntry[] = [];

  for (const [slug, cap] of Object.entries(report.capabilities)) {
    if (cap.computed_status === 'verified-live') continue;

    entries.push({
      slug,
      functional_area: null, // enriquecido opcionalmente por el componente
      computed_status: cap.computed_status,
      scenarios_total: cap.scenarios_total,
      drift_reasons: cap.drift_reasons,
      summary: buildSummary(cap),
      next_action: getNextAction(cap.computed_status),
    });
  }

  return entries;
}

// ────────────────────────────────────────────────────────────────────────────
// Sort: drift first, luego por severidad
// ────────────────────────────────────────────────────────────────────────────

export function sortDriftEntries(entries: DriftEntry[]): DriftEntry[] {
  return [...entries].sort(
    (a, b) => getSeverityOrder(a.computed_status) - getSeverityOrder(b.computed_status)
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Agrupar por agent_owner (prefijo antes del punto en functional_area)
// ────────────────────────────────────────────────────────────────────────────

export function groupByAgent(
  entries: DriftEntry[]
): Map<string, DriftEntry[]> {
  const map = new Map<string, DriftEntry[]>();

  for (const entry of entries) {
    const agent = entry.functional_area?.split('.')[0] ?? 'sin-agente';
    if (!map.has(agent)) map.set(agent, []);
    map.get(agent)!.push(entry);
  }

  return map;
}
