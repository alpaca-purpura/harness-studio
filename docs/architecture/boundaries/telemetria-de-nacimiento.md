---
regla: telemetria-de-nacimiento
version: 2.3
updated: 2026-07-26
status: proposed
ledger: HS-27
sources:
  - url: https://opentelemetry.io/docs/concepts/observability-primer/
    autoridad: oficial
    revisado: 2026-07-23
  - url: https://code.claude.com/docs/en/monitoring-usage.md
    autoridad: oficial
    revisado: 2026-07-24
  - url: https://code.claude.com/docs/en/agent-sdk/observability.md
    autoridad: oficial
    revisado: 2026-07-24
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/INFORME.md
    autoridad: medicion-propia
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/ANEXO-hooks.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by: []
severity: error
---

# Todo arnés nace observable (VISION p9)

## L1 · Principio (estándar de industria)

**Observability by default, no opt-in.** Un sistema producido en masa (aquí: arneses) debe
nacer instrumentado — la telemetría no es algo que se agrega después de un incidente, es parte
del template de creación. *(oficial: OpenTelemetry — observability primer, instrumentación como
parte del ciclo de vida del servicio, no un add-on posterior)*

## L2 · Realización (este árbol Go) — SIN IMPLEMENTAR, ARQUITECTURA RESUELTA (v2.0, 2026-07-24)

**Restricción rectora (no negociable, orden directa del operador 2026-07-24):** ArnesIA es un
producto **instalable** (binario Go único + shell Tauri, `.deb`/`.AppImage`/`.rpm`). Ningún
requisito de telemetría puede exigirle al usuario instalar/operar infraestructura externa
(Langfuse, Postgres, ClickHouse, Docker) — el instalador nunca llama esas dependencias. Esto
**ya era ley** (`vision.md`: *"Langfuse = espejo opcional, jamás dependencia dura"*) — v2.0 la
aplica correctamente a este nodo, que hasta v1.0 quedó vago y podía leerse mal (¿"collector OTLP
embebido" implica correr Langfuse? NO).

**Hallazgo que forzó la revisión:** existe un emisor de telemetría YA CONSTRUIDO y probado en el
repo predecesor congelado (`~/Proyectos/prenter-harness/products/kit/core-harness/telemetry/
emit.py`, KIT-03, 508 líneas) — hooks Stop/SubagentStop/SessionEnd que **parsean el JSONL**
incrementalmente, atribuyen por skill, calculan costo (tabla `TARIFAS` por modelo) y exportan
OTLP a un Langfuse real (contenedor Docker corriendo en esta misma máquina, `~/.prenter/
observatorio/langfuse/`, `docker ps` confirma 6 contenedores activos). El operador aclaró: se
dejó de usar, y **portarlo tal cual violaría dos leyes YA firmadas de este repo**: (1) la
restricción de instalable de arriba (Langfuse como backend real, no opcional) y (2) el boundary
hermano [`conductor-no-parsea-jsonl.md`](conductor-no-parsea-jsonl.md) (severity `high`): el
JSONL es **"formato interno que cambia entre versiones... nunca parsear su schema como
contrato"** — exactamente lo que `emit.py` hace. Es prior art valioso (concepto de atribución,
tabla de costos, walk de privacidad, diseño fail-open) pero el CANAL que usa es el incorrecto
para este árbol.

**Verificación oficial (2026-07-24, `code.claude.com/docs`) que resuelve la arquitectura:**
Claude Code emite telemetría OTel **nativa** (sin hook custom, sin script Python) con
`CLAUDE_CODE_ENABLE_TELEMETRY=1` + `OTEL_EXPORTER_OTLP_ENDPOINT`. Métricas confirmadas:
`claude_code.token.usage` + `claude_code.cost.usage` (USD), con atributos `skill.name` /
`plugin.name` / `tool_name` / `agent.name` / `session.id` — **v2.1 CORRIGE lo que v2.0 afirmaba acá**
(ver ⚠️ abajo): esa atribución **NO es limpia para plugins de terceros**, y nuestros arneses cuentan
como tales. Sigue siendo el dato que necesita la capa Tokens del Mapa, sin tocar el JSONL para nada.
OTLP es protocolo abierto
(HTTP/JSON, HTTP/protobuf o gRPC): *"any backend that accepts OTLP... or a self-hosted
collector"* — un receptor OTLP mínimo casero (no el OpenTelemetry Collector completo, no
Langfuse) es un backend válido.

**Arquitectura resuelta:**
1. **Receptor OTLP embebido EN el propio daemon Go** (nuevo endpoint HTTP junto al resto de la
   API, ningún proceso/contenedor nuevo) — decodifica `POST /v1/metrics` (OTLP) y agrega
   localmente. Esto ES "collector-embebido" del checklist de abajo, correctamente entendido:
   embebido en el binario que YA se instala, no una pieza nueva a desplegar.
2. **`scaffold` (cuando exista, ver BACKLOG) inyecta env vars locales**, no un hook — cada arnés
   nuevo nace con `CLAUDE_CODE_ENABLE_TELEMETRY=1` + `OTEL_METRICS_EXPORTER=otlp` +
   `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:<puerto del daemon>` (loopback, nunca sale de
   la máquina por defecto). Esto ES "scaffold-emite-hook" del checklist, más simple de lo que
   sonaba: Claude Code ya instrumenta solo, no hay `emit.py` que portar.
3. **JSONL sigue exactamente como manda `conductor-no-parsea-jsonl.md`**: enumerar/replay
   (`history/reader.go`, ya construido para el historial B2), NUNCA fuente de tokens/costo/
   atribución — ese dato vive en las métricas OTel, canal correcto y ya verificado.
4. **Langfuse (u OTLP externo) queda como exportador 100% OPCIONAL** — config del operador o de
   un power-user, jamás instalado ni requerido por el producto. El contenedor Langfuse que ya
   corre en esta máquina es infraestructura de OBSERVATORIO propio de alpacapurpura (flota
   propia, fuera de este árbol), no una dependencia de ningún arnés instalado.
5. **Índice local reusa el patrón ya doctrinado** (`indice-desechable-jsonl-es-verdad.md`): map
   in-memory + JSON atómico en `~/.arnesia`, mismo lugar que el índice estructural (CAP-94
   `reindex-tras-turno` es el precedente de "recalcular tras cada turno").

**Por qué queda `proposed` y no se fabrica nada:** la arquitectura está resuelta y documentada,
pero CERO código nuevo se escribió en este cierre — es un paquete propio (toca UI nueva del
Mapa, necesita mockup→spec→PARIDAD como cualquier feature). Escribir `status: enforced` sin
construirlo sería el «pass fabricado» que la doctrina de honestidad prohíbe. Paquete de arranque
→ `docs/product/stories/2026-07-24-telemetria-embebida-otel/INDEX.md`.

## ⚠️ v2.1 — Corrección: la atribución por-componente NO es limpia (2026-07-26)

v2.0 afirmaba que `OTEL_LOG_TOOL_DETAILS=1` *«evita la redacción de terceros»*. **Falso para
tokens/costo.** Doc oficial, textual: *"Third-party plugin skill names are replaced with
`"third-party"`"*, e ídem `plugin.name`, **sin gate documentado** — a diferencia de `workflow.name`,
que sí dice *"unless the gate is set"*.

**Los arneses de ArnesIA se instalan como plugins de un marketplace propio ⇒ caen en «third-party» y
se colapsan a una sola etiqueta.** Amenaza la atribución por-componente incluso en Claude Code.

**Mitigación disponible, no equivalente:** `OTEL_RESOURCE_ATTRIBUTES` es respetado por Claude Code,
Copilot CLI, Codex, Goose y OpenCode — permite inyectar **nuestro** identificador de unidad de
trabajo al spawnear. Pero se fija **por proceso** ⇒ da granularidad de proceso spawneado, **no**
por-skill dentro de una sesión. Resolver en el mockup/spec del paquete.

**Dato hermano:** `claude_code.cost.usage` está documentado como **"Estimated cost"**, no facturación
real. La UI no puede presentarlo como plata gastada sin decirlo.

**Y el estándar tampoco ayuda:** `gen_ai.token.type` de la semconv GenAI admite **solo `input` y
`output`** — cache y reasoning existen únicamente como atributos de *span*, y **no hay convención de
costo en dinero**. Tampoco hay vocabulario para «skill»/«plugin»/«sub-agente». Todo eso es namespace
propio, legítimamente.

## ⚡ v2.2 — Verificación EN VIVO: cinco correcciones medidas (2026-07-26)

v2.0 y v2.1 se apoyaban en documentación. Se corrió el receptor de verdad contra `claude 2.1.220`
(3 corridas, USD 0,044) y **cinco afirmaciones del diseño no sobrevivieron**. Informe + evidencia
cruda: `docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/`.

1. **El canal primario es `/v1/logs`, no `/v1/metrics`.** El log event `claude_code.api_request`
   trae **por request** los 4 buckets de tokens, `cost_usd_micros`, `duration_ms`, `model`,
   `speed`, `query_source`, `prompt.id` y `session.id`. La métrica `token.usage` llega troceada en
   4 puntos. **Se reciben los dos endpoints; el que manda es el de logs.**
2. **El split `ephemeral_5m`/`ephemeral_1h` NO obliga a tocar el transcript.** Viene en el `result`
   del **stream-json**, canal ya sancionado por `conductor-no-parsea-jsonl.md` y ya consumido por
   el árbol. Se cierra la discusión de enmendar el boundary hermano.
3. **`collector/pdata` cuesta +10,79 MB medidos** sobre nuestra base real (`net/http` +
   `modernc.org/sqlite`), no +1,7 MB. Un decodificador OTLP/JSON con la stdlib cuesta **+0,49 MB**
   y se probó contra los payloads reales. **Como controlamos el spawn, se fuerza
   `OTEL_EXPORTER_OTLP_PROTOCOL=http/json`** (y no `http/protobuf`, como decía la recomendación
   anterior). `pdata` queda como plan B para un runtime que no deje elegir protocolo.
   ⚠️ Claude Code emite `intValue` como **número JSON** (off-spec): el decodificador usa
   `json.Number`. Las métricas llegan **Delta monotónicas** ⇒ no hay que diferenciar contadores.
4. **La redacción `third-party` es real** (reproducida con dato propio), **pero `plugin_id_hash` es
   estable entre sesiones y distinto por plugin** ⇒ la atribución **por arnés** se recupera con una
   tabla local `hash → arnés`. La atribución **por skill** no se recupera: `skill_activated` no
   trae hash. **La granularidad honesta es `arnés × caja × sesión × turno`.**
5. 🔴 **La telemetría arrastra PII en CADA punto**: `user.email`, `user.account_uuid`,
   `user.account_id`, `organization.id`, `user.id`. Ninguna versión anterior lo contemplaba. La
   ingesta persiste por **allowlist**, y el forward opcional al OTLP del operador **filtra en el
   borde** — reenviarlo crudo exportaría el email de quien corra el arnés.
   *(Dato a favor del canal: prompts y respuestas llegan `<REDACTED>` por default. La telemetría es
   menos invasiva que el transcript.)*

## ⚡ v2.3 — Segunda tanda en vivo: los hooks, y la premisa de S2 que era falsa (2026-07-26)

Misma jornada, segunda corrida: un hook que vuelca su stdin sobre los 6 eventos, más una prueba
dirigida sobre el bloque `env` de settings. Informe:
`…/verificacion-2026-07-26/ANEXO-hooks.md`. **Cuatro cosas cambian, y una de ellas es de fondo.**

1. **La llave del join es `(session_id, prompt_id)`, no `session.id` solo.** El mismo turno lleva
   el mismo identificador en los dos canales: OTel lo manda como `api_request.prompt.id`, el
   payload del hook como `prompt_id`. **El join es una igualdad de dos campos** — sin heurística
   de tiempo ni de orden — y es a nivel **turno**, que es la granularidad que v2.2 fijó. Con
   `session.id` solo, el join sería a nivel sesión y no alcanza para decir «el 60 % se va en la
   caja Y».
2. 🔴 **El payload del hook trae CONTENIDO en claro**, a diferencia del canal OTel donde llega
   `<REDACTED>`: `UserPromptSubmit.prompt` es el prompt completo, `Stop.last_assistant_message` la
   respuesta del asistente, `PostToolUse.tool_response` lo que la herramienta leyó o escribió.
   ⇒ **la allowlist de v2.2 deja de ser una regla del receptor OTLP y pasa a regir los dos
   caminos de ingesta.** El hook **proyecta a campos declarados y descarta el resto ANTES de
   escribir a ningún lado** — y `telemetria-no-egresa` necesita un hermano que verifique **qué
   campos** escribe, no solo a dónde: un hook que postea su stdin entero a `127.0.0.1` cumple «no
   egresa» y aun así filtra la conversación al almacén local.
3. **Confirmado que los hooks NO traen dinero.** Ninguno de los 6 payloads trae tokens, costo ni
   cache. La separación **hook = proceso · OTel = dinero · daemon = decisión** queda verificada,
   no supuesta. Matiz del mismo día: `tool_decision` y `tool_result` **sí** llegan por OTel, con
   `success`, `duration_ms`, `tool_input_size_bytes` y `tool_result_size_bytes` y **sin
   contenido** ⇒ parte de la señal de proceso también vive en OTel, y el detector B11 («tool
   results obesos») deja de estar bloqueado por falta de dato.
4. 🎯 **La premisa de S2 era falsa: el bloque `env` de un `settings.json` ENCIENDE la telemetría.**
   Probado con marcador distinto por variante, un solo receptor y las env vars del shell
   desarmadas, en tres formas (`--settings <archivo>` · `.claude/settings.json` del proyecto sin
   flags · lo mismo con `--setting-sources project,local`). ⇒ **S2 se parte en dos escenarios que
   son dos niveles de dato distintos**, y la UI los distingue:
   - **`s2-instrumentado`** — el arnés lleva el bloque `env` ⇒ llega **la misma señal que en S1,
     dinero incluido**. (B1 sigue sin aplicar, pero por otra razón: falta el `result` del
     stream-json, no falta la telemetría.)
   - **`s2-degradado`** — solo el hook ⇒ proceso, jamás dinero. Los detectores de dinero se
     apagan **con motivo**.

   ⚠️ **Dónde vive ese bloque `env` es decisión de PRODUCTO, abierta**: en el repo del propio
   arnés (nuestro archivo, sin fricción, cobertura parcial) o en el proyecto del usuario (escribe
   settings de un tercero ⇒ choca con A8 y con el guardrail vigente ⇒ **exige consentimiento
   explícito**). Las dos opciones con sus consecuencias, en
   `…/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md` §7.5.

   ⚠️ **No verificado, y sería el mejor mecanismo de obligación que existe:** si un **plugin**
   puede aportar un bloque `env`. Lo probado es el settings del **proyecto**. Si pudiera, el arnés
   se instrumentaría solo al instalarse sin tocar nada del usuario. La prueba que lo cierra está
   escrita (§7.5), y **necesita control positivo**: los tres primeros intentos de la prueba del
   bloque `env` dieron negativo y el negativo era **falso** — un receptor de una prueba anterior
   seguía ocupando el puerto. **Un negativo sin control no es un resultado.**

**Fail-open, que se evaluó como boundary propio y se resuelve acá.** «La instrumentación jamás
degrada al sistema instrumentado» es doctrina real, pero su único sujeto en este árbol es el hook
que este nodo gobierna: sacarla a un nodo aparte partiría una regla en dos y, generalizada sin
contexto, se leería como «tragarse los errores», que es lo contrario de la honestidad de la casa.
Queda acá, con su matiz explícito: **silencioso hacia el usuario, nunca invisible en el registro.**
El hook no rompe ni demora el trabajo; el hueco aparece después como «sin dato» y en la
conciliación de cobertura, **jamás como 0**.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| scaffold-emite-env-otel | `scaffold` inyecta `CLAUDE_CODE_ENABLE_TELEMETRY`+`OTEL_EXPORTER_OTLP_ENDPOINT`+`OTEL_EXPORTER_OTLP_PROTOCOL=http/json` (loopback) en todo arnés creado — NO un hook custom | error | «arnés nuevo nace sin telemetría» | (pendiente — no existe `scaffold`, ver BACKLOG) |
| collector-otlp-embebido-local | el daemon embebe un receptor OTLP mínimo (HTTP, **`/v1/logs` + `/v1/metrics`**) que recibe SOLO tráfico loopback — ningún proceso/contenedor externo | error | «telemetría emitida pero nadie la recibe» | (pendiente — no existe receptor) |
| jsonl-nunca-fuente-de-tokens | el índice de tokens/costo/atribución NUNCA lee del JSONL (coherencia con `conductor-no-parsea-jsonl.md`) | error | «tokens leídos parseando el JSONL» | (pendiente) |
| langfuse-jamas-dependencia-dura | ningún flujo de instalación/scaffold requiere Langfuse ni infraestructura Docker externa | error | «instalador o scaffold dependen de Langfuse» | (pendiente) |
| telemetria-por-adaptador | la regla de acumulación y la aritmética de tokens (`disjoint`/`inclusive`) son propiedad del ADAPTADOR de runtime, nunca del agregador central | error | «tokens sumados con la regla de otro runtime» | (pendiente — v2.1) |
| telemetria-sin-pii | la ingesta persiste por **allowlist en los DOS caminos** (OTLP y hook); `user.email`/`user.account_*`/`organization.id` y el contenido de la conversación nunca llegan al almacén ni al forward opcional | error | «la telemetría guardó o exportó datos de cuenta o contenido» | (pendiente — v2.3; doctrina en `ingesta-por-allowlist-declarada.md`) |
| cero-post-install | ninguna pieza de telemetría se descarga en post-install: todo compilado en el binario (A) o shipeado como sidecar (B) | error | «el instalador baja algo de internet» | (pendiente — v2.1) |
| hook-es-fail-open | el hook del arnés jamás rompe ni demora el trabajo: exit 0 siempre, stdout vacío, tope de tiempo declarado, sin reintentos | error | «la instrumentación bloqueó un turno del usuario» | (pendiente — v2.3, TestHookNoTardaNiFalla) |
| hook-proyecta-campos | el hook **no reenvía su stdin**: emite solo los campos declarados. No alcanza con verificar el destino, hay que verificar el contenido | error | «el hook filtró la conversación al almacén local» | (pendiente — v2.3, TestHookNoReenviaContenido) |
| escenario-se-deriva | el nivel de instrumentación (`s1`/`s2-instrumentado`/`s2-degradado`) lo DERIVA el receptor de la señal que llegó; un arnés no puede declararlo | error | «un arnés se declara mejor medido de lo que está» | (pendiente — v2.3, TestEscenarioSeDerivaDeLaSenal) |

## Changelog

- 2026-07-26 · v2.3 · **Segunda tanda de verificación en vivo (ANEXO-hooks.md), mismo día.**
  (1) La llave del join pasa a ser **`(session_id, prompt_id)`** — está en los dos canales, el
  join es una igualdad de dos campos y es a nivel turno. (2) 🔴 El payload del hook trae la
  conversación **en claro** ⇒ la allowlist rige los dos caminos de ingesta, y `telemetria-no-egresa`
  gana el check hermano `hook-proyecta-campos`. (3) Confirmado que los hooks no traen dinero;
  `tool_decision`/`tool_result` sí llegan por OTel con tamaños en bytes y sin contenido
  (desbloquea B11, que no entra al MVP). (4) 🎯 **La premisa de S2 era falsa**: el bloque `env`
  de un `settings.json` enciende la telemetría ⇒ S2 se parte en `s2-instrumentado` (misma señal
  que S1, dinero incluido) y `s2-degradado` (solo hook). Dónde vive ese bloque queda **abierto
  como decisión de producto**; si un *plugin* pudiera aportarlo —sin verificar— sería el mecanismo
  de obligación ideal. Se absorbe además la doctrina de **fail-open** que se había evaluado como
  boundary propio (justificación en L2 §v2.3). 3 checks nuevos (**10 en total**) y dos
  reformulados. Sigue `proposed`: cero código. Diseño completo del módulo en
  `docs/product/stories/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md`.

- 2026-07-26 · v2.2 · **Verificación en vivo** (paquete `stories/2026-07-24-telemetria-embebida-otel/`,
  decisiones D14-D16). 5 correcciones medidas contra `claude 2.1.220`: canal primario `/v1/logs`;
  el split 5m/1h no exige tocar el transcript; `pdata` cuesta 6× lo estimado ⇒ OTLP/JSON + stdlib y
  `http/json` en el spawn; `plugin_id_hash` rescata la atribución por arnés; **PII en cada punto**.
  1 check nuevo (7 en total). Sigue `proposed`: cero código.

- 2026-07-26 · v2.1 · **Corrección + multi-runtime** (investigación SOTA, 3 carriles — paquete
  `stories/2026-07-24-telemetria-embebida-otel/`, decisiones D9-D11). (1) **Corregida una afirmación
  falsa de v2.0**: `OTEL_LOG_TOOL_DETAILS` NO evita la redacción `"third-party"` de
  `skill.name`/`plugin.name`, y nuestros arneses caen ahí ⇒ la atribución por-componente está
  amenazada; mitigación parcial vía `OTEL_RESOURCE_ATTRIBUTES`. (2) **OTel no es el canal universal**:
  de 6 runtimes medidos solo 3 dan OTel útil en headless, contra 5 que dan `stream-json` por turno
  (Codex `exec` no emite métricas; Amp y Cursor no tienen OTel) ⇒ el adaptador base es stream-json,
  OTel es enriquecedor. (3) La semconv GenAI **no modela cache ni reasoning en métricas** ni define
  costo en dinero. 2 checks nuevos (6 en total). Sigue `proposed`: cero código.
- 2026-07-24 · v2.0 · **Arquitectura resuelta** (orden del operador, barrido de deuda viva
  HS-27): hallazgo del emisor legacy `emit.py`/KIT-03 (probado, pero canal JSONL+Langfuse
  incompatible con "instalable" + `conductor-no-parsea-jsonl.md`) + verificación oficial de la
  telemetría OTel nativa de Claude Code (`claude_code.token.usage`/`claude_code.cost.usage` +
  atributos `skill.name`/`tool_name`, sin hook custom) → resuelve el diseño: receptor OTLP
  embebido loopback-only en el daemon + scaffold inyecta env vars + JSONL solo enumerar/replay +
  Langfuse 100% opcional, nunca dependencia del producto. 4 checks (reemplazan los 2 de v1.0).
  Sigue `proposed`: cero código nuevo, paquete de arranque en `stories/2026-07-24-telemetria-embebida-otel/`.
- 2026-07-23 · v1.0 · Nodo fundacional — draft de
  `docs/product/research/2026-07-05-arquitectura-inyeccion-knowhow.md` §9 (HS-07) materializado
  como boundary formal (deuda BACKLOG «3 boundaries de research → arch/»). Investigación
  confirmó CERO implementación — nace `proposed`, honesto, ligado a la deuda BACKLOG «telemetría
  JSONL → indexer real» (mismo trabajo pendiente, no duplicar el esfuerzo cuando se construya).
