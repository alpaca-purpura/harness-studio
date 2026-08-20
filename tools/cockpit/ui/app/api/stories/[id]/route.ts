/**
 * GET   /api/stories/{id}?sistema=...  → { story }
 * PATCH /api/stories/{id}?sistema=...  body: { fields?: {...} | field, value } → { story }
 *
 * PATCH solo permite editar fields editables por el operador (cockpit-permissions.md):
 *   release, priority, goal, anti, reuse
 * Resto → 403.
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { z } from 'zod';
import { errorResponse, safeJson } from '../../_lib/responses';
import { readMarkdownWithFrontmatter } from '@/lib/fs-reader';
import { writeMarkdownWithFrontmatter } from '@/lib/fs-writer';
import { storiesPath, archiveRootPath, getSistemas, getSelectableSistemas } from '@/lib/workspace';
import { readdir } from 'node:fs/promises';
import type { Story } from '@/lib/types';

const EDITABLE_FIELDS = ['release', 'priority', 'goal', 'anti', 'reuse'] as const;
type EditableField = (typeof EDITABLE_FIELDS)[number];

function isEditable(field: string): field is EditableField {
  return (EDITABLE_FIELDS as readonly string[]).includes(field);
}

const PatchBodySchema = z.union([
  z.object({
    field: z.enum(EDITABLE_FIELDS),
    value: z.unknown(),
  }),
  z.object({
    fields: z.record(z.string(), z.unknown()),
  }),
]);

async function findStoryPath(sistema: string, storyId: string): Promise<string | null> {
  // 1. Buscar en stories live
  const liveDir = path.join(storiesPath(sistema), storyId);
  try {
    const parsed = await readMarkdownWithFrontmatter(path.join(liveDir, 'checkpoint.md'));
    if (parsed) return liveDir;
  } catch {
    // continuar
  }

  // 2. Buscar en archive (escanea años · {sistema}/docs/archive/{year}/stories/{id}).
  // Usar archiveRootPath (= archive/), NO path.dirname(archivePath(...)) que deja
  // archive/{year} — bug que hacía 404 toda story done/archivada.
  const archiveRoot = archiveRootPath(sistema);
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

async function loadStory(storyDir: string, sistema: string): Promise<Story> {
  const parsed = await readMarkdownWithFrontmatter(path.join(storyDir, 'checkpoint.md'));
  const fm = parsed.frontmatter as Partial<Story>;
  return {
    ...(fm as Story),
    story_id: fm.story_id || path.basename(storyDir),
    path: storyDir,
    sistema,
    release: fm.release ?? null,
    cap_target: fm.cap_target ?? null,
    cap_change_type: fm.cap_change_type ?? null,
    parent_story: fm.parent_story ?? null,
    state: fm.state ?? 'idea',
    body: parsed.content,
  };
}

export async function GET(
  req: NextRequest,
  context: { params: Promise<{ id: string }> }
): Promise<NextResponse> {
  const { id } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSelectableSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const storyDir = await findStoryPath(sistema, id);
  if (!storyDir) {
    return errorResponse('story no encontrada', 404, { story_id: id, sistema });
  }

  try {
    const story = await loadStory(storyDir, sistema);
    return NextResponse.json({ story });
  } catch (err) {
    return errorResponse('error leyendo story', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<{ id: string }> }
): Promise<NextResponse> {
  const { id } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = PatchBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido · esperaba {field, value} o {fields}', 400, {
      issues: parsed.error.issues,
    });
  }

  // Normalizar a record fields
  const patches: Record<string, unknown> =
    'field' in parsed.data
      ? { [parsed.data.field]: parsed.data.value }
      : parsed.data.fields;

  // Validar que todos los fields son editables
  const forbidden = Object.keys(patches).filter((f) => !isEditable(f));
  if (forbidden.length > 0) {
    return errorResponse(
      'algunos campos no son editables desde el cockpit',
      403,
      {
        forbidden_fields: forbidden,
        editable_fields: [...EDITABLE_FIELDS],
        reason:
          'estos fields los gestionan las skills (/po-ux, /architect, /dev-team, /pm-{sistema}). Modifícalos invocando la skill correspondiente.',
      }
    );
  }

  const storyDir = await findStoryPath(sistema, id);
  if (!storyDir) {
    return errorResponse('story no encontrada', 404, { story_id: id, sistema });
  }

  try {
    const ckptPath = path.join(storyDir, 'checkpoint.md');
    const current = await readMarkdownWithFrontmatter(ckptPath);
    const newFrontmatter = {
      ...current.frontmatter,
      ...patches,
      last_modified: new Date().toISOString(),
    };
    await writeMarkdownWithFrontmatter(ckptPath, newFrontmatter, current.content);
    const story = await loadStory(storyDir, sistema);
    return NextResponse.json({ story });
  } catch (err) {
    return errorResponse('error escribiendo checkpoint', 500, {
      detail: (err as Error).message,
    });
  }
}
