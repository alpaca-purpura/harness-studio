package domain

// A Session is a *frente de trabajo* (work-front): one live Claude Code conversation
// bound N:1 to an arnés. It is the central unit of the factory thesis — the app runs
// many of these in parallel (multisesión, HS-03 it.14). The daemon owns the registry
// of open sessions (source of truth); the conversation itself is rehydrated from
// Claude Code via `--resume ClaudeSessionID`, so Conv here is only the lightweight
// transcript the shell replays instantly on reopen.

// SessionStatus is the live state of a session's conductor, mirrored on the rail pip.
type SessionStatus string

const (
	// StatusStreaming — the conductor is generating (pip pulses live-blue).
	StatusStreaming SessionStatus = "streaming"
	// StatusAwait — Claude Code needs the human (permission / decision); pip pulses warn.
	StatusAwait SessionStatus = "await"
	// StatusIdle — parked, waiting for the next turn (pip dim).
	StatusIdle SessionStatus = "idle"
)

// Valid reports whether s is a known status.
func (s SessionStatus) Valid() bool {
	switch s {
	case StatusStreaming, StatusAwait, StatusIdle:
		return true
	default:
		return false
	}
}

// Salud is the health of the arnés behind a session, shown as the rail health dot.
// Mirrors the mockup's ok|warn|crit|info scale.
type Salud string

// The four health grades painted on the rail dot, from healthy to newborn.
const (
	SaludOK   Salud = "ok"   // sano.
	SaludWarn Salud = "warn" // atención.
	SaludCrit Salud = "crit" // señales incompletas.
	SaludInfo Salud = "info" // naciendo.
)

// Rol is who authored a transcript turn. "sys" is a system breadcrumb rendered inline
// (skill_activated, hand-offs) — not a chat bubble.
type Rol string

// The three turn authors: the human, the conductor's assistant, and the inline
// system breadcrumb (see the Rol comment).
const (
	RolUser      Rol = "user"
	RolAssistant Rol = "assistant"
	RolSys       Rol = "sys"
)

// Turn is one entry of a session's lightweight transcript. It is NOT the source of
// truth for the conversation (Claude Code's own JSONL is); it is what the shell shows
// immediately when you switch back to a session, before/while the live stream resumes.
type Turn struct {
	Rol  Rol    `json:"rol"`
	Text string `json:"text"`
}

// Session is a work-front: an open Claude Code conversation over one arnés.
type Session struct {
	// ID is our stable session id (survives daemon restarts); distinct from the
	// Claude Code session id used to --resume the conversation.
	ID string `json:"id"`

	// Frente is the human name of the work-front, auto-derived from the first user
	// message and editable in the rail. It is what the rail card shows.
	Frente string `json:"frente"`

	// Arnes is the id of the harness this front operates on (N sessions may share one).
	Arnes   string `json:"arnes"`
	Empresa string `json:"empresa,omitempty"`
	Puesto  string `json:"puesto,omitempty"`
	Salud   Salud  `json:"salud,omitempty"`

	// Status is the live conductor state; View is the parked view (Mapa|Diag|…);
	// Parked is the free-text "quedaste en …" landing hint.
	Status SessionStatus `json:"status"`
	View   string        `json:"view"`
	Parked string        `json:"parked,omitempty"`

	// ClaudeSessionID is the Claude Code session id captured from the `system/init`
	// event; passing it to `--resume` rehydrates the real conversation after a restart.
	ClaudeSessionID string `json:"claude_session_id,omitempty"`
	Model           string `json:"model,omitempty"`
	// CtxPct is the last-known context-window usage (0–100) for the dock's ctx bar.
	CtxPct int `json:"ctx_pct,omitempty"`

	// Conv is the lightweight replay transcript (see Turn).
	Conv []Turn `json:"conv,omitempty"`
}
