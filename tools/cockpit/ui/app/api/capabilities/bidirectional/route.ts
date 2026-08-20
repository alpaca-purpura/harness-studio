/**
 * GET /api/capabilities/bidirectional?sistema={sistema}
 * → { validation: BidirectionalValidationReport | null, path: string, sistema: string, hint?: string }
 *
 * Lee `{sistema}/docs/product/capabilities/_bidirectional-validation.json`.
 * Producido por `scripts/validate_code_cap_bidirectional.py` (cement 2026-05-28 · Fase B).
 * Si no existe → null + hint.
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { errorResponse } from '../../_lib/responses';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';
import type { BidirectionalValidationReport } from '@/lib/types';

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
    '_bidirectional-validation.json'
  );

  try {
    const raw = await readFile(jsonPath, 'utf-8');
    let data: BidirectionalValidationReport;
    try {
      data = JSON.parse(raw) as BidirectionalValidationReport;
    } catch (parseErr) {
      return errorResponse(
        `_bidirectional-validation.json mal formado: ${(parseErr as Error).message}`,
        500,
        { path: jsonPath }
      );
    }
    return NextResponse.json({ validation: data, path: jsonPath, sistema });
  } catch (err) {
    const nodeErr = err as NodeJS.ErrnoException;
    if (nodeErr.code === 'ENOENT') {
      return NextResponse.json({
        validation: null,
        path: jsonPath,
        sistema,
        hint: `Ejecuta: python3 scripts/validate_code_cap_bidirectional.py --sistema ${sistema}`,
      });
    }
    return errorResponse(
      `error leyendo _bidirectional-validation.json: ${(err as Error).message}`,
      500,
      { path: jsonPath }
    );
  }
}
