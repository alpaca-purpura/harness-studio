# Goal — HS-09 Hito 1 (MVP del Mapa) + Deuda paralela · Storybook = SSOT de FE

> Prompt de ejecución para una sesión fresca de Claude Code (usa subagentes para leer/verificar en
> paralelo; nada se da por hecho sin correrlo). Orden mandado por el operador: **revisar el shell exacto →
> actualizar el mockup DESDE Storybook (con aprobación del operador) → spec detallado (con aprobación del
> operador) → arquitectura as-code → resolver deuda paralela + arrancar Hito 1**. Dos gates de firma
> (mockup, spec) y una fase de arquitectura son BLOQUEANTES antes de programar.

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
