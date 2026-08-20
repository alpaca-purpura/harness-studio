/**
 * cap-ledger.ts tests · 4 ramas (new/fix/extend/derive).
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import {
  applyCapChange,
  createDerivedCap,
  readCapability,
} from '../cap-ledger.js';
import type { ChangeLogEntry } from '../types.js';

let tmpDir: string;

beforeEach(async () => {
  tmpDir = await mkdtemp(path.join(os.tmpdir(), 'cockpit-cap-ledger-'));
});

afterEach(async () => {
  await rm(tmpDir, { recursive: true, force: true });
});

describe('applyCapChange', () => {
  it('type=new crea archivo con change_log[0] y scenarios_added', async () => {
    const capPath = path.join(tmpDir, 'caps', 'scheduling', 'valeria-agenda.yaml');
    const entry: ChangeLogEntry = {
      story_id: 'main-fase2-agenda',
      date: '2026-05-27',
      type: 'new',
      summary: 'Implementación inicial · vista calendario + drag-to-reschedule',
      scenarios_added: ['Vista calendario semanal', 'Drag-to-reschedule'],
      merge_sha: '4562140c',
      status: 'done',
    };

    await applyCapChange({
      capPath,
      entry,
      initialCap: {
        capability_id: 'main.scheduling.agenda',
        module: 'scheduling',
        slug: 'valeria-agenda',
        status: 'live',
        license: 'sistema-local',
        architecture_pattern: 'ADR-004',
      },
    });

    const cap = await readCapability(capPath);
    expect(cap.capability_id).toBe('main.scheduling.agenda');
    expect(cap.module).toBe('scheduling');
    expect(cap.slug).toBe('valeria-agenda');
    expect(cap.created_in_story).toBe('main-fase2-agenda');
    expect(cap.created_date).toBe('2026-05-27');
    expect(cap.last_modified).toBe('2026-05-27');
    expect(cap.change_log).toHaveLength(1);
    expect(cap.change_log[0].type).toBe('new');
    expect(cap.change_log[0].scenarios_added).toEqual([
      'Vista calendario semanal',
      'Drag-to-reschedule',
    ]);
    expect(cap.change_log[0].merge_sha).toBe('4562140c');
  });

  it('type=fix appendea change_log con scenarios_added vacío', async () => {
    const capPath = path.join(tmpDir, 'caps', 'shell', 'layout-5050.yaml');
    // Seed: crear cap base
    await applyCapChange({
      capPath,
      entry: {
        story_id: 'main-fase1-shell-layout-5050',
        date: '2026-05-20',
        type: 'new',
        summary: 'Layout shell 50/50 inicial',
        scenarios_added: ['split 50/50'],
        status: 'done',
      },
      initialCap: {
        capability_id: 'main.shell.layout-5050',
        module: 'shell',
        slug: 'layout-5050',
      },
    });

    // Aplicar fix
    await applyCapChange({
      capPath,
      entry: {
        story_id: 'main-fase1-shell-layout-5050-race-fix',
        date: '2026-05-25',
        type: 'fix',
        summary: 'Fix race condition on resize',
        scenarios_added: [],
        status: 'done',
      },
    });

    const cap = await readCapability(capPath);
    expect(cap.change_log).toHaveLength(2);
    expect(cap.change_log[1].type).toBe('fix');
    expect(cap.change_log[1].scenarios_added).toEqual([]);
    expect(cap.last_modified).toBe('2026-05-25');
  });

  it('type=extend appendea change_log con scenarios_added listados', async () => {
    const capPath = path.join(tmpDir, 'caps', 'lisa', 'marca.yaml');
    // Seed: cap base lisa.marca
    await applyCapChange({
      capPath,
      entry: {
        story_id: 'main-fase2-marca',
        date: '2026-05-24',
        type: 'new',
        summary: 'Lisa marca · skin tokens + voice',
        scenarios_added: ['Skin tokens'],
        status: 'done',
      },
      initialCap: {
        capability_id: 'main.alfa.marca',
        module: 'lisa',
        slug: 'marca',
      },
    });

    // Extend
    await applyCapChange({
      capPath,
      entry: {
        story_id: 'main-fase2-marca-v2',
        date: '2026-05-27',
        type: 'extend',
        summary: 'Color extractor + logo upload',
        scenarios_added: ['Color extractor', 'Logo upload'],
        status: 'done',
      },
    });

    const cap = await readCapability(capPath);
    expect(cap.change_log).toHaveLength(2);
    expect(cap.change_log[1].type).toBe('extend');
    expect(cap.change_log[1].scenarios_added).toEqual(['Color extractor', 'Logo upload']);
    expect(cap.last_modified).toBe('2026-05-27');
  });
});

describe('createDerivedCap', () => {
  it('crea cap hijo + actualiza derives_capabilities[] del padre', async () => {
    // 1. Crear cap padre
    const parentPath = path.join(tmpDir, 'caps', 'scheduling', 'valeria-agenda.yaml');
    await applyCapChange({
      capPath: parentPath,
      entry: {
        story_id: 'main-fase2-agenda',
        date: '2026-05-27',
        type: 'new',
        summary: 'Valeria agenda base',
        scenarios_added: ['Calendar view'],
        status: 'done',
      },
      initialCap: {
        capability_id: 'main.scheduling.agenda',
        module: 'scheduling',
        slug: 'valeria-agenda',
        license: 'sistema-local',
        architecture_pattern: 'ADR-004',
      },
    });

    // 2. Derive
    const childPath = path.join(tmpDir, 'caps', 'scheduling', 'valeria-agenda-mobile.yaml');
    await createDerivedCap({
      parentCapPath: parentPath,
      childCapPath: childPath,
      childCap: {
        capability_id: 'main.scheduling.agenda-mobile',
        module: 'scheduling',
        slug: 'valeria-agenda-mobile',
      },
      spawnedFromStory: 'main-fase3-agenda-mobile',
      date: '2026-06-15',
    });

    // 3. Verificar hijo
    const child = await readCapability(childPath);
    expect(child.parent_cap).toBe('valeria-agenda');
    expect(child.slug).toBe('valeria-agenda-mobile');
    expect(child.change_log).toHaveLength(1);
    expect(child.change_log[0].type).toBe('derive');
    expect(child.change_log[0].scenarios_added).toEqual([]);
    expect(child.created_in_story).toBe('main-fase3-agenda-mobile');
    expect(child.architecture_pattern).toBe('ADR-004'); // heredado del padre

    // 4. Verificar padre actualizado
    const parent = await readCapability(parentPath);
    expect(parent.derives_capabilities).toContain('valeria-agenda-mobile');
    expect(parent.last_modified).toBe('2026-06-15');
  });
});

describe('readCapability · passthrough dimensiones v3.2 (F0 cap-levels)', () => {
  // El cap-drawer del cockpit lee scenarios/business_rules/access/related_capabilities
  // vía readCapability. Antes del fix F0, el loader las dropeaba → el drawer mostraba
  // 0 casos de uso / 0 reglas aunque el YAML las tuviera. Este test las blinda.
  it('preserva scenarios, business_rules, access y related_capabilities del YAML', async () => {
    const capDir = path.join(tmpDir, 'caps', 'inbox');
    await mkdir(capDir, { recursive: true });
    const capPath = path.join(capDir, 'adrian-inbox.yaml');
    const yaml = [
      '---',
      'capability_id: main.inbox.inbox',
      'module: inbox',
      'slug: adrian-inbox',
      'status: live',
      'user_visible: true',
      'user_facing_description: Inbox unificado cross-canal.',
      'scenarios:',
      '  - id: operador-pausa-conversacion',
      '    name: El operador pausa una conversación',
      '    actor: receptionist',
      '    status: live',
      '    given: Conversación activa en modo decide',
      '    when: Clic en Pausar 60 minutos',
      '    then: handler_mode pasa a paused y pause_until queda seteado',
      '    e2e_test: e2e/shell-organism/adrian-inbox-modes.spec.ts',
      '    added_in_story: main-fase2-inbox',
      '    added_date: "2026-06-04"',
      'business_rules:',
      '  - id: inbox-mutation-commits',
      '    rule: Toda mutación del inbox debe commitear',
      '    enforcement: [get_async_session_committing]',
      '    code_ref: backend/src/modules/inbox/api/router.py',
      '    severity: high',
      'access:',
      '  entry_points:',
      '    - path: /adrian/inbox',
      '      navigation: Login entonces ribbon Adrián entonces Inbox',
      '      requires_role: [owner, receptionist]',
      'related_capabilities:',
      '  depends_on: [shell-organism.shell-main]',
      '---',
      'Body resumen del cap.',
      '',
    ].join('\n');
    await writeFile(capPath, yaml, 'utf-8');

    const cap = await readCapability(capPath);

    expect(cap.scenarios).toHaveLength(1);
    expect(cap.scenarios?.[0].id).toBe('operador-pausa-conversacion');
    expect(cap.scenarios?.[0].actor).toBe('receptionist');
    expect(cap.business_rules).toHaveLength(1);
    expect(cap.business_rules?.[0].code_ref).toContain('router.py');
    expect(cap.business_rules?.[0].enforcement).toContain('get_async_session_committing');
    expect(cap.access?.entry_points[0].path).toBe('/adrian/inbox');
    expect(cap.access?.entry_points[0].requires_role).toContain('receptionist');
    expect(cap.related_capabilities?.depends_on).toContain('shell-organism.shell-main');
  });
});
