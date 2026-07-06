package mechanism

import (
	"context"
	"os/exec"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// pass/deferred/error helpers keep the adapters terse.
func result(c domain.Check, v domain.Veredicto, detalle string) domain.CheckResult {
	return domain.CheckResult{Check: c, Veredicto: v, Detalle: detalle}
}

// ── arch-test ───────────────────────────────────────────────────────────────
// ArchTest runs a named Go fitness test (arch_test.go:TestX) as a subprocess and maps
// its outcome. A dangling enforced_by (a test that does not exist) yields VeredictoError
// — "no such test" — never a silent pass (this is how B4's phantom enforcers surface).
type ArchTest struct {
	repoRoot string
	pkg      string // e.g. "./arch/fitness/"
}

// NewArchTest returns an arch-test adapter rooted at repoRoot.
func NewArchTest(repoRoot string) *ArchTest {
	return &ArchTest{repoRoot: repoRoot, pkg: "./arch/fitness/"}
}

func (*ArchTest) Mecanismo() domain.Mecanismo { return domain.MecArchTest }

func (a *ArchTest) Run(ctx context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	test := testNameOf(c.EnforcedBy)
	if test == "" {
		return result(c, domain.VeredictoDiferido,
			"enforcer genérico `arch_test.go` sin función nombrada — sin test específico que correr")
	}
	cmd := exec.CommandContext(ctx, "go", "test", "-run", "^"+test+"$", "-count=1", "-v", a.pkg)
	cmd.Dir = a.repoRoot
	out, err := cmd.CombinedOutput()
	s := string(out)
	switch {
	case strings.Contains(s, "no tests to run"), strings.Contains(s, "no test files"),
		strings.Contains(s, "cannot find"), strings.Contains(s, "unknown"):
		return result(c, domain.VeredictoError,
			"enforced_by cuelga: `"+c.EnforcedBy+"` no existe (no tests to run)")
	case strings.Contains(s, "--- SKIP: "+test):
		// A t.Skip stub: the enforcer exists (no dangling pointer) but does not enforce
		// yet — deferred honestly, never a fabricated pass.
		return result(c, domain.VeredictoDiferido, test+" es stub t.Skip — enforcer aún no activo")
	case err == nil:
		return result(c, domain.VeredictoPass, test+" pasa")
	default:
		return result(c, domain.VeredictoFail, firstFail(s))
	}
}

// testNameOf extracts the Go test name from an enforced_by like "arch_test.go:TestX".
func testNameOf(enforcedBy string) string {
	i := strings.LastIndex(enforcedBy, ":")
	if i < 0 {
		return ""
	}
	name := strings.TrimSpace(enforcedBy[i+1:])
	if strings.HasPrefix(name, "Test") {
		return name
	}
	return ""
}

// firstFail returns the first FAIL/error line of `go test` output for the detail.
func firstFail(out string) string {
	for _, line := range strings.Split(out, "\n") {
		l := strings.TrimSpace(line)
		if strings.Contains(l, "FAIL") || strings.Contains(l, ".go:") {
			return l
		}
	}
	return "test falló"
}

// ── go-arch-lint ────────────────────────────────────────────────────────────
// GoArchLint runs go-arch-lint if the binary is present, else defers honestly.
type GoArchLint struct {
	repoRoot string
	once     bool
	cached   domain.Veredicto
	detalle  string
}

// NewGoArchLint returns a go-arch-lint adapter rooted at repoRoot.
func NewGoArchLint(repoRoot string) *GoArchLint { return &GoArchLint{repoRoot: repoRoot} }

func (*GoArchLint) Mecanismo() domain.Mecanismo { return domain.MecGoArchLint }

func (g *GoArchLint) Run(ctx context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	if _, err := exec.LookPath("go-arch-lint"); err != nil {
		return result(c, domain.VeredictoDiferido, "go-arch-lint no instalado — enforcer externo no corrido aquí")
	}
	if !g.once {
		g.once = true
		cmd := exec.CommandContext(ctx, "go-arch-lint", "check")
		cmd.Dir = g.repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			g.cached, g.detalle = domain.VeredictoFail, firstFail(string(out))
		} else {
			g.cached, g.detalle = domain.VeredictoPass, "go-arch-lint check verde"
		}
	}
	return result(c, g.cached, g.detalle)
}

// ── nl-judge ────────────────────────────────────────────────────────────────
// NLJudge always defers: a judgement/telemetry check has no deterministic enforcer here.
// Honest by construction — a deferred verdict is an open gap, never a fabricated pass.
type NLJudge struct{}

func (NLJudge) Mecanismo() domain.Mecanismo { return domain.MecNLJudge }

func (NLJudge) Run(_ context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	return result(c, domain.VeredictoDiferido, "check de juicio/telemetría — sin enforcer determinista (nl-judge)")
}

// ── static-scan ─────────────────────────────────────────────────────────────
// StaticScan defers in the element path (no concrete target to scan). The arnés path
// runs the real source scans (firewall, single-writer) as built-ins in the use case.
type StaticScan struct{}

func (StaticScan) Mecanismo() domain.Mecanismo { return domain.MecStaticScan }

func (StaticScan) Run(_ context.Context, c domain.Check, t ports.Target) domain.CheckResult {
	if t.Kind != ports.TargetArnes {
		return result(c, domain.VeredictoDiferido, "static-scan sin arnés objetivo — se corre en la ruta de conformidad del arnés")
	}
	return result(c, domain.VeredictoDiferido, "scan estático de este check aún no cableado — diferido honesto")
}

// ── schema-validation ───────────────────────────────────────────────────────
// SchemaAdapter defers in the element path (no instance). Box-contract / graph schema
// validation over a concrete arnés runs as a built-in in the use case, using SchemaSet.
type SchemaAdapter struct{}

func (SchemaAdapter) Mecanismo() domain.Mecanismo { return domain.MecSchema }

func (SchemaAdapter) Run(_ context.Context, c domain.Check, t ports.Target) domain.CheckResult {
	if t.Kind != ports.TargetArnes {
		return result(c, domain.VeredictoDiferido, "schema-validation sin instancia — se corre validando un arnés concreto")
	}
	return result(c, domain.VeredictoDiferido, "validación de schema de este check se cubre en la ruta del arnés")
}
