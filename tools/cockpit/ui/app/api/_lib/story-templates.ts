/**
 * Templates para checkpoint.md + operator-input cuando se crean stories nuevas
 * desde el cockpit (extend-cap, from-done).
 *
 * El archivo se sigue creando con el filename del template del kit
 * (`chris-input.md` // legacy F-1) para no romper los skills del kit read-only;
 * el parser/route aceptan también `operator-input.md`.
 *
 * Schema v2 cement 2026-05-27.
 */

import path from 'node:path';
import type { CapChangeType } from '@/lib/types';
import { writeMarkdownWithFrontmatter, writeFileAtomic } from '@/lib/fs-writer';
import { storiesPath } from '@/lib/workspace';

export interface NewStoryInput {
  sistema: string;
  slug: string; // story_id final = slug
  goal: string;
  release: string;
  capTarget?: string | null;
  capChangeType?: CapChangeType | null;
  parentStory?: string | null;
  spawnedBy?: string;
  /** Agente dueño (heredado del cap/story padre) → board pinta franja + emoji. */
  agentOwner?: string | null;
  /** Módulo del cap (heredado del cap/story padre) → ubicación en el mapa. */
  module?: string | null;
}

function nowIso(): string {
  return new Date().toISOString();
}

function todayIso(): string {
  return new Date().toISOString().substring(0, 10);
}

export interface CreatedStoryPaths {
  storyId: string;
  storyDir: string;
  checkpointPath: string;
  operatorInputPath: string;
}

/**
 * Crea checkpoint.md + operator-input en `{sistema}/docs/product/stories/{slug}/`.
 * Falla si el directorio ya existe (no sobrescribe).
 */
export async function createNewStoryDocs(
  input: NewStoryInput
): Promise<CreatedStoryPaths> {
  const storyDir = path.join(storiesPath(input.sistema), input.slug);
  const checkpointPath = path.join(storyDir, 'checkpoint.md');
  const operatorInputPath = path.join(storyDir, 'chris-input.md'); // legacy F-1 (filename del template del kit)

  // Frontmatter checkpoint
  const checkpointFrontmatter: Record<string, unknown> = {
    story_id: input.slug,
    sistema: input.sistema,
    state: 'idea',
    release: input.release,
    cap_target: input.capTarget ?? null,
    cap_change_type: input.capChangeType ?? null,
    parent_story: input.parentStory ?? null,
    agent_owner: input.agentOwner ?? null,
    module: input.module ?? null,
    last_modified: nowIso(),
    spawned_at: todayIso(),
    spawned_by: input.spawnedBy ?? 'cockpit',
    ratified_by_chris: false, // legacy F-1: key del contrato kit
    parallel_safe: true,
    next_action: 'El operador ratifica goal + cap_target · luego /po-ux o /po refina spec',
    goal: input.goal,
  };

  const checkpointBody = `# ${input.slug} — checkpoint

## Goal

${input.goal}

## Estado

Story creada desde cockpit · pending ratificación del operador para empezar refinement.
`;

  const operatorInputFrontmatter: Record<string, unknown> = {
    story_id: input.slug,
    created_at: nowIso(),
    last_modified: nowIso(),
    notes_count: 0,
    refs_count: 0,
    conversation_count: 0,
  };

  // Body espejo del template del kit (00-chris-input-template.md) — heading y
  // pointer al protocol conservan el naming legacy del kit. // legacy F-1
  const operatorInputBody = `# chris-input.md · ${input.slug}

> Cocina de la story (Notas + Referencias + Conversación). Separada de spec/design/arch.
> Template canónico: \`core-harness/templates/00-chris-input-template.md\`.

## 💭 Notas

## 📎 Referencias

(sin referencias todavía)

## 💬 Conversación
`;

  await writeMarkdownWithFrontmatter(
    checkpointPath,
    checkpointFrontmatter,
    checkpointBody
  );
  await writeMarkdownWithFrontmatter(
    operatorInputPath,
    operatorInputFrontmatter,
    operatorInputBody
  );

  return {
    storyId: input.slug,
    storyDir,
    checkpointPath,
    operatorInputPath,
  };
}

/**
 * Appendea una ref de tipo 'story-ref' a la sección Referencias del operator-input
 * de la story recién creada · helper para from-done para registrar parent.
 */
export async function appendStoryRefToOperatorInput(
  operatorInputPath: string,
  refValue: string,
  comment: string
): Promise<void> {
  // Append directo: ya sabemos cómo se ve el template (sección vacía con "(sin referencias todavía)")
  // Para idempotencia + correctitud, leemos + parseamos + serializamos.
  const { parseOperatorInput, serializeOperatorInput } = await import(
    '@/lib/operator-input-parser'
  );
  const { readFile } = await import('node:fs/promises');

  const raw = await readFile(operatorInputPath, 'utf-8');
  const data = parseOperatorInput(raw);
  data.refs.push({
    type: 'story-ref',
    value: refValue,
    comment,
  });
  data.frontmatter.last_modified = nowIso();
  data.frontmatter.refs_count = data.refs.length;
  const serialized = serializeOperatorInput(data);
  await writeFileAtomic(operatorInputPath, serialized);
}
