# Arquitectura de telemetría de ArnesIA — consolidado (2026-07-26)

> Síntesis de toda la investigación del paquete: `investigacion-runtimes.md` ·
> `investigacion-plataformas.md` · `investigacion-stack-embebible.md` + el informe externo sobre
> `cache-refund`. **Estado: PROPUESTA. Nada firmado — falta 🧑‍⚖️ sobre D9.**
> Fuente de las decisiones: `decisiones.md` D1-D11.

---

# Parte 1 — Todas las recomendaciones, una por una

## Grupo A · Método de medición (origen: `cache-refund`)

| id | recomendación | por qué |
|---|---|---|
| **A1** | **Umbral derivado, no modelo ajustado.** El veredicto es una desigualdad algebraica con constantes citadas (ej. break-even de TTL = `(2−1.25)/(2−0.1) ≈ 39.47 %`), no un fit | Reproducible a mano; se audita, no se cree |
| **A2** | **Contrafactual simétrico.** No reportar «gastaste $X» sino «el mundo alternativo costaba $Y, la diferencia es $Y−X» | Convierte un dashboard en una decisión |
| **A3** | **Declarar el sesgo de cada aproximación** y su dirección, y que el sesgo neto vaya **en contra** de la recomendación | Es la doctrina de honestidad del repo aplicada a números |
| **A4** | **Regla de aceptación:** toda métrica debe ser (1) computada de datos propios, (2) cotizada en dinero, (3) emparejada con **UN** fix concreto. Si no cumple las tres, es opinión, no producto | Criterio de corte para el `spec.md` |
| **A5** | **Filas en $0.00 son honestas**, no datos faltantes — mismo patrón que `sin-check`/`no-reconocido` | Anti pass-fabricado |
| **A6** | **Oracle independiente + test de paridad**: segunda implementación que debe coincidir exacto; las diferencias son correcciones nombradas, no fudge factors | Ya es patrón del repo (cifras se generan) |
| **A7** | **Versionar el score** (`score_version: 1`) | La gente los va a comparar entre releases |
| **A8** | **Nunca escribir settings de terceros por cuenta propia** — solo con backup + confirmación | Guardrail para el chat embebido |

## Grupo B · Detectores de fuga (13, priorizados por impacto × facilidad)

| id | detector | señal | dif. | nota |
|---|---|---|---|---|
| **B1** | re-warm por TTL expirado (`R/C` vs 39.47 %) | `usage` por turno + `timestamp` + sesión | baja | **requiere el split 5m/1h ⇒ depende de D8** |
| **B2** | **costo de la ROTACIÓN de contexto** | evento propio + creación del 1er turno del proceso fresco | baja | **EXCLUSIVO DE ARNESIA** — la rotación es decisión nuestra. Convierte `umbralRot` de número mágico en trade-off cotizado |
| **B3** | model-switch invalidation | `model` por turno | baja | cambiar modelo tira el cache |
| **B4** | **fuga por arnés / empresa / puesto** | join sesión → `Session.Arnes/Empresa/Puesto` | baja | **EXCLUSIVO** — es el producto |
| **B5** | cache escrito y nunca leído (amortización) | creación vs lecturas posteriores | media | `uncachedCost − actualCost < 0` ⇒ *el cache te costó plata* |
| **B6** | sesión abandonada | creación cold sin turno posterior | baja | 100 % desperdicio puro |
| **B7** | compaction rewrite | `compact_boundary` / `query_source="compact"` | baja | cuantifica el costo de cada `/compact` |
| **B8** | overhead de subagentes | `isSidechain` / `query_source` | baja | sidechains son siempre 5m |
| **B9** | churn de `effort`/`speed` | `api_request` | media | cambiar `effort` invalida bloques |
| **B10** | reintentos | `attempt` | media | tokens pagados dos veces |
| **B11** | tool results obesos | `tool.execution.result_tokens` | media | envenenan la ventana → fuerzan rotación (encadena con B2) |
| **B12** | TTL reality check (watchdog de regresión) | split 5m/1h | alta | **depende de D8** |
| **B13** | costo por MCP server / plugin | atributos de atribución | media | encaja con la capa Tokens |

**Familia nueva, destapada al pensar la obligación (Parte 3):** además de fuga de *dinero*, hay
**fuga de PROCESO** — caja que siempre se rechaza en el gate, paso que se reintenta, subagente que
se dispara y no aporta. Requiere eventos de hook, no de OTel.

## Grupo C · Por qué NO usamos Langfuse (y qué queda como escape hatch del operador)

> ⚠️ **Este grupo NO propone que ArnesIA consuma Langfuse.** Propone lo contrario. El nombre anterior
> («Langfuse y el canal») se prestaba a leerlo al revés — corregido 2026-07-26 tras consulta del
> operador.
>
> **El usuario final que no tiene Langfuse —que son TODOS por default— recibe el 100 % de la
> funcionalidad.** La cadena completa vive en el binario que ya se instala:
> `runtime → 127.0.0.1:4318 (receptor en el daemon) → SQLite local → capa de mejora en el Mapa`.
> Ningún paso toca Langfuse. Es D1, orden directa del operador.

| id | recomendación |
|---|---|
| **C1** | **Langfuse es CONSUMIDOR de la señal, no PRODUCTOR.** La fidelidad la define el emisor ⇒ **no usarlo NO baja la calidad de medición**, solo la superficie de análisis. Es el argumento a favor de descartarlo, no de adoptarlo |
| **C2** | **Escape hatch del OPERADOR, apagado por default:** el receptor *puede* reenviar OTLP a un endpoint externo si el operador lo configura deliberadamente. Existe por una sola razón: alpacapurpura ya corre un Langfuse propio (`~/.prenter/observatorio/langfuse/`) para su flota, y esto le deja recuperar traces/evals sin que el producto dependa de nada. **Si ese Langfuse desaparece, no se rompe nada** |
| **C3** | **No clonar Langfuse.** Son 6 subsistemas; necesitamos ~1,5. Los otros 4 (prompt mgmt, evals, datasets, playground) no los usamos, y su peor parte (ClickHouse+Postgres+Redis+worker) es inembebible |

### ⚠️ C2 vs «cero egreso de red» (D12.1) — no es contradicción, son dos niveles

La consulta del operador destapó una ambigüedad que hay que dejar cerrada por escrito, o alguien la
lee como contradicción más adelante —o peor, un arnés termina configurando un forward hacia afuera:

| nivel | quién lo controla | ¿puede egresar? |
|---|---|---|
| **arnés / hook** (escenario S2) | el paquete instalado | **NUNCA.** Solo loopback o disco local. Check duro `telemetria-no-egresa` |
| **daemon** | el operador, config deliberada y explícita | sí — **apagado por default**, con indicador visible en la UI |

**El check `telemetria-no-egresa` ata al ARNÉS, no al daemon.** Y un arnés **no puede alcanzar ni
configurar** el forward del daemon: no es una capacidad expuesta al paquete. Sin esa separación
escrita, las dos reglas se leen como incompatibles y una de las dos se rompe en la implementación.

## Grupo D · Multi-runtime

| id | recomendación |
|---|---|
| **D-1** | **El adaptador base es `stream-json`, OTel es enriquecedor.** 5 de 6 runtimes dan usage por turno en stdout; solo 3 dan OTel útil en headless. **Invierte la jerarquía asumida en D4** |
| **D-2** | **La llave canónica NO puede ser `skill.name`** (concepto exclusivo de Claude Code). Debe ser del terreno propio: **arnés × caja/paso × sesión**. Bloquea el mockup |
| **D-3** | **La tabla de precios es obligatoria** — 5 confirmaciones independientes: solo Claude Code emite costo USD; la semconv GenAI **no define costo en dinero** |
| **D-4** | **Esquema = superset con «no aplica» explícito ≠ 0.** Un 0 donde el concepto no existe es mentira |
| **D-5** | **Los detectores también son per-runtime.** El break-even de TTL depende de la estructura 1.25×/2×/0.1× de Anthropic; OpenAI no cobra por escribir cache ⇒ B1 no aplica a Codex |
| **D-6** | **La regla de acumulación es propiedad del ADAPTADOR**, nunca del agregador. CC acumula en `result` · Codex re-emite `last_token_usage` · Amp es per-message. Tres runtimes, tres reglas |
| **D-7** | **`usage_arithmetic: disjoint \| inclusive` por adaptador.** Anthropic disjunto, OpenAI anidado, OTel dice «inclusive». Sin el flag el error es ~100 % o ~28 %. **Nadie lo tiene** |
| **D-8** | **`OTEL_RESOURCE_ATTRIBUTES` es el vector universal de atribución** (CC, Copilot, Codex, Goose, OpenCode). Límite: se fija **por proceso** |
| **D-9** | **Namespace propio legítimo** para skill/plugin/sub-agente — la semconv no los cubre. No es reinvención |
| **D-10** | **Las métricas OTel pierden cache y reasoning** (`gen_ai.token.type` ∈ {input, output} solamente) ⇒ medir por spans/logs o atributo propio. **Limitación del estándar, no de un vendor** |
| **D-11** | **Aislar config al spawnear** — todos lo permiten (`CODEX_HOME`, `AMP_SETTINGS_FILE`, `GEMINI_CLI_SYSTEM_SETTINGS_PATH`, `COPILOT_HOME`, `CURSOR_CONFIG_DIR`). ⚠️ **Goose no** (sin `GOOSE_HOME` verificado) |
| **D-12** | **Apagar `statsig` en Codex** — su `metrics_exporter` manda telemetría a OpenAI **por default** |
| **D-13** | **Proxy DESCARTADO con evidencia**: `lite claude`/`lite codex` de LiteLLM funcionan pero **exigen API key de pago ⇒ rompen la auth por suscripción**. Por eso todos los trackers OSS leen disco |
| **D-14** | **Seguir ACP** (Agent Client Protocol, Apache-2.0): `usage_update` con tokens + `cost{amount,currency}`. Candidato creíble a estándar. **AG-UI no sirve** (sin usage) |

## Grupo E · Qué vendorizar

| id | recomendación |
|---|---|
| **E1** | **Catálogo de precios: LiteLLM** (raíz, MIT). 2 984 modelos y **el único con cobertura amplia que publica el tier de 1h**. Actualizado varias veces por semana, precios day-0 |
| **E2** | **Patrón de embebido build-time** de `ccusage` (`go:embed` + `go generate`). Filtrado pesa ~182 KB. Refresco de red **opcional**, fallback al embebido |
| **E3** | ⚠️ **NO heredar el bug de ccusage:** descartan `cache_creation_input_token_cost_above_1hr` (un solo campo `cache_create`) — justo el dato que habilita B1/B12 |
| **E4** | **Esquema `ModelUsage` de Helicone** (Apache-2.0, 549 B) — ya es el superset con opcionales que pide D-4, y **el único que modela `write5m`/`write1h`**. Se porta el esquema, **no los parsers** (parsean HTTP crudo; no proxeamos) |
| **E5** | **OpenInference tiene módulo Go nativo** (Apache-2.0) — **se importa, cero port** |
| **E6** | **Regex de alias Bedrock/Vertex de Langfuse** (MIT) — resuelve la normalización de alias cloud gratis |
| **E7** | **Phoenix DESCARTADO** (Elastic v2: obliga a arrastrar sus términos; y todo lo útil deriva de LiteLLM MIT). **Lunary DESCARTADO** (repo 404, borrado) |
| **E8** | **Los 3 primeros tests están dictados por los bugs ajenos:** (a) olvidar el cache **write**, (b) sumar campos que **se solapan**, (c) **aplanar los tiers** al copiar. Ninguna plataforma tiene el costeo de cache correcto |

## Grupo F · Stack embebido

| id | recomendación |
|---|---|
| **F1** | **`collector/pdata`, no `otlpreceiver`.** +1,7 MB vs +8,4 MB / +60 módulos; `pdata` es **v1.63.0 (API Go estable)**, `otlpreceiver` es **v0.157.0** (config estable, API Go **no**) |
| **F2** | **`modernc.org/sqlite`** (ya en `go.mod`) + **rollup horario incremental**: 1,7 ms vs 323 ms crudo (**190×**). Deja a DuckDB sin nada que ganar |
| **F3** | **p95 por histograma de buckets** (347 ms vs 2 092 ms). `modernc` no tiene `percentile_cont()` |
| **F4** | **`OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` en el spawn** — el default de Claude Code es **gRPC :4317**. Una línea elimina gRPC entero |
| **F5** | ⚠️ **`protojson` NO cumple la spec OTLP/JSON** (trace/span IDs en hex, no base64; hex ⊂ base64 ⇒ puede decodificar basura en silencio). Afecta `/v1/logs`, que es donde CC manda `api_request`. **Confirmar con un test de ~20 líneas antes de comprometerse** |
| **F6** | **`ocb` no sirve** — genera un binario standalone, no una librería |
| **F7** | **Ratifica `sin-cgo`** (ya enforced). DuckDB: 240 MB de descarga en CI, 1,08 GB en disco, MinGW-no-MSVC en Windows. **chDB descartado ×3**: sin Windows, `libchdb.so` dinámica, install `curl \| bash` |
| **F8** | **Prometheus TSDB descartado** (292 módulos; floats-por-serie no absorbe eventos con atributos arbitrarios). **Parquet = archivado frío**, no motor |

## Grupo G · Empaquetado y distribución

| id | recomendación |
|---|---|
| **G1** | **Nivel A o B, nunca C.** (A) compilado en el binario · (B) sidecar en el instalador · (C) descargado en post-install ← **se descarta**: rompe offline/corporativo y **`.dmg` no tiene post-install real** |
| **G2** | **El catálogo de precios se embebe (A) + refresco opcional en background con degradación honesta** — si nunca refresca, funciona con los precios del release y **lo dice** |
| **G3** | 🔴 **`tauri#11992`: la notarización de macOS falla con `externalBin`**, que ya usamos. Issue abierto. **Probar temprano, no en la release** |
| **G4** | 🔴 **Azure Artifact Signing es geo-restringido** (USA/Canadá/UE/UK) ⇒ LATAM probablemente no califica. Plan: **cert OV USD 150–300/año** + clave en HSM cloud |
| **G5** | **No pagar EV** — ya no compra bypass de SmartScreen (Microsoft lo removió en 2024) |
| **G6** | 🟠 La reputación de SmartScreen se acumula **por hash de archivo** ⇒ cada release reinicia el contador. Espaciar releases públicos o usar winget/Store |
| **G7** | **Sin sidecar nuevo**: la telemetría va dentro del `arnesia-daemon` que ya existe |

## Grupo H · Posicionamiento

| id | recomendación |
|---|---|
| **H1** | **No competir en cobertura de runtimes.** La categoría «tracker multi-CLI» está saturada y es MIT (ccusage 17,4k ★ / 15 runtimes; tokscale 40+). Dynatrace y Azure ya venden el dashboard |
| **H2** | 🎯 **El hueco real: nadie correlaciona el uso con QUÉ COMPONENTE / ROL / PROCESO estaba corriendo.** Todos agregan por herramienta, modelo, proyecto y día. **La atribución por unidad de trabajo es terreno libre** ⇒ el diseño optimiza por el eje **arnés × empresa × puesto**, no por cobertura |
| **H3** | **Evaluar `api_request` (log event) contra las métricas** como canal primario: trae `cost_usd_micros` + los 4 buckets + `model`/`speed`/`effort`/`query_source` + atribución, **por request**, sin temporalidad delta ni cardinalidad |
| **H4** | **`claude_code.cost.usage` es "Estimated cost"**, no facturación. La UI no puede presentarlo como plata gastada sin decirlo |

---

# Parte 2 — La arquitectura, en capas

```
L0  CONTRATO      arnes.l0.json declara `telemetria:` — enforced por `arnesia conformance`
                  ↓ (sin esto, el arnés no sella ni publica)
L1  EMISIÓN       (a) runtime nativo: OTel | stream-json      ← ArnesIA controla el spawn
                  (b) hooks DEL ARNÉS: eventos de proceso     ← viajan sin ArnesIA (fail-open)
                  (c) eventos propios del daemon: rotación, gate, run
                  ↓
L2  INGESTA       adapters/telemetria/otlp/       (handler net/http + pdata)
                  adapters/telemetria/streamjson/ (parser por runtime)
                  adapters/telemetria/hooks/      (endpoint de eventos de proceso)
                  → ports.TelemetrySink → domain.EventoTelemetria   (UN solo evento canónico)
                  ↓
L3  NORMALIZACIÓN registro por runtime: usage_arithmetic · regla de acumulación · mapa de namespaces
                  ↓
L4  COSTEO        catálogo LiteLLM embebido (go:embed) + calculadora + los 3 tests de E8
                  ↓
L5  ALMACÉN       modernc.org/sqlite — eventos crudos + rollup horario incremental
                  ↓
L6  DETECCIÓN     13 detectores de fuga de dinero + los de fuga de proceso
                  cada adaptador DECLARA cuáles soporta → degradación honesta
                  ↓
L7  SUPERFICIE    capa Tokens del Mapa · tarjeta del proyecto en Portafolio · export opcional OTLP
```

**La llave de atribución atraviesa todas las capas:** `(arnés, instalación, caja/paso, sesión)` —
coherente con la identidad `(home, id)` del Portafolio y con el modelo de terreno.

## Por qué esta forma y no otra

- **Un solo evento canónico (L2)** porque hay dos orígenes heterogéneos. Eso **degrada `pdata` de
  modelo de datos a decodificador de wire format** — y por eso conviene el decodificador barato y
  estable (F1), no el framework completo.
- **L3 separada de L2** porque las reglas de aritmética/acumulación son **dato de tabla por runtime**
  (D-6, D-7), no lógica dispersa en cada parser.
- **L4 separada de L3** porque el costeo falla de tres formas conocidas (E8) y necesita sus propios
  tests, independientes del parseo.
- **L6 declara capacidades** porque los detectores son per-runtime (D-5). Un detector que no aplica
  se muestra «no aplica», nunca 0.

---

# Parte 3 — El mecanismo de OBLIGACIÓN

> *«mecanismos para que los arneses que nosotros creemos, les podamos obligar a tener algún
> mecanismo de medición para que cuando agreguemos un proyecto podamos ver puntos de mejora»*

## El principio ya está firmado, falta el mecanismo

`vision.md` p9 + este mismo boundary: **«todo arnés nace observable — observability by default, no
opt-in»**. Lo que falta no es la doctrina, es **el enforcement**.

## La obligación NO se logra en runtime — se logra en el SELLO

Un arnés no se puede «obligar» a emitir mientras corre; se lo obliga **a nacer instrumentado**, y eso
se chequea antes de publicar. El repo ya tiene el motor: **`arnesia conformance`** + la doctrina
`codigo-traza-a-capability` (R1 integridad · R2 cobertura · R3 gate-commit · R4 estado-generado, con
rollout warn→block). **Se replica el mismo patrón:**

| check | qué exige | severidad |
|---|---|---|
| `arnes-declara-telemetria` | `arnes.l0.json` tiene bloque `telemetria: {version, eventos[], atribucion}` | error |
| `arnes-porta-hook-proceso` | el paquete shipea el hook de eventos de proceso del kit | error |
| `hook-es-fail-open` | el hook no rompe ni bloquea si el daemon no está | error |
| `telemetria-no-egresa` | el hook solo escribe a loopback / disco local, nunca a red externa | error |

**Sin esos checks, el arnés no sella ⇒ no publica.** Esa es la obligación: **un gate, no un truco de
runtime.** Y es coherente con lo que ya se hace — el sello `arnes.l0.json` ya es la puerta.

## Las tres señales, y por qué hacen falta las tres

| señal | quién la emite | qué contesta | ¿sobrevive fuera de ArnesIA? |
|---|---|---|---|
| **tokens/costo** | el runtime (OTel / stream-json) | *cuánto costó* | solo si alguien inyecta las env vars |
| **proceso** | **hooks del propio arnés** | *qué paso, qué gate, cuánto tardó, qué se rechazó* | **sí** — el hook viaja dentro del paquete |
| **decisión** | el daemon de ArnesIA | *rotación de contexto, corrida de caja, gate humano* | no — es nuestro |

⚠️ **Verificado: los payloads de hooks NO traen tokens, costo ni datos de cache.** Traen
`transcript_path`, `session_id`, `prompt_id`, y `PreCompact`/`SessionStart` permiten detectar
fronteras de compactación. ⇒ **el hook es la señal de PROCESO, jamás la de dinero.**

**El JOIN es el producto.** Ni los tokens solos ni el proceso solo dan un «punto de mejora». Juntos
sí: *«este arnés, en este puesto, quema $X — y el 60 % se va en la caja Y, que falla el gate 3 de
cada 4 veces»*. La llave del join es `session.id` + los resource attributes que inyectamos.

**Eso es exactamente el hueco de H2 que nadie llenó.**

## Cómo se enciende sin pedirle nada al usuario

1. **Al crear el arnés** (`scaffold`): estampa el bloque `telemetria:` en el sello + copia el hook
   del kit. Nace instrumentado.
2. **Al spawnear desde ArnesIA**: el daemon inyecta `OTEL_*` + `OTEL_RESOURCE_ATTRIBUTES` con la
   llave de atribución. Cero cooperación del arnés.
3. **Al correr fuera del spawn**: el hook busca el daemon local (archivo de descubrimiento de puerto)
   y postea. Si no lo encuentra, **fail-open silencioso** — nunca rompe el trabajo del usuario.
4. **Al agregar el proyecto al Portafolio**: la tarjeta del proyecto muestra los puntos de mejora
   acumulados de ese arnés en ese contexto.

**Prior art a reusar (no portar):** `emit.py`/KIT-03 ya resolvió el diseño fail-open y el walk de
privacidad («ningún campo `egreso: sensible` cruza el borde sin verificar»). El **concepto** sirve;
el **canal** (parsear JSONL + Langfuse) no.

## Alcance — CERRADO por D12.1: S1 + S2, máquina propia

| escenario | ¿controlamos el spawn? | mecanismo | estado |
|---|---|---|---|
| **S1** — arnés dentro del chat dock de ArnesIA | **sí** | inyección de env al spawn | ✅ **en alcance** |
| **S2** — arnés en la misma máquina, fuera de ArnesIA (terminal/IDE) | no | **hook + descubrimiento del daemon local** | ✅ **en alcance** |
| **S3** — arnés en máquina de un cliente sin ArnesIA | no | requeriría egreso de red | ❌ **FUERA** — choca con `superficie-local-confinada`; decisión de producto/legal, paquete propio si se reabre |

**Todo en `127.0.0.1`. Cero egreso de red.** Como S2 está en alcance, **el mecanismo de obligación
se construye** — es el único caso donde la instrumentación tiene que viajar dentro del arnés.

Tres requisitos que S2 agrega:
1. **archivo de descubrimiento de puerto** (el hook encuentra el daemon sin que nadie le pase nada),
2. **fail-open silencioso obligatorio** (si el daemon no está, jamás rompe ni demora al usuario),
3. **`telemetria-no-egresa` como check duro** (solo loopback o disco local).

## Forma del entregable — CERRADO por D12.2: el MVP es el JOIN

No se entrega tokens solo ni proceso solo. El MVP es la correlación:

> *«este arnés, en este puesto, quema $X — y el 60 % se va en la caja Y, que falla el gate 3 de cada
> 4 veces»*

⚠️ **Expande el alcance original del paquete** (ver D12.3): «capa Tokens» ya no describe el
entregable — es una capa de **mejora** que junta dinero + proceso, y arrastra parte de lo que el
BACKLOG tenía como capa Proceso en `bloqueo`. Renombrar en el mockup.

🟢 **Pero el MVP del join NO depende de D8** (D12.4): el split 5m/1h bloquea solo los detectores
**B1** y **B12**; los otros 11 + todo el eje de proceso no lo necesitan. D8 se firma después, sin
frenar el paquete.
