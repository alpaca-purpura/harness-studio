# ArnesIA · Mapa MVP — `spec.md` (el QUÉ)

> **Paquete Fase C, doc 1 de 3 · HUB.** Define **qué hace** el Mapa (comportamiento/funcional). El
> **cómo se ve al pixel** vive en [`design.md`](./design.md); el **cómo se implementa** en
> [`architecture.md`](./architecture.md). Este doc referencia a ambos con `→ design.md §X` /
> `→ architecture.md §Y`.
>
> **Fuente de verdad:** mockup firmado `mockups/arnesia-mapa-mvp.html` (Gate 1 ✓, commit `0736d2c`).
> Cada requisito es trazable a `mockup:línea` + `shot#` (§12). Reglas: cero invención · lo `deferred`
> se declara honesto · lo que la doctrina no respalda se marca **PROPUESTA**.

---

## 1 · Propósito · alcance · no-goals

### 1.1 · Propósito

El Mapa es la **superficie única** para **entender un arnés de forma intuitiva**: qué componentes lo
forman, cómo se relacionan y cómo fluye el trabajo por el proceso. Es lienzo de lectura primero
(read-only en el MVP); crear/editar/observar son evoluciones aditivas sobre el mismo lienzo
(VISION.md «Mapa = lienzo único», `VISION.md:184-186`).

Norte del usuario: **abrir el Mapa y comprender el arnés sin manual** — geografía fija (Guardia arriba,
proceso al centro, Base abajo), tipo por forma+color+letra, flujo siempre visible, detalle bajo demanda.

### 1.2 · Alcance (MVP = Hito 1 + Hito 2)

Capa **Estructura** solamente (§8): tipo · forma · color · etiqueta · handle · edges · transición del
spine. El Mapa renderiza el **grafo L0** de UN arnés (§2). Dos fixtures en el mockup: **Luana**
(7 fases, denso, ejercita las 10 clases) y **dev-full-cycle** (dogfood real que el backend servirá,
`spec.md §9`).

### 1.3 · No-goals (explícitos)

- **NG-1** · No es React Flow. El sustrato es **HTML bandas/carriles + overlay SVG** ([[hs-09-mapa-sustrato-html-svg]]); React Flow queda reservado al Organigrama. → `architecture.md §2`.
- **NG-2** · No hay capas Tokens/Desempeño/Proceso en el MVP — esperan telemetría JSONL (indexer real). Se declaran **staged**, con métricas «—» (§8). Doctrina define las 4 capas **conmutables** sobre una geografía (`VISION.md:186-187`); el staging es decisión de **plan**, no de doctrina (PROPUESTA de scoping).
- **NG-3** · No hay crear/editar sobre el lienzo (Hito 3+).
- **NG-4** · No hay tablero «Flujo del trabajo» (spine kanban / paquete vivo) — vista hermana futura, telemetry-gated ([[hs-09-mapa-decisiones-doctrina]] decisión 6).
- **NG-5** · No hay realtime (SSE `event: map`) en el MVP (Hito 3).
- **NG-6** · No hay multi-arnés simultáneo en el lienzo (un arnés por vista; el picker es Hito 2).

---

## 2 · Modelo de datos L0 (claves español EXACTAS)

El Mapa consume un `Graph` L0. Claves = las que emite el contrato JSON (español), espejo de
`internal/domain/graph.go`+`box.go` ↔ `docs/architecture/contracts/schema/graph.l0.schema.json`. Detalle visual de
cada campo → `design.md`; decisiones de schema para los campos PROPUESTOS → `architecture.md §5`.

### 2.1 · `Graph` (raíz servida por el endpoint)

| Clave | Tipo | Req | Fuente doctrinal |
|-------|------|-----|------------------|
| `arnes` | objeto (§2.2) | opcional | `graph.go:92-96` (`Graph.Arnes`) |
| `nodos` | array de nodo (§2.3) | **sí** | `graph.go:94` (tag `nodos`, no-omitempty); schema `required:["nodos"]` `graph.l0.schema.json:7` |
| `edges` | array de edge (§2.4) | opcional | `graph.go:95` |

### 2.2 · `arnes` (manifiesto)

| Clave | Tipo | Notas | Fuente |
|-------|------|-------|--------|
| `id` | string | id del arnés | `graph.go:29` |
| `rol` | string | rol×proceso (canónico, ex-`puesto`) | `graph.go:30` |
| `proceso` | string | proceso de compañía | `graph.go:31` |
| `empresa` | string | | `graph.go:32` |
| `reporta_a` | string \| null | organigrama (META de enganche); **emite `null`** | `graph.go:33`; schema `["string","null"]` `:20` |
| `marketplace` | string | canal git | `graph.go:35` |
| `fases` | string[] | **orden de las lanes** (§3.2) | `graph.go:36` |
| `spine` | objeto \| null | estados del trabajo (§2.5) | `graph.go:37,45-50` |

### 2.3 · `nodo` (= `Box`)

| Clave | Tipo | Estado | Fuente |
|-------|------|--------|--------|
| `id` | string | **real** (req) | `box.go:171`; schema req `:79` |
| `clase` | enum 10 (§2.6) | **real** (req) | `box.go:172`; schema `$defs.clase` |
| `nombre` | string | **real** (req) | `box.go:173` |
| `banda` | enum 7 (§2.7) | **real** | `box.go:174`; schema `$defs.banda` |
| `fase` | string \| null | **real** (agnóstico, sin enum) | `box.go:175`; schema `:86` |
| `estado` | string | **real** — transición `"<entra> -> <sale>"` (semántica del `trans` del mockup) | `box.go:176,210` |
| `canal` | enum beta/estable/propuesto/deprecado | **real** — «propuesto» ≈ el `prop` del mockup | `box.go:178`; schema `:87` |
| `procedencia` | enum | **real** — honestidad-de-dato | `box.go:180` |
| `contract` | objeto | **real** — contiene **`caja:bool`** (§2.8) | `box.go:181` |
| `contract.caja` | bool | **real** — un nodo «es caja» ⇔ `contract.caja==true` (`Box.IsCaja()` `box.go:186`) | `box.go:208`; schema `box.contract.schema.json:54-57` |
| **`trans`** | string | **PROPUESTA (mockup-only)** — atajo visual de la transición; su semántica ya vive en `contract.estado` | ausente en dominio/schema |
| **`alw`** | bool | **PROPUESTA (mockup-only)** — «siempre en contexto» de una regla; hoy sólo display en `UX.md:534` | ausente |
| **`prop`** | bool | **PROPUESTA (mockup-only)** — índice semántico opcional; ≈ `canal:"propuesto"` | ausente como campo |
| **`origen`** | enum `estandar`\|`del-puesto` | **PROPUESTA (mockup-only)** — facet firmado; doctrina SILENTE (sólo `procedencia`/`banda`/`canal` existen) | ausente |

> **RF-01 (contrato de campos).** El Mapa DEBE renderizar a partir de las claves reales
> (`id·clase·nombre·banda·fase·estado·canal·procedencia·contract.caja`). Los 4 campos PROPUESTOS
> (`trans·alw·prop·origen`) son marcas visuales que el MVP deriva/simula; su promoción a campo L0 de
> primera clase es una **decisión de schema** (→ `architecture.md §5`). El `nodo` del schema es
> `additionalProperties:true` (`graph.l0.schema.json:80`) → hoy se aceptan **pero se ignoran**.

### 2.4 · `edge`

| Clave | Tipo | Fuente |
|-------|------|--------|
| `de` | string (id nodo origen) | `graph.go:16`; schema req `graph.l0.schema.json:102` |
| `a` | string (id nodo destino) | `graph.go:17` |
| `tipo` | enum `invoca`\|`lee`\|`escribe` | `graph.go:18`; schema enum `graph.l0.schema.json:109` |

### 2.5 · `spine` (estados del trabajo)

`inicial` · `terminales?` · `estados[]` · `transiciones[]{de,a}` (`graph.go:45-56`). El MVP **no dibuja**
el spine como tablero (NG-4); sólo lo usa como fuente de la etiqueta `contract.estado` por caja (§4.4).

### 2.6 · Enum `clase` (10 primitivas)

`skill·subagent·hook·rule·command·mcp·plugin·settings·output-style·statusline` (`box.go:27-38`;
schema `$defs.clase`). Mapa clase→visual → `design.md §4.2`.

### 2.7 · Enum `banda` (7)

`guardia·fase·base·libreria-expertos·meta-harness·terceros·marcas-dormidas` (`box.go:124-132`).
Geografía → §3.

### 2.8 · `contract` (fusionado)

Bloque intención+clasificación+cableado+aceptación (`box.go:195-218`). El MVP sólo lee **`caja`** (§4.2)
y **`estado`** (§4.4). El resto (why/capabilities/necesita/entrega/gate/handoff) alimenta el **Inspector
de Hito 2** (§9.2).

---

## 3 · Regiones (geografía fija) — RF-10…

Detalle visual (tintes, headers, medidas) → `design.md §3`.

- **RF-10.** El Mapa DEBE mostrar exactamente **3 regiones apiladas verticalmente**, en este orden:
  **Guardia** (hooks transversales) → **Proceso** (carriles por fase) → **Base** (conocimiento, reglas y
  soporte). Canónico VISION A6 (`VISION.md:81-83`) + region model (`VISION.md:184-186`). `mockup:406-446`.
- **RF-11.** La región **Guardia** DEBE contener los nodos con `banda=guardia`, renderizados como **chips
  compactos** de 1 línea (§4.3). Si no hay hooks, muestra «— sin hooks —». `mockup:407-413` · shot7.
- **RF-12.** La región **Proceso** DEBE contener **una lane por fase** declarada en `arnes.fases`, en ese
  orden; fases no declaradas pero presentes en nodos se **anexan** (primera aparición). `mockup:396-400,419`.
- **RF-13.** La región **Base** DEBE agrupar los nodos de las bandas `base·libreria-expertos·meta-harness·
  terceros·marcas-dormidas` (§6), sólo las bandas con ≥1 nodo. `mockup:441-444`.
- **RF-14.** Cada región DEBE llevar un header `region-hd` con su título textual (§`design.md §3.1`) y una
  regla horizontal. `mockup:352`.

```gherkin
Feature: Geografía de 3 regiones
  Scenario: Orden fijo de regiones
    Given un arnés cualquiera
    When se renderiza el Mapa
    Then aparecen 3 regiones en orden Guardia, Proceso, Base
    And cada una muestra su título y su regla horizontal

  Scenario: Guardia vacía (arnés dogfood)
    Given el arnés "dev-full-cycle" sin hooks
    When se renderiza el Mapa
    Then la región Guardia muestra "— sin hooks —"      # shot7
```

---

## 4 · Nodo y variantes — RF-20…

Detalle visual de cada variante → `design.md §5-6`.

- **RF-20.** Todo nodo DEBE mostrar: **glyph** (forma+color+letra por clase, `design.md §4`) + **nombre**
  (mono, clamp a 2 líneas) + **handle** derivado por clase (§4.1) + borde-izquierdo de 3px con el color de
  la clase. `mockup:336-351`.
- **RF-21 (handle derivado).** El handle NO es un dato; se deriva de la clase: `command`→id · `skill`→`/id`
  · `subagent`→`@id` · `hook`→«evento» · `rule`→«always-on»/«condicional» (según `alw`) · resto→id.
  `mockup:327-334`. → `design.md §5.2`.
- **RF-22 (caja).** Un nodo con `contract.caja==true` DEBE renderizarse como **caja de proceso**:
  borde-izq 5px + tinte + badge «caja». Es el skill-frente de la fase. `mockup:104-105,343` · shot1.
- **RF-23 (transición del spine).** Una caja con `estado`/`trans` DEBE mostrar una **etiqueta de
  transición** («`entra → sale`», prefijo ◇). Cada caja posee **UNA** transición del spine (METODOLOGIA
  §3, `METODOLOGIA.md:160` + VISION A3 `VISION.md:69-73`). `mockup:114-115,349`.
- **RF-24 (apoyo).** En una fase con caja, los nodos que **no** son caja DEBEN renderizarse como **apoyo**:
  indentados + fondo más tenue, con un hairline `lane-sep` antes del primer apoyo (sin la palabra
  «apoyo»). `mockup:95-96,427-429` · shot1.
- **RF-25 (orden caja-first).** Dentro de cada lane, la(s) caja(s) van **primero** (fila 1) para alinear
  el spine horizontal. `mockup:424`.
- **RF-26 (compacto).** Los hooks de Guardia usan la variante **compacta**: fila de 1 línea, ocultan
  handle/transición/badges. `mockup:99-102,412`.
- **RF-27 (origen — PROPUESTA).** Un nodo con `origen=del-puesto` (llenado en onboarding) DEBE llevar
  **borde-izq punteado**; `origen=estandar` (del kit) = borde sólido. Facet firmado, doctrina SILENTE
  (`mockup:107,229`). → `architecture.md §5`.
- **RF-28 (propuesto — PROPUESTA).** Un nodo con `prop`/`canal=propuesto` (índice semántico opcional)
  DEBE llevar badge «propuesto» (ámbar, borde punteado). `mockup:108,278,344`.

```gherkin
Feature: Nodo y variantes
  Scenario: Caja de proceso con transición
    Given un nodo skill con contract.caja=true y estado "idea -> spec"
    When se renderiza
    Then muestra borde-izquierdo de 5px, tinte de clase y badge "caja"
    And muestra la etiqueta de transición "idea → spec" con prefijo ◇

  Scenario: Nodo de apoyo subordinado
    Given una fase "build" con una caja "builder" y un subagente "test-author"
    When se renderiza la lane
    Then "builder" aparece primero (fila 1)
    And "test-author" aparece indentado, tras un hairline, con fondo tenue

  Scenario: Hook como chip compacto
    Given un nodo hook en banda guardia
    Then se renderiza en una fila de 1 línea sin handle ni transición ni badges

  Scenario: Origen del-puesto (PROPUESTA)
    Given un nodo cuyo id está en el set del-puesto
    Then su borde-izquierdo se dibuja punteado
```

---

## 5 · Edges — RF-30…

Detalle visual (colores/dash/bezier/marcadores) → `design.md §9`.

- **RF-30.** El Mapa DEBE dibujar tres tipos de edge, con distintivo propio: **`invoca`** (rojo `--crit`,
  sólido, con flecha) · **`escribe`** (verde `--ok`, dash `3 3`, con flecha) · **`lee`** (ámbar `--warn`,
  dash `4 4`, **sin** flecha). `mockup:239-243`.
- **RF-31 (spine-always).** Por defecto (sin foco) sólo se dibujan los edges `invoca` (el **backbone**/spine
  del flujo). `mockup:465-467`.
- **RF-32 (hover-reveal).** Al pasar el cursor por un nodo, se revelan **además** sus edges `lee`/`escribe`;
  el edge que toca el foco se resalta (`opacity .95`, `width 2.2`) y el backbone no-tocado se atenúa
  (`opacity .2`). `mockup:465,484` · shot3.
- **RF-33 (foco/dim).** Al enfocar un nodo, los nodos **no relacionados** se atenúan (`opacity .4`).
  `mockup:490-496` · shot3.
- **RF-34 (endpoint oculto).** Si un endpoint está colapsado/oculto (p.ej. reglas plegadas), su edge **no**
  se dibuja (nada de líneas huérfanas). `mockup:471`.
- **RF-35 (bezier).** Cada edge es un bezier cúbico de borde-derecho→borde-izquierdo con
  `dx=max(30,|Δx|/2)`, coords en espacio de contenido (÷z). `mockup:475,478`.

```gherkin
Feature: Edges spine-always + hover-reveal
  Scenario: Vista limpia sin foco
    Given el arnés Luana sin hover
    When se renderiza
    Then se dibujan solo los edges tipo invoca (11 en Luana)    # shot1
    And no se dibuja ningún edge lee ni escribe

  Scenario: Hover revela lecturas/escrituras del nodo
    Given hover sobre "builder"
    Then aparecen sus edges lee hacia "api-mcp" y "react-expert"
    And el backbone invoca no-tocado baja a opacity 0.2
    And los nodos no vecinos bajan a opacity 0.4               # shot3
```

---

## 6 · Región Base — reglas colapsables · knowledge · bandas — RF-40…

Detalle visual (subbands, chips, splits) → `design.md §8`.

- **RF-40.** La banda `base` DEBE partirse en (a) subband **Reglas** colapsable y (b) subband **Knowledge
  & servicios** estática. `mockup:363-394`.
- **RF-41 (reglas colapsables).** La subband Reglas DEBE **colapsar por defecto** (`▸`) y expandirse al
  click (`▾`), porque un arnés real trae muchas reglas (46 en el caso real, `UX.md:522`). `mockup:368,381`
  · shot4/shot5.
- **RF-42 (split de activación).** Expandida, la subband Reglas DEBE separar las reglas en dos grupos:
  **«siempre en contexto (CLAUDE.md)»** (tono `--crit`, reglas `alw:true`) y **«carga condicional
  (paths:)»** (tono `--warn`, reglas `alw:false`). El header muestra conteos «N siempre» / «N condicional».
  `mockup:369-372,378-379`. Norma path-scoped = doctrina (`METODOLOGIA.md:120-122`); el eje `alw`
  por-nodo es **PROPUESTA** (§2.3).
- **RF-43 (knowledge leído).** La subband Knowledge & servicios (mcp/no-reglas) DEBE mostrar el chip
  **«leído por skills»** (tono `--c-knowledge`), sin flecha de colapso. `mockup:389-392`.
- **RF-44 (chip de activación por banda).** Cada banda transversal DEBE mostrar su chip de activación:
  `libreria-expertos·meta-harness·terceros`→«bajo demanda»; `marcas-dormidas`→«dormida · 0 corridas»
  (banda atenuada `opacity .55`). `mockup:231-238,354`. Honestidad = **activación por-nodo** es PROPUESTA
  (`UX.md:534`), no doctrina cementada.
- **RF-45 (redibujo tras colapso).** Al colapsar/expandir Reglas, los edges DEBEN redibujarse (endpoints
  cambian de posición/visibilidad). `mockup:385,471`.

```gherkin
Feature: Base con reglas colapsables
  Scenario: Colapsada por defecto
    Given el arnés Luana con 10 reglas
    When se abre el Mapa
    Then la subband "Reglas" está colapsada (▸)
    And su header muestra "5 siempre" y "5 condicional"        # shot4

  Scenario: Expandir muestra el split
    When se hace click en el header "Reglas"
    Then la flecha pasa a ▾
    And aparecen 2 grupos: "siempre en contexto (CLAUDE.md)" y "carga condicional (paths:)"
    And se redibujan los edges                                 # shot5
```

---

## 7 · Interacciones (navegación) — RF-50…

Detalle visual de controles → `design.md §10`.

- **RF-50 (overview-first).** Al abrir (y al cambiar de fixture), el Mapa DEBE hacer **fit**: encuadrar
  todo el arnés (`z=min((vw-48)/cw,(vh-48)/ch,1)`, `ox=max(24,(vw-cw·z)/2)`, `oy=24`). `mockup:453-459,517`
  · medido `scale(0.906)` en Luana@1680×1000.
- **RF-51 (zoom).** Botones **+** / **−** cambian z en ±0.1 (clamp `[0.4, 2]`); **⤢** = re-fit.
  `ctrl+rueda` hace zoom. `mockup:500-508`.
- **RF-52 (pan).** Arrastrar el fondo del viewport desplaza (`ox/oy`); no arrastra si el mousedown cae en
  un control/nodo/panel. `mockup:505-507`.
- **RF-53 (hover-reveal).** §5 RF-32/33.
- **RF-54 (colapsar reglas).** §6 RF-41.
- **RF-55 (toggle de fixture).** Los botones «Luana» / «dev-full-cycle» cambian el arnés renderizado y
  re-fit; `aria-pressed` marca el activo. `mockup:515` · shot1↔shot7. *(En producción, este toggle será el
  **picker de arneses** de Hito 2, §9.2; en el mockup son 2 fixtures de demo.)*
- **RF-56 (panel de ayuda).** El FAB «?» abre/cierra el panel de leyenda (relaciones, clases, caja,
  regiones, activación, origen, navegación). `mockup:513` · shot6.

```gherkin
Feature: Navegación del lienzo
  Scenario: Overview al abrir
    Given un arnés recién cargado
    When se renderiza
    Then el zoom se ajusta para encuadrar todo el arnés (fit)

  Scenario: Toggle de arnés re-encuadra
    Given la vista en "Luana"
    When se pulsa "dev-full-cycle"
    Then se renderiza el arnés dogfood y se hace fit de nuevo   # shot7
    And el botón "dev-full-cycle" queda aria-pressed=true
```

---

## 8 · Capas — solo Estructura en el MVP

- **RF-60.** El MVP DEBE exponer **únicamente** la capa **Estructura** (tipo=color+forma+etiqueta+handle+
  edges+transición). Las capas **Tokens · Desempeño · Proceso** están **staged** (deshabilitadas), a la
  espera de telemetría JSONL (indexer real). Doctrina define 4 capas **conmutables** sobre una geografía (`VISION.md:186-187`);
  el staging es decisión de plan justificada por honestidad (`METODOLOGIA.md:203-205`, «— cuando no está
  medido»). → `architecture.md §8`.
- **RF-61 (arnés recién nacido).** Cuando no hay telemetría, toda métrica se muestra **«—»**, jamás un
  falso 0. UX firmada lo permite (estado «arnés recién nacido»). `METODOLOGIA.md:203-205`.

```gherkin
Feature: Capa Estructura única
  Scenario: Capas de telemetría deshabilitadas
    Given un arnés sin telemetría JSONL
    Then la capa Estructura está activa
    And Tokens, Desempeño y Proceso aparecen deshabilitadas con tooltip "Necesita telemetría"
    And ninguna métrica numérica se inventa (se muestra "—")
```

---

## 9 · Split Hito 1 vs Hito 2

### 9.1 · Hito 1 — Mapa read-only navegable

Todo lo de §3-§8 **sin** inspector: cargar el grafo del arnés real, renderizar las 3 regiones + nodos +
variantes + edges spine-always/hover-reveal + Base colapsable + navegación (fit/zoom/pan/toggle/ayuda).
Backend: loader del dogfood → índice → `GET /api/harnesses/dev-full-cycle/graph` (hoy sirve un demo
distinto, `architecture.md §4`). FE: `entities/arnes` + `widgets/map-canvas` + swap del `ComingSoon`.

- **RF-70.** El Mapa de Hito 1 DEBE servir el arnés **`dev-full-cycle`** real (5 nodos, 4 edges, spine
  idea→released), no el demo hardcoded actual. → `architecture.md §4`.

### 9.2 · Hito 2 — inspector S3 + picker

- **RF-71 (inspector).** Al **click** en un nodo, DEBE abrirse un **Inspector** (superficie S3, `UX.md:58`)
  con el `contract` fusionado del nodo (why·capabilities·arquetipo·perfil·necesita·entrega·gate·handoff).
  Backend: `getNode` (hoy 501, `router.go:79-81`). → `architecture.md §4`.
- **RF-72 (picker).** Un selector de arneses (reemplaza el toggle de fixtures) alimentado por
  `listHarnesses` (hoy 200 vacío, `router.go:61-63`). → `architecture.md §4`.
- **RF-73.** El inspector es **read-only** en el MVP (editar = Hito 3+).

```gherkin
Feature: Inspector (Hito 2)
  Scenario: Abrir el contrato de una caja
    Given el Mapa de un arnés real
    When se hace click en la caja "spec-writer"
    Then se abre el Inspector con su contract (why, necesita, entrega, gate)
    And no se ofrece edición                                   # read-only MVP
```

---

## 10 · Las 6 decisiones doctrinales (estado de respaldo)

Firmadas por el operador (Gate 1, [[hs-09-mapa-decisiones-doctrina]]); se **cementan as-code** en
`architecture.md §5`. Estado de respaldo verificado `archivo:línea`:

| # | Decisión | Respaldo | Estado |
|---|----------|----------|--------|
| 1 | Región inferior = **Base** (no «Soporte») | `VISION.md:81-83` (A6) + `:184-186` | **Doctrina** (canónico) |
| 2 | Honestidad = **activación por-nodo** (siempre/condicional/demanda/leído/dormida) | honestidad `METODOLOGIA.md:203-205`; MCP deferred `mcp.md:78`; display `UX.md:534` | **PROPUESTA** (principio doctrinal; el facet `alw` no está en schema) |
| 3 | Facet **`origen`** (estandar vs del-puesto) | — (sólo `procedencia`/`banda`/`canal` existen) | **PROPUESTA** (doctrina SILENTE) |
| 4 | Knowledge = **as-code path-scoped** + índice semántico **opcional** | path-scoped `METODOLOGIA.md:120-122,§8.6:353-357`; semántico OPEN `VISION.md:244` | **Doctrina** (norma) + **PROPUESTA** (semántico) |
| 5 | Fase «descubrimiento» = **dato del arnés**, no doctrina | grafo agnóstico `VISION.md:19,164,235` + `METODOLOGIA.md:229` (load-bearing); base-captura `VISION.md:26-28` (contexto, no ancla) | **Doctrina** (agnosticismo) |
| 6 | Paquete = **artefacto/spine**; caja posee **1 transición** | `METODOLOGIA.md:160,191` (§3) + `VISION.md:66-73` (A2/A3) + `§8.3:306` | **Doctrina** |

> Regla de honestidad del paquete: donde el estado es **Doctrina** se afirma como establecido; donde es
> **PROPUESTA** se pinta como propuesta a ratificar (visible en el panel de ayuda del propio mockup:
> «Activación (propuesta)», «Origen (propuesta)», `mockup:188-191`).

---

## 11 · Criterios de aceptación globales (Gherkin resumen)

```gherkin
Feature: Mapa MVP — aceptación integral
  Background:
    Given el sustrato HTML bandas/carriles + overlay SVG (no React Flow)
    And solo la capa Estructura activa

  Scenario: Reconstrucción pixel-idéntica
    Given los tokens de web/tokens/base.tokens.json y las medidas de design.md
    When se implementa el Mapa
    Then coincide con los screenshots shot1..shot7 en light y dark
    And la consola queda limpia

  Scenario: Honestidad de lo no medido
    Then ninguna métrica numérica se inventa
    And las decisiones PROPUESTA se etiquetan como propuesta, no como establecidas
```

---

## 12 · Trazabilidad (requisito → mockup + screenshot)

| Requisito | Elemento | `mockup:línea` | shot# | design.md |
|-----------|----------|----------------|-------|-----------|
| RF-10..14 | 3 regiones + headers | `:54-60,406-446` | 1,2 | §3 |
| RF-20,21 | nodo base + handle | `:91,109-116,327-351` | 1 | §5 |
| RF-22,23 | caja + transición | `:104-105,114-115` | 1,7 | §6.1 (caja) · §5.1 (transición) |
| RF-24,25 | apoyo + caja-first | `:95-96,424-429` | 1 | §6.3,§7 |
| RF-26 | compacto (Guardia) | `:99-102,412` | 1,7 | §6.4 |
| RF-27,28 | origen + propuesto | `:107-108,229,278` | 1 | §6.2 |
| RF-30..35 | edges + hover-reveal | `:239-243,460-497` | 1,3 | §9 |
| RF-40..45 | Base colapsable | `:363-394` | 4,5 | §8 |
| RF-50..56 | navegación | `:451-517` | 1,6,7 | §10 |
| RF-60,61 | capa Estructura | `:196` (comentario) | 1 | — |
| RF-70 | dogfood real | `:308-323` | 7 | §7 |
| RF-71,72 | inspector + picker | *(Hito 2, no en mockup)* | — | — |

---

## Apéndice · Índice cruzado

- UI al pixel de cada elemento → [`design.md`](./design.md).
- Implementación (hexagonal/FSD, loader backend, sync Storybook, cementar las PROPUESTAS) →
  [`architecture.md`](./architecture.md).
- Decisiones firmadas y su respaldo → §10 + [`architecture.md §5`](./architecture.md).
