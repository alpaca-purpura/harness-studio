---
elemento: skill
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/skills
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://agentskills.io/specification
    autoridad: estándar-abierto
    revisado: 2026-07-04
  - url: https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://arxiv.org/html/2601.10338v1
    autoridad: experto
    revisado: 2026-07-04
---

# Skill (SKILL.md · agent skill)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Capacidad file-based invocada por el modelo: carpeta con instrucciones +
scripts/referencias opcionales que Claude descubre por un trigger ligero (name+description) y
carga bajo demanda. Es **conocimiento procedimental** («cómo hacer una tarea recurrente
multi-paso»), distinto de CLAUDE.md (hechos always-on) y de hooks (automatización
determinista no invocada por el modelo). El cuerpo cuesta ~0 tokens hasta que se invoca → es
la palanca principal de dirección **económica en contexto**. *(oficial: overview + steering
blog, 2026-07-04)*

**Ubicación y formato (L1.2).** `<dir>/SKILL.md`, una skill = un directorio. Personal
`~/.claude/skills/<n>/`, proyecto `.claude/skills/<n>/`, plugin `<plugin>/skills/<n>/`
(namespaced `plugin:skill`). YAML frontmatter + cuerpo Markdown. *(oficial: code.claude.com/docs/en/skills)*

**Frontmatter — estándar abierto (L1.3, agentskills.io):** requeridos `name` (≤64,
lowercase alfanumérico+guiones, = nombre del dir, sin «anthropic»/«claude» ni tags XML) y
`description` (1–1024 chars, dice **qué + cuándo**). Opcionales `license`, `compatibility`,
`metadata`, `allowed-tools` (experimental). *(estándar-abierto, publicado 2025-12-18)*

**Frontmatter — extensiones Claude Code (L1.4):** `when_to_use`, `argument-hint`,
`arguments`, `disable-model-invocation`, `user-invocable`, `disallowed-tools`, `model`,
`effort`, `context: fork`, `agent`, `hooks`, `paths` (auto-activación por glob), `shell`.
El listado de skills se trunca a **1.536 chars** (`description`+`when_to_use`).
*(oficial: code.claude.com/docs/en/skills)*

**Progressive disclosure — 3 niveles (L1.5):** (1) metadata name+description **siempre**
cargada, ~100 tok/skill · (2) cuerpo SKILL.md solo al dispararse, **<5.000 tokens
recomendado** · (3) `scripts/` (se ejecutan, no se leen) + `references/` (bajo demanda) +
`assets/` = «prácticamente ilimitado». **SKILL.md <500 líneas**; detalle a archivos aparte,
**referenciados un nivel de profundidad** (sin cadenas anidadas). *(oficial + estándar-abierto)*

**Autoría dirigida por evals (L1.6):** construir evals ANTES de documentar — baseline sin la
skill → ≥3 escenarios → instrucciones mínimas que cierran la brecha. Automatizable con el
plugin `skill-creator` (genera prompts should/shouldn't-trigger y afina la description por
hit-rate). *(oficial: best-practices; skill-creator plugin)*

**Prácticas de autoría (L1.7):** description en **tercera persona**, qué+cuándo, caso clave al
frente (trunca a 1.536) · nombres en **gerundio** (`processing-pdfs`), nunca `helper`/`utils` ·
casar «grados de libertad» con la fragilidad (prosa para juicio, script exacto para operaciones
frágiles) · retar cada párrafo con «¿Claude ya sabe esto?» · scripts deterministas > código
generado para lo frágil · `context: fork` + `agent:` cuando la skill es tarea autocontenida
(no forkear contenido de solo-referencia). *(oficial: best-practices)*

**Seguridad (L1.8):** auditar scripts empaquetados de skills no propias (llamadas de red,
instrucciones que no casan con el propósito). Estudio a escala (arxiv 2601.10338, 2026-01):
**26,1%** de 31k skills con ≥1 vulnerabilidad; skills con script **2,12× más** probables de
tenerla. Incidente real: skill «GIF Creator» descargó ransomware (Cato CTRL, 2025-12). Las
skills no están cubiertas por Zero Data Retention. *(experto + oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Una skill en un arnés nuestro es **una de dos cosas**, y esto lo enforcamos (⇐ VISION A1 +
METODOLOGIA §2):

- **Skill-CAJA** = el frente de una fase del proceso. Posee UNA transición del estado del
  trabajo y declara su **contrato** (§3 METODOLOGIA). ⇐ L1.1 (procedimiento multi-paso) +
  L1.6 (eval-gate = el `contract.gate`).
- **Skill de APOYO** = librería-experta / utilidad / tercero / meta-harness. `contract.caja:
  false` + `rol:`. No posee transición ni gate. ⇐ L1.1 (distinción procedimiento vs referencia).

**Reglas L2 (cada una deriva de L1):**

1. **Frontmatter obligatorio nuestro:** además de `name`+`description` (⇐ L1.3), toda
   skill-caja lleva `version`, `model`, `clase` (contrato L0 `meta.clase`, I-75) y el bloque
   `contract:`. ⇐ L1.4 (Claude Code ya admite `model`; nosotros lo hacemos obligatorio) +
   A2/A4 (el contrato y su gate).
2. **Description = qué + cuándo, tercera persona, ≤1.536 chars, caso al frente.** Adoptado tal
   cual. ⇐ L1.7. *(sin divergencia)*
3. **Prosa mínima de una caja:** `## Cuándo` (precondición = su input del contrato), `## Pasos`,
   `## Guardarraíles`. ⇐ L1.5 (cuerpo <500 líneas) + L1.7 (grados de libertad: `## Pasos`
   modula prosa vs script).
4. **El eval-gate de la skill-caja vive en `contract.gate`** y es **honesto**: `tipo: none`
   cuando no hay eval real (nunca fabricado). ⇐ L1.6 (evals antes de doc) + A4 + METODOLOGIA §3.
   *(esto es MÁS estricto que L1: L1 recomienda evals; nosotros los volvemos un campo del
   contrato que el mapa mide. No es divergencia — es aterrizaje del principio 10.)*
5. **Scripts empaquetados se auditan al importar** (conformación §6). ⇐ L1.8. Un arnés nuestro
   no adopta script sin revisar.
6. **Curaduría agresiva:** skill propia sin activación en 30d = candidata a poda / lazy-load
   (hallazgo de Diagnóstico, no borrado automático). ⇐ L1.7 (economía de contexto) + insight
   de escala de la it.8 (mar de nodos fríos).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| skill-name-format | `name` ≤64, lowercase+guiones, = dir, sin reservados | error | badge «nombre inválido — no carga como skill» | L1.3 |
| skill-desc-missing | `description` presente | error | badge «sin trigger — el modelo nunca la invoca sola» | L1.3 |
| skill-desc-what-when | description trae capacidad **y** cláusula «usar cuando…» | warn | badge «descripción débil (falta el cuándo)» | L1.7 · L2.2 |
| skill-desc-length | description+when_to_use <1.536 chars, caso clave al frente | warn | badge «se trunca en el listado — reordena» | L1.5 |
| skill-desc-pov | tercera persona (sin «I can help»/«you can») | info | badge «voz inconsistente en el matching» | L1.7 |
| skill-body-linecount | cuerpo SKILL.md <500 líneas | warn | badge «cuerpo >500 líneas — mover a references/» | L1.5 |
| skill-ref-depth | referencias a un nivel, sin cadena SKILL→A→B | warn | badge «referencia anidada — lectura parcial» | L1.5 |
| skill-caja-contract | skill-caja declara bloque `contract:` completo | error | badge «caja sin contrato — sin gate ni hand-off» | L2.1 · A2 |
| skill-gate-honest | `contract.gate.tipo` presente; `none` si no hay eval real | warn | badge «gate sin declarar» / chip «SIN GATE» honesto | L2.4 · A4 |
| skill-invocation-guard | skill con efecto irreversible → `disable-model-invocation: true` | error | badge «acción irreversible auto-invocable — sin guardia» | L1.4 · L1.8 |
| skill-script-audit | `scripts/` auditados (sin red no declarada / deps sin pin / ofuscación) | error | badge «script sin auditar — riesgo exfiltración» | L1.8 · L2.5 |
| skill-time-sensitive | sin lenguaje con fecha/deadline fuera de sección «patrones antiguos» | info | badge «contenido caduca» | L1.7 |
| skill-cold-30d | skill propia con 0 activaciones en 30d | info | punteado frío + hallazgo «candidata a poda» | L2.6 |
| skill-has-evals | skill-caja con evals asociados (≥3 casos, delta con/sin) | warn | badge «sin evals — efectividad no medida» | L1.6 · L2.4 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales + agentskills.io (estándar abierto
  publicado 2025-12-18) + estudio de seguridad arxiv 2601.10338. L2 amarra skill-caja/apoyo al
  modelo de fábrica de cajas (contrato + gate honesto). 14 checks. Novedades registradas:
  merge commands→skills, `context: fork`, `paths:` glob-scoped, `shell:` injection, plugin
  skill-creator, features v2.1.145/196/199. · disparado por pasada fundacional del árbol.
