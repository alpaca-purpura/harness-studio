// Package claudecode implements ports.AgentPort by spawning the local `claude`
// binary as a subprocess and speaking stream-json over its stdin/stdout — the
// conductor pattern (I-76 / OBS-16 / OBS-18): Go owns the process and drives it,
// the agent is just an adapter.
//
// Permissions are resolved by the GUI, human-in-the-loop (control_request), and are
// never bypassed (see arch/boundaries/permisos-gui-human-in-the-loop.md). Write/Edit
// therefore do not go in --allowedTools; they pass through the diff-approval flow.
//
// This is a compiling stub: it wires os/exec but real streaming lands in fase 5.
package claudecode

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// errNotImplemented marks the parts of the conductor still owed to fase 5.
var errNotImplemented = errors.New("not implemented")

// Conductor is the Claude Code adapter. bin is the path to the `claude` executable.
type Conductor struct {
	bin string
}

var _ ports.AgentPort = (*Conductor)(nil)

// New returns a Conductor that will spawn the given `claude` binary (e.g. "claude").
func New(bin string) *Conductor {
	return &Conductor{bin: bin}
}

// Spawn builds the conductor subprocess in stream-json mode. The command is wired
// here to pin the contract; starting it and pumping the streams is fase-5 work.
func (c *Conductor) Spawn(ctx context.Context, prompt string) (ports.AgentSession, error) {
	// stream-json both ways: streaming input on stdin, structured events on stdout.
	cmd := exec.CommandContext(ctx, c.bin,
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--verbose",
		"-p", prompt,
	)
	_ = cmd // TODO(fase 5): Start(), pump stdin turns and stdout events onto AgentSession.
	return nil, fmt.Errorf("claudecode: spawn: %w", errNotImplemented)
}
