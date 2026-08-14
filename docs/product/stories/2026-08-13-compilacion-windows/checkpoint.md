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
verified_at: 2026-08-13
---

# checkpoint — compilación Windows (F-0 + F-W1)

## Estado

`developed` — el daemon compila, corre y se testea nativo en Windows sin tocar el camino
Linux (bundle.sh VERBATIM; GOOS=linux/darwin verdes). Verificación en vivo registrada arriba.

## Retomar aquí

1. **Gate PARIDAD 🧑‍⚖️**: el operador corre `python scripts/bundle.py --solo-daemon` +
   `bin\arnesia.exe serve` en su máquina y firma.
2. Siguientes paquetes (BACKLOG § Port Windows): instalador Tauri (.msi/.exe — falta Rust
   MSVC + VS Build Tools), suite completa verde en Windows, rename-trick del self-update.
3. Parte B del plan (cockpit :4300 + CIL + skills) sigue podada en
   `PLAN-cockpit-y-windows-arnesia.md` del workspace padre.
