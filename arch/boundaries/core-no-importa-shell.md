---
regla: core-no-importa-shell
version: 1.0
updated: 2026-07-05
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
  # .go-arch-lint.yml — el boundary se enforça sobre esos tres (cada uno cannotDependOn: [shell]).
  - fitness/.go-arch-lint.yml#domain
  - fitness/.go-arch-lint.yml#usecase
  - fitness/.go-arch-lint.yml#ports
  - fitness/arch_test.go:TestCoreHasNoShellImport
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

- **`internal/core/**` (dominio + casos de uso + puertos) NO importa `internal/shell/**` ni
  ningún paquete de Tauri.** El shell es un cliente del daemon, no al revés. ⇐ L1: hexagonal.
  Nota: «core» es un **agregado conceptual** de los componentes reales `{domain, usecase, ports}`
  de [`.go-arch-lint.yml`](../fitness/.go-arch-lint.yml) (no hay componente `core`); el boundary se
  enforça sobre esos tres + `arch_test.go:TestCoreHasNoShellImport`.
- El shell (`shell/` = crate Tauri + launcher) **solo** conoce: cómo levantar/attach-ear el
  daemon (`attach si :4200 está arriba, si no spawnea`), setear el env de Mint
  (`WEBKIT_DISABLE_DMABUF_RENDERER=1`), y abrir el WebView. No consume el dominio directo.
- Consecuencia que protege la decisión de Tauri-desde-v1: **«servable headless» sigue gratis** —
  el mismo daemon corre sin shell (dev, CI, futuro server). Si mañana el shell fuera un browser o
  un Wails, el core no se entera.
- **Divergencia declarada frente al research:** el research recomendó browser-first por el riesgo
  WebKitGTK en Mint; el operador eligió Tauri-desde-v1 (app de escritorio nativa). Se acepta el
  riesgo y se vuelve load-bearing la mitigación de Mint (ver
  [`../../research/2026-07-05-arquitectura-fase3.md`](../../research/2026-07-05-arquitectura-fase3.md)
  frente A). El boundary NO cambia: sea cual sea el shell, no lo importa el core.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| core-no-shell-import | ningún paquete de `internal/core/**` importa `internal/shell/**` o `tauri` | error | «core importa el shell (acopla el daemon a su envoltorio)» | arch_test.go:TestCoreHasNoShellImport · go-arch-lint#{domain,usecase,ports} |
| shell-solo-composition | solo el shell/launcher levanta o attachea el daemon; el core no auto-lanza ventana | warn | «lógica de ventana en el core» | arch_test.go:TestCoreHasNoShellImport |
| daemon-servable-headless | existe un entrypoint `serve` que corre sin shell (test de humo) | error | «el daemon no arranca sin shell» | arch_test.go |
| mint-env-en-launcher | el launcher setea `WEBKIT_DISABLE_DMABUF_RENDERER` (Tauri, Linux) | warn | banda Guardia «WebView Mint sin mitigación DMABUF» | arch_test.go |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = hexagonal (Cockburn) + sidecar Tauri 2. L2
  amarra shell v1 = Tauri sobre el daemon-core `:4200`; divergencia declarada vs browser-first del
  research (operador eligió Tauri-desde-v1). 4 checks.
