# spec — drawer/inspector del Mapa (QUÉ)

> Ficha HS-09 · Hito 2 · paquete inspector-drawer · 2026-07-07
> Base firmada: decisiones #1–#5 (`decisiones.md`) sobre `mockup-drawer.html` **v6**
> (URL viva en INDEX.md). Cada RF traza a `mockup:línea` de v6 (commit `b2b7104`).
> El drawer actual de la app (`web/src/widgets/map-canvas/ui/inspector.tsx`, Tier A del
> inspector-por-clase) es el punto de partida: estos RF son ADITIVOS sobre él.

## Alcance

Read-only + staged. **NO incluye** (no-goals): editor vivo (Fase 3/4) · detalle/replay
de corridas (Hito 3) · KPIs poblados (indexer JSONL) · extracción per-class Tier B
(firma propia) · acciones de release train vivas. Nada finge funcionar: lo no cableado
va disabled y rotulado.

## RF — Shell del drawer

- **RF-80 (ampliar).** El header lleva ⤢ al costado de ✕ (`mockup:524`). Click → el
  drawer ocupa TODO el espacio visual del mapa; el header queda sticky. [dec #1]
- **RF-81 (colapsar ≠ cerrar).** En expandido: **⤡ colapsa** al drawer lateral normal;
  **✕ CIERRA** el drawer del todo (queda el mapa sin drawer); `Esc` colapsa. Tooltips
  nativos en ambos botones lo explican (`mockup:524-525,552-568`). [dec #5e]
- **RF-82 (tabs).** Tres tabs `Resumen | Contenido | Corridas` (`mockup:527-531`),
  default Resumen, `role=tablist`/`aria-selected`; conmutan panes sin perder el header.
  [dec #4]
- **RF-83 (expandido, UI).** En expandido: tabs **sticky** bajo el header y compactas
  alineadas a la izquierda (`mockup:162-163`); el Resumen fluye en rejilla
  `minmax(280px,1fr)` (`mockup:159`); la botonera es fila completa al pie
  (`mockup:164-166`); Contenido/Corridas usan columna de lectura `max-width:920px`
  (`mockup:167`). En drawer normal las tabs son full-width 3-up. [dec #5a-d]
- **RF-84 (estado vacío).** Sin selección, el inspector muestra la línea de affordance
  (`mockup:541-550`), no un panel en blanco ni ausencia. [dec #3e]

## RF — Tooltips doctrinales

- **RF-85 (secciones).** Cada título de sección lleva una «i» circulada
  (`mockup:125,451-456`) cuyo tooltip dice qué agrupa Y su eje del contrato fusionado.
  Contenido = diccionario `SEC_TIP` (`mockup:309-326`). [dec #2]
- **RF-86 (campos).** Cada nombre de campo va con subrayado punteado + cursor help
  (`mockup:124`); su tooltip = definición del campo + significado del VALOR concreto
  (`DEF_CAMPO`/`DEF_VALOR`/`tipDe`, `mockup:327-362`). Valores PROPUESTA añaden la nota
  de fixture. [dec #2]
- **RF-87 (fuente doctrinal).** El diccionario cita doctrina (METODOLOGIA §3/§4/§8 ·
  VISION A1–A7 · schema L0 · rules L2.6) y al implementar se cementa UNA vez en
  `entities/arnes/model` (la entity es el único dueño; el widget solo lo consume).
  [restricción del operador, dec #2]
- **RF-88 (a11y).** Tooltips accesibles por teclado (`tabindex=0` + `:focus-visible`);
  ningún texto warn a 11px sobre fondo claro — el warn va en fondo `--warn-soft` con
  texto `--foreground` (lección a11y de la fase 2: contraste AA).

## RF — Resumen enriquecido

- **RF-89 (Viene de).** Sección con los edges INVERSOS del nodo con su tipo
  (invoca/lee/escribe), derivados del grafo (`mockup:364-371,509-512`); sin entradas →
  sección ausente. FE: selector inverso sobre `graph.edges` ya cargados. [dec #3a]
- **RF-90 (chips navegables).** `necesita`/`ruta` se dibujan como chips tipados
  (`mockup:373-378,489-499`): origen `tipo: id`; click en chip cuyo destino existe en
  el grafo → SELECCIONA ese nodo en el mapa (mismo mecanismo que click en nodo);
  destino ausente → chip inerte rotulado. Condición de ruta = badge punteado warn
  (`mockup:139,497`). [dec #3b]
- **RF-91 (Hallazgos).** Sección SIEMPRE presente (`mockup:379-388,515-517`):
  determinista del dato — `gate:none` (crit-soft, A4) · `no-reconocido` (warn-soft,
  D-c) · checks rojos de `GET /api/harnesses/{id}/conformance` FILTRADOS por nodo ·
  «Sin hallazgos abiertos» como estado. [dec #3c]
- **RF-92 (botonera staged).** Pie del Resumen: «Editar conversando» (primaria,
  disabled, rotulada Fase 3/4) · «Evaluar A/B» · «Promover a estable» · «Ver en
  Historia» (disabled, rotuladas tren/Historia) + nota staged (`mockup:501-508`).
  [dec #3d]

## RF — Tab Contenido

- **RF-93 (fuente read-only).** Muestra el archivo real del nodo: viewer mono con
  números de línea (`mockup:410-433`), chip «versiona con el arnés», ruta visible.
  Backend nuevo: `GET /api/harnesses/{id}/nodes/{nodeId}/fuente` — lectura CONFINADA
  al dir registrado del arnés (S2; auth Host+Origin+token existente; 404 si el nodo no
  tiene `fuente_path`). [dec #4]
- **RF-94 (no-reconocido raw).** Para `clase:no-reconocido` la tab muestra el artefacto
  TAL CUAL (raw) — la reconciliación D-c se vuelve accionable. [dec #4]
- **RF-95 (acciones staged).** «Editar fuente» + «Editar conversando (dock)» disabled
  con la nota del flujo real (diff antes de confirmar → nace beta → tren). [dec #4]

## RF — Tab Corridas

- **RF-96 (estado honesto).** «Corridas donde actuó» con estado «Sin corridas
  indexadas — llegan con el indexer JSONL (Hito 3)» (`mockup:434-443`); en cajas, nota
  de qué listará primero (runs D2 + sesión CC viva). «Ver todas» disabled. El detalle
  (Conversación/Árbol/Waterfall/replay) queda EXPLÍCITAMENTE fuera. [dec #4]

## Gherkin (aceptación por comportamiento)

```gherkin
Feature: Drawer del Mapa — shell
  Scenario: Ampliar, colapsar y cerrar
    Given el drawer abierto de una caja
    When clickeo ⤢
    Then el drawer ocupa todo el espacio del mapa con header y tabs sticky
    When clickeo ⤡
    Then vuelvo al drawer lateral normal con la misma tab activa
    When clickeo ✕ estando expandido
    Then el drawer se CIERRA y veo el mapa sin drawer          # cerrar ≠ colapsar

Feature: Tooltips doctrinales
  Scenario: Campo con valor explicado
    Given el Resumen de una caja con procedencia "estimado"
    When hago hover (o foco con teclado) sobre "procedencia"
    Then el tooltip explica el campo Y que "estimado" se dibuja atenuado (gris ≠ verde)

Feature: Resumen enriquecido
  Scenario: Navegar por el cableado
    Given la caja review-caja con necesita de "caja:edit-caja"
    When clickeo el chip "edited.md ← caja: edit-caja"
    Then el mapa selecciona el nodo edit-caja y el drawer muestra ese nodo
  Scenario: Hallazgo determinista
    Given una caja con gate.tipo "none"
    Then Hallazgos muestra el hallazgo A4 en tono crit — nunca "sin hallazgos"

Feature: Contenido
  Scenario: Ver la fuente real
    Given un nodo con fuente_path
    When abro la tab Contenido
    Then veo el archivo servido por el daemon (read-only, con números de línea)
  Scenario: Nodo sin fuente
    Given un nodo sin fuente_path
    Then la tab dice el estado honesto (pendiente del reconocedor) — jamás inventa

Feature: Corridas
  Scenario: Estado honesto
    Given cualquier nodo hoy (sin indexer)
    When abro la tab Corridas
    Then leo "Sin corridas indexadas" y por qué — jamás una lista vacía muda
```

## Trazabilidad RF → destino en la app

| RF | mockup:línea (v6) | destino | story=test |
|---|---|---|---|
| RF-80/81 | 524-525 · 552-568 | `widgets/map-canvas/ui/inspector.tsx` + página (estado expandido/cerrado) | expand/colapsa/cierra |
| RF-82/83 | 527-531 · 152-167 | inspector.tsx (tabs + CSS expandido) | tabs + sticky |
| RF-84 | 541-550 | inspector.tsx (prop sin selección) o página | vacío |
| RF-85–88 | 124-125 · 309-362 | `entities/arnes/model/doctrina.ts` (diccionario) + inspector.tsx | tooltip campo+valor |
| RF-89 | 364-371 · 509-512 | `entities/arnes/model/selectors.ts` (edges inversos) + inspector.tsx | viene-de |
| RF-90 | 373-378 · 489-499 | inspector.tsx (chips + onSelect del canvas) | chip navegable + cond |
| RF-91 | 379-388 · 515-517 | inspector.tsx + `shared/api` (conformance por nodo) | gate:none · sin-hallazgos |
| RF-92 | 501-508 | inspector.tsx (footer staged) | botonera disabled |
| RF-93–95 | 410-433 | inspector.tsx (tab) + endpoint Go `GET …/nodes/{id}/fuente` + arch (changelog OpenAPI) | fuente · sin-fuente · raw |
| RF-96 | 434-443 | inspector.tsx (tab) | corridas honesto |

## Cierre

Gate del paquete = PARIDAD.md fila por fila + click-through app vs mockup lado a lado
(consola limpia, screenshots). La firma de ESTE spec habilita la implementación.

---

## Enmienda 2026-07-07 (decisión #7 — supersede RF-84)

**RF-84 (estado vacío) queda SUPERSEDIDO:** el inspector ya no tiene estado vacío. Sin
selección NO se monta (la página lo guarda); `✕` limpia la selección y el drawer
desaparece entero, en drawer normal y en expandido. La línea de affordance «Clic en un
nodo del mapa…» muere. Ver `decisiones.md` #7.
