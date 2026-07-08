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
- **FASE 2 COMPLETA (RF-110..112):** schema aditivo `entrega[].path/plantilla/refina`
  (+`$comment` D3/D9; contratos viejos validan sin cambios) · espejo `domain.Output`
  completo · loader parsea vía json-tags sin cambio (verificado: `path` round-tripea) ·
  conductor: `artifactRef` prefiere `entrega[0].path` y el error de `artifacts.Status`
  viaja VISIBLE en `BoxOutcome.Advertencias` → `RunResult.advertencias` + slog.Warn
  (jamás descartado; `TestConductorArtifactIdentity`) · check `art-es-path` (warn) vivo ·
  dogfood `spec-writer` declara `path: spec.md` · fixture `dev-full-cycle.graph.json`
  sincronizado (sin agrandar drift). **HALLAZGO HONESTO:** `--arnes` ahora 21 checks =
  20 pass + 1 warn-fail — `art-es-path` caza los 3 art-etiqueta reales
  (builder/reviewer/releaser); severidad warn NO bloquea el gate; NO se silencia con
  paths inventados (regla del design).
- **FASE 3 COMPLETA (RF-120..121):** `domain.InsumosDe(g, box)` resuelve cada
  `necesita` contra el grafo (path del productor vía entrega/refina) · el conductor
  statea los REQUERIDOS con path antes de spawnear — faltante = `PrecondicionError`
  (run NO arranca, cero tokens; spawn count 0 verificado) → HTTP **409** con lista
  `faltantes` accionable · `requerido:false` jamás bloquea · `tarea()` inyecta bloque
  de insumos: ruta + `Resumen` (digest sidecar `<art>.digest.md` gana; fallback =
  SOLO frontmatter, cap 2KB rune-safe; el cuerpo del doc JAMÁS viaja — test lo
  garantiza) · puerto `ArtifactReader.Resumen` con confinamiento idéntico a Status.
  Tests: `TestConductorEncadenaPorFilesystem` (3 subtests) + `TestInsumosDe` +
  `TestResumen`. Suite verde, golangci 0.
- **FASE 4 COMPLETA (RF-130..133):** dogfood gana el trío completo —
  `skills/spec-writer/references/plantilla-spec.md` (esqueleto con placeholders
  doble-llave) + `scripts/validate_spec` (UN validador: errores verbosos · estampa
  `status: done` SOLO al pasar · genera `<art>.digest.md` determinista) +
  `hooks/hooks.json`+`spec-guard.sh` (PostToolUse bloquea vía JSON decision:block ·
  Stop respeta stop_hook_active; probados A-E manual y E2E) · `entrega[0].plantilla`
  declarado + fixture sync · **RF-131 verificado en headless REAL** (stream-c.jsonl:
  copia con placeholders → block con 7 violaciones → corrección → done+digest) ·
  **medición p11 REAL en [`medicion-p11.md`](./medicion-p11.md)**: el escritor paga
  +32 % USD por validación determinista; el hand-off ahorra **−90 % de contexto por
  insumo** (digest 75 tok vs spec 750 tok); digest ≤200 tok ✓ D7 · RF-133 hecho
  (forjar-caja del kit materializa plantilla+validador, paso 5 nuevo + refina en
  paso 7). **Hallazgo para F6 (RF-150):** el loader NO emite warns por
  references/scripts (solo lee SKILL.md por skill) pero `hooks/hooks.json` queda
  INVISIBLE al grafo — el TODO del loader («reconocedores se añaden con el primer
  arnés real que los use») SE DISPARÓ: el dogfood ahora usa hooks.
- **Próximo paso:** Fase 5 — Mapa (RF-140..145): port del mockup v2
  (`selectArtefactos` + chips + gutters + toggle + panel ↖ + stories) y PARIDAD.md
  para el gate final. Luego 6 (reconocedor de hooks + nomenclatura + ficha DevStudio).
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
