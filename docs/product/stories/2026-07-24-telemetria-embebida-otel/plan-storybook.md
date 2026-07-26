# Plan de stories — capa «Mejora» del Mapa + tarjeta del Portafolio

> `tipo: plan-pruebas` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> Par de [`spec.md`](./spec.md) (RF-232…RF-286) y [`design.md`](./design.md).
> **En este repo la story ES el test** (`story = test`, `docs/architecture/boundaries/fe-visual-fitness.md`,
> job `visual-fitness` de CI). Este documento define **qué stories escribir y qué asserta cada una**.
> El constructor lo ejecuta sin decidir nada; donde hay una decisión pendiente está marcada como tal.
>
> **Dato de entorno verificado hoy (2026-07-26):** `cd web && npx vitest run --project=storybook
> <archivo>.stories.tsx` corre **headless y pasa** — verificado contra `map-bar.stories.tsx`:
> `1 passed, 3.97 s`. `web/vitest.config.ts` tiene `headless: true` explícito. La nota vieja del repo
> («vitest-browser no corre en background, Chromium no headless») es **falsa y queda derogada**;
> `design.md` §8 y `spec.md` §3 (fila T-21) la repiten y hay que corregirlos en el mismo commit.
>
> **Orden de construcción (spec.md §Orden obligatorio):** los RF sin superficie (§H, RF-282…RF-286)
> van **antes** que cualquier píxel. Este plan cubre la capa visual; §4 dice con qué se verifica lo otro.
>
> ⚠️ **Ningún RF 🎨 se construye antes del 🧑‍⚖️ del mockup** (sigue en iteración 1). Este plan se puede
> escribir y revisar ya; su ejecución espera esa firma.

---

## 0 · Resumen ejecutable

| | |
|---|---|
| Stories planificadas | **124** en **17 archivos** (12 nuevos · 5 supersets de archivos firmados) |
| Componentes nuevos | 11 (4 en `entities/telemetria`, 6 en `widgets/map-canvas`, 1 en `widgets/portafolio`) + 1 promoción a `shared/ui` |
| Componentes extendidos | 6 (`arnes-node` · `map-bar` · `lane` · `inspector` · `map-canvas` · `portafolio-list`) |
| RF 🎨 con cobertura de story | 44 de 50 |
| RF 🎨 **sin** cobertura de story | 6 → §4.2 (mecanismo alternativo declarado, ninguno omitido en silencio) |
| RF sin superficie (§H) | 5 → §4.1 (Go: tabla, integración, conformance, E2E) |
| Riesgos a11y detectados | 3 duros (§2.18), 2 de ellos **rompen el gate hoy** con los tokens que `design.md` §4.2 elige |
| Bloqueantes de arquitectura | 3 (§1.3) — hay que resolverlos **antes** de escribir la primera story |

---

## 1 · Inventario de componentes

### 1.1 Nacen

| Componente | Ruta exacta | Capa y por qué ahí | Nuevo / superset | RF |
|---|---|---|---|---|
| `CifraUSD` | `web/src/entities/telemetria/ui/cifra-usd.tsx` | **entities** — el formato de dinero es dominio (RF-281 exige *un* formateador para las cinco superficies). No puede vivir en `shared/ui` porque `ui-not-domain` (depcruise, `error`) prohíbe que `shared/ui/**` importe `entities/*/model` | nuevo | RF-281 |
| `MarcaConfianza` | `web/src/entities/telemetria/ui/marca-confianza.tsx` | **entities** — los 4 casos de `atribucion_confianza` son un tipo del dominio, no un primitivo de chrome | nuevo | RF-242 · RF-280 |
| `BarraCobertura` | `web/src/entities/telemetria/ui/barra-cobertura.tsx` | **entities** — se dibuja a partir de `Cobertura`, tipo del dominio | nuevo | RF-237 · RF-279 |
| `Sparkline` | `web/src/entities/telemetria/ui/sparkline.tsx` | **entities** — 5 puntos de una serie del dominio + texto equivalente | nuevo | RF-266 · RF-279 |
| `FranjaMejora` | `web/src/widgets/map-canvas/ui/franja-mejora.tsx` | **widgets** — es *chrome* del Mapa (mismo criterio que `MapBar`: el canvas no sabe que existe). Vive en la slice `map-canvas` y no en una slice propia para no chocar con `no-sibling-widget-imports` (depcruise, `error`) | nuevo | RF-234…237 · RF-269…275 |
| `PuntosMejoraList` | `web/src/widgets/map-canvas/ui/puntos-mejora-list.tsx` | **widgets**, misma slice — resuelve **H-1**: la lista vive debajo del canvas, no dentro | nuevo | H-1 · H-2 · RF-246 |
| `PuntoMejoraCard` | `web/src/widgets/map-canvas/ui/punto-mejora-card.tsx` | **widgets**, misma slice — organismo de app compuesto de piezas de `entities/telemetria` | nuevo | RF-246…257 |
| `InspectorMejora` | `web/src/widgets/map-canvas/ui/inspector-mejora.tsx` | **widgets**, misma slice — es el cuerpo de la 4ª tab del inspector, que ya vive ahí | nuevo | RF-258…264 |
| `PoliticaDatosDialog` | `web/src/widgets/map-canvas/ui/politica-datos-dialog.tsx` | **widgets**, misma slice — se abre desde la franja (**H-5**) | nuevo | RF-275 · RF-283 |
| `TablaMejoraPortafolio` | `web/src/widgets/portafolio/ui/tabla-mejora-portafolio.tsx` | **widgets**, slice `portafolio` — las 3 columnas nuevas de la fila. Slice distinta de `map-canvas` a propósito: son dos vistas y `pages/shell` las compone | nuevo | RF-265…268 · H-12 |
| `Skeleton` · `ErrorBody` | `web/src/shared/ui/estado-carga.tsx` | **shared/ui** — promoción de los privados de `portafolio-list.tsx:280-307`. Legal porque son puramente presentacionales (props: `label`, `motivo`, `onReintentar`), cero dominio ⇒ no violan `ui-not-domain`. Los consumen el Portafolio (igual que hoy) y la franja/canvas (**H-6**) | promoción (sin cambio de conducta) | H-6 |

**Sin story** (no son componentes): `entities/telemetria/model/types.ts` ·
`entities/telemetria/model/selectors.ts` (se cubre con `model/selectors.test.ts`, project `unit`) ·
`entities/telemetria/testing/telemetria.ts` (§3) · `entities/telemetria/index.ts` ·
`widgets/map-canvas/model/layers.ts` · `pages/shell/ui/workspace-stage.tsx` y
`portafolio-view.tsx` (capa `pages`: transporte, sin story en este repo — hoy hay 0 stories de `pages/`).

### 1.2 Se extienden (superset estricto, BR-M16)

| Archivo | Qué gana | Nuevo / superset | RF |
|---|---|---|---|
| `web/src/entities/arnes/ui/arnes-node.tsx` | 3 hijos al **final del flujo** (`.mej-cifra` · `.mej-fuga` · `.mej-share`) + línea de motivo cuando no es caja. Sin las props nuevas el DOM es **idéntico** al de hoy | superset | RF-238…243 · RF-280 |
| `web/src/widgets/map-canvas/ui/map-bar.tsx` | Slot `Mejora` encendido; `title` fijo → `l.motivo`; `aria-describedby` → `<span class="sr-only">` | superset | RF-232 · RF-233 · RF-276 |
| `web/src/widgets/map-canvas/ui/lane.tsx` | Tercer hijo en `.lane-hd`: el total de la fase. **`<span className="count">` se conserva** (J-8) | superset | RF-244 |
| `web/src/widgets/map-canvas/ui/inspector.tsx` | Cuarta entrada en `TABS` (línea 111) + cuarto `tabpanel` con el mismo contrato ARIA | superset | RF-258 · RF-277 |
| `web/src/widgets/map-canvas/ui/map-canvas.tsx` | Prop `capa` + `mejora` (mapa `nodeId → CifraCaja`) que pasa a los nodos. **Cero rama de layout** | superset | RF-245 |
| `web/src/widgets/portafolio/ui/portafolio-list.tsx` | 3 celdas opcionales en `Fila`, insertadas **entre chips y dot de salud** | superset | RF-265…268 |

### 1.3 🔴 Tres bloqueantes de arquitectura — resolver ANTES de la primera story

Ninguno es opinable a mano alzada: los tres los caza un enforcer que ya está en `error`.

**BLOQ-1 · `entities/arnes` → `entities/telemetria` es un cross-import de entities.**
`design.md` §1.2 dice que `arnes-node.tsx` gana `mejora?: CifraCaja`. `CifraCaja` vive en
`entities/telemetria`. `dependency-cruiser` **no** tiene regla de hermanos para `entities` (solo
`no-sibling-feature-imports` y `no-sibling-widget-imports`), pero `steiger` **sí**
(`fsd/no-cross-imports`), corre en `pnpm --dir web run fsd`, y `fsd` está dentro de `verify`. El
único escape configurado es `entities/*/@x/**`, y el propio `steiger.config.ts` dice que ese scope
**puede no bastar** («si reporta en el consumidor, este scope no basta … pendiente»).

| opción | consecuencia |
|---|---|
| **A (recomendada)** — `ArnesNode` recibe **primitivos**, cero import: `cifraUSD?: string` · `participacionPct?: number` · `confianza?: "exacta"\|"por-hash"\|"por-proceso"\|"sin-dato"` · `marcaFuga?: string` · `motivoSinDato?: string`. El widget `map-canvas` (que **sí** puede importar las dos entities) compone `CifraCaja → props` | cero riesgo de enforcer. Costo: el literal del chip de confianza y el formato de dinero quedan en dos sitios ⇒ se ata con la story `CopyConfianzaEsUnaSola` (§2.13), que vive en el widget y asserta igualdad literal contra `etiquetaConfianza()` de `entities/telemetria` |
| B — `entities/telemetria/@x/arnes.ts` | usa el escape que el config anticipa, pero el config declara su enforcement **incierto**. Si steiger reporta en el consumidor, `verify` se pone rojo y hay que apagar la regla — degradar un gate por una prop es mal negocio |
| C — mover las marcas a un overlay del widget sobre el nodo | rompe `§2.3` de `design.md` («todo va en el flujo de la tarjeta») y reintroduce el posicionamiento absoluto que J-9 prohibió |

**Se ejecuta A.** Marcado **PROPUESTO** — necesita el visto del arquitecto en `decisiones.md` antes de codear.

**BLOQ-2 · `Clase` no tiene `knowledge`.**
`spec.md` RF-238 nombra «rule, mcp, **knowledge**, subagent, hook, command o settings» y `design.md`
§7.3 fija un copy `sin dato atribuible — el conocimiento se paga en la caja que lo carga`. Pero
`web/src/entities/arnes/model/types.ts` define `Clase` como
`skill|subagent|hook|rule|command|mcp|plugin|settings|output-style|statusline|no-reconocido` —
**no existe `knowledge`**. En este árbol «conocimiento» es una *banda* (`selectBase`) y un token
(`--c-knowledge`), no una clase de nodo.
⇒ El copy de `conocimiento` **no tiene nodo al que colgarse**. Dos salidas: (a) borrar la fila de
`design.md` §7.3 y quedarse con `sin dato · resto`; (b) agregar la clase a `Clase` — que es tocar
el contrato L0 y **no** es alcance de este paquete. **Se propone (a)**; la story
`SinDatoResto` (§2.5) cubre el hueco con el copy genérico.

**BLOQ-3 · el Portafolio no tiene `puesto`.**
RF-265 pide «una fila por arnés × **puesto**» y `design.md` §3.5 lo lista como
`identidad.id + puesto de la instalación`. Pero `EntradaPortafolio`
(`web/src/entities/portafolio/model/types.ts`) **no tiene ningún campo `puesto`**: tiene
`empresas?: string[]` y `instalaciones[].proyecto_path`. J-4 ya declaró que «puesto» es el
vocabulario firmado y que el `MetaChip` vigente **no** se renombra.
⇒ El dato no existe en el wire del Portafolio. Hasta que exista, la fila se agrupa por
`(identidad, instalacion)` y la celda de puesto muestra el copy de ausencia que `design.md` §3.5 ya
prevé: `puesto sin declarar`. La story `SinPuestoDeclarado` (§2.15) lo cementa para que nadie
fabrique un puesto. **Deuda a registrar en `BACKLOG.md`.**

---

## 2 · El plan de stories

Convenciones que valen para todas, calcadas de las stories firmadas
(`map-canvas.stories.tsx` = *fitness fixture of record*, `portafolio-list.stories.tsx`,
`inspector.stories.tsx`):

- `import { expect, fn, userEvent, within } from "storybook/test"`, `within(canvasElement)` en `c`.
- Preferencia de query: `getByRole` → `getByLabelText` → `getByText` → `querySelector` **solo** para
  conteos estructurales (`.pf-fila`, `.artchip`, `.node`), como ya se hace hoy.
- Toda ausencia se asserta con `queryBy…` + `toBeNull()` / `not.toBeInTheDocument()` — nunca «no lo
  vi». Es el patrón G2 del Portafolio y el que hace verificable «jamás un cero».
- `title` de la story = `<capa>/<slice>/<Componente>` (p. ej. `entities/telemetria/CifraUSD`).
- Cada story lleva un comentario de 1-3 líneas que cita el RF y el porqué, como hoy.
- **`parameters: { a11y: { test: "error" } }`** por default (viene de `.storybook/preview.ts`).
  Cualquier `a11y: { test: "todo" }` nuevo tiene que justificarse por escrito en el comentario y
  aparecer en §2.18 — no se baja el gate en silencio.

### 2.1 `entities/telemetria/ui/cifra-usd.stories.tsx` — 4 stories

| story | escenario | asserts | RF |
|---|---|---|---|
| `Estandar` | monto normal | `c.getByText("USD")` presente · `c.getByText("1,92")` presente · el nodo raíz tiene `class` con `num` y `getComputedStyle(el).fontVariantNumeric` contiene `tabular-nums` | RF-281 |
| `SeparadorYDosDecimales` | monto con millar | renderiza `USD 1 234,56` — `c.getByText(/1[\s ]234,56/)` · `c.queryByText(/1,234\.56/)` es `null` (nunca formato EN) | RF-281 |
| `MenorAlCentavo` | `0,004` | `c.queryByText("0,00")` es **`null`** · el texto visible contiene `0,004` (no se redondea a cero) | RF-281 |
| `Ausente` | sin monto | `c.getByText("sin dato")` · `c.queryByText("USD")` es `null` · `c.queryByText("0,00")` es `null` · `c.queryByText("—")` es `null` | RF-281 · BR-M2 |

### 2.2 `entities/telemetria/ui/marca-confianza.stories.tsx` — 5 stories

| story | escenario | asserts | RF |
|---|---|---|---|
| `Exacta` | `confianza="exacta"` | **el componente no renderiza nada**: `canvasElement.querySelector("[data-confianza]")` es `null` · `c.queryByText(/por huella\|por proceso\|aproximad/i)` es `null`. *La ausencia de marca **es** la señal* (design §5.3) | RF-242 |
| `PorHash` | `confianza="por-hash"` | `c.getByText("por huella")` · ese nodo tiene `title` **exactamente** `"Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA."` · `toHaveClass` que incluya la marca de subrayado punteado | RF-242 · RF-274 |
| `PorProceso` | `confianza="por-proceso"` | `c.getByText("por proceso")` · `title` **exactamente** `"Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla."` · el `title` **difiere** del de `PorHash` (assert de desigualdad contra la constante importada) | RF-242 · H-13 |
| `SinDato` | `confianza="sin-dato"` | no hay cifra: `c.queryByText(/USD/)` es `null` · `c.queryByText("0,00")` es `null` · hay texto de motivo | RF-242 |
| `LosCuatroJuntos` | los 4 en fila | `c.getAllByText("por huella")` longitud 1 · `getAllByText("por proceso")` longitud 1 · **`c.queryByText(/aproximad/i)` es `null`** (ninguno se pinta como «atribución aproximada») · los 4 `title` del DOM son 4 strings distintos (`new Set(titles).size === 3`, porque `exacta` no tiene) | RF-242 · H-13 |

### 2.3 `entities/telemetria/ui/barra-cobertura.stories.tsx` — 5 stories

| story | escenario | asserts | RF |
|---|---|---|---|
| `CuatroSegmentos` | 12 exactas · 3 huella · 2 proceso · 1 sin dato | `canvasElement.querySelectorAll(".cov-seg").length` es **4** · anchos proporcionales: cada segmento tiene `style.width` cuyo % ≈ `n/18` (±1 pp) · `c.getByText("12 exactas · 3 por huella · 2 por proceso · 1 sin dato — sobre 18 corridas")` (literal de design §7.2) · `c.getByRole("img")` con `aria-label` **exacto** `"Cobertura de la atribución: 12 corridas exactas, 3 por huella, 2 por proceso, 1 sin dato, sobre 18 corridas."` | RF-237 · RF-279 · H-8 |
| `CategoriaEnCeroNoOcupaLugar` | 0 por proceso | `.cov-seg` longitud **3** · `c.queryByText(/por proceso/)` es `null` (la categoría en cero **no** se nombra) | RF-237 |
| `CoberturaCompleta` | todas exactas | `.cov-seg` longitud **1** · `c.getByText("atribución exacta en las 18 corridas")` | RF-237 |
| `SinCorridas` | 0 corridas | `canvasElement.querySelector(".cov-bar")` es `null` (la barra no se dibuja) — el estado 1 manda | RF-237 · RF-269 |
| `RotuloVisibleSiempre` | ancho 320 px (envuelto) | `await expect(c.getByText("cobertura")).toBeVisible()` — el rótulo no depende del espacio; la barra nunca queda huérfana | H-11 · RF-279 |

### 2.4 `entities/telemetria/ui/sparkline.stories.tsx` — 4 stories

| story | escenario | asserts | RF |
|---|---|---|---|
| `EnAlza` | 5 puntos ascendentes | `c.getByRole("img", { name: "Tendencia en alza en las últimas 5 corridas." })` · 5 barras (`querySelectorAll(".spark-bar").length === 5`) · la última tiene una clase/`data-` que la distingue de las 4 anteriores | RF-266 · RF-279 |
| `Estable` | 5 puntos planos | `getByRole("img", { name: /estable/ })` | RF-266 |
| `ALaBaja` | 5 puntos descendentes | `getByRole("img", { name: /a la baja/ })` | RF-266 |
| `PocasCorridas` | 1 punto | `canvasElement.querySelector(".spark")` es `null` · `c.getByText("pocas corridas para una tendencia")` | RF-266 |

### 2.5 `entities/arnes/ui/arnes-node.stories.tsx` — **+13 stories** (superset del archivo firmado)

> El archivo ya tiene 14 stories firmadas (`Skill`, `Caja`, `CajaGateNone`, …). **Ninguna se toca.**
> Las nuevas usan las props primitivas de BLOQ-1 opción A.
> El archivo hereda `parameters: { a11y: { test: "todo" } }` del `meta` — está justificado en el
> comentario vigente (micro-UI de bajo contraste del diseño firmado) y **no se amplía** por este
> paquete: las marcas nuevas se pintan con `--foreground`, no con `--warn` (§2.18).

| story | escenario | asserts | RF |
|---|---|---|---|
| `MejoraCifraExacta` | caja con `cifraUSD="1,92"`, `participacionPct=40`, `confianza="exacta"` | `c.getByText("USD 1,92")` · `c.getByText("40 %")` · la barra de pie existe con `style.width === "40%"` y `getByRole("img", { name: "40 % del gasto del arnés en esta ventana." })` · **`c.queryByText("por huella")` es `null`** · las 5 marcas de hoy siguen (`caja`, `idea → spec`, `T2`, `≈`, gate) | RF-238 · RF-239 · RF-240 · RF-242 |
| `MejoraPorHuella` | `confianza="por-hash"` | `c.getByText("por huella")` · la cifra sigue presente y **suma** (`getByText("USD 1,92")`) · el chip tiene su `title` largo | RF-242 · RF-274 |
| `MejoraPorProceso` | `confianza="por-proceso"` | `c.getByText("por proceso")` · `title` con la advertencia de mezcla (`/los mezcla/`) · `c.queryByText("por huella")` es `null` | RF-242 · H-13 |
| `MejoraCajaSinCorridas` | caja, ventana vacía | `c.getByText("sin corridas en esta ventana")` · `c.queryByText(/USD/)` es `null` · `c.queryByText("0,00")` es `null` · **no hay barra**: `querySelector(".mej-share")` es `null` | RF-240 · RF-243 |
| `MejoraConMarcaDeFuga` | detector B1 | `c.getByText("re-warm TTL")` · **no** hay icono suelto: `c.queryByText("⚠")` es `null` o el `⚠` tiene `aria-hidden="true"` **y** hay texto hermano | RF-241 |
| `MejoraDosDetectores` | B1 (ahorro mayor) + B3 | se pinta **una sola** marca: `querySelectorAll(".mej-fuga").length === 1` · es la de mayor ahorro: `getByText("re-warm TTL")` · `queryByText("modelo cambiado")` es `null` | RF-241 |
| `SinDatoSubagente` | `clase="subagent"` (`test-author` de Luana) | `c.getByText("sin dato atribuible — el subagente no se distingue en el turno")` · `queryByText(/USD/)` `null` · `queryByText("0")` `null` · el nodo tiene la clase `sindato` | RF-243 · BR-M2 |
| `SinDatoRegla` | `clase="rule"` (`std-spec`) | `c.getByText("sin dato atribuible — una regla no consume por sí misma")` + las mismas 3 ausencias | RF-243 |
| `SinDatoMcp` | `clase="mcp"` (`api-mcp`) | `c.getByText("sin dato atribuible — el costo del MCP está incluido en la caja que lo llama; todavía no se desglosa")` — **el literal corregido por H-10**, no el del mockup `:404` | RF-243 · H-10 |
| `SinDatoHook` | `clase="hook"` (`hook-stop`) | `c.getByText("sin dato atribuible — un hook informa, no consume")` | RF-243 |
| `SinDatoResto` | `clase="command"` | `c.getByText("sin dato atribuible — esta capa mide cajas")` · cubre también el hueco de BLOQ-2 | RF-243 |
| `MejoraDesbordamiento` | nombre de 90 caracteres + `cifraUSD="123 456,78"` + `participacionPct=100` | ninguna marca desborda la celda de 232 px: para `.mej-cifra`, `.mej-fuga` y la línea de motivo, `el.scrollWidth <= el.clientWidth + 1` · el motivo **envuelve**, no se trunca: `getComputedStyle(el).textOverflow !== "ellipsis"` (design §6.3) | design §6.3 |
| `EstructuraIntacta` | **sin ninguna prop de mejora** | el DOM es el de hoy: `querySelector(".mej-cifra")` `null` · `.mej-fuga` `null` · `.mej-share` `null` · `queryByText(/sin dato atribuible/)` `null`. Es el guardián de BR-M16: la capa apagada no deja rastro | RF-245 · BR-M16 |

### 2.6 `widgets/map-canvas/ui/map-bar.stories.tsx` — **+4 stories**

> ⚠️ La story `Default` vigente asserta `getByRole("tab", { name: "Desempeño" })).toBeDisabled()` —
> eso **sigue siendo cierto** y no se toca. No hay assert vigente sobre «Tokens», así que el rename
> no rompe nada en este archivo.

| story | escenario | asserts | RF |
|---|---|---|---|
| `CapaMejoraDisponible` | capa `estructura` | `c.getAllByRole("tab")` longitud **4** · los nombres accesibles en orden son exactamente `["Estructura","Mejora","Desempeño","Proceso"]` (`.map(t => t.textContent)`) · **`c.queryByRole("tab", { name: "Tokens" })` es `null`** · `getByRole("tab", { name: "Mejora" })` **no** está `disabled` | RF-232 |
| `CapaMejoraActiva` | capa `mejora` | `getByRole("tab", { name: "Mejora" })` tiene `aria-selected="true"` · `getByRole("tab", { name: "Estructura" })` tiene `aria-selected="false"` · click en `Estructura` llama `args.onCapa` con `"estructura"` | RF-232 · RF-276 |
| `MotivosHonestos` | slots apagados | el `title` de `Desempeño` es **exactamente** `"La señal ya llega —duración por request y por herramienta—. Falta decidir qué es «desempeño» a nivel Mapa."` · el de `Proceso`, `"Entra en parte por la capa Mejora. Falta mapear todos los eventos a fases del arnés."` · **`c.queryByText(/Necesita telemetría/)` es `null`** y ningún `title` del DOM lo contiene | RF-233 |
| `MotivoAccesible` | slots apagados | `getByRole("tab", { name: "Desempeño" })` tiene `aria-describedby` · el elemento referenciado existe, tiene clase `sr-only` y su `textContent` es el mismo motivo que el `title` (assert de igualdad, para que no driften) | RF-276 |

### 2.7 `widgets/map-canvas/ui/lane.stories.tsx` — **+2 stories**

| story | escenario | asserts | RF |
|---|---|---|---|
| `LaneConTotalMejora` | fase `spec`, total `1,92` | dentro de `.lane-hd` hay **3** hijos · el `<h3>` sigue con el nombre de fase · **`querySelector(".lane-hd .count")` sigue existiendo** con su valor de hoy (J-8: no se sustituye) · `getByText("USD 1,92")` | RF-244 · J-8 |
| `LaneSinDato` | carril cuyas cajas no tienen costo | `c.getByText("sin dato")` · `c.queryByText("USD 0,00")` es `null` · el `.count` sigue presente | RF-244 |

### 2.8 `widgets/map-canvas/ui/franja-mejora.stories.tsx` — 19 stories (NUEVO)

> Dueña de los estados 1, 1b, 2, 3, 3b, 4, 5 y del resumen del 7. Los ids `S <nombre>` que ya cita
> `escenarios.md` se mapean acá: `estado-sin-datos` → `Estado1SinDatos`; `estado-otro-runtime` →
> `Estado4OtroRuntime`; `estado-catalogo-viejo` → `Estado5CatalogoViejo`; `barra-con-disclaimer` →
> `DisclaimerEnSuperficie`; `estado-runtime-no-soportado` → `RuntimeNoSoportado`.

| story | escenario | asserts | RF |
|---|---|---|---|
| `Reposo` | ventana 7 d, datos completos | orden de lectura del DOM: el `select` de ventana precede al total (`compareDocumentPosition` o índice en `querySelectorAll`) · `getByLabelText("ventana")` con valor `"7 días"` · total, disclaimer y cobertura presentes | RF-234…237 |
| `VentanaCambia` | operador elige 30 días | `userEvent.selectOptions(getByLabelText("ventana"), "30 días")` → `args.onVentana` llamado con `"30d"` · las tres opciones existen: `getAllByRole("option")` longitud 3 con textos `7 días`/`30 días`/`todo` | RF-234 |
| `TotalConDenominador` | 61 corridas · 12 sesiones · 4 cajas | `c.getByText("USD 4,82")` (o `getByText("4,82")` + `getByText("USD")` si van en spans distintos) · `c.getByText("61 corridas · 12 sesiones · 4 cajas")` literal | RF-235 |
| `DisclaimerEnSuperficie` | reposo | `const d = c.getByText("estimado por el runtime, no es facturación")` · `await expect(d).toBeVisible()` · **el disclaimer NO vive en un `title`**: `d.tagName !== "TITLE"` y `d.closest("[title]")` es `null` · no está dentro de un `[hidden]` ni de un `details` cerrado | RF-236 |
| `CoberturaCuatroNiveles` | 12/3/2/1 | delega en `<BarraCobertura>` pero asserta la integración: `.cov-seg` longitud 4 · el rótulo `cobertura` visible | RF-237 |
| `Estado1SinDatos` | arnés nunca instrumentado | `c.getByText("Este arnés nunca corrió con telemetría.")` · `c.getByText("Abrí una sesión desde ArnesIA y la medición arranca sola.")` · `c.getByText("— —")` · **`expect(canvasElement.textContent).not.toContain("0,00")`** (assert de tablero-en-cero, el que pide `escenarios.md` A8) · la barra de cobertura no se dibuja | RF-269 |
| `Estado1bSinCorridasEnVentana` | última corrida hace un mes, ventana 7 d | `c.getByText("Sin corridas en los últimos 7 días. La última fue el 12/07.")` · botón `getByRole("button", { name: "Ver todo" })` que llama `args.onVentana` con `"todo"` · el texto **difiere** del de `Estado1SinDatos`: `queryByText(/nunca corrió/)` es `null` | H-9 · T-15 |
| `Estado2CoberturaParcial` | 5 corridas, 3 atribuidas | `c.getByText("USD 3,10 — de 5 corridas, 3")` · `c.getByText("2 corridas quedaron sin atribución.")` · el `3` está en un nodo con el tono `warn` (`toHaveClass`) | RF-270 |
| `Estado3S2SinInstrumentar` | S2 degradado | `c.getByText("corrió fuera de ArnesIA, sin instrumentar")` · `queryByText(/instrumentado/)` devuelve **solo** ese match (no el rótulo del modo instrumentado) | RF-271 · J-10 |
| `Estado3bS2Instrumentado` | S2 instrumentado (H9) | `c.getByText("corrió fuera de ArnesIA, instrumentado")` · `c.getByText("Llega la misma señal que dentro de ArnesIA.")` · **las cifras de dinero se muestran igual que en S1**: `getByText(/USD/)` presente · assert cruzado: el rótulo **no** es el mismo string que en `Estado3S2SinInstrumentar` | RF-271 · J-10 · ANEXO H9 |
| `Estado4OtroRuntime` | runtime sin costo | `c.getByText("USD 0,74")` · `c.getByText("calculado con el catálogo v2026-07-20 — este runtime no reporta costo")` · el número **sí** se muestra (no se oculta por no ser reportado) | RF-272 |
| `Estado5CatalogoViejo` | catálogo del release | `c.getByText("⚠ Precios del release, sin refrescar desde el 20/07.")` · `c.getByText("Funciona sin internet; te avisamos que está funcionando así.")` · **está en la superficie**: `toBeVisible()` y `closest("[title]")` es `null` | RF-273 |
| `RuntimeNoSoportado` | arnés que corre con `codex` | `c.getByText(/todavía no medimos ese runtime/)` · `queryByText("0,00")` `null` — cubre `escenarios.md` A5 | A5 |
| `ResumenQueGuardamos` | reposo | `c.getByText("Nada de tu cuenta. Nada de la conversación.")` · `c.getByText(/Retención 90 días/)` · enlace `getByRole("button", { name: "qué guardamos" })` que llama `args.onPolitica` | RF-275 · H-5 |
| `Cargando` | consulta en vuelo | `getByRole("status")` con `aria-live="polite"` y texto `midiendo…` · **el `select` de ventana sigue habilitado**: `getByLabelText("ventana")` `not.toBeDisabled()` · la barra de cobertura no se dibuja | H-6 · design §5.2 |
| `Error` | fallo de consulta | `c.getByText("No se pudo leer la telemetría — ECONNREFUSED.")` (motivo real inyectado) · botón `Reintentar` habilitado que llama `args.onReintentar` · **la franja no desaparece**: el `select` de ventana sigue en el DOM | H-6 |
| `DaemonCaido` | daemon no responde, hay medición previa | `c.getByText("El daemon no respondió. Lo que ves es la última medición, del 25/07.")` · **las cifras previas siguen**: `getByText("USD 4,82")` presente · es distinguible de `Estado1SinDatos`: `queryByText(/nunca corrió/)` es `null` | H-6 |
| `ReenvioEncendido` | forward on | `c.getByText("reenvío externo encendido → langfuse.local")` · en `Reposo` ese chip **no** existe (assert espejo dentro de esta story sobre el default) | H-7 · D13 |
| `Angosta1024` | decorator de 1024 px | `getByText("cobertura")` visible · `document.body.scrollWidth <= document.body.clientWidth` (la página no scrollea horizontal, design §6.3) | H-11 · design §6.1 |

*(`RuntimeNoSoportado` entra como estado extra de `escenarios.md` A5, además de los 7+1 del spec.)*

### 2.9 `widgets/map-canvas/ui/punto-mejora-card.stories.tsx` — 14 stories (NUEVO)

| story | escenario | asserts | RF |
|---|---|---|---|
| `CompletaB1Atencion` | B1, severidad atención — **la tarjeta insignia** | titular literal `La caja «escribir el spec» reescribe el cache en cada corrida` · lede literal `Gasta USD 0,84 de los 1,92 de la caja (44 %) escribiendo cache que se vence antes de volver a leerse.` · los 4 `dt` presentes: `getByText("Contrafactual")`, `("Umbral")`, `("Confianza")`, `("Sesgo")` · contrafactual literal con **unidad explícita** (`/por corrida/`) · umbral literal `relectura 61 % > break-even (2−1,25)/(2−0,1) = 39,47 %` · confianza `exacta · 14 de 14 corridas con atribución` · fix `Fijar cache_ttl: 1h en esta caja` con el ajuste en `<code>` (`querySelector(".mej-fix code")` no nulo) · `getByText("score v1")` visible **sin hover** · botones `Descartar` y `Proponerlo en el chat` presentes | RF-247…254 |
| `CriticaP1SinUmbral` | P1, severidad crítica | titular literal `La caja «revisar el build» se rechaza en el gate 3 de cada 4 veces` · **`c.queryByText("Umbral")` es `null`** · en su lugar `getByText("Patrón")` con `Los 3 rechazos citan el mismo motivo: «el veredicto no lista hallazgos».` · chip de severidad `crítico` | RF-250 · RF-257 |
| `SesgoSubestima` | B1 | `getByText("Asume 0 lecturas fuera de la ventana de 7 días → subestima el ahorro.")` · la dirección está en negrita: `querySelector("dd strong")?.textContent` contiene `subestima` | RF-252 |
| `SesgoSobreestima` | P1 | `getByText("No descuenta lo que la revisión aporta aunque rechace → sobreestima el desperdicio.")` | RF-252 |
| `SinSesgoIdentificado` | detector sin sesgo | **la fila `Sesgo` SIGUE presente**: `getByText("Sesgo")` · su `dd` dice `No se identificó ningún supuesto que sesgue este cálculo.` — nunca se omite la fila | RF-252 |
| `S1Only` | tarjeta B1 en instalación S2 | `c.getByText("solo con telemetría de ArnesIA")` · el chip es de tono neutro (`toHaveClass` de chip neutro, no `warn`/`crit`) | J-2 |
| `CalculoCerrado` | reposo | `const b = getByRole("button", { name: "ver el cálculo" })` · `b` tiene `aria-expanded="false"` · el bloque referenciado por `aria-controls` está `hidden` | H-4 · RF-250 |
| `CalculoAbierto` | tras click | `userEvent.click(b)` → `b` pasa a `aria-expanded="true"` y su nombre a `ocultar el cálculo` · el bloque muestra la fórmula con los números reemplazados (`getByText(/0,75\s*\/\s*1,9/)` o el literal que fije design) + la ventana + las corridas contadas · **no es un modal**: `queryByRole("dialog")` es `null` | H-4 |
| `Resaltada` | su caja está seleccionada | `c.getByText("↔ caja seleccionada")` — el texto existe, así que la señal **no** depende del grosor del borde | design §5.4 · RF-280 |
| `DescartarLlamaHandler` | click en `Descartar` | `userEvent.click(getByRole("button", { name: "Descartar" }))` → `args.onDescartar` llamado con el id del punto · hay una región `aria-live="polite"` con `Descartado. Se puede volver a mostrar desde la tab Mejora de esa caja.` · **la cifra de la caja no viaja en el handler** (el payload es solo el id) | RF-256 |
| `ProponerAbreChatNoEscribe` | click en `Proponerlo en el chat` | `args.onProponer` llamado **una** vez con `{ puntoId, textoPropuesto }` · **no existe ninguna prop de escritura**: el `meta.args` no declara `onAplicar`/`onEscribir` — assert por tipo (el archivo no compila si existen) + `expect(Object.keys(args)).not.toContain("onAplicar")` | RF-255 · BR-M12 · D17.3 |
| `ProponerDeshabilitadoFueraDeAlcance` | arnés fuera del alcance del chat | el botón está `toBeDisabled()` · su `title` es `Este arnés está fuera del alcance del chat embebido.` · **el motivo también está en texto visible**, no solo en el `title`: `getByText(/fuera del alcance del chat embebido/)` (design §5.4) | RF-255 |
| `SeveridadSinColor` | dos tarjetas, atención y crítico | `getByText("atención")` y `getByText("crítico")` presentes como **texto** · el `⚠` decorativo tiene `aria-hidden="true"` · assert de que la diferencia no es solo el borde: los dos chips tienen `textContent` distintos | RF-257 · RF-280 |
| `TitularSinJerga` | B1 | el titular **no** nombra el id interno: `expect(titular.textContent).not.toMatch(/\bB1\b/)` · tampoco jerga del runtime: `not.toMatch(/ephemeral|cache_creation|TTL_/)` | RF-247 |

> **«Una tarjeta sin contrafactual computable» no existe como story de tarjeta, y es a propósito.**
> `spec.md` RF-246 (A4) y `design.md` §3.3 dicen que **sin contrafactual no hay tarjeta**. Pintarla
> degradada sería violar la regla que el paquete entero defiende. El escenario se verifica en dos
> sitios distintos: `DescartaSinContrafactual` (§2.10, la lista lo filtra) y
> `DetectorSinFixPropuesto` (§2.11, el inspector lo lista, no lo esconde).

### 2.10 `widgets/map-canvas/ui/puntos-mejora-list.stories.tsx` — 7 stories (NUEVO)

| story | escenario | asserts | RF |
|---|---|---|---|
| `DosTarjetasOrdenadas` | B1 (0,53) y B3 (0,21) | `querySelectorAll(".mejora").length === 2` · el **primero** en el DOM es el de mayor ahorro: `cards[0].textContent` contiene `re-warm` · encabezado `Puntos de mejora` con contador `2` | H-1 |
| `DescartaSinContrafactual` | 3 hallazgos: 2 con contrafactual, 1 sin | `querySelectorAll(".mejora").length === 2` · el hallazgo sin contrafactual **no** aparece: `queryByText(/<titular del tercero>/)` es `null` | RF-246 · A4 |
| `VaciaConDatos` | hay datos, 0 puntos | `c.getByText("Hay datos y ningún punto de mejora que pase el corte.")` · `c.getByText("Los seis detectores corrieron sobre 61 corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.")` · los 6 detectores listados: `getAllByRole("listitem")` dentro de la lista de detectores longitud **6** · **no es una sección vacía**: `querySelector(".mejora")` es `null` pero el bloque existe | H-2 |
| `Cargando` | consulta en vuelo | 2 `Skeleton` con la altura de una tarjeta: `querySelectorAll("[data-skeleton='mejora']").length === 2` · `getByRole("status")` | design §5.4 |
| `Error` | fallo | motivo textual + `getByRole("button", { name: "Reintentar" })` que llama `args.onReintentar` | design §5.4 |
| `NotaAlPie` | reposo | `c.getByText("«Proponerlo en el chat» no escribe archivos: abre el chat con el cambio propuesto, y se aplica por el camino de siempre, con sus permisos y su gate.")` | RF-255 · BR-M12 |
| `MuchasTarjetas` | 12 puntos | 12 tarjetas · la página **no** scrollea horizontal: `document.body.scrollWidth <= document.body.clientWidth` · ninguna tarjeta desborda su columna | design §6 |

### 2.11 `widgets/map-canvas/ui/inspector-mejora.stories.tsx` — 17 stories (NUEVO)

| story | escenario | asserts | RF |
|---|---|---|---|
| `Completo` | caja con 1,92 en 14 corridas | encabezado `Tokens · 7 días (14 corridas)` · **6 filas** de bucket: `entrada`, `salida`, `cache · lectura`, `cache · escritura 5 m`, `cache · escritura 1 h`, `razonamiento` · **la suma cierra**: leer los USD de las filas, sumar y comparar con `1,92` (assert aritmético, no de texto) · las 4 secciones presentes por su título | RF-259 |
| `NoAplicaNoEsCero` | runtime sin razonamiento | la fila `razonamiento` tiene una celda con `colSpan === 2` y texto `no aplica en este runtime` · **`queryByText("0")` dentro de esa fila es `null`** · nota `«No aplica» no es 0. Un cero donde el concepto no existe sería mentira.` | RF-260 · RF-286 |
| `CeroLegitimo` | `cache · escritura 1 h` = 0 (dato **real**: `ephemeral_5m: 0` de la corrida 3) | la fila muestra `0` y `0,00` en celdas separadas · **no** tiene `colSpan` · es distinguible de la fila `no aplica`: assert de que las dos filas tienen estructura DOM distinta | RF-260 · RF-286 |
| `ParidadCoinciden` | runtime 1,92 · catálogo 1,92 | `getByText("runtime USD 1,92")` · `getByText("nuestro catálogo USD 1,92")` · `getByText("✓ coinciden")` — **el veredicto es texto**, no solo tono | RF-261 · RF-280 |
| `ParidadDivergen` | 1,92 vs 2,10 | `getByText("⚠ difieren en USD 0,18")` · **los dos números siguen visibles** · `getByText("Guardamos los dos. Si difieren, o nuestro catálogo está viejo o el runtime cambió su tarifa.")` · **no elige**: el DOM contiene las dos causas y ninguna marca de «la correcta» (`queryByText(/correcto\|válido/i)` es `null`) | RF-261 |
| `SinCostoDelRuntime` | runtime que no reporta | `getByText("este runtime no reporta costo — calculado con el catálogo v2026-07-20")` · el número calculado se muestra rotulado como calculado | RF-261 · RF-272 |
| `CatalogoSinConstruir` | estado real de hoy (traza `:665`) | `getByText("catálogo sin construir")` · el costo reportado se muestra igual · sin `0,00` fabricado | design §3.4 |
| `JoinCompleto` | 14 corridas, 3 rechazadas | 4 filas: `corridas` = 14 · `rechazadas en el gate` = 3 · `costo de las rechazadas` en **USD** (`getByText(/USD 0,\d\d/)`) · `rotaciones de contexto` = n | RF-262 |
| `SinSenalDeGate` | corridas sin evento de gate | `getByText("sin señal de gate en estas corridas")` · **las filas de dinero se muestran igual**: `getByText("14")` para corridas y el bucket table intacta · `queryByText("0")` en la fila de rechazadas es `null` | RF-262 |
| `SeisDetectoresConEstado` | los 6 del MVP | `getAllByRole("listitem")` en la sección `Detectores` longitud **6** · cada uno matchea uno de tres patrones: `/· activo$/`, `/· sin hallazgos$/`, o tiene una segunda línea `/^no disponible: /` · **cada estado tiene texto**, el punto de color es refuerzo: `querySelectorAll(".det-dot[aria-hidden='true']").length === 6` | RF-263 · RF-280 |
| `DetectorB1ApagadoEnS2` | S2 degradado | `getByText(/re-warm/)` presente **y** su segunda línea es `no disponible: sin el result del stream-json — este arnés corrió fuera de ArnesIA` · **el detector NO se esconde** · los otros 5 siguen (5 sin ese motivo) | RF-271 · T-09 |
| `DetectorB2Parcial` | 2 de 3 rotaciones afuera | `getByText("no disponible: 2 de 3 rotaciones ocurrieron fuera de ArnesIA")` | RF-263 |
| `SieteNoMedidos` | los fuera del MVP | el bloque de no medidos lista **7** entradas · cada una dice `no medido todavía` · **`queryByText("sin hallazgos")` dentro de ese bloque es `null`** (RF-263 lo prohíbe explícitamente) | RF-263 |
| `DetectorSinFixPropuesto` | hallazgo cotizado sin fix | el detector aparece en la lista con `sin fix propuesto` · **no está oculto** — es la contraparte de `DescartaSinContrafactual` (§2.10) | RF-246 |
| `VentanaHeredada` | ventana de la capa = 30 días | el encabezado dice `Tokens · 30 días (14 corridas)` — la ventana es **la de la capa**, y las 14 corridas son el **denominador**, no la ventana (J-3) · assert de que el string `últimas 14 corridas` **no** está en el DOM | RF-264 · J-3 |
| `NodoNoCaja` | nodo `mcp` | una sola sección: `getByText(/^Esta capa mide cajas\./)` seguida del **mismo motivo que el canvas** — assert de igualdad literal contra la constante de copy usada en `SinDatoMcp` (§2.5) · **cero tablas**: `querySelectorAll("table").length === 0` · `queryByText("0")` es `null` | RF-258 |
| `CargandoYError` | dos estados en un archivo, dos stories | `Cargando`: cada sección con su `Skeleton` y **los títulos ya visibles** (`getByText("Detectores")` presente) · `Error`: motivo + `Reintentar` dentro de la tab | design §5.5 |

*(La última fila son 2 stories: `Cargando` y `Error` — total 17.)*

### 2.12 `widgets/map-canvas/ui/inspector.stories.tsx` — **+3 stories**

> 🔴 **La story vigente `Tabs` asserta `await expect(tabs).toHaveLength(3)` (línea 92). Hay que
> cambiarla a 4 en el mismo commit que agrega la tab, o CI se pone roja.** Es la única modificación
> permitida a una story firmada en este paquete y va declarada en `PARIDAD.md`.

| story | escenario | asserts | RF |
|---|---|---|---|
| `CuatroTabs` | inspector de una caja | `getAllByRole("tab")` longitud **4** con nombres `["Resumen","Contenido","Corridas","Mejora"]` en ese orden · `Resumen` tiene `aria-selected="true"` por default (sin cambio) · las tres vigentes conmutan igual que hoy (se re-corren los asserts de la story `Tabs`) | RF-258 |
| `TabMejoraContratoAria` | click en `Mejora` | la tab tiene `id` propio y `aria-controls` apuntando a un panel existente · ese panel tiene `role="tabpanel"` y `aria-labelledby` de vuelta a la tab · **los 3 paneles inactivos tienen el atributo `hidden`** (`querySelectorAll("[role=tabpanel][hidden]").length === 3`), no `display:none` por CSS | RF-277 |
| `CambioDeNodoVuelveAResumen` | tab `Mejora` abierta, se selecciona otro nodo | tras el rerender con otro `box`, `getByRole("tab", { name: "Resumen" })` vuelve a `aria-selected="true"` — igual que hoy | RF-277 |

### 2.13 `widgets/map-canvas/ui/map-canvas.stories.tsx` — **+4 stories**

> Este archivo es el *fitness fixture of record*. Las 8 stories vigentes **no se tocan**.

| story | escenario | asserts | RF |
|---|---|---|---|
| `CapaMejoraSupersetGeografia` | `dev-full-cycle` con `capa="mejora"` | los 6 asserts de la story `Dogfood` **se repiten idénticos** y pasan · además: `querySelectorAll(".node").length` es el mismo que en `Dogfood` · `querySelectorAll(".lane").length` igual · `querySelectorAll("path")` del `EdgeLayer` igual. Es T-16: misma geografía, mismos nodos, mismos edges | RF-245 · T-16 |
| `CapaMejoraSoloCajasLlevanCifra` | dogfood con cifras en las 4 cajas | `querySelectorAll(".mej-cifra").length === 4` · para cada nodo que **no** es caja (`std-spec`, `hook-posttooluse`, `hook-stop`): `node.querySelector(".mej-cifra")` es `null` **y** `node.textContent` no contiene `"0"` como cifra suelta · el total de los carriles suma el total del arnés | RF-238 · RF-244 · T-17 |
| `ConmutarNoPierdeSeleccion` | `estructura` → `mejora` → `estructura`, con un nodo seleccionado | el nodo sigue con `aria-pressed="true"` tras las dos conmutaciones · el DOM final es equivalente al inicial: `querySelectorAll(".mej-cifra").length === 0` al volver | RF-245 |
| `CopyConfianzaEsUnaSola` | nodo `por-hash` en el canvas + `<MarcaConfianza confianza="por-hash">` renderizado al lado en la misma story | el `textContent` del chip del nodo **es idéntico** al del componente de `entities/telemetria` · lo mismo para el `title`. **Es el candado de BLOQ-1 opción A**: si alguien edita un literal y no el otro, CI se rompe. Vive acá porque un widget sí puede importar las dos entities | BLOQ-1 · RF-242 · RF-281 |

### 2.14 `widgets/map-canvas/ui/politica-datos-dialog.stories.tsx` — 8 stories (NUEVO)

| story | escenario | asserts | RF |
|---|---|---|---|
| `Reposo` | diálogo abierto | `getByRole("dialog")` con `aria-modal="true"` y `aria-labelledby` al título `Qué guardamos de la telemetría` · los 3 bloques presentes · copy literal de «no se guarda»: `No guardamos nada de tu cuenta: ni email, ni identificadores de usuario, ni de organización.` y `Y no guardamos nada del contenido: ni tu prompt, ni la respuesta, ni lo que leyó o escribió una herramienta. Eso llega por el canal de hooks y se descarta antes de escribirse en disco.` | RF-275 · H-14 |
| `CamposDesplegados` | click en `ver los campos exactos` | `aria-expanded` pasa a `true` · la lista muestra los campos **por nombre**: al menos `session_id`, `prompt_id`, `hook_event_name`, `tool_name`, `duration_ms`, `cwd` · **la lista viene por prop, no está hardcodeada en el componente**: se pasa `camposPersistidos` en `args` y se asserta que el DOM refleja exactamente ese array (así la UI no puede divergir de la allowlist real — RF-275 «esa lista es la misma allowlist que aplica la ingesta») | RF-275 · RF-282 |
| `RetencionDesdeConfig` | `retencionDias=45` | `getByText("Se borra solo a los 45 días.")` · **assert de que 90 no está hardcodeado**: `queryByText(/90 días/)` es `null`. J-6 dice que el 90 es PROPUESTO, no firmado | RF-283 · J-6 |
| `Confirmacion` | click en `Borrar la telemetría de este arnés` | título `Borrar la telemetría de «vitalia»` · cuerpo `Se borran 1 284 corridas medidas y los puntos de mejora que salieron de ellas. No se puede deshacer.` · **el botón `Borrar` nace `disabled`** hasta marcar la confirmación; tras marcarla, `toBeEnabled()` | RF-275 |
| `Borrando` | POST en vuelo | botón en espera (`Borrando…`, `aria-busy="true"`) · **el diálogo no se puede cerrar**: `userEvent.keyboard("{Escape}")` → `args.onCerrar` **no** fue llamado; el botón `Cancelar` está `disabled` | design §5.7 |
| `ErrorDeBorrado` | falla | `getByText("No se pudo borrar — permiso denegado. No se borró nada.")` · el diálogo **sigue abierto** (`getByRole("dialog")`) · el botón vuelve a estar operable | design §5.7 |
| `Exito` | borrado ok | `args.onCerrar` llamado · región `aria-live` con `Listo. Este arnés vuelve a estar sin datos de telemetría.` | RF-275 |
| `FocoAtrapado` | teclado | al abrir, el foco está dentro del diálogo · `Tab` desde el último control vuelve al primero (`trapTabKeyDown`) · `Escape` en reposo llama `args.onCerrar` · todo control tiene anillo visible: para cada uno, `focus()` y `getComputedStyle(el).outlineWidth !== "0px"` | RF-278 |

### 2.15 `widgets/portafolio/ui/tabla-mejora-portafolio.stories.tsx` — 9 stories (NUEVO)

| story | escenario | asserts | RF |
|---|---|---|---|
| `ConDato` | vitalia · 0,31/corrida | encabezados `USD/corrida`, `tendencia`, `punto de mejora` · la celda dice `0,31` **sin** prefijo (el `USD` está en el encabezado) y con `tabular-nums` · el encabezado declara explícitamente que es USD por corrida | RF-265 |
| `UnArnesDosPuestos` | mismo arnés, 2 instalaciones | **2 filas**, no una: `querySelectorAll(".pf-mej-fila").length === 2` · cada una con su propio valor de costo (assert de que los dos textos difieren) | RF-265 |
| `SinPuestoDeclarado` | instalación sin puesto (BLOQ-3, el caso real de hoy) | `getByText("puesto sin declarar")` · **no se fabrica un puesto**: `queryByText(/Product Owner/)` es `null` | RF-265 · BLOQ-3 |
| `SinFugas` | arnés con datos, 0 hallazgos | `getByText("✓ sin fugas detectadas")` · es **distinguible de sin-dato**: en la misma story hay una fila sin dato y los dos `textContent` difieren; además el chip de «sin fugas» tiene una clase distinta de la de «sin dato» | RF-267 · RF-280 |
| `SinDato` | arnés nunca instrumentado | **la fila SIGUE en la tabla** · la celda de costo dice `sin dato` · `queryByText("0,00")` en esa fila es `null` · la celda de tendencia está vacía y **marcada como tal** (`aria-label` o texto `—` con rótulo) · la celda de mejora dice `nunca corrió con telemetría` | RF-268 |
| `PocasCorridas` | 1 corrida | `getByText("pocas corridas para una tendencia")` · no hay sparkline | RF-266 |
| `OrdenadaSinDatoAlFinal` | 3 con dato + 2 sin dato, orden por costo | las 3 primeras filas tienen valor numérico y están ordenadas descendentes (assert aritmético sobre los textos parseados) · **hay un separador rotulado** `Sin datos de telemetría` antes de las 2 últimas · **los sin-dato no se intercalan**: el índice de la primera fila sin dato es mayor que el de la última con dato | RF-268 |
| `PieConDisclaimerYConfianza` | reposo | pie de tabla con `Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no aparecen en cero.` · una fila `por-hash` lleva la **misma** marca de duda del canvas: `getByText("por huella")` | H-12 |
| `NombresLargosYNumerosGrandes` | id de 60 chars + `12 345,67` | la columna `punto de mejora` (`1fr`) se encoge primero · **la página no scrollea horizontal**: `document.body.scrollWidth <= document.body.clientWidth` · el scroll vive en `.pf-mej-scroll`, que tiene `tabindex="0"`, `role="region"` y `aria-label` (requisito de axe para contenedor scrolleable, design §6.3) | design §6.3 |

### 2.16 `widgets/portafolio/ui/portafolio-list.stories.tsx` — **+2 stories**

> Las 12 stories firmadas del Slice 1 **no cambian**: las props nuevas son opcionales.

| story | escenario | asserts | RF |
|---|---|---|---|
| `ConMejora` | `entradasDemo` + datos de telemetría | las 3 celdas nuevas aparecen **entre chips y dot de salud**: para una fila, el índice DOM del `.pf-mej-usd` es mayor que el de `.pf-fila-chips` y menor que el del dot · el dot de salud **sigue cerrando la fila** (ancla firmada en la PARIDAD del Slice 1) · el comportamiento de la fila no cambia: click sigue llamando `onAbrir` con la clave exacta | RF-265 · BR-M16 |
| `SinMejoraDOMIntacto` | `entradasDemo` sin props nuevas | `querySelector(".pf-mej-usd")` es `null` · el conteo de celdas de la fila es el del Slice 1 (5) — el guardián del superset | BR-M16 |

### 2.17 `shared/ui/estado-carga.stories.tsx` — 2 stories (NUEVO, por la promoción)

> `fe-visual-fitness` check `story-es-test`: *«cada componente de `shared/ui` + widget tiene story»*.
> La promoción crea componentes en `shared/ui` que hoy no tienen story propia.

| story | escenario | asserts | RF |
|---|---|---|---|
| `SkeletonHonesto` | `<Skeleton label="Cargando portafolio" />` | `getByRole("status", { name: "Cargando portafolio" })` · **cero filas fantasma**: las barras del skeleton tienen `aria-hidden="true"` · `aria-live="polite"` | H-6 · G5 |
| `ErrorConMotivoYReintentar` | motivo real | `getByRole("alert")` con el motivo textual · botón `Reintentar` habilitado que llama el handler · el motivo **no** es genérico: assert de que el texto contiene el string inyectado | H-6 |

### 2.18 a11y — el gate está en `error` y estos tres puntos lo rompen

`.storybook/preview.ts` tiene `a11y: { test: "error" }`. Cada story nueva tiene que pasar axe.
Tres riesgos **medidos**, no intuidos (tokens de `web/src/app/styles/theme.css`, tema claro):

| # | dónde | medición | veredicto |
|---|---|---|---|
| **A11Y-1** 🔴 | **disclaimer «estimado, no facturación»** — `design.md` §4.2 lo pinta `texto --warn` sobre `fondo --warn-soft` | `--warn` `#c96a2e` sobre el compuesto de `--warn-soft` `rgba(201,106,46,.13)` sobre `--card` `#ffffff` = **3,26 : 1**. Mínimo axe para texto: 4,5 : 1 | **falla el gate.** Y contradice al propio `design.md` §4 fin: *«todo texto nuevo sobre `--warn-soft` usa `--foreground`, no `--warn`»*. **La tabla §4.2 gana mal: hay que corregirla a `--foreground`** antes de codear el disclaimer. Es el mismo bug que ya está en `BACKLOG.md` (`.text-warn` a 3,76 : 1 sobre blanco, rompe 4 stories de `new-session-picker`) — este paquete **no lo puede agravar** |
| **A11Y-2** 🔴 | **marca de fuga grave** — `design.md` §4.2: `texto --crit` sobre `--crit-soft` | `--crit` `#c94545` sobre el compuesto de `--crit-soft` `rgba(201,69,69,.12)` sobre blanco = **4,03 : 1** | **falla el gate** por poco. Sobre `--card` puro daría 4,75 : 1 y pasaría. Fix: el texto de la marca en `--foreground`; `--crit` queda para borde y punto (requisito 3 : 1, que sí cumple) |
| **A11Y-3** ⚠️ | **barra de cobertura, segmentos contiguos** — `--heat-4` vs `--heat-3` | compuestos sobre blanco: **1,20 : 1** entre segmentos adyacentes | **axe NO lo caza** (su regla `color-contrast` solo mira texto), así que el gate pasaría con una barra ilegible. Por eso `design.md` §2.2 mete el separador de 1 px de `--card` — y por eso **el assert de verdad es el texto**: `CuatroSegmentos` (§2.3) asserta el `aria-label` completo y el texto de al lado, no los colores. La barra nunca es el único portador (RF-237: *«distinguibles sin depender del color»*) |

Riesgos menores a vigilar, sin veredicto todavía (medir al construir):
`--muted-foreground` `#4d5555` sobre blanco da **7,65 : 1** — pasa, incluso a 9 px (los `dt` de la
tarjeta y el `score v1`). El **tema oscuro** de los tres casos de arriba hay que medirlo aparte:
`--warn` `#e0904f` sobre `--card` `#0c1110` es texto claro sobre fondo oscuro y probablemente pasa,
pero **el toolbar de tema de `.storybook/preview.ts` renderiza las stories en `light` por default**,
así que el gate solo ve el claro. **Acción**: agregar una story de tema oscuro por archivo crítico
(franja y tarjeta) con `globals: { theme: "dark" }`, o aceptar por escrito que el oscuro queda sin
gate automático. **Se propone lo primero** y se cuenta dentro del presupuesto (no suma stories
nuevas de contenido: se aplica a `Reposo` y `CompletaB1Atencion` duplicándolas —
`ReposoDark` y `CompletaB1AtencionDark`, **+2 stories**, contadas en el total de 124).

Regla dura para el constructor: **ninguna story nueva nace con `a11y: { test: "todo" }`**. Si una
falla, se arregla el color o el marcado. El único `todo` heredado es el de
`arnes-node.stories.tsx`/`map-canvas.stories.tsx`, que **no se amplía** — por eso las marcas nuevas
del nodo van en `--foreground`.

---

## 3 · Fixtures

**Regla dura del repo: datos reales de dogfood, jamás inventados.** Cuando un número no lo midió
nadie, se dice.

### 3.1 Dónde viven

| archivo | qué contiene |
|---|---|
| `web/src/entities/telemetria/testing/telemetria.ts` | **nuevo.** Fixtures de telemetría, en dos bloques separados y rotulados: `// ── MEDIDO ──` y `// ── ILUSTRATIVO ──` |
| `web/src/entities/arnes/testing/dev-full-cycle.ts` | **existe, no se toca.** El grafo del dogfood: 4 cajas (`spec-writer` · `builder` · `reviewer` · `releaser`), 1 regla (`std-spec`), 2 hooks. Es el grafo de todas las stories de canvas |
| `web/src/entities/arnes/testing/luana-feature-cycle.ts` | **existe, no se toca.** De acá salen los nodos que `dev-full-cycle` no tiene: `api-mcp` (mcp «acceso a la API»), `test-author` (subagent «escribir pruebas»), `spec-review` (command) — los tres casos de RF-243 |
| `web/src/entities/portafolio/testing/entradas.ts` | **existe, no se toca.** `entradasDemo` (3 entradas reales del E2E de Slice 0). `acme-cli` tiene 2 instalaciones en 2 proyectos ⇒ es el caso real de `UnArnesDosPuestos` |

### 3.2 Bloque `MEDIDO` — sale de `verificacion-2026-07-26/`, byte por byte

| dato | valor | fuente |
|---|---|---|
| forma de `api_request` | `model=claude-haiku-4-5-20251001` · `input_tokens=10` · `output_tokens=39` · `cache_read_tokens=17536` · `cache_creation_tokens=8257` · `cost_usd=0.0184726` · `cost_usd_micros=18473` · `duration_ms=1772` · `speed=normal` · `query_source=sdk` | `INFORME.md` V1 · `evidencia/logs-run1.json` |
| split 5m/1h | `input 10` · `cache_read 21695` · `cache_creation 4099` · `output 35` · **`ephemeral_1h_input_tokens: 4099`** · **`ephemeral_5m_input_tokens: 0`** | `INFORME.md` V2 · `evidencia/result-envelope.json` |
| **el cero legítimo de RF-260** | `ephemeral_5m = 0` — es un **0 real medido**, no inventado. Alimenta `CeroLegitimo` (§2.11) | ídem |
| **la ausencia de RF-260** | `razonamiento` **nunca apareció** en ninguno de los 57 log records ⇒ el «no aplica» también es real, por ausencia. Alimenta `NoAplicaNoEsCero` | `INFORME.md` V1 (inventario de 57 records) |
| `plugin_id_hash` | `083a6411f71e6967` y `29e6a51f5b6dca52` (third-party, user-local) · `5de2da3204573e62` (`rust-analyzer-lsp`, official) | `INFORME.md` V3 |
| huella desconocida | cualquier hash **fuera** de esos tres ⇒ alimenta `sin-dato` sin inventar nada | `INFORME.md` V3 |
| PII que llega y se descarta | `user.email` · `user.account_uuid` · `user.account_id` · `organization.id` · `user.id` | `INFORME.md` V6 |
| contenido que llega por hook | `UserPromptSubmit.prompt` · `Stop.last_assistant_message` · `PostToolUse.tool_response` | `ANEXO-hooks.md` H4 |
| campos de la allowlist del hook | `session_id` · `prompt_id` · `hook_event_name` · `tool_name` · `duration_ms` · `cwd` normalizado | `ANEXO-hooks.md` H4 — **es la lista que `CamposDesplegados` (§2.14) asserta** |
| tamaños de herramienta | `tool_input_size_bytes=143` · `tool_result_size_bytes=9` · `duration_ms=1` | `ANEXO-hooks.md` H8 |
| volumen del experimento | 3 corridas · 57 log records · 8 métricas · 14 puntos · **USD 0,044 en total** | `INFORME.md` cabecera |
| break-even de B1 | `(2−1,25)/(2−0,1) = 39,47 %` — **no es una medición, es álgebra** sobre los multiplicadores del catálogo (write 1h ×2, write 5m ×1,25, read ×0,1). Verificado a mano: 0,75/1,9 = 0,3947 | D16.1 · RF-250 |
| temporalidad | las 4 métricas llegan `Sum · Delta · monotonic` | `INFORME.md` V5.1 |

### 3.3 Bloque `ILUSTRATIVO` — marcarlo, y decir por qué existe igual

**Nadie midió jamás una ventana de 61 corridas instrumentadas.** La única telemetría real que existe
en este repo son las **3 corridas** del 2026-07-26, por un total de **USD 0,044**. Todos estos
números son **constantes de diseño**, no mediciones:

`USD 4,82` · `61 corridas · 12 sesiones · 4 cajas` · `USD 1,92` / `40 %` por caja · `14 corridas` ·
`USD 0,84` / `44 %` · `relectura 61 %` · `USD 0,53/corrida` · `USD 0,31` · `USD 0,67` / `USD 0,89` /
`75 %` · `USD 0,22` · `USD 0,47` · `USD 0,74` · `USD 3,10 — de 5 corridas, 3` · `USD 0,18` de
divergencia · `1 284 corridas` del borrado · `12 exactas · 3 por huella · 2 por proceso · 1 sin dato`.

**Por qué se usan igual, y por qué son load-bearing:** `design.md` §7 fija el **copy literal**, y ese
copy **lleva los números adentro** (`Gasta USD 0,84 de los 1,92 de la caja (44 %)…`). Las stories
assertan el literal (regla 5 de `design.md` §0). Si el fixture no produce exactamente esos números,
el assert de copy falla. O sea: **son constantes del diseño, y el test las trata como tales.**

Obligaciones del constructor:

1. El bloque va precedido por este comentario, textual:
   ```
   // ── ILUSTRATIVO ── NADIE MIDIÓ ESTOS NÚMEROS.
   // La única telemetría real del repo son las 3 corridas de verificacion-2026-07-26/
   // (USD 0,044 en total). Estos valores son CONSTANTES DE DISEÑO: design.md §7 fija el copy
   // literal y el copy lleva el número adentro, así que el fixture tiene que reproducirlos para
   // que los asserts de copy pasen. NO son evidencia de nada. Al primer dato de dogfood real,
   // se reemplazan y se corrigen los literales de design.md §7 en el mismo commit.
   ```
2. La **forma** (campos, tipos, punteros nil vs 0) sale del bloque `MEDIDO`. Nunca se inventa un campo.
3. **`H-8` sigue sin cerrar:** el `aria-label` del mockup (`12+3+2 = 17`) no cierra con el
   denominador (`61 corridas`). Este plan usa **18** como total de la barra (12+3+2+1, con el cuarto
   segmento de H-13) y **declara la unidad** en el texto, como manda RF-237 y propone H-8. Los `61`
   de la franja y los `18` de la cobertura son de **fixtures distintas** a propósito, para no fingir
   una coherencia que el mockup todavía no tiene. **Hay que cerrarlo en la iteración 2 del mockup.**

### 3.4 Cero mocks donde hay dato real

`spec.md` §2: *«Prohibido mock donde hay dato real disponible»*. Concretamente:

- El grafo: siempre `devFullCycle` / `luanaFeatureCycle`, jamás un grafo a mano.
- El Portafolio: siempre `entradasDemo`.
- El motivo de error de transporte: el formato real del daemon
  (`GET /api/portafolio: conexión rechazada (ECONNREFUSED)`), como ya hace
  `portafolio-list.stories.tsx`.
- El motivo de decode Go: el formato real de `derr.Error()`, como ya hace `corruptaEjemplo`.

---

## 4 · Qué NO se verifica con Storybook

### 4.1 Los 5 RF sin superficie (§H del spec) — ninguno es asertable con una story

| RF | por qué no | mecanismo real |
|---|---|---|
| **RF-282** — allowlist en los dos caminos | probar una ausencia en el almacén no se hace desde el DOM | **El método ya está en el spec (§3, fila RF-282) y se reusa tal cual, no se inventa otro:** *test de texto sobre el almacén real* — se corre el hook con un prompt que contiene un **marcador único**, y se busca ese marcador **como subcadena del archivo SQLite**. Si aparece, falla. Va como `TestAllowlistNoPersistePII` y `TestHookNoReenviaContenido` (CAP-121), con el control positivo que exige `escenarios.md` §14: en la misma corrida, un campo de la allowlist que **sí** tiene que estar. Complemento: `TestAllowlistEsListaNoSugerencia` (campo nuevo desconocido → no se persiste **y** el descarte queda contable) |
| **RF-283** — retención y borrado | el test asserta que la consulta no devuelve nada, no que el byte se fue | test de integración Go sobre el `.db` real: `TestPurgaRespetaTTL` · `TestBorradoPorArnesTambienLimpiaRollup` (no queda agregado huérfano). La UI del botón sí tiene story (§2.14); **lo que borra, no** |
| **RF-284** — ni egresa ni guarda de más | no se prueba una ausencia por muestreo | **dos** checks duros de conformance del arnés: `telemetria-no-egresa` (destino) **y su hermano** `hook-proyecta-campos` (qué campos escribe). El primero solo no alcanza — ANEXO H4. Más `TestForwardApagadoPorDefault` y `TestForwardNoReenviaCrudo` |
| **RF-285** — confianza + los dos costos persistidos | es esquema, no píxel | test de tabla Go: `TestDobleCostoSePersisteEntero` · `TestConfianzaDeAgregadoEsLaMinima` · `TestJoinPorSesionYTurno` (igualdad del par `(session_id, prompt_id)`, sin heurística de tiempo) |
| **RF-286** — «no aplica» se persiste como ausencia | ídem | test de tabla Go: `TestNoAplicaNoEsCeroEnElWire` + `TestNoAplicaSobreviveAlRollup`. **La story `NoAplicaNoEsCero` (§2.11) verifica el otro extremo del cable** — que la UI distingue `nil` de `0` — pero no puede verificar que el `nil` llegó así |

Además, **nada de esto se toca desde el FE y hay que decirlo**: el receptor OTLP, el decodificador
`intValue`/`json.Number`, la temporalidad Delta, el esquema SQLite, el join, los 6 detectores y el
catálogo de precios son **Go**, con la matriz completa de `escenarios.md` §A-I (≈100 casos, ids
`U`/`F`/`E2E`/`M`). Este plan **no los duplica**.

### 4.2 Los 6 RF 🎨 que **no** se pueden cerrar con una story

| RF | por qué la story no alcanza | con qué se verifica |
|---|---|---|
| **RF-236** — «estimado» *se lee* | la story asserta que el texto existe y es visible; no que un humano lo lea antes de creer el número | la story `DisclaimerEnSuperficie` (§2.8) cubre lo asertable (existe · visible · no en `title` · no plegado). **Lo que falta es click-through humano del gate PARIDAD a 1440×900**, con captura, exactamente como dice `spec.md` §3 |
| **RF-247 / RF-248** — la calidad del copy | «la frase es buena» no es asertable | la story asserta el **literal** de `design.md` §7 (y `TitularSinJerga` asserta las ausencias). El juicio sobre la frase es del operador en el gate del mockup |
| **RF-252** — el sesgo va **en contra** | «en contra» es un juicio sobre el cálculo, no sobre el DOM | las 3 stories de sesgo (§2.9) assertan que la fila existe siempre y que la dirección está escrita. **Que la dirección sea la correcta** se verifica en Go: revisión por par de cada detector + assert de que `direccion_sesgo` existe y no es neutro (`TestB1BreakEvenTTL` y hermanos, CAP-130) |
| **RF-254** — el score **bumpeado** | testeable el literal, no que alguien haya subido la versión al cambiar la fórmula | la story asserta que `score v1` es visible sin hover. **El bump lo enforza un check de conformance**: si cambia la fórmula de un detector y no cambia `score_version`, falla |
| **RF-274** — determinismo de `plugin_id_hash` entre máquinas | **no verificado** (V7.2): una sola máquina | la story `MejoraPorHuella`/`PorHash` verifica que la UI **dice** «por huella», que es honesto aunque el hash no sea portable. Cerrarlo requiere correr el mismo plugin en otra máquina/usuario — **queda abierto y declarado** |
| **RF-281** — un monto por debajo del centavo | la story asserta que `0,004` no se redondea a `0,00`; **la otra mitad de la regla** («o se agrega en un total que sí sea legible») es una decisión de agregación del backend | `MenorAlCentavo` (§2.1) cubre la mitad del FE. La agregación se verifica en Go (`TestCosteoNoAplanaLosTiers` y el redondeo del rollup) |

Y **T-21 / T-22** de `spec.md` §2:

- **T-21 (a11y con axe)**: **sí corre en Storybook**, headless, con el gate en `error`. La fila de
  `spec.md` §3 que dice *«vitest-browser no corre en background»* es la nota derogada: **corregirla**.
- **T-22 (sin color)**: axe no verifica escala de grises. Se cubre **estructuralmente** en las
  stories (`SeveridadSinColor`, `SeisDetectoresConEstado`, `SinFugas`, `LosCuatroJuntos`: cada estado
  tiene texto propio y el color es refuerzo) **más** una captura en escala de grises en el
  click-through del gate PARIDAD.

### 4.3 Fuera de todo test — click-through humano del gate

Estos van al `PARIDAD.md`, fila por fila, con captura a 1440×900:
RF-236 legible sin hover · la calidad del copy de RF-247/248 · que la geografía «se siente» igual al
conmutar (RF-245, más allá del conteo de nodos) · la captura en escala de grises (T-22) · y el
E2E contra el daemon vivo en `127.0.0.1:4200` para RF-255 (que «Proponerlo» abra el chat de verdad,
con su alcance y su permiso).

---

## 5 · Comandos exactos de verificación

Todo desde `web/` salvo donde se indique. **Orden recomendado: barato → caro.**

```bash
# ── 0 · antes de escribir la primera story: que el árbol esté verde ────────────
cd /home/chalreme/Proyectos/harness-studio/web
npx vitest run --project=storybook          # línea base: todas las stories firmadas pasan
```

```bash
# ── 1 · loop de iteración, por archivo (≈4 s por archivo de 1 story) ──────────
npx vitest run --project=storybook src/entities/telemetria/ui/cifra-usd.stories.tsx
npx vitest run --project=storybook src/widgets/map-canvas/ui/franja-mejora.stories.tsx
# …un archivo por vez, en el orden de §2 (entities → widgets → pages)
```
Pasa ⇒ render + `play` + **axe** verdes para ese archivo. Es el gate real, no un smoke.
Con `--project=storybook` corre **solo** el proyecto de browser; `--project=unit` corre los
`*.test.ts` de Node (los `selectors.test.ts` colocados).

```bash
# ── 2 · los tests unitarios colocados de entities/telemetria ──────────────────
npx vitest run --project=unit src/entities/telemetria/model/selectors.test.ts
```

```bash
# ── 3 · toda la suite visual (el gate de CI, literal) ─────────────────────────
npx vitest --project=storybook run
```
Es exactamente lo que corre el job `visual-fitness` de `.github/workflows/ci.yml`
(precedido allí de `pnpm exec playwright install --with-deps chromium`).
Pasa ⇒ `story-es-test` y `a11y-axe` de `fe-visual-fitness.md` están verdes.

```bash
# ── 4 · los gates estáticos (typecheck + lint + arquitectura + tokens) ────────
pnpm --dir web run verify
```
= `typecheck && lint && depcruise && fsd && stylelint`. **Los tres que importan acá:**
- `depcruise` → `canvas-not-chrome`, `chrome-not-canvas-internals`, `ui-not-domain`,
  `no-sibling-widget-imports` (todos `error`).
- `fsd` (**steiger**) → `fsd/no-cross-imports`. **Es el que caza BLOQ-1** si alguien mete
  `entities/arnes → entities/telemetria`.
- `stylelint` con `strict-value` → cero color literal en `mejora.css` y en las extensiones.

```bash
# ── 5 · tokens y contrato ─────────────────────────────────────────────────────
pnpm --dir web run tokens:build && git -C /home/chalreme/Proyectos/harness-studio diff --exit-code web/src/app/styles/theme.css
```
Pasa ⇒ no se agregó ningún token a mano. `design.md` §4 dice «ningún color nuevo»: este comando lo
prueba.

```bash
# ── 6 · el lado Go (los RF sin superficie de §4.1) ────────────────────────────
cd /home/chalreme/Proyectos/harness-studio
go test ./...
go test ./docs/architecture/fitness/...
```

```bash
# ── 7 · conformance y capabilities (R1·R2·R3·R4) ──────────────────────────────
cd /home/chalreme/Proyectos/harness-studio
make            # dispara el pipeline de conformance del repo
./arnesia conformance --todo   # sin regresión contra la línea base del checkpoint
```
Recordatorio de `capabilities-a-crear.md` §0: **las hojas YAML se crean en el mismo commit que crea
su símbolo** (R3), nunca antes — si no, R1 rompe con punteros colgantes.

```bash
# ── 8 · E2E contra el daemon vivo (RF-255, §4.3) ──────────────────────────────
# el daemon en 127.0.0.1:4200, la app abierta, y el click-through del gate PARIDAD
```

**Regla de corte:** un archivo de stories no se da por hecho hasta que pasa **(1)** su corrida
individual **y** **(4)** `verify`. Un archivo que pasa (1) pero rompe `fsd` está mal ubicado en la
taxonomía, no mal testeado.

---

## 6 · Resumen por componente

| archivo | stories nuevas | nuevo / superset |
|---|---:|---|
| `entities/telemetria/ui/cifra-usd.stories.tsx` | 4 | nuevo |
| `entities/telemetria/ui/marca-confianza.stories.tsx` | 5 | nuevo |
| `entities/telemetria/ui/barra-cobertura.stories.tsx` | 5 | nuevo |
| `entities/telemetria/ui/sparkline.stories.tsx` | 4 | nuevo |
| `entities/arnes/ui/arnes-node.stories.tsx` | 13 | superset (14 firmadas intactas) |
| `widgets/map-canvas/ui/map-bar.stories.tsx` | 4 | superset (1 firmada intacta) |
| `widgets/map-canvas/ui/lane.stories.tsx` | 2 | superset |
| `widgets/map-canvas/ui/franja-mejora.stories.tsx` | 20 (19 + `ReposoDark`) | nuevo |
| `widgets/map-canvas/ui/punto-mejora-card.stories.tsx` | 15 (14 + `CompletaB1AtencionDark`) | nuevo |
| `widgets/map-canvas/ui/puntos-mejora-list.stories.tsx` | 7 | nuevo |
| `widgets/map-canvas/ui/inspector-mejora.stories.tsx` | 17 | nuevo |
| `widgets/map-canvas/ui/inspector.stories.tsx` | 3 | superset (**+1 assert firmado a corregir: `toHaveLength(3)` → `4`**) |
| `widgets/map-canvas/ui/map-canvas.stories.tsx` | 4 | superset (8 firmadas intactas) |
| `widgets/map-canvas/ui/politica-datos-dialog.stories.tsx` | 8 | nuevo |
| `widgets/portafolio/ui/tabla-mejora-portafolio.stories.tsx` | 9 | nuevo |
| `widgets/portafolio/ui/portafolio-list.stories.tsx` | 2 | superset (12 firmadas intactas) |
| `shared/ui/estado-carga.stories.tsx` | 2 | nuevo (promoción) |
| **TOTAL** | **124** | 12 archivos nuevos · 5 supersets |

### Mapa de los 7 estados honestos → story dueña (ninguno queda huérfano)

| estado | story dueña | co-verificado en |
|---|---|---|
| 1 · sin datos | `Estado1SinDatos` (§2.8) | `SinCorridas` (§2.3) · `SinDato` del Portafolio (§2.15) |
| 1b · sin corridas en la ventana (H-9) | `Estado1bSinCorridasEnVentana` (§2.8) | `CajaSinCorridasEnVentana` (§2.11) · `MejoraCajaSinCorridas` (§2.5) |
| 2 · cobertura parcial | `Estado2CoberturaParcial` (§2.8) | `CuatroSegmentos` (§2.3) |
| 3 · S2 **sin instrumentar** | `Estado3S2SinInstrumentar` (§2.8) | `DetectorB1ApagadoEnS2` (§2.11) · `S1Only` (§2.9) |
| 3b · S2 **instrumentado** (J-10) | `Estado3bS2Instrumentado` (§2.8) | — *(los dos modos se assertan como rótulos distintos, en la misma familia)* |
| 4 · otro runtime | `Estado4OtroRuntime` (§2.8) | `SinCostoDelRuntime` (§2.11) |
| 5 · catálogo viejo | `Estado5CatalogoViejo` (§2.8) | `CatalogoSinConstruir` (§2.11) |
| 6 · por huella | `PorHash` (§2.2) | `MejoraPorHuella` (§2.5) · `PieConDisclaimerYConfianza` (§2.15) |
| 7 · qué guardamos | `Reposo` de `politica-datos-dialog` (§2.14) | `ResumenQueGuardamos` (§2.8) · `CamposDesplegados` (§2.14) |

### Los 4 niveles de `atribucion_confianza` → una story cada uno, más el careo

`Exacta` · `PorHash` · `PorProceso` · `SinDato` (§2.2), más `LosCuatroJuntos` que asserta que
**ninguno** se pinta como «aproximada» genérica, y `MejoraCifraExacta` / `MejoraPorHuella` /
`MejoraPorProceso` / `MejoraCajaSinCorridas` (§2.5) que lo repiten **dentro del nodo**, que es donde
el operador los va a ver.

---

## 7 · Lo que este plan destapó y hay que resolver fuera de él

| # | qué | dónde va |
|---|---|---|
| 1 | **BLOQ-1** — cross-import `entities/arnes → entities/telemetria` (§1.3). Opción A propuesta, falta firma | `decisiones.md` de este paquete |
| 2 | **BLOQ-2** — `Clase` no tiene `knowledge`; `design.md` §7.3 tiene copy para un nodo que no existe | `design.md` §7.3 (borrar la fila) |
| 3 | **BLOQ-3** — `EntradaPortafolio` no tiene `puesto`; RF-265 lo pide | `BACKLOG.md` + `spec.md` §4 |
| 4 | **A11Y-1 / A11Y-2** — `design.md` §4.2 elige combinaciones que fallan axe (3,26:1 y 4,03:1) y contradicen su propio §4 | `design.md` §4.2 (corregir a `--foreground`) |
| 5 | **Nota derogada** — `design.md` §8 y `spec.md` §3 (T-21) dicen que vitest-browser no corre headless. **Es falso**, verificado hoy | los dos archivos, mismo commit |
| 6 | **H-8 abierto** — el `aria-label` de la cobertura (17) no cierra con el denominador (61). Este plan usa 18 y declara la unidad; el mockup tiene que cerrarlo | iteración 2 del mockup |
| 7 | **`inspector.stories.tsx:92`** — `toHaveLength(3)` pasa a `4`. Único cambio a una story firmada | `PARIDAD.md`, declarado |
| 8 | **Tema oscuro sin gate** — `preview.ts` renderiza en `light` por default; el oscuro no lo ve axe. Se agregan 2 stories `globals: { theme: "dark" }` | ya incluido en el total |
