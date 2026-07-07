// Pure per-node visual derivations from REAL L0 data (no transport, no React). These are
// the honest half of the node's marks: `caja`, the spine transition label and the
// "propuesto" badge come straight off the contract/canal — never a fixture. The PROPOSAL
// half (origen/alw) lives in ./proposals, kept separate on purpose (spec §2.3).

import type { Box } from "./types"

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
