import { useEffect, useId, useRef, useState } from "react"
import { trapTabKeyDown } from "@/shared/lib/focus-trap"

// PoliticaDatosDialog — «qué guardamos» y el borrado (RF-275 · RF-278 · RF-283 · H-5 · H-14).
//
// Vive **detrás del enlace «qué guardamos» de la franja** (H-5): el mockup lo dibuja como texto
// dentro de un panel didáctico, que no es una ubicación.
//
// La razón de que exista: la promesa *«nada de tu cuenta, nada de la conversación»* tiene que ser
// **contrastable**, no una frase. Por eso:
//
//  1. **La lista de campos viene POR PROP (`camposPersistidos`), no hardcodeada.** RF-275 exige
//     que sea *«la misma allowlist que aplica la ingesta, no una redacción aparte»*. Si el
//     componente la inventa, la promesa deja de ser verificable — y ese es el único valor que
//     tiene un diálogo de privacidad.
//  2. **El `{N}` de retención sale de la config.** ⚠️ El `90` es PROPUESTO, no firmado (J-6):
//     hardcodearlo acá lo convertiría en decisión por omisión.
//  3. **H-14** — el copy dice qué pasa con lo que **sí** llega (el contenido por el canal de
//     hooks) y que se descarta ANTES de escribirse. Callarlo dejaría al lector suponiendo que
//     el contenido nunca sale de su máquina, que es una promesa distinta.
//
// El botón destructivo nace `disabled` hasta que el operador confirma el alcance, y **mientras
// borra el diálogo no se puede cerrar** (mismo patrón que el wizard con POST en vuelo, S1-D19):
// cerrarlo dejaría una operación en vuelo sin dónde reportar su resultado.

export type EstadoBorrado = "reposo" | "confirmando" | "borrando" | "error" | "exito"

export interface PoliticaDatosDialogProps {
  /** El arnés del que se habla. Aparece en el título de la confirmación. */
  arnes: string
  /**
   * La allowlist REAL que aplica la ingesta (`ANEXO-hooks.md` H4). **No la inventa la UI.**
   * El DOM refleja exactamente este array, ni uno más ni uno menos.
   */
  camposPersistidos: readonly string[]
  /** El TTL vigente, de la config. ⚠️ El `90` del mockup es PROPUESTO, no firmado (J-6). */
  retencionDias: number
  retencionPropuesta?: boolean | undefined
  /** Cuántas corridas se van a borrar. Se dice antes, no después. */
  corridasPorBorrar: number
  estado?: EstadoBorrado | undefined
  error?: string | undefined
  onBorrar: () => void
  onCerrar: () => void
}

export function PoliticaDatosDialog({
  arnes,
  camposPersistidos,
  retencionDias,
  retencionPropuesta,
  corridasPorBorrar,
  estado = "reposo",
  error,
  onBorrar,
  onCerrar,
}: PoliticaDatosDialogProps) {
  const ref = useRef<HTMLDivElement>(null)
  const tituloId = useId()
  const camposId = useId()
  const [camposAbiertos, setCamposAbiertos] = useState(false)
  // `borrando` y `error` solo existen DENTRO del paso de confirmación: son estados de la
  // operación destructiva, no del panel didáctico.
  const [confirmando, setConfirmando] = useState(
    estado === "confirmando" || estado === "borrando" || estado === "error",
  )
  const [entendido, setEntendido] = useState(false)
  const borrando = estado === "borrando"

  // Foco inicial dentro del diálogo (RF-278). Sin esto, el teclado queda detrás del modal.
  useEffect(() => {
    ref.current?.focus()
  }, [])

  // Esc cierra EN REPOSO. Mientras borra, no: cerrar dejaría una operación en vuelo sin dónde
  // reportar su resultado, y el operador no sabría si borró o no.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !borrando) onCerrar()
    }
    document.addEventListener("keydown", onKey)
    return () => document.removeEventListener("keydown", onKey)
  }, [borrando, onCerrar])

  return (
    // El `role="dialog"` es el contenedor del focus trap (patrón sancionado,
    // `shared/lib/focus-trap.ts`): el `onKeyDown` es de teclado y no reemplaza a ningún control.
    <div
      ref={ref}
      className="arnesia-mejora pol-dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby={tituloId}
      tabIndex={-1}
      onKeyDown={(e) => ref.current && trapTabKeyDown(ref.current, e)}
    >
      {confirmando ? (
        <>
          <h2 id={tituloId} className="pol-titulo">{`Borrar la telemetría de «${arnes}»`}</h2>
          <p>
            {`Se borran ${corridasPorBorrar.toLocaleString("es-AR").replace(/\./gu, " ")} corridas medidas y los puntos de mejora que salieron de ellas. No se puede deshacer.`}
          </p>
          <label className="pol-confirm">
            <input
              type="checkbox"
              checked={entendido}
              disabled={borrando}
              onChange={(e) => setEntendido(e.target.checked)}
            />
            Entiendo que no se puede deshacer.
          </label>
          {estado === "error" && (
            <p className="pol-error" role="alert">
              {`No se pudo borrar — ${error ?? "motivo desconocido"}. No se borró nada.`}
            </p>
          )}
          <div className="pol-acciones">
            <button
              type="button"
              className="pol-btn"
              disabled={borrando}
              onClick={() => setConfirmando(false)}
            >
              Cancelar
            </button>
            <button
              type="button"
              className="pol-btn pol-btn-destructivo"
              disabled={!entendido || borrando}
              aria-busy={borrando}
              onClick={onBorrar}
            >
              {borrando ? "Borrando…" : "Borrar"}
            </button>
          </div>
        </>
      ) : (
        <>
          <h2 id={tituloId} className="pol-titulo">
            Qué guardamos de la telemetría
          </h2>

          <section className="pol-bloque">
            <h3>Qué NO se guarda</h3>
            <p>
              No guardamos nada de tu cuenta: ni email, ni identificadores de usuario, ni de
              organización.
            </p>
            {/* H-14 — lo que SÍ llega por el canal de hooks y se descarta antes de escribirse.
                Callarlo dejaría al lector suponiendo una promesa distinta de la real. */}
            <p>
              Y no guardamos nada del contenido: ni tu prompt, ni la respuesta, ni lo que leyó o
              escribió una herramienta. Eso llega por el canal de hooks y se descarta antes de
              escribirse en disco.
            </p>
          </section>

          <section className="pol-bloque">
            <h3>Qué SÍ se guarda</h3>
            <p>
              Guardamos, por turno: qué arnés, qué caja, qué sesión, qué modelo, cuántos tokens de
              cada tipo, cuánto costó, cuánto tardó y si el gate lo aceptó o lo rechazó.
            </p>
            <button
              type="button"
              className="pol-link"
              aria-expanded={camposAbiertos}
              aria-controls={camposId}
              onClick={() => setCamposAbiertos((a) => !a)}
            >
              ver los campos exactos
            </button>
            {/* La lista viene por prop: es la MISMA allowlist que aplica la ingesta. Si la UI la
                inventara, la promesa dejaría de ser contrastable. */}
            <ul className="pol-campos" id={camposId} hidden={!camposAbiertos}>
              {camposPersistidos.map((campo) => (
                <li key={campo}>
                  <code>{campo}</code>
                </li>
              ))}
            </ul>
          </section>

          <section className="pol-bloque">
            <h3>Retención</h3>
            <p>
              {`Se borra solo a los ${retencionDias} días.`}
              {retencionPropuesta && (
                <em className="pol-propuesto"> Valor propuesto, sin firmar.</em>
              )}
            </p>
            <div className="pol-acciones">
              <button type="button" className="pol-btn" onClick={onCerrar}>
                Cerrar
              </button>
              <button
                type="button"
                className="pol-btn pol-btn-destructivo"
                onClick={() => setConfirmando(true)}
              >
                Borrar la telemetría de este arnés
              </button>
            </div>
          </section>
        </>
      )}

      <div className="pol-live" aria-live="polite">
        {estado === "exito" && "Listo. Este arnés vuelve a estar sin datos de telemetría."}
      </div>
    </div>
  )
}
