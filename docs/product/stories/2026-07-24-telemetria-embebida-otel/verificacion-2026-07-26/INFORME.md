# Verificación EN VIVO de la telemetría — 2026-07-26

> **Qué es esto:** las tres verificaciones que el paquete listaba como «pendientes antes del spec»
> corridas de verdad contra el binario real (`claude 2.1.220`), no contra la documentación.
> Doctrina de honestidad del repo: *nada se da por «listo» sin verificar en vivo*.
>
> **Método:** receptor OTLP casero de 50 líneas ([`otlpsink.go.txt`](otlpsink.go.txt)) escuchando
> `127.0.0.1:4318`, tres corridas de `claude -p` con las env vars de telemetría apuntando ahí.
> Payloads crudos en [`evidencia/`](evidencia/) (**PII redactada** — ver hallazgo V6).
> Costo total del experimento: **USD 0,044** (3 corridas, `claude-haiku-4-5`).

| corrida | protocolo OTLP | salida | qué probaba |
|---|---|---|---|
| 1 | `http/json` | `--output-format json` | qué señales llegan · si `OTEL_RESOURCE_ATTRIBUTES` sirve |
| 2 | `http/json` | `--output-format json`, invocando una **skill de plugin** (`/caveman-help`) | la redacción `third-party` de D10 |
| 3 | `http/protobuf` | `--output-format stream-json` | wire format binario · el split 5m/1h de D8 |

---

## V1 · `api_request` ES el canal correcto — H3 confirmado

El log event `claude_code.api_request` llegó con **todo lo que hace falta, por request**:

```
event.name        = api_request          model             = claude-haiku-4-5-20251001
prompt.id         = 8e957894-…           input_tokens      = 10
session.id        = 89c8bdde-…           output_tokens     = 39
request_id        = req_011CdQDvJP…      cache_read_tokens = 17536
duration_ms       = 1772                 cache_creation_tokens = 8257
speed             = normal               cost_usd          = 0.0184726
query_source      = sdk                  cost_usd_micros   = 18473
```

Contra la métrica `claude_code.token.usage`, que llega troceada en 4 puntos (`type` ∈
`input`/`output`/`cacheRead`/`cacheCreation`) y hay que recomponer.

⇒ **El canal primario es `/v1/logs` (`api_request`), no `/v1/metrics`.** Las métricas quedan como
señal secundaria. Esto **cierra la decisión (a)** que quedaba abierta y **corrige el diseño de D4**,
que asumía `POST /v1/metrics` como el endpoint del receptor: **hay que recibir los dos, y el que
importa es el de logs.**

Inventario completo de eventos recibidos en una corrida trivial (57 log records):

| evento | qué habilita |
|---|---|
| `api_request` | **dinero + latencia por request** |
| `hook_execution_start` / `_complete` | `total_duration_ms`, `num_blocking`, `num_success`, `num_cancelled` → **señal de proceso, sin hooks propios** |
| `skill_activated` | `invocation_trigger`, `skill.source`, ligado por `prompt.id` |
| `plugin_loaded` | `plugin_id_hash` (ver V3) |
| `mcp_server_connection` | `status`, `duration_ms`, `transport_type` |
| `user_prompt` / `assistant_response` | longitudes; el contenido llega **`<REDACTED>`** por default |
| `tool_*` | no aparecieron (la corrida no usó herramientas) — verificar aparte |

## V2 · D8 SE DISUELVE — el split 5m/1h no obliga a tocar el JSONL

D8 planteaba un dilema binario: **(a)** enmendar `conductor-no-parsea-jsonl.md` para leer 4 campos
del transcript, o **(b)** resignar los detectores B1 y B12. **Hay una tercera vía y estaba a la
vista:** el `result` del **stream-json** lo trae.

```json
"usage": { "input_tokens": 10, "cache_read_input_tokens": 21695,
           "cache_creation_input_tokens": 4099, "output_tokens": 35,
           "cache_creation": { "ephemeral_1h_input_tokens": 4099,
                               "ephemeral_5m_input_tokens": 0 },
           "service_tier": "standard", "speed": "standard",
           "iterations": [ { … mismo desglose por iteración … } ] }
```

Verificado con `--output-format stream-json --verbose` (corrida 3) **y** con `--output-format json`
(corrida 1). El boundary prohíbe parsear **el transcript JSONL**; el stream-json del subproceso es
el canal **sancionado** — el árbol ya lo consume y `arch_test.go` lo tiene cableado
(`TestLiveEventsFromStreamJSON`).

⇒ **B1 (re-warm por TTL) y B12 (watchdog) quedan desbloqueados sin enmendar ningún boundary.**
D8 no se firma: **se cierra por inexistencia del dilema.**

Bonus del mismo envelope: `modelUsage` por modelo con `costUSD`, `contextWindow`, `canonicalModel`
y `provider`; `iterations[]` con el desglose por iteración; y eventos `rate_limit_event` /
`thinking_tokens` en el stream.

## V3 · D10 confirmado con dato real — y apareció la mitigación que faltaba

**La redacción es real y pega exactamente donde duele.** Con una skill de plugin activa, el
`api_request` de esa corrida llegó así:

```
skill.name  = third-party        (en skill_activated llega como `custom_skill`)
plugin.name = third-party
marketplace.name = third-party
```

**Pero `plugin_loaded` trae un identificador que la investigación no había visto:**

| corrida | `plugin.name` | `plugin_id_hash` | `plugin.scope` |
|---|---|---|---|
| 1 | `third-party` | `083a6411f71e6967` | user-local |
| 1 | `third-party` | `29e6a51f5b6dca52` | user-local |
| 1 | `rust-analyzer-lsp` | `5de2da3204573e62` | **official** |
| 2 | `third-party` | `083a6411f71e6967` | user-local |
| 2 | `third-party` | `29e6a51f5b6dca52` | user-local |

**`plugin_id_hash` es estable entre sesiones y distinto por plugin.** Los plugins oficiales
conservan su nombre real; solo los de terceros se redactan — y aun así quedan **distinguibles**.

⇒ **La atribución por ARNÉS es recuperable** con una tabla local `hash → arnés`, que podemos
construir porque **nosotros instalamos el arnés**. Importa sobre todo en **S2** (el usuario corre el
arnés fuera de ArnesIA, donde no inyectamos env vars): ahí el hash es la única señal de identidad.

⚠️ **Lo que NO se recupera:** el nombre de la **skill** dentro del arnés. `skill_activated` no trae
hash. Lo máximo es unir `skill_activated` → `api_request` por `prompt.id` y atribuir **el turno**.
⇒ **La granularidad honesta del producto es `arnés × caja × sesión × turno`, no por-skill.** Es
exactamente el eje del terreno propio que D7.2 pedía — la limitación empuja al diseño correcto.

⚠️ **No verificado:** si `plugin_id_hash` es determinista **entre máquinas**. Si no lo fuera, el
mapeo se aprende localmente igual (spawn controlado una vez al instalar), pero cambia el diseño.

## V4 · `OTEL_RESOURCE_ATTRIBUTES` funciona, y mejor de lo esperado — D9.4 confirmado

Se inyectó `arnesia.arnes=vitalia,arnesia.caja=paso-3,arnesia.instalacion=home-local` y aparece:

1. en los **resource attributes**, y
2. **copiado dentro de CADA data point de métrica y CADA log record.**

Eso significa que el join no depende de correlacionar resource↔punto: **cada `api_request` ya viene
etiquetado con la unidad de trabajo.** La llave canónica del terreno viaja sola.

## V5 · F1 REFUTADO — `pdata` cuesta 6× lo que dice la investigación

Medido, mismo compilador, misma base (`net/http` + `modernc.org/sqlite`, o sea nuestra dependencia real):

| binario | tamaño | delta |
|---|---|---:|
| base | 10,77 MB | — |
| base + **decodificador OTLP/JSON con la stdlib** | 11,26 MB | **+0,49 MB** |
| base + **`collector/pdata`** (`pmetricotlp` + `plogotlp`) | 21,56 MB | **+10,79 MB** |

Referencia: el daemon hoy pesa **22,71 MB** (16,93 MB con `-s -w`). `pdata` lo llevaría a ~33,5 MB
(**+48 %**), en un producto cuyo argumento de venta es «se instala y ya».

**F1 decía «+1,7 MB».** No reproduce.

**Y `protojson` ni siquiera es una alternativa:** `pdata` no expone los `proto.Message` de OTLP, así
que usar `protojson` obliga a vendorizar los `.proto`. La disyuntiva real no era la que planteaba F5.

⇒ **Contrapropuesta medida: OTLP/JSON + `encoding/json`.** Podemos hacerlo porque **controlamos el
spawn** (D12.1 S1) y el hook es nuestro (S2) ⇒ imponemos
`OTEL_EXPORTER_OTLP_PROTOCOL=http/json`. El decodificador completo son ~120 líneas
([`decodificador-stdlib.go.txt`](decodificador-stdlib.go.txt)), probado contra los payloads reales:
**8 métricas · 14 puntos · 57 log records, todos los eventos**.

⚠️ **Gotcha verificado, va al spec:** Claude Code emite `intValue` como **número JSON**, no como
string — **off-spec** (el mapeo protobuf→JSON manda string para int64). Un decodificador estricto
revienta con `cannot unmarshal number into Go struct field … of type string`. Se resuelve con
`json.Number`, que acepta las dos formas. **`pdata` sí lo tolera** — es su único mérito medido acá,
y no vale 10,3 MB.

**Lo que `pdata` sí resuelve bien** (probado): decodifica el protobuf real de Claude Code (27 log
records, 7 puntos) y acepta trace IDs en hex en OTLP/JSON. Queda como plan B si algún runtime
futuro no deja elegir el protocolo.

### V5.1 · Temporalidad — dato que faltaba para el receptor

Las 4 métricas llegan `Sum · temporalidad=Delta · monotonic=true`.

⇒ **El receptor NO tiene que diferenciar contadores acumulados**: suma y listo. Elimina toda una
clase de bug (reinicio de proceso = contador que vuelve a cero) que habría aparecido en producción.

## V6 · 🔴 HALLAZGO NUEVO — la telemetría arrastra PII en cada punto

**Ninguna decisión del paquete lo contemplaba.** Cada data point de métrica y cada log record llega con:

```
user.email        = <el email real de la cuenta>
user.account_uuid = <uuid de cuenta>          user.account_id = user_01…
organization.id   = <uuid de la organización> user.id = <hash sha256>
```

Consecuencias que hay que resolver **antes** de escribir el receptor:

1. **`telemetria-no-egresa` deja de ser higiene y pasa a ser protección de datos.** El escape hatch
   C2 (forward OTLP al Langfuse del operador) **exportaría el email y los ids de cuenta** de quien
   corra el arnés. Si algún día se enciende, tiene que **filtrar en el borde**, no reenviar crudo.
2. **Hay que decidir qué se PERSISTE.** Para el join no hace falta ninguno de esos campos:
   `session.id` + los `arnesia.*` alcanzan. La postura correcta es **allowlist** (guardar solo lo
   declarado), no denylist.
3. **Es un argumento a favor del canal OTLP y en contra de leer disco:** el contenido de prompts y
   respuestas llega **`<REDACTED>`** por defecto. La telemetría es menos invasiva que el transcript.

Lo verificado acá se aplicó a esta misma evidencia: los payloads de [`evidencia/`](evidencia/) están
redactados con el criterio de allowlist antes de entrar a git.

## V7 · Lo que NO se pudo verificar

| # | qué | por qué | cómo cerrarlo |
|---|---|---|---|
| 1 | **`tauri#11992`** — notarización macOS con `externalBin` | no hay máquina macOS | runner macOS en CI, o una prueba manual antes de abrir el canal |
| 2 | Determinismo de `plugin_id_hash` **entre máquinas** | una sola máquina | correr el mismo plugin en otra máquina/usuario |
| 3 | Eventos `tool_*` (`tool_decision`, `tool_result`) | la corrida no usó herramientas | una corrida con `--allowedTools Read` |
| 4 | Comportamiento con **subagentes** (`isSidechain`) y con `/compact` | no ejercitado | corrida larga con subagente + compactación forzada |
| 5 | Los otros 5 runtimes | solo se probó Claude Code | repetir el mismo receptor por runtime |

---

## Qué cambia en el paquete

| decisión previa | qué le pasa |
|---|---|
| **D4.1** «receptor decodifica `POST /v1/metrics`» | **corregido**: el canal primario es `/v1/logs` (`api_request`); metrics es secundario |
| **D8** «leer el JSONL o resignar B1/B12» | **disuelta** — el split viene en el `result` del stream-json (V2) |
| **D9.3 / F1** «`pdata`, +1,7 MB» | **refutado** (+10,79 MB medido) → **OTLP/JSON + stdlib, +0,49 MB** (V5) |
| **D9.4** «`OTEL_RESOURCE_ATTRIBUTES` es el vector» | **confirmado y reforzado** — se copia a cada punto (V4) |
| **D10** «la atribución por componente está amenazada» | **confirmado**, y **parcialmente rescatado** por `plugin_id_hash` (V3) |
| **F4** «forzar `http/protobuf` en el spawn» | **invertido**: forzar **`http/json`**, que es lo que hace barato el decodificador |
| **H3** «evaluar `api_request` como canal primario» | **resuelto a favor** (V1) |
| — | **nuevo**: PII en cada punto (V6) · temporalidad Delta (V5.1) · `intValue` off-spec (V5) |
