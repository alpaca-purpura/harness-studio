// entities/arnes — the agnostic component graph as a domain slice: types (mirroring the L0
// contract), geography selectors (Guardia · lanes · Base), the type→visual KIND map, and the
// ArnesNode card. Public API only; import via "@/entities/arnes", never deep paths.
//
// `devFullCycle` is the recorded dogfood fixture, exported for stories/mockup during Fase B.
// TEMPORAL: once api.getGraph lands (Hito 1) the live graph replaces it; the fixture stays for tests.

export { KIND, type KindVisual } from "./model/kind"
export { handleFor, isCaja, isPropuesto, transLabel } from "./model/node-view"
export { alwFor, isDelPuesto } from "./model/proposals"
export { type Lane, selectBase, selectEdges, selectGuardia, selectLanes } from "./model/selectors"
export type {
  Arnes,
  Banda,
  Box,
  Canal,
  Clase,
  Contract,
  Edge,
  Graph,
  Spine,
  TipoEdge,
  Transicion,
} from "./model/types"
export { devFullCycle } from "./testing/dev-full-cycle"
export { luanaFeatureCycle } from "./testing/luana-feature-cycle"
export { ArnesNode } from "./ui/arnes-node"
