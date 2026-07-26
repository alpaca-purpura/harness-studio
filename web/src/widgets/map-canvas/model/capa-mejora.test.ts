import { describe, expect, it } from "vitest"
import { devFullCycle } from "@/entities/arnes"
import type { ResumenTelemetria } from "@/entities/telemetria"
import { CAJAS_ILUSTRATIVAS, RESUMEN_ILUSTRATIVO } from "@/entities/telemetria"
import {
  coberturaEsParcial,
  denominadorDeBusqueda,
  hayDatosAtribuibles,
  vistaCapaMejora,
} from "./capa-mejora"

// Tests de tabla de la derivación de la capa. Cubren lo que ninguna story de componente puede:
// **las condiciones que deciden qué muestra cada bloque**, que es donde vivieron los cinco
// defectos de composición del paquete (D24 + los 4 críticos de `auditoria-tramo-b.md`).

const NUNCA_CORRIO: ResumenTelemetria = {
  ...RESUMEN_ILUSTRATIVO,
  corridas: 0,
  turnos: 0,
  sesiones: 0,
  cajas: 0,
  costo_reportado_micros: null,
  costo_calculado_micros: null,
  cobertura: {
    esperados: null,
    exacta: 0,
    por_hash: 0,
    por_proceso: 0,
    sin_dato: 0,
    no_llegaron: null,
  },
}

describe("hayDatosAtribuibles — C-2: `corridas` NO es proxy de «hay datos»", () => {
  it("un arnés que nunca corrió no tiene datos", () => {
    expect(hayDatosAtribuibles(NUNCA_CORRIO)).toBe(false)
    expect(hayDatosAtribuibles(null)).toBe(false)
    expect(hayDatosAtribuibles(undefined)).toBe(false)
  })

  it("un costo en 0 NO es medición: 0 no es «se gastó cero», es que no hay nada", () => {
    expect(hayDatosAtribuibles({ ...NUNCA_CORRIO, costo_reportado_micros: 0 })).toBe(false)
  })

  it("⚠️ la confianza del agregado NO sirve como condición (es PeorConfianza)", () => {
    // Una sola corrida sin atribuir deja el agregado en `sin-dato` con 60 perfectas al lado.
    const casiPerfecto: ResumenTelemetria = { ...RESUMEN_ILUSTRATIVO, confianza: "sin-dato" }
    expect(hayDatosAtribuibles(casiPerfecto)).toBe(true)
  })
})

describe("denominadorDeBusqueda — «corrieron sobre N …» no puede decir 0", () => {
  it("usa corridas cuando el contador sirve", () => {
    expect(denominadorDeBusqueda(RESUMEN_ILUSTRATIVO)).toBe(61)
  })
  it("sin resumen es 0 — pero ahí la sección no se dibuja", () => {
    expect(denominadorDeBusqueda(null)).toBe(0)
  })
})

describe("coberturaEsParcial — umbral declarado, no escondido", () => {
  it("el juego coherente del mockup (3 de 61 sin atribuir) no dispara el aviso", () => {
    expect(coberturaEsParcial(RESUMEN_ILUSTRATIVO)).toBe(false)
  })
  it("2 de 5 sin atribuir (40 %) sí lo dispara", () => {
    expect(
      coberturaEsParcial({
        ...RESUMEN_ILUSTRATIVO,
        corridas: 5,
        cobertura: {
          esperados: 5,
          exacta: 3,
          por_hash: 0,
          por_proceso: 0,
          sin_dato: 2,
          no_llegaron: null,
        },
      }),
    ).toBe(true)
  })
  it("sin datos no hay cobertura parcial que avisar", () => {
    expect(coberturaEsParcial(NUNCA_CORRIO)).toBe(false)
  })
})

describe("vistaCapaMejora — los cinco bloques desde una sola fuente", () => {
  const cajas = CAJAS_ILUSTRATIVAS

  it("🔴 sin datos, NINGÚN bloque afirma nada: ni franja, ni canvas, ni carril, ni lista", () => {
    // C-2 · si el estado 1 manda, el canvas tampoco puede seguir cobrando dinero.
    const v = vistaCapaMejora({ resumen: NUNCA_CORRIO, cajas, graph: devFullCycle, activa: true })
    expect(v.hayDatos).toBe(false)
    expect(v.resumenParaFranja).toBeNull()
    expect(v.mejoraPorNodo).toBeUndefined()
    expect(v.totalesPorFase).toBeUndefined()
    expect(v.propsPorNodo).toBeUndefined()
  })

  it("con la capa apagada no deriva nada: el DOM vuelve a ser el de hoy", () => {
    const v = vistaCapaMejora({
      resumen: RESUMEN_ILUSTRATIVO,
      cajas,
      graph: devFullCycle,
      activa: false,
    })
    expect(v.mejoraPorNodo).toBeUndefined()
    expect(v.motivosPorNodo).toBeUndefined()
    expect(v.propsPorNodo).toBeUndefined()
  })

  it("los nodos que no son caja reciben su motivo, y ninguna cifra", () => {
    const v = vistaCapaMejora({
      resumen: RESUMEN_ILUSTRATIVO,
      cajas,
      graph: devFullCycle,
      activa: true,
    })
    const regla = v.propsPorNodo?.get("std-spec")
    expect(regla?.motivoSinDato).toBe("sin dato atribuible — una regla no consume por sí misma")
    expect(regla?.cifraUsd).toBeUndefined()
  })

  it("un carril sin cajas con costo da `null`, no 0 (RF-244)", () => {
    const v = vistaCapaMejora({
      resumen: RESUMEN_ILUSTRATIVO,
      cajas: [],
      graph: devFullCycle,
      activa: true,
    })
    for (const total of v.totalesPorFase?.values() ?? []) expect(total).toBeNull()
  })
})
