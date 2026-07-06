// Recorded fixture: the real dogfood arnés (dogfood/dev-full-cycle.graph.json), transcribed
// as a typed constant so stories are deterministic AND the fixture is type-checked against the
// graph types (a compile-time story=test). If the JSON changes, this must be re-synced — Fase F
// wires the live loader; until then this mirror is the map's fixture of record.
//
// NOT the transport path (that is api.getGraph, Hito 1 live). This lives under entities/*/testing
// so it never ships in the app bundle path.

import type { Graph } from "../model/types"

export const devFullCycle = {
  arnes: {
    id: "dev-full-cycle",
    rol: "Ingeniería · Desarrollo full-cycle",
    proceso: "desarrollo de software end-to-end (idea → released)",
    empresa: "alpacapurpura",
    reporta_a: null,
    canal: "beta",
    marketplace: "alpacapurpura/prenter-marketplace",
    fases: ["spec", "build", "review", "release"],
    spine: {
      inicial: "idea",
      terminales: ["released"],
      estados: ["idea", "spec", "build", "review", "released"],
      transiciones: [
        { de: "idea", a: "spec" },
        { de: "spec", a: "build" },
        { de: "build", a: "review" },
        { de: "review", a: "released" },
        { de: "review", a: "build" },
      ],
    },
  },
  nodos: [
    {
      id: "spec-writer",
      clase: "skill",
      nombre: "escribir el spec",
      banda: "fase",
      fase: "spec",
      estado: "idea -> spec",
      canal: "beta",
      fuente_path: "dogfood/skills/spec-writer.SKILL.md",
      procedencia: "declarado",
      contract: {
        why: "convertir una idea conversada en un spec ejecutable que blinde la deriva",
        capabilities: [
          {
            id: "CAP-01",
            what: "destilar la idea en capacidades con criterio de éxito",
            success: "cada capability tiene un success verificable",
          },
          {
            id: "CAP-02",
            what: "emitir spec.md as-code",
            success: "spec.md valida contra el schema de spec",
          },
        ],
        constraints: ["no inventa requisitos que el usuario no confirmó"],
        non_goals: ["no escribe código", "no diseña la arquitectura"],
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        fase: "spec",
        estado: "idea -> spec",
        necesita: [
          { art: "idea del usuario (conversación grill)", de: "usuario", requerido: true },
          { art: "estándar de spec", de: "base:std-spec", requerido: true },
        ],
        entrega: [{ art: "spec.md", escritor_unico: true }],
        ruta: [
          { a: "builder", si: "gate del spec verde" },
          { a: "humano", si: "la idea no converge en 3 vueltas" },
        ],
        gate: {
          tipo: "manual",
          detalle: "revisión humana del spec contra la intención declarada",
          aceptacion: [
            {
              given: "un spec.md emitido",
              when: "el humano lo revisa",
              then: "cada capability tiene criterio de éxito y no hay non_goal violado",
            },
          ],
          evidencia: "registro de aprobación del spec (telemetría de nacimiento)",
        },
        handoff: { cuando: "la idea no converge tras 3 vueltas de grill", a: "humano" },
      },
    },
    {
      id: "builder",
      clase: "skill",
      nombre: "construir contra el spec",
      banda: "fase",
      fase: "build",
      estado: "spec -> build",
      canal: "beta",
      procedencia: "declarado",
      contract: {
        why: "materializar el spec en código que pasa sus propias pruebas",
        capabilities: [
          {
            id: "CAP-01",
            what: "implementar cada capability del spec",
            success: "los tests de la capability pasan",
          },
        ],
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T3",
        caja: true,
        fase: "build",
        estado: "spec -> build",
        necesita: [{ art: "spec.md", de: "caja:spec-writer", requerido: true }],
        entrega: [{ art: "código + tests", escritor_unico: true }],
        ruta: [{ a: "reviewer" }],
        gate: {
          tipo: "auto",
          detalle: "la suite de tests del build pasa en verde",
          aceptacion: [
            {
              given: "el código del build",
              when: "se corre go test ./...",
              then: "exit 0 sin fallos",
            },
          ],
        },
        handoff: { cuando: "el build no converge en el cap de reparación", a: "humano" },
      },
    },
    {
      id: "reviewer",
      clase: "skill",
      nombre: "revisar el build",
      banda: "fase",
      fase: "review",
      estado: "build -> review",
      canal: "beta",
      procedencia: "declarado",
      contract: {
        why: "asegurar que el build cumple el spec y el estándar antes de release",
        clase: "skill",
        arquetipo: "excepcion",
        perfil_harness: "T2",
        caja: true,
        fase: "review",
        estado: "build -> review",
        necesita: [{ art: "código + tests", de: "caja:builder", requerido: true }],
        entrega: [{ art: "veredicto de review", escritor_unico: true }],
        ruta: [
          { a: "releaser", si: "review verde" },
          { a: "builder", si: "hay hallazgos que corregir" },
        ],
        gate: { tipo: "parcial", detalle: "checks automáticos + juicio humano sobre hallazgos" },
      },
    },
    {
      id: "releaser",
      clase: "skill",
      nombre: "promover a released",
      banda: "fase",
      fase: "release",
      estado: "review -> released",
      canal: "beta",
      procedencia: "declarado",
      contract: {
        why: "publicar el arnés como snapshot inmutable versionado",
        clase: "skill",
        arquetipo: "pipeline",
        perfil_harness: "T1",
        caja: true,
        fase: "release",
        estado: "review -> released",
        necesita: [{ art: "veredicto de review", de: "caja:reviewer", requerido: true }],
        entrega: [{ art: "release@version", escritor_unico: true }],
        gate: {
          tipo: "none",
          detalle: "gate de fidelidad (§8.4) aún no operacionalizado — diferido honesto",
        },
      },
    },
    {
      id: "std-spec",
      clase: "rule",
      nombre: "estándar de spec",
      banda: "base",
    },
  ],
  edges: [
    { de: "spec-writer", a: "std-spec", tipo: "lee" },
    { de: "spec-writer", a: "builder", tipo: "invoca" },
    { de: "builder", a: "reviewer", tipo: "invoca" },
    { de: "reviewer", a: "releaser", tipo: "invoca" },
  ],
} satisfies Graph
