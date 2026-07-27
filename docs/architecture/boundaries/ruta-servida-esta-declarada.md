---
regla: ruta-servida-esta-declarada
version: 1.1
updated: 2026-07-26
status: enforced
ledger: HS-29
sources:
  - url: https://spec.openapis.org/oas/v3.1.1.html
    autoridad: oficial
    revisado: 2026-07-26
  - url: https://martinfowler.com/articles/consumerDrivenContracts.html
    autoridad: experto
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-26-conversaciones-del-panel/relevamiento-as-is.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by:
  # Los 4 deterministas existen y corren verdes desde 2026-07-26 (T5 del paquete
  # conversaciones-del-panel), y su validación NO fue leerlos: se inyectó a mano una ruta falsa
  # en el router y `TestRutaServidaEstaDeclarada` rompió; al quitarla volvió a verde.
  - docs/architecture/fitness/openapi_contract_test.go
  # ⚠ `forma-de-respuesta-unica` NO entra acá: sigue sin enforcer determinista y difiere
  # honesto en su celda (H-H). El nodo es `enforced` con **4 de 5**, y lo dice.
severity: error
---

# Una ruta que el daemon sirve está declarada en el contrato, o está exenta con razón escrita

## L1 · Principio (estándar de industria)

**El contrato de una API describe la superficie que el servicio efectivamente expone; si no la
describe, no es un contrato — es un folleto.**

- **El documento ES la descripción de la superficie.** OpenAPI no define un documento «orientativo»:
  define la descripción formal de la API y de su semántica, y el objeto `paths` es literalmente el
  inventario de los endpoints. *(oficial: OpenAPI 3.1.1, textual: «An OpenAPI Description (OAD)
  formally describes the surface of an API and its semantics» · Paths Object: «Holds the relative
  paths to the individual endpoints and their operations»)*
- **Un contrato se verifica contra el proveedor, no se publica y se confía.** La disciplina de
  consumer-driven contracts existe justamente porque un esquema publicado sin test se vuelve
  obsoleto o directamente ignorado: la garantía la da la corrida, no el archivo.
  *(experto: Ian Robinson, 2006-06-12 — Consumer-Driven Contracts, textual: «By implementing these
  tests, the provider gains a better understanding of how it can evolve the structure of the
  messages it produces without breaking existing functionality in the service community»)*

Corolario propio: **la deuda de contrato tiene que ser contable, no invisible.** La alternativa
realista a «todo declarado desde hoy» no es «nada gateado»: es una **exención declarada con razón**
que sólo puede achicarse. Un hueco listado con su motivo es deuda; un hueco que nadie enumeró es una
mentira estructural con el aspecto de documentación.

## L2 · Realización (este árbol Go+React)

El contrato vive en `docs/architecture/contracts/api/openapi.yaml` y las rutas se registran en
`internal/adapters/transport/http/router.go` con el mux nativo de Go 1.22+ (patrones
`"MÉTODO /ruta"`), por dos verbos: `mux.HandleFunc(` (la mayoría) y `mux.Handle(` (SSE, OTLP y el
SPA). **El enforcer tiene que escanear los dos** — mirar sólo `HandleFunc` deja `GET /api/events`
fuera y produce un falso «declarada sin servir».

- **Mecanismo:** source-scan del router + parseo del `paths:` del YAML con `gopkg.in/yaml.v3` (ya es
  dependencia directa, `go.mod` — cero peso nuevo, y encima es código de test). Mismo estilo
  heurístico-pero-honesto que el repo ya sanciona en `arch_test.go:TestNoJSONLSchemaParsing` y
  `marketplace_shape_test.go:TestDominioNoAdoptaShapeAjeno`.
- **`servers:` con base `/api`:** el YAML declara `/sessions`, el router registra `/api/sessions`.
  La normalización es del enforcer, no del que escribe el contrato.
- **Exención declarada:** `docs/architecture/contracts/api/_sin-declarar.yaml`, una entrada por ruta
  con `ruta:` + `razon:` obligatoria — misma forma y misma disciplina que la allowlist de
  [`codigo-traza-a-capability`](./codigo-traza-a-capability.md) (`_coverage.yaml` +
  `capability_trace_test.go:52-73`). Una entrada sin `razon` es un fail, no una omisión tolerada.
- **Ratchet:** el archivo de exención **sólo puede achicarse**. Una ruta nueva que no está en el
  contrato y no está exenta rompe el build el día que se escribe, que es el único día en que
  arreglarla es barato.

**El estado medido hoy** (generado, no tecleado — `2026-07-26`, HEAD `dd460f3`, contando los dos
verbos de registro):

| | |
|---|---|
| rutas que el router sirve | **54** (medido por AST 2026-07-26; la cifra anterior, 50, salía de un escaneo que no contaba `mux.Handle`) |
| rutas `/api` declaradas en el contrato | **39** |
| **servidas sin declarar** | **11** — las 7 de `telemetria/`, `GET /api/sessions/cerradas/{id}/historial`, `POST /api/portafolio/arneses/{clave}/identificar`, `GET /healthz`, `DELETE /api/telemetria/arneses/{clave}`. Las 11 (+ las 4 estructurales = **15**) viven hoy en [`_sin-declarar.yaml`](../contracts/api/_sin-declarar.yaml) con su razón |
| declaradas sin servir | **0** |
| exentas por estructura (fuera del contrato `/api` a propósito) | `POST /v1/logs` · `POST /v1/metrics` (receptor OTLP, contrato ajeno) · `GET /events` (alias sin prefijo) · `/` (SPA embebido) |

Esas 11 son la **deuda de arranque**: entran a `_sin-declarar.yaml` con su razón el día que nace el
enforcer, y salen a medida que cada módulo declare lo suyo. Lo que el nodo compra desde el minuto
uno es que **la número 12 no exista**.

- **La forma de la respuesta también es contrato.** `GET /api/sessions` devuelve un **array** sin
  `?cerradas=1` y un **objeto** con él (`sessions.go:32` vs `:41`), mientras el YAML declara
  `{type: array}` (`openapi.yaml:614`). Dos formas bajo una operación no son documentables: son dos
  operaciones. Este check no tiene enforcer determinista todavía y **difiere honesto**.
- **Sin codegen no hay gate.** El paso `openapi-gen-check` de CI está inerte por construcción
  (`ci.yml:73`, `if: hashFiles('web/src/shared/api/generated/**') != ''`, y ese directorio no
  existe): el guard es honesto sobre el directorio generado, pero **no guarda nada**. Este nodo es
  el gate que faltaba, y es independiente del codegen: no exige generar tipos, exige que lo servido
  esté declarado.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| ruta-servida-declarada | toda ruta registrada en el router (`mux.Handle` y `mux.HandleFunc`) existe en `openapi.yaml` con su método, o está en `_sin-declarar.yaml` | error | «el daemon sirve algo que el contrato no conoce» | docs/architecture/fitness/openapi_contract_test.go:TestRutaServidaEstaDeclarada |
| ruta-declarada-se-sirve | toda operación del `openapi.yaml` la sirve el router: cero contrato fantasma | error | «el contrato promete un endpoint que no existe» | docs/architecture/fitness/openapi_contract_test.go:TestRutaDeclaradaSeSirve |
| query-param-declarado | todo `r.URL.Query().Get("x")` de un handler tiene su `parameters:` en la operación correspondiente | warn | «un parámetro que cambia el comportamiento y nadie documentó» | docs/architecture/fitness/openapi_contract_test.go:TestQueryParamEstaDeclarado |
| exencion-con-razon | cada entrada de `_sin-declarar.yaml` lleva `razon:` no vacía, y el archivo sólo se achica | error | «deuda de contrato sin dueño ni motivo» | docs/architecture/fitness/openapi_contract_test.go:TestExencionDeContratoTieneRazon |
| forma-de-respuesta-unica | una operación no cambia la FORMA de su respuesta según un query param; dos formas son dos operaciones | error | «un cliente tipado contra el array recibe un objeto» | (pendiente — juicio; hoy sin enforcer determinista) |

## Changelog

- 2026-07-26 · v1.0 · Nodo fundacional (paquete
  `stories/2026-07-26-conversaciones-del-panel/`, etapa de arquitectura). Nace de un hallazgo
  medido, no de una intuición: `grep -c "cerradas" openapi.yaml` → **0**, con dos endpoints en
  producción hace días, y **nada en CI que lo cace** porque el único paso de contrato es condicional
  sobre un directorio generado que no existe. El paquete iba a agregar 4 rutas más por el mismo
  agujero. L1 = el documento describe la superficie (OpenAPI) + el contrato se verifica contra el
  proveedor (consumer-driven contracts); el aporte propio es la **exención contable con ratchet**,
  que es lo que hace la regla adoptable en un árbol que arranca con 11 rutas sin declarar en vez de
  cero. 5 checks: **4 con enforcer nombrado y ninguno escrito todavía**, 1 difiere honesto por falta
  de mecanismo determinista. **Nace `proposed`**; gradúa a `enforced` cuando los 4 corran verdes,
  mismo criterio que `indice-desechable-jsonl-es-verdad` v1.3. NUNCA pass fabricado.

- 2026-07-26 · v1.1 · **`proposed` → `enforced` con 4 de 5.** Los 4 checks deterministas tienen
  enforcer real (`docs/architecture/fitness/openapi_contract_test.go`) y corren verdes. La
  graduación **no se firmó leyendo el test**: se inyectó a mano `GET /api/ruta-falsa-de-prueba`
  en `router.go` y `TestRutaServidaEstaDeclarada` rompió nombrándola; al quitarla volvió a verde.
  Sin esa prueba el enforcer no estaba probado, sólo escrito.

  **Dos cosas que la construcción corrigió del propio nodo, y conviene que queden:**

  1. **La cifra de rutas servidas era 50 y son 54.** El conteo original salía de un escaneo que
     sólo miraba `mux.HandleFunc(`; el enforcer mira **los dos verbos** de registro, y `mux.Handle(`
     aportaba 4 (`/`, `GET /events`, `GET /api/events`, más el `ui` del SPA). El saldo de deuda
     **no cambia**: siguen siendo 11 `/api` sin declarar (15 entradas en la allowlist con las 4
     estructurales).
  2. **`query-param-declarado` nació vacuo y se rehízo.** La primera versión cruzaba «rutas
     registradas en el archivo X» con «queries leídas en el archivo X»; en este árbol `router.go`
     registra todo y no lee ni un query, y `sessions.go` lee dos y no registra nada. Intersección
     vacía, verde mentiroso. Se rehízo por **AST**, emparejando la ruta con el identificador del
     constructor de su handler y leyendo el cuerpo de esa función (closures incluidas). Al
     rehacerlo destapó al toque los dos que nadie había declarado: `?arnes=` y `?cerradas=` de
     `GET /api/sessions`, hoy ya declarados **tal como se comportan**.

  `forma-de-respuesta-unica` **sigue difiriendo honesto** (H-H): no tiene mecanismo determinista.
  El nodo es `enforced` con 4/5 y lo dice en su `enforced_by`.
