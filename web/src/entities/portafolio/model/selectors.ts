// Selectores puros del Portafolio (S1-D6/D7): sin React, sin transporte — importables desde
// cualquier capa (FSD: entities/*/model). Unit-tested por el proyecto vitest `unit`
// (environment node, NO Storybook/browser) — ver selectors.test.ts.

import type { EntradaPortafolio, SaludPortafolio } from "./types"

// saludDe — regla EXACTA de S1-D4 (definida + testeada, G9 «definir regla o quitar»): «no sé»
// ≠ «sano», jamás un verde fabricado.
//   atencion  ⟺ ∃ instalación con deriva "en-deriva" ∨ con aviso ∨ con discrepancias en su origen
//   ok        ⟺ ≥1 instalación y NINGUNA disparó atencion (todas al-hilo, sin aviso ni discrepancias)
//   sin-senal ⟺ el resto (todo deriva-no-evaluable mezclado sin señal de atención, o 0 instalaciones)
export function saludDe(e: EntradaPortafolio): SaludPortafolio {
  const instalaciones = e.instalaciones ?? []

  const necesitaAtencion = instalaciones.some((i) => {
    const discrepancias = i.origen.discrepancias?.length ?? 0
    return i.deriva === "en-deriva" || !!i.aviso || discrepancias > 0
  })
  if (necesitaAtencion) return "atencion"

  const todasAlHilo = instalaciones.length > 0 && instalaciones.every((i) => i.deriva === "al-hilo")
  if (todasAlHilo) return "ok"

  return "sin-senal"
}

// registriesDe — unión de entry.registries ∪ cada instalación.origen.registry, dedup, orden
// estable (primera aparición gana). Vacío ⇒ el caller pinta «desconocido» (G4/BR-3).
export function registriesDe(e: EntradaPortafolio): string[] {
  const vistos = new Set<string>()
  const out: string[] = []
  const agregar = (r: string | undefined) => {
    if (!r || vistos.has(r)) return
    vistos.add(r)
    out.push(r)
  }
  for (const r of e.registries ?? []) agregar(r)
  for (const inst of e.instalaciones ?? []) agregar(inst.origen.registry)
  return out
}

const SIN_EMPRESA = "sin empresa"

// agruparPorEmpresa — lente `empresa` (S1-D8): N:M, una entrada aparece en TANTOS grupos como
// empresas declare (nunca se elige una sola). El grupo «sin empresa» es SIEMPRE el último si
// alguna entrada no declara ninguna — jamás se infiere la empresa del path (G4).
export function agruparPorEmpresa(
  es: EntradaPortafolio[],
): { grupo: string; entradas: EntradaPortafolio[] }[] {
  const orden: string[] = []
  const porEmpresa = new Map<string, EntradaPortafolio[]>()
  const sinEmpresa: EntradaPortafolio[] = []

  for (const e of es) {
    const empresas = e.empresas ?? []
    if (empresas.length === 0) {
      sinEmpresa.push(e)
      continue
    }
    for (const emp of empresas) {
      let grupo = porEmpresa.get(emp)
      if (!grupo) {
        grupo = []
        porEmpresa.set(emp, grupo)
        orden.push(emp)
      }
      grupo.push(e)
    }
  }

  const grupos = orden.map((grupo) => ({ grupo, entradas: porEmpresa.get(grupo) ?? [] }))
  if (sinEmpresa.length > 0) grupos.push({ grupo: SIN_EMPRESA, entradas: sinEmpresa })
  return grupos
}

// filtrarEntradas — buscar (S1-D8): substring case-insensitive sobre id/nombre/descripción.
// Query vacía o solo-espacios ⇒ sin filtro (todas las entradas).
export function filtrarEntradas(es: EntradaPortafolio[], q: string): EntradaPortafolio[] {
  const query = q.trim().toLowerCase()
  if (query === "") return es
  return es.filter((e) => {
    const campos = [e.identidad.id, e.nombre, e.descripcion]
    return campos.some((c) => c?.toLowerCase().includes(query))
  })
}

// idsColisionados — GAP-2/S1-D2 (el índice del Mapa keyea por `identidad.id` pelado, deuda
// viva): dos entradas del Portafolio con el mismo id se pisarían en silencio al observarlas.
// Devuelve SOLO los ids con ≥2 claves — mapa vacío ⇒ ninguna colisión.
export function idsColisionados(es: EntradaPortafolio[]): Map<string, string[]> {
  const porId = new Map<string, string[]>()
  for (const e of es) {
    const claves = porId.get(e.identidad.id) ?? []
    claves.push(e.clave)
    porId.set(e.identidad.id, claves)
  }

  const colisiones = new Map<string, string[]>()
  for (const [id, claves] of porId) {
    if (claves.length >= 2) colisiones.set(id, claves)
  }
  return colisiones
}
