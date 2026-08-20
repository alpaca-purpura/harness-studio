// gates_test.go — fija el comportamiento de los gates SDD (dominio puro de gates.go).
// Antes estas reglas (G1-G2 transiciones · G7-G8 terminalidad · políticas de edición)
// vivían inline en los handlers HTTP y solo se podían probar por e2e de browser.
package main

import (
	"sort"
	"strings"
	"testing"
)

func TestEvaluateOperatorTransition(t *testing.T) {
	cases := []struct {
		name              string
		from, to, reason  string
		want              transitionVerdict
	}{
		{"idea→refining sin reason", "idea", "refining", "", transitionOK},
		{"refining→idea", "refining", "idea", "", transitionOK},
		{"parked→idea", "parked", "idea", "", transitionOK},
		{"idea→parked con reason", "idea", "parked", "se pausa por presupuesto", transitionOK},
		{"idea→dropped reason exacto 10", "idea", "dropped", "1234567890", transitionOK},
		{"idea→parked reason corto", "idea", "parked", "corto", transitionNeedsReason},
		{"idea→parked sin reason", "idea", "parked", "", transitionNeedsReason},
		{"idea→dropped reason solo espacios", "idea", "dropped", "          ", transitionNeedsReason},
		{"idea→developing (estado de rol)", "idea", "developing", "", transitionForbidden},
		{"refined→done (no operador)", "refined", "done", "", transitionForbidden},
		{"idea→done", "idea", "done", "", transitionForbidden},
		{"estado desconocido", "bogus", "refining", "", transitionForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := evaluateOperatorTransition(c.from, c.to, c.reason); got != c.want {
				t.Errorf("evaluateOperatorTransition(%q,%q,%q) = %d, want %d", c.from, c.to, c.reason, got, c.want)
			}
		})
	}
}

func TestTransitionOwnerNote(t *testing.T) {
	if note := transitionOwnerNote("developing"); !strings.Contains(note, "/dev-team") {
		t.Errorf("owner note de 'developing' debe nombrar al rol, got %q", note)
	}
	if note := transitionOwnerNote("done"); !strings.Contains(note, "/pm-") {
		t.Errorf("owner note de 'done' debe nombrar al pm, got %q", note)
	}
	if note := transitionOwnerNote("refining"); note != "transition no permitida desde el cockpit." {
		t.Errorf("target sin owner debe dar el mensaje genérico, got %q", note)
	}
}

func TestIsTerminalState(t *testing.T) {
	cases := map[string]bool{
		"done": true, "dropped": true,
		"idea": false, "reviewing": false, "parked": false, "": false,
	}
	for state, want := range cases {
		if got := isTerminalState(state); got != want {
			t.Errorf("isTerminalState(%q) = %v, want %v", state, got, want)
		}
	}
}

func TestIsStoryState(t *testing.T) {
	for _, s := range storyStates {
		if !isStoryState(s) {
			t.Errorf("isStoryState(%q) = false, debería ser un estado válido", s)
		}
	}
	if isStoryState("bogus") {
		t.Error("isStoryState(\"bogus\") = true, debería ser false")
	}
}

func TestForbiddenStoryFields(t *testing.T) {
	cases := []struct {
		name   string
		fields map[string]any
		want   []string
	}{
		{"todos editables", map[string]any{"release": "v1", "priority": 1, "goal": "x", "anti": "y", "reuse": "z"}, nil},
		{"uno solo editable", map[string]any{"goal": "x"}, nil},
		{"prohibido + editable", map[string]any{"state": "done", "goal": "x"}, []string{"state"}},
		{"varios prohibidos", map[string]any{"state": "done", "story_id": "s1"}, []string{"state", "story_id"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertSameSet(t, forbiddenStoryFields(c.fields), c.want)
		})
	}
}

func TestForbiddenCapFields(t *testing.T) {
	cases := []struct {
		name string
		body map[string]any
		want []string
	}{
		{"status+reason permitidos", map[string]any{"status": "live", "reason": "x"}, nil},
		{"estructural", map[string]any{"change_log": []any{}}, []string{"change_log"}},
		{"varios estructurales", map[string]any{"scenarios": 1, "slug": "s", "status": "live"}, []string{"scenarios", "slug"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertSameSet(t, forbiddenCapFields(c.body), c.want)
		})
	}
}

func TestIsValidCapStatus(t *testing.T) {
	for _, s := range []string{"live", "beta", "deprecated", "sunset"} {
		if !isValidCapStatus(s) {
			t.Errorf("isValidCapStatus(%q) = false, debería ser válido", s)
		}
	}
	for _, s := range []string{"bogus", "", "LIVE", "shipped"} {
		if isValidCapStatus(s) {
			t.Errorf("isValidCapStatus(%q) = true, debería ser inválido", s)
		}
	}
}

// assertSameSet compara dos slices como conjuntos (el orden de iteración de un map
// en Go no es determinista, así que forbidden*Fields no garantiza orden).
func assertSameSet(t *testing.T, got, want []string) {
	t.Helper()
	g := append([]string(nil), got...)
	w := append([]string(nil), want...)
	sort.Strings(g)
	sort.Strings(w)
	if len(g) != len(w) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range g {
		if g[i] != w[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
