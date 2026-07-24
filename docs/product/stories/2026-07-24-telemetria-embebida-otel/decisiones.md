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

## Estado del paquete

**Diseño de arquitectura RESUELTO y documentado** (boundary v2.0 + esta ficha). **CERO código
construido en este cierre** — es correctamente un paquete propio: toca UI nueva del Mapa (capa
Tokens, hoy sin ningún diseño visual — ni siquiera el mockup "destino" la dibuja, ver
`mockups/arnesia-mapa-destino.html:295`, "Fase 2") y un componente Go nuevo (receptor OTLP). Por
disciplina §10 (feature nueva = mockup→spec→PARIDAD), NO se codea salteando ese proceso. Este
paquete es el punto de arranque cuando se retome: falta mockup de la capa Tokens (granularidad
per-skill, dónde vive el número — ¿badge en el nodo? ¿panel del inspector?) → spec del receptor
OTLP + del scaffold de env vars → build → PARIDAD.
