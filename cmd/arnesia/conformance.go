package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// runConformance implements `arnesia conformance <target>` — the ruleset runner
// (METODOLOGIA §6, the pillar P0). It loads knowledge/+arch/ as a ruleset of data and
// runs the checks relevant to the target, emitting real per-check verdicts (never
// markdown). Exit code is non-zero when a blocking (error-severity) check fails.
func runConformance(args []string) error {
	fs := flag.NewFlagSet("conformance", flag.ExitOnError)
	arnes := fs.String("arnes", "", "path to an arnés graph.l0 JSON to validate (contracts + spine + firewall)")
	todo := fs.Bool("todo", false, "run the whole ruleset (all elements)")
	root := fs.String("root", "", "repo root (default: walk up from cwd to go.mod)")
	asJSON := fs.Bool("json", false, "emit the report as JSON")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia conformance [<elemento> | --arnes <path> | --todo] [flags]

  <elemento>       run one knowledge/arch node's checks (e.g. skills, hooks, superficie-local-confinada)
  --arnes <path>   validate an arnés graph.l0 JSON (box contracts + spine consistency + firewall)
  --todo           run the entire ruleset
  --root <dir>     repo root (default: walk up from cwd)
  --json           machine-readable report
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	repoRoot := *root
	if repoRoot == "" {
		r, err := findRepoRoot()
		if err != nil {
			return err
		}
		repoRoot = r
	}

	rs := ruleset.New(repoRoot)
	adapters := []ports.MechanismAdapter{
		mechanism.NewArchTest(repoRoot),
		mechanism.NewGoArchLint(repoRoot),
		mechanism.NLJudge{},
		mechanism.StaticScan{},
		mechanism.SchemaAdapter{},
	}
	svc := usecase.NewConformanceService(repoRoot, rs, adapters)

	target, err := resolveTarget(fs.Arg(0), *arnes, *todo)
	if err != nil {
		fs.Usage()
		return err
	}

	report, err := svc.Run(context.Background(), target)
	if err != nil {
		return err
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		printReport(report)
	}
	if !report.OK() {
		os.Exit(1)
	}
	return nil
}

// resolveTarget builds the conformance Target from the CLI args.
func resolveTarget(elemento, arnes string, todo bool) (ports.Target, error) {
	switch {
	case arnes != "":
		return ports.Target{Kind: ports.TargetArnes, GraphPath: arnes}, nil
	case todo:
		return ports.Target{Kind: ports.TargetTodo}, nil
	case elemento != "":
		return ports.Target{Kind: ports.TargetElemento, Nombre: elemento}, nil
	default:
		return ports.Target{}, fmt.Errorf("falta el target: un <elemento>, --arnes <path> o --todo")
	}
}

// findRepoRoot walks up from cwd to the directory holding go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no se encontró go.mod (¿estás dentro del repo?)")
		}
		dir = parent
	}
}

// printReport prints a human-readable conformance report grouped by verdict.
func printReport(r domain.ConformanceReport) {
	tally := r.Tally()
	fmt.Printf("conformance %s\n", r.Target)
	fmt.Printf("  %d checks · pass %d · fail %d · error %d · deferred %d · n/a %d\n\n",
		len(r.Results), tally[domain.VeredictoPass], tally[domain.VeredictoFail],
		tally[domain.VeredictoError], tally[domain.VeredictoDiferido], tally[domain.VeredictoNoAplica])

	// Order: failures and errors first (most actionable), then pass, then deferred.
	order := map[domain.Veredicto]int{
		domain.VeredictoError: 0, domain.VeredictoFail: 1, domain.VeredictoPass: 2,
		domain.VeredictoDiferido: 3, domain.VeredictoNoAplica: 4,
	}
	sorted := make([]domain.CheckResult, len(r.Results))
	copy(sorted, r.Results)
	sort.SliceStable(sorted, func(i, j int) bool {
		return order[sorted[i].Veredicto] < order[sorted[j].Veredicto]
	})

	for _, res := range sorted {
		fmt.Printf("  %-9s %-8s %-42s %s\n", glyph(res.Veredicto), res.Check.Severidad, res.Check.ID, res.Detalle)
	}
}

func glyph(v domain.Veredicto) string {
	switch v {
	case domain.VeredictoPass:
		return "PASS"
	case domain.VeredictoFail:
		return "FAIL"
	case domain.VeredictoError:
		return "ERROR"
	case domain.VeredictoDiferido:
		return "defer"
	default:
		return string(v)
	}
}
