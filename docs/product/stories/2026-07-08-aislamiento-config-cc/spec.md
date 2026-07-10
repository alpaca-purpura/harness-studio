# Spec — Aislamiento de superficie de config al spawnear CC

> **FIRMADA** 2026-07-08 — autorización del operador vía prompt de arranque del paquete
> (RF-160..RF-167 tal cual redactados). Implementación autorizada.

Sin mockup/`design.md` (paquete backend puro, cero UI). Cada RF trazado a **evidencia**
(URL oficial + cita, en vez de `mockup:línea`) y al **archivo/línea** de código que toca.
Numeración RF-160.. (el máximo usado en el repo hasta 2026-07-08 es RF-151, boton-
actualizar/inspector-drawer/franja-artefactos).

## Evidencia (investigación 2026-07-08, 3 agentes `claude-code-guide`, code.claude.com/docs)

| # | Pregunta | Fuente | Cita/hallazgo |
|---|---|---|---|
| E1 | ¿`--bare` rompe auth de suscripción? | code.claude.com/docs/en/headless (2026-07-08) | «Bare mode skips OAuth and keychain reads. Anthropic authentication must come from ANTHROPIC_API_KEY or an apiKeyHelper» — reconfirma HS-04 |
| E2 | ¿`--safe-mode` rompe auth? | mismo | Preserva auth normal; solo apaga customizaciones (CLAUDE.md/skills/plugins/hooks/MCP/comandos/output-styles/temas/keybindings) |
| E3 | ¿`--safe-mode` + flags explícitos conviven? | inferencia sin cita dura | **No confirmado** — requiere experimento (Fase 3) |
| E4 | ¿Qué carga cada `--setting-sources`? | code.claude.com/docs/en/settings.md | `user`=`~/.claude/settings.json` (incl. `enabledPlugins`) · `project`=`.claude/settings.json` del cwd · `local`=`.claude/settings.local.json` |
| E5 | ¿CLAUDE.md del propio cwd sobrevive si excluyo `user`? | code.claude.com/docs/en/claude-md.md | Sí — CLAUDE.md discovery es mecanismo propio, independiente de `--setting-sources` (solo `--bare` lo apaga entero) |
| E6 | ¿`--strict-mcp-config` tapa conectores claude.ai? | code.claude.com/docs/en/mcp.md, jerarquía «Scope hierarchy and precedence» | Sí — jerarquía local>project>user>plugin>connector; `--mcp-config`+`--strict-mcp-config` = solo lo pasado ahí |
| E7 | ¿Cómo declara un plugin su propio MCP? | mismo, «Plugin-provided MCP servers» | `.mcp.json` en raíz del plugin, campo `mcpServers`, `${CLAUDE_PLUGIN_ROOT}` |
| E8 | ¿`--strict-mcp-config` tapa MCP de plugin? | inferencia de la jerarquía, sin cita literal | **No confirmado** — tratar como bloqueado hasta probar |

## RF

### RF-160 · Baseline real medido antes de cualquier cambio

Spawnear una sesión contra el dogfood (`dev-full-cycle`) con el código ACTUAL (sin
modificar), turno-sonda: *"Lista exhaustivamente qué skills, subagentes, comandos y
servidores MCP tenés disponibles ahora mismo, agrupados por origen si podés inferirlo."*
Capturar la respuesta completa como `baseline.md` en este paquete. Sirve de contraste para
RF-161/162/163.

**Gherkin-equivalente:**
```
Dado el daemon arnesia corriendo con el dogfood dev-full-cycle registrado
Cuando se spawnea una sesión SIN cambios de código y se envía el turno-sonda
Entonces la respuesta lista skills/MCP AJENOS al kit propio de arnesia (ej. golang-*, MCP de cuenta)
Y esa respuesta queda archivada como evidencia baseline
```

### RF-161 · Aislamiento MCP siempre activo

`ports.Injection` (`internal/ports/kit.go:10-17`) suma campo `MCPConfigFile string`.
`Provisioner.materialize` (`internal/adapters/provision/provisioner.go`) escribe
`~/.arnesia/mcp.json` con `{"mcpServers":{}}` si no existe ya un archivo del kit; `Provision`
lo agrega al `Injection` devuelto. `conductor.go:SpawnArgs`
(`internal/adapters/agent/claudecode/conductor.go:82-90`) agrega, siempre que
`opts.Injection.MCPConfigFile != ""`: `--mcp-config <archivo> --strict-mcp-config`.

**Gherkin:**
```
Dado un spawn con Injection.MCPConfigFile poblado
Cuando SpawnArgs arma el argv
Entonces incluye "--mcp-config" "<archivo>" "--strict-mcp-config" en ese orden
Y una sonda contra el dogfood NO lista ningún MCP fuera de los declarados en ese archivo
```

**Verificación:** repetir el turno-sonda de RF-160 tras el cambio → 0 MCP ajenos ·
`arnesia conformance --arnes` sobre dogfood sin regresión (mismo pass/fail que hoy) ·
fitness test nuevo `TestSpawnArgsMCPAislado` (mismo archivo que
`TestPermissionSetParametrizedByRole`, mismo patrón: asserts sobre el argv de `SpawnArgs`).

### RF-162 · Exclusión de settings de usuario

`ports.SpawnOpts` (`internal/ports/agent.go:67-`) o `SpawnArgs` directamente agrega,
incondicional, `--setting-sources project,local`. Sin campo nuevo en `Injection` — es
comportamiento fijo del conductor, no configurable por caller (nadie debería poder
reintroducir `user`).

**Gherkin:**
```
Dado cualquier spawn de SpawnArgs
Cuando arma el argv
Entonces incluye "--setting-sources" "project,local"
Y una sonda contra el dogfood NO lista skills del enabledPlugins personal del operador (golang-*, stitch, etc. — los mismos vistos en el disparador de HS-17)
Y el CLAUDE.md propio del árbol del arnés SIGUE presente en la sonda (no se pierde contexto propio)
```

**Verificación:** turno-sonda antes/después, diff explícito de qué desaparece (debe ser
SOLO lo ajeno) · confirmar que el CLAUDE.md del arnés (ej.
`dogfood/dev-full-cycle/CLAUDE.md`) sigue citado/usado por el modelo · auditoría de
ancestros (`~/.arnesia`, `$HOME`) sin CLAUDE.md perdido (documentar hallazgo aunque sea
"ninguno encontrado") · `TestSpawnArgsSettingSourcesExcludeUser`.

### RF-163 · Experimento `--safe-mode` (resultado abierto, no bloqueante)

Script de prueba (no en `SpawnArgs` todavía) que spawnea manualmente:
`claude --safe-mode --plugin-dir <kit> --append-system-prompt-file <doctrina>
--add-dir <knowhow> -p "<sonda>" --output-format stream-json` contra el dogfood.

**Gherkin:**
```
Dado el spawn de prueba con --safe-mode + los 3 flags de inyección explícitos
Cuando se envía el turno-sonda
Entonces SI el kit propio de arnesia sigue respondiendo igual que el baseline-con-D2+D3 → adoptar --safe-mode como capa extra en SpawnArgs (RF-164)
Entonces SI el kit propio deja de responder (skills/subagentes propios desaparecen) → descartar --safe-mode, documentar en decisiones.md por qué, cerrar RF-163 sin RF-164
```

**Verificación:** transcripción completa del experimento archivada en este paquete
(`experimento-safe-mode.md`) independientemente del resultado — la evidencia negativa
también es evidencia.

### RF-164 · (condicional a RF-163) Sumar `--safe-mode` a `SpawnArgs`

Solo si RF-163 confirma compatibilidad. Mismo patrón que RF-161/162: flag incondicional en
`SpawnArgs`, fitness test nuevo, sonda de verificación.

### RF-165 · Cementado as-code

`docs/architecture/boundaries/superficie-local-confinada.md`: nueva versión (bump de version en
frontmatter), suma 2 (o 3 si RF-164 aplica) checks nuevos a la tabla «Checklist
evaluable» con `enforced_by:` apuntando a los fitness tests de RF-161/162(/164). Sección
L2 suma un párrafo describiendo el eje config-source, con cita a `mcp.md` L1.6
(`mcp-toolsearch-off`/`mcp-unused`) como el hallazgo de costo que esto cierra. Changelog
nuevo en el nodo. `docs/architecture/knowledge/elements/mcp.md`: nota en cambios recientes o changelog —
"el spawn de arnés ahora enforcea `mcp-toolsearch-off`/`mcp-unused` a nivel producto, no
solo advisory" (sin sumar checks nuevos ahí, el enforcement vive en `arch/`). CLAUDE.md
maestro: una línea en "Decisiones técnicas vigentes" documentando los flags nuevos de
`SpawnArgs`.

**Verificación:** `arnesia conformance --todo` corre sin romper (conteo total sube por los
checks nuevos de `arch/`, documentar el número exacto medido, no de memoria — norma del
repo tras HS-16).

### RF-166 · Gate de calidad — sin pérdida de entrega propia de arnesia

Round-trip completo sobre el dogfood real: spawn de sesión Dock (chat) + una corrida T3
de caja (`BoxConductor`) con el código de RF-161/162(/164) aplicado.

**Gherkin:**
```
Dado el dogfood dev-full-cycle con los cambios de RF-161/162(/164) en producción
Cuando se corre una sesión de chat Y una caja T3
Entonces el kit propio (skills/subagentes/comandos del plugin ②) responde IDÉNTICO al baseline pre-cambio
Y el overlay de doctrina (--append-system-prompt-file, ①) sigue presente
Y el knowhow (--add-dir, ①) sigue legible
Y los permisos derivados del rol (eje aparte, permisos-derivan-del-rol) siguen intactos
Y arnesia conformance --arnes sobre el dogfood da el MISMO pass/fail que antes del paquete (sin regresión, sin drift nuevo)
Y la medición de contexto de la sonda (RF-160 vs post-cambio) muestra la baja cuantificada
```

**Verificación:** todo `go test ./... -race` + `golangci-lint` + `pnpm run verify` verdes
· `arnesia conformance --arnes` sin regresión · transcripciones de las 3 sondas
(baseline/post-MCP/post-settings) archivadas en el paquete como evidencia.

### RF-167 · Cierre de ficha HS-17

`LEDGER.md`: entrada de cierre bajo HS-17 (mismo patrón HS-14: diagnóstico → ficha de
cierre con verificación real), resumiendo qué se ejecutó, qué se descartó (`--bare`
siempre, `--safe-mode` según RF-163) y las cifras medidas (conteo de checks `arch/`,
resultado de las sondas). `PARIDAD.md` de este paquete: matriz RF↔commit↔verificación (en
vez de RF↔mockup↔componente, adaptado a paquete sin UI).

## Orden de implementación (mini-plan de commits, sellado)

1. RF-160 — sonda baseline (sin código, solo evidencia archivada).
2. RF-161 — `feat(HS-17): aislamiento MCP siempre activo en SpawnArgs (--strict-mcp-config)`.
3. RF-162 — `feat(HS-17): --setting-sources project,local siempre en SpawnArgs`.
4. RF-163 — experimento `--safe-mode` (script suelto, no toca `SpawnArgs`; documenta resultado).
5. RF-164 (condicional) — `feat(HS-17): --safe-mode en SpawnArgs si RF-163 confirma`.
6. RF-165 — `docs(HS-17): cementa eje config-source en superficie-local-confinada + mcp.md`.
7. RF-166 — gate de calidad, sin commit propio (verificación sobre lo ya commiteado).
8. RF-167 — `docs(HS-17): cierre — ficha de cierre LEDGER + PARIDAD.md`.
