// Fixture multi-actividad: `developer-vitalia` como quedaría tras MA-T3, calcado de la escena A
// del mockup firmado (docs/product/stories/2026-07-30-definicion-de-arnes/
// mockup-mapa-actividades.html) con los enums REALES del wire (gate auto|manual|parcial|none ·
// arquetipo pipeline|excepcion|abierto — el mockup usaba «eval»/«humano»/«guiado», que no
// existen en el contrato L0). Cubre los escenarios del corte:
//  · 4 actividades: historia (4 pasos, salud crit por test-all gate:none) · bugfix (4 pasos, ok)
//    · spike (2 pasos, «decidir» SIN caja = E13, termina en preparación = E5) ·
//    revisar-capability (sin pasos = E7)
//  · 2 cajas sin actividad (E6): chrome-devtools-verify · hipaa-check
//  · cajas compartidas ×N (MA-L4): dev-team ×3 · builder-backend ×2 · commit-push ×2
// La faceta `actividades` viene ESTAMPADA como la estamparía el loader (CAP-151) — el FE no
// re-deriva (spec §1 AUD-2).

import type { Graph } from "../model/types"

export const developerVitaliaActividades = {
  arnes: {
    id: "developer-vitalia",
    nombre: "Developer · Vitalia",
    rol: "Ingeniería · developer",
    proceso: "desarrollo de producto Vitalia (spec → pr@main)",
    empresas: ["vitalia"],
    reporta_a: null,
    canal: "beta",
    marketplace: "alpacapurpura/prenter-marketplace",
    fases: ["preparacion", "construccion", "verificacion", "entrega"],
    actividades: [
      {
        id: "historia",
        estados: ["idea", "refinando", "listo", "en-curso", "en-revision", "done"],
        cierre: "ratifica_capability",
        pasos: [
          { paso: "tomar-spec", caja: "dev-team", artefacto: "plan.md" },
          { paso: "construir", caja: "builder-backend", artefacto: "código+tests" },
          { paso: "verificar-e2e", caja: "test-all", artefacto: "evidencia-e2e" },
          { paso: "entregar", caja: "commit-push", artefacto: "pr@main" },
        ],
      },
      {
        id: "bugfix",
        estados: ["reportado", "reproducido", "corregido", "verificado"],
        cierre: "regresion_cubierta",
        pasos: [
          { paso: "reproducir", caja: "dev-team", artefacto: "repro.md" },
          { paso: "corregir", caja: "builder-backend", artefacto: "fix+test" },
          { paso: "validar", caja: "gate-runner", artefacto: "veredicto" },
          { paso: "entregar", caja: "commit-push", artefacto: "pr@main" },
        ],
      },
      {
        id: "spike",
        estados: ["abierto", "investigando", "cerrado"],
        cierre: "documenta_decision",
        pasos: [
          { paso: "investigar", caja: "dev-team", artefacto: "hallazgos.md" },
          // E13 — paso sin caja aún: la cadena paso→caja nace opcional (AUD-1).
          { paso: "decidir", artefacto: "decision.md" },
        ],
      },
      // E7 — actividad declarada sin procedimiento aún.
      {
        id: "revisar-capability",
        estados: ["pendiente", "revisado"],
        cierre: "veredicto_registrado",
        pasos: [],
      },
    ],
  },
  nodos: [
    // ── Guardia (hooks reales de vitalia-app, recorte a 2) ──
    {
      id: "contract-guard",
      clase: "hook",
      nombre: "contract-guard · PreToolUse",
      banda: "guardia",
      procedencia: "declarado",
    },
    {
      id: "auto-chain-detect",
      clase: "hook",
      nombre: "auto-chain-detect · Stop",
      banda: "guardia",
      procedencia: "declarado",
    },
    // ── Proceso: cajas del corte developer ──
    {
      id: "dev-team",
      clase: "skill",
      nombre: "dev-team · tomar spec y armar plan",
      banda: "fase",
      fase: "preparacion",
      estado: "listo -> en-curso",
      procedencia: "declarado",
      actividades: ["historia", "bugfix", "spike"],
      contract: {
        why: "Router del build: lee el ready package, arma plan ticket-por-ticket y decide owner.",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        capabilities: [
          {
            id: "CAP-dt-1",
            what: "Convierte spec en plan ejecutable",
            success: "plan con tickets y owners",
          },
        ],
        necesita: [{ art: "spec.md", de: "externo:po" }],
        entrega: [{ art: "plan.md", escritor_unico: true }],
        ruta: [{ a: "builder-backend" }],
        gate: { tipo: "manual", detalle: "plan aprobado antes de construir" },
        handoff: { cuando: "plan listo", a: "construccion" },
      },
    },
    {
      id: "builder-backend",
      clase: "subagent",
      nombre: "builder-backend",
      banda: "fase",
      fase: "construccion",
      estado: "en-curso -> construido",
      procedencia: "declarado",
      actividades: ["historia", "bugfix"],
      contract: {
        why: "Implementa los tickets de backend contra el plan (DDD, TDD).",
        arquetipo: "abierto",
        perfil_harness: "T3",
        caja: true,
        necesita: [{ art: "plan.md", de: "caja:dev-team" }],
        entrega: [{ art: "código+tests", escritor_unico: false }],
        ruta: [{ a: "gate-runner" }],
        gate: { tipo: "auto", detalle: "validators corren en gate-runner" },
        handoff: { cuando: "tickets verdes", a: "verificacion" },
      },
    },
    {
      id: "gate-runner",
      clase: "subagent",
      nombre: "gate-runner · validators",
      banda: "fase",
      fase: "verificacion",
      estado: "construido -> validado",
      procedencia: "declarado",
      actividades: ["bugfix"],
      contract: {
        why: "Corre los validators del ready package hasta GREEN.",
        arquetipo: "pipeline",
        perfil_harness: "T2",
        caja: true,
        necesita: [{ art: "código+tests", de: "caja:builder-backend" }],
        entrega: [{ art: "veredicto", escritor_unico: true }],
        ruta: [{ a: "test-all", si: "GREEN" }],
        gate: { tipo: "auto" },
        handoff: { cuando: "GREEN", a: "e2e" },
      },
    },
    {
      id: "test-all",
      clase: "command",
      nombre: "/test-all · e2e completo",
      banda: "fase",
      fase: "verificacion",
      estado: "validado -> verificado",
      procedencia: "declarado",
      actividades: ["historia"],
      contract: {
        why: "Corre la suite completa (back+front+e2e playwright).",
        arquetipo: "pipeline",
        perfil_harness: "T1",
        caja: true,
        necesita: [{ art: "veredicto", de: "caja:gate-runner" }],
        entrega: [{ art: "evidencia-e2e", escritor_unico: true }],
        ruta: [{ a: "commit-push" }],
        gate: { tipo: "none" },
        handoff: { cuando: "suite verde", a: "entrega" },
      },
    },
    {
      id: "commit-push",
      clase: "skill",
      nombre: "commit-push · delegado",
      banda: "fase",
      fase: "entrega",
      estado: "verificado -> entregado",
      procedencia: "declarado",
      actividades: ["historia", "bugfix"],
      contract: {
        why: "Stage + commit + push con guardrails (git-safety).",
        arquetipo: "pipeline",
        perfil_harness: "T1",
        caja: true,
        necesita: [{ art: "evidencia-e2e", de: "caja:test-all" }],
        entrega: [{ art: "pr@main", escritor_unico: true }],
        gate: { tipo: "manual", detalle: "el merge lo aprueba un humano" },
        handoff: { cuando: "PR publicado", a: "auditor (otro arnés)" },
      },
    },
    // ── Cajas SIN actividad (E6): ningún procedimiento las referencia ──
    {
      id: "chrome-devtools-verify",
      clase: "skill",
      nombre: "chrome-devtools-verify · verificar UI en vivo",
      banda: "fase",
      fase: "verificacion",
      procedencia: "declarado",
      contract: {
        why: "Verificación viva de UI vía Chrome DevTools MCP.",
        arquetipo: "abierto",
        perfil_harness: "T2",
        caja: true,
        entrega: [{ art: "evidencia-visual", escritor_unico: false }],
        gate: { tipo: "none" },
      },
    },
    {
      id: "hipaa-check",
      clase: "skill",
      nombre: "hipaa-check · salvaguardas PHI",
      banda: "fase",
      fase: "verificacion",
      procedencia: "declarado",
      contract: {
        why: "Checklist PHI al tocar módulos médicos.",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        entrega: [{ art: "checklist-phi", escritor_unico: true }],
        gate: { tipo: "manual" },
      },
    },
    // ── Base: rules transversales + knowledge (3 caras D18) ──
    {
      id: "architectural-fitness",
      clase: "rule",
      nombre: "architectural-fitness",
      banda: "base",
      procedencia: "declarado",
    },
    {
      id: "git-safety",
      clase: "rule",
      nombre: "git-safety",
      banda: "base",
      procedencia: "declarado",
    },
    {
      id: "tdd-mandatory",
      clase: "rule",
      nombre: "tdd-mandatory (ruta: historia·bugfix)",
      banda: "base",
      procedencia: "declarado",
    },
    {
      id: "vitalia-design-system",
      clase: "skill",
      nombre: "vitalia-design-system (SSoT UI)",
      banda: "base",
      procedencia: "declarado",
    },
    // ── Librería de expertos (recorte) ──
    {
      id: "backend-expert",
      clase: "skill",
      nombre: "backend-expert",
      banda: "libreria-expertos",
      procedencia: "declarado",
    },
    {
      id: "playwright-expert",
      clase: "skill",
      nombre: "playwright-expert",
      banda: "libreria-expertos",
      procedencia: "declarado",
    },
    // ── Meta-harness ──
    {
      id: "vitalia-settings",
      clase: "settings",
      nombre: "permisos del puesto developer",
      banda: "meta-harness",
      procedencia: "declarado",
    },
  ],
  edges: [
    { de: "dev-team", a: "builder-backend", tipo: "invoca" },
    { de: "builder-backend", a: "gate-runner", tipo: "invoca" },
    { de: "gate-runner", a: "test-all", tipo: "invoca" },
    { de: "test-all", a: "commit-push", tipo: "invoca" },
    { de: "builder-backend", a: "backend-expert", tipo: "lee" },
    { de: "test-all", a: "playwright-expert", tipo: "lee" },
    { de: "builder-backend", a: "vitalia-design-system", tipo: "lee" },
  ],
} satisfies Graph
