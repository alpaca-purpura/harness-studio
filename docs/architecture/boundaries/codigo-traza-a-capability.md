---
regla: codigo-traza-a-capability
version: 1.4
updated: 2026-07-14
status: enforced
ledger: HS-20
sources:
  - url: https://cucumber.io/docs/guides/living-documentation/
    autoridad: experto
    revisado: 2026-07-09
  - url: https://pubs.opengroup.org/togaf-standard/business-architecture/business-capabilities.html
    autoridad: oficial
    revisado: 2026-07-09
enforced_by:
  - docs/architecture/fitness/capability_trace_test.go:TestCapabilityPointersResolve
  - docs/architecture/fitness/capability_trace_test.go:TestCapabilityCoverage
  - docs/architecture/fitness/capability_trace_test.go:TestCapabilityStatusConsistent
  - docs/architecture/fitness/capability_trace_test.go:TestCapabilityPointersStable
severity: error
---

# Todo código fuente traza a un capability (`docs/product/capabilities/` = SSoT funcional)

## L1 · Principio (estándar de industria)

**La documentación de lo que un sistema HACE debe estar atada al código, o se pudre.** Dos marcos
mundiales lo formalizan:

- **Living Documentation (BDD / Specification by Example).** La documentación funcional NO es prosa
  aparte: está vinculada al código y a pruebas ejecutables, de modo que si el código cambia y la
  doc no, algo falla y obliga a actualizar. La doc está viva porque está conectada a la fuente.
  *(experto: cucumber.io living documentation; Gojko Adzic, Specification by Example)*
- **Business Capability Map (TOGAF / Enterprise Architecture).** El «estado de cuenta» de lo que un
  sistema sabe hacer, por capability de negocio, independiente de cómo se programó — separado del
  backlog (transacciones) y del historial. *(oficial: TOGAF Business Architecture — Capabilities)*

Corolario operativo (doctrina propia, mandato del operador 2026-07-09): **el árbol
`docs/product/capabilities/` (una hoja YAML por capability + `INDEX.md`) es el SSoT de lo funcional.
Ningún cambio de código fuente existe sin construir o modificar un capability y actualizar el
registro.** La historia de usuario es el *delta*; el capability es el *saldo*.

## L2 · Realización (este árbol)

- `docs/product/capabilities/{module}/{slug}.yaml` (una hoja por capability) lista cada capability
  con `punteros:` al código autoritativo (`file#Símbolo`) y `valida:` su check. Estado
  (`vivo`/`sin-check`/`stub`) derivado del check. `INDEX.md` = índice generado (`cap_doctor.py --index`).
- **Integridad (R1):** cada puntero resuelve a un ARCHIVO real → no hay doc colgante. Honestidad
  del alcance: la parte `#Símbolo` del puntero HOY no se verifica (el enforcer hace stat del path
  recortado en `#`) — un símbolo renombrado pasa en silencio. Resolución de símbolo = deuda BACKLOG.
- **Cobertura (R2):** todo archivo fuente bajo `cmd/`, `internal/`, `web/src/`, `web/src-tauri/src/`
  está reclamado por ≥1 capability → no hay código huérfano. Allowlist explícito y con razón para
  generados, `*_test.go`, `embed_*.go`, boilerplate.
- **Gate de commit (R3):** un commit que toca fuente construye o modifica un capability en `docs/product/capabilities/` (lefthook pre-commit).
- **Estado consistente (R4):** el estado no CONTRADICE su evidencia — `vivo`/`parcial` ⟹ tiene
  `valida:`; `vivo·nc`/`stub` ⟹ sin `valida:` (enforcer determinista, `TestCapabilityStatusConsistent`).
  La derivación LIVE (`vivo ⟺ check verde` corriendo cada test) sigue como cableado CI (deuda BACKLOG).
- **Anti-drift:** el puntero es `file#Símbolo` o `paquete/`, NO `file:línea` cruda (las líneas se
  pudren) — ver `historias/2026-07-09-reorg-docs/doctrina-capabilities.md §3`.
- **Rollout:** este nodo nació `proposed` y graduó a `enforced` (cobertura=100%). Los 4 checks
  determinables corren como arch-test real (R1 `cap-ptr-resuelve` · R2 `cap-sin-huerfano` · R4
  `cap-estado-consistente` · R4 `cap-puntero-estable`); R3 = job lefthook local. Lo único que sigue
  como deuda honesta es la derivación LIVE del estado (correr cada check y flipear el bit) → CI. NUNCA
  pass fabricado.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| cap-ptr-resuelve | cada `punteros:` de `docs/product/capabilities/` resuelve a un ARCHIVO real del árbol (el `#Símbolo` aún no se verifica — deuda) | error | «capability apunta a código inexistente (doc colgante)» | docs/architecture/fitness/capability_trace_test.go:TestCapabilityPointersResolve |
| cap-sin-huerfano | todo archivo fuente (`cmd`/`internal`/`web/src`/`src-tauri/src`) está reclamado por ≥1 capability, salvo allowlist con razón | error | «archivo de código sin capability (huérfano)» | docs/architecture/fitness/capability_trace_test.go:TestCapabilityCoverage |
| cap-commit-toca-registro | un commit que modifica código fuente también construye o modifica un capability en `docs/product/capabilities/` | warn | «código cambiado sin actualizar el SSoT funcional» | lefthook pre-commit (capabilities) — feedback local, no bloquea CI |
| cap-estado-consistente | el `estado` no contradice su evidencia (`vivo`/`parcial`⟹tiene `valida:` · `vivo·nc`/`stub`⟹sin `valida:`); nunca fabricado a mano | warn | «estado de capability incoherente con su check» | docs/architecture/fitness/capability_trace_test.go:TestCapabilityStatusConsistent |
| cap-puntero-estable | los punteros usan `file#Símbolo`/`paquete/`, no `file:línea` cruda (anti-drift) | info | «puntero por número de línea (se pudre al reformatear)» | docs/architecture/fitness/capability_trace_test.go:TestCapabilityPointersStable |

## Changelog

- 2026-07-14 · v1.4 · Auditoría Portafolio: (a) los `enforced_by` apuntaban a
  `fitness/arch_test.go` pero los 4 tests viven en `fitness/capability_trace_test.go` —
  refs corregidas a la ruta real (el motor ahora resuelve rutas repo-relativas, cf.
  `portafolio-identidad-y-deriva-honesta` v1.1); (b) honestidad de R1: el enforcer valida
  el ARCHIVO del puntero, no el `#Símbolo` (stat del path recortado en `#`) — el texto
  decía «archivo/símbolo real» y `cap_doctor.py` afirmaba «go test resuelve el símbolo»,
  ambos overclaim; corregidos + deuda «R1 a nivel símbolo» registrada en BACKLOG.
- 2026-07-09 · v1.0 · Nodo fundacional (HS-18, paquete `reorg-docs`). L1 = Living Documentation
  (BDD) + Business Capability Map (TOGAF). L2 = `CAPABILITIES.md` SSoT + reglas R1-R4. Nació
  `proposed`: 5 checks diferían honesto (enforcer pendiente).
- 2026-07-09 · v1.1 · **Graduado a `enforced`** el mismo día: validador construido
  (`docs/architecture/fitness/capability_trace_test.go`) — **R1 `cap-ptr-resuelve`** (110 punteros canonicalizados
  a rutas repo-relativas, 0 colgantes) y **R2 `cap-sin-huerfano`** (154 archivos fuente: reclamados
  por caps + bloque `<!--coverage-->` + allowlist con razón; 0 huérfanos) PASAN como arch-test real.
  R3 `cap-commit-toca-registro` = job `capabilities` en `lefthook.yml` (feedback local). R4
  `cap-estado-generado` + `cap-puntero-estable` siguen `pendiente` (difieren honesto, no fabrican pass).
- 2026-07-09 · v1.3 · **R4 aterrizado (HS-20):** los 2 checks que quedaban `(pendiente)` pasan a
  arch-test determinista. `cap-estado-consistente` (`TestCapabilityStatusConsistent`) enforcea que el
  estado no contradiga su evidencia (vivo/parcial⟹tiene `valida:` · nc/stub⟹sin `valida:`) sobre las
  82 hojas; `cap-puntero-estable` (`TestCapabilityPointersStable`) rechaza punteros `file:línea`. Ambos
  verdes (0 incoherencias · 0 punteros por línea). Único resto: derivación LIVE del estado → CI.
- 2026-07-09 · v1.2 · **Sync a la homologación de metodología** (paquete
  `docs/product/stories/2026-07-09-homologacion-metodologia/`): el SSoT dejó de ser el monolito
  `CAPABILITIES.md` y pasó a ser el árbol `docs/product/capabilities/` (82 hojas YAML por-cap +
  `INDEX.md` generado); el enforcer R1/R2 (`capability_trace_test.go`) ya lee el árbol YAML desde
  Task 8. Sólo prosa: IDs de check y `enforced_by:` intactos (ruleset `--todo` sin cambio: 247·40·0).
