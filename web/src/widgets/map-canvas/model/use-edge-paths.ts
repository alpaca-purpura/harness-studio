import { type RefObject, useCallback, useLayoutEffect, useState } from "react"
import type { Edge, TipoEdge } from "@/entities/arnes"

// EdgePath is one measured, fully-styled connector: a cubic bezier from the right edge of the
// source node to the left edge of the target, in CONTENT space (screen deltas ÷ z, since the
// SVG lives inside the scaled stage). It carries its own stroke/opacity/width/dash/marker so
// EdgeLayer is a dumb renderer.
export interface EdgePath {
  key: string
  d: string
  stroke: string
  opacity: number
  width: number
  dash?: string | undefined
  marker: boolean
}

// STROKE — style per relation type (design §9.1, mockup:239-243): invoca = spine/backbone
// (crit, solid, arrow) · escribe (ok, dash 3 3, arrow) · lee (warn, dash 4 4, NO arrow).
const STROKE: Record<TipoEdge, { color: string; opacity: number; dash?: string }> = {
  invoca: { color: "var(--crit)", opacity: 0.75 },
  escribe: { color: "var(--ok)", opacity: 0.7, dash: "3 3" },
  lee: { color: "var(--warn)", opacity: 0.65, dash: "4 4" },
}

interface Opts {
  z: number
  focusId?: string | null | undefined
}

// useEdgePaths measures node anchors from the DOM (the signed v3 approach) and returns the
// bezier paths to draw, applying the spine-always + hover-reveal gating (design §9.3):
//  · without focus, only `invoca` (the backbone) is drawn (RF-31);
//  · on hover, the focused node's `lee`/`escribe` are revealed too (RF-32); the touched edge is
//    highlighted (opacity .95, width 2.2) and the untouched backbone dims (opacity .2) (RF-33);
//  · a collapsed/hidden endpoint (width 0) drops its edge — no orphan lines (RF-34).
// Recomputes on content resize and whenever z or the focus change.
export function useEdgePaths(
  contentRef: RefObject<HTMLElement | null>,
  edges: Edge[],
  { z, focusId }: Opts,
): EdgePath[] {
  const [paths, setPaths] = useState<EdgePath[]>([])

  const compute = useCallback(() => {
    const content = contentRef.current
    if (!content) return
    const base = content.getBoundingClientRect()
    const next: EdgePath[] = []
    for (const e of edges) {
      const spine = e.tipo === "invoca"
      const touches = !!focusId && (e.de === focusId || e.a === focusId)
      // Backbone always visible; lee/escribe only for the node under the cursor (no spaghetti).
      if (!spine && !touches) continue
      const from = content.querySelector<HTMLElement>(`[data-node-id="${CSS.escape(e.de)}"]`)
      const to = content.querySelector<HTMLElement>(`[data-node-id="${CSS.escape(e.a)}"]`)
      if (!from || !to) continue
      const f = from.getBoundingClientRect()
      const t = to.getBoundingClientRect()
      if (!f.width || !t.width) continue // endpoint collapsed/hidden (folded rules) → skip
      const x1 = (f.right - base.left) / z
      const y1 = (f.top + f.height / 2 - base.top) / z
      const x2 = (t.left - base.left) / z
      const y2 = (t.top + t.height / 2 - base.top) / z
      const dx = Math.max(30, Math.abs(x2 - x1) / 2)
      const s = STROKE[e.tipo]
      let opacity = s.opacity
      let width = 1.6
      if (focusId) {
        if (touches) {
          opacity = 0.95
          width = 2.2
        } else if (spine) {
          opacity = 0.2
        }
      }
      next.push({
        key: `${e.de}->${e.a}`,
        d: `M ${x1},${y1} C ${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`,
        stroke: s.color,
        opacity,
        width,
        dash: s.dash,
        marker: e.tipo !== "lee",
      })
    }
    setPaths(next)
  }, [contentRef, edges, z, focusId])

  useLayoutEffect(() => {
    const content = contentRef.current
    if (!content) return
    compute()
    const ro = new ResizeObserver(compute)
    ro.observe(content)
    return () => ro.disconnect()
  }, [contentRef, compute])

  return paths
}
