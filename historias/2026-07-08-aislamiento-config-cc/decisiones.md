# Decisiones — Aislamiento de superficie de config al spawnear CC

Una entrada por decisión. Estado `PROPUESTA` (conversada, sin firma del operador) o
`FIRMADA` (operador confirmó). Ninguna línea de código se toca hasta que las decisiones
bloqueantes (D2, D3, D5) estén FIRMADAS vía `spec.md`.

## D1 · Descartar `--bare` como mecanismo de aislamiento

**Qué:** no usar `--bare` para limpiar la superficie de config del spawn, bajo ninguna
combinación.

**Por qué:** `--bare` fuerza autenticación por `ANTHROPIC_API_KEY`/`apiKeyHelper` — bloquea
OAuth/keychain. ArnesIA depende de la suscripción del OPERADOR (Claude Code login normal),
no de una API key propia (constraint fundacional, HS-04: «Sin --bare: la suscripción del
usuario queda intacta», ya en `conductor.go:81` como comentario). Reconfirmado 2026-07-08
con cita nueva: code.claude.com/docs/en/headless — «Bare mode skips OAuth and keychain
reads. Anthropic authentication must come from ANTHROPIC_API_KEY or an apiKeyHelper».

**Estado:** `FIRMADA` (heredada de HS-04, no requiere re-firma — es hecho técnico, no
preferencia).

## D2 · Aislar MCP con `--mcp-config`+`--strict-mcp-config` SIEMPRE

**Qué:** todo spawn de `SpawnArgs` agrega incondicionalmente `--mcp-config <archivo>
--strict-mcp-config`, apuntando a un archivo materializado por el `Provisioner`
(`~/.arnesia/mcp.json`, inicialmente `{"mcpServers":{}}`).

**Por qué:** es el eje mejor documentado (code.claude.com/docs/en/mcp, «Only use MCP
servers from --mcp-config, ignoring all other MCP configurations») y el que causó el
disparador original (62.2k tok de Canva/Gmail/Drive/Calendar/claude_design). Cero riesgo de
auth — MCP y autenticación son ejes ortogonales. Cierra también los conectores
account-level (claude.ai connectors, scope MÁS BAJO que `--mcp-config` en la jerarquía
local>project>user>plugin>connector).

**Riesgo abierto:** no está confirmado si `--strict-mcp-config` también tapa MCP
declarados DENTRO de un plugin cargado por `--plugin-dir` en el mismo comando (jerarquía
sugiere que sí, sin cita literal). Hoy el kit propio de ArnesIA no declara MCP, así que no
bloquea nada existente — pero si el kit algún día quiere su propio MCP, el mecanismo
sancionado es declararlo en `.mcp.json` del propio kit Y sumarlo al archivo agregado que
pasa `--mcp-config` (no depender de que sobreviva el modo estricto sin más).

**Estado:** `FIRMADA` (2026-07-08, autorización del operador vía prompt de arranque del
paquete — firma sobre D2/D3/D5/D6 y `spec.md` RF-160..RF-167 tal cual redactados).

## D3 · `--setting-sources project,local` (excluir `user`)

**Qué:** todo spawn agrega `--setting-sources project,local` — nunca incluye `user`.

**Por qué:** `user` = `~/.claude/settings.json` del OPERADOR, donde viven sus
`enabledPlugins` personales (marketplaces de golang-*, stitch, etc. — visibles en la
sesión interactiva que disparó el diagnóstico). Excluir `user` es ortogonal a auth (no lo
toca) y el CLAUDE.md del propio cwd del arnés sobrevive de todos modos (documentado:
«CLAUDE.md files... loaded in full at launch», independiente del setting-source).

**Riesgo abierto (bloqueante para firmar sin probar):** NO hay cita oficial explícita de
que excluir `user` corte `enabledPlugins` — es inferencia fuerte, no confirmada. Tampoco
está claro si el walk ascendente de CLAUDE.md (ancestros del cwd por filesystem) queda
contenido por esto o sigue igual (evidencia dice que solo `--bare` lo apaga del todo, y
`--bare` está descartado por D1). Mitigación estructural: auditar que `~/.arnesia/
arneses/**` y sus ancestros no tengan CLAUDE.md perdido — residual aceptado, no
bloqueante, pero documentado.

**Estado:** `FIRMADA` (2026-07-08) — condición de verificación (Fase 0/2, sonda real)
sigue en pie: se ejecuta antes de cementar as-code (RF-165), no antes de codear RF-162.

## D4 · `--safe-mode` NO se adopta a priori

**Qué:** no sumar `--safe-mode` al spawn todavía. Se prueba en un experimento aislado
(Fase 3) DESPUÉS de D2+D3, no antes ni junto.

**Por qué:** preserva auth (a diferencia de `--bare`, dato con cita: no fuerza API key) y
es el único flag que también tapa hooks/output-styles/temas/keybindings ajenos — pero no
hay cita oficial que confirme si `--plugin-dir`/`--append-system-prompt-file`/`--add-dir`
explícitos sobreviven cuando se combinan con `--safe-mode` (un agente de investigación lo
marcó como "conflictivo en espíritu" sin cita dura — no es base suficiente para decidir).
Si mata el kit propio de ArnesIA, es inservible para el producto (defeats the purpose:
arnesia necesita que SU inyección sí cargue).

**Criterio de decisión:** spawn de prueba con `--safe-mode` + los 3 flags de inyección ①②
sobre el dogfood real. Si el kit responde igual que sin `--safe-mode` → se adopta como
capa extra. Si no → se descarta, D2+D3 quedan como la solución completa (ya cierran los
dos ejes de mayor costo/riesgo real: MCP + settings personales).

**Estado:** `FIRMADA-DESCARTADA` (2026-07-08, RF-163 ejecutado —
`experimento-safe-mode.md`). Resultado: `--append-system-prompt-file`/`--add-dir` (①)
SOBREVIVEN a `--safe-mode` (citas textuales exactas verificadas), pero `--plugin-dir`
(②, skills `arnesia-kit:auditar-arnes`/`forjar-caja`) NO — desaparecen del todo. Cumple
el criterio de descarte tal cual estaba escrito arriba. RF-164 no se implementa; D2+D3
quedan como la solución completa.

## D5 · Cementar como extensión de `superficie-local-confinada`, no boundary nueva

**Qué:** los checks nuevos (spawn siempre aísla MCP · spawn siempre excluye `user` de
settings) se agregan a `arch/boundaries/superficie-local-confinada.md` (ya vigente desde
HS-06, S1-S6), no a un boundary nuevo.

**Por qué:** es el mismo concern — confinamiento de la superficie que ve una sesión
spawneada. HS-06 ya confinó el eje filesystem/cwd (S1-S6); esto suma el eje config-source/
MCP al MISMO boundary, coherente con cómo el repo ya organiza `arch/` (espeja
`knowledge/` por tema, no por PR). `permisos-derivan-del-rol` se mantiene sin tocar — ese
boundary es sobre INVOCACIÓN (`--allowedTools`), un eje distinto del de carga-a-contexto
que ataca este paquete.

**Estado:** `FIRMADA` (2026-07-08).

## D6 · Verificación SIEMPRE vía sonda real, nunca `/context` interactivo

**Qué:** cada fase se valida spawneando una sesión real contra el dogfood
(`dev-full-cycle`) con un turno-sonda explícito («lista tus skills/tools/servidores MCP
disponibles ahora mismo»), comparando antes/después. `/context` de una sesión interactiva
del operador NO sirve de medición — es una capa distinta (cuenta personal, no arnés
spawneado), como ya se aclaró al abrir HS-17.

**Por qué:** el propio disparador de esta ficha fue confundir esas dos capas. Método
distinto evita repetir el error y da evidencia real, no inferida, para D2/D3/D4.

**Estado:** `FIRMADA` (2026-07-08).
