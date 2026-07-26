import type { BucketToken, EstadoDetector, ParidadCosto, Ventana } from "@/entities/telemetria"

// La forma de lectura del detalle de UNA caja y su mapeo a la tabla de buckets. Vive en `model/`
// —no en la página— porque es derivación, y toda derivación de esta capa que vivió en el
// composition-root terminó siendo uno de los defectos de composición del paquete.

export const ETIQUETA_VENTANA: Readonly<Record<Ventana, string>> = {
  "7d": "7 días",
  "30d": "30 días",
  todo: "todo",
}

/** `domain.DetalleCaja` recortado a lo que la 4ª tab necesita. Los buckets llegan como punteros:
 *  un campo AUSENTE en el JSON es «no aplica», y eso se conserva como `null` (RF-260). */
export interface DetalleCajaWire {
  turnos_totales: number
  tokens?: Partial<Record<string, number>> | undefined
  paridad?: ParidadCosto | undefined
  detectores?: EstadoDetector[] | undefined
}

/** Los seis buckets, en orden fijo. **Un bucket ausente en el wire viaja `null`, no 0**: el
 *  `omitempty` de Go significa «este runtime no tiene el concepto», y un 0 sería mentira. */
export const BUCKETS: readonly { id: BucketToken["id"]; etiqueta: string; campo: string }[] = [
  { id: "entrada", etiqueta: "entrada", campo: "entrada" },
  { id: "salida", etiqueta: "salida", campo: "salida" },
  { id: "cache_lectura", etiqueta: "cache · lectura", campo: "cache_lectura" },
  { id: "cache_escritura_5m", etiqueta: "cache · escritura 5 m", campo: "cache_escritura_5m" },
  { id: "cache_escritura_1h", etiqueta: "cache · escritura 1 h", campo: "cache_escritura_1h" },
  { id: "razonamiento", etiqueta: "razonamiento", campo: "razonamiento" },
]

export function bucketsDe(d: DetalleCajaWire | null): BucketToken[] {
  return BUCKETS.map((b) => ({
    id: b.id,
    etiqueta: b.etiqueta,
    tokens: d?.tokens?.[b.campo] ?? null,
    // ⚠️ El wire NO manda el costo por bucket: `domain.DetalleCaja` trae `Tokens` y `Paridad`,
    // no un desglose de dinero por bucket. La columna USD queda `null` —«no aplica»— hasta que
    // el backend lo mande. Inventarla acá sería costear en el FE, que es justo lo que
    // `design.md` §1.3 prohíbe («entities presenta; no calcula»).
    costo_micros: null,
  }))
}
