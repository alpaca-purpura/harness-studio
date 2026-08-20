# Decisiones — compilación Windows

> Ratificadas por el operador el 2026-08-13 (sesión de plan v3, `PLAN-cockpit-y-windows-arnesia.md`
> en el workspace padre). Cada decisión se firma acá EN EL MISMO TURNO en que se construye.

## CW-D1 · Self-update en Windows = degradación honesta 🧑‍⚖️ RATIFICADA

Windows no permite reemplazar un `.exe` en ejecución (la imagen queda mapeada como sección).
El paso ④ (instalar) en Windows: deja el binario nuevo staged en `arnesia-nuevo.exe` junto al
instalado y **falla con la acción concreta** («cerrá la app y renombralo sobre X»). El reporte
de 5 pasos muestra la verdad — build ✓ · verificar binario ✓ · instalar ✗ accionable ·
reiniciar no-corrido — jamás un «actualizado» que no actualizó (doctrina «gris ≠ verde»).
**Alternativa evaluada y DIFERIDA** (→ deuda): rename-trick NTFS — renombrar el `.exe`
corriente SÍ es legal (sobrescribirlo no): `os.Rename(exe, exe+".old")` → copiar nuevo →
reiniciar → limpiar `.old` al boot. Se difiere para no sumar superficie de fallo en el primer
port (AV/locks).

## CW-D2 · Puertos: el daemon SE QUEDA en 4200; 4300 reservado al cockpit 🧑‍⚖️ RATIFICADA

No hay colisión hoy (el daemon usa 4200, no 4002), pero el 4002 propuesto en el plan v1 para
el cockpit futuro colisionaba con el cockpit de vitalia-app (misma máquina). El 4200 está
cableado en ≥8 lugares (SPA, `lib.rs:30`, allowlist Tauri, `auth.go:53`, settings del dogfood,
tests, skills) — cambiarlo rompería la app instalada. Mapa documentado en README § Desarrollo
en Windows. Follow-up en deuda: parametrizar el puerto del daemon (la SPA hardcodea 4200 —
deuda YA conocida del checkpoint raíz).

## CW-D3 · `bundle.sh` INTACTO; `bundle.py` es el camino Windows 🧑‍⚖️ RATIFICADA (implícita en «Linux como siempre»)

Linux no cambia: `bundle.sh` sigue siendo el camino canónico y el self-update unix lo invoca
VERBATIM. `scripts/bundle.py` es el espejo portable (mismos 3 pasos, mismo sello RF-231 —
el sello vive en el script de bundle, jamás en el Makefile) y es lo que invoca el self-update
Windows (`Build()` de `os_windows.go`). Unificar ambos en un solo script = decisión futura,
no de este paquete.

## CW-D4 · Instalador Tauri Windows — ⚠️ SUPERSEDED por CW-D7 (2026-08-14): YA NO está diferido

*(histórico)* Verificado el 2026-08-13: `rustc`/`cargo` NO instalados → `tauri build` era
imposible en esa ejecución (Tauri no cross-compila). Prereqs documentados en README. Sin
certificado de firma (ratificado ②): SmartScreen mostrará «Ejecutar de todos modos» —
certificado OV/EV sigue siendo deuda para cuando se distribuya. **El bloqueo se levantó en
CW-D7: la toolchain se instaló y el instalador se produce.**

## CW-D7 · Toolchain instalada y instalador Windows CONSTRUIDO 🧑‍⚖️ RATIFICADA (2026-08-14)

Pedido del operador: «instalá Rust/MSVC de los prereqs para sanar la deuda registrada».
Hallazgos que cambiaron el alcance y decisiones tomadas:

- **MSVC NO hacía falta instalarlo.** La máquina ya tenía **Visual Studio Build Tools 2019**
  (16.11.35425) con `VC.Tools.x86.x64` — `cl.exe` 14.29.30133 — y **Windows SDK 10.0.19041**,
  más WebView2 Runtime 151.x. Se REUTILIZA en vez de instalar Build Tools 2022 (ahorra ~5 GB y
  ~30 min). Validado en vivo: el shell Tauri linkea sin tocar `vcvars` (rustc resuelve el
  toolset vía `vswhere`). El README pasa a decir «Build Tools 2019 **o** 2022», que es la
  verdad medida, en vez de exigir la versión más nueva por costumbre.
- **Rust: instalación estándar en C:, no en el disco con más espacio.** C: quedó con 14.47 GB
  tras la limpieza del operador y es el **único SSD** (NVMe); D:/E: son particiones del mismo
  HDD SATA. Con `lto = true` + `codegen-units = 1` el build es I/O-intensivo: compilar en HDD
  costaría varias veces más. Se instaló `rustup` user-scope (sin admin) con
  `--profile minimal` y host `x86_64-pc-windows-msvc` → **Rust 1.97.1**. Integridad del
  `rustup-init.exe` verificada contra el SHA256 oficial de `static.rust-lang.org` (no trae
  firma Authenticode, es lo normal en rustup).
- **El `Cargo.lock` NO se toca.** La nota «correr `cargo update` al instalar» del `Cargo.toml`
  se resuelve CONFIRMANDO que el lock existente resuelve en Windows (compila en 7m24s), no
  regenerándolo: el lock es el contrato compartido con Linux y actualizarlo divergiría las dos
  plataformas sin necesidad. La nota queda como deuda de mantenimiento, no de este paquete.
- **`scripts/installer.ps1` es el espejo de `_installer-build`, no un camino nuevo**: mismo
  guard «una generación publicada NUNCA se pisa», mismo `bundle.py` (mismo sello RF-231),
  `Get-FileHash` en lugar de `sha256sum`, y **sin dev-sync** (en Windows no existe el override
  `~/.local/bin/arnesia` — DA-9: siempre el sidecar empaquetado). No bumpea: equivale a
  `make installer-actual`; el bump sigue siendo `bump.sh` en Linux (`bump.py` = deuda abierta).
- **CI `tauri-windows` solo en push a `main`**, no en cada PR: el bundle tarda decenas de
  minutos. El gate barato de compilación Windows sigue siendo `go-windows` (build+vet+tests
  de la superficie portada), que sí corre siempre.

## CW-D8 · El attach del shell se VERIFICA, no se adivina 🧑‍⚖️ RATIFICADA (2026-08-14)

**Bug reportado por el operador:** instaló el `.exe` y la ventana mostraba únicamente
`arnesia: este build no embebe la UI (scripts/bundle.sh la incluye); la API vive en /api`.

**Causa raíz (investigada, no supuesta):** el instalador estaba bien — el `arnesia-daemon.exe`
instalado es byte-idéntico (SHA256 `B54256E9…`) al que ya se verificó sirviendo la SPA. Quien
respondía era **otro daemon del propio operador corriendo en WSL** sobre el mismo repo
(`/mnt/c/…/bin/arnesia`, `version: dev`, huella `bf914df` del 2026-08-01), compilado con
`go build` suelto **cuando `web/dist` estaba vacío**: su `embeddedUI()` da `nil` y el router
contesta ese 404 (`router.go:66-68`). El shell lo tomó por «el daemon ya está arriba» porque
`daemon_running()` solo hacía `TcpStream::connect` — verificaba que **algo** escuchara, no que
ese algo sirviera la interfaz. Prueba de cierre: al detener ese proceso, la app instalada
levantó su propio sidecar y `GET /` pasó a 200 con la SPA.

**Decisión:** el bind-or-bail pasa de ciego a verificado. `estado_daemon()` sondea con un
`GET /` crudo (HTTP a mano sobre el mismo socket: es una línea de estado, no vale sumar un
cliente HTTP al shell) y `veredicto_daemon()` —regla pura, con 6 tests, misma doctrina que
`concede_captura`— decide entre `Ausente` (spawnear), `ConUI` (attach) y `SinUI`. En `SinUI`
no se attachea (mostraría la respuesta del ocupante) **ni** se intenta spawnear (el bind
moriría con «address already in use»): la ventana muestra una tarjeta que nombra el problema y
el paso siguiente. La verificación va en Rust y no en el JS de `conectando.html` porque un
`fetch` del WebView viaja con el `Origin` de Tauri y chocaría con los gates de `auth.go` —por
eso esa página sondea `/healthz`, el único exento (HS-14 fix ②)—. Sin botón «Reintentar»: el
shell decide al arrancar, y sin reiniciar no hay nada que cambie; ofrecerlo sería mentir.

**Corrección de una evidencia previa (honestidad):** la `dod_evidence` del 2026-08-13 decía
que se había ejercido la app lanzando `target/release/arnesia-app.exe`. Ese binario —producido
por `cargo build`— **apunta al `devUrl` (`http://localhost:5173`) y muestra un error de
conexión de WebView2**, no la app real: el binario válido lo produce `tauri build`, que embebe
los assets. Aquel chequeo solo probaba que el proceso levantaba y abría ventana; se reemplaza
por la verificación contra el binario bundleado.

## CW-D5 · Suite de tests en Windows: superficie portada verde, resto = DEUDA VISIBLE

`go build ./...` + `go vet ./...` verdes en windows/linux/darwin. Tests verdes en Windows:
`internal/adapters/selfupdate` (con tests por-OS nuevos), `cmd/arnesia`,
`internal/adapters/telemetria/store`, `docs/architecture/fitness` (con 2 skips honestos).
**Hallazgo mayor del triage (bug real, corregido):** los tests que «aislaban» el home con
`t.Setenv("HOME")` seguían leyendo/escribiendo el `~/.arnesia` REAL en Windows
(`os.UserHomeDir` lee USERPROFILE) — 12 sitios corregidos con `USERPROFILE`/`APPDATA`; la
contaminación detectada (`~/.arnesia/telemetria.db` creado por el run) se eliminó.

**Deuda registrada (clases de fallo restantes en Windows, por paquete):**

| Paquete | Clase de fallo | Naturaleza |
|---|---|---|
| `marketplace`, `publish`, `traer` | shims `gh`/`git` como scripts sh ejecutables | fixture unix — el runtime usa git/gh reales que SÍ existen en Windows |
| `traer` (copia), `store` (migración), `marketplace` (0000) | bits de ejecución / chmod 0555·0000 como fault-injection | semántica unix inexistente en NTFS |
| `transport/http` (forja/portafolio), `usecase` (conformance) | fixtures JSON que embeben paths con backslash (`C:\U…` = escape inválido) | fixture — construir el body con `json.Marshal` |
| `usecase` (traer frontera ~/.claude), `portafolio` (deriva/scanner), `stt` (rutas /usr/bin), `domain` (planificar-traer/rutas) | mezcla de literales `/` y semántica de árbol | requiere port por-caso |
| `fitness` `TestGoArchLintAdapterCorreDeVerdad` | go-arch-lint reporta violaciones FALSAS con paths backslash | skip honesto en windows; el gate corre en CI ubuntu; diagnosticar/upstreamear |
| `fitness` `TestLiveEventsFromStreamJSON` | fake `claude` = script sh | skip honesto en windows; port del fake (binario Go) pendiente |
| `scripts/estado.sh` (tooling, no test) | `conformance --todo` no termina en Windows (>10 min; probable raíz go-arch-lint) | regen de cifras en Linux/CI; `cap_doctor.py` con stdout UTF-8 forzado (crasheaba en cp1252) |

CI: job `go-windows` corre build+vet del árbol completo + tests de la superficie portada —
NUNCA un `|| true` sobre la suite completa (pass fabricado prohibido).

## CW-D6 · Diferidas explícitas (heredadas del plan)

- **DA-5** `~/.arnesia` → `os.UserConfigDir()` (D6 p.7 de telemetría): paquete propio con
  migración de datos (el store ya tiene maquinaria forward-only).
- **DA-6** goreleaser sin sello de identidad → reporta `dev` honesto; revisar al activar
  release pipeline.
- **DA-8/Parte B** migración de maquinaria de proceso (cockpit vendored :4300 + CIL + skills
  /po /architect /dev-team /auditor): plan podado vigente en `PLAN-cockpit-y-windows-arnesia.md`.
- `bump.py`/`installer.ps1`/Makefile despachador: junto con el paquete del instalador Windows.
