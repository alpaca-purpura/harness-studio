/**
 * POST /api/merge-release
 *  body: { sistema, releaseId, confirmFinal: boolean, verified?: boolean, verificationNote?: string }
 *
 * Gate de shipped (release-protocol.md § 5): un release pasa a `shipped` SOLO si
 *   (a) todas sus stories están en {done, dropped}  Y
 *   (b) el operador confirma la PRUEBA DE COMPORTAMIENTO verde (verified=true) — corrió
 *       la suite de integración + E2E smoke y todo pasa sin romper lo anterior.
 *
 * Dual-confirm:
 *   1. confirmFinal=false → preview { plan: Operation[], executed: false }
 *   2. confirmFinal=true + verified=true → flip status=shipped + shipped_date +
 *      verified_by/verified_at/verification_note en el release.yaml.
 *
 * Lo que SÍ ejecuta: actualizar el release.yaml (status=shipped + sellos de
 * verificación). Lo que NO ejecuta automático: git mv de las stories al archive
 * (cada story ya se squash-mergeó en su Fase F individual) — eso sigue manual o
 * vía /pm-{sistema}. El "pase a producción" es OTRO eje (futuro · ver § 8).
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { z } from 'zod';
import { errorResponse, safeJson } from '../_lib/responses';
import { readRelease, loadStoryStatesForRelease, writeRelease } from '@/lib/release-resolver';
import { storiesPath, archivePath, getSistemas } from '@/lib/workspace';
import type { StoryState } from '@/lib/types';

const BodySchema = z.object({
  sistema: z.string(),
  releaseId: z.string().min(1),
  confirmFinal: z.boolean().optional().default(false),
  verified: z.boolean().optional().default(false),
  verificationNote: z.string().optional(),
});

interface PlanOperation {
  op: 'archive_story' | 'update_release' | 'generate_release_notes';
  description: string;
  source?: string;
  target?: string;
}

const TERMINAL_STATES: ReadonlySet<StoryState> = new Set(['done', 'dropped']);

export async function POST(req: NextRequest): Promise<NextResponse> {
  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = BodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }
  const { sistema, releaseId, confirmFinal, verified, verificationNote } = parsed.data;

  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  let release;
  try {
    release = await readRelease(sistema, releaseId);
  } catch {
    return errorResponse('release no encontrado', 404, { release_id: releaseId, sistema });
  }

  if (release.status === 'shipped') {
    return errorResponse('release ya está shipped', 409, { release_id: releaseId });
  }

  // Verificar todas las stories en done o dropped
  const storyStates = await loadStoryStatesForRelease(release);
  const nonTerminal = release.stories.filter((_, i) => !TERMINAL_STATES.has(storyStates[i]));
  if (nonTerminal.length > 0) {
    return errorResponse(
      'release no está listo para merge · hay stories no terminales',
      409,
      {
        release_id: releaseId,
        non_terminal_stories: nonTerminal,
        states: storyStates,
      }
    );
  }

  // Construir plan
  const year = new Date().getFullYear();
  const archiveBase = archivePath(sistema, year);
  const liveBase = storiesPath(sistema);
  const releaseNotesPath = path.join(
    path.dirname(release.path ?? ''),
    '..',
    'release-notes',
    `${releaseId}.md`
  );

  const plan: PlanOperation[] = [];
  release.stories.forEach((storyId, i) => {
    if (storyStates[i] === 'done') {
      plan.push({
        op: 'archive_story',
        description: `git mv ${storyId} → archive/${year}/stories/${storyId}`,
        source: path.join(liveBase, storyId),
        target: path.join(archiveBase, storyId),
      });
    }
  });
  plan.push({
    op: 'update_release',
    description: `update ${releaseId}.yaml status=shipped + shipped_date=${new Date()
      .toISOString()
      .substring(0, 10)}`,
    target: release.path ?? '',
  });
  plan.push({
    op: 'generate_release_notes',
    description: `generar release-notes/${releaseId}.md con summary de stories shipped`,
    target: releaseNotesPath,
  });

  if (!confirmFinal) {
    return NextResponse.json({ plan, executed: false, preview: true });
  }

  // Gate de comportamiento: shipped requiere confirmación explícita del operador de que
  // corrió la prueba de integración + E2E smoke y dio verde (release-protocol.md § 5).
  if (!verified) {
    return errorResponse(
      'falta confirmar la prueba de comportamiento · marca el check de verificación antes de cerrar el release',
      409,
      { release_id: releaseId, reason: 'verified=false' }
    );
  }

  // Flip a shipped + sellos de verificación (escribe release.yaml).
  const nowIso = new Date().toISOString();
  release.status = 'shipped';
  release.shipped_date = nowIso;
  release.verified_by = 'operador';
  release.verified_at = nowIso;
  release.verification_note = verificationNote?.trim() || null;
  // Reserva el eje de despliegue (futuro): el release nace "no desplegado a prod".
  if (!release.production_status) release.production_status = 'not_deployed';

  try {
    await writeRelease(release);
  } catch (err) {
    return errorResponse('error escribiendo release', 500, {
      detail: (err as Error).message,
    });
  }

  return NextResponse.json({
    plan,
    executed: true,
    release,
    note: 'Release marcado shipped (base sólida verificada). El git mv de las stories al archive (si falta) lo haces manual o vía /pm-{sistema}. El "pase a producción" es otro paso (futuro).',
  });
}
