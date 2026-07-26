import type { KeyboardEvent } from "react"
import { useEffect, useId, useRef } from "react"
import type { CandidatoOrigen } from "@/entities/marketplace"
import { trapTabKeyDown } from "@/shared/lib/focus-trap"

// ResolverOrigenDialog — superficie S7 del paquete 2026-07-23-portafolio-agregar-marketplace
// (AG-D8 decisión 7 + PENDIENTE-01 §3, FIRMADAS). Vive en `widgets/portafolio` porque actúa
// **sobre un ARNÉS**: el marketplace es solo el destino elegido. No hay import cruzado con
// `widgets/marketplace` (`no-sibling-widget-imports`): la página compone los dos.
//
// No es un wizard: UN paso. `role="dialog"` + `aria-modal` + focus trap con `trapTabKeyDown`
// (patrón S1-D18: sin portal, para que `within(canvasElement)` funcione en las stories).
//
// La regla que gobierna todo: **el operador siempre confirma**. Nada premarcado, `Confirmar
// origen` nace `disabled`, el orden es sugerencia del backend (señales blandas) y NUNCA una
// decisión. Y confirmar escribe el `home` declarado: **no clona, no instala, no descarga** (BR-11).

export interface ResolverOrigenDialogProps {
  abierto: boolean
  /** para el título «Resolver origen — <id>». */
  arnesId: string
  /** ya ordenados por el backend por señales blandas (§3.5) — el widget NO reordena. */
  candidatos: CandidatoOrigen[]
  /** el home ya declarado ⇒ el diálogo lo dice. Ausente/"" = identidad provisional. */
  actual?: string | undefined
  /** `null` = «ninguno — dejarlo sin origen» (una ELECCIÓN). `undefined` = nada elegido ⇒
   *  `Confirmar` disabled. La distinción de TRES estados es lo que permite que «ninguno»
   *  habilite el botón sin premarcar nada. */
  eleccion: string | null | undefined
  onElegir: (home: string | null) => void
  confirmando: boolean
  /** slot propio de error (nace nuevo: no hay razón para repetir el parche de `observarError`).
   *  El 409 de colisión de identidad llega acá, literal. */
  error?: string | undefined
  onConfirmar: () => void
  onClose: () => void
}

const TOOLTIP_SIN_ELECCION = "elegí un destino primero"

export function ResolverOrigenDialog({
  abierto,
  arnesId,
  candidatos,
  actual,
  eleccion,
  onElegir,
  confirmando,
  error,
  onConfirmar,
  onClose,
}: ResolverOrigenDialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const cerrarBtnRef = useRef<HTMLButtonElement>(null)
  const titleId = useId()

  // Foco inicial DENTRO del diálogo al abrir (G8). El diálogo se monta sobre el drawer: cerrarlo
  // devuelve el foco al botón que lo abrió, y eso lo hace la PÁGINA (dueña del z-stack).
  useEffect(() => {
    if (abierto) cerrarBtnRef.current?.focus()
  }, [abierto])

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") {
      e.stopPropagation()
      // Mismo criterio que S1-D19: con una escritura en vuelo, cerrar a mitad es peor que
      // esperar un instante.
      if (!confirmando) onClose()
      return
    }
    if (dialogRef.current) trapTabKeyDown(dialogRef.current, e)
  }

  if (!abierto) return null

  // `undefined` es «nada elegido»; `null` («ninguno») SÍ habilita: es una decisión válida.
  const sinElegir = eleccion === undefined

  return (
    <div
      ref={dialogRef}
      className="pf-resolver"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onKeyDown={handleKeyDown}
    >
      <header className="pf-resolver-head">
        <h2 id={titleId}>
          Resolver origen — <span className="mono">{arnesId}</span>
        </h2>
        <button
          type="button"
          ref={cerrarBtnRef}
          className="pf-drawer-cerrar"
          aria-label="Cerrar"
          disabled={confirmando}
          onClick={onClose}
        >
          ✕
        </button>
      </header>

      <p className="pf-wizard-nota">
        Este arnés no puede decir de dónde viene. Elegí su marketplace de origen entre los que
        ArnesIA ya conoce. <b>Nadie decide por vos:</b> el orden es solo una sugerencia por señales
        blandas (nombre y autor del <span className="mono">plugin.json</span>) — vos confirmás.
      </p>

      {actual && (
        <p className="pf-mut">
          ya tiene origen declarado: <span className="mono">{actual}</span> — confirmar lo
          reemplaza.
        </p>
      )}

      {/* El `radiogroup` va en un `<div>`, no en un `<ul>`: un elemento de lista no puede llevar
          un rol interactivo, y meter un `list` ENTRE el radiogroup y sus radios rompería
          `aria-required-children` de axe (el gate está en `error`). */}
      <div className="pf-resolver-opciones" role="radiogroup" aria-label="Marketplace de origen">
        {candidatos.map((c) => {
          const valor = c.repo ?? c.nombre
          return (
            <label key={c.nombre} className="pf-resolver-op">
              <input
                type="radio"
                name="pf-resolver-destino"
                checked={eleccion === valor}
                disabled={confirmando}
                onChange={() => onElegir(valor)}
              />
              <span className="mono">{valor}</span>
              {c.senal && <span className="pf-resolver-senal">{c.senal}</span>}
            </label>
          )
        })}
        {/* «ninguno» es la ÚLTIMA opción y una ELECCIÓN: habilita Confirmar sin inventar un home
            (E-26). Su señal lo dice con esas palabras. */}
        <label className="pf-resolver-op">
          <input
            type="radio"
            name="pf-resolver-destino"
            checked={eleccion === null}
            disabled={confirmando}
            onChange={() => onElegir(null)}
          />
          <span>ninguno — dejarlo sin origen</span>
          <span className="pf-resolver-senal">honesto, no se inventa un home</span>
        </label>
      </div>

      {error && (
        <p role="alert" className="pf-error">
          {error}
        </p>
      )}

      <div className="pf-wizard-pie">
        <button
          type="button"
          className="pf-btn-secundario"
          disabled={confirmando}
          onClick={onClose}
        >
          Cancelar
        </button>
        <div className="derecha">
          <button
            type="button"
            className="pf-btn-primary"
            disabled={sinElegir || confirmando}
            title={sinElegir ? TOOLTIP_SIN_ELECCION : undefined}
            aria-busy={confirmando}
            onClick={onConfirmar}
          >
            {confirmando ? "Confirmando…" : "Confirmar origen"}
          </button>
        </div>
      </div>
    </div>
  )
}
