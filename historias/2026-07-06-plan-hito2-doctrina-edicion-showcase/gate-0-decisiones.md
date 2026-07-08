# Gate 0 — decisiones firmadas (encuadre as-code)

> Ficha HS-09 (fase 5) · plan Hito 2 · 2026-07-06 · **FIRMADO por el operador**
> Regla del retro: ninguna fase de código arranca sin su gate firmado. Este doc cementa las 4 decisiones ANTES de codear.

## D1 — Encoding de doctrina en la tarjeta: **fila-meta completa**

Los 3 ejes de CLASIFICACIÓN (`clase ⊥ arquetipo ⊥ perfil_harness`) + el `gate` + `procedencia` se hacen visibles **en la tarjeta**, no solo en el inspector. Somos el promotor de la metodología: el mapa mismo la muestra.

**Spec de render (cementa `ui-doctrina-visible.md` §diseño):**

| Eje | Marca | Token / mecanismo | Valores |
|---|---|---|---|
| `arquetipo` | char-badge mono | `--muted-foreground` + forma (color NO discrimina) | pipeline=`═` · excepcion=`≈` · abierto=`✳` · no-arnesar=`○` |
| `perfil_harness` | pill mono | `color-mix(in srgb, var(--foreground) N%, transparent)` — opacidad creciente | T1=tenue · T2=media · T3=plena («más autonomía → más presencia») |
| `gate.tipo` | punto de estado | `color.health` + `--c-skill` | auto=`--ok` · parcial=`--warn` · **manual=`--c-skill` (azul)** · **none=`--crit`, anillo punteado** |
| `procedencia` | modulación del borde ya existente | opacidad/estilo de borde (sin color nuevo) | medido=sólido pleno · declarado=normal · estimado/inferido=atenuado · no-declarado=hairline gris |

**Reglas de ubicación:**
- La **fila-meta** (`arquetipo · perfil · gate`) va al pie de la tarjeta **solo cuando `caja:true`**.
- `no-arnesar` (`○`) nunca aparece en fila-meta de caja (fuerza `caja:false`) → se pinta en el nodo de apoyo.
- `procedencia` modula el borde de **todo** nodo (caja o no).

**Compliance (verificado):** **CERO tokens nuevos** — reusa `--muted-foreground`, `--foreground`, `color.health` (`--ok/--warn/--crit`), `--c-skill`. → sin churn de `tokens-sync`, sin tocar `fe-tokens-contrato`. Todo valor de color = `var(--…)` o `color-mix` sobre var (pasa stylelint strict-value).

**⚠ Riesgo a chequear en Gate 2:** `manual=--c-skill` (azul) coincide con el color del glyph de clase `skill` (azul). El operador aprobó «manual=azul» sobre el preview; verificar en la app real que el punto de gate no se confunde con el glyph. Fallback: `--foreground` neutro para manual.

**Fitness:** una **story=test** por valor nuevo — por arquetipo (4), por perfil (3), por gate.tipo (4, incl. `none`), por procedencia (5). Rompen CI si regresan.

## D2 — Superficie del run T3: **endpoint dedicado `POST /api/harnesses/{id}/boxes/{boxId}/run`**

El conductor T3 es su propio recurso, separado del turn conversacional del Dock. Respeta `orquestacion-determinista-entre-cajas` (el loop es código, no el turn humano). Cementa en OpenAPI en Fase 3 (contract-first de ese track).

## D3 — Canal del `control_request`: **Dock derecho + tarjeta de diff inline**

La aprobación humana de escrituras (role/ttl) vive en el Dock, como tarjeta con diff (assistant-ui + CodeMirror 6/merge, stack HS-04). Un solo lugar: conversar + aprobar. Realiza `permisos-gui-human-in-the-loop`.

## D4 — Aislamiento: **main directo, secuencial (sin worktree)**

Trunk-based puro (como el resto del repo). Los tracks NO corren en paralelo: **orden A → B**.
- **Track A primero** (Fase 1 showcase + Fase 2 doctrina-visible) — usa backend ya vivo (`getGraph`/`getNode`), entrega el valor «metodología visible» sin backend nuevo.
- **Track B después** (Fase 3 Fase E backend + Fase 4 join).

Esto ajusta el `INDEX.md`: donde decía «tracks paralelizables / worktree» → **secuencial en main, A→B**.

## Contratos as-code a cementar en su fase (no ahora)

- **Fase 2 (Track A):** las stories = test de las marcas nuevas SON el contrato de fitness visual; el mockup `arnesia-mapa-mvp.html` se regenera desde Storybook con la fila-meta.
- **Fase 3 (Track B):** `arch/contracts/api/openapi.yaml` — 3 rutas (`/boxes/{id}/run`, `/sessions/{id}/permission` real, ruta de conform) contract-first ANTES del handler; changelogs de los 5 boundaries tocados.

## Estado

**Gate 0 = FIRMADO.** Siguiente: **Fase 1 — showcase dogfood** (promover `showcase.graph.json` → `dogfood/`, validar con `arnesia conformance --arnes`, espejar fixture FE + stories) → **Gate 1** (ojo-UI + operador).
