# Decisiones — Homologación de metodología (3 repos)

> Paquete de trabajo. Objetivo: homologar la forma de crear software (estructura de
> CLAUDE.md, skills, docs) entre **harness-studio**, **cockpit** y **dev-studio**, usando
> **luana-vitalia/vitalia** como referencia. Regla de esta sesión: estudiar lo mejor de cada
> uno y **construirlo bien AQUÍ (harness-studio) primero**; replicar a los otros repos en otra
> conversación. Permiso dado para tocar los otros repos, pero no en esta sesión.

## Norte pedido por el operador (cita cruda)

- Todo doc que **no** sea skill ni CLAUDE va a una sola carpeta `docs/`:
  - `docs/product/` — producto actual, capabilities-as-code; leyéndolo se sabe **qué existe y
    cómo está construido**.
  - `docs/architecture/` — arquitectura-as-code + tecnologías + todo lo técnico; leyéndolo se
    sabe **al detalle cómo está construido**.
  - `docs/projects/` — lo que se quiere desarrollar **+ el historial de lo desarrollado**:
    - `epics/` — hitos grandes, rotos en varias historias hasta implementar todo.
    - `stories/` — historias de usuario en carpetas, con documentos **obligatorios** antes de
      desarrollar (spec, arquitectura, diseño, etc.).
    - `research/` — cualquier research; stories y capabilities apuntan aquí a su fuente original.
- Un skill **`/pm`** que conozca todo el proceso, para invocarlo y ubicarse; + otros "arneses"
  que permitan desarrollo ordenado.

## Exploración — lo mejor de cada repo (2026-07-09)

Barrido de 4 agentes en paralelo (resultados completos en la conversación).

- **harness-studio:** CLAUDE.md = router puro + tabla "Necesito X → leo Y"; modelo **4 ejes**
  (LEDGER=pasado · BACKLOG=futuro · ESTADO=presente · CAPABILITIES=saldo); **CAPABILITIES = SSoT
  funcional** con enforcement R1/R2 por arch-tests (`capability_trace_test.go`, punteros
  `file#Símbolo` anti-drift, estado generado); nodos **L1(principio+fuente)/L2(realización) +
  `Checklist evaluable` + `enforced_by:`** idénticos en `arch/` y `knowledge/`; disciplina de
  paquete de trabajo mockup→decisiones→spec→PARIDAD con firmas 🧑‍⚖️; **honestidad**
  (cifras generadas, gaps visibles, sin pass fabricado, drift documentado). Peso: `UX.md` 60k y
  `METODOLOGIA.md` 36k monolíticos; cifras aún tecleadas en prosa (drift recurrente).
- **cockpit:** tríada `sistema/`+`docs/`+`proyecto/`; patrón **SSoT-yaml + vista-md generada,
  editadas mismo evento**; **pre-commit que valida Y regenera** la arquitectura (anti-drift
  mecánico); ledger `CK-NN` como decisión vinculante citada desde cada artefacto.
- **dev-studio:** dos planos **PRODUCTO vs PROYECTO** nunca mezclados; caps con **evidencia viva
  citada** (nunca "compila"/"200"); arch fitness tests + doctrina honestidad (`enforced` exige
  test que pasa hoy, gap real = `❌ TBD`, jamás verde pintado); **`project.config.yaml` = "THE
  SEAM"** (DIP, sentinelas `__FILL_ME__`, `--doctor`); **`.claude/rules/`** (21 guardrails:
  tdd-mandatory, story-closure-gate, anti-duplication, worktree…); **`templates/`** 00→07;
  `docs/process/` (lifecycle, ciclo mejora continua, state machines).
- **luana-vitalia/vitalia (REFERENCIA):** **modelo 4-ejes SDD Release→Story→Capability→Scenario**;
  **capability = YAML vivo** (esquema v3.2: dev_preview + scenarios BDD como unidad atómica +
  change-log + grafo de deps + business_rules); **`SYSTEM-MAP.yaml`** = arquitectura/taxonomía
  as-code renderizable; pipeline story numerado **00-research → 01-spec → 02-design → 03-arch →
  04-validators → 05-guidelines → 06-tickets → 07-merge** con librería de templates compartida;
  **`/pm-{brand}`** = front-door (bootstrap: closure-gate scan → carga checkpoint/backlog/releases
  → menú; state machine 10 estados con WIP caps; intake-handshake; **auto-chain programático** a
  po-ux/architect/dev-team/auditor); enforcement determinista `make new-cap` + `make cap-doctor`
  (G1-G6) + `# cap:` headers → **drift falla el commit**; **live-verify = DoD** (`dod_evidence[]`
  con HTTP/DB/logs reales). Layout: harness (skills/agents/commands/rules) en **root `.claude/`**
  compartido; docs SSoT en **`<proj>/docs/`** por marca; CLAUDE.md = overlay puntero.

**Veredicto de diseño:** vitalia es prácticamente el target ya construido. Homologar =
layout vitalia (`docs/product` · `docs/architecture` · `docs/projects/{epics,stories,research}`)
+ router/honestidad/enforcement-1:1 de harness-studio + gate-mecánico de cockpit +
evidencia-DoD/rules/templates de dev-studio.

**Tensión propia de harness-studio:** es la fábrica de arneses; `arch/`, `knowledge/` y
`CAPABILITIES` son PRODUCTO, no solo overhead de método. La migración a `docs/` debe preservar
que `knowledge/` (metodología-as-code sobre arneses) es el dominio del producto ArnesIA.

## Decisiones (firmas 🧑‍⚖️)

| # | Decisión | Estado |
|---|---|---|
| D1 | **Vehículo de las skills** (`/pm` + arneses) = **plugin marketplace único** (`harness@prenter-marketplace`). Los 3 repos lo consumen; cero fork; backflow. | 🟢 FIRMADA 2026-07-09 |
| D2 | **Formato capability-as-code** = **YAML por-cap + escenarios BDD** (modelo vitalia): dev_preview + scenarios Gherkin + change-log + deps; enforcement `cap_doctor` + `# cap:` headers → drift falla el commit. | 🟢 FIRMADA 2026-07-09 |
| D3 | **Los 3 ejes temporales caen DENTRO de `docs/projects/`**: pasado = stories/epics archivados + `LEDGER.md`; futuro = `BACKLOG.md` + epics abiertos; presente = `STATE.md` (puntero). `projects/` = pasado+presente+futuro del desarrollo. | 🟢 FIRMADA 2026-07-09 |
| D4 | **No construir de cero — extender el plugin existente** y trasladar lo nuestro a su formato (respuesta del operador). El kit ya trae ~70% (rules, process, templates, seam-loader). Se le añaden los huecos de método. | 🟢 FIRMADA 2026-07-09 (dirección) |

## Hallazgo — el kit `harness@prenter-marketplace` v0.5.3 ya existe

Repo fuente `alpacapurpura/prenter-marketplace`, plugin en `plugins/harness/0.5.3`, canal ESTABLE,
pinneado por harness-studio + cockpit + dev-studio. Es un **kit de método + seam-loader +
telemetría**, tech/sistema-agnóstico, lee `project.config.yaml`.

- **Ya trae (reusar):** 21 `rules/` (invariantes de proceso: tdd-mandatory, story-closure-gate,
  anti-duplication, pm-skill-chaining con contrato de auto-chain, paradigm-arquitectura…);
  `process/` (harness-lifecycle, **ticket-states 12**, spec-mapa-funcional, continuous-improvement,
  tech-debt, cockpit-permissions); **5 templates** numerados (`00-chris-input`, `00-research`,
  `00-story`, `01-spec` con Gherkin+matriz-cobertura, `story-ui.yaml` con **máquina 10 estados**
  idea→refining→refined→ready→developing→developed→reviewing→done|parked|dropped);
  `scripts/harness_config.py` (**seam-loader `--doctor`** = patrón para cap_doctor); telemetría;
  git-safety; skill `harness-bootstrap` (instalador).
- **NO trae, por diseño open-closed (capa "proyecto" en cada repo) — nuestros 3 huecos:**
  1. skills de rol **`/pm`** (+ po/architect/auditor/dev-team) — referidas en todos lados,
     nunca extraídas al kit;
  2. **template capability YAML + `cap_doctor.py`** (referido por la CIL L4 pero no shipped);
  3. **scaffolder del árbol `docs/`** (los templates asumen el árbol ya creado).

**Arquitectura de la homologación (2 capas):**

| Capa | Qué | Dónde |
|---|---|---|
| **MÉTODO** (compartido) | `/pm` + arneses · templates · schema capability · `cap_doctor` · scaffolder | **plugin** (nueva versión, backflow) |
| **CONTENIDO** (por-repo) | `docs/{product,architecture,projects}` real · `project.config.yaml` · CLAUDE.md router | **harness-studio** primero |

## Decisiones nuevas (del hallazgo)

| # | Decisión | Estado |
|---|---|---|
| D5 | **Vocabulario = el del kit tal cual**: `docs/product/{stories,capabilities,releases,modules}` + `docs/architecture/`. **Sin `projects/` ni `epics/`.** Máxima homologación con kit + vitalia (ya construido). | 🟢 FIRMADA 2026-07-09 |
| D6 | **Forjar el MÉTODO en harness-studio esta sesión** (`.claude/skills/` local + `docs/` + `cap_doctor` + scaffolder), probar E2E aquí, **upstream al plugin en la sesión de replicación**. Respeta "esta sesión = harness-studio; plugin y otros repos después". | 🟢 FIRMADA 2026-07-09 |

### Reconciliación D3 ↔ D5 (registrada, visible)

D5 elimina `projects/`, así que el **nombre de carpeta** de D3 queda superseded. Los 3 ejes
temporales caen ahora **dentro de `docs/product/`**:

- **presente** = `docs/product/checkpoint.md` (← `ESTADO.md`)
- **futuro** = `docs/product/BACKLOG.md` (← `BACKLOG.md`)
- **pasado** = `docs/product/LEDGER.md` + `ledger/HS-NN.md` + `archive/` (stories cerradas)

El **espíritu** de D3 (pasado≠presente≠futuro en una zona coherente) se mantiene: esa zona es
`docs/product/`. El operador puede vetar esto en el gate.

**Research (del ask original):** el kit/vitalia lo maneja per-story (`00-research-*.md`) + cross-cutting
`learnings/`. Para honrar el pedido original ("capabilities apuntan a su fuente"), se conserva un
cross-cutting `docs/product/research/` que caps y stories pueden citar. A confirmar en el gate.

## Layout objetivo homologado (harness-studio)

```
docs/
  product/            # qué existe + pipeline de trabajo (delta story → saldo capability)
    vision.md         # ← VISION.md (constitución) o puntero
    checkpoint.md     # PRESENTE  ← ESTADO.md
    BACKLOG.md        # FUTURO    ← BACKLOG.md
    LEDGER.md ledger/ # PASADO (decisiones) ← LEDGER.md + ledger/
    releases/         # hitos (vocab kit) — mapea fases HS
    stories/          # ← historias/*  (formato 00-research..07-merge)
    capabilities/     # ← CAPABILITIES.md → YAML por-cap + BDD (D2)
      _template.yaml  capability.schema.yaml  <module>/<slug>.yaml
    modules/  research/  archive/
  architecture/       # ← arch/ + STACK.md  (técnico, cómo está construido)
    INDEX.md  boundaries/ fitness/ conventions/ contracts/ model/  stack.md
  process/            # doctrina de proceso (del kit)
  knowledge/          # DOMINIO PRODUCTO ArnesIA (metodología-as-code) — ubicación diferida
.claude/
  skills/pm/SKILL.md  # /pm front-door (forja local, upstream luego)
  scripts/cap_doctor.py  scripts/scaffold_docs.py
project.config.yaml   # THE SEAM (kit lo consume)
CLAUDE.md             # router → docs/
```

## Rebanada de esta sesión (esqueleto + /pm + 1 cap + 1 story + gates) — NO rompe arch-tests

Regla honesta: **no** muevo `arch/` ni `CAPABILITIES.md` en masa (romperían `capability_trace_test.go`,
`go-arch-lint`, pre-commit). `docs/` nace **al lado**; migración masiva = paquete siguiente, gap VISIBLE.

1. `project.config.yaml` (seam) + `harness_config.py --doctor` verde.
2. `scaffold_docs.py` → árbol `docs/` + READMEs de propiedad + INDEX.
3. `capabilities/_template.yaml` + `capability.schema.yaml` (BDD, D2).
4. Migrar **1 capability real** (loader) → YAML + `# cap:` header en el `.go`.
5. `cap_doctor.py` (patrón `--doctor`): schema + `# cap:` ↔ YAML + puntero resuelve; verde sobre la 1 cap.
6. `/pm` skill: bootstrap (lee seam + checkpoint + BACKLOG + stories abiertas → dónde estás +
   próxima transición legal + auto-chain). Probar invocándola.
7. Migrar **esta historia** → `docs/product/stories/` (dogfood, primera story homologada) + `checkpoint.md`.
8. `checkpoint.md` (← ESTADO) + fila "docs/" en CLAUDE.md router (additivo).
9. `PARIDAD.md` + gate humano.

**Diferido VISIBLE (próximo paquete):** migrar 82 caps; mover `arch/`→`docs/architecture/` con fix de
paths de arch-tests; mover `historias/`→`stories/`; ubicar `knowledge/`; `releases/` real; **upstream del
método al plugin (nueva versión)**; replicar a cockpit + dev-studio.

## D7 — Adopción PLENA del plugin (directiva operador 2026-07-09)

> Cita cruda: *"adoptemos lo que dice el plugin, traslademos todo a como funciona el plugin"*.

Supersede el alcance "1 rebanada" de D6: **esta sesión = trasladar harness-studio ENTERO a la
estructura y método del plugin**, no solo un ejemplo. Regla de honestidad intacta: cada mover se
**verifica verde** (`go test ./...`, `go-arch-lint`, `lefthook`, `--doctor`, `cap_doctor`); lo que
no pueda quedar verde HOY se revierte y queda **VISIBLE** como diferido — jamás roto ni pintado.

**Acoplamientos a respetar al mover (recon previo obligatorio):** `arch/` y `knowledge/` son el
**ruleset que consume el motor** (CAP-26 `parser.go`) + `go-arch-lint.yml` + `arch_test.go` +
`capability_trace_test.go` + `lefthook`; `CAPABILITIES.md` lo parsea `capability_trace_test.go`
(R1/R2) y lo regenera `scripts/estado.sh`. Mover = editar esos consumidores en el mismo paso.

Orden de olas: (1) seam `project.config.yaml` + `/pm` + `docs/` + `cap_doctor` + templates →
(2) ejes temporales ESTADO/BACKLOG/LEDGER → `docs/product/` → (3) `CAPABILITIES.md` → YAML por-cap
+ reescribir `capability_trace_test.go` → (4) `historias/`→`stories/`, `arch/`→`docs/architecture/`,
`knowledge/` (toca motor). Verificar verde tras cada ola.

## Log de ejecución (2026-07-09)

- **T1 ✓** `project.config.yaml` (seam, 12 slots reales Go+React+Tauri) + `scripts/harness_config.py`
  (copia project-layer del kit, ADOPTING §1) → `--doctor` **exit 0**.
- **T2 ✓** `docs/{product,architecture,process}/` + READMEs de propiedad (R1/R2/R3) + templates del
  plugin verbatim (`00-story`, `00-research`, `01-spec`, `story.yaml`) + `capability.template.yaml`
  homologado (schema plugin + punteros anti-drift `file#Símbolo`).
- **T3 ✓** `scripts/cap_doctor.py` (doctor local: schema + enum + unicidad + archivo-existe; `--index`).
- **T4 ✓** `.claude/skills/pm/SKILL.md` — front-door seam-driven (bootstrap: closure-gate scan +
  lee checkpoint/BACKLOG/releases; 10 estados; auto-chain). Bootstrap verificado contra estructura nueva.
- **T5 ✓** `git mv` ESTADO→`docs/product/checkpoint.md`, BACKLOG, LEDGER, `ledger/`. Actualizado
  `scripts/estado.sh` + router `CLAUDE.md`. Cero refs de código rotas.
- **T6 ✓** `scripts/capabilities_to_yaml.py` (converter determinista) → **82 hojas
  `docs/product/capabilities/{module}/{slug}.yaml`** (15 módulos) + `_coverage.yaml`. Enforcer
  `arch/fitness/capability_trace_test.go` R1/R2 **reescrito para leer el árbol YAML** (dependency-free,
  block-aware sobre `pointers:`). `CAPABILITIES.md` retirado → `INDEX.md` (narrativa preservada +
  tabla generada por `cap_doctor --index`). Comentario lefthook actualizado.
  **Verificación: `go build ✓`, `go vet ✓`, `go test ./... 11 ok / 0 FAIL`, `cap_doctor 82 ✓`.**
- **T7 ✓** (operador eligió "relocación física completa ahora"): `git mv` `arch/`+`knowledge/`+`STACK.md`
  → `docs/architecture/`. Reescritos: `//go:embed` (root), `parser.go` dirs, 4 `filepath.Join` por-segmento,
  `iofs.Sub` ×2, `provisioner.copyTree`, `arch_test.resolveSchema`, `.go-arch-lint.yml` (exclude), `ci.yml`,
  `lefthook.yml`, seam, + ~115 refs en un pase determinista (guardas anti doble-reemplazo en `knowledge/`).
  **Verificado:** `go build`✓ · `go test ./...` 11 ok/0 FAIL · `go-arch-lint` OK · `conformance --todo` **247**
  · `--arnes` **21·20·1** (idénticos al baseline) · `estado.sh` pipeline E2E ✓.
- **T8 ✓** `historias/` → `docs/product/stories/` (10 paquetes, git mv; el activo untracked con `mv`) +
  records sueltos → `docs/product/research/` (10). Refs Go (conductor→research, dogfood→stories),
  `go-arch-lint` exclude, router `CLAUDE.md` actualizados. `checkpoint.md` (state=developed) + `PARIDAD.md`
  del paquete creados. `/pm` closure-gate scan ve el paquete (🔴 OPEN).
- **Estado:** homologación en harness-studio COMPLETA y verde. Falta gate humano + paquetes de
  replicación (upstream al kit · cockpit · dev-studio · arneses secundarios).

## D8 — VISION/METODOLOGIA/UX raíz → docs/ (firmada 2026-07-09)

**Decisión (placement por identidad, guiada por los READMEs de propiedad, NO adivinada):**

| Archivo raíz | Home nuevo | Por qué (owner-README que lo declara) |
|---|---|---|
| `VISION.md` | **`docs/product/vision.md`** | `docs/product/README.md:10` YA lo declara como eje-hoja de producto (owner Chris; constitución = norte del producto). `/pm` SKILL.md:94 YA apunta ahí. Cero cambio de README, cero cambio de `/pm`. |
| `UX.md` | **`docs/product/ux.md`** | UX firmada = superficie del producto (qué-hace/cómo-se-ve). Se agrega fila a `product/README` + a la lista R1 de hojas permitidas en raíz de `product/`. |
| `METODOLOGIA.md` | **`docs/process/metodologia.md`** | metodología · disciplina §10 · honestidad §4 · contrato-de-caja §3 = doctrina de *cómo construimos*. `process/README.md:26-29` reserva exactamente el slot «Local (capa proyecto)» para la doctrina específica de harness-studio que extiende la del kit. |

**Casing:** minúsculas (`vision.md`/`ux.md`/`metodologia.md`) homologando `vision.md` ya declarado.
El CONCEPTO en prosa (VISION, METODOLOGIA, UX como nombres propios de la doctrina) queda en
mayúsculas y **no se toca** — solo cambian los *targets* de links markdown + el 1 string con path
en `box.contract.schema.json`. Separación limpia concepto↔archivo (igual que ya hacía product/README).

**Alcance = MOVER + cablear, NO partir contenido.** El split de UX.md (15k) / METODOLOGIA.md (8.5k)
por «backlog+historia mezclados» (BACKLOG `[reorg-docs]`) sigue siendo paquete aparte — esta tarea
solo reubica + repara cableado. `research/*` line-citations históricas (`UX.md:534` en specs frozen
de `mapa-mvp`) quedan como snapshot; solo se recomputan links VIVOS.

**Cero acople de código:** los 3 son doc pura — grep confirma que ningún path los carga (solo
comentarios/`description`); NO están en `//go:embed` (embed = `docs/architecture/*`). Mover = doc-only.

- **T9 ✓** `git mv` los 3 a sus homes. Recompute determinista de los 30 targets de links markdown
  (relpath por profundidad; repara de paso los `../VISION.md` que T7 dejó rotos al mover `knowledge/`).
  Cableado: router `CLAUDE.md` (3 filas) · `README.md` raíz refresco completo (estaba stale pre-homolog:
  citaba CAPABILITIES/ESTADO/STACK inexistentes) · `product/README` (+fila ux +R1) · `process/README`
  (+metodologia en «Local») · `box.contract.schema.json` (1 path) · `project.config.yaml` (1 comment) ·
  `/pm` SKILL (metodologia cableada en disciplina/cierre; vision ya estaba) · `stack.md`. **Verificado:**
  `go build`✓ · `go test ./...` · `go-arch-lint` · `conformance --todo`/`--arnes` (idénticos baseline) ·
  `cap_doctor` · `--doctor` · 0 links rotos a los 3.

