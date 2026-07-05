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
- [`../knowledge/`](../knowledge/INDEX.md) — estándar as code por **elemento de arnés** (122
  checks). `arch/` es el gemelo: estándar as code de **la app ArnesIA misma** (dogfood). Un solo
  runner (`arnesia conformance`) corre ambos árboles — **aspiración pendiente-HS-06**: hoy `arch/`
  ya declara `enforced_by:` por check, pero los checks de `knowledge/` aún no; la ejecución
  conjunta requiere un **contrato de check común** (schema con `enforced_by`/mecanismo por check)
  planificado como ficha HS-06. Hasta entonces cada árbol corre por separado; la aspiración sigue
  viva, solo no es realidad actual.
- [`../LEDGER.md`](../LEDGER.md) — el diario firmado. Cada boundary cita su ficha en `ledger:`.

## Las dos capas (en cada boundary node)

1. **L1 · Principio** — el patrón de arquitectura como estándar de industria (hexagonal,
   ports&adapters, local-first, event-sourcing…), fechado y con fuente. Evidencia, no opinión.
2. **L2 · Realización** — cómo se mapea en ESTE árbol Go+React: qué paquetes son core/shell/
   dominio/transporte/adaptador, y cómo la regla se encarna aquí.

Y cada nodo emite un **`Checklist evaluable`**: la rúbrica as code (checks con severidad + señal)
que un linter (`go-arch-lint`/`depguard`/validación de schema/`arch_test.go`) corre para probar
que el código NO viola la arquitectura. Cada check declara su `enforced_by:` — el link 1:1 a la
check ejecutable que lo guarda (esto ya rige en `arch/`; en `knowledge/` el `enforced_by:` por
check llega con HS-06, ver runner unificado arriba).

## El árbol — boundary nodes

| Nodo | Regla | Estado | Versión | Checks | Enforcer |
|------|-------|--------|---------|--------|----------|
| [`boundaries/core-no-importa-shell.md`](./boundaries/core-no-importa-shell.md) | El daemon-core no depende del shell (Tauri) | 🌱 vivo | 1.0 | 4 | go-arch-lint · arch_test.go |
| [`boundaries/dominio-independiente-de-transporte.md`](./boundaries/dominio-independiente-de-transporte.md) | El dominio no depende de HTTP/SSE/SQLite | 🌱 vivo | 1.0 | 4 | go-arch-lint · depguard |
| [`boundaries/adaptadores-de-agente-intercambiables.md`](./boundaries/adaptadores-de-agente-intercambiables.md) | Claude Code = un adaptador tras `AgentPort` | 🌱 vivo | 1.0 | 4 | go-arch-lint · arch_test.go |
| [`boundaries/indice-desechable-jsonl-es-verdad.md`](./boundaries/indice-desechable-jsonl-es-verdad.md) | JSONL = verdad; SQLite = índice reconstruible | 🌱 vivo | 1.0 | 4 | arch_test.go · schema |
| [`boundaries/conductor-no-parsea-jsonl.md`](./boundaries/conductor-no-parsea-jsonl.md) | El conductor consume stream-json/OTel, no parsea JSONL | 🌱 vivo | 1.0 | 4 | depguard · arch_test.go |
| [`boundaries/permisos-gui-human-in-the-loop.md`](./boundaries/permisos-gui-human-in-the-loop.md) | Deny-by-default; el GUI aprueba cada write vía diff | 🌱 vivo | 1.0 | 5 | arch_test.go |
| [`boundaries/contrato-de-caja-es-fitness-function.md`](./boundaries/contrato-de-caja-es-fitness-function.md) | Validar el `contract:` de caja contra su schema | 🌱 vivo | 1.0 | 4 | schema · arch_test.go |
| [`boundaries/fe-topologia-fsd.md`](./boundaries/fe-topologia-fsd.md) | La SPA se estructura por FSD (import direccional) | 🌱 vivo | 1.0 | 5 | dependency-cruiser · steiger |
| [`boundaries/fe-taxonomia-componentes.md`](./boundaries/fe-taxonomia-componentes.md) | Taxonomía por dirección/pureza, no ladder atómico (canvas⊥chrome) | 🌱 vivo | 1.0 | 4 | dependency-cruiser |
| [`boundaries/fe-transporte-independiente.md`](./boundaries/fe-transporte-independiente.md) | Dominio FE ⊥ transporte; SSE singleton en `app` | 🌱 vivo | 1.0 | 4 | dependency-cruiser |
| [`boundaries/fe-tokens-contrato.md`](./boundaries/fe-tokens-contrato.md) | Tokens DTCG = contrato mockup↔código, cero magic-value | 🌱 vivo | 1.0 | 4 | stylelint · tokens-sync |
| [`boundaries/fe-visual-fitness.md`](./boundaries/fe-visual-fitness.md) | Story = test que rompe CI (fitness visual local) | 🌱 vivo | 1.0 | 5 | vitest + Storybook 10 |

Leyenda de estado: ⏳ en forja · 🌱 vivo (nace, se enforça cuando el código llegue) · 🌳 estable ·
🔍 en-revisión. **Total boundaries: 12 · 51 checks** — fundacional HS-04 (backend, 7 boundaries · 29
checks) + HS-05 (frontend, 5 boundaries · 22 checks). **+ [`conventions/`](./conventions/INDEX.md): 8
convention nodes · 26 checks** (HS-05). **Gran total `arch/`: 77 checks.**

> **Honestidad (heredada de METODOLOGIA §4 / CADENCE):** hoy no hay código todavía (fase 5). Los
> checks están **declarados, no corriendo**: son el ruleset que se activa cuando el módulo `arnesia`
> y la SPA `web/` aterricen. Igual que los 122 checks de `knowledge/` son el linter futuro, estos 77
> (51 checks de boundary + 26 de conventions) son el enforcement futuro de la arquitectura. Los config files
> (`.golangci.yml`, `web/biome.json`, `web/.dependency-cruiser.js`, `lefthook.yml`, `ci.yml`…) están
> **declarados** con `if: hashFiles(...)` / notas de honestidad; los paths (`web/src/**`, module path)
> son **provisionales**. Estado del enforcer = `proposed` hasta que el código exista; luego `enforced`.

## Subdirectorios

- [`model/`](./model/) — **diagramas as code** (render on-demand, nunca hand-sync). C4 nivel
  contexto (`system-context.mmd`, Mermaid) + container (`container.d2`, D2 = binario Go).
- [`contracts/`](./contracts/) — **schema-first, single source of truth**. `schema/` = las 2
  schemas del dominio (L0 `meta.clase` + `contract:` de caja); `api/openapi.yaml` = superficie
  HTTP/SSE. `gen/` = tipos generados Go+TS (quicktype/oapi-codegen; se llena al haber build).
- [`fitness/`](./fitness/) — **las checks ejecutables** (fallan CI): `.go-arch-lint.yml` (grafo
  de imports Go), `arch_test.go` (lo que el linter no expresa). Los enforcers FE viven junto a la SPA
  (`web/.dependency-cruiser.js`, `web/steiger.config.ts`, `web/.stylelintrc.json`) y los de estilo en
  la raíz (`.golangci.yml`, `web/biome.json`, `lefthook.yml`, `.github/workflows/ci.yml`).
- [`conventions/`](./conventions/INDEX.md) — **convenciones de código as code** (HS-05): 8 nodes · 26
  checks de estilo/lint/format/naming/commits/hooks/CI (Go + TS/React + Rust). Hermano de `fitness/`
  (`fitness/` = forma arquitectónica; `conventions/` = estilo de código). Cada nodo referencia el config
  real vía `enforced_by:`.
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
| FE topología (HS-05) | **FSD-lite** (`web/src/{app,pages,widgets,features,entities,shared}`; `pages`=composition-roots por hash-state, sin router) · **dependency-cruiser** (gate) + steiger |
| FE componentes (HS-05) | **shadcn/ui sobre Base UI** (copy-in, lintable) · 6 capas de UI direccionales (**canvas⊥chrome**) · atomic = vocabulario en `shared/ui` |
| Tokens + estilo (HS-05) | **DTCG 2025.10** `.tokens.json` → **Style Dictionary v5** → **Tailwind v4** `@theme`+TS · dark `data-theme` · stylelint anti-magic-value |
| Fitness/convenciones FE (HS-05) | **Storybook 10** (story=test) + a11y axe + regresión visual **local** (sin Chromatic) · **Biome v2.4** + `tsc` strictest · **golangci-lint v2** · **lefthook** |

## Cómo crece

La arquitectura se revisa **al cambiar** (no en cadencia fija como `knowledge/`): cada decisión
estructural nueva = una ficha `HS-NN` + (si crea/cambia un boundary) un nodo aquí con su check.
Detalle en [`CADENCE.md`](./CADENCE.md).
