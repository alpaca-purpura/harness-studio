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
