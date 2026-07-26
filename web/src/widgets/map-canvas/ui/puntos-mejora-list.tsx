import type { EstadoDetector, PuntoMejora } from "@/entities/telemetria"
import { ErrorBody, Skeleton } from "@/shared/ui/estado-carga"
import { DETECTORES_DEL_MVP } from "../model/capa-mejora"
import { PuntoMejoraCard } from "./punto-mejora-card"

// PuntosMejoraList — resuelve **H-1**: dónde vive la tarjeta de un punto de mejora.
//
// **Debajo del canvas, como contenido de la página.** No es un panel flotante ni un modal, y
// **no puede ser un botón dentro del nodo**: el nodo YA es un `<button>` y anidar un botón en
// un botón es DOM inválido (H-1/J-9). Seleccionar una caja resalta su tarjeta y la trae a la
// vista; la relación se dice con TEXTO (`↔ caja seleccionada`), no con un borde más grueso.
//
// **Orden: ahorro contrafactual DESCENDENTE.** Es lo único que ordena sin discutir. El wire ya
// viene ordenado (`telemetria_mejoras.go:125`), y acá se reordena igual: la lista no puede
// depender de que el productor no cambie de opinión.
//
// **Una tarjeta sin contrafactual no existe** (A4): se filtra, no se pinta degradada. El
// hallazgo no se pierde — el inspector lo lista como «sin fix propuesto» (T34). Esconderlo sería
// el gap invisible; pintarlo a medias sería violar la regla que el paquete entero defiende.
//
// 🔴 **`hayDatos` NO es opcional, y ese es el punto.** Sin corridas atribuibles esta sección
// **no se dibuja** (design.md §7.4: «no se dibuja: manda el estado 1 de la franja»). Nació
// opcional-con-default y produjo el defecto que la verificación en la app instalada cazó: sobre
// un arnés que nunca corrió, la sección afirmaba «Hay datos y ningún punto de mejora que pase el
// corte. Los seis detectores corrieron sobre 0 corridas», contradiciendo a la franja de arriba.
// Un default que asume que hay datos es un default que miente cuando no los hay.

export interface PuntosMejoraListProps {
  estado?: "datos" | "cargando" | "error" | undefined
  puntos: readonly PuntoMejora[]
  /**
   * ¿Hubo medición atribuible en la ventana? Lo decide `hayDatosAtribuibles()`
   * (`../model/capa-mejora`), la MISMA función que usa el composition-root — para que la franja
   * y esta sección no puedan contradecirse. `false` ⇒ esta sección no se dibuja.
   *
   * **Obligatorio a propósito**: ver el comentario de cabecera.
   */
  hayDatos: boolean
  /**
   * Los detectores del MVP que **NO pudieron correr**, con su motivo (`no_aplican` del wire).
   * Sin esto, el vacío afirmaría que corrieron los seis incluso en `s2-degradado`, donde B1 no
   * puede correr por falta del `result` del stream-json — la misma mentira, otro escenario.
   */
  noAplican?: readonly EstadoDetector[] | undefined
  /** Denominador del vacío: «los seis detectores corrieron sobre N corridas». */
  corridas?: number | undefined
  /** La caja seleccionada en el canvas: su tarjeta se resalta. */
  cajaSeleccionada?: string | undefined
  /** El arnés está fuera del alcance del chat embebido (guardrail vigente). */
  proponerDeshabilitado?: string | undefined
  /** Opcionales (A-1): sin handler los botones nacen deshabilitados con su motivo. */
  onDescartar?: ((puntoId: string) => void) | undefined
  onProponer?: ((p: { puntoId: string; textoPropuesto: string }) => void) | undefined
  onReintentar: () => void
  error?: string | undefined
}

export function PuntosMejoraList({
  estado = "datos",
  puntos,
  hayDatos,
  noAplican,
  corridas,
  cajaSeleccionada,
  proponerDeshabilitado,
  onDescartar,
  onProponer,
  onReintentar,
  error,
}: PuntosMejoraListProps) {
  // 🔴 Sin medición atribuible, la sección NO se dibuja: manda el estado 1 de la franja
  // (design.md §7.4). Dibujar acá cualquier vacío sería una segunda afirmación sobre el mismo
  // hecho, y la que se leería primero es la de abajo.
  //
  // El estado de transporte gana igual: «no pude preguntar» hay que decirlo aunque todavía no
  // sepamos si hay datos.
  if (!hayDatos && estado === "datos") return null

  // A4 — el filtro es la regla, no una optimización: sin contrafactual computable no hay tarjeta.
  const cotizables = puntos
    .filter((p) => p.contrafactual !== null && p.contrafactual !== "")
    .slice()
    .sort((a, b) => b.diferencia_micros - a.diferencia_micros)

  // Cuántos llegaron a correr. El wire no manda la lista de los que salieron limpios (hueco
  // declarado en PARIDAD §4), pero sí `no_aplican` — así que el CONTEO sí es derivable, y es
  // lo que la frase necesita para no exagerar.
  const corrieron = DETECTORES_DEL_MVP - (noAplican?.length ?? 0)

  return (
    <section className="arnesia-mejora mej-lista" aria-label="Puntos de mejora">
      <header className="mej-lista-hd">
        <h3>Puntos de mejora</h3>
        {estado === "datos" && <span className="mej-lista-n">{cotizables.length}</span>}
      </header>

      {estado === "cargando" ? (
        <Skeleton label="Buscando puntos de mejora" filas={2} altura={148} data="mejora" />
      ) : estado === "error" ? (
        <ErrorBody
          motivo={`No se pudieron calcular los puntos de mejora — ${error ?? "motivo desconocido"}.`}
          onReintentar={onReintentar}
        />
      ) : cotizables.length === 0 ? (
        // H-2 — **nunca una sección vacía**, pero tampoco una que exagere la búsqueda.
        // «No encontramos nada» y «no buscamos» son conclusiones opuestas, y acá se dice
        // exactamente CUÁNTOS detectores llegaron a correr: afirmar los seis cuando B1 no pudo
        // (s2-degradado) es la misma mentira que afirmarlos sobre 0 corridas.
        <div className="mej-vacia">
          <p className="mej-vacia-titulo">Hay datos y ningún punto de mejora que pase el corte.</p>
          <p className="mej-mut">
            {corrieron === DETECTORES_DEL_MVP
              ? `Los seis detectores corrieron sobre ${corridas ?? 0} corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.`
              : `${corrieron} de los seis detectores corrieron sobre ${corridas ?? 0} corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.`}
          </p>
          {noAplican && noAplican.length > 0 && (
            <ul className="mej-vacia-detectores">
              {noAplican.map((d) => (
                <li
                  key={d.detector}
                >{`${d.nombre} — no disponible: ${d.motivo ?? "sin motivo declarado"}`}</li>
              ))}
            </ul>
          )}
        </div>
      ) : (
        <div className="mej-lista-body">
          {cotizables.map((p) => (
            <PuntoMejoraCard
              key={p.id}
              punto={p}
              resaltada={cajaSeleccionada !== undefined && p.caja_id === cajaSeleccionada}
              proponerDeshabilitado={proponerDeshabilitado}
              onDescartar={onDescartar}
              onProponer={onProponer}
            />
          ))}
        </div>
      )}

      {/* RF-255 · BR-M12 — la nota al pie es parte del contrato con el operador, no relleno:
          dice qué hace el botón ANTES de que lo aprete. */}
      <p className="mej-nota">
        «Proponerlo en el chat» no escribe archivos: abre el chat con el cambio propuesto, y se
        aplica por el camino de siempre, con sus permisos y su gate.
      </p>
    </section>
  )
}
