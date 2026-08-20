# Cockpit — copia vendored

Esta carpeta es una **copia** del cockpit de `vitalia-app`, no un submódulo ni una
dependencia. Se usa como herramienta de desarrollo de ArnesIA: lee `docs/product/` de la
raíz del repo y lo pinta (board, roadmap, proceso).

- **Origen:** `vitalia-app/tools/cockpit/`
- **Commit de origen:** `cc7bdfdf2b4bfd4ccb097fb888c9530632b7b70d` (2026-07-30)
  · HEAD del repo fuente al copiar: `cb541ea` (2026-08-01)
- **Copiado el:** 2026-08-18 (paquete `docs/product/stories/2026-08-18-cockpit-y-doctrina/`)
- **Module path:** se conserva el de upstream (`github.com/alpacapurpura/prenter-harness/cockpit-go`)
  para que un diff contra el origen siga siendo legible.

## Por qué vendored y no importado

Son dos repos independientes sin monorepo que los una. Un import cruzado ataría el build de
ArnesIA al ciclo de vida de un repo que no controla. El costo del vendoring es el **drift**,
y se mitiga con dos cosas: esta tabla, y `go/vendored_arnesia_test.go` — si un merge desde
upstream borra una de las adaptaciones, cae un test en vez de romperse el board en silencio.

## Qué NO se copió

`go/cockpit` (binario ELF de 11 MB), `go/ui/` (la UI compilada — se regenera), y
`ui/{node_modules,.next,out}`. Todos gitignoreados acá también.

## Diffs contra upstream

### 1 · Escrituras sobre el pseudo-sistema `platform` (decisión CD-2)

Upstream trata `platform` como contexto de **solo trazabilidad**: `getSelectableSistemas()`
lo incluye para leer y `getSistemas()` lo excluye para escribir. Eso asume un workspace con
sistemas reales (subdirs con `docs/product`) donde la raíz es un agregado transversal.

harness-studio no tiene sistemas reales: su `docs/product/` **vive en la raíz**, así que
todo su contenido ES `platform`. Con la regla upstream el board sería un museo — el drag
daba 400 en `handlers_stories.go`.

| Archivo | Cambio |
|---|---|
| `go/workspace.go` | nuevo `writableSistemas()` + `flagAllowPlatformWrites`. Sin el flag el comportamiento es idéntico a upstream |
| `go/cli.go` | flag `-allow-platform-writes`, propagado al hijo en modo `-d` |
| `go/handlers_{stories,releases,create,regen}.go` | los gates de escritura pasan por `writableSistemas()` |
| `go/handlers_misc.go` | `GET /api/sistemas` devuelve además `escribibles[]` |
| `ui/components/providers/SistemaProvider.tsx` | expone `escribibles` desde el server |
| `ui/components/sistema/board/BoardView.tsx` | el gate de solo-lectura sale de `escribibles`, no de `isPlatform()` hardcodeado |

El punto de la última fila: al habilitar el backend, el board **seguía** bloqueado porque la
UI deducía la escribibilidad por su cuenta. Dos capas decidiendo lo mismo por separado es una
contradicción esperando ocurrir; ahora la autoridad es el servidor, que es quien tiene el flag.

### 2 · Capabilities generadas (decisión CD-4 y CD-5)

| Archivo | Cambio |
|---|---|
| `go/handlers_caps.go` | `user_facing_name` cae a `name` (las caps de ArnesIA usan `name`; sin esto la UI pinta el slug) |
| `go/handlers_caps.go` | PATCH → **409 explicado** si el workspace trae `scripts/capabilities_to_yaml.py`: las caps son artefactos generados (R4) y su enum de status es propio (`vivo · vivo·nc · parcial · stub`), no el de `gates.go` |
| `go/handlers_caps.go` | un `.yaml` que abre `---` y no cierra se lee como documento YAML completo — antes devolvía defaults **fabricados** (`capability_id: ""`, `status: "live"`) para 3 caps del repo |
| `go/handlers_regen.go` | el hint del degradado nombra `scripts/cap_doctor.py --index`, el doctor real de este repo |

### 3 · Vistas que aplican a `platform`

| Archivo | Cambio |
|---|---|
| `ui/lib/platform-context.ts` | `PLATFORM_VIEWS.roadmap: false → true` — acá platform ES el proyecto entero y `docs/product/releases/` existe |
| `ui/components/sistema/{roadmap,drift,map}/*View.tsx` | los guards de render y de fetch consultan `viewAppliesTo()` en vez de repetir `isPlatform()` — una sola fuente decide |

### 4 · Portabilidad Windows

Upstream nunca corrió en Windows. Lo que estaba roto:

| Archivo | Qué fallaba | Fix |
|---|---|---|
| `go/handlers_file.go` | `filepath.Clean` traduce a `\`, y `segmentWhitelisted` parte por `/` → **el file-API rechazaba TODO path**. Y `filepath.IsAbs("/etc/passwd")` es false en Windows → un absoluto POSIX entraba como relativo | `isTraversalSafe` normaliza a slash con `path.Clean` y rechaza absolutos en ambas convenciones (incluido volumen/UNC) |
| `go/handlers_file.go` | la cadena de editores era Linux pura (`xdg-open`, `xed`, `nano`) → «Abrir» fallaba las 5 veces | `defaultEditorChain()` por OS; en Windows `code` → `cmd /c start` |
| `go/watch.go` | `docTypeFor` comparaba con `/capabilities/` → en Windows todo caía en `other` y la UI no refrescaba la tarjeta correcta | `filepath.ToSlash` antes de comparar |
| `go/handlers_regen.go` | `pythonBin` buscaba `.venv/bin/python` y caía a `python3`, que en Windows es el stub falso de WindowsApps | sondea `.venv\Scripts\python.exe` y valida el candidato EJECUTÁNDOLO |
| `go/torre.go` | `sh -c` a secas | `torreShell()`: usa `sh` si existe (Git Bash), y si no reporta `no-medido` con motivo en vez de traducir a `cmd.exe` a la fuerza |
| `go/build.sh` | leía `../../../core-harness/VERSION`, que no existe acá → el binario quedaba sellado `dev` | lee `web/src-tauri/Cargo.toml`, la misma fuente que `bundle.py` e `installer.ps1` |

### 5 · Arranque

Nuevo `cockpit.ps1` (`build|run|stop|status`) — el Makefile del repo host no corre en
Windows y `build.sh`/`build-ui.sh` son bash. Replica su semántica, incluido el stash de
`ui/app/api` (Next con `output:'export'` no admite rutas API) restaurado con `finally`.

## Limitaciones conocidas

- **Los comentarios del frontmatter se pierden al transicionar.** `stringifyFrontmatter`
  re-serializa con `yaml.v3`, que no preserva comentarios: la línea de evidencia que escribe
  `scripts/backfill_checkpoints.py` desaparece en el primer drag. El marcador
  `backfilled: true` sí sobrevive.
- **El primer render pide `?sistema=main`** (`FALLBACK_SISTEMA` del provider) y deja tres
  400 en la consola antes de resolver `platform`. Ruidoso, sin efecto funcional.

## Cómo actualizar desde upstream

1. Traé el árbol nuevo a un lado y diffeá contra esta copia.
2. Aplicá lo que quieras, corré `go test ./...` en `go/` y `npx vitest run` en `ui/`.
3. Si `vendored_arnesia_test.go` falla, el merge borró una adaptación de esta tabla:
   re-aplicala, no borres el test.
4. Actualizá el commit de origen de arriba.
