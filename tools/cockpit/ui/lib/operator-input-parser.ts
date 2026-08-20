/**
 * operator-input parser + serializer (round-trip idempotente).
 *
 * Schema: process-doc del kit (operator-input protocol; filename legacy del
 * template: `chris-input.md` // legacy F-1).
 *
 * 3 secciones secuenciales:
 *   ## 💭 Notas
 *   ## 📎 Referencias
 *   ## 💬 Conversación
 *
 * Cada `note` y `conv entry` abre con `### YYYY-MM-DD HH:MM` (+ author/skill/verdict).
 * Refs son entries `- **emoji tipo** · valor` con opcional comentario `  > comentario`.
 *
 * Compat F-4: al parsear se acepta autor `operador` y `chris` (legacy F-1);
 * al serializar SIEMPRE se escribe `operador`.
 */

import matter from 'gray-matter';
import {
  type OperatorInput,
  type OperatorInputFrontmatter,
  type ConvEntry,
  type ConvVerdict,
  type Note,
  type Ref,
  type RefType,
  EMOJI_TO_REF_TYPE,
  LABEL_TO_VERDICT,
  REF_TYPE_TO_EMOJI,
  VERDICT_TO_LABEL,
} from './types';
import { readFile } from 'node:fs/promises';
import { appendToFile, writeFileAtomic } from './fs-writer';

// ────────────────────────────────────────────────────────────────────────────
// Filenames reconocidos (el nuevo genérico + el legacy del template del kit)
// ────────────────────────────────────────────────────────────────────────────

export const OPERATOR_INPUT_FILENAME = 'operator-input.md';
/** Filename del template del kit read-only — se sigue aceptando. // legacy F-1 */
export const LEGACY_OPERATOR_INPUT_FILENAME = 'chris-input.md';
export const OPERATOR_INPUT_FILENAMES = [
  OPERATOR_INPUT_FILENAME,
  LEGACY_OPERATOR_INPUT_FILENAME, // legacy F-1
] as const;

/** ¿Este basename es un operator-input (nuevo o legacy)? */
export function isOperatorInputFilename(basename: string): boolean {
  return (OPERATOR_INPUT_FILENAMES as readonly string[]).includes(basename.toLowerCase());
}

// ────────────────────────────────────────────────────────────────────────────
// Section headers (constants verbatim del protocolo)
// ────────────────────────────────────────────────────────────────────────────

const HEADER_NOTES = '## 💭 Notas';
const HEADER_REFS = '## 📎 Referencias';
const HEADER_CONV = '## 💬 Conversación';

// Regex helpers
const TIMESTAMP_RE = /^### (\d{4}-\d{2}-\d{2} \d{2}:\d{2})$/;
// Acepta `operador` y el alias `chris` de archivos viejos. // legacy F-1
const CONV_OPERATOR_RE = /^### (\d{4}-\d{2}-\d{2} \d{2}:\d{2}) · 🧑 (?:operador|chris)$/;
const CONV_CLAUDE_RE =
  /^### (\d{4}-\d{2}-\d{2} \d{2}:\d{2}) · 🤖 claude · `\/([^`]+)` · (\S+) (.+)$/u;
// Ref pattern · emoji puede ser uno o varios chars unicode
// Ejemplos:
//   - **🔗 link** · https://example.com
//   - **🖼 img** · refs/foo.png
const REF_RE = /^- \*\*(\S+) ([a-z-]+)\*\* · (.+)$/u;
const REF_COMMENT_RE = /^  > (.+)$/u;

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

/**
 * Extrae comentarios HTML leading (típicamente `<!-- voseo-allowed: ... -->`)
 * que aparecen ANTES del frontmatter. gray-matter requiere frontmatter al
 * inicio · sin esto el archivo no parsea.
 */
function extractLeadingHtmlComments(content: string): {
  leadingComments: string;
  rest: string;
} {
  const lines = content.split('\n');
  const collected: string[] = [];
  let idx = 0;
  while (idx < lines.length) {
    const line = lines[idx];
    if (line.startsWith('<!--')) {
      // single-line comment
      collected.push(line);
      idx += 1;
      continue;
    }
    if (line.trim() === '') {
      // permitir blank line entre comments
      idx += 1;
      continue;
    }
    break;
  }
  return {
    leadingComments: collected.join('\n'),
    rest: lines.slice(idx).join('\n'),
  };
}

// ────────────────────────────────────────────────────────────────────────────
// Parser
// ────────────────────────────────────────────────────────────────────────────

export function parseOperatorInput(content: string): OperatorInput {
  // Strip leading HTML comments (ej. <!-- voseo-allowed --> antes del frontmatter)
  // Preservamos el comment en `preamble` para round-trip idempotente.
  const { leadingComments, rest } = extractLeadingHtmlComments(content);
  const parsed = matter(rest);
  const frontmatter = parsed.data as OperatorInputFrontmatter;
  const body = parsed.content;

  // Tokenizar el body en líneas
  const lines = body.split('\n');

  // Encontrar índices de las 3 secciones
  const idxNotes = lines.findIndex((l) => l.trim() === HEADER_NOTES);
  const idxRefs = lines.findIndex((l) => l.trim() === HEADER_REFS);
  const idxConv = lines.findIndex((l) => l.trim() === HEADER_CONV);

  if (idxNotes === -1 || idxRefs === -1 || idxConv === -1) {
    throw new Error(
      'operator-input malformado: faltan una o más de las 3 secciones (💭 Notas / 📎 Referencias / 💬 Conversación)'
    );
  }

  // Preámbulo = leading HTML comments (si existen) + todo entre frontmatter y la primera sección
  const innerPreambleLines = lines.slice(0, idxNotes);
  const innerPreamble = innerPreambleLines.join('\n').trim();
  const preamble = [leadingComments, innerPreamble].filter(Boolean).join('\n\n').trim();

  const notesLines = lines.slice(idxNotes + 1, idxRefs);
  const refsLines = lines.slice(idxRefs + 1, idxConv);
  const convLines = lines.slice(idxConv + 1);

  const notes = parseNotesSection(notesLines);
  const refs = parseRefsSection(refsLines);
  const conversation = parseConvSection(convLines);

  return {
    frontmatter,
    notes,
    refs,
    conversation,
    preamble: preamble.length > 0 ? preamble : undefined,
  };
}

function parseNotesSection(lines: string[]): Note[] {
  const notes: Note[] = [];
  let current: Note | null = null;
  let accumulator: string[] = [];

  const flush = () => {
    if (current) {
      current.text = accumulator.join('\n').trim();
      if (current.text.length > 0 && !isPlaceholderNote(current.text)) {
        notes.push(current);
      }
      accumulator = [];
      current = null;
    }
  };

  for (const line of lines) {
    const m = TIMESTAMP_RE.exec(line);
    if (m) {
      flush();
      current = { timestamp: m[1], text: '' };
      continue;
    }
    if (current) {
      accumulator.push(line);
    }
    // Si todavía no encontramos timestamp y la línea no es vacía/blockquote/instrucción,
    // la ignoramos (corresponde al texto guía del template).
  }
  flush();
  return notes;
}

/** Detecta entries placeholder del template (texto "Sin notas todavía...") */
function isPlaceholderNote(text: string): boolean {
  const lower = text.toLowerCase();
  return (
    lower.includes('sin notas todavía') ||
    lower.includes('escribe aqui') ||
    lower.includes('escribe aquí')
  );
}

function parseRefsSection(lines: string[]): Ref[] {
  const refs: Ref[] = [];
  let pending: Ref | null = null;

  for (const line of lines) {
    const matchRef = REF_RE.exec(line);
    if (matchRef) {
      // Flush pendiente
      if (pending) refs.push(pending);
      const [, emoji, typeStr, value] = matchRef;
      // Resolver type del emoji (más robusto que el typeStr porque emoji es PK)
      const type = (EMOJI_TO_REF_TYPE[emoji] ?? (typeStr as RefType));
      pending = { type, value: value.trim() };
      continue;
    }
    const matchComment = REF_COMMENT_RE.exec(line);
    if (matchComment && pending) {
      pending.comment = matchComment[1].trim();
      continue;
    }
    // Resto de líneas (blank, blockquote del template) → ignorar
  }
  if (pending) refs.push(pending);
  return refs;
}

function parseConvSection(lines: string[]): ConvEntry[] {
  const entries: ConvEntry[] = [];
  let current: ConvEntry | null = null;
  let accumulator: string[] = [];

  const flush = () => {
    if (current) {
      current.text = accumulator.join('\n').trim();
      entries.push(current);
      accumulator = [];
      current = null;
    }
  };

  for (const line of lines) {
    const claudeMatch = CONV_CLAUDE_RE.exec(line);
    if (claudeMatch) {
      flush();
      const [, timestamp, skill, emoji, rest] = claudeMatch;
      // `rest` viene como "APLICADO" o "DUDA" etc. (label sin emoji)
      // Si label tiene espacios (raro), tomamos primera palabra
      const label = rest.trim().split(/\s+/)[0];
      const verdict: ConvVerdict | undefined =
        LABEL_TO_VERDICT[label] ?? matchVerdictByEmoji(emoji);
      current = {
        timestamp,
        author: 'claude',
        skill,
        verdict,
        text: '',
      };
      continue;
    }
    const operatorMatch = CONV_OPERATOR_RE.exec(line);
    if (operatorMatch) {
      flush();
      current = {
        timestamp: operatorMatch[1],
        author: 'operador',
        text: '',
      };
      continue;
    }
    if (current) {
      accumulator.push(line);
    }
  }
  flush();
  return entries;
}

function matchVerdictByEmoji(emoji: string): ConvVerdict | undefined {
  for (const [verdict, info] of Object.entries(VERDICT_TO_LABEL)) {
    if (info.emoji === emoji) return verdict as ConvVerdict;
  }
  return undefined;
}

// ────────────────────────────────────────────────────────────────────────────
// Serializer (round-trip idempotente)
// ────────────────────────────────────────────────────────────────────────────

export function serializeOperatorInput(data: OperatorInput): string {
  // Separar HTML comments leading (van ANTES del frontmatter) del resto del preámbulo
  const { leadingHtmlComments, innerPreamble } = splitPreamble(data.preamble ?? '');

  const body: string[] = [];

  // Preámbulo "interior" (heading + blockquote opcional dentro del body)
  if (innerPreamble.length > 0) {
    body.push(innerPreamble);
    body.push('');
  }

  // Sección Notas
  body.push(HEADER_NOTES);
  body.push('');
  if (data.notes.length === 0) {
    // No agregamos placeholder · sección vacía (el cockpit muestra hint en UI)
  } else {
    for (const note of data.notes) {
      body.push(`### ${note.timestamp}`);
      body.push(note.text);
      body.push('');
    }
  }

  // Sección Refs
  body.push(HEADER_REFS);
  body.push('');
  if (data.refs.length === 0) {
    body.push('(sin referencias todavía)');
    body.push('');
  } else {
    for (const ref of data.refs) {
      const emoji = REF_TYPE_TO_EMOJI[ref.type];
      body.push(`- **${emoji} ${ref.type}** · ${ref.value}`);
      if (ref.comment) {
        body.push(`  > ${ref.comment}`);
      }
    }
    body.push('');
  }

  // Sección Conversación
  body.push(HEADER_CONV);
  body.push('');
  for (const entry of data.conversation) {
    if (entry.author === 'claude') {
      const verdict = entry.verdict ?? 'applied';
      const { emoji, label } = VERDICT_TO_LABEL[verdict];
      const skill = entry.skill ?? 'unknown';
      body.push(`### ${entry.timestamp} · 🤖 claude · \`/${skill}\` · ${emoji} ${label}`);
    } else {
      // 'operador' (o alias legacy 'chris' parseado de archivos viejos) → se
      // serializa SIEMPRE como 'operador'.
      body.push(`### ${entry.timestamp} · 🧑 operador`);
    }
    body.push(entry.text);
    body.push('');
  }

  // Update counters en frontmatter (autocalc) sin mutar el original
  const frontmatter: OperatorInputFrontmatter = {
    ...data.frontmatter,
    notes_count: data.notes.length,
    refs_count: data.refs.length,
    conversation_count: data.conversation.length,
  };

  const matterOut = matter.stringify(body.join('\n').replace(/\n+$/, '\n'), frontmatter);

  // Si tenemos HTML comments leading, prependemos antes del frontmatter
  if (leadingHtmlComments.length > 0) {
    return `${leadingHtmlComments}\n${matterOut}`;
  }
  return matterOut;
}

/**
 * Split del preamble: HTML comments leading (van ANTES del frontmatter al
 * serializar) vs preamble "interior" (heading + blockquote del cuerpo).
 */
function splitPreamble(preamble: string): {
  leadingHtmlComments: string;
  innerPreamble: string;
} {
  if (!preamble) return { leadingHtmlComments: '', innerPreamble: '' };
  const lines = preamble.split('\n');
  const htmlLines: string[] = [];
  let i = 0;
  while (i < lines.length) {
    const ln = lines[i];
    if (ln.startsWith('<!--')) {
      htmlLines.push(ln);
      i += 1;
      continue;
    }
    if (ln.trim() === '' && htmlLines.length > 0) {
      i += 1;
      continue;
    }
    break;
  }
  return {
    leadingHtmlComments: htmlLines.join('\n'),
    innerPreamble: lines.slice(i).join('\n').trim(),
  };
}

// ────────────────────────────────────────────────────────────────────────────
// Append helper · usado por skills al cerrar un turn
// ────────────────────────────────────────────────────────────────────────────

/**
 * Appendea una entry a la sección Conversación. Atomic: lee el archivo
 * completo, parsea, agrega entry, serializa y escribe atomic.
 *
 * Justificación: append-text-only no funciona acá porque el frontmatter
 * (counters) debe actualizarse — eso requiere parse + serialize.
 */
export async function appendConversationEntry(
  absPath: string,
  entry: ConvEntry
): Promise<void> {
  let data: OperatorInput;
  try {
    const raw = await readFile(absPath, 'utf-8');
    data = parseOperatorInput(raw);
  } catch (err) {
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') {
      throw new Error(`operator-input no existe en ${absPath} · skill debe crearlo primero`);
    }
    throw err;
  }
  data.conversation.push(entry);
  data.frontmatter.last_modified = nowIso();
  data.frontmatter.conversation_count = data.conversation.length;
  const serialized = serializeOperatorInput(data);
  await writeFileAtomic(absPath, serialized);
}

function nowIso(): string {
  const d = new Date();
  const tzOffsetMin = -d.getTimezoneOffset();
  const sign = tzOffsetMin >= 0 ? '+' : '-';
  const abs = Math.abs(tzOffsetMin);
  const hh = String(Math.floor(abs / 60)).padStart(2, '0');
  const mm = String(abs % 60).padStart(2, '0');
  const iso = d.toISOString().replace('Z', '');
  return `${iso.split('.')[0]}${sign}${hh}:${mm}`;
}

/**
 * Variante low-level: appendea un bloque markdown raw al final del archivo.
 * El skill compone el bloque verbatim y el cockpit lo recalcula al
 * próximo read (chokidar dispara → re-parse → counters se recalculan).
 *
 * Útil cuando el skill no quiere overhead de parse/serialize.
 */
export async function appendRawConvBlock(absPath: string, block: string): Promise<void> {
  await appendToFile(absPath, `\n${block.trim()}\n`);
}
