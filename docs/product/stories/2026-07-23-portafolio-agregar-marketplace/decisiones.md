# Decisiones — Portafolio · Agregar de marketplace → clonar + mejorar

> `tipo: decisiones` · paquete `2026-07-23-portafolio-agregar-marketplace`. Una entrada por
> decisión conversada, EN EL MISMO TURNO (METODOLOGIA §10). Estado: PENDIENTE-RESOLVER →
> PROPUESTA → FIRMADA. Una entrada en PENDIENTE-RESOLVER es una nota capturada tal cual la dejó
> el operador — todavía NO es una decisión tomada, es el punto de partida de la próxima
> conversación.

## PENDIENTE-01 · Reconciliación proyecto-cargado ↔ marketplace de origen — FIRMADA (2026-07-23)

**Estado: FIRMADA.** Resuelta en conversación nueva (Sonnet 5 propuso, operador confirmó las 4
sub-preguntas sin ajustes). Ya puede arrancar el mockup del paquete.

Cuando se agrega un **proyecto** a ArnesIA (no un arnés suelto — una carpeta con Claude Code
instalado), hay que resolver cómo se accionan los arneses que ya están instalados ahí dentro:

- Al cargar el proyecto se deberían VER los arneses instalados en él.
- Lo más probable: ese arnés instalado NO tiene todavía un match real contra ningún marketplace
  mapeado dentro de ArnesIA. **El arnés siempre debe poder decir de dónde proviene.**
- Para poder hacer ese match, ArnesIA necesita tener los **marketplaces mapeados** (registrados)
  internamente, no solo descubrir arneses sueltos.
- Cuando SÍ hay match: el sistema debe saber, con evidencia (código/mecanismo, no supuesto):
  de qué marketplace viene, qué arnés específico es, en qué versión está, si está desactualizado
  respecto al origen o no.
- Cuando NO hay match: debería poder **crearle un espacio** (mapeo) en un marketplace que el
  operador elija — no queda huérfano sin origen posible.
- **Para qué sirve esto (el objetivo de fondo):** que el flujo de "reparar" funcione end-to-end
  sin que el operador tenga que actualizar el proyecto a mano y pasarle código por git para que
  el usuario final haga pull. En vez de eso: el operador solo entrega **acceso al plugin +
  marketplace**; el usuario final lo instala, sigue el paso a paso indicado, y al terminar tiene
  todo funcionando según lo diseñado — sin intervención manual de código por parte del operador.

**Por qué queda ABIERTA y no se resuelve ahora:** toca el modelo de identidad/origen de Slice 0
(`domain.Arnes`, `(home,id)`, `deriva`) y potencialmente el registro de marketplaces como
entidad propia dentro de ArnesIA — necesita su propia conversación de diseño antes de tocar el
mockup de este paquete, no una resolución apurada en el medio de otra cosa.

**Las 4 sub-preguntas, resueltas:**

1. **Dónde vive el registro** — extiende Portafolio/Slice 0: nueva lista `marketplaces_conocidos`
   en el store del Portafolio (junto a `portafolio.json`). NO entidad top-level nueva. `Registries[]`
   (S0-D3) ya modela el lado "detectado por provenance"; esto agrega el lado "operador" (marketplaces
   que ArnesIA conoce/administra explícitamente).
2. **Matching** — ambos, en orden: manifiesto (`(home,id)` de `plugin.json`) primero (barato,
   determinista); si no hay `home`, cae a comparar `deriva` (hash) contra `home/plugins/<id>/<v>/`
   de los marketplaces conocidos. Reusa el mecanismo collect-all ya construido (S0-D14:
   lock>git-plugin>manifiesto) — no inventa uno nuevo.
3. **Sin match** — manual con sugerencia: se listan marketplaces conocidos ordenados por señales
   blandas (nombre/autor de `plugin.json`), el operador SIEMPRE confirma el destino explícitamente
   — mismo principio anti-drift que el resto del Portafolio (nunca auto-decide, ver S1-D2 colisión
   bare-id).
4. **Flujo "reparar sin pasar código"** — extiende el outcome 5 (Reparar + Backport) ya en
   `BACKLOG.md`, no un 6to outcome. Este PENDIENTE-01 es lo que le faltaba al item 2 (Agregar) para
   que el item 5 (Reparar) tenga con qué reconciliar — cierra ese bloqueo.

**Próximo paso:** arrancar el mockup (punto 1 del flujo del paquete, `INDEX.md`) con este modelo
como insumo.

## AG-D1 · El paquete se re-escopea a «consolidar la agregación» — FIRMADA (2026-07-24)

**Estado: FIRMADA.** Orden explícita del operador en sesión.

El paquete deja de ser «agregar de marketplace» y pasa a ser **la consolidación de TODA la
agregación de arneses al Portafolio** — por marketplace *y* por proyecto, una sola superficie
coherente. La rama Proyecto ya construida (Slice 1) entra al alcance: no como intocable, sino
como mitad a revisar junto con la nueva.

**Por qué:** la auditoría visual en vivo (`auditoria-storybook-vs-app.md`) mostró que la rama
Proyecto ya construida tiene deuda de diseño propia (W1-W6, L1-L3) que el mockup del Slice 1 sí
resolvía y el código nunca bajó. Diseñar la rama Marketplace *al lado* de eso duplicaría el
problema: dos ramas del mismo wizard con dos criterios distintos de pasos, contadores, defaults
de selección y densidad de lista. Se arreglan juntas o no se arreglan.

**Qué implica:**
- El mockup de la etapa 1 cubre **las dos ramas** del wizard `Agregar al portafolio`, no solo la
  nueva. Superset estricto: nada firmado en la PARIDAD del Slice 1 se quita.
- La rama Proyecto **puede cambiar** donde la auditoría probó divergencia contra su propio
  mockup firmado — eso no es alcance nuevo, es cerrar una brecha existente.
- El spec del paquete hereda las 5 preguntas abiertas de `auditoria-storybook-vs-app.md` §7.
- Los ítems 2 y 5 del outcome del programa (`BACKLOG.md`) siguen donde están; lo que cambia es el
  título y el alcance del ítem 2.

## AG-D2 · El mockup del Portafolio es PRE-rebrand: no se forkean sus colores — FIRMADA (2026-07-24)

**Estado: FIRMADA.** Consecuencia directa de la auditoría, no hay fork de criterio.

`mockups/arnesia-portafolio.html` (2026-07-14) quedó **un día antes** del rebrand PRENTER
(2026-07-15): su `--primary` es ámbar `#d9a35b`/`#a8742c` y el token vigente es
`color.semantic.primary` = `brand-teal-500`. El mockup nuevo se deriva de **Storybook en tema
dark**, no de este `.html`; del `.html` se hereda solo la **estructura** (tabs descriptivas,
pasos, contador de hallazgos), jamás la paleta. Ancla: norma «pegarse al Storybook / tokens DTCG
reales» + `mockups/INDEX.md` regla dura 1-2.

Anotado en el propio archivo (cabecera ⚠ STALE) y en su fila de `mockups/INDEX.md` para que el
próximo lector no repita el error. **No se re-derivó el `.html` ahora**: re-derivarlo ES la etapa
1 de este paquete y tiene que salir superset, no calco.

## PENDIENTE-02 · Las 5 preguntas de diseño que abre la auditoría — PENDIENTE-RESOLVER

Capturadas tal cual salieron de la inspección en vivo; se resuelven en la conversación del mockup
(`auditoria-storybook-vs-app.md` §7 tiene el detalle y la evidencia de cada una):

1. **Pasos simultáneos vs máquina de estados** (W3). Hoy `PasoFuente` desaparece en la rama feliz
   de `candidatos` → no se puede ver ni corregir la ruta escaneada, ni re-escanear, sin cerrar el
   wizard. El mockup los muestra juntos. ¿Cuál gana?
2. **¿Premarcar los hallazgos?** (W5). El mockup dice «Agregar 2» con los checkboxes puestos; el
   código arranca en «Agregar 0». Choca con el principio anti-drift «el operador siempre
   confirma» (S1-D2, PENDIENTE-01 §3) — hay que elegir explícitamente, no heredar por descuido.
3. **Densidad de la lista** (W6). `max-height:280px` en modal de `480px` alcanza para 3½ filas.
   Con marketplace en juego, 30+ arneses es el caso normal. ¿Modal más grande, paginado,
   virtualizado?
4. **La toolbar dice «Marketplace» dos veces** (L1), lente y filtro, sin distinción visual.
   ¿Rótulos de grupo visibles (`VER POR` / `FILTROS`, como el mockup), o se renombra una?
5. **¿Dónde vive la reconciliación?** (mitad B de PENDIENTE-01). Registrar
   `marketplaces_conocidos` y resolver un arnés sin match **no es** «agregar algo al portafolio».
   ¿Tercer paso del wizard, superficie aparte, o acción in-situ sobre la fila (como
   «Identificar» del Slice 2)?

**Avance 2026-07-24:** P3 y P4 resueltas (→ AG-D3, AG-D4). P5 con propuesta del operador en
mesa (→ AG-D5, PROPUESTA). P1 y P2 siguen abiertas — el operador pidió que se expliquen antes de
decidir; explicación dada en sesión, respuesta pendiente.

## AG-D3 · Densidad de listas: modal más grande + lazy loading + buscador + filtro — FIRMADA (2026-07-24)

**Estado: FIRMADA.** Respuesta directa del operador a PENDIENTE-02 §3.

Resuelve W6 de la auditoría (hoy: `max-height:280px` dentro de `max-width:480px` → 3½ filas con 4
hallazgos reales). Las cuatro piezas, juntas:

1. **Modal más grande** — el ancho de `480px` no es negociable contra filas que llevan id mono +
   nombre + versión + registry + chips de tipo/deriva/aviso multilínea.
2. **Lazy loading** — no se renderiza la lista completa de golpe.
3. **Buscador** dentro de la lista de hallazgos/catálogo.
4. **Filtro** dentro de la misma.

**Alcance de esta decisión:** aplica a **toda** lista de arneses-a-elegir del flujo de agregación,
no solo al catálogo de marketplace — la rama Proyecto sufre lo mismo en un monorepo (el walker
baja 4 niveles, C-P-11). El caso de 30+ entradas pasa a ser el normal, no el extremo.

**Interacción con AG-D5:** si el catálogo de marketplace se muda a superficie de página completa
(AG-D5), esta decisión **no se vuelve redundante** — sigue rigiendo para el modal de candidatos de
la rama Proyecto, y las 4 piezas se heredan tal cual en la página del catálogo. Ninguna de las dos
decisiones cancela a la otra.

## AG-D4 · La toolbar lleva rótulos de grupo visibles — FIRMADA (2026-07-24)

**Estado: FIRMADA.** Respuesta directa del operador a PENDIENTE-02 §4.

Cierra el bug L1: hoy la toolbar muestra `Empresa · Plano · Proyecto · Marketplace · Estado ·
Marketplace` — seis pills iguales donde la 4ª es *lente* (agrupar por) y la 6ª es *filtro* (acotar
por), sin ninguna distinción visual. Los `role="group"` + `aria-label` existen
(`portafolio-list.tsx:199,233`), así que un lector de pantalla ya las distingue; un ojo no.

**Decisión:** los rótulos de grupo se hacen **visibles**, como en el mockup — `VER POR […]` para
las lentes y `FILTROS […]` para los filtros. No se renombra ninguna de las dos «Marketplace»: con
el rubro a la vista, «ver por marketplace» y «filtrar por marketplace» se leen sin ambigüedad, y
renombrar rompería vocabulario ya firmado en la PARIDAD del Slice 1.

**Consecuencia sobre AG-D5:** habilita que exista un tercer uso de la palabra (el plano
«Marketplaces») sin volver la superficie ilegible — los tres quedan desambiguados por rubro.
Queda ANOTADO, no decidido, que la lente «marketplace» podría volverse redundante el día que
exista el plano; no se toca ahora.

## AG-D5 · «Marketplaces» como plano hermano de los arneses — PROPUESTA (2026-07-24)

**Estado: PROPUESTA.** Idea del operador («¿podría ser una pestaña a la altura de los arneses, que
pueda dar click y ver sus arneses y agregar a mi portafolio?»), desarrollada a pedido suyo. NO
firmada — falta su confirmación.

**El modelo.** El Portafolio pasa a tener **dos planos hermanos**, conmutables por pestaña arriba
de la lista:

- **Arneses** (default) — la superficie actual: lo que YA tengo. Sin cambios de fondo.
- **Marketplaces** (nuevo) — los `marketplaces_conocidos` de PENDIENTE-01 §1. Una fila por
  marketplace (git url · N arneses en catálogo · última sincronización). Click en una fila →
  **catálogo de ese marketplace**: sus arneses, cada uno etiquetado contra mi portafolio
  (`ya lo tengo` · `disponible` · `lo tengo pero en deriva`), con «Agregar a mi portafolio» por
  fila. Hereda las 4 piezas de AG-D3.

**Por qué esto es mejor que el «paso 3 del wizard» que yo había planteado:**

1. **Deshace la contradicción de origen.** Registrar un marketplace y elegir arneses de su
   catálogo son **dos actos distintos**, y solo el segundo es «agregar algo al portafolio».
   Meterlos en el mismo modal era forzarlos a un molde que no les calza.
2. **Le da casa a la mitad B de PENDIENTE-01.** La reconciliación (arnés sin `home` → asignarle
   uno) necesita un lugar donde ver los marketplaces conocidos. Ese lugar ahora existe. La acción
   sigue siendo in-situ sobre la fila del arnés (como «Identificar» del Slice 2) y su selector se
   alimenta de este plano — el operador SIEMPRE confirma el destino (PENDIENTE-01 §3, intacto).
3. **Desbloquea los ítems 3-5 del outcome sin superficie nueva.** «Publicar» (3), «update-check»
   (4) y «Reparar/Backport» (5) son todos conversaciones *contra el marketplace `home`*. El plano
   es donde naturalmente se ve «3 de tus arneses están desactualizados respecto a este
   marketplace» — hoy no hay dónde poner eso.
4. **Un catálogo no cabe en un modal.** 30+ arneses con búsqueda y filtros es una página, no un
   diálogo. AG-D3 lo admite pero pelea contra el formato; acá el formato deja de pelear.

**Qué pasa con el wizard `+ Agregar`** (queda con sus dos tabs, re-propósito de la de marketplace):

- Tab **Proyecto** — sin cambios: carpeta local → Escanear → elegir hallazgos.
- Tab **Marketplace** — deja de intentar «validar + listar + elegir» dentro del modal. Hace solo
  lo que un modal hace bien: **registrar** (pegar git url → validar `marketplace.json` → guardar
  en `marketplaces_conocidos`). Al terminar, **cierra y aterriza en el plano Marketplaces**, en la
  fila recién registrada, con el catálogo abierto. El elegir-y-clonar pasa allá.

Así el wizard queda como «la puerta por la que entran cosas nuevas» (un proyecto que ya tengo, o
un marketplace que quiero conocer) y el plano queda como «el lugar donde navego y jalo». Coherente
con la etapa de **jalar arneses que ya funcionan** ([[hs-vision-sello-jalar-arneses]]).

**Lo que esta propuesta NO resuelve y hay que decidir aparte:**

- ¿La pestaña vive **dentro** de la página Portafolio, o es una entrada propia del rail? Propuesta:
  dentro — es el mismo dominio y comparte el `+ Agregar`.
- ¿El catálogo se lee **en vivo** del marketplace en cada visita, o se cachea con «última
  sincronización» + botón de refrescar? Propuesta: cachear + refrescar explícito (la app es
  conductor, no proxy; y sin red el plano tiene que seguir siendo legible).
- ¿Qué se muestra de un marketplace **inalcanzable** (repo privado sin `gh`, url muerta)? Tiene que
  degradar honesto, jamás catálogo vacío que parezca «no tiene arneses».

**Las 3 quedaron resueltas en AG-D8** (el operador delegó el criterio UX).

## AG-D6 · Wizard: máquina de estados con Atrás + Cancelar — FIRMADA (2026-07-25)

**Estado: FIRMADA.** Respuesta del operador a PENDIENTE-02 §1.

Se **conserva la máquina de 4 estados** (`fuente` → `escaneando` → `candidatos` → `agregando`) —
no se adopta la vista de pasos simultáneos del mockup. Lo que se agrega es la salida que hoy falta:

- **Atrás** — vuelve al paso anterior. Cierra el agujero real de la auditoría (W3): hoy, escaneada
  una ruta con hallazgos, `PasoFuente` no se re-monta y no hay forma de corregir la ruta ni
  re-escanear sin cerrar el wizard entero.
- **Cancelar** — sale del wizard por completo. Distinto de Atrás: uno retrocede un paso, el otro
  aborta el flujo. Mantiene S1-D9 (cancelar = CERO efectos; el widget no persiste nada solo) y
  S1-D19 (durante `agregando` el cierre sigue bloqueado — hay un POST en vuelo).

Implica que **el mockup NO se calca en este punto**: la estructura de pasos simultáneos que el
`.html` dibuja queda descartada a propósito, no por olvido. Lo que sí se toma del mockup son los
rótulos de paso y el contador de hallazgos («encontrados N arneses», W4) — informan sin obligar a
mostrar todo junto.

## AG-D7 · Los hallazgos NO se premarcan — FIRMADA (2026-07-25)

**Estado: FIRMADA.** Respuesta del operador a PENDIENTE-02 §2: «dejar como está actualmente sin
premarcar, yo marco lo que quiero».

Se **mantiene el comportamiento vigente**: checkboxes vacíos, botón «Agregar 0 al portafolio»
deshabilitado hasta que el operador tilde. Se **descarta** la propuesta intermedia que estaba en
mesa (premarcar salvo filas con advertencia) y se **descarta** el premarcado del mockup (W5).

Refuerza la coherencia con el principio anti-drift del Portafolio — el operador siempre elige
explícitamente, sin excepción por conveniencia (S1-D2, PENDIENTE-01 §3). Nota: esto sube el peso de
AG-D3 (buscador + filtro + lazy loading), porque con 30+ hallazgos «tildar a mano» solo es viable
si se puede buscar y filtrar antes de tildar. Las dos decisiones se sostienen mutuamente.

## AG-D8 · Plano «Marketplaces» — propuesta cerrada de UX (criterio delegado) — FIRMADA 🧑‍⚖️ (2026-07-25)

**Estado: FIRMADA 🧑‍⚖️** por el operador el 2026-07-25 («firmo AG-D8, arranca el mockup»). El
operador había delegado el criterio UX pidiendo explícitamente anclarlo en objetivo / necesidad de
negocio / visión. Supersede el borrador AG-D5 y resuelve sus 3 huecos.

Con esta firma **PENDIENTE-02 queda cerrada por completo** y la etapa 1 (mockup) arranca sin
preguntas de flujo abiertas.

### Anclaje (lo que la visión ya dice, y que corrige el borrador AG-D5)

1. **`vision.md` §Qué mutó — MUERE «la idea de operar/cargar arneses de terceros».** Más
   `CLAUDE.md`: «solo arneses propios». Entonces el plano **no es una tienda para navegar**.
   El borrador AG-D5 lo trataba como catálogo genérico: error de encuadre.
2. **`vision.md` §Ecosistema — el marketplace es el SEAM de entrega**:
   `ArnesIA (fábrica) → publica → marketplace git → instala → proyecto del cliente → ejecuta DevHub`.
3. **`vision.md` §Ecosistema — «ArnesIA es dueña única del observar y el modificar»**; las apps de
   rol solo ejecutan. Traer un arnés del estante es **traerlo para trabajarlo**, nunca «instalarlo
   para usarlo».
4. **Necesidad de negocio (PENDIENTE-01 §4 + §Buyer)** — el vendible es *el arnés con su mejora
   continua*; el operador entrega **solo acceso a plugin + marketplace** y el cliente queda
   funcionando sin handoff manual de código.

**Reencuadre resultante:** el plano Marketplaces es **el estante de lo que vendemos, y el espejo
que dice si lo que el cliente tiene coincide con lo que publicamos.** Sirve dos direcciones:
**saliente** (¿publiqué? ¿mi canónico está adelantado?) y **entrante** (este arnés que encontré en
un proyecto, ¿de qué estante salió y en qué versión?).

### Decisión 1 — dos clases de marketplace, y no son simétricas

- **Propio** — publicamos ahí. Catálogo navegable, **Traer canónico habilitado**, participa de
  publicar / update-check / reparar.
- **De referencia** — aparece en la procedencia de arneses ajenos que el escaneo encuentra (p. ej.
  `anthropics/claude-plugins-official`, visto en un escaneo real). Se registra **solo** para que un
  arnés pueda decir de dónde viene. Catálogo **read-only, sin Traer** — habilitarlo sería
  exactamente «operar arneses de terceros», que la visión mata.

Sin esta partición el plano se convierte en un gestor universal de plugins: justo lo que
`vision.md` §Lo que NO es prohíbe.

### Decisión 2 — la pestaña vive DENTRO de Portafolio

Dos planos hermanos conmutables arriba de la lista: **Arneses** (default, superficie actual) ·
**Marketplaces**. No entrada nueva del rail: mismo dominio, comparten el `+ Agregar`, y el rail ya
lleva 4 entradas. Con AG-D4 (rótulos `VER POR`/`FILTROS` visibles) los tres usos de la palabra
—plano, lente, filtro— quedan desambiguados por rubro.

### Decisión 3 — catálogo cacheado con «última lectura» + refrescar explícito

No se lee en vivo en cada visita: la app es **conductor, no proxy** (§Ecosistema), y sin red el
plano tiene que seguir siendo legible. Cada fila muestra **cuándo se leyó** — mismo espíritu que
`procedencia` (dato fechado y atribuido, no número flotando).

### Decisión 4 — un marketplace inalcanzable degrada honesto

Estados explícitos por fila: `no leído aún` · `sin acceso (gh no autenticado)` · `url no
resuelve` · `leído hace <t>`. **Jamás** catálogo vacío que se lea como «este marketplace no tiene
arneses» — es el mismo pass-fabricado que G3 mató en el Slice 1.

### Decisión 5 — el verbo YA existe: se reusa, no se inventa

`↧ Traer canónico` vive hoy como botón `disabled` en el drawer del arnés
(`portafolio-drawer.tsx:288`, story `portafolio-drawer.stories.tsx:106`, y en el mockup línea 519).
La fila del catálogo usa **ese mismo verbo**, y el botón del drawer es la **segunda puerta** al
mismo acto. Dos puertas, un acto, cero vocabulario nuevo (`mockups/INDEX.md` regla dura 4).

### Decisión 6 — el catálogo muestra situación, no un botón genérico

Por arnés del catálogo, una de estas — cada una con su acción, y la que no aplica no se pinta:

| situación | acción |
|---|---|
| no lo tengo | `↧ Traer canónico` |
| lo tengo · al hilo | sin acción — link a su ficha |
| lo tengo · mi copia adelantada | **Publicar** (ítem 3 del outcome) |
| lo tengo · el estante adelantado | **Actualizar mi copia** (ítem 4) |
| lo tengo · instalaciones en deriva | **Reparar** (ítem 5) |

Esto es lo que hace que el plano **desbloquee los ítems 3-5 del outcome sin superficie nueva**:
son todos conversaciones contra el marketplace `home`, y hoy no existe dónde ponerlas.

### Decisión 7 — la reconciliación se hace en UN lugar, con DOS puertas

La acción de asignarle origen a un arnés huérfano queda **in-situ sobre la fila del arnés** en el
plano Arneses (mismo patrón que «Identificar» del Slice 2), con selector alimentado de
`marketplaces_conocidos`, ordenado por señales blandas, y **el operador siempre confirma**
(PENDIENTE-01 §3, intacto). El plano Marketplaces aporta la **segunda puerta**: un contador
cruzado «N arneses tuyos sin origen resuelto» que lleva al plano Arneses ya filtrado. Un solo
lugar donde se ejecuta; dos formas de llegar.

### Decisión 8 — el wizard `+ Agregar` deja de pelear con el catálogo

- Tab **Proyecto** — sin cambios de fondo, más AG-D6 (Atrás/Cancelar) y AG-D7 (sin premarcar).
- Tab **Marketplace** — hace solo lo que un modal hace bien: **registrar**. Pegar git url →
  validar `marketplace.json` → elegir clase (propio / de referencia, Decisión 1) → guardar en
  `marketplaces_conocidos`. Al terminar **cierra y aterriza en el plano**, en la fila nueva, con el
  catálogo abierto.

El elegir-y-traer **no ocurre dentro del modal**. Wizard = puerta de entrada de cosas nuevas al
conocimiento del sistema; plano = donde navego y jalo. Coherente con la etapa de jalar arneses que
ya funcionan y sellarlos.

### Interacción con AG-D3

AG-D3 (modal más grande + lazy loading + buscador + filtro) **no se cae**: rige para el modal de
candidatos de la rama Proyecto (un monorepo escanea muchos), y sus 4 piezas se heredan tal cual en
la página del catálogo — donde además tienen espacio de sobra. Ninguna cancela a la otra.

### Lo que esta propuesta NO decide

- Cómo se lee el catálogo de un marketplace git (¿`marketplace.json` del default branch por `gh
  api`? ¿clone shallow?) — es mecanismo, va al `spec.md`, no acá.
- Si la lente «marketplace» del plano Arneses se vuelve redundante al existir el plano. Anotado en
  AG-D4, sigue sin decidir. No se toca en este paquete salvo que el mockup lo pida.

## GATE 1 · Mockup FIRMADO 🧑‍⚖️ (2026-07-25)

El operador firmó el dibujo (`mockup-agregacion-consolidada.html`, 6 superficies) y ordenó pasar a
spec. Habilita etapa 3 del flujo (`spec.md` + `design.md`).

## AG-D9 · `known_marketplaces.json` es un eslabón REAL de CC: se lee, no se inventa — FIRMADA (2026-07-25)

**Estado: FIRMADA por evidencia** (no es preferencia: es lo que la máquina real tiene).

Antes de escribir el spec se inspeccionó la máquina. **Claude Code ya lleva su propio registro de
marketplaces** en `~/.claude/plugins/known_marketplaces.json`, con esta forma real (5 entradas
vivas hoy en esta laptop):

```json
{ "prenter-marketplace": {
    "source": { "source": "github", "repo": "alpacapurpura/prenter-marketplace" },
    "installLocation": "/home/chalreme/.claude/plugins/marketplaces/prenter-marketplace",
    "lastUpdated": "2026-07-10T00:36:43.459Z" } }
```

**Decisión:** `marketplaces_conocidos` **NO** es una lista virgen que el operador llena a mano. Es
**collect-all** de dos lados, exactamente el mismo patrón que la cadena de origen (S0-D14):

1. **detectado** — `known_marketplaces.json` de CC (nombre · `owner/repo` · `installLocation` ·
   `lastUpdated`). Da checkout local gratis y fecha de última lectura real.
2. **declarado** — lo que el operador registra por el wizard (marketplaces que CC no conoce todavía,
   o clase `propio`/`de referencia` que CC no modela).

Merge por **nombre de marketplace**; si los dos lados discrepan en `repo`, se muestran las dos
señales (nunca se elige en silencio — mismo criterio que `discrepancias` de `ResolverOrigen`).
Consecuencia práctica: **el plano Marketplaces nace poblado**, no vacío. Y `installLocation` es
justamente la ruta `home/plugins/<id>/<v>/` contra la que `deriva` ya compara — el eslabón cierra
el círculo que faltaba.

## AG-D10 · La forma real de `marketplace.json`: `source` es ruta relativa con la versión adentro — FIRMADA (2026-07-25)

**Estado: FIRMADA por evidencia.** Leído de `~/.claude/plugins/marketplaces/prenter-marketplace/.claude-plugin/marketplace.json`:

```json
{ "name": "prenter-marketplace",
  "owner": { "name": "Prenter", "email": "hola@alpacapurpura.lat" },
  "metadata": { "description": "…" },
  "plugins": [
    { "name": "harness",      "source": "./plugins/harness/0.5.3", "description": "Canal ESTABLE …" },
    { "name": "harness-beta", "source": "./plugins/harness/0.5.3", "description": "Canal BETA …" } ] }
```

Tres cosas que el spec tiene que respetar y que **nadie había escrito**:

1. El archivo vive en **`.claude-plugin/marketplace.json`**, no en la raíz del repo. La validación
   del wizard apunta ahí.
2. **`source` es una ruta relativa que lleva la versión adentro** (`./plugins/harness/0.5.3`). La
   versión del catálogo NO es un campo propio: se deriva de la ruta. Coherente con
   `home/plugins/<id>/<versión>/`, la referencia de `deriva` (ya construida en Slice 0).
3. El árbol real conserva **todas** las versiones lado a lado
   (`plugins/harness/{0.5.0,0.5.1,0.5.2,0.5.3}`) — el catálogo apunta a una, el resto sigue en disco.

## AG-D11 · Fila del catálogo = ENTRADA DE ÍNDICE (canal), no arnés físico — FIRMADA 🧑‍⚖️ (2026-07-25)

**Estado: FIRMADA 🧑‍⚖️.** El operador eligió «una fila por canal» tras ver las dos opciones
renderizadas — exactamente la propuesta de abajo, sin ajustes. Queda **cerrado** el punto de cambio
que el spec §8 dejaba abierto: NO se deduplica por `source`.

El dato real muestra que **dos entradas del catálogo pueden apuntar al MISMO arnés físico**:
`harness` y `harness-beta` comparten `source: ./plugins/harness/0.5.3`. Son **canales**
(estable/beta), no arneses distintos. El mockup firmado asumió 1 fila = 1 arnés, así que este caso
no está dibujado.

**Propuesta:** la fila del catálogo es la **entrada de índice** (o sea: el canal), porque es lo que
el cliente instala — `/plugin install harness@prenter-marketplace` referencia el nombre de la
entrada, no la carpeta. Cuando dos entradas comparten `source`, la segunda se marca con un chip
`mismo contenido que harness` para que nadie crea que son dos arneses. Alternativa descartada:
deduplicar por `source` y mostrar los canales como chips de una sola fila — se descarta porque
esconde el nombre instalable, que es el dato operativo.

**Sub-hallazgo (no bloquea):** `catalogo.json` de prenter (`canales: {estable, beta}` ·
`versiones[]` con `estado: habilitada|deprecada`) es una **convención propia de prenter, NO parte
del estándar `marketplace.json`**. El lector del catálogo se apoya en `marketplace.json` (universal)
y **enriquece si existe** `catalogo.json`, degradando sin ruido cuando no está. Nunca al revés: un
marketplace ajeno no tiene por qué tenerlo.

## AG-D17 · `↧ Traer canónico` entra COMPLETO: local + clone externo — FIRMADA 🧑‍⚖️ (2026-07-25)

**Estado: FIRMADA 🧑‍⚖️.** Resuelve C2 de [`design.md`](./design.md) §2, que era un **hueco del
`spec.md`, no del diseño**: §0 no declaraba `Traer` ni dentro ni fuera, §2.1 decía «se habilita» sin
mecanismo, ningún §4 lo describía y ninguno de los 30 escenarios lo ejercitaba — mientras el texto
del ítem 2 en `BACKLOG.md` sí lo pedía literal («→ checkout `<checkouts>/<home-slug>/<id>/`
(`gh`/PAT) → chat/mejorar») y el mockup firmado lo dibuja como acción principal en 3 de 7 filas.

El operador eligió **alcance completo** sobre las dos alternativas (local-first con externo diferido,
o diferir todo). Se construyen **los dos caminos**:

- **Camino A · local** — cuando CC ya tiene el marketplace clonado (`installLocation` en disco) y el
  `source` es ruta relativa: se **copia la subcarpeta** a `<checkouts>`. Sin red, sin auth, sin git.
  Es el caso de `prenter-marketplace` (`./plugins/harness/0.5.3` ya está en disco), o sea **el caso
  de lo que vendemos**.
- **Camino B · externo** — cuando `source` es objeto (`git-subdir`/`url`/`github`, AG-D13: 220 de 273
  entradas del catálogo oficial): clone shallow con `ref`, verificación de `sha` si viene, extracción
  de `path`, rollback si corta a mitad. Auth `gh` → PAT fallback.

**Asimetría que justifica que esto SÍ entre y Publicar/Actualizar/Reparar NO** (§0 sigue igual):
`Traer` escribe **solo en nuestro propio directorio** (`~/.arnesia/checkouts/`). Los otros tres
escriben en cosas del cliente o en el estante remoto — cada uno con su propio gate.

Reglas que trae consigo: **BR-13..BR-18** (`spec.md` §5) y escenarios **E-76..E-90** (§6). El
mecanismo técnico exacto (contratos, endpoint, adapters, atomicidad, plan de pruebas, capability) lo
baja `design.md` §13.

## AG-D13..D16 · RATIFICADAS (2026-07-25) — la muestra de AG-D10 era un solo marketplace

**Estado: RATIFICADAS.** El arquitecto inspeccionó **los 5** marketplaces de la laptop (AG-D9/D10 se
habían firmado mirando **uno**: prenter, 2 entradas). Aparecieron 4 hechos que cambian el contrato.
Son **evidencia, no preferencia** — se ratifican sin gate de producto; el detalle está en
[`design.md`](./design.md) §1. Superseden parcialmente AG-D10 **sin invertir su espíritu**: nada se
inventa nunca.

| id | hallazgo | qué supersede |
|---|---|---|
| **AG-D13** | **`source` tiene 4 formas reales, no 1.** En `claude-plugins-official` (273 entradas): 53 string-ruta · 78 objeto `git-subdir` · 140 objeto `url` · 2 objeto `github`. Un `source` objeto **no** lleva versión en la ruta: la lleva en `ref`/`sha`, que no son semver. | AG-D10 §2 era cierto para prenter y falso para el 80 % del catálogo oficial. El parser acepta `string \| objeto` y normaliza |
| **AG-D14** | **`version` SÍ es campo estándar de `plugins[]`** — presente en 273/273 (259 `null` explícito, 14 con semver real). El archivo declara `$schema: https://anthropic.com/claude-code/marketplace.schema.json`: **hay esquema oficial** | **BR-2** pasa de «derivación exclusiva» a **precedencia con procedencia anotada**: `campo-version` → `derivada-de-source` → ausente. El espíritu intacto: jamás fabricada |
| **AG-D15** | el catálogo ajeno trae campos que el spec no modela (`strict`, `skills`, `lspServers`, `homepage`, `author`, `displayName`, `keywords`, `category`, `tags`; y raíz `$schema`, `renames`). `owner` es `{name,email?}` **o** `{name,url?}`. `description` vive **top-level** en 4 de 5 y bajo `metadata.` solo en prenter | el adapter **traduce y descarta** (anti-corruption layer); `domain.EntradaCatalogo` no crece un campo por extra ajeno. **`renames` sí se modela**: tiene efecto sobre identidad |
| **AG-D16** | el catálogo oficial es **273 entradas / 159 KB**, no 41 | `ListaLazy` (AG-D3) deja de ser confort y pasa a **requisito**. El caché deja de ser «lindo para offline» y pasa a ser **O(1) por fila**: no se pueden parsear 5 `marketplace.json` en cada `GET` |

## AG-D12 · Vitalia es el caso testigo de «0 hallazgos honesto» — FIRMADA (2026-07-25)

**Estado: FIRMADA por evidencia.** `/home/chalreme/Proyectos/vitalia` (remote
`github.com/alpacapurpura/vitalia`) tiene `.claude/settings.json` + `.claude/skills/crear-pieza-comercial`
pero **ningún `.claude/plugins/`**: es un proyecto con skill local, no con plugin instalado.

Escanearlo devuelve **0 candidatos legítimamente**. Sirve como fixture E2E de la rama honesta que ya
existe («No encontré arneses instalados aquí…») y prueba que el mensaje no miente. El proyecto con
plugins reales para el E2E positivo es `/home/chalreme/Proyectos/luana-platform` (5 plugins de
`claude-plugins-official`) y este mismo repo (`harness@prenter-marketplace` declarado en
`enabledPlugins` sin record de instalación — el aviso que la auditoría ya vio en vivo).
