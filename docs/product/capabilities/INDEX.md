# CAPABILITIES — ArnesIA · índice (SSoT funcional = las hojas `*.yaml`)

> vig: activo · revisar: 2026-08-01
> **La fuente de verdad son las hojas `{module}/{slug}.yaml`** (formato del kit
> `harness@prenter-marketplace`, homologación 2026-07-09). Cada capability apunta al código
> autoritativo (`pointers: file#Símbolo`) + su validador (`valida:`). Esta tabla es un índice
> **GENERADO** con `python3 scripts/cap_doctor.py --index` — no se teclea. `status` GENERADO
> (R4): `vivo`=validado por test/story · `vivo·nc`=implementado sin check · `parcial` · `stub`.
> Enforcement (R1 integridad · R2 cobertura · R3 gate-commit · R4 estado-generado) →
> `docs/architecture/boundaries/codigo-traza-a-capability.md`. Doctor rápido: `scripts/cap_doctor.py`.
> Migrado desde el CAPABILITIES.md monolítico (HS-18) a 82 hojas atómicas (homologación HS-19).

## Índice por módulo

<!--caps:begin-->
_Generado por `cap_doctor.py --index` desde las hojas `*.yaml` (SSoT). No editar a mano._


### `cli-daemon` (5)

- **CAP-01 · Servir daemon** `vivo·nc` · cli-daemon/servir-daemon.yaml — `cmd/arnesia/main.go#runServe`
- **CAP-02 · Indexar arnés (CLI `index`)** `vivo·nc` · cli-daemon/indexar-arnes.yaml — `cmd/arnesia/main.go#runIndex`
- **CAP-03 · Conformance CLI (`elemento|--arnes|--todo`)** `vivo` · cli-daemon/conformance-cli.yaml — `cmd/arnesia/conformance.go#runConformance`
- **CAP-04 · Publicar al marketplace (`publish`)** `stub` · cli-daemon/publicar-al-marketplace.yaml — `cmd/arnesia/main.go#runPublish`
- **CAP-05 · Abrir shell (`open`)** `stub` · cli-daemon/abrir-shell.yaml — `cmd/arnesia/main.go#runOpen`

### `dominio-l0` (9)

- **CAP-06 · Grafo agnóstico L0 (Graph/Arnes/Edge/manifiesto)** `vivo·nc` · dominio-l0/grafo-agnostico-l0.yaml — `internal/domain/graph.go#Graph`
- **CAP-07 · Taxonomía Clase (10 primitivas)** `vivo·nc` · dominio-l0/taxonomia-clase.yaml — `internal/domain/box.go#Clase`
- **CAP-08 · Contrato fusionado de caja (4 ejes)** `vivo` · dominio-l0/contrato-fusionado-de-caja.yaml — `internal/domain/box.go#Contract`
- **CAP-09 · Ejes de autonomía (arquetipo×perfil) + document-as-cache** `vivo` · dominio-l0/ejes-de-autonomia-document-as-cache.yaml — `internal/domain/box.go#Arquetipo`
- **CAP-10 · Resolución de insumos (cableado)** `vivo` · dominio-l0/resolucion-de-insumos.yaml — `internal/domain/box.go#InsumosDe`
- **CAP-11 · FSM determinista de caja (T3)** `vivo` · dominio-l0/fsm-determinista-de-caja.yaml — `internal/domain/caja_fsm.go#AvanzarCaja`
- **CAP-12 · Spine + 5 categorías fijas (interop I-77)** `vivo` · dominio-l0/spine-5-categorias-fijas.yaml — `internal/domain/graph.go#Spine`
- **CAP-13 · Permisos derivados del rol (deny>ask>allow, TTL)** `vivo·nc` · dominio-l0/permisos-derivados-del-rol.yaml — `internal/domain/permission.go#PermissionSet.Decide`
- **CAP-14 · Sesión = frente de trabajo** `vivo·nc` · dominio-l0/sesion-frente-de-trabajo.yaml — `internal/domain/session.go#Session`

### `loader` (6)

- **CAP-15 · Reconocer forma física (plugin|instalado)** `vivo` · loader/reconocer-forma-fisica.yaml — `internal/adapters/loader/loader.go#detectarElementos`
- **CAP-16 · Cargar arnés a grafo L0** `vivo` · loader/cargar-arnes-a-grafo-l0.yaml — `internal/adapters/loader/loader.go#LoadArnes`
- **CAP-17 · Leer manifiesto (degradado honesto)** `vivo` · loader/leer-manifiesto.yaml — `internal/adapters/loader/loader.go#leerManifiesto`
- **CAP-18 · Reconocedores clase→ubicación (8 tipos)** `vivo` · loader/reconocedores-claseubicacion.yaml — `internal/adapters/loader/loader.go#reconocerSkills`
- **CAP-19 · Derivar edges (necesita→lee/invoca)** `vivo` · loader/derivar-edges.yaml — `internal/adapters/loader/edges.go#derivarEdges`
- **CAP-20 · Reconciliación honesta (no-reconocido visible)** `vivo` · loader/reconciliacion-honesta.yaml — `internal/adapters/loader/loader.go#nodoNoReconocido`

### `indice-persistencia` (5)

- **CAP-21 · Índice de arneses en memoria** `vivo` · indice-persistencia/indice-de-arneses-en-memoria.yaml — `internal/adapters/index/store.go#Store`
- **CAP-22 · Reconstrucción del índice (seed)** `parcial` · indice-persistencia/reconstruccion-del-indice.yaml — `internal/adapters/index/store.go#Rebuild`
- **CAP-23 · Observar cambios del corpus (watcher)** `stub` · indice-persistencia/observar-cambios-del-corpus.yaml — `internal/adapters/watch/watcher.go#Watch`
- **CAP-24 · Registro arnés→working-dir (confinamiento cwd + denylist)** `vivo·nc` · indice-persistencia/registro-arnesworking-dir.yaml — `internal/adapters/store/arnes_registry.go#Register`
- **CAP-25 · Persistencia (sesiones + arnés→path, JSON atómico)** `vivo·nc` · indice-persistencia/persistencia.yaml — `internal/adapters/store/registry.go#Save`

### `conformance` (7)

- **CAP-26 · Parsear ruleset a data (knowledge+arch→checks)** `vivo` · conformance/parsear-ruleset-a-data.yaml — `internal/adapters/conformance/ruleset/parser.go#Loader.Load`
- **CAP-27 · Inferir mecanismo + wiring por check** `vivo` · conformance/inferir-mecanismo-wiring-por-check.yaml — `internal/adapters/conformance/ruleset/parser.go#inferMechanism`
- **CAP-28 · Ejecutar checks por mecanismo (arch-test/go-arch-lint/schema/static/nl-judge)** `vivo·nc` · conformance/ejecutar-checks-por-mecanismo.yaml — `internal/adapters/conformance/mechanism/adapters.go#ArchTest.Run`
- **CAP-29 · Built-ins de conformidad del dominio (spine/escritor-único/composición/veredicto)** `vivo` · conformance/built-ins-de-conformidad-del-dominio.yaml — `internal/domain/conformance.go#VerificarSpine`
- **CAP-30 · Servicio de conformidad (rutea target + firewall CC-native)** `vivo` · conformance/servicio-de-conformidad.yaml — `internal/usecase/conformance_service.go#RunGraph`
- **CAP-31 · Scope fabrica|arnes (portabilidad)** `vivo·nc` · conformance/scope-fabrica-arnes.yaml — `internal/adapters/conformance/mechanism/adapters.go`
- **CAP-32 · Doctrina embebida (portable)** `vivo·nc` · conformance/doctrina-embebida.yaml — `embed_doctrina.go#Files`

### `conductor` (10)

- **CAP-33 · Conducir sesión CC persistente (stream-json)** `vivo·nc` · conductor/conducir-sesion-cc-persistente.yaml — `internal/adapters/agent/claudecode/conductor.go#Spawn`
- **CAP-34 · Aislamiento de superficie de config al spawn** `vivo` · conductor/aislamiento-de-superficie-de-config-al-spawn.yaml — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs`
- **CAP-35 · Inyección de doctrina por flags** `vivo·nc` · conductor/inyeccion-de-doctrina-por-flags.yaml — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs`
- **CAP-36 · Cap de turnos (`--max-turns`)** `vivo·nc` · conductor/cap-de-turnos.yaml — `internal/adapters/agent/claudecode/conductor.go#SpawnArgs`
- **CAP-37 · Materializar permisos en flags CC (read-only, deny>allow)** `vivo` · conductor/materializar-permisos-en-flags-cc.yaml — `internal/adapters/agent/claudecode/conductor.go#permissionArgs`
- **CAP-38 · Handshake initialize (arma can_use_tool)** `vivo·nc` · conductor/handshake-initialize.yaml — `internal/adapters/agent/claudecode/conductor.go#sendInitialize`
- **CAP-39 · HITL permisos (forward + respond wire)** `vivo` · conductor/hitl-permisos.yaml — `internal/adapters/agent/claudecode/conductor.go#RespondControl`
- **CAP-40 · Streaming de turno + normalización a AgentEvent** `vivo` · conductor/streaming-de-turno-normalizacion-a-agentevent.yaml — `internal/adapters/agent/claudecode/conductor.go#Send`
- **CAP-41 · Interrupción en banda** `vivo·nc` · conductor/interrupcion-en-banda.yaml — `internal/adapters/agent/claudecode/conductor.go#Interrupt`
- **CAP-42 · Observabilidad de contexto (ctxPct)** `vivo·nc` · conductor/observabilidad-de-contexto.yaml — `internal/adapters/agent/claudecode/conductor.go#ctxPct`

### `provisioning` (2)

- **CAP-43 · Materializar doctrina+kit a ~/.arnesia (idempotente por huella)** `vivo` · provisioning/materializar-doctrina-kit-a-arnesia.yaml — `internal/adapters/provision/provisioner.go#Provision`
- **CAP-44 · Resolver permission-set por rol (deny-by-default)** `vivo·nc` · provisioning/resolver-permission-set-por-rol.yaml — `internal/adapters/permission/provisioner.go#ResolveForRole`

### `handoff` (3)

- **CAP-45 · Leer status del artefacto (señal máquina)** `vivo` · handoff/leer-status-del-artefacto.yaml — `internal/adapters/artifact/reader.go#Status`
- **CAP-46 · Digest acotado (economía de contexto, −90%)** `vivo` · handoff/digest-acotado.yaml — `internal/adapters/artifact/reader.go#Resumen`
- **CAP-47 · Confinamiento de path del artefacto** `vivo` · handoff/confinamiento-de-path-del-artefacto.yaml — `internal/adapters/artifact/reader.go#confinedPath`

### `http-sse` (5)

- **CAP-48 · Confinamiento de superficie local (Host+Origin+token)** `vivo` · http-sse/confinamiento-de-superficie-local.yaml — `internal/adapters/transport/http/auth.go#withAuth`
- **CAP-49 · Router + montaje** `vivo·nc` · http-sse/router-montaje.yaml — `internal/adapters/transport/http/router.go#NewHandler`
- **CAP-50 · UI embebida servida por daemon** `vivo·nc` · http-sse/ui-embebida-servida-por-daemon.yaml — `embed_webdist.go#WebDist`
- **CAP-51 · SSE multiplexado (map/dock/run, replay)** `vivo·nc` · http-sse/sse-multiplexado.yaml — `internal/adapters/transport/sse/broker.go#Publish`
- **CAP-52 · Superficie REST (23 endpoints)** `vivo·nc` · http-sse/superficie-rest.yaml — `internal/adapters/transport/http/router.go#NewHandler`

### `usecases` (7)

- **CAP-53 · Dock: conversación en vivo multisesión** `vivo` · usecases/dock-conversacion-en-vivo-multisesion.yaml — `internal/usecase/session_service.go#Turn`
- **CAP-54 · Gobierno del turno (permisos HITL + interrupt)** `vivo` · usecases/gobierno-del-turno.yaml — `internal/usecase/session_service.go#onControlRequest`
- **CAP-55 · Ejecutar caja T3 desde daemon** `vivo·nc` · usecases/ejecutar-caja-t3-desde-daemon.yaml — `internal/usecase/run_service.go#RunBox`
- **CAP-56 · Orquestación determinista (BoxConductor)** `vivo·nc` · usecases/orquestacion-determinista.yaml — `internal/usecase/box_conductor.go#RunWith`
- **CAP-57 · Servir fuente real de nodo** `vivo` · usecases/servir-fuente-real-de-nodo.yaml — `internal/usecase/fuente_service.go#Fuente`
- **CAP-58 · Mapa/portafolio/inspector (ensamblado)** `vivo` · usecases/mapa-portafolio-inspector.yaml — `internal/usecase/map_service.go#Graph`
- **CAP-59 · Gestión de sesiones CRUD** `vivo·nc` · usecases/gestion-de-sesiones-crud.yaml — `internal/usecase/session_service.go#Create`

### `self-update` (1)

- **CAP-60 · Self-update sin sudo (5 pasos atómico)** `vivo` · self-update/self-update-sin-sudo.yaml — `internal/usecase/selfupdate_service.go#Actualizar`

### `fe-mapa` (7)

- **CAP-61 · Renderizar el Mapa (HTML+SVG)** `vivo` · fe-mapa/renderizar-el-mapa.yaml — `web/src/widgets/map-canvas/ui/map-canvas.tsx#MapCanvas`
- **CAP-62 · Pan/zoom/fit** `vivo·nc` · fe-mapa/pan-zoom-fit.yaml — `web/src/widgets/map-canvas/model/use-viewport.ts#useViewport`
- **CAP-63 · Inspector drawer (Resumen|Contenido|Corridas)** `vivo` · fe-mapa/inspector-drawer.yaml — `web/src/widgets/map-canvas/ui/inspector.tsx#Inspector`
- **CAP-64 · Tab Contenido (fuente real, lazy)** `vivo·nc` · fe-mapa/tab-contenido.yaml — `web/src/widgets/map-canvas/ui/inspector.tsx#Contenido`
- **CAP-65 · Franja de artefactos (chips hand-off, off/auto/todos)** `vivo` · fe-mapa/franja-de-artefactos.yaml — `web/src/widgets/map-canvas/ui/handoff-gutter.tsx#HandoffGutter`
- **CAP-66 · Picker de arnés** `vivo` · fe-mapa/picker-de-arnes.yaml — `web/src/widgets/map-canvas/ui/map-bar.tsx#arnes`
- **CAP-67 · Conmutador de capas** `vivo` · fe-mapa/conmutador-de-capas.yaml — `web/src/widgets/map-canvas/ui/map-bar.tsx#LAYERS`

### `fe-chat` (4)

- **CAP-68 · Chat CC (turno/interrupt)** `vivo·nc` · fe-chat/chat-cc.yaml — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ChatDock`
- **CAP-69 · Acotar alcance (nodo→chip)** `vivo·nc` · fe-chat/acotar-alcance.yaml — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ScopeRow`
- **CAP-70 · Decidir permisos (tarjeta inline)** `vivo` · fe-chat/decidir-permisos.yaml — `web/src/widgets/chat-dock/ui/permission-card.tsx#PermissionCard`
- **CAP-71 · Gate de conformance tras escrituras (RF-117)** `vivo·nc` · fe-chat/gate-de-conformance-tras-escrituras.yaml — `web/src/shared/store/sessions-store.ts`

### `fe-shell` (7)

- **CAP-72 · Rail de sesiones (crear/renombrar/cerrar/switch)** `vivo·nc` · fe-shell/rail-de-sesiones.yaml — `web/src/widgets/session-rail/ui/session-rail.tsx#SessionRail`
- **CAP-73 · View-strip (Mapa|Diag|Corridas|Tren|Hist)** `vivo·nc` · fe-shell/view-strip.yaml — `web/src/widgets/view-strip/ui/view-strip.tsx#ViewStrip`
- **CAP-74 · Topbar breadcrumb + ⌘K dock** `vivo·nc` · fe-shell/topbar-breadcrumb-k-dock.yaml — `web/src/widgets/topbar/ui/topbar.tsx#Topbar`
- **CAP-75 · Navegación global (portafolio/estándar/ajustes)** `vivo·nc` · fe-shell/navegacion-global.yaml — `web/src/pages/shell/ui/global-view.tsx#GlobalView`
- **CAP-76 · Shell composition-root + hash-state routing** `vivo·nc` · fe-shell/shell-composition-root-hash-state-routing.yaml — `web/src/pages/shell/ui/shell-page.tsx#ShellPage`
- **CAP-77 · Botón actualizar (UI self-update)** `vivo` · fe-shell/boton-actualizar.yaml — `web/src/features/self-update/ui/update-card.tsx#UpdateCard`
- **CAP-78 · Transporte FE (REST+SSE+token)** `vivo·nc` · fe-shell/transporte-fe.yaml — `web/src/shared/api/client.ts#api`

### `tauri` (4)

- **CAP-79 · Single-instance reenfoca** `vivo·nc` · tauri/single-instance-reenfoca.yaml — `web/src-tauri/src/lib.rs`
- **CAP-80 · Sidecar del daemon (bind-or-bail, kill al salir)** `vivo·nc` · tauri/sidecar-del-daemon.yaml — `web/src-tauri/src/lib.rs#daemon_running`
- **CAP-81 · Inyección de token (raíz de confianza)** `vivo·nc` · tauri/inyeccion-de-token.yaml — `web/src-tauri/src/lib.rs#mint_token`
- **CAP-82 · Workaround render Linux (WEBKIT_DISABLE_DMABUF)** `vivo·nc` · tauri/workaround-render-linux.yaml — `web/src-tauri/src/main.rs`

### `fe-portafolio` (4)

- **CAP-89 · Lista del Portafolio (lente empresa/plano, buscar, corruptas visibles)** `vivo` · fe-portafolio/lista-del-portafolio.yaml — `web/src/entities/portafolio/model/types.ts#EntradaPortafolio`
- **CAP-90 · Drawer de detalle READ (identidad, facetas, origen de la copia, instalaciones)** `vivo` · fe-portafolio/drawer-detalle-read.yaml — `web/src/entities/portafolio/model/types.ts#Instalacion`
- **CAP-91 · Wizard agregar proyecto (escanear carpeta local → candidatos honestos → elegir)** `vivo` · fe-portafolio/wizard-agregar-proyecto.yaml — `web/src/entities/portafolio/model/types.ts#Candidato`
- **CAP-92 · Abrir en Mapa desde el Portafolio (peek al stage de sesión, colisión visible)** `vivo` · fe-portafolio/abrir-en-mapa.yaml — `web/src/shared/api/client.ts#observarEnMapa`

### `portafolio` (6)

- **CAP-83 · Registrar identidad de arnés (home,id)** `vivo` · portafolio/registrar-identidad.yaml — `internal/domain/portafolio.go#IdentidadArnes`
- **CAP-84 · Escanear proyecto (walker multi-instalación)** `vivo` · portafolio/escanear-proyecto.yaml — `internal/domain/portafolio.go#TipoInstalacion`
- **CAP-85 · Resolver origen (collect-all + reconcilia)** `vivo` · portafolio/resolver-origen.yaml — `internal/domain/portafolio.go#OrigenPortafolio`
- **CAP-86 · Evaluar deriva (hash vs referencia inmutable)** `vivo` · portafolio/evaluar-deriva.yaml — `internal/domain/portafolio.go#EstadoDeriva`
- **CAP-87 · Desvincular arnés del Portafolio** `vivo` · portafolio/desvincular.yaml — `internal/domain/portafolio.go#EntradaPortafolio`
- **CAP-88 · Observar en Mapa (presencia read-only del Portafolio)** `vivo` · portafolio/observar-en-mapa.yaml — `internal/usecase/portafolio.go#PortafolioService.ObservarEnMapa`
<!--caps:end-->

## Cobertura & honestidad (para la doctrina de enforcement)

- **82 capabilities** mapeadas a código real. Estado: ~24 `vivo` (test/story) · ~46 `vivo·nc`
  (sin check automático) · 3 `parcial/stub-dentro` · **6 STUB** (CAP-04 publish, CAP-05 open,
  CAP-23 watcher, CAP-22 seed-parcial, endpoint `listRuns` 501, `mint_token` sin test).
- **⚠ DRIFT CLAUDE.md ↔ código:** CLAUDE.md decía «SQLite puro-Go (modernc, WAL)». Realidad:
  índice = **map in-memory** (CAP-21), persistencia = **archivos JSON atómicos** (CAP-25).
  SQLite es fase 5 futura (comentarios en `index/store.go:1-5`, `store/registry.go:3-4`).
- **FE sin tests:** `web/` NO tiene `*.test.*`/`e2e/`; única validación = `.stories.tsx`. Los
  widgets de chrome (chat-dock, session-rail, topbar, view-strip), pages/shell y stores/api
  quedan `nc` pese a ser capabilities reales. Deuda de validación → BACKLOG.
- **Dead-code candidatos (sin capability):** `domain/graph.go#UnidadDeTrabajo:119` (sin uso) ·
  enums `Banda/Canal/Procedencia/Origen` en box.go (serializados, sin comportamiento propio) ·
  `app-store.ts#toggleTheme:25` (latente, sin control en UI). → decidir: reclamar o borrar.
- **Deuda declarada sin capability:** reconocedores pendientes del loader (`subagent`,
  plugin-raíz, edges de librería/Guardia) — `loader.go:20-28`.
- **Soporte para R2** (código que implementa un cap pero no es su puntero autoritativo):
  `_coverage.yaml`.

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
