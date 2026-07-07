---
name: reviewer
nombre: revisar el build
description: Asegura que el build cumple el spec y el estándar antes de release. Usar cuando builder entrega código + tests y la unidad de trabajo entra a la fase review.
version: 1.0
clase: skill
contract:
  why: "asegurar que el build cumple el spec y el estándar antes de release"
  clase: skill
  arquetipo: excepcion
  perfil_harness: T2
  caja: true
  fase: review
  estado: "build -> review"
  necesita:
    - art: "código + tests"
      de: "caja:builder"
      requerido: true
  entrega:
    - art: "veredicto de review"
      escritor_unico: true
  ruta:
    - a: releaser
      si: "review verde"
    - a: builder
      si: "hay hallazgos que corregir"
  gate:
    tipo: parcial
    detalle: "checks automáticos + juicio humano sobre hallazgos"
---

# reviewer — la caja de la fase `review` del arnés dev-full-cycle

Tercera caja del arnés dogfood. Recibe «código + tests» de `builder` y lo confronta contra
DOS varas: el `spec.md` (¿cada capability quedó implementada con su success probado?) y el
estándar del repo (convenciones, honestidad de errores, cero pass fabricado).

Emite un único artefacto: el «veredicto de review» — verde o lista de hallazgos accionables,
cada hallazgo con su evidencia. Gate parcial: los checks automáticos corren primero; el
juicio humano decide sobre lo que las máquinas no ven. Ruta condicional: veredicto verde →
`releaser`; hallazgos → de vuelta a `builder` (rework, la transición review→build del spine).

No corrige código: revisar y construir no comparten escritor (mutation contract).
