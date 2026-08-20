// voseo-allowed: archivo test contiene glosario voseo verbatim para validar detección en TOOLTIPS values
/**
 * tooltips.ts dictionary tests.
 *
 * Cement 2026-05-28 · refactor fuse /functionality into /map.
 * Estrategia: vitest puro (sin React Testing Library).
 *
 * Validamos:
 *   1. Diccionario tiene las keys esperadas para cada surface (CapDrawer, MapView, DriftView, BoardView, RoadmapView).
 *   2. Todos los values son strings non-empty.
 *   3. Spanish neutro: ninguno contiene voseo frecuente (vos/sos/tenés/podés/sabés/mirá/dejá/poné).
 *   4. Length cap razonable (≤ 320 chars) para que tooltips no se vuelvan ensayos.
 */

import { describe, it, expect } from 'vitest';
import { TOOLTIPS } from '../tooltips.js';

const REQUIRED_KEYS = [
  // CapDrawer schema cap
  'change_log',
  'functional_area',
  'agent_owner',
  'user_visible',
  'nature',
  'parent_cap',
  'derives_capabilities',
  'superseded_by',
  'architecture_pattern',
  'hipaa_lite_overlay',
  'dev_preview',
  'yaml_ledger',
  'created_in_story',
  // v3.2 cap blocks
  'scenarios',
  'edge_cases',
  'story_spec_ref',
  'e2e_test',
  'access_entry_points',
  'requires_role',
  'requires_clinic_scope',
  'forbidden_roles',
  'business_rules',
  'audit_trail',
  'enforcement',
  'related_capabilities',
  'severity',
  'code_files',
  'cap_header',
  // Verification (DriftView + MapView badges)
  'verified_live',
  'declared_live',
  'drift',
  'stub',
  'partial',
  'wip',
  'cross_check_3',
  'cross_check_4',
  'hard_fail',
  'soft_drift',
  'clean',
  'bidirectional_validation',
  // Estados del ciclo: MIGRADO (F6/RN-51) — los tooltips salen de estado.descripcion
  // del descriptor; las keys state_* murieron del diccionario.
  'wip_cap',
  // Harness Backlog estados (BoardView core/transversal)
  'harness_reported',
  'harness_triaged',
  'harness_ratified',
  'harness_applied',
  'harness_verified',
  'harness_deferred',
  'harness_otro',
  // Cockpit-specific
  'system_map',
  'product_health',
  'mapa_implementado',
  'roadmap',
  'drift_tab',
  'infra_role',
  'planned_status',
  'v3_2_badge',
  'release_concept',
] as const;

const VOSEO_PATTERNS = [
  /\bvos\b/i,
  /\bsos\b/i,
  /\btenés\b/i,
  /\bpodés\b/i,
  /\bsabés\b/i,
  /\bhacés\b/i,
  /\bvenís\b/i,
  /\bdecís\b/i,
  /\bmirá\b/i,
  /\bdejá\b/i,
  /\bponé\b/i,
  /\busá\b/i,
  /\belegí\b/i,
  /\bagregá\b/i,
  /\bguardá\b/i,
];

describe('TOOLTIPS dictionary', () => {
  it('contains all required keys', () => {
    for (const key of REQUIRED_KEYS) {
      expect(TOOLTIPS, `falta key: ${key}`).toHaveProperty(key);
    }
  });

  it('does not contain dead atomics keys (Fase 1b · unit is scenario now)', () => {
    // lifecycle.md § 2: atomic + atomic_ref MUERTOS — la unidad atómica es scenario.
    expect(TOOLTIPS).not.toHaveProperty('atomics');
    expect(TOOLTIPS).not.toHaveProperty('atomic_ref');
  });

  it('all values are non-empty strings', () => {
    for (const [key, value] of Object.entries(TOOLTIPS)) {
      expect(typeof value, `${key} debe ser string`).toBe('string');
      expect(value.length, `${key} no puede ser empty`).toBeGreaterThan(0);
    }
  });

  it('all values ≤ 320 chars (cap razonable)', () => {
    for (const [key, value] of Object.entries(TOOLTIPS)) {
      expect(value.length, `${key} excede 320 chars (${value.length})`).toBeLessThanOrEqual(320);
    }
  });

  it('all values are Spanish neutro (no voseo)', () => {
    for (const [key, value] of Object.entries(TOOLTIPS)) {
      for (const pattern of VOSEO_PATTERNS) {
        expect(
          pattern.test(value),
          `${key} contiene voseo: "${pattern.source}" en "${value}"`
        ).toBe(false);
      }
    }
  });
});
