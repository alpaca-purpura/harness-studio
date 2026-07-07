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
}

// HarnessSummary is one entry of GET /api/harnesses (S1 portfolio / the Map picker, RF-72):
// a lightweight row, not the full graph. The domain Graph/Box types live in entities/arnes
// (shared/api must not import upward), so the page maps this to whatever it renders.
export interface HarnessSummary {
  id: string
  rol?: string
  proceso?: string
  empresa?: string
}

// DockFrame is one SSE `dock` event payload. Every frame carries session_id so one
// connection multiplexes N conversations (fase 4 c.1).
export interface DockFrame {
  session_id: string
  run_id?: string
  kind: "status" | "init" | "delta" | "message" | "result" | "error"
  text?: string
  status?: SessionStatus
  ctx_pct?: number
  model?: string
  claude_session_id?: string
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
