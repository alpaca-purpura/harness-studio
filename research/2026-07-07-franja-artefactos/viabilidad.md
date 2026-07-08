# Viabilidad: 4ª franja «Artefactos» — síntesis de la evaluación 5-frentes

Fecha: 2026-07-07 · Método: workflow de 13 subagentes — 5 lectores del as-code (doctrina,
código ejecutable, mapa FE, interop DevStudio, knowledge), 3 investigadores web
(artifact-centric BPM, spec-driven development, prácticas oficiales CC), 1 auditor de
consistencia + panel de 4 lentes (doctrina · UX-mapa · ejecución · abogado del diablo).
Los 4 lentes convergieron sin contradicción sustantiva.

## 1. Veredicto

**La intención es doctrina vigente; la forma propuesta (franja-banda) es la geografía
equivocada; y la mitad de mayor valor de la idea — plantillas + llenado casi determinista
+ verificación de composición — es genuinamente nueva y ataca el hueco doctrina↔motor más
grande del repo.**

- La idea NO es concepto nuevo: **A2 ya la cementa** — toda caja declara qué entra y qué
  sale «como artefactos as code (spec, arquitectura, código, review)» y «la salida de una
  caja es la entrada de la siguiente: el hand-off ES el contrato» (VISION.md:66-68).
- El dato YA existe: `necesita[].art` (+`de`, 7 orígenes, `requerido`) y `entrega[].art`
  (+`escritor_unico`) en el contrato fusionado (box.contract.schema.json:68-102;
  METODOLOGIA §3). El dogfood lo ejerce: spec-writer entrega `spec.md`, builder lo
  necesita `de: caja:spec-writer`.
- La mitad UI YA existe: el Inspector/drawer (RF-90) pinta `necesita`/`entrega` como chips
  tipados navegables al seleccionar la caja (inspector.tsx:467-506). La selección de caja
  existe end-to-end; el mecanismo focus/related+dim también.
- Lo que NO existe: **plantillas** (cero menciones as-code), **identidad del art** (string
  libre con semántica doble), **checks de composición** (prometidos en METODOLOGIA §3/§6,
  UX it.10 y el propio schema; implementado solo `escritor-unico`), **proyección espacial**
  del hand-off en el Mapa, y **precondiciones `necesita`** en el conductor T3.

## 2. Por qué NO franja-banda (y qué sí)

1. **A6:** las bandas son infraestructura compartida que actúa sobre todas las cajas
   (Guardia/Base); el artefacto es el TRABAJO que fluye (A3). Una banda «Artefactos»
   mezcla naturalezas — el enum `banda` (7 valores, cerrado) no la contempla y añadirla
   re-decide geografía firmada (mockup v3 verbatim).
2. **Duplicación:** la capa Proceso (UX it.9 firmada) ya asigna hogar visual al binomio
   `entra {estado}·{artefacto} → sale {estado}·{artefacto}`; el Inspector ya muestra el
   detalle. Tercera superficie sin fuente común = drift visual.
3. **Decisión #6 del Gate 1 (HS-09)** ya fijó la visualización: superficie = transición
   del spine por caja · edges = flujo del artefacto · click→Inspector = contrato ·
   instancias VIVAS moviéndose = tablero «Flujo» como vista APARTE telemetry-gated.
   Una franja con instancias vivas la re-litigaría.
4. **Geometría:** los anclajes de edges son horizontales (borde-derecho→borde-izquierdo);
   una banda al fondo produce beziers verticales largos cruzando carriles en un lienzo ya
   denso.

**Forma recomendada (lente UX + abogado del diablo convergen): «artefactos en el spine» —
chips de hand-off intercalados en el gutter entre carriles consecutivos, a la altura del
spine.** Un solo chip = output de la caja N e input de la N+1 (A2 hecho píxel). 100%
proyección FE derivada del contrato (postura P5/HS-12: «se proyecta, jamás descriptor
aparte»): selector puro `selectArtefactos(g)` que sintetiza pseudo-nodos con
`data-node-id` (el overlay de edges los conecta gratis, use-edge-paths.ts ancla por DOM);
edges `escribe` (caja→chip) y `lee` (chip→caja) — el tipo `escribe` está reservado en el
enum y nunca se deriva (edges.go). Reposo sin toggle = mapa idéntico al actual; seleccionar
caja = chips adyacentes iluminados, resto atenuado; toggle «Artefactos» en MapBar (NO 5ª
capa: las capas firmadas son lentes de telemetría). Inputs `de: usuario|terceros` = chip
huérfano-visible (patrón §4.5, jamás omitidos); entregas sin consumidor = chip dead-end
warn — el mapa muestra lo que conformance verifica. Frontera A3: chips muestran TIPOS
declarados (Estructura), jamás status/instancia viva (eso es del tablero Flujo).

**Restricciones duras (unánimes):** artefacto ≠ 11ª `clase` (enum crece solo si CC suma
primitiva; firewall CC-native) · jamás nodo declarado en el L0 ni colección que re-declare
productor/consumidor (anti-segunda-fuente, pecado que P5 rechazó para I-77) · jamás
reutilizar el enum-5 de `spine.categorias` para estados de artefacto (tocaría el contrato
DevStudio sin firma).

## 3. Plantillas + llenado casi determinista (lo genuinamente nuevo, mecanismo completo)

Ubicación según nomenclatura (sin clase nueva):
- **Estándar corto** = rule de banda Base consumida vía `necesita.de: base:<std-id>` —
  patrón `std-spec` ya vivo en el dogfood.
- **Esqueleto largo** = `skills/<caja-escritora>/references/plantilla-<art>.md` + validador
  en `scripts/` de la MISMA skill (progressive disclosure nivel 3: ~0 tokens hasta usarse;
  `escritor_unico` da dueña única). Command descartado (cmd-migrate-to-skill).
- **Prohibido always-on**: plantilla en CLAUDE.md/rule sin scope viola p11
  (rules-size/token-budget). Just-in-time siempre.

División script↔LLM (dial oficial «degrees of freedom» + check
`skill-script-for-deterministic`):
- **Script** (se ejecuta, su código no entra al contexto): scaffold del artefacto
  (frontmatter document-as-cache §8.3: inputs·status·timestamps + secciones vacías + TOC),
  copia de campos entre artefactos (trazas, IDs, fechas), validador `validate_<art>` con
  errores VERBOSOS (habilita autocorrección), destilador de cierre que emite
  `<art>.digest.md` (frontmatter + resumen ≤200 tok + TOC con anchors).
- **LLM**: solo prosa de juicio (why, decisiones, trade-offs, Gherkin). Variante para
  artefactos muy estructurados: plan-validate-execute con `claude -p --json-schema`
  (una property por sección) → renderer determinista compone el markdown. OJO:
  validación post-hoc con re-prompting, no constrained decoding — presupuestar
  `error_max_structured_output_retries` en la FSM.

Gates (un ÚNICO validador, tres puntos de uso — «nada sin eval»):
- **PostToolUse** (matcher Write|Edit sobre el path del art): corre el validador; bloqueo
  real = JSON `{"decision":"block","reason":…}` (en PostToolUse exit 2 NO bloquea).
  Mantener <500ms o mover el scan pesado al Stop.
- **Stop** = gate final de la caja: mismo validador; si falla, block con lista de
  violaciones; `stop_hook_active` + repairCap del conductor como redes.
- **`gate.tipo: auto`** con Gherkin then = «el artefacto valida contra su plantilla» (A4).
  Sin plantilla → `tipo: none`, jamás fabricado.
- **Anti-señal-falsa:** el `status: done` del frontmatter lo estampa el VALIDADOR al
  pasar, no el LLM — cierra el hueco actual donde el conductor confía en un status libre.

Encadenado sin re-alimentar contexto (p11, el argumento del operador):
- El filesystem del cwd confinado es la memoria compartida entre cajas. Al spawn de la
  caja N+1 el conductor: (a) verifica existencia + `status=done` de cada `necesita`
  requerido ANTES de Spawn (precondición determinista nueva — semántica case-handling: el
  dato disponible habilita, el spine sigue mandando); (b) inyecta en `tarea()` solo RUTAS
  + digest destilado, jamás el doc entero; (c) la skill declara qué secciones lee (TOC →
  lecturas parciales). Patrones harness-profile: Temp File Assembly · Delegated Data
  Access (padre no lee lo delegado). Reducción documentada hasta 98.7% (Anthropic,
  code-execution-with-mcp).

**Alcance de la obligatoriedad:** plantilla + document-as-cache estricto SOLO para
arquetipos `pipeline`/`excepcion` (precedencia arquetipo>perfil); las cajas `abierto`
quedan exentas (§8.1/§8.3) — universalizarlo violaría p1 (burocracia).

¿Bastan los 3 ejes del contrato? **Sí, con extensión aditiva dentro de CABLEADO:**
`entrega[].plantilla?: <ruta>` y `entrega[].path?:` (distinguir artefacto-archivo de
artefacto-etiqueta). No hace falta 4º eje. Bump aditivo de box.contract.schema.json
(additionalProperties:false — patrón reparación P3/HS-12) + espejo `domain.Output`.

## 4. Respaldo externo (resumen)

- **Artifact-centric BPM** (Nigam & Caswell 2003, IBM/BALSA; GSM/Hull 2011; CMMN;
  case handling van der Aalst 2005): 20+ años validan artefactos como entidades modeladas
  con information model + lifecycle. La mejor práctica es EXACTAMENTE el híbrido
  recomendado: contrato de la actividad como única fuente (BPMN ioSpecification) + vista
  proyectada; el riesgo canónico es la doble fuente de verdad proceso↔artefacto y la
  mitigación es derivación/proyección + el proceso REACCIONA al artefacto por eventos
  (sentries CMMN ≈ nuestro gate Gherkin), jamás copia su estado.
- **El manifiesto Agentic BPM (arXiv 2603.18916) NO trata artefactos** — hueco que
  ArnesIA puede llenar como extensión original alineada con framed autonomy (plantilla +
  validación = guard-rails).
- **Spec-driven development 2025-26** (spec-kit, BMAD, OpenSpec, Kiro, Agent-OS): todos
  encadenan documentos con plantillas-programa (elicit forzado, marcadores
  [NEEDS CLARIFICATION], checklists ejecutables como «unit tests del documento»), gates
  por fase y paso de contexto selectivo (sharding a story files autocontenidos; tasks lee
  plan.md, no la spec; delta specs de OpenSpec que COMPONEN al archivar). Falla #1
  reportada: **spec drift** («the spec is fiction») y docs zombis — la franja necesitará
  a futuro un mecanismo de reconciliación artefacto↔realidad (drift gate); declararlo
  deuda explícita.
- **Prácticas oficiales CC**: Template pattern en skills; scripts > prosa para lo frágil;
  plan-validate-execute con intermedio verificable; PostToolUse decision:block / Stop
  gate; `--json-schema` structured output; context engineering (rutas + digests, no docs
  enteros).

## 5. Riesgos / líneas rojas

1. **Identidad de `art` es EL prerequisito**: hoy string libre con semántica doble —
   etiqueta en el contrato, path en el conductor (artifactRef = `Entrega[0].Art`). 3 de 4
   cajas del dogfood tienen art no-path («código + tests», «veredicto de review»,
   «release@version») → `Status` devuelve vacío y **el error se descarta en silencio**
   (box_conductor.go:90). Sin resolverlo, la proyección pinta datos incoherentes y el
   mecanismo cojea.
2. **Doble fuente de verdad**: nada re-declara productor/consumidor/plantillas fuera de
   `necesita`/`entrega`. Los artefactos NO ganan ciclo de vida propio (borrador→aprobado
   duplicaría el spine; si algún día se necesita, eje aparte y jamás el enum-5).
3. **Geografía firmada**: mockup→firma ANTES de código (map-canvas es port verbatim del
   v3); comparar chip-en-gutter vs franja-abajo en el MISMO mockup y llevar ambos al gate
   contra A6 y la decisión #6.
4. **Densidad**: fan-out / entregas múltiples / cajas paralelas apilan chips — el mockup
   fija tope + overflow (+N); cajas paralelas pueden exigir overlay medido por DOM en vez
   de columna flex.
5. **Costo de oportunidad** (abogado del diablo): 3 capas del Mapa vacías esperando
   telemetría, loader que reconoce 2 de 10 clases (banda Guardia impoblable desde disco),
   208 checks deferred. Mitigación: la fase 1 del plan (checks de composición) ES pagar
   deuda prometida, no fachada nueva.

## 6. Impacto DevStudio (respuesta directa: NO hay que actualizar el contrato hoy)

- marketplace.json y catalogo.json: intactos. Forma-plugin: solo suma archivos
  (references/ de plantilla) + campos opcionales en versiones NUEVAS publicadas (jamás
  mutar 0.1.0 in place — patrón P3). El compromiso HS-12 «solo aditivos» lo ampara.
- `spine.categorias` enum-5 (I-77 RN-28): intacto. Línea roja: no reutilizarlo para
  estados de artefacto.
- Postura P5 sale REFORZADA: I-77 se proyecta de spine + contratos de caja, y los
  artefactos ya viven en esos contratos.
- Timing ideal: el publisher fase-5 es stub (`errNotImplemented`) — si la identidad del
  art se firma antes del publish real, el formato nace con el campo, sin migración.
- Cortesía: ficha gemela (patrón DH-18/PB-25) cuando se firme `entrega[].plantilla`/`path`.
- Nomenclatura: v1.2 (bump menor aditivo, patrón v1.1) SOLO si el mockup firmado exige
  plantillas reconocibles en disco — sin reconocedor, cada plantilla suelta generaría
  nodo `no-reconocido` warn; dentro de `skills/<id>/references/` no necesita nada.

## 7. Plan propuesto (fases, cada una con su gate)

1. **Motor — checks de composición** (cero decisión nueva; deuda prometida en
   METODOLOGIA §3/§6, UX it.10, boundary contrato-de-caja-es-fitness-function):
   implementar `sin-huerfanos` (declarado, deferred), `dead-end` (fila nueva),
   `ruta-a-existe` y `art-identidad-coherente` — gemelos de `VerificarEscritorUnico`,
   puro dato del grafo — y cablearlos a la ruta `--arnes`.
2. **Identidad del art**: bump aditivo box.contract.schema.json (`entrega[].path` y
   `entrega[].plantilla` opcionales) + espejo domain + el conductor deja de descartar el
   error de Status; check warn `art-es-path` para cajas pipeline/excepcion.
3. **Plantillas dogfood-first (HS-07)**: spec-writer gana `references/plantilla-spec.md`
   + `scripts/validate_spec` + hooks PostToolUse/Stop + `gate.tipo:auto` Gherkin; el kit
   `forjar-caja` materializa plantilla+validador al forjar cajas pipeline/excepcion.
   **MEDIR tokens antes/después del digest (p11 — el argumento del operador).**
4. **Conductor**: precondición `necesita[]` antes de Spawn + destilado determinista al
   cierre + `tarea()` inyecta rutas+digest.
5. **Mockup de la proyección visual** (mockups/arnesia-mapa-mvp.html, tokens DTCG
   reales, datos del dogfood/showcase): chips de hand-off en el gutter a la altura del
   spine vs franja-abajo, estados reposo/selección/dim/huérfano/dead-end, fan-out y cajas
   paralelas → **gate humano** (re-decide geografía firmada).
6. **FE tras firma**: `selectArtefactos` (selector puro) + HandoffGutter/ArtefactoChip
   (`data-node-id`) + re-ruteo visual invoca→escribe+lee bajo toggle + navegación
   chip↔Inspector; spec.md/design.md trazados + story=test + PARIDAD.md.
7. **Nomenclatura v1.2 + ficha DevStudio** solo si aplica (ver §6).

Las fases 1–4 (MECANISMO) son independientes de la 5–6 (GEOGRAFÍA) y firmables ya.

## 8. Auditoría colateral (hallazgos verificados, ejecutados en vivo)

Aritmética que CUADRA ✓: knowledge 12 nodos/138 checks · arch 16 boundaries/97 checks
(71+26) · 235 checks en `--todo` · 8 boundaries enforced · round-trip dogfood
`index → conformance --arnes` = 15/15 PASS · `go test ./internal/...` verde.

Discrepancias:
1. **CLAUDE.md stale**: «24 pass real + 211 deferred» → hoy **27 pass / 208 deferred**
   (3 arch-tests flipados a verde tras escribirse la cifra). CLAUDE.md:84,126.
2. **Validación de composición prometida en 4 cuerpos, implementada en 0** (ver fase 1).
   Desglose real de los 208 deferred: 185 nl-judge legítimos · 3 go-arch-lint (vivo en
   CI) · 5 que sí corren en `--arnes` · **15 arch-test sin enforcer real** (3 stubs
   `t.Skip` + 12 con enforcer genérico sin función — entre ellos `sin-huerfanos` y
   `gate-honesto`, este último severidad error).
3. **Error de `artifacts.Status` descartado en silencio** (box_conductor.go:90) +
   semántica doble del art (§5.1).
4. **Colisión de elemento «hooks»**: el parser del ruleset usa basename
   (parser.go:105) — knowledge/elements/hooks.md (12 checks CC) y
   arch/conventions/hooks.md (3 checks lefthook) se funden en un elemento de 15;
   las cuentas 138/97 solo cuadran partiendo a mano.
5. **Drift fixture↔loader en el mismo dogfood**: el fixture embebido estampa
   `canal:beta` que el loader jamás estampa; std-spec sale del loader con
   `procedencia:declarado` que el fixture no tiene → el mismo arnés se ve distinto
   seedeado que cargado por «Cargar carpeta». El inspector pinta `canal`
   (inspector.tsx:412).
6. **Campos sin semántica en el motor**: `necesita[].requerido` (solo lo pinta el FE) y
   `ruta[].a` (nunca validado contra nodos existentes).
7. El Mapa por defecto muestra grafos que el loader real no puede producir
   (content-studio-full ejercita 10 clases/7 bandas; loader reconoce skills+CLAUDE.md,
   emite solo fase/base, edges lee/invoca; prefijos libreria:/maquinaria:/terceros:/
   marcas-dormidas: no generan edge).
