# devhub-ui — plataforma per-sistema (post CK-08)

> Hasta CK-08 (2026-07-02) esta UI era una plataforma compartida con Cockpit vía route-groups (`(sistema)`/`(directorio)`) + alias `@cockpit/*`. Stage 4 cortó limpio: Cockpit vive en `products/cockpit/ui` como app Next **independiente** (binario propio `directorio`, puerto 4100, `node_modules` propio). Esta UI es 100% lente sistema — sin route-group `(directorio)`, sin alias, sin symlink, sin import cruzado. Historia completa de la extracción: `products/cockpit/LEDGER.md` (fichas `CK-01`..`CK-08`).

Vistas del proceso SDD de **UN install**: board, roadmap, map, arquitectura, drift, learnings, harness, evolución (lente Ledger por-sistema, I-45). Toda vista lee data **per-sistema** (`?sistema=`).

## Mapa de carpetas
- **`app/`** — board · roadmap · map · arquitectura (redirect legacy → /map) · drift · learnings · harness · evolucion. `layout.tsx` = raíz (html/body + `SistemaProvider`) → `AppShell`. `app/page.tsx` — redirect a /roadmap.
- **`app/api/`** — se stashea entero en `build-ui.sh` para el static export; el Go implementa la paridad. Todos per-sistema (`stories`, `capabilities`, `releases`, `system-map`, `learnings`, `harness`, `cil`, …) — `/api/portfolio`+`/api/negocio` (directorio) ya no existen acá, viven en `products/cockpit/go`.
- **`components/sistema/`** — vistas + piezas del proceso SDD: board, roadmap, map, drift, evolucion, harness, learnings, modals (release), story-drawer, cap-drawer.
- **`components/{layout,providers,ui,platform}/`** — shell y primitivas devhub-owned. `AppShell` monta drawers sistema. `SistemaProvider` = SSoT de navegación de la plataforma + switcher multi-empresa propio (`SistemaSwitcher`, sin `EmpresaDependencias` — borrado en CK-08, sin procedencia que mostrar sin la vista Negocio).
- **`lib/`** (plano) — infraestructura compartida/per-sistema, consumida también por `app/api/` y el shell (types, api-client, fs-reader, nav, workspace, project-config, `sistemas.ts`, …). `sistemas.ts` (CK-08) = copia literal del subconjunto de `products/cockpit/ui/lib/portfolio.ts` que nunca dependió del árbol rico Empresa→Sistema — agrupa las sistema-keys planas que sirve `/api/sistemas` (devhub-owned desde siempre). Duplicación consciente, cero import cruzado.

## Disciplina
- Build/embebido en el binario: `../go/build.sh` (llama `build-ui.sh`). Tests: `pnpm test` (vitest, `lib/**`). Typecheck: `pnpm typecheck`.
- Cero dependencia de `products/cockpit/*` — sin alias `@cockpit/*`, sin symlink de `node_modules`, sin route-group compartido. Módulo que ambas células necesiten → cada una guarda su propia copia (mismo criterio que `directorioMeta` en `go/registry.go`), no se importa cruzado.
- Contrato de datos DevHub→Cockpit: diseñado en CK-08 (Pull API, envelope versionado `GET /api/contrato/directorio?sistema=...`), sin implementar — sin consumidor real todavía. No escribir el cliente/stub antes de que exista una vista que lo necesite.
- Rename comercial del binario/CLI (`cockpit` → nombre DevHub) sigue pendiente — DH-01.
