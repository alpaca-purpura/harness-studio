// Unit tests del store multisesión (proyecto vitest `unit` — Node, sin Chromium). Cubre la
// rama del frame `conversacion` (CV-D2/CV-D10 · design.md §7.2) y el `default:` del switch.
//
// El store de 500 líneas NO tenía ni un test hasta acá: es lo que dejó pasar que
// `appendConv` escribiera en un campo que el wire ya no manda. Estos seis miden lo que el
// tramo 3 agrega, y tres de ellos existen para PROHIBIR una simetría equivocada.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import type { ConversacionActiva, DockFrame, Session } from "@/shared/api"
import { useSessions } from "./sessions-store"

const CONV_A: ConversacionActiva = {
  id: "c1",
  titulo: "por qué el mapa sale vacío",
  titulo_editado: false,
  activa: true,
  turnos: 2,
  ctx_pct: 68,
  rotacion_pendiente: false,
  ultima_interaccion: "2026-07-26T18:02:00Z",
  creada_en: "2026-07-24T10:00:00Z",
  claude_session_id: "4b046945-1f0e-4c0a-9a51-1b2f9c8d7e30",
  model: "claude-opus-5[1m]",
  conv: [
    { rol: "user", text: "no veo nada en el mapa" },
    { rol: "assistant", text: "el manifiesto declara 0 elementos" },
  ],
}

const CONV_NUEVA: ConversacionActiva = {
  id: "c2",
  titulo: "nueva conversación",
  titulo_editado: false,
  activa: true,
  turnos: 0,
  ctx_pct: 0,
  rotacion_pendiente: false,
  creada_en: "2026-07-26T19:00:00Z",
  conv: [],
}

const SESION: Session = {
  id: "s6165ac75",
  frente: "repro del bug de carga",
  arnes: "vitalia",
  status: "idle",
  view: "Mapa",
  cwd: "~/Proyectos/luana-vitalia/vitalia",
  activa: CONV_A,
}

// Estado ENTERO, jamás un merge parcial sobre lo que dejó el test anterior.
function sembrar(parcial: Partial<ReturnType<typeof useSessions.getState>> = {}) {
  useSessions.setState({
    sessions: [structuredClone(SESION)],
    activeId: SESION.id,
    chatOpen: true,
    railCollapsed: false,
    connected: true,
    loaded: true,
    streaming: {},
    msgFlushed: {},
    finalizedRun: {},
    pendingPerms: {},
    scope: {},
    wroteInRun: {},
    convRev: {},
    detalleForzado: {},
    desactivadaTitulo: {},
    ...parcial,
  })
}

const frame = (f: Partial<DockFrame>): DockFrame =>
  ({ session_id: SESION.id, kind: "conversacion", ...f }) as DockFrame

beforeEach(() => sembrar())
afterEach(() => vi.restoreAllMocks())

describe("sessions-store · frame conversacion", () => {
  it("creada reemplaza la activa y limpia los buffers del turno viejo (E-06)", () => {
    sembrar({
      streaming: { [SESION.id]: "a medio ensamblar" },
      msgFlushed: { [SESION.id]: true },
      wroteInRun: { [SESION.id]: true },
      pendingPerms: { [SESION.id]: [{ request_id: "cr-1", tool: "Edit" }] },
    })
    useSessions
      .getState()
      .onDock(
        frame({ conversacion_id: "c2", conversacion_evento: "creada", conversacion: CONV_NUEVA }),
      )
    const st = useSessions.getState()
    expect(st.sessions[0]?.activa.id).toBe("c2")
    expect(st.sessions[0]?.activa.conv).toEqual([])
    expect(st.streaming[SESION.id]).toBeUndefined()
    expect(st.msgFlushed[SESION.id]).toBeUndefined()
    expect(st.wroteInRun[SESION.id]).toBeUndefined()
    expect(st.pendingPerms[SESION.id]).toBeUndefined()
    // El vacío del transcript nuevo NOMBRA la que se desactivó (RF-307 CA-1).
    expect(st.desactivadaTitulo[SESION.id]).toBe("por qué el mapa sale vacío")
    expect(st.convRev[SESION.id]).toBe(1)
  })

  it("finalizedRun NO se limpia en la transición (§4.4)", () => {
    sembrar({ finalizedRun: { [SESION.id]: "r-7" } })
    useSessions
      .getState()
      .onDock(
        frame({ conversacion_id: "c2", conversacion_evento: "creada", conversacion: CONV_NUEVA }),
      )
    // Limpiarlo «por simetría» rompería la idempotencia: un replay del run r-7 volvería a
    // aplicarse sobre el hilo nuevo.
    expect(useSessions.getState().finalizedRun[SESION.id]).toBe("r-7")
  })

  it("scope NO se limpia en la transición (E-48 · RF-352 CA-3)", () => {
    sembrar({ scope: { [SESION.id]: { nodeId: "hipaa-check", clase: "caja" } } })
    useSessions
      .getState()
      .onDock(
        frame({ conversacion_id: "c2", conversacion_evento: "activada", conversacion: CONV_NUEVA }),
      )
    // El alcance pertenece a la SESIÓN, no a la conversación.
    expect(useSessions.getState().scope[SESION.id]?.nodeId).toBe("hipaa-check")
    // Retomar SÍ abre el detalle solo: el cc-id cambió (RF-314).
    expect(useSessions.getState().detalleForzado[SESION.id]).toBe(true)
  })

  it("rotada con turno_idx correcto appendea la marca (E-19)", () => {
    useSessions.getState().onDock(
      frame({
        conversacion_id: "c1",
        conversacion_evento: "rotada",
        turno_idx: 2,
        text: "— contexto rotado, seguimos —",
      }),
    )
    const activa = useSessions.getState().sessions[0]?.activa
    expect(activa?.conv).toHaveLength(3)
    expect(activa?.conv[2]).toEqual({ rol: "sys", text: "— contexto rotado, seguimos —" })
    // El daemon limpió el cc-id al rotar: el del hilo nuevo llega con el próximo `init`.
    expect(activa?.claude_session_id).toBeUndefined()
    expect(useSessions.getState().detalleForzado[SESION.id]).toBe(true)
  })

  it("rotada con turno_idx repetido no duplica el breadcrumb (E-41)", () => {
    const f = frame({
      conversacion_id: "c1",
      conversacion_evento: "rotada",
      turno_idx: 2,
      text: "— contexto rotado, seguimos —",
    })
    useSessions.getState().onDock(f)
    useSessions.getState().onDock(f)
    // El replay llega con la copia local ya más larga y se descarta solo: sin run_id, la
    // llave de idempotencia es el índice.
    expect(useSessions.getState().sessions[0]?.activa.conv).toHaveLength(3)
  })

  it("renombrar una INACTIVA no pisa la activa, pero avisa a la lista", () => {
    const otra: ConversacionActiva = { ...CONV_NUEVA, id: "c9", titulo: "el bug del índice" }
    useSessions
      .getState()
      .onDock(
        frame({ conversacion_id: "c9", conversacion_evento: "renombrada", conversacion: otra }),
      )
    const st = useSessions.getState()
    expect(st.sessions[0]?.activa.id).toBe("c1")
    expect(st.sessions[0]?.activa.titulo).toBe("por qué el mapa sale vacío")
    expect(st.convRev[SESION.id]).toBe(1)
  })

  it("un kind desconocido no rompe y avisa", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {})
    useSessions.getState().onDock({ session_id: SESION.id, kind: "telepatia" } as never)
    expect(warn).toHaveBeenCalled()
    expect(useSessions.getState().sessions[0]?.activa.id).toBe("c1")
  })
})
