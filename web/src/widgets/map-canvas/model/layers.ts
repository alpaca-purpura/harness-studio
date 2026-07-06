// The four conmutable map layers (UX.md S2). MVP renders only Estructura; Tokens/Desempeño/
// Proceso need JSONL telemetry (the real indexer, Hito 3) and are disabled — honest, not hidden.
export type Capa = "estructura" | "tokens" | "perf" | "proceso"

export interface LayerDef {
  id: Capa
  label: string
  disabled?: boolean
}

export const LAYERS: LayerDef[] = [
  { id: "estructura", label: "Estructura" },
  { id: "tokens", label: "Tokens", disabled: true },
  { id: "perf", label: "Desempeño", disabled: true },
  { id: "proceso", label: "Proceso", disabled: true },
]
