import type { GlyphShape } from "@/shared/canvas"
import type { Clase } from "./types"

// KIND maps each L0 Clase to its map visual: a token color (never a raw hex), a glyph shape,
// a mono char and the ES label. Color + shape are a redundant channel (color-blind safety).
//
// GAP (registrar en Fase D): el token `kind` trae 6 colores (skill/agent/hook/knowledge/mcp/rule).
// `Clase` tiene 10 primitivas — command/plugin/settings/output-style/statusline NO tienen color
// propio y caen a `--muted-foreground` (fallback honesto). El dogfood solo usa skill + rule.
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
  command: { color: "var(--muted-foreground)", shape: "rounded", char: "/", label: "Comando" },
  plugin: { color: "var(--muted-foreground)", shape: "rounded", char: "P", label: "Plugin" },
  settings: { color: "var(--muted-foreground)", shape: "rounded", char: "⚙", label: "Ajustes" },
  "output-style": {
    color: "var(--muted-foreground)",
    shape: "rounded",
    char: "◐",
    label: "Output style",
  },
  statusline: {
    color: "var(--muted-foreground)",
    shape: "rounded",
    char: "▭",
    label: "Statusline",
  },
}
