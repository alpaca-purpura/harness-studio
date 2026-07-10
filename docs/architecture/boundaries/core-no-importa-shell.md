---
regla: core-no-importa-shell
version: 1.3
updated: 2026-07-08
status: proposed
ledger: HS-04
sources:
  - url: https://alistair.cockburn.us/hexagonal-architecture/
    autoridad: experto
    revisado: 2026-07-05
  - url: https://v2.tauri.app/develop/sidecar/
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  # «core» = agregado conceptual de {domain, usecase, ports}; NO existe un componente `core` en
  # .go-arch-lint.yml — v3 es allow-list (default-deny): ningún componente del core lista un shell
  # en mayDependOn. El shell real es Rust (web/src-tauri/), fuera del grafo Go; su frontera la
  # guarda arch_test.go:TestCoreHasNoShellImport.
  - fitness/.go-arch-lint.yml#domain
  - fitness/.go-arch-lint.yml#usecase
  - fitness/.go-arch-lint.yml#ports
  - fitness/arch_test.go:TestCoreHasNoShellImport
  - fitness/arch_test.go:TestDaemonServableHeadless
  - fitness/arch_test.go:TestMintEnvInLauncher
severity: critical
---

# El daemon-core no depende del shell

## L1 · Principio (estándar de industria)

**Arquitectura hexagonal / ports & adapters (Cockburn).** El corazón de la aplicación
(dominio + casos de uso) no conoce sus mecanismos de entrega. La UI, el framework de escritorio
y el transporte son *adaptadores* que dependen del core — **nunca al revés**. Un core que no
importa a su envoltorio se puede servir por cualquier canal (CLI, HTTP, WebView nativo) sin
tocar una línea de dominio. *(experto: Cockburn, hexagonal)*

**Sidecar de escritorio (Tauri 2).** Tauri puede embeber un binario externo (`externalBin`) y
spawnearlo como *sidecar*; el WebView carga la SPA y consume la API del sidecar. El shell aporta
solo ventana/tray/updater/deep-link/firma — **cero lógica de dominio**. *(oficial: Tauri sidecar)*

## L2 · Realización (este árbol Go+React)

El `serve` de `arnesia` expone TODO el producto como **HTTP/SSE `:4200` + SPA `go:embed`**. Ese
contrato único es lo que vuelve el shell intercambiable (decisión HS-04: shell v1 = **Tauri 2**,
con el daemon como sidecar `externalBin`; el WebView apunta al daemon). ⇐ L1: sidecar.

- **El core (`internal/{domain,usecase,ports}`) NO importa ningún paquete `shell` ni bindings de
  Tauri/Wails.** El shell real es el crate Rust [`web/src-tauri/`](../../web/src-tauri/) (Tauri 2)
  — un proceso aparte, fuera del grafo de imports Go; el check guarda además que jamás aparezca
  un `shell` Go ni un binding (`tauri`, `wailsapp/wails`). El shell es un cliente del daemon, no
  al revés. ⇐ L1: hexagonal.
  Nota: «core» es un **agregado conceptual** de los componentes reales `{domain, usecase, ports}`
  de [`.go-arch-lint.yml`](../fitness/.go-arch-lint.yml) (no hay componente `core`); el boundary se
  enforça sobre esos tres + `arch_test.go:TestCoreHasNoShellImport`. Desde HS-10 el linter corre
  en CI (`go-arch-lint check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml`,
  deepScan off — el grafo de imports es el enforcement).
- El shell ([`web/src-tauri/`](../../web/src-tauri/) = crate Tauri 2 + launcher) **solo** conoce: cómo levantar/attach-ear el
  daemon (`attach si :4200 está arriba, si no spawnea`), setear el env de Mint
  (`WEBKIT_DISABLE_DMABUF_RENDERER=1`), y abrir el WebView. No consume el dominio directo.
- Consecuencia que protege la decisión de Tauri-desde-v1: **«servable headless» sigue gratis** —
  el mismo daemon corre sin shell (dev, CI, futuro server). Si mañana el shell fuera un browser o
  un Wails, el core no se entera.
- **Divergencia declarada frente al research:** el research recomendó browser-first por el riesgo
  WebKitGTK en Mint; el operador eligió Tauri-desde-v1 (app de escritorio nativa). Se acepta el
  riesgo y se vuelve load-bearing la mitigación de Mint (ver
  [`../../historias/2026-07-05-arquitectura-fase3.md`](../../historias/2026-07-05-arquitectura-fase3.md)
  frente A). El boundary NO cambia: sea cual sea el shell, no lo importa el core.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| core-no-shell-import | ningún paquete del core (`internal/{domain,usecase,ports}`) importa un paquete `shell` ni `tauri`/`wails` | error | «core importa el shell (acopla el daemon a su envoltorio)» | arch_test.go:TestCoreHasNoShellImport · go-arch-lint#{domain,usecase,ports} |
| shell-solo-composition | solo el shell/launcher levanta o attachea el daemon; el core no auto-lanza ventana | warn | «lógica de ventana en el core» | arch_test.go:TestCoreHasNoShellImport |
| daemon-servable-headless | existe un entrypoint `serve` que corre sin shell (test de humo) | error | «el daemon no arranca sin shell» | arch_test.go:TestDaemonServableHeadless |
| mint-env-en-launcher | el launcher setea `WEBKIT_DISABLE_DMABUF_RENDERER` (Tauri, Linux) | warn | banda Guardia «WebView Mint sin mitigación DMABUF» | arch_test.go:TestMintEnvInLauncher |
| single-instance-reenfoca | la 2a instancia no abre una ventana duplicada; reenfoca la existente (`unminimize`+`show`+`set_focus`) | warn | «reapertura tragada en silencio (sin ventana visible)» | web/src-tauri/src/lib.rs (revisión) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = hexagonal (Cockburn) + sidecar Tauri 2. L2
  amarra shell v1 = Tauri sobre el daemon-core `:4200`; divergencia declarada vs browser-first del
  research (operador eligió Tauri-desde-v1). 4 checks.
- 2026-07-07 · v1.1 · Sync HS-10: el shell real es el crate Rust `web/src-tauri/` (Tauri 2), no un
  `shell/` en la raíz — rutas de L2 corregidas; `fitness/.go-arch-lint.yml` corre en CI desde
  HS-10 (`--project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml`, deepScan off — el linter
  de imports es el enforcement; v3 allow-list, sin `cannotDependOn`). Sin cambios de checks ni de
  status.
- 2026-07-08 · v1.2 · HS-14 fix ③: el callback del plugin single-instance estaba vacío (`TODO`) —
  una 2a instancia se tragaba en silencio sin reenfocar la ventana existente. Implementado
  `get_webview_window("main")` + `unminimize`/`show`/`set_focus`; verificado en vivo (binario
  instalado, `gtk-launch` real, foco movido a otra ventana y devuelto tras relanzar, sin ventana
  duplicada). Check nuevo `single-instance-reenfoca` (revisión manual, mismo patrón que
  `mint-env-en-launcher`) → **5 checks**.
- 2026-07-08 · v1.3 · Auditoría colateral (HS-16): `daemon-servable-headless` y
  `mint-env-en-launcher` tenían enforcer bare `arch_test.go` (sin nombre de función) — el
  parser del ruleset solo reconoce `arch_test.go:TestX`, así que quedaban `deferred` pese a
  ser triviales de probar. `TestDaemonServableHeadless` (build+run real de `arnesia serve` como
  subproceso, HOME apuntado a un dir descartable, `GET /healthz` responde) y
  `TestMintEnvInLauncher` (grep sobre `web/src-tauri/src/main.rs`) los conectan a `--todo` sin
  lógica nueva — la feature ya existía en ambos casos.
