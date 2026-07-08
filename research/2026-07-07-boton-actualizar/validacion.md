# Validación REAL — FASE 3 (2026-07-07/08) · COMPLETA, todos los veredictos reales

> Este archivo nació como el «commit trivial» del guion (PROMPT §FASE 3): el binario
> instalado corría la huella del commit anterior; el commit que lo creó (`0a42644`) fue
> la delta que el botón compiló, instaló y confirmó tras el reinicio. Resultados abajo.

## Migración única del operador (documentada, PROMPT §FASE 3 ①)

```bash
scripts/bundle.sh --daemon-only          # SPA + go build → bin/arnesia (sin rustc)
install -m755 bin/arnesia ~/.local/bin/  # UNA sola vez; ~/.local/bin precede en PATH
~/.local/bin/arnesia serve --repo /home/chalreme/Proyectos/harness-studio
```

Estado previo real: `/usr/bin/arnesia` (root, .deb del 13:37) — queda como instalación
inicial de terceros; el ciclo diario ya no lo toca. Desde la migración: cero sudo.

## Camino feliz — el daemon se actualizó A SÍ MISMO ✅

| Evidencia | Valor real |
|---|---|
| Binario instalado ANTES | huella `ed223a4` · 2026-07-08 · `vcs.modified=false` |
| Commit trivial (delta) | `0a42644` (este archivo) |
| Click «Actualizar desde el repo» | checklist REAL: verificar ✓ · build ✓ (`bundle.sh --daemon-only OK`) · verificar binario ✓ (`huella 0a42644 · 10010889 bytes`) · instalar ✓ (`rename atómico, sin sudo`) · reiniciar agendado |
| Reinicio | re-exec in situ — **MISMO PID 930295** (syscall.Exec), la UI reconectó por polling `/api/version` |
| Tarjeta tras reconectar | okbox «Actualizado a 0a42644 — daemon reiniciado, UI reconectada.» + huella nueva en el kv |
| `go version -m ~/.local/bin/arnesia` | `vcs.revision=0a42644…` ✓ |

## Contracasos (todos con veredicto real)

| Caso | Cómo se indujo | Veredicto observado |
|---|---|---|
| **409 doble click** | 2º POST disparado DURANTE el build del 1º | HTTP **409** ✓; lock retenido tras «actualizado» (decisión #9) — go test lo cubre también |
| **Ya al día** | click sin delta nueva | okbox «Ya estás al día — la huella del repo (0a42644) es la del binario corriendo.»; instalar/reiniciar «no corrido»; SIN reinstalar, SIN reinicio; consola 0 |
| **Build roto** | `const rotoAProposito: number = "…"` en `web/src/shared/lib/cn.ts` | paso build **fallo** con `error TS2322` + stderr completo en critbox; verificar-binario/instalar/reiniciar «no corrido»; botón «Reintentar»; binario instalado INTACTO (`0a42644`); consola 0. Archivo restaurado con `git checkout` |
| **No escribible** | `chmod 555 ~/.local/bin` (y restaurado) | pill warn + warnbox con `install -m755 …` + botón disabled title «Migra a ~/.local/bin…»; POST directo → **503** con motivo; consola 0 |
| **Sin --repo** | daemon reiniciado sin flag | origen «repo no configurado (arranca sin --repo)» + disabled title «Configura --repo / ARNESIA_REPO…»; POST → **503** con motivo RF-103; consola 0 |
| **Gates S1 (RF-106)** | curl `Host: evil.com` / `Origin: https://evil.com` | **403** ambos; body del POST (`{"repo":"/tmp/evil"}`) IGNORADO — go test `TestPostSelfUpdateIgnoraParametrosDelRequest` |

## Consola

- Escenarios idle · ya-al-día · build-roto · no-escribible · sin-repo: **0 mensajes**.
- Update real: 2 líneas explicadas (el 409 del propio test doble-POST + el corte del SSE
  por el re-exec) — detalle en PARIDAD §Desviaciones #2.

## Gates de código (todos verdes al cierre)

`tsc` ✓ · `biome ci src` ✓ · `depcruise` ✓ (84 módulos, 0 violaciones) · `steiger` ✓ ·
`stylelint` ✓ · story-tests **71/71** ✓ (8 stories nuevas de update-card) ·
`go build/vet` ✓ · `go test -race ./...` ✓ (adapter 15 tests · usecase 6 · transport 4) ·
`golangci-lint` **0 issues** · `golangci-lint fmt --diff` limpio · `go-arch-lint` OK
(componente `selfupdate` añadido).

**CI de GitHub: 3/3 verde** (run 28915038317, commit `c821665`). El primer push cayó en
el job go — `TestVerificar/repo_completo` asumía `pnpm` en el PATH del runner; corregido
con stub de toolchain hermético (`c821665`), y ganó un caso nuevo «toolchain incompleta».

## Driver

Playwright headless (Chromium) — fallback documentado (PARIDAD §Desviaciones #1; el
browser MCP quedó tomado por el WM, igual que en el click-through del mockup v2).
Script: `valida.mjs` (scratchpad de la sesión) — escenarios idle · update · ya-al-dia ·
build-roto · no-escribible · sin-repo · mockup; screenshots `shots/*.png` revisados
lado a lado contra `mockup-caso-01..05.png`.

## Hallazgo post-validación (2026-07-07 noche) — colisión de nombre rompe el ícono del escritorio

**Síntoma reportado por el operador:** click en el ícono ArnesIA → no abre ventana, sin
error visible. El daemon SÍ estaba vivo (`:4200` respondía 200 en 6ms).

**Causa raíz (verificada):** el `.desktop` del .deb dice `Exec=arnesia` SIN ruta absoluta.
La migración de esta feature instaló el daemon Go como `~/.local/bin/arnesia`, y el PATH
de la sesión (`systemctl --user show-environment`) pone `~/.local/bin` ANTES de
`/usr/bin` → el ícono ejecuta el **daemon Go sin argumentos** (imprime usage y sale, con
`Terminal=false` nadie lo ve) en vez del **shell Tauri** `/usr/bin/arnesia`. La línea 15
de este archivo («queda como instalación inicial; el ciclo diario ya no lo toca») no vio
que el binario del shell en el .deb TAMBIÉN se llama `arnesia`.

**Remediación local aplicada (fuera del repo, reversible):**
- `~/.local/share/applications/ArnesIA.desktop` — override user-level (precede al de
  `/usr/share/applications`), `desktop-file-validate` OK.
- `~/.local/bin/arnesia-shell-launcher` — Exec absoluto a `/usr/bin/arnesia` + **log de
  cada lanzamiento** en `~/.arnesia/logs/shell.log` (timestamp, DISPLAY, sesión; stdout+
  stderr del shell; rotación simple a 1MB) → responde el pedido del operador de un
  «monitoreador» para diagnosticar arranques fallidos.
- Verificado E2E: launcher → ventana `arnesia.Arnesia ArnesIA` abierta, proceso
  `/usr/bin/arnesia` vivo, lanzamiento registrado en el log.

**Observación adicional:** al reusar el daemon externo el shell loguea
`[arnesia] daemon ya activo en 127.0.0.1:4200; no spawneo (dev: WebView sin token)` —
funcionó, pero la ruta «shell sin token contra daemon pre-existente» quedó ejercitada
solo de facto.

**Deuda de producto (candidata a desviación #7 para el gate final):** la colisión
`arnesia` (shell Tauri del bundle) ↔ `arnesia` (daemon del self-update) es del diseño,
no de esta máquina. Opciones para el operador: (a) el bundle genera el `.desktop` con
Exec absoluto (Tauri `desktopTemplate`), (b) renombrar uno de los dos binarios,
(c) self-update instala como `arnesia-daemon` alineado con el sidecar del .deb.

## Hallazgo #2 post-validación (2026-07-07 noche) — el shell embebe SPA propia congelada; self-update NO le llega a la UI del escritorio

**Síntoma reportado por el operador:** «no veo las mejoras del drawer supuestamente
implementadas» — abría el shell de escritorio y la UI seguía sin el inspector-drawer
(RF-80..96, commits `6e61b30..1c7443f`, 17:58–19:15).

**Causa raíz (verificada binario contra binario):** hay DOS copias de la SPA y el shell
usa la suya:
- Shell Tauri (`/usr/bin/arnesia`, .deb del **13:37**): `frontendDist: "../dist"` sin
  `url` en la ventana → el WebView carga la SPA **embebida en el .deb al momento del
  bundle** — `index-BEoG7Efy.js`, **cero** ocurrencias de `arnesia-inspector` (el drawer
  aún no existía; tampoco la tarjeta Ajustes/self-update).
- Daemon (`~/.local/bin/arnesia`, huella `0a42644`): sirve la SPA NUEVA
  (`index-IpYXxoWE.js` — drawer ✓, tarjeta self-update ✓) en `http://127.0.0.1:4200/`.
- El shell solo consume la **API** del daemon (`client.ts` BASE `127.0.0.1:4200`):
  datos vivos, código de UI congelado.

**Consecuencia de diseño:** el botón «Actualizar desde el repo» actualiza daemon + SPA
del daemon, pero la UI del escritorio JAMÁS recibe mejoras sin re-empaquetar e instalar
el .deb (sudo). Peor: el shell congelado ni siquiera muestra la tarjeta de actualización.

**Remediación inmediata (aplicada):** abrir `http://127.0.0.1:4200/` en el browser —
ahí vive la SPA actual con el drawer.

**Deuda de producto (candidata a desviación #8 para el gate final) — opciones:**
(a) la ventana del shell navega a `http://127.0.0.1:4200` (WebviewUrl externa; la SPA
    embebida del .deb queda solo como fallback/urls de error) → toda mejora llega con el
    self-update del daemon; revisar gate S1 Host+Origin para el origin del WebView;
(b) self-update también reempaqueta el shell (pesado: rustc + sudo dpkg — contradice
    «cero sudo»);
(c) aceptar el doble-cuerpo y mostrar skew de versión shell↔daemon en Ajustes.
La (a) alinea con HS-04 («el WebView consume la misma API») y con el valor de esta
feature; decisión del operador en el gate.

### Recomendación técnica para el gate de la candidata #8 (argumentada, 2026-07-07)

**Recomendación: (a)** — la ventana del shell navega a `http://127.0.0.1:4200`; la SPA
embebida del .deb queda reducida a página de fallback («conectando…» / «el daemon no
arrancó»). Razones:

1. **Elimina la CLASE de fallo, no la instancia.** Dos copias de la SPA = skew
   UI↔API garantizado por construcción y una matriz de compatibilidad permanente.
   Con (a) hay UN cuerpo desplegable para todo lo que cambia semanalmente (UI+API):
   el binario Go, con update atómico rename+re-exec YA probado. El shell solo cambia
   cuando cambia el shell (ventana · sidecar · mitigaciones webkit) — churn bajísimo.
2. **Se shippea lo que se valida.** Hoy la validación E2E (Playwright contra `:4200`)
   ejercita una SPA que el usuario de escritorio NO ve. (a) cierra ese hueco de
   integridad de testing.
3. **Es el modelo de la topología que ya firmamos.** Syncthing —inspiración declarada
   en HS-02— sirve su UI desde el daemon; Tailscale/code-server igual. El patrón
   Electron de «actualizar el app entero» (b) existe para apps SIN daemon que sirva
   UI; nosotros ya lo tenemos — duplicarle el trabajo es desperdicio.
4. **(b) rompe compromisos firmados:** sudo (contradice «cero sudo» de esta feature),
   rustc en la máquina del usuario (bundle.sh --daemon-only lo evita a propósito),
   minutos de link de Rust por update. Resuelve instancias por fuerza bruta.
5. **(c) institucionaliza el defecto:** el banner de skew es una disculpa permanente
   + carga de soporte. Parche vestido de feature.

**Costos reales de (a) (verificados contra el código, acotados):**
- **Token bootstrap:** único IPC de la SPA = `invoke('auth_token')` (`client.ts:133`).
  Con URL remota: capability `remote` para ese dominio, o MEJOR eliminar el IPC —
  `initialization_script` del shell inyecta el token como global; browser-dev queda
  igual (global ausente → tokenless). El shell SIGUE siendo raíz de confianza
  (boundary `superficie-local-confinada` HS-06 se revisa, no se rompe).
- **Secuencia de arranque:** el shell YA sondea `:4200` antes de spawnear
  (`lib.rs:60`); navegar tras health-check; fallback embebido si el daemon muere.
- **Cache:** el daemon sirve `index.html` con no-cache (assets Vite ya van hasheados).

Complemento futuro (no ahora): Tauri updater oficial para los cambios raros del
shell binario. Ortogonal a (a).

### Candidata #8 — FIRMADA (a) POR EL OPERADOR Y EJECUTADA (2026-07-07 noche)

Operador: «procede con la opción (a) de una vez». Implementado y verificado E2E:

- **`web/public/conectando.html`** (nueva): página de arranque embebida — sondea
  `GET /api/version` cada 300ms y `location.replace` a `http://127.0.0.1:4200/`;
  a los 15s muestra fallo honesto (apunta a `~/.arnesia/logs/shell.log`) pero sigue
  sondeando. Valores de tokens DTCG resueltos de `web/tokens/base.tokens.json`.
- **`lib.rs`**: ventana programática (sale de `tauri.conf.json`) que nace en
  `conectando.html`; el token viaja al WebView por `initialization_script`
  (`window.__ARNESIA_TOKEN__`, corre en cada documento incluido el origin del daemon)
  SOLO en modo spawn; **muere el comando IPC `auth_token`** (habría exigido abrir IPC
  a origin remoto). Attach queda igual (sin token, Host+Origin).
- **`client.ts`**: `fetchAuthToken` lee el global; fuera `@tauri-apps/api` y
  `@tauri-apps/plugin-shell` de package.json (la SPA ya no toca Tauri — corre idéntica
  en browser y shell).
- **`cmd/arnesia/main.go`**: `Cache-Control: no-cache` en `index.html` (raíz y
  fallback SPA) — tras un self-update la próxima carga recoge los assets nuevos.
- Versión shell 0.1.0 → **0.2.0** (tauri.conf + Cargo.toml).

**Gates:** web `verify` ✓ (tsc·biome·depcruise·steiger·stylelint) · stories **71/71** ✓
· `go build/vet/test` ✓ · `golangci-lint` 0 · `cargo fmt/clippy` 0 · build release ✓.

**E2E real (attach):** binario release (`--features tauri/custom-protocol`; sin la
feature, cargo usa `devUrl` — hallazgo del guion: el CLI de tauri la agrega solo) →
ventana `arnesia.Arnesia` → conectando → **SPA del daemon renderizada EN EL SHELL:
inspector-drawer visible (estado vacío RF-84) + Ajustes en el rail + Mapa dogfood**;
screenshot `shell-e2e5.png` (scratchpad sesión). La ruta spawn+token queda para el
E2E del .deb instalado (al boot :4200 libre → sidecar con token).

**Pendiente del operador:** `sudo dpkg -i` del .deb 0.2.0 (el shell instalado en
`/usr/bin` es root; único paso con sudo, como toda instalación de paquete del sistema).
