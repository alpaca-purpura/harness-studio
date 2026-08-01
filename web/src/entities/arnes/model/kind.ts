import type { GlyphShape } from "@/shared/canvas"
import type { Clase } from "./types"

// KIND maps each L0 Clase to its map visual: a token color (never a raw hex), a glyph shape,
// a mono char and the ES label. Color + shape are a redundant channel (color-blind safety).
//
// Las 10 clases tienen color propio (`--c-*`). El token `kind` emite los 10 tras `tokens:build`
// (base.tokens.json:165-211 → theme.css); las 5 config (command/plugin/settings/output-style/
// statusline) comparten forma `rounded` — la sub-distinción la carga el char + el color.
export interface KindVisual {
  color: string // token reference, e.g. "var(--c-skill)"
  shape: GlyphShape
  char: string
  label: string
}

export const KIND: Record<Clase, KindVisual> = {
  skill: { color: "var(--c-skill)", shape: "square", char: "S", label: "Skill" },
  subagent: { color: "var(--c-agent)", shape: "circle", char: "A", label: "Subagente" },
  hook: { color: "var(--c-hook)", shape: "diamond", char: "H", label: "Hook" },
  rule: { color: "var(--c-rule)", shape: "shield", char: "R", label: "Regla" },
  mcp: { color: "var(--c-mcp)", shape: "hexagon", char: "M", label: "MCP" },
  command: { color: "var(--c-command)", shape: "rounded", char: "/", label: "Comando" },
  plugin: { color: "var(--c-plugin)", shape: "rounded", char: "P", label: "Plugin" },
  settings: { color: "var(--c-settings)", shape: "rounded", char: "⚙", label: "Ajustes" },
  "output-style": {
    color: "var(--c-output-style)",
    shape: "rounded",
    char: "◐",
    label: "Output style",
  },
  statusline: {
    color: "var(--c-statusline)",
    shape: "rounded",
    char: "▭",
    label: "Statusline",
  },
  // Marcador de reconciliación (D-c, nomenclatura §4.5) — no es primitiva: tono warn, nunca
  // un color --c-* de clase. Sin esta entrada un nodo no-reconocido tumbaba el canvas entero
  // al ErrorBoundary (KIND[clase] undefined), lo contrario de «visible con warn».
  "no-reconocido": {
    color: "var(--warn)",
    shape: "rounded",
    char: "?",
    label: "No reconocido",
  },
}

// kindFor — lookup TOTAL sobre KIND (deuda F, §4.5): una clase que el wire trae fuera del
// enum (sin pasar por el marcador del loader) degrada al visual no-reconocido NOMBRANDO el
// valor, en vez de tumbar el lienzo entero al ErrorBoundary (`KIND[clase].color` undefined).
export function kindFor(clase: string): KindVisual {
  const k = (KIND as Record<string, KindVisual>)[clase]
  if (k) return k
  return { ...KIND["no-reconocido"], label: `No reconocido: ${clase}` }
}
