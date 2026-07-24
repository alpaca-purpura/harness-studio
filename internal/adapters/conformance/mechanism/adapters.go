package mechanism

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// pass/deferred/error helpers keep the adapters terse.
func result(c domain.Check, v domain.Veredicto, detalle string) domain.CheckResult {
	return domain.CheckResult{Check: c, Veredicto: v, Detalle: detalle}
}

// ── arch-test ───────────────────────────────────────────────────────────────

// ArchTest runs a named Go test as a subprocess and maps its outcome. The enforced_by
// decide el paquete: el alias histórico "arch_test.go:TestX" corre en el paquete fitness;
// una ruta repo-relativa "dir/foo_test.go:TestX" corre en el paquete de ESE archivo
// (tests colocados junto al código que guardan — patrón del boundary del Portafolio).
// A dangling enforced_by (a test or file that does not exist) yields VeredictoError
// — "no such test" — never a silent pass (this is how B4's phantom enforcers surface).
type ArchTest struct {
	repoRoot string
	pkg      string // e.g. "./docs/architecture/fitness/"
}

// NewArchTest returns an arch-test adapter rooted at repoRoot.
func NewArchTest(repoRoot string) *ArchTest {
	return &ArchTest{repoRoot: repoRoot, pkg: "./docs/architecture/fitness/"}
}

// Mecanismo reports the mechanism this adapter executes (arch-test).
func (*ArchTest) Mecanismo() domain.Mecanismo { return domain.MecArchTest }

// Run executes the check's named fitness test via `go test -run` and maps the outcome:
// pass / fail (first FAIL line) / deferred (t.Skip stub) / error (dangling enforced_by).
func (a *ArchTest) Run(ctx context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	if a.repoRoot == "" {
		// Scope `fabrica` (decisión HS-10): los arch-tests auditan el código FUENTE de
		// ArnesIA y exigen repo + toolchain Go — en un binario instalado no hay ninguno
		// de los dos. Fuera del repo difieren honesto; el scope `arnes` (schema, spine,
		// firewall) es el que viaja al cliente.
		return result(c, domain.VeredictoDiferido,
			"scope fabrica: requiere el repo fuente + toolchain Go (no viaja en el binario instalado)")
	}
	test := testNameOf(c.EnforcedBy)
	if test == "" {
		return result(c, domain.VeredictoDiferido,
			"enforcer genérico `arch_test.go` sin función nombrada — sin test específico que correr")
	}
	pkg, ok := pkgOf(a.repoRoot, c.EnforcedBy, a.pkg)
	if !ok {
		return result(c, domain.VeredictoError,
			"enforced_by cuelga: el archivo de `"+c.EnforcedBy+"` no existe en el árbol")
	}
	// The test name is parsed from the repo's own ruleset (.md as data) and anchored in
	// the regex; the binary is the local `go` toolchain — local-first, never remote input.
	cmd := exec.CommandContext(ctx, "go", "test", "-run", "^"+test+"$", "-count=1", "-v", pkg) //nolint:gosec // G204: test comes from the repo's own docs/architecture/knowledge/+arch/ ruleset, not external input.
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

// pkgOf resolves the Go package dir an enforced_by runs in. Sin ruta con "/" (el alias
// histórico "arch_test.go:TestX") mantiene el paquete fitness por defecto; con ruta
// repo-relativa ("dir/foo_test.go:TestX") corre en el paquete del archivo referido.
// ok=false cuando el archivo referido no existe — enforcer colgante, jamás pass.
func pkgOf(repoRoot, enforcedBy, fallback string) (string, bool) {
	i := strings.LastIndex(enforcedBy, ":")
	if i < 0 {
		return fallback, true
	}
	file := strings.TrimSpace(enforcedBy[:i])
	if !strings.Contains(file, "/") {
		return fallback, true
	}
	if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(file))); err != nil {
		return "", false
	}
	return "./" + path.Dir(file) + "/", true
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

// GoArchLint runs `go run github.com/fe3dback/go-arch-lint@latest check` — the EXACT
// invocation `ci.yml` uses (job `go`, step "go-arch-lint") — via the local `go` toolchain,
// never a preinstalled `go-arch-lint` binary. Bug fixed 2026-07-24 (HS-27, barrido de deuda
// viva): the adapter used to probe `exec.LookPath("go-arch-lint")` first, which is never on
// PATH in dev/CI (both invoke it through `go run`) — so this check deferred locally even
// though CI runs+enforces it for real, a false "no enforcer" reading of a mechanism that
// was actually already live.
type GoArchLint struct {
	repoRoot string
	once     bool
	cached   domain.Veredicto
	detalle  string
}

// NewGoArchLint returns a go-arch-lint adapter rooted at repoRoot.
func NewGoArchLint(repoRoot string) *GoArchLint { return &GoArchLint{repoRoot: repoRoot} }

// Mecanismo reports the mechanism this adapter executes (go-arch-lint).
func (*GoArchLint) Mecanismo() domain.Mecanismo { return domain.MecGoArchLint }

// Run invokes go-arch-lint once (via `go run`, network/module-cache required — same
// requirement `ci.yml` has), caches the verdict, and reuses it for every check routed here.
func (g *GoArchLint) Run(ctx context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	if g.repoRoot == "" {
		// Scope `fabrica` (HS-10): igual que arch-test, el grafo de imports solo existe
		// donde está el código fuente — fuera del repo difiere honesto.
		return result(c, domain.VeredictoDiferido,
			"scope fabrica: requiere el repo fuente (no viaja en el binario instalado)")
	}
	if !g.once {
		g.once = true
		cmd := exec.CommandContext(ctx, "go", "run", "github.com/fe3dback/go-arch-lint@latest",
			"check", "--project-path", ".", "--arch-file", "docs/architecture/fitness/.go-arch-lint.yml", "--output-color=false")
		cmd.Dir = g.repoRoot
		out, err := cmd.CombinedOutput()
		if err == nil {
			g.cached, g.detalle = domain.VeredictoPass, "go-arch-lint check verde"
		} else {
			g.cached, g.detalle = domain.VeredictoFail, firstArchLintNotice(string(out))
		}
	}
	return result(c, g.cached, g.detalle)
}

// firstArchLintNotice returns the first violation line of go-arch-lint's plain-text
// output ("Component X shouldn't depend on Y" / "File Z not attached to any component"),
// falling back to the trimmed full output if the shape doesn't match (format change).
func firstArchLintNotice(out string) string {
	for _, line := range strings.Split(out, "\n") {
		l := strings.TrimSpace(line)
		if strings.Contains(l, "shouldn't depend") || strings.Contains(l, "not attached to any component") {
			return l
		}
	}
	return strings.TrimSpace(out)
}

// ── nl-judge ────────────────────────────────────────────────────────────────

// NLJudge always defers: a judgement/telemetry check has no deterministic enforcer here.
// Honest by construction — a deferred verdict is an open gap, never a fabricated pass.
type NLJudge struct{}

// Mecanismo reports the mechanism this adapter executes (nl-judge).
func (NLJudge) Mecanismo() domain.Mecanismo { return domain.MecNLJudge }

// Run always returns a deferred verdict — see the type comment.
func (NLJudge) Run(_ context.Context, c domain.Check, _ ports.Target) domain.CheckResult {
	return result(c, domain.VeredictoDiferido, "check de juicio/telemetría — sin enforcer determinista (nl-judge)")
}

// ── static-scan ─────────────────────────────────────────────────────────────

// StaticScan defers in the element path (no concrete target to scan). The arnés path
// runs the real source scans (firewall, single-writer) as built-ins in the use case.
type StaticScan struct{}

// Mecanismo reports the mechanism this adapter executes (static-scan).
func (StaticScan) Mecanismo() domain.Mecanismo { return domain.MecStaticScan }

// Run defers for both target kinds — see the type comment for where the real scans live.
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

// Mecanismo reports the mechanism this adapter executes (schema-validation).
func (SchemaAdapter) Mecanismo() domain.Mecanismo { return domain.MecSchema }

// Run defers for both target kinds — see the type comment for where validation runs.
func (SchemaAdapter) Run(_ context.Context, c domain.Check, t ports.Target) domain.CheckResult {
	if t.Kind != ports.TargetArnes {
		return result(c, domain.VeredictoDiferido, "schema-validation sin instancia — se corre validando un arnés concreto")
	}
	return result(c, domain.VeredictoDiferido, "validación de schema de este check se cubre en la ruta del arnés")
}
