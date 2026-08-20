/**
 * Filesystem reader helpers — markdown con frontmatter + YAML.
 */

import { readFile, readdir, stat } from 'node:fs/promises';
import path from 'node:path';
import matter from 'gray-matter';
import YAML from 'yaml';

export interface MarkdownWithFrontmatter {
  frontmatter: Record<string, unknown>;
  content: string;
  /** Raw del archivo (útil para round-trips serializados) */
  raw: string;
}

/** Lee markdown + parse frontmatter via gray-matter */
export async function readMarkdownWithFrontmatter(
  absPath: string
): Promise<MarkdownWithFrontmatter> {
  const raw = await readFile(absPath, 'utf-8');
  const parsed = matter(raw);
  return {
    frontmatter: parsed.data,
    content: parsed.content,
    raw,
  };
}

/** Lee y parsea un archivo YAML standalone (no markdown) */
export async function readYaml<T = unknown>(absPath: string): Promise<T> {
  const raw = await readFile(absPath, 'utf-8');
  // Soporta YAML con tres-dashes leading (estilo frontmatter) o sin ellos
  const trimmed = raw.replace(/^---\s*\n/, '').replace(/\n---\s*$/, '');
  return YAML.parse(trimmed) as T;
}

/** Lee YAML que vive dentro de un archivo .md (frontmatter mode estilo capabilities) */
export async function readYamlFromMarkdown<T = unknown>(absPath: string): Promise<T> {
  const parsed = await readMarkdownWithFrontmatter(absPath);
  return parsed.frontmatter as T;
}

/**
 * Glob simple sin dependencia externa. Soporta:
 * - `*.ext` (un nivel)
 * - `**` (recursivo)
 * - paths absolutos o relativos
 *
 * Devuelve paths absolutos.
 */
export async function globPaths(pattern: string, baseDir?: string): Promise<string[]> {
  const base = baseDir ? path.resolve(baseDir) : process.cwd();
  const segments = pattern.split('/');
  const results: string[] = [];

  async function walk(currentDir: string, segmentIdx: number): Promise<void> {
    if (segmentIdx >= segments.length) {
      // Match exacto al final
      try {
        const stats = await stat(currentDir);
        if (stats.isFile()) results.push(currentDir);
      } catch {
        // skip
      }
      return;
    }

    const segment = segments[segmentIdx];

    if (segment === '**') {
      // Recursive: probar el resto desde currentDir y todos los descendientes
      await walk(currentDir, segmentIdx + 1);
      try {
        const entries = await readdir(currentDir, { withFileTypes: true });
        for (const entry of entries) {
          if (entry.isDirectory()) {
            const child = path.join(currentDir, entry.name);
            await walk(child, segmentIdx); // sigue en **
          }
        }
      } catch {
        // skip
      }
      return;
    }

    // Wildcard simple `*` con extensión opcional
    if (segment.includes('*')) {
      const regex = wildcardToRegex(segment);
      try {
        const entries = await readdir(currentDir, { withFileTypes: true });
        for (const entry of entries) {
          if (regex.test(entry.name)) {
            const child = path.join(currentDir, entry.name);
            if (segmentIdx === segments.length - 1) {
              // Último segmento · match si es file
              if (entry.isFile()) results.push(child);
            } else if (entry.isDirectory()) {
              await walk(child, segmentIdx + 1);
            }
          }
        }
      } catch {
        // skip
      }
      return;
    }

    // Match literal
    const child = path.join(currentDir, segment);
    await walk(child, segmentIdx + 1);
  }

  await walk(base, 0);
  return results.sort();
}

function wildcardToRegex(pattern: string): RegExp {
  // Escapa regex specials excepto `*`
  const escaped = pattern.replace(/[.+?^${}()|[\]\\]/g, '\\$&').replace(/\*/g, '.*');
  return new RegExp(`^${escaped}$`);
}

/** Lista archivos directos de un directorio (no recursivo) */
export async function listFiles(absDir: string): Promise<string[]> {
  const entries = await readdir(absDir, { withFileTypes: true });
  return entries
    .filter((e) => e.isFile())
    .map((e) => path.join(absDir, e.name))
    .sort();
}

/** Lista subdirectorios directos */
export async function listDirs(absDir: string): Promise<string[]> {
  try {
    const entries = await readdir(absDir, { withFileTypes: true });
    return entries
      .filter((e) => e.isDirectory())
      .map((e) => path.join(absDir, e.name))
      .sort();
  } catch {
    return [];
  }
}
