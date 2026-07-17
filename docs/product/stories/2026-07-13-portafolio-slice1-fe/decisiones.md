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

## S1-D18 · Drawer modal (G8): primitivo de a11y propio, NO `@base-ui-components/react/dialog` (T5)

Hallazgo de build (Sonnet 5, T5): el repo no tenía ningún dialog/drawer existente con
`role="dialog"`/`aria-modal`/trap-de-Tab/Esc que reusar (se buscó en `widgets/map-canvas/ui/
inspector.tsx` — el drawer del Mapa, que NO es modal: Esc solo colapsa, sin trap, sin
`role="dialog"`). `@base-ui-components/react` YA es dependencia (`package.json`, sin uso real
todavía) y trae `Dialog` (`@base-ui-components/react/dialog`, `Root/Portal/Popup/Title/Close`)
con trap de foco + dismiss nativo — la opción evidente por "no reinventar el primitivo si ya
existe uno compartido" (instrucción del ticket). Probado: `Dialog.Popup` exige un
`<Dialog.Portal>` ancestro (`useDialogPortalContext` tira `Error` sin él) y `DialogPortal`, por
`default`, monta su contenido vía `FloatingPortal` en `document.body` — **fuera** del árbol que
`within(canvasElement)` recorre en TODAS las stories de este repo (patrón fe-visual-fitness
ya usado en `portafolio-list.stories.tsx` y todo el resto del Storybook). Pasar un `container`
propio a `Dialog.Portal` habría exigido colar una prop de "dónde portalar" a través del
contrato puro `PortafolioDrawerProps` (§2.6, cerrado) — un detalle de test filtrándose al
contrato de dominio, lo que el plan prohíbe (§P.1: no relitigar contratos cerrados). Verificado
además que Base UI tampoco setea `aria-modal` por su cuenta (`DialogPopup.js`/`useRole` de
`floating-ui-react`: solo agregan `role`, nunca `aria-modal`) — habría que agregarlo a mano de
todos modos. → **Decisión:** primitivo propio y chico en `web/src/shared/lib/focus-trap.ts`
(`focusablesEn` + `trapTabKeyDown`, sin dependencias nuevas — reusa lo que ya hay: React +
DOM), documentado inline con el porqué, aplicado en `portafolio-drawer.tsx` (`role="dialog"`
`aria-modal="true"` `aria-labelledby` propios, foco inicial al botón cerrar vía `useEffect`,
Esc vía `onKeyDown` local). Reusable tal cual por el wizard (T6, mismos requisitos G8) sin
tocar este archivo. `@base-ui-components/react` sigue como dependencia intacta (no se quitó del
`package.json`, NO deps nuevas — §P.4); simplemente no se usó su `Dialog` compound para esta
superficie por el conflicto estructural con el patrón de test del repo.

## S1-D19 · Wizard: Esc/✕ NO cierran mientras `estado==="agregando"` (criterio delegado, T6)

El ticket T6 delegó explícitamente el criterio: «Esc llama `onClose` en cualquier paso (salvo
mientras `estado==="agregando"` si te parece más seguro no cerrar a mitad de un POST — usá tu
criterio y documentalo si te desviás)». Decisión: **sí se bloquea** — mientras `estado===
"agregando"` el `POST /api/portafolio/proyectos` de la página (T7) está en vuelo; cerrar el
wizard en ese instante no aborta el POST (a diferencia del escaneo, que sí tiene
`onCancelarEscaneo`/`AbortSignal`, S1-D9) — solo desmontaría la UI mientras el request sigue
vivo en el server, dejando al usuario sin feedback del resultado. Mismo espíritu que
`UpdateCard` (`features/self-update`), que tampoco permite cancelar a mitad de un self-update.
Implementación: el botón ✕ queda `disabled` + `title="agregando en curso — esperá a que
termine"`, y el handler de `Escape` no llama `onClose` en ese estado (`bloqueadoParaCierre`,
`portafolio-wizard.tsx`). Cubierto por la story `Agregando` (assert de `disabled`+`title`+Esc
sin efecto) — la story `A11yModal` prueba el camino normal (Esc SÍ cierra) en `estado:"fuente"`.
Cancelar/cerrar en cualquier OTRO paso sigue siendo CERO efectos (S1-D9 intacto).

## S1-D20 · Confirm de colisión de id: dialog propio EN LA PÁGINA, no un cambio de contrato (T7)

Hallazgo de build (Sonnet 5, T7): el plan (§2.7) pide "confirmación de colisión (S1-D2:
`idsColisionados` + dialog propio, mismo primitivo del confirm de desvincular)" ANTES de
`onObservar`. Pero `PortafolioDrawerProps` (§2.6, cerrado en T5) no tiene ningún slot de
confirmación intermedia — el prop es `onObservar?: (installPath: string) => void`, directo. Los
3 widgets son props puras y su contrato está cerrado (§P.1: no relitigar); tocarlo para sumar un
paso de confirmación habría sido exactamente el tipo de "detalle de test/flujo filtrándose al
contrato de dominio" que S1-D18 ya rechazó por otro motivo. → **Decisión:** el confirm de
colisión vive ENTERAMENTE en `pages/shell/ui/portafolio-view.tsx` (`ColisionConfirmDialog`,
componente local del archivo): la página intercepta el click ANTES de que `onObservar` exista
como tal — su propio `onObservarInstalacion` corre `idsColisionados(entradas)` primero y, si hay
colisión, guarda `{clave, installPath, idMostrado, otraClave}` en estado local y renderiza el
dialog en vez de llamar al backend; `Confirmar` dispara `ejecutarObservar` (el mismo POST que
habría corrido directo). El dialog reusa el vocabulario visual de `pf-drawer`/`pf-btn-*`
(`portafolio.css`, T5) y el primitivo de foco `trapTabKeyDown` (`shared/lib/focus-trap.ts`,
S1-D18) — mismo patrón, cero widget tocado, cero prop nueva en el contrato cerrado. Texto exacto
de S1-D2: «el Mapa de hoy keyea por id pelado — abrir "X" acá re-apunta la vista de "X" de
`<otra clave>`».

## S1-D21 · Error de `onObservar`: banner propio de la página (T7)

`onObservar?: (installPath: string) => void` (§2.6) es fire-and-forget — el widget no espera
ningún resultado ni tiene un slot de error (a diferencia de `desvincularError`, que sí existe
para el DELETE). `POST /api/portafolio/arneses/{clave}/mapa` (S1-D1) puede fallar en la vida
real (400 install_path ajeno tras un re-escaneo que movió la instalación, 404 si la entrada se
desvincula en otra pestaña, 500 si el loader no puede leer el dir) y silenciar ese fallo sería
fabricar un éxito de palabra (BR-8). → **Decisión:** la página guarda `observarError` en estado
propio y pinta un `<p role="alert" className="pf-error">` FUERA del `<PortafolioDrawer>` (por
encima, mismo overlay) cuando lo hay — cero prop nueva en el contrato cerrado, mismo espíritu que
S1-D20 (el contrato no se toca; la página resuelve por fuera). Se limpia al reintentar observar o
al abrir otra fila.

## S1-D22 · Error de `agregarProyecto` (POST /proyectos): reusa el slot `error` del wizard (T7)

El contrato del wizard (§2.6) trae un único slot textual `error?: string` documentado para "el
motivo 400 del backend" del ESCANEO (G5, `ErrorDePath`) — no hay un slot separado para un fallo
del POST final de agregar (`estado==="agregando"` solo pinta el botón bloqueado, sin rama de
error, plan §3 T6: "sin checklist inventada"). Un fallo real ahí (path movido entre el escaneo y
el click de Agregar, C-P-9 en juego) no puede fabricarse como éxito. → **Decisión:** la página
reusa el mismo slot `error` + vuelve `estado` a `"candidatos"` — el usuario ve el motivo real y
puede reintentar desde el mismo formulario de re-escaneo que ya existe para el error de escaneo
(`PasoCandidatos` con `error` truthy). Costo aceptado: la checklist de candidatos deja de
mostrarse mientras el error está visible (el componente widget no distingue "error de escaneo" de
"error de agregar" — mismo slot, mismo render). No se tocó el widget: es la opción más honesta
sin ampliar el contrato cerrado de T6.

## S1-D24 · Auditoría UX del paso Fuente (skill frontend-design) — 4 fixes aplicados (2026-07-15)

Post S1-D23, el operador pidió auditar `Paso1Fuente` con la skill `frontend-design`, enfoque
usabilidad/intuitividad. Verificado en vivo (Storybook `:6006`, screenshot real, `getComputedStyle`),
no a ojo — 4 hallazgos reales, los 4 corregidos y re-verificados (`pnpm run verify` + `vitest`
127/127 + 21/21):

1. **Los 2 ButtonGroup (fuente/modo) se leían como el mismo control repetido** — sin nada que marque
   que son ejes distintos. Fix: label `.pf-faceta-label` (reusado del drawer, no una clase nueva)
   arriba de cada grupo: «ORIGEN» / «CÓMO CARGAR LA RUTA».
2. **`accent-color: auto` en los radios/checkboxes nativos** — el navegador pintaba el punto en su
   azul default, chocando contra el pill teal de fondo. Fix: regla módulo-wide
   `.arnesia-portafolio input[type="radio"|"checkbox"] { accent-color: var(--primary) }` — cubre las
   6 instancias del módulo (wizard + drawer), cero por-componente que perseguir a futuro.
3. **Layout se rompía en modo "elegir"**: input+trigger+Escanear todos en `.pf-wizard-path-row`
   hacía que Escanear saltara de línea sin control cuando el contenido no entraba. Fix: Escanear sale
   de esa fila, vive como hermano directo en `.pf-wizard-fuente` (fila propia siempre, `align-self:
   flex-start` para no stretchear full-width por el flex-column padre).
4. **Botón cerrar `×` (28px, borde `--border`) fácil de perder** — esquina lejana, bajo contraste.
   Fix: 32px, borde `--input` (más contraste), `font-size` del glifo subido, hover feedback nuevo.

Verificado con screenshot real ANTES/DESPUÉS de los 4 (light+dark), incl. el caso "elegir" que tenía
el bug de layout — confirmado resuelto, Escanear queda en su propia fila sin romperse.

**Reubicación (mismo turno, "aplicá donde corresponda"):** el fix #2 (`accent-color`) se movió de
`portafolio.css` (scoped `.arnesia-portafolio`) a `index.css` `@layer base` (global) — es
correctitud de nivel token («todo radio/checkbox nativo usa `--primary`, no el azul del navegador»),
no algo específico del Portafolio; dejarlo scoped hubiera significado repetirlo por módulo cada vez
que otra superficie sume un checkbox. Verificado por grep: HOY ningún otro módulo (Mapa/Ajustes/Chat)
tiene `<input type="radio"|"checkbox">` ni un botón «Cerrar» propio — los fixes #1/#3/#4 (labels,
layout de Escanear, botón cerrar) quedan correctamente scoped a Portafolio, nada más que propagar.
Re-verificado tras el move: `pnpm run verify` + `vitest` 127/127 + 21/21.

## S1-D23 · Wizard Paso Fuente: modo Escribir/Elegir explícito + ButtonGroup — SUPERSEDE S1-D9 (2026-07-15)

Amendment post-review: el operador revisó el build de T1-T8 en vivo (sin firmar el gate aún) y pidió 2
cambios sobre el paso 1 del wizard, antes de firmar. Documentado acá porque cambia comportamiento ya
descrito en S1-D9 — S1-D9 queda **superseded**, no borrada (regla de continuidad del paquete).

**Pedido del operador (2 partes):**
1. En vez de un input SIEMPRE editable + botón «Elegir carpeta…» condicional, el usuario debe poder
   **elegir explícitamente** entre tipear la ruta o abrir el picker nativo — con la ruta resuelta
   mostrada arriba, solo-lectura, y «Escanear» deshabilitado hasta que haya una ruta cargada.
2. El selector «Carpeta local / Repositorio GitHub» (hoy radios pelados) debe leer como un
   **ButtonGroup** — mismo criterio aplica al nuevo selector de modo.

**Diseño resuelto:**

- **Nivel 1 (fuente, sin cambio de markup)** — sigue siendo `<input type="radio">` nativo
  (`role="radiogroup"` implícito), NO se migra a `@base-ui-components/react/toggle-group`: el repo
  usaba de fábrica el patrón «radio + CSS `:has(input:checked)`» (`pf-wizard-radio:has(input:disabled)`
  ya existía, línea 634 de `portafolio.css` antes de este cambio) — reusarlo es menos riesgo que sumar
  un primitivo nuevo, y semánticamente un radiogroup («elegí exactamente una fuente, siempre una
  elegida») es más correcto que un ToggleGroup tipo-toolbar para este caso. Solo cambia el CSS: de
  «radio pelado + label» a segmentos con borde compartido, fondo resaltado en el `:checked` — el
  ButtonGroup pedido es 100% CSS, cero cambio de estructura ni de tests existentes de rol (`getByRole
  ("radio", {name})` sigue funcionando igual).
- **Nivel 2 (modo, NUEVO)** — mismo patrón exacto: radiogroup `Escribir ruta` / `Elegir carpeta`,
  estado LOCAL del widget (`useState<ModoFuente>("escribir")` en `PortafolioWizard`, mismo criterio que
  `path`/`elegidos` ya lifted — S1-D9 original). Default siempre `"escribir"` (funciona con y sin
  Tauri). `Elegir carpeta` disabled+tooltip **`"solo disponible en la app de escritorio"`** cuando
  `onElegirCarpeta` es `undefined` (fuera de Tauri) — mismo criterio «disabled+tooltip, jamás oculto en
  silencio» que ya rige GitHub/Marketplace (G3), pero tooltip DISTINTO de `TOOLTIP_S2` («próximo · S2»)
  porque esto no es un roadmap diferido, es un límite de plataforma permanente (File System Access API
  del browser nunca expone la ruta absoluta real del OS, por diseño de sandboxing — Tauri sí, porque
  corre con privilegios nativos fuera del sandbox del navegador; explicación técnica dada al operador en
  el mismo turno).
- **El input es UN SOLO elemento siempre presente** (no se swap-ea por un `<span>`): en modo `escribir`
  es editable; en modo `elegir` pasa a `readOnly` (sigue con `role="textbox"`, sigue anunciado por
  lectores de pantalla, sigue en el DOM en la misma posición — cero breaking change de selector en los
  tests existentes que hacen `getByRole("textbox", {name: "Ruta del proyecto"})`). Placeholder cambia
  según modo: `"ninguna carpeta elegida todavía"` en `elegir` sin resolver aún.
- **El botón «Elegir carpeta…» (trigger) solo se renderiza en modo `elegir`** (antes: condicionado solo
  a que `onElegirCarpeta` existiera). **Deliberadamente NO se auto-dispara el picker al seleccionar el
  radio** — evaluado y descartado: un radio que lanza un diálogo nativo del SO al ser seleccionado (incl.
  por navegación de flechas de un lector de pantalla) es una sorpresa de foco/acción no pedida, mal
  patrón de a11y. En cambio: seleccionar el radio revela el botón trigger (siempre clickeable, reintentos
  después de cancelar el picker del SO sin fricción — un radio ya-marcado no re-dispara `onChange` en un
  segundo click, por eso el trigger es un botón aparte, no el propio radio).
- **`Escanear` deshabilitado hasta `path.trim() !== ""`** — antes solo dependía del flag `disabled`
  (form-wide, ligado a `estado==="escaneando"`). Ahora: `disabled || path.trim() === ""`, en los 3 sitios
  donde `PasoFuente` se renderiza (paso 1, embebido en escaneando-disabled, retry en candidatos
  vacío/error).

**Impacto en tests existentes (S1, `portafolio-wizard.stories.tsx`):**
- `Paso1Fuente` — extendida (no rota): agrega assert de `Escanear` disabled ANTES de tipear + assert del
  radio `Elegir carpeta` disabled+tooltip nuevo.
- `Paso1FuenteConElegirCarpeta` — el flujo cambia de 1 paso (click directo en «Elegir carpeta…») a 2
  (seleccionar radio `Elegir carpeta` → click en el botón trigger que aparece) — reescrita.
- `A11yModal` — el último focuseable del trap deja de ser `Escanear` (ahora arranca disabled sin ruta) y
  pasa a ser el input (el nuevo radio `Escribir ruta` se suma al medio del loop, cuenta neta de
  focuseables NO cambia: 5). Reescrita la sección de trap hacia atrás.
- `SinHallazgos`/`ErrorDePath`/`CandidatosReales`/`Escaneando`/`Agregando` — sin cambios de fondo (no
  interactúan con el picker ni dependen del último-focuseable).

**Deuda diferida (no bloquea, registrada acá):** `mockups/arnesia-portafolio.html` línea ~300 (radio
pelado «Carpeta local») queda SIN corregir en este turno — S1-D12 ya estableció que el SSoT es Storybook
y el mockup es snapshot derivado corregido solo en cierres de ticket (T8 lo hizo); este cambio es un
amendment pre-firma dentro del mismo slice, no un ticket nuevo. Si el gate se firma antes de tocar el
mockup, queda como ítem de BACKLOG.

## S1-D25 · Homologación del término «arnés» (fase cero de lenguaje, operador 2026-07-16)

Observación del operador pre-firma: antes de seguir iterando, fijar QUÉ llamamos arnés y cómo se
relaciona con el estándar de la industria. Resolución conversada y firmada de palabra en el turno:

- **En la industria (2025-2026), «harness / agent harness» = el runtime que envuelve al modelo**
  (loop de contexto, ejecución de tools, feed de resultados) — Claude Code ES el harness en ese
  vocabulario; «harness engineering» = construir ese runtime. Lo que NOSOTROS empaquetamos encima
  de CC (skills, agents, commands, hooks, CLAUDE.md, MCP, docs, arquitectura as-code por
  rol×proceso) la industria lo llama «plugin» (mecanismo oficial CC) / «skills pack» / «agent
  configuration» — NO existe un término estándar único para el paquete completo.
- **Decisión: «arnés» se mantiene como término de PRODUCTO** (marca: ArnesIA, Prenter Harness) con
  la definición que ya fija `docs/architecture/contracts/nomenclatura-arnes.md` §1 (FIRMADA v1.1):
  paquete de know-how por rol×proceso, distribuido como plugin CC, manifiesto `arnes.l0.json`.
  Regla interna de conversación: arnés = el paquete (la carga); harness (industria) = CC (el
  vehículo). Hacia afuera, «Claude Code plugin» es la traducción que el mundo entiende.
- La definición del operador («todo lo que se instala encima de Claude Code para que el usuario
  trabaje y cumpla sus objetivos») coincide con la doctrina firmada — no hay corrección de fondo,
  solo el deslinde explícito contra el uso industrial del término.

## S1-D26 · Wizard identificable: cadena id→scope, aviso con id, worktrees, agrupación monorepo (operador 2026-07-16)

Observación del operador probando en vivo (pre-firma del gate): agregar `~/Proyectos/luana-vitalia`
(monorepo + worktree git de `luana-platform`) mostraba **11 tarjetas «(sin id)» indistinguibles** +
2 avisos anónimos «enabledPlugins declara X pero sin record». Diagnóstico verificado contra el
código y el registro CC real; 4 fixes con go explícito del operador:

1. **FE — cadena de identificación `id → scope → «(sin id)»`** (`identificadorDe()`, selector puro
   con unit test): la identidad provisional de un hallazgo sin manifiesto ES su scope (RN-IDENT-2 —
   ruta relativa al proyecto o remote), el backend YA lo mandaba en el wire y el FE lo tiraba.
   Aplicado en wizard, lista, drawer y confirm de colisión de la página. Candidato sin id lleva
   chip «sin manifiesto». `TipoInstalacionChip` con `tipo:""` (hallazgo sin forma física, C-P-14)
   ahora no pinta nada — antes: chip vacío; `types.ts` refleja que `tipo:""` es wire real.
2. **Scanner — el aviso sin-record conserva el id** (`IDConocido` = parte-id de la clave
   `enabledPlugins`, misma semántica que la fila del lock DevStudio — la rama estaba inconsistente):
   la tarjeta-aviso muestra «commit-commands», no «(sin id)». Test `TestScannerCCSinRecordConservaID`.
3. **Scanner — cruce CC resuelve worktrees** (`gitCommonDir()`: mecanismo canónico
   `<gitdir>/commondir`, no heurística de nombres): CC keyea el install-record por el path EXACTO
   del proyecto, pero un worktree linkeado comparte el `settings.json` versionado — record de otro
   working tree del MISMO repo ahora resuelve, con **aviso visible** «record de instalación de
   <path> (worktree del mismo repo)», jamás en silencio. Era la causa real de los 2 avisos de
   luana-vitalia (records bajo `luana-platform`). Test `TestScannerCCWorktreeResuelveRecord`.
4. **FE — agrupación por subcarpeta** (`gruposCandidatosDe()`, selector puro con unit test): los
   candidatos de un escaneo se agrupan por primer segmento de `install_path` relativo a
   `proyecto_path` — grupo «proyecto (raíz)» primero (incluye fuera-de-árbol/cache CC, dirs
   ocultos `.claude/…` y avisos sin dir), una sección con heading por subcarpeta. Con UN solo
   grupo la lista queda plana (proyecto simple, cero ruido). Story
   `CandidatosMonorepoAgrupados` calca el caso luana-vitalia post-fix.
5. **Scanner — los remotes del proyecto resuelven en worktrees** (destapado por la verificación
   en vivo del fix 3, preexistente): `ResolverRemotesGit` leía el config del gitdir PROPIO del
   worktree (`.git/worktrees/<n>/config`), que no tiene remotes — viven en el config COMÚN. Ahora
   lee vía `gitCommonDir()`: el root de un worktree gana su scope remoto (antes: scope `"."`).
   Assert agregado a `TestScannerCCWorktreeResuelveRecord`.

Evidencia: Go `./internal/...` verde con los 2 tests nuevos; FE `pnpm run verify` limpio + vitest
unit 27/27 + storybook 128/128 (la story nueva incluida; `ConDatos` de la lista extendida: la
entrada provisional se identifica por scope y el selector de fila se ancla `/^harness\b/`).
**Escaneo VIVO re-corrido contra `~/Proyectos/luana-vitalia` real** (`arnesia portafolio escanear`,
binario del árbol): 15/15 candidatos identificables — root = `github.com/alpacapurpura/luana-platform`,
10 sub-apps por su ruta (`comunify`…`vitalia`), y los 2 ex-avisos anónimos ahora RESUELVEN como
`referenciada-cc` con id (`claude-md-management`, `commit-commands`) + aviso «record de instalación
de …/luana-platform (worktree del mismo repo)». Pendiente del operador: verlo en la app instalada
(el daemon corre el binario viejo hasta reinstalar) y firmar.
