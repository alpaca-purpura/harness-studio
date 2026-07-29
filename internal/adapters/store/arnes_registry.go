// arnes_registry.go implements ports.ArnesRegistry (and thus ports.WorkdirResolver):
// the explicit arnés→working-directory mapping that confines each session's conductor to
// its own tree (boundary permisos-gui `sesion-aislada-por-cwd`, HS-06). The mapping is a
// small JSON file under the user's home (~/.arnesia/arneses.json); an unregistered arnés
// resolves to a dedicated per-arnés fallback dir — NEVER a shared global cwd.

package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/adapters/filelock"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// ArnesRegistry is a file-backed arnés→path registry. Safe for concurrent use.
type ArnesRegistry struct {
	path         string // arneses.json
	fallbackRoot string // dir holding per-arnés fallback trees for unregistered arneses
	home         string // user home, for the protected-path denylist

	mu      sync.Mutex
	entries map[string]string // arnesID → validated absolute path
}

var _ ports.ArnesRegistry = (*ArnesRegistry)(nil)

// NewArnesRegistry returns a registry writing to path (default ~/.arnesia/arneses.json)
// with unregistered arneses falling back under fallbackRoot (default ~/.arnesia/arneses).
// It loads any existing mapping so a restart keeps registered paths.
func NewArnesRegistry(path, fallbackRoot string) (*ArnesRegistry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("arnes registry: resolve home: %w", err)
	}
	if path == "" {
		path = filepath.Join(home, ".arnesia", "arneses.json")
	}
	if fallbackRoot == "" {
		fallbackRoot = filepath.Join(home, ".arnesia", "arneses")
	}
	r := &ArnesRegistry{
		path:         path,
		fallbackRoot: fallbackRoot,
		home:         home,
		entries:      map[string]string{},
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// Resolve returns the working directory for arnesID. A registered arnés yields its
// validated path (registered=true); otherwise a dedicated fallback dir under fallbackRoot
// is created and returned (registered=false). It never returns a shared cwd.
func (r *ArnesRegistry) Resolve(arnesID string) (string, bool, error) {
	r.mu.Lock()
	p, ok := r.entries[arnesID]
	r.mu.Unlock()
	if ok {
		return p, true, nil
	}
	// Fallback: an isolated per-arnés tree. slug keeps a hostile id from escaping the root.
	fallback := filepath.Join(r.fallbackRoot, slug(arnesID))
	if err := os.MkdirAll(fallback, 0o750); err != nil {
		return "", false, fmt.Errorf("arnes registry: mkdir fallback %s: %w", fallback, err)
	}
	return fallback, false, nil
}

// Register validates path and records it for arnesID (replacing any prior entry).
//
// Corre bajo `filelock.Guard` (Fase 1, D2): recarga desde disco antes de mutar — mismo
// mecanismo que `portafolio.Store.Upsert`, para el mismo riesgo (el daemon y un CLI
// standalone son dos procesos con dos cachés en memoria escribiendo `arneses.json`).
func (r *ArnesRegistry) Register(arnesID, path string) error {
	if strings.TrimSpace(arnesID) == "" {
		return errors.New("arnes registry: empty arnés id")
	}
	clean, err := r.validate(path)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return filelock.Guard(r.path, func() error {
		if rerr := r.reloadLocked(); rerr != nil {
			return rerr
		}
		r.entries[arnesID] = clean
		return r.saveLocked()
	})
}

// List returns every registered arnés→path entry.
func (r *ArnesRegistry) List() []ports.ArnesPath {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ports.ArnesPath, 0, len(r.entries))
	for a, p := range r.entries {
		out = append(out, ports.ArnesPath{Arnes: a, Path: p})
	}
	return out
}

// validate enforces the security contract for a registrable working directory: absolute,
// existing directory, canonicalized, and not a protected/containing location. Returns the
// cleaned absolute path.
func (r *ArnesRegistry) validate(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("arnes registry: empty path")
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("arnes registry: path %q must be absolute", path)
	}
	clean := filepath.Clean(path)
	// Canonicalize through symlinks so containment checks can't be tricked by a link.
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		clean = resolved
	}
	info, err := os.Stat(clean)
	if err != nil {
		return "", fmt.Errorf("arnes registry: path %q: %w", clean, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("arnes registry: path %q is not a directory", clean)
	}
	if err := r.checkProtected(clean); err != nil {
		return "", err
	}
	return clean, nil
}

// checkProtected rejects a path that IS, is INSIDE, or CONTAINS a protected location. The
// "contains" direction matters: running claude in $HOME would expose ~/.claude/~/.ssh.
func (r *ArnesRegistry) checkProtected(clean string) error {
	if clean == string(filepath.Separator) || clean == r.home {
		return fmt.Errorf("arnes registry: path %q is a protected root", clean)
	}
	protected := []string{
		filepath.Join(r.home, ".claude"),
		filepath.Join(r.home, ".arnesia"),
		filepath.Join(r.home, ".ssh"),
		filepath.Join(r.home, ".gnupg"),
		filepath.Join(r.home, ".config"),
	}
	for _, p := range protected {
		if clean == p || within(p, clean) || within(clean, p) {
			return fmt.Errorf("arnes registry: path %q overlaps protected location %q", clean, p)
		}
	}
	return nil
}

// within reports whether child is inside parent (parent is an ancestor of child).
func within(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "."
}

// slug reduces an arbitrary arnés id to a single safe path segment (no separators, no
// traversal). Empty ids collapse to "_".
func slug(id string) string {
	id = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, id)
	id = strings.Trim(id, "-")
	if id == "" {
		return "_"
	}
	return id
}

// load reads the persisted mapping. A missing file is a fresh registry, not an error.
func (r *ArnesRegistry) load() error {
	b, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("arnes registry: read %s: %w", r.path, err)
	}
	var list []ports.ArnesPath
	if err := json.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("arnes registry: decode %s: %w", r.path, err)
	}
	for _, e := range list {
		r.entries[e.Arnes] = e.Path
	}
	return nil
}

// reloadLocked descarta el mapeo en memoria y vuelve a leer `r.path` desde cero. Se llama
// DENTRO de un `filelock.Guard` (Fase 1, D2/D3): mismo mecanismo que
// `portafolio.Store.reloadLocked` — un `Register` tiene que partir de lo que hay en disco en
// ESE instante, nunca de un caché que pudo quedar viejo porque otro proceso escribió
// mientras tanto. Caller sostiene r.mu.
func (r *ArnesRegistry) reloadLocked() error {
	r.entries = map[string]string{}
	return r.load()
}

// saveLocked snapshots the mapping to disk atomically (temp + rename). Caller holds r.mu.
func (r *ArnesRegistry) saveLocked() error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("arnes registry: mkdir %s: %w", dir, err)
	}
	list := make([]ports.ArnesPath, 0, len(r.entries))
	for a, p := range r.entries {
		list = append(list, ports.ArnesPath{Arnes: a, Path: p})
	}
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("arnes registry: encode: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".arneses-*.json")
	if err != nil {
		return fmt.Errorf("arnes registry: temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup: a no-op after a successful rename, and on the error paths the
	// write/close error below is the one worth reporting, not the leftover-temp removal.
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close() // the write error is the root cause; Close only releases the fd.
		return fmt.Errorf("arnes registry: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("arnes registry: close temp: %w", err)
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		return fmt.Errorf("arnes registry: rename %s: %w", r.path, err)
	}
	return nil
}
