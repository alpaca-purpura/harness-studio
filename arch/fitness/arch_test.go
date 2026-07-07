// Package fitness holds ArnesIA's architecture fitness functions — tests that FAIL CI
// when the code violates the boundaries declared in arch/boundaries/.
//
// STATUS (2026-07-07, HS-08/HS-09): the `arnesia` module is real and these tests run in CI
// against it. Each test matches a boundary node's `enforced_by:`. The import-boundary tests
// are stdlib-only source scanners over the module tree; the doctrine (HS-07/HS-08) and HS-06
// tests exercise the conductor loop, the role-derived permission-sets and the session service
// with fakes — real pass/fail. Only the checks still gated on telemetry (the JSONL indexer,
// fase 5) remain honest t.Skip TODOs.
//
// This is the arch-side twin of the knowledge/ checklists: `arnesia conformance` parses
// knowledge/ (138 checks) + arch/ (97 checks) = 235 checks as data and runs go-arch-lint +
// these tests + schema validation + the methodology checks in one severity+signal report.
// See arch/CADENCE.md.
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

	"github.com/alpacapurpura/arnesia/internal/adapters/permission"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
	"github.com/google/jsonschema-go/jsonschema"
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
	if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // an unreadable entry must fail the scan, not silently narrow it.
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
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
	}); werr != nil {
		t.Errorf("walk %s: %v", base, werr)
	}
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
	for _, pkg := range []string{
		"internal/domain", "internal/usecase", "internal/ports",
		"internal/adapters/transport",
	} {
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
	if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // an unreadable entry must fail the scan, not silently narrow it.
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, rerr := os.ReadFile(path) //nolint:gosec // G304: test-only scanner; path comes from walking the repo's own source tree.
		if rerr != nil {
			return rerr
		}
		if strings.Contains(string(b), "dangerously-skip-permissions") ||
			strings.Contains(string(b), "bypassPermissions") {
			t.Errorf("%s references a permission-bypass flag — violates permisos-gui-human-in-the-loop", path)
		}
		return nil
	}); werr != nil {
		t.Errorf("walk %s: %v", base, werr)
	}
}

func TestWriteRequiresApproval(t *testing.T) {
	t.Skip("TODO(fase 5): Write/Edit no van en --allowedTools; pasan por el diff-approval del GUI (control_request).")
}

// --- contrato-de-caja-es-fitness-function.md ---

// resolveSchema loads a repo-relative JSON Schema and resolves it for validation.
func resolveSchema(t *testing.T, rel string) *jsonschema.Resolved {
	t.Helper()
	raw := readSourceFile(t, rel)
	if raw == "" {
		t.Skipf("schema %s unavailable (pre-module)", rel)
	}
	var s jsonschema.Schema
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("unmarshal schema %s: %v", rel, err)
	}
	res, err := s.Resolve(nil)
	if err != nil {
		t.Fatalf("resolve schema %s: %v", rel, err)
	}
	return res
}

// mustInstance decodes a JSON literal into the generic value validate expects.
func mustInstance(t *testing.T, doc string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatalf("decode instance: %v", err)
	}
	return v
}

// TestBoxContractValidatesAgainstSchema is the fitness function of the fused box
// contract (contrato-de-caja-es-fitness-function.md, eval-gate A4). Validating a
// contract: block against box.contract.schema.json IS the eval-gate: a well-formed
// fused caja with all three axes passes; each malformed one fails on the right axis.
func TestBoxContractValidatesAgainstSchema(t *testing.T) {
	schema := resolveSchema(t, "arch/contracts/schema/box.contract.schema.json")

	// A real fused caja: INTENCIÓN + CLASIFICACIÓN (3 ejes) + CABLEADO + ACEPTACIÓN.
	validFused := `{
		"why": "convertir la conversación en un spec ejecutable",
		"capabilities": [{"id":"CAP-01","what":"emitir spec","success":"spec valida contra schema"}],
		"constraints": ["no inventa requisitos"],
		"non_goals": ["no construye código"],
		"clase": "skill",
		"arquetipo": "excepcion",
		"perfil_harness": "T2",
		"caja": true,
		"fase": "spec",
		"estado": "grill -> spec",
		"necesita": [{"art":"idea del usuario","de":"usuario","requerido":true}],
		"entrega": [{"art":"spec.md","escritor_unico":true}],
		"ruta": [{"a":"build","si":"gate verde"}],
		"gate": {
			"tipo": "auto",
			"detalle": "gherkin ejecutable",
			"aceptacion": [{"given":"un spec","when":"se valida","then":"cumple el schema"}],
			"evidencia": "registro de auditoría emitido"
		},
		"handoff": {"cuando":"no converge en 3 vueltas","a":"humano"}
	}`
	if err := schema.Validate(mustInstance(t, validFused)); err != nil {
		t.Errorf("a well-formed fused caja must validate, got: %v", err)
	}

	// A support skill (caja=false) needs none of the caja-required axes.
	if err := schema.Validate(mustInstance(t, `{"caja": false, "clase": "rule"}`)); err != nil {
		t.Errorf("a support skill (caja=false) must validate, got: %v", err)
	}

	// no-arnesar is legitimately NOT a caja (out of the process graph).
	if err := schema.Validate(mustInstance(t, `{"caja": false, "clase": "skill", "arquetipo": "no-arnesar"}`)); err != nil {
		t.Errorf("no-arnesar (caja=false) must validate, got: %v", err)
	}

	// Each of these MUST fail — the schema has diente on every axis.
	bad := map[string]string{
		"caja sin why (INTENCIÓN)":              `{"caja":true,"clase":"skill","arquetipo":"pipeline","perfil_harness":"T1","fase":"f","estado":"a -> b","gate":{"tipo":"none"}}`,
		"caja sin arquetipo (CLASIFICACIÓN)":    `{"why":"x","caja":true,"clase":"skill","perfil_harness":"T1","fase":"f","estado":"a -> b","gate":{"tipo":"none"}}`,
		"clase legacy 'agente' (CLASIFICACIÓN)": `{"why":"x","caja":true,"clase":"agente","arquetipo":"pipeline","perfil_harness":"T1","fase":"f","estado":"a -> b","gate":{"tipo":"none"}}`,
		"perfil T4 no es caja (CLASIFICACIÓN)":  `{"why":"x","caja":true,"clase":"skill","arquetipo":"pipeline","perfil_harness":"T4","fase":"f","estado":"a -> b","gate":{"tipo":"none"}}`,
		"estado sin patrón X -> Y (CABLEADO)":   `{"why":"x","caja":true,"clase":"skill","arquetipo":"pipeline","perfil_harness":"T1","fase":"f","estado":"listo","gate":{"tipo":"none"}}`,
		"gate.tipo desconocido (ACEPTACIÓN)":    `{"why":"x","caja":true,"clase":"skill","arquetipo":"pipeline","perfil_harness":"T1","fase":"f","estado":"a -> b","gate":{"tipo":"quiza"}}`,
		"clave fantasma (firewall CC-native)":   `{"caja":false,"clase":"skill","sanctum":"PERSONA"}`,
		"no-arnesar no puede ser caja (§8.1)":   `{"why":"x","caja":true,"clase":"skill","arquetipo":"no-arnesar","perfil_harness":"T1","fase":"f","estado":"a -> b","gate":{"tipo":"none"}}`,
	}
	for name, doc := range bad {
		if err := schema.Validate(mustInstance(t, doc)); err == nil {
			t.Errorf("%s: expected schema rejection, got none", name)
		}
	}
}

// ============================================================================
// HS-07/HS-08 · doctrina v1 boundaries (enforced). The conductor loop and the
// permission spike landed: the enforcer named in each boundary's `enforced_by:`
// EXISTS and RUNS for real (no dangling pointer — resolves B4). The tests below
// drive usecase.BoxConductor and permission.KitProvisioner with scripted fakes
// and emit genuine pass/fail — no t.Skip left in this section.
// ============================================================================

// --- orquestacion-determinista-entre-cajas.md (enforced) ---

// scriptedSession emits one scripted `result` event per Send, so the conductor loop
// advances deterministically under test. The channel is buffered so Send never blocks.
type scriptedSession struct {
	events  chan ports.AgentEvent
	results []ports.AgentEvent
	idx     int
	mu      sync.Mutex
	sends   int
}

func (s *scriptedSession) Send(_ context.Context, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sends++
	ev := ports.AgentEvent{Kind: ports.EventResult, Subtype: "success"}
	if s.idx < len(s.results) {
		ev = s.results[s.idx]
		s.idx++
	}
	s.events <- ev
	return nil
}
func (s *scriptedSession) Events() <-chan ports.AgentEvent { return s.events }
func (s *scriptedSession) Close() error                    { return nil }

type scriptedAgent struct{ sess *scriptedSession }

func (a *scriptedAgent) Spawn(_ context.Context, _ ports.SpawnOpts) (ports.AgentSession, error) {
	return a.sess, nil
}

// scriptedArtifacts returns a scripted document-as-cache status per read.
type scriptedArtifacts struct {
	statuses []string
	idx      int
	reads    int
}

func (a *scriptedArtifacts) Status(_ context.Context, _ string) (string, bool, error) {
	st := "working"
	if a.idx < len(a.statuses) {
		st = a.statuses[a.idx]
		a.idx++
	}
	a.reads++
	return st, true, nil
}

func boxWithRuta(ruta []domain.Route, handoff *domain.Handoff) domain.Box {
	return domain.Box{
		ID: "caja-x", Clase: domain.ClaseSkill, Nombre: "caja x",
		Contract: &domain.Contract{
			Why: "hacer x", Clase: domain.ClaseSkill, Arquetipo: domain.ArqExcepcion,
			Perfil: domain.PerfilT3, Caja: true, Fase: "f", Estado: "a -> b",
			Entrega: []domain.Output{{Art: "art-x"}}, Ruta: ruta, Handoff: handoff,
			Gate: &domain.Gate{Tipo: domain.GateAuto},
		},
	}
}

// TestConductorOwnsBoxRouting enforces orquestacion-determinista-entre-cajas: the Go
// conductor — not the LLM — owns the loop and the hand-off. It reads the result subtype +
// the artifact status (document-as-cache), bounds iterations with a repair cap, blocks on
// non-convergence → handoff, and routes via contract.ruta. Crucially, a misleading chat
// text does NOT change the decision.
func TestConductorOwnsBoxRouting(t *testing.T) {
	newConductor := func(results []ports.AgentEvent, statuses []string, repairCap int) (*usecase.BoxConductor, *scriptedSession, *scriptedArtifacts) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, repairCap+2), results: results}
		arts := &scriptedArtifacts{statuses: statuses}
		return usecase.NewBoxConductor(&scriptedAgent{sess: sess}, arts, repairCap, 40), sess, arts
	}

	t.Run("happy path routes by contract.ruta, not the LLM", func(t *testing.T) {
		c, _, _ := newConductor(
			[]ports.AgentEvent{{Kind: ports.EventResult, Subtype: "success"}, {Kind: ports.EventResult, Subtype: "success"}},
			[]string{"working", "done"}, 5)
		out, err := c.Run(context.Background(), boxWithRuta([]domain.Route{{A: "reviewer"}}, nil))
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.Estado != domain.CajaDone {
			t.Errorf("estado = %q, want done", out.Estado)
		}
		if out.Siguiente != "reviewer" {
			t.Errorf("siguiente = %q, want reviewer (code read contract.ruta)", out.Siguiente)
		}
		if out.Iteraciones != 2 {
			t.Errorf("iteraciones = %d, want 2", out.Iteraciones)
		}
	})

	t.Run("repair cap bounds the loop and blocks → handoff", func(t *testing.T) {
		c, sess, _ := newConductor(nil, []string{"working", "working", "working", "working"}, 3)
		out, err := c.Run(context.Background(), boxWithRuta(nil, &domain.Handoff{Cuando: "no converge", A: "humano"}))
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.Estado != domain.CajaBlocked {
			t.Errorf("estado = %q, want blocked (cap hit)", out.Estado)
		}
		if sess.sends != 3 {
			t.Errorf("sends = %d, want exactly 3 (the cap bounds the loop — not unbounded)", sess.sends)
		}
		if !out.Handoff || out.Siguiente != "humano" {
			t.Errorf("blocked must handoff to humano, got handoff=%v siguiente=%q", out.Handoff, out.Siguiente)
		}
	})

	t.Run("reads artifact status, NOT chat text", func(t *testing.T) {
		// The chat text screams success; the artifact status says working. The conductor
		// must ignore the text and NOT finish — this is the anti-scrape guarantee.
		c, _, arts := newConductor(
			[]ports.AgentEvent{{Kind: ports.EventResult, Subtype: "success", Text: "¡LISTO! done done done ✅"}},
			[]string{"working"}, 1)
		out, err := c.Run(context.Background(), boxWithRuta(nil, &domain.Handoff{Cuando: "x", A: "humano"}))
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.Estado == domain.CajaDone {
			t.Error("conductor finished on misleading chat text — it must read the artifact status, not the chat")
		}
		if arts.reads == 0 {
			t.Error("conductor never read the artifact status (document-as-cache)")
		}
	})

	t.Run("explicit blocked artifact stops immediately", func(t *testing.T) {
		c, sess, _ := newConductor(
			[]ports.AgentEvent{{Kind: ports.EventResult, Subtype: "success"}},
			[]string{"blocked"}, 5)
		out, err := c.Run(context.Background(), boxWithRuta(nil, &domain.Handoff{Cuando: "x", A: "caja:fix"}))
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.Estado != domain.CajaBlocked {
			t.Errorf("estado = %q, want blocked", out.Estado)
		}
		if sess.sends != 1 {
			t.Errorf("sends = %d, want 1 (stopped immediately on blocked)", sess.sends)
		}
	})
}

// --- permisos-derivan-del-rol.md (enforced) ---

// TestPermissionSetParametrizedByRole enforces permisos-derivan-del-rol: the permission-set
// is f(role), not a fixed default — the SAME tool carries different decisions per role —
// and grants are task-based / expiring (least temporal privilege), never perpetual.
func TestPermissionSetParametrizedByRole(t *testing.T) {
	k := permission.NewKitProvisioner()
	ctx := context.Background()

	dev, err := k.ResolveForRole(ctx, "backend-dev")
	if err != nil {
		t.Fatalf("resolve dev: %v", err)
	}
	rev, err := k.ResolveForRole(ctx, "reviewer")
	if err != nil {
		t.Fatalf("resolve reviewer: %v", err)
	}

	// The SAME tool must resolve differently by role — permission = f(role), not a default.
	if got := dev.Decide("Write"); got != domain.DecisionAllow {
		t.Errorf("backend-dev Write = %q, want allow", got)
	}
	if got := rev.Decide("Write"); got != domain.DecisionDeny {
		t.Errorf("reviewer Write = %q, want deny (same tool, different role)", got)
	}
	if dev.Decide("Write") == rev.Decide("Write") {
		t.Error("Write resolves identically for dev and reviewer — permission is NOT parametrized by role")
	}

	// Deny-by-default: an unlisted tool needs approval, never a silent allow.
	if got := dev.Decide("SomeUnlistedTool"); got != domain.DecisionAsk {
		t.Errorf("unlisted tool = %q, want ask (deny-by-default)", got)
	}

	// An unknown role degrades to the base deny-by-default set (mutations denied).
	unknown, err := k.ResolveForRole(ctx, "rol-desconocido")
	if err != nil {
		t.Fatalf("resolve unknown: %v", err)
	}
	if got := unknown.Decide("Write"); got != domain.DecisionDeny {
		t.Errorf("unknown role Write = %q, want deny (least privilege fallback)", got)
	}

	// Grants EXPIRE (least temporal privilege) — not perpetual access.
	base := time.Unix(1_700_000_000, 0)
	g := dev.NuevoGrant("Bash", base) // TTL 15m for backend-dev.
	if !g.Vigente(base.Add(14 * time.Minute)) {
		t.Error("grant should be valid within its TTL")
	}
	if g.Vigente(base.Add(16 * time.Minute)) {
		t.Error("grant must EXPIRE past its TTL — no perpetual permission")
	}

	// A per-task role (TTL 0) yields an already-expired grant: valid only for the turn.
	perTask := domain.PermissionSet{Rol: "x", TTL: 0}
	if perTask.NuevoGrant("Bash", base).Vigente(base.Add(time.Nanosecond)) {
		t.Error("per-task grant (TTL 0) must not persist beyond the approving turn")
	}
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
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))) //nolint:gosec // G304: rel is a repo-relative literal written in this test file, never external input.
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
	if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // an unreadable entry must fail the scan, not silently narrow it.
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, rerr := os.ReadFile(path) //nolint:gosec // G304: test-only scanner; path comes from walking the repo's own source tree.
		if rerr != nil {
			return rerr
		}
		if strings.Contains(string(b), `Access-Control-Allow-Origin", "*"`) {
			t.Errorf("%s sets a wildcard CORS origin — violates superficie-local-confinada", path)
		}
		return nil
	}); werr != nil {
		t.Errorf("walk %s: %v", base, werr)
	}
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
	svc, err := usecase.NewSessionService(context.Background(), agent, fakeStore{}, pub, fakeResolver{path: cwd}, 40, nil)
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
