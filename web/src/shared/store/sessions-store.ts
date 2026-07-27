// Multisesión store (Zustand). It mirrors the daemon's session registry and drives the
// live Dock: optimistic user turns, per-session streaming buffers, and SSE frame
// routing. The daemon is the source of truth (persist + resume); this store is the
// live view of it.

import { create } from "zustand"
import {
  ApiError,
  api,
  connectDock,
  type DockConnection,
  type DockFrame,
  fetchAuthToken,
  type GateReport,
  type NewSession,
  type PermissionAsk,
  type ScopeNode,
  type Session,
} from "@/shared/api"
import { useMapLive } from "./map-live-store"

// Escrituras directas (mismo set que el conductor filtra de --allowedTools): si el turno
// aprobó al menos una, al `result` corre el gate de conformance del arnés (RF-117).
const WRITE_TOOLS = new Set(["Write", "Edit", "MultiEdit", "NotebookEdit"])

interface SessionsState {
  sessions: Session[]
  activeId: string | null
  chatOpen: boolean
  railCollapsed: boolean
  connected: boolean
  loaded: boolean
  // streaming[id] = assistant text assembled from deltas for the in-flight turn.
  streaming: Record<string, string>
  // finalizedRun[id] = the last run_id that received a terminal (result/error) frame. Frames
  // of an already-finalized run are dropped so an SSE reconnect+replay can't double-apply a
  // turn (boundary sesion-viva-consistente `frames-idempotentes-run-id`).
  finalizedRun: Record<string, string>
  // pendingPerms[id] = tarjetas de permiso abiertas de la sesión (RF-113, orden de llegada).
  pendingPerms: Record<string, PermissionAsk[]>
  // scope[id] = chip de alcance del composer (RF-111): nodo del Mapa → archivo real.
  scope: Record<string, ScopeNode | null>
  // wroteInRun[id] = el turno en vuelo aprobó ≥1 escritura ⇒ al result corre el gate (RF-117).
  wroteInRun: Record<string, boolean>
  // msgFlushed[id] = algún frame `message` ya cerró burbuja este turno (CH-D3) ⇒ el
  // result no re-arma el texto desde su propio campo (duplicaría).
  msgFlushed: Record<string, boolean>
  // convRev[id] = revisión de las conversaciones de la sesión. Es el seam con el store del
  // panel (widgets/chat-dock/model): `shared` NO puede importar `widgets`
  // (`shared-no-upward`), así que el panel se SUSCRIBE a este contador y refetchea cuando
  // cambia. Un contador y no la lista: la lista es del panel, esta capa sólo dice «cambió».
  convRev: Record<string, number>
  // detalleForzado[id] = el detalle de identidad se abre SOLO (RF-314) porque el cc-id
  // acaba de cambiar: al retomar y al rotar. Sin esto, el ctx cayendo de 92 % a 12 % y un
  // cc-id nuevo pasarían inadvertidos, que es exactamente lo que CV-D10 quiere evitar.
  detalleForzado: Record<string, boolean>
  // desactivadaTitulo[id] = el TÍTULO de la conversación que quedó inactiva al crear la
  // actual. Lo necesita el vacío del transcript nuevo, que la nombra (RF-307 CA-1). Sale de
  // la copia local —el store SABE cuál era la activa cuando llegó el frame `creada`—, así
  // que no hace falta un campo más en el wire.
  desactivadaTitulo: Record<string, string>

  init: () => Promise<void>
  switchTo: (id: string) => void
  create: (input: NewSession) => Promise<void>
  closeSession: (id: string) => Promise<void>
  rename: (id: string, frente: string) => Promise<void>
  parkView: (view: string) => Promise<void>
  sendTurn: (text: string) => Promise<void>
  resolvePermission: (
    requestId: string,
    decision: "allow" | "deny",
    once?: boolean,
    answers?: Record<string, string>,
  ) => Promise<void>
  interrupt: () => Promise<void>
  setScope: (node: ScopeNode | null) => void
  toggleChat: () => void
  openChat: () => void
  closeChat: () => void
  toggleRail: () => void
  onDock: (frame: DockFrame) => void
}

let dockConn: DockConnection | null = null

// patch replaces one session in the list by id via a mutator.
function patch(list: Session[], id: string, fn: (s: Session) => Session): Session[] {
  return list.map((s) => (s.id === id ? fn(s) : s))
}

// appendConv adds one turn to a session's transcript (local mirror; the daemon's Conv is
// the persisted truth — permisos/gate son rastro vivo de esta vista).
//
// Bajo CV-D3 el transcript vive en la conversación ACTIVA, no en la sesión: `s.activa.conv`.
// Sin `?? []` a propósito — `activa.conv` nunca es opcional en el wire (el daemon manda `[]`,
// nunca `null`), y un fallback acá volvería a hacer indistinguible «vacía» de «no cargada».
function appendConv(
  list: Session[],
  id: string,
  rol: "user" | "assistant" | "sys" | "act",
  text: string,
): Session[] {
  return patch(list, id, (s) => ({
    ...s,
    activa: { ...s.activa, conv: [...s.activa.conv, { rol, text }] },
  }))
}

// bump incrementa la revisión de conversaciones de una sesión (el seam con el panel).
function bump(rev: Record<string, number>, id: string): Record<string, number> {
  return { ...rev, [id]: (rev[id] ?? 0) + 1 }
}

// sinClave devuelve el registro sin una clave. Los cuatro buffers por-turno se LIMPIAN en
// una transición de conversación (el hilo nuevo no hereda el turno a medio ensamblar del
// viejo); `finalizedRun` y `scope` NO — el primero es idempotencia por run_id, que sobrevive
// a la transición, y el segundo es alcance de la SESIÓN (RF-118/RF-352 CA-3).
function sinClave<T>(reg: Record<string, T>, id: string): Record<string, T> {
  const { [id]: _fuera, ...resto } = reg
  return resto
}

export const useSessions = create<SessionsState>((set, get) => ({
  sessions: [],
  activeId: null,
  chatOpen: false,
  railCollapsed: false,
  connected: false,
  loaded: false,
  streaming: {},
  msgFlushed: {},
  finalizedRun: {},
  pendingPerms: {},
  scope: {},
  wroteInRun: {},
  convRev: {},
  detalleForzado: {},
  desactivadaTitulo: {},

  init: async () => {
    // Get the API capability token from the Tauri shell before any request (undefined in the
    // dev browser → the daemon falls back to its Host+Origin gate).
    api.setToken(await fetchAuthToken())
    // The daemon may still be binding :4200 when the WebView mounts (Tauri spawns it
    // as a sidecar concurrently). Retry the first load until it answers.
    let sessions: Session[] = []
    for (let attempt = 0; ; attempt++) {
      try {
        sessions = await api.listSessions()
        break
      } catch (err) {
        if (attempt >= 40) {
          console.error("arnesia: daemon unreachable after retries", err)
          set({ loaded: true })
          return
        }
        await new Promise((r) => setTimeout(r, 500))
      }
    }
    set((st) => ({
      sessions,
      loaded: true,
      activeId: st.activeId ?? sessions[0]?.id ?? null,
    }))
    if (!dockConn) {
      dockConn = connectDock(
        (f) => get().onDock(f),
        (c) => set({ connected: c }),
        // Reindex-en-vivo (RF-187): un turno reindexó un arnés → bump de su revisión;
        // la vista Mapa que lo mire refetchea.
        (m) => useMapLive.getState().bump(m.harness_id),
      )
    }
  },

  switchTo: (id) => set({ activeId: id }),

  create: async (input) => {
    const sess = await api.createSession(input)
    set((st) => ({
      sessions: [...st.sessions, sess],
      activeId: sess.id,
      chatOpen: true,
    }))
  },

  closeSession: async (id) => {
    await api.closeSession(id)
    set((st) => {
      const sessions = st.sessions.filter((s) => s.id !== id)
      const activeId = st.activeId === id ? (sessions[0]?.id ?? null) : st.activeId
      return { sessions, activeId }
    })
  },

  rename: async (id, frente) => {
    const trimmed = frente.trim()
    if (!trimmed) return
    set((st) => ({
      sessions: patch(st.sessions, id, (s) => ({ ...s, frente: trimmed })),
    }))
    await api.renameSession(id, trimmed)
  },

  parkView: async (view) => {
    const id = get().activeId
    if (!id) return
    set((st) => ({
      sessions: patch(st.sessions, id, (s) => ({ ...s, view })),
    }))
    await api.setView(id, view)
  },

  sendTurn: async (text) => {
    const id = get().activeId
    const trimmed = text.trim()
    if (!id || !trimmed) return
    // Client-side guard mirroring the server's one-turn-at-a-time rule (409): never send
    // while this session's turn is in flight (streaming O parked en un permiso).
    const status = get().sessions.find((s) => s.id === id)?.status
    if (status === "streaming" || status === "await") return
    // RF-111: el chip de alcance viaja como línea de contexto ANTEPUESTA al turno — el
    // texto enviado ES el que se ve (transparencia; el daemon persiste este payload).
    const sc = get().scope[id]
    const body = sc
      ? `[alcance: ${sc.clase ?? "nodo"} «${sc.nodeId}»${sc.fuentePath ? ` — archivo ${sc.fuentePath}` : ""}]\n\n${trimmed}`
      : trimmed
    // Optimistic: show the user turn and flip to streaming immediately.
    set((st) => ({
      sessions: patch(st.sessions, id, (s) => ({
        ...s,
        status: "streaming",
        activa: { ...s.activa, conv: [...s.activa.conv, { rol: "user", text: body }] },
      })),
      streaming: { ...st.streaming, [id]: "" },
    }))
    try {
      await api.turn(id, body)
    } catch (err) {
      // 409 = the daemon already has a turn in flight (a race the guard above almost always
      // prevents). Drop the optimistic user turn and leave the live stream untouched.
      if (err instanceof ApiError && err.status === 409) {
        set((st) => ({
          sessions: patch(st.sessions, id, (s) => ({
            ...s,
            activa: { ...s.activa, conv: s.activa.conv.slice(0, -1) },
          })),
        }))
        return
      }
      set((st) => ({
        sessions: appendConv(st.sessions, id, "sys", `error: ${String(err)}`).map((s) =>
          s.id === id ? { ...s, status: "idle" } : s,
        ),
      }))
    }
  },

  // resolvePermission (RF-113): la decisión humana sobre la tarjeta. `once` acota el
  // grant a 1 s (la siguiente petición del mismo tool VUELVE a preguntar). La tarjeta se
  // cierra cuando llega el frame `permission_result` del daemon (él es la verdad).
  resolvePermission: async (requestId, decision, once, answers) => {
    const id = get().activeId
    if (!id) return
    try {
      await api.resolvePermission(id, requestId, decision, once ? 1 : undefined, answers)
    } catch (err) {
      set((st) => ({
        sessions: appendConv(st.sessions, id, "sys", `error al resolver permiso: ${String(err)}`),
      }))
    }
  },

  // interrupt (RF-116): Stop real del turno en vuelo; el cierre llega como `result`.
  interrupt: async () => {
    const id = get().activeId
    if (!id) return
    try {
      await api.interrupt(id)
    } catch (err) {
      set((st) => ({
        sessions: appendConv(st.sessions, id, "sys", `error al interrumpir: ${String(err)}`),
      }))
    }
  },

  // setScope (RF-111): el nodo seleccionado en el Mapa como chip removible del composer,
  // por sesión (RF-118: el alcance pertenece a UNA sesión).
  setScope: (node) => {
    const id = get().activeId
    if (!id) return
    set((st) => ({ scope: { ...st.scope, [id]: node } }))
  },

  toggleChat: () => set((st) => ({ chatOpen: !st.chatOpen })),
  openChat: () => set({ chatOpen: true }),
  closeChat: () => set({ chatOpen: false }),
  toggleRail: () => set((st) => ({ railCollapsed: !st.railCollapsed })),

  onDock: (f) => {
    const id = f.session_id
    // Idempotency: once a run received its terminal frame, drop any further frame for it —
    // this is how an SSE reconnect+replay (or a late delta) can't duplicate a turn.
    if (f.run_id && get().finalizedRun[id] === f.run_id) return
    switch (f.kind) {
      case "status":
        set((st) => ({
          sessions: patch(st.sessions, id, (s) => ({
            ...s,
            status: f.status ?? s.status,
          })),
        }))
        break

      case "init":
        set((st) => ({
          sessions: patch(st.sessions, id, (s) => ({
            ...s,
            activa: {
              ...s.activa,
              claude_session_id: f.claude_session_id ?? s.activa.claude_session_id,
              model: f.model ?? s.activa.model,
            },
          })),
        }))
        break

      case "delta":
        set((st) => ({
          streaming: {
            ...st.streaming,
            [id]: (st.streaming[id] ?? "") + (f.text ?? ""),
          },
          sessions: patch(st.sessions, id, (s) => ({
            ...s,
            status: "streaming",
          })),
        }))
        break

      case "result": {
        const wrote = get().wroteInRun[id] === true
        const arnes = get().sessions.find((s) => s.id === id)?.arnes
        set((st) => {
          // CH-D3: los frames `message` ya cerraron sus burbujas; acá solo cae el
          // remanente del stream, o el fallback de un turno sin messages.
          let finalText = (st.streaming[id] ?? "").trim()
          if (!finalText && !st.msgFlushed[id]) finalText = f.text ?? ""
          const rest = { ...st.streaming }
          delete rest[id]
          const wroteRest = { ...st.wroteInRun }
          delete wroteRest[id]
          const flushedRest = { ...st.msgFlushed }
          delete flushedRest[id]
          return {
            streaming: rest,
            wroteInRun: wroteRest,
            msgFlushed: flushedRest,
            finalizedRun: f.run_id ? { ...st.finalizedRun, [id]: f.run_id } : st.finalizedRun,
            sessions: patch(st.sessions, id, (s) => ({
              ...s,
              status: "idle",
              activa: {
                ...s.activa,
                ctx_pct: f.ctx_pct && f.ctx_pct > 0 ? f.ctx_pct : s.activa.ctx_pct,
                conv: finalText
                  ? [...s.activa.conv, { rol: "assistant", text: finalText }]
                  : s.activa.conv,
              },
            })),
          }
        })
        // RF-117: el turno aprobó escrituras ⇒ el gate del arnés corre y se VE. Nada es
        // «listo» de palabra: el veredicto (o el fallo honesto del fetch) queda en el chat.
        if (wrote && arnes) {
          void (async () => {
            try {
              const report = await api.getConformance<GateReport>(arnes)
              const results = report.results ?? []
              const notPass = results.filter(
                (r) => r.veredicto === "fail" || r.veredicto === "error",
              )
              // Semántica del dominio (ConformanceReport.OK): solo un fallo de severidad
              // ERROR bloquea; un warn es hallazgo visible, jamás un rojo fingido.
              const blockers = notPass.filter((r) => r.check.severidad === "error")
              const warns = notPass.filter((r) => r.check.severidad !== "error")
              const passN = results.filter((r) => r.veredicto === "pass").length
              const warnTail = warns.length
                ? ` · ${warns.length} warn (${warns.map((r) => r.check.id).join(" · ")})`
                : ""
              const verdict = blockers.length
                ? `🛡 gate de conformance ${arnes}: ${blockers.length} BLOQUEANTE — ${blockers.map((r) => r.check.id).join(" · ")}${warnTail}`
                : `🛡 gate de conformance ${arnes}: ${passN}/${passN + notPass.length} pass${warnTail} — sin bloqueos, el arnés sigue conforme`
              set((st) => ({ sessions: appendConv(st.sessions, id, "sys", verdict) }))
            } catch (err) {
              set((st) => ({
                sessions: appendConv(
                  st.sessions,
                  id,
                  "sys",
                  `🛡 gate de conformance no corrió: ${String(err)}`,
                ),
              }))
            }
          })()
        }
        break
      }

      case "error":
        set((st) => {
          const rest = { ...st.streaming }
          delete rest[id]
          const wroteRest = { ...st.wroteInRun }
          delete wroteRest[id]
          return {
            streaming: rest,
            wroteInRun: wroteRest,
            // El turno murió: ninguna tarjeta pendiente sigue viva (el conductor ya no
            // espera respuesta) — cerrar sin decisión es honesto, no silencioso: el error
            // queda en el transcript.
            pendingPerms: { ...st.pendingPerms, [id]: [] },
            finalizedRun: f.run_id ? { ...st.finalizedRun, [id]: f.run_id } : st.finalizedRun,
            sessions: appendConv(st.sessions, id, "sys", `error: ${f.text ?? ""}`).map((s) =>
              s.id === id ? { ...s, status: "idle" } : s,
            ),
          }
        })
        break

      case "permission":
        // RF-113: tarjeta ask→UI. La sesión queda `await` (pip ámbar); la tarjeta vive
        // hasta su permission_result (el daemon es la verdad, no el click local).
        if (f.request_id && f.tool) {
          const ask: PermissionAsk = { request_id: f.request_id, tool: f.tool, input: f.input }
          set((st) => {
            const cur = st.pendingPerms[id] ?? []
            if (cur.some((p) => p.request_id === ask.request_id)) return st
            return {
              pendingPerms: { ...st.pendingPerms, [id]: [...cur, ask] },
              sessions: patch(st.sessions, id, (s) => ({ ...s, status: f.status ?? "await" })),
            }
          })
        }
        break

      case "permission_result":
        // Cierra la tarjeta y deja rastro en el transcript. Un allow de escritura arma
        // el gate del turno (RF-117). También cubre el auto-allow por grant vigente
        // (RF-115 — llega sin tarjeta previa, solo el rastro).
        set((st) => {
          const cur = st.pendingPerms[id] ?? []
          const rest = cur.filter((p) => p.request_id !== f.request_id)
          const allowedWrite = f.decision === "allow" && !!f.tool && WRITE_TOOLS.has(f.tool)
          const mark = f.decision === "allow" ? "✓" : "✕"
          const rastro = `${mark} ${f.tool ?? "tool"}: ${f.decision ?? ""}${f.text ? ` — ${f.text}` : ""}`
          return {
            pendingPerms: { ...st.pendingPerms, [id]: rest },
            wroteInRun: allowedWrite ? { ...st.wroteInRun, [id]: true } : st.wroteInRun,
            sessions: appendConv(
              f.status
                ? patch(st.sessions, id, (s) => ({ ...s, status: f.status ?? s.status }))
                : st.sessions,
              id,
              "sys",
              rastro,
            ),
          }
        })
        break

      case "message":
        // CH-D3: el daemon cerró una burbuja — congelarla como turno y vaciar el
        // stream vivo (los próximos deltas abren burbuja nueva).
        if (f.text) {
          set((st) => {
            const rest = { ...st.streaming }
            delete rest[id]
            return {
              streaming: rest,
              msgFlushed: { ...st.msgFlushed, [id]: true },
              sessions: appendConv(st.sessions, id, "assistant", f.text ?? ""),
            }
          })
        }
        break

      case "act":
        // CH-D2: paso de actividad — rastro `act` en el transcript; el ChatDock agrupa
        // consecutivos en la tarjeta desplegable.
        set((st) => ({
          sessions: appendConv(
            patch(st.sessions, id, (s) => ({ ...s, status: f.status ?? "streaming" })),
            id,
            "act",
            `${f.tool ?? ""} ${f.text ?? ""}`.trim(),
          ),
        }))
        break

      case "conversacion":
        // CV-D2/CV-D10 · design.md §7.2. Cuatro eventos, tres formas de idempotencia:
        //
        //  · `creada`/`activada` traen el estado POST-transición completo ⇒ aplicarlas dos
        //    veces es un `set`, no un append. Reemplazan la activa y limpian los buffers del
        //    turno que se quedó en el hilo viejo.
        //  · `renombrada` puede venir de una conversación INACTIVA (renombrar no exige que sea
        //    la activa): se aplica sólo si es la activa, y siempre bumpea la revisión para que
        //    el panel, si está abierto, se entere.
        //  · `rotada` es un append, así que su llave es `turno_idx`: se appendea SÓLO si la
        //    copia local tiene exactamente esa longitud. Un replay por Last-Event-ID llega con
        //    la copia más larga y se descarta solo (E-41).
        set((st) => {
          const evento = f.conversacion_evento
          const nueva = f.conversacion
          if (evento === "creada" || evento === "activada") {
            if (!nueva) return { convRev: bump(st.convRev, id) }
            const anterior = st.sessions.find((s) => s.id === id)?.activa.titulo
            return {
              // Los buffers del turno viejo NO se heredan…
              streaming: sinClave(st.streaming, id),
              pendingPerms: sinClave(st.pendingPerms, id),
              msgFlushed: sinClave(st.msgFlushed, id),
              wroteInRun: sinClave(st.wroteInRun, id),
              // …pero `finalizedRun` (idempotencia por run_id) y `scope` (alcance de la
              // SESIÓN, RF-352 CA-3) SÍ: limpiarlos «por simetría» sería un bug.
              convRev: bump(st.convRev, id),
              detalleForzado: { ...st.detalleForzado, [id]: evento === "activada" },
              desactivadaTitulo:
                evento === "creada" && anterior
                  ? { ...st.desactivadaTitulo, [id]: anterior }
                  : sinClave(st.desactivadaTitulo, id),
              sessions: patch(st.sessions, id, (s) => ({ ...s, activa: nueva })),
            }
          }
          if (evento === "renombrada") {
            return {
              convRev: bump(st.convRev, id),
              sessions:
                nueva && f.conversacion_id === st.sessions.find((s) => s.id === id)?.activa.id
                  ? patch(st.sessions, id, (s) => ({ ...s, activa: nueva }))
                  : st.sessions,
            }
          }
          if (evento === "rotada") {
            const actual = st.sessions.find((s) => s.id === id)
            if (!actual || actual.activa.conv.length !== f.turno_idx) return st
            return {
              convRev: bump(st.convRev, id),
              detalleForzado: { ...st.detalleForzado, [id]: true },
              sessions: patch(
                appendConv(st.sessions, id, "sys", f.text ?? ""),
                id,
                // El daemon limpió el `ClaudeSessionID` al rotar: el del hilo nuevo llega con
                // el `init` del próximo spawn. Dejar el viejo pintaría una identidad muerta.
                (s) => ({ ...s, activa: { ...s.activa, claude_session_id: undefined } }),
              ),
            }
          }
          return { convRev: bump(st.convRev, id) }
        })
        break

      default:
        // Un `kind` que este build no conoce NO se traga en silencio: el daemon puede ser
        // más nuevo que la SPA (self-update reemplaza el binario, no el bundle servido) y
        // «no pasó nada» sería indistinguible de «el frame se perdió». Agujero preexistente
        // que este paquete ensancha con un kind más, así que lo cierra acá.
        console.warn("arnesia: frame de dock con kind desconocido", f.kind, f)
        break
    }
  },
}))

// selectors -----------------------------------------------------------------

export const selectActive = (st: SessionsState): Session | undefined =>
  st.sessions.find((s) => s.id === st.activeId)

export const selectAttention = (st: SessionsState): number =>
  st.sessions.filter((s) => s.status === "await").length

const NO_PERMS: PermissionAsk[] = []

// selectPendingPerms — las tarjetas de permiso abiertas de la sesión activa (RF-113).
export const selectPendingPerms = (st: SessionsState): PermissionAsk[] =>
  (st.activeId ? st.pendingPerms[st.activeId] : undefined) ?? NO_PERMS

// selectScope — el chip de alcance de la sesión activa (RF-111).
export const selectScope = (st: SessionsState): ScopeNode | null =>
  (st.activeId ? st.scope[st.activeId] : null) ?? null

// Los selectores del cromo de conversación devuelven PRIMITIVOS, sin excepción. Zustand
// compara por identidad: uno que devolviera un objeto nuevo re-renderizaría el dock —y con
// él el transcript entero— en cada frame `delta`. El repo no usa `useShallow` en ningún lado
// y este paquete no lo introduce (arquitectura.md §6.3).

// selectConvActivaId — el id de la conversación activa de la sesión activa.
export const selectConvActivaId = (st: SessionsState): string | undefined =>
  selectActive(st)?.activa.id

// selectCtxPct — el porcentaje de contexto usado. 0 es DATO (BR-CV-9): sin sesión activa
// devuelve `undefined`, que es otra cosa.
export const selectCtxPct = (st: SessionsState): number | undefined =>
  selectActive(st)?.activa.ctx_pct

// selectCtxCaliente — el chip va en variante caliente. Sale de `rotacion_pendiente` del
// wire: el umbral vive en el daemon (`SetUmbralRotacion`) y el FE no lo conoce ni lo teclea.
export const selectCtxCaliente = (st: SessionsState): boolean =>
  selectActive(st)?.activa.rotacion_pendiente ?? false

// selectTituloActiva — el título de la conversación activa (la fila 2 del cromo).
export const selectTituloActiva = (st: SessionsState): string | undefined =>
  selectActive(st)?.activa.titulo
