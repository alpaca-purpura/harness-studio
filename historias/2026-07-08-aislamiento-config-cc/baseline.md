# Sonda: RF-160 baseline (SpawnArgs actual, sin cambios)

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Respuesta del modelo

Inventario actual, agrupado por origen inferido — solo de contexto, sin tools.

## Skills

**Kit propio arnés (arnesia-kit)** — 2
`arnesia-kit:auditar-arnes`, `arnesia-kit:forjar-caja`

**Plugin caveman (marketplace)** — 7
`caveman:cavecrew`, `caveman:caveman`, `caveman:caveman-commit`, `caveman:caveman-compress`, `caveman:caveman-help`, `caveman:caveman-review`, `caveman:caveman-stats`

**Plugin golang (samber/cc-skills-golang)** — 42
benchmark, cli, code-style, concurrency, context, continuous-integration, data-structures, database, dependency-injection, dependency-management, design-patterns, documentation, error-handling, google-wire, graphql, grpc, how-to, lint, modernize, naming, observability, performance, popular-libraries, project-layout, safety, samber-do, samber-hot, samber-lo, samber-mo, samber-oops, samber-ro, samber-slog, security, spf13-cobra, spf13-viper, stay-updated, stretchr-testify, structs-interfaces, swagger, testing, troubleshooting, uber-dig, uber-fx (prefijo `golang-` en todas)

**Bundle Stitch/diseño** — 14
`design-md`, `enhance-prompt`, `react-components`, `remotion`, `shadcn-ui`, `taste-design`, `stitch-code-to-design`, `stitch-extract-design-md`, `stitch-extract-static-html`, `stitch-generate-design`, `stitch-loop`, `stitch-manage-design-system`, `stitch-react-native`, `stitch-upload-to-stitch`

**Core / bundled operador** — 16
`find-skills`, `deep-research`, `dataviz`, `update-config`, `keybindings-help`, `verify`, `code-review`, `simplify`, `fewer-permission-prompts`, `loop`, `schedule`, `claude-api`, `run`, `init`, `review`, `security-review`

## Subagentes (Agent tool)

**Built-in genéricos**: `claude`, `Explore`, `general-purpose`, `Plan`, `statusline-setup`
**Plugin caveman**: `caveman:cavecrew-builder`, `caveman:cavecrew-investigator`, `caveman:cavecrew-reviewer`

## Comandos

Comandos slash = las skills arriba (invocables `/nombre`). Aparte, `/caveman` activo ahora (level: full) — meta-comando del plugin caveman, no listado como skill separada.

## MCP servers

`chrome-devtools`, `claude_design` — ambos **aún conectando**, origen cuenta/operador (no arnés, no kit). Tools no disponibles todavía.

## stderr

```
(vacío)
```