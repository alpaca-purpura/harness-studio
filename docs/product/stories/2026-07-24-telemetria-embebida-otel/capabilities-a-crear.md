# Capabilities a crear — módulo `telemetria/` y vecinos

> **Qué es esto:** la lista de las hojas YAML que este paquete tiene que construir en
> `docs/product/capabilities/`, con su slug, qué hace, el **símbolo que la va a satisfacer** y el
> **check** que sostendrá su estado.
>
> **NO se crean todavía, y es a propósito:** R1 (`cap-ptr-resuelve` + `cap-ptr-simbolo-resuelve`)
> exige que cada puntero resuelva a un archivo **y a un símbolo real** del árbol. El código no
> existe ⇒ crear las hojas hoy rompería el enforcer con 22 punteros colgantes. Se crean **en el
> mismo commit que crea cada símbolo** (R3, gate de commit).
>
> Doctrina: [`codigo-traza-a-capability.md`](../../../architecture/boundaries/codigo-traza-a-capability.md).
> Formato: [`docs/product/_templates/capability.template.yaml`](../../_templates/capability.template.yaml).
> Diseño de referencia: [`arquitectura-modulo.md`](arquitectura-modulo.md).

---

## 0 · Dos prerrequisitos que no son capabilities

1. **`telemetria` hay que agregarlo a `domain_modules` en `project.config.yaml`.** Hoy la lista
   tiene 16 módulos y no lo incluye.
   ⚠️ **Drift ya existente, anotado no resuelto:** varios módulos que YA tienen hojas
   (`http-sse`, `cli-daemon`, `fe-mapa`, `fe-portafolio`, `fe-shell`, `handoff`, `tauri`,
   `usecases`, `dominio-l0`, `indice-persistencia`) **tampoco están** en `domain_modules`, pese a
   que el template dice «uno de `project.config.yaml domain_modules`». O el template miente o la
   lista está incompleta; no lo resuelve este paquete.
2. **Numeración.** El último `cap_num` del árbol es **CAP-117**. Las hojas de este paquete arrancan
   en **CAP-118** y son **22**, o sea CAP-118…CAP-139.

---

## 1 · Módulo nuevo `telemetria/` (17 hojas)

| cap | slug | qué hace | puntero (símbolo que la satisface) | check (`valida:`) |
|---|---|---|---|---|
| **CAP-118** | `receptor-otlp-embebido` | Recibe OTLP/JSON en `/v1/logs` y `/v1/metrics` sobre el mux que ya existe, loopback-only, sin bloquear jamás al emisor: cola acotada, descarte contado y `partialSuccess` en la respuesta | `internal/adapters/telemetria/otlp/receptor.go#Receptor`<br>`…/receptor.go#NewReceptor`<br>`internal/adapters/transport/http/router.go#NewHandler` | `TestReceptorNoBloqueaAlEmisor` · `TestOTLPBajoLosTresGates` · `TestProtobufSeRechazaConMotivo` · `TestPayloadGiganteSeRechazaSinLeerlo` |
| **CAP-119** | `decodificador-otlp-json` | Decodifica el subconjunto del wire OTLP/JSON que usamos, con `json.Number` para el `intValue` off-spec, ignorando lo desconocido y rechazando el lote malo sin tumbar el daemon | `internal/adapters/telemetria/otlp/decodifica.go#DecodificarLogs`<br>`…/decodifica.go#DecodificarMetricas` | `TestIntValueComoNumeroYComoString` · `TestCamposDesconocidosNoRompen` · `TestPayloadMalformado` · `TestCumulativeNoSeAdivina` · `FuzzDecodificarLogs` |
| **CAP-120** | `evento-canonico-de-telemetria` | El único evento que los cuatro emisores producen, con la llave del join `(sesion_id, turno_id)`, los buckets como punteros («no aplica» ≠ 0) y los dos costos | `internal/domain/telemetria.go#EventoTelemetria`<br>`internal/domain/telemetria.go#LlaveJoin`<br>`internal/domain/telemetria.go#Tokens`<br>`internal/ports/telemetria.go#TelemetriaSink` | `TestNoAplicaNoEsCeroEnElWire` · `TestEventoSinSesionSeRechaza` · `TestEventoSinTurnoNoEntraAlJoin` |
| **CAP-121** | `allowlist-de-ingesta` | Persiste solo los campos declarados, en los DOS caminos (OTLP y hook) y dos veces en S2: ni identidad de cuenta, ni contenido de conversación, ni rutas del usuario | `internal/adapters/telemetria/otlp/mapa_cc.go#MapearLogRecord`<br>`internal/adapters/telemetria/hooks/proyecta.go#Proyectar` | `TestAllowlistNoPersistePII` · `TestHookNoReenviaContenido` · `TestAllowlistEsListaNoSugerencia` · `TestCwdDesconocidoNoSeGuardaCrudo` · `TestToolResultBytesSePersiste` |
| **CAP-122** | `almacen-de-telemetria` | SQLite propio (`telemetria.db`, **jamás** dentro de `index.db`), writer serializado bajo WAL, migración **aditiva** y archivado —nunca borrado— ante una ruptura de generación | `internal/adapters/telemetria/store/store.go#Store`<br>`…/store.go#New`<br>`…/migracion.go#Migracion`<br>`…/migracion.go#Aplicar` | `TestTelemetriaDBNoEsElIndice` · `TestMigracionesSonAditivas` · `TestGeneracionArchivaNoBorra` · `TestDBCorruptaSeArchiva` · `TestDosEscritoresSobreElMismoDB` |
| **CAP-123** | `rollup-horario-incremental` | Proyección por hora reanudable por cursor que hace que el tablero cueste ms en vez de cientos de ms, **preservando el `NULL`** a través de la agregación | `internal/adapters/telemetria/store/rollup.go#Actualizar`<br>`…/rollup.go#Recomputar` | `TestNoAplicaSobreviveAlRollup` · `TestRollupReanudaDesdeElCursor` · `TestRollupColapsaCardinalidad` · `TestTurnoCruzaHora` |
| **CAP-124** | `retencion-y-borrado` | TTL por default con purga periódica, y el botón «borrar la telemetría de este arnés» que limpia crudo **y** rollup en una transacción | `internal/usecase/telemetria_retencion.go#Purgar`<br>`…#BorrarArnes` | `TestPurgaRespetaTTL` · `TestBorradoPorArnesTambienLimpiaRollup` · `TestAvisoDeTamano` |
| **CAP-125** | `catalogo-de-precios-embebido` | Catálogo de tarifas compilado en el binario (cero post-install), con versión, rev y sha256 declarados, refresco **apagado por default** y degradación que se dice en pantalla | `internal/adapters/telemetria/catalogo/catalogo.go#Embebido`<br>`…/catalogo.go#Catalogo.Precio`<br>`…/catalogo.go#Catalogo.Version` | `TestRefrescoFallidoNoRompe` · `TestModeloDesconocido` · `TestTarifaParcial` · `TestAliasBedrockVertex` |
| **CAP-126** | `costeo-de-doble-fuente` | Cotiza el uso con nuestro catálogo **y conserva lo que reportó el runtime**: los dos costos se persisten y la divergencia es visible. Incluye los tres tests que los bugs ajenos dictaron | `internal/domain/telemetria_costo.go#CalcularCosto`<br>`…#PrecioModelo`<br>`…#CostoCalculado` | `TestCosteoCobraElCacheWrite` · `TestCosteoNoSumaBucketsQueSeSolapan` · `TestCosteoNoAplanaLosTiers` · `TestDobleCostoSePersisteEntero` · `TestDivergenciaDeCostoEsVisible` · `TestSinNingunCosto` |
| **CAP-127** | `atribucion-y-confianza` | Resuelve a qué unidad de trabajo pertenece cada evento (`exacta`→`por-hash`→`por-proceso`→`sin-dato`), **deriva el escenario de la señal** y garantiza que lo no atribuido no sume al total | `internal/usecase/telemetria_service.go#TelemetriaService.Atribuir`<br>`internal/domain/telemetria.go#Confianza`<br>`internal/domain/telemetria.go#Escenario` | `TestAtribucionPorHash` · `TestAtribucionPorProceso` · `TestSinDatoNoSumaAlTotal` · `TestConfianzaDeAgregadoEsLaMinima` · `TestEscenarioSeDerivaDeLaSenal` |
| **CAP-128** | `conciliacion-de-cobertura` | Cuenta los turnos que **ocurrieron** contra los que se **midieron**, para que un agujero de red (WSL2, contenedor, telemetría apagada) se vea como «N turnos no reportaron» y no como un cero | `internal/usecase/telemetria_service.go#TelemetriaService.Conciliar`<br>`internal/domain/telemetria_vistas.go#Cobertura` | `TestConciliacionCuentaLosNoLlegados` · `TestJoinPorSesionYTurno` · `TestTurnoSoloConProceso` · `TestTurnoSoloConDinero` |
| **CAP-129** | `motor-de-detectores` | El contrato común: cada detector **declara si aplica** según la señal real de la ventana, y el «no aplica» viaja al wire con motivo obligatorio junto a los «no medidos todavía» | `internal/domain/telemetria_deteccion.go#Detector`<br>`…#ContextoDeteccion`<br>`…#Aplicabilidad`<br>`…#PuntoDeMejora` | `TestDetectorQueNoAplicaTraeMotivo` · `TestNoMedidosSeDeclaran` |
| **CAP-130** | `los-seis-detectores-del-mvp` | B4 · P1 · B2 · B6 · B3 · B1: cada uno con su contrafactual, su umbral algebraico citado, su sesgo declarado en contra y **un** fix | `internal/domain/telemetria_deteccion.go#DetectoresMVP`<br>`internal/usecase/telemetria_mejoras.go#TelemetriaService.Mejoras` | `TestB1BreakEvenTTL` · `TestB2CostoDeLaRotacion` · `TestB3CambioDeModelo` · `TestB6SesionAbandonada` · `TestP1CajaQueSeRechaza` · `TestP1ParcialSinGate` · `TestS2DegradadoApagaLosDetectoresDeDinero` · `TestS2InstrumentadoTieneDineroYNoTieneSplit` |
| **CAP-131** | `hook-de-proceso-fail-open` | El hook que viaja dentro del arnés **es el propio binario**: lee el payload por stdin, proyecta a los campos declarados, postea a loopback y **sale 0 siempre**, sin stdout, con tope de 250 ms | `cmd/arnesia/hook.go#runHookProceso`<br>`internal/adapters/telemetria/hooks/proyecta.go#Proyectar` | `TestHookFailOpenSinDaemon` · `TestHookNoTardaNiFalla` · `TestHookStdoutVacio` · `TestHookFichaHuerfana` · `TestHookNoReenviaContenido` |
| **CAP-132** | `ficha-de-descubrimiento-del-daemon` | Publica dónde está el daemon en la ruta de config del SO (`os.UserConfigDir()`), `0600`, **solo después de que el listener acepta**, y la retira al salir | `internal/adapters/telemetria/descubrimiento/ficha.go#Publicar`<br>`…/ficha.go#Leer`<br>`…/ficha.go#Retirar`<br>`internal/domain/telemetria.go#FichaDaemon` | `TestFichaSoloTrasEscuchar` · `TestFichaPermisos0600` · `TestFichaNoEscribibleDegradaHonesto` · `TestFichaSeRelePorInvocacion` |
| **CAP-133** | `token-de-ingesta-acotado` | Un segundo token, separado del de la API: abre **solo** los tres endpoints de ingesta. Filtrarlo concede «escribime telemetría», nunca «conducí un agente» | `internal/adapters/transport/http/auth.go#validaIngesta`<br>`…/auth.go#isRutaIngesta` | `TestTokenDeIngestaNoAbreLaAPI` · `TestOTLPBajoLosTresGates` · `TestSpawnNoFiltraElTokenDeAPI` |
| **CAP-134** | `forward-externo-filtrado` | Escape hatch del operador, **apagado por default**: reenvía el evento ya proyectado, jamás el payload crudo, con indicador visible mientras está encendido | `internal/usecase/telemetria_service.go#TelemetriaService.Forward`<br>`internal/ports/telemetria.go#ForwardOTLP` | `TestForwardApagadoPorDefault` · `TestForwardNoReenviaCrudo` · `TestForwardSoloPorFlagDelOperador` |

---

## 2 · Hojas en módulos que ya existen (5 hojas)

| cap | módulo | slug | qué hace | puntero | check |
|---|---|---|---|---|---|
| **CAP-135** | `conductor` | `spawn-inyecta-telemetria` | El spawn de S1 inyecta las 8 env vars (`http/json`, endpoint loopback, token de ingesta y `OTEL_RESOURCE_ATTRIBUTES` con la llave del terreno). Hoy `cmd.Env` es `nil`: es cambio de código, no de config | `internal/adapters/agent/claudecode/conductor.go#SpawnEnv`<br>`…/conductor.go#Spawn` | `TestSpawnInyectaTelemetria` · `TestSpawnNoFiltraElTokenDeAPI` |
| **CAP-136** | `conductor` | `uso-del-turno-en-el-result` | El `result` del stream-json —que el adaptador **ya** decodifica para `ctxPct`— pasa a exponer el uso del turno con el split 5m/1h. **Sin un segundo parser**: el adaptador del agente sigue siendo el dueño único del protocolo | `internal/ports/agent.go#AgentEvent`<br>`internal/domain/telemetria.go#UsoDelTurno`<br>`internal/adapters/agent/claudecode/conductor.go#parseResult` | `TestResultTraeUsoDelTurno` · `TestSplitTTLLlegaDelStreamJSON` |
| **CAP-137** | `http-sse` | `superficie-telemetria` | Las 7 rutas `/api/telemetria/*` que consume el FE, con el contrato duro de que `null` y `0` no son lo mismo en ninguna de ellas | `internal/adapters/transport/http/telemetria.go#getResumen`<br>`…#getCajas`<br>`…#getDetalleCaja`<br>`…#getMejoras`<br>`…#getPortafolio`<br>`…#deleteTelemetriaArnes`<br>`…#getSalud`<br>`…#postProceso` | `TestNoAplicaNoEsCeroEnElWire` · `TestCifraLlevaConfianza` · `TestBorradoPorArnesTambienLimpiaRollup` |
| **CAP-138** | `cli-daemon` | `telemetria-cli` | `arnesia telemetria resumen\|mejoras\|salud\|purgar\|catalogo`: la vía de verificación E2E sin FE, reusando el MISMO usecase que el HTTP (patrón S0-D9) | `cmd/arnesia/telemetria.go#runTelemetria` | `TestTelemetriaCLIReusaElUsecase` |
| **CAP-139** | `fe-mapa` | `capa-mejora` | La capa «Mejora» del Mapa (ex slot `Tokens`): franja de la barra con ventana temporal, disclaimer de «estimado» y barra de cobertura; marcas en las cajas; «sin dato atribuible» en lo no atribuible; 4ª tab del inspector; tarjeta de punto de mejora; los **7 estados honestos** | `web/src/widgets/map-canvas/model/layers.ts#LAYERS`<br>`web/src/entities/telemetria/ui/valor-o-sin-dato.tsx#ValorOSinDato`<br>`web/src/entities/telemetria/ui/chip-confianza.tsx#ChipConfianza`<br>`web/src/widgets/mejora/ui/tarjeta-mejora.tsx#TarjetaMejora` | story-tests (`vitest --project=storybook`, headless): `estado-sin-datos` · `estado-cobertura-parcial` · `estado-s2-instrumentado` · `estado-s2-degradado` · `estado-otro-runtime` · `estado-catalogo-viejo` · `estado-por-huella` · `estado-que-guardamos` · `barra-con-disclaimer` · `canvas-nodos-no-atribuibles` |

**La fila del Portafolio (D17.2)** — arnés × puesto, con «sin dato» honesto para el que nunca
corrió — **extiende un capability existente** en vez de crear uno: es una columna nueva de
`fe-portafolio/lista-del-portafolio.yaml`. Se agrega ahí como `change_log` con
`type: extend`, no como hoja nueva.

---

## 3 · Cobertura R2 — ningún archivo nuevo queda huérfano

| archivo nuevo/tocado | reclamado por |
|---|---|
| `internal/domain/telemetria.go` | CAP-120 · CAP-127 · CAP-132 |
| `internal/domain/telemetria_costo.go` | CAP-126 |
| `internal/domain/telemetria_deteccion.go` | CAP-129 · CAP-130 |
| `internal/domain/telemetria_vistas.go` | CAP-128 |
| `internal/ports/telemetria.go` | CAP-120 · CAP-134 |
| `internal/usecase/telemetria_service.go` | CAP-127 · CAP-128 · CAP-134 |
| `internal/usecase/telemetria_mejoras.go` | CAP-130 |
| `internal/usecase/telemetria_retencion.go` | CAP-124 |
| `internal/adapters/telemetria/otlp/receptor.go` | CAP-118 |
| `internal/adapters/telemetria/otlp/decodifica.go` | CAP-119 |
| `internal/adapters/telemetria/otlp/mapa_cc.go` | CAP-121 |
| `internal/adapters/telemetria/hooks/proyecta.go` | CAP-121 · CAP-131 |
| `internal/adapters/telemetria/store/store.go` · `migracion.go` | CAP-122 |
| `internal/adapters/telemetria/store/rollup.go` | CAP-123 |
| `internal/adapters/telemetria/catalogo/catalogo.go` | CAP-125 |
| `internal/adapters/telemetria/descubrimiento/ficha.go` | CAP-132 |
| `internal/adapters/transport/http/telemetria.go` | CAP-137 |
| `internal/adapters/transport/http/auth.go` *(tocado)* | CAP-133 (+ CAP-48 ya lo reclama) |
| `internal/adapters/transport/http/router.go` *(tocado)* | CAP-118 (+ CAP-49 ya lo reclama) |
| `internal/adapters/agent/claudecode/conductor.go` *(tocado)* | CAP-135 · CAP-136 (+ las CAP-33…42 ya lo reclaman) |
| `cmd/arnesia/telemetria.go` | CAP-138 |
| `cmd/arnesia/hook.go` | CAP-131 |
| `web/src/entities/telemetria/**` · `web/src/widgets/mejora/**` | CAP-139 |
| `internal/adapters/telemetria/catalogo/precios.json` | **allowlist con razón**: dato generado por `go generate` desde LiteLLM (MIT), no código |

---

## 4 · Plantilla exacta de una de ellas

Copiar esto a `docs/product/capabilities/telemetria/allowlist-de-ingesta.yaml` **en el mismo
commit que crea `Proyectar` y `MapearLogRecord`**, no antes.

```yaml
---
# Capability nueva (paquete 2026-07-24-telemetria-embebida-otel, D15 + ANEXO H4).
# Schema: docs/product/_templates/capability.template.yaml
capability_id: arnesia.telemetria.allowlist-de-ingesta
cap_num: CAP-121
slug: allowlist-de-ingesta
name: "Allowlist de ingesta (ni identidad de cuenta, ni contenido, ni rutas)"
status: vivo          # GENERADO (R4) — no teclear
license: brand-local
module: telemetria
layer: adapter
verb: observar
functional_area: "telemetria.ingesta"
user_visible: false
nature: feature
pointers:
  - "internal/adapters/telemetria/otlp/mapa_cc.go#MapearLogRecord"
  - "internal/adapters/telemetria/hooks/proyecta.go#Proyectar"
  - "internal/domain/telemetria.go#EventoTelemetria"
valida:
  - "TestAllowlistNoPersistePII"
  - "TestHookNoReenviaContenido"
  - "TestAllowlistEsListaNoSugerencia"
  - "TestCwdDesconocidoNoSeGuardaCrudo"
  - "TestToolResultBytesSePersiste"
dev_preview:
  route: null
  how_to_navigate: null
  main_component: null
  api_endpoints:
    - "POST /v1/logs"
    - "POST /v1/metrics"
    - "POST /api/telemetria/proceso"
  e2e_test: null
created_in_story: 2026-07-24-telemetria-embebida-otel
created_date: 2026-MM-DD
last_modified: 2026-MM-DD
source_ref: "docs/product/stories/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md §6.1"
change_log:
  - story_id: 2026-07-24-telemetria-embebida-otel
    date: 2026-MM-DD
    type: new
    summary: "La telemetría que entra se persiste solo por lista declarada: la identidad de cuenta que el runtime manda en cada punto y el contenido de la conversación que manda el hook se descartan en la puerta, en los dos caminos de ingesta."
    merge_sha: null
scenarios:
  - id: pii-en-la-puerta
    name: "Llega telemetría con identidad de cuenta"
    status: live
    given: "el receptor OTLP recibiendo de un runtime que manda email e ids de cuenta en cada punto"
    when: "se ingiere el lote"
    then: "el evento persistido no contiene ninguno de esos campos, y tampoco aparecen como texto en el archivo de la base"
    verifica: "TestAllowlistNoPersistePII"
  - id: hook-no-reenvia-la-conversacion
    name: "El hook recibe el prompt completo por stdin"
    status: live
    given: "un hook de proceso invocado por el runtime con el payload real"
    when: "el hook proyecta y emite"
    then: "lo emitido tiene identificadores, nombre de herramienta y duración; no tiene prompt, respuesta, salida de herramienta ni rutas del usuario"
    verifica: "TestHookNoReenviaContenido"
  - id: campo-nuevo-cae-afuera
    name: "El emisor agrega un campo que no conocíamos"
    status: live
    given: "un payload con un atributo desconocido"
    when: "se ingiere"
    then: "el evento entra con lo que sí está declarado y el campo nuevo se descarta, sin error"
    verifica: "TestCamposDesconocidosNoRompen"
business_rules:
  - id: allowlist-no-denylist
    rule: "Se persiste lo declarado. Una denylist queda desactualizada frente a un emisor que agrega campos; una allowlist deja lo nuevo afuera por default."
    enforcement: ["docs/architecture/boundaries/ingesta-por-allowlist-declarada.md"]
    severity: critical
  - id: allowlist-en-los-dos-caminos
    rule: "El hook proyecta antes de mandar y el daemon re-valida al recibir. No se confía en que el emisor haya filtrado, aunque el emisor sea nuestro propio binario."
    enforcement: ["docs/architecture/boundaries/ingesta-por-allowlist-declarada.md"]
    severity: critical
related_capabilities:
  depends_on:
    - arnesia.telemetria.evento-canonico-de-telemetria
  enables:
    - arnesia.telemetria.almacen-de-telemetria
    - arnesia.telemetria.forward-externo-filtrado
  similar: []
  obsoletes: []
---

# Allowlist de ingesta

La telemetría del runtime llega con `user.email`, `user.account_uuid`, `user.account_id`,
`organization.id` y `user.id` **en cada punto y cada log record**; el payload del hook llega con el
prompt del usuario, la respuesta del asistente y la salida de las herramientas **en claro**. Nada
de eso hace falta: para el join alcanzan `session_id`, `prompt_id` y los `arnesia.*`.

El evento canónico se arma **campo por campo** desde las claves declaradas. No existe una ruta de
código que copie un mapa entero del emisor al evento — y hay un enforcer que lo impide, porque esa
es la única forma de que la allowlist sea una garantía y no una intención.
```
