# `web/src-tauri/` — Shell de escritorio ArnesIA (Tauri 2)

Esqueleto **mínimo y coherente** con las decisiones firmadas de HS-04
(`research/2026-07-05-arquitectura-fase3.md`, `arch/boundaries/core-no-importa-shell.md`).

**Qué es este shell:** una ventana nativa (WebView) que carga la SPA de Vite y consume la
API HTTP/SSE del daemon Go `arnesia` en `http://localhost:4200`. El shell **no importa el
core** — es un cliente del daemon. Aporta solo: ventana, single-instance y (fase futura) el
lanzamiento del daemon como *sidecar*. El mismo daemon corre headless (dev/CI/server) sin shell.

> **Todo lo marcado `// verificar al instalar` (Cargo.toml, main.rs) debe revisarse en una
> máquina con toolchain + red:** este scaffold se escribió sin acceso a `cargo`/crates, así
> que las versiones están fijadas al *major* de cada crate. Correr `cargo update` y fijar minors.

## Estructura

```
web/src-tauri/
├─ Cargo.toml            crate `arnesia` (lib `arnesia_lib` + bin `arnesia-app`, HS-14: nombre
│                        distinto del daemon Go `arnesia` — evita colisión de PATH); deps Tauri 2
├─ tauri.conf.json       config canónica Tauri 2 (schema v2)
├─ build.rs              tauri_build::build()
├─ src/
│  ├─ main.rs            entry desktop: setea WEBKIT_DISABLE_DMABUF_RENDERER, llama run()
│  └─ lib.rs             run(): registra single-instance + (futuro) spawn del sidecar
├─ capabilities/
│  └─ default.json       capability base (core:default sobre la ventana `main`)
└─ binaries/             aquí se coloca el sidecar arnesia-<target-triple> (no versionado)
```

## Arrancar en dev (máquina con toolchain)

```bash
cargo tauri dev
```

Flujo cableado en `tauri.conf.json`:
- `beforeDevCommand: npm run dev` (corre desde `web/`, levanta Vite) →
- `devUrl: http://localhost:5173` (el WebView carga la SPA de dev) →
- la SPA habla con la API del daemon en `:4200`.

Para `cargo tauri build`: `beforeBuildCommand: npm run build` y `frontendDist: ../dist`.

## Mitigaciones Mint / Linux (load-bearing, HS-04)

Mint 22 = Ubuntu 24.04 base. Riesgos que "muerden primero" y sus mitigaciones ya aplicadas:

- **Cliff WebKitGTK 4.0 → 4.1:** Ubuntu 24.04 dropeó `libwebkit2gtk-4.0` ⇒ **Tauri 2 es
  obligatorio** (linka 4.1). Por eso este shell es Tauri 2, no v1.
- **NVIDIA + DMABUF = ventana negra:** `main.rs` setea `WEBKIT_DISABLE_DMABUF_RENDERER=1`
  **antes** de construir el WebView (solo en Linux). Si aún falla, probar además
  `WEBKIT_DISABLE_COMPOSITING_MODE=1`.
- **Doble ventana / segunda instancia:** plugin **single-instance** registrado de primero en
  `run()`; el callback reenfoca la ventana `main` existente (HS-14 fix ③), no abre una segunda.
  (Combinar con el lock propio del daemon `:4200` bind-or-bail / flock.)

Deps de sistema en Mint/Ubuntu 24.04 (verificar al instalar):
`libwebkit2gtk-4.1-dev`, `build-essential`, `libssl-dev`, `libayatana-appindicator3-dev`,
`librsvg2-dev`, `patchelf`.

## Cablear el sidecar (fase futura)

El daemon Go `arnesia` se declara como `externalBin: ["binaries/arnesia"]` en
`tauri.conf.json`. Tauri exige el binario nombrado con el **target-triple** del host:

```bash
# 1. Obtener el triple del host:
TRIPLE=$(rustc -Vv | grep host | cut -d' ' -f2)      # p.ej. x86_64-unknown-linux-gnu

# 2. Cross-compilar el daemon Go y colocarlo con ese sufijo:
GOOS=linux GOARCH=amd64 go build -o web/src-tauri/binaries/arnesia-$TRIPLE ./cmd/arnesia
```

Luego, para **spawnearlo** desde Rust (hoy solo hay un `TODO` en `lib.rs`):
1. Sumar `tauri-plugin-shell` a `Cargo.toml` y registrarlo en `run()`.
2. Añadir el permiso `shell:allow-execute` (con scope del sidecar) a `capabilities/default.json`.
3. En `setup()`: sondar `:4200`; si no responde, `app.shell().sidecar("arnesia")` con
   sub-comando `serve`. Tauri **no** auto-reapea el sidecar en crash → sumar heartbeat/supervisor.

Hoy el `externalBin` solo declara el bundling; el spawn está fuera de este esqueleto v1.

## Iconos

Este scaffold **no incluye iconos binarios**. Antes de bundlear, generarlos desde el logo:

```bash
cargo tauri icon path/al/logo.png     # crea web/src-tauri/icons/*
```

Y añadir el array `bundle.icon` a `tauri.conf.json` (`icons/32x32.png`, `icons/128x128.png`,
`icons/128x128@2x.png`, `icons/icon.icns`, `icons/icon.ico`). `cargo tauri dev` en Linux
arranca sin iconos; el bundle (`build`) sí los requiere. `icons/` está en `.gitignore`.

## Seguridad (CSP)

`app.security.csp` está en `null` por ahora (dev cómodo). **Endurecer luego**: fijar una CSP
que permita solo la SPA local y `connect-src http://localhost:4200` (+ su SSE), antes de release.
