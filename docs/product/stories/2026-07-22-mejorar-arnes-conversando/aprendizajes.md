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

(entradas por ticket se appendean debajo)
