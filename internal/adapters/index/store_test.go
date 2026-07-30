package index

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// newTestStore opens a Store at a fresh temp path with no reg/load — the shape unit
// tests below need (Query/List/Upsert directly), matching every caller elsewhere in
// the repo that constructs a Store without ever calling Rebuild.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "index.db"), nil, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestSeedServesDogfood asserts the index seeds the real dogfood arnés (dev-full-cycle)
// so the Map endpoint serves it (HS-09 Hito 1, architecture §4.1): 5 nodes, 4 edges,
// the spec-writer box carries its contract (caja + estado).
func TestSeedServesDogfood(t *testing.T) {
	s := newTestStore(t)

	g, err := s.Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query(dev-full-cycle) = %v, want nil", err)
	}
	if g.Arnes == nil || g.Arnes.ID != "dev-full-cycle" {
		t.Fatalf("Arnes.ID = %v, want dev-full-cycle", g.Arnes)
	}
	// 7 = 4 cajas + 1 regla + 2 hooks de la Guardia (franja-artefactos F4/F6: el dogfood
	// ganó hooks reales y el reconocedor de nomenclatura §3 los emite).
	if got := len(g.Nodes); got != 7 {
		t.Errorf("len(Nodes) = %d, want 7", got)
	}
	if got := len(g.Edges); got != 4 {
		t.Errorf("len(Edges) = %d, want 4", got)
	}

	spec, ok := g.NodeByID("spec-writer")
	if !ok {
		t.Fatal("NodeByID(spec-writer) not found")
	}
	if !spec.IsCaja() {
		t.Error("spec-writer.IsCaja() = false, want true (contract.caja)")
	}
	if spec.Estado != "idea -> spec" {
		t.Errorf("spec-writer.Estado = %q, want %q", spec.Estado, "idea -> spec")
	}
	if spec.Clase != domain.ClaseSkill {
		t.Errorf("spec-writer.Clase = %q, want skill", spec.Clase)
	}
}

// TestSeedFases asserts the dogfood declares its four phases in order (the lane order the
// Map renders, RF-12) and that the base-band rule is present.
func TestSeedFases(t *testing.T) {
	g, err := newTestStore(t).Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query(dev-full-cycle) = %v", err)
	}
	want := []string{"spec", "build", "review", "release"}
	if len(g.Arnes.Fases) != len(want) {
		t.Fatalf("Fases = %v, want %v", g.Arnes.Fases, want)
	}
	for i, f := range want {
		if string(g.Arnes.Fases[i]) != f {
			t.Errorf("Fases[%d] = %q, want %q", i, g.Arnes.Fases[i], f)
		}
	}
	if std, ok := g.NodeByID("std-spec"); !ok || std.Banda != domain.BandaBase {
		t.Errorf("std-spec banda = %v, want base", std.Banda)
	}
}

// TestSeedServesShowcase asserts the index seeds the showcase arnés (content-studio-full,
// HS-09 Hito 2): a maximal graph whose rol is Editorial (agnostic-to-rubro proof) and that
// exercises all 10 clases so the Map can render every casuistic. Conformance-valid.
func TestSeedServesShowcase(t *testing.T) {
	g, err := newTestStore(t).Query(context.Background(), "content-studio-full")
	if err != nil {
		t.Fatalf("Query(content-studio-full) = %v, want nil", err)
	}
	if g.Arnes == nil || g.Arnes.Rol != "Editorial · Content Lead" {
		t.Fatalf("showcase rol = %v, want Editorial · Content Lead", g.Arnes)
	}
	seen := map[domain.Clase]bool{}
	for _, n := range g.Nodes {
		seen[n.Clase] = true
	}
	for _, c := range []domain.Clase{
		domain.ClaseSkill, domain.ClaseSubagent, domain.ClaseHook, domain.ClaseRule,
		domain.ClaseCommand, domain.ClaseMCP, domain.ClasePlugin, domain.ClaseSettings,
		domain.ClaseOutputStyle, domain.ClaseStatusline,
	} {
		if !seen[c] {
			t.Errorf("showcase missing clase %q — the kitchen-sink must exercise all 10", c)
		}
	}
}

// TestQueryUnknown asserts an unknown harness id is a clean ErrNotFound (the endpoint
// maps it to 404), not a panic or empty graph.
func TestQueryUnknown(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Query(context.Background(), "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Query(nope) err = %v, want ErrNotFound", err)
	}
}

// TestUpsertIndexaBajoLaClaveNoElArnesID — deuda BACKLOG «re-key (home,id,scope)», cerrada
// 2026-07-23: la llave del índice es la que el caller pasa explícito, NUNCA re-derivada del
// propio g.Arnes.ID del grafo. Un grafo cuyo Arnes.ID difiere de la clave sigue siendo
// consultable SOLO por la clave — así se cierra el hueco que hacía colisionar dos arneses
// distintos con el mismo id pelado.
func TestUpsertIndexaBajoLaClaveNoElArnesID(t *testing.T) {
	s := newTestStore(t)
	g := domain.Graph{Arnes: &domain.Arnes{ID: "harness", Marketplace: "acme/repo"}}
	if err := s.Upsert(context.Background(), "acme-repo~harness~", g); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if _, err := s.Query(context.Background(), "harness"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Query(harness) [el bare ID, NO la clave] = %v, want ErrNotFound", err)
	}
	got, err := s.Query(context.Background(), "acme-repo~harness~")
	if err != nil {
		t.Fatalf("Query(clave) = %v, want nil", err)
	}
	if got.Arnes.ID != "harness" {
		t.Errorf("Arnes.ID = %q, want %q (el grafo viaja intacto, solo cambia la llave)", got.Arnes.ID, "harness")
	}
}

// TestUpsertDosArnesesMismoIDNoColisionan — el caso REAL que motivó la deuda: dos arneses de
// homes distintos comparten el mismo `id` pelado ("harness" de dos marketplaces). Antes del
// re-key, el segundo Upsert pisaba en silencio al primero (mismo key = g.Arnes.ID); ahora
// cada uno vive bajo su propia clave calificada.
func TestUpsertDosArnesesMismoIDNoColisionan(t *testing.T) {
	s := newTestStore(t)
	a := domain.Graph{Arnes: &domain.Arnes{ID: "harness", Marketplace: "acme/repo", Nombre: "Acme"}}
	b := domain.Graph{Arnes: &domain.Arnes{ID: "harness", Marketplace: "otro/repo", Nombre: "Otro"}}
	if err := s.Upsert(context.Background(), "acme-repo~harness~", a); err != nil {
		t.Fatalf("Upsert a: %v", err)
	}
	if err := s.Upsert(context.Background(), "otro-repo~harness~", b); err != nil {
		t.Fatalf("Upsert b: %v", err)
	}
	gotA, err := s.Query(context.Background(), "acme-repo~harness~")
	if err != nil || gotA.Arnes.Nombre != "Acme" {
		t.Errorf("Query(clave-a) = %+v, %v — quiero Acme sobreviviente", gotA.Arnes, err)
	}
	gotB, err := s.Query(context.Background(), "otro-repo~harness~")
	if err != nil || gotB.Arnes.Nombre != "Otro" {
		t.Errorf("Query(clave-b) = %+v, %v — quiero Otro sobreviviente, NO pisado por a", gotB.Arnes, err)
	}
}

// TestUpsertClaveVaciaError / TestUpsertSinManifiestoError — ambos guardas honestos: ni una
// llave vacía ni un grafo sin manifiesto son indexables (nada inventado).
func TestUpsertClaveVaciaError(t *testing.T) {
	s := newTestStore(t)
	err := s.Upsert(context.Background(), "", domain.Graph{Arnes: &domain.Arnes{ID: "x"}})
	if err == nil {
		t.Fatal("Upsert con clave vacía = nil, want error")
	}
}

func TestUpsertSinManifiestoError(t *testing.T) {
	s := newTestStore(t)
	err := s.Upsert(context.Background(), "alguna-clave", domain.Graph{})
	if err == nil {
		t.Fatal("Upsert sin Arnes = nil, want error")
	}
}

// TestListPortfolio asserts List returns every seeded harness ordered by clave (the
// portfolio's stable order, RF-72) and includes the real dogfood arnés.
func TestListPortfolio(t *testing.T) {
	entradas, err := newTestStore(t).List(context.Background())
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	claves := make([]string, 0, len(entradas))
	for _, e := range entradas {
		if e.Grafo.Arnes != nil {
			claves = append(claves, e.Clave)
		}
	}
	// New() seeds demo + dev-full-cycle + content-studio-full; sorted by clave.
	want := []string{"content-studio-full", "demo", "dev-full-cycle"}
	if len(claves) != len(want) {
		t.Fatalf("List claves = %v, want %v", claves, want)
	}
	for i, w := range want {
		if claves[i] != w {
			t.Errorf("List claves = %v, want %v", claves, want)
			break
		}
	}
}

// TestListDevuelveLaClaveNoElArnesID es la regresión del cartel «no está en el índice del
// daemon»: una entrada del Portafolio se indexa bajo su clave calificada
// («sin-home~vitalia~vitalia») mientras el manifiesto sigue diciendo «vitalia». List tiene
// que devolver la clave — es la única cuerda que después sirve para Query. Cuando devolvía
// el id interno, un arnés perfectamente cargable se veía como ausente.
func TestListDevuelveLaClaveNoElArnesID(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	const clave = "sin-home~vitalia~vitalia"
	g := domain.Graph{Arnes: &domain.Arnes{ID: "vitalia"}}
	if err := st.Upsert(ctx, clave, g); err != nil {
		t.Fatalf("Upsert(%s) = %v", clave, err)
	}

	entradas, err := st.List(ctx)
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	var hallada *ports.EntradaIndice
	for i := range entradas {
		if entradas[i].Clave == clave {
			hallada = &entradas[i]
			break
		}
	}
	if hallada == nil {
		claves := make([]string, 0, len(entradas))
		for _, e := range entradas {
			claves = append(claves, e.Clave)
		}
		t.Fatalf("List no trajo la clave %q; trajo %v", clave, claves)
	}
	if hallada.Grafo.Arnes == nil || hallada.Grafo.Arnes.ID != "vitalia" {
		t.Errorf("el grafo de %q perdió su manifiesto (arnes.id interno)", clave)
	}
	// La clave listada tiene que servir tal cual para pedir el grafo.
	if _, qerr := st.Query(ctx, hallada.Clave); qerr != nil {
		t.Errorf("Query(clave listada %q) = %v, want nil", hallada.Clave, qerr)
	}
}

// regFija is a minimal ports.ArnesRegistry fake for Rebuild tests: a fixed set of
// arnés→path entries, Register/Resolve unused (Rebuild only calls List).
type regFija []ports.ArnesPath

func (r regFija) Resolve(string) (string, bool, error) {
	return "", false, errors.New("regFija: Resolve no implementado")
}

func (r regFija) Register(string, string) error {
	return errors.New("regFija: Register no implementado")
}

func (r regFija) List() []ports.ArnesPath { return r }

// TestRebuildDesdeArnesRegistry cubre RF-207 escenario 1: N arneses registrados y sanos —
// Rebuild reconstruye el índice exactamente desde ArnesRegistry+loader, y ningún dato
// demo/seed hardcodeado sobrevive.
func TestRebuildDesdeArnesRegistry(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "index.db"), regFija{
		{Arnes: "uno", Path: "/arneses/uno"},
		{Arnes: "dos", Path: "/arneses/dos"},
	}, func(dir string) (domain.Graph, error) {
		id := filepath.Base(dir)
		return domain.Graph{Arnes: &domain.Arnes{ID: id, Nombre: "cargado: " + dir}}, nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if err = s.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	gs, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(gs) != 2 {
		t.Fatalf("List() tras Rebuild = %d grafos, want 2 (ningún seed/demo debe sobrevivir): %+v", len(gs), gs)
	}
	g, err := s.Query(context.Background(), "uno")
	if err != nil {
		t.Fatalf("Query(uno): %v", err)
	}
	if g.Arnes.Nombre != "cargado: /arneses/uno" {
		t.Errorf("Query(uno).Arnes.Nombre = %q, want el grafo real de loader.LoadArnes", g.Arnes.Nombre)
	}
	if _, err := s.Query(context.Background(), "demo"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Query(demo) tras Rebuild = %v, want ErrNotFound (seed data no debe sobrevivir)", err)
	}
}

// TestRebuildEntradaDegradada cubre RF-207 escenario 2: un árbol roto (loader falla) queda
// indexado Degradado:true — nunca ausente, nunca con pass fabricado.
func TestRebuildEntradaDegradada(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "index.db"), regFija{
		{Arnes: "roto", Path: "/arneses/roto"},
	}, func(dir string) (domain.Graph, error) {
		return domain.Graph{}, fmt.Errorf("manifiesto ausente en %s", dir)
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if err = s.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	g, err := s.Query(context.Background(), "roto")
	if err != nil {
		t.Fatalf("Query(roto) = %v, want nil (degradado pero presente)", err)
	}
	if !g.Degradado {
		t.Error("Query(roto).Degradado = false, want true")
	}
	if g.Arnes == nil || g.Arnes.ID != "roto" {
		t.Errorf("Query(roto).Arnes = %+v, want ID=roto (sintetizado desde el registro)", g.Arnes)
	}
}

// TestRebuildRegistroVacio cubre RF-207 escenario 3: ArnesRegistry vacío → List() vacío, sin
// error — y ningún dato de seed queda atrás.
func TestRebuildRegistroVacio(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "index.db"), regFija{}, func(string) (domain.Graph, error) {
		t.Fatal("load no debería llamarse con ArnesRegistry vacío")
		return domain.Graph{}, nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if err = s.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	gs, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(gs) != 0 {
		t.Errorf("List() tras Rebuild con registro vacío = %d grafos, want 0", len(gs))
	}
}

// TestSchemaVersionMismatchWipesFile cubre RF-208 a nivel unitario: un .db existente con una
// schema_version distinta se borra entero (no ALTER TABLE) y New() arranca de un estado
// idéntico al de un path que nunca existió.
func TestSchemaVersionMismatchWipesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.db")

	// Simula un .db viejo: schema_meta con una versión que ya no es la actual, más una fila
	// "old-junk" que NO debe sobrevivir al wipe.
	raw, err := sql.Open(driverName, "file:"+path)
	if err != nil {
		t.Fatalf("abrir .db crudo: %v", err)
	}
	for _, stmt := range []string{
		createMetaTable,
		createGraphsTable,
		`INSERT INTO schema_meta (version) VALUES (-1)`,
		`INSERT INTO graphs (clave, graph_json, updated_at) VALUES ('old-junk', '{"arnes":{"id":"old"}}', 'x')`,
	} {
		if _, eerr := raw.ExecContext(context.Background(), stmt); eerr != nil {
			t.Fatalf("seed .db viejo (%s): %v", stmt, eerr)
		}
	}
	if cerr := raw.Close(); cerr != nil {
		t.Fatalf("close raw: %v", cerr)
	}

	s, err := New(path, regFija{{Arnes: "nuevo", Path: "/arneses/nuevo"}}, func(string) (domain.Graph, error) {
		return domain.Graph{Arnes: &domain.Arnes{ID: "nuevo"}}, nil
	})
	if err != nil {
		t.Fatalf("New tras schema_version vieja: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if _, qerr := s.Query(context.Background(), "old-junk"); !errors.Is(qerr, ErrNotFound) {
		t.Errorf("Query(old-junk) tras el wipe = %v, want ErrNotFound (la fila vieja no debe sobrevivir)", qerr)
	}
	if rerr := s.Rebuild(context.Background()); rerr != nil {
		t.Fatalf("Rebuild tras wipe: %v", rerr)
	}
	if _, qerr := s.Query(context.Background(), "nuevo"); qerr != nil {
		t.Errorf("Query(nuevo) tras Rebuild post-wipe = %v, want nil", qerr)
	}
}

// TestWriterSerializedConcurrentUpsertsSucceed cubre RF-209 escenario 2: Upserts
// concurrentes contra el mismo Store nunca devuelven SQLITE_BUSY/error de lock, y el
// estado final refleja todos (el writer de una sola conexión los serializa).
func TestWriterSerializedConcurrentUpsertsSucceed(t *testing.T) {
	s := newTestStore(t)
	const n = 16
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			g := domain.Graph{Arnes: &domain.Arnes{ID: fmt.Sprintf("concurrente-%d", i)}}
			errs[i] = s.Upsert(context.Background(), fmt.Sprintf("clave-%d", i), g)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("Upsert(clave-%d) concurrente = %v, want nil (writer serializado, sin busy)", i, err)
		}
	}
	for i := range n {
		if _, err := s.Query(context.Background(), fmt.Sprintf("clave-%d", i)); err != nil {
			t.Errorf("Query(clave-%d) tras concurrencia = %v, want nil", i, err)
		}
	}
}
