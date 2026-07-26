# Escenarios de la capa «Mejora» — matriz exhaustiva

> **Para qué existe:** un diseño que solo contempla el caso feliz no se puede firmar. Acá está,
> escenario por escenario, **qué hace el sistema**, **qué ve el usuario** y **cómo se verifica**.
> Compañero de [`arquitectura-modulo.md`](arquitectura-modulo.md); las referencias `§N` apuntan ahí.
>
> **La regla que gobierna toda la matriz:** *la degradación es visible y honesta, nunca un cero
> fabricado.* Un número que no se pudo medir se muestra como «sin dato» con su motivo. Un
> concepto que el runtime no tiene se muestra «no aplica», no `0`. Un detector que no puede
> correr se muestra apagado con la razón, no oculto.
>
> **La regla que gobierna toda la verificación:** un negativo sin control positivo no es un
> resultado (§14). Todo escenario cuyo criterio sea «no llegó nada» lleva, en la misma corrida,
> un evento con marcador único que **sí** debe llegar.

**Leyenda de la columna «cómo se verifica»:**
`U` test unitario colocado · `F` fitness (`docs/architecture/fitness/`) · `S` story-test
(`vitest --project=storybook`, corre headless) · `E2E` corrida real contra `claude` ·
`M` manual con receta escrita · `⚠️` **no verificado todavía** — el escenario está diseñado
pero su comportamiento real no se observó.

---

## A · Captura: los tres escenarios y sus vecinos

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| A1 | **S1** — el usuario abre una sesión en el chat dock de ArnesIA | el spawn inyecta las 8 env vars (§10.2) con `OTEL_RESOURCE_ATTRIBUTES`; llega `api_request` por OTLP + `result` por stream-json + eventos del daemon | señal completa: gasto por caja, los 6 detectores, `atribucion: exacta` | `F TestSpawnInyectaTelemetria` · `E2E` corrida con receptor de puerto efímero |
| A2 | **S2-instrumentado** — el arnés corre en la terminal del usuario y **lleva bloque `env`** en sus settings (H9) | el receptor recibe `api_request` sin `arnesia.corrida` ⇒ deriva `escenario=s2-instrumentado`; no hay `result` ni eventos del daemon | dinero completo; B1 y B2 apagados **con motivo distinto de A3** | `F TestEscenarioSeDerivaDeLaSenal` · `F TestS2InstrumentadoTieneDineroYNoTieneSplit` · `E2E` (ya corrido, H9) |
| A3 | **S2-degradado** — el arnés corre fuera y **no** lleva bloque `env`; solo el hook | solo eventos de proceso; todas las columnas de dinero quedan `NULL` | «proceso medido, dinero no disponible: este arnés no lleva el bloque de telemetría» + cómo activarlo | `F TestS2DegradadoApagaLosDetectoresDeDinero` |
| A4 | El usuario corre el arnés **con ArnesIA cerrada** | el hook falla el POST en ≤250 ms y sale 0. Nada se pierde del trabajo; se pierde ese evento | nada en el momento. Después: «sin dato» para esa ventana | `U TestHookFailOpenSinDaemon` (control positivo: segunda invocación con daemon vivo **sí** llega) |
| A5 | **Runtime que no es Claude Code** (Codex, OpenCode…) | no hay adaptador todavía ⇒ cero eventos. El Portafolio marca el arnés `runtime no soportado` | «este arnés corre con `codex`: todavía no medimos ese runtime» — no una fila en cero | `S estado-runtime-no-soportado` · `U TestRuntimeSinAdaptador` |
| A6 | Runtime que **no reporta costo en USD** (D-3) | `costo_reportado_micros = NULL`; se calcula con el catálogo y se marca la fuente | «USD 0,74 **calculado con catálogo `v2026-07-20`** — este runtime no reporta costo» (estado 4) | `U TestCostoSoloCalculado` · `S estado-otro-runtime` |
| A7 | El usuario **apaga la telemetría** a mano (borra el bloque `env`, o `CLAUDE_CODE_ENABLE_TELEMETRY=0`) | deja de llegar. La conciliación (§6.4) cuenta los turnos esperados no medidos | «3 de 7 turnos no reportaron telemetría» | `U TestConciliacionCuentaLosNoLlegados` |
| A8 | Se instala un arnés nuevo y **nunca corrió** | no hay filas | estado 1: «este arnés nunca corrió con telemetría» + qué hacer. **Nunca «USD 0,00»** | `S estado-sin-datos` (aserta que el texto `0,00` NO está en el DOM) |

---

## B · Red, puerto y descubrimiento

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| B1 | **Daemon caído cuando el arnés emite** (hook) | POST falla → fail-open silencioso, sin reintento, sin escritura a disco, exit 0 | el trabajo no se interrumpe. El hueco aparece luego como «sin dato» | `U TestHookFailOpenSinDaemon` + control positivo |
| B2 | **Daemon caído cuando el runtime emite** (OTLP) | el exportador OTel del runtime reintenta y descarta por su cuenta; nosotros no nos enteramos | «sin dato» para esos turnos, vía conciliación | `E2E` con daemon apagado a mitad de corrida |
| B3 | **Puerto ocupado al arrancar el daemon** | `ListenAndServe` falla ⇒ el daemon no arranca ⇒ **no se escribe la ficha** (§7.2). Una ficha que nombra un puerto muerto es una mentira que se cobra en timeouts | el shell muestra el error de arranque de siempre | `F TestFichaSoloTrasEscuchar` |
| B4 | **El puerto cambia entre arranques** (`--addr` distinto) | la ficha se reescribe atómicamente con el nuevo endpoint. El hook la relee **en cada invocación**, no cachea | transparente | `U TestFichaSeRelePorInvocacion` |
| B5 | El puerto cambió pero un arnés **ya spawneado** tiene el viejo en su env | ese proceso pierde su telemetría hasta que termine. No hay forma de reconfigurar un proceso vivo | conciliación: «N turnos no reportaron» | `M` receta: arrancar sesión, reiniciar daemon en otro puerto, comprobar el contador |
| B6 | **WSL2 / devcontainer**: el agente corre adentro y no alcanza el `127.0.0.1` del host | no llega nada. La conciliación lo cuenta (en S1) | «los turnos de esta sesión no reportaron telemetría — el agente corre en un contenedor y no alcanza al daemon del host» + puntero a la nota de red | `M` receta con devcontainer + control positivo desde el host |
| B7 | **Windows: loopback promiscuo** — otro proceso local postea métricas falsas | pasa el Host gate y el token gate solo si tiene el `token_ingesta` (§7.3). Aun con él, los eventos que no resuelven a un arnés conocido entran `atribucion=sin-dato` y **no suman al total** (A15) | los totales no se mueven; el ruido aparece en «sin dato» y en `salud` | `F TestSinDatoNoSumaAlTotal` · `F TestOTLPBajoLosTresGates` |
| B8 | Un proceso local postea al endpoint **sin token** | `401`. No entra nada | — | `F TestOTLPBajoLosTresGates` |
| B9 | Alguien intenta usar el `token_ingesta` contra `/api/sessions` | `401`. El token acotado **no abre la API** | — | `F TestTokenDeIngestaNoAbreLaAPI` |
| B10 | **Dos daemons** del mismo usuario, dos puertos | último que arranca gana la ficha ⇒ los hooks nuevos van al nuevo. Los spawns de S1 de cada uno siguen yendo al suyo (env por proceso). Los dos escriben el **mismo** `.db` bajo WAL | nada raro; los datos son de los dos | `U TestDosEscritoresSobreElMismoDB` (WAL + `busy_timeout`) |
| B11 | **Dos instalaciones de ArnesIA** (dos usuarios del SO) | HOME distinto ⇒ `.db` distinto ⇒ ficha distinta. Cero interferencia | cada usuario ve lo suyo | `U TestRutasPorHome` |
| B12 | La ficha quedó **huérfana** (daemon murió sin shutdown) | el hook intenta, falla en ≤250 ms, sale 0. **No se hace liveness check por pid**: es más frágil que el timeout | transparente | `U TestHookFichaHuerfana` |
| B13 | El directorio de config del SO **no existe o no es escribible** | el daemon arranca igual, loguea `warn` y `salud.descubrimiento` dice «no publicada» ⇒ el hook de S2 no encuentra el daemon; S1 y `s2-instrumentado` (que van por env/settings, no por ficha) siguen | «la telemetría por hook no está disponible: no pude publicar la ficha del daemon» | `U TestFichaNoEscribibleDegradaHonesto` |
| B14 | Un arnés en **`s2-instrumentado`** postea a `/v1/logs` **sin token** | se acepta bajo Host loopback (**A22**: el token no puede viajar en el bloque `env`, H10.2). El evento entra y, si no resuelve a un arnés conocido, queda `sin-dato` y no suma | transparente cuando el arnés es conocido | `F TestOTLPAceptaSinTokenBajoLoopback` + `F TestTokenDeIngestaNoAbreLaAPI` (la API sigue cerrada) |
| B15 | El operador enciende **`--telemetria-ingesta-token-obligatorio`** | `/v1/*` pasa a exigir token ⇒ **`s2-instrumentado` deja de reportar** salvo que el operador ponga el token literal en un settings **no versionado** | «modo estricto de ingesta: los arneses que corren fuera de ArnesIA dejan de reportar dinero» — se dice, no se descubre | `U TestModoEstrictoApagaS2Instrumentado` |

---

## C · Payload y wire format

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| C1 | **Payload malformado** (JSON roto) | `400`, **el lote entero se descarta** (nunca a medias), contador `payload_invalido++`, `slog.Warn` sin el cuerpo | `salud`: «12 lotes rechazados por formato» | `U TestPayloadMalformado` |
| C2 | **Payload gigante** (> 4 MiB) | `413` vía `MaxBytesReader`, sin leer el cuerpo a memoria | ídem | `F TestPayloadGiganteSeRechazaSinLeerlo` |
| C3 | **Campos desconocidos** (versión nueva del runtime) | `encoding/json` los descarta; el evento entra con lo que sí conoce. **No es un error** — es lo que hace que una versión nueva no rompa el receptor | nada | `U TestCamposDesconocidosNoRompen` |
| C4 | **`intValue` como número JSON** (off-spec, lo que hace Claude Code) | `json.Number` lo acepta | nada | `U TestIntValueComoNumeroYComoString` (las dos formas, mismo resultado) |
| C5 | `intValue` como **string** (un runtime que sí cumple la spec) | ídem | nada | mismo test |
| C6 | Un valor numérico **no parseable** (`"abc"`) | ese atributo se descarta; el resto del evento entra. Un token ilegible no invalida el costo reportado | nada; se cuenta en `salud.atributo_ilegible` | `U TestAtributoNumericoIlegible` |
| C7 | El runtime manda **protobuf** (no eligió `http/json`) | `415` con cuerpo que **nombra la causa y el fix**. Jamás un 200 que finge haber guardado | `salud`: «un emisor está mandando protobuf; falta `OTEL_EXPORTER_OTLP_PROTOCOL=http/json`» | `U TestProtobufSeRechazaConMotivo` |
| C8 | **Métricas Delta** (lo normal, V5.1) | se suman y listo, sin diferenciar contadores | nada | `U TestSumaDelta` |
| C9 | **Métricas Cumulative** (temporalidad 2, no observada) | **no se adivina**: el punto se marca `sin-dato` y se cuenta en `salud.temporalidad_no_soportada`. Fabricar un delta sin estado previo es inventar | `salud` lo dice | `U TestCumulativeNoSeAdivina` |
| C10 | **Reinicio del proceso emisor** a mitad de sesión | con Delta no hay problema: cada lote es un delta propio. Es exactamente la clase de bug que V5.1 eliminó | nada | `U TestReinicioDelEmisorNoDuplica` |
| C11 | **Evento duplicado** (reintento del exportador) | dedupe por `(sesion_id, turno_id, tipo_evento, ts_emisor)` en el writer | nada | `U TestEventoDuplicadoNoSeCuentaDosVeces` |
| C12 | Evento **sin `session_id`** | rechazado con `ErrEventoSinSesion`: no es atribuible ni deduplicable | `salud` lo cuenta | `U TestEventoSinSesionSeRechaza` |
| C13 | Evento sin `prompt_id` (p. ej. `SessionStart`) | entra con `turno_id` vacío: sirve para el total del arnés, **no** para el join a nivel caja | los números por caja no lo incluyen | `U TestEventoSinTurnoNoEntraAlJoin` |
| C14 | **Pánico en el decodificador** | `recover()` en el handler → `500` + `slog.Error`; **el daemon sigue vivo**. Es red de seguridad, no estrategia | nada visible | `U FuzzDecodificarLogs` (sembrado con los 3 payloads reales) |
| C15 | Un lote con **cientos de log records** | se procesan todos; el tope es el de bytes, no el de cantidad | nada | benchmark `BenchmarkReceptorLogs` |
| C16 | **`tool_result` con tamaños** (H8) | se persisten `tool_input_bytes`/`tool_result_bytes`; **cero contenido** | no se muestra todavía (B11 fuera del MVP) pero el dato queda | `F TestToolResultBytesSePersiste` |

---

## D · Tiempo

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| D1 | **Reloj del sistema hacia atrás** (NTP, VM suspendida) | se guardan los dos tiempos (A10). Un `ts_emisor` fuera de `[ts_recibido − 24 h, ts_recibido + 5 min]` se marca `reloj_sospechoso=1` y **la ventana y el rollup usan `ts_recibido`** | nada, salvo en el drill-down: «marca de tiempo del emisor inconsistente» | `U TestRelojHaciaAtrasNoRompeLaVentana` |
| D2 | Reloj hacia **adelante** | mismo tratamiento | ídem | mismo test |
| D3 | El emisor **no manda `timeUnixNano`** | `ts_emisor = NULL`; manda el nuestro | nada | `U TestSinTimestampDelEmisor` |
| D4 | Un turno **cruza la frontera horaria** | cuenta en las dos horas del rollup ⇒ `COUNT(DISTINCT turno_id)` sobreestima levemente. **Sesgo declarado y en contra**: infla el denominador, o sea baja el costo-por-turno que mostramos (§5.5) | el sesgo está escrito en la tarjeta de mejora | `U TestTurnoCruzaHora` (aserta la dirección del sesgo) |
| D5 | Cambio de huso / DST | todo se guarda en **UTC RFC3339**; la UI convierte al mostrar | nada | `U TestTodoEnUTC` |

---

## E · Identidad y atribución

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| E1 | **El mismo arnés instalado en dos homes** | son dos identidades `(home,id)` distintas ⇒ dos `instalacion_id` ⇒ filas separadas. El Portafolio ya modela esto | dos filas, no una suma | `U TestDosHomesNoSeFusionan` |
| E2 | La misma identidad con **dos instalaciones** (canónico + espejo) | `instalacion_id` distingue; el resumen por arnés puede sumarlas si el usuario lo pide, **nunca por default** | «vitalia · canónico» y «vitalia · instalación» separadas | `U TestInstalacionesNoSeSumanSolas` |
| E3 | **Arnés sin sello** (`arnes.l0.json` ausente) | la telemetría igual se guarda si se puede atribuir por hash o cwd; el arnés se muestra `no sellado` | «este arnés no está sellado: la atribución es por huella» | `U TestArnesSinSello` |
| E4 | **Sello sin bloque `telemetria:`** | `arnesia conformance` marca `arnes-declara-telemetria` en fail ⇒ **no sella, no publica** (es el mecanismo de obligación) | el gate de conformance lo dice con el check nombrado | `U TestConformanceArnesSinBloqueTelemetria` |
| E5 | `plugin_id_hash` **desconocido** (nunca lo instalamos nosotros) | no resuelve ⇒ se prueba cwd ⇒ si tampoco, `sin-dato` | «1 corrida sin atribución» en la barra de cobertura | `U TestHashDesconocido` |
| E6 | `plugin_id_hash` conocido pero el nombre está redactado (`third-party`) | se atribuye **por hash** al arnés; la caja queda sin dato | estado 6: «identificado por huella del arnés, no por nombre», con subrayado punteado | `U TestAtribucionPorHash` · `S estado-por-huella` |
| E7 | `cwd` de un hook que **no mapea** a ninguna instalación conocida | `sin-dato`; el `cwd` crudo **no se guarda** (A14), solo su huella | «1 corrida sin atribución» | `U TestCwdDesconocidoNoSeGuardaCrudo` |
| E8 | `cwd` que mapea a una instalación conocida | `atribucion=por-proceso`; da arnés + instalación, **no** caja | subrayado punteado + «atribuido por proceso» al hover | `U TestAtribucionPorProceso` |
| E9 | El mismo turno llega por **los dos canales** (OTel y hook) | el join por `(sesion_id, turno_id)` los une en un `TurnoUnido`; dinero del uno, proceso del otro | una fila, completa | `U TestJoinPorSesionYTurno` |
| E10 | Llega el hook y **nunca** el `api_request` del mismo turno | `TurnoUnido` con proceso y **dinero `null`** — no 0 | la fila del turno dice «costo no medido» | `U TestTurnoSoloConProceso` |
| E11 | Llega el `api_request` y **nunca** el hook | `TurnoUnido` con dinero y sin proceso; los detectores de proceso lo saltan | ídem, del otro lado | `U TestTurnoSoloConDinero` |
| E12 | ⚠️ `plugin_id_hash` **no determinista entre máquinas** (V7.2, sin verificar) | el diseño no lo asume: la tabla se aprende local (`como_se_aprendio="spawn-controlado"`). Si resultara determinista, se podría sembrar al instalar | ninguna diferencia visible | ⚠️ pendiente: correr el mismo plugin en otra máquina |

---

## F · Dinero y catálogo

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| F1 | **Modelo desconocido para el catálogo** | `Precio()` devuelve `ok=false` ⇒ `costo_calculado = NULL`, `sin_tarifa` poblado. **Jamás 0** | «costo calculado no disponible: el catálogo no conoce `modelo-x`» | `U TestModeloDesconocido` |
| F2 | Modelo conocido pero **sin tarifa de cache 1h** | ese bucket va a `sin_tarifa`, `costo_completo=0`; el número es **parcial y lo dice** | «costo parcial: falta la tarifa de escritura de cache 1h» | `U TestTarifaParcial` |
| F3 | **Catálogo viejo** (sin refrescar) | se usa el embebido y se declara su versión | estado 5: ⚠️ «precios del release, sin refrescar desde el 20/07» | `S estado-catalogo-viejo` |
| F4 | **Sin internet** y refresco encendido | falla el refresco, se sigue con el embebido, se registra el intento fallido | mismo aviso que F3 | `U TestRefrescoFallidoNoRompe` |
| F5 | Refresco **apagado** (default, A11) | no se hace ninguna conexión saliente | «precios del release · catálogo `v2026-07-20`» sin ⚠️ (funcionar offline es lo normal, no una anomalía) | `F TestForwardApagadoPorDefault` (mismo dialer fake) |
| F6 | El catálogo devuelve un **precio disparatado** tras un refresco | el costo calculado diverge del reportado ⇒ la divergencia es **visible** (es el oracle A6 corriendo en producción) | «reportado USD 1,92 · calculado USD 19,20 — divergencia 10×» en la 4ª tab | `U TestDivergenciaDeCostoEsVisible` |
| F7 | El runtime **cambia su tarifa** y nuestro catálogo queda atrás | misma divergencia, otro origen. No se elige uno en silencio: se muestran los dos | ídem | mismo test |
| F8 | El runtime **no reporta** costo y el catálogo **tampoco** conoce el modelo | los dos costos `NULL` | «sin dato de costo para esta corrida» — no una fila en cero | `U TestSinNingunCosto` |
| F9 | Se **aplanan los tiers** al copiar el catálogo (bug phoenix#14314) | el test lo impide antes de shipear | — | `U TestCosteoNoAplanaLosTiers` |
| F10 | Se **olvida el cache write** (bug langfuse#14249) | ídem | — | `U TestCosteoCobraElCacheWrite` |
| F11 | Se **suman buckets que se solapan** (bug langfuse#12306) | el flag `aritmetica` del adaptador lo impide | — | `U TestCosteoNoSumaBucketsQueSeSolapan` |
| F12 | El total se lee como si fuera facturación | **imposible por contrato**: `Estimado` es siempre `true` mientras la fuente sea `cost_usd_micros` (H4) | «USD 4,82 estimado · ⓘ estimado por el runtime, no es facturación» | `S barra-con-disclaimer` (aserta el texto) |

---

## G · Proceso

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| G1 | **Corrida sin ninguna caja activa** (chat libre en el Dock) | `caja_id` vacío ⇒ el gasto va al arnés, no a una caja | el total del arnés incluye la sesión; ninguna caja se la atribuye | `U TestGastoSinCaja` |
| G2 | **Sesión abandonada** (escribe cache y nunca lo lee) | B6 la detecta: creación cold sin turno posterior | punto de mejora «sesión abandonada · 100 % desperdicio» + fix «no spawnear / reusar sesión» | `U TestB6SesionAbandonada` |
| G3 | **Rotación de contexto** (RF-195) | el daemon emite `EventoRotacion`; B2 cotiza el `cache_creation` del primer turno del proceso fresco | punto de mejora «costo de la rotación» + fix «subir/bajar `umbralRot`» — convierte un número mágico en un trade-off cotizado | `U TestB2CostoDeLaRotacion` |
| G4 | Rotación en **s2** (cualquiera de los dos modos) | no hay evento de rotación: la rotación es decisión de ArnesIA | B2 apagado con motivo: «la rotación es decisión de ArnesIA» | `F TestS2DegradadoApagaLosDetectoresDeDinero` |
| G5 | **`/compact`** | `query_source="compact"` en el `api_request` ⇒ el turno queda etiquetado. B7 (cuantificar el costo de cada `/compact`) **no entra al MVP** | el turno aparece en el desglose; no hay detector todavía | ⚠️ no verificado (V7.4): hace falta una corrida con compactación forzada |
| G6 | **Subagentes** (`isSidechain`) | los `api_request` del sidechain comparten `session.id` ⇒ suman al total. B8 (overhead de subagentes) fuera del MVP | el total los incluye; no se separan todavía | ⚠️ no verificado (V7.4) |
| G7 | **Rechazo en el gate** | el daemon emite `EventoGate` con `resultado=rechazado`; P1 une el costo del turno con el rechazo | la frase objetivo: «la caja Y falla el gate 3 de cada 4 veces y esas corridas costaron USD Z» | `U TestP1CajaQueSeRechaza` |
| G8 | Rechazo en el gate en **s2-instrumentado** | no hay evento de gate (es del daemon) ⇒ P1 corre **parcial**: detecta reintentos y fracasos de herramienta vía `tool_result`, no rechazos de gate | P1 marcado `cobertura parcial` con motivo, no un ✅ liso | `U TestP1ParcialSinGate` |
| G9 | **Corrida cancelada** por el usuario | `resultado=cancelado`; los tokens ya gastados se contabilizan (se pagaron) | el desglose muestra el gasto y el desenlace | `U TestCorridaCancelada` |
| G10 | **Reintento** de una caja (BoxConductor, `repair-cap`) | `resultado=reintento` por iteración; el costo de las iteraciones suma a la caja | «esta caja costó USD X en 3 iteraciones» | `U TestReintentosSuman` |
| G11 | **Cambio de modelo** a mitad de sesión | B3 lo detecta: cambiar modelo invalida el cache | punto de mejora + fix «fijar el modelo de la caja» | `U TestB3CambioDeModelo` |
| G12 | **Re-warm por TTL** vencido (S1) | B1 aplica el umbral algebraico `(2−1,25)/(2−0,1) ≈ 39,47 %`, citado | tarjeta con contrafactual, umbral, sesgo y `[ver el cálculo]` | `U TestB1BreakEvenTTL` |
| G13 | Un detector **no aplica** | viaja en `no_aplican[]` con motivo obligatorio (§9.2). Omitirlo obligaría al FE a mostrar 0 o nada | detector apagado **con la razón a la vista** | `F TestDetectorQueNoAplicaTraeMotivo` · `S estado-s2-degradado` |
| G14 | Los **7 detectores fuera del MVP** | viajan en `no_medidos[]` con «no medido todavía» (patrón `sin-check`) | lista visible de lo que todavía no medimos — gap declarado, no escondido | `U TestNoMedidosSeDeclaran` |
| G15 | **Nodos no atribuibles** (conocimiento, reglas, hooks, artefactos) | `atribuible=false` + motivo; `costo_micros=null` | atenuados con «sin dato atribuible» — poner una cifra ahí sería inventarla | `S canvas-nodos-no-atribuibles` |

---

## H · Almacén

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| H1 | **DB corrupta** (archivo truncado) | `New()` falla al abrir ⇒ se **archiva** (`…-corrupta-<fecha>.db`) y se crea una nueva vacía. Nunca se borra en silencio | «la base de telemetría estaba dañada; se archivó y se arrancó una nueva» + la ruta del archivo | `U TestDBCorruptaSeArchiva` |
| H2 | **DB bloqueada** (`SQLITE_BUSY`) | `busy_timeout=5000` + writer con `SetMaxOpenConns(1)`: database/sql serializa. Si aun así falla, el lote se descarta y se cuenta | `salud`: «N lotes perdidos por bloqueo» | `U TestDosEscritoresSobreElMismoDB` |
| H3 | **Disco lleno** | el `INSERT` falla ⇒ el lote se descarta, se cuenta, se loguea `warn`. **El daemon sigue** — la telemetría jamás degrada al sistema instrumentado | «no se pudo guardar telemetría: sin espacio en disco» | `U TestDiscoLlenoNoTumbaElDaemon` (fs de prueba con cuota) |
| H4 | **Retención vencida** | purga al boot y cada 6 h: `DELETE FROM evento WHERE ts_recibido < now-90d`; el rollup sobrevive (24 meses) | los totales históricos siguen (rollup); el drill-down de un turno viejo dice «detalle purgado, resumen conservado» | `U TestPurgaRespetaTTL` |
| H5 | El usuario **borra la telemetría de un arnés** (D15.3) | `DELETE FROM evento WHERE arnes_id=?` + `DELETE FROM rollup_hora WHERE arnes_id=?` + re-agregado de las horas tocadas, todo en una transacción | «se borró la telemetría de vitalia (1 284 eventos)» y el arnés vuelve al estado 1 | `U TestBorradoPorArnesTambienLimpiaRollup` |
| H6 | **Upgrade de esquema aditivo** (versión +1) | migración en transacción; los datos viejos quedan, la columna nueva es `NULL` para ellos — que es «no aplica», correcto | nada | `U TestMigracionAditivaConservaDatos` |
| H7 | Alguien intenta una migración **destructiva** | el test lo rompe antes de mergear | — | `U TestMigracionesSonAditivas` |
| H8 | **Ruptura de generación** (no expresable aditivamente) | se **archiva** el archivo (`telemetria-gen1-<fecha>.db.archivada`) y se arranca vacío. **Jamás se borra** — no es reconstruible como el índice | «la historia anterior se archivó en …» en `salud` y en la UI | `F TestGeneracionArchivaNoBorra` |
| H9 | Se borra `index.db` (wipe del índice por mismatch) | `telemetria.db` **no se toca**: son archivos distintos, y esa es la razón de A2 | la telemetría sobrevive a un rebuild del Mapa | `F TestTelemetriaDBNoEsElIndice` |
| H10 | **Cardinalidad explota** (arnés con cajas generadas) | al superar 50 000 filas/mes en el rollup, `caja_id` colapsa a `(otros)` y la fila se marca `cardinalidad_colapsada` | «el desglose por caja se agrupó: demasiadas cajas distintas» — degradación declarada | `U TestRollupColapsaCardinalidad` |
| H11 | El `.db` pasa **500 MB** | aviso en `salud`; se sugiere bajar la retención | «la base de telemetría ocupa 512 MB» + acción | `U TestAvisoDeTamano` |
| H12 | El rollup queda **atrasado** tras una caída | `rollup_cursor` permite reanudar desde el último `evento.id` agregado | nada | `U TestRollupReanudaDesdeElCursor` |
| H13 | Cola del receptor **llena** | descarta, cuenta, y lo dice en la respuesta OTLP (`partialSuccess.rejectedLogRecords`). **El emisor nunca se bloquea** | `salud`: «N puntos descartados por saturación» | `F TestReceptorNoBloqueaAlEmisor` |

---

## I · Operación y ciclo de vida

| # | qué pasa | qué hace el sistema | qué ve el usuario | cómo se verifica |
|---|---|---|---|---|
| I1 | El operador **enciende el forward externo** (C2) | reenvía el `EventoTelemetria` **ya proyectado**, nunca el OTLP crudo (que llevaría el email de quien corra el arnés) | indicador visible mientras está encendido | `F TestForwardNoReenviaCrudo` |
| I2 | Un **arnés** intenta configurar el forward | no es capacidad expuesta al paquete (D13, invariante 2). No hay endpoint ni env que lo permita | — | `F TestForwardSoloPorFlagDelOperador` |
| I3 | **Desinstalación** de ArnesIA | el `.deb`/`.rpm` no borra `~/.arnesia` (dato del usuario). La ficha del daemon sí se borra en el shutdown | los datos quedan; se pueden borrar a mano | `M` receta de desinstalación |
| I4 | **Self-update** con el daemon corriendo | el daemon se reemplaza y reinicia; la ficha se reescribe al volver a escuchar. Los spawns vivos pierden el canal hasta terminar | conciliación cuenta los turnos del hueco | `M` receta con el binario instalado |
| I5 | El hook corre y **el binario `arnesia` no existe** (arnés copiado a otra máquina — S3, fuera de alcance) | ✅ **verificado (H10.3)**: el runtime lo absorbe — `exit 0`, `is_error:false`, stderr vacío, duración normal, y el otro hook de la misma corrida sí se ejecutó. **No hay que envolver el comando ni apoyarse en el `timeout`** | transparente: el turno del usuario no cambia | `E2E` ya corrido (ANEXO §H10.3, con hook de control) |
| I12 | El hook **existe y falla** (bug propio, sale ≠0) | ⚠️ **NO lo cubre H10.3**: esa garantía es para el binario ausente. Un `exit 1` en `UserPromptSubmit` sí puede bloquear el turno ⇒ el contrato «exit 0 siempre» sigue siendo obligación nuestra | debería ser transparente, y lo es solo si respetamos el contrato | `U TestHookNoTardaNiFalla` (aserta exit 0 en todas las ramas, incluidas las de error) |
| I6 | El hook **tarda** más de 250 ms | se abandona y sale 0. La instrumentación nunca demora el trabajo del usuario | nada | `U TestHookNoTardaNiFalla` |
| I7 | El hook **imprime** algo a stdout | no puede: el contrato es stdout vacío (imprimir inyecta texto al contexto del agente) | — | `U TestHookStdoutVacio` |
| I8 | El paquete de un arnés **no shipea el hook** | `arnes-porta-hook-proceso` en fail ⇒ no sella | el check nombrado en el reporte de conformance | `U TestConformanceArnesSinHook` |
| I9 | Un hook **reenvía su stdin entero** a `127.0.0.1` | cumple `telemetria-no-egresa` y **aun así filtra la conversación al almacén local** ⇒ lo caza el check hermano `hook-proyecta-campos` (ANEXO H4) | check en fail con el nombre del campo filtrado | `F TestHookNoReenviaContenido` |
| I10 | Llega PII al receptor (siempre llega: está en cada punto) | se descarta en la puerta por allowlist; nunca toca el `.db` | la política de datos, visible: «nada de tu cuenta, nada del contenido» (estado 7) | `F TestAllowlistNoPersistePII` (busca las 5 claves **como subcadena del archivo `.db`**) · `S estado-que-guardamos` |
| I11 | El operador quiere **auditar** qué guardamos | `arnesia telemetria salud` + el estado 7 de la UI + el `.db` es SQLite legible con cualquier cliente | la lista de campos, no una promesa | `M` receta: abrir el `.db` y listar columnas |

---

## J · Escenarios que este paquete NO cubre, y hay que decirlo

| # | escenario | por qué queda afuera | qué se ve mientras tanto |
|---|---|---|---|
| J1 | **S3** — arnés en la máquina de un cliente sin ArnesIA | D12.1 lo descartó: choca con `superficie-local-confinada` y abre consentimiento/GDPR. Paquete propio si se reabre | nada; no hay canal |
| J2 | **Capa Desempeño** a nivel Mapa | la señal ya llega (`duration_ms` por request **y por herramienta**, H8): falta el **diseño** de qué es «desempeño» en el Mapa | el slot `Desempeño` sigue `disabled`, con tooltip honesto que dice que falta diseño, no señal |
| J3 | **Los otros 7 detectores** de la familia B | no pasan la regla A4 todavía (falta el fix concreto o la cotización) | lista `no_medidos[]` visible |
| J4 | **B11 (tool results obesos)** en particular | **ya no está bloqueado por dato** (H8: `tool_result_size_bytes`), sí por diseño del fix | los bytes se persisten desde el día 1 (A21) |
| J5 | Adaptadores de **otros runtimes** | H1: no competimos en cobertura; el eje es arnés × empresa × puesto | «todavía no medimos ese runtime» (A5) |
| J6 | **Series temporales / tablero** | es lo que ya venden ccusage y Dynatrace; competir ahí es perder (H1) | la ventana temporal existe, el tablero no |
