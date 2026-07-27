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
   `dev-full-cycle`). ⚠ **Corregido por F-1**: acá se decía que el re-key `cca314a` «migró las
   vivas» — es falso, no migró ninguna, y 4 de 5 vivas también tienen id pelado.

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

## CV-D7 · Una conversación ACTIVA por sesión, a la vez

Crear una conversación **desactiva la anterior** — y no hay ningún botón de cerrar: el único
cierre es implícito. No hay N hilos corriendo en paralelo dentro de una sesión. El paralelismo
real sigue siendo el de siempre: N sesiones en el rail. La lista del dock es 1 activa + N
inactivas (ver CV-D12: no se dice «cerrada»).

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

## CV-D11 · Seleccionar una conversación inactiva la RETOMA

No es de solo lectura: seleccionarla la vuelve la activa (`--resume` sobre su `ClaudeSessionID`,
con `CadenaCC`/`Checkpoint`/`Cwd` intactos) y desactiva a la que estaba activa. Es el mismo
movimiento de CV-D7 en la otra dirección: siempre exactamente **una** activa por sesión.

Consecuencia sobre CV-D8: persistir `Conv` al desactivar deja de ser «guardar un archivo muerto»
— es lo que se repinta al retomar, antes de que el stream vivo se reenganche. El mismo rol que
`Conv` ya cumple hoy al cambiar de sesión en el rail (`session.go:59-61`).

## CV-D12 · Vocabulario: activa / inactiva. NO «cerrada»

Si se retoma, no está cerrada. La conversación tiene dos estados y el cierre es reversible:
**activa** (una, la del dock) e **inactiva** (las demás, retomables).

Consecuencia estructural: el registro aparte `~/.arnesia/sesiones-cerradas.json` +
`SetArchivoCerradas` (`cmd/arnesia/main.go:325-328`) **deja de tener sentido** — modelaba un
archivo terminal. Las conversaciones son todas pares dentro de su sesión, con una marcada activa.
Se decide en spec si el archivo desaparece o queda como registro de sesiones enteras cerradas.

## CV-D13 · La fila muestra: última interacción · nº de turnos · ctx final

Y el título (CV-D9). Nada más — la fila no es una tarjeta.

**Campo nuevo requerido:** «última interacción» no existe hoy en `Session`; lo único temporal es
`CerradaEn` (`session.go:121-124`), que es otra cosa. Hace falta un timestamp estampado en cada
turno, no al desactivar — si no, una conversación inactiva mentiría la fecha del último mensaje.
`Turnos` y `CtxPct` ya existen y bajan a la conversación por CV-D3.

## CV-D14 · El cromo del dock baja de 4 filas a 2

Conversada sobre el mockup: el operador contó lo que hay que leer antes del primer mensaje y pidió
recortarlo. Hoy son **cuatro filas** en un dock de 360 px (header · SessionLine · ScopeRow, más la
fila de conversación que agregaba este paquete), y `vitalia` aparece **tres veces** — SessionLine,
chip de Alcance y placeholder del composer — sobre un rail y un breadcrumb que ya lo dicen.

Queda así:

| Fila | Qué lleva | Cuándo |
|---|---|---|
| 1 | pip · frente de la sesión · colapsar | siempre (**intacta**, sin cambios) |
| 2 | ▶ título · **chip de ctx** · 🔍 · ＋ | siempre |
| 3 | chip del nodo en alcance, con ✕ | **solo con un nodo elegido** |

- **El ctx es un chip-disclosure.** Es la única cifra accionable de la `SessionLine` (dispara la
  rotación, RF-195) y se queda a la vista. Un clic abre el detalle con `cc-id · arnés · modelo ·
  cwd` — las mismas cifras del vigente **más el `cwd`**, que hoy no se ve en ningún lado y es el
  confinamiento real del conductor. **Nada se pierde**: deja de costar una fila permanente.
- **El detalle se abre solo al retomar (CV-D11) y al rotar (CV-D10)** — son los dos momentos en que
  el `cc-id` cambia, y eso hay que verlo sin buscarlo.
- **`ScopeRow` deja de ser fila fija.** Su estado vacío gastaba una fila entera en el hint
  «selecciona un nodo en el Mapa para acotar» + un chip del arnés redundante. Con nodo elegido la
  fila aparece y se gana el lugar: cambia qué le estás pidiendo. El ✕ para quitarlo se conserva
  literal (`chat-dock.tsx` L62-69).

Alternativa **descartada** por el operador: sacar también la fila 1 (el frente de la sesión ya está
en la tarjeta resaltada del rail, al lado, y en el breadcrumb del topbar). Habría dejado 1 sola fila
de cromo, pero quita superficie firmada (regla dura 3 de `mockups/INDEX.md`). Se conserva.

## CV-D15 · El glifo de colapsar pasa de `⟩` a `»`

`⟩` (U+27E9) se lee como un paréntesis suelto. `»` es el glifo que el **rail ya usa** para lo mismo
(`session-rail.tsx:96`: `«` colapsar / `»` expandir, la punta apunta hacia donde se mueve el panel).
El dock vive a la derecha, así que colapsarlo lo empuja a la derecha: `»`. Vocabulario reusado, no
inventado. La palabra «colapsar» y el `title` se conservan.

---

## Correcciones de hecho y enmiendas (post-relevamiento, 2026-07-26)

El relevamiento as-is (`relevamiento-as-is.md`) desmintió tres cosas que este archivo afirmaba y
destapó dos requisitos que ninguna decisión cubría. Se corrigen acá, sin reescribir la historia:
las decisiones firmadas siguen firmadas, lo que cambia es el hecho que las fundamentaba.

### F-1 · El re-key `cca314a` NO migró nada — el bug es más grande

Arriba, en §«Lo que se verificó en vivo», este archivo decía que el re-key «migró las vivas, no las
cerradas». **Falso**: el commit no migró ninguna, y verificado contra el daemon vivo **4 de 5
sesiones VIVAS también tienen id pelado**. La consulta `?arnes=sin-home~vitalia~vitalia` esconde
**3 vivas además de las 2 cerradas**.

Consecuencia sobre CV-D6: borrar `sesiones-cerradas.json` **no cura `sessions.json`**. Ver P-1.

### F-2 · CV-D8 se apoyaba en una cifra vieja

«90 turnos = 9,4 KB» era la medición del momento. Medido hoy: **12 095 bytes**. La conclusión no
cambia (100 conversaciones ≈ 1,2 MB, sigue siendo scan en memoria sin FTS5); el número sí.

### F-3 · Archivar destruye `Conv` **y `Checkpoint`** — y hay una capability que lo afirma

`session_historial.go:56-58` hace `cerrada.Conv = nil; cerrada.Checkpoint = ""`, con un test que lo
cementa (`session_historial_test.go:73`) y **CAP-98 afirmándolo como ley de negocio**.

CV-D8 ya decidió persistir `Conv`. Lo que **ninguna decisión cubría** es el `Checkpoint`: CV-D11
promete retomar con el checkpoint intacto, pero hoy se borra. Enmienda: **CV-D8 se extiende al
`Checkpoint`** — sin él, retomar una conversación rotada arranca sin el digest de su propia
rotación, que es exactamente el estado que RF-196 existe para no perder. Implica **modificar CAP-98
y su test**: no es un bug, es una ley que deja de valer bajo CV-D3, y el cambio tiene que ser
explícito y trazado, jamás un test borrado en silencio.

### F-4 · Todo el estado in-flight es por-SESIÓN, no por-conversación

12 campos en `sessionRuntime` (`session_service.go:85-109`: `live`, `curRun`, `resumeRetried`,
`pendingPerm`, `grants`…) y 6 mapas en el FE indexados por `session.id`. Toca el boundary
`sesion-viva-consistente` (**enforced, high**) y sus 4 tests. Ninguna decisión lo cubre porque es
diseño, no producto — **lo resuelve la arquitectura**, y es el tramo más denso del paquete.

---

## CV-D16 · Las sesiones vivas con llave vieja se RE-KEY, no se borran 🧑‍⚖️ FIRMADA 2026-07-26

CV-D6 borra las 3 cerradas por no importantes. Las **vivas** con id pelado son otra cosa: son
conversaciones que el operador sí quiere, y 3 quedan invisibles cuando la UI pregunta por clave
calificada. Estado verificado contra el daemon vivo:

| sesión | `arnes` en disco | |
|---|---|---|
| `s25123a2c` | `vitalia` | pelado |
| `s6165ac75` | `sin-home~vitalia~vitalia` | ok |
| `s0fec7798` | `vitalia` | pelado |
| `sfc512b15` | `vitalia` | pelado |
| `s78b3aeeb` | `arnesia` | pelado |

**Decisión:** el mismo paso que estrena el esquema versionado (problema B de la arquitectura)
**re-key las vivas** a clave calificada `(home,id,scope)`. Nada se borra. **Copia previa
obligatoria** (`~/.arnesia/sessions.json.bak-<sello>`) y el paso tiene que ser **reversible**.

Sin esto, CV-D4 («solo las conversaciones de mi sesión actual») seguiría mintiendo para esas 3.

Alternativa **descartada** por el operador: borrarlas como a las cerradas. Habría ahorrado el
migrador de llaves a costa de perder transcripts vivos.

---

## CV-D17 · La mudanza del picker se adelanta del tramo 5 al 3 — no se puede no hacerla 🧑‍⚖️ PENDIENTE

**Qué pasó.** T21 retira `conversacionesDeArnes` e `historialCerrada` del cliente, porque sus dos
rutas ya no existen (responden 400 con puntero y 404). Su único consumidor —el store
`widgets/session-rail/model/conversaciones-store.ts` y el bloque `ConversacionesDelArnes` del
selector de arnés— **deja de compilar en el mismo instante**. Eso es RF-333/RF-334, que el plan
había puesto en **T30, tramo 5**.

**Por qué el orden del plan ya no aplica.** T30 estaba último con un argumento explícito: «hasta acá
el operador conserva la superficie vieja, aunque muestre 0». El argumento **caducó con el tramo 2**:
esa superficie no muestra 0 hoy, muestra un error de red, porque los endpoints que consulta se
retiraron. Conservarla no era conservar nada — era garantizar un fallo a la vista del operador.

**Las alternativas, y por qué no.**
- *Dejar los dos métodos del cliente hasta T30.* Habría que darles un tipo propio, porque `Session`
  cambió de forma: escribir un contrato TypeScript que describa endpoints que ya no se sirven.
- *Adaptar el store del rail a la forma nueva.* Adaptar código que el mismo paquete va a borrar,
  para que siga llamando a rutas muertas.

**Qué se hizo.** El store y su test se eliminan; el picker pierde el import, la llamada, el render y
el componente. La story `CasoSimple` gana un assert **por ausencia**
(`queryByText(/Conversaciones:/) === null`), que es la evidencia de RF-333.

**Qué NO cambia.** El resto de T30 —el gate humano de la mudanza y la nota de `mockups/INDEX.md`—
sigue en el tramo 5. Y la **capacidad** de leer lo archivado antes de la migración sigue viva del
lado del daemon (CAP-98), ahora sin superficie que la ejerza: declarado en su `change_log`, no
tapado.

**Efecto colateral, verificado:** las 4 stories del picker que el `BACKLOG` daba por rotas por el
`text-warn` de `new-session-picker.tsx:273` ya no lo ejercitan. **La deuda del token `--warn` sigue
abierta** — sólo perdió a este consumidor.

---

## CV-D18 · La recalibración de llaves se CABLEA al arranque del daemon 🧑‍⚖️ FIRMADA 2026-07-27

**Qué se decide.** El daemon, al arrancar, **detecta las llaves a medias y las recalibra solo**,
con **respaldo previo** (`sessions.json.bak-<sello>`) y **una línea de log que diga cuántas
recalibró**. Hoy `cmd/arnesia/main.go:202` pasa `nil` como `ClaveCalificada` a `AbrirRegistro`, así
que `reKey` no corre nunca en el arranque y la decisión firmada no está construida.

**Qué revierte.** La **enmienda que `arquitectura.md` le había hecho a CV-D16** dejándola como
comando manual (`arnesia sesiones recalibrar-llaves`). El argumento del arquitecto era bueno —no
atar el arranque del daemon a que el Portafolio responda— pero la enmienda **la tomó un documento
que el operador nunca leyó**. Es el riesgo exacto que abrió el Gate 2 «por directiva, no por
lectura», y se resolvió de la única forma que corresponde: **preguntándole**.

**Motivo.** CV-D16 firmada dice que **el mismo paso** que estrena el esquema versionado re-key las
vivas. Con el comando manual, **3 de las 5 sesiones del operador quedan invisibles sin que nada se
lo avise** (hallazgo **N-23** del E2E): nada en la superficie le sugiere correr el comando, así que
el efecto práctico es que la decisión no existe. **Una decisión firmada que no corre no está
construida.**

**Alternativas descartadas por el operador.**
- **(a) Avisar en la superficie, con un botón para recalibrar.** Más honesto con «no mutar datos al
  arrancar», pero exige **superficie nueva fuera del mockup firmado** — y el mockup está firmado.
- **(b) Dejarlo manual como está.** Es el estado que produjo N-23.

**Contrapeso declarado, no escondido.** Esto **muta datos del operador al arrancar**, que es justo
lo que el arquitecto quiso evitar. Se acepta, y por eso **el respaldo previo y el log no son
opcionales**, y **la reversión tiene que seguir funcionando**:

1. **Primero la red, después el cambio.** Si el respaldo falla, **no se recalibra nada**. El orden
   no es un detalle de implementación: es la decisión.
2. **Nada en silencio.** El log dice **cuántas** se recalibraron y **cuántas quedaron
   `sin-candidata`**. Lo segundo es tan obligatorio como lo primero: el E2E ya sabe que **una de
   las 5 sale `sin-candidata`** porque su arnés no está en el Portafolio. Eso es correcto y
   honesto — callarlo lo volvería un defecto mudo.
3. **Idempotente.** Arrancar dos veces no duplica respaldos ni vuelve a mover nada.
4. **El motivo del riesgo queda escrito EN EL CÓDIGO**, no sólo acá.

**Consecuencia sobre A-10.** El hallazgo A-10 de la auditoría —«CV-D16 no corre en el binario que
se instala, y eso lo decidió `arquitectura.md`, no el operador»— queda **cerrado por decisión del
operador**, no por argumento técnico del ejecutor. El comando manual **no se retira**: sigue siendo
útil con `--dry-run` y como reversión.

## Abierto

**CV-D17 sigue PENDIENTE de firma.** CV-D1..CV-D16 y **CV-D18** firmadas; F-1..F-4 son
correcciones de hecho, no decisiones nuevas.
