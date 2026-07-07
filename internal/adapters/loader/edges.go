package loader

import (
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// derivarEdges deriva las relaciones del grafo desde los contratos de las cajas — edges
// DERIVADOS, jamás declarados sueltos (§4.3). Reglas finales v1, calibradas para reproducir
// exactamente el set del fixture dev-full-cycle.graph.json:
//
//	R1 (lee):    caja C con necesita[].de == "base:<id>" y nodo <id> presente
//	             → {de: C, a: <id>, tipo: lee}. Dirección lector→conocimiento, como lo
//	             dibuja el Mapa (punteado sin flecha; el `de` es quien consulta la Base).
//	R2 (invoca): caja C con necesita[].de == "caja:<id>" y nodo <id> presente
//	             → {de: <id>, a: C, tipo: invoca} — el upstream invoca/entrega al
//	             consumidor declarado; el DAG de hand-offs es el espejo de las necesidades.
//
// `ruta[].a` NO genera edges en v1 (decisión documentada): §4.3 de la nomenclatura deriva
// edges de necesita/hooks/entrega — no de ruta — y derivar ruta→invoca produciría el edge de
// rework reviewer→builder («hay hallazgos que corregir») que el fixture deliberadamente NO
// dibuja: las ramas de rework/escalate viven en el contrato (ruta/handoff) y en el spine
// (transición review→build), no como cableado del Mapa. Un input `caja:<id>`/`base:<id>`
// cuyo nodo no está presente tampoco genera edge: ese hueco es un hallazgo del gate de
// conformance (input huérfano / declarado-sin-archivo §4.5), no un edge inventado.
func derivarEdges(nodos []domain.Box) []domain.Edge {
	presentes := make(map[string]bool, len(nodos))
	for _, n := range nodos {
		presentes[n.ID] = true
	}

	var edges []domain.Edge
	vistos := map[domain.Edge]bool{} // dedup: el mismo edge derivado dos veces cuenta una.
	agrega := func(e domain.Edge) {
		if !vistos[e] {
			vistos[e] = true
			edges = append(edges, e)
		}
	}

	for _, n := range nodos {
		if !n.IsCaja() {
			continue
		}
		for _, in := range n.Contract.Necesita {
			if id, ok := strings.CutPrefix(in.De, "base:"); ok && presentes[id] {
				agrega(domain.Edge{De: n.ID, A: id, Tipo: domain.EdgeLee}) // R1
			}
			if id, ok := strings.CutPrefix(in.De, "caja:"); ok && presentes[id] {
				agrega(domain.Edge{De: id, A: n.ID, Tipo: domain.EdgeInvoca}) // R2
			}
		}
	}
	return edges
}
