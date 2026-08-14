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
    observed: "⚠️ CORREGIDO 2026-08-14 (CW-D8): esta evidencia era ENGAÑOSA. Ese binario lo produce `cargo build`, que apunta al devUrl (http://localhost:5173) — la ventana abría pero mostraba el error ERR_CONNECTION_REFUSED de WebView2, no la app. El binario válido lo produce `tauri build` (embebe los assets). Lo único que probaba era que el proceso levantaba. Reemplazada por las evidencias de abajo."
  - action: "2026-08-14 · bug reportado por el operador: la app instalada mostraba solo «arnesia: este build no embebe la UI…». Investigación de causa raíz"
    observed: "el instalador estaba BIEN (arnesia-daemon.exe instalado = SHA256 B54256E9… byte-idéntico al ya verificado con la SPA). Respondía OTRO daemon: el del operador en WSL (wslrelay → /mnt/c/…/bin/arnesia, version dev, huella bf914df) compilado sin web/dist. `curl -i :4200/` reprodujo el 404 exacto. Al detenerlo (pkill dirigido), la app instalada levantó su propio sidecar y GET / pasó a 200"
  - action: "reproducción controlada del bug + fix: ocupante que responde 404 en :4200 (scratchpad/ocupante_sin_ui.py) contra el binario de `tauri build`"
    observed: "el shell ya NO se attachea: stderr «:4200 está ocupado por algo que NO sirve la UI: ni attach ni spawn» y la ventana muestra la tarjeta «Otro programa ocupa el puerto» (captura verificada), en vez del 404 crudo"
  - action: "camino feliz con el puerto libre (mismo binario bundleado)"
    observed: "spawnea arnesia-daemon.exe como hijo, GET / → 200, y la ventana muestra la SPA REAL de ArnesIA: rail de sesiones, vistas Mapa/Diag/Corridas/Tren/Hist y el botón Conversar (captura verificada)"
  - action: "gates del shell: cargo test · cargo fmt --check · cargo clippy --all-targets -- -D warnings"
    observed: "6 tests de la decisión del attach en verde; fmt y clippy exit 0 (clippy/rustfmt agregados al toolchain minimal)"
verified_at: 2026-08-14
---

# checkpoint — compilación Windows (F-0 + F-W1)

## Estado

`developed` — el daemon compila, corre y se testea nativo en Windows sin tocar el camino
Linux (bundle.sh VERBATIM; GOOS=linux/darwin verdes). Verificación en vivo registrada arriba.

## Retomar aquí

1. **Gate PARIDAD 🧑‍⚖️**: el operador instala `instaladores/v0.7.1/ArnesIA_0.7.1_x64-setup.exe`
   (SmartScreen mostrará «Ejecutar de todos modos» — sin certificado, decisión ②) y confirma que
   la app abre la interfaz. Si otro daemon ocupa `:4200` (p. ej. uno en WSL), la app ahora lo
   dice con la tarjeta «Otro programa ocupa el puerto» en vez de mostrar su 404 (CW-D8).
   Conviene desinstalar la 0.7.0 previa, que tiene el attach ciego.
2. Siguientes (BACKLOG § Port Windows): suite completa verde en Windows, `bump.py` portable,
   rename-trick del self-update, certificado de firma, `os.UserConfigDir`.
3. Parte B del plan (cockpit :4300 + CIL + skills) sigue podada en
   `PLAN-cockpit-y-windows-arnesia.md` del workspace padre.
