// PROPUESTA (HS-09 · Gate 1 · spec §2.3 · architecture §5 · cementado en Fase D).
// Estado as-code REAL de las dos facetas: `origen` YA es campo L0
// (graph.l0.schema.json $defs.nodo.origen, estandar|del-puesto) pero lo ESTAMPA el
// provisioner al instanciar ③ — hasta que ese provisioning exista, ningún nodo lo trae
// poblado y el Mapa lo dibuja desde estos sets fixture AISLADOS, rotulados PROPUESTA.
// `alw` NUNCA será campo L0: decisión firmada (knowledge/rules.md L2.6) — se DERIVA de
// `fuente_path`/`paths:` de la regla; este set fixture es el stand-in hasta que el
// loader real derive. Aislar los sets aquí (no en el canvas) es la mitigación de
// architecture §9.1: cuando el dato real llegue, `isDelPuesto` pasa a leer node.origen,
// `alwFor` pasa a derivar, y estos sets desaparecen.

// PROPOSED_DEL_PUESTO — ids whose `origen` is "del-puesto" (filled at onboarding with
// rol·proceso·empresa knowledge) → dashed left border. The rest are "estandar" (from the
// agnostic kit) → solid border. Mirrors DEL_PUESTO (mockup:229).
const PROPOSED_DEL_PUESTO: ReadonlySet<string> = new Set([
  "domain-glossary",
  "std-producto",
  "api-conventions",
  "api-mcp",
  "docs-mcp",
  "indice-semantico",
  "react-expert",
  "a11y-expert",
  "clerk-mcp",
  "luana-plugin",
  "guardrails-settings",
  "discovery-interviewer",
  "ux-designer",
])

// The rule always-on axis, as a tri-state (mockup `alw` per rule): a rule id in
// PROPOSED_ALWAYS is "siempre en contexto" (CLAUDE.md); in PROPOSED_CONDITIONAL is "carga
// condicional" (paths:); absent from both is unknown (renders always-on handle, condicional
// group — faithful to the mockup's `box.alw===false?…:…` / `!r.alw` split).
const PROPOSED_ALWAYS: ReadonlySet<string> = new Set([
  "std-producto",
  "domain-glossary",
  "pii-policy",
  "commit-format",
  "security-rules",
])
const PROPOSED_CONDITIONAL: ReadonlySet<string> = new Set([
  "code-style",
  "test-policy",
  "api-conventions",
  "a11y-rules",
  "release-checklist",
])

// isDelPuesto — PROPOSAL: does this node id carry the `origen=del-puesto` facet?
export function isDelPuesto(id: string): boolean {
  return PROPOSED_DEL_PUESTO.has(id)
}

// alwFor — PROPOSAL: the always-on tri-state of a rule id (true=siempre, false=condicional,
// undefined=unknown). Mirrors the per-node `alw` of the mockup.
export function alwFor(id: string): boolean | undefined {
  if (PROPOSED_ALWAYS.has(id)) return true
  if (PROPOSED_CONDITIONAL.has(id)) return false
  return undefined
}
