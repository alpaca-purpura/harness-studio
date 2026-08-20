/**
 * GET /api/value-stream?sistema={sistema}  → { stages: ValueStreamStage[] | null }
 *
 * Expone el slot `value_stream` del seam (project.config.yaml) normalizado para
 * la lente "proceso" del Mapa (F-4: las etapas NO se hardcodean en el cockpit).
 * `stages: null` → slot vacío/__FILL_ME__ · el cliente usa el fallback genérico
 * (DEFAULT_VALUE_STREAM_STAGES de lib/map-zones.ts).
 */

import { NextRequest, NextResponse } from 'next/server';
import { errorResponse } from '../_lib/responses';
import { getSistemas } from '@/lib/workspace';
import { getValueStreamStages } from '@/lib/project-config';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  try {
    const stages = getValueStreamStages(sistema);
    return NextResponse.json({ stages });
  } catch {
    // project.config.yaml ausente/inparseable → el cliente cae al fallback genérico
    return NextResponse.json({ stages: null });
  }
}
