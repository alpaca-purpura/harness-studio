---
regla: no-aplica-no-es-cero
version: 1.0
updated: 2026-07-26
status: enforced
ledger: HS-28
sources:
  - url: https://prometheus.io/docs/prometheus/latest/querying/basics/#staleness
    autoridad: oficial
    revisado: 2026-07-26
  - url: https://www.postgresql.org/docs/current/functions-comparison.html
    autoridad: oficial
    revisado: 2026-07-26
    # Reemplaza a sqlite.org/nulls.html, que se abrió y NO sostiene la afirmación: esa página
    # compara CÓMO tratan el NULL los motores, no define su semántica. La de PostgreSQL sí.
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/investigacion-runtimes.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by:
  - internal/usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio
  - internal/domain/marketplace_situacion_test.go:TestSituacionNoComparableSinVersionEstante
  - internal/adapters/portafolio/deriva_test.go:TestEvaluarDerivaNoEvaluable
severity: error
---

# «No aplica» y «no lo sé» son valores; `0` y `[]` son afirmaciones

## L1 · Principio (estándar de industria)

**La ausencia de dato no es un dato con valor cero.** Es la misma distinción que sostienen dos
sistemas maduros por caminos distintos:

- **Lógica de tres valores.** En SQL, `NULL` significa *desconocido*: no es `0`, no es `''`, y no
  es igual a otro `NULL`. Colapsarlo a un valor concreto convierte una pregunta abierta en una
  respuesta falsa. *(oficial: PostgreSQL — comparison functions, textual: «Ordinary comparison
  operators yield null (signifying "unknown"), not true or false, when either input is null» ·
  «The null value represents an unknown value, and it is not known whether two unknown values are
  equal»)*
- **Staleness markers.** Prometheus marca explícitamente una serie que dejó de reportar, en vez de
  dejar que el último valor se extienda o que el gráfico caiga a 0: un servicio que se cayó y un
  servicio con carga cero **se ven distinto**, y tienen que verse distinto. *(oficial: Prometheus —
  querying basics/staleness, textual: «If a target scrape or rule evaluation no longer returns a
  sample for a time series that was previously present, this time series will be marked as stale»
  · «If a query is evaluated at a sampling timestamp after a time series is marked as stale, then
  no value is returned for that time series»)*

Corolario que este árbol agrega, y que es doctrina de la casa antes que de la industria: **en un
esquema superset multi-proveedor hay una tercera categoría.** No es solo «lo sé» vs «no lo sé»:
está «**este proveedor no tiene el concepto**». Un `0` en «tokens de razonamiento» para un modelo
que no razona no es un dato faltante ni un dato medido: es una **categoría que no existe**, y
mostrarla como 0 invita a sumarla, promediarla y compararla con la de un modelo que sí razona.

## L2 · Realización (este árbol Go+React)

La regla ya vivía repartida en tres lugares del árbol, **cada uno resolviéndola por su cuenta**.
Este nodo la nombra una vez y las tres pasan a ser instancias suyas:

| instancia | qué distingue | dónde |
|---|---|---|
| **`entradas: null` ≠ `[]`** | «no pude leer el catálogo» vs «leí y no declara ninguno» | BR-4, `usecase/marketplace.go`; el `openapi.yaml` ya lo advierte textual: *«convertir el null en [] es el pass fabricado que BR-4 mata»* |
| **`no-comparable` de primera clase** | no es «al día» ni «en deriva»: es que **no hay con qué comparar** | `domain/marketplace_situacion.go` (`portafolio-identidad-y-deriva-honesta` v1.2) |
| **`deriva-no-evaluable` con motivo** | sin versión o sin referencia local, el veredicto es *no evaluable*, jamás `al-hilo` | `adapters/portafolio/deriva.go` (S0-D7) |
| **`deferido` del motor de conformance** | un mecanismo que no puede correr acá difiere honesto, nunca fabrica `pass` | `domain/conformance.go`, `VeredictoDiferido` |
| **`no-reconocido` / `sin-check`** | el loader deja visible lo que no supo clasificar | `adapters/loader/loader.go` |

Y las instancias que agrega el módulo `telemetria/`:

- **Los buckets de tokens son punteros** (`*int64`), no enteros. `nil` = «este runtime no tiene el
  concepto» (D-4/D9.5). Anthropic tiene `cache_creation` con split 5m/1h; OpenAI no cobra por
  escribir cache y **no tiene** el concepto. Un `0` ahí sería mentira, y encima una mentira que
  se suma.
- **`null` sobrevive a la agregación.** El rollup horario usa una suma que preserva el `NULL`: si
  ningún evento del grupo tenía el bucket, el grupo queda `NULL`. Un `COALESCE(x,0)` de más en el
  rollup destruye la distinción justo donde nadie la mira.
- **Un detector que no aplica viaja al wire**, en `no_aplican[]`, **con motivo obligatorio**.
  Omitirlo obliga al FE a elegir entre no mostrar nada (gap escondido) o mostrar 0 (mentira).
- **Un costo que no se pudo calcular es `null`**, y si se calculó a medias es
  `completo:false` + `sin_tarifa:[…]`. Un `0` en dinero se lee como «gratis».
- **El FE tiene un solo lugar** donde se decide cómo se ve un `null`: `ValorOSinDato`. La regla no
  se re-implementa por componente.

**El motivo es obligatorio, no decorativo.** «No aplica» sin razón es un gap escondido con mejor
tipografía. Las tres instancias vivas ya lo cumplen (`deriva-no-evaluable` lleva motivo,
`no-comparable` lleva `sin-senal`+motivo, el `deferido` lleva la razón del mecanismo).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| nulo-no-es-vacio | «no pude leer» viaja como `null` y «leí y está vacío» como `[]`; nunca se normalizan al mismo valor | error | «un ilegible se muestra como vacío (pass fabricado)» | internal/usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio |
| no-comparable-primera-clase | el veredicto «no hay con qué comparar» es un estado propio del dominio, no un `disabled` de la UI | error | «un no-comparable disfrazado de al-día» | internal/domain/marketplace_situacion_test.go:TestSituacionNoComparableSinVersionEstante |
| no-evaluable-con-motivo | un veredicto que no se pudo emitir sale con motivo legible, jamás con el valor optimista | error | «veredicto fabricado por falta de dato» | internal/adapters/portafolio/deriva_test.go:TestEvaluarDerivaNoEvaluable |
| superset-usa-punteros | en un esquema superset multi-proveedor, todo campo que un proveedor puede no tener es puntero/opcional, nunca un entero con default 0 | error | «0 donde el concepto no existe» | (pendiente — TestNoAplicaNoEsCeroEnElWire) |
| no-aplica-sobrevive-al-agregado | la agregación preserva el «no aplica»; ningún `COALESCE(x,0)` lo colapsa en el camino | error | «el rollup convirtió no-aplica en cero» | (pendiente — TestNoAplicaSobreviveAlRollup) |
| no-aplica-viaja-al-wire | lo que no aplica se emite explícito con su motivo; no se omite dejando que el cliente lo interprete | error | «el FE tuvo que inventar el estado ausente» | (pendiente — TestDetectorQueNoAplicaTraeMotivo) |

## Changelog

- 2026-07-26 · v1.0 · Nodo fundacional (HS-28, paquete
  `stories/2026-07-24-telemetria-embebida-otel/`, decisión D16/D-4 FIRMADA). **No inventa
  doctrina: la consolida** — la regla ya estaba resuelta por separado en tres lugares
  (`entradas: null`≠`[]`, `no-comparable`, `deriva-no-evaluable`) y el módulo `telemetria/`
  necesitaba la cuarta y quinta (buckets de tokens como punteros, detectores que no aplican).
  L1 = lógica de tres valores + staleness markers; el aporte propio es la **tercera categoría**
  («el proveedor no tiene el concepto»), que ninguna plataforma de observabilidad LLM modela
  (investigación de runtimes, D-4). 6 checks: **3 nacen con enforcer REAL y verde** (corridos el
  2026-07-26 antes de escribir esta línea), 3 difieren honesto hasta que exista el módulo. Por eso
  el nodo nace `proposed` y no `enforced`: el criterio de la casa es que gradúa cuando corren
  **todos** sus checks (mismo criterio que `indice-desechable-jsonl-es-verdad` v1.3).
