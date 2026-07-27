import type { ReactNode } from "react"
import type { ArtefactosMode, Graph } from "@/entities/arnes"
import type {
  CifraCaja,
  EstadoDetector,
  PuntoMejora,
  ResumenTelemetria,
  Ventana,
} from "@/entities/telemetria"
import { estadosDeLaFranja, vistaCapaMejora } from "../model/capa-mejora"
import { bucketsDe, type DetalleCajaWire, ETIQUETA_VENTANA } from "../model/detalle-caja"
import { FranjaMejora } from "./franja-mejora"
import { InspectorMejora } from "./inspector-mejora"
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
  forwardDestino?: string | undefined
  /** Cuándo se refrescó el catálogo de precios, de `/salud`. `null` = NUNCA (estado 5). */
  catalogoRefrescado?: string | null | undefined
  /** Con el daemon caído se conservan las cifras previas, marcadas como posiblemente viejas
   *  (estado B3). La página lo sabe porque la consulta falló sin respuesta, no por adivinanza. */
  daemonCaido?: boolean | undefined
  fechaUltimaMedicion?: string | undefined
  onPolitica: () => void
  onReintentar: () => void
  /** Opcionales (A-1): sin handler los botones nacen deshabilitados con su motivo, en vez de
   *  fingir que funcionan. Hoy la página no puede cablearlos. */
  onDescartar?: ((puntoId: string) => void) | undefined
  onProponer?: ((p: { puntoId: string; textoPropuesto: string }) => void) | undefined
  /** Motivo por el que no se puede proponer (guardrail de alcance del chat, CH-D6). */
  proponerDeshabilitado?: string | undefined

  // ── 4ª tab del inspector. El CUERPO lo arma este widget (C-3): que lo armara la página es
  // cómo un GET fallido terminó pintándose como dato. ──
  /** El nodo seleccionado, si lo hay: decide si la tab muestra tabla o el motivo de «no es caja». */
  cajaSeleccionada?: { esCaja: boolean; motivoNoCaja: string } | undefined
  detalle?: DetalleCajaWire | null | undefined
  /** Motivo REAL del fallo del detalle. Presente ⇒ la tab muestra su estado de ERROR. */
  detalleError?: string | undefined
  noMedidos?: readonly EstadoDetector[] | undefined
  /**
   * El drawer del Slice anterior. Es estructura, no capa, así que lo monta la página — pero
   * recibe el cuerpo de la 4ª tab YA decidido por este widget.
   */
  inspector?: ((cuerpoMejora: ReactNode | undefined) => ReactNode) | undefined
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
  forwardDestino,
  catalogoRefrescado,
  daemonCaido,
  fechaUltimaMedicion,
  onPolitica,
  onReintentar,
  onDescartar,
  onProponer,
  proponerDeshabilitado,
  cajaSeleccionada,
  detalle,
  detalleError,
  noMedidos,
  inspector,
  cuerpoAlternativo,
}: CapaMejoraStageProps) {
  const activa = capa === "mejora"
  // TODA la derivación, en una función pura y unit-testeada. Que la página no pueda tomar
  // ninguna de estas decisiones es el punto del refactor.
  const vista = vistaCapaMejora({ resumen, cajas, graph, activa })

  // 🔴 C-3 · un GET que falla es estado de TRANSPORTE, no dato. Tragarlo hacía que la 4ª tab
  // afirmara ocho cosas falsas —«0 corridas», seis «no aplica en este runtime» y «catálogo sin
  // construir»— con la nota «"No aplica" no es 0» desplegada EN DEFENSA de la mentira.
  const cuerpoMejora =
    activa && cajaSeleccionada ? (
      <InspectorMejora
        estado={detalleError !== undefined ? "error" : "datos"}
        error={detalleError}
        esCaja={cajaSeleccionada.esCaja}
        motivoNoCaja={cajaSeleccionada.motivoNoCaja}
        ventanaLabel={ETIQUETA_VENTANA[ventana]}
        corridas={detalle?.turnos_totales ?? 0}
        buckets={bucketsDe(detalle ?? null)}
        totalMicros={detalle?.paridad?.reportado_micros ?? null}
        paridad={
          detalle?.paridad ?? {
            reportado_micros: null,
            calculado_micros: null,
            divergencia_pct: null,
            completo: false,
            catalogo_sin_construir: true,
          }
        }
        join={{
          corridas: detalle?.turnos_totales ?? 0,
          // `null`, no 0: sin señal de gate, «ninguna se rechazó» sería una afirmación sobre el
          // proceso que nadie midió (T22 sigue abierto).
          rechazadas: null,
          costo_rechazadas_micros: null,
          rotaciones: null,
        }}
        detectores={detalle?.detectores ?? []}
        noMedidos={noMedidos ?? []}
        onReintentar={onReintentar}
      />
    ) : undefined

  return (
    <div className="flex h-full min-h-0 flex-col">
      {activa && (
        <FranjaMejora
          estado={daemonCaido === true ? "daemon-caido" : estado}
          ventana={ventana}
          onVentana={onVentana}
          resumen={vista.resumenParaFranja}
          escenario={resumen?.escenario}
          retencionDias={retencionDias}
          forwardDestino={forwardDestino}
          onPolitica={onPolitica}
          onReintentar={onReintentar}
          error={error}
          fechaUltimaMedicion={fechaUltimaMedicion}
          {...estadosDeLaFranja(resumen, catalogoRefrescado)}
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
              {inspector?.(cuerpoMejora)}
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
