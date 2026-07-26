import {
  type CifraCaja,
  etiquetaConfianza,
  marcaPrincipal,
  pct,
  tituloConfianza,
  usd,
} from "@/entities/telemetria"
import type { MejoraNodo } from "../ui/lane"

/**
 * `CifraCaja → props primitivas` — **la traducción que D18 exige que viva acá**.
 *
 * `entities/arnes` no importa `entities/telemetria` (nunca, por ningún escape), así que el nodo
 * recibe strings y números. El copy de confianza NO se duplica: sale de `etiquetaConfianza`/
 * `tituloConfianza`, que son la única fuente, y la story `CopyConfianzaEsUnaSola` asserta que
 * el chip del nodo y `<MarcaConfianza>` dicen exactamente lo mismo.
 */
export function propsDeMejora(c: CifraCaja, motivoSinDato: string | undefined): MejoraNodo {
  if (!c.atribuible || c.costo_micros === null) {
    return { motivoSinDato: c.motivo ?? motivoSinDato ?? "sin corridas en esta ventana" }
  }
  const marca = marcaPrincipal(c.marcas)
  const porcentaje = c.parte === undefined ? undefined : Number.parseInt(pct(c.parte) ?? "", 10)
  return {
    cifraUsd: usd(c.costo_micros) ?? undefined,
    participacionPct: Number.isNaN(porcentaje) ? undefined : porcentaje,
    confianza: c.confianza,
    etiquetaConfianza: etiquetaConfianza(c.confianza) ?? undefined,
    tituloConfianza: tituloConfianza(c.confianza) ?? undefined,
    marcaFuga: marca?.nombre,
    marcaFugaGrave: marca?.grave,
  }
}
