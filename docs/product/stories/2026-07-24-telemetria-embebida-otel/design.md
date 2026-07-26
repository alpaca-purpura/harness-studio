# Diseño de UI · Capa «Mejora» del Mapa + tarjeta del Portafolio

> `tipo: design` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> Par de [`spec.md`](./spec.md) (RF-232…RF-286). **El UI al pixel**: orden de secciones, jerarquía,
> tabla de campos, tokens por marca, todos los estados por componente, responsive y **copy exacto**.
> El QUÉ está en el spec; acá está el CÓMO SE VE.
>
> **Cero color literal.** Todo `var(--…)` de [`web/src/app/styles/theme.css`](../../../../web/src/app/styles/theme.css)
> (generado de `web/tokens/base.tokens.json`). `stylelint` con `strict-value` lo enforça y está
> dentro de `pnpm --dir web run verify`.
>
> **SSoT del UI vigente = Storybook**, no el `.html`. Lo que se calca sale de
> `map-bar.tsx` · `map-canvas.tsx` · `lane.tsx` · `arnes-node.tsx` · `inspector.tsx` ·
> `portafolio-list.tsx`, y de sus `.css` scopeados. El mockup es la propuesta; el código vigente es
> la línea base.

## 0 · Cómo leer esto

| sección | qué resuelve |
|---|---|
| §1 | Inventario: qué se reusa tal cual · qué se extiende · qué nace |
| §2 | Superficie por superficie, al pixel |
| §3 | Tabla de campos (etiqueta · fuente · formato · qué se ve cuando falta) |
| §4 | Tokens por marca |
| §5 | Todos los estados, por componente |
| §6 | Responsive y desbordamiento |
| §7 | Copy exacto, todo junto |
| §8 | Gates de lint y a11y |

**Cinco reglas duras que gobiernan cada decisión de abajo:**

1. **Overlay, no rediseño.** La capa Mejora **agrega**; la geografía del canvas no se toca (RF-245).
2. **Solo las cajas llevan cifra.** Todo lo demás dice por qué no (RF-238/RF-243).
3. **Cada número dice cómo se atribuyó** — cuatro casos, cuatro copys (RF-242).
4. **Ningún estado depende solo del color** (RF-280).
5. **El copy es material de diseño.** Está en §7, literal, y el test asserta el literal.

---

## 1 · Inventario de componentes

### 1.1 Se reusa TAL CUAL (cero cambio)

| Componente | Path | Por qué |
|---|---|---|
| `MapCanvas` | `web/src/widgets/map-canvas/ui/map-canvas.tsx` | La geografía no cambia. La capa entra por prop, no por rama de render |
| `Region` · `Band` · `BaseBand` · `HandoffGutter` · `EdgeLayer` | `web/src/widgets/map-canvas/ui/` | Nada de la capa Mejora los toca |
| `Glyph` | `web/src/shared/canvas` | El tipo de nodo sigue siendo el mismo |
| `useViewport` | `web/src/widgets/map-canvas/model/use-viewport.ts` | El pan/zoom es el mismo — **y por eso el canvas NO scrollea** (§6.2) |
| `ErrorBoundary` | `web/src/shared/ui/error-boundary` | Ya envuelve el canvas; cubre también los selectores nuevos |
| `DotSaludPortafolio` · `EmblemaInicial` | `web/src/entities/portafolio/ui/chips.tsx` | La fila del Portafolio conserva su identidad visual |
| `Skeleton` · `ErrorBody` | privados en `web/src/widgets/portafolio/ui/portafolio-list.tsx` | Patrón de carga/error ya firmado — se **promueve** a `shared/ui/` para que el Mapa lo use (§1.4) |
| `trapTabKeyDown` | `web/src/shared/lib/focus-trap.ts` | El diálogo de confirmación de borrado (RF-275) |
| `cn` | `web/src/shared/lib/cn.ts` | — |

### 1.2 Se EXTIENDE (superset, nada se quita)

| Archivo | Qué cambia | RF |
|---|---|---|
| `web/src/widgets/map-canvas/model/layers.ts` | `{ id: "tokens", label: "Tokens", disabled: true }` → `{ id: "mejora", label: "Mejora" }`. **El `id` del tipo `Capa` pasa de `"tokens"` a `"mejora"`** — es un rename de literal de unión, lo agarra `tsc`. `LayerDef` gana `motivo?: string` para el tooltip honesto | RF-232 · RF-233 |
| `web/src/widgets/map-canvas/ui/map-bar.tsx` | El `title` fijo `"Necesita telemetría (indexer JSONL)"` (línea 123) sale y pasa a leerse de `l.motivo`. Se agrega `aria-describedby` al slot deshabilitado apuntando a un `<span class="sr-only">` con el motivo (RF-276) — hoy el motivo vive solo en `title`, que un lector de pantalla puede no anunciar | RF-233 · RF-276 |
| `web/src/widgets/map-canvas/ui/lane.tsx` | El `lane-hd` gana un tercer hijo: el total de la fase. **El `<span className="count">` se conserva** (J-8) | RF-244 |
| `web/src/entities/arnes/ui/arnes-node.tsx` | Prop opcional `mejora?: CifraCaja`. Cuando llega, el nodo agrega tres hijos **al final del flujo** (cifra · marca de fuga · barra de participación) y un pie. Cuando no llega, el nodo renderiza **exactamente** el DOM de hoy | RF-238…RF-243 |
| `web/src/widgets/map-canvas/ui/inspector.tsx` | `TABS` (línea 111) gana una cuarta entrada; `Tab` gana `"mejora"`; se agrega un cuarto `tabpanel` con el mismo contrato de `id`/`aria-controls`/`aria-labelledby`/`hidden` que los tres vigentes | RF-258 · RF-277 |
| `web/src/pages/shell/ui/workspace-stage.tsx` | Es el dueño del estado `capa` (línea 44) y **el único que hace transporte** (`fe-transporte-independiente`). Gana el estado de ventana y la carga de la telemetría; los widgets reciben props puras | RF-234 |
| `web/src/widgets/portafolio/ui/portafolio-list.tsx` | La `Fila` gana tres celdas opcionales (costo/corrida · tendencia · punto de mejora). Opcionales ⇒ las 25 stories firmadas del Slice 1 **no cambian** | RF-265…RF-268 |
| `web/src/pages/shell/ui/portafolio-view.tsx` | Transporte de la telemetría del Portafolio, mismo patrón lazy que `GET /api/marketplaces` | RF-265 |

### 1.3 NACE — entidad `telemetria`

Capa **entities**: es dominio, no feature. No importa canvas, no importa otra entity, no importa
transporte. **Presenta; no calcula.** El join, los detectores y el contrafactual se resuelven en el
dominio Go y viajan resueltos en el wire — igual que `SituacionCatalogo` en el paquete de
marketplace, y por el mismo motivo: cruzar telemetría × arnés en el FE sería cross-import
entity↔entity y `steiger fsd/no-cross-imports` lo rompe.

```
web/src/entities/telemetria/
  index.ts                       barrel público
  model/types.ts                 Confianza · Ventana · CifraCaja · Cobertura · PuntoMejora ·
                                 Detector · EstadoDetector · BucketToken · ParidadCosto
  model/selectors.ts             usd() · etiquetaConfianza() · tonoConfianza() · aviso de confianza ·
                                 etiquetaDetector() · direccionTendencia()
  model/selectors.test.ts        tabla de casos (colocado, patrón sancionado)
  ui/cifra-usd.tsx               <CifraUSD> — el ÚNICO lugar donde se formatea dinero (RF-281)
  ui/marca-confianza.tsx         <MarcaConfianza> — los 4 casos (RF-242)
  ui/barra-cobertura.tsx         <BarraCobertura> — 4 segmentos + texto (RF-237)
  ui/sparkline.tsx               <Sparkline> — 5 puntos + texto equivalente (RF-266)
  ui/*.stories.tsx               una story por estado, con play()
  testing/telemetria.ts          fixtures con shape REAL, copiadas de verificacion-2026-07-26/evidencia/
```

### 1.4 NACE — widgets

| Widget | Path | Rol |
|---|---|---|
| `FranjaMejora` | `web/src/widgets/map-canvas/ui/franja-mejora.tsx` | La `ctxbar`: ventana · total · disclaimer · cobertura · «qué guardamos» · chip de reenvío |
| `PuntosMejoraList` | `web/src/widgets/map-canvas/ui/puntos-mejora-list.tsx` | La lista bajo el canvas (resuelve **H-1**), ordenada por ahorro descendente |
| `PuntoMejoraCard` | `web/src/widgets/map-canvas/ui/punto-mejora-card.tsx` | Una tarjeta = una caja × un detector |
| `InspectorMejora` | `web/src/widgets/map-canvas/ui/inspector-mejora.tsx` | El cuerpo de la 4ª tab (4 secciones) |
| `PoliticaDatosDialog` | `web/src/widgets/map-canvas/ui/politica-datos-dialog.tsx` | «Qué guardamos» + campos persistidos + retención + borrado con confirmación (**H-5**) |
| `TablaMejoraPortafolio` | `web/src/widgets/portafolio/ui/tabla-mejora-portafolio.tsx` | Las tres columnas nuevas del Portafolio |

**Promoción a `shared/ui/` (refactor de movimiento, sin cambio de conducta):** `Skeleton` y
`ErrorBody`, hoy privados en `portafolio-list.tsx:280-307`, pasan a
`web/src/shared/ui/estado-carga.tsx`. Los usan el Portafolio (igual que hoy) y la franja/canvas de
la capa Mejora (**H-6**). Es la misma jugada que `FiltroDisclosure` en el paquete de marketplace.

### 1.5 NACE — CSS

| Archivo | Alcance |
|---|---|
| `web/src/app/styles/mejora.css` (nuevo, importado desde `index.css`) | La franja, la lista de puntos, la tarjeta y el diálogo. Scope `.arnesia-mejora` |
| `web/src/app/styles/map.css` (extensión) | Las marcas dentro del nodo: `.arnesia-map .node .mej-*`, `.arnesia-map .node.sindato`, `.arnesia-map .lane-hd .lane-usd` |
| `web/src/app/styles/inspector.css` (extensión) | La 4ª tab: `.arnesia-inspector .mej-*` |
| `web/src/app/styles/portafolio.css` (extensión) | Las tres columnas: `.pf-mej-*` |

---

## 2 · Superficie por superficie, al pixel

### 2.1 Conmutador de capas (extiende `map-bar.tsx`)

Sin cambio estructural. El `tablist` sigue donde está —al final de la barra, con `ml-auto` cuando
no hay toggle de Artefactos (línea 113)—, mismo `padding: 2px`, mismo `border-radius: var(--radius-lg)`,
mismos cuatro botones a `px-2.5 py-1` / `text-xs`.

**Lo único que cambia visualmente:** el segundo slot dice `Mejora` y ya no está atenuado al 40 %.

```
┌ mapbar ──────────────────────────────────────────────────────────────────────────┐
│ dev-full-cycle   [empresa ·]  [reporta a ·]  [⬡ ·]   [artefactos off|auto|todos]  │
│                                              [ Estructura │ Mejora │ Desemp. │ Proc. ] │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Franja de contexto (`FranjaMejora`) — NUEVA

**Dónde vive:** entre `MapBar` y el canvas, dentro de `workspace-stage.tsx`. Es chrome, no canvas
(mismo criterio que `MapBar`: el canvas no sabe que existe).

**Forma:** se pega al borde inferior de la barra — `border-top: 0`, radio solo abajo —, de modo que
barra + franja se leen como **un** bloque de chrome y no como dos.

**Orden de izquierda a derecha** (es el orden de lectura del dato, no una preferencia estética):

| # | elemento | por qué en ese lugar |
|---|---|---|
| 1 | **Ventana** (`select`) | Es el modificador de todo lo demás. Va primero o el total se lee antes de saber sobre qué |
| 2 | **Total + denominador** (dos líneas) | El número principal |
| 3 | **Disclaimer** «estimado por el runtime, no es facturación» | Pegado al total. Si se va al pie, se lee después de haber creído el número |
| 4 | *(condicional)* **Chip de reenvío externo** | Solo si el forward está encendido. Un estado peligroso no se esconde a la derecha |
| 5 | **Cobertura** (`margin-inline-start: auto`) | Es la calificación del total; cierra la línea |
| 6 | **«qué guardamos»** (enlace) | Última, discreta, siempre presente |

**Espaciado:** `padding: var(--space-3) var(--space-4)` · `gap: var(--space-3) var(--space-5)` ·
`flex-wrap: wrap`. El mockup usa `10px 14px` / `10px 18px` hand-typed; **gana el token** (12/16 y
12/20) — la disciplina de token vale más que 2 px de fidelidad a un `.html`.

**Jerarquía tipográfica:**

| elemento | fuente | tamaño | peso |
|---|---|---|---|
| total | `--font-mono` | `--text-lg` | 700 |
| prefijo `USD` | `--font-mono` | `--text-xs` | 600, `--muted-foreground` |
| denominador | `--font-sans` | `--text-xs` | 400, `--muted-foreground` |
| disclaimer | `--font-sans` | `--text-xs` | 400, tono `--warn` sobre `--warn-soft` |
| cobertura | `--font-sans` | `--text-xs` | 400, `--muted-foreground` |

**Barra de cobertura — cuatro segmentos (RF-237 · H-13).** `width: 132px` · `height: 7px` ·
`border-radius: var(--radius-full)` · `border: 1px solid var(--border)` · `overflow: hidden`.
Los segmentos van en orden de calidad descendente y **llevan un separador de 1 px** de `--card`
entre sí: sin él, `--heat-3` y `--heat-2` (36 % vs. 22 % del mismo tono) no se cuentan de un
vistazo. El rótulo `cobertura` es **visible siempre**, a la izquierda de la barra (**H-11**): al
envolver en pantalla angosta, la barra no queda huérfana.

### 2.3 Canvas — las marcas dentro del nodo (extiende `arnes-node.tsx`)

**Ninguna marca nueva es `position: absolute`.** La esquina superior derecha del nodo ya está
ocupada dos veces —`.caja-badge` y `.prop-badge` comparten `top: 8px; right: 8px` en `map.css:310`
y `:329`— y un tercer badge ahí se superpondría con el de «propuesto» (J-9). Todo lo nuevo va **en
el flujo** de la tarjeta, salvo la barra de participación, que es el pie.

**Orden de hijos del nodo, de arriba abajo** (los 5 primeros son los de hoy, sin mover):

```
1  .caja-badge / .prop-badge        (absoluto, como hoy)
2  .node-top   glifo + nombre       (como hoy)
3  .node-cmd   el handle            (como hoy — el mockup lo omite por brevedad; en el producto SE QUEDA)
4  .node-trans transición del spine (como hoy)
5  .node-meta  arquetipo·perfil·gate(como hoy)
── de acá abajo, solo con la capa Mejora activa ──
6  .mej-cifra  USD 1,92 · 40 %      + <MarcaConfianza> si la confianza no es exacta
7  .mej-fuga   el detector nombrado (0 o 1: el de mayor ahorro)
8  .mej-share  barra de participación, pegada al borde inferior
```

El nodo con capa activa gana `padding-bottom: var(--space-4)` para dejar sitio a la barra de 5 px
sin comerse el `.node-meta`.

**El nodo sigue siendo un `<button>`** (`arnes-node.tsx:52`). Por eso:
`.mej-cifra`, `.mej-fuga` y `.mej-share` son **`<span>`**, jamás controles. Abrir la tarjeta de un
punto de mejora **no** se hace con un botón dentro del nodo: seleccionar la caja lleva el foco a su
tarjeta en la lista de abajo (§2.4). Anidar un botón en un botón es DOM inválido y `biome`
`useValidAriaRole` / el gate a11y lo cazan.

**Encabezado de carril.** `lane-hd` es `display:flex; justify-content:space-between` (map.css:208).
Hoy tiene `<h3>{fase}</h3>` + `<span class="count">`. Pasa a tres hijos:
`<h3>` · `<span class="count">` · `<span class="lane-usd">`. El `count` queda entre los dos y no se
quita (J-8).

**Nodos sin dato atribuible.** `opacity: .62` + `border-left-style: dotted` + una línea de motivo
en `--text-xs`, itálica, `--muted-foreground`. **Seis motivos distintos** según la clase (§7).
La opacidad sola no alcanza (RF-280): el texto es el portador.

### 2.4 Lista de puntos de mejora (`PuntosMejoraList`) — NUEVA · resuelve H-1

**Dónde vive:** debajo del canvas, dentro del área del Mapa, en un bloque propio con su encabezado.
No es un panel flotante ni un modal: es contenido de la página. Colapsable, nace abierta.

**Orden:** por ahorro contrafactual descendente. El de más plata primero — es lo único que ordena
sin discutir.

**Interacción con el canvas:** seleccionar una caja **resalta** su tarjeta y la trae a la vista
(`scrollIntoView({ block: "nearest" })`); enfocar una tarjeta **resalta** su caja en el canvas con
el mismo tratamiento de `related` que ya usa el hover (`map-canvas.tsx:119-128`). Es la relación
que el mockup no dibuja y sin la cual las dos mitades no se hablan.

### 2.5 Tarjeta de punto de mejora (`PuntoMejoraCard`) — NUEVA

**Anatomía, en orden estricto** (el orden es el argumento: qué pasa → cuánto → cuánto se ahorra →
por qué lo creemos → contra qué → qué hacer):

```
┌ .mejora ─────────────────────────────────────────────────────────────────────┐
│ ⚠  La caja «escribir el spec» reescribe el cache en cada corrida        [1]  │
│                                                                              │
│ Gasta USD 0,84 de los 1,92 de la caja (44 %) escribiendo cache que se   [2]  │
│ vence antes de volver a leerse.                                              │
│                                                                              │
│ CONTRAFACTUAL   con TTL de 1 h, las mismas 14 corridas costaban…        [3]  │
│ UMBRAL          relectura 61 % > (2−1,25)/(2−0,1) = 39,47 %  [ver el cálculo]│
│ CONFIANZA       exacta · 14 de 14 corridas con atribución                    │
│ SESGO           asume 0 lecturas fuera de la ventana → subestima el ahorro   │
│ ─────────────────────────────────────────────────────────────────────────    │
│ Fijar `cache_ttl: 1h` en esta caja      score v1  [Descartar] [Proponerlo…]  │
└──────────────────────────────────────────────────────────────────────────────┘
```

| # | elemento | especificación |
|---|---|---|
| — | contenedor | `border-left: 4px solid var(--warn)` (o `--crit`) · resto del borde `color-mix(in srgb, var(--warn) 42%, transparent)` · fondo `color-mix(in srgb, var(--warn) 5%, var(--card))` · `border-radius: var(--radius-md)` · `padding: var(--space-4)` · `gap: var(--space-3)` |
| 1 | titular | `--text-base`, peso 700, `line-height: 1.35`, `text-wrap: balance`. El `⚠` es decorativo (`aria-hidden`) — la severidad la carga el **texto** de la etiqueta de severidad, no el icono (RF-257) |
| 2 | lede | `--text-sm`, `--foreground`, `max-width: 72ch`. Nunca un monto sin su base |
| 3 | `<dl>` | `grid-template-columns: auto 1fr` · `gap: var(--space-1) var(--space-3)`. `dt` en 9 px, 700, uppercase, `letter-spacing: .07em`, `--muted-foreground` · `dd` en `--text-sm` |
| — | fila de fix | `border-top: 1px solid var(--border)` · `padding-top: var(--space-3)` · `flex-wrap: wrap`, el texto del fix con `flex: 1 1 260px` para que los botones bajen enteros y no se partan |

**Etiqueta de severidad (RF-257).** Chip de texto antes del titular: `atención` o `crítico`. Es lo
que hace la severidad legible en escala de grises; el borde de color es refuerzo.

**Etiqueta S1-only (J-2).** Cuando el detector no aplica en S2 —hoy solo B1— la tarjeta lleva un
chip `solo con telemetría de ArnesIA`, en tono neutro. Sin él, en una instalación mayormente S2 la
tarjeta insignia desaparece sin explicación.

**«ver el cálculo» (H-4).** Despliegue **en línea**, dentro de la misma tarjeta, no un modal:
`aria-expanded` en el botón, `id`/`aria-controls` al bloque. Muestra la fórmula con los números
reemplazados, la ventana y las corridas contadas. Cerrado dice `ver el cálculo`; abierto, `ocultar
el cálculo`.

### 2.6 Inspector — 4ª tab (`InspectorMejora`) — NUEVA

Las tres tabs vigentes no se tocan. La cuarta reusa `<Section>` (`inspector.tsx:71`) para los cuatro
títulos, así que hereda el patrón de la «i» doctrinal sin código nuevo.

**Ancho: 340 px** (`inspector.css:13`), no los 380 que dibuja el mockup. Todo lo de abajo se
dimensiona contra 340. Con cuatro tabs, `.dw-tabs` ya tiene `overflow-x: auto` — a 340 px las
cuatro etiquetas entran (`Resumen · Contenido · Corridas · Mejora` ≈ 300 px a `--text-sm` con
`padding: 6px 10px`), así que **no** debe aparecer scroll horizontal en la tira de tabs; si
aparece, la etiqueta es demasiado larga y se corrige la etiqueta, no el contenedor.

**Orden de secciones** (de lo auditable a lo accionable):

1. `Tokens · 7 días (14 corridas)` — tabla de 6 buckets × (tokens, USD)
2. `Costo — reportado vs. calculado` — la caja de paridad
3. `El join — corridas de esta caja` — 4 filas
4. `Detectores` — los 6 del MVP con su estado, y el bloque de los no medidos

**Tabla de buckets:** reusa `.tbl` de `inspector.css`. Columna 1 a la izquierda; columnas 2 y 3
alineadas a la derecha con `font-variant-numeric: tabular-nums`. La fila «no aplica» ocupa las dos
columnas numéricas con `colspan=2`, en itálica y `--muted-foreground` — **no** un guion en cada
celda, que se leería como cero.

**Caja de paridad:** `display: flex; justify-content: space-between` · borde y fondo `--ok` /
`--ok-soft` cuando coinciden, `--warn` / `--warn-soft` cuando difieren. El veredicto es **texto**
(`✓ coinciden` / `⚠ difieren en USD 0,18`), no solo el tono.

**Detectores:** una fila por detector: punto de 7 px + texto. Tres estados de punto (`on` `--warn`,
`clean` `--ok`, `off` `--border`) y **siempre** su texto al lado. El motivo del `off` va en una
segunda línea, `--text-xs`, `--muted-foreground`.

### 2.7 Portafolio — tres columnas nuevas (`TablaMejoraPortafolio`)

La `Fila` de `portafolio-list.tsx:224` hoy tiene 5 celdas: emblema · id+nombre · presencia · chips ·
dot de salud. Las tres nuevas se insertan **entre chips y dot**, para que el dot de salud siga
cerrando la fila (es el ancla visual que la PARIDAD del Slice 1 firmó).

| columna | ancho | alineación |
|---|---|---|
| `USD/corrida` | `min-content`, `white-space: nowrap` | derecha, `tabular-nums` |
| `tendencia` | 40 px fijos | izquierda |
| `punto de mejora` | `1fr`, se encoge primero | izquierda |

**Sparkline:** 5 barras de 5 px con `gap: 2px`, altura máxima 20 px, `align-items: flex-end`. La
última barra en `--primary` pleno; las anteriores en `color-mix(in srgb, var(--primary) 55%,
transparent)`. `role="img"` + `aria-label` con la dirección (RF-266/RF-279).

**Pie de tabla (H-12).** Una línea en `--text-xs` / `--muted-foreground` con el disclaimer de
estimación. Sin ella, un USD pelado en una tabla se lee como facturación.

**Orden con «sin dato» (RF-268).** Al ordenar por costo, los «sin dato» se agrupan **al final**, en
su propio bloque, con un separador rotulado. Nunca intercalados como si valieran 0.

---

## 3 · Tabla de campos

> Columna «cuando falta» = qué se ve exactamente. Nunca vacío, nunca 0 por omisión.

### 3.1 Franja de contexto

| etiqueta visible | fuente del dato | formato | cuando falta |
|---|---|---|---|
| `ventana` | estado local de `workspace-stage.tsx` | `select` de 3 opciones | nunca falta (default `7 días`) |
| total | suma de `costo_reportado_micros` (fallback `costo_calculado_micros`) en la ventana | `USD` + `1 234,56` · coma decimal · 2 decimales · `tabular-nums` | `— —` + copy del estado 1 |
| denominador | `count(corridas)` · `count(distinct sesion_id)` · `count(distinct caja_id)` | `61 corridas · 12 sesiones · 4 cajas` · `tabular-nums` | se omite la parte que no aplica; nunca `0 sesiones` |
| disclaimer | constante | texto plano | nunca falta |
| cobertura | `count` por `atribucion_confianza` | 4 números + etiquetas + total | con 0 corridas no se dibuja la barra; queda el estado 1 |
| chip de reenvío | config del daemon | texto + destino | **no se dibuja** cuando está apagado (el default) |
| `qué guardamos` | — | enlace | nunca falta |

### 3.2 Nodo del canvas (solo cajas)

| etiqueta visible | fuente | formato | cuando falta |
|---|---|---|---|
| cifra | `sum(costo)` de la caja en la ventana | `USD 1,92` · `tabular-nums` | la caja pasa a `sindato` con el copy «sin corridas en esta ventana» |
| participación | `costo_caja / costo_arnes` | `40 %` · entero · sin decimales | se omite si el total del arnés es 0 |
| marca de confianza | `atribucion_confianza` | chip corto + subrayado punteado | `exacta` ⇒ no se dibuja nada |
| marca de fuga | detector de mayor ahorro | nombre corto del detector | sin hallazgos ⇒ no se dibuja |
| barra de participación | ídem participación | ancho = % | sin dato ⇒ no se dibuja |
| total del carril | suma de las cajas de la fase | `USD 1,92` | `sin dato` |

### 3.3 Tarjeta de punto de mejora

| etiqueta visible | fuente | formato | cuando falta |
|---|---|---|---|
| titular | plantilla del detector + nombre de la caja | frase | — (sin caja no hay tarjeta) |
| lede | monto · base · % | `USD 0,84 de los 1,92 de la caja (44 %)` | — |
| `CONTRAFACTUAL` | `costo_con_fix` y la diferencia | monto + unidad explícita (`por corrida` / `en la ventana`) | sin contrafactual **no hay tarjeta** (A4) |
| `UMBRAL` | desigualdad del detector | `61 % > (2−1,25)/(2−0,1) = 39,47 %` · `--font-mono` | la fila **no se pinta**; en su lugar `PATRÓN` |
| `PATRÓN` | agregación de motivos de rechazo | texto | no se pinta si hay `UMBRAL` |
| `CONFIANZA` | `atribucion_confianza` + denominador | `exacta · 14 de 14 corridas con atribución` | jamás falta |
| `SESGO` | supuesto + dirección | texto en itálica, con la dirección en negrita | **jamás se omite la fila** — si no se identificó sesgo, lo dice (RF-252) |
| fix | plantilla del detector | frase con el ajuste en `<code>` | sin fix **no hay tarjeta** |
| score | `score_version` | `score v1` · 9 px · `--muted-foreground` | jamás falta |

### 3.4 Inspector · 4ª tab

| etiqueta visible | fuente | formato | cuando falta |
|---|---|---|---|
| `entrada` / `salida` | `entrada` / `salida` | entero con espacio fino de millar · `tabular-nums` | `no aplica en este runtime` |
| `cache · lectura` | `cache_lectura` | ídem | ídem |
| `cache · escritura 5 m` | `cache_escritura_5m` | ídem | ídem |
| `cache · escritura 1 h` | `cache_escritura_1h` | ídem | **`0` es un dato válido** (el runtime tiene el concepto) |
| `razonamiento` | `razonamiento` | ídem | `no aplica en este runtime` (`colspan=2`) |
| `runtime` | `costo_reportado_micros` | `USD 1,92` | `este runtime no reporta costo` |
| `nuestro catálogo` | `costo_calculado_micros` + `catalogo_version` | `USD 1,92` | `catálogo sin construir` (estado real hoy — ver traza del mockup, `:665`) |
| `corridas` | `count` | entero | `0` es válido |
| `rechazadas en el gate` | `resultado = rechazado` | entero | `sin señal de gate en estas corridas` |
| `costo de las rechazadas` | suma sobre las rechazadas | `USD 0,27` | ídem |
| `rotaciones de contexto` | evento propio de rotación | entero | `sin señal — la rotación es de ArnesIA y estas corridas fueron afuera` |

### 3.5 Portafolio

| etiqueta visible | fuente | formato | cuando falta |
|---|---|---|---|
| `arnés · puesto` | `identidad.id` + puesto de la instalación | mono + `--text-xs` atenuado en 2 líneas | sin puesto: `puesto sin declarar` |
| `USD/corrida` | `costo_total / corridas` en la ventana | `0,31` · `tabular-nums`, **sin** prefijo (está en el encabezado) | `sin dato`, en `--muted-foreground` |
| `tendencia` | 5 últimas corridas agregadas | sparkline + texto equivalente | `—` con el aviso `pocas corridas para una tendencia` |
| `punto de mejora` | el de mayor ahorro | chip: detector + monto | `✓ sin fugas detectadas` / `nunca corrió con telemetría` |

---

## 4 · Tokens por marca

**Ningún color nuevo.** Todo sale de `theme.css`, y todo tiene su par en `[data-theme="dark"]`.

### 4.1 Superficie y estructura

| marca | token |
|---|---|
| fondo de la franja | `color-mix(in srgb, var(--primary) 4%, var(--card))` — la franja se distingue de la barra sin ser otro color |
| fondo de tarjeta / drawer / tabla | `--card` |
| fondo de chip, segmento, pie | `--secondary` |
| bordes | `--border` · borde de input `--input` |
| texto / atenuado | `--foreground` · `--muted-foreground` |
| radios | `--radius-sm` (code inline) · `--radius-md` (tarjetas, botones) · `--radius-lg` (segmentos) · `--radius-full` (chips, barras) |
| sombras | `--shadow-sm` (slot activo del conmutador) · `--shadow-md` (diálogo) |
| tipografía | `--font-display` (títulos de sección) · `--font-sans` (prosa) · `--font-mono` (**todo número y todo identificador**) |
| escala | `--text-xs` 11 · `--text-sm` 12,5 · `--text-base` 14 · `--text-lg` 16 |
| espaciado | `--space-1` 4 · `--space-2` 8 · `--space-3` 12 · `--space-4` 16 · `--space-5` 20 |

### 4.2 Marcas de la capa

| marca | token | por qué ese |
|---|---|---|
| acento de la capa / slot activo / enlace «qué guardamos» / sparkline | `--primary` · `--accent-soft` | Es la marca, no una señal de salud |
| **cobertura · exacta** | `--heat-4` | El ramp de calor es el único ramp monocromo del sistema. **No se lee como `warn`** porque va dentro de una barra rotulada `cobertura`, nunca sola |
| **cobertura · por huella** | `--heat-3` | |
| **cobertura · por proceso** | `--heat-2` | |
| **cobertura · sin dato** | `--border` | Ausencia = el color del hueco, no un tono del ramp |
| separador entre segmentos | `--card`, 1 px | Sin él, `heat-3` y `heat-2` no se cuentan de un vistazo |
| barra de participación (relleno) | `--warn` a `opacity: .85` | Proporción de gasto = atención, no error |
| barra de participación (riel) | `color-mix(in srgb, var(--border) 55%, transparent)` | |
| disclaimer «estimado» | texto `--warn` · fondo `--warn-soft` · borde `color-mix(--warn 35%)` | |
| marca de fuga · grave | texto `--crit` · fondo `--crit-soft` · borde `color-mix(--crit 45%)` | |
| marca de fuga · leve / confianza no exacta | texto `--warn` · fondo `--warn-soft` | Una atribución deducida es «revisá», no «error» |
| tarjeta · severidad atención | borde izquierdo `--warn`, fondo `color-mix(--warn 5%, --card)` | |
| tarjeta · severidad crítica | borde izquierdo `--crit`, fondo `color-mix(--crit 5%, --card)` | |
| paridad de costos · coinciden | `--ok` · `--ok-soft` | |
| paridad de costos · difieren | `--warn` · `--warn-soft` | Difieren ≠ roto: puede ser catálogo viejo |
| detector activo · sin hallazgos · no disponible | `--warn` · `--ok` · `--border` | Ausencia = `--border`, **nunca** `--ok`: «no sé» ≠ «sano» (regla ya enforced en `.pf-dot-salud.sin-senal`) |
| foco | `--ring`, `outline: 2px` + `outline-offset: 1px` | El mismo que ya usa el conmutador |
| chip S1-only / chip neutro | `--secondary` + `--muted-foreground` | Es información, no señal |

**Deuda a11y heredada, no agravada:** `.text-warn` (`#c96a2e` sobre blanco) no llega a 4.5:1 y está
en el `BACKLOG.md`. Este paquete **no la arregla y no la agrava**: todo texto nuevo sobre
`--warn-soft` usa `--foreground`, no `--warn`. `--warn` queda reservado para bordes, puntos y
barras — donde el requisito es 3:1, no 4.5:1.

---

## 5 · Todos los estados, por componente

### 5.1 Slot del conmutador de capas

| estado | tratamiento |
|---|---|
| reposo (no activo) | `--muted-foreground`, fondo transparente |
| hover | fondo `--secondary` |
| foco | `outline: 2px solid var(--ring)`, `offset: 1px` — visible aunque no esté activo |
| seleccionado | fondo `--card`, texto `--foreground`, peso 600, `--shadow-sm`, `aria-selected="true"` |
| deshabilitado con motivo | `opacity: .4`, `cursor: not-allowed`, `disabled`, **motivo en `title` Y en un `<span class="sr-only">` referenciado por `aria-describedby`** |
| cargando | no aplica: el slot no espera dato para existir |
| error | no aplica |

### 5.2 Franja de contexto

| estado | tratamiento |
|---|---|
| reposo | como §2.2 |
| cargando | el total se reemplaza por un `Skeleton` de su ancho; el `select` de ventana **sigue usable**; la barra de cobertura no se dibuja. Copy: `midiendo…` con `aria-live="polite"` |
| error | la franja se mantiene y su cuerpo pasa a: motivo + botón `Reintentar`. **No desaparece** — desaparecer se leería como «no hay capa» |
| daemon no respondió | tono `--warn`, copy propio, y **las cifras previas se conservan marcadas como posiblemente viejas**. Es distinto de «no hay datos» (H-6) |
| vacío · nunca corrió | estado 1: `— —` + copy + qué hacer. La barra de cobertura no se dibuja |
| vacío · sin corridas en esta ventana | copy propio con la fecha de la última corrida + botón `Ver todo` (H-9) |
| cobertura parcial | estado 2: el denominador dice `de 5 corridas, 3` con el `3` en `--warn` |

### 5.3 Nodo con capa Mejora

| estado | tratamiento |
|---|---|
| **confianza `exacta`** | cifra + % + barra. **Sin marca de duda, sin chip, sin `title`** — la ausencia de marca *es* la señal de que es exacta |
| **confianza `por-hash`** | cifra con `border-bottom: 1px dotted currentColor` + `cursor: help` · chip corto **`por huella`** en tono `--warn` · `title` largo (§7.3). La cifra **se suma** al total |
| **confianza `por-proceso`** | mismo subrayado punteado · chip corto **`por proceso`**, mismo tono · `title` largo **distinto**: nombra el directorio como vía y **advierte la mezcla** si ahí corre más de un arnés. La cifra se suma al total |
| **confianza `sin-dato`** | **no hay cifra que marcar.** La caja pasa a `sindato`: sin número, sin barra, sin 0. Suma al cuarto segmento de la cobertura, no al total |
| reposo, nodo sin dato atribuible (no es caja) | `opacity: .62`, borde izquierdo punteado, línea de motivo por clase. **Sin cifra, sin barra, sin 0** |

> Los cuatro se distinguen así: **exacta** por *ausencia de marca*, **por huella** y **por proceso**
> por el *texto del chip* (no por el tono, que es el mismo), y **sin dato** porque *no hay número*.
> Ningún caso se pinta como «atribución aproximada»: eso taparía tres cosas distintas bajo una.
| hover / foco | igual que hoy (`map.css:257`), más el `focus-visible` del `<button>`. La capa **no** agrega interacción nueva al nodo |
| seleccionado | igual que hoy, y además su tarjeta de punto de mejora se resalta y se trae a la vista |
| atenuado por foco de otro nodo (`dim`) | las marcas nuevas heredan la opacidad del nodo, sin regla propia |
| cargando | la geografía se pinta **completa**; las cifras son `Skeleton` del ancho de `USD 0,00`. Nunca se colapsa el layout |
| error | la capa vuelve a `Estructura` y la franja muestra el error. El canvas **no** se rompe: `ErrorBoundary` ya lo cubre |

### 5.4 Tarjeta de punto de mejora

| estado | tratamiento |
|---|---|
| reposo | como §2.5 |
| hover | sin cambio de fondo (no es clickeable como bloque); los botones tienen su propio hover |
| foco dentro | anillo `--ring` en el control enfocado; el contenedor gana `outline` de 1 px cuando el foco entra por `scrollIntoView` desde el canvas |
| resaltada (por selección de su caja) | borde izquierdo a 6 px y `--shadow-sm`. **Además** un texto `↔ caja seleccionada`, para no depender del grosor |
| «ver el cálculo» abierto | `aria-expanded="true"`, bloque desplegado con la fórmula resuelta, botón dice `ocultar el cálculo` |
| botón deshabilitado con motivo | `Proponerlo en el chat` queda `disabled` + `title` cuando el arnés está fuera del alcance del chat embebido (guardrail vigente). El motivo también en texto, no solo en `title` |
| descartada | sale de la lista con `aria-live="polite"` anunciando el descarte y cómo revertirlo |
| cargando | la lista muestra 2 `Skeleton` con la altura de una tarjeta |
| error | la lista muestra motivo + `Reintentar`; el canvas y la franja siguen funcionando |
| vacío | copy de H-2 + la lista de los seis detectores que corrieron. **Nunca** una sección vacía |

### 5.5 Cuarta tab del inspector

| estado | tratamiento |
|---|---|
| reposo | 4 secciones en orden |
| tab no seleccionada | `--muted-foreground`, sin subrayado; el panel con `hidden` |
| tab seleccionada | `--foreground`, peso 600, `border-bottom: 2px solid var(--primary)` |
| foco en la tab | anillo `--ring` |
| nodo no-caja | una sola sección con el motivo, con el **mismo texto** que muestra el canvas para ese nodo. Cero tablas vacías |
| cargando | cada sección con su `Skeleton`; los títulos ya visibles |
| error | motivo + `Reintentar` dentro de la tab; las otras tres tabs siguen intactas |
| vacío (caja sin corridas en la ventana) | copy de «sin corridas en esta ventana» + la fecha de la última |

### 5.6 Fila del Portafolio

| estado | tratamiento |
|---|---|
| reposo, con dato | las 3 columnas nuevas pobladas |
| reposo, sin dato | `sin dato` en `--muted-foreground`, `—` en tendencia, chip neutro en punto de mejora |
| hover / foco / seleccionada | **igual que hoy** (`.pf-fila`, `.pf-fila.seleccionada`, `aria-current`). Las columnas nuevas no cambian el comportamiento de la fila |
| cargando | el `Skeleton` vigente ya cubre la fila entera |
| error | el `ErrorBody` vigente |
| vacío (portafolio vacío) | el `VaciaBody` vigente, sin cambio |
| ordenada por costo | los «sin dato» agrupados al final, con separador rotulado |

### 5.7 Diálogo de política de datos y borrado

| estado | tratamiento |
|---|---|
| reposo | 3 bloques: qué NO se guarda · qué SÍ se guarda (lista de campos) · retención + acción |
| foco | focus trap con `trapTabKeyDown`, `role="dialog"` + `aria-modal="true"`, Esc cierra |
| confirmación de borrado | segundo paso dentro del mismo diálogo. El botón destructivo **nace deshabilitado** hasta que el operador confirma que entendió el alcance |
| borrando | botón en estado de espera, el diálogo **no se puede cerrar** (mismo patrón que el wizard con POST en vuelo, S1-D19) |
| error de borrado | motivo textual; el diálogo queda abierto y el dato intacto |
| éxito | el diálogo cierra y las superficies pasan al estado 1, con `aria-live` anunciando el resultado |

---

## 6 · Responsive y desbordamiento

### 6.1 Breakpoints

La app corre en un WebView de escritorio; el diseño es desktop-first y **no** hay layout móvil.

| ancho | comportamiento |
|---|---|
| **≥ 1280 px** | Todo en una línea: la franja no envuelve; la lista de puntos de mejora a dos columnas |
| **1024–1279 px** | La franja **envuelve** en dos líneas: (ventana + total + disclaimer) arriba, (cobertura + enlace) abajo. El rótulo `cobertura` es lo que impide que la barra quede huérfana (**H-11**). La lista de puntos, a una columna |
| **< 1024 px** | El inspector deja de convivir con la lista de puntos: la lista se colapsa a su encabezado con el contador (`Puntos de mejora · 2`), y el operador la abre a demanda. La franja envuelve a tres líneas |

### 6.2 Desbordamiento del canvas — **corrección al mockup**

El mockup usa `.procrow { overflow-x: auto }` (`:120`). **En el producto no existe esa caja.** Los
carriles viven dentro del `stage` con `transform` que maneja `useViewport`
(`map-canvas.tsx:81`, `map.css:15-31`): el Mapa se recorre con **pan y zoom**, no con scrollbar.

Consecuencias que la implementación tiene que respetar:

- **No se agrega ningún contenedor con `overflow` dentro del canvas.** Uno solo rompería el
  cálculo de `useEdgePaths` (que mide posiciones absolutas para dibujar los edges).
- El `fitView` de montaje (`map-canvas.tsx:134-138`) sigue siendo el que garantiza que el arnés
  entero se vea; con la capa activa los nodos son **más altos** (la barra de participación y la
  cifra suman ~26 px), así que `fitView` va a elegir un zoom menor. Es correcto y es el
  comportamiento overview-first ya firmado (RF-50).
- La barra de participación es `left: 0; right: 0` **del nodo**, no del carril: no se ve afectada
  por el zoom más que por escala.

### 6.3 Desbordamiento de tablas

Las tres tablas nuevas **sí** scrollean, cada una en su propio contenedor con `overflow-x: auto`:

| tabla | contenedor | mínimo antes de scrollear |
|---|---|---|
| buckets del inspector | `.arnesia-inspector .mej-scroll` | 3 columnas. El drawer real mide **340 px** (`inspector.css:13`), no los 380 del mockup: la tabla se dimensiona contra 340 y en modo normal **no** scrollea; en expandido, menos aún |
| join del inspector | ídem | 2 columnas, nunca scrollea |
| Portafolio | `.pf-mej-scroll` | 7 columnas. Por debajo de ~980 px la columna `punto de mejora` se encoge primero (`1fr`), y recién después aparece el scroll |

**Regla:** el `<body>` y la página **jamás** scrollean horizontalmente. El scroll vive dentro del
contenedor de la tabla, y ese contenedor lleva `tabindex="0"` + `role="region"` + `aria-label` para
que sea alcanzable por teclado (requisito de axe cuando un contenedor scrollea).

**Textos largos:** el motivo de un detector no disponible y el motivo de «sin dato atribuible» se
envuelven — **jamás** `text-overflow: ellipsis`. Un motivo truncado es peor que no tenerlo (misma
regla que `AvisoChip`, BR-8 del paquete de marketplace).

---

## 7 · Copy exacto

> Esto es material de diseño, no relleno: el test asserta el literal. Escrito del lado del usuario,
> activo y específico. **Nada de «Error al obtener los datos».**

### 7.1 Conmutador de capas

| dónde | copy |
|---|---|
| slot | `Mejora` |
| tooltip `Desempeño` (disabled) | `La señal ya llega —duración por request y por herramienta—. Falta decidir qué es «desempeño» a nivel Mapa.` |
| tooltip `Proceso` (disabled) | `Entra en parte por la capa Mejora. Falta mapear todos los eventos a fases del arnés.` |

### 7.2 Franja de contexto

| dónde | copy |
|---|---|
| rótulo de ventana | `ventana` · opciones `7 días` · `30 días` · `todo` |
| disclaimer | `estimado por el runtime, no es facturación` |
| rótulo de cobertura | `cobertura` |
| texto de cobertura | `12 exactas · 3 por huella · 2 por proceso · 1 sin dato — sobre 18 corridas` |
| `aria-label` de la barra | `Cobertura de la atribución: 12 corridas exactas, 3 por huella, 2 por proceso, 1 sin dato, sobre 18 corridas.` |
| cobertura completa | `atribución exacta en las 18 corridas` |
| enlace de privacidad | `qué guardamos` |
| chip de reenvío (solo si está encendido) | `reenvío externo encendido → langfuse.local` |
| cargando | `midiendo…` |
| error | `No se pudo leer la telemetría — {motivo}.` + botón `Reintentar` |
| daemon caído | `El daemon no respondió. Lo que ves es la última medición, del {fecha}.` |

### 7.3 Canvas

| dónde | copy |
|---|---|
| total del carril sin dato | `sin dato` |
| `aria-label` de la barra de participación | `40 % del gasto del arnés en esta ventana.` |
| marca de fuga · B1 | `re-warm TTL` |
| marca de fuga · P1 | `rechazo en gate` |
| marca de fuga · B3 | `modelo cambiado` |
| marca de fuga · B6 | `sesión abandonada` |
| marca de fuga · B2 | `rotación cara` |
| marca de confianza · `por-hash` | chip `por huella` · `title`: `Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA.` |
| marca de confianza · `por-proceso` | chip `por proceso` · `title`: `Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla.` |
| sin dato · subagente | `sin dato atribuible — el subagente no se distingue en el turno` |
| sin dato · regla | `sin dato atribuible — una regla no consume por sí misma` |
| sin dato · MCP | `sin dato atribuible — el costo del MCP está incluido en la caja que lo llama; todavía no se desglosa` *(corrige el mockup `:404`, hueco H-10)* |
| sin dato · conocimiento | `sin dato atribuible — el conocimiento se paga en la caja que lo carga` |
| sin dato · hook | `sin dato atribuible — un hook informa, no consume` |
| sin dato · resto | `sin dato atribuible — esta capa mide cajas` |
| caja sin corridas en la ventana | `sin corridas en esta ventana` |

### 7.4 Lista y tarjeta de punto de mejora

| dónde | copy |
|---|---|
| título de la lista | `Puntos de mejora` + contador |
| lista vacía · hay datos | `Hay datos y ningún punto de mejora que pase el corte.` / `Los seis detectores corrieron sobre 61 corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.` |
| lista vacía · no hay datos | (no se dibuja: manda el estado 1 de la franja) |
| severidad | `atención` · `crítico` |
| chip S1-only | `solo con telemetría de ArnesIA` |
| titular B1 | `La caja «escribir el spec» reescribe el cache en cada corrida` |
| lede B1 | `Gasta USD 0,84 de los 1,92 de la caja (44 %) escribiendo cache que se vence antes de volver a leerse.` |
| titular P1 | `La caja «revisar el build» se rechaza en el gate 3 de cada 4 veces` |
| lede P1 | `Las corridas rechazadas costaron USD 0,67 de los 0,89 de la caja (75 %): se paga el trabajo y se descarta el resultado.` |
| rótulos `<dt>` | `Contrafactual` · `Umbral` · `Patrón` · `Confianza` · `Sesgo` |
| contrafactual B1 | `Con TTL de 1 h, las mismas 14 corridas costaban USD 0,31 → diferencia USD 0,53 por corrida.` |
| contrafactual P1 | `Si el gate pasara a la primera, el ciclo costaba USD 0,22 en vez de 0,89.` |
| umbral B1 | `relectura 61 % > break-even (2−1,25)/(2−0,1) = 39,47 %` |
| patrón P1 | `Los 3 rechazos citan el mismo motivo: «el veredicto no lista hallazgos».` |
| confianza | `exacta · 14 de 14 corridas con atribución` |
| confianza degradada | `no exacta · 11 de 14 exactas, 3 por huella` |
| sesgo B1 | `Asume 0 lecturas fuera de la ventana de 7 días → subestima el ahorro.` |
| sesgo P1 | `No descuenta lo que la revisión aporta aunque rechace → sobreestima el desperdicio.` |
| sin sesgo identificado | `No se identificó ningún supuesto que sesgue este cálculo.` |
| fix B1 | `Fijar cache_ttl: 1h en esta caja` |
| fix P1 | `El contrato de la caja no exige hallazgos[] en el entregable` |
| score | `score v1` |
| botón de cálculo | `ver el cálculo` / abierto: `ocultar el cálculo` |
| botones | `Descartar` · `Proponerlo en el chat` |
| `Proponerlo` deshabilitado | `title`: `Este arnés está fuera del alcance del chat embebido.` |
| tras descartar | `Descartado. Se puede volver a mostrar desde la tab Mejora de esa caja.` |
| nota al pie de la lista | `«Proponerlo en el chat» no escribe archivos: abre el chat con el cambio propuesto, y se aplica por el camino de siempre, con sus permisos y su gate.` |

### 7.5 Inspector · 4ª tab

| dónde | copy |
|---|---|
| tab | `Mejora` |
| sección 1 | `Tokens · 7 días (14 corridas)` |
| encabezados | `bucket` · `tokens` · `USD` |
| filas | `entrada` · `salida` · `cache · lectura` · `cache · escritura 5 m` · `cache · escritura 1 h` · `razonamiento` |
| no aplica | `no aplica en este runtime` |
| nota | `«No aplica» no es 0. Un cero donde el concepto no existe sería mentira.` |
| sección 2 | `Costo — reportado vs. calculado` |
| paridad ok | `runtime USD 1,92` · `nuestro catálogo USD 1,92` · `✓ coinciden` |
| paridad divergente | `⚠ difieren en USD 0,18` |
| nota | `Guardamos los dos. Si difieren, o nuestro catálogo está viejo o el runtime cambió su tarifa.` |
| sin costo del runtime | `este runtime no reporta costo — calculado con el catálogo v2026-07-20` |
| sección 3 | `El join — corridas de esta caja` |
| filas | `corridas` · `rechazadas en el gate` · `costo de las rechazadas` · `rotaciones de contexto` |
| sin señal de proceso | `sin señal de gate en estas corridas` |
| sección 4 | `Detectores` |
| activo | `{detector} · activo` |
| limpio | `{detector} · sin hallazgos` |
| no disponible | `{detector}` + segunda línea `no disponible: {motivo}` |
| motivo B1 en S2 | `no disponible: sin el result del stream-json — este arnés corrió fuera de ArnesIA` |
| motivo B2 parcial | `no disponible: 2 de 3 rotaciones ocurrieron fuera de ArnesIA` |
| fuera del MVP | `no medido todavía` |
| nodo no-caja | `Esta capa mide cajas. {motivo del nodo}` |

### 7.6 Portafolio

| dónde | copy |
|---|---|
| encabezados nuevos | `USD/corrida` · `tendencia` · `punto de mejora` |
| sin dato | `sin dato` |
| sin fugas | `✓ sin fugas detectadas` |
| nunca corrió | `nunca corrió con telemetría` |
| pocas corridas | `pocas corridas para una tendencia` |
| `aria-label` de tendencia | `Tendencia en alza en las últimas 5 corridas.` / `estable` / `a la baja` |
| chip de punto de mejora | `⚠ re-warm de cache · USD 0,53/corrida` |
| separador de sin-dato al ordenar | `Sin datos de telemetría` |
| pie de tabla | `Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no aparecen en cero.` |

### 7.7 Los siete estados honestos

| # | copy |
|---|---|
| 1 · sin datos | `— —` / `Este arnés nunca corrió con telemetría.` / `Abrí una sesión desde ArnesIA y la medición arranca sola.` |
| 1b · sin corridas en la ventana (H-9) | `Sin corridas en los últimos 7 días. La última fue el 12/07.` + botón `Ver todo` |
| 2 · cobertura parcial | `USD 3,10 — de 5 corridas, 3` / `2 corridas quedaron sin atribución.` |
| 3 · S2 sin instrumentar | rótulo `corrió fuera de ArnesIA, sin instrumentar` / `re-warm por TTL — no disponible: sin el result del stream-json.` |
| 3b · S2 instrumentado (J-10) | rótulo `corrió fuera de ArnesIA, instrumentado` / `Llega la misma señal que dentro de ArnesIA.` |
| 4 · otro runtime | `USD 0,74` / `calculado con el catálogo v2026-07-20 — este runtime no reporta costo` |
| 5 · catálogo viejo | `⚠ Precios del release, sin refrescar desde el 20/07.` / `Funciona sin internet; te avisamos que está funcionando así.` |
| 6 · por huella | `vitalia USD 0,47` / `identificado por la huella del arnés, no por su nombre` |
| 7 · qué guardamos | ver §7.8 |

### 7.8 Política de datos y borrado (RF-275 · H-14 · ANEXO H4)

**Resumen en la franja** (dos líneas, es lo que se ve sin abrir nada):

```
Nada de tu cuenta. Nada de la conversación.
Retención 90 días · qué guardamos
```

**Diálogo completo:**

| bloque | copy |
|---|---|
| título | `Qué guardamos de la telemetría` |
| qué NO se guarda | `No guardamos nada de tu cuenta: ni email, ni identificadores de usuario, ni de organización.` / `Y no guardamos nada del contenido: ni tu prompt, ni la respuesta, ni lo que leyó o escribió una herramienta. Eso llega por el canal de hooks y se descarta antes de escribirse en disco.` |
| qué SÍ se guarda | `Guardamos, por turno: qué arnés, qué caja, qué sesión, qué modelo, cuántos tokens de cada tipo, cuánto costó, cuánto tardó y si el gate lo aceptó o lo rechazó.` + enlace `ver los campos exactos` (despliega la lista literal, la misma allowlist que aplica la ingesta) |
| retención | `Se borra solo a los {N} días.` |
| acción | botón `Borrar la telemetría de este arnés` |
| confirmación | título `Borrar la telemetría de «vitalia»` / cuerpo `Se borran {N} corridas medidas y los puntos de mejora que salieron de ellas. No se puede deshacer.` / botones `Cancelar` · `Borrar` |
| borrando | `Borrando…` (el diálogo no se puede cerrar) |
| error | `No se pudo borrar — {motivo}. No se borró nada.` |
| éxito | `Listo. Este arnés vuelve a estar sin datos de telemetría.` |

> ⚠️ **`{N} = 90` es un valor PROPUESTO, no firmado** (J-6): D15.3 firmó «TTL por default» sin
> número. La UI lee el valor de la configuración; el 90 del mockup es una propuesta al operador.

---

## 8 · Gates de lint y a11y

| gate | qué verifica acá | comando |
|---|---|---|
| `tsc` | el rename `Capa: "tokens"` → `"mejora"` no deja un consumidor huérfano | `pnpm --dir web run typecheck` |
| `biome` | DOM válido (nada de botón dentro de botón), reglas a11y de base | `pnpm --dir web run lint` |
| `stylelint` + `strict-value` | **cero color literal** en `mejora.css` y en las extensiones | dentro de `verify` |
| `depcruise` + `steiger` | `entities/telemetria` no importa widgets, no importa otra entity, no importa transporte (`fe-transporte-independiente`, `fe-taxonomia-componentes` v1.1) | `pnpm --dir web run fsd` |
| Storybook + `vitest-browser` | una story por estado de §5, con `play()`; gate a11y en `error` (ya en `.storybook/preview.ts`) | ⚠️ **no corre en background** (Chromium no headless): `pnpm --dir web run verify` en sesión interactiva |
| conformance | capabilities nuevas (R1 · R2 · R3 · R4) + `arnesia conformance --todo` sin regresión | `make` / hook `pre-commit` |

**Stories obligatorias** (una por estado, no una por componente):

```
franja-mejora.stories.tsx        reposo · cargando · error · daemon-caído · sin-datos ·
                                 sin-corridas-en-ventana · cobertura-parcial · 4-niveles ·
                                 reenvío-encendido
arnes-node.stories.tsx (+)       caja-con-cifra · caja-por-huella · caja-por-proceso ·
                                 caja-sin-corridas · sindato-subagente · sindato-regla · sindato-mcp
punto-mejora-card.stories.tsx    atención · crítico · S1-only · cálculo-abierto · resaltada ·
                                 proponer-deshabilitado
puntos-mejora-list.stories.tsx   dos-tarjetas · vacía-con-datos · cargando · error
inspector-mejora.stories.tsx     completo · no-aplica · cero-legítimo · paridad-divergente ·
                                 sin-señal-de-gate · nodo-no-caja · cargando · error
tabla-mejora-portafolio.stories  con-dato · sin-dato · sin-fugas · pocas-corridas · ordenada
politica-datos-dialog.stories    reposo · campos-desplegados · confirmación · borrando · error
```

Las fixtures salen de `verificacion-2026-07-26/evidencia/` — **shape real, no inventado**.
