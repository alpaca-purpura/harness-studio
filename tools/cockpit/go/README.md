# cockpit-go — Cockpit SDD como binario standalone

Port en Go del Cockpit Next.js: mismo filesystem-as-DB, misma API JSON (paridad
verificada endpoint a endpoint), misma UI (el frontend Next.js se exporta a
estático y viaja embebido en el binario vía `go:embed`).

**Resultado: un solo archivo de ~9.3MB, cero dependencias npm/Node en runtime.**

```bash
# Workspace único
./cockpit start -workspace ~/Proyectos/demo-environment -port 4000

# Multi-workspace (Fase 3): registrá tus proyectos una vez, corré el daemon
./cockpit add ~/Proyectos/demo-environment
./cockpit auto-add ~/Proyectos          # escanea y registra todo workspace SDD
./cockpit start -d                       # daemon · PID + logs en ~/.cockpit/
./cockpit status
# → http://localhost:4000 · cada brand aparece como "proyecto/brand" en el selector
```

## CLI

| Comando | Qué hace |
|---|---|
| `cockpit [start] [-port N] [-workspace P] [-d]` | server; sin `-workspace` usa el registry (multi) · `-d` = daemon |
| `cockpit stop` / `cockpit status` | controla el daemon (`~/.cockpit/cockpit.pid`, logs en `~/.cockpit/logs/`) |
| `cockpit add PATH [-name N]` / `remove NAME` | gestiona el registry `~/.cockpit/cockpit.yaml` |
| `cockpit auto-add DIR` | registra todo subdir de DIR que sea workspace SDD |
| `cockpit list` | proyectos + brands + conteo de stories |
| `cockpit story search TEXTO` | búsqueda global de stories (id/título/goal) en todos los proyectos |

**Modo multi-workspace:** las brands se exponen compuestas (`proyecto/brand`) —
el brand switcher existente de la UI se convierte en selector global sin ningún
cambio de frontend. El watcher SSE, sessions y el whitelist de `/api/file`
resuelven multi-root. El CIL/harness se sirve del primer proyecto que tenga
`docs/process/harness-backlog.md`.

## Build

Requiere Go 1.23+ y (solo para regenerar la UI) Node 20 + pnpm:

```bash
./build-ui.sh          # exporta cockpit-ui/ → ui/ (Next.js static export)
go build -ldflags="-s -w" -o cockpit .

# cross-compile
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o cockpit-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o cockpit.exe .
```

`build-ui.sh` aparta `cockpit/app/api/` durante el export (las API routes no
existen bajo `output:'export'` — las implementa este server) y restaura al salir.

## Qué implementa

- **Lectura** (paridad JSON 1:1 con el Next.js original, verificada por diff):
  stories (list/single), capabilities (list/single + status/bidirectional/
  code-index/doctor), releases, learnings, system-map, value-stream, sessions,
  harness, cil, operator-input, file, brands.
- **Escritura:** PATCH stories (campos operador), PATCH capabilities (status),
  releases POST/PUT/DELETE, transition (whitelist operador), merge-release
  (dual-confirm), operator-input PATCH, file PUT, extend-cap, from-done,
  refs/upload, open (editor local).
- **Watcher:** `/api/watch` SSE con fsnotify (debounce 200ms, docType detection,
  heartbeat 30s) — hot-reload del board al editar archivos.
- **UI:** export estático de `../cockpit` embebido; el cliente pide las brands a
  `GET /api/brands` al montar (en dev Next.js siguen llegando por SSR).

## Paridad

`e2e/diff-api.py` compara byte-a-byte (estructural) las respuestas JSON del
server Next.js y este binario sobre el mismo workspace. Estado al port:
**18/18 endpoints idénticos** sobre demo-environment.

`e2e/verify.js` (Playwright) corre la suite UI completa contra el binario:
**14/14 checks** (boards por brand, drawer, roadmap shipped, map, brand switcher,
transición de operador vía UI).

```bash
# paridad API (con ambos servers corriendo)
python3 e2e/diff-api.py

# suite UI contra el binario
BASE_URL=http://localhost:4010 node e2e/verify.js
```

## Limitaciones conocidas vs Next.js

- La serialización YAML al escribir puede diferir en detalles cosméticos de
  quoting respecto a gray-matter (semánticamente equivalente; el orden de keys
  del frontmatter se preserva).
- `pidAlive` en Windows es aproximado (no hay signal-0 probe).
- El dev-flow de la UI sigue siendo `cockpit-ui/ + pnpm dev` — este binario es el
  artefacto de distribución, no el entorno de desarrollo del frontend.
