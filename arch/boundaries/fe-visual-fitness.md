---
regla: fe-visual-fitness
version: 1.1
updated: 2026-07-06
status: enforced
ledger: HS-09
sources:
  - url: https://storybook.js.org/docs/writing-tests/integrations/vitest-addon
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://storybook.js.org/releases/10.0
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://vitest.dev/guide/browser/visual-regression-testing
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://storybook.js.org/addons/storybook-addon-vis
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - web/.storybook/main.ts
  - web/vitest.config.ts#storybook
  - ci.yml:visual-fitness
severity: medium
---

# Cada componente tiene una story que ES un test que rompe CI (fitness visual local)

## L1 · Principio (estándar de industria)

**Story-as-test (Storybook 10 + `@storybook/addon-vitest`).** Vía portable stories, cada story corre
como test de componente en **Vitest browser mode (Playwright Chromium, DOM real)**: smoke de render +
la `play` function con asserts. Una story rota → exit-code ≠ 0 → CI roja. Es el único workshop de
componentes con story-as-test nativo. *(oficial: Storybook 10 vitest-addon)*

**a11y como test.** `@storybook/addon-a11y` (axe) se integra al addon-vitest → las violaciones de
accesibilidad rompen CI. *(oficial: Storybook a11y)*

**Regresión visual local (sin SaaS).** Baselines de screenshot en el repo (`storybook-addon-vis` /
`@storybook/test-runner` + Playwright), cero nube — **Chromatic descartado** por chocar con local-first/
privado. *(oficial: Vitest visual regression; experto: storybook-addon-vis)*

## L2 · Realización (este árbol Go+React)

ArnesIA dogfoodea su propio estándar: su FE ES un producto UI pesado (canvas del Mapa en HTML+SVG con
overlay de edges, Dock, merge CM6; React Flow queda para el Organigrama), así que sus componentes se
aíslan en un workshop con checks ejecutables — coherente con la tesis de fitness functions de todo el
repo. ⇐ L1: story-as-test.

- **Storybook 10** (ESM-only; exige Node 20.16+/22.19+/24+, fijado en CI) en `web/.storybook/`;
  `web/vitest.config.ts` define, vía `test.projects` (Vitest 4 removió `defineWorkspace`/
  `vitest.workspace.ts`), el project `storybook` que transforma stories→tests.
- **3 checks en una corrida** (`pnpm test` → `vitest run` → project `storybook`): render+`play` · a11y axe ·
  coverage. **Regresión visual** por `storybook-addon-vis` (baselines en `web/`), 100% local. ⇐ L1: a11y,
  visual. Estado real: **40 story-tests en 11 archivos, verdes** (Playwright Chromium headless).
- **Determinismo Mint-first:** el visual corre en **Playwright Chromium dentro de contenedor Linux**
  (casar con dev Mint, evitar drift de font-AA cross-OS); animaciones y transiciones no-deterministas
  desactivadas en snapshots; el viewport del canvas se mide en tamaño fijo (los 7 shots = 1680×1000).
- **Gotchas del stack** (van en las stories, no en la arquitectura): el canvas del Mapa mide edges con
  `getBoundingClientRect` → necesita contenedor con w/h explícitos (jsdom no mide → browser mode
  obligatorio); Zustand se resetea/hidrata por story en decorator (singleton de módulo); SSE/assistant-ui
  = replay de fixtures grabados (patrón «AG-UI Dojo»), nunca daemon vivo; CodeMirror y React Flow (futuro
  Organigrama) necesitan altura de contenedor + su provider.
- **Umbral honesto:** el nodo es `severity: medium` — Storybook se paga cuando hay ~10-20 componentes
  reutilizables; por debajo, `vitest` browser pelado basta. La story-como-test es obligatoria; el
  catálogo visual es el upside.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| story-es-test | cada componente de `shared/ui` + widget tiene story; `vitest run` project `storybook` verde | error | «componente sin story-test (fitness visual ausente)» | vitest.config.ts#storybook |
| a11y-axe | ninguna story viola reglas axe (configurable a error) | error | banda Base «violación a11y en componente» | addon-a11y + vitest |
| regresion-visual-local | los snapshots casan con las baselines commiteadas (cero SaaS) | warn | «drift visual vs baseline» | storybook-addon-vis |
| sin-chromatic | no hay dependencia de un SaaS de regresión visual (local-first) | error | banda Guardia «UI enviada a SaaS externo» | revisión ci.yml |
| node-version-sb10 | el toolchain/CI fija Node ≥ 20.16/22.19/24 (SB10 ESM-only) | warn | «Node incompatible con Storybook 10» | ci.yml |

## Changelog

- 2026-07-06 · v1.1 · `proposed → enforced` (HS-09, Fase D). El MVP del Mapa es la primera superficie que
  ejercita story-as-test de verdad: **40 story-tests en 11 archivos, verdes** (`pnpm test`, Playwright
  Chromium headless). Migrado el enforcer de `vitest.workspace.ts` (Vitest 4 removió `defineWorkspace`) a
  `vitest.config.ts` (`test.projects`, project `storybook`). Corregido el ejemplo de UI pesada (canvas del
  Mapa HTML+SVG, no «nodos React Flow» — RF se reserva al Organigrama). Sin cambio de checks (5).
- 2026-07-05 · v1.0 · Nodo fundacional FE (HS-05). L1 = Storybook 10 story-as-test (addon-vitest) + a11y
  axe + regresión visual local. L2 = workshop en `web/.storybook`, 3 checks por corrida, determinismo en
  contenedor Linux, Chromatic descartado (local-first), gotchas RF/Zustand/SSE. `severity: medium`
  (umbral ~10-20 componentes). 5 checks.
