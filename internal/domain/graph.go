package domain

import "encoding/json"

// TipoEdge is the kind of relation between two nodes. invoca/escribe are directional
// (drawn with an arrow); lee (reading knowledge) is dotted without an arrow. Mirrors
// $defs.edge.tipo in graph.l0.schema.json.
type TipoEdge string

// The three relation kinds between nodes: invoca (calls into), lee (reads knowledge)
// and escribe (writes an artifact). See the TipoEdge comment for how each is drawn.
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
	ID               string   `json:"id,omitempty"`
	Nombre           string   `json:"nombre,omitempty"`      // human name — canonical for painting the arnés; fallback chain: plugin.json name → id (HS-12).
	Descripcion      string   `json:"descripcion,omitempty"` // short human description; fallback: plugin.json description (HS-12).
	Rol              string   `json:"rol,omitempty"`         // canonical (was `puesto`); the role the arnés serves.
	Proceso          string   `json:"proceso,omitempty"`     // the company process the arnés operationalizes.
	Empresas         []string `json:"empresas,omitempty"`    // N:M facet (S0-D3); UnmarshalJSON below tolerates the legacy scalar `empresa`.
	ReportaA         *string  `json:"reporta_a"`             // id of the arnés it reports to (organigrama). null = root — the META field is required by graph.l0, so nil must emit `null`, not be omitted.
	Canal            Canal    `json:"canal,omitempty"`
	Marketplace      string   `json:"marketplace,omitempty"`       // home autor-declarado, CRUDO (S0-D3): la procedencia de la copia vive en domain.Origen del Portafolio, jamás acá.
	Version          string   `json:"version,omitempty"`           // fuente: plugin.json.version vía el loader (D-DOM-1).
	FuenteManifiesto string   `json:"fuente_manifiesto,omitempty"` // "arnes.l0.json" | "plugin.json" — de dónde salió este manifiesto (BR-3).
	Fases            []Fase   `json:"fases,omitempty"`             // the phases THIS arnés declares (data, not a product enum).
	Spine            *Spine   `json:"spine,omitempty"`             // THIS arnés's work-state spine (data, not a product enum).
}

// UnmarshalJSON acepta el legacy escalar `"empresa":"x"` (pre S0-D3), normalizándolo a
// `Empresas:["x"]`; la forma nueva `"empresas":[...]` pasa directo. Si ambas vienen,
// `empresas` gana (es la forma canónica) — Marshal siempre emite solo `empresas`.
func (a *Arnes) UnmarshalJSON(b []byte) error {
	type alias Arnes // separa el tipo para que json.Unmarshal no reentre en este método.
	aux := struct {
		Empresa *string `json:"empresa,omitempty"`
		*alias
	}{alias: (*alias)(a)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if len(a.Empresas) == 0 && aux.Empresa != nil && *aux.Empresa != "" {
		a.Empresas = []string{*aux.Empresa}
	}
	return nil
}

// Spine is the FORM of an arnés's work-state machine — the shape only, never product
// values. Each arnés declares its own concrete states/transitions as DATA (the luana
// example's idea→…→released lives in a fixture, never here). The product owns the TYPE
// and the consistency checks against what an arnés declared, not the states themselves
// (agnosticism, VISION p3/p7). Mirrors arnes.spine in graph.l0.schema.json.
type Spine struct {
	Inicial      string               `json:"inicial"`                // the single entry state.
	Terminales   []string             `json:"terminales,omitempty"`   // accepting/terminal states.
	Estados      []string             `json:"estados"`                // every legal work-state.
	Categorias   map[string]Categoria `json:"categorias,omitempty"`   // optional state→category map (HS-12) — semantic layer over the per-arnés states.
	Transiciones []Transicion         `json:"transiciones,omitempty"` // the legal edges between states.
}

// Categoria is the FIXED semantic category a spine state may map to (HS-12, interop
// DevStudio — mirrors the ecosystem contract I-77 RN-28, the Azure "state categories"
// pattern). The five values ARE a product enum, deliberately: they are the
// process-agnostic semantic layer a console derives generic behavior from, while the
// state ids they classify remain per-arnés DATA — agnosticism (VISION p3/p7) intact.
type Categoria string

// The five fixed spine categories (identical to I-77 RN-28; immutable since v1).
const (
	CategoriaPropuesto  Categoria = "propuesto"
	CategoriaEnProgreso Categoria = "en-progreso"
	CategoriaCompletado Categoria = "completado"
	CategoriaDescartado Categoria = "descartado"
	CategoriaPausado    Categoria = "pausado"
)

// Valid reports whether c is one of the five fixed spine categories.
func (c Categoria) Valid() bool {
	switch c {
	case CategoriaPropuesto, CategoriaEnProgreso, CategoriaCompletado,
		CategoriaDescartado, CategoriaPausado:
		return true
	}
	return false
}

// Terminal reports whether c marks a terminal state — terminality is DERIVED from the
// category (I-77: categoría ∈ {completado, descartado}), never declared twice.
func (c Categoria) Terminal() bool {
	return c == CategoriaCompletado || c == CategoriaDescartado
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
	// Degradado marca un grafo cargado SIN manifiesto (S1-D27, honra nomenclatura-arnes.md
	// §2): el loader reconoció los nodos (rules/skills/hooks) pero no hay `arnes.l0.json` ni
	// `plugin.json` que selle la identidad. El Mapa lo observa igual con la marca roja
	// `manifiesto-ausente` — visible, nunca inventado; jamás se finge un arnés sellado.
	Degradado bool `json:"degradado,omitempty"`
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
