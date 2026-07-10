---
regla: commits
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://www.conventionalcommits.org/en/v1.0.0/
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://lefthook.dev/examples/commitlint/
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - /lefthook.yml
severity: medium
---

# Commits: Conventional Commits + scope `HS-NN`

## L1 · Principio (estándar de industria)

**Conventional Commits 1.0** (`type(scope): descripción`) da un historial parseable (releases, changelog,
semver). *(estándar: Conventional Commits)*

## L2 · Realización (este repo Go+TS+Rust)

El repo ya usa el formato `feat: HS-NN — …` (conventional con scope de ficha). Se valida en el hook
`commit-msg` de **`/lefthook.yml`** con **regex Node-free** (low-maintenance):
`^(feat|fix|docs|refactor|chore|test|perf|build|ci)(\(.+\))?: ` — cero dependencias. ⇐ L1: parseable.

- Alternativa `commitlint` (`@commitlint/config-conventional` vía `commit-msg`) queda documentada como
  opción si se quiere el ecosistema (case/longitud de subject), pero **reintroduce Node** → no default.
- El scope `HS-NN` y el em-dash del formato de la casa se aceptan (quedan en el subject); enforçar el
  scope `HS-NN` estricto requeriría una regla custom (no obligatoria hoy).

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| conventional-type | el subject empieza con un type válido (`feat/fix/docs/…`) + `: ` | error | «commit sin type Conventional» | /lefthook.yml (commit-msg) |
| node-free | la validación no arrastra Node (regex en lefthook, no commitlint por default) | warn | «commitlint añade Node al toolchain» | revisión |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = Conventional Commits 1.0. L2 = regex Node-free en
  `commit-msg` de lefthook; commitlint documentado como opción no-default; scope `HS-NN` aceptado. 2
  checks.
