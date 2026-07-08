# ArnesIA — UX del producto (fase 2 del gran plan, ficha HS-03)

> **FIRMADA** (HS-03, it.13, 2026-07-05). Por **excepción declarada** (ver METODOLOGIA §7 y ficha
> HS-03 del LEDGER), la firma cierra la **fase** pero este doc queda **VIVO**: sigue creciendo con
> nuevas iteraciones (no se congela). Norte: [`VISION.md`](./VISION.md) v3. Mockup v1 del Mapa =
> herencia validada en esencia (HS-02).
>
> **Sync 2026-07-07 (HS-10, edición mecánica — doc vivo):** S9 pasa de «11 elementos/nodos» a
> **12** (nodo `harness-profile`, HS-07) y el inventario S2 de «7 tipos» a **10 clases** (enum
> canónico L0 de HS-08; tokens de color 6→10 resueltos en HS-09 Fase D). Sin cambio de diseño.

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
| S8 | **Corridas** (pestaña hermana) | Corridas reales JSONL: Conversación/Árbol/Waterfall + attribution paso→mapa + replay |
| S9 | **Estándar** (screen global, cross-arnés) | El árbol de conocimiento vivo (`knowledge/`): 12 elementos con L1/L2/checks (12º = `harness-profile`, HS-07) + botón «Actualizar estándar» → drawer de novedades. Chip global en el header |

**Navegación:** toolbar global = wordmark · picker de arnés (con canal+versión) · pestañas
Mapa / Diagnóstico / Tren / Historia · conmutador de capas (visible solo en Mapa).
**› Actualizado en iteración 13 (shell firmado):** la chrome pasó de toolbar superior a
**Command Rail** (izquierda); las pestañas hermanas son ítems del rail, el picker de arnés vive
en el breadcrumb y el conmutador de capas en la barra del mapa. El Portafolio gana lente
**Organigrama** (arneses por empresa/puesto, «reporta a»). Ver «Iteración 13».
**› Actualizado en iteración 14 (multisesión):** el borde izquierdo pasa de rail-de-vistas a
**rail-de-sesiones** (tabs paralelas tipo WARP, colapsable a gutter); las **vistas** bajan a tira slim
por sesión; el global (Portafolio·Estándar·Ajustes) al pie del rail. Sesión = **frente de trabajo**
(N:1 con arnés) con su conversación CC; chat invocado colapsable; sesiones persisten. Ver «Iteración 14».
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

> **Nota de numeración:** el registro guarda las iteraciones donde se **fijó un acuerdo** (por eso
> hay saltos: it.1 = grill/bifurcaciones sin numerar; it.4/it.5 se disolvieron en las subversiones
> v3.1–v3.2 de it.3 sin acuerdo propio que cementar). No es un hueco: falta mucho trabajo de UX y el
> doc sigue vivo (crece con nuevas iteraciones).

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

## Iteración 13 — Shell firmado: Command Rail (A) + Organigrama de puestos (2026-07-05)

Grill del operador: repensar el **SHELL** (cómo se navega entre vistas) para app de escritorio
con experiencia conversacional — referencia conductor.build «pero donde el fuerte sea lo
visual». Se exploraron **4 paradigmas** en un "Shell Lab" (mismo contenido, conmutable al
tacto): **A Command Rail · B Copilot Home (Claude Desktop) · C Split Studio · D Canvas-native**.
Mockup lab: `mockups/arnesia-shell-lab.html` (artifact `9bcef775-…`, favicon 🧭).

**FIRMADO — shell = A · Command Rail.** Rail de iconos fijo a la izquierda (nav espacial, tipo
IDE/Linear/conductor); el visual (mapa/portafolio) a pantalla casi completa; el chat se
**invoca** (⌘K / «Conversar») como dock derecho contextual — herramienta, no hogar. Descartados:
B (entierra lo visual = contra el fuerte), C Split permanente; **D Canvas-native guardado como
modo/milestone futuro** (el zoom-semántico de galaxia = demo de venta para la app Tauri). Esto
**evoluciona la bifurcación 4** (shell de lentes): capas + vistas hermanas siguen, pero la chrome
es el Command Rail — las «pestañas hermanas» son ítems del rail, el picker de arnés vive en el
breadcrumb, el conmutador de capas en la barra del mapa. **Muere el backlog #1** («dock a 1280px
comía el mapa»): el dock es lateral e intencional, no invade el lienzo.

**Portafolio con 2 lentes (reemplaza el grid plano):**
- **Organigrama (nuevo, hero).** Por empresa, un tablero **libre 2D arrastrable**: nodo = arnés,
  **línea = «reporta a»**, color = salud, badge = hallazgos. Materializa VISION «arneses por rol
  × proceso de compañía»: una empresa cliente = un arnés por puesto.
- **Cuadrícula.** Triage plano peor-primero; cada tarjeta muestra empresa·puesto·reporta-a·
  marketplace.

**Aclaraciones de alcance del operador (cementadas):**
1. **App = fábrica de arneses (crear + mantener), NO cockpit de empresa.** Los puestos SIN arnés
   no aparecen aquí (no contaminar el objetivo); lo demás de la empresa vive en un **cockpit**
   aparte (futuro). Cada arnés lleva metadato **puesto · empresa · reporta-a** (visible en la
   tarjeta y en la barra del mapa al entrar).
2. **No programar para la excepción.** La multi-marca del operador (vitalia/nicolify/comunify) =
   mismo código, una excepción; el arnés real es `luana-platform`. Diseñar para la realidad
   general (empresa = arnés por puesto). En el mockup: `Nordia` = empresa **EJEMPLO** del
   organigrama tipo; `Luana` = **REAL**. (Cierra el pendiente multi-marca de iteración 7.)
3. **Marketplace por empresa (y por arnés).** Dónde publica sus arneses: por defecto el mío
   (`alpacapurpura/prenter-marketplace`); una empresa puede tener **repo propio** que el operador
   crea y **enlaza** aquí → ver / actualizar / **desplegar** sus arneses ahí. Header por empresa:
   repo + tag(mío/propio) + gestionar + ⇢ desplegar. Conecta con Tren/publish (S6) y ⚙ Marketplaces.

**Mockup del shell:** `mockups/arnesia-shell-A-galaxia.html` (artifact `682f3890-…`, favicon 🌌).
Verificado headless (chromium real): 4 paradigmas del lab · organigrama 2D + edges de reporte ·
**drag real** (nodo mueve, línea sigue, sin navegación accidental) · click→mapa con metadata
strip (empresa·puesto·reporta-a·marketplace) · cuadrícula con reporta-a+marketplace · light+dark ·
consola sin errores de app.

**Dos mockups, dos roles (para fase 3):** shell + navegación + organigrama viven en
`arnesia-shell-A-galaxia.html`; el detalle profundo de superficies (4 capas del mapa, inspector
3-pestañas, corridas, tren, diagnóstico, estándar) sigue en `arnesia-mockup-v3.html`. **Al portar
a código se unifican** — el shell A hospeda las superficies de v3.

**Backlog abierto (deferido por el operador para no bloquear la fase de arquitectura):** posición
del organigrama ¿100% libre vs auto-layout + ajuste fino? ¿persistir posiciones como metadato? ·
marketplace por-arnés ¿override o hereda de la empresa? · reporta-a ¿cross-empresa o solo intra? ·
`＋ crear arnés para un puesto` desde el organigrama (dispara dock de creación).

**Cierre de fase 2 UX → siguiente: fase 3 arquitectura**, luego implementar v1 con lo definido.

## Iteración 14 — Multisesión: tabs de sesión tipo WARP en el shell A (2026-07-05)

Grill del operador: necesita **tabs de sesiones paralelas** — trabaja varios arneses a la par y al
cambiar de tab debe caer en el **contexto Y la conversación** de ese arnés (referencia: WARP, tabs a
la izquierda para saltar de conversación y continuar). Choca con el shell A firmado (it.13): el
Command Rail izquierdo era **rail de vistas**; ahora el borde izquierdo lo reclaman las **sesiones**.

**Exploración — «Session Lab», 4 paradigmas conmutables al tacto** (`mockups/arnesia-session-lab.html`,
artifact `2cc55622-…`, favicon 🗂️): **P1 Session Rail** (rail ancho, tarjetas con nombre·estado CC vivo) ·
**P2 Dual-Zone Rail** (un rail 60px: sesiones-avatares arriba, global abajo; vistas → pestañas top) ·
**P3 Conversation Cockpit** (rail sesiones + conversación persistente derecha, WARP literal) ·
**P4 Spaces + Quick-Switch** (gutter colapsado + ⌘1–9/⌘K palette). Verificado headless (chromium real):
4 paradigmas, cambio de sesión restaura vista+selección+conversación, ⌘K/palette, simulación «CC te
necesita», light+dark, consola limpia.

**La joya (por qué multisesión importa, no es solo tabs):** en trabajo paralelo el rail de sesiones es
una **superficie de aviso** — lanzas trabajo en el arnés A (headless CC corriendo), saltas a B, y A te
avisa **◐ «te necesita»** cuando CC pide tu OK/decisión. Estado de sesión de primera clase: **streaming**
(CC generando) · **await** (te necesita — pulso ámbar) · **idle** (en pausa). Contador `N ◐` en la
cabecera del rail.

**FIRMADO — shell de sesión = P1 Session Rail + colapsa a gutter.** Rail izquierdo **ancho** (~224px):
tarjeta por sesión con **nombre del frente**·arnés·empresa·salud·estado CC vivo·dónde quedó; botón **«**
colapsa el rail a un **gutter de pips** (~52px, estilo P4) para recuperar ancho de mapa, **»** lo expande.
Descartados como base: P2/P4 (avatares/pips **no distinguen** dos sesiones del mismo arnés — ver decisión
1); P3-permanente (el operador quiere el chat colapsable, no fijo).

**3 decisiones del operador (cementadas):**
1. **Sesión = frente de trabajo, N:1 con arnés.** Lo normal es 1 sesión por arnés, pero puede abrir
   **varios frentes** sobre el mismo (p. ej. `luana-platform` con «timeout de contract-guard» + «poda de
   skills frías» a la vez). ⇒ la sesión es entidad propia con **nombre de frente**; la tarjeta muestra
   frente + arnés + badge «·2 frentes»; el breadcrumb añade `· frente «…»` cuando hay N>1.
2. **Chat invocado y colapsable** (no persistente). Dock derecho que se abre con ⌘K/«Conversar» y se
   **colapsa** («⟩ colapsar») para ver mejor la información. Muere P3 como base; su conversación-siempre-
   visible queda como posible modo-foco futuro, no default.
3. **Sesiones persisten.** Al reabrir la app, las sesiones se **restauran donde quedaron** (arnés + vista
   + selección + nodo + conversación CC). Cada sesión = un **hash-state con nombre + su conversación CC
   viva** — reusa el estado serializado de OBS-13 (contexto inter-vista) ya diseñado. Chip «⭯ persisten»
   en el rail; toast de restauración al abrir.

**Cómo evoluciona el shell A (supersede la chrome de it.13):** el Command Rail **muta** de «rail de
vistas» a **rail de sesiones**; las **vistas** (Mapa/Diag/Corridas/Tren/Historia) se demotan a **tira slim
por sesión** (a la derecha del rail); el **global** (⌂ Portafolio · ⟳ Estándar · ⚙ Ajustes) baja al pie
del rail. Jerarquía mental nueva: **1º qué sesión/frente** (rail) → **2º qué vista** (tira) → **3º
conversar** (dock ⌘K). El Portafolio/Organigrama (it.13) sigue siendo el **launcher** de sesiones
(«Nueva sesión» elige arnés y nombra el frente).

**Mockup de la dirección firmada:** `mockups/arnesia-shell-A-sessions.html` (artifact `b5e559d0-…`,
favicon 🪟). Verificado headless: modelo N:1 (2 frentes de luana visibles), cambio de sesión restaura
todo, rail colapsa a gutter y switch sigue funcionando, dock invocado colapsable + ⌘K reabre, cerrar
sesión, simulación «te necesita», light+dark, **consola limpia**. Screenshots (expandido + gutter)
autorevisados.

**Decidido por el operador (2026-07-05):** nombre del frente = **auto-derivado del 1er mensaje, editable**
(✎ en la tarjeta → input inline; en el mockup) · **sin límite de sesiones ni advertencia de costo** (el
operador asume las N ventanas de contexto CC; no gatear).

**Backlog abierto (it.14):** ⌘K quick-switch de sesión (palette de P4) como **añadido** al rail, no
reemplazo · ¿dónde persiste la lista de sesiones — índice SQLite desechable vs sidecar propio (los JSONL
siguen siendo fuente de verdad)? · al portar: unificar shell-A-galaxia (it.13) + shell-A-sessions (it.14)
+ detalle v3 en un solo shell.
**Doc VIVO:** it.14 refina el shell **durante** la fase 3 de arquitectura (no reabre la firma de fase 2);
alimenta las specs de fase 4.

## Iteración 12 — S9 Estándar en el mapa: chip global + vista + botón Actualizar (2026-07-04)

> **Actualizado it.13 (2026-07-05):** el drawer se des-andamió — se retiraron el estado vacío
> «◌ Sin novedades reales» y la sección **DEMO etiquetada** («así se verá cuando SÍ haya…»);
> ahora presenta directo las novedades del barrido semanal. Sobreviven la procedencia
> (fuente+fecha por novedad), el tier híbrido (auto vs requiere-OK) y los toasts honestos.
> Razón: en esta etapa todo es mockup de diseño; no se diseña para la excepción real-vs-demo.
> El **principio de producto** (nada se inventa: un barrido real vacío muestra «sin novedades»,
> jamás fabrica) sigue vigente en METODOLOGIA §4 — se retiró el andamiaje, no la honestidad.

Pedido del operador: un **botón en el mockup** para disparar el update del estándar y, justo
después, ver un **resumen de novedades (si hay) + cómo aplicarlas con ejemplos**; y **reformular
la vista contenedor** para que el look-and-feel aloje esa opción y una vista sobre los detalles.

**Shell reformulado:** el estándar es cross-arnés, así que se volvió un **screen global**
(`S.scr:"estandar"`, hermano de portafolio, NO una pestaña del arnés). El header ganó un **chip
global `⟳ Estándar v1.0 · rev 04-jul ●0`** (versión + fecha de revisión + dot de propuestas
pendientes), visible en TODAS las pantallas — incluido dentro de un arnés. `.gear` dejó de ser el
ancla de margen; el chip lo es (mismo patrón que ya usaba gear).

**S9 · Estándar (vista + detalle):** dos paneles. Izquierda: título + explicación + botón
**«⟳ Actualizar estándar»** + métricas (v1.0 · 11 elementos · 121 checks · revisado 04-jul) +
link al buzón; grilla de **11 nodos** (borde por color de tipo, versión, #checks, 🌱 vivo).
Derecha: **rail de detalle** del nodo elegido — **L1** (estándar oficial+expertos), **L2**
(nuestra adaptación) y **tabla de checks** con severidad (error/warn/info). Es la «vista sobre los
detalles» pedida. Deep-link `#/estandar/<nodo>`.

**Botón → drawer de novedades (la joya):** máquina de estados honesta. ① progreso (4 frentes de
fuente ✓, animado; instantáneo con reduced-motion). ② resultado **real: «◌ Sin novedades reales»**
(gris ≠ verde — el árbol nació hoy, 0 delta) — **no se inventan novedades**. ③ sección **DEMO
etiquetada** «así se verá cuando SÍ haya novedades» con 3 tarjetas ejemplo: **evidencia L1
(auto-aplicado, verde)** · **check nuevo (requiere tu OK, ámbar) con arneses afectados** (luana →
3 skills clerk) · **elemento nuevo (Channels)**. Cada una con **«cómo aplicarlo» expandible = diff/
ejemplo concreto** y acciones (Aplicar/Descartar/Ver en el nodo, con toasts honestos). El tier
híbrido queda explícito: evidencia se auto-aplica, lo que toca el mapa espera OK.

**Honestidad (coherente con METODOLOGIA §4):** REAL vs DEMO separados y etiquetados · gris para
«sin novedades» · procedencia (fuente+fecha) en cada novedad · cero botones muertos (toasts).

**Verificado headless (Playwright, chromium real):** 39/39 asserts en verde — shell/chip,
navegación a Estándar, 11 nodos, rail L1/L2/checks, drawer (sin-novedades + 3 DEMO + tier +
afectados), «ver en el nodo», deep-link `#/estandar/hooks`, **regresión del mapa del arnés (20
nodos intactos)**, animación real <5s, **consola limpia**. Screenshots light+dark de vista y
drawer autorevisados. Artifact republicado al mismo URL (favicon 🏭).

**Pendiente que habilita (no hecho aquí):** conectar los checks a Diagnóstico real por arnés
(«ver arneses afectados» hoy es demo) · activar el routine semanal cloud (`/schedule`) que
alimenta el buzón · badge de conformidad por-nodo sobre el mapa del arnés.

## Iteración 11 — metodología as code: árbol de conocimiento vivo (2026-07-04)

Pedido del operador: la metodología tiene que ser **as code** — entender cómo se debe
estructurar cada elemento (skill, hook, rule, subagente, command, mcp, plugin, settings,
output-style, statusline, headless) según las mejores prácticas VIGENTES de Anthropic y los
expertos, para poder **evaluar cada elemento y el conjunto** y que el mapa muestre puntos de
mejora al instante. Requisito clave: conocimiento **vivo** — se investiga cada semana lo nuevo
(comandos como `/goal`, features, eventos), el árbol crece y adiciona, **no es un `.md` que se
lee una vez**. Cruce en dos capas: L1 estándar experto ↔ L2 nuestra forma (obligada a derivar
de L1; «nuestra forma no puede no ser la recomendada»).

**Construido (nuevo directorio `knowledge/`, no toca el mockup todavía):**
- **[`knowledge/INDEX.md`](./knowledge/INDEX.md)** — raíz del árbol (11 nodos, estados, versiones,
  reparto de checks, cross-checks transversales).
- **[`knowledge/CADENCE.md`](./knowledge/CADENCE.md)** — el mecanismo vivo: anatomía de un nodo
  (frontmatter `version/updated/status/fuentes` + L1 + L2 + checklist + changelog), el ritual
  semanal (barrido → triage → append L1 → revisión L2 → evolución de checks → bump → propagar),
  reglas del árbol (aditivo sin pérdida, fuente+fecha obligatorias, L2 deriva de L1).
- **11 nodos `knowledge/elements/*.md`** — skills, hooks, rules, subagents, commands, mcp,
  plugins, settings-permissions, output-styles, statusline, headless-sdk. Cada uno con L1
  (docs oficiales code.claude.com/platform.claude.com + estándares abiertos agentskills.io/
  AGENTS.md/MCP + expertos, todo fechado y con fuente), L2 amarrada a la fábrica de cajas, y su
  checklist evaluable. **121 checks** en total; cada check declara `severidad` + `señal en el
  mapa` (el puente a UX).

**Investigación:** 11 subagentes en paralelo (uno por elemento), fuentes oficiales priorizadas +
expertos + señales de seguridad recientes (arxiv skills, CVE-2025-59536 de hooks, scans MCP).
Novedades capturadas que un doc estático se habría perdido: merge command↔skill (v2.1.3), Tool
Search MCP default-on, `/goal` y set de built-ins nuevos, ~30 eventos de hook, rename Claude Code
SDK→Agent SDK, `--bare`, auto memory, transparencia de costo de plugins.

**Amarres de doctrina:** METODOLOGIA §2–3 ahora **derivan** del árbol (pointer + regla de no-
divergencia); nuevo METODOLOGIA §7 «el estándar as code es un árbol vivo» con la **excepción a
firmado=congelado**: al firmar HS-03 la UX/metodología se congelan pero el árbol NO — sigue
evolucionando por diseño, porque un arnés que ayer cumplía puede necesitar mejora hoy por algo
nuevo (ése es el «punto de mejora» del mapa).

**Pendiente UX que esto habilita (próxima iteración de superficie):** pintar un subconjunto de
los 121 checks como **señal de conformidad al estándar** sobre el mapa — badge por nodo
(«cumple X/Y del estándar del elemento») + entrada en Diagnóstico por check fallado + posible
5ª capa o refuerzo de Desempeño. El árbol ya da la data honesta (severidad + señal); falta la
superficie. **No se tocó el mockup en esta iteración** — es groundwork de conocimiento.

## Iteración 10 — Contrato por caja reemplaza Relaciones (v3.7, 2026-07-04)

Decisión del operador: cada caja-skill debe declarar **qué necesita para operar y qué
entrega, y a quién (condicional)**. Se reemplazó la sección **«Relaciones»** del inspector
(6 direcciones invoca/lee/escribe) por **«Contrato»**, en clave de flujo de datos:
- **necesita ↑** (inputs con origen: usuario / base / caja upstream / librería / maquinaria)
- **entrega** (outputs, artefactos as code)
- **ruta ↓** (a quién entrega, **con condición** — happy path / rework / escalate)
- **viene de** (upstream, computado de quién rutea a esta caja)

**Ruteo condicional real de luana** (inferido del SKILL.md, marcado «por formalizar»):
auditor → commit-push si APPROVED · → dev-team si CHANGES_REQUESTED (rework) · → humano si
ESCALATED; dev-team → auditor normal · → pm-luana si necesita engine lift (R23, la corrida
rl-c). pm-* rutea a po/po-ux/ux-agentico según tipo de story. Contratos poblados para las 9
skills-frente de las cajas; el resto cae a **contrato derivado del grafo** (edges) o «no
declarado». Chips navegables. Verificado headless (Contrato reemplaza Relaciones, ruteo
condicional, fallback, regresión de capas, consola limpia).

**Por qué importa (más que una vista):** el contrato por caja es lo que vuelve reales el
eval-gate por-skill (A4), la detección de precondición = saltos/guía-sin-bloqueo (principio
6), el ruteo condicional (el DAG real), y la validación de composición (huérfanos/mismatch).
Es un **forcing function**: para mostrarlo, la skill debe declararlo. Hoy en luana vive en
prosa → el mockup lo marca «inferido, por formalizar», que es a la vez honesto y la
propuesta de valor. **Siguiente paso acordado:** prompt para kit-dev que formaliza los
contratos en los SKILL.md de todos los arneses (schema `contract:` en frontmatter) para que
ArnesIA los extraiga y remapee sin inferir.

## Iteración 9 — capa Proceso = fábrica de cajas + doctrina de anatomía (v3.6, 2026-07-04)

Raíz: una conversación de modelo con el operador cerró que **un arnés es una fábrica de
cajas de proceso** (input→output), no un carril-cadena. Se firmó como doctrina y se rehízo
la capa Proceso sobre ese modelo.

**Doctrina nueva en `VISION.md` — «Anatomía del arnés» (reglas A1–A7, aditivas, no tocan los
11 principios firmados):** A1 jerarquía fase›caja›maquinaria (caja = una skill que orquesta
agentes/sub-skills y se apoya en rules/hooks/knowledge) · A2 contrato input→output as code
(la salida de una caja = entrada de la siguiente) · A3 **dos estados**: el trabajo lleva su
estado (spine), la caja es dueña de UNA transición + tiene su estado operativo (telemetría)
· A4 **el eval-gate vive en el contrato de salida** (aterriza el principio 10) · A5 la
fábrica no es recta: **rework** (aristas de retorno) + **cajas paralelas** (por marca / tipo)
· A6 la infraestructura compartida (Guardia/Base) vive en bandas, no dentro de una caja ·
A7 crear un arnés = definir sus fases y los contratos entre cajas.

**Capa Proceso rediseñada (mockup):** el strip de flujo se volvió un **pipeline de contratos**
— cada caja como tarjeta con `entra {estado}·{artefacto}` → cajas (skills) → **eval-gate
chip** → `sale {estado}·{artefacto}`, con **rework ↩** y cajas paralelas apiladas. Los
headers de carril muestran el contrato (`entra→sale` + estado del gate). El **eval-gate por
fin tiene hogar visual**: chips auto/manual/parcial/**sin-gate** (rojo punteado).

**El contraste que esto revela (la joya de la iteración):** el **demo dev-fullcycle** (arnés
ideal) tiene eval-gate **auto** en Calidad (12/14, release train) y Despliegue; **luana real
tiene «SIN GATE» en 4 de 6 cajas** — el hueco del principio 10 hecho espacial. Se *ve* que el
arnés real construye y audita pero no evalúa. Conecta 1:1 con el hallazgo crítico «sin evals»
del Diagnóstico real.

**Contratos reales de luana modelados** (del repo): Intake `idea→refining` (pedido→story,
cajas pm-* paralelas) · Spec `refining→refined` (story→01-spec.md, po/po-ux/ux-agentico) ·
Arq `refined→ready` (spec→READY PACKAGE) · Dev `ready→developed` (READY PACKAGE→código,
gate parcial gate-runner) · Cal `developed→done` (código→REVIEW.md, gate manual auditor +
rework CHANGES_REQUESTED→Dev) · Cierre `done→released`. Verificado headless (pipeline 6
cajas en ambos, gates correctos, regresión de las otras 3 capas intacta, consola limpia,
dark/light revisados). Arneses sin `pipeline` (pm-discovery/fin) caen al flujo viejo.

## Iteración 8 — arnés real COMPLETO, sin corte curado (v3.5, 2026-07-04)

Pedido del operador: «muestra los 55 skills completos, no el corte curado; de igual forma
los rules, hooks, TODO — el caso más real posible». Extraídos los bytes EXACTOS de cada
archivo del repo (`wc -c`) y generado el set de nodos por script (`carga = bytes/4` real),
fusionando la telemetría medida donde existe. **141 nodos reales**: 58 skills · 46 rules
(25 proyecto + 21 del kit symlinkeadas) · 10 agentes · 6 hooks · 10 comandos · 8
knowledge/reglas-raíz · 1 MCP. Verificado headless (jsdom + Playwright, 141 nodos render,
consola limpia, screenshots dark/light autorevisados).

**Nueva geografía — bandas transversales** (además de Guardia/6-fases/Base): el spine no
alcanza para lo que no es una fase, así que se añadieron 4 bandas: **Librería de expertos**
(13: backend/frontend/copilot/… + design systems + playwright/chrome-verify) · **Meta-harness
· Git · Utilidades · Comandos** (22) · **Marcas dormidas** (7 PMs sin bootstrap, punteados) ·
**Terceros · Clerk** (12 skills de auth externa). Base ahora lista las 46 rules
individuales (siempre-en-contexto vs condicional) + knowledge + mcp.

**El hallazgo que emerge de mostrar TODO:** superficie enorme, mayoría fría — de 58 skills
solo ~13 se activaron en 30d; **22 componentes propios fríos + 7 marcas dormidas + 12
Clerk + 2 UX deprecadas**. Es un hallazgo real de primera clase (registrado en Diagnóstico
sobre `harnesses-improvement`): candidato a poda / lazy-load. **A esta escala el filtro por
tipo de la leyenda deja de ser adorno y se vuelve el navegador principal** — insight de
producto: el mapa necesita filtros (tipo, caliente/frío, banda) como ciudadanos de primera.

**Refinamientos de honestidad forzados por la escala:**
- **Tipo `command` de primera clase** (glifo `/`, color muted) — antes salía «UNDEFINED».
- **Reglas y knowledge no se «invocan»**: reglas muestran «siempre en contexto» (alw) vs
  «carga condicional»; knowledge «leído por skills», nunca «0× 30d» engañoso.
- **Librerías preloaded no cuentan heat** (`consumoDe` devuelve 0 si `lib`): su costo «se
  cuenta en quien la usa» — evita el doble conteo de la atribución solapada.
- **Heat percentil se recalcula sobre el set real** (141 nodos): dev-team heat máx, el resto
  se distribuye; el mar de nodos fríos queda sin heat (correcto).

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
  se mezcla en los totales). *— Retirado it.13 (2026-07-05): chip REAL/DEMO, KPI «solo reales»,
  orden real-primero y statebar «⛁ ARNÉS REAL» eran andamiaje del mockup; se quitaron (todo es
  mockup de diseño; luana es solo ejemplo, sus skills aún se adaptarán). Sobrevive el resto de
  esta lista — «señales incompletas», «—» honesto, heat percentil, consumo-medido-pisa-estimado,
  procedencia.*
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

## Inventario final (firma) — corte iteración 13 (2026-07-05)

> Baseline **firmado** de HS-03 (it.13, 2026-07-05). El snapshot «corte iteración 6» de abajo se conserva como
> historia. **Dos mockups, dos roles** (se unifican al portar): SHELL/navegación/organigrama =
> `arnesia-shell-A-galaxia.html` (artifact 682f3890) · detalle profundo de superficies =
> `arnesia-mockup-v3.html` (artifact 6a63cdf3) · laboratorio de shell (4 paradigmas, A firmado) =
> `arnesia-shell-lab.html` (artifact 9bcef775).

**I · Shell & navegación (shell-A).** Command Rail izquierdo (brand + nav espacial Portafolio·
Mapa·Corridas·Diagnóstico·Tren·Historia + Estándar + Ajustes al pie) · topbar con breadcrumb
(Portafolio / empresa / picker-arnés ▾ / vista) + chip global ⟳ Estándar + botón «Conversar ⌘K» ·
**chat dock invocado** (derecho, contextual): fila de sesión CC (id·modelo·tokens·barra ctx·%) +
mensajes user/sys/asistente + diff + CTAs «Aplicar→beta»/«Ver en el mapa» + composer; Claude Code
headless detrás. Muere el backlog #1 (el dock ya no come el mapa). Light/dark · reduced-motion ·
focus · toasts honestos.

**II · S1 Portafolio, 2 lentes (shell-A).**
- **Organigrama (hero):** tablero libre 2D **arrastrable** por empresa; nodo = arnés, **línea =
  «reporta a»** (bezier SVG), color = salud, badge = hallazgos, badge «raíz»; drag real (la línea
  sigue, clic-sin-mover = entra al mapa). CTA «＋ Conectar empresa».
- **Cuadrícula:** triage peor-primero; tarjeta con empresa·puesto·**reporta-a**·**marketplace**,
  estado (sano/atención/crítico/naciendo/**señales incompletas**/inactivo), señales + hallazgos.
- **Metadato de arnés de primera clase:** puesto · empresa · reporta-a (en tarjeta, breadcrumb y
  barra del mapa). **Marketplace por empresa:** repo + tag (mío / repo propio) + gestionar +
  ⇢ desplegar. Alcance declarado: aquí SOLO se crean y mantienen arneses; los puestos sin arnés no
  aparecen (lo demás vive en el cockpit, futuro).

**III · Superficies de detalle (v3).**
- **S2 Mapa:** Guardia / carriles-fase / Base + **bandas transversales** (Librería de expertos ·
  Meta-harness·Git·Utilidades·Comandos · Marcas dormidas · Terceros·Clerk). **10 clases**
  color+forma+etiqueta (enum canónico L0 de HS-08: skill·subagent·hook·rule·command·mcp·plugin·
  settings·output-style·statusline — al corte it.13 eran 7 tipos; gap de tokens de color 6→10
  resuelto en HS-09 Fase D). **4 capas con
  contrato de datos:** Estructura · Tokens (carga+30d, fórmula única, heat percentil, Σ carril,
  budget+Pareto, solape declarado) · Desempeño (inv×·éxito%·p95/invocación·sin-uso·gate-fallado·
  grosor de edge) · **Proceso = pipeline de contratos por caja** (entra→cajas→eval-gate chip→sale,
  rework ↩, cajas paralelas, gate auto/manual/parcial/**sin-gate**). Edges direccionales. Leyenda
  = filtro por tipo (pendiente: falta el filtro de `command` — backlog). Replay de corrida sobre
  el mapa (#1..#n, camino, deep-link). Base en alerta + CTA capturar.
- **S3 Inspector (3 pestañas):** Resumen (identidad, métricas inv/éxito/p95/carga/excepciones/
  consumo, **Contrato: necesita↑·entrega·ruta↓ condicional·viene-de** [inferido/derivado/no-
  declarado], hallazgos, acciones) · Contenido (fuente + editar → diff → nota → beta → tren;
  CodeMirror 6) · Corridas.
- **S8 Corridas:** lista JSONL (disparo·fase·fecha·modelo·turnos·in/out·ctx·costo·duración·✓/✕·
  ⚠exc); detalle 3 vistas Conversación / Árbol (contabilidad de contexto por paso) / Waterfall +
  attribution paso→mapa + ▶ reproducir en el mapa.
- **S5 Diagnóstico:** bandeja unificada estático+runtime, severidad, capa sugerida, «ver en el
  mapa», badge de count. **S6 Tren:** horno / eval-gate / listo / publicado; gate-fallado →
  «Reabrir en fábrica»; promote bloqueado por gate; destinos publish. **S7 Historia:** versiones +
  deltas + hitos. **S9 Estándar (global):** 12 nodos (12º = `harness-profile`, HS-07) + rail L1/L2/checks (severidad) + botón
  Actualizar → drawer de novedades (tier híbrido auto vs requiere-OK, arneses afectados, cómo-
  aplicarlo con diff); deep-link `#/estandar/<nodo>`. **S4 Dock fábrica** (equivalente al chat
  dock): 5 modos (editar · crear-componente · crear-arnés · capturar-base · **importar arnés** →
  chequeo constitucional → adopción) + fila de sesión CC + eventos de sistema.

**IV · Global.** ⚙ Marketplaces · deep-link por hash de todo (paga OBS-13) · **honestidad de
producto intacta** (nada se inventa · procedencia · «—» sin medir · gris ≠ verde) · daltonismo
(color+forma) · toasts honestos. Estados: naciendo, sin telemetría, señales incompletas, base
incompleta, gate fallado, sin uso, propuesto, corrida fallida, bandejas vacías.

**Decisiones de arquitectura firmadas:** shell = Command Rail (A) · D Canvas-native = milestone
futuro · contrato de datos por capa · frontera Proceso↔Diagnóstico · p95 por invocación (jamás
promedios) · atribución de tokens solapada (Σ ≠ total) · fórmula de consumo única mapa=inspector ·
heat percentil · edición directa siempre pare beta→tren · Contrato por caja como forcing function ·
**andamiaje REAL/DEMO retirado del mockup, honestidad de producto conservada.**

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

~~1. Dock a 1280px deja mapa a media pantalla~~ — **resuelto en it.13**: Command Rail dejó el chat
   como dock **invocado** (⌘K), no persistente; el visual va casi-fullscreen y el dock ya no comprime el mapa.
2. Vista A/B real (hoy «Evaluar A/B» solo salta al tren).
3. ¿Flujo canónico/ideal por skill como concepto aparte del replay real?
4. Detalle de evals del gate («Ver evals») + telemetría post-deploy por proyecto.
5. Historia: mapa por versión (exige snapshot del índice — decidir en fase 3).
6. Onboarding/captura de base a fondo (hoy solo dock guionado).
7. Multi-proyecto: ¿vista por proyecto instalado?
8. Búsqueda global (componentes, corridas, hallazgos).
9. Accesibilidad teclado completa (hoy parcial) · estados vacíos restantes.
10. Taxonomía: ¿subtipos de regla / clase L0 visible en el nodo?
11. Leyenda del mapa (v3): falta el filtro para el 7º tipo `command` — se renderiza pero no se
    puede filtrar (gap detectado en it.13).
12. Organigrama (shell-A): ¿posición 100% libre vs auto-layout + ajuste fino? ¿persistir
    posiciones como metadato? · marketplace por-arnés ¿override o hereda de la empresa? ·
    reporta-a ¿cross-empresa o solo intra? · «＋ crear arnés para un puesto» desde el organigrama.
13. **Artefactos (D1–D11 firmadas 2026-07-08, LISTO PARA IMPLEMENTAR):** chips de hand-off en el
    spine + identidad del art + plantillas con llenado determinista + checks de composición —
    paquete completo con spec/design/mockup/prompt en
    `research/2026-07-07-franja-artefactos/` (arrancar con su PROMPT.md en sesión nueva).

<!-- Al cerrar cada iteración: registrar sección "Iteración N" + actualizar inventario. -->
