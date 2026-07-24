---
regla: maquinaria-no-contamina-arnes
version: 1.0
updated: 2026-07-23
status: enforced
ledger: HS-11
sources:
  - url: https://code.claude.com/docs/en/cli-reference
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://reproducible-builds.org/docs/build-path/
    autoridad: estándar
    revisado: 2026-07-23
enforced_by:
  - fitness/arch_test.go:TestMaquinariaNoContaminaArnes
severity: error
---

# La config del agente-fábrica no se filtra al artefacto que edita

## L1 · Principio (estándar de industria)

**Aislamiento hermético build↔artefacto.** Un builder (compilador, generador, agente de
fábrica) NUNCA debe dejar residuo de su propia configuración/entorno dentro del artefacto que
produce o edita — el mismo principio que exige *reproducible builds* (build-path independence:
la ruta/entorno del que compila no debe filtrarse al binario) y que en Claude Code se realiza
vía superficies de inyección *session-scoped* (`--plugin-dir`, `--append-system-prompt-file`)
en vez de instalación persistente en el árbol objetivo. *(oficial: CLI reference —
`--plugin-dir`/`--append-system-prompt-file` son flags de sesión, no mutan el filesystem del
target; estándar: reproducible-builds.org — build-path independence)*

## L2 · Realización (este árbol Go)

ArnesIA es un agente-fábrica (① doctrina + ② kit/maquinaria) que opera SOBRE el árbol de un
arnés (③, el producto). La inyección de ①② al proceso `claude` que edita ③ entra **solo por
flags de sesión**, jamás escribiendo en `<ruta-arnés>/.claude/`:

- `internal/ports/kit.go#Injection` — `PluginDirs`/`SystemPromptFile`/`AddDirs`/`MCPConfigFile`
  son todos rutas APP-owned (`~/.arnesia/...`), nunca rutas dentro del arnés.
- `internal/adapters/agent/claudecode/conductor.go#SpawnArgs` traduce `Injection` a
  `--plugin-dir`/`--append-system-prompt-file`/`--add-dir`/`--mcp-config --strict-mcp-config` —
  exportada explícitamente para que el enforcer llame la función real, no un texto simulado.
- **Prohibido `--bare`**: rompe el auth de suscripción del operador (salta OAuth/keychain,
  corrección ya cementada en `conductor-no-parsea-jsonl.md`) — este boundary lo re-verifica
  desde el ángulo de la inyección de doctrina, no del parseo de eventos.
- Degradación honesta: `Injection` en cero (falla de provisión) ⇒ spawn SIN doctrina, jamás
  bloquea la sesión (comentario `SpawnOpts.Injection`).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| flags-no-bare | `SpawnArgs` con `Injection` poblada nunca emite `--bare` | error | «inyección rompe auth de suscripción» | fitness/arch_test.go:TestMaquinariaNoContaminaArnes |
| flags-session-scoped | `SpawnArgs` emite `--plugin-dir`/`--append-system-prompt-file`/`--add-dir`/`--mcp-config` cuando `Injection` está poblada | error | «doctrina sin vía de inyección» | fitness/arch_test.go:TestMaquinariaNoContaminaArnes |

## Changelog

- 2026-07-23 · v1.0 · Nodo fundacional — draft de
  `docs/product/research/2026-07-05-arquitectura-inyeccion-knowhow.md` §9 (HS-07) materializado
  como boundary formal (deuda BACKLOG «3 boundaries de research → arch/», cerrada). Directo a
  `enforced`: la implementación ya existía desde HS-11 (`SpawnArgs`/`Injection`), solo faltaba
  el nodo as-code + el test — construido en el mismo cierre (`TestMaquinariaNoContaminaArnes`,
  llama la función real, cero mock de texto).
