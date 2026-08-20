/**
 * Git wrappers via simple-git.
 * El cockpit usa esto para mostrar metadata útil en las cards de stories
 * (last commit, files modified, branch actual) — NUNCA para mutar repo.
 */

import simpleGit, { type SimpleGit } from 'simple-git';
import { getWorkspaceRoot } from './workspace';

let cachedGit: SimpleGit | null = null;

function git(): SimpleGit {
  if (!cachedGit) {
    cachedGit = simpleGit(getWorkspaceRoot());
  }
  return cachedGit;
}

/** Reset cache · útil para tests */
export function _resetGitCache(): void {
  cachedGit = null;
}

export async function currentBranch(): Promise<string> {
  const branch = await git().revparse(['--abbrev-ref', 'HEAD']);
  return branch.trim();
}

export async function lastCommitSha(short = true): Promise<string> {
  const sha = await git().revparse([short ? '--short' : 'HEAD', 'HEAD']);
  return sha.trim();
}

export interface FileChange {
  path: string;
  /** A=added, M=modified, D=deleted, R=renamed, ?=untracked */
  status: string;
}

/** Lista archivos modificados (staged + unstaged + untracked) del workdir */
export async function repoStatus(): Promise<FileChange[]> {
  const status = await git().status();
  const changes: FileChange[] = [];
  for (const f of status.modified) changes.push({ path: f, status: 'M' });
  for (const f of status.created) changes.push({ path: f, status: 'A' });
  for (const f of status.deleted) changes.push({ path: f, status: 'D' });
  for (const f of status.renamed) changes.push({ path: f.to, status: 'R' });
  for (const f of status.not_added) changes.push({ path: f, status: '?' });
  return changes;
}

/** Diff HEAD vs main para un path · útil para mostrar "qué cambió esta story" */
export async function fileChangesAgainstMain(filePath: string): Promise<string> {
  try {
    return await git().diff(['main...HEAD', '--', filePath]);
  } catch {
    return '';
  }
}

/** Lista commits que tocaron un path concreto (limit N) */
export async function fileCommitHistory(
  filePath: string,
  limit = 10
): Promise<Array<{ hash: string; date: string; message: string }>> {
  try {
    const log = await git().log({
      file: filePath,
      maxCount: limit,
    });
    return log.all.map((c) => ({
      hash: c.hash.substring(0, 8),
      date: c.date,
      message: c.message,
    }));
  } catch {
    return [];
  }
}
