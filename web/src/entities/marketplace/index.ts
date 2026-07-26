// entities/marketplace — el estante de lo que vendemos y el espejo de si el cliente coincide
// (AG-D8, paquete 2026-07-23-portafolio-agregar-marketplace): tipos (espejo exacto del wire de
// design.md §3.1/§3.2/§3.5/§7) + PRESENTADORES puros (la situación la calcula el dominio Go,
// design.md §2 C1) + fixtures con shape REAL de esta máquina + chips de dominio.
//
// Public API only; import via "@/entities/marketplace", nunca deep paths (`no-deep-import`,
// dependency-cruiser). **Esta entidad NO importa `entities/portafolio`** (`steiger
// fsd/no-cross-imports`): los átomos compartidos (`DotSaludPortafolio`, `AvisoChip`) los importa
// el WIDGET, que sí puede ver las dos entidades.

export {
  accionDeFila,
  esNavegable,
  etiquetaDeSituacion,
  filtrarEntradasCatalogo,
  filtrarPorSituacion,
  ordenarMarketplaces,
  rotuloDeTipoSituacion,
  type SaludVisual,
  situacionesDisponibles,
  textoDeLectura,
  tonoDeSituacion,
} from "./model/selectors"
export type {
  Accion,
  AccionCatalogo,
  CaminoTraer,
  CandidatoOrigen,
  CandidatosOrigen,
  Catalogo,
  ClaseMarketplace,
  EntradaCatalogo,
  EntradaCorruptaMarketplace,
  EslabonMarketplace,
  EstadoLectura,
  EstadoTraer,
  ListadoMarketplaces,
  MarketplaceConocido,
  ProcedenciaVersion,
  ResultadoTraer,
  SituacionCatalogo,
  SourceCatalogo,
  TipoLectura,
  TipoSituacion,
  TipoSource,
  Validacion,
  VersionCatalogo,
} from "./model/types"
export {
  candidatosOrigenDemo,
  catalogo273,
  catCruceDebil,
  catDegradadoConCache,
  catLas6Situaciones,
  catNull,
  catOficialMuestra,
  catPrenter,
  catReferenciaLas6,
  catTruncado,
  catVacio,
  entradasOficialMuestra,
  marketplacesDemo,
  mkConDiscrepancia,
  mkNoLeido,
  mkOficialReferencia,
  mkPrenterPropio,
  mkSinAcceso,
  mkUrlNoResuelve,
  situacionesLas6,
  situacionesLas6Referencia,
  validacionCaveman,
  validacionPrenter,
} from "./testing/marketplaces"
export { CanalChip, ClaseChip, EstadoLecturaChip, SituacionChip, ViaChip } from "./ui/chips"
