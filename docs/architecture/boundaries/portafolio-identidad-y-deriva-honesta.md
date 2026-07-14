---
regla: portafolio-identidad-y-deriva-honesta
version: 1.0
updated: 2026-07-14
status: enforced
ledger: HS-22
sources:
  - url: https://theupdateframework.io/security/
    autoridad: estándar
    revisado: 2026-07-13
  - url: https://cucumber.io/docs/guides/living-documentation/
    autoridad: experto
    revisado: 2026-07-09
enforced_by:
  - internal/adapters/portafolio/store_test.go:TestStoreNoFusionaProvisional
  - internal/adapters/portafolio/deriva_test.go:TestDerivaNuncaSemver
  - internal/adapters/portafolio/store_test.go:TestStoreDegradaHonesto
  - internal/domain/portafolio_test.go:TestResolverOrigen
severity: high
---

# El Portafolio identifica sin ambigüedad y deriva por contenido, jamás por conveniencia

## L1 · Principio (estándar de industria)

**La identidad de un artefacto distribuido y su estado de integridad son dos preguntas
distintas, y ninguna de las dos admite una respuesta inventada cuando falta el dato.**

- **Content-addressable verification (hash-based integrity).** Sistemas de distribución de
  paquetes maduros (git, Nix, OCI, TUF) verifican integridad por HASH DE CONTENIDO contra una
  referencia inmutable — nunca por metadata mutable (un número de versión, un timestamp, un
  `git status` que solo responde «¿el usuario commiteó?», no «¿el contenido cambió?»). Un
  string de versión puede repetirse por error humano o reuse; el hash no miente sobre el
  contenido. *(estándar: The Update Framework — «versions are not proof of content»)*
- **Living documentation / identidad calificada, no ambigua.** Dos identidades que PARECEN la
  misma cosa (mismo `id`, sin `home` resuelto todavía) no son la misma entrada hasta que un
  humano las fusiona explícitamente — fusionar automáticamente por coincidencia parcial es
  el error clásico de dedup silencioso que junta cosas distintas. *(experto: living
  documentation — la doc conectada a la fuente, nunca inferida)*

## L2 · Realización (este árbol Go)

`internal/domain/portafolio.go` + `internal/adapters/portafolio/{store,deriva}.go` (Slice 0,
HS-22 — modelo firmado en `docs/product/stories/2026-07-10-spike-carga-arneses/` +
`docs/product/stories/2026-07-13-portafolio-slice0-cimientos/`):

1. **Identidad calificada, nunca fusionada por coincidencia (BR-1/C-ID-2).** La clave del
   Portafolio es `(home, id)` canonicalizado (`domain.IdentidadArnes.Clave()`); una identidad
   sin `home` resoluble es PROVISIONAL (`sin-home~id~scope`) y su `Clave()` es
   estructuralmente distinta de cualquier identidad resuelta con el MISMO `id` — el
   `portafolio.Store.Upsert` nunca las mezcla porque nunca comparten clave. La fusión de una
   provisional con una resuelta (cuando el usuario confirma que son la misma) es una acción
   EXPLÍCITA, fuera de Slice 0.
2. **Deriva por hash de contenido, jamás por semver-string (BR-4).** `domain.EstadoDeriva` se
   decide en `portafolio.EvaluarDeriva` comparando `HashFormaPlugin(instalación)` contra
   `HashFormaPlugin(referencia inmutable)` — dos árboles con la MISMA `version` declarada pero
   contenido distinto dan `en-deriva`, nunca `al-hilo` por coincidencia del string. Sin
   referencia local accesible (marketplace no clonado, versión ausente) → `deriva-no-evaluable`
   con motivo — el default honesto, nunca un `al-día` fabricado.
3. **El store degrada, nunca brickea (BR-11/C-N-4).** `portafolio.Store` decodifica su
   envelope entry-wise: una fila corrupta (o el archivo entero ilegible) se reporta aparte en
   `Listar()` y jamás impide abrir el store ni arrancar el daemon — el anti-patrón que
   `arneses.json` corrupto SÍ tenía antes de HS-21.
4. **Procedencia siempre anotada, conflicto siempre visible (BR-3/C-OR-6).**
   `domain.ResolverOrigen` recolecta TODOS los eslabones disponibles (collect-all, nunca
   «para en el primer hit») y aplica la tabla de autoridad de S0-D1 (lock-devstudio ≙
   cc-plugins > git-plugin > manifiesto); un conflicto entre valores de un mismo campo, o una
   discrepancia home≠registry, queda en `Discrepancias[]` — NUNCA una elección silenciosa.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| identidad-calificada-unica | `Store.Upsert` nunca fusiona una identidad provisional con una resuelta del mismo id (C-ID-2) | error | «el Portafolio juntó dos arneses distintos bajo una card» | portafolio/store_test.go:TestStoreNoFusionaProvisional |
| deriva-nunca-semver | `EvaluarDeriva` decide por hash de contenido; misma version + contenido distinto ⇒ en-deriva | error | «deriva reportada al-día por coincidencia de versión, no de contenido» | portafolio/deriva_test.go:TestDerivaNuncaSemver |
| store-degrada-honesto | una entrada corrupta (o el archivo entero ilegible) del store jamás impide `Listar()` ni el boot | error | «el Portafolio no abre por un byte corrupto» | portafolio/store_test.go:TestStoreDegradaHonesto |
| procedencia-anotada | `ResolverOrigen` anota la fuente de cada dato y nunca resuelve un conflicto en silencio | warn | «origen mostrado sin decir de dónde salió, o conflicto oculto» | domain/portafolio_test.go:TestResolverOrigen |

## Changelog

- 2026-07-14 · v1.0 · Nodo fundacional (HS-22, Slice 0 «Cimientos» del Portafolio). L1 =
  content-addressable verification (TUF/git/Nix/OCI: hash de contenido > metadata mutable) +
  living documentation (identidad calificada, nunca fusión por coincidencia). L2 = los 4
  invariantes del dominio nuevo (`domain/portafolio.go` + `adapters/portafolio/{store,deriva}.go`),
  cada uno con su test Go real ya verde en el árbol. Nace `enforced` (código + tests
  shippean juntos, mismo patrón que `superficie-local-confinada` v1.0) — 4 checks.
