---
regla: indice-desechable-jsonl-es-verdad
version: 1.0
updated: 2026-07-05
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
  # el ban CGO/DuckDB (check sin-cgo) — depguard aún no lo cubre:
  - fitness/arch_test.go:TestNoDuckDBOrCGOStore
severity: high
---

# JSONL es la fuente de verdad; SQLite es un índice reconstruible

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

- **El índice (`modernc.org/sqlite`, WAL) nunca es la fuente de verdad.** Todo lo que vive en
  SQLite se puede regenerar desde JSONL + el event store del stream (ver
  [`conductor-no-parsea-jsonl.md`](./conductor-no-parsea-jsonl.md)). ⇐ L1: proyección.
- **No se migra el índice: se reconstruye.** Una `schema_version` en una meta table; en mismatch →
  borrar el archivo y re-indexar. Migraciones incrementales = prohibidas para el índice (son deuda
  que la reconstrucción hace innecesaria). ⇐ L1: desechable.
- **Concurrencia:** 2 handles `*sql.DB` — writer con `SetMaxOpenConns(1)` (serializa escrituras) +
  reader pooled; `journal_mode=WAL`, `synchronous=NORMAL`, `busy_timeout=5000`, `BEGIN IMMEDIATE`
  en txns de escritura. ⇐ L1: WAL.
- **Rebuild rápido:** cursores por archivo (`{path,inode,size,offset}`) persistidos → reanuda
  incremental tras reinicio; bulk-load con `synchronous=OFF`+`journal_mode=MEMORY`, luego WAL.
- **DuckDB descartado** (es CGO → mata el binario estático cross-compile). Si años de telemetría lo
  exigen: DuckDB out-of-process sobre Parquet, nunca el core.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| index-reconstruible | borrar el `.db` y re-indexar produce el mismo estado consultable (test) | error | «dato solo-en-índice: se pierde si el índice muere» | arch_test.go:TestIndexRebuildsFromJSONL |
| sin-migracion-incremental | no hay migraciones incrementales del índice; hay `schema_version`+rebuild | warn | «migración de un store que debería ser desechable» | arch_test.go:TestSchemaVersionTriggersRebuild |
| writer-serializado | el handle de escritura usa `SetMaxOpenConns(1)`; WAL configurado | warn | «múltiples writers → SQLITE_BUSY» | arch_test.go |
| sin-cgo | el índice usa `modernc.org/sqlite` (no CGO); no hay import de DuckDB en el core | error | «dependencia CGO rompe el binario único» | arch_test.go:TestNoDuckDBOrCGOStore (depguard aún no lo cubre) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = local-first + índice desechable (CQRS-lite).
  L2: SQLite modernc como proyección reconstruible; no-migrar-reconstruir; 2 handles WAL; DuckDB
  descartado por CGO. 4 checks.
