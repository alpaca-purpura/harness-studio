package domain

// TipoEdge is the kind of relation between two nodes. invoca/escribe are directional
// (drawn with an arrow); lee (reading knowledge) is dotted without an arrow. Mirrors
// $defs.edge.tipo in graph.l0.schema.json.
type TipoEdge string

const (
	EdgeInvoca  TipoEdge = "invoca"
	EdgeLee     TipoEdge = "lee"
	EdgeEscribe TipoEdge = "escribe"
)

// Edge is a relation between two nodes. Mirrors $defs.edge.
type Edge struct {
	De   string   `json:"de"`
	A    string   `json:"a"`
	Tipo TipoEdge `json:"tipo"`
}

// Arnes is the manifiesto of the harness that owns a graph (a rol × proceso of a
// company). It carries the META de enganche (rol·proceso·reporta-a·empresa, VISION
// §Linaje) — the seam to the future external L1 (organigrama) system — and, as DATA
// (not a product constant), THIS arnés's own process shape: its Fases and its Spine of
// work-states. ArnesIA is agnostic to any particular process, so those values live here
// per-arnés, never as an enum in the core. Mirrors the top-level `arnes` object in
// graph.l0.schema.json.
type Arnes struct {
	ID          string  `json:"id,omitempty"`
	Rol         string  `json:"rol,omitempty"`     // canonical (was `puesto`); the role the arnés serves.
	Proceso     string  `json:"proceso,omitempty"` // the company process the arnés operationalizes.
	Empresa     string  `json:"empresa,omitempty"`
	ReportaA    *string `json:"reporta_a"` // id of the arnés it reports to (organigrama). null = root — the META field is required by graph.l0, so nil must emit `null`, not be omitted.
	Canal       Canal   `json:"canal,omitempty"`
	Marketplace string  `json:"marketplace,omitempty"`
	Fases       []Fase  `json:"fases,omitempty"` // the phases THIS arnés declares (data, not a product enum).
	Spine       *Spine  `json:"spine,omitempty"` // THIS arnés's work-state spine (data, not a product enum).
}

// Spine is the FORM of an arnés's work-state machine — the shape only, never product
// values. Each arnés declares its own concrete states/transitions as DATA (the luana
// example's idea→…→released lives in a fixture, never here). The product owns the TYPE
// and the consistency checks against what an arnés declared, not the states themselves
// (agnosticism, VISION p3/p7). Mirrors arnes.spine in graph.l0.schema.json.
type Spine struct {
	Inicial      string       `json:"inicial"`                // the single entry state.
	Terminales   []string     `json:"terminales,omitempty"`   // accepting/terminal states.
	Estados      []string     `json:"estados"`                // every legal work-state.
	Transiciones []Transicion `json:"transiciones,omitempty"` // the legal edges between states.
}

// Transicion is one legal edge of a Spine (from a state to a state).
type Transicion struct {
	De string `json:"de"`
	A  string `json:"a"`
}

// TieneEstado reports whether e is one of the spine's declared states.
func (s Spine) TieneEstado(e string) bool {
	for _, x := range s.Estados {
		if x == e {
			return true
		}
	}
	return false
}

// TransicionLegal reports whether de→a is a declared legal transition of the spine.
func (s Spine) TransicionLegal(de, a string) bool {
	for _, t := range s.Transiciones {
		if t.De == de && t.A == a {
			return true
		}
	}
	return false
}

// UnidadDeTrabajo is the agnostic work-item that flows through an arnés's Spine — the
// order-of-work, deliberately NOT called a "Story" (agnosticism: no methodology-specific
// vocabulary leaks into the core). Its Estado/Fase are DATA validated against the arnés's
// declared Spine and Fases, never a product enum.
type UnidadDeTrabajo struct {
	ID     string `json:"id"`
	Estado string `json:"estado,omitempty"`
	Fase   string `json:"fase,omitempty"`
}

// Graph is the agnostic component graph of one arnés: harness manifiesto plus its nodes
// and edges. Mirrors the top-level object of graph.l0.schema.json (only Nodes is
// required there; Arnes is a pointer so a graph fragment with no manifiesto omits it
// cleanly instead of emitting an invalid empty `arnes:{}`).
type Graph struct {
	Arnes *Arnes `json:"arnes,omitempty"`
	Nodes []Box  `json:"nodos"`
	Edges []Edge `json:"edges,omitempty"`
}

// NodeByID returns the node with the given id and whether it was found.
func (g Graph) NodeByID(id string) (Box, bool) {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return Box{}, false
}
