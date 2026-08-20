/**
 * GET /api/capabilities/doctor?sistema={sistema}
 * → { doctor: CapDoctorReport | null, path: string, sistema: string, hint?: string }
 *
 * Health report code↔cap de un vistazo (HB-51 · Capa 8). Reshapea los gates G1-G6 de
 * `{sistema}/docs/product/capabilities/_bidirectional-validation.json` (producido por
 * `scripts/validate_code_cap_bidirectional.py`) en una vista "doctor": sano/deriva por
 * gate + detalles. Mismo SSoT que el pre-commit (5e) y el CLI `make cap-doctor`.
 * Si el JSON no existe → null + hint.
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { errorResponse } from '../../_lib/responses';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';

const GATE_LABEL: Record<string, string> = {
  G1: 'headers huérfanos (→ cap inexistente)',
  G2: 'cajas live vacías (área sin cap)',
  G3: 'caps sin hogar (fa ∉ SYSTEM-MAP)',
  G4: 'paths rotos (declarados, no existen)',
  G5: 'supersesiones rotas',
  G6: 'caps live invisibles en el mapa',
};

interface GateResult {
  total: number;
  pass: number;
  drift: number;
  details?: Array<Record<string, unknown>>;
}

interface BidirectionalJson {
  cap_gates?: Record<string, GateResult>;
  cap_gates_hard?: boolean;
  validated_at?: string;
  summary?: { cap_gate_drift?: number };
}

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
    let data: BidirectionalJson;
    try {
      data = JSON.parse(raw) as BidirectionalJson;
    } catch (parseErr) {
      return errorResponse(
        `_bidirectional-validation.json mal formado: ${(parseErr as Error).message}`,
        500,
        { path: jsonPath }
      );
    }

    const gates = data.cap_gates ?? {};
    const gateView = Object.entries(gates).map(([id, g]) => ({
      id,
      label: GATE_LABEL[id] ?? id,
      drift: g.drift ?? 0,
      total: g.total ?? 0,
      details: g.details ?? [],
    }));
    const totalDrift = gateView.reduce((n, g) => n + g.drift, 0);

    return NextResponse.json({
      doctor: {
        sistema,
        healthy: totalDrift === 0,
        total_drift: totalDrift,
        hard_enforced: Boolean(data.cap_gates_hard),
        validated_at: data.validated_at ?? null,
        gates: gateView,
      },
      path: jsonPath,
      sistema,
    });
  } catch (err) {
    const nodeErr = err as NodeJS.ErrnoException;
    if (nodeErr.code === 'ENOENT') {
      return NextResponse.json({
        doctor: null,
        path: jsonPath,
        sistema,
        hint: `Ejecuta: make cap-doctor SISTEMA=${sistema} (o python3 scripts/validate_code_cap_bidirectional.py --sistema ${sistema})`,
      });
    }
    return errorResponse(
      `error leyendo _bidirectional-validation.json: ${(err as Error).message}`,
      500,
      { path: jsonPath }
    );
  }
}
