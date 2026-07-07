// widgets/map-canvas — the map surface (Guardia · lanes · Base + SVG edges). Canvas side of the
// canvas⊥chrome boundary: imports only shared/* and entities/arnes, never chrome (rail/dock/shell).
// MapBar is the map's CHROME (arnés META + layer tablist, architecture §2.2) — it lives here but
// is mounted by the shell/page ABOVE the pure canvas, never by MapCanvas itself.
export type { Capa } from "./model/layers"
export { Inspector } from "./ui/inspector"
export { MapBar } from "./ui/map-bar"
export { MapCanvas } from "./ui/map-canvas"
