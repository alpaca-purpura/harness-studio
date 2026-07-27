---
regla: sesion-viva-consistente
version: 1.2
updated: 2026-07-26
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
  - fitness/arch_test.go:TestTransicionDeConversacionEsAtomica
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

### El sujeto, re-enunciado: una sesión tiene N conversaciones y exactamente UNA activa (v1.2)

Desde que `Session` se parte en **Sesión** (el frente de trabajo) y **Conversación** (el hilo con el
conductor), «la sesión viva» dejó de ser una frase sin ambigüedad. Se enuncia una vez, acá:

> **El pipe es de la conversación ACTIVA, y hay exactamente una por sesión.** Por eso el runtime
> (`sessionRuntime`, `session_service.go:85-109`) sigue siendo **uno por sesión** y **ninguno de sus
> 12 campos se convierte en un mapa por conversación**: no hay dos hilos que puedan estar en vuelo a
> la vez. Lo que el runtime gana es un puntero, `convActiva`, que dice a quién le pertenece.

Consecuencias exactas sobre los cuatro invariantes de v1.0, que **siguen valiendo sin cambiar de
texto** — cambia sólo a qué se refiere «la sesión»:

- **un-turno-a-la-vez** se sigue guardando por sesión, que bajo la invariante «una activa» **es** por
  conversación. No hace falta un segundo guard.
- **frames-idempotentes-run-id**: `runSeq` se queda **por sesión y monótono**
  (`run_id = "<sessionID>-r<n>"`, `session_service.go:357-358`). Es deliberado: un contador por
  conversación reiniciaría en `r1` al retomar una vieja y **colisionaría** con un `run_id` que el FE
  ya tiene en `finalizedRun[id]`, haciéndole dropear frames legítimos. La invariante se mantiene
  con el contador que ya existe, no con uno nuevo.
- **sin-perdida-silenciosa** y **resume-auto-sana** son del proceso, y el proceso siempre es el de la
  activa. `tryHealResume` limpia el `ClaudeSessionID` **de la conversación**, no de la sesión.

Y **nace la quinta propiedad**, que es lo único que el reparto agrega:

- **transicion-de-conversacion-atomica.** Crear o retomar es **una** transición, no dos operaciones
  encadenadas: exige turno quieto (`status ∉ {streaming, await}` ⇒ **409**, mismo `ErrBusy` que
  `Turn`), cierra el conductor vivo, **deniega los permisos pendientes con motivo** antes de cerrar
  (mismo precedente que `Interrupt`, `session_service.go:848-852` y su publish de `:879`), resetea el
  estado de vuelo (`pendingTurn`/`wasResume`/`sawInit`/`resumeRetried`/`assembling`/`msgFlushed` y
  los `grants` efímeros) y deja **exactamente una activa**. Un fallo a mitad **no es observable**:
  o la transición ocurrió entera, o el estado anterior quedó intacto. ⇐ L1 productor único: dos
  conductores compartiendo el stdin de una sesión es exactamente lo que la propiedad 1 prohíbe.

  Los `grants` se **descartan** en la transición (no se heredan al hilo nuevo): un grant es una
  aprobación que el operador dio dentro de un hilo, y heredarla es aprobar algo que nunca vio.
  ⇐ deny-by-default de [`permisos-gui-human-in-the-loop`](./permisos-gui-human-in-the-loop.md).

**Nota de honestidad heredada, no introducida por v1.2:** dos de las 14 llamadas a `s.publish`
(`session_service.go:626` y `:812`) **no estampan `run_id`**, y `TestFramesCarryRunID`
(`arch_test.go:1924`) sólo ejercita `init`/`delta`/`result`, así que no las caza. Es un hueco del
check, medido el 2026-07-26 y **anotado al BACKLOG**; no se tapa acá ni se le cambia el veredicto al
check. El frame nuevo `conversacion` **tampoco lleva `run_id`** —y es correcto: no pertenece a un
turno—, así que su idempotencia se resuelve por otro camino (índice del turno en `Conv`), declarado
en el check nuevo.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| un-turno-a-la-vez | `Turn` rechaza (ErrBusy→409) un turno mientras la sesión está streaming | error | «turnos concurrentes intercalan stdin/ensamblado» | arch_test.go:TestOneTurnAtATime |
| sin-perdida-silenciosa | ni el conductor ni el broker dropean frames en silencio (bloquean o reconectan+replayan) | error | «frame perdido → dock colgado en streaming» | arch_test.go:TestNoSilentEventDrop |
| frames-idempotentes-run-id | todo frame del dock lleva `run_id`; el consumidor aplica el terminal una sola vez | error | «replay SSE duplica turnos (sin run_id/dedup)» | arch_test.go:TestFramesCarryRunID |
| resume-auto-sana | un `--resume` fallido antes de init reinicia fresh una vez y reenvía el turno | warn | «sesión CC stale = frente muerto permanente» | arch_test.go:TestResumeAutoSana |
| transicion-de-conversacion-atomica | crear/retomar es UNA transición: exige turno quieto (409), cierra el conductor, deniega los permisos pendientes con motivo, descarta los grants y deja exactamente una activa; un fallo deja el estado anterior intacto | error | «dos conductores sobre el mismo stdin, o una sesión sin conversación activa» | arch_test.go:TestTransicionDeConversacionEsAtomica |

## Changelog

- 2026-07-26 · v1.2 · **el quinto check deja de estar pendiente** (T15 del mismo paquete):
  `TestTransicionDeConversacionEsAtomica` existe y pasa. Afirma las cinco cosas que el check
  enuncia, en una corrida contra el servicio real: ErrBusy con el turno en vuelo (para crear y
  para retomar), `Close()` del conductor que se desactiva, `permission_result` deny con un motivo
  que nombra la desactivación y no se confunde con el de `Interrupt`, el mismo tool volviendo a
  preguntar en el hilo nuevo (grants descartados) y —con un store que falla a pedido— el rollback
  que deja el estado anterior intacto **sin** cerrar el conductor. El nodo pasa a **5/5 con
  enforcer real**; los 4 originales siguen verdes **sin que se tocara una línea de su cuerpo**.
  El hueco de `frames-idempotentes-run-id` (2 de 14 `publish` sin `run_id`) **sigue abierto y
  sigue en el BACKLOG**: este enforcer no lo tapa ni le cambia el veredicto.
- 2026-07-26 · v1.2 · **Re-enunciación del sujeto** (paquete
  `stories/2026-07-26-conversaciones-del-panel/`, CV-D3/CV-D7 firmadas): `Session` se parte en
  Sesión + Conversación, así que «la sesión viva» necesitaba decir de quién habla. **Los 4 checks de
  v1.0/v1.1 no cambian de texto ni de veredicto** — cambia el L2, que ahora explica por qué el
  runtime sigue siendo uno por sesión (la invariante «exactamente una activa» lo hace correcto) y
  por qué `runSeq` **no** baja a la conversación (un contador por hilo colisionaría con
  `finalizedRun[id]` del FE al retomar). **+1 check**: `transicion-de-conversacion-atomica`, la única
  propiedad que el reparto agrega — crear/retomar como transición única, con turno quieto, denegación
  de permisos pendientes y descarte de grants. Nace **sin enforcer escrito**
  (`TestTransicionDeConversacionEsAtomica` está nombrado, no existe): el nodo **conserva
  `enforced`** porque sus 4 checks originales siguen verdes, y el quinto se declara pendiente en
  vez de fabricarle un pass. Se anota además, sin taparlo, un hueco medido del check
  `frames-idempotentes-run-id`: 2 de los 14 `s.publish` no estampan `run_id` y el enforcer no los
  ejercita (`session_service.go:626`, `:812`) → BACKLOG. 5 checks.
- 2026-07-05 · v1.0 · Nodo fundacional (HS-06) — ejecuta el scope multisesión cementado en «Siguiente»
  de HS-05 (c.1 run_id · c.3 max-turns · c.4 contención) tras la auditoría del shell Tauri v1. L1 =
  stream confiable (productor único · sin-pérdida · idempotente · auto-sana). L2 = guard ErrBusy→409,
  emit bloqueante + broker shed-on-lag, `run_id` en todo frame + dedup FE, `tryHealResume`. Referencia
  (no duplica) `max-turns-siempre` de permisos-gui. **Nace `enforced`** (código + tests juntos). 4 checks.
- 2026-07-08 · v1.1 · Auditoría colateral (HS-16): `resume-auto-sana` tenía enforcer genérico
  `arch_test.go` (comportamiento ya implementado, solo faltaba el test nombrado).
  `TestResumeAutoSana` fuerza un `--resume` que muere antes de `init` con fakes y confirma el
  heal (respawn fresh sin el id stale + reenvío del turno pendiente). Sin checks nuevos.
