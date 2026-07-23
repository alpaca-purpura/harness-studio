# Aprendizajes por ticket — mejorar-arnes-conversando

> Bitácora acumulativa (protocolo del spec §Protocolo). Cada ticket ABRE leyendo este archivo y
> CIERRA appendeando su entrada: diseño final, gotchas, y qué NO re-descubrir. Objetivo: cada
> ticket gasta menos tokens que el anterior.

## Base (de la investigación del spike + diagnóstico Vitalia, 2026-07-22)

- **Mapa de código verificado (file:line):** spawn/flags → `internal/adapters/agent/claudecode/conductor.go`
  (`SpawnArgs` L59-104, `ctxPct` L554-578, `Close` proceso L383-389). Orquestación →
  `internal/usecase/session_service.go` (`Turn` L256, `spawnLocked` L305 — SIEMPRE pasa
  `Resume=ClaudeSessionID`, `consume`/`EventResult` L385-404 — ahí se guarda `CtxPct` pisándolo y
  se appendea `Conv`, `Close` L229-251 — BORRA la entrada). Kit ② →
  `internal/adapters/provision/provisioner.go` (materializa compartido a `~/.arnesia/`, huella
  sha256, `Injection{PluginDirs,SystemPromptFile,AddDirs,MCPConfigFile}`). Loader →
  `internal/adapters/loader/loader.go` (`LoadArnesInfo` L72-150; `reconocerRegla` = SOLO CLAUDE.md
  raíz). Broker SSE → `internal/adapters/transport/http/broker.go` (`EventMap="map"` L18 SIN USO;
  FE `sse.ts` solo escucha `dock`). Portafolio → `internal/domain/portafolio.go`
  (`IdentidadArnes`/`Canonico`/`Instalacion.Deriva`), `internal/usecase/portafolio.go`
  (`ObservarEnMapa` patrón degradado L297-310). Sesiones → `internal/domain/session.go`
  (`Session.Arnes` L75, `Conv` L94, `ClaudeSessionID` L88, `RolSys` L53), store JSON atómico
  `internal/adapters/store/registry.go`. Índice → `internal/adapters/index/store.go` (in-memory,
  TODO fase-5 SQLite/walk JSONL; `watch/watcher.go` = stub no-op).
- **Vitalia (caso de validación):** instalación en `/home/chalreme/Proyectos/luana-vitalia/vitalia`
  (brand overlay; el harness raíz del worktree viene de `harness@prenter-marketplace`). Sin
  canónico/home en ningún marketplace. Grafo actual 2 nodos/0 edges porque el loader no ve
  `.claude/rules/*.md`. Sello `arnes.l0.json` presente (v0.1.0, sin procedencia). `skills/` solo
  README. Rule `shell-mockup-per-component.md` marcada SUPERSEDED en el README de rules.
- **CLI:** `arnesia conformance --arnes` recibe un GRAPH JSON (no un dir); `arnesia index -o out.json <dir>`
  genera el grafo. `estado.sh` usa `dogfood/dev-full-cycle.graph.json`.
- **Gotchas heredados:** vitest-browser no headless en bg (`pnpm run verify` para FE);
  `initialize` handshake obligatorio (no tocar); lefthook `pre-commit estado-cifras` regenera
  cifras del checkpoint solo.

---

## T-L · Loader reconoce rules/ (RF-183) — CERRADO ✅

- **Diseño final:** reconocedor nuevo `reconocerReglasDir(elementos)` en
  `internal/adapters/loader/regla.go` (WalkDir recursivo sobre `rules/` bajo elementos — cubre
  `.claude/rules/` instalado y `rules/` plugin con el mismo código porque `detectarElementos` ya
  resuelve la base). Un nodo `ClaseRule`/`BandaBase` por `.md`; id = ruta relativa sin `.md`
  (`sub/tema`); nombre = frontmatter opcional vía `nombreDe` (patrón commands); no-.md →
  no-reconocido visible.
- **Desvío vs spec original:** README.md del dir de rules SÍ es nodo rule — el runtime CC carga
  todo .md del dir; fidelidad al runtime > tratarlo como índice aparte. Spec RF-183 corregido.
- **El estándar as-code no necesitó cambio:** `knowledge/elements/rules.md` L1.4 ya lo cubría; el
  atrasado era el loader. `paths:` scoping NO se modela en el grafo v1 (semántica de carga
  condicional del runtime — anotado, no perdido).
- **Verificado:** `TestReconocerReglasDir` verde + suite entera verde + `arnesia index` contra
  Vitalia real: 2 → **6 nodos**.
- **Para el siguiente (T2):** helpers de test loader: `escribir(t,ruta,contenido)` + `ids(nodos)`
  (`loader_test.go:337,347`). `LoadArnes` ≠ `LoadArnesInfo` (el segundo devuelve `Info.Aviso`
  degradado). Posible colisión de id entre reconocedores (rule `hipaa-lite` vs skill homónima) es
  preexistente y tolerada — no resolver acá.

## T2 · Reindex-tras-turno (RF-184/185) — CERRADO ✅

- **Diseño final:** tipo `Reindexer func(ctx, arnesID, cwd)` + builder testeable
  `NewTurnReindexer(idx ports.IndexPort, load func(dir)(domain.Graph,error))` en
  `internal/usecase/session_reindex.go` (loader inyectado como func — usecase no importa
  adapters, patrón `RoleSource`). Cableado vía `SetReindexer` (setter, NO parámetro 10 del
  constructor — menos churn en tests/main). El disparo vive en `consume`/`EventResult`
  (`session_service.go`): se capturan `reindex/arnes/cwd` DENTRO del lock, se llama FUERA (hace
  IO). `sessionRuntime.cwd` nuevo, estampado en `spawnLocked`.
- **Degradación (3 salidas):** sello OK → tal cual; sin sello → síntesis `HuellaPath(cwd)` +
  Degradado (idéntico a `ObservarEnMapa`); carga rota → grafo VACÍO Degradado bajo el id del
  registro. Nunca foto vieja, nunca Upsert con `Arnes` nil.
- **Gotchas descubiertos:** firmas SIN ctx: `svc.Create(domain.Session)` y `svc.Turn(id, text)`.
  Stubs de sesión reusables en `session_permisos_test.go` (package `usecase_test`): `stubAgent`
  (`.sessions[i].events <- ports.AgentEvent{...}`), `stubStore`, `stubResolver{path}`, `stubPub`.
  **R2 coverage MUERDE en `go test` (`TestCapabilityCoverage`)**: archivo Go nuevo sin capability
  = suite roja → crear el YAML en el MISMO ticket (nació `CAP-94 usecases/reindex-tras-turno`;
  próximo libre: CAP-95).
- **Para el siguiente (T3):** el broker ya tiene `EventMap="map"` (broker.go:18). El publisher
  del usecase es `EventPublisher.Publish(eventType, data)` (brokerPublisher en main). Extender
  `NewTurnReindexer` con un `pub EventPublisher` opcional para emitir `map` tras Upsert — un solo
  lugar. FE: `sse.ts` listener + `workspace-stage.tsx` refetch (`useEffect` L59-108 ya fetchea por
  `viewedId`).

## T3 · Push por `event: map` + FE refetch (RF-186/187) — CERRADO ✅

- **Diseño final:** `NewTurnReindexer(idx, load, pub)` publica `mapFrame{harness_id, degradado}`
  por `event: map` tras CADA Upsert (sano o degradado — el FE debe enterarse del roto). FE:
  `connectDock(onFrame, onStatus, onMap?)` (mismo EventSource multiplexado, listener `map` solo
  si hay callback); store nuevo `useMapLive` (`shared/store/map-live-store.ts`: `rev` por arnés +
  `bump`); `sessions-store` cablea `onMap → bump`; `workspace-stage.tsx` efecto SEPARADO por
  `mapRev` que refetchea `getGraph` sin resetear selección/conformance (el efecto de navegación
  queda intacto — refresh ≠ navegación; refetch fallido conserva el grafo visible).
- **Corrección a T2 destapada acá:** el reindex sin sello upsertea bajo el ID DEL REGISTRO, no
  `HuellaPath` — una llave sintética duplicaría la entrada del índice y dejaría stale la que el
  Mapa mira. (La síntesis HuellaPath es del flujo Portafolio/ObservarEnMapa, donde NO hay llave
  previa.) Edge anotado: si el chat cambia el `id` DEL SELLO, el grafo entra bajo el id nuevo y
  el Mapa que mira el viejo queda stale — aceptado en V1, es una edición deliberada del sello.
- **Gotchas:** R2 coverage también cubre `web/src` (map-live-store.ts+test exigieron capability →
  CAP-95 `fe-mapa/reindex-en-vivo`; próximo libre CAP-96). TS strict `TS4111`: `Record` se accede
  con corchetes (`rev["vitalia"]`). Biome formatea distinto que a mano — correr
  `pnpm exec biome format --write` antes de `verify`. Los unit tests FE (proyecto `unit`, Node
  sin Chromium) SÍ corren en bg: `pnpm vitest run --project unit`.
- **Para el siguiente (T4 E2E Vitalia):** daemon corre con `go run ./cmd/arnesia serve` (o el
  binario instalado). El FE dev en `:4200` sirve la SPA embebida del daemon o `pnpm dev`
  (verificar puerto). Sesión contra vitalia YA registrada en `~/.arnesia/arneses.json` (llave
  `vitalia`). Evidencia a `PARIDAD.md`.

## T4 · E2E vivo Vitalia (RF-188) — CERRADO ✅ (evidencia completa en PARIDAD.md)

- **Receta E2E reproducible:** daemon aislado
  `arnesia serve -addr 127.0.0.1:4213 -sessions <scratch>/sessions.json -arneses <scratch>/arneses.json`;
  SSE con `curl -sN /events > sse.log`; sesión `POST /api/sessions {arnes, frente, path, view}`;
  turno `POST /api/sessions/{id}/turn {text}`; tarjetas: frames `kind:permission` en sse.log →
  `POST /api/sessions/{id}/permission {request_id, decision:"allow"}` (aprobador
  `scratchpad/e2e/aprobador.py`). Sin token en dev (Host+Origin only) — curl pasa.
- **3 bugs reales destapados por la corrida** (ninguno visible en tests unitarios):
  1. **`createSession` no indexaba** el path registrado → `roleFor` (consulta el índice) devolvía
     "" al primer spawn → SIN flags de permisos → Write denegado SIN tarjeta (Claude termina el
     turno pidiendo aprobación a nadie). Fix: `onRegistered` (el `loadArnesDir` del composition
     root) llamado en `createSession`, best-effort. Test `TestCreateSessionIndexaAlRegistrar`.
  2. **Un arnés sin `rol` en el sello NO puede escribir por chat** (policy deny-by-default) —
     parte de la reparación de Vitalia fue sellarle
     `rol: "Ingeniería · Desarrollo full-cycle"` (uno de los 3 roles de la policy spike:
     backend-dev · Ingeniería · Desarrollo full-cycle · reviewer). Grounding T10 debería DECIR el
     rol en la tarjeta.
  3. **`ctxPct` ROTO para Fork C:** suma usage ACUMULADO del turno (cache_read re-contado por
     tool-call) → 100 % en un turno con ~10 tool-calls. T7 debe leer el usage del ÚLTIMO API call
     (frames assistant traen `message.usage` — extender `assistantMsg`) y recién ahí armar el
     umbral. `defaultContextWindow=200k`; el model real fue `claude-opus-4-8[1m]` (ventana 1M) —
     verificar la resolución de ventana por `modelUsage[f.Model]` (posible mismatch de llave).
- **Continuidad probada:** daemon reiniciado a mitad de conversación → `--resume` retomó la misma
  sesión y Claude completó lo que había dejado pendiente («staged» el SKILL.md en su contexto).
- **Para el siguiente (T10):** la tarjeta de identidad tiene de dónde salir:
  `PortafolioService.Listar` → match de cwd contra `entrada.Canonico.Path` /
  `Instalaciones[].InstallPath` (usar `canonicalPathPortafolio` para comparar). Incluir el ROL del
  sello en la tarjeta (aprendizaje 2). `Injection.SystemPromptFile` hoy apunta al doctrine.md
  compartido; plan: `ProvisionSession(ctx, sessionID, tarjeta)` en el provisioner que escribe
  `~/.arnesia/sessions/<id>/system.md` = doctrine + tarjeta.

## T10 · Tarjeta de identidad por sesión (RF-189/190) — CERRADO ✅

- **Diseño final:** interfaz `InjectionProvisioner` gana `ProvisionSession(ctx, sessionID, extra)`
  (extra=="" degrada a `Provision` — cero cambio para quien no la usa; `RunService` sigue con
  `Provision`). El provisioner escribe `~/.arnesia/sessions/<id>/system.md` = doctrina ② + extra,
  RE-escrito en cada spawn (la tarjeta refleja el Portafolio ACTUAL). `TarjetaIdentidad(entradas,
  arnesID, cwd, rol)` = función PURA en `session_grounding.go` (canónico / instalación-con-loop-A4
  / suelto honesto; sin rol lo DICE — aprendizaje T4). `SetGrounding` + closure en main sobre
  `portafolioSvc.Listar` + `roleFor`.
- **La sección de reparación A4 vive en la TARJETA, no en doctrine.md compartido** — solo se
  inyecta cuando la copia es instalación (RF-190 como se especificó). Pide a Claude cerrar cada
  respuesta declarando la causa (instalación vs base) — insumo barato para el backport futuro.
- **Gotcha:** un solo implementador real de la interfaz (provision.Provisioner) + stubs de test —
  extender la interfaz fue barato. `Instalacion.InstallPath` ya viene canonicalizado (C-N-3), pero
  igual se canonicaliza ambos lados al comparar. CAP-96 nace (próximo libre CAP-97).
- **Para el siguiente (T1):** el picker FE ya distingue tipos en `copiasDe` (mezcla canónico +
  instalaciones); falta ROTULAR (chip «instalación → sesión de reparación»), payload de create con
  marca opcional, y re-evaluar deriva tras turno. La re-evaluación de deriva vive en el
  Portafolio — buscar qué función la computa (grep `Deriva`/`deriva` en usecase/portafolio.go y
  adapters/portafolio) antes de inventar nada.

## T1 · Picker rotula + deriva viva (RF-191/192/193) — CERRADO ✅

- **Diseño final:** `domain.Session.Reparacion` (bool) seteado por el picker al elegir una copia
  no-canónica (`ArnesElegido.reparacion` → `NewSession` → `createSessionBody` → `Session`); el
  picker ya tenía chips de tipo — solo se agregó el hint «→ reparación» (con tooltip A4) en la
  sub-lista y el chip `reparación` en la SessionCard expandida. `PortafolioService.ReevaluarDeriva
  (ctx, installPath)` re-corre el evaluador (misma resolución home/version que `candidatoDe`:
  `Identidad.Home` → fallback `CanonicalizarRepo(Origen.Registry)`; version de `Origen.Version`) y
  persiste SOLO si cambió. Wiring: el closure del reindexer en main.go encadena
  `turnReindex(...)` + `ReevaluarDeriva(cwd)` — cero plumbing nuevo en SessionService.
- **Gotchas:** los fakes del Portafolio ya existen en `portafolio_test.go`
  (fakePortafolioStore/Scanner/Loader + fakeDerivaEvaluator fijo — para probar cambio de estado
  hace falta un evaluador propio `derivaFija`). El picker calcula `copiasDe` on-the-fly — para
  saber si la copia elegida es canónica se re-busca por path en `crear()`.
- **Deuda anotada:** el chip `reparación` y el hint del picker no tienen story-test propio
  (vitest-browser no corre headless en bg — validación visual va al E2E final / gate humano).
- **Para el siguiente (T7):** PRIMERO reparar `ctxPct` (hallazgo T4: usa usage ACUMULADO del
  turno → 100 % espurio). El fix: capturar `message.usage` del ÚLTIMO frame assistant
  (`assistantMsg` hoy solo parsea Content — extender) y usar ESO en el `EventResult`; el
  `modelUsage[f.Model].contextWindow` puede no matchear la llave del model `[1m]` — mantener el
  fallback pero loguear. Después: histórico en `Session` (slice de ctxPct por turno) + umbral 40 %
  configurable + marca `rotacion-pendiente`.

## T7+T8+T9 · Rotación de contexto completa (RF-194..199) — CERRADOS ✅

- **Fix ctxPct:** `translate(line, **usage)` — el pump (único goroutine) pasa el puntero;
  frames `assistant` traen `message.usage` (el API call real); result usa el último no-cero,
  acumulado solo de fallback. Medido en vivo: 14 % donde antes daba 100 %.
- **Rotación:** `rotarLocked` en `Turn()` (nunca en consume — el turno en curso jamás se corta):
  close proceso + `CadenaCC=append(ClaudeSessionID)` + `Checkpoint=CheckpointMecanico(Conv)` +
  breadcrumb `RolSys` + limpiar resume. El spawn siguiente va por el camino normal (live==nil) y
  el checkpoint viaja DENTRO del system.md por sesión (concat con la tarjeta — un mecanismo, dos
  usos, cero archivo nuevo).
- **E2E vivo:** rotación real contra vitalia con `-rotacion-umbral 1`: cc rotado, cadena
  registrada, breadcrumb en su lugar, y el proceso fresco contestó una pregunta que dependía del
  turno pre-rotación («qué rule respalda esa skill» → hipaa-lite). BONUS: cerró con «Causa
  diagnosticada: instalación» — la tarjeta A4 (T10) opera sola.
- **Gotchas:** `Session.Checkpoint` persiste (sobrevive restarts y se re-inyecta en cada spawn
  posterior hasta la próxima rotación — feature, no bug: el fresco siempre sabe dónde venía).
  Con umbral bajo TODO turno re-marca pendiente — esperable, el default real es 40. En spawn
  fresco el primer ctx puede saltar (~68 %: cache_creation del harness pesado del worktree
  luana-vitalia) — el umbral 40 con ese harness rotaría cada 1-2 turnos: si molesta en la
  práctica, subir umbral o excluir cache_creation del cómputo (decisión de spec futura, anotada).
- **Para el siguiente (T6):** `CadenaCC` ya llena. `~/.claude/projects/<hash>` — averiguar el
  algoritmo de hash del path del cwd ANTES de codear el lector (mirar un dir real: los JSONL de
  las corridas E2E de hoy existen para `/home/chalreme/Proyectos/luana-vitalia/vitalia`).

## T6+T5 · Historial B2 + cierre (RF-200..206) — CERRADOS ✅

- **El «hash» del corpus no es hash:** `~/.claude/projects/<dir>` = el cwd con todo carácter
  no-alfanumérico → `-` (verificado contra el CLI real). `history.Reader` (adapter nuevo, read-only
  — respeta `indice-desechable-jsonl-es-verdad`) parsea user/assistant, salta bloques sin texto.
- **Close archiva ANTES de borrar** (`archivarLocked`): metadata sin Conv a
  `sesiones-cerradas.json` (segundo `store.NewRegistry` — patrón reutilizado). Gotcha real: las
  sesiones nacidas ANTES del estampado de `Cwd` archivaban sin join → backfill vía
  `resolver.Resolve(arnes)` en el archivo. E2E: conversación real reconstruida completa (9 turnos,
  2 JSONLs cosidas por la cadena de rotación, 0 faltantes).
- **FE:** transporte en `model/conversaciones-store.ts` (patrón portafolio-picker-store, widget
  props-puro); tipos de metadata (`cerrada_en`/`turnos`/`cadena_cc`) en `Session` del FE.
- **Cifras:** capabilities 93→98 (CAP-94..98), cobertura 100 %, `--todo` 48 pass/0 fail, dogfood
  sin regresión. `estado.sh` regeneró el checkpoint solo.
- **Nota metodológica del paquete completo:** el protocolo investigación→TDD→verificación→
  aprendizajes CUMPLIÓ su promesa — cada ticket abrió con el archivo al día y no re-descubrió
  nada; los 3 bugs grandes (roleFor sin índice, ctxPct acumulado, Cwd sin estampar) los destapó
  la VERIFICACIÓN REAL (E2E vivo), nunca los tests unitarios. La regla «verificar contra el
  binario/corpus real» va en cada ticket futuro de este circuito.
