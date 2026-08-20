/**
 * POST /api/open
 *  body: { path: string }   // relativo a WORKSPACE_ROOT o absoluto dentro del root
 *  → { ok: true, editor: "<bin-used>" }
 *
 * Spawnea editor para abrir el archivo en el host donde corre el cockpit.
 * Patrón "browser-click → spawn host editor". El cockpit corre local (per-worktree)
 * así que esto abre el editor en la misma máquina del usuario.
 *
 * Resolución del binario:
 *   1. Lee EDITOR_BIN env (puede ser "xed" o "code,xed,xdg-open" como fallback chain)
 *   2. Default chain Linux: ["xdg-open", "code", "xed", "gnome-text-editor", "nano"]
 *   3. Prueba cada bin en orden hasta uno que se ejecute sin error inmediato
 *   4. Si todos fallan → 500 con mensaje accionable (cómo overridear via .env.local)
 *
 * NUNCA usa shell exec (anti shell-injection).
 */

import { NextRequest, NextResponse } from 'next/server';
import { spawn } from 'node:child_process';
import { stat } from 'node:fs/promises';
import { z } from 'zod';
import { errorResponse, resolveWorkspacePath, safeJson } from '../_lib/responses';

const BodySchema = z.object({
  path: z.string().min(1),
});

const DEFAULT_FALLBACK_CHAIN = ['xdg-open', 'code', 'xed', 'gnome-text-editor', 'nano'];

/** Resuelve cadena de editores a probar (env o fallback). */
function resolveEditorChain(): string[] {
  const raw = process.env.EDITOR_BIN;
  if (!raw || raw.trim().length === 0) return DEFAULT_FALLBACK_CHAIN;
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}

interface SpawnAttempt {
  bin: string;
  ok: boolean;
  error?: string;
}

/**
 * Intenta spawn de UN editor. Resuelve cuando:
 * - 'spawn' event dispara (proceso lanzado OK) → ok:true
 * - 'error' event dispara (ENOENT, EACCES, etc.) → ok:false con mensaje
 * - 100ms timeout sin eventos → ok:true (asumimos lanzado; algunos wrappers
 *   no emiten 'spawn' antes de detached)
 *
 * detached + unref garantiza que el editor sobrevive al request.
 */
function trySpawn(bin: string, absPath: string): Promise<SpawnAttempt> {
  return new Promise((resolve) => {
    let settled = false;
    const finish = (result: SpawnAttempt) => {
      if (settled) return;
      settled = true;
      resolve(result);
    };

    try {
      const child = spawn(bin, [absPath], {
        detached: true,
        stdio: 'ignore',
      });

      child.on('error', (err: NodeJS.ErrnoException) => {
        finish({
          bin,
          ok: false,
          error: err.code === 'ENOENT' ? 'no instalado en PATH' : err.message,
        });
      });

      child.on('spawn', () => {
        try {
          child.unref();
        } catch {
          /* noop */
        }
        finish({ bin, ok: true });
      });

      setTimeout(() => {
        try {
          child.unref();
        } catch {
          /* noop */
        }
        finish({ bin, ok: true });
      }, 100);
    } catch (err) {
      finish({
        bin,
        ok: false,
        error: (err as Error).message,
      });
    }
  });
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = BodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }

  const abs = resolveWorkspacePath(parsed.data.path);
  if (!abs) {
    return errorResponse('path fuera del workspace o inválido', 403, {
      path: parsed.data.path,
    });
  }

  try {
    await stat(abs);
  } catch {
    return errorResponse('path no existe en el filesystem', 404, {
      path: parsed.data.path,
    });
  }

  const chain = resolveEditorChain();
  const attempts: SpawnAttempt[] = [];

  for (const bin of chain) {
    const result = await trySpawn(bin, abs);
    attempts.push(result);
    if (result.ok) {
      return NextResponse.json({
        ok: true,
        editor: bin,
        attempts_count: attempts.length,
      });
    }
  }

  return errorResponse(
    'no se pudo abrir el archivo en ningún editor de la cadena',
    500,
    {
      tried: attempts,
      hint:
        'Setea EDITOR_BIN en el .env.local del cockpit con tu editor preferido. ' +
        'Ejemplos: EDITOR_BIN=code · EDITOR_BIN=cursor · EDITOR_BIN=xdg-open · ' +
        'EDITOR_BIN=code,xed,xdg-open (fallback chain).',
      path: parsed.data.path,
    }
  );
}
