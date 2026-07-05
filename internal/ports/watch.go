package ports

import "context"

// WatchOp is the kind of filesystem change observed (fsnotify-shaped, fase 5).
type WatchOp string

const (
	WatchWrite  WatchOp = "write"
	WatchCreate WatchOp = "create"
	WatchRemove WatchOp = "remove"
)

// WatchEvent is a single filesystem change under the watched tree (~/.claude JSONL).
type WatchEvent struct {
	Path string
	Op   WatchOp
}

// WatchPort observes the JSONL corpus and streams change events that trigger an
// incremental reindex.
type WatchPort interface {
	// Watch starts observing and returns a channel that closes when ctx is done.
	Watch(ctx context.Context) (<-chan WatchEvent, error)
}
