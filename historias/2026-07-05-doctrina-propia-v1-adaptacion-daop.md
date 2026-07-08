# Doctrina propia v1 — adaptación de DAOP a la visión ArnesIA

> **Estado:** **RATIFICADA por el operador y BAJADA A AS-CODE (ficha HS-07, 2026-07-05).** Este doc es
> el «documento resumen de referencia a profundizar»; la doctrina vive as-code en: VISION §Linaje ·
> METODOLOGIA §3 (contrato fusionado) + §8 (doctrina de proceso) · `knowledge/elements/harness-profile.md`
> (nodo 12) + checks del firewall en skills/rules/subagents · `arch/boundaries/{orquestacion-determinista-
> entre-cajas,permisos-derivan-del-rol}.md`. Las 3 decisiones (§10: producto-puro+META · dogfood-first ·
> P6/HITL por scope) están firmadas. **Reencuadre: operacionalizamos Agentic BPM — disciplina de proceso,
> NO clon de BMAD (§11).**
> Producto de un estudio 5-frentes (subagentes por capa) que cruzó la investigación externa
> **DAOP v0.2** (`~/Descargas/doctrina-kits-v0.2.md`, derivada de BMAD v6 + Agent SDK) contra
> nuestras capas: [`VISION.md`](../VISION.md) · [`METODOLOGIA.md`](../METODOLOGIA.md) ·
> [`knowledge/`](../knowledge/INDEX.md) · [`arch/`](../arch/INDEX.md) ·
> [`historias/…inyeccion-knowhow.md`](./2026-07-05-arquitectura-inyeccion-knowhow.md).
> **Norte:** adaptar lo bueno de DAOP a NUESTRA visión sin importar los mecanismos BMAD que
> Claude Code no soporta. VISION firmada = intacta.

## 0. Tesis central (lo que el estudio concluyó)

**DAOP completa nuestra capa de CONTENIDO del arnés; nosotros conservamos la capa de SISTEMA.**
No es una doctrina rival — está *escrita para nosotros* (habla en segunda persona: «tu término»,
«tú ya usas worktrees»). Es fuerte donde somos delgados (anatomía del arnés, contrato de intención,
perfil de harness, front-door organizacional MOF→librería→manifiesto→override→permisos) y ciega
donde somos fuertes (modelo visual, event-sourcing confiable, seguridad del daemon, inyección
`--plugin-dir`, mejora continua). **La fusión es complementaria.**

Tres movimientos gobiernan toda la adaptación:
1. **Traducir DAOP hacia abajo** (regla de Rosetta, §1) — resuelve la colisión de vocabulario.
2. **Rechazar los BMAD-ismos** que CC ignora o que rompen decisiones firmadas (§3, el firewall).
3. **Exponer el seam de META** (rol·proceso·reporta-a) para que un sistema L1 EXTERNO (futuro)
   enganche y componga arneses — ArnesIA = **producto puro**, no consultoría de modelado (§6).

## 1. La regla de traducción (keystone — sin esto todo lo demás es trampa)

DAOP re-escala «Arnés» hacia abajo, al nivel de UNA parte-de-proceso, y para el agregado inventa
«manifiesto de rol». Colisiona con nuestro vocabulario firmado. **Mapeo obligatorio al importar
cualquier idea de DAOP:**

| DAOP | Nuestro (firmado) | Evidencia |
|---|---|---|
| **Arnés** (Kit⊕perfil, `arnes.fact.emitir-factura`) | **CAJA** (nodo `clase:skill` + `contract:`) | graph.l0 schema · DAOP §6/§19 |
| Kit = ⟨Contrato,Arquetipo,Elementos,Conocimiento,Permisos⟩ | contrato de caja + maquinaria + Base + permisos | METODOLOGIA §2–3 |
| Perfil-de-harness = ⟨Tipo,Loop,Subagentes,Hooks,Routing,Trigger⟩ | **(hueco — nace nodo 12)** | DAOP §6 ↔ ✗ |
| **Manifiesto de rol** (`rol.<area>.<cargo>`) | **ARNÉS** (objeto `arnes` + su grafo puesto×proceso) | graph.l0 `arnes` · VISION |
| Proceso (5ª capa aparte) | nuestros `edges` / `ruta` (ya fusionado en el arnés) | graph.l0 `edges` |

**Por qué NO renombramos:** nuestra fusión rol×proceso es SUPERIOR para el producto — es la unidad
publicable (marketplace por arnés), la del Organigrama (`reporta_a`), la de la sesión-shell
(frente N:1 con arnés, HS-06). DAOP tiene dos conceptos flacos (manifiesto-plano + proceso-DAG)
donde nosotros tenemos uno coherente. Romperlo cascadearía a marketplace, Organigrama, shell y
modelo de negocio. **VISION intacta; DAOP se absorbe a nivel de caja (loss-less).**

Lo que la colisión SÍ revela: **nuestra caja está sub-especificada** frente a lo que DAOP mete en
«Arnés». Eso se arregla enriqueciendo el contrato de caja (§2), no reescribiendo VISION.

## 2. El contrato de caja enriquecido (fusión de los dos contratos)

Nuestro contrato (METODOLOGIA §3) es de **cableado/topología**; el SPEC kernel de DAOP (§8) es de
**intención/aceptación**. Ejes perpendiculares del mismo objeto → un solo `contract:` fusionado.
**Supersede el §3 actual** (aditivo, no rompe lo escrito):

```yaml
contract:
  # ── INTENCIÓN (SPEC kernel DAOP §8) — NUEVO ──
  why: "<propósito inmutable de esta parte-de-proceso>"      # el "goal" que hoy no existe
  capabilities:
    - id: CAP-01                                             # IDs estables entre versiones
      what: "<qué logra — WHAT, no HOW>"
      success: "<señal concreta y verificable>"
  constraints: ["<solo los que doblan decisiones>"]          # NUEVO
  non_goals:  ["<lo que explícitamente NO hace>"]            # NUEVO

  # ── CLASIFICACIÓN — DOS EJES ORTOGONALES NUEVOS ──
  clase: <skill|agente|hook|knowledge|mcp|regla|command>     # meta.clase L0 (I-75) — qué ELEMENTO es
  arquetipo: <pipeline|excepcion|abierto|no-arnesar>         # §7 DAOP — FORMA del trabajo (NUEVO)
  perfil_harness: <T1|T2|T3>                                 # §5 DAOP — cómo EJECUTA (NUEVO; T4=shell, no caja)

  # ── CABLEADO (nuestro §3, intacto) ──
  caja: true
  fase: <id-de-fase>
  estado: "<estado_entra> -> <estado_sale>"                  # transición del spine que posee
  necesita: [{art, de: "usuario|base:<id>|caja:<skill>|libreria:<skill>|maquinaria:<agente>", requerido}]
  entrega:  [{art, escritor_unico: true}]                    # mutation contract §8.4 → check nuevo
  ruta:     [{a, si}]                                        # condicional; omitir en happy path

  # ── ACEPTACIÓN (gate honesto nuestro + Gherkin ejecutable DAOP) ──
  gate:
    tipo: "auto | manual | parcial | none"
    detalle: "<cómo; si none, POR QUÉ falta>"                # regla de honestidad §3 INTACTA
    aceptacion: [{given, when, then}]                        # Gherkin = gate.detalle hecho ejecutable
    evidencia: "<registro de auditoría emitido (A11 / procedencia §4)>"
```

**Reglas de la fusión (cementan tensiones):**
- **Gherkin = `gate.detalle` ejecutable.** Cuando `tipo:auto`, el Gherkin ES el eval. El slot ya
  existía; DAOP le da forma.
- **Honestidad preservada:** DAOP exige 100% Gherkin *para promover*; nosotros toleramos
  `tipo:none` *para existir* (badge de mejora, principio 10). No se contradicen si se separan:
  **existencia** tolera `none`; **promoción** exige verde. Defender la nuestra en existencia.
- **`escritor_unico` (mutation contract):** un solo artefacto = un solo escritor autorizado →
  check de conformidad nuevo (hoy dos cajas podrían escribir el mismo `art` sin detección). El
  *mecanismo* BMAD (skill `bmad-spec` headless) se rechaza; el *principio* se adopta.

## 3. El firewall CC-native (cruce con nuestro knowledge — lo más importante)

DAOP arrastra mecanismos de BMAD (framework TS/Python con runner propio) que **Claude Code ignora
en silencio** (claves de frontmatter fantasma → hacen CERO) o que **rompen decisiones firmadas**.
Dictamen por mecanismo (fuente: knowledge/ verificado contra docs CC v2.1.x):

| mecanismo DAOP | ¿CC-native? | primitiva CC equivalente | verdict |
|---|---|---|---|
| `persistent_facts` (globs→contexto, §11.1) | **NO** (CC lo ignora) | `.claude/rules/ paths:` + CLAUDE.md + `@import` | **TRADUCIR** |
| `activation_steps_prepend` (pre-fetch, §11.1) | **NO** | `SessionStart` hook + `additionalContext` | **TRADUCIR** |
| `customize.toml` merge 3-capas (§15) | **NO** (motor de merge es del runner BMAD) | precedencia `settings.json` + merge propio en `KitProvisioner` (JSON, no TOML) | **TOOLING-PROPIO** |
| sanctum PERSONA/CREED/BOND (§11.3) | **NO** (archivos muertos, no auto-cargan) | auto-memory `~/.claude/.../MEMORY.md` (≤200 líneas nativo) + CLAUDE.md + `@import` | **TRADUCIR + tirar teatro** |
| step-files capa 4 (§10) | **NO** | conductor headless prompt-a-prompt (`--max-turns 1`) | **RECHAZAR como default** (mata adaptabilidad) |
| filesystem blackboard (§12) | **SÍ (patrón sancionado)** | multi-agent post + `isolation:worktree` + `<veredicto>→<path>` + `SubagentStop` | **ADOPTAR** |
| `AgentDefinition` + `effort` (§6/§12) | **SÍ** | `.claude/agents/*.md` frontmatter (`tools/model/effort/maxTurns`) | **ADOPTAR** (como archivo, no objeto SDK) |
| orquestador Agent SDK (§6/§13) | **NO para Go** (no hay Go SDK) | **conductor Go subprocess + stream-json** (HS-04) | **RECHAZAR mecanismo / adoptar rol** |
| «usa API key / crédito dedicado» (§14) | **N/A** | suscripción-propia SIN `--bare` (historias/…inyeccion-knowhow.md §3) | **RECHAZAR default** (lane opcional para T3 pesado) |

**Check nuevo mission-crítico → `no-phantom-frontmatter`:** ninguna skill/subagente lleva claves
que CC no reconoce (`persistent_facts`, `activation_steps_prepend`, `customize`). Es la defensa
directa contra copiar BMAD ciegamente. **`--bare` sigue prohibido** en la ruta suscripción.

**Riesgo #1 concreto:** escribir `persistent_facts=[...]` en un frontmatter → CC lo ignora → las
reglas de negocio NUNCA cargan → el arnés parece configurado y no lo está. El firewall lo caza.

## 4. Estado, memoria y loop autónomo (todo CC-native, sin SDK)

- **Document-as-cache (A7) — ADOPTAR.** El estado del TRABAJO vive en el artefacto durable
  (frontmatter YAML: inputs·`status`·timestamps + secciones-borrador), no en la conversación.
  Sobrevive compactación (la etapa siguiente relee el doc), habilita pausa/resume/reset limpio.
  Hoy tenemos estado-de-sesión (HS-06) pero NO estado-de-trabajo-en-artefacto. CC no «resume del
  doc solo» → es convención con disciplina: el skill/conductor DEBE releerlo.
- **Memoria dos niveles — YA es CC-native.** session-logs → MEMORY.md curada. Mapea EXACTO a la
  auto-memory de CC (`~/.claude/projects/<p>/memory/MEMORY.md`, ≤200 líneas / 25KB nativo,
  Claude-authored). DAOP lo reinventa como mecanismo BMAD; CC ya lo trae con el mismo tope.
- **Distinción persona-state vs kit-state — ADOPTAR (sin el teatro).** Estado de persona (identidad,
  preferencias, memoria curada) persiste entre resets; estado de Kit (efímero) muere. Válida y
  conecta con la sesión-persistente HS-06. **Pero:** la persona-de-rol es concern del
  **runtime-trabajador (DevHub), no de la fábrica.** El sanctum-teatro (PERSONA/CREED/BOND, «sueño
  no muerte») se RECHAZA — un arnés es framework, no personaje. Además: una «orden permanente» de
  persona que sea regla dura DEBE respaldarse con hook (output-styles son probabilísticos, el
  entrenamiento base las pisa — issue #6450) → check `persona-hard-constraint-hook`.
- **Loop autónomo T3 — ADOPTAR vía conductor Go (NO Agent SDK).** 100% implementable sin SDK:
  - El orquestador = **for-loop en Go** que spawnea `claude -p --max-turns N` por iteración.
  - Lee el evento `result` de stream-json (`subtype` éxito/límite + costo + tokens) + el `status`
    del artefacto — **jamás infiere del texto** del chat.
  - **Cap de reparación explícito** (ej. 5) ADEMÁS de `--max-turns` = doble red. Estado terminal
    `blocked` → handoff humano (mapea a `await` del shell).
  - **Niveles anidados (clave):** la FSM `draft→ready→in_progress→in_review→done/blocked` (§13) es
    el ciclo de vida DENTRO de una caja T3; nuestro **spine `idea→done`** es cross-caja. La caja T3
    corre su FSM interna y solo al `done` interno avanza la transición del spine que posee.
  - Es el «gobierno de presupuesto» del backlog, hecho concreto.

## 5. Orquestación determinista + subagentes

- **A2 orquestación determinista — ADOPTAR.** La secuencia ENTRE cajas es código (el conductor
  ejecuta `ruta`/`si`), la agencia del modelo vive DENTRO de una caja. Resuelve una ambigüedad que
  VISION dejaba abierta (¿quién mueve el trabajo caja→caja?). Decisión de detalle para HS-07:
  ¿el `ruta` lo ejecuta el runner determinista o el conductor conversacional?
- **A14 «el padre no lee lo que delega» — ADOPTAR (check nuevo).** Tenemos el contrato de RETORNO
  del hijo (`<veredicto>→<path>`), NO la disciplina del PADRE. El «lenguaje defensivo de
  orquestación» («tu rol es ORQUESTACIÓN, NO leas los archivos objetivo, lístalos por nombre») es
  el check `orchestrator-no-read-delegated` (vive en el nodo headless/perfil-harness).
- **Los 6 patrones de subagentes — ADOPTAR como catálogo** (no doctrina dura): Delegated Data
  Access · Temp File Assembly · Shared-File · Hierarchical Lead-Worker · Persona-Driven Parallel
  (el auditor no es uno, son varios: adversarial-general + edge-case-hunter) · Evolutionary. El
  filesystem blackboard vive en modo worktree (ya en roadmap). ⚠️ No confundir «filesystem =
  verdad» de DAOP (artefactos de trabajo) con «JSONL de ~/.claude = verdad» nuestro (event-sourcing).
- **A15 degradación elegante — ADOPTAR (barato):** una caja multi-agente cae a ejecución
  secuencial si la maquinaria no está.

## 6. El seam organizacional (decidido: ArnesIA = producto puro + META de enganche)

**Decisión del operador (2026-07-05): ArnesIA NO posee L1.** El modelado del proceso de la empresa
(mapa de procesos, MOF, organigrama, «reporta a») vive en **OTRO sistema — futuro, separado**. La
salida de ese sistema será, un paso más adelante, el **INPUT para *crear* arneses**. ArnesIA es la
fábrica pura (crear·mapear·observar·mejorar), no la consultoría de modelado ni el cockpit
organizacional. Esto **evita el costo de consultoría por-cuenta** que DAOP advierte (4.2) y mantiene
el relato «medio de producción escalable».

**El seam que SÍ es nuestro — la META del arnés.** Cada arnés carga en su metadata **rol · proceso ·
reporta-a · empresa** (ya modelado en el objeto `arnes` de `graph.l0.schema.json`:
id·puesto·empresa·reporta_a·canal·marketplace). Esa META es el **contrato de enganche**: permite que
(a) múltiples arneses se compongan/relacionen y (b) el sistema-de-organigrama externo los consuma y
los produzca. **No construimos el organigrama; exponemos la superficie para que se enganche.**
Reencuadre de los ítems DAOP bajo esta decisión:

- **L1 · Modelado (MOF→grafo) — FUERA DE SCOPE de ArnesIA.** Concern del sistema externo. Lo que
  robamos de DAOP §1/§4 no es la consultoría sino la **disciplina del contrato de META** (qué campos
  debe cargar un arnés para ser enganchable). ISO/MOF ni se menciona en nuestra doctrina — es del
  otro sistema.
- **META de enganche — ENDURECER (nuestro).** Los campos rol·proceso·reporta-a·empresa pasan de
  metadata suelta a **contrato validado** (check: todo arnés declara su META de enganche completa).
  Es el punto de sutura con el futuro sistema L1, hoy ya presente en graph.l0.
- **Gate de fidelidad (§18.6) — ADOPTAR para NUESTROS arneses.** Honestidad de **PROCESO** (≠
  honestidad de DATO, METODOLOGIA §4): validar el flujo feliz contra la operación REAL, no la ideal.
  Como ArnesIA es producto puro, esto = **dogfooding elevado a gate de promoción** de los arneses que
  fabricamos, no auditoría del MOF de un cliente. Encaja con p6 (las excepciones son datos).
- **Reuso de cajas por referencia (A12) — ADAPTAR.** Cajas library-resident referenciadas por
  `id@version` para deduplicar sub-procesos compartidos entre arneses. El **manifiesto de rol /
  organigrama = del sistema externo**; ArnesIA solo garantiza que las cajas sean referenciables y que
  la META permita el enganche. No construimos el backbone cross-arnés; lo habilitamos.
- **Override 3 capas (§15) — ADAPTAR (CC-native/JSON, NO TOML).** Mismo arnés en N empresas sin fork:
  `user > org > base`, merge estructural resuelto en el `KitProvisioner` al hidratar (historias/…inyeccion-knowhow.md
  §8.1). Palanca de productización «1 arnés, N empresas» — 100% nuestra (es config de
  producto). Regla 15.2: identidad/secuencia NO customizable → «la app OPERA las primitivas, no las
  duplica».
- **Permisos = f(rol) del arnés (§16.1) — ADAPTAR (boundary nuevo).** El permission-set se parametriza
  por el rol que hidrata (mismo arnés, permisos distintos por rol), pero la **autoridad del rol viene
  de la META / del sistema externo**, no de una consultoría MOF nuestra. El spike `control_request`
  de HS-07 adopta la semántica de parametrización-por-rol. `iam-por-agente` (§16.2) = concern DevHub.

## 7. Lo que NO cambia — el moat (defender)

1. **El Mapa** (grafo visual, bandas Guardia/carriles/Base, capas Estructura/Tokens/Desempeño/
   Proceso, atribución traza→componente). DAOP es doctrina texto-only, cero modelo espacial.
2. **El loop cerrado de mejora continua** (crear→mapear→observar→mejorar). DAOP para en «published»
   + deprecación; no tiene loop que retroalimente telemetría de producción a re-forjar. Núcleo.
3. **Telemetría de nacimiento** (p9): constitucional, no opt-in. DAOP emite evidencia por-promoción.
4. **Honestidad de dato** (METODOLOGIA §4): procedencia, gris≠verde, «—», heat=percentil. DAOP no
   tiene NADA sobre cómo el producto muestra datos sin mentir.
5. **Inyección 3-cuerpos + `--plugin-dir` sin `--bare` + cero-config del usuario**
   (historias/…inyeccion-knowhow.md). DAOP no tiene mecanismo de distribución ni ve el problema auth-suscripción.
6. **Event-sourcing (JSONL nunca se parsea como API) + seguridad del daemon (superficie-local-
   confinada) + conductor Go local-first.** DAOP tiene cero modelo de sistema.
7. **knowledge/ árbol VIVO con cadencia semanal** anclado a Anthropic/expertos + 122 checks como
   linter. DAOP es un snapshot estático v0.2.
8. **p8 «estándar propio»** → el Mapa puede ASUMIR la constitución (DAOP debe servir arneses
   arbitrarios; nosotros cerramos el universo). Foco convertido en ventaja técnica.

## 8. Deltas as-code (qué se toca)

**Nodo nuevo en `knowledge/` (el mayor aporte estructural):**
- **`elements/harness-profile.md`** (nodo 12) — el «perfil de harness / loop / orquestación» que
  hoy no tiene casa. Aloja: tipos T1–T3, loop-iteration-cap, blocked-state, orchestrator-no-read,
  orchestrator-reads-status, los 6 patrones de subagentes, graceful-degradation, model-routing por
  carga cognitiva. (T4 = shell/sesión, se documenta como capa de runtime, no como caja.)

**Checks nuevos candidatos (para los 122):**
`no-phantom-frontmatter` ★ · `context-injection-native` ★ · `orchestrator-no-read-delegated` ★ ·
`skill-script-for-deterministic` · `skill-document-as-cache` · `persona-hard-constraint-hook` ·
`loop-iteration-cap` · `orchestrator-reads-status` · `cognitive-load-declared` ·
`single-writer-per-artifact` (escritor_unico) · `arquetipo-declarado` · `no-arnesar-clasificado`.
**+ del barrido externo (§11):** `explainability-rationale` · `mece-routing-coverage` ·
`handoff-trigger-declarado` · `tolerancia-declarada` · `autonomia-por-riesgo` · `snapshot-inmutable-
versionado` · `anotacion-humana-a-regresion` · `permiso-efimero-ttl` (spike control_request).

**Nodos que crecen:** skills.md (4 capas + scripts + doc-as-cache) · subagents.md (6 patrones +
padre-no-lee + schema≤200tok) · rules.md (persistent_facts→native + auto-memory shippable) ·
headless-sdk.md (loop-cap + blocked + reads-status).

**Boundaries nuevos candidatos (arch/, HS-07):** `permisos-derivan-del-rol-MOF` ·
`orquestacion-determinista-entre-cajas` (A2) · el `maquinaria-no-contamina-arnes` ya en draft.

**METODOLOGIA:** §3 supersedido por el contrato fusionado (§2 de este doc) · nuevos §: arquetipos
de trabajo · perfil de harness · document-as-cache · gate de fidelidad (honestidad de proceso) ·
cuándo NO arnesar (A13).

## 9. Impacto en roadmap (madurez)

Modelo de madurez 0–4 (DAOP §20) = lente útil. **Estamos en ~0.5–1:** FORMA exquisitamente madura
(11 principios + 122 + 89 checks) pero CERO arneses-producto construidos.

**DECIDIDO (operador, 2026-07-05): dogfood-first.** Forjar a mano el arnés dev-full-cycle end-to-end
(crear→mapear→observar→mejorar sobre nuestro PROPIO proceso) ANTES del compilador/Mapa (nivel 3), para
que el Mapa renderice datos REALES, no mocks. Reordena la fase 5: **el primer entregable de producto
es un arnés real de nivel 2**, y ése alimenta al Mapa. El arnés dev-full-cycle es además el candidato
a primer vendible (el que más dogfoodeamos).

## 10. Decisiones

### Resueltas (operador, 2026-07-05)
1. **Modelo de negocio / L1 → PRODUCTO PURO + META de enganche.** ArnesIA NO posee L1. El modelado
   del proceso/organigrama vive en OTRO sistema (futuro), cuya salida será el input para *crear*
   arneses. Cada arnés carga META (rol·proceso·reporta-a·empresa, ya en graph.l0 `arnes`) = el seam.
   Ver §6 reencuadrado.
2. **Roadmap → DOGFOOD-FIRST.** Arnés dev-full-cycle real end-to-end antes del Mapa/compilador.
   Ver §9.
3. **P6 vs HITL → FRONTERA POR SCOPE RATIFICADA.** La GUÍA de proceso/calidad nunca bloquea (P6, la
   excepción es dato de mejora); la banda Guardia/hooks SÍ bloquea (exit 2) efectos externos
   destructivos/secretos/alto-riesgo. Dos planos, sin contradicción. Requiere nota aditiva en VISION
   que acote el alcance de P6 (no reescribe el principio, lo precisa).

### Abiertas
4. **Persona-state / DevHub:** ¿persistencia de persona-de-rol (memoria dos niveles, sin teatro) en
   el lado ejecución, o sesiones stateless-por-arnés por ahora? (Concern DevHub, no bloquea la
   fábrica.)
5. **Nombre de la doctrina** (DAOP es placeholder; «fábrica de cajas» sigue siendo nuestro spine).
6. **Detalle A2 (HS-07):** ¿el `ruta`/`si` del contrato lo ejecuta el runner determinista o el
   conductor conversacional decide el hand-off?

## 11. Barrido externo (2026-07-05) — el linaje que nos saca de «clon de BMAD»

Segundo estudio (4 subagentes) sobre 7 fuentes externas que aportó el operador: 2 académicas
(arXiv 2504.03693 · el **manifiesto APM** de *Information Systems*), **Sierra AI** (ADLC + enterprise),
**Salesforce Agentforce** (patrones + arquitectura enterprise). **No relitigan nada firmado; dan a la
doctrina el LINAJE y el VOCABULARIO que la vuelven PROPIA, no una adaptación de BMAD.**

### 11.1 El reencuadre estratégico (la respuesta al «¿clon de BMAD?»)

**ArnesIA operacionaliza Agentic Business Process Management (APM)** — una disciplina de PROCESO,
INDEPENDIENTE DE DOMINIO, con ~30 años de linaje BPM. El manifiesto APM («Agentic BPM: A Research
Manifesto», 18 autores del establishment BPM — Dumas, Montali, Rinderle-Ma, Weber…) define el marco:
agentes *process-aware* cuya autonomía se **acota, alinea y opera mediante frames**. Cambia el árbol
de fuentes de la doctrina:
- **BMAD/DAOP** = UN input (mecanismos de autoría, filtrados por el firewall CC-native §3). **Ya no es
  el padre — es una de cuatro fuentes.**
- **APM (académico)** = la TEORÍA DE PROCESO que nos hace conceptualmente independientes de cualquier
  framework de agentes. *«Distinguimos autonomía de automatización… los sistemas APM facilitan
  autonomía, donde los agentes perciben, razonan y eligen cómo actuar dentro de un frame de proceso.»*
  Y la independencia de rubro, literal: *«dejamos intencionalmente abiertos los detalles de realización
  e implementación… process-awareness y framing.»*
- **Sierra ADLC** = validación de que crear→mapear→observar→mejorar es un **ciclo de vida de agente en
  producción** (clase-lifecycle), NO un kit de autoría (clase-BMAD). El valor vive en observar+mejorar.
- **Salesforce Agentforce** = validación del eje **despliegue-en-orgs-reales** (gobernanza operativa).

Posicionamiento: **«ArnesIA fabrica frames (arneses) que dan framed autonomy a Claude Code por
rol×proceso.»** Ancla a una literatura, no a un repo. Un arnés ES el framing mechanism instanciado
(frame normativo + operacional + framed-knowledge + tools) para un rol×proceso.

### 11.2 Upgrade de vocabulario (renombra mejor lo que ya teníamos — costo ~0)

| Nuestro término | Término APM (adoptar) | Qué gana |
|---|---|---|
| banda Guardia + permisos + rules | **frame normativo** (deóntico) + **operacional** | precisión + linaje; el firewall CC-native = «operacionalizar frames normativos sobre CC» |
| arquetipos pipeline vs abierto/excepción | **automation vs framed autonomy** | backbone conceptual |
| loop de mejora (FSM-en-caja vs spine) | **adaptation** (instancia, efímero) vs **evolution** (modelo, persistente) | nombra crisp los dos niveles |
| creación conversacional + dock | **conversational actionability** (Query·Recommend·Create·Execute) | describe qué hace la app sobre el arnés |

**Las 4+1 capacidades APM como checklist de doctrina** («¿este arnés provee las 4?»): framed autonomy ·
**explainability** · conversational actionability · self-modification.

### 11.3 El único gap conceptual nuevo — Explainability

Tenemos honestidad-de-DATO + atribución traza→componente (el Mapa), pero NO doctrina del
**por-qué-el-agente-decidió-X** (rationale accionable — que indique la corrección sin escalar). El
manifiesto lo eleva a capacidad de primera clase. **Cae en nuestro moat (observabilidad/Mapa)** → nueva
faceta evaluable + insumo del Mapa. Candidato a check.

### 11.4 Gap-fillers concretos (enriquecen observar/mejorar + gobernanza — zonas delgadas nuestras)

- **Immutable snapshot** (Sierra) — `arnés@version` = bundle atómico (código+prompts+modelo+knowledge+
  permisos) versionado, A/B-testeable, revertible. Hace concreto el «promote» difuso; refuerza A12.
- **Loop de anotación humana → eval de regresión** (Sierra «Experience Manager») — un SME anota una traza
  real y la anotación se CAPTURA como test de regresión. **El gap más concreto de observar→mejorar**
  (tenemos el dato, no el loop de captura). Superficie de anotación en la lente observar del Mapa.
- **Simulación conversacional vs mock pre-promoción** (Sierra) — eval-gate = correr el arnés contra
  entorno mock antes de promover. Ya tenemos el sustrato (JSONL replay); falta el gate.
- **Bounded-error-rate** (Sierra) — a escala el verde-absoluto es mentira; «systems problem, not
  whack-a-mole». Afina eval-gate + honestidad-de-dato (heat=percentil ya apunta ahí).
- **Tolerancia por capas** (Sierra) — cada `capability`/`constraint` declara su nivel; bajo→hook
  (bloquea), alto→guía. **Formaliza la frontera P6-vs-Guardia (§10.3) sin reescribir P6.**
- **Check MECE de routing** (Salesforce) — cada slice de proceso mapea a exactamente UNA caja (cobertura
  + no-solape). Nuevo check de conformidad del grafo.
- **Permisos efímeros con TTL** (Salesforce «task-based expiring permissions») — enriquece el spike
  `control_request` más allá del set estático por-rol (mínimo privilegio temporal).
- **Handoff-trigger explícito en el contrato** (Salesforce) — campo que declara *cuándo* escala a humano;
  hace ejecutable/auditable la frontera P6/Guardia.
- **Eje autonomía-por-riesgo** (arXiv 2504.03693) — ortogonal a T1–T3 y a arquetipos; el blast-radius del
  efecto externo decide cuánta autonomía. Aterriza en la banda Guardia.
- **Catálogo de métricas de gobernanza** (arXiv 2504.03693) — override-frequency · time-to-escalation ·
  exception-count · audit-completeness → señales de calibración-de-confianza para la capa Desempeño.

### 11.5 Validación de nuestro moat (convergencia independiente)

Cuatro fuentes convergen en tesis ya firmadas: **guardrails fuera del razonamiento** (Salesforce:
«enforced at the reasoning layer — the agent cannot reason its way around them» = banda Guardia/hooks
exit 2) · **handoff-por-referencia** (Salesforce «pass RecordId, never Name» = A14 `<veredicto>→<path>`
+ `escritor_unico`) · **orquestación determinista** · **jerarquía rol×proceso** (Jobs→Topics→Actions ≈
Arnés→Cajas→primitivas) · **audit constitucional** (≈ telemetría de nacimiento p9). Lo que NADIE externo
tiene: **el Mapa espacial**, el **loop cerrado de mejora**, **local-first sin lock-in**, la **inyección
`--plugin-dir`**, la **honestidad-de-dato**.

### 11.6 Qué NO adoptar (disciplina anti-bloat)

- **No academizar la doctrina entera** — 5–6 términos APM load-bearing, no el diccionario (deontic/
  aleatoric/epistemic → watchlist).
- **No la UX conversacional de Sierra** (voz/on-brand/latencia) — es para agentes customer-facing;
  nuestros arneses encapsulan uso interno de CC por rol.
- **No el lock-in de Salesforce** (Trust Layer, Data Cloud, Agent Builder, control-plane SaaS) — el
  principio se conserva, la maquinaria SaaS se descarta.
- **No cimentar en el arXiv 2504.03693** (n=22, cero experiencia operativa real) — direccional, no base.

## Fuentes

- Investigación externa: `~/Descargas/doctrina-kits-v0.2.md` (DAOP v0.2, BMAD v6 + Agent SDK).
- **Barrido externo (4 subagentes, 2026-07-05):** manifiesto APM «Agentic BPM: A Research Manifesto»
  (*Information Systems* vol.140, 2026, DOI 10.1016/j.is.2026.102738; preprint abierto arXiv 2603.18916;
  predecesor arXiv 2201.12855) · «Agentic BPM: Practitioner Perspectives on Agent Governance» (arXiv
  2504.03693) · Sierra AI: ADLC + enterprise-grade agents (sierra.ai/blog) · Salesforce Agentforce:
  agentic-patterns · enterprise-agentic-architecture · agentforce guide (architect.salesforce.com).
- Estudio 5-frentes (subagentes, 2026-07-05): constitución/VISION · metodología/contrato ·
  knowledge/CC-native · arch/conductor/runtime · estrategia/MOF/orgs. Convergencia fuerte en las 3
  tesis (§0). Detalle de cada frente disponible en el hilo de la sesión.
- Nuestras capas: `VISION.md` · `METODOLOGIA.md` · `knowledge/` (11 nodos, 122 checks) · `arch/`
  (14 boundaries, 89 checks) · `historias/2026-07-05-arquitectura-inyeccion-knowhow.md`.
