// Package watch implements ports.WatchPort with fsnotify: it observes the working
// directory of every arnés known to ports.ArnesRegistry (RF-210, the same corpus
// ports.IndexPort.Rebuild reads — never ~/.claude/projects, a separate conversation
// corpus read by internal/adapters/history) and streams filesystem changes so the
// caller can drive an incremental reindex of just the arnés that owns the change.
package watch

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/fsnotify/fsnotify"
)

// Watcher is an fsnotify-backed ports.WatchPort over the arnés trees reg knows about.
type Watcher struct {
	reg ports.ArnesRegistry
}

var _ ports.WatchPort = (*Watcher)(nil)

// New returns a watcher over the working directories reg knows about. reg may be nil
// for callers that never invoke Watch (matches index.New's convention).
func New(reg ports.ArnesRegistry) *Watcher { return &Watcher{reg: reg} }

// Watch observes every arnés directory registered in reg AT THE TIME Watch IS CALLED
// (recursively) and returns a channel of WatchEvent; it closes when ctx is done. A
// directory that fails to watch is logged and skipped, never fatal — one broken tree
// must not blind the others. An arnés registered after Watch starts is not picked up
// mid-run: documented gap (RF-210, no caller needs it yet — restarting the daemon
// re-watches the current set via Rebuild+Watch on the next boot).
func (w *Watcher) Watch(ctx context.Context) (<-chan ports.WatchEvent, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watch: %w", err)
	}
	if w.reg != nil {
		for _, ap := range w.reg.List() {
			if aerr := addRecursive(fsw, ap.Path); aerr != nil {
				slog.Warn("watch: no se pudo observar árbol de arnés", "arnes", ap.Arnes, "path", ap.Path, "err", aerr)
			}
		}
	}

	out := make(chan ports.WatchEvent)
	go func() {
		defer close(out)
		defer func() { _ = fsw.Close() }()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-fsw.Events:
				if !ok {
					return
				}
				// fsnotify no es recursivo (ni en Linux/inotify): una carpeta nueva
				// creada bajo un árbol ya observado necesita su propio Add, si no
				// queda ciego a lo que se cree DENTRO de ella después.
				if ev.Op&fsnotify.Create != 0 {
					if st, serr := os.Stat(ev.Name); serr == nil && st.IsDir() {
						if aerr := addRecursive(fsw, ev.Name); aerr != nil {
							slog.Warn("watch: no se pudo sumar subdirectorio nuevo", "path", ev.Name, "err", aerr)
						}
					}
				}
				op, ok := opFor(ev.Op)
				if !ok {
					continue // metadata-only (Chmod): sin cambio de contenido, se ignora.
				}
				select {
				case out <- ports.WatchEvent{Path: ev.Name, Op: op}:
				case <-ctx.Done():
					return
				}
			case werr, ok := <-fsw.Errors:
				if !ok {
					return
				}
				slog.Warn("watch: fsnotify", "err", werr)
			}
		}
	}()
	return out, nil
}

// addRecursive adds root and every subdirectory under it to fsw — fsnotify only
// watches the exact directories it is told about, so a nested arnés tree
// (.claude/skills/x/, .claude/hooks/, …) needs one Add per directory.
func addRecursive(fsw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fsw.Add(path)
		}
		return nil
	})
}

// opFor maps an fsnotify op to the coarser ports.WatchOp the caller reasons about.
// Rename is reported as a remove (the old path stops existing at that name); a pure
// Chmod carries no content change, so it maps to (_, false) — dropped by the caller.
func opFor(op fsnotify.Op) (ports.WatchOp, bool) {
	switch {
	case op&fsnotify.Create != 0:
		return ports.WatchCreate, true
	case op&fsnotify.Write != 0:
		return ports.WatchWrite, true
	case op&(fsnotify.Remove|fsnotify.Rename) != 0:
		return ports.WatchRemove, true
	default:
		return "", false
	}
}
