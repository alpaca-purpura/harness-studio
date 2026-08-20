/**
 * Filesystem writer helpers — atomic write + append-only.
 *
 * Doctrina: nunca escribir directo al path final. Se escribe a `{path}.tmp`
 * y luego se hace rename (atomic en POSIX). Esto evita que el cockpit lea
 * un archivo a mitad de escritura (chokidar dispararía change con contenido
 * corrupto).
 */

import { mkdir, rename, writeFile, appendFile, readFile } from 'node:fs/promises';
import path from 'node:path';
import matter from 'gray-matter';

/** Asegura que el directorio padre exista antes de escribir */
async function ensureDir(filePath: string): Promise<void> {
  await mkdir(path.dirname(filePath), { recursive: true });
}

/** Write atómico: escribe a `.tmp` y rename al final */
export async function writeFileAtomic(absPath: string, content: string): Promise<void> {
  await ensureDir(absPath);
  const tmp = `${absPath}.tmp`;
  await writeFile(tmp, content, 'utf-8');
  await rename(tmp, absPath);
}

/**
 * Escribe markdown con frontmatter YAML.
 * Usa gray-matter para serializar de forma idempotente.
 */
export async function writeMarkdownWithFrontmatter(
  absPath: string,
  frontmatter: Record<string, unknown>,
  body: string
): Promise<void> {
  const serialized = matter.stringify(body, frontmatter);
  await writeFileAtomic(absPath, serialized);
}

/**
 * Append text al final de un archivo (no atomic en sentido estricto — fs.appendFile
 * usa append-mode del kernel, pero garantiza atomicidad por write a nivel de bytes
 * en sistemas POSIX para writes <= PIPE_BUF).
 *
 * Si el archivo no existe, lo crea. Si existe, garantiza que el contenido
 * nuevo empieza en una nueva línea (insertando `\n` si el archivo no termina
 * con uno).
 */
export async function appendToFile(absPath: string, content: string): Promise<void> {
  await ensureDir(absPath);
  // Verificar si el archivo termina con newline
  try {
    const existing = await readFile(absPath, 'utf-8');
    const needsNewline = existing.length > 0 && !existing.endsWith('\n');
    const payload = needsNewline ? `\n${content}` : content;
    await appendFile(absPath, payload, 'utf-8');
  } catch (err) {
    // Archivo no existe — crearlo
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') {
      await writeFile(absPath, content, 'utf-8');
    } else {
      throw err;
    }
  }
}

/** Write YAML standalone (no markdown) · atomic */
export async function writeYamlAtomic(
  absPath: string,
  data: Record<string, unknown>
): Promise<void> {
  const YAML = await import('yaml');
  const serialized = YAML.stringify(data);
  await writeFileAtomic(absPath, serialized);
}
