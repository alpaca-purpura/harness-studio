// Unit tests del store del picker (proyecto vitest `unit` — Node, sin Chromium; misma disciplina
// que entities/portafolio/model/selectors.test.ts). Cubre RF-7/RF-15: cargando→datos/error, caída
// a listas vacías, refetch en cada apertura, y reset. Spía `api.listPortafolio` (objeto singleton
// exportado) — no mockea el módulo entero.

import { afterEach, describe, expect, it, vi } from "vitest"
import { entradasDemo } from "@/entities/portafolio"
import { api } from "@/shared"
import { usePortafolioPicker } from "./portafolio-picker-store"

afterEach(() => {
  vi.restoreAllMocks()
  usePortafolioPicker.getState().reset()
})

describe("portafolio-picker-store", () => {
  it("cargar puebla entradas/corruptas y pasa a estado datos", async () => {
    vi.spyOn(api, "listPortafolio").mockResolvedValue({
      entradas: entradasDemo,
      corruptas: [{ motivo: "json: cannot unmarshal string into Go struct field" }],
    })
    await usePortafolioPicker.getState().cargar()
    const st = usePortafolioPicker.getState()
    expect(st.estado).toBe("datos")
    expect(st.entradas).toHaveLength(3)
    expect(st.corruptas).toHaveLength(1)
    expect(st.error).toBeUndefined()
  })

  it("cargar sin entradas/corruptas cae a listas vacías, nunca undefined", async () => {
    vi.spyOn(api, "listPortafolio").mockResolvedValue({})
    await usePortafolioPicker.getState().cargar()
    const st = usePortafolioPicker.getState()
    expect(st.estado).toBe("datos")
    expect(st.entradas).toEqual([])
    expect(st.corruptas).toEqual([])
  })

  it("cargar propaga el motivo del fallo a estado error", async () => {
    vi.spyOn(api, "listPortafolio").mockRejectedValue(new Error("ECONNREFUSED"))
    await usePortafolioPicker.getState().cargar()
    const st = usePortafolioPicker.getState()
    expect(st.estado).toBe("error")
    expect(st.error).toBe("ECONNREFUSED")
  })

  it("re-abrir dispara un fetch nuevo (RF-15), no reusa la lista anterior", async () => {
    const spy = vi
      .spyOn(api, "listPortafolio")
      .mockResolvedValueOnce({ entradas: [entradasDemo[0]] })
      .mockResolvedValueOnce({ entradas: entradasDemo })
    await usePortafolioPicker.getState().cargar()
    expect(usePortafolioPicker.getState().entradas).toHaveLength(1)
    await usePortafolioPicker.getState().cargar()
    expect(usePortafolioPicker.getState().entradas).toHaveLength(3)
    expect(spy).toHaveBeenCalledTimes(2)
  })

  it("reset vuelve al estado inicial (cargando, vacío, sin error)", async () => {
    vi.spyOn(api, "listPortafolio").mockResolvedValue({ entradas: entradasDemo })
    await usePortafolioPicker.getState().cargar()
    usePortafolioPicker.getState().reset()
    const st = usePortafolioPicker.getState()
    expect(st.estado).toBe("cargando")
    expect(st.entradas).toEqual([])
    expect(st.error).toBeUndefined()
  })
})
