# Relevamiento AS-IS — Las conversaciones viven en el panel de conversación

> `tipo: relevamiento` · paquete `2026-07-26-conversaciones-del-panel` · relevado 2026-07-26.
> **Alcance:** el ES-ASÍ-HOY y el CON-QUÉ-REGLAS. El QUÉ está firmado en
> [`decisiones.md`](./decisiones.md) (CV-D1..CV-D15) y [`mockup-conversaciones-panel.html`](./mockup-conversaciones-panel.html).
> Este documento **no propone diseño** — describe el terreno y enumera los huecos.
>
> **Contrato de honestidad:** toda afirmación lleva cita `archivo:línea`, `commit-sha` o comando
> corrido. Lo que no se pudo verificar dice `NO VERIFICADO`. Cifras GENERADAS, no tecleadas.
>
> **Contexto de la corrida:** `HEAD = dd460f3` · versión `0.2.24` (`make version`) · daemon vivo
> en `127.0.0.1:4200`, sello `0.2.24.2607261724`, `huella: 7b1d636+sucio` (`GET /api/version`).

---

## 0. Resumen del terreno

| # | Hallazgo | Severidad | Dónde |
|---|---|---|---|
| 1 | **No hay mecanismo de migración de esquema en disco.** `store.Registry` hace `json.Unmarshal` directo a `[]domain.Session`, sin `schema_version` ni versionado de ningún tipo | 🔴 crítico | §2.1 |
| 2 | **`Conv` y `Checkpoint` se destruyen al cerrar** — dos líneas explícitas, con un test que cementa el comportamiento | 🔴 crítico | §2.3 |
| 3 | **El OpenAPI driftó**: cero menciones de `cerradas` en 2 endpoints/params que el código sirve hace días | 🔴 crítico | §3.3 |
| 4 | **El split-brain de llaves también afecta a las sesiones VIVAS**, no solo a las cerradas — corrige la premisa de `decisiones.md:23` | 🔴 crítico | §9.1 |
| 5 | **El estado del turno es por-SESIÓN, no por-conversación** (`curRun`, `resumeRetried`, `pendingPerm`, `grants`, `finalizedRun`) — hueco del modelo no cubierto por CV-D1..D15 | 🔴 crítico | §5.5 |

---

## 1. Modelo de dominio hoy

`internal/domain/session.go` — 128 líneas, un solo agregado. El doc de paquete ya declara la
fusión que CV-D3 rompe (`session.go:3-8`):

> «A Session is a *frente de trabajo* (work-front): **one live Claude Code conversation** bound N:1
> to an arnés.»

La entidad es `Session` (`session.go:68`). Los tipos satélite: `SessionStatus` (`:11`, enum
`streaming|await|idle` con `Valid()` en `:23`), `Salud` (`:34`, enum `ok|warn|crit|info`), `Rol`
(`:46`, enum `user|assistant|sys|act`) y `Turn` (`:62`, `{Rol, Text}` — **sin timestamp**).

### 1.1 Campo por campo

Leyenda de reparto CV-D3: **S** = queda en Sesión · **C** = migra a Conversación · **?** = ambiguo.

| Campo | Línea | Significa | Lo ESCRIBE | Lo LEE | CV-D3 |
|---|---|---|---|---|---|
| `ID` | `:71` | id estable nuestro, sobrevive reinicios; distinto del CC id | `newID()` (`session_service.go:929`) vía `Create` (`:250`) | todo; llave del `map[string]*sessionRuntime` (`:116`) y del `dockFrame.SessionID` (`:53`) | **S** |
| `Frente` | `:74` | nombre humano del frente, auto-derivado del 1er mensaje, editable en el rail | `Turn` si está vacío o es `"nuevo frente"` (`session_service.go:352-354`, vía `deriveFrente` `:915`); `Rename` (`:269`) | tarjeta del rail; header del dock (`chat-dock.tsx:24`) | **?** ver §1.2 |
| `Arnes` | `:78` | id del arnés (N sesiones comparten uno) | `Create` desde el body HTTP (`sessions.go:98`) | `spawnLocked` → `resolver.Resolve` (`session_service.go:386`); filtro de `listSessions` (`sessions.go:25`); filtro de `Cerradas` (`session_historial.go:87`); `roleFor` (`session_service.go:421`) | **S** |
| `Empresa` | `:79` | metadata de encuadre | `Create` (`sessions.go:100`) | rail | **S** |
| `Puesto` | `:80` | idem | `Create` (`sessions.go:101`) | rail | **S** |
| `Salud` | `:81` | punto de salud del arnés en el rail | `Create` (`sessions.go:102`) | rail | **S** |
| `Status` | `:85` | estado vivo del conductor (pip) | `Turn` (`:356`), `consume` (`:452+`), `Close`, spawn fallido (`:366`) | guarda `ErrBusy` (`:339`); pip del rail y del dock (`chat-dock.tsx:23`) | **?** ver §1.2 |
| `View` | `:86` | vista parqueada (Mapa/Diag/…) | `SetView` (`:284`) ← `PATCH` (`sessions.go:134`) | view-strip | **S** |
| `Parked` | `:87` | hint free-text «quedaste en…» | `Create` (`sessions.go:104`) | **sin productor vivo** — declarado en `capabilities/fe-shell/topbar-breadcrumb-k-dock.yaml:15`: «el campo parked no tiene productor vivo, solo seed hardcodeado» | **S** (muerto) |
| `Reparacion` | `:92` | la sesión edita una INSTALACIÓN del Portafolio (ley A4, RF-191) | `Create` (`sessions.go:105`) | chip del rail; tarjeta de identidad (`session_grounding.go`) | **S** |
| `ClaudeSessionID` | `:96` | id de sesión CC capturado del `system/init`; lo que va a `--resume` | `consume`/`EventInit`; se limpia en `rotarLocked` (`session_rotacion.go:61`) | `spawnLocked` (`session_service.go:390`) → `--resume` (`conductor.go:77-78`); `SessionLine` del dock (`chat-dock.tsx:82`) | **C** |
| `Model` | `:97` | modelo del conductor | `consume`/`EventInit` | `spawnLocked` → `--model` (`conductor.go:81`); `SessionLine` (`chat-dock.tsx:85`) | **?** ver §1.2 |
| `CtxPct` | `:99` | último uso conocido de ventana de contexto (0-100) | `consume` desde el `result` | barra ctx del dock (`chat-dock.tsx:87-90`); disparador de rotación | **C** |
| `CtxHist` | `:103` | histórico de `CtxPct` por turno (RF-194); acotado a `maxCtxHist = 500` (`session_service.go:137`) | `consume` | umbral de rotación | **C** |
| `RotacionPendiente` | `:107` | el umbral se cruzó — el PRÓXIMO turno spawnea fresco (RF-195) | `consume` al cruzar `umbralRot` | `Turn` (`session_service.go:347`) | **C** |
| `CadenaCC` | `:111` | los `ClaudeSessionID` previos (rotaciones, RF-198): el join a N JSONLs | `rotarLocked` (`session_rotacion.go:58`); `archivarLocked` (`session_historial.go:47`) | `HistorialCerrada` (`session_historial.go:112`) | **C** |
| `Checkpoint` | `:115` | digest mecánico de la última rotación (RF-196); viaja por el system-prompt | `rotarLocked` (`session_rotacion.go:60`) ← `CheckpointMecanico` (`:25`); **se BORRA en `archivarLocked`** (`session_historial.go:58`) | `spawnLocked` (`session_service.go:406-408`) | **C** |
| `Cwd` | `:119` | dir real del conductor; el join hacia `~/.claude/projects/<dir>/` | `spawnLocked` (`session_service.go:442`); backfill en `archivarLocked` (`session_historial.go:51-54`) | `history.Reader.Turnos` (`reader.go:64`) | **S** (pero lo necesita C — §10.7) |
| `CerradaEn` | `:123` | RFC3339 de cierre — «solo poblada en el registro de cerradas» | `archivarLocked` (`session_historial.go:59`) | fila del picker (`new-session-picker.tsx:282`) | **C** |
| `Turnos` | `:124` | cantidad de turnos archivada | `archivarLocked` (`session_historial.go:56`) | fila del picker (`new-session-picker.tsx:284`) | **C** |
| `Conv` | `:127` | transcript liviano de replay | `Turn` (`:355`), `consume`, `rotarLocked` (`session_rotacion.go:63`); **se ANULA en `archivarLocked`** (`session_historial.go:57`) | `Messages` del dock (`chat-dock.tsx:99`); `CheckpointMecanico` (`session_rotacion.go:25`) | **C** |

### 1.2 Los ambiguos (marcados, no resueltos — son del arquitecto)

1. **`Frente` (`:74`)** — CV-D9 dice que el título auto-derivado pasa a la conversación y «la sesión
   conserva su propio `Frente`». Hoy hay **un solo campo con dos semánticas fusionadas**: el rail lo
   usa como nombre del frente (`session-rail`) y el dock lo usa como título del hilo
   (`chat-dock.tsx:24`). El productor es uno solo (`session_service.go:352-354`) y se dispara con el
   **primer mensaje del usuario** — o sea, hoy el nombre del frente ES el título de la primera
   conversación. Partirlo exige decidir qué hereda el campo existente.

2. **`Status` (`:85`)** — es el estado del *conductor*, no del frente. Bajo CV-D3 pertenece a la
   conversación activa, pero el **pip del rail** lo lee a nivel sesión y `selectAttention`
   (`sessions-store.ts:467`) cuenta sesiones en `await`. Una conversación inactiva no tiene estado de
   conductor; la sesión necesita proyectar el de su activa.

3. **`Model` (`:97`)** — se captura del `init` de CADA proceso CC, así que es por-conversación
   (incluso por-rotación). Pero también se pasa al spawn (`session_service.go:431` →
   `conductor.go:81`), lo que sugiere una preferencia de sesión. Hoy los dos usos comparten campo.

4. **`Cwd` (`:119`)** — es del frente de trabajo (**S**), pero es el **join obligatorio** que
   `history.Reader` necesita para leer la JSONL de una conversación (`reader.go:64`). Si la
   conversación no se lo lleva copiado, cambiar el cwd de la sesión rompe el historial de todas sus
   conversaciones viejas. Ver §10.7.

5. **`Turn` (`:62`) no tiene timestamp.** El struct es exactamente `{Rol, Text}`. CV-D13 pide «última
   interacción» por conversación: **no existe nada de dónde derivarlo**. Lo único temporal en todo el
   dominio es `CerradaEn` (`:123`), que es la fecha de cierre, no la del último mensaje. Ver §10.1.

---

## 2. Persistencia hoy

### 2.1 El adaptador — `internal/adapters/store/registry.go` (97 líneas)

Un solo tipo, `Registry` (`registry.go:20`), `{path string; mu sync.Mutex}`. Satisface
`ports.SessionStore` por assertion estructural (`registry.go:25-28`).

| Aspecto | Cómo es hoy | Cita |
|---|---|---|
| **Formato** | Array JSON plano de `domain.Session`, indentado 2 espacios | `registry.go:73` (`json.MarshalIndent(sessions, "", "  ")`) |
| **Ruta default** | `~/.arnesia/sessions.json` | `registry.go:38` |
| **Locking** | `sync.Mutex` **in-process** en `Load` (`:46`) y `Save` (`:67`). **No hay lock de archivo** — dos daemons sobre el mismo `$HOME` se pisan | `registry.go:22, 46, 67` |
| **Atomicidad de escritura** | Sí: temp en el mismo dir + `os.Rename` (swap atómico en el mismo FS) | `registry.go:77-95`; doc en `:3-5` |
| **Archivo ausente** | No es error → registry vacío (primer arranque) | `registry.go:50-52` |
| **Permisos** | dir `0o750`; en disco los `.json` quedan `-rw-------` (verificado con `ls -la ~/.arnesia/`) | `registry.go:70` |
| **MIGRACIÓN DE ESQUEMA** | **NO EXISTE. Ningún mecanismo.** | ver abajo |
| **Versionado del esquema** | **NO EXISTE.** No hay `schema_version`, ni envelope, ni discriminador | ver abajo |
| **Tests** | **CERO.** `internal/adapters/store` reporta `[no test files]` en `go test ./...` | corrida verificada |

**🔴 Hallazgo central — no hay migración ni versión de esquema.** `Load` hace
`json.Unmarshal(b, &sessions)` desnudo sobre `[]domain.Session` (`registry.go:56-59`). Consecuencias
mecánicas del comportamiento de `encoding/json`:

- Un campo **que desaparece** del struct se ignora en silencio (no hay `DisallowUnknownFields`).
- Un campo **nuevo** queda en su zero-value, indistinguible de «estaba vacío».
- Un cambio de **forma** (array → objeto, o `Session` → `{Sesion, Conversaciones[]}`) hace fallar el
  `Unmarshal` entero → `Load` devuelve error → **el daemon pierde el registro completo**.
- No hay forma de saber con qué versión se escribió un archivo.

Esto contrasta con el índice, que **sí** tiene el mecanismo: `internal/adapters/index/store.go`
lleva `schemaVersion` y una política declarada de wipe-and-rebuild, enforced por
`docs/architecture/fitness/arch_test.go:TestSchemaVersionTriggersRebuild:538`. El registro de
sesiones no tiene nada equivalente — y **no puede copiar la política del índice**, porque el índice
es desechable (se reconstruye del árbol de arneses) y `sessions.json` **es fuente durable**: si se
borra, se pierden las sesiones del operador.

**Deuda relacionada y abierta** — `docs/product/BACKLOG.md:158-162` deja sin política definitiva la
conservación de entradas corruptas: hoy una entrada sobrevive si es JSON sintácticamente válido, y
**se pierde todo si el archivo entero está roto**. Bajo CV-D8 ese archivo pasaría a contener todos
los transcripts (§2.5), así que el radio de daño de una corrupción crece.

### 2.2 Quién escribe, cuándo

El único productor es `SessionService.persistLocked()` (`session_service.go:902-912`): snapshotea
`s.order` → `[]domain.Session` y llama `s.store.Save`. Un fallo se loguea con `slog.Error` y **no
propaga** (`:909-911`).

Se invoca desde: `Create` (`:250`), `Rename` (`:269`), `SetView` (`:284`), `Close` (`:317`), `Turn`
(`:367`, `:373`) y el loop `consume` (`:452+`). O sea: **cada turno reescribe el archivo entero**,
incluidos todos los `Conv` de todas las sesiones.

El cableado del composition root: `store.NewRegistry(...)` en `cmd/arnesia/main.go`, y el registro
de cerradas en `main.go:325-328`:

```go
if cerradasStore, cerr := store.NewRegistry(cerradasPathDefault(*sessionsPath)); cerr != nil {
    slog.Warn("historial: registro de cerradas no disponible", "err", cerr)
} else {
    sessionSvc.SetArchivoCerradas(cerradasStore)
}
```

`cerradasPathDefault` (`main.go:795-805`) resuelve `~/.arnesia/sesiones-cerradas.json`, o el hermano
de `--sessions` si se pasó la flag. **Es el mismo adaptador `Registry`** — mismo formato, mismas
carencias (sin migración, sin versión, sin tests).

`SetArchivoCerradas` (`session_historial.go:23-27`) es un setter con lock. Su doc declara la
degradación honesta: «nil = comportamiento previo (borrado seco)».

### 2.3 🔴 El `Conv` se descarta al cerrar — el código exacto

`internal/usecase/session_historial.go:56-58`, dentro de `archivarLocked`:

```go
cerrada.Turnos = len(cerrada.Conv)
cerrada.Conv = nil
cerrada.Checkpoint = ""
```

La intención está escrita arriba, en `session_historial.go:37-39`:

> «archiva la metadata liviana […] cantidad de turnos y fecha de cierre — **SIN Conv** (B2: la
> JSONL nativa es la verdad del contenido)»

`Close` (`session_service.go:299-324`) archiva **antes** de borrar (`:309`), luego saca la entrada
del `map` y del `order` y persiste.

**Dos consecuencias que el paquete tiene que asumir explícitamente:**

1. **`Checkpoint` también se destruye** (`:58`). CV-D11 promete retomar «con
   `CadenaCC`/`Checkpoint`/`Cwd` intactos» — pero hoy el archivado lo borra. **Conservarlo es un
   requisito NUEVO, no una preservación.** No figura en CV-D1..D15.

2. **Hay un test que cementa el comportamiento actual** —
   `internal/usecase/session_historial_test.go:73`:

   ```go
   if c.Conv != nil {
       t.Error("el Conv NO viaja al archivo (B2: la JSONL nativa es la verdad)")
   }
   ```

   CV-D8 invierte esta decisión, así que ese test **hay que reescribirlo deliberadamente y con
   justificación**. Es exactamente la clase de falla que HS-28 documentó: un test que afirma lo
   contrario del invariante deseado y que ningún otro test caza, porque el test ERA el defecto.

**Verificado en disco** (`~/.arnesia/sesiones-cerradas.json`, 1479 bytes, 3 entradas):

| id | arnes | cwd | cadena_cc | turnos | conv | checkpoint |
|---|---|---|---|---|---|---|
| `s1b38a066` | `vitalia` | `/home/…/luana-vitalia/vitalia` | 1 | 12 | **ausente** | ausente |
| `s408bb085` | `dev-full-cycle` | `/home/…/harness-studio/dogfood/dev-full-cycle` | 1 | 12 | **ausente** | ausente |
| `s020210e3` | `vitalia` | `/home/…/luana-vitalia/vitalia` | — | — | **ausente** | ausente |

`s020210e3` cerró sin ningún turno: sin `claude_session_id`, sin `cadena_cc`, sin `turnos`.
`GET /api/sessions/cerradas/s020210e3/historial` responde `{"faltantes":null,"turnos":null}` — 200
honesto, ni error ni contenido inventado (verificado en vivo).

### 2.4 El corpus JSONL nativo — `internal/adapters/history/reader.go` (124 líneas)

El package doc declara el boundary que lo gobierna (`reader.go:1-5`): «la JSONL es LA fuente de
verdad de una conversación — este adapter solo la LEE (boundary `indice-desechable-jsonl-es-verdad`),
jamás la escribe ni la mueve. ArnesIA aporta el join (`Session.CadenaCC` + cwd)».

| Aspecto | Detalle | Cita |
|---|---|---|
| **Base** | `~/.claude/projects` (vacío → default) | `reader.go:31` |
| **Ruta de una JSONL** | `<base>/<DirParaCwd(cwd)>/<claudeSessionID>.jsonl` | `reader.go:64` |
| **Encoding del dir** | todo carácter **no alfanumérico** → `-`. Verificado 2026-07-22 contra el CLI real | `reader.go:36-49` |
| **Qué extrae** | solo líneas `type: user`/`assistant` con `message` no nulo | `reader.go:79-81` |
| **Qué descarta** | bloques sin texto (`tool_use`/`tool_result`) — «el historial es la conversación, no el trace»; y turnos con texto en blanco | `reader.go:59-61`, `:82-85` |
| **Formas de `content`** | string directo, o array de bloques `{type:text,text}` | `reader.go:105-123` |
| **Buffer** | hasta 4 MiB por línea (`1<<22`) | `reader.go:73` |
| **Tolerancia** | una línea que no parsea **se saltea**, no rompe el historial entero | `reader.go:76-78` |
| **Ausencia** | JSONL faltante → error envuelto; el caller lo reporta honesto | `reader.go:61-62, 66-68` |
| **`Existe()`** | stat sin leer | `reader.go:99-102` |
| **Tests** | 2: `TestDirParaCwd` (`reader_test.go:13`), `TestTurnosLeeCorpus` (`:24`) | — |

**Qué NO garantiza (importante para CV-D8/D11):**

- **No garantiza permanencia.** El corpus lo gestiona Claude Code, no ArnesIA; una JSONL puede
  desaparecer. Por eso `HistorialCerrada` devuelve `faltantes` (`session_historial.go:114-116`).
- **No preserva la actividad.** `RolAct`/`RolSys` (permisos, gate, breadcrumbs de rotación) **no
  están en la JSONL** — son construcción nuestra que solo vive en `Conv`. Un historial reconstruido
  desde JSONL **pierde** todo lo que CAP-100 agregó. Esto es un argumento fuerte a favor de CV-D8 que
  el paquete todavía no usa.
- **No garantiza fidelidad al alcance.** El texto enviado lleva la línea `[alcance: …]` antepuesta
  por el FE (`sessions-store.ts:188-190`), así que el turno en la JSONL es el crudo con el prefijo.
- **No tiene índice ni búsqueda.** Es scan lineal por archivo.

### 2.5 Volumen medido (no estimado)

Corrido sobre `~/.arnesia/sessions.json` (17202 bytes totales):

| sesión | turnos | bytes del `conv` (JSON) |
|---|---|---|
| `s6165ac75` | 90 | **12 095** |
| `s25123a2c` | 4 | 439 |
| `s0fec7798` / `sfc512b15` / `s78b3aeeb` | 0 | 2 c/u |

**Corrección honesta a CV-D8:** `decisiones.md:83-84` dice «90 turnos = **9.4 KB**». Medido hoy, la
misma conversación serializada como JSON son **12 095 bytes (11.8 KB)** — ~25 % más. La conclusión
de CV-D8 (scan en memoria, sin FTS5) **no cambia** con ese delta, pero la cifra citada en la spec
debería ser la medida.

---

## 3. API HTTP hoy

### 3.1 Registro de rutas — `internal/adapters/transport/http/router.go:122-138`

Diez rutas bajo `/api/sessions`, con `net/http` mux nativo (patrones método+path de Go 1.22+):

| # | Ruta | Handler | Línea |
|---|---|---|---|
| 1 | `GET /api/sessions` | `listSessions` | `router.go:123` |
| 2 | `GET /api/sessions/cerradas/{id}/historial` | `historialCerrada` | `router.go:124` |
| 3 | `POST /api/sessions` | `createSession` | `router.go:125` |
| 4 | `GET /api/sessions/{id}` | `getSession` | `router.go:126` |
| 5 | `PATCH /api/sessions/{id}` | `patchSession` | `router.go:127` |
| 6 | `DELETE /api/sessions/{id}` | `deleteSession` | `router.go:128` |
| 7 | `POST /api/sessions/{id}/turn` | `sessionTurn` | `router.go:129` |
| 8 | `POST /api/sessions/{id}/permission` | `resolvePermission` | `router.go:130` |
| 9 | `POST /api/sessions/{id}/interrupt` | `sessionInterrupt` | `router.go:131` |
| 10 | `POST /api/sessions/{id}/dictado` | `postDictado` (condicional) | `router.go:137` |

Nota de orden: la ruta 2 (`/cerradas/{id}/historial`) se registra **antes** que `/{id}` — con el mux
de Go el patrón más específico gana, así que no hay ambigüedad, pero `cerradas` es de facto un id
reservado.

### 3.2 Los dos endpoints del historial, en detalle

**`GET /api/sessions?arnes=…&cerradas=1`** — `sessions.go:18-43`. El handler tiene **dos formas de
respuesta distintas según query param**, lo cual es el drift de contrato más incómodo:

| Caso | Status | Body |
|---|---|---|
| sin `cerradas=1` | 200 | **array** `[]Session` (`sessions.go:32`) |
| `cerradas=1`, OK | 200 | **objeto** `{sesiones: [], cerradas: []}` (`sessions.go:41`) |
| `cerradas=1`, error al cargar cerradas | **200** | `{sesiones: [], cerradas_error: "…"}` (`sessions.go:38`) — degradación honesta: «las vivas viajan igual; el hueco de cerradas se DICE» (`:37`) |

El filtro por arnés es un `==` desnudo sobre el string (`sessions.go:25` para vivas;
`session_historial.go:87` para cerradas). **No normaliza llaves** — de ahí el bug de §9.1.
Validación: ninguna. `arnes=` vacío devuelve todo.

**`GET /api/sessions/cerradas/{id}/historial`** — `sessions.go:47-56`:

| Caso | Status | Body |
|---|---|---|
| OK | 200 | `{turnos: [], faltantes: []}` (`sessions.go:54`) |
| id inexistente **o lector no cableado** | **404** | `{error: "…"}` (`sessions.go:51`) |

⚠️ El 404 **colapsa dos causas distintas**: `HistorialCerrada` devuelve `errNotFound(id)` si no
encuentra el id (`session_historial.go:122`) pero también `errors.New("lector de historial no
cableado")` si `s.historial == nil` (`:101-102`), y el handler mapea **cualquier** error a 404. Un
lector no cableado se reporta como «no existe esa conversación» — una mentira estructural del
contrato.

La reconstrucción cose la cadena (`session_historial.go:112-119`): por cada `ccid` en `CadenaCC`
llama `reader.Turnos(c.Cwd, ccid)`; si falla, suma a `faltantes` y sigue. **Nunca inventa contenido.**

**Verificado en vivo contra el daemon:**

```
GET /api/sessions                                  → 5 vivas (array)
GET /api/sessions?cerradas=1                       → {sesiones:5, cerradas:3}, sin cerradas_error
GET /api/sessions?arnes=vitalia&cerradas=1         → vivas=3 cerradas=2
GET /api/sessions?arnes=sin-home~vitalia~vitalia…  → vivas=1 cerradas=0     ← el bug
GET /api/sessions?arnes=dev-full-cycle&cerradas=1  → vivas=0 cerradas=1
GET /api/sessions/cerradas/s1b38a066/historial     → turnos=13 faltantes=null
GET /api/sessions/cerradas/s020210e3/historial     → turnos=null faltantes=null
```

### 3.3 🔴 El OpenAPI driftó — hallazgo verificado

`docs/architecture/contracts/api/openapi.yaml` declara 6 paths de sesiones: `/sessions` (`:605`),
`/sessions/{id}` (`:630`), `/turn` (`:665`), `/permission` (`:687`), `/interrupt` (`:723`),
`/dictado` (`:742`).

**`grep -c "cerradas" docs/architecture/contracts/api/openapi.yaml` → `0`.**

O sea, el spec **no conoce nada del historial B2**, que está en producción y sirviendo:

| Lo que el código hace | Lo que el spec dice |
|---|---|
| `GET /api/sessions/cerradas/{id}/historial` (`router.go:124`) | **el path no existe** |
| `?arnes=` y `?cerradas=1` en `GET /sessions` (`sessions.go:20,31`) | **sin `parameters:`** (`openapi.yaml:606-614`) |
| respuesta objeto `{sesiones, cerradas}` con `cerradas=1` (`sessions.go:41`) | `schema: {type: array, items: Session}` (`openapi.yaml:614`) — **contradicho** |
| `cerradas_error` como degradación honesta 200 (`sessions.go:38`) | no modelado |

**Y el drift no lo caza nadie.** El paso de CI que verificaría la generación desde el spec es
condicional y hoy está inerte — `.github/workflows/ci.yml:71-77` corre solo
`if: hashFiles('web/src/shared/api/generated/**') != ''`, y ese directorio no existe. El FE tipa a
mano en `web/src/shared/api/client.ts`. El boundary `fe-transporte-independiente` exige que los tipos
de `shared/api/generated/**` deriven del OpenAPI, pero como no hay generados, la regla no muerde.

**Consecuencia para el paquete:** cualquier endpoint nuevo de conversaciones va a nacer con el mismo
agujero salvo que se decida explícitamente sincronizar el spec. Hoy el OpenAPI **no es contrato
enforced**, es documentación con drift.

---

## 4. Estado del FE hoy

### 4.1 `web/src/shared/store/sessions-store.ts` (479 líneas)

Zustand. Su doc de cabecera (`:1-4`) ya declara el acoplamiento: «Multisesión store […] mirrors the
daemon's session registry and drives the live Dock».

**Estado (`:26-47`)** — nueve mapas, **todos indexados por `session.id`**:

| Campo | Línea | Qué guarda |
|---|---|---|
| `sessions: Session[]` | `:27` | la lista espejo del daemon |
| `activeId: string \| null` | `:28` | **la sesión activa — no hay noción de conversación activa** |
| `streaming: Record<string,string>` | `:34` | texto del assistant en vuelo, por sesión |
| `finalizedRun: Record<string,string>` | `:38` | último `run_id` terminal, por sesión (idempotencia) |
| `pendingPerms: Record<string,PermissionAsk[]>` | `:40` | tarjetas de permiso abiertas, por sesión |
| `scope: Record<string,ScopeNode\|null>` | `:42` | chip de alcance, por sesión (RF-118) |
| `wroteInRun: Record<string,boolean>` | `:44` | el turno aprobó escrituras ⇒ gate |
| `msgFlushed: Record<string,boolean>` | `:47` | ya cerró burbuja este turno (CH-D3) |
| `chatOpen`, `railCollapsed`, `connected`, `loaded` | `:29-32` | chrome |

**SSE:** singleton `dockConn` (`:71`), abierto una vez en `init` (`:128-136`) vía `connectDock`. El
router de frames es `onDock` (`:262-459`), un switch de 10 kinds: `status`, `init`, `delta`,
`result`, `error`, `permission`, `permission_result`, `message`, `act`. **Todos rutean por
`f.session_id`** (`:263`) y aplican con `patch(list, id, …)` (`:74-76`) o `appendConv` (`:80-87`).

### 4.2 Puntos de ruptura con CV-D3 — lista con cita

| # | Punto | Cita | Por qué rompe |
|---|---|---|---|
| R1 | `activeId` es el único puntero de foco | `sessions-store.ts:28`, `:126`, `:139` | Con N conversaciones por sesión hace falta un segundo eje (`activeConvId`). `switchTo` es literalmente `set({ activeId: id })` (`:139`) — **no carga nada**, porque hoy la sesión ya trae su `conv` embebido |
| R2 | `conv` vive dentro de `Session` | `sessions-store.ts:86`, `:196`, `:324` | El transcript se muta con `patch`/`appendConv` sobre la sesión. Con conversaciones separadas hay que decidir si el frame trae `conversation_id` o si el FE lo infiere |
| R3 | `sendTurn` escribe al `activeId` | `:177-180` | Manda al frente, no al hilo. Necesita saber a qué conversación |
| R4 | Guarda cliente de un-turno-a-la-vez por sesión | `:183-184` (`if status === "streaming" \|\| "await") return`) | Espeja el `ErrBusy` del daemon (`session_service.go:339`). Si el guard queda a nivel sesión con conversaciones múltiples, es correcto por CV-D7 (una activa) pero hay que re-anclarlo |
| R5 | `finalizedRun[id]` por sesión | `:38`, `:266`, `:318` | Idempotencia de frames (boundary `frames-idempotentes-run-id`). Retomar una conversación vieja con runs previos puede colisionar con el `run_id` de la activa. **Ver §5.5** |
| R6 | `pendingPerms[id]` por sesión | `:40`, `:394-401` | El `INDEX.md:66` del paquete deja abierto «qué pasa con permisos pendientes de la conversación que se desactiva». Hoy no hay respuesta: las tarjetas cuelgan de la sesión |
| R7 | `streaming[id]` / `msgFlushed[id]` / `wroteInRun[id]` por sesión | `:34`, `:44`, `:47` | Buffers del turno en vuelo; pertenecen a la conversación activa |
| R8 | `scope[id]` por sesión | `:42`, `:251-255` | RF-118 dice «el alcance pertenece a UNA sesión». CV-D14 lo conserva a nivel sesión ⇒ probablemente **no** migra. Decisión del arquitecto |
| R9 | `selectActive` / `selectAttention` / `selectPendingPerms` / `selectScope` | `:464-478` | Los 4 selectores públicos asumen 1:1. `selectAttention` (`:467`) cuenta sesiones en `await` para el badge del rail |
| R10 | `closeSession` elimina la sesión entera | `:150-157` | Bajo CV-D12 «cerrar» deja de ser el verbo de la conversación. El cierre de SESIÓN sigue existiendo pero ya no implica cerrar un hilo |
| R11 | `create` abre el chat y hace foco | `:141-148` | Crear **sesión** ≠ crear **conversación** (CV-D2 pone el `＋` en el dock) |

### 4.3 `web/src/shared/api/client.ts` (400 líneas)

- **`BASE` hardcodea `:4200`** — `client.ts:21`:
  `const BASE = import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"`. Ver §9.3.
- Métodos de sesión: `listSessions` (`:214`), `createSession`, `closeSession`, `renameSession`,
  `setView`, `turn`, `resolvePermission`, `interrupt`.
- **Los dos del historial B2** (`client.ts:216-225`), tipados a mano:
  - `conversacionesDeArnes(arnes)` → `{sesiones: Session[]; cerradas?: Session[]; cerradas_error?: string}`
  - `historialCerrada(id)` → `{turnos: Turn[]; faltantes?: string[]}`

  Modelan fielmente la respuesta bimorfa del §3.2, incluido el `cerradas_error` opcional.
  **Ruptura:** ambos cuelgan de un **string de arnés**, que es exactamente lo que CV-D5 elimina.

### 4.4 `web/src/widgets/chat-dock/ui/chat-dock.tsx` (375 líneas)

Composición actual (`chat-dock.tsx:19-41`) — las **4 filas de cromo** que CV-D14 baja a 2:

| Fila | Componente | Línea | Qué pinta |
|---|---|---|---|
| 1 | header inline | `:21-34` | `<Pip>` + `active.frente` + botón colapsar |
| 2 | `<SessionLine>` | `:36`, def. `:78-93` | `◍ cc-id[0:8]` · `arnes` · `model` · `ctx` barra + `%` |
| 3 | `<ScopeRow>` | `:37`, def. `:46-76` | `Alcance:` + chip `arnés <b>` **fijo** + chip del nodo o hint |
| 4 | `<Messages>` / `<Composer>` | `:38-39` | cuerpo |

Detalles que las decisiones nombran literalmente:
- **`⟩ colapsar`** en `chat-dock.tsx:32` — el glifo que CV-D15 cambia a `»`. El `title` es «Colapsar
  el dock (⌘K para reabrir)» (`:29`), y CV-D15 lo conserva.
- **El `✕` de quitar alcance** en `chat-dock.tsx:62-69`, que CV-D14 manda conservar literal.
- **`vitalia` aparece 3 veces**, como dice CV-D14: `SessionLine` (`:84`), chip de alcance (`:53`) y
  el `placeholder` del composer (§ `Composer`, `:268+`).
- El estado vacío del `ScopeRow` gasta la fila entera en el hint «selecciona un nodo en el Mapa para
  acotar» (`:72`) — lo que CV-D14 elimina.

**Ruptura clave (verificada):** `chat-dock.tsx` **no importa `conversaciones-store` en ningún lado**
— su único import de estado es `sessions-store` vía el barrel `@/shared` (`chat-dock.tsx:3`). El dock
lee `active.conv` directo (`:99`) y no tiene lista, ni buscador, ni `＋`. Confirma el punto 1 de
`decisiones.md:11-12`.

**Deuda de verificación:** `chat-dock.tsx` (375 líneas) **no tiene stories**. Solo sus hijos
`permission-card.tsx` y `dictado-button.tsx` están cubiertos. Ver §8.

### 4.5 `web/src/widgets/session-rail/**`

| Archivo | Qué es | Reclamado por |
|---|---|---|
| `ui/session-rail.tsx` | el rail; `«`/`»` colapsar en `:96` (el glifo que CV-D15 reusa) | CAP-72 |
| `ui/new-session-picker.tsx` | picker inline; `crear` manda `arnes: e.clave` (`:125`) | CAP-72 |
| `model/portafolio-picker-store.ts` (+`.test.ts`) | selección de arnés del Portafolio | CAP-72 |
| **`model/conversaciones-store.ts`** (+`.test.ts`) | **el store que se muda al dock** | **CAP-98**, no CAP-72 |
| `ui/new-session-picker.stories.tsx` | 7 stories con `play()` | allowlist R2 |

**`conversaciones-store.ts`** (65 líneas) — zustand con `{arnes, vivas, cerradas, error, historialDe,
turnos, faltantes}` y tres acciones: `cargar(arnes)`, `abrirHistorial(id)`, `reset()`. Su
`cargar` propaga el `cerradas_error` del backend como campo visible («El backend reporta el hueco de
cerradas aparte (degradación honesta) — se muestra»).

**Rupturas:** `cargar` toma un **string de arnés**, no un `session_id` (CV-D5); el estado
`historialDe`/`turnos` modela **una** expansión a la vez, no una conversación activa retomable
(CV-D11); y no hay noción de activa/inactiva (CV-D12), solo `vivas`/`cerradas`.

**`ConversacionesDelArnes`** (`new-session-picker.tsx:263`) es el bloque que CV-D2 elimina: contenedor
`max-h-44` con `text-[10px]` (`:267`), contador «N abiertas · M cerradas» (`:269-272`), el
`text-warn` del error (`:273`), y por cada cerrada un botón que expande el transcript con turnos
truncados a una línea (`:296-299`). Se dispara desde `new-session-picker.tsx:111`
(`cargar(e.clave)` — con la clave calificada).

---

## 5. Conductor y ciclo de vida

### 5.1 El spawn — `internal/usecase/session_service.go:385-447`

Orden de operaciones:

1. `resolver.Resolve(r.meta.Arnes)` → `cwd` (`:386`). **Falla ⇒ falla el spawn** — no hay fallback a
   scratch dir.
2. `resume := r.meta.ClaudeSessionID` (`:390`).
3. Tarjeta de identidad por sesión (`:399-401`) vía `s.grounding`.
4. **Checkpoint de rotación** (`:406-408`): si `Checkpoint != ""`, se concatena a la tarjeta bajo el
   header `## Checkpoint de rotación (la conversación CONTINÚA)`. «un mecanismo, dos usos — MC-D6».
5. `injector.ProvisionSession(ctx, id, tarjeta)` (`:410`) — **la key es el `id` de SESIÓN**. Con N
   conversaciones, el system-prompt por sesión se pisaría entre hilos. Ver §10.6.
6. Permisos del rol (`:419-428`) vía `roleFor(arnes)` + `perms.ResolveForRole`. Degradación honesta
   con warn si el rol no resuelve.
7. `agent.Spawn(ctx, ports.SpawnOpts{Resume, Model, Cwd, MaxTurns, Injection, Permisos})` (`:429-436`).
8. Post-spawn (`:440-445`): `r.live`, `r.cwd`, **`r.meta.Cwd = cwd`** («persistido: el join del
   historial B2 sobrevive al cierre»), `r.wasResume = resume != ""`, `r.sawInit = false`, y
   `go s.consume(id, live)`.

**Los flags** — `internal/adapters/agent/claudecode/conductor.go`, función `SpawnArgs`:
`--resume <id>` (`:77-78`), `--model` (`:81`), `--max-turns` (`:87`), `--plugin-dir` (`:93`),
`--append-system-prompt-file` (`:96`), `--add-dir` (`:99`), `--mcp-config` + `--strict-mcp-config`
(`:106`), `--allowedTools` (`:147`), `--disallowedTools` (`:150`). El proceso se lanza en `:231`.

### 5.2 El turno — `session_service.go:329-381`

1. **Guarda de un-turno-a-la-vez** (`:339-342`): `streaming` **o** `await` ⇒ `ErrBusy` → 409. El
   comentario explica por qué `await` cuenta (`:337-338`): el turno sigue en vuelo, parqueado en un
   permiso.
2. **Rotación** (`:347-349`): si `RotacionPendiente`, `rotarLocked(r)` **antes** de todo. «Se rota
   ENTRE turnos por construcción (nunca streaming acá)» (`:346`).
3. **Auto-derive del `Frente`** (`:352-354`) si está vacío o es `"nuevo frente"`.
4. Append del turno user a `Conv` (`:355`), `Status = streaming` (`:356`), `runSeq++` y
   `runID = "<id>-r<n>"` (`:357-358`), `pendingTurn = text`, `resumeRetried = false` (`:360-361`).
5. Spawn perezoso si `r.live == nil` (`:364-371`).
6. `persistLocked()` + unlock, luego publish del frame `status` y `live.Send` (`:373-379`).

### 5.3 Rotación por contexto — `internal/usecase/session_rotacion.go` (64 líneas)

`rotarLocked` (`:52-64`), con `s.mu` tomado:

```go
if r.live != nil { _ = r.live.Close(); r.live = nil }        // :53-56
if r.meta.ClaudeSessionID != "" {
    r.meta.CadenaCC = append(r.meta.CadenaCC, r.meta.ClaudeSessionID)  // :58
}
r.meta.Checkpoint = CheckpointMecanico(r.meta.Conv)          // :60
r.meta.ClaudeSessionID = ""                                  // :61  → próximo spawn SIN --resume
r.meta.RotacionPendiente = false                             // :62
r.meta.Conv = append(r.meta.Conv, domain.Turn{Rol: RolSys, Text: breadcrumbRotacion})  // :63
```

`breadcrumbRotacion = "— contexto rotado, seguimos —"` (`:12`). `CheckpointMecanico` (`:25-46`) toma
los últimos `checkpointTurnos = 6` (`:15`), recorta cada uno a `checkpointMaxTurno = 700` (`:19`), y
arma un digest determinístico **sin LLM** (MC-D5) con la instrucción «NO saludes de nuevo ni pidas
que repitan nada — continuá donde quedó» (`:35`).

**CV-D10 pide que esto siga siendo invisible.** El mecanismo ya lo es: `Session.ID` no cambia,
`Conv` no se corta, y el breadcrumb `RolSys` queda inline. **Migra tal cual a la conversación**;
`CadenaCC` ya cose los N `ClaudeSessionID` en un hilo lógico.

El umbral se cablea en `main.go` con `sessionSvc.SetUmbralRotacion(*rotUmbral)` (flag
`-rotacion-umbral`, default 40 %). `maxCtxHist = 500` acota `CtxHist` (`session_service.go:135-137`).

**Estado vivo verificado:** la sesión `s25123a2c` está en disco con `ctx_pct: 68`,
`rotacion_pendiente: true` — o sea, hay una rotación armada esperando el próximo turno. Buen caso de
prueba para el paquete.

### 5.4 Interrupción y permisos

- `Interrupt` (`session_service.go:853`) + `ErrNadaQueInterrumpir` (`:846`). RF-116.
- `onControlRequest` (`:637`) — el gate HITL. Incluye el guardrail del **paquete cerrado** (CAP-99):
  `ProtegerPaqueteCerrado` (`:159`) + `refiereAlguno` (`:174`) deniegan en el acto cualquier
  `control_request` que mencione el árbol propio de ArnesIA, **antes** del auto-allow por grant y del
  click humano.
- `ResolvePermission` (`:725`) con `PermissionResolution` (`:695`); grants efímeros en
  `r.grants map[string]domain.Grant` (`:101`).
- `askUserQuestionUpdatedInput` (`:828`) para el tool `AskUserQuestion` (`:30`).

### 5.5 🔴 Qué pasa al cambiar de sesión activa — y el hueco de CV-D11

**Hoy, cambiar de sesión activa NO toca el conductor.** `switchTo` es literalmente
`set({ activeId: id })` (`sessions-store.ts:139`). No hay fetch, no hay `--resume`, no hay nada: el
transcript se repinta desde `active.conv`, que **ya viaja en la lista** de `GET /api/sessions`. Los
conductores de las otras sesiones **siguen vivos** en el daemon (`s.rt` es un map de runtimes
concurrentes). El `--resume` solo entra cuando `r.live == nil` y hay un `ClaudeSessionID` guardado —
típicamente tras reiniciar el daemon.

**Qué haría falta para retomar una conversación inactiva (CV-D11):**

1. **Un `ClaudeSessionID` que siga siendo válido.** El de una conversación de hace días tiene alta
   probabilidad de estar GC'd por Claude Code. Existe el auto-heal — `tryHealResume`
   (`session_service.go:600`), enforced por `arch_test.go:TestResumeAutoSana:1957`: si un spawn con
   `--resume` muere **antes** de un `init`, se limpia el `ClaudeSessionID` stale y se respawnea fresh
   **una vez** (`resumeRetried` `:96`). Pero un respawn fresco **pierde el hilo** salvo que le llegue
   el `Checkpoint` — que hoy se borra al archivar (§2.3).
2. **`Conv` persistido** para repintar antes de que el stream vivo se reenganche (CV-D8/D11). Hoy no
   está (§2.3).
3. **`Checkpoint` y `Cwd` conservados** — `Cwd` sí sobrevive (`session_historial.go:51-54` incluso lo
   backfillea), `Checkpoint` no.
4. **🔴 Desacoplar el estado del turno de la sesión.** Este es el hueco no cubierto por CV-D1..D15.
   Todo el estado in-flight vive en `sessionRuntime` (`session_service.go:85-109`), **una instancia
   por SESIÓN**:

   | Campo | Línea | Por qué duele con N conversaciones |
   |---|---|---|
   | `live ports.AgentSession` | `:87` | **un** proceso CC por sesión. CV-D7 (una activa) lo hace correcto, pero retomar exige cerrar el vivo y spawnear el de la conversación retomada |
   | `runSeq` / `curRun` | `:88-89` | el `run_id` es `"<sessionID>-r<n>"` (`:358`). Si el contador es de sesión, dos conversaciones comparten espacio de nombres |
   | `pendingTurn`, `wasResume`, `sawInit`, `resumeRetried` | `:93-96` | la máquina del auto-heal es por-sesión; retomar una vieja resetearía el estado de la activa |
   | `pendingPerm map[string]pendingPermission` | `:100` | permisos en vuelo — el `INDEX.md:66` deja esto abierto explícitamente |
   | `grants map[string]domain.Grant` | `:101` | grants efímeros: ¿se heredan al retomar otra conversación del mismo frente? |
   | `assembling`, `msgFlushed` | `:90`, `:105` | buffers del turno |
   | `cwd` | `:108` | resuelto al spawn |

   Y su espejo en el FE: `finalizedRun` (`sessions-store.ts:38`) dropea frames de un run ya terminado
   — con `run_id`s de dos conversaciones conviviendo, un frame legítimo puede quedar descartado.

   **Este reparto no está decidido en CV-D1..D15 y es el trabajo de arquitectura más denso del
   paquete.** Toca un boundary `enforced` de severidad `high` (§6.2).

---

## 6. Arquitectura as-code que aplica

`docs/architecture/INDEX.md:78` declara **26 boundaries**; verificado:
`ls docs/architecture/boundaries/*.md | wc -l` → **26**.

⚠️ **El INDEX tiene drift de conteo propio**: dice «26 boundaries · 136 checks» (`INDEX.md:78`) pero
más abajo «gran total arch/: 123 checks» (`INDEX.md:139-141`) y arriba «247 checks» (`:21-22`),
mientras `arch_test.go:11-12` dice 238. La única cifra viva y auto-regenerada es la del checkpoint
(§8.5). El INDEX también está **stale en ≥10 filas** de versión/estado (p.ej. marca
`ingesta-por-allowlist-declarada` como «pendiente» cuando el archivo dice `status: enforced` con 7
enforcers). **Tratar el INDEX como mapa, no como verdad**; el frontmatter de cada nodo manda.

### 6.1 Los boundaries que este cambio toca

🔴 directo · 🟡 colateral

| Boundary | Status / severity | Qué exige | Impacto |
|---|---|---|---|
| `conductor-no-parsea-jsonl.md` | `proposed` / high (`:5,20`) | §6.2 | 🔴 |
| `codigo-traza-a-capability.md` | **`enforced` / error** (`:5,20`) | §6.3 | 🔴 |
| `sesion-viva-consistente.md` | **`enforced`** / high | §6.4 | 🔴 |
| `dominio-independiente-de-transporte.md` | enforced | `internal/domain/**` no importa `net/http`, SSE, `database/sql`, `modernc.org/sqlite` | 🔴 entidades nuevas en `domain` |
| `indice-desechable-jsonl-es-verdad.md` | `proposed` / high | §6.5 | 🔴 |
| `ingesta-por-allowlist-declarada.md` | **`enforced`** | lo que entra de un emisor ajeno se persiste **por allowlist declarada, aplicada en la puerta, no en la consulta**; nada de contenido de conversación, rutas del usuario ni identidad de cuenta. Se declara aplicable «más allá de la telemetría… el `stream-json` del runtime» (`:84-87`) | 🔴 persistir `Conv` ES ingesta |
| `no-aplica-no-es-cero.md` | **`enforced` / error** | «no pude leer» viaja `null`, «leí y está vacío» `[]`; campos opcionales son punteros, nunca `0` con default | 🔴 `Conv == nil` (no cargado) ≠ `Conv == []` (vacío) |
| `permisos-gui-human-in-the-loop.md` | enforced | deny-by-default, Write/Edit por diff-approval, `--max-turns` siempre, **sesión aislada por cwd** | 🔴 el cwd cuelga del frente |
| `fe-topologia-fsd.md` | enforced | FSD direccional; ninguna capa importa una superior; widget↛widget; public API por `index.ts` | 🔴 el store se muda de `session-rail` a `chat-dock` |
| `fe-taxonomia-componentes.md` | enforced | canvas⊥chrome; `shared/ui` no toca dominio; **widget↛widget** | 🔴 idem |
| `fe-transporte-independiente.md` | enforced | `entities/*/model/**` no importa `shared/api`; SSE singleton solo en `app/realtime/**`; tipos de `shared/api/generated/**` derivan del OpenAPI | 🔴 |
| `fe-visual-fitness.md` | enforced | cada componente de `shared/ui` + widget tiene story = test; axe a11y; sin Chromatic | 🔴 |
| `superficie-local-confinada.md` | enforced (9 checks) | host-gate loopback, Origin allowlist, cero `ACAO: *`, token constante-time | 🟡 endpoints nuevos |
| `fe-tokens-contrato.md` | enforced | tokens DTCG = SSoT; cero magic-value en CSS | 🟡 UI nueva |
| `core-no-importa-shell.md` | enforced | `internal/{domain,usecase,ports}` no importan shell/Tauri | 🟡 |
| `adaptadores-de-agente-intercambiables.md` | enforced | Claude Code tras `AgentPort`; cero fuga de términos stream-json fuera del adapter | 🟡 |
| `peso-del-binario-es-presupuesto.md` | enforced | una dependencia se MIDE antes de adoptarse | 🟡 si se suma un índice de texto |

**Fuera de alcance** (los 9 restantes): `orquestacion-determinista-entre-cajas`,
`permisos-derivan-del-rol`, `contrato-de-caja-es-fitness-function`,
`portafolio-identidad-y-deriva-honesta`, `marketplace-referencia-es-solo-procedencia`,
`maquinaria-no-contamina-arnes`, `doctrina-una-fuente-dos-targets`, `telemetria-de-nacimiento`,
`cifra-viaja-con-su-confianza`.

### 6.2 🔴 `conductor-no-parsea-jsonl.md` — qué exige del código nuevo

`version: 1.1` · `status: proposed` · `severity: high` (`:2-20`). Enforcers:
`arch_test.go:TestNoJSONLSchemaParsing` + `arch_test.go:TestLiveEventsFromStreamJSON` (`:17-19`).

**Los tres roles fijos** (`:35-47`):

- **stream-json = live/primario.** «El adaptador conductor consume NDJSON de stdout y **captura el
  stream a NUESTRO event store** — ese store es la API estable interna» (`:37-38`).
- **OTel = sidechannel** para hooks/skills/tool_decision/auth (`:41-43`).
- **JSONL = enumerar/replay**: «**nunca parsear su schema como contrato**. Es puntero, no API»
  (`:44-45`).
- «`% contexto` es métrica derivada nuestra y se etiqueta como tal» (`:46-47`).

**Veredicto sobre CV-D8: persistir `Conv` y buscar sobre él es LEGAL, y el boundary casi lo ordena.**
El nodo manda capturar el stream a un event store propio (`:37-38`); `Conv` es exactamente eso, y el
dominio ya lo documenta así: «It is NOT the source of truth for the conversation (Claude Code's own
JSONL is)» (`session.go:59-61`). Buscar sobre `Conv` es leer nuestra API interna estable, no el
schema ajeno. **CV-D8 ya lo dice bien** (`decisiones.md:90-92`); conviene anclarlo al check
`no-jsonl-parse` en la spec, porque «persistir el transcript» leído ingenuamente suena a duplicar la
fuente de verdad.

**Lo que sí está prohibido:** decodificar el schema del JSONL **fuera de
`internal/adapters/history`** (único paquete exento, con razón escrita en
`arch_test.go:268-275`); alimentar el dock por tail del JSONL; presentar `% contexto` como medido.

⚠️ **TRAMPA MECÁNICA — falso positivo casi garantizado.** `TestNoJSONLSchemaParsing`
(`arch_test.go:328`) es un **source-scan heurístico**, no semántico. Busca decoders
(`json.Unmarshal`, `json.NewDecoder`, `json.RawMessage` — `arch_test.go:279`) cerca de pistas de
contexto (`"transcript"`, `"jsonl"`, `"~/.claude/projects"`, `".claude/projects"` —
`arch_test.go:283`), **en la misma línea y en la línea previa** (el comentario que la encabeza),
case-insensitive (`:288-317`), sobre `internal/usecase`, `internal/domain` y
`internal/adapters/telemetria` (`:355-357`).

→ Si un archivo nuevo de `domain`/`usecase` documenta `Conv` como «la copia liviana del **transcript**»
justo arriba de un `json.RawMessage`, **el test pinta rojo aunque el código sea 100 % legal**. Dos
salidas legítimas: redactar el comentario sin la palabra-pista adyacente al decoder, o sumar el
prefijo al mapa `jsonlExento` con razón escrita — el propio test dice que eso «exige tocar este test,
que es exactamente el punto» (`arch_test.go:265-267`).

**Riesgo adicional del mismo nodo:** el check `ctx-derivado-etiquetado` («`% contexto` se computa y se
marca como métrica derivada») está hoy `deferred`. **CV-D14 asciende el ctx a chip-disclosure
protagonista.** Si se pinta sin la etiqueta de «derivada», el check pasa de diferido a fail real.

### 6.3 🔴 `codigo-traza-a-capability.md` — R1/R2/R3/R4

`version: 1.5` · **`status: enforced`** · **`severity: error`** (`:2-20`).

| Regla | Qué exige del código nuevo | Rollout | Enforcer |
|---|---|---|---|
| **R1 integridad** (`:47-53`) | cada `pointers:` resuelve a un **archivo** real | **BLOCK** (`cap-ptr-resuelve`, `:73`) | `capability_trace_test.go:TestCapabilityPointersResolve:129` |
| **R1 símbolo** (v1.5) | el `#Símbolo` está **declarado**: Go exacto vía `go/parser`; TS/TSX/Rust por regex de declaración top-level, miembro de interface/object-literal, o especificador de import | **BLOCK** (`cap-ptr-simbolo-resuelve`, `:74`) | `TestCapabilityPointerSymbolsResolve:193` + `capability_symbol_resolve_test.go` |
| **R2 cobertura** (`:54-56`) | **todo** archivo bajo `cmd/`, `internal/`, `web/src/`, `web/src-tauri/src/` reclamado por ≥1 capability, o allowlist **con razón** | **BLOCK** (`cap-sin-huerfano`, `:75`) | `TestCapabilityCoverage:359`; allowlist en `capability_trace_test.go:52-73` + `docs/product/capabilities/_coverage.yaml` |
| **R3 gate-commit** (`:57`) | un commit que toca fuente construye o modifica un capability | **WARN** — «feedback local, no bloquea CI» (`:76`) | `lefthook.yml:28-33`, glob `*.{go,ts,tsx,rs}` |
| **R4 estado** (`:58-60`) | `vivo`/`parcial` ⟹ tiene `valida:`; `vivo·nc`/`stub` ⟹ **sin** `valida:` | **WARN** (`cap-estado-consistente`, `:77`) | `TestCapabilityStatusConsistent:318` |
| **R4 puntero estable** | punteros `file#Símbolo` o `paquete/`, **nunca `file:línea`** | **INFO** (`cap-puntero-estable`, `:78`) | `TestCapabilityPointersStable:342` |

**Deuda declarada** (`:60,65-67`): la derivación LIVE (`vivo ⟺ check verde` corriendo el test) **no
existe** — está en BACKLOG. «NUNCA pass fabricado».

**Qué exige, concretamente, de este paquete:**

1. Renombrar o partir `Session` **rompe R1-símbolo** en todo puntero
   `internal/domain/session.go#Session`. Los capabilities afectados están en §7.
2. Cada archivo nuevo (`internal/domain/conversacion.go`, `internal/adapters/store/…`,
   `web/src/widgets/chat-dock/model/…`) necesita hoja de capability o entrada en `_coverage.yaml`
   **antes** de commitear, o R2 pinta rojo.
3. Mover `conversaciones-store.ts` de `session-rail/model/` a `chat-dock/` rompe R1 sobre
   `docs/product/capabilities/usecases/historial-de-conversaciones.yaml:14-15` **y** su
   `valida: conversaciones-store.test.ts` (`:21`).
4. Si el paquete agrega tests a capabilities hoy `vivo·nc`, **R4 exige flipear `status` a `vivo` en el
   mismo commit**.

### 6.4 🔴 `sesion-viva-consistente.md` — el nodo que gobierna lo que se parte

`status: enforced`, severity high. Su L2 nombra `internal/usecase/session_service.go` (`:48`).
Cuatro propiedades, cada una con enforcer:

| Check | Qué exige | Enforcer |
|---|---|---|
| `un-turno-a-la-vez` | `Turn` rechaza con `ErrBusy` si `status == streaming` → transporte 409 | `arch_test.go:TestOneTurnAtATime:1911` |
| `frames-idempotentes-run-id` | todo `dockFrame` lleva `run_id`; el FE trackea `finalizedRun[id]` y dropea frames de un run terminado | `arch_test.go:TestFramesCarryRunID:1924` |
| `sin-perdida-silenciosa` | conductor y broker no dropean frames | `arch_test.go:TestNoSilentEventDrop:1699` |
| `resume-auto-sana` | `tryHealResume`: spawn con `--resume` que muere antes de `init` ⇒ limpia el `ClaudeSessionID` stale y respawnea fresh **una vez** | `arch_test.go:TestResumeAutoSana:1957` |

**CV-D11 cae de lleno acá.** Los 4 invariantes están hoy implementados **por sesión**; con N
conversaciones hay que decidir el nivel al que aplican (§5.5). Es el boundary de mayor severidad que
el paquete toca, y su L2 apunta al archivo que se parte. Probable bump de versión del nodo.

Adyacente: `arch_test.go:TestSessionSpawnsInArnesPath:1891` («cada sesión spawnea en la ruta de SU
arnés») y `TestArnesPathContainment:2002`.

### 6.5 `indice-desechable-jsonl-es-verdad.md` — aclaración importante

`version: 1.3` · `status: proposed` · high. La v1.2/v1.3 aclara explícitamente (`:47-62`):

> «La fuente durable real **NO es literalmente la conversación JSONL de `~/.claude/projects`** […] El
> índice guarda `domain.Graph` […] **La conversación JSONL de sesiones de chat es un corpus SEPARADO,
> con su propio adapter de solo-lectura (`internal/adapters/history/reader.go`) que no alimenta este
> índice.**»

**Traducción para el paquete: persistir conversaciones en `~/.arnesia/` NO cae bajo este boundary**
mientras se use el store JSON atómico. Los que aplican son `ingesta-por-allowlist-declarada` y
`no-aplica-no-es-cero`.

⚠️ **Pero revive una deuda:** la decisión D1 del nodo (`:72-77`) dejó los cursores
`{path,inode,size,offset}` fuera de fase 5, con la nota de que «revive como deuda futura si el índice
algún día indexa algo que sí crece por apéndice». **Un corpus de conversaciones crece por apéndice.**

Si en cambio se abriera una base propia, el precedente es telemetría: componente propio,
`canUse: [sqlite]`, y la nota de `.go-arch-lint.yml:198-200` de que «`index` y `telemetria-store` no
se conocen: son dos bases con doctrinas opuestas (una desechable, la otra irrecuperable)».

Sus 4 checks: `index-reconstruible` (error), `sin-migracion-incremental` (warn),
`writer-serializado` (warn), `sin-cgo` (error).

### 6.6 El grafo de imports — `.go-arch-lint.yml` (rompe CI por omisión)

**v3 = ALLOW-LIST, default-deny**: «todo import interno no listado en `mayDependOn` está prohibido»
(`.go-arch-lint.yml:6-7`). `depOnAnyVendor: false`, `deepScan: false` (`:15-19`).

- 🔴 Si el paquete crea `internal/adapters/<algo>/`, **hay que agregarle `components:` + `deps:`** o
  el linter lo rechaza por ausencia. Precedente literal: `logfile` «NO tiene entrada acá a propósito:
  ausencia = default-deny total» (`:184-187`).
- 🔴 Si algo nuevo entra por `cmd`, sumarlo a la línea `cmd: mayDependOn: [...]` (`:212`).
- Nota: `history: mayDependOn: [domain]` (`:176-177`) — `history` **no puede importar `ports`**.

Corre en CI como paso propio (`.github/workflows/ci.yml:29-30`).

### 6.7 Fitness tests — los que este cambio puede romper

`docs/architecture/fitness/`: 8 archivos, **75 funciones `Test*`** (generado).

| Test | Archivo:línea | Riesgo |
|---|---|---|
| `TestCapabilityPointerSymbolsResolve` | `capability_trace_test.go:193` | 🔴 **rompe seguro** al partir `Session` |
| `TestCapabilityPointersResolve` | `capability_trace_test.go:129` | 🔴 **rompe seguro** |
| `TestCapabilityCoverage` | `capability_trace_test.go:359` | 🔴 **rompe seguro** con archivos nuevos |
| `TestCapabilityStatusConsistent` | `capability_trace_test.go:318` | 🔴 si se agregan `valida:` sin flipear `status` |
| `TestCapabilityPointersStable` | `capability_trace_test.go:342` | 🟡 |
| `TestNoJSONLSchemaParsing` | `arch_test.go:328` | 🔴 **falso positivo probable** (§6.2) |
| `TestOneTurnAtATime` | `arch_test.go:1911` | 🔴 §6.4 |
| `TestFramesCarryRunID` | `arch_test.go:1924` | 🔴 §6.4 |
| `TestResumeAutoSana` | `arch_test.go:1957` | 🔴 §6.4 |
| `TestNoSilentEventDrop` | `arch_test.go:1699` | 🔴 §6.4 |
| `TestSessionSpawnsInArnesPath` | `arch_test.go:1891` | 🔴 el sujeto es «sesión» |
| `TestDomainIndependentOfTransport` | `arch_test.go:206` | 🔴 entidades nuevas en `domain` |
| `TestLiveEventsFromStreamJSON` | `arch_test.go:398` | 🔴 si se toca el adapter conductor |
| `TestDominioNoAdoptaShapeAjeno` | `marketplace_shape_test.go:32` | 🟡 hoy solo escanea `domain/marketplace*.go`, **pero es el molde exacto** para un enforcer «Conversación no adopta el shape del JSONL nativo» |
| `TestIndexRebuildsFromJSONL` / `TestSchemaVersionTriggersRebuild` / `TestWriterSerializedSingleConn` | `arch_test.go:489`, `:538`, `:588` | 🟡 |
| `TestAgentPortHasNoConcreteLeak` | `arch_test.go:222` | 🟡 |
| `TestVersionManifestsInSync` | `arch_test.go:2027` | 🟡 si se bumpea |
| `TestChangelogExisteYTieneForma` / `…CubreLaVersion…` | `changelog_test.go:96`, `:163` | 🟡 obligatorio al bumpear |

Precedente de layout útil (`hs17_config_source_test.go:1-6`): se pueden poner tests nuevos en un
archivo propio (`conversacion_test.go`) porque «la discovery de Go es por paquete, no por nombre de
archivo»; el `arch_test.go:TestX` de los `enforced_by` es una convención que el parser matchea
**textualmente**.

### 6.8 Convenciones

`docs/architecture/conventions/INDEX.md:17-18`: «Regla dura: **toda convención en este árbol DEBE
romper CI** (si no, es una nota)». 9 nodos · 31 checks.

| Nodo | Aplica | Config |
|---|---|---|
| `go-style.md` | 🔴 | `.golangci.yml` (~30 linters) |
| `naming.md` | 🔴 **el split ES un ejercicio de naming** | `.golangci.yml` revive + `web/biome.json` `useNamingConvention` `strictCase` |
| `ts-style.md` | 🔴 | Biome v2.4 |
| `ts-types.md` | 🔴 | `@tsconfig/strictest` |
| `versionado.md` | 🔴 si se publica | 9 checks; `changelog-por-version` (error), `sync-3-manifiestos` (error), `dev-daemon-sincronizado` (warn) |
| `commits.md` / `git-hooks.md` / `ci.md` | 🟡 | lefthook + CI |
| `editor.md` | ⚪ | `.editorconfig` |

**Precedente de naming del árbol:** los nombres de dominio van **en español sin tildes** (`Salud`,
`Rol`, `Reparacion`, `CadenaCC`, `RotacionPendiente` — `session.go:32-115`; paquetes `portafolio`,
`traer`, `conformance`). `Conversacion` es consistente.

### 6.9 `docs/architecture/knowledge/` — no aplica al código de la app

12 nodos · 138 checks. Su INDEX (`knowledge/INDEX.md:1-11`) lo declara: es «la metodología as code»
sobre cómo estructurar los **elementos de un arnés** — **dominio-producto de ArnesIA, NO arquitectura
de la app**. Cadencia semanal por barrido de ecosistema (`knowledge/CADENCE.md:1-7`), opuesta a la
event-driven de los boundaries.

🟡 Único punto de contacto: `elements/headless-sdk.md` es corregido explícitamente por
`conductor-no-parsea-jsonl.md:48-51` — «`--bare` **rompe el auth de suscripción**… `-p` hará `--bare`
default futuro → **pinear flags explícitos**». Sigue vigente si el paquete toca el spawn.

Nota operativa: el parser del motor lee `knowledge/elements` + `boundaries` + `conventions`
(`internal/adapters/conformance/ruleset/parser.go:38-46`, salteando `INDEX*`/`CADENCE*` en `:65-68`).
**Un boundary nuevo entra al motor por existir como `.md` con tabla `Checklist evaluable`** — cero
recompilación (`parser.go:1-7`).

---

## 7. Capabilities

`docs/product/capabilities/` — **139 hojas** (`ls docs/product/capabilities/*/*.yaml | wc -l`), 18
módulos. Distribución: **84 vivo · 40 vivo·nc · 13 parcial · 2 stub** (`checkpoint.md:190`).

**Aclaración de nomenclatura** (importante para escribir hojas): los campos reales del schema son
`pointers:` (la traza), `status:` (el estado, enum `vivo|vivo·nc|parcial|stub`) y `valida:` (los
tests). **No existen los campos `afirma`, `traza` ni `estado`** — «qué afirma» vive en `name` +
`scenarios[]` + `business_rules[]` + la prosa markdown bajo el segundo `---`. Template canónico:
`docs/product/_templates/capability.template.yaml`. Mejor ejemplo del schema nuevo completo:
`docs/product/capabilities/fe-chat/dictar-en-el-composer.yaml:1-173`.

⚠️ **`capabilities/INDEX.md` está STALE en 22 hojas** — le falta el módulo `telemetria` entero (17) y
desactualiza `cli-daemon`, `conductor` (+2), `http-sse` (+1) y `fe-mapa` (+1). Suma 117 vs 139 reales.
Causa: `GROUP_ORDER` en `scripts/cap_doctor.py:90-94` no incluye `telemetria`. Fix:
`python3 scripts/cap_doctor.py --index`. **Correrlo antes de empezar.**

### 7.1 Las capabilities que este paquete toca

| CAP | archivo | status | Qué afirma | `pointers` |
|---|---|---|---|---|
| **CAP-14** | `dominio-l0/sesion-frente-de-trabajo.yaml:5` | `vivo·nc` | «Sesión = frente de trabajo». **Cuerpo genérico HS-18; `valida`/`scenarios`/`business_rules` VACÍOS** | `internal/domain/session.go#Session`, `#Salud` (`:11-13`) |
| **CAP-25** | `indice-persistencia/persistencia.yaml:5` | `vivo·nc` | «Persistencia (sesiones + arnés→path, JSON atómico)». **Cero evidencia, cero afirmación** (`:14-16`) | `store/registry.go#Save`, `store/arnes_registry.go#saveLocked` |
| **CAP-59** | `usecases/gestion-de-sesiones-crud.yaml:5` | `vivo·nc` | «Gestión de sesiones CRUD»; `valida: []` (`:15`) | `session_service.go#Create`, `#Rename`, `#Close` |
| **CAP-53** | `usecases/dock-conversacion-en-vivo-multisesion.yaml:5` | `vivo` | «Dock: conversación en vivo multisesión» | `session_service.go#Turn`, `#spawnLocked`, `#tryHealResume`; `valida: session_permisos_test.go` |
| **CAP-54** | `usecases/gobierno-del-turno.yaml:5` | `vivo` | «Gobierno del turno (permisos HITL + interrupt)» | `#onControlRequest`, `#ResolvePermission`, `#Interrupt` |
| **CAP-97** | `usecases/rotacion-de-contexto.yaml:5` | `vivo` | «Rotación de contexto invisible (conversación infinita)». Cuerpo (`:29-34`): ctxPct real por turno → umbral → proceso fresco con checkpoint; «`Session.ID` no cambia; el usuario ve UNA conversación continua» | `session_rotacion.go` (+test); `valida`: `TestCtxHistYUmbralRotacion`, `TestRotacionInvisible`, `TestCtxPctUsaUltimoUsage` |
| **CAP-98** | `usecases/historial-de-conversaciones.yaml:5` | `vivo` | 🔴 **«Historial de conversaciones por arnés (B2: JSONL nativa = verdad)»**. Cuerpo (`:33-38`): «`Close()` archiva la metadata liviana […] **sin `Conv`: la JSONL nativa es la verdad**» + los 2 endpoints + «FE: sección Conversaciones del picker» | `session_historial.go`, `history/reader.go`, `session-rail/model/conversaciones-store.ts` (+test); `valida` (`:16-21`): `TestCloseArchivaMetadata`, `TestHistorialCerradaCoseCadena`, `TestTurnosLeeCorpus`, `TestDirParaCwd`, `conversaciones-store.test.ts` |
| **CAP-96** | `provisioning/tarjeta-identidad-por-sesion.yaml:5` | `vivo` | system-prompt POR SESIÓN en `~/.arnesia/sessions/<id>/system.md`; «Misma infra que usará el checkpoint de rotación» | `session_grounding.go`, `provisioner.go#ProvisionSession` |
| **CAP-72** | `fe-shell/rail-de-sesiones.yaml:5` | `vivo·nc` | «Rail de sesiones (crear/renombrar/cerrar/**switch**)». Scenario `:21` describe el picker inline 224→360px; prosa `:33-37`: «El historial previo (`useConversaciones.cargar`) sigue la misma clave» | `session-rail.tsx#SessionRail`, `new-session-picker.tsx#NewSessionPicker`, `portafolio-picker-store.ts#usePortafolioPicker`, … ; `valida: []` (`:18`) |
| **CAP-74** | `fe-shell/topbar-breadcrumb-k-dock.yaml:5` | `vivo·nc` | breadcrumb sin empresa ni `parked`; arnés = etiqueta de solo lectura; «Conversar ⌘K baja a su propia línea» (`:15-16`) | `topbar.tsx#Topbar`; `valida: []` |
| **CAP-68** | `fe-chat/chat-cc.yaml:5` | `vivo·nc` | «Chat CC (turno/interrupt)» — cuerpo genérico HS-18, `valida: []` | `chat-dock.tsx#ChatDock`, `sessions-store.ts#sendTurn` |
| **CAP-69** | `fe-chat/acotar-alcance.yaml:5` | `vivo·nc` | «Acotar alcance (nodo→chip)» | `chat-dock.tsx#ScopeRow`, `sessions-store.ts#setScope` |
| **CAP-70** | `fe-chat/decidir-permisos.yaml:5` | `vivo` | «Decidir permisos (tarjeta inline)» | `permission-card.tsx#PermissionCard`; `valida: permission-card.stories.tsx` |
| **CAP-71** | `fe-chat/gate-de-conformance-tras-escrituras.yaml:5` | `vivo·nc` | «Gate de conformance tras escrituras (RF-117)» | `sessions-store.ts` (archivo entero, **sin símbolo**) |
| **CAP-100** | `fe-chat/conversacion-legible.yaml:5` | `vivo` | «Conversación legible: actividad visible · burbuja por paso · markdown». Cuerpo (`:35-41`): «**Persistido en Conv → el historial rehidrata idéntico**» | `conductor.go#assistantEvents`, `session_service.go#consume`, `chat-dock.tsx#ActivityCard`, `markdown.tsx#Md`; 4 tests Go |
| **CAP-99** | `usecases/paquete-cerrado-en-el-gate.yaml:5` | `vivo` | el gate deniega el árbol propio de ArnesIA, antes del auto-allow y del click humano | `session_service.go#ProtegerPaqueteCerrado`, `#refiereAlguno` |
| **CAP-52** | `http-sse/superficie-rest.yaml:5-7` | `vivo·nc` | «Superficie REST (23 endpoints)» — pointer único `router.go#NewHandler`, `valida: []` | — |
| **CAP-115** | `fe-chat/dictar-en-el-composer.yaml:5` | `vivo` | dictado en el composer; 9 scenarios BDD | `dictado-button.tsx`, `chat-dock.tsx#Composer`, … |

**Gaps notables:**

1. **«switch de sesión» no tiene capability propia** — solo aparece en el `name` de CAP-72. Sin
   scenario, sin test. Y es el comportamiento que CV-D11 extiende a conversaciones.
2. **CAP-14 (la entidad que se parte) está vacía**: `valida`, `scenarios` y `business_rules` todos
   vacíos, cuerpo genérico. **No hay red de tests que sostenga el split.**
3. **CAP-25 (la persistencia) está igual de vacía** y su adaptador tiene 0 tests (§8.1).
4. **CAP-98 afirma literalmente lo contrario de CV-D8.** Su cuerpo (`:33-34`) dice «sin `Conv`: la
   JSONL nativa es la verdad». Eso es un **cambio de business rule**, no un delta de implementación:
   hay que reescribir el cuerpo de la hoja, no solo sumarle un pointer. Y sus 5 `valida:` se rompen.

### 7.2 Capabilities NUEVAS o MODIFICADAS que este paquete exige (propuesta, no escrita)

**Nuevas:**

| Propuesta | Módulo | Cubriría |
|---|---|---|
| `conversacion-como-entidad` | `dominio-l0` | la entidad `Conversacion` y su relación 1:N con `Sesion` (CV-D3) |
| `conversaciones-en-el-dock` | `fe-chat` | lista + buscador + `＋` dentro del dock (CV-D2/D4) |
| `retomar-conversacion-inactiva` | `usecases` | `--resume` sobre una inactiva; una activa por sesión (CV-D7/D11/D12) |
| `buscar-en-el-transcript` | `usecases` o `fe-chat` | scan sobre `Conv` (CV-D8) |
| `titulo-de-conversacion` | `usecases` | auto-derivado + editable (CV-D9) |
| `migracion-de-esquema-en-disco` | `indice-persistencia` | el mecanismo que hoy no existe (§10.3) |
| `ultima-interaccion-por-turno` | `dominio-l0` | el timestamp nuevo (CV-D13, §10.1) |

**Modificadas (obligatorio):**

| CAP | Qué cambia | Regla que fuerza |
|---|---|---|
| **CAP-98** | reescribir el cuerpo: `Conv` **sí** se persiste; el registro de cerradas se resignifica (CV-D12); los 2 endpoints por-arnés se reemplazan; el store FE se muda | R1 (pointers), R4 (status/valida) |
| **CAP-14** | pointers al nuevo `session.go#Sesion` + `conversacion.go#Conversacion`; poblar scenarios | R1-símbolo |
| **CAP-72** | quitar `ConversacionesDelArnes` del alcance; corregir la prosa `:33-37` sobre la clave | R1 |
| **CAP-68/69/71** | el dock gana superficie; los pointers a `chat-dock.tsx#…` cambian | R1-símbolo |
| **CAP-53/59/54** | el turno, el CRUD y el gobierno pasan a colgar de la conversación | R1-símbolo |
| **CAP-97** | la rotación cuelga de la conversación (CV-D10) | R1 |
| **CAP-100** | «persistido en Conv → el historial rehidrata idéntico» se vuelve **verdad también para las inactivas** | prosa |
| **CAP-96** | el system-prompt por sesión vs. por conversación (§10.6) | R1 |
| **CAP-25** | formato nuevo + migración | R1, R4 |
| **CAP-52** | «23 endpoints» cambia de número | prosa |

⚠️ **CAP-72, CAP-74, CAP-68, CAP-69, CAP-71, CAP-59, CAP-14, CAP-25 son todas `vivo·nc` con
`valida: []`.** Si el paquete les agrega tests, **R4 exige flipear el `status` a `vivo` en el mismo
commit** o `TestCapabilityStatusConsistent` falla.

---

## 8. Tests e infraestructura de verificación

### 8.1 Go — cifras generadas

`go test ./...` corrido: **26 paquetes `ok`, 8 `[no test files]`, 0 `FAIL`**.
**622 funciones `Test*` en 82 archivos** (`grep -rh "^func Test"`).

| Directorio | Tests | Archivos |
|---|---|---|
| `internal/adapters` | 338 | 47 |
| `internal/usecase` | 126 | 16 |
| `internal/domain` | 75 | 9 |
| `docs/architecture/fitness` | 75 | 8 |
| `cmd/arnesia` | 8 | 2 |
| **Total** | **622** | **82** |

**Tests del área:**

| Archivo | Tests |
|---|---|
| `internal/usecase/session_historial_test.go` | `TestCloseArchivaMetadata:38`, `TestHistorialCerradaCoseCadena:82` |
| `internal/usecase/session_rotacion_test.go` | `TestCtxHistYUmbralRotacion:16`, `TestRotacionInvisible:67` |
| `internal/usecase/session_burbujas_test.go` | `TestTurnoSeParteEnBurbujasPorActividad:15`, `TestResultSinMensajesConservaFallback:55` |
| `internal/usecase/session_permisos_test.go` | 5 tests (`:103`-`:221`), incl. `TestControlRequestSobrePaqueteCerradoSeDeniegaSinTarjeta:180` |
| `internal/usecase/session_grounding_test.go` | 4 tests (`:29`-`:76`) |
| `internal/usecase/session_reindex_test.go` | 4 tests (`:43`-`:112`) |
| `internal/adapters/agent/claudecode/conductor_test.go` | 8 tests, incl. `TestUserTurnWireSinCamposExtra:54`, `TestCtxPctUsaUltimoUsage:216` |
| `internal/adapters/history/reader_test.go` | `TestDirParaCwd:13`, `TestTurnosLeeCorpus:24` |
| `internal/adapters/transport/http/sessions_test.go` | **1 solo**: `TestCreateSessionIndexaAlRegistrar:42` |

🔴 **Tres huecos de cobertura en el área exacta del paquete:**

1. **`internal/adapters/store` → `[no test files]`.** La persistencia JSON que se va a reformatear
   **no tiene un solo test**.
2. **`internal/usecase/session_service.go` (953 líneas) no tiene `session_service_test.go`.** Se cubre
   indirectamente desde los 6 archivos temáticos.
3. **`sessions_test.go` cubre 1 de 10 endpoints.** El gap más grande del lado Go.

### 8.2 Cómo se corre — el Makefile

**7 targets** (`Makefile:50`): `version` (`:52`), `changelog` (`:55`), `bump-patch` (`:61`),
`bump-minor` (`:64`), `bump-major` (`:67`), `installer` (`:70`), `dev-sync` (`:87`).

🔴 **NO HAY TARGET DE TESTS.** El Makefile es exclusivamente de versionado/empaquetado. Los tests se
corren a mano: `go test ./...` y `pnpm exec vitest --project=…`.

### 8.3 El motor `arnesia conformance`

Invocación (`cmd/arnesia/conformance.go:33-40`, wireado en `main.go:66`):

```
arnesia conformance [<elemento> | --arnes <path> | --todo] [--root <dir>] [--json]
```

Ruleset desde el repo si existe, `go:embed` si no (`conformance.go:46-63`). 5 adapters de mecanismo
(`:71-77`): `ArchTest`, `GoArchLint`, `NLJudge`, `StaticScan`, `SchemaAdapter`. Exit 1 si
`!report.OK()` (`:100-102`).

**Cifras vivas generadas hoy:**

```
conformance todo    → 311 checks · pass 81 · fail 0 · error 0 · deferred 230 · n/a 0   (exit 0)
conformance arnes:dev-full-cycle → 21 checks · pass 20 · fail 1 · error 0 · deferred 0
```

**230 de 311 (74 %) están `deferred`** — checks de juicio sin enforcer determinista (nl-judge). El
`fail 1` del dogfood es el warn honesto `art-es-path`, declarado en `checkpoint.md:187`.

⚠️ **`arnesia conformance` NO corre directo en CI.** Entra indirecto vía
`scripts/estado.sh --check` (`ci.yml:35-39`), y solo para detectar **drift de cifras** en el
checkpoint, no para bloquear por veredicto. El gate duro de facto sobre los boundaries es
**`go test ./...` (que incluye `docs/architecture/fitness/`) + `go-arch-lint`**.

### 8.4 Frontend

**`web/vitest.config.ts:17-42` define 2 proyectos** (Vitest 4 quitó `defineWorkspace`):

| Proyecto | Líneas | Entorno | Include |
|---|---|---|---|
| `unit` | `:20-27` | `environment: "node"` | `src/**/*.test.ts` |
| `storybook` | `:28-40` | browser **Chromium headless**, `provider: playwright()` | `*.stories.tsx` vía `storybookTest({configDir: ".storybook"})` |

Doctrina en el propio archivo (`:7-8`): «story = test, enforced_by `fe-visual-fitness.md` ·
`ci.yml:visual-fitness`».

**Versiones** (`web/package.json`): `vitest ^4.1.9` (`:67`), `@vitest/browser ^4.1.9` (`:55`),
`@vitest/browser-playwright 4.1.9` pin exacto (`:56`), `playwright ^1.61.1` (`:58`),
**`storybook ^10.4.6`** (`:60`), `@storybook/addon-a11y ^10.4.6` (`:46`), `vite ^8.1.3`,
`typescript ^6.0.3`. `pnpm@9.15.9`, `engines.node >=20.16`.

**Scripts** (`:11-26`): `test` = `vitest run` (**los dos proyectos**, `:24`); `verify` =
`typecheck && lint && depcruise && fsd && stylelint` (**sin tests**, `:25`).

**Stories: 42 archivos / 383 `export const`** (generado). Los del área:

| Archivo | Stories | `play()` | a11y |
|---|---|---|---|
| `chat-dock/ui/dictado-button.stories.tsx` | 13 | 13 | `todo` (`:62`) |
| `chat-dock/ui/permission-card.stories.tsx` | 6 | 6 | `todo` (`:11`) |
| `session-rail/ui/new-session-picker.stories.tsx` | 7 | 7 | **`error`** (hereda el global) |

🔴 **`chat-dock.tsx` (375 líneas) NO tiene stories.** El widget central del dock — el que este paquete
rediseña — **no está cubierto por el gate visual**. Tampoco `markdown.tsx`.

**a11y/axe:** el default global es **bloqueante** — `web/.storybook/preview.ts:7`:
`a11y: { test: "error" }`. Hay 14 escapes a `test: "todo"` en 13 archivos (incluidos los 2 del
chat-dock). **28 de 42 archivos corren axe en modo error.**

**Unit tests: 8 archivos / 122 tests**, incluido
`web/src/widgets/session-rail/model/conversaciones-store.test.ts` (51 líneas, 3 tests) — el
precedente directo:
- `:14` «cargar puebla vivas + cerradas y conserva el error honesto de cerradas»
- `:28` «abrirHistorial trae turnos + faltantes honestos»
- `:43` «cargar con fallo deja el error visible, jamás datos fabricados»

**Corridas reales:**

```
vitest --project=unit run       → 8 files / 122 tests PASSED (292ms)          ✅
vitest --project=storybook run  → 1 file FAILED / 41 passed; 4 tests FAILED   ❌
```

🔴 **Los 4 fallos caen exactamente sobre el widget que este paquete toca**
(`new-session-picker.stories.tsx > Caso Simple / Caso Ambiguo / Colision / Cancelar`). Causa: axe
`color-contrast` sobre el `text-warn` del **historial de conversaciones cerradas**:

```
<span class="text-warn">historial de cerradas: Failed to fetch</span>
contrast 3.76 (fg #c96a2e, bg #ffffff, 10px). Expected 4.5:1
```

El elemento es `new-session-picker.tsx:273`, dentro de `ConversacionesDelArnes` (`:263`) — **el bloque
que CV-D2 elimina**. Hay un segundo `text-warn` igual en `:290`. Ver §9.5.

### 8.5 E2E y app instalada

**No hay runner E2E de Playwright.** No existe `playwright.config.ts`. Playwright entra **solo** como
provider de browser de Vitest (`vitest.config.ts:2`, `:35-37`) y se instala en CI
(`ci.yml:68`: `pnpm exec playwright install --with-deps chromium`).

Lo que sí hay, **manual/ad-hoc**:

| Artefacto | Qué es |
|---|---|
| `.playwright-mcp/` | **153 PNG** de sesiones E2E vía el MCP de Playwright. Evidencia visual, no suite reejecutable |
| `docs/product/stories/2026-07-08-chat-cc-funcional/e2e/casuistica.mjs` | script E2E de casuística del chat |
| `docs/product/stories/2026-07-08-chat-cc-funcional/e2e/mock-claude.sh` | mock ejecutable de `claude` |
| `docs/product/stories/2026-07-08-chat-cc-funcional/e2e/real.mjs` | corrida contra Claude real |
| `internal/adapters/stt/local/local_e2e_test.go` | único `_test.go` E2E (STT) |

→ **La verificación contra la app instalada es humana + MCP-Playwright + screenshots. No hay gate
automatizado.**

**`make installer`** (`Makefile:70-85`), depende de `bump-patch`: limpia el bundle → `bundle.sh` →
`instaladores/vX.Y.Z/` → falla duro si no salió instalador (`:76`) → `sha256sum` (`:77`) → **avisa de
`make dev-sync` si existe `~/.local/bin/arnesia`** (`:80-85`).

**`scripts/bundle.sh`** (65 líneas, 3 pasos): SPA (`:21-24`), daemon con
`-X …selfupdate.Version/Build/Compilado` (`:26-48`), Tauri + sidecar (`:55-63`). Modo
`--daemon-only` (`:17-18`, `:50-53`) corta tras el paso 2.
⚠️ **`bundle.sh` no corre ningún test.** Nada impide bundlear con la suite roja.

**`~/.local/bin/arnesia`** = `DEV_DAEMON` (`Makefile:43`), el **override local del self-update**: el
shell Tauri instalado (`web/src-tauri/src/lib.rs`, `override_local`) **prefiere SIEMPRE ese archivo
sobre el sidecar empaquetado** si existe. Ver §9.2.

**`make dev-sync`** (`Makefile:87-102`): sin override sale con 0 y avisa (`:88-91`); con override →
`bundle.sh --daemon-only` → `install -m755 … .new` → `mv -f` atómico (`:94-95`) → `pkill -f` con el
truco del corchete `[a]rnesia` para no auto-matarse (`:96-101`).

**Bundles:** `instaladores/` con **23 carpetas versionadas**, de `v0.2.2` a **`v0.2.24`** (la de
HEAD), con `.deb` + `.AppImage` + `.rpm` + `checksums.txt`. Vía paralela: `.goreleaser.yaml` publica
**solo el binario standalone**, y aclara (`:2-4`) que los instaladores de escritorio salen de
`bundle.sh`.

### 8.6 Gates: lefthook y CI

**`lefthook.yml`** — 2 stages, 8 jobs. **NO existe stage `pre-push`** (ver §9.6).

| Stage | Job | Glob | Corre |
|---|---|---|---|
| pre-commit | `go-lint` (`:10-14`) | `*.go` | `golangci-lint run --new-from-rev=HEAD ./...` |
| pre-commit | `go-fmt` (`:15-17`) | `*.go` | `golangci-lint fmt {staged_files}` |
| pre-commit | `ts` (`:18-23`) | `*.{ts,tsx}` | `biome check --write --staged` |
| pre-commit | `rust` (`:24-27`) | `*.rs` | `cargo fmt` |
| pre-commit | **`capabilities`** (`:28-33`) | `*.{go,ts,tsx,rs}` | `go test ./docs/architecture/fitness/ -run 'TestCapability(PointersResolve\|Coverage)'` (~4 ms) |
| pre-commit | `changelog` (`:34-40`) | manifiestos | 2 tests de changelog (~30 ms) |
| pre-commit | `estado-cifras` (`:41-52`) | `*.{go,ts,tsx,rs,yaml,md}` | `estado.sh --check`; si driftó regenera **+ `git add`** (~20-50 s) |
| commit-msg | `conventional` (`:54-58`) | — | regex `^(feat\|fix\|docs\|refactor\|chore\|test\|perf\|build\|ci)(\(.+\))?: ` |

**Lo que lefthook NO corre:** `go test ./...`, ningún vitest, `tsc`, `depcruise`, `steiger`,
`stylelint`. Cabecera honesta (`lefthook.yml:4-5`): «lefthook = feedback rápido (**saltable con
`--no-verify`**); el gate DURO es CI».

⚠️ **Y hoy ni siquiera se dispara**: `.git/hooks/` contiene solo los 14 `.sample` de git y
`core.hooksPath` no está seteado → **lefthook nunca corre en este working copy**, pese a estar
instalado (v1.13.6). Falta `lefthook install`.

**`.github/workflows/ci.yml`** — 3 jobs paralelos, steps secuenciales fail-fast:

| Job | Steps | Cita |
|---|---|---|
| **`go`** | `golangci-lint` → `fmt --diff` → **`go-arch-lint`** → **`go test ./... -race`** → `go build` → **`estado.sh --check`** | `:19-39` |
| **`ts`** | `biome ci` → `tsc --noEmit` → `depcruise` → `steiger` → `stylelint` → **`vitest --project=unit`** → **`playwright install chromium && vitest --project=storybook`** (id `visual-fitness`) → tokens diff → openapi diff (**condicional, inerte**) | `:41-77` |
| **`rust`** | build del sidecar → apt webkit2gtk 4.1 → `cargo clippy -- -D warnings` → `cargo fmt --check` | `:79-98` |

🔴 **Estado REAL de CI:** última corrida verde **2026-07-17**; **19 de las últimas 20 corridas en
rojo**; rojas consecutivas desde 2026-07-20. El step que rompe es
**`fitness visual (Storybook 10 = tests)`** (`ci.yml:66-68`) — **exactamente los 4 fallos de
`new-session-picker`** reproducidos local. Además hay **79 commits locales sin pushear**
(`git rev-list --count origin/main..HEAD`).

→ **Este paquete no tiene hoy una línea base verde de CI contra la cual medirse**, y el widget que
rompe es el que va a tocar.

---

## 9. Riesgos y trampas conocidas

### 9.1 🔴 El re-key `cca314a` — la premisa del paquete está mal citada

`git show --stat cca314a`: 21 archivos, +306/−200, 2026-07-23. Introdujo `clave` explícita en
`IndexPort.Upsert` y sus 3 callers.

**No hubo migración de datos. Cero.** El mensaje del propio commit:

> «No route or persistence-format changes: `{id}` stays an opaque string […] ArnesRegistry/Session.Arnes
> on disk are untyped strings — **old sessions keep resolving under their old bare id, nothing to
> migrate**.»

Confirmado en `docs/product/ledger/HS-23.md`: «cero migración de datos persistidos […] las sesiones
viejas con id pelado siguen resolviendo exactamente igual». El único cambio fue en el FE y solo para
sesiones **nuevas**: `new-session-picker.tsx:111` (`cargar(e.clave)`) y `:125` (`arnes: e.clave`).

**🔴 CORRECCIÓN A `decisiones.md:23`**, que dice «El re-key `cca314a` migró las vivas, no las
cerradas». **Es inexacto y la spec no debe heredarlo.** No se migró nada; simplemente las sesiones
creadas *después* nacen con clave calificada.

**Verificado en disco** (`~/.arnesia/sessions.json`, 5 vivas):

| id | `arnes` | conv | ctx_pct |
|---|---|---|---|
| `s25123a2c` | `vitalia` **(pelado)** | 4 | 68 (`rotacion_pendiente: true`) |
| `s6165ac75` | `sin-home~vitalia~vitalia` (calificado) | 90 | 19 |
| `s0fec7798` | `vitalia` **(pelado)** | 0 | — |
| `sfc512b15` | `vitalia` **(pelado)** | 0 | — |
| `s78b3aeeb` | `arnesia` **(pelado)** | 0 | — |

**4 de 5 sesiones VIVAS también están con id pelado.** El split-brain **no es exclusivo de las
cerradas** — está igual de vivo en `sessions.json`.

Confirmado contra el daemon (§3.2): `?arnes=vitalia` → **vivas=3, cerradas=2**;
`?arnes=sin-home~vitalia~vitalia` → **vivas=1, cerradas=0**. La consulta con clave calificada
esconde 3 sesiones vivas, no solo las cerradas.

**Consecuencia:** CV-D6 borra `sesiones-cerradas.json` y declara el bug muerto, pero **no borra
`sessions.json`**. Si el modelo nuevo cuelga conversaciones de `session_id` (CV-D5) pero la *sesión*
sigue llevando `Arnes string` y alguien filtra por clave calificada, esas 4 sesiones vivas siguen
siendo invisibles. El filtro es un `==` desnudo: `session_historial.go:87` y `sessions.go:25`.

Otro punto donde la ley de llaves se filtra: el backfill de `Cwd` en `archivarLocked` depende de
`s.resolver.Resolve(cerrada.Arnes)` (`session_historial.go:51-54`) — y `Resolve("vitalia")` ≠
`Resolve("sin-home~vitalia~vitalia")`.

### 9.2 Sello de build y self-update

El identificador completo es **`X.Y.Z.AAMMDDHHMM`** (`docs/architecture/conventions/versionado.md:134-139`),
inyectado en `scripts/bundle.sh:41-47` (`BUILD="$(date '+%y%m%d%H%M')"` + `-ldflags -X`), **nunca en
el Makefile**. Verificado en vivo: `GET /api/version` → `"version":"0.2.24.2607261724"`,
`"compilado":"2026-07-26 17:24"`, `"huella":"7b1d636+sucio"`, `"sucio":true`.

⚠️ **GOTCHA de incidente real (2026-07-25)** — `Makefile:27-32`:

> «`make installer` arma el .deb/.rpm/.AppImage con un sidecar fresco, pero el shell Tauri instalado
> (`web/src-tauri/src/lib.rs`, `override_local`) **PREFIERE SIEMPRE `~/.local/bin/arnesia`** por sobre
> ese sidecar si ese archivo existe […] instalar un .deb nuevo **NO cambia nada visible** hasta que
> también corras `make dev-sync`»

Reforzado por el check as-code `dev-daemon-sincronizado` (`versionado.md:154`), cuya señal en el mapa
es «instalé un .deb nuevo y la UI sigue vieja».

**En esta máquina el override EXISTE y está sincronizado**: `~/.local/bin/arnesia` y `bin/arnesia`
tienen mismo tamaño (18538761) y misma hora (jul 26 17:24).

🔴 **Consecuencia directa:** este paquete cambia el formato de `~/.arnesia/sessions.json`. **Verificar
E2E sin `make dev-sync` = medir el daemon viejo con el formato viejo** y concluir que el código nuevo
no anda. Y `dev-sync` sale con `exit 0` silencioso si no hay override (`Makefile:87-91`).

**Sub-gotcha ya resuelto (no repetir):** `pkill -f` matcheaba su propia línea y se auto-mataba; el fix
es el corchete `[a]rnesia` (`Makefile:96-101`).

**`pathAumentado`** — `internal/adapters/selfupdate/updater.go:179-181`: las apps `.desktop` no
heredan `~/.profile`, así que `exec.LookPath` no ve binarios de `~/.local/bin`. El mismo root cause
está replicado en `internal/adapters/stt/local/local.go:129`. **Si este paquete agrega cualquier
subproceso o resolución de binario, usar el mismo patrón.**

Deuda abierta: `BACKLOG.md:86-88` — «El sello a través del botón «Actualizar», de punta a punta.
`state: deuda`».

### 9.3 El SPA con `:4200` hardcodeado

`web/src/shared/api/client.ts:21`:
```ts
const BASE = import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"
```

Superficies acopladas al mismo número: `web/src-tauri/src/lib.rs:30`
(`const DAEMON_ADDR: &str = "127.0.0.1:4200"`), `web/src-tauri/capabilities/default.json:7`, y el
comentario de reintento en `sessions-store.ts:107` (40 reintentos × 500 ms).

**El gotcha, documentado** — `BACKLOG.md:77-82`:

> «`state: bug`. Destapado verificando RF-231 en vivo: con el daemon en otro puerto, **la UI entera
> queda en `Failed to fetch`** aunque el daemon responda perfecto por `curl`. **Impide correr una
> segunda instancia o mover el puerto sin recompilar el FE.**»

🔴 **Impacto directo:** CV-D6 exige ejecutar el borrado «con el daemon detenido y con copia previa».
Verificar el paquete con un daemon aislado (para no tocar el `~/.arnesia` real) requiere `:4200`, o
levantar el vite dev server con `VITE_ARNESIA_API`. No hay tercera vía sin recompilar el FE.

### 9.4 El riesgo de wire de HS-26 — repetible

`docs/product/ledger/HS-26.md`, textual:

> «el E2E vivo con claude REAL cazó una regresión de wire — `contentText` se comparte entre parsear
> frames y MANDAR el turno user; **los campos nuevos sin `omitempty` rompieron la API (400 «Extra
> inputs are not permitted») y envenenaron un mensaje del historial CC de vitalia**. Fix + test de
> regresión (`TestUserTurnWireSinCamposExtra`) + JSONL curado a mano.»

🔴 **Este paquete agrega campos al dominio de sesión/conversación.** Si alguno se cuela al wire del
turno user sin `omitempty`, se repite el incidente exacto — **incluida la corrupción de la JSONL
nativa del operador**. El test de regresión existe: `conductor_test.go:54`.

### 9.5 Las 4 stories rojas — declararlas, no cobrárselas

`BACKLOG.md:61-66` y `:207-211` documentan el token: «`--warn` (`#c96a2e`) sobre `--card` (`#ffffff`)
a 10px da **3.76:1**, bajo el mínimo 4.5 de axe. **Rompe 4 stories de
`session-rail/new-session-picker`**».

Son **exactamente** las stories del bloque `ConversacionesDelArnes` que CV-D2 saca del rail. Si el
paquete elimina ese bloque, **esas 4 stories rojas desaparecen solas** — hay que **declararlo como
efecto colateral, no presentarlo como que se arregló el token**. El problema de contraste sigue vivo
en cualquier otro `.text-warn` a 10px, incluido el segundo (`new-session-picker.tsx:290`) y
cualquiera que el dock nuevo pinte para errores o `faltantes`.

### 9.6 Gotchas de lefthook — con evidencia E2E

`docs/product/ledger/HS-21.md:48-53`, textual:

> «**`pre-push` DESCARTADO** (lefthook 1.13.6 lo saltea con «no matching push files» en push real […]
> solo corre con `--force` que git no pasa) · **`stage_fixed` DESCARTADO** (no stagea el checkpoint
> cuando el trigger es otro archivo) · **`pre-commit` + `git add` ELEGIDO** y verificado con commit
> REAL»

Espejado como comentario vivo en `lefthook.yml:46-50`.

**Trampa colateral de HS-28:** verificar contra un **worktree limpio** destapó que `main` no
compilaba — cuatro paquetes que `main.go` importa nunca se habían commiteado, y `.gitignore` se comía
un fixture de tests. **Verificar contra worktree limpio y revisar `.gitignore` antes de agregar
fixtures.**

### 9.7 🔴 El BACKLOG describe un índice que ya no existe

`BACKLOG.md:286-292` y `:325-328` dicen (fechado «re-verificado 2026-07-24, SIGUEN bloqueados»):

> «esperan SQLite fase 5 real (hoy: `internal/adapters/index/store.go` sigue `// TODO(fase 5)`,
> CAP-21/22/23 siguen vivo·in-memory/parcial/stub)»

**Eso es stale.** El código está construido y commiteado: `c2c61ba feat(indice): SQLite real +
watcher fsnotify — fase 5 del stack (CAP-21/22/23)` (2026-07-24). `index/store.go:1-6` dice
«implements `ports.IndexPort` over **modernc.org/sqlite** (pure Go, no CGO)». `~/.arnesia/index.db`
existe (53 KB + WAL de 90 KB). CAP-21 `vivo`, CAP-22 `vivo`, CAP-23 `parcial`. Y
`grep "TODO(fase 5)"` ya no devuelve nada en `index/`.

🔴 **Por qué importa:** CV-D8 justifica el scan en memoria con «sin FTS5, sin índice nuevo, **sin
tocar `~/.arnesia/index.db`**». Esa premisa describe el índice viejo. Hoy `index.db` **es** SQLite
real con WAL, writer serializado y watcher fsnotify. **La decisión de no usarlo sigue siendo
defendible por volumen (§2.5), pero el razonamiento citado se apoya en una foto que ya no existe y
hay que re-fundamentarlo en la spec.**

Nota adicional: `index/store.go` documenta explícitamente que **NO es el corpus de conversaciones**
(«NOT the ~/.claude conversation JSONL, a separate corpus read by internal/adapters/history»), y
`schemaVersion = 1` con la política de que un bump **borra el `.db` entero** (RF-208).

### 9.8 Estado del proyecto — `checkpoint.md`

**Fase 5 (Implementación) EN CURSO** (`checkpoint.md:7-11`); última ficha cerrada HS-27 (stale: HS-28
ya existe en `docs/product/ledger/HS-28.md`).

**5 paquetes activos simultáneos** — el WIP está alto: dictado por voz, identidad de build,
versionado+changelog, portafolio·agregación, telemetría/capa Mejora.

⚠️ **Este paquete NO figura en `checkpoint.md`** (grep sin resultados), pese a que su `INDEX.md:19-24`
declara Gate 1 🧑‍⚖️ FIRMADO 2026-07-26. Deuda de sincronización.

**Cifras vivas** (bloque generado `checkpoint.md:185-191`, `scripts/estado.sh`):
```
ruleset --todo:  311 checks · pass 81 · fail 0 · error 0 · deferred 230 · n/a 0
dogfood --arnes: 21 checks · pass 20 · fail 1  (warn honesto art-es-path)
arch/: 26 boundaries · knowledge/: 12 nodos · 138 checks
capabilities: 139 — 84 vivo · 40 vivo·nc · 13 parcial · 2 stub · cobertura 100%
```

⚠️ La línea «**cobertura 100%** (0 huérfanos, 0 punteros colgantes)» es **texto hardcodeado** en
`scripts/estado.sh:57` — **no se mide**. La medición real la hace el go test. Es una afirmación, no
una cifra generada.

`estado.sh --check` corrido ahora: **`cifras en sync con el estado real ✓`** (exit 0).

### 9.9 Otra deuda del BACKLOG que toca el paquete

| Cita | Ítem | Por qué toca |
|---|---|---|
| `BACKLOG.md:11-24` | ítem padre «atar chat-cc-funcional + historial + rotación en UNA experiencia continua», con forks **B** («`Close()` borra el historial en vez de archivarlo») y **C** | **STALE**: B y C ya se resolvieron (HS-25), pero el ítem sigue abierto describiendo el mundo pre-B2. Su nota `:22-24` («el re-key… **sin relación directa**») quedó **contradicha**: la relación resultó directa |
| `BACKLOG.md:354-355` | «Persistencia de la lista de sesiones: índice SQLite desechable vs sidecar propio» | **Decisión no tomada** que el paquete fuerza: al persistir `Conv` elige de facto el sidecar JSON. Registrarlo como decisión explícita, no colateral |
| `BACKLOG.md:240-249` | quedan 4 sub-ítems de la fase presentación, incl. **cola de turnos (reemplazar el 409)** | Interactúa con CV-D7/D11: hoy un 2º turno da 409 (`un-turno-a-la-vez`) |
| `BACKLOG.md:353` | «⌘K quick-switch de sesión (palette P4)» | Diseño futuro sobre el rail que este paquete re-escopa |
| `BACKLOG.md:344` | «Búsqueda global (componentes, corridas, hallazgos)» | CV-D8 crea un buscador acotado; riesgo de colisión conceptual |
| `BACKLOG.md:53-60` | T7 STT sin motor | El dock que CV-D14 recorta **ya tiene el botón de dictado** (CAP-115); recortar cromo puede pisar esa superficie |
| `BACKLOG.md:158-162` | política de entradas corruptas del store, sin definir | Con `Conv` persistido, un archivo roto se lleva **todos** los transcripts (§2.1) |
| `BACKLOG.md:302-309` | `ctx-derivado-etiquetado` probablemente stale desde `b91a061` | CV-D14 asciende el ctx a protagonista (§6.2) |

---

## 10. Huecos que la arquitectura tendrá que resolver

Lista numerada de lo que **no existe hoy** y el paquete necesita. **Sin propuesta de solución** — eso
es del arquitecto.

1. **Timestamp de última interacción por conversación.** `domain.Turn` es exactamente
   `{Rol, Text}` (`session.go:62-65`) — sin tiempo. Lo único temporal del agregado es `CerradaEn`
   (`:123`), que es la fecha de cierre. CV-D13 exige «última interacción» por fila, y sin un
   timestamp estampado **en cada turno** (no al desactivar) una conversación inactiva mentiría la
   fecha de su último mensaje.

2. **Persistencia de `Conv` al desactivar.** Hoy se anula explícitamente en
   `session_historial.go:57`, con un test que lo cementa (`session_historial_test.go:73`) y una
   capability que lo afirma como ley (CAP-98, `:33-34`). CV-D8 y CV-D11 dependen de invertirlo.

3. **Migración del esquema en disco.** `store.Registry` no tiene `schema_version`, ni envelope, ni
   detección de formato, ni ruta de upgrade (`registry.go:45-61`). Partir `Session` cambia la forma
   del array. Hoy un `Unmarshal` que falla devuelve error y el daemon pierde el registro entero.
   **CV-D6 resuelve el archivo de cerradas por borrado, pero no dice nada de `sessions.json`**, que
   tiene 5 sesiones vivas con 94 turnos acumulados.

4. **Endpoint(s) de conversación.** No existe ninguna ruta con `conversation` en el path
   (`router.go:122-138`). Los dos actuales cuelgan de un **string de arnés**
   (`/api/sessions?arnes=…&cerradas=1` y `/api/sessions/cerradas/{id}/historial`), que es lo que CV-D5
   elimina. Falta: listar conversaciones de una sesión, crear, retomar (CV-D11), renombrar (CV-D9),
   y traer el transcript de una inactiva.

5. **Búsqueda en texto.** No hay ninguna superficie de búsqueda sobre transcripts — ni endpoint, ni
   usecase, ni índice. El FE tampoco tiene buscador en el dock. CV-D8 lo pide sobre `Conv`.

6. **Reparto del estado in-flight entre Sesión y Conversación.** `sessionRuntime`
   (`session_service.go:85-109`) tiene 12 campos hoy por-sesión: `live`, `runSeq`, `curRun`,
   `assembling`, `pendingTurn`, `wasResume`, `sawInit`, `resumeRetried`, `pendingPerm`, `grants`,
   `msgFlushed`, `cwd`. Su espejo FE son 6 mapas por `session.id` (`sessions-store.ts:34-47`). **Toca
   el boundary `sesion-viva-consistente` (`enforced`, high) y sus 4 tests.** No cubierto por
   CV-D1..D15 (§5.5).

7. **El join hacia el corpus JSONL cuando el `Cwd` es de la sesión.** `history.Reader.Turnos` exige
   `(cwd, claudeSessionID)` (`reader.go:63-64`). Si el `Cwd` queda en la Sesión (**S**) y la Sesión
   puede cambiar de cwd, el historial de las conversaciones viejas se rompe en silencio. Hoy el
   backfill de `archivarLocked` (`session_historial.go:51-54`) además depende de
   `resolver.Resolve(Arnes)`, que arrastra el problema de llaves (§9.1).

8. **Identidad del system-prompt por sesión con N conversaciones.**
   `injector.ProvisionSession(ctx, id, tarjeta)` (`session_service.go:410`) usa el **`id` de sesión**
   como key; escribe en `~/.arnesia/sessions/<id>/system.md` (CAP-96). El `Checkpoint` de rotación
   viaja por ese mismo archivo (`:406-408`). Con N conversaciones por sesión, dos hilos se pisarían
   el system-prompt.

9. **Qué pasa con los permisos pendientes de la conversación que se desactiva.** El `INDEX.md:66` del
   paquete lo deja abierto explícitamente. Hoy `pendingPerm` (`session_service.go:100`) y
   `pendingPerms[id]` (`sessions-store.ts:40`) cuelgan de la sesión, y el conductor queda **bloqueado
   en el canal de control** esperando la decisión. Desactivar sin resolver deja un proceso colgado.

10. **Resignificación (o eliminación) de `sesiones-cerradas.json` + `SetArchivoCerradas`.** CV-D12
    dice que el registro «deja de tener sentido» y que se decide en spec si desaparece o queda como
    registro de sesiones enteras cerradas. Afecta `main.go:325-328`, `cerradasPathDefault`
    (`main.go:795-805`), `SetArchivoCerradas` (`session_historial.go:23`), `Cerradas` (`:71`),
    `HistorialCerrada` (`:97`) y los 2 endpoints.

11. **Sincronización del OpenAPI.** El spec ya driftó con los 2 endpoints B2 (§3.3) y **nada lo caza**
    (el paso de CI es condicional e inerte). Decidir si el contrato nuevo se declara o se acepta el
    drift explícitamente.

12. **`Frente` con dos semánticas fusionadas.** Un solo campo (`session.go:74`), un solo productor
    (`session_service.go:352-354`), dos lectores con significados distintos (rail = nombre del frente;
    dock = título del hilo). CV-D9 los separa; hay que decidir qué hereda el campo existente y qué
    pasa con los `Frente` ya derivados en disco.

13. **Proyección del `Status` de la conversación activa hacia la sesión.** El pip del rail
    (`session-rail.tsx`) y `selectAttention` (`sessions-store.ts:467`) leen `Session.Status`. Bajo
    CV-D3 el estado del conductor es de la conversación; la sesión necesita proyectarlo.

14. **Cobertura de test inexistente en las 3 superficies que se van a reescribir:**
    `internal/adapters/store` (0 tests), `session_service.go` (sin test propio),
    `transport/http/sessions.go` (1 test para 10 endpoints). Y `chat-dock.tsx` sin stories (§8.4).

15. **Contrato de `Conv == nil` vs `Conv == []`.** El boundary `no-aplica-no-es-cero` (`enforced`,
    **error**) exige que «no pude leer» viaje `null` y «leí y está vacío» viaje `[]`. Con `Conv`
    persistido y conversaciones que se cargan perezosamente, esa distinción pasa a ser observable en
    el wire y hay que modelarla (¿puntero? ¿campo `cargado`?).

---

## Apéndice — comandos corridos para este relevamiento

```
git rev-parse --short HEAD                     → dd460f3
make version                                   → 0.2.24
curl -s http://127.0.0.1:4200/api/version      → 0.2.24.2607261724 · huella 7b1d636+sucio
go test ./...                                  → 26 ok · 8 no-test-files · 0 FAIL
go run ./cmd/arnesia conformance --todo        → 311 checks · pass 81 · fail 0 · deferred 230
go run ./cmd/arnesia conformance --arnes …     → 21 checks · pass 20 · fail 1
bash scripts/estado.sh --check                 → cifras en sync ✓ (exit 0)
pnpm exec vitest --project=unit run            → 8 files / 122 tests PASSED
pnpm exec vitest --project=storybook run       → 1 file FAILED / 4 tests FAILED (new-session-picker)
grep -rh "^func Test" … | wc -l                → 622 en 82 archivos
ls docs/architecture/boundaries/*.md | wc -l   → 26
ls docs/product/capabilities/*/*.yaml | wc -l  → 139
grep -c "cerradas" …/openapi.yaml              → 0
git show --stat cca314a                        → 21 archivos, +306/−200
git rev-list --count origin/main..HEAD         → 79 commits sin pushear
curl GET /api/sessions?arnes=…&cerradas=1      → ver §3.2
python3 (lectura de ~/.arnesia/*.json)         → ver §2.3, §2.5, §9.1
```

**NO VERIFICADO** (declarado, no medido):
- `go test ./... -race` (CI lo corre; acá se corrió sin `-race`; `docs/architecture/fitness` ya tarda 36 s sin él).
- `cargo clippy` / `cargo fmt --check` del job `rust`.
- Los 3 scripts E2E de `docs/product/stories/2026-07-08-chat-cc-funcional/e2e/`.
- Si el `.deb` de `instaladores/v0.2.24/` levanta (no se instaló nada).
- Si el gate 🧑‍⚖️ del paquete fase-5 SQLite fue firmado fuera del repo (su `PARIDAD.md` lo declara ABIERTO y no hay ficha de ledger).
- El conteo check-por-check de cada boundary (los números del §6 son los **declarados** en los frontmatter/INDEX, que ya se probó que driftean).
