import type { CSSProperties } from "react"
import type { EdgePath } from "../model/use-edge-paths"

// EdgeLayer is the SVG overlay that draws the measured connectors over the bands/lanes. Styling by
// tipo mirrors the signed v3 mockup: invoca = solid + arrow · lee = dotted, no arrow · escribe =
// accent-dashed + arrow. Non-interactive (pointer-events:none) so node clicks pass through.

const STYLE: Record<EdgePath["tipo"], CSSProperties> = {
  invoca: { stroke: "var(--input)", opacity: 0.55 },
  lee: { stroke: "var(--input)", opacity: 0.3, strokeDasharray: "4 4" },
  escribe: { stroke: "var(--primary)", opacity: 0.5, strokeDasharray: "2 3" },
}

export function EdgeLayer({ paths }: { paths: EdgePath[] }) {
  return (
    <svg
      aria-hidden
      className="pointer-events-none absolute inset-0 z-[3] h-full w-full overflow-visible"
    >
      <title>Conexiones entre nodos del arnés</title>
      <defs>
        <marker
          id="arnes-arrow"
          viewBox="0 0 10 10"
          refX="8"
          refY="5"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M0,0 L10,5 L0,10 z" fill="context-stroke" />
        </marker>
      </defs>
      {paths.map((p) => (
        <path
          key={p.key}
          d={p.d}
          fill="none"
          strokeWidth={1.4}
          style={STYLE[p.tipo]}
          markerEnd={p.tipo === "lee" ? undefined : "url(#arnes-arrow)"}
        />
      ))}
    </svg>
  )
}
