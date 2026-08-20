/**
 * Tests del reader de build-claims (ADR-009 single-hub).
 * Crea un `.session-locks/` temporal con WORKSPACE_ROOT override y verifica el
 * parsing + filtrado de PIDs muertos + el map story→sesión.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { readActiveSessions, sessionsByStory } from '../sessions';
import { _resetWorkspaceCache } from '../workspace';

let root: string;

function writeLock(file: string, line: string): void {
  writeFileSync(path.join(root, '.session-locks', file), `${line}\n`, 'utf-8');
}

beforeEach(() => {
  root = mkdtempSync(path.join(tmpdir(), 'cockpit-locks-'));
  mkdirSync(path.join(root, '.session-locks'));
  process.env.WORKSPACE_ROOT = root;
  _resetWorkspaceCache();
});

afterEach(() => {
  delete process.env.WORKSPACE_ROOT;
  _resetWorkspaceCache();
  rmSync(root, { recursive: true, force: true });
});

describe('readActiveSessions', () => {
  it('parsea un build-claim vivo (PID actual) con story + lane + bucket', async () => {
    writeLock(
      'code__scheduling.lock',
      `${process.pid} dev-team 2026-05-28T23:00:00-05:00 main-fase2-embudo A code:scheduling`
    );
    const sessions = await readActiveSessions();
    expect(sessions).toHaveLength(1);
    expect(sessions[0]).toMatchObject({
      pid: process.pid,
      skill: 'dev-team',
      storyId: 'main-fase2-embudo',
      lane: 'A',
      bucket: 'code:scheduling',
      startedAt: '2026-05-28T23:00:00-05:00',
    });
  });

  it('filtra locks con PID muerto', async () => {
    // PID 2^31-ish improbable que exista
    writeLock('docs.lock', `2147480000 po-ux 2026-05-28T23:00:00-05:00 — — docs`);
    const sessions = await readActiveSessions();
    expect(sessions).toHaveLength(0);
  });

  it('reconstruye el bucket desde el filename si el campo falta (lock legacy 3-campos)', async () => {
    writeLock('code__crm.lock', `${process.pid} dev-team 2026-05-28T23:00:00-05:00`);
    const sessions = await readActiveSessions();
    expect(sessions).toHaveLength(1);
    expect(sessions[0].bucket).toBe('code:crm');
    expect(sessions[0].storyId).toBeNull();
    expect(sessions[0].lane).toBeNull();
  });

  it('devuelve [] si no existe .session-locks/', async () => {
    rmSync(path.join(root, '.session-locks'), { recursive: true, force: true });
    expect(await readActiveSessions()).toEqual([]);
  });
});

describe('sessionsByStory', () => {
  it('mapea story_id → sesión y omite locks sin story (docs)', async () => {
    writeLock(
      'code__scheduling.lock',
      `${process.pid} dev-team 2026-05-28T23:00:00-05:00 story-a A code:scheduling`
    );
    writeLock(
      'docs.lock',
      `${process.pid} po-ux 2026-05-28T23:00:00-05:00 — operador docs`
    );
    const map = await sessionsByStory();
    expect(Object.keys(map)).toEqual(['story-a']);
    expect(map['story-a'].lane).toBe('A');
  });
});
