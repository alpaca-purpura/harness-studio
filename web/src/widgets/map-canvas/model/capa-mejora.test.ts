import { describe, expect, it } from "vitest"
import { devFullCycle } from "@/entities/arnes"
import type { ResumenTelemetria } from "@/entities/telemetria"
import { CAJAS_ILUSTRATIVAS, RESUMEN_ILUSTRATIVO } from "@/entities/telemetria"
import {
  coberturaEsParcial,
  denominadorDeBusqueda,
  estadosDeLaFranja,
  hayDatosAtribuibles,
  vistaCapaMejora,
} from "./capa-mejora"

// Tests de tabla de la derivación de la capa. Cubren lo que ninguna story de componente puede:
// **las condiciones que deciden qué muestra cada bloque**, que es donde vivieron los cinco
// defectos de composición del paquete (D24 + los 4 críticos de `auditoria-tramo-b.md`).

/** El wire de `s2-instrumentado`: `corridas: 0` por construcción, todo lo demás poblado. */
const S2_INSTRUMENTADO: ResumenTelemetria = {
  ...RESUMEN_ILUSTRATIVO,
  escenario: "s2-instrumentado",
  // `COUNT(DISTINCT corrida_id)` y `corrida_id` es NULL fuera de S1 (consultas.go:95).
  corridas: 0,
  turnos: 58,
  costo_reportado_micros: 1_920_000,
  cobertura: {
    esperados: null,
    exacta: 58,
    por_hash: 0,
    por_proceso: 0,
    sin_dato: 3,
    no_llegaron: null,
  },
}

const NUNCA_CORRIO: ResumenTelemetria = {
  ...RESUMEN_ILUSTRATIVO,
  corridas: 0,
  // D26.4 — «nunca corrió» EXIGE no haber corrido nunca. Con historial fuera de la ventana
  // el estado correcto es el 1b, y decir «nunca corrió» sería falso.
  ultima_corrida: null,
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
  it("🔴 s2-instrumentado tiene datos aunque `corridas` sea 0", () => {
    // Éste es el crítico C-2: el escenario que PARIDAD declara MAYORITARIO. Mirando `corridas`,
    // la franja decía «nunca corrió» con USD 1,92 medidos y el candado escondía el entregable.
    expect(S2_INSTRUMENTADO.corridas).toBe(0)
    expect(hayDatosAtribuibles(S2_INSTRUMENTADO)).toBe(true)
  })

  it("un arnés que nunca corrió no tiene datos", () => {
    expect(hayDatosAtribuibles(NUNCA_CORRIO)).toBe(false)
    expect(hayDatosAtribuibles(null)).toBe(false)
    expect(hayDatosAtribuibles(undefined)).toBe(false)
  })

  it("turnos que llegaron pero NO se pudieron atribuir no son datos atribuibles (D24)", () => {
    // Intenté agregar un fallback por `turnos > 0` «para cubrir S2» y rompió esto: sin
    // atribución no hay a qué colgar un hallazgo, así que «los detectores corrieron y no
    // encontraron nada» seguiría siendo falso. No hacía falta — en s2 la cobertura viene
    // poblada y la primera condición ya alcanza.
    expect(
      hayDatosAtribuibles({
        ...NUNCA_CORRIO,
        turnos: 4,
        costo_reportado_micros: 12_000,
        cobertura: { ...NUNCA_CORRIO.cobertura, sin_dato: 4 },
      }),
    ).toBe(false)
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
  it("cae a turnos fuera de S1, donde `corridas` es 0 por construcción", () => {
    expect(denominadorDeBusqueda(S2_INSTRUMENTADO)).toBe(58)
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

  it("🔴 en s2-instrumentado los cinco bloques SÍ muestran el dato", () => {
    const v = vistaCapaMejora({
      resumen: S2_INSTRUMENTADO,
      cajas,
      graph: devFullCycle,
      activa: true,
    })
    expect(v.hayDatos).toBe(true)
    // La franja recibe el resumen ⇒ puede alcanzar su rama `s2-instrumentado`, que antes era
    // inalcanzable porque la rama `resumen == null` ganaba primero.
    expect(v.resumenParaFranja).not.toBeNull()
    expect(v.resumenParaFranja?.escenario).toBe("s2-instrumentado")
    // Y la lista de puntos —el entregable del paquete— no se esconde.
    expect(v.mejoraPorNodo?.size).toBe(cajas.length)
    expect(v.propsPorNodo?.get("spec-writer")?.cifraUsd).toBe("1,92")
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

// D26.4 — los cuatro estados que estaban STORIADOS Y MUERTOS. Cada uno sale de un hecho del
// wire, y cada uno se prueba con su control negativo al lado: un derivador que enciende
// siempre pasaría un test de un solo lado.
describe("estadosDeLaFranja — los cuatro estados que la app no alcanzaba (D26.4)", () => {
  it("1b · con historial fuera de la ventana NO dice «nunca corrió»", () => {
    const r = { ...NUNCA_CORRIO, ultima_corrida: "2026-07-01T10:00:00Z" }
    expect(estadosDeLaFranja(r, undefined).ultimaCorridaFuera).toBe("01/07/2026")
    // Control negativo: sin historial, el estado correcto es el 1 y este no aplica.
    expect(estadosDeLaFranja(NUNCA_CORRIO, undefined).ultimaCorridaFuera).toBeUndefined()
    // Y con corridas DENTRO de la ventana tampoco: no hay nada que explicar.
    expect(
      estadosDeLaFranja({ ...RESUMEN_ILUSTRATIVO, corridas: 61 }, undefined).ultimaCorridaFuera,
    ).toBeUndefined()
  })

  it("4 · el dinero que puso nuestro catálogo se declara como tal", () => {
    const r = { ...RESUMEN_ILUSTRATIVO, costo_reportado_micros: null, costo_calculado_micros: 4820 }
    expect(estadosDeLaFranja(r, undefined).costoDelCatalogo).toBe(true)
    // Control negativo: con costo del runtime, la cifra NO es nuestra estimación.
    expect(estadosDeLaFranja(RESUMEN_ILUSTRATIVO, undefined).costoDelCatalogo).toBeUndefined()
  })

  it("5 · un catálogo que nunca se refrescó lo dice, con su versión", () => {
    const r = {
      ...RESUMEN_ILUSTRATIVO,
      catalogo: { ...RESUMEN_ILUSTRATIVO.catalogo, refrescado: null },
    }
    expect(estadosDeLaFranja(r, undefined).catalogoSinRefrescar).toBe(r.catalogo.version)
    // Control negativo: refrescado ⇒ no hay aviso.
    expect(estadosDeLaFranja(r, "2026-07-26T00:00:00Z").catalogoSinRefrescar).toBeUndefined()
  })

  it("runtime no soportado · se DICE, no se muestra un cero", () => {
    const r = { ...RESUMEN_ILUSTRATIVO, runtime: "aider", runtime_soportado: false }
    expect(estadosDeLaFranja(r, undefined).runtimeNoSoportado).toBe("aider")
    // Control negativo doble: soportado, y sin runtime declarado.
    expect(estadosDeLaFranja(RESUMEN_ILUSTRATIVO, undefined).runtimeNoSoportado).toBeUndefined()
    expect(
      estadosDeLaFranja({ ...RESUMEN_ILUSTRATIVO, runtime: "" }, undefined).runtimeNoSoportado,
    ).toBeUndefined()
  })
})
