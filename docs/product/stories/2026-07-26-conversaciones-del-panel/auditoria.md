# Auditoría independiente — `2026-07-26-conversaciones-del-panel`

> `tipo: auditoría` · worktree `.claude/worktrees/audit`, rama `audit-conversaciones`, base `6d85861`
> (la punta de los 6 tramos). Todas las cifras de este informe las **generó el auditor**, corriendo
> los comandos en este worktree; ninguna se copió de `PARIDAD.md` ni de `verificacion-e2e/INFORME.md`.
> `~/.arnesia/` NO se tocó: `sessions.json` md5 `b1689d1513e8d3c3fa21692c511ed793` al abrir y al
> cerrar. **T32 NO se ejecutó.**
>
> Los hallazgos declarados (N-1, N-5…N-23) **no son el entregable**. Lo que sigue es lo que nadie
> vio. Se numeran `A-n` para no colisionar con la serie `N-n` del paquete.

---

## 1 · Veredicto

**No, todavía no. Falta una tanda de código, no una firma.**

1. **La lista de conversaciones —el entregable entero del paquete— no se puede usar para lo que
   existe**: el título nunca se auto-deriva y la fecha de última interacción nunca se escribe. Las
   dos son decisiones **firmadas** (CV-D9, CV-D13) y las dos están **sin construir** (A-1, A-2).
2. Consecuencia concreta, reproducida contra el servicio real: después de trabajar un rato, el
   operador abre el panel y ve **N filas idénticas**, todas `nueva conversación`, todas
   `sin fecha · N turnos · ctx X %`. No puede distinguir una de otra, que es exactamente el
   problema que originó el paquete.
3. **No hay riesgo de pérdida de datos** en la migración v1→v2 del registro real del operador
   (verificado campo por campo sobre su archivo, sin excepciones) — el cimiento está sano.
4. Lo demás está por encima de la media del repo: build, tests, contrato, boundaries y capabilities
   son sólidos y las cifras de conformance que el paquete declara son **exactas**.
5. **Condición para firmar PARIDAD:** cerrar A-1 y A-2 (son chicos: dos escrituras en `Turn`), y
   decidir A-3 y A-6, que son decisiones del operador tomadas en su nombre bajo el Gate 2.

---

## 2 · Hallazgos nuevos

### 🔴 A-1 · El título de la conversación **nunca** se auto-deriva — CV-D9 / RF-303 sin construir

**Qué.** `domain.Conversacion.DerivarTitulo` (`internal/domain/conversacion.go:243`) tiene **cero
llamadores de producción**. Sus únicos llamadores están en
`internal/domain/conversacion_test.go:379,393,403`. En el camino real del turno
(`internal/usecase/session_service.go:491-496`) lo que se deriva es el **`Frente` de la SESIÓN** —
el comportamiento anterior al paquete, intacto — y el turno se appendea a `conv.Conv` sin tocar el
título de la conversación.

```
grep -rn "DerivarTitulo" --include=*.go .   # sólo conversacion.go y su _test.go
```

**Cómo se reproduce.** Test contra el servicio real (corrido por el auditor, con el harness
`newSvc`/`stubAgent` del propio paquete; el archivo temporal se borró tras medir):

```
FRENTE de la sesión       = "ciclo full-cycle · spec→released"
TITULO de la conversación = "nueva conversación"      ← RF-303 ROTO
ULTIMA_INTERACCION        = ""                        ← RF-304 ROTO (A-2)
turnos                    = 2
```

**Impacto.** Toda conversación que el operador cree se llama `nueva conversación` para siempre,
salvo que la renombre a mano una por una. La lista deja de ser una lista: es N filas homónimas.
CV-D9 dice literalmente «título auto-derivado del primer mensaje, **editable**» — se construyó sólo
la mitad editable.

**Arreglo.** En `session_service.go`, junto a la derivación del `Frente` (L491-492), llamar
`conv.DerivarTitulo(deriveFrente(text))` sobre la conversación activa. La ley de precedencia
(«gana el operador si ya editó») ya está escrita y probada en el dominio. Ticket + capability +
test que falle sin la llamada.

---

### 🔴 A-2 · `UltimaInteraccion` **nunca** se escribe — CV-D13 / RF-304 sin construir

**Qué.** La única asignación de `UltimaInteraccion` en todo el código de producción es la que la
pone **vacía**: `internal/adapters/store/migracion.go:114` (`UltimaInteraccion: ""`, correcta, es la
migración que no inventa fechas). El campo se **declara** (`internal/domain/conversacion.go:64`), se
**proyecta** al DTO (`internal/usecase/session_conversaciones.go:83`) y se **lee** para desempatar
(`conversacion.go:341-342`), pero nadie lo estampa nunca.

Esto es exactamente lo que CV-D13 marcó como el requisito nuevo del paquete: «hace falta un
timestamp estampado en **cada turno**, no al desactivar — si no, una conversación inactiva mentiría
la fecha del último mensaje».

**Cómo se reproduce.** Ver A-1 (misma corrida): `ULTIMA_INTERACCION = ""` después de un turno
completo con `result`.

**Impacto — triple.**
1. **La fila miente por omisión.** `web/src/widgets/chat-dock/ui/conversacion-fila.tsx:107` pinta
   `sin fecha` cuando el campo está vacío. Está bien programado; el problema es que ese camino
   **no es el borde, es el único**. Toda fila con turnos dice `sin fecha · N turnos · ctx X %`.
2. **El orden se degrada en silencio.** RF-320 CA-1 ordena por última interacción y cae al desempate
   `creada_en` (`conversacion.go:345`), o sea orden de creación. Nadie lo nota: no hay error, hay
   un orden plausible y equivocado.
3. **La reparación de la invariante elige mal.** `NormalizarConversaciones` activa «la de
   `UltimaInteraccion` más reciente» (`conversacion.go:282`, `masReciente()` L336-354). Con el campo
   siempre vacío, la primera rama del `switch` nunca se toma y gana la de `CreadaEn` más reciente —
   la más nueva, no la última usada.

**Y la doctrina lo cementa mal.** La capability nueva
`docs/product/capabilities/fe-chat/panel-de-conversaciones.yaml:97` declara como regla: «con turnos
pero sin fecha registrada (**una conversación migrada**) se dice “sin fecha”». Presenta como
excepción de migración lo que en el producto es el caso general. El SSoT funcional describe un
sistema que no es el que hay.

**Arreglo.** Estampar `conv.UltimaInteraccion = time.Now().UTC().Format(time.RFC3339)` en el mismo
lugar donde se appendea el turno del usuario (`session_service.go:496`) y, si se quiere fidelidad
al literal «última interacción», también al cerrar el turno del asistente. Corregir la redacción de
`panel-de-conversaciones.yaml:97` para que «sin fecha» vuelva a ser el borde que dice ser.

> **A-1 + A-2 juntos son el hallazgo del informe.** Por separado son dos campos; juntos son la
> lista entera. Y explican por qué se escaparon: `DerivarTitulo` **tiene** tests unitarios verdes
> (sobre código muerto), las stories del panel **hardcodean** títulos y fechas plausibles, y el E2E
> midió una migración —donde `sin fecha` es correcto— en vez de una conversación nacida en la app.
> Tres capas de verificación verde sobre una función que nadie llama.

---

### 🔴 A-3 · La migración puede perder `Checkpoint` y `CadenaCC` sin que ningún test se entere

**Qué.** `TestDeV1aV2ConservaTodoElFixtureReal`
(`internal/adapters/store/migracion_test.go`, se presenta como «el corazón del ticket. Se compara
**CAMPO POR CAMPO**») corre sobre `testdata/sessions-v1-real.json`, que es —verificado— el registro
**real** del operador, byte a byte idéntico (md5 `b1689d15…`, 17 202 B, 5 sesiones). Eso es buena
práctica. El problema es que **ese fixture no contiene 5 de los 17 campos de `sesionV1`**:
`cadena_cc`, `cerrada_en`, `checkpoint`, `parked`, `puesto` no están en el archivo. Un test
campo-por-campo no puede comparar un campo que el insumo no trae.

**Cómo se reproduce** (mutación, con el árbol restaurado después):

| mutación en `internal/adapters/store/migracion.go` | `go test ./...` |
|---|---|
| `Checkpoint: v.Checkpoint,` → `Checkpoint: "",` | 🔴 **verde** |
| `CadenaCC: v.CadenaCC,` → `CadenaCC: nil,` | 🔴 **verde** |
| `CtxHist: v.CtxHist,` → `nil` | ✅ rojo |
| `RotacionPendiente:` → `false` | ✅ rojo |
| `Cwd: v.Cwd,` → `""` | ✅ rojo |

**Impacto.** Son precisamente los dos campos que la **enmienda F-3** volvió portantes: CV-D11
promete retomar «con `CadenaCC`/`Checkpoint`/`Cwd` intactos», y F-3 extendió CV-D8 al `Checkpoint`
con el argumento de que sin él «retomar una conversación rotada arranca sin el digest de su propia
rotación». Hoy la migración los lleva bien — pero no hay red: cualquier refactor futuro los tira en
silencio y el árbol queda verde. Para el operador el síntoma sería mudo: una conversación rotada que
al retomarse arranca amnésica.

**Arreglo.** Un fixture sintético adicional (no reemplazar el real: su valor es otro) con los 17
campos poblados, y afirmar la igualdad campo por campo sobre él. El fixture real se queda como
prueba de que el archivo del operador sobrevive.

---

### 🔴 A-4 · El foco inicial del panel se pierde: se enfoca antes de que exista qué enfocar

**Qué.** `web/src/widgets/chat-dock/ui/conversaciones-panel.tsx:66-71` decide el foco en un
`useEffect(..., [])` que corre **una sola vez, al montar**:

```tsx
useEffect(() => {
  if (focoInicial === "buscador") buscadorRef.current?.focus()
  else listaRef.current?.focus()
}, [])
```

Pero al abrir, el store hace `set({ ...INICIAL, ... })`
(`web/src/widgets/chat-dock/model/conversaciones-store.ts:134`) y `INICIAL` deja el estado en
`cargando`; el render de ese primer frame devuelve `<Esqueleto />`
(`conversaciones-panel.tsx:80`) — **no existen todavía ni el `listbox` ni el `<input>`**. Los dos
`ref` valen `null`, el `?.focus()` es un no-op silencioso y el efecto **no vuelve a correr** (el
`biome-ignore` de la línea 67 lo declara deliberado: «el foco se decide UNA vez, al abrir»).

**Cómo se reproduce.** Abrir el panel con el `▶` o con el `🔍`. El foco se queda donde estaba (el
botón que se apretó). Las flechas no mueven el cursor de la lista y escribir no va al buscador
hasta tabular a mano.

**Impacto.** El camino de teclado del panel no arranca. Para un panel cuya navegación por teclado
está bien construida (`role="listbox"` + `aria-activedescendant` + flechas + Enter, L115-130), el
único paso que falta es el primero — y sin él, nada de lo demás se alcanza sin mouse.

**Arreglo.** Disparar el foco cuando el contenido aparece, no al montar: depender del estado
(`[estado]`) y enfocar al pasar a `listo`, o usar un callback-ref. El intento del `biome-ignore`
(no re-enfocar en cada render) se conserva con un flag de «ya enfoqué una vez».

---

### 🔴 A-5 · El cursor de la lista no se hace visible al moverse con flechas

**Qué.** No hay ni un `scrollIntoView` en el panel de conversaciones. El único del widget es el del
transcript (`web/src/widgets/chat-dock/ui/chat-dock.tsx:175`). El contenedor del listbox es
`overflow-y-auto` (`conversaciones-panel.tsx:131`) y el cursor se mueve sólo por estado
(`setCursor`, L121-123), sin mover el scroll.

**Cómo se reproduce.** Con más conversaciones de las que entran en el alto del panel, poner foco en
la lista y bajar con `↓`. El cursor visual y `aria-activedescendant` avanzan; la vista no. A partir
de la primera fila fuera del viewport, el operador navega a ciegas.

**Impacto.** Se agrava con A-4 (hay que llegar al foco a mano) y es justo el escenario que el
paquete midió como realista: 50 conversaciones. Para lector de pantalla es menos grave —
`aria-activedescendant` se anuncia igual— pero para teclado con vista es navegación ciega.

**Arreglo.** `useEffect` sobre `cursor` que llame `scrollIntoView({ block: "nearest" })` sobre el
elemento `${id}-${marcada.id}`.

---

### 🔴 A-6 · `Escape` no cierra el panel

**Qué.** Sólo hay dos manejadores de `Escape` en el widget: el del renombre
(`web/src/widgets/chat-dock/ui/conversacion-row.tsx:104`) y el del `<input>` del buscador
(`conversaciones-panel.tsx:184`). El `onKeyDown` del listbox
(`conversaciones-panel.tsx:119-130`) sólo atiende `ArrowDown`, `ArrowUp`, `Enter` y `Space`. El
contenedor del panel no tiene manejador.

**Cómo se reproduce.** Abrir el panel, poner foco en la lista, apretar `Escape`. No pasa nada.

**Impacto.** El panel abre **en sitio** y tapa el transcript (es la decisión C-4, correcta: no es un
`<dialog>`, no hay backdrop). Pero un panel que tapa y no responde a `Escape` obliga a volver al
mouse o a tabular hasta el `▶` para cerrarlo. Es la convención más universal de la plataforma y es
la única salida barata.

**Arreglo.** Manejar `Escape` en el contenedor del panel → `onCerrar()`, y devolver el foco al
control que lo abrió (el `▶` o el `🔍`), que hoy tampoco se restituye.

---

## 3 · Hallazgos 🟠

### 🟠 A-7 · `convActiva` —la razón de ser de `sesion-viva-consistente` v1.2— no está enforzado

**Qué.** El nodo se extendió a v1.2 con el argumento de que «el runtime sigue siendo uno por sesión
y **gana un puntero que dice a quién le pertenece**»
(`internal/usecase/session_service.go:124-129`). El quinto check nuevo,
`transicion-de-conversacion-atomica`, se declara enforzado por
`TestTransicionDeConversacionEsAtomica` (`docs/architecture/fitness/arch_test.go:2074`).

Mutación: borrar la asignación del puntero en la transición
(`internal/usecase/session_conversaciones.go:336`, `r.convActiva = nueva.ID` → `_ = nueva.ID`) deja
**verde** el paquete `./docs/architecture/fitness/` — el enforcer pasa igual. También queda verde
`./internal/usecase/`. El mecanismo que el nodo v1.2 agregó no lo prueba nadie; con él roto, el
detector `chequearDueño` (L158-164) nunca avisaría de la divergencia que existe para detectar.

Es N-12 otra vez, en forma más sutil: allá el enforcer no podía correr; acá corre y pasa sin tocar
lo suyo.

**Arreglo.** Agregar al enforcer una aserción sobre el puntero (o sobre el `slog.Warn` de
`chequearDueño`) que falle sin la línea 336. Mientras tanto, el changelog del nodo debería decir
«5 de 5 checks con enforcer, 4 con enforcer que prueba lo suyo».

### 🟠 A-8 · `respaldarSiEsViejoLocked` —la red que impide pisar el archivo v1— no tiene test

**Qué.** `internal/adapters/store/registry.go:157-159` respalda antes de escribir, con un comentario
que dice exactamente cuánto vale: «El respaldo va acá y no en el llamador para que no haya un camino
de escritura que se lo saltee — **el que se lo saltea es el que borra el archivo del operador**».

Mutación: reemplazar ese bloque por `if false { return nil }` deja **verde**
`./internal/adapters/store/` y **verde** `./docs/architecture/fitness/`. (El check
`migracion-forward-only-con-respaldo` sí se cae si se saca el respaldo del camino de
`AbrirRegistro` — pero no si se saca el de `Save`, que es el que cubre todos los demás caminos de
escritura.)

**Impacto.** El guard más caro del paquete —el único entre el operador y la pérdida de su registro
por un camino de escritura no previsto— es la única línea sin red. Hoy funciona; nada avisa si
mañana no.

**Arreglo.** Test: registro v1 en disco + `Save()` directo (sin `AbrirRegistro`) ⇒ existe
`<ruta>.v1-<sello>.bak` con el contenido original.

### 🟠 A-9 · `sesiones-cerradas.json` se abandona en silencio, y el fallback que lo justifica quedó inalcanzable

**Qué.** `cmd/arnesia/main.go:873` cambia el archivo de archivadas a `sesiones-archivadas.json`, y
el comentario (L864-866) declara: «El `sesiones-cerradas.json` viejo **NO se lee ni se migra** — el
operador lo borra a mano (CV-D6)».

Para las 3 sesiones que el operador declaró no importantes, eso es CV-D6 y está bien. **Los dos
problemas son otros:**

1. **Es la única cosa que este paquete hace en silencio.** Todo el resto del arranque pasa por
   `store.Informe` + `loguearInforme` con su lema explícito («nada de esto ocurre en silencio»):
   migración, respaldo, cuarentena, esquema futuro, reparaciones, recalibraciones. El abandono de un
   archivo de datos del operador **no emite una sola línea**. Y no aplica sólo a las 3 conocidas:
   **cualquier sesión que el operador archive con el binario que tiene instalado hoy, antes de
   actualizar, cae en el archivo abandonado** sin que nada se lo diga.
2. **Deja un fallback inalcanzable.** `internal/usecase/session_historial.go:110-112` documenta el
   camino JSONL como «FALLBACK para las sesiones archivadas **ANTES de la migración**, cuyo
   transcript no se guardó». Esas sesiones viven en el archivo que ya no se lee. Todo lo que hay en
   `sesiones-archivadas.json` lo escribió el binario nuevo, que siempre conserva `Conv`. El caso que
   justifica el código no puede ocurrir por ese camino.

**Arreglo.** Detectar el archivo viejo al arrancar y emitir **una línea** en el `Informe`
(«existe `sesiones-cerradas.json` con N sesiones; este binario ya no lo lee — ver T32»). Es
consistente con el resto del paquete y cuesta diez líneas. Y corregir el comentario del fallback
para que describa el caso que sí puede darse (una sesión archivada sin turnos).

### 🟠 A-10 · CV-D16 no corre en el binario que se instala — y eso lo decidió `arquitectura.md`, no el operador

**Qué.** `cmd/arnesia/main.go:202` pasa **`nil`** como `ClaveCalificada` a `AbrirRegistro`, con lo
cual `reKey` nunca corre en el arranque (`internal/adapters/store/migracion.go:224`,
`if inf.Migro && clave != nil`). El motivo está escrito y es defendible (L197-200: no atar el
arranque del daemon a que el Portafolio responda).

**Por qué lo levanto igual.** CV-D16 está **firmada 🧑‍⚖️** y dice: «**el mismo paso** que estrena el
esquema versionado **re-key las vivas**». Lo que se construyó es un comando manual
(`arnesia sesiones recalibrar-llaves`) que **nada le sugiere al operador correr**. El hecho está
declarado como **N-23**; lo que agrego es su naturaleza: no es una deuda, es **una decisión firmada
cambiada por un documento que el operador nunca leyó**, que es precisamente el riesgo que el Gate 2
«por directiva, no por lectura» abrió. Debe presentarse al operador como enmienda a CV-D16, no como
hallazgo técnico.

**Arreglo.** O bien cablear el re-key después del Portafolio (y persistir), o bien —lo mínimo—
emitir en el `Informe` de arranque «N sesiones con llave a medias; corré
`arnesia sesiones recalibrar-llaves --dry-run`». Y llevar la enmienda a `decisiones.md`.

### 🟠 A-11 · Dos guards de la transición sin ningún test que los sostenga

Mutaciones que sobreviven — ningún test se pone rojo:

| mutación en `internal/usecase/session_conversaciones.go` | qué deja de valer | alcance medido |
|---|---|---|
| `r.pendingPerm = nil` (L317) → no-op | las tarjetas de permiso del hilo viejo sobreviven en el runtime del hilo nuevo | `./internal/usecase/` **y** `./docs/architecture/fitness/` |
| `r.assembling.Reset()` (L330) → no-op | el buffer de ensamblado a medias del hilo viejo se hereda | **`./...` (árbol entero)** |

Contraste honesto, para que se vea que el enforcer **sí** sirve: `r.grants = nil` (L331) y
`r.live = nil` (L338) **sí** ponen rojo a `TestTransicionDeConversacionEsAtomica`. El check cubre
lo que enuncia en sus puntos (2) y (4); los otros dos resets del paso 5 no los cubre nadie.

**Impacto.** Bajo, porque hoy el código es correcto — pero el paso 5 se documenta como una unidad
(«el estado de vuelo se resetea») y sólo la mitad tiene red.

**Arreglo.** Extender el enforcer: tras la transición, `pendingPerm` vacío y `assembling` vacío.

---

## 4 · Hallazgos 🟡

### 🟡 A-12 · Las dos cifras del gate de PARIDAD están tecleadas, y las dos están mal

Contadas por el auditor sobre las tablas de `PARIDAD.md` (script, no a ojo):

| tabla | declara | contado | delta |
|---|---|---|---|
| §2 (mockup vs producto) | **30 ✅ · 5 ⚠️ · 0 ❌** | **31 ✅ · 4 ⚠️ · 0 ❌** | +1 ✅ / −1 ⚠️ |
| §2.1 (sólo la app instalada) | **17 ✅ · 1 ⚠️ · 2 ❌** | **18 ✅ · 1 ⚠️ · 2 ❌** | +1 ✅ |

La causa probable de §2 es benigna y hasta buena noticia: N-1 se cerró en T23 y su fila pasó de ⚠️ a
✅ sin que el resumen se recontara (el párrafo que sigue todavía enumera «las cinco desviaciones»).
Lo levanto igual porque la regla dura del repo es «**las cifras se GENERAN, no se teclean**», y son
las dos cifras que el operador va a leer para decidir si firma. Un gate humano no debería apoyarse
en un conteo a mano.

**Arreglo.** Que `estado.sh` (o un script del paquete) cuente las marcas de las tablas de PARIDAD,
como ya se hace con el resto del checkpoint.

### 🟡 A-13 · `RecorteDeTitulo` parte UTF-8 a la mitad — preexistente, pero el paquete le suma consumidores

**Qué.** `internal/domain/conversacion.go:102-108` compara **bytes** (`len(text) > 48`) y corta
**bytes** (`text[:48]`). Con acentos —o sea, en castellano— el corte puede caer dentro de una runa.

**Reproducido** (test temporal, borrado después):

```
in=108 bytes out="aaaaaaa…aaa\xc3…" valid=false   ← UTF-8 inválido
```

`encoding/json` reemplaza el byte huérfano por `U+FFFD`, así que el operador ve un `�` al final del
título, y eso se **persiste** en `sessions.json`.

**Honestidad sobre el origen:** el bug **es preexistente** — venía de `deriveFrente`
(`main:internal/usecase/session_service.go:918`) y el paquete lo movió al dominio tal cual, sin
empeorarlo. Lo que sí hizo el paquete es **duplicarle los consumidores**: ahora también bautiza las
conversaciones migradas (`migracion.go:141`) y —en cuanto se arregle A-1— el título de toda
conversación nueva. El mismo patrón está en `CheckpointMecanico`
(`internal/usecase/session_rotacion.go:39-40`, `texto[:700]`), también preexistente, y ahí el byte
roto viaja al conductor.

**Arreglo.** Cortar por runas (`[]rune(text)[:48]`) en los dos lugares. Es de una línea cada uno.
No lo apliqué: la regla del encargo es encontrar, no parchear, y esto merece su ticket.

---

## 5 · Verificación de las cifras ajenas

Todo corrido por el auditor en este worktree, con `pnpm install --frozen-lockfile` propio (no el
symlink de `node_modules`).

| afirmado | por | medido por el auditor | veredicto |
|---|---|---|---|
| `conformance --todo` → `323 · pass 91 · fail 0 · error 0 · deferred 232` | `INDEX.md`, `PARIDAD.md` §1.14 | **`323 checks · pass 91 · fail 0 · error 0 · deferred 232 · n/a 0`** | ✅ **exacto** |
| `go build ./...` verde | tramos | verde | ✅ |
| `go vet ./...` verde | tramos | verde | ✅ |
| `go test ./... -race` verde | tramos | **27 paquetes ok · 0 FAIL · 7 sin tests** | ✅ |
| `tsc --noEmit` → 0 errores | §1.14 | **0 errores** | ✅ |
| `biome ci` limpio | §1.14 | **178 archivos, 0 errores** (1 info: sugerencia `biome migrate`) | ✅ |
| `steiger` sin problemas | §1.14 | **`✔ No problems found!`** | ✅ |
| `depcruise` sin violaciones | §1.14 | **`✔ no dependency violations found (194 módulos, 601 dependencias)`** | ✅ |
| `vitest --project=unit` | §1.14 | **9 archivos · 136 tests · 136 pass** | ✅ |
| `vitest --project=storybook` | §1.14 | **47 archivos · 444 tests · 444 pass** (con el error de consola esperado de la story `FallaUnaVez`) | ✅ |
| `estado.sh --check` en sync | Paquete A de N2 | **«cifras en sync con el estado real ✓», exit 0** | ✅ |
| PARIDAD §2 = 30 ✅ · 5 ⚠️ | `PARIDAD.md` | **31 ✅ · 4 ⚠️** | ❌ **A-12** |
| PARIDAD §2.1 = 17 ✅ · 1 ⚠️ · 2 ❌ | `PARIDAD.md` | **18 ✅ · 1 ⚠️ · 2 ❌** | ❌ **A-12** |
| el fixture de migración es el registro real del operador | T9 | **md5 `b1689d15…` idéntico, 17 202 B, 5 sesiones, mismos ids y mismos conteos de turnos** | ✅ **confirmado** |
| `go-arch-lint` no se corrió | `INDEX.md` | tampoco pude correrlo (ver §8) | ✅ declarado honesto |

**Lo que la tabla no dice y hay que decir:** las cifras del toolchain son verdes y son propias, pero
**siguen siendo locales**. Nadie pusheó; el gate de T6 (CI 3/3) sigue abierto y ninguna de estas
corridas lo sustituye.

---

## 6 · Cobertura real — verificado vs. declarado

### 6.1 · Lo que sí está sólido (una línea, como manda el encargo)

Migración v1→v2 sobre el archivo real del operador: **verificada campo por campo por el auditor**.
Los 15 campos que su `sessions.json` realmente contiene tienen destino explícito en v2 (10 a
`Session`, 5 a `Conversacion`); **ninguno se pierde**. Archivado, rotación (checkpoint, cadena,
`turno_idx`), cuarentena, esquema futuro y búsqueda tienen tests que **se caen** cuando se rompe el
código (mutación verificada, 5 de 5 rojos). Capabilities: 18 hojas tocadas, 7 nuevas — el paquete
cumple `codigo-traza-a-capability` con holgura.

### 6.2 · El agujero de cobertura, por camino crítico

Mutación aplicada al código de producción y **árbol restaurado en cada caso** (verificado: `git
status` limpio y `go build ./...` verde al terminar). Alcance de la corrida por fila: las de
migración/`Save` se midieron contra `./internal/adapters/store/`, las cinco marcadas «árbol entero»
contra `./...`, y las de la transición contra `./internal/usecase/` **y** `./docs/architecture/fitness/`
—que es donde vive el enforcer— para no acusar de decoración a un test que sí existe en otro paquete.

| camino crítico | mutación | resultado |
|---|---|---|
| transición · rollback del estado de vuelo | quitar `restaurarVuelo` | ✅ rojo |
| transición · deny de permisos con motivo | no emitir los frames | ✅ rojo |
| transición · descarte de grants | `r.grants = nil` → no-op | ✅ rojo (fitness) |
| transición · soltar el conductor viejo | `r.live = nil` → no-op | ✅ rojo (fitness) |
| **transición · puntero `convActiva`** | `r.convActiva = nueva.ID` → no-op | 🔴 **verde** (A-7) |
| **transición · limpiar `pendingPerm`** | → no-op | 🔴 **verde** (A-11) |
| **transición · limpiar el ensamblado** | → no-op | 🔴 **verde** (A-11) |
| búsqueda · el total es el total | `total := 0` | ✅ rojo |
| migración · `CtxHist` / `RotacionPendiente` / `Cwd` | tirar el campo | ✅ rojo (×3) |
| **migración · `Checkpoint`** | tirar el campo | 🔴 **verde** (A-3) |
| **migración · `CadenaCC`** | tirar el campo | 🔴 **verde** (A-3) |
| migración · respaldo previo en `AbrirRegistro` | quitarlo | ✅ rojo |
| **`Save` · respaldo del v1 en disco** | `if false` | 🔴 **verde** (A-8) |
| `Save` · no pisar un esquema futuro | quitar el guard | ✅ rojo |
| archivado · conservar las conversaciones | `nil` | ✅ rojo |
| archivado · devolver el transcript | no appendear | ✅ rojo |
| rotación · escribir el checkpoint | `""` | ✅ rojo |
| rotación · encadenar el cc-id | no-op | ✅ rojo |
| rotación · `turno_idx` de la marca | `idx := 0` | ✅ rojo |

**19 mutaciones · 14 rojas (74 %) · 5 verdes.** Para un paquete de este tamaño es una tasa buena;
las 5 supervivientes están todas concentradas en dos zonas (el paso 5 de la transición y los campos
que el fixture real no tiene), y ninguna es casual.

### 6.3 · RF y escenarios — declarado con un límite

**No completé el barrido uno-por-uno de los 58 RF ni de los 50 escenarios E-01…E-50.** Lo que sí
tengo, verificado en el código y por reproducción, son **cuatro RF que NO están como se
especificaron**, y son los que importan:

| RF | estado | evidencia |
|---|---|---|
| **RF-303** (título auto-derivado) | **AUSENTE** | `DerivarTitulo` sin llamadores de producción — A-1 |
| **RF-304** (timestamp por turno) | **AUSENTE** | `UltimaInteraccion` nunca se escribe — A-2 |
| **RF-320 CA-1** (orden por última interacción) | **DISTINTO** | cae siempre al desempate `creada_en` — A-2 |
| **RF-333/334** (mudanza del picker) | implementado, **adelantado de tramo** | CV-D17, sin firmar |

Sobre CV-D12 (vocabulario «activa/inactiva», nunca «cerrada»): el E2E verificó **0 ocurrencias de
«cerrada» en el copy del bundle servido** y eso lo doy por bueno. Queda legado **intencional y
declarado** del lado del daemon (`CerradaEn`, `SetArchivoCerradas`, `Cerradas`,
`HistorialCerrada`), que describe sesiones archivadas, no conversaciones — es otro sujeto y el
propio código lo argumenta. **No completé el barrido de la superficie de la API ni de los YAML de
capabilities**, así que no puedo afirmar que no quede ninguno visible por ahí.

Distinción honesta que el encargo pide y que sí puedo dar: **las tres capas de verificación del
paquete (tests unitarios, stories, E2E) fueron verdes sobre A-1 y A-2**. Eso significa que
«cubierto por un test» y «hace lo que dice» se separaron al menos una vez de forma grave, y que las
tablas de cobertura E-01…E-50 de `plan-pruebas.md` deben leerse con esa desconfianza hasta que
alguien las recorra una por una.

---

## 7 · UX

**Contra el mockup firmado, el producto es fiel.** Las 8 desviaciones declaradas (§2 y §2.1) las
revisé y **7 están bien justificadas**: la del `✎` es superset autorizado (C-11), las de contraste
son el precedente ya firmado del repo (C-3/C-13), la del texto de la marca de rotación es la
contradicción C-5 resuelta a favor de no reescribir transcripts ya persistidos —el argumento es
correcto—, y las dos ❌ de §2.1 (N-21, N-22) están declaradas como lo que son. La octava, **N-19**
(el `＋` bloqueado no ofrece `(■ para interrumpir)`), no es una racionalización: es una decisión de
contenido que el paquete dejó explícitamente esperando al operador. Está bien dejada abierta.

**Lo que el mockup no podía revelar son A-4, A-5 y A-6** (§2): el panel está bien dibujado y bien
construido en su estructura ARIA —`role="listbox"`, `aria-activedescendant`, flechas, Enter, el
vacío que nombra el término buscado, el esqueleto que calca el contrato ARIA del picker— y sin
embargo **el camino de teclado no arranca, no scrollea y no se sale con `Escape`**. Es un caso
clásico: la accesibilidad estática está, la dinámica no. Ninguna story lo caza porque las stories
montan el componente ya con datos, saltándose el frame de `cargando` que es el que rompe el foco.

**La pregunta que resume todo — el operador abre esto sin haber leído nada:**

- **¿Entiende qué está viendo?** Sí. Dos filas de cromo, el `▶ título`, la lista en sitio. El
  recorte de CV-D14 se nota y es una mejora real.
- **¿Descubre que puede crear y retomar?** Sí. El `＋` y el `▶` están donde el patrón de cualquier
  UI de chat los pone, y el E2E lo confirmó contra el binario instalado.
- **¿Algo lo puede confundir?** **Sí, y es grave: la lista misma.** Con A-1 y A-2 vivos, después de
  media jornada tiene cinco filas que dicen todas `nueva conversación` y todas `sin fecha`. La
  función que la lista existe para cumplir —«¿cuál era aquella conversación?»— no se puede cumplir.
  Peor: **es indistinguible de un bug de carga**. Lo más probable es que crea que la lista está
  rota, no que los títulos no se derivan.
- **¿Puede perder trabajo?** **No encontré ningún camino de pérdida.** La transición es atómica con
  rollback total y probado, el archivado conserva `Conv` y `Checkpoint`, la cuarentena preserva el
  archivo roto sin pisarlo, el esquema futuro bloquea la escritura, la migración respalda antes y no
  pierde ningún campo del registro real, y el re-key sólo toca `Arnes`. Esta parte del paquete está
  bien hecha.

**Contraste:** **no lo medí yo mismo.** Ver §8.

---

## 8 · Lo que NO pude auditar, y por qué

Un auditor que no declara sus límites es el peor de todos. Estos son los míos:

1. **El barrido uno-por-uno de los 58 RF y los 50 escenarios E-01…E-50.** Prioricé la reproducción
   técnica y el análisis por mutación, y el presupuesto se consumió ahí. Lo que reporto en §6.3 son
   los RF que verifiqué directamente; **el resto no está ni verificado ni desmentido por mí**. La
   tabla de cobertura de `plan-pruebas.md` sigue sin auditoría independiente.
2. **El contraste WCAG medido con mis propios números.** No corrí el cálculo sobre los pares
   fg/bg de los componentes nuevos. Las desviaciones de contraste declaradas (C-3 en 3,76:1, C-13
   en 2,51:1) **quedan sin verificación independiente**; las acepto como declaradas, no como
   medidas por mí.
3. **El lector de pantalla real.** Leí el marcado ARIA y lo encontré bien construido, pero **no
   corrí un lector de pantalla** ni verifiqué qué se anuncia al crear, retomar, rotar o fallar. Que
   los atributos estén no prueba que el anuncio llegue.
4. **La app corriendo.** No levanté el daemon ni el navegador: todo mi trabajo de FE es lectura de
   código, stories y las capturas del paquete. A-4, A-5 y A-6 están derivados del código con
   certeza alta (la cadena `INICIAL → cargando → Esqueleto → refs null` es directa), pero **no los
   vi fallar en pantalla**.
5. **CI.** Sigue sin observarse, como declara el paquete. Nada de lo que corrí lo sustituye.
6. **`go-arch-lint`.** No está en el `PATH` de este entorno ni tiene target en el `Makefile` —
   mismo límite que declaró el tramo 4, confirmado por mí.
7. **T32 y el estado de `~/.arnesia/`.** Por orden explícita no ejecuté el borrado ni toqué nada
   bajo `~/.arnesia/`. `sessions.json` verificado idéntico al abrir y al cerrar
   (`b1689d1513e8d3c3fa21692c511ed793`).
8. **Los 3 tramos medidos sobre árbol compartido (N-6).** No re-medí los tramos 0-2 en aislamiento;
   sus cifras siguen sin ser propias y ese hallazgo sigue vigente.

---

## 9 · Lo que esto significa para el gate de PARIDAD

**A-1 y A-2 invalidan la firma de dos decisiones firmadas** (CV-D9 y CV-D13): lo construido no hace
lo que dicen. No es un matiz de implementación — es la mitad de lo que la fila de la lista tenía que
mostrar. **El operador no debería firmar PARIDAD hasta que estén cerradas**, porque firmaría la
paridad de un dibujo con un producto que en la pantalla real se ve distinto del mockup (donde cada
fila tiene su título y su fecha).

**A-10 es una enmienda a CV-D16 disfrazada de deuda técnica**, y es el ejemplo exacto del riesgo que
abrió el Gate 2 «por directiva, no por lectura»: `arquitectura.md` cambió el mecanismo de una
decisión firmada, con buen argumento, sin que el operador lo viera. **Debe presentarse como enmienda
para firmar o rechazar**, no como hallazgo.

**CV-D17 sigue sin firmar** y su contenido (adelantar RF-333/334 al tramo 3) me parece
técnicamente inevitable y bien argumentado — pero la firma es del operador, no mía.
