import { type RefObject, useCallback, useEffect, useRef, useState } from "react"

// useViewport drives the map's pan/zoom/fit over a CSS transform on `.stage` (mockup:451-508).
// The stage lives inside the scaled transform, so edges divide their coords by z (see
// use-edge-paths). No graph engine — a fixed swim-lane geography moved by three lines of transform.
//
//  · fit()  — overview-first: frame the whole arnés (RF-50). z = min((vw-48)/cw,(vh-48)/ch,1);
//             ox = max(24,(vw-cw·z)/2); oy = 24. Called on mount + on graph change.
//  · zoomIn/zoomOut — ±0.1, clamped [0.4, 2] (RF-51). ctrl+wheel zooms.
//  · drag-pan — mousedown on the viewport background (not on a control/node/panel) pans ox/oy.

export interface Viewport {
  z: number
  transform: string
  grabbing: boolean
  fitView: () => void
  zoomIn: () => void
  zoomOut: () => void
}

const clampZoom = (z: number) => Math.min(2, Math.max(0.4, +z.toFixed(2)))

export function useViewport(
  viewportRef: RefObject<HTMLElement | null>,
  contentRef: RefObject<HTMLElement | null>,
): Viewport {
  const [z, setZ] = useState(1)
  const [ox, setOx] = useState(24)
  const [oy, setOy] = useState(24)
  const [grabbing, setGrabbing] = useState(false)

  // Latest {ox,oy} for the pan math without re-subscribing listeners on every move.
  const posRef = useRef({ ox, oy })
  useEffect(() => {
    posRef.current = { ox, oy }
  }, [ox, oy])

  const fitView = useCallback(() => {
    const c = contentRef.current
    const vp = viewportRef.current
    if (!c || !vp) return
    const cw = c.scrollWidth
    const ch = c.scrollHeight
    const vw = vp.clientWidth
    const vh = vp.clientHeight
    if (!cw || !ch) return
    const nz = +Math.min((vw - 48) / cw, (vh - 48) / ch, 1).toFixed(3)
    setZ(nz)
    setOx(Math.max(24, (vw - cw * nz) / 2))
    setOy(24)
  }, [contentRef, viewportRef])

  const zoomIn = useCallback(() => setZ((v) => clampZoom(v + 0.1)), [])
  const zoomOut = useCallback(() => setZ((v) => clampZoom(v - 0.1)), [])

  // Pan (drag) + ctrl-wheel zoom, attached once to the viewport element.
  useEffect(() => {
    const vp = viewportRef.current
    if (!vp) return
    let drag: { sx: number; sy: number; ox: number; oy: number } | null = null

    const onDown = (e: MouseEvent) => {
      const target = e.target as HTMLElement | null
      if (target?.closest(".ctl, .help-fab, .help-panel, .node")) return
      drag = { sx: e.clientX, sy: e.clientY, ox: posRef.current.ox, oy: posRef.current.oy }
      setGrabbing(true)
    }
    const onMove = (e: MouseEvent) => {
      if (!drag) return
      setOx(drag.ox + (e.clientX - drag.sx))
      setOy(drag.oy + (e.clientY - drag.sy))
    }
    const onUp = () => {
      drag = null
      setGrabbing(false)
    }
    const onWheel = (e: WheelEvent) => {
      if (!e.ctrlKey) return
      e.preventDefault()
      setZ((v) => clampZoom(v - Math.sign(e.deltaY) * 0.1))
    }

    vp.addEventListener("mousedown", onDown)
    window.addEventListener("mousemove", onMove)
    window.addEventListener("mouseup", onUp)
    vp.addEventListener("wheel", onWheel, { passive: false })
    return () => {
      vp.removeEventListener("mousedown", onDown)
      window.removeEventListener("mousemove", onMove)
      window.removeEventListener("mouseup", onUp)
      vp.removeEventListener("wheel", onWheel)
    }
  }, [viewportRef])

  return {
    z,
    transform: `translate(${ox}px, ${oy}px) scale(${z})`,
    grabbing,
    fitView,
    zoomIn,
    zoomOut,
  }
}
