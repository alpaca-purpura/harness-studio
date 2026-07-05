---
regla: ci
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://github.com/golangci/golangci-lint-action
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://docs.github.com/en/actions/how-tos/manage-workflow-runs/require-status-checks-before-merging
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - /.github/workflows/ci.yml
severity: critical
---

# Todo check rompe el merge (GitHub Actions, required status checks)

## L1 · Principio (estándar de industria)

**El gate duro es CI, no el hook local.** Los hooks son saltables (`--no-verify`); la garantía es un
**required status check** que bloquea el merge si un job falla. Cada convención/boundary de este árbol
debe tener su job. *(oficial: golangci-lint-action; GitHub required checks)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`/.github/workflows/ci.yml`**, jobs paralelos, cada uno un required check en `main`:
- **Go:** `golangci/golangci-lint-action` (v7 para v2) → `golangci-lint run` + `golangci-lint fmt --diff`
  · `go test ./... -race` · `go build` · `go-arch-lint check` (grafo de [`../fitness/`](../fitness/)).
- **TS:** `biome ci .` · `tsc --noEmit` · `dependency-cruiser` (boundaries FE) · `stylelint` (tokens) ·
  `vitest --project=storybook` (fitness visual) · `steiger` (FSD-estructural).
- **Rust (shell):** `cargo clippy --all-targets -- -D warnings` · `cargo fmt --check`.
- **Contratos:** `openapi`/`json-schema` gen-check (los tipos generados están en sync) + `tokens-sync`
  (`.tokens.json` ⟷ `@theme`).
- **Node ≥ 20.16/22.19/24** (Storybook 10 ESM-only). ⇐ L1: required checks.

Trunk-based (push directo a `main`): los checks corren post-push y **rompen el árbol** si algo falla; o
se usa merge queue. Este job es el que vuelve «enforced» todos los `status: proposed` cuando el código
aterrice. ⇐ L1: el gate duro es CI.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| ci-go | job Go (golangci-lint + test -race + build + go-arch-lint) verde | error | banda Guardia «CI Go roja» | ci.yml |
| ci-ts | job TS (biome + tsc + dependency-cruiser + stylelint + vitest-storybook) verde | error | banda Guardia «CI TS roja» | ci.yml |
| ci-rust | job Rust (clippy -D warnings + fmt --check) verde | warn | «CI shell roja» | ci.yml |
| ci-contratos | tipos generados y tokens en sync (gen-check + tokens-sync) | error | «drift contrato↔código» | ci.yml |
| required-checks | los jobs son required status checks en `main` (bloquean merge/rompen árbol) | error | «check no bloqueante (gate blando)» | ci.yml + branch protection |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = el gate duro es CI (required checks), no el hook. L2
  = `/.github/workflows/ci.yml` con jobs Go/TS/Rust/contratos paralelos; enumera todos los enforcers de
  boundaries+conventions; el job que vuelve «enforced» los proposed. 5 checks.
