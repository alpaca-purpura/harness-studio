// widgets/portafolio — la superficie del Portafolio (Slice 1, plan §2.6): props puras, CERO
// transporte (fe-transporte-independiente) — el fetch vive en pages/shell/ui/portafolio-view.tsx
// (T7). Public API only; import via "@/widgets/portafolio", nunca deep paths (no-deep-import).

export { PortafolioDrawer, type PortafolioDrawerProps } from "./ui/portafolio-drawer"
export { PortafolioList, type PortafolioListProps } from "./ui/portafolio-list"
export { PortafolioWizard, type PortafolioWizardProps } from "./ui/portafolio-wizard"
export {
  ResolverOrigenDialog,
  type ResolverOrigenDialogProps,
} from "./ui/resolver-origen-dialog"
