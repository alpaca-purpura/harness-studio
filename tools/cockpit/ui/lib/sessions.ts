/**
 * Active sessions reader — lee `.session-locks/*.lock` del workspace root y los
 * mapea a build-claims vivos (ADR-009 single-hub worktree).
 *
 * Cada lock lo escribe `scripts/git/session-lock.sh acquire` con el formato:
 *
 *   PID SKILL TIMESTAMP STORY_ID LANE BUCKET
 *
 * (los locks legacy de 3 campos `PID SKILL TIMESTAMP` siguen parseando — story_id,
 * lane y bucket quedan en `null`).
 *
 * Un lock cuyo PID ya no está vivo se considera muerto y se omite (el script lo
 * auto-libera en el próximo `acquire`; acá solo lo filtramos de la vista).
 */

import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';
import { getWorkspaceRoot } from './workspace';
import type { ActiveSession } from './types';

export type { ActiveSession };

/** `true` si el proceso `pid` sigue vivo (signal 0 = probe, no mata). */
function pidAlive(pid: number): boolean {
  if (!Number.isInteger(pid) || pid <= 0) return false;
  try {
    process.kill(pid, 0);
    return true;
  } catch (err) {
    // EPERM = vivo pero de otro usuario (lo tratamos como vivo). ESRCH = muerto.
    return (err as NodeJS.ErrnoException).code === 'EPERM';
  }
}

const SENTINEL = new Set(['—', '-', '', undefined]);
function field(v: string | undefined): string | null {
  return v && !SENTINEL.has(v) ? v : null;
}

/** Reconstruye el bucket original desde el filename saneado (`__` → `:`). */
function bucketFromFilename(file: string): string {
  return path.basename(file, '.lock').replace(/__/g, ':');
}

/**
 * Lee todos los build-claims vivos del worktree actual.
 * Tolerante a errores: si `.session-locks/` no existe, devuelve `[]`.
 */
export async function readActiveSessions(): Promise<ActiveSession[]> {
  let root: string;
  try {
    root = getWorkspaceRoot();
  } catch {
    return [];
  }
  const lockDir = path.join(root, '.session-locks');

  let entries: string[];
  try {
    entries = (await readdir(lockDir)).filter((f) => f.endsWith('.lock'));
  } catch {
    return []; // dir no existe → sin sesiones activas
  }

  const sessions: ActiveSession[] = [];
  for (const file of entries) {
    let firstLine: string;
    try {
      const raw = await readFile(path.join(lockDir, file), 'utf-8');
      firstLine = raw.split('\n')[0]?.trim() ?? '';
    } catch {
      continue;
    }
    if (!firstLine) continue;

    const parts = firstLine.split(/\s+/);
    const pid = Number.parseInt(parts[0] ?? '', 10);
    if (!pidAlive(pid)) continue; // lock muerto → fuera de la vista

    sessions.push({
      bucket: field(parts[5]) ?? bucketFromFilename(file),
      pid,
      skill: field(parts[1]) ?? 'unknown',
      startedAt: field(parts[2]),
      storyId: field(parts[3]),
      lane: field(parts[4]),
    });
  }
  return sessions;
}

/**
 * Map story_id → sesión que la está construyendo (build-claim vivo).
 * Si dos locks vivos reclaman la misma story (raro), gana el más reciente.
 */
export async function sessionsByStory(): Promise<Record<string, ActiveSession>> {
  const sessions = await readActiveSessions();
  const map: Record<string, ActiveSession> = {};
  for (const s of sessions) {
    if (!s.storyId) continue;
    const prev = map[s.storyId];
    if (!prev || (s.startedAt ?? '') >= (prev.startedAt ?? '')) {
      map[s.storyId] = s;
    }
  }
  return map;
}
