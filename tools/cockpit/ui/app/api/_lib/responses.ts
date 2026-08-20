/**
 * Helpers compartidos para route handlers.
 *
 * - errorResponse: NextResponse JSON con error + status code
 * - safeJson: parse body como JSON con error handling defensivo
 * - validatePath: anti path-traversal · whitelist de subdirs
 */

import { NextResponse } from 'next/server';
import path from 'node:path';
import { getSistemas, getWorkspaceRoot } from '@/lib/workspace';

export interface ApiError {
  error: string;
  detail?: string;
  [key: string]: unknown;
}

export function errorResponse(
  message: string,
  status: number,
  extra: Record<string, unknown> = {}
): NextResponse<ApiError> {
  return NextResponse.json<ApiError>({ error: message, ...extra }, { status });
}

export async function safeJson<T = unknown>(req: Request): Promise<T | null> {
  try {
    return (await req.json()) as T;
  } catch {
    return null;
  }
}

/**
 * Whitelist de prefixes (relativos al workspace root) accesibles vía /api/file.
 * Excluye explícitamente código de aplicación (BE/FE source).
 *
 * Permitido:
 *   - {sistema}/docs/** (sistemas derivadas del seam + discovery · sin lista hardcodeada, F-4)
 *   - .claude/**
 *   - docs/**
 */
function isWhitelistedRelPath(relPath: string): boolean {
  // Path normalizado · sin leading ./
  const clean = relPath.replace(/^\.\//, '');

  // Reject path traversal explícito
  if (clean.includes('..')) return false;

  const segments = clean.split('/');
  if (segments.length === 0) return false;

  const first = segments[0];

  // Sistema-scoped: solo dentro de docs/
  if (getSistemas().includes(first)) {
    return segments[1] === 'docs';
  }

  // Cross-cutting permitido
  if (first === '.claude') return true;
  if (first === 'docs') return true;

  return false;
}

/**
 * Resuelve un relPath (recibido del cliente) a absPath, validando whitelist
 * y anti-traversal. Devuelve `null` si el path no pasa validación.
 */
export function resolveSafePath(relPath: string): string | null {
  if (!relPath || typeof relPath !== 'string') return null;
  if (path.isAbsolute(relPath)) return null;

  // Normalizar y verificar que no escapa el root
  const normalized = path.normalize(relPath);
  if (normalized.startsWith('..') || normalized.includes('/../')) return null;

  if (!isWhitelistedRelPath(normalized)) return null;

  const root = getWorkspaceRoot();
  const absPath = path.resolve(root, normalized);

  // Final sanity: absPath debe estar dentro del root
  const rel = path.relative(root, absPath);
  if (rel.startsWith('..') || path.isAbsolute(rel)) return null;

  return absPath;
}

/**
 * Variante para paths que solo deben caer dentro del workspace root (sin
 * whitelist de subdirs específicos). Usado por `/api/open` que abre cualquier
 * archivo del repo para editar pero requiere protección anti-traversal.
 */
export function resolveWorkspacePath(relOrAbs: string): string | null {
  if (!relOrAbs || typeof relOrAbs !== 'string') return null;

  const root = getWorkspaceRoot();
  let abs: string;

  if (path.isAbsolute(relOrAbs)) {
    abs = path.normalize(relOrAbs);
  } else {
    abs = path.resolve(root, path.normalize(relOrAbs));
  }

  const rel = path.relative(root, abs);
  if (rel.startsWith('..') || path.isAbsolute(rel)) return null;

  return abs;
}
