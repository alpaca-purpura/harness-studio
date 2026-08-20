/**
 * GET    /api/releases?sistema={sistema}       → { releases: Release[] }
 * POST   /api/releases body:{release_id, sistema, name, ...}  → { release }
 * PUT    /api/releases?id=F2&sistema={sistema} body:{name?, description?, ...}  → { release }
 * DELETE /api/releases?id=F2&sistema={sistema} → { ok: true, archived_path }
 *
 * Per cockpit-permissions.md:
 *   PUT solo edita: name, description, target_date, order, stories
 *   NO edita: status, shipped_date, release_id, sistema, created_at, created_by
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { rename, mkdir } from 'node:fs/promises';
import matter from 'gray-matter';
import { z } from 'zod';
import { errorResponse, safeJson } from '../_lib/responses';
import { listReleases, readRelease, writeRelease } from '@/lib/release-resolver';
import { writeFileAtomic } from '@/lib/fs-writer';
import { releasesPath, getSistemas, getSelectableSistemas } from '@/lib/workspace';
import type { Release, ReleaseStatus } from '@/lib/types';

const CreateBodySchema = z.object({
  release_id: z.string().regex(/^[A-Za-z0-9_-]+$/, 'solo alfanuméricos, _ y -'),
  sistema: z.string(),
  name: z.string().min(1),
  description: z.string(),
  target_date: z.string().nullable().optional(),
  order: z.number().int().optional(),
  stories: z.array(z.string()).optional(),
});

const UpdateBodySchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
  target_date: z.string().nullable().optional(),
  order: z.number().int().optional(),
  stories: z.array(z.string()).optional(),
});

const READ_ONLY_FIELDS = [
  'status',
  'shipped_date',
  'release_id',
  'sistema',
  'created_at',
  'created_by',
];

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  // platform aceptado en GET (read-only): no tiene releases → lista vacía graceful.
  if (!getSelectableSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  try {
    const releases = await listReleases(sistema);
    return NextResponse.json({ releases });
  } catch (err) {
    return errorResponse('error listando releases', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = CreateBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }
  const data = parsed.data;

  if (!getSistemas().includes(data.sistema)) {
    return errorResponse(`sistema desconocido: ${data.sistema}`, 400);
  }

  // Verificar que no exista
  try {
    await readRelease(data.sistema, data.release_id);
    return errorResponse('release ya existe', 409, {
      release_id: data.release_id,
      sistema: data.sistema,
    });
  } catch {
    // expected · no existe
  }

  const nowIso = new Date().toISOString();
  const release: Release = {
    release_id: data.release_id,
    sistema: data.sistema,
    name: data.name,
    description: data.description,
    status: 'planning' as ReleaseStatus,
    target_date: data.target_date ?? null,
    shipped_date: null,
    order: data.order ?? 0,
    created_at: nowIso,
    created_by: 'operador',
    stories: data.stories ?? [],
    verified_by: null,
    verified_at: null,
    verification_note: null,
    production_status: 'not_deployed',
    production_version: null,
    production_scheduled_at: null,
    deployed_at: null,
    release_branch: null,
    maps_legacy_outcome: null,
    maps_legacy_phase: null,
    body: `# ${data.release_id} · ${data.name}\n\n> ${data.description}\n\n## Stories incluidas\n\n## Notas del release\n\n(libre · el operador escribe aquí contexto adicional · decisiones · gotchas)\n`,
    path: path.join(releasesPath(data.sistema), `${data.release_id}.yaml`),
  };

  try {
    await writeRelease(release);
    return NextResponse.json({ release });
  } catch (err) {
    return errorResponse('error escribiendo release', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  const releaseId = req.nextUrl.searchParams.get('id');
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!releaseId) return errorResponse('query param "id" requerido', 400);
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  // Check forbidden fields
  if (typeof body === 'object' && body !== null) {
    const sent = Object.keys(body as Record<string, unknown>);
    const forbidden = sent.filter((k) => READ_ONLY_FIELDS.includes(k));
    if (forbidden.length > 0) {
      return errorResponse('algunos fields no son editables', 403, {
        forbidden_fields: forbidden,
        reason:
          'status y shipped_date se auto-calculan; release_id/sistema/created_at/created_by son inmutables.',
      });
    }
  }

  const parsed = UpdateBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }

  let release: Release;
  try {
    release = await readRelease(sistema, releaseId);
  } catch {
    return errorResponse('release no encontrado', 404, { release_id: releaseId, sistema });
  }

  // Releases shipped son inmutables (entregados · base sólida). No se editan
  // nombre/descripción/etc. Correcciones excepcionales van por /pm-{sistema}.
  if (release.status === 'shipped') {
    return errorResponse('no se puede editar un release shipped', 403, {
      release_id: releaseId,
      reason:
        'releases ya entregados son inmutables. Usa /pm-{sistema} para correcciones excepcionales.',
    });
  }

  const updated: Release = {
    ...release,
    ...parsed.data,
  };

  try {
    await writeRelease(updated);
    return NextResponse.json({ release: updated });
  } catch (err) {
    return errorResponse('error escribiendo release', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function DELETE(req: NextRequest): Promise<NextResponse> {
  const releaseId = req.nextUrl.searchParams.get('id');
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!releaseId) return errorResponse('query param "id" requerido', 400);
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  let release: Release;
  try {
    release = await readRelease(sistema, releaseId);
  } catch {
    return errorResponse('release no encontrado', 404, { release_id: releaseId, sistema });
  }

  if (release.status === 'shipped') {
    return errorResponse('no se puede archivar release shipped', 403, {
      release_id: releaseId,
      reason: 'releases ya entregados son inmutables. Usa /pm-{sistema} para correcciones excepcionales.',
    });
  }

  const archiveDir = path.join(releasesPath(sistema), '_archived');
  const archivedPath = path.join(archiveDir, `${releaseId}.yaml`);

  try {
    await mkdir(archiveDir, { recursive: true });
    if (release.path) {
      await rename(release.path, archivedPath);
    } else {
      // fallback: serializar y escribir
      const { body, path: _p, ...frontmatter } = release;
      const serialized = matter.stringify(body ?? '', frontmatter as Record<string, unknown>);
      await writeFileAtomic(archivedPath, serialized);
    }
    return NextResponse.json({ ok: true, archived_path: archivedPath });
  } catch (err) {
    return errorResponse('error archivando release', 500, {
      detail: (err as Error).message,
    });
  }
}
