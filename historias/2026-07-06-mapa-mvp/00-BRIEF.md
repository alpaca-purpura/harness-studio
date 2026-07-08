# Fase C — BRIEF (LEER PRIMERO) · Paquete de spec del Mapa MVP en 3 documentos

> Este archivo es el **contrato de la tarea**. Léelo COMPLETO antes de escribir una línea. Es el
> «primer paso» del goal. Produce tres documentos + una fase final de revisión. Gate 2 = aprobación
> explícita del operador.

---

## 0 · Misión

Congelar el **mockup firmado** (`mockups/arnesia-mapa-mvp.html`, Gate 1 ✓ commit `0736d2c`) en un
paquete firmable de **TRES documentos** que, juntos, permitan reconstruir el Mapa **pixel-idéntico sin
ver el mockup**:

| Doc | Ámbito | Pregunta que responde |
|-----|--------|-----------------------|
| **`spec.md`** | El **QUÉ** — comportamiento/funcional | ¿Qué hace el Mapa? Todo lo que está en el mockup. |
| **`design.md`** | El **UI exacto** — visual | ¿Cómo se ve, al pixel? Tokens/medidas/estados para quedar idéntico. |
| **`architecture.md`** | El **CÓMO** — técnico senior | ¿Cómo se implementa bien? Arquitectura as-code + plan hexagonal/FSD. |

`spec.md` es el **hub**: referencia explícitamente a `design.md` (detalle UI) y a `architecture.md`
(plan técnico). Los tres viven en `research/2026-07-06-mapa-mvp/`.

**Al final, fase de REVISIÓN adversarial obligatoria** (§5).

---

## 1 · Regla de oro (transversal, NO negociable)

- **Cero invención.** Cada valor visual del spec/design cita las CUATRO referencias cuando aplique:
  `{token de web/tokens/base.tokens.json}` + `{archivo:línea del mockup}` + `{screenshot en mockups/.shots/}`
  + `{componente Storybook + props}`.
- **Cero magic-value:** solo `var(--…)` / nombres de token. Ningún hex suelto (norma
  [[hs-norma-pegarse-storybook]]).
- **Cero contradicción con doctrina** (VISION/METODOLOGIA/knowledge/arch). Verifica con **subagentes**,
  `archivo:línea`. Donde la doctrina calle, decláralo PROPUESTA (no lo pintes como establecido).
- **No alterar el mockup firmado** ni las decisiones firmadas.
- **Honestidad:** lo staged/deferred se declara explícito, jamás verde falso.

---

## 2 · Estado / contexto (verificar, no confiar)

- **Gate 1 (mockup) FIRMADO** — `mockups/arnesia-mapa-mvp.html` (sustrato **HTML bandas/carriles + SVG
  overlay**, NO React Flow — [[hs-09-mapa-sustrato-html-svg]]).
- **6 decisiones doctrinales firmadas** ([[hs-09-mapa-decisiones-doctrina]], PROPUESTAS a cementar
  as-code en `architecture.md`): (1) Base canónico (VISION A6) · (2) activación honesta por-nodo
  (siempre/condicional/bajo-demanda/leído/dormida) · (3) facet `origen` (estándar del kit vs del-puesto
  llenado en onboarding) · (4) knowledge as-code path-scoped = norma + índice semántico opcional · (5)
  discovery = dato del arnés (no doctrina) · (6) paquete de trabajo = artefacto/spine (cada caja posee 1
  transición; detalle al click).
- Plan madre: `research/2026-07-06-plan-goal-hito1-mapa-y-deuda.md` (Fases A–F).

---

## 3 · Paso previo — re-capturar referencias visuales (chrome-devtools, NO inventar)

Abre el mockup firmado en chrome-devtools con la fixture **Luana** y captura screenshots por ESTADO,
**commiteados en `mockups/.shots/`** con nombres estables (el spec/design los referencian; nada de
scratchpad efímero):

1. `mapa-overview-light` · 2. `mapa-overview-dark` · 3. `mapa-hover-nodo` (foco + dim + lee/escribe) ·
4. `mapa-reglas-colapsadas` · 5. `mapa-reglas-expandidas` (split siempre/condicional) ·
6. `mapa-ayuda` (panel de leyenda) · 7. `mapa-dogfood`.

Consola limpia en cada uno (adjunta evidencia). Mide los valores reales con `evaluate_script` (ancho de
lane, fit-zoom, overflow, conteos, truncamiento) y cítalos como datos, no supuestos.

---

## 4 · Los tres documentos — contenido exhaustivo + buenas prácticas

### A) `spec.md` — el QUÉ (funcional / comportamiento)

**Buenas prácticas de spec:** requisitos atómicos numerados y verificables · sin ambigüedad ·
no-goals explícitos · criterios de aceptación testeable (Gherkin) · cada requisito trazable a su origen.

Debe cubrir, sin huecos (todo lo que HAY en el mockup):
- **Propósito, alcance, no-goals.** Norte del usuario: *entender el mapa de arneses de forma intuitiva*.
- **Requisitos funcionales numerados** de CADA elemento y comportamiento:
  - Regiones (Guardia · Proceso · Base) y su geografía fija.
  - Nodo y variantes (caja · apoyo · compacto · puesto · dim · hover) + handle derivado por clase +
    transición del spine en cajas.
  - Edges: `invoca` (siempre) · `escribe` · `lee`; regla **spine-always + hover-reveal**.
  - Base: subband **Reglas colapsable** (split siempre/condicional) · Knowledge & servicios ·
    librería/meta/terceros con **chip de activación** · marcas-dormidas atenuada.
  - Interacciones: overview-first (fit), zoom +/−, ctrl+wheel, pan, hover-reveal, colapsar reglas,
    toggle Luana↔dogfood, panel de ayuda.
- **Modelo de datos L0** (claves español EXACTAS): `nodos`(id·clase·banda·fase?·caja?·alw?·prop?·trans?·
  origen?) · `arnes`(id·empresa·rol·proceso·reporta_a·marketplace·fases[]·spine?) · `edges`(de·a·tipo).
  Mapear ↔ `internal/domain/graph.go`+`box.go` ↔ `arch/contracts/schema/graph.l0.schema.json`. Marcar
  los campos PROPUESTOS (activacion/origen/trans/alw) que exigen decisión de schema (→ `architecture.md`).
- **Aceptación Gherkin** por comportamiento (layout · edges · hover · colapso · fit · toggle · inspector).
- **Capas:** solo **Estructura** en el MVP (tipo=color+forma+etiqueta+handle+edges+transición).
  Tokens/Desempeño/Proceso = staged (telemetría JSONL). UX «arnés recién nacido» → métricas «—».
- **Split Hito 1** (Mapa read-only navegable) **vs Hito 2** (inspector S3 al click — donde el paquete se
  profundiza).
- **Referencias explícitas:** «detalle visual → `design.md §X`», «implementación → `architecture.md §Y`».
- **Trazabilidad:** cada requisito ↔ `mockup archivo:línea` + `screenshot#`.

### B) `design.md` — el UI exacto (para que quede idéntico)

**Buenas prácticas de design-system doc:** fuente visual única · cada medida anotada · TODOS los estados
e interacciones visuales cubiertos · light + dark · cero magic-value (solo tokens).

Debe cubrir:
- **Tabla completa de tokens** ↔ `base.tokens.json`: los 11 `--c-*` (skill·agent·hook·knowledge·mcp·rule
  + command·plugin·settings·output-style·statusline; hex light/dark), semantic (background/card/secondary/
  foreground/muted-foreground/border/input/primary), ok/warn/crit, radius (md 8/lg 9/full), text
  (xs 11/sm 12.5/base 14), fonts sans/mono.
- **Glyph:** 6 formas (clip-paths exactos) · 11 chars por clase · 17px · color=clase →
  `web/src/shared/canvas/glyph.tsx`. Tabla clase→{shape·char·color-token·label}.
- **Nodo:** anatomía (glyph + nombre 2-line-clamp + handle + trans? + caja-badge? + prop-badge?) +
  border-left 3px; y CADA variante con sus deltas exactos:
  - `.caja` (border-left 5px + tint color-mix 9% + badge) · `.support` (margin-left 14px + secondary
    55%) · `.compact` (Guardia: fila, 1 línea, oculta handle/trans/badges) · `.puesto` (border-left
    **dashed** = facet origen) · `.dim` (.4) · hover (border=--tc + ring 22%).
- **Regiones (3 territorios):** padding 14/16 · radius 14 · bg+border color-mix por región
  (hook/primary/knowledge) · region-hd (h2 uppercase + rule line).
- **Proceso:** lane 232 · lane-hd (fase + count) · **orden caja-first** (spine fila 1) · hairline
  lane-sep + tier apoyo indentado.
- **Base:** header canónico · Reglas colapsable (toggle ▸/▾ · split «siempre en contexto (CLAUDE.md)» vs
  «carga condicional (paths:)» · chips de conteo · colapsa por defecto) · Knowledge & servicios (chip
  «leído por skills») · bands con **chip de activación** (mapa ACT) · marcas-dormidas dimmed.
- **Edges:** invoca (crit·flecha) · escribe (ok·dash 3 3·flecha) · lee (warn·dash 4 4·sin flecha) ·
  bezier cúbico (dx=max(30,|Δx|/2)) · foco: spine no-tocado op .2, edges del nodo op .95/w2.2.
- **Facet origen y prop-badge** (índice semántico `propuesto`).
- **Light + dark** para cada estado. **Screenshots por estado con redlines/anotaciones.**

### C) `architecture.md` — el CÓMO técnico (plan senior)

**Buenas prácticas de tech-design/RFC:** contexto→decisión→consecuencias · principios explícitos ·
alternativas consideradas · honestidad de lo deferred · plan accionable y secuenciado.

Debe cubrir:
1. **Estado actual — revisar A DETALLE nuestra arquitectura as-code:** boundaries FE
   (`arch/boundaries/fe-topologia-fsd`, `fe-taxonomia-componentes` **canvas⊥chrome**,
   `fe-transporte-independiente`, `fe-tokens-contrato`, `fe-visual-fitness`) + backend; modelo C4
   (`arch/model/`); contratos (`arch/contracts/` schemas L0 + box); mapeo `go-arch-lint`; convenciones
   (`arch/conventions/`). Confirmar qué existe / qué falta / qué está `proposed` vs `enforced`.
2. **Arquitectura propuesta del Mapa — hexagonal + FSD-lite:** dominio puro (sin dependencia de
   transporte/UI) · puertos · adapters intercambiables. `widgets/map-canvas` [canvas] ⊥ chrome ·
   `entities/arnes` NO importa `shared/api` (el fetch lo dispara la page) · SSE singleton en
   `app/realtime/` (Hito 3).
3. **Desglose de módulos + flujo de datos:** `api.getGraph` → store `entities/arnes` → selectores
   (selectGuardia/selectLanes/selectBase/selectEdges) → `<MapCanvas>`; algoritmo de layout de carriles
   **determinista** (una columna por `arnes.fases[]` en orden; posiciones fijas).
4. **Backend Hito 1:** loader `dogfood/dev-full-cycle.graph.json` → índice (flag `--seed-graphs` o
   extender `seed()`/`Rebuild` en `internal/adapters/index/store.go`) → `GET /api/harnesses/dev-full-cycle/
   graph` sirve el arnés real (hoy demo de 2 nodos). **Hito 2:** `getNode` (Box+contract), `listHarnesses`.
5. **Cambios de contrato/schema:** los campos PROPUESTOS (origen/activacion/trans/alw) → decisión de
   schema L0 (`graph.l0.schema.json` + `box.contract.schema.json` + `domain`). Qué boundaries mover
   `proposed→enforced`. Cementar aquí las 6 decisiones firmadas (§2).
6. **Sincronización Storybook (SSOT):** el mockup divergió; por cada componente el **delta exacto** para
   el look firmado (`glyph` +5 clases · `arnes-node` variantes+trans+prop · `map-canvas`/`band`/`lane`
   regiones/caja-first/apoyo/Guardia-compacta · `edge-layer` spine+hover) + **componentes NUEVOS**
   (`region`, `base-band`/`rules-subband` colapsable, `activation-chip`, `transition-tag`, `inspector`).
   **Story = test** cada uno.
7. **Estrategia de pruebas:** story=test (fitness visual) · `go build/vet/test ./... -race` · `tsc`
   strictest · Biome · **dependency-cruiser** (arreglar el `.dependency-cruiser.js` roto — deuda
   registrada) · stylelint anti-magic-value.
8. **Escalabilidad (sin sobre-ingeniería, con costuras):** N arneses en índice · N sesiones · indexer
   JSONL real que reemplaza el stub · realtime SSE `map` (Hito 3). Puertos listos para crecer.
9. **Riesgos / trade-offs / alternativas · secuencia de implementación paso a paso a nivel senior**
   (orden, dependencias, PRs por trozo coherente). Principios explícitos: hexagonal · FSD-lite · SOLID ·
   cero magic-value · honestidad de deferred.

---

## 5 · Fase final — REVISIÓN (obligatoria, con subagentes)

Al terminar los 3 documentos, corre una revisión adversarial y **surface el resultado en el chat**:
1. **Consistencia** entre los 3 (spec↔design↔architecture sin contradicción; las referencias cruzadas
   resuelven).
2. **Round-trip de paridad:** recorrer el mockup elemento por elemento ⇒ cada uno está en `spec.md`
   (comportamiento) y `design.md` (visual) con sus referencias; recorrer las referencias ⇒ cada
   token/línea/screenshot/componente existe y es exacto.
3. **Doctrina:** cero contradicción (subagentes, `archivo:línea`).
4. **Completitud:** nada del mockup falta; lo deferred declarado honesto.
Corregir hasta que la revisión salga verde. Publica en el chat el **checklist de paridad visual** (la
tabla `elemento · token · mockup archivo:línea · componente Storybook · screenshot#`) como evidencia.

---

## 6 · Definición de hecho / Gate 2

Los 3 documentos completos y cross-referenced, `spec.md` referenciando `design.md` + `architecture.md`,
revisión final verde con el checklist de paridad surfaced. Entonces **presentar al operador y conseguir
su aprobación explícita (Gate 2)** antes de implementar. Iterar hasta el «apruebo». Trunk-based
(`main`), commit por trozo coherente, firma `Co-Authored-By: Claude Opus 4.8 (1M context)`.

---

## 7 · Anclas (rutas exactas)

- **Mockup firmado:** `mockups/arnesia-mapa-mvp.html` · screenshots → `mockups/.shots/`
- **Tokens SSOT:** `web/tokens/base.tokens.json`
- **Storybook (SSOT FE):** `web/src/shared/canvas/glyph.tsx` · `web/src/entities/arnes/`
  (model/{types,selectors,kind}, ui/arnes-node, testing/dev-full-cycle) ·
  `web/src/widgets/map-canvas/` (ui/{map-canvas,map-bar,band,lane,edge-layer}, model/{layers,use-edge-paths})
- **Seam de montaje:** `web/src/pages/shell/ui/workspace-stage.tsx` (canvas⊥chrome; hoy `<ComingSoon>`)
- **Backend:** `internal/domain/graph.go`+`box.go` · `internal/adapters/index/store.go` ·
  `dogfood/dev-full-cycle.graph.json` · `arch/contracts/schema/{graph.l0,box.contract}.schema.json`
- **Arch as-code:** `arch/boundaries/fe-*.md` · `arch/model/` · `arch/contracts/` · `arch/conventions/` ·
  `.go-arch-lint.yml`
- **Doctrina:** `VISION.md` (A6) · `METODOLOGIA.md` (§3 contrato, §8.3 document-as-cache) · `UX.md` (S2/S3)
  · `knowledge/` (nodes)
- **Memorias:** `hs-09-mapa-decisiones-doctrina` · `hs-09-mapa-sustrato-html-svg` · `hs-09-fase5-mapa` ·
  `hs-norma-pegarse-storybook` · `hs-arquitectura-fe`

---

## 8 · Disciplina de ejecución

Subagentes para lecturas/verificación en paralelo + verificación **adversarial** (por cada afirmación
visual, un subagente intenta refutar que la referencia sea exacta). Nada «hecho» sin abrir el archivo
real. Commit por trozo. Si algo queda diferido, decláralo honesto.
