// widgets/portafolio — la superficie del Portafolio (Slice 1, plan §2.6): props puras, CERO
// transporte (fe-transporte-independiente) — el fetch vive en pages/shell/ui/portafolio-view.tsx
// (T7). Public API only; import via "@/widgets/portafolio", nunca deep paths (no-deep-import).

export { PortafolioList, type PortafolioListProps } from "./ui/portafolio-list"
