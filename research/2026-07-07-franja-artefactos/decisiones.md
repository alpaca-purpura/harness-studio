# Decisiones del paquete franja-artefactos

Registro en el mismo turno (disciplina §10). Fundamento de cada una en
[`viabilidad.md`](./viabilidad.md).

**Estado (2026-07-08): D1–D11 FIRMADAS 🧑‍⚖️** — D1/D3–D8 («firmemos todo») · D2 sobre el
mockup v1 (geometría **A · chips en spine**) · D9–D11 con el OK a la casuística y la
orden de pasar a backlog implementable («agrégalo al backlog… para programarlo»).
Catálogo en [`casuistica.md`](./casuistica.md); demo en el mockup v2. **Próxima firma
pendiente: spec.md + design.md** (gate de la sesión de implementación).

## D1 — Artefacto = proyección del contrato, jamás primera clase del L0 — ✅ FIRMADA
El artefacto NO es 11ª `clase` (enum crece solo con primitivas CC-native), NO es nodo
declarado del manifiesto, NO es valor nuevo del enum `banda`. Toda superficie visual se
DERIVA de `necesita[]`/`entrega[]` (postura P5/HS-12: se proyecta, jamás descriptor
aparte). Propuesta: cementarlo con una línea en VISION §Linaje.

## D2 — Geografía: chips de hand-off a la altura del spine, no banda al fondo — ✅ FIRMADA (2026-07-08, sobre el mockup: «se ve y entiende mucho mejor»)
La proyección visual vive en el gutter entre carriles consecutivos: un chip por hand-off
(output de caja N ≡ input de caja N+1), edges `escribe`/`lee` derivados, revelado por
selección + toggle en MapBar (no 5ª capa). Inputs de usuario/terceros = chip
huérfano-visible; entregas sin consumidor = chip dead-end warn. Frontera A3: solo TIPOS
declarados; instancias vivas = tablero Flujo aparte (decisión #6 Gate 1 intacta).
**El mockup compara chip-en-gutter vs franja-abajo y el gate humano decide.**

## D3 — Identidad del art (prerequisito de todo) — ✅ FIRMADA
`entrega[].path?` (artefacto-archivo vs artefacto-etiqueta) y `entrega[].plantilla?`
(ruta a la reference de la skill escritora), aditivos en box.contract.schema.json
(patrón reparación P3) + espejo domain.Output. El conductor deja de descartar el error
de `Status` en silencio. Check warn `art-es-path` para cajas pipeline/excepcion.

## D4 — Plantilla: rule Base corta + reference de la skill-caja escritora — ✅ FIRMADA
Estándar corto = rule de banda Base (`base:<std-id>`, patrón std-spec). Esqueleto largo
= `skills/<caja>/references/plantilla-<art>.md` + validador en `scripts/`. Jamás
always-on (p11). Command descartado. Nomenclatura v1.2 solo si el mockup exige
plantillas reconocibles fuera de references/.

## D5 — Llenado casi determinista: script para estructura, LLM para juicio — ✅ FIRMADA
Scaffold/copias/validación/digest = scripts (dial low-freedom, check
skill-script-for-deterministic). Prosa de juicio = LLM; opcional plan-validate-execute
con `--json-schema` + renderer determinista. `status: done` lo estampa el validador,
no el LLM. Obligatorio SOLO para arquetipos pipeline/excepcion (`abierto` exento,
§8.1/§8.3).

## D6 — Gates: un validador, tres puntos (PostToolUse · Stop · gate.tipo:auto Gherkin) — ✅ FIRMADA
PostToolUse bloquea con JSON decision:block (exit 2 no bloquea ahí); Stop = gate final
con stop_hook_active; el Gherkin del gate auto referencia la validación de plantilla.
Sin plantilla → tipo: none, jamás fabricado.

## D7 — Encadenado por filesystem: rutas + digest, jamás el doc entero — ✅ FIRMADA
Conductor verifica precondiciones `necesita[]` (existencia + status=done de requeridos)
antes de Spawn; `tarea()` inyecta rutas + `<art>.digest.md` (≤200 tok) generado
determinísticamente al cierre. Medir tokens antes/después (p11).

## D8 — Motor primero: checks de composición prometidos → vivos — ✅ FIRMADA
`sin-huerfanos` (deferred→vivo), `dead-end` (fila nueva), `ruta-a-existe`,
`art-identidad-coherente` — gemelos de VerificarEscritorUnico, cableados a `--arnes`.
Es deuda ya prometida (METODOLOGIA §3/§6, UX it.10) — ejecutable sin ninguna otra firma.

## D9 — Mejora de artefacto: `entrega[].refina` (cadena de revisiones) — ✅ FIRMADA
Cuando el output ES el mismo artefacto de entrada mejorado (draft.md → draft.md editado,
factura.pdf → factura.pdf validada): la entrega declara `refina: "<art>"`. Reglas:
(a) el refinador DEBE tener ese mismo `art` en su `necesita` (check `refina-coherente`);
(b) `escritor_unico` pasa a evaluarse por REVISIÓN — la cadena es lineal, un solo escritor
por transición del art; dos entregas del mismo art SIN `refina` siguen siendo hallazgo;
(c) la versión (v2, v3…) se DERIVA de la cadena, jamás se declara; (d) los consumidores
aguas abajo referencian la revisión por `de: caja:<refinador>`. Distinto del REWORK
(ruta de vuelta al MISMO escritor que itera su propia entrega — no es refina, no crea
revisión nueva; vive en `ruta[]` y en el tablero Flujo). Visual: el chip se repite en el
gutter del refinador con badge `↻ v2`. Aditivo a box.contract.schema.json (patrón P3).

## D10 — Artefactos sin plantilla propia (factura) y opacos — ✅ FIRMADA
La plantilla SOLO aplica a entregas que el arnés PRODUCE (D4). Un input que llega del
mundo (factura, orden de compra, aprobación) = `necesita.de: usuario|terceros` — jamás
lleva plantilla; su control es ADMISIÓN: precondición de existencia del conductor (D7,
para `requerido: true`) + validación de forma en `scripts/` de la skill consumidora
(existe, mime/formato, campos mínimos extraíbles) — encarnada en el componente, sin campo
nuevo de contrato (p1). Entregas OPACAS (pdf, binarios, datasets): `path` sí, plantilla
no; el gate valida forma por script; document-as-cache NO aplica al binario — el status
del trabajo vive en el document-as-cache de la caja (digest sidecar determinista = deuda
declarada, no se promete). Etiquetas sin path («código + tests») siguen D3.

## D11 — Densidad y largo alcance en el Mapa (vista A firmada) — ✅ FIRMADA
(a) El chip vive UNA sola vez: en el gutter del hand-off donde NACE (tras su productor);
consumidores lejanos = edge `lee` largo, jamás chip duplicado. (b) Al SELECCIONAR una
caja, su gutter de entrada muestra el **panel de entrada**: referencias compactas ↖ de
TODOS sus `necesita` de largo alcance (índice navegable — click selecciona al productor);
los chips reales se iluminan donde viven. (c) Tope de densidad: máx 3 chips visibles por
gutter + botón «+N más» que expande; los chips relacionados con la selección tienen
prioridad sobre el tope. (d) `requerido: false` = edge tenue + tag «opcional» en la
referencia. (e) Externos (usuario/terceros) se pintan en el gutter de entrada de su
consumidor (patrón huérfano-visible §4.5, jamás omitidos).

## Impacto DevStudio (constatación, no decisión)
Nada que actualizar hoy: todo aditivo bajo el compromiso HS-12; enum-5 intacto (línea
roja: no reutilizarlo para estados de artefacto); ficha gemela de cortesía al firmar D3.
