// Package ports declares the interfaces the use cases need. It depends only on the
// domain: the use cases talk to these ports, and concrete adapters (agent, index,
// watch, publish) implement them. Only the composition root (cmd) wires the two.
package ports

import "context"

// AgentEvent is one stream-json message emitted by the conductor subprocess. Raw is
// the untouched JSONL line — enumerate/replay it, never re-decode its schema inside a
// use case (see arch/boundaries/conductor-no-parsea-jsonl.md).
type AgentEvent struct {
	Type string // stream-json message type, e.g. "assistant" | "tool_use" | "result".
	Raw  []byte
}

// AgentSession is a live `claude` conductor: streaming input in, stream-json out.
type AgentSession interface {
	// Send streams one user turn to the subprocess stdin.
	Send(ctx context.Context, turn string) error
	// Events yields the subprocess stream-json output until the session closes.
	Events() <-chan AgentEvent
	// Close terminates the subprocess.
	Close() error
}

// AgentPort spawns and drives a Claude Code conductor (conductor pattern, I-76). It is
// deliberately interchangeable: nothing outside the concrete adapter knows which agent
// backs it (see arch/boundaries/adaptadores-de-agente-intercambiables.md).
type AgentPort interface {
	// Spawn starts a conductor for prompt and returns the live session.
	Spawn(ctx context.Context, prompt string) (AgentSession, error)
}
