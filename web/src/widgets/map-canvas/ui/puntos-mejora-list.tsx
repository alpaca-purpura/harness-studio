import type { EstadoDetector, PuntoMejora } from "@/entities/telemetria"
import { ErrorBody, Skeleton } from "@/shared/ui/estado-carga"
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

export interface PuntosMejoraListProps {
  estado?: "datos" | "cargando" | "error" | undefined
  puntos: readonly PuntoMejora[]
  /** Los detectores que SÍ corrieron. Alimentan el vacío honesto de H-2. */
  detectores?: readonly EstadoDetector[] | undefined
  /** Denominador del vacío: «los seis detectores corrieron sobre N corridas». */
  corridas?: number | undefined
  /** La caja seleccionada en el canvas: su tarjeta se resalta. */
  cajaSeleccionada?: string | undefined
  /** El arnés está fuera del alcance del chat embebido (guardrail vigente). */
  proponerDeshabilitado?: string | undefined
  onDescartar: (puntoId: string) => void
  onProponer: (p: { puntoId: string; textoPropuesto: string }) => void
  onReintentar: () => void
  error?: string | undefined
}

export function PuntosMejoraList({
  estado = "datos",
  puntos,
  detectores,
  corridas,
  cajaSeleccionada,
  proponerDeshabilitado,
  onDescartar,
  onProponer,
  onReintentar,
  error,
}: PuntosMejoraListProps) {
  // A4 — el filtro es la regla, no una optimización: sin contrafactual computable no hay tarjeta.
  const cotizables = puntos
    .filter((p) => p.contrafactual !== null && p.contrafactual !== "")
    .slice()
    .sort((a, b) => b.diferencia_micros - a.diferencia_micros)

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
        // H-2 — **nunca una sección vacía**. «No encontramos nada» y «no buscamos» son
        // conclusiones opuestas, y la única forma de distinguirlas es listar quién corrió.
        <div className="mej-vacia">
          <p className="mej-vacia-titulo">Hay datos y ningún punto de mejora que pase el corte.</p>
          <p className="mej-mut">
            {`Los seis detectores corrieron sobre ${corridas ?? 0} corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.`}
          </p>
          {detectores && detectores.length > 0 && (
            <ul className="mej-vacia-detectores">
              {detectores.map((d) => (
                <li key={d.detector}>{`${d.nombre} · sin hallazgos`}</li>
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
