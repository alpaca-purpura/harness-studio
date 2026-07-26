import { toggleEnSet } from "@/shared/lib/toggle-en-set"

// FiltroDisclosure — botón que despliega un panel de chips toggleables (multi-select). Mismo
// patrón de disclosure ya sancionado por `rules-subband.tsx` (aria-expanded, sin portal, sin
// click-outside): la afordancia es DISTINTA de una lente (acota la lista, no la reagrupa).
//
// PROMOVIDO de `widgets/portafolio/ui/portafolio-list.tsx`, donde vivía privado
// (paquete 2026-07-23-portafolio-agregar-marketplace §8.2). Es refactor de MOVIMIENTO, no de
// conducta: mismo DOM, mismas clases, mismo `aria-expanded` en el botón — el criterio de éxito
// del refactor es que `FiltroEstadoAcota`/`FiltroMarketplaceAcota` sigan pasando sin tocarlas.
//
// Genérico `<T extends string>` con `labelDe` inyectado a propósito: `shared/ui/**` NO PUEDE
// importar `entities/*` (dependency-cruiser `shared-no-upward`, severidad `error`), así que el
// vocabulario de dominio (`SALUD_LABEL`, situaciones del catálogo) entra por prop y esta
// molécula no conoce ningún dominio.
export interface FiltroDisclosureProps<T extends string> {
  etiqueta: string
  abierto: boolean
  onToggleAbierto: () => void
  panelId: string
  valores: readonly T[]
  labelDe: (v: T) => string
  seleccion: ReadonlySet<T>
  onCambiar: (s: ReadonlySet<T>) => void
  vacio?: string | undefined
}

export function FiltroDisclosure<T extends string>({
  etiqueta,
  abierto,
  onToggleAbierto,
  panelId,
  valores,
  labelDe,
  seleccion,
  onCambiar,
  vacio,
}: FiltroDisclosureProps<T>) {
  return (
    <div className="pf-filtro">
      <button
        type="button"
        className="pf-filtro-btn"
        aria-expanded={abierto}
        aria-controls={panelId}
        onClick={onToggleAbierto}
      >
        {etiqueta}
        {seleccion.size > 0 && ` (${seleccion.size})`}
      </button>
      {abierto && (
        <div
          id={panelId}
          className="pf-filtro-panel"
          role="group"
          aria-label={`Filtrar por ${etiqueta.toLowerCase()}`}
        >
          {valores.length === 0 && vacio && <p className="pf-mut">{vacio}</p>}
          {valores.map((v) => (
            <button
              key={v}
              type="button"
              className="pf-lente-btn"
              aria-pressed={seleccion.has(v)}
              onClick={() => onCambiar(toggleEnSet(seleccion, v))}
            >
              {labelDe(v)}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
