import { type RefObject, useCallback, useLayoutEffect, useState } from "react"
import type { Edge } from "@/entities/arnes"

// EdgePath is one measured connector: a cubic bezier from the right edge of the source node to
// the left edge of the target node, in coordinates relative to the content box.
export interface EdgePath {
  key: string
  d: string
  tipo: Edge["tipo"]
}

// useEdgePaths measures node anchors from the DOM (the signed v3 approach: nodes flow in HTML,
// edges are an SVG overlay measured by getBoundingClientRect) and returns bezier paths. Recomputes
// on content resize and when the edge set changes. This is why the map is HTML+SVG, not a graph
// engine — the geography is fixed swim-lanes, the connectors are drawn over them.
export function useEdgePaths(contentRef: RefObject<HTMLElement | null>, edges: Edge[]): EdgePath[] {
  const [paths, setPaths] = useState<EdgePath[]>([])

  const compute = useCallback(() => {
    const content = contentRef.current
    if (!content) return
    const base = content.getBoundingClientRect()
    const next: EdgePath[] = []
    for (const e of edges) {
      const from = content.querySelector<HTMLElement>(`[data-node-id="${CSS.escape(e.de)}"]`)
      const to = content.querySelector<HTMLElement>(`[data-node-id="${CSS.escape(e.a)}"]`)
      if (!from || !to) continue
      const f = from.getBoundingClientRect()
      const t = to.getBoundingClientRect()
      const x1 = f.right - base.left
      const y1 = f.top + f.height / 2 - base.top
      const x2 = t.left - base.left
      const y2 = t.top + t.height / 2 - base.top
      const dx = Math.max(28, Math.abs(x2 - x1) / 2)
      next.push({
        key: `${e.de}->${e.a}`,
        d: `M ${x1},${y1} C ${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`,
        tipo: e.tipo,
      })
    }
    setPaths(next)
  }, [contentRef, edges])

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
