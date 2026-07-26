# Arquitectura · Las conversaciones viven en el panel de conversación

> `tipo: arquitectura` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Etapa 3 (METODOLOGIA §10), par de [`spec.md`](./spec.md) (RF-300…RF-357 · E-01…E-50),
> [`design.md`](./design.md) (C-1…C-13) y [`plan-storybook.md`](./plan-storybook.md).
> Línea base: [`relevamiento-as-is.md`](./relevamiento-as-is.md). Ley: `decisiones.md` CV-D1…CV-D15
> **firmadas** + las enmiendas F-1…F-4. **P-1 NO está firmada** y este documento la trata como tal.
>
> **Contrato de honestidad.** Toda afirmación sobre el código lleva `archivo:línea` verificada hoy
> contra `HEAD = dd460f3` y contra el daemon vivo `0.2.24.2607261724` en `127.0.0.1:4200`. Las
> cifras se generaron; ninguna se tecleó. Lo que no tiene respuesta se declara en §9, no se rellena.
>
> **Qué NO hace este documento.** No relitiga CV-D1..D15. No escribe código. No firma gates.

## 0 · Estado y las seis decisiones de fondo

| # | Problema | Decisión de fondo | §|
|---|---|---|---|
| **A** | El estado in-flight es por-SESIÓN (F-4) | **Nada de `sessionRuntime` baja a la conversación.** CV-D7 («una activa») hace que el runtime por sesión YA sea el runtime de la conversación activa. Gana un puntero `convActiva` y una transición atómica. `runSeq` **se queda por sesión y monótono** a propósito | §1.4 · §3 · §4 |
| **B** | No existe migración de esquema en disco | **Envelope versionado + cadena de migradores forward-only + respaldo con sello + cuarentena del corrupto + solo-lectura ante esquema futuro.** El precedente del índice (wipe-and-rebuild) es **ilegal** acá y el porqué queda as-code | §2 |
| **C** | La rotación no emite ningún frame (C-6) | `rotarLocked` **no publica**: acumula y **devuelve** el frame, que `Turn` emite tras soltar el lock — que es la disciplina que el servicio ya tiene. Frame nuevo `conversacion`, idempotente **por índice de turno**, no por `run_id` | §4.3 · §5.2 |
| **D** | `Conv` y `Checkpoint` se destruyen al archivar (F-3) y CAP-98 lo afirma | Las 3 líneas se invierten; el test **se renombra e invierte** (no se borra) y CAP-98 cambia de `business_rule` con entrada en su `change_log` + ficha de ledger. `Turnos` deja de persistirse: **se deriva** | §1.2 · §7.3 |
| **E** | 4 de 5 sesiones vivas con id pelado (F-1) | **CV-D16 🧑‍⚖️ FIRMADA:** el re-key es un **paso de la migración**, no un extra. Nada se borra. Respaldo obligatorio `sessions.json.bak-<sello>` + **dos caminos de reversión especificados**. Lo que no se puede decidir **no se decide y se dice** | §2.6 |
| **F** | No hay línea base verde | Tramo 0: `lefthook install` → story del dock vigente → **aplicar C-3 a los 2 `text-warn` del picker** (2 líneas, la regla que el paquete ya decidió) → declarar el OpenAPI **como es hoy** → recién ahí el enforcer de drift | §7.4 · `plan-desarrollo.md` T1-T6 |

**Lo que se escribió en `docs/architecture/`** (§7.5): 2 boundaries nuevos + 1 re-enunciado. Nada más
— el resto es producto y vive en capabilities.

---

## 1 · Modelo de datos

### 1.1 La relación

```
Session  (el frente de trabajo — lo que vive en el rail)
  │  ID · Frente · Arnes · Empresa · Puesto · Salud · Status* · View · Parked
  │  Reparacion · Cwd · CerradaEn
  │
  └── Conversaciones []Conversacion        1 — N,  N ≥ 1  y  exactamente 1 con Activa==true
        │  ID · Titulo · TituloEditado · Activa · CreadaEn · UltimaInteraccion
        │  ClaudeSessionID · Model · CtxPct · CtxHist · RotacionPendiente
        │  CadenaCC · Checkpoint · Conv []Turn
        │
        └── Conv []Turn      (domain.Turn NO cambia: {Rol, Text})
              └── CadenaCC ──► N JSONL nativas en ~/.claude/projects/<dir(Cwd)>/<ccid>.jsonl
                               (join: Cwd de la SESIÓN + ccid de la CONVERSACIÓN)

* Status es PROYECCIÓN: es el estado del conductor de la conversación activa (§1.5).
```

### 1.2 `domain.Conversacion` — archivo nuevo `internal/domain/conversacion.go`

Campo por campo. «Viene de» cita la línea de `internal/domain/session.go` de la que baja.

| Campo Go | Tipo | JSON tag | Viene de | Dueño / por qué |
|---|---|---|---|---|
| `ID` | `string` | `id` | — **nuevo** | La conversación. `cv<8hex>`, mismo generador que `newID()` (`session_service.go:929`) con otro prefijo, para que un id de conversación nunca se confunda con uno de sesión al leer un log |
| `Titulo` | `string` | `titulo` | `Frente` (`:75`) **por derivación, no por herencia** | La conversación (CV-D9). La **Sesión conserva su `Frente`**: son dos nombres y ambos siguen existiendo (RF-303 CA-3) |
| `TituloEditado` | `bool` | `titulo_editado,omitempty` | — **nuevo** | La conversación. Con `true`, ningún turno re-deriva (RF-303 CA-1) |
| `Activa` | `bool` | `activa` | — **nuevo** | La conversación. **Sin `omitempty`**: `false` es un dato, no una ausencia (BR-CV-9, `no-aplica-no-es-cero`) |
| `CreadaEn` | `string` | `creada_en` | — **nuevo** | La conversación. RFC3339 UTC. Es el **desempate de orden** para las de 0 turnos (RF-320 CA-2), que no tienen fecha de interacción y a las que BR-CV-14 prohíbe inventarles una |
| `UltimaInteraccion` | `string` | `ultima_interaccion,omitempty` | — **nuevo** (RF-304) | La conversación. RFC3339 UTC. Vacío **⟺** `len(Conv)==0`. `omitempty` es correcto acá y sólo acá: el vacío ya significa «no hubo turno», y la UI lo pinta con literal propio (`mockup:667`) |
| `ClaudeSessionID` | `string` | `claude_session_id,omitempty` | `:96` | La conversación. Es lo que va a `--resume` (`session_service.go:390`) |
| `Model` | `string` | `model,omitempty` | `:97` | **La conversación.** Se captura del `init` de CADA proceso (§1.2 del relevamiento, ambigüedad 3). El `SpawnOpts.Model` que se le pasa al spawn sale de la misma conversación: un solo campo, un solo dueño |
| `CtxPct` | `int` | `ctx_pct,omitempty` | `:99` | La conversación |
| `CtxHist` | `[]int` | `ctx_hist,omitempty` | `:103` | La conversación. Sigue acotado por `maxCtxHist = 500` (`session_service.go:137`) |
| `RotacionPendiente` | `bool` | `rotacion_pendiente,omitempty` | `:107` | La conversación. **Es también el insumo del `caliente` del `CtxChip`** (design.md §3.3): el FE no conoce el umbral y no lo va a teclear |
| `CadenaCC` | `[]string` | `cadena_cc,omitempty` | `:111` | La conversación (CV-D10). Cose los N `ClaudeSessionID` de sus rotaciones |
| `Checkpoint` | `string` | `checkpoint,omitempty` | `:115` | La conversación. **Deja de destruirse al archivar** (F-3). Viaja al spawn por `session_service.go:406-408`. **No viaja al FE** (§6.1) |
| `Conv` | `[]Turn` | `conv` | `:127` | La conversación. **Sin `omitempty`, deliberado**: con `omitempty`, `nil` («no cargado») y `[]` («leí y está vacío») se serializan igual — exactamente lo que `no-aplica-no-es-cero` (`enforced`/**error**) prohíbe. La distinción se resuelve **por tipo**, no por presencia: el DTO de la lista no tiene el campo (§5.1) |

**Lo que NO se agrega, y es una decisión:**

- **`domain.Turn` no cambia.** Sigue siendo `{Rol, Text}` (`session.go:62-65`). CV-D13 pide «última
  interacción» y RF-304 CA-1 la ubica en la **conversación**, estampada en el mismo punto donde se
  appendea el turno (`session_service.go:355` y la rama `result` de `consume`, `:497-518`). Un
  timestamp por `Turn` costaría 90 campos nuevos en la conversación más larga para responder una
  pregunta que se responde con uno. Y hay un motivo más fuerte: **`Turn` es el tipo que roza el wire
  del turno user**, y agregarle un campo es literalmente el incidente HS-26 («los campos nuevos sin
  `omitempty` rompieron la API con 400 y envenenaron un mensaje del historial CC de vitalia»,
  `ledger/HS-26.md`, test de regresión `conductor_test.go:TestUserTurnWireSinCamposExtra:54`). No se
  toca.
- **`Turnos` deja de existir como campo persistido.** Hoy es `Turnos int` (`session.go:124`),
  escrito una sola vez en `archivarLocked` (`session_historial.go:56`) porque `Conv` se tiraba. Con
  `Conv` persistido siempre, un campo que duplica `len(Conv)` es un drift esperando ocurrir. Se
  **deriva**: `func (c Conversacion) NumTurnos() int { return len(c.Conv) }`, y el DTO del wire lo
  manda calculado. Cumple RF-305 CA-2 literalmente («`Turnos` sigue existiendo y es `len(Conv)`») y
  además lo hace imposible de mentir.

### 1.3 `domain.Session` — lo que queda, lo que gana, lo que pierde

**Se conserva el nombre `Session`.** `naming.md` no exige el rename y el árbol ya convive con
`Session`/`SessionStatus`/`SessionService` en inglés junto a `Salud`/`Rol`/`Reparacion` en español.
Renombrar a `Sesion` rompería el `#Símbolo` de **9 capabilities** (§7.3) sin comprar nada, en un
paquete cuyo presupuesto de riesgo ya se gasta en partir el agregado. `Conversacion` sí nace en
español, consistente con `Reparacion` (`session.go:92`).

| Campo | Qué pasa | Por qué |
|---|---|---|
| `ID` `:71` · `Frente` `:75` · `Arnes` `:78` · `Empresa` `:79` · `Puesto` `:80` · `Salud` `:81` · `View` `:86` · `Parked` `:87` · `Reparacion` `:92` | **quedan, sin cambios** | Son del frente de trabajo |
| `Cwd` `:119` | **queda en la Sesión** | Lo resuelve `spawnLocked` desde `s.resolver.Resolve(r.meta.Arnes)` (`session_service.go:386`), que es **por arnés**: dos conversaciones de una sesión corren en el mismo cwd por construcción. Ver §9-H-A por qué esto es un riesgo declarado y no un descuido |
| `Status` `:85` | **queda, pasa a ser proyección** | §1.5 |
| `CerradaEn` `:123` | **queda en la Sesión** | Pasa a significar «esta SESIÓN se archivó» (RF-306). Ya no describe una conversación |
| `Turnos` `:124` | **se elimina** | §1.2 |
| `ClaudeSessionID` `:96` · `Model` `:97` · `CtxPct` `:99` · `CtxHist` `:103` · `RotacionPendiente` `:107` · `CadenaCC` `:111` · `Checkpoint` `:115` · `Conv` `:127` | **bajan a `Conversacion`** | CV-D3 / RF-300 |
| `Conversaciones []Conversacion` | **nace**, tag `conversaciones` | El agregado. Sin `omitempty`: una sesión con `[]` es una sesión rota, y tiene que verse |

### 1.4 Dónde vive la invariante — y dónde NO

**La invariante principal: `len(Conversaciones) ≥ 1` ∧ exactamente una con `Activa==true`.**

Vive en **`internal/domain`**, en operaciones puras sobre el agregado. No en el handler, no en el
store, **no en el FE**. Las tres transiciones son métodos del dominio y son testeables sin un solo
fake:

```go
// internal/domain/conversacion.go

// ErrConvNoEncontrada / ErrTituloVacio / ErrSinActiva son los sentinelas del agregado.
var (
    ErrConvNoEncontrada = errors.New("domain: la conversación no pertenece a esta sesión")
    ErrTituloVacio      = errors.New("domain: el título no puede quedar vacío")
    ErrSinActiva        = errors.New("domain: la sesión quedó sin conversación activa")
)

// Activa devuelve la conversación activa. El bool es el guardrail de un registro
// editado a mano; bajo la invariante siempre es true.
func (s *Session) Activa() (*Conversacion, bool)

// CrearConversacion desactiva la activa y agrega una nueva activa. Devuelve el id de
// la que quedó inactiva ("" si la sesión no tenía ninguna). Es UNA transición (BR-CV-4).
func (s *Session) CrearConversacion(id string, ahora time.Time) (nueva *Conversacion, desactivada string)

// ActivarConversacion mueve la marca de activa a cid. No-op si cid ya es la activa
// (E-15). ErrConvNoEncontrada si cid no está en ESTA sesión (BR-CV-2 — jamás se busca
// globalmente).
func (s *Session) ActivarConversacion(cid string) (desactivada string, err error)

// RenombrarConversacion aplica el título recortado y marca TituloEditado.
// ErrTituloVacio si queda vacío; el título NO cambia (RF-344 CA-1).
func (s *Session) RenombrarConversacion(cid, titulo string) error

// VerificarUnaActiva es el predicado puro de la invariante. Lo corre el test de
// dominio después de CADA transición, y el store después de cargar.
func VerificarUnaActiva(s Session) error

// NormalizarConversaciones REPARA la invariante y DEVUELVE qué reparó, para que el
// arranque lo diga en el log. Nunca repara en silencio (E-01).
//   - 0 conversaciones            ⇒ crea una activa vacía
//   - ninguna activa              ⇒ activa la de UltimaInteraccion más reciente
//                                   (desempate: CreadaEn más reciente)
//   - ≥2 activas                  ⇒ conserva esa misma y desactiva el resto
func (s *Session) NormalizarConversaciones(ahora time.Time) []string
```

**Los tres puntos donde se enforza, en orden de defensa:**

1. **Dominio (la única fuente):** las 4 operaciones de arriba no pueden expresar un estado
   inválido. Test: `internal/domain/conversacion_test.go:TestInvarianteUnaActivaTrasCadaTransicion`
   (table-driven sobre las 4, verificando `VerificarUnaActiva` después de cada una).
2. **Carga desde disco:** `NormalizarConversaciones` corre en `Registry.Load`, después de migrar y
   antes de devolver. Un registro editado a mano, una migración parcial o un archivo de otra versión
   no pueden meter una sesión inválida en memoria. Lo que reparó viaja en `store.Informe.Reparaciones`
   (§2.3) y `main.go` lo loguea con `slog.Warn`, una línea por reparación.
3. **Persistencia:** `persistLocked` corre `VerificarUnaActiva` sobre el snapshot y **se niega a
   guardar** un registro que la viole, logueando el id de la sesión culpable. Es un cinturón: si
   alguna vez llega, el archivo en disco no se contamina.

**El FE no la enforza y no la asume «por si acaso»:** `Session.activa` es un campo **no opcional**
del wire (§6.1), así que el tipo hace imposible el `if (!activa)` defensivo que terminaría
escondiendo el bug. Si el daemon mandara una sesión sin activa, el FE rompe ruidoso — que es lo
correcto — en vez de pintar un dock vacío.

### 1.5 `Status`: la proyección, y por qué no se duplica

El pip del rail y `selectAttention` (`sessions-store.ts:467`) leen `Session.Status`. Bajo CV-D3 el
estado del conductor es de la conversación. **No se agrega `Conversacion.Status`.** Motivo: sólo la
activa tiene conductor (BR-CV-1), así que un `Status` por conversación tendría un valor real y N−1
valores inventados — y `no-aplica-no-es-cero` prohíbe exactamente eso. `Session.Status` **es** el
estado del conductor de la activa, se escribe donde ya se escribe (`Turn` `:356`, `consume` `:452+`,
`Close`, spawn fallido `:366`) y no cambia una línea. El rail y `selectAttention` siguen andando sin
tocarse.

---

## 2 · Persistencia

### 2.1 Layout de archivos

| Ruta | Qué es | Quién la escribe | Estado tras el paquete |
|---|---|---|---|
| `~/.arnesia/sesiones.json` | **NUEVO** · envelope v2 · sesiones vivas **con** sus conversaciones | `store.Registry` | el registro vivo |
| `~/.arnesia/sessions.json` | v1 desnudo (17 202 B, 5 sesiones) | nadie, después de migrar | **intacto**, respaldo y ruta de downgrade (RF-335 CA-2) |
| `~/.arnesia/sessions.json.bak-<AAMMDDHHMM>` | **NUEVO** · copia byte a byte del v1 **antes** de tocar nada | `AbrirRegistro`, una vez | El respaldo que CV-D16 exige, y el insumo de la reversión selectiva del re-key (§2.6) |
| `~/.arnesia/sesiones-archivadas.json` | **NUEVO** · envelope v2 · sesiones **enteras** cerradas, con sus conversaciones completas | `archivarLocked` | reemplaza el registro de cerradas (RF-306) |
| `~/.arnesia/sesiones-cerradas.json` | v1 · 3 entradas (1 479 B) | nadie | **lo borra el operador a mano** (RF-337/CV-D6). La migración **no lo lee**: el dato se elimina, no se convierte |
| `~/.arnesia/sessions.json.bak` | 522 B, 2026-07-05, preexistente | nadie | no se toca |
| `~/.arnesia/sessions/<id>/system.md` | tarjeta de identidad por sesión (CAP-96) | `injector.ProvisionSession` | **no cambia** — ver §4.5 |

**Por qué archivo nuevo y no migración in-place, esta vez y sólo esta vez.** El envelope se estrena
acá; mientras no exista, no hay forma de que un binario viejo sepa que el archivo cambió de forma.
Escribir `sesiones.json` y dejar `sessions.json` quieto compra un **downgrade gratis**: el daemon
`0.2.24` sigue leyendo lo suyo y el operador puede volver atrás sin restaurar nada. De v2 en
adelante el envelope hace innecesario el truco y la migración es in-place con `.bak-<sello>`, que es
lo que el boundary describe como la regla. El rename es el **bootstrap**, no la política.

### 2.2 El envelope

```jsonc
{
  "schema_version": 2,
  "escrito_por": "0.2.25.2607262100",   // sello de build (versionado.md §134-139)
  "escrito_en":  "2026-07-26T21:00:00Z",
  "sesiones": [ { "id": "s25123a2c", "...": "...", "conversaciones": [ … ] } ]
}
```

```go
// internal/adapters/store/esquema.go — archivo nuevo

// EsquemaActual es la forma que este binario escribe. Bumpearlo exige sumar un migrador.
const EsquemaActual = 2

type sobre struct {
    SchemaVersion int             `json:"schema_version"`
    EscritoPor    string          `json:"escrito_por"`
    EscritoEn     string          `json:"escrito_en"`
    Sesiones      json.RawMessage `json:"sesiones"`
}

// migrador lleva el payload de la versión N a la N+1. Falla ruidoso: un campo que no se
// supo llevar es un error, jamás un zero-value.
type migrador func(payload json.RawMessage) (json.RawMessage, error)

var migradores = map[int]migrador{1: deV1aV2}
```

**Detección, no adivinanza.** `detectarVersion(b []byte) (int, error)` mira el primer byte no-blanco:

| primer byte | significa | acción |
|---|---|---|
| `[` | array desnudo = **v1 implícito** (la forma que `registry.go:73` escribía) | migrar 1→2 |
| `{` | envelope | leer `schema_version`; `0`/ausente ⇒ error, no v1 tácito |
| otro / vacío | no es JSON | cuarentena (§2.4) |

### 2.3 La máquina de migración

```go
// AbrirRegistro resuelve qué archivo manda, migra si hace falta y deja el Registry listo.
// Es lo único que cmd/arnesia/main.go llama; NewRegistry sigue siendo el constructor tonto
// y ports.SessionStore NO cambia de firma (arnes_registry.go y los callers no se enteran).
func AbrirRegistro(rutaV2, rutaLegado, sello string, clave ClaveCalificada) (*Registry, Informe, error)

// ClaveCalificada resuelve un id de arnés pelado a su clave (home,id,scope) — CV-D16.
// ok=false cuando NO hay exactamente una candidata: entonces la llave se deja como está y
// el motivo viaja al Informe. Jamás adivina (no-aplica-no-es-cero).
// Se INYECTA: `store` sólo puede depender de [domain, ports] (.go-arch-lint.yml:171-172),
// así que no importa el Portafolio — lo cablea cmd/arnesia/main.go, que sí puede (:212).
type ClaveCalificada func(idPelado, cwd string) (clave string, ok bool, motivo string)

// Informe es lo que el arranque LOGUEA. Nada de esto ocurre en silencio.
type Informe struct {
    Migro         bool     // corrió una migración en este arranque
    DesdeVersion  int      // 0 si no migró
    RespaldoEn    string   // "" si no migró
    Corrupto      bool
    CuarentenaEn  string   // ruta del .corrupto-<sello>
    EsquemaFuturo bool     // el archivo lo escribió un binario más nuevo
    Reparaciones  []string // lo que NormalizarConversaciones arregló (§1.4)
    Recalibradas  []Recalibracion // el re-key de CV-D16, una fila por sesión
}

// Recalibracion es el rastro auditable de CV-D16: qué sesión, de qué llave a cuál, y
// cuando NO se movió, por qué. Se loguea entera al arrancar y se transcribe a PARIDAD.md.
type Recalibracion struct {
    SesionID string
    Antes    string
    Despues  string // == Antes cuando no se pudo decidir
    Motivo   string // "resuelta-por-cwd" | "resuelta-por-id" | "sin-candidata" | "ambigua: N" | "ya-calificada"
}

// SoloLectura reporta si el registro está bloqueado y por qué. Toda mutación consulta
// esto ANTES de tocar memoria (§2.4, modo E).
func (r *Registry) SoloLectura() (bool, string)
```

Secuencia de `AbrirRegistro`:

```
1. ¿existe rutaV2?
   sí → leerlo                          ── es el camino de todos los arranques después del primero
   no → ¿existe rutaLegado?
        no → registro vacío, EsquemaActual, sin migración      (primer arranque)
        sí → leerlo                                            (migración v1→v2, una vez)
2. detectarVersion
   > EsquemaActual  → modo SOLO-LECTURA, Informe.EsquemaFuturo, NO se escribe nada  (modo E)
   ilegible         → cuarentena, registro vacío, Informe.Corrupto                  (modo C)
   == EsquemaActual → seguir al paso 4                                              (IDEMPOTENCIA)
   <  EsquemaActual → paso 3
3. RESPALDAR PRIMERO, antes de cualquier escritura:
     v1 → ~/.arnesia/sessions.json.bak-<sello>     (el nombre que CV-D16 fija)
     v≥2 → <ruta>.v<N>-<sello>.bak                 (la forma genérica del boundary)
   luego aplicar migradores[N], migradores[N+1], … hasta EsquemaActual
   un migrador que falla ⇒ NO se escribe nada, se devuelve el error con la versión y el paso
4. RE-KEY (CV-D16), sólo en el tramo 1→2 y sólo sobre `Arnes`:
   por cada sesión, clave(Arnes, Cwd) → si ok, Arnes = clave; si no, se DEJA como está.
   Cada sesión emite una Recalibracion, se mueva o no. Nada se borra, nada se fusiona.
5. NormalizarConversaciones(ahora) por sesión → Informe.Reparaciones
6. Save atómico a rutaV2 (temp en el mismo dir + rename, registry.go:77-95 sin cambios)
   rutaLegado NO se toca, NO se borra, NO se renombra
```

**El paso 4 está separado del 3 a propósito.** El migrador de **forma** (`deV1aV2`) es puro: no
conoce el Portafolio, no hace IO, y se testea con un fixture JSON. El re-key es una migración de
**dato**, depende de una autoridad externa y puede no poder decidir. Fusionarlos haría intestable la
parte que sí es determinista y ataría el arranque del daemon a que el Portafolio responda.

**`deV1aV2`, exacto** (RF-335): por cada `Session` v1 se emite una `Session` v2 con **una**
conversación activa que hereda `claude_session_id`, `model`, `ctx_pct`, `ctx_hist`, `cadena_cc`,
`checkpoint`, `rotacion_pendiente` y `conv`; `titulo = deriveFrente(primer turno rol=user)` o
`"nueva conversación"` si no hay ninguno; `titulo_editado = false`; `creada_en = ahora` (es lo único
honesto: la fecha real no existe); **`ultima_interaccion = ""` aunque tenga 90 turnos** — el dato no
existe y BR-CV-14 prohíbe inventarlo (H-4). La sesión conserva `id`, `frente`, `arnes`, `empresa`,
`puesto`, `salud`, `status`, `view`, `parked`, `reparacion`, `cwd`. **`Arnes` no se toca** (§2.6).

Resultado verificable contra esta máquina: **5 sesiones → 5 sesiones × 1 conversación**; `s6165ac75`
conserva sus 90 turnos y su `ctx_pct: 19`; `s25123a2c` conserva sus 4 turnos, `ctx_pct: 68` y
`rotacion_pendiente: true`; `s0fec7798`/`sfc512b15`/`s78b3aeeb` conservan 0 turnos y salen con
`ultima_interaccion` vacío y `conv: []` (no `null`).

### 2.4 Los seis modos de fallo, uno por uno

| | Situación | Qué hace | Qué NO hace |
|---|---|---|---|
| **A** | **Archivo ausente** | Registro vacío, `EsquemaActual` estampado en el primer `Save`. Es el primer arranque | No es error (comportamiento vigente, `registry.go:50-52`) |
| **B** | **Crash a mitad de la migración** | El respaldo ya existe y el original sigue entero; el `Save` es temp+rename (`registry.go:77-95`): queda **el viejo entero o el nuevo entero**. Al reintentar, el paso 1 vuelve a encontrar `rutaLegado` y rehace todo (E-47) | Nunca medio archivo |
| **C** | **JSON inválido / basura** | Renombra a `<ruta>.corrupto-<AAMMDDHHMM>`, arranca con registro vacío, `slog.Error` **con la ruta de la cuarentena**. El rail muestra 0 sesiones (E-44) | No sobreescribe · no usa `seedSessions` (`session_service.go:947-951`) para tapar: la semilla es para registro **ausente**, que es otra cosa |
| **D** | **Sin permisos de escritura** | `Save` ya envuelve el error con la ruta (`registry.go:70-95`). **Cambio real:** `persistLocked` deja de tragárselo (hoy `slog.Error` y sigue, `session_service.go:909-911`) y **devuelve** el error; crear/retomar/renombrar **revierten su cambio en memoria** y responden `500` con el motivo del filesystem (E-45, RF-338) | El estado en memoria no diverge del disco en silencio |
| **E** | **`schema_version` > `EsquemaActual`** (el operador corrió una build nueva y volvió a la vieja) | **Modo solo-lectura**: el registro carga vacío, `SoloLectura()` devuelve `true` + motivo, y **toda** mutación (`Create`, `Turn`, crear/activar/renombrar conversación) responde `503` con el motivo antes de tocar memoria | **No lee a medias · no pisa.** Degradar a vacío y después persistir destruiría el archivo nuevo con el binario viejo: es el único modo de fallo irreversible, y el que este modo existe para impedir |
| **F** | **Dos daemons sobre el mismo `$HOME`** | Nada: `Registry` tiene `sync.Mutex` in-process (`registry.go:22`) y **no** lock de archivo. Se pisan | **Hueco declarado, preexistente y NO resuelto acá** (§9-H-C). El envelope lo empeora marginalmente porque el archivo es más grande y la ventana de escritura más ancha |

### 2.5 El tamaño del archivo: presupuesto, no promesa

`persistLocked` (`session_service.go:902-912`) **serializa el registro entero** con
`json.MarshalIndent` (`registry.go:73`) y lo reescribe, y se invoca desde `Create` (`:250`),
`Rename` (`:269`), `SetView` (`:284`), `Close` (`:317`), `Turn` (`:367`, `:373`) **y desde `consume`
en cada `init`/`message`/`act`/`result`** (`:465`, `:515`, `:556`, `:573`). Es decir: **una
reescritura completa por cada paso de actividad de un turno.**

Hoy son 17 KB. Con `Conv` persistido en N conversaciones (H-7), 5 sesiones × 4 conversaciones ×
12 KB ≈ **240 KB por evento de actividad**. No está medido bajo carga real y **no se pre-optimiza**.
Lo que sí se hace es poner el presupuesto y el disparador, para que nadie lo descubra en producción:

> **Presupuesto:** un `persistLocked` con el registro proyectado a 20 conversaciones tiene que
> costar **< 15 ms** p95 en la máquina del operador. **Disparador:** si lo supera, el corte es
> **un archivo por sesión** bajo `~/.arnesia/sessions/<id>/conversaciones.json` — carpeta que **ya
> existe** (la usa la tarjeta de identidad, CAP-96, `session_service.go:410`) y que reduce la
> reescritura a la sesión que cambió. El envelope de §2.2 se aplica igual a cada archivo, así que el
> corte **no es un rediseño**: es cambiar el `path` del `Registry`.

El ticket que mide está en el plan (`plan-desarrollo.md` T14) y **no es opcional**: sin la medición,
la afirmación «alcanza» sería exactamente la clase de pass fabricado que el repo prohíbe.

### 2.6 CV-D16 · el re-key de las llaves vivas — paso de la migración, reversible

**🧑‍⚖️ FIRMADA 2026-07-26** (`decisiones.md:225-246`). Ya no es una propuesta abierta: es un
requisito, y vive **dentro** del mecanismo versionado (paso 4 de §2.3), no en un comando aparte.

**Nada se borra.** CV-D16 y CV-D6 son decisiones distintas sobre datos distintos y **no se
unifican**: `sesiones-cerradas.json` lo borra el operador a mano (RF-337); las 4 sesiones **vivas**
con id pelado se **re-key** a clave calificada `(home,id,scope)`.

| sesión | `arnes` en disco | `cwd` en disco | cómo se resuelve |
|---|---|---|---|
| `s25123a2c` | `vitalia` | `/home/…/luana-vitalia/vitalia` | **por cwd** — la vía fuerte |
| `s6165ac75` | `sin-home~vitalia~vitalia` | `/home/…/luana-vitalia/vitalia` | `ya-calificada`, no se toca |
| `s0fec7798` | `vitalia` | **vacío** | **por id** (fallback) |
| `sfc512b15` | `vitalia` | **vacío** | **por id** (fallback) |
| `s78b3aeeb` | `arnesia` | **vacío** | **por id** (fallback) |

**Hallazgo que cambia el algoritmo, verificado en vivo hoy:** 3 de las 4 a re-key tienen **`cwd`
vacío** (nunca spawnearon: 0 turnos). Desambiguar por `cwd` es imposible para ellas. El algoritmo
tiene que tener dos vías y decir cuál usó:

```
clave(idPelado, cwd):
  1. si idPelado ya es calificado            → ok=false, motivo "ya-calificada"    (no se toca)
  2. si cwd != "":
       candidatas = entradas del Portafolio cuyo id == idPelado Y que resuelven a ese cwd
       len == 1 → ok, motivo "resuelta-por-cwd"
  3. si no resolvió por cwd:
       candidatas = entradas del Portafolio cuyo id == idPelado
       len == 1 → ok, motivo "resuelta-por-id"
       len == 0 → ok=false, motivo "sin-candidata"
       len >= 2 → ok=false, motivo "ambigua: N"
```

**Lo que no se puede decidir, no se decide.** Con `sin-candidata` o `ambigua`, la llave **queda
exactamente como está** y la fila del `Informe.Recalibradas` lo dice con su motivo. Es
`no-aplica-no-es-cero` aplicado a una migración: elegir una de dos candidatas al azar rompería
`portafolio-identidad-y-deriva-honesta` («identidad `(home,id)` calificada, **nunca fusión por
coincidencia**»). **Honestidad de alcance: la arquitectura NO puede garantizar 4 de 4** — depende de
qué tenga el Portafolio del operador. Lo que garantiza es que ninguna se pierde, ninguna se fusiona,
y que el resultado real se ve **antes** de aplicarse (abajo).

**Las dos reversiones, especificadas.** «Reversible» sin procedimiento es una promesa:

| escenario | procedimiento exacto | por qué funciona |
|---|---|---|
| **R1 · volver al binario viejo** (deshacer todo: esquema + re-key) | `pkill -f 'arnesia serve'` · `rm ~/.arnesia/sesiones.json` · arrancar el daemon `0.2.24` | `sessions.json` **nunca se tocó** (§2.1): el daemon viejo lee exactamente lo que leía antes. El `.bak-<sello>` es cinturón, no se necesita |
| **R2 · deshacer SÓLO el re-key, quedándose en el esquema nuevo** | `pkill -f 'arnesia serve'` · `arnesia sesiones recalibrar-llaves --revertir --desde ~/.arnesia/sessions.json.bak-<sello>` · arrancar | El re-key cambia **un solo campo por sesión** (`Arnes`) y el `.bak` conserva el valor original **indexado por `id` de sesión**. El comando reescribe `Arnes` sobre el `sesiones.json` vigente y **no toca las conversaciones**: los turnos que se generaron después de migrar se conservan |

R2 es lo que hace que «reversible» sea verdad y no un eufemismo de «restaurá el backup y perdé lo
que hiciste desde entonces».

**La herramienta se construye igual, con tres modos** (`cmd/arnesia/sesiones.go`):

```
arnesia sesiones recalibrar-llaves [--sessions <path>]            # dry-run: imprime la tabla, no escribe
arnesia sesiones recalibrar-llaves --revertir --desde <bak>       # R2
arnesia sesiones recalibrar-llaves --aplicar                      # red de seguridad: re-key idempotente
```

El **`--dry-run` no es opcional en el procedimiento**: se corre **antes** de arrancar el binario
nuevo por primera vez, y su salida se transcribe a `PARIDAD.md`. Así el operador ve las 5 filas con
su motivo *antes* de que la migración las aplique — que es lo que convierte «nada se borra» en algo
observable y no en una afirmación. `--aplicar` existe para el caso en que el Portafolio no estuviera
disponible al arrancar (`sin-candidata` en todas): se vuelve a correr después, y es idempotente
porque el paso 1 del algoritmo saltea las ya calificadas.

**Qué arregla, honestamente.** El re-key **no** es necesario para que el panel funcione: CV-D5 y
BR-CV-2 hacen que las conversaciones cuelguen de `session_id`, así que listar y buscar andan igual
con la llave pelada. Lo que arregla son **dos consumidores del mismo bug**, ambos fuera de la
superficie nueva: el filtro `?arnes=` de `GET /api/sessions` (`sessions.go:25`, `==` desnudo, que
hoy esconde 3 sesiones vivas) y el backfill de `Cwd` en `archivarLocked`
(`session_historial.go:51-54`, que llama `resolver.Resolve(cerrada.Arnes)` y falla con la llave
pelada). CV-D16 los cierra a los dos de una.

---

## 3 · Ciclo de vida de una conversación

### 3.1 La máquina

**Dos estados, no cinco.** «Retomada» y «rotada» **no son estados** — son transiciones y eventos
sobre estados que ya existen. Modelarlos como estados sería exactamente la confusión que CV-D10
(«la rotación es invisible, misma conversación») y CV-D12 («activa/inactiva, no cerrada») firmaron
para evitar.

```
                    T1 nacer
         ∅ ──────────────────────────► ACTIVA ◄─┐
                                       │  ▲     │ T4 rotar (self-loop)
                          T2 desactivar│  │     │ el ClaudeSessionID cambia,
                                       │  │T3   │ la ENTRADA no
                                       ▼  │retomar
                                    INACTIVA
```

| T | de → a | Quién dispara | Precondición | Efecto |
|---|---|---|---|---|
| **T1** | `∅ → activa` | **operador** (`＋`, RF-307) · **sistema** (`Create` de sesión RF-301 · `NormalizarConversaciones` al cargar, E-01) | turno quieto (`status ∉ {streaming, await}`) cuando la dispara el operador | Nace con `titulo="nueva conversación"`, 0 turnos, `ctx 0 %`. **Acoplada a T2** |
| **T2** | `activa → inactiva` | **nunca sola** | — | Persiste `Conv` y `Checkpoint` (RF-305). Deniega los permisos pendientes con motivo (RF-311). Cierra el conductor |
| **T3** | `inactiva → activa` | **operador** (clic en fila, RF-310) | turno quieto | **Acoplada a T2.** No spawnea: el spawn es perezoso, en el `Turn` siguiente (RF-343 CA-4) |
| **T4** | `activa → activa` | **sistema**, al inicio del turno siguiente al cruce de umbral (`session_service.go:347-349`) | `RotacionPendiente == true` | `rotarLocked` (`session_rotacion.go:52-64`) sin cambios de lógica + **emite frame** (§4.3). Cero entradas nuevas en la lista |

### 3.2 Las transiciones prohibidas, y qué las impide

| Prohibida | Qué la impide |
|---|---|
| Quedar con **0 activas** | `CrearConversacion`/`ActivarConversacion` activan y desactivan en la misma operación; no hay una función que sólo desactive. `VerificarUnaActiva` en `persistLocked` es el cinturón |
| Tener **2 activas** | Ídem. Un registro editado a mano lo repara `NormalizarConversaciones` **diciéndolo** |
| **Borrar** una conversación | **No existe la operación.** No hay endpoint, no hay método de dominio. Cerrar la **sesión** entera las archiva a todas con su `Conv` (E-50) |
| **Cerrar** una conversación sin reemplazo | No hay botón (RF-309) ni ruta. El único cierre es implícito, dentro de T1/T3 |
| T1/T3 con **turno en vuelo** | Guard re-evaluado **bajo el lock** ⇒ `ErrBusy` → **409** (§4.2). El deshabilitado del FE (RF-312) es cortesía, no la autoridad |
| **Rotar una inactiva** | Sólo la activa tiene conductor y sólo ella recibe turnos. `rotarLocked` se llama desde `Turn`, que opera sobre la activa (E-20) |
| **Retomar una conversación de OTRA sesión** | `ActivarConversacion` busca `cid` **dentro de `s.Conversaciones`** y devuelve `ErrConvNoEncontrada` si no está → **404**. Nunca hay una búsqueda global por `cid` (BR-CV-2, RF-343 CA-3) |
| Que una transición **quede a medias** | §4.2 |

### 3.3 Qué le pasa al proceso en cada transición

| | T1 crear | T3 retomar | T4 rotar |
|---|---|---|---|
| `r.live` | se cierra **fuera del lock** | se cierra **fuera del lock** | `r.live.Close()` dentro de `rotarLocked` (`:53-56`) — **vigente, no se toca** |
| `ClaudeSessionID` de la que se activa | `""` (nace sin) | el suyo → próximo spawn con `--resume` (`:390`) | se limpia (`:61`) → próximo spawn **sin** `--resume` |
| `CadenaCC` | `nil` | intacta | `append(CadenaCC, ccid)` (`:58`) |
| `Checkpoint` | `""` | **intacto** (F-3: hoy se borraba al archivar) | `CheckpointMecanico(Conv)` (`:60`) |
| `Conv` | `[]` | intacto, **se repinta antes de enganchar el stream** (BR-CV-7) | `append(Conv, {RolSys, breadcrumbRotacion})` (`:63`) |
| spawn | **no** (perezoso) | **no** (perezoso) | **no** (el `Turn` que la disparó sigue y spawnea) |
| frame | `conversacion{creada}` | `conversacion{activada}` | `conversacion{rotada, turno_idx}` |

---

## 4 · Concurrencia

### 4.1 El modelo de locking que ya existe, y que no se cambia

- **Un solo candado:** `SessionService.mu sync.Mutex` (`session_service.go:115`). `sessionRuntime`
  **no tiene mutex propio**: está protegido siempre por `s.mu`.
- **Regla invariante observada en las 953 líneas:** `s.publish(...)` y **todo IO**
  (`live.Send`, `live.RespondControl`, `reindex`) ocurren **fuera** del lock. Los patrones son
  `defer s.mu.Unlock()` para las operaciones puras (`:227`, `:240`, `:251`, `:270`, `:285`) y
  Lock/Unlock manual para las que hacen IO (`Close` `:300`/`:318`, `Turn` `:330`/`:368`/`:374`,
  `consume` × 7).
- **Guardia de proceso stale, ya vigente:** todo tramo de `consume` empieza con
  `if r := s.rt[id]; r != nil && r.live == live` (`:458`, `:473`, `:485`, `:532`, `:550`, `:570`,
  `:586`).
- **Hallazgo declarado, no resuelto acá:** `spawnLocked` hace IO pesado (`Resolve`,
  `ProvisionSession`, `ResolveForRole`, `Spawn`) **bajo `s.mu`** (llamado desde `Turn` `:365` y
  `tryHealResume` `:611`), bloqueando `List`/`Get`/`Turn` de **todas** las sesiones mientras arranca
  un proceso. Preexistente. Este paquete **no lo empeora** porque la transición **no spawnea**
  (§3.3) — pero tampoco lo arregla, y queda en §9-H-D.

### 4.2 La transición atómica

Calca el patrón exacto de `Close` (`session_service.go:299-324`), que es el único del servicio que
ya cierra un proceso: tomar el lock → mutar → guardarse el `live` → soltar → `live.Close()`.

```go
// internal/usecase/session_conversaciones.go — archivo nuevo

func (s *SessionService) CrearConversacion(id string) (domain.Conversacion, string, error)
func (s *SessionService) ActivarConversacion(id, cid string) (domain.Conversacion, string, error)
func (s *SessionService) RenombrarConversacion(id, cid, titulo string) (domain.Conversacion, error)
func (s *SessionService) Conversaciones(id, q string) ([]ConversacionResumen, int, error)

// transicionLocked es el cuerpo compartido de crear y retomar. Caller holds s.mu.
// Devuelve el live a cerrar (el caller lo cierra DESPUÉS de soltar) y los frames a
// publicar (el caller los publica DESPUÉS de soltar) — la disciplina del servicio.
func (s *SessionService) transicionLocked(
    r *sessionRuntime, aplicar func(*domain.Session) (string, error),
) (viejo ports.AgentSession, frames []dockFrame, err error)
```

Orden exacto de `transicionLocked`, bajo el lock:

```
1. GUARD (re-evaluado acá, no en el cliente):
   r.meta.Status ∈ {streaming, await}  ⇒  return ErrBusy        → transporte 409
   registro en solo-lectura (§2.4-E)   ⇒  return ErrSoloLectura → transporte 503
2. SNAPSHOT de rollback: copia del agregado (Session con sus conversaciones).
3. aplicar(r.meta)  — la operación PURA del dominio (§1.4). Si falla ⇒ restaurar y salir.
4. Permisos pendientes de la que se desactiva: por cada request_id en r.pendingPerm,
   armar el frame permission_result{decision:"deny", text:"conversación desactivada por
   el operador"} y vaciar el mapa.        ← precedente literal: Interrupt, :848-852/:879
5. Resetear el estado de vuelo:  pendingTurn="" · wasResume=false · sawInit=false ·
   resumeRetried=false · assembling.Reset() · msgFlushed=false · grants = map vacío.
   runSeq NO se toca: es monótono por sesión, a propósito (§4.4).
6. r.convActiva = <id de la nueva activa>;  viejo := r.live;  r.live = nil.
7. persistLocked()  →  si devuelve error: RESTAURAR el snapshot del paso 2, r.live = viejo,
   y salir con el error. El disco manda; en memoria no queda lo que no se guardó.
8. armar frames: conversacion{creada|activada} + los permission_result del paso 4.
```

Y fuera del lock: `viejo.Close()` (si no era `nil`), después `s.publish(f)` por cada frame **en el
orden en que se armaron**.

**Un fallo nunca deja media transición** (RF-308 CA-2): los únicos pasos que pueden fallar son 1
(antes de mutar), 3 (rollback trivial) y 7 (rollback del snapshot). El `Close` del proceso viejo
ocurre **después** de que el disco confirmó, así que un fallo de persistencia no deja al operador
sin conductor.

### 4.3 La rotación emite, y emite sin romper la disciplina del lock

**Por qué hoy no emite** (C-6 / H-8): `rotarLocked` se llama desde `Turn` **con `s.mu` tomado**
(`session_service.go:347-349`), y en este servicio **nada publica bajo el lock**. No es un olvido:
es que el lugar donde está no puede publicar. La corrección respeta la regla en vez de romperla:

```go
// session_rotacion.go — la ÚNICA firma que cambia. El cuerpo (`:53-63`) no cambia una línea
// de lógica: cierra el live, encadena, checkpointea, limpia el resume y deja el breadcrumb.
func (s *SessionService) rotarLocked(r *sessionRuntime) dockFrame
//   devuelve: dockFrame{
//     SessionID: r.meta.ID,
//     Kind: "conversacion",
//     ConversacionID: <id de la activa>,
//     ConversacionEvento: "rotada",
//     TurnoIdx: len(Conv) - 1,       // índice donde quedó el breadcrumb
//     Text: breadcrumbRotacion,      // el literal de session_rotacion.go:12 (C-5: gana el código)
//     CtxPct: 0,                     // el ctx real llega con el result del turno nuevo
//   }
```

Y en `Turn`, junto al `status` que ya se publica tras soltar el lock (`:376`):

```go
// dentro del lock, :347-349
var fRot *dockFrame
if r.meta.Activa().RotacionPendiente { f := s.rotarLocked(r); fRot = &f }
...
s.mu.Unlock()
if fRot != nil { s.publish(*fRot) }                    // ← PRIMERO la rotación
s.publish(dockFrame{... Kind: "status" ...})           // ← después el status del turno
```

**El orden importa y es este**: la rotación agrega el breadcrumb a `Conv` **antes** de que se
appendee el turno del usuario (`:355`), así que el frame tiene que llegar antes para que el
transcript del FE quede en el mismo orden cronológico que el `Conv` del backend. Sin eso, la marca
aparecería **después** del mensaje que la disparó.

`rotarLocked` **no lleva `run_id`** en su frame, y es correcto: el breadcrumb no pertenece a un
turno. La idempotencia se resuelve por `turno_idx` (§4.4).

### 4.4 Idempotencia, y por qué `runSeq` se queda por sesión

**El `run_id` no baja a la conversación.** `runSeq`/`curRun` (`session_service.go:88-89`) generan
`"<sessionID>-r<n>"` (`:357-358`) desde un contador **en memoria, monótono, por sesión**. Un contador
por conversación reiniciaría en `r1` al retomar una vieja y produciría un `run_id` que el FE ya tiene
en `finalizedRun[id]` (`sessions-store.ts:38`, `:266`, `:318`) ⇒ **dropearía frames legítimos**
(R5 del relevamiento). Manteniéndolo por sesión, dos conversaciones **no pueden** compartir un
`run_id` dentro de una vida del daemon, y el dedup del FE sigue siendo correcto sin tocar una línea.

**El frame `conversacion` es idempotente por otro camino**, porque no tiene `run_id`:

| evento | idempotencia | mecanismo |
|---|---|---|
| `creada` / `activada` | **declarativa** | El frame trae el **estado post-transición** de la conversación (id, título, activa, ctx). Aplicarlo dos veces es un `set`, no un `append`: mismo resultado |
| `renombrada` | **declarativa** | Ídem: trae el título final |
| `rotada` | **por secuencia** | Trae `turno_idx` = el índice donde quedó el breadcrumb en `Conv`. El FE appendea **sólo si `conv.length === turno_idx`**. Un replay por `Last-Event-ID` llega con `conv.length > turno_idx` ⇒ se dropea |

Esto cierra E-41 sin inventar un canal ni un contador nuevo, y es lo que `plan-storybook.md` U-09
pide probar.

### 4.5 Las carreras identificadas, con su mitigación

| id | Carrera | Mitigación | Código |
|---|---|---|---|
| **CR-1** | Dos `POST …/conversaciones` casi simultáneos | `s.mu` serializa. La segunda encuentra la primera activa con 0 turnos y **crea otra igual** — E-09 lo declara legal. Invariante intacta: 2 vacías, 1 activa | §4.2 paso 3 |
| **CR-2** | El operador manda un turno entre el render del `＋` habilitado y el `POST` | El guard se re-evalúa **bajo el lock** ⇒ 409 con motivo. El FE no altera nada (E-07/E-39) | §4.2 paso 1 |
| **CR-3** | Frames del conductor **viejo** llegan después de la transición | **Ya resuelto por código vigente**: `transicionLocked` pone `r.live = nil`, y los 7 tramos de `consume` chequean `r.live == live` (`:458`, `:473`, `:485`, `:532`, `:550`, `:570`, `:586`) y descartan. **Cero código nuevo** | verificado |
| **CR-4** | El FE aplica un frame de la vieja al transcript de la nueva | `runSeq` monótono (§4.4) + el frame `activada` **reemplaza** `activa.conv` con el de la retomada, y llega **después** de que el daemon cerró el `live` viejo | §6.2 |
| **CR-5** | SSE reconecta y replaya `conversacion{rotada}` ⇒ breadcrumb duplicado | `turno_idx` (§4.4) | §6.2 |
| **CR-6** | Dos vistas (app Tauri + `:4200` en el navegador) sobre la misma sesión | El daemon es la verdad; **las dos** reciben los mismos frames por el SSE singleton (`sessions-store.ts:128-136`) y el 409 protege de dos turnos. El frame `conversacion` es justamente lo que impide que la lista de la otra vista quede stale (E-40) | §5.2 |
| **CR-7** | `persistLocked` bajo `s.mu` con el registro inflado por N `Conv` | §2.5 — presupuesto + disparador de corte, con ticket de medición obligatorio |
| **CR-8** | Dos daemons sobre el mismo `$HOME` | **No mitigada.** Preexistente (`registry.go` sin lock de archivo) → §9-H-C | — |

### 4.6 El system-prompt por sesión con N conversaciones (hueco 8 del relevamiento)

`injector.ProvisionSession(ctx, id, tarjeta)` (`session_service.go:410`) usa el **id de SESIÓN** como
key y escribe `~/.arnesia/sessions/<id>/system.md` (CAP-96); el `Checkpoint` viaja por ese mismo
archivo (`:406-408`). Con N conversaciones **dos hilos se pisarían el system-prompt**.

**No se pisan, y no hay que cambiar nada.** Bajo CV-D7 sólo la activa spawnea, y `spawnLocked` corre
**bajo `s.mu`**, así que la escritura de `system.md` y el `Spawn` que la consume son una sección
crítica: no puede haber dos provisiones concurrentes de la misma sesión. Lo que sí cambia es **de
dónde sale el `Checkpoint`** que se concatena en `:406-408`: de `r.meta.Checkpoint` pasa a
`r.meta.Activa().Checkpoint`. Una línea. **CAP-96 se conserva como está** y se anota en su prosa que
la key sigue siendo la sesión porque la conversación activa es una.

---

## 5 · API

### 5.1 Endpoints nuevos

Todos bajo `internal/adapters/transport/http/sessions_conversaciones.go` (archivo nuevo; los
handlers de `sessions.go` no se tocan salvo `listSessions`). Registro en `router.go`, **antes** de
`GET /api/sessions/{id}` para que el mux no ambigüe.

| # | Ruta | 2xx | Errores | Semántica |
|---|---|---|---|---|
| 1 | `GET /api/sessions/{id}/conversaciones` | `200 {conversaciones:[…], total}` | `404` sesión inexistente | RF-340. **No devuelve `conv`** (12 KB × N). `total` = cantidad total de la sesión |
| 2 | `GET /api/sessions/{id}/conversaciones?q=<texto>` | `200` ídem + `fragmento` por entrada | `404` | RF-341. Sólo viajan las que coinciden; **`total` sigue siendo el total** (el denominador de «N de M», RF-318). `q` en blanco ≡ sin `q` (E-26) |
| 3 | `POST /api/sessions/{id}/conversaciones` | `201 {nueva, desactivada}` | `409` turno en vuelo · `404` · `500` fs · `503` solo-lectura | RF-342. **Una** operación: desactiva y activa (RF-308 CA-1) |
| 4 | `POST /api/sessions/{id}/conversaciones/{cid}/activar` | `200 {activada, desactivada}` | `409` · `404` cid ajeno a esa sesión · `500` · `503` | RF-343. **No spawnea** (CA-4) |
| 5 | `PATCH /api/sessions/{id}/conversaciones/{cid}` body `{titulo}` | `200 {conversacion}` | `400` título vacío (**y el título no cambia**) · `404` · `500` · `503` | RF-344. Marca `titulo_editado:true`. Duplicados legales (RF-357/E-31) |

**Los dos DTO, y por qué son dos tipos y no uno con `omitempty`:**

```go
// ConversacionResumen es lo que viaja en la LISTA. No tiene Conv — no por omisión, sino
// porque el tipo no lo declara. Así `nil` vs `[]` no es expresable acá (no-aplica-no-es-cero).
type ConversacionResumen struct {
    ID, Titulo        string  `json:"id" / "titulo"`
    TituloEditado     bool    `json:"titulo_editado"`
    Activa            bool    `json:"activa"`
    Turnos            int     `json:"turnos"`               // derivado: len(Conv)
    CtxPct            int     `json:"ctx_pct"`              // 0 es dato (BR-CV-9)
    RotacionPendiente bool    `json:"rotacion_pendiente"`   // insumo del `caliente` del chip
    UltimaInteraccion string  `json:"ultima_interaccion,omitempty"`
    CreadaEn          string  `json:"creada_en"`
    ClaudeSessionID   string  `json:"claude_session_id,omitempty"`
    Model             string  `json:"model,omitempty"`
    Fragmento         string  `json:"fragmento,omitempty"`  // sólo con ?q=
}

// ConversacionActiva es lo que viaja DENTRO de una Session (GET /api/sessions y los
// frames): el resumen + el transcript. Es la única que trae Conv.
type ConversacionActiva struct {
    ConversacionResumen
    Conv     []domain.Turn `json:"conv"`               // SIN omitempty
    CadenaCC []string      `json:"cadena_cc,omitempty"`
}
```

`Checkpoint` **no viaja al FE en ningún DTO**: es el digest del system-prompt, el FE no lo pinta en
ninguna superficie del mockup, y mandarlo sería exportar contenido de conversación por un cable que
nadie lee.

### 5.2 El frame nuevo

```go
// session_service.go — dockFrame (:53-66) gana 3 campos, todos opcionales.
ConversacionID     string `json:"conversacion_id,omitempty"`
ConversacionEvento string `json:"conversacion_evento,omitempty"` // creada|activada|renombrada|rotada
TurnoIdx           *int   `json:"turno_idx,omitempty"`           // sólo en `rotada` (§4.4)
```

`Kind: "conversacion"`. Lo emiten `CrearConversacion`, `ActivarConversacion`,
`RenombrarConversacion` y `rotarLocked`. **Sin `run_id`** (§4.4). Lleva además el
`ConversacionResumen` de la conversación resultante, para que la idempotencia sea declarativa.

⚠️ **Gotcha de wire, con precedente de incidente real:** los 3 campos van **con `omitempty`**. HS-26
demostró que un campo nuevo sin `omitempty` en un tipo que roza el wire del turno user devuelve
`400 «Extra inputs are not permitted»` y **envenena la JSONL nativa del operador**
(`ledger/HS-26.md`; regresión cazada por `conductor_test.go:TestUserTurnWireSinCamposExtra:54`).
`dockFrame` es de salida y no debería tocar ese camino, pero la disciplina se aplica igual.

### 5.3 Endpoints modificados y retirados

| | Hoy | Después | Motivo |
|---|---|---|---|
| `GET /api/sessions` | array, o `{sesiones, cerradas}` con `?cerradas=1` (`sessions.go:32` vs `:41`) | **siempre array**. `?arnes=` se conserva (`:25`). **`?cerradas=1` ⇒ `400`** con motivo «el parámetro `cerradas` se retiró: las conversaciones viven en `GET /api/sessions/{id}/conversaciones`» | RF-345 CA-1. Un `400` con puntero es mejor que ignorarlo en silencio o devolver 200 vacío (BR-CV-10). Y elimina la respuesta bimorfa que hace indocumentable la operación (`ruta-servida-esta-declarada`, check `forma-de-respuesta-unica`) |
| `GET /api/sessions/cerradas/{id}/historial` (`router.go:124`) | reconstruye desde la JSONL | **se retira.** `404` por ausencia de ruta | RF-345 CA-2. Bajo CV-D8 el transcript de una archivada viaja en su propio `Conv`: no hay nada que reconstruir. `HistorialCerrada` (`session_historial.go:97-123`) y `history.Reader` **se conservan como capacidad del dominio, sin ruta HTTP** — siguen siendo el único fallback para registros archivados **antes** de la migración, cuyo `Conv` no existe |
| `POST /api/sessions` | crea la sesión | crea la sesión **con su conversación inicial** activa, `titulo="nueva conversación"` (RF-301 CA-3 — **no** hereda el `"nuevo frente"` de `session_service.go:261`) | RF-301 CA-1. ⚠ RF-301 CA-1 pide `201`; **verificar el código actual antes de cambiarlo**: si hoy responde `200`, cambiarlo es un breaking change del contrato y va declarado en el ticket, no de contrabando |
| `DELETE /api/sessions/{id}` | archiva metadata sin `Conv` | archiva la sesión **con todas sus conversaciones completas** en `sesiones-archivadas.json` | RF-306 · E-50. `canClose` sigue exigiendo ≥2 sesiones (`session-rail.tsx:193`) |

**Neto de rutas: −1, +5.**

### 5.4 El diff exacto contra `docs/architecture/contracts/api/openapi.yaml`

Estado medido hoy (generado): el router sirve **50** rutas, el contrato declara **39**, hay **11
`/api` servidas sin declarar** y **0 fantasma**. El paquete deja el saldo en **10 sin declarar**
(retira una de la deuda) y no agrega ninguna.

**(a) `paths:` — AGREGAR, después de `/sessions/{id}` (`:630-664`) y antes de `/sessions/{id}/turn`
(`:665`):**

```yaml
  /sessions/{id}/conversaciones:
    parameters: [ { $ref: '#/components/parameters/SessionId' } ]
    get:
      summary: Conversaciones de una sesión (CV-D3/CV-D4 — sólo las de ESTA sesión)
      parameters:
        - name: q
          in: query
          required: false
          schema: { type: string }
          description: >-
            Filtra por título Y por el texto de cada turno de `conv`, insensible a
            mayúsculas y acentos (NFD + strip de diacríticos). En blanco ≡ sin filtro.
            Con `q`, cada entrada suma `fragmento` y `total` sigue siendo el TOTAL de la
            sesión (el denominador de «N de M coinciden»).
      responses:
        "200":
          description: la lista. NUNCA trae `conv` — son 12 KB por conversación.
          content: { application/json: { schema: { $ref: '#/components/schemas/ConversacionesListado' } } }
        "404": { description: la sesión no existe. Jamás 200 con lista vacía. }
    post:
      summary: Crear una conversación (desactiva la anterior — UNA transición, BR-CV-4)
      responses:
        "201":
          content:
            application/json:
              schema:
                type: object
                required: [nueva]
                properties:
                  nueva:        { $ref: '#/components/schemas/ConversacionActiva' }
                  desactivada:  { type: string, nullable: true,
                                  description: id de la que quedó inactiva; null si no había ninguna }
        "404": { description: la sesión no existe }
        "409": { description: hay un turno en vuelo (streaming/await) — mismo criterio que /turn }
        "500": { description: no se pudo persistir; el estado anterior queda intacto }
        "503": { description: registro en solo-lectura (esquema en disco más nuevo que el binario) }

  /sessions/{id}/conversaciones/{cid}:
    parameters:
      - { $ref: '#/components/parameters/SessionId' }
      - { name: cid, in: path, required: true, schema: { type: string } }
    patch:
      summary: Renombrar (RF-344) — títulos duplicados son legales, no hay unicidad
      requestBody:
        required: true
        content: { application/json: { schema: { type: object, required: [titulo],
                   properties: { titulo: { type: string } } } } }
      responses:
        "200": { content: { application/json: { schema: { $ref: '#/components/schemas/Conversacion' } } } }
        "400": { description: título vacío o sólo espacios — el título NO cambia }
        "404": { description: la conversación no pertenece a esta sesión }
        "500": { description: no se pudo persistir }
        "503": { description: registro en solo-lectura }

  /sessions/{id}/conversaciones/{cid}/activar:
    parameters:
      - { $ref: '#/components/parameters/SessionId' }
      - { name: cid, in: path, required: true, schema: { type: string } }
    post:
      summary: Retomar (CV-D11) — NO spawnea; el --resume ocurre en el turno siguiente
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                required: [activada]
                properties:
                  activada:    { $ref: '#/components/schemas/ConversacionActiva' }
                  desactivada: { type: string, nullable: true }
        "404": { description: la conversación no pertenece a esta sesión (nunca se busca global) }
        "409": { description: hay un turno en vuelo }
        "500": { description: no se pudo persistir }
        "503": { description: registro en solo-lectura }
```

**(b) `/sessions:` (`:605-629`) — AGREGAR `parameters:` (hoy no tiene, `:606-614`):**

```yaml
      parameters:
        - name: arnes
          in: query
          required: false
          schema: { type: string }
          description: >-
            Filtra por id de arnés con comparación EXACTA (`sessions.go:25`): NO normaliza
            llaves. Medido 2026-07-26 (pre-migración): 4 de 5 sesiones vivas tenían id
            pelado y 3 quedaban invisibles ante una consulta con clave calificada. El
            re-key de la migración v1→v2 (CV-D16) las lleva a `(home,id,scope)`; las que
            no se pudieron resolver conservan su llave pelada y figuran en el log de
            arranque con su motivo. El filtro sigue siendo exacto a propósito: normalizar
            acá escondería el problema en vez de arreglarlo.
        - name: cerradas
          in: query
          required: false
          deprecated: true
          schema: { type: string, enum: ["1"] }
          description: >-
            RETIRADO. Responde 400 con puntero a
            GET /api/sessions/{id}/conversaciones. Se declara para que el retiro sea
            visible en el contrato, no un 400 sorpresa.
```

Y su `"200"` queda **`{type: array, items: Session}`** — que es lo que ya dice (`:614`) y que ahora
**es verdad**, porque la respuesta deja de ser bimorfa.

**(c) `components/schemas/Session` (`:856-878`) — REESCRIBIR:**

- **quitar:** `claude_session_id`, `model`, `ctx_pct`, `conv` (con su `items.rol` de
  `enum: [user, assistant, sys]`, `:877`, **stale desde `RolAct`**, `session.go:56`), `turnos`,
  `cadena_cc`.
- **agregar:** `cwd`, `reparacion`, `cerrada_en`, y **`activa`** → `$ref: ConversacionActiva`,
  **`required`**.
- **el `enum` de `rol` se corrige a `[user, assistant, sys, act]`** en el schema `Turn`, que pasa a
  ser componente propio en vez de estar inline en `conv.items` — hoy sólo existe ahí, y el paquete
  lo necesita desde dos lugares.

**(d) `components/schemas` — AGREGAR 4:** `Turn`, `Conversacion` (el resumen), `ConversacionActiva`
(el resumen + `conv` + `cadena_cc`), `ConversacionesListado` (`{conversaciones, total}`).

**(e) `components/parameters` — AGREGAR `SessionId`** (hoy el `{id}` se repite inline en 6 paths).

**(f) `docs/architecture/contracts/api/_sin-declarar.yaml` — CREAR** con las 11 medidas hoy, cada
una con su `razon:`; el paquete la deja en 10 al retirar
`GET /api/sessions/cerradas/{id}/historial`. Es la allowlist del boundary nuevo (§7.5) y **sólo
puede achicarse**.

**(g) Lo que NO se toca:** `/sessions/{id}/turn` (`:665`), `/permission` (`:687`), `/interrupt`
(`:723`), `/dictado` (`:742`) y `/events` (`:812`). El schema de `DockFrame` — si existe en el yaml —
gana los 3 campos de §5.2; si no existe, entra a `_sin-declarar.yaml` con razón, no se inventa.

---

## 6 · Frontend

### 6.1 La forma del wire

```ts
// web/src/shared/api/types.ts — Session (:16-38) se parte

export interface Turn { rol: Rol; text: string }            // sin cambios (:11-14)

export interface Conversacion {
  id: string
  titulo: string
  titulo_editado: boolean
  activa: boolean
  turnos: number
  ctx_pct: number                    // 0 es dato, no ausencia (BR-CV-9)
  rotacion_pendiente: boolean        // el `caliente` del chip sale de acá, NO de un >=40 en el FE
  creada_en: string
  ultima_interaccion?: string | undefined   // vacío ⟺ turnos === 0 (BR-CV-14)
  claude_session_id?: string | undefined
  model?: string | undefined
  fragmento?: string | undefined     // sólo con ?q=
}

export interface ConversacionActiva extends Conversacion {
  conv: Turn[]                       // SIN `?` — siempre presente, `[]` cuando está vacía
  cadena_cc?: string[] | undefined
}

export interface ConversacionesListado { conversaciones: Conversacion[]; total: number }

export interface Session {
  id: string; frente: string; arnes: string
  empresa?: string; puesto?: string; salud?: Salud
  status: SessionStatus; view: string; parked?: string
  reparacion?: boolean
  cwd?: string | undefined           // NUEVO en el wire — lo pide el detalle (RF-328)
  activa: ConversacionActiva         // NO opcional: BR-CV-1 lo garantiza (§1.4)
}
```

`GET /api/sessions` sigue costando lo que cuesta hoy: **sólo la activa trae `conv`**. Las demás
conversaciones llegan por el endpoint 1, sin transcript. Los `| undefined` explícitos son
obligatorios (`@tsconfig/strictest` ⇒ `exactOptionalPropertyTypes`, `package.json:51`), misma
disciplina que `Session` ya tiene (`types.ts:26-27`).

`cadena_cc` hoy está declarado (`types.ts:37`) y **no lo consume nadie** (verificado). Se conserva
en `ConversacionActiva` porque el detalle de identidad puede querer decir «este hilo tiene N
rotaciones», pero **si el ticket no lo pinta, se borra**: un campo del wire sin consumidor es deuda.

### 6.2 Los dos stores y el seam entre ellos

| store | dónde | qué guarda | por qué ahí |
|---|---|---|---|
| `useSessions` | `web/src/shared/store/sessions-store.ts` (vigente) | el registro espejo + los **6 mapas por `session.id`**, que **no cambian de llave** (`:34-47`) | Bajo CV-D7 hay una conversación activa por sesión: la llave por sesión ya **es** por conversación activa (§7.1) |
| `useConversaciones` | `web/src/widgets/chat-dock/model/conversaciones-store.ts` (**nuevo**) | el panel: `estado`, `conversaciones`, `total`, `busqueda`, `abierta`, `focoInicial`, `retomando`, `retomaFallo` | Patrón «el store del widget envuelve `api.*`» ya sancionado (`portafolio-picker-store.ts:1-8`). Los componentes de `ui/` son props-puras |

**El seam, y por qué va en esa dirección.** El frame `conversacion` llega a `useSessions.onDock`
(`sessions-store.ts:262-459`), pero el que tiene que refrescar la lista es el store del widget. Que
`shared/store` importe `widgets/chat-dock` es un **error de CI**: `shared-no-upward`
(`.dependency-cruiser.js:41-48`). La dirección legal es la inversa:

```ts
// sessions-store.ts gana UN campo, no una dependencia:
convRev: Record<string, number>     // bump por cada frame kind==="conversacion" de esa sesión

// conversaciones-store.ts (widgets → shared: descendente, legal) se suscribe:
useSessions.subscribe(
  (s) => s.convRev[get().sesionId ?? ""] ?? 0,
  () => { if (get().abierta) void get().reintentar() },
)
```

`no-sibling-widget-imports` (`.dependency-cruiser.js:56-63`, **severity error**) es además lo que
hace que la mudanza de `session-rail/model/conversaciones-store.ts` al dock tenga que ser **física**,
no un re-export (RF-334 CA-3). El archivo viejo y su test **se eliminan**.

**La rama nueva de `onDock`** (`case "conversacion"`), por evento:

| evento | qué hace |
|---|---|
| `creada` / `activada` | `patch(sessions, id, s => ({...s, activa: f.conversacion}))` — el frame trae el estado completo. **Limpia** `streaming[id]`, `pendingPerms[id]`, `msgFlushed[id]`, `wroteInRun[id]`. **NO limpia** `finalizedRun[id]` (§4.4) ni `scope[id]` (el alcance es de la sesión, RF-118/RF-352 CA-3). Bump de `convRev[id]` |
| `renombrada` | `patch` del `titulo`. Bump |
| `rotada` | si `conv.length === f.turno_idx` ⇒ `appendConv(list, id, "sys", f.text)` (`sessions-store.ts:80-87`, **ya existe**); si no, dropea. Bump |

⚠️ `onDock` **no tiene `default:`** (verificado, `:267-458`): un `kind` desconocido cae en silencio.
El ticket agrega el `default:` con un `console.warn` — es una línea y cierra un agujero preexistente
que este paquete ensancha al agregar un kind.

### 6.3 Selectores, y cómo NO se re-renderiza el transcript

**Regla dura: los selectores devuelven primitivos.** Es el estilo vigente
(`chat-dock.tsx:14`: `useSessions((s) => (active ? s.streaming[active.id] : undefined))`). Un
selector que devuelve un objeto nuevo re-renderiza en **cada** cambio del store con la igualdad por
default de zustand, y el repo **no usa `useShallow` en ningún lado**. Nada de este paquete lo
introduce.

```ts
// shared/store/sessions-store.ts — junto a selectActive (:464)
export const selectActive        = …                                  // sin cambios
export const selectConvActivaId  = (s) => selectActive(s)?.activa.id ?? null
export const selectCtxPct        = (s) => selectActive(s)?.activa.ctx_pct ?? 0
export const selectCtxCaliente   = (s) => selectActive(s)?.activa.rotacion_pendiente === true
export const selectTituloActiva  = (s) => selectActive(s)?.activa.titulo ?? ""
```

**El transcript no se re-renderiza al buscar** por tres razones acumulativas, en orden de fuerza:

1. **Está desmontado.** Con la lista abierta, `ConversacionesPanel` **ocupa el área del transcript**
   en lugar de `<Messages>` (design.md §1.2). No hay nada que re-renderizar.
2. **La búsqueda vive en otro store.** Teclear muta `useConversaciones.busqueda`; `Messages`
   suscribe a `useSessions`. Cero notificaciones cruzadas.
3. **Cuando el transcript sí está montado**, `Messages` recibe `session` y `streaming` por props
   desde `ChatDock` (`chat-dock.tsx:38`) y lo único que lo hace re-renderizar es un cambio en
   `activa.conv` o en `streaming[id]` — que es exactamente cuando **tiene** que re-renderizar.

`ChatDock` sigue siendo el único que lee stores; `ConversacionRow`, `CtxChip`,
`ConversacionesPanel` y `ConversacionFila` son props-puras (design.md §3). El transporte vive en el
store del widget, no en la página: patrón `portafolio-picker-store`, no inventado acá.

---

## 7 · Boundaries y fitness

### 7.1 Los que se tocan

| Boundary | Estado | Qué le pasa | Enforcer afectado |
|---|---|---|---|
| **`sesion-viva-consistente`** | `enforced`/high | **Re-enunciado → v1.2.** Los 4 checks conservan su texto: el runtime sigue siendo por sesión porque «una activa» lo hace correcto. **+1 check** `transicion-de-conversacion-atomica` | 4 verdes intactos + `arch_test.go:TestTransicionDeConversacionEsAtomica` (**nace declarado, sin escribir**) |
| **`conductor-no-parsea-jsonl`** | `proposed`/high | **Nada cambia — y hay que decirlo.** Persistir `Conv` y buscar sobre él es lo que su L1 **ordena** («captura el stream a NUESTRO event store — ese store es la API estable interna», `:37-38`). Buscar sobre `Conv` es leer nuestra API interna, no el schema ajeno | ⚠ **trampa mecánica**: `TestNoJSONLSchemaParsing` (`arch_test.go:328`) es un source-scan que busca decoders (`json.Unmarshal`/`RawMessage`) **en la misma línea o la previa** de las pistas `transcript`/`jsonl`/`.claude/projects`. `esquema.go` usa `json.RawMessage` y `conversacion.go` documenta `Conv`. **Regla para el que construye:** ningún comentario adyacente a un decoder lleva esas palabras — §7.6 |
| **`ingesta-por-allowlist-declarada`** | `enforced` | **La allowlist no cambia.** `Conv` ya se persiste hoy para las sesiones vivas (`session.go:127`); CV-D8 **no agrega ingesta, deja de destruir la que ya hay**. Ningún campo nuevo entra desde el emisor ajeno: `Turn` no cambia | sin cambios |
| **`no-aplica-no-es-cero`** | `enforced`/**error** | **Aplica tres veces** y las tres están resueltas por tipo, no por convención: `Conv` sin `omitempty` (§1.2) · el DTO de la lista **no declara** `conv` (§5.1) · `ultima_interaccion` vacío ⟺ 0 turnos, con literal propio en la UI (BR-CV-14) · la recalibración de llaves deja `sin-candidata`/`ambigua` en vez de adivinar (§2.6) | sin cambios |
| **`dominio-independiente-de-transporte`** | `enforced` | `internal/domain/conversacion.go` **no importa** `net/http`, SSE ni `database/sql` | `arch_test.go:TestDomainIndependentOfTransport:206` |
| **`codigo-traza-a-capability`** | `enforced`/**error** | **Rompe seguro** al partir `Session`: todo puntero `session.go#Session` y los archivos nuevos. §7.3 | `capability_trace_test.go:TestCapabilityPointersResolve:129` · `TestCapabilityPointerSymbolsResolve:193` · `TestCapabilityCoverage:359` · `TestCapabilityStatusConsistent:318` |
| **`fe-topologia-fsd`** · **`fe-taxonomia-componentes`** | `enforced` | La mudanza del store es **física**; el seam es descendente (§6.2) | `depcruise`: `no-sibling-widget-imports` (`:56-63`), `shared-no-upward` (`:41-48`) |
| **`fe-visual-fitness`** | `enforced` | 56 stories nuevas; **ninguna baja `a11y` a `"todo"`** (RF-353 CA-4) | `vitest --project=storybook` |
| **`permisos-gui-human-in-the-loop`** | `enforced` | Los **grants se descartan** en la transición (deny-by-default): heredarlos aprobaría algo que el operador nunca vio en ese hilo | cubierto por el check nuevo |
| **`indice-desechable-jsonl-es-verdad`** | `proposed`/high | **No aplica**, y el nodo lo dice él mismo (`:47-62`): el corpus de chat es separado, con su propio adapter read-only. Persistir en `~/.arnesia/` con el store JSON atómico **no cae acá**. ⚠ La premisa de CV-D8 «sin tocar `index.db`» describía el índice **viejo**; hoy `index.db` es SQLite real con WAL (`c2c61ba`). **La decisión no cambia y se re-funda por volumen** (12 281 B × 100 ≈ 1,2 MB), no por «el índice no existe» | sin cambios |

### 7.2 Los tests Go que rompen seguro, y qué se hace con cada uno

| Test | Por qué rompe | Qué se hace |
|---|---|---|
| `session_historial_test.go:73` (`TestCloseArchivaMetadata`) | Asserta literal `if c.Conv != nil { t.Error(...) }`. CV-D8 lo invierte | **Se renombra a `TestCloseArchivaConTranscript` y se invierte la aserción.** No se borra: el rename hace visible el cambio de ley en el diff y en el `valida:` de CAP-98. Se le suma `TestCloseConservaCheckpoint` (F-3, que ninguna decisión previa cubría) |
| `TestCapabilityPointerSymbolsResolve:193` · `TestCapabilityPointersResolve:129` | Los punteros a `session.go#Session` y a `conversaciones-store.ts` (mudado) dejan de resolver | Se actualizan las 9 capabilities (§7.3) **en el mismo commit** — R1 es `error` |
| `TestCapabilityCoverage:359` | Archivos nuevos sin capability = huérfanos | Cada archivo nuevo entra a una hoja o a `_coverage.yaml` **con razón** |
| `TestCapabilityStatusConsistent:318` | 8 capabilities son `vivo·nc` con `valida: []`; sumarles tests exige flipear a `vivo` | R4: el flip va en el mismo commit |
| `TestOneTurnAtATime:1911` · `TestFramesCarryRunID:1924` · `TestResumeAutoSana:1957` · `TestNoSilentEventDrop:1699` · `TestSessionSpawnsInArnesPath:1891` | **Los 5 hacen `svc.List()[0].ID` y dependen de que `newTestService` deje exactamente una sesión sembrada en la posición 0.** Cambiar `Create`/`seedSessions` para que nazcan con conversación toca ese helper | Se ajusta **`newTestService`/`seedSessions`, no los tests**: siguen midiendo lo mismo. Es el canario de que el reparto no cambió el comportamiento del pipe |
| `TestNoJSONLSchemaParsing:328` | Falso positivo probable (§7.1) | Redacción de comentarios (§7.6). La segunda salida legítima — sumar el prefijo a `jsonlExento` — **queda prohibida en este paquete**: el propio test dice que eso «exige tocar este test, que es exactamente el punto» (`arch_test.go:265-267`), y acá no hay nada que eximir |
| `TestDomainIndependentOfTransport:206` | Entidades nuevas en `domain` | Se cumple por construcción |

**Lo que NO rompe y conviene saberlo:** `session_rotacion_test.go` (`TestCtxHistYUmbralRotacion:16`,
`TestRotacionInvisible:67`) sigue midiendo lo mismo — `rotarLocked` cambia de **firma** (devuelve un
frame) pero no de lógica. Los tests se ajustan al valor de retorno, no a un comportamiento nuevo.

### 7.3 Capabilities — la traza del cambio de ley de CAP-98

**Nuevas** (bloque desde **CAP-140**, verificado: el máximo en uso es CAP-139):

| CAP | archivo | cubre |
|---|---|---|
| **CAP-140** | `dominio-l0/conversacion-como-entidad.yaml` | la entidad, la invariante «una activa», las 4 operaciones puras |
| **CAP-141** | `usecases/conversaciones-de-una-sesion.yaml` | crear · retomar · renombrar, con la transición atómica |
| **CAP-142** | `usecases/buscar-en-el-transcript.yaml` | el scan en memoria sobre `Conv` + fragmento + normalización |
| **CAP-143** | `indice-persistencia/migracion-de-esquema-en-disco.yaml` | el envelope, la cadena, la cuarentena, el modo solo-lectura |
| **CAP-144** | `fe-chat/panel-de-conversaciones.yaml` | lista + buscador + `＋` + retomar en el dock |
| **CAP-145** | `http-sse/frame-de-conversacion.yaml` | el `kind` nuevo y su idempotencia |

**Modificadas (obligatorio):** CAP-14 (`sesion-frente-de-trabajo`, hoy con `valida`/`scenarios`/
`business_rules` **vacíos**) · CAP-25 (`persistencia`) · CAP-52 (`superficie-rest`, el número de
endpoints) · CAP-53 · CAP-59 · CAP-68 (`chat-cc`) · CAP-69 (`acotar-alcance`) · CAP-72
(`rail-de-sesiones`, pierde `ConversacionesDelArnes` y su prosa `:33-37` sobre la clave) · CAP-96
(`tarjeta-identidad-por-sesion`, §4.6) · CAP-97 (`rotacion-de-contexto`) · CAP-98 · CAP-100.

**El procedimiento para CAP-98, que es un cambio de LEY, no de implementación.** Su cuerpo dice hoy,
textual (`:33-34`): *«archiva la metadata liviana … **sin `Conv`: la JSONL nativa es la verdad**»*.
Los cuatro pasos, en el **mismo commit**, y ninguno es opcional:

1. **El test se renombra e invierte** (§7.2). Un test que afirmaba la ley vieja no se borra: se
   convierte en el que afirma la nueva, y el `git log` muestra la conversión. Es la lección de HS-28
   («un test que afirma lo contrario del invariante deseado y que ningún otro test caza, porque el
   test ERA el defecto»).
2. **El cuerpo de la hoja se reescribe**, no se le suma un pointer. La oración nueva:
   > `Close()` archiva la sesión entera con sus conversaciones **completas** — `Conv`, `Checkpoint`,
   > `CadenaCC` y `Cwd`. La JSONL nativa sigue siendo la verdad **del contenido**; `Conv` es nuestra
   > copia de presentación (boundary `conductor-no-parsea-jsonl`, L1: capturar el stream a nuestro
   > event store) y es lo único que sobrevive a un GC del corpus, lo único sobre lo que se puede
   > buscar (CV-D8) y lo que se repinta al retomar (CV-D11). El lector JSONL se conserva como
   > fallback para registros archivados **antes** de la migración v1→v2, cuyo `Conv` no existe.
3. **`change_log` gana su entrada** — es el mecanismo de trazado que el schema ya tiene
   (`_templates/capability.template.yaml`): `story_id: 2026-07-26-conversaciones-del-panel`,
   `type: derive`, `summary: "la ley se invierte: Conv y Checkpoint se persisten al archivar
   (CV-D8 + enmienda F-3)"`. Y `business_rules[]` gana la regla con su
   `enforcement: [docs/architecture/boundaries/sesion-viva-consistente.md]`.
4. **Ficha de ledger.** Un cambio de regla de negocio es historia, y la historia vive en
   `docs/product/LEDGER.md` → `ledger/HS-29.md`. Sin esa línea, dentro de seis meses el cambio
   parece un bug fix.

**Qué garantiza la ley nueva, en una oración verificable:** *al archivar una sesión, ninguna
conversación pierde su transcript ni su checkpoint; lo que se guarda es lo que el operador vio, y es
sobre eso que se busca y se repinta.* Se verifica con `TestCloseArchivaConTranscript` +
`TestCloseConservaCheckpoint`, y se observa en disco: `sesiones-archivadas.json` trae `conv` con N
turnos y `checkpoint` no vacío si lo había (RF-305 CA-1).

**Antes de empezar:** `capabilities/INDEX.md` está **stale en 22 hojas** (le falta el módulo
`telemetria` entero). Fix: `python3 scripts/cap_doctor.py --index`. Correrlo primero, o el INDEX
queda peor.

### 7.4 Llegar a verde, y quedarse

| Hecho medido | Consecuencia |
|---|---|
| CI rojo desde **2026-07-20**, 19 de las últimas 20 corridas; el step que rompe es `fitness visual` (`ci.yml:66-68`) | El paquete **no tiene línea base** contra la cual medirse |
| La causa son las **4 stories de `new-session-picker`**: `text-warn` `#c96a2e` sobre `#ffffff` a 10px = **3,76:1** (`new-session-picker.tsx:273` y `:290`) | Es **contraste**, no la existencia del bloque |
| **79 commits sin pushear** | CI ni siquiera vio el trabajo de los últimos días |
| **lefthook no está instalado** en este working copy (`.git/hooks/` sólo tiene los 14 `.sample`, `core.hooksPath` sin setear) | Los 8 jobs pre-commit **nunca corren** |
| `openapi.yaml` con **11 rutas `/api` sin declarar** y el paso `openapi-gen-check` **inerte** (`ci.yml:73`, condicional sobre un directorio que no existe) | El drift no lo caza nadie |

**El orden del tramo 0, y por qué es ese** (detalle en `plan-desarrollo.md` T1-T6):

1. `lefthook install`. Sin el gate local, todo lo demás se descubre en CI.
2. **`chat-dock.stories.tsx` con el dock de HOY** (A-01..A-03, `plan-storybook.md` §2.1). H-1 es
   bloqueante: un superset sin baseline no es verificable.
3. **Aplicar C-3 a los dos `text-warn` del picker** — `text-foreground` + señal `--warn` no textual.
   **Dos líneas.** Es la regla que este mismo paquete ya decidió para su superficie nueva, aplicada
   al consumidor que la rompe hoy. Las 4 stories pasan **por la razón correcta** (contraste
   arreglado), no por desaparición del bloque — que ocurriría recién en el tramo 7 (RF-333) y
   dejaría CI rojo hasta entonces. La deuda del **token** sigue abierta (`BACKLOG.md:59-64`) y
   PARIDAD tiene que decirlo así.
4. **Declarar el OpenAPI como es HOY**: las 5 rutas nuevas todavía no existen, pero
   `_sin-declarar.yaml` sí, con las 11 y sus razones. No se puede poner un gate de drift sin sacar
   primero el drift de la ecuación.
5. **Escribir `openapi_contract_test.go`** (4 enforcers). Desde este punto, la ruta 12 sin declarar
   no existe.
6. **Correr el gate completo local y recién ahí pushear** los 79 commits, y **confirmar la corrida
   de CI en verde**. Si sale roja por otra causa, es un stop-the-line: el paquete no empieza sobre
   un árbol rojo.

### 7.5 Lo que se escribió en `docs/architecture/` — y por qué eso y no más

| Archivo | Qué es | Por qué trasciende al paquete |
|---|---|---|
| **`boundaries/archivo-durable-declara-su-esquema.md`** (nuevo, `proposed`, 6 checks) | Envelope + migración forward-only con respaldo + cuarentena + solo-lectura ante esquema futuro + **prohibición del wipe sobre lo durable** | El repo tiene **al menos 4 archivos durables** (`sesiones.json`, `sesiones-archivadas.json`, el registro de arneses, la ficha de descubrimiento de telemetría) y **ninguno** tenía la disciplina. La próxima vez que una forma cambie, la pregunta ya está contestada. Y la partición derivada⊥durable es la que el árbol venía tratando por omisión, con `index/store.go` como único precedente y con la política **opuesta** |
| **`boundaries/ruta-servida-esta-declarada.md`** (nuevo, `proposed`, 5 checks) | Ratchet router ⟷ OpenAPI con exención declarada y razón obligatoria | El drift no es de este paquete: son **11 rutas** de 4 módulos distintos, y **nada** lo cazaba. Sin el nodo, las 5 rutas nuevas nacen por el mismo agujero |
| **`boundaries/sesion-viva-consistente.md`** v1.1 → **v1.2** (+1 check) | Re-enunciación del sujeto («la sesión viva» = la conversación activa) + `transicion-de-conversacion-atomica` | Su L2 apunta al archivo que se parte (`session_service.go`). Dejarlo sin re-enunciar sería dejar `enforced`/high un nodo que ya no dice de quién habla |
| **`INDEX.md`** | +2 filas, `sesion-viva-consistente` corregido de 1.0/4 a **1.2/5** (la fila estaba stale respecto de su propio archivo, que iba por 1.1), totales 26/136 → **28/148**, nota de la pasada con las cifras generadas | DoD del árbol |

**Los tres candidatos que se descartaron, con razón escrita:**

- **«El buscador entra al texto del transcript.»** Es **producto**, no arquitectura: vive en
  CAP-142. Generalizado sin sujeto se leería como licencia para indexar contenido del operador, que
  es justo lo que `ingesta-por-allowlist-declarada` acota.
- **«Una activa por sesión.»** Es una **invariante de agregado**, no un boundary: no cruza capas ni
  módulos, se enforza con un predicado puro en `domain` y su test. Un boundary que sólo habla de un
  struct es decoración.
- **«El estado in-flight vive con el proceso.»** Sería una re-enunciación de lo que
  `sesion-viva-consistente` ya dice. Se absorbió en su v1.2 en vez de abrir un nodo que compite por
  el mismo sujeto — mismo criterio con que «fail-open de la instrumentación» se absorbió en
  `telemetria-de-nacimiento` v2.3.

### 7.6 Dos reglas mecánicas para el que construye

1. **Comentarios lejos de los decoders.** `TestNoJSONLSchemaParsing` (`arch_test.go:279-317`) marca
   rojo si `json.Unmarshal`/`json.NewDecoder`/`json.RawMessage` aparece en la misma línea o en la
   **siguiente** a un comentario con `transcript`, `jsonl`, `~/.claude/projects` o
   `.claude/projects`, case-insensitive, sobre `internal/usecase`, `internal/domain` y
   `internal/adapters/telemetria`. `esquema.go` (que usa `json.RawMessage`) y `conversacion.go` (que
   documenta `Conv`) caen justo ahí. **Se redacta el comentario, no se toca el test.**
2. 🔴 **Un check con enforcer que todavía no existe se declara `(pendiente — TestX)`, con el nombre
   PELADO.** Descubierto corriéndolo, no leyéndolo: el parser del motor extrae cualquier patrón
   `archivo_test.go:TestX` de la **celda del checklist** (no del frontmatter) e intenta invocarlo. Si
   el test no existe, el veredicto es **`error`**, `arnesia conformance --todo` sale con **exit 1**,
   y como `scripts/estado.sh` corre con `set -euo pipefail` **el drift-gate de CI se cae entero**.
   Medido: con los 12 checks nuevos declarados «bien» el motor pasó de `311 · error 0` a
   `323 · error 11`. Con la celda en `(pendiente — TestX)` y el enforcer **fuera** de `enforced_by:`,
   vuelve a `323 · error 0 · deferred 242` — los 12 difieren honesto. Es la convención que
   `no-aplica-no-es-cero` e `ingesta-por-allowlist-declarada` ya usaban y que no está escrita en
   ningún lado. **Al escribir el test, el ticket hace las dos mitades:** la celda pasa a
   `docs/architecture/fitness/…:TestX` y la entrada se suma a `enforced_by:`.
3. **`go-arch-lint` es allow-list default-deny** (`.go-arch-lint.yml:6-7`). No hace falta componente
   nuevo: el código nuevo vive en `store` (`mayDependOn: [domain, ports]`, `:171-172`), `usecase`
   (`:125-126`), `domain` (sin deps por diseño) y `transport-http`. **Si alguien crea
   `internal/adapters/<algo>/`, sin entrada en el yml el linter lo rechaza por ausencia** — el
   precedente literal es `logfile` (`:184-187`).

---

## 8 · Escenarios E-01…E-50 → respuesta arquitectónica

`resp.` = qué componente responde y cómo. Ninguna fila se rellena: las que no tienen respuesta se
declaran (§9).

### 8.1 Forma de la lista

| E | Componente que responde | Cómo |
|---|---|---|
| **E-01** sin conversaciones | `domain.Session.NormalizarConversaciones` (§1.4) | Al cargar: crea una activa vacía **y devuelve la reparación**, que `main.go` loguea. Nunca se pinta una lista vacía. Cinturón: `VerificarUnaActiva` en `persistLocked` |
| **E-02** una sola | `ConversacionesPanel` (props-pura) | `conversaciones.length === 1` ⇒ no renderiza el buscador; el rótulo dice `1 conversación de la sesión X` |
| **E-03** ≥4 | `ConversacionesPanel` + endpoint 1 | Lista scrolleable en el área del transcript; el `total` del wire alimenta el rótulo |
| **E-04** 0 turnos | `ConversacionFila` + `ConversacionResumen` | `turnos:0` y `ctx_pct:0` viajan **sin `omitempty`**: la fila pinta `sin turnos todavía · ctx 0 %`. 0 no se esconde |
| **E-05** sin `claude_session_id` | `CtxChip`/`IdentidadDetalle` | El campo es `omitempty`; ausente ⇒ literal vigente `◍ sin sesión CC` (`chat-dock.tsx:82`) |

### 8.2 Crear

| E | Componente | Cómo |
|---|---|---|
| **E-06** crear quieta | `SessionService.CrearConversacion` → `domain.CrearConversacion` | Transición atómica (§4.2). La anterior queda inactiva con su `Conv` persistido. Frame `creada` |
| **E-07** crear con `streaming` | **Dos capas.** FE: `＋` `disabled` + `title` (RF-312). Servidor: guard bajo lock ⇒ `ErrBusy` → **409** | §4.2 paso 1. El cliente no es la autoridad (CR-2) |
| **E-08** crear con `await` | Ídem, con motivo «esperá tu decisión de permiso» | `await` cuenta como turno en vuelo — mismo criterio que `Turn` (`session_service.go:337-338`) |
| **E-09** crear dos veces rápido | `s.mu` serializa | Dos vacías, 1 activa. **Legal por diseño**: no hay deduplicación «por vacía». Se distinguen por `creada_en` (CR-1) |

### 8.3 Retomar

| E | Componente | Cómo |
|---|---|---|
| **E-10** retomar quieta | `ActivarConversacion` + frame `activada` + `useConversaciones.retomando` | Repinta desde `Conv` **antes** de enganchar el stream (BR-CV-7). Franja efímera; detalle abierto solo |
| **E-11** retomar con turno en vuelo | Ídem E-07: `aria-disabled` + `title` en el FE, **409** en el servidor | — |
| **E-12** proceso `claude` inexistente | **Nada especial: es el camino normal** | `r.live` es memoria del daemon; tras un reinicio **ninguna** conversación tiene proceso. `spawnLocked` es perezoso (`session_service.go:364-370`) |
| **E-13** `--resume` que el CLI no reconoce | `tryHealResume` (`:600-630`), **vigente** + marca inline nueva | Limpia el `ClaudeSessionID` **de la conversación** (`:609`), respawnea fresh una vez (`resumeRetried`, `:608`), reenvía el turno (`:624-628`). El `Checkpoint` viaja por `:406-408`. **Nuevo:** frame `rotada` con texto `⟳ hilo reiniciado · checkpoint` — bajo CV-D11 el operador acaba de pedir ese hilo y un cambio silencioso de `cc-id` es indistinguible de una rotación |
| **E-14** `cwd` inexistente | `spawnLocked` `:386-388` + `Turn` `:365-369` | Activar **funciona** (es dato nuestro). El fallo aparece al primer turno: el error del resolver, **con la ruta**, y la conversación vuelve a `idle` |
| **E-15** retomar la activa | `domain.ActivarConversacion` devuelve no-op | Cero llamadas al conductor, cero repintado. La lista se cierra |
| **E-16** retomar de 0 turnos | `ConversacionRow` | Sin `claude_session_id` ⇒ **la franja `--resume` no se dibuja**. El transcript queda vacío con el copy de RF-307 |

### 8.4 Rotación

| E | Componente | Cómo |
|---|---|---|
| **E-17** rotación en medio de un turno | **Imposible por construcción, vigente** | `RotacionPendiente` se marca al `result` (`:505-507`) y se ejecuta al **inicio del turno siguiente** (`:347-349`, comentario literal: «Se rota ENTRE turnos por construcción»). No se toca |
| **E-18** rotación con el dock colapsado | `rotarLocked` + `appendConv` | La rotación ocurre igual; el frame llega igual; `appendConv` mete la marca en su lugar cronológico. `chatOpen` no lo toca ningún camino de este paquete (RF-315 CA-1). Sin toast |
| **E-19** rotación con la lista abierta | El frame `rotada` **no cambia** `conversaciones[]` | Bump de `convRev` ⇒ el panel refetchea y ve **N**, no N+1; el `ctx_pct` de la fila baja |
| **E-20** rotación en una inactiva | **Imposible**: sólo la activa recibe turnos | §3.2 |

### 8.5 Buscador

| E | Componente | Cómo |
|---|---|---|
| **E-21** buscar con una sola | `ConversacionesPanel` | El buscador no existe — nada que filtrar |
| **E-22** sin coincidencias | `ConversacionesPanel` estado 6 + `total` del wire | «Ninguna conversación de esta sesión menciona «X»» + «Se buscó en el título y en el texto de las **4**» (el `total`, no un literal) + `Limpiar búsqueda` |
| **E-23** acentos y mayúsculas | **En el servidor**, `usecase.normalizar()` | NFD + strip de diacríticos + lowercase, **un solo lugar** (RF-341 CA-3). Es superset de `coincide()` (`new-session-picker.tsx:69-74`), que sólo hace `toLowerCase()` |
| **E-24** la activa entre resultados | `ConversacionFila` | Participa como cualquiera; si no coincide **no aparece** y el `N de M` lo refleja |
| **E-25** transcript enorme | Scan lineal en `Conversaciones(id, q)` sobre `sessionRuntime.meta` **en memoria** | Medido: la más larga son **12 281 B**; 100 conversaciones ≈ 1,2 MB. Sin FTS5, sin tocar `index.db`. ⚠ El fixture del test es la conversación **real** de 90 turnos copiada a `testdata/`, no una sintética |
| **E-26** query en blanco | `strings.TrimSpace(q) == ""` ≡ sin `q` | Lista completa, sin fragmentos |
| **E-27** coincide sólo en el título | El servidor no emite `fragmento` | `omitempty` ⇒ la fila no dibuja tercera línea. No hay nada que explicar |
| **E-28** buscar con turno en vuelo | El `GET` **no pasa por el guard** de transición | Buscar no cambia estado. Sólo las acciones se bloquean |
| **E-29** muchas coincidencias | El servidor recorta **el primer** match | Un fragmento por fila. La fila no es un visor de resultados |

### 8.6 Renombrar

| E | Componente | Cómo |
|---|---|---|
| **E-30** a vacío | Tres capas coherentes | `ConversacionRow` descarta (calca `session-rail.tsx:199-200`) · el store no llama a la API (`sessions-store.ts:160-161`) · el servidor **400** y el título no cambia (`domain.ErrTituloVacio`) |
| **E-31** título duplicado | **No hay unicidad que validar** | La identidad es el `id`. Dos hilos sobre el mismo tema con el mismo nombre son un caso real; los distingue la fila (fecha + turnos + ctx) |
| **E-32** Escape | `ConversacionRow` estado local | Descarta, sale de edición, foco al botón del título (`session-rail.tsx:267-270`) |
| **E-33** renombrar mientras llega un turno | Permitido: renombrar **no toca el conductor** | No pasa por `transicionLocked` ni por el guard de turno. Si ese turno hubiera derivado el título, gana el operador: `titulo_editado` ya es `true` |
| **E-34** blur sin Enter ni Escape | Confirma | Calca `session-rail.tsx:263` (`onBlur={commit}`) |

### 8.7 Transporte y concurrencia

| E | Componente | Cómo |
|---|---|---|
| **E-35** daemon caído al listar | `useConversaciones.estado = "error"` + el motivo **tal cual** | `ConversacionesPanel` estado 2: `role="alert"` + `Reintentar`. **Jamás «0 conversaciones»** (BR-CV-10) |
| **E-36** daemon caído al retomar | El store **no altera nada** | La activa no cambia, la lista queda abierta para reintentar |
| **E-37** daemon caído al crear | Ídem | Nada se crea, nada se desactiva |
| **E-38** timeout | Mismo camino de transporte, motivo «no respondió a tiempo» | **No** se agrega timeout propio en `client.ts`: el tope vive en el daemon (`client.ts:283-284`) |
| **E-39** 409 | `ApiError.status` (`client.ts:33-42`) | Se muestra el motivo del 409 («hay un turno en vuelo»), no un genérico. El FE ya distingue 409 (`sessions-store.ts:205`) |
| **E-40** dos vistas | El **frame `conversacion`** | El daemon es la verdad; ambas vistas reciben los mismos frames por el SSE singleton (`sessions-store.ts:128-136`); el 409 protege de dos turnos. Esto es exactamente lo que el frame nuevo existe para resolver — sin él la otra vista queda stale |
| **E-41** SSE reconecta durante la retoma | `turno_idx` para `rotada`; **declarativo** para el resto | §4.4. El `run_id`/dedup no aplica: el frame no pertenece a un turno |
| **E-42** cambiar de sesión con la lista abierta | `useConversaciones` suscrito a `activeId` | La lista **se cierra** y descarta la búsqueda: su alcance es la sesión activa (CV-D4) y quedarse abierta sería mentir. Calca `cerrarPicker` (`session-rail.tsx:44-47`) |

### 8.8 Datos en disco y entorno

| E | Componente | Cómo |
|---|---|---|
| **E-43** instalación vieja | `store.AbrirRegistro` + `deV1aV2` | 5 sesiones → 5 × 1 conversación; `sessions.json` **intacto**; `ultima_interaccion` **vacío** (H-4, no se inventa) |
| **E-44** archivos corruptos | Modo C (§2.4) | Cuarentena `.corrupto-<sello>`, registro vacío, `slog.Error` **con la ruta**. `seedSessions` no tapa nada |
| **E-45** sin permisos de escritura | Modo D (§2.4) | `persistLocked` **propaga**; la mutación revierte y responde 500 con el motivo del fs |
| **E-46** eliminar las 3 cerradas | **Procedimiento manual del operador** (RF-337, 6 pasos) | El daemon **no lo hace solo**: borrar datos del operador en silencio al actualizar sería lo contrario de la honestidad del repo. La arquitectura sólo garantiza que **la migración no los lee** (§2.1) y que no romper si no se corre |
| **E-47** crash a mitad de la migración | Modo B (§2.4) | Respaldo previo + temp+rename. Al reintentar, idempotencia |
| **E-48** nodo del alcance desaparece tras reindex | `scope[id]` **sigue por sesión** | Engancha en el bump de reindex-vivo que ya existe (`sessions-store.ts:135`); rastro `sys` en el transcript; la fila 3 desaparece. **No baja a la conversación**: es el nodo del Mapa que estás mirando, no el tema del hilo (RF-352 CA-3) |
| **E-49** el arnés desaparece del Portafolio | **Nada en cascada** | El registro de sesiones es independiente del Portafolio. Listar y buscar dependen de `session_id` (BR-CV-2). El fallo aparece al primer turno, con el motivo del resolver |
| **E-50** cerrar la sesión entera | `archivarLocked` reescrito | Las N conversaciones se archivan **con su `Conv` y su `Checkpoint`** en `sesiones-archivadas.json`. El `✕` del rail se conserva; `canClose` sigue exigiendo ≥2 sesiones (`session-rail.tsx:193`) |

**Cobertura: 50 de 50 con respuesta arquitectónica.** Ninguno queda sin componente. E-46 tiene una
respuesta **deliberadamente no automática** y así está declarada, no tapada.

---

## 9 · Huecos abiertos — declarados, no resueltos

| id | Hueco | Estado |
|---|---|---|
| **H-A** | **El `Cwd` es de la Sesión y la conversación lo necesita para el join del corpus.** `history.Reader.Turnos` exige `(cwd, ccid)` (`reader.go:63-64`). Si la sesión cambiara de cwd, el historial JSONL de sus conversaciones viejas se rompe **en silencio** | **No ocurre hoy**: el cwd sale de `resolver.Resolve(Arnes)` (`session_service.go:386`) y el arnés de una sesión no cambia. Se declara porque el día que exista «mover una sesión de arnés», este es el que se rompe. Con `Conv` persistido (CV-D8) el daño baja de «se pierde el historial» a «se pierde el fallback», que es exactamente por qué CV-D8 vale más de lo que su decisión afirma |
| **H-B** | **`ultima_interaccion` de lo migrado queda vacío** | H-4 del spec. El dato no existe y no se puede inferir. La fila dice `sin fecha · N turnos` |
| **H-C** | **Dos daemons sobre el mismo `$HOME` se pisan.** `Registry` tiene mutex in-process, no lock de archivo (`registry.go:22`) | Preexistente. **No se resuelve acá** y el envelope lo empeora marginalmente (archivo más grande, ventana de escritura más ancha) → BACKLOG |
| **H-D** | **`spawnLocked` hace IO bajo `s.mu`** y bloquea el servicio entero mientras arranca un proceso | Preexistente, medido hoy. Este paquete **no lo empeora** (la transición no spawnea) ni lo arregla → BACKLOG |
| **H-E** | **2 de los 14 `s.publish` no estampan `run_id`** (`session_service.go:626`, `:812`) y `TestFramesCarryRunID` no los ejercita (sólo `init`/`delta`/`result`) | Hueco del **check**, no del código nuevo. Anotado en el changelog de `sesion-viva-consistente` v1.2 → BACKLOG. **No se tapa cambiándole el veredicto al check** |
| **H-F** | **El tamaño de `sesiones.json` con `Conv` × N no está medido bajo carga real** | §2.5: presupuesto (< 15 ms p95 con 20 conversaciones) + disparador (archivo por sesión) + **ticket de medición obligatorio** (T14). Sin la medición, «alcanza» sería un pass fabricado |
| **H-G** | **`deriveFrente` corta por bytes, no por runas** (`session_service.go:918`: `text[:maxLen]`) — puede partir un UTF-8 a la mitad | Preexistente. RF-303 CA-2 manda **reusar** `deriveFrente`, así que el bug viaja al título de la conversación. Se declara; arreglarlo es un one-liner de otro paquete |
| **H-H** | **`forma-de-respuesta-unica` del boundary nuevo no tiene enforcer determinista** | Difiere honesto en la tabla de checks. Este paquete **elimina la única instancia conocida** (`?cerradas=1`), pero no puede probar que no haya otra |
| **H-I** | **El re-key (CV-D16) no puede garantizar 4 de 4.** 3 de las 4 sesiones a recalibrar tienen **`cwd` vacío** (0 turnos, nunca spawnearon), así que dependen del fallback por id y de qué tenga el Portafolio del operador | **Resuelto en forma, no en resultado** (§2.6): las que no se puedan decidir conservan su llave con motivo (`sin-candidata`/`ambigua: N`), nunca se fusionan ni se borran, y el `--dry-run` **obligatorio** muestra el resultado real **antes** de aplicarlo. Lo que quede sin resolver se transcribe a `PARIDAD.md` como hueco, no como éxito |

---

## Apéndice · comandos corridos para esta arquitectura

```
git rev-parse --short HEAD                                  → dd460f3
git rev-list --count origin/main..HEAD                      → 79
ls .git/hooks/ | grep -v sample | wc -l                     → 0   (lefthook NO instalado)
git config core.hooksPath                                   → (vacío)
curl -s http://127.0.0.1:4200/api/version                   → 0.2.24.2607261724 · huella 7b1d636+sucio
curl -s http://127.0.0.1:4200/api/sessions                  → 5 sesiones; s6165ac75=90 turnos ctx 19;
                                                              s25123a2c=4 turnos ctx 68 rot=true;
                                                              3 con 0 turnos y cwd nulo
grep -c "cerradas" docs/architecture/contracts/api/openapi.yaml   → 0
python3 (router mux.Handle+HandleFunc ⟷ openapi paths)      → sirve 50 · declara 39 ·
                                                              11 sin declarar · 0 fantasma
ls internal/adapters/store/*_test.go                        → (ninguno)
grep -rhoE "CAP-[0-9]+" docs/ | sort -V | tail -1           → CAP-139  ⇒ el bloque nuevo arranca en CAP-140
```

**NO VERIFICADO** (declarado, no medido):

- El costo real de `persistLocked` con el registro inflado (H-F) — es el objeto del ticket T14.
- Si `POST /api/sessions` responde hoy `200` o `201` (§5.3): el ticket lo verifica antes de tocarlo.
- `go test ./... -race` (el relevamiento corrió sin `-race`; CI sí lo corre).
- Si `DockFrame` tiene schema propio en `openapi.yaml` (§5.4-g).
