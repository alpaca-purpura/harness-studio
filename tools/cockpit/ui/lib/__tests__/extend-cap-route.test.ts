/**
 * Tests para POST /api/extend-cap
 *
 * Regresión bug 2026-05-30: módulos con guion bajo (sistema_studio, sales_agent,
 * offer_studio, public_landing) eran rechazados por el SLUG_REGEX kebab-only
 * → "body inválido / slug inválido" + toast "Error al guardar" en el cockpit.
 *
 * Estrategia: mock de las deps de la route (cap-ledger, workspace,
 * story-templates, node:fs/promises) para aislar la validación + el wiring.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// ────────────────────────────────────────────────────────────────────────────
// Mocks
// ────────────────────────────────────────────────────────────────────────────

vi.mock('@/lib/workspace', () => ({
  getSistemas: vi.fn(() => ['main', 'acme']),
  capabilitiesPath: vi.fn((sistema: string) => `/fake/${sistema}/capabilities`),
  storiesPath: vi.fn((sistema: string) => `/fake/${sistema}/stories`),
}));

vi.mock('@/lib/cap-ledger', () => ({
  readCapability: vi.fn(async () => ({
    id: 'fake-cap',
    module: 'sistema_studio',
    agent_owner: 'lisa',
  })),
}));

vi.mock('../../app/api/_lib/story-templates', () => ({
  createNewStoryDocs: vi.fn(async (input: { slug: string }) => ({
    storyId: input.slug,
    storyDir: `/fake/stories/${input.slug}`,
    checkpointPath: `/fake/stories/${input.slug}/checkpoint.md`,
    operatorInputPath: `/fake/stories/${input.slug}/chris-input.md`, // legacy F-1 filename
  })),
}));

vi.mock('node:fs/promises', () => ({
  // stat se usa para verificar que la story NO existe ya → rechazar = ENOENT
  stat: vi.fn(async () => {
    throw Object.assign(new Error('ENOENT'), { code: 'ENOENT' });
  }),
}));

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

interface ExtendCapBody {
  sistema: string;
  parentCap: { module: string; slug: string };
  capChangeType: 'fix' | 'extend' | 'derive';
  newStorySlug: string;
  goal: string;
  release: string;
  derivedName?: string;
}

function makeRequest(body: ExtendCapBody): Request {
  return new Request('http://localhost:3000/api/extend-cap', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

async function getHandler() {
  const mod = await import('../../app/api/extend-cap/route.js');
  return mod.POST;
}

function baseBody(overrides: Partial<ExtendCapBody> = {}): ExtendCapBody {
  return {
    sistema: 'main',
    parentCap: { module: 'sistema_studio', slug: 'lisa-marca' },
    capChangeType: 'extend',
    newStorySlug: 'main-marca-multi-sede',
    goal: 'Extender lisa-marca para soportar multiples clinicas en el mismo tenant',
    release: 'F2',
    ...overrides,
  };
}

// ────────────────────────────────────────────────────────────────────────────
// Tests
// ────────────────────────────────────────────────────────────────────────────

describe('POST /api/extend-cap', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('test_accepts_module_with_underscore', async () => {
    // sistema_studio tiene guion bajo → antes rompía con "body inválido".
    const POST = await getHandler();
    const res = await POST(makeRequest(baseBody()) as Parameters<typeof POST>[0]);
    const body = (await res.json()) as { storyId?: string; error?: string };

    expect(res.status).toBe(200);
    expect(body.error).toBeUndefined();
    expect(body.storyId).toBe('main-marca-multi-sede');
  });

  it('test_accepts_cap_slug_with_underscore', async () => {
    // cap slug legacy con guion bajo (wizard_sistema_studio_slice_1).
    const POST = await getHandler();
    const res = await POST(
      makeRequest(
        baseBody({
          parentCap: { module: 'sistema_studio', slug: 'wizard_sistema_studio_slice_1' },
        })
      ) as Parameters<typeof POST>[0]
    );
    expect(res.status).toBe(200);
  });

  it('test_accepts_all_underscore_modules', async () => {
    const POST = await getHandler();
    for (const mod of ['sales_agent', 'offer_studio', 'public_landing']) {
      const res = await POST(
        makeRequest(
          baseBody({ parentCap: { module: mod, slug: 'algun-cap' } })
        ) as Parameters<typeof POST>[0]
      );
      expect(res.status, `module=${mod}`).toBe(200);
    }
  });

  it('test_rejects_module_with_invalid_char', async () => {
    // un espacio o mayuscula sigue siendo invalido (no es barra libre).
    const POST = await getHandler();
    const res = await POST(
      makeRequest(
        baseBody({ parentCap: { module: 'sistema studio', slug: 'lisa-marca' } })
      ) as Parameters<typeof POST>[0]
    );
    expect(res.status).toBe(400);
  });

  it('test_rejects_story_slug_with_underscore', async () => {
    // las stories nuevas siguen kebab-case → underscore rechazado.
    const POST = await getHandler();
    const res = await POST(
      makeRequest(
        baseBody({ newStorySlug: 'main_marca_invalida' })
      ) as Parameters<typeof POST>[0]
    );
    expect(res.status).toBe(400);
  });

  it('test_derive_requires_derived_name', async () => {
    const POST = await getHandler();
    const res = await POST(
      makeRequest(
        baseBody({ capChangeType: 'derive', derivedName: undefined })
      ) as Parameters<typeof POST>[0]
    );
    const body = (await res.json()) as { error: string };
    expect(res.status).toBe(400);
    expect(body.error).toContain('derivedName');
  });

  it('test_extend_sets_cap_target_to_parent_slug', async () => {
    const { createNewStoryDocs } = await import('../../app/api/_lib/story-templates.js');
    const POST = await getHandler();
    await POST(makeRequest(baseBody()) as Parameters<typeof POST>[0]);

    expect(createNewStoryDocs).toHaveBeenCalledWith(
      expect.objectContaining({
        sistema: 'main',
        capTarget: 'lisa-marca',
        capChangeType: 'extend',
        parentStory: null,
        spawnedBy: 'cockpit-extend-cap',
      })
    );
  });

  it('test_propagates_agent_owner_and_module_from_parent_cap', async () => {
    // Regresión: la story creada debe heredar agent_owner + module del cap padre
    // → el board pinta el agente/módulo en vez de "—".
    const { createNewStoryDocs } = await import('../../app/api/_lib/story-templates.js');
    const POST = await getHandler();
    await POST(makeRequest(baseBody()) as Parameters<typeof POST>[0]);

    expect(createNewStoryDocs).toHaveBeenCalledWith(
      expect.objectContaining({
        agentOwner: 'lisa',
        module: 'sistema_studio',
      })
    );
  });
});
