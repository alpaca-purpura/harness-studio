# Decisiones del paquete franja-artefactos

Registro en el mismo turno (disciplina §10). Fundamento de cada una en
[`viabilidad.md`](./viabilidad.md).

**Estado (2026-07-08): D1 · D3 · D4 · D5 · D6 · D7 · D8 FIRMADAS 🧑‍⚖️ por el operador
(«firmemos todo»). D2 FIRMADA PARCIAL: proyección-jamás-banda-declarada queda firmada
dentro de D1; la GEOMETRÍA (A · chips en spine vs B · franja abajo) se decide sobre
[`mockup-artefactos.html`](./mockup-artefactos.html) — gate abierto.**

## D1 — Artefacto = proyección del contrato, jamás primera clase del L0 — ✅ FIRMADA
El artefacto NO es 11ª `clase` (enum crece solo con primitivas CC-native), NO es nodo
declarado del manifiesto, NO es valor nuevo del enum `banda`. Toda superficie visual se
DERIVA de `necesita[]`/`entrega[]` (postura P5/HS-12: se proyecta, jamás descriptor
aparte). Propuesta: cementarlo con una línea en VISION §Linaje.

## D2 — Geografía: chips de hand-off a la altura del spine, no banda al fondo — 🚧 GATE ABIERTO (mockup)
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

## Impacto DevStudio (constatación, no decisión)
Nada que actualizar hoy: todo aditivo bajo el compromiso HS-12; enum-5 intacto (línea
roja: no reutilizarlo para estados de artefacto); ficha gemela de cortesía al firmar D3.
