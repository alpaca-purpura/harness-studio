package usecase_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

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

func newService(t *testing.T) (*usecase.ConformanceService, string) {
	t.Helper()
	root := repoRoot(t)
	adapters := []ports.MechanismAdapter{
		mechanism.NewArchTest(root), mechanism.NewGoArchLint(root),
		mechanism.NLJudge{},
		mechanism.StaticScan{},
		mechanism.SchemaAdapter{},
	}
	schemas := mechanism.NewSchemaSet(filepath.Join(root, "docs", "architecture", "contracts", "schema"))
	return usecase.NewConformanceService(root, ruleset.New(root), schemas, adapters), root
}

// TestDogfoodArnesConforms is G1: the dev-full-cycle dogfood arnés validates green — every
// fused box contract passes and the spine consistency holds against the arnés's own spine.
func TestDogfoodArnesConforms(t *testing.T) {
	svc, root := newService(t)
	rep, err := svc.Run(context.Background(), ports.Target{
		Kind: ports.TargetArnes, GraphPath: filepath.Join(root, "dogfood", "dev-full-cycle.graph.json"),
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !rep.OK() {
		for _, r := range rep.Results {
			if r.Veredicto == domain.VeredictoFail || r.Veredicto == domain.VeredictoError {
				t.Errorf("blocking %s: %s — %s", r.Veredicto, r.Check.ID, r.Detalle)
			}
		}
		t.Fatal("dogfood arnés must conform (G1)")
	}
	// Every fused axis must be recognized: the box-contract checks ran and passed.
	var cajaChecks int
	for _, r := range rep.Results {
		if r.Check.ID[:min(len(r.Check.ID), 20)] == "caja-contract-schema" {
			cajaChecks++
			if r.Veredicto != domain.VeredictoPass {
				t.Errorf("caja %s did not pass: %s", r.Check.ID, r.Detalle)
			}
		}
	}
	if cajaChecks != 4 {
		t.Errorf("expected 4 caja-contract checks, got %d", cajaChecks)
	}
	// The firewall must actually run (spec-writer has a real fuente_path) and pass.
	for _, r := range rep.Results {
		if r.Check.ID == "no-phantom-frontmatter" && r.Veredicto != domain.VeredictoPass {
			t.Errorf("firewall should run and pass, got %s — %s", r.Veredicto, r.Detalle)
		}
	}
}

// TestBrokenArnesFailsRightChecks is the adversarial twin: a malformed arnés must fail the
// specific checks — a missing why (schema), an illegal transition and an out-of-spine
// state (spine), a phantom frontmatter key (firewall) — never a silent pass.
func TestBrokenArnesFailsRightChecks(t *testing.T) {
	svc, _ := newService(t)
	dir := t.TempDir()

	// A node source file with a phantom frontmatter key the firewall must catch.
	badSkill := filepath.Join(dir, "bad.SKILL.md")
	if err := os.WriteFile(badSkill, []byte("---\nname: bad\nsanctum: PERSONA\n---\nx\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	broken := `{
      "arnes": {
        "id": "roto", "rol": "x", "proceso": "y", "empresa": "z", "reporta_a": null,
        "fases": ["spec"],
        "spine": { "inicial": "idea", "terminales": ["done"], "estados": ["idea","spec","done"],
          "transiciones": [ {"de":"idea","a":"spec"} ] }
      },
      "nodos": [
        { "id": "sin-why", "clase": "skill", "nombre": "caja sin why", "banda": "fase", "fase": "spec",
          "estado": "idea -> spec", "fuente_path": "` + badSkill + `",
          "contract": { "clase":"skill", "arquetipo":"excepcion", "perfil_harness":"T1",
            "caja": true, "fase": "spec", "estado": "idea -> spec", "gate": { "tipo": "none" } } },
        { "id": "fuera-de-spine", "clase": "skill", "nombre": "estado ilegal", "banda": "fase", "fase": "spec",
          "estado": "spec -> released",
          "contract": { "why":"x","clase":"skill","arquetipo":"pipeline","perfil_harness":"T1",
            "caja": true, "fase": "spec", "estado": "spec -> released", "gate": { "tipo": "none" } } }
      ]
    }`
	path := filepath.Join(dir, "roto.graph.json")
	if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := svc.Run(context.Background(), ports.Target{Kind: ports.TargetArnes, GraphPath: path})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if rep.OK() {
		t.Fatal("a broken arnés must NOT conform")
	}
	got := map[string]domain.Veredicto{}
	for _, r := range rep.Results {
		got[r.Check.ID] = r.Veredicto
	}
	// The caja without `why` must fail schema validation.
	if got["caja-contract-schema:sin-why"] != domain.VeredictoFail {
		t.Errorf("caja sin why: want fail, got %q", got["caja-contract-schema:sin-why"])
	}
	// 'released' is not in the declared spine → estado-en-spine must fail.
	if got["estado-en-spine-declarado"] != domain.VeredictoFail {
		t.Errorf("estado-en-spine: want fail, got %q", got["estado-en-spine-declarado"])
	}
	// spec->released is not a legal transition.
	if got["transicion-legal"] != domain.VeredictoFail {
		t.Errorf("transicion-legal: want fail, got %q", got["transicion-legal"])
	}
	// The firewall must catch the phantom `sanctum:` key.
	if got["no-phantom-frontmatter"] != domain.VeredictoFail {
		t.Errorf("firewall: want fail on phantom key, got %q", got["no-phantom-frontmatter"])
	}
	// 'done' is declared but unreachable from 'idea' → spine-cobertura must fail.
	if got["spine-cobertura"] != domain.VeredictoFail {
		t.Errorf("spine-cobertura: want fail (done unreachable), got %q", got["spine-cobertura"])
	}
}
