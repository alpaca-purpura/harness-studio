# docs/ — SSoT de ArnesIA (método del kit `harness@prenter-marketplace`)

> Toda la documentación que **no** es skill ni `CLAUDE.md` vive aquí, organizada como manda el
> plugin. Homologación firmada 2026-07-09 (paquete `historias/2026-07-09-homologacion-metodologia/`,
> D7 adopción plena). El mismo árbol se replica a cockpit y dev-studio.

## Tres zonas

| Zona | Qué contiene | Leyéndola sabés… |
|---|---|---|
| [`product/`](./product/) | qué existe HOY + pipeline de trabajo (capabilities, stories, releases, modules, checkpoint, BACKLOG, LEDGER) | **qué hace el sistema y cómo está construido** |
| [`architecture/`](./architecture/) | arquitectura-as-code + tecnologías + todo lo técnico (boundaries, fitness, contracts, conventions, model, stack) | **al detalle cómo está construido** |
| [`process/`](./process/) | doctrina de proceso del kit (lifecycle, gates, roles, mejora continua) | **cómo se construye software con arneses** |

## El seam

[`../project.config.yaml`](../project.config.yaml) = THE SEAM. Único store de valores TECH/SISTEMA
que el CORE del kit lee. Doctor: `python3 scripts/harness_config.py --doctor`.

## El front-door

`/pm` (`.claude/skills/pm/`) — orientación: dónde estamos en el proceso, próxima transición legal,
auto-chain a los demás arneses. Bootstrap: escanea stories abiertas → carga checkpoint/BACKLOG/releases.

## Convención dura

- **product = delta→saldo:** la story es el DELTA; la capability es el SALDO. Nada entra a
  `capabilities/` sin evidencia viva (`arnesia conformance` / test citado).
- **Nada suelto en `docs/` raíz de una zona** salvo su `README.md`. Contenido ad-hoc → subcarpeta.
- **Auto-generado no se teclea:** cifras de checkpoint/BACKLOG se GENERAN (`scripts/estado.sh`,
  `cap_doctor.py`). Editar la fuente, regenerar la vista.
