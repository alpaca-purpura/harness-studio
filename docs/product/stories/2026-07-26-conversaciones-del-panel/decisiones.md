# Decisiones — Las conversaciones viven en el panel de conversación

> `tipo: decisiones` · paquete `2026-07-26-conversaciones-del-panel` · conversadas con el operador
> el 2026-07-26. Origen: el operador pidió ver el historial de una conversación y crear una nueva,
> y NO encontró ninguna de las dos cosas en el panel. Estado del código verificado en vivo contra
> el binario instalado (`~/.local/bin/arnesia serve --addr 127.0.0.1:4200`, build del 2026-07-26
> 10:11) antes de escribir cada decisión — no de memoria.

## Lo que se verificó en vivo primero (el punto de partida real)

1. **El dock de conversación no tiene historial ni «＋ nueva».** `chat-dock.tsx` pinta solo los
   turnos de la sesión activa; no importa `conversaciones-store` en ningún lado.
2. **El historial B2 existe pero está escondido en el picker del rail** —
   `web/src/widgets/session-rail/ui/new-session-picker.tsx:263` (`ConversacionesDelArnes`): hay que
   apretar `＋ Nueva sesión` y seleccionar un arnés para verlo. Bloque de 10px, `max-h-44`, cada
   turno truncado a una línea.
3. **Y aun así muestra 0.** Verificado contra la API del daemon:
   `GET /api/sessions?arnes=vitalia&cerradas=1` → **2 cerradas**;
   `GET /api/sessions?arnes=sin-home~vitalia~vitalia&cerradas=1` → **0**. El picker consulta con la
   clave calificada (`new-session-picker.tsx:111`, `cargar(e.clave)`) pero
   `~/.arnesia/sesiones-cerradas.json` guarda las 3 cerradas con el **id pelado** (`vitalia`,
   `dev-full-cycle`). El re-key `cca314a` migró las vivas, no las cerradas.

## CV-D1 · `Hist` es el historial del ARNÉS, no de la conversación

La vista `Hist` del view-strip (hoy stub «próximamente») es para los **cambios del arnés en el
tiempo**. No es el lugar del transcript de una conversación y **no se toca en este paquete**.
Corrige la lectura del ejecutor en la sesión previa, que la había propuesto como destino del
historial.

## CV-D2 · Todo lo de conversación vive en el panel de conversación

Crear · listar · buscar · abrir transcript: **todo en el dock**, siguiendo el patrón estándar de
las UIs de chat con IA (lista de conversaciones + buscador + `＋ nueva` + click para cargar). El
bloque `ConversacionesDelArnes` **sale del picker del rail**: el rail crea sesiones, no
conversaciones.

## CV-D3 · Sesión CONTIENE N conversaciones (entidad nueva)

**La decisión de fondo.** Hoy el código es 1:1 — `internal/domain/session.go:68`, `Session` lleva
`ClaudeSessionID`, `Conv []Turn`, `CadenaCC`, `Checkpoint`: la sesión **es** la conversación.

Se parte en dos:

- **Sesión** = el frente de trabajo (arnés + cwd + reparación + vista). Es lo que vive en el rail.
- **Conversación** = un hilo con el conductor dentro de esa sesión. Se lleva `ClaudeSessionID`,
  `Conv`, `CadenaCC`, `Checkpoint`, `CtxPct`/`CtxHist`, `RotacionPendiente`, `CerradaEn`, `Turnos`.

```
sesión B ─┬─ conv 1
          ├─ conv 2
          └─ conv 3   ← el dock ve SOLO estas
```

## CV-D4 · El dock ve SOLO las conversaciones de la sesión activa

Nunca las de otras sesiones ni las de otros arneses. El alcance del listado y del buscador es la
sesión activa, punto.

## CV-D5 · La ley de llaves se arregla, no se hereda

El bug del punto 3 (cerradas por id pelado vs consulta por clave calificada) **no se replica** en
el modelo nuevo: la conversación cuelga de un `session_id`, no de un string de arnés
reconstruido. Qué pasa con las 3 cerradas ya en disco → CV-D6.

## CV-D6 · Las 3 cerradas de hoy se ELIMINAN

`~/.arnesia/sesiones-cerradas.json` (`s1b38a066`, `s408bb085`, `s020210e3`) se borra. El operador
las declaró no importantes. **No se escribe migración de llaves** — el bug muere con los datos, y
CV-D5 impide que vuelva. Es un borrado de datos del operador, autorizado explícitamente el
2026-07-26; se ejecuta con el daemon detenido y con copia previa del archivo.

## CV-D7 · Una conversación viva por sesión, a la vez

Crear una conversación **cierra la anterior**. No hay N hilos vivos en paralelo dentro de una
sesión. El paralelismo real sigue siendo el de siempre: N sesiones en el rail. La lista del dock
es entonces 1 activa + N cerradas.

## CV-D8 · El buscador busca el TEXTO del transcript, no solo el título

Sobre `Conv []Turn` — lo que el operador **vio**, no la JSONL cruda. Medido, no estimado: la
conversación viva más larga en disco hoy (`s6165ac75`) son 90 turnos = **9.4 KB**; 100
conversaciones ≈ 1 MB. Scan en memoria: sin FTS5, sin índice nuevo, sin tocar
`~/.arnesia/index.db`.

**Exige un cambio de comportamiento:** hoy al cerrar una sesión el daemon **tira el `conv`** —
verificado, las 3 cerradas guardan metadata + `turnos: N` y el transcript se reconstruye leyendo
la JSONL nativa (`internal/adapters/history/reader.go`; corpus de 1.2 MB–84 MB por proyecto). Al
cerrar, `Conv` **se persiste**. La JSONL nativa sigue siendo la SSoT del contenido (boundary
`conductor-no-parsea-jsonl`); `Conv` es la copia liviana de presentación, y es sobre esa que se
busca.

## CV-D9 · Título auto-derivado, editable por el operador

Igual que `Frente` hoy (`session.go:73-75`): se deriva del primer mensaje del usuario y el
operador lo puede editar. Ahora vive en la conversación, no en la sesión — la sesión conserva su
propio `Frente` (el nombre del frente de trabajo).

## CV-D10 · La rotación por contexto es INVISIBLE — no corta la conversación

RF-195/198: cuando el uso de contexto cruza el umbral, el próximo turno spawnea un proceso fresco
con checkpoint. Eso es **mecánica**, no cambio de tema: la conversación **sigue siendo la misma
entrada** en la lista. `CadenaCC` (`session.go:109-111`) ya cose los N `ClaudeSessionID` en un
hilo lógico — se conserva tal cual, ahora colgando de la conversación.

En el transcript la rotación se ve como **marca inline** («contexto rotado · checkpoint»), jamás
como corte de hilo.

> Registro honesto: la respuesta inicial del operador fue «rotación = conversación nueva». El
> ejecutor señaló que contradecía su propia regla «conversación = tema» (la rotación la dispara el
> `%` de contexto, no un tema nuevo, así que cortar ahí parte un tema en dos entradas) y el
> operador eligió invisible. La regla que manda es **conversación = tema**.

---

## Abierto (aún NO conversado — no inventar)

- Cómo se **cierra** una conversación desde el dock: ¿botón explícito, o solo implícito al crear
  otra (CV-D7)?
- Si una conversación cerrada es **reanudable** (`--resume` sobre su `ClaudeSessionID`) o es de
  solo lectura.
- Qué muestra cada fila de la lista además del título (fecha · nº de turnos · ctx final).
