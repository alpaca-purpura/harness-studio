import { describe, expect, it } from 'vitest';

import { scenarioTruth } from '../cap-badges';

describe('scenarioTruth · badge de verdad del cap-drawer (N1)', () => {
  it('✅ verificado live cuando verified_real es objeto {at,how} (HB-58 cableado)', () => {
    const t = scenarioTruth({
      verified_real: { at: '2026-06-03T18:35:00Z', how: 'PATCH 200 + row en DB' },
      e2e_test: 'e2e/x.spec.ts',
    });
    expect(t.icon).toBe('✅');
    expect(t.tip).toContain('2026-06-03');
    expect(t.tip).toContain('PATCH 200');
  });

  it('✅ también con verified_real booleano (compat legacy)', () => {
    const t = scenarioTruth({ verified_real: true, e2e_test: null });
    expect(t.icon).toBe('✅');
  });

  it('🟠 con test cuando hay e2e_test pero NO verified_real', () => {
    const t = scenarioTruth({ verified_real: null, e2e_test: 'e2e/x.spec.ts' });
    expect(t.icon).toBe('🟠');
    expect(t.tip).toContain('sin evidencia de live-verify');
  });

  it('⚪ sin test cuando no hay ni verified_real ni e2e_test (deuda L4)', () => {
    const t = scenarioTruth({ verified_real: null, e2e_test: null });
    expect(t.icon).toBe('⚪');
    expect(t.tip).toContain('L4');
  });
});
