---
regla: sesion-viva-consistente
version: 1.1
updated: 2026-07-08
status: enforced
ledger: HS-06
sources:
  - url: https://html.spec.whatwg.org/multipage/server-sent-events.html
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://go.dev/ref/mem
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/sessions
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestOneTurnAtATime
  - fitness/arch_test.go:TestFramesCarryRunID
  - fitness/arch_test.go:TestNoSilentEventDrop
  - fitness/arch_test.go:TestResumeAutoSana
severity: high
---

# El pipe conductor↔dock es confiable: guardado, sin pérdida, idempotente, auto-sana

## L1 · Principio (estándar de industria)

**Un stream de eventos que maneja una máquina de estados debe ser lifecycle-guardado, sin pérdida
silenciosa e idempotente ante reentrega.** Cuatro propiedades, cada una de un estándar:

- **Un productor por recurso serializado.** Dos turnos escribiendo el mismo stdin intercalan JSON y
  corrompen el buffer que se ensambla; el modelo de concurrencia exige un solo escritor a la vez.
  *(experto: Go memory model — ownership del canal)*
- **Sin pérdida silenciosa; back-pressure o reconexión.** Dropear un frame (buffer lleno) deja la UI
  colgada (un `result` perdido = «streaming» eterno). O se hace back-pressure al productor, o se
  descarta al *consumidor lento* para que **reconecte y replaye por `Last-Event-ID`** — nunca se
  pierde en silencio. *(oficial: WHATWG SSE — Last-Event-ID + reconnection)*
- **Frames idempotentes.** Una reconexión SSE replaya; aplicar dos veces un `result` duplica el
  turno. Cada frame lleva un id de turno (`run_id`) y el consumidor aplica una sola vez. *(oficial:
  SSE replay; patrón at-least-once → dedup)*
- **Resume que auto-sana.** El `session_id` de Claude Code puede quedar stale (JSONL GC'd); reintentar
  el mismo id falla en loop. Detectar el fallo y reiniciar fresh una vez es la recuperación mínima.
  *(oficial: sessions — resume por id explícito)*

## L2 · Realización (este árbol Go+React)

Todo vive en `internal/usecase/session_service.go` + los adaptadores conductor/SSE + el store FE:

- **un-turno-a-la-vez.** `Turn` rechaza con `ErrBusy` si `status == streaming` → el transporte
  responde **409**; el FE (`sessions-store.ts`) también guarda antes del optimistic. ⇐ L1 productor.
- **sin-perdida-silenciosa.** El conductor (`conductor.go` `emit`) hace **send bloqueante** (el
  consumer siempre drena → back-pressure correcto, cero drop). El broker SSE (`broker.go`) **descarta
  al subscriber rezagado** (cierra su `done`) en vez de dropear el frame → el `EventSource` reconecta
  y replaya del history (256) por `Last-Event-ID`. ⇐ L1 sin-pérdida.
- **frames-idempotentes-run-id.** `dockFrame` lleva `run_id` (= el turno en vuelo, `curRun`) estampado
  en TODO frame (init/delta/result/error), no solo status. El FE trackea `finalizedRun[id]` y **dropea
  cualquier frame de un run ya terminado** (replay o delta tardío). ⇐ L1 idempotencia.
- **resume-auto-sana.** `tryHealResume`: si un spawn con `--resume` muere/errorea **antes de un
  `init`**, se limpia el `ClaudeSessionID` stale, se respawnea fresh **una vez** (`resumeRetried`) y se
  reenvía el turno pendiente. Silencioso en éxito; solo un reinicio fallido emite error. ⇐ L1 resume.
- **`--max-turns` siempre.** Toda corrida fija el cap (`SpawnOpts.MaxTurns` → `conductor.go`); realiza
  el check `max-turns-siempre` de [`permisos-gui-human-in-the-loop`](./permisos-gui-human-in-the-loop.md)
  (no se duplica aquí) y contiene el threat model de [`superficie-local-confinada`](./superficie-local-confinada.md).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| un-turno-a-la-vez | `Turn` rechaza (ErrBusy→409) un turno mientras la sesión está streaming | error | «turnos concurrentes intercalan stdin/ensamblado» | arch_test.go:TestOneTurnAtATime |
| sin-perdida-silenciosa | ni el conductor ni el broker dropean frames en silencio (bloquean o reconectan+replayan) | error | «frame perdido → dock colgado en streaming» | arch_test.go:TestNoSilentEventDrop |
| frames-idempotentes-run-id | todo frame del dock lleva `run_id`; el consumidor aplica el terminal una sola vez | error | «replay SSE duplica turnos (sin run_id/dedup)» | arch_test.go:TestFramesCarryRunID |
| resume-auto-sana | un `--resume` fallido antes de init reinicia fresh una vez y reenvía el turno | warn | «sesión CC stale = frente muerto permanente» | arch_test.go:TestResumeAutoSana |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-06) — ejecuta el scope multisesión cementado en «Siguiente»
  de HS-05 (c.1 run_id · c.3 max-turns · c.4 contención) tras la auditoría del shell Tauri v1. L1 =
  stream confiable (productor único · sin-pérdida · idempotente · auto-sana). L2 = guard ErrBusy→409,
  emit bloqueante + broker shed-on-lag, `run_id` en todo frame + dedup FE, `tryHealResume`. Referencia
  (no duplica) `max-turns-siempre` de permisos-gui. **Nace `enforced`** (código + tests juntos). 4 checks.
- 2026-07-08 · v1.1 · Auditoría colateral (HS-16): `resume-auto-sana` tenía enforcer genérico
  `arch_test.go` (comportamiento ya implementado, solo faltaba el test nombrado).
  `TestResumeAutoSana` fuerza un `--resume` que muere antes de `init` con fakes y confirma el
  heal (respawn fresh sin el id stale + reenvío del turno pendiente). Sin checks nuevos.
