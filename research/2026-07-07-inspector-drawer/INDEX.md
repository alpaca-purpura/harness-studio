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

- [x] mockup-drawer.html clonado del drawer actual (v1, 2026-07-07) — **artifact
      (publicar SIEMPRE a esta URL):** https://claude.ai/code/artifact/99e127f9-b7f2-48c8-b506-ae0fb376743d
- [ ] iteraciones sobre el mockup → firma
- [ ] decisiones.md (crece por iteración)
- [ ] spec.md + design.md → firma
- [ ] implementación + stories
- [ ] PARIDAD.md verificada → gate final

## Retomar aquí

> Se actualiza al cierre de cada turno de trabajo (METODOLOGIA §10). Una sesión nueva
> lee esto y continúa como si fuera la misma conversación.

- **Último hecho (2026-07-07):** mockup **v2** publicado (misma URL) con la
  **decisión #1 en PROPUESTA** (`decisiones.md`): botón «ampliar» ⤢ junto al de
  cerrar → el drawer ocupa todo el espacio visual del mapa; ⤡/Esc restauran; en
  amplio las secciones fluyen en columnas (`minmax(280px,1fr)`). INTERACTIVO en el
  mockup — verificado ojo-UI (toggle+Esc+aria-pressed, consola limpia). v1 = clon
  fiel, 15 casuísticas 01–15, stubs «sintético» (08 · 09 · 15).
- **Decisión #1 FIRMADA** (operador, 2026-07-07): botón «ampliar» tal como está en v2.
- **Próximo paso:** el operador dicta el siguiente cambio → `decisiones.md` en el
  mismo turno + vN a la MISMA URL.
- **Firmas pendientes:** mockup (ni siquiera v1 aún) · Tier B (campo `meta` per-class
  en L0 — se firmará en `decisiones.md` si las iteraciones piden dato que hoy no viaja).
- **Contexto caliente:** el operador quiere pedir cambios iterativos sobre el mockup SIN
  tocar código; el inspector de contrato (constraints/non_goals/Gherkin/evidencia/
  mini-spine, `ui-doctrina-visible.md` §Inspector) queda EN PAUSA — sus decisiones ahora
  se toman dentro de este paquete, no por fuera.

## Contexto que este paquete NO duplica (solo enlaza)

- Estado actual del drawer: `web/src/widgets/map-canvas/ui/inspector.tsx` (Tier A del
  inspector por clase YA implementado: encuadre por clase · Fuente · Activación).
- Tabla clase→campos y tiers: `../2026-07-06-plan-hito2-doctrina-edicion-showcase/inspector-por-clase.md`.
- Huecos de contrato pendientes (constraints/non_goals/Gherkin/evidencia/mini-spine):
  `../2026-07-06-plan-hito2-doctrina-edicion-showcase/ui-doctrina-visible.md` §Inspector.
- Contrato de datos: `web/src/entities/arnes/model/types.ts` (espejo L0) ·
  `arch/contracts/schema/graph.l0.schema.json`.
- Dato de prueba: `../2026-07-06-plan-hito2-doctrina-edicion-showcase/showcase.graph.json`.
