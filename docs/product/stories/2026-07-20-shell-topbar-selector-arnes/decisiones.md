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

## Verificación del mockup (no solo lectura de código)

Cada iteración se revisó con Chrome headless real (`google-chrome --headless=new`, screenshots
capturados y leídos), ambos temas, no solo razonamiento sobre el CSS. Cazó 2 bugs reales de layout
antes de firmar:
1. Texto largo (`install_path`) sin `min-width: 0` en una cadena de flexbox anidada forzaba el ancho
   del panel más allá de la ventana — `overflow: hidden` en un ancestro no alcanza si el hijo con
   `white-space: nowrap` nunca recibe permiso de encogerse.
2. La misma clase de bug a nivel CSS Grid: una columna `1fr` **sí puede desbordar** si su contenido no
   puede achicarse — hace falta `minmax(0, 1fr)`, no alcanza con `1fr` a secas.
