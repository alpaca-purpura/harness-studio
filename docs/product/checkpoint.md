# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-27** (barrido de deuda viva, 6 ítems — filtros Portafolio construidos ·
arquitectura de telemetría OTel resuelta · CI reparado [gosec + go-arch-lint + 10 lint] ·
botón «Correr» pausado por el operador, 2026-07-24). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Retomar aquí: MVP DE 1 DÍA — EN CURSO (2026-07-30).** Tres paquetes activos, corte y decisiones firmados
  vía aprobación del plan de sesión (prioridad 1 = B):
  [`stories/2026-07-30-volverlo-de-arnesia-y-publicar/`](./stories/2026-07-30-volverlo-de-arnesia-y-publicar/INDEX.md)
  (B1 sello válido + B2 publicar write-side) ·
  [`stories/2026-07-30-arnesia-en-el-proyecto/`](./stories/2026-07-30-arnesia-en-el-proyecto/INDEX.md)
  (semilla `.arnesia/` + `arnesia init` + enmienda doctrine) ·
  [`stories/2026-07-30-ciclo-conversacional-cables/`](./stories/2026-07-30-ciclo-conversacional-cables/INDEX.md)
  (Dock desde tarjeta de mejora + editar-conversando). Validación directo en laptop por bloque.
- **Las conversaciones viven en el panel — MERGEADO A MAIN (v0.4.0, merge `bb53421`; el aviso
  previo de «no integrado» quedó stale y se corrigió 2026-07-30).** Paquete
  [`stories/2026-07-26-conversaciones-del-panel/`](./stories/2026-07-26-conversaciones-del-panel/INDEX.md).
  Quedan: borrado manual de CV-D6 (T32) y las deudas N-21/N-22/N-23 (en `BACKLOG.md`); la firma
  🧑‍⚖️ de PARIDAD del paquete sigue del operador.
  **Origen:** el operador pidió ver el historial de una conversación y crear otra, y no encontró
  ninguna de las dos. Detrás había un problema de modelo: `Session` **era** la conversación.
  **Tramos 1 y 2:** los 9 campos del diálogo bajan a `Conversacion`; sobre versionado con migración
  forward-only, copia previa, cuarentena y modo solo-lectura; re-key de las llaves a medias (CV-D16);
  una ley invertida —archivar deja de tirar el transcript y el checkpoint—; la transición atómica con
  rollback total, la rotación que emite frame, el buscador dentro del texto y las 4 rutas nuevas con
  su contrato.
  **Tramos 3 y 4 — el frente, y ahora SE VE:** el cromo del dock baja de 4 filas a 2, la identidad
  técnica pasa a un chip desplegable que además dice **en qué carpeta corre** el conductor, la lista
  abre **en sitio** con buscador que resalta el fragmento que coincidió, y crear · retomar ·
  renombrar quedan cableados. **61 stories** nuevas y **17 tests** de store, todas verdes con el gate
  a11y entero. `tsc` fue de 0 a **18 errores** con T21 —era el objetivo: volver al compilador el
  inventario de lo que el tramo 2 había roto— y de vuelta a **0**.
  **Se cerró N-1**, una violación de contraste **que se envía hoy** (el `◍ <cc-id>` en `--primary`
  sobre `--secondary`, 2,21:1): al mudarse esa línea al detalle pasó a `--foreground`, y las 3
  stories del baseline **dejaron de apagar `color-contrast`**.
  **3 hallazgos nuevos:** el panel re-renderizaba el transcript entero en cada tecla del buscador
  (N-15); un `aria-controls` apuntando a la nada rompe axe cuando `aria-expanded="true"` —lo cazó el
  gate en la primera corrida— (N-16); y el contrato `<ul role="listbox">` del diseño no compila
  contra el linter del repo, se realiza sobre `<div>` con el contrato ARIA idéntico (N-17).
  **PARIDAD dibujo→producto ya se puede leer: 30 ✅ · 5 ⚠️ · 0 ❌**, con **26 capturas** en
  `verificacion-tramo3/` (13 escenas × 2 temas).
  Gate 🧑‍⚖️ de PARIDAD del paquete **sin firmar**: es del operador.


- **Dictado por voz en el composer — CONSTRUIDO Y VERIFICADO POR TESTS, falta el gate en vivo
  (2026-07-26).** El spike cerró y se ejecutó completo el mismo hilo. **Las 7 decisiones FIRMADAS 🧑‍⚖️**
  (2ª ronda): V-D1 = **A2 local `base`** con la sub-decisión resuelta como **adaptador por PATH**
  (detecta el motor instalado y degrada VISIBLE si no hay ninguno, en vez de elegir binario sobre
  evidencia que no existe → deuda **T7**); V-D2 glosario global + 2-3 turnos; V-D4 limpieza por defecto
  con escape a crudo; V-D5 audio no persistido; V-D6/V-D7 por no-objeción.
  **Construido, RF-215 primero** (el orden importaba: `wry 0.55.1` no maneja `permission-request` en
  Linux, así que sin el puente `getUserMedia` **no rechaza — queda pendiente para siempre** y el botón
  moriría mudo en la app instalada aunque ande en `pnpm dev`): puente Rust `webkit2gtk` que concede
  **solo audio** · `TranscriptionPort` + adaptador por PATH con transcodificado a cargo del adaptador ·
  endpoint `POST /api/sessions/{id}/dictado` que **siempre dice si el texto quedó `limpio` o `crudo`** ·
  limpieza como spawn **endurecido** (1 turno, 11 tools negadas, `--setting-sources project,local`,
  cableado a fitness porque el dictado es entrada NO confiable) · FE con las **etapas nombradas**
  (ningún spinner anónimo), medidor de nivel real, tope de 3 min que conserva lo grabado, y el composer
  poblado **editable que jamás se auto-envía**.
  **Verificación real:** Rust 4/4 · `go test ./...` verde · fitness verde · `npm run verify` verde ·
  13 stories del dictado (30/30 archivos del árbol corridos) · conformance
  `277 checks · pass 66 · fail 0` · capabilities **111 → 115** (4 nuevas, las 4 `vivo` por R4) ·
  `arch/` **21 → 22 boundaries**. Los fitness cazaron 4 cosas durante la obra (componente `stt` sin
  declarar en el grafo, 2 punteros R1 que no resolvían, 1 archivo sin capability) — el arnés muerde.
  **✅ CADENA COMPLETA PROBADA CON VOZ REAL (2026-07-26).** Se instaló un motor de verdad (venv
  aislado sin sudo) y corrió el pipeline entero por el daemon real: el STT devolvió «**warming**» y
  «**danon**», y la limpieza con contexto **los reparó a `warning`/`daemon`** y resolvió «la tarjeta
  esa de permisos» → **`PermissionCard`**, con `estado: "limpio"` en 13.4 s. Es lo que el spike §1.8
  nunca hizo: ahí se midió un script de Python, acá corrió el adaptador Go del paquete. E2E
  reproducible versionado (skipea sin motor, jamás finge).
  **Correr los binarios destapó 4 bugs que los fakes no veían:** se leía `stdout` en vez del `.txt`
  (habría poblado el composer con la charla del CLI), `--output_dir -` inválido, falta de
  `--device cpu` que hace **explotar** al CLI sin CUDA, y `faster-whisper` ofrecido como binario
  cuando **no expone ejecutable**. Y uno mayor: `exec.LookPath` habría dejado el motor **invisible
  para la app instalada** (el `.desktop` no hereda `~/.profile`) — mismo root cause que ya arregló
  `selfupdate.pathAumentado`, replicado acá y probado con `$PATH` limpio. También se arregló un
  footgun del `Makefile`: `make dev-sync` reportaba «Terminado» siempre porque su `pkill -f` se
  auto-mataba.
  **Instalador v0.2.20 armado** (`instaladores/v0.2.20/`, .deb/.rpm/.AppImage) + `make dev-sync`.
  **Falta SOLO el tramo del MICRÓFONO** contra el binario instalado para firmar 🧑‍⚖️ — ningún test lo
  reemplaza. Abierto además **T7** (qué motor se empaqueta; falta medir `whisper.cpp` y el
  **`audio/mp4` real**, que nunca se transcribió) y un **hallazgo ajeno**: `--warn` sobre `--card` en
  tema claro da 3.76:1 y rompe 4 stories de `session-rail` — al BACKLOG, no arreglado al voleo acá.
  [`stories/2026-07-25-spike-voz-dictado/INDEX.md`](stories/2026-07-25-spike-voz-dictado/INDEX.md)
  · [`PARIDAD.md`](stories/2026-07-25-spike-voz-dictado/PARIDAD.md)
  · [`spec.md`](stories/2026-07-25-spike-voz-dictado/spec.md)
  · [`decisiones.md`](stories/2026-07-25-spike-voz-dictado/decisiones.md)
- **Identidad de build en Ajustes — CONSTRUIDO Y REFINADO, falta el gate en vivo (2026-07-26).**
  Pedido del operador: «que en Ajustes aparezca un último número que sea el build». Se midió el
  agujero antes de tocar nada: la huella es el **commit**, no el build — ese mismo día se bundleó
  v0.2.21 **dos veces** desde el mismo árbol (dos binarios, huella idéntica) — y el semver ni
  siquiera estaba en el binario. Quedó **`arnesia v0.2.21.2607260225`** (semver + sello
  `AAMMDDHHMM`: cambia siempre, monótono, cero estado en el repo), con el commit debajo y un aviso
  accionable en la superficie cuando en disco hay un build más nuevo (`cerrá y reabrí la app` /
  `corré make dev-sync`). El sello se inyecta en **`bundle.sh` y no en el Makefile** a propósito:
  el self-update corre ese mismo script, así que la app **no pierde su identidad al actualizarse a
  sí misma**. Sin identidad inyectada dice `dev` — jamás un número inventado.
  **Verificación:** 12 tests Go de identidad · 15 stories de la tarjeta · daemon real (`GET
  /api/version` con y sin binario más nuevo en disco) · `go test ./...` y fitness verdes (R1-R4:
  `identidad.go` reclamado por CAP-60). **Refinamiento 2026-07-26** (2ª pasada, cero código): se
  agregaron `design.md`, la capa humana del spec (mapa funcional · RN-1..8 · AC-1..9 · matriz de
  cobertura sin huecos), B-D5/B-D6 (por qué sin mockup nuevo · los 4 límites del aviso declarados),
  y se re-derivó el snapshot `mockup-actualizar.html` (v3, + caso 06 del aviso, `charset` reparado).
  **VISTO EN VIVO (AC-9a, 2026-07-26):** se compiló, se levantó el daemon sellado con estado aislado
  (sin tocar `~/.arnesia` ni `~/.local/bin`) y se forzaron los **4 estados** en el navegador — al día ·
  build sin instalar · «cerrá y reabrí» · sin sellar (capturas en el `shots/` del paquete). Salieron
  **dos correcciones a la propia doc**: el límite «se evalúa por request» era peor de lo real (la
  tarjeta **refetchea al entrar a Ajustes**, sin recargar), y un **hallazgo ajeno** — el SPA tiene
  `127.0.0.1:4200` hardcodeado, en otro puerto la UI entera dice `Failed to fetch` (al BACKLOG).
  **Falta SOLO AC-9b**: la ventana Tauri instalada (`make installer` + sudo) para firmar 🧑‍⚖️.
  [`stories/2026-07-26-identidad-de-build/INDEX.md`](stories/2026-07-26-identidad-de-build/INDEX.md)
  · [`PARIDAD.md`](stories/2026-07-26-identidad-de-build/PARIDAD.md)
- **Versionado + changelog metodológicos — CONSTRUIDO Y PROBADO E2E, falta el ciclo real
  (2026-07-26).** Orden del operador: que actualizar la versión **y** decir qué se agrega/corrige/
  elimina no dependa de que alguien se lo recuerde a nadie. Dos huecos medidos: el repo tenía **20
  releases** en `instaladores/` y **cero changelog**, y `versionado.md` (v1.1, `enforced`) ni
  mencionaba el sello de build de RF-231. Quedó: `CHANGELOG.md` (Keep a Changelog, 6 categorías
  cerradas, `convencion-desde: 0.2.22` — **historia previa NO reconstruida**, sería inventar) ·
  `scripts/changelog.py` (check/add/release) · **`scripts/bump.sh` como punto único** que valida
  ANTES de tocar manifiestos y promueve `[Sin publicar]` → `## [X.Y.Z] — fecha` ·
  `make bump-patch|bump-minor|bump-major` con criterio escrito · `changelog_test.go` (3 tests, CI) ·
  job `changelog` de lefthook · `versionado.md` **v1.2** (4 checks nuevos + «identidad de build ≠
  versión de release») · la regla en `CLAUDE.md` y en metodología §10 para que una sesión nueva la
  lea sola. **E2E corrido de verdad y revertido:** bump 0.2.21→0.2.22 con promoción · segundo bump
  **abortado** por changelog vacío **sin tocar un solo archivo** · `bump-minor` 0.2.22→0.3.0 ·
  alias/idempotencia de `add` · rechazo de categoría inventada. Falta **AC-9**: un ciclo de
  publicación real para firmar 🧑‍⚖️.
  [`stories/2026-07-26-versionado-y-changelog-metodologicos/INDEX.md`](stories/2026-07-26-versionado-y-changelog-metodologicos/INDEX.md)
- **Portafolio · consolidar la agregación (proyecto + marketplace) — ACTIVO, en etapa de mockup
  (2026-07-24).** Re-escopeado por orden del operador (AG-D1): el paquete pasa de «agregar de
  marketplace» a consolidar TODA la agregación al Portafolio. **Auditoría visual en vivo hecha**
  (mockup · Storybook · app real a 1440×900): la app **sí** respeta el Storybook — pixel-idéntica
  a la story en dark; el que driftó es el mockup `.html`, stale en paleta (ámbar pre-rebrand) y
  adelantado en estructura (affordances que nunca bajaron a código). 1 bug real destapado en la
  toolbar («Marketplace» dos veces, lente y filtro sin distinción). Cero cambios en código de app
  a propósito — los hallazgos son insumo del spec, no bugfixes al voleo. **PENDIENTE-02 CERRADA
  2026-07-25:** 7 decisiones firmadas (AG-D1..D4, D6, D7 + PENDIENTE-01). **Único bloqueo del
  mockup: firma 🧑‍⚖️ de AG-D8** — plano «Marketplaces» reencuadrado contra `vision.md` (no es una
  tienda: es el estante de lo que vendemos + el espejo de si el cliente coincide; partición
  propio / de-referencia porque «operar arneses de terceros» está muerto por visión).
  [`stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md`](stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md)
  · [`auditoria-storybook-vs-app.md`](stories/2026-07-23-portafolio-agregar-marketplace/auditoria-storybook-vs-app.md)
- **Botón «Correr» de una caja — mockup publicado, PAUSADO por el operador (2026-07-24, HS-27).**
  Tras la explicación funcional del botón, el operador prefirió seguir con el resto del barrido
  antes de firmar. Retomar en [`stories/2026-07-23-boton-correr-caja/INDEX.md`](stories/2026-07-23-boton-correr-caja/INDEX.md).
- **Capa «Mejora» del Mapa (ex «capa Tokens») — investigación CERRADA y VERIFICADA EN VIVO, sin
  construir (HS-27/HS-28, 2026-07-26).** Tres carriles de investigación SOTA (13 runtimes ·
  plataformas OSS + licencias · stack embebible) + **verificación empírica contra `claude 2.1.220`**
  (receptor OTLP casero, 3 corridas, USD 0,044) que **corrigió 5 afirmaciones del diseño**: el canal
  primario es `/v1/logs` (`api_request`), no `/v1/metrics` · el split de cache 5m/1h viene en el
  `result` del stream-json ⇒ **D8 se cerró sin enmendar ningún boundary** · `pdata` cuesta
  **+10,79 MB** medidos (no +1,7) ⇒ OTLP/JSON + stdlib, **+0,49 MB** · `plugin_id_hash` rescata la
  atribución por arnés pese a la redacción `third-party` · 🔴 **la telemetría arrastra PII**
  (email + ids de cuenta en cada punto). **El MVP es el JOIN dinero × proceso** (D12.2), alcance
  S1 + S2, cero egreso. Refinamiento pre-mockup hecho: 6 detectores del MVP, evento canónico con
  `atribucion_confianza`, renombre de la capa. **Falta la firma 🧑‍⚖️ de D9/D11/D13/D14/D15/D16 y
  después el mockup.** Boundary v2.2 →
  [`stories/2026-07-24-telemetria-embebida-otel/INDEX.md`](stories/2026-07-24-telemetria-embebida-otel/INDEX.md)
  · [`verificacion-2026-07-26/INFORME.md`](stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/INFORME.md).
- **Chat dock · legibilidad y ergonomía (2026-07-22) — CERRADO Y FIRMADO 🧑‍⚖️ (HS-26).**
  Las 6 quejas + cariño markdown, verificadas E2E vivo (claude real contra vitalia); CAP-99 +
  CAP-100. Paquete [`stories/2026-07-22-chat-dock-ux/`](stories/2026-07-22-chat-dock-ux/INDEX.md).
  Deuda viva (no reabre): `make installer` para llevarlo al escritorio → `BACKLOG.md`; errores
  posteriores = bugfix en paquete nuevo.
- **Programa «Portafolio · ciclo de vida del arnés» — modelo FIRMADO 🧑‍⚖️ (HS-22); Slice 0
  FIRMADO 🧑‍⚖️ (HS-23); Slice 1 «FE» FIRMADO 🧑‍⚖️ (HS-25); Slice 2 «sello/Identificar» FIRMADO 🧑‍⚖️ (HS-25).**
  El spike de CARGA se reencuadró en el front-door del ciclo de vida (agregar de marketplace/proyecto · observar · mejorar ·
  publicar · actualizar · reparar). Modelo firmado: identidad **`(home,id)`** · N:M:M · **canónico** (editable) + N **instalaciones**
  (read-only) · ley anti-drift descriptiva · **`deriva`**. Revisión adversaria (4 subagentes) probó que el FE no es construible sin
  cimientos → recut **Slice 0 «Cimientos» (dominio) → Slice 1 «FE»** (épica en `BACKLOG.md`). Paquete
  [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md) (specs + casuística + revisión + mockup v2).
  **Hand-off:** Fable 5 (arquitecto) planificó → Sonnet 5 (constructor) ejecutó Slice 0 T1-T9 y Slice 1
  T1-T8 — los 8 tickets YA en `main` (T8 = `b35ff6c`, commiteado por el operador). **Auditoría post-build
  de la carga (Fable 5, 2026-07-14) EJECUTADA:** build sano; CI-rojo por lint preexistente REPARADO,
  motor conformance extendido a tests colocados (pass 42→46), docs stale sincronizadas →
  [`stories/2026-07-14-auditoria-portafolio-carga/`](stories/2026-07-14-auditoria-portafolio-carga/informe.md).
  Ver «Retomar aquí».
- **Fase 1 «forja-ciclo-vivo» — PAUSADA** ([`stories/2026-07-10-forja-ciclo-vivo/`](stories/2026-07-10-forja-ciclo-vivo/INDEX.md)):
  Slice 1a (golden scaffold `docs/terreno/` dim piloto `forma-trabajo` + `docs/wip/`) **construido, pendiente firma
  🧑‍⚖️**; verificado verde (F-D4). **Aclaración F-D5:** la expertise de forja se encarna en el **kit ② embebido**
  (cuerpos ①②, HS-10) — se inyecta al chat in-app; NO es `docs/terreno/` top-level (eso es un ③ dogfood mal-ubicado,
  migrará, no revertir). Firmado «docs/terreno absorbe architecture» → bajo revisión (P9 en terreno-conocimiento).
- **Modelo de TERRENO FIRMADO 🧑‍⚖️ 2026-07-10** ([`stories/2026-07-10-terreno-conocimiento/`](stories/2026-07-10-terreno-conocimiento/INDEX.md):
  D0-D20 + `arnes.yaml` + `estructura-terreno.html` v4). Sigue válido como **schema que el forjador aplica**.
  Slice 0/1/2 los 3 FIRMADOS (`ledger/HS-23.md`, `ledger/HS-25.md`). Slice 1 3 superficies (Lista/
  lente-empresa · Wizard-Proyecto/carpeta-local · Drawer READ) + Observar/Desvincular, E2E vivo (GAP-1
  cerrado); Slice 2 huella de path + modo degradado del Mapa + acción «Identificar» in-situ, E2E vivo
  (circuito crudo→degradado→identificar→sellado). Próximo paso del programa (item 2, «Agregar de
  marketplace → clonar + mejorar») desbloqueado en `BACKLOG.md`.
  Pausadas: **Fase 1 forja** (Slice 1a pendiente firma) · **terreno** (schema firmado). Contexto: memorias
  `hs-forja-fase1-carga-spike` · `hs-terreno-modelo-firmado` · `hs-linea-base-ui-storybook-ssot` · `hs-chat-cc-funcional`.
- **Rebrand PRENTER (2026-07-15) — FIRMADO 🧑‍⚖️ (HS-25).** Tokens dark-first en `web/`, paleta
  funcional del Mapa intacta. Deuda residual (no bloquea): re-derivar `arnesia-shell-A-galaxia.html` +
  re-estampar `mockups/INDEX.md` → `BACKLOG.md`.
- **Shell — Topbar sin empresa + selector de arnés (2026-07-20) — FIRMADO 🧑‍⚖️ (HS-25).**
  `topbar.tsx`/`session-rail.tsx`/`new-session-picker.tsx`/`portafolio-picker-store.ts`, 168/168 verde.
- **Paquete `mejorar-arnes-conversando` — IMPLEMENTADO COMPLETO 2026-07-22
  (ejecución autónoma /goal), pendiente SOLO gate humano 🧑‍⚖️ de PARIDAD.**
  [`stories/2026-07-22-mejorar-arnes-conversando/INDEX.md`](stories/2026-07-22-mejorar-arnes-conversando/INDEX.md) —
  spike cerrado (Forks A4/B2/C + grounding firmados, MC-D1..D8) → `spec.md` RF-183..206 → los 11
  tickets landeados con E2E vivo (evidencia en `PARIDAD.md`): loader reconoce `.claude/rules/`
  (Vitalia 2→7 nodos) · reindex-tras-turno + `event: map` + refetch FE (Mapa en vivo) · tarjeta de
  identidad por sesión (el chat SABE qué arnés/copia edita; loop A4 operó solo: «Causa
  diagnosticada: instalación») · sesión de reparación rotulada + deriva re-evaluada · `ctxPct`
  REPARADO (medía acumulado: 100 % espurio → 14 % real) + rotación invisible probada
  (`ClaudeSessionID` rotó, checkpoint mecánico, conversación continua) · historial B2 (Close
  archiva metadata + lector JSONL nativo cosió 9 turnos de 2 JSONLs, endpoints + picker). Reparación
  REAL de Vitalia hecha por chat: sello con rol + skill `hipaa-check`. 3 bugs destapados por E2E y
  reparados (roleFor sin índice al crear sesión · ctxPct acumulado · Cwd sin estampar). 5
  capabilities nuevas (CAP-94..98), cobertura 100 %. Deuda honesta en `PARIDAD.md` §Estado global.
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `328 checks · pass 93 · fail 0 · error 0 · deferred 235 · n/a 0` (medido 2026-07-30, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-30
- **arch/:** 28 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 149 — 102 vivo · 32 vivo·nc · 13 parcial · 2 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
