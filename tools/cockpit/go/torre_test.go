package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitT corre git en dir con identidad fija (los runners no tienen user.name global).
func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=torre-test", "-c", "user.email=t@t"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func repoDemo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitT(t, dir, "init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("uno\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitT(t, dir, "add", ".")
	gitT(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func cat(t *testing.T, v map[string]any) string {
	t.Helper()
	s, _ := v["categoria"].(string)
	return s
}

// ── eje REPO: categorías EXACTAS de la SPEC §4 ──────────────────────────────

func TestTorreRepoCategorias(t *testing.T) {
	dir := repoDemo(t)

	t.Run("repo limpio sin upstream ni tags", func(t *testing.T) {
		repo := torreRepo(dir)
		if repo["repo_root"] == "" {
			t.Error("repo_root debe reportarse (RN-01)")
		}
		if c := cat(t, repo["rama"].(map[string]any)); c != "sin-upstream" {
			t.Errorf("rama: esperaba sin-upstream, obtuve %s", c)
		}
		if c := cat(t, repo["en_vuelo"].(map[string]any)); c != "limpio" {
			t.Errorf("en_vuelo: esperaba limpio, obtuve %s", c)
		}
		if c := cat(t, repo["ultimo_tag"].(map[string]any)); c != "sin-tag" {
			t.Errorf("ultimo_tag: esperaba sin-tag, obtuve %s", c)
		}
		for _, col := range []string{"rama", "en_vuelo", "ultimo_tag"} {
			if repo[col].(map[string]any)["medido_en"] == nil {
				t.Errorf("%s sin medido_en (RN-17)", col)
			}
		}
	})

	t.Run("archivo sucio → en-vuelo con XY porcelain", func(t *testing.T) {
		torreFactsCache = map[string]torreFactsEntry{} // sin cache entre subtests
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("dos\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ev := torreEnVuelo(dir)
		if cat(t, ev) != "en-vuelo" || ev["total"] != 1 {
			t.Fatalf("esperaba en-vuelo con 1 archivo: %+v", ev)
		}
		archivos := ev["archivos"].([]map[string]any)
		if archivos[0]["path"] != "a.txt" || archivos[0]["estado"] != " M" {
			t.Errorf("porcelain XY mal parseado (el espacio de X importa): %+v", archivos[0])
		}
		gitT(t, dir, "checkout", "--", ".")
	})

	t.Run("tag + commits_desde", func(t *testing.T) {
		gitT(t, dir, "tag", "v0.1.0")
		if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitT(t, dir, "add", ".")
		gitT(t, dir, "commit", "-q", "-m", "post-tag")
		ut := torreUltimoTag(dir)
		if cat(t, ut) != "taggeado" || ut["tag"] != "v0.1.0" || ut["commits_desde"] != 1 {
			t.Errorf("esperaba taggeado v0.1.0 +1: %+v", ut)
		}
	})

	t.Run("upstream sincronizada / adelante / divergida", func(t *testing.T) {
		clon := t.TempDir()
		gitT(t, clon, "clone", "-q", dir, ".")
		if c := cat(t, torreRama(clon)); c != "sincronizada" {
			t.Errorf("esperaba sincronizada, obtuve %s", c)
		}
		if err := os.WriteFile(filepath.Join(clon, "c.txt"), []byte("c\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitT(t, clon, "add", ".")
		gitT(t, clon, "commit", "-q", "-m", "local")
		v := torreRama(clon)
		if cat(t, v) != "adelante" || v["ahead"] != 1 || v["behind"] != 0 {
			t.Errorf("esperaba adelante 1/0: %+v", v)
		}
	})

	t.Run("no-repo → eje entero no-medido", func(t *testing.T) {
		repo := torreRepo(t.TempDir())
		for _, col := range []string{"rama", "en_vuelo", "ultimo_tag"} {
			if c := cat(t, repo[col].(map[string]any)); c != "no-medido" {
				t.Errorf("%s: esperaba no-medido, obtuve %s (no-medido ≠ rojo, RN-13)", col, c)
			}
		}
	})
}

// ── gate_fabrica (RN-05/RN-13/RN-16) ────────────────────────────────────────

func TestTorreGateFabrica(t *testing.T) {
	dir := t.TempDir()

	t.Run("verde y rojo con nativo al lado", func(t *testing.T) {
		v := torreRunGate(dir, "echo ok")
		if cat(t, v) != "verde" || v["exit_code"] != 0 || !strings.Contains(v["resumen"].(string), "ok") {
			t.Errorf("esperaba verde exit 0: %+v", v)
		}
		v = torreRunGate(dir, "echo drift detectado; exit 3")
		if cat(t, v) != "rojo" || v["exit_code"] != 3 {
			t.Errorf("rojo = el check corrió y FALLÓ (RN-13): %+v", v)
		}
	})

	t.Run("sin declaración → sin-gate, no rojo", func(t *testing.T) {
		v := torreGateFabrica(registryProject{Name: "x", Path: dir}, false)
		if cat(t, v) != "sin-gate" {
			t.Errorf("esperaba sin-gate (RN-05): %+v", v)
		}
	})

	t.Run("declarado pero aún no medido → no-medido; ?medir lo mide", func(t *testing.T) {
		torreGateCache = map[string]map[string]any{}
		p := registryProject{Name: "y", Path: dir, GateCheck: "echo listo"}
		if c := cat(t, torreGateFabrica(p, false)); c != "no-medido" {
			t.Errorf("sin boot ni ?medir debe ser no-medido: %s", c)
		}
		if c := cat(t, torreGateFabrica(p, true)); c != "verde" {
			t.Errorf("?medir=gate_fabrica debe medir: %s", c)
		}
		// segunda llamada sin medir: sirve el cache (jamás re-corre en loop — RN-16)
		if c := cat(t, torreGateFabrica(p, false)); c != "verde" {
			t.Errorf("debe servir el cache: %s", c)
		}
	})
}

// ── eje PROYECTO (RN-03 convención · RN-11 contrato ledger · RN-02 nullable) ──

func TestTorreProyectos(t *testing.T) {
	ws := t.TempDir()

	// célula CON espejo máquina (shape RN-10)
	celula := filepath.Join(ws, "products", "demo")
	if err := os.MkdirAll(celula, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(celula, "LEDGER.md"), []byte("# Ledger demo\n"), 0o644)
	os.WriteFile(filepath.Join(celula, "ledger.yaml"), []byte(
		"sistema: demo\nfuente: products/demo/LEDGER.md\n"+
			"fichas:\n- {id: DM-01, titulo: fundación, estado: decidida, vigencia: vigente, nota: ''}\n"+
			"- {id: DM-02, titulo: segunda, estado: capturada, vigencia: vigente, nota: ''}\n"+
			"log:\n- {fecha: '2026-07-01', decision: una, fichas: [DM-01]}\n"+
			"- {fecha: '2026-07-03', decision: otra, fichas: [DM-02]}\n"), 0o644)

	// célula SIN espejo → sin-ledger honesto (RN-02)
	pelada := filepath.Join(ws, "products", "pelada")
	os.MkdirAll(pelada, 0o755)
	os.WriteFile(filepath.Join(pelada, "LEDGER.md"), []byte("# Ledger\n"), 0o644)

	// sistema sdd por convención con board
	stories := filepath.Join(ws, "app", "docs", "product", "stories", "S-001")
	os.MkdirAll(stories, 0o755)
	os.WriteFile(filepath.Join(stories, "story.md"), []byte("---\nstory_id: S-001\nstate: developing\n---\ncuerpo\n"), 0o644)

	proyectos := torreProyectos(ws)
	porSlug := map[string]map[string]any{}
	for _, p := range proyectos {
		porSlug[p["slug"].(string)] = p
	}

	demo := porSlug["demo"]
	if demo == nil || demo["tipo"] != "celula" {
		t.Fatalf("products/demo/LEDGER.md debe descubrirse como célula (RN-03): %+v", proyectos)
	}
	ledger := demo["ledger"].(map[string]any)
	if cat(t, ledger) != "con-ledger" || ledger["total_fichas"] != 2 || ledger["ultima_fecha"] != "2026-07-03" {
		t.Errorf("contrato RN-11 roto: %+v", ledger)
	}
	if uf := ledger["ultima_ficha"].(map[string]any); uf["id"] != "DM-02" {
		t.Errorf("ultima_ficha debe ser la última del espejo: %+v", uf)
	}

	if c := cat(t, porSlug["pelada"]["ledger"].(map[string]any)); c != "sin-ledger" {
		t.Errorf("célula sin ledger.yaml → sin-ledger (RN-11): %s", c)
	}

	app := porSlug["app"]
	if app == nil || app["tipo"] != "sdd" {
		t.Fatalf("app/docs/product debe descubrirse como sdd (RN-03): %+v", proyectos)
	}
	board := app["board"].(map[string]any)
	if cat(t, board) != "con-board" || board["total_stories"] != 1 {
		t.Errorf("board con 1 story: %+v", board)
	}
	if board["stories"].(map[string]int)["developing"] != 1 {
		t.Errorf("conteo por estado roto: %+v", board["stories"])
	}

	// F3: el campo arquitectura es ADITIVO en TODO proyecto (RN-25); sdd → sin-arquitectura v1 (RN-20)
	for _, slug := range []string{"demo", "pelada", "app"} {
		arq, ok := porSlug[slug]["arquitectura"].(map[string]any)
		if !ok {
			t.Fatalf("%s: proyectos[] debe traer arquitectura (RN-25)", slug)
		}
		if c := cat(t, arq); c != "sin-arquitectura" {
			t.Errorf("%s sin arquitectura.yaml → sin-arquitectura (RN-02), obtuve %s", slug, c)
		}
	}
}

// ── lente arquitectura (RN-20/RN-22/RN-25) ──────────────────────────────────

func TestTorreArquitectura(t *testing.T) {
	dir := t.TempDir()

	t.Run("archivo ausente → sin-arquitectura, jamás inventa", func(t *testing.T) {
		v := torreArquitectura(filepath.Join(dir, "no-existe.yaml"))
		if cat(t, v) != "sin-arquitectura" || v["medido_en"] == nil {
			t.Errorf("RN-02/RN-17 rotos: %+v", v)
		}
	})

	t.Run("YAML ilegible → no-medido con motivo, no otra cosa", func(t *testing.T) {
		path := filepath.Join(dir, "rota.yaml")
		os.WriteFile(path, []byte("meta: [esto no: cierra"), 0o644)
		v := torreArquitectura(path)
		if cat(t, v) != "no-medido" || v["motivo"] == nil {
			t.Errorf("RN-13 roto: %+v", v)
		}
	})

	t.Run("modelo válido → con-arquitectura + nativo completo; clase explícita gana", func(t *testing.T) {
		path := filepath.Join(dir, "arquitectura.yaml")
		os.WriteFile(path, []byte(
			"meta:\n  id: demo-p9\n  clase: modelo\n  nombre: \"Demo\"\n  version: 0.1.0\n  proposito: probar\n"+
				"planos:\n  - { id: uno, nombre: \"Uno\", sub: \"banda\" }\n"+
				"tipos:\n  - { id: modulo, label: \"Módulo\" }\n"+
				"componentes:\n  - { id: a, tipo: modulo, plano: uno, nombre: \"A\", estado: activo, fichas: [DM-01] }\n"+
				"  - { id: b, tipo: modulo, plano: uno, nombre: \"B\", estado: activo, fichas: [DM-01] }\n"+
				"relaciones:\n  - { from: a, to: b, tipo: usa }\n"), 0o644)
		v := torreArquitectura(path)
		if cat(t, v) != "con-arquitectura" || v["clase"] != "modelo" || v["id"] != "demo-p9" {
			t.Fatalf("veredicto RN-25 roto: %+v", v)
		}
		if v["total_componentes"] != 2 || v["total_relaciones"] != 1 {
			t.Errorf("totales del nativo rotos: %+v", v)
		}
		if v["planos"] == nil || v["componentes"] == nil || v["relaciones"] == nil {
			t.Errorf("el modelo viaja inline como nativo (RN-25): %+v", v)
		}
	})

	t.Run("meta sin clase → default modelo (RN-22: el curado ES arquitectura-como-dato)", func(t *testing.T) {
		path := filepath.Join(dir, "sin-clase.yaml")
		os.WriteFile(path, []byte("meta:\n  id: x\nplanos: []\ntipos: []\ncomponentes: []\nrelaciones: []\n"), 0o644)
		v := torreArquitectura(path)
		if v["clase"] != "modelo" {
			t.Errorf("default de clase debe ser modelo (RN-22): %+v", v)
		}
	})

	t.Run("célula CON arquitectura.yaml en el workspace → el proyecto la trae", func(t *testing.T) {
		ws := t.TempDir()
		celula := filepath.Join(ws, "products", "conarq")
		os.MkdirAll(celula, 0o755)
		os.WriteFile(filepath.Join(celula, "LEDGER.md"), []byte("# Ledger\n"), 0o644)
		os.WriteFile(filepath.Join(celula, "arquitectura.yaml"), []byte(
			"meta: { id: conarq-p9 }\nplanos: []\ntipos: []\ncomponentes: []\nrelaciones: []\n"), 0o644)
		proyectos := torreProyectos(ws)
		if len(proyectos) != 1 {
			t.Fatalf("esperaba 1 proyecto: %+v", proyectos)
		}
		arq := proyectos[0]["arquitectura"].(map[string]any)
		if cat(t, arq) != "con-arquitectura" || arq["id"] != "conarq-p9" {
			t.Errorf("RN-20 roto: %+v", arq)
		}
	})
}

// ── la fila (RN-15/RN-17) ───────────────────────────────────────────────────

func TestTorreSistemasFilaNoDesaparece(t *testing.T) {
	dir := repoDemo(t)
	reg := &registry{Projects: []registryProject{
		{Name: "vivo", Nombre: "Sistema vivo", Path: dir, Active: true},
		{Name: "fantasma", Path: "/no/existe/en/disco", Active: true},
		{Name: "sin-ws", Path: "", Active: false},
	}}
	torreFactsCache = map[string]torreFactsEntry{}
	sistemas := torreSistemas(reg, false)
	if len(sistemas) != 3 {
		t.Fatalf("las filas no desaparecen (RN-17): esperaba 3, obtuve %d", len(sistemas))
	}
	for _, s := range sistemas[1:] {
		repo := s["repo"].(map[string]any)
		for _, col := range []string{"rama", "en_vuelo", "ultimo_tag", "gate_fabrica"} {
			if c := cat(t, repo[col].(map[string]any)); c != "no-medido" {
				t.Errorf("%s/%s: workspace ausente → todo no-medido (RN-17), obtuve %s", s["slug"], col, c)
			}
		}
	}
	if sistemas[0]["nombre"] != "Sistema vivo" || sistemas[1]["nombre"] != "fantasma" {
		t.Errorf("nombre aditivo con fallback al slug: %v / %v", sistemas[0]["nombre"], sistemas[1]["nombre"])
	}
}
