// Domain types of the agnostic component graph (L0), mirroring the Go source of truth in
// internal/domain/{graph,box}.go and arch/contracts/schema/{graph.l0,box.contract}.schema.json.
// Keys are Spanish exactly as the JSON contract emits them (nodos/clase/banda/fase/…).
//
// TEMPORAL (HS-09 Fase B): hand-authored to unblock the map SSOT. These should be GENERATED
// from the JSON schema (quicktype) into shared/api and the entity should wrap them — see
// entities/README and arch/boundaries/fe-transporte-independiente.md. Placed here (not in
// shared/api) so entities/arnes/model stays free of a transport import (dependency-cruiser
// `domain-not-transport`). Fase D reconciles the generated-types seam.

// Clase — the ten CC-native placeable primitives (meta.clase / I-75). ⊥ Banda ⊥ PerfilHarness.
export type Clase =
  | "skill"
  | "subagent"
  | "hook"
  | "rule"
  | "command"
  | "mcp"
  | "plugin"
  | "settings"
  | "output-style"
  | "statusline"

// Banda — the fixed map region a node lives in.
export type Banda =
  | "guardia"
  | "fase"
  | "base"
  | "libreria-expertos"
  | "meta-harness"
  | "marcas-dormidas"
  | "terceros"

// TipoEdge — invoca/escribe are directional (arrow); lee is dotted, no arrow.
export type TipoEdge = "invoca" | "lee" | "escribe"

export type Canal = "beta" | "estable" | "propuesto" | "deprecado"
export type Arquetipo = "pipeline" | "excepcion" | "abierto" | "no-arnesar"
export type PerfilHarness = "T1" | "T2" | "T3"
export type GateTipo = "auto" | "manual" | "parcial" | "none"
export type Procedencia = "medido" | "estimado" | "declarado" | "inferido" | "no-declarado"

export interface Transicion {
  de: string
  a: string
}

// Spine — the FORM of an arnés's work-state machine; concrete values are per-arnés data.
export interface Spine {
  inicial: string
  terminales?: string[]
  estados: string[]
  transiciones?: Transicion[]
}

// Arnes — the harness manifiesto (META de enganche + this arnés's declared fases/spine).
export interface Arnes {
  id?: string
  rol?: string
  proceso?: string
  empresa?: string
  reporta_a: string | null
  canal?: Canal
  marketplace?: string
  fases?: string[]
  spine?: Spine
}

export interface Capability {
  id: string
  what: string
  success: string
}

export interface Input {
  art: string
  de: string
  requerido?: boolean
}

export interface Output {
  art: string
  escritor_unico?: boolean
}

export interface Route {
  a: string
  si?: string
}

export interface Acceptance {
  given: string
  when: string
  then: string
}

export interface Gate {
  tipo: GateTipo
  detalle?: string
  aceptacion?: Acceptance[]
  evidencia?: string
}

export interface Handoff {
  cuando: string
  a: string
}

// Contract — the fused `contract:` block of a process box (doctrina v1). Three axes.
export interface Contract {
  why?: string
  capabilities?: Capability[]
  constraints?: string[]
  non_goals?: string[]
  clase?: Clase
  arquetipo?: Arquetipo
  perfil_harness?: PerfilHarness
  caja: boolean
  fase?: string
  estado?: string
  necesita?: Input[]
  entrega?: Output[]
  ruta?: Route[]
  gate?: Gate
  handoff?: Handoff
}

// Box — a node in the graph (a component of an arnés). Only id/clase/nombre are required.
export interface Box {
  id: string
  clase: Clase
  nombre: string
  banda?: Banda
  fase?: string
  estado?: string
  reporta_a?: string
  canal?: Canal
  fuente_path?: string
  procedencia?: Procedencia
  contract?: Contract
}

export interface Edge {
  de: string
  a: string
  tipo: TipoEdge
}

// Graph — the agnostic component graph of one arnés.
export interface Graph {
  arnes?: Arnes
  nodos: Box[]
  edges?: Edge[]
}
