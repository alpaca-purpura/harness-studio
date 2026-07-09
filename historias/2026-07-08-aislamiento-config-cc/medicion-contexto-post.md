# Sonda: RF-166 medición post (RF-161+RF-162 aplicados)

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --setting-sources project,local --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow --mcp-config /home/chalreme/.arnesia/mcp.json --strict-mcp-config
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Uso de contexto (result.usage)

input_tokens: 3193
cache_creation_input_tokens: 16186
cache_read_input_tokens: 15206
total (input+cache_creation+cache_read): 34585

## Respuesta del modelo

Desde mi contexto actual (system prompt + listados inyectados), sin ejecutar nada:

## Subagentes (agent types para la tool `Agent`)
Origen: nativos del harness (SDK/Claude Code). Ninguno inferible como del kit propio.
- `claude` — catch-all (tools: `*`)
- `Explore` — búsqueda read-only fan-out
- `general-purpose` — investigación/búsqueda multi-paso (tools: `*`)
- `Plan` — arquitecto de planes de implementación
- `statusline-setup` — configura la statusline

## Skills (tool `Skill`)

**Kit propio del arnés** (prefijo `arnesia-kit:` → la maquinaria ② inyectada por flags, tu doctrina):
- `arnesia-kit:auditar-arnes` — auditar el arnés del cwd contra la doctrina
- `arnesia-kit:forjar-caja` — crear una caja de proceso nueva

**Built-in de Claude Code / harness** (Anthropic-bundled; no llevan prefijo de plugin):
- `deep-research` — harness de investigación multi-fuente
- `dataviz` — diseño de gráficos/visualizaciones
- `claude-api` — referencia de la API de Claude / SDK
- `update-config` — configurar `settings.json` (hooks, permisos, env)
- `keybindings-help` — customizar `~/.claude/keybindings.json`
- `verify` — verificar un cambio end-to-end
- `code-review` — review del diff (incl. `ultra`, `--comment`, `--fix`)
- `simplify` — limpieza de reuse/simplificación
- `fewer-permission-prompts` — allowlist de Bash/MCP read-only
- `loop` — correr un prompt/comando en intervalo recurrente
- `schedule` — agentes cloud programados (cron/routines)
- `run` — lanzar y manejar la app del proyecto
- `init` — crear `CLAUDE.md`
- `review` — review de un PR de GitHub
- `security-review` — review de seguridad del branch

*(Los últimos —`init`, `review`, `security-review`, `code-review`— son los clásicos slash-commands built-in; no veo comandos custom del operador en `.claude/commands/` expuestos en mi contexto.)*

## Comandos
No tengo una lista separada de "comandos" más allá de las skills invocables como `/<nombre>`. Todo lo de arriba en la sección Skills es lo que puedo invocar por slash. No hay comandos custom adicionales visibles en mi contexto.

## Servidores MCP
**Ninguno visible ahora mismo.** No hay ninguna tool con forma de MCP (namespace de servidor) en mi contexto ni en la lista de deferred tools. Esto es coherente con la doctrina HS-17: aislamiento de superficie de config (`--strict-mcp-config`), donde ningún arnés hereda MCP de cuenta — aunque nota que ESTA es tu sesión de fábrica (①), no un arnés spawneado, así que la ausencia refleja que simplemente no hay MCP conectado a esta sesión.

## Tools diferidas (harness, NO son MCP — schema se carga con `ToolSearch`)
Nativas del harness, no de plugin ni de cuenta:
`CronCreate` · `CronDelete` · `CronList` · `DesignSync` · `EnterWorktree` · `ExitWorktree` · `Monitor` · `NotebookEdit` · `PushNotification` · `RemoteTrigger` · `SendMessage` · `TaskCreate` · `TaskGet` · `TaskList` · `TaskOutput` · `TaskStop` · `TaskUpdate` · `WebFetch` · `WebSearch`

---

**Honestidad sobre la inferencia de origen (gris ≠ verde):** la separación "kit propio vs built-in" es sólida solo para las 2 skills con prefijo `arnesia-kit:`. Para el resto, "built-in Anthropic vs plugin del operador" es *inferencia* por convención de nombres (sin prefijo de plugin); confirmarlo requeriría leer `settings.json`/`~/.claude`, que pediste no tocar. Lo marco como no-verificado.

## stderr

```
(vacío)
```