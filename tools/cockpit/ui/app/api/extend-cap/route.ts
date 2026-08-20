/**
 * POST /api/extend-cap
 *  body: {
 *    sistema: string,
 *    parentCap: { module: string, slug: string },
 *    capChangeType: 'fix' | 'extend' | 'derive',
 *    newStorySlug: string,
 *    goal: string,
 *    release: string,
 *    derivedName?: string   // requerido si capChangeType=derive
 *  }
 *  → { storyId, slug, checkpointPath, operatorInputPath }
 *
 * Crea checkpoint.md + operator-input declarando parentCap + capChangeType.
 * Si capChangeType=derive, declara parent_story=null (es story propia) y
 * cap_target=derivedName (el cap hijo NO se crea ahora · se materializa al
 * merge en Fase F.3 via createDerivedCap).
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { z } from 'zod';
import { errorResponse, safeJson } from '../_lib/responses';
import { createNewStoryDocs } from '../_lib/story-templates';
import { readCapability } from '@/lib/cap-ledger';
import { capabilitiesPath, storiesPath, getSistemas } from '@/lib/workspace';
import { stat } from 'node:fs/promises';

// Story slugs nuevos siguen convención kebab-case (a-z 0-9 -).
const STORY_SLUG_REGEX = /^[a-z0-9][a-z0-9-]*$/;
// Referencias a módulos/caps EXISTENTES en filesystem pueden contener guion bajo
// (sistema_studio, sales_agent, offer_studio, public_landing, wizard_sistema_studio_slice_1).
// El regex de referencia DEBE aceptar `_` o rechaza caps válidos al extender.
const REF_SLUG_REGEX = /^[a-z0-9][a-z0-9_-]*$/;

const BodySchema = z.object({
  sistema: z.string(),
  parentCap: z.object({
    module: z.string().regex(REF_SLUG_REGEX, 'module inválido (a-z 0-9 _ -)'),
    slug: z.string().regex(REF_SLUG_REGEX, 'cap slug inválido (a-z 0-9 _ -)'),
  }),
  capChangeType: z.enum(['fix', 'extend', 'derive']),
  newStorySlug: z.string().regex(STORY_SLUG_REGEX, 'newStorySlug inválido (solo a-z 0-9 -)'),
  goal: z.string().min(10, 'goal debe tener ≥10 chars'),
  release: z.string().min(1),
  derivedName: z.string().regex(STORY_SLUG_REGEX, 'derivedName inválido (solo a-z 0-9 -)').optional(),
});

export async function POST(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = BodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }
  const data = parsed.data;

  if (!getSistemas().includes(data.sistema)) {
    return errorResponse(`sistema desconocido: ${data.sistema}`, 400);
  }

  // Validar parentCap existe
  const parentCapPath = path.join(
    capabilitiesPath(data.sistema),
    data.parentCap.module,
    `${data.parentCap.slug}.yaml`
  );
  let parentCapData;
  try {
    parentCapData = await readCapability(parentCapPath);
  } catch {
    return errorResponse('parentCap no encontrada', 404, {
      parent: `${data.parentCap.module}/${data.parentCap.slug}`,
    });
  }

  // Si type=derive, requiere derivedName
  if (data.capChangeType === 'derive' && !data.derivedName) {
    return errorResponse(
      'capChangeType=derive requiere "derivedName" (slug del cap hijo a crear al merge)',
      400
    );
  }

  // Verificar que la story no existe ya
  const storyDir = path.join(storiesPath(data.sistema), data.newStorySlug);
  try {
    await stat(storyDir);
    return errorResponse('story ya existe', 409, {
      story_id: data.newStorySlug,
    });
  } catch {
    // expected · no existe
  }

  // capTarget según tipo:
  //   - fix/extend → apunta al cap padre (mismo slug)
  //   - derive     → apunta al derivedName (cap hijo futuro)
  const capTarget =
    data.capChangeType === 'derive'
      ? data.derivedName!
      : data.parentCap.slug;

  try {
    const result = await createNewStoryDocs({
      sistema: data.sistema,
      slug: data.newStorySlug,
      goal: data.goal,
      release: data.release,
      capTarget,
      capChangeType: data.capChangeType,
      parentStory: null,
      // Hereda agente + módulo del cap padre → el board pinta agente/módulo
      // en vez de "—". (derive crea un cap hijo nuevo pero mantiene la raíz.)
      agentOwner: parentCapData.agent_owner ?? null,
      module: parentCapData.module || data.parentCap.module,
      spawnedBy: 'cockpit-extend-cap',
    });

    return NextResponse.json({
      storyId: result.storyId,
      slug: result.storyId,
      checkpointPath: result.checkpointPath,
      operatorInputPath: result.operatorInputPath,
    });
  } catch (err) {
    return errorResponse('error creando story', 500, {
      detail: (err as Error).message,
    });
  }
}
