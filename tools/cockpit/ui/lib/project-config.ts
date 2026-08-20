/**
 * project-config.ts — the cockpit's reader of the harness DIP seam `project.config.yaml`
 * (W5b · 2026-06-09 · consumer class C). SAME single store the python loader
 * (`scripts/harness_config.py`) reads; this is the SECOND parser (the `yaml` pkg, already
 * a cockpit dep) but ONE store — the cross-runtime fixture test
 * (`__tests__/project-config.test.ts`) guards parity so a yaml edit that breaks one runtime
 * fails CI (RESEARCH-loader-mechanism.md Q3 cross-class note).
 *
 * SERVER-ONLY: uses `getWorkspaceRoot()` (node:fs + git). NEVER import from a `'use client'`
 * module. The client-imported roster/zone const modules (`agent-meta.ts`, `map-zones.ts`)
 * adopt the seam via a server-boundary / codegen follow-up (W5b flag) — not by fs-reading
 * in a client bundle.
 *
 * Doctrina: docs/process/harness-refactor-charter-2026-06-08.md §3 (the seam) · §7 (cockpit
 * renders the CORE read-schema).
 */

import { readFileSync } from 'node:fs';
import path from 'node:path';
import YAML from 'yaml';
import { getWorkspaceRoot } from './workspace';

export const FILL_SENTINEL = '__FILL_ME__';

export interface SistemaPorts {
  backend: number;
  frontend: number;
  cockpit: number;
}

export interface SistemaEntry {
  slug: string;
  vertical: string;
  status: 'active' | 'shipped' | 'placeholder' | 'pending-bootstrap' | 'blocked';
  ports?: SistemaPorts;
  dev_app_url?: string;
  compliance?: string;
}

export interface AgentMeta {
  slug: string;
  name: string;
  emoji: string;
  color: string;
  role?: string;
  subtitle?: string;
}

export interface ValueStreamStage {
  id: string;
  name: string;
  order: number;
  description: string;
  boxIds: string[];
}

export interface ProjectConfig {
  meta: { product: string; config_version: number; description?: string };
  // El seam (project.config.yaml) puede declarar la key nueva `sistemas:` (I-52) o la
  // legacy `brands:` (seams pre-rename) — getProjectConfig() normaliza ambas (el loader python igual).
  sistemas: {
    cockpit_main_port: number;
    loop_order: string[];
    active: SistemaEntry[];
    pending_bootstrap: { slug: string; vertical: string }[];
  };
  engine_prefix: {
    python_glob: string;
    python_module_prefix: string;
    ts_scope: string;
    python_package_count: number;
    [k: string]: unknown;
  };
  // slots: per-sistema record, flat single-sistema (forma del template del installer),
  // o __FILL_ME__ (string) hasta que el adopter los llene
  agent_roster: Record<string, AgentMeta[] | string> | AgentMeta[] | string;
  value_stream:
    | Record<string, ValueStreamStage[] | string[] | string>
    | ValueStreamStage[]
    | string[]
    | string;
  [slot: string]: unknown;
}

let cache: ProjectConfig | null = null;

/** Read + parse + cache the single seam store. Module-level cache (long-lived dev process). */
export function getProjectConfig(refresh = false): ProjectConfig {
  if (cache && !refresh) return cache;
  const raw = readFileSync(path.join(getWorkspaceRoot(), 'project.config.yaml'), 'utf-8');
  const parsed = YAML.parse(raw) as ProjectConfig;
  // Back-compat (I-52): seams pre-rename declaran `brands:` en vez de `sistemas:`.
  const legacy = (parsed as Record<string, unknown>).brands;
  if (parsed.sistemas == null && legacy != null) {
    parsed.sistemas = legacy as ProjectConfig['sistemas'];
  }
  cache = parsed;
  return cache;
}

export function isUnfilled(value: unknown): boolean {
  return value === FILL_SENTINEL;
}

/**
 * A sistema's agent roster, or null when the slot is unfilled (`__FILL_ME__`).
 * Acepta slot per-sistema (`agent_roster: {main: [...]}`) o flat single-sistema
 * (`agent_roster: [...]` — forma del template del installer).
 */
export function getAgentRoster(sistema: string): AgentMeta[] | null {
  const slot = getProjectConfig().agent_roster;
  if (Array.isArray(slot)) return slot as AgentMeta[]; // flat single-sistema
  if (!slot || typeof slot !== 'object') return null;
  const r = (slot as Record<string, AgentMeta[] | string>)[sistema];
  return Array.isArray(r) ? (r as AgentMeta[]) : null;
}

/** A sistema's value-stream stages, or null when unfilled / only canonical names. */
export function getValueStream(sistema: string): ValueStreamStage[] | null {
  const slot = getProjectConfig().value_stream;
  const v = Array.isArray(slot)
    ? slot // flat single-sistema (forma del template del installer)
    : slot && typeof slot === 'object'
      ? (slot as Record<string, ValueStreamStage[] | string[] | string>)[sistema]
      : null;
  if (Array.isArray(v) && v.length > 0 && typeof v[0] === 'object') {
    return v as ValueStreamStage[];
  }
  return null;
}

/**
 * Etapas del value-stream NORMALIZADAS para la lente "proceso" del mapa:
 *  - etapas objeto → boxIds validados contra el `agent_roster` (si hay roster);
 *  - lista de nombres canónicos (`[captar, convertir, ...]`) → etapas sin agentes;
 *  - slot vacío / __FILL_ME__ → null (el caller usa el fallback genérico).
 */
export function getValueStreamStages(
  sistema: string
): { id: string; name: string; description: string; boxIds: string[] }[] | null {
  const slot = getProjectConfig().value_stream;
  const v = Array.isArray(slot)
    ? slot
    : slot && typeof slot === 'object'
      ? (slot as Record<string, ValueStreamStage[] | string[] | string>)[sistema]
      : null;
  if (!Array.isArray(v) || v.length === 0) return null;

  const rosterSlugs = new Set((getAgentRoster(sistema) ?? []).map((a) => a.slug));

  if (typeof v[0] === 'string') {
    // Solo nombres canónicos → etapas sin agentes asignados
    return (v as string[]).map((name) => ({
      id: String(name).toLowerCase().replace(/\s+/g, '-'),
      name: String(name),
      description: '',
      boxIds: [],
    }));
  }

  return (v as ValueStreamStage[])
    .slice()
    .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
    .map((s) => ({
      id: s.id,
      name: s.name,
      description: s.description ?? '',
      // boxIds del seam · si hay roster, solo los que mapean a un agente real
      boxIds: (s.boxIds ?? []).filter((id) => rosterSlugs.size === 0 || rosterSlugs.has(id)),
    }));
}

export function getActiveSistemaSlugs(): string[] {
  return getProjectConfig().sistemas.active.map((b) => b.slug);
}

/** Reset cache · for tests. */
export function _resetProjectConfigCache(): void {
  cache = null;
}
