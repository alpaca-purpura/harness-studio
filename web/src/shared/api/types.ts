// Contract types mirroring the daemon's domain (internal/domain/session.go) and the
// Dock SSE frame (internal/usecase/session_service.go dockFrame). These are the shared
// vocabulary the whole SPA speaks; keep them in lockstep with the Go side.

export type SessionStatus = "streaming" | "await" | "idle"
export type Salud = "ok" | "warn" | "crit" | "info"
// "act" es un paso de actividad del turno (CH-D2/D3): texto "<tool> <blanco>"; el dock
// agrupa consecutivos en una tarjeta desplegable, jamás como burbuja.
export type Rol = "user" | "assistant" | "sys" | "act"

export interface Turn {
  rol: Rol
  text: string
}

// Conversacion es UNA conversación de una sesión, tal como viaja en la lista del panel
// (`ConversacionResumen` del daemon, session_conversaciones.go:46-59). NO trae los turnos:
// el tipo ni siquiera los declara, así que «no cargado» no es expresable acá.
//
// Los campos que el daemon serializa SIN `omitempty` van acá sin `?`: `ctx_pct` a 0 y
// `turnos` a 0 son DATOS (BR-CV-9), no ausencias, y un `?` invitaría al `?? 0` que borra la
// distinción. `ultima_interaccion` sí es opcional: viaja vacía cuando no hubo ningún turno,
// y el FE la DICE («sin fecha»), no la inventa (BR-CV-14).
export interface Conversacion {
  id: string
  /** título auto-derivado del primer turno `user`, o "nueva conversación" (RF-303). */
  titulo: string
  /** true ⇒ ningún turno vuelve a re-derivar el título (RF-303). */
  titulo_editado: boolean
  /** exactamente una por sesión (BR-CV-1). */
  activa: boolean
  turnos: number
  /** 0 es dato, no ausencia (BR-CV-9). */
  ctx_pct: number
  /** el `caliente` del chip sale de acá: el umbral vive en el daemon, el FE no lo conoce. */
  rotacion_pendiente: boolean
  /** RFC3339 UTC. VACÍA cuando no hubo turnos — no se inventa (BR-CV-14). */
  ultima_interaccion?: string | undefined
  creada_en: string
  claude_session_id?: string | undefined
  model?: string | undefined
  /** sólo con `?q=`: el fragmento que coincidió, ya recortado por el daemon (RF-322). */
  fragmento?: string | undefined
}

// ConversacionActiva es la conversación viva de una sesión: el resumen MÁS su transcript.
// Es la única forma que trae `conv`, y viaja en dos lugares — dentro de la sesión y dentro
// del frame de la transición, para que el FE repinte sin una segunda vuelta.
export interface ConversacionActiva extends Conversacion {
  /** SIN `?`: `[]` y «no la cargué» no pueden significar lo mismo. */
  conv: Turn[]
  /** los ids de Claude Code previos de este hilo (sus rotaciones). */
  cadena_cc?: string[] | undefined
}

// ConversacionesListado es la respuesta de GET /api/sessions/{id}/conversaciones.
export interface ConversacionesListado {
  conversaciones: Conversacion[]
  /** el total de la SESIÓN, no el de coincidencias: es el denominador de «N de M» (RF-318). */
  total: number
}

// Session es el frente de trabajo. Bajo CV-D3 la sesión CONTIENE N conversaciones, así que
// `claude_session_id`, `model`, `ctx_pct`, `conv`, `turnos` y `cadena_cc` ya NO viven acá:
// bajaron a `activa`. El wire es `sessionWire` (sessions.go:26-38).
//
// `activa` NO es opcional, y es deliberado: la invariante garantiza que existe (BR-CV-1) y un
// campo opcional invitaría al `if (!activa)` defensivo que escondería el día en que no exista.
// El daemon loguea `error` antes que inventar una.
export interface Session {
  id: string
  frente: string
  arnes: string
  empresa?: string
  puesto?: string
  salud?: Salud
  status: SessionStatus
  view: string
  parked?: string
  // Sesión de reparación (RF-191, ley A4): abierta contra una instalación del Portafolio.
  reparacion?: boolean
  /** el directorio de trabajo del conductor — el confinamiento real (RF-328 CA-2). */
  cwd?: string | undefined
  // Metadata de archivo (RF-200) — solo poblada en sesiones CERRADAS.
  cerrada_en?: string
  activa: ConversacionActiva
}

// NewSession is the create payload. path, when set, registers the arnés's working directory
// (the dir its conductor runs claude in) in the same call — per-session confinement (S2).
export interface NewSession {
  arnes: string
  frente?: string
  empresa?: string
  puesto?: string
  salud?: Salud
  view?: string
  parked?: string
  path?: string
  reparacion?: boolean
}

// HarnessSummary is one entry of GET /api/harnesses (S1 portfolio / the Map picker, RF-72):
// a lightweight row, not the full graph. The domain Graph/Box types live in entities/arnes
// (shared/api must not import upward), so the page maps this to whatever it renders.
export interface HarnessSummary {
  // La CLAVE del índice del daemon (`sin-home~vitalia~vitalia`) — el mismo espacio de llaves
  // que `session.arnes` y que `getGraph(id)`. NO es el `arnes.id` del manifiesto: comparar
  // contra ése hacía que una sesión cargable se viera «no está en el índice del daemon».
  id: string
  rol?: string
  proceso?: string
  empresas?: string[]
}

// DockFrame is one SSE `dock` event payload. Every frame carries session_id so one
// connection multiplexes N conversations (fase 4 c.1). kind=permission es la tarjeta
// ask→UI de un control_request (RF-113: request_id + tool + input crudo para pintar el
// diff; la sesión pasa a `await`); permission_result la cierra con la decisión efectiva.
// kind=conversacion es el frame del panel (CV-D2): lo emiten crear, retomar, renombrar y la
// rotación. NO lleva `run_id` — no pertenece a un turno — así que el dedup por run finalizado
// no aplica: la idempotencia es declarativa para creada/activada/renombrada (traen el estado
// final: aplicarlas dos veces es un `set`) y por `turno_idx` para `rotada`.
export interface DockFrame {
  session_id: string
  run_id?: string
  kind:
    | "status"
    | "init"
    | "delta"
    | "message"
    | "act"
    | "result"
    | "error"
    | "permission"
    | "permission_result"
    | "conversacion"
  text?: string
  status?: SessionStatus
  ctx_pct?: number
  model?: string
  claude_session_id?: string
  request_id?: string
  tool?: string
  input?: unknown
  decision?: "allow" | "deny"
  // Los 4 campos del frame `conversacion` (design.md §7.2).
  conversacion_id?: string
  conversacion_evento?: "creada" | "activada" | "renombrada" | "rotada"
  /** sólo en `rotada`: el índice del turno donde va la marca — su llave de idempotencia. */
  turno_idx?: number
  /** el estado POST-transición, para repintar sin una segunda vuelta. */
  conversacion?: ConversacionActiva
}

// MapFrame is one SSE `map` event payload (RF-186/RF-187): the daemon reindexed an arnés
// after a chat turn; a mounted Map viewing that harness refetches its graph live.
export interface MapFrame {
  harness_id: string
  degradado?: boolean
}

// PermissionAsk is one pending control_request card of a session (RF-113).
export interface PermissionAsk {
  request_id: string
  tool: string
  input?: unknown
}

// ScopeNode is the removable composer scope chip (RF-111, decisión #1): a Map-selected
// node resolved to its real file by the loader's nomenclatura (fuente_path). Los campos
// admiten undefined explícito (se construyen por spread desde el nodo del grafo —
// exactOptionalPropertyTypes).
export interface ScopeNode {
  nodeId: string
  clase?: string | undefined
  fuentePath?: string | undefined
}

// GateCheckResult / GateReport mirror the daemon's ConformanceReport (RF-117 — solo lo
// que la tarjeta del gate pinta: veredicto por check).
export interface GateCheckResult {
  veredicto: string
  detalle?: string
  check: { id: string; severidad?: string }
}

export interface GateReport {
  target?: string
  results?: GateCheckResult[]
}

// VIEWS is the per-session view strip (mockup it.14). label + glyph.
export const VIEWS: ReadonlyArray<readonly [string, string]> = [
  ["Mapa", "⬡"],
  ["Diag", "⚠"],
  ["Corridas", "▤"],
  ["Tren", "⇢"],
  ["Hist", "◷"],
]

// GLOBAL_VIEWS are the daemon-wide destinations at the rail foot (not per session).
export const GLOBAL_VIEWS: ReadonlyArray<readonly [string, string, string]> = [
  ["portafolio", "⌂", "Portafolio"],
  ["estandar", "⟳", "Estándar"],
  ["ajustes", "⚙", "Ajustes"],
]

export const SALUD_LABEL: Record<Salud, string> = {
  ok: "sano",
  warn: "atención",
  crit: "señales incompletas",
  info: "naciendo",
}

export const STATUS_LABEL: Record<SessionStatus, string> = {
  streaming: "generando…",
  await: "te necesita",
  idle: "en pausa",
}

// ── Dictado por voz (paquete 2026-07-25-spike-voz-dictado, RF-222/RF-223) ──────────────

// EstadoDictado dice si el texto salió ordenado o es el transcripto crudo. La distinción es
// obligatoria en el wire: pasar un crudo por limpio sería un pass fabricado.
export type EstadoDictado = "limpio" | "crudo"

// Dictado is the daemon's answer to one dictation: el texto que va al composer, más cuán
// honestos estamos siendo sobre él.
export interface Dictado {
  texto: string
  estado: EstadoDictado
  // motivo explica por qué quedó crudo. Ausente cuando estado es "limpio".
  motivo?: string
  // motor es el STT que produjo el transcripto (qué motor corrió cambia qué errores esperar).
  motor?: string
}

// DisponibilidadDictado — si se puede dictar y, si no, POR QUÉ y qué instalar (RF-227).
// El composer la consulta al montar: sin esto, la única forma de enterarse de que falta el
// motor sería grabar tres minutos y fallar al final.
export interface DisponibilidadDictado {
  disponible: boolean
  motor?: string
  motivo?: string
  instalar?: string[]
}

// ── Diagnóstico de fallos (paquete 2026-07-25-spike-voz-dictado, RF-230) ───────────────

// EventoDiagnostico es un fallo del FE contado con detalle suficiente para arreglarlo sin
// reproducirlo.
//
// `detalle` es libre a propósito: cada fallo tiene sus propias variables (un dictado necesita
// muestras/pico/hz, un crash de render necesita el stack) y encorsetarlas en un schema fijo
// haría que la próxima falla desconocida no tuviera dónde contarse. El daemon lo escribe tal
// cual al log; nadie parsea su forma como contrato.
export interface EventoDiagnostico {
  // origen dice qué parte lo emitió ("fe").
  origen: string
  // evento es la clave estable para grepear el log ("dictado.sin-audio").
  evento: string
  mensaje?: string
  detalle?: Record<string, unknown>
}
