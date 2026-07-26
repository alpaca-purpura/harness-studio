# Spec funcional · Las conversaciones viven en el panel de conversación (RF-300…RF-357)

> `tipo: spec` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Etapa 3 del flujo (METODOLOGIA §10), habilitada por el **🧑‍⚖️ GATE 1 firmado el 2026-07-26**
> ([`INDEX.md`](./INDEX.md) §GATE 1: el operador firmó el mockup en su iteración 2).
> Par de [`design.md`](./design.md) (el UI al pixel) y [`plan-storybook.md`](./plan-storybook.md)
> (qué story asserta qué).
>
> **Numeración.** Se usa **RF-300 en adelante**. Verificado antes de escribir: el último RF en uso en
> todo `docs/` es **RF-286** (`stories/2026-07-24-telemetria-embebida-otel/spec.md`), y
> `grep -rhoE "RF-3[0-9]{2}" docs/` devuelve **vacío** — el bloque 300 está libre entero.
>
> **El QUÉ, no el cómo.** Componentes, props, tokens y estados visuales viven en `design.md`.
> Las decisiones **CV-D1..CV-D15 están FIRMADAS y no se relitigan**: donde una resulta inviable, este
> documento lo dice con la evidencia y elige la alternativa más cercana a su espíritu (§2).
>
> **Trazabilidad.** Cada RF 🎨 cita `mockup-conversaciones-panel.html:<línea>` — línea real,
> verificada contra el archivo de 961 líneas. Cada afirmación sobre el código vigente cita
> `archivo:línea`, leída hoy; nada de memoria. Los RF sin superficie lo declaran en §L.

## Estado de este documento — LEER ANTES DE CONSTRUIR

| | |
|---|---|
| ✅ **Firmado que este spec ejecuta** | Mockup 🧑‍⚖️ (it. 2) + **CV-D1…CV-D15**, las 15 |
| 🔬 **Verificado en vivo hoy** | daemon `0.2.23.2607261611` en `127.0.0.1:4200`: `GET /api/sessions` → **5 sesiones**, la más larga con **90 turnos / 12 281 B de `conv`**; `~/.arnesia/sesiones-cerradas.json` → **3 entradas**, las tres **sin `conv`** y con `turnos` sólo en dos (`s020210e3` no tiene ni `turnos` ni `claude_session_id`) |
| 🔴 **Desviaciones declaradas (3)** | **(1)** CV-D8 midió «90 turnos = 9,4 KB»; hoy son **12,0 KB** (la conversación siguió creciendo). El orden de magnitud y la conclusión —scan en memoria, sin FTS5— **no cambian** (§1 BR-CV-8). **(2)** El mockup pinta el error de lectura con `--warn` sobre `--card` (`:887`), que **rompe el gate a11y** — deuda ya abierta en `BACKLOG.md:59-64` y `:207-211`; gana el gate, ver `design.md` §Contradicciones C-3. **(3)** CV-D14 dice «4 filas → 2»; con nodo en el alcance son **3** por construcción (`mockup:433`), que es lo que el propio mockup dibuja en §2C. La cuenta firmada es del **estado de reposo** |
| 🕳 **Gaps del modelo que este spec construye** | «última interacción» **no existe** hoy (`session.go` no tiene timestamp por turno; lo único temporal es `CerradaEn`, `session.go:123`) · `Conv` se **tira** al desactivar (`session_historial.go:57`, literal `cerrada.Conv = nil`) |
| ⚠️ **Sin cobertura de story hoy** | `chat-dock.tsx` **no tiene story**: el directorio sólo trae `dictado-button.stories.tsx` y `permission-card.stories.tsx`. Todo lo que este paquete construye nace sin baseline visual — ver `plan-storybook.md` §1.3 |
| 📌 **Orden obligatorio de construcción** | **(1)** modelo + migración + borrado CV-D6 (§A, §F) → **(2)** API (§G) → **(3)** UI (§B, §C, §D). El borrado de datos del operador (RF-337) va **antes** de que exista una UI que los liste: si no, se lista lo que se va a borrar |

---

## 0 · Alcance y no-alcance

**Dentro — una sola superficie y un solo modelo:**

| # | Qué | Estado hoy |
|---|---|---|
| M | **`Conversacion` como entidad**, colgando de `session_id` | inexistente: `Session` **es** la conversación (`internal/domain/session.go:68`) |
| S | **El panel de conversación** (`chat-dock`) gana lista + buscador + `＋ nueva` + retomar, y baja el cromo a 2 filas | `chat-dock.tsx:12-42` no importa ningún store de conversaciones |
| R | El picker del rail **pierde** el bloque `ConversacionesDelArnes` (mudanza declarada) | `new-session-picker.tsx:235` lo renderiza, `:263-307` lo define |
| D | **Los datos en disco** migran; las 3 cerradas de hoy se eliminan (CV-D6) | `~/.arnesia/sessions.json` (5 sesiones) + `~/.arnesia/sesiones-cerradas.json` (3) |

**Fuera de alcance (explícito):**

- **La vista `Hist` del view-strip.** CV-D1: es el historial del ARNÉS. No se toca — ni su stub
  «próximamente» ni el `VIEWS` de `shared/api/types.ts:130-136`.
- **Conversaciones de otras sesiones o de otros arneses.** CV-D4: el alcance del listado y del
  buscador es la sesión activa. Un buscador global es otro paquete.
- **N conversaciones corriendo en paralelo dentro de una sesión.** CV-D7: una activa. El paralelismo
  real sigue siendo N sesiones en el rail (`session-rail.tsx:130-142`).
- **FTS5 / índice nuevo / tocar `~/.arnesia/index.db`.** CV-D8: scan en memoria.
- **Reescribir el corpus JSONL nativo.** Sigue siendo la SSoT del contenido
  (`boundaries/indice-desechable-jsonl-es-verdad.md`); `Conv` es copia liviana de presentación.
- **Botón de cerrar una conversación.** CV-D7: el único cierre es implícito. La ausencia es
  requisito, no omisión (RF-309).
- **El rebrand del token `--warn`.** La deuda a11y es preexistente (`BACKLOG.md:59-64`); este paquete
  **no la arrastra a su superficie nueva** (RF-353), pero tampoco la arregla.

---

## 1 · Reglas de negocio (BR-CV)

| id | regla | por qué |
|---|---|---|
| **BR-CV-1** | Una **Sesión** tiene **N Conversaciones** y **exactamente una activa**, siempre. Nunca cero | CV-D3 + CV-D7. Con cero activas el composer no tiene a quién hablarle (`chat-dock.tsx:268-375` manda a `sendTurn`, que resuelve por `activeId`, `sessions-store.ts:178`) |
| **BR-CV-2** | La conversación cuelga de un **`session_id`**, jamás de un string de arnés reconstruido | CV-D5. El bug que originó el paquete (cerradas por id pelado `vitalia` vs consulta por clave calificada `sin-home~vitalia~vitalia`) **nace de reconstruir la llave**; con `session_id` no es expresable |
| **BR-CV-3** | Vocabulario **activa / inactiva**. La palabra «cerrada» **no aparece** en ninguna superficie nueva | CV-D12. Si se retoma, no estaba cerrada |
| **BR-CV-4** | Crear y retomar son **la misma transición**: desactivan la que estaba y activan otra. No hay estado intermedio | CV-D7 + CV-D11 · `mockup:688-691` |
| **BR-CV-5** | La rotación por contexto **no crea ni corta una conversación**. Es una marca inline | CV-D10. La dispara el `%` de contexto (`session_service.go:505-507`), no un tema nuevo; cortar ahí partiría un tema en dos entradas |
| **BR-CV-6** | Toda **transición desactiva-activa exige turno quieto**: `status` ∉ {`streaming`, `await`} | Mismo motivo que `un-turno-a-la-vez` (`boundaries/sesion-viva-consistente.md`): el conductor viejo tiene un turno abierto sobre su stdin y un permiso pendiente sin dueño |
| **BR-CV-7** | Al desactivar, **`Conv` se persiste**. Al retomar, se repinta desde `Conv` **antes** de reenganchar el stream | CV-D8 + CV-D11. Sin esto no hay búsqueda por texto ni repintado — es el mismo rol que `Conv` ya cumple al cambiar de sesión en el rail (`session.go:59-61`) |
| **BR-CV-8** | El buscador **entra al texto** del transcript (`Conv`), no sólo al título, y **muestra el fragmento** que coincidió | CV-D8 · `mockup:590`. Sin snippet el operador no sabe por qué apareció esa fila |
| **BR-CV-9** | **`0` no se esconde.** 0 turnos, 0 % de ctx y «sin sesión CC» son datos, no vacíos | `mockup:544` (`0 turnos · ctx 0 %`), `:667` («sin turnos todavía»), `:764-765` |
| **BR-CV-10** | Un fallo de lectura **nunca** degrada a «0 conversaciones»: dice el motivo y ofrece Reintentar | `mockup:887-890` · honestidad del repo (CLAUDE.md) — un 0 fabricado es peor que un error |
| **BR-CV-11** | **Superset estricto**: nada firmado del dock se quita. Lo que se mueve se declara como mudanza con origen y destino | `mockups/INDEX.md` regla dura 3 · `mockup:806-846` (§5) |
| **BR-CV-12** | El **borrado de datos del operador** (CV-D6) se ejecuta una sola vez, con el daemon detenido, con copia previa fechada, y **dice qué borró** | CV-D6. Es dato del operador, no basura del build |
| **BR-CV-13** | La migración de disco es **idempotente y sin pérdida**: correrla dos veces da el mismo resultado, y un archivo que no se puede leer **se preserva**, no se pisa | `store.Registry.Save` ya es atómico (temp + `rename`, `registry.go:65-97`); la migración hereda esa disciplina |
| **BR-CV-14** | Ninguna cifra de la lista se **inventa**: «última interacción» sin turnos dice «sin turnos todavía», no una fecha | CV-D13 · `mockup:667` · callout `mockup:936-943` |

---

## A · Modelo y vocabulario (7 RF)

### RF-300 — `Conversacion` es una entidad propia; la sesión deja de ser el hilo
`internal/domain/session.go:68-128` (lo que se parte) · CV-D3.

Los campos que hoy viven en `Session` y **bajan** a `Conversacion`: `ClaudeSessionID` (`:96`),
`Model` (`:97`), `CtxPct` (`:99`), `CtxHist` (`:103`), `RotacionPendiente` (`:107`), `CadenaCC`
(`:111`), `Checkpoint` (`:115`), `Conv` (`:127`), `Turnos` (`:124`).
Los que **quedan** en `Session`: `ID` (`:71`), `Frente` (`:75`), `Arnes` (`:78`), `Empresa`/`Puesto`
(`:79-80`), `Salud` (`:81`), `Status` (`:85`), `View` (`:87`), `Parked` (`:88`), `Reparacion`
(`:92`), `Cwd` (`:119`).

> **`Cwd` se queda en la Sesión.** Es el confinamiento del arnés, no del hilo: lo resuelve
> `spawnLocked` desde `s.resolver.Resolve(r.meta.Arnes)` (`session_service.go:386`), que es por
> arnés. Dos conversaciones de una misma sesión corren en el mismo `cwd` por construcción. La
> conversación lo **lee** para el detalle (RF-330) y para el join del corpus.

```gherkin
Escenario: una sesión persiste con sus conversaciones
  Dado una sesión con 3 conversaciones, una activa
  Cuando el daemon se reinicia
  Entonces la sesión vuelve con sus 3 conversaciones
  Y la misma sigue marcada activa
  Y el ClaudeSessionID de cada una es el suyo, no el de la sesión
```

**Criterios de aceptación**
1. `GET /api/sessions/{id}/conversaciones` devuelve N entradas con `id`, `titulo`, `activa`,
   `turnos`, `ctx_pct`, `ultima_interaccion`.
2. El JSON de una `Session` **ya no** trae `claude_session_id`, `ctx_pct`, `conv`, `cadena_cc`,
   `checkpoint`, `ctx_hist`, `rotacion_pendiente`.
3. `Session.Cwd` sigue presente y con el mismo valor que hoy.

### RF-301 — Toda sesión tiene siempre al menos una conversación, y exactamente una activa
CV-D3 · CV-D7 · BR-CV-1. Sin superficie propia: se ve por consecuencia en `mockup:663-670` (§3D).

```gherkin
Escenario: sesión recién creada
  Cuando se crea una sesión desde «＋ Nueva sesión» del rail
  Entonces nace con UNA conversación activa, de 0 turnos, titulada "nueva conversación"

Escenario: invariante tras cualquier transición
  Dado una sesión con N conversaciones
  Cuando se crea, se retoma o se rota
  Entonces sigue habiendo exactamente 1 activa y N o N+1 en total
```

**Criterios de aceptación**
1. `POST /api/sessions` responde `201` con la sesión **y** su conversación inicial.
2. No existe ninguna operación de la API que deje una sesión con 0 conversaciones ni con 2 activas
   (test de invariante en el dominio, no en el handler).
3. El título de la conversación inicial es `"nueva conversación"` — literal de `mockup:653`, `:666`,
   `:741`. **No** hereda el `"nuevo frente"` de la sesión (`session_service.go:261`).

### RF-302 — Vocabulario: activa / inactiva. «Cerrada» desaparece del producto 🎨
`mockup:524` (el único rótulo: `activa`) · `:928` (traza) · CV-D12.

Las inactivas **no llevan rótulo**: la ausencia de `activa` ya las nombra (`mockup:526-546`).

```gherkin
Escenario: la lista no dice "cerrada"
  Dado una sesión con 1 activa y 3 inactivas
  Cuando se abre la lista
  Entonces exactamente una fila muestra el texto "activa"
  Y ninguna fila muestra "cerrada", "abierta" ni "archivada"
```

**Criterios de aceptación**
1. `grep -ri "cerrada" web/src/widgets/chat-dock/` no encuentra copy visible.
2. El campo del wire se llama `activa: boolean`, no `cerrada`.

### RF-303 — Título auto-derivado del primer mensaje del usuario, editable, y una vez editado no se re-deriva 🎨
`mockup:466` (input de edición) · `:349` (título en reposo) · CV-D9 ·
código análogo `session_service.go:352-354` + `deriveFrente` (`:915-926`).

```gherkin
Escenario: derivación
  Dado una conversación de 0 turnos titulada "nueva conversación"
  Cuando el operador manda su primer turno "no veo nada en el mapa, ¿podés repararlo?"
  Entonces el título pasa a derivarse de ese texto

Escenario: no se re-deriva
  Dado una conversación cuyo título el operador editó a "el bug del índice"
  Cuando llegan 20 turnos más
  Entonces el título sigue siendo "el bug del índice"
```

**Criterios de aceptación**
1. La conversación lleva una marca de «título editado por el operador»; con la marca puesta, ningún
   turno vuelve a tocar el título.
2. La derivación reusa `deriveFrente` (`session_service.go:915-926`) — **no** se escribe un
   derivador nuevo.
3. El `Frente` de la **sesión** conserva su propia derivación y su propio renombrado
   (`session-rail.tsx:256-273`): son dos nombres distintos y ambos siguen existiendo.

### RF-304 — «Última interacción»: timestamp estampado en cada turno
CV-D13. **Campo nuevo — hoy no existe.** Verificado: `session.go` no tiene ningún timestamp por
turno; lo único temporal es `CerradaEn` (`:123`), que se estampa **al cerrar**
(`session_historial.go:59`) y por eso mentiría la fecha del último mensaje.

```gherkin
Escenario: se estampa al turno, no al desactivar
  Dado una conversación cuyo último turno fue el 24 de julio
  Cuando el operador la desactiva el 26 de julio
  Entonces su "última interacción" sigue diciendo 24 de julio

Escenario: sin turnos, sin fecha
  Dado una conversación de 0 turnos
  Cuando se lista
  Entonces la celda dice "sin turnos todavía", no una fecha
```

**Criterios de aceptación**
1. `Conversacion.UltimaInteraccion` (RFC3339, UTC) se escribe en el **mismo** punto donde hoy se
   appendea el turno del usuario (`session_service.go:355`) y donde se appendea el del assistant
   (en `consume`, rama `result`, `session_service.go:497-518`).
2. Con `Turnos == 0` el campo va **vacío** (`omitempty`) y la UI pinta el literal de `mockup:667`.
3. El valor viaja en el wire y el FE **no lo calcula**: formatea lo que recibe.

### RF-305 — `Conv` se persiste al desactivar
CV-D8 + CV-D11 · BR-CV-7. **Cambio de comportamiento verificado:** hoy `archivarLocked` hace
literal `cerrada.Conv = nil` (`internal/usecase/session_historial.go:57`) y
`cerrada.Checkpoint = ""` (`:58`), y guarda sólo `Turnos = len(cerrada.Conv)` (`:56`). Las 3
entradas de `~/.arnesia/sesiones-cerradas.json` lo confirman: ninguna trae `conv`.

```gherkin
Escenario: el transcript sobrevive a la desactivación
  Dado una conversación activa con 41 turnos
  Cuando el operador crea otra con ＋
  Entonces la primera queda inactiva con sus 41 turnos EN DISCO
  Y su Checkpoint y su CadenaCC se conservan
```

**Criterios de aceptación**
1. Tras desactivar, el registro en disco contiene `conv` con los N turnos y `checkpoint` no vacío si
   lo había.
2. `Turnos` sigue existiendo y es `len(Conv)` — es lo que la fila muestra sin cargar el transcript.
3. La JSONL nativa **no se toca**: `boundaries/conductor-no-parsea-jsonl.md` y
   `indice-desechable-jsonl-es-verdad.md` siguen enforced.

### RF-306 — El registro aparte de cerradas desaparece
`mockup:840-844` (§5B: «desaparece … el registro aparte `sesiones-cerradas.json`») · CV-D12.

`SetArchivoCerradas` (`internal/usecase/session_historial.go:23-27`, cableado en
`cmd/arnesia/main.go:325-329`, con la ruta que arma `cerradasPathDefault`,
`cmd/arnesia/main.go:796-805`) modelaba un archivo **terminal**. Bajo CV-D12 no hay archivo
terminal: las conversaciones son pares dentro de su sesión.

**Decisión de este spec** (CV-D12 la dejó explícitamente abierta: *«Se decide en spec si el archivo
desaparece o queda como registro de sesiones enteras cerradas»*):

> **El archivo desaparece como registro de CONVERSACIONES y sobrevive como registro de SESIONES
> enteras cerradas**, renombrado a `~/.arnesia/sesiones-cerradas.json` → **`~/.arnesia/sesiones-archivadas.json`**.
> Motivo: cerrar una **sesión** sí es terminal (`session_service.go:299-324` la saca del registro y
> del rail), y su metadata sigue siendo el join que el historial B2 necesita
> (`session_historial.go:97-123`). Lo que deja de existir es «conversación cerrada» como categoría.

**Criterios de aceptación**
1. `Cerradas()` (`session_historial.go:71-92`) y `HistorialCerrada()` (`:97-123`) se conservan pero
   operan sobre **sesiones archivadas**, no sobre conversaciones.
2. Ninguna superficie del FE consulta ese archivo (RF-333).
3. Si el archivo nuevo no se puede crear, el daemon degrada honesto con `warn` — igual que hoy
   (`main.go:325-327`), jamás bloquea el arranque.

---

## B · Ciclo de vida de una conversación (10 RF)

### RF-307 — `＋ nueva` crea una conversación en la sesión activa 🎨
`mockup:355` (`＋`, `title="Nueva conversación — desactiva la actual"`) · `:733-767` (§4B) · CV-D2.

```gherkin
Escenario: crear
  Dado una conversación activa con 90 turnos, ctx 68 %
  Cuando el operador aprieta ＋
  Entonces nace una conversación activa titulada "nueva conversación", 0 turnos, ctx 0 %
  Y el transcript queda vacío con el mensaje que nombra la que se desactivó
  Y el chip de ctx muestra 0 %, NO se esconde
```

**Criterios de aceptación**
1. El vacío dice, literal de `mockup:749-752`: «Pídele un cambio a **vitalia**. «por qué el mapa
   sale vacío» quedó guardada — la retomás desde ▶». **Nombra** la que se desactivó.
2. El vacío vigente («Esta conversación ES la sesión Claude Code del frente …»,
   `chat-dock.tsx:110-115`) **se reemplaza**: bajo CV-D3 dejó de ser verdad. Es la única
   eliminación de copy firmado del paquete, y su autorización es CV-D3.
3. El chip de ctx a 0 % se pinta igual (`mockup:742`); el `◍ sin sesión CC` vigente
   (`chat-dock.tsx:82`) sobrevive **dentro del detalle** (`mockup:764-765`).

### RF-308 — Crear desactiva la anterior, en una transición
CV-D7 · BR-CV-4.

**Criterios de aceptación**
1. El backend expone **una** operación (`POST …/conversaciones`) que desactiva y activa; el FE no
   encadena dos llamadas.
2. Un fallo deja el estado **anterior intacto**: no queda una desactivada sin reemplazo.
3. Al desactivar se aplica RF-305 (se persiste `Conv`) y RF-311 (permisos pendientes).

### RF-309 — No hay botón de cerrar 🎨
`mockup:369-372` (§2A note: «**No hay botón de cerrar** — a propósito») · CV-D7.

**Criterios de aceptación**
1. Ni la fila de conversación (`mockup:348-357`) ni las filas de la lista (`mockup:518-546`) exponen
   un control de cierre, ni con hover.
2. Contrasta con el rail, que **sí** tiene `✕ Cerrar sesión` (`session-rail.tsx:241-253`) y lo
   conserva: cerrar la **sesión** entera sigue existiendo (RF-350).

### RF-310 — Seleccionar una conversación inactiva la RETOMA 🎨
`mockup:694-731` (§4A) · `:526-546` (filas seleccionables) · CV-D11.

```gherkin
Escenario: retomar
  Dado la lista abierta con "sellar el arnés con arnes.l0.json" inactiva
  Cuando el operador la selecciona
  Entonces esa pasa a activa y la anterior a inactiva
  Y el transcript se repinta desde su Conv persistido
  Y aparece la franja "Retomando la conversación… --resume d303a93f"
  Y el detalle de identidad se abre solo
```

**Criterios de aceptación**
1. La franja es **efímera**: desaparece al primer frame del conductor nuevo, o al fallar (RF-348).
   Literal de `mockup:709-712`.
2. El repintado ocurre **antes** de enganchar el stream (BR-CV-7).
3. `CadenaCC`, `Checkpoint` y `Cwd` viajan intactos al spawn — es el mismo `spawnLocked`
   (`session_service.go:385-447`), con `Resume = ClaudeSessionID` de **esa** conversación (`:390`).
4. **No es solo lectura**: seleccionar activa. No existe un modo «ver sin retomar».

### RF-311 — Al desactivar, los permisos pendientes se deniegan con motivo
Sin superficie propia (el rastro cae en el transcript, que ya existe:
`sessions-store.ts:405-428`). Resuelve el punto que el mockup dejó abierto
(`mockup:954-959`: *«qué pasa con permisos pendientes de la conversación que se desactiva»*).

**Precedente que se reusa, no se inventa:** `Interrupt` ya deniega los asks pendientes con motivo
antes de interrumpir (`session_service.go:848-852`, doc-comment: *«pending permission asks are
denied first (motivo "interrumpido por el operador")»*).

```gherkin
Escenario: permiso pendiente al desactivar
  Dado una conversación en `await` con una tarjeta de permiso abierta
  Cuando el operador la desactiva
  Entonces la tarjeta se resuelve como deny con motivo "conversación desactivada por el operador"
  Y el rastro ✕ queda en el transcript de ESA conversación
  Y el conductor no queda bloqueado esperando una respuesta que nadie va a dar
```

**Criterios de aceptación**
1. El motivo es distinguible del de `Interrupt`.
2. `pendingPerms[id]` del store se vacía para esa conversación (`sessions-store.ts:378`, mismo
   mecanismo que la rama `error`).
3. **Nota:** por BR-CV-6 este escenario sólo ocurre si RF-312 lo permite; con la regla estricta de
   RF-312 el caso se vuelve inalcanzable desde la UI y queda como guardrail del dominio.

### RF-312 — Con un turno en vuelo, `＋` y las filas de la lista quedan deshabilitados **con motivo** 🎨
Superficie nueva (superset del mockup, que dibuja `＋` siempre habilitado) · BR-CV-6 ·
precedente `chat-dock.tsx:293` (`busy = streaming || await`).

Resuelve dos escenarios que el mockup no dibuja pero el encargo exige: crear/retomar con turno en
vuelo y con permiso pendiente.

```gherkin
Escenario: crear con el turno en vuelo
  Dado la conversación activa en `streaming`
  Entonces ＋ está deshabilitado con title "esperá a que termine el turno (■ para interrumpir)"

Escenario: retomar con un permiso pendiente
  Dado la conversación activa en `await`
  Cuando se abre la lista
  Entonces las filas inactivas están deshabilitadas con title "esperá tu decisión de permiso"
  Y la fila activa sigue visible y legible
```

**Criterios de aceptación**
1. El deshabilitado **siempre** trae `title` con el motivo — nunca un control apagado y mudo
   (patrón ya sancionado: `new-session-picker.tsx:363`).
2. El buscador **sigue funcionando** con el turno en vuelo: buscar no cambia estado.
3. El servidor **también** rechaza (`409`), no sólo el cliente: mismo patrón que `Turn`/`ErrBusy`
   (`session_service.go:339-342` → 409, `boundaries/sesion-viva-consistente.md` check
   `un-turno-a-la-vez`).

### RF-313 — La rotación por contexto es invisible: misma conversación, marca inline 🎨
`mockup:769-802` (§4C) · `:785` (`⟳ contexto rotado · checkpoint`) · CV-D10.

Lo vigente que se conserva tal cual: el disparo por umbral (`session_service.go:505-507`),
`rotarLocked` (`session_rotacion.go:52-64`), el encadenado de `CadenaCC` (`:58`), el `Checkpoint`
mecánico (`:60`) y el breadcrumb `RolSys` (`:63`, constante `breadcrumbRotacion` = `"— contexto
rotado, seguimos —"`, `session_rotacion.go:12`).

```gherkin
Escenario: rotación con la lista abierta
  Dado la lista abierta con 4 conversaciones
  Cuando el próximo turno dispara la rotación
  Entonces la lista sigue mostrando 4, no 5
  Y el título de la activa no cambia
  Y el ctx de su fila baja al valor nuevo

Escenario: la marca es un respiro del hilo, no una burbuja
  Cuando se pinta el transcript rotado
  Entonces la marca va centrada, punteada, sin cola de burbuja
  Y el detalle de identidad se abre solo mostrando el cc-id NUEVO
```

**Criterios de aceptación**
1. **Cero** entradas nuevas en la lista tras rotar.
2. El texto de la marca es el que ya se persiste (`session_rotacion.go:12`); el dibujo `mockup:785`
   dice «contexto rotado · checkpoint» — **gana el código**, porque ese string ya está en `Conv` de
   conversaciones vivas y reescribirlo dejaría transcripts viejos con la marca vieja. Ver
   `design.md` §Contradicciones C-5.
3. `CadenaCC` de la conversación acumula el `ClaudeSessionID` anterior — el join del historial sigue
   funcionando (`session_historial.go:112-118`).
4. 🔴 **La marca tiene que llegar EN VIVO, y hoy no llega.** Verificado: `session_rotacion.go` no
   tiene **ni una** llamada a `s.publish(...)` — las 14 del usecase están todas en
   `session_service.go` (`:376`, `:468`, `:478`, `:518`, `:538`, `:561`, `:576`, `:617`, `:626`,
   `:656`, `:680`, `:687`, `:812`, `:879`). El breadcrumb sólo se persiste en el `Conv` del backend,
   y el `conv` del FE es un espejo propio armado a partir de frames
   (`sessions-store.ts:80-87 appendConv`): **el operador no vería la marca hasta recargar**. Este RF
   exige emitir un frame en `rotarLocked` — el `kind` reusable es `message`/`act`… ninguno encaja
   (son de turno). Va con el frame nuevo de H-2. Ver §N H-8.

### RF-314 — El detalle de identidad se abre solo al retomar y al rotar 🎨
`mockup:703` + `:713-715` (§4A, `ctxchip open` + detalle) · `:774` + `:780-782` (§4C) · CV-D14.

**Criterios de aceptación**
1. Son los **dos únicos** momentos en que se abre sin click: son los dos en que cambia el `cc-id`.
2. Se cierra con el mismo click que lo abre; su estado no persiste entre conversaciones.

### RF-315 — Rotación y creación funcionan con el dock colapsado
Sin superficie (el dock colapsado no dibuja nada: `shell-page.tsx:53`, `width: 0`).

```gherkin
Escenario: rotación con el dock colapsado
  Dado el dock colapsado (⌘K)
  Cuando un turno dispara la rotación
  Entonces la rotación ocurre igual y la marca queda en Conv
  Y al reabrir el dock la marca está en su lugar cronológico
  Y NO se abre el dock solo ni se muestra ningún toast
```

**Criterios de aceptación**
1. El estado `chatOpen` (`sessions-store.ts:92`) no lo toca ningún camino de este paquete salvo
   `create` de sesión, que ya lo abre hoy (`:146`).
2. La lista y el detalle se **desmontan** al colapsar; al reabrir arrancan cerrados.

### RF-316 — Retomar la conversación que ya está activa es un no-op
Sin superficie.

**Criterios de aceptación**
1. Click en la fila `activa` (`mockup:518-525`) cierra la lista y no dispara ninguna llamada.
2. No se re-spawnea el conductor ni se repinta el transcript.

---

## C · La lista y el buscador (9 RF)

### RF-317 — La lista se abre EN SITIO y toma el área del transcript 🎨
`mockup:489-493` (lead §3) · `:511-549` (§3A) · CV-D2.

Idioma reusado, no inventado: es el gesto que `NewSessionPicker` ya usa en el rail
(`session-rail.tsx:101-110` la renderiza dentro del propio `<aside>`; TS-D6/D10 «100 % inline, cero
popover/backdrop»).

**Criterios de aceptación**
1. El header del dock, la fila de conversación y el composer **siguen visibles** con la lista
   abierta (`mockup:499-510` + `:550-554`).
2. Cero `<dialog>`, cero backdrop, cero portal.
3. Cierra con `Cancelar` (`mockup:548`), con `Escape` y con el mismo chevron que la abrió
   (`mockup:504`).
4. **Ancho** (punto que el mockup dejó abierto, `mockup:954-959`): la lista ocupa el ancho del dock,
   sea cual sea. El dock ya es redimensionable entre 300 px y el 60 % de la ventana
   (`shell-page.tsx:16-18`) y la lista no introduce mínimo propio: las filas truncan
   (`mockup:222-223`, `text-overflow: ellipsis`).

### RF-318 — El alcance se DICE, no se deduce 🎨
`mockup:516` (`4 conversaciones de la sesión **repro del bug de carga**`) · `:583` (`2 de 4
coinciden`) · `:661` · CV-D4.

**Criterios de aceptación**
1. El rótulo nombra la **sesión** por su `Frente`, no por el arnés.
2. Con filtro activo el rótulo cambia a `N de M coinciden` (`mockup:583`).
3. La lista **jamás** trae una conversación de otra sesión (test de dominio, no de UI).

### RF-319 — Cada fila: título · última interacción · nº de turnos · ctx final 🎨
`mockup:518-546` · `:522` (`hace 4 min · 90 turnos · ctx 68 %`) · `:667` · CV-D13.

**Criterios de aceptación**
1. Cuatro datos y ninguno más — la fila no es una tarjeta (CV-D13).
2. Formato de fecha relativo cerca / absoluto lejos, tal como el dibujo: `hace 4 min` · `ayer 18:02`
   · `24 jul` · `23 jul` (`mockup:522`, `:530`, `:537`, `:544`).
3. Sin turnos ⇒ `sin turnos todavía · ctx 0 %` (`mockup:667`), BR-CV-14.

### RF-320 — Orden: última interacción descendente, con la activa siempre visible 🎨
`mockup:954-959` (el propio dibujo declara este orden y lo manda a spec) · `:518-546` (el orden
dibujado).

**Criterios de aceptación**
1. Orden primario: `ultima_interaccion` descendente.
2. Las de 0 turnos (sin fecha) van **al final**, ordenadas por creación descendente — no se les
   inventa una fecha para ordenarlas (BR-CV-14).
3. La activa **no se fija arriba**: se ordena como cualquier otra (en el dibujo queda primera porque
   es la más reciente). Lo que se garantiza es que **está en la lista**, incluso filtrada (RF-323).

### RF-321 — El buscador entra al TEXTO del transcript 🎨
`mockup:578-582` (input) · `:590`, `:599` (snippets) · `:611-615` (note) · CV-D8 · BR-CV-8.

**Criterios de aceptación**
1. Busca en `titulo` **y** en el texto de cada `Turn` de `Conv`.
2. Sin FTS5, sin índice, sin tocar `~/.arnesia/index.db`. Medido hoy: la conversación más larga en
   disco son **12 281 B**; 100 conversaciones ≈ 1,2 MB — un scan lineal en memoria.
3. **Incremental**, sin Enter (punto que el mockup dejó abierto, `mockup:954-959`). Motivo: es el
   comportamiento que el repo ya sancionó para el otro buscador de la misma familia — `coincide()`
   de `new-session-picker.tsx:69-74` corre en cada `onChange` (`:192`), sin debounce, sobre una
   lista del mismo orden de magnitud.
4. La comparación es **case- y accent-insensitive** (RF-324).

### RF-322 — Cada coincidencia muestra su fragmento 🎨
`mockup:590` (`…el <mark>manifiesto</mark> declara 0 elementos…`) · `:599` · CV-D8.

**Criterios de aceptación**
1. Un fragmento por fila: el **primero** que coincide, con elipsis a ambos lados.
2. El término coincidente va resaltado (`<mark>`), no en negrita ni en color solo.
3. Si la coincidencia es sólo en el **título**, no se dibuja fragmento (no hay nada que explicar).

### RF-323 — La conversación activa participa de la búsqueda como cualquier otra 🎨
`mockup:585-593` (la fila `activa` aparece entre los resultados) · `:614` (note).

**Criterios de aceptación**
1. Si la activa no coincide, **no aparece** — y el rótulo `N de M` lo refleja.
2. Si aparece, conserva su tinte, su radio marcado y su rótulo `activa`.

### RF-324 — Búsqueda insensible a mayúsculas y acentos
Sin superficie (es comportamiento del matcher).

```gherkin
Escenario: acentos
  Dado una conversación cuyo transcript dice "el manifiesto está vacío"
  Cuando el operador busca "VACIO"
  Entonces esa conversación coincide
```

**Criterios de aceptación**
1. Normalización NFD + strip de diacríticos en ambos lados, más `toLowerCase()`.
2. Es **superset** de `coincide()` (`new-session-picker.tsx:69-74`), que hoy sólo hace
   `toLowerCase()` — el buscador nuevo no puede ser peor que el viejo.

### RF-325 — Los dos vacíos del buscador dicen dónde buscaron 🎨
`mockup:634-638` (§3C, sin coincidencias) · `:660-676` (§3D, una sola conversación).

**Criterios de aceptación**
1. Sin coincidencias: «Ninguna conversación de esta sesión menciona **«telemetría»**» + «Se buscó en
   el título y en el texto de las 4.» + `Limpiar búsqueda` (literales de `mockup:635-637`).
2. Con **una sola** conversación el buscador **no se dibuja** (`mockup:681`: «nada que filtrar») y
   la lista muestra el mensaje de `mockup:672-675`.
3. Ninguno de los dos usa el copy «sin resultados» pelado.

---

## D · El cromo comprimido (7 RF)

### RF-326 — El dock baja de 4 filas de cromo a 2 🎨
`mockup:286-327` (§1, el antes: 4 filas) · `:341-377` (§2A, el después: 2) · `:332-338` (lead) ·
CV-D14.

| Fila | Qué lleva | Cuándo | Código vigente que absorbe |
|---|---|---|---|
| 1 | pip · frente de la sesión · colapsar | siempre, **intacta** | `chat-dock.tsx:21-34` |
| 2 | ▶ título · chip de ctx · 🔍 · ＋ | siempre | nueva; absorbe el ctx de `SessionLine` (`:78-93`) |
| 3 | chip del nodo en alcance, con ✕ | **sólo con nodo elegido** | `ScopeRow` (`:46-76`), condicionada |

**Criterios de aceptación**
1. En reposo y sin nodo: **exactamente 2 filas** entre el borde superior del dock y el primer
   mensaje.
2. La fila 1 no cambia ni un píxel salvo el glifo (RF-332).
3. `SessionLine` (`chat-dock.tsx:78-93`) **desaparece como fila fija**; sus 4 datos sobreviven
   (RF-327, RF-328) — es una mudanza, no un borrado (BR-CV-11).

### RF-327 — El ctx es un chip-disclosure 🎨
`mockup:350-352` (cerrado) · `:388-390` (abierto) · CV-D14.

**Criterios de aceptación**
1. Conserva la barra de 44×5 y el `N%` de la `SessionLine` vigente
   (`chat-dock.tsx:87-90` → `mockup:351`).
2. Es la **única cifra accionable** que queda a la vista: dispara la rotación (RF-195,
   `session_service.go:505-507`).
3. A 0 % se pinta igual, no se esconde (`mockup:742`, BR-CV-9).

### RF-328 — El detalle lleva cc-id · arnés · modelo · **cwd** 🎨
`mockup:396-399` · `:410-415` (note) · CV-D14.

**Criterios de aceptación**
1. Los 4 datos de la `SessionLine` vigente están **todos** (`chat-dock.tsx:81-90`): `◍ <cc-id[0:8]>`
   / `sin sesión CC`, arnés, modelo, ctx.
2. **Gana** el `cwd` (`mockup:398`), que hoy no se ve en ninguna superficie y es el confinamiento
   real del conductor (`session_service.go:386`, `:442`).
3. Nada se pierde: es un superset estricto de la fila que se retira.

### RF-329 — `ScopeRow` deja de ser fila fija 🎨
`mockup:433-439` (§2C, la fila con nodo) · `:450-455` (note) · CV-D14 · RF-110/RF-111.

**Criterios de aceptación**
1. Sin nodo: la fila **no existe** en el DOM. Su estado vacío gastaba una fila entera en el hint
   «selecciona un nodo en el Mapa para acotar» (`chat-dock.tsx:72`) + un chip punteado del arnés
   (`:52-54`), redundante con el rail, el topbar y el placeholder del composer.
2. Con nodo: aparece el chip **literal** del vigente — cuadrito de clase, `clase`, `nodeId`,
   `fuentePath` truncado y el **✕ para quitarlo** (`chat-dock.tsx:55-70` → `mockup:434-438`).
3. El `✕` sigue llamando a `setScope(null)` (`chat-dock.tsx:64`); el comportamiento de RF-111 no
   cambia.
4. **Lo que se quita**, declarado: el rótulo «Alcance:» (`chat-dock.tsx:51`), el chip punteado del
   arnés (`:52-54`) y el hint del estado vacío (`:72`). Autoriza CV-D14.

### RF-330 — El detalle se abre y se cierra con un clic 🎨
`mockup:184-189` (estilos `.ctxchip`/`.open`) · `:388` (`ctxchip open`).

**Criterios de aceptación**
1. `aria-expanded` refleja el estado; el detalle es el `aria-controls`.
2. El chip abierto se distingue por fondo + borde, no sólo por color (`mockup:186`).

### RF-331 — Renombrar en línea, con el mismo gesto que el rail 🎨
`mockup:458-484` (§2D) · `:478-483` (note) · CV-D9.

**Criterios de aceptación**
1. `✎ → input → Enter confirma, Escape descarta` — calca `session-rail.tsx:256-273`.
2. Mientras se edita, el input **toma la fila entera** (`mockup:466`, `.titlein` con `width: 100%`
   en `:202`): el ctx y las acciones no compiten con el cursor.
3. Título vacío ⇒ **descarta**, no borra (RF-341). Es lo que ya hace el rail
   (`session-rail.tsx:199-200`: `if (value.trim() && …) rename(); else setValue(s.frente)`) y lo que
   ya hace el store (`sessions-store.ts:160-161`: `if (!trimmed) return`).

### RF-332 — El glifo de colapsar pasa de `⟩` a `»` 🎨
`mockup:346` (`» colapsar`) · `:299` (§1 conserva el `⟩` vigente) · CV-D15.

**Criterios de aceptación**
1. `chat-dock.tsx:32` pasa de `⟩ colapsar` a `» colapsar`.
2. La palabra «colapsar» y el `title="Colapsar el dock (⌘K para reabrir)"` (`chat-dock.tsx:29`) se
   conservan literales.
3. Es el glifo que el rail ya usa para lo mismo (`session-rail.tsx:96`: `{collapsed ? "»" : "«"}`).

---

## E · Lo que se mueve (2 RF)

### RF-333 — `ConversacionesDelArnes` sale del picker del rail 🎨
`mockup:806-846` (§5, la mudanza declarada) · `new-session-picker.tsx:235` (el render) ·
`:263-307` (el componente) · CV-D2.

**Es mudanza, no borrado** (BR-CV-11 · `mockups/INDEX.md` regla 3). Destino declarado:

| Qué hacía en el picker | Dónde vive ahora |
|---|---|
| listar conversaciones con turnos y fecha | §3A — la lista (RF-317…RF-320) |
| expandir una cerrada a sus turnos truncados a una línea | §3B — el buscador sobre el texto (RF-321) y el transcript completo al retomar |
| — (no existía) | §4A — **retomar** (RF-310) |

**Criterios de aceptación**
1. `new-session-picker.tsx` pierde el import de `conversaciones-store` (`:15`), la llamada
   `cargar(e.clave)` (`:111`), el render (`:235`) y el componente (`:263-307`).
2. El picker queda **sólo con lo suyo**: elegir el arnés y la copia. Nada más se le quita
   (`mockup:836-838`).
3. Se cierra de paso el bug que originó el paquete — verificado hoy: el picker consulta por clave
   calificada (`new-session-picker.tsx:111`) y el registro guarda id pelado (`vitalia`,
   `dev-full-cycle` en `~/.arnesia/sesiones-cerradas.json`), así que mostraba **0**. No se arregla:
   **se elimina el dato** (CV-D6/RF-337) y CV-D5/BR-CV-2 impide que vuelva.
4. Las 4 stories de `new-session-picker.stories.tsx` que hoy fallan el gate a11y por el `text-warn`
   de este bloque (`BACKLOG.md:207-211`, `new-session-picker.tsx:273`) **dejan de fallar** por
   desaparición de la causa. Es un efecto colateral, no el objetivo.

### RF-334 — El store `conversaciones-store` se muda al dock, reescrito
`web/src/widgets/session-rail/model/conversaciones-store.ts` (64 líneas) → nueva slice del widget
`chat-dock` · CV-D2.

**Criterios de aceptación**
1. Su modelo actual (`arnes` / `vivas` / `cerradas` / `historialDe`, `:8-13`) **no sobrevive**: era
   por arnés y por «cerradas». El nuevo es por `session_id` (BR-CV-2, BR-CV-3).
2. Su test hermano `conversaciones-store.test.ts` se reescribe con él (corre en el project `unit`,
   `vitest.config.ts:20-27`).
3. `depcruise` `no-sibling-widget-imports` (`.dependency-cruiser.js`) prohíbe que `session-rail`
   importe del `chat-dock`: la mudanza es física, no un re-export.

---

## F · Migración de datos en disco (5 RF)

> Contexto verificado: `~/.arnesia/sessions.json` (17 202 B, 5 sesiones) y
> `~/.arnesia/sesiones-cerradas.json` (1 479 B, 3 entradas). Ambos los escribe el mismo
> `store.Registry` (`registry.go:32-41`), con `Save` atómico por temp + `rename`
> (`registry.go:65-97`) y `Load` que trata «archivo ausente» como registro vacío (`:50-52`).

### RF-335 — Migración del formato vigente al formato nuevo, al arrancar el daemon

```gherkin
Escenario: instalación vieja
  Dado un ~/.arnesia/sessions.json del formato actual con 5 sesiones
  Cuando el daemon arranca con el binario nuevo
  Entonces cada sesión queda con exactamente UNA conversación activa
  Y esa conversación hereda claude_session_id, model, ctx_pct, ctx_hist, cadena_cc,
    checkpoint, rotacion_pendiente, conv y turnos de la sesión vieja
  Y la sesión conserva id, frente, arnes, empresa, puesto, salud, status, view, parked,
    reparacion y cwd
  Y el título de la conversación se deriva de su primer turno de rol `user`, o queda
    "nueva conversación" si no hay ninguno
```

**Criterios de aceptación**
1. `ultima_interaccion` de la conversación migrada queda **vacío** — no existe el dato para
   inferirlo, y BR-CV-14 prohíbe inventarlo. La fila dirá «sin fecha» con sus `N turnos` reales.
   Es un hueco declarado, no un pass fabricado.
2. La migración escribe **un archivo nuevo** (`~/.arnesia/sesiones.json` v2) y **deja
   `sessions.json` intacto** como respaldo; el daemon prefiere el nuevo si existe.
3. Es **idempotente** (BR-CV-13): con el archivo v2 presente, la migración no corre.
4. Verificable contra la máquina real: 5 sesiones → 5 sesiones × 1 conversación; la de 90 turnos
   conserva sus 90 (`s6165ac75`); las 3 de 0 turnos conservan 0.

### RF-336 — Archivos corruptos: se preservan, no se pisan

```gherkin
Escenario: sessions.json ilegible
  Dado un ~/.arnesia/sessions.json con JSON inválido
  Cuando el daemon arranca
  Entonces NO se sobreescribe
  Y se renombra a sessions.json.corrupto-<AAMMDDHHMM>
  Y el daemon arranca con registro vacío y lo DICE en el log con la ruta del respaldo
  Y el rail muestra 0 sesiones, no sesiones inventadas
```

**Criterios de aceptación**
1. `Load` hoy devuelve error en JSON inválido (`registry.go:57-59`) y el arranque **no** define qué
   hacer: este RF lo define.
2. El respaldo lleva timestamp — un segundo arranque no pisa el primero.
3. Ninguna semilla ilustrativa (`seedSessions`, `session_service.go:947-951`) se usa para tapar una
   corrupción: sólo aplica a registro **ausente**, que es otra cosa.

### RF-337 — Las 3 conversaciones cerradas de hoy se ELIMINAN (procedimiento seguro)
CV-D6 · BR-CV-12 · `mockup:922` (traza: «sin superficie (dato)»).

Las tres, verificadas hoy en `~/.arnesia/sesiones-cerradas.json`:

| id | frente | arnes (id **pelado**) | cerrada_en | turnos |
|---|---|---|---|---|
| `s1b38a066` | «quiero saber por qué solo veo el readme y el hi…» | `vitalia` | 2026-07-24T01:44:57Z | 12 |
| `s408bb085` | «ciclo full-cycle · spec→released» | `dev-full-cycle` | 2026-07-25T02:05:11Z | 12 |
| `s020210e3` | «nuevo frente» | `vitalia` | 2026-07-25T19:59:34Z | — (sin turnos ni cc-id) |

**Procedimiento — exactamente estos pasos, en este orden:**

1. **Detener el daemon.** `pkill -f 'arnesia serve'` y **confirmar** con `curl -sf
   http://127.0.0.1:4200/api/version` → debe fallar. Con el daemon vivo, `persistLocked`
   (`session_service.go:902`) o `archivarLocked` (`session_historial.go:65`) pueden reescribir el
   archivo entre la copia y el borrado.
2. **Copia previa fechada**, fuera del árbol que se toca:
   `cp ~/.arnesia/sesiones-cerradas.json ~/.arnesia/sesiones-cerradas.json.bak-<AAAAMMDD-HHMM>`.
   Precedente en la misma carpeta: `sessions.json.bak` ya existe (522 B, 2026-07-05).
3. **Verificar la copia**: `python3 -c "import json;print(len(json.load(open(...))))"` → `3`.
4. **Borrar el archivo original** — el archivo entero, no entradas sueltas: las tres se van.
   `rm ~/.arnesia/sesiones-cerradas.json`.
5. **Arrancar el daemon** y confirmar `GET /api/sessions?cerradas=1` → `cerradas: []` (o el 404 del
   endpoint retirado, según RF-345).
6. **Dejar rastro**: una línea en `PARIDAD.md` del paquete con la fecha, los 3 ids y la ruta del
   `.bak`.

**Qué NO se borra**, explícito:
- `~/.arnesia/sessions.json` (las 5 sesiones vivas) — lo migra RF-335.
- Las JSONL nativas de `~/.claude/projects/` — ArnesIA no las escribe ni las borra
  (`boundaries/indice-desechable-jsonl-es-verdad.md`); el contenido real de esas 3 conversaciones
  sigue en disco, fuera de ArnesIA.
- `~/.arnesia/index.db` ni `telemetria.db`.

**Criterios de aceptación**
1. **No se escribe migración de llaves** (CV-D6, literal): el bug muere con los datos.
2. El procedimiento es **manual y del operador**, no un paso automático del arranque: borrar datos
   del operador en silencio al actualizar sería exactamente lo contrario de la honestidad del repo.
3. Si el operador **no** lo corre, el sistema no rompe: RF-345 hace que el endpoint viejo devuelva
   404 y RF-333 quita al único consumidor.

### RF-338 — Sin permisos de escritura: se dice, no se rompe

```gherkin
Escenario: ~/.arnesia no escribible
  Dado ~/.arnesia con permisos de sólo lectura
  Cuando el operador crea una conversación
  Entonces la operación falla con 500 y el motivo real del filesystem
  Y la UI muestra el motivo en la superficie, no en la consola
  Y la conversación NO aparece en la lista como si se hubiera creado
```

**Criterios de aceptación**
1. `Save` ya devuelve el error envuelto con la ruta (`registry.go:70-95`); hoy `persistLocked`
   (`session_service.go:902-912`) lo consume — este RF exige que el fallo **llegue al operador** en
   la operación que lo causó.
2. Estado en memoria y estado en disco no divergen en silencio: si no persistió, no se muestra como
   persistido.

### RF-339 — Crash a mitad de la migración no deja el registro a medias
BR-CV-13.

**Criterios de aceptación**
1. La migración usa el mismo temp+`rename` de `registry.go:77-95`; un crash deja o el archivo viejo
   entero o el nuevo entero, nunca medio archivo.
2. Al reintentar, RF-335 punto 3 (idempotencia) aplica.

---

## G · API (7 RF)

> Hoy las rutas de sesión viven en `internal/adapters/transport/http/router.go:123-131` y sus
> handlers en `sessions.go`. **Hueco declarado:** ni `GET /api/sessions?arnes=&cerradas=1` ni
> `GET /api/sessions/cerradas/{id}/historial` están en `docs/architecture/contracts/api/openapi.yaml`
> (`grep -n "cerradas\|historial"` → vacío). El contrato viene con deuda antes de este paquete.

### RF-340 — `GET /api/sessions/{id}/conversaciones`
Reemplaza el uso que el FE hacía de `conversacionesDeArnes` (`client.ts:218-221`).

**Criterios de aceptación**
1. `200` con `{ conversaciones: [...] }`; cada una: `id`, `titulo`, `titulo_editado`, `activa`,
   `turnos`, `ctx_pct`, `ultima_interaccion`, `claude_session_id`, `model`.
2. **No** devuelve `conv` (el transcript) — la lista no lo necesita y son 12 KB por conversación.
3. `404` si la sesión no existe. **Nunca** `200` con lista vacía para una sesión inexistente
   (BR-CV-10).

### RF-341 — `GET /api/sessions/{id}/conversaciones?q=<texto>` — la búsqueda corre en el daemon

**Decisión, con su motivo:** el filtro vive en el **backend**, no en el FE. RF-340 no manda `conv`
por el cable; mandarlo sólo para que el FE busque significaría 1,2 MB por apertura de lista. El
daemon ya tiene los transcripts en memoria (`sessionRuntime.meta.Conv`).

**Criterios de aceptación**
1. Con `q` no vacío, cada entrada suma `fragmento` (el snippet de RF-322) y las que no coinciden
   **no viajan**; la respuesta trae `total` para el rótulo `N de M` (RF-318).
2. `q` sólo con espacios ≡ sin `q`.
3. La normalización de RF-324 se aplica en el servidor: es un único lugar, no dos.

### RF-342 — `POST /api/sessions/{id}/conversaciones` — crear (desactivando la anterior)

**Criterios de aceptación**
1. `201` con la conversación nueva **y** el id de la que quedó inactiva.
2. `409` con motivo si la activa está `streaming` o `await` (RF-312, BR-CV-6) — mismo código que
   `Turn` usa hoy (`session_service.go:339-342`, `openapi.yaml:683-686`).
3. `404` si la sesión no existe.

### RF-343 — `POST /api/sessions/{id}/conversaciones/{cid}/activar` — retomar

**Criterios de aceptación**
1. `200` con la conversación activada.
2. `409` con motivo si hay turno en vuelo.
3. `404` si la conversación no pertenece a esa sesión — **no** se busca globalmente por `cid`
   (BR-CV-2).
4. La activación **no** spawnea el conductor: el spawn sigue siendo perezoso, en el primer `Turn`
   (`session_service.go:364-370`). La franja `--resume` (RF-310) se pinta con lo que la respuesta
   informa, y el `--resume` real ocurre en el turno siguiente.

### RF-344 — `PATCH /api/sessions/{id}/conversaciones/{cid}` — renombrar

**Criterios de aceptación**
1. Body `{ "titulo": "…" }`. `titulo` vacío o sólo espacios ⇒ `400`, y el título **no cambia** —
   mismo criterio que `Rename` hoy (`session_service.go:276`: `if frente = TrimSpace(frente); frente != ""`).
2. Un `PATCH` exitoso marca `titulo_editado: true` (RF-303).
3. **Títulos duplicados son legales** (RF-357): no hay unicidad que validar.

### RF-345 — Las rutas del historial B2 por arnés se retiran

**Criterios de aceptación**
1. `GET /api/sessions?arnes=…&cerradas=1` deja de aceptar `cerradas=1`; `?arnes=` sigue filtrando
   sesiones (`sessions.go:20-30`, se conserva).
2. `GET /api/sessions/cerradas/{id}/historial` (`router.go:124`) se retira o se re-apunta a sesiones
   archivadas (RF-306). Sea cual sea, **no queda una ruta viva que devuelva `200` con datos vacíos**
   — o sirve, o responde 404 con motivo.
3. `client.ts:218-225` pierde `conversacionesDeArnes` y `historialCerrada`.

### RF-346 — El contrato OpenAPI se actualiza en el mismo commit
`docs/architecture/contracts/api/openapi.yaml:605-741` (lo vigente) · `:856-878` (schema `Session`).

**Criterios de aceptación**
1. Se agregan las 5 rutas nuevas y se retiran/ajustan las de RF-345.
2. Se agrega el schema `Conversacion` y se corrige `Session`: hoy el schema está **stale** de antes
   de este paquete — `conv.items.rol` declara `enum: [user, assistant, sys]` (`:877`) pero el
   dominio tiene 4 roles desde `RolAct` (`session.go:56`), y faltan `cwd`, `cadena_cc`,
   `checkpoint`, `ctx_hist`, `rotacion_pendiente`, `reparacion`, `cerrada_en`, `turnos`.
3. `web/src/shared/api/types.ts:16-38` (`Session`) se parte igual que el dominio.

---

## H · Errores y degradación (6 RF)

### RF-347 — Daemon caído: se dice el motivo, con Reintentar 🎨
`mockup:878-904` (§6B) · `:899-903` (note) · BR-CV-10.

**Criterios de aceptación**
1. El motivo va **en la superficie**, no en un tooltip ni en la consola. Copy de `mockup:888-889`.
2. Botón `Reintentar` — mismo patrón que `onReintentar` del picker
   (`new-session-picker.tsx:143-156`).
3. **No** se degrada a «0 conversaciones»: eso sería un pass fabricado.
4. Vale para las tres operaciones: listar, crear y retomar. Al crear o retomar, además, **el estado
   anterior queda intacto** (RF-308 CA-2).
5. **El color del motivo NO es `--warn`** — ver `design.md` §Contradicciones C-3: el dibujo usa
   `--warn` sobre `--card`, que da **3,76:1** en tema claro y rompe el gate a11y (deuda abierta,
   `BACKLOG.md:59-64`).

### RF-348 — `--resume` que el CLI ya no reconoce: se auto-sana, y se ve

**Lo vigente que ya cubre el caso:** `tryHealResume` (`session_service.go:600-630`) detecta un
`--resume` cuyo proceso murió o erroró **antes** del `init`, limpia el `ClaudeSessionID` stale
(`:609`), respawnea fresco **una vez** (`resumeRetried`, `:608`) y **reenvía el turno pendiente**
(`:624-628`). Está enforced: `boundaries/sesion-viva-consistente.md` check `resume-auto-sana`,
`fitness/arch_test.go:TestResumeAutoSana`.

```gherkin
Escenario: retomar una conversación cuya JSONL fue GC'd
  Dado una conversación inactiva con claude_session_id "d303a93f…"
  Y esa sesión CC ya no existe para el CLI
  Cuando el operador la retoma y manda un turno
  Entonces el spawn con --resume falla antes del init
  Y el daemon reinicia fresco UNA vez y reenvía el turno
  Y el transcript sigue mostrando los turnos viejos desde Conv
  Y el detalle muestra el cc-id NUEVO
```

**Criterios de aceptación**
1. El heal es **silencioso en éxito** (comportamiento vigente, `:598-599`) — pero **este paquete
   agrega una marca inline** en el transcript, porque bajo CV-D11 el operador acaba de pedir
   explícitamente esa conversación y un cambio silencioso de `cc-id` es indistinguible de una
   rotación. Marca sugerida, misma forma que la de rotación (`mockup:785`): `⟳ hilo reiniciado ·
   checkpoint`.
2. `Conv` **no se pierde**: el repintado de RF-310 ya ocurrió y no depende del proceso.
3. El `Checkpoint` de la conversación viaja al proceso fresco por el mismo camino que la rotación
   (`session_service.go:406-408`).
4. Un **segundo** fallo emite `error` visible (`:617`), no un loop.

### RF-349 — El proceso `claude` de una conversación ya no existe: es el camino normal, no un error

**Criterios de aceptación**
1. El conductor vivo (`sessionRuntime.live`) es **memoria del daemon**, no estado persistido: tras
   un reinicio del daemon **ninguna** conversación tiene proceso, y eso ya es lo normal hoy
   (`spawnLocked` corre en el primer `Turn`, `session_service.go:364-370`).
2. Retomar una conversación sin proceso **no muestra error**: muestra la franja de RF-310 y espera
   al primer turno.
3. Sólo se muestra error si el spawn falla de verdad (RF-350).

### RF-350 — El `cwd` de la conversación ya no existe en disco

```gherkin
Escenario: la carpeta del arnés fue borrada o movida
  Dado una conversación cuya sesión apunta a /home/…/vitalia
  Y esa carpeta ya no existe
  Cuando el operador retoma y manda un turno
  Entonces el spawn falla con el motivo real del resolver
  Y la conversación vuelve a `idle` (no queda colgada en streaming)
  Y el motivo queda en el transcript, nombrando la ruta
```

**Criterios de aceptación**
1. `spawnLocked` ya envuelve el fallo con el arnés (`session_service.go:386-388`:
   `resolve arnés %q workdir`) y `Turn` ya revierte a `idle` (`:365-369`). Este RF exige que el
   mensaje **nombre la ruta**, no sólo el arnés.
2. La conversación **no se pierde**: sigue en la lista, con sus turnos.
3. Es distinguible del caso RF-348 (resume stale): son dos fallos distintos y dos copys distintos.

### RF-351 — El arnés de la sesión desaparece del Portafolio con conversaciones vivas

```gherkin
Escenario: desvincular del Portafolio un arnés con sesión abierta
  Dado una sesión sobre `vitalia` con 4 conversaciones
  Cuando el operador desvincula `vitalia` desde la vista Portafolio
  Entonces la sesión sigue en el rail con sus 4 conversaciones
  Y la lista se sigue pudiendo abrir y buscar (los datos son nuestros)
  Y el primer turno nuevo falla con el motivo del resolver, como RF-350
```

**Criterios de aceptación**
1. **Nada se borra en cascada.** El registro de sesiones es independiente del Portafolio; borrar
   sesiones del operador porque desvinculó un arnés sería pérdida silenciosa de datos.
2. Listar y buscar **no** dependen del arnés: dependen de `session_id` (BR-CV-2).
3. El fallo aparece cuando se intenta usar el conductor, y dice por qué.

### RF-352 — El nodo del alcance deja de existir tras un reindex

```gherkin
Escenario: el nodo elegido desaparece
  Dado un nodo `hipaa-check` en el chip de alcance
  Cuando un turno reindexa el arnés y ese nodo ya no está en el grafo
  Entonces el chip de alcance se limpia solo
  Y queda un rastro `sys` en el transcript diciendo qué nodo se quitó y por qué
  Y la fila 3 del cromo desaparece (RF-329)
```

**Criterios de aceptación**
1. El reindex-en-vivo ya existe y ya avisa al FE: `map-live-store` bump por
   `harness_id` (`sessions-store.ts:135`). Este RF engancha ahí, no inventa un canal.
2. El alcance **no se conserva apuntando a un nodo fantasma**: mandarlo al conductor como línea de
   contexto (`sessions-store.ts:189`) sería mandarle una referencia falsa.
3. El scope es por sesión (`sessions-store.ts:251-255`, RF-118) y **sigue siéndolo** — no baja a la
   conversación. Motivo: es el nodo del Mapa que estás mirando, no el tema del hilo.

---

## I · Accesibilidad (5 RF)

### RF-353 — Todo texto y todo control nuevos pasan el gate axe en los dos temas
`web/.storybook/preview.ts:7` (`a11y: { test: "error" }` global) · `mockup:945-952` (contraste
medido del dibujo).

**Criterios de aceptación**
1. Texto normal ≥ 4,5:1; no-textual ≥ 3:1, en `light` **y** en `dark`.
2. Los dos elementos que el mockup ya midió se respetan tal cual: `«activa»` y `＋` van en
   `--foreground` (16,39:1 claro / 13,2:1 oscuro), **no** en `--primary` (2,20:1 y 2,51:1 — medido y
   re-verificado en este spec).
3. **Ningún elemento nuevo usa `--warn` como color de texto sobre `--card`, `--secondary`,
   `--background` o `--accent-soft`** — medidos hoy: 3,76 · 3,31 · 3,53 · 3,29:1, los cuatro por
   debajo de 4,5. Ver `design.md` §Contradicciones C-3.
4. Ninguna story del paquete baja `a11y` a `"todo"` para pasar.

### RF-354 — La lista es navegable por teclado
Sin panel propio (el mockup dibuja hover/foco, no la mecánica).

**Criterios de aceptación**
1. `↑`/`↓` mueven entre filas, `Enter`/`Espacio` activan, `Escape` cierra la lista y devuelve el
   foco al chevron que la abrió, `Tab` sale del grupo.
2. Foco visible en toda fila, con anillo, no sólo con fondo.
3. Las filas deshabilitadas por RF-312 son **anunciadas** como tales (`aria-disabled` + el motivo
   accesible), no simplemente salteadas.

### RF-355 — El foco al abrir la lista va al buscador; sin buscador, a la fila activa

**Criterios de aceptación**
1. Precedente reusado: `NewSessionPicker` pone `autoFocus` en su buscador
   (`new-session-picker.tsx:184-189`, con el `noAutofocus` de biome apagado a propósito).
2. Con una sola conversación no hay buscador (RF-325) ⇒ el foco va a la fila activa.
3. Al cerrar, el foco vuelve al control que la abrió — nunca al `<body>`.

### RF-356 — Los estados se anuncian; nada depende sólo del color

**Criterios de aceptación**
1. La fila activa se distingue por el radio marcado (`mockup:518-520`) **y** el rótulo `activa`
   (`:524`), no sólo por el tinte.
2. El esqueleto de carga lleva `role="status"` + `aria-live="polite"` + `aria-label` — calca
   `PickerSkeleton` (`new-session-picker.tsx:309-326`), tal como el mockup manda (`:872-875`).
3. El error de lectura lleva `role="alert"` — calca `new-session-picker.tsx:145`.
4. El chip de ctx expone su valor como texto (`68%`), no sólo como barra
   (`chat-dock.tsx:90` ya lo hace hoy y se conserva).

### RF-357 — `⌘K` sigue siendo el atajo del dock, y no gana hermanos
`web/src/app/App.tsx:15-17` · `chat-dock.tsx:29` (el `title` lo promete).

**Criterios de aceptación**
1. Este paquete **no** agrega atajos globales. La lista y el buscador se abren desde sus controles.
2. `⌘K` con la lista abierta colapsa el dock entero (comportamiento vigente de `toggleChat`,
   `sessions-store.ts:257`); la lista se desmonta (RF-315).

---

## J · No-objetivos (explícitos)

| No se hace | Por qué |
|---|---|
| Historial del **arnés** en `Hist` | CV-D1 — otro concepto, otro paquete |
| Buscador **global** sobre todas las sesiones | CV-D4 |
| N conversaciones **en paralelo** en una sesión | CV-D7 |
| Migrar las llaves de las 3 cerradas | CV-D6 — el dato se borra |
| Reparar el token `--warn` | preexistente, `BACKLOG.md:59-64`; este paquete sólo evita reintroducirlo |
| Arreglar el schema `Session` del OpenAPI **más allá** de lo que este paquete parte | RF-346 lo deja consistente; los campos ya stale se corrigen de paso porque tocarlos es inevitable |
| Persistir el estado abierto/cerrado de la lista entre sesiones | el dock ya persiste su **ancho** (`shell-page.tsx:15`), no su contenido |

---

## K · Cobertura de escenarios — 44 casos con su respuesta especificada

**El corazón del encargo.** Cada fila define **qué hace el sistema**, no sólo qué se espera.
`resp.` = la respuesta especificada. `RF` = quién la manda.

### K.1 · Forma de la lista

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-01** | sesión **sin ninguna** conversación | **imposible por invariante**: toda sesión nace con una (BR-CV-1). Si el registro en disco trajera una sesión sin conversaciones (migración parcial, edición a mano), el daemon **le crea una activa vacía al cargar** y lo dice en el log. Nunca se pinta una sesión con lista vacía | RF-301 · RF-335 |
| **E-02** | **una sola** conversación | la lista se abre igual, con el rótulo `1 conversación de la sesión X`; el **buscador no se dibuja**; debajo, el mensaje «Esta sesión recién arranca. Cuando abras otra con ＋, esta queda acá y la podés retomar» (`mockup:672-675`) | RF-325 |
| **E-03** | **muchas** (≥ 4) | lista scrolleable dentro del área del transcript; el buscador aparece; el rótulo dice el total | RF-317 · RF-321 |
| **E-04** | una de **0 turnos** | se muestra **igual**: `nueva conversación` · `sin turnos todavía · ctx 0 %`. 0 no se esconde (`mockup:544`, `:667`) | BR-CV-9 · RF-319 |
| **E-05** | conversación con turnos pero **sin `claude_session_id`** (creada, nunca spawneada) | fila normal; el detalle muestra `◍ sin sesión CC` (literal vigente, `chat-dock.tsx:82`) dentro del disclosure | RF-328 |

### K.2 · Crear

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-06** | crear con la sesión **quieta** | nace activa, 0 turnos, ctx 0 %; la anterior queda inactiva con su `Conv` persistido; el vacío nombra a la desactivada | RF-307 · RF-308 |
| **E-07** | crear con un **turno en vuelo** (`streaming`) | `＋` **deshabilitado** con `title` «esperá a que termine el turno (■ para interrumpir)». Si igual llega la llamada (carrera), el servidor responde **409** y el FE no altera nada | RF-312 · RF-342 |
| **E-08** | crear con un **permiso pendiente** (`await`) | idéntico a E-07, con motivo «esperá tu decisión de permiso». El operador tiene dos salidas legales: resolver la tarjeta, o `■` interrumpir (`chat-dock.tsx:351-359`) — ambas ya existen | RF-312 |
| **E-09** | crear **dos veces seguidas** rápido | la segunda encuentra la primera en 0 turnos; **se crea igual** (no hay deduplicación por «vacía»): dos conversaciones vacías son legales y el operador las ve y las distingue por su fecha de creación | RF-320 |

### K.3 · Retomar

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-10** | retomar con la sesión quieta | franja `--resume <ccid[0:8]>`; transcript repintado desde `Conv`; detalle abierto solo; la anterior queda inactiva | RF-310 · RF-314 |
| **E-11** | retomar **con turno en vuelo** | filas inactivas deshabilitadas con motivo; **409** del servidor si llega la llamada | RF-312 · RF-343 |
| **E-12** | retomar una cuyo **proceso `claude` ya no existe** | **no es un error**: es el estado normal tras cualquier reinicio del daemon. Se activa, se repinta desde `Conv`, y el spawn ocurre perezoso en el turno siguiente | RF-349 |
| **E-13** | retomar una cuyo **`ClaudeSessionID` el CLI ya no reconoce** (`--resume` falla) | el spawn muere antes del `init`; `tryHealResume` limpia el id stale, respawnea **fresco una vez** y reenvía el turno; `Conv` intacto; **marca inline nueva** «⟳ hilo reiniciado · checkpoint» para que el cambio de `cc-id` no sea silencioso; el `Checkpoint` viaja al proceso fresco. Un segundo fallo ⇒ `error` visible | RF-348 |
| **E-14** | retomar una cuyo **`cwd` ya no existe** | activar **funciona** (es dato nuestro); el fallo aparece al primer turno: `spawnLocked` erra, la conversación vuelve a `idle` y el motivo con **la ruta** queda en el transcript | RF-350 |
| **E-15** | retomar la que **ya está activa** | no-op: cierra la lista, cero llamadas, cero re-spawn | RF-316 |
| **E-16** | retomar una de **0 turnos** | se activa; el transcript queda vacío con el copy de RF-307; no hay `--resume` que hacer (sin `cc-id`) ⇒ **la franja no se dibuja** | RF-310 · RF-328 |

### K.4 · Rotación

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-17** | rotación **en medio de un turno** | **no ocurre por construcción**: `RotacionPendiente` se marca al `result` (`session_service.go:505-507`) y se ejecuta al **inicio del turno siguiente** (`:347-349`, comentario literal: «Se rota ENTRE turnos por construcción»). Este paquete no toca ese mecanismo | RF-313 |
| **E-18** | rotación con el **dock colapsado** | ocurre igual; la marca queda en `Conv`; al reabrir está en su lugar cronológico; el dock **no** se abre solo, no hay toast | RF-315 |
| **E-19** | rotación con **la lista abierta** | la lista sigue mostrando **N**, no N+1; el título no cambia; el ctx de la fila baja al valor nuevo | RF-313 |
| **E-20** | rotación en una conversación **inactiva** | **imposible**: sólo la activa tiene conductor y sólo ella recibe turnos (BR-CV-1) | RF-301 |

### K.5 · Buscador

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-21** | buscar con **una sola** conversación | el buscador **no existe** — nada que filtrar (`mockup:681`) | RF-325 |
| **E-22** | **sin coincidencias** | vacío que **dice dónde buscó y cuántas miró**: «Ninguna conversación de esta sesión menciona «X»» + «Se buscó en el título y en el texto de las 4.» + `Limpiar búsqueda` | RF-325 |
| **E-23** | **acentos y mayúsculas** | «VACIO» encuentra «vacío». Normalización NFD + strip de diacríticos + lowercase, en el servidor | RF-324 · RF-341 |
| **E-24** | la **activa** entre los resultados | participa como cualquier otra, conservando tinte, radio y rótulo `activa`; si no coincide, **no aparece** | RF-323 |
| **E-25** | **transcript enorme** | scan lineal en memoria en el daemon; sólo viajan las que coinciden + su fragmento. Medido hoy: 12 281 B la más larga; 100 conversaciones ≈ 1,2 MB. Si alguna vez deja de alcanzar, el upgrade es FTS5 — **fuera de alcance** (CV-D8) | RF-341 |
| **E-26** | query en **blanco / sólo espacios** | ≡ sin query: lista completa, sin fragmentos, rótulo con el total | RF-341 |
| **E-27** | coincide **sólo en el título** | la fila aparece **sin fragmento** — no hay nada que explicar | RF-322 |
| **E-28** | buscar **con turno en vuelo** | funciona: buscar no cambia estado. Sólo las **acciones** están deshabilitadas | RF-312 |
| **E-29** | el término coincide **muchas veces** en un transcript | se muestra **un** fragmento: el primero. La fila no es un visor de resultados | RF-322 |

### K.6 · Renombrar

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-30** | renombrar a **vacío** | **descarta**: el título vuelve al anterior. Servidor: `400`, título intacto. Calca `session-rail.tsx:199-200` y `sessions-store.ts:160-161` | RF-331 · RF-344 |
| **E-31** | renombrar a un **título duplicado** | **se acepta**. No hay unicidad: la identidad es el `id`, y dos conversaciones sobre el mismo tema con el mismo nombre son un caso real. Las distingue la fila (fecha + turnos + ctx) | RF-344 |
| **E-32** | **Escape** durante la edición | descarta, sale del modo edición, el foco vuelve al título. Calca `session-rail.tsx:267-270` | RF-331 |
| **E-33** | renombrar **mientras llega un turno** | permitido: renombrar no toca el conductor. El turno sigue llegando y el transcript sigue creciendo detrás del input. Si el turno **derivaría** el título (primer turno de una conversación nueva), gana **el operador**: `titulo_editado` ya quedó en `true` | RF-303 · RF-331 |
| **E-34** | **blur** sin Enter ni Escape | confirma, como el rail (`session-rail.tsx:263`, `onBlur={commit}`) — vocabulario reusado, no inventado | RF-331 |

### K.7 · Transporte y concurrencia

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-35** | **daemon caído al listar** | «No se pudieron leer las conversaciones — daemon no responde» + `Reintentar`. **Nunca** «0 conversaciones» | RF-347 |
| **E-36** | **daemon caído al retomar** | la activa **no cambia**; el motivo se muestra; la lista queda abierta para reintentar | RF-347 · RF-308 |
| **E-37** | **daemon caído al crear** | idem: nada se crea, nada se desactiva, el motivo se muestra | RF-347 · RF-308 |
| **E-38** | **timeout** | tratado como fallo de transporte, con el mismo copy + motivo «no respondió a tiempo». **No** se agrega timeout propio en `client.ts` — el repo ya decidió que el tope vive en el daemon (`client.ts:283-284`, dictado) | RF-347 |
| **E-39** | respuesta **409** | es la respuesta *correcta* a un turno en vuelo, no un error de transporte: se muestra el motivo del 409 («hay un turno en vuelo»), no un genérico. El FE ya distingue 409 por `ApiError.status` (`client.ts:33-42`, `sessions-store.ts:205`) | RF-312 · RF-342 |
| **E-40** | **dos ventanas de la app** sobre la misma sesión | **Verificado:** el shell Tauri tiene single-instance y una segunda instancia **reenfoca** la existente en vez de duplicar (`web/src-tauri/src/lib.rs:52-59`, `unminimize + show + set_focus`; capability `tauri/single-instance-reenfoca.yaml` = CAP-79, check `single-instance-reenfoca` en `boundaries/core-no-importa-shell.md:78`, severidad `warn`). El caso real que queda es **el daemon sirviendo el SPA en `:4200` abierto en un navegador además de la app**. Respuesta: el daemon es la verdad; ambas vistas reciben los mismos frames por SSE (`connectDock`, `sessions-store.ts:129-136`) y el 409 protege de dos turnos. **Lo que este paquete debe garantizar:** crear/retomar/renombrar emiten un frame que las otras vistas aplican, para que la lista no quede stale. **Hueco declarado:** hoy `DockFrame` (`types.ts:68-90`) no tiene `kind` para cambios de conversación — hace falta uno nuevo, y el `run_id`/dedup existente (`sessions-store.ts:266`) no aplica porque no son frames de turno | RF-342 · RF-343 |
| **E-41** | **SSE reconecta** durante la retoma | el broker ya replaya por `Last-Event-ID` y el FE dropea frames de runs finalizados (`boundaries/sesion-viva-consistente.md` checks `sin-perdida-silenciosa` y `frames-idempotentes-run-id`). El frame de conversación de E-40 **debe ser idempotente por su propio id**, no por `run_id` | RF-343 |
| **E-42** | **cambiar de sesión en el rail** con la lista abierta | la lista se **cierra** y se descarta la búsqueda: su alcance es la sesión activa (CV-D4) y quedarse abierta mostrando otra sesión sería mentir. Calca la disciplina de `cerrarPicker` (`session-rail.tsx:44-47`: «cerrar descarta búsqueda/selección — cero efectos») | RF-318 |

### K.8 · Datos en disco y entorno

| id | escenario | respuesta especificada | RF |
|---|---|---|---|
| **E-43** | instalación vieja con `sessions.json` + `sesiones-cerradas.json` del formato actual | migración al arrancar: 5 sesiones → 5 × 1 conversación; `sessions.json` **se conserva** como respaldo; `ultima_interaccion` queda **vacío** (no se inventa); `sesiones-cerradas.json` lo trata E-46 | RF-335 |
| **E-44** | archivos **corruptos** | no se pisan: se renombran a `.corrupto-<AAMMDDHHMM>`, el daemon arranca con registro vacío y **lo dice** con la ruta del respaldo. Ninguna semilla ilustrativa tapa la corrupción | RF-336 |
| **E-45** | **sin permisos de escritura** | la operación falla con el motivo real del filesystem, en la superficie; el estado en memoria no diverge del disco en silencio | RF-338 |
| **E-46** | **CV-D6: eliminar las 3 cerradas** | procedimiento de 6 pasos: detener daemon → confirmar muerto → copia `.bak-<fecha>` → verificar la copia (3 entradas) → `rm` del archivo **entero** → arrancar y confirmar → rastro en `PARIDAD.md`. **No se borran** `sessions.json`, las JSONL nativas, `index.db` ni `telemetria.db`. Manual y del operador, jamás automático al actualizar | RF-337 |
| **E-47** | **crash a mitad de la migración** | temp + `rename` (`registry.go:77-95`): o el viejo entero o el nuevo entero. Al reintentar, la idempotencia de RF-335 aplica | RF-339 |
| **E-48** | **nodo del Mapa desaparece tras reindex** | el chip de alcance se limpia solo, con rastro `sys` que nombra el nodo y el motivo; la fila 3 del cromo desaparece. Engancha en el bump de reindex-vivo que ya existe (`sessions-store.ts:135`) | RF-352 |
| **E-49** | **el arnés desaparece del Portafolio** con conversaciones vivas | **nada en cascada**: la sesión y sus conversaciones siguen; listar y buscar funcionan (dependen de `session_id`); el fallo aparece al primer turno, con el motivo del resolver | RF-351 |
| **E-50** | **cerrar la sesión entera** con N conversaciones | el `✕` del rail se conserva (`session-rail.tsx:241-253`); al cerrar, **las N conversaciones se archivan con la sesión** (con su `Conv`, RF-305) en `sesiones-archivadas.json`; el rail pierde la tarjeta. `canClose` sigue exigiendo ≥2 sesiones (`session-rail.tsx:193`) | RF-306 |

---

## L · Trazabilidad — RF → panel del mockup → decisión

| RF | panel (`mockup-conversaciones-panel.html`) | CV-D |
|---|---|---|
| RF-300 | — **sin superficie, a propósito**: es el modelo. Se ve por consecuencia en §3A (`:516`) y §4B (`:761-766`) | CV-D3 |
| RF-301 | — sin superficie (invariante); consecuencia visible en §3D `:663-670` | CV-D3 · CV-D7 |
| RF-302 | §3A `:524` · §7 `:928` | CV-D12 |
| RF-303 | §2D `:466` · §2A `:349` | CV-D9 |
| RF-304 | §3A `:522` · §3D `:667` · callout `:936-943` | CV-D13 |
| RF-305 | — sin superficie directa; la habilita §3B `:590` (snippet) y §4A `:726` (repintado) | CV-D8 · CV-D11 |
| RF-306 | §5B `:840-844` | CV-D12 |
| RF-307 | §4B `:733-767` · `＋` en `:355` | CV-D2 · CV-D7 |
| RF-308 | §4 lead `:688-691` | CV-D7 |
| RF-309 | §2A note `:369-372` (**la ausencia es el artefacto**) | CV-D7 |
| RF-310 | §4A `:694-731` | CV-D11 |
| RF-311 | — **sin superficie, a propósito**: el rastro cae en el transcript, que ya existe (`sessions-store.ts:405-428`). El mockup lo dejó abierto en `:954-959` | CV-D7 |
| RF-312 | — **superset del mockup**: el dibujo pinta `＋` siempre habilitado porque es estático. Lo manda BR-CV-6, con el precedente `busy` de `chat-dock.tsx:293` | CV-D7 · CV-D11 |
| RF-313 | §4C `:769-802` · marca `:785` | CV-D10 |
| RF-314 | §4A `:703`+`:713-715` · §4C `:774`+`:780-782` | CV-D14 |
| RF-315 | — sin superficie (el dock colapsado no dibuja nada) | CV-D10 |
| RF-316 | §3A `:518-525` (la fila activa) | CV-D11 |
| RF-317 | §3 lead `:489-493` · §3A `:511-549` | CV-D2 |
| RF-318 | §3A `:516` · §3B `:583` · §3D `:661` | CV-D4 |
| RF-319 | §3A `:522`, `:530`, `:537`, `:544` · §3D `:667` | CV-D13 |
| RF-320 | §3A `:517-547` (el orden dibujado) · callout `:954-959` | CV-D13 |
| RF-321 | §3B `:578-582`, note `:611-615` | CV-D8 |
| RF-322 | §3B `:590`, `:599` | CV-D8 |
| RF-323 | §3B `:585-593`, note `:614` | CV-D8 |
| RF-324 | — sin superficie (matcher) | CV-D8 |
| RF-325 | §3C `:634-638` · §3D `:660-676` | CV-D8 |
| RF-326 | §1 `:294-327` (el antes) · §2A `:341-377` (el después) | CV-D14 |
| RF-327 | §2A `:350-352` · §2B `:388-390` | CV-D14 |
| RF-328 | §2B `:396-399`, note `:410-415` | CV-D14 |
| RF-329 | §2C `:433-439`, note `:450-455` | CV-D14 |
| RF-330 | CSS `:184-189` · §2B `:388` | CV-D14 |
| RF-331 | §2D `:458-484` | CV-D9 |
| RF-332 | §2A `:346` · §1 `:299` (conserva el vigente) | CV-D15 |
| RF-333 | §5A `:814-827` · §5B `:829-845` | CV-D2 |
| RF-334 | — **sin superficie, a propósito**: es el transporte del widget | CV-D2 · CV-D5 |
| RF-335 | — **sin superficie, a propósito**: corre al arrancar el daemon. Su consecuencia visible es que `ultima_interaccion` sale vacío, y el mockup ya declaró ese gap (`:936-943`) | CV-D3 |
| RF-336 | — sin superficie (arranque) | — |
| RF-337 | §7 `:922` (declarado «sin superficie (dato)») | CV-D6 |
| RF-338 | — sin superficie propia; reusa el estado de error de §6B `:878-904` | — |
| RF-339 | — sin superficie | — |
| RF-340…RF-346 | — **sin superficie, a propósito**: es el cable. Todo lo que devuelven se pinta en §3 y §4 | CV-D2 · CV-D4 · CV-D12 |
| RF-347 | §6B `:878-904` · §6A `:856-876` (el cargando) | — |
| RF-348 | — **superset**: el mockup dibuja la franja de retoma en su variante feliz (`:709-712`) y su variante mala (`.resumebar.bad`, CSS `:242`) **pero no la usa en ningún panel**. Este RF la usa | CV-D11 |
| RF-349 | — sin superficie (es el camino normal) | CV-D11 |
| RF-350 | — sin superficie propia; el motivo cae en el transcript | CV-D11 |
| RF-351 | — sin superficie propia | CV-D5 |
| RF-352 | §2C `:433-439` (la fila que desaparece) | CV-D14 |
| RF-353 | callout `:945-952` (el método de medición) | — |
| RF-354…RF-356 | §6A note `:872-875` (el patrón `role="status"`) — el resto **sin superficie a propósito**: la a11y es contrato, no dibujo | — |
| RF-357 | §2A `:346` (`title` que promete ⌘K) | CV-D15 |

**Decisiones sin RF propio, declaradas:**

| CV-D | por qué no tiene RF | dónde vive |
|---|---|---|
| **CV-D1** | fuera de alcance total: el view-strip no se toca | §0 · `mockup:917` |
| **CV-D5** | es una **propiedad** del modelo, no un requisito verificable aparte | BR-CV-2, que atraviesa RF-300, RF-340, RF-343, RF-351 |

---

## M · Capabilities a crear/modificar

Doctrina: **ningún cambio de código sin capability**
(`boundaries/codigo-traza-a-capability.md`). El CAP más alto en uso hoy es **CAP-139**
(verificado: `grep -rhoE "CAP-[0-9]+" docs/` → máx. 139), así que el bloque nuevo arranca en
**CAP-140**.

| acción | capability | módulo | qué cubre |
|---|---|---|---|
| **nueva** | `conversaciones-de-una-sesion` | `usecases` | el modelo Sesión→N Conversaciones, la invariante de una activa, crear/retomar/renombrar |
| **nueva** | `buscar-en-el-transcript` | `usecases` | el scan en memoria sobre `Conv` + fragmento |
| **nueva** | `panel-de-conversaciones` | `fe-chat` | lista + buscador + `＋` + retomar en el dock |
| **modificar** | `chat-cc` (**CAP-68**, `fe-chat/chat-cc.yaml`) | `fe-chat` | el cromo baja a 2 filas; el ctx pasa a chip-disclosure |
| **modificar** | `historial-de-conversaciones` (**CAP-98**, `usecases/historial-de-conversaciones.yaml`) | `usecases` | deja de ser por-arnés/por-cerrada; sus punteros a `conversaciones-store.ts` y su `valida: TestCloseArchivaMetadata` cambian |
| **modificar** | `gestion-de-sesiones-crud` (**CAP-59**) | `usecases` | `Close` archiva con `Conv`; `Create` nace con conversación |
| **modificar** | `rail-de-sesiones` (**CAP-72**, `fe-shell/rail-de-sesiones.yaml`) | `fe-shell` | el picker pierde `ConversacionesDelArnes` |
| **modificar** | `acotar-alcance` (`fe-chat/acotar-alcance.yaml`) | `fe-chat` | `ScopeRow` deja de ser fila fija |
| **modificar** | `rotacion-de-contexto` (`usecases/`) | `usecases` | la rotación pasa a colgar de la conversación |
| **modificar** | `superficie-rest` (`http-sse/`) | `http-sse` | 5 rutas nuevas, 2 retiradas |

---

## N · Huecos declarados (nada tapado)

| id | hueco | estado |
|---|---|---|
| **H-1** | **`chat-dock.tsx` no tiene story.** Todo el widget que este paquete reescribe nace sin baseline visual — sólo hay stories de `dictado-button` y `permission-card` | se cierra en `plan-storybook.md` §1.3, que exige la story del dock **antes** de tocarlo |
| **H-2** | **No hay `kind` de `DockFrame` para cambios de conversación** (`types.ts:68-90`) | E-40/E-41 lo exigen; el diseño del frame va en `design.md` §7 |
| **H-3** | **El OpenAPI no documenta las rutas B2** (`?cerradas=1`, `/cerradas/{id}/historial`) — deuda **anterior** a este paquete | RF-346 la cierra al retirarlas |
| **H-4** | **`ultima_interaccion` de lo migrado queda vacío**: el dato no existe y no se puede inferir | declarado, no inventado (BR-CV-14) |
| **H-5** | **El contraste de `--warn`** rompe el gate a11y y el mockup lo usa (`:887`) | RF-353 lo prohíbe en superficie nueva; el fix del token sigue en `BACKLOG.md:59-64` |
| **H-6** | **CV-D8 midió 9,4 KB donde hoy hay 12,0 KB** | declarado en §Estado; no cambia la conclusión |
| **H-7** | **`Conv` persistido de N conversaciones infla `sesiones.json`**: 5 sesiones × 4 conversaciones × 12 KB ≈ 240 KB en un archivo que hoy son 17 KB, releído y reescrito entero en cada `persistLocked` (`session_service.go:902-912`, snapshot total con `json.MarshalIndent`, `registry.go:73`) | **no medido bajo carga real.** Si molesta, el upgrade natural es un archivo por sesión bajo `~/.arnesia/sessions/<id>/` (carpeta que **ya existe**, la usa el checkpoint de rotación, RF-196). Fuera de alcance; anotado para que nadie lo descubra en producción |
| **H-8** | 🔴 **La rotación no emite ningún frame SSE.** `session_rotacion.go` tiene cero `s.publish`; el breadcrumb vive sólo en el `Conv` del backend. El §4C del mockup dibuja una marca que **hoy el operador no vería hasta recargar** | lo cierra RF-313 CA-4 junto con el frame nuevo de H-2. **Es un hallazgo de este spec, no una decisión firmada** — hay que decidir si el frame es de conversación (H-2) o uno propio |
| **H-9** | **`DirParaCwd` mapea `_`, `.` y `-` todos a `-`** (`internal/adapters/history/reader.go:39-49`): dos `cwd` distintos pueden colisionar en el mismo directorio de corpus | preexistente, sin relación con este paquete. Anotado porque RF-306 conserva el lector; no se toca |
