import type { ChangeEvent, ReactNode } from "react"

// BuscadorFiltro — composición: un `input[type=search]` + N `FiltroDisclosure` ya armados por
// el caller (AG-D3: buscador + filtro dentro de toda lista de arneses-a-elegir del flujo de
// agregación, porque AG-D7 obliga a tildar a mano y el caso de 30+ filas es el normal).
//
// **No sabe filtrar**: solo compone y rotula. Los selectores de filtrado son del dominio
// (`entities/*/model/selectors.ts`) y `shared/ui/**` no puede importarlos (`shared-no-upward`,
// severidad `error`) — el caller pasa `busqueda`/`onBusqueda` ya cableados.
export interface BuscadorFiltroProps {
  busqueda: string
  onBusqueda: (q: string) => void
  placeholder: string
  ariaLabel: string
  /** los `FiltroDisclosure` (o cualquier control) que acompañan al buscador. */
  filtros?: ReactNode | undefined
}

export function BuscadorFiltro({
  busqueda,
  onBusqueda,
  placeholder,
  ariaLabel,
  filtros,
}: BuscadorFiltroProps) {
  return (
    <div className="pf-toolbar">
      <input
        type="search"
        className="pf-buscar"
        aria-label={ariaLabel}
        placeholder={placeholder}
        value={busqueda}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onBusqueda(e.target.value)}
      />
      {filtros}
    </div>
  )
}
