package domain

import (
	"sort"
	"strings"
)

// This file models the conformance layer of the doctrine (HS-08, the pillar P0): the
// common check contract shared by knowledge/ and arch/, the report a run produces, and
// the spine-consistency checks parametrized by an arnés's DECLARED spine (never a
// product enum — agnosticism, VISION p3/p7). The core owns the TYPES and the
// consistency logic; HOW each check runs lives behind a mechanism adapter (a port).

// Mecanismo is how a check is executed. Each check declares one; the conformance runner
// routes to the matching adapter. Adding a mechanism = a new adapter, core untouched.
type Mecanismo string

const (
	// MecSchema — validate a target JSON against a JSON Schema (jsonschema-go).
	MecSchema Mecanismo = "schema-validation"
	// MecGoArchLint — run go-arch-lint against the import graph.
	MecGoArchLint Mecanismo = "go-arch-lint"
	// MecArchTest — run a named Go fitness test (arch_test.go:TestX).
	MecArchTest Mecanismo = "arch-test"
	// MecStaticScan — a deterministic source/text scan (grep-like).
	MecStaticScan Mecanismo = "static-scan"
	// MecNLJudge — a judgement/telemetry check with no deterministic enforcer here; it is
	// reported `deferred` honestly, never fabricated as pass (VISION p6/p10).
	MecNLJudge Mecanismo = "nl-judge"
)

// Valid reports whether m is a known mechanism.
func (m Mecanismo) Valid() bool {
	switch m {
	case MecSchema, MecGoArchLint, MecArchTest, MecStaticScan, MecNLJudge:
		return true
	default:
		return false
	}
}

// Severidad is a check's weight: error breaks the standard, warn smells, info is a
// possible improvement (knowledge/ + arch/ convention).
type Severidad string

const (
	// SevError — the check breaks the standard; a fail/error here blocks the gate.
	SevError Severidad = "error"
	// SevWarn — a smell: reported in the verdict, never blocks the gate.
	SevWarn Severidad = "warn"
	// SevInfo — a possible improvement; informative only.
	SevInfo Severidad = "info"
)

// Check is the common check contract: one row of a `Checklist evaluable` (knowledge) or
// an arch boundary/convention, parsed as DATA (principle 2: knowledge is data, not
// code). Adding/changing a check = editing the .md, never the engine.
type Check struct {
	ID         string    `json:"id"`
	Elemento   string    `json:"elemento"`              // owning node: skills, hooks, orquestacion-…, go-style, or "conformance" (built-in).
	Mecanismo  Mecanismo `json:"mecanismo"`             // how it runs.
	EnforcedBy string    `json:"enforced_by,omitempty"` // concrete enforcer: "arch_test.go:TestX" | "box.contract.schema.json" | "go-arch-lint" | "".
	Severidad  Severidad `json:"severidad"`
	Senal      string    `json:"senal,omitempty"` // map signal painted on failure.
	Que        string    `json:"que,omitempty"`   // human "qué chequea".
}

// Veredicto is the outcome of running one check against a target.
type Veredicto string

const (
	// VeredictoPass — the check ran and the target satisfies it.
	VeredictoPass Veredicto = "pass"
	// VeredictoFail — the check ran and the target violates it.
	VeredictoFail Veredicto = "fail"
	// VeredictoDiferido — no deterministic mechanism ran here (nl-judge, or a tool
	// absent). HONEST: never counted as pass, never fabricated (VISION p6/p10).
	VeredictoDiferido Veredicto = "deferred"
	// VeredictoError — the mechanism itself failed (e.g. a dangling enforced_by naming a
	// test that does not exist → "no such test"). Distinct from Fail.
	VeredictoError Veredicto = "error"
	// VeredictoNoAplica — the check does not apply to this target (skipped cleanly).
	VeredictoNoAplica Veredicto = "n/a"
)

// CheckResult is one check's verdict against a target, with a human detail line.
type CheckResult struct {
	Check     Check     `json:"check"`
	Veredicto Veredicto `json:"veredicto"`
	Detalle   string    `json:"detalle,omitempty"`
}

// Ruleset is the parsed union of knowledge/ + arch/ checks — the ruleset of data the
// RulesetPort loads. The engine never hardcodes checks; it reads this.
type Ruleset struct {
	Checks []Check `json:"checks"`
}

// ForElemento returns the checks owned by a given element/boundary node.
func (rs Ruleset) ForElemento(elemento string) []Check {
	var out []Check
	for _, c := range rs.Checks {
		if c.Elemento == elemento {
			out = append(out, c)
		}
	}
	return out
}

// Elementos returns the distinct element names present in the ruleset, in first-seen order.
func (rs Ruleset) Elementos() []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range rs.Checks {
		if !seen[c.Elemento] {
			seen[c.Elemento] = true
			out = append(out, c.Elemento)
		}
	}
	return out
}

// ConformanceReport is what a conformance run emits: the verdicts per check for a target.
type ConformanceReport struct {
	Target  string        `json:"target"`
	Results []CheckResult `json:"results"`
}

// Tally counts results by verdict.
func (r ConformanceReport) Tally() map[Veredicto]int {
	t := map[Veredicto]int{}
	for _, res := range r.Results {
		t[res.Veredicto]++
	}
	return t
}

// OK reports whether the report has no blocking failures: no error-severity check
// Failed or Errored. Deferred/warn/info never block (honesty: a deferred check is an
// open gap, not a pass, but it does not fail the gate). Promotion gates are stricter.
func (r ConformanceReport) OK() bool {
	for _, res := range r.Results {
		if res.Check.Severidad != SevError {
			continue
		}
		if res.Veredicto == VeredictoFail || res.Veredicto == VeredictoError {
			return false
		}
	}
	return true
}

// ParseTransicion splits a box's `estado` ("X -> Y") into its endpoints. ok is false if
// the string is not a single well-formed transition. The spine values themselves are
// never a product constant — only the FORM is (agnosticism).
func ParseTransicion(estado string) (de, a string, ok bool) {
	parts := strings.Split(estado, "->")
	if len(parts) != 2 {
		return "", "", false
	}
	de = strings.TrimSpace(parts[0])
	a = strings.TrimSpace(parts[1])
	if de == "" || a == "" {
		return "", "", false
	}
	return de, a, true
}

// spineCheck builds a built-in conformance Check (elemento="conformance").
func spineCheck(id string, sev Severidad, que string) Check {
	return Check{
		ID:         id,
		Elemento:   "conformance",
		Mecanismo:  MecStaticScan,
		EnforcedBy: "domain.VerificarSpine",
		Severidad:  sev,
		Que:        que,
	}
}

// VerificarSpine runs the spine/fase consistency checks of a graph against the spine and
// fases its OWN arnés declared (METODOLOGIA §6 turned executable). It knows no concrete
// state — only whether what the boxes declare is consistent with what the arnés declared.
// Returns one CheckResult per built-in check. If the arnés declares no spine, the
// state-level checks report deferred (nothing to validate against — honest).
func VerificarSpine(g Graph) []CheckResult {
	var out []CheckResult
	var spine *Spine
	var fases []Fase
	if g.Arnes != nil {
		spine = g.Arnes.Spine
		fases = g.Arnes.Fases
	}

	// Collect the process boxes (cajas) and their declared transitions.
	type cajaTx struct {
		id     string
		estado string
		de, a  string
		ok     bool
	}
	var cajas []cajaTx
	for _, n := range g.Nodes {
		if !n.IsCaja() {
			continue
		}
		est := string(n.Estado)
		if est == "" && n.Contract != nil {
			est = n.Contract.Estado
		}
		de, a, ok := ParseTransicion(est)
		cajas = append(cajas, cajaTx{id: n.ID, estado: est, de: de, a: a, ok: ok})
	}

	// spine-auto-consistente: el spine declarado es coherente consigo mismo, antes de
	// cotejar cajas — inicial/terminales/extremos-de-transición ∈ estados.
	{
		c := spineCheck("spine-auto-consistente", SevError,
			"el spine declarado es coherente: inicial · terminales · extremos de transiciones ∈ estados")
		switch spine {
		case nil:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara spine",
			})
		default:
			var bad []string
			if spine.Inicial != "" && !spine.TieneEstado(spine.Inicial) {
				bad = append(bad, "inicial '"+spine.Inicial+"' ∉ estados")
			}
			for _, tstate := range spine.Terminales {
				if !spine.TieneEstado(tstate) {
					bad = append(bad, "terminal '"+tstate+"' ∉ estados")
				}
			}
			for _, tr := range spine.Transiciones {
				if !spine.TieneEstado(tr.De) {
					bad = append(bad, "transición: '"+tr.De+"' ∉ estados")
				}
				if !spine.TieneEstado(tr.A) {
					bad = append(bad, "transición: '"+tr.A+"' ∉ estados")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "spine internamente coherente"))
		}
	}

	// estado-en-spine-declarado
	{
		c := spineCheck("estado-en-spine-declarado", SevError,
			"cada extremo de la transición de una caja ∈ arnes.spine.estados")
		switch spine {
		case nil:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara spine — nada contra qué validar (agnóstico)",
			})
		default:
			var bad []string
			for _, cj := range cajas {
				if !cj.ok {
					bad = append(bad, cj.id+" (estado mal formado)")
					continue
				}
				if !spine.TieneEstado(cj.de) {
					bad = append(bad, cj.id+": '"+cj.de+"' ∉ spine")
				}
				if !spine.TieneEstado(cj.a) {
					bad = append(bad, cj.id+": '"+cj.a+"' ∉ spine")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "todo estado de caja ∈ spine declarado"))
		}
	}

	// transicion-legal
	{
		c := spineCheck("transicion-legal", SevError,
			"cada transición de caja X→Y ∈ arnes.spine.transiciones")
		switch {
		case spine == nil || len(spine.Transiciones) == 0:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara transiciones legales — no se puede validar legalidad",
			})
		default:
			var bad []string
			for _, cj := range cajas {
				if cj.ok && !spine.TransicionLegal(cj.de, cj.a) {
					bad = append(bad, cj.id+": '"+cj.de+" -> "+cj.a+"' no es transición legal")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "toda transición de caja es legal"))
		}
	}

	// una-transicion-por-caja (dos cajas no poseen la misma transición — mutation contract del spine)
	{
		c := spineCheck("una-transicion-por-caja", SevWarn,
			"dos cajas no declaran la misma transición de estado (un dueño por transición)")
		seen := map[string]string{}
		var bad []string
		for _, cj := range cajas {
			if !cj.ok {
				continue
			}
			key := cj.de + " -> " + cj.a
			if prev, dup := seen[key]; dup {
				bad = append(bad, "'"+key+"' la poseen "+prev+" y "+cj.id)
			} else {
				seen[key] = cj.id
			}
		}
		out = append(out, veredictoDeLista(c, bad, "cada transición tiene un solo dueño"))
	}

	// categoria-estado-existe (HS-12, interop DevStudio): el mapa OPCIONAL
	// spine.categorias clasifica estados per-arnés con las 5 categorías fijas (I-77
	// RN-28). Claves ∈ estados y valores ∈ enum; ausente → diferido honesto.
	{
		c := spineCheck("categoria-estado-existe", SevWarn,
			"cada clave de spine.categorias ∈ estados y cada valor es una de las 5 categorías fijas")
		switch {
		case spine == nil || len(spine.Categorias) == 0:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara categorías de spine — mapa opcional (HS-12)",
			})
		default:
			var bad []string
			for _, e := range spine.Estados {
				if cat, ok := spine.Categorias[e]; ok && !cat.Valid() {
					bad = append(bad, "'"+e+"': categoría '"+string(cat)+"' fuera del enum fijo")
				}
			}
			var huerfanas []string
			for e := range spine.Categorias {
				if !spine.TieneEstado(e) {
					huerfanas = append(huerfanas, "'"+e+"' ∉ estados")
				}
			}
			sort.Strings(huerfanas)
			bad = append(bad, huerfanas...)
			out = append(out, veredictoDeLista(c, bad, "categorías consistentes con el spine"))
		}
	}

	// terminal-categoria-coherente (HS-12): la terminalidad se DERIVA de la categoría
	// (I-77) — todo terminal declarado debe mapear a completado|descartado.
	{
		c := spineCheck("terminal-categoria-coherente", SevWarn,
			"todo estado terminal mapea a categoría completado|descartado (terminalidad derivada, I-77)")
		switch {
		case spine == nil || len(spine.Categorias) == 0:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara categorías de spine — mapa opcional (HS-12)",
			})
		default:
			var bad []string
			for _, t := range spine.Terminales {
				cat, ok := spine.Categorias[t]
				switch {
				case !ok:
					bad = append(bad, "terminal '"+t+"' sin categoría")
				case !cat.Terminal():
					bad = append(bad, "terminal '"+t+"' con categoría '"+string(cat)+"' (no terminal)")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "terminales coherentes con su categoría"))
		}
	}

	// spine-cobertura (todo estado alcanzable desde inicial; sin estados huérfanos/inalcanzables)
	{
		c := spineCheck("spine-cobertura", SevWarn,
			"todo estado del spine es alcanzable desde `inicial` (sin huérfanos/inalcanzables)")
		switch {
		case spine == nil:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara spine",
			})
		case spine.Inicial == "" || len(spine.Transiciones) == 0:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "spine sin `inicial` o sin transiciones — cobertura no computable",
			})
		default:
			reachable := reachableFrom(spine.Inicial, spine.Transiciones)
			var bad []string
			for _, e := range spine.Estados {
				if !reachable[e] {
					bad = append(bad, "'"+e+"' inalcanzable desde '"+spine.Inicial+"'")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "todo estado alcanzable desde inicial"))
		}
	}

	// fase-en-fases-declaradas (nodo.fase ∈ arnes.fases; cada fase cubierta por ≥1 caja)
	{
		c := spineCheck("fase-en-fases-declaradas", SevError,
			"cada nodo.fase ∈ arnes.fases; cada fase declarada cubierta por ≥1 caja")
		switch {
		case len(fases) == 0:
			out = append(out, CheckResult{
				Check: c, Veredicto: VeredictoDiferido,
				Detalle: "el arnés no declara fases — nada contra qué validar",
			})
		default:
			declared := map[string]bool{}
			for _, f := range fases {
				declared[string(f)] = false // false = aún no cubierta
			}
			var bad []string
			for _, n := range g.Nodes {
				f := string(n.Fase)
				if f == "" {
					continue
				}
				if _, ok := declared[f]; !ok {
					bad = append(bad, n.ID+": fase '"+f+"' ∉ arnes.fases")
				} else if n.IsCaja() {
					declared[f] = true
				}
			}
			for f, cubierta := range declared {
				if !cubierta {
					bad = append(bad, "fase '"+f+"' sin caja que la cubra")
				}
			}
			out = append(out, veredictoDeLista(c, bad, "fases consistentes y cubiertas"))
		}
	}

	return out
}

// VerificarEscritorUnico enforces the mutation contract (METODOLOGIA §3, checks huérfanos
// M12): an artifact has ONE authorized writer. Two cajas whose `entrega` declares the same
// `art` with escritor_unico (the default) are a conformance finding. The ONLY legal gate
// to multi-writing is `refina` (D9): an output declaring itself a revision leaves this
// count — its legality (linear chain, no cycles) is VerificarRefinaCoherente's job. Two
// entregas of the same art WITHOUT refina remain a finding, as always. Reads only the
// declared contracts — no product state.
func VerificarEscritorUnico(g Graph) CheckResult {
	c := Check{
		ID: "escritor-unico", Elemento: "conformance", Mecanismo: MecStaticScan,
		EnforcedBy: "domain.VerificarEscritorUnico", Severidad: SevError,
		Que: "un solo escritor autorizado por artefacto (dos cajas escribiendo el mismo art SIN refina = hallazgo)",
	}
	writers := map[string][]string{} // art → caja ids that claim single-writer (non-revision).
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, o := range n.Contract.Entrega {
			if o.Refina != "" {
				continue // revisión declarada: la valida refina-coherente, no este check.
			}
			// escritor_unico defaults true (schema default): nil counts as claiming it.
			if o.EscritorUnico == nil || *o.EscritorUnico {
				writers[o.Art] = append(writers[o.Art], n.ID)
			}
		}
	}
	var bad []string
	for art, ws := range writers {
		if len(ws) > 1 {
			bad = append(bad, "'"+art+"' lo escriben "+strings.Join(ws, ", "))
		}
	}
	return veredictoDeLista(c, bad, "cada artefacto tiene un solo escritor")
}

// ── composición del cableado (franja-artefactos D8/D9, RF-100..104) ─────────────
// The hand-off between cajas IS the contract (A2). These checks make the promised
// composition rules executable on the `--arnes` path: every declared input has its
// producer, every output has a consumer (or its caja is terminal), every route lands
// on a real box, and refina chains are linear. All read ONLY declared contracts.

// composicionCheck builds a built-in composition Check (elemento="conformance").
func composicionCheck(id string, enforcedBy string, sev Severidad, que string) Check {
	return Check{
		ID:         id,
		Elemento:   "conformance",
		Mecanismo:  MecStaticScan,
		EnforcedBy: enforcedBy,
		Severidad:  sev,
		Que:        que,
	}
}

// entregaIndex maps cajaID → set of arts the caja produces. An output with `refina`
// counts as producing the refined art (the revision IS the art, D9c) besides its own.
func entregaIndex(g Graph) map[string]map[string]bool {
	idx := map[string]map[string]bool{}
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, o := range n.Contract.Entrega {
			if idx[n.ID] == nil {
				idx[n.ID] = map[string]bool{}
			}
			if o.Art != "" {
				idx[n.ID][o.Art] = true
			}
			if o.Refina != "" {
				idx[n.ID][o.Refina] = true
			}
		}
	}
	return idx
}

// deCaja extracts the caja id of a `necesita.de` reference ("caja:X" → "X", ok). Any
// other origin (usuario, terceros:*, base:*, libreria:*, maquinaria:*, marcas-dormidas:*)
// is a component or an external input — never an orphan candidate (D10, C3-C6).
func deCaja(de string) (string, bool) {
	id, found := strings.CutPrefix(de, "caja:")
	if !found || id == "" {
		return "", false
	}
	return id, true
}

// VerificarSinHuerfanos (RF-100, warn): every `necesita.de: caja:X` must name a producer
// that actually delivers (or refines) that art. External origins (usuario/terceros:*) and
// Base components are NOT orphans — they are admitted, not produced (D10).
func VerificarSinHuerfanos(g Graph) CheckResult {
	c := composicionCheck("sin-huerfanos", "domain.VerificarSinHuerfanos", SevWarn,
		"ningún `necesita` referencia un artefacto que ninguna caja `entrega` upstream")
	idx := entregaIndex(g)
	var bad []string
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, in := range n.Contract.Necesita {
			prod, isCaja := deCaja(in.De)
			if !isCaja {
				continue
			}
			if !idx[prod][in.Art] {
				bad = append(bad, n.ID+": necesita '"+in.Art+"' de caja:"+prod+" que no la entrega")
			}
		}
	}
	return veredictoDeLista(c, bad, "todo input declarado tiene productor")
}

// VerificarDeadEnds (RF-101, warn): an output no caja consumes is a dead-end — UNLESS its
// caja is terminal (its transition lands on a terminal state of the declared spine):
// terminality is DERIVED, never declared (HS-12/C15). Without spine/terminales the check
// defers honestly (nothing to derive against).
func VerificarDeadEnds(g Graph) CheckResult {
	c := composicionCheck("dead-end", "domain.VerificarDeadEnds", SevWarn,
		"toda entrega tiene consumidor, salvo que su caja sea terminal (terminalidad derivada del spine)")
	if g.Arnes == nil || g.Arnes.Spine == nil || len(g.Arnes.Spine.Terminales) == 0 {
		return CheckResult{
			Check: c, Veredicto: VeredictoDiferido,
			Detalle: "el arnés no declara spine/terminales — terminalidad no derivable",
		}
	}
	terminal := map[string]bool{}
	for _, t := range g.Arnes.Spine.Terminales {
		terminal[t] = true
	}
	// consumo: (productor, art) consumido por algún necesita.
	consumido := map[string]bool{}
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, in := range n.Contract.Necesita {
			if prod, isCaja := deCaja(in.De); isCaja {
				consumido[prod+"\x00"+in.Art] = true
			}
		}
	}
	var bad []string
	for _, n := range g.Nodes {
		if !n.IsCaja() {
			continue
		}
		est := string(n.Estado)
		if est == "" {
			est = n.Contract.Estado
		}
		_, a, ok := ParseTransicion(est)
		if !ok {
			continue // estado mal formado: lo reporta estado-en-spine-declarado, no este check.
		}
		if terminal[a] {
			continue // salida del proceso (C15): sin consumidor ≠ dead-end.
		}
		for _, o := range n.Contract.Entrega {
			ok := consumido[n.ID+"\x00"+o.Art] ||
				(o.Refina != "" && consumido[n.ID+"\x00"+o.Refina])
			if !ok {
				bad = append(bad, n.ID+": entrega '"+o.Art+"' sin consumidor (caja no terminal)")
			}
		}
	}
	return veredictoDeLista(c, bad, "toda entrega de caja no terminal tiene consumidor")
}

// VerificarRutaExiste (RF-102, error): every `ruta[].a` must land on an existing caja or
// the literal `humano` (the P6 escalation target).
func VerificarRutaExiste(g Graph) CheckResult {
	c := composicionCheck("ruta-a-existe", "domain.VerificarRutaExiste", SevError,
		"cada `ruta[].a` apunta a una caja existente o al literal `humano`")
	cajas := map[string]bool{}
	for _, n := range g.Nodes {
		if n.IsCaja() {
			cajas[n.ID] = true
		}
	}
	var bad []string
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, r := range n.Contract.Ruta {
			if r.A != "humano" && !cajas[r.A] {
				bad = append(bad, n.ID+": ruta a '"+r.A+"' que no existe")
			}
		}
	}
	return veredictoDeLista(c, bad, "toda ruta aterriza en caja existente o humano")
}

// VerificarArtIdentidad (RF-103, error): `necesita {art:A, de:caja:X}` where X EXISTS
// demands that X delivers (or refines) exactly A — a name mismatch (entrega «spec.md» vs
// necesita «espec.md», C21) is never fused silently. A reference to a non-existent caja
// is the orphan case (sin-huerfanos), not an identity mismatch.
func VerificarArtIdentidad(g Graph) CheckResult {
	c := composicionCheck("art-identidad-coherente", "domain.VerificarArtIdentidad", SevError,
		"si el productor referido existe, entrega (o refina) exactamente el art que se le pide")
	idx := entregaIndex(g)
	nodos := map[string]bool{}
	for _, n := range g.Nodes {
		nodos[n.ID] = true
	}
	var bad []string
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, in := range n.Contract.Necesita {
			prod, isCaja := deCaja(in.De)
			if !isCaja || !nodos[prod] {
				continue
			}
			if !idx[prod][in.Art] {
				bad = append(bad, n.ID+": necesita '"+in.Art+"' de caja:"+prod+" pero no está entre sus entregas")
			}
		}
	}
	return veredictoDeLista(c, bad, "identidad de artefactos coherente entre necesita y entrega")
}

// VerificarArtEsPath (RF-112, warn): a pipeline/excepcion caja promises an exact,
// verifiable output (document-as-cache) — its primary entrega should declare `path`
// (artefacto-archivo, D3). A label-only entrega («código + tests») is honest but the
// conductor cannot stat it: the warn is the tooth, never silenced with invented paths.
// `abierto` is exempt (§8.1: restricts scope, not steps).
func VerificarArtEsPath(g Graph) CheckResult {
	c := composicionCheck("art-es-path", "domain.VerificarArtEsPath", SevWarn,
		"cajas pipeline/excepcion declaran `path` en su entrega primaria (artefacto-archivo, D3)")
	var bad []string
	for _, n := range g.Nodes {
		if !n.IsCaja() {
			continue
		}
		arq := n.Contract.Arquetipo
		if arq != ArqPipeline && arq != ArqExcepcion {
			continue
		}
		if len(n.Contract.Entrega) == 0 {
			continue // sin entrega no hay identidad que exigir (otro problema, otro check).
		}
		if n.Contract.Entrega[0].Path == "" {
			bad = append(bad, n.ID+": entrega '"+n.Contract.Entrega[0].Art+"' sin path (etiqueta)")
		}
	}
	return veredictoDeLista(c, bad, "toda caja pipeline/excepcion declara path en su entrega primaria")
}

// VerificarRefinaCoherente (RF-104, error): `refina` is the only legal gate to
// multi-writing (D9) and it must be a linear chain of revisions: (a) the refinador MUST
// need the very art it refines; (b) if that need comes from a caja, that caja must be a
// writer of the art; (c) no cycles; (d) no branching — one refinador per source writer.
// A chain may root on an external input (factura from terceros, C11) — that is legal.
func VerificarRefinaCoherente(g Graph) CheckResult {
	c := composicionCheck("refina-coherente", "domain.VerificarRefinaCoherente", SevError,
		"toda cadena de refina es lineal: el refinador necesita el art que refina, sin ciclos ni ramas")
	idx := entregaIndex(g)
	var bad []string

	// fuente(R, art) = de dónde toma R el art que refina (su necesita para ese art).
	fuente := func(n Box, art string) (string, bool) {
		for _, in := range n.Contract.Necesita {
			if in.Art == art {
				return in.De, true
			}
		}
		return "", false
	}

	// ramas: (art, productor-fuente) → refinadores que beben de él.
	ramas := map[string][]string{}

	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, o := range n.Contract.Entrega {
			if o.Refina == "" {
				continue
			}
			de, tiene := fuente(n, o.Refina)
			if !tiene {
				bad = append(bad, n.ID+": refina '"+o.Refina+"' sin necesitarla (D9a)")
				continue
			}
			if prod, isCaja := deCaja(de); isCaja {
				if !idx[prod][o.Refina] {
					bad = append(bad, n.ID+": refina '"+o.Refina+"' desde caja:"+prod+" que no la escribe")
				}
				ramas[o.Refina+"\x00"+prod] = append(ramas[o.Refina+"\x00"+prod], n.ID)
			}
			// externo (usuario/terceros): raíz legal de la cadena (C11) — nada que cotejar.
		}
	}

	for key, refs := range ramas {
		if len(refs) > 1 {
			art, _, _ := strings.Cut(key, "\x00")
			sort.Strings(refs)
			bad = append(bad, "cadena de '"+art+"' no lineal: "+strings.Join(refs, ", ")+" refinan desde el mismo escritor")
		}
	}

	// ciclos: caminar la cadena de cada refinador siguiendo sus fuentes caja→caja.
	byID := map[string]Box{}
	for _, n := range g.Nodes {
		byID[n.ID] = n
	}
	for _, n := range g.Nodes {
		if n.Contract == nil {
			continue
		}
		for _, o := range n.Contract.Entrega {
			if o.Refina == "" {
				continue
			}
			visited := map[string]bool{n.ID: true}
			cur := n
			art := o.Refina
			for {
				de, tiene := fuente(cur, art)
				if !tiene {
					break
				}
				prod, isCaja := deCaja(de)
				if !isCaja {
					break // raíz externa: cadena termina legal.
				}
				if visited[prod] {
					bad = append(bad, "ciclo en la cadena de refina de '"+art+"' (via "+prod+")")
					break
				}
				visited[prod] = true
				nxt, ok := byID[prod]
				if !ok || nxt.Contract == nil {
					break
				}
				cur = nxt
			}
		}
	}

	sort.Strings(bad)
	return veredictoDeLista(c, bad, "cadenas de refina lineales y coherentes")
}

// veredictoDeLista turns a list of violations into a CheckResult: empty = pass.
func veredictoDeLista(c Check, violaciones []string, okDetalle string) CheckResult {
	if len(violaciones) == 0 {
		return CheckResult{Check: c, Veredicto: VeredictoPass, Detalle: okDetalle}
	}
	return CheckResult{Check: c, Veredicto: VeredictoFail, Detalle: strings.Join(violaciones, "; ")}
}

// reachableFrom returns the set of states reachable from `start` over the transitions.
func reachableFrom(start string, txs []Transicion) map[string]bool {
	adj := map[string][]string{}
	for _, t := range txs {
		adj[t.De] = append(adj[t.De], t.A)
	}
	seen := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nxt := range adj[cur] {
			if !seen[nxt] {
				seen[nxt] = true
				queue = append(queue, nxt)
			}
		}
	}
	return seen
}
