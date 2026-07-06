# knowledge/ — árbol de conocimiento vivo de ArnesIA

> **La metodología as code.** Cómo se debe estructurar cada elemento de un arnés (skill,
> hook, rule, subagent, command, MCP, plugin, settings, output-style, statusline, headless,
> **perfil de harness**) según las mejores prácticas actuales de Anthropic y los expertos — y
> cómo lo adaptamos a nuestra forma de trabajo. **No es un doc que se lee una vez: es un árbol
> que crece cada semana** (mecanismo en [`CADENCE.md`](./CADENCE.md)).
>
> **Linaje (doctrina v1, 2026-07-05):** este árbol es la realización as-code de que
> **operacionalizamos Agentic BPM** — un arnés da *framed autonomy* a Claude Code por rol×proceso.
> Cada L1 se ancla a primitivas CC nativas (firewall §8.6 de METODOLOGIA); las fuentes académicas/
> industria (manifiesto Agentic BPM, Sierra ADLC, Salesforce) son insumos, no el padre. Detalle:
> [`../research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`](../research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md).

## Cómo se relaciona con el resto del repo

- [`../VISION.md`](../VISION.md) — constitución (11 principios + anatomía A1–A7). El **norte**.
- [`../METODOLOGIA.md`](../METODOLOGIA.md) — reglas de negocio / doctrina operativa. **Deriva**
  de este árbol: «qué debe tener cada componente» (§2–3) se apoya en los nodos de aquí.
- [`../UX.md`](../UX.md) — la UX que consume los checks (los puntos de mejora que el mapa pinta
  salen de los `Checklist evaluable` de cada nodo).
- [`../research/2026-07-04-salud-trazas-edicion.md`](../research/2026-07-04-salud-trazas-edicion.md)
  — investigación one-shot de salud/trazas/edición. Este árbol es la **evolución viva** de esa
  idea de «investigar antes de inventar», aplicada a la anatomía de los componentes.

## Las dos capas (en cada nodo)

1. **L1 · Estándar** — qué recomiendan Anthropic (oficial) y los mejores de internet. Fechado,
   con fuente. Evidencia, no opinión.
2. **L2 · Nuestra adaptación** — nuestra forma de trabajo, **obligada a derivar de L1**.
   Divergencia = marcada y justificada, nunca silenciosa.

Y cada nodo emite un **`Checklist evaluable`**: la rúbrica as code (checks con severidad +
señal en el mapa) que vuelve el estándar algo que ArnesIA puede **medir** por componente.

## El árbol — nodos

| Nodo | Elemento | Estado | Versión | Checks |
|------|----------|--------|---------|--------|
| [`elements/skills.md`](./elements/skills.md) | Skills (SKILL.md, agent skills) | 🌱 vivo | 1.1 | 17 |
| [`elements/hooks.md`](./elements/hooks.md) | Hooks (settings.json events) | 🌱 vivo | 1.0 | 12 |
| [`elements/rules.md`](./elements/rules.md) | Rules / memoria (CLAUDE.md, AGENTS.md, imports) | 🌱 vivo | 1.1 | 13 |
| [`elements/subagents.md`](./elements/subagents.md) | Subagents (.claude/agents) | 🌱 vivo | 1.1 | 13 |
| [`elements/commands.md`](./elements/commands.md) | Slash commands (custom + built-in) | 🌱 vivo | 1.0 | 11 |
| [`elements/mcp.md`](./elements/mcp.md) | MCP servers | 🌱 vivo | 1.0 | 11 |
| [`elements/plugins.md`](./elements/plugins.md) | Plugins & marketplaces | 🌱 vivo | 1.0 | 12 |
| [`elements/settings-permissions.md`](./elements/settings-permissions.md) | Settings & permisos | 🌱 vivo | 1.0 | 11 |
| [`elements/output-styles.md`](./elements/output-styles.md) | Output styles | 🌱 vivo | 1.0 | 8 |
| [`elements/statusline.md`](./elements/statusline.md) | Status line | 🌱 vivo | 1.0 | 8 |
| [`elements/headless-sdk.md`](./elements/headless-sdk.md) | Headless / Agent SDK | 🌱 vivo | 1.1 | 11 |
| [`elements/harness-profile.md`](./elements/harness-profile.md) | Perfil de harness (loop · subagentes · routing) | 🌱 vivo | 1.0 | 11 |

Leyenda de estado: ⏳ en forja · 🌱 vivo (nace, se actualiza) · 🌳 estable · 🔍 en-revisión.
**Total: 12 nodos · 138 checks evaluables · pasada fundacional 2026-07-04; headless→v1.1 en HS-04;
**doctrina v1 (HS-07, 2026-07-05):** nace el nodo 12 `harness-profile` (11 checks) + skills/rules/
subagents bumpeados (+5 checks del firewall CC-native y perfil de harness).**

## Cómo crece

Cada semana: barrido de novedades → triage → append a L1 → revisión de L2 → evolución de
checks → bump + changelog → propagar a METODOLOGIA/UX. Detalle y reglas del árbol en
[`CADENCE.md`](./CADENCE.md).

## Índice de checks (agregado)

La unión de los `Checklist evaluable` de los 12 nodos = **138 checks** = el ruleset que el
linter de conformidad de ArnesIA correrá sobre un arnés (METODOLOGIA §6). Reparto por elemento
en la tabla de arriba. Severidades: `error` rompe el estándar · `warn` huele mal · `info` mejora
posible. Cada check declara su **señal en el mapa** (columna 4 de cada nodo) — ése es el puente a
la UX: qué badge/estado pinta ArnesIA en Diagnóstico / capa Desempeño / bandas.

**Transversales que cruzan nodos** (cross-checks — un check en un nodo referencia a otro):
- **regla ↔ hook ↔ permiso**: una «nunca X» destructiva ([[rules]]) exige un hook que la enforce
  ([[hooks]]) o un deny ([[settings-permissions]]); si es solo prosa = advisory, hallazgo.
- **secreto en config**: mismo check en [[settings-permissions]], [[mcp]], [[plugins]], [[statusline]].
- **costo de contexto always-on**: [[rules]] + [[mcp]] + [[plugins]] suman contra el presupuesto
  (principio 11); atribución solapada declarada (METODOLOGIA §4).
- **frío / sin uso 30d**: [[skills]], [[subagents]] (maquinaria muerta), [[mcp]], [[plugins]] —
  superficie fría = hallazgo de poda.
- **anti-telephone**: contrato de retorno `<veredicto>→<path>` en [[subagents]], que el conductor
  headless ([[headless-sdk]]) consume estructurado.

> **Estado (HS-08):** el motor **`arnesia conformance` ya está construido** (fase 4, dogfood-first) —
> estos 138 checks (+ los de `arch/`) se parsean a datos y se corren; la UX (it.11+) los pinta sobre el mapa.
> **Ejecución unificada knowledge + `arch/` por un solo runner (`arnesia conformance`) — construida en
> HS-08** (hexagonal: `RulesetPort` parsea AMBOS árboles = 235 checks a datos; `ConformancePort` + adapters
> por mecanismo arch-test/schema-validation/go-arch-lint/static-scan/nl-judge). Muchos checks siguen siendo
> telemetría/juicio-NL — los cubre el adapter `nl-judge`; el contrato de check común (mecanismo por check) ya existe.
