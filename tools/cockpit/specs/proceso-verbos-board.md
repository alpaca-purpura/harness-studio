# SPEC — Verbos CDEvents + board gobernado por el descriptor (F6) — CONGELADA

> **Congelada 2026-07-03** (épica Torre de Control, F6 — la fase de cierre). Continúa la
> numeración de la épica (RN-01..RN-44 en `torre-read-only.md`, `torre-arquitectura.md`,
> `proceso-descriptor.md`, `delivery-cockpit.md`). Decisiones D15-D17: checkpoint expirado
> (60s, AFK) → recomendadas aplicadas (mecánica de la casa). La implementación ejecuta ESTO
> sin retocar diseño; cambio de diseño = nueva ficha DH-NN + edición explícita aquí.

## 0 · Qué es

El principio 5 del norte (CDEvents: "qué acaba de pasar", no solo "en qué estado está")
hecho dato, y la muerte del último hardcode del ciclo en la UI: el campo `verbo` reservado
en F4 se llena e interpreta, y las definiciones duplicadas del board (D10/D13) migran a
`/api/proceso` — el board queda GOBERNADO por el descriptor. Con esto las tres capas del
norte quedan cerradas de punta a punta y la épica puede cerrar.

## 1 · Vocabulario de verbos (D15: solo verbos como dato, sin log)

- **RN-45** · `transicion.verbo` pasa de RESERVADO a **INTERPRETADO**: formato
  `<sujeto>.<predicado>` (regex `^[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$`), validado por el gen
  cuando está presente (sigue opcional en el contrato — otra instancia puede no nombrarlos).
  El verbo nombra el **EVENTO**, no la transición: NO es único (dos transiciones pueden
  emitir el mismo evento — `story.parked` desde `idea` y desde `refining`).
- **RN-46** · Instancia `sdd-default`: vocabulario chico estilo CDEvents — sujetos
  `{story, spec, build, review}` + predicados en pasado:

  | transición | verbo |
  |---|---|
  | idea→refining | `spec.started` |
  | refining→refined | `story.refined` |
  | refined→ready | `story.queued` |
  | ready→developing | `build.started` |
  | developing→developed | `build.finished` |
  | developed→reviewing | `review.started` |
  | reviewing→done | `story.merged` |
  | idea→parked · refining→parked | `story.parked` |
  | idea→dropped · refining→dropped | `story.dropped` |
  | refining→idea | `story.backlogged` |
  | parked→idea | `story.reactivated` |

- **RN-47** · El verbo viaja como DATO, sin log persistente (D15): `/api/proceso` ya lo
  serializa; la respuesta de `POST /api/transition` gana `verbo` (el evento nombrado que
  acaba de ocurrir); `GET /api/delivery/plantilla` gana `tramo.verbo`. Timeline/endpoint
  de eventos persistido = ficha DH-NN post-épica (ES la costura con P4 del norte —
  frontera #1: dato PUBLICADO por contrato, jamás lectura cruzada).

## 2 · Campos aditivos del contrato (para que las defs de la UI tengan fuente)

- **RN-48** · El schema L0 gana, ADITIVOS y opcionales:
  - `transicion.nombre` (str) — etiqueta humana de la ACCIÓN ("Empezar a refinar"): la
    fuente de los labels de botón que hoy hardcodea `OPERATOR_ALLOWED_TRANSITIONS.verb`.
  - `estado.wip` (int > 0) — límite WIP del estado; ausente = sin límite: la fuente de
    `WIP_CAPS`.
  El gen valida ambos (nombre = str no vacía; wip = int > 0). Instancia: `nombre` en las
  7 transiciones de operador (las strings actuales), `wip` en los 6 estados con cap hoy
  (refining 3 · refined 5 · ready 5 · developing 3 · developed 1 · reviewing 1).
  Cambio de instancia del kit → entry en su CHANGELOG (KIT-08; la versión la decide P3).
- **RN-49** · Motor genérico de razones: `handleTransition` deja de conocer
  `parked`/`dropped` — una transición con `requiere_razon` escribe
  `{targetState}_reason` en el frontmatter (byte-igual para sdd-default; otro descriptor
  gana el mismo comportamiento gratis).

## 3 · Migración de las defs de la UI (D16: completa)

- **RN-50** · **ProcesoProvider**: un fetch a `/api/proceso` al montar el shell sistema;
  helpers PUROS en `lib/proceso.ts` (testeados con fixture): orden canónico ·
  happy-path (categoría ∉ {pausado, descartado}) · terminal por categoría · wip ·
  descripción · transiciones de operador (con nombre/verbo) · `isOperatorAllowed` ·
  dueño render (join `"{arnes} ({nota})"` con `" o "` — byte-igual al Go/RN-31) ·
  `editOrder` (índice en `estados[]`; categoría `pausado` → índice del inicial;
  `descartado` → ∞ — reproduce el STATE_ORDER de edit-permissions).
- **RN-51** · Las defs migran así (muere el literal, la vista no cambia con sdd-default):

  | def hardcodeada | fuente en el descriptor |
  |---|---|
  | `STATES_ORDER` (BoardView) | orden de `estados[]` (+ filtro parked/dropped POR CATEGORÍA) |
  | `WIP_CAPS` + chips de cabecera (BoardView) | `estado.wip` |
  | `OPERATOR_ALLOWED_TRANSITIONS` + `isOperatorAllowed` (types.ts) | transiciones `ejecutor: operador` (labels = `nombre`) |
  | `STATE_TOOLTIPS` (BoardColumn) | `estado.descripcion` (los `TOOLTIPS.state_*` mueren del diccionario) |
  | `STATE_CLASSES` (Badge) | theme de PRESENTACIÓN por id nativo (paleta actual) + **fallback por categoría** (contrato) — mismo look hoy, otro descriptor no rompe |
  | `STATE_OWNERS` (OperatorTransitions) | `duenos` render |
  | `STATE_OWNERS` + whitelist (dev-route `app/api/transition`) | lee el MISMO espejo server-side (lib compartida con `app/api/proceso/route.ts`) |
  | `STATE_ORDER` (edit-permissions) | `editOrder` del provider — `getEditPermission` recibe el orden derivado; `LOCK_FROM_STATE` **queda** (política de ENTIDAD por artefacto, fuera del descriptor — F4 §6) |

- **RN-52** · Hallazgo F6 — literales del ciclo FUERA de las 8 defs mapeadas, en
  `ProcesoTab` (story-drawer): `MAIN_FLOW` → happy-path del provider · `STEP_META`
  (who/what) → `duenos` render + `descripcion` · `STEP_CURRENT_CLS` → theme por id +
  fallback categoría · el literal `developed` del form gate G → el gate con
  `autoridad.rol: operador` y momento-ESTADO. Los paneles `chris_verify`/`reconciled` y
  el chip G·R quedan (narrativa de entidad del SDD v5, no topología del proceso).
  El header del board ("10 estados v4") deriva: `{n} estados · {descriptor.nombre}`.

## 4 · Lanzar sesión desde el drawer (D17: entra)

- **RN-53** · Botón «Lanzar sesión de delivery» en la card de acciones de `ProcesoTab`
  → navega a `/delivery?story=<id>`; `DeliveryView` preselecciona la story por query
  param (la plantilla se re-deriva sola — RN-39 intacto). Cierra el fork que D13 dejó
  para esta fase.

## 5 · Tests (regla de la fase)

- **RN-54** · Los tests que fijaban hardcodes migrados se **REEMPLAZAN** por tests del
  contrato: `lib/__tests__/proceso.test.ts` (fixture espejo de sdd-default + un alterno
  con estados inventados → derivaciones correctas); `edit-permissions.test.ts` construye
  el orden desde fixture (la matriz artifact × state se conserva). Go: verbos
  (parse + formato + lookup), respuesta de transition con `verbo`, `tramo.verbo` en
  plantilla. Suite Go + vitest verdes SIN tocar tests viejos no relacionados.

## 6 · Fuera de alcance

Log/timeline persistente de eventos de delivery (ficha post-épica — la costura P4) ·
enforcement de transiciones de rol (los skills siguen ejecutando) · editor del descriptor ·
`?sistema=` en `/api/proceso` (un binario = un descriptor embebido, F5+ si hace falta).
