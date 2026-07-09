# Sonda: RF-166 medición baseline (sin --mcp-config/--setting-sources)

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Uso de contexto (result.usage)

input_tokens: 12959
cache_creation_input_tokens: 19832
cache_read_input_tokens: 23323
total (input+cache_creation+cache_read): 56114

## Respuesta del modelo

**Skills — kit propio arnés (`arnesia-kit`, marketplace prenter):**
- arnesia-kit:auditar-arnes
- arnesia-kit:forjar-caja

**Skills + subagentes — plugin `caveman` (operador):**
Skills: caveman:cavecrew, caveman:caveman, caveman:caveman-commit, caveman:caveman-compress, caveman:caveman-help, caveman:caveman-review, caveman:caveman-stats
Subagentes: caveman:cavecrew-builder, caveman:cavecrew-investigator, caveman:cavecrew-reviewer

**Skills — marketplace `samber/cc-skills-golang` (~50, prefijo `golang-*`):**
Cluster completo Go: benchmark, cli, code-style, concurrency, context, continuous-integration, data-structures, database, dependency-injection, dependency-management, design-patterns, documentation, error-handling, google-wire, graphql, grpc, how-to (orquestador), lint, modernize, naming, observability, performance, popular-libraries, project-layout, safety, samber-do, samber-hot, samber-lo, samber-mo, samber-oops, samber-ro, samber-slog, security, spf13-cobra, spf13-viper, stay-updated, stretchr-testify, structs-interfaces, swagger, testing, troubleshooting, uber-dig, uber-fx.

**Skills — bundle Stitch/diseño (sin prefijo visible, plugin diseño):**
design-md, enhance-prompt, react-components, remotion, shadcn-ui, stitch-code-to-design, stitch-extract-design-md, stitch-extract-static-html, stitch-generate-design, stitch-loop, stitch-manage-design-system, stitch-react-native, stitch-upload-to-stitch, taste-design

**Skills — nativas Claude Code (sin namespace):**
find-skills, deep-research, update-config, keybindings-help, verify, code-review, simplify, fewer-permission-prompts, loop, schedule, claude-api, run, init, review, security-review, dataviz

**Subagentes nativos (Agent tool):**
claude (catch-all), Explore (búsqueda read-only), general-purpose, Plan (arquitecto), statusline-setup

**MCP servers:** ninguno operativo ahora — 6 listados TODOS en estado "still connecting" (no invocables aún): `chrome-devtools`, `claude.ai Canva`, `claude.ai Gmail`, `claude.ai Google Calendar`, `claude.ai Google Drive`, `claude_design`. Todos cuenta/operador, ninguno del arnés.

**Extra — tools diferidas (no skill/MCP, requieren ToolSearch para cargar schema):** CronCreate/Delete/List, DesignSync, EnterWorktree/ExitWorktree, Monitor, NotebookEdit, PushNotification, RemoteTrigger, SendMessage, TaskCreate/Get/List/Output/Stop/Update, WebFetch, WebSearch.

## stderr

```
(vacío)
```