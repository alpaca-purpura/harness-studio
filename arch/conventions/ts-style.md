---
regla: ts-style
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://biomejs.dev/linter/rules/use-exhaustive-dependencies/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://biomejs.dev/linter/rules/use-hook-at-top-level/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://www.pkgpulse.com/guides/biome-vs-eslint-vs-oxlint-2026
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - web/biome.json
severity: high
---

# TS/React: Biome v2.4 (lint + format + organize-imports + react-hooks) en un binario

## L1 · Principio (estándar de industria)

**Biome v2.4 (jul-2026)** = un binario Rust: formatter (~97% Prettier-compat) + linter (200+ reglas) +
organize-imports nativo + **type-aware linting** (~75% de lo que caza typescript-eslint) + lee
`.editorconfig`. **El dato que cambia el debate 2026:** Biome promovió `useExhaustiveDependencies` +
`useHookAtTopLevel` a **STABLE** → el histórico «quédate con ESLint porque react-hooks es no-negociable»
ya no aplica. *(oficial: Biome react-hooks rules)*

**Para equipo chico low-maintenance:** un config, ~20× más rápido, cero Node solo-para-lintear (alinea
con daemon Go + lefthook Go). El gap type-aware (~25%) lo cubre `tsc --noEmit` estricto (ver
[`ts-types.md`](./ts-types.md)). *(experto: Biome vs ESLint vs oxlint 2026)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`web/biome.json`**. Primario para format + lint + `organizeImports` + react-hooks
(`useExhaustiveDependencies` con hooks estables custom, `useHookAtTopLevel`). ⇐ L1: un binario. Gate de
tipos = `tsc --noEmit` (nodo `ts-types`). `biome ci .` (modo CI, sin auto-fix) rompe en cualquier
violación de lint/format/orden de imports.

- **oxlint descartado** (formatter `oxfmt` inmaduro 2026). ⇐ L1.
- **Criterio de auditoría para saltar a ESLint** (documentado, no silencioso): si
  `no-floating-promises`/`no-misused-promises`/`no-unsafe-*` al 100% se vuelven red de seguridad dura
  load-bearing → ESLint 9 flat + typescript-eslint `strict-type-checked` + Prettier (costo: +2
  herramientas, Node, CI más lento). Hoy no compensa; `tsc` estricto + Biome gana. ⇐ L1: gap type-aware.
- **Distribución:** Biome se baja por npm o binario standalone; para mantener el árbol Go puro se puede
  instalar el binario (como lefthook), sin arrastrar Node al toolchain de un dev Go.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| biome-ci | `biome ci .` = exit 0 (lint + format + orden de imports) | error | «violación de estilo/lint TS» | web/biome.json |
| react-hooks | `useExhaustiveDependencies` + `useHookAtTopLevel` en verde | error | «hook con deps incompletas o fuera de tope» | web/biome.json |
| organize-imports | los imports están ordenados (assist nativo, determinístico) | warn | «imports desordenados» | web/biome.json |
| no-eslint-prettier | no coexisten ESLint+Prettier con Biome (una sola fuente de estilo) | warn | «doble formateador (drift de estilo)» | revisión |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = Biome v2.4 (react-hooks stable) + gap type-aware
  cubierto por tsc. L2 = `web/biome.json` primario, `biome ci` rompe CI, oxlint descartado, criterio
  explícito para saltar a ESLint. 4 checks.
