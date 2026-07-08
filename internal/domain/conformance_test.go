package domain

import (
	"strings"
	"testing"
)

func ptrBool(b bool) *bool { return &b }

func TestVerificarEscritorUnico(t *testing.T) {
	// Two cajas writing the same art (both default single-writer) → fail.
	g := Graph{Nodes: []Box{
		{ID: "a", Contract: &Contract{Caja: true, Entrega: []Output{{Art: "spec.md"}}}},
		{ID: "b", Contract: &Contract{Caja: true, Entrega: []Output{{Art: "spec.md"}}}},
	}}
	if r := VerificarEscritorUnico(g); r.Veredicto != VeredictoFail {
		t.Errorf("two writers of spec.md: want fail, got %s", r.Veredicto)
	}

	// One explicitly relinquishes escritor_unico → no conflict.
	g2 := Graph{Nodes: []Box{
		{ID: "a", Contract: &Contract{Caja: true, Entrega: []Output{{Art: "spec.md"}}}},
		{ID: "b", Contract: &Contract{Caja: true, Entrega: []Output{{Art: "spec.md", EscritorUnico: ptrBool(false)}}}},
	}}
	if r := VerificarEscritorUnico(g2); r.Veredicto != VeredictoPass {
		t.Errorf("one writer relinquishes: want pass, got %s (%s)", r.Veredicto, r.Detalle)
	}
}

func TestVerificarSpineAutoConsistente(t *testing.T) {
	// A transition endpoint not in `estados` is a self-inconsistency.
	g := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Estados: []string{"a", "b"},
		Transiciones: []Transicion{{De: "a", A: "b"}, {De: "b", A: "phantom"}},
	}}}
	var got Veredicto
	for _, r := range VerificarSpine(g) {
		if r.Check.ID == "spine-auto-consistente" {
			got = r.Veredicto
		}
	}
	if got != VeredictoFail {
		t.Errorf("phantom transition endpoint: want fail, got %s", got)
	}

	// A coherent spine passes.
	g2 := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Terminales: []string{"b"}, Estados: []string{"a", "b"},
		Transiciones: []Transicion{{De: "a", A: "b"}},
	}}}
	for _, r := range VerificarSpine(g2) {
		if r.Check.ID == "spine-auto-consistente" && r.Veredicto != VeredictoPass {
			t.Errorf("coherent spine: want pass, got %s (%s)", r.Veredicto, r.Detalle)
		}
	}
}

func TestVerificarSpineCategorias(t *testing.T) {
	spineVeredicto := func(g Graph, id string) (Veredicto, string) {
		for _, r := range VerificarSpine(g) {
			if r.Check.ID == id {
				return r.Veredicto, r.Detalle
			}
		}
		t.Fatalf("check %q ausente", id)
		return "", ""
	}

	// Sin categorías → ambos checks difieren honesto (mapa opcional, HS-12).
	sinCat := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Estados: []string{"a", "b"},
	}}}
	for _, id := range []string{"categoria-estado-existe", "terminal-categoria-coherente"} {
		if v, _ := spineVeredicto(sinCat, id); v != VeredictoDiferido {
			t.Errorf("%s sin categorías: want diferido, got %s", id, v)
		}
	}

	// Clave fuera de estados + valor fuera del enum → fail.
	mal := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Estados: []string{"a", "b"},
		Categorias: map[string]Categoria{"a": "inventada", "fantasma": CategoriaPausado},
	}}}
	if v, d := spineVeredicto(mal, "categoria-estado-existe"); v != VeredictoFail {
		t.Errorf("categorías rotas: want fail, got %s (%s)", v, d)
	}

	// Terminal con categoría no terminal (o sin categoría) → fail; coherente → pass.
	tMal := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Terminales: []string{"b"}, Estados: []string{"a", "b"},
		Categorias: map[string]Categoria{"a": CategoriaPropuesto, "b": CategoriaEnProgreso},
	}}}
	if v, d := spineVeredicto(tMal, "terminal-categoria-coherente"); v != VeredictoFail {
		t.Errorf("terminal en-progreso: want fail, got %s (%s)", v, d)
	}
	ok := Graph{Arnes: &Arnes{Spine: &Spine{
		Inicial: "a", Terminales: []string{"b"}, Estados: []string{"a", "b"},
		Categorias: map[string]Categoria{"a": CategoriaPropuesto, "b": CategoriaCompletado},
	}}}
	for _, id := range []string{"categoria-estado-existe", "terminal-categoria-coherente"} {
		if v, d := spineVeredicto(ok, id); v != VeredictoPass {
			t.Errorf("%s coherente: want pass, got %s (%s)", id, v, d)
		}
	}
}

func TestAvanzarCaja(t *testing.T) {
	cases := []struct {
		name   string
		actual EstadoCaja
		senal  SenalIteracion
		want   EstadoCaja
	}{
		{"error subtype blocks", CajaWorking, SenalIteracion{Subtipo: SubtipoError}, CajaBlocked},
		{"blocked artifact blocks", CajaDraft, SenalIteracion{EstadoArtefacto: "blocked"}, CajaBlocked},
		{"done artifact finishes", CajaWorking, SenalIteracion{EstadoArtefacto: "done"}, CajaDone},
		{"otherwise keeps working", CajaDraft, SenalIteracion{Subtipo: SubtipoLimite, EstadoArtefacto: "wip"}, CajaWorking},
		{"terminal is sticky", CajaDone, SenalIteracion{EstadoArtefacto: "blocked"}, CajaDone},
	}
	for _, c := range cases {
		if got := AvanzarCaja(c.actual, c.senal); got != c.want {
			t.Errorf("%s: AvanzarCaja(%s,%+v) = %s, want %s", c.name, c.actual, c.senal, got, c.want)
		}
	}
}

func TestRutaSiguiente(t *testing.T) {
	ct := &Contract{
		Ruta:    []Route{{A: "builder", Si: "verde"}, {A: "revisor"}},
		Handoff: &Handoff{Cuando: "no converge", A: "humano"},
	}
	// blocked → handoff target.
	if got := RutaSiguiente(ct, CajaBlocked, SenalIteracion{}); got != "humano" {
		t.Errorf("blocked routes to %q, want humano", got)
	}
	// done with a matching `si` → the conditional route.
	if got := RutaSiguiente(ct, CajaDone, SenalIteracion{EstadoArtefacto: "verde"}); got != "builder" {
		t.Errorf("done+verde routes to %q, want builder", got)
	}
	// done without a matching `si` → the happy path (no `si`).
	if got := RutaSiguiente(ct, CajaDone, SenalIteracion{EstadoArtefacto: "otro"}); got != "revisor" {
		t.Errorf("done+other routes to %q, want revisor (happy path)", got)
	}
	// non-terminal → no routing yet.
	if got := RutaSiguiente(ct, CajaWorking, SenalIteracion{}); got != "" {
		t.Errorf("working routes to %q, want empty", got)
	}
}

func TestRequiereDocumentAsCache(t *testing.T) {
	// abierto is exempt even at T2/T3 (B8 resolution: arquetipo > perfil).
	if RequiereDocumentAsCache(ArqAbierto, PerfilT3) {
		t.Error("arquetipo abierto must be EXEMPT from strict document-as-cache")
	}
	// pipeline/excepcion at T2/T3 require it.
	if !RequiereDocumentAsCache(ArqPipeline, PerfilT2) {
		t.Error("pipeline T2 must require document-as-cache")
	}
	// T1 never requires it.
	if RequiereDocumentAsCache(ArqPipeline, PerfilT1) {
		t.Error("T1 must not require document-as-cache")
	}
}

// ── composición del cableado (franja-artefactos, RF-100..104) ────────────────────

// cajaCon builds a minimal process box for composition tests.
func cajaCon(id, estado string, necesita []Input, entrega []Output, ruta []Route) Box {
	return Box{ID: id, Contract: &Contract{
		Caja: true, Estado: estado, Necesita: necesita, Entrega: entrega, Ruta: ruta,
	}}
}

func TestVerificarSinHuerfanos(t *testing.T) {
	// Orphan: b needs an art from caja:a that a does not deliver → warn/fail.
	g := Graph{Nodes: []Box{
		cajaCon("a", "", nil, []Output{{Art: "otra-cosa.md"}}, nil),
		cajaCon("b", "", []Input{{Art: "spec.md", De: "caja:a"}}, nil, nil),
	}}
	r := VerificarSinHuerfanos(g)
	if r.Veredicto != VeredictoFail {
		t.Errorf("huérfano: want fail, got %s (%s)", r.Veredicto, r.Detalle)
	}

	// External inputs are never orphans (usuario, terceros:*), and a wired pair passes.
	g2 := Graph{Nodes: []Box{
		cajaCon("a", "", []Input{{Art: "idea", De: "usuario"}, {Art: "factura.pdf", De: "terceros:proveedor"}},
			[]Output{{Art: "spec.md"}}, nil),
		cajaCon("b", "", []Input{{Art: "spec.md", De: "caja:a"}}, nil, nil),
	}}
	if r := VerificarSinHuerfanos(g2); r.Veredicto != VeredictoPass {
		t.Errorf("externos+cableado: want pass, got %s (%s)", r.Veredicto, r.Detalle)
	}

	// A refina output counts as producing the refined art (D9).
	g3 := Graph{Nodes: []Box{
		cajaCon("valida", "", []Input{{Art: "factura.pdf", De: "terceros:prov"}},
			[]Output{{Art: "factura.pdf", Refina: "factura.pdf"}}, nil),
		cajaCon("paga", "", []Input{{Art: "factura.pdf", De: "caja:valida"}}, nil, nil),
	}}
	if r := VerificarSinHuerfanos(g3); r.Veredicto != VeredictoPass {
		t.Errorf("consumo de revisión: want pass, got %s (%s)", r.Veredicto, r.Detalle)
	}
}

func TestVerificarDeadEnds(t *testing.T) {
	spine := &Arnes{Spine: &Spine{
		Inicial: "idea", Terminales: []string{"released"},
		Estados:      []string{"idea", "build", "released"},
		Transiciones: []Transicion{{De: "idea", A: "build"}, {De: "build", A: "released"}},
	}}

	// «notas de build» sin consumidor en caja NO terminal → fail (RF-101).
	g := Graph{Arnes: spine, Nodes: []Box{
		cajaCon("build", "idea -> build", nil, []Output{{Art: "notas de build"}}, nil),
		cajaCon("release", "build -> released", []Input{}, []Output{{Art: "release@v"}}, nil),
	}}
	r := VerificarDeadEnds(g)
	if r.Veredicto != VeredictoFail || !strings.Contains(r.Detalle, "notas de build") {
		t.Errorf("dead-end real: want fail con 'notas de build', got %s (%s)", r.Veredicto, r.Detalle)
	}
	// La entrega de la caja terminal (release@v) NO es dead-end (C15) — no aparece.
	if strings.Contains(r.Detalle, "release@v") {
		t.Errorf("entrega terminal reportada como dead-end: %s", r.Detalle)
	}

	// Sin spine/terminales → deferred honesto.
	g2 := Graph{Nodes: []Box{cajaCon("a", "x -> y", nil, []Output{{Art: "algo"}}, nil)}}
	if r := VerificarDeadEnds(g2); r.Veredicto != VeredictoDiferido {
		t.Errorf("sin spine: want deferred, got %s", r.Veredicto)
	}
}

func TestVerificarRutaExiste(t *testing.T) {
	g := Graph{Nodes: []Box{
		cajaCon("a", "", nil, nil, []Route{{A: "fantasma"}, {A: "humano"}, {A: "b"}}),
		cajaCon("b", "", nil, nil, nil),
	}}
	r := VerificarRutaExiste(g)
	if r.Veredicto != VeredictoFail || !strings.Contains(r.Detalle, "fantasma") {
		t.Errorf("ruta colgante: want fail con 'fantasma', got %s (%s)", r.Veredicto, r.Detalle)
	}
	if strings.Contains(r.Detalle, "humano") {
		t.Errorf("humano es destino legal, no debe reportarse: %s", r.Detalle)
	}
}

func TestVerificarArtIdentidad(t *testing.T) {
	// Mismatch C21: a entrega spec.md pero b pide espec.md de a → error.
	g := Graph{Nodes: []Box{
		cajaCon("a", "", nil, []Output{{Art: "spec.md"}}, nil),
		cajaCon("b", "", []Input{{Art: "espec.md", De: "caja:a"}}, nil, nil),
	}}
	if r := VerificarArtIdentidad(g); r.Veredicto != VeredictoFail {
		t.Errorf("mismatch de nombres: want fail, got %s (%s)", r.Veredicto, r.Detalle)
	}

	// Productor inexistente NO es mismatch de identidad (lo cubre sin-huerfanos).
	g2 := Graph{Nodes: []Box{
		cajaCon("b", "", []Input{{Art: "spec.md", De: "caja:nadie"}}, nil, nil),
	}}
	if r := VerificarArtIdentidad(g2); r.Veredicto != VeredictoPass {
		t.Errorf("productor ausente: want pass aquí (huérfano es otro check), got %s (%s)", r.Veredicto, r.Detalle)
	}
}

func TestVerificarRefinaCoherente(t *testing.T) {
	// Cadena cobranza-like: terceros → valida (refina) → paga. Lineal, raíz externa → pass.
	ok := Graph{Nodes: []Box{
		cajaCon("valida", "", []Input{{Art: "factura.pdf", De: "terceros:prov"}},
			[]Output{{Art: "factura.pdf", Refina: "factura.pdf"}}, nil),
		cajaCon("paga", "", []Input{{Art: "factura.pdf", De: "caja:valida"}},
			[]Output{{Art: "comprobante.pdf"}}, nil),
	}}
	if r := VerificarRefinaCoherente(ok); r.Veredicto != VeredictoPass {
		t.Errorf("cadena legal: want pass, got %s (%s)", r.Veredicto, r.Detalle)
	}

	// Refina sin necesita (D9a) → fail.
	sinNec := Graph{Nodes: []Box{
		cajaCon("r", "", nil, []Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
	}}
	if r := VerificarRefinaCoherente(sinNec); r.Veredicto != VeredictoFail {
		t.Errorf("refina sin necesita: want fail, got %s (%s)", r.Veredicto, r.Detalle)
	}

	// Ciclo: r1 refina desde r2 y r2 refina desde r1 → fail.
	ciclo := Graph{Nodes: []Box{
		cajaCon("r1", "", []Input{{Art: "doc.md", De: "caja:r2"}},
			[]Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
		cajaCon("r2", "", []Input{{Art: "doc.md", De: "caja:r1"}},
			[]Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
	}}
	if r := VerificarRefinaCoherente(ciclo); r.Veredicto != VeredictoFail || !strings.Contains(r.Detalle, "ciclo") {
		t.Errorf("ciclo: want fail con 'ciclo', got %s (%s)", r.Veredicto, r.Detalle)
	}

	// Rama (no lineal): dos refinadores beben del mismo escritor → fail.
	rama := Graph{Nodes: []Box{
		cajaCon("raiz", "", nil, []Output{{Art: "doc.md"}}, nil),
		cajaCon("rA", "", []Input{{Art: "doc.md", De: "caja:raiz"}},
			[]Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
		cajaCon("rB", "", []Input{{Art: "doc.md", De: "caja:raiz"}},
			[]Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
	}}
	if r := VerificarRefinaCoherente(rama); r.Veredicto != VeredictoFail || !strings.Contains(r.Detalle, "no lineal") {
		t.Errorf("rama: want fail con 'no lineal', got %s (%s)", r.Veredicto, r.Detalle)
	}
}

func TestEscritorUnicoConRefina(t *testing.T) {
	// RF-104: dos entregas del mismo art SIN refina = error (como hoy).
	sin := Graph{Nodes: []Box{
		cajaCon("a", "", nil, []Output{{Art: "doc.md"}}, nil),
		cajaCon("b", "", nil, []Output{{Art: "doc.md"}}, nil),
	}}
	if r := VerificarEscritorUnico(sin); r.Veredicto != VeredictoFail {
		t.Errorf("2 escritores sin refina: want fail, got %s", r.Veredicto)
	}

	// CON refina = cadena legal: el refinador sale del conteo de escritor-unico.
	con := Graph{Nodes: []Box{
		cajaCon("a", "", nil, []Output{{Art: "doc.md"}}, nil),
		cajaCon("b", "", []Input{{Art: "doc.md", De: "caja:a"}},
			[]Output{{Art: "doc.md", Refina: "doc.md"}}, nil),
	}}
	if r := VerificarEscritorUnico(con); r.Veredicto != VeredictoPass {
		t.Errorf("cadena refina: want pass, got %s (%s)", r.Veredicto, r.Detalle)
	}
}
