// Artefactos — la proyección del hand-off (franja-artefactos D1–D11, RF-140/142/144).
// Un chip = una REVISIÓN de un artefacto con productor único, o un input externo; TODO se
// DERIVA de necesita[]/entrega[] de los contratos (D1: el artefacto JAMÁS es nodo del L0,
// banda ni clase 11ª). Puro: sin React, sin transporte — el widget solo pinta lo que esto
// devuelve (canvas ⊥ chrome). Espejo FE de la derivación firmada en el mockup v2
// (research/2026-07-07-franja-artefactos/mockup-artefactos.html:420-466).

import type { Edge, Graph, Input } from "./types"

// ArtefactosMode — el toggle de MapBar (RF-143): off = mapa actual idéntico (cero DOM
// extra) · auto = chips solo al seleccionar una caja · todos = siempre visibles.
export type ArtefactosMode = "off" | "auto" | "todos"

// CAP — tope de densidad por gutter (D11c firmada): máx 3 chips visibles; el resto tras
// «+N más». Los relacionados con la selección SALTAN el tope.
export const CAP_GUTTER = 3

export interface ConsumidorChip {
  id: string
  requerido: boolean
}

// ChipArtefacto — un chip derivado. `after` = fase del productor (el gutter donde NACE,
// D11a); `before` = fase del consumidor (externos se pintan en el gutter de entrada de su
// consumidor, D11e). `version` se deriva de la cadena refina (D9c), jamás se declara.
export interface ChipArtefacto {
  id: string
  art: string
  externo: boolean
  origen?: "usuario" | "terceros"
  productor: string | null
  consumidores: ConsumidorChip[]
  path: boolean
  plantilla: boolean
  opaco: boolean
  refina: string | null
  final: boolean
  dead: boolean
  after?: string | undefined
  before?: string | undefined
  version: number
}

const slug = (s: string) =>
  s
    .toLowerCase()
    .replace(/[^a-z0-9@.]+/g, "-")
    .replace(/^-|-$/g, "")

const esExterno = (de: string) => de === "usuario" || de.startsWith("terceros")

// Legible por humanos/LLM → documento; cualquier otra extensión → opaco (pdf/binarios,
// D10/C20: icono relleno, jamás plantilla). Sin path no hay opacidad que derivar.
const TEXT_EXT = new Set(["md", "txt", "json", "yaml", "yml", "csv"])
function esOpaco(path: string | undefined): boolean {
  if (!path) return false
  const dot = path.lastIndexOf(".")
  if (dot < 0) return false
  return !TEXT_EXT.has(path.slice(dot + 1).toLowerCase())
}

// Terminalidad DERIVADA del spine (C15/HS-12, espejo de domain.VerificarDeadEnds): la
// caja cuya transición aterriza en un terminal declarado. Sin spine/terminales no se
// deriva nada (ni final ni dead — honesto, como el check Go que difiere).
function terminales(g: Graph): ReadonlySet<string> | null {
  const t = g.arnes?.spine?.terminales
  if (!t || t.length === 0) return null
  return new Set(t)
}
function esTerminal(estado: string | undefined, term: ReadonlySet<string>): boolean {
  const parts = (estado ?? "").split("->")
  if (parts.length !== 2) return false
  return term.has(parts[1]?.trim() ?? "")
}

// selectArtefactos — deriva los chips del grafo (RF-140). Orden estable: primero los
// externos (por caja consumidora), luego las entregas (por caja productora) — el gutter
// reordena por prioridad al pintar.
export function selectArtefactos(g: Graph): ChipArtefacto[] {
  const cajas = g.nodos.filter((n) => n.contract?.caja)
  const term = terminales(g)

  // consumo: "<productor>::<art>" → consumidores con su `requerido` (D11d).
  const consumo = new Map<string, ConsumidorChip[]>()
  const chips: ChipArtefacto[] = []

  for (const c of cajas) {
    for (const n of c.contract?.necesita ?? []) {
      if (n.de.startsWith("caja:")) {
        const k = `${n.de.slice(5)}::${n.art}`
        const list = consumo.get(k) ?? []
        list.push({ id: c.id, requerido: n.requerido !== false })
        consumo.set(k, list)
      } else if (esExterno(n.de)) {
        chips.push({
          id: `art-ext-${slug(n.art)}-${c.id}`,
          art: n.art,
          externo: true,
          origen: n.de === "usuario" ? "usuario" : "terceros",
          productor: null,
          consumidores: [{ id: c.id, requerido: n.requerido !== false }],
          path: false,
          plantilla: false,
          opaco: false,
          refina: null,
          final: false,
          dead: false,
          before: c.fase,
          version: 1,
        })
      }
      // base:/libreria:/maquinaria:/marcas-dormidas: = componentes, no chips (C5/C6).
    }
  }

  for (const c of cajas) {
    for (const e of c.contract?.entrega ?? []) {
      const cons = consumo.get(`${c.id}::${e.art}`) ?? []
      const final = term !== null && esTerminal(c.estado ?? c.contract?.estado, term)
      chips.push({
        id: `art-${slug(e.art)}-${c.id}`,
        art: e.art,
        externo: false,
        productor: c.id,
        consumidores: cons,
        path: !!e.path,
        plantilla: !!e.plantilla,
        opaco: esOpaco(e.path),
        refina: e.refina ?? null,
        // dead-end (C16): sin consumidor en caja NO terminal — solo derivable con spine.
        final,
        dead: term !== null && cons.length === 0 && !final,
        after: c.fase,
        version: 1,
      })
    }
  }

  derivarVersiones(g, chips)
  return chips
}

// derivarVersiones — v = v(revisión anterior) + 1 siguiendo la cadena refina (D9c);
// visited-set corta ciclos declarados por error (el motor los caza como hallazgo).
function derivarVersiones(g: Graph, chips: ChipArtefacto[]): void {
  const byProdArt = new Map(
    chips.filter((c) => c.productor).map((c) => [`${c.productor}::${c.art}`, c]),
  )
  const nodos = new Map(g.nodos.map((n) => [n.id, n]))
  const fuente = (prodId: string, art: string): Input | undefined =>
    nodos.get(prodId)?.contract?.necesita?.find((n) => n.art === art && n.de.startsWith("caja:"))

  const ver = (c: ChipArtefacto, seen: Set<string>): number => {
    if (!c.refina || !c.productor || seen.has(c.id)) return c.version
    seen.add(c.id)
    const src = fuente(c.productor, c.refina)
    const parent = src ? byProdArt.get(`${src.de.slice(5)}::${c.refina}`) : undefined
    c.version = parent ? ver(parent, seen) + 1 : 2
    return c.version
  }
  for (const c of chips) ver(c, new Set())
}

// ── visibilidad (mockup syncChips, RF-143/144) ─────────────────────────────────────

export interface GutterPlan {
  visibles: ChipArtefacto[]
  ocultos: number
  // El botón se muestra también expandido («− plegar») cuando hay más de CAP elegibles.
  expandible: boolean
}

const relacionado = (c: ChipArtefacto, sel: string | undefined) =>
  !!sel && (c.productor === sel || c.id === sel || c.consumidores.some((k) => k.id === sel))

// planGutter — qué chips de un gutter se pintan: elegible (todos, o relacionado con la
// selección en auto) → prioridad relacionados (saltan el tope D11c) → CAP salvo expandido.
export function planGutter(
  chips: ChipArtefacto[],
  opts: { mode: ArtefactosMode; selectedId?: string | undefined; expanded?: boolean },
): GutterPlan {
  if (opts.mode === "off") return { visibles: [], ocultos: 0, expandible: false }
  const elig = chips.filter((c) => opts.mode === "todos" || relacionado(c, opts.selectedId))
  const rels = elig.filter((c) => relacionado(c, opts.selectedId))
  const shown = new Set(rels)
  for (const c of elig) {
    if (!opts.expanded && shown.size >= CAP_GUTTER) break
    shown.add(c)
  }
  const visibles = elig.filter((c) => shown.has(c))
  return {
    visibles,
    ocultos: elig.length - visibles.length,
    expandible: elig.length > CAP_GUTTER,
  }
}

// ── panel de entrada (D11b, RF-144) ────────────────────────────────────────────────

// RefEntrada — una referencia ↖ compacta: índice navegable de un necesita de largo
// alcance de la caja seleccionada. JAMÁS un segundo chip (el real se ilumina donde vive).
export interface RefEntrada {
  art: string
  productor: string
  opcional: boolean
}

// selectRefsEntrada — los necesita `caja:` de la caja cuyo productor NO vive en la fase
// adyacente anterior (esos ya tienen su chip en el gutter de entrada).
export function selectRefsEntrada(
  g: Graph,
  cajaId: string,
  fasePrevia: string | null,
): RefEntrada[] {
  const caja = g.nodos.find((n) => n.id === cajaId)
  if (!caja?.contract?.caja) return []
  const nodos = new Map(g.nodos.map((n) => [n.id, n]))
  const refs: RefEntrada[] = []
  for (const n of caja.contract.necesita ?? []) {
    if (!n.de.startsWith("caja:")) continue
    const pid = n.de.slice(5)
    const adyacente = fasePrevia !== null && nodos.get(pid)?.fase === fasePrevia
    if (!adyacente) refs.push({ art: n.art, productor: pid, opcional: n.requerido === false })
  }
  return refs
}

// ── edges derivados (RF-142) ───────────────────────────────────────────────────────

// ArtEdge — un edge de la proyección: escribe caja→chip, lee chip→consumidor. `art`
// marca que se pinta SIEMPRE que su chip esté visible (no gated por hover); `opt` = más
// tenue (D11d).
export interface ArtEdge extends Edge {
  art: true
  opt?: boolean
}

// artEdges — los edges de los chips VISIBLES + el set de invoca a suprimir: con el chip
// del hand-off visible, el invoca productor→consumidor se sustituye por escribe+lee.
export function artEdges(
  chips: ChipArtefacto[],
  visibles: ReadonlySet<string>,
): { edges: ArtEdge[]; suprimidos: ReadonlySet<string> } {
  const edges: ArtEdge[] = []
  const suprimidos = new Set<string>()
  for (const c of chips) {
    if (!visibles.has(c.id)) continue
    if (c.productor) {
      edges.push({ de: c.productor, a: c.id, tipo: "escribe", art: true })
      for (const k of c.consumidores) {
        edges.push({ de: c.id, a: k.id, tipo: "lee", art: true, opt: !k.requerido })
        suprimidos.add(`${c.productor}>${k.id}`)
      }
    } else {
      for (const k of c.consumidores) {
        edges.push({ de: c.id, a: k.id, tipo: "lee", art: true, opt: !k.requerido })
      }
    }
  }
  return { edges, suprimidos }
}
