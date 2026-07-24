package ports

import "context"

// WatchOp is the kind of filesystem change observed (fsnotify-shaped, fase 5).
type WatchOp string

// The three change kinds the watcher reports: an appended/updated file, a new file,
// and a removed file under the watched arnés tree.
const (
	WatchWrite  WatchOp = "write"
	WatchCreate WatchOp = "create"
	WatchRemove WatchOp = "remove"
)

// WatchEvent is a single filesystem change under the watched arnés tree (the harness
// directories the Portafolio knows about — see IndexPort, not ~/.claude JSONL).
type WatchEvent struct {
	Path string
	Op   WatchOp
}

// WatchPort observes the arnés tree (the same corpus IndexPort.Rebuild reads) and
// streams change events that trigger an incremental reindex.
type WatchPort interface {
	// Watch starts observing and returns a channel that closes when ctx is done.
	Watch(ctx context.Context) (<-chan WatchEvent, error)
}
