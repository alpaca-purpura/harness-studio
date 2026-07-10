# docs/product/ — qué existe + pipeline de trabajo

> SSoT funcional. Owner: `/pm`. Modelo del kit: **Release → Story → Capability → Scenario**.
> La story es el DELTA; la capability es el SALDO. Los 3 ejes temporales viven acá (D3).

## Contenido

| Path | Qué | Eje | Owner |
|---|---|---|---|
| `vision.md` | constitución (11 principios + anatomía A1–A7) | — | Chris |
| `ux.md` | UX firmada del producto (superficie · inventario · log de iteraciones) | — | Chris + `/po-ux` |
| `checkpoint.md` | estado global HOY (fase + paquete activo + cifras vivas) | **presente** | `/pm` (cifras GENERADAS) |
| `BACKLOG.md` | lo-que-viene (solo items abiertos; cerrado se borra) | **futuro** | `/pm` |
| `LEDGER.md` + `ledger/HS-NN.md` | decisiones firmadas (historia) | **pasado** | `/pm` |
| `releases/{id}.yaml` | hitos (contenedor temporal de stories) | futuro | `/pm` |
| `stories/{id}/` | unidades de trabajo (00-research…07-merge + checkpoint) | presente | `/pm` + handoffs |
| `capabilities/{module}/{slug}.yaml` | qué existe (SSoT, YAML por-cap + BDD) | saldo | `/pm` ratifica al merge |
| `modules/{module}.md` | narrativa por módulo de negocio | — | `/pm` |
| `research/` | research cross-cutting; caps y stories citan su fuente aquí | — | Chris + `/pm` |
| `archive/{año}/stories/` | stories `done` (snapshot inmutable) | pasado | auto al merge |
| `_templates/` | plantillas del kit (00-story, 00-research, 01-spec, story.yaml, capability) | — | kit (upstream) |

## Reglas de escritura (heredadas del kit · pm-vitalia R1/R2/R3)

- **R1** — nada suelto en `product/` raíz salvo `README.md`, las 4 hojas de eje
  (`vision`/`checkpoint`/`BACKLOG`/`LEDGER`) y `ux.md` (superficie firmada). El resto va a su subcarpeta.
- **R2** — story `state: done` se auto-mueve a `archive/{año}/stories/` en el commit de cierre.
- **R3** — auto-generados NO se editan a mano (`checkpoint.md` cifras, tablas de `modules/`).
  Editar la fuente (capabilities/stories/conformance), regenerar la vista.

## Capabilities — enforcement (`cap_doctor.py` + arch-test)

Cada `capabilities/{module}/{slug}.yaml` cumple `docs/architecture/boundaries/codigo-traza-a-capability.md`:
**R1** punteros `file#Símbolo` resuelven · **R2** todo archivo de código reclamado por ≥1 cap ·
**R3** commit que toca código toca su cap · **R4** `status` GENERADO, no tecleado.
