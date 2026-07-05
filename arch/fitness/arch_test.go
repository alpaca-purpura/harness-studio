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
// This is the arch-side twin of knowledge/'s 122-check linter: `arnesia conformance` runs
// go-arch-lint + these tests + schema validation + the methodology checks in one severity+signal
// report. See arch/CADENCE.md.
package fitness

import (
	"context"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
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

// ============================================================================
// HS-06 · superficie-local-confinada + sesion-viva-consistente (enforced)
// These run for real against the module — no skip.
// ============================================================================

// readSourceFile returns a repo-relative file's contents, or "" pre-module (no-op).
func readSourceFile(t *testing.T, rel string) string {
	t.Helper()
	root := repoRoot()
	if root == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// --- superficie-local-confinada.md ---

func TestNoWildcardCORS(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	base := filepath.Join(root, "internal", "adapters", "transport")
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
		if strings.Contains(string(b), `Access-Control-Allow-Origin", "*"`) {
			t.Errorf("%s sets a wildcard CORS origin — violates superficie-local-confinada", path)
		}
		return nil
	})
}

func TestLocalSurfaceConfined(t *testing.T) {
	auth := readSourceFile(t, "internal/adapters/transport/http/auth.go")
	if auth == "" {
		return
	}
	for _, must := range []string{"ConstantTimeCompare", "Host", "Origin", "Access-Control-Allow-Origin"} {
		if !strings.Contains(auth, must) {
			t.Errorf("auth.go missing %q — the confinement gate is incomplete", must)
		}
	}
	router := readSourceFile(t, "internal/adapters/transport/http/router.go")
	if !strings.Contains(router, "withAuth(") {
		t.Errorf("router.go does not wrap the mux in withAuth — surface is unconfined")
	}
	if strings.Contains(router, "withCORS(") {
		t.Errorf("router.go still uses the old open withCORS wrapper")
	}
}

// --- sesion-viva-consistente.md (source scans) ---

func TestNoSilentEventDrop(t *testing.T) {
	cond := readSourceFile(t, "internal/adapters/agent/claudecode/conductor.go")
	if cond == "" {
		return
	}
	i := strings.Index(cond, "func (s *ccSession) emit(")
	if i < 0 {
		t.Fatal("conductor emit not found")
	}
	body := cond[i:min(i+220, len(cond))]
	if !strings.Contains(body, "s.events <- ev") {
		t.Errorf("conductor emit is not a blocking send — frames can be dropped")
	}
	if strings.Contains(body, "default:") {
		t.Errorf("conductor emit still has a select/default drop — silent event loss")
	}
	broker := readSourceFile(t, "internal/adapters/transport/sse/broker.go")
	if !strings.Contains(broker, "disconnect") {
		t.Errorf("broker does not shed lagging subscribers — silent event loss on backpressure")
	}
}

func TestMaxTurnsAlways(t *testing.T) {
	cond := readSourceFile(t, "internal/adapters/agent/claudecode/conductor.go")
	if cond == "" {
		return
	}
	if !strings.Contains(cond, "--max-turns") {
		t.Errorf("conductor does not pass --max-turns — violates permisos-gui max-turns-siempre")
	}
}

// --- sesion-viva-consistente.md + permisos-gui (behavioral, with fakes) ---

type fakeSession struct {
	events chan ports.AgentEvent
	mu     sync.Mutex
	sent   []string
}

func (f *fakeSession) Send(_ context.Context, turn string) error {
	f.mu.Lock()
	f.sent = append(f.sent, turn)
	f.mu.Unlock()
	return nil
}
func (f *fakeSession) Events() <-chan ports.AgentEvent { return f.events }
func (f *fakeSession) Close() error                    { return nil }

type fakeAgent struct {
	mu       sync.Mutex
	spawns   []ports.SpawnOpts
	sessions []*fakeSession
}

func (a *fakeAgent) Spawn(_ context.Context, opts ports.SpawnOpts) (ports.AgentSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.spawns = append(a.spawns, opts)
	s := &fakeSession{events: make(chan ports.AgentEvent, 16)}
	a.sessions = append(a.sessions, s)
	return s, nil
}

type fakeResolver struct{ path string }

func (r fakeResolver) Resolve(string) (string, bool, error) { return r.path, true, nil }

type fakeStore struct{}

func (fakeStore) Load(context.Context) ([]domain.Session, error) { return nil, nil }
func (fakeStore) Save(context.Context, []domain.Session) error   { return nil }

type fakePub struct {
	mu     sync.Mutex
	frames []map[string]any
}

func (p *fakePub) Publish(_ string, data []byte) {
	var m map[string]any
	if json.Unmarshal(data, &m) == nil {
		p.mu.Lock()
		p.frames = append(p.frames, m)
		p.mu.Unlock()
	}
}

func (p *fakePub) snapshot() []map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]map[string]any, len(p.frames))
	copy(out, p.frames)
	return out
}

func newTestService(t *testing.T, agent ports.AgentPort, pub usecase.EventPublisher, cwd string) *usecase.SessionService {
	t.Helper()
	svc, err := usecase.NewSessionService(context.Background(), agent, fakeStore{}, pub, fakeResolver{path: cwd}, 40)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return svc
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestSessionSpawnsInArnesPath(t *testing.T) {
	agent := &fakeAgent{}
	svc := newTestService(t, agent, &fakePub{}, "/tmp/arnes-x")
	id := svc.List()[0].ID
	if err := svc.Turn(id, "hola"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	agent.mu.Lock()
	defer agent.mu.Unlock()
	if len(agent.spawns) != 1 {
		t.Fatalf("want 1 spawn, got %d", len(agent.spawns))
	}
	if got := agent.spawns[0].Cwd; got != "/tmp/arnes-x" {
		t.Errorf("conductor cwd = %q, want the arnés path (never a shared global cwd)", got)
	}
	if got := agent.spawns[0].MaxTurns; got != 40 {
		t.Errorf("MaxTurns = %d, want 40 (cap must propagate)", got)
	}
}

func TestOneTurnAtATime(t *testing.T) {
	agent := &fakeAgent{}
	svc := newTestService(t, agent, &fakePub{}, t.TempDir())
	id := svc.List()[0].ID
	if err := svc.Turn(id, "one"); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	// Status is now streaming (no result event delivered). A second turn must be rejected.
	if err := svc.Turn(id, "two"); !errors.Is(err, usecase.ErrBusy) {
		t.Errorf("second turn while streaming: got %v, want ErrBusy", err)
	}
}

func TestFramesCarryRunID(t *testing.T) {
	agent := &fakeAgent{}
	pub := &fakePub{}
	svc := newTestService(t, agent, pub, t.TempDir())
	id := svc.List()[0].ID
	if err := svc.Turn(id, "hi"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	sess := agent.sessions[0]
	sess.events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc1", Model: "m"}
	sess.events <- ports.AgentEvent{Kind: ports.EventDelta, Text: "x"}
	sess.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "done", CtxPct: 5}
	close(sess.events)

	waitFor(t, 2*time.Second, func() bool {
		for _, f := range pub.snapshot() {
			if f["kind"] == "result" {
				return true
			}
		}
		return false
	})
	for _, f := range pub.snapshot() {
		if rid, ok := f["run_id"].(string); !ok || rid == "" {
			t.Errorf("dock frame kind=%v has no run_id — an SSE replay cannot be deduped", f["kind"])
		}
	}
}

func TestArnesPathContainment(t *testing.T) {
	reg, err := store.NewArnesRegistry(filepath.Join(t.TempDir(), "arneses.json"), filepath.Join(t.TempDir(), "fallback"))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	home, _ := os.UserHomeDir()
	for _, bad := range []string{"/", home, filepath.Join(home, ".claude"), filepath.Join(home, ".ssh"), "relative/dir"} {
		if err := reg.Register("x", bad); err == nil {
			t.Errorf("Register(%q) should be rejected (protected/invalid path)", bad)
		}
	}
	good := t.TempDir()
	if err := reg.Register("x", good); err != nil {
		t.Errorf("Register(%q) should succeed: %v", good, err)
	}
}
