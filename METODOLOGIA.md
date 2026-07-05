# ArnesIA — Metodología de arneses (cómo los creamos, mantenemos y qué debe tener cada componente)

> **Doctrina operativa VIVA** — crece con cada iteración de la fase UX (HS-03) y las que
> sigan. Norte constitucional: [`VISION.md`](./VISION.md) (11 principios + anatomía A1–A7).
> Este documento **baja la anatomía a reglas de negocio concretas que el producto va a
> enforcar**. Registro de iteraciones y detalle de UX: [`UX.md`](./UX.md). Cementado hasta la
> **iteración 11 de HS-03** (2026-07-04).
>
> **Base de evidencia = [`knowledge/`](./knowledge/INDEX.md) (árbol de conocimiento VIVO).**
> El «qué debe tener cada componente» (§2–3) **deriva** del estándar as code por elemento
> (skills, hooks, rules, subagents, commands, mcp, plugins, settings, output-styles, statusline,
> headless). Cada nodo del árbol lleva dos capas — **L1 estándar oficial+expertos** y **L2 nuestra
> adaptación** (obligada a derivar de L1) — y emite una **rúbrica de checks evaluables** (121 al
> corte fundacional 2026-07-04) que es el ruleset de conformidad de §6. El árbol se actualiza
> **cada semana** (mecanismo en [`knowledge/CADENCE.md`](./knowledge/CADENCE.md)); cuando cambia,
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

## 2. Qué debe tener cada componente — estructura obligatoria (EN CONSTRUCCIÓN)

> **Fuente de verdad = el árbol [`knowledge/elements/`](./knowledge/INDEX.md).** Un nodo por
> elemento con L1 (estándar oficial+expertos, con fuente y fecha), L2 (nuestra adaptación) y su
> checklist evaluable. Lo de abajo es el **resumen de negocio**; el detalle vivo y las fuentes
> viven en el árbol. Cuando el árbol crezca (cadencia semanal), este resumen se re-alinea.

Reglas por tipo de componente. Se cementan a medida que las acordamos; hoy firme lo de
skills-caja y el bloque de contrato. Estándar completo por elemento en el árbol: skills-caja/
apoyo ([`skills`](./knowledge/elements/skills.md)) · Guardia ([`hooks`](./knowledge/elements/hooks.md),
[`settings-permissions`](./knowledge/elements/settings-permissions.md)) · Base
([`rules`](./knowledge/elements/rules.md)) · maquinaria ([`subagents`](./knowledge/elements/subagents.md)) ·
[`commands`](./knowledge/elements/commands.md) · terceros ([`mcp`](./knowledge/elements/mcp.md)) ·
distribución ([`plugins`](./knowledge/elements/plugins.md)) ·
[`output-styles`](./knowledge/elements/output-styles.md) ·
[`statusline`](./knowledge/elements/statusline.md) · motor conductor ([`headless-sdk`](./knowledge/elements/headless-sdk.md)).

**Skill que es CAJA de proceso:**
- Frontmatter: `name`, `description`, `version`, `model`, `clase` (contrato L0 `meta.clase`,
  I-75) y el bloque **`contract:`** (§3).
- Prosa mínima: `## Cuándo` (precondición = su input), `## Pasos`, `## Guardarraíles`.

**Skill de APOYO** (librería experta · utilidad · tercero · meta-harness):
- Frontmatter con `contract.caja: false` + `rol:` (`libreria-experta | utilidad | tercero |
  meta-harness`). No es una caja: no declara transición de estado ni gate.

**Agentes · hooks · rules · knowledge:** reglas de estructura **pendientes de acordar** en
próximas iteraciones (los agentes ya tienen contrato implícito anti-telephone
«`<veredicto> → <path>`»; los hooks su evento; las rules su ámbito always-on/condicional).

## 3. El contrato de caja — schema formal (acordado iteración 10)

Bloque `contract:` en el frontmatter del SKILL.md. **Inferible de la prosa hoy, formal
mañana.** Es lo que vuelve reales el eval-gate por-skill, la detección de precondición
(saltos de fase / guía sin bloqueo) y la validación de composición (huérfanos/mismatch).

```yaml
contract:
  caja: true                       # true si es skill-frente de una fase
  fase: <id-de-fase>
  estado: "<estado_entra> -> <estado_sale>"   # la transición que posee (spine de estados)
  necesita:                        # inputs — qué necesita para operar bien
    - art: "<artefacto o precondición>"
      de: "usuario | base:<id> | caja:<skill> | libreria:<skill> | maquinaria:<agente>"
      requerido: true
  entrega:                         # outputs — qué produce como artefacto as code
    - art: "<artefacto>"
  ruta:                            # a quién entrega, CONDICIONAL
    - a: "<skill | rol>"
      si: "<condición>"            # omitir para el happy path
  gate:                            # cómo se evalúa su SALIDA (A4) — HONESTO
    tipo: "auto | manual | parcial | none"
    detalle: "<cómo; si no hay eval, por qué falta>"
```

**Regla de honestidad del gate:** `tipo: none` cuando la skill NO tiene eval real. Nunca se
fabrica un eval para rellenar — el objetivo es que los huecos se vean (es el hueco del
principio 10 hecho dato).

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
- **REAL vs DEMO separados:** los arneses reales llevan chip REAL; los de demostración, DEMO;
  los KPIs del portafolio suman **solo arneses reales**.

## 5. Insights de producto que guían el diseño (emergentes)

- **A escala (100+ nodos), el filtro es navegación de primera clase.** El mapa necesita
  filtrar por tipo / caliente-frío / banda como ciudadanos de primera, no como adorno.
  (Descubierto con el arnés real de 141 nodos.)
- **La superficie fría es un hallazgo.** Skills sin uso, marcas dormidas, terceros,
  deprecadas — hay que distinguirlas y el arnés debe cazar el bloat (poda / lazy-load).
- **Dos vistas distintas conviven:** el **mapa de COMPONENTES** (piezas del arnés) vs el
  **tablero de FLUJO del trabajo** (spine de 10 estados, WIP caps, historias moviéndose). La
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

## 7. El estándar as code es un árbol vivo (no un doc congelado)

- **ArnesIA fija un paradigma propio (principio 8), pero anclado a las mejores prácticas
  vigentes de Anthropic y los expertos** — no a una opinión estática. Ese anclaje vive en
  [`knowledge/`](./knowledge/INDEX.md) como árbol versionado: L1 (evidencia oficial+experta,
  fechada y con fuente) + L2 (nuestra adaptación, que **deriva** de L1) + checks evaluables.
- **Se investiga y actualiza cada semana** (nuevos comandos tipo `/goal`, features, eventos de
  hook, deprecaciones). El árbol **crece y adiciona**, no se reescribe (lo viejo se marca
  `deprecado`). Mecanismo en [`knowledge/CADENCE.md`](./knowledge/CADENCE.md).
- **Excepción a «firmado = congelado»:** cuando HS-03 se firme, la UX y esta metodología se
  congelan; **el árbol de conocimiento NO** — por diseño sigue evolucionando, porque el estándar
  que ArnesIA enforca tiene que seguir al ecosistema. Un arnés que ayer cumplía puede necesitar
  mejora hoy porque salió algo nuevo: ése es exactamente el «punto de mejora» que el mapa muestra.
- **Los 121 checks son el ruleset de conformidad (§6)** hecho dato: el linter que ArnesIA correrá
  (fase 5) y la fuente de los badges de mejora que la UX empieza a pintar (it.11+).

## Estado

Documento vivo. **Cementado hasta la iteración 11 de HS-03.** Crece con los comentarios de
UX del operador (quedan muchos). Cuando HS-03 se firme, esta metodología se congela junto a
la UX (regla Rust: firmado = congelado) y pasa a ser la base de las specs de fase 4 — **salvo el
árbol [`knowledge/`](./knowledge/INDEX.md), que por diseño sigue vivo** (§7).
