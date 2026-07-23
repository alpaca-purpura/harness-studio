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
