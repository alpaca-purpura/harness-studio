# PARIDAD — fix: repo configurable para self-update

> Contrato: spec.md (RF-108..111) ↔ código real. Gate: lado Go verificado E2E en vivo
> (daemon real, headless); lado Tauri (picker nativo) requiere sesión con display —
> gate humano explícito, honesto (no fingido headless).

| RF | Elemento | Componente real | Evidencia | Estado |
|---|---|---|---|---|
| RF-108 | persistencia + precedencia (`--repo`/env > persistido > vacío) | `internal/adapters/selfupdate/repo_store.go#RepoStore` + `cmd/arnesia/main.go` runServe | `go test -race` (repo_store_test.go, 5 sub-tests) + **E2E vivo**: daemon sin `--repo` → `PUT` válido → `kill` → relanzado sin `--repo`, mismo `$HOME` → `GET /api/version` devuelve el repo persistido; relanzado CON `--repo /tmp/otro-repo-explicito` → el flag PISA lo persistido | ✅ |
| RF-109 | `PUT /api/self-update/repo` (valida con reglas de Verificar, fija en caliente, persiste; inválido = 400 sin tocar nada) | `internal/ports/selfupdate.go#ConfigurarRepo` + `internal/adapters/selfupdate/updater.go#ConfigurarRepo` + `internal/usecase/selfupdate_service.go#ConfigurarRepo` + `internal/adapters/transport/http/selfupdate.go#putSelfUpdateRepo` | `go test -race` (updater_test.go `TestConfigurarRepo` 3 casos · selfupdate_service_test.go `TestConfigurarRepo` 4 casos · selfupdate_test.go 3 casos de wire) + **E2E vivo**: `PUT` con `/tmp` (sin go.mod) → 400 con motivo exacto; `PUT` con el repo real de ArnesIA → 200 + `GET /api/version` refleja el repo activo | ✅ |
| RF-110 | selector nativo en Ajustes (botón «Elegir carpeta…» SOLO dentro de Tauri) | `web/src/features/self-update/ui/update-card.tsx#UpdateCard` (prop `onElegirRepo`) + `web/src/pages/shell/ui/global-view.tsx#AjustesView` + `web/src/shared/lib/platform.ts#isTauri` | stories `SinRepoConSelectorNativo` · `ConfigurandoRepo` · `RepoConfigInvalido` (vitest+play, 93/93 tests OK) + `tsc --noEmit` · `biome check` · `steiger` · `depcruise` · `stylelint` todos verdes | 🔶 código+stories verificados; **falta interacción real del diálogo nativo con display** (gate humano, ver abajo) |
| RF-111 | `PUT` bajo `withAuth`; `POST /api/self-update` sigue con cero parámetros (RF-106 intacto) | `internal/adapters/transport/http/router.go` (misma `withAuth` que todo `/api/*`) | `go test`: `TestPostSelfUpdateIgnoraParametrosDelRequest` (ya existía, sigue verde sin tocar) + `TestPutSelfUpdateRepoInvalido400`/`Valido`/`SinPath400` (nuevos) | ✅ |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar en vivo (display) · ✅ verificado en vivo.

## Evidencia Go E2E (headless, sesión 2026-07-13)

```
# daemon SIN --repo:
GET /api/version → repo:""

# PUT inválido:
PUT /self-update/repo {"path":"/tmp"} → 400
  "el repo configurado no es un árbol Go legible: open /tmp/go.mod: no such file or directory"

# PUT válido (repo real de ArnesIA):
PUT /self-update/repo {"path":"/home/chalreme/Proyectos/harness-studio"} → 200
  "repo /home/chalreme/Proyectos/harness-studio · módulo esperado · go/pnpm/bash presentes"
GET /api/version → repo:"/home/chalreme/Proyectos/harness-studio"

# persistido en ~/.arnesia/self-update.json:
{"repo": "/home/chalreme/Proyectos/harness-studio"}

# kill + relanzar SIN --repo, mismo HOME:
GET /api/version → repo:"/home/chalreme/Proyectos/harness-studio"  (releyó el persistido)

# relanzar CON --repo /tmp/otro-repo-explicito:
GET /api/version → repo:"/tmp/otro-repo-explicito"  (el flag pisa lo persistido — precedencia OK)
```

`go build ./...` · `go vet ./...` · `go test -race ./internal/adapters/selfupdate/... ./internal/usecase/... ./internal/adapters/transport/http/...` · `go test ./...` (repo completo, incluye `TestCapabilityCoverage`/`TestCapabilityPointersResolve`) — todos verdes.

## Evidencia FE (sesión 2026-07-13)

`pnpm run typecheck` · `pnpm run lint` (biome) · `pnpm run fsd` (steiger) · `pnpm run depcruise` · `pnpm run stylelint` · `pnpm run test` (vitest, 93/93 incluye las 3 stories nuevas con `play()`) — todos verdes. `cargo check` en `web/src-tauri` compiló con `tauri-plugin-dialog` + el permiso `dialog:allow-open` (valida contra el schema real del plugin al build).

## Desviaciones registradas

1. **`PUT /api/self-update/repo` responde 400 para TODO error** (candidato inválido Y
   fallo de persistencia del store, p.ej. disco lleno) — no distingue con un código
   distinto. Un fallo de persistencia no es semánticamente «bad request», pero
   separar los códigos exigía un sentinel adicional que el bugfix no justificaba
   (caso raro; el mensaje trae el motivo exacto igual). Documentado, no maquillado.
2. **El picker nativo (RF-110) no se ejecutó de verdad** — esta sesión corrió headless
   (background job), sin display para abrir la ventana Tauri y click-through el
   diálogo de carpeta del OS. Sí se verificó: compila (`cargo check`), el permiso del
   capability es válido, el wire FE→API es correcto (stories con `play()` simulan el
   click y confirman `onElegirRepo`/`configurandoRepo`/`repoConfigError`), y la lógica
   de `onElegirRepo` en `global-view.tsx` fue revisada línea por línea. Lo que falta es
   estrictamente la interacción humana con el diálogo del sistema operativo.

## Firma

- [ ] 🧑‍⚖️ **Gate humano** — pendiente: abrir la app Tauri en una sesión con display
  (`pnpm --dir web tauri dev`, o el `.deb`/`.AppImage` empaquetado), ir a ⚙ Ajustes con
  el daemon arrancado SIN `--repo`, click «Elegir carpeta…», confirmar que el diálogo
  nativo abre, elegir un repo válido y ver la tarjeta pasar a identidad honesta; probar
  también un repo inválido y confirmar el mensaje inline. Con esa firma el paquete
  CIERRA.
