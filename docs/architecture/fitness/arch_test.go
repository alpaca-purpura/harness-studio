// Package fitness holds ArnesIA's architecture fitness functions — tests that FAIL CI
// when the code violates the boundaries declared in docs/architecture/boundaries/.
//
// STATUS (2026-07-07, HS-08/HS-09): the `arnesia` module is real and these tests run in CI
// against it. Each test matches a boundary node's `enforced_by:`. The import-boundary tests
// are stdlib-only source scanners over the module tree; the doctrine (HS-07/HS-08) and HS-06
// tests exercise the conductor loop, the role-derived permission-sets and the session service
// with fakes — real pass/fail. Only the checks still gated on telemetry (the JSONL indexer,
// fase 5) remain honest t.Skip TODOs.
//
// This is the arch-side twin of the docs/architecture/knowledge/ checklists: `arnesia conformance` parses
// docs/architecture/knowledge/ (138 checks) + arch/ (100 checks) = 238 checks as data and runs go-arch-lint +
// these tests + schema validation + the methodology checks in one severity+signal report.
// See docs/architecture/CADENCE.md.
package fitness

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/agent/claudecode"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
	"github.com/alpacapurpura/arnesia/internal/adapters/permission"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
	"github.com/google/jsonschema-go/jsonschema"
	_ "modernc.org/sqlite" // driver "sqlite", solo para el .db crudo de TestSchemaVersionTriggersRebuild
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

// TestDaemonServableHeadless enforces daemon-servable-headless (HS-16, auditoría colateral
// franja-artefactos §8.2): a real smoke test — builds and runs `arnesia serve` as a subprocess
// (not an import scan) and confirms it answers HTTP with NO shell/display involved. The build
// uses the real toolchain env (module cache etc); only the SERVE process gets a throwaway HOME,
// so its provisioning (~/.arnesia, session/arnés registries) never touches the operator's real
// one and the test never re-downloads modules under a fake HOME.
func TestDaemonServableHeadless(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bin := filepath.Join(t.TempDir(), "arnesia-headless-test")
	//nolint:gosec // G204: fixed args (go build -o <tmp> ./cmd/arnesia); bin is this test's own t.TempDir(), never external input.
	build := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/arnesia")
	build.Dir = root
	if out, berr := build.CombinedOutput(); berr != nil {
		t.Fatalf("build arnesia: %v\n%s", berr, out)
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve a free port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close() // released for the subprocess; a small TOCTOU race is accepted here.

	//nolint:gosec // G204: bin is the binary this test just built to its own t.TempDir(), never external input.
	cmd := exec.CommandContext(ctx, bin, "serve", "--addr", addr)
	cmd.Env = []string{"HOME=" + t.TempDir(), "PATH=" + os.Getenv("PATH")} // headless: no shell, no display, no DISPLAY var at all.
	if serr := cmd.Start(); serr != nil {
		t.Fatalf("start serve: %v", serr)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()

	url := "http://" + addr + "/healthz"
	deadline := time.Now().Add(20 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if rerr != nil {
			t.Fatalf("build healthz request: %v", rerr)
		}
		resp, gerr := http.DefaultClient.Do(req)
		if gerr == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return // headless boot confirmed — a shell is never required to serve.
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = gerr
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("serve never answered /healthz headless: %v", lastErr)
}

// TestMintEnvInLauncher enforces mint-env-en-launcher: the Tauri launcher must set
// WEBKIT_DISABLE_DMABUF_RENDERER before the WebView boots — the Mint mitigation is
// load-bearing for the Tauri-desde-v1 fork (HS-04's declared divergence from browser-first).
func TestMintEnvInLauncher(t *testing.T) {
	main := readSourceFile(t, "web/src-tauri/src/main.rs")
	if main == "" {
		return
	}
	if !strings.Contains(main, "WEBKIT_DISABLE_DMABUF_RENDERER") {
		t.Error("web/src-tauri/src/main.rs no setea WEBKIT_DISABLE_DMABUF_RENDERER — mitigación Mint ausente")
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

// minimalAgent is a second, from-scratch ports.AgentPort — proof the seam is real: nothing in
// usecase needs claudecode-specific behavior to drive a turn.
type minimalAgent struct{}

func (minimalAgent) Spawn(context.Context, ports.SpawnOpts) (ports.AgentSession, error) {
	events := make(chan ports.AgentEvent, 1)
	events <- ports.AgentEvent{Kind: ports.EventResult, Subtype: "success"}
	return &minimalSession{events: events}, nil
}

type minimalSession struct{ events chan ports.AgentEvent }

func (s *minimalSession) Send(context.Context, string) error { return nil }
func (s *minimalSession) Events() <-chan ports.AgentEvent    { return s.events }
func (s *minimalSession) Interrupt(context.Context) error    { return nil }
func (s *minimalSession) Close() error                       { return nil }
func (s *minimalSession) RespondControl(context.Context, string, ports.ControlDecision) error {
	return nil
}

// TestSegundoAdaptadorPosible enforces segundo-adaptador-posible: a second, unrelated
// ports.AgentPort implementation (not claudecode) drives a real usecase turn with zero
// changes to usecase code — the interface seam is real, not aspirational.
func TestSegundoAdaptadorPosible(t *testing.T) {
	svc := newTestService(t, minimalAgent{}, &fakePub{}, t.TempDir())
	id := svc.List()[0].ID
	if err := svc.Turn(id, "hola desde un segundo adaptador"); err != nil {
		t.Fatalf("un segundo ports.AgentPort (no claudecode) debe drivear un turno sin cambios en usecase: %v", err)
	}
}

// --- conductor-no-parsea-jsonl.md ---

// jsonlExento lista, POR RUTA y con la razón escrita, los paquetes a los que este check NO
// aplica. Es una lista por prefijo de ruta, no un patrón: agregar un segundo paquete exige
// tocar este test, que es exactamente el punto.
var jsonlExento = map[string]string{
	// internal/adapters/history es EL lector de replay del historial B2 (RF-200/201): su
	// trabajo ES leer el JSONL de ~/.claude para reconstruir conversaciones cerradas. Ya está
	// declarado así en .go-arch-lint.yml (`history: mayDependOn: [domain]`) y en el propio
	// boundary conductor-no-parsea-jsonl.md, que separa «enumerar/replayar» de «tratar el
	// schema como contrato de los eventos vivos».
	"internal/adapters/history": "lector de replay del historial B2 — enumerar/replayar es su oficio",
}

// jsonlDecoders son las llamadas de decodificación que el scan busca.
var jsonlDecoders = []string{"json.Unmarshal", "json.NewDecoder", "json.RawMessage"}

// jsonlPistas son las marcas que convierten una decodificación cualquiera en «decodificación
// del schema del transcript». Se buscan en la MISMA línea y en la línea previa (el comentario
// que la encabeza), que es donde vive el contexto en este árbol.
var jsonlPistas = []string{"transcript", "jsonl", "~/.claude/projects", ".claude/projects"}

// escaneaJSONLSchema recorre las rutas dadas y devuelve los hallazgos «archivo:línea».
// Se expone como función (y no inline en el test) para poder correrla contra un fixture en
// memoria: sin eso, un scanner roto que no encuentra nada daría verde por vacío.
func escaneaJSONLSchema(t *testing.T, fuentes map[string]string) []string {
	t.Helper()
	var hallazgos []string
	for ruta, src := range fuentes {
		lineas := strings.Split(src, "\n")
		for i, ln := range lineas {
			bajo := strings.ToLower(ln)
			tieneDecoder := false
			for _, d := range jsonlDecoders {
				if strings.Contains(ln, d) {
					tieneDecoder = true
					break
				}
			}
			if !tieneDecoder {
				continue
			}
			contexto := bajo
			if i > 0 {
				contexto = strings.ToLower(lineas[i-1]) + "\n" + bajo
			}
			for _, p := range jsonlPistas {
				if strings.Contains(contexto, p) {
					hallazgos = append(hallazgos, fmt.Sprintf("%s:%d: %s", ruta, i+1, strings.TrimSpace(ln)))
					break
				}
			}
		}
	}
	slices.Sort(hallazgos)
	return hallazgos
}

// TestNoJSONLSchemaParsing enforces conductor-no-parsea-jsonl: el JSONL de ~/.claude se
// ENUMERA y se REPLAYA (eso es del lector de historial), pero su schema no se decodifica como
// contrato de los eventos vivos — esos salen del stream-json del subproceso conductor, y el
// adaptador claudecode es su único dueño.
//
// Dejó de ser un t.Skip con TODO (deuda D8/D14.4): un check que no corre es un pass fabricado
// con otro nombre. Ahora es un source-scan con exención declarada por ruta + control positivo.
func TestNoJSONLSchemaParsing(t *testing.T) {
	// ── Control positivo (§14 del plan: un negativo sin control no es un resultado) ──
	// Un fixture en memoria con una llamada que el detector SÍ debe marcar. Sin esto, un
	// scanner roto (regex mal, walk vacío) daría verde con `len(hallazgos) == 0` sobre el
	// árbol real y nadie lo notaría.
	control := escaneaJSONLSchema(t, map[string]string{
		"fixture/positivo.go": "func leer() {\n\t// decodifica el transcript de la sesión\n\tjson.Unmarshal(raw, &t)\n}",
		"fixture/negativo.go": "func otro() {\n\t// frame stream-json del subproceso\n\tjson.Unmarshal(line, &f)\n}",
	})
	if len(control) != 1 {
		t.Fatalf("control positivo: el scanner debería marcar exactamente 1 llamada del fixture, marcó %d (%v) — el detector está roto y su verde sobre el árbol real no significaría nada", len(control), control)
	}
	if !strings.Contains(control[0], "fixture/positivo.go") {
		t.Fatalf("control positivo: el scanner marcó la línea equivocada: %v", control[0])
	}

	// ── El scan real ──
	root := repoRoot()
	if root == "" {
		t.Fatal("repoRoot vacío: el scan no puede correr y no puede pasar por omisión")
	}
	fuentes := map[string]string{}
	for _, prefijo := range []string{"internal/usecase", "internal/domain", "internal/adapters/telemetria"} {
		base := filepath.Join(root, filepath.FromSlash(prefijo))
		if _, err := os.Stat(base); err != nil {
			continue // el árbol de telemetría puede no existir todavía; los otros dos sí.
		}
		if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return rerr
			}
			rel = filepath.ToSlash(rel)
			for exento := range jsonlExento {
				if strings.HasPrefix(rel, exento) {
					return nil
				}
			}
			b, oerr := os.ReadFile(path) //nolint:gosec // ruta derivada del propio árbol del repo.
			if oerr != nil {
				return oerr
			}
			fuentes[rel] = string(b)
			return nil
		}); werr != nil {
			t.Fatalf("walk %s: %v", base, werr)
		}
	}
	if len(fuentes) == 0 {
		t.Fatal("el scan no leyó ningún archivo: sin corpus, su verde sería vacío, no una garantía")
	}
	for _, h := range escaneaJSONLSchema(t, fuentes) {
		t.Errorf("%s — decodifica el schema del transcript JSONL fuera del lector de historial; "+
			"viola conductor-no-parsea-jsonl.md (los eventos vivos salen del stream-json). "+
			"Exenciones declaradas hoy: %v", h, jsonlExento)
	}
}

// TestLiveEventsFromStreamJSON (real desde Fase E): los eventos vivos salen del
// stream-json del SUBPROCESO conductor — nunca de tail del JSONL. Un binario fake emite
// frames stream-json reales por stdout; el adaptador claudecode debe traducirlos a los
// eventos normalizados que alimentan el Dock, y responder el control_request por stdin
// (el canal de vuelta del human-in-the-loop). Sin `claude` real: el protocolo es lo
// que se prueba.
func TestLiveEventsFromStreamJSON(t *testing.T) {
	dir := t.TempDir()
	stdinCopy := filepath.Join(dir, "stdin-recibido.ndjson")
	frames := []string{
		`{"type":"system","subtype":"init","session_id":"cc-fit","model":"claude-x"}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"hola"}}}`,
		`{"type":"control_request","request_id":"req-9","request":{"subtype":"can_use_tool","tool_name":"Write","input":{"file_path":"a.md"}}}`,
		`{"type":"result","subtype":"success","result":"ok"}`,
	}
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	for _, f := range frames {
		script.WriteString("echo '" + f + "'\n")
	}
	script.WriteString("cat > " + stdinCopy + "\n")
	bin := filepath.Join(dir, "claude-fake")
	if err := os.WriteFile(bin, []byte(script.String()), 0o700); err != nil { //nolint:gosec // G306: test fixture must be executable.
		t.Fatalf("write fake binary: %v", err)
	}

	sess, err := claudecode.New(bin).Spawn(context.Background(), ports.SpawnOpts{MaxTurns: 3})
	if err != nil {
		t.Fatalf("spawn fake: %v", err)
	}

	got := map[ports.AgentEventKind]ports.AgentEvent{}
	deadline := time.After(5 * time.Second)
	for got[ports.EventResult].Kind == "" {
		select {
		case ev, ok := <-sess.Events():
			if !ok {
				t.Fatal("el stream cerró antes del result")
			}
			if _, seen := got[ev.Kind]; !seen {
				got[ev.Kind] = ev
			}
		case <-deadline:
			t.Fatal("timeout esperando los frames stream-json del subproceso")
		}
	}

	if init := got[ports.EventInit]; init.ClaudeSessionID != "cc-fit" || init.Model != "claude-x" {
		t.Errorf("init = %+v, want session cc-fit / model claude-x (del frame system/init)", init)
	}
	if delta := got[ports.EventDelta]; delta.Text != "hola" {
		t.Errorf("delta.Text = %q, want hola (del stream_event, no del JSONL)", delta.Text)
	}
	ctrl := got[ports.EventControlRequest]
	if ctrl.RequestID != "req-9" || ctrl.Tool != "Write" {
		t.Errorf("control_request = %+v, want req-9/Write reenviado (no descartado)", ctrl)
	}

	// La vuelta: responder el control_request viaja por stdin como control_response.
	if rerr := sess.RespondControl(context.Background(), "req-9", ports.ControlDecision{Allow: true, UpdatedInput: ctrl.Input}); rerr != nil {
		t.Fatalf("respond control: %v", rerr)
	}
	_ = sess.Close() // cierra stdin → el fake persiste lo recibido y termina.

	stdin, err := os.ReadFile(stdinCopy) //nolint:gosec // G304: fixture path built in this test.
	if err != nil {
		t.Fatalf("leer stdin capturado: %v", err)
	}
	for _, must := range []string{`"type":"control_response"`, `"request_id":"req-9"`, `"behavior":"allow"`} {
		if !strings.Contains(string(stdin), must) {
			t.Errorf("el stdin del conductor no lleva %s — got %q", must, stdin)
		}
	}
}

// --- indice-desechable-jsonl-es-verdad.md (comportamiento) ---

// regDogfood is a minimal ports.ArnesRegistry fake pointing at the fábrica's own real
// dogfood arnés on disk (dogfood/dev-full-cycle) — these boundary-level tests exercise
// the REAL production chain (index.New + loader.LoadArnes) over real fixture data, not
// a synthetic stub; the exhaustive Rebuild scenario coverage with fakes lives next to
// the code in internal/adapters/index/store_test.go.
type regDogfood []ports.ArnesPath

func (r regDogfood) Resolve(string) (string, bool, error) {
	return "", false, errors.New("regDogfood: Resolve no implementado")
}

func (r regDogfood) Register(string, string) error {
	return errors.New("regDogfood: Register no implementado")
}

func (r regDogfood) List() []ports.ArnesPath { return r }

// TestIndexRebuildsFromJSONL enforces index-reconstruible (RF-207/RF-211): deleting the
// .db file entirely and re-indexing from ArnesRegistry reproduces the exact same
// queryable state — the index is a disposable projection, never a second original.
func TestIndexRebuildsFromJSONL(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	reg := regDogfood{{Arnes: "dev-full-cycle", Path: filepath.Join(root, "dogfood", "dev-full-cycle")}}
	dbPath := filepath.Join(t.TempDir(), "index.db")

	idx, err := index.New(dbPath, reg, loader.LoadArnes)
	if err != nil {
		t.Fatalf("index.New: %v", err)
	}
	if err = idx.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	before, err := idx.Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query antes de borrar el .db: %v", err)
	}
	if err = idx.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// El corazón del check: borrar el .db entero (no un truncate, no una migración) y
	// re-indexar desde la fuente reproduce el MISMO estado consultable.
	if err = os.Remove(dbPath); err != nil {
		t.Fatalf("borrar .db: %v", err)
	}
	idx2, err := index.New(dbPath, reg, loader.LoadArnes)
	if err != nil {
		t.Fatalf("index.New tras borrar el .db: %v", err)
	}
	defer func() { _ = idx2.Close() }()
	if err = idx2.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild tras borrar el .db: %v", err)
	}
	after, err := idx2.Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query tras re-indexar: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("el estado consultable cambió tras borrar+re-indexar el .db:\n antes:   %+v\n después: %+v", before, after)
	}
}

// TestSchemaVersionTriggersRebuild enforces sin-migracion-incremental (RF-208): a .db
// whose schema_meta.version disagrees with the binary's is discarded WHOLE (no ALTER
// TABLE, no row migration) and recreated fresh — proven against the real production
// chain, same as TestIndexRebuildsFromJSONL above.
func TestSchemaVersionTriggersRebuild(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	dbPath := filepath.Join(t.TempDir(), "index.db")

	// Simula un .db "viejo": schema_meta con una versión que el binario actual ya no
	// reconoce, más una fila fantasma que NO debe sobrevivir (nunca se migra una fila).
	raw, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("abrir .db crudo: %v", err)
	}
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS schema_meta (version INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS graphs (clave TEXT PRIMARY KEY, graph_json TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`INSERT INTO schema_meta (version) VALUES (-1)`, // -1: nunca es una schema_version real del binario.
		`INSERT INTO graphs (clave, graph_json, updated_at) VALUES ('fantasma', '{"arnes":{"id":"fantasma"}}', 'x')`,
	} {
		if _, eerr := raw.ExecContext(context.Background(), stmt); eerr != nil {
			t.Fatalf("seed .db viejo (%s): %v", stmt, eerr)
		}
	}
	if cerr := raw.Close(); cerr != nil {
		t.Fatalf("close raw: %v", cerr)
	}

	reg := regDogfood{{Arnes: "dev-full-cycle", Path: filepath.Join(root, "dogfood", "dev-full-cycle")}}
	idx, err := index.New(dbPath, reg, loader.LoadArnes)
	if err != nil {
		t.Fatalf("index.New sobre .db con schema_version vieja: %v", err)
	}
	defer func() { _ = idx.Close() }()

	if _, qerr := idx.Query(context.Background(), "fantasma"); !errors.Is(qerr, index.ErrNotFound) {
		t.Errorf("Query(fantasma) tras el mismatch = %v, want ErrNotFound (nunca migra, borra entero)", qerr)
	}
	if rerr := idx.Rebuild(context.Background()); rerr != nil {
		t.Fatalf("Rebuild sobre el .db recreado: %v", rerr)
	}
	if _, qerr := idx.Query(context.Background(), "dev-full-cycle"); qerr != nil {
		t.Errorf("Query(dev-full-cycle) tras Rebuild post-wipe = %v, want nil", qerr)
	}
}

// TestWriterSerializedSingleConn enforces writer-serializado (RF-209): el handle de
// escritura configura SetMaxOpenConns(1) + WAL + busy_timeout=5000 — chequeo
// estructural; la prueba COMPORTAMENTAL (Upserts concurrentes que nunca devuelven
// SQLITE_BUSY) vive junto al código en
// internal/adapters/index/store_test.go:TestWriterSerializedConcurrentUpsertsSucceed.
func TestWriterSerializedSingleConn(t *testing.T) {
	src := readSourceFile(t, filepath.Join("internal", "adapters", "index", "store.go"))
	if src == "" {
		return
	}
	for _, must := range []string{
		"writer.SetMaxOpenConns(1)",
		"journal_mode(WAL)",
		"busy_timeout(5000)",
	} {
		if !strings.Contains(src, must) {
			t.Errorf("internal/adapters/index/store.go no configura %q — writer sin serializar (RF-209)", must)
		}
	}
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

// TestWriteRequiresApproval (real desde Fase E) — write-requiere-aprobacion: Write/Edit
// jamás se pre-aprueban en --allowedTools (aunque el rol los permita, «allow del rol» =
// aprobable, no automático) y cada escritura pasa por el diff-approval del GUI: el
// control_request llega como tarjeta al Dock, la autoridad del rol manda (deny del rol
// gana al click humano) y el allow mintea un grant EFÍMERO que evita re-preguntar solo
// mientras está vigente.
func TestWriteRequiresApproval(t *testing.T) {
	// (1) Flags del spawn: la materialización CC-native del permission-set.
	k := permission.NewKitProvisioner()
	dev, err := k.ResolveForRole(context.Background(), "backend-dev")
	if err != nil {
		t.Fatalf("resolve backend-dev: %v", err)
	}
	args := claudecode.SpawnArgs(ports.SpawnOpts{MaxTurns: 40, Permisos: dev})
	joined := " " + strings.Join(args, " ") + " "
	for _, must := range []string{" --permission-mode default ", " --permission-prompt-tool stdio ", " --max-turns 40 "} {
		if !strings.Contains(joined, must) {
			t.Errorf("spawn args sin %q — got %q", must, joined)
		}
	}
	allowed := ""
	for i, a := range args {
		if a == "--allowedTools" && i+1 < len(args) {
			allowed = args[i+1]
		}
	}
	for _, writeTool := range []string{"Write", "Edit"} {
		if strings.Contains(allowed, writeTool) {
			t.Errorf("--allowedTools %q pre-aprueba %s — la escritura debe pasar por el diff-approval del GUI", allowed, writeTool)
		}
	}

	// (2) El loop humano end-to-end con fakes (sin claude real).
	agent := &fakeAgent{}
	pub := &fakePub{}
	svc := newTestService(t, agent, pub, t.TempDir())
	id := svc.List()[0].ID
	if terr := svc.Turn(id, "edita el spec"); terr != nil {
		t.Fatalf("turn: %v", terr)
	}
	sess := agent.sessions[0]

	askFrame := func(reqID string) func() bool {
		return func() bool {
			for _, f := range pub.snapshot() {
				if f["kind"] == "permission" && f["request_id"] == reqID {
					return true
				}
			}
			return false
		}
	}

	// Un Write pide permiso → tarjeta `permission` en el Dock (ask→UI, D3).
	sess.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-1", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`)}
	waitFor(t, 2*time.Second, askFrame("cr-1"))

	// El deny del ROL gana al click humano: reviewer deniega Write aunque se apruebe.
	res, err := svc.ResolvePermission(id, "cr-1", "allow", "reviewer", 0, nil)
	if err != nil {
		t.Fatalf("resolve cr-1: %v", err)
	}
	if res.Efectiva != domain.DecisionDeny {
		t.Errorf("reviewer + click allow ⇒ efectiva %q, want deny (la autoridad del rol se impone)", res.Efectiva)
	}
	if got := sess.respondedSnapshot(); len(got) != 1 || got[0].requestID != "cr-1" || got[0].decision.Allow {
		t.Errorf("control_response = %+v, want deny de cr-1 hacia el conductor", got)
	}

	// backend-dev SÍ puede aprobar Write — con grant que EXPIRA (nunca perpetuo).
	sess.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-2", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`)}
	waitFor(t, 2*time.Second, askFrame("cr-2"))
	res, err = svc.ResolvePermission(id, "cr-2", "allow", "backend-dev", time.Minute, nil)
	if err != nil {
		t.Fatalf("resolve cr-2: %v", err)
	}
	if res.Efectiva != domain.DecisionAllow || res.Expira == nil {
		t.Fatalf("resolución = %+v, want allow con expiración (grant efímero)", res)
	}
	if remaining := time.Until(*res.Expira); remaining > time.Minute+time.Second {
		t.Errorf("el ttl del operador debía ACOTAR el grant a ≤1m, expira en %v", remaining)
	}

	// Mientras el grant está vigente, el MISMO tool no re-pregunta: auto-allow sin tarjeta.
	sess.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-3", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`)}
	waitFor(t, 3*time.Second, func() bool { return len(sess.respondedSnapshot()) == 3 })
	if got := sess.respondedSnapshot(); !got[2].decision.Allow {
		t.Errorf("con grant vigente el Write debía auto-aprobarse, got %+v", got[2])
	}
	if askFrame("cr-3")() {
		t.Error("cr-3 generó tarjeta de ask pese al grant vigente — re-pregunta de más")
	}
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
	schema := resolveSchema(t, "docs/architecture/contracts/schema/box.contract.schema.json")

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

// TestDogfoodComposicionFabricaConforma wires sin-huerfanos/dead-end/ruta-a-existe/
// refina-coherente into `--todo` (HS-16, auditoría colateral franja-artefactos §8.2): these
// enforcers already exist and pass for real under `--arnes` (domain.VerificarX, franja-artefactos
// F1) but the ruleset parser only recognizes the `arch_test.go:TestX` pattern in a check's
// `enforcer` column — the plain-prose form fell to nl-judge/deferred even though nothing was
// missing. No new validation logic: this reuses the same ConformanceService against the
// fábrica's own dogfood (the same pattern as usecase.TestDogfoodArnesConforms) so the fábrica
// proves it satisfies its own composition rules, not just third-party arneses.
func TestDogfoodComposicionFabricaConforma(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	adapters := []ports.MechanismAdapter{
		mechanism.NewArchTest(root), mechanism.NewGoArchLint(root),
		mechanism.NLJudge{},
		mechanism.StaticScan{},
		mechanism.SchemaAdapter{},
	}
	schemas := mechanism.NewSchemaSet(filepath.Join(root, "docs", "architecture", "contracts", "schema"))
	svc := usecase.NewConformanceService(root, ruleset.New(root), schemas, adapters)
	rep, err := svc.Run(context.Background(), ports.Target{
		Kind: ports.TargetArnes, GraphPath: filepath.Join(root, "dogfood", "dev-full-cycle.graph.json"),
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []string{"sin-huerfanos", "dead-end", "ruta-a-existe", "refina-coherente"}
	found := map[string]bool{}
	for _, r := range rep.Results {
		if !slices.Contains(want, r.Check.ID) {
			continue
		}
		found[r.Check.ID] = true
		if r.Veredicto != domain.VeredictoPass {
			t.Errorf("%s: %s — %s (el dogfood de la fábrica debe cumplir su propia composición)",
				r.Check.ID, r.Veredicto, r.Detalle)
		}
	}
	for _, id := range want {
		if !found[id] {
			t.Errorf("check %q ausente del reporte --arnes", id)
		}
	}
}

// TestGoArchLintAdapterCorreDeVerdad (HS-27, barrido de deuda viva 2026-07-24): el adapter
// GoArchLint probaba `exec.LookPath("go-arch-lint")` antes de correr — un binario que NUNCA
// está en el PATH ni en dev ni en CI (ambos invocan `go run github.com/fe3dback/go-arch-lint@
// latest`, ver ci.yml step "go-arch-lint") — así que el motor difería localmente 3 checks que
// CI corre y hace cumplir de verdad (agent-port-existe · adapter-solo-en-root ·
// domain-no-http, los únicos 3 checks del ruleset cuyo enforcer resuelve al mecanismo
// go-arch-lint puro, sin cita `arch_test.go` que los desvíe a ArchTest primero). Corregido el
// adapter para invocar EXACTAMENTE el mismo comando que ci.yml.
//
// Scope `fabrica` (no `arnes`): estos 3 checks auditan el grafo de imports de ArnesIA MISMA
// (`internal/domain` no importa `net/http`, etc.) — no tienen nada que ver con un arnés
// target, así que corren bajo `Target{Kind: TargetTodo}` (el sweep que usa `--todo`), no
// `TargetArnes` (ese va por `RunGraph`, el set hardcodeado de 21 checks — ver runArnes — que
// NUNCA pasa por el ruteo por-mecanismo del ruleset; sin-huerfanos/dead-end/etc de
// TestDogfoodComposicionFabricaConforma pasan ahí porque RunGraph los construye a mano con el
// MISMO id, no porque el ruteo los alcance).
func TestGoArchLintAdapterCorreDeVerdad(t *testing.T) {
	root := repoRoot()
	if root == "" {
		return
	}
	adapters := []ports.MechanismAdapter{
		mechanism.NewArchTest(root), mechanism.NewGoArchLint(root),
		mechanism.NLJudge{},
		mechanism.StaticScan{},
		mechanism.SchemaAdapter{},
	}
	schemas := mechanism.NewSchemaSet(filepath.Join(root, "docs", "architecture", "contracts", "schema"))
	svc := usecase.NewConformanceService(root, ruleset.New(root), schemas, adapters)
	rep, err := svc.Run(context.Background(), ports.Target{Kind: ports.TargetTodo})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []string{"agent-port-existe", "adapter-solo-en-root", "domain-no-http"}
	found := map[string]bool{}
	for _, r := range rep.Results {
		if !slices.Contains(want, r.Check.ID) {
			continue
		}
		found[r.Check.ID] = true
		if r.Check.Mecanismo != domain.MecGoArchLint {
			t.Errorf("%s: mecanismo = %s, quiero go-arch-lint (¿cambió el enforcer del .md?)", r.Check.ID, r.Check.Mecanismo)
		}
		if r.Veredicto != domain.VeredictoPass {
			t.Errorf("%s: %s — %s (el propio árbol de ArnesIA debe cumplir su grafo de imports)",
				r.Check.ID, r.Veredicto, r.Detalle)
		}
	}
	for _, id := range want {
		if !found[id] {
			t.Errorf("check %q ausente del reporte --todo", id)
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
	prompts []string
}

func (s *scriptedSession) Send(_ context.Context, prompt string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sends++
	s.prompts = append(s.prompts, prompt)
	ev := ports.AgentEvent{Kind: ports.EventResult, Subtype: "success"}
	if s.idx < len(s.results) {
		ev = s.results[s.idx]
		s.idx++
	}
	s.events <- ev
	return nil
}
func (s *scriptedSession) Events() <-chan ports.AgentEvent { return s.events }
func (s *scriptedSession) Interrupt(context.Context) error { return nil }
func (s *scriptedSession) Close() error                    { return nil }
func (s *scriptedSession) RespondControl(context.Context, string, ports.ControlDecision) error {
	return nil
}

type scriptedAgent struct {
	sess   *scriptedSession
	spawns int
}

func (a *scriptedAgent) Spawn(_ context.Context, _ ports.SpawnOpts) (ports.AgentSession, error) {
	a.spawns++
	return a.sess, nil
}

// scriptedArtifacts returns a scripted document-as-cache status per read.
type scriptedArtifacts struct {
	statuses []string
	idx      int
	reads    int
}

func (a *scriptedArtifacts) Status(_ context.Context, _, _ string) (string, bool, error) {
	st := "working"
	if a.idx < len(a.statuses) {
		st = a.statuses[a.idx]
		a.idx++
	}
	a.reads++
	return st, true, nil
}

func (a *scriptedArtifacts) Resumen(context.Context, string, string) (string, error) {
	return "", nil
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
			[]string{"working", "done"}, 5,
		)
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
			[]string{"working"}, 1,
		)
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
			[]string{"blocked"}, 5,
		)
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

// trackingArtifacts records the ref the conductor reads and can fail the read.
type trackingArtifacts struct {
	lastRef string
	err     error
}

func (a *trackingArtifacts) Status(_ context.Context, _, ref string) (string, bool, error) {
	a.lastRef = ref
	if a.err != nil {
		return "", false, a.err
	}
	return "done", true, nil
}

func (a *trackingArtifacts) Resumen(context.Context, string, string) (string, error) {
	return "", nil
}

// TestConductorArtifactIdentity enforces RF-111 (franja-artefactos D3): the conductor
// reads the entrega's declared `path` when present (artefacto-archivo wins over the art
// label) and an artifact-status read error is NEVER discarded silently — it surfaces in
// the outcome's Advertencias without breaking the loop.
func TestConductorArtifactIdentity(t *testing.T) {
	t.Run("entrega.path wins over the art label", func(t *testing.T) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, 3)}
		arts := &trackingArtifacts{}
		c := usecase.NewBoxConductor(&scriptedAgent{sess: sess}, arts, 3, 40)
		box := boxWithRuta(nil, nil)
		box.Contract.Entrega = []domain.Output{{Art: "spec.md", Path: "docs/spec.md"}}
		if _, err := c.Run(context.Background(), box); err != nil {
			t.Fatalf("run: %v", err)
		}
		if arts.lastRef != "docs/spec.md" {
			t.Errorf("conductor leyó %q, want docs/spec.md (entrega.path manda sobre art)", arts.lastRef)
		}
	})

	t.Run("status read error is visible, never silent", func(t *testing.T) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, 3)}
		arts := &trackingArtifacts{err: errors.New("frontmatter roto")}
		c := usecase.NewBoxConductor(&scriptedAgent{sess: sess}, arts, 1, 40)
		out, err := c.Run(context.Background(), boxWithRuta(nil, &domain.Handoff{Cuando: "x", A: "humano"}))
		if err != nil {
			t.Fatalf("run: %v (el error de status NO debe romper el loop)", err)
		}
		if len(out.Advertencias) == 0 {
			t.Fatal("error de artifacts.Status descartado en silencio — debe viajar en Advertencias (RF-111)")
		}
		if !strings.Contains(out.Advertencias[0], "frontmatter roto") {
			t.Errorf("advertencia %q no conserva la causa", out.Advertencias[0])
		}
	})
}

// fsArtifacts scripts a per-path filesystem view: existence + digest per artifact ref.
type fsArtifacts struct {
	existe  map[string]bool
	resumen map[string]string
}

func (a *fsArtifacts) Status(_ context.Context, _, ref string) (string, bool, error) {
	return "working", a.existe[ref], nil
}

func (a *fsArtifacts) Resumen(_ context.Context, _, ref string) (string, error) {
	return a.resumen[ref], nil
}

// TestConductorEncadenaPorFilesystem enforces D7 (franja-artefactos, RF-120/121): the
// hand-off travels by filesystem — the conductor stats the REQUIRED inputs before any
// spawn (missing → the run never starts, zero tokens) and tarea() cites rutas + digest,
// never the document body.
func TestConductorEncadenaPorFilesystem(t *testing.T) {
	boxConInsumos := func() domain.Box {
		b := boxWithRuta(nil, &domain.Handoff{Cuando: "x", A: "humano"})
		b.Contract.Necesita = []domain.Input{{Art: "spec.md", De: "caja:spec-writer"}}
		return b
	}
	insumos := []domain.Insumo{{
		Art: "spec.md", De: "caja:spec-writer", Productor: "spec-writer",
		Path: "spec.md", Requerido: true,
	}}

	t.Run("precondición incumplida: no spawnea, cero tokens", func(t *testing.T) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, 3)}
		agent := &scriptedAgent{sess: sess}
		arts := &fsArtifacts{existe: map[string]bool{}} // spec.md NO existe.
		c := usecase.NewBoxConductor(agent, arts, 3, 40)
		_, err := c.RunWith(context.Background(), boxConInsumos(), insumos, ports.SpawnOpts{Cwd: "/arnes"})
		var pre *usecase.PrecondicionError
		if !errors.As(err, &pre) {
			t.Fatalf("want PrecondicionError, got %v", err)
		}
		if len(pre.Faltantes) != 1 || !strings.Contains(pre.Faltantes[0], "spec.md") {
			t.Errorf("faltantes = %v, want lista con spec.md", pre.Faltantes)
		}
		if agent.spawns != 0 {
			t.Errorf("spawns = %d, want 0 (el run NO arranca — cero tokens)", agent.spawns)
		}
	})

	t.Run("requerido:false jamás bloquea", func(t *testing.T) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, 3)}
		agent := &scriptedAgent{sess: sess}
		arts := &fsArtifacts{existe: map[string]bool{}}
		c := usecase.NewBoxConductor(agent, arts, 1, 40)
		opcional := []domain.Insumo{{Art: "diseño.md", De: "caja:designer", Path: "diseño.md", Requerido: false}}
		if _, err := c.RunWith(context.Background(), boxConInsumos(), opcional, ports.SpawnOpts{Cwd: "/arnes"}); err != nil {
			t.Fatalf("insumo opcional ausente bloqueó el run: %v", err)
		}
		if agent.spawns != 1 {
			t.Errorf("spawns = %d, want 1", agent.spawns)
		}
	})

	t.Run("tarea() cita rutas + digest, jamás el documento (RF-121)", func(t *testing.T) {
		sess := &scriptedSession{events: make(chan ports.AgentEvent, 3)}
		agent := &scriptedAgent{sess: sess}
		arts := &fsArtifacts{
			existe:  map[string]bool{"spec.md": true},
			resumen: map[string]string{"spec.md": "status: done\ncapabilities: CAP-01"},
		}
		c := usecase.NewBoxConductor(agent, arts, 1, 40)
		if _, err := c.RunWith(context.Background(), boxConInsumos(), insumos, ports.SpawnOpts{Cwd: "/arnes"}); err != nil {
			t.Fatalf("run: %v", err)
		}
		if len(sess.prompts) == 0 {
			t.Fatal("sin prompt spawneado")
		}
		p := sess.prompts[0]
		if !strings.Contains(p, "ruta: spec.md") {
			t.Errorf("el prompt no cita la ruta del insumo:\n%s", p)
		}
		if !strings.Contains(p, "CAP-01") {
			t.Errorf("el prompt no incluye el resumen determinista:\n%s", p)
		}
	})
}

// controlRequestSession emite un control_request seguido de un result por cada Send —
// dirige TestConductorAutoResuelveControlRequest sin subproceso real. Guarda cada
// ControlDecision que recibió para asertar cómo el conductor respondió.
type controlRequestSession struct {
	events    chan ports.AgentEvent
	tool      string
	mu        sync.Mutex
	responded []ports.ControlDecision
}

func (s *controlRequestSession) Send(_ context.Context, _ string) error {
	s.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "req-1", Tool: s.tool, ToolUseID: "tu-1"}
	s.events <- ports.AgentEvent{Kind: ports.EventResult, Subtype: "success"}
	return nil
}
func (s *controlRequestSession) Events() <-chan ports.AgentEvent { return s.events }
func (s *controlRequestSession) Interrupt(context.Context) error { return nil }
func (s *controlRequestSession) Close() error                    { return nil }
func (s *controlRequestSession) RespondControl(_ context.Context, _ string, d ports.ControlDecision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responded = append(s.responded, d)
	return nil
}

// singleSessionAgent spawnea siempre la MISMA sesión ya construida — helper genérico para
// probar `RunWith` con una sesión custom (a diferencia de scriptedAgent, atado a *scriptedSession).
type singleSessionAgent struct{ sess ports.AgentSession }

func (a *singleSessionAgent) Spawn(context.Context, ports.SpawnOpts) (ports.AgentSession, error) {
	return a.sess, nil
}

// TestConductorAutoResuelveControlRequest — deuda BACKLOG «run async + gate post-run»
// (2026-07-23): un box run es headless (sin humano en el loop), así que `awaitResult` DEBE
// resolver todo `control_request` contra el PermissionSet del rol en vez de colgarse
// esperando un click que nunca llega (bug real que este ticket cazó y cerró: antes de este
// fix, `awaitResult` descartaba EventControlRequest en silencio — cualquier rol con
// permiso de escritura habría colgado el subproceso para siempre, porque `--permission-
// prompt-tool stdio` rutea TODO write tool por acá sin excepción, `escrituraDirecta` en
// claudecode/conductor.go).
func TestConductorAutoResuelveControlRequest(t *testing.T) {
	t.Run("Allow del rol -> auto-allow, EscribioAlgo=true", func(t *testing.T) {
		sess := &controlRequestSession{events: make(chan ports.AgentEvent, 4), tool: "Write"}
		arts := &scriptedArtifacts{statuses: []string{"done"}}
		c := usecase.NewBoxConductor(&singleSessionAgent{sess: sess}, arts, 1, 40)
		ps := domain.PermissionSet{Rol: "dev", Allow: []string{"Write"}}

		out, err := c.RunWith(context.Background(), boxWithRuta([]domain.Route{{A: "reviewer"}}, nil), nil,
			ports.SpawnOpts{Cwd: "/arnes", Permisos: ps})
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if !out.EscribioAlgo {
			t.Error("EscribioAlgo = false, want true (Write auto-permitido por el rol)")
		}
		if len(sess.responded) != 1 || !sess.responded[0].Allow {
			t.Fatalf("RespondControl no auto-permitió: %+v", sess.responded)
		}
	})

	t.Run("Ask/no-listado -> auto-deny (deny-by-default, sin humano a quien preguntarle), jamás cuelga", func(t *testing.T) {
		sess := &controlRequestSession{events: make(chan ports.AgentEvent, 4), tool: "Bash"}
		arts := &scriptedArtifacts{statuses: []string{"done"}}
		c := usecase.NewBoxConductor(&singleSessionAgent{sess: sess}, arts, 1, 40)
		ps := domain.PermissionSet{Rol: "dev"} // "Bash" no listado ⇒ Decide()==Ask.

		out, err := c.RunWith(context.Background(), boxWithRuta([]domain.Route{{A: "reviewer"}}, nil), nil,
			ports.SpawnOpts{Cwd: "/arnes", Permisos: ps})
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.EscribioAlgo {
			t.Error("EscribioAlgo = true, want false (Bash autodenegado, nada se escribió)")
		}
		if len(sess.responded) != 1 || sess.responded[0].Allow {
			t.Fatalf("RespondControl no denegó: %+v", sess.responded)
		}
		if sess.responded[0].Message == "" {
			t.Error("deny sin motivo visible — el operador no vería por qué el run se frenó ahí")
		}
	})

	t.Run("Deny explícito del rol -> auto-deny", func(t *testing.T) {
		sess := &controlRequestSession{events: make(chan ports.AgentEvent, 4), tool: "Edit"}
		arts := &scriptedArtifacts{statuses: []string{"done"}}
		c := usecase.NewBoxConductor(&singleSessionAgent{sess: sess}, arts, 1, 40)
		ps := domain.PermissionSet{Rol: "reader", Deny: []string{"Edit"}}

		out, err := c.RunWith(context.Background(), boxWithRuta([]domain.Route{{A: "reviewer"}}, nil), nil,
			ports.SpawnOpts{Cwd: "/arnes", Permisos: ps})
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if out.EscribioAlgo {
			t.Error("EscribioAlgo = true, want false (Edit denegado por el rol)")
		}
		if len(sess.responded) != 1 || sess.responded[0].Allow {
			t.Fatalf("RespondControl no denegó: %+v", sess.responded)
		}
	})
}

// --- RunService async (deuda BACKLOG «run async + gate post-run», 2026-07-23) ---

type fakeIndexPort struct{ g domain.Graph }

func (f *fakeIndexPort) Rebuild(context.Context) error                       { return nil }
func (f *fakeIndexPort) Query(context.Context, string) (domain.Graph, error) { return f.g, nil }
func (f *fakeIndexPort) List(context.Context) ([]ports.EntradaIndice, error) { return nil, nil }
func (f *fakeIndexPort) Upsert(context.Context, string, domain.Graph) error  { return nil }

type fakePermissionPort struct{ ps domain.PermissionSet }

func (f *fakePermissionPort) ResolveForRole(context.Context, string) (domain.PermissionSet, error) {
	return f.ps, nil
}

type fakeWorkdirResolver struct{ path string }

func (f *fakeWorkdirResolver) Resolve(string) (string, bool, error) { return f.path, true, nil }

type fakeRunPublisher struct {
	mu     sync.Mutex
	frames [][]byte
}

func (f *fakeRunPublisher) Publish(_ string, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.frames = append(f.frames, data)
}

func (f *fakeRunPublisher) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.frames)
}

// blockingConfPort — el gate post-run: cuenta cuántas veces se llamó RunGraph (asserta que
// SOLO corre cuando el run escribió algo).
type countingConfPort struct {
	mu    sync.Mutex
	calls int
}

func (c *countingConfPort) Run(context.Context, ports.Target) (domain.ConformanceReport, error) {
	return domain.ConformanceReport{}, nil
}

func (c *countingConfPort) RunGraph(context.Context, []byte, string) (domain.ConformanceReport, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return domain.ConformanceReport{Target: "arnes"}, nil
}

func (c *countingConfPort) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func waitForRunTerminal(t *testing.T, runs *usecase.RunService, runID string) usecase.RunStatus {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st, ok := runs.GetRun(runID)
		if !ok {
			t.Fatalf("GetRun(%q): no encontrado", runID)
		}
		if st.Estado != "corriendo" {
			return st
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("run %q no terminó en 2s (¿colgado?)", runID)
	return usecase.RunStatus{}
}

// TestRunAsyncDevuelveRunIDInmediato — StartRun (POST /boxes/{id}/run) devuelve el run_id
// SIN esperar a que el conductor termine (el punto entero de la deuda: antes bloqueaba la
// respuesta HTTP hasta el desenlace). El estado pasa corriendo→terminado vía GetRun.
func TestRunAsyncDevuelveRunIDInmediato(t *testing.T) {
	g := domain.Graph{
		Arnes: &domain.Arnes{Rol: "dev"},
		Nodes: []domain.Box{boxWithRuta([]domain.Route{{A: "reviewer"}}, nil)},
	}
	sess := &scriptedSession{
		events:  make(chan ports.AgentEvent, 3),
		results: []ports.AgentEvent{{Kind: ports.EventResult, Subtype: "success"}},
	}
	conductor := usecase.NewBoxConductor(&scriptedAgent{sess: sess}, &scriptedArtifacts{statuses: []string{"done"}}, 3, 40)
	pub := &fakeRunPublisher{}
	runs := usecase.NewRunService(
		&fakeIndexPort{g: g}, conductor, &fakePermissionPort{ps: domain.PermissionSet{Rol: "dev"}},
		&fakeWorkdirResolver{path: "/arnes"}, nil, pub, nil, nil,
	)

	runID, err := runs.StartRun(context.Background(), "arnes-x", "caja-x")
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	if runID == "" {
		t.Fatal("StartRun devolvió run_id vacío")
	}
	// El registro YA existe apenas vuelve StartRun (antes de que el goroutine termine) —
	// eso es lo que hace posible el 202 inmediato.
	st0, ok := runs.GetRun(runID)
	if !ok {
		t.Fatalf("GetRun(%q) inmediatamente después de StartRun: no encontrado", runID)
	}
	if st0.Estado != "corriendo" && st0.Estado != "terminado" {
		t.Errorf("estado inicial = %q, want corriendo (o terminado si la goroutine ya corrió)", st0.Estado)
	}

	final := waitForRunTerminal(t, runs, runID)
	if final.Estado != "terminado" {
		t.Fatalf("estado final = %q, want terminado (result=%+v err=%q)", final.Estado, final.Result, final.Error)
	}
	if final.Result == nil || final.Result.RunID != runID {
		t.Fatalf("Result = %+v, want RunID=%q", final.Result, runID)
	}
	// event: run sigue viajando por SSE en paralelo (started + finished) — sin cambio de
	// contrato para quien ya escuchaba /events antes de esta deuda.
	if got := pub.count(); got != 2 {
		t.Errorf("frames publicados = %d, want 2 (started + finished)", got)
	}

	if _, ok := runs.GetRun("run-inexistente"); ok {
		t.Error("GetRun de un run_id inexistente devolvió ok=true")
	}
}

// TestRunAsyncValidacionSigueSincrona — 404/422 (arnés/nodo inexistente, nodo sin
// contrato) le llegan al caller EN StartRun, nunca escondidos detrás de un 202 que
// después falla en silencio (la validación es barata — cero tokens quemados).
func TestRunAsyncValidacionSigueSincrona(t *testing.T) {
	runs := usecase.NewRunService(
		&fakeIndexPort{g: domain.Graph{Nodes: []domain.Box{{ID: "no-es-caja"}}}},
		usecase.NewBoxConductor(&scriptedAgent{sess: &scriptedSession{events: make(chan ports.AgentEvent, 1)}}, &scriptedArtifacts{}, 1, 40),
		&fakePermissionPort{}, &fakeWorkdirResolver{path: "/arnes"}, nil, &fakeRunPublisher{}, nil, nil,
	)

	if _, err := runs.StartRun(context.Background(), "arnes-x", "no-existe"); !errors.Is(err, usecase.ErrNodoNoExiste) {
		t.Errorf("nodo inexistente: err = %v, want ErrNodoNoExiste", err)
	}
	if _, err := runs.StartRun(context.Background(), "arnes-x", "no-es-caja"); !errors.Is(err, usecase.ErrNoEsCaja) {
		t.Errorf("nodo sin contrato: err = %v, want ErrNoEsCaja", err)
	}
}

// TestRunAsyncGatePostRunSoloSiEscribio — el gate post-run (CAP-70/71: mismo patrón que
// el chat corre conformance tras un turno con escrituras) SOLO llama RunGraph cuando el
// run auto-permitió ≥1 escritura — un run 100% lectura no paga ese costo.
func TestRunAsyncGatePostRunSoloSiEscribio(t *testing.T) {
	g := domain.Graph{
		Arnes: &domain.Arnes{Rol: "dev"},
		Nodes: []domain.Box{boxWithRuta([]domain.Route{{A: "reviewer"}}, nil)},
	}

	t.Run("con escritura -> gate corre", func(t *testing.T) {
		sess := &controlRequestSession{events: make(chan ports.AgentEvent, 4), tool: "Write"}
		conductor := usecase.NewBoxConductor(&singleSessionAgent{sess: sess}, &scriptedArtifacts{statuses: []string{"done"}}, 1, 40)
		conf := &countingConfPort{}
		runs := usecase.NewRunService(
			&fakeIndexPort{g: g}, conductor, &fakePermissionPort{ps: domain.PermissionSet{Rol: "dev", Allow: []string{"Write"}}},
			&fakeWorkdirResolver{path: "/arnes"}, nil, &fakeRunPublisher{}, conf, func(string) string { return "/arnes" },
		)
		runID, err := runs.StartRun(context.Background(), "arnes-x", "caja-x")
		if err != nil {
			t.Fatalf("StartRun: %v", err)
		}
		final := waitForRunTerminal(t, runs, runID)
		if final.Result == nil || final.Result.Conformance == nil {
			t.Fatalf("Conformance = nil, want el reporte del gate post-run (result=%+v)", final.Result)
		}
		if conf.count() != 1 {
			t.Errorf("RunGraph llamado %d veces, want 1", conf.count())
		}
	})

	t.Run("sin escritura -> gate NO corre", func(t *testing.T) {
		sess := &scriptedSession{
			events:  make(chan ports.AgentEvent, 3),
			results: []ports.AgentEvent{{Kind: ports.EventResult, Subtype: "success"}},
		}
		conductor := usecase.NewBoxConductor(&scriptedAgent{sess: sess}, &scriptedArtifacts{statuses: []string{"done"}}, 3, 40)
		conf := &countingConfPort{}
		runs := usecase.NewRunService(
			&fakeIndexPort{g: g}, conductor, &fakePermissionPort{ps: domain.PermissionSet{Rol: "dev"}},
			&fakeWorkdirResolver{path: "/arnes"}, nil, &fakeRunPublisher{}, conf, func(string) string { return "/arnes" },
		)
		runID, err := runs.StartRun(context.Background(), "arnes-x", "caja-x")
		if err != nil {
			t.Fatalf("StartRun: %v", err)
		}
		final := waitForRunTerminal(t, runs, runID)
		if final.Result == nil {
			t.Fatalf("Result = nil")
		}
		if final.Result.Conformance != nil {
			t.Errorf("Conformance = %+v, want nil (run 100%% lectura, gate no debía correr)", final.Result.Conformance)
		}
		if conf.count() != 0 {
			t.Errorf("RunGraph llamado %d veces, want 0", conf.count())
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

// TestHealthzCORSReflected is HS-14's regression test: /healthz used to bypass CORS entirely
// (early-return before the Origin gate ran), so a WebView fetch() from its own origin
// (tauri://localhost, distinct from 127.0.0.1:4200) saw a healthy 200 as a network error —
// indistinguishable from "the daemon is down". Scoped to the /healthz branch specifically (not
// a whole-file scan) so this can't pass just because /api/version sets the header elsewhere.
func TestHealthzCORSReflected(t *testing.T) {
	auth := readSourceFile(t, "internal/adapters/transport/http/auth.go")
	if auth == "" {
		return
	}
	i := strings.Index(auth, `r.URL.Path == "/healthz"`)
	if i < 0 {
		t.Fatal("healthz branch not found in auth.go")
	}
	j := strings.Index(auth[i:], "next.ServeHTTP(w, r)")
	if j < 0 {
		t.Fatal("healthz branch never falls through to next.ServeHTTP")
	}
	branch := auth[i : i+j]
	if !strings.Contains(branch, "Access-Control-Allow-Origin") {
		t.Errorf("the /healthz branch skips CORS reflection — a cross-origin WebView fetch() " +
			"sees a healthy daemon as a network error (HS-14 regression)")
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

// TestMaquinariaNoContaminaArnes — boundary maquinaria-no-contamina-arnes (research
// 2026-07-05-arquitectura-inyeccion-knowhow.md §9, materializado 2026-07-23): la doctrina ①②
// entra SOLO por flags de sesión — nunca `--bare` (rompe el auth de suscripción) y nunca un
// archivo escrito dentro del árbol del arnés. `SpawnArgs` es la superficie de enforcement real
// (se exporta para esto, ver su doc comment).
func TestMaquinariaNoContaminaArnes(t *testing.T) {
	args := claudecode.SpawnArgs(ports.SpawnOpts{
		Injection: ports.Injection{
			PluginDirs:       []string{"/app-owned/kit-a", "/app-owned/kit-b"},
			SystemPromptFile: "/app-owned/doctrine.md",
			AddDirs:          []string{"/app-owned/knowhow"},
			MCPConfigFile:    "/app-owned/mcp.json",
		},
	})
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--bare") {
		t.Fatalf("SpawnArgs incluye --bare: rompe el auth de suscripción del operador (headless-sdk)")
	}
	for _, want := range []string{"--plugin-dir", "--append-system-prompt-file", "--add-dir", "--mcp-config"} {
		if !strings.Contains(joined, want) {
			t.Errorf("SpawnArgs con Injection poblada no emite %s — la doctrina no entraría por flag", want)
		}
	}
}

// --- sesion-viva-consistente.md + permisos-gui (behavioral, with fakes) ---

type fakeSession struct {
	events chan ports.AgentEvent
	mu     sync.Mutex
	sent   []string
	// responded records every RespondControl (the control_response wire the daemon sent).
	responded []respondedControl
	// closed cuenta los Close() — la transición de conversación tiene que cerrar el
	// conductor del hilo que deja de estar activo.
	closed int
}

type respondedControl struct {
	requestID string
	decision  ports.ControlDecision
}

func (f *fakeSession) Send(_ context.Context, turn string) error {
	f.mu.Lock()
	f.sent = append(f.sent, turn)
	f.mu.Unlock()
	return nil
}
func (f *fakeSession) Events() <-chan ports.AgentEvent { return f.events }
func (f *fakeSession) Interrupt(context.Context) error { return nil }

func (f *fakeSession) Close() error {
	f.mu.Lock()
	f.closed++
	f.mu.Unlock()
	return nil
}

// closedSnapshot es la lectura lock-safe de cuántas veces se cerró este conductor. La
// necesita `transicion-de-conversacion-atomica`: «cierra el conductor vivo» es una de las
// cinco cosas que el check afirma, y sin contarlo sólo se podría inferir del respawn.
func (f *fakeSession) closedSnapshot() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func (f *fakeSession) RespondControl(_ context.Context, requestID string, d ports.ControlDecision) error {
	f.mu.Lock()
	f.responded = append(f.responded, respondedControl{requestID: requestID, decision: d})
	f.mu.Unlock()
	return nil
}

func (f *fakeSession) respondedSnapshot() []respondedControl {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]respondedControl, len(f.responded))
	copy(out, f.responded)
	return out
}

func (f *fakeSession) sentSnapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.sent))
	copy(out, f.sent)
	return out
}

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

// spawnsSnapshot is the lock-safe read of a.spawns — needed whenever a heal (an async
// respawn from the consume goroutine) can race a polling read from the test goroutine.
func (a *fakeAgent) spawnsSnapshot() []ports.SpawnOpts {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]ports.SpawnOpts, len(a.spawns))
	copy(out, a.spawns)
	return out
}

func (a *fakeAgent) sessionsSnapshot() []*fakeSession {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]*fakeSession, len(a.sessions))
	copy(out, a.sessions)
	return out
}

type fakeResolver struct{ path string }

func (r fakeResolver) Resolve(string) (string, bool, error) { return r.path, true, nil }

type fakeStore struct{}

func (fakeStore) Load(context.Context) ([]domain.Session, error) { return nil, nil }
func (fakeStore) Save(context.Context, []domain.Session) error   { return nil }

// storeFalible es fakeStore con un interruptor: modela el filesystem que deja de aceptar
// escrituras a mitad de una sesión de trabajo (disco lleno, montaje read-only). Lo necesita
// la mitad del check que dice «un fallo deja el estado anterior intacto»: sin un disco que
// falle no hay forma de afirmarlo, y afirmarlo sin probarlo sería el pass fabricado.
type storeFalible struct {
	mu   sync.Mutex
	roto bool
}

func (*storeFalible) Load(context.Context) ([]domain.Session, error) { return nil, nil }

func (s *storeFalible) Save(context.Context, []domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.roto {
		return errors.New("no space left on device")
	}
	return nil
}

func (s *storeFalible) romper() { s.setRoto(true) }
func (s *storeFalible) sanar()  { s.setRoto(false) }

func (s *storeFalible) setRoto(v bool) {
	s.mu.Lock()
	s.roto = v
	s.mu.Unlock()
}

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
	return newTestServiceConStore(t, agent, pub, cwd, fakeStore{})
}

// newTestServiceConStore es el mismo helper con el store inyectable. Se factorizó al escribir
// el enforcer de la transición atómica, que necesita un disco que falle; los llamadores de
// `newTestService` no cambian ni una línea — siguen sembrando exactamente una sesión en
// `List()[0]`, que es de lo que dependen los cuatro checks originales de este boundary.
func newTestServiceConStore(t *testing.T, agent ports.AgentPort, pub usecase.EventPublisher, cwd string, store ports.SessionStore) *usecase.SessionService {
	t.Helper()
	// The REAL role provisioner backs the permission seam: the fitness runs against the
	// same authority the daemon wires (permisos-derivan-del-rol).
	svc, err := usecase.NewSessionService(context.Background(), agent, store, pub, fakeResolver{path: cwd}, 40, nil, permission.NewKitProvisioner(), nil)
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

// TestResumeAutoSana enforces resume-auto-sana: a `--resume` whose process dies before ever
// emitting `init` (stale ClaudeSessionID — e.g. CC pruned it) must self-heal ONCE — restart
// fresh (dropping the stale id) and resend the in-flight turn — never wedge the session or
// surface a raw error for a condition the daemon can recover from by itself.
func TestResumeAutoSana(t *testing.T) {
	agent := &fakeAgent{}
	pub := &fakePub{}
	svc := newTestService(t, agent, pub, t.TempDir())
	id := svc.List()[0].ID

	// Turn 1: fresh spawn (no resume yet), inits with a session id, then its process dies —
	// that id is what the NEXT turn will try (and fail) to resume.
	//
	// El turno 1 NO manda EventResult a propósito, y esto es sincronización, no estilo: el
	// handle vivo lo suelta la goroutine de consume DESPUÉS de que el canal cierra
	// (`session_service.go`, «Channel closed»), y ese paso no publica nada. Con un
	// EventResult, `Status` ya quedaba en Idle antes del cierre, así que esperar Idle no
	// esperaba a nadie: el Turn 2 corría contra un `r.live` todavía no soltado, no
	// spawneaba, y el test fallaba ~1 de cada 100 corridas bajo carga (flake preexistente,
	// reproducido en la base 306b80c). Sin EventResult, `Status` sigue en Streaming hasta
	// que la limpieza del cierre lo baja a Idle — o sea que esperar Idle espera EXACTAMENTE
	// al paso que suelta el handle. Ninguna aserción cambia; sólo deja de haber carrera.
	if err := svc.Turn(id, "primero"); err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	sess1 := agent.sessionsSnapshot()[0]
	sess1.events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-stale"}
	close(sess1.events) // process life ends (Turn 2 must see r.live==nil to spawn again).
	waitFor(t, 2*time.Second, func() bool { return svc.List()[0].Status == domain.StatusIdle })

	// Turn 2: spawnLocked sees the stale id and asks --resume. The process dies before ever
	// emitting init (channel closes with nothing sent) — a failed resume, not a real error.
	if err := svc.Turn(id, "segundo — el resume se pierde"); err != nil {
		t.Fatalf("turn 2: %v", err)
	}
	if spawns := agent.spawnsSnapshot(); len(spawns) != 2 || spawns[1].Resume != "cc-stale" {
		t.Fatalf("el 2do spawn debía pedir --resume cc-stale, got %+v", spawns)
	}
	sess2 := agent.sessionsSnapshot()[1]
	close(sess2.events)

	// The heal must respawn FRESH (stale id dropped) and resend the pending turn — no manual
	// retry, no wedged session, no raw error frame for a condition the daemon self-heals.
	// The heal itself runs asynchronously (from the consume goroutine, after the close above),
	// so every read past this point must go through the lock-safe snapshots — a plain field
	// read here would race the heal's write.
	waitFor(t, 2*time.Second, func() bool { return len(agent.spawnsSnapshot()) == 3 })
	spawns := agent.spawnsSnapshot()
	if spawns[2].Resume != "" {
		t.Errorf("el heal debía respawnear SIN --resume (id stale descartado), got %q", spawns[2].Resume)
	}
	sess3 := agent.sessionsSnapshot()[2]
	waitFor(t, 2*time.Second, func() bool { return len(sess3.sentSnapshot()) > 0 })
	if got := sess3.sentSnapshot(); got[0] != "segundo — el resume se pierde" {
		t.Errorf("el heal no reenvió el turno pendiente, got %v", got)
	}
}

// TestTransicionDeConversacionEsAtomica es el enforcer del quinto check de
// `sesion-viva-consistente` v1.2 (`transicion-de-conversacion-atomica`). Nació declarado
// pendiente en el nodo, sin cuerpo, para no fabricarle un pass; este es el cuerpo.
//
// Afirma las cinco cosas que el check dice, en una sola corrida contra el servicio real:
//
//  1. exige turno quieto — con un turno en vuelo la transición devuelve ErrBusy (→ 409);
//  2. cierra el conductor vivo — el proceso del hilo que se desactiva recibe su Close();
//  3. deniega los permisos pendientes CON MOTIVO, y el motivo nombra la desactivación;
//  4. descarta los grants efímeros — el mismo tool vuelve a preguntar en el hilo nuevo;
//  5. deja exactamente una activa, y un fallo de persistencia deja el estado anterior
//     intacto (ni media transición, ni una sesión sin conductor).
//
// Lo que NO afirma, y por eso no se enuncia: nada sobre el `run_id` del frame nuevo. El frame
// `conversacion` no lleva run_id a propósito (no pertenece a un turno) y su idempotencia es
// declarativa — trae el estado final. Eso lo cubren los tests del usecase, no este check.
func TestTransicionDeConversacionEsAtomica(t *testing.T) {
	agent := &fakeAgent{}
	pub := &fakePub{}
	disco := &storeFalible{}
	svc := newTestServiceConStore(t, agent, pub, t.TempDir(), disco)
	id := svc.List()[0].ID

	// (1) Con el turno en vuelo, las dos transiciones se rechazan con el MISMO error que
	// `Turn` usa para el turno concurrente: dos conductores sobre un stdin es exactamente
	// lo que la propiedad «un productor por recurso serializado» prohíbe.
	if err := svc.Turn(id, "un pedido largo"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	if _, _, err := svc.CrearConversacion(id); !errors.Is(err, usecase.ErrBusy) {
		t.Errorf("crear con el turno en vuelo = %v, want ErrBusy (→ 409)", err)
	}
	if _, _, err := svc.ActivarConversacion(id, "cv-cualquiera"); !errors.Is(err, usecase.ErrBusy) {
		t.Errorf("retomar con el turno en vuelo = %v, want ErrBusy (→ 409)", err)
	}

	// El conductor pide permiso y el operador lo aprueba: eso mintea un grant efímero. La
	// segunda pregunta por el mismo tool ya no llega al humano — se auto-aprueba.
	viejo := agent.sessionsSnapshot()[0]
	viejo.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-1", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`)}
	waitFor(t, 2*time.Second, func() bool { return svc.List()[0].Status == domain.StatusAwait })
	if _, err := svc.ResolvePermission(id, "cr-1", "allow", "backend-dev", time.Minute, nil); err != nil {
		t.Fatalf("resolve cr-1: %v", err)
	}
	// Una tarjeta que queda ABIERTA: el conductor pregunta y nadie contesta. El `result`
	// del turno devuelve la sesión a idle sin limpiarla — ese es el estado real en el que
	// la transición se encuentra un permiso pendiente.
	viejo.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-2", Tool: "Bash", Input: []byte(`{"command":"ls"}`)}
	waitFor(t, 2*time.Second, func() bool { return svc.List()[0].Status == domain.StatusAwait })
	viejo.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo"}
	waitFor(t, 2*time.Second, func() bool { return svc.List()[0].Status == domain.StatusIdle })

	// (5b) Con el disco roto, la transición NO ocurre y el estado anterior queda intacto.
	antes := svc.List()[0]
	activaAntes, _ := antes.Activa()
	disco.romper()
	if _, _, err := svc.CrearConversacion(id); err == nil {
		t.Fatal("crear con el disco roto tiene que fallar: en memoria no puede quedar lo que no se guardó")
	}
	ahora := svc.List()[0]
	if len(ahora.Conversaciones) != len(antes.Conversaciones) {
		t.Errorf("el rollback dejó %d conversaciones, había %d", len(ahora.Conversaciones), len(antes.Conversaciones))
	}
	if c, ok := ahora.Activa(); !ok || c.ID != activaAntes.ID {
		t.Errorf("el rollback cambió la activa (%v), tenía que dejar %q", c, activaAntes.ID)
	}
	if n := viejo.closedSnapshot(); n != 0 {
		t.Errorf("el conductor se cerró %d vez/veces en una transición que falló: la sesión quedaría sin proceso", n)
	}

	// La transición de verdad, con el disco sano.
	disco.sanar()
	nueva, desactivada, err := svc.CrearConversacion(id)
	if err != nil {
		t.Fatalf("crear: %v", err)
	}
	if desactivada != activaAntes.ID {
		t.Errorf("desactivada = %q, want %q — la operación tiene que nombrar cuál desactivó", desactivada, activaAntes.ID)
	}

	// (5a) Exactamente una activa, y es la nueva.
	final := svc.List()[0]
	if err := domain.VerificarUnaActiva(final); err != nil {
		t.Fatalf("la invariante se rompió tras la transición: %v", err)
	}
	if c, _ := final.Activa(); c.ID != nueva.ID {
		t.Errorf("la activa es %q, want la recién creada %q", c.ID, nueva.ID)
	}

	// (2) El conductor del hilo que se desactivó está cerrado.
	waitFor(t, 2*time.Second, func() bool { return viejo.closedSnapshot() == 1 })

	// (3) La tarjeta que quedó pendiente se resolvió como deny, con un motivo que nombra la
	// desactivación y NO se confunde con el de Interrupt.
	var cerrada map[string]any
	for _, f := range pub.snapshot() {
		if f["kind"] == "permission_result" && f["request_id"] == "cr-2" {
			cerrada = f
		}
	}
	if cerrada == nil {
		t.Fatal("la tarjeta pendiente quedó abierta: el conductor esperaría una respuesta que nadie va a dar")
	}
	if cerrada["decision"] != string(domain.DecisionDeny) {
		t.Errorf("decision = %v, want deny (deny-by-default)", cerrada["decision"])
	}
	if motivo, _ := cerrada["text"].(string); !strings.Contains(motivo, "desactivada") || strings.Contains(motivo, "interrumpido") {
		t.Errorf("motivo = %q: tiene que nombrar la desactivación y ser distinguible del de Interrupt", motivo)
	}

	// (4) Los grants no se heredan: el mismo tool que ya se había aprobado vuelve a
	// preguntar en el hilo nuevo. Heredarlo sería aprobar algo que el operador nunca vio acá.
	if err := svc.Turn(id, "seguimos en el hilo nuevo"); err != nil {
		t.Fatalf("turn en el hilo nuevo: %v", err)
	}
	fresco := agent.sessionsSnapshot()[1]
	fresco.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-3", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`)}
	waitFor(t, 2*time.Second, func() bool {
		for _, f := range pub.snapshot() {
			if f["request_id"] == "cr-3" {
				return true
			}
		}
		return false
	})
	for _, f := range pub.snapshot() {
		if f["request_id"] != "cr-3" {
			continue
		}
		if f["kind"] != "permission" {
			t.Errorf("cr-3 salió como %v: el grant del hilo anterior se heredó (deny-by-default violado)", f["kind"])
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

// --- versionado.md (source scans) ---

// TestVersionManifestsInSync guards docs/architecture/conventions/versionado.md: the desktop
// bundle version lives in three manifests (Cargo.toml is the source of truth — Tauri falls
// back to it — but tauri.conf.json/package.json stay explicit per Tauri's own recommendation)
// that `make bump-patch` keeps in lockstep. A hand-edit to just one of them, or a stray "v"
// prefix (which Keygen's Release.version rejects, see the licensing story), drifts silently
// until someone diffs a shipped installer against its own about box.
func TestVersionManifestsInSync(t *testing.T) {
	cargo := readSourceFile(t, "web/src-tauri/Cargo.toml")
	tauriConf := readSourceFile(t, "web/src-tauri/tauri.conf.json")
	pkg := readSourceFile(t, "web/package.json")
	if cargo == "" || tauriConf == "" || pkg == "" {
		return
	}

	cargoVer := regexp.MustCompile(`(?m)^version = "([^"]+)"`).FindStringSubmatch(cargo)
	tauriVer := regexp.MustCompile(`"version":\s*"([^"]+)"`).FindStringSubmatch(tauriConf)
	pkgVer := regexp.MustCompile(`"version":\s*"([^"]+)"`).FindStringSubmatch(pkg)
	if cargoVer == nil {
		t.Fatal("no [package].version en web/src-tauri/Cargo.toml")
	}
	if tauriVer == nil {
		t.Fatal("no \"version\" en web/src-tauri/tauri.conf.json")
	}
	if pkgVer == nil {
		t.Fatal("no \"version\" en web/package.json")
	}

	v := cargoVer[1]
	if strings.HasPrefix(v, "v") {
		t.Errorf("Cargo.toml version=%q lleva prefijo v — Keygen.Release.version lo rechaza", v)
	}
	if tauriVer[1] != v {
		t.Errorf("tauri.conf.json version=%q != Cargo.toml version=%q (drift — correr `make bump-patch`)", tauriVer[1], v)
	}
	if pkgVer[1] != v {
		t.Errorf("package.json version=%q != Cargo.toml version=%q (drift — correr `make bump-patch`)", pkgVer[1], v)
	}
}

// TestElPunteroDeConversacionSigueALaTransicion — A-7 (auditoría 2026-07-26).
//
// `sesion-viva-consistente` se extendió a v1.2 con el argumento de que «el runtime sigue
// siendo uno por sesión y **gana un puntero que dice a quién le pertenece**». El quinto
// check nuevo se declaraba enforzado por `TestTransicionDeConversacionEsAtomica` — pero
// borrar la asignación del puntero (`r.convActiva = nueva.ID`) dejaba ese test VERDE. El
// mecanismo que el nodo v1.2 agregó no lo probaba nadie: es N-12 otra vez, en forma más
// sutil (allá el enforcer no podía correr; acá corría y pasaba sin tocar lo suyo).
//
// Cómo se observa un campo privado sin exportarlo para el test: por su CONSECUENCIA. El
// puntero existe para que `chequearDueño` pueda avisar cuando el runtime y el agregado se
// separan. Si la transición no lo mueve, el runtime se queda creyendo que le pertenece la
// conversación VIEJA, y el primer turno sobre la nueva emite el warn de divergencia. Con la
// asignación puesta, ese warn no aparece nunca. Se afirma sobre el log, que es la superficie
// que el propio nodo eligió para este detector.
func TestElPunteroDeConversacionSigueALaTransicion(t *testing.T) {
	agent := &fakeAgent{}
	svc := newTestService(t, agent, &fakePub{}, t.TempDir())
	id := svc.List()[0].ID

	// Un turno en la conversación original: deja el puntero del runtime apuntando a ella.
	if err := svc.Turn(id, "el primer tema"); err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	sess1 := agent.sessionsSnapshot()[0]
	sess1.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo"}
	waitFor(t, 2*time.Second, func() bool { return svc.List()[0].Status == domain.StatusIdle })

	vieja, ok := svc.List()[0].Activa()
	if !ok {
		t.Fatal("precondición: la sesión tiene que tener una activa")
	}

	// La transición. A partir de acá el runtime tiene que pertenecer a la conversación nueva.
	nueva, desactivada, err := svc.CrearConversacion(id)
	if err != nil {
		t.Fatalf("crear: %v", err)
	}
	if desactivada != vieja.ID {
		t.Fatalf("la desactivada fue %q, esperaba %q", desactivada, vieja.ID)
	}

	// Desde acá se escucha el log: el turno siguiente pasa por `activa()` ⇒ `chequearDueño`.
	previo := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previo) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))

	if err := svc.Turn(id, "el tema nuevo"); err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	if s := buf.String(); strings.Contains(s, "no coinciden en cuál es la conversación activa") {
		t.Errorf("la transición NO movió el puntero del runtime: el detector de divergencia se "+
			"disparó en el primer turno de la conversación nueva %q.\nlog:\n%s", nueva.ID, s)
	}

	// Y el turno cayó donde tenía que caer: en la conversación NUEVA, no en la vieja. Sin
	// esto, un puntero correcto con el turno en el hilo equivocado pasaría igual.
	final := svc.List()[0]
	for _, c := range final.Conversaciones {
		switch c.ID {
		case nueva.ID:
			if c.NumTurnos() == 0 {
				t.Error("el turno no cayó en la conversación nueva")
			}
		case vieja.ID:
			if c.NumTurnos() != vieja.NumTurnos() {
				t.Errorf("la conversación vieja creció de %d a %d turnos después de desactivarse",
					vieja.NumTurnos(), c.NumTurnos())
			}
		}
	}
}
