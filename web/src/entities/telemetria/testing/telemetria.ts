import type {
  BucketToken,
  CifraCaja,
  Cobertura,
  EstadoDetector,
  FilaPortafolio,
  ParidadCosto,
  PuntoMejora,
  ResumenTelemetria,
  VersionCatalogo,
} from "../model/types"

// Fixtures de telemetría, en DOS bloques separados y rotulados (plan-storybook §3).
//
// Regla dura del repo: datos reales de dogfood, jamás inventados. Cuando un número no lo midió
// nadie, se dice — y acá se dice arriba de cada bloque, no en una nota al pie.

// ══════════════════════════════════════════════════════════════════════════════════════════
// ── MEDIDO ── sale de docs/product/stories/2026-07-24-telemetria-embebida-otel/
//              verificacion-2026-07-26/evidencia/, byte por byte.
// ══════════════════════════════════════════════════════════════════════════════════════════
//
// ⚠️ DISCREPANCIA DECLARADA (no resuelta a mano alzada): `plan-storybook.md` §3.2 cita el split
// `input 10 · cache_read 21695 · cache_creation 4099 · output 35`, que es el de `INFORME.md`
// V2 (líneas 61-63) — OTRA de las tres corridas. El archivo golden versionado
// (`evidencia/result-envelope.json`) y el `api_request` de `evidencia/logs-run1.json` traen
// `input 10 · cache_read 17536 · cache_creation 8257 · output 39 · cost_usd_micros 18473`, y
// coinciden ENTRE SÍ. Gana el archivo, que es el golden file citado por la regla 4 del plan.
// Los dos hechos que el plan necesita de esta corrida sobreviven intactos:
//   · `ephemeral_5m_input_tokens: 0` — el CERO LEGÍTIMO de RF-260, medido, no inventado;
//   · `razonamiento` NUNCA apareció en los 57 log records — el «no aplica» también es real.

/** El turno medido: `evidencia/logs-run1.json`, log record `api_request`. */
export const TURNO_MEDIDO = {
  arnes: "vitalia",
  caja: "paso-3",
  instalacion: "home-local",
  sesion_id: "89c8bdde-913f-407f-8959-51fffb5cc27c",
  prompt_id: "8e957894-aefe-4be1-a3b6-1fb4bb31b190",
  modelo: "claude-haiku-4-5-20251001",
  input_tokens: 10,
  output_tokens: 39,
  cache_read_tokens: 17536,
  cache_creation_tokens: 8257,
  cost_usd_micros: 18473,
  duration_ms: 1772,
  speed: "normal",
  query_source: "sdk",
} as const

/**
 * Los buckets del turno medido, con la FORMA que el inspector pinta.
 *
 * `cache_escritura_1h = 8257` y `cache_escritura_5m = 0` salen de
 * `evidencia/result-envelope.json#usage.cache_creation`. El **0 de 5 m es real** (el runtime
 * tiene el concepto y midió cero) y el `null` de `razonamiento` **también** (nunca llegó en
 * ningún record): son las dos ausencias distintas que RF-260 exige distinguir.
 */
export const BUCKETS_MEDIDOS: readonly BucketToken[] = [
  { id: "entrada", etiqueta: "entrada", tokens: 10, costo_micros: 8 },
  { id: "salida", etiqueta: "salida", tokens: 39, costo_micros: 156 },
  { id: "cache_lectura", etiqueta: "cache · lectura", tokens: 17536, costo_micros: 1403 },
  { id: "cache_escritura_5m", etiqueta: "cache · escritura 5 m", tokens: 0, costo_micros: 0 },
  {
    id: "cache_escritura_1h",
    etiqueta: "cache · escritura 1 h",
    tokens: 8257,
    costo_micros: 16906,
  },
  { id: "razonamiento", etiqueta: "razonamiento", tokens: null, costo_micros: null },
]

/** Las tres huellas de plugin observadas (`INFORME.md` V3). Cualquier otra ⇒ `sin-dato`. */
export const HUELLAS_MEDIDAS = {
  third_party_a: "083a6411f71e6967",
  third_party_b: "29e6a51f5b6dca52",
  oficial_rust_analyzer: "5de2da3204573e62",
} as const

/**
 * La PII que llega en cada punto y se descarta antes de escribir (`INFORME.md` V6) y el
 * contenido que llega por el canal de hooks (`ANEXO-hooks.md` H4).
 */
export const PII_QUE_LLEGA_Y_SE_DESCARTA: readonly string[] = [
  "user.email",
  "user.account_uuid",
  "user.account_id",
  "organization.id",
  "user.id",
]

/**
 * La allowlist REAL del canal de hooks (`ANEXO-hooks.md` H4). Es la lista que
 * `CamposDesplegados` asserta: RF-275 exige que la UI muestre **esta** lista, no una redacción
 * aparte — si la UI la inventa, la promesa deja de ser verificable.
 */
export const CAMPOS_PERSISTIDOS_HOOK: readonly string[] = [
  "session_id",
  "prompt_id",
  "hook_event_name",
  "tool_name",
  "duration_ms",
  "cwd",
]

/** El catálogo tal como lo declara el wire hoy. `refrescado: null` = NUNCA, no «hace 0 h». */
export const CATALOGO_MEDIDO: VersionCatalogo = {
  version: "2026-07-20",
  rev: "d4a1f0c",
  modelos: 0,
  refrescado: null,
}

// ══════════════════════════════════════════════════════════════════════════════════════════
// ── ILUSTRATIVO ── NADIE MIDIÓ ESTOS NÚMEROS.
// La única telemetría real del repo son las 3 corridas de verificacion-2026-07-26/
// (USD 0,044 en total). Estos valores son CONSTANTES DE DISEÑO: design.md §7 fija el copy
// literal y el copy lleva el número adentro, así que el fixture tiene que reproducirlos para
// que los asserts de copy pasen. NO son evidencia de nada. Al primer dato de dogfood real,
// se reemplazan y se corrigen los literales de design.md §7 en el mismo commit.
// ══════════════════════════════════════════════════════════════════════════════════════════
//
// El juego de la ITERACIÓN 2 del mockup CIERRA entre sí (design.md cabecera, hueco H-8):
//   61 corridas · 58 con atribución (44 exactas + 9 huella + 5 proceso; 3 sin dato no suman)
//   12 sesiones · 4 cajas · USD 4,82 = 1,92 + 1,54 + 0,89 + 0,47 · 40+32+18+10 = 100 %

export const COBERTURA_ILUSTRATIVA: Cobertura = {
  esperados: 61,
  exacta: 44,
  por_hash: 9,
  por_proceso: 5,
  sin_dato: 3,
  no_llegaron: 0,
}

export const RESUMEN_ILUSTRATIVO: ResumenTelemetria = {
  desde: "2026-07-19T00:00:00Z",
  hasta: "2026-07-26T00:00:00Z",
  estimado: true,
  ultima_corrida: "2026-07-26T12:00:00Z",
  runtime: "claude-code",
  runtime_soportado: true,
  costo_reportado_micros: 4_820_000,
  costo_calculado_micros: 4_820_000,
  costo_completo: true,
  divergencia_pct: 0,
  divergencia_sospechosa: false,
  corridas: 61,
  sesiones: 12,
  turnos: 148,
  cajas: 4,
  escenario: "s1",
  confianza: "exacta",
  cobertura: COBERTURA_ILUSTRATIVA,
  catalogo: CATALOGO_MEDIDO,
}

/** Las 4 cajas del dogfood `dev-full-cycle`, con el reparto que cierra en 100 %. */
export const CAJAS_ILUSTRATIVAS: readonly CifraCaja[] = [
  {
    caja_id: "spec-writer",
    nombre: "escribir el spec",
    atribuible: true,
    costo_micros: 1_920_000,
    parte: 0.4,
    confianza: "exacta",
    corridas: 14,
    marcas: [{ detector: "b1-rewarm-por-ttl", nombre: "re-warm TTL", grave: false }],
  },
  {
    caja_id: "builder",
    nombre: "construir",
    atribuible: true,
    costo_micros: 1_540_000,
    parte: 0.32,
    confianza: "por-hash",
    corridas: 31,
  },
  {
    caja_id: "reviewer",
    nombre: "revisar el build",
    atribuible: true,
    costo_micros: 890_000,
    parte: 0.18,
    confianza: "exacta",
    corridas: 4,
    marcas: [
      { detector: "p1-caja-que-consume-y-se-rechaza", nombre: "rechazo en gate", grave: true },
    ],
  },
  {
    caja_id: "releaser",
    nombre: "publicar",
    atribuible: true,
    costo_micros: 470_000,
    parte: 0.1,
    confianza: "por-proceso",
    corridas: 9,
  },
]

/**
 * Los buckets de la tarjeta insignia: **la FORMA sale del bloque MEDIDO** (el `0` real de
 * `ephemeral_5m` y la ausencia real de `razonamiento`), **los MONTOS son ilustrativos** y suman
 * exactamente los `USD 1,92` que design.md §7.5 fija como copy de la caja.
 *
 * Los dos fixtures existen a propósito y no se pueden fundir: `BUCKETS_MEDIDOS` trae 18 473
 * micros repartidos en centésimas de centavo, y a dos decimales —que es como se muestra el
 * dinero (RF-281)— ninguna fila suma nada. Un assert aritmético sobre esos valores no puede
 * cerrar, y bajar la precisión del formateador para que cierre sería arreglar el termómetro.
 */
export const BUCKETS_ILUSTRATIVOS: readonly BucketToken[] = [
  { id: "entrada", etiqueta: "entrada", tokens: 12_400, costo_micros: 50_000 },
  { id: "salida", etiqueta: "salida", tokens: 8_900, costo_micros: 310_000 },
  { id: "cache_lectura", etiqueta: "cache · lectura", tokens: 214_800, costo_micros: 220_000 },
  // El CERO REAL medido: el runtime tiene el concepto y midió cero (RF-260).
  { id: "cache_escritura_5m", etiqueta: "cache · escritura 5 m", tokens: 0, costo_micros: 0 },
  {
    id: "cache_escritura_1h",
    etiqueta: "cache · escritura 1 h",
    tokens: 96_300,
    costo_micros: 1_340_000,
  },
  // La AUSENCIA REAL: `razonamiento` nunca apareció en ninguno de los 57 log records.
  { id: "razonamiento", etiqueta: "razonamiento", tokens: null, costo_micros: null },
]

export const PARIDAD_COINCIDEN: ParidadCosto = {
  reportado_micros: 1_920_000,
  calculado_micros: 1_920_000,
  divergencia_pct: 0,
  completo: true,
  catalogo_version: "2026-07-20",
}

export const PARIDAD_DIVERGEN: ParidadCosto = {
  reportado_micros: 1_920_000,
  calculado_micros: 2_100_000,
  divergencia_pct: 9.375,
  completo: true,
  catalogo_version: "2026-07-20",
}

/** La tarjeta insignia: B1 sobre la caja del spec (design §2.5 y §7.4, copy literal). */
export const PUNTO_B1: PuntoMejora = {
  id: "b1-spec-writer",
  detector: "b1-rewarm-por-ttl",
  score_version: 1,
  titulo: "La caja «escribir el spec» reescribe el cache en cada corrida",
  lede: "Gasta USD 0,84 de los 1,92 de la caja (44 %) escribiendo cache que se vence antes de volver a leerse.",
  caja_id: "spec-writer",
  gasto_micros: 840_000,
  contrafactual_micros: 310_000,
  parte_del_total: 0.44,
  contrafactual:
    "Con TTL de 1 h, las mismas 14 corridas costaban USD 0,31 → diferencia USD 0,53 en la ventana (USD 0,04 por corrida).",
  diferencia_micros: 530_000,
  umbral: "relectura 61 % > break-even (2−1,25)/(2−0,1) = 39,47 %",
  sesgo: "Asume 0 lecturas fuera de la ventana de 7 días → subestima el ahorro.",
  direccion_sesgo: "subestima",
  fix: "Fijar cache_ttl: 1h en esta caja",
  fix_codigo: "cache_ttl: 1h",
  confianza: "exacta",
  corridas_usadas: 14,
  corridas_totales: 14,
  grave: false,
  solo_s1: true,
  calculo:
    "relectura = 0,75 / 1,9 = 39,47 % de break-even · medido 61 % sobre 14 corridas en los últimos 7 días",
}

/** P1: severidad crítica, y **sin umbral** — se decide por patrón, no por una desigualdad. */
export const PUNTO_P1: PuntoMejora = {
  id: "p1-reviewer",
  detector: "p1-caja-que-consume-y-se-rechaza",
  score_version: 1,
  titulo: "La caja «revisar el build» se rechaza en el gate 3 de cada 4 veces",
  lede: "Las corridas rechazadas costaron USD 0,67 de los 0,89 de la caja (75 %): se paga el trabajo y se descarta el resultado.",
  caja_id: "reviewer",
  gasto_micros: 670_000,
  contrafactual_micros: 220_000,
  parte_del_total: 0.75,
  contrafactual:
    "Si el gate pasara a la primera, el ciclo costaba USD 0,22 en vez de 0,89 → diferencia USD 0,67 en la ventana (USD 0,17 por corrida, sobre 4).",
  diferencia_micros: 670_000,
  patron: "Los 3 rechazos citan el mismo motivo: «el veredicto no lista hallazgos».",
  sesgo: "No descuenta lo que la revisión aporta aunque rechace → sobreestima el desperdicio.",
  direccion_sesgo: "sobreestima",
  fix: "El contrato de la caja no exige hallazgos[] en el entregable",
  fix_codigo: "hallazgos[]",
  confianza: "exacta",
  corridas_usadas: 4,
  corridas_totales: 4,
  grave: true,
  calculo: "3 rechazos de 4 corridas · costo de las rechazadas USD 0,67 sobre los 0,89 de la caja",
}

/** B3, el segundo punto de la lista: menor ahorro, va después. */
export const PUNTO_B3: PuntoMejora = {
  id: "b3-builder",
  detector: "b3-cambio-de-modelo-invalida-cache",
  score_version: 1,
  titulo: "La caja «construir» cambia de modelo a mitad de corrida",
  lede: "Gasta USD 0,21 de los 1,54 de la caja (14 %) reconstruyendo un cache que el cambio de modelo invalidó.",
  caja_id: "builder",
  gasto_micros: 210_000,
  contrafactual_micros: 1_330_000,
  parte_del_total: 0.14,
  contrafactual:
    "Con un solo modelo, las mismas 31 corridas costaban USD 1,33 → diferencia USD 0,21 en la ventana (USD 0,01 por corrida).",
  diferencia_micros: 210_000,
  confianza_detalle: "no exacta · 28 de 31 exactas, 3 por huella",
  umbral: "2 modelos distintos en la misma sesión ≥ 1",
  sesgo: null,
  direccion_sesgo: "",
  fix: "Fijar el modelo de la caja en su contrato",
  fix_codigo: "model: claude-sonnet-4-6",
  confianza: "por-hash",
  corridas_usadas: 28,
  corridas_totales: 31,
  grave: false,
}

/** Los seis detectores del MVP con su estado (D16.1). */
export const DETECTORES_MVP: readonly EstadoDetector[] = [
  {
    detector: "b1-rewarm-por-ttl",
    nombre: "re-warm por TTL",
    aplica: true,
    hallazgos: 1,
  },
  {
    detector: "p1-caja-que-consume-y-se-rechaza",
    nombre: "caja que consume y se rechaza",
    aplica: true,
    hallazgos: 1,
  },
  {
    detector: "b2-costo-de-la-rotacion",
    nombre: "costo de la rotación",
    aplica: true,
    hallazgos: 0,
  },
  { detector: "b6-sesion-abandonada", nombre: "sesión abandonada", aplica: true, hallazgos: 0 },
  {
    detector: "b3-cambio-de-modelo-invalida-cache",
    nombre: "cambio de modelo invalida cache",
    aplica: true,
    hallazgos: 0,
  },
  {
    detector: "b4-gasto-por-arnes-empresa-puesto",
    nombre: "gasto por arnés × empresa × puesto",
    aplica: true,
    hallazgos: 0,
  },
]

/** Los siete que están FUERA del MVP. Dicen `no medido todavía`, jamás `sin hallazgos`. */
export const DETECTORES_NO_MEDIDOS: readonly EstadoDetector[] = [
  { detector: "b5-tool-cara", nombre: "herramienta cara", aplica: false, hallazgos: 0 },
  {
    detector: "b7-reintento-en-cadena",
    nombre: "reintento en cadena",
    aplica: false,
    hallazgos: 0,
  },
  { detector: "b8-prompt-inflado", nombre: "prompt inflado", aplica: false, hallazgos: 0 },
  { detector: "p2-handoff-perdido", nombre: "hand-off perdido", aplica: false, hallazgos: 0 },
  {
    detector: "p3-fase-sin-entregable",
    nombre: "fase sin entregable",
    aplica: false,
    hallazgos: 0,
  },
  { detector: "p4-gate-sin-criterio", nombre: "gate sin criterio", aplica: false, hallazgos: 0 },
  { detector: "p5-subagente-huerfano", nombre: "subagente huérfano", aplica: false, hallazgos: 0 },
]

/**
 * Filas del Portafolio. `acme-cli` va DOS veces —dos instalaciones en dos proyectos, el caso
 * real de `entradasDemo`— porque RF-265 se cumple **por instalación**, no por puesto (D20).
 *
 * `puesto: null` en las tres: **es el caso NORMAL de hoy**, no un borde. Ningún arnés del
 * dogfood declara `rol`, y la auditoría §UX lo confirma: la cláusula «en este puesto» se
 * renderiza «puesto sin declarar» en el 100 % de los casos.
 *
 * `serie` va `undefined` a propósito: el wire NO manda ninguna serie (auditoría M10). El
 * sparkline lo dice; no lo dibuja.
 */
export const FILAS_PORTAFOLIO_ILUSTRATIVAS: readonly FilaPortafolio[] = [
  {
    arnes_id: "vitalia",
    clave: "alpacapurpura~vitalia",
    nombre: "Vitalia",
    instalacion_id: "inst-vitalia-1",
    puesto: null,
    empresas: ["alpacapurpura"],
    corridas: 61,
    costo_micros: 4_820_000,
    costo_por_corrida: 310_000,
    confianza: "exacta",
    serie: [240_000, 260_000, 255_000, 290_000, 310_000],
    puntos_de_mejora: 1,
    punto: {
      nombre: "re-warm de cache",
      monto_micros: 120_000,
      unidad: "corrida",
      grave: false,
    },
  },
  {
    arnes_id: "acme-cli",
    clave: "acme~acme-cli",
    nombre: "ACME CLI",
    instalacion_id: "inst-acme-1",
    puesto: null,
    corridas: 22,
    costo_micros: 1_980_000,
    costo_por_corrida: 90_000,
    confianza: "por-hash",
    serie: [95_000, 92_000, 91_000, 90_000, 90_000],
    puntos_de_mejora: 0,
  },
  {
    arnes_id: "acme-cli",
    clave: "acme~acme-cli",
    nombre: "ACME CLI",
    instalacion_id: "inst-acme-2",
    puesto: null,
    corridas: 7,
    costo_micros: 420_000,
    costo_por_corrida: 60_000,
    confianza: "por-proceso",
    serie: [70_000, 66_000, 64_000, 61_000, 60_000],
    puntos_de_mejora: 0,
  },
]

/** La fila del que nunca corrió: `null` en todo lo que es dinero. **Jamás 0.** */
export const FILA_SIN_DATO: FilaPortafolio = {
  arnes_id: "harness",
  clave: "prenter~harness",
  nombre: "harness",
  instalacion_id: "inst-harness-1",
  puesto: null,
  corridas: 0,
  costo_micros: null,
  costo_por_corrida: null,
  confianza: "sin-dato",
  puntos_de_mejora: 0,
}
