import type { Banda } from "@/entities/arnes"

// The Base region iterates SUPPORT_BANDS in order, rendering only the bands with ≥1 node
// (mockup:218-238, RF-13/RF-44). The `base` band is special — it renders the collapsible
// rules + knowledge subbands (BaseBand); the rest are simple blocks with an activation chip.
//
// PROPUESTA (design §8.3 · spec §10 decision 2): the activation axis (siempre/condicional/
// demanda/leído/dormida) is the honesty of a component, NOT its banda — a principle to
// ratify. Here it is fixture metadata, labeled PROPUESTA in the help panel.

export type ActKind = "siempre" | "condicional" | "demanda" | "leido" | "dormida" | "mixta"

export interface SupportBand {
  id: Banda
  label: string
  act: ActKind
}

export const SUPPORT_BANDS: readonly SupportBand[] = [
  { id: "base", label: "Base · reglas y knowledge", act: "mixta" },
  { id: "libreria-expertos", label: "Librería de expertos", act: "demanda" },
  { id: "meta-harness", label: "Meta-harness · config del arnés", act: "demanda" },
  { id: "terceros", label: "Terceros", act: "demanda" },
  { id: "marcas-dormidas", label: "Marcas dormidas", act: "dormida" },
]

// ACT — the activation chip label + tone token per kind (design §8.3, mockup:231-238).
export const ACT: Record<ActKind, { label: string; tone: string }> = {
  siempre: { label: "siempre en contexto", tone: "var(--crit)" },
  condicional: { label: "carga condicional", tone: "var(--warn)" },
  demanda: { label: "bajo demanda", tone: "var(--muted-foreground)" },
  leido: { label: "leído por skills", tone: "var(--c-knowledge)" },
  dormida: { label: "dormida · 0 corridas", tone: "var(--muted-foreground)" },
  mixta: { label: "", tone: "var(--muted-foreground)" },
}
