# Investigación — plataformas OSS de observabilidad LLM: qué se puede vendorizar (2026-07-26)

> Insumo para D7 (esquema normalizado) y para la pregunta «¿construimos nuestro propio Langfuse?».
> **`[V]` = verificado leyendo `LICENSE`/fuente · `[I]` = inferido.**

## ⚠️ CORRECCIÓN a `boundaries/telemetria-de-nacimiento.md` v2.0

El boundary afirma que `OTEL_LOG_TOOL_DETAILS=1` *"evita la redacción de terceros"*. **Es falso para
la atribución de tokens/costo** `[V, doc oficial textual]`:

> *"Third-party plugin skill names are replaced with `"third-party"`"* — y lo mismo para
> `plugin.name`, **sin gate documentado**, a diferencia de `workflow.name` que sí dice *"unless the
> gate is set"*.

**Los arneses de ArnesIA se instalan como plugins de un marketplace propio ⇒ caen en "third-party" y
se colapsan a una sola etiqueta.** Esto amenaza la atribución por-componente —el caso de uso central
de la capa Tokens— **incluso en Claude Code**. Hay que resolverlo en el mockup/spec, no descubrirlo
al implementar. Corregir el boundary a v2.1 cuando se firme D9.

Segundo dato del mismo tenor: `claude_code.cost.usage` está documentado como **"Estimated cost"**
`[V]`, no facturación real. La UI no puede presentarlo como plata gastada sin decirlo.

## Veredicto de fondo

**No existe backend "lite" adoptable.** Ninguna plataforma grande es empotrable: Langfuse exige
Postgres + ClickHouse + Redis + S3/MinIO + worker; SigNoz exige ClickHouse + Zookeeper + collector.
Ninguna ofrece single-binary con SQLite. ⇒ **el receptor OTLP se construye en Go, y es la parte
fácil** (handler `net/http` + `go.opentelemetry.io/proto/otlp`, acumulando deltas — Claude Code
exporta con temporalidad `delta` por defecto `[V]`).

Lo que sí se extrae es **dato y esquema**, no plataforma.

## Licencias — ¿vendorizable en producto comercial cerrado?

| Plataforma | Licencia raíz | Carve-out | ¿Vendorizar? |
|---|---|---|---|
| **LiteLLM** | MIT `[V]` | `enterprise/` → licencia propia `[V]` | **SÍ** — el catálogo está en la **raíz**, MIT |
| **Helicone** | Apache-2.0 `[V]` | ninguno hallado `[I]` | **SÍ** |
| **Langfuse** | MIT `[V]` | `ee/`, `web/src/ee/`, `worker/src/ee/` `[V]` | **PARCIAL** — el catálogo vive en `worker/src/constants/`, fuera del carve-out |
| **SigNoz** | MIT `[V]` | `ee/`, `cmd/enterprise/` `[V]` | PARCIAL |
| **Phoenix** | **Elastic v2** `[V]` | — | **NO TOMAR** — obliga a arrastrar sus términos; y todo lo útil es derivado de LiteLLM (MIT), tomamos el original |
| **OpenInference** | Apache-2.0 `[V]` | — | **SÍ** |
| **Lunary** | — | — | **NO — repo 404, borrado** `[V]` |
| OpenLLMetry · OpenLIT · Laminar · Opik · MLflow | Apache-2.0 (solo SPDX) `[V]` | no verificado | probable SÍ `[I]` |
| Weave · Braintrust · Literal AI · Logfire | **no verificado** | **no verificado** | **sin dato** |

## Catálogo de precios — LiteLLM gana, medido

| Catálogo | Lic. | Modelos | cache_read | cache_write | **cache 1h** | tier largo |
|---|---|---|---|---|---|---|
| **LiteLLM** `model_prices_and_context_window.json` | MIT | **2 984** | 710 | 243 | **124** | 95 |
| models.dev `api.json` | MIT | 5 755 | 2 575 | 960 | **0** | 254 |
| genai-prices (Pydantic) | MIT | 1 442 | 345 | 113 | **0** | 18 |
| Langfuse `default-model-prices.json` | MIT | 165 | 73 | 24 | **24** | 7 |
| ccusage | MIT | *sin catálogo propio* — consume LiteLLM | | | | |

**LiteLLM es el único catálogo con cobertura amplia que publica el tier de 1 hora de Anthropic**
(`cache_creation_input_token_cost_above_1hr`, 124 modelos). Langfuse lo tiene pero con 165 modelos y
**solo OpenAI/Anthropic/Google — cero DeepSeek, xAI, Mistral, Llama, Qwen, Ollama** `[V]`: inservible
para multi-runtime. Mantenimiento LiteLLM `[V]`: varios commits por semana, precios **day-0**.

Esquema real `[V]` (`claude-sonnet-4-5`) — matriz completa cache × contexto-largo:

```json
"cache_creation_input_token_cost": 3.75e-6,
"cache_creation_input_token_cost_above_1hr": 6e-6,
"cache_creation_input_token_cost_above_1hr_above_200k_tokens": 1.2e-5,
"cache_read_input_token_cost": 3e-7,
"input_cost_per_token_above_200k_tokens": 6e-6,
"prompt_cache_min_tokens": 1024
```

### Cómo consumirlo sin violar la restricción offline

`ccusage` ya resolvió el patrón y es transportable a Go `[V]`: su `build.rs` descarga el JSON **en
tiempo de compilación** desde un rev fijado, lo compacta, y lo empotra con `include_str!`. Refresco
de red en runtime = **opcional**, con fallback al embebido.

**Equivalente Go: `go:embed` + un paso `go generate`.** Costo medido: filtrado a Anthropic+OpenAI el
JSON pesa **~182 KB** (1,67 MB sin filtrar). Trivial de empotrar.

⚠️ **Trampa a NO heredar:** ccusage **descarta** `cache_creation_input_token_cost_above_1hr` — su
struct tiene un solo `cache_create`. Portar su `build.rs` tal cual reproduce el bug.

## Esquema normalizado — Helicone ya lo tiene

`packages/cost/usage/types.ts`, Apache-2.0, **549 bytes** `[V]`. Es exactamente el superset con
«no aplica» **opcional (no 0)** que exige D7.4, y **la única capa hallada que modela el split 5m/1h
como esquema de primera clase**:

```ts
export interface ModelUsage {
  input: number;  output: number;
  cacheDetails?: { cachedInput: number; write5m?: number; write1h?: number };
  cacheDurationHours?: number;
  thinking?: number;
  image?: ModalityUsage; audio?: ModalityUsage; video?: ModalityUsage; file?: ModalityUsage;
  web_search?: number;
  cost?: number;
}
```

**Salvedad:** sus *processors* por proveedor parsean cuerpos HTTP crudos (Helicone es proxy, incluso
reensambla SSE). ArnesIA no proxea ⇒ **se porta el esquema, no los parsers**.

## La contradicción de aritmética está en la especificación, no es un bug

| Convención | ¿`input_tokens` incluye los cacheados? |
|---|---|
| **OTel semconv** `gen_ai.usage.input_tokens` | **SÍ** — *"SHOULD include all types of input tokens, including cached tokens"* (estabilidad *Development*) |
| **Anthropic / Claude Code** | **NO** — buckets disjuntos (`input`/`output`/`cacheRead`/`cacheCreation`) |
| **OpenAI / Codex** | `cached_input_tokens` es **subconjunto** de `input_tokens` |

**Las dos convenciones son contradictorias y ambas están «bien» según su fuente.** Sin un flag
explícito por adaptador, el error es de **~100 %** en una dirección o **~28 %** en la otra.
Ninguna plataforma revisada expone ese flag como campo del esquema — Helicone lo resuelve
implícitamente con un processor por proveedor.

⇒ **Recomendación: hacerlo explícito** (`usage_arithmetic: disjoint | inclusive` por adaptador).
Eso es **más de lo que hace el estado del arte**.

## Ninguna plataforma tiene el costeo de cache correcto

| Issue | Estado | Patrón |
|---|---|---|
| **langfuse#14249** | **ABIERTO** | omite el cache write del costo — $0.20 vs $0.28 real ⇒ **−28 %** |
| **langfuse#12306** | **cerrado como `stale`, nunca arreglado** | doble conteo: 130 213 tokens se muestran como **260 421 (~2×)**; *"cache hit rate appears ~50% when it's actually ~99%"* |
| **phoenix#14314** | ABIERTO | su `sync_models.py` aplana los tiers de LiteLLM ⇒ *"zero threshold_based customizations across 264 models"* — **sub-factura todo prompt largo** |
| phoenix#13379 · opik#6976/#6969/#6982/#5618 · openllmetry#4060 | cerrados | cached tokens a tarifa de output · sin descontar · tiered no aplicado |
| **el propio Claude Code** `[V]` | pre-v2.1.214 | `usage` en varios frames inflaba `cost.usage` ~una request extra por frame |

**Tres modos de fallo, siempre los mismos** — (a) olvidar el cache write, (b) sumar campos que se
solapan, (c) aplanar los tiers al copiar de LiteLLM. **Son los tres primeros tests a escribir.**

## Ranking: qué extraer

| # | Componente | Origen | Lic. | Nota |
|---|---|---|---|---|
| 1 | **Dataset de precios** | LiteLLM (raíz) | MIT | Es dato, no código: `go:embed` + aviso MIT |
| 2 | **Patrón de embebido build-time** | ccusage `build.rs` | MIT | Offline-first ya probado. **Agregar el tier 1h que ellos omiten** |
| 3 | **Esquema `ModelUsage`** | Helicone `packages/cost/usage/types.ts` | Apache-2.0 | 549 B. Único con `write5m`/`write1h` |
| 4 | **Constantes semconv en Go** | OpenInference `go/openinference-semantic-conventions` | Apache-2.0 | **Se importa como dependencia Go — cero port** |
| 5 | **Reglas declarativas de matching** (`match: {starts_with, or}`, `base`/`tiers`) | genai-prices `data.schema.json` | MIT | Tomar el **esquema**, no la data |
| 6 | **Regex de alias Bedrock/Vertex** | Langfuse `default-model-prices.json` | MIT | `(eu\.|us\.|apac\.)?anthropic\.claude-…` resuelve alias cloud gratis |
| 7 | Processors por proveedor | Helicone | Apache-2.0 | Bajo hoy — solo sirven si algún día proxeamos |

**Reimplementar de cero (no hay nada que copiar):**
1. Receptor OTLP embebido en Go — no hay backend lite adoptable.
2. **Mapeo de namespaces híbridos** (`claude_code.*` / `codex.*` / `gemini_cli.*` → llave canónica
   del terreno). **Nadie lo tiene** `[V, por ausencia]`.
3. **Flag `usage_arithmetic` por adaptador** — el fix de langfuse#12306 que la industria no hizo.

## Canal de ingesta: evaluar `api_request` contra las métricas

El log event `claude_code.api_request` trae **por request** `cost_usd_micros` (entero), los 4 buckets
de tokens, `model`, `speed`, `effort`, `query_source` y toda la atribución `[V]`. Granularidad exacta
sin lidiar con temporalidad delta ni cardinalidad de métricas. **Evaluarlo en `spec.md`** — puede ser
mejor canal que `claude_code.token.usage`.

## Efecto sobre D8

**Se refuerza, no se desbloquea.** Confirmado por dos vías independientes: la telemetría OTel de
Claude Code **no expone el split 5m/1h** en ningún canal, y **ningún catálogo salvo LiteLLM siquiera
publica el precio de 1h**. La disyuntiva (a) leer 4 campos numéricos del JSONL vs (b) resignar los
detectores de fuga sigue abierta tal cual.
