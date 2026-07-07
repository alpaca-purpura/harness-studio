# Inspector/drawer del Mapa — iteración por mockup (paquete de trabajo)

> Ficha HS-09 (fase 5) · Hito 2 · profundización de Fase 2 · 2026-07-07
> Origen: el operador quiere iterar QUÉ se ve al seleccionar cada elemento del mapa
> SIN tocar el código — mockup HTML clonado del drawer real, decisiones firmadas,
> specs, y recién entonces implementación. Mismo ritmo probado del Mapa
> (`research/2026-07-06-mapa-mvp/`): decisión → as-code → build → verificar humano.

## Flujo y gates

1. **Mockup** (`mockup-drawer.html`) — clon fiel del drawer actual, se itera con el
   operador hasta 🧑‍⚖️ **firma del mockup**. Reglas: tokens DTCG reales (nunca estilos
   inventados — norma «pegarse al Storybook») · datos REALES del showcase embebidos
   (el mismo JSON que sirve el daemon) · una casuística por clase (caja rica · skill
   apoyo · regla siempre/condicional · hook · mcp · plugin · settings · output-style ·
   statusline · subagent · no-reconocido).
2. **Decisiones** (`decisiones.md`) — una entrada por decisión conversada, estado
   PROPUESTA → FIRMADA, con el porqué. Si un cambio pide dato que hoy no viaja en L0,
   la decisión Tier B (campo `meta` per-class, ver `../2026-07-06-plan-hito2-doctrina-
   edicion-showcase/inspector-por-clase.md`) se firma AQUÍ.
3. **Specs** — 🧑‍⚖️ **firma del paquete**:
   - `spec.md` — RF numerados + Gherkin, cada RF trazado a `mockup:línea`.
   - `design.md` — UI al pixel: secciones, orden, tabla clase→campos, tokens por marca.
4. **Implementación** — contra el spec firmado, en `inspector.tsx`/`entities/arnes`
   (+ loader/schema si Tier B se firmó). Story=test por marca nueva.
5. **Paridad** (`PARIDAD.md`) — matriz mockup ↔ componente real ↔ story ↔ RF; 🧑‍⚖️
   **gate final**: click-through app vs mockup lado a lado, consola limpia,
   screenshots revisados.

## Estado

- [ ] mockup-drawer.html clonado del drawer actual (pendiente — lo pide el operador)
- [ ] iteraciones sobre el mockup → firma
- [ ] decisiones.md (crece por iteración)
- [ ] spec.md + design.md → firma
- [ ] implementación + stories
- [ ] PARIDAD.md verificada → gate final

## Contexto que este paquete NO duplica (solo enlaza)

- Estado actual del drawer: `web/src/widgets/map-canvas/ui/inspector.tsx` (Tier A del
  inspector por clase YA implementado: encuadre por clase · Fuente · Activación).
- Tabla clase→campos y tiers: `../2026-07-06-plan-hito2-doctrina-edicion-showcase/inspector-por-clase.md`.
- Huecos de contrato pendientes (constraints/non_goals/Gherkin/evidencia/mini-spine):
  `../2026-07-06-plan-hito2-doctrina-edicion-showcase/ui-doctrina-visible.md` §Inspector.
- Contrato de datos: `web/src/entities/arnes/model/types.ts` (espejo L0) ·
  `arch/contracts/schema/graph.l0.schema.json`.
- Dato de prueba: `../2026-07-06-plan-hito2-doctrina-edicion-showcase/showcase.graph.json`.
