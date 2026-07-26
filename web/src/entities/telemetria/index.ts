// Public API de la entidad `telemetria` (FSD). Los consumidores importan de acá, jamás por
// deep-import a `model/` o `ui/`.
//
// 🔴 `entities/arnes` NO importa esta entity. Nunca, por ningún escape (D18): el nodo del canvas
// recibe PRIMITIVOS y es el widget `map-canvas` —que sí puede importar las dos— el que compone
// `CifraCaja → props`. La igualdad literal del copy entre las dos superficies la ata la story
// `CopyConfianzaEsUnaSola`.

export {
  ariaTendencia,
  direccionTendencia,
  ETIQUETA_CONFIANZA,
  entero,
  etiquetaConfianza,
  etiquetaDetector,
  marcaPrincipal,
  NOMBRE_DETECTOR,
  pct,
  type SegmentoCobertura,
  segmentosCobertura,
  TITULO_CONFIANZA,
  TONO_CONFIANZA,
  tituloConfianza,
  tonoConfianza,
  totalCobertura,
  usd,
} from "./model/selectors"
export type {
  BucketId,
  BucketToken,
  CifraCaja,
  Cobertura,
  Confianza,
  Escenario,
  EstadoDetector,
  FilaPortafolio,
  MarcaDeFuga,
  ParidadCosto,
  PuntoMejora,
  ResumenTelemetria,
  SaludTelemetria,
  Tendencia,
  Ventana,
  VersionCatalogo,
} from "./model/types"
export { BarraCobertura, type BarraCoberturaProps } from "./ui/barra-cobertura"
export { CifraUsd, type CifraUsdProps } from "./ui/cifra-usd"
export { MarcaConfianza, type MarcaConfianzaProps } from "./ui/marca-confianza"
export { Sparkline, type SparklineProps } from "./ui/sparkline"
