# Spike de medición — ¿el contrato fusionado gasta contexto? (debate #6)

> 2026-07-07 · ordenado por el operador («corre el spike de medición») · principio 11:
> economía de contexto MEDIBLE — se decide con dato, no con intuición.

## Método

Dos sesiones REALES de Claude Code (`claude -p … --plugin-dir dogfood/dev-full-cycle
--max-turns 1 --output-format json`, cwd limpio en scratchpad, mismo entorno del usuario):

- **A (baseline):** «Responde únicamente: OK» — plugin cargado, skill NO invocada.
  Sesión `c96b4929`.
- **B (activación):** `/spec-writer idea: …` — la skill del dogfood con el contrato
  fusionado MÁS pesado (frontmatter ≈1.5KB, contrato ≈1.3KB ≈ ~330 tok).
  Sesión `bf5cc6fc`.

Fuente de verdad: el JSONL de `~/.claude/projects/...` de cada sesión (enumerar, jamás
parsear a mano el vivo) + `usage` del resultado.

## Resultados

| Métrica | A (baseline) | B (skill activada) |
|---|---|---|
| input_tokens | 12 821 | 12 804 |
| cache_read | 15 190 | 18 332 |
| cache_creation | 3 352 | 629 |
| contexto total aprox. | ~31.4k | ~31.8k |
| costo | $0.21 | $0.18 |

**Lo que entró al contexto al activar la skill (JSONL de B):**

1. Turno user de expansión del comando: 243 chars ≈ **60 tok**
   (`<command-name>/dev-full-cycle:spec-writer</command-name>` + args).
2. Inyección de la skill: 749 chars ≈ **187 tok** — «Base directory for this skill: …» +
   **SOLO EL CUERPO** markdown del SKILL.md (desde `# spec-writer — la caja…`) + ARGUMENTS.

**Lo que NO entró:** el frontmatter COMPLETO — `perfil_harness` y `escritor_unico`
aparecen **0 veces** en todo el JSONL; el único «arquetipo» es prosa del cuerpo.
El contrato fusionado costó **0 tokens** en la activación.

## Veredicto (debate #6)

1. **El yml hermano es INNECESARIO para skills.** Claude Code ya se comporta como el
   operador quería: al activar una skill inyecta solo el cuerpo; el frontmatter es
   metadata que CC lee para descubrimiento (name/description) y el resto lo ignora sin
   costo. El **contrato fusionado firmado (METODOLOGIA §3) queda VALIDADO gratis**:
   as-code para la fábrica (loader/conformance/drawer), 0 tokens para el runtime.
2. **Costo real de una activación** de la caja más pesada del dogfood: ~247 tok
   (expansión + cuerpo). El presupuesto de diseño va al CUERPO de la skill, no al
   contrato.
3. **La única superficie sensible siguen siendo las rules** (CLAUDE.md siempre en
   contexto) — allí ya rige la convención de identidad mínima (loader HS-11); la meta
   de reglas jamás va al CLAUDE.md.
4. **Pendientes honestos:** (a) el listado name+description de skills SÍ vive en el
   system prompt siempre (no aislable por JSONL; estimado ~40-60 tok por skill — es
   carga ÚTIL: es lo que permite a CC invocarlas); (b) subagents/commands comparten
   formato frontmatter — mismo mecanismo presumido, NO medido; (c) esto describe el
   CC instalado hoy — si su comportamiento cambiara, un check de conformance puede
   vigilar la invariante «frontmatter no entra al contexto».

## Consecuencia para el paquete inspector-drawer

El drawer puede mostrar TODO el contrato (Gherkin, evidencia, ejes) sin culpa de
contexto: la información de entendimiento vive as-code en el componente y no cuesta
tokens en el uso diario. No hay cambio de nomenclatura ni enmienda doctrinal que hacer.
