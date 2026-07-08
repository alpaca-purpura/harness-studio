// Package provision implements ports.InjectionProvisioner: materializes the embedded
// doctrine+kit (cuerpos ①+②) to the app-owned dir (~/.arnesia) and hands back the
// Injection every conductor spawn carries. Idempotente por huella de contenido: el
// primer Provision de cada versión del binario re-escribe; los siguientes no tocan
// disco. El usuario jamás configura nada (constraint de UX de la investigación de inyección,
// FIRMADO HS-10) y nada de esto se escribe en el árbol de un arnés (② ↛ ③).
package provision

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Provisioner materializes kitFS (el plugin ②, raíz `kit/`) and the knowledge elements
// of doctrinaFS (los checklists ①) under baseDir.
type Provisioner struct {
	baseDir    string
	kitFS      fs.FS
	doctrinaFS fs.FS

	mu   sync.Mutex
	done *ports.Injection // cached tras el primer Provision exitoso del proceso.
}

var _ ports.InjectionProvisioner = (*Provisioner)(nil)

// New returns a Provisioner that materializes under baseDir (empty = ~/.arnesia).
func New(baseDir string, kitFS, doctrinaFS fs.FS) (*Provisioner, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("provision: home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".arnesia")
	}
	return &Provisioner{baseDir: baseDir, kitFS: kitFS, doctrinaFS: doctrinaFS}, nil
}

// Provision materializes (idempotent) and returns the Injection. La huella de todo el
// contenido embebido gobierna el refresh: binario nuevo → doctrina nueva en disco.
func (p *Provisioner) Provision(_ context.Context) (ports.Injection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.done != nil {
		return *p.done, nil
	}

	fp, err := fingerprint(p.kitFS, p.doctrinaFS)
	if err != nil {
		return ports.Injection{}, fmt.Errorf("provision: fingerprint: %w", err)
	}
	stampPath := filepath.Join(p.baseDir, ".doctrina-version")
	if prev, err := os.ReadFile(stampPath); err != nil || strings.TrimSpace(string(prev)) != fp { //nolint:gosec // G304: ruta propia bajo ~/.arnesia.
		if err := p.materialize(); err != nil {
			return ports.Injection{}, err
		}
		if err := os.WriteFile(stampPath, []byte(fp+"\n"), 0o600); err != nil {
			return ports.Injection{}, fmt.Errorf("provision: stamp: %w", err)
		}
	}

	inj := ports.Injection{
		PluginDirs:       []string{filepath.Join(p.baseDir, "kit")},
		SystemPromptFile: filepath.Join(p.baseDir, "doctrine.md"),
		AddDirs:          []string{filepath.Join(p.baseDir, "knowhow")},
	}
	p.done = &inj
	return inj, nil
}

// materialize writes kit/ (plugin ② completo), doctrine.md (overlay ①) and knowhow/
// (los nodos del estándar, referencia read-only) under baseDir.
func (p *Provisioner) materialize() error {
	if err := copyTree(p.kitFS, "kit", filepath.Join(p.baseDir, "kit")); err != nil {
		return fmt.Errorf("provision: kit: %w", err)
	}
	// El overlay vive dentro del kit embebido pero se inyecta como archivo suelto
	// (--append-system-prompt-file no lee de plugins).
	overlay, err := fs.ReadFile(p.kitFS, "kit/doctrine.md")
	if err != nil {
		return fmt.Errorf("provision: doctrine.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(p.baseDir, "doctrine.md"), overlay, 0o600); err != nil {
		return fmt.Errorf("provision: doctrine.md: %w", err)
	}
	if err := copyTree(p.doctrinaFS, "knowledge/elements", filepath.Join(p.baseDir, "knowhow")); err != nil {
		return fmt.Errorf("provision: knowhow: %w", err)
	}
	return nil
}

// copyTree copies srcRoot (inside fsys) under dstRoot, flattening srcRoot's prefix.
func copyTree(fsys fs.FS, srcRoot, dstRoot string) error {
	return fs.WalkDir(fsys, srcRoot, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, srcRoot), "/")
		dst := filepath.Join(dstRoot, filepath.FromSlash(rel))
		if de.IsDir() {
			return os.MkdirAll(dst, 0o750)
		}
		b, rerr := fs.ReadFile(fsys, p)
		if rerr != nil {
			return rerr
		}
		return os.WriteFile(dst, b, 0o600)
	})
}

// fingerprint hashes every embedded path+content so ANY doctrine/kit change (a new
// binary) re-materializes on first use.
func fingerprint(trees ...fs.FS) (string, error) {
	h := sha256.New()
	for _, fsys := range trees {
		var paths []string
		if err := fs.WalkDir(fsys, ".", func(p string, de fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !de.IsDir() {
				paths = append(paths, p)
			}
			return nil
		}); err != nil {
			return "", err
		}
		sort.Strings(paths)
		for _, p := range paths {
			b, err := fs.ReadFile(fsys, p)
			if err != nil {
				return "", err
			}
			// hash.Hash.Write jamás devuelve error (contrato de la stdlib).
			_, _ = fmt.Fprintf(h, "%s\n", p)
			_, _ = h.Write(b)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
