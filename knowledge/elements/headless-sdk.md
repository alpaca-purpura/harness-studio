---
elemento: headless
version: 1.1
updated: 2026-07-05
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/headless
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/agent-sdk/overview
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/agent-sdk/migration-guide
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/sessions
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/monitoring-usage
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/best-practices
    autoridad: oficial
    revisado: 2026-07-04
---

# Headless / Agent SDK (conductor programático)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Headless = `claude -p "prompt"` (`--print`): invocación no interactiva que
corre el loop de agente completo y retorna sin TUI. El **Claude Agent SDK** (Python
`claude-agent-sdk`, TS `@anthropic-ai/claude-agent-sdk`) es la forma librería del mismo motor,
para embeber el agente en un programa padre. Es la superficie exacta que un **conductor** usa
para dirigir Claude Code (stdin/stdout o SDK). *(oficial: headless + agent-sdk/overview)*

**Output/input (L1.2):** `--output-format text|json|stream-json`; `--json-schema '<schema>'` +
`json` fuerza salida validada en `structured_output`. `--input-format text|stream-json`. Eventos
stream: `system/init` (session id, model, tools, MCP, plugins), `system/api_retry` (attempt,
retry_delay_ms, error), `stream_event`, `assistant`, `result`. *(oficial: headless)*

**Sesión (L1.3):** `-c`/`--continue` (más reciente en cwd), `-r`/`--resume <id>`,
`--fork-session`, `--no-session-persistence`. Scoped al project dir + worktrees. *(oficial: sessions)*

**Permisos (L1.4):** `--allowedTools`/`--disallowedTools` (rule syntax), `--permission-mode
{default|acceptEdits|plan|auto|dontAsk|bypassPermissions}`, `--dangerously-skip-permissions`
(= bypass). `--bare` salta hooks/skills/plugins/MCP/CLAUDE.md discovery — **«modo recomendado
para scripted/SDK, será el default de `-p`»**; pasar solo lo necesario vía `--settings`/`--mcp-config`/
`--agents`. **⚠ CORRECCIÓN (HS-04, 2026-07-05, frente B):** `--bare` **también salta OAuth/keychain
discovery** → exige `ANTHROPIC_API_KEY`/`apiKeyHelper`. Un conductor que monta el login Pro/Max del
usuario **no puede usar `--bare`** (le rompe el auth de suscripción). Y como `-p` hará `--bare` su
default futuro, **pinear flags explícitos, no heredar el default de `-p`**. *(oficial: headless +
permission-modes + authentication)*

**Transcript (L1.5):** JSONL en `~/.claude/projects/<p>/<session>.jsonl`; **formato interno,
cambia entre versiones** → guía oficial: usar `--output-format json/stream-json`, `/export`, hook
`transcript_path` o el SDK, **no parsear el JSONL directo**. *(oficial: sessions)*

**Prácticas (L1.6):** parsear `json`/`stream-json`, no scrapear texto · capturar `session_id` y
resumir por ID explícito (no `--continue` bajo concurrencia) · `--bare` en CI · scope de permisos
con `--allowedTools`/`dontAsk` (deny-by-default), no blanket skip · `--permission-mode auto`
(clasificador) para autonomía guardada · **`--max-turns` en casi toda corrida desatendida** (cap
de costo/loop) · manejar `system/api_retry` para backoff · paso de verificación adversarial (subagente
revisando el diff) antes de dar por hecho · **OTel** (`CLAUDE_CODE_ENABLE_TELEMETRY=1`) para flotas ·
`settingSources: []` para aislar de drift local. *(oficial: headless + best-practices + monitoring)*

**Anti-patrones (L1.7):** `--dangerously-skip-permissions`/bypass en producción («no protege
contra prompt injection») · parsear texto libre en vez de json · **parsear el JSONL interno como
API estable** · sin `--max-turns` en fan-out (loop/costo runaway) · `Bash(*)` en automatización ·
`--continue` bajo concurrencia · automatización sin telemetría (sin audit trail). *(oficial)*

**Novedades (L1.8):** **rename Claude Code SDK → Claude Agent SDK** (~2025-09-29, secundario):
`@anthropic-ai/claude-code`→`-agent-sdk`, `ClaudeCodeOptions`→`ClaudeAgentOptions` · v0.1.0: el
system prompt ya NO es el default (opt-in `preset:"claude_code"`) · **`--bare`** (futuro default
de `-p`) · modo `auto` (v2.1.83+) · `system/api_retry` event · stdin cap 10MB (v2.1.128) · tracing
distribuido beta (`CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`, span `interaction→llm_request/hook/tool`)
· GitHub Actions v1.0 GA (`claude_args`) · **Managed Agents** (REST hosted) como 3ª opción. Regla
de marca: partners dicen «Powered by Claude», **no** «Claude Code». *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

El headless/SDK es **el motor de la fábrica conversacional**: ArnesIA dirige Claude Code headless
por detrás con el **patrón conductor** (I-76/OBS-16/OBS-18, ya en VISION y CLAUDE.md). No es un
componente DENTRO de un arnés cliente — es cómo ArnesIA OPERA la creación/edición. Su estándar
gobierna al producto ArnesIA mismo (dogfood) más que a los arneses que fabrica.

1. **stream-json + schema, jamás scrapear texto ni el JSONL:** el conductor de ArnesIA consume
   `--output-format stream-json` / SDK, y el **sensor** lee el JSONL como fuente de verdad pero
   sabiendo que su formato es interno (índice desechable — coherente con VISION «SQLite desechable,
   JSONL fuente de verdad», y con la trampa ya vivida en it.7 de parsear JSONL). ⇐ L1.5/L1.6.
2. **Sesión pineada por ID:** cada dock↔sesión CC (S4) resume por `session_id` explícito, nunca
   `--continue` (el dock corre múltiples conversaciones). ⇐ L1.6. Ya modelado en S4/S8.
3. **`--max-turns` + permisos acotados en todo build headless:** el build de un arnés (J3) corre
   con turn-cap y `--allowedTools`, nunca bypass. ⇐ L1.6/L1.7 + principio 10 (nada sin control).
4. **Telemetría de nacimiento vía OTel:** todo run headless emite OTel — es el principio 9
   («telemetría de nacimiento, no opt-in») hecho flag. ⇐ L1.6 + [[hooks]] L2.5 (`telemetry-emit`).
5. **`--bare` SOLO cuando estás deliberadamente sobre API key** (corregido HS-04): aísla el build
   del `~/.claude` del operador, pero **salta el auth de suscripción** — por eso el conductor
   interactivo (que monta el login Pro/Max) NO lo usa; para aislar sin romper auth, pasar el arnés
   explícito vía `--settings`/`--mcp-config`/`--agents` y `settingSources:[]`, dejando el auth
   intacto. `--bare` reservado para CI/fan-out con API key propia. ⇐ L1.4/L1.6.
6. **`% contexto` es métrica DERIVADA nuestra** (corregido HS-04): no existe en stream-json ni OTel;
   se computa `(input+cacheRead+cacheCreation)/ventana-del-modelo` y se etiqueta como derivada (no
   se presenta como medida). Alimenta el header de sesión del dock (S4) y el árbol de corrida (S8).
   ⇐ L1.2.

## Checklist evaluable

> Estos checks aplican sobre todo al **conductor de ArnesIA** y a cualquier automatización que un
> arnés nuestro incluya (CI, fan-out). Son de dogfood además de de conformación.

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| headless-no-blanket-skip | evita `--dangerously-skip-permissions`/bypass; usa allowlist/`dontAsk` | error | banda Guardia «automatización con permisos saltados» | L1.4 · L2.3 |
| headless-structured-output | `-p` scripted usa `--output-format json/stream-json`, no texto | warn | «`-p` sin `--output-format` piped a parser» | L1.6 · L2.1 |
| headless-session-pin | captura `session_id` y resume por ID, no `--continue` bajo concurrencia | warn | «orquestación con `--continue` — ambiguo» | L1.6 · L2.2 |
| headless-bare-ci | CI/scripted **sobre API key** pasa `--bare`; conductor de suscripción aísla con `--settings`/`--mcp-config`/`settingSources:[]` (jamás `--bare`, rompe auth) | info | «CI sin aislar — hereda ambiente» / «`--bare` en conductor de suscripción — auth roto» | L1.4 · L2.5 |
| headless-ctx-derivado | el `% contexto` se computa y se etiqueta como derivado (no existe en stream-json/OTel) | info | «% contexto presentado como medido» | L1.2 · L2.6 |
| headless-max-turns | corridas desatendidas/fan-out fijan `--max-turns`/`maxTurns` | error | «sin turn-cap — loop/costo runaway» | L1.6 · L2.3 |
| headless-perm-scope | tools por `--allowedTools`, no `Bash(*)`/wildcard en automatización | error | banda Guardia «wildcard en config headless» | L1.7 · L2.3 |
| headless-telemetry | flota headless emite OTel (`CLAUDE_CODE_ENABLE_TELEMETRY=1`) | warn | «sin telemetría — sin audit trail» | L1.6 · L2.4 |
| headless-retry-aware | conductor consume `system/api_retry` (backoff/visibilidad) | warn | «stream-json sin manejar api_retry» | L1.6 |
| headless-no-jsonl-parse | el conductor no depende del schema interno del JSONL | error | «parseo directo del JSONL (formato inestable)» | L1.5 · L2.1 |
| headless-verify-gate | tarea desatendida incluye paso de verificación antes de «hecho» | warn | «sin gate de verificación (test/lint/review)» | L1.6 |

## Changelog

- 2026-07-05 · v1.1 · **Corrección load-bearing (HS-04, fase 3, frente B):** `--bare` **también
  salta OAuth/keychain** → rompe el auth de suscripción; el conductor interactivo NO lo usa (se
  aísla con `--settings`/`--mcp-config`/`settingSources:[]`), reservado a CI con API key. Añadido:
  `% contexto` = métrica derivada nuestra (no existe en stream-json/OTel). L2.5 reescrita, L2.6
  nueva. **1 alta neta**: `ctx-derivado` nuevo; `bare-ci` fue **corrección** de un check preexistente
  (v1.0), no alta → **11 checks** (headless v1.0=10 → v1.1=11). Disparado por la
  investigación de arquitectura `historias/2026-07-05-arquitectura-fase3.md`.
- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (headless, agent-sdk/overview +
  migration, sessions, monitoring, best-practices). L2 amarra headless/SDK = motor de la fábrica
  conversacional de ArnesIA (patrón conductor I-76), stream-json + sensor JSONL-como-verdad-índice-
  desechable, sesión pineada (S4/S8), max-turns + permisos acotados, OTel = principio 9. 10 checks
  (dogfood + conformación). Novedades: rename Agent SDK, `--bare`, modo `auto`, tracing beta,
  Managed Agents, GH Actions v1.0. · pasada fundacional.
