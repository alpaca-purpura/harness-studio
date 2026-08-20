/**
 * Regresión GET /api/stories — las stories `done` viven en
 * `{sistema}/docs/archive/{year}/stories/` y DEBEN aparecer (is_archived: true).
 *
 * Bug fix 2026-05-29: el route derivaba el archive root con
 * `path.dirname(archivePath(sistema,'0000'))` → `…/archive/0000` (un nivel de más) →
 * el readdir fallaba silencioso → la columna `done` del board quedaba vacía aunque
 * hubiera 29 stories archivadas. Fix: `archiveRootPath(sistema)` → `…/archive`.
 *
 * Usa un workspace temporal real (sin mocks) para ejercitar la resolución de paths.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { _resetWorkspaceCache } from '../workspace';

let root: string;

function writeStory(rel: string, frontmatter: string): void {
  const dir = path.join(root, rel);
  mkdirSync(dir, { recursive: true });
  writeFileSync(path.join(dir, 'checkpoint.md'), `${frontmatter}\n# body\n`, 'utf-8');
}

async function getHandler() {
  const mod = await import('../../app/api/stories/route.js');
  return mod.GET;
}

function makeReq(sistema: string) {
  const req = new Request(`http://localhost:3000/api/stories?sistema=${sistema}`);
  return Object.assign(req, { nextUrl: new URL(req.url) });
}

beforeEach(() => {
  root = mkdtempSync(path.join(tmpdir(), 'cockpit-stories-'));
  process.env.WORKSPACE_ROOT = root;
  _resetWorkspaceCache();
  // Story activa
  writeStory(
    'main/docs/product/stories/main-live-example',
    '---\nstory_id: main-live-example\nstate: refining\n---'
  );
  // Story done archivada (el caso del bug)
  writeStory(
    'main/docs/archive/2026/stories/main-done-example',
    '---\nstory_id: main-done-example\nstate: done\nrelease: F1\n---'
  );
});

afterEach(() => {
  delete process.env.WORKSPACE_ROOT;
  _resetWorkspaceCache();
  rmSync(root, { recursive: true, force: true });
});

describe('GET /api/stories — archived done stories', () => {
  it('incluye stories done archivadas con is_archived: true', async () => {
    const GET = await getHandler();
    const res = await GET(makeReq('main') as Parameters<typeof GET>[0]);
    const body = (await res.json()) as {
      stories: Array<{ story_id: string; state: string; is_archived: boolean }>;
    };

    expect(res.status).toBe(200);
    const done = body.stories.find((s) => s.story_id === 'main-done-example');
    expect(done, 'la story done archivada debe aparecer (bug archive root)').toBeDefined();
    expect(done?.state).toBe('done');
    expect(done?.is_archived).toBe(true);
  });

  it('sigue incluyendo stories activas (is_archived: false)', async () => {
    const GET = await getHandler();
    const res = await GET(makeReq('main') as Parameters<typeof GET>[0]);
    const body = (await res.json()) as {
      stories: Array<{ story_id: string; is_archived: boolean }>;
    };
    const live = body.stories.find((s) => s.story_id === 'main-live-example');
    expect(live?.is_archived).toBe(false);
  });

  it('deduplica story_id que existe live + archivado (prefiere archivada + flag dup_collision)', async () => {
    // Mismo story_id en ambos lados → colisión (caso real: copia live + copia done del mismo id)
    writeStory(
      'main/docs/product/stories/main-colision',
      '---\nstory_id: main-colision\nstate: idea\nrelease: F0\n---'
    );
    writeStory(
      'main/docs/archive/2026/stories/main-colision',
      '---\nstory_id: main-colision\nstate: done\nrelease: F0\n---'
    );
    const GET = await getHandler();
    const res = await GET(makeReq('main') as Parameters<typeof GET>[0]);
    const body = (await res.json()) as {
      stories: Array<{
        story_id: string;
        state: string;
        is_archived: boolean;
        dup_collision?: boolean;
      }>;
    };
    const hits = body.stories.filter((s) => s.story_id === 'main-colision');
    expect(hits, 'debe aparecer exactamente una vez (sin crash de key)').toHaveLength(1);
    expect(hits[0].is_archived).toBe(true); // prefiere la archivada (canónica)
    expect(hits[0].state).toBe('done');
    expect(hits[0].dup_collision).toBe(true); // marcada para que la UI avise
  });
});
