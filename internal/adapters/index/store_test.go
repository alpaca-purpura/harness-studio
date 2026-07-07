package index

import (
	"context"
	"errors"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestSeedServesDogfood asserts the index seeds the real dogfood arnés (dev-full-cycle)
// so the Map endpoint serves it (HS-09 Hito 1, architecture §4.1): 5 nodes, 4 edges,
// the spec-writer box carries its contract (caja + estado).
func TestSeedServesDogfood(t *testing.T) {
	s := New()

	g, err := s.Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query(dev-full-cycle) = %v, want nil", err)
	}
	if g.Arnes == nil || g.Arnes.ID != "dev-full-cycle" {
		t.Fatalf("Arnes.ID = %v, want dev-full-cycle", g.Arnes)
	}
	if got := len(g.Nodes); got != 5 {
		t.Errorf("len(Nodes) = %d, want 5", got)
	}
	if got := len(g.Edges); got != 4 {
		t.Errorf("len(Edges) = %d, want 4", got)
	}

	spec, ok := g.NodeByID("spec-writer")
	if !ok {
		t.Fatal("NodeByID(spec-writer) not found")
	}
	if !spec.IsCaja() {
		t.Error("spec-writer.IsCaja() = false, want true (contract.caja)")
	}
	if spec.Estado != "idea -> spec" {
		t.Errorf("spec-writer.Estado = %q, want %q", spec.Estado, "idea -> spec")
	}
	if spec.Clase != domain.ClaseSkill {
		t.Errorf("spec-writer.Clase = %q, want skill", spec.Clase)
	}
}

// TestSeedFases asserts the dogfood declares its four phases in order (the lane order the
// Map renders, RF-12) and that the base-band rule is present.
func TestSeedFases(t *testing.T) {
	g, err := New().Query(context.Background(), "dev-full-cycle")
	if err != nil {
		t.Fatalf("Query(dev-full-cycle) = %v", err)
	}
	want := []string{"spec", "build", "review", "release"}
	if len(g.Arnes.Fases) != len(want) {
		t.Fatalf("Fases = %v, want %v", g.Arnes.Fases, want)
	}
	for i, f := range want {
		if string(g.Arnes.Fases[i]) != f {
			t.Errorf("Fases[%d] = %q, want %q", i, g.Arnes.Fases[i], f)
		}
	}
	if std, ok := g.NodeByID("std-spec"); !ok || std.Banda != domain.BandaBase {
		t.Errorf("std-spec banda = %v, want base", std.Banda)
	}
}

// TestSeedServesShowcase asserts the index seeds the showcase arnés (content-studio-full,
// HS-09 Hito 2): a maximal graph whose rol is Editorial (agnostic-to-rubro proof) and that
// exercises all 10 clases so the Map can render every casuistic. Conformance-valid.
func TestSeedServesShowcase(t *testing.T) {
	g, err := New().Query(context.Background(), "content-studio-full")
	if err != nil {
		t.Fatalf("Query(content-studio-full) = %v, want nil", err)
	}
	if g.Arnes == nil || g.Arnes.Rol != "Editorial · Content Lead" {
		t.Fatalf("showcase rol = %v, want Editorial · Content Lead", g.Arnes)
	}
	seen := map[domain.Clase]bool{}
	for _, n := range g.Nodes {
		seen[n.Clase] = true
	}
	for _, c := range []domain.Clase{
		domain.ClaseSkill, domain.ClaseSubagent, domain.ClaseHook, domain.ClaseRule,
		domain.ClaseCommand, domain.ClaseMCP, domain.ClasePlugin, domain.ClaseSettings,
		domain.ClaseOutputStyle, domain.ClaseStatusline,
	} {
		if !seen[c] {
			t.Errorf("showcase missing clase %q — the kitchen-sink must exercise all 10", c)
		}
	}
}

// TestQueryUnknown asserts an unknown harness id is a clean ErrNotFound (the endpoint
// maps it to 404), not a panic or empty graph.
func TestQueryUnknown(t *testing.T) {
	s := New()
	if _, err := s.Query(context.Background(), "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Query(nope) err = %v, want ErrNotFound", err)
	}
}

// TestListPortfolio asserts List returns every seeded harness ordered by id (the picker's
// stable portfolio, RF-72) and includes the real dogfood arnés.
func TestListPortfolio(t *testing.T) {
	gs, err := New().List(context.Background())
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	ids := make([]string, 0, len(gs))
	for _, g := range gs {
		if g.Arnes != nil {
			ids = append(ids, g.Arnes.ID)
		}
	}
	// New() seeds demo + dev-full-cycle + content-studio-full; sorted by id.
	want := []string{"content-studio-full", "demo", "dev-full-cycle"}
	if len(ids) != len(want) {
		t.Fatalf("List ids = %v, want %v", ids, want)
	}
	for i, w := range want {
		if ids[i] != w {
			t.Errorf("List ids = %v, want %v", ids, want)
			break
		}
	}
}
