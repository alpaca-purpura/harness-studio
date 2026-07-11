# Research — ¿4º nivel "Proyecto"? Definición vs Ejecución (2026-07-10)

> Insumo de co-diseño. Pregunta: ¿el trabajo-en-vuelo (planificación + paquetes/épicas/historias +
> artefactos llenados = "foto viva") merece un nivel propio, separado de Proceso(método) y Producto(durable)?
> 2 investigaciones: taxonomía PM (web) + Vitalia (local, `/home/chalreme/Proyectos/luana-vitalia/vitalia`).

## A · Taxonomía PM + veredicto (web)

**Jerarquía work-items:** Theme/pillar → Initiative → **Epic** → Feature → **Story** → Task/Subtask → Spike.
Soporte: backlog · roadmap · PRD · sprint/iteration planning. Regla: todo trazable a estrategia.

**Crítica de "épica":** cajón de sastre · sin criterio de cierre (vive por trimestres) · sesga a OUTPUT
("Build billing portal") vs OUTCOME ("Enable self-serve subscription"). Alternativas a robar:
- **Opportunity Solution Tree** (Teresa Torres): Outcome → Opportunity → Solution → Test. Separa el
  FUTURO DESEADO (outcome/oportunidad, medible) del TRABAJO (solución/experimento).
- **Shape Up** (Basecamp): **appetite** (tiempo fijo) · **pitch** (problem/appetite/solution/rabbit-holes/
  no-gos) · **bet** 6 semanas · sin backlog acumulado · cool-down. Robar: bet-con-appetite-y-no-gos.

**Veredicto sobre el 4º nivel — SÍ, con respaldo fuerte (type vs instance):**
- **Essence/SEMAT:** 3 alphas ORTOGONALES — `Way-of-Working`(método) · `Software System`(producto) ·
  **`Work`(esfuerzo en vuelo, con máquina de estados** Initiated→Prepared→Started→Under-Control→
  Concluded→Closed). El trabajo concreto NO es ni el método ni el producto.
- **Value-stream / Flow (Kersten):** el flow/work-item es la INSTANCIA que fluye (WIP, flow-time),
  medible solo si es entidad de 1ª clase, distinta de la plantilla que la generó.
- **Shape Up:** separa shaping(definir) de building(ejecutar).
**En contra / avisos:** (a) riesgo de **4º cajón-de-sastre** si no tiene criterio de cierre; (b) Essence
trata Work como **alpha con ESTADOS, no como carpeta anidada** — modelarlo como instancia-con-estados, no
jerarquía de directorios; (c) Shape Up alerta contra backlogs acumulados; (d) sobre-estructura para equipos
chicos. **Mitigación:** el paquete debe portar appetite + Definition-of-Done + máquina de estados.
**Agnóstico:** universal — legal=**matter**(usa precedents/templates→deliverables) · consultoría=**engagement**
(scope/budget/%complete/WIP) · finanzas=**deal/análisis**. Todos separan método-de-la-firma / caso-concreto /
activo-entregado. Fuentes: atlassian epics-stories-themes · producttalk.org/opportunity-solution-trees ·
basecamp.com/shapeup · cacm.acm.org (Essence alphas) · agility-at-scale (flow metrics) · netsuite PS.

## B · Vitalia (local) — lo prueba en la práctica + muestra el modo de falla

Ruta real: `/home/chalreme/Proyectos/luana-vitalia/vitalia` (el `~/Producto/...` no existe).

**Lo que YA hace bien (robar):**
- **Método EXTERNALIZADO del brand:** protocolos (`capability-protocol`, `release-protocol`, `chris-input-
  protocol`) + templates + skills viven en el monorepo raíz `docs/process/` + `.claude/`. El brand solo
  tiene TRABAJO + overlay. → separación definición(kit) vs instancia(proyecto) real.
- **Story = unidad de trabajo** con pipeline numerado `00-research→01-spec→02-design→03-arch→04-validators→
  05-guidelines→06-tickets→07-merge`. **Máquina de 10 estados** (idea→refining→…→done + parked/dropped) +
  **WIP caps por estado** (no sprints).
- **Capability = durable SSoT, story = efímera.** La story se ratifica en capability al merge
  (`created_in_story`, `change_log[].story_id`, `merge_sha`). → **el cierre = ratificar capability**.
- **Foto viva AUTO-GENERADA:** `BACKLOG.md` "DO NOT EDIT" desde artefactos-fuente + `cap-doctor` gates +
  índice bidireccional code↔cap. → anti-drift real (= nuestra doctrina cifras-generadas).
- **template↔instancia** fuerte: `_template.yaml` → cap instanciado; chris-input separa cocina/conversación
  de outputs ratificados.

**Modo de FALLA (evitar) — "el concepto de épica a mejorar":**
- **4 agrupadores compitiendo:** `outcomes/` (referenciado en README/checkpoint pero **el directorio NO
  EXISTE** — colgado) · `releases/F0-F8` (lo supersede) · `Fase1/Fase2` (congelado en los SLUGS, ortogonal
  al release: una story "fase2" cae en F7) · `areas/`+`modules/`. Migración a medias.
- **ID de story codifica un agrupador (fase) distinto de su campo `release`** → agrupador disfrazado de identidad.
- **Estado-vivo mezclado con historia:** `checkpoint.md` de story = log-dump en YAML (campos como
  `next_action` son párrafos-bitácora). El brand-checkpoint (389 líneas) arrastra config histórica congelada.
- **Lifecycle pesado/desigual:** 2 stories con 29-48 archivos, el resto con 2. Sub-slices (`OLA-1/OLA-2`)
  improvisados como prosa, no como work-units hijos con estado.

**Mejoras que propone (foco arnés):** (1) UNA sola abstracción de agrupación con schema + directorio que
EXISTA; sacar `fase` de los slugs; reconstruir `outcomes/`. (2) Partir el story-checkpoint en header-fino-
de-estado-vivo vs bitácora-append-only. (3) Sub-slices = work-units hijos con estado propio.
**Lección transferible:** separar **definición-de-proceso (plantilla/protocolo → kit)** de **artefacto-
llenado (instancia story/cap → producto)**, y mantener el snapshot generado distinto de la historia append-only.

## C · Síntesis para el modelo (mi lectura)

Los dos convergen: el 4º nivel es correcto PERO **no es un "4º área" del terreno — es un PLANO de EJECUCIÓN
(instancias con estados)**, peer al plano de definición. Esto evita el cajón-de-sastre (Essence: Work=alpha-
con-estados, no contenedor) y lo respalda Vitalia (método en kit, trabajo con estados en el brand).
- **DEFINICIÓN** (`docs/terreno/`, el arnés): dimensiones + método + plantillas + conocimiento. Estático, forjado.
- **EJECUCIÓN** (`docs/proyecto/`, el trabajo vivo): paquetes (outcome→historia→sub) con ESTADOS +
  planificación + artefactos-del-proceso LLENADOS + foto auto-generada. = el "Proyecto layer" ya existente.
- **PRODUCTO DURABLE** (código + `capabilities/` SSoT): lo construido. **La capability = el puente y el
  CIERRE** (un paquete cierra cuando ratifica su capability → mata el "cajón sin cierre").
- **Fix épica:** jerarquía limpia por OUTCOME (Opportunity Solution Tree) + appetite (Shape Up); NO codificar
  el agrupador en el slug; release/fase = tag ORTOGONAL; snapshot-vivo (auto-gen) ≠ bitácora (LEDGER);
  pipeline que ESCALA al tamaño del paquete.
