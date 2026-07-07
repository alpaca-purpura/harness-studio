// Package publish implements ports.PublishPort by shelling out to git to push a
// harness to its marketplace repo (the release train, KIT-06).
//
// This is a compiling stub: it wires the git command but the real publish (with the
// evals-gate) lands in fase 5.
package publish

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// errNotImplemented marks the publish flow still owed to fase 5.
var errNotImplemented = errors.New("not implemented")

// Publisher is the git-backed ports.PublishPort. git is the executable path.
type Publisher struct {
	git string
}

var _ ports.PublishPort = (*Publisher)(nil)

// New returns a Publisher using the given git binary (e.g. "git").
func New(git string) *Publisher {
	return &Publisher{git: git}
}

// Publish wires the git push to target. Running it (behind the evals-gate) is fase-5
// work.
func (p *Publisher) Publish(ctx context.Context, harnessID, target string) error {
	cmd := exec.CommandContext(ctx, p.git, "push", target, harnessID) //nolint:gosec // G204: p.git is local configuration and target/harnessID come from the operator's own CLI flags; the command is wired but never run (fase-5 stub).
	_ = cmd                                                           // TODO(fase 5): run the release-train publish once the evals-gate passes.
	return fmt.Errorf("publish %q -> %q: %w", harnessID, target, errNotImplemented)
}
