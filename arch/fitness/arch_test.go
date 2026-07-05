// Package fitness holds ArnesIA's architecture fitness functions — tests that FAIL CI
// when the code violates the boundaries declared in arch/boundaries/.
//
// STATUS (2026-07-05, HS-04): there is no product Go code yet (phase 5). These tests are the
// declared enforcement that matches each boundary node's `enforced_by:`. The import-boundary
// tests below are stdlib-only source scanners — they work the moment the module lands and there
// is nothing to run against until then (they no-op cleanly on a missing tree). The behavioral
// tests skip with a TODO. When the module exists (`go mod init`), drop this file at the repo
// root's `arch/fitness/` and it runs in CI alongside `go-arch-lint check`.
//
// This is the arch-side twin of knowledge/'s 121-check linter: `arnesia conformance` runs
// go-arch-lint + these tests + schema validation + the methodology checks in one severity+signal
// report. See arch/CADENCE.md.
package fitness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test's cwd until it finds a go.mod (the future module root).
// Returns "" if none is found yet (pre-phase-5) so scanners can no-op instead of failing.
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// importsOf returns every import path used by .go files under root/pkgGlob (a path prefix
// relative to the repo root, e.g. "internal/domain"). Missing dirs yield no imports.
func importsOf(t *testing.T, pkgPrefix string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	root := repoRoot()
	if root == "" {
		return out
	}
	base := filepath.Join(root, filepath.FromSlash(pkgPrefix))
	if _, err := os.Stat(base); err != nil {
		return out // package tree doesn't exist yet — nothing to enforce.
	}
	fset := token.NewFileSet()
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("parse %s: %v", path, perr)
			return nil
		}
		for _, imp := range f.Imports {
			out[path] = append(out[path], strings.Trim(imp.Path.Value, `"`))
		}
		return nil
	})
	return out
}

// assertNoImport fails if any file under pkgPrefix imports something matching a forbidden substr.
func assertNoImport(t *testing.T, pkgPrefix string, forbidden []string, boundary string) {
	t.Helper()
	for file, imps := range importsOf(t, pkgPrefix) {
		for _, imp := range imps {
			for _, bad := range forbidden {
				if strings.Contains(imp, bad) {
					t.Errorf("%s imports %q — violates boundary %s", file, imp, boundary)
				}
			}
		}
	}
}

// --- core-no-importa-shell.md ---

func TestCoreHasNoShellImport(t *testing.T) {
	for _, core := range []string{"internal/domain", "internal/usecase", "internal/ports"} {
		assertNoImport(t, core, []string{"/shell", "wailsapp/wails", "tauri"}, "core-no-importa-shell")
	}
}

// --- dominio-independiente-de-transporte.md ---

func TestDomainIndependentOfTransport(t *testing.T) {
	assertNoImport(t, "internal/domain",
		[]string{"net/http", "database/sql", "modernc.org/sqlite", "/transport/", "/sse"},
		"dominio-independiente-de-transporte")
}

// --- indice-desechable-jsonl-es-verdad.md (sin-cgo: DuckDB nunca en el core) ---

func TestNoDuckDBOrCGOStore(t *testing.T) {
	for _, pkg := range []string{"internal/domain", "internal/usecase", "internal/adapters/index"} {
		assertNoImport(t, pkg, []string{"duckdb", "mattn/go-sqlite3"}, "indice-desechable-jsonl-es-verdad")
	}
}

// --- adaptadores-de-agente-intercambiables.md ---

func TestAgentPortHasNoConcreteLeak(t *testing.T) {
	// Fuera del adaptador concreto, nadie importa el paquete claudecode.
	for _, pkg := range []string{"internal/domain", "internal/usecase", "internal/ports",
		"internal/adapters/transport"} {
		assertNoImport(t, pkg, []string{"adapters/agent/claudecode"}, "adaptadores-de-agente-intercambiables")
	}
}

// --- conductor-no-parsea-jsonl.md ---
// El JSONL solo se enumera/replaya; su schema no se decodifica en casos de uso.
// (Chequeo estructural completo requiere el código; placeholder honesto por ahora.)

func TestNoJSONLSchemaParsing(t *testing.T) {
	t.Skip("TODO(fase 5): asegurar que ningún caso de uso json.Unmarshal-ea el transcript interno; " +
		"los eventos vivos vienen de stream-json (ver conductor-no-parsea-jsonl.md).")
}

func TestLiveEventsFromStreamJSON(t *testing.T) {
	t.Skip("TODO(fase 5): el bus de eventos del dock se alimenta del stream-json del conductor, no de tail del JSONL.")
}

// --- indice-desechable-jsonl-es-verdad.md (comportamiento) ---

func TestIndexRebuildsFromJSONL(t *testing.T) {
	t.Skip("TODO(fase 5): borrar el .db y re-indexar produce el mismo estado consultable (índice desechable).")
}

func TestSchemaVersionTriggersRebuild(t *testing.T) {
	t.Skip("TODO(fase 5): mismatch de schema_version borra y reconstruye; no hay migración incremental.")
}

// --- permisos-gui-human-in-the-loop.md ---

func TestNoBypassPermissions(t *testing.T) {
	// Ningún archivo del conductor pasa el flag de bypass (grep estructural sobre el árbol).
	root := repoRoot()
	if root == "" {
		return
	}
	base := filepath.Join(root, "internal", "adapters", "agent")
	if _, err := os.Stat(base); err != nil {
		return
	}
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		if strings.Contains(string(b), "dangerously-skip-permissions") ||
			strings.Contains(string(b), "bypassPermissions") {
			t.Errorf("%s references a permission-bypass flag — violates permisos-gui-human-in-the-loop", path)
		}
		return nil
	})
}

func TestWriteRequiresApproval(t *testing.T) {
	t.Skip("TODO(fase 5): Write/Edit no van en --allowedTools; pasan por el diff-approval del GUI (control_request).")
}

// --- contrato-de-caja-es-fitness-function.md ---

func TestBoxContractValidatesAgainstSchema(t *testing.T) {
	t.Skip("TODO(fase 5): cada contract: de caja valida contra contracts/schema/box.contract.schema.json " +
		"vía google/jsonschema-go — la fitness function del dominio (eval-gate A4, huérfanos, gate honesto).")
}
