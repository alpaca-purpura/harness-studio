# Análisis — drawer del mockup v3 (arnesia-mockup-v3.html) vs nuestro drawer

> 2026-07-07 · pedido del operador: revisar el inspector del mockup navegable
> (`mockups/arnesia-mockup-v3.html#/a/luana-vitalia/mapa?…&sel=s-dev-team`) con ojo-UI
> (Chrome DevTools) y recomendar qué traer al paquete inspector-drawer.
> Evidencia: `scratchpad/{v3-drawer-devteam,v3-drawer-acciones}.png` + innerText de las 3 tabs.

## Qué tiene el drawer v3 (inventario observado)

1. **Tabs** `Resumen | Contenido | Corridas`.
2. **Header**: glyph+clase · id · `sin versión` · chip canal `ESTABLE`.
3. **Descripción en prosa** del nodo (1–2 líneas, ej. «Router del developer… Componente más pesado»).
4. **KPI tiles** (Resumen): invocaciones 30d · éxito (`—` «sin criterio») · p95 · tokens al cargar ·
   tokens por invocación · excepciones de proceso 30d · consumo 30d con nota «cada componente rinde
   cuentas de sus tokens (principio 11)». Honestidad integrada: `—` cuando no hay dato/criterio.
5. **CONTRATO** con etiqueta de honestidad «· inferido de la prosa · por formalizar»:
   NECESITA↑ como **chips tipados y navegables** (`READY PACKAGE [ready] · caja Arquitectura`,
   `CONTEXT-BRIEF · maquinaria · context-builder`) · ENTREGA · RUTA↓ con **condición como badge**
   (`si necesita engine lift (R23)`) · **VIENE DE** (edges inversos: architect, auditor).
6. **HALLAZGOS**: caja de hallazgos por nodo («Mayor consumidor del arnés… candidato a acotar vía
   CONTEXT-BRIEF») o «Sin hallazgos abiertos».
7. **Acciones al pie**: `Editar conversando` (primaria) · `Evaluar A/B contra v anterior` ·
   `Promover a estable` · `Ver en Historia`.
8. **Tab Contenido**: la FUENTE del componente (frontmatter+md) + chip «versiona con el arnés» +
   acciones `Editar fuente` / `Editar conversando (dock)` + nota del flujo (CodeMirror 6, diff →
   beta → tren, versiones inmutables).
9. **Tab Corridas**: corridas donde el nodo actuó (fecha · duración · resultado) + link «ver todas».
10. **Estado vacío** del inspector: «Clic en un nodo: identidad, métricas, hallazgos, contenido
    editable y sus corridas».
11. Para apoyo (hook): mismo esqueleto, contrato = una línea de encuadre (nuestro Tier A ya lo
    supera con encuadre por clase + pendientes honestos).

## Veredicto — qué traería (mi criterio, filtrado por doctrina y dato disponible)

### A · Traer YA (el dato existe hoy en el grafo)
- **VIENE DE (edges inversos)** — quién invoca/alimenta a ESTE nodo. Los edges ya viajan en el
  grafo; solo mostramos el sentido saliente. Barato y de alto valor de lectura.
- **Chips tipados y navegables en NECESITA/RUTA** — hoy texto plano («← base:brand-voice»); como
  chip con el origen tipado (caja:/base:/maquinaria:/usuario) y click → selecciona ese nodo en el
  mapa. La condición de RUTA como badge diferenciado.
- **Estado vacío del inspector** — una línea de affordance en vez de nada.
- **HALLAZGOS (versión conformance)** — sin telemetría ya podemos derivar hallazgos deterministas
  del nodo: `gate:none` (A4) · `no-reconocido` (D-c) · checks rojos del endpoint
  `GET /api/harnesses/{id}/conformance` filtrados por nodo. «Sin hallazgos abiertos» como estado.
- **Acciones al pie (staged)** — `Editar conversando` es doctrina (crear/editar = acciones sobre el
  mapa; Fase 3/4 del Hito 2 ya tiene el backend E). Traer la botonera con lo no-cableado disabled
  y rotulado, jamás fingiendo que funciona.

### B · Traer el DISEÑO ya, rotulado «necesita telemetría» (Hito 3 lo puebla)
- **KPI tiles** con la honestidad del v3 (`—` = sin dato · «sin criterio» explícito · nota
  principio 11). Es la joya del drawer v3 y calza con nuestras capas Tokens/Desempeño staged.
- **Tab Corridas** (corridas donde actuó) — espera indexer JSONL.

### C · Traer con Tier B (necesita dato per-class que el loader aún no extrae)
- **Descripción en prosa bajo el header** — para cajas ya existe (`why`); para apoyo llega con
  `description` del frontmatter (ya listado en inspector-por-clase.md Tier B).
- **Tab Contenido (fuente read-only)** — tenemos `fuente_path`; el daemon puede servir el archivo
  confinado. Editor (CodeMirror + diff→beta→tren) = Fase 3/4, NO de este paquete.
- **Versión en el header** — dato del release train (KIT-06); cuando exista.

### D · NO traer / lo nuestro es mejor
- El encuadre de apoyo del v3 («Componente de apoyo — infra compartida») — nuestro Tier A por
  clase + pendientes honestos es más rico.
- El v3 NO dibuja arquetipo/perfil/gate/procedencia/origen — nuestra Clasificación + tooltips
  doctrinales (#2) va más profundo; se conserva.
- «Evaluar A/B» y «Promover a estable» como acciones vivas — tren de release fuera del alcance de
  este paquete; solo staged.

## Profundización: tabs Contenido y Corridas (pedido del operador, mismo día)

> Revisado en UI (DevTools, screenshots `v3-tab-{contenido,corridas}.png` + vista global
> `v3-corridas-view.png`) Y en el código fuente del mockup (data model `RUNS` línea 1008,
> `renderRunDet` línea 1891).

### Qué es la tab CONTENIDO (v3)

La FUENTE del componente, visible en el drawer: viewer de código con números de línea
(frontmatter YAML + markdown), chip «versiona con el arnés», y dos acciones: **Editar
fuente** (CodeMirror 6 embebido; guardar = diff antes de confirmar + nota de cambio →
nace beta → tren; versiones inmutables, deploy = mover canal jamás editar historia) y
**Editar conversando (dock)**.

### Qué es la tab CORRIDAS (v3) — el modelo completo

**Corrida = una sesión REAL de Claude Code tal como ocurrió** (el JSONL de `~/.claude`;
su propio subtítulo: «Nada es flujo ideal: es lo que el usuario vivió»). Data model por
corrida: `sid` (session id) · fecha · componente disparador · fase · modelo · turnos ·
tokens in/out/caché · costo · ctx máx % · duración · exitosa/fallida · excepciones de
proceso · `pasos[]` tipados (user/hook/skill/llm/tool/agent-con-hijos-sidechain/cmd/
compact/err/asst), cada paso con tokens, tiempo, **contexto acumulado** y anotación de
qué añadió a la ventana.

- **Tab del drawer** = «CORRIDAS DONDE ACTUÓ»: lista filtrada al nodo (dot estado ·
  fecha · duración · resultado) + «Ver todas las corridas del arnés».
- **Vista global Corridas** = lista del arnés con filtros Todas/Fallidas.
- **Detalle de corrida** (vista propia, NO drawer): stats bar (badge «CORRIDA REAL ·
  JSONL») + 3 vistas — **Conversación** · **Árbol** (pasos con barra de contexto
  acumulado + link «ver componente en el mapa →») · **Waterfall** (timeline por tipo) +
  **«▶ Reproducir en el mapa»** (replay de la corrida sobre el mapa).

### Propuesta de integración a NUESTRO drawer (por etapas honestas)

**Estructura: tabs `Resumen | Contenido | Corridas`** en el drawer (header identidad
fijo encima, como v3). Resumen = todo lo de v4.

**Contenido — etapa 1 EN ESTE PAQUETE (read-only):**
- Dato: `fuente_path` ya viaja en L0 para todas las clases; el daemon tiene el dir del
  arnés registrado (S2, confinado). Falta solo un endpoint pequeño
  `GET /api/harnesses/{id}/nodes/{nodeId}/fuente` (lectura confinada al dir del arnés,
  auth Host+Origin+token ya existente).
- Drawer: viewer mono read-only con números de línea + chip «versiona con el arnés».
  Universal por clase (rule→CLAUDE.md · mcp/settings→JSON · **no-reconocido→ver el
  artefacto que el loader no entendió** = reconciliación D-c accionable).
- Acciones staged: «Editar fuente» y «Editar conversando» disabled rotuladas Fase 3/4
  (backend Fase E vivo; CodeMirror 6/merge ya firmado en el stack HS-04 para el dock).
- En el MOCKUP: fuente reconstruida del dato real del grafo, rotulada «reconstruido —
  el showcase no vive en disco»; jamás contenido inventado en silencio.

**Corridas — etapa 1 EN ESTE PAQUETE (diseño + dato disponible):**
- El modelo v3 ES nuestro event sourcing firmado (HS-04: stream-json vivo + JSONL
  enumerar/replay) — la tab es la cara visible de la deuda «telemetría/indexer JSONL».
- Drawer: lista «Corridas donde actuó» con estado honesto «necesita indexer JSONL» +
  lo que YA existe: corridas de caja del endpoint `POST …/boxes/{boxId}/run` (D2) y la
  sesión CC viva de la sesión-frente (claude_session_id persistido) cuando aplique.
- El DETALLE de corrida (Conversación/Árbol/Waterfall + replay en el mapa) = superficie
  propia fuera del drawer → capa Proceso / vista Corridas (Hito 3, UX S8). La tab solo
  LISTA y enlaza.

**Qué NO entra ahora:** editor vivo (Fase 3/4) · detalle/replay de corridas · costo/ctx
por corrida (indexer). Todo staged rotulado, nada finge funcionar.

## Estado

Análisis entregado al operador — las que apruebe se vuelven decisiones #3+ en `decisiones.md`
y entran al mockup como vN. Bloque §A aprobado y materializado (decisión #3, v4).
Propuesta de tabs Contenido/Corridas esperando firma (→ decisión #4).
