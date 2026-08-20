/**
 * Tests para GET /api/capabilities/status
 *
 * Estrategia: mock `node:fs/promises` + `@/lib/workspace` para
 * aislar la lógica de la route sin necesidad de disco real.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ComputedStatusReport } from '../types.js';

// ────────────────────────────────────────────────────────────────────────────
// Mocks
// ────────────────────────────────────────────────────────────────────────────

vi.mock('node:fs/promises', () => ({
  readFile: vi.fn(),
}));

vi.mock('@/lib/workspace', () => ({
  getWorkspaceRoot: vi.fn(() => '/fake/workspace'),
  getSistemas: vi.fn(() => ['main', 'acme']),
  _resetWorkspaceCache: vi.fn(),
}));

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

function makeRequest(sistema: string): Request {
  return new Request(
    `http://localhost:3000/api/capabilities/status?sistema=${sistema}`
  );
}

// Importar después de los mocks (lazy para asegurar hoisting)
async function getHandler() {
  const mod = await import(
    '../../app/api/capabilities/status/route.js'
  );
  return mod.GET;
}

// ────────────────────────────────────────────────────────────────────────────
// Tests
// ────────────────────────────────────────────────────────────────────────────

describe('GET /api/capabilities/status', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('test_returns_data_when_file_exists', async () => {
    const { readFile } = await import('node:fs/promises');
    const mockReport: ComputedStatusReport = {
      computed_at: '2026-05-28T11:35:00-05:00',
      sistema: 'main',
      capabilities: {
        'shell-main': {
          declared_status: 'live',
          computed_status: 'declared-live',
          scenarios_total: 7,
          scenarios_verified: 0,
          verification_total: 0,
          verification_pass: 0,
          drift_reasons: ['e2e_test missing in all 7 scenarios'],
        },
      },
      summary: {
        total_caps: 64,
        verified_live: 0,
        declared_live: 1,
        partial: 0,
        wip: 0,
        stub: 63,
        drift: 0,
        deprecated: 0,
        sunset: 0,
      },
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(readFile).mockResolvedValueOnce(JSON.stringify(mockReport) as any);

    const GET = await getHandler();
    const req = makeRequest('main');
    // Cast to NextRequest-like (has nextUrl.searchParams)
    const nextReq = Object.assign(req, {
      nextUrl: new URL(req.url),
    });

    const res = await GET(nextReq as Parameters<typeof GET>[0]);
    const body = await res.json() as { status: ComputedStatusReport; sistema: string; path: string };

    expect(res.status).toBe(200);
    expect(body.status).toBeDefined();
    expect(body.status.sistema).toBe('main');
    expect(body.status.summary.total_caps).toBe(64);
    expect(body.status.capabilities['shell-main'].computed_status).toBe('declared-live');
    expect(body.sistema).toBe('main');
    expect(body.path).toContain('_status-computed.json');
  });

  it('test_returns_null_when_file_missing', async () => {
    const { readFile } = await import('node:fs/promises');
    const enoentErr = Object.assign(new Error('ENOENT: no such file'), {
      code: 'ENOENT',
    });
    vi.mocked(readFile).mockRejectedValueOnce(enoentErr);

    const GET = await getHandler();
    const req = makeRequest('main');
    const nextReq = Object.assign(req, {
      nextUrl: new URL(req.url),
    });

    const res = await GET(nextReq as Parameters<typeof GET>[0]);
    const body = await res.json() as { status: null; hint: string; path: string; sistema: string };

    expect(res.status).toBe(200);
    expect(body.status).toBeNull();
    expect(body.hint).toContain('compute_capability_status.py');
    expect(body.hint).toContain('main');
    expect(body.path).toContain('_status-computed.json');
  });
});
