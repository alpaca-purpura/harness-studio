# SPEC — Descriptor de proceso (v1) — CONGELADA

> **Congelada 2026-07-03** (épica Torre de Control, F4 · fichas `I-77` + `DH-07`). La
> implementación ejecuta ESTO sin retocar diseño; cambio de diseño = nueva ficha DH-NN +
> edición explícita aquí. Continúa la numeración de la épica (RN-01..RN-26 en
> `torre-read-only.md` y `torre-arquitectura.md`). Decisiones D8-D10: checkpoint expirado →
> recomendadas aplicadas (mecánica de la casa), registradas en
> `../epicas/torre-de-control/ESTADO.md`, `DH-07` e `I-77`.

## 0 · Qué es

El proceso de la empresa como DATO (I-77, espejo de I-72): un **descriptor** declara
estados·categorías·transiciones·gates·dueños; el **kit lo shipea** (instancia por defecto),
**DevHub lo interpreta y renderiza** (motor + consola), el **schema es L0**. El ciclo SDD de
10 estados deja de ser hardcode en `gates.go` y pasa a ser la instancia 1. "Otra empresa,
otro descriptor, cero cambio de consola" = la prueba del corte.

## 1 · Contrato de 3 planos (I-77)

- **RN-27** · Tres dueños, un contrato: **schema L0**
  `.claude/harness/schema/process-descriptor.schema.yaml` (categorías fijas del contrato +
  shape de nodos — hermano de `harness-descriptor` en el eje proceso; NO confundir con
  `process.schema.yaml`, que es la capa ENGAGEMENT de service — I-20) · **instancia SSoT en
  el kit** `products/kit/core-harness/process/sdd-default.yaml` (la superficie `process/`
  del plugin pasa de prosa a dato; viaja y se versiona con el kit — CHANGELOG KIT-08) ·
  **DevHub consume un espejo GENERADO+gated** (RN-30) — patrón copia-propia-sin-import
  (D7/RN-21 + disciplina CK-08), cero lectura runtime de paths ajenos.

## 2 · Schema (L0)

- **RN-28** · Nodos del descriptor:
  - `descriptor:` `{id, nombre, version (int, de la instancia), descripcion?}`.
  - `categorias` = **enum FIJO del schema** (regla Azure/D4, inmutable desde v1):
    `propuesto · en-progreso · completado · descartado · pausado`. Todo comportamiento
    genérico del motor/consola se ata a la categoría; el `id` nativo del estado se preserva
    al lado (DevLake `status`+`original_status`). Terminalidad = **derivada**: categoría ∈
    `{completado, descartado}`.
  - `estados[]` `{id (nativo, inmutable), categoria (enum), nombre?, descripcion?,
    inicial? (bool — exactamente UNO true)}`. El orden del array ES el orden canónico de
    render/board.
  - `transiciones[]` `{de, a, ejecutor: operador|rol, requiere_razon? (bool, solo
    operador), verbo? (str — RESERVADO para F6/CDEvents, el schema no lo exige ni lo
    interpreta)}`. Las de `ejecutor: rol` son declarativas en v1 (documentan el ciclo, se
    renderizan; el motor NO las hace cumplir — hoy los roles transicionan vía skill
    escribiendo checkpoint.md, y eso no cambia en F4).
  - `gates[]` `{id, nombre, momento (estado o transición "de→a" al que aplica), autoridad
    (binding), checklist[] (≥1 item — Essence: el gate ES su checklist)}`.
  - `duenos:` mapa `estado → [{rol, arnes, nota?}]` — **bindings nombrados** (principio 2,
    ASL/XState/`ruleKey`): matan el hardcode `stateOwners`. Solo estados cuya LLEGADA
    ejecuta un rol (no el operador).
  - `parametros:` `{razon_minima (int >0)}` — hoy `transitionReasonMinLen=10`.

## 3 · Instancia 1 — `sdd-default` (round-trip contra gates.go)

- **RN-29** · Los 10 estados con su categoría: `idea→propuesto` ·
  `refining/refined/ready/developing/developed/reviewing→en-progreso` · `done→completado` ·
  `dropped→descartado` · `parked→pausado`. `inicial: idea`. Transiciones de operador =
  EXACTAMENTE las de `operatorAllowed` (idea→refining · idea→parked⚡ · idea→dropped⚡ ·
  refining→idea · refining→parked⚡ · refining→dropped⚡ · parked→idea; ⚡ =
  `requiere_razon`). Dueños = EXACTAMENTE los de `stateOwners` como bindings (ej. `done:
  [{rol: pm, arnes: "/pm-{sistema}", nota: "Fase F MERGE"}]`; `refined` lleva 2 bindings —
  architect y pm). Transiciones de rol completan el ciclo feliz
  (refining→refined→ready→developing→developed→reviewing→done). Gates v1 = los que el motor
  HACE CUMPLIR hoy, como checklist: `razon-de-cierre` (operador, parked/dropped),
  `verificacion-operador` ("gate G" de handlers_verify — aplica en `developed`),
  `merge-gate` (release: toda story en categoría terminal). `parametros.razon_minima: 10`.
  **Round-trip**: las tablas derivadas == las literales pre-F4; los tests Go existentes
  (gates_test.go y toda la suite) pasan SIN tocarse.

## 4 · Espejo gated + motor

- **RN-30** · `tooling/scripts/gen_proceso_descriptor.py` (fábrica al servicio del contrato
  de ecosistema — I-77, no aplica el atajo RN-12): **valida** la instancia del kit contra el
  schema L0 (ids únicos · categoria ∈ enum · exactamente 1 inicial · `de`/`a`/`momento`/
  `duenos` joinean estados · `requiere_razon` solo en ejecutor operador · checklist ≥1 ·
  razon_minima >0) y **emite** `products/devhub/go/process/sdd-default.yaml` (cabecera
  GENERADO). Registrado en `gen_all.py` STEPS + el espejo en `repo-map.yaml::generated`
  (drift = gate rojo). Validación falla → exit 1, no emite (SC-30).
- **RN-31** · El motor deriva, la API del dominio NO cambia: `gates.go` conserva
  `isStoryState · evaluateOperatorTransition · transitionOwnerNote · isTerminalState` con
  las mismas firmas; sus tablas (`storyStates · operatorAllowed · stateOwners`) se
  inicializan desde el descriptor embebido (`go:embed` del espejo RN-30, parse en
  `proceso.go`). Terminalidad = categoría ∈ {completado, descartado}. Nota de owner = join
  de bindings `"{arnes} ({nota})"` con `" o "` (reproduce byte-a-byte las strings
  actuales). Descriptor embebido ilegible/inválido → **panic al boot** (fail-fast: un motor
  sin proceso válido no arranca; jamás un proceso a medias).

## 5 · API + vista (D10: render mínimo, board intacto)

- **RN-32** · `GET /api/proceso` → el descriptor completo + derivados:
  `{descriptor: {id, nombre, version, descripcion}, categorias[], estados[] (con
  `terminal` derivado), transiciones[], gates[], duenos{}, parametros{}, fuente}`.
  Sin `?sistema=` en v1: el binario tiene UN descriptor embebido (el default del kit);
  descriptor por workspace = extensión F5+.
- **RN-33** · Vista nueva `/proceso` (nav item propio): rinde TODO desde `/api/proceso` —
  pipeline de estados en su orden canónico con **categoría como badge + id nativo al lado**
  (D4), transiciones agrupadas por ejecutor (operador vs rol), gates con su checklist
  desplegada, dueños como bindings (rol + arnés + nota). Cero literal de estado en el
  componente nuevo: si el descriptor cambia, la vista cambia sola. La prueba "otro
  descriptor, cero cambio de consola" = test Go con un descriptor alterno de fixture
  (estados inventados) → derivaciones correctas sin tocar código del motor.
- El board y las 8 definiciones duplicadas de la UI (STATES_ORDER · STATE_CLASSES ·
  STATE_OWNERS ×2 · OPERATOR_ALLOWED_TRANSITIONS · tooltips · WIP_CAPS · edit-permissions)
  NO se tocan en F4 — migran a `/api/proceso` en F5/F6 con el contrato ya probado (D10).

## 6 · Fuera de alcance (v1)

Board consumiendo el descriptor (F5/F6) · verbos CDEvents (F6 — campo reservado) ·
descriptor por workspace/override runtime (F5+) · enforcement de transiciones de rol (los
skills siguen siendo los ejecutores) · políticas de edición de campos de story/capability
(política de ENTIDAD, no de proceso) · editor/escritura del descriptor (la consola es
read-only sobre él) · estados de ticket (capa de abajo — `ticket-states.md` sigue prosa).
