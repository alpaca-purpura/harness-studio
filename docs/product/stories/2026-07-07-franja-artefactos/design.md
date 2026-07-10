---
status: firmado (🧑‍⚖️ 2026-07-08, «dale Go»)
spec: spec.md (RF-100..151)
---

# design — diseño técnico por fase

Arquitectura de referencia: hexagonal Go (`internal/domain` puro · `usecase` ·
`adapters`) + FSD-lite FE (`entities → widgets → pages`). Boundaries vigentes: 8
`enforced` — este paquete NO cruza ninguno (verificar con `go-arch-lint` + depcruise en
cada fase). Todo cambio de schema = ADITIVO con `$comment` de versión (patrón P3).

## Fase 1 — checks de composición (solo Go puro + as-code)

**Dónde:** `internal/domain/conformance.go` — gemelos de `VerificarEscritorUnico`
(:444-469, el patrón a copiar: recorre cajas, acumula hallazgos `domain.Hallazgo` con
severidad). Cablear en `internal/usecase/conformance_service.go` (:130-172, RunGraph:
schema → spine → escritor-unico → firewall → **+ composición**).

- `VerificarSinHuerfanos(g)`: index entregas por `(cajaID, art)`; para cada
  `necesita{de: caja:X, art:A}` sin entrada → warn. Cuidado: `refina` cuenta como entrega
  del art (RF-103/104 comparten el índice).
- `VerificarDeadEnds(g)`: terminalidad = la caja cuya transición (`estado`) llega a un
  estado terminal del spine (`arnes.spine.terminales`) — derivada, jamás declarada.
- `VerificarRutaExiste(g)`: destinos válidos = ids de caja + literal `humano`.
- `VerificarRefinaCoherente(g)` + ajuste de `VerificarEscritorUnico`: agrupar entregas
  por art; multi-escritor legal ⇔ todas menos la raíz declaran `refina` y la cadena
  productor→refinador es lineal (detectar ciclo con visited-set).

**As-code:** filas en `docs/architecture/boundaries/contrato-de-caja-es-fitness-function.md` con
`enforced_by:` la función Go (sin-huerfanos ya existe deferred → apuntar al enforcer
real; 3 filas nuevas). Actualizar aritmética en `docs/architecture/INDEX.md` (97→100) y donde el
CLAUDE.md la cite. Los checks corren en la ruta `--arnes` (datos), NO en `--todo`.

**Trampa conocida:** el parser del ruleset colapsa elementos por basename
(`parser.go:105`, colisión docs/architecture/knowledge/hooks ↔ conventions/hooks) — las filas nuevas van
en el boundary, no crear archivo nuevo con nombre colisionante.

## Fase 2 — identidad del art

**Schema:** `docs/architecture/contracts/schema/box.contract.schema.json` — en `$defs` de entrega
(:88-98 hoy `art`+`escritor_unico`, `additionalProperties:false`): sumar `path`
(string, relativa al cwd del arnés), `plantilla` (string, relativa a la skill),
`refina` (string). `$comment` explicando D3/D9.
**Go espejo:** `internal/domain/box.go` `Output{Art, EscritorUnico}` → `+Path,
Plantilla, Refina` + parser del loader (frontmatter YAML de SKILL.md).
**Conductor:** `internal/usecase/box_conductor.go:74` `artifactRef` → preferir
`Entrega[0].Path`; `:90` el `err` de `artifacts.Status` se propaga al log del run y al
`RunResult` (no rompe el loop: status vacío sigue siendo válido, pero VISIBLE).
**Check:** `art-es-path` en domain (warn, solo `arquetipo ∈ {pipeline, excepcion}`).
**Dogfood:** declarar `path: spec.md` en spec-writer (el único art-archivo real hoy).

## Fase 3 — precondiciones + tarea()

`internal/usecase/run_service.go` / `box_conductor.go`: antes de Spawn, resolver
`necesita` requeridos: para `de: caja:X` con productor que declara `path` → stat del
archivo bajo el cwd confinado (reusar el confinamiento de `adapters/artifact`); faltante
→ el run termina en estado «precondición incumplida» con la lista (NO spawnea, NO quema
tokens). `usuario|terceros|base:` no se statean (admisión D10 es de la skill).
`tarea()`: construir el bloque de contexto con `ruta` + primeras N líneas de frontmatter
o `<art>.digest.md` si existe. Tests con arnés fixture en tmp.

## Fase 4 — plantillas dogfood (archivos del arnés, NO del producto)

```
dogfood/dev-full-cycle/
  skills/spec-writer/
    references/plantilla-spec.md      # esqueleto + TOC + frontmatter document-as-cache
    scripts/validate_spec             # sh o go run: estructura + estampa status:done
  hooks/hooks.json                    # PostToolUse(Write|Edit spec.md) + Stop → validador
```
Recordar semántica de hooks: PostToolUse bloquea SOLO con JSON `{"decision":"block",
"reason":…}` (exit 2 ahí NO bloquea); Stop sí bloquea con decision/exit-2 y debe chequear
`stop_hook_active`. El validador es UNO (mismo script en ambos + citado por el Gherkin).
Verificar tras esto RF-150 (¿el loader emite no-reconocido por references/scripts/hooks
del dogfood?) — si ensucia, nomenclatura v1.2.
**Medición p11 (RF-132):** correr spec-writer headless 2×: (a) main actual, (b) con
plantilla+digest; anotar `usage` de stream-json en `medicion-p11.md`.

## Fase 5 — Mapa (port del mockup v2)

```
entities/arnes/model/artefactos.ts    # selectArtefactos(g) puro + tipos ChipArtefacto
entities/arnes/ui/artefacto-chip.tsx  # chip (estados) — story por marca
widgets/map-canvas/ui/handoff-gutter.tsx
widgets/map-canvas/…                  # integración: gutters intercalados, refs, cap+N
widgets/map-canvas/model/use-edge-paths.ts  # SIN cambios de motor (ancla por data-node-id)
```
- La derivación vive en `entities` (regla canvas⊥chrome intacta; el widget solo pinta).
- El grafo servido ya incluye `contract` por nodo (el inspector lo consume) — cero
  fetch nuevo. Si `contract.entrega/necesita` no viajara completo en `/graph`, PARAR y
  resolver en el daemon (aditivo), no duplicar fetch por nodo.
- Supresión del invoca: en el cálculo de edges del canvas, no en el backend — el loader
  NO cambia en esta fase (el edge `escribe` derivado en el loader Go es mejora posterior
  opcional; el FE deriva su propia proyección desde contracts, como el mockup).
- Toggle en MapBar (`off·auto·todos`) + persistencia en hash-state como el resto.
- Estados y prioridades EXACTOS del mockup v2: CAP=3, hand-off>externo, relacionados
  saltan tope, refs solo de la caja seleccionada, chip click→productor.
- Stories = fitness visual (Storybook 10, story=test); asserts del click-through v2
  (documentados en INDEX «Último hecho») se portan como interaction tests.
- Cierre: `PARIDAD.md` (mockup↔componente↔story↔RF) + gate humano lado a lado.

## Orden y gates

1. Fase 1 → commit (independiente, deuda ya prometida).
2. Fase 2 → commit (desbloquea 3 y 4).
3. Fases 3 y 4 en cualquier orden → commits (4 produce la medición p11).
4. Fase 5 → PARIDAD → **gate final humano**.
5. Fase 6 al final (RF-150 se evalúa tras la 4; RF-151 tras la 2).

## Riesgos para el implementador

- `additionalProperties:false` en el schema: olvidar el espejo Go rompe el round-trip —
  correr `arnesia index dogfood/… && conformance --arnes` tras CADA cambio de schema.
- El fixture embebido (`internal/adapters/index/store.go:151`) NO pasa por el loader:
  si se le agregan campos nuevos, mantener coherencia con la salida real del loader
  (drift ya conocido en canal/procedencia — no agrandarlo).
- Los 3 art-etiqueta del dogfood son honestos: NO inventarles path para silenciar
  `art-es-path` — el warn es el diente.
- Hooks del dogfood corren en CADA Write del spawn: validador <500ms o mover a Stop.
- FE: no tocar `use-edge-paths` salvo lo imprescindible — el motor ya ancla por DOM.
