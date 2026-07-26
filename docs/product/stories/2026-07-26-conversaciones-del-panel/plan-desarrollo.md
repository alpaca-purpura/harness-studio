# Plan de desarrollo · Las conversaciones viven en el panel de conversación

> `tipo: plan` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Ejecuta [`arquitectura.md`](./arquitectura.md) contra [`spec.md`](./spec.md) (RF-300…RF-357),
> [`design.md`](./design.md) (C-1…C-13) y [`plan-storybook.md`](./plan-storybook.md) (las 56 stories).
> Ley: `decisiones.md` **CV-D1…CV-D16 firmadas** + F-1…F-4.
>
> **33 tickets · 6 tramos.** El orden es por **dependencia**, no por preferencia. Un tramo no
> arranca hasta que el anterior está verde. Próxima capability libre: **CAP-140** (verificado:
> `grep -rhoE "CAP-[0-9]+" docs/ | sort -V | tail -1` → CAP-139).
>
> **Comandos que se repiten** (verificados en `Makefile`, `web/package.json`, `ci.yml`):
> `go test ./...` · `go test ./docs/architecture/fitness/ -run TestX` ·
> `pnpm --dir web exec vitest --project=unit run` ·
> `pnpm --dir web exec vitest --project=storybook run [archivo]` · `pnpm --dir web run verify`.
> ⚠ `pnpm run verify` **NO corre tests** (`package.json:25`) y **no hay target de tests en el
> Makefile**: los tres se corren a mano.
>
> **Regla que atraviesa los 33:** ningún commit toca fuente sin construir o modificar un capability
> (`codigo-traza-a-capability`, R2 `error`). Y toda decisión conversada se escribe en
> `decisiones.md` **en el mismo turno** (METODOLOGIA §10).

---

## Tramo 0 · Llegar a verde y quedarse (problema F) — T1…T6

**Sin esto no hay línea base.** CI rojo desde 2026-07-20 (19 de las últimas 20), 79 commits sin
pushear, lefthook sin instalar, y `chat-dock.tsx` sin story. **Nada del tramo 1 empieza antes.**

---

### T1 · Instalar el gate local

- **Qué:** `lefthook install` en este working copy.
- **Por qué:** `.git/hooks/` tiene sólo los 14 `.sample` y `core.hooksPath` está sin setear
  (verificado): los **8 jobs pre-commit nunca corrieron acá**, incluido `capabilities` (R3) y
  `estado-cifras`. Todo lo que este paquete rompe se descubriría en CI en vez de al commitear.
- **Archivos:** ninguno del repo. Es entorno.
- **Diseño:** `lefthook install`; verificar con `ls .git/hooks/pre-commit` y `git config core.hooksPath`.
- **Capability:** ninguna — no toca fuente. **Es la única excepción del plan y por eso está sola.**
- **Verificación:** `ls .git/hooks/ | grep -v sample` devuelve `pre-commit` y `commit-msg`.
- **Validación:** un commit de prueba dispara `capabilities` (~4 ms) y `estado-cifras` (~20-50 s).
- **DoD:** ☐ hooks presentes ☐ un commit real los dispara ☐ `estado.sh --check` corre y no rompe.
- **Riesgo:** `estado-cifras` tarda 20-50 s por commit. **Reversión:** `lefthook uninstall`.

---

### T2 · 🔴 Story del `ChatDock` VIGENTE — bloqueante (H-1)

- **Qué:** crear `chat-dock.stories.tsx` con 3 stories del dock **tal como está hoy**, y que pasen
  en verde **antes del primer cambio**.
- **Por qué:** H-1 / `plan-storybook.md` §1.3. El widget que este paquete reescribe entero
  (375 líneas) no tiene baseline: sin él, no hay forma de probar que el superset (BR-CV-11) no rompió
  el composer, las burbujas ni la tarjeta de actividad.
- **Archivos:** `web/src/widgets/chat-dock/ui/chat-dock.stories.tsx` (**nuevo**).
- **Diseño:** patrón 42/42 del repo, sin inventar nada:
  ```tsx
  import type { Meta, StoryObj } from "@storybook/react-vite"
  import { expect, fn, userEvent, within } from "storybook/test"   // NO "@storybook/test": no existe
  const escenario = (parcial: Partial<SessionsState>) => (Story: () => ReactNode) => { … }
  ```
  El decorator escribe el estado **entero** de `useSessions`, nunca un merge parcial — calca
  `dictado-button.stories.tsx:16-29`, el único archivo del repo que toca un store zustand en stories.
  Stories: **A-01 `VigenteReposo`** (asserta las 4 filas de cromo: header · `◍` de `SessionLine` ·
  `Alcance:` del `ScopeRow` · composer) · **A-02 `VigenteConPermiso`** · **A-03 `VigenteStreaming`**.
- **Capability:** modifica **CAP-68** (`fe-chat/chat-cc.yaml`, hoy `vivo·nc` con `valida: []`) →
  suma `valida: [chat-dock.stories.tsx]` y **R4 exige flipear `status` a `vivo` en el mismo commit**.
- **Verificación:** `pnpm --dir web exec vitest --project=storybook run src/widgets/chat-dock/ui/chat-dock.stories.tsx`
  → 3/3. Cubre el **baseline**, no un escenario E-nn.
- **Validación:** las 3 describen el dock que el operador ve hoy; una captura contra la app viva
  confirma la equivalencia.
- **DoD:** ☐ 3 verdes ☐ ninguna baja `a11y` a `"todo"` (`grep -n a11y` en el archivo) ☐ CAP-68
  `vivo` con su `valida` ☐ `TestCapabilityStatusConsistent` verde.
- **Riesgo:** el dock lee `useSessions` directo (`chat-dock.tsx:13-15`) y el decorator puede filtrar
  estado entre stories. **Mitigación:** estado entero, siempre. **Reversión:** borrar el archivo.

---

### T3 · CI a verde: aplicar C-3 a los dos `text-warn` del picker

- **Qué:** cambiar `text-warn` → `text-foreground` + señal `--warn` **no textual** (borde izquierdo
  3px) en `new-session-picker.tsx:273` y `:290`.
- **Por qué:** es **la causa medida** del CI rojo: `#c96a2e` sobre `#ffffff` a 10px = **3,76:1**,
  bajo el mínimo 4,5 de axe, rompiendo 4 stories (`BACKLOG.md:207-211`). Es exactamente la
  realización de **C-3** que `design.md` ya decidió para la superficie nueva, aplicada al consumidor
  que la rompe hoy. **No se espera a RF-333** (tramo 5): eso dejaría CI rojo 5 tramos.
- **Archivos:** `web/src/widgets/session-rail/ui/new-session-picker.tsx` (2 líneas).
- **Diseño:** `className="text-warn"` → `className="border-warn border-l-[3px] pl-1.5 text-foreground"`.
  Contraste resultante: `--foreground`/`--card` = **18,74:1** claro / 16,09 oscuro. La señal de
  alarma la da el borde (no textual, 3,76 ≥ 3 ✓).
- **Capability:** modifica **CAP-72** (`fe-shell/rail-de-sesiones.yaml`) — nota de que el error de
  historial deja de usar `--warn` como color de texto.
- **Verificación:** `pnpm --dir web exec vitest --project=storybook run src/widgets/session-rail/ui/new-session-picker.stories.tsx`
  → **7/7** (hoy 3/7). Cubre el desbloqueo de todo el plan.
- **Validación:** `pnpm --dir web exec vitest --project=storybook run` completo → **0 failed**.
- **DoD:** ☐ las 4 stories que fallaban pasan ☐ **la deuda del token sigue abierta** en
  `BACKLOG.md:59-64` y se anota en `PARIDAD.md` que pasaron **por contraste corregido, no por el
  token** ☐ ninguna story bajó `a11y`.
- **Riesgo:** ninguno funcional (es color). **Reversión:** revertir 2 líneas.

---

### T4 · Declarar el OpenAPI como ES HOY + la allowlist de deuda

- **Qué:** documentar en `openapi.yaml` lo que el router sirve hoy y **no** está declarado, o
  ponerlo en `_sin-declarar.yaml` con razón.
- **Por qué:** no se puede poner un gate de drift sin sacar primero el drift de la ecuación. Medido
  hoy (generado): **50 rutas servidas · 39 declaradas · 11 `/api` sin declarar · 0 fantasma**.
- **Archivos:** `docs/architecture/contracts/api/openapi.yaml` ·
  `docs/architecture/contracts/api/_sin-declarar.yaml` (**nuevo**).
- **Diseño:** las 11 entran a `_sin-declarar.yaml`, cada una con `ruta:` + `razon:` **obligatoria**:
  las 7 de `telemetria/` («el módulo nace en `stories/2026-07-24-telemetria-embebida-otel/`, declara
  su contrato al cerrar»), `GET /api/sessions/cerradas/{id}/historial` («se retira en T18 de este
  paquete»), `POST /api/portafolio/arneses/{clave}/identificar`, `GET /healthz`,
  `DELETE /api/telemetria/arneses/{clave}`. Y las 4 estructurales: `POST /v1/logs`,
  `POST /v1/metrics` (contrato OTLP ajeno), `GET /events` (alias sin prefijo), `/` (SPA).
- **Capability:** modifica **CAP-52** (`http-sse/superficie-rest.yaml`) — el número de endpoints y
  la deuda contable.
- **Verificación:** T5 la mide; acá se prepara el terreno.
- **Validación:** el archivo lista 11 + 4 y **ninguna razón vacía**.
- **DoD:** ☐ 15 entradas con razón ☐ ninguna ruta servida queda sin declarar ni sin exención.
- **Riesgo:** una entrada con razón floja se vuelve permiso permanente. **Mitigación:** el ratchet de
  T5 impide que crezca. **Reversión:** borrar el archivo.

---

### T5 · El enforcer de drift del OpenAPI (problema F, segunda mitad)

- **Qué:** escribir los 4 enforcers del boundary `ruta-servida-esta-declarada`.
- **Por qué:** el único paso de contrato de CI está **inerte** (`ci.yml:73`, condicional sobre
  `web/src/shared/api/generated/**`, que no existe). Sin este gate, las 5 rutas del tramo 2 nacen por
  el mismo agujero por el que se fueron las 11.
- **Archivos:** `docs/architecture/fitness/openapi_contract_test.go` (**nuevo**) ·
  `docs/architecture/boundaries/ruta-servida-esta-declarada.md` (`proposed` → **`enforced`**) ·
  `docs/architecture/INDEX.md` (la fila).
- **Diseño:**
  ```go
  func rutasDelRouter(t *testing.T) map[string]bool   // regex sobre mux.Handle( Y mux.HandleFunc(
  func rutasDelContrato(t *testing.T) map[string]bool // yaml.v3 sobre paths: × métodos, base /api
  func exenciones(t *testing.T) map[string]string     // _sin-declarar.yaml → ruta:razon
  func TestRutaServidaEstaDeclarada(t *testing.T)
  func TestRutaDeclaradaSeSirve(t *testing.T)
  func TestQueryParamEstaDeclarado(t *testing.T)      // r.URL.Query().Get("x") ⟷ parameters:
  func TestExencionDeContratoTieneRazon(t *testing.T)
  ```
  ⚠ **Escanear los dos verbos.** Mirar sólo `HandleFunc` deja `GET /api/events` fuera
  (`router.go:67` usa `mux.Handle`) y produce un falso «declarada sin servir». Verificado.
  `gopkg.in/yaml.v3` ya es dependencia directa (`go.mod`): cero peso nuevo, y es código de test
  (`peso-del-binario-es-presupuesto` satisfecho).
- **Capability:** **nueva CAP-146** `http-sse/contrato-declarado.yaml` (el enforcer es código y
  necesita hoja) — o entrada en `_coverage.yaml` si se prefiere tratarlo como fitness puro;
  **decidir explícitamente, no por omisión**.
- **Verificación:** `go test ./docs/architecture/fitness/ -run 'TestRuta|TestQueryParam|TestExencion'`
  → 4/4 verdes contra el árbol de hoy.
- **Validación:** agregar a mano una ruta falsa en `router.go` ⇒ `TestRutaServidaEstaDeclarada`
  **falla**; quitarla ⇒ verde. Sin esa prueba, el enforcer no está probado.
- **DoD:** ☐ 4 verdes ☐ el fallo inducido rompe ☐ boundary a `enforced` con su changelog ☐ INDEX
  actualizado ☐ `_sin-declarar.yaml` con 15 entradas.
- **Riesgo:** falsos positivos por una forma de registro no contemplada. **Mitigación:** la
  validación por fallo inducido + la allowlist. **Reversión:** borrar el test y volver el boundary a
  `proposed`.

---

### T6 · 🧑‍⚖️ Gate de línea base: gate local completo → push → CI verde

- **Qué:** correr el gate entero local, pushear los 79 commits, **confirmar la corrida de CI en
  verde**.
- **Por qué:** el paquete no empieza sobre un árbol rojo. Y CI no vio el trabajo de los últimos días.
- **Archivos:** ninguno.
- **Diseño:** en este orden — `go test ./... -race` · `go run github.com/fe3dback/go-arch-lint@latest check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml` ·
  `bash scripts/estado.sh --check` · `pnpm --dir web run verify` ·
  `pnpm --dir web exec vitest --project=unit run` · `pnpm --dir web exec vitest --project=storybook run` ·
  `git push` · `gh run watch`.
- **Capability:** ninguna (no toca fuente). Los commits de T2-T5 ya traen las suyas.
- **Verificación:** `gh run list --limit 1` → `success` en los 3 jobs (`go`, `ts`, `rust`).
- **Validación:** la corrida verde es el artefacto. Se cita su URL en `PARIDAD.md`.
- **DoD:** ☐ los 6 comandos locales verdes ☐ push hecho ☐ **3/3 jobs verdes en CI** ☐ URL en
  `PARIDAD.md`.
- **Riesgo:** **CI sale rojo por otra causa** (`rust`/`clippy` y `-race` no se corrieron local en el
  relevamiento — declarado NO VERIFICADO). **Es un stop-the-line:** se arregla antes de seguir, no se
  sigue con el árbol rojo. **Reversión:** ninguna; es un gate.

---

## Tramo 1 · El modelo y el disco (problemas A, B, D, E) — T7…T14

**El borrado y la migración van ANTES de que exista una UI que liste** (spec §Estado, orden
obligatorio): si no, se lista lo que se va a migrar.

---

### T7 · `domain.Conversacion` + las 4 operaciones puras + la invariante

- **Qué:** la entidad nueva y las operaciones donde vive la invariante «exactamente una activa».
- **Por qué:** RF-300, RF-301, RF-303, RF-304 · CV-D3/CV-D7/CV-D9/CV-D13 · `arquitectura.md` §1.2/§1.4.
- **Archivos:** `internal/domain/conversacion.go` (**nuevo**) ·
  `internal/domain/conversacion_test.go` (**nuevo**).
- **Diseño:** el struct completo de `arquitectura.md` §1.2 (14 campos; `Conv []Turn` **sin
  `omitempty`**; `Turnos` **no existe**, se deriva con `NumTurnos()`), más:
  ```go
  var ErrConvNoEncontrada, ErrTituloVacio, ErrSinActiva error
  func (s *Session) Activa() (*Conversacion, bool)
  func (s *Session) CrearConversacion(id string, ahora time.Time) (nueva *Conversacion, desactivada string)
  func (s *Session) ActivarConversacion(cid string) (desactivada string, err error)
  func (s *Session) RenombrarConversacion(cid, titulo string) error
  func VerificarUnaActiva(s Session) error
  func (s *Session) NormalizarConversaciones(ahora time.Time) []string
  func (c Conversacion) NumTurnos() int
  ```
  ⚠ **`domain.Turn` NO se toca** — es el tipo que roza el wire del turno user y agregarle un campo
  es el incidente HS-26 (`conductor_test.go:TestUserTurnWireSinCamposExtra:54`).
  ⚠ Ningún comentario del archivo pone `transcript`/`jsonl`/`.claude/projects` **adyacente** a un
  decoder (`arquitectura.md` §7.6).
- **Capability:** **nueva CAP-140** `dominio-l0/conversacion-como-entidad.yaml`, con
  `business_rules` (BR-CV-1..BR-CV-4) y `scenarios` BDD.
- **Verificación:** `go test ./internal/domain/ -run TestInvariante -v`
  - `TestInvarianteUnaActivaTrasCadaTransicion` (table-driven sobre las 4) → **E-01, E-15, E-20**
  - `TestCrearDesactivaLaAnterior` → **E-06**
  - `TestActivarConversacionAjenaEs404` → **E-11** (parte de dominio de RF-343 CA-3)
  - `TestRenombrarVacioNoCambiaElTitulo` → **E-30**
  - `TestNormalizarReparaYLoDice` (0 conv · 0 activas · 2 activas) → **E-01**
  - `TestTituloEditadoNoSeReDeriva` → **E-33**
- **Validación:** las 4 operaciones son puras: se prueban **sin un solo fake**. Si un test necesita
  un `fakeAgent`, la invariante quedó en el lugar equivocado.
- **DoD:** ☐ 6 tests verdes ☐ `VerificarUnaActiva` corre después de cada transición en el test ☐
  CAP-140 con `status: vivo` y su `valida:` ☐ `TestDomainIndependentOfTransport:206` verde ☐
  `TestNoJSONLSchemaParsing:328` verde.
- **Riesgo:** falso positivo de `TestNoJSONLSchemaParsing` por la palabra «transcript» en la doc de
  `Conv`. **Mitigación:** redactar el comentario, **jamás** sumar el prefijo a `jsonlExento`.
  **Reversión:** borrar 2 archivos; nada más los importa todavía.

---

### T8 · `Session` se parte, y el árbol vuelve a compilar

- **Qué:** quitar los 9 campos que bajan, agregar `Conversaciones []Conversacion`, y arreglar todos
  los call-sites.
- **Por qué:** RF-300 · CV-D3.
- **Archivos:** `internal/domain/session.go` (`:96-127`) · `internal/usecase/session_service.go`
  (`consume` `:452+`, `spawnLocked` `:390`/`:406-408`/`:431`, `Turn` `:347-355`) ·
  `internal/usecase/session_rotacion.go` (`:57-63`) · `internal/usecase/session_historial.go`
  (`:46-58`, `:112`) · `internal/adapters/transport/http/sessions.go`.
- **Diseño:** mecánico y guiado por el compilador. `r.meta.X` → `r.meta.Activa().X` para los 9. Dos
  puntos que **no** son mecánicos:
  - `spawnLocked:406-408` — el `Checkpoint` que se concatena a la tarjeta sale de
    `r.meta.Activa().Checkpoint`. Una línea (`arquitectura.md` §4.6).
  - `session_historial.go:112` — `HistorialCerrada` recorre `c.CadenaCC` de la **sesión**; pasa a
    recorrer las de sus conversaciones.
  **`Session` NO se renombra a `Sesion`** (`arquitectura.md` §1.3): rompería el `#Símbolo` de 9
  capabilities sin comprar nada.
- **Capability:** modifica **CAP-14** (`dominio-l0/sesion-frente-de-trabajo.yaml`, hoy con
  `valida`/`scenarios`/`business_rules` **vacíos**) — punteros nuevos + scenarios poblados;
  y **CAP-53/CAP-59/CAP-97** (punteros `#Símbolo`).
- **Verificación:** `go build ./...` · `go test ./... -race`. Los 5 tests de
  `sesion-viva-consistente` (`TestOneTurnAtATime:1911`, `TestFramesCarryRunID:1924`,
  `TestResumeAutoSana:1957`, `TestNoSilentEventDrop:1699`, `TestSessionSpawnsInArnesPath:1891`)
  **tienen que seguir verdes sin cambiarles el cuerpo** — es el canario de que el reparto no cambió
  el pipe.
- **Validación:** `go test ./docs/architecture/fitness/ -run 'TestCapabilityPointer'` verde: ningún
  puntero quedó colgando.
- **DoD:** ☐ compila ☐ `go test ./... -race` verde ☐ los 5 del boundary verdes **sin tocar su
  cuerpo** ☐ R1-símbolo verde ☐ `go-arch-lint` verde.
- **Riesgo:** 🔴 **el más alto del paquete.** Un `r.meta.X` mal traducido en `consume` corrompe el
  estado del turno en silencio. **Mitigación:** el compilador caza el 100 % de los renombres porque
  los campos **se eliminan** (no se dejan deprecados) — un campo eliminado es un error de compilación,
  un campo dejado como alias es un bug latente. **Reversión:** `git revert` del commit; T7 no depende
  de esto.

---

### T9 · Envelope versionado + cadena de migradores (problema B)

- **Qué:** el mecanismo de `arquitectura.md` §2.2/§2.3, y los **primeros tests que
  `internal/adapters/store` haya tenido**.
- **Por qué:** RF-335, RF-339 · boundary `archivo-durable-declara-su-esquema` · E-43, E-47.
- **Archivos:** `internal/adapters/store/esquema.go` (**nuevo**) ·
  `internal/adapters/store/migracion.go` (**nuevo**) · `internal/adapters/store/registry.go`
  (`Load` `:45-61` y `Save` `:65-97` envuelven el sobre) ·
  `internal/adapters/store/migracion_test.go` (**nuevo**) ·
  `internal/adapters/store/testdata/sessions-v1-real.json` (**nuevo** — copia del archivo REAL de
  17 202 B).
- **Diseño:** `const EsquemaActual = 2` · `type sobre struct{…}` · `type migrador func(json.RawMessage) (json.RawMessage, error)` ·
  `var migradores = map[int]migrador{1: deV1aV2}` · `func detectarVersion(b []byte) (int, error)` ·
  `func AbrirRegistro(rutaV2, rutaLegado, sello string, clave ClaveCalificada) (*Registry, Informe, error)`.
  `deV1aV2` es **puro**: no hace IO, no conoce el Portafolio (el re-key es T11, paso aparte).
  `ports.SessionStore` **no cambia de firma** ⇒ `arnes_registry.go` y los callers no se enteran.
- **Capability:** **nueva CAP-143** `indice-persistencia/migracion-de-esquema-en-disco.yaml`; y
  modifica **CAP-25** (`persistencia`, hoy `vivo·nc` con cero evidencia) → `vivo` con `valida`.
- **Verificación:** `go test ./internal/adapters/store/ -v`
  - `TestSaveEstampaEsquemaActual` → check `envelope-versionado`
  - `TestDetectarVersionArrayDesnudoEsV1` · `TestDetectarVersionSobre`
  - `TestMigracionRespaldaAntesDeEscribir` → check `migracion-forward-only-con-respaldo`
  - `TestMigracionEsIdempotente` (correrla 2× da el mismo byte) → **E-43**
  - `TestDeV1aV2ConservaTodoElFixtureReal` (5 sesiones → 5×1 conv; `s6165ac75` con sus 90 turnos;
    `ultima_interaccion` **vacío**; `conv: []` y no `null` en las de 0 turnos) → **E-43, H-B**
  - `TestCrashAMitadDejaElViejoEntero` (falla el `Save`, se relee el original) → **E-47**
- **Validación:** correr el daemon nuevo con `--sessions` apuntando a una **copia** del
  `~/.arnesia/sessions.json` real y comparar `GET /api/sessions` antes/después: mismas 5 sesiones,
  mismos turnos, mismos `ctx_pct`.
- **DoD:** ☐ 7 tests verdes ☐ el fixture es el archivo **real**, no uno sintético ☐ `sessions.json`
  del fixture **no se modifica** por la corrida ☐ CAP-143 + CAP-25 ☐ `go-arch-lint` verde (`store`
  sigue con `mayDependOn: [domain, ports]`).
- **Riesgo:** un campo que `deV1aV2` olvide llevar se pierde en silencio. **Mitigación:** el test
  compara **campo por campo** contra el fixture real, no un `len()`. **Reversión:** el daemon viejo
  lee `sessions.json`, que nunca se toca.

---

### T10 · Cuarentena, esquema futuro, y el error de `Save` que llega al operador

- **Qué:** los modos C, D y E de `arquitectura.md` §2.4.
- **Por qué:** RF-336, RF-338 · E-44, E-45 · checks `corrupto-se-preserva` y
  `esquema-futuro-no-se-degrada`.
- **Archivos:** `internal/adapters/store/registry.go` · `internal/usecase/session_service.go`
  (`persistLocked` `:902-912`) · `cmd/arnesia/main.go` (`:325-334`).
- **Diseño:**
  ```go
  func (r *Registry) SoloLectura() (bool, string)
  func (s *SessionService) persistLocked() error     // ← ANTES era func() sin retorno (:902)
  ```
  **Quién propaga y quién no, deliberado:** las mutaciones que el operador acaba de pedir
  (`CrearConversacion`, `ActivarConversacion`, `RenombrarConversacion`, `Create`, `Rename`,
  `SetView`, `Close`) **propagan y revierten**. `Turn` y `consume` **siguen logueando y siguiendo**:
  un turno en vuelo no se puede deshacer, y abortarlo por un fallo de disco dejaría al conductor
  hablando solo. Es la misma asimetría que el repo ya tiene entre operación y stream.
- **Capability:** modifica **CAP-143** (los modos de fallo) y **CAP-25**.
- **Verificación:** `go test ./internal/adapters/store/ ./internal/usecase/ -run 'Corrupto|Futuro|SinPermisos'`
  - `TestArchivoCorruptoSePreservaYSeDice` (JSON inválido ⇒ `.corrupto-<sello>` + registro vacío +
    el original **intacto** en la cuarentena) → **E-44**
  - `TestCorrupcionNoUsaSeed` (`seedSessions` no aparece) → **E-44**
  - `TestEsquemaFuturoNoSePisa` (`schema_version: 99` ⇒ `SoloLectura()==true` y **el archivo no
    cambia un byte** tras el arranque) → modo E
  - `TestPersistFallidoRevierteLaMutacion` (fs de sólo lectura ⇒ 500 y la conversación **no** aparece
    en `List()`) → **E-45**
- **Validación:** `chmod 500 ~/.arnesia` en una copia, crear una conversación por `curl`, ver el 500
  con el motivo del filesystem y confirmar que `GET .../conversaciones` no la lista.
- **DoD:** ☐ 4 tests verdes ☐ el `.corrupto-` lleva sello y un segundo arranque no lo pisa ☐ el log
  de arranque **nombra la ruta** de la cuarentena ☐ modo E no escribe **nada**.
- **Riesgo:** que `persistLocked` devolviendo error se ignore con `_ =` en algún caller.
  **Mitigación:** `errcheck` está en `.golangci.yml`. **Reversión:** revertir la firma.

---

### T11 · CV-D16 · el re-key de las llaves vivas (problema E)

- **Qué:** el paso 4 de la migración + el comando de inspección y reversión.
- **Por qué:** **CV-D16 🧑‍⚖️ FIRMADA** (`decisiones.md:225-246`) · `arquitectura.md` §2.6.
- **Archivos:** `internal/adapters/store/rekey.go` (**nuevo**) ·
  `internal/adapters/store/rekey_test.go` (**nuevo**) · `cmd/arnesia/sesiones.go` (**nuevo**) ·
  `cmd/arnesia/main.go` (cablea `ClaveCalificada` desde el Portafolio).
- **Diseño:**
  ```go
  type ClaveCalificada func(idPelado, cwd string) (clave string, ok bool, motivo string)
  type Recalibracion struct{ SesionID, Antes, Despues, Motivo string }
  func reKey(sesiones []domain.Session, clave ClaveCalificada) []Recalibracion
  ```
  Algoritmo de dos vías (`arquitectura.md` §2.6): por `cwd` cuando hay, **por id como fallback**
  cuando el `cwd` está vacío — **3 de las 4 a recalibrar tienen `cwd` vacío**, verificado en vivo.
  Cero o ≥2 candidatas ⇒ **se deja como está** con motivo `sin-candidata`/`ambigua: N`. `store` **no
  importa el Portafolio** (`.go-arch-lint.yml:171-172`): la función se inyecta desde `cmd`, que sí
  puede (`:212`).
  Comando, 3 modos: `recalibrar-llaves` (dry-run, default) · `--revertir --desde <bak>` (R2) ·
  `--aplicar` (idempotente).
- **Capability:** modifica **CAP-143**; y **CAP-25**.
- **Verificación:** `go test ./internal/adapters/store/ -run TestReKey -v`
  - `TestReKeyPorCwd` · `TestReKeyPorIdCuandoNoHayCwd` (el caso real de 3 de 4)
  - `TestReKeyAmbiguaNoDecide` y `TestReKeySinCandidataNoDecide` (la llave **no cambia** y el motivo
    viaja) → `no-aplica-no-es-cero`
  - `TestReKeyEsIdempotente` (2ª corrida: todas `ya-calificada`)
  - `TestRevertirRestauraSoloElArnes` (las conversaciones creadas después **sobreviven**) → **R2**
- **Validación (obligatoria, del operador):** correr el **`--dry-run` contra el
  `~/.arnesia/sessions.json` real ANTES de arrancar el binario nuevo** y **transcribir la tabla de 5
  filas a `PARIDAD.md`**. Eso convierte «nada se borra» en algo observable.
- **DoD:** ☐ 6 tests verdes ☐ `sessions.json.bak-<sello>` creado **antes** de cualquier escritura ☐
  el log de arranque imprime las 5 `Recalibracion` con motivo ☐ **R1 y R2 probadas a mano** contra
  una copia ☐ la tabla del dry-run en `PARIDAD.md`.
- **Riesgo:** que una llave se resuelva a la entrada equivocada del Portafolio. **Mitigación:** «una
  sola candidata o no se decide» + el dry-run previo, que lo hace visible antes de aplicar.
  **Reversión:** **dos caminos especificados** (`arquitectura.md` §2.6): R1 = `rm sesiones.json`
  (`sessions.json` nunca se tocó); R2 = `--revertir --desde <bak>` sin perder las conversaciones
  nuevas.

---

### T12 · La ley se invierte: `Conv` y `Checkpoint` sobreviven al archivado (problema D)

- **Qué:** invertir `session_historial.go:56-58`, renombrar y **trazar** el cambio de CAP-98.
- **Por qué:** RF-305, RF-306 · CV-D8 + **enmienda F-3** · E-50.
- **Archivos:** `internal/usecase/session_historial.go` (`:37-68`, `:71-92`, `:97-123`) ·
  `internal/usecase/session_historial_test.go` (`:38`, **`:73`**) · `cmd/arnesia/main.go`
  (`cerradasPathDefault` `:794-805` → `sesiones-archivadas.json`) ·
  `docs/product/capabilities/usecases/historial-de-conversaciones.yaml` ·
  `docs/product/ledger/HS-29.md` (**nuevo**) · `docs/product/LEDGER.md`.
- **Diseño:** se eliminan las 3 líneas `cerrada.Turnos = len(cerrada.Conv)` / `cerrada.Conv = nil` /
  `cerrada.Checkpoint = ""` (`:56-58`). `archivarLocked` archiva la **sesión entera con sus
  conversaciones completas**. `CerradaEn` pasa a significar «esta SESIÓN se archivó».
  **Procedimiento de 4 pasos, en el MISMO commit** (`arquitectura.md` §7.3):
  1. `TestCloseArchivaMetadata` → **`TestCloseArchivaConTranscript`**, aserción **invertida**
     (`if c.Conv == nil { t.Error(...) }`). **No se borra:** el rename hace visible en el diff que
     una ley cambió. + `TestCloseConservaCheckpoint` (F-3, que ninguna decisión previa cubría).
  2. El **cuerpo** de CAP-98 se reescribe (la oración exacta está en `arquitectura.md` §7.3), no se
     le suma un pointer.
  3. `change_log` de CAP-98 gana `{story_id, date, type: derive, summary}` y `business_rules[]` la
     regla con su `enforcement`.
  4. `ledger/HS-29.md` registra el cambio de regla de negocio.
- **Capability:** modifica **CAP-98** (los 4 pasos) y **CAP-59** (`Close` archiva con `Conv`).
- **Verificación:** `go test ./internal/usecase/ -run 'TestClose|TestHistorial' -v`
  - `TestCloseArchivaConTranscript` → **E-50**
  - `TestCloseConservaCheckpoint` → **E-50**, F-3
  - `TestHistorialCerradaCoseCadena` (existente, `:82`) sigue verde: el lector JSONL se conserva como
    fallback para archivados **pre-migración**
- **Validación:** cerrar una sesión con 2 conversaciones por `curl` y leer
  `~/.arnesia/sesiones-archivadas.json`: trae `conv` con los N turnos y `checkpoint` no vacío.
- **DoD:** ☐ los 3 tests verdes ☐ **el test viejo aparece como renombrado en el diff, no como
  borrado** ☐ CAP-98 con cuerpo nuevo + `change_log` + `business_rules` ☐ `ledger/HS-29.md` escrito
  ☐ `LEDGER.md` indexado ☐ `TestCapabilityStatusConsistent` verde.
- **Riesgo:** que alguien lea el diff como «borraron un test que molestaba». **Mitigación:** el
  rename + los 4 pasos + la ficha de ledger. **Reversión:** revertir el commit entero, incluida la
  capability.

---

### T13 · Cableado: `AbrirRegistro` en `main.go` y el `Informe` en el log

- **Qué:** que el daemon use el mecanismo nuevo y **diga** lo que hizo al arrancar.
- **Por qué:** RF-306 CA-3 · E-01, E-43, E-44.
- **Archivos:** `cmd/arnesia/main.go` (`:312-334`, `:794-805`).
- **Diseño:** `store.NewRegistry(...)` → `store.AbrirRegistro(sesionesV2, sessionsLegado, sello, clave)`.
  Se loguea, una línea cada uno: `Migro`+`DesdeVersion`+`RespaldoEn` · cada `Reparaciones[i]` con
  `slog.Warn` · cada `Recalibradas[i]` con su motivo · `CuarentenaEn` con `slog.Error`. Si el
  registro archivado no se puede crear, **degrada honesto con `warn` y jamás bloquea el arranque** —
  comportamiento vigente (`main.go:325-327`), que RF-306 CA-3 manda conservar.
- **Capability:** modifica **CAP-143** y **CAP-25**.
- **Verificación:** `go test ./cmd/arnesia/ -run TestArranque` ·
  `TestArranqueMigraYLoguea` (fixture v1 ⇒ el `Informe` trae `Migro:true` y las 5 `Recalibracion`).
- **Validación:** arrancar contra una copia del `~/.arnesia` real y **leer el log**: tiene que decir
  qué migró, dónde está el respaldo y qué llaves movió. Un arranque silencioso es un fallo del ticket.
- **DoD:** ☐ el log dice las 4 cosas ☐ un segundo arranque **no** vuelve a migrar (idempotencia) ☐
  `sessions.json` intacto ☐ `sesiones.json` con `schema_version: 2`.
- **Riesgo:** ruido en el log en cada arranque. **Mitigación:** el `Informe` sólo emite lo que ocurrió;
  un arranque sin novedades no loguea nada. **Reversión:** `rm ~/.arnesia/sesiones.json`.

---

### T14 · Medir `persistLocked` con el registro inflado (H-F) — no opcional

- **Qué:** medir el costo real y decidir con el número si hace falta el corte a archivo-por-sesión.
- **Por qué:** H-7 del spec · `arquitectura.md` §2.5. `persistLocked` reescribe el registro **entero**
  por cada `init`/`message`/`act`/`result` (`:465`, `:515`, `:556`, `:573`) y ahora el registro
  incluye N `Conv`.
- **Archivos:** `internal/usecase/session_persist_bench_test.go` (**nuevo**) ·
  `docs/product/BACKLOG.md` (el resultado).
- **Diseño:** `func BenchmarkPersistLocked20Conversaciones(b *testing.B)` con el registro proyectado
  a 20 conversaciones × el fixture real de 12 281 B ≈ 240 KB.
  **Presupuesto: < 15 ms p95.** **Disparador:** si lo supera, el corte es
  `~/.arnesia/sessions/<id>/conversaciones.json` — carpeta que **ya existe** (CAP-96,
  `session_service.go:410`) y que reduce la reescritura a la sesión que cambió. El envelope se aplica
  igual a cada archivo, así que **no es un rediseño: es cambiar el `path` del `Registry`**.
- **Capability:** modifica **CAP-25**.
- **Verificación:** `go test ./internal/usecase/ -bench BenchmarkPersistLocked -benchtime 20x -run '^$'`.
- **Validación:** el número se escribe en `PARIDAD.md`. **Si no se midió, no se puede afirmar que
  alcanza** — eso sería el pass fabricado que el repo prohíbe.
- **DoD:** ☐ benchmark corrido ☐ el número (no «parece rápido») en `PARIDAD.md` ☐ si supera el
  presupuesto, ticket de corte abierto en `BACKLOG.md` **antes** de seguir al tramo 2.
- **Riesgo:** que se saltee «porque anda bien en 5 sesiones». **Mitigación:** está en el DoD del
  tramo. **Reversión:** n/a (es medición).

---

## Tramo 2 · Usecase y API — T15…T20

---

### T15 · La transición atómica: crear · retomar · renombrar (problema A)

- **Qué:** `transicionLocked` y las 3 operaciones del servicio.
- **Por qué:** RF-307, RF-308, RF-310, RF-311, RF-312, RF-316 · CV-D7/CV-D11 ·
  `sesion-viva-consistente` v1.2 check `transicion-de-conversacion-atomica`.
- **Archivos:** `internal/usecase/session_conversaciones.go` (**nuevo**) ·
  `internal/usecase/session_service.go` (`sessionRuntime` `:85-109` gana `convActiva string`) ·
  `internal/usecase/session_conversaciones_test.go` (**nuevo**) ·
  `docs/architecture/fitness/arch_test.go` (el enforcer nuevo).
- **Diseño:** las firmas y los 8 pasos de `arquitectura.md` §4.2. **Ningún campo de `sessionRuntime`
  se convierte en mapa**: el runtime sigue siendo uno por sesión porque CV-D7 lo hace correcto.
  `runSeq` **NO se toca** (monótono por sesión, §4.4). El patrón de cierre calca `Close`
  (`:300`/`:318`/`:320-322`): tomar el lock, mutar, guardarse el `live`, soltar, `live.Close()`.
  Los `grants` se **descartan** (deny-by-default).
- **Capability:** **nueva CAP-141** `usecases/conversaciones-de-una-sesion.yaml`; modifica
  **CAP-53**, **CAP-54** (los permisos pendientes se deniegan en la transición) y **CAP-59**.
- **Verificación:**
  - `go test ./internal/usecase/ -run TestConversaciones -v`:
    `TestCrearConTurnoEnVueloEsErrBusy` → **E-07** · `TestRetomarConTurnoEnVueloEsErrBusy` → **E-11**
    · `TestTransicionDeniegaPermisosPendientesConMotivo` → **E-08**, RF-311 ·
    `TestFalloDePersistenciaDejaElEstadoAnterior` → **E-36, E-37** ·
    `TestCrearDosVecesRapidoDejaUnaActiva` → **E-09** ·
    `TestRetomarLaActivaEsNoOp` → **E-15** · `TestFramesDelConductorViejoSeDescartan` → **CR-3**
  - `go test ./docs/architecture/fitness/ -run TestTransicionDeConversacionEsAtomica` → el check 5
    del boundary
- **Validación:** con el daemon vivo y una sesión en `streaming`, `curl -X POST .../conversaciones`
  → **409 con motivo**. Con la sesión quieta → 201 y `GET .../conversaciones` muestra 2, una activa.
- **DoD:** ☐ 7 tests + el enforcer verdes ☐ `sesion-viva-consistente` sube a **5/5 checks con
  enforcer real** y su changelog v1.2 se completa ☐ los 4 checks originales **siguen verdes sin tocar
  su cuerpo** ☐ INDEX actualizado.
- **Riesgo:** un `publish` que se cuele **bajo** el lock rompe la disciplina de las 953 líneas.
  **Mitigación:** `transicionLocked` **devuelve** los frames, no los publica. **Reversión:** revertir
  el archivo nuevo; nada anterior depende de él.

---

### T16 · La rotación emite (problema C · C-6 · H-8)

- **Qué:** `rotarLocked` devuelve el frame; `Turn` lo publica tras soltar el lock.
- **Por qué:** RF-313 CA-4 · CV-D10 · `design.md` C-6 · el §4C del mockup **no es realizable hoy**.
- **Archivos:** `internal/usecase/session_rotacion.go` (`:52-64` — **firma**, no lógica) ·
  `internal/usecase/session_service.go` (`Turn` `:347-349` y `:376`) ·
  `internal/usecase/session_rotacion_test.go` (`:16`, `:67`).
- **Diseño:** `func (s *SessionService) rotarLocked(r *sessionRuntime) dockFrame`. El cuerpo
  (`:53-63`) **no cambia una línea de lógica**. El frame:
  `{Kind:"conversacion", ConversacionEvento:"rotada", ConversacionID:…, TurnoIdx: len(Conv)-1, Text: breadcrumbRotacion}`.
  **Orden obligatorio** tras el `Unlock`: `publish(rotacion)` **y después** `publish(status)` — el
  breadcrumb entra a `Conv` antes del turno del usuario (`:355`), así que si el frame llegara después
  la marca aparecería debajo del mensaje que la disparó. Sin `run_id` (no pertenece a un turno).
  **C-5: gana el código** — el texto es `breadcrumbRotacion` (`session_rotacion.go:12`), no el del
  dibujo: reescribir la constante dejaría los transcripts ya persistidos con la marca vieja.
- **Capability:** modifica **CAP-97** (`rotacion-de-contexto`) y **nueva CAP-145**
  `http-sse/frame-de-conversacion.yaml`.
- **Verificación:** `go test ./internal/usecase/ -run 'TestRotacion|TestCtxHist' -v`
  - `TestRotacionInvisible` (`:67`, existente) sigue verde, ajustado al valor de retorno
  - `TestRotacionEmiteFrameConTurnoIdx` → **E-19**
  - `TestRotacionFrameLlegaAntesDelStatus` → orden
  - `TestRotacionNoCreaConversacionNueva` → **E-19, E-20**
- **Validación:** E2E con `-rotacion-umbral 5`, dock abierto: la marca `— contexto rotado, seguimos —`
  aparece **sin recargar**, centrada, y la lista sigue mostrando N.
- **DoD:** ☐ 4 tests verdes ☐ **cero `s.publish` dentro de `rotarLocked`** (grep) ☐ el orden
  verificado ☐ H-8 cerrado en `spec.md`.
- **Riesgo:** el frame llega desordenado bajo carga. **Mitigación:** los dos `publish` son
  consecutivos en el mismo goroutine tras el `Unlock`. **Reversión:** volver la firma a `void`.

---

### T17 · La búsqueda: scan en memoria + fragmento + normalización

- **Qué:** `Conversaciones(id, q)` con el matcher.
- **Por qué:** RF-321, RF-322, RF-324, RF-341 · CV-D8 · BR-CV-8.
- **Archivos:** `internal/usecase/session_conversaciones.go` ·
  `internal/usecase/session_busqueda_test.go` (**nuevo**) ·
  `internal/usecase/testdata/conv-90-turnos.json` (**nuevo** — el transcript **real** de
  `s6165ac75`, 12 281 B).
- **Diseño:**
  ```go
  func (s *SessionService) Conversaciones(id, q string) ([]ConversacionResumen, int, error)
  func normalizar(s string) string          // NFD + strip de diacríticos + lower — UN solo lugar
  func fragmentoDe(conv []domain.Turn, q string) string   // el PRIMER match, elipsis a ambos lados
  ```
  El scan corre sobre `sessionRuntime.meta` **en memoria** (el daemon ya los tiene): sin FTS5, sin
  índice, sin tocar `~/.arnesia/index.db`. ⚠ La premisa de CV-D8 «sin tocar `index.db`» describía el
  índice **viejo**; hoy `index.db` es SQLite real con WAL (`c2c61ba`) — **la decisión no cambia y se
  re-funda por volumen**: 12 281 B × 100 ≈ 1,2 MB. `total` es el **total de la sesión**, no el de
  coincidencias (es el denominador de «N de M»).
- **Capability:** **nueva CAP-142** `usecases/buscar-en-el-transcript.yaml`.
- **Verificación:** `go test ./internal/usecase/ -run TestBusqueda -v`
  - `TestBuscaEnTituloYEnTexto` → **E-03**
  - `TestBusquedaInsensibleAAcentosYMayusculas` («VACIO» encuentra «vacío») → **E-23**
  - `TestQueryEnBlancoEsSinQuery` → **E-26**
  - `TestFragmentoEsElPrimerMatch` → **E-29**
  - `TestCoincideSoloEnTituloNoTraeFragmento` → **E-27**
  - `TestBusquedaSobreTranscriptRealDe90Turnos` (fixture real, **no sintético**) → **E-25**
  - `TestTotalEsElTotalNoElDeCoincidencias` → **E-22**, RF-318
- **Validación:** `curl '.../conversaciones?q=manifiesto'` contra el daemon con datos reales: sólo
  las que coinciden, con `fragmento`, y `total` = el total de la sesión.
- **DoD:** ☐ 7 verdes ☐ el fixture es el transcript **real** ☐ `normalizar` existe **una sola vez**
  (grep) ☐ cero referencias a `index.db`.
- **Riesgo:** `normalizar` duplicada en el FE. **Mitigación:** el FE **no filtra**: manda `q` y pinta
  lo que llega. **Reversión:** borrar el método; la lista sin `q` sigue andando.

---

### T18 · Handlers HTTP: 5 rutas nuevas, 1 retirada, 1 parámetro retirado

- **Qué:** el transporte.
- **Por qué:** RF-340…RF-345 · `arquitectura.md` §5.1/§5.3.
- **Archivos:** `internal/adapters/transport/http/sessions_conversaciones.go` (**nuevo**) ·
  `internal/adapters/transport/http/router.go` (`:122-131`) ·
  `internal/adapters/transport/http/sessions.go` (`listSessions` `:18-43`; `historialCerrada`
  `:47-56` se retira del router) ·
  `internal/adapters/transport/http/sessions_conversaciones_test.go` (**nuevo**).
- **Diseño:** las 5 rutas registradas **antes** de `GET /api/sessions/{id}` (`router.go:126`) para
  que el mux no ambigüe. Mapeo de errores: `ErrBusy`→**409** (mismo criterio que `Turn`,
  `session_service.go:339-342`) · `ErrConvNoEncontrada`→**404** · `ErrTituloVacio`→**400** ·
  error de `Save`→**500** con el motivo del fs · `SoloLectura`→**503**.
  `?cerradas=1` → **400 con puntero** al endpoint nuevo (mejor que ignorarlo o devolver 200 vacío).
  `GET /api/sessions` vuelve a ser **siempre un array**, lo que elimina la única respuesta bimorfa
  conocida (check `forma-de-respuesta-unica`).
  `historialCerrada` **se retira del router**; `HistorialCerrada` y `history.Reader` **se conservan
  como capacidad del dominio, sin ruta**: siguen siendo el fallback para archivados pre-migración.
- **Capability:** modifica **CAP-52** (`superficie-rest`, el número de endpoints) y **CAP-98**
  (los 2 endpoints por-arnés se reemplazan).
- **Verificación:** `go test ./internal/adapters/transport/http/ -run TestConversaciones -v`
  — **cierra el gap más grande del lado Go** (`sessions_test.go` cubre hoy **1 de 10** endpoints):
  `TestListarDevuelve404SiLaSesionNoExiste` → **E-35**, RF-340 CA-3 ·
  `TestCrearDevuelve409ConTurnoEnVuelo` → **E-07, E-39** ·
  `TestActivarConvAjenaDevuelve404` → **RF-343 CA-3** ·
  `TestRenombrarVacioDevuelve400YNoCambia` → **E-30** ·
  `TestCerradas1DevuelveGon400ConPuntero` → **RF-345** ·
  `TestListSessionsSiempreEsArray` → `forma-de-respuesta-unica`
- **Validación:** los 5 `curl` contra el daemon instalado, con sus códigos.
- **DoD:** ☐ 6 tests verdes ☐ las 5 rutas responden ☐ la ruta B2 devuelve 404 ☐
  **`TestRutaServidaEstaDeclarada` FALLA** (el contrato todavía no las tiene) — es lo esperado, lo
  cierra T19 ☐ `_sin-declarar.yaml` pierde la entrada de `/cerradas/{id}/historial`.
- **Riesgo:** orden de registro en el mux. **Mitigación:** el test de 404 lo prueba.
  **Reversión:** revertir `router.go`.

---

### T19 · El contrato: 5 paths, 4 schemas, `Session` corregido (RF-346)

- **Qué:** el diff exacto de `arquitectura.md` §5.4.
- **Por qué:** RF-346 · H-3 · el gate de T5 **está rojo desde T18** y este ticket lo cierra.
- **Archivos:** `docs/architecture/contracts/api/openapi.yaml` (`:605-629`, `:856-878`) ·
  `docs/architecture/contracts/api/_sin-declarar.yaml`.
- **Diseño:** (a) los 3 paths nuevos con sus 5 operaciones y **todos** sus códigos (200/201/400/404/
  409/500/503) · (b) `parameters:` de `/sessions` con `arnes` y `cerradas` (`deprecated: true`) ·
  (c) `Session` **pierde** `claude_session_id`, `model`, `ctx_pct`, `conv`, `turnos`, `cadena_cc` y
  **gana** `cwd`, `reparacion`, `cerrada_en`, `activa` (`$ref` + **required**) · (d) 4 schemas
  nuevos: `Turn` (con el `enum` corregido a **`[user, assistant, sys, act]`** — hoy dice
  `[user, assistant, sys]` en `:877`, **stale desde `RolAct`**, `session.go:56`), `Conversacion`,
  `ConversacionActiva`, `ConversacionesListado` · (e) `components/parameters/SessionId`.
- **Capability:** modifica **CAP-52**.
- **Verificación:** `go test ./docs/architecture/fitness/ -run 'TestRuta|TestQueryParam|TestExencion'`
  → **4/4 verdes de nuevo**. Es el ticket que devuelve el gate al verde.
- **Validación:** cada respuesta de los 5 `curl` de T18 valida contra su schema.
- **DoD:** ☐ los 4 enforcers verdes ☐ el `enum` de `rol` con los 4 valores ☐ `_sin-declarar.yaml`
  en **10** entradas `/api` (de 11) ☐ ninguna ruta nueva sin declarar.
- **Riesgo:** YAML mal formado rompe el parser del enforcer. **Mitigación:** el propio test lo caza.
  **Reversión:** revertir el yaml (y el gate vuelve a rojo, que es la señal correcta).

---

### T20 · Toda sesión nace con su conversación (RF-301) + el helper de tests

- **Qué:** `Create` y `seedSessions` producen una conversación activa; ajustar `newTestService`.
- **Por qué:** RF-301 CA-1/CA-3 · E-01.
- **Archivos:** `internal/usecase/session_service.go` (`Create` `:250-266`, `seedSessions`
  `:947-951`) · `docs/architecture/fitness/arch_test.go` (`newTestService`).
- **Diseño:** `Create` llama `sess.CrearConversacion(newConvID(), time.Now().UTC())` antes de
  persistir. El título es `"nueva conversación"` — **no** hereda el `"nuevo frente"` de `:261`, que
  es de la **sesión** (RF-301 CA-3). ⚠ **`newTestService` deja exactamente una sesión sembrada en
  `List()[0]` y 5 tests del boundary dependen de eso** (`arch_test.go:1961`, `:1914`, `:1928`,
  `:1892`, más `TestNoSilentEventDrop`): **se ajusta el helper, no los tests** — siguen midiendo lo
  mismo, y que sigan verdes es el canario de que el reparto no cambió el pipe.
  ⚠ RF-301 CA-1 pide `201` en `POST /api/sessions`: **verificar el código actual antes de cambiarlo**;
  si hoy responde `200`, el cambio es breaking y va **declarado** en el commit y en el OpenAPI, no de
  contrabando.
- **Capability:** modifica **CAP-59** (`Create` nace con conversación) y **CAP-14**.
- **Verificación:** `go test ./internal/usecase/ ./docs/architecture/fitness/ -race`
  - `TestCreateNaceConUnaConversacionActiva` → **E-01**, RF-301
  - `TestTituloInicialNoHeredaNuevoFrente` → RF-301 CA-3
  - los 5 del boundary **verdes sin tocar su cuerpo**
- **Validación:** `curl -X POST /api/sessions` → la respuesta trae `activa` con `titulo: "nueva
  conversación"`, `turnos: 0`, `conv: []`.
- **DoD:** ☐ 2 tests nuevos verdes ☐ **los 5 del boundary verdes con el cuerpo intacto** ☐ el código
  de estado de `POST /api/sessions` verificado y declarado.
- **Riesgo:** tocar `newTestService` invalida silenciosamente los 5 enforcers. **Mitigación:** el DoD
  exige que pasen **sin editar sus aserciones**. **Reversión:** revertir el helper.

---

## Tramo 3 · El cromo del dock (RF-326…RF-332) — T21…T25

---

### T21 · Los tipos del wire y los 5 métodos del cliente

- **Qué:** partir `Session` en TS y cablear el transporte.
- **Por qué:** RF-300, RF-340…RF-345 · `arquitectura.md` §6.1.
- **Archivos:** `web/src/shared/api/types.ts` (`:16-38`, `:68-90`) ·
  `web/src/shared/api/client.ts` (`:214-225`).
- **Diseño:** los 4 tipos de `arquitectura.md` §6.1 con `| undefined` explícitos
  (`@tsconfig/strictest` ⇒ `exactOptionalPropertyTypes`, `package.json:51`). `Session.activa:
  ConversacionActiva` **no opcional** — el tipo hace imposible el `if (!activa)` defensivo que
  escondería el bug. `conv: Turn[]` **sin `?`**. `DockFrame` gana `conversacion_id?`,
  `conversacion_evento?`, `turno_idx?`. `client.ts` gana `conversaciones(id, q?)` (con `signal`,
  porque teclear rápido pisa peticiones — mismo criterio que `escanearProyecto`, `:119`),
  `crearConversacion`, `activarConversacion`, `renombrarConversacion` (**sin** `signal`: abortar una
  mutación a mitad es peor que esperar, `client.ts:205-207`), y **pierde** `conversacionesDeArnes`
  (`:218-221`) y `historialCerrada` (`:222-225`).
  ⚠ `cadena_cc` (`types.ts:37`) hoy **no lo consume nadie** (verificado): si T23 no lo pinta,
  **se borra**. Un campo del wire sin consumidor es deuda.
- **Capability:** modifica **CAP-68** y **CAP-52**.
- **Verificación:** `pnpm --dir web exec tsc --noEmit` — el compilador enumera cada call-site roto.
- **Validación:** el tipo del wire coincide campo por campo con la respuesta real de
  `curl /api/sessions`.
- **DoD:** ☐ `tsc` verde ☐ los 2 métodos viejos borrados ☐ `pnpm --dir web run verify` verde.
- **Riesgo:** cascada de errores de tipo. **Mitigación:** es el objetivo — el compilador es el
  inventario. **Reversión:** revertir 2 archivos.

---

### T22 · `sessions-store`: `activa`, `convRev`, la rama del frame, el `default:`

- **Qué:** el store compartido absorbe el modelo nuevo.
- **Por qué:** `arquitectura.md` §6.2 · E-40, E-41.
- **Archivos:** `web/src/shared/store/sessions-store.ts` (`:26-69`, `:80-87`, `:177-219`,
  `:262-459`, `:464-478`) · `web/src/shared/store/sessions-store.test.ts` (**nuevo** — el store de
  478 líneas **no tiene test hoy**).
- **Diseño:** los **6 mapas conservan su llave `session.id`** (§7.1: bajo CV-D7 «por sesión» ya es
  «por conversación activa»). `appendConv` (`:80-87`) escribe en `s.activa.conv`. **Gana**
  `convRev: Record<string, number>`. **Rama nueva** `case "conversacion"` con los 4 eventos
  (`arquitectura.md` §6.2): `creada`/`activada` reemplazan `activa` y **limpian** `streaming[id]`,
  `pendingPerms[id]`, `msgFlushed[id]`, `wroteInRun[id]`, pero **NO** `finalizedRun[id]` (§4.4) ni
  `scope[id]` (RF-352 CA-3); `rotada` appendea **sólo si `conv.length === turno_idx`**.
  ⚠ `onDock` **no tiene `default:`** (verificado, `:267-458`): se agrega con `console.warn`. Es una
  línea y cierra un agujero preexistente que este paquete ensancha.
  Selectores nuevos, **todos primitivos** (§6.3): `selectConvActivaId`, `selectCtxPct`,
  `selectCtxCaliente`, `selectTituloActiva`. **Prohibido** un selector que devuelva objeto sin
  `useShallow` — el repo no lo usa en ningún lado y este paquete no lo introduce.
- **Capability:** modifica **CAP-68**, **CAP-71**; **CAP-145** (el frame).
- **Verificación:** `pnpm --dir web exec vitest --project=unit run src/shared/store/sessions-store.test.ts`
  - `frame conversacion creada reemplaza la activa y limpia los buffers` → **E-06**
  - `frame rotada con turno_idx repetido no duplica el breadcrumb` → **E-41**
  - `frame rotada con turno_idx correcto appendea` → **E-19**
  - `finalizedRun NO se limpia en la transición` → §4.4
  - `scope NO se limpia en la transición` → **E-48**
  - `un kind desconocido no rompe y avisa` → el `default:`
- **Validación:** dos pestañas sobre `:4200`; crear una conversación en una y ver que la lista de la
  otra se actualiza (**E-40**).
- **DoD:** ☐ 6 tests verdes ☐ `default:` presente ☐ cero selectores que devuelvan objeto ☐
  `depcruise` verde (**`shared` no importa `widgets`**).
- **Riesgo:** limpiar `finalizedRun` por «simetría» rompe la idempotencia. **Mitigación:** hay un
  test que lo prohíbe explícitamente. **Reversión:** revertir el store.

---

### T23 · `CtxChip` + `IdentidadDetalle` (RF-327, RF-328, RF-330)

- **Qué:** el chip-disclosure con los 4 datos + el `cwd`.
- **Por qué:** CV-D14 · RF-327/328/330 · `design.md` §3.3/§4.2.
- **Archivos:** `web/src/widgets/chat-dock/ui/ctx-chip.tsx` (**nuevo**) ·
  `web/src/widgets/chat-dock/ui/ctx-chip.stories.tsx` (**nuevo**, 10 stories).
- **Diseño:** `CtxChipProps` de `design.md` §3.3. **`caliente` viene por prop**, derivado de
  `rotacion_pendiente` del wire: el umbral vive en el daemon (`SetUmbralRotacion`,
  `session_service.go:141-145`, default 40 por `main.go:120`) y el FE **no lo conoce ni lo teclea**.
  **C-3 realizado:** el **número** en `--foreground`; la **barra** en `--warn` (no textual, 3,76 ≥ 3 ✓).
  A 0 % se pinta igual (BR-CV-9). `aria-expanded` + `aria-controls`; el `%` va **en el nombre
  accesible del botón**, no sólo en la barra (que es `aria-hidden`).
- **Capability:** **nueva CAP-144** `fe-chat/panel-de-conversaciones.yaml` (arranca acá).
- **Verificación:** `pnpm --dir web exec vitest --project=storybook run src/widgets/chat-dock/ui/ctx-chip.stories.tsx`
  → C-01…C-10 de `plan-storybook.md` §2.3. **C-02 `Caliente`** asserta por **estilo computado** que
  el número **no** usa `--warn` → **C-3, RF-353**. **C-03 `Cero`** → **E-04**. **C-07
  `DetalleSinSesionCC`** → **E-05**.
- **Validación:** el detalle muestra `cc-id · arnés · modelo · cwd` — los 4 de la `SessionLine`
  vigente (`chat-dock.tsx:81-90`) **más** el `cwd`, que hoy no se ve en ninguna superficie.
- **DoD:** ☐ 10 stories verdes ☐ ninguna con `a11y: "todo"` ☐ el assert de color de C-02 pasa en
  **light y dark** ☐ nada de la `SessionLine` se perdió (superset, BR-CV-11).
- **Riesgo:** el assert de estilo computado es frágil. **Mitigación:** es el **único** de estilo del
  plan y está justificado (C-3 es la contradicción más cara). **Reversión:** borrar 2 archivos.

---

### T24 · `ConversacionRow` (RF-326, RF-331)

- **Qué:** la fila 2 del cromo: `▶ título · chip ctx · 🔍 · ＋`, con renombrado en línea.
- **Por qué:** CV-D14/CV-D9 · RF-326/331 · `design.md` §3.2/§4.1.
- **Archivos:** `web/src/widgets/chat-dock/ui/conversacion-row.tsx` (**nuevo**) ·
  `conversacion-row.stories.tsx` (**nuevo**, 11).
- **Diseño:** `ConversacionRowProps` de `design.md` §3.2. Renombrado **calca `session-rail.tsx:256-273`**:
  `✎ → input → Enter confirma · Escape descarta · blur confirma`; vacío ⇒ **descarta**
  (`:199-200`). Editando, el input **toma la fila entera** y el ctx y las acciones no se renderizan.
  **C-1: el spec gana sobre el mockup** — con `status ∈ {streaming, await}` el `＋` va `disabled`
  **con `title`** (patrón sancionado, `new-session-picker.tsx:362-363`); nunca un control apagado y
  mudo. **C-10:** dos controles, dos destinos de foco — `▶` ⇒ filas, `🔍` ⇒ buscador.
- **Capability:** modifica **CAP-144**.
- **Verificación:** `vitest --project=storybook run src/widgets/chat-dock/ui/conversacion-row.stories.tsx`
  → B-01…B-11. **B-06** → **E-06** · **B-07** → **E-07** · **B-09** → **E-30** · **B-10** →
  **E-32** · **B-11** → **E-34** · B-04/B-05 → C-10/RF-355.
- **Validación:** el gesto de renombrado es **indistinguible** del rail, lado a lado.
- **DoD:** ☐ 11 verdes ☐ `＋` deshabilitado **siempre** con `title` ☐ ninguna con `a11y: "todo"`.
- **Riesgo:** divergencia del gesto respecto del rail. **Mitigación:** B-08…B-11 lo cubren.
  **Reversión:** borrar 2 archivos.

---

### T25 · `ScopeRow` condicional · glifo `»` · `ChatDock` recompuesto (RF-329, RF-332, RF-326)

- **Qué:** bajar el cromo de 4 filas a 2.
- **Por qué:** CV-D14/CV-D15 · RF-326/329/332.
- **Archivos:** `web/src/widgets/chat-dock/ui/chat-dock.tsx` (`:21-42`, `:32`, `:46-76`, `:78-93`,
  `:95-154`) · `chat-dock.stories.tsx` (T2 gana A-04…A-07, A-17).
- **Diseño:** `ScopeRow` (`:46-76`) devuelve **`null` sin scope** (la fila **no existe en el DOM**);
  con scope conserva el chip **literal** (`:55-70`) incluido el `✕` que llama a `setScope(null)`
  (`:64`). **Se quitan, declarado:** el rótulo «Alcance:» (`:51`), el chip punteado del arnés
  (`:52-54`) y el hint del estado vacío (`:72`). `SessionLine` (`:78-93`) **se retira como fila fija**
  y su cuerpo se muda a `IdentidadDetalle` (T23) — **mudanza, no borrado** (BR-CV-11).
  `⟩ colapsar` (`:32`) → `» colapsar`; la palabra y el `title` (`:29`) **literales**.
  **NO se tocan** (criterio de aceptación): `Composer` (`:268-375`), `fitComposer` (`:263-266`),
  `Bubble` (`:233-258`), `ActivityCard` (`:183-231`), `agrupar` (`:160-172`), `rotulo` (`:176-179`).
  `Messages` **gana** el copy de vacío nuevo (reemplaza `:110-115`, única eliminación de copy firmado
  del paquete, autorizada por CV-D3) y pinta la marca de rotación como `RolSys` centrado, que
  `Bubble` (`:234-240`) **ya hace** — **cero código nuevo para la marca**.
- **Capability:** modifica **CAP-68**, **CAP-69** (`acotar-alcance`).
- **Verificación:** A-04 `Reposo` asserta **exactamente 2 filas**
  (`queryByText("Alcance:") === null` ∧ `queryByText(/◍/) === null` ∧ el botón de ctx presente ∧
  `» colapsar`) → **RF-326, RF-332, E-03**; A-06 `ConNodoEnAlcance` asserta **3** → **RF-329**;
  A-09 `ConversacionRecienCreada` → **E-06**; A-05/A-17 en dark → **RF-353**.
  Y **A-01…A-03 de T2 tienen que seguir verdes tras ajustarlas al superset**: es la prueba de que el
  composer, las burbujas y la tarjeta quedaron intactos.
- **Validación:** medir en la app: ~46 px más de transcript en reposo (`design.md` §6).
- **DoD:** ☐ A-01…A-09, A-17 verdes ☐ **exactamente 2 filas** en reposo ☐ **3** con nodo (C-2) ☐ el
  `✕` sigue llamando a `setScope(null)` ☐ los 6 símbolos de «NO se tocan» sin cambios (diff vacío).
- **Riesgo:** 🔴 romper superficie firmada de HS-26 sin darse cuenta. **Mitigación:** el baseline de
  T2. **Reversión:** revertir `chat-dock.tsx`.

---

## Tramo 4 · La lista y el buscador (RF-317…RF-325) — T26…T29

---

### T26 · `useConversaciones` — el store del widget (RF-334)

- **Qué:** el transporte del panel, en el dock.
- **Por qué:** RF-334 · CV-D2 · `design.md` §3.6 · `arquitectura.md` §6.2.
- **Archivos:** `web/src/widgets/chat-dock/model/conversaciones-store.ts` (**nuevo**) ·
  `conversaciones-store.test.ts` (**nuevo**).
- **Diseño:** `ConversacionesState` de `design.md` §3.6. Tres disciplinas **heredadas, no
  inventadas**: refetch en **cada** apertura (`portafolio-picker-store.ts:22-25`) · `cerrar()`
  descarta búsqueda y estado (`session-rail.tsx:43-47`) · el error del backend se guarda **tal cual**
  (`portafolio-picker-store.ts:45-47`). El seam con `useSessions` es una **suscripción descendente**
  a `convRev` (`arquitectura.md` §6.2): `shared` **no puede** importar `widgets`
  (`shared-no-upward`, `.dependency-cruiser.js:41-48`).
- **Capability:** modifica **CAP-144**.
- **Verificación:** `pnpm --dir web exec vitest --project=unit run src/widgets/chat-dock/model/conversaciones-store.test.ts`
  → U-01…U-10 de `plan-storybook.md` §2.7:
  U-01 refetch → RF-334 · U-02/U-03 → **E-42** · U-04 → **E-35** · U-05 → **E-37** ·
  U-06 → **E-36** · U-07 → **E-39** · U-08 → **E-38** · U-09 → **E-41** · U-10 → **E-30**.
- **Validación:** con el daemon caído, abrir la lista: motivo + `Reintentar`, **jamás «0
  conversaciones»** (BR-CV-10).
- **DoD:** ☐ 10 tests verdes ☐ `depcruise` verde ☐ el error nunca se reescribe.
- **Riesgo:** una suscripción que no se limpia. **Mitigación:** `cerrar()` la desuscribe; U-02 lo
  cubre. **Reversión:** borrar 2 archivos.

---

### T27 · `ConversacionFila` (RF-319, RF-322, RF-323)

- **Qué:** la fila: título · última interacción · nº turnos · ctx, con `<mark>`.
- **Por qué:** CV-D13 · `design.md` §3.5/§4.4 (**8 estados**).
- **Archivos:** `web/src/widgets/chat-dock/ui/conversacion-fila.tsx` (**nuevo**) ·
  `conversacion-fila.stories.tsx` (**nuevo**, 9).
- **Diseño:** `ConversacionFilaProps` de `design.md` §3.5. **Cuatro datos y ninguno más** — la fila
  **no es una tarjeta** (CV-D13). Fecha relativa cerca / absoluta lejos; **el FE formatea lo que
  recibe, no lo calcula** (RF-304 CA-3). `turnos: 0` ⇒ `sin turnos todavía · ctx 0 %`;
  `ultima_interaccion` vacío **con** turnos ⇒ `sin fecha · 41 turnos` (**H-4/H-B: migrada — se dice,
  no se inventa**). El `<mark>` va como `<span>` propio, no dependiendo del reset del navegador
  (`design.md` §9). La fila activa se distingue por **radio relleno + rótulo `activa` + fondo**, no
  sólo por el tinte (RF-356).
- **Capability:** modifica **CAP-144**.
- **Verificación:** `vitest --project=storybook run src/widgets/chat-dock/ui/conversacion-fila.stories.tsx`
  → E-01s…E-09s. **E-03s** → **E-04** · **E-04s** → **H-4** · **E-05s/E-06s** → **RF-322, E-23** ·
  **E-07s** → **E-27** · **E-08s** → **E-11** · E-09s dark → **RF-353**.
- **Validación:** los 4 formatos de fecha del dibujo (`hace 4 min` · `ayer 18:02` · `24 jul` ·
  `23 jul`).
- **DoD:** ☐ 9 verdes ☐ los 8 estados de `design.md` §4.4 cubiertos ☐ ninguna fecha inventada ☐
  ninguna con `a11y: "todo"`.
- **Riesgo:** formateo de fecha dependiente del locale del runner. **Mitigación:** las stories fijan
  la fecha de referencia. **Reversión:** borrar 2 archivos.

---

### T28 · `ConversacionesPanel` (RF-317, RF-318, RF-320, RF-321, RF-325)

- **Qué:** el organismo: buscador + rótulo + lista + pie, **en sitio**.
- **Por qué:** CV-D2/CV-D4 · `design.md` §3.4/§4.3 (**8 estados**) · §8.2.
- **Archivos:** `web/src/widgets/chat-dock/ui/conversaciones-panel.tsx` (**nuevo**) ·
  `conversaciones-panel.stories.tsx` (**nuevo**, 14).
- **Diseño:** `ConversacionesPanelProps` de `design.md` §3.4. **Abre EN SITIO** y toma el área del
  transcript: **cero `<dialog>`, cero backdrop, cero portal** — el gesto que `NewSessionPicker` ya
  usa (`session-rail.tsx:101-110`). **C-8: la estructura la gana el repo** —
  `<ul role="listbox">` → `<li role="option" aria-selected>`, no el `<button><span>` del dibujo; el
  aspecto queda idéntico. **C-9: el pie sólo lleva `Cancelar`** — seleccionar **ES** la acción; nadie
  agrega un «Retomar» por simetría. Orden: `ultima_interaccion` **desc**, las de 0 turnos **al final**
  por `creada_en` desc (RF-320 CA-2); la activa **no se fija arriba**. Con **1** conversación el
  buscador **no se dibuja** (RF-325 CA-2). Búsqueda **incremental sin Enter** (precedente
  `new-session-picker.tsx:69-74` + `:192`). **El buscador sigue activo con turno en vuelo** (E-28).
  Esqueleto: **calca el contrato ARIA de `PickerSkeleton`** (`new-session-picker.tsx:309-326`) en
  Tailwind, **sin** las clases `pf-*` (C-4: arrastrarían el scope CSS del Portafolio al dock).
- **Capability:** modifica **CAP-144**.
- **Verificación:** `vitest --project=storybook run src/widgets/chat-dock/ui/conversaciones-panel.stories.tsx`
  → D-01…D-14. **D-01/D-02** → **E-35** · **D-05** → **E-02, E-21** · **D-06/D-07** → **E-24** ·
  **D-08/D-09** → **E-22** · **D-10** → **E-11, E-28** · **D-11** → **E-10** · **D-12** → **E-15** ·
  **D-13** → **RF-354** · **D-14** → **RF-317**.
- **Validación:** con el dock estirado a 300 px y al 60 % de la ventana, las filas **truncan** y la
  lista no introduce mínimo propio (RF-317 CA-4).
- **DoD:** ☐ 14 verdes ☐ **cero** `<dialog>`/backdrop/portal (grep) ☐ header, fila 2 y composer
  **siguen visibles** con la lista abierta ☐ `↑`/`↓`/`Enter`/`Escape` andan ☐ ninguna con
  `a11y: "todo"`.
- **Riesgo:** `role="listbox"` mal implementado rompe axe. **Mitigación:** el gate está en `error`.
  **Reversión:** borrar 2 archivos.

---

### T29 · Cablear el panel al dock: retomar, crear, franja, vacíos

- **Qué:** el comportamiento vivo (RF-307…RF-316).
- **Por qué:** CV-D7/CV-D10/CV-D11 · RF-307/310/313/314/315/316.
- **Archivos:** `web/src/widgets/chat-dock/ui/chat-dock.tsx` · `chat-dock.stories.tsx`
  (A-08, A-10…A-16).
- **Diseño:** `ChatDock` es el **único** que lee stores; los 4 componentes son props-puras.
  Con `abierta`, `<ConversacionesPanel>` **reemplaza** `<Messages>` (§6.3: por eso buscar no
  re-renderiza el transcript). La franja de retoma es **efímera**: desaparece al primer frame del
  conductor nuevo o al fallar; **no se dibuja sin `claude_session_id`** (E-16). **C-11: `.resumebar.bad`
  se usa** para RF-348, con texto en `--foreground` sobre `--crit-soft` (15,89:1 ✓), **nunca**
  `--warn` (3,19 ✗). El detalle se abre solo **al retomar y al rotar** — los dos momentos en que
  cambia el `cc-id` (RF-314). Con el dock colapsado la lista y el detalle **se desmontan** y al
  reabrir arrancan cerrados (RF-315 CA-2); `chatOpen` **no lo toca ningún camino de este paquete**.
- **Capability:** modifica **CAP-144**, **CAP-68**, **CAP-100** (el historial rehidrata idéntico,
  ahora también para las inactivas).
- **Verificación:** A-08 → **E-03** · A-10 → `design.md` §4.6#2 · **A-11** → **E-10** · **A-12** →
  **E-16** · **A-13** → **E-13** · **A-14** → **E-19** · **A-15** → **E-07, E-28** · **A-16** →
  **E-08**.
- **Validación:** en la app instalada: crear, retomar, buscar, renombrar, y **ver la marca de
  rotación llegar sin recargar**.
- **DoD:** ☐ las 17 stories del dock verdes ☐ el transcript **no se re-renderiza** al teclear en el
  buscador (React DevTools Profiler) ☐ `⌘K` sin hermanos nuevos (RF-357).
- **Riesgo:** re-render del transcript en cada tecla. **Mitigación:** el panel reemplaza `Messages`
  + store separado + selectores primitivos. **Reversión:** revertir `chat-dock.tsx`.

---

## Tramo 5 · La mudanza y el cierre — T30…T33

---

### T30 · El picker pierde `ConversacionesDelArnes` (RF-333, RF-334)

- **Qué:** completar la mudanza declarada. **Último a propósito:** hasta acá el operador conserva la
  superficie vieja.
- **Por qué:** CV-D2 · RF-333/334 · BR-CV-11 (`mockups/INDEX.md` regla dura 3).
- **Archivos:** `web/src/widgets/session-rail/ui/new-session-picker.tsx` (`:15`, `:111`, `:235`,
  `:263-307`) · `new-session-picker.stories.tsx` · **elimina**
  `web/src/widgets/session-rail/model/conversaciones-store.ts` y `conversaciones-store.test.ts`.
- **Diseño:** se quitan el import, la llamada `cargar(e.clave)` (`:111`), el render (`:235`) y el
  componente (`:263-307`). El picker queda **sólo con lo suyo**: elegir el arnés y la copia.
  `no-sibling-widget-imports` (`.dependency-cruiser.js:56-63`, **error**) hace que la mudanza sea
  **física**, no un re-export. De paso desaparece la llamada que violaba la doctrina «props PURAS:
  CERO transporte» del propio archivo (`:18-24` vs `:111`).
- **Capability:** modifica **CAP-72** (pierde el bloque; se corrige su prosa `:33-37` sobre la clave)
  y **CAP-98** (pierde los punteros al store viejo y su `valida: conversaciones-store.test.ts`).
- **Verificación:** `vitest --project=storybook run src/widgets/session-rail/ui/new-session-picker.stories.tsx`
  → 7/7 con el assert nuevo `queryByText(/Conversaciones:/) === null` (F-01).
  `go test ./docs/architecture/fitness/ -run TestCapabilityPointer` verde (los punteros al archivo
  borrado tienen que haber desaparecido de CAP-98).
- **Validación:** abrir «＋ Nueva sesión» en la app: el bloque no está, y crear una sesión sigue
  andando.
- **DoD:** ☐ 7 stories verdes ☐ los 2 archivos **borrados** ☐ R1 verde ☐ **`PARIDAD.md` anota que
  las 4 stories rojas históricas ya estaban verdes desde T3 por contraste — no por esta
  desaparición**, y que la deuda del token `--warn` sigue abierta.
- **Riesgo:** presentar el efecto colateral como el arreglo. **Mitigación:** está en el DoD.
  **Reversión:** `git revert`.

---

### T31 · E2E contra la app INSTALADA

- **Qué:** el circuito completo del operador. Detalle en [`plan-pruebas.md`](./plan-pruebas.md) §3.
- **Por qué:** requisito del operador: cuando él instale, tiene que ver exactamente lo construido.
  El paquete de dictado descubrió 4 bugs que los fakes no veían.
- **Archivos:** `docs/product/stories/2026-07-26-conversaciones-del-panel/e2e/` (**nuevo**):
  `entorno.sh`, `conversaciones.mjs`, `mock-claude.sh`.
- **Diseño:** `make installer` → **`make dev-sync`** → `~/.local/bin/arnesia` → daemon en `:4200`.
  ⚠ **`make dev-sync` no es opcional:** el shell Tauri instalado **prefiere SIEMPRE**
  `~/.local/bin/arnesia` sobre el sidecar (`Makefile:27-32`), así que verificar sin él es **medir el
  daemon viejo con el formato viejo**. Playwright **crudo** (`web/node_modules/playwright`, v1.61.1);
  **`@playwright/test` no está instalado** y **no se usa**. **No se depende de ningún MCP de Chrome**
  (pueden estar tomados por otra sesión). Precedente: `stories/2026-07-08-chat-cc-funcional/e2e/`,
  cuyo import por **ruta absoluta hardcodeada** se **corrige** acá con `fileURLToPath` + `resolve`.
  ⚠ El SPA tiene `:4200` **hardcodeado** (`client.ts:21`): el daemon de prueba corre en `:4200` con
  `HOME` apuntando a un sandbox (`plan-pruebas.md` §4).
- **Capability:** ninguna nueva; el `dev_preview.e2e_test` de **CAP-144** apunta acá.
- **Verificación:** `bash e2e/entorno.sh && node e2e/conversaciones.mjs` → 5 escenarios:
  **E-14** (`cwd` movido) · **E-18** (rotación con dock colapsado, `-rotacion-umbral 5`) ·
  **E-33** (renombrar mientras stremea) · **E-40** (dos vistas) · **E-49** (desvincular del Portafolio).
- **Validación:** capturas por escenario + consola limpia.
- **DoD:** ☐ los 5 pasan contra el binario instalado ☐ `make dev-sync` corrido ☐ `GET /api/version`
  coincide con el sello recién construido ☐ `~/.arnesia` real **no tocado** ☐ capturas en el paquete.
- **Riesgo:** ensuciar el `~/.arnesia` del operador. **Mitigación:** `HOME` sandbox
  (`plan-pruebas.md` §4). **Reversión:** borrar el sandbox.

---

### T32 · 🧑‍⚖️ CV-D6 · el borrado de las 3 cerradas (E-46)

- **Qué:** el procedimiento de 6 pasos de RF-337. **Manual, del operador.**
- **Por qué:** CV-D6 · BR-CV-12. **CV-D6 y CV-D16 son decisiones distintas sobre datos distintos y
  no se unifican:** acá se **borra** `sesiones-cerradas.json`; las vivas se **re-key** (T11).
- **Archivos:** `PARIDAD.md` (el rastro). **Ningún archivo de código.**
- **Diseño:** los 6 pasos, en orden: detener el daemon y **confirmar** con `curl` que falla →
  `cp` a `~/.arnesia/sesiones-cerradas.json.bak-<AAAAMMDD-HHMM>` → **verificar la copia** (`3`
  entradas) → `rm` del archivo **entero** → arrancar y confirmar → **una línea en `PARIDAD.md`** con
  fecha, los 3 ids (`s1b38a066`, `s408bb085`, `s020210e3`) y la ruta del `.bak`.
- **Capability:** ninguna (no toca fuente).
- **Verificación:** **ninguna automatizada, a propósito** (`plan-storybook.md` §4.3): un test tendría
  que borrar el archivo real o fingir uno, y las dos opciones son peores que un rastro escrito. Es el
  **único escenario del paquete sin test**, y está declarado.
- **Validación:** `GET /api/sessions?cerradas=1` → **400** con el puntero (RF-345), no un 200 vacío.
- **DoD:** ☐ los 6 pasos ejecutados **en orden** ☐ `.bak` verificado con 3 entradas ☐ la línea en
  `PARIDAD.md` ☐ el daemon arranca sin el archivo.
- **Riesgo:** correrlo con el daemon vivo ⇒ `persistLocked` (`:902`) o `archivarLocked`
  (`session_historial.go:65`) reescriben entre la copia y el borrado. **Mitigación:** el paso 1 exige
  confirmar que el daemon murió. **Reversión:** `cp` del `.bak`.

---

### T33 · Cierre: capabilities, cifras, changelog, PARIDAD

- **Qué:** dejar el SSoT y las cifras en sync.
- **Por qué:** METODOLOGIA §10 · `codigo-traza-a-capability` R1-R4 · `mockups/INDEX.md` regla 5 ·
  `versionado.md` §changelog-y-bump.
- **Archivos:** `docs/product/capabilities/**` (6 nuevas + 12 modificadas) ·
  `docs/product/capabilities/INDEX.md` · `docs/product/checkpoint.md` · `docs/product/BACKLOG.md` ·
  `CHANGELOG.md` · `mockups/INDEX.md` · `PARIDAD.md` · `docs/product/ledger/HS-29.md`.
- **Diseño:** ⚠ **primero** `python3 scripts/cap_doctor.py --index` — el INDEX está **stale en 22
  hojas** (le falta el módulo `telemetria` entero: `GROUP_ORDER` de `scripts/cap_doctor.py:90-94` no
  lo incluye). Después: las 6 nuevas (CAP-140…CAP-145, +CAP-146 si T5 la creó) y las 12 modificadas.
  `bash scripts/estado.sh` regenera el bloque de cifras (**nunca se teclean**). El changelog se anota
  con `python3 scripts/changelog.py add Agregado "las conversaciones viven en el panel: lista,
  buscador, ＋ nueva y retomar"` — y `Cambiado` para el cromo de 2 filas, `Corregido` para el drift
  del OpenAPI. `BACKLOG.md` recibe los huecos H-C, H-D, H-E, H-G y el resultado de T14.
- **Capability:** todas.
- **Verificación:** `go test ./docs/architecture/fitness/ -run TestCapability` → los **5** verdes ·
  `bash scripts/estado.sh --check` → exit 0 · `go test ./docs/architecture/fitness/ -run TestChangelog`.
- **Validación:** `PARIDAD.md` completa: mockup ↔ componente ↔ story ↔ RF, fila por fila, con las
  desviaciones declaradas (las 3 de `spec.md` §Estado + las 13 contradicciones de `design.md`).
- **DoD:** ☐ R1/R2/R3/R4 verdes ☐ `estado.sh --check` exit 0 ☐ CHANGELOG con entradas ☐
  `mockups/INDEX.md` con la fila del paquete ☐ `checkpoint.md` **nombra este paquete** (hoy no
  figura, deuda declarada en el relevamiento §9.8) ☐ PARIDAD completa ☐ `ledger/HS-29.md`.
- **Riesgo:** cerrar con capabilities a medias. **Mitigación:** R1/R2 son `error` y rompen CI.
  **Reversión:** n/a.

---

## Cobertura E-01…E-50 → ticket → mecanismo

**50 de 50 cubiertos.** Varios llevan más de un mecanismo (dominio + transporte + UI).

| E | Ticket(s) | Mecanismo de verificación |
|---|---|---|
| E-01 sin conversaciones | **T7**, T20, T13 | Go: `TestNormalizarReparaYLoDice`, `TestCreateNaceConUnaConversacionActiva` |
| E-02 una sola | **T28** | story `D-05` |
| E-03 muchas | **T25**, T28, T17 | story `A-04`, `D-04` + Go `TestBuscaEnTituloYEnTexto` |
| E-04 de 0 turnos | **T23**, T27 | story `C-03`, `E-03s` |
| E-05 sin `claude_session_id` | **T23** | story `C-07` |
| E-06 crear quieta | **T7**, T24, T22 | Go `TestCrearDesactivaLaAnterior` + story `B-06`, `A-09` + unit |
| E-07 crear con `streaming` | **T15**, T18, T24 | Go `TestCrearConTurnoEnVueloEsErrBusy` + http `Test…409` + story `A-15`/`B-07` |
| E-08 crear con `await` | **T15**, T29 | Go `TestTransicionDeniegaPermisosPendientesConMotivo` + story `A-16` |
| E-09 crear dos veces rápido | **T15** | Go `TestCrearDosVecesRapidoDejaUnaActiva` |
| E-10 retomar quieta | **T29**, T28 | story `A-11`, `D-11` |
| E-11 retomar con turno en vuelo | **T15**, T18, T27, T28 | Go + http `TestActivarConvAjenaDevuelve404` + story `D-10`, `E-08s` |
| E-12 proceso inexistente | **T15** | Go: el spawn perezoso; sin código nuevo, con test que lo fija |
| E-13 `--resume` que falla | **T29** | story `A-13` + Go `TestResumeAutoSana:1957` (**ya existe**) |
| E-14 `cwd` inexistente | **T31** | **E2E vivo** (`mv` de la carpeta) |
| E-15 retomar la activa | **T7**, T15, T28 | Go `TestRetomarLaActivaEsNoOp` + story `D-12` |
| E-16 retomar de 0 turnos | **T29** | story `A-12` |
| E-17 rotación en medio de un turno | **T16** | Go: `session_rotacion_test.go` (**ya existe**) — imposible por construcción |
| E-18 rotación con dock colapsado | **T31** | **E2E vivo** (`-rotacion-umbral 5`, `⌘K` cerrado) |
| E-19 rotación con la lista abierta | **T16**, T22, T29 | Go `TestRotacionNoCreaConversacionNueva` + unit + story `A-14` |
| E-20 rotación en una inactiva | **T7**, T16 | Go: imposible por invariante |
| E-21 buscar con una sola | **T28** | story `D-05` |
| E-22 sin coincidencias | **T17**, T28 | Go `TestTotalEsElTotalNoElDeCoincidencias` + story `D-08`/`D-09` |
| E-23 acentos y mayúsculas | **T17**, T27 | Go `TestBusquedaInsensibleAAcentosYMayusculas` + story `E-06s` |
| E-24 la activa entre resultados | **T28** | story `D-07` |
| E-25 transcript enorme | **T17** | Go `TestBusquedaSobreTranscriptRealDe90Turnos` (**fixture real**) |
| E-26 query en blanco | **T17** | Go `TestQueryEnBlancoEsSinQuery` |
| E-27 sólo en el título | **T17**, T27 | Go `TestCoincideSoloEnTituloNoTraeFragmento` + story `E-07s` |
| E-28 buscar con turno en vuelo | **T28**, T24 | story `A-15`, `D-10` (el `searchbox` habilitado) |
| E-29 muchas coincidencias | **T17** | Go `TestFragmentoEsElPrimerMatch` |
| E-30 renombrar a vacío | **T7**, T18, T24, T26 | Go `TestRenombrarVacioNoCambiaElTitulo` + http `400` + story `B-09` + unit `U-10` |
| E-31 título duplicado | **T7** | Go: no hay unicidad que validar — el test lo **fija** |
| E-32 Escape | **T24** | story `B-10` |
| E-33 renombrar mientras llega un turno | **T7**, T31 | Go `TestTituloEditadoNoSeReDeriva` + **E2E vivo** |
| E-34 blur | **T24** | story `B-11` |
| E-35 daemon caído al listar | **T26**, T28, T18 | unit `U-04` + story `D-01`/`D-02` + http 404 |
| E-36 daemon caído al retomar | **T15**, T26 | Go `TestFalloDePersistenciaDejaElEstadoAnterior` + unit `U-06` |
| E-37 daemon caído al crear | **T15**, T26 | ídem + unit `U-05` |
| E-38 timeout | **T26** | unit `U-08` |
| E-39 respuesta 409 | **T18**, T26 | http + unit `U-07` |
| E-40 dos vistas | **T22**, T31 | unit + **E2E vivo** (app + navegador) |
| E-41 SSE reconecta | **T22**, T26 | unit: `turno_idx` repetido no duplica (`U-09`) |
| E-42 cambiar de sesión con la lista abierta | **T26** | unit `U-02`, `U-03` |
| E-43 instalación vieja | **T9** | Go `TestDeV1aV2ConservaTodoElFixtureReal`, `TestMigracionEsIdempotente` |
| E-44 archivos corruptos | **T10** | Go `TestArchivoCorruptoSePreservaYSeDice`, `TestCorrupcionNoUsaSeed` |
| E-45 sin permisos de escritura | **T10** | Go `TestPersistFallidoRevierteLaMutacion` |
| E-46 borrado CV-D6 | **T32** | ⚠ **manual, con rastro en `PARIDAD.md`** — el único sin test, declarado |
| E-47 crash a mitad de la migración | **T9** | Go `TestCrashAMitadDejaElViejoEntero` |
| E-48 nodo desaparece tras reindex | **T22** | unit: `scope[id]` **no** se limpia en la transición |
| E-49 arnés desaparece del Portafolio | **T31** | **E2E vivo** |
| E-50 cerrar la sesión entera | **T12** | Go `TestCloseArchivaConTranscript`, `TestCloseConservaCheckpoint` |

**Sin cobertura automatizada, declarado: 1 de 50** — **E-46** (`plan-storybook.md` §4.3). Es dato del
operador, el procedimiento exige el daemon detenido y copia previa, y un test tendría que borrar el
archivo real o fingir uno. Se cubre con rastro escrito, no con un pass.

**Además, dos cosas que este plan cubre y no son escenarios E-nn**, porque nacieron del relevamiento:
la **rotación que no emitía frame** (H-8/C-6 → T16) y el **drift del OpenAPI** (H-3 → T5/T19).

---

## Resumen por tramo

| Tramo | Tickets | Qué cierra | Gate para pasar |
|---|---|---|---|
| **0** llegar a verde | T1-T6 | Problema F · H-1 · el enforcer de contrato | 🧑‍⚖️ **CI 3/3 verde** (T6) |
| **1** modelo y disco | T7-T14 | A (modelo) · B · D · E/CV-D16 | `go test ./... -race` verde + el número de T14 |
| **2** usecase y API | T15-T20 | A (transición) · C · RF-307…312, 340…346 | los 5 checks del boundary + los 4 del contrato |
| **3** cromo | T21-T25 | RF-326…332 | 2 filas en reposo, 3 con nodo |
| **4** lista y buscador | T26-T29 | RF-317…325, 307…316 | las 56 stories verdes |
| **5** mudanza y cierre | T30-T33 | RF-333, 334, 337 · SSoT | 🧑‍⚖️ **PARIDAD** |
