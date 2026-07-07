import type { CSSProperties } from "react"
import { Glyph } from "@/shared/canvas"
import { cn } from "@/shared/lib/cn"
import { KIND } from "../model/kind"
import { handleFor, isCaja, isPropuesto, transLabel } from "../model/node-view"
import { alwFor, isDelPuesto } from "../model/proposals"
import type { Box } from "../model/types"

// ArnesNode is the map's node card (`.node`, mockup:91,336-351). Structure layer (MVP): the
// type is carried by the glyph (shape+color+char) + the derived handle; the name is mono,
// 2-line clamped. It DERIVES its marks: caja/transición/propuesto from REAL data (contract·
// estado·canal via node-view), origen/alw as PROPUESTA from the isolated fixture sets
// (proposals). Positional/interaction state (support·compact·dim·selected) comes from the
// canvas. Styling lives in app/styles/map.css (scoped .arnesia-map) — this emits the mockup DOM.

// NodeStyle extends CSSProperties with the per-node type-color custom property (--tc), which
// map.css reads for the left border, tint, badges and transition tag.
type NodeStyle = CSSProperties & { "--tc"?: string }

interface ArnesNodeProps {
  box: Box
  // Positional/interaction flags owned by the canvas (not derivable from the node alone).
  support?: boolean | undefined
  compact?: boolean | undefined
  dim?: boolean | undefined
  selected?: boolean | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function ArnesNode({ box, support, compact, dim, selected, onSelect }: ArnesNodeProps) {
  const k = KIND[box.clase]
  const caja = isCaja(box)
  const puesto = isDelPuesto(box.id)
  const propuesto = isPropuesto(box)
  const alw = alwFor(box.id)
  const trans = transLabel(box)
  const style: NodeStyle = { "--tc": k.color }

  return (
    <button
      type="button"
      data-node-id={box.id}
      aria-pressed={selected}
      onClick={() => onSelect?.(box.id)}
      style={style}
      className={cn(
        "node",
        caja && "caja",
        puesto && "puesto",
        support && "support",
        compact && "compact",
        dim && "dim",
        selected && "selected",
      )}
    >
      {caja && <span className="caja-badge">caja</span>}
      {propuesto && <span className="prop-badge">propuesto</span>}
      <span className="node-top">
        <Glyph color={k.color} char={k.char} shape={k.shape} />
        <span className="node-nm">{box.nombre}</span>
      </span>
      <span className="node-cmd">{handleFor(box, alw)}</span>
      {trans && (
        <span className="node-trans" title="transición del spine que posee esta caja">
          {trans}
        </span>
      )}
    </button>
  )
}
