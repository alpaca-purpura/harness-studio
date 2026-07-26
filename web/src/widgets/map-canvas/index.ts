// widgets/map-canvas — the map surface (Guardia · lanes · Base + SVG edges). Canvas side of the
// canvas⊥chrome boundary: imports only shared/* and entities/arnes, never chrome (rail/dock/shell).
// MapBar is the map's CHROME (arnés META + layer tablist, architecture §2.2) — it lives here but
// is mounted by the shell/page ABOVE the pure canvas, never by MapCanvas itself.
//
// La capa «Mejora» (paquete 2026-07-24) vive en esta MISMA slice y no en una propia: una slice
// `widgets/mejora/` chocaría con `no-sibling-widget-imports`, que está en `error`. Es también el
// único widget que puede importar las DOS entities (`arnes` y `telemetria`), y por eso es acá
// donde `CifraCaja` se compone en las props primitivas del nodo (D18).
export {
  coberturaEsParcial,
  DETECTORES_DEL_MVP,
  hayDatosAtribuibles,
} from "./model/capa-mejora"
export type { Capa, LayerDef } from "./model/layers"
export { LAYERS } from "./model/layers"
export { FranjaMejora, type FranjaMejoraProps } from "./ui/franja-mejora"
export { Inspector } from "./ui/inspector"
export {
  InspectorMejora,
  type InspectorMejoraProps,
  type JoinDeLaCaja,
} from "./ui/inspector-mejora"
export { MapBar } from "./ui/map-bar"
export { MapCanvas } from "./ui/map-canvas"
export {
  PoliticaDatosDialog,
  type PoliticaDatosDialogProps,
} from "./ui/politica-datos-dialog"
export { PuntoMejoraCard, type PuntoMejoraCardProps } from "./ui/punto-mejora-card"
export { PuntosMejoraList, type PuntosMejoraListProps } from "./ui/puntos-mejora-list"
