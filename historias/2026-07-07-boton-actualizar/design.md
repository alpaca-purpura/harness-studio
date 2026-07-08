# design — botón «Actualizar» (UI al pixel + técnica)

> Compañero de `spec.md` (mismo corte: mockup v2 firmado). Tokens = DTCG de
> `web/tokens/base.tokens.json` → `var(--…)`; lo que no está aquí se lee del mockup.

## Anatomía (tarjeta, en la vista Ajustes)

```
┌ Ajustes (vista global; breadcrumb «alpacapurpura / Ajustes») ─────────┐
│ ┌ tarjeta 460px, bg-card, border, radius-md, p-16, gap-14 ──────────┐ │
│ │ VERSIÓN Y ACTUALIZACIÓN            (h3 xs 600 uppercase muted)    │ │
│ │ daemon        arnesia · huella 1c7443f · 2026-07-07   (kv)        │ │
│ │ instalado en  ~/.local/bin/arnesia [pill ok|warn]                 │ │
│ │ origen        /ruta/del/repo | «repo no configurado»              │ │
│ │ [checklist de pasos — solo tras correr]                           │ │
│ │ [Actualizar desde el repo]  (btn primary; disabled según estado)  │ │
│ │ nota 10px muted (toolchain · tren remoto después)                 │ │
│ └───────────────────────────────────────────────────────────────────┘ │
│ (línea muted: «más ajustes llegan: marketplaces · daemon»)            │
└───────────────────────────────────────────────────────────────────────┘
```

## Marcas y tokens (mockup:56-137)

- tarjeta: `bg-card` · `border-border` · `radius-md` · padding 16 · gap 14 · 460px máx.
- kv: k 110px `muted-foreground` xs · v mono `foreground`.
- pill: 9px 700 pill; ok = borde `--ok` fondo `--ok-soft`; warn = borde `--warn` fondo
  `--warn-soft`; texto SIEMPRE `--foreground` (AA — lección del drawer).
- botón: borde `--primary`, fondo `--accent-soft`, texto `--foreground` 600; disabled
  opacity .68 + title del porqué.
- checklist: dot 8px — pendiente `--border` · en curso `--primary` · ok `--ok` · fallo
  `--crit`; texto xs, nombre del paso en 600, detalle mono.
- cajas de desenlace: okbox `--ok-soft` · critbox `--crit-soft` (stderr en mono) ·
  warnbox `--warn-soft` — texto `--foreground` en todas.

## Estados de la tarjeta (máquina)

`idle` → (click) → `actualizando` (botón disabled, sin checklist inventada) → respuesta:
- pasos con fallo → `error` (checklist real + critbox stderr + «Reintentar»)
- «ya al día» → `idle` con okbox «ya estás al día» (sin reinicio)
- instalado → `reiniciando` (polling `/healthz`+`/api/version`, nota «el daemon se
  reinicia…») → huella nueva → `exito` (okbox) · timeout 30s → `error` honesto.
`no-escribible` y `sin-repo` → botón disabled + su caja explicativa (RF-102/103).

## FE (FSD)

- `features/self-update/ui/update-card.tsx` — UI PURA: recibe `version` (RF-107),
  `estado`, `resultado` y callbacks (`onUpdate`); jamás fetchea (patrón Inspector).
- `pages/shell/ui/global-view.tsx` — composition-root: transporte (`api.getVersion`,
  `api.selfUpdate`), polling del reinicio, y monta la tarjeta en la vista Ajustes.
- `shared/api/client.ts`: `getVersion()` · `selfUpdate()` (POST largo: sin timeout FE).
- Story=test por estado (idle/actualizando/error/no-escribible/sin-repo/éxito).

## Go (hexagonal)

- **Puerto** `ports.SelfUpdater`: `Verificar(ctx)` · `Build(ctx)` · `VerificarBinario(ctx)`
  · `Instalar(ctx)` — cada paso devuelve `(detalle string, err error)`; y
  `ports.VersionInfo` para RF-107 (o función en el adapter).
- **Adapter** `internal/adapters/selfupdate/`: ejecuta `scripts/bundle.sh` (cwd=repo,
  contexto cancelable), verifica el binario producido (`bin/arnesia`: existe, ejecutable,
  huella buildinfo ≠/= corriente), instala write-tmp→`os.Rename` en el dir de
  `os.Executable()` (mismo filesystem = atómico). Huella: `debug.ReadBuildInfo()`
  (`vcs.revision`/`vcs.time`) — el binario destino se lee con `go version -m` o
  buildinfo del archivo (adapter decide; sin ejecutar el binario nuevo).
- **Usecase** `usecase.SelfUpdateService`: orquesta pasos, corta al primer fallo,
  arma el reporte `[{paso, estado(ok|fallo|no-corrido), detalle}]`; mutex → 409;
  expone `Reiniciar()` que el transport agenda POST-respuesta (goroutine + delay ≤1s,
  `syscall.Exec` del path del ejecutable con `os.Args`).
- **Transport**: `POST /api/self-update` (200 reporte · 409 en vuelo · 503 si
  no-escribible/sin-repo con el motivo) · `GET /api/version`. OpenAPI + bump.
- **Config**: flag `--repo` / env `ARNESIA_REPO` en `serve`. go-arch-lint: componente
  `selfupdate` mayDependOn `[domain, ports]`; usecase/transport sin cambios de regla.
- **Tests**: adapter (instalar atómico + no-escribible + binario inexistente) ·
  usecase (corte al primer fallo · reporte · 409) — `-race`.

## A11y / honestidad

Botón disabled SIEMPRE con `title` del porqué · checklist solo con veredictos reales
(jamás pasos «en vivo» inventados: la animación del mockup era demo) · stderr visible
· contraste AA ambos temas · reduced-motion respetado (sin spinners animados críticos).
