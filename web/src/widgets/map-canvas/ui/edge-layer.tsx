import type { EdgePath } from "../model/use-edge-paths"
import type { SecuenciaActividad } from "../model/use-secuencia-paths"

// EdgeLayer is the SVG overlay drawn over the bands/lanes (mockup:152-158,460-497). It is a
// dumb renderer: useEdgePaths already measured geometry and resolved each path's stroke/opacity/
// width/dash/marker (spine-always + hover-reveal gating). It lives INSIDE `.content` so its
// coordinate space matches the measured anchors; `pointer-events:none` lets node clicks pass
// through. Styling (position/overflow) is in map.css (`.arnesia-map svg.edges`).
//
// `secuencia` (MA-T5, N1): el procedimiento de la actividad focada dibujado EN ESTA MISMA capa
// — trazo sólido `--primary` con flecha + círculo numerado por paso. Estilo NUEVO sin colisión:
// los dash ya significan escribe (3 3) · lee (4 4) · opcional (2 6), jamás dash acá.
// `undefined`/vacía ⇒ cero DOM extra (superset MA-L3).

export function EdgeLayer({
  paths,
  secuencia,
}: {
  paths: EdgePath[]
  secuencia?: SecuenciaActividad | undefined
}) {
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
      {secuencia && (secuencia.trazos.length > 0 || secuencia.numeros.length > 0) && (
        <g className="secuencia-actividad">
          {secuencia.trazos.map((t) => (
            <path
              key={t.key}
              className="seq-trazo"
              d={t.d}
              fill="none"
              stroke="var(--primary)"
              strokeWidth={2.6}
              opacity={0.95}
              markerEnd="url(#arnes-arr)"
            />
          ))}
          {secuencia.numeros.map((num) => (
            <g key={num.key} className="seq-num">
              <circle cx={num.cx} cy={num.cy} r={9} fill="var(--primary)" />
              <text
                x={num.cx}
                y={num.cy + 3.5}
                textAnchor="middle"
                fontSize={10}
                fontWeight={700}
                fill="var(--primary-foreground)"
              >
                {num.n}
              </text>
            </g>
          ))}
        </g>
      )}
    </svg>
  )
}
