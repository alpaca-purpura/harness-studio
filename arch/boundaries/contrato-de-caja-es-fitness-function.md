---
regla: contrato-de-caja-es-fitness-function
version: 1.4
updated: 2026-07-08
status: enforced
ledger: HS-08
sources:
  - url: https://json-schema.org/draft/2020-12
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://opensource.googleblog.com/
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestBoxContractValidatesAgainstSchema
  - fitness/arch_test.go:TestDogfoodComposicionFabricaConforma
  - contracts/schema/box.contract.schema.json
severity: high
---

# El dominio ES el spec vivo: validar el `contract:` de caja contra su schema

## L1 · Principio (estándar de industria)

**Schema-first / contract-first: el modelo de dominio es el single source of truth.** Cuando el
dominio ya tiene forma de schema, ese schema (JSON Schema 2020-12) se vuelve la única fuente:
genera tipos (Go + TS) y **valida instancias** — la doc no puede driftar del código porque ambos
salen del mismo archivo. Validar cada instancia contra el schema **es** una fitness function del
dominio. *(estándar: JSON Schema 2020-12; oficial: google/jsonschema-go — validate + infer,
production-grade, ene-2026)*

## L2 · Realización (este árbol Go+React)

El dominio de ArnesIA ya es schema-shaped: el contrato L0 `meta.clase` (I-75) y el bloque
`contract:` por caja (METODOLOGIA §3). El frontmatter YAML de cada SKILL.md-caja es una
**instancia** de un schema.

- **Single source:** [`../contracts/schema/box.contract.schema.json`](../contracts/schema/box.contract.schema.json)
  (el `contract:` de caja: `caja/fase/estado/necesita/entrega/ruta/gate`) y
  [`../contracts/schema/graph.l0.schema.json`](../contracts/schema/graph.l0.schema.json)
  (`meta.clase` + identidad del nodo agnóstico). ⇐ L1: schema-first.
- **Codegen:** `quicktype` genera los tipos Go **y** TS desde estos schemas → el mapa (React) y el
  indexer (Go) comparten la misma forma sin escribirla dos veces. Los tipos van a
  `contracts/gen/{go,ts}/`. ⇐ L1: single source.
- **Validación = fitness function:** al importar/indexar un arnés, `arnesia` valida cada
  `contract:` contra el schema (`google/jsonschema-go`). Esto vuelve **ejecutable** lo que
  METODOLOGIA §3/§6 pide: el eval-gate por-caja (A4), la detección de precondición (saltos de
  fase / guía sin bloqueo), y la validación de composición (huérfanos = input sin productor
  upstream; dead-ends = output que nadie consume). ⇐ L1: validar instancias.
- **Puente con `knowledge/`:** este check es la versión-arquitectura del estándar de `knowledge/`
  (138 checks). El motor **`arnesia conformance`** (construido en HS-08) parsea `knowledge/` +
  `arch/` = **247 checks a datos** (`--todo`, 2026-07-09) y corre schema-validación (aquí) + go-arch-lint + los checks de
  metodología, mismo reporte severidad+señal.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| contract-valida-schema | cada `contract:` de caja valida contra `box.contract.schema.json` | error | «caja con contrato inválido/incompleto» | arch_test.go:TestBoxContractValidatesAgainstSchema |
| tipos-generados | los tipos Go/TS del contrato salen de `quicktype`, no a mano | warn | «tipo de contrato a mano — riesgo de drift» | schema/codegen |
| sin-huerfanos | ningún `necesita` referencia un artefacto que ninguna caja `entrega` upstream | warn | capa Proceso «input huérfano (sin productor)» | domain.VerificarSinHuerfanos (ruta `--arnes`); `--todo`: arch_test.go:TestDogfoodComposicionFabricaConforma |
| gate-honesto | `gate.tipo: none` cuando no hay eval real (jamás fabricar un eval) | error | capa Proceso «SIN GATE» (principio 10 hecho dato) | arch_test.go |
| dead-end | toda `entrega` tiene consumidor, salvo caja terminal (terminalidad DERIVADA del spine, HS-12) | warn | chip «sin consumidor» en el gutter (franja-artefactos D2) | domain.VerificarDeadEnds (ruta `--arnes`); `--todo`: arch_test.go:TestDogfoodComposicionFabricaConforma |
| ruta-a-existe | cada `ruta[].a` apunta a caja existente o al literal `humano` | error | «ruta colgante» en la caja | domain.VerificarRutaExiste (ruta `--arnes`); `--todo`: arch_test.go:TestDogfoodComposicionFabricaConforma |
| refina-coherente | toda cadena `refina` es lineal: el refinador necesita el art que refina, sin ciclos ni ramas (D9) | error | chip `↻ vN` incoherente en el gutter del refinador | domain.VerificarRefinaCoherente (ruta `--arnes`); `--todo`: arch_test.go:TestDogfoodComposicionFabricaConforma |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = schema-first (JSON Schema 2020-12 +
  google/jsonschema-go). L2: `contract:` de caja y `meta.clase` L0 como single source → quicktype
  (Go+TS) + validación = fitness function del eval-gate A4 y la composición (huérfanos/dead-ends/
  gate honesto). Puente con el estándar de knowledge. 4 checks.
- 2026-07-07 · v1.2 · Sync HS-10: la referencia «linter de 122 checks» apuntaba a un conteo y a un
  runner futuros — el motor real es `arnesia conformance` (construido en HS-08; `knowledge/` = 138
  checks, `knowledge/`+`arch/` = 235 checks a datos). Sin cambios de checks ni de status.
- 2026-07-08 · v1.3 · Franja-artefactos Fase 1 (D8/D9, RF-100..104): `sin-huerfanos` pasa de
  promesa a enforcer vivo (`domain.VerificarSinHuerfanos`); +3 filas nuevas `dead-end` ·
  `ruta-a-existe` · `refina-coherente` (4→7 checks; árbol 97→100, ruleset 235→238). Los 5
  enforcers (incluye `art-identidad-coherente`, built-in del motor como los checks de spine)
  corren en la ruta `--arnes`; en `--todo` las filas difieren honesto (mecanismo no ejecutable
  desde el ruleset). `escritor-unico` ajustado: `refina` es la única puerta legal a la
  multi-escritura (cadena lineal, jamás paralela — D9b).
- 2026-07-08 · v1.4 · Auditoría colateral (HS-16): `sin-huerfanos`/`dead-end`/`ruta-a-existe`/
  `refina-coherente` estaban mal-clasificadas en `--todo` (el parser del ruleset solo reconoce
  el patrón `arch_test.go:TestX` en la columna `enforcer`; el texto `domain.VerificarX (ruta
  --arnes)` caía a `nl-judge`, aunque el enforcer real YA corre y pasa vía `--arnes`). Fix:
  `TestDogfoodComposicionFabricaConforma` (`arch/fitness/arch_test.go`) reusa el mismo
  `ConformanceService` contra el dogfood real de la fábrica y exige `pass` en los 4 — ahora
  visibles como `arch-test` real en `--todo`, sin escribir lógica nueva (la validación ya
  existía). `gate-honesto` queda tal cual (no existe `domain.VerificarGateHonesto` — necesita
  una decisión de diseño previa, deuda declarada, no tocado en esta pasada).
