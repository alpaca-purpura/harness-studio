package ruleset

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// repoRoot walks up until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestLoadRuleset(t *testing.T) {
	rs, err := New(repoRoot(t)).Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rs.Checks) < 200 {
		t.Fatalf("expected the full ruleset (>200 checks), got %d", len(rs.Checks))
	}

	// knowledge: skills declares 17 checks (docs/architecture/knowledge/INDEX.md).
	if n := len(rs.ForElemento("skills")); n != 17 {
		t.Errorf("skills element: got %d checks, want 17", n)
	}
	// harness-profile declares 11.
	if n := len(rs.ForElemento("harness-profile")); n != 11 {
		t.Errorf("harness-profile element: got %d checks, want 11", n)
	}

	// Every check has an id, an elemento and a known mechanism.
	byMec := map[domain.Mecanismo]int{}
	for _, c := range rs.Checks {
		if c.ID == "" || c.Elemento == "" {
			t.Errorf("check with empty id/elemento: %+v", c)
		}
		if !c.Mecanismo.Valid() {
			t.Errorf("check %s/%s has invalid mechanism %q", c.Elemento, c.ID, c.Mecanismo)
		}
		byMec[c.Mecanismo]++
	}

	// The arch boundaries wire real arch-test enforcers; at least a few must be present.
	if byMec[domain.MecArchTest] < 4 {
		t.Errorf("expected several arch-test checks inferred, got %d", byMec[domain.MecArchTest])
	}
	t.Logf("ruleset: %d checks across %d elements; by mechanism: %v",
		len(rs.Checks), len(rs.Elementos()), byMec)
}
