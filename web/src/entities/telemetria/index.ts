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
  MOTIVO_SIN_DATO,
  marcaPrincipal,
  motivoSinDato,
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
// Fixtures del dogfood, expuestas por la Public API igual que `devFullCycle` y `entradasDemo`
// (steiger `fsd/no-public-api-sidestep` prohíbe el deep-import a `testing/`). Van en dos bloques
// rotulados MEDIDO / ILUSTRATIVO: cuando un número no lo midió nadie, el archivo lo dice.
export {
  BUCKETS_ILUSTRATIVOS,
  BUCKETS_MEDIDOS,
  CAJAS_ILUSTRATIVAS,
  CAMPOS_PERSISTIDOS_HOOK,
  CATALOGO_MEDIDO,
  COBERTURA_ILUSTRATIVA,
  DETECTORES_MVP,
  DETECTORES_NO_MEDIDOS,
  FILA_SIN_DATO,
  FILAS_PORTAFOLIO_ILUSTRATIVAS,
  HUELLAS_MEDIDAS,
  PARIDAD_COINCIDEN,
  PARIDAD_DIVERGEN,
  PII_QUE_LLEGA_Y_SE_DESCARTA,
  PUNTO_B1,
  PUNTO_B3,
  PUNTO_P1,
  RESUMEN_ILUSTRATIVO,
  TURNO_MEDIDO,
} from "./testing/telemetria"
export { BarraCobertura, type BarraCoberturaProps } from "./ui/barra-cobertura"
export { CifraUsd, type CifraUsdProps } from "./ui/cifra-usd"
export { MarcaConfianza, type MarcaConfianzaProps } from "./ui/marca-confianza"
export { Sparkline, type SparklineProps } from "./ui/sparkline"
