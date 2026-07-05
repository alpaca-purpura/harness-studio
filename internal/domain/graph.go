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

// Arnes is the metadata of the harness that owns a graph (a rol × proceso of a
// company). Mirrors the top-level `arnes` object in graph.l0.schema.json.
type Arnes struct {
	ID          string  `json:"id,omitempty"`
	Puesto      string  `json:"puesto,omitempty"`
	Empresa     string  `json:"empresa,omitempty"`
	ReportaA    *string `json:"reporta_a,omitempty"`
	Canal       Canal   `json:"canal,omitempty"`
	Marketplace string  `json:"marketplace,omitempty"`
}

// Graph is the agnostic component graph of one arnés: harness metadata plus its nodes
// and edges. Mirrors the top-level object of graph.l0.schema.json (only Nodes is
// required there).
type Graph struct {
	Arnes Arnes  `json:"arnes,omitempty"`
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
