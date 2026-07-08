---
name: spec-writer
nombre: escribir el spec
description: Convierte una idea conversada en un spec ejecutable. Usar cuando el usuario quiere arrancar una unidad de trabajo desde una idea y aún no hay spec.
version: 1.0
clase: skill
contract:
  why: "convertir una idea conversada en un spec ejecutable que blinde la deriva"
  capabilities:
    - id: CAP-01
      what: "destilar la idea en capacidades con criterio de éxito"
      success: "cada capability tiene un success verificable"
    - id: CAP-02
      what: "emitir spec.md as-code"
      success: "spec.md valida contra el schema de spec"
  constraints:
    - "no inventa requisitos que el usuario no confirmó"
  non_goals:
    - "no escribe código"
    - "no diseña la arquitectura"
  clase: skill
  arquetipo: excepcion
  perfil_harness: T2
  caja: true
  fase: spec
  estado: "idea -> spec"
  necesita:
    - art: "idea del usuario (conversación grill)"
      de: usuario
      requerido: true
    - art: "estándar de spec"
      de: "base:std-spec"
      requerido: true
  entrega:
    - art: spec.md
      path: spec.md
      escritor_unico: true
  ruta:
    - a: builder
      si: "gate del spec verde"
    - a: humano
      si: "la idea no converge en 3 vueltas"
  gate:
    tipo: manual
    detalle: "revisión humana del spec contra la intención declarada"
    aceptacion:
      - given: "un spec.md emitido"
        when: "el humano lo revisa"
        then: "cada capability tiene criterio de éxito y no hay non_goal violado"
    evidencia: "registro de aprobación del spec (telemetría de nacimiento)"
  handoff:
    cuando: "la idea no converge tras 3 vueltas de grill"
    a: humano
---

# spec-writer — la caja de la fase `spec` del arnés dev-full-cycle

Primera caja del arnés dogfood. Toma la idea del usuario (conversación grill) + el estándar
de spec de la banda Base, y emite `spec.md` as-code. Document-as-cache: el estado del trabajo
vive en `spec.md` (frontmatter `status:`), no en la conversación (perfil T2, arquetipo excepción).

El conductor no infiere del texto: lee el `result` + el `status` del artefacto. Si la idea no
converge en 3 vueltas de grill, hace `handoff` a humano (frontera P6/Guardia).
