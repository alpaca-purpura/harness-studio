# Paquete: compilación Windows (F-0 + F-W1)

> **Qué es:** el daemon `arnesia` compila, corre y se testea NATIVO en Windows sin romper
> Linux/macOS; el repo se vuelve operable desde una máquina Windows (EOL, hooks, python); y
> **el instalador de escritorio Windows (.msi + .exe) se construye** (ampliación 2026-08-14,
> CW-D7 — la toolchain se instaló y `instaladores/v0.7.0/` existe).
> **Qué NO es:** la migración de la maquinaria de proceso de vitalia-app (cockpit/CIL/skills),
> DIFERIDA — ver el plan `PLAN-cockpit-y-windows-arnesia.md` (workspace padre).

## Retomar aquí

- Estado: **construido y verificado en vivo** (ver `checkpoint.md` § dod_evidence), instalador
  incluido. Gate PARIDAD 🧑‍⚖️ pendiente: el operador instala el `.exe` y confirma que la app
  levanta su propio sidecar.
- Siguiente paquete natural: **migración del cockpit** (Parte B del plan) y/o el resto de la
  deuda Windows (suite completa, `bump.py`, rename-trick, certificado de firma).

## Artefactos

| Doc | Qué tiene |
|---|---|
| `decisiones.md` | Decisiones ①–⑤ ratificadas + deuda registrada (suite Windows, rename-trick, certificado, UserConfigDir) |
| `checkpoint.md` | Estado + dod_evidence de la verificación en vivo |

## Cambios (resumen por área)

- **`internal/adapters/selfupdate/`** — split por-OS: `os_{unix,windows}.go` +
  `instalar_{unix,windows}.go`; unix conserva rename atómico + re-exec VERBATIM; Windows
  degrada honesto (staged `arnesia-nuevo.exe` + instalar=fallo accionable). Era EL bloqueante
  de `GOOS=windows go build ./...`.
- **`cmd/arnesia/main.go` + `internal/adapters/stt/local/`** — descubrimiento de binarios
  (`claude`, motores STT) con candidatos Windows + PATHEXT, gateado por `runtime.GOOS`.
- **`scripts/bundle.py`** — espejo portable de `bundle.sh` (mismo sello RF-231); es lo que
  invoca el self-update Windows. `bundle.sh` INTACTO (camino canónico Linux).
- **Tests** — fix sistémico de aislamiento (`USERPROFILE`/`APPDATA` junto a `HOME` — en Windows
  `os.UserHomeDir` lee USERPROFILE y los tests contaminaban el `~/.arnesia` REAL), fixtures
  `filepath.FromSlash`, sufijo `.exe` en binarios de prueba, split de tests por-OS en selfupdate.
- **`.gitattributes` + shim `PYTHON` (Makefile/lefthook) + README § Desarrollo en Windows** —
  el repo committeable desde Windows; mapa de puertos documentado (4200 daemon · 4300 reservado
  cockpit · 4002 cockpit vitalia).
- **CI** — job `go-windows` (windows-latest): build+vet del árbol completo + tests de la
  superficie portada. Y `tauri-windows` (solo en push a `main`): sidecar + `tauri build` +
  artifacts con el `.msi`/`.exe`.
- **`scripts/installer.ps1`** (ampliación 2026-08-14) — espejo de `_installer-build` del
  `Makefile` con utilidades nativas (`Get-FileHash` en vez de `sha256sum`, sin `dev-sync`);
  produce `instaladores/vX.Y.Z/` + `checksums.txt`. Toolchain: Rust 1.97.1 sobre el MSVC 2019
  que la máquina ya tenía.
- **Attach verificado (CW-D8, v0.7.1)** — `web/src-tauri/src/lib.rs`: `daemon_running` (TCP a
  secas) → `estado_daemon` + `veredicto_daemon` con 6 tests. Cierra el bug reportado: con otro
  `arnesia serve` sin la SPA ocupando `:4200` (caso real: el daemon del operador en WSL), la app
  instalada mostraba el 404 crudo del ocupante; ahora la ventana explica el conflicto
  (`web/public/conectando.html`, tercer estado). El gate de ese check corre en CI (`cargo test`
  agregado al job `rust`). Shim de python también en `bump.sh` + `changelog.py` con stdout UTF-8
  (la consola cp1252 abortaba el bump por un `✓`).
