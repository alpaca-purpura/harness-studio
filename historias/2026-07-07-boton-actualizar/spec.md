# spec — botón «Actualizar» · self-update sin sudo (QUÉ)

> Ficha HS-11 (deuda del instalador) · paquete boton-actualizar · 2026-07-07
> Base firmada: decisiones #1–#3 (rumbo) y #5 (ubicación) sobre `mockup-actualizar.html`
> **v2 FIRMADO** (operador «firmo el mockup», 2026-07-07). Cada RF traza a `mockup:línea`
> de v2 (commit `923ec3a`). Numeración continúa la del Hito (inspector cerró en RF-96).

## Alcance

Self-update del binario instalado (daemon+UI embebida) desde el build local del repo,
en espacio de usuario (`~/.local/bin`), cero sudo. **NO incluye** (no-goals): releases
remotos/GitHub (tren KIT-06 — ni siquiera staged en la UI v1) · auto-check periódico de
versión · migración automática desde /usr/bin (manual, una vez, documentada) · resto de
tarjetas de Ajustes (marketplaces · daemon — futuras). Nada finge funcionar.

## RF — vista Ajustes

- **RF-100 (la vista nace).** El ⚙ del pie del rail deja de dar `ComingSoon`: la vista
  global Ajustes renderiza tarjetas; v1 = SOLO «Versión y actualización»
  (`mockup:223-247`, decisión #5) + una línea muted al pie con lo que llega después
  (marketplaces · daemon) — texto, jamás tarjetas vacías fingiendo. Breadcrumb/encabezado
  coherente con el shell (`mockup:239`).

## RF — tarjeta «Versión y actualización»

- **RF-101 (identidad honesta).** La tarjeta muestra: `daemon` = huella VCS embebida en
  el binario (revision corta + fecha; binario sin huella → «versión no embebida» dicho,
  `mockup:209`); `instalado en` = ruta REAL del ejecutable (`os.Executable()`) con pill
  de estado: verde «espacio de usuario · sin sudo» si el proceso puede escribirlo
  (`mockup:163`) / warn «root · actualizar pide sudo» si no (`mockup:210`); `origen` =
  ruta del repo configurada en el daemon, o «repo no configurado» (`mockup:164,211`).
  El dato viene de un endpoint nuevo `GET /api/version` — la UI jamás lo inventa.
- **RF-102 (no-actualizable: sin escritura).** Ejecutable NO escribible por el usuario
  (p.ej. `/usr/bin`, root) → botón disabled + warnbox con la migración de una sola vez
  (`install -m755 <repo>/bin/arnesia ~/.local/bin/`; PATH precede) (`mockup:207-216`).
- **RF-103 (no-actualizable: sin repo).** Daemon arrancado sin repo configurado →
  botón disabled + estado «repo no configurado» (`mockup:211`). La ruta del repo se
  configura AL DAEMON (flag `--repo` / env `ARNESIA_REPO`), jamás desde la UI/request.

## RF — el self-update

- **RF-104 (flujo de pasos).** `POST /api/self-update` ejecuta en orden (`mockup:141-147`):
  ① `verificar` — repo configurado, es el árbol esperado, toolchain presente (go · pnpm);
  ② `build` — `scripts/bundle.sh` (SPA + go build; huella VCS embebida);
  ③ `verificar binario` — existe · ejecutable · huella distinta de la corriente se REPORTA
  (igual huella = «ya estás al día», éxito sin reinstalar);
  ④ `instalar` — write a tmp EN EL MISMO directorio del ejecutable → `rename` atómico;
  ⑤ `reiniciar` — el daemon se re-ejecuta (mismo binario path, mismos args) TRAS responder.
  La respuesta 200 llega al terminar ④ con el veredicto de CADA paso (la UI pinta la
  checklist con datos reales, `mockup:151-157`); durante la espera la UI muestra
  «Actualizando…» bloqueado (`mockup:173-181`). Un paso fallido corta el flujo: los
  posteriores quedan «no corridos» y el error viaja con stderr/detalle (`mockup:195-203`)
  — el binario instalado NO se toca si ③ no pasó (`mockup:200-201`).
- **RF-105 (reinicio + reconexión).** Tras el 200 de «instalado», el daemon agenda el
  re-exec (≤1s) y la UI hace polling a `/healthz` + `GET /api/version` hasta ver la
  huella NUEVA → estado éxito «Actualizado a <huella> — daemon reiniciado, UI
  reconectada» (`mockup:184-191`). Timeout de polling → error honesto («el daemon no
  volvió — revisa el proceso»), jamás un éxito fingido.
- **RF-106 (seguridad, decisión #4 → cementada).** El endpoint vive BAJO `withAuth`
  (Host+Origin+token, superficie-local-confinada). El request NO lleva parámetros: ni
  rutas, ni flags — repo y destino los conoce SOLO el daemon (flag/env + `os.Executable()`).
  Concurrencia: un self-update a la vez (409 si hay uno en vuelo). El build corre con
  cwd = repo configurado y sin heredar credenciales extra.
- **RF-107 (endpoint de identidad).** `GET /api/version` → `{ huella, fecha, instalado_en,
  escribible, repo }` — la fuente de RF-101 y del polling de RF-105. Huella = VCS del
  buildinfo de Go (`vcs.revision`/`vcs.time`); binario sin VCS info → `"dev"` honesto.

## Gherkin

```gherkin
Feature: Ajustes · Versión y actualización
  Scenario: La vista nace con la tarjeta
    Given el shell abierto
    When clickeo ⚙ Ajustes en el pie del rail
    Then veo la tarjeta «Versión y actualización» con huella, ruta e instalación reales

  Scenario: Actualizar desde el repo (camino feliz)
    Given el binario instalado en ~/.local/bin (escribible) y el repo configurado
    When clickeo «Actualizar desde el repo»
    Then el botón queda «Actualizando…» y al terminar veo la checklist real de pasos
    And el daemon se reinicia y la tarjeta muestra la huella NUEVA — sin sudo en ningún paso

  Scenario: Build falla
    Given un repo que no compila
    When corro el self-update
    Then veo el paso build en rojo con su stderr y el binario instalado queda INTACTO

  Scenario: Binario no escribible (estado actual del operador)
    Given el binario en /usr/bin (dueño root)
    Then el botón está disabled y la warnbox explica la migración única a ~/.local/bin

  Scenario: Ya al día
    Given el repo en la misma huella que el binario corriendo
    When corro el self-update
    Then el flujo termina en «ya estás al día» sin reinstalar ni reiniciar
```

## Trazabilidad RF → destino en la app

| RF | mockup:línea (v2) | destino | story/test |
|---|---|---|---|
| RF-100 | 223-247 | `pages/shell/ui/global-view.tsx` (vista Ajustes real) | vista con tarjeta |
| RF-101 | 159-171 · 209-211 | `features/self-update/ui/update-card.tsx` (UI pura; la página inyecta datos/callbacks) + `GET /api/version` | identidad honesta |
| RF-102 | 207-216 | update-card (estado no-escribible) | disabled + migración |
| RF-103 | 211 | update-card (estado sin-repo) | disabled + honesto |
| RF-104 | 141-157 · 173-181 · 195-203 | Go: `ports.SelfUpdater` + `internal/adapters/selfupdate/` + `usecase.SelfUpdateService` + `POST /api/self-update` + OpenAPI | go test: pasos/fallos/atómico |
| RF-105 | 184-191 | update-card (polling healthz+version) + re-exec en daemon | éxito/timeout honesto |
| RF-106 | — (decisión #4) | withAuth + sin params + 409 concurrencia | go test |
| RF-107 | 161 · 188 | `GET /api/version` (buildinfo VCS) + client.ts | version honesta |

## Cierre

Gate del paquete = PARIDAD.md fila por fila + click-through real (daemon actualizándose
a sí mismo de verdad) con consola limpia. La firma de ESTE spec habilita el código.
