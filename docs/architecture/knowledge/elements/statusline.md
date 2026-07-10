---
elemento: statusline
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/statusline
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/settings
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://www.aihero.dev/creating-the-perfect-claude-code-status-line
    autoridad: experto
    revisado: 2026-07-04
  - url: https://github.com/sirmalloc/ccstatusline
    autoridad: experto
    revisado: 2026-07-04
---

# Status line (statusLine)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Barra persistente que corre un comando shell, recibe un JSON de estado por
stdin y muestra su stdout. Corre local, **0 tokens API**; da telemetría de un vistazo (contexto %,
costo, git, modelo) sin gastar un turno conversacional. *(oficial: statusline)*

**Config (L1.2):** en settings.json:
```json
{ "statusLine": { "type": "command", "command": "~/.claude/statusline.sh", "padding": 2 } }
```
`refreshInterval` (≥1s, para relojes/subagentes de fondo), `hideVimModeIndicator`. Sibling
`subagentStatusLine` (una línea JSON por fila del panel de agentes). `/statusline <descripción>`
lo genera. `disableAllHooks: true` también lo apaga. *(oficial)*

**Stdin JSON — campos (L1.3):** `cwd`, `session_id`, `prompt_id?` (v2.1.196+), `transcript_path`,
`version`, `model{id,display_name}`, `workspace{current_dir,project_dir,git_worktree?,repo?}`,
`output_style.name`, `cost{total_cost_usd,total_duration_ms,total_lines_added/removed}`,
`context_window{total_input_tokens,total_output_tokens,context_window_size,used_percentage,
remaining_percentage,current_usage?}`, `exceeds_200k_tokens`, `effort.level?`, `thinking.enabled`,
`rate_limits{five_hour?,seven_day?}`, `vim.mode?`, `agent.name?`, `pr{number?,url?,review_state?}`,
`worktree{...}`. `?` = puede faltar. *(oficial)*

**Contrato de salida (L1.4):** solo stdout (stderr rompe); multi-línea con varios echo; ANSI +
OSC 8 hyperlinks; `COLUMNS`/`LINES` (no `tput cols`, requiere ≥2.1.153); exit≠0 o vacío = barra
en blanco. *(oficial)*

**Prácticas (L1.5):** **cachear shell-outs caros** (git status/diff) a temp keyed por
`session_id` (estable, no PID) con TTL ~5s · output corto (trunca/wrapea) · mostrar **contexto %,
costo, modelo** (dejan decidir compactar sin gastar turno) · manejo defensivo de campos null
(`// 0`, `?.`) · `refreshInterval` solo para segmentos time-based · cross-platform (forward
slashes, PowerShell explícito) · `printf '%b'` > `echo -e` para OSC 8. *(oficial + aihero + ccstatusline)*

**Anti-patrones (L1.6):** `git status`/`diff` sin cache en cada invocación (causa #1 de lag) ·
scripts lentos/bloqueantes (quedan stale) · PID como cache key (cambia siempre) · escribir a
stderr / exit≠0 (blanquea) · multi-línea con ANSI/OSC pesado · asumir campos presentes · recomputar
contexto % con fórmula equivocada (debe ser input-only, mejor confiar en `used_percentage`) ·
(inferido) echo de env/secretos a la barra visible. *(oficial + comunidad)*

**Novedades (L1.7):** `rate_limits` Pro/Max · **semántica de `total_input/output_tokens` cambió
a LIVE (v2.1.132)** — breaking para scripts que sumaban · `COLUMNS`/`LINES` (v2.1.153) ·
`prompt_id` (v2.1.196, correla con OTel `prompt.id`) · `footerLinksRegexes` (v2.1.176, alternativa
no-script) · `subagentStatusLine` · campos `pr.*`/`worktree.*`/`effort`/`thinking`. Versión de
introducción de `statusLine`: no verificada (predata los markers hallados). *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

El statusline es **visibilidad, no dirección** — no cambia lo que Claude hace. En un arnés
nuestro importa por dos razones conectadas al producto: (1) surface honesto de contexto/costo
(coherente con el principio 11 y con lo que ArnesIA mide), y (2) su `prompt_id`/`session_id`
correlan con el mismo JSONL que el **sensor** de ArnesIA consume (S8 Corridas). No es una caja
ni banda; es utilidad de sesión.

1. **Debe surfacear contexto % + costo + modelo:** un arnés nuestro sin statusline deja al
   operador ciego al costo/presión de contexto. ⇐ L1.5 + principio 11. (Coherencia con la
   contabilidad de contexto por paso que ya modelamos en S8.)
2. **Rápido y con cache:** git shell-out cacheado por `session_id`; nada bloqueante. ⇐ L1.5/L1.6.
3. **Cero secretos en la barra:** nunca echo de env/tokens. ⇐ L1.6 (coherente con la regla de
   honestidad y con [[settings-permissions]] L2.5).
4. **Null-safe:** maneja campos ausentes (pre-primera-API, post-compaction). ⇐ L1.5.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| statusline-configured | `statusLine.type=="command"` y el script existe/ejecutable | info | badge «sin statusline — sin visibilidad de contexto/costo» | L1.2 · L2.1 |
| statusline-context | output/source referencia `context_window.used_percentage` | info | badge «no muestra % de contexto — riesgo de compactar tarde» | L1.5 · L2.1 |
| statusline-cost-model | referencia `cost.total_cost_usd` y `model.display_name` | info | badge «no muestra costo/modelo — visibilidad incompleta» | L1.5 · L2.1 |
| statusline-fast | wall-time del comando (mock stdin) <~300ms | warn | badge «statusline lenta — queda stale bajo uso» | L1.5 · L2.2 |
| statusline-git-cache | si hace `git status`/`diff`, cachea por `session_id` con TTL | warn | badge «git sin cache — spawn pesado por refresh» | L1.5 · L2.2 |
| statusline-null-safe | usa fallbacks (`// 0`, `?.`, `or 0`) en campos opcionales | info | badge «no maneja campos null — puede romper/vaciar» | L1.5 · L2.4 |
| statusline-stdout-only | solo escribe stdout (sin stderr / exit≠0 accidental) | warn | badge «puede quedar en blanco por error no manejado» | L1.4 |
| statusline-no-secret | no hace echo de env/tokens/`.env` al output | error | banda Guardia «statusline expone posibles secretos» | L1.6 · L2.3 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (statusline, settings) + aihero +
  ccstatusline. L2 amarra statusline = visibilidad (no dirección), surface honesto contexto/costo
  (principio 11) + correlación `prompt_id`/`session_id` con el sensor de ArnesIA (S8). 8 checks.
  Novedades: `total_input/output_tokens` LIVE (v2.1.132, breaking), `COLUMNS`/`LINES`, `prompt_id`,
  `footerLinksRegexes`, `subagentStatusLine`. · pasada fundacional.
