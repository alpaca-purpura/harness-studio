# arch/ — arquitectura y diseño técnico as code de ArnesIA

> **La arquitectura vive aquí, no en un doc que se pudre.** Cada regla arquitectónica
> estructural (boundary) = un nodo con **L1** (principio con fuente) ↔ **L2** (realización en
> este árbol Go) + una **tabla de checks ejecutables**. La parte «por qué» vive en el
> [`../LEDGER.md`](../LEDGER.md) (fichas firmadas, sin duplicar); la parte «prueba» vive en
> [`fitness/`](./fitness/); el «cómo se ve» en [`model/`](./model/); el «contrato de datos» en
> [`contracts/`](./contracts/). Espeja el patrón de [`../knowledge/`](../knowledge/INDEX.md).
> Norte: [`../VISION.md`](../VISION.md) · evidencia L1:
> [`../research/2026-07-05-arquitectura-fase3.md`](../research/2026-07-05-arquitectura-fase3.md).

## Cómo se relaciona con el resto del repo

- [`../VISION.md`](../VISION.md) — constitución (11 principios + anatomía A1–A7) y **decisiones
  técnicas fundacionales** (HS-02). El **norte**; esta capa las aterriza y las enforça.
- [`../METODOLOGIA.md`](../METODOLOGIA.md) — reglas de negocio. El `contract:` de caja (§3) es
  **el mismo schema** que valida [`contracts/schema/box.contract.schema.json`](./contracts/schema/box.contract.schema.json).
- [`../knowledge/`](../knowledge/INDEX.md) — estándar as code por **elemento de arnés** (121
  checks). `arch/` es el gemelo: estándar as code de **la app ArnesIA misma** (dogfood). Un solo
  runner (`arnesia conformance`) corre ambos árboles.
- [`../LEDGER.md`](../LEDGER.md) — el diario firmado. Cada boundary cita su ficha en `ledger:`.

## Las dos capas (en cada boundary node)

1. **L1 · Principio** — el patrón de arquitectura como estándar de industria (hexagonal,
   ports&adapters, local-first, event-sourcing…), fechado y con fuente. Evidencia, no opinión.
2. **L2 · Realización** — cómo se mapea en ESTE árbol Go+React: qué paquetes son core/shell/
   dominio/transporte/adaptador, y cómo la regla se encarna aquí.

Y cada nodo emite un **`Checklist evaluable`**: la rúbrica as code (checks con severidad + señal)
que un linter (`go-arch-lint`/`depguard`/validación de schema/`arch_test.go`) corre para probar
que el código NO viola la arquitectura. Cada check declara su `enforced_by:` — el link 1:1 a la
check ejecutable que lo guarda.

## El árbol — boundary nodes

| Nodo | Regla | Estado | Versión | Checks | Enforcer |
|------|-------|--------|---------|--------|----------|
| [`boundaries/core-no-importa-shell.md`](./boundaries/core-no-importa-shell.md) | El daemon-core no depende del shell (Tauri) | 🌱 vivo | 1.0 | 4 | go-arch-lint · depguard |
| [`boundaries/dominio-independiente-de-transporte.md`](./boundaries/dominio-independiente-de-transporte.md) | El dominio no depende de HTTP/SSE/SQLite | 🌱 vivo | 1.0 | 4 | go-arch-lint · depguard |
| [`boundaries/adaptadores-de-agente-intercambiables.md`](./boundaries/adaptadores-de-agente-intercambiables.md) | Claude Code = un adaptador tras `AgentPort` | 🌱 vivo | 1.0 | 4 | go-arch-lint · arch_test.go |
| [`boundaries/indice-desechable-jsonl-es-verdad.md`](./boundaries/indice-desechable-jsonl-es-verdad.md) | JSONL = verdad; SQLite = índice reconstruible | 🌱 vivo | 1.0 | 4 | arch_test.go · schema |
| [`boundaries/conductor-no-parsea-jsonl.md`](./boundaries/conductor-no-parsea-jsonl.md) | El conductor consume stream-json/OTel, no parsea JSONL | 🌱 vivo | 1.0 | 4 | depguard · arch_test.go |
| [`boundaries/permisos-gui-human-in-the-loop.md`](./boundaries/permisos-gui-human-in-the-loop.md) | Deny-by-default; el GUI aprueba cada write vía diff | 🌱 vivo | 1.0 | 5 | arch_test.go |
| [`boundaries/contrato-de-caja-es-fitness-function.md`](./boundaries/contrato-de-caja-es-fitness-function.md) | Validar el `contract:` de caja contra su schema | 🌱 vivo | 1.0 | 4 | schema · arch_test.go |

Leyenda de estado: ⏳ en forja · 🌱 vivo (nace, se enforça cuando el código llegue) · 🌳 estable ·
🔍 en-revisión. **Total: 7 boundaries · 29 checks · pasada fundacional 2026-07-05 (HS-04).**

> **Honestidad (heredada de METODOLOGIA §4 / CADENCE):** hoy no hay código Go todavía (fase 5).
> Los checks están **declarados, no corriendo**: son el ruleset que se activa cuando el módulo
> `arnesia` aterrice. Igual que los 121 checks de `knowledge/` son el linter futuro, estos 29 son
> el enforcement futuro de la arquitectura. Estado del enforcer = `proposed` hasta que el código
> exista; luego `enforced`.

## Subdirectorios

- [`model/`](./model/) — **diagramas as code** (render on-demand, nunca hand-sync). C4 nivel
  contexto (`system-context.mmd`, Mermaid) + container (`container.d2`, D2 = binario Go).
- [`contracts/`](./contracts/) — **schema-first, single source of truth**. `schema/` = las 2
  schemas del dominio (L0 `meta.clase` + `contract:` de caja); `api/openapi.yaml` = superficie
  HTTP/SSE. `gen/` = tipos generados Go+TS (quicktype/oapi-codegen; se llena al haber build).
- [`fitness/`](./fitness/) — **las checks ejecutables** (fallan CI): `.go-arch-lint.yml` (grafo
  de imports), `arch_test.go` (lo que el linter no expresa).
- `decisions/` — MADR completos, **solo cuando hace falta** el tratamiento de opciones que una
  ficha no aguanta. Vacío por ahora (el LEDGER cubre el «por qué»).

## El stack (resumen; detalle y fuentes en el research doc y los nodos)

| Capa | Pick |
|---|---|
| Shell v1 | **Tauri 2** (daemon = sidecar `externalBin`); ruta a «app vendible» aditiva |
| Backend | binario Go único `arnesia` (serve/open/index/publish) |
| Conexión CC | subproceso `claude` + **stream-json** por stdin/stdout (patrón conductor) |
| Índice | **modernc.org/sqlite** (puro-Go, WAL, desechable) · JSONL = verdad |
| Transporte | **SSE** multiplexado (`event: map\|dock\|run`) |
| Frontend | Vite+React SPA `go:embed` · **React Flow 12** · **Zustand** + hash-state |
| Dock | **AG-UI** (taxonomía, emisor Go propio) · **assistant-ui** · **CodeMirror 6** + merge |
| Arch as code | JSON Schema 2020-12 → quicktype (Go+TS) · go-arch-lint + depguard · D2+Mermaid · MADR-proyección del LEDGER |

## Cómo crece

La arquitectura se revisa **al cambiar** (no en cadencia fija como `knowledge/`): cada decisión
estructural nueva = una ficha `HS-NN` + (si crea/cambia un boundary) un nodo aquí con su check.
Detalle en [`CADENCE.md`](./CADENCE.md).
