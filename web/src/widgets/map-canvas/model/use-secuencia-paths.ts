import { type RefObject, useCallback, useLayoutEffect, useState } from "react"

// useSecuenciaPaths — la medición del PROCEDIMIENTO focado (N1, MA-T5): el mismo patrón
// bezier-desde-el-DOM de use-edge-paths (anclas por data-node-id, deltas de pantalla ÷ z,
// ResizeObserver), pero para la secuencia caja→caja en ORDEN de pasos. Devuelve los trazos
// (sólidos --primary, el estilo NUEVO sin colisión: los dash ya significan escribe/lee/
// opcional) y el círculo numerado en la esquina sup-izq de cada ancla — incluidos los ghosts
// E13 (`ghost-<paso>`), que anclan igual que una caja real. Con `anclas` vacío devuelve
// vacío y no toca el DOM: sin foco el overlay no existe (superset estricto, MA-L3).

export interface TrazoSecuencia {
  key: string
  d: string
}

export interface NumeroSecuencia {
  key: string
  cx: number
  cy: number
  n: number
}

export interface SecuenciaActividad {
  trazos: TrazoSecuencia[]
  numeros: NumeroSecuencia[]
}

const VACIA: SecuenciaActividad = { trazos: [], numeros: [] }

export function useSecuenciaPaths(
  contentRef: RefObject<HTMLElement | null>,
  anclas: readonly string[],
  z: number,
): SecuenciaActividad {
  const [secuencia, setSecuencia] = useState<SecuenciaActividad>(VACIA)

  const compute = useCallback(() => {
    const content = contentRef.current
    if (!content) return
    if (anclas.length === 0) {
      setSecuencia((prev) => (prev === VACIA ? prev : VACIA))
      return
    }
    const base = content.getBoundingClientRect()
    const els = anclas.map((id) =>
      content.querySelector<HTMLElement>(`[data-node-id="${CSS.escape(id)}"]`),
    )
    const trazos: TrazoSecuencia[] = []
    const numeros: NumeroSecuencia[] = []
    els.forEach((el, i) => {
      if (!el) return // ancla ausente del DOM → el paso no se dibuja, jamás una línea huérfana
      const r = el.getBoundingClientRect()
      if (!r.width) return
      const next = els[i + 1]
      if (next) {
        const rn = next.getBoundingClientRect()
        if (rn.width) {
          const x1 = (r.right - base.left) / z
          const y1 = (r.top + r.height / 2 - base.top) / z
          const x2 = (rn.left - base.left) / z
          const y2 = (rn.top + rn.height / 2 - base.top) / z
          const dx = Math.max(34, Math.abs(x2 - x1) / 2)
          trazos.push({
            key: `${anclas[i]}->${anclas[i + 1]}`,
            d: `M ${x1},${y1} C ${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`,
          })
        }
      }
      // El número del paso vive en la esquina sup-izq del ancla (mockup:685-692).
      numeros.push({
        key: `n-${anclas[i]}`,
        cx: (r.left - base.left) / z - 9,
        cy: (r.top - base.top) / z + 9,
        n: i + 1,
      })
    })
    setSecuencia({ trazos, numeros })
  }, [contentRef, anclas, z])

  useLayoutEffect(() => {
    const content = contentRef.current
    if (!content) return
    compute()
    const ro = new ResizeObserver(compute)
    ro.observe(content)
    return () => ro.disconnect()
  }, [contentRef, compute])

  return secuencia
}
