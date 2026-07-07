---
name: builder
nombre: construir contra el spec
description: Materializa el spec en código que pasa sus propias pruebas. Usar cuando hay un spec.md con gate verde y la unidad de trabajo entra a la fase build.
version: 1.0
clase: skill
contract:
  why: "materializar el spec en código que pasa sus propias pruebas"
  capabilities:
    - id: CAP-01
      what: "implementar cada capability del spec"
      success: "los tests de la capability pasan"
  clase: skill
  arquetipo: excepcion
  perfil_harness: T3
  caja: true
  fase: build
  estado: "spec -> build"
  necesita:
    - art: spec.md
      de: "caja:spec-writer"
      requerido: true
  entrega:
    - art: "código + tests"
      escritor_unico: true
  ruta:
    - a: reviewer
  gate:
    tipo: auto
    detalle: "la suite de tests del build pasa en verde"
    aceptacion:
      - given: "el código del build"
        when: "se corre go test ./..."
        then: "exit 0 sin fallos"
  handoff:
    cuando: "el build no converge en el cap de reparación"
    a: humano
---

# builder — la caja de la fase `build` del arnés dev-full-cycle

Segunda caja del arnés dogfood. Toma el `spec.md` aprobado que entrega `spec-writer` y lo
materializa: por cada capability del spec escribe la implementación Y el test que prueba su
`success`. No inventa alcance — el spec es el contrato; lo que no está en una capability no
se construye (la deriva se blinda en la fase anterior, no aquí).

Perfil T3: el conductor Go es dueño del loop (spawn `claude -p --max-turns N`), lee el
`result` + el status del artefacto, nunca infiere del texto. Gate auto: la unidad solo
avanza si `go test ./...` sale en verde — el Gherkin del gate ES el eval. Si el build no
converge dentro del cap de reparación, `handoff` a humano (P6/Guardia); jamás un pass
fabricado.

Entrega «código + tests» como escritor único y rutea a `reviewer`.
