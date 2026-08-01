# Informe de estado real — insumo de research (2026-08-01)

> Auditoría con evidencia archivo:línea, hecha el mismo día de las decisiones. Dos frentes:
> **(1)** anatomía de `~/Proyectos/vitalia-app` como espécimen del escenario B, **(2)** qué
> maquinaria existe hoy en arnesia para ese escenario. Honesto: dice lo que NO existe.

## 0. Disparador

El operador no veía las «actividades» en la app instalada. Diagnóstico: versión correcta
(0.7.0, binario fresco, feature completa E2E en el binario — probado con `arnesia index` sobre
el canónico), pero **ningún arnés indexado tenía `arnes.yaml` en su raíz** — la derivación es
silenciosa-legal cuando falta (`internal/adapters/loader/actividades.go:64`). El canónico
`prenter-marketplace/plugins/developer-vitalia/0.1.0/` SÍ lo tiene; la instalación no lo
recibió jamás. Fix manual aplicado (copia a `~/Proyectos/vitalia-app/`), chips vivos. El fix
sistemático = T0 (ING-D9).

## 1. Espécimen: anatomía de vitalia-app

- **169 archivos as-code tracked en `.claude/`**: 30 skills · 11 agents · 51 rules · 10
  commands · 4 hooks · 1 workflow JS · settings.
- **Cero plugins instalados para este path** (`installed_plugins.json` no lo contiene). El
  `CLAUDE.md:3` lo declara: harness **materializado** (archivos propios, sin plugin). El
  checkout upstream vive en `~/.arnesia/checkouts/prenter-marketplace/developer-vitalia/`.
- **Procedencia mixta medida contra el checkout**: 11/11 agents + 10/10 commands + 4/4 hooks +
  49/51 rules idénticos (provisto-por-plugin) · **4 skills con DRIFT local**
  (`architect`/`auditor`/`commit-push`/`dev-team`) · **15 skills creados-por-usuario** — y la
  mitad propia ES la cadena de proceso (`/pm-vitalia → /po|/po-ux → /architect → /dev-team →
  /auditor`) + 2 rules propias. Hooks re-cableados local (`${CLAUDE_PROJECT_DIR}` + un `Stop`
  hook que el plugin no tiene).
- **Capas de terceros**: `skills-lock.json` (12 skills Clerk, `computedHash`) materializadas en
  `.agents/skills/` FUERA de `.claude/` — de-referencia, jamás propio.
- **`/pm-vitalia` = el orientador** (396 líneas): tabla de superficies owner→path · máquina de
  10 estados con WIP caps · bootstrap 2 pasos (closure-gate scan + carga checkpoint/BACKLOG) ·
  prior-art scan · tabla «Chris dice → acción» (17 filas, auto-chain a skills) · gates de
  cierre · 24 punteros salientes — el nodo más conectado del grafo.
- **El grafo de punteros YA está materializado en el proyecto**: `docs/process/DOCS-GRAPH.md`
  (autogen, 920 docs, clases SSoT-vivo/detail-pareado) + `_docs-graph.json` (aristas
  `{src,dst,kind}`) + `scan_harness_pointers.py` anti-rot. ArnesIA no inventa el grafo — lo lee.
- **WIP-home nítido**: `vitalia/docs/product/stories/{id}/` cuya carpeta crece con el estado
  (2 archivos en `idea` → 46 en `developing`) → `archive/2026/stories/` al cerrar (49
  archivadas). Frontmatter schema v2 (~25 campos). Templates en `docs/specs/templates/`.
- **3 desalineaciones sello↔realidad** (base de ING-D6):
  1. Spine declarado (5 estados del checkout / 4 tipos del `arnes.yaml`) ≠ spine operado (los
     34 checkpoints usan SOLO los 10 estados macro de `/pm-vitalia`).
  2. `vitalia/arnes.l0.json` declara spine fantasma (`cambio-propuesto → cambio-verificado`)
     sin contraparte en ningún artefacto.
  3. WIP caps en 3 sitios con valores distintos (SKILL.md ≤1 · CLAUDE.md ≤3 ·
     `project.config.yaml` 3/10/2).
- El l0 de la raíz es **subset degradado** del upstream (perdió `version`, `fases`, `spine`) →
  `deriva-no-evaluable` en `~/.arnesia/portafolio.json`.

## 2. Maquinaria existente en arnesia — mapa de gaps

| Pieza del escenario B | Estado (evidencia) |
|---|---|
| Detectar instalaciones | ✅ 4 detectores + monorepo depth 4 (`scanner.go:46-92`): `.claude/` · materializada · lock DevStudio · referenciada-CC |
| Carpeta cruda sin nada | ❌ `[]` sin error → wizard: «este asistente no la reconoce» (`portafolio-wizard.tsx:308`) — callejón sin salida |
| Inventario skills/rules/commands/hooks/mcp/settings | ✅ loader (`loader.go:311`, `regla.go:25`, `soporte.go`); no-entendido → nodo `no-reconocido` visible |
| Inventario **agents** | ❌ reconocedor inexistente — TODO declarado (`loader.go:20-24`, `nomenclatura-arnes.md:68`) |
| Contenedor `plugin` como nodo | ❌ inexistente (misma nota) |
| Clasificación knowledge/as-code/WIP | ❌ cero código; vocabulario solo en prosa. `domain.Clase` = 10 primitivas técnicas, OTRO eje |
| «Identificar» (S1-D28) | ✅ sella identidad in-situ (l0 validado contra schema + plugin.json mínimo + re-key); NO inventaría/clasifica/arma actividades. «Identificar actividades» = futuro nombrado (spec E8, definicion-de-arnes) |
| `arnesia init` + doctor | ✅ siembra `.arnesia/` (19 piezas, semilla EMBEBIDA) en cualquier path no-protegido; ❌ no lee el proyecto — `--arnes-yaml` = gap Fase 2 declarado (`parser.go:6-7`); sin FE (`sembrar-semilla.yaml` route:null) |
| Gate de completitud D19 | ❌ el bloque `gate:` de `semilla/arnes.yaml` no lo lee nadie |
| Forja conversacional | ❌ `forja-ciclo-vivo` ⏸ pausada desde 2026-07-10; `internal/adapters/forja/` = solo sembrador (parser 161 + scaffolder 222 + doctor 67 líneas) |
| Editar desde el Mapa | ❌ read-only TOTAL: rutas del grafo solo GET (`router.go:86-97`); «Editar fuente» disabled hardcodeado (`inspector.tsx:744`); CTAs «conversando» solo abren dock (vacío el de E7, `actividad-chips.tsx:135`) |
| Actividades desde `arnes.yaml` | ✅ seam cerrado (MA-T1b, `actividades.go:64`) — derivación al indexar, ausente = silencioso-legal |
| Manos del conductor (permisos Write) | ✅ **resuelta v0.7.0** (`4d14e75`, deuda D) — informes anteriores a esa fecha están stale en esta fila |
| Publicar (B2) | ✅ copia árbol completo del canónico incl. `arnes.yaml` (`publisher.go:306-352`; excluye solo `.git/`, `.in_use`, `.orphaned_at`) + conformance gate + tag |
| Instalación recibe `arnes.yaml` | ❌ **gap T0**: nadie lo propaga a instalaciones (§0) |
| Refrescar | ✅ `git pull --ff-only` del checkout, solo marketplace `propio`, degradado visible (CAP-154 parcial); deuda E-ter: `catalogo.json` mono-plugin |
| Alta plugin forjado → canónico | ✅ `Adoptar` (`adoptar.go:54`, CAP-155 parcial) |
| Deriva semilla vs lock | ❌ lock se escribe, no se evalúa (gap Fase 2) |

## 3. La frase que resume el gap

Del informe de la única forja real
(`stories/2026-07-30-definicion-de-arnes/informe-forja-conversacional.md`):

> «el criterio y la verificación fueron conversacionales; **la materialización fue mía porque
> el producto me negó las manos del conductor**.»

## 4. Contexto externo (2026)

La industria converge en lo mismo que este paquete: skills de agentes estructurados como
knowledge graph con ontología + quality gates («graph engineering», «context graphs») en vez de
.md sueltos. El modelo de terreno D0-D20 + regla de las 3 caras ya es esa ontología, propia.
Referencias: cognee.ai/blog/tutorials/structure-your-skills-with-cognee ·
atlan.com/know/ai-agent/knowledge-graph-for-ai-agents ·
github.com/codejunkie99/graph-engineering.
