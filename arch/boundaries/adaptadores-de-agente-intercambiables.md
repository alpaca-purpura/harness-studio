---
regla: adaptadores-de-agente-intercambiables
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-04
sources:
  - url: https://code.claude.com/docs/en/agent-sdk/overview
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://alistair.cockburn.us/hexagonal-architecture/
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - fitness/.go-arch-lint.yml#agent-adapters
  - fitness/arch_test.go:TestAgentPortHasNoConcreteLeak
severity: high
---

# Claude Code es un adaptador tras `AgentPort`, no el core

## L1 · Principio (estándar de industria)

**Ports & adapters aplicado al motor de agente.** El agente que dirige la creación/edición es un
**detalle intercambiable**, no el núcleo. VISION lo dice explícito: «el núcleo del producto es un
grafo de componentes agnóstico y cada agente es solo un adaptador» (OpenCode, Codex, pi,
Antigravity pueden entrar). El motor headless de Claude Code (subproceso `claude` hablando
stream-json por stdin/stdout) es UNA implementación de un puerto. *(oficial: agent-sdk/overview —
el SDK y el CLI hablan el mismo protocolo; experto: hexagonal)*

## L2 · Realización (este árbol Go+React)

- El dominio define **`AgentPort`** (crear/continuar sesión, enviar turno, stream de eventos,
  resolver permiso). `internal/adapters/agent/claudecode/` lo implementa spawneando el `claude`
  local. ⇐ L1: adaptador.
- **Solo la composition-root (`cmd/arnesia`) importa el adaptador concreto.** El resto del código
  depende de `AgentPort`, no de `claudecode`. Meter OpenCode mañana = un paquete nuevo
  `adapters/agent/opencode/`, cero cambios en el dominio. ⇐ L1: ports.
- **El tipo concreto no se filtra.** Nada fuera del adaptador conoce `stream-json`, flags de
  `claude`, o el shape de `control_request`. Si un caso de uso menciona `--max-turns`, es un leak.
- **Decisión HS-04 (frente B):** el adaptador es **subproceso-conductor**, no SDK-sidecar (el SDK
  es un wrapper del mismo binario → no aporta lo que Go no pueda, y mete runtime Node). Managed
  Agents descartado (sandbox remoto, no toca el filesystem local). El detalle del conductor
  (flags, discovery, auth, event-sourcing) vive en
  [`conductor-no-parsea-jsonl.md`](./conductor-no-parsea-jsonl.md) y
  [`permisos-gui-human-in-the-loop.md`](./permisos-gui-human-in-the-loop.md).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| agent-port-existe | el dominio define `AgentPort`; los casos de uso dependen de él | error | «conductor cableado directo, sin puerto» | go-arch-lint#agent-adapters |
| adapter-solo-en-root | solo `cmd/**` importa `adapters/agent/claudecode`; el resto usa `AgentPort` | error | «adaptador concreto importado fuera de la raíz» | go-arch-lint#agent-adapters |
| no-leak-concreto | ningún flag/término stream-json de `claude` aparece fuera del adaptador | warn | «detalle de Claude Code filtrado al dominio» | arch_test.go:TestAgentPortHasNoConcreteLeak |
| segundo-adaptador-posible | agregar un adaptador nuevo no requiere tocar `domain/**` (test de humo con fake) | info | «grafo no-agnóstico: el motor está soldado» | arch_test.go |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = grafo agnóstico + adaptador (VISION) sobre
  ports&adapters. L2: `AgentPort` en el dominio, `claudecode` como impl subproceso-conductor
  (SDK-sidecar y Managed Agents descartados), concreto solo en la composition-root. 4 checks.
