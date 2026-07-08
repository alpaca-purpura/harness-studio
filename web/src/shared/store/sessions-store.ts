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
function appendConv(
  list: Session[],
  id: string,
  rol: "user" | "assistant" | "sys",
  text: string,
): Session[] {
  return patch(list, id, (s) => ({ ...s, conv: [...(s.conv ?? []), { rol, text }] }))
}

export const useSessions = create<SessionsState>((set, get) => ({
  sessions: [],
  activeId: null,
  chatOpen: false,
  railCollapsed: false,
  connected: false,
  loaded: false,
  streaming: {},
  finalizedRun: {},
  pendingPerms: {},
  scope: {},
  wroteInRun: {},

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
        conv: [...(s.conv ?? []), { rol: "user", text: body }],
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
          sessions: patch(st.sessions, id, (s) => ({ ...s, conv: (s.conv ?? []).slice(0, -1) })),
        }))
        return
      }
      set((st) => ({
        sessions: patch(st.sessions, id, (s) => ({
          ...s,
          status: "idle",
          conv: [...(s.conv ?? []), { rol: "sys", text: `error: ${String(err)}` }],
        })),
      }))
    }
  },

  // resolvePermission (RF-113): la decisión humana sobre la tarjeta. `once` acota el
  // grant a 1 s (la siguiente petición del mismo tool VUELVE a preguntar). La tarjeta se
  // cierra cuando llega el frame `permission_result` del daemon (él es la verdad).
  resolvePermission: async (requestId, decision, once) => {
    const id = get().activeId
    if (!id) return
    try {
      await api.resolvePermission(id, requestId, decision, once ? 1 : undefined)
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
            claude_session_id: f.claude_session_id ?? s.claude_session_id,
            model: f.model ?? s.model,
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
          const finalText = (st.streaming[id] ?? "").trim() || (f.text ?? "")
          const rest = { ...st.streaming }
          delete rest[id]
          const wroteRest = { ...st.wroteInRun }
          delete wroteRest[id]
          return {
            streaming: rest,
            wroteInRun: wroteRest,
            finalizedRun: f.run_id ? { ...st.finalizedRun, [id]: f.run_id } : st.finalizedRun,
            sessions: patch(st.sessions, id, (s) => ({
              ...s,
              status: "idle",
              ctx_pct: f.ctx_pct && f.ctx_pct > 0 ? f.ctx_pct : s.ctx_pct,
              conv: [...(s.conv ?? []), { rol: "assistant", text: finalText }],
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
              const notPass = results.filter((r) => r.veredicto === "fail" || r.veredicto === "error")
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
            sessions: patch(st.sessions, id, (s) => ({
              ...s,
              status: "idle",
              conv: [...(s.conv ?? []), { rol: "sys", text: `error: ${f.text ?? ""}` }],
            })),
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
        // Full assistant message: superseded by delta assembly; ignored here.
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
