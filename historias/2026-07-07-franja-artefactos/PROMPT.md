# Prompt de arranque — sesión de implementación

Copia y pega esto en una conversación nueva de Claude Code (cwd = raíz del repo):

---

Retoma el paquete de trabajo `historias/2026-07-07-franja-artefactos/` y continúa como si
fueras la misma conversación (disciplina METODOLOGIA §10).

Orden de lectura obligatorio antes de tocar NADA:
1. `historias/2026-07-07-franja-artefactos/INDEX.md` («Retomar aquí»)
2. `decisiones.md` — D1–D11 FIRMADAS; son ley, no se relitigan
3. `spec.md` (RF-100..151) y `design.md` (diseño técnico por fase) — estado
   `pendiente-de-firma`
4. `casuistica.md` y `mockup-artefactos.html` (ábrelo en navegador: es la referencia
   visual firmada; asserts del click-through documentados en INDEX)

Primer gate: preséntame un resumen de spec.md + design.md (qué se construye por fase, qué
NO, riesgos) y espera mi firma 🧑‍⚖️. Si algo del spec contradice el código actual del
repo, repórtalo ANTES de la firma, no lo arregles en silencio.

Tras la firma, implementa por fases EN ORDEN (design.md §Orden y gates): Fase 1 checks de
composición → Fase 2 identidad del art → Fases 3/4 conductor + plantillas dogfood (con
medición p11 real) → Fase 5 Mapa (port del mockup v2, stories por marca, PARIDAD.md al
cierre) → Fase 6 nomenclatura/ficha DevStudio. Reglas duras: cada fase termina con la
suite verde (`go test ./...`, golangci, round-trip `arnesia index dogfood/dev-full-cycle`
+ `conformance --arnes` TODO PASS, y en FE vitest+stories+depcruise+biome), un commit a
main y el «Retomar aquí» del INDEX actualizado. Todo cambio de schema es ADITIVO (patrón
P3/HS-12). El artefacto JAMÁS es nodo del L0, banda ni clase 11ª. Nada de pass fabricado:
si un check nuevo caza algo real del dogfood, es hallazgo visible, no se silencia. Al
terminar la Fase 5 me entregas PARIDAD.md para el gate final lado a lado contra el mockup.

---

(Contexto mínimo si la sesión lo pide: el paquete nació de la idea «4ª franja Artefactos»;
la evaluación completa vive en `viabilidad.md`; la geometría firmada es chips de hand-off
en el spine, todo derivado de `necesita[]/entrega[]` de los contratos de caja.)
