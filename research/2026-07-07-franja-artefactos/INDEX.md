# Paquete: franja Artefactos (proyección del hand-off + plantillas deterministas)

**Origen:** idea del operador (2026-07-07): 4ª franja «Artefactos» en el Mapa — documentos
intermedios (spec.md, design.md…) como input/output encadenado entre cajas de proceso;
al seleccionar una caja se pintan sus artefactos de entrada y salida; cada artefacto con
plantilla que la caja llena de forma casi determinista (scripts, hooks), con calidad y
ahorro de tokens.

**Evaluación ejecutada:** workflow de 13 subagentes (5 lectores as-code · 3 investigadores
web · 1 auditor · panel de 4 lentes con abogado del diablo), 2026-07-07. Síntesis completa
en [`viabilidad.md`](./viabilidad.md) · decisiones propuestas en
[`decisiones.md`](./decisiones.md).

## Retomar aquí

- **Último hecho (2026-07-08):** operador FIRMÓ D1 + D3–D8 («firmemos todo»). Mockup del
  gate D2 construido y verificado: [`mockup-artefactos.html`](./mockup-artefactos.html) —
  copia fiel de `mockups/arnesia-mapa-mvp.html` + 3 vistas conmutables (Actual · A
  chips-en-spine · B franja-abajo) + toggle todos/auto + selección de caja. Datos: dogfood
  REAL (contratos de los SKILL.md) + showcase Luana (fan-out spec.md→2 cajas, apilado 2
  chips/gutter, dead-end «notas de build», entrega final). Click-through ✅: consola
  limpia, 9 chips Luana, invoca 11→8 (3 hand-offs sustituidos por escribe+lee), reposo en
  auto = mapa idéntico al actual, vista Actual intacta. Screenshots en `shots/`. Artifact:
  https://claude.ai/code/artifact/5e506b6a-7ed3-4e87-9005-61c71b84cd6c
- **Próximo paso:** 🧑‍⚖️ gate D2 — el operador recorre el mockup y decide la geometría
  (A · spine recomendada vs B · franja vs híbrido). Tras la firma: `spec.md`+`design.md`
  (RF trazados a mockup:línea) y recién ahí código (§10). En paralelo ya está habilitada
  la fase 1 del plan (checks de composición, D8 firmada — no depende de la geometría).
- **Firmas:** D1 · D3–D8 ✅ (2026-07-08) · D2 geometría 🚧 abierta sobre el mockup.
  Código intacto: nada se toca hasta specs firmados (METODOLOGIA §10).
- **Hallazgos de auditoría colaterales** (independientes de la firma, ver §8 de
  viabilidad.md): cifra stale en CLAUDE.md («24 pass/211 deferred» → real 27/208) ·
  error de `artifacts.Status` descartado en silencio (`box_conductor.go:90`) · colisión
  de elemento «hooks» knowledge↔conventions en el parser del ruleset · drift
  fixture↔loader (canal/procedencia) en el mismo dogfood.
- **DevStudio:** NO hay que actualizar el contrato hoy. Todo el diseño es aditivo
  (compromiso HS-12 «solo aditivos» lo ampara). Ficha gemela de cortesía (patrón
  DH-18/PB-25) solo cuando se firme el campo `entrega[].plantilla`/`path`.
