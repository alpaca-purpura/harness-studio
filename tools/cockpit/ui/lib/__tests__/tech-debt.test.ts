/**
 * tech-debt parser tests · parsea la tabla markdown de docs/process/tech-debt.md
 * (carril L3 del CIL) a TechDebtItem[]. Cubre: fila normal, estado en negrita/
 * sufijo, los 3 emojis de severidad propios (🔴🟡🔵), skip de header/separador/
 * placeholder `—`/prosa, y un smoke contra el archivo real.
 */

import { describe, expect, it } from 'vitest';
import { existsSync } from 'node:fs';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import {
  parseTechDebt,
  countTdByEstado,
  TD_ESTADO_ORDER,
  type TechDebtItem,
} from '../tech-debt.js';

// Smoke opcional contra el registro real del WORKSPACE adopter (si está presente).
// El cockpit es standalone: sin workspace el smoke se salta (los unit tests de
// arriba cubren el parser con fixture inline).
const WORKSPACE = process.env.WORKSPACE_ROOT ?? path.resolve(__dirname, '../../../..');
const REAL_FILE = path.join(WORKSPACE, 'docs', 'process', 'tech-debt.md');
const HAS_REAL_FILE = existsSync(REAL_FILE);

const FIXTURE = `# Tech-Debt Register — carril L3 del CIL

> Deuda de código/infra pura. severidad: 🔴 bloquea-pronto · 🟡 fricción · 🔵 mejora.

## Registro

| ID | fecha | sev | deuda (problema → causa raíz → refuerzo) | estado | ref |
|---|---|---|---|---|---|
| — | — | — | (placeholder · no es una fila real) | — | — |
| TD-1 | 2026-06-05 | 🔴 | migración 021 BYTEA → causa drift en clone | **applied** | abc123 |
| TD-2 | 2026-06-05 | 🟡 | ci-parity Dockerfile sin stage test | reported | sentinel |
| TD-3 | 2026-06-06 | 🔵 | refactor helper duplicado | **deferred** (cont. 2) | — |

## Referencias

Esto es prosa · TD-fake no debe parsearse.
`;

describe('parseTechDebt', () => {
  it('parsea solo filas TD-N (skip header, separador, placeholder — y prosa)', () => {
    const items = parseTechDebt(FIXTURE);
    expect(items).toHaveLength(3);
    expect(items.map((i) => i.id)).toEqual(['TD-1', 'TD-2', 'TD-3']);
  });

  it('mapea las 6 columnas de una fila normal', () => {
    const td1 = parseTechDebt(FIXTURE).find((i) => i.id === 'TD-1') as TechDebtItem;
    expect(td1.num).toBe(1);
    expect(td1.fecha).toBe('2026-06-05');
    expect(td1.sevEmoji).toBe('🔴');
    expect(td1.sevLabel).toBe('bloquea-pronto');
    expect(td1.item).toBe('migración 021 BYTEA → causa drift en clone');
    expect(td1.estado).toBe('applied');
    expect(td1.estadoRaw).toBe('**applied**');
    expect(td1.ref).toBe('abc123');
  });

  it('mapea los 3 emojis de severidad propios de L3', () => {
    const items = parseTechDebt(FIXTURE);
    expect(items.find((i) => i.id === 'TD-1')?.sevLabel).toBe('bloquea-pronto'); // 🔴
    expect(items.find((i) => i.id === 'TD-2')?.sevLabel).toBe('friccion'); // 🟡
    expect(items.find((i) => i.id === 'TD-3')?.sevLabel).toBe('mejora'); // 🔵
  });

  it('extrae el estado canónico ignorando negrita y sufijo "(cont. N)"', () => {
    const items = parseTechDebt(FIXTURE);
    expect(items.find((i) => i.id === 'TD-2')?.estado).toBe('reported');
    expect(items.find((i) => i.id === 'TD-3')?.estado).toBe('deferred');
    expect(items.find((i) => i.id === 'TD-3')?.estadoRaw).toBe('**deferred** (cont. 2)');
  });

  it('no crashea con string vacío ni con solo el placeholder', () => {
    expect(parseTechDebt('')).toEqual([]);
    expect(parseTechDebt('| — | — | — | x | — | — |')).toEqual([]);
  });

  it('countTdByEstado agrupa por estado canónico', () => {
    const counts = countTdByEstado(parseTechDebt(FIXTURE));
    expect(counts.applied).toBe(1);
    expect(counts.reported).toBe(1);
    expect(counts.deferred).toBe(1);
  });

  it('TD_ESTADO_ORDER reusa el lifecycle compartido del CIL', () => {
    expect(TD_ESTADO_ORDER).toEqual([
      'reported',
      'triaged',
      'ratified',
      'applied',
      'verified',
      'deferred',
    ]);
  });
});

describe('parseTechDebt · smoke contra el archivo real (skip sin workspace adopter)', () => {
  it.skipIf(!HAS_REAL_FILE)('parsea docs/process/tech-debt.md sin crashear', async () => {
    const md = await readFile(REAL_FILE, 'utf-8');
    const items = parseTechDebt(md);
    // El registro arranca con placeholder `—` → 0 items hasta que dev/auditor appendean.
    for (const it of items) {
      expect(it.id).toMatch(/^TD-\d+$/);
      expect([...TD_ESTADO_ORDER, 'otro']).toContain(it.estado);
    }
  });
});
