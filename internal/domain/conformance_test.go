package domain

import "testing"

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
