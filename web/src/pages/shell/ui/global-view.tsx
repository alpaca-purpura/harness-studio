import { ComingSoon, useAppStore } from "@/shared"

const GLOBAL: Record<string, { glyph: string; title: string; note: string }> = {
  portafolio: {
    glyph: "⌂",
    title: "Portafolio",
    note: "Organigrama (arneses por empresa/puesto, «reporta a») ↔ Cuadrícula. Llega después del MVP del Mapa.",
  },
  estandar: {
    glyph: "⟳",
    title: "Estándar as code",
    note: "El árbol de conocimiento vivo (11 nodos · 122 checks) y su cadencia semanal. Próximamente.",
  },
  ajustes: {
    glyph: "⚙",
    title: "Ajustes · Marketplaces",
    note: "Marketplaces por empresa/arnés, config de proyecto y del daemon. Próximamente.",
  },
}

// GlobalView renders the daemon-wide destinations (rail foot). All are staged for now.
export function GlobalView() {
  const route = useAppStore((s) => s.view)
  const g = GLOBAL[route]
  if (!g) return <ComingSoon title="Vista" />
  return <ComingSoon glyph={g.glyph} title={g.title} note={g.note} />
}
