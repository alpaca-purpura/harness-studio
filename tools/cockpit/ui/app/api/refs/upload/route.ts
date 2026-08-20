/**
 * POST /api/refs/upload
 *  FormData: { file: Blob, sistema: string, storyId: string, comment?: string }
 *  → { relativePath, displayName }
 *
 * Validaciones:
 *   - Extensión whitelist: png|jpg|jpeg|gif|webp|svg|pdf|md|txt
 *   - Sanitiza filename: [^a-zA-Z0-9._-] → _ · max 80 chars
 *   - Anti path-traversal: NUNCA permite `..` o `/` en filename
 *   - Collision: counter suffix (foo.png → foo-2.png)
 *   - Max size 10MB
 *   - Escribe blob a {sistema}/docs/product/stories/{storyId}/refs/{filename}
 *   - Appendea entry al operator-input sección Referencias
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { mkdir, stat, writeFile, readdir } from 'node:fs/promises';
import { errorResponse } from '../../_lib/responses';
import { storiesPath, archiveRootPath, getSistemas } from '@/lib/workspace';
import {
  parseOperatorInput,
  serializeOperatorInput,
  OPERATOR_INPUT_FILENAMES,
} from '@/lib/operator-input-parser';
import { writeFileAtomic } from '@/lib/fs-writer';
import { readFile } from 'node:fs/promises';
import type { RefType } from '@/lib/types';

const ALLOWED_EXTENSIONS = new Set([
  'png',
  'jpg',
  'jpeg',
  'gif',
  'webp',
  'svg',
  'pdf',
  'md',
  'txt',
]);

const MAX_SIZE_BYTES = 10 * 1024 * 1024; // 10MB

function sanitizeFilename(name: string): string {
  // Strip path components · solo basename
  const base = path.basename(name);
  // Reemplazar caracteres no seguros
  let safe = base.replace(/[^a-zA-Z0-9._-]/g, '_');
  // Reject paths traversal residuales
  safe = safe.replace(/^\.+/, ''); // strip leading dots
  if (safe.length === 0) safe = 'unnamed';
  // Max 80 chars
  if (safe.length > 80) {
    const ext = path.extname(safe);
    const stem = path.basename(safe, ext).substring(0, 80 - ext.length);
    safe = `${stem}${ext}`;
  }
  return safe;
}

function getExtension(filename: string): string {
  const ext = path.extname(filename).toLowerCase().replace(/^\./, '');
  return ext;
}

function refTypeForExtension(ext: string): RefType {
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext)) return 'img';
  if (ext === 'pdf') return 'doc';
  if (ext === 'md' || ext === 'txt') return 'doc';
  return 'doc';
}

async function findStoryDir(sistema: string, storyId: string): Promise<string | null> {
  const liveDir = path.join(storiesPath(sistema), storyId);
  try {
    await stat(liveDir);
    return liveDir;
  } catch {
    // try archive
  }
  const archiveRoot = archiveRootPath(sistema); // = archive/ (NO path.dirname(archivePath) → archive/{year} bug)
  try {
    const years = await readdir(archiveRoot, { withFileTypes: true });
    for (const y of years) {
      if (!y.isDirectory()) continue;
      const candidate = path.join(archiveRoot, y.name, 'stories', storyId);
      try {
        await stat(candidate);
        return candidate;
      } catch {
        // continuar
      }
    }
  } catch {
    // archive dir no existe
  }
  return null;
}

async function resolveCollision(dir: string, filename: string): Promise<string> {
  let candidate = filename;
  let abs = path.join(dir, candidate);
  let counter = 2;
  const ext = path.extname(filename);
  const stem = path.basename(filename, ext);

  // Loop hasta encontrar nombre libre · cap 1000 para evitar loop infinito
  for (let i = 0; i < 1000; i++) {
    try {
      await stat(abs);
      // existe · siguiente intento
      candidate = `${stem}-${counter}${ext}`;
      abs = path.join(dir, candidate);
      counter += 1;
    } catch {
      // no existe · OK
      return candidate;
    }
  }
  throw new Error('demasiadas colisiones de nombre');
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  let formData: FormData;
  try {
    formData = await req.formData();
  } catch {
    return errorResponse('body multipart inválido', 400);
  }

  const file = formData.get('file');
  const sistema = formData.get('sistema');
  const storyId = formData.get('storyId');
  const commentRaw = formData.get('comment');

  if (!(file instanceof File)) {
    return errorResponse('campo "file" requerido (Blob)', 400);
  }
  if (typeof sistema !== 'string' || !sistema) {
    return errorResponse('campo "sistema" requerido', 400);
  }
  if (typeof storyId !== 'string' || !storyId) {
    return errorResponse('campo "storyId" requerido', 400);
  }
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  const comment = typeof commentRaw === 'string' ? commentRaw : undefined;

  // Size check
  if (file.size > MAX_SIZE_BYTES) {
    return errorResponse(
      `archivo excede ${MAX_SIZE_BYTES / 1024 / 1024}MB`,
      413,
      { size: file.size, max: MAX_SIZE_BYTES }
    );
  }

  // Sanitize + validate extension
  const rawName = file.name || 'unnamed';
  const safeName = sanitizeFilename(rawName);
  const ext = getExtension(safeName);
  if (!ALLOWED_EXTENSIONS.has(ext)) {
    return errorResponse('extensión no permitida', 400, {
      extension: ext,
      allowed: [...ALLOWED_EXTENSIONS],
    });
  }

  const storyDir = await findStoryDir(sistema, storyId);
  if (!storyDir) {
    return errorResponse('story no encontrada', 404, { story_id: storyId, sistema });
  }

  const refsDir = path.join(storyDir, 'refs');

  let finalName: string;
  let absPath: string;
  try {
    await mkdir(refsDir, { recursive: true });
    finalName = await resolveCollision(refsDir, safeName);
    absPath = path.join(refsDir, finalName);
  } catch (err) {
    return errorResponse('error preparando ruta destino', 500, {
      detail: (err as Error).message,
    });
  }

  // Escribir blob a disco
  try {
    const arrayBuffer = await file.arrayBuffer();
    await writeFile(absPath, Buffer.from(arrayBuffer));
  } catch (err) {
    return errorResponse('error escribiendo archivo', 500, {
      detail: (err as Error).message,
    });
  }

  // Appendea entry al operator-input (si existe · acepta filename legacy F-1)
  for (const inputFilename of OPERATOR_INPUT_FILENAMES) {
    const operatorInputPath = path.join(storyDir, inputFilename);
    try {
      const raw = await readFile(operatorInputPath, 'utf-8');
      const data = parseOperatorInput(raw);
      const refType = refTypeForExtension(ext);
      data.refs.push({
        type: refType,
        value: `refs/${finalName}`,
        comment,
      });
      data.frontmatter.last_modified = new Date().toISOString();
      data.frontmatter.refs_count = data.refs.length;
      const serialized = serializeOperatorInput(data);
      await writeFileAtomic(operatorInputPath, serialized);
      break;
    } catch {
      // operator-input no existe o falla parse · OK · upload sigue válido sin registrar entry
    }
  }

  return NextResponse.json({
    relativePath: `refs/${finalName}`,
    displayName: finalName,
    absolutePath: absPath,
  });
}
