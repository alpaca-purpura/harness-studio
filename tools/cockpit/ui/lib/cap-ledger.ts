/**
 * Capability ledger logic — schema v2 (cement 2026-05-27).
 *
 * 4 ramas según `cap_change_type`:
 *   - new      → crea cap YAML inicial + change_log[0]
 *   - fix      → append change_log entry
 *   - extend   → append change_log entry (scenarios se materializan vía skill al merge)
 *   - derive   → crea cap hijo declarando parent_cap + actualiza derives_capabilities[] del padre
 *
 * Doc canónico: docs/process/capability-protocol.md.
 */

import { readFile } from 'node:fs/promises';
import path from 'node:path';
import matter from 'gray-matter';
import type {
  Capability,
  ChangeLogEntry,
  CapChangeType,
} from './types';
import { writeFileAtomic } from './fs-writer';

// ────────────────────────────────────────────────────────────────────────────
// Read / write capability YAML
// ────────────────────────────────────────────────────────────────────────────

/**
 * Lee un capability YAML. Acepta tanto archivos puros YAML como markdown con
 * frontmatter — el cap YAML estándar viene como markdown con frontmatter +
 * body de resumen, así que parseamos ambos formatos.
 */
export async function readCapability(absPath: string): Promise<Capability> {
  const raw = await readFile(absPath, 'utf-8');
  const parsed = matter(raw);
  const cap = parsed.data as Partial<Capability>;
  // Defaults defensivos (legacy files podían tener fields ausentes)
  return {
    capability_id: cap.capability_id ?? '',
    module: cap.module ?? '',
    slug: cap.slug ?? '',
    status: cap.status ?? 'live',
    license: cap.license ?? 'sistema-local',
    created_in_story: cap.created_in_story ?? cap.story_introduced ?? '',
    created_date: cap.created_date ?? cap.date_introduced ?? '',
    last_modified: cap.last_modified ?? cap.date_updated ?? cap.created_date ?? '',
    package_version: cap.package_version ?? null,
    package_path: cap.package_path ?? null,
    architecture_pattern: cap.architecture_pattern ?? null,
    hipaa_lite_overlay: cap.hipaa_lite_overlay ?? false,
    parent_cap: cap.parent_cap ?? cap.extends_capability ?? null,
    derives_capabilities: cap.derives_capabilities ?? [],
    change_log: (cap.change_log ?? []) as ChangeLogEntry[],
    date_introduced: cap.date_introduced ?? null,
    story_introduced: cap.story_introduced ?? null,
    date_updated: cap.date_updated ?? null,
    extends_capability: cap.extends_capability ?? null,
    tech_module: cap.tech_module ?? cap.module ?? null,
    agent_owner: cap.agent_owner ?? null,
    functional_area: cap.functional_area ?? null,
    user_visible: cap.user_visible ?? true,
    nature: cap.nature ?? null,
    user_facing_name: cap.user_facing_name ?? null,
    user_facing_description: cap.user_facing_description ?? null,
    // v3.2 cap blocks (F0 cap-levels): el cap-drawer del cockpit los lee para
    // mostrar casos de uso (scenarios) + reglas de negocio + acceso + dependencias.
    // Sin este passthrough el drawer mostraba 0 aunque el YAML los tuviera.
    access: cap.access ?? null,
    scenarios: cap.scenarios ?? null,
    business_rules: cap.business_rules ?? null,
    related_capabilities: cap.related_capabilities ?? null,
    dev_preview: cap.dev_preview ?? null,
    superseded_by: cap.superseded_by ?? null,
    body: parsed.content || '',
    path: absPath,
  };
}

/** Serializa cap a markdown con frontmatter (atomic write) */
export async function writeCapability(cap: Capability): Promise<void> {
  if (!cap.path) throw new Error('cap.path requerido para writeCapability');
  // Quitar fields helpers que no son parte del schema persistido
  const { body, path: _, ...frontmatter } = cap;
  const serialized = matter.stringify(body ?? '', frontmatter as Record<string, unknown>);
  await writeFileAtomic(cap.path, serialized);
}

// ────────────────────────────────────────────────────────────────────────────
// applyCapChange — 3 ramas (new/fix/extend) sobre un cap target existente o nuevo
// ────────────────────────────────────────────────────────────────────────────

export interface ApplyCapChangeOptions {
  /** Path absoluto al cap YAML (existe para fix/extend · puede no existir para new) */
  capPath: string;
  /** Entry a appendear al change_log */
  entry: ChangeLogEntry;
  /**
   * Capability inicial · solo se usa cuando type=new (file no existe aún)
   * Debe incluir capability_id, module, slug, license, etc.
   */
  initialCap?: Partial<Capability>;
}

/**
 * Aplica un cambio al cap YAML según el `entry.type`.
 *
 * - new    → crea archivo + change_log[0]
 * - fix    → append change_log
 * - extend → append change_log (scenarios se materializan vía skill al merge)
 * - derive → caso especial: usa `createDerivedCap` en lugar de esta función
 */
export async function applyCapChange(options: ApplyCapChangeOptions): Promise<void> {
  const { capPath, entry, initialCap } = options;

  if (entry.type === 'derive') {
    throw new Error(
      `applyCapChange no soporta type='derive' · usa createDerivedCap en su lugar`
    );
  }

  if (entry.type === 'new') {
    // Crear cap nuevo desde cero
    if (!initialCap) {
      throw new Error("type='new' requiere initialCap con metadata mínima");
    }
    const cap: Capability = {
      capability_id: initialCap.capability_id ?? '',
      module: initialCap.module ?? '',
      slug: initialCap.slug ?? '',
      status: initialCap.status ?? 'live',
      license: initialCap.license ?? 'sistema-local',
      created_in_story: entry.story_id,
      created_date: entry.date,
      last_modified: entry.date,
      package_version: initialCap.package_version ?? null,
      package_path: initialCap.package_path ?? null,
      architecture_pattern: initialCap.architecture_pattern ?? null,
      hipaa_lite_overlay: initialCap.hipaa_lite_overlay ?? false,
      parent_cap: initialCap.parent_cap ?? null,
      derives_capabilities: initialCap.derives_capabilities ?? [],
      change_log: [
        {
          ...entry,
          scenarios_added: entry.scenarios_added ?? [],
        },
      ],
      body: initialCap.body ?? `# ${initialCap.capability_id ?? initialCap.slug ?? ''}\n`,
      path: capPath,
    };
    await writeCapability(cap);
    return;
  }

  // fix o extend · cap target debe existir
  const cap = await readCapability(capPath);

  // Ni fix ni extend tocan scenarios desde el cockpit: los scenarios se
  // materializan vía la skill al merge (Fase F.3). El cockpit solo registra
  // la intención en el change_log.
  cap.change_log.push({
    ...entry,
    scenarios_added: entry.scenarios_added ?? [],
  });

  cap.last_modified = entry.date;
  await writeCapability(cap);
}

// ────────────────────────────────────────────────────────────────────────────
// createDerivedCap — caso especial type=derive
// ────────────────────────────────────────────────────────────────────────────

export interface CreateDerivedCapOptions {
  /** Path absoluto al cap padre (debe existir) */
  parentCapPath: string;
  /** Path absoluto donde se creará el cap hijo */
  childCapPath: string;
  /** Metadata mínima del cap hijo */
  childCap: Partial<Capability> & {
    capability_id: string;
    module: string;
    slug: string;
  };
  /** Story que spawnea el derive · llena change_log[0] del hijo */
  spawnedFromStory: string;
  /** Fecha del cambio · default today */
  date?: string;
  /** Summary del change_log[0] del hijo · default genérico */
  summary?: string;
  /** SHA del merge · opcional */
  mergeSha?: string | null;
}

/**
 * Crea un cap derivado:
 * 1. Lee cap padre (debe existir)
 * 2. Crea cap hijo con parent_cap declarado + change_log[0] type='derive'
 * 3. Update padre: append derives_capabilities[hijo_slug] + last_modified
 *
 * Order matters: crear hijo PRIMERO (con parent_cap declarado), luego update padre.
 * Si falla la actualización del padre, el hijo queda creado pero el padre no
 * lo referencia · auditor flagea inconsistencia.
 */
export async function createDerivedCap(options: CreateDerivedCapOptions): Promise<void> {
  const {
    parentCapPath,
    childCapPath,
    childCap,
    spawnedFromStory,
    date,
    summary,
    mergeSha,
  } = options;

  // 1. Validar padre existe
  const parent = await readCapability(parentCapPath);
  if (!parent.capability_id) {
    throw new Error(`Cap padre en ${parentCapPath} es inválido (sin capability_id)`);
  }

  const effectiveDate = date ?? todayIso();

  // 2. Crear cap hijo
  const child: Capability = {
    capability_id: childCap.capability_id,
    module: childCap.module,
    slug: childCap.slug,
    status: childCap.status ?? 'live',
    license: childCap.license ?? parent.license,
    created_in_story: spawnedFromStory,
    created_date: effectiveDate,
    last_modified: effectiveDate,
    package_version: childCap.package_version ?? null,
    package_path: childCap.package_path ?? null,
    architecture_pattern: childCap.architecture_pattern ?? parent.architecture_pattern ?? null,
    hipaa_lite_overlay: childCap.hipaa_lite_overlay ?? parent.hipaa_lite_overlay ?? false,
    parent_cap: parent.slug,
    derives_capabilities: childCap.derives_capabilities ?? [],
    change_log: [
      {
        story_id: spawnedFromStory,
        date: effectiveDate,
        type: 'derive',
        summary:
          summary ??
          `Cap derivada de ${parent.slug} · scope diferenciado · cementada en story ${spawnedFromStory}`,
        scenarios_added: [],
        merge_sha: mergeSha ?? null,
        status: 'done',
      },
    ],
    body: childCap.body ?? `# ${childCap.capability_id}\n\nDerivada de \`${parent.slug}\`.\n`,
    path: childCapPath,
  };
  await writeCapability(child);

  // 3. Actualizar padre
  if (!parent.derives_capabilities.includes(child.slug)) {
    parent.derives_capabilities.push(child.slug);
  }
  parent.last_modified = effectiveDate;
  await writeCapability(parent);
}

function todayIso(): string {
  return new Date().toISOString().substring(0, 10);
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers públicos
// ────────────────────────────────────────────────────────────────────────────

/** Path canónico de un capability YAML dado sistema + module + slug */
export function capabilityFilePath(
  sistema: string,
  module: string,
  slug: string,
  sistemaProductPath: string
): string {
  return path.join(sistemaProductPath, 'capabilities', module, `${slug}.yaml`);
}

/** Devuelve true si el cap_change_type es uno de los 4 válidos */
export function isValidCapChangeType(value: string): value is CapChangeType {
  return value === 'new' || value === 'fix' || value === 'extend' || value === 'derive';
}
