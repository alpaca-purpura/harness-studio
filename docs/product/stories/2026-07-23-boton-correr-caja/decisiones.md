# Decisiones — botón «Correr» de una caja (paquete HS-11)

> Disciplina METODOLOGIA §10: cada decisión conversada se escribe acá EN EL MISMO TURNO.

## D1 — Alcance del paquete (2026-07-23)

**Ciclo completo mockup → decisiones → spec → PARIDAD** (no ticket chico), a pedido del
operador. Backend ya vivo y probado (`CAP-55`, commit `f7a48f6`); lo que falta es 100% FE:
tipos TS, client methods, listener SSE, botón real en `inspector.tsx`, story, capability FE.

## D2 — Mecanismo de feedback en vivo: SSE `event: run`, no polling

**SSE `event: run`** (ya emitido real por el broker — `internal/adapters/transport/sse/broker.go:20`,
publicado en `run_service.go` en `StartRun`/`runAsync`/`finishError`; wireado end-to-end en
`cmd/arnesia/main.go:163,231,298`). `GET …/boxes/{id}/runs/{runId}` se usa SOLO de resync: al
montar el drawer (por si una corrida quedó en curso de una sesión anterior) o si el evento SSE
no llegó. Nunca un loop `setInterval`.

**Por qué:** el backend ya resuelve esto mejor que un poll — menos tráfico, consistente con el
patrón ya usado en `sessions-store.ts` (fetch disparado por evento, no por temporizador). El
único polling real del repo (`pollReinicio`, self-update) es para un caso distinto (esperar que
un binario NUEVO responda tras reiniciar) — no aplica acá, donde el propio proceso ya emite el
desenlace.

**Gap a cerrar:** hoy `web/src/shared/api/sse.ts` (`connectDock`) solo escucha `dock` y `map` —
`event: run` llega al browser y se descarta sin listener. Hay que extender `connectDock` (o
agregar un stream paralelo) para escuchar `run`.

## D3 — Error 409 (precondición incumplida): mensaje inline, no toast

El repo no tiene una primitiva de toast hoy (todo feedback del Inspector es texto en secciones
del drawer). El 409 (`precondicionBody{faltantes}`) se muestra como un `<p class="hallazgo crit">`
dentro de la sección «Última corrida», listando los `faltantes` — mismo patrón visual que
`Hallazgos` (`inspector.tsx:291-332`, ya usa `.hallazgo.crit`/`.hallazgo.warn`). El botón
«Correr» sigue visible y habilitado (el operador puede resolver el bloqueo afuera y reintentar
sin recargar nada).

**Nota de spec pendiente:** el 409 de esta ruta NO está documentado en `openapi.yaml` (gap
código↔spec real, detectado en la investigación) — el spec.md de este paquete debe fijarlo
igual, citando el código (`run.go:43-47`) como fuente, y una tarea aparte (fuera de este
paquete) debería cerrar el gap en el OpenAPI.

## D4 — Qué NO cambia (superset estricto)

- La prosa «Sin corridas indexadas — llegan con el indexer JSONL (Hito 3)» queda INTACTA arriba
  del botón — el listado histórico de corridas sigue sin existir.
- El botón «Ver todas las corridas del arnés» sigue `disabled` (Hito 3, alcance distinto).
- Nodos que no son caja (`box.contract?.caja !== true`) no ven ningún botón — mismo criterio de
  gating que hoy (`isCaja()`, `entities/arnes/model/node-view.ts:8-11`).
- La story `CorridasHonesta` (`inspector.stories.tsx:104-111`) se actualiza para seguir
  cubriendo el estado placeholder + se agregan stories nuevas para el flujo del botón — no se
  borra cobertura existente.

## Mockup

`mockup-boton-correr.html` — publicado como Artifact (gallery de 5 estados: idle interactivo ·
corriendo · éxito con advertencias · error 409 · error 500 · nodo no-caja). Pendiente firma 🧑‍⚖️
del operador antes de escribir `spec.md`.
