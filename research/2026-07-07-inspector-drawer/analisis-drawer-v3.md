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

## Estado

Análisis entregado al operador — las que apruebe se vuelven decisiones #3+ en `decisiones.md`
y entran al mockup como vN.
