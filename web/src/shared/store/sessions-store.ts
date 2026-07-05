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
  type NewSession,
  type Session,
} from "@/shared/api"

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

  init: () => Promise<void>
  switchTo: (id: string) => void
  create: (input: NewSession) => Promise<void>
  closeSession: (id: string) => Promise<void>
  rename: (id: string, frente: string) => Promise<void>
  parkView: (view: string) => Promise<void>
  sendTurn: (text: string) => Promise<void>
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

export const useSessions = create<SessionsState>((set, get) => ({
  sessions: [],
  activeId: null,
  chatOpen: false,
  railCollapsed: false,
  connected: false,
  loaded: false,
  streaming: {},
  finalizedRun: {},

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
    const body = text.trim()
    if (!id || !body) return
    // Client-side guard mirroring the server's one-turn-at-a-time rule (409): never send
    // while this session is already streaming.
    if (get().sessions.find((s) => s.id === id)?.status === "streaming") return
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

      case "result":
        set((st) => {
          const finalText = (st.streaming[id] ?? "").trim() || (f.text ?? "")
          const rest = { ...st.streaming }
          delete rest[id]
          return {
            streaming: rest,
            finalizedRun: f.run_id ? { ...st.finalizedRun, [id]: f.run_id } : st.finalizedRun,
            sessions: patch(st.sessions, id, (s) => ({
              ...s,
              status: "idle",
              ctx_pct: f.ctx_pct && f.ctx_pct > 0 ? f.ctx_pct : s.ctx_pct,
              conv: [...(s.conv ?? []), { rol: "assistant", text: finalText }],
            })),
          }
        })
        break

      case "error":
        set((st) => {
          const rest = { ...st.streaming }
          delete rest[id]
          return {
            streaming: rest,
            finalizedRun: f.run_id ? { ...st.finalizedRun, [id]: f.run_id } : st.finalizedRun,
            sessions: patch(st.sessions, id, (s) => ({
              ...s,
              status: "idle",
              conv: [...(s.conv ?? []), { rol: "sys", text: `error: ${f.text ?? ""}` }],
            })),
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
