// handlers_lifecycle_test.go — fija el parser de tablas markdown del CIL (handlers_lifecycle.go).
// El parser (estados, severidad, carril, ref) es no trivial y alimenta /api/harness y /api/cil.
package main

import "testing"

func TestNormalizeEstado(t *testing.T) {
	cases := map[string]string{
		"reported":          "reported",
		"**Triaged**":       "triaged",       // strip emphasis + lowercase
		"verified ✅":        "verified",       // primera palabra alfabética
		"deferred (low)":    "deferred",
		"garbage":           "otro",
		"":                  "otro",
		"APPLIED":           "applied",
	}
	for raw, want := range cases {
		if got := normalizeEstado(raw); got != want {
			t.Errorf("normalizeEstado(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestStripEmphasis(t *testing.T) {
	cases := map[string]string{
		"**bold**":   "bold",
		"`code`":     "code",
		"__under__":  "under",
		"**A** `b`":  "A b",
		"plain":      "plain",
	}
	for in, want := range cases {
		if got := stripEmphasis(in); got != want {
			t.Errorf("stripEmphasis(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitTableRow(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"| a | b | c |", []string{"a", "b", "c"}},
		{"|x|", []string{"x"}},
		{"sin pipe", nil},
		{"   | trim | me |  ", []string{"trim", "me"}},
	}
	for _, c := range cases {
		got := splitTableRow(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitTableRow(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitTableRow(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestParseLifecycleTable(t *testing.T) {
	md := "" +
		"| ID | Fecha | Sev | Item | Estado | Ref |\n" +
		"|---|---|---|---|---|---|\n" +
		"| HB-7 | 2026-06-01 | 🔴 | **Algo crítico** [L2] | reported | docs/x.md |\n" +
		"| TD-3 | 2026-06-02 | 🟡 | otra cosa | triaged | docs/y.md |\n" + // id no-HB → se ignora
		"texto suelto que no es fila\n"

	items := parseLifecycleTable(md, hbIDRe, "HB-", harnessSevMap, true)
	if len(items) != 1 {
		t.Fatalf("esperaba 1 item HB (header/separator/TD/texto ignorados), got %d", len(items))
	}
	it := items[0]
	checks := map[string]any{
		"id":       "HB-7",
		"num":      7,
		"fecha":    "2026-06-01",
		"sevEmoji": "🔴",
		"sevLabel": "silent-killer",
		"estado":   "reported",
		"title":    "Algo crítico [L2]",
		"carril":   "L2",
		"ref":      "docs/x.md",
	}
	for k, want := range checks {
		if it[k] != want {
			t.Errorf("item[%q] = %v, want %v", k, it[k], want)
		}
	}
}

func TestCountByEstadoAndOpen(t *testing.T) {
	items := []lifecycleItem{
		{"estado": "reported"}, {"estado": "reported"}, {"estado": "verified"},
		{"estado": "applied"}, {"estado": "deferred"},
	}
	counts := countByEstado(items)
	if counts["reported"] != 2 || counts["verified"] != 1 || counts["applied"] != 1 || counts["deferred"] != 1 {
		t.Errorf("countByEstado = %v", counts)
	}
	// open = reported(2) + applied(1); verified y deferred NO son open.
	if got := countOpen(items); got != 3 {
		t.Errorf("countOpen = %d, want 3 (reported×2 + applied×1)", got)
	}
}
