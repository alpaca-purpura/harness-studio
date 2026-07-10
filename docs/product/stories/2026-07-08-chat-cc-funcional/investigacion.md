# Investigación — chat CC completamente funcional (4 frentes)

> Paquete `2026-07-08-chat-cc-funcional` · agentes: ① GitHub CC-UIs · ② assistant-ui/AG-UI
> · ③ repo as-is · ④ doctrina/scoping. Frentes ③ y ④ cerrados; ① y ② se anexan al llegar.

## Frente ③ — Qué tenemos YA (no empezamos de cero: ~60%)

### Pipeline backend vivo (se reusa entero)

- **Conductor** `internal/adapters/agent/claudecode/conductor.go` — proceso `claude`
  persistente multi-turno: `-p --input-format stream-json --output-format stream-json
  --include-partial-messages --verbose` + `--resume` + `--model` + `--max-turns` +
  inyección (`--plugin-dir`/`--append-system-prompt-file`/`--add-dir`) + permisos
  (`permissionArgs:117`: `--permission-mode default`, `--allowedTools` solo lectura —
  Write/Edit/MultiEdit/NotebookEdit SIEMPRE filtrados a aprobación GUI—,
  `--disallowedTools`, `--permission-prompt-tool stdio`). `cmd.Dir = opts.Cwd` (S2).
  `translate:398` normaliza wire→`AgentEvent` (init/delta/message/result/control_request).
- **SSE** `internal/adapters/transport/sse/broker.go` — 1 conexión `GET /events`,
  3 canales `map|dock|run`, `Last-Event-ID` replay, ring 256. OJO: taxonomía = frames
  propios (`dockFrame`), NO AG-UI todavía.
- **dockFrame** `session_service.go:45-58` — `kind ∈ status|init|delta|message|result|
  error|permission|permission_result` + `request_id/tool/input/decision/ctx_pct/...`.
- **Sesiones** — N:1 arnés, persistidas (`~/.arnesia/sessions.json`), `ErrBusy` 409
  un-turno-a-la-vez, resume auto-sana (`tryHealResume:420`), estados
  `streaming|await|idle` de dominio.
- **Permisos Fase E COMPLETOS (Go)** — `onControlRequest:457` (grant vigente →
  auto-allow; si no → pending + `await` + frame `permission`) ·
  `ResolvePermission:525` (deny-del-rol gana SIEMPRE, allow acuña Grant TTL,
  operador solo ESTRECHA TTL) · roles spike `permission/provisioner.go`
  (backend-dev 15m / reviewer 5m / base deny TTL 0).
- **Fuente por nodo** — `FuenteService.Fuente(harnessID,nodeID)` resuelve
  `fuente_path` confinado al dir del arnés → `GET /api/harnesses/{id}/nodes/{nodeId}/fuente`.
  Base directa del **chip de alcance** (decisión #1).
- **Gate** — `GET /api/harnesses/{id}/conformance` (RunGraph: manifiesto-schema ·
  contrato-caja-schema · spine · escritor-único · firewall no-phantom-frontmatter).
- **T3** — `POST /api/harnesses/{id}/boxes/{boxId}/run` síncrono con `Permisos` del
  rol del arnés (`run_service.go:124`). Sin caller FE.

### FE as-is

- Chat hand-rolled funcional E2E: `widgets/chat-dock` (burbujas, deltas en vivo,
  typing dots, ctx bar) + `shared/store/sessions-store.ts` (init→token→sessions→SSE,
  optimistic turn, 409). Rail multisesión completo. ⌘K toggle.
- **Sin instalar:** `@assistant-ui/react`, `@codemirror/*`, AG-UI. Deps actuales:
  React 19.2, Zustand 5, Base UI 1.0-rc, Tailwind 4, lucide.
- Selección de nodo alimenta SOLO el Inspector (`workspace-stage.tsx`) — no llega al
  composer.

### Gaps a "chat completamente funcional"

1. FE ignora `permission|permission_result` (`types.ts:60` ni los tipa; `onDock` sin
   case; sin `api.resolvePermission`). Humano-en-el-loop OSCURO en FE.
2. `spawnLocked:307` NO pasa `Permisos` al Dock ⇒ sesiones interactivas sin
   `--permission-prompt-tool stdio`; control_request quizá nunca dispara en Dock
   (solo T3 lo pasa). Fix chico.
3. Wire `control_response` verificado solo con fakes (`conductor.go:250-259`) —
   spike contra binario real = deuda HS-04 reservada.
4. Tarjetas ricas inexistentes: diff Edit/Write, Bash output, TodoWrite, thinking,
   Task/subagent, costo/tokens.
5. Chip de alcance (decisión #1) no cableado; gate card (decisión #4) no cableada.
6. Watcher stub (`watch/watcher.go`) — editar no refresca Mapa (`main.go:218` loop
   vacío, TODO fase 5).
7. `GET /runs` 501 · `sse.ts` solo escucha `dock` (ni `map` ni `run`).

## Frente ④ — Doctrina como arneses propios + scoping (todo el mecanismo existe)

- **Kit (cuerpo ②)** = plugin `arnesia-kit` v0.1.0: overlay `kit/doctrine.md`
  (A1–A7, contrato fusionado §3, honestidad, firewall, frontera §9) + 2 meta-skills:
  `forjar-caja` (crear caja conforme: lee `arnes.l0.json`, clasifica 3 ejes, escribe
  `skills/<id>/SKILL.md`, gate honesto, edges, corre conformance) y `auditar-arnes`
  (auditoría read-only + plan de mejora). **El "arnés que modifica arneses" YA existe**
  — chico; crecerá con skills tipo `mejorar-caja`, `renombrar-estado`, etc.
- **Inyección** (`provision/provisioner.go`): huella sha256 → `~/.arnesia/{kit,
  doctrine.md,knowhow}` → flags al spawn. Degradación honesta si falla (spawn sin
  doctrina, jamás bloquea). Knowhow = 12 checklists de `docs/architecture/knowledge/elements/`.
- **Regla dura §9: ② jamás se escribe en ③** (METODOLOGIA.md:385-408; también en el
  overlay y en «Prohibido» de forjar-caja).
- **Scoping disponible:** cwd por arnés (`WorkdirResolver` + registry validado
  anti-dirs-protegidos) · `--add-dir` · `--plugin-dir` · system-prompt-file. Sesión↔
  arnés FIJA: cambiar de arnés = cambiar/crear sesión, jamás re-apuntar una viva.
- **Nodo→archivo:** tabla clase→ubicación (nomenclatura v1.1 §3) — skill=
  `skills/<id>/SKILL.md`, hook=`hooks/hooks.json`, rule=`CLAUDE.md`, mcp=`.mcp.json`,
  etc. Loader hoy reconoce skills+rules (resto TODO honesto).
- **Anatomía del arnés en disco** (dogfood): `arnes.l0.json` (fases+spine+META) ·
  `.claude-plugin/plugin.json` · `CLAUDE.md` · `skills/*/SKILL.md` con `contract:`
  fusionado en frontmatter (why/capabilities/clase/arquetipo/perfil_harness/caja/
  fase/estado/necesita/entrega/ruta/gate/handoff).
- **Invariantes que el chat editor debe honrar:** A1–A7 (VISION.md:62-87) · aditivo
  sin pérdida · spine declarado (jamás inventar estados) · una transición por caja ·
  escritor único por artefacto · firewall no-phantom · gate honesto (`none` permitido)
  · META completa en manifiesto. `RunGraph` los machaca a error/warn — el gate card
  del chat (decisión #4) es su superficie natural.
- **Permisos de una sesión "modificar arnés":** rol con escritura (p.ej. backend-dev)
  ⇒ Edit/Write "aprobables" (jamás auto: `escrituraDirecta` los saca de
  `--allowedTools`); toda escritura pasa por tarjeta de aprobación con diff.

## Frente ① — Proyectos GitHub chat-sobre-Claude-Code (2026-07-08, fuente leída)

### Panorama (consolidación H1-2026)

- Muertos/archivados: opcode/claudia (oct-25) · CUI (mar-26) · claude-code-webui
  (may-26) · Crystal→closed. Vivos: **siteboon/claudecodeui** (12.5k★, el más parecido
  a ArnesIA) · **happy** (22.5k★) · **vibe-kanban** (27.3k★, community) · agentapi.
- 3 patrones de transporte sobreviven: **stream-json+control protocol** (el más rico;
  vibe-kanban, claude-code-chat) · SDK oficial `query()` (siteboon, happy) · PTY-scrape
  (agentapi — frágil, NO seguir). **Nuestra decisión HS-04 (Go conductor stream-json
  bidireccional) = exactamente el patrón del segmento ganador.**

### Wire-format verificado (cierra el spike `control_response` de HS-04 en papel)

Fuentes: `query.py` del SDK Python oficial + claude-code-chat `extension.ts:1972-2093`
+ vibe-kanban `claude.rs`. Solo emite `can_use_tool` si spawneas con
**`--permission-prompt-tool stdio`** (literal no documentado pero real, usado por los 3).

- Envelope: `{"type":"control_request","request_id":"req_N_x","request":{"subtype":…}}`
  ↔ `{"type":"control_response","response":{"subtype":"success"|"error",
  "request_id":…,"response":{…}}}`.
- `can_use_tool` request: `{subtype, tool_name, input, tool_use_id,
  permission_suggestions:[PermissionUpdate], blocked_path?, decision_reason?}`.
- Allow: `{"behavior":"allow","updatedInput":{…REQUERIDO…},
  "updatedPermissions":[…suggestions elegidas…],"toolUseID":…}` · Deny:
  `{"behavior":"deny","message":"…REQUERIDO…","interrupt":true,"toolUseID":…}`.
- **`interrupt` in-band**: `{"request":{"subtype":"interrupt"}}` por stdin (adiós
  kill-only). También `set_permission_mode`, `set_model`, `get_context_usage`.
- **`initialize`** (app→CLI al arrancar): registra hooks PreToolUse programáticos
  (`hookCallbackIds` → llegan como `hook_callback` por el mismo canal — gates que
  aplican SIEMPRE, incluso bajo allow-rules) y devuelve `account.subscriptionType`,
  comandos soportados, output styles.
- `AskUserQuestion` y `ExitPlanMode` llegan por el MISMO canal `can_use_tool`
  (respuesta vía `updatedInput`).
- Flags extra útiles: `--replay-user-messages` (ack de stdin) ·
  `--resume-session-at <uuid>` (rewind truncado — poco conocido) · `--fork-session` ·
  `--include-hook-events`.

### Convergencia de los mejores (tabla corta)

| Capacidad | Best-in-class | Cómo |
|---|---|---|
| Transporte | vibe-kanban | proceso persistente stream-json bidireccional (JAMÁS spawn-por-turno: clase de bugs new-session-id) — **ya lo hacemos** |
| Permisos | vibe-kanban/claude-code-chat | `--permission-prompt-tool stdio` + `can_use_tool` in-band; MCP-broker = generación anterior |
| "Always allow" | claude-code-chat | echo de `permission_suggestions` como `updatedPermissions` (persiste en settings.local del arnés) — NOSOTROS: grants TTL en daemon (más estricto, correcto) |
| Interrupt | siteboon/happy | control_request interrupt; kill escalonado solo fallback |
| Sesiones | siteboon | ID-app propio ↔ session_id CC (CAMBIA en cada resume); frames con `seq`+resubscribe `lastSeq` — nuestro broker ya tiene Last-Event-ID |
| Descubrir session_id | happy | hook `SessionStart` inyectado — robusto a resume/compact |
| Render tools | siteboon | registro declarativo tool→config; diffs CodeMirror merge; subagents por `parent_tool_use_id` |
| Streaming fino | vibe-kanban | `--include-partial-messages` → text_delta — **ya lo hacemos** |
| Coste | claude-code-chat | `usage`+`total_cost_usd` del result; JAMÁS precios hardcodeados |
| Testing | CUI | binario `claude` MOCK que emite JSONL válido |

### Potholes con fuente (los que nos tocan)

1. `--resume` emite session_id NUEVO + JSONL nuevo duplicando historia — vigilar.
2. `CLAUDE_CODE_ENTRYPOINT` ≠ `cli` esconde sesiones del picker `claude --resume`
   (happy #1202) — elegir valor a propósito.
3. `ANTHROPIC_API_KEY` en env pisa OAuth de suscripción SILENCIOSO → cargos
   (siteboon #568) — el daemon controla el env del hijo.
4. 401 puede llegar como TEXTO del asistente, no error estructurado.
5. `bypassPermissions`/allow-rules ⇒ callback de permisos NO dispara — gates
   universales van en hooks PreToolUse (vía `initialize`).
6. `~/.claude` = solo-lectura para apps de terceros (escribir JSONL ajeno = invasivo,
   incompatible con `claude --resume`) — coincide con nuestra doctrina.
7. Strippear `NODE_OPTIONS`/`VSCODE_INSPECTOR_OPTIONS` del env del hijo.

## Frente ② — assistant-ui + AG-UI + CodeMirror merge (estado del arte, 2026-07-08)

### assistant-ui (`@assistant-ui/react` 0.14.26, 2026-07-04 — aún 0.x)

- **Runtime para nosotros = `useExternalStoreRuntime`** — doc oficial: elegirlo cuando
  "ya guardas mensajes en zustand… control total de estado/persistencia/sync". Adapter
  por capacidades: `messages`, `onNew`, `isRunning`, **`isSendDisabled`** (= estado
  await-permiso: input usable, send bloqueado), `onAddToolResult`/`onResumeToolCall`
  (HITL), `convertMessage`. Deltas = REEMPLAZAR el objeto mensaje en el store por delta
  (inmutable, jamás push); throttle ~50ms recomendado (gotcha #4051 React 19 + Vite).
- **APIs actuales (v0.14):** `useAui`/`useAuiState` (v0.12 renombró useAssistantApi);
  `MessagePrimitive.Parts` con children render-fn; **`makeAssistantToolUI` DEPRECATED**
  → `defineToolkit` + `Tools({toolkit})` con entries `type:"backend"` (render-only,
  el tool ejecuta daemon-side). Render props: `{args, argsText, status, result,
  addResult, resume, approval}`.
- **Aprobaciones nativas = mapping 1:1 a nuestro control_request+grants**: part
  tool-call con `approval:{id}` + status `requires-action`; renderer recibe
  `respondToApproval`; `options:[{kind:"allow-once"|"allow-always"|…, grants:["git *"]}]`.
  Doc: "persistence is host-owned" (= nuestros grants TTL en el daemon). ✔ decisión #3.
- **Composer:** slash/mentions primera clase — `ComposerPrimitive.Unstable_TriggerPopover
  char="/"` + `unstable_useSlashCommandAdapter`/`unstable_useMentionAdapter`. ✔ chip/@.
- Reasoning part `type:"reasoning"` + componentes copy-in (auto-abre en streaming);
  widgets backend-driven = parts `data-*` + `makeAssistantDataUI` (encaja gate card).
- Styling: SOLO ruta Tailwind+shadcn copy-in (registry `r.assistant-ui.com`) —
  `@assistant-ui/styles` deprecated. Alineado con nuestra FE arch (shadcn/Base UI).
- Gotchas: lockstep de versiones `@assistant-ui/*` + limpiar `.vite/deps` al subir ·
  bundle 154kB gzip · virtualización por turno (@tanstack/react-virtual, receta oficial,
  Viewport autoscroll OFF) recién a ~150-200 msgs · markdown memoizado por bloque
  (`react-streamdown` con `defer`) · Tailwind v4 no escanea node_modules → `@source`.

### AG-UI (`@ag-ui/*` 0.0.57)

- Taxonomía 34 eventos verificada en fuente: RUN_STARTED/FINISHED/ERROR · STEP_* ·
  TEXT_MESSAGE_START/CONTENT/END(/CHUNK) · TOOL_CALL_START/ARGS/END/RESULT ·
  REASONING_* (THINKING_* deprecated) · STATE_SNAPSHOT/**STATE_DELTA (JSON Patch
  RFC 6902)** · MESSAGES_SNAPSHOT · ACTIVITY_* · RAW · CUSTOM. Wire = `data: <json>\n\n`,
  `type` SCREAMING_SNAKE, campos camelCase.
- **Interrupts primera clase** (2026): `RUN_FINISHED{outcome:{type:"interrupt",
  interrupts:[{id, reason:"tool_call"|"confirmation"|…, responseSchema, expiresAt}]}}`
  + resume en el siguiente `RunAgentInput.resume[]`. Modelo limpio para control_request.
- **Puente oficial existe**: `@assistant-ui/react-ag-ui` 0.0.44 (`useAgUiRuntime`) —
  pero asume HttpAgent (POST→SSE por run). Nuestro SSE único multiplexado ⇒ ruta
  sancionada = **ExternalStoreRuntime + reducer AG-UI propio en Zustand** (o custom
  `AbstractAgent`). Confirma decisión #2.
- **SDK Go comunitario en el repo AG-UI**: `sdks/community/go` — constantes de eventos +
  `sse.SSEWriter` + validación de secuencia. Importable o copiable para el emisor del
  daemon (hoy emite dockFrames propios: migrar/mapear).

### CodeMirror merge (`@codemirror/merge` 6.12.2)

- Chat diff = **`unifiedMergeView`** read-only: `mergeControls:false` +
  `EditorView.editable.of(false)` + `EditorState.readOnly.of(true)` +
  `collapseUnchanged` + `diffConfig.timeout`. Accept/reject por chunk existe
  (`acceptChunk`/`rejectChunk`, `mergeControls` como función = botones custom) para la
  fase aprobación-con-diff editable.
- Para diffs chicos en transcript: `presentableDiff()` + markup propio = lo más barato
  (patrón claudecodeui: solo líneas cambiadas, sin syntax highlight en transcript).
- React: sin wrapper first-party — mount en useEffect + destroy; cada diff = instancia
  CM completa ⇒ lazy-mount off-screen.

### Patrones de producción (claudecodeui = mejor referencia leída en fuente)

- Diffs Edit/Write: colapsados por defecto, borde ámbar, solo-líneas-cambiadas,
  el input ES el display (`hideOnSuccess`). Bash: fila colapsada `$ cmd` + badge
  "N lines" + spinner; expandido `max-h-80` scroll interno; auto-abre UNA vez en error.
- TodoWrite → **MIGRACIÓN: desde CC v2.1.142 son TaskCreate/TaskUpdate/TaskGet/TaskList**
  (estado por task-ID que llega en el tool_result, no en el input) — parsear defensivo,
  render checklist + progreso.
- Subagent/Task: sección colapsada con "Currently: {tool}" vivo + "View tool history (N)"
  anidado — jamás transcript completo del hijo.
- Costo/uso: widget a nivel SESIÓN (pill tokens + gauge de contexto), no badge por
  mensaje. = nuestra ctx bar.
- Permisos: banner sobre el composer con la REGLA derivada visible (`Bash(git *)`) para
  que "recordar" sea informado; batching de pendientes de la misma regla;
  ExitPlanMode como tarjeta especial inline.
- Ejemplo oficial más cercano a nuestro daemon: `assistant-ui/examples/with-opencode`
  (assistant-ui sobre un servidor de coding-agent).

## Síntesis final (4 frentes)

1. **Validación externa de HS-04:** nuestro conductor (proceso persistente, stream-json
   bidireccional, `--include-partial-messages`) es EXACTAMENTE el patrón del segmento
   ganador (vibe-kanban/claude-code-chat y lo que el SDK oficial hace por dentro).
   No se cambia el transporte.
2. **El spike `control_response` queda cerrado en papel** (wire verificado en SDK
   Python + 2 repos): falta solo confirmar contra el binario real. Nuestro
   `controlResponseLine` debe emitir `updatedInput` (requerido en allow) y `message`
   (requerido en deny) — revisar en spec.
3. **Fix backend #1:** `spawnLocked` debe pasar `Permisos` (hoy solo T3 lo hace) para
   que el Dock reciba `can_use_tool`. **Fix #2:** mandar `initialize` al arrancar
   (subscriptionType + hooks PreToolUse programáticos futuros). **Fix #3:** interrupt
   in-band (hoy no hay Stop real).
4. **FE:** ExternalStoreRuntime (assistant-ui 0.14) + reducer AG-UI/dockFrames en
   Zustand + `defineToolkit` backend-render + approval nativo (mapea 1:1 a grants TTL)
   + unifiedMergeView read-only para diffs + slash/mentions primera clase del composer.
   El emisor Go puede adoptar taxonomía AG-UI (SDK Go comunitario copiable) o mapear
   dockFrames→AG-UI en FE; decisión de spec.
5. **Doctrina:** el kit ya inyecta las 2 meta-skills; el chat "modificar arnés" es
   sesión normal (cwd=arnés, rol con escritura) + chip de alcance (FuenteService) +
   gate conformance visible (endpoint existente). Los invariantes A1–A7/spine/escritor-
   único/firewall YA los machaca RunGraph — el chat solo los hace visibles.
6. Orden: mockup (este paquete) → firma → spec/design → implementación FE + 3 fixes Go
   + spike real de control_response contra el binario.
