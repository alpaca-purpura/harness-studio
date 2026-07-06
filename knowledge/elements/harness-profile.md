---
elemento: harness-profile
version: 1.0
updated: 2026-07-05
status: vivo
fuentes:
  - url: https://arxiv.org/abs/2603.18916
    autoridad: académica
    nota: "Agentic BPM: A Research Manifesto (Information Systems 2026, DOI 10.1016/j.is.2026.102738). Framed autonomy · autonomy≠automation · 4+1 capacidades."
    revisado: 2026-07-05
  - url: https://arxiv.org/abs/2504.03693
    autoridad: académica
    nota: "Agentic BPM: Practitioner Perspectives on Agent Governance. Autonomía-por-riesgo · métricas de gobernanza. (n=22, direccional)"
    revisado: 2026-07-05
  - url: https://sierra.ai/blog/agent-development-life-cycle
    autoridad: experto
    nota: "ADLC — ciclo de vida de agente en producción; immutable snapshot; bounded-error-rate; anotación humana→regresión."
    revisado: 2026-07-05
  - url: https://sierra.ai/blog/enterprise-grade-agents
    autoridad: experto
    revisado: 2026-07-05
  - url: https://architect.salesforce.com/docs/architect/fundamentals/guide/agentic-patterns.html
    autoridad: experto
    nota: "Patrones agénticos; guardrails enforced-at-reasoning-layer; permisos task-based que expiran; handoff explícito."
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/headless
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/sub-agents
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://github.com/humanlayer/12-factor-agents
    autoridad: experto
    revisado: 2026-07-05
---

# Perfil de harness (cómo una caja ejecuta el loop, subagentes y routing)

> **El nodo 12 — el segundo eje del contrato de caja.** Los otros 11 nodos describen *qué es* cada
> elemento (skill, hook, rule…). Éste describe *cómo ejecuta*: tipo de loop, orquestación de
> subagentes, routing de modelo, punto de escalada. Nace de la **doctrina v1** (2026-07-05):
> operacionalizamos **Agentic BPM** — un arnés da *framed autonomy* a Claude Code por rol×proceso.
> `clase` (I-75) ⊥ `perfil_harness`: una skill-caja (clase=skill) puede ser T1 o T3.

## L1 · Estándar (académico + industria + primitivas CC)

**Framed autonomy — autonomía ≠ automatización (L1.1, APM).** El manifiesto *Agentic BPM* distingue:
la **automatización** ejecuta tareas predefinidas exactamente como se especifican; la **autonomía**
deja que el agente perciba, razone y **elija cómo actuar DENTRO de un frame de proceso**. El frame se
compone de una capa **normativa** (deóntica: obligaciones / prohibiciones / permisos) y una
**operacional** (procedimientos). Un agente APM debe proveer **4 capacidades**: *framed autonomy ·
explainability · conversational actionability · self-modification* (esta última con dos niveles:
**adaptation** = instancia, efímera; **evolution** = modelo, persistente). *(académica: arXiv 2603.18916)*

**El orquestador dueña el loop (L1.2, 12-Factor + Agent SDK + CC).** El control-flow es del código
que llama, no del modelo (12-Factor «own your control flow»). El fin del loop se lee de una señal
**legible por máquina** — el `ResultMessage`/evento `result` (subtype éxito/límite, costo, tokens) +
el `status` del artefacto — **nunca del texto del chat**. Tope duro `max_turns` + un **cap de
iteraciones explícito** para loops de reparación (doble red); estado terminal `blocked` → handoff.
En Go **no hay Agent SDK oficial** → el orquestador es el **conductor subprocess + stream-json** (los
SDK TS/Py spawnean el mismo binario). *(oficial: headless, cli-reference; experto: 12-factor-agents)*

**Subagentes = cajas negras sin estado; el filesystem es la verdad (L1.3, CC + patrones).** El
contexto del padre queda diminuto (punteros + plan); cada subagente recibe instrucciones y devuelve
**solo su resultado final** (aislamiento de contexto). Seis patrones de orquestación: *Delegated Data
Access · Temp File Assembly · Shared-File · Hierarchical Lead-Worker · Persona-Driven Parallel ·
Evolutionary*. **Error #1: el padre lee lo que delega** → los tokens ya se gastaron; fix = lenguaje
defensivo («tu rol es ORQUESTACIÓN, NO leas los archivos objetivo, lístalos por nombre»). Salida con
**schema estricto** + summary ≤~200 tok. *(oficial: sub-agents + multi-agent research post)*

**Routing por carga cognitiva (L1.4).** Modelo por clase de carga: alta (planificar/auditar) →
Opus/Fable · media (construir) → Sonnet · baja (recorrer/extraer) → Haiku + `effort: low`. Cada
elemento con LLM declara su clase; el multi-agente se justifica contra su costo (~15× un agente
único), no se asume. *(oficial + APM §economía)*

**Guardrails fuera del razonamiento (L1.5, Salesforce + CC).** Los límites duros se imponen en una
capa que **el agente no puede razonar para esquivar** — hooks deterministas (exit 2), no instrucción
de prompt. Los permisos pueden ser **task-based y expirar** (mínimo privilegio temporal), no acceso
perpetuo. *(experto: Salesforce Agentforce; oficial: hooks)*

**Ciclo de vida en producción (L1.6, Sierra ADLC).** Un agente serio no para en «published»: recorre
Development → **Release (immutable snapshot** = bundle atómico código+prompts+modelo+knowledge+
permisos, A/B-testeable) → **QA** (SME anota una traza real; la anotación se **captura como eval de
regresión**) → **Testing** (simulación conversacional vs mock antes de tocar producción) → mejora.
Filosofía: de «perfect adherence» a **bounded error rates** («systems problem, not whack-a-mole»).
*(experto: Sierra ADLC + enterprise-grade agents)*

**Autonomía acotada por riesgo (L1.7, práctica).** Cuánta autonomía dar depende del *blast-radius*
del efecto: rutinario → autónomo; cambios en sistemas fuente / decisiones de alto valor → humano.
Métricas de gobernanza: override-frequency · time-to-escalation · exception-count · audit-completeness.
*(académica: arXiv 2504.03693 — direccional, n=22 sin experiencia operativa)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

El perfil de harness es el **segundo eje del contrato de caja** (METODOLOGIA §3, §8.2). Un arnés
nuestro ES un *framing mechanism* instanciado: banda Guardia + permisos + rules = **frame normativo +
operacional**; la caja da **framed autonomy** a Claude Code para una etapa del proceso.

1. **`perfil_harness` obligatorio por caja (T1–T3):** T1 pase único · T2 workflow stateful
   (document-as-cache obligatorio) · T3 worker autónomo. **T4 = shell/sesión, NO una caja** (el
   agente-de-rol persistente vive en la sesión, HS-06). ⇐ L1.1/L1.2.
2. **El conductor Go dueña el loop, no el Agent SDK.** T3 = for-loop en Go que spawnea
   `claude -p --max-turns N`, lee `result` (stream-json) + `status` del artefacto, aplica **cap de
   reparación explícito**, termina en `blocked`→handoff. La FSM interna (adaptation) anida dentro de
   la caja; el spine `idea→done` (evolution) es cross-caja. ⇐ L1.2 + boundary
   `adaptadores-de-agente-intercambiables`.
3. **El padre no lee lo que delega + schema estricto de retorno.** Extiende el anti-telephone de
   [[subagents]] (contrato de RETORNO) con la disciplina de ENTRADA (el orquestador no toca el
   material delegado). ⇐ L1.3.
4. **Explainability = 5ª faceta evaluable del arnés.** Además de honestidad-de-dato (§4) y atribución
   traza→componente (el Mapa), el arnés emite el **rationale accionable** de por qué decidió X (que
   indique la corrección sin escalar). Cae en nuestro moat de observabilidad. ⇐ L1.1 (capacidad APM).
5. **Autonomía acotada por riesgo, impuesta en la Guardia.** El nivel de autonomía se declara y el
   efecto peligroso se bloquea en hooks (exit 2), nunca por prompt. Frontera P6/Guardia (§8.5). ⇐
   L1.5/L1.7.
6. **Bounded-error-rate + immutable snapshot + anotación→regresión** aterrizan en observar/mejorar:
   `arnés@version` como bundle atómico (refuerza reuso por referencia A12); la anotación humana de una
   traza se captura como eval de regresión (superficie en la lente observar del Mapa); heat=percentil
   (§4) ya apunta al bounded-error. ⇐ L1.6. *(candidatos de producto, no solo checks)*
7. **Degradación elegante:** una caja multi-agente cae a ejecución secuencial si la maquinaria no está.
   ⇐ L1.3.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| perfil-tipo-declarado | la caja declara `perfil_harness` (T1/T2/T3) | error | badge «caja sin perfil de harness — el conductor no sabe si loopear» | L1.1 · L2.1 |
| arquetipo-declarado | la caja declara `arquetipo` (pipeline/excepcion/abierto/no-arnesar) | error | badge «caja sin arquetipo — forma del trabajo indefinida» | L1.1 · METODOLOGIA §8.1 |
| loop-iteration-cap | caja T3 con cap de reparación explícito **además** de `--max-turns` + estado terminal `blocked` | error | banda Guardia «loop sin cap — riesgo de bucle/quema» | L1.2 · L2.2 |
| orchestrator-reads-status | continue/stop/block se decide leyendo `result`+`status`, no el texto del chat | warn | «orquestador scrapea texto — frágil» | L1.2 · L2.2 |
| orchestrator-no-read-delegated | el padre/conductor NO lee el material que delega (lenguaje defensivo presente) | warn | capa Tokens «padre lee lo delegado — patrón derrotado» | L1.3 · L2.3 |
| cognitive-load-declared | cada elemento con LLM declara clase de carga → modelo; multi-agente justificado vs costo | info | «modelo no deliberado / fan-out sin justificar (~15×)» | L1.4 · L2 |
| explainability-rationale | la caja emite rationale accionable de sus decisiones (no solo output) | warn | lente Desempeño «sin explicabilidad — decisión opaca» | L1.1 · L2.4 |
| autonomia-por-riesgo | el efecto de alto blast-radius se bloquea en hook (exit 2), no por prompt | error | banda Guardia «autonomía no acotada en efecto peligroso» | L1.5 · L1.7 · L2.5 |
| tolerancia-declarada | `capability`/`constraint` de bajo nivel de tolerancia respaldada por hook, no solo guía | warn | «regla dura solo advisory (ver par regla↔hook)» | L1.5 · METODOLOGIA §8.5 |
| handoff-trigger-declarado | caja T2/T3 de alto riesgo declara `contract.handoff.cuando` | warn | capa Proceso «sin punto de escalada declarado» | L1.7 · METODOLOGIA §3 |
| graceful-degradation | caja multi-agente degrada a secuencial si la maquinaria falta | info | «sin fallback secuencial» | L1.3 · L2.7 |

## Changelog

- 2026-07-05 · v1.0 · **Nodo fundacional (doctrina v1, ficha HS-07).** Nace el 12º elemento: el
  perfil de harness = 2º eje del contrato de caja (⊥ `meta.clase`). L1 de fuentes académicas
  (manifiesto Agentic BPM arXiv 2603.18916 + gobernanza arXiv 2504.03693) + industria (Sierra ADLC,
  Salesforce Agentforce) + primitivas CC (headless, sub-agents, hooks). L2 amarra: perfil T1–T3 por
  caja (T4=shell), conductor Go dueña el loop (no Agent SDK), padre-no-lee-lo-delegado, explainability
  como 5ª faceta, autonomía-por-riesgo en la Guardia, bounded-error/snapshot/anotación→regresión en
  observar/mejorar. 11 checks. **Reencuadre doctrinal: operacionalizamos Agentic BPM, no clonamos un
  framework de agentes.** · disparado por el barrido externo de 7 fuentes (VISION §Linaje).
