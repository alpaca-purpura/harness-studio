// Pure per-node visual derivations from REAL L0 data (no transport, no React). These are
// the honest half of the node's marks: `caja`, the spine transition label and the
// "propuesto" badge come straight off the contract/canal — never a fixture. The PROPOSAL
// half (origen/alw) lives in ./proposals, kept separate on purpose (spec §2.3).

import type { Arquetipo, Box, GateTipo } from "./types"

// isCaja — the node is a process box (skill-front of its phase) ⇔ contract.caja===true.
// Mirrors Box.IsCaja() (box.go:186). RF-22.
export function isCaja(box: Box): boolean {
  return box.contract?.caja === true
}

// isPropuesto — the "propuesto" badge derives from the release channel (canal==="propuesto"),
// NOT a mockup-only `prop` field (architecture §5.1). RF-28.
export function isPropuesto(box: Box): boolean {
  return box.canal === "propuesto"
}

// transLabel — the spine transition a box owns ("<entra> → <sale>"), from the REAL
// contract.estado / box.estado (box.go:210). Normalizes the ASCII "->" of the data to the
// "→" of the signed mockup. undefined when the node owns no transition. RF-23.
export function transLabel(box: Box): string | undefined {
  const estado = box.estado ?? box.contract?.estado
  if (!estado) return undefined
  return estado.replace(/\s*->\s*/, " → ")
}

// ARQUETIPO_MARK — the FORM-of-work axis (§8.1, autonomía≠automatización) as a mono char-badge.
// Color does NOT discriminate (the badge uses --muted-foreground); the char + tooltip carry it,
// so the arquetipo never competes with the clase color. From REAL contract.arquetipo. HS-09 Fase 2.
export const ARQUETIPO_MARK: Record<Arquetipo, { char: string; label: string }> = {
  pipeline: { char: "═", label: "pipeline · determinista" },
  excepcion: { char: "≈", label: "excepción · flujo feliz + casos raros" },
  abierto: { char: "✳", label: "abierto · generativo" },
  "no-arnesar": { char: "○", label: "no-arnesar · juicio puro (sale del grafo)" },
}

// GATE_TONE — the eval-gate honesty (A4). A status dot toned by gate.tipo. `none` is a FINDING,
// not a blank: drawn crit + hollow (dotted ring) so the hole is impossible to miss (gris ≠ verde,
// METODOLOGIA §3/§4). `manual` = a human decides (skill-blue). From REAL contract.gate.tipo.
export const GATE_TONE: Record<GateTipo, { color: string; label: string; hollow: boolean }> = {
  auto: { color: "var(--ok)", label: "gate auto — eval automático", hollow: false },
  parcial: { color: "var(--warn)", label: "gate parcial — auto + juicio", hollow: false },
  manual: { color: "var(--c-skill)", label: "gate manual — decide un humano", hollow: false },
  none: { color: "var(--crit)", label: "gate none — SIN eval (hallazgo)", hollow: true },
}

// handleFor — the invocation handle, DERIVED from the class (not a datum): command→id ·
// skill→/id · subagent→@id · hook→"evento" · rule→always-on/condicional (by the PROPOSAL
// `alw`) · rest→id. Mirrors comando() (mockup:327-334). RF-21.
export function handleFor(box: Box, alw?: boolean): string {
  switch (box.clase) {
    case "command":
      return box.id
    case "skill":
      return `/${box.id}`
    case "subagent":
      return `@${box.id}`
    case "hook":
      return "evento"
    case "rule":
      return alw === false ? "condicional" : "always-on"
    default:
      return box.id
  }
}
