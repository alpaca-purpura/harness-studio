# Arquitectura — índice SQLite real + watcher fsnotify (fase 5)

> Fase 3 del paquete (agente 3, 2026-07-24). Base: `hallazgos.md` (T0) + `decisiones.md` (D1-D6) +
> `spec.md` (RF-207..RF-214). Este documento describe el diseño AL DETALLE MÍNIMO que satisface los
> RF — no reintroduce nada que D1/D3/D4 ya descartaron (cursores de reanudación, rename del
> boundary, tablas normalizadas). Escrito junto con la implementación (ambas en la misma corrida);
> describe lo que el código realmente hace, verificado contra los 5 comandos de CI + `conformance
> --todo` (ver `PARIDAD.md`).

## 0 · Archivos tocados/creados

| Archivo | Qué cambió |
|---|---|
| `go.mod`/`go.sum` | `+modernc.org/sqlite v1.54.0`, `+github.com/fsnotify/fsnotify v1.10.1` (y transitivas). `go` directive `1.23.0→1.25.0` (mínimo que `go get modernc.org/sqlite@latest` exige — inofensivo, CI usa `go-version: "stable"`, sin pin). |
| `internal/adapters/index/store.go` | **Reescrito completo** (mismo archivo, no split — ver §1 «por qué un solo archivo»). `Store` pasa de `map[string]domain.Graph` a `modernc.org/sqlite` (2 handles). |
| `internal/adapters/index/store_test.go` | Tests existentes adaptados a la nueva firma de `New` (mismos nombres — los capabilities que los listan en `valida:` no se rompen) + 6 tests nuevos (RF-207/208/209, TDD). |
| `internal/adapters/watch/watcher.go` | **Reescrito completo** (27 líneas de stub → `fsnotify` real). |
| `internal/adapters/watch/watcher_test.go` | **Nuevo** — el stub no tenía tests (nada que romper). |
| `cmd/arnesia/main.go` | Rewire: reordena `arnesReg` antes de `idx` (Rebuild la necesita) · el loop de boot que poblaba el índice se elimina (vive dentro de `Rebuild`) · el consumidor de `events` (antes TODO vacío) reindexa incremental · `runIndex` (subcomando CLI) adaptado a la firma nueva de `index.New` · +1 helper `ownerOf`. |
| `docs/architecture/fitness/arch_test.go` | Los 2 `t.Skip` activados con aserciones reales + 1 test nuevo (`TestWriterSerializedSingleConn`) + 1 fake (`regDogfood`) + imports nuevos. |
| `docs/architecture/fitness/.go-arch-lint.yml` | +2 vendors (`sqlite`, `fsnotify`) + `canUse` en los componentes `index`/`watch`. |
| `docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md` | v1.2→v1.3: `enforced_by` gana `TestWriterSerializedSingleConn`, checklist `writer-serializado` nombrado, L2 corregido a la realidad cableada, changelog. |
| `docs/product/capabilities/indice-persistencia/{reconstruccion-del-indice,observar-cambios-del-corpus,indice-de-arneses-en-memoria}.yaml` | RF-212 — pointers/valida/status actualizados (detalle en §5). |
| `internal/usecase/map_service_test.go`, `internal/usecase/fuente_service_test.go` | Colateral inevitable: ambos llamaban `index.New()` sin argumentos; adaptados a la firma nueva (`filepath.Join(t.TempDir(), "index.db"), nil, nil`) — cero cambio de intención de los tests. |
| `internal/adapters/loader/loader_test.go` | 2 líneas: `os.Chdir`+`os.Getwd`+`t.Cleanup` manual → `t.Chdir()`. Causal: el bump de `go.mod` a 1.25.0 (arriba) hizo que `usetesting` empezara a sugerir `t.Chdir()` (stdlib desde go1.24) donde antes (go1.23 declarado) no aplicaba — `golangci-lint run ./...` completo lo marcaba rojo; se corrige en la raíz, no se lo deja como deuda ajena. |

## 1 · Decisión: reescribir `store.go`/`watcher.go` in-place, un solo archivo

`hallazgos.md` ya señalaba que estos dos archivos eran «el objeto de la Fase 3», y `spec.md`/
`decisiones.md` no piden separar el store SQLite en varios archivos. Se mantiene **un archivo por
paquete** (no `store.go`+`schema.go`+`rebuild.go`) por una razón concreta de este repo: el boundary
`codigo-traza-a-capability.md` (R2, `TestCapabilityCoverage`) exige que **todo archivo fuente** bajo
`internal/`/`cmd/` esté reclamado por ≥1 capability — cada archivo nuevo es una entrada más a
mantener en `docs/product/capabilities/`. Ladder regla 4 (fewest files): dividir el store en 3-4
archivos de ~100 líneas cada uno no gana nada (mismo paquete, mismos tests, sin reuso cruzado) y
sólo agrega bookkeeping de capabilities. Un archivo de 412 líneas es grande pero legible por
secciones (constructor → Rebuild → CRUD → seed → helpers privados).

## 2 · `internal/adapters/index/store.go` — diseño

### 2.1 Tipo y constructor

```go
type Store struct {
    writer *sql.DB
    reader *sql.DB
    reg    ports.ArnesRegistry                      // puede ser nil
    load   func(dir string) (domain.Graph, error)    // puede ser nil
}

func New(path string, reg ports.ArnesRegistry, load func(dir string) (domain.Graph, error)) (*Store, error)
func (s *Store) Close() error   // no es parte de ports.IndexPort; hygiene opcional
```

`reg`/`load` viven en el `Store` (no como parámetros de `Rebuild`) porque `ports.IndexPort.Rebuild`
tiene la firma fija `Rebuild(ctx context.Context) error` (spec: «ningún caller cambia») — la única
forma de que `Rebuild()` sepa su fuente sin tocar el puerto es inyectarla al construir. `reg`/`load`
nil es válido: cualquier caller que solo ejercite `Query`/`List`/`Upsert` (tests, o un `Store` que
nunca hace boot completo) no los necesita; un `Rebuild()` sobre un `Store` así simplemente vacía la
tabla (mismo resultado que «ArnesRegistry vacío», RF-207 escenario 3).

`path == ""` → default `~/.arnesia/index.db` (mismo patrón que `store.NewRegistry`/
`store.NewArnesRegistry`, `resolvePath`).

### 2.2 Schema SQL exacto

```sql
CREATE TABLE IF NOT EXISTS schema_meta (version INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS graphs (
    clave      TEXT PRIMARY KEY,
    graph_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

Una tabla JSON-blob (D4) — `clave` es la llave autoritativa que el caller ya decide (invariante
heredada, sin cambio), `graph_json` es el `domain.Graph` serializado con `encoding/json` (mismo
formato que `decodeGraph` ya usaba para los grafos embebidos), `updated_at` es
`time.Now().UTC().Format(time.RFC3339)`. `schemaVersion = 1` (constante Go); la tabla `schema_meta`
tiene como máximo una fila.

### 2.3 Dos handles + PRAGMAs (DSN de `modernc.org/sqlite`)

```go
writerDSN := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate"
writer, _ := sql.Open("sqlite", writerDSN)
writer.SetMaxOpenConns(1)

readerDSN := "file:" + path + "?_pragma=busy_timeout(5000)"
reader, _ := sql.Open("sqlite", readerDSN)   // sin límite de conexiones — pool real
```

`modernc.org/sqlite` (confirmado leyendo `driver.go`/`sqlite.go` del módulo, v1.54.0) soporta
`_pragma=<nombre>(<valor>)` repetible y `_txlock=immediate` (→ toda `BeginTx` en ese handle emite
`BEGIN IMMEDIATE`) como parámetros de DSN — no hace falta ejecutar `PRAGMA ...` a mano tras abrir.
`writer.SetMaxOpenConns(1)` es la pieza que serializa: `database/sql` encola cualquier segundo
`Exec`/`BeginTx` concurrente hasta que el primero suelta la única conexión — nunca hay dos
escritores simultáneos dentro de este proceso, así que `SQLITE_BUSY` de contención intra-proceso no
puede ocurrir (`busy_timeout=5000` queda como defensa extra para el caso cross-proceso, p.ej. un
`arnesia index <dir>` corriendo a la vez que el daemon).

### 2.4 `New` — apertura + `schema_version` (RF-208)

```
1. resolvePath(path)
2. mkdir -p del dir contenedor
3. openHandles(path) → writer, reader
4. ensureSchema(ctx, writer):
     CREATE TABLE IF NOT EXISTS (x2)
     SELECT version FROM schema_meta LIMIT 1
       sql.ErrNoRows  → INSERT version=schemaVersion (fresh)      → nil
       version == schemaVersion                                   → nil
       version != schemaVersion                                   → errSchemaMismatch (sentinel)
       otro error                                                  → error real
5. si errSchemaMismatch: Close(writer+reader) → wipeFile(path, path-wal, path-shm)
                          → reabrir (paso 3) → ensureSchema de nuevo (debe dar nil; si no, error)
6. s := &Store{...}; s.seed(); return s, nil
```

`wipeFile` borra el `.db` **y sus sidecars WAL/SHM** — un detalle que un wipe ingenuo (`os.Remove`
solo del `.db`) podría dejar atrás si el archivo estaba en modo WAL activo. Ningún `ALTER TABLE` en
ningún camino — el único remedio a un mismatch es recrear desde cero (L1 del boundary: desechable).

### 2.5 `Rebuild` (RF-207)

```go
func (s *Store) Rebuild(ctx context.Context) error {
    entries := []ports.ArnesPath{}
    if s.reg != nil { entries = s.reg.List() }
    tx, _ := s.writer.BeginTx(ctx, nil)      // _txlock=immediate del DSN → BEGIN IMMEDIATE real
    defer tx.Rollback()                       // no-op tras Commit
    tx.ExecContext(ctx, `DELETE FROM graphs`) // limpia TODO — ver «por qué borra todo» abajo
    for _, ap := range entries {
        g, err := s.load(ap.Path)             // s.loadGraph: nil-safe
        si err != nil → g = Graph{Nodes: []Box{}, Degradado: true}
        si g.Arnes == nil → g.Arnes = &Arnes{ID: ap.Arnes, Nombre: ap.Arnes}; g.Degradado = true
        tx.ExecContext(ctx, upsertGraphSQL, ap.Arnes, json(g), now)
    }
    return tx.Commit()
}
```

**Por qué borra TODO (incluido lo que `Upsert` insertó fuera del boot)**: RF-207 es explícito —
«ningún dato demo/seed hardcodeado sobrevive en el índice» tras un `Rebuild()`. La consecuencia
directa (documentada en D6 y en `PARIDAD.md`) es que una entrada del Portafolio «observada en el
Mapa» vía `ObservarEnMapa` tampoco sobrevive un `Rebuild()` — es el mismo gap que D6 ya aceptó
explícitamente, no uno nuevo.

**Por qué una sola transacción**: `DELETE`+N `INSERT` atómicos evitan un estado intermedio
observable (un lector que llegue entre el DELETE y el primer INSERT vería la tabla vacía bajo WAL
si no estuvieran en la misma tx — con la tx, los lectores ven el snapshot PRE-Rebuild hasta el
`Commit`, y el snapshot POST-Rebuild después). Reusa el mismo patrón de degradado que
`usecase.NewTurnReindexer` (`internal/usecase/session_reindex.go:44-59`) — no llama a esa función
(usecase no se importa desde un adapter), pero replica su forma en ~8 líneas, aceptable (ladder: no
vale una abstracción compartida para 8 líneas usadas 2 veces en capas distintas).

**Por qué NO llama a `s.Upsert` dentro del loop**: `s.Upsert` toma `s.writer` directamente — con
`SetMaxOpenConns(1)` y la conexión ya prestada a la `tx` abierta, un segundo `Exec` sobre `s.writer`
se auto-bloquearía esperando esa misma conexión (deadlock). El loop usa `tx.ExecContext` con la
misma sentencia `upsertGraphSQL` en vez de reusar el método público.

### 2.6 `Query`/`List`/`Upsert` — sin cambio de contrato

Mismas firmas, mismas validaciones (`clave==""`/`g.Arnes==nil` → error), mismo comportamiento
observable que la versión in-memory — ahora sobre `s.reader.QueryRowContext`/`QueryContext` y
`s.writer.ExecContext` en vez del mapa+mutex. `List` sigue devolviendo `ORDER BY clave` (antes era
un sort manual del slice de claves; ahora lo hace SQL).

### 2.7 `seed()` — se mantiene, con una precisión sobre su ciclo de vida

`New()` sigue sembrando `demo` + `dev-full-cycle` + `content-studio-full` (mismo contenido literal
que la versión in-memory) vía `s.Upsert(context.Background(), ...)`. Esto es deliberado y documentado
en el propio doc-comment de `New`: sirve a callers que construyen un `Store` y **nunca** llaman
`Rebuild` (varios tests del repo, ver §4) — pero en producción (`cmd/arnesia/main.go`) `Rebuild()`
se llama SIEMPRE inmediatamente después de `New()`, y `Rebuild` borra todo antes de repoblar desde
`ArnesRegistry` — así que en el daemon real el seed es puramente transitorio (existe entre el
`New()` y el `Rebuild()` del mismo boot, invisible a cualquier caller externo). Ver §6 para la
consecuencia UX de esto (la sesión ilustrativa de primer-uso que apunta a `dev-full-cycle`).

## 3 · `internal/adapters/watch/watcher.go` — diseño

```go
type Watcher struct { reg ports.ArnesRegistry }   // puede ser nil
func New(reg ports.ArnesRegistry) *Watcher
func (w *Watcher) Watch(ctx context.Context) (<-chan ports.WatchEvent, error)
```

`Watch`:
1. `fsnotify.NewWatcher()` (real, cross-platform; en Linux usa inotify).
2. Para cada `ap := range reg.List()` (si `reg != nil`): `addRecursive(fsw, ap.Path)` — fsnotify
   **no es recursivo** (ni en Linux/inotify), así que cada subdirectorio del árbol de un arnés
   (`.claude/skills/x/`, `.claude/hooks/`, …) necesita su propio `fsw.Add(path)`; `addRecursive`
   hace un `filepath.WalkDir` y suma cada directorio. Un árbol que falla (permiso denegado, etc.) se
   loggea (`slog.Warn`) y se salta — un árbol roto no debe cegar a los demás.
3. Goroutine consumidora: `select` entre `ctx.Done()`, `fsw.Events` y `fsw.Errors`.
   - `fsnotify.Create` sobre un directorio → se suma recursivamente al watch-set EN CALIENTE (si
     no, un `mkdir` seguido de escribir adentro quedaría ciego — ver `addRecursive` en el evento).
   - `opFor(ev.Op)` mapea `Create→WatchCreate`, `Write→WatchWrite`, `Remove|Rename→WatchRemove`,
     `Chmod` (solo) → descartado (metadata sin cambio de contenido).
   - El evento se envía por el channel de salida (`select` contra `ctx.Done()` también, para no
     bloquear para siempre si nadie más consume).
4. `ctx.Done()` → cierra `fsw` y el channel de salida.

## 4 · Reordenamiento de `cmd/arnesia/main.go` (`runServe`)

**Antes:** `idx := index.New()` (sin args) → `idx.Rebuild(ctx)` (re-seed) → … → `arnesReg, err :=
store.NewArnesRegistry(...)` → loop manual `for ap := range arnesReg.List() { loadArnesDir(...) }` →
`watcher := watch.New()` (sin args).

**Ahora:**
```
arnesReg, err := store.NewArnesRegistry(*arnesesPath, *arnesRoot)   // PRIMERO — Rebuild lo necesita
idx, err := index.New(*indexPath, arnesReg, loader.LoadArnes)
defer idx.Close()
idx.Rebuild(ctx)                                                    // repuebla TODO desde arnesReg
loadArnesDir := func(id, path string) error { ... }                 // SIGUE viva — la usa PUT /api/arneses/{id}
...
watcher := watch.New(arnesReg)
events, err := watcher.Watch(ctx)
...
go func() {
    for ev := range events {
        ap, ok := ownerOf(arnesReg, ev.Path)
        if !ok { continue }
        turnReindex(ctx, ap.Arnes, ap.Path)   // el MISMO Reindexer que ya usa el chat (session_reindex.go)
    }
}()
```

`ownerOf(reg, path)` — nuevo helper privado en `main.go` (composition root, mismo lugar que
`arnesLoaderFunc`/`brokerPublisher`): recorre `reg.List()` buscando la entrada cuyo `Path` sea
`path` o del cual `path` cuelgue (`strings.HasPrefix(path, ap.Path+separador)`), evitando falsos
positivos tipo `/a/arnes-2` bajo `/a/arnes`.

`turnReindex` (la variable `usecase.NewTurnReindexer(idx, loader.LoadArnes, brokerPublisher{broker})`
que YA existía para el reindex tras cada turno de chat) se **reusa literalmente** para el consumidor
del watcher — no se construye una segunda instancia. Es la mecanización directa de D5: «reusar la
forma existente es la opción más lazy que sigue siendo correcta».

`loadArnesDir` sigue existiendo tal cual (usada por `PUT /api/arneses/{id}`, el botón «Cargar» del
FE) — solo se retira el loop de boot que la invocaba manualmente, porque `Rebuild()` ahora hace ese
trabajo internamente.

`runIndex` (subcomando CLI `arnesia index` sin dir) cambia de comportamiento: ya no puede «re-seedear
el índice» porque `Rebuild()` ahora significa «reconstruir desde ArnesRegistry», y este subcomando
no tiene uno que ofrecer. Pasa a ser un smoke test: abre un `.db` descartable en un dir temporal
(`os.MkdirTemp`, borrado al salir) solo para confirmar que `New()` (y por ende el seed embebido)
decodifica sin error — nunca llama `Rebuild`.

## 5 · Capabilities (RF-212)

| Capability | Antes | Después | Pointers nuevos/retirados |
|---|---|---|---|
| CAP-21 `indice-de-arneses-en-memoria` | `vivo` | `vivo` (sin cambio funcional) | sin cambio — `Store`/`Store.Upsert`/`Store.Query` siguen resolviendo tal cual |
| CAP-22 `reconstruccion-del-indice` | `parcial`, `valida:` con placeholder `~/.claude` | `vivo` | `+Store.Rebuild, +New`, `-Rebuild` (bare, ahora `Store.Rebuild`) · `valida:` reemplazado por los 6 tests reales de RF-207/208 |
| CAP-23 `observar-cambios-del-corpus` | `stub`, `pointers: [Watch]`, `valida: []` | `parcial` (gap honesto: no re-observa en caliente) | `+Watcher, +New, +Watcher.Watch` · `valida:` con los 4 tests del watcher |

`status` de CAP-22 sube a `vivo` (no `parcial`) porque ya no hay gap entre lo que el nombre promete
y el código: `Rebuild()` reconstruye de verdad. CAP-23 sube solo a `parcial` (no `vivo`) porque el
gap de RF-210 (arnés nuevo en caliente) sigue abierto — ver §6.

## 6 · Deuda honesta / gaps documentados (no maquillados)

1. **Arnés registrado con el daemon ya corriendo no se suma al watch-set** (RF-210, sub-escenario
   3 del propio spec, marcado ahí como aceptablemente diferible). El watcher solo observa lo que
   `reg.List()` devolvía al momento de `Watch(ctx)` (boot). Reiniciar el daemon vuelve a alinear
   `Rebuild()`+`Watch()` con el registro actual. No hay caller hoy que dependa de esto (D1: el
   Portafolio es la vía explícita para «arnés nuevo en caliente», vía `ObservarEnMapa`).
2. **Debounce de ráfagas fsnotify: no implementado.** Un `git checkout`/edición masiva dentro del
   árbol de un arnés dispara un evento por archivo tocado → un `LoadArnes`+`Upsert`+`Publish` por
   cada uno. `decisiones.md` D1 ya establece que el corpus (árbol de arneses) es chico y releerlo
   completo es barato (microsegundos) — no hay caller real que sufra por esto hoy. `ponytail`: si
   algún día un arnés con cientos de archivos hace de esto un problema medible, el upgrade es un
   debounce por-arnés con un `time.Timer` — no construido porque nadie lo necesita todavía.
3. **La sesión ilustrativa de primer-uso pierde su demo tras el primer `Rebuild()`.**
   `internal/usecase/session_service.go:seedSessions()` arma una sesión de ejemplo apuntando al
   arnés `dev-full-cycle` (comentario propio: «the real indexed dogfood arnés») asumiendo que ese
   id siempre está en el índice. RF-207 exige literalmente que ningún seed sobreviva a `Rebuild()`
   — y `Rebuild()` corre siempre al boot — así que en una instalación fresca (`ArnesRegistry`
   vacío, sin el arnés `dev-full-cycle` registrado) esa sesión ilustrativa ahora apunta a un id que
   el Mapa no encuentra (404 honesto vía `MapService.Graph`, no un crash). **Este hallazgo es mío,
   no del spec** — ni `hallazgos.md` ni `decisiones.md` examinaron `session_service.go`. No lo
   arreglo acá: es una decisión de UX de primer-uso (¿la sesión ilustrativa debería auto-registrar
   `dogfood/dev-full-cycle` en `ArnesRegistry`? ¿debería el FE tolerar un Mapa vacío en esa sesión
   demo?) fuera del alcance backend de este paquete — queda anotado en `PARIDAD.md` para que la
   Fase 4 (revisión humana) decida si amerita un ticket propio.
4. **`golangci-lint run ./...` completo no estaba en la lista de comandos de la consigna** (solo
   `golangci-lint fmt --diff`), pero CI sí lo corre (`.github/workflows/ci.yml`, step
   `golangci-lint-action`). Se corrió igual y se llevó a 0 issues — incluidas 2 líneas de
   `internal/adapters/loader/loader_test.go` que el bump de `go.mod` (1.23.0→1.25.0, efecto
   colateral de `go get modernc.org/sqlite@latest`) volvió señalables (`t.Chdir()` disponible desde
   go1.24). Root-cause fix, no symptom patch — ver tabla §0.

## 7 · Lo que NO se construyó (por decisión ya firmada, no reabierta acá)

- Cursores de reanudación `{path,inode,size,offset}` — D1.
- Bulk-load `synchronous=OFF`+`journal_mode=MEMORY` (T6 del plan de `INDEX.md`) — nunca promovido a
  RF por `spec.md`; mismo razonamiento de escala que D1 (corpus chico, sin caller que lo necesite).
- Tablas normalizadas por nodo/edge — D4.
- Rename del boundary `indice-desechable-jsonl-es-verdad` — D3.
- Que `Rebuild()` re-observe automáticamente entradas del Portafolio — D6 (consecuencia aceptada).
