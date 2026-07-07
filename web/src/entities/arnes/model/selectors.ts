// Pure projections of a Graph into the map's fixed geography: Guardia band (top) · one lane
// per declared fase (middle, in arnes.fases order) · Base band (bottom). These are the SSOT
// of "which node goes where"; the canvas only lays out what these return. No React, no
// transport — safe to import from anywhere (kept in entities/*/model per FSD).

import type { Banda, Box, ConformanceResult, Edge, Graph, TipoEdge } from "./types"

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

// Every banda the fixed geography knows how to place. Graphs arrive as JSON at runtime —
// nothing guarantees the TS union — so anything else (absent OR outside the contract enum)
// is «banda desconocida».
const KNOWN_BANDS: ReadonlySet<string> = new Set<string>(["guardia", "fase", ...BASE_BANDS])

// Where banda-desconocida nodes land: the `base` band of the Base region (canónico VISION A6),
// the catch-all support band the canvas already renders.
const FALLBACK_BAND: Banda = "base"

// selectGuardia — the transversal hooks band (top).
export function selectGuardia(g: Graph): Box[] {
  return g.nodos.filter((n) => n.banda === "guardia")
}

// selectBandaDesconocida — ids of the nodes whose banda falls in NO known region (absent or
// unknown). Reconciliación honesta (arch/contracts/nomenclatura-arnes.md §4.5): these nodes
// must stay VISIBLE — selectSoporte folds them into FALLBACK_BAND, and this set is the
// detectable mark for consumers. El badge visual sobre el nodo llega cuando el WIP del Hito 2
// (node-view/arnes-node) landee — hasta entonces el dato queda expuesto aquí, sin tocar esos
// archivos en obra.
export function selectBandaDesconocida(g: Graph): ReadonlySet<string> {
  const ids = new Set<string>()
  for (const n of g.nodos) {
    if (n.banda === undefined || !KNOWN_BANDS.has(n.banda)) ids.add(n.id)
  }
  return ids
}

// selectBase — knowledge / rules / mcp / library bands (bottom), PLUS every banda-desconocida
// node: they belong nowhere else, and invisible is a lie (§4.5).
export function selectBase(g: Graph): Box[] {
  const desconocida = selectBandaDesconocida(g)
  return g.nodos.filter(
    (n) => (n.banda !== undefined && BASE_BANDS.has(n.banda)) || desconocida.has(n.id),
  )
}

// selectSoporte — the nodes of ONE support band, exactly as the canvas groups the Base region
// (map-canvas iterates SUPPORT_BANDS). The fallback band also receives the banda-desconocida
// nodes so they render visibly instead of vanishing in silence.
export function selectSoporte(g: Graph, banda: Banda): Box[] {
  if (banda !== FALLBACK_BAND) return g.nodos.filter((n) => n.banda === banda)
  const desconocida = selectBandaDesconocida(g)
  return g.nodos.filter((n) => n.banda === FALLBACK_BAND || desconocida.has(n.id))
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

// VieneDe — one inverse edge of a node: who acts ON it and how (RF-89).
export interface VieneDe {
  de: string
  tipo: TipoEdge
}

// selectVieneDe — the INVERSE edges of a node, derived from the already-loaded
// graph.edges (never declared by hand): every edge whose target is nodeId, as
// (source, tipo). Empty array ⇒ the drawer hides the section (faithful to the data).
export function selectVieneDe(g: Graph, nodeId: string): VieneDe[] {
  return (g.edges ?? []).filter((e) => e.a === nodeId).map((e) => ({ de: e.de, tipo: e.tipo }))
}

// selectHallazgosConformance — the RED checks (fail|error) of a harness conformance
// report that mention nodeId (RF-91). The arnes-scope checks are GRAPH-level with the
// violating node ids listed inside `detalle` («id (motivo); id2 (…)», domain.
// veredictoDeLista) — per-node attribution does not exist as data in the report, so a
// word-boundary match over detalle is the honest filter (see plan-implementacion.md §3;
// if the engine ever attributes nodes as data, this selector simplifies).
export function selectHallazgosConformance(
  results: readonly ConformanceResult[],
  nodeId: string,
): ConformanceResult[] {
  const word = new RegExp(`(^|[^\\w-])${nodeId.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}($|[^\\w-])`)
  return results.filter(
    (r) => (r.veredicto === "fail" || r.veredicto === "error") && word.test(r.detalle ?? ""),
  )
}
