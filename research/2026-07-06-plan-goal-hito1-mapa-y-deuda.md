# Goal — HS-09 Hito 1 (MVP del Mapa) + Deuda paralela · Storybook = SSOT de FE

> Prompt de ejecución para una sesión fresca de Claude Code (usa subagentes para leer/verificar en
> paralelo; nada se da por hecho sin correrlo). Orden mandado por el operador: **revisar el shell exacto →
> actualizar el mockup DESDE Storybook (con aprobación del operador) → spec detallado (con aprobación del
> operador) → arquitectura as-code → resolver deuda paralela + arrancar Hito 1**. Dos gates de firma
> (mockup, spec) y una fase de arquitectura son BLOQUEANTES antes de programar.

---

## ⚑ Bitácora de avance · estado · aprendizajes (act. 2026-07-06)

### Avance por fase
- **Fase A ✓ COMPLETA** (commit `a8721e9`). Entregable `research/2026-07-06-shell-map.md` — shell exacto
  `archivo:línea` (3 subagentes verificados). Seam del Mapa = `workspace-stage.tsx:40-46`, branch
  `s.view==="Mapa"` → `<MapCanvas arnesId={s.arnes}/>`.
- **Fase B ✓ COMPLETA — 🛑 Gate 1 (mockup) FIRMADO por el operador (2026-07-06, commit `0736d2c`).**
  Iteración crítica final en chrome-devtools con fixture **Luana** (goal: ser el crítico de la UX). Aprobado
  tras bajar 6 decisiones doctrinales (verificadas por 3 subagentes, cero contradicción) + los 3 refinamientos:
  - **Base canónica** (elimina el falso «Soporte · siempre en contexto», VISION A6) · **activación honesta
    por-nodo** (siempre/condicional/bajo-demanda/leído/dormida) · **reglas colapsables** (siempre-on vs `paths:`).
  - **Facet `origen`** (estándar del kit vs del-puesto llenado en onboarding = borde punteado) · **índice
    semántico opcional** (`propuesto`; knowledge as-code path-scoped = norma, semántico opcional, NO dep dura).
  - **Discovery = dato del arnés** (no doctrina; fixture) · **paquete de trabajo = artefacto/spine**: cada caja
    muestra la transición que posee (`idea→…→operando`); detalle al click = Inspector (Hito 2); paquete vivo
    moviéndose = tablero «Flujo del trabajo» (lente futura, en debate).
  - **Refinamientos**: Guardia compacta (hooks=chips) · apoyo sin la palabra (hairline+indent) · **gap 6→10
    colores RESUELTO** (5 tokens `kind` nuevos en `web/tokens/base.tokens.json`, propuestos HS-09).
  - Base de interacción firmada: overview-first (fit real) · spine siempre + lee/escribe on-hover · nombres
    2 líneas · `<!doctype>` (consola limpia). Memoria: `hs-09-mapa-decisiones-doctrina`.
  - **Las 6 decisiones son PROPUESTAS a cementar as-code (VISION/METODOLOGIA/knowledge/arch) en Fase D.**
- **Fase B (registro previo) — construido y verde** (typecheck+biome+storybook build; render light/dark):
  - Componentes SSOT (commit `a8721e9`): `shared/canvas/glyph` · `entities/arnes` (types/selectors/kind/
    ArnesNode/fixtures) · `widgets/map-canvas` (MapBar/Band/Lane/EdgeLayer/MapCanvas + use-edge-paths) + stories.
  - Fixture completo **Luana** (commit `5207e94`) — empresa canónica de los mockups; 7 fases · 24 nodos ·
    10 clases · edges 3 tipos.
  - Mockup **v2** (commit `4b438db`) tras feedback del operador (7 cambios, abajo). Data-driven, toggle
    Luana↔dogfood. Artifact `08d31cde-3f91-4808-a14f-fe8d31ae7572`.
- **Deuda · diseño de arquitectura ✓** (commit `3a8a278`) — `research/2026-07-06-deuda-backend-arch.md`
  (track independiente; análisis read-only, sin código).
- **Fase C ✓ COMPLETA — 🛑 Gate 2 (spec) FIRMADO por el operador (2026-07-06).** Paquete en
  `research/2026-07-06-mapa-mvp/`: `spec.md` (QUÉ — RF numerados + Gherkin, cada uno trazado a
  `mockup:línea`+shot) · `design.md` (UI al pixel — tokens/medidas) · `architecture.md` (CÓMO —
  hexagonal+FSD-lite, loader backend, secuencia de PRs, plan de cementado de las 6 decisiones) ·
  `PARIDAD.md` (round-trip visual + deferred honesto). Fiel al mockup firmado; corrige un dato stale
  (el seed es **3 nodos/1 edge keyed «demo»**, no 2 → `dev-full-cycle` da **404** hoy).
- **⚠️ Auditoría 5-subagentes (2026-07-06, build/test real) destapó INVERSIÓN DE ORDEN.** Se escribió
  **código adelantándose a los gates**: Fase F (Hito 1) YA implementada en el working-tree (loader del
  dogfood + `<MapCanvas>` montado, adiós `ComingSoon` + adelanto de Hito 2 inspector/picker) con **build
  verde** (go build/vet/test `-race` · tsc · biome · depcruise · steiger · stylelint), **sin** cerrar
  Fase D (arquitectura as-code) ni Fase E (deuda), y con Fase C **sin firmar**. El **código respeta la
  doctrina** (backend agnóstico `fase`/`estado`=dato; canvas⊥chrome; tokens `var(--…)`; contrato fusionado;
  las PROPUESTAs `origen`/`alw` aisladas en `proposals.ts` y etiquetadas). PERO: (i) la doctrina as-code
  (VISION/METODOLOGIA/knowledge/arch) **NO refleja** las 6 decisiones que el código ya materializó; (ii)
  🚩 `arch/boundaries/fe-visual-fitness.md` quedó flipado a `status: enforced` apuntando a un enforcer
  **BORRADO** (`web/vitest.workspace.ts`) = **pass fabricado** (viola «no fabricar pass»).
- **Decisión del operador: RETRO-AJUSTAR EN ORDEN.** (1) ✓ **Gate 2 FIRMADO**. (2) **Fase D de verdad**:
  cementar las 6 decisiones as-code + arreglar el `enforced` falso + cuerpos de boundaries + re-sincronizar
  conteos (knowledge/arch) + C4/contratos + graduar las PROPUESTAs (`origen`/`alw`) de fixture a L0
  (decisión doctrinal pendiente en D — tensión `origen` vs `maquinaria-no-contamina-arnes`). (3) **RECIÉN
  AHÍ** commitear el código como Fase F. Deuda E queda **diferida honesta**.
- **Fases D → F — EJECUTADAS en orden** (commits `9cd8e77` D · `31a7532` + `c013dc3` F). Fase D cerrada
  ANTES del código: pass fabricado corregido (`fe-visual-fitness`) · 3 boundaries FE enforced con
  enforcers verificados corriendo (→ 8 enforced, 97 checks intacto) · 6 decisiones as-code (`origen` en
  schema, `alw` derivado en `rules.md`) · C4 + propagación HTML+SVG a VISION/CLAUDE. Fase F = **Hito 1 del
  Mapa VIVO** sobre el dogfood real (loader + `<MapCanvas>` montado), todo verde (go/tsc/biome/depcruise/
  steiger/stylelint/40 story-tests). Fase E (deuda backend) sigue diferida honesta.

### Decisiones firmadas / tomadas
1. **Sustrato del Mapa = HTML bandas/carriles + SVG overlay** (NO React Flow). Operador eligió; RF se reserva
   al Organigrama. Revierte "React Flow para el Mapa" de CLAUDE.md/HS-04. → memoria `hs-09-mapa-sustrato-html-svg`.
2. **7 cambios del operador sobre el mockup (v2):** (1) vista limpia + ayuda en botón «?» flotante · (2) edges
   **invoca=rojo · escribe=verde · lee=ámbar** · (3) tag de comando/invocación por nodo · (4) zoom +/−/ajustar
   + pan + más aire · (5) **distintivo de caja de proceso** (`contract.caja`) · (6) separadores de región
   Guardia/Proceso/Soporte · (7) región **Soporte** con sub-bandas.
3. **Iteración de diseño en el mockup** (superficie rápida, menos tokens); los componentes de Storybook (SSOT
   real) se sincronizan al **congelar el look**. Divergencia temporal declarada.

### Aprendizajes (doctrina — investigados, no inventados)
- **Caja de proceso = `contract.caja == true`** (skill-frente de una fase; posee 1 transición + gate). vs skill
  de **apoyo** (`caja:false` + `rol:`). En capa Estructura la doctrina NO marca visualmente la caja (solo la
  capa Proceso lo hacía) → el distintivo es una decisión NUEVA del operador. (`box.go:186-208`, `skills.md:93-100`)
- **"Soporte" NO es término doctrinal — el canónico es "Base"** (VISION A6). ⚠️ pendiente decisión del operador:
  ¿oficializar Base→Soporte en VISION/METODOLOGIA (Fase D) o mantener "Base"?
- **"Knowledge" NO es una clase** — se pliega en `rule` (firewall CC-native §8.6). Los "tipos de soporte" reales
  = **bandas** `base · libreria-expertos · meta-harness · terceros · marcas-dormidas` (no un subtipo "Knowledge").
- **No existe campo "comando/invocación"** en el modelo — el `id` es la clave. El handle se **deriva por clase**
  (command/skill→`/id` · subagent→name · hook→evento · rule→always-on).

### Aprendizajes (técnicos / deuda destapada → Fase D)
- `.dependency-cruiser.js` **roto**: `module.exports` en repo `type:module` ESM + globs `canvas⊥chrome` con
  rutas stale (`widgets/command-rail|dock`, `app/shell` inexistentes; el chrome real = `session-rail|chat-dock|
  topbar|view-strip` + `pages/shell`). Enforcer inerte.
- `vitest.workspace.ts` obsoleto (vitest 4 dropeó `defineWorkspace` → `test.projects`) → los **story-tests no
  ejecutan** aún; render verificado por chrome-devtools en su lugar. Playwright chromium instalado (fallback).
- **Gap 6→10 colores:** tokens `kind` traen 6; `Clase` tiene 10 → command/plugin/settings/output-style/
  statusline caen a `--muted-foreground`. Extender tokens en Fase D.
- `biome.json`: `noThenProperty` off (Gherkin `then` = vocabulario de contrato de caja).
- **Meta-proceso:** el goal por Stop-hook entra en loop cuando choca con un gate de aprobación HUMANO — el hook
  no distingue "bloqueado en humano" de "parado antes de tiempo". El operador limpió el goal con `/goal clear`.

### Pendiente inmediato (Gate 1 ya firmado → siguiente ola)
1. ✓ **Gate 1 FIRMADO** (2026-07-06). Base vs Soporte resuelto = **Base** (canónico). Sub-bandas conservadas
   como doctrinales pero re-encuadradas por **activación**.
2. **Fase C (spec)** → 🛑 Gate 2 → **Fase D (arquitectura, cementar las 6 decisiones as-code)** → **E/F**.
3. **Sincronizar los componentes de Storybook (SSOT) con el mockup firmado v3**: regiones tinte, caja+badge,
   spine-align + tier apoyo, activación/reglas colapsables, facet `origen`, transición del spine por caja,
   Guardia compacta, tokens `kind` 6→10. (El mockup divergió del Storybook durante la iteración rápida.)
4. **Bajar as-code (Fase D)** las 6 decisiones firmadas (VISION/METODOLOGIA/knowledge/arch): término «Base»,
   facet `activacion`, facet `origen`, knowledge as-code+semántico-opcional, discovery=data, la superficie del
   paquete/spine en la capa Estructura + el Inspector Hito 2.

### Preguntas abiertas de la deuda (de `…deuda-backend-arch.md`)
Superficie del conductor T3 (`POST /boxes/{id}/run`?) · `control_request` ask→UI · worktree vs main · ¿arrancar
la deuda en paralelo ya?

---

## 0 · Contexto y doctrina (leer ANTES de tocar nada)

Trabajas en **ArnesIA** (`/home/chalreme/Proyectos/harness-studio`), la fábrica de arneses por rol × proceso.
Estás en **fase 5 (Implementación), ficha HS-09 — el Mapa primero**. Lee, en este orden:

1. `CLAUDE.md` (norte del producto + Estado + Próximo).
2. `LEDGER.md` → ficha **HS-09** (el plan por hitos + la deuda paralela) y **HS-08** (el motor y el dogfood).
3. Memoria: `hs-09-fase5-mapa`, `hs-08-doctrina-ejecutable`, `hs-norma-pegarse-storybook`, `hs-arquitectura-fe`.
4. As-code FE: `arch/boundaries/fe-*.md` (topologia-fsd · taxonomia-componentes **canvas⊥chrome** ·
   transporte-independiente · tokens-contrato · visual-fitness) y `arch/conventions/`.
5. UX firmada del Mapa: `UX.md` (S2 lienzo/carriles/capas · S3 inspector) y `METODOLOGIA.md` (§3 contrato).

**Doctrina no-negociable de esta ola:**
- **Storybook es el SSOT del frontend.** Todo visual —incluidos los mockups y artifacts— **deriva de
  componentes reales de Storybook + tokens DTCG de `web/tokens/base.tokens.json`**. Prohibido inventar
  estilos o valores hex a mano. El mockup deja de ser hecho-a-mano: se **genera/actualiza desde Storybook**.
- **Backend agnóstico:** `fase`/`estado` son DATOS del arnés, nunca enums de producto. El Mapa lee lo que
  el manifiesto declara (`arnes.fases[]`, `arnes.spine`).
- **Hexagonal (backend) + FSD-lite (FE).** El Mapa = `widgets/map-canvas`; datos de dominio en
  `entities/arnes`; SSE singleton solo en `app/realtime/`. **canvas⊥chrome**: el canvas no importa chrome
  (rail/dock/shell) ni el chrome hace deep-import a internos del canvas — solo props/store.
- **Cero magic-values:** solo `var(--…)`; las 4 capas del Mapa = recoloreo `data-capa` de
  `--node-fill`/`--node-accent`, no paletas nuevas.
- **Nada se entrega sin correrlo:** `go build/vet/test ./...` verde (`-race`), `tsc` strictest, Biome limpio,
  dependency-cruiser limpio, **story = test** presente, y click-through real con screenshots + consola limpia.
- **No se programa sin arquitectura (NORMA):** antes de CUALQUIER código se revisa la arquitectura as-code
  (`arch/boundaries` + `arch/model` C4 + `arch/contracts` + mapeo `go-arch-lint`) y **se actualiza cuando el
  cambio lo exige** (puertos/adapters/componentes nuevos). No se codifica «por codificar»: la solución debe ser
  **escalable** y respetar los **principios de arquitectura hexagonal** (dominio puro; puertos; adapters
  intercambiables; el dominio NO depende de transporte/UI). Es una fase propia (D), bloqueante.
- **As-code al día:** al mover un boundary de `proposed`→`enforced`, edita su `status:` y re-sincroniza los
  conteos (knowledge/arch). No fabricar `pass`: un check sin enforcer determinista queda `deferred` honesto.
- **Git:** trunk-based (`main`), commit por trozo coherente, mensaje convencional con la firma Co-Authored-By.

---

## Fase A · Revisar el shell EXACTAMENTE (read-only → mapa preciso)

Objetivo: entender, `archivo:línea`, cómo funciona el shell de punta a punta, y **dónde exactamente se
enchufa el Mapa sin violar canvas⊥chrome**. Lanza subagentes en paralelo (Tauri+daemon / transporte+SSE /
FE shell), pero **verifica cada afirmación contra el archivo real** — no confíes en los docs.

Cubrir:
- **Tauri:** `web/src-tauri/src/lib.rs` (mint del `auth_token`, spawn del sidecar `externalBin`,
  single-instance, kill-on-exit), `main.rs` (`WEBKIT_DISABLE_DMABUF_RENDERER=1`), `tauri.conf.json`
  (`externalBin`, `devUrl`, `frontendDist`), `capabilities/default.json`.
- **Daemon (`arnesia serve`):** `cmd/arnesia/main.go::runServe` — cómo cablea index · agent (conductor CC) ·
  stores · watcher · broker · sessionSvc · http; flags (`-addr -claude -sessions -arneses -arnes-root
  -max-turns -auth-token`).
- **Transporte:** `internal/adapters/transport/http/router.go` (rutas + `withAuth`: Host anti-rebind, Origin
  allowlist reflejado, token constant-time) · `sse/broker.go` (multiplex `map|dock|run`, ring buffer,
  `Last-Event-ID` replay, shed-on-lag) · `sessions.go` (turn→202, `ErrBusy`→409) · `arneses.go`.
- **FE shell:** `web/src/pages/shell/ui/workspace-stage.tsx` (**el punto de montaje del Mapa**: hoy
  `<ComingSoon>` para `view==='mapa'`) · `widgets/{session-rail,topbar,view-strip,chat-dock}` ·
  `shared/store/{app-store (hash-state + tema), sessions-store}` · `shared/api/{client,sse,types}`.
- **Modelo de sesión:** N sesiones N:1 con arnés, subproceso CC por sesión, ruteo por `session_id`, dedup
  por `run_id`.

**Entregable:** `research/2026-07-06-shell-map.md` — el cableado exacto + el **seam de montaje del Mapa**
(cómo `workspace-stage` monta una vista nueva pasando `arnesId` como prop → respeta canvas⊥chrome) + cómo
una `widget`/`entity` nueva entra sin romper la topología FSD.

---

## Fase B · Storybook = SSOT → actualizar el mockup DESDE Storybook

Aquí nace la capa visual del Mapa **como stories reales** (esto ya es parte de Hito 1). El mockup se vuelve
un **artefacto derivado** de Storybook, no una fuente paralela. **Parte de lo que YA construimos — adaptar,
no descartar.**

0. **Partir de la referencia previa y ENTENDERLA.** Leer y comprender los mockups existentes
   (`mockups/arnesia-mockup-v3.html` = detalle más rico de la superficie/capas/inspector;
   `mockups/arnesia-shell-A-galaxia.html` = `mapView()` con `.mapbar`/`.band`/`.lanes`/`.lane`/`.node`) + las
   decisiones UX firmadas (`UX.md` S2 lienzo/carriles/capas · S3 inspector). **Extraer lo mejor** —la
   geografía de bandas/carriles, la barra del mapa con las capas, el inspector, la jerarquía visual, las
   interacciones— y **cargarlo a los componentes de Storybook**. La intención de diseño previa es el INSUMO
   que se preserva y mejora; no se tira. Anota qué se conserva, qué se mejora y por qué.
1. **Construir los componentes del canvas como stories** (SSOT), consumiendo solo `var(--…)` de
   `base.tokens.json`: `NodeShell`, `ProcessBox` (caja de fase), `GuardiaBand`, `BaseBand`, `BandLane`,
   `MapCanvas` (alimentado por el fixture `dogfood/dev-full-cycle.graph.json` como fixture grabado), y el
   panel `Inspector` (S3). **Story = test** (Storybook 10 + addon-vitest); React Flow exige
   `ReactFlowProvider` + contenedor con w/h; `fitView`/animaciones OFF en snapshot.
2. **Verificar** que Storybook levanta y las stories renderizan (screenshots + consola limpia; `tsc`
   strictest + Biome + dependency-cruiser limpios; stylelint anti-magic-value verde).
3. **Regenerar el mockup DESDE Storybook** (preservando la intención de diseño extraída en el paso 0): el
   mockup del Mapa debe reflejar los visuales reales de los componentes + los tokens. Los hex **deben calzar**
   con `base.tokens.json`. **Publicar al MISMO artifact** (parámetro `url`), click-through con asserts +
   screenshots revisados.
4. **🛑 GATE DE APROBACIÓN DEL OPERADOR.** Presentar el mockup actualizado (URL del artifact + screenshots +
   la nota de qué se conservó/mejoró) y **conseguir aprobación explícita del operador ANTES de escribir el
   spec (Fase C)**. Iterar sobre el feedback hasta el «apruebo». El mockup aprobado es la **referencia
   firmada** del Mapa; recién ahí se procede. NO redactar el spec sobre un mockup no aprobado.

Regla de oro: Storybook gana **solo en la implementación visual** (tokens, estilos, hex, forma exacta del
componente) — ahí el mockup se corrige contra Storybook. Pero la **intención de diseño** (layout de bandas/
carriles, capas, inspector, flujo de interacción) viene del mockup previo + UX firmada y **se conserva/mejora**;
Storybook NO la reinventa. Y sobre el resultado manda **la aprobación del operador** (paso 4).

---

## Fase C · Spec del MVP del Mapa CON TODO EL DETALLE

Con el mapa del shell (A) + Storybook + mockup actualizado (B) como insumos, escribe el spec firmable del
Mapa MVP (Hito 1 + Hito 2). Entregable: `research/2026-07-06-spec-mapa-mvp.md` (o donde corresponda como
as-code). Debe cubrir, sin huecos:

- **Backend — loader del dogfood:** `internal/adapters/index/store.go` — leer
  `dogfood/dev-full-cycle.graph.json`, `Unmarshal` a `domain.Graph`, registrar bajo id `dev-full-cycle`
  (flag `--seed-graphs <dir>` en `serve` o extender `seed()`/`Rebuild`). Con esto
  `GET /api/harnesses/dev-full-cycle/graph` sirve el arnés real (hoy sirve un demo de 2 nodos). Único cambio
  backend imprescindible del Hito 1.
- **FE transporte:** tipos TS del grafo con **claves español exactas** (`nodos`/`clase`/`banda`/`fase`/…),
  idealmente generados del schema/OpenAPI; `api.getGraph(id)` en `client.ts`; `sse.ts` escuchando `map`.
- **FE dominio:** slice `entities/arnes` (store con el `Graph` + selectores `selectGuardia`,
  `selectLanes(fases)`, `selectBase`, `selectEdges`); **no importa `shared/api`** (el fetch lo dispara la
  page/feature).
- **FE canvas (el grueso):** `widgets/map-canvas` — `<ReactFlow>` + **algoritmo de layout de carriles
  custom, determinista** (Guardia fila superior · una columna por `arnes.fases[]` en orden · Base fila
  inferior; posiciones fijas, NO auto-layout). Mapear `clase`→color/glyph, `edge.tipo`: `lee`=punteado sin
  flecha, `invoca`/`escribe`=flecha. Swap del `<ComingSoon>` por `<MapCanvas arnesId={s.arnes}/>`.
- **Inspector S3 (Hito 2):** backend `getNode` (Box+contract) + `listHarnesses` (hoy 501/`[]`); panel
  derecho identidad+contrato (el contrato fusionado ya lo alimenta); click nodo→inspector.
- **Capas:** solo **Estructura** en el MVP (tipo=color+forma+etiqueta + canal + edges). Tokens/Desempeño/
  Proceso quedan staged (requieren telemetría JSONL → indexer real; UX «arnés recién nacido» permite
  métricas «—»).
- **Aceptación (Gherkin) por comportamiento**, boundaries tocados, contratos/schemas, y las **stories como
  fitness**. Este spec es el artefacto que se firma antes de la implementación completa.

**🛑 GATE DE APROBACIÓN DEL OPERADOR.** Presentar el spec completo y **conseguir aprobación explícita del
operador ANTES de la Fase D (arquitectura) y de cualquier implementación (Fases E–F)**. Iterar sobre el
feedback hasta el «apruebo». Sin spec firmado no se avanza.

---

## Fase D · Arquitectura as-code (ANTES de programar — norma)

**No se programa por programar.** Con el spec aprobado (C) + el mapa del shell (A) como base, revisar y
**actualizar la arquitectura as-code** antes de escribir el código de las Fases E–F. La solución debe ser
**escalable** y respetar **arquitectura hexagonal** (dominio puro sin dependencias de transporte/UI; puertos;
adapters intercambiables). Cubrir:

- **Boundaries FE del Mapa** (`arch/boundaries/fe-*.md`): confirmar/actualizar que `widgets/map-canvas` (canvas)
  ⊥ chrome, `entities/arnes` no importa `shared/api`, SSE singleton en `app/realtime/`, tokens-contrato. Si el
  Mapa introduce un patrón nuevo (el algoritmo de carriles, la hidratación por selectores), reflejarlo en el
  boundary + `.dependency-cruiser.js` + `steiger.config.ts`.
- **Boundaries backend de la deuda** (`arch/boundaries/*.md`): el adapter concreto de `ports.ArtifactReader` y
  el cableado de `BoxConductor`/`KitProvisioner` caen bajo `conductor-no-parsea-jsonl`,
  `adaptadores-de-agente-intercambiables`, `permisos-derivan-del-rol`, `dominio-independiente-de-transporte`; el
  endpoint `control_request` bajo `permisos-gui-human-in-the-loop`. Mover `status: proposed→enforced` los que el
  nuevo código enforce de verdad; añadir el mapeo de componentes nuevos en `.go-arch-lint.yml`.
- **Modelo C4** (`arch/model/`): actualizar contexto/contenedor si nace un flujo nuevo (el Mapa como vista, el
  loader del índice, el endpoint de permisos).
- **Contratos** (`arch/contracts/`): si el Mapa necesita tipos generados (grafo TS) o `control_request` un shape
  nuevo, cementarlo en el schema/OpenAPI + `gen/`.
- **Escalabilidad:** dejar asentado cómo esto escala más allá del dogfood (N arneses en el índice, N sesiones,
  realtime SSE `map` en Hito 3, indexer JSONL real que reemplaza el stub) — sin sobre-ingeniería, pero con las
  costuras (puertos) listas para crecer.

Si un cambio toca un boundary **firmado/enforced**, señalarlo al operador antes de alterarlo. Salida: `arch/`
actualizado y coherente (conteos re-sincronizados) — **el mapa que el código de E–F debe respetar**.

---

## Fase E · Resolver la Deuda paralela (track backend, concurrente al Mapa)

Independiente del Mapa; puede correr en paralelo (subagentes / worktrees). Cada punto con tests; mover el
`status:` del boundary a `enforced` cuando aplique; sin fabricar `pass`.

1. **Cablear `BoxConductor` + `KitProvisioner` al daemon.** Diseñar e implementar un adapter concreto de
   `ports.ArtifactReader` (lee `result`+`status` del artefacto — **NUNCA el texto del chat**; la garantía
   anti-scrape está testeada). Cablear `BoxConductor` en `runServe` y `KitProvisioner` en la ruta de decisión
   de permisos. Hoy ambos existen reales+testeados pero solo instanciados en `arch/fitness`.
2. **Endpoint HTTP `control_request` con `role`/`ttl`** (hoy `resolvePermission` → 501). Definir la forma
   JSON-RPC (el spike que HS-04 reservó), cablear `PermissionSet`/`Grant` con TTL (`Grant.Vigente`).
3. **go-arch-lint operativo:** añadir `.go-arch-lint.yml` + asegurar el binario, para que los 3 boundaries
   dejen de `deferred` (hoy inoperable: sin binario ni config).
4. **`deferred`→real:** cablear los linters externos (Biome / golangci-lint / `tsc` / dependency-cruiser) a
   CI para que esos checks dejen de ser nl-judge deferred; que `conformance --todo` sea honesto sobre qué
   está realmente enforced (hoy: 235 checks pero **211 deferred**, solo 24 pass real).

---

## Fase F · Arrancar Hito 1 en vivo (enchufar los componentes de Storybook al app)

- Implementar el loader backend (~30 LOC) + los tipos+`api.getGraph`+`entities/arnes` + montar `<MapCanvas>`
  en `workspace-stage` para `view==='mapa'`, hidratando desde `GET graph` (fetch one-shot al seleccionar el
  arnés — el realtime SSE `map` es Hito 3, diferible).
- **Correrlo:** `arnesia serve` + el shell (Tauri o `devUrl`), navegar al Mapa, ver el arnés `dev-full-cycle`
  renderizado con sus carriles (Guardia vacía · spec/build/review/release · Base con `std-spec`). Screenshots
  + consola limpia.

---

## Gates / definición de hecho

- **G-shell:** `research/2026-07-06-shell-map.md` exacto (verificado por muestreo contra los archivos).
- **G-ssot:** stories del Mapa verdes (story=test); **mockup regenerado DESDE Storybook** preservando la
  intención de diseño previa (v3/galaxia + UX firmada), hex calzando con `base.tokens.json`; publicado al
  mismo artifact + click-through; **y APROBADO explícitamente por el operador** (referencia firmada del Mapa).
- **G-spec:** spec del Mapa MVP completo y coherente as-code (boundaries/contratos/Gherkin), **construido
  sobre el mockup ya aprobado**, **y APROBADO explícitamente por el operador antes de las Fases D–F** (sin spec
  firmado no se avanza a arquitectura ni a código).
- **G-arch:** `arch/` revisado y actualizado ANTES de programar (boundaries FE+backend, modelo C4, contratos,
  mapeo `go-arch-lint`); solución hexagonal + escalable; boundaries firmados no alterados sin aviso; conteos
  re-sincronizados.
- **G-debt:** conductor+permisos cableados (adapter `ArtifactReader`, testeado) · `control_request` vivo ·
  go-arch-lint operativo · conteo `deferred` reducido con honestidad; `go build/vet/test -race` verde.
- **G-hito1:** el arnés dogfood renderiza en el Mapa vivo (bandas + carriles), navegable, consola limpia,
  screenshots.
- **Cierre:** conteos as-code re-sincronizados (knowledge/arch), ficha **HS-09** actualizada en LEDGER,
  commit(s) por trozo. Si algo queda diferido, decláralo honesto (no lo pintes verde).

**Disciplina de ejecución:** subagentes para lecturas/verificación en paralelo + verificación adversarial de
afirmaciones; commit por trozo coherente; nada «hecho» sin correrlo. Si `firmado=congelado` aplica a un
entregable, fírmalo explícito con el operador antes de congelar.
