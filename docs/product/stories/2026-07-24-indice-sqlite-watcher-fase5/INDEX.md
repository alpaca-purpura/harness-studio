# Índice SQLite real + watcher fsnotify — fase 5 del stack (paquete de trabajo)

> Origen: barrido de deuda viva HS-27 (2026-07-24) — el operador pidió "explicame más" sobre el
> ítem BACKLOG "índice SQLite (CAP-21) + watcher fsnotify (CAP-23) + seed→JSONL (CAP-22) — sale
> la fase-5 del stack · bloqueo", y tras la explicación pidió consolidar todo lo investigado en
> un paquete para construir en sesión nueva. **CERO código construido en este cierre** — es
> investigación + plan de ataque, no implementación.

## Qué es hoy (real, con cita)

- **Índice** (`internal/adapters/index/store.go`): `Store` = `map[string]domain.Graph` +
  `sync.RWMutex`. `New()` llama `seed()`, que carga un arnés demo hardcodeado + 2 grafos
  embebidos (`dogfood.DevFullCycleJSON`/`ContentStudioFullJSON`). `Rebuild(ctx)` (línea 43-46)
  **no lee nada del disco real** — vuelve a llamar `seed()`. `Upsert`/`Query`/`List` sí son reales
  y con tests (`TestListPortfolio`, `TestUpsertIndexaBajoLaClaveNoElArnesID`, etc.).
- **Watcher** (`internal/adapters/watch/watcher.go`): stub literal, 27 líneas. `Watch(ctx)`
  devuelve un channel que nunca emite y se cierra con el ctx. El propio comentario del archivo:
  `// TODO(fase 5): fsnotify — watch ~/.claude/projects for JSONL changes...`.
- **Capabilities:** CAP-21 `vivo` (el map in-memory funciona) · CAP-22 `parcial` (`Rebuild`
  existe pero es el mismo seed) · CAP-23 `stub` (watcher no hace nada) —
  `docs/product/capabilities/indice-persistencia/*.yaml`.
- **Quién llama qué hoy** (`cmd/arnesia/main.go:119,133,432`): `idx.Rebuild(ctx)` al boot (re-seed
  nomás) + `idx.Upsert(ctx, id, g)` cuando el loader real carga un directorio (flujo "Cargar" del
  Portafolio, ya real desde HS-11). El indexado que SÍ importa hoy pasa por `Upsert` invocado
  desde casos de uso reales (`portafolio.go:316`, `session_reindex.go:57` tras cada turno de
  chat), nunca por `Rebuild`.

## Diseño ya firmado (no hay que inventarlo, solo construirlo)

`docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md` (v1.1, `status: proposed`,
firmado como boundary L1/L2 desde HS-04):

- **`modernc.org/sqlite`** — SQLite pure-Go, **sin CGO** (CGO rompe el binario estático
  cross-compile; por eso DuckDB queda descartado del core).
- **2 handles `*sql.DB`:** writer con `SetMaxOpenConns(1)` (serializa escritura) + reader pooled.
  `journal_mode=WAL` + `synchronous=NORMAL` + `busy_timeout=5000` + `BEGIN IMMEDIATE` en
  transacciones de escritura.
- **No hay migraciones incrementales.** Una tabla meta con `schema_version`; en mismatch → se
  borra el `.db` entero y se re-indexa desde la fuente. El índice es desechable por diseño
  (CQRS-lite): si se corrompe, se tira y se reproyecta.
- **Rebuild rápido:** cursores `{path,inode,size,offset}` persistidos → reanuda incremental tras
  reinicio. Bulk-load con `synchronous=OFF`+`journal_mode=MEMORY`, después vuelve a WAL.

## Los 3 checks `deferred` que esto desbloquea (evidencia exacta)

`docs/architecture/fitness/arch_test.go`:

- `TestIndexRebuildsFromJSONL` (líneas 344-346) — `t.Skip("TODO(fase 5): borrar el .db y
  re-indexar produce el mismo estado consultable")`.
- `TestSchemaVersionTriggersRebuild` (líneas 348-350) — `t.Skip("TODO(fase 5): mismatch de
  schema_version borra y reconstruye; no hay migración incremental")`.
- `writer-serializado` — enforcer genérico `arch_test.go` sin test nombrado (mismo patrón que
  causaba `deferred` en `ctx-derivado-etiquetado`, ver ese paquete).

`TestNoDuckDBOrCGOStore` (línea 209) **ya existe y ya pasa** (solo verifica ausencia de imports
`duckdb`/`mattn/go-sqlite3`) — no es parte de esta deuda, no tocar.

## ⚠ Pregunta abierta descubierta investigando (resolver ANTES de codear, no asumir)

El wording "JSONL corpus" del TODO en `store.go:4-5` y del propio boundary doc **no aplica
literalmente a este índice**. El índice guarda `domain.Graph` (composición estructural de un
arnés: nodos/edges de `.claude/` — skills/hooks/rules), no conversaciones. La fuente durable real
para "qué arneses existen y qué contienen" ya es el **Portafolio**, construido en HS-22/23:

- `ports.PortafolioStore.Listar()` (`internal/ports/portafolio.go:33-41`) — registro de entradas
  conocidas (canónico + instalaciones).
- `ports.PortafolioScanner.Escanear(ctx, root)` — walker físico read-only del disco.
- `ports.ArnesLoader.Load(dir)` — wrapper sobre `loader.LoadArnes` (el loader real que ya arma
  `domain.Graph` desde `.claude/`, ya vivo desde HS-11 "tres puentes").

Un `Rebuild()` real probablemente es: `Listar()`/`Escanear()` → por cada entrada, `Load(dir)` →
`Upsert(clave, g)` en el nuevo store SQLite. **NO** es un parser de JSONL de sesiones — eso ya
tiene su propio adapter separado y funcionando (`internal/adapters/history/reader.go`, boundary
`conductor-no-parsea-jsonl.md`, usado para historial de chat, no para el índice estructural).

La terminología "JSONL" del boundary/TODO es la analogía L1 genérica (Claude Code transcripts
como ejemplo de industria de "índice desechable"), heredada sin ajustar al comentario L2
concreto de `store.go`. **Confirmar esta lectura al arrancar el spec** — si está mal, todo el
plan de abajo apunta al corpus equivocado.

## Puertos existentes a reusar (cero cambio de firma en usecase)

`ports.IndexPort` (`internal/ports/index.go`) ya es la interfaz que consumen `RunService`,
`MapService`, `Reindexer` (session_reindex.go), `PortafolioService`, `FuenteService` — el nuevo
store SQLite implementa la MISMA interfaz (`Rebuild`/`Query`/`List`/`Upsert`), ningún caller
cambia. Igual con `ports.WatchPort` (`internal/ports/watch.go`, tipos `WatchEvent`/`WatchOp` ya
definidos, listos para un adapter fsnotify real).

## Plan de ataque sugerido (tickets, orden sugerido)

1. **T0 — resolver la pregunta abierta de arriba.** Sin esto, T1-T4 pueden apuntar mal.
2. **T1 — schema + 2 handles.** Tabla(s) para `domain.Graph` (decidir: 1 tabla JSON-blob por
   clave vs. tablas normalizadas por nodo/edge — empezar por blob, es lo mínimo que satisface
   "reconstruible", normalizar después si hace falta query granular) + tabla meta
   `schema_version` + PRAGMAs (WAL/NORMAL/busy_timeout=5000) + `SetMaxOpenConns(1)` en el writer.
3. **T2 — `Rebuild` real** según T0: recorre harnesses conocidos → `Load` → `Upsert`. Mismatch de
   `schema_version` → borrar `.db` → re-rebuild completo (nunca migración).
4. **T3 — cursores de reanudación** `{path,inode,size,offset}` — **condicional a T0**: solo
   aplica si Rebuild también necesita reanudar una lectura incremental de archivos grandes; si la
   fuente es solo el árbol de harnesses (finito, chico), este ticket puede no ser necesario —
   decidir en el spec, no construir "porque el boundary lo menciona".
5. **T4 — watcher fsnotify real.** Directorio(s) a observar = el mismo que resuelva T0 (¿árbol de
   harnesses `.claude/plugins/`+checkouts del Portafolio? ¿o también algo de sesiones?). Emite
   `ports.WatchEvent` → dispara reindex incremental (precedente ya vivo:
   `usecase/session_reindex.go`, "reindex tras evento" ya funciona para el flujo de chat).
6. **T5 — activar los 2 `t.Skip`** (`TestIndexRebuildsFromJSONL`,
   `TestSchemaVersionTriggersRebuild`) con implementación real + nombrar el enforcer de
   `writer-serializado` (hoy genérico). Correr `go run ./cmd/arnesia conformance --todo` y
   confirmar que los 3 checks pasan de `deferred` a `pass`.
7. **T6 — bulk-load rápido**: `synchronous=OFF`+`journal_mode=MEMORY` durante la carga masiva
   inicial, vuelve a WAL al terminar.
8. **T7 — PARIDAD** del paquete (gate humano 🧑‍⚖️, disciplina §10).

## Fuera de alcance explícito

- Capas **Desempeño**/**Proceso** del Mapa — dependen del canal OTel (`hooks-desde-otel`),
  confirmado fuera de esta arquitectura (ver `stories/2026-07-24-telemetria-embebida-otel/`).
- Capa **Tokens** del Mapa — paquete propio ya abierto, no bloqueado por este.
- Historial de chat (`history.Reader`) — adapter separado, ya funciona por lectura directa
  on-demand, no depende de este índice ni lo alimenta.
- **derivación LIVE de capabilities** (`vivo ⟺ test corriendo`) — item BACKLOG relacionado pero
  distinto; puede reusar la infraestructura de este paquete después, no es parte de él.

## Estado

- [x] **Fase 1 — investigación** (agente 1, 2026-07-24): `hallazgos.md` — T0 resuelto con
  evidencia dura (corpus = árbol de arneses vía Portafolio, no JSONL de conversación), 2 drifts
  de comentarios corregidos (`ports/index.go`, `ports/watch.go`, `map_service.go`), boundary
  `indice-desechable-jsonl-es-verdad.md` corregido a v1.2.
- [x] **Fase 2 — spec funcional** (agente 2, 2026-07-24): `decisiones.md` (D1-D6) + `spec.md`
  (RF-207..RF-214, Gherkin por RF, flujo HOY vs. OBJETIVO). Decisiones cerradas: cursores de
  reanudación fuera de alcance · CAP-22 se limpia al tocar el yaml · el nombre del boundary NO
  se renombra · schema = 1 tabla JSON-blob · `Rebuild()` reconstruye desde `ArnesRegistry` (no
  el Portafolio) · watcher observa `ArnesRegistry` y dispara reindex incremental (no `Rebuild()`
  completo por evento).
- [x] **Fase 3 — implementación** (agente 3, 2026-07-24): `arquitectura.md` (diseño al detalle) +
  código real + `PARIDAD.md` (evidencia). Store SQLite real (`modernc.org/sqlite`, 2 handles WAL,
  `schema_meta`, wipe-and-recreate) · `Rebuild()` real sobre `ArnesRegistry` (RF-207) · watcher
  `fsnotify` real + consumidor en `main.go` (reindex incremental por-arnés, RF-210, reemplaza el
  TODO que vivía en `main.go:300-305`) · `TestIndexRebuildsFromJSONL` +
  `TestSchemaVersionTriggersRebuild` activados (RF-208) · enforcer `writer-serializado` nombrado
  (`TestWriterSerializedSingleConn`, RF-209) · CAP-21/22/23 actualizados (RF-212). Los 3 checks que
  eran `deferred` pasan a `pass` real: `conformance --todo` `pass 54→57 · deferred 212→209 · fail
  0` (exacto +3/−3). Los 5 comandos de CI + `golangci-lint run ./...` (no listado en la consigna
  pero corre en CI) + `estado.sh --check` — todos verdes, `estado.sh` regenerado y committeable.
  2 desviaciones honestas documentadas en `PARIDAD.md` (gap RF-210 "arnés nuevo en caliente" — ya
  anticipado por el propio spec como diferible — y un hallazgo propio sobre la sesión ilustrativa
  de primer-uso, `session_service.go:seedSessions()`, fuera del alcance backend de este paquete).
- [x] **Fase 4 — revisión visual/UX** (agente 4, 2026-07-24): `revision-visual.md` — los 5
  escenarios de la consigna, app real levantada y navegada con `claude-in-chrome`, entorno REAL
  del operador (`dev-full-cycle`+`vitalia` en `~/.arnesia/arneses.json`, no editado). **Ningún bug
  encontrado, cero fixes de código.** Boot con datos reales (no demo, `GET /api/harnesses` =
  exactamente los 2 arneses) · consola limpia en 3 pasadas · watcher fsnotify confirmado en vivo de
  punta a punta (alta Y baja de un nodo, SSE `event: map` + refetch de la FE sin recargar,
  `dogfood/` quedó limpio) · `kill -9`+reinicio del daemon con cero pérdida · gap de
  `seedSessions()` confirmado no-reproducible en este entorno (como se esperaba). Checklist técnico
  de `PARIDAD.md` actualizado — **falta SOLO la firma 🧑‍⚖️ del operador** (los 3 ítems de criterio
  editorial: gap RF-210, gap `seedSessions()`, graduar el boundary a `enforced`).

## Retomar aquí

**Fase 4 CERRADA (revisión visual/UX, sin bugs).** El paquete queda COMPLETO salvo la firma
humana. Una sesión nueva que retome esto arranca leyendo: `PARIDAD.md` (checklist al final, 2 de 5
ítems ya tildados por agentes, 3 pendientes de criterio editorial del operador) →
`revision-visual.md` (evidencia de la revisión visual) → (si hace falta más contexto)
`arquitectura.md`/`decisiones.md`/`hallazgos.md`. El siguiente paso concreto es que el operador
revise los 3 ítems de criterio editorial del checklist y firme 🧑‍⚖️ — no queda trabajo de código
ni de agente pendiente en este paquete.
