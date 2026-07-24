# Hallazgos — índice SQLite real + watcher fsnotify (fase 5)

> Segunda pasada de investigación (2026-07-24), sobre `INDEX.md` de este paquete. Confirma con
> evidencia fresca los hallazgos preliminares, resuelve **T0** con cita de flujo real, y corrige
> 2 drifts que la primera pasada no había tocado (boundary doc + comentarios de código en 3
> archivos). Cero código de índice/watcher construido — eso es Fase 3, agente 3.

## 1. Estado real de código

### `internal/adapters/index/store.go` (íntegro, 174 líneas)

- `Store` = `map[string]domain.Graph` + `sync.RWMutex` (líneas 26-29). **Cero persistencia a
  disco** — no hay `os.WriteFile`, no hay ningún archivo `.json`/`.db` involucrado en todo el
  paquete (confirmado por grep: `ls internal/adapters/index/` solo tiene `store.go` +
  `store_test.go`).
- `New()` (línea 35-39) llama `seed()`.
- `Rebuild(ctx)` (línea 43-46) **no lee nada del disco real** — vuelve a llamar `seed()`.
- `Query`/`List`/`Upsert` (líneas 49-90) son reales, con tests. `Upsert` exige `clave != ""` y
  `g.Arnes != nil` (línea 79-90) — un grafo sin manifiesto no es indexable.
- `seed()` (línea 93-163) carga un arnés demo hardcodeado + 2 grafos embebidos
  (`dogfood.DevFullCycleJSON`/`ContentStudioFullJSON`).
- Comentarios de wording "JSONL" en líneas 1-2 y 42 (package doc + TODO de `Rebuild`) —
  **dejados intactos deliberadamente**: la instrucción de tarea nombra explícitamente `store.go`
  como archivo a NO tocar ("código de producción del propio índice"). Ver §6 — quedan
  documentados acá como corrección pendiente para cuando el agente 3 reescriba el archivo.

### `internal/adapters/watch/watcher.go` (íntegro, 30 líneas)

- Stub literal. `Watch(ctx)` (línea 23-30) devuelve un channel que nunca emite y se cierra con
  el ctx. Comentario de package (líneas 1-5) trae el mismo wording "JSONL" — igual que arriba,
  **no tocado** por instrucción explícita.

### `internal/ports/index.go` — interfaz `IndexPort` (corregido, ver §6)

Comentario original decía "IndexPort is the disposable index over the JSONL source of truth" y
"Rebuild reconstructs the whole index from the JSONL corpus" — desalineado con lo que el código
real hace. Corregido a wording preciso (líneas 9-15 ahora).

### `internal/ports/watch.go` — interfaz `WatchPort` (corregido, ver §6)

Mismo problema: 4 comentarios con wording "JSONL corpus" que en realidad describen el árbol de
arneses. Corregidos.

### Callers reales (grep exhaustivo)

- `cmd/arnesia/main.go:118-121` — boot: `idx := index.New()` + `idx.Rebuild(ctx)` (re-seed, no
  lee disco).
- `cmd/arnesia/main.go:125-134` — `loadArnesDir(id, path)`: wrapper que llama
  `loader.LoadArnes(path)` y luego `idx.Upsert(ctx, id, g)` bajo la clave que el caller ya
  decidió (nunca re-derivada de `g.Arnes.ID`).
- `cmd/arnesia/main.go:148-154` — al boot, cada entrada de `arnesReg.List()` se re-carga al
  índice vía `loadArnesDir` (el directorio es la verdad, el índice se repuebla).
- `cmd/arnesia/main.go:156-160` — `watcher := watch.New()` + `watcher.Watch(ctx)` — el channel
  de eventos se guarda en `events` pero **nunca se consume de forma útil**: el goroutine en
  líneas 300-305 hace `for range events { // TODO(fase 5): idx.Rebuild(ctx) then
  broker.Publish(...) }` — cuerpo vacío, drena el channel y no hace nada (el channel del stub
  jamás emite igual, así que hoy es 100% muerto).
- `cmd/arnesia/main.go:431-436` (`runIndex`, subcomando CLI `arnesia index` sin argumento) —
  mismo patrón: `index.New()` + `Rebuild` (re-seed) + un `fmt.Println` que dice explícitamente
  *"arnesia index: rebuilt (in-memory; el índice SQLite llega con el indexer JSONL)"* — el
  propio mensaje de CLI usa el wording "indexer JSONL" heredado (no tocado; es un string de
  producto, no un comentario, fuera del alcance de "comentario engañoso" y el subcomando es
  legítimamente sobre reconstruir vía disco cuando se le da un `<dir>`).
- `cmd/arnesia/main.go:477-483` (`newPortafolioService`) — el 5° puerto del `PortafolioService`
  es el `ports.IndexPort` real (`idx` en `serve`, `nil` en el subcomando CLI `portafolio`).
- `internal/usecase/portafolio.go:316` — `ObservarEnMapa`: `s.indice.Upsert(ctx, clave, g)`
  tras `s.cargar.Load(installPath)` (línea 292) — el `cargar` es `ports.ArnesLoader`, wrapper
  de `loader.LoadArnes`. **Confirma T0**: indexa el árbol físico de un arnés, no una
  conversación.
- `internal/usecase/session_reindex.go:57` — `NewTurnReindexer`: `idx.Upsert(ctx, arnesID, g)`
  donde `g, err := load(cwd)` (línea 44) y `load` es el parámetro `func(dir string)
  (domain.Graph, error)` inyectado desde `main.go` como `loader.LoadArnes` (composición real:
  `cmd/arnesia/main.go:290`, `usecase.NewTurnReindexer(idx, loader.LoadArnes, ...)`). Se dispara
  DESPUÉS de cada turno del chat (RF-184) — recarga el **árbol de archivos del cwd de la
  sesión**, no la transcript JSONL de esa sesión.
- `internal/usecase/map_service.go`, `run_service.go:113`, `fuente_service.go:45` — todos
  consumen `index.Query`/`index.List` de solo lectura; ninguno escribe.
- Otros usuarios de `IndexPort` en `internal/usecase/`: `fuente_service.go`, `portafolio.go`,
  `portafolio_test.go` (mock), `run_service.go`, `session_reindex.go`, `map_service.go` — los 6
  archivos que aparecen en el grep de la instrucción, todos confirmados arriba.

### `internal/ports/portafolio.go` (firmas exactas, confirmado)

```go
type ArnesLoader interface { Load(dir string) (domain.Graph, error) }
type PortafolioScanner interface {
    Escanear(ctx context.Context, root string) ([]domain.HallazgoInstalacion, error)
}
type PortafolioStore interface {
    Listar() ([]domain.EntradaPortafolio, []domain.EntradaCorrupta)
    Upsert(e domain.EntradaPortafolio) error
    Desvincular(clave string) (bool, error)
    Checkouts() []string
}
```

### `internal/adapters/history/reader.go` (confirmado FUERA de alcance)

Paquete separado, sin relación al índice: `Reader.Turnos(cwd, claudeSessionID)` parsea
`~/.claude/projects/<dir-del-cwd>/<ccid>.jsonl` línea por línea para reconstruir turnos
user/assistant del chat (RF-201, boundary `conductor-no-parsea-jsonl.md`). Es LA lectura
legítima de JSONL de conversaciones del repo — nadie lo llama desde `index`/`watch`, ni al
revés. Wireado en `main.go:283-287` (`history.New("")` → `sessionSvc.SetHistoryReader`), 100%
desacoplado del `IndexPort`.

## 2. Resolución de T0 (con evidencia dura)

**El corpus real que el índice indexa es el árbol de arneses (`.claude/` de cada directorio
registrado/escaneado), vía el Portafolio — NO son transcripts JSONL de sesiones de chat.**

Evidencia:

1. `session_reindex.go:57` — el dato que entra a `idx.Upsert` es el resultado de
   `loader.LoadArnes(cwd)`, que parsea `.claude/skills|hooks|rules` etc. del directorio de
   trabajo (nomenclatura-arnes.md v1) — cero relación con `.jsonl`.
2. `portafolio.go:316` — mismo patrón: `s.cargar.Load(installPath)` → `s.indice.Upsert`.
   `cargar` es `ports.ArnesLoader`, y su única implementación real
   (`arnesLoaderFunc(loader.LoadArnes)` en `main.go:482`) es el mismo loader de árbol.
3. `main.go:125-134` (`loadArnesDir`) — tercera vía de entrada, mismo loader.
4. Los 3 puntos de entrada a `Upsert` en todo el repo (`main.go` boot/CLI, `portafolio.go`,
   `session_reindex.go`) llaman, directa o indirectamente, a `loader.LoadArnes` — **nunca** hay
   un parser de líneas `.jsonl` en el camino hacia `Upsert`.
5. El parser de JSONL real del repo (`history.Reader.Turnos`) vive en un paquete
   (`internal/adapters/history/`) que ningún caller conecta a `IndexPort`/`WatchPort`.

**Conclusión:** el wording "JSONL" en `store.go`, `ports/index.go`, `ports/watch.go` (y
`map_service.go`, que la primera pasada no había señalado) es la analogía L1 de industria
heredada sin ajustar al L2 concreto de este árbol — texto explicativo, no dicta comportamiento
(nadie lo usa como excusa de diseño real). Corregido en comentarios donde la instrucción lo
permitía (§6); el boundary doc también corregido con esta resolución documentada in extenso
(v1.2, ver §3).

Consecuencia para el diseño (insumo para `spec.md`): `Rebuild()` real = `Listar()`/`Escanear()`
del Portafolio → por cada entrada conocida, `Load(dir)` → `Upsert(clave, g)` en el store SQLite
nuevo. El watcher fsnotify real debe observar el/los directorio(s) de árboles de arneses
registrados (no `~/.claude/projects`).

## 3. Diseño ya firmado (boundary doc, ahora v1.2)

`docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md` — **corregido esta pasada**
(v1.1 → v1.2, `status` sigue `proposed`, mismos 4 checks/enforcers, sin cambios de severidad):

- **`modernc.org/sqlite`** — SQLite pure-Go, sin CGO (por eso DuckDB queda descartado del core;
  `TestNoDuckDBOrCGOStore` ya pasa hoy — ver §4).
- **2 handles `*sql.DB`:** writer con `SetMaxOpenConns(1)` (serializa escritura) + reader
  pooled. `journal_mode=WAL` + `synchronous=NORMAL` + `busy_timeout=5000` + `BEGIN IMMEDIATE`
  en transacciones de escritura.
- **No hay migraciones incrementales.** Tabla meta `schema_version`; en mismatch → se borra el
  `.db` entero y se re-indexa desde la fuente (índice desechable, CQRS-lite).
- **Rebuild rápido:** cursores `{path,inode,size,offset}` persistidos → reanuda incremental tras
  reinicio (nota: esto tiene más sentido para archivos grandes con offset — evaluar en el spec
  si aplica al árbol de arneses, que es finito y chico, o si es vestigial del wording JSONL
  original; ver ticket T3 del INDEX.md, marcado condicional a T0).
- Bulk-load con `synchronous=OFF`+`journal_mode=MEMORY`, después vuelve a WAL.
- **Corregido v1.2:** (a) el "store JSON atómico" que v1.1 atribuía al índice **no existe** — el
  índice hoy es únicamente el map in-memory, cero disco; el JSON atómico real del árbol
  (`internal/adapters/store/registry.go` + `arnes_registry.go`) persiste sesiones y el registro
  arnés→path, una cosa completamente distinta que v1.1 conflaba con el índice. (b) T0 resuelto y
  documentado in extenso en el propio boundary (la fuente durable es el árbol de arneses vía
  Portafolio, no la conversación JSONL).

## 4. Los 3 checks `deferred` — evidencia exacta

`docs/architecture/fitness/arch_test.go`:

- `TestIndexRebuildsFromJSONL` (líneas 344-346): cuerpo completo es un único
  `t.Skip("TODO(fase 5): borrar el .db y re-indexar produce el mismo estado consultable (índice
  desechable).")` — sin ninguna aserción.
- `TestSchemaVersionTriggersRebuild` (líneas 348-350): igual, único
  `t.Skip("TODO(fase 5): mismatch de schema_version borra y reconstruye; no hay migración
  incremental.")`.
- **`writer-serializado`** (fila del checklist, columna enforcer = literal `arch_test.go`, sin
  `:TestX`): el motor de conformance (`internal/adapters/conformance/mechanism/adapters.go`,
  función `ArchTest.Run`, líneas 43-82) llama `testNameOf(c.EnforcedBy)` (líneas 103-111); si el
  `enforced_by` no trae `:` con un nombre de test después, `testNameOf` devuelve `""` y `Run`
  responde inmediatamente `VeredictoDiferido` con detalle *"enforcer genérico `arch_test.go` sin
  función nombrada — sin test específico que correr"* (línea 54-55 de `adapters.go`). Es un
  mecanismo genérico, no algo específico de este boundary — el mismo patrón que causa
  `deferred` en `ctx-derivado-etiquetado` (confirmado, mismo código).
- `TestNoDuckDBOrCGOStore` (línea 209-213): **ya existe y pasa hoy** — verificado corriendo
  `go test -run '^TestNoDuckDBOrCGOStore$' -v ./docs/architecture/fitness/...` → `PASS`. Solo
  verifica ausencia de imports `duckdb`/`mattn/go-sqlite3` en `internal/domain`,
  `internal/usecase`, `internal/adapters/index` vía `assertNoImport` (arch_test.go:104-115,
  helper genérico que grepea imports reales del árbol Go). No es parte de esta deuda — no
  tocado, confirmado.

## 5. Los 3 capability YAML — estado verificado

`docs/product/capabilities/indice-persistencia/`:

- **`indice-de-arneses-en-memoria.yaml`** (CAP-21, `status: vivo`) — punteros
  `internal/adapters/index/store.go#Store`, `#Store.Upsert`, `#Store.Query` — **todos
  resuelven** (símbolos confirmados en el archivo). `valida:` lista 5 tests
  (`TestListPortfolio`, `TestUpsertIndexaBajoLaClaveNoElArnesID`,
  `TestUpsertDosArnesesMismoIDNoColisionan`, `TestUpsertClaveVaciaError`,
  `TestUpsertSinManifiestoError`) — **todos existen** en `store_test.go` (confirmado por grep de
  `^func Test`). Estado `vivo` es coherente con la realidad (Upsert/Query/List son reales y
  testeados). **No editado.**
- **`reconstruccion-del-indice.yaml`** (CAP-22, `status: parcial`) — punteros
  `internal/adapters/index/store.go#Rebuild`, `#seed` — **resuelven**. `valida:` lista
  `TestSeedServesDogfood` (existe, confirmado) + una entrada rara `~/.claude` (no es un nombre
  de test Go — probablemente un placeholder heredado del barrido HS-18 apuntando a la fuente
  aspiracional). No es una inconsistencia de `status` (que sigue siendo `parcial`, correcto: el
  `Rebuild` real existe como función pero es el mismo `seed`), así que no la traté como
  "sorpresa" que ameritara edición — la dejo señalada acá para que el agente 2 decida si vale
  limpiarla al tocar este yaml en la Fase 3. **No editado.**
- **`observar-cambios-del-corpus.yaml`** (CAP-23, `status: stub`) — puntero
  `internal/adapters/watch/watcher.go#Watch` — **resuelve**. `valida: []` (vacío, coherente con
  stub sin tests). Estado `stub` es correcto: el archivo es 100% no-op. **No editado.**

Nota: el directorio también tiene `persistencia.yaml` (CAP-25) y
`registro-arnesworking-dir.yaml` (CAP-24), fuera del alcance pedido — los leí para descartar
confusión con el índice: **CAP-25 es el store JSON atómico real del repo**
(`internal/adapters/store/registry.go#Save` + `arnes_registry.go#saveLocked`), pero persiste
**sesiones y el registro arnés→path**, no el grafo del índice — es justamente la fuente de la
confusión que corregí en el boundary doc v1.2 (§3). Ninguno de los dos yaml necesitó edición.

## 6. Corregido en esta pasada

| Archivo | Qué cambié | Motivo |
|---|---|---|
| `docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md` | v1.1→v1.2: (a) borré la afirmación falsa "store JSON atómico" del índice (no existe — es solo map in-memory); (b) agregué la resolución de T0 in extenso (fuente durable = árbol de arneses vía Portafolio, no JSONL de conversación); (c) changelog con fecha+motivo | Drift confirmado por grep: `internal/adapters/index/` no tiene ningún `os.WriteFile`/`atomic`; el "JSON atómico" real (`internal/adapters/store/`) persiste otra cosa (sesiones/registro), no el grafo. El doc afirmaba algo que el código no hace. |
| `docs/architecture/INDEX.md` (fila de la tabla de boundaries) | Versión 1.1→1.2, resumen "JSONL = verdad; el índice (in-memory+JSON)…" → "Árbol de arneses = verdad; el índice (in-memory, sin persistencia hoy)…" | Sincronizar con el boundary doc corregido; confirmé que esta tabla es manual, no generada por script. |
| `internal/ports/index.go` (comentarios de `IndexPort`, líneas 9-15) | "JSONL source of truth"/"JSONL corpus" → wording preciso: árbol de arneses vía Portafolio, referencia cruzada a `history` para la distinción | Wording engañoso — texto explicativo que no dicta comportamiento (T0), permitido corregir por instrucción explícita de tarea (`ports/index.go` nombrado). |
| `internal/ports/watch.go` (comentarios de `WatchOp`/`WatchEvent`/`WatchPort`, 4 lugares) | Mismo ajuste: "JSONL corpus"/"~/.claude JSONL" → "árbol de arneses"/"watched arnés tree" | Igual, archivo nombrado explícitamente en la instrucción. |
| `internal/usecase/map_service.go` (comentarios de `MapService`/`Rebuild`, líneas 12-13 y 45) | Mismo ajuste — no estaba en la lista explícita de la instrucción pero es el mismo wording engañoso, descubierto en la misma investigación (grep de "JSONL corpus" en `internal/`) | Coherencia: dejar este comentario con el wording viejo mientras se corrigen `ports/index.go`/`ports/watch.go` hubiera sido inconsistente — es la misma categoría de "comentario, no lógica". |

Verificación tras los cambios: `go build ./...` limpio; `go test ./internal/... ./docs/architecture/fitness/...` — todo `ok` (incluye `TestNoDuckDBOrCGOStore` pasando).

**Deliberadamente NO tocado** (aunque tienen el mismo wording "JSONL" engañoso):
`internal/adapters/index/store.go` (líneas 1-2, 42) e `internal/adapters/watch/watcher.go`
(líneas 1-5) — la instrucción de tarea los nombra explícitamente en la lista de archivos a NO
tocar ("código de producción del propio índice/watcher... esos son el objeto de la Fase 3, no
tuyo"), aunque en la sección 3 de la consigna se sugería que sí podían tocarse como
"comentario chico". Ante la contradicción, prevalece la prohibición explícita y nombrada por
archivo. **Queda para el agente 3**, que de todos modos va a reescribir estos dos archivos al
construir el store real.

## 7. Gaps reales (no tocar, son la Fase 3)

- **Store SQLite real** (`modernc.org/sqlite`, WAL, 2 handles, `schema_version`, cursores de
  reanudación) — no existe ni una línea; `internal/adapters/index/` solo tiene el map
  in-memory.
- **`Rebuild()` real** — hoy siempre re-seedea; falta: `Listar()`/`Escanear()` del Portafolio →
  `Load(dir)` por cada entrada → `Upsert`.
- **Watcher fsnotify real** — `watcher.go` es un stub de 27 líneas; el channel jamás emite.
- **Consumo del canal de eventos en `main.go:300-305`** — hoy el goroutine que debería reindexar
  tras un evento de filesystem tiene el cuerpo vacío (`for range events { // TODO }`) — cuando
  el watcher real exista, este goroutine también necesita implementarse (llamar
  `idx.Rebuild(ctx)` + `broker.Publish(sse.EventMap, delta)` según el TODO ya escrito ahí).
- **Los 2 `t.Skip`** (`TestIndexRebuildsFromJSONL`, `TestSchemaVersionTriggersRebuild`) — activar
  con implementación real.
- **Nombrar el enforcer `writer-serializado`** con un test específico (hoy genérico
  `arch_test.go`, por eso `deferred`).
- **Decisión de diseño abierta** (para el spec, no para mí): si los cursores
  `{path,inode,size,offset}` de "Rebuild rápido" tienen sentido para un corpus finito y chico
  como el árbol de arneses (a diferencia de archivos JSONL grandes que crecen por líneas) — el
  `INDEX.md` de este paquete ya lo marca como T3 condicional a T0; con T0 resuelto (corpus =
  árbol de arneses, no líneas JSONL), este ticket probablemente se simplifica o cae — decisión
  del agente 2/3, no mía.
- **Nombre del boundary/slug `indice-desechable-jsonl-es-verdad`** — el propio id de la regla
  sigue codificando "jsonl" en el slug pese a que L2 ya no lo dice literalmente (el título H1 y
  el `regla:` del frontmatter tampoco los toqué). Renombrar el archivo tocaría
  `enforced_by`/referencias cruzadas en `arch_test.go`, `BACKLOG.md`, `ledger/HS-16.md`, y otros
  3 paquetes de `stories/` que lo citan por nombre (confirmado por grep) — decisión de diseño
  más grande, no una corrección trivial de esta pasada. Señalado para que el agente 2 lo
  considere explícitamente en el spec (mantener el nombre como analogía L1 histórica, o
  renombrar con su propio ticket de migración de referencias).

## 8. `domain.Graph` — lo que el futuro store SQLite tiene que persistir

`internal/domain/graph.go`:

```go
type Graph struct {
    Arnes     *Arnes `json:"arnes,omitempty"`
    Nodes     []Box  `json:"nodos"`
    Edges     []Edge `json:"edges,omitempty"`
    Degradado bool   `json:"degradado,omitempty"`
}
```

- `Arnes` (línea 32-46): manifiesto — `ID`, `Nombre`, `Descripcion`, `Rol`, `Proceso`,
  `Empresas []string`, `ReportaA *string`, `Canal`, `Marketplace`, `Version`,
  `FuenteManifiesto`, `Fases []Fase`, `Spine *Spine`.
- `Spine` (línea 71-77): `Inicial string`, `Terminales []string`, `Estados []string`,
  `Categorias map[string]Categoria`, `Transiciones []Transicion{De,A}`.
- `Nodes []Box` (`internal/domain/box.go:198-213`): `ID`, `Clase`, `Nombre`, `Banda`, `Fase`,
  `Estado`, `ReportaA`, `Canal`, `FuentePath`, `Procedencia`, `Origen`, `Contract *Contract`
  (bloque grande: why/capabilities/constraints/non_goals/clase/arquetipo/perfil/caja/fase/
  estado/necesita/entrega/ruta/gate/handoff — `box.go:226-249`).
- `Edges []Edge` (`graph.go:19-23`): `De string`, `A string`, `Tipo TipoEdge` (invoca|lee|escribe).

Para T1 del plan (schema), el punto de partida más lazy es **1 tabla JSON-blob por clave**
(columna `clave TEXT PRIMARY KEY`, `graph_json BLOB/TEXT`, `updated_at`) — serializa/deserializa
con el mismo `encoding/json` que ya usa `decodeGraph` (store.go:167-173); normalizar a tablas
por nodo/edge solo si hace falta query granular (SQL `WHERE`) sobre nodos individuales, cosa que
hoy ningún caller necesita (`Query`/`List`/`Upsert` operan siempre sobre el `Graph` entero).

## Archivos tocados (confirmado por `git status`)

```
 M docs/architecture/INDEX.md
 M docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md
 M internal/ports/index.go
 M internal/ports/watch.go
 M internal/usecase/map_service.go
?? docs/product/stories/2026-07-24-indice-sqlite-watcher-fase5/hallazgos.md   (este archivo)
```

`docs/product/BACKLOG.md` aparece modificado pero es **preexistente de la sesión anterior**
(ya estaba en `git status` al arrancar esta tarea, según el estado inicial reportado) — no lo
toqué en esta pasada.
