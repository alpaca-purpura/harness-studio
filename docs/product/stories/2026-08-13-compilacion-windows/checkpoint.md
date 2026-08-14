---
story_id: 2026-08-13-compilacion-windows
state: developed            # construido + verificado en vivo; falta gate PARIDAD 🧑‍⚖️ del operador
module: self-update
cap_target: arnesia.self-update.self-update-sin-sudo (extend)
chris_verify:
  signoff: false            # pendiente: el operador ejercita arnesia.exe en su máquina
defer_audit: false
dod_live_verified: true
dod_evidence:
  - action: "python scripts/bundle.py --solo-daemon en Windows (pnpm build SPA + go build sellado)"
    observed: "bin/arnesia.exe 21.9 MB con sello 0.7.0.2608132155 (bajo el techo de 25 MB)"
  - action: "arnesia.exe serve --addr 127.0.0.1:4262 con USERPROFILE/HOME/APPDATA sandboxeados"
    observed: "GET /healthz = {status:ok} · GET /api/version = huella 51cef76+sucio + version 0.7.0.2608132155 + compilado (RF-231 intacto en Windows) · GET / = SPA embebida 200 · daemon detenido limpio, cero footprint en el ~ real"
  - action: "go build ./... con GOOS=windows, linux y darwin + go vet"
    observed: "los 3 targets compilan; vet limpio (el bloqueante selfupdate eliminado)"
  - action: "go test en Windows: selfupdate · cmd/arnesia · telemetria/store · fitness (gates capability/changelog/headless)"
    observed: "VERDES (con 2 skips honestos en fitness); resto de la suite = deuda CW-D5 con inventario por clase"
  - action: "2026-08-14 · instalar toolchain: rustup-init user-scope (SHA256 verificado contra static.rust-lang.org) --profile minimal --default-host x86_64-pc-windows-msvc"
    observed: "Rust 1.97.1 · rustc -vV host = x86_64-pc-windows-msvc · linkeo real probado (hello world compila y corre) contra el MSVC 2019 YA instalado (cl.exe 14.29.30133 + Win10SDK 19041) — no hizo falta instalar Build Tools 2022"
  - action: "cargo build --release del shell Tauri (web/src-tauri) con el Cargo.lock intacto"
    observed: "Finished release en 7m24s · arnesia-app.exe 4.5 MB · único warning = mensaje informativo del linker MSVC (import library del cdylib), no un fallo"
  - action: "powershell -File scripts/installer.ps1 (guard + bundle.py + recolección + checksums)"
    observed: "instaladores/v0.7.0/ con ArnesIA_0.7.0_x64-setup.exe (8.62 MB, NSIS) + ArnesIA_0.7.0_x64_en-US.msi (11.09 MB, WiX) + checksums.txt SHA256; Tauri descargó NSIS/WiX solo"
  - action: "live-verify de la app real: lanzar target/release/arnesia-app.exe"
    observed: "ventana «ArnesIA» abierta, WebView2 151.0.4129.78 hijo con user-data-dir lat.alpacapurpura.arnesia, cierre limpio. NO ejercitó el spawn del sidecar: :4200 ya estaba ocupado por el daemon del operador corriendo en WSL (wslrelay) → el shell entró en modo attach, que es su comportamiento correcto. El sidecar Windows es byte-idéntico al bin/arnesia.exe ya verificado sirviendo /healthz + /api/version + SPA"
verified_at: 2026-08-14
---

# checkpoint — compilación Windows (F-0 + F-W1)

## Estado

`developed` — el daemon compila, corre y se testea nativo en Windows sin tocar el camino
Linux (bundle.sh VERBATIM; GOOS=linux/darwin verdes). Verificación en vivo registrada arriba.

## Retomar aquí

1. **Gate PARIDAD 🧑‍⚖️**: el operador instala `instaladores/v0.7.0/ArnesIA_0.7.0_x64-setup.exe`
   (SmartScreen mostrará «Ejecutar de todos modos» — sin certificado, decisión ②), abre la app
   instalada y confirma que levanta su propio sidecar (para eso conviene que NO haya otro daemon
   escuchando en `:4200`, como el de WSL durante la verificación del 2026-08-14), y firma.
2. Siguientes (BACKLOG § Port Windows): suite completa verde en Windows, `bump.py` portable,
   rename-trick del self-update, certificado de firma, `os.UserConfigDir`.
3. Parte B del plan (cockpit :4300 + CIL + skills) sigue podada en
   `PLAN-cockpit-y-windows-arnesia.md` del workspace padre.
