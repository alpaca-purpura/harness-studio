/**
 * GET /api/file?path=<rel>     → { path, content }
 * PUT /api/file body:{path, content}  → { ok: true }
 *
 * Whitelist: solo paths dentro de workspace root + dentro de subdirs permitidos
 * ({sistema}/docs/, .claude/, docs/).
 */

import { NextRequest, NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import { z } from 'zod';
import { errorResponse, resolveSafePath, safeJson } from '../_lib/responses';
import { writeFileAtomic } from '@/lib/fs-writer';

const PutBodySchema = z.object({
  path: z.string().min(1),
  content: z.string(),
});

export async function GET(req: NextRequest): Promise<NextResponse> {
  const relPath = req.nextUrl.searchParams.get('path');
  if (!relPath) {
    return errorResponse('query param "path" requerido', 400);
  }
  const abs = resolveSafePath(relPath);
  if (!abs) {
    return errorResponse('path fuera del whitelist o inválido', 403, { path: relPath });
  }
  try {
    const content = await readFile(abs, 'utf-8');
    return NextResponse.json({ path: relPath, content });
  } catch (err) {
    const code = (err as NodeJS.ErrnoException).code;
    if (code === 'ENOENT') {
      return errorResponse('archivo no existe', 404, { path: relPath });
    }
    return errorResponse('error leyendo archivo', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = PutBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }
  const { path: relPath, content } = parsed.data;

  const abs = resolveSafePath(relPath);
  if (!abs) {
    return errorResponse('path fuera del whitelist o inválido', 403, { path: relPath });
  }

  try {
    await writeFileAtomic(abs, content);
    return NextResponse.json({ ok: true });
  } catch (err) {
    return errorResponse('error escribiendo archivo', 500, {
      detail: (err as Error).message,
    });
  }
}
