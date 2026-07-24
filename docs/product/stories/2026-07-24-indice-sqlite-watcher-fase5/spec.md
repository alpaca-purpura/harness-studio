# spec — Índice SQLite real + watcher fsnotify (fase 5 del stack)

> Paquete `2026-07-24-indice-sqlite-watcher-fase5` · base: `hallazgos.md` (investigación agente 1,
> T0 resuelto con evidencia) + `decisiones.md` (D1-D6, agente 2). Paquete **100 % backend, sin
> mockup** — el Mapa ya consume `ports.IndexPort`/`ports.WatchPort` sin cambio de contrato; cada
> RF traza a citas `archivo:línea` en vez de `mockup:línea` (disciplina §10, precedente
> `2026-07-13-portafolio-slice0-cimientos/plan-implementacion.md`). Ejecución ordenada por el
> operador vía `/goal` 2026-07-24: autónoma, hasta cerrar el paquete; el gate humano final de
> PARIDAD queda abierto para el operador. **RF arrancan en RF-207** (RF-206 = último usado en el
> repo, `2026-07-22-mejorar-arnes-conversando/spec.md`).

## 0 · Lo que este paquete NO decide (ya resuelto, no relitigar)

- **T0 — el corpus real** (`hallazgos.md` §2): el índice guarda `domain.Graph` — la composición
  estructural de un arnés (nodos/edges de `.claude/`: skills/hooks/rules) — cargada por
  `loader.LoadArnes(dir)`. **NO** es un parser de líneas JSONL de conversación; ese corpus
  (`~/.claude/projects/*.jsonl`) es un adapter separado y ya funcionando
  (`internal/adapters/history/reader.go`, boundary `conductor-no-parsea-jsonl.md`) que ningún
  caller conecta a `IndexPort`/`WatchPort`.
- **D1-D6** (`decisiones.md`): cursores de reanudación fuera de alcance · limpieza de CAP-22 ·
  nombre del boundary se mantiene · schema = 1 tabla JSON-blob · watcher observa
  `ArnesRegistry`, reindex incremental (no `Rebuild()` completo por evento) · `Rebuild()` fuente
  = `ArnesRegistry`, no el Portafolio.

## 1 · Flujo de información — HOY vs. OBJETIVO

### HOY (confirmado, `hallazgos.md` §1)

```
boot (main.go:118-121)
  idx := index.New()          → seed() hardcodeado (demo + 2 grafos embebidos dogfood)
  idx.Rebuild(ctx)             → vuelve a llamar seed() — NO lee disco, NO usa ArnesRegistry

boot (main.go:148-154)
  for ap := range arnesReg.List() {          // arneses.json, CAP-24
    g, _ := loader.LoadArnes(ap.Path)
    idx.Upsert(ctx, ap.Arnes, g)             // el loop vive en main.go, NO dentro de Rebuild()
  }

watcher := watch.New()                       // stub — Watch(ctx) devuelve channel que jamás emite
events, _ := watcher.Watch(ctx)
go func() {
  for range events { /* TODO(fase 5): cuerpo vacío */ }   // main.go:300-305, código muerto
}()

turno de chat → SessionService.consume → NewTurnReindexer (session_reindex.go:57)
  g, _ := loader.LoadArnes(cwd)
  idx.Upsert(ctx, arnesID, g)                → publica event:map {harness_id}   [VIVO, RF-184/186]

FE "Observar en Mapa" → PortafolioService.ObservarEnMapa (portafolio.go:316)
  g, _ := s.cargar.Load(installPath)
  s.indice.Upsert(ctx, clave, g)             [VIVO, per-instalación, opt-in]

daemon cae → índice (map in-memory) se pierde ENTERO; próximo boot repuebla vía el loop de
  arriba (arnesReg) — el árbol en disco sobrevive, el índice no necesita sobrevivir. Cero
  persistencia SQLite hoy: `internal/adapters/index/store.go` es únicamente
  `map[string]domain.Graph` + `sync.RWMutex`.
```

### OBJETIVO (este paquete construye)

```
boot (main.go, refactor)
  idx := index.New(store-sqlite, arnesReg, loader.LoadArnes)   // Rebuild ya sabe su fuente
  idx.Rebuild(ctx)
    → abre/crea el .db (WAL, 2 handles: writer SetMaxOpenConns(1) + reader pooled)
    → si schema_version del .db ≠ la del binario: borra el .db entero, empieza de cero (RF-208)
    → por cada ap := range arnesReg.List():
        g, err := loader.LoadArnes(ap.Path)
        err != nil → Upsert(ap.Arnes, gradado Degradado:true)   // nunca desaparece en silencio
        err == nil → Upsert(ap.Arnes, g)
    → el loop de main.go:148-154 SE ELIMINA (vive adentro de Rebuild ahora) — RF-207

watcher := watch.NewFsnotify(arnesReg)        // observa el working-dir de cada ap.Path conocido
events, _ := watcher.Watch(ctx)
go func() {
  for ev := range events {
    ap, ok := arnesRegDueño(ev.Path)          // ¿qué entrada de ArnesRegistry es prefijo de ev.Path?
    if !ok { continue }                        // evento huérfano — se ignora, no Upsert
    g, err := loader.LoadArnes(ap.Path)
    err != nil → idx.Upsert(ap.Arnes, gradado)
    err == nil → idx.Upsert(ap.Arnes, g)
    broker.Publish(sse.EventMap, {harness_id: ap.Arnes})        // RF-210
  }
}()

turno de chat / ObservarEnMapa: SIN CAMBIOS — mismo IndexPort, mismo comportamiento (RF-184/186,
  ObservarEnMapa siguen vivos tal cual; ver decisiones.md D6 para por qué el Portafolio queda
  fuera de Rebuild()).

daemon cae → el .db SQLite puede sobrevivir en disco (WAL) O perderse — CUALQUIERA de los dos es
  aceptable (RF-211): Rebuild() al boot siguiente reconstruye el mismo estado consultable desde
  ArnesRegistry + el árbol en disco, que es la única fuente que DEBE sobrevivir. El .db es
  desechable por diseño (boundary indice-desechable-jsonl-es-verdad.md v1.2).
```

## 2 · RF numerados (continúa RF-183..RF-206 ya usados en el repo)

### RF-207 · `Rebuild()` reconstruye desde el árbol de arneses real (ArnesRegistry)

`Rebuild(ctx)` deja de re-sembrar datos demo (`internal/adapters/index/store.go:41-46`) y
reconstruye iterando `ports.ArnesRegistry.List()` (`internal/adapters/store/arnes_registry.go:93-101`,
CAP-24) → `loader.LoadArnes(path)` → `Upsert(clave, g)`, replicando el loop que hoy vive fuera,
en `cmd/arnesia/main.go:148-154` (`decisiones.md` D6).

```gherkin
Característica: Rebuild reconstruye el índice desde el árbol de arneses conocido

  Escenario: N arneses registrados y sanos
    Dado que ArnesRegistry tiene N entradas (id, path) con árboles cargables
    Cuando se llama Rebuild(ctx)
    Entonces el índice contiene exactamente N grafos
    Y cada uno es consultable vía Query(id) con el contenido real de loader.LoadArnes(path)
    Y ningún dato demo/seed hardcodeado sobrevive en el índice

  Escenario: una entrada registrada con árbol roto
    Dado una entrada de ArnesRegistry cuyo path ya no es cargable (manifiesto corrupto o ausente)
    Cuando se llama Rebuild(ctx)
    Entonces esa entrada queda indexada con Degradado:true (mismo patrón que
      NewTurnReindexer, session_reindex.go:44-59), nunca ausente del índice ni con pass fabricado

  Escenario: ArnesRegistry vacío
    Dado que ArnesRegistry no tiene entradas
    Cuando se llama Rebuild(ctx)
    Entonces List(ctx) devuelve un slice vacío, sin error
```

### RF-208 · Mismatch de `schema_version` dispara rebuild completo, nunca migración incremental

```gherkin
Característica: el índice SQLite es desechable, no migrable

  Escenario: el .db existe con una schema_version distinta a la esperada
    Dado un archivo .db en disco con tabla meta schema_version = N-1
    Y el binario actual espera schema_version = N
    Cuando el daemon abre/inicializa el store
    Entonces el archivo .db se borra entero (no hay ALTER TABLE ni migración de filas)
    Y se re-ejecuta Rebuild() completo (RF-207) sobre el .db nuevo
    Y el estado consultable resultante es el mismo que si el .db nunca hubiera existido

  Escenario: el .db no existe (primer boot)
    Dado que no hay archivo .db en la ruta configurada
    Cuando el daemon inicializa el store
    Entonces se crea el .db con la schema_version actual y se corre Rebuild()
```

### RF-209 · El writer serializa escrituras (`SetMaxOpenConns(1)`)

```gherkin
Característica: un solo escritor evita SQLITE_BUSY

  Escenario: el store se abre
    Cuando se inicializa el store SQLite
    Entonces el handle de escritura tiene SetMaxOpenConns(1)
    Y journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000 están configurados
    Y el handle de lectura es un pool separado (no comparte el límite de 1 conexión)

  Escenario: dos Upsert concurrentes
    Dado dos goroutines llamando Upsert al mismo tiempo sobre claves distintas
    Cuando ambas corren concurrentemente
    Entonces ninguna devuelve SQLITE_BUSY/error de lock
    Y el estado final refleja ambos Upserts (serializados por el único writer, no perdidos)
```
Enforcer a nombrar (hoy genérico `arch_test.go`, boundary
`indice-desechable-jsonl-es-verdad.md` fila `writer-serializado`): agente 3 escribe
`TestWriterSerializedSingleConn` (o nombre equivalente) y actualiza `enforced_by` en el frontmatter
del boundary con `fitness/arch_test.go:TestWriterSerializedSingleConn`.

### RF-210 · El watcher fsnotify detecta cambios y dispara reindex incremental (no `Rebuild()` completo)

Resuelve la pregunta 5 de la consigna + `decisiones.md` D5: el watcher observa el working-dir de
cada entrada de `ArnesRegistry`; un evento dispara recarga+Upsert de **solo el arnés dueño del
path**, reusando la forma de `NewTurnReindexer` (`internal/usecase/session_reindex.go:38-68`) —
**reemplaza** el TODO literal de `cmd/arnesia/main.go:300-305`, que si se implementara tal cual
(`idx.Rebuild(ctx)` completo por cada evento) recargaría TODOS los arneses por cada archivo
tocado de UNO solo.

```gherkin
Característica: cambios en disco disparan reindex incremental vía fsnotify

  Escenario: un archivo cambia dentro del árbol de un arnés registrado
    Dado un arnés "a" registrado en ArnesRegistry con path P
    Y el watcher corriendo
    Cuando un archivo bajo P se escribe/crea/borra
    Entonces WatchPort emite un WatchEvent{Path bajo P, Op}
    Y el consumidor resuelve "a" como dueño (P es prefijo de Path)
    Y recarga SOLO "a" (loader.LoadArnes(P)) y hace Upsert(ctx, "a", g)
    Y publica event:map {harness_id:"a"} por el broker SSE
    Y NINGÚN otro arnés registrado se recarga por este evento

  Escenario: un evento no pertenece a ningún arnés registrado
    Dado un WatchEvent cuyo Path no cae bajo ningún ap.Path de ArnesRegistry
    Cuando el consumidor lo procesa
    Entonces se ignora (sin Upsert, sin publish, sin error)

  Escenario: se registra un arnés nuevo con el daemon ya corriendo
    Dado el watcher ya corriendo antes de que se registre un arnés nuevo (sesión nueva)
    Cuando el arnés se registra (POST /api/arneses o creación de sesión)
    Entonces su directorio queda observado sin reiniciar el daemon
    O, si esto no es viable en la primera iteración, el gap queda documentado explícito en
      PARIDAD.md (RF-214) — nunca silenciado como si funcionara
```

### RF-211 · Daemon cae y vuelve — cero pérdida de datos indexables

```gherkin
Característica: el árbol en disco es la única fuente que debe sobrevivir

  Escenario: kill -9 y reinicio
    Dado el daemon corriendo con M arneses indexados (sanos, ninguno degradado)
    Cuando el proceso termina abruptamente y se reinicia
    Entonces Rebuild() (RF-207) repuebla el índice completo desde ArnesRegistry + el árbol en
      disco de cada arnés
    Y el conjunto de arneses consultables tras el reinicio es el mismo que antes de caer
    Y esto es cierto exista o no un .db SQLite previo en disco (borrarlo a mano no pierde nada
      que el árbol + ArnesRegistry no puedan reconstruir)
```

### RF-212 · Capabilities reflejan el estado real post-implementación

`docs/product/capabilities/indice-persistencia/`: `reconstruccion-del-indice.yaml` (CAP-22)
retira el placeholder `~/.claude` de `valida:` (`decisiones.md` D2) y lista los tests reales que
cubren RF-207/208; su `status` sube de `parcial` a `vivo` cuando `Rebuild` deja de ser `seed()`.
`observar-cambios-del-corpus.yaml` (CAP-23) gana `pointers`/`valida` reales del watcher (RF-210)
y sube de `stub` a `vivo`/`parcial` según cobertura real. `indice-de-arneses-en-memoria.yaml`
(CAP-21) — sin cambio funcional esperado (Query/List/Upsert no cambian de firma), pero su
`pointers` deben seguir resolviendo si el archivo se reescribe. R1/R2/R4
(`codigo-traza-a-capability.md`) sin huérfanos nuevos.

### RF-213 · Los 3 checks `deferred` pasan a `pass`

`docs/architecture/fitness/arch_test.go`: `TestIndexRebuildsFromJSONL` (líneas 344-346) y
`TestSchemaVersionTriggersRebuild` (líneas 348-350) dejan de ser `t.Skip(...)` y verifican
RF-207/208 con aserciones reales. `writer-serializado` deja de ser el enforcer genérico
`arch_test.go` (ver RF-209). `go run ./cmd/arnesia conformance --todo` corre estos 3 en `pass`,
sin checks `deferred` nuevos introducidos por este paquete.

### RF-214 · `PARIDAD.md` del paquete

Tabla RF-207..RF-213 → evidencia (test real que pasa + salida de `conformance --todo`). Cualquier
desviación (p.ej. el sub-escenario de RF-210 sobre arneses registrados en caliente, si no resultó
viable) queda documentada explícita, nunca maquillada. Gate 🧑‍⚖️ **ABIERTO** para el operador —
este spec no lo firma ni lo simula.

## 3 · Criterio de aceptación consolidado

- `go build ./... && go test ./... -race` verde.
- `go test ./docs/architecture/fitness/...` verde, incluyendo los 3 tests de RF-213 en pass real
  (no skip).
- `go run ./cmd/arnesia conformance --todo`: pass ≥ el conteo actual + 3 (los 3 que salen de
  `deferred`), fail 0.
- Boundary `indice-desechable-jsonl-es-verdad.md` gana un changelog v1.3: `Rebuild()`/watcher
  reales cableados, `enforced_by` de `writer-serializado` nombrado, sin cambio de `status`
  (pasa a `enforced` solo si los 4 checks del checklist están en pass real — evaluar al cierre).
- `bash scripts/estado.sh --check` verde (cifras regeneradas, nunca tecleadas).

## 4 · Fuera de alcance (hereda `INDEX.md`, sin cambios)

- Capas Desempeño/Proceso del Mapa (dependen del canal OTel, paquete
  `2026-07-24-telemetria-embebida-otel` aparte).
- Capa Tokens del Mapa (paquete propio, no bloqueado por este).
- Historial de chat (`internal/adapters/history/reader.go`) — adapter separado, de solo lectura,
  no alimenta ni depende de este índice.
- Derivación LIVE de capabilities (`vivo ⟺ test corriendo`) — item BACKLOG relacionado pero
  distinto.
- Cursores de reanudación `{path,inode,size,offset}` (D1) — deuda futura condicional a que el
  corpus indexado incluya algún día archivos que crecen por apéndice.
- Rename del boundary `indice-desechable-jsonl-es-verdad` (D3) — se mantiene el nombre.
- Que el índice re-observe automáticamente entradas del Portafolio tras un reinicio (D6,
  consecuencia aceptada) — `ObservarEnMapa` sigue siendo la única vía, sin cambios.
- Tablas SQLite normalizadas por nodo/edge (D4) — 1 tabla JSON-blob alcanza a todo caller real
  hoy.

## 5 · Coherencia con visión (referencia, no relitigar)

`docs/product/vision.md:189-192` («Sensor local-first»): *"Los JSONL de `~/.claude` son la
fuente de verdad (Claude Code ya es el colector); el sensor indexa — daemon caído = cero
pérdida. Índice desechable [...] el índice SQLite es fase 5 futura."* Este paquete ES esa fase 5:
el `.db` SQLite pasa de "futuro" a construido, pero la fuente de verdad para ESTE índice
(distinto del historial de chat que sí lee JSONL vía `history.Reader`) es el árbol de arneses en
disco (T0). RF-211 es la traducción literal de "daemon caído = cero pérdida" a un escenario
Gherkin verificable.
