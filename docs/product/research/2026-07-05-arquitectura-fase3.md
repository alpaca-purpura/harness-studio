# Investigación de arquitectura — fase 3 (HS-04, 2026-07-05)

> **5 frentes en paralelo (subagentes), verificados contra la web vigente (julio 2026).**
> Es la **evidencia L1** que respaldan los boundary nodes de [`../arch/`](../docs/architecture/INDEX.md).
> Condensado: lo decisivo para la propuesta. Prioridad de fuentes = docs oficiales
> (code.claude.com/docs, platform.claude.com) › estándares/proyectos maduros › expertos.
> Norte: [`../vision.md`](../vision.md) (decisiones técnicas fundacionales HS-02).

## Frente A — Shell de escritorio y empaque multiplataforma

**Veredicto:** el core no cambia según el shell — `serve` expone el producto como HTTP/SSE
`:4200` + SPA `go:embed`; ese contrato hace el shell intercambiable. **Decisión del operador
(HS-04): shell v1 = Tauri 2** (no browser-first), con el daemon como **sidecar `externalBin`**;
la misma build de Vite corre en el WebView y consume la misma API. Tauri añade solo ventana
nativa, tray, auto-updater, deep-link, firma. La ruta a la «app vendible» es aditiva (wrap), no
rewrite — consistente con VISION.

**Gotchas de Linux WebView (muerden primero en Mint = Ubuntu 24.04 base, Cinnamon, kernel 6.8;
ahora load-bearing porque elegimos Tauri desde v1):**
- **Cliff 4.0→4.1:** Ubuntu 24.04 dropeó `libwebkit2gtk-4.0`; **Tauri v1 muere en Mint 22 →
  Tauri 2 obligatorio** (linka 4.1). #1 bug «la app no abre» en el mundo real.
- **NVIDIA + DMABUF = ventana negra/basura.** Mitigación: `WEBKIT_DISABLE_DMABUF_RENDERER=1`
  (y/o `WEBKIT_DISABLE_COMPOSITING_MODE=1`) desde el launcher. Portátiles Mint+NVIDIA comunes.
- **Ghost-DOM al maximizar/restaurar** en webkit2gtk-4.1 2.48.x tras `apt upgrade` — la MISMA
  app regresa cuando el distro bumpea WebKit debajo. No controlamos la versión; la controla el
  distro. Riesgo real para un canvas pesado (React Flow 150 nodos + SSE) de larga duración.
- **Wayland vs X11** (Cinnamon X11-default hoy, migrando): compositing/HiDPI/IME difieren.

**Lifecycle:** «attach si el daemon está arriba, si no spawnea» (mantiene headless gratis) ·
Tauri **no auto-reapea** el sidecar en crash duro → heartbeat/supervisor · single-instance
plugin (ventana) **+ lock propio del daemon** (`:4200` bind-or-bail / flock) · deep-link
`arnesia://` vía plugin (ojo bug single-instance+deep-link #12726) · auto-update: **el daemon
se auto-actualiza por package manager** (GoReleaser → apt/brew/scoop); el updater de Tauri
firma el wrapper. Wails v3 = **alpha** en jul-2026 (sin GA) y disolvería el core headless en su
bridge → descartado. Electron descartado (bundle 100–250 MB).

Fuentes: v2.tauri.app/develop/sidecar · v2.tauri.app/develop/debug/linux-graphics · tauri
issues #9662 (cliff 4.0/4.1), #13157 (ghost-DOM), #12726 (single-instance+deep-link),
discussion #8524 · v2.tauri.app/plugin/{updater,single-instance,deep-linking} · wails issue
#5052 (v3 alpha, sin GA) · Phoronix Mint 22 = Ubuntu 24.04.

## Frente B — Conexión al Claude Code local (el conductor)

**Veredicto:** **Go spawnea el `claude` local como subproceso y habla stream-json por
stdin/stdout.** Rank: subproceso-conductor › SDK-sidecar ›› managed agents. Razón decisiva: el
Agent SDK es un wrapper que spawnea el MISMO binario y habla el mismo protocolo NDJSON — no
aporta nada que Go no pueda, y mete runtime Node en una distribución Go. El TS SDK **trae su
propio `claude` bundled** → manejaría un Claude bundled, no el install del usuario (pelea con
«manejar el install LOCAL»). Managed Agents (REST hosted, beta GA 2026-04-08) corre en sandbox
remoto → **no toca `~/.claude` ni el filesystem** → descartado.

**Dos correcciones load-bearing al nodo [`headless-sdk`](../docs/architecture/knowledge/elements/headless-sdk.md):**
1. **`--bare` ROMPE el auth de suscripción.** `--bare` salta OAuth/keychain y exige
   `ANTHROPIC_API_KEY`/`apiKeyHelper`. Un conductor que monta el login Pro/Max del operador NO
   puede default-earlo. Úsalo solo cuando estás deliberadamente sobre API key. Y `-p` hará
   `--bare` el default futuro → **pinear flags explícitos, no heredar defaults de `-p`.**
2. **No existe señal de «% contexto»** en stream-json ni OTel. Se computa:
   `(input + cacheRead + cacheCreation) / ventana del modelo`. Es métrica derivada nuestra.

**Loop vivo multi-turno (sanctioned):**
```
claude -p --input-format stream-json --output-format stream-json --verbose \
  --session-id <UUID-por-conversación> --include-partial-messages \
  --permission-prompt-tool stdio --max-turns <N> [--model …] [--add-dir <arnés>]
```
`--input-format stream-json` = UN proceso vivo leyendo turnos NDJSON de stdin (modo recomendado
para interactivo/multi-turno; soporta cola, interrupts, imágenes, permisos) · `--session-id`
que generamos y pineamos · **nunca `--continue`** bajo concurrencia (carga «la más reciente en
cwd» → race) · `--fork-session` para variantes · `--max-turns` = kill-switch (en `-p` sale con
error al toparse) · **stdin cap 10 MB** (v2.1.128) → specs/imágenes grandes por path, no inline.

**Tipos de mensaje stream-json (fuente de eventos live):** `system/init` (session_id, model,
tools, mcp_servers, apiKeySource, cwd, plugins) · `assistant` (turno completo; **skill activada
= un `tool_use` llamado `Skill`**) · `user` (tool_result) · `stream_event` (deltas: ttft_ms,
text_delta, input_json_delta, usage) · `result` (total_cost_usd, usage, num_turns,
`permission_denials`) · `system/api_retry` (attempt, error: authentication_failed/rate_limit/
overloaded — **detector de auth/rate**) · **`system/compact_boundary`** (compaction SÍ está en
stream-json — no hace falta JSONL).

**Event sourcing en 3 fuentes (sin caer en la trampa del JSONL):**
- **stream-json = live/primario** → turnos, tool_use/result, deltas de tokens, TTFT, costo,
  permission_denials, compaction, api_retry.
- **OTel = sidechannel** (`CLAUDE_CODE_ENABLE_TELEMETRY=1`, endpoint → el daemon) para lo que
  stream-json no limpia: `hook_*` (los hooks corren DENTRO del CLI, **solo OTel los ve**),
  `skill_activated` (nombre limpio), `tool_decision` (auditoría de permisos), `compaction`,
  `mcp_server_connection`, `auth`. Métricas `token.usage`/`cost.usage` por type/model/
  query_source. **Sin evento de slash-command dedicado** (va en `user_prompt` con command_name;
  igual lo inyectamos nosotros).
- **JSONL = enumerar/replay** de sesiones pasadas; **nunca parsear su schema como API estable.**
  Capturamos el stream a NUESTRO event store → ese store es la API estable. (Coherente con
  VISION «SQLite desechable, JSONL fuente de verdad» y la trampa ya vivida en UX it.7.)

**Discovery del binario (Linux-first):** override de config (siempre 1º) → `exec.LookPath` →
nativo `~/.local/bin/claude` → Homebrew `/opt/homebrew|/usr/local` → npm global (`npm prefix
-g`) → confirmar con `claude --version`. **Auth:** subproceso como el mismo usuario **hereda el
login** (Linux `~/.claude/.credentials.json` 0600; macOS Keychain; respeta `CLAUDE_CONFIG_DIR`).
**Ojo:** `ANTHROPIC_API_KEY` en el env del hijo **pisa** la suscripción OAuth → si el usuario es
Pro/Max, **scrub la key del env del hijo**. Detección limpia: no-instalado = ENOENT · instalado
= `--version` ok · no-logueado = probe `claude -p "ok" --output-format json --max-turns 1` y leer
`apiKeySource` del `system/init` / `api_retry:authentication_failed`. **macOS daemon**: leer
Keychain OAuth puede pedir unlock (issue #9403) — Linux (archivo 0600) limpio; testear macOS.

**Permisos = GUI es el human-in-the-loop:** modo `default` (ahora «Manual»), deny-by-default,
NO pre-aprobar Write/Edit. Claude pide edit → `control_request:can_use_tool` con el input
(`file_path`/`old_string`/`new_string`) → **pintamos el diff nosotros** → `control_response`
allow (echo `updatedInput`) / deny (con `message`). Orden de evaluación fijo: PreToolUse hook →
deny → ask → mode → allow → canUseTool → PostToolUse (los auto-aprobados **no** llegan al
callback → `--allowedTools` solo read-only genuino). Modos por fase: `plan` en grill/spec ·
`dontAsk` en evals-gate/promote (sin humano) · **`bypassPermissions` NUNCA**. Paths protegidos
(`.git`, `.claude`, `.mcp.json`, dotfiles) nunca auto-aprobados salvo bypass = guardrail gratis.
Decisión PreToolUse **`defer`** (2026): deja salir el proceso y reanudar de la sesión persistida
mientras el humano tarda en aprobar (no sostener N procesos bloqueados).

**Concurrencia/lifecycle:** N sesiones = N procesos, cada uno su cwd + su **process-group**
(`setpgid`; matar el grupo → mueren Bash/MCP hijos) · aislamiento = cwd separado (worktrees para
git concurrente; **race #34645 en `.git/config.lock`** → serializar creación de worktrees ·
worktree no-interactivo **no se auto-limpia**) · clase SingletonLock: todos comparten `~/.claude`
→ contención de escritura bajo paralelismo (no documentada → **load-test**; `CLAUDE_CONFIG_DIR`
por-lane solo como último recurso, parte el auth) · backpressure: consumir stdout NDJSON continuo
o el pipe lleno bloquea el hijo.

**Semi-documentado → verificar contra el binario instalado antes de construir (primer spike
fase 4):** el protocolo `control_request`/`can_use_tool` (`--permission-prompt-tool stdio`,
shape JSON-RPC exacto) — oficial solo para el SDK, issue #24594 abierto. Referencia:
`Roasbeef/claude-agent-sdk-go/docs/cli-protocol.md`. **ToS:** manejar el login local del propio
usuario ≠ «ofrecer login de claude.ai» (cláusula del SDK) — distinto y a favor, pero confirmar
con legal si ArnesIA se distribuye a terceros. **Versión vigente:** v2.1.199 (2026-07-02);
v2.1.200 añade alias `manual`/label «Manual» para `default`.

Fuentes: code.claude.com/docs/en/{headless, cli-reference, agent-sdk/overview,
agent-sdk/streaming-vs-single-mode, agent-sdk/streaming-output, authentication, permission-modes,
agent-sdk/permissions, agent-sdk/user-input, monitoring-usage, worktrees} · anthropics/claude-code
issues #24594, #34645, #9403 · Roasbeef/claude-agent-sdk-go/docs/cli-protocol.md.

## Frente C — Contrato del dock (UI agéntica de streaming)

**Veredicto:** **conformar a la taxonomía AG-UI sobre SSE** (subset text/tool/lifecycle +
`CUSTOM`), **emisor Go de primera mano** (no depender del SDK Go community, que es tier-community
sin scaffolding de handler HTTP; emitir `data: {json}\n\n` es trivial). AG-UI (CopilotKit): 16
eventos en 5 grupos, **transport-agnóstico (SSE default), backend-agnóstico** (sin dependencia
de LangGraph — el framing de que asume backends LangGraph está **desactualizado**). Lo pesado
(STATE_DELTA colaborativo, binario, webhooks) es **opt-in** → no lo emitimos. Los eventos
CC-específicos (skill_activated, hook, /command, compaction, header de sesión, diffs, CTAs) van
como **`CUSTOM`** (y `STATE_*` para el header como estado vivo). Razón decisiva: hace el
consumidor React casi gratis (assistant-ui trae runtime AG-UI) + **interop futura** (AWS Bedrock
AgentCore, LangGraph, Google ADK, CrewAI, Mastra ya hablan AG-UI) a costo ~0 → importa para
DevHub/marketplace.

**Capa React del dock:** **assistant-ui** (MIT, «forever free», React 18/19, **sin Next**, corre
en Vite SPA, bring-your-own-backend). `@assistant-ui/react-ag-ui` apunta a cualquier endpoint
SSE AG-UI **incl. servidores no-JS/Go**. Filas de sistema/header/diff = **`data` parts** que
proveemos (modelo allowlist = component-selection, exactamente nuestra restricción).
`makeAssistantToolUI` mapea tool calls → componentes. Out-of-box: streaming markdown, highlight,
autoscroll, retries, a11y. Fallback si NO emitimos AG-UI: `LocalRuntime`/`ExternalStoreRuntime`
(SSE bespoke) → assistant-ui es correcto en cualquier caso, de-riskea el frente. **Costo:** es
thread/mensaje-céntrico → filas no-conversacionales densas van como `data` parts DENTRO del
timeline; validar ergonomía en un spike de 1 día. **No AI Elements** (asume Next + wire-format
del AI SDK). **No hand-roll** (reconstruir streaming/markdown/autoscroll por nada).

**Diff:** **`@codemirror/merge`** (`unifiedMergeView` inline; `allowInlineDiffs` word-level) —
reusa el stack CM6 ya elegido para el editor inline → un solo árbol de dependencias, look
consistente entre «editar» y «revisar diff».

**Component-selection, sin ambigüedad.** Format-authoring (A2UI/mdocUI/Thesys C1) solo gana
cuando el SET de UI es abierto y compuesto por el modelo — NO intersecta ArnesIA (dock = set fijo
de tipos app-owned sobre stream tipado). Adoptar un DSL aquí añade parser + superficie de
render/validación/seguridad + no-determinismo por **cero** beneficio. **Verdicto: trampa. El
modelo no autora UI del dock.** Ortogonal a Tauri (mismo stream en el WebView) y a
marketplace/DevHub (upside: interop AG-UI sin capa de traducción).

Fuentes: docs.ag-ui.com/concepts/architecture · github.com/ag-ui-protocol/ag-ui (sdk/go
overview — flag community) · assistant-ui.com/docs/runtimes/ag-ui + runtimes/custom + primitives/
message-part (MIT, data parts) · npmjs/@codemirror/merge · developers.googleblog A2UI v0.9
(2026-07) · copilotkit.ai/blog AgentCore AG-UI (2026-03) · techcrunch 2026-05 (CopilotKit).

## Frente D — Local-first: watch, índice, transporte, grafo

Los 4 picks pre-hechos **aguantan.**
- **Watch:** fsnotify da señal «sucio»; **reader propio por byte-offset** (no dejar que fsnotify
  lea). Cursor `{path, inode, size, offset, partialLineBuf}` en meta table; retener bytes sin
  `\n` hasta el próximo evento. **Rotación:** `size < offset` → truncado, releer de 0; cambio de
  inode → archivo nuevo. Tratar CREATE/WRITE/RENAME como «releer» y **re-adjuntar el watch** tras
  rename (inotify dropea el watch al mover el inode). Debounce 50–150 ms. inotify **no recursa** →
  `WalkDir` + `Add()` cada subdir + watch `IN_CREATE` para dirs nuevos. Manejar `IN_Q_OVERFLOW`
  (bursts → pierdes eventos → rescan de cursores). Rebuild rápido: una transacción,
  `synchronous=OFF`+`journal_mode=MEMORY` durante el bulk, luego WAL/NORMAL; persistir cursores
  para reanudar incremental; en mismatch de schema-version → `rm` y reconstruir.
- **Índice:** **modernc.org/sqlite confirmado** (puro-Go/CGO-free = un binario cross-compile).
  Bench 2026-03-23 Go 1.26: modernc ~1.5–2× más lento en inserts pero competitivo/mejor en
  queries y en concurrente 8-goroutine; el gap de escritura es irrelevante para carga
  read-dominant. **DuckDB rechazado** (`go-duckdb` exige CGO → mata el binario estático; si años
  de telemetría lo exigen, DuckDB out-of-process sobre Parquet, nunca el core). WAL: `journal_
  mode=WAL, synchronous=NORMAL, busy_timeout=5000` + **2 `*sql.DB`**: writer `SetMaxOpenConns(1)`
  (serializa escrituras, evita SQLITE_BUSY) + reader pooled; `BEGIN IMMEDIATE` en txns de
  escritura. **Migraciones = no migrar, reconstruir** (schema-version en meta table; mismatch →
  borrar y re-indexar desde JSONL = feature del diseño).
- **Transporte:** **SSE confirmado** sobre WebSocket — reconexión + `Last-Event-ID` = «índice
  desechable, JSONL es verdad, replay el stream»; client→server raro = POST plano; corre trivial
  en el server `go:embed` y el WebView Tauri. **UNA conexión multiplexada** con eventos tipados
  (`event: map|dock|run`) demux client-side. Gotcha localhost: el cap de 6-conexiones/origen es
  real bajo **HTTP/1.1** (localhost plano ES HTTP/1.1; Go sirve h2 a browsers solo con TLS) → el
  multiplexado a 1 conexión lo esquiva; **no añadir TLS localhost solo por h2**.
- **Grafo:** **React Flow 12 confirmado** (v12.11.x; 150 nodos = zona dulce, el «muro» es 5k+).
  Memoizar cada nodo/edge (`React.memo` fuera del render padre) · suscribir por selector
  (`useStore(selector)`), nunca leer el array completo en hijos · selección por **CSS**
  (`.selected`), no mutando `data` · `onlyRenderVisibleElements` **saltar** a esta escala (puede
  añadir costo al panear) · **conmutar las 4 capas SIN remount** (toggle `hidden`/`data.layer`/
  CSS — remount = el cliff real) · edges: `zIndex` explícito, evitar SVG animado en 100+ ·
  **mantener el layout de carriles custom (~200 líneas)**; elkjs solo si el sub-grafo Base lo pide.
- **Estado FE + deep-link:** **Zustand + hash-sync a mano** (tupla propia
  `arnés·vista·capa·sel·run·replay` como blob compacto en `location.hash`; hidratar store al
  init; `store.subscribe()` → escribir hash con **`history.replaceState`** + debounce). Usar
  **hash, no query** (no round-trip al server, corre en Tauri). SSE → **reducer al store**
  (deltas a un grafo vivo, no cache request/response; TanStack Query **no** tiene SSE built-in →
  reservarlo para endpoints discretos). nuqs v2 (ya tiene adapter plain-React SPA) = fallback si
  no queremos hand-roll; TanStack Router = overkill (una sola vista con estado, no árbol de rutas).

**Gotchas Mint primero:** **inotify watch limit** (`fs.inotify.max_user_watches` ~8k default;
`~/.claude` con muchos subdirs lo agota) → `/etc/sysctl.d/99-arnesia.conf`
`fs.inotify.max_user_watches=524288` + `sysctl --system`; **detectar el fallo y avisar, no morir
en silencio** · `max_user_instances` (128), `max_queued_events` (16384) · nunca NFS/SMB
(`~/.claude` es local) · CGO-free construye en Mint limpio sin `build-essential`.

Fuentes: cvilsmeier/go-sqlite-bench (2026-03-23, Go 1.26) · duckdb/duckdb-go issue #243 (CGO) ·
sqlite.org/wal.html · websocket.org/comparisons/sse · go.dev/doc/go1.24 (UnencryptedHTTP2) ·
reactflow.dev/learn/advanced-use/performance (v12.11.1) · nuqs.dev/blog/nuqs-2 · TanStack/query
#7581 (streamedQuery experimental) · pkg.go.dev/fsnotify · nxadm/tail.

## Frente E — Arquitectura as code (cómo vive en el repo, no como doc que se pudre)

**Stack mínimo de alta palanca (sesgo low-maintenance, equipo chico, Go+React):**
- **Decisiones:** frontmatter **MADR 4.0** como *proyección delgada del LEDGER* — no un log
  paralelo. La ficha `HS-NN` sigue canónica; el boundary node lleva `ledger:` + `enforced_by:`.
  MADR completo (`arch/decisions/`) solo cuando una decisión necesita el tratamiento de
  opciones-consideradas que una ficha no aguanta. (Patrón `zircote/structured-madr`: MADR
  machine-readable + JSON Schema + GitHub Action validator.)
- **Diagramas:** **Mermaid inline** (renderiza gratis en GitHub, cero tooling) para el 80% +
  **D2** (Terrastruct, **binario Go único**, layout dagre/ELK) para las 2-3 vistas C4 container
  reales. **Skip Structurizr** (JVM/Docker) y **TypeSpec** (Node, sin emitter Go oficial) hasta
  que un need concreto aparezca. Todo diffea en git y renderiza on-demand.
- **Fitness functions (la parte ejecutable que rompe CI):** **go-arch-lint** (v1.15.0
  2026-05-04, config `version: 3` en `.go-arch-lint.yml`: componentes por glob + `deps`
  mayDependOn) para el grafo de imports (core⊥shell, dominio⊥transporte, adaptadores solo desde
  la composition-root) + **depguard** (en golangci-lint) para bans duros de import (dominio no
  importa `net/http`/`database/sql`). Lo que el linter no expresa (swap de adaptador) = un
  `arch_test.go` a mano. Inspiración: ArchUnit/Spring Modulith/jMolecules (Q1 2026),
  dependency-cruiser (TS, usable en la SPA).
- **Schema-first (el dominio ES el spec vivo):** **JSON Schema 2020-12** como single source para
  `meta.clase` L0 (I-75) y el `contract:` de caja (METODOLOGIA §3) — el frontmatter YAML de cada
  caja es una *instancia*. **quicktype** (2026-03-26) genera tipos **Go y TS** de un schema;
  **google/jsonschema-go** (Google Go team, ene-2026, production-grade) valida en runtime + en el
  linter. **Validar el `contract:` contra su schema ES la fitness function del dominio** = el
  eval-gate A4 y la detección de huérfanos/mismatch hechos ejecutables. API HTTP/SSE (concern
  aparte): **OpenAPI 3.1** + **oapi-codegen/v2** (Go) + **openapi-typescript** (TS).
- **Un solo runner:** subcomando `arnesia conformance` corre `docs/architecture/knowledge/` (121 checks de
  metodología — 121 al momento de escribir → 122 tras la corrección de headless) **y**
  `docs/architecture/fitness` (go-arch-lint + validación de schema) juntos, mismo reporte
  severidad+señal. Literalmente «la conformidad de arquitectura corre como linter igual que el
  árbol de metodología».

**Árbol propuesto (espeja `docs/architecture/knowledge/`):** ver [`../docs/architecture/INDEX.md`](../docs/architecture/INDEX.md). Cada
boundary node = frontmatter (`version/updated/status/ledger/sources/enforced_by/severity`) + **L1**
(principio con fuente) + **L2** (realización en este árbol Go) + tabla de checks (severidad +
señal). Separación limpia: check ejecutable → `fitness/` · diagrama → `model/` (render on-demand,
nunca hand-sync) · el «por qué» → LEDGER · regla viva + mapeo de enforcement → `boundaries/` (la
única prosa a mano, mínima).

Fuentes: adr.github.io/madr (MADR 4.0.0) · github.com/terrastruct/d2 · github.com/fe3dback/
go-arch-lint (v1.15.0) · pkg.go.dev/OpenPeeDeeP/depguard · github.com/glideapps/quicktype ·
opensource.googleblog google/jsonschema-go (ene-2026) · github.com/oapi-codegen/oapi-codegen/v2 ·
zircote/structured-madr · platformtoolsmith.com (ADR→fitness).

## Flags de confianza (no confirmado / verificar)

- **Shape JSON-RPC de `control_request`/`can_use_tool`** — semi-documentado (oficial solo SDK,
  issue #24594) → **spike contra el binario instalado antes de construir el dock** (frente B).
- **Versión exacta de Tauri 2** (GA desde 2024-10; ~v2.11.x mid-2026, maduro) y de D2/Structurizr
  — confirmar semver antes de pinear.
- **Contención de escritura en `~/.claude` compartido** bajo alta concurrencia — no documentada →
  load-test (frente B/D).
- **macOS daemon + Keychain OAuth** (unlock prompts #9403) — Linux limpio; testear macOS.
- **ToS** subproceso-del-login-propio vs cláusula SDK «no ofrecer login claude.ai» — legal si se
  distribuye.
- **TypeSpec sin emitter Go oficial** (community only); «YADR» YAML-ADR mencionado mar-2026, poco
  verificado.
