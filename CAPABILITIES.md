# CAPABILITIES — ArnesIA · SSoT funcional (docs-as-code)

> vig: activo · revisar: 2026-08-01
> **Single Source of Truth de lo que el sistema HACE.** Derivado del CÓDIGO REAL (barrido de 7
> subagentes sobre cmd/internal/web/src-tauri, 2026-07-09), no de fichas. Cada capability apunta
> al código autoritativo (`file#Símbolo`) y a su validador. NO es historia (`ledger/`) ni backlog.
> **Estado GENERADO** (D2/RF-181): `vivo`=validado por test/story · `vivo·nc`=implementado sin
> check automático · `parcial`=parcialmente implementado · `stub`=declarado, no implementado.
> Doctrina de enforcement (hook «nada de código sin capability») →
> `arch/boundaries/codigo-traza-a-capability.md`. FIRMADO 2026-07-09 (ledger HS-18).
> Un solo archivo por ahora; si crece se parte en `capabilities/<area>.md` (hojas atómicas) + índice.

## A · Superficie CLI + daemon
- **CAP-01 · Servir daemon** `infra·nc` — `cmd/arnesia/main.go#runServe:90` — composition-root :4200 (cablea adapters+usecases+HTTP/SSE+UI).
- **CAP-02 · Indexar arnés (CLI `index`)** `mapear·nc` — `cmd/arnesia/main.go#runIndex:334` — dir→graph.l0 JSON.
- **CAP-03 · Conformance CLI (`elemento|--arnes|--todo`)** `observar·vivo` — `cmd/arnesia/conformance.go#runConformance:26` — valida: `internal/domain/conformance_test.go`.
- **CAP-04 · Publicar al marketplace (`publish`)** `mejorar·STUB` — `cmd/arnesia/main.go#runPublish:380` + `internal/adapters/publish/publisher.go#Publish` → `errNotImplemented`.
- **CAP-05 · Abrir shell (`open`)** `infra·STUB` — `cmd/arnesia/main.go#runOpen:326` (TODO fase 5).

## B · Dominio L0 (modelo agnóstico)
- **CAP-06 · Grafo agnóstico L0 (Graph/Arnes/Edge/manifiesto)** `mapear·nc` — `internal/domain/graph.go#Graph:129,#Arnes:30,#Edge:10`.
- **CAP-07 · Taxonomía Clase (10 primitivas)** `mapear·nc` — `internal/domain/box.go#Clase:29,#Clase.Valid:50`.
- **CAP-08 · Contrato fusionado de caja (4 ejes)** `crear·vivo` — `internal/domain/box.go#Contract:226,#IsCaja:217` — valida: `box.contract.schema.json`.
- **CAP-09 · Ejes de autonomía (arquetipo×perfil) + document-as-cache** `crear·vivo` — `internal/domain/box.go#Arquetipo:62,#PerfilHarness:94,#RequiereDocumentAsCache:122` — valida: `TestRequiereDocumentAsCache`.
- **CAP-10 · Resolución de insumos (cableado)** `mapear·vivo` — `internal/domain/box.go#InsumosDe:307` — valida: `TestInsumosDe`.
- **CAP-11 · FSM determinista de caja (T3)** `crear·vivo` — `internal/domain/caja_fsm.go#AvanzarCaja:66,#RutaSiguiente:86` — valida: `TestAvanzarCaja,TestRutaSiguiente`.
- **CAP-12 · Spine + 5 categorías fijas (interop I-77)** `mapear·vivo` — `internal/domain/graph.go#Spine:49,#Categoria:62,#TransicionLegal:106` — valida: `TestVerificarSpineCategorias`.
- **CAP-13 · Permisos derivados del rol (deny>ask>allow, TTL)** `crear·nc` — `internal/domain/permission.go#PermissionSet.Decide:37,#NuevoGrant:68`.
- **CAP-14 · Sesión = frente de trabajo** `crear·nc` — `internal/domain/session.go#Session:65,#Salud:32`.

## C · Reconocimiento & carga (loader)
- **CAP-15 · Reconocer forma física (plugin|instalado)** `mapear·vivo` — `internal/adapters/loader/loader.go#detectarElementos:129,#LoadArnes:57` — valida: `TestLoaderInstalado,TestLoaderNoEsArnes`.
- **CAP-16 · Cargar arnés a grafo L0** `mapear·vivo` — `internal/adapters/loader/loader.go#LoadArnes:57` (estampa fuente_path/clase/procedencia) — valida: `TestLoaderDevFullCycle`.
- **CAP-17 · Leer manifiesto (degradado honesto)** `mapear·vivo` — `internal/adapters/loader/loader.go#leerManifiesto:143` — valida: `TestLoaderSinManifiesto`.
- **CAP-18 · Reconocedores clase→ubicación (8 tipos)** `mapear·vivo` — skill `internal/adapters/loader/loader.go#reconocerSkills:164` · hook `#reconocerHooks:280` · rule `internal/adapters/loader/regla.go#reconocerRegla:25` · command/output-style `internal/adapters/loader/soporte.go#reconocerArchivosSoporte:26` · mcp `internal/adapters/loader/soporte.go#reconocerMCP:71` · settings/statusline `internal/adapters/loader/soporte.go#reconocerSettings:119` — valida: `TestReconocer*`.
- **CAP-19 · Derivar edges (necesita→lee/invoca)** `mapear·vivo` — `internal/adapters/loader/edges.go#derivarEdges:27` — valida: `TestLoaderDevFullCycle`.
- **CAP-20 · Reconciliación honesta (no-reconocido visible)** `mapear·vivo` — `internal/adapters/loader/loader.go#nodoNoReconocido:338` — valida: `TestLoaderNoReconocido`.

## D · Índice & persistencia
- **CAP-21 · Índice de arneses en memoria** `crear·vivo` ⚠ — `internal/adapters/index/store.go#Store:26,#Upsert:76,#Query:49` — valida: `TestListPortfolio`. **NO es SQLite** (in-memory map; SQLite=fase 5).
- **CAP-22 · Reconstrucción del índice (seed)** `crear·parcial` — `internal/adapters/index/store.go#Rebuild:43,#seed:96` — valida: `TestSeedServesDogfood`. STUB: re-siembra embebidos, TODO recorrer JSONL `~/.claude`.
- **CAP-23 · Observar cambios del corpus (watcher)** `observar·STUB` — `internal/adapters/watch/watcher.go#Watch:23` — no-op (TODO fsnotify).
- **CAP-24 · Registro arnés→working-dir (confinamiento cwd + denylist)** `infra·nc` — `internal/adapters/store/arnes_registry.go#Register:78,#Resolve:62,#checkProtected:134`.
- **CAP-25 · Persistencia (sesiones + arnés→path, JSON atómico)** `infra·nc` ⚠ — `internal/adapters/store/registry.go#Save:65`; `internal/adapters/store/arnes_registry.go#saveLocked:200`. **JSON temp+rename, NO SQLite.**

## E · Conformance & doctrina (motor)
- **CAP-26 · Parsear ruleset a data (knowledge+arch→checks)** `mapear·vivo` — `internal/adapters/conformance/ruleset/parser.go#Loader.Load:52,#checklistRows:148` — valida: `TestLoadRuleset`.
- **CAP-27 · Inferir mecanismo + wiring por check** `mapear·vivo` — `internal/adapters/conformance/ruleset/parser.go#inferMechanism:207` — valida: `internal/adapters/conformance/ruleset/parser_test.go`.
- **CAP-28 · Ejecutar checks por mecanismo (arch-test/go-arch-lint/schema/static/nl-judge)** `observar·nc` — `internal/adapters/conformance/mechanism/adapters.go#ArchTest.Run:37,#GoArchLint.Run:115`; `internal/adapters/conformance/mechanism/schema.go#Validate:81`.
- **CAP-29 · Built-ins de conformidad del dominio (spine/escritor-único/composición/veredicto)** `observar·vivo` — `internal/domain/conformance.go#VerificarSpine:186,#VerificarEscritorUnico:447,#VerificarSinHuerfanos:532,#VerificarDeadEnds:558,#VerificarRutaExiste:612` — valida: `internal/domain/conformance_test.go` (12+ tests).
- **CAP-30 · Servicio de conformidad (rutea target + firewall CC-native)** `observar·vivo` — `internal/usecase/conformance_service.go#RunGraph:130,#firewallScan:200` — valida: `TestDogfoodArnesConforms,TestBrokenArnesFailsRightChecks`.
- **CAP-31 · Scope fabrica|arnes (portabilidad)** `infra·nc` — `internal/adapters/conformance/mechanism/adapters.go` (gating por `TargetArnes`/`repoRoot`).
- **CAP-32 · Doctrina embebida (portable)** `infra·nc` — `embed_doctrina.go#Files:21,#Kit:32` (go:embed ruleset+kit).

## F · Conductor CC & spawn
- **CAP-33 · Conducir sesión CC persistente (stream-json)** `crear·nc` — `internal/adapters/agent/claudecode/conductor.go#Spawn:152,#New:49`.
- **CAP-34 · Aislamiento de superficie de config al spawn** `infra·vivo` — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs:59` (`--setting-sources project,local`:70 · `--mcp-config`+`--strict-mcp-config`:100) — valida: `TestSpawnArgsSettingSourcesExcludeUser,TestSpawnArgsMCPAislado`.
- **CAP-35 · Inyección de doctrina por flags** `crear·nc` — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs:87` (`--plugin-dir`/`--append-system-prompt-file`/`--add-dir`).
- **CAP-36 · Cap de turnos (`--max-turns`)** `infra·nc` — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs:81`.
- **CAP-37 · Materializar permisos en flags CC (read-only, deny>allow)** `mapear·vivo` — `internal/adapters/agent/claudecode/conductor.go#permissionArgs:130` — valida: `TestPermissionArgsMaterialization`.
- **CAP-38 · Handshake initialize (arma can_use_tool)** `crear·nc` — `internal/adapters/agent/claudecode/conductor.go#sendInitialize:197`.
- **CAP-39 · HITL permisos (forward + respond wire)** `mapear·vivo` — `internal/adapters/agent/claudecode/conductor.go#RespondControl:337,#translate:516` — valida: `TestTranslateForwardsControlRequest,TestControlResponseLineWireFormat`.
- **CAP-40 · Streaming de turno + normalización a AgentEvent** `mapear·vivo` — `internal/adapters/agent/claudecode/conductor.go#Send:247,#pump:440,#translate:484` — valida: `TestTranslateForwardsControlRequest`.
- **CAP-41 · Interrupción en banda** `mapear·nc` — `internal/adapters/agent/claudecode/conductor.go#Interrupt:363`.
- **CAP-42 · Observabilidad de contexto (ctxPct)** `observar·nc` — `internal/adapters/agent/claudecode/conductor.go#ctxPct:554`.

## G · Provisioning & permisos
- **CAP-43 · Materializar doctrina+kit a ~/.arnesia (idempotente por huella)** `crear·vivo` — `internal/adapters/provision/provisioner.go#Provision:52,#fingerprint:143,#materializeMCPConfig:109` — valida: `TestProvisionMaterializesAndIsIdempotent`.
- **CAP-44 · Resolver permission-set por rol (deny-by-default)** `mapear·nc` — `internal/adapters/permission/provisioner.go#ResolveForRole:69`.

## H · Hand-off / artefactos
- **CAP-45 · Leer status del artefacto (señal máquina)** `mapear·vivo` — `internal/adapters/artifact/reader.go#Status:36` — valida: `TestStatusPresente,TestStatusAusente,TestFrontmatterRoto`.
- **CAP-46 · Digest acotado (economía de contexto, −90%)** `mapear·vivo` — `internal/adapters/artifact/reader.go#Resumen:63,#capBytes:105` — valida: `TestResumen`.
- **CAP-47 · Confinamiento de path del artefacto** `infra·vivo` — `internal/adapters/artifact/reader.go#confinedPath:118`; `internal/adapters/artifact/fuente.go#LeerConfinado:25` — valida: `TestPathQueIntentaEscapar,TestFuenteReaderConfinamiento`.

## I · Superficie HTTP/SSE
- **CAP-48 · Confinamiento de superficie local (Host+Origin+token)** `infra·vivo` — `internal/adapters/transport/http/auth.go#withAuth:65,#validToken:140` — valida: `internal/adapters/transport/http/auth_test.go`.
- **CAP-49 · Router + montaje** `infra·nc` — `internal/adapters/transport/http/router.go#NewHandler:27`.
- **CAP-50 · UI embebida servida por daemon** `observar·nc` — `embed_webdist.go#WebDist:14`; `internal/adapters/transport/http/router.go:34` (404 honesto sin dist).
- **CAP-51 · SSE multiplexado (map/dock/run, replay)** `observar·nc` — `internal/adapters/transport/sse/broker.go#Publish:69,#ServeHTTP:105,#replay:157`.
- **CAP-52 · Superficie REST (23 endpoints)** `funcional·nc` — `internal/adapters/transport/http/router.go#NewHandler:27` (tabla → Apéndice). Todos envueltos por `withAuth`.

## J · Usecases (aplicación)
- **CAP-53 · Dock: conversación en vivo multisesión** `crear·vivo` — `internal/usecase/session_service.go#Turn:256,#spawnLocked:305,#tryHealResume:450` — valida: `internal/usecase/session_permisos_test.go`.
- **CAP-54 · Gobierno del turno (permisos HITL + interrupt)** `observar·vivo` — `internal/usecase/session_service.go#onControlRequest:487,#ResolvePermission:555,#Interrupt:654` — valida: `TestResolvePermissionUsesArnesRole`.
- **CAP-55 · Ejecutar caja T3 desde daemon** `mejorar·nc` — `internal/usecase/run_service.go#RunBox:84`.
- **CAP-56 · Orquestación determinista (BoxConductor)** `infra·nc` — `internal/usecase/box_conductor.go#RunWith:74,#precondiciones:184,#tarea:142`.
- **CAP-57 · Servir fuente real de nodo** `observar·vivo` — `internal/usecase/fuente_service.go#Fuente:44` — valida: `TestFuenteService`.
- **CAP-58 · Mapa/portafolio/inspector (ensamblado)** `mapear·vivo` — `internal/usecase/map_service.go#Graph:24,#Harnesses:29,#Node:35` — valida: `TestMapServiceNode,TestMapServiceHarnesses`.
- **CAP-59 · Gestión de sesiones CRUD** `crear·nc` — `internal/usecase/session_service.go#Create:180,#Rename:199,#Close:229`.

## K · Self-update
- **CAP-60 · Self-update sin sudo (5 pasos atómico)** `mejorar·vivo` — `internal/usecase/selfupdate_service.go#Actualizar:91`; `internal/adapters/selfupdate/updater.go#Build:93,#Instalar:133,#Reiniciar:166` — valida: `internal/usecase/selfupdate_service_test.go,updater_test.go`.

## L · FE — Mapa
- **CAP-61 · Renderizar el Mapa (HTML+SVG)** `observar·vivo` — `web/src/widgets/map-canvas/ui/map-canvas.tsx#MapCanvas:53`; `web/src/entities/arnes/ui/arnes-node.tsx#ArnesNode:38` — valida: `web/src/widgets/map-canvas/ui/map-canvas.stories.tsx`.
- **CAP-62 · Pan/zoom/fit** `observar·nc` — `web/src/widgets/map-canvas/model/use-viewport.ts#useViewport`.
- **CAP-63 · Inspector drawer (Resumen|Contenido|Corridas)** `observar·vivo` — `web/src/widgets/map-canvas/ui/inspector.tsx#Inspector:133` — valida: `web/src/widgets/map-canvas/ui/inspector.stories.tsx`.
- **CAP-64 · Tab Contenido (fuente real, lazy)** `observar·nc` — `web/src/widgets/map-canvas/ui/inspector.tsx#Contenido:566`; `web/src/shared/api/client.ts#getNodeFuente:78`.
- **CAP-65 · Franja de artefactos (chips hand-off, off/auto/todos)** `observar·vivo` — `web/src/widgets/map-canvas/ui/handoff-gutter.tsx#HandoffGutter:24`; `web/src/entities/arnes/ui/artefacto-chip.tsx#ArtefactoChip:17` — valida: `web/src/widgets/map-canvas/ui/handoff-gutter.stories.tsx`.
- **CAP-66 · Picker de arnés** `mapear·vivo` — `web/src/widgets/map-canvas/ui/map-bar.tsx#arnes-picker:62` — valida: `web/src/widgets/map-canvas/ui/map-bar.stories.tsx`.
- **CAP-67 · Conmutador de capas** `observar·vivo` — `web/src/widgets/map-canvas/ui/map-bar.tsx#LAYERS:110`; `web/src/widgets/map-canvas/model/layers.ts` — valida: `web/src/widgets/map-canvas/ui/map-bar.stories.tsx`. Solo Estructura activa (resto «necesita telemetría»).

## M · FE — Chat/Dock
- **CAP-68 · Chat CC (turno/interrupt)** `crear·nc` — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ChatDock:10`; `web/src/shared/store/sessions-store.ts#sendTurn:168`.
- **CAP-69 · Acotar alcance (nodo→chip)** `mejorar·nc` — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ScopeRow:44`; `web/src/shared/store/sessions-store.ts#setScope:242`.
- **CAP-70 · Decidir permisos (tarjeta inline)** `mejorar·vivo` — `web/src/widgets/chat-dock/ui/permission-card.tsx#PermissionCard:7`; `web/src/shared/store/sessions-store.ts#resolvePermission:215` — valida: `web/src/widgets/chat-dock/ui/permission-card.stories.tsx`.
- **CAP-71 · Gate de conformance tras escrituras (RF-117)** `observar·nc` — `web/src/shared/store/sessions-store.ts` rama result :291.

## N · FE — Shell multisesión
- **CAP-72 · Rail de sesiones (crear/renombrar/cerrar/switch)** `crear·nc` — `web/src/widgets/session-rail/ui/session-rail.tsx#SessionRail:12`; `web/src/shared/store/sessions-store.ts#create:132`.
- **CAP-73 · View-strip (Mapa|Diag|Corridas|Tren|Hist)** `observar·nc` — `web/src/widgets/view-strip/ui/view-strip.tsx#ViewStrip:6`. Solo Mapa real.
- **CAP-74 · Topbar breadcrumb + ⌘K dock** `observar·nc` — `web/src/widgets/topbar/ui/topbar.tsx#Topbar:6`.
- **CAP-75 · Navegación global (portafolio/estándar/ajustes)** `infra·nc` — `web/src/pages/shell/ui/global-view.tsx#GlobalView:152`. Solo Ajustes real.
- **CAP-76 · Shell composition-root + hash-state routing** `infra·nc` — `web/src/pages/shell/ui/shell-page.tsx#ShellPage:16`; `web/src/shared/store/app-store.ts#bindHashState:38`.
- **CAP-77 · Botón actualizar (UI self-update)** `mejorar·vivo` — `web/src/features/self-update/ui/update-card.tsx#UpdateCard:55` — valida: `web/src/features/self-update/ui/update-card.stories.tsx`.
- **CAP-78 · Transporte FE (REST+SSE+token)** `infra·nc` — `web/src/shared/api/client.ts#api:51`; `web/src/shared/api/sse.ts#connectDock:16`.

## O · Shell Tauri
- **CAP-79 · Single-instance reenfoca** `infra·nc` — `web/src-tauri/src/lib.rs:52` (HS-14 fix ③).
- **CAP-80 · Sidecar del daemon (bind-or-bail, kill al salir)** `infra·nc` — `web/src-tauri/src/lib.rs:86,#daemon_running:141`.
- **CAP-81 · Inyección de token (raíz de confianza)** `infra·nc` — `web/src-tauri/src/lib.rs#mint_token:133` (env al daemon + init-script al WebView).
- **CAP-82 · Workaround render Linux (WEBKIT_DISABLE_DMABUF)** `infra·nc` — `web/src-tauri/src/main.rs:9`.

---

## Cobertura interna (soporte reclamado por su capability — R2)

Archivos de código que IMPLEMENTAN un capability ya listado pero no son su puntero autoritativo
(chrome/handlers). Reclamados aquí para cobertura (R2); el validador `capability-trace` los lee.

<!--coverage-->
- CAP-52 · handlers HTTP: `internal/adapters/transport/http/arneses.go` · `internal/adapters/transport/http/sessions.go` · `internal/adapters/transport/http/run.go` · `internal/adapters/transport/http/conformance.go` · `internal/adapters/transport/http/fuente.go` · `internal/adapters/transport/http/selfupdate.go`
- CAP-76 · shell stage: `web/src/pages/shell/ui/workspace-stage.tsx`
- CAP-61 · chrome interno del Mapa: `web/src/widgets/map-canvas/ui/band.tsx` · `web/src/widgets/map-canvas/ui/lane.tsx` · `web/src/widgets/map-canvas/ui/region.tsx` · `web/src/widgets/map-canvas/ui/base-band.tsx` · `web/src/widgets/map-canvas/ui/edge-layer.tsx` · `web/src/widgets/map-canvas/ui/activation-chip.tsx` · `web/src/widgets/map-canvas/ui/rules-subband.tsx` · `web/src/widgets/map-canvas/ui/help-panel.tsx` · `web/src/widgets/map-canvas/model/bands.ts` · `web/src/widgets/map-canvas/model/use-edge-paths.ts`
<!--/coverage-->

## Cobertura & honestidad (para la doctrina de enforcement)

- **82 capabilities** mapeadas a código real. Estado: ~24 `vivo` (test/story) · ~46 `vivo·nc`
  (sin check automático) · 3 `parcial/stub-dentro` · **6 STUB** (CAP-04 publish, CAP-05 open,
  CAP-23 watcher, CAP-22 seed-parcial, endpoint `listRuns` 501, `mint_token` sin test).
- **⚠ DRIFT CLAUDE.md ↔ código:** CLAUDE.md dice «SQLite puro-Go (modernc, WAL)». Realidad:
  índice = **map in-memory** (CAP-21), persistencia = **archivos JSON atómicos** (CAP-25).
  SQLite es fase 5 futura (comentarios en `index/store.go:1-5`, `store/registry.go:3-4`).
  → corregir en el router/ESTADO al migrar.
- **FE sin tests:** `web/` NO tiene `*.test.*`/`e2e/`; única validación = `.stories.tsx`. Los
  widgets de chrome (chat-dock, session-rail, topbar, view-strip), pages/shell y stores/api
  quedan `nc` pese a ser capabilities reales. Deuda de validación → BACKLOG.
- **Dead-code candidatos (sin capability):** `domain/graph.go#UnidadDeTrabajo:119` (sin uso) ·
  enums `Banda/Canal/Procedencia/Origen` en box.go (serializados, sin comportamiento propio) ·
  `app-store.ts#toggleTheme:25` (latente, sin control en UI). → decidir: reclamar o borrar.
- **Deuda declarada sin capability:** reconocedores pendientes del loader (`subagent`,
  plugin-raíz, edges de librería/Guardia) — `loader.go:20-28`.

## Apéndice — 23 endpoints REST (`transport/http/router.go`)

```
/                                              UI embebida | 404          CAP-50
GET  /healthz                                  liveness                   (auth-exento)
GET  /events, /api/events                      SSE                        CAP-51
GET  /api/version                              self-update identidad      CAP-60
POST /api/self-update                          self-update aplicar        CAP-60
GET  /api/harnesses                            portafolio                 CAP-58
GET  /api/harnesses/{id}/graph                 grafo                      CAP-58
GET  /api/harnesses/{id}/nodes/{nid}           inspector                  CAP-58
GET  /api/harnesses/{id}/nodes/{nid}/fuente    fuente real                CAP-57
GET  /api/harnesses/{id}/runs                  STUB 501                   —
GET  /api/harnesses/{id}/conformance           auditoría                  CAP-30
POST /api/harnesses/{id}/boxes/{bid}/run       correr caja T3             CAP-55
GET  /api/arneses · PUT /api/arneses/{id}      registro arnés→path        CAP-24
GET/POST/GET/PATCH/DELETE /api/sessions[...]   multisesión CRUD           CAP-59
POST /api/sessions/{id}/turn                   turno                      CAP-53
POST /api/sessions/{id}/permission             permisos HITL              CAP-54
POST /api/sessions/{id}/interrupt              interrupción               CAP-54
```
