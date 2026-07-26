import { useId, useState } from "react"
import type { PuntoMejora } from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"

// PuntoMejoraCard — una tarjeta = una caja × un detector (design §2.5, §5.4, §7.4).
//
// **La anatomía es el argumento**, no una preferencia de layout:
//   qué pasa (titular) → cuánto (lede) → cuánto se ahorra (contrafactual) → por qué lo creemos
//   (umbral o patrón) → con qué confianza → contra qué (sesgo) → qué hacer (fix).
// Mover el contrafactual debajo del sesgo convertiría la tarjeta en un reproche con nota al pie.
//
// Las cinco cosas que A4 exige — número, contrafactual, umbral citado, sesgo EN CONTRA y UN fix
// — son obligatorias. **Sin contrafactual no hay tarjeta**: la lista la filtra (no se pinta
// degradada) y el inspector la lista como «sin fix propuesto».
//
// 🔴 **`Proponerlo en el chat` NO ESCRIBE ARCHIVOS** (D17.3 · BR-M12). No existe ninguna prop de
// escritura en este componente, y la story lo asserta por ausencia: abre el chat con el cambio
// propuesto, y el cambio se aplica por el camino de siempre, con sus permisos y su gate.

export interface PuntoMejoraCardProps {
  punto: PuntoMejora
  /** Su caja está seleccionada en el canvas. La señal lleva TEXTO, no solo un borde de 6 px. */
  resaltada?: boolean | undefined
  /** Motivo por el que no se puede proponer. Va en `title` **y** en texto visible (design §5.4). */
  proponerDeshabilitado?: string | undefined
  onDescartar: (puntoId: string) => void
  onProponer: (p: { puntoId: string; textoPropuesto: string }) => void
  // 🔴 NO HAY `onAplicar` NI `onEscribir`. Su ausencia es el contrato (BR-M12).
}

export function PuntoMejoraCard({
  punto,
  resaltada,
  proponerDeshabilitado,
  onDescartar,
  onProponer,
}: PuntoMejoraCardProps) {
  const [abierto, setAbierto] = useState(false)
  const [descartado, setDescartado] = useState(false)
  const calcId = useId()
  const severidad = punto.grave ? "crítico" : "atención"

  const confianza =
    punto.confianza === "exacta"
      ? `exacta · ${punto.corridas_usadas} de ${punto.corridas_totales} corridas con atribución`
      : (punto.confianza_detalle ??
        `no exacta · ${punto.corridas_usadas} de ${punto.corridas_totales} corridas con atribución`)

  return (
    <article
      className={cn("mejora", punto.grave && "grave", resaltada && "resaltada")}
      data-punto={punto.id}
      aria-labelledby={`${calcId}-titular`}
    >
      <header className="mej-head">
        {/* RF-257 · RF-280 — la severidad es TEXTO. El borde de color es refuerzo y el ⚠ es
            decorativo: en escala de grises la tarjeta sigue diciendo cuál es cuál. */}
        <span className={cn("mej-sev", punto.grave ? "sev-crit" : "sev-warn")}>{severidad}</span>
        {/* J-2 — sin este chip, en una instalación mayormente S2 la tarjeta insignia
            desaparecería sin explicación. Tono neutro: es información, no señal. */}
        {punto.solo_s1 && <span className="mej-chip-neutro">solo con telemetría de ArnesIA</span>}
        {resaltada && <span className="mej-resaltada">↔ caja seleccionada</span>}
      </header>

      <h4 className="mej-titular" id={`${calcId}-titular`}>
        <span aria-hidden="true">⚠</span> {punto.titulo}
      </h4>

      <p className="mej-lede">{punto.lede}</p>

      <dl className="mej-dl">
        <dt>Contrafactual</dt>
        <dd>{punto.contrafactual}</dd>

        {/* RF-250 — `Umbral` solo cuando el detector SE DECIDE POR UN UMBRAL. P1 no: se decide
            por un patrón de motivos de rechazo, y pintarle una desigualdad inventada sería
            fabricar el rigor que no tiene. */}
        {punto.umbral !== undefined ? (
          <>
            <dt>Umbral</dt>
            <dd className="mej-umbral">
              <span className="mono">{punto.umbral}</span>
              {punto.calculo !== undefined && (
                <button
                  type="button"
                  className="mej-calc-btn"
                  aria-expanded={abierto}
                  aria-controls={`${calcId}-calc`}
                  onClick={() => setAbierto((a) => !a)}
                >
                  {abierto ? "ocultar el cálculo" : "ver el cálculo"}
                </button>
              )}
            </dd>
          </>
        ) : punto.patron !== undefined ? (
          <>
            <dt>Patrón</dt>
            <dd>
              {punto.patron}
              {punto.calculo !== undefined && (
                <button
                  type="button"
                  className="mej-calc-btn"
                  aria-expanded={abierto}
                  aria-controls={`${calcId}-calc`}
                  onClick={() => setAbierto((a) => !a)}
                >
                  {abierto ? "ocultar el cálculo" : "ver el cálculo"}
                </button>
              )}
            </dd>
          </>
        ) : null}

        <dt>Confianza</dt>
        <dd>{confianza}</dd>

        {/* RF-252 — la fila `Sesgo` NUNCA se omite. Si no se identificó ninguno, lo DICE: una
            fila ausente se lee como «no hay sesgo», que es una afirmación distinta de «no
            buscamos». La dirección va en negrita porque es lo que cambia la lectura. */}
        <dt>Sesgo</dt>
        <dd className="mej-sesgo">
          {punto.sesgo === null ? (
            "No se identificó ningún supuesto que sesgue este cálculo."
          ) : (
            <SesgoConDireccion texto={punto.sesgo} direccion={punto.direccion_sesgo} />
          )}
        </dd>
      </dl>

      {/* H-4 — «ver el cálculo» es un despliegue EN LÍNEA, no un modal: un modal para auditar
          un número te saca de la tarjeta que estabas leyendo. */}
      {punto.calculo !== undefined && (
        <div className="mej-calc" id={`${calcId}-calc`} hidden={!abierto}>
          <span className="mono">{punto.calculo}</span>
        </div>
      )}

      <footer className="mej-fix-row">
        <p className="mej-fix">
          {punto.fix_codigo !== undefined ? partirEnCodigo(punto.fix, punto.fix_codigo) : punto.fix}
        </p>
        <span className="mej-score">{`score v${punto.score_version}`}</span>
        <button
          type="button"
          className="mej-btn"
          onClick={() => {
            setDescartado(true)
            onDescartar(punto.id)
          }}
        >
          Descartar
        </button>
        <button
          type="button"
          className="mej-btn mej-btn-primary"
          disabled={proponerDeshabilitado !== undefined}
          title={proponerDeshabilitado}
          onClick={() => onProponer({ puntoId: punto.id, textoPropuesto: punto.fix })}
        >
          Proponerlo en el chat
        </button>
      </footer>

      {/* El motivo del deshabilitado también en TEXTO: un `title` no llega por teclado ni por
          touch, y un botón muerto sin explicación se lee como un bug. */}
      {proponerDeshabilitado !== undefined && <p className="mej-mut">{proponerDeshabilitado}</p>}

      <div className="mej-live" aria-live="polite">
        {descartado && "Descartado. Se puede volver a mostrar desde la tab Mejora de esa caja."}
      </div>
    </article>
  )
}

/** La dirección del sesgo va en `<strong>`: es la palabra que decide si el número es piso o techo. */
function SesgoConDireccion({ texto, direccion }: { texto: string; direccion: string }) {
  const i = texto.indexOf(direccion)
  if (i < 0) return <>{texto}</>
  return (
    <>
      {texto.slice(0, i)}
      <strong>{direccion}</strong>
      {texto.slice(i + direccion.length)}
    </>
  )
}

/** El ajuste concreto va en `<code>`: es lo que se copia y se pega, no prosa. */
function partirEnCodigo(fix: string, codigo: string) {
  const i = fix.indexOf(codigo)
  if (i < 0) return fix
  return (
    <>
      {fix.slice(0, i)}
      <code>{codigo}</code>
      {fix.slice(i + codigo.length)}
    </>
  )
}
