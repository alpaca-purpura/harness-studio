// Skeleton · ErrorBody — los dos estados honestos de carga y de fallo, compartidos por el
// Portafolio y por la capa «Mejora» del Mapa (H-6).
//
// PROMOVIDOS de `widgets/portafolio/ui/portafolio-list.tsx`, donde vivían privados
// (plan-desarrollo T28 · design.md §1.4). Es refactor de MOVIMIENTO, no de conducta: mismo DOM,
// mismas clases, mismo contrato ARIA — el criterio de éxito es que las 12 stories firmadas de
// `portafolio-list.stories.tsx` sigan pasando SIN tocarlas. Misma jugada que `FiltroDisclosure`
// en el paquete de marketplace.
//
// Puramente presentacionales (props `label`, `motivo`, `onReintentar`): cero dominio ⇒ no violan
// `ui-not-domain` (dependency-cruiser, severidad `error`). El texto de dominio —«Cargando
// portafolio», «No se pudo cargar el portafolio — …»— lo inyecta el llamador: `shared/ui/**` no
// puede importar `entities/*` (`shared-no-upward`) y tampoco debe conocer el vocabulario de
// ninguna superficie.
//
// Las clases `pf-*` se conservan A PROPÓSITO (mismo criterio que `FiltroDisclosure`, que
// conservó `pf-filtro`): renombrarlas obligaría a reescribir el CSS ya firmado del Portafolio,
// que es exactamente el cambio de conducta que un refactor de movimiento no puede permitirse.
// La capa Mejora las estiliza bajo su propio scope (`mejora.css`).

export interface SkeletonProps {
  /** Qué se está cargando, dicho: es el nombre accesible del `role="status"`. */
  label: string
  /** Cuántas barras fantasma. 3 = el default histórico del Portafolio. */
  filas?: number | undefined
  /** Alto de cada barra, cuando el llamador necesita la silueta de SU fila (design §5.4). */
  altura?: number | undefined
  /** Marca de datos para asserts estructurales de las stories (p. ej. `mejora`). */
  data?: string | undefined
}

// Skeleton — CERO filas fantasma que se puedan confundir con datos: las barras son
// `aria-hidden`, y lo único que un lector de pantalla anuncia es el `label`.
export function Skeleton({ label, filas = 3, altura, data }: SkeletonProps) {
  return (
    <div className="pf-skeleton" role="status" aria-live="polite" aria-label={label}>
      {Array.from({ length: filas }, (_, i) => (
        <span
          // biome-ignore lint/suspicious/noArrayIndexKey: barras decorativas idénticas — el índice ES la identidad.
          key={i}
          className="pf-skeleton-fila"
          aria-hidden="true"
          data-skeleton={data}
          style={altura === undefined ? undefined : { height: altura }}
        />
      ))}
    </div>
  )
}

export interface ErrorBodyProps {
  /** El motivo REAL, ya redactado por el llamador. Nunca «Error al obtener los datos». */
  motivo: string
  onReintentar: () => void
}

// ErrorBody — el motivo textual va en un `role="alert"` y el reintento es un botón real. El
// motivo lo arma el llamador porque nombra su superficie; acá no se inventa ninguno.
export function ErrorBody({ motivo, onReintentar }: ErrorBodyProps) {
  return (
    <div className="pf-estado-vacio">
      <p className="pf-mut" role="alert">
        {motivo}
      </p>
      <button type="button" className="pf-btn-primary" onClick={onReintentar}>
        Reintentar
      </button>
    </div>
  )
}
