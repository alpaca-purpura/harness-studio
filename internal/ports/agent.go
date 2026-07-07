// Package ports declares the interfaces the use cases need. It depends only on the
// domain: the use cases talk to these ports, and concrete adapters (agent, index,
// watch, publish) implement them. Only the composition root (cmd) wires the two.
package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// AgentEventKind is the normalized meaning of a conductor event. The claudecode
// adapter owns the Claude Code stream-json protocol and translates its raw frames
// into these; nothing upstream re-decodes the protocol (see
// arch/boundaries/conductor-no-parsea-jsonl.md — the on-disk JSONL corpus is for
// enumerate/replay, and the live stream is interpreted only here, by its expert).
type AgentEventKind string

const (
	// EventInit — the `system/init` frame: ClaudeSessionID and Model are set. Capture
	// ClaudeSessionID to later `--resume` this conversation.
	EventInit AgentEventKind = "init"
	// EventDelta — one token chunk of the current assistant message (Text).
	EventDelta AgentEventKind = "delta"
	// EventMessage — a complete assistant message (Text); a fallback when partial
	// deltas were not requested/emitted.
	EventMessage AgentEventKind = "message"
	// EventResult — the turn finished. CtxPct carries context-window usage (0–100).
	EventResult AgentEventKind = "result"
	// EventError — the conductor failed (Text is the reason).
	EventError AgentEventKind = "error"
	// EventControlRequest — a `control_request:can_use_tool` frame: Claude Code asks
	// permission to run a tool (RequestID + Tool + raw Input). The adapter FORWARDS it
	// (never auto-resolves): the daemon decides via the role's PermissionSet and the
	// human-in-the-loop Dock (boundaries permisos-derivan-del-rol +
	// permisos-gui-human-in-the-loop), answering through AgentSession.RespondControl.
	EventControlRequest AgentEventKind = "control_request"
)

// AgentEvent is one normalized message from the conductor. Raw is the untouched
// stream-json line for enumerate/replay; the typed fields are what the Dock renders.
type AgentEvent struct {
	Kind            AgentEventKind
	Text            string
	ClaudeSessionID string
	Model           string
	CtxPct          int
	// Subtype carries the `result` frame subtype (e.g. "success", "error_max_turns").
	// The T3 conductor reads THIS machine signal to decide continue/stop — never the chat
	// text (arch/boundaries/orquestacion-determinista-entre-cajas.md).
	Subtype string
	// RequestID / Tool / Input carry a forwarded control_request (Kind=
	// EventControlRequest): the id to answer with RespondControl, the tool Claude Code
	// wants to run, and the raw JSON input of the call (the GUI paints the diff from it).
	RequestID string
	Tool      string
	Input     []byte
	Raw       []byte
}

// SpawnOpts parameterizes a conductor. Resume, when non-empty, continues an existing
// Claude Code conversation instead of starting a fresh one.
type SpawnOpts struct {
	// Resume is a Claude Code session id to `--resume`; empty starts a new session.
	Resume string
	// Model overrides the default model (e.g. "claude-opus-4-8"); empty uses the CLI default.
	Model string
	// Cwd is the working directory the conductor runs in; empty inherits the daemon's.
	// It is resolved per session from the session's arnés (see WorkdirResolver) so each
	// conductor is confined to its own tree — never a shared global cwd.
	Cwd string
	// MaxTurns caps the agent loop of a single turn (`--max-turns`); 0 means "unset"
	// (no cap). Every unattended conductor run fixes it — a hijacked/looping turn must
	// not run unbounded (permisos-gui `max-turns-siempre`, headless-sdk `headless-max-turns`).
	MaxTurns int
	// Injection carries the doctrine bodies ①+② as spawn flags (--plugin-dir /
	// --append-system-prompt-file / --add-dir; HS-11 puente 2). Zero value = spawn sin
	// doctrina (degradación honesta, jamás bloquea la sesión).
	Injection Injection
	// Permisos is the role-derived permission-set the spawn materializes in CC-native
	// flags (boundary permisos-derivan-del-rol: permission = f(rol), resolved by the
	// PermissionPort at hydration). Zero value = no extra restriction (the interactive
	// Dock spawn keeps Claude Code's own defaults).
	Permisos domain.PermissionSet
}

// ControlDecision is the daemon's answer to a forwarded control_request. On allow the
// original tool input is echoed back (UpdatedInput, raw JSON); on deny Message says why.
type ControlDecision struct {
	Allow        bool
	UpdatedInput []byte
	Message      string
}

// AgentSession is a live `claude` conductor: streaming user turns in, normalized
// events out, over a persistent subprocess.
type AgentSession interface {
	// Send streams one user turn to the subprocess stdin. Safe to call across turns
	// while the session stays alive.
	Send(ctx context.Context, turn string) error
	// Events yields normalized events until the session closes.
	Events() <-chan AgentEvent
	// RespondControl answers a pending control_request (EventControlRequest) over the
	// agent's control channel (stdin for the claudecode adapter). The daemon — role
	// permission-set + human approval — is the only caller; the adapter never decides.
	RespondControl(ctx context.Context, requestID string, d ControlDecision) error
	// Close terminates the subprocess (closing stdin, then waiting).
	Close() error
}

// AgentPort spawns and drives a Claude Code conductor (conductor pattern, I-76). It is
// deliberately interchangeable: nothing outside the concrete adapter knows which agent
// backs it (see arch/boundaries/adaptadores-de-agente-intercambiables.md).
type AgentPort interface {
	// Spawn starts a persistent conductor and returns the live session.
	Spawn(ctx context.Context, opts SpawnOpts) (AgentSession, error)
}
