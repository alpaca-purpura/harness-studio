---
regla: portafolio-identidad-y-deriva-honesta
version: 1.2
updated: 2026-07-25
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
  - internal/domain/marketplace_test.go:TestMergeMarketplacesAnotaDiscrepancia
  - internal/usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio
  - internal/domain/marketplace_test.go:TestVersionCatalogoNuncaInventada
  - internal/domain/marketplace_situacion_test.go:TestSituacionNoComparableGana
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

**Extensión v1.2 — el mismo principio aplicado al ESTANTE** (paquete
`docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`, `design.md` §11.2). La pregunta
«¿de dónde salió esto y coincide con lo que publicamos?» tiene los mismos dos modos de mentir que
la identidad y la deriva: inventar el dato que falta, o callar que falta:

5. **El registro de marketplaces es collect-all, y una discrepancia se muestra.**
   `domain.MergeMarketplaces` mergea por `nombre` el lado detectado
   (`~/.claude/plugins/known_marketplaces.json` de Claude Code — un 4º eslabón REAL, no una lista
   virgen que el operador llena a mano) con el lado declarado por el operador. Un `repo` distinto
   entre eslabones queda en `Discrepancias[]` con **los dos valores crudos visibles** — exactamente
   el criterio de `ResolverOrigen`, aplicado un nivel arriba.
6. **Un catálogo no legible viaja `null`, nunca `[]`.** `domain.Catalogo.Entradas` es `nil`
   cuando no hubo lectura válida (serializa a `null`: el campo NO lleva `omitempty`) y
   `[]EntradaCatalogo{}` **solo** cuando el marketplace declara `plugins: []` de verdad. Una lista
   vacía se lee como «este marketplace no tiene arneses»: es el mismo pass fabricado que el
   criterio G3 mató en el Slice 1, con otra ropa.
7. **La versión del catálogo nunca se inventa.** `domain.VersionDeEntrada` aplica precedencia
   con procedencia anotada — `plugins[].version` (campo estándar, verificado: 14 semver reales +
   259 `null` explícitos en el catálogo oficial) > último segmento de la ruta del `source` que
   parsee semver > **ausente**. Y dos versiones se comparan como semver o **no se comparan**: si
   alguna no parsea, el resultado es `no-comparable`, jamás una comparación de strings (mismo
   espíritu que «deriva nunca por semver»).
8. **`no-comparable` es una rama de primera clase.** `domain.CalcularSituacion` tiene 6 ramas y
   la 6ta existe porque sin ella el sistema tendría que elegir entre mentir (`al-hilo` por defecto)
   o callarse. Falta cualquier insumo (sin versión del estante, sin canónico, versión no-semver,
   deriva no evaluable, N coincidencias de identidad) ⇒ `no-comparable` **con motivo**; y
   `no-comparable` gana sobre toda afirmación positiva. `al-hilo` es la ÚLTIMA rama de la tabla: se
   alcanza solo cuando ningún insumo falta y ninguna instalación tiene señal.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| identidad-calificada-unica | `Store.Upsert` nunca fusiona una identidad provisional con una resuelta del mismo id (C-ID-2) | error | «el Portafolio juntó dos arneses distintos bajo una card» | internal/adapters/portafolio/store_test.go:TestStoreNoFusionaProvisional |
| deriva-nunca-semver | `EvaluarDeriva` decide por hash de contenido; misma version + contenido distinto ⇒ en-deriva | error | «deriva reportada al-día por coincidencia de versión, no de contenido» | internal/adapters/portafolio/deriva_test.go:TestDerivaNuncaSemver |
| store-degrada-honesto | una entrada corrupta (o el archivo entero ilegible) del store jamás impide `Listar()` ni el boot | error | «el Portafolio no abre por un byte corrupto» | internal/adapters/portafolio/store_test.go:TestStoreDegradaHonesto |
| procedencia-anotada | `ResolverOrigen` anota la fuente de cada dato y nunca resuelve un conflicto en silencio | warn | «origen mostrado sin decir de dónde salió, o conflicto oculto» | internal/domain/portafolio_test.go:TestResolverOrigen |
| registro-marketplaces-collect-all | `MergeMarketplaces` mergea por nombre los dos eslabones (CC + declarado) y deja toda discrepancia de `repo` visible, sin elegir | error | «el registro de marketplaces eligió un repo en silencio, o nació vacío ignorando lo que CC ya conoce» | internal/domain/marketplace_test.go:TestMergeMarketplacesAnotaDiscrepancia |
| catalogo-nunca-vacio-fabricado | un catálogo no legible viaja `entradas: null` + motivo; solo un `plugins: []` REAL viaja `[]` con su fecha de lectura | error | «catálogo vacío que se lee como “este marketplace no tiene arneses”» | internal/usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio |
| version-catalogo-nunca-inventada | la versión de una fila sale del campo o del último segmento semver del `source`; si no, ausente — y la comparación es semver o `no-comparable`, nunca de strings | error | «versión de catálogo fabricada, o comparada como string» | internal/domain/marketplace_test.go:TestVersionCatalogoNuncaInventada |
| situacion-no-comparable-de-primera-clase | falta cualquier insumo ⇒ `no-comparable` con motivo; `no-comparable` gana sobre toda afirmación positiva y `al-hilo` es la última rama | error | «al-hilo afirmado sin poder evaluarlo (verde fabricado en el estante)» | internal/domain/marketplace_situacion_test.go:TestSituacionNoComparableGana |

## Changelog

- 2026-07-25 · v1.2 · **Extensión al ESTANTE** (paquete
  `docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`, etapa 3 · diseño técnico).
  El mismo L1 —identidad y estado de integridad no admiten respuesta inventada cuando falta el
  dato— aplicado a la pregunta nueva del paquete: «¿de dónde salió este arnés y coincide con lo que
  publicamos?». 4 invariantes nuevos (L2 §5-§8): registro collect-all con discrepancia visible ·
  catálogo ilegible viaja `null` y no `[]` · versión del catálogo con precedencia y procedencia,
  nunca inventada, y comparación semver-o-nada · `no-comparable` como rama de primera clase que
  gana a toda afirmación positiva. **La evidencia nueva llegó de la máquina real y corrigió el
  spec:** `plugins[].source` tiene 4 formas (no 1), `version` SÍ es campo estándar del formato
  (259 `null` + 14 semver reales en el catálogo oficial), y ese catálogo son 273 filas / 159 KB
  (no 41) — detalle en el `decisiones.md` del paquete (AG-D13..AG-D16) y en su `design.md` §1.
  El nodo **sigue `enforced`**: los 4 checks originales más los 4 nuevos, **con enforcer real
  cableado** desde la etapa 4 (implementación) del mismo día — tests COLOCADOS junto al código, el
  patrón sancionado en v1.1. 4 → **8 checks**, los 8 corriendo.
- 2026-07-14 · v1.1 · Auditoría post-build: los 4 enforcers son tests COLOCADOS junto al
  código (no viven en `fitness/`), y el motor `arnesia conformance` solo sabía correr el
  paquete fitness — los 4 checks salían `deferred` en `--todo` pese a correr en CI. Se
  extendió el motor (parser `reGoTest` + `mechanism.pkgOf`): un enforcer con ruta
  repo-relativa `dir/foo_test.go:TestX` ahora clasifica arch-test y corre en SU paquete.
  Las celdas de la tabla pasan a ruta completa (el path ES el wiring). Este patrón
  (test colocado > test-proxy en fitness, cf. `healthz-refleja-cors` HS-14) queda
  sancionado para boundaries nuevos.
- 2026-07-14 · v1.0 · Nodo fundacional (HS-22, Slice 0 «Cimientos» del Portafolio). L1 =
  content-addressable verification (TUF/git/Nix/OCI: hash de contenido > metadata mutable) +
  living documentation (identidad calificada, nunca fusión por coincidencia). L2 = los 4
  invariantes del dominio nuevo (`domain/portafolio.go` + `adapters/portafolio/{store,deriva}.go`),
  cada uno con su test Go real ya verde en el árbol. Nace `enforced` (código + tests
  shippean juntos, mismo patrón que `superficie-local-confinada` v1.0) — 4 checks.
