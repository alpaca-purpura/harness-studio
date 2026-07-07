// PROPUESTA (HS-09 · Gate 1 · spec §2.3 · architecture §5). Two map facets are NOT L0
// fields yet — doctrine is SILENT on `origen`, and the rule `alw` (always-on) axis lives
// only as display (UX.md:534, no schema field). The Map draws them as PROPOSAL from these
// ISOLATED fixture sets — never read off a node as if it were data — and labels them
// PROPUESTA in the help panel, until they are ratified as-code (schema + Box). Isolating
// the sets here (not in the canvas) is the mitigation of architecture §9.1.
//
// These ids are demo-fixture data (Luana + dogfood). When `origen`/`alw` land as real L0
// fields, `isDelPuesto`/`alwFor` derive from the node and these sets disappear.

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
