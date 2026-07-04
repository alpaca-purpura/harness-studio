# Investigación UX iteración 3 — salud de producción · trazas · edición de contenido

> 3 frentes web + verificación local de JSONL (2026-07-04, fase UX HS-03). Condensado:
> lo decisivo para la propuesta. Detalle completo en la transcripción de sesión.

## Frente 1 — Cómo visualizan salud los productos de LLM observability

Cubiertos: Langfuse, LangSmith, Braintrust, W&B Weave, Datadog LLM Obs, Arize Phoenix,
AgentOps, Helicone. Patrones que repiten TODOS (el estándar de facto):

1. **Fila de KPIs + serie de tendencia**: volumen · error rate · costo · latencia · tokens.
2. **Latencia en cuantiles** (p50/p95/p99 + TTFT), jamás promedios.
3. **Descomposición de costo/tokens**: por tipo de token Y Pareto top-N (modelos, features).
4. **Error por paso/tool**, no solo global (per-tool error rate: LangSmith Tools, Phoenix).
5. **Calidad como serie de primera clase**: eval scores online + feedback junto a métricas
   ops — «nada en el dashboard sin eval» ya es norma de industria (valida principio 10).
6. **Todo punto del dashboard clickea hasta su traza** — dashboards = puertas a trazas,
   nunca destinos finales.
7. **Segmentación slice-by-anything con top-N** (metadata como group-by, cap top-5).
8. Emergentes 2025-26: clustering de tráfico por tópico (Datadog Patterns), NL→chart
   (Braintrust Loop), session replay + grafo (AgentOps), comparación por versión de prompt.

## Frente 2 — Patrones SRE/fleet health aplicables al portafolio

- **RED** (Rate/Errors/Duration) = vista de servicio; **USE** = recursos; Golden Signals ≈
  unión. Clave operativa (Wilkie): misma tarjeta para TODOS los servicios → cualquiera
  puede leer cualquier arnés.
- Mapeo a arneses: Rate=corridas/día · Errors=corridas fallidas+fallos de tool ·
  Duration=latencia corrida · Saturación≈% ventana de contexto + presupuesto tokens ·
  **5ª señal AI: eval/quality score** (ningún método clásico cubre corrección).
- **SLO/error budget**: % restante (stat con umbral) + burndown vs ideal; burn rate
  multiventana (14.4x/1h paging estándar Google). Para ArnesIA: el contrato de mejora
  continua con el cliente ES un SLO — candidato fase 3+.
- **Scorecards** (Backstage Soundcheck/Cortex/OpsLevel): nota compuesta (nivel/medalla) +
  sub-scores por categoría + qué-arreglar-siguiente; portafolio grande = heatmap
  entidad×dimensión (Cortex birds-eye), no un dashboard por servicio.
- **Release health** (Argo Rollouts/LaunchDarkly): rollout = barra de pasos discretos con
  veredicto por gate inspeccionable; delta nuevo-vs-baseline con banda de confianza;
  auto-rollback por regresión. Estados de máquina: Progressing/Paused/Degraded/Healthy.
- **Diseño de estado**: semáforo SOLO con umbrales documentados y cuantitativos; rojo/ámbar/
  verde reservados para estado (nunca decorativos); color + icono/forma (8% daltónicos);
  **gris = sin datos/inactivo ≠ verde**; sparkline junto a cada KPI (tendencia + punto);
  pirámide invertida: overview → drill-down → dato crudo.

## Frente 3 — Trazas de una corrida + contenido versionado

**Idiomas visuales de traza (repetidos en LangSmith/Langfuse/OTel GenAI):**
1. **Árbol colapsable + panel de detalle** = base universal (click paso → inputs/outputs/
   tokens/latencia).
2. **Waterfall/timeline como segunda vista** del mismo run (latencia/paralelismo) — nunca
   única vista.
3. **Render conversación como tercera vista** (LangSmith Messages tab, Langfuse session
   replay) — «leerla como chat» vs «debuggearla como traza».
4. Grafo de topología emergente (Langfuse agent graph beta).
5. Tokens+latencia por paso, costo agregado en raíz (costo = usage × tabla de precios).
6. Colapso: tool calls consecutivos agrupados; subagentes = subárboles expandibles.
7. **Attribution: click en paso → definición del componente que lo produjo.**
8. Correlación por un id por turno de usuario (promptId/trace id).

**OTel GenAI semconv** (en Development): `invoke_agent` → `chat {model}` / `execute_tool
{tool}`; attrs `gen_ai.usage.*`, `gen_ai.tool.name`; Claude Code exporta OTel (metrics +
events + spans beta con jerarquía interaction→llm_request/hook/tool, subagentes anidados).

**JSONL de ~/.claude (verificado local, fuente de verdad del sensor):** `parentUuid` (árbol/
DAG) · `promptId`/`requestId` (correlación) · `message.model` exacto + `usage` completo
(input/output/cache_creation/cache_read + tier + speed) por turno · `tool_use`/`tool_result`
enlazados (`sourceToolUseID`) · subagentes en archivos hermanos (`subagents/agent-*.jsonl`,
`isSidechain`) · hooks (`stop_hook_summary` con comando+duración) · skills
(`attributionSkill`, attachment `skill_listing`) · comandos slash (`<command-name>`) ·
compaction. **Todo lo que la UX de corridas necesita ya existe en el JSONL.**

**Contenido versionado (Langfuse Prompt Mgmt/Braintrust/PromptLayer) — idioma universal:**
- **Versiones inmutables + labels movibles** (deploy/rollback = mover label, jamás editar
  historia) → mapea 1:1 a nuestro git+semver+canales beta/estable (KIT-06). No inventar.
- Diff entre versiones en todas; diff-before-save (PromptLayer); notas estilo commit.
- Playground junto al historial: editar y evaluar nunca separados (valida dock+tren).
- Versión ↔ trazas que la usaron (linkage estándar).
- **Editor embebido: CodeMirror 6, no Monaco** (300KB vs 5-10MB; consenso: Monaco solo si
  el producto ES un IDE; Replit y Sourcegraph migraron a CM6).

## Fuentes principales

Frente 1: langfuse.com/docs/metrics · docs.langchain.com/langsmith/dashboards ·
braintrust.dev/docs/observe · weave-docs.wandb.ai (monitors) · docs.datadoghq.com/llm_observability ·
arize.com/docs/phoenix · docs.agentops.ai · docs.helicone.ai.
Frente 2: grafana.com/blog (RED) · sre.google/workbook/alerting-on-slos ·
backstage.spotify.com (Soundcheck) · docs.cortex.io (birds-eye) · docs.opslevel.com ·
argo-rollouts.readthedocs.io · launchdarkly.com/docs (guarded rollouts) ·
grafana.com/docs (best practices).
Frente 3: opentelemetry.io/docs/specs/semconv/gen-ai · code.claude.com/docs/en/monitoring-usage ·
langfuse.com (timeline/agent-graphs/sessions/prompt-version-control) ·
docs.langchain.com/langsmith/view-traces · braintrust.dev/docs/evaluate/playgrounds ·
docs.promptlayer.com (registry) · sourcegraph.com/blog/migrating-monaco-codemirror ·
JSONL local `~/.claude/projects/*/`.
