import { describe, expect, it } from "vitest"
import montos from "../testing/montos.golden.json"
import {
  ariaTendencia,
  direccionTendencia,
  ETIQUETA_CONFIANZA,
  entero,
  etiquetaConfianza,
  etiquetaDetector,
  marcaPrincipal,
  pct,
  segmentosCobertura,
  TITULO_CONFIANZA,
  tonoConfianza,
  totalCobertura,
  usd,
} from "./selectors"
import type { Cobertura } from "./types"

// Tests de tabla del proyecto `unit` (S1-D7, patrón colocado sancionado). Cubren lo que una
// story no puede: los bordes aritméticos del formateo y **el candado 2 de D18** — que el copy
// canónico de confianza es UNA sola fuente y coincide con `design.md` §7.3, literal.

describe("usd — el único formateador de dinero (RF-281)", () => {
  it("formatea con coma decimal y dos decimales", () => {
    expect(usd(1_920_000)).toBe("1,92")
  })

  it("agrupa el millar con espacio fino, jamás con el formato EN", () => {
    const s = usd(1_234_560_000) as string
    expect(s).toMatch(/^1[\s ]234,56$/u)
    expect(s).not.toContain(".")
  })

  it("NO redondea a 0,00 un monto que existe (la mentira barata)", () => {
    expect(usd(4000)).toBe("0,004")
    expect(usd(400)).toBe("0,0004")
    expect(usd(4)).toBe("0,000004")
  })

  it("el cero REAL sí se dice cero — es un dato válido, no una ausencia", () => {
    expect(usd(0)).toBe("0,00")
  })

  it("null y undefined devuelven null: la ausencia la resuelve el componente", () => {
    expect(usd(null)).toBeNull()
    expect(usd(undefined)).toBeNull()
  })

  it("un negativo lleva signo menos tipográfico, no un guion ASCII", () => {
    expect(usd(-1_920_000)).toBe("−1,92")
  })
})

describe("pct y entero", () => {
  it("pct redondea a entero, sin decimales", () => {
    expect(pct(0.4)).toBe("40 %")
    expect(pct(0.444)).toBe("44 %")
  })
  it("pct de null es null — jamás «0 %»", () => {
    expect(pct(null)).toBeNull()
    expect(pct(undefined)).toBeNull()
  })
  it("entero agrupa el millar y respeta el cero", () => {
    expect(entero(17536)).toMatch(/^17[\s ]536$/u)
    expect(entero(0)).toBe("0")
    expect(entero(null)).toBeNull()
  })
})

describe("copy de confianza — candado 2 de D18 (una sola fuente, literal de design §7.3)", () => {
  it("exacta no tiene etiqueta ni title: la AUSENCIA de marca es la señal", () => {
    expect(ETIQUETA_CONFIANZA.exacta).toBeNull()
    expect(TITULO_CONFIANZA.exacta).toBeNull()
    expect(etiquetaConfianza("exacta")).toBeNull()
    expect(tonoConfianza("exacta")).toBe("ninguno")
  })

  it("los literales son EXACTAMENTE los de design.md §7.3", () => {
    expect(ETIQUETA_CONFIANZA["por-hash"]).toBe("por huella")
    expect(ETIQUETA_CONFIANZA["por-proceso"]).toBe("por proceso")
    expect(TITULO_CONFIANZA["por-hash"]).toBe(
      "Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA.",
    )
    expect(TITULO_CONFIANZA["por-proceso"]).toBe(
      "Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla.",
    )
  })

  it("ningún caso se pinta como «aproximada» — taparía tres cosas distintas bajo una", () => {
    for (const v of [...Object.values(ETIQUETA_CONFIANZA), ...Object.values(TITULO_CONFIANZA)]) {
      if (v !== null) expect(v).not.toMatch(/aproximad/i)
    }
  })

  it("los tres títulos visibles son tres strings DISTINTOS (H-13)", () => {
    const ts = Object.values(TITULO_CONFIANZA).filter((t): t is string => t !== null)
    expect(ts).toHaveLength(3)
    expect(new Set(ts).size).toBe(3)
  })
})

describe("cobertura", () => {
  const c: Cobertura = {
    esperados: 61,
    exacta: 44,
    por_hash: 9,
    por_proceso: 5,
    sin_dato: 3,
    no_llegaron: 0,
  }

  it("el total es la suma de los cuatro cubos — su propio denominador", () => {
    expect(totalCobertura(c)).toBe(61)
  })

  it("una categoría en cero NO se devuelve: no se dibuja ni se nombra", () => {
    expect(segmentosCobertura(c)).toHaveLength(4)
    expect(segmentosCobertura({ ...c, por_proceso: 0 })).toHaveLength(3)
    expect(segmentosCobertura({ ...c, por_proceso: 0 }).map((s) => s.id)).not.toContain(
      "por-proceso",
    )
  })

  it("los segmentos salen en orden de calidad descendente", () => {
    expect(segmentosCobertura(c).map((s) => s.id)).toEqual([
      "exacta",
      "por-hash",
      "por-proceso",
      "sin-dato",
    ])
  })
})

describe("tendencia", () => {
  it("con menos de dos puntos devuelve null — «estable» sería una afirmación no medida", () => {
    expect(direccionTendencia(undefined)).toBeNull()
    expect(direccionTendencia([])).toBeNull()
    expect(direccionTendencia([100])).toBeNull()
  })

  it("clasifica alza / baja / estable con un umbral del 5 %", () => {
    expect(direccionTendencia([100, 200])).toBe("alza")
    expect(direccionTendencia([200, 100])).toBe("baja")
    expect(direccionTendencia([100, 102])).toBe("estable")
  })

  it("el aria-label lleva la dirección en palabras y el conteo real", () => {
    expect(ariaTendencia("alza", 5)).toBe("Tendencia en alza en las últimas 5 corridas.")
    expect(ariaTendencia("baja", 3)).toBe("Tendencia a la baja en las últimas 3 corridas.")
  })
})

describe("detectores", () => {
  it("cada detector del MVP tiene nombre corto (jamás un ⚠ sin nombre)", () => {
    expect(etiquetaDetector("b1-rewarm-por-ttl")).toBe("re-warm TTL")
    expect(etiquetaDetector("p1-caja-que-consume-y-se-rechaza")).toBe("rechazo en gate")
    expect(etiquetaDetector("b3-cambio-de-modelo-invalida-cache")).toBe("modelo cambiado")
    expect(etiquetaDetector("b6-sesion-abandonada")).toBe("sesión abandonada")
    expect(etiquetaDetector("b2-costo-de-la-rotacion")).toBe("rotación cara")
  })

  it("un detector desconocido devuelve su id, nunca una cadena vacía", () => {
    expect(etiquetaDetector("z9-inventado")).toBe("z9-inventado")
  })

  it("marcaPrincipal toma la primera — el wire ya ordena por ahorro descendente", () => {
    expect(marcaPrincipal(undefined)).toBeNull()
    expect(marcaPrincipal([])).toBeNull()
    expect(marcaPrincipal([{ n: 1 }, { n: 2 }])).toEqual({ n: 1 })
  })
})

describe("un solo formato de dinero de los dos lados del wire (RF-281 · D25)", () => {
  // El contrafactual de una tarjeta lo redacta el dominio Go —con el monto YA formateado
  // adentro de la frase— y el resto de la superficie lo formatea acá con `usd()`. Si los dos
  // formateadores se separan, la misma cifra se escribe distinto en la misma tarjeta.
  //
  // Un test de tabla por lado no alcanzaría: dos tablas distintas se editan por separado y
  // divergen sin que nada se ponga rojo. Por eso la tabla es **una sola** y vive en el árbol.
  it.each(montos)("$micros → $texto ($porque)", ({ micros, texto }) => {
    expect(usd(micros)).toBe(texto)
  })

  it("cubre los bordes que RF-281 nombra, no solo el caso feliz", () => {
    expect(montos.length).toBeGreaterThanOrEqual(6)
    // El cero real, el sub-centavo, la agrupación de miles y el negativo tienen que estar:
    // sin ellos la tabla pasaría con un formateador que redondea a «0,00».
    expect(montos.some((c) => c.micros === 0)).toBe(true)
    expect(montos.some((c) => c.micros > 0 && c.micros < 10_000)).toBe(true)
    expect(montos.some((c) => c.micros > 1_000_000_000)).toBe(true)
    expect(montos.some((c) => c.micros < 0)).toBe(true)
  })
})
