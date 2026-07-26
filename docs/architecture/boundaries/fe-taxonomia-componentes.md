---
regla: fe-taxonomia-componentes
version: 1.2
updated: 2026-07-25
status: enforced
ledger: HS-09
sources:
  - url: https://feature-sliced.design/blog/atomic-design-architecture
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://bradfrost.com/blog/post/extending-atomic-design/
    autoridad: experto
    revisado: 2026-07-05
  - url: https://reactflow.dev/learn/customization/custom-nodes
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://ui.shadcn.com/docs/changelog
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - web/.dependency-cruiser.js#canvas-not-chrome
  - web/.dependency-cruiser.js#chrome-not-canvas-internals
  - web/.dependency-cruiser.js#ui-not-domain
  - web/.dependency-cruiser.js#no-sibling-widget-imports
severity: high
---

# La taxonomía de componentes enforça dirección y pureza, no un ladder atómico

## L1 · Principio (estándar de industria)

**Atomic Design = taxonomía, no topología.** El consenso maduro 2026 (incl. el propio Brad Frost:
«taxonomía, no dogma») es que la jerarquía de 5 niveles (atoms/molecules/organisms/templates/pages) es
buen *vocabulario mental* pero mala *estructura de carpetas* — la frontera atom/molecule/organism es
difusa, y esa ambigüedad es el modo de falla. Lo enforçable no es «de qué está hecho un componente»
sino **quién puede importar a quién**. *(experto: Frost; estándar: FSD atomic-vs-topology)*

**Componente de canvas ≠ átomo reutilizable.** Un nodo del Mapa (HTML+SVG con geografía de bandas/
carriles y edges medidos por `getBoundingClientRect`) vive dentro del viewport del canvas, con
posición/hover/foco gobernados por su store — no es un primitivo de chrome recomponible. (El Organigrama
usará React Flow, mismo principio: custom node dentro de `ReactFlowProvider`.) La separación de runtimes
(canvas vs shell) es una frontera real. *(oficial: React Flow custom nodes — patrón análogo)*

**Primitivos copy-in.** shadcn/ui (sobre Base UI) trae los primitivos como archivos EN el repo →
lintables y enforçables (no opacos en `node_modules`), bundle chico, sin `ThemeProvider` lock-in.
*(oficial: shadcn changelog)*

## L2 · Realización (este árbol Go+React)

Atomic vive DENTRO de `shared/ui` como vocabulario; la estructura real = **6 capas de UI direccionales**
(no compiten con FSD del nodo [`fe-topologia-fsd`](./fe-topologia-fsd.md); lo refinan):

| Capa | Path | Ejemplos ArnesIA | No importa |
|---|---|---|---|
| primitives (átomos) | `shared/ui/primitives/` | Button, Chip, Tooltip, Input, Icon, Popover | canvas, chrome, features, store |
| compositions (moléculas) | `shared/ui/` | DockRow, Toolbar, Menu, Card, SearchField | canvas, store, dominio |
| canvas (rama hermana) | `shared/canvas/`, `entities/*/ui`, `widgets/map-canvas` | NodeShell, Port, Handle, BandLane, ProcessBox, bandas | **chrome**, views |
| entities/features ui (organismos de dominio) | `entities/*/ui`, `features/*/ui` | ArnesCard, fila de timeline del Dock | canvas internals, otra feature |
| chrome/shells (organismos de app) | `widgets/*`, `app/shell` | Command Rail, Dock, Organigrama, Grid, Dialogs | **canvas internals** (solo props/store) |
| views/pages | `pages/*` | Vista Mapa, Vista Portafolio | ser importado |

- **canvas ⊥ chrome = la frontera de mayor valor** (runtimes distintos): `widgets/map-canvas` (canvas) no
  importa el chrome real (`widgets/{session-rail,chat-dock,topbar,view-strip}`, `pages/shell`), ni al revés
  a los internos del canvas (solo props/store). ⇐ L1: canvas ≠ átomo. El Dock (assistant-ui/AG-UI) es
  **chrome**, no canvas. **Verificado enforced (HS-09):** `pnpm depcruise` verde sobre el Mapa real (76
  módulos, 0 violaciones); `canvas-not-chrome`·`chrome-not-canvas-internals`·`ui-not-domain` en `error`.
- **primitives puros:** no tocan store ni dominio; consumen **solo tokens semánticos** (`var(--…)`),
  nunca hex (ver [`fe-tokens-contrato`](./fe-tokens-contrato.md)). ⇐ L1: atomic = vocabulario.
- **Primitivos = shadcn sobre Base UI** (copy-in en `shared/ui/primitives/`); React Flow UI se trae por
  el mismo registry shadcn. MUI/Mantine/Chakra descartados (provider lock-in, bundle 2-3×). ⇐ L1:
  copy-in.
- **NO se enforça «atom no importa organism»** (teatro con tiers difusos); se enforça dirección
  (canvas⊥chrome) y pureza (ui⊥dominio). ⇐ L1: quién importa a quién.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| canvas-not-chrome | el canvas (`map-canvas`, `shared/canvas`) no importa chrome (`session-rail`, `chat-dock`, `topbar`, `view-strip`, `pages/shell`) | error | «nodo del Mapa importa el Rail/Dock (mezcla runtimes canvas/shell)» | dependency-cruiser#canvas-not-chrome |
| chrome-not-canvas-internals | el chrome no hace deep-import a internos del canvas (solo props/store) | error | «chrome alcanza los internos del canvas (Mapa HTML+SVG)» | dependency-cruiser#chrome-not-canvas-internals |
| ui-not-domain | `shared/ui/**` no importa `entities/features/*/model` ni el store | error | «primitivo/molécula acoplado al dominio» | dependency-cruiser#ui-not-domain |
| no-sibling-widget-imports | un widget no importa otro widget (cross-slice de chrome): la página los compone | error | «dos organismos de app acoplados entre sí en vez de por la vista» | dependency-cruiser#no-sibling-widget-imports |
| primitives-solo-semanticos | `shared/ui/**` no usa colores/valores crudos, solo `var(--…)` | warn | banda Base «primitivo con valor mágico (no tokenizado)» | stylelint (ver fe-tokens-contrato) |

## Changelog

- 2026-07-25 · v1.2 · **+1 check `no-sibling-widget-imports`** (paquete
  `docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`, `design.md` §11.4). La capa
  `widgets` era la única capa de slices sin regla de cross-import: `no-sibling-feature-imports` ya
  cubría `features/`, pero dos widgets podían acoplarse libremente. El paquete de marketplaces es el
  primero que pone **dos widgets de la misma vista** (`widgets/marketplace` y `widgets/portafolio`,
  compuestos por `pages/shell/ui/portafolio-view.tsx`), donde el acoplamiento sería invisible sin
  gate. Regla espejo de la de features, `severity: error`. **Verificada en vivo antes de agregarla:**
  el árbol de hoy tiene **cero** imports cruzados entre widgets, así que la regla nace verde
  (`depcruise src`: «no dependency violations found», 122 módulos / 314 dependencias). 4 → **5
  checks**; sin cambio de `status` (sigue `enforced`).

- 2026-07-06 · v1.1 · `proposed → enforced` (HS-09, Fase D). El Mapa MVP es la primera superficie que
  ejercita **canvas⊥chrome** de verdad: `widgets/map-canvas` importa solo `entities/arnes` + `shared/*`,
  cero chrome; `pnpm depcruise` verde (76 módulos, 0 violaciones) con `canvas-not-chrome`·
  `chrome-not-canvas-internals`·`ui-not-domain` en `error`. Corregidos los nombres stale del chrome
  (`command-rail`/`dock` → los reales `session-rail`/`chat-dock`/`topbar`/`view-strip` + `pages/shell`,
  casando el enforcer) y el ejemplo canvas≠átomo (Mapa HTML+SVG; React Flow queda para el Organigrama).
  Sin cambio de checks (4).
- 2026-07-05 · v1.0 · Nodo fundacional FE (HS-05). L1 = atomic-es-taxonomía (Frost/FSD) + canvas≠átomo
  (React Flow) + shadcn copy-in. L2 = 6 capas de UI direccionales dentro de FSD; canvas⊥chrome como
  frontera estrella; shadcn sobre Base UI. No se enforça el ladder de 5. 4 checks.
