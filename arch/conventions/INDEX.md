# arch/conventions/ — convenciones de código as code de ArnesIA

> **Las convenciones de programación viven como config que rompe CI, no como prosa de «buenas
> prácticas».** Hermano de [`../fitness/`](../fitness/): `fitness/` guarda la **forma arquitectónica**
> (grafo de imports, boundaries); `conventions/` guarda el **estilo/código** (lint, format, naming,
> commits, hooks, CI). Ambos son checks corribles con `enforced_by:`; distinta intención. Espeja el
> patrón de [`../boundaries/`](../INDEX.md) (nodo L1 principio+fuente ↔ L2 realización + checklist).
> Norte: [`../../VISION.md`](../../VISION.md) · evidencia L1:
> [`../../historias/2026-07-05-fe-arch-atomic-storybook-convenciones.md`](../../historias/2026-07-05-fe-arch-atomic-storybook-convenciones.md)
> (frente D) · el «por qué» firmado: ficha **HS-05** del [`../../LEDGER.md`](../../LEDGER.md).

## La disciplina (idéntica a boundaries)

Cada nodo = **L1** (la convención como estándar de industria, con fuente y fecha) ↔ **L2** (cómo se
realiza en ESTE repo Go+TS+Rust) + un **checklist evaluable** cuyo `enforced_by:` apunta al **archivo de
config REAL** (que vive donde la herramienta lo espera: raíz o `web/`), no a una copia. El nodo NO
duplica la config; la referencia — igual que `boundaries/` referencia `go-arch-lint`. Regla dura: **toda
convención en este árbol DEBE romper CI** (si no, es una nota, va al doc de historias o la ficha).

## El árbol — convention nodes

| Nodo | Convención | Estado | Config (enforced_by) | Rompe CI vía |
|------|-----------|--------|----------------------|--------------|
| [`go-style.md`](./go-style.md) | Go idiomático, ~30 linters de alto valor, errores wrapped | 🌱 vivo | `/.golangci.yml` | `golangci-lint run` |
| [`ts-style.md`](./ts-style.md) | Biome v2.4 (lint+format+organize-imports+react-hooks) | 🌱 vivo | `web/biome.json` | `biome ci .` |
| [`ts-types.md`](./ts-types.md) | TS estricto (`@tsconfig/strictest`), tipos honestos | 🌱 vivo | `web/tsconfig.json` | `tsc --noEmit` |
| [`naming.md`](./naming.md) | Naming = check, no doc (revive + useNamingConvention) | 🌱 vivo | `/.golangci.yml` · `web/biome.json` | lint |
| [`commits.md`](./commits.md) | Conventional + scope `HS-NN` | 🌱 vivo | `/lefthook.yml` (commit-msg) | `commit-msg` hook |
| [`git-hooks.md`](./git-hooks.md) | Gate local pre-commit (lefthook, binario Go) | 🌱 vivo | `/lefthook.yml` | hook exit≠0 |
| [`editor.md`](./editor.md) | Indentación/EOL única (EditorConfig) | 🌱 vivo | `/.editorconfig` | (vía format checks) |
| [`ci.md`](./ci.md) | Todo check rompe el merge (GitHub Actions) | 🌱 vivo | `/.github/workflows/ci.yml` | required status checks |

Leyenda: ⏳ en forja · 🌱 vivo (nace, se enforça cuando el código llegue) · 🌳 estable · 🔍 en-revisión.
**Total: 8 convention nodes · 26 checks · pasada fundacional 2026-07-05 (HS-05).**

> **Honestidad (heredada de METODOLOGIA §4 / CADENCE):** el código es real (daemon Go
> `internal/`+`cmd/`, SPA `web/`, shell Rust `web/src-tauri/` — HS-06/HS-08) y los configs
> **corren local Y en CI desde HS-10** (`.github/workflows/ci.yml` verde 3/3: go+ts+rust;
> `lefthook` instalado como gate local). Los paths (`web/src/**`, module path Go) dejaron de ser
> provisionales.

## Decisiones clave (detalle y fuentes en el doc de historias, frente D)

| Capa | Pick | Descartado |
|---|---|---|
| Go lint+format | **golangci-lint v2** (~30 linters) + gofumpt/goimports | linters ruidosos (`lll`, `wsl`, `exhaustruct`, `mnd`, `err113`…) |
| TS lint+format | **Biome v2.4** + `tsc --noEmit` | ESLint+Prettier (más herramientas, Node en el loop) · oxlint (formatter inmaduro) |
| tsconfig | **`@tsconfig/strictest`** + overrides Vite | strict manual |
| pre-commit | **lefthook** (binario Go, políglota) | husky+lint-staged (exige Node) |
| commits | conventional + `HS-NN`, regex en lefthook | commitlint (reintroduce Node; opcional) |

## Cómo crece

Como [`../CADENCE.md`](../CADENCE.md): **al cambiar, no por calendario.** Subir versión de una
herramienta o cambiar una regla = bump del nodo + changelog. Mecanismo en
[`./CADENCE.md`](./CADENCE.md). El runner futuro **`arnesia conformance`** corre `knowledge/` + `arch/`
(boundaries + fitness) + `arch/conventions/` en un solo reporte.
