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
  agruparPorMarketplace,
  agruparPorProyecto,
  filtrarEntradas,
  filtrarPorMarketplace,
  filtrarPorSalud,
  gruposCandidatosDe,
  identificadorDe,
  idsColisionados,
  marketplacesDisponibles,
  registriesDe,
  saludDe,
} from "./selectors"
import type { Candidato, EntradaPortafolio, Instalacion } from "./types"

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

describe("agruparPorProyecto", () => {
  it("N:M — una entrada instalada en 2 proyectos se duplica en los 2 grupos", () => {
    const e = entrada({
      clave: "n-m",
      instalaciones: [
        instalacion({ proyecto_path: "~/Proyectos/a" }),
        instalacion({ proyecto_path: "~/Proyectos/b" }),
      ],
    })
    const grupos = agruparPorProyecto([e])
    expect(grupos.map((g) => g.grupo)).toEqual(["~/Proyectos/a", "~/Proyectos/b"])
    expect(grupos[0]?.entradas).toEqual([e])
    expect(grupos[1]?.entradas).toEqual([e])
  })

  it("misma entrada con 2 instalaciones EN el mismo proyecto no se duplica en el grupo", () => {
    const e = entrada({
      clave: "mismo-proyecto",
      instalaciones: [
        instalacion({ proyecto_path: "~/Proyectos/a", install_path: "~/Proyectos/a/x" }),
        instalacion({ proyecto_path: "~/Proyectos/a", install_path: "~/Proyectos/a/y" }),
      ],
    })
    const grupos = agruparPorProyecto([e])
    expect(grupos).toEqual([{ grupo: "~/Proyectos/a", entradas: [e] }])
  })

  it("«sin proyecto instalado» va SIEMPRE al final si alguna entrada no tiene instalaciones", () => {
    const conProyecto = entrada({ clave: "con-proyecto", instalaciones: [instalacion()] })
    const sinProyecto = entrada({ clave: "sin-proyecto" })
    const grupos = agruparPorProyecto([sinProyecto, conProyecto])
    expect(grupos.at(-1)?.grupo).toBe("sin proyecto instalado")
    expect(grupos.at(-1)?.entradas).toEqual([sinProyecto])
  })

  it("sin ninguna entrada instalada ⇒ un único grupo «sin proyecto instalado»", () => {
    const soloCanonico = entrada({ clave: "solo-canonico" })
    const grupos = agruparPorProyecto([soloCanonico])
    expect(grupos).toEqual([{ grupo: "sin proyecto instalado", entradas: [soloCanonico] }])
  })
})

describe("agruparPorMarketplace", () => {
  it("N:M — una entrada con 2 registries se duplica en los 2 grupos", () => {
    const e = entrada({ clave: "n-m", registries: ["github.com/a/a", "github.com/b/b"] })
    const grupos = agruparPorMarketplace([e])
    expect(grupos.map((g) => g.grupo)).toEqual(["github.com/a/a", "github.com/b/b"])
    expect(grupos[0]?.entradas).toEqual([e])
    expect(grupos[1]?.entradas).toEqual([e])
  })

  it("«origen desconocido» va SIEMPRE al final si alguna entrada no resuelve registry", () => {
    const conRegistry = entrada({ clave: "con-registry", registries: ["github.com/a/a"] })
    const sinRegistry = entrada({ clave: "sin-registry" })
    const grupos = agruparPorMarketplace([sinRegistry, conRegistry])
    expect(grupos.at(-1)?.grupo).toBe("origen desconocido")
    expect(grupos.at(-1)?.entradas).toEqual([sinRegistry])
  })

  it("sin ningún registry en ninguna entrada ⇒ un único grupo «origen desconocido»", () => {
    const grupos = agruparPorMarketplace([entradaProyectoInstaladoProvisional])
    expect(grupos).toEqual([
      { grupo: "origen desconocido", entradas: [entradaProyectoInstaladoProvisional] },
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

describe("filtrarPorSalud", () => {
  const sana = entrada({
    clave: "sana",
    instalaciones: [instalacion({ deriva: "al-hilo" })],
  })
  const atencion = entrada({
    clave: "atencion",
    instalaciones: [instalacion({ deriva: "en-deriva" })],
  })
  const es = [sana, atencion]

  it("set vacío ⇒ sin filtro (todas las entradas)", () => {
    expect(filtrarPorSalud(es, new Set())).toEqual(es)
  })

  it("un solo valor seleccionado ⇒ solo esas entradas", () => {
    expect(filtrarPorSalud(es, new Set(["atencion"]))).toEqual([atencion])
  })

  it("multi-select: 2 valores ⇒ unión", () => {
    expect(filtrarPorSalud(es, new Set(["ok", "atencion"]))).toEqual(es)
  })

  it("sin match ⇒ array vacío", () => {
    expect(filtrarPorSalud(es, new Set(["sin-senal"]))).toEqual([])
  })
})

describe("filtrarPorMarketplace", () => {
  const a = entrada({ clave: "a", registries: ["github.com/a/a"] })
  const b = entrada({ clave: "b", registries: ["github.com/b/b"] })
  const sinRegistry = entrada({ clave: "sin-registry" })
  const es = [a, b, sinRegistry]

  it("set vacío ⇒ sin filtro (todas las entradas)", () => {
    expect(filtrarPorMarketplace(es, new Set())).toEqual(es)
  })

  it("un registry seleccionado ⇒ solo las entradas que lo tienen", () => {
    expect(filtrarPorMarketplace(es, new Set(["github.com/a/a"]))).toEqual([a])
  })

  it("multi-select: 2 registries ⇒ unión, sin-registry queda afuera", () => {
    expect(filtrarPorMarketplace(es, new Set(["github.com/a/a", "github.com/b/b"]))).toEqual([a, b])
  })
})

describe("marketplacesDisponibles", () => {
  it("valores distintos, orden de primera aparición, dedup entre entradas", () => {
    const a = entrada({ clave: "a", registries: ["github.com/a/a", "github.com/b/b"] })
    const b = entrada({ clave: "b", registries: ["github.com/b/b"] })
    expect(marketplacesDisponibles([a, b])).toEqual(["github.com/a/a", "github.com/b/b"])
  })

  it("sin ningún registry ⇒ array vacío", () => {
    expect(marketplacesDisponibles([entradaProyectoInstaladoProvisional])).toEqual([])
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

// candidato — helper mínimo para los selectores de S1-D26 (Candidato NO tiene fixture en
// testing/entradas.ts por diseño — resultado no-persistido, mismo criterio que T6).
function candidato(
  overrides: Omit<Partial<Candidato>, "instalacion"> & { instalacion?: Partial<Instalacion> },
): Candidato {
  const { instalacion: instOverrides, ...resto } = overrides
  return {
    clave: "clave-x",
    identidad: { id: "x" },
    instalacion: instalacion(instOverrides),
    ...resto,
  }
}

describe("identificadorDe", () => {
  it("id manda cuando existe", () => {
    expect(identificadorDe({ id: "harness", scope: "algo" })).toBe("harness")
  })

  it("sin id ⇒ el scope discrimina (RN-IDENT-2 — ruta relativa o remote del proyecto)", () => {
    expect(identificadorDe({ id: "", scope: "comunify" })).toBe("comunify")
  })

  it("sin id ni scope ⇒ «(sin id)» honesto, último recurso", () => {
    expect(identificadorDe({ id: "" })).toBe("(sin id)")
  })
})

describe("gruposCandidatosDe", () => {
  const raiz = "/home/u/Proyectos/mono"

  it("monorepo real: raíz primero, una subcarpeta por grupo, ocultos y fuera-de-árbol a la raíz", () => {
    const enRaiz = candidato({
      clave: "c-raiz",
      identidad: { id: "", scope: "github.com/acme/mono" },
      instalacion: { proyecto_path: raiz, install_path: raiz, tipo: "proyecto-instalado" },
    })
    const enSub = candidato({
      clave: "c-sub",
      identidad: { id: "", scope: "comunify" },
      instalacion: {
        proyecto_path: raiz,
        install_path: `${raiz}/comunify`,
        tipo: "proyecto-instalado",
      },
    })
    const enSubProfundo = candidato({
      clave: "c-sub2",
      identidad: { id: "", scope: "apps/fitflow" },
      instalacion: {
        proyecto_path: raiz,
        install_path: `${raiz}/apps/fitflow`,
        tipo: "proyecto-instalado",
      },
    })
    const materializadaNivelProyecto = candidato({
      clave: "c-mat",
      identidad: { id: "bar-cli" },
      instalacion: { proyecto_path: raiz, install_path: `${raiz}/.claude/plugins/bar-cli` },
    })
    const cacheCc = candidato({
      clave: "c-cc",
      identidad: { id: "harness" },
      instalacion: {
        proyecto_path: raiz,
        install_path: "/home/u/.claude/plugins/cache/mkt/harness/1.0.0",
        tipo: "referenciada-cc",
      },
    })
    const avisoSinDir = candidato({
      clave: "c-aviso",
      identidad: { id: "commit-commands" },
      instalacion: { proyecto_path: raiz, install_path: "", aviso: "sin record de instalación" },
    })

    const grupos = gruposCandidatosDe([
      enSub,
      enRaiz,
      enSubProfundo,
      materializadaNivelProyecto,
      cacheCc,
      avisoSinDir,
    ])
    expect(grupos.map((g) => g.grupo)).toEqual(["proyecto (raíz)", "comunify/", "apps/"])
    expect(grupos[0]?.candidatos.map((c) => c.clave)).toEqual([
      "c-raiz",
      "c-mat",
      "c-cc",
      "c-aviso",
    ])
    expect(grupos[1]?.candidatos.map((c) => c.clave)).toEqual(["c-sub"])
    expect(grupos[2]?.candidatos.map((c) => c.clave)).toEqual(["c-sub2"])
  })

  it("proyecto simple (todo en raíz) ⇒ un único grupo", () => {
    const unico = candidato({ instalacion: { proyecto_path: raiz, install_path: raiz } })
    expect(gruposCandidatosDe([unico]).map((g) => g.grupo)).toEqual(["proyecto (raíz)"])
  })

  it("0 candidatos ⇒ 0 grupos (nada inventado)", () => {
    expect(gruposCandidatosDe([])).toEqual([])
  })
})
