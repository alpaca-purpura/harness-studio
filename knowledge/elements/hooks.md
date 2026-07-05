---
elemento: hook
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/hooks
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/hooks-guide
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/security
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://research.checkpoint.com/2026/rce-and-api-token-exfiltration-through-claude-code-project-files-cve-2025-59536/
    autoridad: experto
    revisado: 2026-07-04
  - url: https://claudelog.com/mechanics/hooks/
    autoridad: experto
    revisado: 2026-07-04
  - url: https://ranjankumar.in/hooks-policy-as-code-agent-enforcement
    autoridad: experto
    revisado: 2026-07-04
---

# Hooks (eventos de ciclo de vida)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Comandos shell (o http/mcp/prompt/agent) que corren automáticamente en puntos
fijos del ciclo. Dan **control determinista**: reglas que SIEMPRE disparan (formateo, bloquear
comandos peligrosos, auditoría) en vez de confiar en el juicio del modelo. Para juicio que
necesita razonamiento existen types `prompt`/`agent` como punto medio. *(oficial: hooks-guide)*

**Dónde (L1.2):** `~/.claude/settings.json`, `.claude/settings.json` (compartible), `.local.json`,
managed, plugin `hooks/hooks.json`, o frontmatter de skill/subagente (scoped). *(oficial)*

**Eventos (L1.3):** la superficie creció a **~30 eventos**. Los core: `SessionStart`,
`UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Notification`, `Stop`, `SubagentStop`,
`SessionEnd`, `PreCompact`. Nuevos (multi-agente/teams/MCP/watching): `Setup`, `PermissionRequest`,
`PermissionDenied`, `PostToolUseFailure`, `PostToolBatch`, `SubagentStart`, `TaskCreated`,
`TaskCompleted`, `StopFailure`, `TeammateIdle`, `InstructionsLoaded`, `ConfigChange`, `CwdChanged`,
`FileChanged`, `WorktreeCreate`/`Remove`, `PostCompact`, `Elicitation`/`Result`, `MessageDisplay`. *(oficial: hooks)*

**Types y matchers (L1.4):** 5 types: `command` (shell), `http`, `mcp_tool`, `prompt` (Haiku
sí/no), `agent` (**experimental**, prefiere command en prod). Matcher = string exacto/lista `|`/
regex/`*`. Campo `if` (permission-syntax `Bash(git *)`, v2.1.85+) filtra por nombre+args pero es
**best-effort, fail-open** → «usa el sistema de permisos para un allow/deny duro, no un hook». *(oficial)*

**Exit codes (L1.5, clave):** `0` = ok (stdout parseado como JSON) · **`2` = bloqueante** (stdout
ignorado, stderr = razón) · **otro = NO bloqueante** (continúa). El error #1 de novatos: usar
exit `1` para bloquear — no bloquea nada. Qué eventos pueden bloquear con exit-2 varía (PreToolUse,
UserPromptSubmit, Stop, PreCompact… sí; PostToolUse, Notification, SessionStart… no, ya pasó). *(oficial)*

**JSON output (L1.6):** `continue`, `suppressOutput`, `systemMessage`; `decision`/`reason`;
`additionalContext` (inyecta contexto que Claude lee como texto — un `UserPromptSubmit` usa esto
en vez de bloquear); `hookSpecificOutput.permissionDecision` (`allow`/`deny`/`ask`/`defer`) +
`updatedInput` (reescribe args, v2.0.10+) para PreToolUse. Strings cap 10.000 chars. Timeouts:
command/http/mcp 600s default (UserPromptSubmit 30s), `prompt` 30s, `agent` 60s. *(oficial)*

**Seguridad (L1.7, verbatim):** «Command hooks corren con TODOS los permisos de tu usuario…
revisa y prueba todo hook antes de añadirlo». Bullets: validar/sanitizar input · **siempre
comillas `"$VAR"`** · bloquear traversal `..` · **rutas absolutas** o `${CLAUDE_PROJECT_DIR}` ·
saltar `.env`/`.git/`/keys. `allowManagedHooksOnly` para lockdown org. **CVE-2025-59536** (feb
2026): un hook `SessionStart` de repo no confiable corría al abrir la carpeta (RCE/reverse-shell
antes de review) — fix: diálogo de config-no-confiable reforzado. *(oficial + Check Point)*

**Prácticas (L1.8):** casar evento con trabajo (**PreToolUse** = único que previene antes;
**PostToolUse** = formateo/validación/audit; **Stop** = gate de completitud; **SessionStart** =
cargar contexto) · exit-2 para bloquear · Stop hook revisa `stop_hook_active` (cap de 8 bloqueos) ·
**PreToolUse rápido (<500ms)**, scans/tests pesados a PostToolUse/async · `additionalContext` para
informar, no bloquear · exec form (`args:[]`) evita bugs de quoting · commitear hooks de seguridad
del equipo a `.claude/settings.json` · auditar `ConfigChange`. *(oficial + claudelog + Ranjan Kumar)*

**Anti-patrones (L1.9):** exit-1 para «bloquear» (no hace nada) · `$VAR` sin comillas · rutas
relativas · enforcement de regla dura solo por prompt (el modelo racionaliza la violación) ·
PreToolUse pesado (latencia en cada tool call) · dos PreToolUse reescribiendo el mismo `updatedInput`
(carrera) · perfil de shell con `echo` incondicional que corrompe el JSON stdout · matcher
`PermissionRequest` `.*` que auto-aprueba todo. *(oficial + expertos)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Los hooks son la **banda Guardia** del mapa (VISION A6: infraestructura transversal, actúa sobre
TODAS las cajas). Ya modelamos 6 hooks reales de luana (auto-chain, learning-detect, overlay-check,
contract-guard, validate-session-close, telemetry-emit). El hook es donde una regla pasa de
advisory a **garantía**.

1. **Regla dura ⇒ hook, no solo prosa:** toda «nunca X» destructiva/de-secreto de [[rules]] tiene
   su PreToolUse que la enforca. ⇐ L1.1/L1.8 + [[rules]] L2.2. El par regla↔hook es un cross-check.
2. **Exit-2 para bloquear, verificado:** un hook «guardián» que use exit-1 = bug silencioso
   (no protege nada). ⇐ L1.5. Check de error.
3. **PreToolUse liviano:** la Guardia no puede meter latencia en cada tool call; scans pesados a
   PostToolUse/async. ⇐ L1.8.
4. **Guardia versionada y de fuente confiable:** al importar (conformación §6) se valida que los
   hooks estén en config versionada y no vengan de repo no confiable (lección CVE-2025-59536). ⇐ L1.7.
5. **`telemetry-emit` = el sensor:** el hook de telemetría del kit (KIT-03, OTLP GenAI) es el
   egress que ArnesIA consume — no se instrumenta de cero. ⇐ L1.3 (`Stop`/`PostToolUse` como
   puntos de emisión) + METODOLOGIA §5. *(pieza de producto, no solo check)*

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| hook-exit2-block | hook con lógica de bloqueo usa exit `2`, no `1`/otro | error | banda Guardia «hook usa exit 1: no bloquea nada» | L1.5 · L2.2 |
| hook-quote-vars | strings de command referencian `$VAR` sin comillas | warn | «variable sin comillas — injection/globbing» | L1.7 |
| hook-timeout | declara `timeout` explícito para script no trivial | info | «sin timeout (default 10min puede colgar la corrida)» | L1.6 |
| hook-abs-path | `command` con ruta absoluta o `${CLAUDE_PROJECT_DIR}` | error | «ruta relativa — falla si cwd difiere» | L1.7 |
| hook-sensitive-guard | PreToolUse Edit/Write chequea `.env`/`.git/`/keys antes de permitir | warn | «sin guardia de archivos sensibles» | L1.7 |
| hook-stop-reentrant | Stop/SubagentStop bloqueante revisa `stop_hook_active` | warn | «riesgo de bucle: no revisa stop_hook_active (cap 8)» | L1.8 |
| hook-committed | hooks de seguridad en config versionada, no solo local | info | banda Guardia «guardia no versionada — no compartida» | L1.8 · L2.4 |
| hook-pretooluse-light | PreToolUse sin test suites/curl/scans pesados inline | warn | «PreToolUse pesado — latencia en cada tool call» | L1.8 · L2.3 |
| hook-single-updatedinput | ≤1 PreToolUse por matcher reescribe `updatedInput` | warn | «múltiples hooks reescriben input — orden no determinista» | L1.9 |
| hook-permreq-scoped | `PermissionRequest` que auto-`allow` usa matcher estrecho, nunca `*` | error | banda Guardia «auto-approve sin matcher — aprueba todo» | L1.9 |
| hook-if-not-security | hook que confía solo en `if` para deny crítico sin regla de permiso de respaldo | warn | «`if` es fail-open — no es límite de seguridad» | L1.4 |
| hook-rule-pair | regla dura de banda Base tiene su hook que la enforca | warn | «regla sin hook — solo advisory» | L2.1 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (hooks, hooks-guide, security) +
  Check Point (CVE-2025-59536) + claudelog + Ranjan Kumar. L2 amarra hooks = banda Guardia, par
  regla↔hook (advisory→garantía), exit-2 verificado, PreToolUse liviano, `telemetry-emit` = sensor.
  12 checks. Novedades: ~30 eventos (de ~9), `updatedInput` v2.0.10, `PermissionRequest`, 5 types,
  async, CVE de config no confiable. · pasada fundacional.
