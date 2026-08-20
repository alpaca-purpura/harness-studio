// proceso_test.go — fija el ROUND-TRIP del descriptor de proceso (I-77 · RN-29/RN-31):
// la instancia sdd-default que shipea el kit reproduce EXACTO el hardcode pre-F4 de
// gates.go; y un descriptor ALTERNO deriva correcto sin tocar código del motor (la
// prueba "otra empresa, otro descriptor, cero cambio de consola" — RN-33).
package main

import (
	"reflect"
	"regexp"
	"testing"
)

// El hardcode pre-F4, congelado como expectativa (si el kit cambia su instancia,
// ESTOS literales avisan que el comportamiento observable cambió).
var legacyStates = []string{
	"idea", "refining", "refined", "ready", "developing",
	"developed", "reviewing", "done", "parked", "dropped",
}

var legacyOperatorAllowed = map[string][]operatorTransition{
	"idea":     {{"refining", false}, {"parked", true}, {"dropped", true}},
	"refining": {{"idea", false}, {"parked", true}, {"dropped", true}},
	"parked":   {{"idea", false}},
}

var legacyStateOwners = map[string]string{
	"refined":    "/architect (cierra spec) o /pm-{sistema} (ratifica)",
	"ready":      "/architect (ready package completo)",
	"developing": "/dev-team (builder spawn)",
	"developed":  "/dev-team (validators GREEN)",
	"reviewing":  "/auditor (AUTO-HANDOFF desde developed)",
	"done":       "/pm-{sistema} (Fase F MERGE)",
}

func TestProcesoEmbebidoParsea(t *testing.T) {
	if proceso.Descriptor.ID != "sdd-default" {
		t.Errorf("descriptor.id = %q, want sdd-default", proceso.Descriptor.ID)
	}
	if proceso.Descriptor.Version != 2 {
		t.Errorf("descriptor.version = %d, want 2 (v2 = F6: verbos + nombres + wip)", proceso.Descriptor.Version)
	}
	if len(proceso.Gates) != 3 {
		t.Errorf("gates = %d, want 3 (razon-de-cierre · verificacion-operador · merge-gate)", len(proceso.Gates))
	}
	for _, g := range proceso.Gates {
		if len(g.Checklist) == 0 {
			t.Errorf("gate %s sin checklist — el gate ES su checklist", g.ID)
		}
	}
}

func TestProcesoRoundTripLegacy(t *testing.T) {
	if !reflect.DeepEqual(storyStates, legacyStates) {
		t.Errorf("storyStates derivados ≠ hardcode pre-F4:\n got  %v\n want %v", storyStates, legacyStates)
	}
	if !reflect.DeepEqual(operatorAllowed, legacyOperatorAllowed) {
		t.Errorf("operatorAllowed derivado ≠ hardcode pre-F4:\n got  %v\n want %v", operatorAllowed, legacyOperatorAllowed)
	}
	if !reflect.DeepEqual(stateOwners, legacyStateOwners) {
		t.Errorf("stateOwners derivado ≠ hardcode pre-F4 (byte-a-byte):\n got  %v\n want %v", stateOwners, legacyStateOwners)
	}
	if transitionReasonMinLen != 10 {
		t.Errorf("transitionReasonMinLen = %d, want 10", transitionReasonMinLen)
	}
	// Terminalidad por categoría == terminalidad pre-F4 por nombre.
	for _, s := range legacyStates {
		want := s == "done" || s == "dropped"
		if got := isTerminalState(s); got != want {
			t.Errorf("isTerminalState(%q) = %v, want %v", s, got, want)
		}
	}
}

// Un descriptor de OTRA empresa: estados inventados, mismas categorías del contrato.
// El motor deriva todo sin un solo literal suyo en el código (RN-33).
const descriptorAlterno = `
descriptor: { id: cto-prod, nombre: Ciclo CTO con paso a producción, version: 1 }
estados:
  - { id: triage,   categoria: propuesto, inicial: true }
  - { id: build,    categoria: en-progreso }
  - { id: staging,  categoria: en-progreso }
  - { id: prod,     categoria: completado }
  - { id: killed,   categoria: descartado }
  - { id: frozen,   categoria: pausado }
transiciones:
  - { de: triage, a: build,  ejecutor: operador }
  - { de: triage, a: killed, ejecutor: operador, requiere_razon: true }
  - { de: build,  a: staging, ejecutor: rol }
  - { de: staging, a: prod,  ejecutor: rol }
gates:
  - { id: go-live, nombre: Paso a producción, momento: staging,
      autoridad: { rol: sre, arnes: "/sre" },
      checklist: ["smoke en staging verde", "rollback ensayado"] }
duenos:
  staging:
    - { rol: release-eng, arnes: "/release", nota: "empaqueta" }
    - { rol: sre, arnes: "/sre", nota: "verifica" }
parametros: { razon_minima: 5 }
`

func TestProcesoDescriptorAlterno(t *testing.T) {
	p, err := parseProceso([]byte(descriptorAlterno))
	if err != nil {
		t.Fatalf("descriptor alterno no parsea: %v", err)
	}
	wantStates := []string{"triage", "build", "staging", "prod", "killed", "frozen"}
	if !reflect.DeepEqual(p.estadoIDs(), wantStates) {
		t.Errorf("estadoIDs = %v, want %v", p.estadoIDs(), wantStates)
	}
	// terminalidad SIEMPRE por categoría — nombres que gates.go jamás vio.
	for estado, want := range map[string]bool{
		"prod": true, "killed": true, "triage": false, "staging": false, "frozen": false, "bogus": false,
	} {
		if got := p.esTerminal(estado); got != want {
			t.Errorf("esTerminal(%q) = %v, want %v", estado, got, want)
		}
	}
	ops := p.transicionesOperador()
	if len(ops) != 1 || len(ops["triage"]) != 2 {
		t.Errorf("transicionesOperador = %v, want triage con 2", ops)
	}
	if ops["triage"][1] != (operatorTransition{to: "killed", requiresReason: true}) {
		t.Errorf("triage→killed debe requerir razón: %v", ops["triage"])
	}
	// bindings múltiples renderizan con " o " — misma regla que el default.
	if got, want := p.duenosRender()["staging"], "/release (empaqueta) o /sre (verifica)"; got != want {
		t.Errorf("duenosRender(staging) = %q, want %q", got, want)
	}
}

// F6 (RN-45/RN-47/RN-48): los verbos nombran EVENTOS, el wip viaja en el estado.
func TestProcesoVerbos(t *testing.T) {
	verboRe := regexp.MustCompile(`^[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$`)
	for _, tr := range proceso.Transiciones {
		if tr.Verbo == "" {
			t.Errorf("transición %s→%s sin verbo — la instancia sdd-default los nombra todos", tr.De, tr.A)
			continue
		}
		if !verboRe.MatchString(tr.Verbo) {
			t.Errorf("verbo %q no cumple <sujeto>.<predicado>", tr.Verbo)
		}
	}
	for _, c := range []struct{ de, a, want string }{
		{"idea", "refining", "spec.started"},
		{"refining", "refined", "story.refined"},
		{"reviewing", "done", "story.merged"},
		{"parked", "idea", "story.reactivated"},
		{"done", "idea", ""}, // transición inexistente → sin evento
	} {
		if got := proceso.verboDe(c.de, c.a); got != c.want {
			t.Errorf("verboDe(%s, %s) = %q, want %q", c.de, c.a, got, c.want)
		}
	}
	// El verbo NO es único: nombra el evento, no la transición (story.parked ×2).
	parked := 0
	for _, tr := range proceso.Transiciones {
		if tr.Verbo == "story.parked" {
			parked++
		}
	}
	if parked != 2 {
		t.Errorf("story.parked aparece %d veces, want 2 (desde idea y refining)", parked)
	}
	// WIP como dato del estado (RN-48): 0 = sin límite.
	wips := map[string]int{}
	for _, e := range proceso.Estados {
		wips[e.ID] = e.Wip
	}
	for estado, want := range map[string]int{"refining": 3, "refined": 5, "developed": 1, "idea": 0, "done": 0} {
		if wips[estado] != want {
			t.Errorf("wip(%s) = %d, want %d", estado, wips[estado], want)
		}
	}
}

func TestParseProcesoInvalido(t *testing.T) {
	cases := map[string]string{
		"sin estados":        `descriptor: {id: x, nombre: x, version: 1}`,
		"estado duplicado":   `{descriptor: {id: x, nombre: x, version: 1}, estados: [{id: a, categoria: propuesto, inicial: true}, {id: a, categoria: pausado}], parametros: {razon_minima: 5}}`,
		"categoria inválida": `{descriptor: {id: x, nombre: x, version: 1}, estados: [{id: a, categoria: bogus, inicial: true}], parametros: {razon_minima: 5}}`,
		"sin inicial":        `{descriptor: {id: x, nombre: x, version: 1}, estados: [{id: a, categoria: propuesto}], parametros: {razon_minima: 5}}`,
		"transición rota":    `{descriptor: {id: x, nombre: x, version: 1}, estados: [{id: a, categoria: propuesto, inicial: true}], transiciones: [{de: a, a: zz, ejecutor: operador}], parametros: {razon_minima: 5}}`,
		"razon_minima 0":     `{descriptor: {id: x, nombre: x, version: 1}, estados: [{id: a, categoria: propuesto, inicial: true}]}`,
	}
	for name, raw := range cases {
		if _, err := parseProceso([]byte(raw)); err == nil {
			t.Errorf("%s: parseProceso aceptó un descriptor inválido", name)
		}
	}
}
