---
regla: indice-desechable-jsonl-es-verdad
version: 1.3
updated: 2026-07-24
status: proposed
ledger: HS-04
sources:
  - url: https://code.claude.com/docs/en/sessions
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://www.sqlite.org/wal.html
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestIndexRebuildsFromJSONL
  - fitness/arch_test.go:TestSchemaVersionTriggersRebuild
  - fitness/arch_test.go:TestWriterSerializedSingleConn
  # el ban CGO/DuckDB (check sin-cgo) — depguard aún no lo cubre:
  - fitness/arch_test.go:TestNoDuckDBOrCGOStore
severity: high
---

# JSONL es la fuente de verdad; el índice es reconstruible

## L1 · Principio (estándar de industria)

**Local-first con índice desechable (CQRS-lite).** Cuando existe una fuente de verdad durable
(aquí: los transcripts JSONL de `~/.claude`, que Claude Code ya escribe), el store de consulta es
una **proyección reconstruible**, no un segundo original. Si el índice se corrompe, se borra y se
reproyecta desde la verdad. Esto elimina toda una clase de bugs de migración/consistencia: el
índice puede cambiar de schema libremente porque se puede tirar. *(oficial: sessions — el JSONL
es interno y cambia entre versiones; SQLite WAL para el índice)*

## L2 · Realización (este árbol Go+React)

VISION lo firma: «los JSONL de `~/.claude` son la fuente de verdad; índice SQLite desechable;
daemon caído = cero pérdida».

- **El índice nunca es la fuente de verdad — pero desde fase 5 SÍ persiste (v1.3).**
  `internal/adapters/index/store.go` es `modernc.org/sqlite` real (2 handles `*sql.DB`, WAL);
  si el daemon cae, el `.db` puede sobrevivir o perderse — CUALQUIERA de los dos es aceptable,
  porque `Rebuild()` al boot siguiente repuebla el mismo estado consultable desde la fuente. (El
  "store JSON atómico" que sí existe aparte en el árbol — `internal/adapters/store/registry.go` +
  `arnes_registry.go` — persiste OTRA cosa: el registro de sesiones y el mapeo arnés→path, no el
  grafo indexado; no confundir los dos.)
- **La fuente durable real NO es literalmente la conversación JSONL de `~/.claude/projects`**
  (T0 del paquete `2026-07-24-indice-sqlite-watcher-fase5`, corregido v1.2). El índice guarda
  `domain.Graph` — la composición estructural de un arnés (nodos/edges de `.claude/`:
  skills/hooks/rules) — no historial de chat. **Corregido v1.3:** la fuente durable de la que
  `Rebuild()` reconstruye es `ports.ArnesRegistry` (`internal/adapters/store/arnes_registry.go`,
  el registro simple `id→path` — **NO** el Portafolio, que v1.2 nombraba como insumo antes de que
  el spec resolviera esta ambigüedad: una `EntradaPortafolio` puede tener N instalaciones para la
  misma identidad, y `Rebuild()` no puede inventar cuál gana sin una política que nadie pidió; el
  Portafolio sigue alimentando el índice solo vía `ObservarEnMapa`, explícito, per-instalación,
  sin cambios). `ports.ArnesRegistry.List()` → `loader.LoadArnes(path)` (vivo desde HS-11) →
  `Upsert`. La conversación JSONL de sesiones de chat es un corpus SEPARADO, con su propio adapter
  de solo-lectura (`internal/adapters/history/reader.go`, boundary
  [`conductor-no-parsea-jsonl.md`](./conductor-no-parsea-jsonl.md)) que no alimenta este índice.
  El nombre "JSONL" de esta regla es la analogía L1 de industria (Claude Code transcripts como
  ejemplo genérico de "índice desechable sobre una fuente durable"); el corpus L2 concreto de
  ESTE árbol es el árbol de arneses, no transcripts. Todo lo que vive en el índice se puede
  regenerar re-escaneando y re-cargando ese árbol. ⇐ L1: proyección.
- **No se migra el índice: se reconstruye.** Una tabla `schema_meta` con `version`; en mismatch →
  se borra el `.db` entero (más sus sidecars `-wal`/`-shm`) y se re-indexa. Migraciones
  incrementales = prohibidas para el índice (son deuda que la reconstrucción hace innecesaria).
  ⇐ L1: desechable.
- **Concurrencia (v1.3, cableado):** 2 handles `*sql.DB` — writer con `SetMaxOpenConns(1)`
  (serializa escrituras — database/sql encola cualquier segundo Exec/tx hasta que el primero
  suelta la única conexión) + reader pooled sin límite; `journal_mode=WAL`, `synchronous=NORMAL`,
  `busy_timeout=5000`, `_txlock=immediate` (⇒ `BEGIN IMMEDIATE`) en las transacciones de escritura
  del writer. ⇐ L1: WAL.
- **Cursores de reanudación `{path,inode,size,offset}` — decidido FUERA de fase 5 (D1,
  `2026-07-24-indice-sqlite-watcher-fase5/decisiones.md`).** El corpus real (árbol de arneses) es
  chico y no crece por apéndice como un JSONL de sesiones — `Rebuild()` relee cada árbol completo
  en cada corrida, barato a esta escala. Revive como deuda futura si el índice algún día indexa
  algo que sí crece por apéndice. Bulk-load (`synchronous=OFF`+`journal_mode=MEMORY` durante la
  carga masiva) tampoco se construyó — mismo razonamiento de escala, no hay caller real todavía.
- **DuckDB descartado** (es CGO → mata el binario estático cross-compile). Si años de telemetría lo
  exigen: DuckDB out-of-process sobre Parquet, nunca el core.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| index-reconstruible | borrar el `.db` y re-indexar produce el mismo estado consultable (test) | error | «dato solo-en-índice: se pierde si el índice muere» | arch_test.go:TestIndexRebuildsFromJSONL |
| sin-migracion-incremental | no hay migraciones incrementales del índice; hay `schema_version`+rebuild | warn | «migración de un store que debería ser desechable» | arch_test.go:TestSchemaVersionTriggersRebuild |
| writer-serializado | el handle de escritura usa `SetMaxOpenConns(1)`; WAL configurado | warn | «múltiples writers → SQLITE_BUSY» | arch_test.go:TestWriterSerializedSingleConn |
| sin-cgo | el índice usa `modernc.org/sqlite` (no CGO); no hay import de DuckDB en el core | error | «dependencia CGO rompe el binario único» | arch_test.go:TestNoDuckDBOrCGOStore (depguard aún no lo cubre) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = local-first + índice desechable (CQRS-lite).
  L2: SQLite modernc como proyección reconstruible; no-migrar-reconstruir; 2 handles WAL; DuckDB
  descartado por CGO. 4 checks.
- 2026-07-09 · v1.1 · **Sync HS-18 (reorg-docs).** El índice ACTUAL es un **map in-memory + store
  JSON atómico** (reconstruible desde JSONL), no SQLite: `modernc.org/sqlite`+WAL pasa a ser el
  **target de escala de fase 5** (aún no cableado). Se reencuadra el título y el bullet líder; los
  detalles WAL/2-handles quedan como diseño de fase 5 (aditivo, no se borran). Principio (JSONL =
  verdad, índice desechable) intacto; sin cambios de checks ni de status (`proposed`).
- 2026-07-24 · v1.2 · **Corrección de drift, investigación previa a `spec.md` (paquete
  `2026-07-24-indice-sqlite-watcher-fase5`).** Dos afirmaciones de v1.1 no resistieron la
  verificación contra el código real: (1) el "store JSON atómico" del índice **no existe** —
  `internal/adapters/index/store.go` es únicamente `map[string]domain.Graph` + `sync.RWMutex`,
  cero escritura a disco (el JSON atómico real del árbol, `internal/adapters/store/`, persiste
  sesiones y el registro arnés→path, no el grafo — dos cosas distintas que el texto anterior
  conflaba); (2) resuelto **T0**: la fuente durable de la que el índice se reconstruye NO es
  literalmente la conversación JSONL de `~/.claude/projects` — es el árbol de arneses en disco
  vía el Portafolio (`ports.PortafolioStore.Listar()`/`PortafolioScanner.Escanear()` +
  `ports.ArnesLoader.Load()`), confirmado siguiendo el flujo real de
  `cmd/arnesia/main.go`→`internal/usecase/portafolio.go:316` (`ObservarEnMapa`) y
  `internal/usecase/session_reindex.go:57` (`NewTurnReindexer`, `load(cwd)` = `loader.LoadArnes`
  sobre el cwd de la sesión, nunca un parser de líneas JSONL). Sin cambios de checks/enforcers ni
  de `status` (`proposed`); el gap de construcción (SQLite real + fsnotify real) sigue abierto,
  fase 3 de ese paquete.
- 2026-07-24 · v1.3 · **Fase 3 del paquete construida: el `.db` real queda cableado.**
  `internal/adapters/index/store.go` reescrito sobre `modernc.org/sqlite` (2 handles, WAL,
  `schema_meta`, wipe-and-recreate en mismatch — RF-207/208); `internal/adapters/watch/watcher.go`
  reescrito sobre `fsnotify` real, observando el árbol de cada entrada de `ArnesRegistry` (RF-210);
  `cmd/arnesia/main.go` rewireado: el loop de boot que poblaba el índice ahora vive DENTRO de
  `Rebuild()`, y el consumidor del canal de eventos (antes un TODO vacío) reindexa incremental
  SOLO el arnés dueño del path que cambió (reusa `usecase.NewTurnReindexer`, el mismo patrón que
  ya vivía en `session_reindex.go`), nunca un `Rebuild()` completo por evento. **Corrección de
  fondo respecto a v1.2:** la fuente que `Rebuild()` lee es `ports.ArnesRegistry`, no el
  Portafolio (decisión D6 del spec — ver bullet L2 arriba). `enforced_by` gana
  `TestWriterSerializedSingleConn` (chequeo estructural: `SetMaxOpenConns(1)`+WAL+busy_timeout en
  el código fuente; la prueba comportamental — Upserts concurrentes sin `SQLITE_BUSY` — vive junto
  al código en `internal/adapters/index/store_test.go:TestWriterSerializedConcurrentUpsertsSucceed`)
  — antes enforcer genérico `arch_test.go` sin test nombrado, ahora nombrado. Los 4 checks del
  checklist corren en `pass` real (`index-reconstruible`, `sin-migracion-incremental`,
  `writer-serializado`, `sin-cgo`) — con eso, esta regla calificaría para graduar de `proposed` a
  `enforced` (mismo criterio que graduó `codigo-traza-a-capability.md` v1.1), pero esa promoción de
  `status` es criterio editorial del operador, no se fuerza acá. Decisiones D1 (cursores de
  reanudación) y D3 (nombre del boundary, no se renombra) quedan documentadas in extenso arriba;
  deuda honesta que sigue abierta: un arnés registrado con el daemon ya corriendo no se suma al
  watch-set en caliente (el watcher solo observa lo que `ArnesRegistry` tenía al arrancar) — ver
  `PARIDAD.md` del paquete para el detalle completo.
