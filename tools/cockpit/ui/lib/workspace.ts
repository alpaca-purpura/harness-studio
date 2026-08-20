/**
 * Workspace resolver — encuentra el root del workspace y detecta sistemas activos.
 *
 * Estrategia root:
 * 1. `process.env.WORKSPACE_ROOT` (override explícito)
 * 2. `git rev-parse --show-toplevel` desde cwd
 * 3. Walk hacia arriba buscando `pnpm-workspace.yaml` o `.git/`
 *
 * Sistemas (F-4 · sin lista hardcodeada): se derivan del seam
 * `project.config.yaml` (`sistemas.active[].slug`) + discovery en disco
 * (todo dir `{root}/{slug}/docs/product/`).
 */

import { execSync } from 'node:child_process';
import { existsSync, readdirSync, statSync } from 'node:fs';
import path from 'node:path';
import { PLATFORM_SLUG } from './platform-context';
// Ciclo benigno workspace ⇄ project-config: ambos solo usan funciones del otro
// en CALL-time (hoisted), nunca en module-load — seguro en ESM + bundler.
import { getActiveSistemaSlugs } from './project-config';

let cachedRoot: string | null = null;

/** Devuelve el path absoluto al workspace root del adopter. Throw si no se encuentra. */
export function getWorkspaceRoot(): string {
  if (cachedRoot) return cachedRoot;

  if (process.env.WORKSPACE_ROOT) {
    const root = path.resolve(process.env.WORKSPACE_ROOT);
    if (!existsSync(root)) {
      throw new Error(`WORKSPACE_ROOT no existe en disco: ${root}`);
    }
    cachedRoot = root;
    return root;
  }

  // Intento 1: git rev-parse desde cwd
  try {
    const root = execSync('git rev-parse --show-toplevel', {
      encoding: 'utf-8',
      cwd: process.cwd(),
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
    if (root && existsSync(root)) {
      cachedRoot = root;
      return root;
    }
  } catch {
    // git no disponible o no estamos en un repo · seguir
  }

  // Intento 2: walk ancestros buscando pnpm-workspace.yaml
  let dir = process.cwd();
  while (dir !== path.dirname(dir)) {
    if (
      existsSync(path.join(dir, 'pnpm-workspace.yaml')) ||
      existsSync(path.join(dir, '.git'))
    ) {
      cachedRoot = dir;
      return dir;
    }
    dir = path.dirname(dir);
  }

  throw new Error(
    'No se pudo resolver WORKSPACE_ROOT · setea env var WORKSPACE_ROOT o ejecuta desde dentro de un git repo'
  );
}

/** Reset cache · útil para tests */
export function _resetWorkspaceCache(): void {
  cachedRoot = null;
}

/** ¿`{root}/{slug}/docs/product/` existe en disco? */
function isBootstrappedSistema(root: string, slug: string): boolean {
  const productDir = path.join(root, slug, 'docs', 'product');
  return existsSync(productDir) && statSync(productDir).isDirectory();
}

/** Slug sano (sin path traversal · sin segmentos raros). */
function isValidSlug(slug: string): boolean {
  return /^[a-z0-9][a-z0-9_-]*$/i.test(slug);
}

/** Slugs declarados en `sistemas.active` del seam. [] si el seam no existe/parsea. */
function seamSistemaSlugs(): string[] {
  try {
    return getActiveSistemaSlugs().filter(isValidSlug);
  } catch {
    return [];
  }
}

/**
 * Sistemas disponibles: las del seam (`sistemas.active`, en su orden) que estén
 * bootstrapeadas en disco + cualquier otra `{slug}/docs/product/` descubierta
 * en disco (orden alfabético al final). Sin lista hardcodeada (F-4).
 */
export function getSistemas(): string[] {
  const root = getWorkspaceRoot();

  const fromSeam = seamSistemaSlugs().filter((slug) => isBootstrappedSistema(root, slug));

  let discovered: string[] = [];
  try {
    discovered = readdirSync(root, { withFileTypes: true })
      .filter((d) => d.isDirectory() && !d.name.startsWith('.') && isValidSlug(d.name))
      .map((d) => d.name)
      .filter((slug) => !fromSeam.includes(slug) && isBootstrappedSistema(root, slug))
      .sort();
  } catch {
    // root ilegible · seguimos solo con el seam
  }

  return [...fromSeam, ...discovered];
}

/**
 * Sistemas seleccionables en el cockpit = sistemas reales bootstrapeadas + la
 * pseudo-sistema `platform` (Vía A) si el `docs/product/` raíz existe. Platform va
 * AL FINAL (las reales primero, para no robar el default del worktree).
 */
export function getSelectableSistemas(): string[] {
  const real = getSistemas();
  const root = getWorkspaceRoot();
  const hasPlatform = existsSync(path.join(root, 'docs', 'product'));
  return hasPlatform ? [...real, PLATFORM_SLUG] : real;
}

/**
 * Raíz de docs de un contexto. Sistema real → `{root}/{sistema}/docs`. Pseudo-sistema
 * `platform` → `{root}/docs` (docs platform-level, owner /pm). Punto único
 * de reruteo: todos los path-builders de abajo derivan de acá.
 */
export function sistemaDocsRoot(sistema: string): string {
  const root = getWorkspaceRoot();
  return sistema === PLATFORM_SLUG
    ? path.join(root, 'docs')
    : path.join(root, sistema, 'docs');
}

/** Path al directorio `docs/product/` del contexto. Valida slug sano + bootstrapeado. */
export function sistemaPath(sistema: string): string {
  if (sistema !== PLATFORM_SLUG && !isValidSlug(sistema)) {
    throw new Error(`sistema inválido: ${sistema}`);
  }
  const p = path.join(sistemaDocsRoot(sistema), 'product');
  if (!existsSync(p)) {
    throw new Error(`Contexto ${sistema} no bootstrapeado · falta ${p}`);
  }
  return p;
}

/** Path al directorio archive de stories done del año dado */
export function archivePath(sistema: string, year: number | string): string {
  return path.join(sistemaDocsRoot(sistema), 'archive', String(year), 'stories');
}

/**
 * Path al root del archive del sistema (`{sistema}/docs/archive`), bajo el cual viven
 * los subdirectorios por año (`2026/stories/...`). Usar para iterar todos los años
 * de stories done — NO derivar via `path.dirname(archivePath(...))` (deja un nivel
 * de más: `archive/{year}` en vez de `archive`).
 */
export function archiveRootPath(sistema: string): string {
  return path.join(sistemaDocsRoot(sistema), 'archive');
}

/** Path al directorio de capabilities del sistema */
export function capabilitiesPath(sistema: string): string {
  return path.join(sistemaPath(sistema), 'capabilities');
}

/** Path al directorio de releases del sistema */
export function releasesPath(sistema: string): string {
  return path.join(sistemaPath(sistema), 'releases');
}

/** Path al directorio de stories activas del sistema */
export function storiesPath(sistema: string): string {
  return path.join(sistemaPath(sistema), 'stories');
}

/** Path al directorio de learnings del sistema */
export function learningsPath(sistema: string): string {
  return path.join(sistemaDocsRoot(sistema), 'learnings');
}
