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

- **Último hecho (2026-07-08):** **D1–D8 FIRMADAS** (D2 resuelta sobre el mockup v1:
  geometría **A · chips en spine** — «se ve y entiende mucho mejor»). El operador planteó
  4 escenarios (largo alcance · mejora del mismo artefacto · 8+ inputs · sin plantilla
  tipo factura) → catálogo completo en [`casuistica.md`](./casuistica.md) (C1–C24 + 4
  reglas transversales) + **decisiones D9 (refina) · D10 (admisión/opacos) · D11
  (densidad: panel de entrada ↖ + tope +N)** en decisiones.md, y **mockup v2** que las
  demuestra: ejemplo nuevo «Cobranza» (factura externa sin plantilla, opacos, refina
  ↻v2, gutter con +1 más) + Luana con fan-in de 4 inputs (3 opcionales de largo alcance
  → panel de entrada al seleccionar «promover»). Click-through v2 ✅ (asserts 0 errores,
  consola limpia, vista Actual 0 residuos); screenshots 05–07 en `shots/`. Artifact
  (misma URL): https://claude.ai/code/artifact/5e506b6a-7ed3-4e87-9005-61c71b84cd6c
- **Próximo paso:** 🧑‍⚖️ firma de D9–D11 sobre el mockup v2. Con eso el paquete queda
  listo para `spec.md`+`design.md` (RF trazados a mockup:línea) → código (§10). La fase 1
  del plan (checks de composición, D8) sigue habilitada e independiente.
- **Firmas:** D1–D8 ✅ (2026-07-08) · D9–D11 🚧 propuestas (mockup v2 las demuestra).
  Código intacto: nada se toca hasta specs firmados (METODOLOGIA §10).
- **Hallazgos de auditoría colaterales** (independientes de la firma, ver §8 de
  viabilidad.md): cifra stale en CLAUDE.md («24 pass/211 deferred» → real 27/208) ·
  error de `artifacts.Status` descartado en silencio (`box_conductor.go:90`) · colisión
  de elemento «hooks» knowledge↔conventions en el parser del ruleset · drift
  fixture↔loader (canal/procedencia) en el mismo dogfood.
- **DevStudio:** NO hay que actualizar el contrato hoy. Todo el diseño es aditivo
  (compromiso HS-12 «solo aditivos» lo ampara). Ficha gemela de cortesía (patrón
  DH-18/PB-25) solo cuando se firme el campo `entrega[].plantilla`/`path`.
