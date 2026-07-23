// Unit tests del store de conversaciones por arnés (RF-203, proyecto vitest `unit`).
// Mismo patrón que portafolio-picker-store.test.ts: espía el singleton `api`.

import { afterEach, describe, expect, it, vi } from "vitest"
import { api } from "@/shared"
import { useConversaciones } from "./conversaciones-store"

afterEach(() => {
  vi.restoreAllMocks()
  useConversaciones.getState().reset()
})

describe("conversaciones-store", () => {
  it("cargar puebla vivas + cerradas y conserva el error honesto de cerradas", async () => {
    vi.spyOn(api, "conversacionesDeArnes").mockResolvedValue({
      sesiones: [{ id: "s1", frente: "reparar", arnes: "vitalia", status: "idle", view: "Mapa" }],
      cerradas: [
        { id: "s0", frente: "vieja", arnes: "vitalia", status: "idle", view: "Mapa", turnos: 4 },
      ],
    })
    await useConversaciones.getState().cargar("vitalia")
    const st = useConversaciones.getState()
    expect(st.vivas).toHaveLength(1)
    expect(st.cerradas).toHaveLength(1)
    expect(st.error).toBeUndefined()
  })

  it("abrirHistorial trae turnos + faltantes honestos", async () => {
    vi.spyOn(api, "historialCerrada").mockResolvedValue({
      turnos: [
        { rol: "user", text: "uno" },
        { rol: "assistant", text: "dos" },
      ],
      faltantes: ["cc-borrada"],
    })
    await useConversaciones.getState().abrirHistorial("s0")
    const st = useConversaciones.getState()
    expect(st.historialDe).toBe("s0")
    expect(st.turnos).toHaveLength(2)
    expect(st.faltantes).toEqual(["cc-borrada"])
  })

  it("cargar con fallo deja el error visible, jamás datos fabricados", async () => {
    vi.spyOn(api, "conversacionesDeArnes").mockRejectedValue(new Error("daemon caído"))
    await useConversaciones.getState().cargar("vitalia")
    const st = useConversaciones.getState()
    expect(st.error).toBe("daemon caído")
    expect(st.vivas).toHaveLength(0)
    expect(st.cerradas).toHaveLength(0)
  })
})
