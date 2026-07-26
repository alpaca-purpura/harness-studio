import type { ReactNode } from "react"
import type { ArtefactosMode, Graph } from "@/entities/arnes"
import type {
  CifraCaja,
  EstadoDetector,
  PuntoMejora,
  ResumenTelemetria,
  Ventana,
} from "@/entities/telemetria"
import { vistaCapaMejora } from "../model/capa-mejora"
import { FranjaMejora } from "./franja-mejora"
import { MapCanvas } from "./map-canvas"
import { PuntosMejoraList } from "./puntos-mejora-list"

// CapaMejoraStage — **la composición de la capa «Mejora», con stories**.
//
// 🔴 Existe por una razón de método, no de estética. Los cuatro críticos de la auditoría del
// Tramo B son de la misma clase que D24 —bloques defendibles por separado que juntos afirman
// algo falso— y **los cuatro vivían en `pages/shell/ui/workspace-stage.tsx`**, que es el único
// lugar del repo donde se componen las cinco superficies que hablan del mismo hecho y el único
// que **no tiene stories**. El candado de D24 no los vio porque montaba dos bloques de los cinco.
//
// Mover la composición acá la vuelve **testeable**: este widget recibe los objetos crudos del
// wire, deriva TODO con `vistaCapaMejora()` (una función pura, unit-testeada) y decide qué
// muestra cada bloque. La página queda con lo suyo —transporte y estado— y ya no puede
// contradecirse a sí misma, porque no toma ninguna de estas decisiones.
//
// Los dos invariantes de layout que la auditoría midió rotos (C-1):
//  · **el canvas conserva su altura**: la capa es un superset del Mapa (RF-245 · BR-M16), así que
//    el canvas lleva un piso propio y la lista NO puede comérselo por ser hermana flex;
//  · **la lista scrollea dentro de sí misma**, acotada, en vez de crecer sin límite. Medido antes
//    del fix: 739 px → 257 px con una tarjeta → **0 px con cuatro**.

export interface CapaMejoraStageProps {
  capa: "estructura" | "mejora" | "perf" | "proceso"
  graph: Graph | null
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
  artefactos?: ArtefactosMode | undefined

  // ── Datos CRUDOS del wire. El stage deriva; la página no interpreta. ──
  estado?: "datos" | "cargando" | "error" | undefined
  resumen: ResumenTelemetria | null
  cajas: readonly CifraCaja[]
  puntos: readonly PuntoMejora[]
  noAplican: readonly EstadoDetector[]
  error?: string | undefined

  ventana: Ventana
  onVentana: (v: Ventana) => void
  /** Retención y reenvío salen de `GET /api/telemetria/salud`; sin ese dato NO se inventan. */
  retencionDias?: number | undefined
  retencionPropuesta?: boolean | undefined
  forwardDestino?: string | undefined
  onPolitica: () => void
  onReintentar: () => void
  onDescartar: (puntoId: string) => void
  onProponer: (p: { puntoId: string; textoPropuesto: string }) => void
  /** Motivo por el que no se puede proponer (guardrail de alcance del chat, CH-D6). */
  proponerDeshabilitado?: string | undefined

  /** El drawer del Slice anterior. Es estructura, no capa: entra como slot. */
  inspector?: ReactNode | undefined
  /** Estados de carga/error del GRAFO (no de la telemetría), que la página ya resuelve. */
  cuerpoAlternativo?: ReactNode | undefined
}

export function CapaMejoraStage({
  capa,
  graph,
  selectedId,
  onSelect,
  artefactos,
  estado = "datos",
  resumen,
  cajas,
  puntos,
  noAplican,
  error,
  ventana,
  onVentana,
  retencionDias,
  retencionPropuesta,
  forwardDestino,
  onPolitica,
  onReintentar,
  onDescartar,
  onProponer,
  proponerDeshabilitado,
  inspector,
  cuerpoAlternativo,
}: CapaMejoraStageProps) {
  const activa = capa === "mejora"
  // TODA la derivación, en una función pura y unit-testeada. Que la página no pueda tomar
  // ninguna de estas decisiones es el punto del refactor.
  const vista = vistaCapaMejora({ resumen, cajas, graph, activa })

  return (
    <div className="flex h-full min-h-0 flex-col">
      {activa && (
        <FranjaMejora
          estado={estado}
          ventana={ventana}
          onVentana={onVentana}
          resumen={vista.resumenParaFranja}
          escenario={resumen?.escenario}
          retencionDias={retencionDias}
          retencionPropuesta={retencionPropuesta}
          forwardDestino={forwardDestino}
          onPolitica={onPolitica}
          onReintentar={onReintentar}
          error={error}
        />
      )}

      {/* C-1 · el canvas lleva su PISO. `min-h-0` solo (lo que había) le dice a flexbox que
          puede encogerlo sin límite, y la lista se lo comía hasta 0 px. */}
      <div className="mapa-canvas-slot relative min-h-0 flex-1">
        {cuerpoAlternativo ??
          (graph === null ? null : (
            <>
              <MapCanvas
                graph={graph}
                selectedId={selectedId}
                onSelect={onSelect}
                artefactos={artefactos}
                capa={capa}
                mejora={vista.mejoraPorNodo}
                totalesPorFase={vista.totalesPorFase}
                motivosSinDato={vista.motivosPorNodo}
              />
              {inspector}
            </>
          ))}
      </div>

      {/* H-1 — la lista vive DEBAJO del canvas, como contenido de la página… pero **acotada**:
          es un superset del Mapa, no un reemplazo. */}
      {activa && (
        <div className="mapa-lista-slot">
          <PuntosMejoraList
            estado={estado}
            puntos={puntos}
            hayDatos={vista.hayDatos}
            noAplican={noAplican}
            corridas={vista.denominadorDeBusqueda}
            cajaSeleccionada={selectedId}
            proponerDeshabilitado={proponerDeshabilitado}
            onDescartar={onDescartar}
            onProponer={onProponer}
            onReintentar={onReintentar}
            error={error}
          />
        </div>
      )}
    </div>
  )
}
