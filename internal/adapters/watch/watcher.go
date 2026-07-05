// Package watch implements ports.WatchPort. Stub: it returns a channel that never
// fires and closes when the context is done.
//
// TODO(fase 5): fsnotify — watch ~/.claude/projects for JSONL changes and emit one
// WatchEvent per change to drive incremental reindexing.
package watch

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Watcher is a no-op ports.WatchPort until the fsnotify adapter lands.
type Watcher struct{}

var _ ports.WatchPort = (*Watcher)(nil)

// New returns a stub watcher.
func New() *Watcher { return &Watcher{} }

// Watch returns a channel that carries no events and is closed when ctx is done.
func (w *Watcher) Watch(ctx context.Context) (<-chan ports.WatchEvent, error) {
	ch := make(chan ports.WatchEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}
