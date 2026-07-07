import type { EdgePath } from "../model/use-edge-paths"

// EdgeLayer is the SVG overlay drawn over the bands/lanes (mockup:152-158,460-497). It is a
// dumb renderer: useEdgePaths already measured geometry and resolved each path's stroke/opacity/
// width/dash/marker (spine-always + hover-reveal gating). It lives INSIDE `.content` so its
// coordinate space matches the measured anchors; `pointer-events:none` lets node clicks pass
// through. Styling (position/overflow) is in map.css (`.arnesia-map svg.edges`).

export function EdgeLayer({ paths }: { paths: EdgePath[] }) {
  return (
    <svg className="edges" aria-hidden>
      <title>Relaciones entre nodos del arnés</title>
      <defs>
        <marker
          id="arnes-arr"
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
          stroke={p.stroke}
          strokeWidth={p.width}
          strokeDasharray={p.dash}
          opacity={p.opacity}
          markerEnd={p.marker ? "url(#arnes-arr)" : undefined}
        />
      ))}
    </svg>
  )
}
