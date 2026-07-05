---
regla: conductor-no-parsea-jsonl
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-04
sources:
  - url: https://code.claude.com/docs/en/headless
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/monitoring-usage
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/sessions
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestNoJSONLSchemaParsing
  - fitness/arch_test.go:TestLiveEventsFromStreamJSON
severity: high
---

# El conductor consume stream-json/OTel; nunca parsea el JSONL como API estable

## L1 · Principio (estándar de industria)

**Event sourcing sobre contratos estables, no sobre formatos internos.** Claude Code emite
eventos en vivo por `--output-format stream-json` (contrato soportado) y telemetría por OTel; el
transcript JSONL de `~/.claude` es **formato interno que cambia entre versiones** — la guía
oficial dice explícitamente: no parsear el JSONL directo; usar stream-json / `/export` / hook
`transcript_path` / SDK. *(oficial: headless, sessions, monitoring-usage)*

## L2 · Realización (este árbol Go+React)

Event sourcing en 3 fuentes con roles fijos (la trampa de parsear JSONL ya se vivió en UX it.7):

- **stream-json = live/primario.** El adaptador conductor consume NDJSON de stdout y **captura el
  stream a NUESTRO event store** — ese store es la API estable interna. De aquí salen: turnos,
  `tool_use`/`tool_result`, deltas de tokens/TTFT, `result` (costo/usage/`permission_denials`),
  `system/compact_boundary`, `system/api_retry`. Skill activada = un `tool_use` `Skill`. ⇐ L1.
- **OTel = sidechannel** (`CLAUDE_CODE_ENABLE_TELEMETRY=1`, endpoint → el daemon) para lo que
  stream-json no limpia: `hook_*` (los hooks corren DENTRO del CLI → **solo OTel los ve**),
  `skill_activated` (nombre limpio), `tool_decision`, `mcp_server_connection`, `auth`. ⇐ L1.
- **JSONL = enumerar/replay** de sesiones pasadas (listar corridas, reconstruir el índice), **nunca
  parsear su schema como contrato**. Es puntero, no API. ⇐ L1.
- **`% contexto` es métrica derivada nuestra** (no existe en stream-json ni OTel):
  `(input+cacheRead+cacheCreation)/ventana`. Se computa en el adaptador, se etiqueta como derivada.
- **Corrección al nodo [`knowledge/headless-sdk`](../../knowledge/elements/headless-sdk.md)
  (frente B):** `--bare` **rompe el auth de suscripción** (salta OAuth/keychain) → no default-earlo
  para el login Pro/Max del operador; y `-p` hará `--bare` default futuro → **pinear flags
  explícitos**.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| no-jsonl-parse | ningún caso de uso decodifica el schema interno del JSONL (solo enumera/replaya) | error | «parseo directo del JSONL (formato inestable)» | arch_test.go:TestNoJSONLSchemaParsing |
| live-desde-stream-json | los eventos vivos del dock salen de stream-json, no de tail del JSONL | error | «dock alimentado del JSONL en vez de stream-json» | arch_test.go:TestLiveEventsFromStreamJSON |
| hooks-desde-otel | eventos de hook/skill limpios vienen de OTel, no inferidos del texto | warn | «hooks/skills sin fuente OTel — inferidos» | arch_test.go |
| ctx-derivado-etiquetado | `% contexto` se computa y se marca como métrica derivada | info | «% contexto presentado como si fuera medido» | arch_test.go |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = event sourcing sobre contratos estables. L2:
  stream-json (live) + OTel (hooks/skills) + JSONL (enumerar/replay, nunca parsear); `%contexto`
  derivado; corrección `--bare` rompe auth de suscripción. 4 checks.
