// widgets/marketplace — chrome del plano Marketplaces (S2) y de sus catálogos (S3/S4), paquete
// 2026-07-23-portafolio-agregar-marketplace. Props puras, CERO transporte
// (`fe-transporte-independiente`): el fetch/refetch/AbortController viven en
// `pages/shell/ui/portafolio-view.tsx`, que compone este widget con `widgets/portafolio`.
//
// **No importa `widgets/portafolio` ni al revés** (`no-sibling-widget-imports`, severidad
// `error`): dos widgets hermanos que se toquen crean un acoplamiento invisible; la página los
// compone.

export { MarketplaceCatalogo, type MarketplaceCatalogoProps } from "./ui/marketplace-catalogo"
export { MarketplaceList, type MarketplaceListProps } from "./ui/marketplace-list"
