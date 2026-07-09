package fitness

// Enforcer de `arch/boundaries/codigo-traza-a-capability.md` (HS-18):
// R1 integridad — cada puntero de CAPABILITIES.md resuelve a un archivo real (docs no cuelgan).
// R2 cobertura — todo archivo fuente está reclamado por ≥1 capability (o allowlist con razón):
//                 no hay código huérfano de capability.
// CAPABILITIES.md = SSoT funcional (docs-as-code, Living Documentation / BDD).

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var capBacktick = regexp.MustCompile("`([^`]+)`")

var capSourceExts = []string{".go", ".ts", ".tsx", ".rs"}

// capFileToken extrae la ruta de un token de puntero (`ruta#Símbolo:línea` → `ruta`),
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

func capLines(t *testing.T) (root string, lines []string) {
	t.Helper()
	root = repoRoot()
	// Ruta fija del repo (raíz + nombre constante), no input externo.
	b, err := os.ReadFile(filepath.Join(root, "CAPABILITIES.md")) //nolint:gosec // archivo fijo del repo
	if err != nil {
		t.Fatalf("no se pudo leer CAPABILITIES.md: %v", err)
	}
	return root, strings.Split(string(b), "\n")
}

// TestCapabilityPointersResolve — R1: todo puntero de una línea `- **CAP` existe en el árbol.
func TestCapabilityPointersResolve(t *testing.T) {
	root, lines := capLines(t)
	var bad []string
	for _, ln := range lines {
		if !strings.HasPrefix(strings.TrimSpace(ln), "- **CAP") {
			continue
		}
		for _, m := range capBacktick.FindAllStringSubmatch(ln, -1) {
			p, ok := capFileToken(m[1])
			if !ok {
				continue
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err != nil {
				bad = append(bad, p)
			}
		}
	}
	if len(bad) > 0 {
		t.Fatalf("R1: punteros de CAPABILITIES.md que no resuelven a código real: %v", bad)
	}
}

// TestCapabilityCoverage — R2: ningún archivo fuente queda huérfano de capability.
func TestCapabilityCoverage(t *testing.T) {
	root, lines := capLines(t)
	claimed := map[string]bool{}
	inCoverage := false
	for _, ln := range lines {
		if strings.Contains(ln, "<!--coverage-->") {
			inCoverage = true
		}
		if strings.Contains(ln, "<!--/coverage-->") {
			inCoverage = false
		}
		if !strings.HasPrefix(strings.TrimSpace(ln), "- **CAP") && !inCoverage {
			continue
		}
		for _, m := range capBacktick.FindAllStringSubmatch(ln, -1) {
			if p, ok := capFileToken(m[1]); ok {
				claimed[p] = true
			}
		}
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
		t.Fatalf("R2: %d archivo(s) de código sin capability (reclamar en CAPABILITIES.md o allowlist con razón): %v", len(orphans), orphans)
	}
}
