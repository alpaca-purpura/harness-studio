# Decisiones — Portafolio · Slice 1 «FE Portafolio» (arquitectura del build)

> `tipo: decisiones` · paquete `2026-07-13-portafolio-slice1-fe` · deriva del modelo FIRMADO HS-22 +
> cimientos FIRMADOS HS-23 (`stories/2026-07-13-portafolio-slice0-cimientos/`). Decisiones TÉCNICAS del
> arquitecto (Fable 5, 2026-07-13) que cierran los huecos que la spec (§10) y la revisión adversaria
> (G1-G9) dejaron abiertos para el FE. El ejecutor (Sonnet 5) NO las relitiga; si una resulta inviable
> en código, documenta el porqué AQUÍ (nueva S1-D) y elige la alternativa más cercana al espíritu
> (honestidad > limpieza). **Todo se verificó contra el código REAL de Slice 0** (no contra la spec
> pre-build): `internal/{domain,ports,usecase}/portafolio.go` · `internal/adapters/portafolio/` ·
> `internal/adapters/transport/http/{portafolio,router,arneses}.go` · `web/src/` completo.

## Los 3 GAPS reales encontrados (flag al operador — mismo ejercicio que destapó S0-D12..D15)

**GAP-1 · «Abrir en Mapa» de una instalación `referenciada-cc` es IMPOSIBLE con el contrato vigente.**
La única vía de indexar un arnés al Mapa es `PUT /api/arneses/{id}` (registra cwd + indexa vía
`onArnesRegistered`), y `ArnesRegistry.checkProtected` (`internal/adapters/store/arnes_registry.go`)
**rechaza todo path bajo `~/.claude`** — exactamente donde vive el cache CC
(`~/.claude/plugins/cache/…`), que es la forma de instalación MÁS COMÚN real (S0-D1: «NINGÚN proyecto
tiene `.claude/plugins/<id>/` físico hoy»). Sin backend nuevo, el ítem «Observar (Abrir en Mapa)» del
alcance quedaría vacío en el caso dominante. → resuelto en **S1-D1** (endpoint de observación read-only).

**GAP-2 · La deuda del índice bare-id (S0-D6) se ACTIVA con este slice y NO se cierra acá.**
`index.Store.Upsert` keyea por `g.Arnes.ID` pelado: dos identidades del Portafolio con el mismo `id`
(p.ej. `harness` de dos homes distintos) se PISAN en silencio al observarlas — la violación exacta de
A2/C-ID-2 que la revisión adversaria marcó. El BACKLOG decía «re-key … cuando Abrir en Mapa lo exija»;
lo exige ahora, pero el re-key global (index + MapService + endpoints + picker + sesiones) es cirugía
de otro paquete. → resuelto acotado en **S1-D2** (colisión VISIBLE con confirmación; deuda sigue viva).

**GAP-3 · `EntradaPortafolio.Registries` existe pero NADA lo puebla, y el merge lo pisa.**
`usecase.AgregarProyecto` construye la entrada sin `Registries` aunque `origen.Registry` esté resuelto;
y `store.Upsert` hace `merged := e` (la entrada nueva REEMPLAZA facetas — un re-agregar con menos datos
borra `Registries`/`Empresas` previos). El facet «marketplaces» del drawer (spec §4) no tendría fuente.
→ resuelto en **S1-D3** (poblar al agregar + merge por unión).

**Notas menores verificadas (no bloquean, el plan las absorbe):** (a) `POST /api/portafolio/proyectos`
responde `[]EntradaPortafolio` SIN `clave` (el wire `entradaWire` solo existe en el GET) — el FE
re-fetchea el listado tras agregar, no parsea la respuesta del POST; (b) el escaneo es un request
síncrono — cancelable igual: el `ctx` del server muere cuando el cliente aborta (C-P-13 cubierto con
`AbortSignal`); (c) `GLOBAL_VIEWS`/`GlobalView` ya tienen la ruta `portafolio` staged como `ComingSoon`
— el swap es quirúrgico; (d) no existe infra de unit-test FE (vitest solo corre el proyecto
`storybook`) — ver S1-D7.

## S1-D1 · «Observar / Abrir en Mapa» = endpoint de observación READ-ONLY (cierra GAP-1)

`POST /api/portafolio/arneses/{clave}/mapa` con body `{"install_path": "..."}`:
- Busca la entrada por `clave` en el store del Portafolio (404 si no está — solo se observa lo
  PERSISTIDO, no candidatos).
- Valida que `install_path` sea una presencia REAL de esa entrada (∈ `instalaciones[].install_path` o
  `== canonico.path`, comparación canónica) — 400 si no pertenece (el endpoint jamás carga un dir
  arbitrario: el path viaja, pero la autoridad es el store).
- Carga el dir con el loader (read-only, mismo `ports.ArnesLoader`) y lo publica al índice del Mapa vía
  `ports.IndexPort.Upsert` (puerto YA existente). Responde `{"id": "<bare-id>", "indexed": true}`.
- **NO registra cwd** (`arneses.json` intacto): el registro de arnés→path es confinamiento de SESIÓN
  (otra cosa, otro boundary); observar es la semántica de la ley anti-drift — espejo read-only. La tab
  Contenido/fuente del inspector degrada honesto con el mensaje existente («sin directorio registrado»),
  y `checkProtected` deja de ser obstáculo porque acá no hay registro. Un dir no cargable → 400 con el
  motivo del loader, jamás un grafo inventado.
- `PortafolioService` gana el 5° puerto `indice ports.IndexPort` (cmd lo cablea con el `idx` que ya
  existe); el subcomando CLI pasa `nil` → `ObservarEnMapa` con índice nil = error honesto «la
  observación en Mapa requiere el daemon» (el CLI no la necesita).

## S1-D2 · Índice bare-id: NO se re-keyea en este slice; la colisión se hace VISIBLE (acota GAP-2)

El re-key del índice a identidad calificada sigue siendo deuda de BACKLOG (paquete propio — toca
MapService, endpoints, picker, sesiones). Lo que este slice SÍ hace: el FE computa las colisiones de
`identidad.id` entre entradas del Portafolio (`idsColisionados()` — selector puro con test) y, antes de
observar una entrada cuyo id colisiona, muestra **confirmación explícita**: «el Mapa de hoy keyea por id
pelado — abrir “X” acá re-apunta la vista de “X” de <otra clave>». Elección informada, jamás pisada
silenciosa (C-ID-2 en espíritu). La respuesta del endpoint devuelve el `id` efectivo para que el FE
apunte el Mapa a lo que REALMENTE quedó indexado.

## S1-D3 · `Registries` se puebla y el merge une facetas (cierra GAP-3)

- `AgregarProyecto`: si `c.Instalacion.Origen.Registry != ""` → `e.Registries =
  [CanonicalizarRepo(registry) | crudo si no parsea]` (el crudo visible es dato, no se descarta).
- `store.Upsert` (merge): `Registries` y `Empresas` = **unión dedup** (existente ∪ nueva); `Nombre`/
  `Descripcion` siguen last-write-wins (el escaneo fresco manda); un upsert con faceta vacía NO borra la
  existente. Test dedicado.
- El drawer pinta el facet «marketplaces» desde `registries` de la entrada ∪ `origen.registry` de cada
  instalación (selector puro `registriesDe()`) — si el conjunto queda vacío → chip `desconocido` (G4/BR-3).

## S1-D4 · Regla de salud del dot (cierra G9 — «definir regla o quitar»: se define)

Selector puro **`saludDe(entrada): "ok" | "atencion" | "sin-senal"`** con unit test (G9 exige regla
TESTEADA, no vibes):
- `atencion` ⟺ ∃ instalación con `deriva == "en-deriva"` ∨ `aviso != ""` ∨ `origen.discrepancias`
  no vacío.
- `ok` ⟺ hay ≥1 instalación y TODAS están `al-hilo` sin aviso ni discrepancias.
- `sin-senal` ⟺ el resto (todo `deriva-no-evaluable`, o 0 instalaciones): **«no sé» ≠ «sano»** — dot
  hueco/muted con `aria-label="salud: sin señal"`, JAMÁS verde fabricado.
- Sin `crit` en S1: ningún dato del backend lo computa hoy — no se inventa un rojo. Vocabulario PROPIO
  del Portafolio (`SaludPortafolio`), NO se reusa el `Salud` de sesión (ok/warn/crit/info = otro
  significado, «no pisar vocabulario»).

## S1-D5 · Update honesto (cierra G2)

La FILA no pinta ningún flag de update (no hay dato — el flag `⬆ vX` del mockup muere). El DRAWER
muestra una línea fija por entrada: «update: **no-verificado**» con tooltip «update-check llega en
Slice 4» (BR-8 + RN-UPD-1). Nada parpadea, nada promete.

## S1-D6 · Placement FSD-lite + estilos (verificado contra boundaries enforced)

- **`entities/portafolio/`** — `model/types.ts` (espejo EXACTO del wire de Slice 0) ·
  `model/selectors.ts` (puros) · `testing/entradas.ts` (fixtures calcadas de salidas REALES del E2E de
  Slice 0) · `ui/` (chips/dot de dominio, presentacionales).
- **`widgets/portafolio/ui/`** — `portafolio-list.tsx` · `portafolio-drawer.tsx` ·
  `portafolio-wizard.tsx` (+ sus `.stories.tsx`). Widgets = props puras, CERO transporte
  (`fe-transporte-independiente`; dependency-cruiser lo enforcea).
- **`pages/shell/ui/portafolio-view.tsx`** — composition-root con TODO el transporte (patrón
  `AjustesView` de RF-100, mismo archivo-hermano en pages/shell — pages no se importan entre sí).
  `GlobalView` rutea `portafolio` → `PortafolioView` (muere el `ComingSoon`).
- **`shared/api/client.ts`** — 5 métodos nuevos GENÉRICOS (`<T>`, domain-free como `getGraph`):
  `listPortafolio` · `escanearProyecto(path, signal?)` · `agregarProyecto(path, elegidos)` ·
  `desvincularDelPortafolio(clave)` · `observarEnMapa(clave, installPath)`.
- **Estilos:** `app/styles/portafolio.css` scoped `.arnesia-portafolio` (importado desde `index.css`),
  SOLO tokens DTCG (stylelint strict-value ya lo rompe si no) + `@media (prefers-reduced-motion:
  reduce)` (G8). Primitivos: los existentes (`Button`, Base UI copy-in si falta Dialog/Checkbox) — cero
  entradas nuevas en `package.json`.

## S1-D7 · Infra de test FE: proyecto vitest `unit` (node) para los selectores

Hoy `vitest.config.ts` solo tiene el proyecto `storybook` (story = test, browser). Los selectores G9/G4/
G6 necesitan tests de función pura sin levantar Chromium → se agrega el proyecto **`unit`** (environment
node, include `src/**/*.test.ts`) + paso de CI `pnpm exec vitest --project=unit run` en el job `ts`.
Las stories SIGUEN siendo el fitness visual (fe-visual-fitness intacto); esto lo complementa, no lo
reemplaza. Ataja además la deuda registrada «parte de FE sin tests (solo stories)».

## S1-D8 · Toolbar mínima honesta (acota spec-usabilidad §2 al alcance del BACKLOG)

Construido: **buscar** (substring sobre id/nombre/descripción) + **lentes `empresa` (default) y
`plano`**. La lente empresa agrupa por `empresas[]` (N:M: una entrada aparece en N grupos) con grupo
**«sin empresa»** obligatorio al final — la realidad medida es que la mayoría de instalaciones CC no
declaran empresa (C-N-15/G4: jamás inferirla del path). Lentes `proyecto`/`marketplace` y filtros
`estado`/`marketplace` del mockup = **disabled + tooltip «próximo»** (BR-8) — no tienen slice asignado,
no se les inventa uno.

## S1-D9 · Wizard rama Proyecto: fuente y cancelación

- Input de path de texto SIEMPRE (funciona en browser y desktop, S-D2) + botón «Elegir carpeta…» SOLO
  dentro de Tauri (`isTauri()` + `@tauri-apps/plugin-dialog`, patrón exacto de RF-110 en `AjustesView`).
- Radio «Repositorio GitHub» y tab «Marketplace» = **disabled + tooltip «próximo · S2»**; muere la
  validación ✓ hardcodeada del mockup (G3) — la rama disabled no simula NINGÚN resultado.
- Escaneo cancelable: `AbortSignal` en el fetch (el `ctx` del server se cancela con la desconexión —
  C-P-13 sin backend nuevo). Cancelar/cerrar el wizard = CERO efectos (no persiste nada; el POST
  /escaneos ya no persiste por diseño).

## S1-D10 · Candidatos del escaneo: ya-presentes y canónicos

- Candidato cuya `clave` ya existe en el Portafolio → badge **«ya en el portafolio»**, checkbox
  HABILITADO (re-agregar = re-escanear idempotente, C-P-8 — actualiza instalaciones, no duplica).
- `es_canonico: true` (RN-IDENT-4) → el candidato se pinta como **CANÓNICO** con copy propio («checkout
  editable — se registra como canónico, no como instalación»), no como espejo.
- Cada candidato muestra: id (mono) + nombre · `v<version>` o **`v?`** · registry resuelto o **«origen
  desconocido»** · tipo de instalación · chip deriva · chip aviso si lo hay. Todo del wire, nada local.

## S1-D11 · Tooltips de lo diferido: Reparar y Backport = S5 (resuelve H2 para la UI)

`spec-usabilidad.md` §4 decía «Backport (S2)»; el BACKLOG re-cortado (post-HS-23) los agrupa en el item
**5. Reparar + Backport**. El BACKLOG manda (es el estado vivo): ambos botones disabled + tooltip
«próximo · S5». «Mejorar (chat)» y «Traer canónico» = S2; «Publicar» = S3; update = S4 (S1-D5).

## S1-D12 · El mockup se corrige EN SITIO y el SSoT pasa a Storybook

El alcance del BACKLOG dice «fix mockup G1-G9» textual: `mockups/arnesia-portafolio.html` se corrige en
T8 (fuera los datos fabricados: `drift:N`, `⬆ v1.2.0`, empresa-del-path «Vitalia», validación ✓ del
marketplace; entra el vocabulario real: `deriva`, `no-verificado`, keys por clave) y su fila en
`mockups/INDEX.md` se re-estampa: estado → «✅ construido (Slice 1) — SSoT = stories `portafolio-*`;
snapshot corregido @<commit>». NO se crea un `.html` nuevo: el baseline vigente del Portafolio queda en
Storybook (regla 1 del INDEX), el html es snapshot derivado y fechado.

## S1-D13 · «Abrir en Mapa» necesita una sesión activa — disabled honesto sin ella

El Mapa vive en el stage de sesión (`WorkspaceStage`; la ruta global `portafolio` es otra superficie).
Flujo con sesión activa: confirmar colisión si aplica (S1-D2) → `POST …/mapa` → `setMapaPeek(id)` en
`app-store` (campo nuevo) → `parkView("Mapa")` + `setView(<ruta no-global>)`; `WorkspaceStage` consume
el peek (re-fetchea el listado de harnesses, apunta `viewedId`, limpia el peek). Sin sesión activa: el
botón queda **disabled + tooltip** «necesita una sesión abierta — el Mapa vive en el stage de sesión»
(no se auto-crea una sesión: crear sesión lanza un conductor, efecto demasiado grande para un click de
observación).

## S1-D14 · Vocabulario en la UI: «deriva» firmado; «origen» solo como «origen de la copia»

Etiquetas: `al-hilo` / `en-deriva` / `deriva-no-evaluable` tal cual (F-C firmado). La sección del drawer
que pinta `OrigenPortafolio` se rotula **«origen de la copia»** (registry · versión · discrepancias ·
eslabones) para no chocar con `origen` L0 (estandar/del-puesto, del provisioner) — misma distinción que
el backend hizo con `OrigenPortafolio` vs `Origen` (S0-D12). `procedencia`, `canal`, `insumos`, `banda`
NO se usan para conceptos nuevos (regla 4 de `mockups/INDEX.md`). La trazabilidad BR-3 se expone: los
`eslabones[]` crudos viven en un `<details>` colapsable «trazabilidad» dentro de cada instalación.

## S1-D15 · Frescura del dato: lo persistido es una FOTO del último escaneo (no se finge vivo)

`GET /api/portafolio` devuelve lo que se persistió al agregar — deriva/avisos NO se re-evalúan en cada
listado. El drawer lo DICE: línea muted «datos del último escaneo — agregado <fecha `agregado`>» +
la vía de refresco es re-agregar el proyecto (wizard, C-P-8 idempotente). No se construye auto-refresh
ni re-evaluación en background en S1 (sería otra superficie de backend); no se maquilla la foto como
tiempo-real. El estado `no-encontrada` (C-P-9) aparece cuando un RE-escaneo lo detecta, no por magia.

## S1-D16 · Capability YAML: el campo `status:` POR-SCENARIO del template rompe el enforcer real (T1)

Hallazgo de build (Sonnet 5, T1): `docs/product/_templates/capability.template.yaml` documenta cada
scenario BDD con su propio `status: live|wip|deprecated` anidado. Pero
`docs/architecture/fitness/capability_trace_test.go#capMetas` (el enforcer R4 real, `TestCapabilityStatusConsistent`)
NO parsea YAML: escanea línea por línea y toma como «el status de la capability» CUALQUIER línea cuyo
trim empiece con `status:`, sin mirar indentación/anidamiento — así que el último `status: live` de un
scenario le PISA el `status: vivo` real del root y el check falla («status "live" fuera de enum
{vivo,vivo·nc,parcial,stub}»). Confirmado: ningún capability YAML existente en el repo usa hoy la forma
BDD completa (`id/name/status/given/when/then/verifica`) con scenarios NO vacíos — los 2 que tienen
`scenarios:` poblado (`fe-shell/boton-actualizar.yaml`, `fe-shell/rail-de-sesiones.yaml`,
`self-update/self-update-sin-sudo.yaml`) usan la forma LEGACY (lista plana de strings, sin campo
`status` anidado), y los 5 de `portafolio/` creados en Slice 0 dejan `scenarios: []`. `scripts/cap_doctor.py`
(el otro validador) SÍ usa `yaml.safe_load` de verdad y no tiene este bug — la inconsistencia es SOLO
del arch-test Go. → **Decisión:** `observar-en-mapa.yaml` (T1) usa la forma BDD completa (id/name/
given/when/then/verifica) pero **sin el campo `status:` por-scenario** — la única forma de tener
scenarios BDD reales Y pasar el gate `go test ./docs/architecture/fitness/...` hoy. NO se tocó
`capability_trace_test.go` (fuera del alcance de T1 — es un enforcer compartido, arreglarlo bien pide
YAML real y una pasada sobre las 82 hojas existentes). Deuda para BACKLOG: `capMetas` debería parsear
YAML de verdad (o al menos ignorar líneas con indentación > 0) antes de que alguien más pise esta misma
trampa con scenarios BDD poblados.

## S1-D17 · steiger `fsd/inconsistent-naming` es un falso positivo EN sobre nombres ES (T2)

Hallazgo de build (Sonnet 5, T2): al crear `entities/portafolio/` (segundo slice bajo `entities/`,
junto a `entities/arnes/`), `pnpm exec steiger src` empezó a fallar con «Inconsistent pluralization of
slice names. Prefer all plural names», auto-fix propuesto: renombrar el directorio `portafolio` →
`portafolios`. Investigado en el código del plugin (`@feature-sliced/steiger-plugin` 0.6.0,
`inconsistent-naming` check): usa una librería de pluralización **en inglés** sobre los nombres de
slice; "arnes" termina en "s" así que la heurística EN lo clasifica "plural", "portafolio" no termina
en "s" así que lo clasifica "singular" — el checker exige que TODOS los slices de una capa compartan
la misma clasificación. Es un artefacto de que el dominio está en **español**, no una violación FSD
real (ambos nombres son sustantivos singulares en español). Renombrar `entities/arnes` (consumido en
decenas de imports ya mergeados) está fuera de alcance de T2, y cualquier slice nuevo en español volvería
a chocar con la misma heurística. → **Decisión:** `web/steiger.config.ts` apaga
`fsd/inconsistent-naming` globalmente (mismo patrón que el apagado ya existente de
`fsd/insignificant-slice`, con razón documentada inline) — steiger es gate secundario/BETA
(`enforced_by` primario = dependency-cruiser, según el propio comentario de cabecera del config); las
demás reglas de steiger (public-api, no-cross-imports, layer-direction) siguen activas y enforced.
