# spec — fix: repo configurable para self-update (QUÉ)

> Paquete bugfix · 2026-07-13 · continúa la numeración RF del paquete `2026-07-07-boton-actualizar`
> (RF-100..107). Decisiones base: `decisiones.md` #1–#5, FIRMADAS 2026-07-13.

## Alcance

Cierra el hueco: una instalación empaquetada (Tauri/.deb) arranca el daemon sin
`--repo`/`ARNESIA_REPO` y no tiene forma de fijarlo salvo relanzar por CLI. Se agrega una
vía de CONFIGURACIÓN (persistida, validada, gated a gesto nativo) — separada del
mecanismo de update ya firmado (RF-104..107), que no cambia. **NO incluye** (no-goals):
releases remotos/GitHub (KIT-06, sigue fuera) · auto-clonar el repo si no existe
localmente (el operador sigue siendo dueño de tener el checkout) · exponer el path por
un input de texto libre (decisión #3).

## RF — persistencia + precedencia

- **RF-108 (repo persistido).** Nuevo store `~/.arnesia/self-update.json`
  (`{"repo": "<path>"}`, escritura atómica tmp+rename — mismo patrón que
  `portafolio.json`/`arneses.json`). Al boot (`cmd/arnesia/main.go` runServe): si
  `--repo`/`ARNESIA_REPO` viene vacío, se lee el store; si trae un valor, se usa como
  repo inicial del `Updater`. Ausencia de ambos → estado `sinRepo` de siempre (RF-103
  intacto, cero regresión).

## RF — endpoint de configuración

- **RF-109 (`PUT /api/self-update/repo`).** Body `{"path": "<abs>"}`. Bajo `withAuth`
  (mismo confinamiento que todo `/api/*`). Valida con las MISMAS reglas del paso
  Verificar (RF-104 ①: repo existe · `go.mod` módulo esperado · `scripts/bundle.sh` ·
  toolchain go/pnpm/bash) ANTES de aceptar. Válido → fija el repo en caliente en el
  `Updater` corriendo (sin reiniciar el daemon) Y lo persiste (RF-108); responde 200 con
  el detalle de la validación. Inválido → 400 con el motivo exacto, NO persiste, NO
  toca el repo activo. `POST /api/self-update` (RF-104/RF-106) **no cambia una línea**:
  sigue ignorando el body por completo — la superficie que ejecuta código no gana
  parámetros nuevos.

## RF — UI

- **RF-110 (selector nativo en Ajustes).** La tarjeta «Versión y actualización», en el
  estado `sinRepo`, gana un botón «Elegir carpeta…» que abre el diálogo nativo de
  carpeta de Tauri (`@tauri-apps/plugin-dialog`, `open({directory:true})`) — SOLO si la
  app corre dentro de Tauri (detectado en runtime); en un browser plano el botón no
  aparece y el mensaje actual («repo no configurado (arranca sin --repo)») queda tal
  cual, apuntando a la vía CLI para dev. Tras elegir carpeta: `PUT
  /api/self-update/repo`; éxito → refresca `GET /api/version` (la tarjeta pasa a
  identidad honesta normal, RF-101); error de validación → mensaje inline con el motivo
  exacto del servidor (nunca un `alert()` ni un fallo mudo).

## RF — seguridad

- **RF-111 (reafirma RF-106, no lo reabre).** `PUT /api/self-update/repo` vive bajo el
  mismo `withAuth` que el resto de `/api/*` (Host+Origin+token). El path SOLO llega por
  este endpoint dedicado — nunca por `POST /api/self-update`, que sigue con cero
  parámetros. Un path que falla la validación jamás se persiste ni se activa (no hay
  ventana de repo-a-medio-configurar).

## Gherkin

```gherkin
Feature: Ajustes · configurar repo del self-update
  Scenario: Instalación empaquetada sin repo, primera vez
    Given el daemon arrancó sin --repo/ARNESIA_REPO y sin store persistido
    When abro Ajustes en la app Tauri
    Then veo «repo no configurado» + botón «Elegir carpeta…»

  Scenario: Configurar un repo válido
    Given elijo con el diálogo nativo un directorio que ES el árbol arnesia
    When se dispara el PUT /api/self-update/repo
    Then la tarjeta pasa a identidad honesta (huella · ruta · repo) y «Actualizar» se habilita

  Scenario: Configurar un repo inválido
    Given elijo un directorio que NO es el módulo esperado (o le falta bundle.sh/toolchain)
    When se dispara el PUT /api/self-update/repo
    Then veo el motivo exacto inline, el repo NO se persiste, «Actualizar» sigue disabled

  Scenario: Persistencia sobrevive un reinicio
    Given configuré un repo válido en una sesión previa
    When el daemon reinicia SIN --repo/ARNESIA_REPO
    Then arranca con el repo persistido — sin volver a configurarlo

  Scenario: Browser plano (dev, sin Tauri)
    Given abro Ajustes en un browser normal (no la app de escritorio)
    Then NO veo el botón «Elegir carpeta…» — el mensaje sigue apuntando a --repo/ARNESIA_REPO
```

## Trazabilidad RF → destino en la app

| RF | destino | story/test |
|---|---|---|
| RF-108 | `internal/adapters/selfupdate/repo_store.go` (nuevo) + `cmd/arnesia/main.go` runServe (precedencia) | go test store atómico + boot con/sin persistido |
| RF-109 | `internal/ports/selfupdate.go` (`ConfigurarRepo`) + `internal/adapters/selfupdate/updater.go` + `internal/usecase/selfupdate_service.go` + `internal/adapters/transport/http/selfupdate.go` (`PUT /api/self-update/repo`) + `router.go` | go test -race: válido/inválido/persistencia |
| RF-110 | `web/src/features/self-update/ui/update-card.tsx` (prop `onElegirRepo?`) + `web/src/pages/shell/ui/global-view.tsx` (detecta Tauri, llama al plugin-dialog, transporta) | story UpdateCard con/sin `onElegirRepo`; verificación manual en la app Tauri (gate humano — requiere display) |
| RF-111 | `withAuth` existente + go test (400 inválido no persiste · 200 válido persiste+activa) | go test |

## Cierre

Gate del paquete = PARIDAD.md + verificación E2E del lado Go (daemon real, PUT con
repo válido/inválido, reinicio con persistido) + revisión humana del lado Tauri (picker
nativo, requiere sesión con display — no automatizable headless). La firma de ESTE spec
habilita el código.
