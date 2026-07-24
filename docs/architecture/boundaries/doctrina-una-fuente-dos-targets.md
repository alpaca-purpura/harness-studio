---
regla: doctrina-una-fuente-dos-targets
version: 1.0
updated: 2026-07-23
status: proposed
ledger: HS-11
sources:
  - url: https://code.claude.com/docs/en/plugins
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://en.wikipedia.org/wiki/Single_source_of_truth
    autoridad: estándar
    revisado: 2026-07-23
enforced_by: []
severity: error
---

# Un ruleset ejecutable y su guía generativa no deben divergir

## L1 · Principio (estándar de industria)

**Single source of truth con múltiples targets derivados.** Cuando la misma doctrina debe
alimentar dos consumidores de forma distinta (un motor que la EJECUTA como checks, y un agente
que la LEE como guía prosa), ambos targets deben derivar de la MISMA fuente en build — nunca
mantenerse a mano por separado, porque divergen en silencio. *(estándar: single source of
truth; oficial: Claude Code plugins — `kit/` es la superficie de guía que el CLI consume vía
`--plugin-dir`)*

## L2 · Realización (este árbol Go) — GAP honesto

`embed_doctrina.go` embebe DOS árboles por separado:

- **Target A (ejecutable):** `//go:embed docs/architecture/knowledge docs/architecture/boundaries
  docs/architecture/conventions docs/architecture/contracts/schema` — el `RulesetPort`
  (`internal/adapters/conformance/mechanism/schema.go`) parsea estos `.md`+schema EN VIVO;
  no hay un `ruleset.json` intermedio generado (el plan original de HS-07 asumía uno; la
  implementación real de HS-08/HS-10 embebe el árbol fuente directo — MEJOR: cero paso de
  generación que pueda quedar stale).
- **Target B (guía):** `//go:embed all:kit` → `kit/doctrine.md` — prosa que el conductor
  inyecta vía `--append-system-prompt-file`. **`kit/doctrine.md` está escrito A MANO**,
  referenciando METODOLOGIA §3/VISION A1-A7 por número — NO se deriva automáticamente de
  `docs/architecture/knowledge/` ni de `docs/process/metodologia.md`.

**El gap real:** no existe ningún check que detecte si `kit/doctrine.md` queda stale respecto a
la doctrina fuente (p.ej. METODOLOGIA §3 cambia de forma y `kit/doctrine.md` sigue citando la
forma vieja). El principio L1 no está enforced — está observado y registrado. `status: proposed`
hasta que exista un enforcer real (candidato: check de presencia/freshness de las referencias
citadas por sección, o mover `kit/doctrine.md` a plantilla generada).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| dos-targets-embebidos | ambos targets (ruleset ejecutable + `kit/doctrine.md`) están embebidos en el binario, ninguno es archivo externo opcional | error | «target de doctrina no embebido — puede faltar en el binario instalado» | (pendiente — hoy verificación manual de `embed_doctrina.go`) |
| kit-doctrine-no-diverge | `kit/doctrine.md` no cita una sección/número de METODOLOGIA/VISION que ya no existe en esa forma | error | «guía prosa del kit desincronizada de la doctrina fuente» | (pendiente — deuda BACKLOG, requiere diseño de freshness-check) |

## Changelog

- 2026-07-23 · v1.0 · Nodo fundacional — draft de
  `docs/product/research/2026-07-05-arquitectura-inyeccion-knowhow.md` §9 (HS-07) materializado
  como boundary formal (deuda BACKLOG «3 boundaries de research → arch/», parcialmente cerrada:
  el nodo existe y el gap queda VISIBLE en vez de enterrado en un ledger). Investigación en este
  cierre confirmó que la arquitectura real diverge del draft (sin `ruleset.json` intermedio —
  mejora sobre el plan) pero que `kit/doctrine.md` es prosa mantenida a mano sin drift-check —
  `status: proposed`, no se fabrica un enforcer que no existe. Construir el freshness-check
  queda en BACKLOG como ítem propio (no bloqueante).
