/**
 * GET /api/stories?sistema={sistema}  → { stories: Story[] }
 *
 * Agrega todos los checkpoints active + archived del sistema, parseando
 * frontmatter v2 via lib/fs-reader. Marca `is_archived` para cada story.
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { readdir, readFile } from 'node:fs/promises';
import { errorResponse } from '../_lib/responses';
import { readMarkdownWithFrontmatter } from '@/lib/fs-reader';
import { storiesPath, archiveRootPath, getSelectableSistemas } from '@/lib/workspace';
import type { Story } from '@/lib/types';

interface StoryWithArchive extends Story {
  is_archived: boolean;
}

async function safeListDirs(dir: string): Promise<string[]> {
  try {
    const entries = await readdir(dir, { withFileTypes: true });
    return entries.filter((e) => e.isDirectory()).map((e) => path.join(dir, e.name));
  } catch {
    return [];
  }
}

async function readCheckpoint(
  storyDir: string,
  sistema: string,
  isArchived: boolean
): Promise<StoryWithArchive | null> {
  const ckptPath = path.join(storyDir, 'checkpoint.md');
  try {
    const parsed = await readMarkdownWithFrontmatter(ckptPath);
    const fm = parsed.frontmatter as Partial<Story>;
    const storyId = fm.story_id || path.basename(storyDir);
    return {
      ...(fm as Story),
      story_id: storyId,
      path: storyDir,
      sistema,
      release: fm.release ?? null,
      cap_target: fm.cap_target ?? null,
      cap_change_type: fm.cap_change_type ?? null,
      parent_story: fm.parent_story ?? null,
      state: fm.state ?? 'idea',
      body: parsed.content,
      is_archived: isArchived,
    };
  } catch (err) {
    // Frontmatter malformado (ej. key duplicada → YAML inválido). NO silenciar:
    // rescatamos `state` via regex para ubicar la card en su columna real y
    // marcamos parse_error → el board muestra un badge rojo "⚠ checkpoint inválido".
    // Antes esto retornaba null (la story desaparecía) o caía a state=idea
    // (la story aparecía en la columna equivocada sin avisar) — confusión silenciosa.
    let raw = '';
    try {
      raw = await readFile(ckptPath, 'utf-8');
    } catch {
      return null; // checkpoint.md ni siquiera existe/legible → no es una story
    }
    const stateMatch = raw.match(/^state:[ \t]*([a-z]+)/m);
    return {
      story_id: path.basename(storyDir),
      path: storyDir,
      sistema,
      release: null,
      cap_target: null,
      cap_change_type: null,
      parent_story: null,
      state: (stateMatch?.[1] as Story['state']) ?? 'idea',
      body: raw,
      is_archived: isArchived,
      parse_error: (err as Error).message.split('\n')[0],
    };
  }
}

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) {
    return errorResponse('query param "sistema" requerido', 400);
  }

  const validSistemas = getSelectableSistemas();
  if (!validSistemas.includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400, {
      valid_sistemas: validSistemas,
    });
  }

  try {
    // Stories activas
    const liveDirs = await safeListDirs(storiesPath(sistema));
    const liveStories = await Promise.all(
      liveDirs.map((d) => readCheckpoint(d, sistema, false))
    );

    // Stories archivadas (escanea todos los años presentes bajo {sistema}/docs/archive/)
    const archiveRoot = archiveRootPath(sistema);
    let archivedDirs: string[] = [];
    try {
      const years = await readdir(archiveRoot, { withFileTypes: true });
      for (const y of years) {
        if (y.isDirectory()) {
          const storiesYearDir = path.join(archiveRoot, y.name, 'stories');
          const dirs = await safeListDirs(storiesYearDir);
          archivedDirs = archivedDirs.concat(dirs);
        }
      }
    } catch {
      // archive dir no existe · OK
    }

    const archivedStories = await Promise.all(
      archivedDirs.map((d) => readCheckpoint(d, sistema, true))
    );

    const all = [...liveStories, ...archivedStories].filter(
      (s): s is StoryWithArchive => s !== null
    );

    // Dedup por story_id (un id puede existir live + archivado por colisión de data,
    // ej. un stub re-creado con el id de una story ya done). React/dnd exigen ids
    // únicos → sin dedup la UI crashea. Preferimos la copia archivada (terminal/
    // canónica) y marcamos `dup_collision` para que la UI lo haga visible.
    const byId = new Map<string, StoryWithArchive>();
    const collisions = new Set<string>();
    for (const s of all) {
      const prev = byId.get(s.story_id);
      if (!prev) {
        byId.set(s.story_id, s);
        continue;
      }
      collisions.add(s.story_id);
      // archivada gana sobre live; si ya teníamos archivada, la mantenemos.
      if (s.is_archived && !prev.is_archived) byId.set(s.story_id, s);
    }
    for (const id of collisions) {
      const s = byId.get(id);
      if (s) s.dup_collision = true;
    }
    const stories = [...byId.values()];

    return NextResponse.json({ stories });
  } catch (err) {
    return errorResponse('error agregando stories', 500, {
      detail: (err as Error).message,
    });
  }
}
