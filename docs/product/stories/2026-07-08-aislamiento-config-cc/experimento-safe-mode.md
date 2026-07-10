# Sonda: RF-163 --safe-mode + inyección explícita

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --safe-mode --setting-sources project,local --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow --mcp-config /home/chalreme/.arnesia/mcp.json --strict-mcp-config
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Respuesta del modelo

Acá va el inventario, armado **solo desde mi contexto/system prompt actual** (sin ejecutar nada ni leer archivos). Marco explícitamente lo que puedo inferir vs. lo que no — gris ≠ verde.

## Subagentes (Agent tool)
Origen: definiciones de agente del harness (registro `.claude/agents/*` o SDK). No puedo distinguir cuáles son built-in vs. propios del operador desde el contexto.

- **claude** — catch-all, tools `*`
- **Explore** — búsqueda read-only, fan-out (sin Edit/Write/Agent)
- **general-purpose** — investigación multi-paso, tools `*`
- **Plan** — arquitecto de planes de implementación (read-only)

## Skills (Skill tool)
Origen: no vienen etiquetadas por plugin/marketplace en mi contexto, así que la atribución de origen es **inferencia, no dato firmado**. Todas tienen pinta de skills estándar del harness Claude Code (ninguna es una caja del arnés `dev-full-cycle` — esas viven en `skills/<id>/SKILL.md` del repo y NO están cargadas como invocables aquí).

- deep-research
- dataviz
- update-config
- keybindings-help
- verify
- code-review
- simplify
- fewer-permission-prompts
- loop
- schedule
- claude-api
- run
- init
- review
- security-review

## Comandos (slash)
No tengo una lista de comandos separada de las skills — en este harness los `/<skill-name>` **son** las skills de arriba. Los únicos comandos que mi guía de sesión menciona explícitamente y que NO son invocables por mí:

- **/code-review ultra** (y su alias deprecado **/ultrareview**) — review multi-agente en la nube, disparado por el usuario y facturado; **no puedo lanzarlo yo**.

## Servidores MCP
**Ninguno visible en mi contexto.** No aparece ninguna tool con prefijo `mcp__…` ni ningún servidor MCP declarado. (El system prompt sí dice que agentes de workflow *pueden* alcanzar MCP de la sesión vía ToolSearch, pero desde mi contexto actual no hay evidencia de ningún servidor MCP conectado — lo dejo como **no-reconocido/ausente**, no como "cero confirmado".)

## Tools del harness diferidas (no son MCP)
Nombres conocidos vía ToolSearch, esquema aún no cargado. Son herramientas **nativas del harness**, no MCP:

- Tareas/agentes: TaskCreate, TaskGet, TaskList, TaskOutput, TaskStop, TaskUpdate, SendMessage, Monitor
- Cron/scheduling: CronCreate, CronDelete, CronList
- Worktree: EnterWorktree, ExitWorktree
- Web: WebFetch, WebSearch
- Otros: DesignSync, NotebookEdit, PushNotification, RemoteTrigger

## Tools activas directamente (top del prompt)
Agent, Bash, Edit, Read, Write, ReportFindings, ScheduleWakeup, Skill, ToolSearch, Workflow.

---

**Límites de honestidad de esta respuesta:**
- No puedo separar de forma firme "kit propio del arnés" vs. "plugins/marketplaces del operador" vs. "MCP de cuenta": mi system prompt **no etiqueta el origen** de skills ni subagentes. Todo lo de arriba clasificado por origen es inferencia.
- Las **cajas del arnés** `dev-full-cycle` (sus `skills/<id>/SKILL.md`) **no** figuran como skills invocables en este contexto — para enumerarlas de verdad habría que leer el repo, y me pediste no leer archivos.

## stderr

```
(vacío)
```

## Hallazgo #1 — el kit propio (②, plugin `arnesia-kit`) DESAPARECE

Comparado contra `post-mcp.md`/`post-settings.md` (mismos flags SIN `--safe-mode`, donde
la sección Skills lista explícitamente `arnesia-kit:auditar-arnes` y
`arnesia-kit:forjar-caja`): en esta corrida CON `--safe-mode` la lista de Skills NO
incluye ninguna entrada `arnesia-kit:*` — el modelo ni siquiera las menciona como
"invocables". `--safe-mode` mata el plugin cargado por `--plugin-dir` pese a ser
explícito.

## Sonda dirigida #2 — overlay ① (`--append-system-prompt-file`+`--add-dir`)

Pregunta puntual: ¿tenés cargado un texto de sistema que arranque con "Doctrina
ArnesIA"? ¿tenés `--add-dir` con knowhow? Respuesta:

> **1) ¿Texto de sistema "Doctrina ArnesIA"?** → Sí. `# Doctrina ArnesIA — overlay del
> conductor (cuerpo ①, siempre presente)`
> **2) ¿`--add-dir` con knowhow?** → Sí. `/home/chalreme/.arnesia/knowhow` — listó
> `rules.md, hooks.md, mcp.md, skills.md, subagents.md, settings-permissions.md,
> harness-profile.md, commands.md, headless-sdk.md, output-styles.md, plugins.md,
> statusline.md`.

Ambas citas coinciden exacto con el contenido real de `~/.arnesia/doctrine.md` y
`~/.arnesia/knowhow/`. **`--append-system-prompt-file` y `--add-dir` SOBREVIVEN a
`--safe-mode`** — solo `--plugin-dir` (el mecanismo de skills/plugin) muere.

## Veredicto (criterio D4)

`--safe-mode` mata SELECTIVAMENTE la mitad del kit propio: el overlay de doctrina ①
sobrevive, pero las skills invocables `arnesia-kit:auditar-arnes`/`forjar-caja` (②) NO
— exactamente el escenario que D4 marca como "skills/subagentes propios desaparecen →
descartar `--safe-mode`". **Se descarta.** D2 (`--mcp-config`+`--strict-mcp-config`) +
D3 (`--setting-sources project,local`) ya cierran los dos ejes de mayor costo/riesgo
real (MCP de cuenta + `enabledPlugins`/hooks personales) sin este costo — RF-164 no se
implementa.