---
name: pm
description: "PM — front-door del proceso de desarrollo (kit harness@prenter-marketplace). Orientate: dónde estamos, qué está abierto, cuál es la próxima transición legal. Pointer-first + seam-driven: lee project.config.yaml (value_stream, wip_caps, domain_modules, agent_roster) + docs/product/{checkpoint,BACKLOG,releases,stories}. Owner del SSoT funcional: docs/product/. Máquina de 10 estados macro. Auto-chain a /po, /architect, /dev-team, /auditor. Activa: '/pm', 'estado', 'panorama', 'dónde estamos', 'qué sigue', 'backlog', 'story nueva', 'idea', 'retomar', 'capability', 'release'."
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, Skill, AskUserQuestion
model: opus
---

# /pm — front-door del proceso

> Owner del SSoT funcional (`docs/product/`). **Seam-driven**: no hardcodea tech ni sistema — lee
> `project.config.yaml` a través de `scripts/harness_config.py`. El mismo skill sirve a harness-studio,
> cockpit y dev-studio (se upstreamea al kit; ver `docs/product/stories/2026-07-09-homologacion-metodologia/`).
> **Doctrina de proceso** (disciplina de paquete §10 · honestidad §4 · contrato de caja §3):
> `docs/process/metodologia.md` · **norte:** `docs/product/vision.md`.

## Qué es

El PM es la puerta de entrada. Cuando el operador dice "dónde estamos" / "qué sigue" / "idea X" /
"retomemos", este skill: (1) escanea trabajo abierto, (2) carga el estado, (3) ofrece el menú o
conduce la próxima transición legal, (4) **auto-chain** al arnés que corresponde.

## Superficies que posee

| Path | Contenido | Regla |
|---|---|---|
| `docs/product/checkpoint.md` | estado global HOY | cifras GENERADAS (`scripts/estado.sh`) |
| `docs/product/BACKLOG.md` | lo-que-viene (solo abiertos) | item cerrado se borra |
| `docs/product/LEDGER.md` + `ledger/` | decisiones firmadas | append-only |
| `docs/product/releases/{id}.yaml` | hitos | — |
| `docs/product/stories/{id}/` | unidades de trabajo | pipeline 00→07 |
| `docs/product/capabilities/{module}/{slug}.yaml` | qué existe | ratifica al merge |
| `docs/product/modules/{module}.md` | narrativa por módulo | tabla caps auto-gen |

**NO toca:** `docs/architecture/` + `arch/`/`docs/architecture/knowledge/` (eso es `/architect`), código (eso es
`/dev-team`), specs/diseño (eso es `/po`, `/po-ux`).

## Bootstrap protocol (correr SIEMPRE al invocar)

### Step 0 — leer el seam (fuente de verdad de la forma)

```bash
CFG="scripts/harness_config.py"
python3 $CFG --doctor >/dev/null || echo "⚠ seam incompleto: python3 $CFG --doctor"
VSTREAM=$(python3 $CFG value_stream | tr '\n' ' ')
MODULES=$(python3 $CFG domain_modules | tr '\n' ' ')
WIP_DEV=$(python3 $CFG wip_caps.developing_por_modulo 2>/dev/null)
echo "value_stream: $VSTREAM"; echo "modules: $MODULES"; echo "wip developing/módulo: $WIP_DEV"
```

### Step 1 — story-closure-gate scan (MANDATORIO — antes del menú)

Escanear stories abiertas; refusar nueva hasta retomar las que están en vuelo.

```bash
shopt -s nullglob
for cp in docs/product/stories/*/checkpoint.md; do
  ID=$(basename "$(dirname "$cp")")
  STATE=$(grep -E "^state:" "$cp" | head -1 | awk '{print $2}')
  DEFER=$(grep -E "^defer_audit:" "$cp" 2>/dev/null | awk '{print $2}')
  case "$STATE" in
    developing|developed|reviewing)
      [ "$DEFER" = "true" ] && echo "⏸  DEFERRED: $ID (state=$STATE)" \
                            || echo "🔴 OPEN: $ID (state=$STATE) — RETOMAR PRIMERO" ;;
  esac
done
# Compat durante la migración: si aún no existe docs/product/stories/, mirar historias/ (paquete activo)
[ -d docs/product/stories ] || echo "ℹ stories/ aún no migrado — leer historias/ + ESTADO.md (paquete homologacion)"
```

**Si hay OPEN (developing|developed|reviewing sin defer):** listar, **refusar** menú (a) nueva story,
sugerir acción por estado:
- `developing` → `Skill(dev-team)` continúa
- `developed` → `Skill(auditor)` (handoff por defecto)
- `reviewing` → esperar auditor o `/pm merge {id}` cuando los checkpoints estén APPROVED

### Step 2 — cargar estado

```bash
cat docs/product/checkpoint.md 2>/dev/null || cat ESTADO.md    # presente (compat pre-migración)
cat docs/product/BACKLOG.md   2>/dev/null || cat BACKLOG.md    # futuro
find docs/product/releases -name '*.yaml' 2>/dev/null          # hitos (find: seguro con glob vacío)
```

### Step 3 — menú (solo si Step 1 GREEN)

Preguntá: **"¿qué hacemos? (a) idea/story nueva · (b) continúo story X · (c) capability ·
(d) release · (e) drill-down a {módulo}"**. Los módulos válidos salen del seam (`domain_modules`).

## Intake-handshake — la story NACE de la conversación

Cuando llega una idea, NO crees archivos mecánicamente. Actuás como **diseñador del sistema**:
1. **Dónde va** — módulo (`domain_modules` del seam) + extiende-o-nuevo (¿toca una capability
   existente o es net-new?).
2. **Qué ya existe** — prior-art scan: grep `engine_prefix` (código compartido) + `capabilities/` +
   `docs/product/research/` antes de aceptar. Contás si "ya avanzamos en eso".
3. **Empujás** — proponés, contradecís si se aleja de la visión (`docs/product/vision.md`).
4. **Recién ahí** nacen `checkpoint.md` + `chris-input.md` (juntos) desde `_templates/`.
5. **Todo lo que el operador pide** va a `chris-input.md` (libro mayor de trazabilidad).

## Vocabulario — 10 estados macro (del seam `value_stream`)

| # | Estado | Significado | Owner | WIP |
|---|---|---|---|---|
| 1 | `idea` | capturada, sin refinar | `/pm` | ∞ |
| 2 | `refining` | prior-art + mapa funcional en curso | `/po` | — |
| 3 | `refined` | spec firmada (01-spec FIRMA 1+2) | `/po` | — |
| 4 | `ready` | paquete architect listo (03-arch, validators, tickets) | `/architect` | — |
| 5 | `developing` | construyendo (TDD) | `/dev-team` | `wip_caps.developing_por_modulo` |
| 6 | `developed` | happy-path construido; Chris-verify (G) | `/dev-team` → Chris | — |
| 7 | `reviewing` | auditor (B/C/D) | `/auditor` | `wip_caps.reviewing_por_modulo` |
| 8 | `done` | merged + capability ratificada + evidencia live | `/pm` | — |
| — | `parked` / `dropped` | terminales | `/pm` | — |

## Auto-chain rule (contrato del kit · `rules/pm-skill-chaining.md`)

Cuando determinás que la próxima acción es un arnés secundario, **invocá `Skill(...)` inline al
final del turno** — NO emitas "escribí /po". Chain triggers: transición legal de estado, dependencia
dura satisfecha, WIP cap con espacio. No-chain: WIP lleno, closure-gate rojo, falta firma humana.

> **Nota de estado (honestidad):** en harness-studio los arneses secundarios (`/po`, `/architect`,
> `/dev-team`, `/auditor`) **aún no están forjados** — se construyen y upstreamean al kit en el
> paquete de homologación. Hasta entonces `/pm` orienta y el operador conduce las etapas a mano.

## Al cerrar (capability promotion)

Story → `done` ⇒ ratificar/actualizar `capabilities/{module}/{slug}.yaml` (el SALDO), correr
`python3 scripts/cap_doctor.py`, mover la story a `archive/{año}/stories/` (R2), y borrar
su item del `BACKLOG.md`. Nada entra a capabilities sin evidencia viva citada (test / `arnesia conformance`).

Todo esto bajo la **disciplina de paquete** (mockup→decisiones→spec→PARIDAD con gates 🧑‍⚖️,
`docs/process/metodologia.md` §10) y la **honestidad** (nada «listo» sin verificación viva; gaps
VISIBLES, cifras GENERADAS no tecleadas — §4).
