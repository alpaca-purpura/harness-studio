package loader_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		if g.Canal != w.Canal {
			t.Errorf("%s: canal got %q, quiero %q", w.ID, g.Canal, w.Canal)
		}
		if g.Procedencia != w.Procedencia {
			t.Errorf("%s: procedencia got %q, quiero %q", w.ID, g.Procedencia, w.Procedencia)
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

// TestLoaderSinManifiesto cubre el modo degradado honesto (§2): SIN NINGÚN manifiesto
// (ni arnes.l0.json ni plugin.json — forma instalada pelada) el grafo sale con Arnes nil
// y SIN error — los nodos igual se reconocen; el check rojo lo decide el caller. (Un
// plugin.json presente SIN arnes.l0.json ya no cae acá desde C-N-14/T3: ese caso puebla
// Arnes vía fallback — ver TestLoaderFallbackPluginJSON.)
func TestLoaderSinManifiesto(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude", "skills", "hola", "SKILL.md"),
		"---\nname: hola\nnombre: saludar\n---\n\n# hola\n")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes != nil {
		t.Errorf("sin ningún manifiesto el Arnes debe ser nil, got %+v", g.Arnes)
	}
	if !g.Degradado {
		t.Error("sin manifiesto el grafo debe marcarse Degradado (S1-D27, contrato §2)")
	}
	if n, ok := g.NodeByID("hola"); !ok || n.Clase != domain.ClaseSkill || n.Nombre != "saludar" {
		t.Errorf("la skill debe reconocerse igual en modo degradado, got %+v (ok=%v)", n, ok)
	}

	// El aviso `manifiesto-ausente` es la señal única del degradado (LoadArnesInfo).
	_, info, ierr := loader.LoadArnesInfo(dir)
	if ierr != nil {
		t.Fatalf("LoadArnesInfo: %v", ierr)
	}
	if !strings.Contains(info.Aviso, "manifiesto-ausente") {
		t.Errorf("aviso = %q, quiero que nombre `manifiesto-ausente`", info.Aviso)
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

// TestLoaderFallbackPluginJSON cubre C-N-14 (D-DOM-4, el caso más común: un plugin de
// marketplace normal, CON plugin.json y SIN arnes.l0.json): el fallback puebla
// ID/Nombre/Descripcion/Version desde plugin.json y estampa FuenteManifiesto —
// "bloqueante de cimientos" resuelto, ya no `Arnes==nil` ni el id ausente.
func TestLoaderFallbackPluginJSON(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"),
		`{"name":"harness-x","displayName":"Harness X","description":"un plugin normal","version":"2.1.0"}`)

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes == nil {
		t.Fatal("con plugin.json el fallback debe poblar Arnes, no dejarlo nil (C-N-14)")
	}
	want := domain.Arnes{
		ID:               "harness-x",
		Nombre:           "Harness X",
		Descripcion:      "un plugin normal",
		Version:          "2.1.0",
		FuenteManifiesto: "plugin.json",
	}
	if !reflect.DeepEqual(*g.Arnes, want) {
		t.Errorf("Arnes = %+v, quiero %+v", *g.Arnes, want)
	}
}

// TestLoaderVersionDesdePluginJSON cubre §2.4 punto 1: con arnes.l0.json presente Y
// plugin.json, la version SIGUE saliendo del plugin.json (arnes.l0.json no la trae) —
// el manifiesto real manda en ID/Nombre/Descripcion, plugin.json solo completa lo que
// falta.
func TestLoaderVersionDesdePluginJSON(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"),
		`{"name":"harness-x","version":"3.4.5"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"),
		`{"id":"harness-x","nombre":"Harness X real","rol":"r","proceso":"p","empresas":["a"],"reporta_a":null}`)

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes == nil {
		t.Fatal("Arnes no debe ser nil")
	}
	if g.Arnes.Version != "3.4.5" {
		t.Errorf("Version = %q, quiero 3.4.5 (desde plugin.json)", g.Arnes.Version)
	}
	if g.Arnes.Nombre != "Harness X real" {
		t.Errorf("Nombre = %q, arnes.l0.json manda sobre plugin.json", g.Arnes.Nombre)
	}
	if g.Arnes.FuenteManifiesto != "arnes.l0.json" {
		t.Errorf("FuenteManifiesto = %q, quiero arnes.l0.json (es la fuente que manda)", g.Arnes.FuenteManifiesto)
	}
}

// TestLoaderPluginJSONCorrupto cubre el degradado honesto: un plugin.json que no parsea,
// SIN arnes.l0.json, no debe ser un error fatal — Arnes sale nil con un aviso visible
// (LoadArnesInfo lo expone; LoadArnes lo descarta por compatibilidad, pero el load no
// truena).
func TestLoaderPluginJSONCorrupto(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{esto no es json`)

	g, info, err := loader.LoadArnesInfo(dir)
	if err != nil {
		t.Fatalf("un plugin.json corrupto no debe ser error fatal: %v", err)
	}
	if g.Arnes != nil {
		t.Errorf("Arnes debe quedar nil ante un plugin.json ilegible, got %+v", g.Arnes)
	}
	if info.Aviso == "" {
		t.Error("quiero un aviso visible cuando plugin.json no parsea")
	}
}

// TestLoaderIDsDiscrepantes cubre RN-IDENT-3: arnes.l0.id ≠ plugin.json.name → arnes.l0
// gana el ID (es la fuente que el autor escribió a mano) pero la discrepancia queda
// anotada como aviso visible, nunca elegida en silencio.
func TestLoaderIDsDiscrepantes(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"nombre-del-plugin"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"),
		`{"id":"nombre-del-manifiesto","rol":"r","proceso":"p","empresas":["a"],"reporta_a":null}`)

	g, info, err := loader.LoadArnesInfo(dir)
	if err != nil {
		t.Fatalf("LoadArnesInfo: %v", err)
	}
	if g.Arnes == nil || g.Arnes.ID != "nombre-del-manifiesto" {
		t.Fatalf("arnes.l0.id debe ganar, got %+v", g.Arnes)
	}
	if info.Aviso == "" {
		t.Error("quiero un aviso visible ante la discrepancia de ids (RN-IDENT-3)")
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
	dirSchemas := filepath.Join(root, "docs", "architecture", "contracts", "schema")
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

// armaPlugin crea el marcador mínimo de forma-plugin (§1) en un dir temporal nuevo.
func armaPlugin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"x"}`)
	return dir
}

// TestReconocerCommandYOutputStyle cubre las filas `command`/`output-style` de nomenclatura
// §3 (auditoría colateral HS-16): archivos sueltos <id>.md, frontmatter OPCIONAL (a
// diferencia de skill, su ausencia NO vuelve el nodo no-reconocido) y un archivo/dir suelto
// que no matchea sí lo vuelve no-reconocido visible (§4.5).
func TestReconocerCommandYOutputStyle(t *testing.T) {
	dir := armaPlugin(t)
	escribir(t, filepath.Join(dir, "commands", "spec.md"), "# /spec\n\nescribe el spec.\n")
	escribir(t, filepath.Join(dir, "commands", "review.md"), "---\nnombre: revisar código\n---\n\n# /review\n")
	escribir(t, filepath.Join(dir, "commands", "rogue.txt"), "no es un comando")
	escribir(t, filepath.Join(dir, "output-styles", "conciso.md"), "# conciso\n\nrespuestas cortas.\n")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatal(err)
	}

	spec, ok := g.NodeByID("spec")
	if !ok || spec.Clase != domain.ClaseCommand || spec.Nombre != "spec" || spec.Banda != "" {
		t.Errorf("spec (sin frontmatter) mal emitido: %+v (ok=%v)", spec, ok)
	}
	review, ok := g.NodeByID("review")
	if !ok || review.Clase != domain.ClaseCommand || review.Nombre != "revisar código" {
		t.Errorf("review (con frontmatter) mal emitido: %+v (ok=%v)", review, ok)
	}
	if _, rogueOk := g.NodeByID("rogue.txt"); !rogueOk {
		t.Error("rogue.txt (no-.md bajo commands/) ausente — debía ser no-reconocido visible")
	}
	for _, n := range g.Nodes {
		if n.ID == "rogue.txt" && n.Clase != domain.ClaseNoReconocido {
			t.Errorf("rogue.txt clase = %q, want no-reconocido", n.Clase)
		}
	}
	conciso, ok := g.NodeByID("conciso")
	if !ok || conciso.Clase != domain.ClaseOutputStyle {
		t.Errorf("conciso (output-style) mal emitido: %+v (ok=%v)", conciso, ok)
	}
}

// TestReconocerMCP cubre la fila `mcp` de nomenclatura §3: un server declarado = un nodo,
// vive en la RAÍZ del arnés (nunca bajo `.claude/`, ni en forma instalada); no parseable/vacío
// → no-reconocido visible (§4.5).
func TestReconocerMCP(t *testing.T) {
	t.Run("un server = un nodo", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, ".mcp.json"), `{"mcpServers":{"linear":{"command":"npx","args":["linear-mcp"]}}}`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		n, ok := g.NodeByID("mcp-linear")
		if !ok || n.Clase != domain.ClaseMCP || n.Nombre != "linear" {
			t.Errorf("mcp-linear mal emitido: %+v (ok=%v)", n, ok)
		}
	})

	t.Run(".mcp.json roto → no-reconocido visible", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, ".mcp.json"), `{esto no es json`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 1 || g.Nodes[0].Clase != domain.ClaseNoReconocido {
			t.Errorf("want 1 nodo no-reconocido, got %+v", g.Nodes)
		}
	})

	t.Run("sin .mcp.json → cero nodos", func(t *testing.T) {
		g, err := loader.LoadArnes(armaPlugin(t))
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 0 {
			t.Errorf("nodos = %d, want 0", len(g.Nodes))
		}
	})
}

// TestReconocerSettings cubre las filas `settings`/`statusline`/`hook` forma-instalada de
// nomenclatura §3: las tres celdas viven en el MISMO settings.json.
func TestReconocerSettings(t *testing.T) {
	t.Run("solo settings.json → un nodo settings", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, "settings.json"), `{"env":{"FOO":"bar"}}`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 1 {
			t.Fatalf("nodos = %d, want 1 (solo settings)", len(g.Nodes))
		}
		n, ok := g.NodeByID("settings")
		if !ok || n.Clase != domain.ClaseSettings {
			t.Errorf("settings mal emitido: %+v (ok=%v)", n, ok)
		}
	})

	t.Run("statusLine presente suma el nodo statusline", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, "settings.json"), `{"statusLine":{"type":"command","command":"echo hola"}}`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := g.NodeByID("settings"); !ok {
			t.Error("settings ausente")
		}
		n, ok := g.NodeByID("statusline")
		if !ok || n.Clase != domain.ClaseStatusline {
			t.Errorf("statusline mal emitido: %+v (ok=%v)", n, ok)
		}
	})

	t.Run("hooks dentro de settings.json = forma instalada de la Guardia", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, "settings.json"),
			`{"hooks":{"Stop":[{"hooks":[]}]}}`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		n, ok := g.NodeByID("hook-stop")
		if !ok || n.Banda != domain.BandaGuardia || n.Clase != domain.ClaseHook {
			t.Errorf("hook-stop (forma instalada) mal emitido: %+v (ok=%v)", n, ok)
		}
	})

	t.Run("settings.json roto → no-reconocido visible", func(t *testing.T) {
		dir := armaPlugin(t)
		escribir(t, filepath.Join(dir, "settings.json"), `{esto no es json`)
		g, err := loader.LoadArnes(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 1 || g.Nodes[0].Clase != domain.ClaseNoReconocido {
			t.Errorf("want 1 nodo no-reconocido, got %+v", g.Nodes)
		}
	})

	t.Run("sin settings.json → cero nodos", func(t *testing.T) {
		g, err := loader.LoadArnes(armaPlugin(t))
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Nodes) != 0 {
			t.Errorf("nodos = %d, want 0", len(g.Nodes))
		}
	})
}

// TestReconocerReglasDir cubre RF-183 (mejorar-arnes-conversando T-L): el directorio de
// rules (`.claude/rules/` instalado · `rules/` plugin) es first-class en el runtime
// (knowledge/elements/rules.md L1.4 — un tema por archivo, recursivo) y el loader lo emite
// como nodos `rule` de banda Base, README incluido (Claude Code carga TODO .md del dir).
func TestReconocerReglasDir(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, "arnes.l0.json"), `{"id":"overlay","nombre":"Overlay","version":"0.1.0"}`)
	escribir(t, filepath.Join(dir, ".claude", "rules", "hipaa-lite.md"),
		"---\nnombre: HIPAA lite\n---\n\n# salvaguardas\n")
	escribir(t, filepath.Join(dir, ".claude", "rules", "README.md"), "# índice de rules\n")
	escribir(t, filepath.Join(dir, ".claude", "rules", "sub", "tema.md"), "# tema anidado\n")
	escribir(t, filepath.Join(dir, ".claude", "rules", "suelto.txt"), "no es una rule")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	casos := []struct{ id, nombre string }{
		{"hipaa-lite", "HIPAA lite"},
		{"README", "README"},
		{"sub/tema", "sub/tema"},
	}
	for _, c := range casos {
		n, ok := g.NodeByID(c.id)
		if !ok {
			t.Fatalf("falta la rule %q; nodos: %v", c.id, ids(g.Nodes))
		}
		if n.Clase != domain.ClaseRule || n.Banda != domain.BandaBase {
			t.Errorf("%s: clase/banda got %q/%q, quiero rule/base", c.id, n.Clase, n.Banda)
		}
		if n.Nombre != c.nombre {
			t.Errorf("%s: nombre got %q, quiero %q", c.id, n.Nombre, c.nombre)
		}
		if n.FuentePath == "" {
			t.Errorf("%s: fuente_path vacío", c.id)
		}
	}
	// Lo no-.md bajo rules/ es visible como no-reconocido (§4.5), jamás descarte silencioso.
	if n, ok := g.NodeByID("suelto.txt"); !ok || n.Clase != domain.ClaseNoReconocido {
		t.Errorf("suelto.txt debe ser no-reconocido visible, got %+v (ok=%v)", n, ok)
	}
}
