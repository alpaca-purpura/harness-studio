// Package store implements ports.SessionStore as a single JSON file under the user's
// home (~/.arnesia/sessions.json). It is deliberately trivial: the session *registry*
// is small and low-churn, and the disposable SQLite index (fase 5) is a separate
// concern. Writes are atomic (temp file + rename) so a crash mid-save never corrupts
// the registry.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Registry is a file-backed session store. Safe for concurrent use.
type Registry struct {
	path string
	mu   sync.Mutex
}

var _ = (interface {
	Load(context.Context) ([]domain.Session, error)
	Save(context.Context, []domain.Session) error
})((*Registry)(nil))

// NewRegistry returns a Registry writing to path. If path is empty it defaults to
// ~/.arnesia/sessions.json. The parent directory is created on first Save.
func NewRegistry(path string) (*Registry, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("store: resolve home: %w", err)
		}
		path = filepath.Join(home, ".arnesia", "sessions.json")
	}
	return &Registry{path: path}, nil
}

// Load reads the persisted sessions. A missing file is not an error — it yields an
// empty registry (first run).
func (r *Registry) Load(_ context.Context) ([]domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return []domain.Session{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", r.path, err)
	}
	var sessions []domain.Session
	if err := json.Unmarshal(b, &sessions); err != nil {
		return nil, fmt.Errorf("store: decode %s: %w", r.path, err)
	}
	return sessions, nil
}

// Save atomically replaces the file with sessions (temp file in the same directory,
// then rename — an atomic swap on the same filesystem).
func (r *Registry) Save(_ context.Context, sessions []domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("store: mkdir %s: %w", dir, err)
	}
	b, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".sessions-*.json")
	if err != nil {
		return fmt.Errorf("store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename.

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("store: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("store: close temp: %w", err)
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		return fmt.Errorf("store: rename %s: %w", r.path, err)
	}
	return nil
}
