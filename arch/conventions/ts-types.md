---
regla: ts-types
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://www.npmjs.com/package/@tsconfig/strictest
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://www.typescriptlang.org/tsconfig
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - web/tsconfig.json
severity: high
---

# Tipos honestos: `@tsconfig/strictest` + overrides Vite; `tsc --noEmit` es el gate

## L1 · Principio (estándar de industria)

**`@tsconfig/strictest`** activa el set completo de honestidad de tipos: `strict`,
`noUncheckedIndexedAccess` (lookups honestos sobre `undefined`), `exactOptionalPropertyTypes` (opcional ≠
«T | undefined» implícito), `noPropertyAccessFromIndexSignature`, `noImplicitOverride`,
`noImplicitReturns`, `noFallthroughCasesInSwitch`, `noUnusedLocals/Parameters`, `isolatedModules`. El
compilador ES el check: `tsc --noEmit` rompe CI en cualquier error de tipo. *(oficial: @tsconfig/
strictest)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`web/tsconfig.json`** extiende `@tsconfig/strictest` + overrides Vite/React: `target: ES2022`,
`module: ESNext`, `moduleResolution: "Bundler"`, `jsx: "react-jsx"`, `noEmit: true`,
`verbatimModuleSyntax: true`, `lib: [ES2023, DOM, DOM.Iterable]`. `tsc --noEmit` cubre el gap type-aware
que Biome (~75%) no caza — juntos cierran la red de tipos. ⇐ L1: tsc = gate.

- **El par que importa:** `noUncheckedIndexedAccess` + `exactOptionalPropertyTypes` (honestidad real
  sobre `undefined`). ⇐ L1.
- **Gotcha documentado** (no silencioso): `exactOptionalPropertyTypes` genera fricción con spread de
  props React y algunos tipos de terceros → mantener; el patrón es `| undefined` explícito o `Partial`.
- **Tipos del dominio se generan, no se escriben:** los tipos de `shared/api/generated` derivan del
  OpenAPI/JSON-Schema (ver [`../boundaries/fe-transporte-independiente.md`](../boundaries/fe-transporte-independiente.md)
  y `../contracts/`), no a mano → evita drift.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| tsc-noemit | `tsc --noEmit` = exit 0 (cero errores de tipo) | error | «error de tipo (red de tipos)» | web/tsconfig.json |
| extiende-strictest | `tsconfig.json` extiende `@tsconfig/strictest` (no strict a mano parcial) | warn | «tsconfig no-strictest (honestidad de tipos incompleta)» | revisión |
| unchecked-indexed | `noUncheckedIndexedAccess` + `exactOptionalPropertyTypes` activos | error | «lookup/opcional deshonesto sobre undefined» | web/tsconfig.json |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = `@tsconfig/strictest` + `tsc --noEmit` como gate. L2
  = `web/tsconfig.json` con overrides Vite; el par unchecked-indexed/exact-optional; gotcha documentado;
  tipos del dominio generados. 3 checks.
