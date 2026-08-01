// Derivaciones puras del catálogo multi-actividad (MA-T4/T5, spec 2026-07-30-definicion-de-arnes).
// El catálogo `arnes.actividades[]` y la faceta `box.actividades` llegan YA derivados por el
// loader al indexar (MA-T1b / CAP-151): acá NUNCA se re-deriva la faceta — solo se proyecta
// (qué cajas quedaron sin actividad, qué fases toca un procedimiento, qué salud agrega).
// Sin catálogo (arnés legacy, MA-L5/E8) todo devuelve vacío y el Mapa queda como hoy.
//
// La salud es HONESTA (spec §3): worst-of de los hallazgos FE ya existentes de sus cajas
// (gate:none crit · no-reconocido warn · paso sin caja / caja ausente warn) — la misma fuente
// del inspector, jamás una cifra tecleada ni un dato nuevo.

import { isCaja } from "./node-view"
import type { Actividad, Box, Graph } from "./types"

// SaludActividad — el agregado worst-of que colorea el dot del chip N0.
export type SaludActividad = "ok" | "warn" | "crit"

const peor = (a: SaludActividad, b: SaludActividad): SaludActividad => {
  if (a === "crit" || b === "crit") return "crit"
  if (a === "warn" || b === "warn") return "warn"
  return "ok"
}

// selectActividades — el catálogo del arnés, o [] (arnés sin tipos declarados, E1/E8).
export function selectActividades(g: Graph): Actividad[] {
  return g.arnes?.actividades ?? []
}

// cajasSinActividad — las cajas que NINGÚN procedimiento referencia (grupo `sin-actividad`,
// E6 — visible, insumo de poda §6). SOLO existe con catálogo declarado: sin tipos no hay
// pregunta que responder (MA-L5). La faceta viene del loader (box.actividades), no se re-deriva.
export function cajasSinActividad(g: Graph): Box[] {
  if (selectActividades(g).length === 0) return []
  return g.nodos.filter((n) => isCaja(n) && (n.actividades ?? []).length === 0)
}

// fasesDeActividad — el Set de fases que el procedimiento TOCA (las fases de las cajas de sus
// pasos). Alimenta E5: las fases fuera del set se atenúan enteras — se VE que el spike
// termina antes. Pasos sin caja (E13) no suman fase.
export function fasesDeActividad(g: Graph, actividad: Actividad): ReadonlySet<string> {
  const fases = new Set<string>()
  for (const p of actividad.pasos ?? []) {
    if (!p.caja) continue
    const n = g.nodos.find((x) => x.id === p.caja)
    if (n?.fase) fases.add(n.fase)
  }
  return fases
}

// saludDeActividad — worst-of derivado SOLO del grafo (sin dato nuevo, igual que el mockup):
// paso sin caja o caja ausente del grafo → warn · caja no-reconocido → warn ·
// caja con gate:none → crit (A4: el hueco de eval es un hallazgo, no un blanco).
export function saludDeActividad(g: Graph, actividad: Actividad): SaludActividad {
  let s: SaludActividad = "ok"
  for (const p of actividad.pasos ?? []) {
    if (!p.caja) {
      s = peor(s, "warn")
      continue
    }
    const n = g.nodos.find((x) => x.id === p.caja)
    if (!n) {
      s = peor(s, "warn")
      continue
    }
    if (n.clase === "no-reconocido") s = peor(s, "warn")
    if (n.contract?.gate?.tipo === "none") s = peor(s, "crit")
  }
  return s
}

// saludDeGrupo — el mismo worst-of para un grupo suelto de cajas (el chip `sin-actividad`).
export function saludDeGrupo(cajas: readonly Box[]): SaludActividad {
  let s: SaludActividad = "ok"
  for (const n of cajas) {
    if (n.contract?.gate?.tipo === "none") s = peor(s, "crit")
    else if (n.clase === "no-reconocido") s = peor(s, "warn")
  }
  return s
}
