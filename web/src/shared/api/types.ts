// Contract types mirroring the daemon's domain (internal/domain/session.go) and the
// Dock SSE frame (internal/usecase/session_service.go dockFrame). These are the shared
// vocabulary the whole SPA speaks; keep them in lockstep with the Go side.

export type SessionStatus = "streaming" | "await" | "idle"
export type Salud = "ok" | "warn" | "crit" | "info"
export type Rol = "user" | "assistant" | "sys"

export interface Turn {
  rol: Rol
  text: string
}

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
  // These are built via spreads (`?? prev`), so they may be explicitly undefined —
  // allowed under exactOptionalPropertyTypes only if the type includes undefined.
  claude_session_id?: string | undefined
  model?: string | undefined
  ctx_pct?: number | undefined
  conv?: Turn[]
  // Sesión de reparación (RF-191, ley A4): abierta contra una instalación del Portafolio.
  reparacion?: boolean
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
  id: string
  rol?: string
  proceso?: string
  empresas?: string[]
}

// DockFrame is one SSE `dock` event payload. Every frame carries session_id so one
// connection multiplexes N conversations (fase 4 c.1). kind=permission es la tarjeta
// ask→UI de un control_request (RF-113: request_id + tool + input crudo para pintar el
// diff; la sesión pasa a `await`); permission_result la cierra con la decisión efectiva.
export interface DockFrame {
  session_id: string
  run_id?: string
  kind:
    | "status"
    | "init"
    | "delta"
    | "message"
    | "result"
    | "error"
    | "permission"
    | "permission_result"
  text?: string
  status?: SessionStatus
  ctx_pct?: number
  model?: string
  claude_session_id?: string
  request_id?: string
  tool?: string
  input?: unknown
  decision?: "allow" | "deny"
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
