# Changelog

Qué cambió en cada versión **publicada** de ArnesIA, para quien la instala — no por qué se decidió
(eso vive en [`docs/product/LEDGER.md`](docs/product/LEDGER.md)) ni qué sabe hacer el sistema hoy
(eso es [`docs/product/capabilities/`](docs/product/capabilities/INDEX.md)).

Formato: [Keep a Changelog 1.1.0](https://keepachangelog.com/es-ES/1.1.0/) · versionado:
[SemVer 2.0.0](https://semver.org/lang/es/). Categorías, las 6 canónicas y ninguna más:
**Agregado · Cambiado · Deprecado · Eliminado · Corregido · Seguridad**.

## Cómo se escribe (no es opcional)

- **En el mismo turno en que se construye**, igual que `decisiones.md` (metodología §10):
  `python3 scripts/changelog.py add Agregado "lo que hiciste"`
- **El bump lo promueve solo:** `make bump-patch|bump-minor|bump-major` mueve `[Sin publicar]` a
  una sección versionada con fecha. Si `[Sin publicar]` está vacía, **el bump falla y no toca
  ningún manifiesto** — no se puede publicar una versión muda.
- Enforcado por `scripts/changelog.py` (gate del bump) ·
  `docs/architecture/fitness/changelog_test.go` (CI) · `lefthook.yml` job `changelog` (pre-commit).
- La regla completa: [`docs/architecture/conventions/versionado.md`](docs/architecture/conventions/versionado.md).

<!-- convencion-desde: 0.2.22 -->

## [Sin publicar]

### Agregado
- story baseline del panel de conversación (chat-dock.stories.tsx): el widget que no tenía ninguna ahora fija sus 4 filas de cromo, su tarjeta de permiso y su turno en vuelo
- gate de contrato del daemon: una ruta que se sirve está declarada en openapi.yaml o exenta con razón escrita (4 enforcers + allowlist de 15 entradas)
- La conversación pasa a ser una entidad propia del dominio, con su título editable, su marca de activa y su ciclo de vida: una sesión contiene N conversaciones y exactamente una activa, y esa invariante se repara al cargar diciendo qué reparó.
- El registro de sesiones viaja en un sobre que dice de qué versión es, qué build lo escribió y cuándo, y el arranque sabe migrarlo con copia previa obligatoria. El archivo de la versión anterior queda intacto: volver atrás no necesita restaurar nada.
- El arranque del daemon dice qué le hizo al registro de sesiones: qué migró, dónde quedó la copia previa, qué reparó y qué llaves movió. Un arranque sin novedades no imprime nada.
- Las conversaciones de una sesión: crear una nueva y retomar una anterior sin perder ninguna de las dos. Cambiar de hilo es una sola transición — o pasa entera, o el estado anterior queda intacto.
- El buscador del panel entra al TEXTO de la conversación, no sólo al título: encuentra por lo que se dijo, ignora acentos y mayúsculas, y muestra el pedazo donde coincidió.
- El daemon expone las conversaciones de una sesión: listarlas, buscarlas, crear una nueva, retomar una anterior y renombrarla.
- El panel de conversación muestra las conversaciones de su sesión: se listan, se buscan por lo que se dijo adentro, se crean y se retoman sin salir del dock
- Verificación de punta a punta contra la aplicación instalada: los seis guiones del plan corren contra el binario que el operador ejecuta, con los datos de sesiones aislados en una copia, y dejan su informe con capturas
- Fixture sintético de migración con los 20 campos del esquema v1 poblados, más un guard reflexivo: 6 mutaciones que antes sobrevivían (checkpoint, cadena_cc, status, puesto, parked, cerrada_en) ahora ponen el árbol rojo.

### Cambiado
- La sesión deja de ser la conversación: el id de Claude Code, el modelo, el uso de contexto, la cadena de rotaciones, el checkpoint y el transcript bajan a la conversación que los tiene. La sesión se queda con el frente de trabajo.
- Cerrar un frente de trabajo ya no tira el transcript ni el checkpoint de sus conversaciones: se archiva lo que el operador vio, que es lo único que sobrevive a una limpieza del corpus de Claude Code y lo único sobre lo que se puede buscar.
- Toda sesión nace con su conversación activa. Antes nacía sin ninguna y la primera lectura la creaba avisando de una «reparación» de algo que nunca estuvo roto.
- `GET /api/sessions` vuelve a devolver SIEMPRE un arreglo y cada sesión lleva su conversación activa en vez de todas sus conversaciones con los diálogos completos.
- El cromo del dock baja de cuatro filas a dos: la identidad técnica del proceso pasa a un chip desplegable —que ahora también dice en qué carpeta corre— y la fila de alcance sólo aparece cuando hay un nodo elegido

### Deprecado

### Eliminado
- Se retiran las rutas del historial por arnés: el parámetro `cerradas` de la lista de sesiones ahora responde con el puntero a su reemplazo, y la ruta que reconstruía turnos desde el corpus nativo ya no se sirve. La capacidad se conserva para lo archivado antes de la migración.
- El selector de arnés ya no lista conversaciones: pertenecían a una sesión, no a un arnés, y sus dos rutas se retiraron con el modelo nuevo

### Corregido
- declarada la violación a11y preexistente del dock: el cc-id de la SessionLine va en --primary sobre --secondary (2,21:1)
- el error de historial del picker deja de pintarse en --warn (3,76:1, bajo el mínimo de axe): el texto va en --foreground y la alarma en un borde no textual — con eso el job visual-fitness de CI vuelve a verde
- declarados ?arnes= y ?cerradas= de GET /api/sessions, que se servían sin figurar en el contrato
- Lo que el daemon devuelve de una sesión es un instante y ya no una ventana al registro vivo: mientras el conductor trabajaba, la lectura ya entregada se movía sola.
- Las sesiones cuya llave de arnés quedó a medias se recalibran a la clave completa: tres de las cinco del registro real estaban invisibles cuando la interfaz preguntaba por la clave. Nada se borra, nada se fusiona, y se ve antes de aplicarse.
- Leer un registro de sesiones escrito por una versión anterior ya no tira en silencio lo que cambió de lugar — le costaba los 90 turnos de la conversación más larga en disco.
- La marca de «contexto rotado» ahora aparece en el diálogo sin recargar la aplicación. Antes se guardaba en disco y el operador no la veía hasta reabrir.
- El identificador de la sesión de Claude Code dejó de pintarse con un color que no llegaba al contraste mínimo de texto sobre el fondo del dock
- La migración del registro de sesiones se probó sobre el archivo real del operador (en copia): las cinco sesiones y la conversación de noventa turnos sobreviven, el archivo de la versión anterior queda intacto y volver a arrancar no vuelve a migrar
- Se registran seis hallazgos que sólo aparecieron al probar contra la aplicación instalada, entre ellos que la marca de contexto rotado puede perderse sin aviso cuando la misma sesión se mira desde dos ventanas
- El título de una conversación se deriva de su primer mensaje (CV-D9/RF-303): estaba firmado, escrito en el dominio y sin cablear — toda conversación se llamaba «nueva conversación» para siempre.
- La fecha de última interacción se estampa en cada turno (CV-D13/RF-304): nunca se escribía, así que toda fila de la lista decía «sin fecha» y el orden del panel caía en silencio al de creación.
- TestResumeAutoSana dejó de ser flaky: esperaba un Idle que ya estaba puesto, así que el turno 2 corría contra un handle vivo todavía no soltado (falla reproducida en la base, ~1 de 100 bajo carga).

### Seguridad
- Un registro de sesiones ilegible se guarda entero con su sello en vez de pisarse, y uno escrito por una versión más nueva deja el daemon en solo-lectura en vez de destruirlo. Una mutación que no se pudo guardar ya no queda viva en memoria.

## [0.2.24] — 2026-07-26

### Agregado

### Cambiado

### Deprecado

### Eliminado

### Corregido
- Capa Mejora: encender la capa ya no colapsa el canvas del Mapa; el escenario s2-instrumentado deja de decir «nunca corrió» mientras muestra gasto; un error del detalle ya no se pinta como dato; y la fila del Portafolio tiene una sola implementación

### Seguridad

## [0.2.23] — 2026-07-26

### Agregado
- Capa «Mejora» del Mapa: telemetría embebida con ingesta OTLP y hook, costeo con catálogo propio, join dinero×proceso por (sesión,turno), 6 detectores de fuga y superficie en Mapa y Portafolio

### Cambiado

### Deprecado

### Eliminado

### Corregido

### Seguridad
- La telemetría se ingiere por allowlist en los dos caminos: no se persiste identidad de cuenta ni contenido de conversación, y la ruta del proyecto se guarda como huella

## [0.2.22] — 2026-07-26

### Agregado
- Identidad de build en Ajustes: la tarjeta muestra `arnesia vX.Y.Z.AAMMDDHHMM` (semver + sello de compilación) y cuándo se compiló, con el commit debajo — dos builds del mismo commit ya se distinguen (RF-231).
- Aviso en Ajustes cuando en disco hay un build más nuevo que el que está corriendo, con la acción concreta: cerrar y reabrir la app, o `make dev-sync` (RF-231).
- `CHANGELOG.md` + `scripts/changelog.py`: el bump de versión ahora exige y promueve el registro de qué se agrega, corrige o elimina.
- `make bump-minor` y `make bump-major`, con el criterio de cuándo usar cada uno escrito en la convención de versionado.
- Dictado por voz en el composer del chat: grabar → transcribir → ordenar → poblar, con las etapas nombradas y el permiso de micrófono concedido por el shell.
- Marketplaces en el Portafolio: plano propio, catálogo por marketplace con situación por fila, wizard de registro, traer canónico y asignar origen a un arnés huérfano.
- Log del daemon a archivo rotativo, y el WebView deja rastro ahí — un incidente ya no se pierde.
- Registro obligatorio de cambios por versión: CHANGELOG.md + `make bump-minor`/`bump-major` + gate que impide publicar una versión que no dice qué trae (RF-232).

### Cambiado
- El sello de build se inyecta en `scripts/bundle.sh`, el único camino por el que pasan el build a mano, `make dev-sync` y el self-update: la app ya no pierde su identidad al actualizarse a sí misma.

### Corregido
- El `charset` faltante en el mockup de Ajustes, que se veía bien por `file://` y salía mojibake servido por HTTP.

## [0.2.21] y anteriores

**Sin changelog reconstruido, a propósito.** Hay 20 releases en `instaladores/` (v0.2.2 …
v0.2.21) anteriores a esta convención: nadie puede jurar hoy qué entró exactamente en cada una, y
rellenarlas de memoria sería inventar. La historia real de ese tramo vive en dos lugares
verificables:

- **decisiones y fichas:** [`docs/product/LEDGER.md`](docs/product/LEDGER.md) → `docs/product/ledger/HS-NN.md`
- **el árbol:** `git log` — cada iteración firmada se commiteó a `main` (trunk-based)

El changelog es exigible **desde 0.2.22** (marcador `convencion-desde` arriba); el enforcement no
valida nada anterior.
