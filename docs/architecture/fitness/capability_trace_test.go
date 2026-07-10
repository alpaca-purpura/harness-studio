package fitness

// Enforcer de `docs/architecture/boundaries/codigo-traza-a-capability.md` (HS-18 · homologación HS-19 2026-07-09):
// R1 integridad — cada puntero de una capability resuelve a un archivo real (docs no cuelgan).
// R2 cobertura — todo archivo fuente está reclamado por ≥1 capability (o allowlist con razón):
//                 no hay código huérfano de capability.
// SSoT funcional = docs/product/capabilities/{module}/{slug}.yaml (YAML por-cap + BDD, formato
// del kit harness@prenter-marketplace). Migrado desde CAPABILITIES.md; el doctor rápido local es
// scripts/cap_doctor.py. Los punteros viajan como tokens entre comillas `"ruta#Símbolo"` en el
// bloque `pointers:` (+ soporte en _coverage.yaml `support_files:`).

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var capQuoted = regexp.MustCompile(`"([^"]+)"`)

var capSourceExts = []string{".go", ".ts", ".tsx", ".rs"}

// capFileToken extrae la ruta de un token de puntero (`ruta#Símbolo` → `ruta`),
// devolviendo (ruta, true) solo si termina en una extensión de código fuente.
func capFileToken(tok string) (string, bool) {
	p := tok
	if i := strings.IndexAny(tok, "#:,"); i >= 0 {
		p = tok[:i]
	}
	p = strings.TrimSpace(p)
	for _, e := range capSourceExts {
		if strings.HasSuffix(p, e) {
			return p, true
		}
	}
	return "", false
}

func capHasSourceExt(p string) bool {
	for _, e := range capSourceExts {
		if strings.HasSuffix(p, e) {
			return true
		}
	}
	return false
}

// capAllowed = código de soporte que legítimamente NO es una capability (con razón).
func capAllowed(p string) bool {
	switch {
	case strings.HasSuffix(p, "_test.go"): // pruebas
	case strings.HasSuffix(p, ".stories.tsx"): // fitness visual, no código de producto
	case strings.HasSuffix(p, ".d.ts"): // decls generadas (vite-env)
	case strings.HasSuffix(p, "/index.ts"): // barrels FSD
	case strings.Contains(p, "/entities/arnes/testing/"): // fixtures dogfood
	case strings.Contains(p, "/entities/arnes/model/"): // derivación (no capability)
	case strings.Contains(p, "/shared/ui/"): // átomos UI transversales
	case strings.Contains(p, "/shared/canvas/"): // glyph/lienzo chrome
	case strings.Contains(p, "/shared/lib/"): // cn() etc
	case strings.Contains(p, "/shared/config/"): // tokens
	case strings.HasSuffix(p, "/shared/api/types.ts"): // DTOs
	case strings.HasPrefix(p, "web/src/app/"): // bootstrap App/main
	case strings.Contains(p, "/features/self-update/model/"): // DTOs
	case strings.HasPrefix(p, "internal/ports/"): // contratos (impls reclamadas)
	case p == "internal/adapters/loader/frontmatter.go": // soporte de reconocedores
	default:
		return false
	}
	return true
}

// capClaims recorre las hojas capability (+ _coverage.yaml) y colecta las RUTAS de código
// reclamadas: los tokens entre comillas SOLO del bloque `pointers:` (y `support_files:` en
// _coverage.yaml), nunca de `valida:` (que lleva nombres/ficheros de test). La forma es
// machine-generated (scripts/capabilities_to_yaml.py) y estable.
func capClaims(t *testing.T) (root string, claims []string) {
	t.Helper()
	root = repoRoot()
	capDir := filepath.Join(root, "docs", "product", "capabilities")
	var files []string
	werr := filepath.WalkDir(capDir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() && strings.HasSuffix(path, ".yaml") {
			files = append(files, path)
		}
		return nil
	})
	if werr != nil {
		t.Fatalf("no se pudo recorrer docs/product/capabilities: %v", werr)
	}
	if len(files) == 0 {
		t.Fatalf("no hay capabilities en docs/product/capabilities (¿migración incompleta?)")
	}
	for _, f := range files {
		b, err := os.ReadFile(f) //nolint:gosec // rutas del propio árbol del repo
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", f, err)
		}
		inBlock := false
		for _, ln := range strings.Split(string(b), "\n") {
			trimmed := strings.TrimSpace(ln)
			if trimmed == "pointers:" || trimmed == "support_files:" {
				inBlock = true
				continue
			}
			if !inBlock {
				continue
			}
			if strings.HasPrefix(trimmed, "- ") {
				if m := capQuoted.FindStringSubmatch(ln); m != nil {
					if p, ok := capFileToken(m[1]); ok {
						claims = append(claims, p)
					}
				}
				continue
			}
			inBlock = false // dedent / nueva clave top-level cierra el bloque
		}
	}
	return root, claims
}

// TestCapabilityPointersResolve — R1: todo puntero de una capability existe en el árbol.
func TestCapabilityPointersResolve(t *testing.T) {
	root, claims := capClaims(t)
	var bad []string
	for _, p := range claims {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err != nil {
			bad = append(bad, p)
		}
	}
	if len(bad) > 0 {
		t.Fatalf("R1: punteros de capabilities que no resuelven a código real: %v", bad)
	}
}

// TestCapabilityCoverage — R2: ningún archivo fuente queda huérfano de capability.
func TestCapabilityCoverage(t *testing.T) {
	root, claims := capClaims(t)
	claimed := map[string]bool{}
	for _, p := range claims {
		claimed[p] = true
	}

	var orphans []string
	for _, d := range []string{"cmd", "internal", filepath.Join("web", "src"), filepath.Join("web", "src-tauri", "src")} {
		base := filepath.Join(root, d)
		werr := filepath.WalkDir(base, func(path string, de fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if de.IsDir() {
				switch de.Name() {
				case "node_modules", "dist", "target":
					return fs.SkipDir
				}
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			if !capHasSourceExt(rel) {
				return nil
			}
			if claimed[rel] || capAllowed(rel) {
				return nil
			}
			orphans = append(orphans, rel)
			return nil
		})
		if werr != nil {
			t.Fatalf("no se pudo recorrer %s: %v", d, werr)
		}
	}
	if len(orphans) > 0 {
		t.Fatalf("R2: %d archivo(s) de código sin capability (reclamar en docs/product/capabilities/ o allowlist con razón): %v", len(orphans), orphans)
	}
}
