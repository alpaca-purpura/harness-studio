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
- [x] iteraciones sobre el mockup → **FIRMADO** (v6; decisiones #1–#5 firmadas, #6
      resuelto por medición — «Firmo de momento», 2026-07-07)
- [x] decisiones.md (6 entradas, todas cerradas)
- [x] spec.md (RF-80..96 + Gherkin + trazabilidad) + design.md (UI al pixel) — 🧑‍⚖️
      **PAQUETE FIRMADO** (operador «firmo», 2026-07-07) → implementación AUTORIZADA
- [ ] implementación + stories
- [ ] PARIDAD.md verificada → gate final (esqueleto listo, 16 filas)

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
- **Decisión #2 en PROPUESTA (v3):** tooltips en dos niveles — «i» por título de sección
  (qué agrupa) + campos con subrayado punteado (tooltip = qué significa el campo Y qué
  significa el valor concreto). **Restricción firmada del operador: contenido 100%
  consistente con la doctrina** — diccionario `SEC_TIP`/`DEF_CAMPO`/`DEF_VALOR` en el
  mockup con fuentes citadas (METODOLOGIA §3/§4/§8 · VISION A1–A7 · schema L0 · rules
  L2.6); al implementar se cementa en la entity FE. Verificado ojo-UI: 65 campos con
  tip · 51 «i» de sección · consola limpia.
- **Decisión #3 FIRMADA + materializada (v4):** bloque «traer YA» del análisis v3
  (`analisis-drawer-v3.md §A`) — VIENE DE (edges inversos reales con tipo; solo 5 nodos
  del showcase los tienen, el resto oculta la sección) · chips tipados NAVEGABLES en
  Necesita/Ruta (click → salta a la casuística; fuera del mockup = data-off rotulado) ·
  condición de ruta como badge (caso real: tarjeta 16 review-caja «si cambios mayores
  solicitados») · HALLAZGOS por nodo (gate:none crit · no-reconocido warn · «sin
  hallazgos» + nota del endpoint conformance) · botonera staged (Editar conversando
  primaria disabled «Fase 3/4»; resto rotulado tren/Historia) · tarjeta 00 estado vacío.
  Ahora 17 tarjetas (00–16; 16 = review-caja, ruta condicional real). Verificado ojo-UI
  (navegación de chips OK, consola limpia).
- **Decisión #4 FIRMADA + materializada (v5):** tabs `Resumen | Contenido | Corridas`.
  Contenido = fuente read-only con números de línea (reconstruida del dato real y
  rotulada: cajas/subagentes → frontmatter+contract YAML · rule → convención CLAUDE.md ·
  clases sin forma → estado honesto · no-reconocido → nota raw D-c) + chip «versiona
  con el arnés» + acciones staged (Editar fuente/conversando, Fase 3/4). Corridas =
  estado honesto «necesita indexer JSONL» + para cajas la nota de qué lista primero
  (runs D2 + sesión viva) + «Ver todas» disabled (detalle/replay = Hito 3). Análisis
  de base: `analisis-drawer-v3.md` §tabs (modelo completo del v3 leído de su código:
  RUNS/pasos tipados/3 vistas/replay). Verificado ojo-UI, consola limpia.
- **Decisión #5 FIRMADA + v6:** UI del expandido (tabs sticky+compactas · botonera fila
  al pie · columna lectura 920px) + semántica ✕ cierra ≠ ⤡ colapsa. Fix v5.1 (pane
  colado por especificidad).
- **Debate #6 RESUELTO POR MEDICIÓN** (`spike-medicion-contrato.md`): CC inyecta solo
  el CUERPO de la skill al activar (~187 tok); el contrato fusionado en frontmatter =
  0 tokens de contexto. El yml hermano es innecesario para skills; contrato fusionado
  validado. Rules/CLAUDE.md = única superficie sensible (ya cubierta).
- **#2 FIRMADA («Firmo de momento» — redacciones ajustables en paridad) → mockup v6 =
  FIRMADO completo.** Specs escritos: `spec.md` (RF-80..96, Gherkin, trazabilidad a
  mockup:línea de v6/`b2b7104`) + `design.md` (anatomía, tokens, tabla clase→contenido,
  estados, a11y) + `PARIDAD.md` esqueleto (16 filas).
- **PAQUETE FIRMADO (2026-07-07).** Implementación autorizada; el prompt riguroso de
  arranque vive en [`PROMPT-implementacion.md`](./PROMPT-implementacion.md) (fase de
  revisión previa → diseño técnico → implementación por RF → validación tests +
  Chrome DevTools funcional Y visual → PARIDAD fila por fila).
- **FASE 1 SELLADA (2026-07-07):** diseño técnico + mini-plan de 7 commits en
  [`plan-implementacion.md`](./plan-implementacion.md). Claves: diccionario →
  `entities/arnes/model/doctrina.ts` · viene-de/hallazgos → selectors.ts · conformance
  por nodo = FILTRO EN FE (justificado: la atribución por nodo no existe como dato en el
  reporte) · transporte queda en la página (inyecta `conformance` + `loadFuente`) ·
  endpoint fuente = puerto `FuenteReader` + adapter confinado (patrón artifact) +
  `FuenteService` + ruta text/plain + OpenAPI 0.2.0-hs09 · interpretación RF-81↔RF-84
  registrada (tras ✕ queda el drawer VACÍO affordance — a consultar en paridad).
- **Próximo paso:** FASE 2 — implementar commits 1..6 del plan; luego FASE 3 validación
  Chrome DevTools + PARIDAD fila por fila.
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
