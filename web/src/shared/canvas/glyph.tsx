import type { CSSProperties } from "react"

// Glyph is the dumb canvas primitive for a node's type mark: a colored shape holding a mono
// letter. Color + SHAPE together encode the type (redundant channel for color-blind reading —
// an explicit map principle). It knows nothing about the arnés domain; the caller (entities/arnes)
// maps a `clase` to a color/shape/char. Decorative → aria-hidden (the node name carries meaning).

export type GlyphShape = "square" | "circle" | "diamond" | "hexagon" | "shield" | "rounded"

const SHAPE: Record<GlyphShape, CSSProperties> = {
  square: { borderRadius: 3 },
  circle: { borderRadius: "50%" },
  rounded: { borderRadius: 4 },
  diamond: { clipPath: "polygon(50% 0, 100% 50%, 50% 100%, 0 50%)" },
  hexagon: { clipPath: "polygon(25% 0, 75% 0, 100% 50%, 75% 100%, 25% 100%, 0 50%)" },
  shield: { clipPath: "polygon(0 0, 100% 0, 100% 62%, 50% 100%, 0 62%)" },
}

interface GlyphProps {
  color: string // a token reference, e.g. "var(--c-skill)"
  char: string
  shape: GlyphShape
  size?: number
}

export function Glyph({ color, char, shape, size = 17 }: GlyphProps) {
  return (
    <span
      aria-hidden
      className="inline-grid shrink-0 place-items-center font-mono font-semibold leading-none text-[color:var(--background)]"
      style={{
        background: color,
        width: size,
        height: size,
        fontSize: Math.round(size * 0.58),
        ...SHAPE[shape],
      }}
    >
      {char}
    </span>
  )
}
