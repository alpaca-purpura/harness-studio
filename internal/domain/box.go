// Package domain holds ArnesIA's core types: the agnostic component graph (the L0
// contract, meta.clase / I-75). It imports nothing internal — the graph is the truth,
// every agent is only an adapter. The JSON field tags mirror the single source of
// truth in arch/contracts/schema/graph.l0.schema.json and box.contract.schema.json;
// from those schemas the Go (indexer) and TS (React Flow map) types are generated.
package domain

// Clase is the L0 class of a component (meta.clase, I-75). The set is additive: it
// grows if the Claude Code surface adds types. Mirrors $defs.clase in
// graph.l0.schema.json.
type Clase string

// The seven first-class node types mapped in mockup v3 (color + shape + label).
const (
	ClaseSkill     Clase = "skill"
	ClaseAgente    Clase = "agente"
	ClaseHook      Clase = "hook"
	ClaseKnowledge Clase = "knowledge"
	ClaseMCP       Clase = "mcp"
	ClaseRegla     Clase = "regla"
	ClaseCommand   Clase = "command"
)

// Valid reports whether c is one of the known L0 classes.
func (c Clase) Valid() bool {
	switch c {
	case ClaseSkill, ClaseAgente, ClaseHook, ClaseKnowledge, ClaseMCP, ClaseRegla, ClaseCommand:
		return true
	default:
		return false
	}
}

// Banda is the region of the map's fixed geography where a node lives. Mirrors
// $defs.banda in graph.l0.schema.json.
type Banda string

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

const (
	CanalBeta      Canal = "beta"
	CanalEstable   Canal = "estable"
	CanalPropuesto Canal = "propuesto"
	CanalDeprecado Canal = "deprecado"
)

// Procedencia records where a node's data came from — the honesty axis of
// METODOLOGIA §4 ("medido" overrides "estimado"). Mirrors nodo.procedencia.
type Procedencia string

const (
	ProcMedido      Procedencia = "medido"
	ProcEstimado    Procedencia = "estimado"
	ProcDeclarado   Procedencia = "declarado"
	ProcInferido    Procedencia = "inferido"
	ProcNoDeclarado Procedencia = "no-declarado"
)

// Fase is the id of a process phase a box belongs to (null in Guardia/Base).
type Fase string

// Estado is the work-state transition a process box owns ("<entra> -> <sale>"),
// as declared in the box contract.
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
	Contract    *Contract   `json:"contract,omitempty"`
}

// IsCaja reports whether the box is a process box (the skill-front of a phase), i.e.
// it carries a contract with caja=true.
func (b Box) IsCaja() bool {
	return b.Contract != nil && b.Contract.Caja
}

// Contract is the `contract:` block of a process-box SKILL.md. Mirrors
// box.contract.schema.json — validating each instance against that schema is the
// eval-gate A4 fitness function.
type Contract struct {
	Caja     bool     `json:"caja"`
	Fase     string   `json:"fase,omitempty"`
	Estado   string   `json:"estado,omitempty"`
	Necesita []Input  `json:"necesita,omitempty"`
	Entrega  []Output `json:"entrega,omitempty"`
	Ruta     []Route  `json:"ruta,omitempty"`
	Gate     *Gate    `json:"gate,omitempty"`
}

// Input is a precondition the box needs to operate (contract.necesita). An input with
// no upstream producer is an orphan (a finding).
type Input struct {
	Art       string `json:"art"`
	De        string `json:"de"`
	Requerido *bool  `json:"requerido,omitempty"`
}

// Output is an as-code artifact the box produces (contract.entrega). An output nobody
// consumes is a dead-end (a finding).
type Output struct {
	Art string `json:"art"`
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

const (
	GateAuto    GateTipo = "auto"
	GateManual  GateTipo = "manual"
	GateParcial GateTipo = "parcial"
	GateNone    GateTipo = "none"
)

// Gate declares how a box's output is evaluated (contract.gate).
type Gate struct {
	Tipo    GateTipo `json:"tipo"`
	Detalle string   `json:"detalle,omitempty"`
}
