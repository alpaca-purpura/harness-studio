# Plan Hito 2 — Mapa doctrina-completa + edición gobernada con Claude Code

> Ficha **HS-09** (fase 5 · Implementación) · 2026-07-06
> Antecede: [`../2026-07-06-plan-goal-hito1-mapa-y-deuda.md`](../../research/2026-07-06-plan-goal-hito1-mapa-y-deuda.md) (Hito 1 landeado) · specs Hito 1: [`../2026-07-06-mapa-mvp/`](../2026-07-06-mapa-mvp/)
> Deuda backend de base: [`../2026-07-06-deuda-backend-arch.md`](../../research/2026-07-06-deuda-backend-arch.md)

## Norte del hito

Dejar **el Mapa correctamente funcionando** y **poder modificar las cosas con Claude Code**, con la **metodología visible** tanto en el Mapa como en el detalle de cada nodo, y con **espacios de verificación humana** en cada corte. Todo bajo docs-as-code + arquitectura-as-code + convenciones enforced.

Dos decisiones firmadas por el operador (2026-07-06) que gobiernan el alcance:
- **Dogfood** = *showcase aparte* + `dev-full-cycle` real → [`dogfood-showcase-spec.md`](./dogfood-showcase-spec.md) + [`showcase.graph.json`](./showcase.graph.json).
- **Editar con CC** = *Fase E completa gobernada* → [`edit-con-cc-fase-e.md`](./edit-con-cc-fase-e.md).

## Estado de partida (auditado por 5 subagentes, 2026-07-06)

- **Mapa:** Hito 1 ✓ vivo end-to-end + gran parte de Hito 2 (inspector + picker) ya construida. Índice **en-memoria** (SQLite = Hito 3). Loader = `go:embed`. Sirve `demo` (3 nodos) + `dev-full-cycle` (5 nodos, real).
- **Dogfood pobre:** solo 2/10 clases, 2/7 bandas, 1 canal. → lo resuelve el showcase.
- **Editar con CC:** todas las piezas del patrón conductor **construidas y test-cubiertas pero desconectadas**; el daemon es read-only para el grafo. Único camino vivo = Dock interactivo sin gobierno.
- **Gates:** 8 boundaries enforced; depcruise/vitest/colores ya arreglados (Fase D). Deuda abierta = **Fase E backend** + **go-arch-lint binario** + **deferred→CI**.
- **Doctrina invisible:** tarjeta omite `arquetipo`/`perfil_harness`/honestidad-de-`gate`; inspector omite `constraints`/`non_goals`/Gherkin/`evidencia`. → lo resuelve [`ui-doctrina-visible.md`](./ui-doctrina-visible.md).

## Dos tracks

- **Track A — Mapa doctrina-completa (FE + fixture).** Showcase + marcas de tarjeta (arquetipo/perfil/gate honesto) + inspector completo. Sin backend nuevo (usa `getGraph`/`getNode` ya vivos).
- **Track B — Edición gobernada con CC (backend, Fase E).** ArtifactReader + BoxConductor + KitProvisioner + `control_request` role/ttl + `/boxes/{id}/run` + gate de conformance + GUI diff-approval.

Los tracks corren **secuencial en main, orden A → B** (Gate 0 D4 — sin worktree, trunk-based puro). Track A entrega «metodología visible» sin backend nuevo; Track B se juntan en Fase 4.

## Regla de orden (lección del retro HS-09)

El retro encontró una **inversión de orden** (se codeó antes de firmar los gates; un `enforced` fabricado apuntaba a un enforcer borrado). **Este plan front-loadea el as-code:** ninguna fase de código empieza sin su decisión/contrato firmado. Rítmo probado: **decisión → as-code → build → verificar humano**.

## Fases + gates de verificación humana

### Fase 0 — Encuadre y decisiones as-code · ✅ **Gate 0 FIRMADO** → [`gate-0-decisiones.md`](./gate-0-decisiones.md)
- **D1** fila-meta completa en cajas (arquetipo/perfil/gate/procedencia; **cero tokens nuevos**).
- **D2** run T3 = endpoint dedicado `/boxes/{id}/run`. · **D3** `control_request` = Dock + diff inline. · **D4** main secuencial A→B (sin worktree).
- Contratos as-code se cementan en su fase (Fase 2 = stories+mockup; Fase 3 = OpenAPI 3 rutas + changelogs).

### Fase 1 — Showcase dogfood (data, conformance-valid) · ✅ **EJECUTADA** → 🧑‍⚖️ **Gate 1 (pendiente firma operador)**
- ✅ Promovido a `dogfood/content-studio-full.graph.json` + `//go:embed ContentStudioFullJSON` (`dogfood/embed.go`) + seed en `index/store.go` (refactor `decodeGraph` genérico, seed de ambos arneses).
- ✅ **Conformance VERDE:** `arnesia conformance --arnes dogfood/content-studio-full.graph.json` → **16 checks · 15 pass · 0 fail · 0 error · 1 deferred** (firewall diferido: `fuente_path` no escaneable — honesto). 7 contratos de caja válidos · spine coherente · escritor-único · transiciones legales.
- ✅ Borrador corregido: quitado `_meta` (root `additionalProperties:false`) + quitado `alw` de reglas (dato muerto — el FE deriva `alw` de fixture, ruling Fase D, no del L0).
- ✅ Backend verde: `go build/test -race` (tests de índice actualizados a 3 arneses + `TestSeedServesShowcase` = fitness de las 10 clases).
- ✅ **Ojo-UI mío (evidencia Gate 1):** levanté daemon `:4200` + vite `:5173`, navegué a la app real, cambié el picker a «Editorial · Content Lead» → el showcase **renderiza en su esplendor** (`scratchpad/showcase-live.png`): 7 carriles, Guardia 3 hooks, cajas en spine horizontal, Base completa (Reglas · Knowledge con badge PROPUESTO · Librería · Meta-harness con las 4 clases config · Terceros · Marcas dormidas atenuada). **Sin mojibake** (charset Vite OK). Consola limpia salvo 1 `404` **pre-existente** (`/api/harnesses/backend-nordia/graph` — sesión persistida a un arnés no sembrado, NO del showcase).
- **Hallazgos que alimentan fases siguientes:** (a) doctrina invisible confirmada — arquetipo/perfil/gate NO en tarjeta → **Fase 2**. (b) `origen:del-puesto` + split `siempre/condicional` de reglas aún NO salen del dato (FE los deriva de `proposals.ts`) → cablear en Fase 2 (añadir ids del showcase al fixture) o Hito 3 (provisioning). (c) fixture FE `content-studio-full.ts` + stories → **diferido a Fase 2** (donde las marcas nuevas necesitan story=test; el render live ya prueba el dato).
- ✅ **Gate 1 FIRMADO** (operador, 2026-07-06).
- ✅ **404 stale corregido** (pedido del operador): `web/src/pages/shell/ui/workspace-stage.tsx` — el fetch del grafo ahora espera a que `listHarnesses` cargue y **salta el request** si el arnés de la sesión no está indexado (un 404 solo ensucia consola; el browser lo loguea igual, así que la cura es no pedirlo). Degrada a estado limpio «no indexado» con el picker como escape hatch. Verificado en vivo: consola **0 mensajes**, **cero request** a `backend-nordia/graph` (`scratchpad/backend-nordia-degrade.png`). Gates FE verdes (tsc · biome · depcruise 76 mód/0 viol).

### Fase 2 — Doctrina visible (Track A) · 🧑‍⚖️ **Gate 2**
- ✅ **Tarjeta (hecho):** meta-row en cajas — `arquetipo` (char-badge gris) + `perfil_harness` (pill T1/T2/T3 opacidad creciente) + `gate.tipo` (punto de salud; **`none` = anillo hollow crit = hallazgo**) + `procedencia` (modulación borde-izq: estimado/inferido punteado · no-declarado gris). Derivadas de dato REAL (`node-view.ts`: `ARQUETIPO_MARK`/`GATE_TONE`). Cero tokens nuevos. Archivos: `entities/arnes/{model/node-view.ts,ui/arnes-node.tsx,ui/arnes-node.stories.tsx}` + `app/styles/map.css`.
- ✅ **Story = test (hecho):** `Caja` enriquecida + `CajaGateNone` (honestidad) + `CajaPipelineAuto`. **42 story-tests verdes** (eran 40). tsc strictest · biome · stylelint · depcruise verdes.
- ✅ **Ojo-UI mío (parcial):** verificado live en el showcase — las 7 cajas muestran su clasificación; los 2 `gate:none` (draft/measure) surgen como anillo hollow; riesgo manual=azul vs skill-azul **descartado** (separados arriba/abajo). Menor: punto gate 8px, matiz sutil a zoom default (aria-label/tooltip cargan el significado). Evidencia: `scratchpad/{showcase-marks,caja-brief,caja-draft}.png`.
- ⏳ **Inspector (pendiente):** `constraints` + `non_goals` + Gherkin (`gate.aceptacion`) + `evidencia` + mini-spine + stories.
- ⏳ **Inspector por clase** (pedido operador 2026-07-07) → [`inspector-por-clase.md`](./inspector-por-clase.md): qué ve el usuario por tipo de elemento. **Tier A** (dato existente, pre-Gate 2): encuadre por clase en vez del genérico «Sin contrato», sección Fuente, Activación de reglas (`alw` PROPUESTA), fix `no-reconocido` en `Clase`/`KIND` FE (D-c: hoy tumbaría el canvas). **Tier B** (post-firma): campo `meta` per-class en L0 + 8 reconocedores pendientes de nomenclatura §3.
- ⏳ **Cablear `origen`/split-reglas desde el dato** (hoy FE-fixture-backed en `proposals.ts`).
- **Gate 2 (ojo-UI mío + operador):** al cerrar tarjeta + inspector — doctrina inconfundible en mapa Y nodo; comparar showcase vs dev-full-cycle.

### Fase 3 — Edición gobernada backend (Track B / Fase E) · 🧑‍⚖️ **Gate 3**
- Construir `adapters/artifact/reader.go` (`status:` frontmatter, confinado, nunca chat).
- Seam de permisos en `SpawnOpts`; el adapter `claudecode` reenvía `control_request`/`canUseTool`.
- Cablear `BoxConductor` + `KitProvisioner` en `runServe`; publisher `EventRun` en el broker SSE.
- `POST /boxes/{id}/run` (loop T3) + `POST /sessions/{id}/permission` real (role/ttl, `Grant` con TTL) + ruta de conformance como gate.
- Instalar **go-arch-lint** (destraba 3 boundaries) + flip de tests `t.Skip`→real.
- **Gate 3 (operador):** dry-run backend — editar una caja, aprobar diff, el gate de conformance bloquea/pasa; `go test -race` + go-arch-lint + fitness verdes.

### Fase 4 — Editar desde el Mapa (A ⨯ B) · 🧑‍⚖️ **Gate 4**
- FE: acción «editar con CC» en el nodo → sesión CC confinada a la caja/arnés; tarjeta de aprobación de diff (assistant-ui + CodeMirror 6/merge); veredicto de conformance pintado en el Mapa (diagnóstico unificado, VISION §196).
- **Gate 4 (operador + ojo-UI mío):** click-through completo — editar caja desde el mapa → CC propone → humano aprueba diff → conformance re-valida → mapa actualiza. Screenshots + consola limpia.

### Fase 5 — Cierre docs-as-code · 🧑‍⚖️ **Gate 5**
- Sync: VISION / METODOLOGIA / CLAUDE / LEDGER (ficha HS-09) / `docs/architecture/INDEX.md` + changelogs de boundaries / `docs/architecture/knowledge/CADENCE.md` si se tocó.
- Reconciliar conteos (checks enforced/deferred honestos — no repetir el «235 0-fail» engañoso).
- **Gate 5 (operador):** firma de cierre del hito.

## Checklist de cumplimiento (todo código nuevo lo pasa)

**FE (verde hoy — mantener):** `tsc --noEmit` strictest · `biome ci .` · `depcruise` (FSD + **canvas⊥chrome** + transport⊥domain) · `steiger` · `stylelint` (color = `var(--…)`) · `vitest --project=storybook` (story=test). Boundaries tocados: `fe-taxonomia-componentes`, `fe-tokens-contrato`, `fe-visual-fitness` (enforced), `fe-topologia-fsd`, `fe-transporte-independiente` (depcruise-gated).

**Backend:** `golangci-lint run` + `fmt --diff` vacío · `go test -race` · `arch_test.go` fitness (los 5 enforced siguen verdes; flip de skips) · `box.contract.schema.json` · **`go-arch-lint check`** (instalar binario) · OpenAPI/JSON-schema gen-check.

**Contratos/commits:** OpenAPI + schemas versionados · tokens-sync · Conventional Commits `type(HS-09): …` + trailer `Co-Authored-By`.

**Docs-as-code:** todo cambio de doctrina o arquitectura se refleja en `docs/architecture/knowledge/` o `arch/` ANTES o junto al código (nunca prosa suelta). Boundaries con changelog actualizado.

## Fuera de alcance (Hito 3 — no bloquea)

Indexer JSONL vivo + SQLite modernc · SSE `map` deltas realtime · capas Tokens/Desempeño/Proceso (necesitan telemetría) · tipos TS generados de schema (quicktype). El plan deja estos honestos-deshabilitados, no ocultos.

## Índice del paquete

- [`INDEX.md`](./INDEX.md) — este plan.
- [`gate-0-decisiones.md`](./gate-0-decisiones.md) — **4 decisiones firmadas** (encuadre as-code).
- [`dogfood-showcase-spec.md`](./dogfood-showcase-spec.md) — matriz de casuística del showcase.
- [`showcase.graph.json`](./showcase.graph.json) — el arnés kitchen-sink (borrador, valida en Fase 1).
- [`edit-con-cc-fase-e.md`](./edit-con-cc-fase-e.md) — arquitectura de edición gobernada.
- [`ui-doctrina-visible.md`](./ui-doctrina-visible.md) — revisión ojo-UI + diseño tarjeta/inspector.
