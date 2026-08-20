/**
 * map-zones.ts — render por zona (SYSTEM-MAP v2.0) + lente proceso.
 * Estrategia: vitest puro (sin React), fixtures mínimas genéricas espejo del
 * schema v2.0 (F-4: cero roster de producto; etapas vienen del seam o del
 * fallback genérico DEFAULT_VALUE_STREAM_STAGES).
 */

import { describe, it, expect } from 'vitest';
import {
  normalizeZoneBoxes,
  boxZoneIndex,
  capsByFunctionalArea,
  buildZoneTree,
  findOrphanCaps,
  findSupervisor,
  buildProcessLens,
  DEFAULT_VALUE_STREAM_STAGES,
  type ValueStreamStage,
} from '../map-zones';
import type {
  AgentDefinition,
  Capability,
  SystemMap,
  SystemMapZone,
} from '../types';

// ── Fixtures mínimas genéricas (espejo SYSTEM-MAP v2.0) ─────────────────────

function area(id: string, status: 'live' | 'planned' = 'live') {
  return { id, name: id, status } as const;
}

const AGENTS: AgentDefinition[] = [
  { id: 'alfa', emoji: '🏷', name: 'Alfa', subtitle: 'Catálogo', functional_areas: [area('marca'), area('servicios')] },
  { id: 'beta', emoji: '📅', name: 'Beta', subtitle: 'Operar', functional_areas: [area('agenda'), area('bookings')] },
  { id: 'gama', emoji: '💼', name: 'Gama', subtitle: 'Vender', functional_areas: [area('embudo')] },
  { id: 'delta', emoji: '📣', name: 'Delta', subtitle: 'Marketing', functional_areas: [area('atribucion')] },
  { id: 'eco', emoji: '🌟', name: 'Eco', subtitle: 'Retención', functional_areas: [area('reputacion')] },
  // supervisor (sin caja de valor)
  { id: 'nucleo', emoji: '🧭', name: 'Núcleo', subtitle: 'Supervisor', role: 'supervisor', status: 'supervisor', functional_areas: [] },
];

// Etapas como las declararía el adopter en `value_stream` del seam
const SEAM_STAGES: ValueStreamStage[] = [
  { id: 'atraer', name: 'Atraer', description: '', boxIds: ['alfa', 'delta'] },
  { id: 'vender', name: 'Vender', description: '', boxIds: ['gama'] },
  { id: 'operar', name: 'Operar', description: '', boxIds: ['beta'] },
  { id: 'fidelizar', name: 'Fidelizar', description: '', boxIds: ['eco'] },
];

const ZONES: SystemMapZone[] = [
  {
    id: 'agentes',
    name: 'Agentes',
    tier: 'core',
    user_visible: true,
    boxes: ['alfa', 'beta', 'gama', 'delta', 'eco'],
  },
  {
    id: 'plataforma',
    name: 'Plataforma',
    tier: 'supporting',
    user_visible: true,
    boxes: [
      { id: 'acceso', name: 'Acceso', functional_areas: [area('auth'), area('iam')] },
      { id: 'configuracion', name: 'Configuración', functional_areas: [area('admin')] },
    ],
  },
  {
    id: 'infraestructura',
    name: 'Infraestructura',
    tier: 'enabling',
    user_visible: false,
    boxes: [
      { id: 'observabilidad', name: 'Observabilidad', functional_areas: [area('observability')] },
    ],
  },
];

const SYSTEM_MAP: Pick<SystemMap, 'zones' | 'agents'> = { zones: ZONES, agents: AGENTS };

function cap(slug: string, functional_area: string | null, extra: Partial<Capability> = {}): Capability {
  return {
    capability_id: `acme-${slug}`,
    module: 'x',
    slug,
    status: 'live',
    license: 'sistema-local',
    created_in_story: 's',
    created_date: '2026-05-30',
    last_modified: '2026-05-30',
    parent_cap: null,
    derives_capabilities: [],
    change_log: [],
    functional_area,
    ...extra,
  } as Capability;
}

const CAPS: Capability[] = [
  cap('marca-1', 'alfa.marca'),
  cap('agenda-1', 'beta.agenda'),
  cap('agenda-2', 'beta.agenda'),
  cap('auth-1', 'acceso.auth'),
  cap('obs-1', 'observabilidad.observability'),
  cap('dead', 'beta.agenda', { superseded_by: 'agenda-1' }), // superseded → ignorada
  cap('orphan-1', 'beta.unknown-area'), // área inexistente → huérfana
  cap('no-fa', null), // sin functional_area → huérfana
];

// ── Tests ───────────────────────────────────────────────────────────────────

describe('normalizeZoneBoxes', () => {
  it('resuelve refs string (agentes) a la definición del agente', () => {
    const boxes = normalizeZoneBoxes(ZONES[0], AGENTS);
    expect(boxes.map((b) => b.id)).toEqual(['alfa', 'beta', 'gama', 'delta', 'eco']);
    const beta = boxes.find((b) => b.id === 'beta')!;
    expect(beta.emoji).toBe('📅');
    expect(beta.functional_areas.map((a) => a.id)).toEqual(['agenda', 'bookings']);
    expect(beta.zoneId).toBe('agentes');
  });

  it('passthrough de cajas objeto (plataforma/infra) con emoji de caja', () => {
    const boxes = normalizeZoneBoxes(ZONES[1], AGENTS);
    expect(boxes.map((b) => b.id)).toEqual(['acceso', 'configuracion']);
    expect(boxes[0].emoji).toBe('🔐');
    expect(boxes[0].functional_areas.map((a) => a.id)).toEqual(['auth', 'iam']);
  });

  it('omite refs string a agentes inexistentes sin romper', () => {
    const z: SystemMapZone = { ...ZONES[0], boxes: ['alfa', 'fantasma'] };
    const boxes = normalizeZoneBoxes(z, AGENTS);
    expect(boxes.map((b) => b.id)).toEqual(['alfa']);
  });

  it('el supervisor NO aparece como caja de la zona agentes', () => {
    const boxes = normalizeZoneBoxes(ZONES[0], AGENTS);
    expect(boxes.map((b) => b.id)).not.toContain('nucleo');
  });
});

describe('boxZoneIndex', () => {
  it('deriva la zona de cada caja', () => {
    const idx = boxZoneIndex(ZONES);
    expect(idx.get('alfa')).toBe('agentes');
    expect(idx.get('acceso')).toBe('plataforma');
    expect(idx.get('observabilidad')).toBe('infraestructura');
    expect(idx.get('inexistente')).toBeUndefined();
  });
  it('tolera zones undefined', () => {
    expect(boxZoneIndex(undefined).size).toBe(0);
  });
});

describe('capsByFunctionalArea', () => {
  it('agrupa por functional_area e ignora superseded y sin-fa', () => {
    const m = capsByFunctionalArea(CAPS);
    expect(m.get('beta.agenda')!.map((c) => c.slug)).toEqual(['agenda-1', 'agenda-2']);
    expect(m.get('alfa.marca')!).toHaveLength(1);
    expect(m.has('beta.unknown-area')).toBe(true); // se agrupa aunque la caja no la tenga
  });
});

describe('buildZoneTree', () => {
  const tree = buildZoneTree(SYSTEM_MAP, capsByFunctionalArea(CAPS));

  it('produce una rama por zona en orden', () => {
    expect(tree.map((z) => z.zone.id)).toEqual(['agentes', 'plataforma', 'infraestructura']);
  });

  it('adjunta caps por fullId `${box}.${area}`', () => {
    const agentes = tree.find((z) => z.zone.id === 'agentes')!;
    const beta = agentes.boxes.find((b) => b.box.id === 'beta')!;
    const agenda = beta.areas.find((a) => a.area.id === 'agenda')!;
    expect(agenda.fullId).toBe('beta.agenda');
    expect(agenda.caps.map((c) => c.slug)).toEqual(['agenda-1', 'agenda-2']);
    expect(beta.totalCaps).toBe(2);
  });

  it('cuenta caps por zona', () => {
    const plataforma = tree.find((z) => z.zone.id === 'plataforma')!;
    expect(plataforma.totalCaps).toBe(1); // acceso.auth
  });

  it('zones ausentes → árbol vacío', () => {
    expect(buildZoneTree({ zones: undefined, agents: AGENTS }, new Map())).toEqual([]);
  });
});

describe('findOrphanCaps', () => {
  it('detecta caps sin caja/área conocida + sin functional_area', () => {
    const tree = buildZoneTree(SYSTEM_MAP, capsByFunctionalArea(CAPS));
    const orphans = findOrphanCaps(CAPS, tree).map((c) => c.slug).sort();
    expect(orphans).toEqual(['no-fa', 'orphan-1']);
  });
});

describe('findSupervisor', () => {
  it('encuentra al supervisor por role/status supervisor', () => {
    expect(findSupervisor(AGENTS)?.id).toBe('nucleo');
  });
  it('null si no hay supervisor', () => {
    expect(findSupervisor(AGENTS.filter((a) => a.id !== 'nucleo'))).toBeNull();
  });
});

describe('buildProcessLens', () => {
  const tree = buildZoneTree(SYSTEM_MAP, capsByFunctionalArea(CAPS));
  const stages = buildProcessLens(tree, SEAM_STAGES);

  it('mapea las cajas agentes a las etapas del value-stream del seam', () => {
    const atraer = stages.find((s) => s.stage.id === 'atraer')!;
    expect(atraer.boxes.map((b) => b.box.id).sort()).toEqual(['alfa', 'delta']);
    const operar = stages.find((s) => s.stage.id === 'operar')!;
    expect(operar.boxes.map((b) => b.box.id)).toEqual(['beta']);
    expect(operar.totalCaps).toBe(2);
  });

  it('cobertura total: toda caja agentes aparece en alguna etapa (anti-isla)', () => {
    const tree = buildZoneTree(SYSTEM_MAP, capsByFunctionalArea(CAPS));
    const agentes = tree.find((z) => z.zone.id === 'agentes')!;
    const allBoxIds = agentes.boxes.map((b) => b.box.id).sort();
    const covered = buildProcessLens(tree, SEAM_STAGES)
      .flatMap((s) => s.boxes.map((b) => b.box.id))
      .sort();
    expect(covered).toEqual(allBoxIds);
  });

  it('cajas agentes sin etapa caen en "Otros"', () => {
    const extra: AgentDefinition = { id: 'alfa', emoji: '🏷', name: 'Alfa', subtitle: '', functional_areas: [] };
    const z: SystemMapZone = { ...ZONES[0], boxes: ['alfa', 'nuevo-agente'] };
    const agentsPlus = [...AGENTS, { ...extra, id: 'nuevo-agente' as never }];
    const tree2 = buildZoneTree({ zones: [z], agents: agentsPlus }, new Map());
    const stages2 = buildProcessLens(tree2, SEAM_STAGES);
    const otros = stages2.find((s) => s.stage.id === 'otros');
    expect(otros?.boxes.map((b) => b.box.id)).toEqual(['nuevo-agente']);
  });

  it('fallback genérico (slot vacío): 4 etapas sin agentes → todo cae en "Otros"', () => {
    const stagesDefault = buildProcessLens(tree); // sin etapas del seam
    expect(stagesDefault.map((s) => s.stage.id)).toEqual([
      'descubrir',
      'construir',
      'operar',
      'mejorar',
      'otros',
    ]);
    expect(DEFAULT_VALUE_STREAM_STAGES.every((s) => s.boxIds.length === 0)).toBe(true);
    const otros = stagesDefault.find((s) => s.stage.id === 'otros')!;
    expect(otros.boxes.map((b) => b.box.id).sort()).toEqual(['alfa', 'beta', 'delta', 'eco', 'gama']);
  });
});
