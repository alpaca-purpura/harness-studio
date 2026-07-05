---
regla: go-style
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://golangci-lint.run/docs/configuration/file/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://golangci-lint.run/docs/product/migration-guide/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://gist.github.com/maratori/47a4d00457a92aa426dbd48a18776322
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - /.golangci.yml
severity: high
---

# Go idiomático: ~30 linters de alto valor, errores wrapped, cero ruido

## L1 · Principio (estándar de industria)

**golangci-lint v2** es el agregador estándar (2026). Cambios estructurales vigentes: `version: "2"`
obligatorio; `staticcheck` **absorbe `gosimple`+`stylecheck`** (no listarlos); formatters
(`gofumpt`/`goimports`) en sección `formatters:` aparte (`golangci-lint fmt`), no en `linters`; default
`standard` = 5 linters → ampliar. Un daemon con FS/exec/red/SQLite necesita cobertura de correctitud +
seguridad + logging. *(oficial: golangci-lint v2)*

**Curar el set, no habilitar todo.** El consenso (golden-config) rechaza linters dogmáticos/ruidosos que
generan fricción sin valor. *(experto: maratori golden-config)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`/.golangci.yml`** (`version: "2"`). Sobre `default: standard`, `enable:`:
- **correctitud/errores:** `errcheck errorlint nilerr nilnesserr bodyclose noctx rowserrcheck
  sqlclosecheck contextcheck fatcontext makezero` (relevantes por HTTP + SQLite modernc + context).
- **seguridad:** `gosec` (daemon con FS/exec/red) · `depguard` (bans de import duros — complementa el
  grafo de [`../fitness/.go-arch-lint.yml`](../fitness/.go-arch-lint.yml): dominio no importa
  `net/http`/`database/sql`). ⇐ L1: cobertura daemon.
- **logging:** `sloglint` (ya usan slog → fuerza key-value, atributos consistentes).
- **modernización:** `copyloopvar intrange perfsprint usestdlibvars usetesting modernize unconvert
  wastedassign`.
- **estilo/API:** `revive` (naming, ver [`naming.md`](./naming.md)) · `gocritic unparam predeclared
  nolintlint` (obliga justificar cada `//nolint`).
- **tests:** `testifylint tparallel`.
- **formatters:** `gofumpt goimports` (sección aparte; `golangci-lint fmt --diff` rompe CI si el diff no
  está vacío).

**Evitados por ruido** (documentado como criterio): `lll wsl nlreturn err113 exhaustruct varnamelen mnd
godox gochecknoglobals gochecknoinits dupl tagliatelle maintidx`. `go vet`+`staticcheck` corren DENTRO
de golangci-lint (no invocar aparte). ⇐ L1: curar el set.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| golangci-run | `golangci-lint run ./...` = exit 0 (sin hallazgos) | error | «hallazgo de lint Go (correctitud/seguridad/logging)» | /.golangci.yml |
| go-format | `golangci-lint fmt --diff` vacío (gofumpt+goimports) | error | «Go sin formatear (gofumpt/goimports)» | /.golangci.yml (formatters) |
| version-2 | el config declara `version: "2"` y no lista linters fusionados (`gosimple`/`stylecheck`) | warn | «config golangci-lint v1 (parseo roto en v2)» | revisión |
| depguard-boundaries | `depguard` prohíbe imports que violan las capas hexagonales | error | banda Guardia «import prohibido (dominio↔transporte)» | /.golangci.yml (depguard) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = golangci-lint v2 + golden-config curado. L2 = set de
  ~30 linters de alto valor para un daemon (correctitud/seguridad/logging/modernización) + gofumpt/
  goimports en formatters; lista de evitados; depguard complementa el grafo hexagonal. 4 checks.
