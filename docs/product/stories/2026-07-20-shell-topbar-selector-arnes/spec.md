# spec — Shell: Topbar sin empresa + selector de arnés (QUÉ)

> Paquete `2026-07-20-shell-topbar-selector-arnes` · base firmada: `decisiones.md` TS-D1..D17
> sobre `mockups/arnesia-shell-topbar-selector-arnes.html` (mockup FIRMADO 🧑‍⚖️ 2026-07-20, commit
> `5ca97c2`). Cada RF traza a `mockup:línea`. Cero cambios de backend (Go) — todo el dato ya existe
> vía `GET /api/portafolio` (Slice 0/1, HS-22/23); este paquete es 100% FE.

## Alcance

**Toca:** `web/src/widgets/topbar/ui/topbar.tsx` · `web/src/widgets/session-rail/ui/session-rail.tsx`
+ 2 archivos NUEVOS que el alcance firmado implica para no violar boundaries existentes (ver TS-D10/D17):
`web/src/widgets/session-rail/ui/new-session-picker.tsx` (componente props-puras, sin transporte) y
`web/src/widgets/session-rail/model/portafolio-picker-store.ts` (store Zustand, el único punto que
llama `api.listPortafolio`).

**No incluye** (no-goals): cambios de backend · rediseño de `SessionCard` · resolver
`workspace-stage.tsx:235` (efecto secundario documentado en TS-D16, backlog aparte) · wizard de
escaneo dentro del picker (si el portafolio está vacío, el picker deriva a la vista Portafolio, TS-D11)
· cambiar cómo se cierra/renombra/switchea una sesión existente (sin tocar).

## RF — Topbar

- **RF-1 (empresa fuera del breadcrumb).** `Topbar` deja de pintar `active.empresa` como primer
  segmento (`topbar.tsx:22`, hoy `alpacapurpura` fijo). El breadcrumb queda `arnés / vista`
  (`mockup:367-370` vs `355-360` del estado «Hoy»). [TS-D1]
- **RF-2 (chip de arnés → etiqueta plana).** El chip `{active.arnes} ▾` (`topbar.tsx:24-26`) pierde
  el `▾` y el estilo de control: pasa a etiqueta de solo lectura, borde punteado, sin `onClick`
  (`mockup:368`, clase `.chip-arnes.plain`). El arnés de una sesión es fijo de por vida — no se
  sugiere un cambio que no existe. [TS-D2]
- **RF-3 (Conversar a su propia línea).** El botón Conversar deja la fila del breadcrumb
  (`topbar.tsx:40-55`, hoy con `ml-auto` compitiendo por ancho) y baja a una segunda línea,
  alineado a la derecha (`mockup:365-374`, clase `.topbar.row2`). [TS-D3]
- **RF-4 (`quedaste en: …` sale de la Topbar).** El bloque condicional `active.parked`
  (`topbar.tsx:34-38`) se elimina — campo diseñado pero sin productor real en vivo (solo seed de
  `session_service.go:750-752`); mostrarlo sería fingir un dato que no existe. [TS-D4]

## RF — Rail: elegir arnés al abrir sesión

- **RF-5 (el picker vive en la ventana, ensancha el rail).** `NewSessionButton`
  (`session-rail.tsx:259-292`) deja de abrir `window.prompt()`; abre un panel INLINE dentro del
  propio `<aside>` de `SessionRail` — el `<aside>` crece de 224px a 360px con la misma
  `transition-[width]` que ya tiene (`session-rail.tsx:26-31`); mientras el picker está abierto, la
  lista de sesiones y el pie (nav global + tema) quedan ocultos. Cero `position:absolute`, cero
  backdrop, cero popover (`mockup:417-419`, adaptado — ver TS-D10 para el porqué del layout real
  distinto del mockup). [TS-D6, TS-D10]
- **RF-6 (buscador en vivo).** Input de búsqueda que filtra la lista por id, nombre y empresas en
  cada tecleo, sin debounce — es un filtro local sobre datos ya cargados (`mockup:424-427,653-664`).
  Contador de resultados no requerido (el mockup no lo tiene). [TS-D6]
- **RF-7 (fuente de datos = Portafolio real).** La lista sale de `GET /api/portafolio`
  (`api.listPortafolio<PortafolioListado>()`), vía `portafolio-picker-store.ts` — nunca la fixture
  de Storybook (`entities/arnes/testing/*`) ni un catálogo aparte. [TS-D7, TS-D17]
- **RF-8 (fila del picker — campos).** Cada fila reusa los componentes presentacionales YA
  construidos en `entities/portafolio` (cero reinvención): `EmblemaInicial` (inicial+color
  determinista) · `identificadorDe(identidad)` (id, con fallback `"(sin id)"`) · `entrada.nombre` ·
  chips de `entrada.empresas` (N:M, TODAS, nunca una sola) · presencia `◆ canónico vX.X` y/o
  `▣ N instalac.` · `DerivaChip` cuando hay una instalación `en-deriva` · `DotSaludPortafolio`
  (`saludDe(entrada)`) (`mockup:432-493`). [TS-D7]
- **RF-9 (colisión de id → chip distintivo).** Cuando `idsColisionados(entradas)` marca la
  `identidad.id` de una fila, se agrega un chip con `identidad.home ?? identidad.scope` (truncado)
  para diferenciarla de su gemela visual — no bloquea la selección (la fila sigue siendo por
  `clave`, única). [TS-D12]
- **RF-10 (selección en 2 niveles: identidad → copia).** Click en una fila con **1 copia total**
  (`(canonico?1:0)+instalaciones.length`) resuelve el path solo — no hay sub-lista. Con **2+
  copias**, se despliega `.pf-copies` con cada una (canónico incluido, tipo `canónico`
  rotulado; instalaciones con su `TipoInstalacionChip` + `install_path`/`proyecto_path` truncado +
  `DerivaChip` si aplica) — hay que elegir una copia antes de habilitar Crear
  (`mockup:449-476,596-627`). Con **0 copias**, la fila queda deshabilitada (sin radio), tooltip
  «sin copia local registrada» — no seleccionable. [TS-D8, TS-D15]
- **RF-11 (confirmación resuelta + Crear condicionado).** Tras resolver el path (simple o vía
  copia), aparece `usará <path>` en mono (`mockup:500,630-633`); el botón «Crear sesión» queda
  `disabled` hasta que haya `clave` Y `path` resueltos (`mockup:504-505,608-636`). [TS-D8]
- **RF-12 (Cancelar = cero efectos).** «Cancelar» cierra el picker, descarta búsqueda/selección y
  restaura el rail a su ancho normal — sin crear nada, sin re-fetch pendiente colgado
  (`mockup:503,595`).
- **RF-13 (estado vacío del portafolio).** 0 `entradas` → copy «Tu portafolio está vacío» + botón
  «Ir a Portafolio» que cierra el picker y navega `setView("portafolio")`. [TS-D11]
- **RF-14 (cargando / error al abrir).** Al abrir el picker: estado `cargando` (skeleton compacto,
  mismo espíritu que `Skeleton` de `portafolio-list.tsx`) mientras resuelve el fetch; estado `error`
  → mensaje + botón «Reintentar» que vuelve a llamar `cargar()`. [TS-D13]
- **RF-15 (refetch en cada apertura).** Cada apertura del picker dispara un `GET /api/portafolio`
  nuevo — no se reusa una lista cacheada de una apertura anterior de la misma corrida de la app.
  [TS-D14]
- **RF-16 (creación real — payload).** «Crear sesión» llama `useSessions().create({ arnes:
  identificadorDe(identidad), empresa: empresas.join(" · ") || undefined, salud: "info", view:
  "Mapa", path: <copia elegida> })` — sin `puesto` (sin dato real que escribir) y sin el
  `window.prompt()` de ruta libre, que se elimina entero (la ruta sale SIEMPRE de la copia
  elegida, nunca de texto tipeado). [TS-D5, TS-D8 (último bullet), TS-D16]
- **RF-17 (búsqueda y selección se resetean al cerrar/crear).** Cancelar y Crear-exitoso dejan el
  picker en su estado inicial (sin búsqueda, sin fila seleccionada) para la próxima apertura —
  mismo comportamiento que `resetPicker()` del mockup (`mockup:642-651`).

## Gherkin (aceptación por comportamiento)

```gherkin
Feature: Topbar sin empresa
  Scenario: Breadcrumb sin empresa, chip plano, Conversar en su línea
    Given una sesión activa cualquiera
    When se renderiza la Topbar
    Then el breadcrumb NO muestra ninguna empresa
    And el chip del arnés no tiene "▾" ni onClick
    And el botón Conversar está en su propia línea, no compite con el breadcrumb
    And no hay ningún texto "quedaste en:"

Feature: Elegir arnés al abrir sesión — caso simple
  Scenario: Una identidad con una sola copia
    Given el Portafolio tiene "cobranza-proveedores" con 1 canónico y 0 instalaciones
    When abro "＋ Nueva sesión" y clickeo esa fila
    Then el path queda resuelto sin sub-lista
    And "Crear sesión" está habilitado
    When clickeo "Crear sesión"
    Then se crea una sesión con arnes="cobranza-proveedores" y el path del canónico

Feature: Elegir arnés al abrir sesión — caso ambiguo
  Scenario: Una identidad con 2+ copias
    Given "luana-feature-cycle" tiene 2 instalaciones (una en-deriva)
    When clickeo esa fila
    Then se despliega la sub-lista de copias y "Crear sesión" sigue deshabilitado
    When elijo una copia
    Then aparece "usará <path>" y "Crear sesión" se habilita

Feature: Elegir arnés al abrir sesión — estados de datos
  Scenario: Portafolio vacío
    Given el Portafolio no tiene entradas
    When abro "＋ Nueva sesión"
    Then veo el copy de vacío y un botón "Ir a Portafolio"
  Scenario: El fetch falla
    Given GET /api/portafolio devuelve error
    When abro "＋ Nueva sesión"
    Then veo el mensaje de error y un botón "Reintentar"
  Scenario: Reapertura re-fetchea
    Given ya abrí y cerré el picker una vez
    When lo abro de nuevo
    Then se dispara un GET /api/portafolio nuevo, no la lista vieja en memoria

Feature: Buscador
  Scenario: Filtro en vivo
    Given el picker abierto con 3+ entradas
    When tipeo un id parcial
    Then solo quedan visibles las filas que matchean id/nombre/empresas
  Scenario: Sin resultados
    Given un texto que no matchea nada
    Then veo "Ningún arnés coincide" + acción para limpiar la búsqueda

Feature: Cancelar
  Scenario: Cero efectos
    Given una fila (o copia) seleccionada en el picker
    When clickeo "Cancelar"
    Then el picker se cierra, el rail vuelve a su ancho normal, y nada se creó
```

## Trazabilidad RF → destino en la app

| RF | mockup:línea | destino | story=test |
|---|---|---|---|
| RF-1..4 | 355-360 · 365-374 | `widgets/topbar/ui/topbar.tsx` | click-through PARIDAD (sin story: widget conectado a `useSessions`, mismo criterio ya vigente para este archivo) |
| RF-5 | 417-419 (adaptado, TS-D10) | `widgets/session-rail/ui/session-rail.tsx` (ancho + estado `pickerOpen`) | click-through PARIDAD |
| RF-6..11,13..15,17 | 424-500 · 596-664 | `widgets/session-rail/ui/new-session-picker.tsx` (props-puras) | `new-session-picker.stories.tsx`, `play()` por escenario (simple · ambiguo · vacío · error · colisión · buscar) |
| RF-7,15 | — | `widgets/session-rail/model/portafolio-picker-store.ts` | vitest unit (estado cargando→datos/error, refetch) |
| RF-12,17 | 503,595,642-651 | `new-session-picker.tsx` | `play()` cancelar-resetea |
| RF-16 | — | `session-rail.tsx` (wiring `onCrear` → `useSessions().create`) | click-through PARIDAD (transporte real de sesión ya cubierto por E2E existente) |

## Cierre

Gate del paquete = `PARIDAD.md` fila por fila + click-through app vs mockup lado a lado (consola
limpia, screenshots, ambos temas). La firma de ESTE spec + `design.md` habilita la implementación.
