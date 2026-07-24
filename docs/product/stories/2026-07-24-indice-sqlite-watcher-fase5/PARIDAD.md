# PARIDAD — índice SQLite real + watcher fsnotify (fase 5)

> RF-207..RF-214 de `spec.md` → evidencia técnica real (test que pasa + output de `conformance
> --todo`). Escrito por el agente 3 (implementación), 2026-07-24. **Gate 🧑‍⚖️ de operador: ABIERTO
> — no simulado, no fabricado.** Esta tabla es el insumo para que el operador lo firme en Fase 4.

## Tabla RF → evidencia

| RF | Qué pedía | Evidencia real |
|---|---|---|
| **RF-207** — `Rebuild()` reconstruye desde `ArnesRegistry` | 3 escenarios Gherkin: N sanos · 1 roto degradado · registro vacío | `internal/adapters/index/store_test.go`: `TestRebuildDesdeArnesRegistry` (N sanos, ningún seed sobrevive), `TestRebuildEntradaDegradada` (loader falla → `Degradado:true`, presente igual), `TestRebuildRegistroVacio` (`List()` vacío, sin error). Más `docs/architecture/fitness/arch_test.go:TestIndexRebuildsFromJSONL` (mismo mecanismo contra el dogfood REAL en disco). Todos `PASS`. |
| **RF-208** — mismatch de `schema_version` → wipe completo, nunca migración | 2 escenarios: `.db` viejo · `.db` inexistente | `internal/adapters/index/store_test.go:TestSchemaVersionMismatchWipesFile` (fila `old-junk` sembrada a mano en un `.db` con `version=-1`; tras `New()` esa fila NO existe — `ErrNotFound` — y `Rebuild()` funciona sobre el `.db` recreado). Más `docs/architecture/fitness/arch_test.go:TestSchemaVersionTriggersRebuild` (mismo patrón contra el dogfood real). Ambos `PASS`. |
| **RF-209** — writer serializado (`SetMaxOpenConns(1)`) | 2 escenarios: config estructural · Upserts concurrentes | Estructural: `docs/architecture/fitness/arch_test.go:TestWriterSerializedSingleConn` (grep del código fuente: `writer.SetMaxOpenConns(1)`, `journal_mode(WAL)`, `busy_timeout(5000)` — los 3 presentes). Comportamental: `internal/adapters/index/store_test.go:TestWriterSerializedConcurrentUpsertsSucceed` (16 goroutines, `Upsert` concurrente, `-race` limpio, 0 errores, las 16 claves quedan consultables). Ambos `PASS`; `go test ./... -race` global también limpio. |
| **RF-210** — watcher fsnotify + reindex incremental (no `Rebuild()` completo por evento) | 2 escenarios base + 1 sub-escenario condicional (arnés nuevo en caliente) | `internal/adapters/watch/watcher_test.go`: `TestWatchEmitsEventUnderRegisteredPath`, `TestWatchObservesNestedSubdirCreatedAfterStart`, `TestWatchIgnoresUnregisteredTree`, `TestWatchClosesWhenContextDone` — todos `PASS`. El consumidor en `cmd/arnesia/main.go` (resolución de dueño + reindex de SOLO ese arnés + publish scoped) reusa `usecase.NewTurnReindexer`, ya probado en producción por el flujo de chat (`internal/usecase/session_reindex.go`); su pieza propia (`ownerOf`) tiene test dedicado: `cmd/arnesia/main_test.go:TestOwnerOf` (5 casos, incluye el caso "vecino con prefijo similar no colisiona"). **Sub-escenario "arnés nuevo en caliente": NO implementado** — ver «Desviaciones» abajo, tal como el propio RF-210 lo permite explícitamente («si esto no es viable en la primera iteración, el gap queda documentado explícito»). |
| **RF-211** — daemon cae y vuelve, cero pérdida | 1 escenario: kill -9 y reinicio | Sin un test literal de `kill -9` (no aplicable a un test unitario), la propiedad se reduce exactamente a RF-207/208: `Rebuild()` reconstruye el mismo estado consultable exista o no un `.db` previo — que es lo que `TestIndexRebuildsFromJSONL` prueba explícito (borra el `.db`, reconstruye, compara con `reflect.DeepEqual` contra el estado pre-borrado). `PASS`. |
| **RF-212** — capabilities reflejan el estado real | CAP-21/22/23 actualizados, R1/R2/R4 sin huérfanos | 3 yaml editados (`docs/product/capabilities/indice-persistencia/`) — detalle en `arquitectura.md` §5. `docs/architecture/fitness/capability_trace_test.go`: `TestCapabilityPointersResolve`, `TestCapabilityPointerSymbolsResolve`, `TestCapabilityStatusConsistent`, `TestCapabilityPointersStable`, `TestCapabilityCoverage` — los 5, `PASS`. |
| **RF-213** — los 3 checks `deferred` pasan a `pass` | `conformance --todo`: pass +3, fail 0, sin `deferred` nuevos | Ver «Corrida de `conformance --todo`» abajo: `index-reconstruible`, `sin-migracion-incremental`, `writer-serializado` los 3 en `PASS`. `sin-cgo` (`TestNoDuckDBOrCGOStore`) seguía `PASS` desde antes (no contaba para el +3, confirmado por `hallazgos.md`). Total: pass 54→57 (+3 exacto), fail 0, deferred 212→209 (−3 exacto). |
| **RF-214** — `PARIDAD.md` | Este documento | — |

## Corrida de `conformance --todo` (extracto, 2026-07-24)

```
  PASS      error    index-reconstruible                        TestIndexRebuildsFromJSONL pasa
  PASS      warn     sin-migracion-incremental                  TestSchemaVersionTriggersRebuild pasa
  PASS      warn     writer-serializado                         TestWriterSerializedSingleConn pasa
  PASS      error    sin-cgo                                    TestNoDuckDBOrCGOStore pasa
  ...
  266 checks · pass 57 · fail 0 · error 0 · deferred 209 · n/a 0
```

## Los 5 comandos de CI + `conformance --todo` — todos verdes

```
$ golangci-lint fmt --diff                                                                → limpio
$ golangci-lint run ./...   (no listado en la consigna, pero SÍ corre en CI — ver Desviaciones) → 0 issues
$ go run github.com/fe3dback/go-arch-lint@latest check --project-path . \
    --arch-file docs/architecture/fitness/.go-arch-lint.yml                               → OK - No warnings found
$ go test ./... -race                                                                      → ok, todos los paquetes
$ go build ./...                                                                            → limpio
$ CGO_ENABLED=0 go build ./...   (requisito duro del boundary, confirmado explícito)        → limpio
$ bash scripts/estado.sh --check                                                            → en sync ✓
$ go run ./cmd/arnesia conformance --todo                                                   → pass 57 · fail 0 · deferred 209
```

## Desviaciones del spec (honestas, ninguna maquillada)

1. **RF-210, sub-escenario "arnés nuevo en caliente": no viable en esta iteración, tal como el
   propio spec anticipaba.** El watcher solo observa las entradas de `ArnesRegistry` que existían
   al momento de `Watch(ctx)` (boot). Registrar un arnés con el daemon ya corriendo no lo suma al
   watch-set — hace falta reiniciar el daemon para que `Rebuild()`+`Watch()` vuelvan a alinearse
   con el registro actual. Razón: sumar dinámicamente requeriría que `main.go` (o el propio
   `Watcher`) se suscriba a los eventos de `Register()` de `ArnesRegistry`, que hoy es un simple
   store de archivo sin ningún mecanismo de notificación — construirlo sería trabajo especulativo
   sin caller real hoy (el Portafolio ya cubre "arnés nuevo" vía `ObservarEnMapa`, explícito).
   Documentado en `arquitectura.md` §6.1 y en `observar-cambios-del-corpus.yaml` (CAP-23,
   `status: parcial`, no `vivo`, precisamente por este gap).
2. **Hallazgo propio, fuera de los RF numerados: la sesión ilustrativa de primer-uso
   (`internal/usecase/session_service.go:seedSessions()`) apunta al arnés `dev-full-cycle` asumiendo
   que sobrevive en el índice.** RF-207 exige explícito que ningún seed sobreviva a `Rebuild()`, y
   `Rebuild()` corre siempre al boot — así que en una instalación fresca (sin `dev-full-cycle`
   registrado en `ArnesRegistry`) esa sesión demo ahora ve un Mapa vacío/404 honesto en vez del
   grafo de ejemplo. Ni `hallazgos.md` ni `decisiones.md` examinaron este archivo — lo descubrí
   implementando RF-207 al pie de la letra. **No lo arreglé**: es una decisión de UX de
   primer-uso (¿la sesión demo debería auto-registrar el dogfood? ¿el FE debería tolerar un Mapa
   vacío en esa sesión?) fuera del alcance 100%-backend de este paquete. Detalle completo en
   `arquitectura.md` §6.3 — queda para que Fase 4 decida si amerita un ticket propio.
3. **`golangci-lint run ./...` (lint completo) no estaba en la lista de 5 comandos de la
   consigna** (solo `golangci-lint fmt --diff`), pero SÍ es un job separado en
   `.github/workflows/ci.yml` (`golangci-lint-action`). Se corrió de todos modos y se llevó a 0
   issues, incluyendo un fix de causa-raíz en `internal/adapters/loader/loader_test.go` (2 líneas:
   `os.Chdir`/`os.Getwd` manual → `t.Chdir()`) que el bump de `go.mod` a `1.25.0` (efecto colateral
   inevitable de `go get modernc.org/sqlite@latest`) volvió señalable por el linter `usetesting`
   (`t.Chdir()` es stdlib desde go1.24, antes no aplicaba con `go 1.23.0` declarado). Detalle en
   `arquitectura.md` §0/§6.4.
4. **T6 (bulk-load `synchronous=OFF`+`journal_mode=MEMORY`) y los cursores de reanudación
   `{path,inode,size,offset}` — no construidos.** Ninguno de los dos llegó a ser un RF numerado en
   `spec.md` (D1 los descarta explícito por escala; T6 nunca se promovió). No es una desviación del
   spec cerrado, es la ejecución literal de D1 — se documenta acá solo para que quede visible en la
   misma tabla que el resto.

## Revisión visual/UX (Fase 4, agente 4, 2026-07-24)

App real levantada (`127.0.0.1:4200`, entorno REAL del operador — `~/.arnesia/arneses.json` con
`dev-full-cycle`+`vitalia`, no editado), navegada con `claude-in-chrome`. Los 5 escenarios de la
consigna, con evidencia (screenshots + respuestas de API citadas) en
`revision-visual.md`. **Ningún bug encontrado — cero fixes de código en esta fase.**

- Escenario 1 (boot con datos reales, no demo): OK — `GET /api/harnesses` devuelve exactamente
  `dev-full-cycle`+`vitalia`, cero rastro de `content-studio-full`.
- Escenario 2 (consola limpia): OK — 3 pasadas (boot, tras watcher, tras reinicio), 0 errores/warnings.
- Escenario 3 (watcher fsnotify en vivo): OK — nodo nuevo apareció Y desapareció en el Mapa SOLO
  (sin recargar), vía SSE `event: map` + refetch; `dogfood/` quedó limpio tras la prueba.
- Escenario 4 (reinicio del daemon): OK — `kill -9` + reinicio, mismo estado exacto, cero pérdida.
- Escenario 5 (gap `seedSessions()`): confirmado no-reproducible en este entorno, como se esperaba
  (`dev-full-cycle` sí está registrado acá) — gap sigue documentado, sin cambios.

## Checklist de verificación (para la firma del operador, Fase 4)

- [x] Los 3 checks (`index-reconstruible`, `sin-migracion-incremental`, `writer-serializado`) se
      revisaron corriendo `go run ./cmd/arnesia conformance --todo` en vivo (no solo leyendo este
      documento). — confirmado por Fase 3 (agente 3); no vuelto a re-correr en Fase 4 porque no se
      tocó código (ver `revision-visual.md`).
- [x] Revisión visual/UX de los 5 escenarios (Fase 4) — ver `revision-visual.md`, sin bugs.
- [ ] El gap de RF-210 (arnés nuevo en caliente) se considera aceptable para esta iteración, o se
      abre un ticket de seguimiento.
- [ ] El hallazgo de `seedSessions()`/`dev-full-cycle` (desviación 2) se considera aceptable, o se
      abre un ticket de UX de primer-uso.
- [ ] `docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md` — decidir si v1.3
      (4/4 checks del checklist en `pass` real) amerita graduar `status: proposed → enforced`
      (criterio editorial, NO decidido por el agente — ver changelog del boundary).
- [ ] Firma 🧑‍⚖️ del operador: **[ ] PENDIENTE**
