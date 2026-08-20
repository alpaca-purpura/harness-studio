/**
 * POST /api/transition (paridad dev-mode — el Go sirve esto en el binario)
 *  body: { sistema: string, storyId: string, targetState: StoryState, reason?: string }
 *  → { ok: true, newState, verbo }
 *
 * F6/RN-51: la whitelist, los dueños y la razón mínima ya NO viven hardcodeados acá —
 * se derivan del MISMO espejo gated del descriptor que el binario embebe
 * (lib/proceso-server.ts). Transición con requiere_razon escribe `{target}_reason`
 * (RN-49, genérico); al salir de un estado de categoría `pausado` se limpia su reason.
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { readdir } from 'node:fs/promises';
import { z } from 'zod';
import { errorResponse, safeJson } from '../_lib/responses';
import { readMarkdownWithFrontmatter } from '@/lib/fs-reader';
import { writeMarkdownWithFrontmatter } from '@/lib/fs-writer';
import { storiesPath, archiveRootPath, getSistemas } from '@/lib/workspace';
import {
  categoriaDe,
  duenoRender,
  esEstado,
  leerEspejo,
  transicionOperador,
} from '@/lib/proceso-server';
import type { StoryState } from '@/lib/types';

const BodySchema = z.object({
  sistema: z.string(),
  storyId: z.string().min(1),
  targetState: z.string().min(1),
  reason: z.string().optional(),
});

async function findStoryDir(sistema: string, storyId: string): Promise<string | null> {
  const liveDir = path.join(storiesPath(sistema), storyId);
  try {
    await readMarkdownWithFrontmatter(path.join(liveDir, 'checkpoint.md'));
    return liveDir;
  } catch {
    // continuar
  }

  const archiveRoot = archiveRootPath(sistema); // = archive/ (NO path.dirname(archivePath) → archive/{year} bug)
  try {
    const years = await readdir(archiveRoot, { withFileTypes: true });
    for (const y of years) {
      if (!y.isDirectory()) continue;
      const candidate = path.join(archiveRoot, y.name, 'stories', storyId);
      try {
        await readMarkdownWithFrontmatter(path.join(candidate, 'checkpoint.md'));
        return candidate;
      } catch {
        // continuar
      }
    }
  } catch {
    // archive dir no existe
  }
  return null;
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = BodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }
  const { sistema, storyId, targetState, reason } = parsed.data;

  let proceso;
  try {
    proceso = await leerEspejo();
  } catch (e) {
    return errorResponse(
      `espejo del descriptor ilegible: ${e instanceof Error ? e.message : e}`,
      500
    );
  }

  if (!esEstado(proceso, targetState)) {
    return errorResponse('body inválido', 400, { targetState });
  }

  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const storyDir = await findStoryDir(sistema, storyId);
  if (!storyDir) {
    return errorResponse('story no encontrada', 404, { story_id: storyId, sistema });
  }

  const checkpointPath = path.join(storyDir, 'checkpoint.md');
  const current = await readMarkdownWithFrontmatter(checkpointPath);
  const fromState = ((current.frontmatter as { state?: string }).state ||
    'idea') as StoryState;

  // Validar whitelist — derivada del descriptor
  const transition = transicionOperador(proceso, fromState, targetState);
  if (!transition) {
    const owner = duenoRender(proceso, targetState);
    const ownerNote = owner
      ? `esta transition la ejecuta ${owner}. Invoca la skill desde Claude Code.`
      : 'transition no permitida desde el cockpit.';
    return errorResponse('transition_forbidden', 403, {
      from: fromState,
      to: targetState,
      reason: ownerNote,
    });
  }

  // Validar reason cuando se requiere (mínimo = parámetro del descriptor)
  const razonMinima = proceso.parametros.razon_minima;
  if (transition.requiere_razon) {
    if (!reason || reason.trim().length < razonMinima) {
      return errorResponse(`reason obligatorio · ≥${razonMinima} chars`, 400, {
        from: fromState,
        to: targetState,
        verb: transition.nombre ?? transition.verbo,
      });
    }
  }

  const updatedFm: Record<string, unknown> = {
    ...current.frontmatter,
    state: targetState,
    last_modified: new Date().toISOString(),
  };

  // Razón genérica (RN-49): requiere_razon → {target}_reason; salir de una pausa
  // limpia su reason (por CATEGORÍA, no por nombre).
  if (transition.requiere_razon) {
    updatedFm[`${targetState}_reason`] = reason;
  } else if (categoriaDe(proceso, fromState) === 'pausado') {
    updatedFm[`${fromState}_reason`] = null;
  }

  try {
    await writeMarkdownWithFrontmatter(checkpointPath, updatedFm, current.content);
    return NextResponse.json({
      ok: true,
      newState: targetState,
      // El evento nombrado (RN-47): el verbo viaja como DATO en la respuesta.
      verbo: transition.verbo ?? '',
    });
  } catch (err) {
    return errorResponse('error escribiendo checkpoint', 500, {
      detail: (err as Error).message,
    });
  }
}
