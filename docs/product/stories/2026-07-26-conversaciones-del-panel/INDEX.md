# Paquete — Las conversaciones viven en el panel de conversación

> Origen: el operador (2026-07-26) pidió ver el historial de una conversación y crear una nueva.
> No encontró ninguna de las dos en el panel. La verificación en vivo le dio la razón: el
> historial existe pero está escondido en el picker del rail, y ahí muestra 0 por un bug de
> llave. Ver [`decisiones.md`](./decisiones.md).

## Estado

| Etapa | Estado |
|---|---|
| decisiones | ✅ **completas** — CV-D1..D13, cero puntos abiertos |
| mockup | ⏳ **escrito, esperando firma 🧑‍⚖️** — [`mockup-conversaciones-panel.html`](./mockup-conversaciones-panel.html), 7 secciones, registrado en [`mockups/INDEX.md`](../../../../mockups/INDEX.md) |
| spec | ⬜ |
| implementar | ⬜ |
| PARIDAD | ⬜ |

## Lo firmado hasta acá

- **CV-D1** — `Hist` = historial del arnés, no de la conversación. Fuera de alcance.
- **CV-D2** — crear · listar · buscar · abrir transcript: todo en el dock. Sale del picker del rail.
- **CV-D3** — **sesión CONTIENE N conversaciones**. Parte `Session` (`internal/domain/session.go:68`)
  en dos entidades.
- **CV-D4** — el dock lista SOLO las conversaciones de la sesión activa.
- **CV-D5** — la conversación cuelga de un `session_id`, no de un string de arnés.
- **CV-D6** — las 3 cerradas de hoy se eliminan; sin migración de llaves.
- **CV-D7** — una conversación viva por sesión; crear cierra la anterior.
- **CV-D8** — el buscador busca el texto del transcript (`Conv`), que ahora **se persiste al cerrar**.
- **CV-D9** — título auto-derivado del primer mensaje, editable.
- **CV-D10** — la rotación por contexto es invisible: misma conversación, marca inline.
- **CV-D11** — seleccionar una inactiva la **retoma** (`--resume`); siempre una activa por sesión.
- **CV-D12** — vocabulario **activa/inactiva**, no «cerrada». El registro `sesiones-cerradas.json` pierde sentido.
- **CV-D13** — la fila muestra última interacción · nº turnos · ctx final. Exige timestamp nuevo por turno.

## Radio de impacto (relevado, no estimado)

| Capa | Qué se toca |
|---|---|
| Dominio | `internal/domain/session.go:68` — partir `Session`; `ClaudeSessionID`/`Conv`/`CadenaCC`/`Checkpoint`/`CtxPct`/`CtxHist`/`RotacionPendiente` migran a `Conversacion` |
| Persistencia | `~/.arnesia/sessions.json` + `sesiones-cerradas.json` (registro `store.NewRegistry`, `cmd/arnesia/main.go:325`) — formato nuevo + migración de lo existente |
| API | `/api/sessions?arnes=…&cerradas=1` y `/api/sessions/cerradas/{id}/historial` (`web/src/shared/api/client.ts:214-225`) → endpoints por sesión |
| FE store | `web/src/shared/store/sessions-store.ts` + `web/src/widgets/session-rail/model/conversaciones-store.ts` (se muda al dock) |
| FE UI | `web/src/widgets/chat-dock/ui/chat-dock.tsx` (gana lista+buscador+＋); `new-session-picker.tsx:263` (pierde `ConversacionesDelArnes`) |
| Capabilities | ningún cambio de código sin capability (doctrina `codigo-traza-a-capability`) — CAP nuevas por definir en spec |

## Retomar aquí

**Gate 1 🧑‍⚖️: mirar el mockup y firmarlo o pedir iteración.** Abrir
[`mockup-conversaciones-panel.html`](./mockup-conversaciones-panel.html) en el navegador (alterna
tema con el botón de arriba). Recién con la firma sigue la etapa 3 (`spec.md` + `design.md` +
`plan-pruebas.md`).

Lo que el mockup deja abierto a propósito, para resolver en spec: ancho de la lista con el dock
estirado (CH-D5) · búsqueda incremental o con Enter · orden de la lista · qué pasa con permisos
pendientes de la conversación que se desactiva.

Dos gaps del modelo que la spec tiene que construir, no dibujar:

1. **«última interacción» no existe** — `session.go` no guarda timestamp por turno.
2. **`Conv` se tira al desactivar** — sin persistirlo no hay búsqueda por texto (CV-D8) ni repintado
   al retomar (CV-D11).

El bug de llave que motivó el paquete (cerradas por id pelado vs consulta por clave calificada)
**ya no se arregla**: CV-D6 borra esos datos y CV-D5 impide que reaparezca.
