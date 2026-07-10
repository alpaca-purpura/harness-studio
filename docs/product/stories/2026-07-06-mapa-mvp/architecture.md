# ArnesIA · Mapa MVP — `architecture.md` (el CÓMO senior)

> **Paquete Fase C, doc 3 de 3.** Responde **cómo implementar bien** el comportamiento de
> [`spec.md`](./spec.md) con el look de [`design.md`](./design.md). Método RFC: contexto → decisión →
> consecuencias · alternativas · honestidad de lo `deferred` · plan secuenciado. Principios explícitos:
> **hexagonal · FSD-lite · SOLID · cero magic-value · honestidad de deferred · canvas ⊥ chrome**.
>
> Norma: nada «hecho» sin abrir el archivo real; lo diferido se declara honesto (no verde falso).

---

## 1 · Estado actual — auditoría de la arquitectura as-code

Verificado `archivo:línea` (3 subagentes de lectura, Fase C). Distingue **existe / falta /
`proposed` vs `enforced`**.

### 1.1 · Boundaries FE (`docs/architecture/boundaries/fe-*.md`)

| Boundary | Regla (H1) | `status:` |
|----------|-----------|-----------|
| `fe-topologia-fsd.md` | «La SPA se estructura por Feature-Sliced Design (import direccional)» (`:29`) | **`proposed`** (`:5`) |
| `fe-taxonomia-componentes.md` | «La taxonomía enforça dirección y pureza… **canvas ⊥ chrome = la frontera de mayor valor**» (`:27,60`) | **`proposed`** (`:5`) |
| `fe-transporte-independiente.md` | «El dominio FE no depende del transporte; SSE es singleton en `app`» (`:23`) | **`proposed`** (`:5`) |
| `fe-tokens-contrato.md` | «Los design tokens son el contrato mockup↔código (DTCG SSOT, cero valores mágicos)» (`:26`) | **`proposed`** (`:5`) |
| `fe-visual-fitness.md` | «Cada componente tiene una story que ES un test que rompe CI» (`:27`) | **`proposed`** (`:5`) |

> **Consecuencia:** los 5 boundaries que gobiernan el Mapa están **`proposed`, ninguno `enforced`**. El
> Mapa MVP es la primera superficie que los ejercita de verdad → §5.3 propone qué mover a `enforced`.

### 1.2 · Modelo C4 (`docs/architecture/model/`)

Existen `container.d2` + `system-context.mmd`. **No hay `INDEX.md`** en `model/` (el índice C4 vive en
`docs/architecture/INDEX.md`). El Mapa como container FE no está dibujado explícito → deuda menor (añadir el nodo
«Map surface» al `container.d2`, §9).

### 1.3 · Contratos (`docs/architecture/contracts/`)

- `schema/graph.l0.schema.json` — raíz `required:["nodos"]`, `additionalProperties:false` (`:7-8`);
  `$defs.nodo` `required:["id","clase","nombre"]` con **`additionalProperties:true`** (`:79-80`);
  `$defs.clase` (10), `$defs.banda` (7), `$defs.edge` (`additionalProperties:false`, tipo enum).
- `schema/box.contract.schema.json` — `required:["caja"]`, `additionalProperties:false` (`:7-8`);
  condicional `if caja===true then required [why,clase,arquetipo,perfil_harness,fase,estado,gate]` (`:165`).
- `api/openapi.yaml` + `gen/README.md` existen. **No hay `INDEX.md`** a nivel `contracts/`.

### 1.4 · Dominio + adapters backend (Go, hexagonal ya real)

| Pieza | Ubicación | Estado |
|-------|-----------|--------|
| Tipos L0 | `internal/domain/graph.go` (Graph/Arnes/Edge/Spine) + `box.go` (Box/Contract/Banda/Clase) | **real, testeado** |
| Puerto índice | `internal/ports/index.go:11-16` (`IndexPort{Rebuild,Query}`) | **real** |
| Adapter índice | `internal/adapters/index/store.go` — `map[string]domain.Graph` in-memory (TODO SQLite `:4-6`) | **real, stub de datos** |
| Use-case | `internal/usecase/map_service.go:24-26` (`MapService.Graph` → `index.Query`) | **real** |
| Transporte | `internal/adapters/transport/http/router.go` (pkg `httpapi`) | **real** |

### 1.5 · Fitness / convenciones — **deuda registrada (honesta)**

| Herramienta | Estado real | Evidencia |
|-------------|-------------|-----------|
| `go-arch-lint` | declarado en 3 boundaries; **binario + `.go-arch-lint.yml` ausentes** → inoperable | [[hs-09-fase5-mapa]] |
| `dependency-cruiser` | **roto**: usa `module.exports` en repo `type:module` ESM (`"module is not defined"`); globs `canvas-not-chrome` apuntan a rutas stale (`widgets/command-rail\|dock`, `app/shell` inexistentes) | [[hs-09-mapa-sustrato-html-svg]] |
| `vitest.workspace.ts` | usa `defineWorkspace` (removido en vitest 4.1.9 → requiere `test.projects`) → **story=test no ejecuta** | [[hs-09-mapa-sustrato-html-svg]] |
| tokens regen | `base.tokens.json` trae 11 `--c-*`; `theme.css`/`tokens.ts` sólo emiten **6** (`tokens:build` no re-corrido) | §6a |
| 211 checks del ruleset | `deferred` (linters externos no cableados a CI) | [[hs-08-doctrina-ejecutable]] |

> **Estos 5 ítems NO bloquean el Mapa read-only**, pero SÍ bloquean el enforcement real (story=test,
> canvas⊥chrome, cero-magic-value). El plan (§9) los agenda como trozo propio antes de declarar
> boundaries `enforced`.

### 1.6 · Convenciones (`docs/architecture/conventions/`)

8 nodos, **todos `status: proposed`** (verificado línea 5 de cada): `go-style.md` · `ts-style.md` ·
`ts-types.md` · `naming.md` · `commits.md` · `hooks.md` · `editor.md` · `ci.md` (+ `INDEX.md` + `CADENCE.md`).
Herramientas declaradas, **no cableadas a CI aún**: **golangci-lint v2** (Go) · **Biome v2.4** + `tsc`
@tsconfig/strictest (TS) · **lefthook** (pre-commit, binario Go) · **stylelint** (anti-magic-value CSS).
Estado real: config files declarados con `if: hashFiles` en CI → **`proposed`, no corriendo**
([[hs-arquitectura-fe]]). El Mapa MVP debe pasarlos localmente (Biome/`tsc`/stylelint verdes) aunque el
gate de CI siga diferido. → §7 (estrategia de pruebas).

---

## 2 · Arquitectura propuesta del Mapa — hexagonal + FSD-lite

### 2.1 · Principios (invariantes de diseño)

1. **Dominio puro** — el modelo del grafo (`entities/arnes/model`) NO importa transporte ni UI; es datos
   + selectores puros. (Espeja el `internal/domain` de Go.)
2. **canvas ⊥ chrome** (regla estrella, `fe-taxonomia-componentes.md:60`) — `widgets/map-canvas` es
   **canvas**: no conoce el rail de sesiones, el topbar ni el dock. El chrome invoca al canvas, nunca al
   revés.
3. **Transporte independiente** (`fe-transporte-independiente.md`) — `entities/arnes` **no** importa
   `shared/api`; el fetch lo dispara la **page** (composition-root) y pasa el grafo al store. El SSE
   (Hito 3) será singleton en `app/realtime/`.
4. **Tokens = contrato** (`fe-tokens-contrato.md`) — cero magic-value; sólo `var(--…)`.
5. **Story = test** (`fe-visual-fitness.md`) — cada componente nuevo trae su story-test.

### 2.2 · Ubicación FSD-lite (una capa por responsabilidad)

```
web/src/
  shared/canvas/glyph.tsx            # primitivo tonto (forma+color+char)
  entities/arnes/                    # DOMINIO del grafo (puro, sin transporte)
    model/{types,kind,selectors}.ts  # tipos L0 + clase→visual + proyecciones por banda/fase
    ui/arnes-node.tsx                # el .node + variantes
    testing/{dev-full-cycle,luana-feature-cycle}.ts
  widgets/map-canvas/                # CANVAS (⊥ chrome)
    ui/{map-canvas,band,lane,edge-layer, region, base-band, rules-subband,
        activation-chip, transition-tag, inspector}.tsx
    model/{layers,use-edge-paths,use-viewport}.ts
  pages/shell/ui/workspace-stage.tsx # composition-root: dispara api.getGraph, monta <MapCanvas>
  shared/api/                        # cliente HTTP (lo usa la page, no la entity)
```

> `map-bar.tsx` actual (id + META chips + tablist de capas) es **chrome extra que el mockup no tiene**
> (el mockup asume que la barra la pone el shell). Decisión: **mantener MapBar como chrome del shell**,
> fuera del canvas puro; el canvas expone la capa activa por prop, no la controla. → §6e (componentes)
> + §6g (`map-canvas`).

### 2.3 · Por qué HTML+SVG y no React Flow (decisión firmada)

React Flow impone su propio modelo de nodos/edges/viewport; el Mapa necesita **geografía fija swim-lane**
(bandas + carriles) que RF no da — habría que pelear su layout. El mockup v3 ya demuestra el patrón:
flex HTML + edges medidos por `getBoundingClientRect`. **RF queda para el Organigrama** (lienzo libre 2D).
[[hs-09-mapa-sustrato-html-svg]]. Alternativa considerada y descartada: RF con layout custom (coste de
fricción > beneficio; el pan/zoom lo resolvemos con una transform CSS de 3 líneas, §3.3).

> **Deuda doctrinal (stale):** `VISION.md:204` y el bloque de decisiones técnicas de CLAUDE.md aún dicen
> «Mapa: React Flow 12 + layout de carriles custom». La reversión a HTML+SVG está **firmada** (Fase B /
> Gate 1, [[hs-09-mapa-sustrato-html-svg]]) pero esas dos líneas **no se han actualizado** → tarea de
> propagación (no bloquea el Mapa; sí debe corregirse para que la doctrina no contradiga la implementación).

---

## 3 · Desglose de módulos + flujo de datos

### 3.1 · Flujo (Hito 1)

```
pages/shell/workspace-stage (composition-root)
  └─ api.getGraph("dev-full-cycle")            # shared/api → GET /api/harnesses/{id}/graph
       └─ store entities/arnes (setGraph)      # dominio puro, sin transporte
            └─ selectores puros:
                 selectGuardia(g)   → banda==="guardia"
                 selectLanes(g)     → 1 col por arnes.fases[] (+ append no-declaradas)
                 selectBase(g)      → bandas base/libreria/meta/terceros/dormidas
                 selectEdges(g)     → g.edges ?? []
            └─ <MapCanvas graph capa="estructura">   # widgets/map-canvas (canvas ⊥ chrome)
                 ├─ <Region r-guardia>  → <Band compact>       (chips)
                 ├─ <Region r-proceso>  → <Lane × fases>       (caja-first + support + hairline)
                 ├─ <Region r-soporte>  → <BaseBand>           (rules-subband + knowledge) + <Band×>
                 └─ <EdgeLayer>         → use-edge-paths (spine-always + hover-reveal, ÷z)
```

`selectLanes`/`selectGuardia`/`selectBase`/`selectEdges` ya existen (`entities/arnes/model/selectors.ts`).
El **layout de carriles es determinista**: orden = `arnes.fases[]`; sin física, sin auto-layout — posiciones
por CSS flex. Esto hace los snapshots de story reproducibles (a diferencia de un motor de grafo).

### 3.2 · Algoritmo de lane (determinista)

1. `fases = [...arnes.fases]`; por cada nodo `banda==="fase"` cuya `fase` no esté en `fases`, **append**
   (primera aparición). (`selFases`, espejo de `mockup:396-400`.)
2. Por lane: nodos de esa fase, **ordenados caja-first** (`contract.caja` primero).
3. Insertar `lane-sep` antes del primer no-caja si la lane tiene alguna caja.

### 3.3 · Viewport (pan/zoom/fit) — `use-viewport.ts` (nuevo)

Estado `{z, ox, oy}`; `transform: translate(ox,oy) scale(z)` sobre `.stage`. `fit()`, zoom ±0.1
clamp `[0.4,2]`, ctrl-wheel, drag-pan (ignora targets `.ctl/.node/.help`). Los edges dividen coords por
`z` (el SVG vive dentro del stage escalado). Espejo directo de `mockup:451-508`. Hash-state opcional
(zustand) para persistir zoom por sesión (fuera de MVP).

---

## 4 · Backend

### 4.1 · Hito 1 — loader del dogfood (ÚNICO cambio imprescindible)

**Hecho verificado (corrige CLAUDE.md):** el índice hoy sirve un **demo keyed `"demo"` de 3 nodos / 1
edge** (`store.go:55-111`), **no** 2 nodos, y **no** bajo el id `dev-full-cycle`. Por tanto
`GET /api/harnesses/dev-full-cycle/graph` → **404** hoy. (CLAUDE.md dice «demo de 2 nodos» — inexacto;
se registra la corrección.)

**Cambio (~30 LOC, sin tocar interfaz ni handler ni schema):** en `internal/adapters/index/store.go`,
extender `seed()` (o un `--seed-graphs`) para:

```go
b, err := os.ReadFile("dogfood/dev-full-cycle.graph.json")
var g domain.Graph
_ = json.Unmarshal(b, &g)          // las claves JSON ya casan con los tags Go (nodos/clase/banda/…)
s.graphs[g.Arnes.ID] = g           // key = "dev-full-cycle"
```

- **Por qué unmarshala limpio:** las claves del JSON (`nodos·clase·banda·fase·contract·de·a·tipo`) son
  exactamente los tags de `domain.Box`/`Edge` (subagente backend, verificado campo a campo).
- **Sin cambio de contrato:** el `Box` de Go ignora silenciosamente los campos PROPUESTOS si aparecieran
  (no hay ninguno en el dogfood json real).
- `Rebuild()` (`store.go:38`) se mantiene re-sembrando; la caminata JSONL real es Hito 3.
- El endpoint (`router.go:36` → `getHarnessGraph` → `MapService.Graph` → `IndexPort.Query`) **no cambia**.

### 4.2 · Hito 2 — inspector + picker

| Endpoint | Hoy | Hito 2 |
|----------|-----|--------|
| `GET /api/harnesses/{id}/nodes/{nodeId}` (`router.go:37,79-81`) | **501** `notImplemented` | devolver `Box`+`Contract` para el Inspector (RF-71) |
| `GET /api/harnesses` (`router.go:35,61-63`) | **200 `[]` vacío** | listar arneses del índice para el picker (RF-72) |

Ambos son lecturas del mismo índice; sin nuevas dependencias.

---

## 5 · Contrato / schema — campos PROPUESTOS + cementar las 6 decisiones

### 5.1 · Los 4 campos PROPUESTOS del mockup

Hoy `origen/alw/trans/prop` **no existen** en `domain` ni en schema (verificado). El `$defs.nodo` es
`additionalProperties:true` (`graph.l0.schema.json:80`) → un nodo que los traiga **pasa validación pero
se ignora** (Go los descarta al unmarshal). Promoverlos a primera clase =:

| Campo | Semántica ya existente | Decisión propuesta |
|-------|------------------------|--------------------|
| `trans` | `contract.estado` («entra -> sale») ya lo tiene (`box.go:210`) | **No añadir campo nuevo.** El MVP deriva la etiqueta de `contract.estado`; `trans` es sólo el string ya presente. |
| `prop` | `canal:"propuesto"` ya lo expresa (`box.go:141`, schema `:87`) | **No añadir campo nuevo.** El badge se deriva de `canal==="propuesto"`. |
| `alw` | «always-on» de una regla; hoy sólo display (`UX.md:534`) | **Campo nuevo `alw:bool` en nodo** (o derivarlo de `fuente_path`/scope de la regla). Requiere schema + `Box`. **PROPUESTA a ratificar.** |
| `origen` | ninguna (`procedencia`/`banda`/`canal` no lo cubren) | **Campo nuevo `origen: "estandar"\|"del-puesto"`.** Requiere schema + `Box` + decisión de **cómo se llena** (onboarding). **PROPUESTA — doctrina SILENTE.** ⚠️ Tensión con `maquinaria-no-contamina-arnes` (el kit genérico vive fuera del grafo; todo nodo del grafo es instancia ③). |

> **Estrategia MVP (honesta):** el Mapa dibuja `caja`/`transición`/`propuesto` desde datos **reales**
> (`contract.caja`, `contract.estado`, `canal`); dibuja `origen`/`alw` desde **heurística/fixture**
> (sets `DEL_PUESTO`/`alw` del mockup) y los etiqueta **PROPUESTA** en la UI, hasta que §5.2 los ratifique
> as-code. Nada se pinta como establecido sin campo real.

### 5.2 · Cementar las 6 decisiones firmadas (as-code)

| # | Decisión | Dónde cementarla | Cambio |
|---|----------|------------------|--------|
| 1 | Base canónico | ya en `VISION.md:81-83`; alinear título CSS `r-soporte`→semántica Base | doc + comentario (ninguna deuda de schema) |
| 2 | Activación por-nodo | `graph.l0.schema.json $defs.nodo` + `Box` + `docs/architecture/knowledge/rules.md` | **añadir `alw`** (o derivar); mover honestidad a check evaluable. PROPUESTA |
| 3 | Facet `origen` | `graph.l0.schema.json $defs.nodo` (+ enum) + `Box.Origen` + doctrina de onboarding | **añadir `origen`**; resolver la tensión maquinaria/instancia. PROPUESTA |
| 4 | Knowledge path-scoped + semántico opcional | ya doctrina `METODOLOGIA.md:120-122,§8.6`; el índice semántico = MCP `canal:propuesto` | ninguna deuda dura; el semántico = **debate abierto** (`VISION.md:244`), jamás dependencia dura (framing Langfuse `VISION.md:147,191`) |
| 5 | Fase = dato del arnés | ya doctrina (`fase` sin enum, `box.go:160`; schema `:86`) | ninguna — confirmar que el layout no universaliza fases |
| 6 | Paquete = artefacto/spine, caja 1 transición | ya doctrina `METODOLOGIA.md:160,191` + `VISION.md:69-73` | el Mapa muestra 1 transición/caja; el tablero «Flujo» (paquete vivo) es vista aparte, telemetry-gated |

### 5.3 · Boundaries a mover `proposed → enforced`

Cuando el Mapa MVP aterrice **con** su tooling arreglado (§7): `fe-tokens-contrato` (cero-magic-value en
el Mapa, vía stylelint) y `fe-visual-fitness` (story=test del Mapa corriendo) son los primeros candidatos
a `enforced`. `fe-taxonomia-componentes` (canvas⊥chrome) requiere primero arreglar `dependency-cruiser`
(§7). Los otros dos quedan `proposed` hasta tener SSE (Hito 3).

### 5.4 · Resolución de literales de medida (deuda de magic-value no-color)

`design.md §12` difiere aquí los **literales de medida** del mockup firmado (no son color; la regla dura
anti-magic-value es de color, `hs-norma-pegarse-storybook`). Resolución propuesta al portar:

| Literal (mockup) | Uso | Resolución |
|------------------|-----|------------|
| `radius:14px` | `.region` (`:54`) | añadir `--radius-2xl:14px` a `base.tokens.json` (o mapear a `--radius-xl:12`) |
| `border-radius:6px` | `.node-cmd` (`:112`) | `--radius-sm:6px` (ya en source, `base.tokens.json:250`) |
| `font-size:9px/10px` | badges, glyph, `.tw` | escala micro bajo `--text-xs:11`; documentar o añadir `--text-2xs` |
| `box-shadow rgba(0,0,0,.12/.18/.22)` | controles/paneles | añadir token de sombra (`--shadow-sm/md/lg`); hoy no existe → deuda |
| `margin-left:14px`, gaps 5/6/7/8/10/16/18 | spacings | snap a `space` base-4 (`--space-*`) donde encaje; documentar el resto |

Ninguno bloquea el Mapa; se cementan al portar cada componente (stylelint no los marca — sólo vigila color).

---

## 6 · Sincronización Storybook (SSOT) — deltas exactos por componente

El mockup **divergió** del Storybook (es un prototipo data-driven independiente). Para el look firmado,
por componente:

### 6a · Tokens (regenerar) — `theme.css` + `tokens.ts`
Correr `tokens:build` para emitir los 5 `--c-*` que faltan (`command/plugin/settings/output-style/
statusline`, ya en `base.tokens.json:191-210`). Luego **repuntar `kind.ts`** (`:23-37`) de
`var(--muted-foreground)` a `var(--c-command)`… **Bloqueante del look** (sin esto las 5 clases config
son grises indistinguibles).

### 6b · `glyph.tsx`
`size` default **15→17**; `fontSize` queda `round(size*.58)`=~10 (OK). Clip-paths **ya idénticos** al
mockup (`:11-16` == `mockup:117-120`). +5 stories de las clases config.

### 6c · `arnes-node.tsx` (el mayor delta)
Añadir variantes/props (hoy sólo `hover`+`selected`): `caja` (border-left 5px + tint + `caja-badge`),
`support` (indent + bg tenue), `compact` (fila 1 línea), `puesto` (border dashed), `dim` (opacity .4),
`node-cmd` (handle derivado por clase), `node-trans` (etiqueta transición), `prop-badge`. Nuevos props:
`variant`/`caja`/`dim`/`handle`/`trans`/`origen`/`prop`. Medidas → `design.md §5-6`.

### 6d · `band.tsx` / `lane.tsx`
`lane` width `min-w-44 flex-1` → **`width:232px` fijo**; `band` node-wrapper `w-44` (176) → **232px**.
`lane` añade **orden caja-first** + `lane-sep` + `.support`. `band` gana modo **compact** (Guardia).

### 6e · Componentes NUEVOS (no existen hoy)
`region` (3 territorios tintados) · `base-band` + `rules-subband` (colapsable, split siempre/condicional)
· `activation-chip` (`act-chip`) · `transition-tag` (`node-trans`, si se separa del nodo) · `inspector`
(panel S3, Hito 2). `map-bar` **se mantiene como chrome del shell** (no entra al canvas puro).

### 6f · `edge-layer.tsx` + `use-edge-paths.ts`
Recolorear: `invoca` `--input`(gris)→**`--crit`**; `lee` `--input`→**`--warn`** dash `4 4`; `escribe`
`--primary`→**`--ok`** dash `3 3`. Ancho 1.4→**1.6** (2.2 en foco). Añadir **gating** spine-always +
hover-reveal + `÷z`. Falta además crear `edge-layer.stories.tsx` (no existe).

### 6g · `map-canvas.tsx`
Envolver en `<Region>`s (hoy `<Band>`/`<Lane>` planos); añadir `use-viewport` (pan/zoom/fit) + hover-focus
+ subbands colapsables; swap del `<ComingSoon>` en `workspace-stage.tsx:41-46` por `<MapCanvas>`.

> **Story = test** para cada componente tocado/nuevo (fitness visual `fe-visual-fitness`).

---

## 7 · Estrategia de pruebas

| Nivel | Herramienta | Acción MVP |
|-------|-------------|------------|
| Fitness visual | Storybook 10 story=test | 1 story-test por componente (glyph/arnes-node×variantes/region/lane/base-band/edge-layer/map-canvas) con fixtures Luana+dogfood. **Requiere arreglar `vitest.workspace.ts`** (`defineWorkspace`→`test.projects`) o no ejecuta |
| Tipos | `tsc` @tsconfig/strictest | verde |
| Lint TS | Biome v2.4 | verde |
| Boundaries FE | dependency-cruiser | **arreglar primero** (`module.exports`→ESM; globs canvas⊥chrome a rutas reales `session-rail\|chat-dock\|topbar\|view-strip` + `pages/shell`) |
| Anti-magic-value | stylelint | verde en el Mapa (sólo `var(--…)`) |
| Backend | `go build/vet/test ./... -race` | verde; test del loader dogfood (unmarshal → 5 nodos/4 edges) |
| Contrato | validar `dev-full-cycle.graph.json` contra `graph.l0.schema.json` | verde |

Paridad visual final: click-through en chrome-devtools con asserts + screenshots revisados + **consola
limpia** (los 7 shots de este paquete son la línea base).

---

## 8 · Escalabilidad (costuras sin sobre-ingeniería)

- **N arneses en índice** — `map[id]Graph` ya soporta N; el picker (Hito 2) los lista. SQLite (modernc)
  reemplaza el `map` cuando el volumen lo pida (TODO ya marcado `store.go:4-6`), sin tocar `IndexPort`.
- **N sesiones** — el canvas es stateless respecto a la sesión (recibe el grafo por prop); multisesión
  vive en el shell, no en el canvas (canvas⊥chrome lo garantiza).
- **Indexer JSONL real** — reemplaza el `seed()` stub detrás de `IndexPort` (Hito 3); el FE no se entera
  (mismo endpoint/forma).
- **Realtime** — `event: map` por SSE singleton en `app/realtime/`; el store `entities/arnes` acepta un
  `patchGraph`. El puerto ya está pensado (`fe-transporte-independiente`).
- **Capas de telemetría** — Tokens/Desempeño/Proceso son `data-capa` sobre la MISMA geografía
  (`fe-tokens-contrato.md:71-73`: no son 4 paletas), se activan cuando el indexer emita métricas; hasta
  entonces «—».

---

## 9 · Riesgos · trade-offs · alternativas · secuencia

### 9.1 · Riesgos / trade-offs

| Riesgo | Mitigación |
|--------|------------|
| Campos PROPUESTOS (`origen/alw`) sin schema → tentación de hardcodear sets en el FE | Etiquetar PROPUESTA en la UI; aislar los sets en `entities/arnes/model` como fixture, no en el canvas; ratificar as-code (§5.2) antes de considerarlos dato |
| Tensión `origen` vs `maquinaria-no-contamina-arnes` (kit genérico fuera del grafo) | Decisión de arquitectura pendiente: ¿el `origen` se computa al provisionar (instancia ③) o se declara? — abrir ficha antes de tocar schema |
| Tooling roto (depcruise/vitest) da falsa sensación de verde | §7 lo arregla ANTES de declarar boundaries `enforced`; hasta entonces, honesto: «story-tests escritas, no ejecutando» |
| `additionalProperties:true` en `$defs.nodo` esconde typos de campo | Al cementar `alw`/`origen`, evaluar cerrar el nodo a `false` (como arnes/edge/contract ya están) |
| Snapshots de canvas no deterministas | El layout es CSS-flex determinista (no motor de grafo) → snapshots estables; medir en viewport fijo (1680×1000, como los shots) |

### 9.2 · Alternativas consideradas

- **React Flow para el Mapa** — descartada (§2.3); RF → Organigrama.
- **Añadir `origen`/`alw` al schema ya** — descartada para el MVP (doctrina SILENTE/PROPUESTA; se dibujan
  como propuesta primero, se ratifican después).
- **Fusionar MapBar al canvas** — descartada (rompe canvas⊥chrome); MapBar = chrome del shell.

### 9.3 · Secuencia de implementación (PRs por trozo coherente)

1. **PR-A · Backend loader** — `store.go` carga `dev-full-cycle.graph.json` bajo su id + test
   (unmarshal → 5 nodos/4 edges). Endpoint ya sirve. *(Desbloquea datos reales.)*
2. **PR-B · Tokens + glyph** — regenerar `theme.css`/`tokens.ts` (5 `--c-*`), repuntar `kind.ts`, glyph
   15→17. *(Desbloquea color correcto.)*
3. **PR-C · arnes-node variantes** — caja/support/compact/puesto/dim + handle + node-trans + badges +
   stories. Medidas de `design.md §5-6`.
4. **PR-D · Regiones + lanes + Base** — `region`, `lane` (232 + caja-first + support/hairline), `band`
   compact (Guardia), `base-band`/`rules-subband` colapsable + `activation-chip` + stories.
5. **PR-E · Edges** — recolorear crit/ok/warn + gating spine-always/hover-reveal + `÷z` +
   `edge-layer.stories.tsx`.
6. **PR-F · Viewport + montaje** — `use-viewport` (pan/zoom/fit) + hover-focus + **swap `ComingSoon`** en
   `workspace-stage.tsx` + click-through chrome-devtools (paridad contra los 7 shots).
7. **PR-G · Tooling (paralelo/antes de enforce)** — arreglar `vitest.workspace.ts` + `dependency-cruiser`;
   correr story=test; mover `fe-tokens-contrato`/`fe-visual-fitness` a `enforced`.
8. **Hito 2 (post-MVP)** — `getNode`/`listHarnesses` reales + `inspector` + picker.

Dependencias: A independiente; B independiente; C→depende de B; D→depende de C; E→depende de C/D; F→depende
de D/E; G paralelo (necesario antes de `enforced`). A+B pueden ir en paralelo.

---

## Apéndice · Índice cruzado

- Comportamiento (RF numerados + Gherkin) → [`spec.md`](./spec.md).
- Medidas/tokens al pixel → [`design.md`](./design.md).
- Decisiones firmadas y su respaldo doctrinal → [`spec.md §10`](./spec.md) + §5.2 de este doc.
- Deuda paralela (no bloquea el Mapa) → §1.5 + [[hs-09-fase5-mapa]].
