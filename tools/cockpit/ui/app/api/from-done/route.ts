/**
 * POST /api/from-done
 *  body: {
 *    sistema: string,
 *    parentStoryId: string,
 *    newStorySlug: string,
 *    goal: string,
 *    release: string
 *  }
 *  → { storyId, checkpointPath, operatorInputPath }
 *
 * Crea una story nueva basada en una done. Auto-fill parent_story + appendea
 * ref "story-ref" al operator-input de la nueva ("Spawned from parent: X done").
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { stat, readdir } from 'node:fs/promises';
import { z } from 'zod';
import { errorResponse, safeJson } from '../_lib/responses';
import { createNewStoryDocs, appendStoryRefToOperatorInput } from '../_lib/story-templates';
import { storiesPath, archiveRootPath, getSistemas } from '@/lib/workspace';
import { readMarkdownWithFrontmatter } from '@/lib/fs-reader';
import type { Story } from '@/lib/types';

const SLUG_REGEX = /^[a-z0-9][a-z0-9-]*$/;

const BodySchema = z.object({
  sistema: z.string(),
  parentStoryId: z.string().min(1),
  newStorySlug: z.string().regex(SLUG_REGEX, 'newStorySlug inválido (solo a-z 0-9 -)'),
  goal: z.string().min(10, 'goal debe tener ≥10 chars'),
  release: z.string().min(1),
});

async function findParentStory(
  sistema: string,
  parentId: string
): Promise<{ dir: string; story: Partial<Story> } | null> {
  const candidates: string[] = [path.join(storiesPath(sistema), parentId)];

  const archiveRoot = archiveRootPath(sistema); // = archive/ (NO path.dirname(archivePath) → archive/{year} bug)
  try {
    const years = await readdir(archiveRoot, { withFileTypes: true });
    for (const y of years) {
      if (y.isDirectory()) {
        candidates.push(path.join(archiveRoot, y.name, 'stories', parentId));
      }
    }
  } catch {
    // archive dir no existe
  }

  for (const dir of candidates) {
    try {
      const parsed = await readMarkdownWithFrontmatter(path.join(dir, 'checkpoint.md'));
      return { dir, story: parsed.frontmatter as Partial<Story> };
    } catch {
      // try next
    }
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
  const data = parsed.data;

  if (!getSistemas().includes(data.sistema)) {
    return errorResponse(`sistema desconocido: ${data.sistema}`, 400);
  }

  // Validar parent existe
  const parent = await findParentStory(data.sistema, data.parentStoryId);
  if (!parent) {
    return errorResponse('parent story no encontrada', 404, {
      parent_story_id: data.parentStoryId,
    });
  }

  if (parent.story.state !== 'done') {
    return errorResponse('parent story no está en state=done', 400, {
      parent_story_id: data.parentStoryId,
      parent_state: parent.story.state,
      reason: 'solo se permite spawnar desde stories ya done (cementadas).',
    });
  }

  // Verificar que la story nueva no existe
  const newDir = path.join(storiesPath(data.sistema), data.newStorySlug);
  try {
    await stat(newDir);
    return errorResponse('story ya existe', 409, {
      story_id: data.newStorySlug,
    });
  } catch {
    // expected · no existe
  }

  try {
    const result = await createNewStoryDocs({
      sistema: data.sistema,
      slug: data.newStorySlug,
      goal: data.goal,
      release: data.release,
      capTarget: parent.story.cap_target ?? null,
      capChangeType: parent.story.cap_target ? 'extend' : null,
      parentStory: data.parentStoryId,
      // Hereda agente + módulo de la story padre → board pinta agente/módulo.
      agentOwner: parent.story.agent_owner ?? null,
      module: parent.story.module ?? null,
      spawnedBy: 'cockpit-from-done',
    });

    // Appendea ref al operator-input
    await appendStoryRefToOperatorInput(
      result.operatorInputPath,
      data.parentStoryId,
      `Spawned from parent: ${data.parentStoryId} (done)`
    );

    return NextResponse.json({
      storyId: result.storyId,
      checkpointPath: result.checkpointPath,
      operatorInputPath: result.operatorInputPath,
    });
  } catch (err) {
    return errorResponse('error creando story', 500, {
      detail: (err as Error).message,
    });
  }
}
