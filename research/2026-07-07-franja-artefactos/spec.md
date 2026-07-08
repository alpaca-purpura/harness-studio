---
status: firmado (🧑‍⚖️ 2026-07-08, «dale Go»)
paquete: franja-artefactos
decisiones: D1-D11 firmadas (decisiones.md)
mockup: mockup-artefactos.html (v2, click-through verde 2026-07-08)
---

# spec — Artefactos: identidad + checks + plantillas + proyección en el Mapa

**why:** el hand-off entre cajas ES el contrato (A2); hoy es dato invisible y sin diente —
este paquete le da identidad ejecutable (D3), verificación (D8/D9), plantillas con llenado
casi determinista (D4–D7) y proyección visual en el Mapa (D2/D11), midiendo el ahorro de
tokens (p11).

**Reglas de juego:** todo ADITIVO (jamás romper schema publicado — patrón P3/HS-12) ·
artefacto JAMÁS nodo del L0 ni banda ni clase 11ª (D1) · el Mapa pinta TIPOS, el Flujo
pinta INSTANCIAS (A3, decisión #6 Gate 1) · DevStudio: cero cambio de contrato; ficha
gemela de cortesía al aterrizar RF-110.

## non_goals

- NO tablero «Flujo del trabajo» (instancias vivas — telemetry-gated, vista aparte).
- NO plantillas para inputs externos (factura se ADMITE, D10).
- NO universalizar llenado determinista a cajas `abierto` (§8.1).
- NO editar el mapa desde los chips (read-only, como el Mapa actual).
- NO digest sidecar para binarios opacos (deuda declarada D10).
- NO codegen ni rediseño del conductor más allá de lo listado.

## RFs — Fase 1 · Motor: checks de composición (D8/D9; sin dependencias)

| RF | Qué | success (eval) |
|----|-----|----------------|
| RF-100 | Check `sin-huerfanos` (warn) VIVO en ruta `--arnes`: todo `necesita.de: caja:<id>` debe tener productor que entregue ese art. `usuario|terceros` no son huérfanos (externos visibles) | test unitario + dogfood sigue 15/15; arnés sintético con huérfano lo reporta |
| RF-101 | Check `dead-end` (warn): entrega sin consumidor de caja NO terminal (terminalidad derivada del spine) | «notas de build» sintético reporta; `release@version` del dogfood NO reporta |
| RF-102 | Check `ruta-a-existe` (error): `ruta[].a` apunta a caja existente o `humano` | test con destino inexistente falla |
| RF-103 | Check `art-identidad-coherente` (error): `necesita {art:A, de:caja:X}` exige que X entregue (o refine) A | mismatch sintético falla; dogfood pasa |
| RF-104 | `escritor-unico` ajustado a `refina` (D9): 2 entregas del mismo art SIN refina = error (como hoy); CON refina = cadena lineal válida. Check `refina-coherente` (error): el refinador necesita el art que refina; sin ciclos | cadena cobranza-like pasa; refina sin necesita falla; ciclo falla |

Cementación as-code: filas de check en el boundary
`arch/boundaries/contrato-de-caja-es-fitness-function.md` (sin-huerfanos ya declarado
deferred → vivo; dead-end/ruta-a-existe/refina-coherente = filas nuevas) — la aritmética
de checks del INDEX se actualiza.

## RFs — Fase 2 · Identidad del art (D3/D9, contrato + conductor)

| RF | Qué | success |
|----|-----|---------|
| RF-110 | `box.contract.schema.json` aditivo: `entrega[].path` (string), `entrega[].plantilla` (string, ruta relativa a la skill), `entrega[].refina` (string) — todos opcionales + espejo `domain.Output` + parser | schema valida contratos viejos SIN cambios; nuevos campos parseados |
| RF-111 | Conductor T3: `artifactRef` usa `entrega[0].path` si existe (fallback art como hoy); el error de `artifacts.Status` deja de descartarse en silencio (se loguea y viaja al resultado del run) | test: art no-path ya no lee status fantasma; error visible |
| RF-112 | Check `art-es-path` (warn): cajas `pipeline|excepcion` con `entrega[0]` sin `path` | dogfood reporta los 3 art-etiqueta (honesto) |

## RFs — Fase 3 · Conductor: encadenado (D7)

| RF | Qué | success |
|----|-----|---------|
| RF-120 | Precondición pre-Spawn: para cada `necesita` con `requerido:true` y path resoluble (productor con `path`), verificar existencia; si falta → el run NO arranca y reporta qué falta. `requerido:false` jamás bloquea | test: run sin spec.md en disco → 4xx/resultado «precondición incumplida» con lista |
| RF-121 | `tarea()` inyecta RUTAS + frontmatter/digest del artefacto de entrada (si existe `<art>.digest.md`), jamás el doc completo | inspección del prompt spawneado: sin contenido del doc |

## RFs — Fase 4 · Plantillas dogfood-first (D4/D5/D6; mide p11)

| RF | Qué | success |
|----|-----|---------|
| RF-130 | `spec-writer` del dogfood gana `references/plantilla-spec.md` (esqueleto: frontmatter status/inputs/timestamps + why + CAP-NN + non_goals, TOC) + `scripts/validate_spec` (errores VERBOSOS; estampa `status: done` SOLO al pasar) + `entrega[0].plantilla` declarado | validador rechaza spec incompleto con mensaje accionable; acepta el spec bien llenado |
| RF-131 | Hooks del arnés dogfood: PostToolUse (Write\|Edit sobre spec.md → validador, bloqueo JSON `decision:block`) + Stop (gate final, respeta `stop_hook_active`) | secuencia headless real: escribir spec inválido → block con violaciones; corregir → pasa |
| RF-132 | Medición p11 registrada en el paquete: tokens de una corrida spec-writer ANTES (estándar en prosa) vs DESPUÉS (plantilla+digest+rutas) | tabla en `medicion-p11.md` con números reales |
| RF-133 | (backflow I-59, puede diferirse a PR separado) `kit/skills/forjar-caja` materializa plantilla+validador al forjar cajas pipeline/excepcion | forjar caja nueva produce los archivos |

## RFs — Fase 5 · Mapa (D2/D11 — port del mockup v2, FSD-lite)

Trazabilidad = `mockup-artefactos.html` (v2). PARIDAD.md al cierre (mockup↔componente↔story↔RF).

| RF | mockup:línea | Qué | success |
|----|-------------|-----|---------|
| RF-140 | 420-454 | `selectArtefactos(g)` selector puro en `entities/arnes/model`: deriva chips (handoff/externo/dead/final/opaco/etiqueta/plantilla) + versión refina DERIVADA (ciclos cortados) + consumidores con `req` | unit tests: dogfood 5 chips · fan-out 3 consumidores · refina v2 · dead/final correctos |
| RF-141 | 123-141, 490-508 | `ArtefactoChip` (papel esquina doblada; opaco=relleno; etiqueta=cursiva; tags ↻vN/plantilla/externo/sin-consumidor/salida-del-proceso) + `HandoffGutter` entre carriles, `data-node-id` | stories por marca; overlay de edges los conecta sin tocar el motor |
| RF-142 | 456-462, 689, 706 | Edges derivados: `escribe` caja→chip, `lee` chip→consumidor (opcional = tenue `2 6`); supresión del `invoca` del hand-off SOLO si su chip está visible | story de reposo = mapa actual idéntico; con chips, invoca sustituido |
| RF-143 | ctl.view | Toggle «Artefactos» en MapBar: `off · auto · todos` (auto = solo al seleccionar; reposo idéntico al mapa actual) | off = cero DOM extra; auto reposo = snapshot igual al actual |
| RF-144 | 144-152, 509-516, 594-604, 644-676 | Panel de entrada ↖ (refs de largo alcance de la caja seleccionada, click navega al productor, tag opcional) + tope `CAP=3` por gutter con prioridad hand-off>externo y relacionados-saltan-tope + «+N más»/«− plegar» | asserts del click-through v2 reproducidos como tests |
| RF-145 | click handlers | Click chip → selecciona productor; selección consistente con inspector/drawer (RF-90: misma fuente `necesita/entrega`, cero fetch nuevo — el contract ya viaja en GET /graph) | navegación cruzada chip↔caja↔drawer verde |

## RFs — Fase 6 · Nomenclatura y reconciliación (condicional)

| RF | Qué | success |
|----|-----|---------|
| RF-150 | Verificar reconciliación del loader con `skills/<id>/references/` y `scripts/` (¿emite `no-reconocido` por archivos extra dentro de la skill?). Si ensucia: nomenclatura v1.2 (fila plantilla → metadato del nodo caja, changelog firmado patrón v1.1). Si no ensucia: cerrar RF con evidencia | `arnesia index dogfood/dev-full-cycle` sin warns nuevos tras RF-130 |
| RF-151 | Ficha gemela de cortesía a DevStudio (patrón DH-18/PB-25) anunciando campos opcionales de forma-plugin (RF-110) | ficha en LEDGER + espejo enviado |

## Criterios de aceptación transversales

- `go test ./...` + golangci verdes · round-trip dogfood `index → conformance --arnes`
  sigue TODO PASS (con los checks nuevos incluidos) · FE: vitest + stories + depcruise +
  biome verdes · consola limpia en click-through real.
- Cada fase termina con commit a main y actualización del «Retomar aquí» del INDEX.
- Gate final humano: PARIDAD.md lado a lado contra el mockup v2.
