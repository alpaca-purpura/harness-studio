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
