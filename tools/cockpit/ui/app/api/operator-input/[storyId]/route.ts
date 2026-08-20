/**
 * GET   /api/operator-input/{storyId}?sistema={sistema}  → { operatorInput }
 * PATCH /api/operator-input/{storyId}?sistema={sistema}
 *        body: { section: 'notes'|'refs'|'conversation', entry: <Entry> }
 *                                                     → { operatorInput }
 *
 * Permisos (cockpit-permissions.md):
 *   - notes / refs  → CRUD completo (acá implementamos APPEND solamente · cockpit
 *                     UI puede usar GET+PUT raw vía /api/file para edits inline)
 *   - conversation  → solo APPEND · author DEBE ser el operador (Claude appendea
 *                     via skill usando appendConversationEntry en lib, no este endpoint)
 *
 * Filename en disco: `operator-input.md` o el legacy `chris-input.md` del
 * template del kit. // legacy F-1
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { readFile, readdir } from 'node:fs/promises';
import { z } from 'zod';
import { errorResponse, safeJson } from '../../_lib/responses';
import {
  parseOperatorInput,
  serializeOperatorInput,
  appendConversationEntry,
  OPERATOR_INPUT_FILENAMES,
} from '@/lib/operator-input-parser';
import { writeFileAtomic } from '@/lib/fs-writer';
import { storiesPath, archiveRootPath, getSistemas, getSelectableSistemas } from '@/lib/workspace';
import type { OperatorInput, ConvEntry, Note, Ref, RefType } from '@/lib/types';

const NoteEntrySchema = z.object({
  timestamp: z.string().optional(),
  text: z.string().min(1),
});

const RefEntrySchema = z.object({
  type: z.enum(['link', 'img', 'text', 'story-ref', 'learning-ref', 'doc']),
  value: z.string().min(1),
  comment: z.string().optional(),
});

const ConvEntrySchema = z.object({
  timestamp: z.string().optional(),
  // 'chris' aceptado como alias de clientes viejos. // legacy F-1
  author: z.enum(['operador', 'chris']),
  text: z.string().min(1),
});

const PatchBodySchema = z.discriminatedUnion('section', [
  z.object({ section: z.literal('notes'), entry: NoteEntrySchema }),
  z.object({ section: z.literal('refs'), entry: RefEntrySchema }),
  z.object({ section: z.literal('conversation'), entry: ConvEntrySchema }),
]);

function nowTimestamp(): string {
  // "YYYY-MM-DD HH:MM" en tz local
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}`;
}

async function findOperatorInputPath(sistema: string, storyId: string): Promise<string | null> {
  const dirs: string[] = [path.join(storiesPath(sistema), storyId)];

  const archiveRoot = archiveRootPath(sistema); // = archive/ (NO path.dirname(archivePath) → archive/{year} bug)
  try {
    const years = await readdir(archiveRoot, { withFileTypes: true });
    for (const y of years) {
      if (y.isDirectory()) {
        dirs.push(path.join(archiveRoot, y.name, 'stories', storyId));
      }
    }
  } catch {
    // archive dir no existe
  }

  for (const dir of dirs) {
    // operator-input.md primero; chris-input.md como fallback. // legacy F-1
    for (const filename of OPERATOR_INPUT_FILENAMES) {
      const candidate = path.join(dir, filename);
      try {
        await readFile(candidate, 'utf-8');
        return candidate;
      } catch {
        // continuar
      }
    }
  }
  return null;
}

async function loadOperatorInput(absPath: string): Promise<OperatorInput> {
  const raw = await readFile(absPath, 'utf-8');
  return parseOperatorInput(raw);
}

export async function GET(
  req: NextRequest,
  context: { params: Promise<{ storyId: string }> }
): Promise<NextResponse> {
  const { storyId } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSelectableSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const ciPath = await findOperatorInputPath(sistema, storyId);
  if (!ciPath) {
    return errorResponse('operator-input no encontrado', 404, { story_id: storyId, sistema });
  }

  try {
    const operatorInput = await loadOperatorInput(ciPath);
    return NextResponse.json({ operatorInput });
  } catch (err) {
    return errorResponse('error parseando operator-input', 500, {
      detail: (err as Error).message,
    });
  }
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<{ storyId: string }> }
): Promise<NextResponse> {
  const { storyId } = await context.params;
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const body = await safeJson(req);
  if (!body) return errorResponse('body JSON inválido', 400);

  const parsed = PatchBodySchema.safeParse(body);
  if (!parsed.success) {
    return errorResponse('body inválido', 400, { issues: parsed.error.issues });
  }

  const ciPath = await findOperatorInputPath(sistema, storyId);
  if (!ciPath) {
    return errorResponse('operator-input no encontrado', 404, { story_id: storyId, sistema });
  }

  try {
    if (parsed.data.section === 'conversation') {
      // Append vía helper dedicado del lib · author se normaliza a 'operador'
      const entry: ConvEntry = {
        timestamp: parsed.data.entry.timestamp || nowTimestamp(),
        author: 'operador',
        text: parsed.data.entry.text,
      };
      await appendConversationEntry(ciPath, entry);
      const operatorInput = await loadOperatorInput(ciPath);
      return NextResponse.json({ operatorInput });
    }

    // Notes o Refs · parse + append + serialize
    const data = await loadOperatorInput(ciPath);
    if (parsed.data.section === 'notes') {
      const note: Note = {
        timestamp: parsed.data.entry.timestamp || nowTimestamp(),
        text: parsed.data.entry.text,
      };
      data.notes.push(note);
    } else {
      // refs
      const ref: Ref = {
        type: parsed.data.entry.type as RefType,
        value: parsed.data.entry.value,
        comment: parsed.data.entry.comment,
      };
      data.refs.push(ref);
    }
    data.frontmatter.last_modified = new Date().toISOString();
    data.frontmatter.notes_count = data.notes.length;
    data.frontmatter.refs_count = data.refs.length;
    data.frontmatter.conversation_count = data.conversation.length;

    const serialized = serializeOperatorInput(data);
    await writeFileAtomic(ciPath, serialized);
    return NextResponse.json({ operatorInput: data });
  } catch (err) {
    return errorResponse('error actualizando operator-input', 500, {
      detail: (err as Error).message,
    });
  }
}
