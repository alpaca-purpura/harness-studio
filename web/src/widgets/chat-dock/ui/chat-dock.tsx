import { useEffect, useId, useRef, useState } from "react"
import type { Resultado, Session, Turn } from "@/shared"
import { selectActive, selectPendingPerms, selectScope, useAppStore, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { Pip } from "@/shared/ui/indicators"
import { useConversaciones } from "../model/conversaciones-store"
import { ConversacionRow, motivoBloqueo } from "./conversacion-row"
import { ConversacionesPanel } from "./conversaciones-panel"
import { DictadoAviso, DictadoButton, VoiceBar } from "./dictado-button"
import { Md } from "./markdown"
import { PermissionCard } from "./permission-card"

// ChatDock is the invoked, collapsible conversation (mockup it.14): the live Claude
// Code session of the active work-front. Sending a turn streams the reply here.
export function ChatDock() {
  const active = useSessions(selectActive)
  const streaming = useSessions((s) => (active ? s.streaming[active.id] : undefined))
  const closeChat = useSessions((s) => s.closeChat)
  const detalleForzado = useSessions((s) =>
    active ? (s.detalleForzado[active.id] ?? false) : false,
  )
  // Selectores PRIMITIVOS, uno por dato: `useConversaciones()` sin selector devuelve el
  // objeto entero, cuya identidad cambia en cada `set` — y con ella se re-renderizaría el
  // transcript completo en cada tecla del buscador (arquitectura.md §6.3).
  const panelAbierta = useConversaciones((s) => s.abierta)
  const panelSesion = useConversaciones((s) => s.sesionId)
  const panelEstado = useConversaciones((s) => s.estado)
  const panelError = useConversaciones((s) => s.error)
  const panelLista = useConversaciones((s) => s.conversaciones)
  const panelTotal = useConversaciones((s) => s.total)
  const panelBusqueda = useConversaciones((s) => s.busqueda)
  const panelFoco = useConversaciones((s) => s.focoInicial)
  const retomando = useConversaciones((s) => s.retomando)
  const fallo = useConversaciones((s) => s.fallo)
  const panelId = useId()

  if (!active) return null

  const abierta = panelAbierta && panelSesion === active.id
  const acciones = useConversaciones.getState()

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-none items-center gap-2 border-b border-border px-3 py-2.5">
        <span className="flex min-w-0 items-center gap-1.5 text-xs font-semibold">
          <Pip status={active.status} />
          <span className="truncate">{active.frente}</span>
        </span>
        <button
          type="button"
          onClick={closeChat}
          title="Colapsar el dock (⌘K para reabrir)"
          className="ml-auto flex items-center gap-1.5 whitespace-nowrap rounded-md px-2 py-1 text-[11px] text-muted-foreground hover:bg-secondary hover:text-foreground"
        >
          » colapsar
        </button>
      </div>

      {/* Fila 2 — SIEMPRE. Absorbe el ctx de la `SessionLine` retirada como chip-disclosure
          (CV-D14): la identidad técnica sigue existiendo, a un clic, y con el `cwd` que hasta
          hoy no se veía en ninguna superficie. */}
      <ConversacionRow
        activa={active.activa}
        arnes={active.arnes}
        cwd={active.cwd}
        status={active.status}
        listaAbierta={abierta}
        detalleForzado={detalleForzado}
        panelId={panelId}
        onToggleLista={(foco) => (abierta ? acciones.cerrar() : acciones.abrir(active.id, foco))}
        onNueva={() => void acciones.crear()}
        onRenombrar={(t) => void acciones.renombrar(active.activa.id, t)}
      />

      {/* Franja de retoma — efímera mientras `--resume` rehidrata (RF-310), y en variante
          `bad` cuando la transición falló (C-11/RF-348). Va DEBAJO de la fila, no adentro:
          es un estado del hilo, no del control. Sin `claude_session_id` no se dibuja el
          `--resume`: no hay proceso viejo que retomar y anunciarlo sería inventarlo (E-16). */}
      {retomando !== null && (
        <div
          role="status"
          className="flex flex-none items-center gap-1.5 border-b border-border bg-accent-soft px-3 py-1 text-[10.5px] text-foreground"
        >
          <span>Retomando la conversación…</span>
          {(() => {
            const cc = panelLista.find((c) => c.id === retomando)?.claude_session_id
            return cc ? (
              <span className="ml-auto font-mono text-[9.5px] text-muted-foreground">
                --resume {cc.slice(0, 8)}
              </span>
            ) : null
          })()}
        </div>
      )}
      {fallo !== undefined && (
        <div
          role="alert"
          className="flex flex-none items-center gap-1.5 border-b border-border bg-crit-soft px-3 py-1 text-[10.5px] text-foreground"
        >
          {fallo}
        </div>
      )}

      {/* Fila 3 — SÓLO con un nodo elegido (RF-329). */}
      <ScopeRow />

      {abierta ? (
        <ConversacionesPanel
          id={panelId}
          estado={panelEstado}
          error={panelError}
          conversaciones={panelLista}
          total={panelTotal}
          frenteSesion={active.frente}
          busqueda={panelBusqueda}
          bloqueadoMotivo={motivoBloqueo(active.status)}
          focoInicial={panelFoco}
          onBusqueda={(q) => void acciones.buscar(q)}
          onElegir={(cid) => void acciones.activar(cid)}
          onCancelar={acciones.cerrar}
          onReintentar={() => void acciones.reintentar()}
        />
      ) : (
        <Messages session={active} streaming={streaming} />
      )}
      <Composer />
    </div>
  )
}

// ScopeRow (RF-110/RF-111/RF-329, mockup §2C): el chip removible del nodo seleccionado en el
// Mapa, resuelto a su archivo real.
//
// Deja de ser fila fija: SIN nodo devuelve `null` y la fila no existe en el DOM. Lo que se
// quita, declarado (CV-D14 lo autoriza): el rótulo «Alcance:», el chip punteado del arnés
// —redundante con el rail, el topbar y el placeholder del composer— y el hint del vacío, que
// gastaba una fila entera del dock en una instrucción. Con nodo, el chip es LITERAL el
// vigente, ✕ incluido.
function ScopeRow() {
  const scope = useSessions(selectScope)
  const setScope = useSessions((st) => st.setScope)
  if (!scope) return null
  return (
    <div className="flex flex-none flex-wrap items-center gap-1.5 border-b border-border px-3 py-1.5 text-[10px]">
      <span className="inline-flex min-w-0 items-center gap-1 rounded-full border border-border bg-secondary px-2 py-px">
        <span className="size-1.5 flex-none rounded-[2px] bg-skill" aria-hidden />
        {scope.clase ?? "nodo"} <b className="font-mono">{scope.nodeId}</b>
        {scope.fuentePath && (
          <span className="truncate font-mono text-muted-foreground">{scope.fuentePath}</span>
        )}
        <button
          type="button"
          title="quitar del alcance"
          onClick={() => setScope(null)}
          className="flex-none text-muted-foreground hover:text-destructive"
        >
          ✕
        </button>
      </span>
    </div>
  )
}

function Messages({ session: s, streaming }: { session: Session; streaming: string | undefined }) {
  const endRef = useRef<HTMLDivElement>(null)
  const pending = useSessions(selectPendingPerms)
  const resolvePermission = useSessions((st) => st.resolvePermission)
  const desactivada = useSessions((st) => st.desactivadaTitulo[s.id])
  const conv = s.activa.conv
  const grupos = agrupar(conv)
  const showLive = s.status === "streaming"

  // biome-ignore lint/correctness/useExhaustiveDependencies: these deps are intentional scroll triggers; the body only reads the ref.
  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" })
  }, [conv.length, streaming, showLive, pending.length])

  return (
    <div className="flex flex-1 flex-col gap-2 overflow-auto p-3">
      {conv.length === 0 && !showLive && (
        // El vacío vigente decía «Esta conversación ES la sesión Claude Code del frente …»:
        // bajo CV-D3 dejó de ser verdad —la sesión CONTIENE N conversaciones— y por eso
        // cambia. Es la única eliminación de copy firmado del paquete, autorizada por CV-D3.
        // La segunda oración sólo aparece si HAY una anterior que nombrar (§4.6 #2).
        <p className="m-auto max-w-[85%] text-center text-[11px] text-muted-foreground">
          Pídele un cambio a <b className="text-foreground">{s.arnes}</b>.
          {desactivada && <> «{desactivada}» quedó guardada — la retomás desde ▶.</>}
        </p>
      )}
      {grupos.map((g, i) =>
        g.kind === "act" ? (
          <ActivityCard
            // biome-ignore lint/suspicious/noArrayIndexKey: conv is append-only and immutable, so the index is a stable identity.
            key={i}
            pasos={g.pasos}
            live={showLive && !streaming && i === grupos.length - 1}
          />
        ) : (
          // biome-ignore lint/suspicious/noArrayIndexKey: conv is append-only and immutable, so the index is a stable identity.
          <Bubble key={i} turn={g.turn} />
        ),
      )}
      {showLive && (
        <div className="max-w-[92%] self-start rounded-[10px] rounded-bl-[3px] border border-border bg-secondary px-2.5 py-2 text-xs leading-relaxed">
          {streaming ? (
            <Md>{streaming}</Md>
          ) : (
            <span className="typing">
              <i />
              <i />
              <i />
            </span>
          )}
        </div>
      )}
      {pending.map((ask) => (
        <PermissionCard
          key={ask.request_id}
          ask={ask}
          onResolve={(decision, opts) =>
            void resolvePermission(ask.request_id, decision, opts?.once, opts?.answers)
          }
        />
      ))}
      <div ref={endRef} />
    </div>
  )
}

// agrupar colapsa los turnos `act` consecutivos en un grupo (la tarjeta de actividad,
// CH-D2); el resto pasa como turno suelto en su orden real.
type Grupo = { kind: "turn"; turn: Turn } | { kind: "act"; pasos: string[] }

function agrupar(conv: Turn[]): Grupo[] {
  const out: Grupo[] = []
  for (const t of conv) {
    const last = out[out.length - 1]
    if (t.rol === "act") {
      if (last?.kind === "act") last.pasos.push(t.text)
      else out.push({ kind: "act", pasos: [t.text] })
    } else {
      out.push({ kind: "turn", turn: t })
    }
  }
  return out
}

// paso "<tool> <blanco>" → rótulo legible; «thinking» se muestra como razonamiento,
// jamás su contenido.
function rotulo(paso: string) {
  const [tool = "", ...resto] = paso.split(" ")
  return { tool: tool === "thinking" ? "pensó" : tool, obj: resto.join(" ") }
}

// ActivityCard (CH-D2): la fila punteada discreta entre burbujas. Viva = muestra la
// herramienta en curso y queda abierta; cerrada = resumen de un renglón, chevron despliega.
function ActivityCard({ pasos, live }: { pasos: string[]; live: boolean }) {
  const cur = rotulo(pasos[pasos.length - 1] ?? "")
  return (
    <details
      open={live || undefined}
      className="group max-w-[92%] self-stretch rounded-md border border-dashed border-border text-[11px]"
    >
      <summary
        className={cn(
          "flex cursor-pointer select-none items-center gap-1.5 px-2.5 py-1 [&::-webkit-details-marker]:hidden",
          live ? "text-foreground" : "text-muted-foreground",
        )}
      >
        <span className={cn("flex-none", live && "animate-pulse text-skill")} aria-hidden>
          ⚙
        </span>
        {live ? (
          <span className="min-w-0 truncate">
            trabajando — <b>{cur.tool}</b>{" "}
            {cur.obj && <span className="font-mono text-muted-foreground">{cur.obj}</span>}
          </span>
        ) : (
          <span>
            {pasos.length} {pasos.length === 1 ? "paso" : "pasos"}
          </span>
        )}
        <span className="ml-auto flex-none text-[9px] transition-transform group-open:rotate-90">
          ▶
        </span>
      </summary>
      <div className="border-t border-dashed border-border px-2.5 py-1 font-mono text-[10px] leading-[1.9] text-muted-foreground">
        {pasos.map((p, i) => {
          const { tool, obj } = rotulo(p)
          const running = live && i === pasos.length - 1
          return (
            // biome-ignore lint/suspicious/noArrayIndexKey: pasos is append-only, the index is a stable identity.
            <div key={i} className="flex items-baseline gap-1.5">
              <span className={cn("flex-none", running ? "animate-pulse text-skill" : "text-ok")}>
                {running ? "⟳" : "✓"}
              </span>
              <b className="flex-none text-foreground">{tool}</b>
              <span className="min-w-0 truncate">{obj}</span>
            </div>
          )
        })}
      </div>
    </details>
  )
}

function Bubble({ turn: t }: { turn: Turn }) {
  if (t.rol === "sys") {
    return (
      <div className="self-center rounded-md border border-dashed border-input bg-secondary px-2 py-0.5 font-mono text-[9.5px] text-muted-foreground">
        {t.text}
      </div>
    )
  }
  return (
    <div
      className={cn(
        "max-w-[92%] rounded-[10px] border px-2.5 py-2 text-xs leading-relaxed",
        t.rol === "user"
          ? "self-end whitespace-pre-wrap rounded-br-[3px] border-primary bg-accent-soft"
          : "self-start rounded-bl-[3px] border-border bg-secondary",
      )}
    >
      {t.rol === "assistant" ? (
        // CH-D4: la respuesta se ve renderizada (markdown), no cruda.
        <Md>{t.text}</Md>
      ) : (
        t.text
      )}
    </div>
  )
}

// fitComposer ajusta la altura al contenido; el tope son 3 líneas (max-h del textarea) y
// de ahí scroll interno. Vacío NO se dimensiona: el placeholder envuelto inflaría
// scrollHeight (WebKitGTK no soporta field-sizing:content → JS).
function fitComposer(ta: HTMLTextAreaElement) {
  ta.style.height = "auto"
  if (ta.value) ta.style.height = `${ta.scrollHeight}px`
}

function Composer() {
  const sendTurn = useSessions((s) => s.sendTurn)
  const interrupt = useSessions((s) => s.interrupt)
  const active = useSessions(selectActive)
  const [value, setValue] = useState("")
  const propuestaChat = useAppStore((st) => st.propuestaChat)
  const setPropuestaChat = useAppStore((st) => st.setPropuestaChat)
  const taRef = useRef<HTMLTextAreaElement>(null)
  // crudoMotivo marca que el texto de abajo vino SIN ordenar (RF-227). Se limpia en cuanto
  // el operador toca el campo: a partir de ahí el texto es suyo, no el crudo del dictado.
  const [crudoMotivo, setCrudoMotivo] = useState<string | undefined>()

  // CH-D1: auto-grow con el contenido (fitComposer); el ResizeObserver recalcula el
  // wrap cuando el dock cambia de ancho (CH-D5).
  // biome-ignore lint/correctness/useExhaustiveDependencies: value es el trigger intencional; el body solo lee el ref.
  useEffect(() => {
    if (taRef.current) fitComposer(taRef.current)
  }, [value])
  useEffect(() => {
    const ta = taRef.current
    if (!ta) return
    const ro = new ResizeObserver(() => fitComposer(ta))
    ro.observe(ta)
    return () => ro.disconnect()
  }, [])
  // El turno está en vuelo mientras streaming O await (parked en un permiso): el server
  // responde 409 a un segundo turno en ambos — el composer lo refleja (RF-116).
  const busy = active?.status === "streaming" || active?.status === "await"

  const submit = () => {
    const text = value.trim()
    if (!text || busy) return
    setValue("")
    setCrudoMotivo(undefined)
    void sendTurn(text)
  }

  // D26.4 · RF-255 — «Proponerlo en el chat» llega por acá. Es el MISMO trato que el dictado
  // y por la misma razón: **puebla el composer y no envía**. Un clic en una tarjeta no puede
  // convertirse en un turno real contra el código del operador; el fix se aplica por el camino
  // de siempre, con sus permisos y su gate (BR-M12 · D17.3).
  //
  // El buzón se consume UNA vez y se limpia — patrón `mapaPeek`. Sin limpiarlo, volver a abrir
  // el Dock repondría una propuesta vieja encima de lo que el operador estuviera escribiendo.
  useEffect(() => {
    if (propuestaChat === null) return
    setValue(propuestaChat)
    setPropuestaChat(null)
    requestAnimationFrame(() => {
      const ta = taRef.current
      if (!ta) return
      ta.focus()
      ta.setSelectionRange(ta.value.length, ta.value.length)
    })
  }, [propuestaChat, setPropuestaChat])

  // recibirDictado puebla el composer con lo dictado (RF-226). Lo que NO hace, y es el
  // punto: **no envía**. Auto-enviar convertiría un error de transcripción en un turno real
  // contra el código del operador.
  const recibirDictado = (r: Resultado) => {
    setValue(r.texto)
    setCrudoMotivo(r.estado === "crudo" ? (r.motivo ?? "falló el paso de limpieza") : undefined)
    // Foco al final: lo primero que uno hace con un dictado es corregirle una palabra.
    requestAnimationFrame(() => {
      const ta = taRef.current
      if (!ta) return
      ta.focus()
      ta.setSelectionRange(ta.value.length, ta.value.length)
    })
  }

  return (
    <>
      <VoiceBar />
      <div className="flex flex-none items-end gap-2 border-t border-border p-3">
        <textarea
          ref={taRef}
          value={value}
          rows={1}
          placeholder={
            active?.status === "await"
              ? "esperando tu decisión de permiso…"
              : busy
                ? "generando… (■ para interrumpir)"
                : `Pídele un cambio a ${active?.arnes ?? "…"}`
          }
          onChange={(e) => {
            setValue(e.target.value)
            // El operador editó: ya no es «el crudo del dictado», es su texto.
            if (crudoMotivo) setCrudoMotivo(undefined)
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault()
              submit()
            }
          }}
          className={cn(
            "max-h-[66px] min-h-[36px] flex-1 resize-none rounded-lg border border-border bg-secondary px-2.5 py-2 text-xs text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none",
            // El crudo se DISTINGUE del limpio: pasarlo por limpio sería un pass fabricado.
            crudoMotivo && "border-warn focus:border-warn",
          )}
        />
        {active && !busy && <DictadoButton sesionId={active.id} onTexto={recibirDictado} />}
        {busy ? (
          <button
            type="button"
            onClick={() => void interrupt()}
            title="Interrumpir el turno (in-band, la sesión sigue viva)"
            className="grid size-9 flex-none place-items-center rounded-lg bg-destructive text-sm text-destructive-foreground"
          >
            ■
          </button>
        ) : (
          <button
            type="button"
            onClick={submit}
            disabled={!value.trim()}
            title="Enviar (Claude Code headless detrás)"
            className="grid size-9 flex-none place-items-center rounded-lg bg-primary text-sm text-primary-foreground disabled:opacity-40"
          >
            ↑
          </button>
        )}
      </div>
      <DictadoAviso crudoMotivo={crudoMotivo} />
    </>
  )
}
