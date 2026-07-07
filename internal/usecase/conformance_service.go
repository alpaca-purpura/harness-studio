package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// ConformanceService is the conformance engine (ports.ConformancePort). It orchestrates:
// it loads the ruleset of DATA, routes each check to its mechanism adapter, and — for an
// arnés target — runs the built-in schema + spine + firewall checks that give the
// dogfood its real verdicts. The engine never hardcodes a check or a state value.
type ConformanceService struct {
	repoRoot string
	ruleset  ports.RulesetPort
	schemas  ports.SchemaValidator
	adapters map[domain.Mecanismo]ports.MechanismAdapter

	loaded *domain.Ruleset // cached ruleset.
}

var _ ports.ConformancePort = (*ConformanceService)(nil)

// NewConformanceService wires the engine: the ruleset port, the schema validator (el
// composition root cablea el SchemaSet concreto — el usecase no importa adapters,
// boundary dominio-independiente-de-transporte, cazado por go-arch-lint en HS-10) and
// the mechanism adapters.
func NewConformanceService(repoRoot string, rs ports.RulesetPort, schemas ports.SchemaValidator, adapters []ports.MechanismAdapter) *ConformanceService {
	m := map[domain.Mecanismo]ports.MechanismAdapter{}
	for _, a := range adapters {
		m[a.Mecanismo()] = a
	}
	return &ConformanceService{
		repoRoot: repoRoot,
		ruleset:  rs,
		schemas:  schemas,
		adapters: m,
	}
}

// Run executes the conformance run for a target and returns its report.
func (s *ConformanceService) Run(ctx context.Context, target ports.Target) (domain.ConformanceReport, error) {
	switch target.Kind {
	case ports.TargetArnes:
		return s.runArnes(ctx, target)
	case ports.TargetElemento:
		return s.runElemento(ctx, target, target.Nombre)
	case ports.TargetTodo:
		return s.runTodo(ctx, target)
	default:
		return domain.ConformanceReport{}, fmt.Errorf("target inválido: %q", target.Kind)
	}
}

// runElemento routes each of an element's checks to its mechanism adapter.
func (s *ConformanceService) runElemento(ctx context.Context, target ports.Target, elemento string) (domain.ConformanceReport, error) {
	rs, err := s.ensureRuleset(ctx)
	if err != nil {
		return domain.ConformanceReport{}, err
	}
	checks := rs.ForElemento(elemento)
	if len(checks) == 0 {
		return domain.ConformanceReport{}, fmt.Errorf("elemento desconocido: %q (sin checks en el ruleset)", elemento)
	}
	rep := domain.ConformanceReport{Target: "elemento:" + elemento}
	for _, c := range checks {
		rep.Results = append(rep.Results, s.route(ctx, c, target))
	}
	return rep, nil
}

// runTodo runs every element in the ruleset.
func (s *ConformanceService) runTodo(ctx context.Context, target ports.Target) (domain.ConformanceReport, error) {
	rs, err := s.ensureRuleset(ctx)
	if err != nil {
		return domain.ConformanceReport{}, err
	}
	rep := domain.ConformanceReport{Target: "todo"}
	for _, c := range rs.Checks {
		rep.Results = append(rep.Results, s.route(ctx, c, target))
	}
	return rep, nil
}

// route sends one check to its mechanism adapter; an unknown mechanism defers honestly.
func (s *ConformanceService) route(ctx context.Context, c domain.Check, target ports.Target) domain.CheckResult {
	a, ok := s.adapters[c.Mecanismo]
	if !ok {
		return domain.CheckResult{
			Check: c, Veredicto: domain.VeredictoDiferido,
			Detalle: "sin adaptador para el mecanismo " + string(c.Mecanismo),
		}
	}
	return a.Run(ctx, c, target)
}

func (s *ConformanceService) ensureRuleset(ctx context.Context) (domain.Ruleset, error) {
	if s.loaded != nil {
		return *s.loaded, nil
	}
	rs, err := s.ruleset.Load(ctx)
	if err != nil {
		return rs, err
	}
	s.loaded = &rs
	return rs, nil
}

// ── arnés conformance path (the dogfood's real verdicts, G1) ─────────────────

// runArnes validates a concrete arnés graph from disk (la vía CLI `--arnes`).
func (s *ConformanceService) runArnes(ctx context.Context, target ports.Target) (domain.ConformanceReport, error) {
	raw, err := os.ReadFile(target.GraphPath)
	if err != nil {
		return domain.ConformanceReport{}, fmt.Errorf("leer arnés %s: %w", target.GraphPath, err)
	}
	return s.RunGraph(ctx, raw, s.repoRoot)
}

// RunGraph validates an in-memory arnés graph: schema (manifiesto + each caja contract)
// + spine/fase consistency + the CC-native firewall. These built-ins give real
// pass/fail — el scope `arnes` (HS-10): lo que viaja al binario instalado, sin repo
// fuente ni toolchain. baseDir resuelve los fuente_path relativos para el firewall.
func (s *ConformanceService) RunGraph(_ context.Context, raw []byte, baseDir string) (domain.ConformanceReport, error) {
	var g domain.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		return domain.ConformanceReport{}, fmt.Errorf("decodificar arnés: %w", err)
	}
	rep := domain.ConformanceReport{Target: "arnes:" + arnesLabel(g, "")}

	// 1) Manifiesto + estructura del grafo contra graph.l0 (incluye META required y, vía
	//    $ref, cada nodo.contract contra box.contract).
	rep.Results = append(rep.Results, s.schemaCheck(
		"arnes-manifiesto-schema", domain.SevError,
		"el grafo del arnés valida contra graph.l0.schema.json (META completa + nodos)",
		"graph.l0.schema.json", anyOf(raw)))

	// 2) Cada caja: su contrato fusionado contra box.contract (granular por caja).
	for _, n := range g.Nodes {
		if !n.IsCaja() {
			continue
		}
		c := domain.Check{
			ID: "caja-contract-schema:" + n.ID, Elemento: "conformance",
			Mecanismo: domain.MecSchema, EnforcedBy: "box.contract.schema.json",
			Severidad: domain.SevError,
			Que:       "el contrato fusionado de la caja valida (intención+clasificación+cableado+aceptación)",
		}
		if err := s.schemas.ValidateJSON("box.contract.schema.json", n.Contract); err != nil {
			rep.Results = append(rep.Results, domain.CheckResult{Check: c, Veredicto: domain.VeredictoFail, Detalle: err.Error()})
		} else {
			rep.Results = append(rep.Results, domain.CheckResult{Check: c, Veredicto: domain.VeredictoPass, Detalle: "contrato válido"})
		}
	}

	// 3) Consistencia de estado/fase contra el spine DECLARADO del arnés (agnóstico).
	rep.Results = append(rep.Results, domain.VerificarSpine(g)...)

	// 4) Mutation contract: un solo escritor autorizado por artefacto.
	rep.Results = append(rep.Results, domain.VerificarEscritorUnico(g))

	// 5) Firewall CC-native: sin claves fantasma en los fuentes de los nodos.
	rep.Results = append(rep.Results, s.firewallScan(g))

	return rep, nil
}

// schemaCheck validates an instance against a schema file and returns a CheckResult.
func (s *ConformanceService) schemaCheck(id string, sev domain.Severidad, que, schemaFile string, instance any) domain.CheckResult {
	c := domain.Check{
		ID: id, Elemento: "conformance", Mecanismo: domain.MecSchema,
		EnforcedBy: schemaFile, Severidad: sev, Que: que,
	}
	if err := s.schemas.Validate(schemaFile, instance); err != nil {
		return domain.CheckResult{Check: c, Veredicto: domain.VeredictoFail, Detalle: err.Error()}
	}
	return domain.CheckResult{Check: c, Veredicto: domain.VeredictoPass, Detalle: "válido contra " + schemaFile}
}

// firewallScan enforces the CC-native firewall (§8.6): no phantom frontmatter keys in any
// node's source file. Keys CC ignores silently would do NOTHING — declaring them is a
// finding. Scans FuentePath files that exist; if none, defers honestly.
func (s *ConformanceService) firewallScan(g domain.Graph) domain.CheckResult {
	c := domain.Check{
		ID: "no-phantom-frontmatter", Elemento: "conformance",
		Mecanismo: domain.MecStaticScan, EnforcedBy: "usecase.firewallScan", Severidad: domain.SevError,
		Que: "sin claves de frontmatter que CC ignora (persistent_facts/activation_steps_prepend/customize/sanctum)",
	}
	forbidden := []string{"persistent_facts", "activation_steps_prepend", "customize", "sanctum"}
	var scanned int
	var findings []string
	for _, n := range g.Nodes {
		if n.FuentePath == "" {
			continue
		}
		p := n.FuentePath
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.repoRoot, p) // repo-relative fuente_path.
		}
		b, err := os.ReadFile(p) //nolint:gosec // G304: fuente_path is declared by the arnés graph under validation; reading it IS the firewall scan (local-first).
		if err != nil {
			continue
		}
		scanned++
		for _, k := range forbidden {
			if strings.Contains(string(b), k+":") {
				findings = append(findings, n.ID+": clave fantasma `"+k+"`")
			}
		}
	}
	switch {
	case scanned == 0:
		return domain.CheckResult{
			Check: c, Veredicto: domain.VeredictoDiferido,
			Detalle: "ningún nodo declara fuente_path escaneable — firewall diferido",
		}
	case len(findings) > 0:
		return domain.CheckResult{Check: c, Veredicto: domain.VeredictoFail, Detalle: strings.Join(findings, "; ")}
	default:
		return domain.CheckResult{
			Check: c, Veredicto: domain.VeredictoPass,
			Detalle: fmt.Sprintf("%d fuentes sin claves fantasma", scanned),
		}
	}
}

// anyOf decodes raw JSON into the generic shape the validator expects.
func anyOf(raw []byte) any {
	var v any
	_ = json.Unmarshal(raw, &v)
	return v
}

// arnesLabel returns the arnés id if present, else the file basename.
func arnesLabel(g domain.Graph, path string) string {
	if g.Arnes != nil && g.Arnes.ID != "" {
		return g.Arnes.ID
	}
	return filepath.Base(path)
}
