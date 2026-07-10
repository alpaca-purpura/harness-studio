# docs/process/ — doctrina de proceso (del kit)

> Owner: kit `harness@prenter-marketplace` (upstream, backflow). El CÓMO se construye software con
> arneses: lifecycle idea→done, gates, roles, DoD, safety, mejora continua. Tech/sistema-agnóstico.

## Fuente

La doctrina de proceso la **ship el plugin** (no se forkea aquí):
`~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.3/process/` +
`.../rules/` (21 invariantes always-on) + `.../templates/`.

| Doc del kit | Qué |
|---|---|
| `harness-lifecycle.md` | HLP — meta-proceso de mantenimiento del kit |
| `ticket-states.md` | máquina de 12 estados de ticket |
| `spec-mapa-funcional.md` | capa spec SDD: mapa funcional + Gherkin + matriz de cobertura |
| `continuous-improvement.md` | CIL 4-carril (incl. L4 capability-desfasada → `cap_doctor`) |
| `tech-debt.md` | registro L3 |
| `cockpit-permissions.md` | ownership de transiciones (humano vs agente) |

## Estados macro (10) — value_stream del seam

`idea → refining → refined → ready → developing → developed → reviewing → done` (+ `parked`/`dropped`).
Definidos en `project.config.yaml:value_stream`. La skill `/pm` los conduce.

## Local (capa proyecto)

Lo específico de harness-studio que extiende la doctrina de proceso vive acá; su parte genérica se
propone upstream al kit, no se forkea.

| Doc local | Qué |
|---|---|
| [`metodologia.md`](./metodologia.md) | doctrina operativa de ArnesIA: fábrica de cajas de proceso, **contrato de caja §3**, **honestidad §4**, **disciplina de paquete de trabajo §10**, arquetipos §8. Deriva de `docs/architecture/knowledge/`; el `contract:` de caja lo formaliza `docs/architecture/contracts/schema/box.contract.schema.json`. |

> Nota: `metodologia.md` mezcla reglas-de-negocio del dominio + método; su split por-concern es
> paquete aparte (BACKLOG `[reorg-docs]`). Aquí se reubicó (homologación 2026-07-09, D8), no se partió.
