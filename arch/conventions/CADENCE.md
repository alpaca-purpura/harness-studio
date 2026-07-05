# CADENCE — cómo viven las convenciones de código

> **Las convenciones no se revisan por calendario; se revisan al cambiar** — cuando una regla nace,
> muta, o una herramienta sube de versión mayor. Hereda el mecanismo de [`../CADENCE.md`](../CADENCE.md)
> (el árbol de arquitectura); este doc solo anota lo específico de `conventions/`. El «por qué» firmado
> vive en el [`../../LEDGER.md`](../../LEDGER.md) (ficha HS-05); nunca se duplica aquí.

## Por qué existe (y por qué NO es un STYLEGUIDE.md que se pudre)

Un documento de estilo en prosa («usa nombres descriptivos», «maneja los errores») no rompe nada: el
código deriva y la doc miente. La regla de la casa: **la convención vive como config ejecutable.** El
naming es una regla de `revive`/`useNamingConvention`, no un párrafo; el formato es `gofumpt`/`biome`,
no una guía. Si una convención no se puede romper en CI, no entra a este árbol — va al research doc o la
ficha.

## Anatomía de un convention node (idéntica a boundaries)

```markdown
---
regla: <slug>
version: <n.m>
updated: <YYYY-MM-DD>
status: proposed | enforced | advisory | superseded
ledger: HS-NN
sources: [{ url, autoridad, revisado }]     # L1: la convención como estándar, fechada
enforced_by: [ <ruta al config REAL> ]      # /.golangci.yml, web/biome.json, /lefthook.yml…
severity: critical | high | medium
---
## L1 · Principio (estándar de industria)
## L2 · Realización (este repo Go+TS+Rust)   # con ⇐ L1: <ancla> en cada mapeo
## Checklist evaluable                        # tabla id · qué chequea · severidad · señal · enforcer
## Changelog
```

## El ritual (al cambiar)

1. **Cambio de convención** (nueva regla, o `golangci-lint`/`biome`/`lefthook` sube mayor) → nace en una
   conversación/ficha `HS-NN`.
2. **¿Rompe CI?** Si la regla se puede enforçar → nodo aquí (nuevo o bump) + edición del config real. Si
   es preferencia sin enforcer → no entra (research doc/ficha).
3. **L1 + L2.** Convención con fuente + cómo se realiza. Divergencia L2 vs L1 → `⚠ divergencia`
   justificada, nunca silenciosa.
4. **Config real + `enforced_by:`.** El nodo referencia el archivo (`/.golangci.yml`, `web/biome.json`…);
   se edita el archivo, no una copia en el nodo.
5. **Bump + changelog.** Subir `version`/`updated`; anotar la versión de la herramienta.

## Reglas del árbol (no re-negociar)

- **El config es la verdad; el nodo lo referencia.** Cero duplicación de reglas en el markdown.
- **Toda convención tiene `enforced_by:`** — si no rompe CI, es una nota.
- **Superar, no borrar** (aditivo): una convención superada → `status: superseded` con puntero.
- **lefthook = feedback, CI = enforcement.** El hook local es saltable (`--no-verify`); el required
  status check de `ci.yml` es la garantía dura. Nunca se pinta «enforced» sin el check en CI.
- **Honestidad:** `status: proposed` mientras no haya código; paths y module path **provisionales**
  hasta scaffold.

## Hacia dónde va

En fase 5, cuando `arnesia` y `web/` aterricen: `golangci-lint run` + `biome ci` + `tsc --noEmit` +
`cargo clippy` corren en CI y **rompen el merge**; `lefthook` da el feedback local; `arnesia
conformance` suma estos 26 checks a los de `boundaries/`, `fitness/` y `knowledge/`. La misma disciplina
que exigimos a los arneses que fabricamos, aplicada a la fábrica (dogfood).
