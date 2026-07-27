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

### Cambiado
- La sesión deja de ser la conversación: el id de Claude Code, el modelo, el uso de contexto, la cadena de rotaciones, el checkpoint y el transcript bajan a la conversación que los tiene. La sesión se queda con el frente de trabajo.

### Deprecado

### Eliminado

### Corregido
- declarada la violación a11y preexistente del dock: el cc-id de la SessionLine va en --primary sobre --secondary (2,21:1)
- el error de historial del picker deja de pintarse en --warn (3,76:1, bajo el mínimo de axe): el texto va en --foreground y la alarma en un borde no textual — con eso el job visual-fitness de CI vuelve a verde
- declarados ?arnes= y ?cerradas= de GET /api/sessions, que se servían sin figurar en el contrato
- Lo que el daemon devuelve de una sesión es un instante y ya no una ventana al registro vivo: mientras el conductor trabajaba, la lectura ya entregada se movía sola.

### Seguridad

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
