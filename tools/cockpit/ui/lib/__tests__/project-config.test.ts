/**
 * project-config.ts tests — lector del seam `project.config.yaml` (W5b · F-4).
 *
 * Self-contained: escribe un seam fixture genérico ('acme' / sistema 'main') en un
 * tmpdir y apunta WORKSPACE_ROOT ahí — no depende de ningún workspace real.
 * Cubre: parse + cache, slots per-sistema, slots flat single-sistema (forma del
 * template del installer), __FILL_ME__ → null, y la normalización de etapas
 * para la lente proceso (getValueStreamStages).
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import {
  _resetProjectConfigCache,
  getActiveSistemaSlugs,
  getAgentRoster,
  getProjectConfig,
  getValueStream,
  getValueStreamStages,
  isUnfilled,
} from '../project-config.js';
import { _resetWorkspaceCache } from '../workspace.js';

let root: string;
let prevWorkspaceRoot: string | undefined;

const SEAM_PER_SISTEMA = `meta:
  product: acme
  config_version: 1
  description: producto de prueba
brands:
  cockpit_main_port: 4000
  loop_order: [main]
  active:
    - slug: main
      vertical: ecommerce
      status: active
  pending_bootstrap: []
engine_prefix:
  python_glob: core/acme-core-*
  python_module_prefix: acme_core_
  ts_scope: "@acme/*"
  python_package_count: 3
agent_roster:
  main:
    - slug: alfa
      name: Alfa
      emoji: "🏷"
      color: "#10b981"
    - slug: beta
      name: Beta
      emoji: "📅"
      color: "#3b82f6"
  otra: __FILL_ME__
value_stream:
  main:
    - id: atraer
      name: Atraer
      order: 1
      description: etapa 1
      boxIds: [alfa]
    - id: operar
      name: Operar
      order: 2
      description: etapa 2
      boxIds: [beta, fantasma]
  otra: __FILL_ME__
`;

const SEAM_FLAT = `meta:
  product: acme
  config_version: 1
brands:
  cockpit_main_port: 4000
  loop_order: [main]
  active:
    - slug: main
      vertical: ecommerce
      status: active
  pending_bootstrap: []
engine_prefix:
  python_glob: core/acme-core-*
  python_module_prefix: acme_core_
  ts_scope: "@acme/*"
  python_package_count: 3
agent_roster: __FILL_ME__
value_stream: [captar, convertir, operar, retener]
`;

function writeSeam(yaml: string): void {
  writeFileSync(path.join(root, 'project.config.yaml'), yaml);
  _resetProjectConfigCache();
}

beforeEach(() => {
  root = mkdtempSync(path.join(tmpdir(), 'cockpit-seam-'));
  prevWorkspaceRoot = process.env.WORKSPACE_ROOT;
  process.env.WORKSPACE_ROOT = root;
  _resetWorkspaceCache();
  _resetProjectConfigCache();
});

afterEach(() => {
  if (prevWorkspaceRoot === undefined) delete process.env.WORKSPACE_ROOT;
  else process.env.WORKSPACE_ROOT = prevWorkspaceRoot;
  _resetWorkspaceCache();
  _resetProjectConfigCache();
  rmSync(root, { recursive: true, force: true });
});

describe('project-config (seam class C · fixture genérico)', () => {
  it('parsea meta + engine_prefix del store', () => {
    writeSeam(SEAM_PER_SISTEMA);
    const cfg = getProjectConfig();
    expect(cfg.meta.product).toBe('acme');
    expect(cfg.engine_prefix.python_package_count).toBe(3);
    expect(cfg.engine_prefix.python_glob).toBe('core/acme-core-*');
  });

  it('active sistema slugs vienen de sistemas.active (brands: legacy via shim I-52)', () => {
    writeSeam(SEAM_PER_SISTEMA);
    expect(getActiveSistemaSlugs()).toEqual(['main']);
  });

  it('back-compat I-52: la key nueva `sistemas:` resuelve igual que `brands:` legacy', () => {
    writeSeam(SEAM_FLAT.replace('brands:', 'sistemas:'));
    expect(getActiveSistemaSlugs()).toEqual(['main']);
    expect(getProjectConfig().sistemas.active[0].vertical).toBe('ecommerce');
  });

  it('agent roster per-sistema parsea con colores hex', () => {
    writeSeam(SEAM_PER_SISTEMA);
    const roster = getAgentRoster('main');
    expect(roster).not.toBeNull();
    const alfa = roster!.find((a) => a.slug === 'alfa');
    expect(alfa?.color).toBe('#10b981');
  });

  it('slot per-sistema __FILL_ME__ → null', () => {
    writeSeam(SEAM_PER_SISTEMA);
    expect(getAgentRoster('otra')).toBeNull();
    expect(getValueStream('otra')).toBeNull();
    expect(isUnfilled('__FILL_ME__')).toBe(true);
  });

  it('value-stream per-sistema parsea (atraer primero)', () => {
    writeSeam(SEAM_PER_SISTEMA);
    const vs = getValueStream('main');
    expect(vs).not.toBeNull();
    expect(vs![0].id).toBe('atraer');
  });

  it('getValueStreamStages normaliza + filtra boxIds contra el roster', () => {
    writeSeam(SEAM_PER_SISTEMA);
    const stages = getValueStreamStages('main');
    expect(stages).not.toBeNull();
    expect(stages!.map((s) => s.id)).toEqual(['atraer', 'operar']);
    // 'fantasma' no está en el roster → filtrado (etapa queda con los que mapean)
    expect(stages![1].boxIds).toEqual(['beta']);
  });

  it('slots flat single-sistema (forma del template): roster __FILL_ME__ → null · value_stream string[] → etapas sin agentes', () => {
    writeSeam(SEAM_FLAT);
    expect(getAgentRoster('main')).toBeNull();
    expect(getValueStream('main')).toBeNull(); // solo nombres canónicos
    const stages = getValueStreamStages('main');
    expect(stages).not.toBeNull();
    expect(stages!.map((s) => s.id)).toEqual(['captar', 'convertir', 'operar', 'retener']);
    expect(stages!.every((s) => s.boxIds.length === 0)).toBe(true);
  });
});
