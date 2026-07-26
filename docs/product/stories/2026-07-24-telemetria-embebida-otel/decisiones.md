# Decisiones — telemetría embebida vía OTel nativo (capa Tokens del Mapa)

> Disciplina METODOLOGIA §10: cada decisión conversada se escribe acá EN EL MISMO TURNO.
> Origen: barrido de deuda viva 2026-07-23/24 (HS-27), ítem "telemetría JSONL → indexer real".

## D1 — Restricción rectora del operador (2026-07-24)

> «arnesia es un producto instalable, si usamos langfuse vamos a tener que portarlo o llamarlo
> como dependencia al momento de instalarlo. No vamos a entregar un manual con el paso a paso
> al usuario para que instale y conecte langfuse... piensa en cómo debe ser nuestra
> arquitectura y con eso toma una decisión»

Ninguna pieza de la arquitectura de telemetría puede exigir infraestructura externa
(Langfuse/Postgres/ClickHouse/Docker) al usuario final. El instalador (`.deb`/`.AppImage`/`.rpm`)
nunca llama esas dependencias. Esta restricción YA era ley en `vision.md` ("Langfuse = espejo
opcional, jamás dependencia dura") — D1 la hace explícita como criterio de diseño, no solo prosa.

## D2 — Hallazgo que forzó revisar el ángulo Langfuse

Existe un emisor de telemetría YA CONSTRUIDO en el repo predecesor congelado
(`~/Proyectos/prenter-harness/products/kit/core-harness/telemetry/emit.py`, KIT-03, 508
líneas, spec `products/kit/specs/KIT-03-telemetria-embebida.md`): hooks Stop/SubagentStop/
SessionEnd parsean el JSONL incrementalmente, atribuyen por skill (`attributionSkill`), calculan
costo (tabla `TARIFAS` por modelo), aplican un walk de privacidad ("THE KEY": ningún campo
`egreso: sensible` cruza el borde sin verificar) y exportan OTLP a Langfuse. Hay un Langfuse
real corriendo en esta máquina AHORA (`docker ps`: 6 contenedores, `~/.prenter/observatorio/
langfuse/`) con datos de un batch de pruebas viejo. El operador confirmó: se dejó de usar.

**Por qué NO se porta tal cual:** viola dos leyes ya firmadas de ESTE repo:
1. D1 (instalable, Langfuse nunca dependencia dura) — el diseño legacy asume un Langfuse real.
2. [`docs/architecture/boundaries/conductor-no-parsea-jsonl.md`](../../architecture/boundaries/conductor-no-parsea-jsonl.md)
   (`severity: high`): *"el transcript JSONL... es formato interno que cambia entre versiones...
   no parsear el JSONL directo"* — exactamente lo que hace `emit.py`.

Es prior art valioso como REFERENCIA (concepto de atribución por-skill, tabla de costos por
modelo, ética fail-open, walk de privacidad) — no como código a portar.

## D3 — Verificación oficial que resuelve el canal correcto (2026-07-24)

Consultada la doc oficial de Claude Code (`code.claude.com/docs/en/monitoring-usage.md` +
`.../agent-sdk/observability.md`):

- `CLAUDE_CODE_ENABLE_TELEMETRY=1` + `OTEL_EXPORTER_OTLP_ENDPOINT` activan telemetría OTel
  **nativa** — sin hook custom, sin script.
- Métricas confirmadas: `claude_code.token.usage` + `claude_code.cost.usage` (USD).
- Atributos de atribución LIMPIA: `skill.name` / `plugin.name` / `tool_name` / `agent.name` /
  `session.id` (`OTEL_LOG_TOOL_DETAILS=1` evita la redacción por defecto de terceros).
- OTLP es protocolo abierto (HTTP/JSON, HTTP/protobuf, gRPC) — *"any backend that accepts
  OTLP... or a self-hosted collector"*. Un receptor mínimo casero (no el OpenTelemetry Collector
  completo, no Langfuse) es un backend válido.
- Funciona headless (`claude -p`, Agent SDK) igual que interactivo, por `session.id`.
- Traces completos siguen en beta (`CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`); métricas
  (token/cost) NO están marcadas beta.

## D4 — Arquitectura resuelta (FIRMADA, ejecutada como boundary v2.0)

1. **Receptor OTLP embebido loopback-only, dentro del propio daemon Go** — un endpoint HTTP más
   (`POST /v1/metrics`) junto al resto de la API existente. Cero proceso/contenedor nuevo.
2. **`scaffold` (futuro) inyecta env vars**, no un hook Python: `CLAUDE_CODE_ENABLE_TELEMETRY=1`
   + `OTEL_METRICS_EXPORTER=otlp` + `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:<puerto>`.
3. **JSONL sigue siendo SOLO enumerar/replay** (`history/reader.go`, ya construido para el
   historial B2) — nunca fuente de tokens/costo/atribución. Coherente con
   `conductor-no-parsea-jsonl.md`, sin excepciones nuevas.
4. **Langfuse (o cualquier OTLP externo) = exportador 100% opcional**, config del operador o de
   un power-user — jamás instalado ni requerido por el producto. El Langfuse corriendo hoy es
   infraestructura de observatorio propio de alpacapurpura (flota), fuera de este árbol.
5. **Índice local reusa el patrón ya doctrinado**: map in-memory + JSON atómico en `~/.arnesia`
   (mismo lugar que el índice estructural; `indice-desechable-jsonl-es-verdad.md`), agregado por
   arnés/sesión/skill desde las métricas OTel recibidas.

Detalle completo + checklist evaluable actualizado (v2.0) en
[`docs/architecture/boundaries/telemetria-de-nacimiento.md`](../../architecture/boundaries/telemetria-de-nacimiento.md).

## D5 — Corrección honesta sobre una decisión anterior de esta MISMA sesión

Antes de este hallazgo, el operador había elegido (AskUserQuestion previa) un plan más chico:
"per-nodo parcial vía Skill/Task, parseando JSONL". **D3/D4 SUPERSEDEN esa elección** — el canal
OTel nativo da atribución por-skill limpia (`skill.name`) sin ninguna de las desventajas de
parsear JSONL (formato inestable, turnos ambiguos, sin costo en USD listo). No se descarta en
silencio: se documenta acá que la arquitectura mejor disponible reemplaza la primera respuesta,
descubierta un paso después en la misma investigación.

## D6 — Cero instalación manual, cross-platform Windows/macOS/Linux (2026-07-25)

> «definitivamente no podemos pedirle al usuario que instale algo manualmente, sea lo que sea lo
> debemos instalar nosotros al momento de instalar arnesia y debemos pensar que debe funcionar en
> windows y MAC»

Endurece D1: no es solo «Langfuse no» — es **ninguna pieza, de ningún tipo, instalada a mano**.
El receptor OTLP embebido en el binario Go ya cumple (no hay nada nuevo que instalar). Lo que
queda expuesto es todo lo que en Windows/macOS **le pide un paso manual al usuario aunque no sea
un `install`**:

1. **Firma de código = requisito, no opcional.** Sin Apple Developer ID + notarización, Gatekeeper
   obliga a «click derecho → Abrir» o `xattr -d`. Sin cert de firma Windows, SmartScreen obliga a
   «Más información → Ejecutar de todos modos». **Eso ES un paso manual** y viola D6. El sidecar
   `arnesia-daemon` necesita su propia firma + hardened runtime en macOS, no alcanza con el `.app`.
2. **Bundles Windows/macOS necesitan runners de esa plataforma** — Tauri no cross-compila el
   bundle. `tauri.conf.json` ya declara `"targets": "all"` y el `Makefile:58` ya copia
   `.dmg`/`.msi`/`.exe`, pero hoy se buildea solo en Linux: los targets existen en config, no en CI.
3. **Bind loopback estricto (`127.0.0.1`, jamás `0.0.0.0`/`::`)** — bindear a todas las interfaces
   dispara el diálogo de firewall de Windows = paso manual. `cmd/arnesia/main.go:100` ya está bien;
   el endpoint OTLP no debe cambiarlo.
4. **Auth también en el canal OTLP** — en Windows el loopback es promiscuo: cualquier proceso local
   puede POSTear métricas falsas o leer el índice. `OTEL_EXPORTER_OTLP_HEADERS` permite token sin
   romper el protocolo abierto.
5. **Descubrimiento de puerto.** Puerto fijo `4200` colisiona; y el puerto entra en config de un
   tercero (el runtime) — si cambia entre arranques, la telemetría se pierde EN SILENCIO. Hace
   falta puerto estable + archivo de descubrimiento + re-escritura de la config del runtime.
6. **WSL2 / devcontainers: agujero real.** Un agente corriendo dentro de WSL2 o de un contenedor
   NO alcanza el `127.0.0.1` del host (salvo `mirrored` networking en Win11 23H2+). Caso frecuente.
   La UI debe decir «sin dato» honesto, nunca 0 tokens.
7. **Rutas por SO**: `~/.arnesia` → `os.UserConfigDir()` (`%APPDATA%` / `~/Library/Application
   Support`). Y la codificación de proyecto tipo `DirParaCwd` cambia con paths `C:\…`.
8. **Auto-update por SO**: `updater.go` es Linux-céntrico — Windows no puede reemplazar un `.exe`
   en ejecución (rename + restart) y macOS debe preservar la firma.

## D7 — Multi-runtime: la telemetría es un PUERTO, no un receptor OTLP (2026-07-25)

> «no solo trabajaremos con claude code, en el futuro implementaremos arneses para codex, open
> code, y así»

Ya es ley del árbol (`boundaries/adaptadores-de-agente-intercambiables.md`, `AgentPort`;
`vision.md`: *«el mercado puede mutar a otros agentes (OpenCode, Codex, pi, Antigravity)»*). D4 se
**re-lee**: lo resuelto no es «la arquitectura de telemetría», es **el adaptador `claudecode`**.
Consecuencias a resolver en `spec.md`:

1. **La telemetría OTel nativa es exclusiva de Claude Code.** Ningún otro runtime emite
   `claude_code.token.usage`. Cada runtime = su propio canal (Codex: rollouts + `config.toml`;
   OpenCode: su storage propio). N adaptadores, N fragilidades de versión.
2. **La llave de atribución NO puede ser `skill.name`** — es un concepto de CC que Codex no tiene.
   Debe ser canónica del terreno propio (**arnés × caja/paso × sesión**), y cada adaptador mapea lo
   suyo a esa llave. **Esto bloquea el mockup**: si la capa Tokens se dibuja sobre `skill.name`, no
   sobrevive al segundo runtime.
3. **El costo en USD deja de venir gratis.** CC emite `cost.usage`; los demás probablemente solo
   tokens ⇒ vuelve la tabla de precios que descartamos en D2, ahora inevitable como *fallback
   per-runtime*, con «sin dato» honesto cuando el modelo es desconocido.
4. **La taxonomía de tokens diverge.** Anthropic: `cache_creation` (5m/1h) + `cache_read`. OpenAI:
   `cached_tokens` (automático, sin costo de escritura) + `reasoning_tokens`. El esquema
   normalizado debe ser **superset con «no aplica» explícito ≠ 0** — un 0 donde el concepto no
   existe es mentira. Corolario: **los detectores de fuga también son per-runtime** (el break-even
   de cache TTL depende de la estructura 1.25×/2×/0.1× de Anthropic; no aplica a OpenAI).
5. **Mapeo de semconv**: si un runtime emite `gen_ai.*` en vez de `claude_code.*`, el receptor mapea
   ambos. La semconv GenAI sigue en estabilidad *Development* — versionar el mapeo.
6. **Cada adaptador declara qué detectores soporta**; la UI degrada por runtime (mismo patrón que
   el Mapa degradado ya firmado), nunca ceros fabricados.
7. **Proxy descartado como estrategia general**: además del riesgo de ToS al interceptar tráfico
   autenticado de terceros, instalar un CA cert local rompe D6.

## D8 — DECISIÓN ABIERTA: el split de TTL 5m/1h no existe en OTel (2026-07-25)

Investigación externa (subagente, 2026-07-25) verificó contra doc oficial: los campos
`cache_creation.ephemeral_5m_input_tokens` / `ephemeral_1h_input_tokens` **NO están en la
telemetría OTel de Claude Code** — solo en el JSONL, o en `claude_code.api_response_body` (que
exige `OTEL_LOG_RAW_API_BODIES=1`, o sea la conversación entera en el canal de telemetría:
incompatible con la postura de privacidad).

Sin ese split no se pueden construir los dos detectores de **fuga** más valiosos (re-warm por TTL
expirado con break-even ≈ 39.47 %, y watchdog de regresión de TTL). Los caminos son:
(a) enmendar `conductor-no-parsea-jsonl.md` para permitir leer **4 campos numéricos de `usage` +
`timestamp`, jamás contenido**; (b) resignar esos detectores.

Dato que pesa: el check que lo prohíbe (`fitness/arch_test.go:TestNoJSONLSchemaParsing`) está
**`t.Skip`eado** y `internal/adapters/history/reader.go` ya parsea schema JSONL — la excepción de
facto ya existe. **Decidir explícitamente, no por omisión.** Sin firma, esta queda ABIERTA.

## D9 — Arquitectura consolidada tras la investigación SOTA (2026-07-26) · ⏳ SIN FIRMAR

> Origen: encargo del operador *«investigación sobre el SOTA de telemetría para múltiples
> proveedores y si podemos extraer lo necesario de GitHub y crear nuestra propia versión»*.
> Tres carriles de investigación cerrados →
> [`investigacion-runtimes.md`](investigacion-runtimes.md) ·
> [`investigacion-plataformas.md`](investigacion-plataformas.md) ·
> [`investigacion-stack-embebible.md`](investigacion-stack-embebible.md).
> **Esto es RECOMENDACIÓN de la investigación, no decisión firmada. Falta 🧑‍⚖️.**

### D9.1 — «Crear nuestro propio Langfuse» es la pregunta equivocada

**No existe backend lite adoptable** (todas exigen Postgres+ClickHouse+Redis+worker). Y no hace
falta: lo que necesitamos son ~1,5 de los 6 subsistemas de Langfuse. **Lo vendorizable es DATO y
ESQUEMA, no plataforma:**

| # | qué | origen | licencia |
|---|---|---|---|
| 1 | catálogo de precios (2 984 modelos, **único con tier 1h**) | LiteLLM, raíz | MIT |
| 2 | patrón de embebido build-time (`go:embed` + `go generate`) | ccusage `build.rs` | MIT |
| 3 | esquema `ModelUsage` (549 B, `write5m`/`write1h` opcionales) | Helicone `packages/cost/usage/types.ts` | Apache-2.0 |
| 4 | constantes semconv **en Go** (se importa, cero port) | OpenInference | Apache-2.0 |
| 5 | regex de alias Bedrock/Vertex | Langfuse `worker/src/constants/` | MIT |

**Descartados:** Phoenix (**Elastic v2** — obliga a arrastrar sus términos; y todo lo útil deriva de
LiteLLM MIT) · Lunary (**repo 404, borrado**).

### D9.2 — Dos carriles de ingesta, un solo evento canónico

OTel **no** es el canal universal: solo 3 de 6 runtimes lo dan útil en headless, mientras el
`stream-json` por turno lo dan 5 de 6. **El adaptador base es stream-json; OTel es enriquecedor.**

Diseño: un `ports.TelemetrySink` con dos adaptadores (`otlp/` y `streamjson/`), ambos emitiendo el
mismo `domain.EventoTelemetria` → un solo writer serializado. Encaja con la hexagonal vigente.

Corolario: como hay dos orígenes heterogéneos, **el esquema canónico no puede ser el modelo de datos
de OTel** — `pdata` queda degradado a *decodificador de wire format*, lo que refuerza elegir el
decodificador barato y estable.

### D9.3 — Stack embebido (medido, no opinado)

| capa | elección | peso |
|---|---|---:|
| recibir OTLP | handler `net/http` propio + **`collector/pdata`** (`pmetricotlp`/`plogotlp`) | **+1,7 MB** |
| almacenar | **`modernc.org/sqlite`** (ya en `go.mod`) + rollup horario incremental | 0 |
| consultar | SQL sobre el rollup; p95 por histograma de buckets | 0 |

`pdata` v1.63.0 tiene **API Go estable**; `otlpreceiver` está en v0.157.0 (config estable, **API Go
no**) y cuesta **+8,4 MB / +60 módulos** por gRPC/TLS/auth que en loopback no se usan.
Rollup horario: **1,7 ms vs 323 ms crudo (190×)** — deja a DuckDB sin nada que ganar.

**Ratifica `sin-cgo`** (`arch_test.go:TestNoDuckDBOrCGOStore`, ya enforced): DuckDB = 240 MB de
descarga en CI, 1,08 GB en disco, MinGW-no-MSVC en Windows. **chDB descartado por tres motivos
simultáneos:** sin Windows · `libchdb.so` dinámica (el usuario instalaría algo, viola D6) ·
install `curl | bash`.

**Config obligatoria del spawn:** `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` **además** del
endpoint — el default de Claude Code es **gRPC :4317**.

### D9.4 — `OTEL_RESOURCE_ATTRIBUTES` es el vector universal de atribución

Lo respetan Claude Code, Copilot CLI, Codex, Goose y OpenCode. Es la palanca para inyectar
**nuestro** identificador de unidad de trabajo (arnés · caja · sesión) al spawnear, sin depender del
vocabulario de cada runtime — y **mitiga la redacción «third-party»** (ver D10).

⚠️ **Límite honesto:** los resource attributes se fijan **por proceso** ⇒ atribución a granularidad
de proceso spawneado, **no** por-skill dentro de una sesión. **Pregunta abierta para el mockup.**

### D9.5 — El esquema normalizado necesita un campo que nadie tiene

Las dos convenciones son contradictorias y ambas están «bien» según su fuente:

| convención | ¿`input_tokens` incluye los cacheados? |
|---|---|
| OTel semconv | **SÍ** — *"SHOULD include all types of input tokens, including cached tokens"* |
| Anthropic / Claude Code | **NO** — buckets disjuntos |
| OpenAI / Codex | `cached_input_tokens` ⊂ `input_tokens` |

Sin flag explícito el error es **~100 %** en un sentido o **~28 %** en el otro. **Ninguna plataforma
lo expone como campo.** ⇒ **`usage_arithmetic: disjoint | inclusive` por adaptador.** Eso nos pone
por encima del estado del arte, y es el fix de `langfuse#12306` que la industria no hizo.

Segundo campo obligatorio: **la regla de acumulación es propiedad del adaptador** (Claude Code
acumula en `result` · Codex re-emite `last_token_usage` · Amp es per-message). Tres runtimes, tres
reglas.

### D9.6 — Los tres primeros tests están dictados por los bugs ajenos

Ninguna plataforma tiene el costeo de cache correcto (`langfuse#14249` abierto · `#12306` cerrado
como `stale` **sin arreglar** · `phoenix#14314` · 4 issues en Opik). **Tres modos de fallo, siempre
los mismos** — son los tres primeros tests a escribir:
1. olvidar el cache **write** en el costo,
2. sumar campos que **se solapan**,
3. **aplanar los tiers** al copiar el catálogo.

### D9.7 — Proxy DESCARTADO con evidencia

LiteLLM tiene el comando `lite` que lanza coding agents ruteando por su proxy (`lite claude`,
`lite codex`, `lite opencode`) — **pero exige API key de pago: rompe la autenticación por
suscripción.** Por eso *todos* los trackers OSS eligieron leer disco. Cierra D7.7, que hasta acá se
apoyaba solo en el argumento de ToS y del CA cert.

### D9.8 — El hueco de mercado confirma la tesis del producto

La categoría «tracker multi-CLI» está **saturada y es MIT**: ccusage (17,4k ★, Rust, 15 runtimes,
pricing de LiteLLM), tokscale (40+ runtimes), TokenTracker, caut. Dynatrace y Azure Monitor ya
venden el dashboard multi-agente.

> **Pero ninguno correlaciona el uso con QUÉ COMPONENTE / ROL / PROCESO estaba corriendo.** Todos
> agregan por herramienta, modelo, proyecto y día. **La atribución por unidad de trabajo es terreno
> libre.**

⇒ Lo diferencial **no es medir tokens** — es el eje **arnés × empresa × puesto**. El diseño debe
optimizar por eso, no por competir en cobertura de runtimes.

### D9.9 — DECISIÓN ABIERTA: ¿parser propio o shell-out a `ccusage`?

La investigación recomienda **no** construir parser de disco propio y hacer shell-out a
`ccusage <source> --json` (MIT, Rust, sin runtime Node, 15 runtimes). Contras reales a pesar:
- Es un **sidecar nuevo** que hay que shipear y firmar por plataforma (nivel B de D6 — admisible,
  pero suma superficie, y arrastra el bug `tauri#11992` multiplicado por dos binarios).
- **ccusage descarta `cache_creation_input_token_cost_above_1hr`** (un solo campo `cache_create`)
  ⇒ pierde exactamente el dato que habilita los detectores de fuga.
- Acopla nuestro roadmap al suyo.

**Sin firma. Decidir explícitamente en el `spec.md`.**

## D10 — Corrección a `telemetria-de-nacimiento.md` v2.0 (2026-07-26)

El boundary v2.0 afirma que `OTEL_LOG_TOOL_DETAILS=1` *«evita la redacción de terceros»*.
**Es falso para la atribución de tokens/costo.** Doc oficial, textual:

> *"Third-party plugin skill names are replaced with `"third-party"`"* — ídem `plugin.name`,
> **sin gate documentado**, a diferencia de `workflow.name` que sí dice *"unless the gate is set"*.

**Nuestros arneses se instalan como plugins de un marketplace propio ⇒ caen en «third-party» y se
colapsan a una sola etiqueta.** Amenaza la atribución por-componente —el caso de uso central de la
capa Tokens— **incluso en Claude Code**.

Mitigación disponible: **D9.4** (`OTEL_RESOURCE_ATTRIBUTES`), con el límite de granularidad
por-proceso ahí documentado. **No es equivalente**: hay que resolverlo en el mockup/spec.

Dato hermano: `claude_code.cost.usage` está documentado como **"Estimated cost"**, no facturación
real. La UI no puede presentarlo como plata gastada sin decirlo.

⇒ **Boundary a corregir a v2.1** cuando se firme este bloque.

## D11 — Restricción de empaquetado: A o B, nunca C (2026-07-26)

Sobre *«o que al instalar obligue a extraer las dependencias previo a la instalación»*: tres niveles
posibles — **(A)** compilado dentro del binario Go · **(B)** sidecar shipeado en el instalador ·
**(C)** descargado de internet en post-install.

**C se descarta.** Rompe instalación offline y redes corporativas, agrega un modo de falla nuevo
(app instalada pero rota), y **`.dmg` no tiene post-install real** (es drag-to-Applications; haría
falta `.pkg`, que Tauri no bundlea) ⇒ ni siquiera es uniforme entre plataformas. Formalmente no es
«manual del usuario», pero el usuario igual ve el fallo.

**El `.dmg` sin post-install es el mínimo común denominador que gobierna el diseño.** La solución
correcta es **cero post-install en las tres plataformas** — exactamente lo que da un binario pure-Go
embebido (D9.3).

Excepción tentadora: el catálogo de precios envejece. Solución = **embeber (A) + refresco opcional
en background con degradación honesta** — si nunca refresca, funciona con los precios del release
y **lo dice**.

### Riesgos operativos altos destapados (nuevos, no había registro)

1. 🔴 **`tauri#11992`** — la notarización de macOS falla con `externalBin`: devuelve
   `The signature of the binary is invalid`; sacándolo, notariza bien. Issue **abierto**.
   **Ya usamos `externalBin`** ⇒ pega de frente al abrir el canal macOS. Probar temprano.
2. 🔴 **Azure Artifact Signing es geo-restringido** (orgs USA/Canadá/UE/UK; individuos solo
   USA/Canadá). Operación latinoamericana ⇒ **probablemente no califica**. Plan realista:
   **cert OV USD 150–300/año** + clave en HSM cloud (exigido por CA/Browser Forum desde jun-2023).
   **EV ya no compra bypass de SmartScreen** (Microsoft lo removió en 2024) — no pagarlo por eso.
3. 🟠 **La reputación de SmartScreen se acumula por hash de archivo** ⇒ cada release reinicia el
   contador. Vamos por v0.2.18. Argumento para espaciar releases públicos o usar winget/Store.
4. 🟠 **macOS Sequoia eliminó el bypass por Control-click** ⇒ sin firma son 5 pasos y 2 diálogos de
   miedo. Inaceptable para un vendible.

## D12 — Alcance de la medición y forma del MVP (2026-07-26, elección del operador)

### D12.1 — Alcance: **S1 + S2, máquina propia. S3 FUERA.**

| escenario | ¿en alcance? |
|---|---|
| **S1** — arnés dentro del chat dock de ArnesIA (controlamos el spawn) | ✅ |
| **S2** — arnés en la MISMA máquina, fuera de ArnesIA (terminal/IDE del usuario) | ✅ |
| **S3** — arnés en la máquina de un cliente sin ArnesIA instalada | ❌ **fuera** |

**Todo queda en `127.0.0.1`. Cero egreso de red.** S3 queda descartado explícitamente: chocaba de
frente con [`superficie-local-confinada.md`](../../architecture/boundaries/superficie-local-confinada.md)
y abría consentimiento, datos de terceros y GDPR — decisión de producto y legal, no de arquitectura.
Si alguna vez se reabre, es paquete propio.

**Consecuencia directa: el mecanismo de obligación SÍ se construye.** S2 es justamente el caso donde
no controlamos el spawn ⇒ la instrumentación tiene que viajar **dentro del arnés**. Los 4 checks de
conformance de [`arquitectura-telemetria.md`](arquitectura-telemetria.md) §Parte 3 quedan en alcance:
`arnes-declara-telemetria` · `arnes-porta-hook-proceso` · `hook-es-fail-open` · `telemetria-no-egresa`.

Requisitos que S2 agrega y S1 no tenía:
1. **Archivo de descubrimiento de puerto** — el hook necesita encontrar el daemon sin que nadie le
   pase el endpoint (retoma D6.5).
2. **Fail-open silencioso obligatorio** — si el daemon no está, el hook **jamás** rompe ni demora el
   trabajo del usuario. Concepto reusado de `emit.py`/KIT-03 (D2), no su canal.
3. **`telemetria-no-egresa` es un check duro**, no un buen deseo: el hook solo escribe a loopback o a
   disco local.

### D12.2 — El MVP es el JOIN, no una capa sola

Elegido: **no entregar tokens solos ni proceso solo — el MVP es la correlación.**

> *«este arnés, en este puesto, quema $X — y el 60 % se va en la caja Y, que falla el gate 3 de cada
> 4 veces»*

Es el hueco de D9.8 que nadie llenó (ccusage, tokscale, Dynatrace: todos agregan por herramienta,
modelo, proyecto y día; **ninguno por unidad de trabajo**). La llave del join es `session.id` + los
resource attributes que inyectamos.

### ⚠️ D12.3 — Esto EXPANDE el alcance del paquete. Registrado, no absorbido en silencio

El paquete se abrió como **«capa Tokens del Mapa»**, y su `INDEX.md` declara explícitamente que las
capas **Desempeño** y **Proceso** están *«fuera de alcance… siguen genuinamente bloqueadas, no
recortables»*, con `bloqueo` en el BACKLOG.

**D12.2 tira esa frontera abajo**: el join necesita la señal de proceso, que era justamente la capa
Proceso. Consecuencias a asumir:

1. **El nombre «capa Tokens» ya no describe el entregable.** Es una capa de *mejora* que junta
   dinero + proceso. Renombrar en el mockup.
2. **El ítem del BACKLOG «capas Desempeño/Proceso» deja de ser `bloqueo` puro** — Proceso entra
   parcialmente por esta puerta. Desempeño (latencia/reintentos) sigue afuera.
3. **El mockup tiene que dibujar el join**, no un badge de tokens por nodo. Es un diseño más difícil
   y es lo que hay que llevar al gate.
4. **Más capabilities**: el módulo `telemetria/` cubre ingesta + costeo + detección + hooks de
   proceso, no solo la capa visual.

**Es la elección del operador, hecha con la expansión a la vista. No es scope creep silencioso.**

### 🟢 D12.4 — Buena noticia de secuenciación: el MVP del join NO depende de D8

D8 (¿leer el split `ephemeral_5m/1h` del JSONL?) bloquea **solo** los detectores **B1** (re-warm por
TTL) y **B12** (watchdog de TTL). **Los otros 11 detectores + todo el eje de proceso no lo necesitan.**

⇒ **D8 deja de ser bloqueante del MVP** y pasa a ser bloqueante de dos detectores específicos. Se
puede firmar después, sin frenar el paquete. (D9.9 —parser propio vs shell-out a `ccusage`— sigue
igual de abierta, pero también es posterior al MVP.)

## D13 — Aclaración: el egreso del daemon y el del arnés son niveles distintos (2026-07-26)

Consulta del operador: *«¿estás considerando que ArnesIA consuma Langfuse? ¿cómo has pensado para
los casos de los usuarios que no tienen instalado Langfuse?»*

**Respuesta: no. ArnesIA NO consume Langfuse.** El usuario sin Langfuse —el default de todos—
recibe el **100 %** de la funcionalidad: `runtime → 127.0.0.1:4318 → SQLite local → capa de mejora`,
todo dentro del binario que ya se instala. El grupo C de
[`arquitectura-telemetria.md`](arquitectura-telemetria.md) estaba mal titulado y se prestaba a
leerse al revés; corregido.

**Pero la consulta destapó una ambigüedad real que queda cerrada acá:** D12.1 dice «cero egreso de
red» y C2 describe un forward OTLP opcional. **No es contradicción — son dos niveles**, y hay que
escribirlo o la implementación rompe uno de los dos:

| nivel | quién controla | ¿egreso? |
|---|---|---|
| **arnés / hook** (S2) | el paquete instalado | **NUNCA** — loopback o disco local. Check duro `telemetria-no-egresa` |
| **daemon** | el operador, config deliberada | sí — **off por default**, con indicador visible en la UI |

Dos invariantes que se derivan y deben quedar en el `spec.md`:
1. **`telemetria-no-egresa` ata al ARNÉS, no al daemon.**
2. **Un arnés no puede alcanzar ni configurar el forward del daemon** — no es capacidad expuesta al
   paquete. (Coherente con el guardrail ya vigente de que el alcance del chat embebido excluye el
   paquete propio.)

## Estado del paquete

**Diseño de arquitectura RESUELTO y documentado** (boundary v2.0 + esta ficha). **CERO código
construido en este cierre** — es correctamente un paquete propio: toca UI nueva del Mapa (capa
Tokens, hoy sin ningún diseño visual — ni siquiera el mockup "destino" la dibuja, ver
`mockups/arnesia-mapa-destino.html:295`, "Fase 2") y un componente Go nuevo (receptor OTLP). Por
disciplina §10 (feature nueva = mockup→spec→PARIDAD), NO se codea salteando ese proceso. Este
paquete es el punto de arranque cuando se retome: falta mockup de la capa Tokens (granularidad
per-skill, dónde vive el número — ¿badge en el nodo? ¿panel del inspector?) → spec del receptor
OTLP + del scaffold de env vars → build → PARIDAD.
