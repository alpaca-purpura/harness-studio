// Las cuatro capas conmutables del Mapa (UX.md S2).
//
// El segundo slot se llamaba `Tokens` y estaba apagado «hasta que exista el indexer JSONL».
// Ya no: la señal llega (verificación en vivo 2026-07-26, V1) y el slot se enciende como
// **`Mejora`** (D17.1 · RF-232). El renombre del literal de unión lo agarra `tsc`, así que no
// queda ningún consumidor huérfano.
//
// 🔴 Es una **desviación declarada de un baseline firmado** (`mockups/INDEX.md` regla 3): la
// superficie vigente decía `Tokens`. Va al gate como tal. Los cuatro slots se conservan, en el
// mismo orden — el conmutador no pierde ninguna posición.
export type Capa = "estructura" | "mejora" | "perf" | "proceso"

export interface LayerDef {
  id: Capa
  label: string
  disabled?: boolean
  /**
   * Por qué está apagado, dicho de verdad (RF-233 · RF-276).
   *
   * El `title` fijo que vivía hardcodeado en `map-bar.tsx` decía «Necesita telemetría (indexer
   * JSONL)» — **eso ya no es cierto para ninguno de los dos**: la señal de duración por request
   * y por herramienta LLEGA hoy (V1). Lo que falta es el diseño. Un motivo que miente es peor
   * que no tener motivo: manda a construir lo que ya está construido.
   */
  motivo?: string
}

export const LAYERS: LayerDef[] = [
  { id: "estructura", label: "Estructura" },
  { id: "mejora", label: "Mejora" },
  {
    id: "perf",
    label: "Desempeño",
    disabled: true,
    motivo:
      "La señal ya llega —duración por request y por herramienta—. Falta decidir qué es «desempeño» a nivel Mapa.",
  },
  {
    id: "proceso",
    label: "Proceso",
    disabled: true,
    motivo: "Entra en parte por la capa Mejora. Falta mapear todos los eventos a fases del arnés.",
  },
]
