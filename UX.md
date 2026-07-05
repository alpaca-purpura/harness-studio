# ArnesIA — UX del producto (fase 2 del gran plan, ficha HS-03)

> **BORRADOR — en forja.** Se congela y firma al cierre de HS-03 (regla Rust: firmado =
> congelado). Norte: [`VISION.md`](./VISION.md) v3. Mockup v1 del Mapa = herencia validada
> en esencia (HS-02).

## Bifurcaciones firmadas (grill 2026-07-04)

1. **Alcance: todas las superficies parejas.** Diseño fino firmable de Mapa, fábrica
   conversacional, tren de release y diagnóstico — cero deuda UX para fases 3-5.
2. **Fábrica conversacional = dock en el lienzo.** Panel acoplado al mapa; los componentes
   nacen y mutan EN el mapa mientras se conversa (propuesto punteado → beta → estable).
   Honra "crear y editar = acciones sobre el mapa, no vistas aparte".
3. **Entrada = portafolio de arneses.** Home con la producción de la fábrica de un vistazo
   (salud + telemetría resumida por arnés) → clic entra al mapa. Picker en toolbar para
   saltar sin volver.
4. **Herencia shell de lentes: capas + vistas hermanas.** Lentes espaciales
   (Estructura/Tokens/Desempeño/Proceso) = capas sobre la geografía del mapa. Lentes no
   espaciales (Historia, Diagnóstico) = pestañas hermanas. El **contrato-de-lente
   sobrevive como principio de arquitectura de capas** (se diseña en fase 3); la UI del
   shell cero-dep muere. Deuda heredada que esta UX SÍ resuelve: contexto inter-vista
   (OBS-13: el deep-link "con ese nodo" se perdía entre lentes).

## Journeys del operador (fábrica: crear · mapear · observar · mejorar)

- **J1 · Observar la producción.** Abrir app → portafolio (salud de N arneses) → entrar a
  un arnés → Mapa capa Estructura → conmutar capas → clic en nodo → inspector (métricas,
  hallazgos, acciones).
- **J2 · Diagnosticar y corregir.** Pestaña Diagnóstico (hallazgos estáticos + runtime,
  misma bandeja) → hallazgo → "ver en el mapa" (foco nodo + capa relevante) → «Editar» →
  dock fábrica abre CON contexto del hallazgo → conversación → cambio queda en beta →
  eval A/B → promote → publish. El loop detectar→forjar→propagar (OBS-18) en una sola
  superficie.
- **J3 · Crear arnés desde cero.** Portafolio → «Nuevo arnés» → grill conversacional
  (rol × proceso de la compañía → fases del proceso = carriles del mapa) → spec → build
  headless con progreso visible (los nodos aparecen punteados en el mapa mientras nacen)
  → beta → gate → promote → publish al marketplace git del proyecto. Telemetría conectada
  de nacimiento (principio 9) — no hay paso "conectar".
- **J4 · Crear/editar componente suelto.** Acción sobre el mapa («+» en carril o banda, o
  «Editar» en inspector) → dock con contexto de fase/tipo → nodo propuesto punteado →
  mismo tren que J3.
- **J5 · Tren de release.** Pestaña Tren: cola de candidatos beta, estado del eval-gate
  por candidato, promotes históricos, publish por proyecto destino. Nada se promueve sin
  eval (principio 10).
- **J6 · Historia.** Pestaña Historia: línea de versiones del arnés, diff de componentes
  entre versiones (+/−/~), hitos ligados a fichas del ledger del arnés.
- **J7 · Base incompleta.** El arnés exige base as code antes de operar (principio 2):
  banda Base en alerta con faltantes explícitos → «Capturar base» abre el dock en modo
  captura. Guía sin bloqueo (principio 6): advierte y registra excepción, no detiene.

## Superficies

| # | Superficie | Rol |
|---|-----------|-----|
| S1 | **Portafolio** (home) | Producción de la fábrica: cards por arnés — salud, versión/canal, corridas 30d, hallazgos abiertos, proyecto(s) donde está instalado |
| S2 | **Mapa** (lienzo único) | Geografía fija: banda Guardia · carriles por fase · banda Base. 4 capas conmutables. Edges invocación/lectura/escritura |
| S3 | **Inspector** (panel derecho) | Identidad + métricas + hallazgos + acciones del nodo (Editar · Evaluar A/B · Promover · Ver historia) |
| S4 | **Dock fábrica** (panel conversacional del lienzo) | Modos: crear arnés · crear componente · editar componente · capturar base. Muta el mapa en vivo; Claude Code headless detrás (patrón conductor) |
| S5 | **Diagnóstico** (pestaña hermana) | Bandeja unificada estático+runtime; severidad; cada hallazgo → foco en mapa |
| S6 | **Tren** (pestaña hermana) | beta → eval-gate → promote → publish (release train KIT-06 operado, no duplicado) |
| S7 | **Historia** (pestaña hermana) | Versiones, diffs, hitos del arnés |

**Navegación:** toolbar global = wordmark · picker de arnés (con canal+versión) · pestañas
Mapa / Diagnóstico / Tren / Historia · conmutador de capas (visible solo en Mapa).
**Contexto inter-vista (paga la deuda OBS-13):** toda entidad (nodo, hallazgo, candidato,
versión) navega a cualquier vista conservando foco — estado serializado en URL hash
(arnés · vista · capa · selección) ⇒ deep-link compartible.

## Estados que la UX debe cubrir

- Arnés recién nacido, sin corridas: mapa completo pero métricas "—"; badge «sin
  telemetría aún» (nunca inventar datos; procedencia obligatoria, herencia OBS-08).
- Base incompleta: banda Base en alerta, faltantes nombrados (J7).
- Build headless en curso: nodos naciendo punteados + stream de progreso en dock.
- Daemon caído / reconexión: los JSONL son la fuente de verdad — aviso «índice
  desactualizado», jamás pérdida.
- Candidato beta con gate fallado: visible en Tren + pintado en mapa (capa Desempeño).

## Iteración 2 — feedback del operador (2026-07-04, sobre mockup v2.1 auditado)

Mockup v2.1: `mockups/arnesia-mockup-v2.html` (= artifact). Auditoría con click-through real:
43 asserts en verde; bug de race del ruteo (dock moría al abrir) cazado y corregido.

**Veredictos:** dock fábrica ✓ funciona · frontera capas↔pestañas ✓ (profundizar QUÉ datos
por capa) · tren ✓ inicialmente · inter-vista ✓ aparente. Expectativa: **7-10 iteraciones**
de mockup antes de congelar.

**Requerimientos nuevos (iteración 3):**

1. **Salud de la producción** — investigar el estándar actual (LLM observability + SRE
   fleet health) y proponer; no inventar de cero. (Investigación 3 frentes lanzada.)
2. **Ver y EDITAR el contenido de cada elemento** — el inspector debe abrir el fuente
   (SKILL.md, hook, knowledge) para ver y editar, no solo métricas.
3. **Flujo real por invocación** — cómo se invoca cada componente: la corrida vista desde
   la perspectiva del usuario que la usa (traza paso a paso).
4. **Dock ↔ sesión de Claude Code**: cada conversación del dock asociada a su conversación
   CC real; visible: modelo usado, tokens, contexto consumido; eventos de skills del
   sistema y comandos (p.ej. `/goal`).
5. **Importar arnés existente** — flujo para cargar un arnés ya creado por el operador
   (carpeta local / repo git). Falta en el portafolio.
6. **Conectar marketplace** — dónde y cómo se declara/conecta el marketplace git (la config
   por proyecto es debate de fase 3; el FLUJO UX se diseña aquí).

**Debilidades abiertas detectadas en la auditoría:** dock a 1280px deja el mapa a media
pantalla (¿comprimir carriles / flotar / colapsar inspector?) · capa Proceso débil, se
solapa con Diagnóstico · «Evaluar A/B» enlaza al tren pero no existe vista A/B.

## Iteración 3 — CONSTRUIDA en mockup v3 (2026-07-04)

Operador aprobó la propuesta completa («mostrando todo lo posible, apegándonos a las
mejores prácticas mapeadas»). `mockups/arnesia-mockup-v3.html` (= artifact, mismo URL).
Auditado: 63 asserts de click-through en verde + pasada visual + consola limpia.
Superficies nuevas vs v2: portafolio RED+eval con semáforo por umbral · **S8 Corridas**
(lista + detalle con Conversación/Árbol/Waterfall + attribution paso→mapa) · inspector
con pestañas Resumen/Contenido/Corridas (editor + diff-before-save + nota → beta→tren) ·
dock con fila de sesión CC (sessionId→corrida · modelo · tokens · % contexto · costo +
eventos skill_activated/hooks/comandos) · Importar arnés (chequeo constitucional → 
adopción) · modal Marketplaces (⚙).

**v3.1 (feedback iteración: «¿cómo sé qué tipo es?» · «¿quién llama a quién y en qué
orden?»):** tipo etiquetado bajo el nombre de CADA nodo + glifos con FORMA distinta por
tipo (rombo=hook · círculo=agente · hexágono=MCP · cuadrado=knowledge; color+forma, no
solo color) · leyenda = filtro interactivo por tipo · flechas direccionales en edges de
invocación/escritura · inspector: sección **Relaciones** (invoca→ / ←invocado por / lee /
←leído por / escribe→ / ←escrito por, chips navegables) · **replay de corrida sobre el
mapa**: desde el detalle de corrida «▶ Reproducir en el mapa» — badges numerados #1..#n
en los nodos en orden de ejecución + secuencia completa en banda superior + camino
resaltado; deep-linkeable (`&replay=r1`). 21 asserts nuevos en verde.

**v3.2 (feedback: flechas tapadas · ¿corridas ideales o reales? + contexto por paso ·
faltan rules):** z-index corregido — edges SOBRE los contenedores, nodos sobre edges ·
**tipo `regla` de primera clase** (6º tipo: escudo oliva, banda Base, «siempre en
contexto», métrica = sesiones que la cargan × su costo — CLAUDE.md y standards dejan de
ser invisibles) · corridas marcadas **CORRIDA REAL · JSONL** (nada es flujo ideal) ·
**contabilidad de contexto por paso** en el árbol: barra de ventana + acumulado k/% + qué
añadió cada paso (+180 hook, +13.2k SKILL+spec, pico 82k en e2e, compaction −41.3k en
verde, subagente = «ventana propia, no ocupa la principal»). 20 asserts nuevos en verde.

## La propuesta que v3 materializa

Base: [`research/2026-07-04-salud-trazas-edicion.md`](./research/2026-07-04-salud-trazas-edicion.md)
(3 frentes web + JSONL local verificado).

1. **Salud de la producción (portafolio v2).** Tarjeta IDÉNTICA por arnés (uniformidad
   RED): semáforo derivado de umbral documentado, jamás a mano — estados Sano / Atención /
   Crítico / En rollout / Naciendo / **Inactivo (gris ≠ verde)** — color+icono. 5 señales:
   corridas/día · éxito % · duración p95 · tokens/costo 30d · **eval score (5ª señal AI)**.
   Sparkline en cada número. Release en vuelo = mini barra de gates (Argo-style). Pirámide:
   card («¿está bien?») → dashboard del arnés («¿qué anda mal?») → corrida («¿por qué?»);
   todo número clickea hasta su traza. Heatmap arnés×dimensión cuando pasemos de ~15.
2. **Inspector con pestañas Resumen | Contenido | Corridas.** Contenido = fuente real
   (SKILL.md/hook/knowledge) con editor **CodeMirror 6** embebido; editar = diff-before-save
   + nota de cambio → nace beta → tren (versiones inmutables + labels movibles = nuestro
   git+semver+canales, mapeo 1:1). Diff entre versiones. Cada versión enlaza las corridas
   que la usaron. Edición directa (quirúrgica) convive con dock (conversacional).
3. **S8 · Corridas — superficie nueva** (5ª pestaña). Las 3 vistas estándar del mismo run:
   **Conversación** (perspectiva del usuario — lo pedido) · **Árbol+detalle** (debug:
   tokens/latencia por paso) · **Waterfall** (paralelismo). Attribution: click en paso →
   componente en el mapa (inter-vista otra vez). Todo sale del JSONL ya verificado
   (parentUuid, promptId, usage, tool_use↔result, subagentes, hooks, skills, comandos).
4. **Dock ↔ sesión CC.** Header: chip sessionId (→ su corrida en S8) · modelo · tokens
   acumulados · barra % contexto · costo (usage × tabla precios). Eventos de sistema como
   filas sys: skill_activated (trigger slash/proactivo/anidado), hooks, comandos (/goal),
   compaction.
5. **Importar arnés existente.** Portafolio → «Importar» (card hermana de Nuevo): carpeta
   local · repo git · marketplace conectado. Import = leer estructura → grafo → **chequeo
   constitucional** (base/telemetría/proceso) → reporte de brechas → adopción (telemetría
   conectada + tren activo). Sigue siendo «solo arneses propios»: importar = adoptar lo
   nuestro preexistente.
6. **Conectar marketplace.** Ajustes → Marketplaces → añadir remote git (URL + canal) →
   aparece como destino de publish en el Tren. (Dónde vive la config = debate fase 3; el
   flujo UX queda aquí.)

## Iteración 6 — datos exactos por capa + rediseño capa Proceso (v3.3, 2026-07-04)

Tema: backlog #1 (deuda desde iteración 2) + #2. Veredicto sobre Proceso: **rediseño, no
fusión** — la capa responde una pregunta que Diagnóstico no puede (flujo y conformidad del
proceso declarado, principio 1); el solape muere al firmar la frontera.

**Contrato de datos por capa (firmable):** cada capa = una pregunta + datos exactos +
procedencia; el botón de capa lleva su pregunta como tooltip y la leyenda muestra
umbrales/procedencia por capa.

1. **Estructura** — ¿qué hay y cómo se conecta? Tipo (color+forma+etiqueta) · canal ·
   edges direccionales. Fuente: índice estático del arnés.
2. **Tokens** — ¿en qué se van los tokens? Nodo: `carga X · 30d Y` con **fórmula única
   mapa=inspector** (reglas = sesiones×carga; resto = inv×(tokens-por-invocación‖carga)) ·
   heat por umbral documentado (▓≥20k ▓▓≥75k ▓▓▓≥200k ▓▓▓▓≥400k) · Σ carril carga+30d ·
   franja presupuesto + **Pareto top-3 clickeable** · nota honesta: **la atribución se
   solapa** (knowledge/reglas cuentan dentro de quien las carga — Σ ≠ total facturado).
   Fuente: usage JSONL.
3. **Desempeño** — ¿qué funciona, qué falla, qué sobra? Nodo: inv× · éxito% coloreado por
   umbral (≥90/85) · **p95 por invocación** (cuantiles, jamás promedios — estándar F1) ·
   sin-uso punteado · gate-fallado rojo · grosor de edge = volumen · franja computada
   «peor éxito» top-3 + sin-uso + puerta a Diagnóstico. Fuente: corridas JSONL 30d.
4. **Proceso** — ¿fluye el proceso declarado? Franja de **flujo fase→fase** con volúmenes ·
   **saltos de fase como chips warn → abren SU corrida real** · tipología de excepciones
   (salto/fallo-de-skill/orden-interno, suma = chips de carril) · nodo = excepciones que
   protagonizó (skills/agentes) o su maquinaria (hooks Guardia: inyecciones/registros/
   excepciones detectadas) · knowledge/reglas/mcp atenuados con «—» (no protagonizan
   excepciones) · cumplimiento = corridas de la fase sin excepción.

**Frontera firmada Proceso ↔ Diagnóstico:** Proceso = flujo agregado y conformidad (dónde
y cuánto se rompe el proceso); Diagnóstico = hallazgos accionables por componente (qué
arreglar). Una excepción repetida se vuelve hallazgo (lo hace registro-excepciones); el
hallazgo vive en Diagnóstico, el flujo en Proceso.

**Además:** corrida **r3** nueva (salto Specs⇢Desarrollo con guía sin bloqueo, principio
6) cierra el loop capa→salto→corrida→replay; badge «⚠ n exc» en fila y detalle de corrida ·
inspector gana métricas p95 + excepciones 30d · **corrección de honestidad**: constructor
(725k, ventana propia) supera a implementar-feature (418k) — hallazgo reformulado a «mayor
consumidor de la ventana principal» y tok30 del arnés dev ajustado 1.94M→2.6M ($38; KPI
portafolio $44) para que la suma sea defendible. Auditado: ~50 asserts nuevos en verde +
reload de deep-link + dark/light + consola limpia.

## Iteración 7 — arnés REAL como data de prueba: luana-platform/vitalia (v3.4, 2026-07-04)

Pedido del operador: cargar SU proyecto real (`~/Proyectos/luana-vitalia`) al mockup para
juzgar si el producto le sirve, viendo su estado real. Investigación con 3 subagentes en
paralelo (estructura del repo · inventario de componentes · telemetría de 194 JSONL de
`~/.claude`, 2 pasadas). Todo verificado headless con jsdom + Playwright (chromium real,
~50 asserts en verde, screenshots dark/light autorevisados, consola limpia) — **el
chrome-devtools MCP se cayó al matar el chrome zombie que tenía el lock del lane
`luana-vitalia-I` (justo el bug SingletonLock de la corrida rl-b); Playwright headless lo
reemplazó sin conflicto de lane**.

**Qué se modeló (todo REAL, procedencia explícita):** el arnés = **pipeline dev
multi-marca** de luana (worktree vitalia, rama `wip/vitalia`). Geografía: Guardia = **6
hooks reales** (auto-chain, learning-detect, overlay-check, contract-guard,
validate-session-close, telemetry-emit del kit `harness@0.5.2`) · **6 fases** = spine de
proceso (Intake·Priorización → Spec&UX → Arquitectura → Desarrollo → Calidad → Cierre&Merge)
con sus skills/agentes reales (pm-luana/pm-vitalia · po/po-ux/ux-agentico · architect +
architect-orchestrator · dev-team + context-builder/validator + builders be/fe/agentic +
gate-runner + backend/frontend-expert · auditor + auditors be/fe/agentic ·
commit-push/handoff/pase-produccion) · Base = reglas always-on (CLAUDE.md, AGENTS.md,
overlay vitalia, 46 rules) + knowledge (project.config.yaml costura, lifecycle,
capability-protocol, harness-arch, paradigm) + mcp-tessl. **39 nodos, 18 edges** (cadena
real dev-team→builders→auditor).

**Telemetría real integrada (194 sesiones, 30d):** consumo por componente = suma real de
tokens del JSONL (`c30` pisa la fórmula estimada) — dev-team 6.7M out (heat máx), pm-luana
5.3M, pm-vitalia 3.2M, builder-fe 2.5M. 3 corridas reales paso a paso (rl-a pm-vitalia
admin · rl-b SingletonLock con subagente + 2 errores · rl-c dev-team OLA-2 con
builder-agentic re-spawn por API 529). Historia real (kit v0.5.2 · charter 3-capas ·
bootstrap 27 pkgs).

**Hallazgos REALES en Diagnóstico (no inventados, trazables al repo):** 2 críticos —
`auditor-agentic` 0 lanzamientos pese a `builder-agentic` activo (código agentic construido
pero **nunca auditado**) · **ningún componente define criterio de éxito ni evals** (el gate
«nada sin eval» del principio 10 no está instrumentado → salud «señales incompletas», no
puede ponerse verde). Warns — `context-validator` 0 lanzamientos (brief sin validar) ·
`pase-produccion` DEFERRED (sin owner de deploy) · incoherencia de fases A–F vs {G,R,C,D}.
Infos — **telemetry-emit del kit YA es el sensor que ArnesIA necesita (KIT-03)** · dev-team
consumo · doc drift (glossary roto, ADR dir vacío) · mcp-tessl sin uso.

**Decisiones de honestidad forzadas por la data real (nuevas):**
- **Chips REAL vs DEMO** por arnés; **KPIs del portafolio = solo arneses reales** (demo no
  se mezcla en los totales).
- **Éxito/p95 por componente = «—» honesto**: el JSONL no trae criterio de éxito de corrida
  → no se inventa; es en sí un hallazgo. Estado del arnés = **«señales incompletas»** (5ª
  rama del semáforo) cuando hay telemetría viva pero sin criterio de calidad.
- **Heat pasa de umbral absoluto a percentil del arnés (p50/75/90/97)**: el absoluto no
  escala entre un arnés de 2.6M y uno de 81M tok/30d. (Contrato de capa Tokens actualizado.)
- **Consumo medido pisa consumo estimado** (`c30`) — la fórmula queda solo para lo no medido
  (reglas always-on). Librerías preloaded (backend/frontend-expert) = «se cuenta en quien la
  usa» (atribución solapada explícita).
- **ctx por paso y contabilidad de contexto = «—» en corridas reales**: no se extrajo del
  JSONL en esta pasada; no se fabrica (la contabilidad por paso de los demos era sintética).
- **Tren vacío** para el arnés real (versiona con git del monorepo, aún no adoptado por el
  tren de ArnesIA).

**Pendiente para próximas iteraciones (feedback del operador que sigue):** contabilidad de
contexto por paso desde JSONL real · criterio de éxito/eval (¿cómo lo definiría el
operador?) · ¿mostrar los ~55 skills completos o el corte curado? · empleados-IA del
producto (Valeria·Lisa…) vs roster de dev — ¿dos vistas? · multi-marca (vitalia/nicolify/
comunify) ¿un arnés con overlays o varios?

## Inventario de funcionalidades — corte iteración 6 (mockup v3.3, 2026-07-04)

> Fuente de verdad para retomar en cualquier sesión. Mockup vigente:
> `mockups/arnesia-mockup-v3.html` (v2 preservado como historia). Artifact único:
> https://claude.ai/code/artifact/6a63cdf3-e6ee-442e-b17c-c659995baec1
> Todo lo listado está CONSTRUIDO y verificado con click-through (≈175 asserts acumulados).

**S1 · Portafolio.** 6 KPIs (arneses · corridas 30d · tokens 30d · costo $ · hallazgos ·
candidatos tren). Tarjeta idéntica por arnés (RED+eval): semáforo por umbral documentado
con tooltip (sano ✓ / atención ▲ / crítico ✕ / naciendo ◐ / **gris ◌ sin datos ≠ verde**),
5 señales (corr/día+sparkline · éxito%+sparkline · p95 · tok·$ 30d · último eval-gate),
chips de tren en vuelo, proyectos instalados, hallazgos; orden worst-first. Cards
**Importar arnés** y **Nuevo arnés**.

**S2 · Mapa.** Geografía Guardia / carriles-fase / Base. **6 tipos** con color+forma+
etiqueta visible (skill azul ▢ · agente violeta ● · hook rosa ◆ · knowledge verde ■ ·
mcp gris ⬡ · **regla oliva escudo — «siempre en contexto»**). **4 capas con contrato de
datos firmado (iteración 6)** — cada una con su pregunta en tooltip y nota de leyenda con
umbrales/procedencia: Estructura (tipos+edges) · Tokens (nodo carga+consumo-30d fórmula
única, heat por umbral, Σ carril, budget + Pareto top-3, nota de solape de atribución) ·
Desempeño (inv×, éxito% umbral, p95/invocación, sin-uso, gate-fallado, grosor de edge,
franja peor-éxito) · Proceso (flujo fase→fase, saltos→corrida, tipología de excepciones,
nodo con excepciones/maquinaria, cumplimiento por carril). **Edges direccionales con
flecha** (inv/write; know punteado sin flecha) pintados SOBRE contenedores. Leyenda =
filtro por tipo. «+» por carril → dock. Statebars: naciendo / sin-telemetría / **replay**.
Banda Base en alerta con faltantes + CTA capturar. **Replay de corrida sobre el mapa**:
badges #1..#n en orden real, secuencia narrada, camino resaltado. Deep-link por hash de
TODO (arnés · vista · capa · sel · dock · run · vista-traza · replay).

**S3 · Inspector (3 pestañas).** *Resumen:* identidad+canal, métricas (inv 30d · éxito ·
**p95/invocación** · carga · por-invocación · **excepciones de proceso 30d** · consumo 30d
— fórmula única compartida con el mapa; reglas: sesiones-que-cargan × costo), **Relaciones**
en 6 direcciones (invoca→ / ←invocado-por / lee / ←leído-por / escribe→ / ←escrito-por,
chips navegables), hallazgos, acciones (Editar conversando · A/B→tren · Promover si beta ·
Ver historia). *Contenido:* fuente real con números de línea + chips versión/diff/historial
+ **Editar fuente → diff altas/bajas → nota de cambio → confirmar → beta → tren** (editor
real: CodeMirror 6, decidido por investigación). *Corridas:* las corridas donde actuó.

**S4 · Dock fábrica.** 5 modos: editar (genérico por nodo + 2 narrativas ricas:
auditoria-tecnica y mcp-tracker gate-fallado) · crear-componente (desde «+» del carril) ·
crear-arnés (grill → carriles ghost → spec → forjar) · capturar-base · **importar arnés**
(fuente local/git/marketplace → chequeo constitucional 5 checks → reporte de adopción).
**Fila de sesión CC**: sessionId clickeable → SU corrida en S8 · modelo · tokens · barra
% contexto · costo. Eventos de sistema como filas: skill_activated (con trigger), hooks,
comandos (/goal), compaction. Rail conmuta inspector↔dock.

**S8 · Corridas.** Lista de corridas REALES (chip «CORRIDA REAL · JSONL» — nada es flujo
ideal): disparo · fase · fecha · modelo · turnos · in/out · ctx máx · costo · duración ·
✓/✕ · **badge «⚠ n exc» si registró excepciones de proceso** (también en el detalle);
filtros todas/fallidas. 3 corridas demo en dev (r1 feliz · r2 fallo · **r3 salto de fase
con guía sin bloqueo**). Detalle con 3 vistas: **Conversación** (burbujas + filas sys)
· **Árbol** (tokens+duración por paso, subagente anidado, attribution «ver componente en
el mapa», **contabilidad de contexto por paso**: barra de ventana + acumulado k/% + qué
añadió cada paso + pico + compaction en verde con delta negativo + subagente = «ventana
propia») · **Waterfall** (barras temporales por tipo). Botón **▶ Reproducir en el mapa**.

**S5 · Diagnóstico.** Bandeja unificada estático+runtime, severidad, capa sugerida,
filtros, «ver en el mapa» → foco+capa. Badge de count en la pestaña.

**S6 · Tren.** 4 columnas (horno / eval-gate / listo / publicado); candidato con barra de
evals; gate fallado → «Reabrir en fábrica» abre el dock con la narrativa del fallo;
promote bloqueado por el gate; destinos de publish por proyecto.

**S7 · Historia.** Versiones + deltas (+nuevos ~mutados −retirados) + hitos ligados a
fichas; actual → mapa; viejas → snapshot pendiente (debate fase 3).

**Global.** Picker de arneses · pestañas con badge · ⚙ Marketplaces (lista + conectar
remote git) · toasts honestos en acciones de fase 3-4 (cero botones muertos) · light+dark
· reduced-motion · focus visible · color+forma (daltonismo). Estados cubiertos: naciendo,
sin telemetría, base incompleta, gate fallado, sin uso, propuesto, corrida fallida,
bandejas vacías.

**Decisiones acumuladas además de las 4 del grill:** 5 señales RED+eval (eval = último
gate hasta tener evals continuos) · corridas siempre reales (JSONL) · regla = 6º tipo de
primera clase · contabilidad de contexto visible por paso · CodeMirror 6 · edición directa
SIEMPRE pare beta que va al tren (principio 10, sin borradores fuera del tren) ·
**contrato de datos por capa: una pregunta + datos exactos + procedencia + umbral
documentado (iteración 6)** · **capa Proceso rediseñada, frontera con Diagnóstico firmada
(flujo/conformidad vs hallazgos accionables)** · **p95 por invocación, jamás promedios** ·
**atribución de tokens declarada como solapada (Σ ≠ total facturado)** · **fórmula de
consumo única mapa=inspector**.

## Backlog de profundización (iteraciones 7+)

~~1. Qué datos exactos por capa~~ · ~~2. Capa Proceso vs Diagnóstico~~ — **resueltos en
iteración 6** (contrato de capas + rediseño Proceso).

1. Dock a 1280px deja mapa a media pantalla — ¿comprimir carriles / flotar / colapsar?
2. Vista A/B real (hoy «Evaluar A/B» solo salta al tren).
3. ¿Flujo canónico/ideal por skill como concepto aparte del replay real?
4. Detalle de evals del gate («Ver evals») + telemetría post-deploy por proyecto.
5. Historia: mapa por versión (exige snapshot del índice — decidir en fase 3).
6. Onboarding/captura de base a fondo (hoy solo dock guionado).
7. Multi-proyecto: ¿vista por proyecto instalado?
8. Búsqueda global (componentes, corridas, hallazgos).
9. Accesibilidad teclado completa (hoy parcial) · estados vacíos restantes.
10. Taxonomía: ¿subtipos de regla / clase L0 visible en el nodo?

<!-- Al cerrar cada iteración: registrar sección "Iteración N" + actualizar inventario. -->
