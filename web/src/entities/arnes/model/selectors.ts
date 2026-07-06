// Pure projections of a Graph into the map's fixed geography: Guardia band (top) · one lane
// per declared fase (middle, in arnes.fases order) · Base band (bottom). These are the SSOT
// of "which node goes where"; the canvas only lays out what these return. No React, no
// transport — safe to import from anywhere (kept in entities/*/model per FSD).

import type { Banda, Box, Edge, Graph } from "./types"

// A lane is one process phase with the boxes that declared that fase.
export interface Lane {
  fase: string
  nodos: Box[]
}

// Bands that render in the bottom Base region (everything that is neither Guardia nor Fase).
const BASE_BANDS: ReadonlySet<Banda> = new Set<Banda>([
  "base",
  "libreria-expertos",
  "meta-harness",
  "marcas-dormidas",
  "terceros",
])

// selectGuardia — the transversal hooks band (top).
export function selectGuardia(g: Graph): Box[] {
  return g.nodos.filter((n) => n.banda === "guardia")
}

// selectBase — knowledge / rules / mcp / library bands (bottom).
export function selectBase(g: Graph): Box[] {
  return g.nodos.filter((n) => n.banda !== undefined && BASE_BANDS.has(n.banda))
}

// selectLanes — the phase lanes, ordered by the arnés's declared `fases`. Fase-band nodes
// with a fase not in the declared list are appended in first-seen order (honest: they still
// render, rather than vanishing). Phases with no node still get an (empty) lane so the
// process reads end-to-end.
export function selectLanes(g: Graph): Lane[] {
  const declared = g.arnes?.fases ?? []
  const faseNodes = g.nodos.filter((n) => n.banda === "fase")
  const order: string[] = [...declared]
  for (const n of faseNodes) {
    const f = n.fase ?? ""
    if (f && !order.includes(f)) order.push(f)
  }
  return order.map((fase) => ({
    fase,
    nodos: faseNodes.filter((n) => (n.fase ?? "") === fase),
  }))
}

// selectEdges — the relations (invoca/lee/escribe).
export function selectEdges(g: Graph): Edge[] {
  return g.edges ?? []
}
