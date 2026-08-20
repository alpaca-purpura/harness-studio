/**
 * GET /api/capabilities/status?sistema={sistema}
 * → { status: ComputedStatusReport | null, path: string, sistema: string, hint?: string }
 *
 * Lee `{sistema}/docs/product/capabilities/_status-computed.json`.
 * Ese archivo lo produce `scripts/compute_capability_status.py`.
 * Si el archivo no existe (script no corrido aún) → devuelve status: null + hint, no error.
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { errorResponse } from '../../_lib/responses';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';
import type { ComputedStatusReport } from '@/lib/types';

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
    '_status-computed.json'
  );

  try {
    const raw = await readFile(jsonPath, 'utf-8');
    let data: ComputedStatusReport;
    try {
      data = JSON.parse(raw) as ComputedStatusReport;
    } catch (parseErr) {
      return errorResponse(
        `_status-computed.json mal formado: ${(parseErr as Error).message}`,
        500,
        { path: jsonPath }
      );
    }
    return NextResponse.json({ status: data, path: jsonPath, sistema });
  } catch (err) {
    // ENOENT: archivo no existe (script aún no corrido) → respuesta 200 con null
    const nodeErr = err as NodeJS.ErrnoException;
    if (nodeErr.code === 'ENOENT') {
      return NextResponse.json({
        status: null,
        path: jsonPath,
        sistema,
        hint: `Ejecuta: python3 scripts/compute_capability_status.py --sistema ${sistema}`,
      });
    }
    // Otro error inesperado
    return errorResponse(
      `error leyendo _status-computed.json: ${(err as Error).message}`,
      500,
      { path: jsonPath }
    );
  }
}
