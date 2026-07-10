// Package ruleset implements ports.RulesetPort: it parses the docs/architecture/knowledge/ and arch/
// trees into a domain.Ruleset of DATA (principle 2 — the knowledge is data, not code).
// Each `Checklist evaluable` table row becomes a domain.Check; a check's mechanism and
// enforced_by are either declared explicitly in the node's `conformance:` frontmatter
// map (the wired subset) or inferred from the table's enforcer column (arch) / defaulted
// to nl-judge (knowledge). Changing the standard = editing those .md files; the engine
// never recompiles.
package ruleset

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"gopkg.in/yaml.v3"
)

// Loader parses the docs/architecture/knowledge/ + arch/ trees of an fs.FS (el disco del repo en dev; el
// embed del binario en una instalación de cliente — misma lógica, cero divergencia:
// boundary doctrina-una-fuente-dos-targets de la investigación de inyección, firmado HS-10).
type Loader struct {
	fsys fs.FS
	dirs []string
}

var _ ports.RulesetPort = (*Loader)(nil)

// New returns a Loader over the on-disk repo tree rooted at repoRoot.
func New(repoRoot string) *Loader { return NewFromFS(os.DirFS(repoRoot)) }

// NewFromFS returns a Loader over any fs.FS carrying the three check-bearing trees:
// docs/architecture/knowledge/elements, docs/architecture/boundaries and docs/architecture/conventions (rutas slash, raíz del FS).
func NewFromFS(fsys fs.FS) *Loader {
	return &Loader{
		fsys: fsys,
		dirs: []string{
			"docs/architecture/knowledge/elements",
			"docs/architecture/boundaries",
			"docs/architecture/conventions",
		},
	}
}

// Load walks the trees and returns the union ruleset. A missing tree is not an error
// (the engine runs against whatever exists).
func (l *Loader) Load(_ context.Context) (domain.Ruleset, error) {
	var rs domain.Ruleset
	for _, d := range l.dirs {
		if _, err := fs.Stat(l.fsys, d); err != nil {
			continue
		}
		err := fs.WalkDir(l.fsys, d, func(p string, de fs.DirEntry, err error) error {
			if err != nil {
				return err // an unreadable entry must surface, not silently shrink the ruleset.
			}
			if de.IsDir() || !strings.HasSuffix(p, ".md") {
				return nil
			}
			name := strings.TrimSuffix(path.Base(p), ".md")
			if strings.HasPrefix(name, "INDEX") || strings.HasPrefix(name, "CADENCE") {
				return nil
			}
			checks, perr := parseFile(l.fsys, p)
			if perr != nil {
				return fmt.Errorf("%s: %w", p, perr)
			}
			rs.Checks = append(rs.Checks, checks...)
			return nil
		})
		if err != nil {
			return rs, err
		}
	}
	return rs, nil
}

// frontmatter is the subset of a node's YAML frontmatter the ruleset needs. The element
// name itself comes from the filename (robust across elemento/regla/convencion keys).
type frontmatter struct {
	Conformance map[string]struct {
		Mecanismo  string `yaml:"mecanismo"`
		EnforcedBy string `yaml:"enforced_by"`
	} `yaml:"conformance"`
}

var (
	sepFrontmatter = regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	reArchTest     = regexp.MustCompile(`arch_test\.go:(Test\w+)`)
	reTestName     = regexp.MustCompile(`\b(Test\w+)\b`)
)

// parseFile reads one node .md and returns its checks. elemento = the file basename.
func parseFile(fsys fs.FS, p string) ([]domain.Check, error) {
	raw, err := fs.ReadFile(fsys, p)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	elemento := strings.TrimSuffix(path.Base(p), ".md")

	var fm frontmatter
	if m := sepFrontmatter.FindStringSubmatch(text); m != nil {
		// Best-effort: a frontmatter that does not carry a `conformance:` map simply
		// yields an empty override set, never a hard error.
		_ = yaml.Unmarshal([]byte(m[1]), &fm)
	}

	rows := checklistRows(text)
	checks := make([]domain.Check, 0, len(rows))
	for _, cells := range rows {
		if len(cells) < 5 {
			continue
		}
		id := cleanCell(cells[0])
		if id == "" || id == "id" {
			continue // header or blank
		}
		enforcerCol := cleanCell(cells[4])
		mec, enf := inferMechanism(enforcerCol)
		if ov, ok := fm.Conformance[id]; ok { // explicit wiring wins.
			if ov.Mecanismo != "" {
				mec = domain.Mecanismo(ov.Mecanismo)
			}
			if ov.EnforcedBy != "" {
				enf = ov.EnforcedBy
			}
		}
		checks = append(checks, domain.Check{
			ID:         id,
			Elemento:   elemento,
			Mecanismo:  mec,
			EnforcedBy: enf,
			Severidad:  parseSeveridad(cleanCell(cells[2])),
			Senal:      cleanCell(cells[3]),
			Que:        cleanCell(cells[1]),
		})
	}
	return checks, nil
}

// checklistRows returns the data rows of the first `| id |`-headed table, as split cells.
func checklistRows(text string) [][]string {
	var rows [][]string
	inTable := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			if inTable {
				break // table ended.
			}
			continue
		}
		if strings.HasPrefix(trimmed, "| id |") || strings.HasPrefix(trimmed, "|id|") {
			inTable = true
			continue // skip header.
		}
		if !inTable {
			continue
		}
		if strings.Contains(trimmed, "---") && strings.Trim(trimmed, "|-: ") == "" {
			continue // separator row.
		}
		rows = append(rows, splitRow(trimmed))
	}
	return rows
}

// splitRow splits a markdown table row into its cells (drops the leading/trailing pipe).
func splitRow(row string) []string {
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	parts := strings.Split(row, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// cleanCell strips markdown emphasis/backticks noise for id/severity matching.
func cleanCell(s string) string {
	return strings.TrimSpace(s)
}

// parseSeveridad maps a cell to a Severidad, defaulting to warn for unknown text.
func parseSeveridad(s string) domain.Severidad {
	switch strings.ToLower(strings.Trim(s, "`* ")) {
	case "error":
		return domain.SevError
	case "warn", "warning":
		return domain.SevWarn
	case "info":
		return domain.SevInfo
	default:
		return domain.SevWarn
	}
}

// inferMechanism derives a mechanism + enforced_by from a table's enforcer/deriva cell.
// Knowledge "deriva de" cells (L1.x/L2.x provenance) and external-linter enforcers we do
// not execute fall back to nl-judge — reported deferred, never fabricated as pass.
func inferMechanism(enforcer string) (domain.Mecanismo, string) {
	e := strings.Trim(enforcer, "`* ")
	switch {
	case reArchTest.MatchString(e):
		m := reArchTest.FindStringSubmatch(e)
		return domain.MecArchTest, "arch_test.go:" + m[1]
	case strings.Contains(e, "arch_test.go"):
		if m := reTestName.FindStringSubmatch(e); m != nil {
			return domain.MecArchTest, "arch_test.go:" + m[1]
		}
		return domain.MecArchTest, "arch_test.go"
	case strings.Contains(e, "go-arch-lint"):
		return domain.MecGoArchLint, "go-arch-lint"
	case strings.Contains(e, ".schema.json"), strings.Contains(e, "box.contract"), strings.Contains(e, "graph.l0"):
		return domain.MecSchema, e
	default:
		// Provenance refs (L1.x/L2.x) and external linters (biome, golangci-lint, tsc,
		// dependency-cruiser, stylelint, steiger, lefthook) we don't run here.
		return domain.MecNLJudge, ""
	}
}
