import type { ReactNode } from "react"
import { Fragment, useEffect, useRef, useState } from "react"
import { cn } from "@/shared/lib/cn"

// ListaLazy — render incremental por centinela `IntersectionObserver` (AG-D3 + AG-D16).
//
// Por qué NO es confort: el catálogo oficial real son **273 filas** (`claude-plugins-official`,
// verificado 2026-07-25) × (emblema + chips + situación + acción). Montarlas de una se nota en
// una máquina real. 273 es el caso NORMAL, no el extremo.
//
// **Nunca trunca en silencio** (E-28): cuando quedan filas sin montar lo DICE con el número
// EXACTO de faltantes y ofrece un botón que monta el paso siguiente. El botón no es decoración:
// `IntersectionObserver` puede no disparar nunca (contenedor sin scroll) y en jsdom no existe —
// sin él, un faltante se volvería un recorte mudo.
//
// Genérica y tonta: `shared/ui/**` no puede importar `entities/*` (`shared-no-upward`,
// severidad `error`), así que el vocabulario entra por `render`/`claveDe`/`etiquetaRestantes`.
export interface ListaLazyProps<T> {
  items: readonly T[]
  /** debe devolver un `<li>`: ListaLazy pone el `<ul>` y la clave, el caller la fila. */
  render: (item: T, i: number) => ReactNode
  claveDe: (item: T) => string
  paso?: number | undefined
  /** default: `… N más (se cargan al bajar)` — el número EXACTO, jamás «y algunos más». */
  etiquetaRestantes?: ((n: number) => string) | undefined
  claseLista?: string | undefined
  /** clase extra del contenedor con scroll (el centinela vive DENTRO de él). */
  claseScroll?: string | undefined
}

const PASO_DEFAULT = 20

function restantesDefault(n: number): string {
  return `… ${n} más (se cargan al bajar)`
}

export function ListaLazy<T>({
  items,
  render,
  claveDe,
  paso,
  etiquetaRestantes,
  claseLista,
  claseScroll,
}: ListaLazyProps<T>) {
  const salto = paso ?? PASO_DEFAULT
  const total = items.length
  const [montadas, setMontadas] = useState(salto)
  const pieRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // items nuevos (otro catálogo, otro escaneo) ⇒ el conteo vuelve al primer paso: montar 200
  // filas heredadas de la lista anterior sería un salto de layout sin motivo.
  // `items` está en deps a propósito: el efecto no LEE items, reacciona a que la LISTA CAMBIÓ de
  // identidad (otro catálogo, otro escaneo).
  // biome-ignore lint/correctness/useExhaustiveDependencies: ver comentario de arriba
  useEffect(() => {
    setMontadas(salto)
  }, [salto, items])

  // El centinela ES el pie: cuando el pie entra en la caja de scroll, se monta el paso
  // siguiente. Se re-observa tras cada paso (`montadas` en deps) porque el pie se corre hacia
  // abajo y puede seguir intersectando.
  //
  // Guarda deliberada: si el contenedor NO tiene scroll (`scrollHeight <= clientHeight`), el
  // pie está a la vista desde el arranque y auto-avanzar montaría todo de una — «se cargan al
  // bajar» no puede cumplirse donde no hay dónde bajar. En ese caso el botón es la ÚNICA vía,
  // que es exactamente la razón por la que el botón existe (§8.2).
  //
  // `montadas` está en deps a propósito: tras cada paso el pie se corre y hay que RE-observarlo,
  // porque puede seguir intersectando.
  // biome-ignore lint/correctness/useExhaustiveDependencies: ver comentario de arriba
  useEffect(() => {
    const pie = pieRef.current
    const caja = scrollRef.current
    if (pie && caja && typeof IntersectionObserver !== "undefined") {
      const obs = new IntersectionObserver(
        (entradas) => {
          if (caja.scrollHeight <= caja.clientHeight) return
          if (entradas.some((e) => e.isIntersecting)) {
            setMontadas((n) => Math.min(n + salto, total))
          }
        },
        { root: caja },
      )
      obs.observe(pie)
      return () => obs.disconnect()
    }
    return undefined
  }, [salto, total, montadas])

  const visibles = items.slice(0, montadas)
  const restantes = total - visibles.length
  const etiqueta = etiquetaRestantes ?? restantesDefault

  return (
    <div ref={scrollRef} className={cn("pf-lazy", claseScroll)}>
      <ul className={claseLista}>
        {visibles.map((item, i) => (
          <Fragment key={claveDe(item)}>{render(item, i)}</Fragment>
        ))}
      </ul>
      {restantes > 0 && (
        <div ref={pieRef} className="cat-lazy">
          <span>{etiqueta(restantes)}</span>
          <button
            type="button"
            className="pf-btn-mini"
            onClick={() => setMontadas((n) => Math.min(n + salto, total))}
          >
            Cargar más
          </button>
        </div>
      )}
    </div>
  )
}
