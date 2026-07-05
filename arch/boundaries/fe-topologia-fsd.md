---
regla: fe-topologia-fsd
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://feature-sliced.design/docs/reference/layers
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://feature-sliced.design/docs/reference/public-api
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://github.com/sverweij/dependency-cruiser/blob/main/doc/rules-reference.md
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://github.com/feature-sliced/steiger
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - web/.dependency-cruiser.js#layer-direction
  - web/.dependency-cruiser.js#shared-no-upward
  - web/.dependency-cruiser.js#no-sibling-feature-imports
  - web/.dependency-cruiser.js#no-deep-import
  - web/steiger.config.ts
severity: critical
---

# La SPA se estructura por Feature-Sliced Design (import direccional)

## L1 · Principio (estándar de industria)

**Feature-Sliced Design 2.x.** Una arquitectura FE en capas (`app › pages › widgets › features ›
entities › shared`) con **una regla de import formal: una slice solo importa capas estrictamente
inferiores**. Es la única topología React mainstream 2026 con un invariante de import concreto que un
linter rompe en CI — a diferencia de atomic-only o feature-folders planas (que organizan por tipo o
por nombre, sin dirección enforçable) o bulletproof-react (boundary global↔feature bidireccional, no
enforçable). *(estándar: FSD layers reference)*

**Public API por slice.** Cada slice expone un `index.ts` (barrel) como su superficie; el resto del
código importa la slice por su public API, nunca por deep-import a sus internos → el interior es
refactorable sin romper consumidores. Excepción por bundle: `shared/ui` y `shared/lib` exponen public
API **por componente**, no un mega-barrel. *(estándar: FSD public API)*

**Grafo de imports as-code.** `dependency-cruiser` valida el grafo *resuelto* real (incl. dynamic y
type-only imports), rompe CI con exit-code y emite DOT/SVG diffeable — es el análogo TS de
`go-arch-lint`. *(oficial: dependency-cruiser)*

## L2 · Realización (este árbol Go+React)

La SPA vive en **`web/src/**`** (Vite root; `web/dist` es lo que `go:embed` sirve). Las 6 capas FSD son
carpetas load-bearing bajo `web/src/`, adaptadas a una **app de escritorio de lienzo único sin router**:

- **`pages` = composition-roots seleccionados por hash-state**, NO rutas de URL. ⇐ L1: layers. Los 2-3
  roots (`pages/workspace`, `pages/portfolio`) se intercambian por estado Zustand + deep-link
  `arnesia://` (Tauri), no por react-router. **Divergencia declarada:** la doc FSD asume routing; aquí
  `pages` = los pocos roots que ensamblan widgets, seleccionados por estado. El nombre `pages` se
  mantiene (steiger/presets lo esperan).
- **`widgets` = bloques grandes autosuficientes** (la definición oficial contempla «independent blocks
  within a single page»): `widgets/command-rail`, `widgets/map-canvas`, `widgets/dock`,
  `widgets/org-chart`, `widgets/portfolio-grid`.
- **`entities`** = dominio (`entities/{arnes,hook,knowledge-node,company,run}`); **`features`** =
  interacciones (`features/{edit-inline,create-arnes,command-k,toggle-layer,run-exec}`); **`shared`** =
  kit (`shared/{ui,api,lib,config}`).
- **Cross-import en `entities`** (crítico para el Organigrama «reporta a»): vía la Public API `@x` de FSD
  (`entities/company/@x/arnes.ts`), no import lateral directo. ⇐ L1: public API.
- **Barrels:** por slice OK; `shared/ui|lib` per-componente (bundle `go:embed`); **solo re-export puro**
  (bug Vite 8 #21966: inline-export + re-export en el mismo archivo rompe tree-shaking).

Enforcer único de grafo = `web/.dependency-cruiser.js` (CI-gate, espeja go-arch-lint); `steiger` cubre
lo FSD-estructural que el cruiser no da (public-api, naming de slice) — **secundario** (beta 0.5.0, no
extensible → nunca el único gate). `eslint-plugin-boundaries` **descartado** (duplicaría el grafo →
drift de dos verdades). ⇐ L1: grafo as-code.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| fsd-layer-direction | ninguna capa importa una capa superior (`entities`↛`features/widgets/pages/app`, etc.) | error | «import contra-corriente FSD (acopla la capa a su consumidor)» | dependency-cruiser#layer-direction |
| fsd-shared-no-upward | `shared/**` no importa `entities/features/widgets/pages/app` | error | «shared (la base) importa hacia arriba» | dependency-cruiser#shared-no-upward |
| fsd-no-sibling-feature | una feature no importa otra feature (cross-slice prohibido) | error | «feature acopla a feature hermana» | dependency-cruiser#no-sibling-feature-imports |
| fsd-public-api | se importa la slice por su `index.ts`, no por deep-import a internos | warn | «deep-import salta la public API de la slice» | dependency-cruiser#no-deep-import · steiger public-api |
| fsd-barrel-puro | los `index.ts` son re-export puro (no inline-export + re-export) | warn | banda Base «barrel mixto rompe tree-shaking (Vite #21966)» | steiger · revisión |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional FE (HS-05). L1 = FSD 2.x + public API + dependency-cruiser. L2
  fija `web/src/**`, reinterpreta `pages`=composition-roots-por-hash-state (divergencia declarada:
  app de escritorio sin router), cross-import `@x`, barrels re-export-puro. Enforcer único =
  dependency-cruiser (eslint-plugin-boundaries descartado por drift). 5 checks.
