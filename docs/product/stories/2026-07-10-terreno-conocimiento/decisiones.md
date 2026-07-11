# Decisiones — Terreno · Dimensiones · Conocimiento (co-diseño 2026-07-10)

> Paquete de co-diseño (realineación core ciclo-arnés). **MODELO BASE FIRMADO 🧑‍⚖️ 2026-07-10 (Chris).**
> Firma cubre: D0-D20 + `arnes.yaml` (dogfood dev) + `estructura-terreno.html` v4. Cambio posterior = nueva `D` con firma.
> Lo aún no cerrado NO bloquea: vive en «Por profundizar (diferido)» abajo.
> Insumos: [`research-terrenos.md`](./research-terrenos.md) (metodologías dev) ·
> [`research-marcos-naming.md`](./research-marcos-naming.md) (marcos + naming, Wave 1).
> Cada decisión conversada se escribe acá EN EL MISMO TURNO (disciplina §10).

## D0 · Terreno ≠ Conocimiento (distinción conceptual) — FIJADA
- **Terreno** = lo que el proyecto DEBE definir/construir; se **trabaja** (write → obra).
- **Conocimiento** = lo que se necesita para APLICAR el terreno; de las decisiones del terreno
  **nacen** las rules + skills de conocimiento (Regla de Rosetta). Se **carga** (read → contexto).
- El arnés LEE conocimiento para saber cómo TOCAR el terreno.

## D1 · Naming (triada con respaldo metodológico) — DECIDIDA
| palabra | qué es | respaldo |
|---|---|---|
| **Terreno** | el mapa/landscape COMPLETO (todas las dimensiones juntas) | Wardley/TOGAF *Architecture Landscape* |
| **Dimensión** | cada eje que el proyecto define y puebla | OLAP/dimensional-modeling |
| **Estado / view** | la instancia por-proyecto, con salud vacío/parcial/lleno | ISO 42010 *view* + Essence *alpha states* |
- Corrige el mismatch: el borrador usaba «terreno» para el eje. **Terreno = el todo · Dimensión = el eje.**
- Cada dimensión = un «alpha» (Essence): tiene estados de salud → es tu capa **Llenado**, ahora fundada.

## D2 · Macro-agrupación rubro-agnóstica — DECIDIDA
- Macro = **3 áreas de preocupación de Essence/SEMAT**: **Cliente/Valor · Solución · Emprendimiento**
  (fuerza incluir Cliente y Emprendimiento, que faltaban por completo).
- Cada dimensión lleva además su **interrogativa Zachman** `[What·How·Where·Who·When·Why]`
  (lenguaje natural → sirve a dev, contabilidad, finanzas, cualquier rubro).

## D3 · Calidad = overlay transversal (con excepciones pesadas) — DECIDIDA
- Base: partición **estructura vs calidad** triangulada por ISO 42010 (viewpoint/perspective),
  ISO 25010:2023 (9 atributos) y AOP (cross-cutting). La calidad **permea**, no es un lugar → **overlay**
  que proyecta checks (BDD/capabilities) sobre cada dimensión estructural, no carpetas peer.
- **Excepción decidida:** **Seguridad · Testing · Compliance** conservan **carpeta propia** (peso
  operativo) *y* además proyectan checks (doble naturaleza). El resto del overlay (Performance ·
  Reliability · Maintainability · Flexibility · Compatibility · Safety) queda solo como overlay.

## D4 · Set de dimensiones (dev) — PROPUESTO (base para product-map; afinable)
`[Zachman]` · `⟳`=auto-derivado · ★=descubierta en la research (faltaba)
```
CLIENTE/VALOR (Essence·Customer)
  ★Oportunidad/Valor [Why] · ★Stakeholders [Who] · ★Requisitos [What]
SOLUCIÓN (Essence·Solution)
  Dominio [What] (subdominios core/supporting/generic) · Arquitectura [How] (estilo·ADR·metodología DDD/hexag/FSD)
  Datos/Información [What] · Interacción/UX [Who/How] (Storybook=SSoT·tokens) · Tecnologías [Where⟳] (SBOM)
  ★Concurrencia [How] · ★Contexto/Alcance [Where]
EMPRENDIMIENTO (Essence·Endeavour)
  Proceso/Trabajo [When/How] (espina fases→cajas→ROL·tramos, FIJO del arnés)
  ★Equipo [Who] (RACI·permisos) · ★Forma-de-trabajo [How] · ★Economía/Valor [Why]
CALIDAD ▓overlay ISO 25010 (9)▓  Functional-Suitability·Performance·Compatibility·Interaction-Capability
  ·Reliability·Security*·Maintainability(⊃Testing*)·Flexibility(⊃scalability,Evolution)·Safety  (+Compliance*)
  * = conservan carpeta (D3).  Observabilidad ⊂ Maintainability(analyzability)+Operations.
```

## D5 · Conocimiento co-locado (no carpeta global) — DECIDIDA (research E)
- Cada dimensión estructural tiene su `knowledge/` local (rules+skills que nacen de sus decisiones).
- Traza **bidireccional** (dato↔enforcement), estilo `codigo-traza-a-capability`.
- En el mapa: conocimiento = **capa de enforcement pegada al nodo**, NO región peer.
- Rules/skills OPERATIVAS del harness (gates) siguen en `.claude/` (kit); las de CONOCIMIENTO DEL
  PROYECTO se co-locan bajo su dimensión.

## D6 · Navegación anti-token — DECIDIDA (ya es doctrina del repo)
- Index-First + hojas atómicas (Zettelkasten) + **punteros auto-derivados del código** (SBOM ·
  storybook `index.json` · manifiestos · Structurizr DSL). 3-4 saltos: `CLAUDE.md → docs/<eje>/INDEX
  → <dimensión>/INDEX → <slug>.yaml`. Extiende reorg HS-18/19 + `estado.sh --check`.

## D7 · Packaging as-code (Wave 2) — DECIDIDA en dirección
- Cada **dimensión estructural = un `SKILL.md`** (frontmatter `description` selector + cuerpo escueto +
  refs lazy) → co-locación física + Index-First (progressive disclosure 3 capas: ~500 tok arranque).
- **Activación 4-modos** (Kiro/Cursor): `always | glob(fileMatch) | demand(description) | manual` — es
  nuestro low-token exacto. El loader carga por glob del archivo tocado + tag por rol/caja.
- Publicar **`AGENTS.md`** como fachada multi-herramienta del `CLAUDE.md` (closest-file-wins anidado).
- Unidad vendible = **expansion-pack/plugin** (arnés por rubro) vía `marketplace.json` (ya lo hacemos).

## D8 · Mecanismo sello/deriva/init/doctor (Wave 2 · robado de cruft/copier/projen) — DECIDIDA
- **Sello** = `.copier-answers.yml`/`.cruft.json` ENDURECIDO: fuente + **versión canónica pineada** +
  inputs resueltos + **firma de fábrica** + **hash por-archivo** del render canónico (baseline).
- **Deriva por-archivo** = three-way merge estilo `copier update`: re-render versión-vieja con inputs
  guardados → "old-generated"; diff vs working-tree clasifica: `original`(idéntico) · `modificado-usuario`
  (difiere) · `generado`(marker de propiedad) · `no-reconocido`(sin match) · seed-once(`skip_if_exists`).
- **Doctor** = re-síntesis idempotente (projen) + limpieza-por-**marker de propiedad** (solo toca lo
  `generado`, **jamás** `no-reconocido`/código de usuario) + `update` 3-way para lo modificado.
- **Drift-gate** = projen anti-tamper + `cruft check` en CI = honestidad automática N2 (ya iniciada HS-21).
- **Propiedad PARCIAL** (no total como projen): el arnés solo posee lo que el proceso EXIGE.
- **Aporte genuino (nadie lo tiene):** el **loop-forward invierte cruft** — la fábrica **cosecha** el diff
  local, lo promueve a canónico, re-publica; las instancias corren `update` y su artificio se reemplaza.

## D9 · Schema de "hoja de conocimiento" (Wave 2 · aider/llms.txt/Kiro/Diátaxis) — DECIDIDA
```yaml
---
id: <dimension>/<slug>
tipo: reference            # Diátaxis: reference|how-to|explanation|tutorial
dimension: <dimension>
resumen: "<1 línea>"       # L1 → va al INDEX de la dimensión
activacion: glob           # always|glob|demand|manual
globs: ["<path/**>"]
punteros_auto:             # GENERADOS (tree-sitter/deps/PageRank), no tecleados; verificados en CI
  - <path>#<símbolo>
  - capability: <module>/<slug>
budget_tok: <N>
---
# cuerpo atómico (L2, ≤~400 tok). Sin historia (→LEDGER), sin cifras a mano (se generan).
```
- INDEX-first por dimensión (solo resumen+link) · punteros vivos auto-derivados (repo-map/aider) ·
  carga 3-4 saltos `CLAUDE.md→docs/<eje>/INDEX→<dimensión>/INDEX→hoja→puntero-a-código`.

## D10 · ArnesIA = FORJA (no marketplace de packs ajenos) — CORRIGE modelo previo
- NOSOTROS forjamos cada arnés con nuestra metodología. El dimension-set le PERTENECE al arnés
  (inversión de propiedad); como lo creamos determinísticamente, lo sabemos leer/mantener/cosechar.
- El marketplace distribuye NUESTROS arneses (solo arneses propios — seteamos el estándar). NO es un
  pack opinado de terceros que el usuario compra; es un arnés que forjamos para el rol.
- Reemplaza el framing "arnés-pack golden-path comprado" de turnos previos.

## D11 · Naming de áreas + regla per-arnés — DECIDIDA (triada CONFIRMADA)
- Áreas canónicas (motor): Cliente→**Demanda** · Solución→**Producción** · Emprendimiento→**Organización**. ✅ confirmada.
- **Regla dura (E):** el nombre canónico es interno (el motor traza por él); **cada arnés DEBE
  renombrar** áreas y dimensiones al lenguaje de su usuario. Modelo: `{id_canónico, label_del_arnés}`.
  dev: Producción→**Producto**; finance: Demanda→Mandato · Producción→Informe/Modelo · Organización→Metodología.

## D12 · Forjador determinista + gates + dogfood — DECIDIDA (respuestas 1,2,4,5)
- **Canónicas obligatorias (1):** el motor exige un mínimo por área → **gate de completitud** en la forja.
- **Calidad por-rubro (2):** overlay = concepto universal; **el catálogo de atributos lo trae cada arnés**
  (software=ISO 25010; finanzas=exactitud/auditabilidad/cumplimiento/…).
- **Determinismo total (4):** el forjador NO improvisa — rellena PLANTILLAS (nunca de cero); todo as-code.
  **Gates de forja** validan: canónicas cubiertas · calidad mapeada · naming del rubro provisto (E) ·
  cada slot con receta · cada rule/skill trazable a decisión. Extiende `arnesia conformance` + PARIDAD.
- **Forjador = motor + dogfood (5):** forja usando NUESTRA metodología (fases→cajas→gates→PARIDAD).
  ArnesIA se forja con ArnesIA.

## D13 · Dimensiones CANÓNICAS + gate de completitud — DECIDIDA · ⚠ set consolidado/actualizado en **D19**
**10 canónicas obligatorias + declaración de calidad** (el mínimo que el motor exige en todo arnés):
| área | id canónico | Zachman | universal |
|---|---|---|---|
| Demanda | `encargo` | Why | por qué existe el trabajo / oportunidad |
| Demanda | `stakeholders` | Who | para quién · afectados |
| Demanda | `requisitos` | What | qué debe lograr/responder |
| Producción | `producto` | How/What | estructura del entregable |
| Producción | `datos` | What | insumos / materia prima de datos |
| Producción | `contexto` | Where | perímetro / alcance |
| Organización | `proceso` | When/How | espina fases→cajas→ROL (FIJO) |
| Organización | `equipo` | Who | quién ejecuta · RACI |
| Organización | `forma-trabajo` | How | convenciones · prácticas · disciplina |
| Organización | `economia` | Why | costo vs valor |
| Calidad | `calidad` (overlay) | — | declarar catálogo ≥1 atributo; contenido por-rubro |
- `economia` y `contexto` = **obligatorias** (decidido: máximo rigor).
- **`producto`: el rubro ELIGE** — expandir en sub-nodos navegables (dev: producto→arquitectura→FE/BE)
  o agregar dimensiones hermanas en Producción. El gate solo exige las canónicas presentes; lo demás libre.
- Cada canónica declara `{id, label_rubro (E, obligatorio), área, zachman, receta}`.
- **Gate de completitud (la forja FALLA si):** falta área o canónica · canónica sin `label_rubro` ·
  sin `receta` (plantilla+scraping+preguntas) · sin catálogo de calidad · *(rec)* decisión sin rule/skill
  trazable. Extiende `arnesia conformance` + gates PARIDAD. Todo determinista/verificable.

## D14 · Áreas — naming FINAL (4 territorios) — DECIDIDA
- **Propósito** (antes Cliente/Demanda — "Demanda" no se entendía: es por-qué/para-quién/qué-se-pide).
- **Producto** (antes Solución/Producción — "Producción"=maquinaria, mal; el área es lo producido).
- **Organización** (antes Emprendimiento).
- **+ 4º territorio: WIP** (el trabajo vivo). Rename per-arnés sigue obligatorio (E).
- Los cuatro: **Propósito · Producto · Organización · WIP** + **Calidad** (overlay).

## D15 · WIP = 4º territorio (trabajo vivo · naturaleza INSTANCIA) — DECIDIDA
- Es un área **peer** a las otras 3, PERO de naturaleza **instancia** (no definición): paquetes de trabajo
  con **ESTADOS** + planificación + artefactos-del-proceso **LLENADOS** + foto auto-generada.
  (Distinto de `proceso`, que guarda el MÉTODO + las PLANTILLAS.)
- **Regla anti-cajón (diseño del user):** toda carpeta de trabajo lleva **estado** + **criterio de cierre**
  (ratifica capability) + al terminar pasa a **`done/`** → orden permanente, cierre real.
- Respaldo: Essence `Work`-alpha (con estados) + value-stream flow-item + Vitalia (método externalizado,
  story 10-estados + WIP-caps). Corrige mi propuesta de "plano": el user lo quiere como 4º ÁREA; el
  done/+estados cubren el riesgo de cajón-de-sastre que yo señalé.
- Tipos de planificación **templetizados por-rubro** (dev: outcome→historia→sub · finanzas: engagement→análisis).

## D16 · Capability = cierre + constitución del Producto — DECIDIDA
- **capability = detalle de lo implementado** (unidad). **Producto = constitución de muchas capabilities,
  FUNCIONALES Y NO-FUNCIONALES.**
- El **overlay de Calidad produce las capabilities NO-funcionales** → integra Calidad al Producto (deja de flotar).
- **Frontera WIP/Producto = la capability:** un paquete CIERRA al ratificar su(s) capability(es).
  (Vitalia lo prueba: cap durable + `created_in_story`/`merge_sha`; story efímera.)

## D17 · Fix del concepto "épica" (PM research + anti-Vitalia-fail) — DECIDIDA
- UNA jerarquía limpia **outcome→paquete→sub**, cada uno instancia-con-estado + criterio de cierre.
  Framing por **OUTCOME** (Opportunity Solution Tree), no épica-cajón-de-output.
- **Release/fase = TAG ortogonal**, jamás en el slug/id (Vitalia mete `fase2` en el slug → choca con `release`).
- **Snapshot-vivo (auto-gen, DO-NOT-EDIT) ≠ bitácora append-only (LEDGER);** nada de prosa-log en YAML.
- Pipeline de artefactos que **ESCALA al appetite** (Shape-Up), no forzar 00→07 en todo.
- Insumo: `research-proyecto-ejecucion.md`.

## D18 · Regla de las 3 caras (principio de ubicación agnóstico) — DECIDIDA
Todo *concern* (requisitos, ux, arquitectura, datos, seguridad…) puede mostrar hasta **3 caras**, y la
cara decide DÓNDE vive — independiente del rubro:
1. **NORMA / definición** (durable, la comparte todo el proyecto) → **dimensión de terreno**
   (Propósito/Producto/Organización según su naturaleza).
2. **PASO** (cuándo se toca · qué rol · qué plantilla) → **caja del `proceso`** (Organización).
3. **ARTEFACTO** (la plantilla LLENADA para este paquete) → vive **dentro del paquete** (WIP) y lo refina.
   El **paquete** = vessel/instancia raíz (su schema lo define `gestion-trabajo`); los artefactos = docs-hijos.
   Ver **D20** (corrige el rótulo: "ARTEFACTO → dentro del paquete", NO "= la instancia").
- **Discriminador:** ¿regla/estándar estable que todos comparten? → dimensión. ¿momento donde alguien hace
  algo? → caja de proceso. ¿lo producido para este paquete? → WIP.
- **Resultado estructural (coherencia del modelo):** las 3 caras mapean **1:1** al modelo — NORMA→dimensión ·
  PASO→`proceso` · ARTEFACTO→WIP. **Un solo hogar por cara.** `proceso` agrega la cara-PASO de todo concern;
  WIP agrega la cara-ARTEFACTO de todo concern. Por eso "dónde va cada cosa" es mecánico.
- **Ejemplo UX:** design-system/tokens/Storybook = NORMA (dimensión de Producto) · "pasa por UX" = PASO (caja
  de proceso) · el mockup de este feature = ARTEFACTO (WIP).
- **Uso:** se aplica al re-barrer las 10 canónicas y a toda dimensión futura (mata el conflado type/instance).

## D19 · Consolidación del set canónico (aplica D18; supersede D13/D4 donde cambió) — DECIDIDA
**11 canónicas de DEFINICIÓN + 2 overlays; WIP = territorio de INSTANCIA (no aporta canónicas).**

| territorio (naturaleza · Essence) | id canónico | Zachman | universal | cambio vs D13 |
|---|---|---|---|---|
| **PROPÓSITO** (definición · Customer) | `encargo` | Why | por qué existe el trabajo / oportunidad | = |
| | `stakeholders` | Who | para quién · afectados | = |
| | `necesidad` | What | el problema/outcome a resolver (nivel-problema, estable) | **RENOMBRA** ex-`requisitos` (su cara problema) |
| **PRODUCTO** (definición · Solution) | `requisitos`/alcance | What | contrato: qué debe CUMPLIR la solución (durable; peso por rubro) | **MIGRA** desde Propósito (su cara solución) |
| | `producto` | How/What | estructura del entregable = Σ capabilities | = |
| | `datos` | What | insumos / materia prima de datos | = |
| | `contexto` | Where | perímetro / integraciones / sistemas externos (C4) — deslindado de `requisitos`[What] | afinado (F2) |
| **ORGANIZACIÓN** (definición · Endeavour) | `proceso` | When/How | la SECUENCIA: espina fases→cajas→ROL | = |
| | `equipo` | Who | quién ejecuta · RACI | = |
| | `forma-trabajo` | How | convenciones · prácticas · disciplina | = |
| | `gestion-trabajo` | How | **NUEVA**: schema del work-management (máquina de estados · jerarquía outcome→paquete→sub · backlog/planning · WIP-caps · `done/` · cierre=ratifica capability). Forja el **esqueleto vacío de WIP** | **NUEVA** |
| **WIP** (INSTANCIA · Work) | — | — | paquetes que se pueblan CONTRA el schema de `gestion-trabajo`; ratifican capability → Producto. Hogar de la cara-ARTEFACTO (D18) | naturaleza instancia (no canónica) |
| **OVERLAY** | `calidad` | — | catálogo ISO 25010 / por-rubro → capabilities NO-funcionales | = (D3) |
| **OVERLAY** | `economia` | Why | costo/valor permea toda decisión; carve-out para modelo económico durable | **PASA a overlay** (ex-canónica Organización · F3) |

- **Deslinde `gestion-trabajo` vs `proceso`:** `proceso`=la SECUENCIA de fases/pasos que un paquete atraviesa;
  `gestion-trabajo`=el CONTENEDOR/lifecycle/planificación (estados, jerarquía, backlog, cierre). La máquina de
  estados se ALINEA con las fases de `proceso`, no las duplica.
- **WIP = territorio de instancia:** el forjador monta su esqueleto vacío desde `gestion-trabajo`; el contenido
  churnea en runtime. No aporta canónicas (por eso el barrido D18 no lo lista).
- **Gate de completitud (actualiza D13):** la forja FALLA si falta una de las **11 canónicas**, o un catálogo
  de cada overlay (calidad, economía), o `label_rubro` (E), o `receta` por canónica, o *(rec)* decisión sin
  rule/skill trazable. Todo determinista/verificable (extiende `arnesia conformance` + gates PARIDAD).
- **Respaldo:** Essence (Way-of-Working[definición] vs Work[instancia]; Requirements en área Solution, no
  Customer) · Vitalia (schema/plantillas en el kit, work-items en el brand).
- **Resuelve:** F1 (`requisitos` migra a Producto, `necesidad` queda en Propósito; el territorio NO se renombra —
  se nombra por tema, no por miembro) · F2 (contexto[Where] vs requisitos[What]) · F3 (`economia`→overlay) ·
  acotación WIP (su schema es NORMA→`gestion-trabajo`, definición; el contenido es ARTEFACTO→WIP).
- **Pendiente de refresco (no bloquea):** `estructura-terreno.html` (visual v3) y `HANDOFF.md` aún dicen "10
  canónicas / requisitos en Propósito / economia canónica" → actualizar al rebuild.

## D20 · Paquete vs Artefacto (2 niveles) + multi-pipeline por tipo-de-paquete — DECIDIDA
- **Vocabulario afilado (corrección del user):** la INSTANCIA de WIP es el **paquete de trabajo** (vessel con
  estados; su schema lo define `gestion-trabajo`). Los **artefactos** son los documentos intermedios que cada
  PASO del proceso llena y que se **acumulan DENTRO del paquete, refinándolo** (dev: la historia = paquete; su
  research/mockup/spec/tickets = artefactos). Modelo: **WIP = árbol de artefactos** con el paquete como raíz;
  la plantilla vacía vive en `proceso`, el artefacto llenado en el paquete.
- **Multi-pipeline (hallazgo del caso RRHH):** un arnés tiene **N pipelines, uno por TIPO de paquete** — no uno
  solo. Dev lo escondía (casi todo = "historia" con un ciclo); RRHH lo grita (vacante ≠ desempeño ≠ caso
  laboral); finanzas en medio (valuación ≠ auditoría ≠ presupuesto). Entonces:
  - **`gestion-trabajo`** declara los **tipos-de-paquete**, cada uno con `{jerarquia, estados, criterio_cierre}`.
  - **`proceso`** deja de ser *un* spine → es un **mapa `{tipo-de-paquete → spine (pasos→rol→plantilla→artefacto)}`**.
  - Generaliza D17 ("pipeline escala al appetite") → **múltiples pipelines por tipo**.
- **`capability` cambia de FORMA por rubro, no de ROL:** feature (dev) · modelo/informe reutilizable (finanzas) ·
  registro-durable + playbook (RRHH). Rol constante: lo durable que el paquete ratifica al cerrar → constituye
  el Producto. En rubros de servicio el output es más evento/registro que artefacto reutilizable; la parte que
  enriquece el Producto durable es el **playbook/metodología reutilizable** + el **registro auditable**.
- Baja al **`arnes.yaml`** (dogfood dev) de este paquete.

## Por profundizar (diferido — modelo base firmado; revisar al avanzar)
No bloquean lo firmado; se cocean cuando su etapa llegue:
- **P1 · `capability` en rubros de servicio:** RRHH/consultoría → output = evento/registro + playbook, no
  artefacto reutilizable. Precisar qué ratifica el cierre y qué entra al Producto durable. (abierto desde D20)
- **P2 · Mecanismo `expande`:** cómo se declaran las sub-dimensiones de `producto` (dev: arquitectura→FE/BE ·
  ux · tecnologías · concurrencia) — ¿sub-nodos con su propio `{zachman, receta, knowledge}`? Schema del sub-árbol.
- **P3 · `cond` de los spines:** cómo evalúa el motor `tiene_ui` / `toca_arquitectura` (flags del paquete ·
  detección auto · pregunta). Define el appetite escalable real (D17/D20).
- **P4 · `receta.scraping`:** implementación del auto-derivado por dimensión (SBOM · storybook-index ·
  structurizr-dsl) + su verificación en CI (D6/D9).
- **P5 · Knowledge hoja (D9) ↔ `arnes.yaml`:** cablear el schema de hoja de conocimiento con las rutas
  `knowledge/` que declaran las dimensiones. Traza bidireccional.
- **P6 · Reconciliar con capabilities existentes (82):** el terreno = capa de DEFINICIÓN encima de
  `capabilities/` (= Producto durable). Mapear dimensiones ↔ modules sin romper R1-R4.
- **P7 · Refresco `HANDOFF.md` §1-§7** (quedó stale; el banner ya apunta a D18-D20).
- **P8 · Overlays carve-out:** seguridad · testing · compliance · economía-durable → carpeta + proyección; formalizar.

## Abiertas (siguiente)
- ✅ Modelo cerrado: 4 territorios (Propósito·Producto·Organización·WIP) + Calidad · 3 planos (definición/
  ejecución/producto) · capability=cierre · fix-épica. Wave 1+2+proyecto research todo hecho.
- ✅ `estructura-terreno.html` actualizado al modelo nuevo (v3).
- **Diseñar el schema `arnes.yaml`** (linchpin as-code) — probablemente como instancia del arnés-dev (dogfood).
- **Dogfood:** forjar el arnés-dev (Vitalia) para validar end-to-end.
- Rebuild del mapa-UI DESTINO.
- **Rebuild del product-map** «DESTINO» como superset del baseline con ESTE modelo.
  ⚠ `mockups/arnesia-mapa-destino.html` (draft previo) quedó **STALE**: se construyó con el modelo
  conflacionado (todo bajo «Terreno»); se rehará con Terreno=todo / Dimensión=eje / Calidad=overlay.

## Nota sobre las 4 decisiones del mockup destino previo
Deriva (capa autoría), Sello (manifiesto+hashes), capas-overlay-apilable siguen VIGENTES. La #2
(«región Terreno») queda **reencuadrada**: no es una región — es el landscape completo agrupado por
Essence, con Conocimiento co-locado como capa, no región peer.
