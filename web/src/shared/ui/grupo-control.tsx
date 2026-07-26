import type { ReactNode } from "react"

// GrupoControl — rótulo VISIBLE + grupo de controles (AG-D4, cierra el defecto L1 de la
// auditoría: hoy la toolbar del Portafolio muestra `Empresa · Plano · Proyecto · Marketplace ·
// Estado · Marketplace` — seis pills iguales donde la 4ª es una LENTE y la 6ª un FILTRO, sin
// ninguna distinción a la vista. Los `role="group"` + `aria-label` ya existían, así que un
// lector de pantalla los distinguía; un ojo, no).
//
// El rótulo pasa a texto visible SIN perder el `role="group"` + `aria-label` existente: no se
// degrada a11y, se agrega afordancia visual. Las mayúsculas las pone el CSS
// (`text-transform: uppercase`), no el dato — el DOM dice «Ver por», igual que el mockup, así
// que `getByText("Ver por")` funciona (E-27).
//
// Genérico y tonto a propósito: `shared/ui/**` no puede importar `entities/*`
// (`shared-no-upward`, severidad `error`).
export interface GrupoControlProps {
  /** «Ver por» | «Filtros» — el texto tal cual va al DOM; las mayúsculas son CSS. */
  rotulo: string
  /** el `aria-label` EXISTENTE del grupo se conserva (no se degrada a11y). */
  ariaLabel: string
  children: ReactNode
  /** hint opcional a la derecha (defecto L3 de la auditoría: el mockup lo tenía). */
  hint?: string | undefined
  /** clase del contenedor interno de los controles (`pf-lentes` | `pf-filtros`). */
  claseControles?: string | undefined
}

export function GrupoControl({
  rotulo,
  ariaLabel,
  children,
  hint,
  claseControles,
}: GrupoControlProps) {
  return (
    <div className="pf-grupo-control">
      <span className="pf-grupo-control-rotulo">{rotulo}</span>
      <div className={claseControles ?? "pf-lentes"} role="group" aria-label={ariaLabel}>
        {children}
      </div>
      {hint && <span className="pf-grupo-control-hint">{hint}</span>}
    </div>
  )
}
