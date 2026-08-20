/**
 * GET   /api/capabilities/{module}/{cap}?sistema={sistema}   → { capability }
 * PATCH /api/capabilities/{module}/{cap}?sistema={sistema}
 *        body: { status: 'live'|'beta'|'deprecated'|'sunset', reason: string }
 *                                                        → { capability }
 *
 * PATCH solo permite editar `status` con razón (cementada en change_log
 * como entry type='fix'). NO permite editar scenarios ni change_log directos
 * — eso vive en `/api/extend-cap` via story creation.
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { z } from 'zod';
import { errorResponse, safeJson } from '../../../_lib/responses';
import { readCapability, writeCapability } from '@/lib/cap-ledger';
import { capabilitiesPath, getSistemas } from '@/lib/workspace';

const PatchBodySchema = z.object({
  status: z.enum(['live', 'beta', 'deprecated', 'sunset']),
  reason: z.string().min(10, 'reason debe tener ≥10 chars'),
});

function todayIso(): string {
  return new Date().toISOString().substring(0, 10);
}

function capYamlPath(sistema: string, module: string, slug: string): string {
  return path.join(capabilitiesPath(sistema), module, `${slug}.yaml`);
}

export async function GET(
  req: NextRequest,
  context: { params: Promise<{ module: string; cap: string }> }
): Promise<NextResponse> {
  const { module: moduleName, cap: capSlug } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  try {
    const absPath = capYamlPath(sistema, moduleName, capSlug);
    const capability = await readCapability(absPath);
    return NextResponse.json({ capability });
  } catch (err) {
    const code = (err as NodeJS.ErrnoException).code;
    if (code === 'ENOENT') {
      return errorResponse('capability no encontrada', 404, {
        sistema,
        module: moduleName,
        cap: capSlug,
      });
    }
    return errorResponse('error leyendo capability', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<{ module: string; cap: string }> }
): Promise<NextResponse> {
  const { module: moduleName, cap: capSlug } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  // Reject explícito si trae fields prohibidos
  if (typeof body === 'object' && body !== null) {
    const sent = Object.keys(body as Record<string, unknown>);
    const forbidden = sent.filter(
      (k) => k === 'scenarios' || k === 'change_log' || k === 'capability_id' || k === 'slug' || k === 'module' || k === 'license' || k === 'parent_cap' || k === 'derives_capabilities'
    );
    if (forbidden.length > 0) {
      return errorResponse(
        'edición directa de scenarios, change_log y campos estructurales prohibida',
        403,
        {
          forbidden_fields: forbidden,
          reason:
            'modificar scenarios o change_log requiere crear una story con cap_change_type. Usa /api/extend-cap o el botón "Extender" del cockpit.',
        }
      );
    }
  }

  const parsed = PatchBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }

  const absPath = capYamlPath(sistema, moduleName, capSlug);
  let cap;
  try {
    cap = await readCapability(absPath);
  } catch (err) {
    const code = (err as NodeJS.ErrnoException).code;
    if (code === 'ENOENT') {
      return errorResponse('capability no encontrada', 404);
    }
    throw err;
  }

  // Append change_log entry tipo 'fix' con razón
  const date = todayIso();
  cap.status = parsed.data.status;
  cap.last_modified = date;
  cap.change_log.push({
    story_id: 'cockpit-status-change',
    date,
    type: 'fix',
    summary: `Status cambiado a ${parsed.data.status} · razón: ${parsed.data.reason}`,
    scenarios_added: [],
    merge_sha: null,
    status: 'done',
  });

  try {
    await writeCapability(cap);
    return NextResponse.json({ capability: cap });
  } catch (err) {
    return errorResponse('error escribiendo capability', 500, {
      detail: (err as Error).message,
    });
  }
}
