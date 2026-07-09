---
regla: fe-transporte-independiente
version: 1.1
updated: 2026-07-09
status: proposed
ledger: HS-05
sources:
  - url: https://feature-sliced.design/blog/zustand-simple-state-guide
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://github.com/pmndrs/zustand/discussions/2496
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://alistair.cockburn.us/hexagonal-architecture/
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - web/.dependency-cruiser.js#domain-not-transport
  - web/.dependency-cruiser.js#sse-singleton
severity: high
---

# El dominio FE no depende del transporte; SSE es singleton en `app`

## L1 · Principio (estándar de industria)

**Hexagonal en el frontend.** El mismo principio del backend ([`dominio-independiente-de-transporte`](./dominio-independiente-de-transporte.md))
aplica en la SPA: el modelo de dominio no conoce cómo llegan los datos (fetch, SSE). El transporte es un
adaptador; el dominio consume datos ya normalizados. *(experto: Cockburn, hexagonal)*

**Store-per-slice, no store global.** Zustand con un store por slice casa con los boundaries del módulo
y evita el «dumping ground donde cualquier feature muta cualquier cosa». *(estándar: FSD+Zustand;
oficial: zustand slices-vs-stores)*

**Una conexión SSE, bootstrapeada una vez.** El transporte de streaming es un recurso singleton; abrirlo
desde una feature multiplica conexiones (el cap de 6/origen HTTP/1.1 es real en localhost). *(deriva de
la decisión de transporte HS-04: 1 conexión multiplexada `event: map|dock|run`.)*

## L2 · Realización (este árbol Go+React)

- **`shared/api` = transporte puro:** cliente fetch + **tipos generados del OpenAPI** (`shared/api/
  generated/`, del contrato `arch/contracts/api/openapi.yaml`) + la **conexión SSE única (singleton)**.
  No importa `entities/features/widgets/pages`. ⇐ L1: hexagonal.
- **El estado vive por capa** (store-per-slice): dominio en `entities/*/model/*.store.ts` · interacción
  en `features/*/model/*.store.ts` · **estado de vista del canvas del Mapa (HTML+SVG)**
  (nodes/edges/viewport/selección) en `widgets/map-canvas/model/` — separado del dominio; los nodos del
  Mapa se hidratan desde selectores de las entity-stores, **no se duplica la verdad** (React Flow queda
  reservado al Organigrama, HS-09). ⇐ L1: store-per-slice.
- **El reducer SSE→store vive en `app/realtime/`:** bootstrapea la conexión una sola vez, traduce los
  eventos tipados `map|dock|run` a acciones sobre los stores de dominio. El dominio nunca sabe que hay
  SSE; consume datos normalizados (espejo del boundary backend dominio⊥transporte). ⇐ L1: SSE singleton.
- **Regla de oro:** `entities/*/model` (dominio) jamás importa `shared/api` para *escuchar*; el flujo
  SSE→dominio es unidireccional vía el reducer en `app`. Para *pedir*, la slice usa su segmento `api/`
  sobre `shared/api`. Solo `app/realtime` toca la conexión SSE cruda.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| domain-not-transport | `entities/*/model/**` no importa `shared/api` (transporte) | error | «el dominio FE escucha el transporte directo» | dependency-cruiser#domain-not-transport |
| sse-singleton | solo `app/realtime/**` importa la conexión SSE cruda de `shared/api`; ninguna feature | error | banda Guardia «feature abre SSE → N conexiones» | dependency-cruiser#sse-singleton |
| transport-puro | `shared/api/**` no importa `entities/features/widgets/pages/app` | error | «transporte acoplado al dominio/UI» | dependency-cruiser#shared-no-upward |
| tipos-generados | los tipos de `shared/api/generated/**` derivan del OpenAPI, no escritos a mano | warn | «tipo de API a mano (drift con el contrato)» | ci.yml:openapi-gen-check |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional FE (HS-05). L1 = hexagonal en el FE + store-per-slice + SSE
  singleton. L2 = `shared/api` transporte puro, reducer SSE→store en `app/realtime`, estado por capa,
  tipos generados del OpenAPI. Espejo FE del boundary backend dominio⊥transporte. 4 checks.
- 2026-07-09 · v1.1 · **Sync HS-18.** El canvas del Mapa es **HTML+SVG** (no React Flow, que queda
  para el Organigrama, HS-09); se relabela «estado de vista de React Flow» → «estado de vista del
  canvas del Mapa». Sin cambios de checks ni de status.
