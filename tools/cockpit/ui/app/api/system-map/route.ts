/**
 * GET /api/system-map?sistema={sistema}  → { system_map: SystemMap, path: string }
 *
 * Lee {sistema}/docs/architecture/SYSTEM-MAP.yaml (schema v1.0).
 * SSoT estructural del producto.
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import YAML from 'yaml';
import { errorResponse } from '../_lib/responses';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';
import type { SystemMap } from '@/lib/types';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const wsRoot = getWorkspaceRoot();
  const yamlPath = path.join(wsRoot, sistema, 'docs', 'architecture', 'SYSTEM-MAP.yaml');

  try {
    const raw = await readFile(yamlPath, 'utf-8');
    const data = YAML.parse(raw) as SystemMap;
    return NextResponse.json({ system_map: data, path: yamlPath });
  } catch (err) {
    return errorResponse(`error leyendo SYSTEM-MAP.yaml: ${(err as Error).message}`, 500);
  }
}
