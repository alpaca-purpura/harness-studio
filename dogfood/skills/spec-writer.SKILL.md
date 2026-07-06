---
name: spec-writer
description: Convierte una idea conversada en un spec ejecutable. Usar cuando el usuario quiere arrancar una unidad de trabajo desde una idea y aún no hay spec.
version: 1.0
clase: skill
contract:
  why: "convertir una idea conversada en un spec ejecutable que blinde la deriva"
  clase: skill
  arquetipo: excepcion
  perfil_harness: T2
  caja: true
  fase: spec
  estado: "idea -> spec"
  gate:
    tipo: manual
    detalle: "revisión humana del spec contra la intención declarada"
---

# spec-writer — la caja de la fase `spec` del arnés dev-full-cycle

Primera caja del arnés dogfood. Toma la idea del usuario (conversación grill) + el estándar
de spec de la banda Base, y emite `spec.md` as-code. Document-as-cache: el estado del trabajo
vive en `spec.md` (frontmatter `status:`), no en la conversación (perfil T2, arquetipo excepción).

El conductor no infiere del texto: lee el `result` + el `status` del artefacto. Si la idea no
converge en 3 vueltas de grill, hace `handoff` a humano (frontera P6/Guardia).
