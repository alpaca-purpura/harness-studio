// entities/portafolio — identidad (home,id) + origen de la copia + deriva, como slice de
// dominio (S1-D6): tipos (espejo exacto del wire de Slice 0/1) + selectores puros (S1-D4/D8/D2)
// + fixtures honestas del E2E (S1-D del paquete) + UI de dominio (chips/dot, T3). Public API
// only; import via "@/entities/portafolio", nunca deep paths (no-deep-import, dependency-cruiser).

export {
  agruparPorEmpresa,
  agruparPorMarketplace,
  agruparPorProyecto,
  filtrarEntradas,
  gruposCandidatosDe,
  identificadorDe,
  idsColisionados,
  registriesDe,
  saludDe,
} from "./model/selectors"
export type {
  Candidato,
  Canonico,
  EntradaCorrupta,
  EntradaPortafolio,
  EslabonOrigen,
  EstadoDeriva,
  IdentidadArnes,
  Instalacion,
  LentePortafolio,
  OrigenCopia,
  PortafolioListado,
  SaludPortafolio,
  TipoInstalacion,
} from "./model/types"
export {
  entradaCanonicaCompleta,
  entradaHarnessEnDeriva,
  entradaProyectoInstaladoProvisional,
  entradasDemo,
} from "./testing/entradas"
export {
  AvisoChip,
  colorDeterminista,
  DerivaChip,
  DotSaludPortafolio,
  EmblemaInicial,
  TipoInstalacionChip,
} from "./ui/chips"
