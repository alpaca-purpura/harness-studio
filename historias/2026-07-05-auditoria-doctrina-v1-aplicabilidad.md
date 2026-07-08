# Auditoría de la doctrina propia v1 (HS-07) — foco: aplicabilidad para HS-08 dogfood

> **Fecha:** 2026-07-05 · **Alcance:** todo lo bajado as-code de la doctrina v1 (VISION §Linaje,
> METODOLOGIA §3/§8, nodo `harness-profile`, 2 boundaries HS-07, contrato/schemas, y su fuente
> `research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`). **Método:** 5 subagentes read-only en
> paralelo (coherencia documental · árbol knowledge · arch/boundaries+schema · aplicabilidad ·
> fidelidad DAOP+fundamento de fuentes) + cross-check manual del schema, el dominio Go y los fitness
> tests. **Lente rectora:** ¿se puede APLICAR la doctrina (forjar el arnés `dev-full-cycle` real)
> siguiéndola escrita, o solo tenerla guardada?

## Veredicto

La doctrina está **madura como forma y como guía de pensamiento**, pero **NO como especificación
ejecutable**. La FORMA cuadra (knowledge = 12 nodos · 138 checks, verificado check por check; arch =
16 boundaries · 97 checks, aritmética consistente; `harness-profile` estructuralmente paralelo a los
otros 11; T1–T4 idénticos entre METODOLOGIA §8.2 y el nodo; Regla de Rosetta a nivel de sustantivos
respetada; firewall bien fundado técnicamente). Lo que **no ocurrió** es la capa *ejecutable* de
HS-07: el schema que debía cargar el contrato fusionado sigue en su versión it.10 y **rechaza** un
contrato firmado real; los `enforced_by:` de los 2 boundaries nuevos apuntan a tests inexistentes; falta
el **meta-modelo para que cada arnés declare SU spine de estados** + su check de consistencia (NO un enum
universal — ver corrección de agnosticismo en B2); y el validador/conductor/scaffold que darían sentido al
dogfood no están construidos.

**Patrón raíz:** HS-07 tocó VISION y METODOLOGIA para *añadir* doctrina en prosa, pero (a) no
reconcilió los artefactos de estado/roadmap de esos mismos docs, y (b) no bajó la prosa a schema +
fitness. CLAUDE.md, LEDGER y los dos INDEX están al día; **VISION y METODOLOGIA arrastran el snapshot
HS-05** y son hoy guías poco confiables para forjar. Además hay un **defecto de fundamento de
fuentes** (lavado de autoridad académica sobre axiomas de framework) que pone en riesgo el relato "no
somos clon de BMAD", y **una contradicción lógica importada sin resolver** (document-as-cache).

**Bottom line despiadado:** un autor no puede terminar el `contract:` de UNA sola caja sin adivinar
—porque el schema fusionado lo rechazaría y el arnés aún no tiene dónde declarar SU spine— y no hay red
(validador) que le diga si lo hizo bien. Los dos primeros bloqueantes (schema fusionado, semántica de
`clase`) son puramente de doctrina/schema; el meta-modelo de spine per-arnés (B2) aterriza CON el arnés
dogfood. Todo ANTES de forjar el dev-full-cycle.

> **Corrección (post-auditoría, 2026-07-05):** el hallazgo B2 estaba MAL enmarcado — trataba el spine
> como un "vocabulario cerrado central" del producto a enumerar. **Error de agnosticismo:** ArnesIA CREA
> arneses para procesos arbitrarios; NO tiene un spine universal. El as-code ya es agnóstico-correcto
> (`domain.Fase`/`Estado` son `string` sin enum, deliberado). B2 recalibrado abajo a **media** = falta el
> meta-modelo per-arnés, no una constante. Ver `research/2026-07-05-plan-goal-resolver-doctrina-v1.md`
> Principio Rector 8 + Ola 0.1/0.5.

| Severidad | # hallazgos | Naturaleza |
|---|---|---|
| **Alta (bloquea aplicar)** | 8 | contrato sin diente · `clase` roto · enforcers fantasma · herramienta no construida · lavado de autoridad · contradicción importada |
| **Media** | 13 | meta-modelo de spine per-arnés (B2 recalibrado) · roadmap/conteos stale · META sin aterrizaje · capacidades no operacionalizadas · firewall parcial · sin árbol de decisión |
| **Baja** | 8 | deuda semántica · ruido taxonómico · solapes no declarados · seam de un solo lado |

---

## BLOQUEANTES (severidad alta)

### B1 · El contrato fusionado no tiene diente: el schema está congelado en it.10 y RECHAZARÍA un contrato real
Confirmado por 3 subagentes + cross-check manual.

- **Dónde:**
  - `arch/contracts/schema/box.contract.schema.json:5` la propia descripción se ancla a "iteración 10
    de HS-03" (pre-doctrina); `:8` `"additionalProperties": false` (también en `gate` y en items de
    `entrega`).
  - `METODOLOGIA.md:137-138` afirma "Schema validado por `box.contract.schema.json`"; `arch/INDEX.md:16-17`
    dice "es **el mismo schema**". **Falso hoy.**
  - `internal/domain/box.go:104-112` el struct `Contract` espeja el schema viejo (solo
    `Caja/Fase/Estado/Necesita/Entrega/Ruta/Gate`; `Gate` = solo `Tipo/Detalle`; `Output` = solo `Art`).
  - `arch/fitness/arch_test.go:191` `TestBoxContractValidatesAgainstSchema` está en `t.Skip("TODO fase 5")`.
- **Qué falta en el schema** (todos presentes en METODOLOGIA §3:140-183, ausentes en el schema; con
  `additionalProperties:false` cada ausencia = **rechazo activo**): `why`, `capabilities[id/what/success]`,
  `constraints`, `non_goals`, `clase`, `arquetipo`, `perfil_harness`, `gate.aceptacion[given/when/then]`,
  `gate.evidencia`, `entrega[].escritor_unico`, `handoff[cuando/a]`. Lo único fiel es el eje **CABLEADO**
  (caja/fase/estado/necesita/entrega.art/ruta/gate.tipo/detalle), incluidos sus patrones regex.
- **Sustento:** la propia doctrina dice que validar contra ese schema "ES la fitness function del
  eval-gate A4" (schema `:5`). Hoy esa fitness ni existe (skip) ni cubre el contrato real. CLAUDE.md
  vende "contrato FUSIONADO … bajado as-code": la mitad as-code no ocurrió.
- **Impacto al APLICAR:** es el bloqueador #1 de HS-08. El primer `contract:` real llevará
  `why/capabilities/arquetipo/perfil_harness/gate.aceptacion/handoff/escritor_unico` → `arnesia conformance`
  **falla en todo contrato bien formado**; o si se afloja `additionalProperties`, valida-pero-ignora
  los ejes nuevos (el eval-gate A4, el Gherkin ejecutable, el mutation-contract y la clasificación
  quedan sin enforcement).
- **Cerrar:** reescribir `box.contract.schema.json` + `domain.Contract` al contrato fusionado (3 ejes)
  y activar `TestBoxContractValidatesAgainstSchema`.

### B2 · [RECALIBRADO a MEDIA] Falta el meta-modelo para que cada arnés declare SU spine + su check de consistencia
> **El encuadre original de este hallazgo era ERRÓNEO** (violaba el agnosticismo). Corregido tras el
> re-audit de agnosticismo. El texto de abajo es la versión correcta; conservado como registro (aditivo).

- **Encuadre correcto:** ArnesIA es **agnóstico al proceso** (VISION p3/p7): CREA arneses para
  roles/procesos arbitrarios, así que **NO tiene un spine universal**. "No enumeramos los estados" es
  **correcto por diseño**, no un hueco. El as-code YA es agnóstico-correcto: `box.contract.schema.json:22`
  valida `estado` solo por patrón `X -> Y` (sin enum), `fase` es string libre, y `domain.Fase`/`Estado`
  son `type string` **sin `const`/`Valid()`** — deliberado, a diferencia de `Clase`/`Banda`/`GateTipo`
  (que sí tienen enum, porque son doctrina). `fase` (agrupador de cajas) ⊥ `estado` (transición del
  work-item) **ya están decididos como ejes distintos** (A1 vs §3, dos campos separados).
- **El hueco REAL (síntoma legítimo):** hoy el schema valida UNA transición aislada, pero **nada permite
  a un arnés declarar SU spine ni SUS fases**, así que la fitness function **no puede** validar lo que la
  doctrina promete (METODOLOGIA §6): que las transiciones encadenen, sin estados/fases huérfanos, con
  integridad `nodo.fase ∈ fases-declaradas`. `graph.l0.schema.json:5` lo admite ("por-formalizar… fase 4").
- **Aclaración de las citas mal leídas:** `METODOLOGIA.md:227` "spine de 10 estados" es un **smell
  documental** (prosa de §5, insight de un arnés real de 141 nodos) — no impone nada (no hay enum detrás);
  además el ejemplo luana solo tiene ~7 estados (`UX.md:500-504`). `UX.md:555` "6 fases = spine" es el
  **eje-fase** del mismo ejemplo luana, NO el eje-estado — no hay "10 vs 6" que reconciliar: son dos ejes
  ortogonales de UN arnés ejemplo. La FSM `draft→…→blocked` sigue siendo la FSM interna de caja T3, distinta.
- **Impacto al APLICAR:** el autor del dev-full-cycle no puede escribir `estado:` **hasta que ESE arnés
  declare su spine** — pero eso es dato del arnés (aterriza CON el dogfood), no una constante ausente del
  producto. Naturaleza: item normal de fase 4/5, no bloqueante-alto.
- **Cerrar (agnóstico):** (1) `domain.Spine` como **TIPO/forma** (estados + transiciones legales), cero
  valores en el core; (2) el **manifiesto del ARNÉS** (`graph.l0`) declara `fases:[...]` +
  `spine:{inicial, terminales, estados:[...]}` como DATO per-arnés; (3) `contract.estado` queda como
  patrón (schema de caja intacto); (4) checks de conformidad **parametrizados por el spine declarado**:
  `estado-en-spine-declarado`, `transicion-legal`, `una-transicion-por-caja`, `spine-cobertura`,
  `fase-en-fases-declaradas`. Los `idea→…→released`/6 fases viven como **fixture del arnés dogfood**, no en
  `domain`/schema del producto. Ver plan Ola 0.1/0.5 + 1.6.

### B3 · `meta.clase` (I-75) tiene DOS ontologías incompatibles, ambas citando el mismo I-75
- **Dónde:**
  - `graph.l0.schema.json:44` enum = `["guardia","fase","base","libreria-expertos","meta-harness","marcas-dormidas","terceros"]` (banda/rol del mapa), comentado "meta.clase, I-75".
  - `METODOLOGIA.md:152` `clase: <skill|agente|hook|knowledge|mcp|regla|command>` (tipo-de-elemento), "meta.clase L0 (I-75)".
  - `harness-profile.md:42-43` construye su tesis central sobre `clase ⊥ perfil_harness` asumiendo la 2ª ("una skill-caja (clase=skill) puede ser T1 o T3").
- **Además:** el enum de METODOLOGIA usa labels legacy (`agente`≠`subagent`, `regla`≠`rule`), incluye
  `knowledge` (que no es un elemento) y **omite** plugin/settings/output-style/statusline/headless/harness-profile
  → no cubre los 12 nodos de `knowledge/`. `graph.l0.schema.json:38` `$comment` dice "TODO **HS-06**:
  converger enum a los 11 tipos" — HS-06 ya pasó, HS-07 subió a 12 y no lo reconcilió. Y `clase` está
  **ausente** de `box.contract.schema.json` (B1) y **sin check** en ningún nodo de knowledge (arquetipo
  y perfil sí tienen; clase no).
- **Impacto al APLICAR:** `meta.clase: skill` (doctrina) no valida contra el enum L0 (que espera `fase`);
  nadie sabe qué valor poner. La ortogonalidad `clase ⊥ perfil` —fundamento del nodo 12— es ambigua
  porque `clase` significa dos cosas.
- **Cerrar:** decidir la ontología única de `clase` (tipo-de-elemento, alineada a los 12 nodos, labels
  canónicos), separarla del eje banda/rol del mapa si hace falta un segundo campo, y darle check.

### B4 · Los 2 boundaries de doctrina declaran `enforced_by:` a tests que NO existen (ni como stub)
- **Dónde:** `arch/boundaries/orquestacion-determinista-entre-cajas.md:17` → `fitness/arch_test.go:TestConductorOwnsBoxRouting`;
  `arch/boundaries/permisos-derivan-del-rol.md:17` → `fitness/arch_test.go:TestPermissionSetParametrizedByRole`.
  **Ninguna de las dos funciones existe** en `arch_test.go` (21 funcs `Test*`, verificado). 3 de los 4
  checks de cada boundary nombran enforcer genérico `"arch_test.go"` sin función.
- **Sustento:** el patrón establecido del repo es que todo boundary `proposed` trae un test placeholder
  con `t.Skip` (existen `TestNoJSONLSchemaParsing`, `TestWriteRequiresApproval`, etc.). HS-07 rompió
  ese patrón: ni el stub. Es honesto en el estado (nacen `proposed`/🌱, no se disfrazan de enforced,
  a diferencia de los 2 de HS-06 que nacen enforced y **pasan**), pero el `enforced_by:` cuelga a vacío.
- **Impacto al APLICAR:** cuando `arnesia conformance` lea `enforced_by:` y ejecute la función nombrada,
  dará "no such test", no un skip legible → se rompe la enumeración de checks. Además, checks como
  `no-infiere-del-texto`, `ruta-ejecutada-por-codigo`, `agencia-dentro-del-frame` son propiedades
  semánticas sin diseño de cómo un fitness test las observaría.
- **Cerrar:** crear los stubs `t.Skip` mínimos ya, y diseñar cómo se observa "el conductor Go ejecuta
  `contract.ruta`, no el LLM".

### B5 · `arnesia conformance` no existe: los 138 checks son markdown no ejecutable → contradicción de secuencia con dogfood-first
- **Dónde:** `knowledge/INDEX.md:83-88` (autoconfesión) "hoy estos checks NO llevan `enforced_by`…";
  `arch/INDEX.md:22-25` idem; el único `arnesia conformance` en el repo es un **comentario**
  (`arch_test.go:11`); `grep conformance **/*.go` = cero implementación. `METODOLOGIA.md:262` promete
  "El linter que ArnesIA correrá (**fase 5**)".
- **Sustento:** dogfood-first (HS-08, fase 4) manda **conformar** el arnés (§6: cargar→verificar→corregir→remapear)
  con un validador que se construye **después** (fase 5). El validador que da sentido al dogfood no
  existe cuando el dogfood ocurre.
- **Impacto al APLICAR:** hueco de secuencia. Hay que decidir **explícitamente** cómo se valida el
  dev-full-cycle sin el runner (revisión manual contra las tablas, o codificar primero un subconjunto
  fundacional de checks) — o adelantar un `arnesia conformance` mínimo.
- **Cerrar:** decisión de secuencia + (mínimo) codificar el subconjunto de checks que el dogfood
  necesita (schema del contrato, huérfanos/dead-ends, spine).

### B6 · El conductor T3 está diseñado pero no construido → el arquetipo T3 no es aplicable HOY
- **Dónde:** `METODOLOGIA.md:291-295` y `harness-profile.md:104-106` describen T3 (conductor Go dueña
  el loop, `--max-turns N`, lee `result`+`status`, cap de reparación, `blocked→handoff`, ejecuta
  `contract.ruta`). Lo que EXISTE en el conductor real: `--max-turns` (`TestMaxTurnsAlways` pasa) y
  turno-a-la-vez (`TestOneTurnAtATime`). **NO existe:** loop de reparación, lectura de `status` del
  artefacto, estado terminal `blocked`, ejecución de `contract.ruta`. La FSM `draft→…→blocked` no está
  en el dominio (grep = 0).
- **Impacto al APLICAR:** solo T1 (un pase) es genuinamente ejecutable hoy. Ninguna caja del
  dev-full-cycle puede declararse `perfil_harness: T3` hasta construir el for-loop del conductor. Es
  **dependencia dura de código**, no de doctrina.
- **Cerrar:** construir el loop del conductor (spawn iterado + repair-cap + lectura de `status` +
  `blocked→handoff`) y crear `TestConductorOwnsBoxRouting`.

### B7 · Lavado de autoridad: axiomas de DAOP/BMAD/12-Factor re-atribuidos al manifiesto académico "Agentic BPM"
Smoking gun: el subagente leyó el fuente `~/Descargas/doctrina-kits-v0.2.md`.

- **Dónde:** `METODOLOGIA.md:301` "### 8.3 Document-as-cache … (**A7 de APM**)"; `arch/boundaries/orquestacion-determinista-entre-cajas.md:25,58` "(**A2 de Agentic BPM** + 12-Factor)". Fuente:
  `~/Descargas/doctrina-kits-v0.2.md:59` "**A2 — Orquestación determinista**" y `:64` "**A7 — Estado en
  el artefacto…** (Es el *document-as-cache* de **BMAD** y el 'unify execution & business state' de
  12-Factor.)"
- **Discrepancia:** los "A#" son los **Principios de DAOP**, no del manifiesto APM. DAOP mismo define A7
  como "el document-as-cache de BMAD" y A2 lo ancla a 12-Factor. La bajada les estampa "de APM" →
  atribuye al paper académico conceptos que la fuente atribuye a BMAD/12-Factor. El barrido APM del
  research (§11) **nunca** lista document-as-cache como tenet APM.
- **Impacto al APLICAR:** cita fantasma — quien vaya a la bibliografía a fundamentar "document-as-cache"
  con el manifiesto APM no lo hallará ahí. Erosiona la confianza en TODAS las citas académicas y
  compromete el relato "no clon de BMAD" (el concepto ES de BMAD, re-etiquetado).
- **Cerrar:** re-atribuir "A7/A2" a DAOP/BMAD/12-Factor; separar lo que el manifiesto APM sí aporta
  (framed autonomy, autonomy≠automation, 4+1, adaptation/evolution) de lo que viene de framework.
  **Verificar fuera de banda** que arXiv 2603.18916 existe y dice lo atribuido (ver M9).

### B8 · Contradicción lógica importada sin resolver: document-as-cache obligatorio (T2) vs exento (arquetipo abierto)
- **Dónde:** `METODOLOGIA.md:290` "T2 · … document-as-cache **obligatorio**"; `:305` "**Obligatorio en
  T2/T3**"; `:280` "abierto — … puede acumular estado de sesión (**excepción a document-as-cache
  estricto**)". Origen: `~/Descargas/doctrina-kits-v0.2.md:202,240` marca esto como `[PENDIENTE 7.A]`.
- **Discrepancia:** `arquetipo` y `perfil_harness` son ejes **ortogonales** (METODOLOGIA §3:151-155;
  `harness-profile.md:43`). Luego una caja puede ser legítimamente `arquetipo:abierto` **y** `perfil:T2`:
  §8.2/§8.3 la **exigen** document-as-cache, §8.1 la **exime**. DAOP lo dejó abierto; ArnesIA lo bajó
  **como firmado** sin resolver, y lo empeoró al partir el eje único de DAOP en dos ortogonales.
- **Impacto al APLICAR:** muerde en el caso más interesante (workflows generativos largos con estado).
  El linter no sabrá si exigir el frontmatter `status:` en esa caja; el conductor T3 que "relee el doc"
  no tiene contrato claro cuando la caja es abierta.
- **Cerrar:** heredar la recomendación de DAOP `[PENDIENTE 7.A]` ("estricto para pipeline/excepción;
  abierto = estado de sesión con destilado al cierre") **o** declarar precedencia arquetipo > perfil.

### B9 · Colisión de etiquetas "A#": la anatomía propia A1–A7 de VISION vs los axiomas DAOP A2/A7/A12 en crudo
- **Dónde:** anatomía PROPIA firmada `VISION.md:61-87` (A2 = contrato input→output; A7 = "crear un arnés
  = definir sus fases"). Axiomas DAOP en crudo: `arch/boundaries/orquestacion-determinista-entre-cajas.md:25` (A2 = orquestación determinista), `METODOLOGIA.md:301` (A7 = document-as-cache),
  `harness-profile.md:117` "reuso por referencia **A12**" (sin definir — vocabulario DAOP filtrándose a
  un nodo de knowledge, justo lo que la Regla de Rosetta buscaba impedir).
- **Impacto al APLICAR:** un lector que vea "A7" no puede saber si es "crear=definir fases" (VISION) o
  "document-as-cache" (METODOLOGIA), y ambos son norte constitucional. Cualquier check/spec que
  referencie "A2"/"A7" es ambiguo.
- **Cerrar:** renombrar el sistema de axiomas importados (p.ej. `DAOP-A2`) o traducirlos a nombres
  propios; purgar "A12" del nodo de knowledge.

---

## IMPORTANTES (severidad media)

### M1 · VISION "El gran plan" arrastra el snapshot HS-05 (roadmap y conteos stale, y se auto-contradice)
- `VISION.md:225` "arch/ = **77 checks**" → real **97** (`arch/INDEX.md:67`). Faltan +12 de HS-06 y +8
  de HS-07 (los 2 boundaries de doctrina quedan invisibles).
- `VISION.md:227` "| 4 | Definición de specs | **siguiente (HS-06)** |" → real HS-08 (`LEDGER.md:330`,
  `CLAUDE.md:88`). La tabla ni menciona que HS-06/HS-07 ocurrieron.
- **Auto-contradicción:** `VISION.md:227` "MVP = **Mapa primero**" vs `VISION.md:124-126` (§Linaje,
  RATIFICADO) "el arnés real … **ANTES del Mapa** … para que el Mapa renderice datos reales, no mocks".
  La propia doctrina v1 introdujo la contradicción y no reconcilió la tabla.
- **Impacto:** quien planifique leyendo la tabla construye el Mapa primero (con mocks) — exactamente lo
  que dogfood-first quería evitar.

### M2 · METODOLOGIA se contradice a sí misma en el conteo y en la ficha de specs
- `METODOLOGIA.md:18` "(**121** al corte fundacional)" vs `:261` "Los **138** checks son el ruleset de
  conformidad". El mismo doc describe "el ruleset de §6" con dos totales.
- `METODOLOGIA.md:348` "las specs de fase 4 (**HS-06**)" — mientras `:349` ya reconoce "Doctrina v1
  (ficha HS-07)". Verdad: specs = HS-08.
- **Impacto:** METODOLOGIA es el norte de reglas de negocio de HS-08; su cabecera y §Estado mandan a
  números y fichas obsoletos.

### M3 · La "META de enganche" (decisión firmada) no tiene aterrizaje as-code
- Firmada en `VISION.md:119-122` y `CLAUDE.md:69-71`: "cada arnés carga rol·proceso·reporta-a·empresa".
  Pero el único schema (contrato §3, **por-CAJA**) no la modela, y **no existe manifiesto a nivel de
  ARNÉS** que la contenga. `harness-profile` (nodo 12) cubre el loop de ejecución, no la META.
- `graph.l0.schema.json:14-21` el objeto `arnes` tiene `{id, puesto, empresa, reporta_a, canal,
  marketplace}` → **falta `proceso`**, el rol está como `puesto` (otro nombre), y **nada es required**
  (el propio `arnes` es opcional). El check `meta-de-enganche-completa` (`permisos-derivan-del-rol.md:57`,
  enforcer "schema graph.l0") es por eso **inejecutable**: valida-todo-vacío.
- **Impacto:** HS-08 forja el arnés real; la doctrina exige que cargue la META pero quien siga
  METODOLOGIA §3 no encuentra dónde escribirla → riesgo directo de omitirla.
- **Cerrar:** agregar `proceso` (y decidir `puesto` vs `rol`) a `arnes`, marcar los 4 required, y
  decidir dónde vive la META a nivel de ARNÉS (manifiesto, no contrato de caja).

### M4 · Permisos-por-rol + TTL es aspiracional, y el spike está mal atribuido
- `arch/boundaries/permisos-derivan-del-rol.md:45` "Permiso efímero (spike **HS-07**)" — HS-07 fue
  doctrina-as-code, **no** hizo el spike; CLAUDE.md lo agenda para HS-08.
- `arch/contracts/api/openapi.yaml:176-197` `POST /sessions/{id}/permission` es **stub allow/deny HITL**:
  body `{requestId, decision, updatedInput, message}` — **sin `role`, sin `ttl`, sin `expiresAt`, sin
  scope por-tarea**. El `KitProvisioner` que parametrizaría por rol no existe (grep = 0).
- **Impacto:** el boundary "permisos-derivan-del-rol", en su mitad rol-céntrica y TTL, es documento sin
  realización. No hay forma de expresar "este permiso expira en N" ni de que el gate lo verifique.

### M5 · `skills.md` no absorbió el contrato fusionado, y `clase` no tiene check en ningún lado
- `skills.md:95-98` (L2.1) lista solo "`version`, `model`, `clase` … y el bloque `contract:`"; el check
  `skill-caja-contract` (`:124`) exige "bloque `contract:` **completo**" sin enumerar sub-campos ni
  referenciar el schema (vago, no objetivamente verificable). **No menciona `arquetipo` ni
  `perfil_harness`** (que METODOLOGIA §3 volvió parte del contrato). `clase` no tiene check en ningún
  nodo (arquetipo y perfil sí lo tienen en `harness-profile`).
- **Impacto:** quien forje una skill-caja leyendo su nodo canónico no verá que debe declarar
  arquetipo/perfil, y un `clase` ausente/mal-tipado no lo caza nada.

### M6 · Los 6 patrones de subagente: triple redirección a un detalle que no existe
- `subagents.md:108-111` y `METODOLOGIA.md:299` remiten a `harness-profile.md` "por el detalle de los 6
  patrones", pero `harness-profile.md:66-69` solo **nombra** los 6 (Delegated Data Access · Temp File
  Assembly · Shared-File · Hierarchical Lead-Worker · Persona-Driven Parallel · Evolutionary) sin
  explicar cuándo/cómo usar cada uno.
- **Impacto:** HS-08 es forjar el arnés con subagentes; el autor necesita saber QUÉ patrón elegir y
  CÓMO cablearlo — el árbol le da 6 nombres pelados. Es el hueco más directo contra dogfood-first.

### M7 · Sin árbol de decisión para `arquetipo` / `perfil_harness`; los checks solo verifican presencia, no correctitud
- `METODOLOGIA.md:273-297` define QUÉ es cada arquetipo/perfil pero no CÓMO decidir el de una caja. Los
  checks `arquetipo-declarado` / `perfil-tipo-declarado` (`harness-profile.md:127-128`) solo verifican
  que el campo **esté**, no que sea **correcto** (una caja generativa puede declararse `pipeline` y nada
  la caza). La frase estrella "saber cuándo NO arnesar es doctrina" (`METODOLOGIA.md:281-283`) no se
  operacionaliza: no hay check `no-arnesar-clasificado` (grep = 0) ni criterio.
- **Cerrar:** árbol de decisión (2-3 preguntas) para arquetipo y para T1/T2/T3, más regla operable de
  "cuándo NO arnesar".

### M8 · Explainability y las "4+1 capacidades": nombradas, no operacionalizadas, con aritmética incoherente
- Conteo roto propagado a doc firmado: `harness-profile.md:51` "APM debe proveer **4** capacidades"
  (incluye explainability); `research §11.2:333` y `VISION.md:110-111` las llaman "**4+1**" listando
  esas mismas 4; `research §11.3` presenta explainability como el "+1 nuevo". Si explainability es una
  de las 4 nativas, no es "+1"; y la quinta nunca se define.
- Sin bajada a mecanismo: `explainability-rationale` (`harness-profile.md:133`, warn) no tiene **campo en
  el contrato** para el rationale, ni mecanismo (¿telemetry-emit? ¿artefacto document-as-cache?), ni
  formato. `conversational actionability` (1 de las 4) no tiene ni L2 ni check en el nodo.
- **Impacto:** el checklist "¿este arnés provee las capacidades?" no sabe si son 4 o 5, y explainability
  —justo la faceta que se promueve a "5ª faceta evaluable"— queda sin dónde vivir.

### M9 · El manifiesto APM (arXiv 2603.18916) es autoridad load-bearing e inverificable
- `research §11.1` + `harness-profile.md:7-10` + `METODOLOGIA.md:331` + `VISION.md:98`. El research SÍ
  aporta quotes (bien), pero (a) inverificables y calzan sospechosamente sobre posiciones que ArnesIA/DAOP
  ya tenían; (b) el research admite que la función es retórica ("el linaje que nos saca de 'clon de
  BMAD'"); (c) el "predecesor arXiv 2201.12855" sería de **enero 2022**, dudoso para "Agentic BPM". El
  DOI/journal (Information Systems, Elsevier) y los autores (Dumas, Montali, Rinderle-Ma, Weber) son
  reales — persuasivo pero no verifica el artículo ni las quotes.
- **Impacto:** el diferenciador estratégico central ("disciplina de proceso académica, no framework de
  agentes") es tan sólido como una cita no auditada. **Verificar en web antes de usar en material de
  producto.**

### M10 · El "framing mechanism" tiene distinto número de componentes según el doc
- `VISION.md:98` y `research §11.1:322` = **4 componentes** (normativo + operacional + conocimiento +
  tools); `harness-profile.md:50` (L1, nodo autoritativo del estándar APM) = **2 capas** (normativa +
  operacional). Los "+2" parecen construidos para mapear sobre la banda Base y los MCP de ArnesIA, no
  citados del manifiesto.
- **Impacto:** "un arnés ES un framing mechanism instanciado" (load-bearing para el posicionamiento) no
  tiene descomposición estable; el estándar L1 solo respalda 2 de las 4 partes.

### M11 · Gate de fidelidad (§8.4) y document-as-cache reread (§8.3): convención sin quién la enforce
- `METODOLOGIA.md:301-306` (§8.3) dice literal "es **convención con disciplina**: la caja/conductor DEBE
  releerlo" — sin hook ni check que verifique el reread (`skill-document-as-cache` solo chequea que el
  estado esté en el frontmatter, no que se relea). §8.4 (gate de fidelidad = "dogfooding elevado a gate
  de promoción") **no tiene ningún check** (grep `fidelidad` = 0) ni procedimiento (cómo se corre el
  eval, dónde vive el doc, qué lo pasa/falla, quién lo dispara en promote).
- **Cerrar:** hook `SessionStart`/`PreCompact` o convención de conductor que fuerce el reread; y
  procedimiento operable del gate de fidelidad.

### M12 · Checks huérfanos prometidos en el research que nunca aterrizaron
- `research/…daop.md:242-249` prometió ~12 checks; **no llegaron** (grep = 0): `single-writer-per-artifact`
  (¡el de `escritor_unico`!), `no-arnesar-clasificado`, `mece-routing-coverage`, `snapshot-inmutable-versionado`,
  `anotacion-humana-a-regresion`, `persona-hard-constraint-hook`. Caso grave: `escritor_unico` es campo
  **obligatorio** del contrato (`METODOLOGIA.md:145,190`) pero (a) ningún check lo enforça y (b) el schema
  lo **rechaza** (B1). Un mutation-contract mandado por prosa, imposible de declarar y sin validación.

---

## MENORES (severidad baja)

- **m1 · CLAUDE.md desglose de boundaries no suma:** `CLAUDE.md:101-104` "9 backend + 5 FE" = 14, luego
  "+2 HS-06 +2 HS-07" → 18; el real es **11 backend + 5 FE = 16**. Ningún desglose cierra limpio.
- **m2 · Ficha huérfana del runner unificado:** `knowledge/INDEX.md:86-88` y `arch/INDEX.md:22-25`
  prometen "el contrato de check común… planificado como **ficha HS-06**"; HS-06 se ejecutó como
  endurecimiento y no lo entregó (los checks de knowledge siguen sin `enforced_by`). Destino real sin
  reasignar (candidato HS-08+).
- **m3 · `version/updated` sin bumpear:** `skills.md:2-3`, `subagents.md:3-4`, `rules.md:3-4` siguen en
  `1.0 / 2026-07-04` pese a changelog v1.1 e INDEX 1.1. CADENCE §0 lee `updated` como marca de agua →
  la próxima corrida semanal computa mal el watermark. (Contraste: `headless-sdk.md` sí subió a 1.1.)
- **m4 · `no-phantom-frontmatter` cubre 3 de 4 claves:** `skills.md:131` omite `sanctum` (PERSONA/CREED)
  y escribe `customize` (no `customize.toml`); su título ("solo claves que CC reconoce" = allowlist) es
  más fuerte que el detalle (denylist de 3). `rules.md:108` nombra solo 2 de 4. Un arnés con `sanctum:`
  pasaría el firewall silenciosamente. (`CLAUDE.md:67` y `LEDGER.md:301` también listan 3; la lista
  completa vive solo en `METODOLOGIA.md:326`.)
- **m5 · Solape de checks no declarado:** `autonomia-por-riesgo` y `tolerancia-declarada`
  (`harness-profile.md:134-135`) duplican el cross-check `rules-hard-enforcement` / `hook-rule-pair` /
  `perm-advisory-only` sin cross-referenciar ni actualizar la sección "regla↔hook↔permiso" de INDEX →
  un defecto real encenderá ~4 badges e inflará el conteo de hallazgos.
- **m6 · "evolution" mal mapeado:** `harness-profile.md:104-105` y `METODOLOGIA.md:296-297` mapean
  "evolution (persistente)" al **spine del work-item** (que es ejecución/instancia), contradiciendo la
  propia `VISION.md:108-109` (evolution = mejora persistente del modelo). Error de categoría; deuda
  semántica.
- **m7 · Ciclo de vida ADLC (L1.6) sin check:** `harness-profile.md:116-119` empaca snapshot-inmutable +
  bounded-error + anotación→regresión y se auto-marca "candidatos de producto, no solo checks". Honesto,
  pero deja un eje entero del estándar sin poder medirse.
- **m8 · Seam L1 externo, contrato de un solo lado marcado RATIFICADO:** `VISION.md:119-122` +
  `permisos-derivan-del-rol.md:44-48` delegan la autoridad rol→permisos a un "sistema externo futuro"
  con cero especificación. El boundary solo puede enforzar el **mecanismo**, no el **contenido**. Los
  campos META de `graph.l0` se fijan hoy sin saber qué necesitará ese sistema (riesgo de drift).
- **m9 · Ruido taxonómico en `harness-profile`:** `arquetipo-declarado` (`:128`) vive en el nodo de
  "cómo ejecuta" pero arquetipo es el eje de FORMA (§8.1); `cognitive-load-declared` (`:132`) deriva de
  "L1.4 · **L2**" (L2 sin número, rompe el patrón de trazabilidad). `nota "77 previos"` (`arch/INDEX.md:76`)
  stale (debería ser 85). "122 checks" stale en 4 sitios de arch. Conteo fundacional 121 (METODOLOGIA)
  vs 122 (LEDGER HS-03). El research resumen reporta conteos pre-bajada (122/89/11 nodos/14 boundaries).

---

## Lo que SÍ está bien (calibración del escepticismo)

- **Conteos núcleo correctos:** knowledge = 12 nodos · **138** checks (columna suma 138, verificado por
  nodo); arch = 16 boundaries · **71** boundary checks + **26** conventions = **97** (aritmética
  consistente). Los outliers son VISION (77) y el desglose de CLAUDE.md.
- **`harness-profile` estructuralmente sólido:** fuentes fechadas con autoridad, L1.1–L1.7 con cita,
  L2.1–L2.7 con `⇐ L1.x`, 11 checks, changelog. T1–T4 **idénticos** a METODOLOGIA §8.2.
- **Regla de Rosetta a nivel de sustantivos RESPETADA:** `arnés = framework por rol×proceso` intacto
  (`VISION.md:11-14`); DAOP-«Arnés»→CAJA y DAOP-«manifiesto»→ARNÉS se respetan (la colisión está en los
  axiomas A#, no en los sustantivos núcleo).
- **Firewall bien fundado técnicamente:** las claves prohibidas existen literalmente en el fuente DAOP;
  "CC ignora frontmatter desconocido" es trivialmente cierto por schema. `no-phantom-frontmatter` y
  `context-injection-native` aterrizaron (pese a m4).
- **Fidelidad de bajada material:** todo lo que el research firmó (contrato fusionado, nodo 12, 2
  boundaries, 3 decisiones) aterrizó. **No hay doctrina huérfana material ni doctrina fabricada** sin
  respaldo. El defecto no es fidelidad de bajada — es **fundamento de fuentes** (B7/M9) y
  **contradicciones importadas** (B8), más la capa **ejecutable ausente** (B1–B6).

---

## Orden de ataque para HS-08 (dependencias duras)

Cada ítem **desbloquea** los de abajo. Los 3 primeros son doctrina/schema (baratos); 4–7 son código y
definen hasta dónde llega el dogfood "a mano" vs "forjado".

1. **Meta-modelo de spine per-arnés** (B2 recalibrado, agnóstico): `domain.Spine` como TIPO + el
   manifiesto del arnés declara SUS `fases`/`spine` como dato + checks de consistencia parametrizados.
   **NO** enum universal en el core. Los valores concretos aterrizan como fixture con el arnés dogfood.
   *(Desbloquea escribir `estado:` en ESE arnés, no en el producto.)*
2. **Reescribir `box.contract.schema.json` + `domain.Contract` al contrato fusionado** y activar
   `TestBoxContractValidatesAgainstSchema`. *(B1.)*
3. **Fijar la ontología única de `clase`** (alineada a los 12 nodos) + darle check; resolver el cruce
   `abierto×T2` de document-as-cache (B8); desambiguar el sistema de axiomas A# (B9). *(B3/B8/B9 —
   doctrina pura.)*
4. **Definir cómo se valida el dev-full-cycle sin `arnesia conformance`** (revisión manual vs codificar
   el subconjunto fundacional). *(B5 — resuelve la contradicción de secuencia.)*
5. **Construir el loop del conductor T3** (repair-cap + lectura de `status` + `blocked→handoff` + ejecución
   de `contract.ruta`) y crear `TestConductorOwnsBoxRouting`. *(B6 — habilita `perfil_harness: T3`.)*
6. **Spike `control_request` + `KitProvisioner`** (rol + TTL) y crear `TestPermissionSetParametrizedByRole`;
   sube `permisos-derivan-del-rol` de proposed a enforced; agregar `proceso`/required a `graph.l0` (M3/M4).
7. **Scaffold + plantilla embebida del "arnés-de-crear"** si el forjado no es 100% manual.
8. **Bocetar el proceso dev-full-cycle** (fases + caja por fase) — output natural del dogfood, pero
   necesita 1–3 listos para expresarse. *(B... / M6 patrones de subagente.)*
9. **Barrido de sincronización de VISION + METODOLOGIA** (tabla del gran plan, 77→97, HS-06→HS-08,
   121→138, MVP-Mapa vs dogfood, version bumps, firewall 4 claves, re-atribución A7/A2) — barato, alto
   retorno en confiabilidad de la guía. *(M1/M2/B7/B9/m1–m4/m9.)*
10. **Verificar en web arXiv 2603.18916** antes de citar el posicionamiento APM en material de producto. *(M9.)*
