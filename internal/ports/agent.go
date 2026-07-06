// Package ports declares the interfaces the use cases need. It depends only on the
// domain: the use cases talk to these ports, and concrete adapters (agent, index,
// watch, publish) implement them. Only the composition root (cmd) wires the two.
package ports

import "context"

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
	Raw     []byte
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
}

// AgentSession is a live `claude` conductor: streaming user turns in, normalized
// events out, over a persistent subprocess.
type AgentSession interface {
	// Send streams one user turn to the subprocess stdin. Safe to call across turns
	// while the session stays alive.
	Send(ctx context.Context, turn string) error
	// Events yields normalized events until the session closes.
	Events() <-chan AgentEvent
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
