// Tipos de la entidad `telemetria` — espejo EXACTO del wire de Go
// (`internal/domain/telemetria_vistas.go` + `telemetria_deteccion.go`), en snake_case porque el
// wire es snake_case y traducirlo acá sería inventar una segunda verdad.
//
// Regla de contrato que gobierna el archivo entero, heredada del backend y enforced en la UI por
// las stories: **`null` y `0` significan cosas distintas.** Todo lo que puede faltar es
// `number | null` explícito, jamás un `number` con 0 de relleno. Convertir un `null` en `0` es
// el pass fabricado que el paquete entero existe para no cometer
// (boundary `no-aplica-no-es-cero`).
//
// `entities` PRESENTA; no calcula (design.md §1.3). El join, los detectores y el contrafactual
// llegan resueltos del dominio Go: cruzarlos en el FE sería cross-import entity↔entity y
// `steiger fsd/no-cross-imports` lo rompe.

/** Los cuatro casos de `atribucion_confianza` (RF-242). Se distinguen por TEXTO, no por tono. */
export type Confianza = "exacta" | "por-hash" | "por-proceso" | "sin-dato"

/** La ventana temporal de la capa. Es estado de la página; el wire recibe `desde`/`hasta`. */
export type Ventana = "7d" | "30d" | "todo"

/** Cómo corrió el arnés. S2 se parte en dos desde el ANEXO H9: el bloque `env` decide. */
export type Escenario = "s1" | "s2-instrumentado" | "s2-degradado"

/** Dirección de la tendencia. `null` = no hay serie suficiente para afirmar nada. */
export type Tendencia = "alza" | "estable" | "baja"

/** Los seis buckets de token del superset multi-proveedor (`domain.Tokens`). */
export type BucketId =
  | "entrada"
  | "salida"
  | "cache_lectura"
  | "cache_escritura_5m"
  | "cache_escritura_1h"
  | "cache_escritura_sin_tier"
  | "razonamiento"

/**
 * Cobertura de la atribución. `esperados`/`no_llegaron` viajan explícitos y pueden ser `null`:
 * en S2 no hay denominador independiente, y obligarle al FE a restar sería obligarlo a inventar.
 */
export interface Cobertura {
  esperados: number | null
  exacta: number
  por_hash: number
  por_proceso: number
  sin_dato: number
  no_llegaron: number | null
}

/** Un detector nombrado sobre una caja. NUNCA un ⚠ genérico (`domain.MarcaDeFuga`). */
export interface MarcaDeFuga {
  detector: string
  nombre: string
  grave: boolean
}

/**
 * El gasto de UNA caja en la ventana — `domain.GastoCaja`. Las cajas sin dato viajan igual, con
 * `atribuible: false` + `motivo`: omitirlas obligaría al FE a inventar por qué faltan.
 *
 * Es el tipo que el widget `map-canvas` compone en props PRIMITIVAS para `ArnesNode` (D18):
 * `entities/arnes` no importa `entities/telemetria`, nunca, por ningún escape.
 */
export interface CifraCaja {
  caja_id: string
  nombre: string
  atribuible: boolean
  motivo?: string | undefined
  /** `null`, no 0, cuando no es atribuible. */
  costo_micros: number | null
  /** Participación en el total del arnés, en 0..1. Ausente si el total es 0. */
  parte?: number | undefined
  confianza: Confianza
  corridas: number
  /** Distingue «no se pudo asignar a esta caja» de «se asignó y no hubo dinero». */
  sin_atribucion?: boolean | undefined
  marcas?: readonly MarcaDeFuga[] | undefined
}

/** Una fila de la tabla de buckets del inspector. `tokens: null` = el runtime no tiene el concepto. */
export interface BucketToken {
  id: BucketId
  etiqueta: string
  /** `null` = no aplica en este runtime. `0` = el runtime tiene el concepto y midió cero. */
  tokens: number | null
  costo_micros: number | null
}

/** `domain.ParidadCosto` — los dos costos y su divergencia. La UI muestra los dos y NO elige. */
export interface ParidadCosto {
  reportado_micros: number | null
  calculado_micros: number | null
  /** `null` cuando falta alguno: sin dos números no hay divergencia, y un 0 diría «coinciden». */
  divergencia_pct: number | null
  completo: boolean
  sin_tarifa?: readonly string[] | undefined
  catalogo_version?: string | undefined
  /** El estado real de hoy: el catálogo embebido todavía no se construyó. */
  catalogo_sin_construir?: boolean | undefined
}

/** `domain.VersionCatalogoPrecios` — con qué precios se costeó. */
export interface VersionCatalogo {
  version: string
  rev?: string | undefined
  modelos?: number | undefined
  /** `null` = NUNCA se refrescó. No es «hace 0 h». */
  refrescado: string | null
}

/**
 * `domain.PuntoDeMejora` — la unidad del entregable: un número, su contrafactual, su umbral
 * citado, su sesgo declarado EN CONTRA y UN fix. Sin las cinco cosas no hay tarjeta (regla A4).
 */
export interface PuntoMejora {
  id: string
  detector: string
  score_version: number
  titulo: string
  lede: string
  /** Ausente cuando el detector agrega la ventana entera y no una caja (B1 · B2 · B3 · B6). */
  caja_id?: string | undefined
  gasto_micros: number
  parte_del_total: number
  /** Qué habría costado el mundo alternativo, en micros. La prosa de abajo es su lectura. */
  contrafactual_micros: number
  /**
   * El mundo alternativo con las MISMAS corridas (A2), redactado con su UNIDAD explícita
   * («por corrida» / «en la ventana»): RF-249 lo exige porque «USD 0,53» sobre una caja de 14
   * corridas se leyó de dos maneras distintas en la iteración 1 del mockup.
   *
   * **`null` ⇒ la tarjeta NO EXISTE** (regla A4). No se pinta degradada: la lista la filtra y
   * el inspector la lista como «sin fix propuesto». Una tarjeta sin contrafactual es un
   * reproche, no una recomendación.
   *
   * **La arma el dominio Go** (`internal/domain/telemetria_prosa.go#PuntoDeMejora.Redactar`),
   * jamás el FE — D25, y antes de D25 lo decía `design.md` §1.3. Contra el wire real llega
   * siempre presente: el backend no publica el punto cuando no la puede armar. El `null` sigue
   * tipado porque las fixtures storian ese caso y el filtro de la lista es la regla.
   */
  contrafactual: string | null
  /** El ahorro, en micros. Es la clave de orden de la lista: lo de más plata primero. */
  diferencia_micros: number
  /** El desglose de la confianza cuando NO es exacta (design §7.4). */
  confianza_detalle?: string | undefined
  /** La desigualdad algebraica CITADA (A1). Ausente ⇒ se pinta `patron` en su lugar. */
  umbral?: string | undefined
  /** Lo que reemplaza al umbral cuando el detector no se decide por uno (P1). */
  patron?: string | undefined
  /** Se declara y va EN CONTRA de la recomendación (A3). `null` ⇒ la fila lo DICE igual. */
  sesgo: string | null
  /** «subestima» | «sobreestima». Va en negrita dentro del sesgo: es la palabra que decide si
   *  el número es un piso o un techo. Vacío solo cuando `sesgo` es `null` (RF-252). */
  direccion_sesgo: string
  fix: string
  /** El ajuste concreto, en `<code>`. */
  fix_codigo?: string | undefined
  confianza: Confianza
  corridas_usadas: number
  corridas_totales: number
  grave: boolean
  /** El detector no aplica fuera de S1 (hoy solo B1) — chip `solo con telemetría de ArnesIA`. */
  solo_s1?: boolean | undefined
  /** La fórmula resuelta que muestra «ver el cálculo» (H-4). */
  calculo?: string | undefined
}

/**
 * `domain.EstadoDetector` — un detector con su veredicto de aplicabilidad. `motivo` es
 * OBLIGATORIO cuando `aplica` es false: un detector apagado sin razón es un gap escondido.
 */
export interface RespuestaMejoras {
  puntos: readonly PuntoMejora[]
  no_aplican: readonly EstadoDetector[]
  no_medidos: readonly EstadoDetector[]
  escenario: Escenario
  /** Cuántos puntos se ocultaron porque el operador los descartó (D26.4). Viaja para que la
   *  lista pueda decirlo: una lista más corta sin explicación se lee como «no hay más». */
  descartados: number
}

export interface EstadoDetector {
  detector: string
  nombre: string
  aplica: boolean
  motivo?: string | undefined
  cobertura_parcial?: boolean | undefined
  hallazgos: number
  /** Encontró algo cotizable pero NO tiene fix propuesto: la contraparte de A4. No se esconde
   *  —el inspector lo lista—, pero tampoco genera tarjeta. */
  sin_fix?: boolean | undefined
}

/** `domain.ResumenTelemetria` — lo que alimenta la franja del Mapa. */
export interface ResumenTelemetria {
  desde: string
  hasta: string
  estimado: boolean
  costo_reportado_micros: number | null
  costo_calculado_micros: number | null
  costo_completo: boolean | null
  sin_tarifa?: readonly string[] | undefined
  divergencia_pct: number | null
  divergencia_sospechosa: boolean
  corridas: number
  sesiones: number
  turnos: number
  cajas: number
  escenario: Escenario
  confianza: Confianza
  cobertura: Cobertura
  catalogo: VersionCatalogo
  /**
   * La última medición del arnés **ignorando la ventana** (D26.4, estado 1b). `null` = nunca
   * hubo una, y ese sí es «este arnés nunca corrió». Sin este campo, «0 corridas en 7 días» y
   * «nunca corrió» se decían igual — y sobre un arnés con historial la segunda es falsa.
   */
  ultima_corrida: string | null
  /** Con qué corre, del dato real de la ventana. Vacío = ningún evento lo dijo. */
  runtime?: string | undefined
  /** Si sabemos medir ese runtime. `false` con `runtime` presente = «todavía no lo medimos»,
   *  que se DICE en vez de mostrar un cero. */
  runtime_soportado: boolean
}

/** `domain.FilaPortafolio` — una fila = arnés × instalación (D20). */
export interface FilaPortafolio {
  arnes_id: string
  clave: string
  nombre: string
  instalacion_id: string
  /** `null` = el arnés no declara `rol`. La UI dice «puesto sin declarar» — el caso NORMAL hoy. */
  puesto: string | null
  empresas?: readonly string[] | undefined
  corridas: number
  costo_micros: number | null
  /** `null` para el que nunca corrió. JAMÁS 0. */
  costo_por_corrida: number | null
  confianza: Confianza
  /** `""` cuando el backend no la calculó (auditoría M10: hoy nunca se calcula). */
  tendencia?: Tendencia | "" | undefined
  /**
   * La serie de las últimas corridas que dibuja el sparkline.
   *
   * ⚠️ HOY EL WIRE NO LA MANDA: `domain.FilaPortafolio` no tiene ningún campo de serie
   * (auditoría §UX ítem 6 · M10). Es opcional a propósito y su ausencia se DICE en pantalla
   * («pocas corridas para una tendencia»), jamás se dibuja una línea inventada.
   */
  serie?: readonly number[] | undefined
  /** Un 0 acá SÍ es dato: «se midió y no se encontró nada» ≠ `costo_micros: null`. */
  puntos_de_mejora: number
  /**
   * El punto de mayor ahorro, ya resuelto, para el chip de la fila. Puede faltar **aunque
   * `puntos_de_mejora > 0`**: la UI dice cuántos hay, no pone un ✓ (que afirmaría lo contrario).
   */
  punto?:
    | { nombre: string; monto_micros: number; unidad: "corrida" | "ventana"; grave: boolean }
    | undefined
}

/** `domain.SaludTelemetria`, recortado a lo que la superficie necesita. */
export interface SaludTelemetria {
  retencion_dias: number
  /** El número NO está firmado (J-6 · parada P2): la UI lo rotula como propuesto. */
  forward: boolean
  forward_destino?: string | undefined
  catalogo: VersionCatalogo
  ultima_recepcion: string | null
  almacen_disponible: boolean
  almacen_motivo?: string | undefined
}
