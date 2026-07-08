// entities/arnes — the agnostic component graph as a domain slice: types (mirroring the L0
// contract), geography selectors (Guardia · lanes · Base), the type→visual KIND map, and the
// ArnesNode card. Public API only; import via "@/entities/arnes", never deep paths.
//
// `devFullCycle` is the recorded dogfood fixture, exported for stories/mockup during Fase B.
// TEMPORAL: once api.getGraph lands (Hito 1) the live graph replaces it; the fixture stays for tests.

export {
  type ArtEdge,
  type ArtefactosMode,
  artEdges,
  CAP_GUTTER,
  type ChipArtefacto,
  type ConsumidorChip,
  type GutterPlan,
  planGutter,
  type RefEntrada,
  selectArtefactos,
  selectRefsEntrada,
} from "./model/artefactos"
export { DEF_CAMPO, DEF_VALOR, PROP_NOTE, SEC_TIP, tipDe } from "./model/doctrina"
export { KIND, type KindVisual } from "./model/kind"
export { handleFor, isCaja, isPropuesto, transLabel } from "./model/node-view"
export { alwFor, isDelPuesto } from "./model/proposals"
export {
  type Lane,
  selectBandaDesconocida,
  selectBase,
  selectEdges,
  selectGuardia,
  selectHallazgosConformance,
  selectLanes,
  selectSoporte,
  selectVieneDe,
  type VieneDe,
} from "./model/selectors"
export type {
  Arnes,
  Banda,
  Box,
  Canal,
  Clase,
  ConformanceCheck,
  ConformanceResult,
  Contract,
  Edge,
  Graph,
  Spine,
  TipoEdge,
  Transicion,
  Veredicto,
} from "./model/types"
export { cobranzaProveedores } from "./testing/cobranza-proveedores"
export { devFullCycle } from "./testing/dev-full-cycle"
export { luanaFeatureCycle } from "./testing/luana-feature-cycle"
export { ArnesNode } from "./ui/arnes-node"
export { ArtefactoChip } from "./ui/artefacto-chip"
