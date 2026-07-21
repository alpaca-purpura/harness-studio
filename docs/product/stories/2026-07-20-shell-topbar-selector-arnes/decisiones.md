# Decisiones — Shell · Topbar sin empresa + selector de arnés

> `tipo: decisiones` · paquete `2026-07-20-shell-topbar-selector-arnes` · conversadas y firmadas
> 🧑‍⚖️ por el operador el mismo día, en una sola sesión de diseño (mockup iterado 3 veces). El
> ejecutor de la spec/build NO relitiga estas decisiones; si una resulta inviable en código, documenta
> el porqué AQUÍ (nueva TS-D) y elige la alternativa más cercana al espíritu.

## TS-D1 · Se saca `empresa` del breadcrumb de la Topbar

`Topbar` (`web/src/widgets/topbar/ui/topbar.tsx:22`) pintaba `active.empresa ?? "—"` como primer
segmento del breadcrumb. Es dato hardcodeado de la fixture de sesión (siempre `alpacapurpura` en el
dogfood) y, más de fondo, el modelo real del Portafolio es **N:M** arnés↔empresa — no hay «la» empresa
de un arnés para mostrar en una línea fija. Se elimina el segmento; el breadcrumb queda `arnés / vista`.

## TS-D2 · El chip de arnés deja de simular un dropdown

El chip dibujaba `{active.arnes} ▾` sin `onClick` — un affordance de cambio que no existe y no debería
sugerirse: **el arnés de una sesión es fijo desde que se crea** (1 sesión = 1 conversación con 1
arnés; cambiar de arnés = abrir otra sesión, nunca mutar la actual). Pasa a etiqueta de solo lectura
(borde punteado, sin flecha).

## TS-D3 · Conversar baja a su propia línea

Competía por ancho con el breadcrumb en la misma fila. Pasa a una segunda línea, alineado a la
derecha, sin `margin-left: auto` empujando contra texto variable.

## TS-D4 · `quedaste en: …` sale de la Topbar — es un campo sin cablear

Verificado contra código antes de tocar nada (no se diseña sobre un supuesto): `Session.parked`
(`internal/domain/session.go`) es free-text pensado como «qué mirabas dentro de la vista» — distinto
de `view` (qué pestaña). Pero **ningún usecase lo escribe en vivo**; el único productor es el seed
hardcodeado de seteo inicial (`session_service.go:748-753`, ej. `"spec-writer"`, `"hallazgo éxito
87%"`). Mostrarlo en el mockup como si funcionara habría sido deshonesto (norma del repo: nada se da
por hecho sin verificar). Se saca de la Topbar hasta que exista un productor real; si se retoma, es
paquete propio (necesita decidir QUIÉN lo escribe y cuándo — no es parte de este alcance).

## TS-D5 · La elección de arnés se muda a «＋ Nueva sesión»

`NewSessionButton` (`session-rail.tsx:259-292`) crea la sesión con `arnes: "nuevo-arnes"` fijo y solo
pregunta la ruta por `window.prompt()`. Ningún punto de la UI deja elegir un arnés real — se corrige
ahí: el arnés se elige UNA vez, al abrir la sesión; la sesión ya creada no vuelve a preguntar (ya
cablea con TS-D2: fijo de por vida de la sesión).

## TS-D6 · El picker vive EN la ventana — nunca como popover — con buscador real

**Primer intento rechazado por el operador:** un popover anclado al botón (`position: absolute`,
ancho fijo ~264px). Motivo del rechazo: «debería estar en la misma ventana no como popup», más «puede
ser que tengamos muchos» arneses. Se rediseñó como panel INLINE dentro de la columna canvas del rail
(mismo layout, mismo flujo — cero `position: absolute`, cero backdrop propio), con un buscador real
(filtra por id/nombre/empresas en vivo) y la lista con `max-height` + scroll propio — escala a un
portafolio grande sin hacer crecer la ventana sin límite.

## TS-D7 · La lista es EL Portafolio real del operador, no un catálogo aparte

**Segundo intento rechazado por el operador:** el picker (ya inline, con buscador) seguía mostrando
3 arneses de fixture de Storybook (`entities/arnes/testing/*.ts` — datos del dogfood del Mapa, no del
Portafolio). Motivo del rechazo: «yo debo poder escoger… de entre los arneses que tengo en mi
portafolio, ya que en mi portafolio ya hice el trabajo de agregar arneses». Se investigó el modelo
real (`internal/domain/portafolio.go`, `internal/usecase/portafolio.go`,
`web/src/widgets/portafolio/ui/portafolio-list.tsx`) y se rehizo la fila del picker para reflejarlo
campo a campo: `empresas[]` (plural, N:M — no una sola empresa inventada), `◆ canónico vX.X` /
`▣ N instalac.`, chip `en-deriva` cuando corresponde, punto de salud. Fuente real: `GET /api/portafolio`
(`EntradaPortafolio[]`, keyeadas por `clave`), el mismo endpoint que alimenta la Lista del Portafolio.

## TS-D8 · Selección en 2 niveles: identidad → copia (cuando hay ambigüedad)

Consecuencia directa de TS-D7: una `EntradaPortafolio` puede tener 0-1 `canonico` **y** N
`instalaciones` — no hay un único path implícito. Se eligió:
- **Caso simple** (solo `canonico`, o una única instalación): elegir la fila alcanza; la ruta se
  resuelve sola y se confirma en una línea (`usará <path>`).
- **Caso ambiguo** (2+ copias): al elegir la fila se despliega una sub-lista con cada copia (`tipo` +
  `install_path`/`proyecto_path` truncado + chip de deriva si aplica) — mismo patrón que ya usa el
  Drawer para «Observar en Mapa» (`onObservar(path)`, referencia por `install_path` string, sin id
  numérico). «Crear sesión» queda deshabilitado hasta elegir la copia.
- Se **eliminó** el campo de texto libre «Ruta (opcional)» que tenía el picker en el intento anterior
  — ya no hace falta tipear nada, la ruta sale de la copia elegida. (El `window.prompt()` de ruta del
  flujo VIEJO, «Hoy», se deja intacto en el mockup solo como comparación — ilustra lo que se reemplaza.)

## TS-D9 · Deuda visible para la spec (no bloquea el Gate 1, pero debe resolverse antes de construir)

No maquetado a fondo; la spec tiene que decidirlo, no asumirlo:
- **Portafolio vacío** (0 `EntradaPortafolio`): copy + acción del estado vacío del picker (¿enlaza a
  la vista Portafolio para agregar uno?).
- **Colisión de `clave`/`id`** entre entradas — ver `idsColisionados` en
  `entities/portafolio/model/selectors.ts`, deuda ya conocida del Portafolio (GAP-2, S1-D2); el picker
  no puede asumir que `identidad.id` es único.
- **`GET /api/portafolio` falla o tarda** al abrir «＋ Nueva sesión» — estado de carga/error del panel
  inline (hoy el mockup asume datos ya presentes).
- Confirmar si el picker necesita re-fetchear el portafolio cada vez que se abre (por si se agregó un
  arnés nuevo desde la vista Portafolio en la misma sesión de app) o si alcanza con el estado ya
  cargado en el store.

## TS-D10 · Host del picker: el rail se ensancha (no una columna canvas aparte)

El mockup demuestra el picker en una segunda columna «canvas» junto al rail — pero esa
grilla de 2 columnas es una simplificación de demo. En la app real, esa columna es
`WorkspaceStage`/`GlobalView`, montada por `pages/shell/ui/shell-page.tsx` — fuera del
alcance firmado de este paquete (solo `topbar.tsx` + `session-rail.tsx`). Se decide: el
propio `<aside>` de `SessionRail` (`session-rail.tsx:26-31`) crece de ancho cuando el
picker está abierto — ya tiene `transition-[width]` (colapsado 52px / normal 224px);
se agrega un tercer ancho fijo (360px, el mismo valor que `ChatDock` ya usa en
`shell-page.tsx:37` — no se inventa un número nuevo). Mientras el picker está abierto se
ocultan la lista de sesiones y el pie del rail (nav global + tema); solo header + picker
quedan visibles. Sigue siendo 100% inline (cero `position:absolute`, cero backdrop, cero
popover) — más todavía que el mockup: es la MISMA caja creciendo, no una vecina — y
mantiene el paquete dentro del alcance firmado.

## TS-D11 · Estado vacío del picker

0 `EntradaPortafolio`: copy + CTA «Ir a Portafolio» (cierra el picker y navega
`setView("portafolio")`, la misma ruta global que ya usa el pie del rail vía
`GLOBAL_VIEWS`, `shared/api/types.ts:123`). Mismo espíritu de copy que `VaciaBody` de
`portafolio-list.tsx:230-237` («Tu portafolio está vacío» + acción), pero apuntando a la
vista Portafolio en vez de abrir el wizard directamente (el picker no tiene wizard propio).

## TS-D12 · Colisión de `clave`/`id` — chip distintivo, no bloqueo

La fila del picker selecciona por `clave` (única por diseño del wire) — la colisión de
GAP-2/S1-D2 (`idsColisionados`, `entities/portafolio/model/selectors.ts:134-147`, ya
construido) no puede duplicar la selección real. Pero si `identificadorDe()` de dos filas
coincide, se ve texto idéntico — confuso aunque no ambiguo en el dato. Se decide: cuando
`idsColisionados` marca una fila, se agrega un chip extra con `identidad.home ??
identidad.scope` (truncado) para diferenciarla a simple vista. A diferencia del diálogo de
confirmación que `PortafolioView` sí necesita antes de «Observar en Mapa»
(`portafolio-view.tsx:41-95` — ESE índice del Mapa sí keyea por id pelado, GAP-2 real y
con efecto real), crear una sesión no toca ese índice — no hace falta bloquear con un
diálogo, alcanza con el chip.

## TS-D13 · `GET /api/portafolio` falla o tarda al abrir el picker

Mismos 3 estados que ya usa `PortafolioList` (`portafolio-list.tsx` — prop `estado:
"cargando"|"error"|"datos"`): skeleton compacto mientras carga, mensaje + botón
«Reintentar» si falla (mismo copy pattern que `ErrorBody`, `portafolio-list.tsx:211-228`).
No se inventa un estado nuevo — se calca el que ya existe y ya se entiende.

## TS-D14 · Refetch en cada apertura del picker

Se decide re-fetchear `GET /api/portafolio` cada vez que se abre el picker — no cachear
entre aperturas. Motivo: el operador pudo agregar un arnés desde la vista Portafolio en la
misma corrida de la app (otra pestaña/ruta global de la misma sesión de escritorio);
mostrarle una lista vieja al momento de elegir sería deshonesto. Costo aceptable: un GET
liviano, el mismo endpoint que ya paga la Lista/Drawer del Portafolio.

## TS-D15 · Cómputo de «copias seleccionables»: canónico + instalaciones, casos 0/1/2+

Formaliza TS-D8, que dejó implícito qué cuenta como «copia»: el total de una identidad =
`(canonico ? 1 : 0) + instalaciones.length`. **0 copias** (entrada agregada sin canónico
local ni instalaciones vivas — posible, no maquetado) → fila deshabilitada, sin radio,
tooltip «sin copia local registrada»; no se puede crear sesión desde ahí. **1 copia** →
caso simple (auto-resuelto), sin cambios sobre lo ya firmado. **2+ copias** → sub-lista con
TODAS, canónico incluido con tipo rotulado `canónico` — el mockup solo maquetó el caso de
2 instalaciones sin canónico (`mockups/arnesia-shell-topbar-selector-arnes.html:449-476`);
esto completa el caso mixto que el propio texto de TS-D8 ya anticipaba («0-1 canónico Y N
instalaciones») sin dibujarlo.

## TS-D16 · Qué escribe la sesión nueva (payload real, no el hardcode viejo)

El picker reemplaza `arnes: "nuevo-arnes"` (hardcode) por datos reales de la
`EntradaPortafolio` elegida:
- `arnes` = `identificadorDe(identidad)` — el MISMO accessor que ya usan Lista/Drawer
  (`entities/portafolio/model/selectors.ts:11-13`), incluido su fallback honesto
  `"(sin id)"` para identidades provisionales sin `id` ni `scope`.
- `empresa` = `empresas.join(" · ")` cuando hay alguna. TS-D1 sigue vigente (no hay «la»
  empresa de un arnés) — pero mostrar TODAS unidas no es fabricar una, es mostrar el dato
  N:M real tal cual; distinto de elegir una arbitraria.
- `puesto` queda SIN escribir. El Portafolio no tiene ese concepto — el `"—"` viejo era
  relleno, no dato (mismo criterio que mató `quedaste en:` en TS-D4).
- `salud: "info"` y `view: "Mapa"` se mantienen — siguen siendo ciertos para una sesión
  recién nacida.
- `path` = la copia elegida (`canonico.path` o `instalacion.install_path`) — confina el
  cwd del conductor (S2), mismo campo que ya usa «Observar en Mapa» del Drawer para el
  mismo arnés (`portafolio-drawer.tsx:272,353`).

**Efecto secundario conocido, fuera del alcance de este paquete:**
`workspace-stage.tsx:235` pinta `{s.empresa} · {s.puesto}` en el header de una sesión; con
`puesto` vacío el separador `·` queda colgando para sesiones creadas por el picker nuevo.
Se documenta acá a propósito — ese archivo no está en el alcance firmado de este paquete,
no se toca de rebote; queda BACKLOG si molesta en el uso real.

## TS-D17 · El fetch del picker vive en un store propio del widget, no en la página

A diferencia de `PortafolioView` (página = composition-root, dueña del transporte —
`portafolio-view.tsx:1-25`), `SessionRail` es un widget que hoy solo consume stores
(`useSessions`/`useAppStore`), nunca `shared/api` directo. Se mantiene el mismo patrón:
nuevo store `widgets/session-rail/model/portafolio-picker-store.ts` (Zustand, envuelve
`api.listPortafolio`). Dos razones: (1) `entities/portafolio/model/**` NO puede importar
`shared/api` — boundary `fe-transporte-independiente`, check `domain-not-transport` es
ERROR (`web/.dependency-cruiser.js`); (2) evita tocar `shell-page.tsx` para exponer
transporte desde la página, que ampliaría el alcance firmado (TS-D10 ya lo evitó para el
layout, esto lo evita para el fetch).

## TS-D18 · Revisión en vivo post-build (operador + Chrome real contra el daemon instalado) — 4 ajustes

El operador instaló el `.deb` v0.2.10 y no vio los cambios: causa raíz ajena a este paquete —
`~/.local/bin/arnesia` (override local del self-update, HS-11) tenía un binario del 17-jul, previo
a estos commits; `lib.rs:97` prioriza ese override sobre el sidecar empaquetado. Corregido
(pisado con el binario fresco). Con el daemon ya al día, el operador revisó la app real (no el
mockup) y pidió reconciliar 4 puntos — verificados con Chrome real contra `:4200` (mismo patrón que
«Verificación del mockup» abajo, ahora contra la app corriendo, token inyectado vía
`window.__ARNESIA_TOKEN__`), no solo lectura de código:

1. **Conversar vuelve a la misma fila que el breadcrumb** (reabre TS-D3). El motivo original de
   TS-D3 —competía por ancho— se resuelve distinto esta vez: el breadcrumb es `min-w-0` +
   `overflow-hidden` + `truncate` en `<b>{view}</b>`, y el botón es `flex-none` — no puede volver a
   empujar. `topbar.tsx`.
2. **El picker de `MapBar` (RF-72) se relabelea, no se saca.** No contradice TS-D2 en los hechos —
   `viewedId` es estado local de `WorkspaceStage` (vista previa), nunca muta `session.arnes` — pero
   sentado justo debajo de un breadcrumb que TS-D2 dejó deliberadamente sin `▾` ni `onClick`, un
   `<select aria-label="Elegir arnés">` con options reales lee como la contradicción exacta que
   TS-D2 quiso evitar. Se relabelea: caption visible «vista», `aria-label`/`title` explícitos
   («vista previa de otro arnés — no cambia el arnés fijo de la sesión»). `map-bar.tsx`.
3. **Se deduplica el stack de header** (3 lugares repetían arnés/rol/empresa): el banner de
   `WorkspaceStage` pierde `{s.empresa} · {s.puesto}` (mismo motivo que TS-D1 — `Session.empresa`
   es fixture hardcodeada, N:M sin «la» empresa — y ya duplicado por los chips reales de `MapBar`);
   el chip `rol` de `MapBar` se saca (duplica el valor ya visible del `<select>` "vista").
   `workspace-stage.tsx` + `map-bar.tsx`.
4. **`MapBar` empresa deja de leer solo `empresas[0]`** — mismo criterio N:M que ya se aplicó en
   TS-D1/TS-D16 (`new-session-picker.tsx` ya usaba `empresas.join(" · ")`): se une el join en vez
   de mostrar arbitrariamente la primera y esconder el resto. `map-bar.tsx`.

Verificado: `pnpm run verify` (typecheck+lint+depcruise+fsd+stylelint) verde, `vitest run` 168/168
sin cambios (ninguna story/test fijaba los strings tocados), y visualmente contra `:4200` real
(screenshot antes/después, accessibility tree confirmando `combobox` renombrado y banner sin
`empresa`/`rol`). No cierra el Gate final — `PARIDAD.md` sigue con la firma 🧑‍⚖️ pendiente, ahora
sobre esta versión ajustada.

## TS-D19 · `Topbar` absorbe el banner suelto de `WorkspaceStage` — un solo header de sesión

TS-D18 dedupó contenido entre 3 bandas apiladas pero dejó 2 headers de sesión compitiendo:
`Topbar` (arnés/vista + Conversar) y el `<header>` propio de `workspace-stage.tsx` (arnés grande +
status + salud), ambos permanentes en TODA vista (verificado en vivo: el banner de
`WorkspaceStage` seguía en `Diag`, sin Mapa de por medio) — el mismo dato de identidad partido en
dos cajas que nadie diseñó juntas. Se fusionan en una sola fila de `Topbar`: arnés (etiqueta fija,
TS-D2) · status (`Pip`+`STATUS_LABEL`) · salud (`HealthDot`+`SALUD_LABEL`) · `/` · vista · Conversar.
El `<header>` de `WorkspaceStage` se elimina — el componente ya no pinta chrome de sesión propio.
`topbar.tsx` + `workspace-stage.tsx`.

## TS-D20 · El picker de `MapBar` (RF-72) deja de ser un `<select>` de uso libre

El operador, ya con TS-D18 corregido, siguió viendo el dropdown y objetó de fondo (no solo el
label): "un desplegable que me permite cambiar de arnés dentro de una sesión no tiene sentido" —
correcto, y el relabel de TS-D18 fue cosmético, no arregló la causa. Se releyó el comentario
original de RF-72 (`workspace-stage.tsx`, previo a este cambio): el propósito NUNCA fue un switcher
de uso general — era el escape-hatch para cuando el arnés de la sesión **falla al cargar o no está
indexado** (`loadErr`, ya existente, es exactamente esa señal). La implementación lo montaba
siempre que el portafolio cargara, sin condición — el scope se corrió de "recuperación" a "switcher
casual". Se corrige gateando la interactividad:

- **Camino feliz** (arnés carga bien, `viewedId === arnesId`): texto de solo lectura, mismo
  principio que el Topbar (TS-D2). Sin `<select>`.
- **Error real** (`showPicker = Boolean(loadErr)`): el `<select>` reaparece, relabeleado "ver
  otro" — ahora es honesto, es LA corrección para el caso en que falla.
- **Peek** (`viewedId !== arnesId` sin error — "Observar en Mapa" del Portafolio, GAP-1): texto de
  solo lectura + botón «vista previa · volver» (reusa el mismo `onPick`, sin prop nueva de
  transporte) — nunca un dropdown genérico; el arnés de la sesión (`ownId`) se mantiene visible e
  intacto en `Topbar` todo el tiempo.

Los 3 estados se probaron EN VIVO contra el daemon real (no solo lectura de código): camino feliz
(`dev-full-cycle`, sin picker) → error real (`nuevo-arnes`, no indexado, reaparece "ver otro") →
peek (`Portafolio → Observar en Mapa` de `vitalia` sobre la sesión `nuevo-arnes` → aparece "vista
previa · volver", `Topbar` sigue mostrando `nuevo-arnes`) → volver restaura el error de
`nuevo-arnes` (esperado, es su estado real). `map-bar.tsx` + `workspace-stage.tsx` (nuevas props
`showPicker`/`ownId`).

Verificado: `pnpm run verify` verde, `vitest run` 168/168 sin cambios, 3 estados confirmados con
Chrome real + accessibility tree. `PARIDAD.md` sigue pendiente de firma 🧑‍⚖️, ahora sobre esta
versión.

## Verificación del mockup (no solo lectura de código)

Cada iteración se revisó con Chrome headless real (`google-chrome --headless=new`, screenshots
capturados y leídos), ambos temas, no solo razonamiento sobre el CSS. Cazó 2 bugs reales de layout
antes de firmar:
1. Texto largo (`install_path`) sin `min-width: 0` en una cadena de flexbox anidada forzaba el ancho
   del panel más allá de la ventana — `overflow: hidden` en un ancestro no alcanza si el hijo con
   `white-space: nowrap` nunca recibe permiso de encogerse.
2. La misma clase de bug a nivel CSS Grid: una columna `1fr` **sí puede desbordar** si su contenido no
   puede achicarse — hace falta `minmax(0, 1fr)`, no alcanza con `1fr` a secas.
