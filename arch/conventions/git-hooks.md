---
regla: hooks
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://lefthook.dev/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://www.pkgpulse.com/guides/husky-vs-lefthook-vs-lint-staged-git-hooks-nodejs-2026
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - /lefthook.yml
severity: medium
---

# Gate local pre-commit con lefthook (binario Go, políglota)

## L1 · Principio (estándar de industria)

**lefthook** (binario Go, paralelo, políglota) gana sobre husky+lint-staged en un repo **Go+TS mixto**:
un config único (`lefthook.yml`), ~10× más rápido, cero arranque Node, `{staged_files}` integrado (lo
que hacía lint-staged). No obliga a instalar Node solo para commitear — un dev Go no necesita el
toolchain Node. *(oficial: lefthook; experto: hooks comparados 2026)*

**lefthook = feedback, CI = enforcement.** El hook local es saltable con `--no-verify`; la garantía dura
es el required status check de CI (ver [`ci.md`](./ci.md)). *(experto)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`/lefthook.yml`**. `pre-commit` paralelo por glob: `*.go` → `golangci-lint run
--new-from-rev=HEAD {staged_files}` + `golangci-lint fmt` · `*.{ts,tsx}` → `biome check --write --staged
{staged_files}` · `*.rs` → `cargo fmt`. `commit-msg` → validación de [`commits.md`](./commits.md). ⇐ L1:
políglota, staged-only.

- Instalable Node-free (binario o `go install github.com/evilmartians/lefthook@latest`) para mantener el
  árbol Go puro. ⇐ L1: cero Node.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| lefthook-precommit | `lefthook.yml` corre lint/format Go+TS+Rust sobre `{staged_files}` en pre-commit | warn | «pre-commit no cablea el lint por lenguaje» | /lefthook.yml |
| node-free-hooks | el gate local no exige Node para un dev Go (lefthook binario) | warn | «hooks atados a Node (husky)» | revisión |
| ci-es-la-garantia | existe el required check de CI (el hook local es saltable) | error | banda Guardia «sin gate duro en CI (hook saltable)» | ci.yml (ver ci.md) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = lefthook binario Go políglota > husky; hook =
  feedback, CI = enforcement. L2 = `/lefthook.yml` pre-commit por glob (Go/TS/Rust) staged-only, Node-
  free. 3 checks.
