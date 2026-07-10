# PARIDAD — Aislamiento de superficie de config al spawnear CC (HS-17)

Paquete backend puro, sin UI: matriz RF↔commit↔verificación (en vez de
RF↔mockup↔componente↔RF, que no aplica sin mockup).

| RF | qué | commit | verificación real |
|---|---|---|---|
| RF-160 | sonda baseline, sin código | `973cfdc` | `baseline.md`: spawn pre-cambio contra el dogfood real filtra 6 MCP de cuenta + ~50 skills ajenas |
| RF-161 | `--mcp-config`+`--strict-mcp-config` siempre | `973cfdc` | `TestSpawnArgsMCPAislado` + `TestProvisionMaterializesAndIsIdempotent` (extendido) verdes · `post-mcp.md`: 6→0 MCP de cuenta, kit propio intacto · `conformance --arnes` sin regresión |
| RF-162 | `--setting-sources project,local` incondicional | `55cc8db` | `TestSpawnArgsSettingSourcesExcludeUser` verde · `post-settings.md`: ~50 skills ajenas → 0, kit propio + CLAUDE.md del arnés intactos (cita textual exacta) · hallazgo residual CLAUDE.md ascendente documentado, no bloqueante · `conformance --arnes` sin regresión |
| RF-163 | experimento `--safe-mode` | `21a33f5` | `experimento-safe-mode.md`: overlay ① sobrevive, plugin ② (kit propio) NO — evidencia real decisiva, criterio de descarte cumplido |
| RF-164 | (condicional) `--safe-mode` en `SpawnArgs` | — | NO SE IMPLEMENTA — RF-163 dio resultado de descarte, documentado en `decisiones.md` D4 |
| RF-165 | cementado as-code | `648a863` | `superficie-local-confinada.md` v1.2→v1.3 (7→9 checks) · `mcp.md` v1.1 · `CLAUDE.md` línea nueva · `conformance superficie-local-confinada` 9/9 (1 defer preexistente) |
| RF-166 | gate de calidad | `9e0e52c` | `gate-calidad.md`: suites verdes, `conformance --arnes` sin regresión (20/1), round-trip kit+doctrina+knowhow+permisos intactos, medición de contexto cuantificada (`input_tokens` −75.4%, total −38.4%) |
| RF-167 | cierre de ficha | (este commit) | `LEDGER.md` entrada de cierre HS-17 + este `PARIDAD.md` |

## Decisiones (estado final)

| id | qué | estado |
|---|---|---|
| D1 | descartar `--bare` | `FIRMADA` (heredada HS-04) |
| D2 | `--mcp-config`+`--strict-mcp-config` siempre | `FIRMADA` — implementada RF-161 |
| D3 | `--setting-sources project,local` | `FIRMADA` — implementada RF-162 |
| D4 | `--safe-mode` condicional | `FIRMADA-DESCARTADA` — RF-163 dio resultado de descarte |
| D5 | cementar en `superficie-local-confinada`, no boundary nueva | `FIRMADA` — implementada RF-165 |
| D6 | verificación siempre vía sonda real | `FIRMADA` — aplicada en las 6 fases |

## Archivos de evidencia (este paquete)

`baseline.md` · `post-mcp.md` · `post-settings.md` · `experimento-safe-mode.md` ·
`medicion-contexto-baseline.md` · `medicion-contexto-post.md` · `gate-calidad.md` ·
`sonda/sonda.mjs` (protocolo stream-json reutilizable) · `sonda/sonda-claudemd.mjs` ·
`sonda/provision.go` (desechable, `//go:build ignore`).

## Código tocado

`internal/ports/kit.go` (campo `MCPConfigFile`) · `internal/adapters/provision/
provisioner.go` (+test) · `internal/adapters/agent/claudecode/conductor.go` (+test) ·
`docs/architecture/boundaries/superficie-local-confinada.md` · `docs/architecture/fitness/
hs17_config_source_test.go` (nuevo) · `docs/architecture/knowledge/elements/mcp.md` · `CLAUDE.md`.

## Nota de higiene de repo

Ejecutado con otra sesión trabajando en paralelo el mismo working tree (deuda HS-16 sin
commitear al arrancar). Cada commit de este paquete usó `git add <archivos específicos>`,
nunca `-A`/`.`. Única excepción de bajo riesgo: un rename puro pre-stageado
(`docs/architecture/conventions/hooks.md→git-hooks.md`, 0 cambio de contenido) viajó en el commit
`973cfdc` porque `git commit` sube todo el índice — sin pérdida ni mezcla de contenido de
código ajeno con el propio.
