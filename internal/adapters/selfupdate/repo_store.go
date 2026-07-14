package selfupdate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// RepoStore implements ports.RepoConfigStore: persiste el path del repo configurado
// vía UI cuando el daemon arrancó sin --repo/ARNESIA_REPO (bugfix fix-repo-self-update,
// RF-108). Un único campo, mismo patrón de escritura atómica que
// internal/adapters/portafolio/store.go#saveLocked.
type RepoStore struct {
	path string
	mu   sync.Mutex
}

var _ ports.RepoConfigStore = (*RepoStore)(nil)

// repoDoc es la forma en disco de ~/.arnesia/self-update.json.
type repoDoc struct {
	Repo string `json:"repo"`
}

// NewRepoStore abre (o prepara) el store en path (default
// ~/.arnesia/self-update.json). Un archivo ausente NO es error — Leer devuelve "".
func NewRepoStore(path string) (*RepoStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("selfupdate: repo store: resolve home: %w", err)
		}
		path = filepath.Join(home, ".arnesia", "self-update.json")
	}
	return &RepoStore{path: path}, nil
}

// Leer devuelve el path persistido, o "" si el archivo no existe aún. Un archivo
// presente pero ilegible/corrupto degrada a "" (honesto, jamás aborta el boot del
// daemon por un dato de configuración secundario).
func (s *RepoStore) Leer() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("selfupdate: repo store: leer %s: %w", s.path, err)
	}
	var doc repoDoc
	if uerr := json.Unmarshal(b, &doc); uerr != nil {
		return "", nil // corrupto: se trata como "nunca configurado", no como fallo de boot.
	}
	return doc.Repo, nil
}

// Guardar persiste path (escritura atómica tmp+rename), reemplazando cualquier valor
// previo.
func (s *RepoStore) Guardar(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := json.MarshalIndent(repoDoc{Repo: path}, "", "  ")
	if err != nil {
		return fmt.Errorf("selfupdate: repo store: encode: %w", err)
	}

	dir := filepath.Dir(s.path)
	if merr := os.MkdirAll(dir, 0o750); merr != nil {
		return fmt.Errorf("selfupdate: repo store: mkdir %s: %w", dir, merr)
	}
	tmp, err := os.CreateTemp(dir, ".self-update-*.json")
	if err != nil {
		return fmt.Errorf("selfupdate: repo store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op tras un rename OK.
	if _, werr := tmp.Write(b); werr != nil {
		_ = tmp.Close()
		return fmt.Errorf("selfupdate: repo store: write temp: %w", werr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("selfupdate: repo store: close temp: %w", cerr)
	}
	if rerr := os.Rename(tmpName, s.path); rerr != nil {
		return fmt.Errorf("selfupdate: repo store: rename %s: %w", s.path, rerr)
	}
	return nil
}
