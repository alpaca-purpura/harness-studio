// Package domain holds ArnesIA's core types: the agnostic component graph (the L0
// contract, meta.clase / I-75) and the fused box contract (doctrina v1, HS-08). It
// imports nothing internal — the graph is the truth, every agent is only an adapter.
// The JSON field tags mirror the single source of truth in
// arch/contracts/schema/graph.l0.schema.json and box.contract.schema.json; from those
// schemas the Go (indexer) and TS (React Flow map) types are generated.
//
// Agnosticism rule (VISION p3/p7): the product enumerates only DOCTRINE axes as
// constants — Clase, Arquetipo, PerfilHarness, Banda, GateTipo, Canal. Process shape
// (Fase, Estado, the Spine values) is DATA each arnés declares, never a product
// constant: those are `string` with no enum/Valid(). See internal/domain/graph.go.
package domain

// Clase is the L0 class of a component (meta.clase, I-75): which CC-native primitive
// this node IS. It is ONE of three orthogonal axes and must not be conflated with the
// others: Clase (which primitive) ⊥ Banda (map region) ⊥ PerfilHarness (how it runs).
// Additive: the set grows only if the Claude Code surface adds a placeable primitive.
// Mirrors $defs.clase in graph.l0.schema.json.
type Clase string

// The ten CC-native placeable primitives, aligned to the knowledge/ element nodes that
// describe a component that can live as a graph node. Labels are canonical (subagent,
// rule — not the legacy agente/regla). Deliberately EXCLUDED from clase: `headless`
// (a run-mode / maquinaria, not a node type) and `harness-profile` (the orthogonal
// PerfilHarness axis). The legacy `knowledge` value folds into `rule` — Base-band
// knowledge is CC-native memory/imports (a rule), per the firewall (METODOLOGIA §8.6).
const (
	ClaseSkill       Clase = "skill"
	ClaseSubagent    Clase = "subagent"
	ClaseHook        Clase = "hook"
	ClaseRule        Clase = "rule"
	ClaseCommand     Clase = "command"
	ClaseMCP         Clase = "mcp"
	ClasePlugin      Clase = "plugin"
	ClaseSettings    Clase = "settings"
	ClaseOutputStyle Clase = "output-style"
	ClaseStatusline  Clase = "statusline"

	// ClaseNoReconocido NO es una 11ª primitiva: es el marcador de reconciliación
	// honesta del loader (nomenclatura-arnes.md §4.5, D-c firmada HS-10) — un archivo
	// que ningún reconocedor entiende se emite VISIBLE con esta clase y un check warn,
	// jamás se oculta ni se descarta. Excluida de Valid(): las 10 canónicas siguen
	// siendo 10.
	ClaseNoReconocido Clase = "no-reconocido"
)

// Valid reports whether c is one of the ten canonical L0 classes.
func (c Clase) Valid() bool {
	switch c {
	case ClaseSkill, ClaseSubagent, ClaseHook, ClaseRule, ClaseCommand,
		ClaseMCP, ClasePlugin, ClaseSettings, ClaseOutputStyle, ClaseStatusline:
		return true
	default:
		return false
	}
}

// Arquetipo is the FORM of the work a box does (METODOLOGIA §8.1) — autonomy ≠
// automation. It is a doctrine axis (product constant), orthogonal to PerfilHarness.
type Arquetipo string

const (
	// ArqPipeline — deterministic / verifiable. Output = exact artifact; gate auto;
	// strict document-as-cache. (Automation.)
	ArqPipeline Arquetipo = "pipeline"
	// ArqExcepcion — happy path + rare cases. Output = invariants; gate at branches.
	// (Framed autonomy, bounded.)
	ArqExcepcion Arquetipo = "excepcion"
	// ArqAbierto — generative / relational. Restricts scope, not steps; may keep session
	// state (exempt from strict document-as-cache — see PrecedeArquetipoSobrePerfil).
	// (Full framed autonomy.)
	ArqAbierto Arquetipo = "abierto"
	// ArqNoArnesar — pure judgement that does not decompose into contracted parts. NOT a
	// box: classified explicitly and left out of the graph. Knowing when NOT to arnesar
	// is doctrine, not omission (METODOLOGIA §8.1).
	ArqNoArnesar Arquetipo = "no-arnesar"
)

// Valid reports whether a is one of the four archetypes.
func (a Arquetipo) Valid() bool {
	switch a {
	case ArqPipeline, ArqExcepcion, ArqAbierto, ArqNoArnesar:
		return true
	default:
		return false
	}
}

// PerfilHarness is HOW a box executes its loop (METODOLOGIA §8.2, nodo harness-profile)
// — the second axis of the contract, orthogonal to Clase. T4 (shell/session) is NOT a
// box, so it is not a valid box profile.
type PerfilHarness string

const (
	// PerfilT1 — single pass, no loop, no subagent; cheap model, low effort.
	PerfilT1 PerfilHarness = "T1"
	// PerfilT2 — multi-step, stateful, often interactive; document-as-cache obligatory
	// unless Arquetipo is abierto (PrecedeArquetipoSobrePerfil).
	PerfilT2 PerfilHarness = "T2"
	// PerfilT3 — unattended worker loop; the Go conductor owns the loop (never the Agent
	// SDK): spawns `claude -p --max-turns N`, reads result+status, repair-cap, blocked→handoff.
	PerfilT3 PerfilHarness = "T3"
)

// Valid reports whether p is one of the three box profiles (T4 = shell, not a box).
func (p PerfilHarness) Valid() bool {
	switch p {
	case PerfilT1, PerfilT2, PerfilT3:
		return true
	default:
		return false
	}
}

// RequiereDocumentAsCache resolves B8 (the imported abierto×T2 contradiction):
// Arquetipo takes precedence over PerfilHarness for document-as-cache. A box requires
// strict document-as-cache when its profile is T2/T3 AND its archetype is not abierto;
// an `abierto` box is exempt (it keeps session state with a distillation at close),
// even at T2/T3. Inherits DAOP [PENDIENTE 7.A]. See METODOLOGIA §8.3.
func RequiereDocumentAsCache(arq Arquetipo, perfil PerfilHarness) bool {
	if arq == ArqAbierto {
		return false
	}
	return perfil == PerfilT2 || perfil == PerfilT3
}

// Banda is the region of the map's fixed geography where a node lives — the map-region
// axis, orthogonal to Clase. Mirrors $defs.banda in graph.l0.schema.json.
type Banda string

// The map's fixed regions, as $defs.banda enumerates them: the Guardia band (hooks)
// across the top, the per-fase process lanes, the Base band (knowledge) at the bottom,
// plus the side shelves (expert library, meta-harness, dormant marks, third-party).
const (
	BandaGuardia         Banda = "guardia"
	BandaFase            Banda = "fase"
	BandaBase            Banda = "base"
	BandaLibreriaExperto Banda = "libreria-expertos"
	BandaMetaHarness     Banda = "meta-harness"
	BandaMarcasDormidas  Banda = "marcas-dormidas"
	BandaTerceros        Banda = "terceros"
)

// Canal is the release channel of an arnés or a node. Superset of both enums in the
// schema (arnes.canal = beta|estable; nodo.canal adds propuesto|deprecado).
type Canal string

// The four channels: an arnés releases as beta|estable (KIT-06 release train); a node
// additionally moves through propuesto (not yet accepted) and deprecado (on the way out).
const (
	CanalBeta      Canal = "beta"
	CanalEstable   Canal = "estable"
	CanalPropuesto Canal = "propuesto"
	CanalDeprecado Canal = "deprecado"
)

// Procedencia records where a node's data came from — the honesty axis of
// METODOLOGIA §4 ("medido" overrides "estimado"). Mirrors nodo.procedencia.
type Procedencia string

// The five provenance grades of nodo.procedencia. "medido" (real telemetry) always
// overrides "estimado"; "no-declarado" makes the absence of data explicit, never silent.
const (
	ProcMedido      Procedencia = "medido"
	ProcEstimado    Procedencia = "estimado"
	ProcDeclarado   Procedencia = "declarado"
	ProcInferido    Procedencia = "inferido"
	ProcNoDeclarado Procedencia = "no-declarado"
)

// Origen is nodo.origen (graph.l0.schema.json, HS-09 Fase D): who put the node in the
// INSTANCE of the arnés — the agnostic kit ("estandar") vs the rol·proceso·empresa
// onboarding ("del-puesto"). The PROVISIONER stamps it (never the author); orthogonal to
// Procedencia (data honesty) and Canal (release).
type Origen string

// The two instance origins of nodo.origen.
const (
	OrigenEstandar  Origen = "estandar"
	OrigenDelPuesto Origen = "del-puesto"
)

// Fase is the id of a process phase a box belongs to (null in Guardia/Base). It is
// DATA the arnés declares (see Arnes.Fases), NOT a product constant — deliberately a
// bare string with no enum: ArnesIA is agnostic to any arnés's process (VISION p3/p7).
type Fase string

// Estado is the work-state transition a process box owns ("<entra> -> <sale>"), as
// declared in the box contract. Like Fase it is DATA (validated against the arnés's
// declared Spine, not a product enum) — the agnosticism guarantee.
type Estado string

// Box is a node in the agnostic graph — a component of an arnés. Mirrors $defs.nodo
// in graph.l0.schema.json (which allows additional properties, so Estado/ReportaA are
// schema-legal convenience fields). Only ID, Clase and Nombre are required.
type Box struct {
	ID          string      `json:"id"`
	Clase       Clase       `json:"clase"`
	Nombre      string      `json:"nombre"`
	Banda       Banda       `json:"banda,omitempty"`
	Fase        Fase        `json:"fase,omitempty"`
	Estado      Estado      `json:"estado,omitempty"`
	ReportaA    string      `json:"reporta_a,omitempty"`
	Canal       Canal       `json:"canal,omitempty"`
	FuentePath  string      `json:"fuente_path,omitempty"`
	Procedencia Procedencia `json:"procedencia,omitempty"`
	// Origen (L0 $defs.nodo.origen, HS-09 Fase D): estandar|del-puesto, lo estampa el
	// provisioner. Sin este espejo el daemon TIRABA el campo al round-trip por el struct.
	Origen   Origen    `json:"origen,omitempty"`
	Contract *Contract `json:"contract,omitempty"`
}

// IsCaja reports whether the box is a process box (the skill-front of a phase), i.e.
// it carries a contract with caja=true.
func (b Box) IsCaja() bool {
	return b.Contract != nil && b.Contract.Caja
}

// Contract is the fused `contract:` block of a process-box SKILL.md (METODOLOGIA §3,
// doctrina v1). One contract, three perpendicular axes: INTENCIÓN (what the box
// promises), CLASIFICACIÓN (clase/arquetipo/perfil), CABLEADO (how it wires) and
// ACEPTACIÓN (how its output is proven). Mirrors box.contract.schema.json — validating
// each instance against that schema IS the eval-gate A4 fitness function.
type Contract struct {
	// ── INTENCIÓN ──
	Why          string       `json:"why,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Constraints  []string     `json:"constraints,omitempty"`
	NonGoals     []string     `json:"non_goals,omitempty"`

	// ── CLASIFICACIÓN (tres ejes ortogonales) ──
	Clase     Clase         `json:"clase,omitempty"`
	Arquetipo Arquetipo     `json:"arquetipo,omitempty"`
	Perfil    PerfilHarness `json:"perfil_harness,omitempty"`

	// ── CABLEADO (it.10, intacto) ──
	Caja     bool     `json:"caja"`
	Fase     string   `json:"fase,omitempty"`
	Estado   string   `json:"estado,omitempty"`
	Necesita []Input  `json:"necesita,omitempty"`
	Entrega  []Output `json:"entrega,omitempty"`
	Ruta     []Route  `json:"ruta,omitempty"`

	// ── ACEPTACIÓN ──
	Gate    *Gate    `json:"gate,omitempty"`
	Handoff *Handoff `json:"handoff,omitempty"`
}

// Capability is one promised outcome of a box (contract.capabilities), with a stable id
// across contract versions and a concrete, verifiable success signal.
type Capability struct {
	ID      string `json:"id"`
	What    string `json:"what"`
	Success string `json:"success"`
}

// Input is a precondition the box needs to operate (contract.necesita). An input with
// no upstream producer is an orphan (a finding).
type Input struct {
	Art       string `json:"art"`
	De        string `json:"de"`
	Requerido *bool  `json:"requerido,omitempty"`
}

// Output is an as-code artifact the box produces (contract.entrega). An output nobody
// consumes is a dead-end (a finding). EscritorUnico declares the mutation contract: one
// authorized writer per artifact (two boxes writing the same art = a finding). Refina
// (D9, franja-artefactos) declares this output as a REVISION of an existing art: the
// only legal gate to multi-writing, always a linear chain — the version (v2, v3…) is
// derived from the chain, never declared.
type Output struct {
	Art           string `json:"art"`
	EscritorUnico *bool  `json:"escritor_unico,omitempty"`
	Refina        string `json:"refina,omitempty"`
}

// Route is a conditional hand-off target (contract.ruta): the real DAG of happy
// path / rework / escalate.
type Route struct {
	A  string `json:"a"`
	Si string `json:"si,omitempty"`
}

// GateTipo is how the box's output is evaluated. GateNone is honest: the skill has no
// real eval (the gap of principle 10 turned into data). Mirrors contract.gate.tipo.
type GateTipo string

// The four gate types: auto (the Gherkin acceptance IS the executable eval), manual
// (a human signs off), parcial (mixed), none (the box declares it has no real eval —
// an honest gap, never a hidden one).
const (
	GateAuto    GateTipo = "auto"
	GateManual  GateTipo = "manual"
	GateParcial GateTipo = "parcial"
	GateNone    GateTipo = "none"
)

// Gate declares how a box's output is evaluated (contract.gate). When Tipo=auto the
// Aceptacion (Gherkin) IS the executable eval; Evidencia is the audit record emitted
// (telemetría de nacimiento, principle 9).
type Gate struct {
	Tipo       GateTipo     `json:"tipo"`
	Detalle    string       `json:"detalle,omitempty"`
	Aceptacion []Acceptance `json:"aceptacion,omitempty"`
	Evidencia  string       `json:"evidencia,omitempty"`
}

// Acceptance is one Gherkin clause of a box's acceptance gate (gate.aceptacion) — the
// gate.detalle turned executable.
type Acceptance struct {
	Given string `json:"given"`
	When  string `json:"when"`
	Then  string `json:"then"`
}

// Handoff makes the P6/Guardia boundary a datum: when a high-risk / non-converging box
// escalates and to whom (contract.handoff).
type Handoff struct {
	Cuando string `json:"cuando"`
	A      string `json:"a"`
}
