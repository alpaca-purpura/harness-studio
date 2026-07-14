// Unit tests de los selectores puros del Portafolio (proyecto vitest `unit`, S1-D7 — Node, sin
// Chromium). Nombres/casos EXACTOS del plan de implementación §3 T2. Complementa el fitness
// visual (`--project=storybook`), no lo reemplaza.

import { describe, expect, it } from "vitest"
import {
  entradaCanonicaCompleta,
  entradaHarnessEnDeriva,
  entradaProyectoInstaladoProvisional,
} from "../testing/entradas"
import {
  agruparPorEmpresa,
  filtrarEntradas,
  idsColisionados,
  registriesDe,
  saludDe,
} from "./selectors"
import type { EntradaPortafolio, Instalacion } from "./types"

// instalacion — helper mínimo con los defaults "limpios" (al-hilo, sin aviso/discrepancias);
// cada test overridea SOLO lo que le importa a esa rama.
function instalacion(overrides: Partial<Instalacion> = {}): Instalacion {
  return {
    proyecto_path: "~/Proyectos/x",
    install_path: "~/Proyectos/x/.claude/plugins/y",
    tipo: "materializada",
    origen: {},
    deriva: "al-hilo",
    ...overrides,
  }
}

// entrada — helper mínimo; clave/identidad arbitrarias salvo que el test las necesite.
function entrada(overrides: Partial<EntradaPortafolio> = {}): EntradaPortafolio {
  return {
    clave: "clave-x",
    identidad: { id: "x" },
    ...overrides,
  }
}

describe("saludDe", () => {
  it("en-deriva → atencion", () => {
    const e = entrada({ instalaciones: [instalacion({ deriva: "en-deriva" })] })
    expect(saludDe(e)).toBe("atencion")
  })

  it("aviso → atencion", () => {
    const e = entrada({ instalaciones: [instalacion({ aviso: "sin record de instalación" })] })
    expect(saludDe(e)).toBe("atencion")
  })

  it("discrepancias → atencion", () => {
    const e = entrada({
      instalaciones: [
        instalacion({ origen: { discrepancias: ["registry: valores en conflicto"] } }),
      ],
    })
    expect(saludDe(e)).toBe("atencion")
  })

  it("todo al-hilo limpio → ok", () => {
    const e = entrada({
      instalaciones: [instalacion({ deriva: "al-hilo" }), instalacion({ deriva: "al-hilo" })],
    })
    expect(saludDe(e)).toBe("ok")
  })

  it("todo no-evaluable → sin-senal", () => {
    const e = entrada({
      instalaciones: [instalacion({ deriva: "deriva-no-evaluable" })],
    })
    expect(saludDe(e)).toBe("sin-senal")
    // grounding real: la entrada (b) del E2E de Slice 0 es exactamente este caso.
    expect(saludDe(entradaProyectoInstaladoProvisional)).toBe("sin-senal")
  })

  it("0 instalaciones → sin-senal", () => {
    expect(saludDe(entrada())).toBe("sin-senal")
    expect(saludDe(entrada({ instalaciones: [] }))).toBe("sin-senal")
  })

  it("grounding real: en-deriva del E2E de Slice 0 → atencion", () => {
    expect(saludDe(entradaHarnessEnDeriva)).toBe("atencion")
  })

  it("grounding real: aviso+discrepancias de la fixture sintética → atencion", () => {
    expect(saludDe(entradaCanonicaCompleta)).toBe("atencion")
  })
})

describe("agruparPorEmpresa", () => {
  it("N:M — una entrada con 2 empresas se duplica en los 2 grupos", () => {
    const e = entrada({ clave: "n-m", empresas: ["alpacapurpura", "acme"] })
    const grupos = agruparPorEmpresa([e])
    expect(grupos.map((g) => g.grupo)).toEqual(["alpacapurpura", "acme"])
    expect(grupos[0]?.entradas).toEqual([e])
    expect(grupos[1]?.entradas).toEqual([e])
  })

  it("«sin empresa» va SIEMPRE al final si alguna entrada no declara ninguna", () => {
    const conEmpresa = entrada({ clave: "con-empresa", empresas: ["alpacapurpura"] })
    const sinEmpresa = entrada({ clave: "sin-empresa" })
    const grupos = agruparPorEmpresa([sinEmpresa, conEmpresa])
    expect(grupos.map((g) => g.grupo)).toEqual(["alpacapurpura", "sin empresa"])
    expect(grupos.at(-1)?.grupo).toBe("sin empresa")
    expect(grupos.at(-1)?.entradas).toEqual([sinEmpresa])
  })

  it("sin ninguna entrada con empresa ⇒ un único grupo «sin empresa»", () => {
    const grupos = agruparPorEmpresa([entradaProyectoInstaladoProvisional])
    expect(grupos).toEqual([
      { grupo: "sin empresa", entradas: [entradaProyectoInstaladoProvisional] },
    ])
  })
})

describe("registriesDe", () => {
  it("unión entry.registries ∪ instalaciones[].origen.registry, dedup", () => {
    const e = entrada({
      registries: ["github.com/acme/acme-cli"],
      instalaciones: [
        instalacion({ origen: { registry: "github.com/acme/acme-cli" } }), // duplicado
        instalacion({ origen: { registry: "github.com/otro/repo" } }),
        instalacion({ origen: {} }), // sin registry: no aporta
      ],
    })
    expect(registriesDe(e)).toEqual(["github.com/acme/acme-cli", "github.com/otro/repo"])
  })

  it("sin ningún registry ⇒ array vacío", () => {
    expect(registriesDe(entradaProyectoInstaladoProvisional)).toEqual([])
  })

  it("grounding real: la entrada (a) del E2E resuelve su registry (entry canonicalizado + origen crudo, distintos strings, ambos visibles)", () => {
    expect(registriesDe(entradaHarnessEnDeriva)).toEqual([
      "github.com/alpacapurpura/prenter-marketplace",
      "alpacapurpura/prenter-marketplace",
    ])
  })
})

describe("filtrarEntradas", () => {
  const acme = entrada({ clave: "acme", identidad: { id: "acme-cli" }, nombre: "Acme CLI" })
  const harness = entrada({
    clave: "harness",
    identidad: { id: "harness" },
    descripcion: "arnés dogfood",
  })
  const es = [acme, harness]

  it("substring case-insensitive sobre id", () => {
    expect(filtrarEntradas(es, "ACME")).toEqual([acme])
  })

  it("substring case-insensitive sobre nombre", () => {
    expect(filtrarEntradas(es, "cli")).toEqual([acme])
  })

  it("substring case-insensitive sobre descripción", () => {
    expect(filtrarEntradas(es, "dogfood")).toEqual([harness])
  })

  it("sin match ⇒ array vacío", () => {
    expect(filtrarEntradas(es, "no-existe")).toEqual([])
  })

  it("query vacía ⇒ todas las entradas", () => {
    expect(filtrarEntradas(es, "")).toEqual(es)
    expect(filtrarEntradas(es, "   ")).toEqual(es)
  })
})

describe("idsColisionados", () => {
  it("colisión detectada: dos claves con el mismo identidad.id", () => {
    const a = entrada({ clave: "clave-a", identidad: { id: "harness", scope: "proyecto-a" } })
    const b = entrada({ clave: "clave-b", identidad: { id: "harness", scope: "proyecto-b" } })
    const colisiones = idsColisionados([a, b])
    expect(colisiones.get("harness")).toEqual(["clave-a", "clave-b"])
    expect(colisiones.size).toBe(1)
  })

  it("ids únicos ⇒ mapa vacío", () => {
    const a = entrada({ clave: "clave-a", identidad: { id: "harness" } })
    const b = entrada({ clave: "clave-b", identidad: { id: "otro" } })
    expect(idsColisionados([a, b]).size).toBe(0)
  })
})
