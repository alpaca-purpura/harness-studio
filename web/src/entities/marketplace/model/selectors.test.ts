// Unit tests de los PRESENTADORES puros de Marketplace (proyecto vitest `unit` — Node, sin
// Chromium; S1-D7). Capa **F** del plan de pruebas del paquete
// 2026-07-23-portafolio-agregar-marketplace. Complementa el fitness visual
// (`--project=storybook`), no lo reemplaza.
//
// Lo que estos tests NO prueban a propósito: el CÁLCULO de la situación. Eso vive en el dominio
// Go (`CalcularSituacion`/`AccionDeSituacion`, capa D del plan) y viaja resuelto en el wire
// (design.md §2 C1). Acá se prueba presentación: qué texto y qué tono sale de un veredicto dado.

import { describe, expect, it } from "vitest"
import {
  catNull,
  catPrenter,
  catVacio,
  marketplacesDemo,
  mkNoLeido,
  mkOficialReferencia,
  mkPrenterPropio,
  mkSinAcceso,
  mkUrlNoResuelve,
  situacionesLas6,
} from "../testing/marketplaces"
import {
  accionDeFila,
  esNavegable,
  etiquetaDeSituacion,
  filtrarEntradasCatalogo,
  filtrarPorSituacion,
  ordenarMarketplaces,
  situacionesDisponibles,
  textoDeLectura,
  tonoDeSituacion,
} from "./selectors"
import type { EntradaCatalogo, SituacionCatalogo, TipoSituacion } from "./types"

// AHORA fijo: los textos de lectura tienen que ser DETERMINISTAS (sin Intl, sin reloj real) —
// por eso `ahora` es un parámetro y no un `new Date()` adentro del selector.
const AHORA = new Date("2026-07-25T14:11:33Z")

function situacion(s: SituacionCatalogo): SituacionCatalogo {
  return s
}

describe("etiquetaDeSituacion", () => {
  it("nombra las 6 ramas con los literales del mockup firmado", () => {
    expect(etiquetaDeSituacion(situacion({ tipo: "no-lo-tengo" }))).toBe("no lo tengo")
    expect(etiquetaDeSituacion(situacion({ tipo: "al-hilo" }))).toBe("lo tengo · al hilo")
    expect(etiquetaDeSituacion(situacion({ tipo: "no-comparable", motivo: "x" }))).toBe(
      "no comparable",
    )
  })

  it("muestra las DOS versiones del wire cuando hay divergencia (jamás una cifra tecleada)", () => {
    expect(
      etiquetaDeSituacion(
        situacion({ tipo: "mi-copia-adelantada", mia: "0.5.4", estante: "0.5.3" }),
      ),
    ).toBe("mi copia adelantada — estante v0.5.3 · tu canónico v0.5.4")
    expect(
      etiquetaDeSituacion(
        situacion({ tipo: "estante-adelantado", mia: "0.2.0", estante: "0.3.1" }),
      ),
    ).toBe("el estante adelantado — estante v0.3.1 · tu canónico v0.2.0")
  })

  it("degrada al rótulo pelado si el wire no mandó las versiones (no inventa un «v?»)", () => {
    expect(etiquetaDeSituacion(situacion({ tipo: "estante-adelantado" }))).toBe(
      "el estante adelantado",
    )
  })

  it("pluraliza las instalaciones en deriva con la cuenta del wire", () => {
    expect(etiquetaDeSituacion(situacion({ tipo: "instalaciones-en-deriva", cuantas: 1 }))).toBe(
      "1 instalación en deriva",
    )
    expect(etiquetaDeSituacion(situacion({ tipo: "instalaciones-en-deriva", cuantas: 3 }))).toBe(
      "3 instalaciones en deriva",
    )
  })
})

describe("tonoDeSituacion", () => {
  it("solo `al-hilo` es `ok`", () => {
    expect(tonoDeSituacion(situacion({ tipo: "al-hilo" }))).toBe("ok")
  })

  it("las tres ramas accionables son `atencion`", () => {
    expect(tonoDeSituacion(situacion({ tipo: "mi-copia-adelantada" }))).toBe("atencion")
    expect(tonoDeSituacion(situacion({ tipo: "estante-adelantado" }))).toBe("atencion")
    expect(tonoDeSituacion(situacion({ tipo: "instalaciones-en-deriva" }))).toBe("atencion")
  })

  it("`no-comparable` y `no-lo-tengo` son `sin-senal` — «no sé» NUNCA se pinta como «sano»", () => {
    expect(tonoDeSituacion(situacion({ tipo: "no-comparable", motivo: "x" }))).toBe("sin-senal")
    expect(tonoDeSituacion(situacion({ tipo: "no-lo-tengo" }))).toBe("sin-senal")
    // regla dura de spec §1: ninguna rama que no sea al-hilo puede compartir el tono de ok.
    for (const tipo of ["no-comparable", "no-lo-tengo"] as TipoSituacion[]) {
      expect(tonoDeSituacion(situacion({ tipo }))).not.toBe("ok")
    }
  })
})

describe("textoDeLectura", () => {
  it("dice CUÁNDO se leyó, en pasos deterministas (BR-3)", () => {
    expect(textoDeLectura(mkPrenterPropio.lectura, AHORA)).toBe("leído hace 4 min")
    expect(textoDeLectura(mkOficialReferencia.lectura, AHORA)).toBe("leído hace 2 días")
    expect(
      textoDeLectura({ tipo: "leido", cuando: "2026-07-25T14:11:10Z", entradas: 1 }, AHORA),
    ).toBe("leído hace menos de 1 min")
    expect(
      textoDeLectura({ tipo: "leido", cuando: "2026-07-25T11:11:33Z", entradas: 1 }, AHORA),
    ).toBe("leído hace 3 h")
    expect(
      textoDeLectura({ tipo: "leido", cuando: "2026-07-24T13:11:33Z", entradas: 1 }, AHORA),
    ).toBe("leído hace 1 día")
  })

  it("nunca inventa una fecha: sin `cuando` lo dice, y con basura no crashea", () => {
    expect(textoDeLectura({ tipo: "leido" }, AHORA)).toBe("leído (sin fecha registrada)")
    expect(textoDeLectura({ tipo: "leido", cuando: "no-una-fecha" }, AHORA)).toBe(
      "leído hace fecha desconocida",
    )
  })

  it("los 3 degradados llevan su MOTIVO completo (BR-4: nunca un degradado mudo)", () => {
    expect(textoDeLectura(mkNoLeido.lectura, AHORA)).toBe("no leído aún")
    expect(textoDeLectura(mkSinAcceso.lectura, AHORA)).toBe(
      "sin acceso — gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual",
    )
    expect(textoDeLectura(mkUrlNoResuelve.lectura, AHORA)).toBe(
      "url no resuelve — marketplace.json inválido: unexpected end of JSON input (offset 4096)",
    )
  })

  it("con caché viejo + degradado actual dice LAS DOS cosas (BR-3 + BR-4 juntas)", () => {
    expect(
      textoDeLectura(
        {
          tipo: "sin-acceso",
          cuando: "2026-07-23T14:07:33Z",
          entradas: 1,
          motivo: "dial tcp: lookup github.com: no such host",
        },
        AHORA,
      ),
    ).toBe("leído hace 2 días · ahora sin acceso — dial tcp: lookup github.com: no such host")
  })
})

describe("esNavegable / accionDeFila", () => {
  it("solo una fila LEÍDA navega al catálogo (spec §4.1.4, BR-4)", () => {
    expect(esNavegable(mkPrenterPropio)).toBe(true)
    expect(esNavegable(mkOficialReferencia)).toBe(true)
    expect(esNavegable(mkNoLeido)).toBe(false)
    expect(esNavegable(mkSinAcceso)).toBe(false)
    expect(esNavegable(mkUrlNoResuelve)).toBe(false)
  })

  it("la acción de la fila distingue «nunca leí» de «ya falló»", () => {
    expect(accionDeFila(mkPrenterPropio)).toBe("ver-catalogo")
    expect(accionDeFila(mkNoLeido)).toBe("leer-catalogo")
    expect(accionDeFila(mkSinAcceso)).toBe("reintentar")
    expect(accionDeFila(mkUrlNoResuelve)).toBe("reintentar")
  })
})

describe("ordenarMarketplaces", () => {
  it("propios legibles → propios no legibles → referencia, y no muta el array entrante", () => {
    const antes = [...marketplacesDemo]
    const orden = ordenarMarketplaces(marketplacesDemo).map((m) => m.nombre)
    expect(orden).toEqual([
      "prenter-marketplace",
      "nordia-plugins-rrhh",
      "vitalia-arneses",
      "caveman",
      "claude-plugins-official",
    ])
    expect(marketplacesDemo).toEqual(antes)
  })
})

describe("filtrarEntradasCatalogo", () => {
  const entradas = catPrenter.entradas as EntradaCatalogo[]

  it("query vacía o solo-espacios ⇒ sin filtro", () => {
    expect(filtrarEntradasCatalogo(entradas, "")).toHaveLength(2)
    expect(filtrarEntradasCatalogo(entradas, "   ")).toHaveLength(2)
  })

  it("busca por nombre, descripción, versión y `source.crudo` (case-insensitive)", () => {
    expect(filtrarEntradasCatalogo(entradas, "BETA").map((e) => e.nombre)).toEqual(["harness-beta"])
    expect(filtrarEntradasCatalogo(entradas, "pre-promotion").map((e) => e.nombre)).toEqual([
      "harness-beta",
    ])
    expect(filtrarEntradasCatalogo(entradas, "0.5.3")).toHaveLength(2)
    expect(filtrarEntradasCatalogo(entradas, "./plugins/harness")).toHaveLength(2)
    expect(filtrarEntradasCatalogo(entradas, "zzz-no-existe")).toHaveLength(0)
  })
})

describe("filtrarPorSituacion / situacionesDisponibles", () => {
  it("set vacío ⇒ sin filtro; con valores, acota sin reagrupar", () => {
    expect(filtrarPorSituacion(situacionesLas6, new Set())).toHaveLength(6)
    expect(
      filtrarPorSituacion(situacionesLas6, new Set<TipoSituacion>(["no-comparable"])).map(
        (e) => e.nombre,
      ),
    ).toEqual(["legacy-onboarding"])
    expect(
      filtrarPorSituacion(
        situacionesLas6,
        new Set<TipoSituacion>(["no-lo-tengo", "instalaciones-en-deriva"]),
      ),
    ).toHaveLength(2)
  })

  it("solo ofrece las situaciones PRESENTES, en orden canónico", () => {
    expect(situacionesDisponibles(situacionesLas6)).toEqual([
      "no-lo-tengo",
      "al-hilo",
      "mi-copia-adelantada",
      "estante-adelantado",
      "instalaciones-en-deriva",
      "no-comparable",
    ])
    const soloDos = situacionesLas6.filter((e) =>
      ["no-comparable", "al-hilo"].includes(e.situacion.tipo),
    )
    expect(situacionesDisponibles(soloDos)).toEqual(["al-hilo", "no-comparable"])
    expect(situacionesDisponibles([])).toEqual([])
  })
})

describe("entradas: null ≠ [] (BR-4, design.md §2 C4)", () => {
  it("`null` es «no sé» y `[]` es «leí y no declara ninguno» — el tipo los distingue", () => {
    expect(catNull.entradas).toBeNull()
    expect(catNull.lectura.motivo).toBeTruthy()
    expect(catVacio.entradas).toEqual([])
    expect(catVacio.lectura.tipo).toBe("leido")
    expect(catVacio.lectura.entradas).toBe(0)
    // la fixture del degradado sin lectura NO trae cuenta de entradas: no hay nada que contar.
    expect(catNull.lectura.entradas).toBeUndefined()
  })
})

describe("las acciones del wire son la autoridad (BR-10 + §13.11)", () => {
  it("7 de las 8 celdas siguen `habilitada: false`; la única abierta es propio × no-lo-tengo", () => {
    const habilitadas = situacionesLas6.filter((e) => e.accion.habilitada)
    expect(habilitadas.map((e) => e.situacion.tipo)).toEqual(["no-lo-tengo"])
    expect(habilitadas[0]?.accion.verbo).toBe("traer-canonico")
  })

  it("cada acción deshabilitada trae su motivo LITERAL (el tooltip no se inventa en el FE)", () => {
    const porVerbo = new Map(situacionesLas6.map((e) => [e.accion.verbo, e.accion.motivo]))
    expect(porVerbo.get("publicar")).toBe(
      "Publicar se construye en su propio paquete (ítem 3 del outcome)",
    )
    expect(porVerbo.get("actualizar-mi-copia")).toBe(
      "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)",
    )
    expect(porVerbo.get("reparar")).toBe(
      "Reparar se construye en su propio paquete (ítem 5 del outcome)",
    )
  })
})
