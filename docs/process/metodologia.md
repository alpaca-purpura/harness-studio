# ArnesIA — Metodología de arneses (cómo los creamos, mantenemos y qué debe tener cada componente)

> **Doctrina operativa VIVA** — crece con cada iteración de la fase UX (HS-03) y las que
> sigan. Norte constitucional: [`../product/vision.md`](../product/vision.md) (11 principios + anatomía A1–A7).
> Este documento **baja la anatomía a reglas de negocio concretas que el producto va a
> enforcar**. Registro de iteraciones y detalle de UX: [`../product/ux.md`](../product/ux.md). Cementado hasta la
> **iteración 13 de HS-03** (2026-07-05; §4 reconciliado con el retiro del andamiaje REAL vs
> DEMO del mockup); **§2 extendido 2026-07-05 — estructura de subagente · hook · rule/conocimiento
> cementada (derivada de `docs/architecture/knowledge/`); solo restan commands/mcp/plugins/settings/output-styles/
> statusline/headless).** **Doctrina v1 bajada as-code 2026-07-05 (§3 contrato fusionado + §8 doctrina
> de proceso / framed autonomy): operacionalizamos Agentic BPM, no clonamos BMAD — ver VISION §Linaje.**
>
> **Base de evidencia = [`docs/architecture/knowledge/`](../architecture/knowledge/INDEX.md) (árbol de conocimiento VIVO).**
> El «qué debe tener cada componente» (§2–3) **deriva** del estándar as code por elemento
> (skills, hooks, rules, subagents, commands, mcp, plugins, settings, output-styles, statusline,
> headless). Cada nodo del árbol lleva dos capas — **L1 estándar oficial+expertos** y **L2 nuestra
> adaptación** (obligada a derivar de L1) — y emite una **rúbrica de checks evaluables** (121 al
> corte fundacional 2026-07-04; **138 al corte HS-07** con el nodo `harness-profile`) que es el
> ruleset de conformidad de §6. El árbol se actualiza
> **cada semana** (mecanismo en [`docs/architecture/knowledge/CADENCE.md`](../architecture/knowledge/CADENCE.md)); cuando cambia,
> esta metodología se re-alinea. Regla dura: nuestra forma de trabajo no puede divergir del
> estándar de los expertos sin marcarlo y justificarlo.

## 0. ArnesIA es dueño de CREAR y de MANTENER

- ArnesIA no solo crea arneses: es el **único lugar donde se mantienen y se corrigen**
  (honra el principio 8 «estándar propio» y «ArnesIA es dueña única del observar y el
  modificar» de VISION).
- **Si un arnés o un skill no cumple el estándar, se carga en ArnesIA y se corrige AQUÍ.**
  No se parchea suelto en el proyecto del cliente: ArnesIA es la fábrica y el taller.
- **Nosotros definimos qué debe tener cada componente** (§2–3); ArnesIA lo verifica al
  importar/operar y lo **conforma** (reescribe a la estructura estándar).
- **Primer trabajo concreto acordado:** modificar los skills existentes —empezando por
  `luana-platform`— para que sigan la estructura que definamos. La primera pieza son los
  **contratos de caja** (§3); el prompt para arrancar esa conformación está en UX.md
  (iteración 10) y se ejecuta en kit-dev.

  > **Nota de supersesión (2026-07-07, sync HS-10):** este «primer trabajo» quedó SUPERADO por
  > **dogfood-first** (HS-07): el primer arnés real conformado fue **`dev-full-cycle`** (HS-08,
  > `dogfood/`), sobre nuestro propio proceso y ANTES del Mapa. `luana-platform` queda como
  > **candidato futuro** de conformación, no mandato.

## 1. El modelo — fábrica de cajas de proceso (resumen; detalle en VISION §Anatomía A1–A7)

Un arnés es una **fábrica**: el trabajo entra, cruza cajas y sale transformado.

- **Fase › caja › maquinaria.** Una **caja de proceso = una skill** (el frente de una etapa)
  que orquesta su **maquinaria dedicada** (agentes, sub-skills) y se apoya en infraestructura
  compartida (rules, hooks, knowledge).
- **Contrato input → output.** La salida de una caja es la entrada de la siguiente. El
  hand-off ES el contrato.
- **Dos estados, no confundir.** El **trabajo** lleva su estado (el spine idea→done); la
  **caja es dueña de UNA transición** de ese estado y además tiene su estado operativo
  (telemetría).
- **El eval-gate vive en el contrato de salida.**
- **La fábrica no es una recta:** hay **retrabajo** (aristas de retorno) y **cajas en
  paralelo** (una línea por variante: marca, tipo de trabajo).
- **La infraestructura compartida (Guardia/Base) vive en bandas,** no dentro de una caja.

## 2. Qué debe tener cada componente — estructura obligatoria (núcleo de fábrica FIRME)

> **Fuente de verdad = el árbol [`docs/architecture/knowledge/elements/`](../architecture/knowledge/INDEX.md).** Un nodo por
> elemento con L1 (estándar oficial+expertos, con fuente y fecha), L2 (nuestra adaptación) y su
> checklist evaluable. Lo de abajo es el **resumen de negocio**; el detalle vivo y las fuentes
> viven en el árbol. Cuando el árbol crezca (cadencia semanal), este resumen se re-alinea.

Reglas por tipo de componente. Se cementan a medida que las acordamos; **firme el núcleo de
fábrica — skill-caja · skill-apoyo · subagente · hook · rule/conocimiento — + el bloque de
contrato** (§3). Distribución/infra (commands · mcp · plugins · settings · output-styles ·
statusline · headless) siguen derivando del árbol, sin bajar aún a regla de negocio.
Estándar completo por elemento en el árbol: skills-caja/
apoyo ([`skills`](../architecture/knowledge/elements/skills.md)) · Guardia ([`hooks`](../architecture/knowledge/elements/hooks.md),
[`settings-permissions`](../architecture/knowledge/elements/settings-permissions.md)) · Base
([`rules`](../architecture/knowledge/elements/rules.md)) · maquinaria ([`subagents`](../architecture/knowledge/elements/subagents.md)) ·
[`commands`](../architecture/knowledge/elements/commands.md) · terceros ([`mcp`](../architecture/knowledge/elements/mcp.md)) ·
distribución ([`plugins`](../architecture/knowledge/elements/plugins.md)) ·
[`output-styles`](../architecture/knowledge/elements/output-styles.md) ·
[`statusline`](../architecture/knowledge/elements/statusline.md) · motor conductor ([`headless-sdk`](../architecture/knowledge/elements/headless-sdk.md)) ·
**perfil de harness** ([`harness-profile`](../architecture/knowledge/elements/harness-profile.md), nodo 12 — cómo la
caja ejecuta el loop / subagentes / routing; doctrina v1 §8.2).

> **Firewall CC-native (doctrina v1 §8.6):** todo frontmatter usa solo claves que Claude Code
> reconoce. Prohibido `persistent_facts` / `activation_steps_prepend` / `customize.toml` / sanctum
> (CC los ignora en silencio → hacen CERO). Check `no-phantom-frontmatter`. Nos mantiene doctrina
> PROPIA anclada a CC.

**Skill que es CAJA de proceso:**
- Frontmatter: `name`, `description`, `version`, `model`, `clase` (contrato L0 `meta.clase`,
  I-75) y el bloque **`contract:`** (§3).
- Prosa mínima: `## Cuándo` (precondición = su input), `## Pasos`, `## Guardarraíles`.

**Skill de APOYO** (librería experta · utilidad · tercero · meta-harness):
- Frontmatter con `contract.caja: false` + `rol:` (`libreria-experta | utilidad | tercero |
  meta-harness`). No es una caja: no declara transición de estado ni gate.

**Subagente = MAQUINARIA de una caja** (no es caja; ⇐ [`subagents`](../architecture/knowledge/elements/subagents.md) L2):
- Frontmatter: `name` (lowercase-hyphen, único por scope) · `description` **con condición de
  disparo concreta** (no rol vago — la auto-delegación rutea por aquí) · `tools` **allowlist
  explícita** (jamás omitido = acceso total incl. MCP) · `model` deliberado por rol (no `inherit`
  silencioso). `isolation: worktree` si escribe en cajas paralelas (A5).
- Cuerpo (system prompt): **contrato de retorno anti-telephone** obligatorio — declara la salida
  como `<veredicto> → <path>` (o estructurado equivalente), **nunca «resume / investiga»**.
- NO declara `contract.estado` ni `gate`: sirve a la caja que lo invoca (aparece como
  `de: maquinaria:<agente>` en el contrato de ésa).
- Honestidad / hallazgos: si tiene `Write/Edit/Bash` el cuerpo justifica modificar (vs
  describirse «solo-lectura») · descripciones solapadas → consolidar (auto-delegación rota) ·
  0 lanzamientos 30d con caja activa = **maquinaria muerta** (poda) · `bypassPermissions` = flag
  de Guardia.

**Hook = GUARDIA** (banda transversal, actúa sobre TODAS las cajas — no vive dentro de una;
⇐ [`hooks`](../architecture/knowledge/elements/hooks.md) L2):
- Estructura: declara **evento** + matcher (exacto / lista `|` / regex; **nunca `*` para
  auto-`allow`**) + **type** (`command` en prod; `agent` experimental = evitar). Vive en config
  **versionada** de fuente confiable (lección CVE-2025-59536).
- Bloqueo: **exit `2`** (exit `1` para «bloquear» = bug silencioso, no bloquea nada).
- Seguridad dura: `"$VAR"` con comillas · ruta absoluta o `${CLAUDE_PROJECT_DIR}` · guardia de
  `.env` / `.git/` / keys · **PreToolUse liviano** (<500 ms; scans/tests pesados → PostToolUse/async).
- Par **regla ↔ hook**: toda «nunca X» destructiva / de-secreto de la banda Base tiene su
  PreToolUse que la enforca — sin él la regla es solo advisory (hallazgo).
- `telemetry-emit` (KIT-03, OTLP) = el sensor: **nace en todo arnés** (telemetría de nacimiento,
  principio 9) — no es opt-in.

**Rule / conocimiento = BASE** (banda always-on, infraestructura compartida — no una caja;
⇐ [`rules`](../architecture/knowledge/elements/rules.md) L2):
- Dos ámbitos, no confundir: **always-on** (`CLAUDE.md` / rules sin scope, se paga CADA turno)
  vs **conocimiento scoped** (`.claude/rules/` con `paths:` — carga solo al tocar archivos que
  matchean; retrieval just-in-time > pre-load).
- Estructura: específico > verboso · headers/bullets · `file:line` en vez de pegar código (se
  pone stale) · `@import` que resuelve, depth ≤4, sin ciclo · AGENTS.md por `@import`/symlink,
  **nunca duplicado** (CC no lo lee nativo).
- Presupuesto medible (principio 11): `CLAUDE.md` + rules sin scope **<~200 líneas**; total
  always-on bajo el techo, medido y pintado en banda Base — superarlo = hallazgo.
- Honestidad: **sin contradicciones cross-capa** (error — el modelo elige arbitrario) · regla
  dura ⇒ hook (ver Guardia) · procedimiento multi-paso → **skill**, no rule.

## 3. El contrato de caja — schema fusionado (it.10 + doctrina v1, 2026-07-05)

Bloque `contract:` en el frontmatter del SKILL.md. **Un solo contrato con tres ejes
perpendiculares:** INTENCIÓN (qué promete la caja), CABLEADO (cómo se conecta) y ACEPTACIÓN
(cómo se prueba su salida). Es lo que vuelve reales el eval-gate por-caja, la detección de
precondición (guía sin bloqueo) y la validación de composición (huérfanos/mismatch). El eje de
intención + aceptación se sumó de la doctrina v1 (SPEC-kernel + Gherkin ejecutable); el cableado
es el §3 original de it.10, intacto. Schema validado por
[`docs/architecture/contracts/schema/box.contract.schema.json`](../architecture/contracts/schema/box.contract.schema.json).

```yaml
contract:
  # ── INTENCIÓN (qué promete esta parte-de-proceso) ──
  why: "<propósito inmutable>"                 # el "goal"; blinda la deriva
  capabilities:
    - id: CAP-01                               # IDs estables entre versiones (semver del contrato)
      what: "<qué logra — WHAT, no HOW>"
      success: "<señal concreta y verificable>"
  constraints: ["<solo los que doblan decisiones>"]
  non_goals:  ["<lo que explícitamente NO hace>"]

  # ── CLASIFICACIÓN (tres ejes ortogonales — clase ⊥ arquetipo ⊥ perfil; ver §8) ──
  clase: <skill|subagent|hook|rule|command|mcp|plugin|settings|output-style|statusline>  # meta.clase L0 (I-75) — qué ELEMENTO es (10 primitivas CC-native canónicas; B3)
  arquetipo: <pipeline|excepcion|abierto|no-arnesar>       # FORMA del trabajo (§8.1)
  perfil_harness: <T1|T2|T3>                               # cómo EJECUTA (§8.2; T4 = shell, no caja)

  # ── CABLEADO (it.10, intacto) ──
  caja: true
  fase: <id-de-fase>
  estado: "<estado_entra> -> <estado_sale>"    # la transición que posee (spine de estados)
  necesita:
    - art: "<artefacto o precondición>"
      de: "usuario | base:<id> | caja:<skill> | libreria:<skill> | maquinaria:<agente> | terceros:<skill> | marcas-dormidas:<skill>"
      requerido: true
  entrega:
    - art: "<artefacto>"
      escritor_unico: true                     # un solo escritor autorizado por artefacto (mutation contract)
  ruta:
    - a: "<skill | rol>"
      si: "<condición>"                        # omitir para el happy path

  # ── ACEPTACIÓN (cómo se evalúa la SALIDA — A4 — HONESTO) ──
  gate:
    tipo: "auto | manual | parcial | none"
    detalle: "<cómo; si none, POR QUÉ falta>"
    aceptacion:                                # Gherkin = gate.detalle hecho ejecutable
      - given: "<precondición>"
        when:  "<acción>"
        then:  "<invariante verificable>"
    evidencia: "<registro de auditoría emitido — telemetría de nacimiento (p9)>"
  handoff:                                     # cuándo escala a humano (frontera P6/Guardia hecha dato)
    cuando: "<condición de alto riesgo / no-convergencia>"
    a: "humano | caja:<skill>"
```

**Reglas del contrato fusionado:**
- **Honestidad del gate (intacta):** `tipo: none` cuando la caja NO tiene eval real — jamás se
  fabrica un eval. El hueco se ve (principio 10 hecho dato). **Existir tolera `none`; PROMOVER
  exige gate verde** (dos umbrales distintos, no se contradicen).
- **Gherkin ejecutable:** cuando `tipo: auto`, el `aceptacion` (Gherkin) ES el eval.
- **`escritor_unico`:** un artefacto = un solo escritor autorizado. Dos cajas escribiendo el
  mismo `art` = hallazgo de conformidad.
- **`arquetipo` obligatorio:** clasifica la forma del trabajo — arregla que el contrato dejaba de
  lado el trabajo abierto/no-arnesable (antes solo cabía como `gate: none`). Ver §8.1.
- **`necesita.de` acepta 7 orígenes** *(sync 2026-07-07, HS-10: la prosa corría con 5; el schema ya
  aceptaba 7 — las bandas legítimas del mapa)*. A los 5 de it.10 (usuario · base · caja · libreria ·
  maquinaria) se suman:
  - **`terceros:<skill>`** — el input lo provee una skill/integración de terceros (banda Terceros
    del mapa, it.8; p.ej. Clerk): dependencia externa declarada, no maquinaria propia.
  - **`marcas-dormidas:<skill>`** — el input referencia una variante de marca dormida (banda Marcas
    dormidas; origen cementado en el Gate 1 del Mapa, HS-09): la dependencia se declara honesta sin
    fabricar actividad de un componente inactivo.

## 4. Reglas de honestidad y medición (acordadas iteraciones 6–8)

Cómo el producto muestra datos, para que jamás mienta:

- **Procedencia obligatoria.** Carga de tokens = `bytes/4` del fuente; consumo = suma real
  del JSONL (30d); **gris ≠ verde** (sin datos ≠ sano).
- **No inventar.** Éxito, p95 o eval por componente = **«—»** cuando no está medido; el
  arnés que tiene telemetría viva pero sin criterio de éxito queda en estado **«señales
  incompletas»**, no verde.
- **Heat = percentil del arnés** (p50/75/90/97), no umbral absoluto — escala igual un arnés
  de 2M o de 80M tokens/30d.
- **Consumo medido pisa al estimado.** La fórmula solo cubre lo no medido (reglas always-on).
- **Atribución solapada declarada:** knowledge y reglas cuentan dentro de quien las carga →
  Σ ≠ total facturado. Las **librerías preloaded no cuentan heat** («se cuenta en quien la
  usa»).
- **Contrato de datos por capa:** cada capa del mapa = una pregunta + datos exactos +
  procedencia + umbral documentado (Estructura · Tokens · Desempeño · Proceso).
- **Datos con procedencia, sin split demo/real** (andamiaje retirado it.13, 2026-07-05): el
  producto opera **solo arneses propios** (VISION §8) — no hay tarjetas de demostración que
  separar; cada arnés carga la procedencia de sus datos y los KPIs agregan datos reales por
  construcción. El chip REAL/DEMO, el KPI «solo reales» y el orden real-primero eran andamiaje
  del mockup para distinguir luana de tarjetas ficticias; mueren con el mockup. La honestidad
  (nada se inventa · gris ≠ verde · «—» sin medir) sigue **intacta**.

## 5. Insights de producto que guían el diseño (emergentes)

- **A escala (100+ nodos), el filtro es navegación de primera clase.** El mapa necesita
  filtrar por tipo / caliente-frío / banda como ciudadanos de primera, no como adorno.
  (Descubierto con el arnés real de 141 nodos.)
- **La superficie fría es un hallazgo.** Skills sin uso, marcas dormidas, terceros,
  deprecadas — hay que distinguirlas y el arnés debe cazar el bloat (poda / lazy-load).
- **Dos vistas distintas conviven:** el **mapa de COMPONENTES** (piezas del arnés) vs el
  **tablero de FLUJO del trabajo** (spine de estados del arnés —ejemplo: el del arnés luana—, WIP
  caps, historias moviéndose; el producto es agnóstico: no fija un número de estados). La
  capa Proceso muestra el flujo de contratos; el tablero de estados como vista aparte sigue
  **en debate**.
- **El sensor ya existe:** el `telemetry-emit` del kit (KIT-03, OTLP GenAI-semconv) es el
  sensor que ArnesIA consume — no se instrumenta de cero, se conecta su egress.

## 6. Proceso de conformación (cargar → verificar → corregir → remapear)

1. **Cargar/importar** un arnés a ArnesIA.
2. **Chequeo de conformidad** contra §2–3 (¿cada caja tiene contrato? ¿gates declarados?
   ¿estructura de frontmatter? ¿huérfanos?).
3. **Reporte de brechas:** cajas con gate `none`, inputs sin productor upstream (huérfanos),
   outputs que nadie consume (dead-ends), skills sin contrato.
4. **Corregir AQUÍ** — ArnesIA reescribe el componente a la estructura estándar.
5. **Re-extraer contratos → remapear** con el match correcto.

Loop en curso: el prompt de kit-dev (UX.md it.10) formaliza los contratos en los SKILL.md y
emite `docs/process/contracts.index.yaml` + `contracts-gaps.md`, que ArnesIA extrae para
reemplazar los contratos inferidos por los reales.

> **Nota de supersesión (2026-07-07, sync HS-10):** ese loop de kit-dev/luana quedó SUPERADO por
> **dogfood-first** (HS-07): los contratos reales nacieron en el arnés `dev-full-cycle` (HS-08),
> validados de verdad por `arnesia conformance` (ruta `--arnes`, 13/13 verde). Luana queda como
> candidato futuro de conformación; el prompt it.10 se conserva como historia.

## 7. El estándar as code es un árbol vivo (no un doc congelado)

- **ArnesIA fija un paradigma propio (principio 8), pero anclado a las mejores prácticas
  vigentes de Anthropic y los expertos** — no a una opinión estática. Ese anclaje vive en
  [`docs/architecture/knowledge/`](../architecture/knowledge/INDEX.md) como árbol versionado: L1 (evidencia oficial+experta,
  fechada y con fuente) + L2 (nuestra adaptación, que **deriva** de L1) + checks evaluables.
- **Se investiga y actualiza cada semana** (nuevos comandos tipo `/goal`, features, eventos de
  hook, deprecaciones). El árbol **crece y adiciona**, no se reescribe (lo viejo se marca
  `deprecado`). Mecanismo en [`docs/architecture/knowledge/CADENCE.md`](../architecture/knowledge/CADENCE.md).
- **Excepción a «firmado = congelado» (decidida 2026-07-05):** HS-03 está **firmada**, pero la
  firma congela la **fase**, no estos documentos: `UX.md` y esta metodología quedan **docs VIVOS**
  por decisión del operador — siguen creciendo con nuevas iteraciones. **El árbol de conocimiento
  tampoco se congela** — por diseño sigue al ecosistema. Un arnés que ayer cumplía puede necesitar
  mejora hoy porque salió algo nuevo: ése es exactamente el «punto de mejora» que el mapa muestra.
- **Los 138 checks son el ruleset de conformidad (§6)** hecho dato: el linter que ArnesIA **YA corre**
  (`arnesia conformance`, construido en HS-08). La ruta `--arnes` da veredictos deterministas reales
  (schema + spine + firewall + escritor-único, verde en el dogfood, con test adversarial); la mayoría del
  ruleset de knowledge queda `deferred` hasta cablear sus enforcers (linters externos, nl-judge). Es la
  fuente de los badges de mejora que la UX empieza a pintar (it.11+).

## 8. Doctrina de proceso — framed autonomy (doctrina v1, 2026-07-05)

> Bajada as-code de [`historias/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`](../product/research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md),
> ratificada por el operador. **Operacionalizamos Agentic BPM** (VISION §Linaje): un arnés da
> *framed autonomy* a Claude Code por rol×proceso. Detalle por elemento del perfil de harness =
> nodo nuevo [`docs/architecture/knowledge/elements/harness-profile.md`](../architecture/knowledge/elements/harness-profile.md).

### 8.1 Arquetipos de trabajo (la FORMA — autonomía ≠ automatización)

Toda caja declara su `arquetipo`. Cuatro:
- **pipeline** — determinista / verificable. Output = artefacto exacto; gate auto; document-as-cache
  estricto. (Es **automatización**.)
- **excepcion** — flujo feliz + casos raros. Output = invariantes; gate en bifurcaciones; checkpoints
  en handoffs. (**Framed autonomy** acotada.)
- **abierto** — generativo / relacional. Restringe *scope*, no *pasos*; gate = guardrails + aceptación
  (parcial/manual); puede acumular estado de sesión (excepción a document-as-cache estricto). (**Framed
  autonomy** plena.)
- **no-arnesar** — juicio puro / político / relacional que no descompone en partes con contrato. **NO
  es una caja:** se clasifica explícitamente y sale del grafo (asistente general). Saber cuándo NO
  arnesar es doctrina, no omisión.

### 8.2 Perfil de harness (cómo EJECUTA — eje ortogonal a `clase`)

Toda caja declara su `perfil_harness`. `clase` dice qué ELEMENTO es (skill/hook/…); el perfil dice
cómo corre el loop:
- **T1 · Tarea** — un pase, sin loop, sin subagente; modelo barato, effort bajo.
- **T2 · Workflow** — multi-paso, stateful, a menudo interactivo; **document-as-cache** obligatorio (§8.3).
- **T3 · Worker autónomo** — loop desatendido; **el conductor Go dueña el loop** (NO Agent SDK):
  spawnea `claude -p --max-turns N`, lee el evento `result` de stream-json + el `status` del artefacto
  (jamás infiere del texto), aplica **cap de reparación explícito** además de `--max-turns`, y termina
  en `blocked` → handoff. La FSM interna `draft→…→blocked` (**adaptation**) anida DENTRO de la caja; el
  spine `idea→done` (**evolution** / cross-caja) solo avanza al `done` interno.
- **T4 = shell / sesión, NO una caja.** El agente-de-rol persistente (menú, persona) es la sesión del
  shell (frente N:1 con arnés, HS-06), no una caja de fábrica.

Detalle, patrones de subagentes y checks del perfil = nodo `harness-profile.md`.

### 8.3 Document-as-cache — el estado vive en el artefacto (DAOP-A7 · document-as-cache de BMAD)

El estado del TRABAJO vive en un artefacto durable (frontmatter YAML: inputs · `status` · timestamps
+ secciones-borrador), NO en la conversación. Sobrevive compactación (la etapa siguiente relee el
doc), habilita pausa/resume/reset limpio. CC no «resume del doc solo» → es convención con disciplina:
la caja/conductor DEBE releerlo.

**Precedencia `arquetipo` > `perfil_harness` (resuelve el cruce abierto×T2 — hereda DAOP `[PENDIENTE
7.A]`):** document-as-cache estricto es **obligatorio para `pipeline`/`excepcion` en T2/T3**. Una caja
`arquetipo: abierto` está **EXENTA** del document-as-cache estricto aun si su perfil es T2/T3 —
acumula estado de sesión y **destila al cierre** (§8.1). Cuando `arquetipo` y `perfil_harness`
apuntan a exigencias distintas, manda `arquetipo` (es la FORMA del trabajo). Operacionalizado en el
dominio como `RequiereDocumentAsCache(arquetipo, perfil)`.

> **Nota de linaje (no clon de BMAD):** el concepto *document-as-cache* viene de **BMAD** vía DAOP
> (axioma **DAOP-A7**), anclado en 12-Factor «unify execution & business state» — **no** es un tenet
> del manifiesto académico Agentic BPM. El manifiesto APM aporta *framed autonomy*, *autonomía≠
> automatización*, las **4 capacidades** (framed autonomy · explainability · conversational
> actionability · self-modification) y *adaptation/evolution*; los mecanismos de framework
> (document-as-cache, orquestación determinista) vienen de BMAD/12-Factor. Ver §8.6 (firewall) y la
> auditoría B7.

### 8.4 Gate de fidelidad — honestidad de PROCESO (≠ honestidad de DATO §4)

La §4 protege que los NÚMEROS no mientan; ésta protege que el FLUJO no encode la operación idealizada.
Validar el camino feliz de la caja contra el **uso REAL** del rol, no el ideal documentado. Como
ArnesIA es producto puro, esto = **dogfooding elevado a gate de promoción** de nuestros arneses.
Encaja con p6 (las excepciones son datos de mejora).

**Procedimiento (borrador operable):** en el `promote` de un arnés (release train KIT-06), (1) se toma
una traza REAL de una corrida del rol (JSONL del arnés dogfood), (2) se compara el camino recorrido
contra el `ruta`/`estado` declarado en los contratos de sus cajas, (3) toda divergencia (una caja que
el uso real saltó, un handoff que nunca se disparó, una fase muerta) es un **hallazgo de fidelidad** que
bloquea la promoción hasta reconciliar contrato↔uso. El eval vive junto al arnés; lo dispara el gate de
`promote`. **Estado: `deferred` (HS-09)** — hoy es procedimiento escrito, NO un check ejecutable
(`mecanismo: nl-judge`); no hay `enforced_by` que fabrique un veredicto. Honesto: gap declarado, no
disfrazado. Su realización depende de tener trazas reales del dogfood corriendo end-to-end.

### 8.5 Frontera P6 ↔ Guardia + tolerancia por capas

Ratificado (VISION §Linaje): la **guía** de proceso/calidad nunca bloquea (p6); la **banda Guardia**
SÍ bloquea (exit 2) efectos externos peligrosos + HITL. Cada `capability`/`constraint` del contrato
puede declarar su **nivel de tolerancia**: bajo → hook que bloquea; alto → guía que advierte. El
`handoff` del contrato (§3) hace ejecutable el *cuándo escala a humano*.

### 8.6 Firewall CC-native — no clonar mecanismos que Claude Code ignora

Todo lo que declaramos debe ser **primitiva nativa de Claude Code** (verificado en `docs/architecture/knowledge/`).
Prohibido escribir en el frontmatter claves que CC ignora en silencio (harían CERO):
`persistent_facts`, `activation_steps_prepend`, `customize.toml`, sanctum PERSONA/CREED. El
conocimiento estático entra por `CLAUDE.md` / rules con `paths:` / `@import` / `SessionStart` hook; la
memoria por auto-memory nativa (≤200 líneas). Check `no-phantom-frontmatter` ([`skills`](../architecture/knowledge/elements/skills.md)).
**Es lo que nos mantiene doctrina PROPIA anclada a CC, no adaptación de un framework ajeno.**

**Para profundizar (bibliografía):** manifiesto Agentic BPM ([arXiv 2603.18916](https://arxiv.org/abs/2603.18916)
· *Information Systems* 2026, DOI [10.1016/j.is.2026.102738](https://doi.org/10.1016/j.is.2026.102738))
· gobernanza práctica ([arXiv 2504.03693](https://arxiv.org/abs/2504.03693)) · Sierra ADLC
(https://sierra.ai/blog/agent-development-life-cycle · .../enterprise-grade-agents) · Salesforce
Agentforce (https://architect.salesforce.com/docs/architect/fundamentals/guide/agentic-patterns.html)
· 12-Factor Agents (https://github.com/humanlayer/12-factor-agents) · DAOP v0.2 (insumo local, filtrado).

## 9. Los 3 cuerpos — posesión e inyección del know-how (FIRMADO 2026-07-07, HS-10)

> Definición canónica — hasta esta firma solo vivía como investigación. Detalle y evidencia:
> [`historias/2026-07-05-arquitectura-inyeccion-knowhow.md`](../product/research/2026-07-05-arquitectura-inyeccion-knowhow.md)
> (VIGENTE por la firma HS-10) · nomenclatura de reconocimiento archivo→grafo:
> [`docs/architecture/contracts/nomenclatura-arnes.md`](../architecture/contracts/nomenclatura-arnes.md) (v1 FIRMADA).

El know-how de la fábrica vive en TRES cuerpos, cada uno con su posesión y su vía de inyección:

- **① Doctrina** — `docs/architecture/knowledge/` + `arch/` (el estándar as code, §7). Se compila y viaja
  **`go:embed` dentro del binario** `arnesia`: el conformance es portable a cualquier máquina,
  sin repo fuente ni contexto LLM.
- **② Maquinaria / kit** — las skills + overlay que **ENCARNAN** la doctrina (el saber-hacer de
  fábrica). Se **materializa a `~/.arnesia/`** y se **inyecta al `claude` spawneado por flags**
  (plugin CC propio, session-scoped, progressive disclosure); jamás se instala en el proyecto
  del arnés.
- **③ Arnés-producto** — el vendible: su repo / su `.claude/`. Se **edita vía cwd** (la sesión
  CC trabaja dentro de él, HS-06 S2) y se **publica al marketplace git** (release train KIT-06).

**Regla dura: ② jamás se escribe en el árbol de ③.** La maquinaria de fábrica no se filtra al
producto — lo que ③ contiene es solo suyo, más lo que el provisioning ESTAMPA deliberadamente
(p.ej. el facet `origen` del grafo). Nota de resolución: el schema L0
(`graph.l0.schema.json`, campo `origen`) cita «los 3 cuerpos» — esta sección es su definición
canónica.

## 10. Disciplina de desarrollo por paquete de trabajo (FIRMADA 2026-07-07)

> Origen: orden del operador (sesión inspector-drawer, HS-09/Hito 2): TODA funcionalidad
> nueva se desarrolla por este flujo, y NADA vive solo en la conversación — una sesión
> nueva retoma como si fuera la misma. Paquete de referencia (plantilla viva):
> [`historias/2026-07-07-inspector-drawer/`](../product/stories/2026-07-07-inspector-drawer/INDEX.md).
> **Carpeta renombrada `research/` → `historias/` (HS-15, 2026-07-08):** nombre más intuitivo
> para un humano que supervisa (calca `vitalia/docs/product/stories/`); el flujo y los
> archivos del paquete no cambian, solo el nombre del contenedor.

**Unidad = paquete de trabajo**: carpeta `historias/AAAA-MM-DD-<slug>/` por funcionalidad.
El código NO se toca hasta que el paquete lo autorice (specs firmados).

### Los archivos del paquete — qué contiene cada uno y CUÁNDO se llena

| Archivo | Qué contiene | Cuándo se llena |
|---|---|---|
| `INDEX.md` | Encuadre · flujo con gates · checklist **Estado** · sección **«Retomar aquí»** (último hecho · próximo paso concreto · firmas pendientes) | Nace con el paquete; «Retomar aquí» se actualiza **al cierre de cada turno de trabajo** |
| `mockup-*.html` | Clon iterable de la superficie, con changelog de iteraciones en comentario de cabecera | Antes de conversar cambios; una versión por iteración |
| `decisiones.md` | Una entrada por decisión: qué · porqué · estado PROPUESTA→FIRMADA | **En el MISMO turno en que se conversa** — jamás al final |
| `spec.md` | El QUÉ: RF numerados + Gherkin, **cada RF trazado a `mockup:línea`** | Tras la firma del mockup |
| `design.md` | El UI al pixel: secciones/orden · tabla de campos · tokens por marca · estados | Junto con spec.md |
| `PARIDAD.md` | Matriz mockup ↔ componente real ↔ story=test ↔ RF | Durante la implementación, fila por fila |

### El flujo (gates humanos 🧑‍⚖️, en orden)

1. **Mockup** → iterar con el operador → 🧑‍⚖️ firma del mockup.
2. **Decisiones** → cada cambio conversado queda en `decisiones.md` al instante.
3. **Specs** (`spec.md` + `design.md`) → 🧑‍⚖️ firma del paquete.
4. **Implementación** contra el spec firmado; story=test por marca nueva; gates
   técnicos verdes (tsc · biome · depcruise · stylelint · steiger · vitest · go race · lint).
5. **Paridad** → 🧑‍⚖️ gate final: click-through app vs mockup lado a lado, consola
   limpia, screenshots revisados.

### Reglas duras de continuidad (anti-pérdida de contexto)

1. **La conversación jamás es el único registro.** Toda decisión, hallazgo u orden del
   operador se escribe en el archivo del paquete que corresponde EN EL MISMO TURNO.
2. **«Retomar aquí» siempre al día.** Si el operador dice «seguimos en otra conversación»,
   el INDEX.md ya lo contiene todo; la actualización es continua, no un ritual de cierre.
3. **Cada iteración firmada se commitea a main** (trunk-based) — git es la memoria durable.
4. **Sesión nueva arranca así:** `CLAUDE.md` (router puro) → [`checkpoint.md`](../product/checkpoint.md) nombra
   el paquete activo → leer su INDEX.md («Retomar aquí») → `decisiones.md` → continuar
   exactamente donde quedó. *(Reorg 4-ejes HS-18: el backlog GLOBAL y los gates pendientes
   viven en [`BACKLOG.md`](../product/BACKLOG.md), el «ahora» en `checkpoint.md`; el INDEX del paquete sigue
   siendo el puntero de continuidad DEL PAQUETE, no del proyecto.)*
5. **Mockups fieles o no sirven:** tokens DTCG reales (norma «pegarse al Storybook»),
   datos REALES (showcase/dogfood, jamás inventados), publicar siempre al MISMO artifact.
6. **Nada llega al código sin spec firmado; nada se firma sin verse** (click-through con
   asserts + screenshots + consola limpia — disciplina UX.md, aquí obligatoria por fase).
7. **Ningún cambio de código sin capability** (FIRMADO 2026-07-09, HS-18). Todo commit que toca
   fuente (`cmd/` · `internal/` · `web/src`) construye o modifica un capability en
   [`CAPABILITIES.md`](../product/capabilities/INDEX.md) — el **SSoT funcional**: qué HACE el sistema, con
   puntero al código autoritativo (`file#Símbolo`) y a su check. La historia de usuario es el
   *delta*; el capability es el *saldo*. Doctrina **enforced**:
   [`docs/architecture/boundaries/codigo-traza-a-capability.md`](../architecture/boundaries/codigo-traza-a-capability.md)
   (Living Documentation + Business Capability Map; validador R1/R2 + job lefthook `capabilities`).

## Estado

Documento vivo. **HS-03 FIRMADA** (it.13, 2026-07-05; §4 al día: retirado el andamiaje REAL vs
DEMO del mockup; la honestidad sigue). **§2 extendido 2026-07-05:** cementada la estructura
obligatoria de **subagente · hook · rule/conocimiento** (derivada de los nodos `docs/architecture/knowledge/` ya
firmados — no doctrina nueva, promoción de checks L2 a regla de negocio) → el núcleo de la
fábrica (skill · maquinaria · Guardia · Base) queda firme; restan solo los elementos de
distribución/infra (commands · mcp · plugins · settings · output-styles · statusline · headless).
Por **excepción declarada** (§7), la firma congela la fase pero **NO** esta metodología ni la UX:
ambas siguen creciendo con los comentarios del operador (quedan muchos) y son la base de las specs
de fase 4 (HS-08). El árbol [`docs/architecture/knowledge/`](../architecture/knowledge/INDEX.md) también sigue vivo por diseño (§7).
**Doctrina v1 (2026-07-05, ficha HS-07):** cruce de DAOP/BMAD + barrido de 7 fuentes externas (manifiesto
Agentic BPM, Sierra ADLC, Salesforce Agentforce) → bajada as-code: §3 contrato fusionado · §8 doctrina de
proceso (framed autonomy) · nodo `harness-profile` (nº12) · nota de linaje en VISION · 2 boundaries nuevos.
Reencuadre: **operacionalizamos Agentic BPM, no clonamos un framework**; la doctrina es PROPIA, basada en
proceso e independiente de rubro.
**Sync 2026-07-07 (HS-10):** §3 `necesita.de` al día con el schema (5→7 orígenes: +`terceros:` ·
+`marcas-dormidas:`) · notas de supersesión en §0 y §6 (dogfood-first: el primer arnés real fue
`dev-full-cycle`, HS-08; luana = candidato futuro, no mandato) · nueva **§9 «Los 3 cuerpos»**
(posesión e inyección del know-how — definición canónica FIRMADA HS-10, antes solo en investigación).
**§10 «Disciplina de desarrollo por paquete de trabajo» (2026-07-07, orden del operador):** toda
funcionalidad nueva itera por mockup en `historias/<fecha>-<slug>/` con decisiones/spec/design/PARIDAD
y reglas de continuidad entre sesiones — la conversación jamás es el único registro.
**Sync 2026-07-09 (HS-18, paquete `reorg-docs`):** alineación con el **árbol de docs de 4 ejes** —
[`BACKLOG.md`](../product/BACKLOG.md) (lo-que-viene) · [`checkpoint.md`](../product/checkpoint.md) (ahora + cifras GENERADAS por
`arnesia conformance`, no tecleadas) · [`CAPABILITIES.md`](../product/capabilities/INDEX.md) (SSoT funcional) ·
[`LEDGER.md`](../product/LEDGER.md) índice → `ledger/HS-NN.md` (fichas atómicas); `CLAUDE.md` = **router puro**,
ya no repositorio de estado/backlog/stack (ése vive en `STACK.md`). **§10** reconciliado (el paquete
activo lo nombra `checkpoint.md`, el backlog global vive en `BACKLOG.md`; el INDEX del paquete sigue siendo
continuidad DEL paquete) + **nueva regla dura 7** «ningún cambio de código sin capability»
(`docs/architecture/boundaries/codigo-traza-a-capability.md`, **enforced**).
