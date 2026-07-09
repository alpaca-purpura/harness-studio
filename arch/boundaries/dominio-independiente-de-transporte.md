---
regla: dominio-independiente-de-transporte
version: 1.1
updated: 2026-07-09
status: proposed
ledger: HS-04
sources:
  - url: https://alistair.cockburn.us/hexagonal-architecture/
    autoridad: experto
    revisado: 2026-07-05
  - url: https://go.dev/doc/effective_go
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/.go-arch-lint.yml#domain
  - fitness/.go-arch-lint.yml#transport-http
  - fitness/.go-arch-lint.yml#transport-sse
  # cubre modernc.org/sqlite + /transport/ + /sse (lo que depguard no banea):
  - fitness/arch_test.go:TestDomainIndependentOfTransport
severity: critical
---

# El dominio no depende del transporte ni del almacenamiento

## L1 · Principio (estándar de industria)

**Ports & adapters + dependency inversion.** El modelo de dominio y los casos de uso se expresan
en términos del negocio, no de HTTP, SSE, SQL o JSON. El transporte (servidor HTTP/SSE) y la
persistencia (el store del índice) son adaptadores que **implementan puertos** que el dominio define; el
dominio no los importa. Esto permite testear el dominio sin red ni disco, y cambiar de transporte
o de store sin tocar reglas de negocio. *(experto: Cockburn; oficial: Go — «accept interfaces,
return structs»)*

## L2 · Realización (este árbol Go+React)

- **`internal/domain/**` (el grafo agnóstico de componentes, el spine de estados, los contratos)
  NO importa `net/http`, `database/sql`, `modernc.org/sqlite`, ni el paquete SSE.** ⇐ L1: DI.
- El dominio define **puertos** (`GraphStore`, `RunStore`, `EventBus`); los adaptadores viven en
  `internal/adapters/` (`index/` = store [map in-memory + JSON atómico hoy; SQLite modernc en fase 5], `transport/http/`, `transport/sse/`). Los adaptadores
  importan el dominio; el dominio no los importa.
- **El transporte no contiene reglas de negocio.** Un handler HTTP traduce request↔caso-de-uso y
  nada más; si un handler decide algo del dominio (p.ej. cuándo un arnés está «sano»), es un
  smell — esa lógica vive en el dominio y se testea sin HTTP. ⇐ L1: ports.
- **Los tipos del wire se generan, no se mezclan con el dominio.** Los DTOs del API salen de
  [`../contracts/`](../contracts/) (OpenAPI→Go); el mapeo DTO↔dominio vive en el adaptador de
  transporte, no en el dominio. Esto es lo que mantiene la API sin driftar de la doc.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| domain-no-http | `internal/domain/**` no importa `net/http` ni el paquete SSE | error | «dominio acoplado al transporte» | depguard / go-arch-lint#domain |
| domain-no-sql | `internal/domain/**` no importa `database/sql` ni `modernc.org/sqlite` | error | «dominio acoplado al store» | depguard (database/sql) · arch_test.go:TestDomainIndependentOfTransport (modernc.org/sqlite) · go-arch-lint#domain |
| transport-sin-reglas | los handlers no derivan estado de salud/eval/gate (solo traducen) | warn | «regla de negocio en el handler» | arch_test.go |
| dtos-generados | los tipos del wire vienen de `contracts/gen`, no escritos a mano en el dominio | info | «DTO a mano — riesgo de drift doc↔código» | schema/codegen |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = ports&adapters + DI. L2: dominio ⊥ HTTP/SSE/
  SQLite vía puertos; adaptadores en `internal/adapters/`; DTOs generados desde `contracts/`. 4
  checks.
- 2026-07-09 · v1.1 · **Sync HS-18.** El `index/` store ACTUAL es map in-memory + JSON atómico
  (SQLite modernc = fase 5); se relabela «SQLite» → «el store» donde se citaba como persistencia
  vigente. El principio (dominio ⊥ transporte/store) y la ban-list de imports (`database/sql`,
  `modernc.org/sqlite`, guardrail anti-coupling futuro) quedan intactos. Sin cambios de checks/status.
