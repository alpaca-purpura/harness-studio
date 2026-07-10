---
regla: naming
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://github.com/mgechev/revive
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://biomejs.dev/linter/rules/use-naming-convention/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://go.dev/wiki/CodeReviewComments
    autoridad: estándar
    revisado: 2026-07-05
enforced_by:
  - /.golangci.yml
  - web/biome.json
severity: medium
---

# El naming es un check, no un documento

## L1 · Principio (estándar de industria)

**Go define su naming** (MixedCaps, acrónimos en mayúsculas `HTTP`/`ID`, receivers cortos, `ErrXxx`,
error strings en minúscula sin punto) — no se inventa, se enforça con `revive`. *(estándar: Go Code
Review Comments; oficial: revive)*

**TS/React:** PascalCase para componentes/tipos, camelCase para variables/funciones — enforçado con
`useNamingConvention` de Biome (configurable por selector). *(oficial: Biome useNamingConvention)*

La convención de naming **vive en el config y rompe CI**, nunca en un `.md`.

## L2 · Realización (este repo Go+TS+Rust)

- **Go** (`/.golangci.yml`, `revive.rules`): `var-naming exported receiver-naming unexported-return
  package-comments error-naming error-strings context-as-argument increment-decrement
  redefines-builtin-id`. `govet` complementa (shadow strict; `fieldalignment` OFF por ruido). ⇐ L1: Go
  naming.
- **TS/React** (`web/biome.json`, `useNamingConvention`): `strictCase`, PascalCase para
  componentes/tipos, camelCase para variables/funciones, CONSTANT_CASE opcional. Cubre ~90%; el resto
  vive en revisión. ⇐ L1: TS naming.
- **Rust** (shell): `rustfmt`/`clippy` traen las convenciones de naming del lenguaje (crate mínimo, sin
  reglas extra).

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| go-naming | `revive` en verde (var-naming, error-naming, error-strings, receiver-naming…) | error | «naming Go no idiomático» | /.golangci.yml (revive) |
| ts-naming | `useNamingConvention` en verde (PascalCase componentes/tipos, camelCase resto) | error | «naming TS no convencional» | web/biome.json |
| error-strings | error strings Go en minúscula, sin puntuación final | warn | «error string capitalizado/con punto» | /.golangci.yml (revive) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = Go define su naming (revive) + TS Biome
  useNamingConvention; naming = check no doc. L2 = reglas revive Go + useNamingConvention TS; Rust vía
  clippy. 3 checks.
