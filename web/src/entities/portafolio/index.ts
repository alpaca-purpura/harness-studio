// entities/portafolio — identidad (home,id) + origen de la copia + deriva, como slice de
// dominio (S1-D6): tipos (espejo exacto del wire de Slice 0/1) + selectores puros (S1-D4/D8/D2)
// + fixtures honestas del E2E (S1-D del paquete). Public API only; import via
// "@/entities/portafolio", nunca deep paths (no-deep-import, dependency-cruiser).
//
// UI de dominio (chips/dot) llega en T3 — este barrel crece con el código (mismo patrón que
// entities/arnes/index.ts).

export {
  agruparPorEmpresa,
  filtrarEntradas,
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
