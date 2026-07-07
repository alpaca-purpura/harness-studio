// Rich fixture: a COMPLETE arnés (Luana · ciclo de feature de producto end-to-end) — the dense
// example the signed mockup renders (mockups/arnesia-mapa-mvp.html GRAPHS.luana). It exercises
// the whole Map surface: all 7 fases, the 7 process cajas aligned into the horizontal spine, the
// 10 clases (glyph shape+color), the 3 edge types, 10 base rules (5 always-on + 5 conditional),
// mcp/librería/meta/terceros bands. This mirror is the parity fixture of record for shot1..shot6.
//
// Data faithfulness to the map's contract:
//  · a node is a CAJA ⇔ contract.caja===true; only cajas carry `estado` (the spine transition
//    they own — RF-23) so ONLY they render a transition tag. Non-caja fase nodes carry no estado.
//  · the "propuesto" badge derives from canal==="propuesto" (indice-semantico), not a mockup field.
//  · the PROPOSAL facets (origen=del-puesto dashed border, rule alw split) are NOT node fields —
//    they derive from the isolated fixture sets in ../model/proposals (spec §2.3, architecture §5).
// Type-checked against Graph (a compile-time story=test).

import type { Graph } from "../model/types"

export const luanaFeatureCycle = {
  arnes: {
    id: "luana-feature-cycle",
    rol: "Producto · Ciclo de feature end-to-end",
    proceso: "de necesidad del usuario a feature en producción y operada",
    empresa: "luana",
    reporta_a: "luana-plataforma",
    canal: "estable",
    marketplace: "alpacapurpura/prenter-marketplace",
    fases: ["descubrimiento", "spec", "diseño", "build", "review", "release", "operación"],
    spine: {
      inicial: "idea",
      terminales: ["operando"],
      estados: [
        "idea",
        "necesidad",
        "spec",
        "diseño",
        "construido",
        "revisado",
        "liberado",
        "operando",
      ],
      transiciones: [
        { de: "idea", a: "necesidad" },
        { de: "necesidad", a: "spec" },
        { de: "spec", a: "diseño" },
        { de: "diseño", a: "construido" },
        { de: "construido", a: "revisado" },
        { de: "revisado", a: "liberado" },
        { de: "liberado", a: "operando" },
      ],
    },
  },
  nodos: [
    // ── Guardia (hooks transversales) ──
    { id: "guard-format", clase: "hook", nombre: "formato pre-write", banda: "guardia" },
    { id: "guard-pii", clase: "hook", nombre: "guardia PII", banda: "guardia" },
    { id: "telemetry-emit", clase: "hook", nombre: "telemetría", banda: "guardia" },

    // ── Fase: descubrimiento ──
    {
      id: "discovery-interviewer",
      clase: "subagent",
      nombre: "entrevistar usuario",
      banda: "fase",
      fase: "descubrimiento",
    },
    {
      id: "need-distiller",
      clase: "skill",
      nombre: "destilar necesidad",
      banda: "fase",
      fase: "descubrimiento",
      estado: "idea -> necesidad",
      contract: { caja: true },
    },

    // ── Fase: spec ──
    {
      id: "spec-writer",
      clase: "skill",
      nombre: "escribir el spec",
      banda: "fase",
      fase: "spec",
      estado: "necesidad -> spec",
      contract: {
        why: "convertir la necesidad en un spec ejecutable y trazable",
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        fase: "spec",
        estado: "necesidad -> spec",
        necesita: [
          { art: "necesidad destilada", de: "caja:need-distiller", requerido: true },
          { art: "estándar de producto", de: "base:std-producto", requerido: true },
        ],
        entrega: [{ art: "spec.md", escritor_unico: true }],
        ruta: [{ a: "ux-designer", si: "spec verde" }],
        gate: { tipo: "manual", detalle: "revisión de producto del spec" },
      },
    },
    { id: "spec-review", clase: "command", nombre: "/spec-review", banda: "fase", fase: "spec" },

    // ── Fase: diseño ──
    {
      id: "ux-designer",
      clase: "skill",
      nombre: "diseñar experiencia",
      banda: "fase",
      fase: "diseño",
      estado: "spec -> diseño",
      contract: { caja: true },
    },
    {
      id: "design-critic",
      clase: "subagent",
      nombre: "criticar diseño",
      banda: "fase",
      fase: "diseño",
    },

    // ── Fase: build ──
    {
      id: "builder",
      clase: "skill",
      nombre: "construir feature",
      banda: "fase",
      fase: "build",
      estado: "diseño -> construido",
      contract: {
        why: "materializar el diseño en código que pasa sus pruebas",
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T3",
        caja: true,
        fase: "build",
        estado: "diseño -> construido",
        necesita: [{ art: "diseño aprobado", de: "caja:ux-designer", requerido: true }],
        entrega: [{ art: "código + tests", escritor_unico: true }],
        ruta: [{ a: "reviewer" }],
        gate: { tipo: "auto", detalle: "la suite de tests pasa en verde" },
      },
    },
    {
      id: "test-author",
      clase: "subagent",
      nombre: "escribir pruebas",
      banda: "fase",
      fase: "build",
    },

    // ── Fase: review ──
    {
      id: "reviewer",
      clase: "skill",
      nombre: "revisar el build",
      banda: "fase",
      fase: "review",
      estado: "construido -> revisado",
      contract: { caja: true },
    },
    {
      id: "security-reviewer",
      clase: "subagent",
      nombre: "revisar seguridad",
      banda: "fase",
      fase: "review",
    },

    // ── Fase: release ──
    {
      id: "releaser",
      clase: "skill",
      nombre: "promover a prod",
      banda: "fase",
      fase: "release",
      estado: "revisado -> liberado",
      contract: { caja: true },
    },
    { id: "promote", clase: "command", nombre: "/promote", banda: "fase", fase: "release" },

    // ── Fase: operación ──
    {
      id: "ops-monitor",
      clase: "skill",
      nombre: "monitorear prod",
      banda: "fase",
      fase: "operación",
      estado: "liberado -> operando",
      contract: { caja: true },
    },
    {
      id: "incident-triage",
      clase: "subagent",
      nombre: "triage incidentes",
      banda: "fase",
      fase: "operación",
    },

    // ── Base · reglas (10: 5 always-on + 5 conditional, split via ../model/proposals) ──
    { id: "std-producto", clase: "rule", nombre: "estándar de producto", banda: "base" },
    { id: "domain-glossary", clase: "rule", nombre: "glosario de dominio", banda: "base" },
    { id: "pii-policy", clase: "rule", nombre: "política PII", banda: "base" },
    { id: "commit-format", clase: "rule", nombre: "formato de commits", banda: "base" },
    { id: "security-rules", clase: "rule", nombre: "reglas de seguridad", banda: "base" },
    { id: "code-style", clase: "rule", nombre: "estilo de código", banda: "base" },
    { id: "test-policy", clase: "rule", nombre: "política de pruebas", banda: "base" },
    { id: "api-conventions", clase: "rule", nombre: "convenciones de API", banda: "base" },
    { id: "a11y-rules", clase: "rule", nombre: "reglas de accesibilidad", banda: "base" },
    { id: "release-checklist", clase: "rule", nombre: "checklist de release", banda: "base" },

    // ── Base · knowledge & servicios (mcp) ──
    { id: "api-mcp", clase: "mcp", nombre: "acceso a la API", banda: "base" },
    { id: "docs-mcp", clase: "mcp", nombre: "base de conocimiento", banda: "base" },
    {
      id: "indice-semantico",
      clase: "mcp",
      nombre: "índice semántico",
      banda: "base",
      canal: "propuesto",
    },

    // ── Librería de expertos ──
    { id: "react-expert", clase: "skill", nombre: "experto React", banda: "libreria-expertos" },
    {
      id: "a11y-expert",
      clase: "skill",
      nombre: "experto accesibilidad",
      banda: "libreria-expertos",
    },

    // ── Meta-harness (config del arnés) ──
    { id: "luana-plugin", clase: "plugin", nombre: "kit luana", banda: "meta-harness" },
    {
      id: "guardrails-settings",
      clase: "settings",
      nombre: "permisos por rol",
      banda: "meta-harness",
    },
    {
      id: "product-style",
      clase: "output-style",
      nombre: "estilo de producto",
      banda: "meta-harness",
    },
    {
      id: "health-statusline",
      clase: "statusline",
      nombre: "salud del arnés",
      banda: "meta-harness",
    },

    // ── Terceros ──
    { id: "clerk-mcp", clase: "mcp", nombre: "Clerk · auth", banda: "terceros" },
  ],
  edges: [
    // cadena de invocación por las fases (11 · el backbone/spine)
    { de: "discovery-interviewer", a: "need-distiller", tipo: "invoca" },
    { de: "need-distiller", a: "spec-writer", tipo: "invoca" },
    { de: "spec-writer", a: "ux-designer", tipo: "invoca" },
    { de: "ux-designer", a: "design-critic", tipo: "invoca" },
    { de: "design-critic", a: "builder", tipo: "invoca" },
    { de: "builder", a: "test-author", tipo: "invoca" },
    { de: "test-author", a: "reviewer", tipo: "invoca" },
    { de: "reviewer", a: "security-reviewer", tipo: "invoca" },
    { de: "security-reviewer", a: "releaser", tipo: "invoca" },
    { de: "releaser", a: "ops-monitor", tipo: "invoca" },
    { de: "ops-monitor", a: "incident-triage", tipo: "invoca" },
    // lecturas de la Base (punteado, sólo en hover)
    { de: "spec-writer", a: "std-producto", tipo: "lee" },
    { de: "spec-writer", a: "domain-glossary", tipo: "lee" },
    { de: "need-distiller", a: "docs-mcp", tipo: "lee" },
    { de: "builder", a: "api-mcp", tipo: "lee" },
    { de: "builder", a: "react-expert", tipo: "lee" },
    { de: "reviewer", a: "std-producto", tipo: "lee" },
    // escritura de un hook a knowledge (accent, sólo en hover)
    { de: "guard-pii", a: "domain-glossary", tipo: "escribe" },
  ],
} satisfies Graph
