/**
 * GET /api/harness  → { items: HarnessItem[], counts: Record<estado, number>, source: string }
 *
 * Lee docs/process/harness-backlog.md (SSoT del issue-tracker del harness) y lo
 * parsea a items estructurados para el board read-only del cockpit.
 *
 * NO es sistema-scoped: el harness-backlog es transversal (un solo archivo en la
 * raíz del workspace, vive en cualquier worktree). El cockpit solo LEE — la
 * captura/transición de estado sigue siendo del HLP (/harness-issue + lotes).
 */

import { NextResponse } from 'next/server';
import path from 'node:path';
import { readFile } from 'node:fs/promises';
import { errorResponse } from '../_lib/responses';
import { getWorkspaceRoot } from '@/lib/workspace';
import { parseHarnessBacklog, countByEstado } from '@/lib/harness-backlog';

const REL_SOURCE = 'docs/process/harness-backlog.md';

export async function GET(): Promise<NextResponse> {
  try {
    const abs = path.join(getWorkspaceRoot(), REL_SOURCE);
    let md: string;
    try {
      md = await readFile(abs, 'utf-8');
    } catch {
      // El archivo no existe en este worktree → board vacío, no es error fatal.
      return NextResponse.json({ items: [], counts: {}, source: REL_SOURCE });
    }
    const items = parseHarnessBacklog(md);
    return NextResponse.json({
      items,
      counts: countByEstado(items),
      source: REL_SOURCE,
    });
  } catch (err) {
    return errorResponse('error leyendo harness-backlog', 500, {
      detail: (err as Error).message,
    });
  }
}
