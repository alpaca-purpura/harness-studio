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

	// Los 4 checks del boundary del Portafolio referencian tests COLOCADOS junto al
	// código (ruta repo-relativa, no arch_test.go) — deben clasificar arch-test, no
	// caer a nl-judge (si cayeran, el motor los reportaría deferred pese a que el
	// enforcer existe y corre en CI).
	for _, c := range rs.ForElemento("portafolio-identidad-y-deriva-honesta") {
		if c.Mecanismo != domain.MecArchTest {
			t.Errorf("check %s: mecanismo %q, quiero arch-test (test colocado)", c.ID, c.Mecanismo)
		}
	}
	t.Logf("ruleset: %d checks across %d elements; by mechanism: %v",
		len(rs.Checks), len(rs.Elementos()), byMec)
}

func TestInferMechanism(t *testing.T) {
	cases := []struct {
		enforcer string
		mec      domain.Mecanismo
		enf      string
	}{
		{"arch_test.go:TestCoreNoImportaShell", domain.MecArchTest, "arch_test.go:TestCoreNoImportaShell"},
		{"`fitness/arch_test.go:TestX`", domain.MecArchTest, "arch_test.go:TestX"},
		{
			"internal/adapters/portafolio/store_test.go:TestStoreDegradaHonesto", domain.MecArchTest,
			"internal/adapters/portafolio/store_test.go:TestStoreDegradaHonesto",
		},
		{
			"docs/architecture/fitness/capability_trace_test.go:TestCapabilityCoverage", domain.MecArchTest,
			"docs/architecture/fitness/capability_trace_test.go:TestCapabilityCoverage",
		},
		{"go-arch-lint check", domain.MecGoArchLint, "go-arch-lint"},
		{"graph.l0.schema.json", domain.MecSchema, "graph.l0.schema.json"},
		{"golangci-lint run", domain.MecNLJudge, ""},
		{"L1.2", domain.MecNLJudge, ""},
	}
	for _, c := range cases {
		mec, enf := inferMechanism(c.enforcer)
		if mec != c.mec || enf != c.enf {
			t.Errorf("inferMechanism(%q) = (%q, %q), quiero (%q, %q)", c.enforcer, mec, enf, c.mec, c.enf)
		}
	}
}
