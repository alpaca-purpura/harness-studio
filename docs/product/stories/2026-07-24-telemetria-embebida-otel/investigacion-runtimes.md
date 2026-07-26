# Investigación — señal de telemetría por runtime de agente (2026-07-26)

> Insumo para D7 (multi-runtime). Recolectado por investigación externa verificada contra repos y
> doc oficial. **`[V]` = verificado leyendo la fuente · `[?]` = NO verificado, no usar como hecho.**
> Estado: PARCIAL — faltan OpenCode, Cline, Goose, Aider, Crush + estado 2026 de semconv GenAI.

## Tabla comparativa

| runtime | OTel útil en **headless** | namespace | `usage` en stream-json | costo USD |
|---|---|---|---|---|
| Claude Code | ✅ | `claude_code.*` | ✅ (`result` acumulado — trampa) | ✅ **único** |
| Gemini CLI | ✅ | híbrido `gemini_cli.*` + `gen_ai.*` | ✅ tokens | ❌ |
| Qwen Code | ✅ | `qwen-code.*` (guion) + `gen_ai.*` | ✅ (5 campos propios) | ❌ |
| Codex | ❌ **`exec` sin métricas** | híbrido `codex.*` + attrs `gen_ai.*` | ✅ `turn.completed.usage` | ❌ |
| Amp | ❌ sin OTel | — | ✅ **el más rico (split 5m/1h)** | ❌ en el stream |
| Cursor | ❌ sin OTel | — | ❌ sin usage | ❌ |

**stream-json: 5/6 · OTel útil en headless: 3/6.** El canal por-turno estructurado es más universal
que OTel. Invierte la jerarquía asumida en D4: **adaptador base = stream-json; OTel = enriquecedor**
(latencia, reintentos, atribución por skill/MCP) donde exista.

## Los tres hallazgos que restringen el esquema normalizado

### 1. La aritmética de tokens NO es uniforme entre proveedores `[V]`

| proveedor | relación |
|---|---|
| **Anthropic** | `input_tokens` **excluye** los cacheados — `cache_read` es campo aparte, aditivo |
| **OpenAI/Codex** | `cached_input_tokens` es **subconjunto de** `input_tokens` — sumarlos duplica |

Un esquema que trate ambos igual sobrecuenta o subcuenta **según el proveedor**. Es la clase de bug
de `langfuse#12306` (reportaba ~50 % de hit rate donde lo real era ~99 %).
⇒ **El esquema normalizado debe declarar, por proveedor, si los campos son disjuntos o anidados.**
Es dato de tabla, no lógica de código.

### 2. Trampa de acumulación, una variante por runtime `[V]`

- **Claude Code**: el `usage` del frame `result` es acumulado (`cache_read` re-contado por tool-call)
  — *"sobreestima brutalmente (100 % espurio medido en vivo)"*, ya documentado en `conductor.go`.
- **Codex**: `last_token_usage` se **re-emite sin cambios** en updates de rate-limit → sumarlo
  sobrecuenta. Correcto = **diferenciar `total_token_usage` sucesivos**.
- **Amp**: `usage` es per-message en los eventos `assistant` → hay que acumularlo (opuesto).

Tres runtimes, tres reglas distintas de agregación. **La regla de acumulación es propiedad del
adaptador, nunca del agregador central.**

### 3. Nadie adopta `gen_ai.*` puro `[V]`

Los tres runtimes con OTel usan namespace **híbrido**: nombre propio + atributos `gen_ai.*` donde
existe semconv, propios donde no. Y Gemini expone **tres vocabularios de token incompatibles dentro
del mismo producto** (métrica `input/output/thought/cache/tool` · stats JSON
`input/prompt/candidates/total/cached/thoughts/tool` · `GenAIUsageDetails` con `*_token_count`).

## Contrato del adaptador (derivado, no uniforme)

Tres mecanismos, **no uno**:

| pieza | Claude Code | Gemini / Qwen | Codex | Amp | Cursor |
|---|---|---|---|---|---|
| aislar config | (ya resuelto, `superficie-local-confinada`) | `GEMINI_CLI_SYSTEM_SETTINGS_PATH` · `QWEN_CODE_SYSTEM_SETTINGS_PATH` (+`_DEFAULTS_`) | `CODEX_HOME` + `--ignore-user-config` + `--ephemeral` | `AMP_SETTINGS_FILE` / `--settings-file` | `CURSOR_CONFIG_DIR` |
| activar señal | env `CLAUDE_CODE_ENABLE_TELEMETRY` + `OTEL_EXPORTER_OTLP_ENDPOINT` | env `GEMINI_TELEMETRY_*` / `QWEN_TELEMETRY_*` (**los flags `--telemetry*` ya no existen en Gemini**) | **escribir bloque `[otel]` en `$CODEX_HOME/config.toml` o `-c`** — no lee `OTEL_EXPORTER_OTLP_*` | n/a | n/a |

**El aislamiento de config SÍ es universal. La activación NO.** Codex obliga a escribir un archivo
de config aislado — el adaptador necesita ese tercer mecanismo.

## Notas por runtime

### Codex `[V]`
- OTel nativo real (crate Rust `codex-rs/otel/`), opt-in (`exporter = "none"` default).
- **Gap crítico (issue #12913):** `codex exec` = logs+traces, **sin métricas**; `codex mcp-server`
  no inicializa OTel. Solo el TUI tiene las tres señales. `exec` es justamente el modo que usaríamos.
- ⚠️ **Privacidad:** `metrics_exporter` default = **`statsig`** (telemetría hacia OpenAI). Todo arnés
  Codex que forjemos debe apagarlo explícitamente.
- `[otel]` bloqueado desde el `.codex/config.toml` de proyecto ("telemetry routing keys").
- Fuente buena de tokens: `codex exec --json` → `turn.completed.usage`
  `{input_tokens, cached_input_tokens, output_tokens}`.
- Rollouts en `$CODEX_HOME/sessions/<YYYY>/<MM>/<DD>/rollout-*.jsonl`; schema **inestable**
  (hay backend SQLite paralelo; los parsers de terceros soportan 3 generaciones de formato).

### Amp `[V]`
- Sin OTel. Pero su `--stream-json` publica schema en TypeScript y trae
  `usage.cache_creation.{ephemeral_5m,ephemeral_1h}_input_tokens` — **el mismo split que la
  telemetría OTel de Claude Code NO expone** (ver D8).
- Compatibilidad con Claude Code **explícita** ⇒ un solo parser cubre ambos.
- Costo USD existe en el producto (`amp usage`, setting `amp.showCosts`) pero **no viaja en el stream**.

### Gemini CLI `[V]`
- Métrica `gemini_cli.token.usage`, attr `type` ∈ `input`·`output`·**`thought`**·`cache`·`tool`.
- Precedencia **env > settings**; endpoint 127.0.0.1 arbitrario OK; target solo `local`|`gcp`.
- Headless `-o json` → `stats.models[].tokens` con tokens, **sin costo**.

### Qwen Code `[V]`
- No es fork trivial: prefijo con guion, **spans propios declarados** (Gemini no tiene ninguno
  verificable), y añade `gen_ai.usage.cache_read.input_tokens` / `cache_creation.input_tokens`.
- Adaptador aparte, no un alias de Gemini.

### Cursor CLI `[V]`
- El peor caso: sin OTel, binario cerrado, `result` de `--output-format stream-json` **sin `usage`
  ni costo**. Fallback = Admin API (`POST /teams/filtered-usage-events`) — batch por equipo, requiere
  plan team/enterprise ⇒ **inútil para atribución por corrida**.

## NO verificado — no usar como hecho

- `usage` en el `result` de `cursor-agent` (staff lo afirmó en foro 27-feb-2026; la doc lo contradice
  — requiere correr el binario).
- Nombres de span de Gemini CLI · schema completo de `~/.gemini/tmp/<id>/logs.json` · `GEMINI_DIR`
  como env var · valores del attr `type` en `qwen-code.token.usage`.
- `AMP_URL` · `amp.telemetry.enabled` (nombre probablemente incorrecto; no existe en el manual).
- Rutas de sesión de Cursor (`~/.cursor/chats/` SQLite) y de Amp
  (`~/.local/share/amp/threads/`) — ambas de terceros.
- Nombres de span de Codex · que `reasoning_output_tokens`/`total_tokens` aparezcan en
  `turn.completed.usage`.
- Env vars `OTEL_EXPORTER_OTLP_*` en Codex: no figuran en doc ni en `otel/src/config.rs`
  ⇒ asumir que **no** funcionan.

## Segunda tanda (2026-07-26) — 7 runtimes más

**Corrección al alcance:** faltaba **GitHub Copilot CLI** en la lista original, y resulta de los
mejores del lote.

| runtime | OTel útil headless | namespace | fallback en disco | costo USD | dificultad |
|---|---|---|---|---|---|
| **Copilot CLI** | ✅ | **`gen_ai.*` + `copilot_chat.*`** — el más limpio | file exporter JSONL `~/.copilot/otel/*.jsonl` | ❌ | **2** |
| **OpenCode** | ⚠️ logs+traces, **sin métricas** | `opencode` hardcodeado | SQLite `opencode.db` | campo existe, **suele venir 0** | **2** |
| **Cline** | ✅ métricas+logs, sin traces | **propio, cero `gen_ai`** | `ui_messages.json` (usage como **string JSON dentro de `text`**) | ✅ | **2** |
| **Goose** | ✅ (feature compilada por default) | propio | SQLite `sessions.db` **schema v13 migrable** | columna existe, en práctica vacía | **3** |
| **Aider** | ❌ (PostHog/Mixpanel) | — | **`--analytics-log` JSONL** evento `message_send` | ✅ | **2** |
| **Crush** | ❌ (PostHog) | — | SQLite `crush.db` — **solo por sesión, sin cache ni reasoning** | columna `cost` | **3** |
| **Factory Droid** | no verificado | no verificado | `droid exec --output-format stream-jsonrpc` | no verificado | **3** |

**Taxonomía de tokens — mejor y peor:**
- **Mejor:** OpenCode `tokens {input, output, reasoning, cache {read, write}}` + `cost`, con
  granularidad **por paso** (`StepFinishPart`). Gemini `input/output/thought/cache/tool`.
- **Peor:** Crush y Aider (solo prompt/completion).
- **Ninguno excepto Amp** expone el split `ephemeral_5m`/`ephemeral_1h`.

**Detalles operativos que importan:**
- **Copilot CLI**: se auto-activa con solo definir `OTEL_EXPORTER_OTLP_ENDPOINT`; aísla con
  `COPILOT_HOME`. **Solo `otlp-http`, no gRPC.**
- **Goose**: usa las **env vars OTel estándar** ⇒ el más fácil de encender. Pero ⚠️ **no hay forma
  verificada de aislar su config** — no existe `GOOSE_CONFIG_PATH`/`GOOSE_HOME`/`--config`. Si lo
  spawneamos, **comparte la config global del usuario**. Además `--no-session` manda la sesión a
  `/dev/null`.
- **Cline**: env con prefijo propio `CLINE_OTEL_*` (no las `OTEL_*` estándar) y **pisan la config
  del dashboard enterprise**. Tiene **tres nomenclaturas distintas para lo mismo** según la capa
  (disco / SDK / OTel).
- **Aider**: la vía sana es `--analytics-log eventos.jsonl --no-analytics` — escribe local **sin
  mandar nada a la nube**. `--config <archivo>` saltea la búsqueda de `.aider.conf.yml`.
  **No parsear `.aider.chat.history.md`** (volcado de consola, frágil).
- **OpenCode**: bug a no repetir — `process.exit()` mata el `BatchSpanProcessor` antes del flush
  ⇒ spans perdidos (issue #13438).

## Estado 2026 de las convenciones GenAI de OTel `[V]`

**Se mudaron a repo propio `open-telemetry/semantic-conventions-genai`** (v1.42.0, jun-2026). La
mudanza fue **organizativa por cadencia de releases, NO una graduación a estable**. La página vieja
redirige y los `gen_ai.*` del registry principal figuran como *Deprecated / movidos*.

**Estabilidad: TODO en `Development`.** Ni una métrica ni un atributo `gen_ai.*` llegó a `Stable`.

### 🔴 El hallazgo que decide el canal de ingesta

**`gen_ai.token.type` admite EXCLUSIVAMENTE `input` y `output`.** No hay valor para cache ni para
reasoning. Esos conceptos existen **solo como atributos de SPAN**, nivel *Recomendado*:

```
gen_ai.usage.cache_creation.input_tokens   — escritos a un cache del proveedor
gen_ai.usage.cache_read.input_tokens       — servidos desde el cache
gen_ai.usage.reasoning.output_tokens       — usados para razonamiento
```

⇒ **Medir por MÉTRICAS OTel pierde cache y reasoning. Hay que medir por spans/logs, o inventar
atributo propio.** Esto es coherente con lo que ya sabíamos de Claude Code (el split 5m/1h no está
en las métricas) — pero ahora se sabe que **es limitación del estándar, no de un vendor**.

### Costo en dinero: NO existe en la spec `[V]`

`gen_ai.usage.cost_usd` circula como convención de facto de gateways, no es estándar. La
recomendación explícita es calcularlo desde tabla de precios propia. **Quinta confirmación** de que
el catálogo de precios es obligatorio.

### Convención de agentes: sí existe, pero incompleta para nosotros `[V]`

`gen_ai.agent.{id,name,description,version}` · `gen_ai.conversation.id` ·
`gen_ai.tool.{name,type,description,call.id,call.arguments,call.result,definitions}`.
Operaciones: `chat`, `create_agent`, `invoke_agent`, `execute_tool`, `invoke_workflow`, …

**No hay nada para «sub-agente», «skill» ni «plugin»** ⇒ ese vocabulario es namespace propio de
ArnesIA. No estamos reinventando: el estándar no lo cubre.

## 🔑 `OTEL_RESOURCE_ATTRIBUTES` — el vector universal de atribución

Lo respetan **Claude Code, Copilot CLI, Codex, Goose y OpenCode** `[V]`. Es la palanca para inyectar
**nuestro** identificador de unidad de trabajo (arnés · caja · sesión) al spawnear, sin depender del
vocabulario de cada runtime.

**Mitiga el problema de la redacción «third-party»** documentado en
[`investigacion-plataformas.md`](investigacion-plataformas.md): si `skill.name`/`plugin.name` se
colapsan, la atribución igual viaja por resource attributes.

⚠️ **Pero no lo reemplaza del todo:** los resource attributes se fijan **por proceso**. Dan
atribución a granularidad de proceso spawneado (un arnés/caja por proceso), **no** por-skill dentro
de una misma sesión. Si la capa Tokens quiere el número **por nodo del Mapa**, hace falta o bien un
proceso por unidad, o bien recuperar `skill.name` sin redactar. **Pregunta abierta para el mockup.**

## La categoría «tracker multi-CLI» está saturada — y es MIT

| proyecto | lic. | lenguaje | runtimes | costo USD | actividad |
|---|---|---|---|---|---|
| **ccusage** | MIT | **Rust** | **15** | sí, pricing de **LiteLLM** | push 2026-07-26, **17,4k ★** |
| **tokscale** | MIT | Rust+TS | **40+** (suma Cline, Roo, Cursor, Devin, Zed) | sí, LiteLLM+OpenRouter | push 2026-07-25, 4,6k ★ |
| TokenTracker | MIT | JS | 27-28 | sí, 2 200+ modelos offline | 1,1k ★ |
| caut | NOASSERTION | Rust | 16+ | solo Claude/Codex | enfocado en **cuota restante** |
| vibe-coding-dashboards (VictoriaMetrics) | Apache-2.0 | Grafana | 4 | ❌ | único dashboard OSS multi-CLI vía OTel |

Comercial ya llegó: **Dynatrace** monitorea Claude Code/Gemini/Codex/OpenCode/Copilot vía OTel;
**Azure Monitor** ofrece "single pane of glass" para seis agentes.

### 🎯 El hueco que nadie llenó

> **Ninguno correlaciona el uso con QUÉ COMPONENTE / ROL / PROCESO estaba corriendo.** Todos agregan
> por herramienta, modelo, proyecto y día. **La atribución por unidad de trabajo es terreno libre.**

Confirma la tesis del producto: el eje **arnés × empresa × puesto** no lo tiene nadie, y no es la
métrica de tokens lo que nos diferencia — es el eje de atribución.

## Estrategia proxy: descartada con evidencia

**LiteLLM** trae el comando `lite` que lanza coding agents ruteando su tráfico por el proxy
(`lite claude` setea `ANTHROPIC_BASE_URL`+`ANTHROPIC_AUTH_TOKEN`; `lite codex`/`lite opencode` setean
`OPENAI_BASE_URL`) `[V]`. Funciona en los tres CLIs grandes **pero exige API key de pago: rompe la
autenticación por suscripción.** Por eso *todos* los trackers OSS eligieron leer disco. Cierra la
discusión abierta en D7.7.

## Candidato a estándar futuro: ACP

**Agent Client Protocol** (Apache-2.0) transporta `usage_update` dentro de `session/update`, con
token counts requeridos y `cost` opcional con `amount` + `currency` ISO 4217 `[V]`. Hoy lo
implementan pocos agentes y el struct vive tras la feature `unstable_session_usage` `[I]`.
**Vale seguirlo** — si se estandariza, es el canal correcto. **AG-UI no sirve**: no tiene ningún
campo de usage ni costo `[V]`.

## Riesgos abiertos por runtime

1. **Goose**: sin aislamiento de config verificado ⇒ compartiría la config global del usuario.
2. **Codex**: telemetría a OpenAI (`statsig`) **activa por defecto** — apagarla explícitamente.
3. **Goose y Crush**: schema SQLite declarado **no-API-pública**, rompe en migraciones.
4. **Cursor**: sin fallback útil. Atribución por componente **imposible sin proxy**.
