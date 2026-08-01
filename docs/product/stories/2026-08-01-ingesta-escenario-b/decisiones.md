# Decisiones — Ingesta escenario B (2026-08-01)

> Conversadas y **FIRMADAS 🧑‍⚖️ 2026-08-01** vía aprobación conversacional del operador
> («Si, arma el paquete completo, haz todo») tras el análisis con evidencia
> ([`informe-estado-real.md`](./informe-estado-real.md)). Transcripción fiel; NO relitigar la
> dirección — los PENDIENTES de abajo sí son diseño abierto.

## ING-D1 · Escenario B = puerta principal (90 % del uso)

Dos puertas: **A)** proyecto con arnés formal instalado (observar/mejorar — lo de hoy) y
**B)** proyecto crudo que hace las cosas bien → procesar, ordenar, empaquetar, publicar a
marketplace propio, **reinstalar formal**, telemetría. B domina el uso real. La fábrica deja de
asumir que el arnés ya existe.

**Por qué:** el operador entra el 90 % del tiempo a proyectos vivos con know-how disperso; la
única forja real que existió (developer-vitalia, carril D) fue MANUAL porque no había camino de
producto («la materialización fue mía porque el producto me negó las manos del conductor»).

## ING-D2 · Mapa = editor manual first-class

Crear fases/artefactos/skills/hooks, cablear, reapuntar — **con clicks sobre el Mapa, y eso
escribe los archivos reales**. Lo conversacional AYUDA pero no reemplaza: el Mapa existe para
que el operador cablee a mano. Remarcado explícito por el operador («debo poder crear y
reapuntar manualmente en el mapa»).

**Por qué:** hoy el Mapa es 100 % solo-lectura (cero rutas de escritura del grafo,
`router.go:86-97`; «Editar fuente» disabled hardcodeado, `inspector.tsx:744`) — contradice la
visión firmada («crear y editar son acciones sobre el mapa», `vision.md:186`).

## ING-D3 · Inventario TOTAL del as-code

La ingesta inventaría todo: skills · rules · commands · hooks · **agents** (reconocedor hoy
inexistente, `loader.go:20-24`) · mcp · settings · **el contenedor plugin como nodo** — y
estampa **procedencia**: provisto-por-plugin / creado-por-usuario / suelto / **drifteado**
(copiado del plugin y editado local) / de-referencia (lockfiles de terceros tipo
`skills-lock.json`). Lo no-entendido sigue saliendo `no-reconocido` VISIBLE, jamás omitido.

**Por qué:** vitalia-app real: 11 agents invisibles para el loader actual; 4 skills del plugin
con drift local; 12 skills Clerk de terceros en `.agents/`. Sin procedencia no se puede decidir
qué se empaqueta como propio.

## ING-D4 · Clasificación por facetas + leer el grafo real

La ingesta clasifica cada pieza: **knowledge** (saber puro) · **knowledge-as-code** ·
**docs-as-code** · **process-as-code** · **WIP-home** (dónde viven las instancias de trabajo:
historias → archivado). Y **lee el grafo de punteros REAL del proyecto** (cadenas .md/.yml tipo
`CLAUDE.md → skill → checkpoint → stories/`) en vez de inventarlo — vitalia-app ya lo
materializa solo (`docs/process/DOCS-GRAPH.md`, 920 docs, autogenerado).

**Por qué:** hoy la clasificación es CERO código (solo prosa en docs); el enum `Clase` (10
primitivas técnicas) es OTRO eje y no se toca. Alineado con la regla de las 3 caras (D18) y las
11 canónicas (D19) del terreno.

## ING-D5 · Carpeta cruda deja de ser callejón sin salida

Apuntar a un dir sin `.claude/` ni plugin ni lock hoy devuelve `[]` y el wizard admite por
escrito que no reconoce fuentes (`portafolio-wizard.tsx:308`). Pasa a ofrecer **modo ingesta**:
inventario de lo que haya (docs, scripts, estructura) — la fábrica es **agnóstica al rubro**
por constitución (principio 7): un proyecto de finanzas o consultoría sin `.claude/` también
se ingiere.

## ING-D6 · Extraer lee el proceso REAL, no confía en manifiestos

Las desalineaciones sello↔realidad se MUESTRAN, no se tapan: en vitalia-app el spine declarado
(5 estados) no es el operado (10 estados de `/pm-vitalia`), el overlay declara un spine
fantasma, y los WIP caps viven en 3 sitios con valores distintos. La ingesta contrasta lo
declarado contra lo operado (mismo espíritu que la ley anti-drift descriptiva del Portafolio).

## ING-D7 · El arnés empaqueta PROCESO, jamás instancias

WIP-home se identifica y se declara (plantillas, estados, gates, estructura de carpetas de
historia → archivado), pero las historias/instancias vivas NO viajan en el paquete.
(Refuerza D20: paquete ≠ artefacto.)

## ING-D8 · Skill orientador = cara visible del arnés

El patrón `/pm` / `/pm-vitalia` (know-how del proceso: superficies con owner→path, máquina de
estados, tabla «operador dice → acción», punteros) se reconoce como **la pieza que encamina y
sabe qué puede/no puede hacer el paquete** — Q&A/orientador. La ingesta lo detecta y lo trata
como nodo de proceso central (el más conectado del grafo), no como un skill más.

## ING-D9 · Quick-win: instalaciones llevan `arnes.yaml`

`publish` SÍ copia `arnes.yaml` al marketplace (`publisher.go:306`), pero las **instalaciones**
nunca lo reciben — Identificar solo escribe el l0 → actividades invisibles en toda instalación
(bug vivido 2026-08-01; fix manual: copiado a `~/Proyectos/vitalia-app/`, chips vivos). El fix
sistemático es un ticket chico y desacoplable (T0) — mecanismo exacto en §PENDIENTES.

## ING-D10 · Absorbe forja-ciclo-vivo

La historia ⏸ `2026-07-10-forja-ciclo-vivo` se reencuadra: **forjar de cero = subcaso de la
ingesta** (ingesta con inventario vacío). Sus piezas vivas (scaffolder, semilla, gate D19
declarado, Slice 1a construido) son motor de este paquete, no se tiran. El backlog «Fase 1 ·
Ciclo de forja» apunta acá.

---

## ING-D11..D15 · Forma del mockup — FIRMADAS 🧑‍⚖️ 2026-08-01

> Aprobación conversacional del operador («Sigo todas tus propuestas») sobre las propuestas
> P-A..P-E de [`auditoria-previa-mockup.md`](./auditoria-previa-mockup.md), que las derivó de la
> medición real del loader contra los dos árboles. **Las preguntas Q1-Q8 de esa auditoría siguen
> ABIERTAS** — están abajo en PENDIENTES como P6.

### ING-D11 · Dos gates, ingesta primero (era P-A)

El mockup se parte en dos etapas con gate 🧑‍⚖️ propio: **(1) ingesta**, **(2) Mapa editor**.

**Por qué:** la ingesta produce el material que el editor edita — diseñar el editor antes de
saber qué sale del inventario es diseñar sobre supuestos. Además la ingesta ya tiene caso de
prueba medido (vitalia-app: 103 nodos, 1 caja, 98 en Base) y el editor todavía no.

### ING-D12 · La ingesta es una TABLA DE DECISIÓN, no un wizard de 6 pasos (era P-B)

El inventario ya es automático (el loader emite los 103 nodos hoy). Lo humano es una sola
pregunta por pieza: **¿propio / de-referencia / suelto?** y **¿es paso del proceso?** Superficie =
tabla densa con multi-select, agrupada por autoría, con acciones en lote. El wizard actual se
conserva como puerta (escenario A intacto, superset estricto) y gana una tercera rama:
«no encontré arnés — ¿inventariar igual?» (ING-D5).

### ING-D13 · El primer gesto del Mapa editor = «esto es un paso del proceso» (era P-C)

Convierte un nodo de Base en caja estampando `contract.caja: true` + `fase:` en el frontmatter.
Determinista, reversible, diffeable.

**Por qué:** es el gesto de mayor palanca **medido** — la distancia entre el proyecto crudo y el
arnés publicado son 5 bloques de frontmatter + 2 archivos de raíz (auditoría §1.1). Si el mockup
dibujara un solo gesto de escritura, tiene que ser ese.

### ING-D14 · La banda Base se colapsa por AUTORÍA antes que por clase (era P-D)

Subbandas plegadas con contador, expandibles bajo demanda, buscador de la map-bar como escape.

**Por qué:** en un proyecto crudo «45 rules» no ayuda a decidir; «38 provistos por el plugin · 5
propios · 12 de terceros» sí — **lo propio es lo que se empaqueta**. Ataca el muro medido de
90-95 % de nodos en una sola banda (auditoría §1.2), que el baseline no cubre (su fixture mayor
tiene 38 nodos).

### ING-D15 · Escenario del mockup = vitalia-app real, cifras medidas (era P-E)

Nada de datos ilustrativos para el inventario: las cifras salen de `arnesia index` y son
reproducibles con un comando. Mismo criterio que hizo funcionar el gate de MA-T6.

---

## PENDIENTES (diseño abierto — NO firmados)

- **P6 · Q1-Q8 de la auditoría previa** ([`auditoria-previa-mockup.md`](./auditoria-previa-mockup.md)
  §Preguntas abiertas) — **Q1-Q2 BLOQUEAN el mockup de ingesta**: quién escribe cuando se hace
  click en el Mapa (Go / conductor / híbrido) · si el muro de 100 nodos se ataca en el Mapa o en
  la ingesta. Q3 quedó resuelta por ING-D11. Las demás lo moldean: nombre del eje de autoría
  (Q4, choca con `procedencia` L0) · vocabulario final de facetas (Q5, ver P2) · el `CLAUDE.md`
  que no pasa el reconocedor (Q6) · los 3 `pasos[].caja` que no son cajas en el arnés YA
  publicado (Q7) · si T0 sale ya y con qué mecanismo (Q8, ver P1).

- **P1 · Mecanismo de T0** (propagar `arnes.yaml` a instalaciones): ¿(a) Identificar/sellar lo
  escribe junto al l0 cuando el canónico lo tiene? ¿(b) Refrescar/Adoptar sincroniza hacia las
  instalaciones? ¿(c) el loader hace fallback al `arnes.yaml` del canónico vía Portafolio cuando
  la instalación no lo trae? — (c) no toca disco ajeno; (a) requiere canónico conocido.
  Recomendación preliminar: (c) para leer + (a) al sellar. Decidir en el diseño de T0.
- **P2 · Vocabulario final de facetas** (ING-D4): nombres exactos y si mapean 1:1 a territorios
  del terreno (D19) o son eje aparte.
- **P3 · UI de edición del Mapa**: qué gestos (click-crear en carril, drag-cablear, panel de
  alta), qué escribe cada gesto, y el contrato de seguridad (diff antes de confirmar, nace beta
  → tren, como ya anticipa el tooltip de «Editar fuente»). → etapa mockup.
- **P4 · Reconocedor de agents**: cómo se vincula un `agents/<id>.md` a la caja que lo invoca
  (el TODO histórico de `loader.go:22`).
- **P5 · Re-extracción/backflow**: proyecto ya arnesado que siguió evolucionando local — cómo se
  cosecha el drift hacia el canónico sin pisar (conecta con deuda de dogfood y con «mejoras se
  upstreamean, jamás fork silencioso»).
