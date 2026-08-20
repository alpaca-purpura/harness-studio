/**
 * GET /api/capabilities/code-index?sistema={sistema}
 * → { index: CodeIndexReport | null, path: string, sistema: string, hint?: string }
 *
 * Lee `{sistema}/docs/product/capabilities/_code-index.json`.
 * Producido por `scripts/generate_code_to_cap_index.py` (cement 2026-05-28 · Fase A).
 * Si no existe → null + hint.
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { errorResponse } from '../../_lib/responses';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';
import type { CodeIndexReport } from '@/lib/types';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const wsRoot = getWorkspaceRoot();
  const jsonPath = path.join(
    wsRoot,
    sistema,
    'docs',
    'product',
    'capabilities',
    '_code-index.json'
  );

  try {
    const raw = await readFile(jsonPath, 'utf-8');
    let data: CodeIndexReport;
    try {
      data = JSON.parse(raw) as CodeIndexReport;
    } catch (parseErr) {
      return errorResponse(
        `_code-index.json mal formado: ${(parseErr as Error).message}`,
        500,
        { path: jsonPath }
      );
    }
    return NextResponse.json({ index: data, path: jsonPath, sistema });
  } catch (err) {
    const nodeErr = err as NodeJS.ErrnoException;
    if (nodeErr.code === 'ENOENT') {
      return NextResponse.json({
        index: null,
        path: jsonPath,
        sistema,
        hint: `Ejecuta: python3 scripts/generate_code_to_cap_index.py --sistema ${sistema}`,
      });
    }
    return errorResponse(
      `error leyendo _code-index.json: ${(err as Error).message}`,
      500,
      { path: jsonPath }
    );
  }
}
