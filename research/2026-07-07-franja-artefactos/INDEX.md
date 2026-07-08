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

- **Último hecho (2026-07-08):** **spec.md + design.md FIRMADOS** («dale Go», tras
  verificación código↔spec sin contradicciones bloqueantes — 3 matices menores
  registrados en decisiones.md) · **FASE 1 COMPLETA (RF-100..104):** los 5 checks de
  composición VIVOS en `internal/domain/conformance.go`
  (`VerificarSinHuerfanos`/`DeadEnds`/`RutaExiste`/`ArtIdentidad`/`RefinaCoherente`),
  `VerificarEscritorUnico` ajustado a `refina` (única puerta legal a multi-escritura,
  cadena lineal), `domain.Output.Refina` añadido (schema llega en Fase 2), cableados en
  `RunGraph` (4b). Tests sintéticos cazan huérfano/dead-end/ruta colgante/mismatch
  C21/refina-sin-necesita/ciclo/rama. **Round-trip dogfood 20/20 PASS** (antes 15/15) ·
  `--todo` 238 checks = 27 pass + 211 deferred (filas nuevas difieren honesto) · go test
  11 paquetes ok · golangci 0 issues. As-code: boundary
  `contrato-de-caja-es-fitness-function` v1.3 (4→7 filas) · arch/INDEX 97→100 ·
  CLAUDE.md sincronizado (cifra stale 24/211→27/211-de-238 medida hoy).
- **Próximo paso:** Fase 2 — identidad del art (RF-110..112): schema aditivo
  `entrega[].path/plantilla/refina` + espejo Go completo + parser + conductor
  (`artifactRef` con path; error de `Status` visible) + check `art-es-path` + dogfood
  `path: spec.md`. Luego 3/4 (conductor+plantillas, medición p11) → 5 (Mapa+PARIDAD) → 6.
- **Firmas:** D1–D11 ✅ · spec.md+design.md ✅ (2026-07-08). Gate final pendiente =
  PARIDAD.md lado a lado tras Fase 5.
- **Hallazgos de auditoría colaterales** (independientes de la firma, ver §8 de
  viabilidad.md): cifra stale en CLAUDE.md («24 pass/211 deferred» → real 27/208) ·
  error de `artifacts.Status` descartado en silencio (`box_conductor.go:90`) · colisión
  de elemento «hooks» knowledge↔conventions en el parser del ruleset · drift
  fixture↔loader (canal/procedencia) en el mismo dogfood.
- **DevStudio:** NO hay que actualizar el contrato hoy. Todo el diseño es aditivo
  (compromiso HS-12 «solo aditivos» lo ampara). Ficha gemela de cortesía (patrón
  DH-18/PB-25) solo cuando se firme el campo `entrega[].plantilla`/`path`.
