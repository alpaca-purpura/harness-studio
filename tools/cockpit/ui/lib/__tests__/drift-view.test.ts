/**
 * Tests para drift-helpers.ts
 *
 * Estrategia: tests unitarios puros de las funciones de filtrado/ordenamiento
 * sin necesidad de DOM (React Testing Library no instalado — vitest puro).
 */

import { describe, it, expect } from 'vitest';
import type { ComputedStatusReport } from '../types.js';
import {
  filterDriftEntries,
  sortDriftEntries,
  getNextAction,
  getSeverityOrder,
  buildSummary,
  groupByAgent,
  type DriftEntry,
} from '../drift-helpers.js';

// ────────────────────────────────────────────────────────────────────────────
// Fixtures
// ────────────────────────────────────────────────────────────────────────────

function makeReport(overrides?: Partial<ComputedStatusReport>): ComputedStatusReport {
  return {
    computed_at: '2026-05-28T11:35:00-05:00',
    sistema: 'main',
    capabilities: {
      'clinics-crud': {
        declared_status: 'live',
        computed_status: 'stub',
        scenarios_total: 0,
        scenarios_verified: 0,
        verification_total: 0,
        verification_pass: 0,
        drift_reasons: ['no scenarios declared'],
      },
      'shell-main': {
        declared_status: 'live',
        computed_status: 'drift',
        scenarios_total: 3,
        scenarios_verified: 0,
        verification_total: 3,
        verification_pass: 0,
        drift_reasons: ['verification path does not exist: frontend/src/features/shell/Shell.tsx'],
      },
      'auth-login': {
        declared_status: 'live',
        computed_status: 'verified-live',
        scenarios_total: 5,
        scenarios_verified: 5,
        verification_total: 5,
        verification_pass: 5,
        drift_reasons: [],
      },
    },
    summary: {
      total_caps: 3,
      verified_live: 1,
      declared_live: 0,
      partial: 0,
      wip: 0,
      stub: 1,
      drift: 1,
      deprecated: 0,
      sunset: 0,
    },
    ...overrides,
  };
}

// ────────────────────────────────────────────────────────────────────────────
// filterDriftEntries
// ────────────────────────────────────────────────────────────────────────────

describe('filterDriftEntries', () => {
  it('test_excludes_verified_live', () => {
    const report = makeReport();
    const entries = filterDriftEntries(report);

    const slugs = entries.map((e) => e.slug);
    expect(slugs).not.toContain('auth-login');
  });

  it('test_includes_stub_and_drift', () => {
    const report = makeReport();
    const entries = filterDriftEntries(report);

    const slugs = entries.map((e) => e.slug);
    expect(slugs).toContain('clinics-crud');
    expect(slugs).toContain('shell-main');
  });

  it('test_count_matches_non_verified', () => {
    const report = makeReport();
    const entries = filterDriftEntries(report);

    // 3 caps total - 1 verified-live = 2 drift entries
    expect(entries).toHaveLength(2);
  });

  it('test_empty_when_all_verified_live', () => {
    const allVerified: ComputedStatusReport = {
      computed_at: '2026-05-28T11:35:00-05:00',
      sistema: 'main',
      capabilities: {
        'cap-a': {
          declared_status: 'live',
          computed_status: 'verified-live',
          scenarios_total: 2,
          scenarios_verified: 2,
          verification_total: 2,
          verification_pass: 2,
          drift_reasons: [],
        },
      },
      summary: {
        total_caps: 1,
        verified_live: 1,
        declared_live: 0,
        partial: 0,
        wip: 0,
        stub: 0,
        drift: 0,
        deprecated: 0,
        sunset: 0,
      },
    };

    const entries = filterDriftEntries(allVerified);
    expect(entries).toHaveLength(0);
  });

  it('test_entry_has_correct_fields', () => {
    const report = makeReport();
    const entries = filterDriftEntries(report);
    const driftEntry = entries.find((e) => e.slug === 'shell-main');

    expect(driftEntry).toBeDefined();
    expect(driftEntry!.computed_status).toBe('drift');
    expect(driftEntry!.scenarios_total).toBe(3);
    expect(driftEntry!.drift_reasons).toHaveLength(1);
    expect(driftEntry!.next_action).toContain('URGENTE');
  });
});

// ────────────────────────────────────────────────────────────────────────────
// sortDriftEntries
// ────────────────────────────────────────────────────────────────────────────

describe('sortDriftEntries', () => {
  it('test_drift_appears_before_stub', () => {
    const entries: DriftEntry[] = [
      {
        slug: 'stub-cap',
        functional_area: null,
        computed_status: 'stub',
        scenarios_total: 0,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
      {
        slug: 'drift-cap',
        functional_area: null,
        computed_status: 'drift',
        scenarios_total: 3,
        drift_reasons: ['path missing'],
        summary: '',
        next_action: '',
      },
    ];

    const sorted = sortDriftEntries(entries);
    expect(sorted[0].slug).toBe('drift-cap');
    expect(sorted[1].slug).toBe('stub-cap');
  });

  it('test_severity_order_full', () => {
    const statuses = ['stub', 'declared-live', 'drift'] as const;
    const entries: DriftEntry[] = statuses.map((s) => ({
      slug: `cap-${s}`,
      functional_area: null,
      computed_status: s,
      scenarios_total: 1,
      drift_reasons: [],
      summary: '',
      next_action: '',
    }));

    const sorted = sortDriftEntries(entries);
    expect(sorted[0].computed_status).toBe('drift');
    expect(sorted[1].computed_status).toBe('declared-live');
    expect(sorted[2].computed_status).toBe('stub');
  });

  it('test_does_not_mutate_original', () => {
    const original: DriftEntry[] = [
      {
        slug: 'b',
        functional_area: null,
        computed_status: 'stub',
        scenarios_total: 0,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
      {
        slug: 'a',
        functional_area: null,
        computed_status: 'drift',
        scenarios_total: 1,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
    ];

    sortDriftEntries(original);
    // El original no se mutó
    expect(original[0].slug).toBe('b');
    expect(original[1].slug).toBe('a');
  });
});

// ────────────────────────────────────────────────────────────────────────────
// getNextAction
// ────────────────────────────────────────────────────────────────────────────

describe('getNextAction', () => {
  it('test_stub_action', () => {
    const action = getNextAction('stub');
    expect(action).toContain('Fase F.3');
  });

  it('test_drift_action_is_urgent', () => {
    const action = getNextAction('drift');
    expect(action).toContain('URGENTE');
  });

  it('test_declared_live_action', () => {
    const action = getNextAction('declared-live');
    expect(action).toContain('e2e_test');
  });

  it('test_deprecated_action', () => {
    const action = getNextAction('deprecated');
    expect(action).toContain('end-of-life');
  });
});

// ────────────────────────────────────────────────────────────────────────────
// getSeverityOrder
// ────────────────────────────────────────────────────────────────────────────

describe('getSeverityOrder', () => {
  it('test_drift_has_lowest_number', () => {
    expect(getSeverityOrder('drift')).toBeLessThan(getSeverityOrder('stub'));
    expect(getSeverityOrder('drift')).toBeLessThan(getSeverityOrder('declared-live'));
  });

  it('test_verified_live_has_highest_number', () => {
    expect(getSeverityOrder('verified-live')).toBeGreaterThan(getSeverityOrder('stub'));
  });
});

// ────────────────────────────────────────────────────────────────────────────
// buildSummary
// ────────────────────────────────────────────────────────────────────────────

describe('buildSummary', () => {
  it('test_stub_with_no_scenarios', () => {
    const summary = buildSummary({
      declared_status: 'live',
      computed_status: 'stub',
      scenarios_total: 0,
      scenarios_verified: 0,
      verification_total: 0,
      verification_pass: 0,
      drift_reasons: [],
    });
    expect(summary).toContain('Sin scenarios');
  });

  it('test_drift_shows_first_reason', () => {
    const summary = buildSummary({
      declared_status: 'live',
      computed_status: 'drift',
      scenarios_total: 3,
      scenarios_verified: 0,
      verification_total: 3,
      verification_pass: 0,
      drift_reasons: ['path X does not exist'],
    });
    expect(summary).toContain('path X');
  });
});

// ────────────────────────────────────────────────────────────────────────────
// groupByAgent
// ────────────────────────────────────────────────────────────────────────────

describe('groupByAgent', () => {
  it('test_groups_by_prefix', () => {
    const entries: DriftEntry[] = [
      {
        slug: 'cap-a',
        functional_area: 'lisa.servicios',
        computed_status: 'stub',
        scenarios_total: 0,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
      {
        slug: 'cap-b',
        functional_area: 'valeria.agenda',
        computed_status: 'drift',
        scenarios_total: 1,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
      {
        slug: 'cap-c',
        functional_area: 'lisa.otro',
        computed_status: 'declared-live',
        scenarios_total: 2,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
    ];

    const grouped = groupByAgent(entries);
    expect(grouped.get('lisa')).toHaveLength(2);
    expect(grouped.get('valeria')).toHaveLength(1);
  });

  it('test_null_functional_area_goes_to_sin_agente', () => {
    const entries: DriftEntry[] = [
      {
        slug: 'orphan',
        functional_area: null,
        computed_status: 'stub',
        scenarios_total: 0,
        drift_reasons: [],
        summary: '',
        next_action: '',
      },
    ];

    const grouped = groupByAgent(entries);
    expect(grouped.get('sin-agente')).toHaveLength(1);
  });
});
