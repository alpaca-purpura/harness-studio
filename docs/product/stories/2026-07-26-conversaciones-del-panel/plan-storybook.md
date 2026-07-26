# Plan de stories · Las conversaciones viven en el panel de conversación

> `tipo: plan-pruebas` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Par de [`spec.md`](./spec.md) (RF-300…RF-357 · E-01…E-50) y [`design.md`](./design.md).
> **En este repo la story ES el test** (`docs/architecture/boundaries/fe-visual-fitness.md`, job
> `visual-fitness` de CI). Este documento define **qué stories escribir y qué asserta cada una**.
> El constructor lo ejecuta sin decidir nada; donde hay decisión pendiente, está marcada como tal.

## 0 · Resumen ejecutable

| | |
|---|---|
| Stories planificadas | **56** en **6 archivos** (5 nuevos · 1 superset de un archivo firmado) |
| Tests unitarios (project `unit`) | **1 archivo nuevo** (`conversaciones-store.test.ts`) · **1 archivo eliminado** (el del rail) |
| Componentes nuevos con story | 4 de 4 |
| Escenarios del spec cubiertos por story | **26 de 50** |
| Escenarios cubiertos por otro mecanismo, declarado | **23 de 50** (§4) |
| Escenarios **sin test automatizable**, declarados | **1** (E-46, el borrado de datos del operador — §4.3) |
| 🔴 Bloqueante | **`chat-dock.tsx` no tiene story hoy.** §1.3 |
| 🕳 Huecos del tooling | 3 (§5.3) |

---

## 1 · Cómo se corre y con qué se verifica

### 1.1 Comandos reales (verificados hoy en `web/package.json`)

| qué | comando | de dónde |
|---|---|---|
| **todas** las stories como test | `pnpm --dir web exec vitest --project=storybook run` | `vitest.config.ts:28-40` |
| **una** story | `pnpm --dir web exec vitest --project=storybook run <archivo>.stories.tsx` | ídem |
| los unitarios puros | `pnpm --dir web exec vitest --project=unit run` | `vitest.config.ts:20-27` |
| los dos | `pnpm --dir web run test` (= `vitest run`) | `package.json:24` |
| el workshop | `pnpm --dir web run storybook` (`-p 6006`) | `package.json:22` |
| los gates estáticos | `pnpm --dir web run verify` | `package.json:25` |

### 1.2 Qué corre en CI, y qué no

`.github/workflows/ci.yml:65-68`:

```yaml
- name: unit (selectores puros — S1-D7, Node, sin Chromium)
  run: pnpm exec vitest --project=unit run
- name: fitness visual (Storybook 10 = tests)
  id: visual-fitness
  run: pnpm exec playwright install --with-deps chromium && pnpm exec vitest --project=storybook run
```

⚠️ **`npm run verify` NO corre las stories** (`package.json:25`: typecheck · lint · depcruise · fsd ·
stylelint, y nada más). Quien construya este paquete tiene que correr `--project=storybook` a mano;
el único que lo obliga es CI.

### 1.3 🔴 Bloqueante — `chat-dock.tsx` no tiene story

Verificado: `web/src/widgets/chat-dock/ui/` tiene **42 archivos de story en todo el repo** y ninguno
es del dock. Los dos que hay en la carpeta son de piezas internas:
`dictado-button.stories.tsx` y `permission-card.stories.tsx`.

**Consecuencia:** el widget que este paquete reescribe entero no tiene baseline. Un superset sin
baseline no es verificable: no hay forma de probar que el composer, las burbujas y la tarjeta de
actividad quedaron intactos (`design.md` §1.3).

**Orden, no sugerencia:** el archivo `chat-dock.stories.tsx` se crea **antes** de tocar
`chat-dock.tsx`, con las 3 stories del vigente (§2.1 A-01…A-03). Se escriben contra el código de hoy
y **tienen que pasar en verde antes del primer cambio**.

### 1.4 Patrón obligatorio (calcado del repo, no inventado)

Verificado sobre los 42 archivos: **42/42** usan exactamente esta forma.

```tsx
import type { Meta, StoryObj } from "@storybook/react-vite"      // 42/42
import { expect, fn, userEvent, within } from "storybook/test"    // 42/42 — NO "@storybook/test"

const meta = { component: X, args: { onY: fn() }, … } satisfies Meta<typeof X>
export default meta
type Story = StoryObj<typeof meta>
```

- `@storybook/test` **no existe** en este árbol (0 ocurrencias): el subpath correcto es
  `storybook/test`.
- `vitest-browser-react` y `@testing-library/*` **no están instalados**. No se usan.
- **Sin `setupFiles`**: las preview annotations (tema + a11y) las auto-aplica
  `@storybook/addon-vitest` 10.3+ (`vitest.config.ts:11`, comentario explícito).
- **a11y en `error` por default** (`preview.ts:7`: `a11y: { test: "error" }`). Ninguna story de este
  paquete lo baja a `"todo"` — RF-353 CA-4.
- **Tema**: `globals: { theme: "dark" }` por story; el decorator global estampa
  `document.documentElement.dataset.theme` (`preview.ts:24-29`).

### 1.5 Cómo se stubbean los stores

`ChatDock` lee `useSessions` directo (`chat-dock.tsx:13-15`) y `ConversacionRow`/`Panel` son
props-puras (`design.md` §3). Dos regímenes:

| componente | régimen | precedente |
|---|---|---|
| `ConversacionRow` · `CtxChip` · `ConversacionesPanel` · `ConversacionFila` | **args puros + `fn()`** | `new-session-picker.stories.tsx:28-35` (`onCrear`/`onCancelar`/… todos `fn()`), `permission-card.stories.tsx:13` |
| `ChatDock` (el widget entero) | **decorator que reescribe el estado COMPLETO del store** antes de renderizar | `dictado-button.stories.tsx:16-29` — **el único archivo del repo que toca un store zustand en stories**. El decorator escribe **todos** los defaults explícitos, lo que lo hace idempotente entre stories |

```tsx
// helper del archivo, calcado de dictado-button.stories.tsx:16-29
const escenario = (parcial: Partial<SessionsState>, conv: Partial<ConversacionesState> = {}) =>
  (Story: () => ReactNode) => {
    useSessions.setState({ sessions: [], activeId: null, chatOpen: true, streaming: {},
      pendingPerms: {}, scope: {}, finalizedRun: {}, wroteInRun: {}, msgFlushed: {},
      connected: true, loaded: true, railCollapsed: false, ...parcial })
    useConversaciones.setState({ sesionId: null, estado: "datos", conversaciones: [], total: 0,
      busqueda: "", abierta: false, focoInicial: "filas", retomando: null, ...conv })
    return <div style={{ width: 360, height: 520 }}><Story /></div>
  }
```

**Regla dura:** el decorator escribe el estado **entero**, nunca un merge parcial sobre lo que dejó
la story anterior. Es lo que hace que el orden de ejecución no importe.

---

## 2 · Las 56 stories

### 2.1 `web/src/widgets/chat-dock/ui/chat-dock.stories.tsx` — **NUEVO** (17)

Fixtures: una sesión `repro del bug de carga` / arnés `vitalia` / cwd
`~/Proyectos/luana-vitalia/vitalia`, y 4 conversaciones — **los mismos datos del mockup §3A**, que a
su vez salen de la máquina real (`GET /api/sessions` verificado hoy).

| id | story | args / escenario | `play()` asserta | cubre |
|---|---|---|---|---|
| **A-01** | `VigenteReposo` | el dock de HOY, sin tocar nada | 4 filas de cromo presentes: header · SessionLine (`◍`, `ctx`) · ScopeRow («Alcance:») · composer | **baseline §1.3** |
| **A-02** | `VigenteConPermiso` | `pendingPerms` con 1 ask | la `PermissionCard` renderiza | baseline |
| **A-03** | `VigenteStreaming` | `status: "streaming"`, `streaming[id]: "…"` | el `■` presente, el `↑` ausente | baseline |
| **A-04** | `Reposo` | activa `por qué el mapa sale vacío`, 90 turnos, ctx 68 %, sin nodo | **exactamente 2 filas** de cromo: `queryByText("Alcance:") === null` ∧ `queryByText(/◍/) === null` ∧ `getByRole("button",{name:/contexto 68/})` presente ∧ `getByRole("button",{name:/» colapsar/})` | RF-326 · RF-332 · E-03 |
| **A-05** | `ReposoDark` | ídem, `globals:{theme:"dark"}` | mismo assert estructural | RF-353 |
| **A-06** | `ConNodoEnAlcance` | `scope: {clase:"caja", nodeId:"hipaa-check", fuentePath:"/hipaa-check"}` | **3** filas ∧ `getByText("hipaa-check")` ∧ `queryByText("Alcance:") === null` ∧ el `✕` con `title="quitar del alcance"` | RF-329 |
| **A-07** | `DetalleDesplegado` | `play` abre el chip | `getByText("~/Proyectos/luana-vitalia/vitalia")` ∧ `getByText(/4b046945/)` ∧ `aria-expanded="true"` | RF-328 · RF-330 |
| **A-08** | `ListaAbierta` | `abierta: true`, 4 conversaciones | `getByRole("listbox")` ∧ 4 `option` ∧ el composer **sigue visible** (`getByPlaceholderText(/Pídele un cambio/)`) | RF-317 |
| **A-09** | `ConversacionRecienCreada` | activa 0 turnos, `desactivada:"por qué el mapa sale vacío"` | el copy nuevo presente ∧ `queryByText(/Esta conversación ES la sesión/) === null` ∧ el chip dice `0%` | RF-307 · BR-CV-9 |
| **A-10** | `PrimeraConversacionDeLaSesion` | 0 turnos, sin desactivada | el copy **sin** la 2ª oración | §4.6 #2 |
| **A-11** | `Retomando` | `retomando:"d303a93f"`, `status:"streaming"` | `getByRole("status")` con «Retomando la conversación…» ∧ `getByText("--resume d303a93f")` ∧ el detalle **abierto solo** | RF-310 · RF-314 · E-10 |
| **A-12** | `RetomandoSinResume` | conversación de 0 turnos, sin `claude_session_id` | la franja **NO** se dibuja | E-16 |
| **A-13** | `RetomaFallida` | `retomaFallo:"la sesión de Claude Code ya no existe — reinicié el hilo"` | `getByRole("alert")` con el motivo ∧ la franja en variante `bad` | RF-348 · E-13 · C-11 |
| **A-14** | `TranscriptConRotacion` | `conv` con un `RolSys` `"— contexto rotado, seguimos —"` entre burbujas | el texto exacto presente **centrado** ∧ el chip bajó a 12 % ∧ el detalle abierto | RF-313 · C-5 |
| **A-15** | `TurnoEnVuelo` | `status:"streaming"` | `＋` con `disabled` ∧ `title` con «esperá a que termine el turno» ∧ el `🔍` **habilitado** | RF-312 · E-07 · E-28 |
| **A-16** | `PermisoPendiente` | `status:"await"` + 1 `pendingPerm` | `＋` `disabled` con motivo de permiso ∧ la `PermissionCard` sigue renderizando | RF-312 · E-08 |
| **A-17** | `ListaAbiertaDark` | A-08 en oscuro | mismo assert | RF-353 |

### 2.2 `web/src/widgets/chat-dock/ui/conversacion-row.stories.tsx` — **NUEVO** (11)

Args base: `activa`, `arnes:"vitalia"`, `status:"idle"`, `listaAbierta:false`,
`detalleForzado:false`, `onToggleLista:fn()`, `onNueva:fn()`, `onRenombrar:fn()`.

| id | story | escenario | `play()` asserta | cubre |
|---|---|---|---|---|
| **B-01** | `Reposo` | — | los 4 controles presentes con sus nombres accesibles | §4.1 #1 |
| **B-02** | `ListaAbierta` | `listaAbierta:true` | `aria-expanded="true"` en el botón del título | §4.1 #4 |
| **B-03** | `TituloLargo` | 120 chars | el `title` nativo trae el completo | §4.1 #7 |
| **B-04** | `AbrirPorChevron` | — | click en el título ⇒ `onToggleLista` llamado con `"filas"` | C-10 · RF-355 |
| **B-05** | `AbrirPorLupa` | — | click en 🔍 ⇒ `onToggleLista` llamado con `"buscador"` | C-10 · RF-355 |
| **B-06** | `CrearNueva` | — | click en ＋ ⇒ `onNueva` llamado 1 vez | RF-307 · E-06 |
| **B-07** | `Bloqueada` | `status:"streaming"` | `＋` `disabled` ∧ `title` con el motivo ∧ `onNueva` **nunca** llamado tras el click | RF-312 · E-07 |
| **B-08** | `RenombrarConfirma` | `play`: ✎ → limpiar → «el bug del índice» → `Enter` | `onRenombrar` llamado con `"el bug del índice"` ∧ el ctx y las acciones **no están** durante la edición | RF-331 · §4.1 #5 |
| **B-09** | `RenombrarVacio` | `play`: ✎ → borrar todo → `Enter` | `onRenombrar` **nunca** llamado ∧ el título vuelve al anterior | E-30 |
| **B-10** | `RenombrarEscape` | `play`: ✎ → tipear → `Escape` | `onRenombrar` **nunca** llamado ∧ sale de edición ∧ el foco vuelve al botón del título | E-32 |
| **B-11** | `RenombrarBlur` | `play`: ✎ → tipear → `Tab` | `onRenombrar` llamado (blur confirma, como el rail) | E-34 |

> **E-31 (título duplicado)** y **E-33 (renombrar con turno llegando)** no tienen story propia: el
> componente no puede distinguirlos (no conoce los otros títulos ni el stream). Van a §4.

### 2.3 `web/src/widgets/chat-dock/ui/ctx-chip.stories.tsx` — **NUEVO** (10)

| id | story | escenario | `play()` asserta | cubre |
|---|---|---|---|---|
| **C-01** | `Frio` | `ctxPct:68, caliente:false, abierto:false` | nombre accesible contiene `68` ∧ la barra `aria-hidden` | §4.2 #1 · RF-356 |
| **C-02** | `Caliente` | `caliente:true` | **el número NO usa `--warn`** (assert de estilo computado: `color` ≠ el valor de `--warn`) ∧ la barra **sí** | **C-3** · RF-353 |
| **C-03** | `Cero` | `ctxPct:0` | el chip se renderiza y dice `0%` — no se esconde | BR-CV-9 · E-04 |
| **C-04** | `Abierto` | `abierto:true` | `aria-expanded="true"` ∧ el detalle en el DOM | RF-330 |
| **C-05** | `Toggle` | `play`: click, click | `onToggle` llamado 2 veces | RF-330 |
| **C-06** | `DetalleCompleto` | cc-id + arnés + modelo + cwd | los **cuatro** presentes | RF-328 |
| **C-07** | `DetalleSinSesionCC` | `claudeSessionId: undefined` | `getByText("◍ sin sesión CC")` — literal vigente | E-05 · §4.2 detalle #2 |
| **C-08** | `DetalleSinCwd` | `cwd: undefined` | la línea del cwd **no existe** (no hay `—` de relleno) | §4.2 detalle #3 |
| **C-09** | `DetalleCwdLargo` | cwd de 90 chars | truncado ∧ `title` completo | §4.2 detalle #4 |
| **C-10** | `CalienteDark` | C-02 en oscuro | mismo assert de color | RF-353 |

### 2.4 `web/src/widgets/chat-dock/ui/conversaciones-panel.stories.tsx` — **NUEVO** (14)

| id | story | escenario | `play()` asserta | cubre |
|---|---|---|---|---|
| **D-01** | `Cargando` | `estado:"cargando"` | `getByRole("status")` con `aria-label="Cargando conversaciones"` ∧ **cero** `option` | §4.3 #1 · E-35 |
| **D-02** | `Error` | `estado:"error", error:"daemon no responde"` | `getByRole("alert")` con el motivo ∧ botón `Reintentar` ∧ **`queryByText(/0 conversaciones/) === null`** | RF-347 · BR-CV-10 · E-35 |
| **D-03** | `ErrorReintentar` | ídem | click ⇒ `onReintentar` llamado 1 vez | RF-347 |
| **D-04** | `Cuatro` | 4 conversaciones | rótulo `4 conversaciones de la sesión «repro del bug de carga»` ∧ 4 `option` ∧ buscador presente | RF-318 · E-03 |
| **D-05** | `UnaSola` | 1 conversación | **`queryByRole("searchbox") === null`** ∧ el mensaje «Esta sesión recién arranca…» | RF-325 · E-02 · E-21 |
| **D-06** | `BuscandoConCoincidencias` | `busqueda:"manifiesto"`, 2 de 4, con `fragmento` | rótulo `2 de 4 coinciden` ∧ 2 `option` ∧ 2 `<mark>` con el término | RF-321 · RF-322 |
| **D-07** | `BuscandoConLaActiva` | la activa entre los 2 resultados | la fila activa conserva `aria-selected="true"` y el rótulo `activa` | RF-323 · E-24 |
| **D-08** | `SinCoincidencias` | `busqueda:"telemetría"`, 0 resultados | el copy con **el término entre comillas** ∧ «Se buscó en el título y en el texto de las 4.» ∧ `Limpiar búsqueda` ∧ el pie `Cancelar` sigue | RF-325 · E-22 |
| **D-09** | `LimpiarBusqueda` | ídem | click ⇒ `onBusqueda("")` | E-22 |
| **D-10** | `Bloqueado` | `bloqueadoMotivo:"esperá tu decisión de permiso"` | todas las `option` con `aria-disabled="true"` ∧ el `searchbox` **habilitado** ∧ click en una fila ⇒ `onElegir` **nunca** llamado | RF-312 · E-11 · E-28 |
| **D-11** | `ElegirInactiva` | — | click en la 2ª ⇒ `onElegir` con **su id exacto** | RF-310 · E-10 |
| **D-12** | `ClickEnLaActiva` | — | click en la activa ⇒ `onElegir` **con el id de la activa** (el no-op lo resuelve el store, §3) | E-15 |
| **D-13** | `TecladoFlechas` | `play`: `↓ ↓ Enter` | `onElegir` con el id de la **tercera** fila ∧ `aria-activedescendant` fue cambiando | RF-354 |
| **D-14** | `EscapeCierra` | `play`: `Escape` | `onCancelar` llamado 1 vez | RF-317 · RF-354 |

### 2.5 `web/src/widgets/chat-dock/ui/conversacion-fila.stories.tsx` — **NUEVO** (9)

| id | story | escenario | `play()` asserta | cubre |
|---|---|---|---|---|
| **E-01s** | `Inactiva` | 41 turnos, ctx 34 %, `ayer 18:02` | los 4 datos ∧ **sin** rótulo `activa` | RF-319 · RF-302 |
| **E-02s** | `Activa` | — | rótulo `activa` ∧ `aria-selected="true"` ∧ el radio relleno presente (no sólo el color) | RF-302 · RF-356 |
| **E-03s** | `SinTurnos` | `turnos:0, ultima_interaccion:""` | `getByText(/sin turnos todavía/)` ∧ `ctx 0 %` ∧ **ninguna fecha inventada** | BR-CV-14 · E-04 |
| **E-04s** | `SinFecha` | `turnos:41, ultima_interaccion:""` (migrada) | `getByText(/sin fecha/)` ∧ los 41 turnos | H-4 · §4.4 #6 |
| **E-05s** | `ConFragmento` | `fragmento:"…el manifiesto declara 0 elementos…"`, `termino:"manifiesto"` | el `<mark>` envuelve **exactamente** el término | RF-322 |
| **E-06s** | `ConFragmentoAcentuado` | `termino:"vacio"`, fragmento con «vacío» | el `<mark>` resalta «vacío» — el resaltado no depende de igualdad literal | E-23 |
| **E-07s** | `SinFragmento` | coincidencia sólo en el título | **cero** `<mark>` y cero tercera línea | RF-322 · E-27 |
| **E-08s** | `Deshabilitada` | `deshabilitadaMotivo:"…"` | `aria-disabled="true"` ∧ `title` con el motivo ∧ `onElegir` **nunca** tras el click | C-1 · RF-354 |
| **E-09s** | `ActivaDark` | E-02s en oscuro | mismo assert | RF-353 |

### 2.6 `web/src/widgets/session-rail/ui/new-session-picker.stories.tsx` — **SUPERSET** (−4 asserts)

| id | story | qué cambia | cubre |
|---|---|---|---|
| **F-01** | las stories que hoy renderizan `ConversacionesDelArnes` | **pierden** el bloque y sus asserts; ganan uno nuevo: `queryByText(/Conversaciones:/) === null` | RF-333 |

> **Efecto colateral verificado:** las **4 stories que hoy fallan el gate a11y** por el `text-warn`
> de `new-session-picker.tsx:273` (`BACKLOG.md:207-211`) dejan de fallar **por desaparición de la
> causa**. Eso hay que anotarlo en `PARIDAD.md`: la deuda del token `--warn` **sigue abierta**, sólo
> que este consumidor ya no la ejercita.

### 2.7 `web/src/widgets/chat-dock/model/conversaciones-store.test.ts` — **NUEVO** (project `unit`)

No es story: es transporte puro, y el project `unit` corre en Node sin Chromium
(`vitest.config.ts:20-27`). Precedente: `conversaciones-store.test.ts` del rail (que **se elimina**)
y `portafolio-picker-store.test.ts`.

| id | test | asserta | cubre |
|---|---|---|---|
| **U-01** | `abrir refetchea siempre` | dos `abrir()` seguidos ⇒ 2 llamadas al fetch; jamás lista cacheada | RF-334 |
| **U-02** | `cerrar descarta la búsqueda` | tras `buscar("x")` + `cerrar()`, `busqueda === ""` ∧ `estado` inicial | E-42 |
| **U-03** | `cambiar de sesión cierra la lista` | `abrir("s1")` + `abrir("s2")` ⇒ la lista no queda mostrando `s1` | E-42 |
| **U-04** | `fallo al listar guarda el motivo tal cual` | `estado === "error"` ∧ `error` = el mensaje del `ApiError`, sin reescribir ∧ `conversaciones === []` **y `estado !== "datos"`** | E-35 · BR-CV-10 |
| **U-05** | `fallo al crear no altera el estado` | tras un `crear()` que rechaza, la lista y la activa son las de antes | E-37 |
| **U-06** | `fallo al retomar no altera el estado` | ídem con `activar()` | E-36 |
| **U-07** | `409 se distingue del resto` | `ApiError{status:409}` ⇒ el motivo dice «turno en vuelo», no el genérico | E-39 |
| **U-08** | `timeout se reporta como fallo de transporte` | un reject de `fetch` ⇒ mismo camino que U-04, con su motivo | E-38 |
| **U-09** | `el frame de conversación es idempotente por id+evento` | aplicar dos veces el mismo `{conversacion_id, conversacion_evento}` no duplica nada | E-41 · §7.2 |
| **U-10** | `renombrar a vacío no llama a la API` | `renombrar(cid, "   ")` ⇒ 0 llamadas | E-30 |

---

## 3 · Cobertura escenario → story

| escenario | cubierto por | mecanismo |
|---|---|---|
| E-01 sesión sin conversaciones | — | **Go** (invariante de dominio) |
| E-02 una sola | **D-05** | story |
| E-03 muchas | **A-04 · D-04** | story |
| E-04 una de 0 turnos | **C-03 · E-03s** | story |
| E-05 sin `claude_session_id` | **C-07** | story |
| E-06 crear quieta | **B-06** | story |
| E-07 crear con streaming | **A-15 · B-07** | story |
| E-08 crear con `await` | **A-16** | story |
| E-09 crear dos veces | — | **Go** |
| E-10 retomar quieta | **A-11 · D-11** | story |
| E-11 retomar con turno en vuelo | **D-10** | story |
| E-12 proceso `claude` inexistente | — | **Go** |
| E-13 `--resume` falla | **A-13** (la superficie) | story + **Go** (`TestResumeAutoSana`, ya existe) |
| E-14 `cwd` inexistente | — | **Go** + **E2E vivo** |
| E-15 retomar la activa | **D-12** | story + **U** |
| E-16 retomar de 0 turnos | **A-12** | story |
| E-17 rotación en medio de turno | — | **Go** (`session_rotacion_test.go`, ya existe) |
| E-18 rotación con dock colapsado | — | **U** + **E2E vivo** |
| E-19 rotación con lista abierta | **A-14** | story |
| E-20 rotación en inactiva | — | **Go** (imposible por invariante) |
| E-21 buscar con una sola | **D-05** | story |
| E-22 sin coincidencias | **D-08 · D-09** | story |
| E-23 acentos y mayúsculas | **E-06s** (el resaltado) | story + **Go** (la normalización, RF-341) |
| E-24 la activa entre resultados | **D-07** | story |
| E-25 transcript enorme | — | **Go** (test con el fixture real de 12 281 B) |
| E-26 query en blanco | — | **Go** |
| E-27 coincide sólo en el título | **E-07s** | story |
| E-28 buscar con turno en vuelo | **A-15 · D-10** | story |
| E-29 muchas coincidencias | — | **Go** (un solo fragmento) |
| E-30 renombrar a vacío | **B-09** | story + **U-10** |
| E-31 título duplicado | — | **Go** (no hay unicidad que validar) |
| E-32 Escape | **B-10** | story |
| E-33 renombrar con turno llegando | — | **E2E vivo** (§4.2) |
| E-34 blur | **B-11** | story |
| E-35 daemon caído al listar | **D-01 · D-02** | story + **U-04** |
| E-36 daemon caído al retomar | — | **U-06** |
| E-37 daemon caído al crear | — | **U-05** |
| E-38 timeout | — | **U-08** |
| E-39 respuesta 409 | — | **U-07** |
| E-40 dos ventanas | — | **E2E vivo** + **U-09** |
| E-41 SSE reconecta | — | **U-09** |
| E-42 cambiar de sesión con lista abierta | — | **U-02 · U-03** |
| E-43 migración | — | **Go** |
| E-44 archivos corruptos | — | **Go** |
| E-45 sin permisos de escritura | — | **Go** |
| E-46 **borrado CV-D6** | — | **manual, con rastro en `PARIDAD.md`** (§4.3) |
| E-47 crash a mitad de la migración | — | **Go** |
| E-48 nodo desaparece tras reindex | — | **U** + **E2E vivo** |
| E-49 arnés desaparece del Portafolio | — | **E2E vivo** |
| E-50 cerrar la sesión entera | — | **Go** |

**26 por story · 10 por unit · 15 por Go · 5 por E2E vivo · 1 manual.** (Varios escenarios llevan
más de un mecanismo; el total de filas es 50.)

---

## 4 · Lo que NO cubre una story, y con qué se cubre

### 4.1 Go — 15 escenarios

Van a `internal/usecase/session_conversaciones_test.go` (nuevo) y a los existentes
`session_rotacion_test.go` / `session_historial_test.go`. Ninguno es de UI: son invariantes de
dominio, migración de disco y el matcher de búsqueda.

**Fixture obligatoria, no inventada:** la conversación real de 90 turnos / 12 281 B que hoy vive en
`~/.arnesia/sessions.json` (`s6165ac75`) se copia a `testdata/` para E-25. Prohibido generar un
transcript sintético cuando hay uno real (`plan-pruebas` del paquete de marketplace, §0).

### 4.2 E2E vivo — 5 escenarios

Contra el binario instalado, no contra `pnpm dev`. Precedente y motivo: el paquete de dictado
descubrió 4 bugs que los fakes no veían (`checkpoint.md`).

| escenario | cómo |
|---|---|
| E-14 `cwd` inexistente | `mv` de la carpeta del arnés → mandar un turno → ver el motivo con la ruta |
| E-18 rotación con dock colapsado | `-rotacion-umbral 5` → turnos hasta rotar con `⌘K` cerrado → reabrir y ver la marca |
| E-33 renombrar con turno llegando | mandar un turno largo y renombrar mientras stremea |
| E-40 dos ventanas | app instalada + `http://127.0.0.1:4200` en el navegador, misma sesión |
| E-49 arnés desaparece | desvincular desde la vista Portafolio con la sesión abierta |

### 4.3 E-46 — el único sin test, y a propósito

El borrado de las 3 conversaciones cerradas (CV-D6, RF-337) **no se automatiza**: es dato del
operador, el procedimiento exige el daemon detenido y una copia previa, y un test que lo ejercite
tendría que borrar el archivo real o fingir uno — las dos opciones son peores que un rastro escrito.

**Lo que sí se exige:** una entrada en `PARIDAD.md` con fecha, los 3 ids (`s1b38a066`, `s408bb085`,
`s020210e3`), la ruta del `.bak` y el resultado del `GET` de verificación.

---

## 5 · Estado del tooling

### 5.1 Lo que hay (verificado hoy)

| pieza | versión | dónde |
|---|---|---|
| `storybook` | `^10.4.6` | `package.json:60` |
| `@storybook/react-vite` | `^10.4.6` | `:48` |
| `@storybook/addon-a11y` | `^10.4.6` | `:46` — **sí existe**, y `preview.ts:7` lo pone en `error` |
| `@storybook/addon-vitest` | `^10.4.6` | `:47` |
| `vitest` | `^4.1.9` | `:67` |
| `@vitest/browser` · `@vitest/browser-playwright` | `^4.1.9` · `4.1.9` (pin) | `:55-56` |
| `playwright` | `^1.61.1` | `:58` |
| browser mode | Chromium, **`headless: true`** | `vitest.config.ts:35-38` |

### 5.2 Lo que NO hay — declarado, no inventado

| falta | consecuencia |
|---|---|
| `vitest-browser-react` | no se usa; el render lo hace `storybookTest` |
| `@storybook/test` | el subpath correcto es `storybook/test` (42/42 archivos) |
| `@testing-library/*`, `jsdom`, `happy-dom` | no hay entorno DOM alternativo; todo corre en Chromium real |
| script `test:storybook` | sólo `test` (= los dos projects). Correr uno solo es con `--project=` |

### 5.3 Huecos del tooling — 3

| id | hueco | evidencia | impacto en este paquete |
|---|---|---|---|
| **T-1** | **No hay regresión visual por screenshot.** `boundaries/fe-visual-fitness.md` dice que las baselines van por `storybook-addon-vis`, pero **no está instalado** (0 ocurrencias en `package.json` y en `.storybook/`) | `package.json` | «story = test» significa hoy **render + `play` + axe**, no diff de píxeles. Todos los asserts de este plan son **estructurales o de rol**, nunca «se ve igual». El único assert de estilo es **C-02** (el color del número caliente), y es puntual y justificado |
| **T-2** | **`npm run verify` no corre las stories** | `package.json:25` | el constructor tiene que correr `--project=storybook` a mano; CI lo obliga (`ci.yml:67-68`) |
| **T-3** | **Cero stories de `pages/`** | 42 stories, ninguna en `pages/` | `shell-page.tsx` (la composición del dock en el `<aside>`) **no está cubierta**. Este paquete no lo cambia (`design.md` §1.3), así que el hueco no crece — pero E-42 (cambiar de sesión con la lista abierta) depende de esa composición y por eso va a `unit`, no a story |

### 5.4 Nota derogatoria confirmada

La nota vieja del repo «vitest-browser no corre en background / Chromium no headless» está
**derogada**: `vitest.config.ts:37` tiene `headless: true` explícito y el CI lo corre sin display
(`ci.yml:68`). La memoria del proyecto ya lo registra como desmentido (2026-07-26). Este plan **no
lo re-litiga**: las 56 stories se corren headless.

---

## 6 · Definición de hecho

Un tramo de `design.md` §10 está cerrado cuando:

1. Sus stories existen y **pasan en verde** con
   `pnpm --dir web exec vitest --project=storybook run`.
2. **Ninguna** bajó `a11y` a `"todo"` (RF-353 CA-4) — verificable con
   `grep -n 'a11y' web/src/widgets/chat-dock/ui/*.stories.tsx`.
3. Los estados de `design.md` §4 tienen al menos una story cada uno: **8 + 6 + 4 + 8 + 8 + 4 + 4 =
   42 estados**, cubiertos por las 56 stories (varias cubren dos).
4. `pnpm --dir web run verify` verde (typecheck · biome · depcruise · steiger · stylelint).
5. `pnpm --dir web exec vitest --project=unit run` verde, incluyendo los 10 tests de §2.7.
6. `go test ./...` verde con los tests de §4.1.
7. La fila de `mockups/INDEX.md` de este paquete se actualiza (DoD, regla dura 5).
