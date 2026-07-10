# Sonda: RF-161 post-MCP (--mcp-config + --strict-mcp-config)

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow --mcp-config /home/chalreme/.arnesia/mcp.json --strict-mcp-config
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Respuesta del modelo

## Skills (vía Skill tool)

**Kit propio arnés (`arnesia-kit:`)**
- `arnesia-kit:auditar-arnes`
- `arnesia-kit:forjar-caja`

**Plugin caveman (`caveman:`)**
- `caveman:cavecrew`
- `caveman:caveman`
- `caveman:caveman-commit`
- `caveman:caveman-compress`
- `caveman:caveman-help`
- `caveman:caveman-review`
- `caveman:caveman-stats`

**Plugin samber/cc-skills-golang** (orquestador `golang-how-to` + ~40 específicos)
`golang-how-to`, `golang-benchmark`, `golang-cli`, `golang-code-style`, `golang-concurrency`, `golang-context`, `golang-continuous-integration`, `golang-data-structures`, `golang-database`, `golang-dependency-injection`, `golang-dependency-management`, `golang-design-patterns`, `golang-documentation`, `golang-error-handling`, `golang-google-wire`, `golang-graphql`, `golang-grpc`, `golang-lint`, `golang-modernize`, `golang-naming`, `golang-observability`, `golang-performance`, `golang-popular-libraries`, `golang-project-layout`, `golang-safety`, `golang-samber-do`, `golang-samber-hot`, `golang-samber-lo`, `golang-samber-mo`, `golang-samber-oops`, `golang-samber-ro`, `golang-samber-slog`, `golang-security`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-stay-updated`, `golang-stretchr-testify`, `golang-structs-interfaces`, `golang-swagger`, `golang-testing`, `golang-troubleshooting`, `golang-uber-dig`, `golang-uber-fx`

**Plugin Stitch/design**
`design-md`, `enhance-prompt`, `react-components`, `remotion`, `shadcn-ui`, `stitch-code-to-design`, `stitch-extract-design-md`, `stitch-extract-static-html`, `stitch-generate-design`, `stitch-loop`, `stitch-manage-design-system`, `stitch-react-native`, `stitch-upload-to-stitch`, `taste-design`

**Core/standalone (sin namespace, origen builtin o marketplace suelto)**
`deep-research`, `dataviz`, `update-config`, `keybindings-help`, `verify`, `code-review`, `simplify`, `fewer-permission-prompts`, `loop`, `schedule`, `claude-api`, `run`, `init`, `review`, `security-review`

Nota: nombre = comando slash tamién (ej `/code-review`, `/simplify`, `/loop`, `/schedule`).

## Subagentes (Agent tool, `subagent_type`)

**Plugin caveman**
- `caveman:cavecrew-builder` — edit quirúrgico 1-2 archivos
- `caveman:cavecrew-investigator` — locator read-only
- `caveman:cavecrew-reviewer` — review diff/PR

**Core/builtin**
- `claude` — catch-all default FleetView
- `Explore` — búsqueda read-only rápida
- `general-purpose` — research/multi-step
- `Plan` — arquitecto de plan implementación
- `statusline-setup` — config statusline

## MCP servers

Ninguno visible en contexto actual. No hay tools `mcp__*` en mi lista. No puedo confirmar sin chequear config real.

## stderr

```
(vacío)
```