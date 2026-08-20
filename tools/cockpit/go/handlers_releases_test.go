// handlers_releases_test.go — fija el merge-gate G7/G8 (classifyReleaseStories) con
// un lector de estado inyectado, sin tocar disco. Antes esta regla solo se ejercía
// por HTTP + filesystem real.
package main

import "testing"

func TestClassifyReleaseStories(t *testing.T) {
	// fake reader: estado por id; "" = story no encontrada.
	fake := map[string]string{"s1": "done", "s2": "developing", "s3": "dropped", "s4": ""}
	read := func(sistema, id string) string { return fake[id] }

	cases := []struct {
		name        string
		ids         []any
		wantStates  []string
		wantNonTerm []string
	}{
		{
			name:        "mezcla: terminal · no-terminal · sin estado",
			ids:         []any{"s1", "s2", "s3", "s4"},
			wantStates:  []string{"done", "developing", "dropped"}, // "" no entra a states
			wantNonTerm: []string{"s2", "s4"},                       // developing + "" bloquean
		},
		{
			name:        "todas terminales → listo para merge",
			ids:         []any{"s1", "s3"},
			wantStates:  []string{"done", "dropped"},
			wantNonTerm: nil,
		},
		{
			name:        "story desconocida bloquea",
			ids:         []any{"s1", "fantasma"},
			wantStates:  []string{"done"},
			wantNonTerm: []string{"fantasma"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotStates, gotNonTerm := classifyReleaseStories("acme", c.ids, read)
			assertSeq(t, "states", gotStates, c.wantStates)
			assertSeq(t, "nonTerminal", gotNonTerm, c.wantNonTerm)
		})
	}
}

// assertSeq compara dos slices en ORDEN (classifyReleaseStories preserva el orden de ids).
func assertSeq(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}
