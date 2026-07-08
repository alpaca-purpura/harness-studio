package loader_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// repoRoot sube desde el cwd del test hasta encontrar go.mod (misma convención que los
// tests de usecase).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod no encontrado")
		}
		dir = parent
	}
}

// chdirRepoRoot mueve el cwd a la raíz del repo (restaurado en cleanup) para poder cargar
// el arnés por su dir repo-relativo: así los fuente_path estampados salen repo-relativos,
// EXACTAMENTE como los declara el fixture (ver godoc de LoadArnes sobre la base heredada).
func chdirRepoRoot(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return root
}

// jsonDe serializa v a JSON o revienta el test — para comparar por igualdad de marshal.
func jsonDe(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestLoaderDevFullCycle es el contrato central del Puente 1: cargar el arnés dogfood real
// en disco (forma plugin) reproduce el grafo del fixture escrito a mano — mismo manifiesto,
// mismos nodos (clase/banda/fase/estado/fuente_path/contract) y mismo SET de edges — y el
// grafo derivado valida contra graph.l0.schema.json.
func TestLoaderDevFullCycle(t *testing.T) {
	root := chdirRepoRoot(t)

	raw, err := os.ReadFile(filepath.Join("dogfood", "dev-full-cycle.graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	var quiero domain.Graph
	if uerr := json.Unmarshal(raw, &quiero); uerr != nil {
		t.Fatal(uerr)
	}

	got, err := loader.LoadArnes(filepath.Join("dogfood", "dev-full-cycle"))
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}

	// El manifiesto sale de arnes.l0.json VERBATIM → igualdad total por marshal.
	if a, b := jsonDe(t, got.Arnes), jsonDe(t, quiero.Arnes); !bytes.Equal(a, b) {
		t.Errorf("arnes derivado ≠ fixture:\n  got:    %s\n  quiero: %s", a, b)
	}

	// Nodos: mismo conteo y, por id, mismos ejes que el fixture declara.
	if len(got.Nodes) != len(quiero.Nodes) {
		t.Errorf("nodos: got %d, quiero %d (got=%v)", len(got.Nodes), len(quiero.Nodes), ids(got.Nodes))
	}
	for _, w := range quiero.Nodes {
		g, ok := got.NodeByID(w.ID)
		if !ok {
			t.Errorf("nodo %q ausente en el grafo derivado", w.ID)
			continue
		}
		if g.Clase != w.Clase {
			t.Errorf("%s: clase got %q, quiero %q", w.ID, g.Clase, w.Clase)
		}
		if g.Banda != w.Banda {
			t.Errorf("%s: banda got %q, quiero %q", w.ID, g.Banda, w.Banda)
		}
		if g.Fase != w.Fase {
			t.Errorf("%s: fase got %q, quiero %q", w.ID, g.Fase, w.Fase)
		}
		if g.Estado != w.Estado {
			t.Errorf("%s: estado got %q, quiero %q", w.ID, g.Estado, w.Estado)
		}
		if g.FuentePath != w.FuentePath {
			t.Errorf("%s: fuente_path got %q, quiero %q", w.ID, g.FuentePath, w.FuentePath)
		}
		if a, b := jsonDe(t, g.Contract), jsonDe(t, w.Contract); !bytes.Equal(a, b) {
			t.Errorf("%s: contract derivado ≠ fixture:\n  got:    %s\n  quiero: %s", w.ID, a, b)
		}
	}

	// Edges como SET (domain.Edge es comparable): derivados ≡ declarados en el fixture.
	gotSet := edgeSet(got.Edges)
	quieroSet := edgeSet(quiero.Edges)
	for e := range quieroSet {
		if !gotSet[e] {
			t.Errorf("edge del fixture no derivado: %+v", e)
		}
	}
	for e := range gotSet {
		if !quieroSet[e] {
			t.Errorf("edge derivado que el fixture no declara: %+v", e)
		}
	}

	validarContraSchema(t, root, got)
}

// TestLoaderNoReconocido cubre la reconciliación §4.5 (D-c firmada): un dir de skill sin
// SKILL.md se emite como nodo no-reconocido VISIBLE — jamás descarte silencioso ni error
// fatal.
func TestLoaderNoReconocido(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"rogue-arnes"}`)
	escribir(t, filepath.Join(dir, "skills", "rogue", "notas.txt"), "esto no es una skill")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	n, ok := g.NodeByID("rogue")
	if !ok {
		t.Fatalf("el nodo no-reconocido debe ser VISIBLE; nodos: %v", ids(g.Nodes))
	}
	if n.Clase != domain.ClaseNoReconocido {
		t.Errorf("clase got %q, quiero %q", n.Clase, domain.ClaseNoReconocido)
	}
	if n.Nombre != "rogue (no reconocido)" {
		t.Errorf("nombre got %q", n.Nombre)
	}
	if n.FuentePath == "" {
		t.Error("fuente_path debe quedar estampado también en lo no reconocido")
	}
}

// TestLoaderSinManifiesto cubre el modo degradado honesto (§2): sin arnes.l0.json el grafo
// sale con Arnes nil y SIN error — los nodos igual se reconocen; el check rojo lo decide el
// caller.
func TestLoaderSinManifiesto(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"sin-manifiesto"}`)
	escribir(t, filepath.Join(dir, "skills", "hola", "SKILL.md"),
		"---\nname: hola\nnombre: saludar\n---\n\n# hola\n")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes != nil {
		t.Errorf("sin manifiesto el Arnes debe ser nil, got %+v", g.Arnes)
	}
	if n, ok := g.NodeByID("hola"); !ok || n.Clase != domain.ClaseSkill || n.Nombre != "saludar" {
		t.Errorf("la skill debe reconocerse igual en modo degradado, got %+v (ok=%v)", n, ok)
	}
}

// TestLoaderNoEsArnes cubre el detector §1: un dir sin ninguna de las dos formas físicas es
// un error explícito — jamás un grafo vacío.
func TestLoaderNoEsArnes(t *testing.T) {
	if _, err := loader.LoadArnes(t.TempDir()); !errors.Is(err, loader.ErrNoEsArnes) {
		t.Fatalf("quiero ErrNoEsArnes, got %v", err)
	}
}

// TestLoaderInstalado cubre la segunda forma física (§1, D-b firmada): con `.claude/` y sin
// plugin.json los reconocedores escanean bajo `.claude/`, y el fuente_path lo refleja.
func TestLoaderInstalado(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude", "skills", "hola", "SKILL.md"),
		"---\nname: hola\nnombre: saludar\n---\n\n# hola\n")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	n, ok := g.NodeByID("hola")
	if !ok || n.Clase != domain.ClaseSkill {
		t.Fatalf("skill instalada no reconocida, got %+v (ok=%v)", n, ok)
	}
	quiero := filepath.ToSlash(filepath.Join(dir, ".claude", "skills", "hola", "SKILL.md"))
	if n.FuentePath != quiero {
		t.Errorf("fuente_path got %q, quiero %q", n.FuentePath, quiero)
	}
}

// ── helpers ──

// escribir crea un archivo (y sus dirs) dentro del fixture temporal del test.
func escribir(t *testing.T, ruta, contenido string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
}

func ids(nodos []domain.Box) []string {
	out := make([]string, 0, len(nodos))
	for _, n := range nodos {
		out = append(out, n.ID)
	}
	return out
}

func edgeSet(edges []domain.Edge) map[domain.Edge]bool {
	s := make(map[domain.Edge]bool, len(edges))
	for _, e := range edges {
		s[e] = true
	}
	return s
}

// validarContraSchema valida el grafo derivado contra graph.l0.schema.json resolviendo el
// $ref hermano (box.contract.schema.json) desde el mismo dir — mismo patrón siblingLoader
// que internal/adapters/conformance/mechanism/schema.go, replicado aquí a propósito para no
// importar ese package (territorio de otro hilo).
func validarContraSchema(t *testing.T, root string, g domain.Graph) {
	t.Helper()
	dirSchemas := filepath.Join(root, "arch", "contracts", "schema")
	carga := func(nombre string) (*jsonschema.Schema, error) {
		raw, err := os.ReadFile(filepath.Join(dirSchemas, filepath.Base(nombre))) //nolint:gosec // G304: dir de schemas del repo, fijo en el test.
		if err != nil {
			return nil, err
		}
		var s jsonschema.Schema
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	}
	raiz, err := carga("graph.l0.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	res, err := raiz.Resolve(&jsonschema.ResolveOptions{
		Loader: func(u *url.URL) (*jsonschema.Schema, error) { return carga(filepath.Base(u.Path)) },
	})
	if err != nil {
		t.Fatal(err)
	}
	var inst any
	if err := json.Unmarshal(jsonDe(t, g), &inst); err != nil {
		t.Fatal(err)
	}
	if err := res.Validate(inst); err != nil {
		t.Errorf("el grafo derivado no valida contra graph.l0.schema.json: %v", err)
	}
}

// TestReconocerHooks cubre la fila `hook` de nomenclatura §3 (franja-artefactos F6):
// una entrada = un nodo de banda Guardia; hooks.json roto = no-reconocido VISIBLE (§4.5).
func TestReconocerHooks(t *testing.T) {
	arma := func(t *testing.T, hooksJSON string) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".claude-plugin"), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude-plugin", "plugin.json"), []byte(`{"name":"x"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if hooksJSON != "" {
			if err := os.MkdirAll(filepath.Join(dir, "hooks"), 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "hooks", "hooks.json"), []byte(hooksJSON), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	t.Run("una entrada = un nodo de Guardia", func(t *testing.T) {
		g, err := loader.LoadArnes(arma(t,
			`{"hooks":{"PostToolUse":[{"matcher":"Write|Edit","hooks":[]}],"Stop":[{"hooks":[]}]}}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 2 {
			t.Fatalf("nodos = %d, want 2 (uno por entrada)", len(g.Nodes))
		}
		post, ok := g.NodeByID("hook-posttooluse")
		if !ok || post.Banda != domain.BandaGuardia || post.Clase != domain.ClaseHook {
			t.Errorf("hook-posttooluse mal emitido: %+v", post)
		}
		if post.Nombre != "PostToolUse · Write|Edit" {
			t.Errorf("nombre = %q, want evento · matcher", post.Nombre)
		}
		if _, ok := g.NodeByID("hook-stop"); !ok {
			t.Error("hook-stop ausente")
		}
	})

	t.Run("hooks.json roto → no-reconocido visible, jamás descarte", func(t *testing.T) {
		g, err := loader.LoadArnes(arma(t, `{esto no es json`))
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 1 || g.Nodes[0].Clase != domain.ClaseNoReconocido {
			t.Errorf("want 1 nodo no-reconocido, got %+v", g.Nodes)
		}
	})

	t.Run("sin hooks/ → cero nodos, cero drama", func(t *testing.T) {
		g, err := loader.LoadArnes(arma(t, ""))
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 0 {
			t.Errorf("nodos = %d, want 0", len(g.Nodes))
		}
	})
}
