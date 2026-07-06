---
regla: orquestacion-determinista-entre-cajas
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-07
sources:
  - url: https://github.com/humanlayer/12-factor-agents
    autoridad: experto
    nota: "Own your control flow — la secuencia la dueña el código, no el modelo."
    revisado: 2026-07-05
  - url: https://arxiv.org/abs/2603.18916
    autoridad: académica
    nota: "Agentic BPM: A2 — orquestación determinista, agencia acotada (framed autonomy)."
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestConductorOwnsBoxRouting
severity: high
---

# La secuencia ENTRE cajas es código; la agencia del modelo vive DENTRO de una caja

## L1 · Principio (estándar de industria)

**Orquestación determinista, agencia acotada (A2 de Agentic BPM + 12-Factor «own your control
flow»).** En un sistema de proceso, **la secuencia entre unidades de trabajo es determinista** — la
mueve código, no la decisión del LLM. La autonomía del modelo (*framed autonomy*) vive **dentro** de
una unidad, acotada por su frame. El fin de una unidad se lee de una señal **legible por máquina**
(evento `result` + `status` del artefacto), **jamás del texto del chat**. *(experto:
12-factor-agents; académica: arXiv 2603.18916, distinción autonomía≠automatización)*

## L2 · Realización (este árbol Go+React)

El **conductor Go** dueña el control-flow; el `contract:` de caja (METODOLOGIA §3) declara el ruteo.

- **El conductor ejecuta `ruta`/`si`:** el hand-off caja→caja lo decide código Go leyendo el
  `contract.ruta` (condición `si`), **no** el conductor conversacional ni el LLM. ⇐ L1 (secuencia
  determinista). *(Detalle abierto HS-07: runner determinista vs conductor — ver research doctrina §10.6.)*
- **El loop es del conductor, no del skill:** ningún skill «llama» a un loop; el T3 corre como
  for-loop en Go (`claude -p --max-turns`) que lee `result.subtype` + `status`. ⇐ boundary
  `adaptadores-de-agente-intercambiables` + `conductor-no-parsea-jsonl`.
- **No inferir del texto:** continue/stop/block sale del evento `result` de stream-json + el `status`
  del artefacto (document-as-cache), nunca de scrapear el chat. ⇐ L1 + `conductor-no-parsea-jsonl` L2.
- **La agencia vive en la caja:** dentro de una caja de arquetipo `abierto`/`excepcion` el modelo
  elige cómo actuar dentro del frame; eso NO se determiniza — se acota con la Guardia. ⇐ framed autonomy.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| conductor-dueña-el-loop | el loop de una caja T3 lo corre el conductor Go, no un skill que «llama» loop | error | capa Proceso «loop en el skill — control-flow invertido» | arch_test.go:TestConductorOwnsBoxRouting |
| ruta-ejecutada-por-codigo | el hand-off caja→caja lo decide código leyendo `contract.ruta/si`, no el LLM | warn | capa Proceso «hand-off no determinista» | arch_test.go |
| no-infiere-del-texto | continue/stop/block se decide de `result`+`status`, no del texto del chat | error | «orquestador scrapea texto — frágil» | arch_test.go |
| agencia-dentro-del-frame | la autonomía del modelo está acotada por la Guardia/permisos de la caja | warn | banda Guardia «agencia sin frame» | arch_test.go |

## Changelog

- 2026-07-05 · v1.0 · Nodo draft (HS-07, doctrina v1). L1 = A2 de Agentic BPM + 12-Factor. L2: el
  conductor Go dueña el loop y ejecuta `contract.ruta`; la agencia vive dentro de la caja (framed
  autonomy). 4 checks · `status: proposed` (se enforça cuando el conductor aterrice el ruteo). ·
  disparado por la doctrina v1 (VISION §Linaje).
