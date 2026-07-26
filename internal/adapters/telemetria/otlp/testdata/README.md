# `testdata/` — golden files MEDIDOS, no fixtures inventados

Estos archivos son **copias byte por byte** de
`docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/evidencia/`:
payloads que Claude Code `2.1.220` emitió de verdad en las corridas del 2026-07-26 (USD 0,044 de
tokens reales). No se editan a mano.

| archivo | qué trae |
|---|---|
| `logs-run1.json` | 8 log records · `api_request` con los 4 buckets, `cost_usd_micros`, `duration_ms` |
| `logs-run2-con-skill.json` | ídem, con una skill cargada |
| `logs-run4-con-tools.json` | 5 log records · `tool_decision` y `tool_result` con sus tamaños en bytes |
| `metrics-run1.json` | 3 métricas · `Sum · Delta · monotonic` |
| `result-envelope.json` | el frame `result` del stream-json, con el split `ephemeral_5m`/`ephemeral_1h` |

Dos cosas que estos archivos PRUEBAN y por eso están acá y no en un fixture escrito a mano:

1. **`intValue` llega como número JSON**, off-spec (el mapeo protobuf→JSON manda string). Un
   decodificador con `*string` revienta; uno con `*int64` reventaría con un runtime que sí cumple
   la spec. De ahí `json.Number`.
2. **La PII llega en cada log record** (`user.email`, `user.account_uuid`, `user.account_id`,
   `user.id`, `organization.id`). Acá viaja **redactada** (`<REDACTADO-POR-ARNESIA>`) porque son
   datos reales del operador; los tests de allowlist agregan además una fixture con PII **simulada**
   para poder asertar sobre un valor buscable.
