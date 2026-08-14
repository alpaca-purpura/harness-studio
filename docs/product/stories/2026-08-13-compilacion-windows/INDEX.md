# Paquete: compilación Windows (F-0 + F-W1)

> **Qué es:** el daemon `arnesia` compila, corre y se testea NATIVO en Windows sin romper
> Linux/macOS, y el repo se vuelve operable desde una máquina Windows (EOL, hooks, python).
> **Qué NO es:** el instalador de escritorio Tauri Windows (.msi/.exe) ni la migración de la
> maquinaria de proceso de vitalia-app (cockpit/CIL/skills) — ambos DIFERIDOS, ver
> `decisiones.md` § deuda y el plan `PLAN-cockpit-y-windows-arnesia.md` (workspace padre).

## Retomar aquí

- Estado: **construido y verificado en vivo** (ver `checkpoint.md` § dod_evidence). Gate
  PARIDAD 🧑‍⚖️ pendiente de firma del operador.
- Siguiente paquete natural: **instalador Tauri Windows** (prereqs: Rust `x86_64-pc-windows-msvc`
  + VS Build Tools — hoy NO instalados en la máquina) y/o **migración cockpit** (Parte B del plan).

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
  superficie portada.
