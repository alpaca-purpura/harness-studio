/**
 * GET /api/sessions  → { sessions: ActiveSession[], by_story: Record<string, ActiveSession> }
 *
 * Lee los build-claims vivos de `.session-locks/` del worktree actual (hub único,
 * ADR-009). El board pinta un badge "🔨 lane" sobre las stories en construcción.
 *
 * Filesystem-as-DB + runtime: no toca git. Si no hay locks → listas vacías.
 */

import { NextResponse } from 'next/server';
import { errorResponse } from '../_lib/responses';
import { readActiveSessions, sessionsByStory } from '@/lib/sessions';

// Estado runtime que cambia entre requests → nunca cachear.
export const dynamic = 'force-dynamic';
export const revalidate = 0;

export async function GET(): Promise<NextResponse> {
  try {
    const [sessions, byStory] = await Promise.all([
      readActiveSessions(),
      sessionsByStory(),
    ]);
    return NextResponse.json({ sessions, by_story: byStory });
  } catch (err) {
    return errorResponse('error leyendo sesiones activas', 500, {
      detail: (err as Error).message,
    });
  }
}
