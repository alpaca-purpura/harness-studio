import type { Cobertura, Confianza, Tendencia } from "./types"

// Selectores puros de la entidad `telemetria`. Presentan; no calculan dominio.
//
// **Este archivo es la ÚNICA fuente del copy de confianza y del formato de dinero** (D18 candado
// 2 · RF-281). El nodo del canvas no duplica el string: lo recibe por prop desde el widget, que
// lo lee de acá. Lo que se duplica es el TIPO (una unión de 4 literales), que es lo que `tsc`
// sí puede cuidar. La igualdad literal entre las dos superficies la ata la story
// `CopyConfianzaEsUnaSola` (map-canvas.stories.tsx, D18 candado 1).

/** El literal que se muestra por cada caso de confianza. `exacta` NO tiene: la ausencia de marca
 *  ES la señal (design §5.3). Ninguno dice «aproximada» — taparía tres cosas distintas bajo una. */
export const ETIQUETA_CONFIANZA: Readonly<Record<Confianza, string | null>> = {
  exacta: null,
  "por-hash": "por huella",
  "por-proceso": "por proceso",
  "sin-dato": "sin dato atribuible",
}

/** El `title` largo de cada caso. Tres strings DISTINTOS (design §7.3): la diferencia entre
 *  «por huella» y «por proceso» es material — una identifica, la otra puede estar mezclando. */
export const TITULO_CONFIANZA: Readonly<Record<Confianza, string | null>> = {
  exacta: null,
  "por-hash":
    "Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA.",
  "por-proceso":
    "Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla.",
  "sin-dato":
    "Estas corridas no se pudieron atribuir a ninguna caja. No suman al total: quedan contadas en la cobertura.",
}

/** El tono de la marca. Los tres visibles comparten tono a propósito: **el texto es el que
 *  distingue** (RF-280), nunca el color. `--foreground` sobre `--warn-soft` (D21). */
export const TONO_CONFIANZA: Readonly<Record<Confianza, "ninguno" | "duda">> = {
  exacta: "ninguno",
  "por-hash": "duda",
  "por-proceso": "duda",
  "sin-dato": "duda",
}

export function etiquetaConfianza(c: Confianza): string | null {
  return ETIQUETA_CONFIANZA[c]
}

export function tituloConfianza(c: Confianza): string | null {
  return TITULO_CONFIANZA[c]
}

export function tonoConfianza(c: Confianza): "ninguno" | "duda" {
  return TONO_CONFIANZA[c]
}

// ── Dinero (RF-281) ────────────────────────────────────────────────────────────────────────
//
// Un solo formateador para las CINCO superficies. Reglas, todas asertadas:
//  · coma decimal y separador de millar de espacio fino — jamás el formato EN (`1,234.56`);
//  · dos decimales por default, pero **`0,004` no se redondea a `0,00`**: un monto que existe
//    y se muestra como cero es una mentira barata, y es la que este paquete existe para no decir;
//  · `null` ⇒ `null`, y la superficie dice `sin dato`. Nunca `0,00`, nunca `—` a secas.

const ESPACIO_FINO = " "

/** Agrupa de a tres con espacio fino, sin tocar la parte decimal. */
function agrupar(entero: string): string {
  return entero.replace(/\B(?=(\d{3})+(?!\d))/g, ESPACIO_FINO)
}

/**
 * `usd(micros)` → `"1 234,56"` (sin el prefijo `USD`, que la UI pinta aparte para poder
 * atenuarlo). `null`/`undefined` ⇒ `null`: la ausencia la resuelve el componente, no el string.
 */
export function usd(micros: number | null | undefined): string | null {
  if (micros === null || micros === undefined || Number.isNaN(micros)) return null
  const dolares = micros / 1_000_000
  // Sube decimales hasta que el número deje de renderizarse como cero (tope: la precisión real
  // del dato, que son micros = 6 decimales). Un monto no nulo NUNCA sale «0,00».
  let decimales = 2
  while (decimales < 6 && dolares !== 0 && Number(dolares.toFixed(decimales)) === 0) decimales += 1
  const fijo = Math.abs(dolares).toFixed(decimales)
  const [entero = "0", dec = ""] = fijo.split(".")
  const signo = dolares < 0 ? "−" : ""
  return `${signo}${agrupar(entero)},${dec}`
}

/** Porcentaje entero, sin decimales (design §3.2). `null` ⇒ `null`, jamás `0 %`. */
export function pct(parte: number | null | undefined): string | null {
  if (parte === null || parte === undefined || Number.isNaN(parte)) return null
  return `${Math.round(parte * 100)} %`
}

/** Enteros con separador de millar de espacio fino (los tokens del inspector, design §3.4). */
export function entero(n: number | null | undefined): string | null {
  if (n === null || n === undefined || Number.isNaN(n)) return null
  return agrupar(String(Math.trunc(Math.abs(n))))
}

// ── Cobertura ──────────────────────────────────────────────────────────────────────────────

export interface SegmentoCobertura {
  id: "exacta" | "por-hash" | "por-proceso" | "sin-dato"
  /** Plural, como se lee en el texto de al lado: `12 exactas · 3 por huella · …`. */
  etiqueta: string
  n: number
  /** Cómo se nombra en el `aria-label`: `12 corridas exactas, 3 por huella, …`. */
  aria: string
}

const SEGMENTOS: readonly { id: SegmentoCobertura["id"]; etiqueta: string; aria: string }[] = [
  { id: "exacta", etiqueta: "exactas", aria: "corridas exactas" },
  { id: "por-hash", etiqueta: "por huella", aria: "por huella" },
  { id: "por-proceso", etiqueta: "por proceso", aria: "por proceso" },
  { id: "sin-dato", etiqueta: "sin dato", aria: "sin dato" },
]

/** Total de corridas de la barra: la suma de los cuatro cubos, que es su propio denominador. */
export function totalCobertura(c: Cobertura): number {
  return c.exacta + c.por_hash + c.por_proceso + c.sin_dato
}

/**
 * Los segmentos a dibujar, en orden de calidad descendente. **Una categoría en cero no se
 * dibuja ni se nombra** (design §2.2): un segmento de ancho 0 con su etiqueta al lado sería
 * ruido que el ojo cuenta igual.
 */
export function segmentosCobertura(c: Cobertura): SegmentoCobertura[] {
  const n: Record<SegmentoCobertura["id"], number> = {
    exacta: c.exacta,
    "por-hash": c.por_hash,
    "por-proceso": c.por_proceso,
    "sin-dato": c.sin_dato,
  }
  return SEGMENTOS.filter((s) => n[s.id] > 0).map((s) => ({ ...s, n: n[s.id] }))
}

// ── Tendencia ──────────────────────────────────────────────────────────────────────────────

/**
 * Dirección de una serie. `null` con menos de dos puntos: con un punto no hay tendencia, y
 * dibujar una línea plana diría «estable», que es una afirmación que nadie midió.
 */
export function direccionTendencia(serie: readonly number[] | undefined): Tendencia | null {
  if (!serie || serie.length < 2) return null
  const primero = serie[0] as number
  const ultimo = serie[serie.length - 1] as number
  if (primero === 0) return ultimo === 0 ? "estable" : "alza"
  const delta = (ultimo - primero) / Math.abs(primero)
  if (delta > 0.05) return "alza"
  if (delta < -0.05) return "baja"
  return "estable"
}

const TENDENCIA_TEXTO: Readonly<Record<Tendencia, string>> = {
  alza: "en alza",
  estable: "estable",
  baja: "a la baja",
}

/** `aria-label` del sparkline (design §7.6). Lleva el conteo: 5 corridas no es «siempre». */
export function ariaTendencia(t: Tendencia, corridas: number): string {
  return `Tendencia ${TENDENCIA_TEXTO[t]} en las últimas ${corridas} corridas.`
}

// ── Detectores ─────────────────────────────────────────────────────────────────────────────

/** Nombre corto de cada detector (design §7.3). Es lo que hace que una marca de fuga NOMBRE
 *  al detector en vez de ser un ⚠ que obliga a adivinar. */
export const NOMBRE_DETECTOR: Readonly<Record<string, string>> = {
  "b1-rewarm-por-ttl": "re-warm TTL",
  "p1-caja-que-consume-y-se-rechaza": "rechazo en gate",
  "b3-cambio-de-modelo-invalida-cache": "modelo cambiado",
  "b6-sesion-abandonada": "sesión abandonada",
  "b2-costo-de-la-rotacion": "rotación cara",
  "b4-gasto-por-arnes-empresa-puesto": "concentración de gasto",
}

/** El nombre corto, o el id crudo si el detector es desconocido — **jamás vacío**: un detector
 *  sin nombre es exactamente el ⚠ genérico que el contrato prohíbe. */
export function etiquetaDetector(id: string): string {
  return NOMBRE_DETECTOR[id] ?? id
}

/**
 * La marca a pintar en el nodo: la de MAYOR ahorro. Cero o una, nunca dos (RF-241) — dos ⚠ en
 * un nodo de 232 px no se leen, se acumulan.
 *
 * Es `marcas[0]` y no un re-orden acá porque **el wire ya viene ordenado por ahorro
 * descendente**: `telemetria_mejoras.go:125` ordena los puntos por `DiferenciaMicros` desc y
 * `telemetria_service.go:269` construye las marcas en ese mismo recorrido. Reordenar en el FE
 * por `grave` cambiaría el criterio a espaldas del dominio, que es el que tiene el número.
 */
export function marcaPrincipal<T>(marcas: readonly T[] | undefined): T | null {
  if (!marcas || marcas.length === 0) return null
  return marcas[0] as T
}
