# Diseño de la superficie · Las conversaciones viven en el panel de conversación

> `tipo: design` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Par de [`spec.md`](./spec.md) (RF-300…RF-357) y [`plan-storybook.md`](./plan-storybook.md).
> Habilitado por el **🧑‍⚖️ GATE 1** del mockup (it. 2, 2026-07-26).
>
> **El UI al pixel.** Qué componente nace, cuál se modifica, cuál NO se toca; su capa FSD con el
> motivo; sus props tipadas; sus estados visuales exhaustivos; sus tokens **medidos**.
>
> **SSoT del UI = Storybook** (`mockups/INDEX.md` regla dura 1). El `.html` firmado es el contrato
> visual de ESTE paquete; el baseline del widget es el código
> (`web/src/widgets/chat-dock/ui/chat-dock.tsx`, 375 líneas).

## 0 · Cómo leer esto

| sección | qué resuelve |
|---|---|
| §1 | el árbol de archivos exacto: qué nace, qué cambia, qué **no se toca** |
| §2 | **Contradicciones detectadas** — cada choque mockup ↔ código ↔ decisión, con veredicto |
| §3 | props y estado, con el tipo |
| §4 | estados visuales exhaustivos por componente |
| §5 | tokens, **medidos** — ninguno inventado |
| §6 | geometría y layout |
| §7 | el wire (tipos TS + frame nuevo) |
| §8 | a11y: roles, `aria-*`, foco, teclado |
| §9 | dónde vive el estilo (Tailwind vs CSS) |
| §10 | orden de construcción |

---

## 1 · Contrato de componentes

### 1.1 Nacen

| Componente | Ruta exacta | Capa FSD y por qué | RF |
|---|---|---|---|
| `ConversacionRow` | `web/src/widgets/chat-dock/ui/conversacion-row.tsx` | **widgets**, slice `chat-dock`. Es *chrome* del dock: la taxonomía `canvas-not-chrome` de `.dependency-cruiser.js` lista `widgets/chat-dock` como chrome, y este componente es literalmente una fila del dock. No es `shared/ui` porque conoce `Conversacion` (dominio) ⇒ `ui-not-domain` lo prohibiría | RF-326…RF-332 |
| `CtxChip` + `IdentidadDetalle` | `web/src/widgets/chat-dock/ui/ctx-chip.tsx` | **widgets**, misma slice. Archivo propio y no dentro de `conversacion-row.tsx` porque tiene **estado propio** (abierto/cerrado), 5 estados visuales y dos disparadores automáticos (RF-314) — es la pieza con más superficie de test del paquete | RF-327 · RF-328 · RF-330 |
| `ConversacionesPanel` | `web/src/widgets/chat-dock/ui/conversaciones-panel.tsx` | **widgets**, misma slice. Es el organismo que toma el área del transcript: buscador + rótulo de alcance + lista + pie | RF-317…RF-325 |
| `ConversacionFila` | `web/src/widgets/chat-dock/ui/conversacion-fila.tsx` | **widgets**, misma slice. Separado del panel porque son **8 estados visuales** (§4.4) y el panel sólo compone | RF-319 · RF-322 · RF-323 |
| `useConversaciones` (store nuevo) | `web/src/widgets/chat-dock/model/conversaciones-store.ts` | **widgets/model**. Precedente literal: `portafolio-picker-store.ts:1-8` («SessionRail es un widget que hoy solo consume stores … se resuelve con este store propio del widget que envuelve `api.*` — mismo patrón "el store consume el transporte" que ya usa sessions-store, no la página»). Los componentes de arriba son **props-puras** | RF-334 |
| `Conversacion` (tipo) | `web/src/shared/api/types.ts` | **shared**. Precedente decisivo: `Session` vive ahí (`types.ts:16-38`) y **no tiene slice de `entities/`**. `Conversacion` es su hermana; abrirle una entity para un único consumidor sería inventar asimetría | RF-300 |

> **Por qué NO nace `entities/conversacion/`.** Tres razones, en orden de peso: (1) el precedente de
> `Session`, que es exactamente el mismo tipo de objeto y vive en `shared/api/types.ts`; (2)
> `domain-not-transport` (`.dependency-cruiser.js`, `error`) prohíbe que `entities/*/model` importe
> `shared/api`, así que la entity necesitaría duplicar el tipo del wire; (3) `fsd/insignificant-slice`
> está **apagado** en `steiger.config.ts` justamente porque el repo no quiere pelear por esto — o
> sea, no hay enforcer que empuje a crear la slice, y sí hay uno (`domain-not-transport`) que la
> encarece.

### 1.2 Se modifican (superset estricto — BR-CV-11)

| Archivo | Qué gana / qué pierde | RF |
|---|---|---|
| `web/src/widgets/chat-dock/ui/chat-dock.tsx` (`ChatDock`, `:12-42`) | **gana** `<ConversacionRow>` bajo el header y el `<ConversacionesPanel>` en lugar de `<Messages>` cuando la lista está abierta. **Pierde** `<SessionLine>` como fila fija. **Cambia** `⟩`→`»` (`:32`) | RF-326 · RF-332 |
| ídem, `ScopeRow` (`:46-76`) | pasa a renderizar `null` sin `scope`; con `scope` conserva el chip **literal** (`:55-70`) y pierde el rótulo «Alcance:» (`:51`), el chip punteado del arnés (`:52-54`) y el hint (`:72`) | RF-329 |
| ídem, `SessionLine` (`:78-93`) | **se retira como componente exportado en la columna** y su cuerpo se muda a `IdentidadDetalle`, con `cwd` agregado. Es mudanza, no borrado | RF-328 |
| ídem, `Messages` (`:95-154`) | **gana** el copy de vacío nuevo (reemplaza `:110-115`) y el render de la marca inline de rotación como `RolSys` centrado, que ya existe en `Bubble` (`:234-240`) — **cero código nuevo para la marca** | RF-307 · RF-313 |
| `web/src/shared/api/client.ts` | **gana** 5 métodos (`conversaciones`, `crearConversacion`, `activarConversacion`, `renombrarConversacion`, y el `q` de búsqueda como parámetro del primero). **Pierde** `conversacionesDeArnes` (`:218-221`) y `historialCerrada` (`:222-225`) | RF-340…RF-345 |
| `web/src/shared/api/types.ts` | `Session` pierde `claude_session_id`, `model`, `ctx_pct`, `conv`, `cadena_cc`, `turnos`, `cerrada_en` (`:26-37`); `DockFrame` (`:68-90`) gana un `kind` nuevo (§7.2) | RF-300 · H-2 |
| `web/src/shared/store/sessions-store.ts` | `onDock` (`:262-459`) gana la rama del `kind` nuevo; `appendConv` (`:80-87`) y todo el ruteo por `session_id` **no cambian** (siguen siendo por sesión: la conversación activa es una sola) | H-2 |
| `web/src/widgets/session-rail/ui/new-session-picker.tsx` | **pierde** el import (`:15`), la llamada `cargar(e.clave)` (`:111`), el render (`:235`) y el componente `ConversacionesDelArnes` (`:263-307`) | RF-333 |
| `web/src/widgets/session-rail/model/conversaciones-store.ts` (64 líneas) + `conversaciones-store.test.ts` | **se eliminan** — el store se reescribe en el dock. `no-sibling-widget-imports` (`.dependency-cruiser.js`) hace que no haya alternativa a mudarlo físicamente | RF-334 |

### 1.3 NO se tocan — y es un criterio de aceptación

| Archivo / símbolo | por qué |
|---|---|
| `chat-dock.tsx` → `Composer` (`:268-375`), `fitComposer` (`:263-266`) | la fila del composer no cambia. Incluye los 3 placeholders (`:326-332`), el botón `■` (`:351-359`), el `↑` (`:360-369`), el `ResizeObserver` (`:284-290`) |
| `chat-dock.tsx` → `Bubble` (`:233-258`), `ActivityCard` (`:183-231`), `agrupar` (`:160-172`), `rotulo` (`:176-179`) | superficie firmada en la PARIDAD de HS-26 (chat dock legible) |
| `markdown.tsx`, `permission-card.tsx`, `dictado-button.tsx` (`VoiceBar`, `DictadoAviso`) | fuera de alcance total |
| `web/src/pages/shell/ui/shell-page.tsx` | el resize del dock (`:13-18`, `:61-84`) ya cubre RF-317 CA-4. **Cero cambios** |
| `web/src/widgets/session-rail/ui/session-rail.tsx` | el rail crea **sesiones**, no conversaciones (CV-D2). El `✕ Cerrar sesión` (`:241-253`) y el renombrado del frente (`:256-273`) se conservan |
| `web/src/widgets/view-strip/`, `widgets/topbar/`, `widgets/map-canvas/` | CV-D1 y alcance |
| `web/src/shared/ui/indicators.tsx` (`Pip`, `HealthDot`) | se reusan tal cual |

---

## 2 · Contradicciones detectadas

Ninguna se resuelve en silencio. Cada una: qué choca, opciones, veredicto **y por qué**.

| id | contradicción | opciones | veredicto |
|---|---|---|---|
| **C-1** | **El mockup dibuja `＋` y las filas de la lista siempre habilitados** (`:355`, `:526-546`), pero BR-CV-6 exige que con un turno en vuelo (`streaming`/`await`) no se pueda crear ni retomar — y el servidor va a responder `409` (`session_service.go:339-342`). | (a) spec gana: deshabilitado con `title` · (b) mockup gana: siempre habilitado, el 409 se muestra como error | **(a) spec gana.** El mockup es un dibujo estático sin backend: no tiene forma de dibujar un estado que depende de `status`. Mostrar un error 409 después de un click que se veía legal es peor UX que un control apagado que dice por qué. **Es superset**: el mockup no pierde nada, gana un estado que no dibujó. Patrón ya sancionado: `new-session-picker.tsx:362-363` (`disabled` + `title="sin copia local registrada"`) |
| **C-2** | **CV-D14 dice «el cromo baja de 4 filas a 2»**, pero §2C del propio mockup dibuja **3** (`:420-439`: header + convrow + scope). | (a) la decisión miente · (b) el conteo es del reposo | **(b).** La tabla de CV-D14 lo dice ella misma: fila 3 «**solo con un nodo elegido**». El «2» es el reposo, que es el 90 % del tiempo; con nodo son 3 y la fila **se gana el lugar** (`mockup:450-452`). No hay contradicción de fondo — se declara para que nadie cuente mal en la PARIDAD |
| **C-3** | 🔴 **El mockup pinta texto en `--warn`** (`:887-888`, el error de lectura; `:187` `.ctxchip.hot`). **Medido hoy: `--warn` `#c96a2e` sobre `--card` `#ffffff` = 3,76:1** — bajo el mínimo 4,5 de axe para texto. Sobre `--secondary` 3,31, sobre `--background` 3,53, sobre `--accent-soft` 3,29, sobre `--crit-soft` 3,19. En oscuro pasa (7,50:1). Y **ya es deuda abierta**: `BACKLOG.md:59-64` y `:207-211` («rompe 4 stories de `session-rail/new-session-picker`»). | (a) gate a11y gana: el texto va en `--foreground`, `--warn` queda como señal **no textual** (borde/punto: 3,76 ≥ 3 ✓) · (b) mockup gana y se baja `a11y` a `"todo"` en las stories · (c) arreglar el token | **(a) gate gana.** (b) es exactamente lo que `preview.ts:7` (`a11y: { test: "error" }`) existe para impedir, y este paquete no puede introducir la cuarta y quinta story rota. (c) toca `web/tokens/base.tokens.json` y el BACKLOG ya lo declaró «merece su propio paquete». **Realización:** §5.3 |
| **C-4** | **`shared/ui/estado-carga.tsx` ya expone `Skeleton` y `ErrorBody`** con el contrato ARIA exacto que el mockup pide — pero usan clases `pf-*` (`pf-skeleton`, `pf-estado-vacio`, `pf-btn-primary`), que viven en el CSS del Portafolio. Reusarlas obligaría a envolver el dock en `className="arnesia-portafolio"`, que es lo que `new-session-picker.tsx:133` tuvo que hacer. Y el mockup apunta a **otro** componente: `PickerSkeleton` (`new-session-picker.tsx:309-326`), que es Tailwind puro. | (a) reusar `shared/ui/estado-carga` + scope `arnesia-portafolio` en el dock · (b) calcar el **contrato ARIA** de `PickerSkeleton` en Tailwind, dentro del dock · (c) promover `PickerSkeleton` a `shared/ui` | **(b), y es lo que el mockup manda** (`:872-875`: «Mismo esqueleto que `PickerSkeleton` … No se inventa un spinner nuevo»). (a) arrastra el scope CSS del Portafolio a un widget que hoy es 100 % Tailwind — acoplamiento nuevo por 20 líneas. (c) sería tentador pero `PickerSkeleton` desaparece con RF-333: promover algo que se está por borrar es trabajo perdido. **Lo que se reusa es el contrato**: `role="status"` + `aria-live="polite"` + `aria-label` + barras `aria-hidden` |
| **C-5** | **El texto de la marca de rotación difiere.** Mockup `:785`: «⟳ contexto rotado · checkpoint». Código: `breadcrumbRotacion = "— contexto rotado, seguimos —"` (`session_rotacion.go:12`), ya persistido en el `Conv` de conversaciones vivas. | (a) código gana · (b) mockup gana y se reescribe la constante | **(a) código gana.** Cambiar la constante dejaría los transcripts ya persistidos con la marca vieja y los nuevos con la nueva — dos marcas para lo mismo, sin migración posible (el `Conv` es dato del operador). El **aspecto** del mockup sí gana: la marca se pinta centrada, punteada, mono, `--muted-foreground` (`mockup:237-238`), que es exactamente lo que `Bubble` ya hace con `RolSys` (`chat-dock.tsx:234-240`) |
| **C-6** | 🔴 **El §4C del mockup no es realizable hoy: la rotación no emite ningún frame SSE.** Verificado: `session_rotacion.go` no tiene ni una llamada a `s.publish(...)`; las 14 del usecase están todas en `session_service.go`. El breadcrumb se persiste sólo en el `Conv` del backend, y el `conv` del FE es un espejo propio armado por `appendConv` (`sessions-store.ts:80-87`) desde frames. | (a) emitir un frame al rotar · (b) que el FE refetchee la sesión tras cada `result` · (c) dejarlo (la marca aparece al recargar) | **(a).** (b) es un GET por turno para un evento que ocurre cada ~40 % de contexto. (c) hace que el ctx caiga de 92 % a 12 % **sin explicación**, que es literalmente lo que el mockup dice que hay que evitar (`:798-800`). Va con el frame nuevo de §7.2 |
| **C-7** | **§1 del mockup conserva `⟩ colapsar`** (`:299`) mientras §2-§6 usan `» colapsar`. | — | **No es contradicción: es el antes y el después.** §1 está etiquetado `vigente` (`:295`). Se declara para que nadie «corrija» §1 en una futura re-derivación |
| **C-8** | **La estructura de la lista del mockup es `<button>` con `<span>` hijos** (`:518-525`), pero el patrón de lista accesible que el repo ya sancionó es `<ul>` → `<li>` → `<button>` (`new-session-picker.tsx:211-223`, con la nota de `:424-425` sobre por qué `role="group"` y no `<ul>` en la sub-lista). | (a) repo gana la estructura, mockup gana el aspecto · (b) mockup literal | **(a).** El mockup es HTML de dibujo; su `<button><span class="body">` no expresa «lista de N opciones» a un lector de pantalla. Se usa `<ul role="listbox">` con `<li role="option" aria-selected>` (§8.2) — el aspecto queda **idéntico** |
| **C-9** | **El pie de la lista sólo tiene `Cancelar`** (`mockup:548`), mientras el picker del rail tiene `Cancelar` + `Crear sesión` (`new-session-picker.tsx:237-253`). | — | **El mockup gana y tiene razón**: en el picker, elegir un arnés no crea nada (hace falta confirmar); acá **seleccionar ES la acción** (CV-D11, RF-310). Un botón «Retomar» sería un segundo paso para algo que ya pasó. Se declara para que nadie lo agregue «por simetría» |
| **C-10** | **Dos controles abren la misma superficie**: el chevron del título (`mockup:349`, `:504` rotado 90°) y el `🔍` (`:354`). | (a) unificar en uno · (b) conservar los dos con destinos de foco distintos | **(b), como el mockup.** Son dos intenciones: «ver mis conversaciones» y «buscar en ellas». El destino del foco las distingue (RF-355, §8.3): el chevron ⇒ foco en la fila activa; el `🔍` ⇒ foco en el buscador. Con una sola conversación no hay buscador (RF-325) ⇒ el `🔍` se oculta, no se deshabilita |
| **C-11** | **`.resumebar.bad` está definido en el CSS del mockup** (`:242`, `background: var(--crit-soft)`) **pero ningún panel lo usa.** | (a) desecharlo · (b) usarlo para el fallo de retoma | **(b).** RF-348 lo necesita (el `--resume` que falla y se auto-sana). Es superset autorizado: el dibujo ya lo preveía y se olvidó de pintarlo. **Con la corrección de C-3**: el texto va en `--foreground` sobre `--crit-soft` (15,89:1 claro / 14,12:1 oscuro ✓), no en `--warn` (3,19:1 ✗) |
| **C-12** | **El schema `Session` del OpenAPI ya está stale antes de este paquete**: `conv.items.rol` declara `enum: [user, assistant, sys]` (`openapi.yaml:877`) pero el dominio tiene 4 roles desde `RolAct` (`internal/domain/session.go:56`), y faltan `cwd`, `cadena_cc`, `checkpoint`, `ctx_hist`, `rotacion_pendiente`, `reparacion`, `cerrada_en`, `turnos`. Tampoco documenta `?arnes=`, `?cerradas=1` ni `/sessions/cerradas/{id}/historial`. | — | **Se corrige en el mismo commit** (RF-346). No es alcance nuevo: partir `Session` obliga a tocar ese schema de todos modos, y dejarlo mitad-corregido sería peor |
| **C-13** | **El anillo de foco del repo no llega a 3:1 en tema claro.** `new-session-picker.tsx:193` usa `focus:outline-2 focus:outline-primary`; medido: `--primary` `#00b7aa` sobre `--card` = **2,51:1**; `--ring` `#1fc6b8` = **2,14:1**. | (a) conservar el precedente y declarar la medición · (b) inventar un anillo propio para el dock | **(a).** WCAG 2.2 SC 2.4.11 no está en el ruleset por defecto de axe, así que ni rompe el gate ni lo tapa. Inventar un anillo distinto **sólo en el dock** crearía dos vocabularios de foco en la misma app, que es peor que una deuda declarada. **Se anota al BACKLOG** junto a la deuda de `--warn` (mismo origen: valores de token) |

---

## 3 · Props y estado

Todos los componentes de `ui/` son **props-puras**: cero `api.*`, cero `useConversaciones` adentro.
El cableado ocurre en `ChatDock`, que es el único que lee stores. Precedente:
`new-session-picker.tsx:18-24` («Props PURAS: CERO transporte»).

### 3.1 El tipo del wire

```ts
// web/src/shared/api/types.ts — al lado de Session (:16-38)

export interface Conversacion {
  id: string
  /** título auto-derivado del primer turno `user`, o "nueva conversación". */
  titulo: string
  /** true ⇒ ningún turno vuelve a re-derivar el título (RF-303). */
  titulo_editado: boolean
  /** exactamente una por sesión (BR-CV-1). */
  activa: boolean
  turnos: number
  ctx_pct?: number | undefined
  /** RFC3339 UTC. VACÍO cuando turnos === 0 — no se inventa (BR-CV-14). */
  ultima_interaccion?: string | undefined
  claude_session_id?: string | undefined
  model?: string | undefined
  /** solo con ?q= — el fragmento que coincidió, ya recortado por el daemon (RF-322). */
  fragmento?: string | undefined
}

export interface ConversacionesListado {
  conversaciones: Conversacion[]
  /** el denominador del rótulo «N de M coinciden» (RF-318). */
  total: number
}
```

> Los `| undefined` explícitos son obligatorios: `tsconfig` extiende `@tsconfig/strictest`
> (`package.json:51`) ⇒ `exactOptionalPropertyTypes`. Misma disciplina que `Session`
> (`types.ts:26-27`).

### 3.2 `ConversacionRow`

```ts
export interface ConversacionRowProps {
  /** la conversación activa. Siempre existe (BR-CV-1). */
  activa: Conversacion
  /** el arnés y el cwd salen de la SESIÓN, no de la conversación (RF-300). */
  arnes: string
  cwd?: string | undefined
  /** `streaming`|`await` ⇒ ＋ y las filas quedan deshabilitados con motivo (RF-312). */
  status: SessionStatus
  /** la lista está desplegada ⇒ el chevron rota 90° (mockup:504). */
  listaAbierta: boolean
  /** el detalle de identidad se abre solo al retomar y al rotar (RF-314). */
  detalleForzado: boolean
  onToggleLista: (foco: "filas" | "buscador") => void
  onNueva: () => void
  onRenombrar: (titulo: string) => void
}
```

Estado local: `editando: boolean` + `valor: string` (calca `session-rail.tsx:194-195`).
`detalleAbierto: boolean` vive en `CtxChip` y se **fuerza** desde `detalleForzado`.

### 3.3 `CtxChip`

```ts
export interface CtxChipProps {
  ctxPct: number            // 0 cuando no hay dato — 0 es dato (BR-CV-9)
  claudeSessionId?: string | undefined  // ausente ⇒ "sin sesión CC" (chat-dock.tsx:82)
  arnes: string
  model?: string | undefined
  cwd?: string | undefined
  /** ctxPct >= umbral ⇒ variante `hot` (la barra en --warn, el número NO — C-3). */
  caliente: boolean
  abierto: boolean
  onToggle: () => void
}
```

> **`caliente` viene por prop, no se calcula.** El umbral vive en el daemon
> (`SetUmbralRotacion`, `session_service.go:141-145`, default 40 por
> `cmd/arnesia/main.go:120`) y el FE **no lo conoce**. Se deriva de `rotacion_pendiente` del wire,
> no de un `>= 40` tecleado en el FE.

### 3.4 `ConversacionesPanel`

```ts
export type PanelEstado = "cargando" | "error" | "datos"   // calca PickerEstado
                                                            // (portafolio-picker-store.ts:15)

export interface ConversacionesPanelProps {
  estado: PanelEstado
  error?: string | undefined
  conversaciones: Conversacion[]
  total: number
  /** el Frente de la SESIÓN — el rótulo lo nombra (RF-318, mockup:516). */
  frenteSesion: string
  busqueda: string
  /** `streaming`|`await` ⇒ filas deshabilitadas con motivo (RF-312). */
  bloqueadoMotivo?: string | undefined
  /** foco inicial: "buscador" si entró por 🔍, "filas" si entró por ▶ (C-10). */
  focoInicial: "filas" | "buscador"
  onBusqueda: (q: string) => void
  onElegir: (id: string) => void
  onCancelar: () => void
  onReintentar: () => void
}
```

### 3.5 `ConversacionFila`

```ts
export interface ConversacionFilaProps {
  conversacion: Conversacion
  /** aria-disabled + title; nunca un control apagado y mudo (C-1). */
  deshabilitadaMotivo?: string | undefined
  /** para <mark> — el daemon manda el fragmento, el FE resalta el término (RF-322). */
  termino?: string | undefined
  onElegir: () => void
}
```

### 3.6 El store del widget

```ts
// web/src/widgets/chat-dock/model/conversaciones-store.ts
interface ConversacionesState {
  sesionId: string | null
  estado: PanelEstado
  error?: string | undefined
  conversaciones: Conversacion[]
  total: number
  busqueda: string
  abierta: boolean
  focoInicial: "filas" | "buscador"
  /** true mientras --resume rehidrata: pinta la franja efímera (RF-310). */
  retomando: string | null
  /** el motivo del fallo de retoma: pinta .resumebar.bad (C-11, RF-348). */
  retomaFallo?: string | undefined

  abrir: (sesionId: string, foco: "filas" | "buscador") => void
  cerrar: () => void
  buscar: (q: string) => Promise<void>
  crear: () => Promise<void>
  activar: (cid: string) => Promise<void>
  renombrar: (cid: string, titulo: string) => Promise<void>
  reintentar: () => Promise<void>
}
```

**Disciplinas heredadas, no inventadas:**
1. **Refetch en cada apertura**, jamás lista cacheada de una apertura anterior —
   `portafolio-picker-store.ts:22-25` (RF-15/TS-D14). El operador pudo haber rotado o mandado un
   turno desde la última vez.
2. **`cerrar()` descarta búsqueda y estado** — `session-rail.tsx:43-47` («cerrar descarta
   búsqueda/selección … cero efectos»).
3. **El error del backend se guarda tal cual**, sin reescribirlo — `conversaciones-store.ts:39-41`
   y `portafolio-picker-store.ts:45-47` (`e instanceof Error ? e.message : String(e)`).

---

## 4 · Estados visuales exhaustivos

### 4.1 `ConversacionRow` (fila 2 del cromo)

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **reposo** | `▶ título` · chip ctx · `🔍` · `＋`. Chevron sin rotar | `:348-357` |
| 2 | **hover en el título** | fondo `--secondary` en el botón del título, radio `--radius-sm` | CSS `:177` |
| 3 | **foco en cualquier control** | anillo `outline-2 outline-primary` (precedente `new-session-picker.tsx:193`, C-13) | — |
| 4 | **lista abierta** | chevron rotado 90° | `:504` |
| 5 | **editando el título** | el `input` **toma la fila entera**; ctx y acciones no se renderizan | `:465-467` |
| 6 | **turno en vuelo** | `＋` con `disabled` + `title` del motivo; opacidad 40 % (calca `chat-dock.tsx:366`, `disabled:opacity-40`) | superset (C-1) |
| 7 | **título largo** | truncado con elipsis; el `title` nativo lleva el completo | CSS `:179` |
| 8 | **retomando** | la franja `resumebar` se inserta **debajo** de la fila, no dentro | `:709-712` |

### 4.2 `CtxChip`

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **cerrado, frío** | barra 44×5 con relleno `--primary` + `N%` en `--muted-foreground`, sin borde | `:350-352`, CSS `:184` |
| 2 | **hover** | fondo `--secondary`, borde `--border`, texto `--foreground` | CSS `:185` |
| 3 | **abierto** | mismo tratamiento que hover, persistente; `aria-expanded="true"` | `:388`, CSS `:186` |
| 4 | **caliente** (`rotacion_pendiente`) | **la barra** en `--warn`; el número **NO** cambia de color (C-3) | CSS `:187-188` corregido |
| 5 | **0 %** | barra vacía + `0%`. **No se esconde** (BR-CV-9) | `:654`, `:742` |
| 6 | **foco** | anillo, igual que 4.1 #3 | — |

`IdentidadDetalle` (el cuerpo desplegado):

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **con sesión CC** | línea 1 `◍ 4b046945 · vitalia · claude-opus-5[1m]`; línea 2 `cwd ~/Proyectos/…` | `:396-399` |
| 2 | **sin sesión CC** | `◍ sin sesión CC` — literal vigente (`chat-dock.tsx:82`) | `:764-765` |
| 3 | **sin cwd** | la línea 2 **no se dibuja** (no se inventa un `—`) | — |
| 4 | **cwd larguísimo** | truncado con elipsis, `title` con el completo | — |

### 4.3 `ConversacionesPanel`

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **cargando** | 3 barras fantasma `--secondary` con opacidad 1 / .7 / .4, `role="status"` | `:864-869` |
| 2 | **error** | motivo + `Reintentar`. **Texto en `--foreground`**, señal `--warn` no textual (C-3) | `:886-890` |
| 3 | **datos, N ≥ 2** | buscador + rótulo `N conversaciones de la sesión X` + filas + pie `Cancelar` | `:511-549` |
| 4 | **datos, N = 1** | **sin buscador**; rótulo `1 conversación de la sesión X`; la fila; el mensaje de `:672-675` | `:660-677` |
| 5 | **buscando, con coincidencias** | rótulo `N de M coinciden`; filas con `<mark>` en el fragmento | `:578-602` |
| 6 | **buscando, sin coincidencias** | el vacío que dice dónde buscó + `Limpiar búsqueda`; el pie `Cancelar` se conserva | `:629-640` |
| 7 | **bloqueado por turno en vuelo** | filas `aria-disabled` + `title`; el buscador **sigue activo** | superset (C-1) |
| 8 | **input con foco** | borde `--primary` | CSS `:209` |

### 4.4 `ConversacionFila`

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **inactiva, reposo** | radio vacío, título, meta mono | `:526-532` |
| 2 | **inactiva, hover** | fondo `--secondary` | CSS `:216` |
| 3 | **activa** | borde `--primary` + fondo `--accent-soft` + radio relleno + rótulo `activa` en `--foreground` | `:518-525`, CSS `:217-220`, `:226-228` |
| 4 | **con fragmento** | tercera línea con `<mark>` sobre `--accent-soft` | `:590`, CSS `:225` |
| 5 | **0 turnos** | meta `sin turnos todavía · ctx 0 %` | `:667` |
| 6 | **sin `ultima_interaccion`** (migrada, con turnos) | meta `sin fecha · 41 turnos · ctx 34 %` — se dice, no se inventa (H-4) | superset |
| 7 | **deshabilitada** | opacidad 50 % + `cursor-not-allowed` + `title` (calca `new-session-picker.tsx:367`) | superset (C-1) |
| 8 | **foco** | anillo, sin depender del fondo | — |

### 4.5 `ScopeRow` modificado

| # | estado | qué se ve |
|---|---|---|
| 1 | **sin nodo** | **no existe en el DOM** |
| 2 | **con nodo** | el chip literal del vigente (`chat-dock.tsx:55-70`) |
| 3 | **hover en el ✕** | `hover:text-destructive` — literal vigente (`chat-dock.tsx:66`) |
| 4 | **`fuentePath` ausente** | el chip no lo dibuja — condicional ya vigente (`chat-dock.tsx:59-61`) |

### 4.6 `Messages` — los dos añadidos

| # | estado | qué se ve | mockup |
|---|---|---|---|
| 1 | **vacío de conversación nueva** | «Pídele un cambio a **arnés**. «<título anterior>» quedó guardada — la retomás desde ▶» | `:749-752` |
| 2 | **vacío de la primera conversación** de una sesión | mismo copy **sin** la segunda oración (no hay anterior que nombrar) | superset |
| 3 | **marca de rotación** | `RolSys` centrado, punteado, mono — **cero código nuevo** (`Bubble`, `chat-dock.tsx:234-240`) | `:785` |
| 4 | **marca de hilo reiniciado** (RF-348) | idéntica, otro texto | superset (C-11) |

---

## 5 · Tokens

**Sólo DTCG reales de `web/src/app/styles/theme.css`** (83 propiedades en `:root`, 45 sobrescritas
en `:root[data-theme="dark"]`). Nada inventado, nada de la paleta ámbar pre-rebrand.

### 5.1 Tabla de uso

| elemento | token | por qué |
|---|---|---|
| fondo del dock y de los paneles | `--card` | ya vigente |
| separadores de fila | `--border` | ya vigente (`chat-dock.tsx:21`) |
| título de la conversación | `--foreground`, `--text-xs`, `600` | ya vigente para el frente (`chat-dock.tsx:22`) |
| meta de la fila (fecha · turnos · ctx) | `--muted-foreground`, `--font-mono`, `9.5px` | calca `SessionLine` (`chat-dock.tsx:80`) |
| relleno de la barra de ctx (frío) | `--primary` | ya vigente (`chat-dock.tsx:88`) |
| relleno de la barra de ctx (**caliente**) | `--warn` | **no textual** ⇒ mínimo 3:1. Medido: 3,76 claro / 7,50 oscuro ✓ |
| número del ctx | `--muted-foreground` (reposo) / `--foreground` (hover, abierto, **caliente**) | **nunca `--warn`** (C-3) |
| fila activa: borde | `--primary` | no textual, 2,51 claro ✗ para 3:1 — **se acompaña** de `--accent-soft` de fondo + el radio relleno + el rótulo `activa` (RF-356: nada depende sólo del color) |
| fila activa: fondo | `--accent-soft` | ya vigente (`chat-dock.tsx:246`) |
| rótulo `activa` | `--foreground` | medido: **16,39:1** claro sobre `--accent-soft` / **13,2:1** oscuro. En `--primary` daría 2,20 ✗ (`mockup:226-227`) |
| glifo `＋` | `--foreground` | medido: 18,74:1 claro. En `--primary` daría **2,51** ✗ (`mockup:198-199`) |
| `＋` hover | fondo `--accent-soft` | el acento entra por hover, no por color solo (`mockup:201`) |
| `<mark>` del fragmento | fondo `--accent-soft`, texto `--foreground` | 16,39:1 ✓ (`mockup:225`) |
| franja de retoma (ok) | fondo `--accent-soft`, texto `--foreground` | 16,39 / 13,2 ✓ (`mockup:241`) |
| franja de retoma (fallo) | fondo `--crit-soft`, texto `--foreground` | medido **15,89:1** claro / **14,12:1** oscuro ✓. En `--warn` daría 3,19 ✗ (C-11) |
| error de lectura | texto `--foreground`; regla/punto en `--warn` | C-3 |
| esqueleto de carga | `--secondary` + `--border` | calca `PickerSkeleton` (`new-session-picker.tsx:321`) |
| input del buscador | fondo `--secondary`, borde `--border`, foco `--primary` | calca `chat-dock.tsx:345` |
| radios | `--radius-sm` (4) chips y filas internas · `--radius-md` (8) filas de la lista e inputs · `--radius-full` barra de ctx y radio | ya vigentes |
| espaciado | `--space-1/2/3` | ya vigentes |
| tipografías | `--font-sans` cuerpo · `--font-mono` meta e identidad | ya vigentes |

### 5.2 Contrastes medidos (método: WCAG 2.x relative luminance, alfa compuesto sobre el fondo real)

| par | claro | oscuro | umbral | veredicto |
|---|---|---|---|---|
| `--foreground` / `--card` | **18,74** | **16,09** | 4,5 | ✓ |
| `--foreground` / `--accent-soft` | **16,39** | **13,18** | 4,5 | ✓ (coincide con el mockup: 16,39 / 13,22) |
| `--foreground` / `--secondary` | **16,50** | **14,90** | 4,5 | ✓ |
| `--foreground` / `--crit-soft` | **15,89** | **14,12** | 4,5 | ✓ |
| `--muted-foreground` / `--card` | **7,65** | **6,81** | 4,5 | ✓ |
| `--muted-foreground` / `--secondary` | **6,73** | **6,60** | 4,5 | ✓ |
| `--muted-foreground` / `--accent-soft` | **6,69** | **5,58** | 4,5 | ✓ |
| `--warn` / `--card` — **texto** | **3,76** | 7,50 | 4,5 | ✗ **claro** → C-3 |
| `--warn` / `--card` — **no textual** | 3,76 | 7,50 | 3,0 | ✓ (la barra caliente) |
| `--warn` / `--crit-soft` — texto | **3,19** | — | 4,5 | ✗ → C-11 |
| `--primary` / `--card` | 2,51 | 8,91 | 3,0 | ✗ **claro** — por eso `＋` va en `--foreground` |
| `--primary` / `--card` como **borde de fila activa** | 2,51 | 8,91 | 3,0 | ✗ claro → se acompaña de fondo + radio + rótulo (RF-356) |
| `--ring` / `--card` (anillo de foco) | 2,14 | — | 3,0 (SC 2.4.11) | ✗ → C-13, deuda declarada |

### 5.3 La corrección de C-3, concreta

```
mockup:187-188            →  realización
.ctxchip.hot {               .ctxchip[data-caliente] {
  color: var(--warn);          color: var(--foreground);   /* el NÚMERO */
}                            }
.ctxchip.hot .ctxbar i {     .ctxchip[data-caliente] .barra > i {
  background: var(--warn);     background: var(--warn);    /* la BARRA: no textual, 3,76 ≥ 3 ✓ */
}                            }

mockup:887-888            →  el motivo en --foreground; el tono de alarma lo da
.empty { color: var(--warn) }  un borde-izquierdo de 3px en --warn (no textual)
```

---

## 6 · Geometría

| medida | valor | de dónde sale |
|---|---|---|
| ancho del dock | 300 px mín · 60 % de la ventana máx · 360 default | `shell-page.tsx:16-18`, **no cambia** |
| ancho de la lista | **el del dock** | RF-317 CA-4: sin mínimo propio; las filas truncan |
| alto de la fila 2 (cromo) | ~26 px (`padding: 5px 12px 5px 10px`) | `mockup:175` |
| alto de la fila 3 (alcance) | ~24 px, **sólo con nodo** | `mockup:156` |
| alto de una fila de la lista | ~34 px sin fragmento · ~50 px con | `mockup:215`, `:224` |
| barra de ctx | 44 × 5 px, `--radius-full` | literal vigente (`chat-dock.tsx:87`) |
| botones de icono | 24 × 24 px | `mockup:196` |
| área de la lista | `min-height: 96px`, crece hasta empujar el composer | `mockup:205` |
| ganancia neta | **~46 px** de transcript en reposo (2 filas retiradas de ~23 px) | §1 vs §2A del mockup, mismo alto de panel |

---

## 7 · El wire

### 7.1 Cliente

```ts
// web/src/shared/api/client.ts — al lado de listSessions (:214)
conversaciones: (sesionId: string, q?: string) =>
  req<ConversacionesListado>(
    `/api/sessions/${encodeURIComponent(sesionId)}/conversaciones${q ? `?q=${encodeURIComponent(q)}` : ""}`,
  ),
crearConversacion: (sesionId: string) =>
  req<{ nueva: Conversacion; desactivada: string | null }>(
    `/api/sessions/${encodeURIComponent(sesionId)}/conversaciones`, { method: "POST" }),
activarConversacion: (sesionId: string, cid: string) =>
  req<Conversacion>(`/api/sessions/${…}/conversaciones/${…}/activar`, { method: "POST" }),
renombrarConversacion: (sesionId: string, cid: string, titulo: string) =>
  req<Conversacion>(`/api/sessions/${…}/conversaciones/${…}`,
    { method: "PATCH", body: JSON.stringify({ titulo }) }),
```

> **Sin timeout propio** y **sin `AbortSignal` en las mutaciones**: el tope vive en el daemon
> (precedente literal `client.ts:283-284`) y abortar una mutación a mitad es peor que esperar
> (precedente `client.ts:205-207`). El `GET` de búsqueda **sí** lleva `signal`, porque teclear
> rápido genera peticiones que se pisan — mismo criterio que `escanearProyecto` (`client.ts:119`).

### 7.2 El frame nuevo (cierra H-2 y C-6)

```ts
// web/src/shared/api/types.ts — DockFrame (:68-90) gana un kind
kind: … | "conversacion"

// campos nuevos, todos opcionales (el frame sigue siendo uno solo para todo el dock)
conversacion_id?: string
conversacion_evento?: "creada" | "activada" | "rotada" | "renombrada"
```

**Reglas del frame:**
1. **No lleva `run_id`**: no pertenece a un turno. Por eso el dedup de `finalizedRun`
   (`sessions-store.ts:266`) **no aplica** — la idempotencia es por `conversacion_id` +
   `conversacion_evento` (E-41).
2. Lo emiten: `crear`, `activar`, `renombrar` **y `rotarLocked`** (`session_rotacion.go:52-64`, que
   hoy no emite nada — C-6).
3. El FE lo usa para (a) refrescar la lista si está abierta y (b) appendear la marca de rotación al
   `conv` local vía `appendConv` (`sessions-store.ts:80-87`), que ya existe.

---

## 8 · Accesibilidad

### 8.1 `ConversacionRow`

| elemento | contrato |
|---|---|
| botón del título | `aria-expanded={listaAbierta}` · `aria-controls="<id>-panel"` · nombre accesible = el título · `title` con el título completo cuando trunca |
| chip de ctx | `<button aria-expanded aria-controls="<id>-detalle">` · nombre accesible: `«contexto 68 %, ver identidad de la conversación»` — el `%` **en el nombre**, no sólo en la barra |
| barra de ctx | `aria-hidden` (la cifra ya está en el nombre del botón) |
| `🔍` | `aria-label="Buscar en las conversaciones de esta sesión"` (literal `mockup:354`) |
| `＋` | `aria-label="Nueva conversación — desactiva la actual"` (literal `mockup:355`); deshabilitado ⇒ `disabled` + `title` con el motivo |
| input de renombrado | `aria-label="Título de la conversación"` · `autoFocus` (precedente `session-rail.tsx:259`) |

### 8.2 `ConversacionesPanel` — la lista

```
<section aria-label="Conversaciones de la sesión «repro del bug de carga»">
  <input type="search" aria-label="Buscar en estas conversaciones"
         aria-describedby="cv-alcance" />
  <p id="cv-alcance" aria-live="polite">4 conversaciones de la sesión «…»</p>
  <ul role="listbox" aria-label="Conversaciones" aria-activedescendant="cv-3">
    <li role="option" id="cv-1" aria-selected="true"> … </li>
    <li role="option" id="cv-2" aria-selected="false" aria-disabled="true"> … </li>
  </ul>
</section>
```

- **`<ul role="listbox">` y no `<button>` suelto** (C-8): expresa «N opciones, una elegida».
- El rótulo de alcance es `aria-live="polite"`: al filtrar, `N de M coinciden` se **anuncia**
  (RF-356).
- Las filas deshabilitadas llevan `aria-disabled="true"` (no `disabled`): siguen siendo alcanzables
  por lector de pantalla, que es lo que hace que el motivo se pueda oír. Precedente
  `new-session-picker.tsx:362-363`.
- Esqueleto: `role="status"` + `aria-live="polite"` + `aria-label="Cargando conversaciones"`, barras
  `aria-hidden` — calca `PickerSkeleton` (`new-session-picker.tsx:311-323`).
- Error: `role="alert"` en el motivo — calca `new-session-picker.tsx:145`.

### 8.3 Orden de foco y teclado

| gesto | resultado |
|---|---|
| `Tab` dentro del cromo | título → chip ctx → `🔍` → `＋` → (fila 3 si existe: chip → `✕`) → transcript → composer |
| abrir con **▶** | foco a la **fila activa** de la lista (C-10) |
| abrir con **🔍** | foco al **buscador** (precedente `autoFocus`, `new-session-picker.tsx:184-189`) |
| `↑` / `↓` en la lista | mueven `aria-activedescendant` entre filas; saltan las deshabilitadas sólo con `↑↓` repetido, nunca sin anunciar |
| `Enter` / `Espacio` en una fila | la elige (retoma) |
| `Escape` con la lista abierta | cierra la lista, descarta la búsqueda, **devuelve el foco** al control que la abrió |
| `Escape` editando el título | descarta, sale de edición, foco al botón del título (`session-rail.tsx:267-270`) |
| `Enter` editando | confirma (`session-rail.tsx:266`) |
| `blur` editando | confirma (`session-rail.tsx:263`) |
| `⌘K` / `Ctrl+K` | colapsa el dock entero — **sin cambios** (`App.tsx:15-17`); la lista se desmonta |

**Ningún atajo global nuevo** (RF-357).

### 8.4 Anuncios de estado

| evento | anuncio |
|---|---|
| lista filtrada | `N de M coinciden` por el `aria-live` del rótulo |
| retoma en curso | la franja lleva `role="status"`; texto «Retomando la conversación…» |
| retoma fallida | `role="alert"` con el motivo |
| conversación creada | el vacío del transcript ya nombra la desactivada — no hace falta un `aria-live` extra |
| rotación | la marca `RolSys` entra al transcript, que ya scrollea solo (`chat-dock.tsx:104-106`) |

---

## 9 · Dónde vive el estilo

**Todo Tailwind, dentro de los componentes.** El widget `chat-dock` hoy es 100 % utilidades
(`chat-dock.tsx` no importa ningún `.css`), y este paquete **no introduce un archivo CSS nuevo**.

- **No se usan las clases `pf-*`** (C-4): arrastrarían el scope `arnesia-portafolio` al dock.
- **No se usa `shell.css`** salvo por `Pip`, que ya lo hace (`indicators.tsx:7`).
- `stylelint` (`package.json:21`) sólo mira `src/**/*.css` ⇒ cero superficie nueva para él.
- El `<mark>` necesita `[&_mark]:bg-accent-soft [&_mark]:text-foreground` o un `<span>` propio; se
  prefiere el `<span className="…">` para no depender del reset del navegador.

---

## 10 · Orden de construcción

| # | tramo | por qué antes que el siguiente |
|---|---|---|
| **0** | **Story del `ChatDock` vigente** (`plan-storybook.md` §1.3) | el widget que se reescribe **no tiene story**: sin baseline no hay forma de probar que el superset no rompió nada |
| **1** | Modelo Go + migración + **borrado CV-D6** (RF-300…RF-306, RF-335…RF-339) | el borrado de datos va antes de que exista UI que los liste |
| **2** | API + OpenAPI (RF-340…RF-346) | la UI es props-puras y necesita el tipo |
| **3** | Frame nuevo + `rotarLocked` que publica (§7.2, C-6) | sin esto §4C del mockup no es realizable |
| **4** | `CtxChip` + `ConversacionRow` + `ScopeRow` condicional (RF-326…RF-332) | el cromo primero: es lo que cambia el alto disponible |
| **5** | `ConversacionesPanel` + `ConversacionFila` + store (RF-317…RF-325) | |
| **6** | Retomar/crear/rotar en vivo (RF-307…RF-316) | necesita 4 y 5 |
| **7** | Mudanza del picker (RF-333, RF-334) | **último**: hasta acá el operador conserva la superficie vieja, aunque muestre 0 |
