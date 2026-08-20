/**
 * workspace · resolución de paths para la pseudo-sistema "platform" (Vía A · HB-27)
 * + discovery genérico de sistemas (F-4: sin lista hardcodeada — toda
 * `{slug}/docs/product/` en disco cuenta como sistema).
 *
 * Platform rerutea al `docs/` RAÍZ del workspace (no `{sistema}/docs/`). Verifica
 * que los path-builders y el selector de sistemas la tratan first-class sin romper
 * la resolución de los sistemas reales.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { mkdtempSync, mkdirSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import {
  _resetWorkspaceCache,
  sistemaDocsRoot,
  storiesPath,
  archiveRootPath,
  learningsPath,
  sistemaPath,
  getSistemas,
  getSelectableSistemas,
} from '../workspace';
import { _resetProjectConfigCache } from '../project-config';

let root: string;

beforeEach(() => {
  root = mkdtempSync(path.join(tmpdir(), 'cockpit-ws-'));
  process.env.WORKSPACE_ROOT = root;
  _resetWorkspaceCache();
  _resetProjectConfigCache();
  // Sistema real bootstrapeada + docs platform-level en root
  mkdirSync(path.join(root, 'main', 'docs', 'product', 'stories'), { recursive: true });
  mkdirSync(path.join(root, 'docs', 'product', 'stories'), { recursive: true });
});

afterEach(() => {
  delete process.env.WORKSPACE_ROOT;
  _resetWorkspaceCache();
  _resetProjectConfigCache();
  rmSync(root, { recursive: true, force: true });
});

describe('workspace · platform pseudo-sistema', () => {
  it('sistemaDocsRoot(platform) apunta al docs/ raíz; sistema real a {sistema}/docs', () => {
    expect(sistemaDocsRoot('platform')).toBe(path.join(root, 'docs'));
    expect(sistemaDocsRoot('main')).toBe(path.join(root, 'main', 'docs'));
  });

  it('storiesPath(platform) resuelve a docs/product/stories raíz', () => {
    expect(storiesPath('platform')).toBe(path.join(root, 'docs', 'product', 'stories'));
    expect(storiesPath('main')).toBe(
      path.join(root, 'main', 'docs', 'product', 'stories')
    );
  });

  it('archiveRootPath + learningsPath de platform van al root', () => {
    expect(archiveRootPath('platform')).toBe(path.join(root, 'docs', 'archive'));
    expect(learningsPath('platform')).toBe(path.join(root, 'docs', 'learnings'));
  });

  it('sistemaPath(platform) NO tira error (es un slug válido)', () => {
    expect(() => sistemaPath('platform')).not.toThrow();
    expect(sistemaPath('platform')).toBe(path.join(root, 'docs', 'product'));
  });

  it('sistemaPath rechaza sistemas no bootstrapeadas (sin docs/product en disco)', () => {
    expect(() => sistemaPath('sistema-fantasma')).toThrow(/no bootstrapeado/);
  });

  it('sistemaPath rechaza slugs con path traversal', () => {
    expect(() => sistemaPath('../etc')).toThrow(/inválido/);
  });

  it('getSistemas descubre cualquier {slug}/docs/product en disco (sin lista hardcodeada)', () => {
    mkdirSync(path.join(root, 'acme', 'docs', 'product'), { recursive: true });
    mkdirSync(path.join(root, 'sin-product', 'docs'), { recursive: true }); // no cuenta
    const sistemas = getSistemas();
    expect(sistemas).toContain('main');
    expect(sistemas).toContain('acme');
    expect(sistemas).not.toContain('sin-product');
  });

  it('getSelectableSistemas incluye platform además de los sistemas reales', () => {
    const selectable = getSelectableSistemas();
    expect(selectable).toContain('main');
    expect(selectable).toContain('platform');
    // platform va al final (las reales primero)
    expect(selectable[selectable.length - 1]).toBe('platform');
  });
});
