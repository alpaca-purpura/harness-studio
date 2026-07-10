# Investigación FE — arquitectura, atomic, storybook, convenciones (HS-05 prep, 2026-07-05)

> **5 frentes en paralelo (subagentes), verificados contra la web vigente (julio 2026).**
> Es la **evidencia L1** que respaldará los boundary/convention nodes del frontend en
> [`../arch/`](../../architecture/INDEX.md) (gemelo FE de los 7 boundaries backend de HS-04). Condensado:
> lo decisivo para la propuesta. Prioridad de fuentes = docs oficiales › estándares/proyectos
> maduros › expertos. Norte: [`../vision.md`](../vision.md) · contexto: **app instalable de
> escritorio** (Tauri 2 · daemon sidecar · SPA `go:embed` · local-first · Linux/Mint-first ·
> equipo 1-3 · sesgo low-maintenance). Complementa
> [`2026-07-05-arquitectura-fase3.md`](./2026-07-05-arquitectura-fase3.md) (backend/stack HS-04).

## Reconciliaciones cruzadas (resueltas antes de cementar)

Cuatro tensiones aparecieron entre frentes; resueltas aquí para que los nodos sean coherentes:

1. **FSD (topología) ≠ atomic (taxonomía).** Frente A pide FSD-lite (carpetas/capas/dirección de
   import); frente B pide una taxonomía de niveles de componente. **No compiten:** FSD es la
   topología del módulo; atomic vive DENTRO de `shared/ui` como vocabulario de nombres. Se adoptan
   ambos en su rol. *(B lo dijo textual: «atomic es taxonomía, FSD es topología».)*
2. **Un solo enforcer de grafo, no dos.** A recomienda `dependency-cruiser` como CI-gate y
   **descarta** `eslint-plugin-boundaries` (que B había propuesto) para evitar drift de dos grafos.
   **Resuelto: dependency-cruiser = gate** (declarativo standalone, espeja go-arch-lint, emite
   DOT/SVG para `docs/architecture/model/`); `steiger` = check FSD-estructural secundario (beta, no único gate);
   `eslint-plugin-boundaries` descartado.
3. **Convención de vars: shadcn vs mockup.** shadcn/assistant-ui esperan `--background/--foreground/
   --border/--ring`; los mockups firmados usan `--bg/--surface/--line`. **Resuelto: la capa
   semántica canónica adopta la convención shadcn** (renombrar, no shim), para no mantener dos
   vocabularios; el `.tokens.json` DTCG emite esos nombres y los mockups se leen como fuente de
   VALORES, no de nombres.
4. **Biome no reemplaza al grafo.** Biome (estilo/format/naming/hooks) y dependency-cruiser
   (arquitectura de imports) son concerns distintos; conviven. El stack FE de fitness = Biome +
   dependency-cruiser + stylelint + `tsc --noEmit` + Storybook/Vitest (cada uno un check distinto).

## Frente A — Arquitectura de módulos/carpetas del frontend

**Veredicto:** **Feature-Sliced Design 2.x adaptado («FSD-lite»)** como topología; **`dependency-
cruiser` como CI-gate** (espejo 1:1 de go-arch-lint) + **`steiger`** (preset `recommended`) como
check FSD-estructural secundario. Descartados: atomic-only y feature-folders planas (sin dirección
de import enforçable = nada que rompa CI útil), bulletproof-react (boundary global↔feature
bidireccional, no enforçable), `eslint-plugin-boundaries` (duplica el grafo → drift). El riesgo de
over-engineering para 1-3 se neutraliza con FSD-lite: las 6 capas existen como nombres load-bearing
pero las slices pueden estar vacías/ausentes («no tienes que usar todas las capas» — doc oficial).

**Reinterpretación clave (app de escritorio, lienzo único, SIN router — la doc FSD no lo cubre):**
- **`pages` = composition-roots seleccionados por hash-state**, NO rutas de URL. Los 2-3 roots
  (`workspace`/`map`/`portfolio`) se intercambian por estado Zustand + deep-link `arnesia://` (Tauri),
  no por react-router. Mantener el nombre `pages` (los presets de steiger/FSD lo esperan).
- **`widgets` = bloques grandes autosuficientes** — la definición oficial contempla «independent
  blocks within a single page», que es exacto para un lienzo único.

**Mapeo Command Rail A → capas:**

| Elemento UX | Capa | Slice |
|---|---|---|
| shell, providers, bootstrap SSE, theme, hash-state shim | `app` | `providers/ · store/ · styles/ · realtime/` |
| Command Rail (rail iconos) | `widgets` | `widgets/command-rail` |
| Mapa (canvas React Flow, `nodeTypes`, layout carriles) | `widgets` | `widgets/map-canvas` |
| Organigrama (2D libre) · Cuadrícula | `widgets` | `widgets/org-chart` · `widgets/portfolio-grid` |
| Dock (chat assistant-ui/AG-UI) | `widgets` | `widgets/dock` |
| pantallas que ensamblan Rail+visual+Dock | `pages` | `pages/workspace · pages/portfolio` |
| arnés · hook · nodo-conocimiento · company · puesto · run (datos) | `entities` | `entities/*` |
| editar-inline · crear-arnés · ⌘K · toggle-capa · ejecutar · filtrar | `features` | `features/*` |
| UI kit · cliente API+tipos OpenAPI · SSE · hash-lib · store-creator | `shared` | `shared/ui · shared/api · shared/lib · shared/config` |

**Estado (Zustand store-per-slice, NO global):** dominio en `entities/*/model/*.store.ts` ·
interacción en `features/*/model/*.store.ts` · **estado de vista de React Flow** (nodes/edges/
viewport/selección) en `widgets/map-canvas/model/` (separado del dominio; hidratar RF-nodes desde
selectores de entity-stores, no duplicar la verdad). Boundary transporte⊥dominio (espejo backend):
`shared/api` = transporte puro (fetch + tipos OpenAPI generados + **la conexión SSE única, singleton**);
el **reducer SSE→store vive en `app/realtime/`** (bootstrap una sola vez; si lo abre una feature →
N conexiones); el dominio consume datos ya normalizados, nunca escucha SSE directo.

**React Flow custom nodes = 3 piezas across capas** (no una capa nueva): el dato = `entity`; el
componente que lo pinta = segmento `ui` de esa entity (`entities/arnes/ui/ArnesNode.tsx`); el canvas
(`<ReactFlow>` + `ReactFlowProvider` + registro `nodeTypes`/`edgeTypes` + layout carriles) = `widget`
(`widgets/map-canvas`). `nodeTypes` DEBE ser objeto estable fuera del render (`widgets/map-canvas/
config/nodeTypes.ts`).

**Barrels:** `index.ts` por slice = OK (cohesivo, refactorable; steiger `public-api` lo enforça) ·
**mega-barrel en `shared/ui|lib` PROHIBIDO** (per-componente por bundle) · **solo re-export puro**
(bug Vite 8 #21966: mezclar inline-export + re-export rompe tree-shaking) · dependency-cruiser prohíbe
deep-imports que salten el public API.

**Cross-import en `entities` (crítico para el Organigrama «reporta a»):** usar la Public API `@x` de
FSD (`entities/company/@x/arnes.ts`) — sin esto, la relación arnés→arnés/company fuerza a violar la
regla de capa. steiger y dependency-cruiser la respetan.

Regla de grafo as-code (dependency-cruiser, rompe CI):
```js
// .dependency-cruiser.js — espeja go-arch-lint (declarativo, exit≠0 en violación)
forbidden: [
  { name:'no-sibling-feature-imports', severity:'error',
    from:{ path:'^src/features/([^/]+)/.+' },
    to:{ path:'^src/features/([^/]+)/.+', pathNot:'^src/features/$1/.+' } },   // $1 = misma feature
  { name:'shared-no-upward', severity:'error',
    from:{ path:'^src/shared/' }, to:{ path:'^src/(entities|features|widgets|pages|app)/' } },
  { name:'layer-direction', severity:'error',
    from:{ path:'^src/entities/' }, to:{ path:'^src/(features|widgets|pages|app)/' } },
]
```

Fuentes: feature-sliced.design/docs/reference/{layers,public-api} · /docs/get-started/overview ·
/blog/zustand-simple-state-guide · /blog/mastering-eslint-config · github.com/feature-sliced/steiger ·
github.com/sverweij/dependency-cruiser (rules-reference: from/to/pathNot/$1/orphans) ·
xebia.com (dependency-cruiser FE architecture) · vite.dev/guide/performance (avoid barrel files) ·
github.com/vitejs/vite/issues/21966 (barrel tree-shaking) · github.com/pmndrs/zustand/discussions/2496
(slices vs stores) · reactflow.dev/learn/customization/custom-nodes.

## Frente B — Atomic design y taxonomía de componentes

**Veredicto:** **NO atomic clásico de 5 tiers como estructura de carpetas.** Consenso maduro 2026
(incl. el propio Brad Frost): la jerarquía de 5 es buen *vocabulario mental* pero mala *taxonomía de
carpetas* — la frontera atom/molecule/organism es difusa y esa ambigüedad ES el modo de falla. Lo
enforçable no es «de qué está hecho» sino «quién importa a quién». **Taxonomía plana + semántica +
direccional, estilo FSD-lite**, con atoms/molecules ⊂ `shared/ui`.

**Taxonomía concreta ArnesIA (6 capas de UI, cada una un boundary node con `enforced_by:`):**

| Capa | Path | Ejemplos | Prohibido importar |
|---|---|---|---|
| **primitives** (átomos) | `shared/ui/primitives/` | Button, Chip, Tooltip, Input, Icon, Popover | canvas, chrome, features, store |
| **compositions** (moléculas) | `shared/ui/` | DockRow, Toolbar, Menu, Card, SearchField | canvas, store, dominio |
| **canvas** (rama hermana) | `shared/canvas/`, `entities/*/ui`, `widgets/map-canvas` | NodeShell, Port, Handle, BandLane, ProcessBox, bandas | **chrome**, views |
| **entities/features ui** (organismos de dominio) | `entities/*/ui`, `features/*/ui` | ArnesCard, fila de timeline del Dock | canvas internals, otra feature |
| **chrome/shells** (organismos de app + layout) | `widgets/*`, `app/shell` | Command Rail, Dock, Organigrama, Grid, Dialogs | **canvas internals** (solo props/store) |
| **views/pages** | `pages/*` | Vista Mapa, Vista Portafolio | ser importado por nadie |

Mnemónica: **primitivos renderizan · canvas dibuja el grafo · chrome orquesta el shell · views montan.**
El insight load-bearing: un custom node **no es un átomo reutilizable** (vive dentro del
`ReactFlowProvider`, con handles/ports/selección gobernados por el store del canvas) → el canvas es
**rama hermana** de `shared/ui`, no un sub-nivel del ladder atómico. La frontera de mayor valor a
enforçar = **canvas ⊥ chrome** (runtimes distintos), no «atom no importa organism» (teatro cuando los
tiers son difusos).

**shadcn/ui = base de primitivos (SÍ) — sobre Base UI, no Radix legacy.** Razones dado escritorio +
go:embed: (1) copy-in = «own every file» → componentes lintables/enforçables en el repo (espeja
`arch/`), no opacos en `node_modules`; (2) bundle ~35-50 KB vs Mantine 80-120 / MUI 120-180 (importa
bajo go:embed); (3) sin `ThemeProvider` lock-in → tokens propios vía CSS-vars; (4) Vite de fábrica,
sin Next. Capa headless = **Base UI** (mantenido por MUI, activo, default greenfield 2026; Radix bajó
velocidad tras adquisición WorkOS). Rechazados MUI/Mantine/Chakra (provider lock-in, bundle 2-3×). Ojo:
**React Flow UI ya distribuye sus nodos por el CLI/registry de shadcn** → nodos como copy-in en el
mismo registro (pero asumen Tailwind + tokens shadcn → alinea con el frente E).

**Enforcement:** SÍ, pero **dirección y pureza, no el ladder de 5.** Las 3 fronteras que pagan el
lint: canvas⊥chrome · ui⊥dominio/store · no cross-slice. Vía el mismo `dependency-cruiser` del frente
A (una sola fuente de grafo). Los primitivos consumen **solo tokens SEMÁNTICOS** (`var(--…)`), nunca
hex ni tokens crudos → re-tematizar = reasignar semánticos, cero cambios de componente (enforçable con
stylelint, frente E).

Fuentes: feature-sliced.design/blog/atomic-design-architecture · bradfrost.com/blog/post/extending-
atomic-design · designsystemscollective.com (atomic 2025) · mentorcruise.com (atomic fails enterprise)
· qt.io (labels don't matter) · greatfrontend.com/blog/top-headless-ui-libraries-for-react-in-2026
(Radix velocity) · certificates.dev (shadcn/Radix/Base UI) · ui.shadcn.com/docs/changelog (Vite, Base
UI) · shadcndeck.com (bundle sizes) · xyflow.com/blog/react-flow-components (nodos por registry).
**No verificado:** «+70% adopción headless» y rangos KB = orden de magnitud (posts comparativos, no
benchmark oficial); fecha del switch shadcn→Base UI (dic-2025 según changelog, confirmar al citar).

## Frente C — Storybook como fitness visual

**Veredicto:** **SÍ Storybook — v10** (SB 10.0 oct-2025, latest 10.4.x; ESM-only) + **`@storybook/
addon-vitest`** → **una story ES un test que rompe CI** (portable stories corriendo en Vitest browser
mode con Playwright Chromium, DOM real). Es la única opción con story-as-test nativo = calza la tesis
dogfood de ArnesIA (fitness functions ejecutables). **Regresión visual 100% local; Chromatic
descartado** (SaaS que sube la UI a la nube → choca frontal con local-first/privado). Descartados
Ladle (sin play/story-as-test) e Histoire (Vue-first, beta perpetuo, muerto para React). Fallback más
ligero = vitest 4 browser pelado (sin catálogo/controls/design-review) — no elegido porque el producto
ES UI-canvas-heavy y autorreferencial.

**Cómo se vuelve check as-code** (boundary node `visual-fitness`, `enforced_by:` = comandos):
1. `vitest --project=storybook` → smoke de render + `play` (interaction asserts). Rompe CI.
2. **a11y axe** (`@storybook/addon-a11y` integrado al addon-vitest) → violaciones rompen CI. Gratis.
3. **regresión visual local** (baselines en el repo, cero SaaS): `storybook-addon-vis` (usa
   `vitest-plugin-vis`) — el más alineado; alternativas `@storycap-testrun/browser` o
   `@storybook/test-runner` + Playwright `toHaveScreenshot()`. Ojo: **`addon-vitest` NO cablea
   `toMatchScreenshot`** de Vitest 4 → la regresión de *stories* va por uno de esos 3.
4. **Determinismo Mint-first:** correr el visual en Playwright Chromium **dentro de contenedor Linux**
   (casar con dev Mint, evitar drift de font-AA cross-OS); desactivar animaciones y `fitView`
   no-determinista.

**Break-even honesto:** ~10-20 componentes reutilizables reales, o cuando el catálogo/design-review
tiene valor propio. ArnesIA cruza el umbral por (a) producto UI pesado e iterativo (nodos RF, Dock,
merge CM6) y (b) coherencia dogfood (story=test-que-rompe-CI). En un CRUD normal ganaría vitest pelado.

**Gotchas del stack:** React Flow necesita **contenedor con w/h explícitos + browser mode** (jsdom no
mide → canvas en blanco); decorator `<div style={{width:800,height:600}}><ReactFlowProvider>…`. Zustand
= singleton de módulo → **resetear/hidratar por story** en decorator (`setState(initial,true)`); mock
global con `sb.mock` (SB10, solo en `.storybook/preview`). SSE = nunca pegar al daemon vivo → sembrar
el store o replay de eventos grabados. assistant-ui/AG-UI Dock = **runtime AG-UI mock por replay de
fixture** (patrón «AG-UI Dojo»). CodeMirror 6/merge = browser mode + altura de contenedor. Entorno: SB10
exige **Node 20.16+ / 22.19+ / 24+** (fijar en CI). Tema: `@storybook/addon-themes`
(`withThemeByDataAttribute`) togglea el dark del app en el preview.

Fuentes: storybook.js.org/releases/10.0 · /blog/storybook-10 · /docs/writing-tests/integrations/
vitest-addon · /docs/api/portable-stories/portable-stories-vitest · vitest.dev/guide/browser/visual-
regression-testing · infoq.com (Vitest 4 browser mode) · github.com/storybookjs/storybook/discussions/
32930 (addon-vitest ≠ toMatchScreenshot) · storybook.js.org/addons/storybook-addon-vis · delta-qa.com +
itnext.io (visual sin Chromatic) · reactflow.dev/api-reference/react-flow-provider · learn.microsoft.com
(AG-UI Dojo testing) · github.com/pmndrs/zustand/discussions/2626 · ladle.dev + github.com/tajo/ladle/
issues/284 (sin play) · github.com/histoire-dev/histoire/releases (beta ene-2026). **No verificado:**
compat exacta `storybook-addon-vis` con SB10.4+Vitest4 (spike); determinismo real de snapshots de canvas
React Flow (posible flakiness → fijar dims medidas).

## Frente D — Convenciones de código as code (Go + TS/React + repo)

**Veredicto por capa:**

| Capa | Decisión | Config | Rompe CI |
|---|---|---|---|
| Go lint | **golangci-lint v2** (`version: "2"`), ~30 linters curados | `.golangci.yml` | `golangci-lint run` |
| Go format | **gofumpt + goimports** (sección `formatters:`, no linters) | `.golangci.yml` | `golangci-lint fmt --diff` |
| TS/React | **Biome v2.4** (lint+format+organize-imports+**react-hooks stable**) + `tsc --noEmit` | `biome.json` | `biome ci .` / `tsc --noEmit` |
| tsconfig | extiende **`@tsconfig/strictest`** + overrides Vite | `tsconfig.json` | `tsc --noEmit` |
| Pre-commit | **lefthook** (binario Go, políglota, paralelo — no husky) | `lefthook.yml` | hook exit≠0 |
| Commits | regex en lefthook `commit-msg` (Node-free) o commitlint | `lefthook.yml` | `commit-msg` |
| Rust shell | rustfmt `--check` + clippy `-D warnings` | (defaults) | `cargo clippy` |
| Repo | EditorConfig | `.editorconfig` | (vía format checks) |

**Go (golangci-lint v2, cambios estructurales vigentes):** `version: "2"` obligatorio · `staticcheck`
**absorbe gosimple+stylecheck** (no listarlos) · formatters (`gofumpt/goimports`) en sección aparte
(`golangci-lint fmt`) · default `standard` = solo 5 linters → ampliar. Set de alto valor para un
daemon: correctitud (`errcheck, errorlint, nilerr, nilnesserr, bodyclose, noctx, rowserrcheck,
sqlclosecheck, contextcheck, fatcontext, makezero`) · seguridad (`gosec`, `depguard` = ya en uso pa'
boundaries hexagonales) · **`sloglint`** (ya usan slog) · modernización (`copyloopvar, intrange,
perfsprint, usestdlibvars, usetesting, modernize, unconvert, wastedassign`) · estilo (`revive`
[reemplaza golint], `gocritic, unparam, predeclared, nolintlint`) · tests (`testifylint, tparallel`).
**Naming Go = enforced con `revive`** (`var-naming, exported, receiver-naming, error-naming,
error-strings, context-as-argument`), no inventar convención. **Evitar por ruido:** `lll, wsl, nlreturn,
err113, exhaustruct, varnamelen, mnd, godox, gochecknoglobals, dupl, tagliatelle`. `go vet`+`staticcheck`
corren DENTRO de golangci-lint (no invocar aparte).

**TS/React — el dato que cambia el debate 2026:** **Biome v2.4 promovió `useExhaustiveDependencies` +
`useHookAtTopLevel` a STABLE** → el histórico «quédate con ESLint porque react-hooks es no-negociable»
ya no aplica. **Biome v2.4 primario** (un binario Rust: format ~97% Prettier-compat + 200+ reglas +
organize-imports nativo + type-aware ~75%) **+ `tsc --noEmit` con `@tsconfig/strictest` como compuerta
de tipos** (cubre el gap type-aware: unsafe access, undefined, optional). Menor mantenimiento, un
config, ~20× más rápido, **cero Node solo-para-lintear** (alinea con daemon Go + lefthook Go). **Criterio
de auditoría para saltar a ESLint:** si `no-floating-promises`/`no-misused-promises`/`no-unsafe-*` al
100% son red de seguridad dura load-bearing → ESLint 9 flat + typescript-eslint `strict-type-checked` +
Prettier (costo: +2 herramientas, Node, CI más lento). oxlint descartado (formatter `oxfmt` inmaduro
2026). tsconfig: `@tsconfig/strictest` (`noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, etc.)
+ overrides Vite (`moduleResolution:"Bundler"`, `verbatimModuleSyntax`, `jsx:"react-jsx"`, `noEmit`).
Naming TS = `useNamingConvention` de Biome (config, no doc).

**Repo:** **lefthook** (binario Go, políglota, paralelo, `{staged_files}` integrado — gana sobre
husky+lint-staged que exige Node) corre golangci-lint + biome + (Rust) por glob. **Commits:** ya usan
`feat: HS-NN — …`; validar con regex en `commit-msg` (Node-free) o commitlint (reintroduce Node, solo
si quieren el ecosistema). **CI GitHub Actions** = jobs paralelos como required status checks:
`golangci/golangci-lint-action` (v7 para v2) · `biome ci .` + `tsc --noEmit` · `go test ./... -race` +
`go build` · `cargo clippy -D warnings` + `cargo fmt --check`. lefthook = feedback rápido (saltable con
`--no-verify`); **CI = enforcement duro**.

**Disposición as-code (propuesta):** **`docs/architecture/conventions/`** como hermano de `docs/architecture/fitness/` —
`fitness/` = forma arquitectónica (go-arch-lint, boundaries), `conventions/` = estilo/código (lint,
format, naming, commits). Nodos: `go-style.md · ts-style.md · ts-types.md · naming.md · commits.md ·
hooks.md · editor.md · ci.md`, cada uno L1 principio+fuente ↔ L2 realización + **`enforced_by:` → el
config REAL en la raíz** (`.golangci.yml`, `biome.json`, `tsconfig.json`, `lefthook.yml`,
`.editorconfig`, `.github/workflows/ci.yml`). El runner futuro **`arnesia conformance`** corre
`docs/architecture/knowledge/` + `arch/` + `docs/architecture/conventions/` (3 árboles). `docs/architecture/conventions/CADENCE.md` = «se revisa al
cambiar convención o subir versión de tooling».

Fuentes: golangci-lint.run/docs/{configuration,product/migration-guide} · ldez.github.io (v2 announce) ·
gist maratori (golden config) · github.com/mgechev/revive · pkgpulse.com (Biome vs ESLint vs oxlint 2026)
· biomejs.dev/linter/rules/{use-exhaustive-dependencies,use-hook-at-top-level} · typescript-eslint.io/
users/configs · npmjs.com/package/@tsconfig/strictest · pkgpulse.com + andymadge.com (lefthook vs husky
2026) · lefthook.dev · editorconfig.org · github.com/golangci/golangci-lint-action. **No verificado:**
«787 reglas oxlint» y «~75% floating-promise coverage Biome» = orden de magnitud (guías secundarias).

## Frente E — Design tokens + sistema de styling

**Veredicto (3 decisiones para go:embed + canvas + escritorio):**

| Eje | Decisión | Por qué |
|---|---|---|
| Formato tokens | **DTCG 2025.10 `.tokens.json` = SSOT** (adoptar, no prematuro) | Spec alcanzó **primera versión estable 28-oct-2025**; media type `application/design-tokens+json`. Contrato versionable que `arch/` exige. |
| Pipeline | **DTCG → Style Dictionary v5 → Tailwind `@theme` CSS + TS types** (un build-step) | SD v5.5 consume DTCG 2025.10 nativo; corre en build, no viaja en go:embed. Fallback cero-tooling: hand-authorear `@theme` (los mockups ya son CSS vars) y dropear DTCG (pierdes portabilidad + TS types). |
| Styling | **Tailwind v4** (`@tailwindcss/vite`), no vanilla-extract/CSS Modules/CSS-in-JS | `@theme` CSS-first = tokens SON CSS vars · Oxide/Rust → bundle **5-20 KB** (crítico go:embed) · zero-runtime · **React Flow 12 ya migró a CSS-vars** → nodos consumen los mismos tokens · coherencia con shadcn/Base UI/assistant-ui (Tailwind-native) · dark mode `data-theme` trivial. |

**Tokens = contrato mockup↔código (verificado: los hex del `.tokens.json` calzan exacto con
`mockups/arnesia-shell-A-galaxia.html`).** Modelo 3-capas: primitive (raw) → **semantic (intent, lo que
consumen los componentes)** → component. Set semántico base extraído (dual-theme claro/oscuro):
`--bg --surface --surface2 --line --line2 --text --muted --accent --accent-soft`. Dominio — **tipo de
componente** (mapea a `docs/architecture/knowledge/`): `--c-skill #2f74b5 · --c-agent #7455c2 · --c-hook #bb4c85 ·
--c-know #20826f · --c-mcp #586c80 · --c-rule #7a7a24`. Dominio — **salud de trazas**: `--ok #2e8f5e ·
--warn #c96a2e · --crit #c94545` + `-soft`. Dominio — **desempeño/heatmap**: v3 define escala
`--heat1..4`, shell-A la colapsa a `--heat` único → **cementar la escala `--heat-1..4` de v3** (fuente
más rica; resolver la divergencia a favor de v3). Tipografía: `--sans` (system-ui stack) · `--mono`
(nombres/IDs/métricas).

**Extracción — asimétrica:** color/tema = **lift mecánico** de los bloques `:root` (ya son CSS vars).
**spacing/radius/type NO están tokenizados en la fuente** (hardcodeados: radios `{4..14}px`, font-sizes
`{8.5..16}px`, sombras ad-hoc) → **curación manual** (snap a escala nombrada `--radius-sm/md/lg`,
`--fs-*`, `--shadow-node/hover`), no extracción automática (una tool solo vomita cada magic number).
Proceso: script que LISTA valores distintos para informar la curación → luego hand-author el
`.tokens.json`.

**Dark mode (la estrategia de los mockups, y la que pide CLAUDE.md):**
```css
@media (prefers-color-scheme: dark){ :root{ …dark… } }   /* señal por defecto */
:root[data-theme="light"]{ …light… }                     /* override que siempre gana… */
:root[data-theme="dark"]{ …dark… }
```
Toggle estampa `data-theme` en root; Zustand+hash-state guarda la preferencia. **Las 4 capas del Mapa
(Estructura/Tokens/Desempeño/Proceso) NO son colores — son modos de recoloreo:** un `data-capa` reasigna
un token indirecto (`--node-fill`/`--node-accent`), no 4 paletas. Salud/tipo/heat quedan como tokens
semánticos reutilizables que las capas remapean.

**Enforcement (doble candado):** (a) Tailwind ya enforça en utility-land (solo expone el tema); (b)
para CSS a mano (overrides React Flow, nodos, CodeMirror) **stylelint + `scale-unlimited/declaration-
strict-value`** prohíbe valores mágicos → obliga `var(--…)`:
```json
// .stylelintrc.json → enforced_by de los nodos de tokens
{ "plugins": ["stylelint-declaration-strict-value"],
  "rules": { "scale-unlimited/declaration-strict-value": [
    ["/color$/","background-color","border-color","fill","stroke","box-shadow"],
    { "ignoreValues":["transparent","currentColor","inherit","none"], "disableFix": true } ] } }
```
Segundo check pal runner `arnesia conformance`: **assert de sync `.tokens.json` ⟷ `@theme` generado**
(diff del output de SD5 vs committed) → detecta divergencia mockup↔código.

**Convención de vars — RESUELTO (ver reconciliación 3):** shadcn/assistant-ui esperan `--background/
--foreground/--border/--ring`; se **adopta la convención shadcn como capa semántica canónica**
(renombrar, no shim); los mockups se leen como fuente de VALORES.

Fuentes: w3.org/community/design-tokens (estable 2025.10) · designtokens.org/tr/2025.10/format · faq ·
styledictionary.com/reference/utils/dtcg · npmjs.com/package/style-dictionary · style-dictionary issues
#1494 #1398 (composites) · terrazzo.app/docs · tailwindcss.com/blog/tailwindcss-v4 · reactflow.dev/learn/
customization/theming + /examples/styling/tailwind · github.com/AndyOGo/stylelint-declaration-strict-value
· michaelmang.dev (linting design tokens). **No verificado:** convención exacta de vars de la versión
actual de assistant-ui (asumida shadcn-style); estado preciso de bugs SD5 composite (abiertos, no
cerrados verificados).

## Flags de confianza (no confirmado / verificar antes de cementar)

- **`pages` sin router / lienzo único** — reinterpretación (composition-roots por hash-state) inferida
  de las definiciones FSD oficiales, NO doctrina citada. Documentar en el nodo pa' que el equipo no la
  confunda con rutas.
- **`steiger` beta (0.5.0) y no-extensible** — nunca el único gate; los invariantes propios
  (transporte⊥dominio, single-SSE) van en dependency-cruiser.
- **shadcn ≈ Tailwind por defecto** — dependencia dura B↔E: decidir tokens/Tailwind ANTES de copiar
  componentes shadcn, o divergís del registry.
- **Determinismo de snapshots de canvas React Flow** — posible flakiness; fijar dims medidas +
  desactivar anim/`fitView`; correr visual en contenedor Linux. Spike antes de confiar.
- **Biome type-aware ≠ 100%** — vender `tsc --noEmit` como la red real de tipos, no Biome.
- **`exactOptionalPropertyTypes`** — fricción con spread de props React / tipos de terceros; mantener,
  documentar patrón.
- **Composites DTCG (typography/dimension) con bugs abiertos en SD5** (#1494, #1398) — mantener
  tipografía/dimensión como tokens simples al inicio.
- **DTCG v3 vs shell-A divergen en `--heat`** — resolver a favor de la escala `--heat-1..4` de v3 al
  cementar el `.tokens.json`, o quedan dos verdades.
- **Node 20.16+/22.19+/24+ exigido por SB10** — fijar en CI y en el toolchain.
