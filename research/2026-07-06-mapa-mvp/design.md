# ArnesIA · Mapa MVP — `design.md` (el UI exacto)

> **Paquete Fase C, doc 2 de 3.** Hermano de [`spec.md`](./spec.md) (el QUÉ) y
> [`architecture.md`](./architecture.md) (el CÓMO). Este documento congela el **look al pixel** del
> mockup firmado `mockups/arnesia-mapa-mvp.html` (Gate 1 ✓, commit `0736d2c`) para reconstruirlo
> **sin ver el mockup**. Regla de oro: **cero magic-value** — todo valor de color es `var(--token)`;
> las medidas px se citan con su `archivo:línea` del mockup + screenshot + componente Storybook.
>
> **Convención de referencia.** Cada bloque cita: **`{token}`** (de `web/tokens/base.tokens.json` →
> var CSS corta `--x`) · **`mockup:NN`** (`mockups/arnesia-mapa-mvp.html`) · **`shot#`**
> (`mockups/.shots/…png`) · **`SB`** (componente Storybook en `web/src/…`). Los deltas contra el
> componente actual viven en [`architecture.md §6`](./architecture.md).

---

## 0 · Índice de referencias visuales (screenshots)

Capturadas con chrome-devtools, fixture **Luana**, viewport **1680×1000**, **consola limpia** (Paso 0
del brief). Valores medidos con `evaluate_script` (no supuestos), citados como datos.

| # | Archivo | Estado | Medición real |
|---|---------|--------|---------------|
| **shot1** | `mockups/.shots/mapa-overview-light.png` | Overview, tema **light**, fit al abrir | fit-zoom `scale(0.906)`; content `1802×979`px; 3 regiones; 37 nodos; 11 edges invoca |
| **shot2** | `mockups/.shots/mapa-overview-dark.png` | Overview, tema **dark** | `--background:#14171c`; `--c-skill:#5b9dd9` (dark) |
| **shot3** | `mockups/.shots/mapa-hover-nodo.png` | Hover sobre `builder` | 32 nodos `.dim`; 13 edges (11 spine + 2 `lee` revelados) |
| **shot4** | `mockups/.shots/mapa-reglas-colapsadas.png` | Base **enfocada** (z=1, oy −432), subband Reglas **colapsada** (default) | `.subband.collapsed`; flecha `▸`; encuadre propio de la región Base (distinto de shot1) |
| **shot5** | `mockups/.shots/mapa-reglas-expandidas.png` | Base **enfocada** (z=1, mismo encuadre que shot4), Reglas **expandida** | flecha `▾`; 2 rule-grps (siempre/condicional); 10 nodos regla (5 always-on + 5 condicional) |
| **shot6** | `mockups/.shots/mapa-ayuda.png` | Panel de ayuda/leyenda abierto | 6 glyphs de leyenda; headings: Relaciones·Clases·Caja·Regiones·Activación·Origen·Navegación |
| **shot7** | `mockups/.shots/mapa-dogfood.png` | Toggle a `dev-full-cycle` (dogfood) | 4 lanes (spec·build·review·release); 5 nodos; 4 cajas; **3 edges invoca renderizados** (4 en datos: el `lee` spec-writer→std-spec queda oculto, endpoint en Reglas colapsada); Guardia `— sin hooks —` |

> **Redline / anotación** (cómo leer los shots): en shot1/shot2 el orden vertical es Guardia→Proceso→Base;
> las cajas quedan alineadas en la **fila 1** de cada lane (spine horizontal); el borde-izq **punteado** =
> `.puesto`; el badge `PROPUESTO` (ámbar) en `índice semántico`. En shot3 el nodo enfocado lleva **ring**
> y los no-vecinos `opacity:.4`; las 2 líneas **ámbar punteadas** son `lee`. shot4→shot5 es el **mismo
> encuadre de la región Base** (z=1, oy −432) con la subband Reglas plegada→desplegada (before/after
> genuino; ambas son capturas distintas del overview, no duplicados). shot7 muestra el arnés dogfood real
> (lo que el MVP servirá).

---

## 1 · Tabla completa de tokens (SSOT `web/tokens/base.tokens.json`)

Los componentes consumen **la var corta** (`--c-skill`, `--ok`, `--primary`…), **nunca** `--color-kind-skill`
ni el hex crudo (verificado en `tokens.ts`, ver [`architecture.md §6a/§1`](./architecture.md)). El mockup
inlinea estos MISMOS valores en `:root` (`mockup:10-20` light) y los duplica en
`@media (prefers-color-scheme:dark)` (`mockup:21-27`) + `:root[data-theme]` (`mockup:28-41`).

### 1.1 · `color.kind` — tipo de componente (11 vars `--c-*`)

| Var CSS | Token DTCG | Light | Dark | `base.tokens.json` | mockup |
|---------|-----------|-------|------|--------------------|--------|
| `--c-skill` | `color.kind.skill` | `#2f74b5` | `#5b9dd9` | `:167-170` | `:14` |
| `--c-agent` | `color.kind.agent` | `#7455c2` | `#9b7fd4` | `:171-174` | `:14` |
| `--c-hook` | `color.kind.hook` | `#bb4c85` | `#d46a9e` | `:175-178` | `:14` |
| `--c-knowledge` | `color.kind.knowledge` | `#20826f` | `#4fb3a1` | `:179-182` | `:14` |
| `--c-mcp` | `color.kind.mcp` | `#586c80` | `#8195a9` | `:183-186` | `:14` |
| `--c-rule` | `color.kind.rule` | `#7a7a24` | `#b5b558` | `:187-190` | `:14` |
| `--c-command` | `color.kind.command` | `#0f7f96` | `#49aec7` | `:191-194` | `:14` |
| `--c-plugin` | `color.kind.plugin` | `#9c7b1a` | `#cdae54` | `:195-198` | `:14` |
| `--c-settings` | `color.kind.settings` | `#766753` | `#a4917c` | `:199-202` | `:14` |
| `--c-output-style` | `color.kind.output-style` | `#a85274` | `#cc7d9c` | `:203-206` | `:14` |
| `--c-statusline` | `color.kind.statusline` | `#5f7a52` | `#8aa87d` | `:207-210` | `:14` |

> ⚠️ **Deuda honesta** (no altera el look objetivo, sí el estado del código): los 5 últimos
> (`command…statusline`) están en el **source** `base.tokens.json` pero **no regenerados** a
> `web/src/app/styles/theme.css`/`tokens.ts`; hoy `kind.ts` cae a `--muted-foreground`. El mockup NO
> tiene esa deuda (los inlinea). Cerrar el gap = `architecture.md §6a`.

### 1.2 · `color.semantic` — neutrales + marca (nombres shadcn)

| Var CSS | Light | Dark | `base.tokens.json` | mockup |
|---------|-------|------|--------------------|--------|
| `--background` | `#f2f4f7` | `#14171c` | `:41-44` | `:11,22` |
| `--card` | `#ffffff` | `#1b2027` | `:49-52` | `:11,29` |
| `--secondary` | `#e9edf2` | `#242b34` | `:73-76` | `:11,29` |
| `--foreground` | `#212832` | `#e8ecf1` | `:45-48` | `:12,23` |
| `--muted-foreground` | `#5d6a79` | `#94a1af` | `:85-88` | `:12,23` |
| `--border` | `#d5dbe3` | `#2d3743` | `:109-112` | `:13,24` |
| `--input` | `#c3cbd6` | `#3a4655` | `:113-116` | `:13,24` |
| `--primary` | `#a8742c` | `#d9a35b` | `:65-68` | `:13,24` |

### 1.3 · `color.health` — semántica de edges

| Var CSS | Light | Dark | `base.tokens.json` | mockup |
|---------|-------|------|--------------------|--------|
| `--ok` (escribe) | `#2e8f5e` | `#57b380` | `:214-217` | `:15,26` |
| `--warn` (lee) | `#c96a2e` | `#e0904f` | `:218-221` | `:15,26` |
| `--crit` (invoca) | `#c94545` | `#e25b5b` | `:222-225` | `:15,26` |

### 1.4 · Escalas no-color

| Grupo | Valores | `base.tokens.json` | mockup |
|-------|---------|--------------------|--------|
| `radius` | `--radius-md:8px` · `--radius-lg:9px` · `--radius-full:9999px` (source también `sm:6` `xl:12`) | `:247-255` | `:16` |
| `text` | `--text-xs:11px` · `--text-sm:12.5px` · `--text-base:14px` | `:269-278` | `:17` |
| `font` | `--font-sans:system-ui,-apple-system,"Segoe UI",Roboto,sans-serif` · `--font-mono:ui-monospace,"SFMono-Regular",Menlo,Consolas,"Liberation Mono",monospace` | `:279-300` | `:18-19` |

> **Radios adicionales usados como literal en el mockup** (a snap-ear a token en el port, hoy magic-value
> en el HTML): `radius:14px` en `.region` (`mockup:54`) y `border-radius:6px` en `.node-cmd` (`mockup:112`,
> único uso). → nota de deuda en `architecture.md §5.4`: mapear a `--radius-xl` (12) / `--radius-sm` (6)
> o añadir `--radius-2xl`. En este doc se documentan como **valor del mockup firmado**, no como invención.

---

## 2 · Layout global (viewport · stage · content · regiones)

Estructura DOM del lienzo (`mockup:148-200`): `.app > .viewport#viewport > .stage#stage >
.content#content` + `svg.edges#edges` overlay dentro de `.content`.

| Elemento | Reglas exactas | mockup |
|----------|----------------|--------|
| `body` (reset) | `*{box-sizing:border-box}` (`:43`) · `body{margin:0; background:var(--background); color:var(--foreground); font-family:var(--font-sans); font-size:var(--text-base); -webkit-font-smoothing:antialiased}` | `:43-44` |
| `.app` | `position:relative; height:100vh; overflow:hidden` | `:46` |
| `.viewport` | `position:absolute; inset:0; overflow:hidden; cursor:grab; background:var(--background)` — `.grabbing`→`cursor:grabbing` | `:47-48` |
| `.stage` | `position:absolute; top:0; left:0; transform-origin:0 0; will-change:transform` — transform = `translate(ox,oy) scale(z)` | `:49` |
| `.content` | `position:relative; display:flex; flex-direction:column; gap:18px; padding:24px` | `:50` |
| `svg.edges` | `position:absolute; inset:0; width:100%; height:100%; pointer-events:none; z-index:3; overflow:visible` | `:51` |

**Overview-first medido (shot1):** al abrir, `fit()` → `transform: translate(24px,24px) scale(0.906)` con
viewport 1680×1000 y content 1802×979. El stage vive **dentro** de la transform escalada; por eso los
edges dividen sus coords por `z` (§9). SB: `widgets/map-canvas/ui/map-canvas.tsx`.

---

## 3 · Región (3 territorios con tinte propio)

Tres bloques `.region` (`mockup:54-60`, construidos `mockup:406-446`). Cada región = tinte
`color-mix` sobre `--card` + header con regla horizontal. Screenshots: shot1 (light) / shot2 (dark).

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.region` (base) | `display:flex; flex-direction:column; gap:10px; padding:14px 16px; border-radius:14px; border:1px solid transparent` | `:54` |
| `.region-hd` | `display:flex; align-items:center; gap:10px` | `:58` |
| `.region-hd h2` | `margin:0; font-size:var(--text-xs); font-weight:700; letter-spacing:.08em; text-transform:uppercase; color:var(--muted-foreground)` | `:59` |
| `.region-hd .rule` | `flex:1; height:1px; background:var(--border)` | `:60` |

### 3.1 · Tinte + título por región

| Región | `background` | `border-color` | Título (texto exacto) | mockup |
|--------|-------------|----------------|------------------------|--------|
| `.r-guardia` | `color-mix(in srgb, var(--c-hook) 6%, var(--card))` | `color-mix(in srgb, var(--c-hook) 16%, var(--border))` | «Guardia · hooks transversales» | `:55` / `:408` |
| `.r-proceso` | `color-mix(in srgb, var(--primary) 5%, var(--card))` | `color-mix(in srgb, var(--primary) 14%, var(--border))` | «Proceso · carriles por fase» | `:56` / `:417` |
| `.r-soporte` | `color-mix(in srgb, var(--c-knowledge) 6%, var(--card))` | `color-mix(in srgb, var(--c-knowledge) 16%, var(--border))` | «Base · conocimiento, reglas y soporte del arnés» | `:57` / `:440` |

> Nota semántica (ver `spec.md §3`): la clase CSS conserva el nombre histórico `r-soporte` pero el
> **título** y la doctrina son **«Base»** (canónico VISION A6, `VISION.md:81-83`). SB nuevo: `region`
> (no existe hoy — `architecture.md §6e`).

---

## 4 · Glyph (forma + color + char por clase)

Primitivo `.glyph` (`mockup:116-120`) — cuadro de **17×17px**, fondo = color de la clase (`--tc`),
letra en color `--background` (knockout). SB: `web/src/shared/canvas/glyph.tsx` (**hoy 15px** → delta
`architecture.md §6b`; los clip-paths son idénticos).

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| caja | `width:17px; height:17px` | `:116` |
| layout | `display:grid; place-items:center; flex:none` | `:116` |
| fondo / texto | `background:var(--tc); color:var(--background)` | `:116` |
| fuente | `font-family:var(--font-mono); font-weight:700; font-size:10px; line-height:1` | `:116` |

### 4.1 · Las 6 formas (border-radius / clip-path exactos)

| Forma | Definición CSS | mockup | `glyph.tsx` |
|-------|----------------|--------|-------------|
| `square` | `border-radius:3px` | `:117` | `:11` |
| `circle` | `border-radius:50%` | `:117` | `:12` |
| `rounded` | `border-radius:4px` | `:117` | `:13` |
| `diamond` | `clip-path:polygon(50% 0,100% 50%,50% 100%,0 50%)` | `:118` | `:14` |
| `hexagon` | `clip-path:polygon(25% 0,75% 0,100% 50%,75% 100%,25% 100%,0 50%)` | `:119` | `:15` |
| `shield` | `clip-path:polygon(0 0,100% 0,100% 62%,50% 100%,0 62%)` | `:120` | `:16` |

### 4.2 · Tabla clase → {forma · char · color · label}

Mapa `KIND` (`mockup:203-214`); SB: `web/src/entities/arnes/model/kind.ts:18-37`.

| Clase | Forma | Char | Color (var) | Label | mockup |
|-------|-------|------|-------------|-------|--------|
| `skill` | square | `S` | `--c-skill` | Skill | `:204` |
| `subagent` | circle | `A` | `--c-agent` | Subagente | `:205` |
| `hook` | diamond | `H` | `--c-hook` | Hook | `:206` |
| `rule` | shield | `R` | `--c-rule` | Regla | `:207` |
| `mcp` | hexagon | `M` | `--c-mcp` | MCP | `:208` |
| `command` | rounded | `/` | `--c-command` | Comando | `:209` |
| `plugin` | rounded | `P` | `--c-plugin` | Plugin | `:210` |
| `settings` | rounded | `⚙` | `--c-settings` | Ajustes | `:211` |
| `output-style` | rounded | `◐` | `--c-output-style` | Output style | `:212` |
| `statusline` | rounded | `▭` | `--c-statusline` | Statusline | `:213` |

> Las 5 clases config comparten forma `rounded`; la **sub-distinción la carga el char + el color**
> `--c-*` propio (por eso el gap 6→10 debe cerrarse). La leyenda del panel de ayuda deduplica por forma
> (`mockup:511`) → muestra 6 glyphs (shot6).

---

## 5 · Nodo — anatomía base

`.node` = `<button>` (`mockup:91`, construido `mockup:336-351`). SB: `entities/arnes/ui/arnes-node.tsx`.

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| layout | `position:relative; display:flex; flex-direction:column; gap:5px; width:100%; text-align:left; cursor:pointer` | `:91` |
| borde | `border:1px solid var(--border); border-left:3px solid var(--tc,var(--border))` | `:91` |
| radio / fondo | `border-radius:var(--radius-md)` (8px) `; background:var(--secondary)` | `:91` |
| padding / fuente | `padding:10px 12px; font:inherit; color:var(--foreground)` | `:91` |
| transición | `transition:border-color .12s, box-shadow .12s` | `:91` |

### 5.1 · Sub-elementos

| Sub-elemento | Reglas | mockup |
|--------------|--------|--------|
| `.node-top` | `display:flex; align-items:flex-start; gap:7px; padding-right:40px` (deja sitio a los badges) | `:109` |
| `.node-nm` | `font-family:var(--font-mono); font-size:var(--text-sm); font-weight:600; color:var(--foreground); line-height:1.3; -webkit-line-clamp:2` (clamp a **2 líneas**) `; overflow:hidden` | `:110` |
| `.node-ty` | (definido pero no usado en el nodo — el tipo lo dicen forma+color+char+handle) `font-size:var(--text-xs); text-transform:uppercase; letter-spacing:.04em; color:var(--muted-foreground)` | `:111` |
| `.node-cmd` (handle) | `align-self:flex-start; font-family:var(--font-mono); font-size:var(--text-xs); color:var(--foreground); background:var(--card); border:1px solid var(--border); border-radius:6px; padding:1px 7px; ellipsis; white-space:nowrap` | `:112` |
| `.node-trans` (transición spine) | `align-self:flex-start; inline-flex; gap:4px; font-family:var(--font-mono); font-size:9px; color:var(--tc); background:color-mix(in srgb,var(--tc) 8%,transparent); border:1px solid color-mix(in srgb,var(--tc) 30%,var(--border)); border-radius:var(--radius-full); padding:0 7px; margin-top:1px` — `::before` = `"◇"` 8px opacity .8 | `:114-115` |

### 5.2 · Handle derivado por clase — `comando(box)`

El handle (`.node-cmd`) **no es un dato**: se deriva de la clase (`mockup:327-334`). Regla:

| Clase | Handle | mockup |
|-------|--------|--------|
| `command` | `box.id` (ya trae `/`) | `:328` |
| `skill` | `"/" + box.id` (user-invocable) | `:329` |
| `subagent` | `"@" + box.id` (delegación) | `:330` |
| `hook` | `"evento"` | `:331` |
| `rule` | `box.alw===false ? "condicional" : "always-on"` | `:332` |
| resto (mcp/plugin/settings/output-style/statusline) | `box.id` | `:333` |

Ejemplos en shot1: `/need-distiller`, `@discovery-interviewer`, `/spec-review`, `api-mcp`.

---

## 6 · Variantes del nodo (deltas exactos)

Cada variante = clase CSS aditiva sobre `.node`. Screenshots: caja/puesto/prop en shot1; compact en
shot1 (Guardia); support+hairline en shot1 (lanes); dim+hover en shot3.

### 6.1 · `.caja` — skill-frente de la fase (`contract.caja`)

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.node.caja` | `border-left-width:5px; background:color-mix(in srgb,var(--tc) 9%,var(--secondary)); box-shadow:inset 0 0 0 1px color-mix(in srgb,var(--tc) 22%,transparent)` | `:104` |
| `.caja-badge` | `position:absolute; top:8px; right:8px; font-family:var(--font-mono); font-size:9px; font-weight:700; letter-spacing:.06em; text-transform:uppercase; color:var(--tc); border:1px solid var(--tc); border-radius:var(--radius-full); padding:0 6px; opacity:.9` — texto «caja» | `:105` / `:343` |

### 6.2 · `.puesto` — facet `origen` = del-puesto (PROPUESTA)

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.node.puesto` | `border-left-style:dashed` (borde-izq punteado, sutil) | `:107` |
| `.prop-badge` | `position:absolute; top:8px; right:8px; font-family:var(--font-mono); font-size:9px; font-weight:700; letter-spacing:.05em; text-transform:uppercase; color:var(--warn); border:1px dashed var(--warn); border-radius:var(--radius-full); padding:0 6px; opacity:.85` — texto «propuesto» | `:108` / `:344` |

Los 13 ids `del-puesto` (borde punteado en shot1) vienen del set `DEL_PUESTO` (`mockup:229`).
`prop-badge` sólo en `indice-semantico` (`box.prop`, `mockup:278`). Ambos son PROPUESTA (no hay campo
L0 hoy — `spec.md §2` / `architecture.md §5`).

### 6.3 · `.support` — apoyo subordinado a la caja

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.node.support` | `margin-left:14px; background:color-mix(in srgb,var(--secondary) 55%,transparent)` (indentado + fondo más tenue) | `:95` |
| `.lane-sep` (hairline antes del 1.er apoyo) | `height:1px; background:var(--border); margin:3px 12px 1px 14px; opacity:.7` | `:96` |

El apoyo se marca con **indent + hairline**, sin la palabra «apoyo» (decisión del operador). Regla de
asignación: nodo de fase que **no** es caja, cuando la fase tiene alguna caja (`mockup:427`).

### 6.4 · `.compact` — hooks de Guardia (chip de 1 línea)

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.node.compact` | `width:auto; flex-direction:row; align-items:center; gap:8px; padding:7px 12px` | `:99` |
| `.node.compact .node-top` | `padding-right:0` (sin reserva de badges, ocultos) | `:100` |
| `.node.compact .node-nm` | `-webkit-line-clamp:1; white-space:nowrap; line-height:1.2` | `:101` |
| oculta | `.node-cmd, .node-trans, .caja-badge, .prop-badge → display:none` | `:102` |
| fila contenedora | `.guardia-row{align-items:center}` | `:98` |

### 6.5 · `.dim` + hover (foco de atención)

| Propiedad | Valor | mockup |
|-----------|-------|--------|
| `.node.dim` | `opacity:.4` (nodos no relacionados al enfocar) | `:93` |
| `.node:hover` | `border-color:var(--tc,var(--input)); box-shadow:0 0 0 2px color-mix(in srgb,var(--tc) 22%,transparent)` | `:92` |

En shot3 el nodo enfocado (`builder`) recibe el ring `:hover` y sus vecinos quedan legibles; los demás
`.dim`. El `.dim` se aplica/quita en `drawEdges` (`mockup:490-496`).

---

## 7 · Región Proceso — carriles por fase

Contenedor `.lanes` (`mockup:84-89`, construido `mockup:416-435`). SB: `widgets/map-canvas/ui/lane.tsx`
(**hoy `min-w-44 flex-1`** = 176px que crece → delta a `width:232px` fijo, `architecture.md §6d`).

| Elemento | Reglas | mockup |
|----------|--------|--------|
| `.lanes` | `display:flex; align-items:flex-start; gap:16px` | `:84` |
| `.lane` | `flex:none; width:232px; display:flex; flex-direction:column; border:1px solid var(--border); border-radius:var(--radius-lg)` (9px) `; background:var(--card)` | `:85` |
| `.lane-hd` | `display:flex; align-items:center; justify-content:space-between; gap:8px; border-bottom:1px solid var(--border); padding:9px 12px` | `:86` |
| `.lane-hd h3` | `margin:0; font-size:var(--text-sm); font-weight:600` (nombre de fase) | `:87` |
| `.count` (pill) | `font-family:var(--font-mono); font-size:var(--text-xs); color:var(--muted-foreground); border:1px solid var(--border); background:var(--secondary); border-radius:var(--radius-full); padding:1px 7px` | `:88` |
| `.lane-body` | `display:flex; flex-direction:column; gap:10px; padding:12px` | `:89` |

**Orden caja-first** (`mockup:424`): dentro de cada lane, la(s) caja(s) se ordenan primero
(`sort (b.caja?1:0)-(a.caja?1:0)`) para que las cajas queden alineadas en la **fila 1** = spine
horizontal (visible en shot1: la fila superior de cajas cruza las 7 lanes). Debajo, el `.lane-sep`
hairline y luego los `.support` indentados.

Lanes = una columna por `arnes.fases[]` en orden declarado, más las fases no declaradas que aparezcan en
nodos (append, primera aparición) — `selFases` `mockup:396-400`. Luana → 7 lanes; dogfood → 4 (shot7).

---

## 8 · Región Base — reglas colapsables · knowledge · bandas

Base (`.r-soporte`) itera `SUPPORT_BANDS` (`mockup:218-224`); la banda `base` usa `baseBand()`, las
demás `bandBlock()` (`mockup:441-444`). SB nuevos: `base-band`/`rules-subband`/`activation-chip`
(`architecture.md §6e`).

### 8.1 · `SUPPORT_BANDS` (orden + activación + label)

| id | Label (texto exacto) | `act` | Colapsable | mockup |
|----|----------------------|-------|-----------|--------|
| `base` | «Base · reglas y knowledge» | `mixta` | **sí** | `:219` |
| `libreria-expertos` | «Librería de expertos» | `demanda` | no | `:220` |
| `meta-harness` | «Meta-harness · config del arnés» | `demanda` | no | `:221` |
| `terceros` | «Terceros» | `demanda` | no | `:222` |
| `marcas-dormidas` | «Marcas dormidas» | `dormida` | no | `:223` |

Sólo se renderiza la banda si tiene nodos (`mockup:443`). Luana no trae `marcas-dormidas` (0 nodos) →
no aparece; sí aparecen libreria/meta/terceros (shot1).

### 8.2 · `.band` / `.band-hd` / chip de activación

| Elemento | Reglas | mockup |
|----------|--------|--------|
| `.band` | `display:flex; flex-direction:column; gap:8px` | `:62` |
| `.band.dormida` | `opacity:.55` | `:63` |
| `.band-hd` | `display:flex; align-items:center; gap:8px` | `:64` |
| `.band-hd h3` | `margin:0; font-size:var(--text-sm); font-weight:600` | `:65` |
| `.band-hd span` + `.count` | `.band-hd span{font-size:var(--text-xs); color:var(--muted-foreground)}` (`:66`) + pill `.count` (mismo del lane, definido `:88`) | `:66,88` |
| `.act-chip` | `font-family:var(--font-mono); font-size:9px; font-weight:600; letter-spacing:.02em; color:var(--ac); border:1px solid color-mix(in srgb,var(--ac) 45%,transparent); background:color-mix(in srgb,var(--ac) 11%,transparent); border-radius:var(--radius-full); padding:1px 7px; white-space:nowrap` | `:68` |

### 8.3 · Mapa de activación `ACT` (label + tono `--ac`)

| Clave | Label | Tono (`--ac`) | mockup |
|-------|-------|---------------|--------|
| `siempre` | «siempre en contexto» | `var(--crit)` | `:232` |
| `condicional` | «carga condicional» | `var(--warn)` | `:233` |
| `demanda` | «bajo demanda» | `var(--muted-foreground)` | `:234` |
| `leido` | «leído por skills» | `var(--c-knowledge)` | `:235` |
| `dormida` | «dormida · 0 corridas» | `var(--muted-foreground)` | `:236` |
| `mixta` | «» (sin chip) | `var(--muted-foreground)` | `:237` |

### 8.4 · `baseBand()` — subband Reglas colapsable + Knowledge

`baseBand` (`mockup:363-394`; contenedor `.band.base-band{gap:10px}` `mockup:70`) parte los nodos:
`rules` (clase `rule`, split `alw`) + `rest` (no-rule).

**Subband Reglas** (colapsable, `collapsed` por defecto — shot4):

| Elemento | Reglas | mockup |
|----------|--------|--------|
| `.subband` | `border:1px solid var(--border); border-radius:var(--radius-md)` (8px) `; background:color-mix(in srgb,var(--secondary) 38%,transparent)` | `:71` |
| `.subband-hd` (botón) | `width:100%; display:flex; align-items:center; gap:8px; padding:8px 12px; background:none; border:0; color:var(--foreground); font:inherit; cursor:pointer; text-align:left` | `:72` |
| `.subband-hd h4` | `margin:0; font-size:var(--text-sm); font-weight:600` — texto «Reglas» | `:74` |
| `.tw` (flecha) | `font-size:11px; color:var(--muted-foreground); width:10px` — `▸` colapsado / `▾` abierto | `:75,384` |
| chips del header | `.count` (total) + `.act-chip --crit` «N siempre» + `.act-chip --warn` «N condicional» | `:369-372` |
| `.subband-body` | `display:flex; flex-direction:column; gap:10px; padding:2px 12px 12px` | `:76` |
| `.subband.collapsed .subband-body` | `display:none` | `:77` |

**Rule-groups dentro del body** (shot5, expandido):

| Elemento | Reglas | mockup |
|----------|--------|--------|
| `.rule-grp` | `display:flex; flex-direction:column; gap:6px` | `:78` |
| `.rule-grp-hd` | `font-size:9px; font-weight:700; text-transform:uppercase; letter-spacing:.07em; color:var(--ac); opacity:.85` | `:79` |
| grupo 1 | «siempre en contexto (CLAUDE.md)», tono `--crit`, nodos con `alw:true` | `:378` |
| grupo 2 | «carga condicional (paths:)», tono `--warn`, nodos con `alw:false` | `:379` |
| `.row` / `.node-w` | `.row{display:flex; flex-wrap:wrap; gap:10px}` · `.row .node-w{width:232px}` | `:80-81` |

**Subband Knowledge & servicios** (estática, no colapsable — shot1):

| Elemento | Reglas | mockup |
|----------|--------|--------|
| `.subband-hd.static` | `cursor:default` (sin toggle) | `:73` |
| header | h4 «Knowledge & servicios» + `.count` + `.act-chip --c-knowledge` «leído por skills» | `:390` |

### 8.5 · `bandBlock()` — bandas simples (libreria/meta/terceros/dormidas)

`bandBlock` (`mockup:355-361`): `.band-hd` (label + act-chip + count) + `.row` de nodos `.node-w`
(232px). `marcas-dormidas` añade `.dormida` (`opacity:.55`). `.empty` (`font-size:var(--text-xs);
font-style:italic; color:var(--muted-foreground)`, `mockup:82`) para bandas/guardia vacías
(«— sin hooks —» en dogfood, shot7).

---

## 9 · Edges (SVG overlay)

`drawEdges` (`mockup:460-497`). SB: `widgets/map-canvas/ui/edge-layer.tsx` + `model/use-edge-paths.ts`
(**hoy colores neutros `--input`/`--primary`, sin gating** → delta `architecture.md §6f`).

### 9.1 · Estilo por tipo — `STROKE`

| Tipo | Color | Opacity base | Dash | Flecha | mockup |
|------|-------|--------------|------|--------|--------|
| `invoca` (spine/flujo) | `var(--crit)` | `.75` | — (sólida) | sí (`marker-end`) | `:240` |
| `escribe` | `var(--ok)` | `.7` | `3 3` | sí | `:241` |
| `lee` | `var(--warn)` | `.65` | `4 4` | **no** | `:242` |

Marcador flecha `#arr` (`mockup:154-156`): `viewBox 0 0 10 10; refX 8; refY 5; markerWidth 6;
markerHeight 6; orient auto-start-reverse; fill context-stroke`.

### 9.2 · Geometría (bezier cúbico)

- Anclas: de = borde **derecho** medio del nodo `de`; a = borde **izquierdo** medio del nodo `a`
  (`getBoundingClientRect`, coords divididas por `z`, `mockup:473-474`).
- `dx = Math.max(30, |x2-x1|/2)` (`mockup:475`).
- path: `M x1,y1 C x1+dx,y1 x2-dx,y2 x2,y2` (`mockup:478`).
- ancho base `1.6` (`mockup:483`).

### 9.3 · Regla spine-always + hover-reveal + foco

| Situación | Comportamiento | mockup |
|-----------|----------------|--------|
| Sin foco | sólo se dibujan edges `invoca` (backbone); `lee`/`escribe` ocultos | `:465-467` |
| Hover en nodo | se dibujan además los `lee`/`escribe` que **tocan** el nodo enfocado | `:465` |
| Edge que toca el foco | `opacity:.95; stroke-width:2.2` | `:484` |
| Spine no-tocado (con foco activo) | `opacity:.2` (se atenúa el backbone) | `:484` |
| Endpoint colapsado/oculto (reglas plegadas) | no se dibuja (`!f.width||!t.width` → skip) | `:471` |

En shot3 (hover `builder`): 11 edges spine (atenuados) + 2 `lee` ámbar (builder→api-mcp, builder→
react-expert) resaltados. En shot1 (sin foco): sólo 11 edges `invoca` rojos.

---

## 10 · Controles flotantes

| Control | Reglas | mockup |
|---------|--------|--------|
| `.ctl` (base) | `position:absolute; display:flex; gap:6px; z-index:20` (heredado por zoom+example) | `:123` |
| `.ctl.zoom` | `left:16px; bottom:16px; flex-direction:column` (+ `.ctl` base) — botones `+` / `−` / `⤢` | `:124,163-167` |
| `.ctl button` | `width:36px; height:36px; display:grid; place-items:center; border:1px solid var(--border); background:var(--card); color:var(--foreground); border-radius:var(--radius-md); box-shadow:0 1px 3px rgba(0,0,0,.12)` | `:126` |
| `.ctl.example` | posición desde `.ctl` (`:123`) + `right:16px; top:16px` (`:125`) + `border:1px solid var(--border); background:var(--card); border-radius:var(--radius-md); box-shadow:0 1px 3px rgba(0,0,0,.12); padding:3px; gap:2px` (`:127`) — botones «Luana» / «dev-full-cycle» | `:123,125,127,168-171` |
| `.ctl.example button` | `padding:5px 11px; font-size:var(--text-xs); border:0; background:transparent; color:var(--muted-foreground)` | `:128` |
| `…button[aria-pressed=true]` | `background:var(--secondary); color:var(--foreground); font-weight:600` | `:129` |
| `.help-fab` | `position:absolute; right:16px; bottom:16px; width:40px; height:40px; border-radius:var(--radius-full); border:1px solid var(--border); background:var(--card); color:var(--foreground); font-size:18px; box-shadow:0 2px 6px rgba(0,0,0,.18)` — «?» | `:132,173` |
| `.help-panel` | `position:absolute; right:16px; bottom:64px; width:300px; max-height:76vh; overflow:auto; border:1px solid var(--border); background:var(--card); border-radius:var(--radius-lg); box-shadow:0 8px 24px rgba(0,0,0,.22); padding:16px` — `[hidden]→display:none` | `:133-134` |
| `.help-panel` (tipografía) | `h3{font-size:var(--text-sm); font-weight:650}` (`:135`) · `h4{font-size:var(--text-xs); text-transform:uppercase; letter-spacing:.06em; color:var(--muted-foreground)}` (`:136`) · `p,li{font-size:var(--text-sm); color:var(--muted-foreground)}` (`:137`) · `ul{padding-left:16px}` (`:138`) · `b{color:var(--foreground)}` (`:139`) | `:135-139` |
| leyenda (`.lg-*`) | `.lg-row{display:flex; align-items:center; gap:8px}` (`:140`) · `.lg-line{width:30px; border-top:2px solid var(--crit)}` · `.lg-line.lee{2px dashed var(--warn)}` · `.lg-line.escribe{2px dashed var(--ok)}` (`:141-143`) · `.lg-glyphs{display:flex; flex-wrap:wrap; gap:9px}` (`:144`) · `.lg-g{display:flex; align-items:center; gap:5px; font-size:var(--text-xs)}` (`:145`) | `:140-145` |

**Panel de ayuda** (shot6) — headings + contenido (`mockup:174-199`): título «ArnesIA · Mapa — MVP
(Hito 1)» · «Relaciones» (leyenda de líneas invoca/escribe/lee, `.lg-line`/`.lg-line.escribe`/`.lg-line.lee`
`mockup:141-143`) · «Clases (forma + color)» (6 glyphs deduplicados por forma) · «Caja de proceso» ·
«Regiones» · «Activación (propuesta)» · «Origen (propuesta)» · «Navegación». Todo el texto declara
explícitamente qué es **propuesta** (activación/origen) vs canónico (Base).

---

## 11 · Estados — matriz light + dark

Cada estado del Mapa se ve idéntico en ambos temas salvo el swap de tokens (§1). El theming es por
`prefers-color-scheme` (`mockup:21-27`) **y** `:root[data-theme]` (`mockup:28-41`) — doble mecanismo,
mismos valores.

| Estado | Screenshot light | Screenshot dark | Fuente de estado |
|--------|------------------|-----------------|------------------|
| Overview (fit) | shot1 | shot2 | `fit()` al abrir (`mockup:517`) |
| Hover nodo (dim + lee/escribe) | shot3 | *(idéntico, swap tokens)* | `mouseover` (`mockup:521`) |
| Reglas colapsadas | shot4 | *(idéntico)* | `.subband.collapsed` default (`mockup:368`) |
| Reglas expandidas | shot5 | *(idéntico)* | click `.subband-hd` (`mockup:381`) |
| Panel ayuda | shot6 | *(idéntico)* | click `.help-fab` (`mockup:513`) |
| Dogfood (arnés real) | shot7 | *(idéntico)* | toggle `dev-full-cycle` (`mockup:515`) |

> **Confirmación de paridad de tema:** medido `--background` = `#f2f4f7` (light, shot1) y `#14171c`
> (dark, shot2); `--c-skill` = `#2f74b5` (light) / `#5b9dd9` (dark). Ambos = valores de §1.

---

## 12 · No-magic-value — auditoría

Todo color del mockup es `var(--…)` o `color-mix(in srgb, var(--…) N%, …)`. **Únicos literales**
presentes (a resolver en el port, declarados como valor del mockup firmado — no invención):

| Literal | Dónde | Resolución propuesta (`architecture.md §5.4`) |
|---------|-------|---------------------------------------------|
| `radius:14px` | `.region` `mockup:54` | token `--radius-xl` (12) o nuevo `--radius-2xl` |
| `border-radius:6px` | `.node-cmd` `mockup:112` (único uso) | token `--radius-sm` (6, ya en source) |
| `font-size:9px` / `10px` / `11px` | badges, glyph, `.tw` | escala micro (bajo `--text-xs`); documentar o añadir `--text-2xs` |
| `box-shadow rgba(0,0,0,.12/.18/.22)` | controles/paneles | token de sombra (no existe hoy; deuda) |
| `margin-left:14px`, `gap` 5/6/7/8/10/16/18px | spacings | snap a `space` base-4 (`--space-*`) donde encaje |

Estos son **medidas**, no colores; la regla anti-magic-value dura (stylelint) aplica a **color**
(`hs-norma-pegarse-storybook`). Los px se documentan aquí con su línea para reproducir el pixel exacto.

---

## Apéndice · Índice cruzado

- Comportamiento (el QUÉ) de cada elemento → [`spec.md`](./spec.md).
- Deltas componente↔mockup, plan de sync Storybook, tokens a regenerar → [`architecture.md §6`](./architecture.md).
- Doctrina de Base/activación/origen → [`spec.md §2`](./spec.md) + [`architecture.md §5`](./architecture.md).
