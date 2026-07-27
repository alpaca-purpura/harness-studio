// Unit tests del store del panel (proyecto vitest `unit` — Node, sin Chromium). U-01…U-10 de
// `plan-storybook.md` §2.7. Espía el singleton `api` exportado, no mockea el módulo entero:
// misma disciplina que `portafolio-picker-store.test.ts`.

import { afterEach, describe, expect, it, vi } from "vitest"
import { ApiError, api, useSessions } from "@/shared"
import { useConversaciones } from "./conversaciones-store"

const LISTA = {
  conversaciones: [
    {
      id: "c-mapa",
      titulo: "por qué el mapa sale vacío",
      titulo_editado: false,
      activa: true,
      turnos: 90,
      ctx_pct: 68,
      rotacion_pendiente: false,
      creada_en: "2026-07-24T10:00:00Z",
    },
  ],
  total: 4,
}

afterEach(() => {
  useConversaciones.getState().cerrar()
  useSessions.setState({ convRev: {} })
  vi.restoreAllMocks()
})

// listo espera a que la carga en vuelo termine: `abrir` dispara el fetch sin devolver la
// promesa (el panel se pinta en `cargando` antes de que llegue nada).
const listo = () =>
  vi.waitFor(() => expect(useConversaciones.getState().estado).not.toBe("cargando"))

describe("conversaciones-store", () => {
  it("U-01 · abrir refetchea SIEMPRE, jamás la lista de la apertura anterior", async () => {
    const spy = vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    // Entre dos aperturas el operador pudo mandar un turno o rotar: una lista cacheada
    // mentiría sobre el estado del hilo.
    expect(spy).toHaveBeenCalledTimes(2)
    expect(useConversaciones.getState().total).toBe(4)
  })

  it("U-02 · cerrar descarta la búsqueda y el estado (E-42)", async () => {
    vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    useConversaciones.getState().abrir("s1", "buscador")
    await listo()
    await useConversaciones.getState().buscar("manifiesto")
    expect(useConversaciones.getState().busqueda).toBe("manifiesto")
    useConversaciones.getState().cerrar()
    const st = useConversaciones.getState()
    expect(st.busqueda).toBe("")
    expect(st.abierta).toBe(false)
    expect(st.conversaciones).toEqual([])
  })

  it("U-03 · cambiar de sesión no deja la lista mostrando la anterior (E-42)", async () => {
    const spy = vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    useConversaciones.getState().abrir("s2", "filas")
    expect(useConversaciones.getState().sesionId).toBe("s2")
    // La lista se vacía ANTES de que llegue la de s2: pintar la de s1 bajo el rótulo de s2
    // sería atribuirle conversaciones ajenas a una sesión (CV-D4).
    expect(useConversaciones.getState().conversaciones).toEqual([])
    await listo()
    expect(spy).toHaveBeenLastCalledWith("s2", undefined, expect.anything())
  })

  it("U-04 · el motivo del fallo se guarda TAL CUAL, y nunca «0 conversaciones» (E-35)", async () => {
    vi.spyOn(api, "conversaciones").mockRejectedValue(new Error("ECONNREFUSED 127.0.0.1:4200"))
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    const st = useConversaciones.getState()
    expect(st.estado).toBe("error")
    expect(st.error).toBe("ECONNREFUSED 127.0.0.1:4200")
    expect(st.conversaciones).toEqual([])
  })

  it("U-05 · un crear que falla deja el estado anterior intacto (E-37)", async () => {
    vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    vi.spyOn(api, "crearConversacion").mockRejectedValue(new Error("disco lleno"))
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    await useConversaciones.getState().crear()
    const st = useConversaciones.getState()
    expect(st.fallo).toBe("disco lleno")
    expect(st.estado).toBe("datos")
    expect(st.conversaciones).toHaveLength(1)
    expect(st.total).toBe(4)
  })

  it("U-06 · un retomar que falla deja el estado anterior intacto (E-36)", async () => {
    vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    vi.spyOn(api, "activarConversacion").mockRejectedValue(new Error("daemon caído"))
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    await useConversaciones.getState().activar("c-sellar")
    const st = useConversaciones.getState()
    expect(st.fallo).toBe("daemon caído")
    expect(st.retomando).toBeNull()
    expect(st.abierta).toBe(true)
    expect(st.conversaciones).toHaveLength(1)
  })

  it("U-07 · el 409 se distingue del resto: es un turno en vuelo, no una caída (E-39)", async () => {
    vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    vi.spyOn(api, "crearConversacion").mockRejectedValue(
      new ApiError(409, "arnesia POST /api/sessions/s1/conversaciones: 409 turno en curso"),
    )
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    await useConversaciones.getState().crear()
    expect(useConversaciones.getState().fallo).toContain("hay un turno en vuelo")
    // Y sin perder el original: el operador tiene que poder leer qué dijo el daemon.
    expect(useConversaciones.getState().fallo).toContain("409")
  })

  it("U-08 · un fallo de transporte va por el mismo camino, con su motivo (E-38)", async () => {
    vi.spyOn(api, "conversaciones").mockRejectedValue(new TypeError("network timeout"))
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    expect(useConversaciones.getState().estado).toBe("error")
    expect(useConversaciones.getState().error).toBe("network timeout")
  })

  it("U-09 · un bump de convRev refetchea UNA vez; sin bump, ninguna (E-40/E-41)", async () => {
    const spy = vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    expect(spy).toHaveBeenCalledTimes(1)
    // Otro campo del store compartido cambia: la lista NO se recarga.
    useSessions.setState({ activeId: "s1" })
    expect(spy).toHaveBeenCalledTimes(1)
    // La revisión de ESTA sesión cambia: se recarga.
    useSessions.setState({ convRev: { s1: 1 } })
    await vi.waitFor(() => expect(spy).toHaveBeenCalledTimes(2))
    // La de OTRA sesión cambia: no.
    useSessions.setState({ convRev: { s1: 1, s2: 9 } })
    await listo()
    expect(spy).toHaveBeenCalledTimes(2)
  })

  it("U-10 · renombrar a vacío no llama a la API (E-30)", async () => {
    vi.spyOn(api, "conversaciones").mockResolvedValue(LISTA)
    const spy = vi.spyOn(api, "renombrarConversacion")
    useConversaciones.getState().abrir("s1", "filas")
    await listo()
    await useConversaciones.getState().renombrar("c-mapa", "   ")
    expect(spy).not.toHaveBeenCalled()
  })
})
