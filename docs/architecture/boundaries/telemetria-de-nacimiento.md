---
regla: telemetria-de-nacimiento
version: 1.0
updated: 2026-07-23
status: proposed
ledger: HS-11
sources:
  - url: https://opentelemetry.io/docs/concepts/observability-primer/
    autoridad: oficial
    revisado: 2026-07-23
enforced_by: []
severity: error
---

# Todo arnés nace observable (VISION p9)

## L1 · Principio (estándar de industria)

**Observability by default, no opt-in.** Un sistema producido en masa (aquí: arneses) debe
nacer instrumentado — la telemetría no es algo que se agrega después de un incidente, es parte
del template de creación. *(oficial: OpenTelemetry — observability primer, instrumentación como
parte del ciclo de vida del servicio, no un add-on posterior)*

## L2 · Realización (este árbol Go) — SIN IMPLEMENTAR

Investigación de este cierre (2026-07-23): `grep`s de `telemetry-emit`/`OTLP`/`otlp`/`scaffold`
sobre `internal/` y `cmd/` no encuentran NINGÚN código — no hay collector OTLP embebido, no hay
`scaffold(arnesID, clase)` que materialice un hook `telemetry-emit` en el `.claude/settings.json`
de un arnés nuevo. El draft de HS-07 (`docs/product/research/2026-07-05-arquitectura-inyeccion-knowhow.md`
§8.7 y §9) sigue siendo exactamente eso — un draft. Coincide con la deuda BACKLOG independiente
«telemetría JSONL → indexer real» (`bloqueo`, desbloquea las capas Tokens/Desempeño/Proceso del
Mapa): es el MISMO trabajo sin construir, visto desde dos ángulos (boundary de arquitectura vs.
feature de producto).

**Por qué queda `proposed` y no se fabrica nada:** materializar este nodo es hacer VISIBLE el
gap, no simularlo. Escribir un `status: enforced` sin código sería exactamente el «pass
fabricado» que la doctrina de honestidad prohíbe.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| scaffold-emite-hook | `scaffold` materializa `telemetry-emit` en el `.claude/settings.json` de todo arnés creado | error | «arnés nuevo nace sin telemetría» | (pendiente — no existe `scaffold`, ver BACKLOG) |
| collector-embebido | el daemon embebe un collector OTLP que recibe la telemetría del hook | error | «telemetría emitida pero nadie la recibe» | (pendiente — no existe collector) |

## Changelog

- 2026-07-23 · v1.0 · Nodo fundacional — draft de
  `docs/product/research/2026-07-05-arquitectura-inyeccion-knowhow.md` §9 (HS-07) materializado
  como boundary formal (deuda BACKLOG «3 boundaries de research → arch/»). Investigación
  confirmó CERO implementación — nace `proposed`, honesto, ligado a la deuda BACKLOG «telemetría
  JSONL → indexer real» (mismo trabajo pendiente, no duplicar el esfuerzo cuando se construya).
