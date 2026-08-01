import type { CSSProperties } from "react"
import {
  type Actividad,
  cajasSinActividad,
  type Graph,
  type SaludActividad,
  saludDeActividad,
  saludDeGrupo,
  selectActividades,
} from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"

// ActividadChips — la fila N0 del Mapa multi-actividad (MA-T4, spec §3): un chip por tipo de
// paquete declarado (dot de salud worst-of · id mono · N cajas) + el chip `sin-actividad`
// (E6, honestidad: jamás se oculta) + el degradado E7 con su CTA al chat. Es CHROME que el
// dueño del estado (workspace-stage / la story) monta BAJO la MapBar — mismo patrón
// prop+callback que `capa`/`artefactos`; el componente no guarda foco propio.
//
// MA-L5/E8: sin tipos declarados devuelve `null` — la fila NO existe, cero DOM nuevo, el
// Mapa queda EXACTAMENTE como hoy. Estilos en map.css, scoped `.arnesia-actividades`.

type DotStyle = CSSProperties & { "--sd"?: string }

export interface ActividadChipsProps {
  graph: Graph
  /** Actividad focada: un id del catálogo o `"sin-actividad"`. `undefined` = panorama N0. */
  actividadFoco?: string | undefined
  /** Focar/des-focar (toggle). `null` = volver al panorama (MA-L1). */
  onActividadFoco: (id: string | null) => void
  /** Pre-resaltado por hover (solo sin foco, mockup:616): atenúa sin entrar al foco N1. */
  onActividadPre?: ((id: string | null) => void) | undefined
  /** CTA del degradado E7 («forjar el procedimiento conversando»). Sin handler no se muestra. */
  onForjarConversando?: (() => void) | undefined
}

// Las cajas DISTINTAS que el procedimiento referencia (los pasos sin caja no cuentan: E13
// los hace visibles en el foco, no en el conteo).
function cajasDelProcedimiento(a: Actividad): number {
  return new Set((a.pasos ?? []).filter((p) => p.caja !== undefined).map((p) => p.caja)).size
}

function tituloDe(a: Actividad): string | undefined {
  const partes: string[] = []
  if (a.estados?.length) partes.push(`estados: ${a.estados.join(" → ")}`)
  if (a.cierre !== undefined) partes.push(`cierre: ${a.cierre}`)
  return partes.length > 0 ? partes.join(" · ") : undefined
}

function Dot({ salud }: { salud: SaludActividad }) {
  return (
    <span
      className="dot"
      style={{ "--sd": `var(--${salud})` } as DotStyle}
      title={`salud ${salud} = worst-of de hallazgos existentes`}
    />
  )
}

export function ActividadChips({
  graph,
  actividadFoco,
  onActividadFoco,
  onActividadPre,
  onForjarConversando,
}: ActividadChipsProps) {
  const actividades = selectActividades(graph)
  // MA-L5/E8 — arnés sin tipos declarados: la fila no existe. Cero selector, cero invento.
  if (actividades.length === 0) return null

  const sinActividad = cajasSinActividad(graph)
  const hayE7 = actividades.some((a) => (a.pasos ?? []).length === 0)
  const toggle = (id: string) => onActividadFoco(actividadFoco === id ? null : id)

  return (
    <div className="arnesia-actividades" role="group" aria-label="Actividades del arnés">
      <span className="lbl">actividades</span>
      {actividades.map((a) => {
        const sinProcedimiento = (a.pasos ?? []).length === 0
        if (sinProcedimiento) {
          // E7 — degradado honesto: dashed + italic, NO focable (disabled).
          return (
            <button
              key={a.id}
              type="button"
              className="achip degradado"
              disabled
              aria-pressed={false}
              title="Actividad declarada en tipos_paquete sin pasos — MA-L5: degradado honesto"
            >
              <span className="nm">{a.id}</span>
              <span className="n">sin procedimiento aún (E7)</span>
            </button>
          )
        }
        const n = cajasDelProcedimiento(a)
        return (
          <button
            key={a.id}
            type="button"
            className="achip"
            aria-pressed={actividadFoco === a.id}
            title={tituloDe(a)}
            onClick={() => toggle(a.id)}
            onMouseEnter={() => {
              if (actividadFoco === undefined) onActividadPre?.(a.id)
            }}
            onMouseLeave={() => {
              if (actividadFoco === undefined) onActividadPre?.(null)
            }}
          >
            <Dot salud={saludDeActividad(graph, a)} />
            <span className="nm">{a.id}</span>
            <span className="n">
              {n} caja{n === 1 ? "" : "s"}
            </span>
          </button>
        )
      })}
      {sinActividad.length > 0 && (
        // E6 — el grupo sin-actividad es VISIBLE y focable (insumo de poda §6).
        <button
          type="button"
          className={cn("achip", "sinact")}
          aria-pressed={actividadFoco === "sin-actividad"}
          title="E6 — cajas que ningún procedimiento referencia: candidatas a declarar o podar"
          onClick={() => toggle("sin-actividad")}
        >
          <Dot salud={saludDeGrupo(sinActividad)} />
          <span className="nm">sin-actividad</span>
          <span className="n">
            {sinActividad.length} caja{sinActividad.length === 1 ? "" : "s"}
          </span>
        </button>
      )}
      {hayE7 && onForjarConversando !== undefined && (
        <button
          type="button"
          className="cta-chat"
          title="Abre el chat para forjar el procedimiento de la actividad sin pasos (E7)"
          onClick={onForjarConversando}
        >
          ⌨ forjar el procedimiento conversando
        </button>
      )}
    </div>
  )
}
