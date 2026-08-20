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


### `cli-daemon` (7)

- **CAP-01 · Servir daemon** `vivo·nc` · cli-daemon/servir-daemon.yaml — `cmd/arnesia/main.go#runServe`
- **CAP-02 · Indexar arnés (CLI `index`)** `vivo·nc` · cli-daemon/indexar-arnes.yaml — `cmd/arnesia/main.go#runIndex`
- **CAP-03 · Conformance CLI (`elemento|--arnes|--todo`)** `vivo` · cli-daemon/conformance-cli.yaml — `cmd/arnesia/conformance.go#runConformance`
- **CAP-04 · Publicar al marketplace (`publish`)** `parcial` · cli-daemon/publicar-al-marketplace.yaml — `cmd/arnesia/main.go#runPublish`
- **CAP-05 · Abrir shell (`open`)** `stub` · cli-daemon/abrir-shell.yaml — `cmd/arnesia/main.go#runOpen`
- **CAP-117 · Log del daemon a archivo rotativo (el incidente deja rastro)** `vivo` · cli-daemon/log-del-daemon.yaml — `internal/adapters/logfile/logfile.go#Writer`
- **CAP-138 · CLI de telemetría (la superficie observable sin pantalla)** `vivo` · cli-daemon/telemetria-cli.yaml — `cmd/arnesia/telemetria.go#runTelemetria`

### `dominio-l0` (10)

- **CAP-06 · Grafo agnóstico L0 (Graph/Arnes/Edge/manifiesto)** `vivo·nc` · dominio-l0/grafo-agnostico-l0.yaml — `internal/domain/graph.go#Graph`
- **CAP-07 · Taxonomía Clase (10 primitivas)** `vivo·nc` · dominio-l0/taxonomia-clase.yaml — `internal/domain/box.go#Clase`
- **CAP-08 · Contrato fusionado de caja (4 ejes)** `vivo` · dominio-l0/contrato-fusionado-de-caja.yaml — `internal/domain/box.go#Contract`
- **CAP-09 · Ejes de autonomía (arquetipo×perfil) + document-as-cache** `vivo` · dominio-l0/ejes-de-autonomia-document-as-cache.yaml — `internal/domain/box.go#Arquetipo`
- **CAP-10 · Resolución de insumos (cableado)** `vivo` · dominio-l0/resolucion-de-insumos.yaml — `internal/domain/box.go#InsumosDe`
- **CAP-11 · FSM determinista de caja (T3)** `vivo` · dominio-l0/fsm-determinista-de-caja.yaml — `internal/domain/caja_fsm.go#AvanzarCaja`
- **CAP-12 · Spine + 5 categorías fijas (interop I-77)** `vivo` · dominio-l0/spine-5-categorias-fijas.yaml — `internal/domain/graph.go#Spine`
- **CAP-13 · Permisos derivados del rol (deny>ask>allow, TTL)** `vivo·nc` · dominio-l0/permisos-derivados-del-rol.yaml — `internal/domain/permission.go#PermissionSet.Decide`
- **CAP-14 · Sesión = frente de trabajo** `vivo` · dominio-l0/sesion-frente-de-trabajo.yaml — `internal/domain/session.go#Session`
- **CAP-140 · La conversación es una entidad, y la sesión la contiene** `vivo` · dominio-l0/conversacion-como-entidad.yaml — `internal/domain/conversacion.go#Conversacion`

### `loader` (7)

- **CAP-15 · Reconocer forma física (plugin|instalado)** `vivo` · loader/reconocer-forma-fisica.yaml — `internal/adapters/loader/loader.go#detectarElementos`
- **CAP-151 · Derivar actividades (arnes.yaml → catálogo + faceta por caja)** `vivo` · loader/derivar-actividades.yaml — `internal/adapters/loader/actividades.go#leerActividades`
- **CAP-16 · Cargar arnés a grafo L0** `vivo` · loader/cargar-arnes-a-grafo-l0.yaml — `internal/adapters/loader/loader.go#LoadArnes`
- **CAP-17 · Leer manifiesto (degradado honesto)** `vivo` · loader/leer-manifiesto.yaml — `internal/adapters/loader/loader.go#leerManifiesto`
- **CAP-18 · Reconocedores clase→ubicación (8 tipos)** `vivo` · loader/reconocedores-claseubicacion.yaml — `internal/adapters/loader/loader.go#reconocerSkills`
- **CAP-19 · Derivar edges (necesita→lee/invoca)** `vivo` · loader/derivar-edges.yaml — `internal/adapters/loader/edges.go#derivarEdges`
- **CAP-20 · Reconciliación honesta (no-reconocido visible)** `vivo` · loader/reconciliacion-honesta.yaml — `internal/adapters/loader/loader.go#nodoNoReconocido`

### `indice-persistencia` (7)

- **CAP-141 · El archivo durable declara su esquema y sabe migrarse** `vivo` · indice-persistencia/migracion-de-esquema-en-disco.yaml — `internal/adapters/store/esquema.go#EsquemaActual`
- **CAP-142 · Las sesiones con llave pelada se recalibran solas al arrancar, y nada se borra** `vivo` · indice-persistencia/recalibracion-de-llaves-de-sesion.yaml — `internal/adapters/store/rekey.go#ClaveCalificada`
- **CAP-21 · Índice de arneses en memoria** `vivo` · indice-persistencia/indice-de-arneses-en-memoria.yaml — `internal/adapters/index/store.go#Store`
- **CAP-22 · Reconstrucción del índice (ArnesRegistry)** `vivo` · indice-persistencia/reconstruccion-del-indice.yaml — `internal/adapters/index/store.go#Store.Rebuild`
- **CAP-23 · Observar cambios del corpus (watcher)** `parcial` · indice-persistencia/observar-cambios-del-corpus.yaml — `internal/adapters/watch/watcher.go#Watcher`
- **CAP-24 · Registro arnés→working-dir (confinamiento cwd + denylist)** `vivo·nc` · indice-persistencia/registro-arnesworking-dir.yaml — `internal/adapters/store/arnes_registry.go#Register`
- **CAP-25 · Persistencia (sesiones + arnés→path, JSON atómico)** `vivo` · indice-persistencia/persistencia.yaml — `internal/adapters/store/registry.go#Save`

### `conformance` (7)

- **CAP-26 · Parsear ruleset a data (knowledge+arch→checks)** `vivo` · conformance/parsear-ruleset-a-data.yaml — `internal/adapters/conformance/ruleset/parser.go#Loader.Load`
- **CAP-27 · Inferir mecanismo + wiring por check** `vivo` · conformance/inferir-mecanismo-wiring-por-check.yaml — `internal/adapters/conformance/ruleset/parser.go#inferMechanism`
- **CAP-28 · Ejecutar checks por mecanismo (arch-test/go-arch-lint/schema/static/nl-judge)** `vivo` · conformance/ejecutar-checks-por-mecanismo.yaml — `internal/adapters/conformance/mechanism/adapters.go#ArchTest.Run`
- **CAP-29 · Built-ins de conformidad del dominio (spine/escritor-único/composición/veredicto)** `vivo` · conformance/built-ins-de-conformidad-del-dominio.yaml — `internal/domain/conformance.go#VerificarSpine`
- **CAP-30 · Servicio de conformidad (rutea target + firewall CC-native)** `vivo` · conformance/servicio-de-conformidad.yaml — `internal/usecase/conformance_service.go#RunGraph`
- **CAP-31 · Scope fabrica|arnes (portabilidad)** `vivo·nc` · conformance/scope-fabrica-arnes.yaml — `internal/adapters/conformance/mechanism/adapters.go`
- **CAP-32 · Doctrina embebida (portable)** `vivo·nc` · conformance/doctrina-embebida.yaml — `embed_doctrina.go#Files`

### `conductor` (12)

- **CAP-135 · Toda sesión que ArnesIA lanza nace instrumentada** `vivo` · conductor/spawn-inyecta-telemetria.yaml — `internal/adapters/agent/claudecode/conductor.go#SpawnEnv`
- **CAP-136 · El uso del turno que el cierre del subproceso ya traía y se tiraba** `vivo` · conductor/uso-del-turno-en-el-result.yaml — `internal/adapters/agent/claudecode/conductor.go#parseResult`
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

### `provisioning` (3)

- **CAP-43 · Materializar doctrina+kit a ~/.arnesia (idempotente por huella)** `vivo` · provisioning/materializar-doctrina-kit-a-arnesia.yaml — `internal/adapters/provision/provisioner.go#Provision`
- **CAP-44 · Resolver permission-set por rol (deny-by-default)** `vivo·nc` · provisioning/resolver-permission-set-por-rol.yaml — `internal/adapters/permission/provisioner.go#ResolveForRole`
- **CAP-96 · Tarjeta de identidad por sesión (grounding del chat)** `vivo` · provisioning/tarjeta-identidad-por-sesion.yaml — `internal/usecase/session_grounding.go`

### `handoff` (3)

- **CAP-45 · Leer status del artefacto (señal máquina)** `vivo` · handoff/leer-status-del-artefacto.yaml — `internal/adapters/artifact/reader.go#Status`
- **CAP-46 · Digest acotado (economía de contexto, −90%)** `vivo` · handoff/digest-acotado.yaml — `internal/adapters/artifact/reader.go#Resumen`
- **CAP-47 · Confinamiento de path del artefacto** `vivo` · handoff/confinamiento-de-path-del-artefacto.yaml — `internal/adapters/artifact/reader.go#confinedPath`

### `http-sse` (10)

- **CAP-106 · Superficie REST de marketplaces (7 endpoints, 400 ≠ 503 ≠ 409, `entradas: null`)** `parcial` · http-sse/superficie-marketplaces.yaml — `internal/adapters/transport/http/router.go#NewHandler`
- **CAP-114 · Superficie HTTP del dictado (subir audio · disponibilidad)** `vivo` · http-sse/superficie-dictado.yaml — `internal/adapters/transport/http/dictado.go#postDictado`
- **CAP-116 · Diagnóstico de fallos del FE (el WebView deja rastro en el log)** `vivo` · http-sse/diagnostico-de-fallos.yaml — `internal/adapters/transport/http/diagnostico.go#postDiagnostico`
- **CAP-137 · Superficie HTTP de telemetría (donde `null` y `0` no se confunden)** `vivo` · http-sse/superficie-telemetria.yaml — `internal/adapters/transport/http/telemetria.go#getResumen`
- **CAP-144 · El panel de conversaciones se entera en vivo** `vivo` · http-sse/frame-de-conversacion.yaml — `internal/usecase/session_service.go#dockFrame`
- **CAP-48 · Confinamiento de superficie local (Host+Origin+token)** `vivo` · http-sse/confinamiento-de-superficie-local.yaml — `internal/adapters/transport/http/auth.go#withAuth`
- **CAP-49 · Router + montaje** `vivo·nc` · http-sse/router-montaje.yaml — `internal/adapters/transport/http/router.go#NewHandler`
- **CAP-50 · UI embebida servida por daemon** `vivo·nc` · http-sse/ui-embebida-servida-por-daemon.yaml — `embed_webdist.go#WebDist`
- **CAP-51 · SSE multiplexado (map/dock/run, replay)** `vivo·nc` · http-sse/sse-multiplexado.yaml — `internal/adapters/transport/sse/broker.go#Publish`
- **CAP-52 · Superficie REST (58 rutas servidas · 43 declaradas · 14 exentas con razón)** `vivo` · http-sse/superficie-rest.yaml — `internal/adapters/transport/http/router.go#NewHandler`

### `usecases` (15)

- **CAP-113 · Dictado: transcribir y ordenar con contexto de dominio** `vivo` · usecases/dictado-transcribir-y-ordenar.yaml — `internal/ports/dictado.go#TranscriptionPort`
- **CAP-143 · Crear, retomar y renombrar conversaciones de una sesión** `vivo` · usecases/conversaciones-de-una-sesion.yaml — `internal/usecase/session_conversaciones.go#CrearConversacion`
- **CAP-145 · Buscar por lo que se dijo, no por cómo se llama** `vivo` · usecases/buscar-en-el-transcript.yaml — `internal/usecase/session_conversaciones.go#Conversaciones`
- **CAP-147 · El título de la conversación se deriva de su primer mensaje y su fecha se estampa en cada turno** `vivo` · usecases/titulo-y-fecha-de-la-conversacion.yaml — `internal/usecase/session_service.go#SessionService.Turn`
- **CAP-53 · Dock: conversación en vivo multisesión** `vivo` · usecases/dock-conversacion-en-vivo-multisesion.yaml — `internal/usecase/session_service.go#Turn`
- **CAP-54 · Gobierno del turno (permisos HITL + interrupt)** `vivo` · usecases/gobierno-del-turno.yaml — `internal/usecase/session_service.go#onControlRequest`
- **CAP-55 · Ejecutar caja T3 desde daemon (async + gate post-run)** `vivo·nc` · usecases/ejecutar-caja-t3-desde-daemon.yaml — `internal/usecase/run_service.go#RunService.StartRun`
- **CAP-56 · Orquestación determinista (BoxConductor)** `vivo·nc` · usecases/orquestacion-determinista.yaml — `internal/usecase/box_conductor.go#RunWith`
- **CAP-57 · Servir fuente real de nodo** `vivo` · usecases/servir-fuente-real-de-nodo.yaml — `internal/usecase/fuente_service.go#Fuente`
- **CAP-58 · Mapa/portafolio/inspector (ensamblado)** `vivo` · usecases/mapa-portafolio-inspector.yaml — `internal/usecase/map_service.go#Graph`
- **CAP-59 · Gestión de sesiones CRUD** `vivo` · usecases/gestion-de-sesiones-crud.yaml — `internal/usecase/session_service.go#Create`
- **CAP-94 · Reindex del Mapa tras cada turno del chat** `vivo` · usecases/reindex-tras-turno.yaml — `internal/usecase/session_reindex.go`
- **CAP-97 · Rotación de contexto invisible (conversación infinita)** `vivo` · usecases/rotacion-de-contexto.yaml — `internal/usecase/session_rotacion.go`
- **CAP-98 · Historial de conversaciones archivadas (transcript propio + JSONL de fallback)** `vivo` · usecases/historial-de-conversaciones.yaml — `internal/usecase/session_historial.go`
- **CAP-99 · Paquete cerrado: el gate deniega el árbol propio de arnesia** `vivo` · usecases/paquete-cerrado-en-el-gate.yaml — `internal/usecase/session_service.go#ProtegerPaqueteCerrado`

### `self-update` (1)

- **CAP-60 · Self-update sin sudo (5 pasos atómico)** `vivo` · self-update/self-update-sin-sudo.yaml — `internal/usecase/selfupdate_service.go#Actualizar`

### `portafolio` (15)

- **CAP-102 · Registro de marketplaces conocidos (collect-all CC + declarado)** `parcial` · portafolio/registrar-marketplace.yaml — `internal/domain/repo_ref.go#CanonicalizarRepo`
- **CAP-103 · Leer el catálogo de un marketplace (local primero, remoto por gh, caché fechado)** `parcial` · portafolio/leer-catalogo-marketplace.yaml — `internal/domain/repo_ref.go#CanonicalizarRepo`
- **CAP-104 · Situación de una fila del catálogo (6 ramas, precedencia, acción por clase)** `parcial` · portafolio/situacion-de-catalogo.yaml — `internal/domain/portafolio.go#EntradaPortafolio`
- **CAP-105 · Asignar origen a un arnés huérfano (re-key por home declarado, sin clonar)** `parcial` · portafolio/asignar-origen.yaml — `internal/usecase/portafolio.go#PortafolioService`
- **CAP-111 · Traer canónico (materializar el arnés del estante: local o clone externo, atómico)** `parcial` · portafolio/traer-canonico.yaml — `internal/domain/portafolio.go#Canonico`
- **CAP-148 · Publicar (write-side prenter-marketplace: gate de conformance → copia versionada → push sin force → tag)** `parcial` · portafolio/publicar.yaml — `internal/domain/publicar.go#SolicitudPublicacion`
- **CAP-154 · Sincronizar el checkout de un marketplace propio al Refrescar (pull ff-only)** `parcial` · portafolio/sincronizar-catalogo.yaml — `internal/ports/marketplace.go#CatalogoSync`
- **CAP-155 · Adoptar como canónico — el camino a canónico de un plugin forjado local** `parcial` · portafolio/adoptar-como-canonico.yaml — `internal/usecase/adoptar.go#Adoptar`
- **CAP-83 · Registrar identidad de arnés (home,id)** `vivo` · portafolio/registrar-identidad.yaml — `internal/domain/portafolio.go#IdentidadArnes`
- **CAP-84 · Escanear proyecto (walker multi-instalación)** `vivo` · portafolio/escanear-proyecto.yaml — `internal/domain/portafolio.go#TipoInstalacion`
- **CAP-85 · Resolver origen (collect-all + reconcilia)** `vivo` · portafolio/resolver-origen.yaml — `internal/domain/portafolio.go#OrigenPortafolio`
- **CAP-86 · Evaluar deriva (hash vs referencia inmutable)** `vivo` · portafolio/evaluar-deriva.yaml — `internal/domain/portafolio.go#EstadoDeriva`
- **CAP-87 · Desvincular arnés del Portafolio** `vivo` · portafolio/desvincular.yaml — `internal/domain/portafolio.go#EntradaPortafolio`
- **CAP-88 · Observar en Mapa (presencia read-only del Portafolio)** `vivo` · portafolio/observar-en-mapa.yaml — `internal/usecase/portafolio.go#PortafolioService.ObservarEnMapa`
- **CAP-93 · Identificar (sellar arnes.l0.json in-situ)** `vivo` · portafolio/identificar.yaml — `internal/usecase/portafolio.go#PortafolioService.Identificar`

### `fe-mapa` (11)

- **CAP-139 · Capa «Mejora» del Mapa — cifra, confianza, cobertura y puntos de mejora** `parcial` · fe-mapa/capa-mejora.yaml — `web/src/entities/telemetria/model/types.ts#CifraCaja`
- **CAP-152 · Chips de actividad (N0) — panorama multi-actividad del Mapa** `vivo` · fe-mapa/chips-de-actividad.yaml — `web/src/widgets/map-canvas/ui/actividad-chips.tsx#ActividadChips`
- **CAP-153 · Foco de actividad (N1) — secuencia del procedimiento, atenuación y radio de impacto** `vivo` · fe-mapa/foco-de-actividad.yaml — `web/src/pages/shell/ui/workspace-stage.tsx#WorkspaceStage`
- **CAP-61 · Renderizar el Mapa (HTML+SVG)** `vivo` · fe-mapa/renderizar-el-mapa.yaml — `web/src/widgets/map-canvas/ui/map-canvas.tsx#MapCanvas`
- **CAP-62 · Pan/zoom/fit** `vivo·nc` · fe-mapa/pan-zoom-fit.yaml — `web/src/widgets/map-canvas/model/use-viewport.ts#useViewport`
- **CAP-63 · Inspector drawer (Resumen|Contenido|Corridas)** `vivo` · fe-mapa/inspector-drawer.yaml — `web/src/widgets/map-canvas/ui/inspector.tsx#Inspector`
- **CAP-64 · Tab Contenido (fuente real, lazy)** `vivo` · fe-mapa/tab-contenido.yaml — `web/src/widgets/map-canvas/ui/inspector.tsx#Contenido`
- **CAP-65 · Franja de artefactos (chips hand-off, off/auto/todos)** `vivo` · fe-mapa/franja-de-artefactos.yaml — `web/src/widgets/map-canvas/ui/handoff-gutter.tsx#HandoffGutter`
- **CAP-66 · Picker de arnés** `vivo` · fe-mapa/picker-de-arnes.yaml — `web/src/widgets/map-canvas/ui/map-bar.tsx#arnes`
- **CAP-67 · Conmutador de capas** `vivo` · fe-mapa/conmutador-de-capas.yaml — `web/src/widgets/map-canvas/ui/map-bar.tsx#LAYERS`
- **CAP-95 · Mapa en vivo: refetch al reindexarse el arnés por chat** `vivo` · fe-mapa/reindex-en-vivo.yaml — `web/src/shared/store/map-live-store.ts`

### `fe-chat` (7)

- **CAP-100 · Conversación legible: actividad visible · burbuja por paso · markdown** `vivo` · fe-chat/conversacion-legible.yaml — `internal/adapters/agent/claudecode/conductor.go#assistantEvents`
- **CAP-115 · Dictar en el composer (grabar → transcribir → ordenar → poblar)** `vivo` · fe-chat/dictar-en-el-composer.yaml — `web/src/widgets/chat-dock/ui/dictado-button.tsx#DictadoButton`
- **CAP-146 · Panel de conversaciones: listarlas, buscarlas, crearlas y retomarlas desde el dock** `vivo` · fe-chat/panel-de-conversaciones.yaml — `web/src/widgets/chat-dock/ui/conversacion-row.tsx#ConversacionRow`
- **CAP-68 · Chat CC (turno/interrupt)** `vivo` · fe-chat/chat-cc.yaml — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ChatDock`
- **CAP-69 · Acotar alcance (nodo→chip)** `vivo` · fe-chat/acotar-alcance.yaml — `web/src/widgets/chat-dock/ui/chat-dock.tsx#ScopeRow`
- **CAP-70 · Decidir permisos (tarjeta inline)** `vivo` · fe-chat/decidir-permisos.yaml — `web/src/widgets/chat-dock/ui/permission-card.tsx#PermissionCard`
- **CAP-71 · Gate de conformance tras escrituras (RF-117)** `vivo·nc` · fe-chat/gate-de-conformance-tras-escrituras.yaml — `web/src/shared/store/sessions-store.ts`

### `fe-shell` (7)

- **CAP-72 · Rail de sesiones (crear/renombrar/cerrar/switch)** `vivo` · fe-shell/rail-de-sesiones.yaml — `web/src/widgets/session-rail/ui/session-rail.tsx#SessionRail`
- **CAP-73 · View-strip (Mapa|Diag|Corridas|Tren|Hist)** `vivo·nc` · fe-shell/view-strip.yaml — `web/src/widgets/view-strip/ui/view-strip.tsx#ViewStrip`
- **CAP-74 · Topbar breadcrumb + ⌘K dock** `vivo·nc` · fe-shell/topbar-breadcrumb-k-dock.yaml — `web/src/widgets/topbar/ui/topbar.tsx#Topbar`
- **CAP-75 · Navegación global (portafolio/estándar/ajustes)** `vivo·nc` · fe-shell/navegacion-global.yaml — `web/src/pages/shell/ui/global-view.tsx#GlobalView`
- **CAP-76 · Shell composition-root + hash-state routing** `vivo·nc` · fe-shell/shell-composition-root-hash-state-routing.yaml — `web/src/pages/shell/ui/shell-page.tsx#ShellPage`
- **CAP-77 · Botón actualizar (UI self-update)** `vivo` · fe-shell/boton-actualizar.yaml — `web/src/features/self-update/ui/update-card.tsx#UpdateCard`
- **CAP-78 · Transporte FE (REST+SSE+token)** `vivo·nc` · fe-shell/transporte-fe.yaml — `web/src/shared/api/client.ts#api`

### `fe-portafolio` (8)

- **CAP-107 · Plano Marketplaces (plano hermano de Arneses, filas honestas, contador cruzado)** `parcial` · fe-portafolio/plano-marketplaces.yaml — `web/src/pages/shell/ui/portafolio-view.tsx#PortafolioView`
- **CAP-108 · Catálogo de un marketplace (situación por fila, buscador+filtro+lazy, referencia read-only)** `parcial` · fe-portafolio/catalogo-de-marketplace.yaml — `web/src/pages/shell/ui/portafolio-view.tsx#PortafolioView`
- **CAP-109 · Wizard rama Marketplace (registrar: validar real → clase → aterrizar en el catálogo)** `parcial` · fe-portafolio/wizard-registrar-marketplace.yaml — `web/src/widgets/portafolio/ui/portafolio-wizard.tsx#PortafolioWizard`
- **CAP-110 · Diálogo Resolver origen (in-situ sobre la fila, nada premarcado, «ninguno» explícito)** `parcial` · fe-portafolio/resolver-origen-dialogo.yaml — `web/src/widgets/portafolio/ui/portafolio-drawer.tsx#PortafolioDrawer`
- **CAP-89 · Lista del Portafolio (lentes empresa/plano/proyecto/marketplace, buscar, corruptas visibles)** `vivo` · fe-portafolio/lista-del-portafolio.yaml — `web/src/entities/portafolio/model/types.ts#EntradaPortafolio`
- **CAP-90 · Drawer de detalle READ (identidad, facetas, origen de la copia, instalaciones)** `vivo` · fe-portafolio/drawer-detalle-read.yaml — `web/src/entities/portafolio/model/types.ts#Instalacion`
- **CAP-91 · Wizard agregar proyecto (escanear carpeta local → candidatos honestos → elegir)** `vivo` · fe-portafolio/wizard-agregar-proyecto.yaml — `web/src/entities/portafolio/model/types.ts#Candidato`
- **CAP-92 · Abrir en Mapa desde el Portafolio (peek al stage de sesión)** `vivo` · fe-portafolio/abrir-en-mapa.yaml — `web/src/shared/api/client.ts#observarEnMapa`

### `tauri` (6)

- **CAP-101 · Deep-link arnesia:// enfoca la app instalada** `vivo·nc` · tauri/deep-link-enfoca-app.yaml — `web/src-tauri/src/lib.rs`
- **CAP-112 · El shell concede el permiso de micrófono del WebView** `vivo` · tauri/permiso-de-microfono.yaml — `web/src-tauri/src/lib.rs#conceder_permiso_de_microfono`
- **CAP-79 · Single-instance reenfoca** `vivo·nc` · tauri/single-instance-reenfoca.yaml — `web/src-tauri/src/lib.rs`
- **CAP-80 · Sidecar del daemon (bind-or-bail verificado, kill al salir)** `vivo` · tauri/sidecar-del-daemon.yaml — `web/src-tauri/src/lib.rs#estado_daemon`
- **CAP-81 · Inyección de token (raíz de confianza)** `vivo·nc` · tauri/inyeccion-de-token.yaml — `web/src-tauri/src/lib.rs#mint_token`
- **CAP-82 · Workaround render Linux (WEBKIT_DISABLE_DMABUF)** `vivo·nc` · tauri/workaround-render-linux.yaml — `web/src-tauri/src/main.rs`

### `forja` (2)

- **CAP-149 · Sembrar semilla .arnesia/ (arnesia init)** `vivo` · forja/sembrar-semilla.yaml — `internal/adapters/forja/parser.go#ParseSemilla`
- **CAP-150 · Chequear salud de la semilla (doctor v0, arnesia init --check)** `vivo` · forja/chequear-semilla.yaml — `internal/adapters/forja/doctor.go#Adapter.Chequear`

### `telemetria` (19)

- **CAP-118 · Receptor OTLP embebido (nunca bloquea al emisor, nunca finge haber guardado)** `vivo` · telemetria/receptor-otlp-embebido.yaml — `internal/adapters/telemetria/otlp/receptor.go#Receptor`
- **CAP-119 · Decodificador OTLP/JSON con la stdlib (+0,49 MB, no +10,79 MB)** `vivo` · telemetria/decodificador-otlp-json.yaml — `internal/adapters/telemetria/otlp/decodifica.go#DecodificarLogs`
- **CAP-120 · Evento canónico de telemetría (una puerta de escritura, «no aplica» ≠ 0)** `vivo` · telemetria/evento-canonico-de-telemetria.yaml — `internal/domain/telemetria.go#EventoTelemetria`
- **CAP-121 · Allowlist de ingesta (ni identidad de cuenta, ni contenido, ni rutas)** `vivo` · telemetria/allowlist-de-ingesta.yaml — `internal/adapters/telemetria/otlp/mapa_cc.go#MapearLogRecord`
- **CAP-122 · Almacén de telemetría (base propia, migración aditiva, archivado que nunca borra)** `vivo` · telemetria/almacen-de-telemetria.yaml — `internal/adapters/telemetria/store/store.go#Store`
- **CAP-123 · Rollup horario incremental (el «no aplica» sobrevive a la suma)** `parcial` · telemetria/rollup-horario-incremental.yaml — `internal/adapters/telemetria/store/rollup.go#Rollup`
- **CAP-124 · Retención y borrado (el botón borra también el agregado)** `vivo` · telemetria/retencion-y-borrado.yaml — `internal/usecase/telemetria_retencion.go#TelemetriaService.Purgar`
- **CAP-125 · Catálogo de precios embebido (cero post-install, con procedencia declarada)** `vivo` · telemetria/catalogo-de-precios-embebido.yaml — `internal/adapters/telemetria/catalogo/catalogo.go#Embebido`
- **CAP-126 · Costeo de doble fuente (lo que dijo el runtime y lo que dice nuestro catálogo)** `vivo` · telemetria/costeo-de-doble-fuente.yaml — `internal/domain/telemetria_costo.go#CalcularCosto`
- **CAP-127 · Atribución y confianza (el orden de preferencia, y el escenario derivado de la señal)** `vivo` · telemetria/atribucion-y-confianza.yaml — `internal/usecase/telemetria_service.go#TelemetriaService.Atribuir`
- **CAP-128 · Conciliación de cobertura (cuántos turnos hubo, no cuántos medimos)** `vivo` · telemetria/conciliacion-de-cobertura.yaml — `internal/usecase/telemetria_service.go#TelemetriaService.Conciliar`
- **CAP-129 · Motor de detectores (cada uno declara si puede correr, y por qué no)** `vivo` · telemetria/motor-de-detectores.yaml — `internal/domain/telemetria_deteccion.go#Detector`
- **CAP-130 · Los seis detectores del MVP (número, contrafactual, umbral, sesgo en contra y un fix)** `vivo` · telemetria/los-seis-detectores-del-mvp.yaml — `internal/domain/telemetria_deteccion.go#DetectoresMVP`
- **CAP-131 · Hook de proceso que nunca rompe un turno (el hook es el propio binario)** `vivo` · telemetria/hook-de-proceso-fail-open.yaml — `cmd/arnesia/hook.go#runHook`
- **CAP-132 · Ficha de descubrimiento del daemon (publicada solo cuando ya escucha)** `vivo` · telemetria/ficha-de-descubrimiento-del-daemon.yaml — `internal/adapters/telemetria/descubrimiento/ficha.go#Ficha`
- **CAP-133 · Token de ingesta acotado (escribir telemetría no es conducir un agente)** `vivo` · telemetria/token-de-ingesta-acotado.yaml — `internal/adapters/transport/http/auth.go#validaIngesta`
- **CAP-134 · Forward externo filtrado (apagado por default, y solo el operador lo enciende)** `vivo` · telemetria/forward-externo-filtrado.yaml — `internal/adapters/telemetria/forward/forward.go#Forward`
- **CAP-141 · La frase del punto de mejora se arma en el dominio (y el dinero se formatea una sola vez)** `vivo` · telemetria/prosa-del-punto-de-mejora.yaml — `internal/domain/telemetria_prosa.go#PuntoDeMejora.Redactar`
- **CAP-142 · El arnés lleva su propia instrumentación (y no escribe en el árbol de nadie)** `vivo` · telemetria/bloque-env-en-el-paquete-del-arnes.yaml — `internal/adapters/agent/claudecode/conductor.go#VariablesTelemetria`
<!--caps:end-->

## Cobertura & honestidad (para la doctrina de enforcement)

- **92 capabilities** mapeadas a código real (82 de HS-18 + 6 `portafolio/` + 4 `fe-portafolio/`
  del programa Portafolio, HS-22/23). La distribución por estado es cifra VIVA — se genera con
  `estado.sh` en `docs/product/checkpoint.md`, acá no se teclea.
- **⚠ DRIFT CLAUDE.md ↔ código:** CLAUDE.md decía «SQLite puro-Go (modernc, WAL)». Realidad:
  índice = **map in-memory** (CAP-21), persistencia = **archivos JSON atómicos** (CAP-25).
  SQLite es fase 5 futura (comentarios en `index/store.go:1-5`, `store/registry.go:3-4`).
- **FE con tests unitarios desde Slice 1 Portafolio (S1-D7):** existe el proyecto vitest `unit`
  (node) — `web/src/entities/portafolio/model/selectors.test.ts` (26 tests) corre en CI job `ts`.
  El resto del FE sigue validándose solo por `.stories.tsx`: los widgets de chrome (chat-dock,
  session-rail, topbar, view-strip), pages/shell y stores/api quedan `nc` pese a ser capabilities
  reales. Deuda de validación → BACKLOG.
- **Dead-code candidatos (sin capability):** `domain/graph.go#UnidadDeTrabajo:119` (sin uso) ·
  enums `Banda/Canal/Procedencia/Origen` en box.go (serializados, sin comportamiento propio) ·
  `app-store.ts#toggleTheme:25` (latente, sin control en UI). → decidir: reclamar o borrar.
- **Deuda declarada sin capability:** reconocedores pendientes del loader (`subagent`,
  plugin-raíz, edges de librería/Guardia) — `loader.go:20-28`.
- **Soporte para R2** (código que implementa un cap pero no es su puntero autoritativo):
  `_coverage.yaml`.

## Apéndice — 28 endpoints REST (`transport/http/router.go`)

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
GET  /api/portafolio                           listado + corruptas        CAP-83
POST /api/portafolio/escaneos                  escanear (no persiste)     CAP-84
POST /api/portafolio/proyectos                 agregar elegidos           CAP-84
DELETE /api/portafolio/arneses/{clave}         desvincular                CAP-87
POST /api/portafolio/arneses/{clave}/mapa      observar en Mapa           CAP-88
```
