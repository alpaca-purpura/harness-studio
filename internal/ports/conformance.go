package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// This file declares the ports of the conformance engine (HS-08, pillar P0). The core
// use case orchestrates; it never knows HOW a check runs or WHERE the ruleset comes
// from. RulesetPort + MechanismAdapter are outbound (driven); ConformancePort is the
// inbound (driving) port the CLI/HTTP call.

// TargetKind distinguishes what a conformance run points at.
type TargetKind string

const (
	// TargetElemento — run one docs/architecture/knowledge/arch element node's checks (e.g. "skills").
	TargetElemento TargetKind = "elemento"
	// TargetArnes — validate a concrete arnés graph (its box contracts + spine consistency).
	TargetArnes TargetKind = "arnes"
	// TargetTodo — run the whole ruleset.
	TargetTodo TargetKind = "todo"
)

// Target is what `arnesia conformance <target>` points at.
type Target struct {
	Kind TargetKind
	// Nombre is the element name when Kind=elemento.
	Nombre string
	// GraphPath is the path to a graph.l0 JSON (an arnés manifiesto) when Kind=arnes.
	GraphPath string
}

// RulesetPort loads docs/architecture/knowledge/ + arch/ as a parsed ruleset of DATA (principle 2). Its
// adapter is a markdown frontmatter+table parser; changing the standard = changing the
// .md files this port re-reads, never the engine.
type RulesetPort interface {
	Load(ctx context.Context) (domain.Ruleset, error)
}

// SchemaValidator validates an instance against a named JSON schema of the contract set
// (graph.l0 + box.contract). Outbound port: the use case asks "does this validate?" and
// never knows the schema engine nor where the .json files live — the composition root
// wires the concrete set (hoy disco del repo; mañana go:embed portable, misma interfaz).
type SchemaValidator interface {
	// Validate checks instance against the schema file (e.g. "graph.l0.schema.json").
	Validate(schemaFile string, instance any) error
	// ValidateJSON round-trips v through JSON before validating (for typed structs).
	ValidateJSON(schemaFile string, v any) error
}

// MechanismAdapter runs a single check against a target and returns its verdict. One
// adapter per domain.Mecanismo; the runner routes by check.Mecanismo. A mechanism that
// cannot execute here returns VeredictoDiferido (honest), never a fabricated pass.
type MechanismAdapter interface {
	// Mecanismo reports which mechanism this adapter handles.
	Mecanismo() domain.Mecanismo
	// Run evaluates one check against the target.
	Run(ctx context.Context, check domain.Check, target Target) domain.CheckResult
}

// ConformancePort is the inbound port: "run the ruleset relevant to target → report".
// The CLI and the HTTP handler depend on this, not on the concrete engine.
type ConformancePort interface {
	Run(ctx context.Context, target Target) (domain.ConformanceReport, error)
	// RunGraph validates an in-memory arnés graph (raw graph.l0 JSON) — la vía del
	// endpoint del daemon: el Mapa audita lo que el índice ya sirve, sin pasar por
	// disco. baseDir resuelve los fuente_path relativos del grafo para el firewall
	// scan; vacío = firewall diferido honesto si ningún path resuelve.
	RunGraph(ctx context.Context, raw []byte, baseDir string) (domain.ConformanceReport, error)
}
