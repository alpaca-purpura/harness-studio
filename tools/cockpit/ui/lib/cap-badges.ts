/**
 * cap-badges · lógica PURA del "badge de verdad" del cap-drawer (N1 cap-levels-proposal §2).
 *
 * Sin React → testeable en aislamiento. Deriva el badge del DATO REAL del scenario,
 * NO de un campo decorativo:
 *   ✅ verificado live  → `verified_real` presente (evidencia de live-verify · HB-58 cableado)
 *   🟠 con test         → `e2e_test` declarado, sin verified_real
 *   ⚪ sin test         → ninguno → deuda de verificación (carril L4 del CIL)
 */

import type { CapScenario } from './types';

export interface ScenarioTruth {
  icon: string;
  label: string;
  /** clases Tailwind del Pill */
  cls: string;
  /** tooltip explicativo (evidencia o deuda) */
  tip: string;
}

export function scenarioTruth(s: Pick<CapScenario, 'verified_real' | 'e2e_test'>): ScenarioTruth {
  const vr = s.verified_real;
  if (vr) {
    const how =
      typeof vr === 'object' && vr !== null ? `${vr.at} — ${vr.how}` : 'verificado live';
    return { icon: '✅', label: 'verificado live', cls: 'bg-[#064e3b] text-[#6ee7b7]', tip: how };
  }
  if (s.e2e_test) {
    return {
      icon: '🟠',
      label: 'con test',
      cls: 'bg-[#713f12] text-[#fbbf24]',
      tip: 'e2e_test declarado, pero sin evidencia de live-verify (verified_real) registrada.',
    };
  }
  return {
    icon: '⚪',
    label: 'sin test',
    cls: 'bg-[#27272a] text-[#a1a1aa]',
    tip: 'Caso de uso declarado live SIN e2e_test → deuda de verificación (carril L4 del CIL).',
  };
}
