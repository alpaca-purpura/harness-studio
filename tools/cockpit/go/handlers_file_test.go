// handlers_file_test.go — fija la defensa del file-API (handlers_file.go): el núcleo
// de seguridad path-traversal + el whitelist, ambos puros. Código sensible: de él
// depende que /api/file no lea fuera del workspace.
package main

import "testing"

func TestIsTraversalSafe(t *testing.T) {
	cases := []struct {
		in        string
		wantClean string
		wantOK    bool
	}{
		{"docs/ARCHITECTURE.md", "docs/ARCHITECTURE.md", true},
		{"./docs/x.md", "docs/x.md", true},
		{"client-acme/docs/SYSTEM-MAP.md", "client-acme/docs/SYSTEM-MAP.md", true},
		{"foo/../bar", "bar", true}, // se normaliza dentro del root → seguro
		{"", "", false},
		{"/etc/passwd", "", false},          // absoluto
		{"../secrets", "", false},           // parent traversal
		{"a/../../b", "", false},            // escapa: Clean → ../b
		{"docs/../../etc/passwd", "", false}, // escapa por subir de más
		{"docs/sub/../../../etc", "", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			clean, ok := isTraversalSafe(c.in)
			if ok != c.wantOK {
				t.Fatalf("isTraversalSafe(%q) ok = %v, want %v", c.in, ok, c.wantOK)
			}
			if ok && clean != c.wantClean {
				t.Errorf("isTraversalSafe(%q) clean = %q, want %q", c.in, clean, c.wantClean)
			}
		})
	}
}

func TestSegmentWhitelisted(t *testing.T) {
	sistemas := []string{"client-acme", "client-xyz"}
	cases := []struct {
		in   string
		want bool
	}{
		{"client-acme/docs/SYSTEM-MAP.md", true}, // sistema + docs
		{".claude/skills/x/SKILL.md", true},      // .claude
		{"docs/ARCHITECTURE.md", true},           // docs raíz
		{"docs", true},
		{".claude", true},
		{"client-acme/specs/x.md", false}, // sistema pero no docs
		{"client-acme", false},            // sistema sin segundo segmento
		{"random/x.md", false},            // no whitelisteado
		{"node_modules/evil.js", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := segmentWhitelisted(c.in, sistemas); got != c.want {
				t.Errorf("segmentWhitelisted(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
