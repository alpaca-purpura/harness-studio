import { useEffect, useRef, useState } from "react"
import type { Conversacion, SessionStatus } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { CtxChip, IdentidadDetalle } from "./ctx-chip"

// ConversacionRow (RF-326/RF-331 · CV-D9/CV-D14 · design.md §3.2/§4.1) — la fila 2 del cromo:
// `▶ título · ✎ · chip de ctx · 🔍 · ＋`, con renombrado en línea.
//
// Props PURAS: cero transporte. El cableado vive en `ChatDock`.

// motivoBloqueo dice POR QUÉ no se puede crear ni retomar ahora mismo, o `undefined` cuando
// sí se puede. Vive acá y no en cada consumidor porque el panel (tramo 4) muestra el mismo
// motivo en sus filas: dos copias del texto serían dos textos que se separan.
//
// El servidor re-evalúa el guard y responde 409 (`transicionLocked`, paso 1): esto no lo
// reemplaza, lo anticipa. Un control que se ve legal y después falla es peor que uno apagado
// que dice por qué (C-1, patrón sancionado en `new-session-picker.tsx:362-363`).
export function motivoBloqueo(status: SessionStatus): string | undefined {
  if (status === "streaming") return "esperá a que termine el turno en vuelo"
  if (status === "await") return "esperá tu decisión de permiso"
  return undefined
}

export interface ConversacionRowProps {
  /** la conversación activa. Siempre existe (BR-CV-1). */
  activa: Conversacion
  /** el arnés y el cwd salen de la SESIÓN, no de la conversación (RF-300). */
  arnes: string
  cwd?: string | undefined
  /** `streaming`|`await` ⇒ ＋ queda deshabilitado CON MOTIVO (RF-312). */
  status: SessionStatus
  /** la lista está desplegada ⇒ el chevron rota 90° y `aria-expanded` lo dice. */
  listaAbierta: boolean
  /** el detalle de identidad se abre solo al retomar y al rotar (RF-314). */
  detalleForzado: boolean
  /** el id del panel que el botón del título controla (`aria-controls`). */
  panelId: string
  onToggleLista: (foco: "filas" | "buscador") => void
  onNueva: () => void
  onRenombrar: (titulo: string) => void
}

export function ConversacionRow({
  activa,
  arnes,
  cwd,
  status,
  listaAbierta,
  detalleForzado,
  panelId,
  onToggleLista,
  onNueva,
  onRenombrar,
}: ConversacionRowProps) {
  const [editando, setEditando] = useState(false)
  const [valor, setValor] = useState(activa.titulo)
  const [devolverFoco, setDevolverFoco] = useState(false)
  const [detalleAbierto, setDetalleAbierto] = useState(false)
  const tituloRef = useRef<HTMLButtonElement>(null)
  const detalleId = `${panelId}-detalle`

  // RF-314: al retomar y al rotar, el cc-id de la conversación CAMBIA. El detalle se abre
  // solo para que eso se vea sin buscarlo. Sólo en el flanco: si el operador lo cierra a
  // mano, no se le vuelve a abrir hasta la próxima transición.
  useEffect(() => {
    if (detalleForzado) setDetalleAbierto(true)
  }, [detalleForzado])

  // Escape devuelve el foco al botón del título (design.md §8.3). Va en un efecto y no en el
  // handler porque el botón todavía no existe cuando el handler corre: el input se desmonta
  // en el mismo render en que el botón se monta.
  useEffect(() => {
    if (devolverFoco && !editando) {
      setDevolverFoco(false)
      tituloRef.current?.focus()
    }
  }, [devolverFoco, editando])

  // Calcado de `session-rail.tsx:196-201`: confirma sólo si hay texto y cambió; vacío
  // DESCARTA (nunca borra el título, RF-331 CA-3).
  const confirmar = () => {
    setEditando(false)
    const t = valor.trim()
    if (t && t !== activa.titulo) onRenombrar(t)
    else setValor(activa.titulo)
  }

  const bloqueo = motivoBloqueo(status)

  return (
    <>
      <div className="flex flex-none items-center gap-1.5 border-b border-border py-1 pr-3 pl-2.5">
        {editando ? (
          // El input toma la FILA ENTERA: el ctx y las acciones no compiten con el cursor
          // (RF-331 CA-2, mockup:466).
          <input
            autoFocus
            aria-label="Título de la conversación"
            value={valor}
            onChange={(e) => setValor(e.target.value)}
            onBlur={confirmar}
            onKeyDown={(e) => {
              if (e.key === "Enter") confirmar()
              if (e.key === "Escape") {
                setValor(activa.titulo)
                setEditando(false)
                setDevolverFoco(true)
              }
            }}
            className="w-full min-w-0 rounded-sm border border-primary bg-card px-1.5 py-px text-xs font-semibold text-foreground focus:outline-none"
          />
        ) : (
          <>
            <button
              ref={tituloRef}
              type="button"
              onClick={() => onToggleLista("filas")}
              aria-expanded={listaAbierta}
              aria-controls={panelId}
              title={activa.titulo}
              className="flex min-w-0 flex-1 items-center gap-1.5 rounded-sm px-1.5 py-0.5 text-left text-xs font-semibold text-foreground hover:bg-secondary focus:outline-2 focus:outline-primary"
            >
              <span
                aria-hidden
                className={cn(
                  "flex-none text-[8px] text-muted-foreground transition-transform",
                  listaAbierta && "rotate-90",
                )}
              >
                ▶
              </span>
              <span className="truncate">{activa.titulo}</span>
            </button>

            {/* ✎ · el disparador del renombrado. El dibujo describe el gesto (§2D: «✎ →
                input → Enter confirma») pero no dibuja el botón; el rail lo esconde hasta el
                hover (`group-hover:flex`), que lo deja fuera del alcance del teclado. Acá es
                un control visible más, del mismo tamaño que los otros dos. */}
            <button
              type="button"
              aria-label="Renombrar la conversación"
              title="Renombrar la conversación"
              onClick={() => {
                setValor(activa.titulo)
                setEditando(true)
              }}
              className="grid size-6 flex-none place-items-center rounded-sm text-[11px] text-muted-foreground hover:bg-secondary hover:text-foreground focus:outline-2 focus:outline-primary"
            >
              ✎
            </button>

            <CtxChip
              ctxPct={activa.ctx_pct}
              claudeSessionId={activa.claude_session_id}
              arnes={arnes}
              model={activa.model}
              cwd={cwd}
              caliente={activa.rotacion_pendiente}
              abierto={detalleAbierto}
              onToggle={() => setDetalleAbierto((v) => !v)}
              detalleId={detalleId}
            />

            <span className="flex flex-none items-center gap-0.5">
              <button
                type="button"
                aria-label="Buscar en las conversaciones de esta sesión"
                title="Buscar en las conversaciones de esta sesión"
                onClick={() => onToggleLista("buscador")}
                className="grid size-6 flex-none place-items-center rounded-sm text-[11px] text-muted-foreground hover:bg-secondary hover:text-foreground focus:outline-2 focus:outline-primary"
              >
                🔍
              </button>
              {/* ＋ en `--foreground`, NO en `--primary`: medido, 2,51:1 sobre `--card` en
                  tema claro contra el mínimo no-textual de 3:1. El acento entra por el hover
                  (`--accent-soft`), no por el color solo. */}
              <button
                type="button"
                aria-label="Nueva conversación — desactiva la actual"
                title={bloqueo ?? "Nueva conversación — desactiva la actual"}
                disabled={bloqueo !== undefined}
                onClick={onNueva}
                className="grid size-6 flex-none place-items-center rounded-sm text-sm text-foreground hover:bg-accent-soft focus:outline-2 focus:outline-primary disabled:cursor-not-allowed disabled:opacity-40"
              >
                ＋
              </button>
            </span>
          </>
        )}
      </div>

      {detalleAbierto && (
        <IdentidadDetalle
          id={detalleId}
          claudeSessionId={activa.claude_session_id}
          arnes={arnes}
          model={activa.model}
          cwd={cwd}
        />
      )}
    </>
  )
}
