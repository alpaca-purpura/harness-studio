/**
 * Release resolver — read / list / recompute status de releases.
 *
 * Doc canónico: docs/process/release-protocol.md.
 *
 * Recompute logic (Sección 5):
 *   - all stories in {done, dropped}        → ready_to_merge
 *   - any in {refining..reviewing}          → in_progress
 *   - all in idea                            → planning
 *   - else                                   → release.status (no-op)
 */

import path from 'node:path';
import matter from 'gray-matter';
import { readFile } from 'node:fs/promises';
import type { Release, ReleaseStatus, Story, StoryState } from './types';
import { listFiles, readMarkdownWithFrontmatter } from './fs-reader';
import { releasesPath, storiesPath, archivePath } from './workspace';
import { writeFileAtomic } from './fs-writer';

// ────────────────────────────────────────────────────────────────────────────
// Read / list
// ────────────────────────────────────────────────────────────────────────────

/** Lee un release YAML por sistema + release_id */
export async function readRelease(sistema: string, releaseId: string): Promise<Release> {
  const absPath = path.join(releasesPath(sistema), `${releaseId}.yaml`);
  return readReleaseFromPath(absPath);
}

async function readReleaseFromPath(absPath: string): Promise<Release> {
  const raw = await readFile(absPath, 'utf-8');
  const parsed = matter(raw);
  const data = parsed.data as Partial<Release>;
  return {
    release_id: data.release_id ?? path.basename(absPath, '.yaml'),
    sistema: data.sistema ?? '',
    name: data.name ?? '',
    description: data.description ?? '',
    status: data.status ?? 'planning',
    target_date: data.target_date ?? null,
    shipped_date: data.shipped_date ?? null,
    order: data.order ?? 0,
    created_at: data.created_at ?? '',
    created_by: data.created_by ?? 'operador', // archivos viejos pueden traer 'chris' (legacy F-1)
    stories: data.stories ?? [],
    verified_by: data.verified_by ?? null,
    verified_at: data.verified_at ?? null,
    verification_note: data.verification_note ?? null,
    production_status: data.production_status ?? null,
    production_version: data.production_version ?? null,
    production_scheduled_at: data.production_scheduled_at ?? null,
    deployed_at: data.deployed_at ?? null,
    release_branch: data.release_branch ?? null,
    maps_legacy_outcome: data.maps_legacy_outcome ?? null,
    maps_legacy_phase: data.maps_legacy_phase ?? null,
    body: parsed.content,
    path: absPath,
  };
}

/** Lista todos los releases del sistema ordenados por `order` ASC */
export async function listReleases(sistema: string): Promise<Release[]> {
  let files: string[] = [];
  try {
    files = await listFiles(releasesPath(sistema));
  } catch {
    return [];
  }
  const yamlFiles = files.filter((f) => f.endsWith('.yaml'));
  const releases = await Promise.all(yamlFiles.map(readReleaseFromPath));
  return releases.sort((a, b) => a.order - b.order);
}

/** Escribe un release atomic */
export async function writeRelease(release: Release): Promise<void> {
  if (!release.path) {
    throw new Error('release.path requerido para writeRelease');
  }
  const { body, path: _, ...frontmatter } = release;
  const serialized = matter.stringify(body ?? '', frontmatter as Record<string, unknown>);
  await writeFileAtomic(release.path, serialized);
}

// ────────────────────────────────────────────────────────────────────────────
// Recompute logic
// ────────────────────────────────────────────────────────────────────────────

const IN_PROGRESS_STATES: ReadonlySet<StoryState> = new Set([
  'refining',
  'refined',
  'ready',
  'developing',
  'developed',
  'reviewing',
]);

const TERMINAL_STATES: ReadonlySet<StoryState> = new Set(['done', 'dropped']);

/**
 * Recomputa el status de un release a partir de los estados actuales de sus
 * stories. Asume que `release.stories[]` están provistas con su `state` actual
 * (cargado vía `loadStoryStatesForRelease`).
 */
export function recomputeReleaseStatus(
  release: Pick<Release, 'status'>,
  storyStates: StoryState[]
): ReleaseStatus {
  // shipped es TERMINAL en el eje de integración — nunca se demota (un release
  // entregado es inmutable). Ver release-protocol.md § 2 + cockpit-permissions.md.
  if (release.status === 'shipped') {
    return 'shipped';
  }
  if (storyStates.length === 0) {
    // Release sin stories asignadas
    return 'planning';
  }
  if (storyStates.every((s) => TERMINAL_STATES.has(s))) {
    return 'ready_to_merge';
  }
  if (storyStates.some((s) => IN_PROGRESS_STATES.has(s))) {
    return 'in_progress';
  }
  if (storyStates.every((s) => s === 'idea')) {
    return 'planning';
  }
  return release.status as ReleaseStatus;
}

/**
 * Resuelve el state de cada story listada en `release.stories[]` leyendo su
 * checkpoint.md de disco. Maneja stories live (`{sistema}/docs/product/stories/`)
 * y archived (`{sistema}/docs/archive/{year}/stories/`).
 */
export async function loadStoryStatesForRelease(release: Release): Promise<StoryState[]> {
  const states: StoryState[] = [];
  const liveBase = storiesPath(release.sistema);
  const currentYear = new Date().getFullYear();
  const archiveBase = archivePath(release.sistema, currentYear);

  for (const storyId of release.stories) {
    const liveCkpt = path.join(liveBase, storyId, 'checkpoint.md');
    const archivedCkpt = path.join(archiveBase, storyId, 'checkpoint.md');
    let state: StoryState | null = null;
    for (const candidate of [liveCkpt, archivedCkpt]) {
      try {
        const parsed = await readMarkdownWithFrontmatter(candidate);
        const fm = parsed.frontmatter as { state?: StoryState };
        if (fm.state) {
          state = fm.state;
          break;
        }
      } catch {
        // try next
      }
    }
    if (state) states.push(state);
  }

  return states;
}

/**
 * Lee story.checkpoint → devuelve release_id si está declarado.
 * Útil cuando se quiere mapear story → release sin cargar todos los releases.
 */
export async function mapStoryToRelease(
  storyId: string,
  sistema: string
): Promise<string | null> {
  const liveBase = storiesPath(sistema);
  const currentYear = new Date().getFullYear();
  const archiveBase = archivePath(sistema, currentYear);

  for (const candidate of [
    path.join(liveBase, storyId, 'checkpoint.md'),
    path.join(archiveBase, storyId, 'checkpoint.md'),
  ]) {
    try {
      const parsed = await readMarkdownWithFrontmatter(candidate);
      const fm = parsed.frontmatter as Pick<Story, 'release'>;
      return fm.release ?? null;
    } catch {
      // try next
    }
  }
  return null;
}
