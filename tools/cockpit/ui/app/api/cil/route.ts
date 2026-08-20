/**
 * GET /api/cil  → board consolidado de los 4 carriles del CIL (read-only)
 * ────────────────────────────────────────────────────────────────────────────
 * El CIL (Continuous Improvement Loop · proceso v5 §5.7) tiene 4 carriles:
 *   L1 harness/proceso   → docs/process/harness-backlog.md      (items + estado)
 *   L2 producto/skills-arq → docs/learnings/**                  (count + link /learnings)
 *   L3 deuda técnica     → docs/process/tech-debt.md            (items + estado)
 *   L4 capability-desfasada → cap-doctor por sistema              (link /drift · per-sistema)
 *
 * Este endpoint hace el "monitor único": L1+L3 traen sus items (mismo lifecycle),
 * L2 trae un conteo del home transversal, L4 queda como link (es per-sistema y ya
 * tiene su vista `/drift`). Transversal — NO sistema-scoped. El cockpit solo LEE;
 * la captura/transición sigue siendo del HLP + dev-team/auditor.
 */

import { NextResponse } from 'next/server';
import path from 'node:path';
import { readFile, readdir } from 'node:fs/promises';
import { errorResponse } from '../_lib/responses';
import { getWorkspaceRoot } from '@/lib/workspace';
import { parseHarnessBacklog, countByEstado, type HarnessItem } from '@/lib/harness-backlog';
import { parseTechDebt, countTdByEstado, type TechDebtItem } from '@/lib/tech-debt';
import { countOpen } from '@/lib/md-lifecycle-table';

const L1_SOURCE = 'docs/process/harness-backlog.md';
const L3_SOURCE = 'docs/process/tech-debt.md';
const L2_SOURCE = 'docs/learnings';

async function readSafe(abs: string): Promise<string | null> {
  try {
    return await readFile(abs, 'utf-8');
  } catch {
    return null;
  }
}

/** Cuenta `.md` recursivamente bajo un dir (best-effort · ignora errores de fs). */
async function countMarkdown(dir: string): Promise<number> {
  let total = 0;
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return 0;
  }
  for (const e of entries) {
    const full = path.join(dir, e.name);
    if (e.isDirectory()) {
      total += await countMarkdown(full);
    } else if (e.isFile() && e.name.endsWith('.md')) {
      total += 1;
    }
  }
  return total;
}

export interface CilBoardResponse {
  l1: { items: HarnessItem[]; counts: Record<string, number>; open: number; source: string };
  l3: { items: TechDebtItem[]; counts: Record<string, number>; open: number; source: string };
  l2: { count: number; source: string; link: string };
  l4: { source: string; link: string; note: string };
}

export async function GET(): Promise<NextResponse> {
  try {
    const root = getWorkspaceRoot();

    const l1Md = await readSafe(path.join(root, L1_SOURCE));
    const l1Items = l1Md ? parseHarnessBacklog(l1Md) : [];

    const l3Md = await readSafe(path.join(root, L3_SOURCE));
    const l3Items = l3Md ? parseTechDebt(l3Md) : [];

    const l2Count = await countMarkdown(path.join(root, L2_SOURCE));

    const body: CilBoardResponse = {
      l1: {
        items: l1Items,
        counts: countByEstado(l1Items),
        open: countOpen(l1Items),
        source: L1_SOURCE,
      },
      l3: {
        items: l3Items,
        counts: countTdByEstado(l3Items),
        open: countOpen(l3Items),
        source: L3_SOURCE,
      },
      l2: { count: l2Count, source: L2_SOURCE, link: '/learnings' },
      l4: {
        source: 'cap-doctor (por sistema)',
        link: '/drift',
        note: 'Capabilities desfasadas se monitorean por sistema en la vista Drift.',
      },
    };

    return NextResponse.json(body);
  } catch (err) {
    return errorResponse('error agregando CIL', 500, { detail: (err as Error).message });
  }
}
