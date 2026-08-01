import { describe, expect, it } from "vitest"
import { devFullCycle } from "../testing/dev-full-cycle"
import { developerVitaliaActividades } from "../testing/developer-vitalia-actividades"
import {
  cajasSinActividad,
  fasesDeActividad,
  saludDeActividad,
  saludDeGrupo,
  selectActividades,
} from "./actividades"
import type { Actividad, Graph } from "./types"

// Tests de tabla del proyecto `unit` (patrón colocado sancionado, S1-D7). Cubren la mitad
// que una story no ve: los bordes del worst-of y la regla MA-L5 (sin catálogo → todo vacío).

const g: Graph = developerVitaliaActividades
const actividad = (id: string): Actividad => {
  const a = selectActividades(g).find((x) => x.id === id)
  if (!a) throw new Error(`actividad ${id} no está en el fixture`)
  return a
}

describe("selectActividades — catálogo o [] (MA-L5/E8)", () => {
  it("devuelve las 4 actividades del fixture en orden de documento", () => {
    expect(selectActividades(g).map((a) => a.id)).toEqual([
      "historia",
      "bugfix",
      "spike",
      "revisar-capability",
    ])
  })

  it("arnés legacy sin tipos → [] (jamás inventa)", () => {
    expect(selectActividades(devFullCycle)).toEqual([])
  })
})

describe("cajasSinActividad — grupo E6, SOLO con catálogo", () => {
  it("junta las cajas con faceta vacía", () => {
    expect(cajasSinActividad(g).map((n) => n.id)).toEqual(["chrome-devtools-verify", "hipaa-check"])
  })

  it("sin catálogo no hay pregunta: [] aunque haya cajas sin faceta", () => {
    expect(cajasSinActividad(devFullCycle)).toEqual([])
  })
})

describe("fasesDeActividad — el alcance real del procedimiento (E5)", () => {
  it("historia toca las 4 fases", () => {
    expect([...fasesDeActividad(g, actividad("historia"))].sort()).toEqual([
      "construccion",
      "entrega",
      "preparacion",
      "verificacion",
    ])
  })

  it("el spike TERMINA ANTES: solo preparación (el paso sin caja no suma fase)", () => {
    expect([...fasesDeActividad(g, actividad("spike"))]).toEqual(["preparacion"])
  })
})

describe("saludDeActividad — worst-of de hallazgos existentes, sin dato nuevo", () => {
  it("historia = crit (test-all lleva gate:none)", () => {
    expect(saludDeActividad(g, actividad("historia"))).toBe("crit")
  })

  it("bugfix = ok (todas sus cajas con gate real)", () => {
    expect(saludDeActividad(g, actividad("bugfix"))).toBe("ok")
  })

  it("spike = warn (paso «decidir» sin caja, E13)", () => {
    expect(saludDeActividad(g, actividad("spike"))).toBe("warn")
  })

  it("caja referenciada AUSENTE del grafo → warn (hueco visible, no silencio)", () => {
    const rota: Actividad = { id: "x", pasos: [{ paso: "p1", caja: "no-existe" }] }
    expect(saludDeActividad(g, rota)).toBe("warn")
  })

  it("sin pasos (E7) → ok: no hay hallazgo que agregar, el degradado lo dice el chip", () => {
    expect(saludDeActividad(g, actividad("revisar-capability"))).toBe("ok")
  })
})

describe("saludDeGrupo — el dot del chip sin-actividad", () => {
  it("crit si alguna caja del grupo tiene gate:none", () => {
    expect(saludDeGrupo(cajasSinActividad(g))).toBe("crit")
  })

  it("ok con grupo vacío", () => {
    expect(saludDeGrupo([])).toBe("ok")
  })
})
