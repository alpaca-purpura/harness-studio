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
	"fmt"
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

// capPointerTokens recorre las hojas capability (+ _coverage.yaml) y devuelve los tokens CRUDOS
// del bloque `pointers:`/`support_files:` (sin recortar en `#`) — insumo de la resolución de
// símbolo, que necesita la parte `#Símbolo` que `capClaims` descarta.
func capPointerTokens(t *testing.T) (root string, tokens []string) {
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
					tokens = append(tokens, m[1])
				}
				continue
			}
			inBlock = false
		}
	}
	return root, tokens
}

// TestCapabilityPointerSymbolsResolve — R1 a nivel símbolo (BACKLOG, auditoría 2026-07-14): la
// parte `#Símbolo` de un puntero `file#Símbolo` debe existir DECLARADA en el archivo — no solo
// que el archivo exista (eso ya lo cubre TestCapabilityPointersResolve). Un símbolo renombrado o
// borrado sin actualizar el capability ahora rompe el enforcer en vez de pasar en silencio.
func TestCapabilityPointerSymbolsResolve(t *testing.T) {
	root, tokens := capPointerTokens(t)
	var bad []string
	for _, tok := range tokens {
		file, symbol, ok := capPointerParts(tok)
		if !ok || symbol == "" {
			continue // sin extensión de código fuente, o puntero legítimo sin símbolo (archivo/paquete)
		}
		abs := filepath.Join(root, filepath.FromSlash(file))
		if _, err := os.Stat(abs); err != nil {
			continue // archivo inexistente: ya lo reporta TestCapabilityPointersResolve, no duplicar
		}
		if resolves, why := capSymbolResolves(abs, symbol); !resolves {
			bad = append(bad, fmt.Sprintf("%s#%s (%s)", file, symbol, why))
		}
	}
	if len(bad) > 0 {
		t.Fatalf("R1 símbolo: %d puntero(s) cuyo #Símbolo no resuelve en el archivo:\n  %s", len(bad), strings.Join(bad, "\n  "))
	}
}

// capMeta = frontmatter de una hoja capability, lo mínimo para R4 (estado ⟺ evidencia · puntero estable).
type capMeta struct {
	file        string
	status      string
	validaCount int
	pointers    []string
}

// capMetas parsea el frontmatter de cada hoja capability (línea a línea, mismo estilo que capClaims;
// sin dep de YAML en Go). Extrae `status:`, cuenta las entradas de `valida:` y colecta los `pointers:`.
// Excluye `_coverage.yaml` (no es una capability, no lleva status).
func capMetas(t *testing.T) (metas []capMeta) {
	t.Helper()
	root := repoRoot()
	capDir := filepath.Join(root, "docs", "product", "capabilities")
	var files []string
	werr := filepath.WalkDir(capDir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() && strings.HasSuffix(path, ".yaml") && filepath.Base(path) != "_coverage.yaml" {
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
		rel, _ := filepath.Rel(root, f)
		metas = append(metas, parseCapMeta(filepath.ToSlash(rel), string(b)))
	}
	return metas
}

// parseCapMeta parsea el frontmatter de UNA hoja capability (línea a línea, sin dep de YAML en
// Go). `status:` solo cuenta a columna 0 (clave top-level) — el `status:` anidado bajo
// `scenarios[]` (live/wip/deprecated, otro enum) va indentado y NO debe pisar el status real.
func parseCapMeta(rel, content string) capMeta {
	m := capMeta{file: rel}
	inValida, inPointers := false, false
	for _, ln := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(ln, "status:"): // col 0 = clave top-level; nested (scenarios[].status) va indentado
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "status:"))
			if i := strings.Index(v, "#"); i >= 0 { // corta el comentario inline
				v = strings.TrimSpace(v[:i])
			}
			m.status = v
		case trimmed == "valida:":
			inValida, inPointers = true, false
		case trimmed == "pointers:":
			inPointers, inValida = true, false
		case inValida && strings.HasPrefix(trimmed, "- "):
			m.validaCount++
		case inPointers && strings.HasPrefix(trimmed, "- "):
			if mm := capQuoted.FindStringSubmatch(ln); mm != nil {
				m.pointers = append(m.pointers, mm[1])
			}
		default:
			inValida, inPointers = false, false
		}
	}
	return m
}

// TestParseCapMetaIgnoraStatusAnidadoDeEscenario — regresión (BACKLOG «capMetas no parsea YAML
// real», S1-D16): un scenario BDD con su propio `status: live` (indentado, enum live/wip/deprecated)
// no debe pisar el `status:` root de la capability (enum vivo/vivo·nc/parcial/stub).
func TestParseCapMetaIgnoraStatusAnidadoDeEscenario(t *testing.T) {
	fixture := `---
status: vivo
valida:
  - TestAlgo
pointers:
  - "internal/algo.go#Algo"
scenarios:
  - id: caso-1
    name: "Caso 1"
    status: live
    given: "x"
    when: "y"
    then: "z"
---
`
	m := parseCapMeta("fixture.yaml", fixture)
	if m.status != "vivo" {
		t.Fatalf("status root pisado por scenarios[].status: got %q, want %q", m.status, "vivo")
	}
}

// TestCapabilityStatusConsistent — R4 `cap-estado-generado` (forma determinista): el estado no puede
// CONTRADECIR su evidencia. `vivo`/`parcial` ⟹ ≥1 `valida:` (hay check que respalda); `vivo·nc`/`stub`
// ⟹ 0 `valida:` (por definición sin check). Esto mata el estado fabricado a mano de forma determinista.
// La derivación LIVE (correr cada check y flipear el bit según pase/falle) sigue como cableado CI —
// deuda honesta en BACKLOG; este enforcer garantiza la consistencia, no la frescura del resultado.
func TestCapabilityStatusConsistent(t *testing.T) {
	metas := capMetas(t)
	var bad []string
	for _, m := range metas {
		switch m.status {
		case "vivo", "parcial":
			if m.validaCount == 0 {
				bad = append(bad, fmt.Sprintf("%s: status %q sin `valida:` (evidencia ausente)", m.file, m.status))
			}
		case "vivo·nc", "stub":
			if m.validaCount > 0 {
				bad = append(bad, fmt.Sprintf("%s: status %q con %d `valida:` (contradice: nc/stub = sin check)", m.file, m.status, m.validaCount))
			}
		default:
			bad = append(bad, fmt.Sprintf("%s: status %q fuera de enum {vivo,vivo·nc,parcial,stub}", m.file, m.status))
		}
	}
	if len(bad) > 0 {
		t.Fatalf("R4 estado-consistente: %d cap(s) con estado que contradice su evidencia:\n  %s", len(bad), strings.Join(bad, "\n  "))
	}
}

// TestCapabilityPointersStable — R4 `cap-puntero-estable`: los punteros usan `file#Símbolo` / `paquete/`,
// NUNCA `file:línea` cruda (las líneas se pudren al reformatear el código). Anti-drift determinista.
func TestCapabilityPointersStable(t *testing.T) {
	metas := capMetas(t)
	lineRe := regexp.MustCompile(`\.(go|ts|tsx|rs):\d`)
	var bad []string
	for _, m := range metas {
		for _, p := range m.pointers {
			if lineRe.MatchString(p) {
				bad = append(bad, fmt.Sprintf("%s: %q", m.file, p))
			}
		}
	}
	if len(bad) > 0 {
		t.Fatalf("R4 puntero-estable: %d puntero(s) por número de línea (usar `#Símbolo`):\n  %s", len(bad), strings.Join(bad, "\n  "))
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
