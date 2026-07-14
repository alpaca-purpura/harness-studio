package portafolio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// HashFormaPlugin computa un sha256 agregado y DETERMINISTA del árbol de dir: paths
// relativos (con '/', ordenados) + contenido de cada archivo. Excluye `.git/`, `.in_use`,
// `.orphaned_at` (S0-D7: metadata del cache CC / clones, no forma-plugin). Es la única
// vía de veredicto de deriva (BR-4) — NUNCA semver-string, NUNCA `git status`.
func HashFormaPlugin(dir string) (string, error) {
	var archivos []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		nombre := d.Name()
		if d.IsDir() && nombre == ".git" {
			return filepath.SkipDir
		}
		if nombre == ".in_use" || nombre == ".orphaned_at" {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		archivos = append(archivos, relSlash)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("hash forma-plugin %s: %w", dir, err)
	}
	sort.Strings(archivos)

	h := sha256.New()
	for _, rel := range archivos {
		b, rerr := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))) //nolint:gosec // G304: ruta bajo el árbol que el caller pidió hashear.
		if rerr != nil {
			return "", fmt.Errorf("hash forma-plugin: leer %s: %w", rel, rerr)
		}
		h.Write([]byte(rel))
		h.Write([]byte{0})
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Referencias resuelve dónde vive, LOCALMENTE, la copia de referencia inmutable de un
// arnés (S0-D7: sin red, sin gh/PAT — Slice 0 solo mira lo que ya está en el disco).
type Referencias struct {
	// CCPluginsDir default ~/.claude/plugins — INYECTABLE para tests.
	CCPluginsDir string
}

func (r *Referencias) ccPluginsDir() string {
	if r.CCPluginsDir != "" {
		return r.CCPluginsDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "plugins")
}

// RutaReferencia busca, entre los marketplaces que `known_marketplaces.json` conoce
// LOCALMENTE, el que canonicaliza al mismo `home` y arma
// "<installLocation>/plugins/<id>/<version>/". ok=false si ningún marketplace local
// resuelve a ese home, o si el dir de esa versión no existe — nada se fabrica.
func (r *Referencias) RutaReferencia(home, id, version string) (dir string, ok bool) {
	if home == "" || id == "" || version == "" {
		return "", false
	}
	marketplaces, mok := leerKnownMarketplaces(filepath.Join(r.ccPluginsDir(), "known_marketplaces.json"))
	if !mok {
		return "", false
	}
	for _, m := range marketplaces {
		canon, cok := domain.CanonicalizarRepo(m.Source.Repo)
		if !cok || canon != home {
			continue
		}
		candidato := filepath.Join(m.InstallLocation, "plugins", id, version)
		if fi, serr := os.Stat(candidato); serr == nil && fi.IsDir() {
			return candidato, true
		}
	}
	return "", false
}

// EvaluarDeriva compara installDir contra su referencia inmutable (BR-4): sin version
// conocida, o sin referencia local accesible, → (DerivaNoEvaluable, motivo) — el default
// honesto de S0-D7. Con referencia: hash(install) == hash(ref) → al-hilo; distinto →
// en-deriva. Dos árboles con LA MISMA version en plugin.json pero contenido distinto SÍ
// son en-deriva — el veredicto sale del hash, jamás del string de versión (BR-4).
func EvaluarDeriva(installDir string, refs *Referencias, home, id, version string) (domain.EstadoDeriva, string) {
	if version == "" {
		return domain.DerivaNoEvaluable, "sin version instalada conocida"
	}
	refDir, ok := refs.RutaReferencia(home, id, version)
	if !ok {
		return domain.DerivaNoEvaluable, "sin referencia local accesible (marketplace no clonado, o esa versión ausente del checkout)"
	}
	hashInstall, err := HashFormaPlugin(installDir)
	if err != nil {
		return domain.DerivaNoEvaluable, fmt.Sprintf("no se pudo hashear la instalación: %v", err)
	}
	hashRef, err := HashFormaPlugin(refDir)
	if err != nil {
		return domain.DerivaNoEvaluable, fmt.Sprintf("no se pudo hashear la referencia: %v", err)
	}
	if hashInstall == hashRef {
		return domain.DerivaAlHilo, ""
	}
	return domain.DerivaEnDeriva, "hash de contenido distinto de la referencia " + refDir
}
