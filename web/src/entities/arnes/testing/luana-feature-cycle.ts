// Rich fixture: a COMPLETE arnés (Luana · ciclo de feature de producto end-to-end) to see the map
// with full information — todas las bandas pobladas, las 10 clases (glyph forma+color), edges de
// los 3 tipos (invoca/lee/escribe), META con reporta_a no-null. `luana` es la empresa de ejemplo
// canónica de los mockups firmados (v3/galaxia). Datos ilustrativos; su función es ejercitar la
// superficie del Mapa con densidad realista. Type-checked contra Graph (story=test en compilación).

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
      terminales: ["operado"],
      estados: [
        "idea",
        "descubierto",
        "especificado",
        "diseñado",
        "construido",
        "revisado",
        "liberado",
        "operado",
      ],
      transiciones: [
        { de: "idea", a: "descubierto" },
        { de: "descubierto", a: "especificado" },
        { de: "especificado", a: "diseñado" },
        { de: "diseñado", a: "construido" },
        { de: "construido", a: "revisado" },
        { de: "revisado", a: "liberado" },
        { de: "liberado", a: "operado" },
        { de: "revisado", a: "construido" },
        { de: "operado", a: "especificado" },
      ],
    },
  },
  nodos: [
    // ── Guardia (hooks transversales) ──
    {
      id: "guard-format",
      clase: "hook",
      nombre: "formato pre-write",
      banda: "guardia",
      canal: "estable",
    },
    {
      id: "guard-pii",
      clase: "hook",
      nombre: "guardia PII/compliance",
      banda: "guardia",
      canal: "estable",
    },
    {
      id: "telemetry-emit",
      clase: "hook",
      nombre: "telemetría de nacimiento",
      banda: "guardia",
      canal: "estable",
    },

    // ── Fase: descubrimiento ──
    {
      id: "discovery-interviewer",
      clase: "subagent",
      nombre: "entrevistar al usuario",
      banda: "fase",
      fase: "descubrimiento",
      estado: "idea -> descubierto",
      canal: "estable",
    },
    {
      id: "need-distiller",
      clase: "skill",
      nombre: "destilar la necesidad",
      banda: "fase",
      fase: "descubrimiento",
      estado: "idea -> descubierto",
      canal: "estable",
    },

    // ── Fase: spec ──
    {
      id: "spec-writer",
      clase: "skill",
      nombre: "escribir el spec",
      banda: "fase",
      fase: "spec",
      estado: "descubierto -> especificado",
      canal: "estable",
      contract: {
        why: "convertir la necesidad en un spec ejecutable y trazable",
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        fase: "spec",
        estado: "descubierto -> especificado",
        necesita: [
          { art: "necesidad destilada", de: "caja:need-distiller", requerido: true },
          { art: "estándar de producto", de: "base:std-producto", requerido: true },
        ],
        entrega: [{ art: "spec.md", escritor_unico: true }],
        ruta: [{ a: "ux-designer", si: "spec verde" }],
        gate: { tipo: "manual", detalle: "revisión de producto del spec" },
      },
    },
    {
      id: "spec-review",
      clase: "command",
      nombre: "/spec-review",
      banda: "fase",
      fase: "spec",
      canal: "estable",
    },

    // ── Fase: diseño ──
    {
      id: "ux-designer",
      clase: "skill",
      nombre: "diseñar la experiencia",
      banda: "fase",
      fase: "diseño",
      estado: "especificado -> diseñado",
      canal: "estable",
    },
    {
      id: "design-critic",
      clase: "subagent",
      nombre: "criticar el diseño",
      banda: "fase",
      fase: "diseño",
      estado: "especificado -> diseñado",
      canal: "beta",
    },

    // ── Fase: build ──
    {
      id: "builder",
      clase: "skill",
      nombre: "construir la feature",
      banda: "fase",
      fase: "build",
      estado: "diseñado -> construido",
      canal: "estable",
      contract: {
        why: "materializar el diseño en código que pasa sus pruebas",
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T3",
        caja: true,
        fase: "build",
        estado: "diseñado -> construido",
        necesita: [{ art: "diseño aprobado", de: "caja:ux-designer", requerido: true }],
        entrega: [{ art: "código + tests", escritor_unico: true }],
        ruta: [{ a: "test-author" }],
        gate: { tipo: "auto", detalle: "la suite de tests pasa en verde" },
      },
    },
    {
      id: "test-author",
      clase: "subagent",
      nombre: "escribir las pruebas",
      banda: "fase",
      fase: "build",
      estado: "diseñado -> construido",
      canal: "estable",
    },

    // ── Fase: review ──
    {
      id: "reviewer",
      clase: "skill",
      nombre: "revisar el build",
      banda: "fase",
      fase: "review",
      estado: "construido -> revisado",
      canal: "estable",
    },
    {
      id: "security-reviewer",
      clase: "subagent",
      nombre: "revisar seguridad",
      banda: "fase",
      fase: "review",
      estado: "construido -> revisado",
      canal: "estable",
    },

    // ── Fase: release ──
    {
      id: "releaser",
      clase: "skill",
      nombre: "promover a producción",
      banda: "fase",
      fase: "release",
      estado: "revisado -> liberado",
      canal: "estable",
    },
    {
      id: "promote",
      clase: "command",
      nombre: "/promote",
      banda: "fase",
      fase: "release",
      canal: "estable",
    },

    // ── Fase: operación ──
    {
      id: "ops-monitor",
      clase: "skill",
      nombre: "monitorear en producción",
      banda: "fase",
      fase: "operación",
      estado: "liberado -> operado",
      canal: "estable",
    },
    {
      id: "incident-triage",
      clase: "subagent",
      nombre: "triage de incidentes",
      banda: "fase",
      fase: "operación",
      estado: "liberado -> operado",
      canal: "estable",
    },

    // ── Base (reglas · knowledge · mcp) ──
    {
      id: "std-producto",
      clase: "rule",
      nombre: "estándar de producto",
      banda: "base",
      canal: "estable",
    },
    {
      id: "domain-glossary",
      clase: "rule",
      nombre: "glosario de dominio",
      banda: "base",
      canal: "estable",
    },
    { id: "api-mcp", clase: "mcp", nombre: "acceso a la API", banda: "base", canal: "estable" },
    { id: "docs-mcp", clase: "mcp", nombre: "base de conocimiento", banda: "base", canal: "beta" },

    // ── Meta-harness (config del arnés — cae en la banda Base) ──
    {
      id: "luana-plugin",
      clase: "plugin",
      nombre: "kit luana",
      banda: "meta-harness",
      canal: "estable",
    },
    {
      id: "guardrails-settings",
      clase: "settings",
      nombre: "permisos por rol",
      banda: "meta-harness",
      canal: "estable",
    },
    {
      id: "product-style",
      clase: "output-style",
      nombre: "estilo de producto",
      banda: "meta-harness",
      canal: "estable",
    },
    {
      id: "health-statusline",
      clase: "statusline",
      nombre: "salud del arnés",
      banda: "meta-harness",
      canal: "estable",
    },
  ],
  edges: [
    // cadena de invocación por las fases
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
    // lecturas de la Base (punteado)
    { de: "spec-writer", a: "std-producto", tipo: "lee" },
    { de: "spec-writer", a: "domain-glossary", tipo: "lee" },
    { de: "need-distiller", a: "docs-mcp", tipo: "lee" },
    { de: "builder", a: "api-mcp", tipo: "lee" },
    { de: "reviewer", a: "std-producto", tipo: "lee" },
    // escritura de un hook a knowledge (accent)
    { de: "guard-pii", a: "domain-glossary", tipo: "escribe" },
  ],
} satisfies Graph
