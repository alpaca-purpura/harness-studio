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
	// sello de build del binario que escribe, estampado en el sobre. Vacío cuando el
	// Registry se construyó con NewRegistry a secas: el sobre lo omite en vez de mentir.
	sello string
	mu    sync.Mutex
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
// empty registry (first run). Lee las dos formas: el sobre versionado que este binario
// escribe y el array desnudo que se escribía antes, porque un registro sin migrar
// (el de sesiones archivadas, p. ej.) sigue siendo legible.
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
	version, err := detectarVersion(b)
	if err != nil {
		return nil, fmt.Errorf("store: %s: %w", r.path, err)
	}
	if version > EsquemaActual {
		return nil, fmt.Errorf(
			"store: %s lo escribió un binario más nuevo (esquema %d, este entiende %d)",
			r.path, version, EsquemaActual)
	}
	payload, err := payloadDe(b, version)
	if err != nil {
		return nil, fmt.Errorf("store: %s: %w", r.path, err)
	}
	var sessions []domain.Session
	if err := json.Unmarshal(payload, &sessions); err != nil {
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
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("store: mkdir %s: %w", dir, err)
	}
	payload, err := json.Marshal(sessions)
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}
	b, err := envolver(payload, r.sello)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".sessions-*.json")
	if err != nil {
		return fmt.Errorf("store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup: a no-op after a successful rename, and on the error paths the
	// write/close error below is the one worth reporting, not the leftover-temp removal.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close() // the write error is the root cause; Close only releases the fd.
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
