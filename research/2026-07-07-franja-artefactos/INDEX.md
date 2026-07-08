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

- **Último hecho (2026-07-08):** **D1–D11 FIRMADAS** — el paquete pasó a **BACKLOG
  IMPLEMENTABLE**: [`spec.md`](./spec.md) (RF-100..151, 6 fases, trazado a
  mockup:línea/casuística/decisiones) + [`design.md`](./design.md) (diseño técnico por
  fase: archivos, funciones-gemelo, orden, gates, riesgos) escritos y en estado
  `pendiente-de-firma`. Mockup v2 = referencia visual firmada (geometría A · chips en
  spine; Cobranza demuestra refina ↻v2/opacos/+N; Luana el fan-in con panel ↖; asserts
  del click-through: 0 errores, consola limpia, vista Actual 0 residuos). Artifact:
  https://claude.ai/code/artifact/5e506b6a-7ed3-4e87-9005-61c71b84cd6c · Backlog UX.md
  item 13.
- **Próximo paso:** sesión de implementación NUEVA arranca pegando
  [`PROMPT.md`](./PROMPT.md). Primer gate de esa sesión: 🧑‍⚖️ firma de spec.md+design.md;
  luego fases 1→6 en orden, cada una con suite verde + commit + este «Retomar aquí»
  actualizado; cierre = PARIDAD.md + gate final lado a lado.
- **Firmas:** D1–D11 ✅ (2026-07-08) · spec.md+design.md 🚧 pendientes (gate de la sesión
  de implementación). Código intacto hasta esa firma (METODOLOGIA §10).
- **Hallazgos de auditoría colaterales** (independientes de la firma, ver §8 de
  viabilidad.md): cifra stale en CLAUDE.md («24 pass/211 deferred» → real 27/208) ·
  error de `artifacts.Status` descartado en silencio (`box_conductor.go:90`) · colisión
  de elemento «hooks» knowledge↔conventions en el parser del ruleset · drift
  fixture↔loader (canal/procedencia) en el mismo dogfood.
- **DevStudio:** NO hay que actualizar el contrato hoy. Todo el diseño es aditivo
  (compromiso HS-12 «solo aditivos» lo ampara). Ficha gemela de cortesía (patrón
  DH-18/PB-25) solo cuando se firme el campo `entrega[].plantilla`/`path`.
