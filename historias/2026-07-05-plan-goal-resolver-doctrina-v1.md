# PROMPT `/goal` — Resolver la doctrina v1 hasta hacerla ejecutable (HS-08)

> **Cómo se usa:** este archivo ES el argumento del comando `/goal`. Pégalo (o referéncialo) tras
> `/goal`. El objetivo es un LOOP: `/goal` itera ola por ola, cada una cierra con un gate verde, y el
> goal **no termina** hasta que la auditoría final pasa. Fuente de verdad de los hallazgos =
> [`historias/2026-07-05-auditoria-doctrina-v1-aplicabilidad.md`](./2026-07-05-auditoria-doctrina-v1-aplicabilidad.md)
> (referéncialo con `BNN`/`MNN`/`mNN`).

---

## OBJETIVO (definition of done)

Llevar la doctrina propia v1 de **prosa referenciada** a **spec ejecutable e invocable**, de modo que
un autor pueda forjar el arnés `dev-full-cycle` real y validarlo verde end-to-end. **El goal termina
SOLO cuando los 4 gates finales pasan:**

1. **G1 · Contrato verde:** un `contract:` de caja real (el primero del dev-full-cycle) valida contra
   el schema fusionado y pasa `arnesia conformance` sin errores — con TODOS sus ejes (intención +
   clasificación + cableado + aceptación) reconocidos, no ignorados.
2. **G2 · Knowledge invocable:** los checks del árbol `knowledge/` se **ejecutan** (no solo se leen)
   vía el puerto de conformidad; correr un elemento devuelve un `ConformanceReport` con veredictos
   reales por check, no markdown.
3. **G3 · Sin punteros colgantes:** cero `enforced_by:` apuntando a tests inexistentes; cero cita
   fantasma; cero conteo/roadmap stale entre VISION/METODOLOGIA/CLAUDE/LEDGER/INDEX.
4. **G4 · Auditoría final limpia:** re-correr la auditoría 5-frentes sobre el estado resuelto → cero
   hallazgos de severidad alta abiertos; los diferidos quedan marcados con honestidad (`status:` +
   ficha responsable), no como huecos silenciosos.

---

## PRINCIPIOS RECTORES (no negociables)

1. **Hexagonal / ports & adapters como columna.** El core del dominio (modelo de la doctrina: `Box`,
   `Contract` fusionado, `Spine` **como TIPO/forma** —no valores—, `Arnes` manifiesto, `UnidadDeTrabajo`
   agnóstica, `ConformanceReport`) **no conoce** ningún detalle de CC, de formato de check, ni de
   transporte. Todo lo que **cambia rápido** vive detrás de un puerto, como **datos + adaptador
   intercambiable**. Regla dura: si algo cambia cada semana (primitivas de CC, los 138 checks, el
   estándar por elemento) o cambia **por-arnés** (fases, estados del spine, proceso), **no puede vivir
   como valor en el core** — va como dato tras puerto. Esto es lo que da alta escalabilidad y
   modificabilidad.
2. **El conocimiento es DATO, no código.** El árbol `knowledge/` (12 nodos · 138 checks) y `arch/`
   (16 boundaries · 97 checks) se cargan como **ruleset parseado** (frontmatter + tabla → estructura),
   no se hardcodean. Agregar/cambiar un check = editar datos, jamás tocar el motor. La cadencia semanal
   de knowledge sigue viva sin recompilar el core.
3. **Investigación POR PUNTO antes de corregir.** Cada ítem arranca con un subagente de research
   (verifica el hallazgo contra el código real + busca el estándar vigente de CC/industria si aplica)
   y produce un mini-diseño ANTES de escribir. Nada se corrige a ciegas.
4. **Subagentes para todo lo paralelizable.** Research, implementación aislada por ítem, y verificación
   adversarial corren como subagentes. El hilo principal orquesta y sintetiza (patrón conductor: no
   lee lo que delega, lo enruta por nombre).
5. **Honestidad as-code (principios 6 y 10 de VISION).** Jamás fabricar un eval/check para rellenar:
   `gate: none`/`status: proposed`/`diferido` con el PORQUÉ visible es válido; un hueco disfrazado no.
   `enforced_by:` solo apunta a algo que existe (aunque sea `t.Skip` legible).
6. **Aditivo y sin pérdida (principio 4).** No se reescribe la doctrina firmada; se la baja a schema y
   se reconcilian los stale. Lo viejo se marca `deprecado`, no se borra silencioso. Cada cambio
   estructural = ficha `HS-NN` + (si crea/cambia boundary) nodo en `arch/` con su check.
7. **Cada puerto nuevo NACE como boundary.** `ConformancePort`, `RulesetPort`, `PermissionPort`,
   `ConductorPort` se declaran en `arch/boundaries/` con su check `enforced_by:` real (o `t.Skip`
   honesto), coherentes con los boundaries previos (`adaptadores-de-agente-intercambiables`,
   `dominio-independiente-de-transporte`, `conductor-no-parsea-jsonl`, `contrato-de-caja-es-fitness-function`).
8. **Agnóstico al proceso y a los estados (principios 3 y 7 de VISION — corrección clave).** ArnesIA
   CREA arneses para roles/procesos arbitrarios; **NO** tiene un spine ni un juego de fases universal.
   `fase` (agrupador de cajas) y `estado` (transición del work-item) son **DATOS que cada arnés declara**,
   nunca constantes del producto — el as-code ya lo respeta (`domain.Fase`/`Estado` son `string` sin
   `const`/`Valid()`, a diferencia de `Clase`/`Banda`/`GateTipo`). Regla de bolsillo: **enumerables como
   constantes del producto = solo ejes de doctrina** (`arquetipo`, `perfil_harness`, `clase`, `gate.tipo`);
   `fase`/`estado` NO. El producto solo provee el **TIPO** (forma del spine) + la **validación de
   consistencia** contra lo que el arnés declaró. El work-item se nombra agnóstico (`UnidadDeTrabajo`/
   orden-de-trabajo), **jamás `Story`**. Los "6 fases / ~7 estados" de luana son **fixture de un arnés
   ejemplo**, no modelo del producto.

---

## EL PUNTO PILAR (nuevo, elevado a primera clase)

**P0 · El estándar `knowledge/` está mapeado pero NO cableado a ningún motor que lo invoque** (hallazgo
B5 + huérfano m2). Hoy los 138 checks son rúbricas markdown sin `enforced_by:`; `arnesia conformance`
no existe; el runner unificado quedó huérfano. **La doctrina cita el conocimiento, no lo enchufa.**

Resolverlo es el corazón del plan y del diseño hexagonal:

- **`RulesetPort`** (nuevo, salida): carga `knowledge/` + `arch/` como un **ruleset de datos** (cada
  check = `{id, elemento, severidad, mecanismo, enforced_by, señal}`). Adaptador = parser de
  frontmatter+tabla. Cambiar el conocimiento = cambiar datos que este puerto relee.
- **`ConformancePort`** (nuevo, entrada del caso de uso): "corre el ruleset del elemento X contra el
  target Y → `ConformanceReport`". El core orquesta; no sabe CÓMO corre cada check.
- **Adaptadores de mecanismo** (intercambiables, uno por tipo de check): `schema-validation`
  (jsonschema-go), `go-arch-lint`, `arch_test`, `NL-judge` (checks de juicio/telemetría), `static-scan`.
  Cada check declara su `mecanismo` → el runner rutea al adaptador. Agregar un mecanismo nuevo = un
  adaptador nuevo, core intacto.
- **`enforced_by:` en los nodos de knowledge**, empezando por el subconjunto que el dogfood necesita
  (contrato, **consistencia-de-estado contra el spine DECLARADO del arnés**, huérfanos/dead-ends,
  firewall). El resto se marca `mecanismo: nl-judge` / `diferido` con honestidad.

Salida: `arnesia conformance <elemento|arnés>` corre checks reales y emite veredictos. **G2 depende de
esto.**

---

## MÉTODO POR PUNTO (loop reutilizable — aplícalo a CADA ítem de cada ola)

```
1. RESEARCH (subagente):     verifica el hallazgo en el código real + estándar vigente CC/industria
                             → devuelve <mini-diseño hexagonal → path> (qué core, qué puerto, qué adaptador)
2. DISEÑO (hilo principal):  aprueba/ajusta el mini-diseño; confirma que el core no toca detalle volátil
3. IMPLEMENTA (subagente):   aislado por ítem (isolation: worktree si toca archivos compartidos);
                             schema/dominio/adaptador + test que lo prueba
4. VERIFICA (subagente):     adversarial — intenta romper el fix; corre `go test ./...` + el check nuevo;
                             confirma que un caso inválido FALLA y uno válido PASA
5. GATE:                     el ítem cierra solo si su verificación es verde y no rompió boundaries previos
```

---

## PLAN EN OLAS (orden = dependencia dura; cada ola es un gate)

### OLA 0 · Fundamentos de doctrina + schema (baratos, desbloquean todo)
- **0.1** **Meta-modelo de spine/fases per-arnés** (B2 REFORMULADO — NO enum universal): el dominio da
  el **TIPO** `Spine` (forma = estados + transiciones legales) y la `UnidadDeTrabajo` agnóstica (jamás
  `Story`); **cero estados/fases horneados en el core ni en el schema de caja**. Afirmar A3 —ya acordado—:
  `fase` (agrupador de cajas) ⊥ `estado` (transición del work-item) son ejes distintos; **ojo, es una
  ortogonalidad DISTINTA de la de A3-original** (estado-del-trabajo vs estado-operativo-de-caja/telemetría).
  Los valores concretos (`idea→…→released`, 6 fases) viven como **fixture del arnés dogfood** (Ola 5.2).
  *Puerto: dato del manifiesto (0.5), no core.*
- **0.5** **Schema del manifiesto del ARNÉS** (adelantado de la vieja 4.2 — hogar correcto del spine, dato
  per-arnés): en el bloque `arnes` (`graph.l0` / `domain.graph`) declarar `fases: [...]`,
  `spine: {inicial, terminales, estados: [...]}` **+** la META de enganche `rol·proceso·reporta-a·empresa`
  (M3; agregar `proceso`, decidir `puesto`↔`rol`, marcar required). Todo son DATOS del arnés. **Debe
  existir ANTES de la Ola 1** (el check de consistencia de estado se parametriza contra este manifiesto).
- **0.2** Reescribir **`box.contract.schema.json` + `domain.Contract`** al contrato fusionado (B1): los 3
  ejes (intención/clasificación/cableado/aceptación), enums `arquetipo|perfil_harness|clase|gate.tipo`,
  `escritor_unico`, `handoff`, `gate.aceptacion` Gherkin. Activar `TestBoxContractValidatesAgainstSchema`.
  **NO añadir enum de `estado`/`fase`** — quedan patrón `X -> Y` / string libre (agnóstico); su validación
  semántica es el check per-arnés de la Ola 1 contra el manifiesto 0.5, no el schema de caja.
- **0.3** Fijar la **ontología única de `clase`** (B3): un solo enum alineado a los 12 nodos (labels
  canónicos `subagent`/`rule`/…), separado del eje banda/rol del mapa si hace falta un 2º campo; darle
  check. Actualizar `graph.l0.schema.json` (quitar el enum legacy de 7). *(Recordar las 3 capas de
  tipo, sin confundirlas: **QUÉ es** = `clase` + caja/apoyo · **QUÉ forma** = `arquetipo` · **CÓMO corre**
  = `perfil_harness` [T2 = Workflow]. "Workflow" NO se perdió; es T2.)*
- **0.4** Doctrina pura: resolver **`abierto×T2`** de document-as-cache (B8, heredar `[PENDIENTE 7.A]` o
  precedencia arquetipo>perfil); **re-atribuir A2/A7** a DAOP/BMAD/12-Factor (B7); **desambiguar el
  sistema A#** (B9, renombrar los axiomas importados; purgar "A12" de knowledge).
- **Gate 0:** el schema fusionado valida un contrato de ejemplo con los 3 ejes; el dominio tiene el **TIPO**
  `Spine` y el manifiesto del arnés declara SU spine/fases (0.5), **sin enum de estado/fase en el schema
  de caja**; `clase` tiene un solo significado; cero contradicción `abierto×T2`; cero cita mal atribuida.

### OLA 1 · Cableado ejecutable — P0 (el pilar)
- **1.1** Diseñar el **contrato de check común** (m2): schema del check `{id, elemento, mecanismo,
  enforced_by, severidad, señal}` que knowledge y arch comparten.
- **1.2** Construir **`RulesetPort` + adaptador de carga** (parser de los .md → ruleset de datos).
- **1.3** Construir **`ConformancePort` + los adaptadores de mecanismo** (schema-validation, go-arch-lint,
  arch_test, nl-judge, static-scan) — hexagonal estricto.
- **1.4** Implementar **`arnesia conformance <target>`** (CLI mínimo) que corre el subconjunto
  fundacional y emite `ConformanceReport`.
- **1.5** Poblar **`enforced_by:`** en los nodos de knowledge del subconjunto dogfood; marcar el resto
  `nl-judge`/`diferido` con honestidad.
- **1.6** **Checks de consistencia de estado PARAMETRIZADOS por el spine declarado del arnés (0.5)** —
  NO contra enum fijo: `estado-en-spine-declarado` (X,Y ∈ `arnes.spine.estados`), `transicion-legal`
  (X→Y ∈ transiciones), `una-transicion-por-caja`, `spine-cobertura` (sin estados huérfanos/inalcanzables
  desde `inicial`), `fase-en-fases-declaradas` (`nodo.fase ∈ arnes.fases`; cada fase cubierta por ≥1 caja).
  Convierte la promesa de METODOLOGIA §6 ("las transiciones encadenan, sin huérfanos") en check ejecutable
  **sin conocer ningún estado concreto** — solo consistencia contra lo declarado. Carga el spine como dato
  vía `RulesetPort`.
- **Gate 1 (= G2):** `arnesia conformance` corre y devuelve veredictos reales por check para al menos
  skills/subagents/hooks/contrato **+ los checks de consistencia de estado contra el spine declarado**.

### OLA 2 · Motor de ejecución (código, tras puertos)
- **2.1** Construir el **loop del conductor T3** (B6) tras `ConductorPort`: spawn iterado (`--max-turns`)
  + cap de reparación + lectura de `result`+`status` del artefacto + `blocked→handoff` + ejecución de
  `contract.ruta`. Modelar la FSM interna `draft→…→blocked` en el dominio. Crear
  **`TestConductorOwnsBoxRouting`** (B4) — sube `orquestacion-determinista-entre-cajas` de proposed a enforced.
- **2.2** **Spike `control_request` + `KitProvisioner`** (M4) tras `PermissionPort`: resolver
  permission-set por rol + TTL/por-tarea; endpoint `control_request` con `role`/`ttl`. (El `proceso`/`rol`
  required en `graph.l0` ya lo aterrizó la Ola 0.5.) Crear **`TestPermissionSetParametrizedByRole`** (B4)
  — sube `permisos-derivan-del-rol` a enforced.
- **Gate 2:** una caja `perfil_harness: T3` corre su loop real; un permiso se concede por rol y expira;
  cero enforcer fantasma.

### OLA 3 · Operacionalizar la doctrina blanda (todo pasa a check ejecutable o diferido honesto)
- **3.1** **Árbol de decisión** `arquetipo`/`perfil_harness` + regla operable "cuándo NO arnesar" +
  check `no-arnesar-clasificado` (M7).
- **3.2** **Explainability** con campo en el contrato + mecanismo (dónde vive el rationale) + reconciliar
  "4+1"→enumeración estable; `conversational actionability` con L2/check; "framing mechanism" con nº de
  componentes estable (M8/M10).
- **3.3** **Gate de fidelidad** (§8.4) con procedimiento + check; **document-as-cache reread** enforced
  por hook/convención de conductor (M11).
- **3.4** Aterrizar **checks huérfanos** (M12): `single-writer`/`escritor_unico`, `mece-routing-coverage`,
  `snapshot-inmutable`, `anotacion-humana-a-regresion`, `persona-hard-constraint-hook`.
- **3.5** **skills.md** absorbe arquetipo/perfil + check de contrato objetivamente verificable (M5);
  **firewall** con las 4 claves + allowlist (m4); **6 patrones de subagente** detallados con
  cuándo/cómo (M6); declarar los solapes de check (m5).
- **Gate 3:** cada capacidad/gate de la doctrina o tiene check corrible o está `diferido` con ficha.

### OLA 4 · Sincronización de docs
- **4.1** **Barrido VISION + METODOLOGIA** (M1/M2/m1/m3/m9): tabla del gran plan (77→97, HS-06→HS-08),
  "MVP=Mapa primero"→dogfood-first, 121→138, desglose de boundaries 11+5, bumps `version/updated` en
  skills/rules/subagents, ficha huérfana del runner reasignada, "122 checks" stale en arch. **+ purgar el
  smell "spine de 10 estados" (METODOLOGIA §227)** → "spine de estados del arnés" sin número, marcado como
  ejemplo luana (agnosticismo, hallazgo del re-audit).
- **4.2** Reconciliar la doctrina con el manifiesto del arnés de la Ola 0.5: METODOLOGIA §3 debe apuntar
  que `spine`/`fases`/META viven en el **manifiesto del ARNÉS**, no en el contrato de caja; hacer
  ejecutable `meta-de-enganche-completa` contra el schema 0.5.
- **Gate 4:** cero número/roadmap stale cross-doc; cero conteo de estados horneado en prosa; la doctrina
  apunta correctamente al manifiesto del arnés.

### OLA 5 · AUDITORÍA FINAL (corroborar que todo funciona)
- **5.1** Re-lanzar la **auditoría 5-frentes** (coherencia · knowledge · arch · aplicabilidad · fidelidad)
  como subagentes sobre el estado resuelto.
- **5.2** El **arnés dogfood dev-full-cycle declara SU spine y SUS fases** (fixture per-arnés — aquí viven
  los `idea→…→released` / 6 fases de ejemplo, NO en el core). Correr **`arnesia conformance`** sobre su
  **primer contrato de caja real** → debe validar verde con los 3 ejes **+ consistencia contra el spine
  declarado del arnés** (G1).
- **5.3** Verificar en web **arXiv 2603.18916** + quotes antes de dejar el posicionamiento APM en pie (M9).
- **5.4** Confirmar **G1–G4**. Si algún gate falla → el goal NO termina: vuelve a la ola del hallazgo
  abierto y re-itera.

---

## ORQUESTACIÓN DE SUBAGENTES

- **Research (por ítem):** 1 subagente lector — verifica el hallazgo en código + estándar vigente →
  mini-diseño hexagonal. Paralelizar los ítems independientes de una misma ola.
- **Implementación:** 1 subagente por ítem, `isolation: worktree` si toca schema/dominio compartido
  (evita colisión entre ítems paralelos de la misma ola).
- **Verificación adversarial:** 1 subagente que intenta ROMPER el fix (caso inválido debe fallar, válido
  pasar) + corre `go test ./...`. Nada cierra sin este paso.
- **Síntesis:** el hilo principal (conductor) enruta, no lee lo que delega; cierra cada gate y decide la
  siguiente ola.

---

## GUARDARRAÍLES / INVARIANTES (revisar en cada gate)

- No romper boundaries `enforced` existentes (HS-06: superficie-local-confinada, sesion-viva-consistente;
  permisos-gui). `go test ./arch/fitness/...` sigue verde.
- `core` no importa `shell`; `domain` no importa transporte; el conductor no parsea JSONL (boundaries vigentes).
- Aditivo: lo firmado no se reescribe, se baja y reconcilia; deprecados marcados.
- Honestidad: cero eval/check fabricado; cero `enforced_by` colgante; diferidos con PORQUÉ + ficha.
- Cada cambio estructural = ficha `HS-NN` en LEDGER + nodo/boundary en `arch/` con su check.

---

## ITERACIÓN (`/goal`)

`/goal` recorre las olas 0→5 en orden. Cada ola cierra con su gate; si el gate falla, itera dentro de la
ola. Tras la ola 5, evalúa **G1–G4**: si alguno falla, reabre la ola del hallazgo y repite. **El goal
solo se declara resuelto cuando G1–G4 pasan simultáneamente** — un contrato de caja dogfood real valida
verde, los checks de knowledge se invocan de verdad, no quedan punteros colgantes, y la auditoría final
está limpia.
