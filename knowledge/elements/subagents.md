---
elemento: subagent
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/sub-agents
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://claude.com/blog/subagents-in-claude-code
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://www.anthropic.com/engineering/multi-agent-research-system
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://stevekinney.com/courses/ai-development/subagent-anti-patterns
    autoridad: experto
    revisado: 2026-07-04
  - url: https://www.pubnub.com/blog/best-practices-for-claude-code-sub-agents/
    autoridad: experto
    revisado: 2026-07-04
---

# Subagent (.claude/agents/*.md)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Archivo Markdown + frontmatter YAML que define un asistente especializado y
acotado a una tarea, con **ventana de contexto propia**, su system prompt, sus tools y su
modelo. Solo su **mensaje final + metadata** vuelve al padre → es el mecanismo primario de
**aislamiento de contexto** (mantener exploración/logs/investigaciones fuera del hilo
principal). *(oficial: code.claude.com/docs/en/sub-agents)*

**Ubicación y frontmatter (L1.2).** `.claude/agents/*.md` (proyecto, sube al root) y
`~/.claude/agents/`. Requeridos solo `name` (lowercase-hyphen, único por scope) y `description`
(dispara la auto-delegación). Opcionales: `tools` (allowlist; omitir = hereda TODO incl. MCP),
`disallowedTools`, `model` (`sonnet|opus|haiku|fable|inherit`), `permissionMode`, `maxTurns`,
`skills` (precarga contenido completo), `mcpServers`, `hooks`, `memory`, `background`, `effort`,
`isolation: worktree`, `color`. Cuerpo = system prompt. Recibe solo ese prompt + entorno + el
mensaje de tarea; **nunca** la historia del padre (salvo fork). *(oficial)*

**Prácticas (L1.3):** responsabilidad única («cada subagente excele en UNA tarea») ·
description con **condición de disparo concreta** («Reviews code for security issues before
commits» rutea mejor que «security expert») · **tools mínimas** (least privilege; omitir =
escalada de privilegio implícita) · modelo por costo/complejidad (Haiku para volumen barato,
Opus para análisis) · **contrato de retorno estructurado** (objetivo + formato de salida +
límites) — el paper multi-agente: subagentes guardan trabajo en sistemas externos y pasan
**referencias ligeras**, no resúmenes con pérdida. Versionar en git los de proyecto.
*(oficial + Kinney + PubNub)*

**Anti-telephone (L1.4, clave):** el resumen de un subagente describe **intención, no efecto**
→ «verifica el diff, no el resumen». Sin contrato de retorno («investiga esto») la salida es
inusable. Fix estructural de Anthropic: artefacto-por-referencia, no re-resumen. *(oficial +
Kinney)*

**Cuándo subagente vs skill vs inline (L1.5):** subagente cuando el sub-trabajo ensuciaría el
hilo con resultados intermedios que no volverás a mirar (explorar 10+ archivos, 3+ piezas
independientes); skill cuando quieres ver/dirigir cada paso en el hilo principal. El paralelismo
multi-agente ganó 90,2% vs single-agent Opus **a ~15× el costo en tokens**. *(oficial: anchor blog + multi-agent post)*

**Novedades (L1.6):** `memory` (conocimiento persistente cross-sesión) · subagentes anidados
hasta profundidad 5 (v2.1.172, requiere `Agent` en tools) · fork mode (hereda TODO el contexto
del padre) · `isolation: worktree` · background-by-default (v2.1.198) · Explore ahora hereda el
modelo del padre (v2.1.198) · `Task`→`Agent` (v2.1.63) · `Agent(name1,name2)` allowlist de
spawn · **Agent Teams** para workers que se comunican (los subagentes NO se comunican entre sí
— coordinarlos ad-hoc es anti-patrón oficial). *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Un subagente en un arnés nuestro es **maquinaria dedicada de una caja** (⇐ VISION A1: la caja
= skill-frente que orquesta agentes/sub-skills). No es una caja: no posee transición de estado
ni gate propio; sirve a la caja que lo invoca.

1. **Contrato de retorno anti-telephone obligatorio:** todo agente nuestro declara su salida en
   clave `<veredicto> → <path>` (o formato estructurado equivalente), nunca «resume». ⇐ L1.4.
   *(ya es doctrina en METODOLOGIA §2; aquí queda su raíz en L1)*
2. **Tools least-privilege explícitas:** `tools` como lista, nunca omitido (omitir = acceso
   total incl. MCP). ⇐ L1.3. Un agente que escribe pero se describe «solo-lectura» = hallazgo.
3. **Modelo deliberado por rol:** builders/reviewers eligen modelo por costo/complejidad, no
   `inherit` silencioso cuando la tarea difiere del default. ⇐ L1.3 + principio 11.
4. **Responsabilidad única, sin solape:** N agentes con descripciones solapadas rompen la
   auto-delegación → se detecta y se consolida. ⇐ L1.1/L1.3. Conecta con el hallazgo real de
   luana (`auditor-agentic` 0 lanzamientos = maquinaria muerta de una caja).
5. **Aislamiento para escritura paralela:** cajas paralelas (A5, por marca/tipo) que escriben
   → `isolation: worktree`. ⇐ L1.6 + A5.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| agent-required-fields | frontmatter con `name` y `description` | error | badge «no delegable — falta campo» | L1.2 |
| agent-name-unique | `name` lowercase-hyphen, único en su scope | error | badge «nombre duplicado» | L1.2 |
| agent-desc-trigger | description da condición de disparo concreta, no rol vago | warn | badge «descripción vaga — auto-delegación falla» | L1.3 |
| agent-tools-scoped | `tools` es allowlist explícita, no omitida | warn | badge «agente con acceso total a tools» | L1.3 · L2.2 |
| agent-return-contract | cuerpo especifica forma de salida (`<veredicto>→<path>`), no «resume/investiga» | warn | badge «sin contrato de retorno — riesgo telephone» | L1.4 · L2.1 |
| agent-write-vs-role | si tiene Write/Edit/Bash, el cuerpo justifica modificar (vs rol solo-lectura) | warn | badge «escribe pero se describe solo-lectura» | L2.2 |
| agent-bypass-flagged | `permissionMode: bypassPermissions` presente | error | banda Guardia «modo sin fricción — revisar» | L1.2 |
| agent-single-responsibility | no enumera múltiples tipos de tarea sin relación | warn | badge «agente multipropósito — dividir» | L1.1 · L2.4 |
| agent-model-deliberate | `model` fijado cuando la tarea difiere del default | info | badge «modelo hereda por defecto» | L1.3 · L2.3 |
| agent-overlap | descripciones solapadas entre agentes del árbol | warn | «N agentes solapados — auto-delegación no confiable» | L1.3 · L2.4 |
| agent-cold-30d | maquinaria de una caja con 0 lanzamientos en 30d pese a caja activa | warn | punteado + hallazgo «maquinaria muerta» | L2.4 |
| agent-vcs | agente de proyecto versionado en git | info | «agente de proyecto no versionado» | L1.3 |
| agent-return-schema-strict | la salida cumple un schema JSON estricto + summary ≤~200 tok (no prosa libre) | warn | badge «retorno sin schema — el orquestador re-parsea» | doctrina v1 · [[harness-profile]] L1.3 |

> **Los 6 patrones de orquestación** (Delegated Data Access · Temp File Assembly · Shared-File ·
> Hierarchical Lead-Worker · Persona-Driven Parallel · Evolutionary) y la disciplina «el padre no lee
> lo que delega» viven en el nodo [`harness-profile`](./harness-profile.md) (doctrina v1 §8.2), que
> gobierna el *cómo ejecuta*; este nodo gobierna *qué es* un subagente (maquinaria de caja).

## Changelog

- 2026-07-05 · v1.1 · **Doctrina v1 (HS-07):** +1 check `agent-return-schema-strict` (schema JSON +
  summary ≤200 tok, extiende el anti-telephone de retorno). Los 6 patrones de orquestación + «padre no
  lee lo delegado» se registran en el nodo nuevo [[harness-profile]]. 12→13 checks. · barrido externo.
- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales + blog subagentes (2026-04-07) +
  multi-agent research post + Kinney/PubNub. L2 amarra subagente = maquinaria dedicada de una
  caja; anti-telephone `<veredicto>→<path>` anclado a L1.4. 12 checks. Novedades: `memory`,
  anidados prof. 5, fork mode, worktree isolation, background-default, Agent Teams. · disparado
  por pasada fundacional del árbol.
