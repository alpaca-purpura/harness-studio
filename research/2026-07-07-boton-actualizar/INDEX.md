# Botón «Actualizar» — self-update sin sudo (paquete de trabajo)

> Ficha HS-11 (deuda del instalador) · 2026-07-07
> Origen: el operador instala cada build con `sudo dpkg -i` (el .deb escribe en
> /usr/bin, dueño root — ESE es el sudo). Pide un botón en la propia app.
> Disciplina METODOLOGIA §10: mockup → decisiones → spec/design → 🧑‍⚖️ → código.

## Rumbo firmado (2026-07-07, respuestas del operador)

Self-update del **binario instalado** · mecanismo **`~/.local/bin`** (espacio de
usuario, cero sudo; el .deb queda para instalación inicial de terceros) · origen
**build local del repo** (el ciclo real de dogfood). Detalle en `decisiones.md`.

## Flujo y gates

1. **Mockup** (`mockup-actualizar.html`) — vista Ajustes con la sección «Versión y
   actualización»; tokens DTCG reales; datos REALES (huella del repo, ruta de
   instalación actual /usr/bin root); estados honestos. Iterar → 🧑‍⚖️ firma.
2. **Decisiones** (`decisiones.md`) — cada cambio conversado, al instante.
3. **Specs** (`spec.md` + `design.md`) → 🧑‍⚖️ firma del paquete.
4. **Implementación** — story=test + tests Go; gates verdes; el endpoint queda BAJO el
   confinamiento S1 (Host+Origin+token) y jamás acepta paths del request.
5. **Paridad** (`PARIDAD.md`) → gate final.

## Estado

- [x] rumbo firmado (alcance · mecanismo · origen)
- [x] mockup **v2 FIRMADO** (operador «firmo el mockup», 2026-07-07) — **artifact
      (publicar SIEMPRE a esta URL):** https://claude.ai/code/artifact/458c147a-b78b-461c-a54b-db2d741ac30b
- [x] spec.md (RF-100..107) + design.md — 🧑‍⚖️ **PAQUETE FIRMADO** (operador
      «Ok, firmo», 2026-07-07) → implementación AUTORIZADA
- [ ] implementación + stories + tests
- [ ] PARIDAD → gate final

## Diseño técnico SELLADO (FASE 1, 2026-07-07 — contraste RF↔existente hecho; código autorizado)

Contraste clave contra lo existente: `features/` está VACÍA (README) — `features/self-update`
estrena la capa · `bin/` ya está gitignored pero `bundle.sh` HOY no emite `bin/arnesia` ni
tiene modo sin-Tauri (decisión #6) · `withAuth` ya confina TODO `/api/*` (el endpoint hereda
S1 gratis) · `os.Executable()` post-rename da `(deleted)` en Linux (decisión #8).

**Dónde vive cada pieza:**

| Pieza | Ubicación | Nota |
|---|---|---|
| Puerto | `internal/ports/selfupdate.go` | `VersionInfo{Huella,Fecha,InstaladoEn,Repo,Escribible,Sucio}` + `SelfUpdater{Version() · Verificar(ctx) · Build(ctx) · VerificarBinario(ctx)→(huella,detalle,err) · Instalar(ctx) · Reiniciar()}` — cada paso `(detalle, err)`; el «ya al día» lo decide el USECASE comparando huellas (regla de negocio, no del adapter) |
| Adapter | `internal/adapters/selfupdate/updater.go` | `New(repo)` captura exePath (#8). Version = `debug.ReadBuildInfo()` + probe de escritura (tmp-file en el DIR del ejecutable — el rename exige write en el dir) · Verificar = repo existe + `go.mod` module esperado + `scripts/bundle.sh` + LookPath go/pnpm · Build = `bundle.sh --daemon-only` cwd=repo, `CommandContext` (cancelable), stderr-tail como detalle · VerificarBinario = `debug/buildinfo.ReadFile(repo/bin/arnesia)` SIN ejecutarlo · Instalar = CreateTemp EN el dir del exe → chmod 0755 → copy → `os.Rename` (mismo fs = atómico; tmp se limpia en fallo) · Reiniciar = `syscall.Exec(exePath, os.Args, os.Environ())` |
| Usecase | `internal/usecase/selfupdate_service.go` | `SelfUpdateService` con `sync.Mutex.TryLock` → sentinel `ErrActualizacionEnCurso` (409) · `ErrNoActualizable` (503: no-escribible/sin-repo, con motivo) · corta al primer fallo, pasos posteriores `no-corrido` · reporte `{resultado: actualizado\|ya-al-dia\|fallo, huella_nueva?, pasos[{paso,estado(ok\|fallo\|no-corrido\|agendado),detalle}]}` (nombres de paso = los 5 del mockup) · «actualizado» ⇒ paso reiniciar `agendado` y el lock NO se libera (#9) |
| Transport | `internal/adapters/transport/http/selfupdate.go` + rutas en `router.go` | `GET /api/version` · `POST /api/self-update` (cuerpo IGNORADO — cero params, RF-106); mapea sentinels 409/503 como `fuente.go`; tras responder `actualizado` agenda `go { sleep 600ms; Reiniciar() }` (post-respuesta, RF-105) |
| Config | `cmd/arnesia/main.go` runServe | flag `--repo` default `$ARNESIA_REPO`; wiring `selfupdate.New(*repo)` → `usecase.NewSelfUpdateService` → `NewHandler` |
| Build | `scripts/bundle.sh` | emite `bin/arnesia` en paso 2 (cp al sidecar) + `--daemon-only` salta Tauri Y el cálculo de TRIPLE (rustc no es prereq del self-update) — decisión #6 |
| OpenAPI | `arch/contracts/api/openapi.yaml` | `+ /version` · `+ /self-update` · bump `0.3.0-hs11` |
| go-arch-lint | `arch/fitness/.go-arch-lint.yml` | componente `selfupdate: in: internal/adapters/selfupdate/**` mayDependOn `[domain, ports]`; `cmd` lo suma |
| FE feature | `web/src/features/self-update/` (`index.ts` + `model/types.ts` + `ui/update-card.tsx`) | UI PURA (props `version/estado/reporte/error` + `onUpdate`), jamás fetchea — patrón Inspector; tipos del payload en `model` |
| FE página | `web/src/pages/shell/ui/global-view.tsx` | ruta `ajustes` deja `ComingSoon` → vista Ajustes real (vhead `alpacapurpura / Ajustes` + tarjeta + línea muted de futuras); transporte + máquina de estados + polling reinicio (arranca ~1.5s, interval 800ms, timeout 30s, solo `/api/version` — #9) |
| FE client | `web/src/shared/api/client.ts` | `getVersion<T>()` · `selfUpdate<T>()` (POST sin timeout FE) |
| CSS | `web/src/app/styles/ajustes.css` (scoped `.arnesia-ajustes`, tokens `var(--…)` verbatim del mockup) + import en `index.css` | patrón inspector.css |
| Stories | `features/self-update/ui/update-card.stories.tsx` | story=test por estado: idle · actualizando · error · no-escribible · sin-repo · éxito (+ click → onUpdate) |
| Tests Go | `internal/adapters/selfupdate/updater_test.go` (repo fake con bundle.sh stub · binario temporal compilado con `go build` en TestMain · instalar atómico · no-escribible · binario inexistente) · `internal/usecase/selfupdate_service_test.go` (corte al primer fallo · reporte · 409 concurrente · ya-al-día · no-actualizable) — `-race` | RF-104/106/107 |

**Mini-plan de commits (sellado):**
1. `feat(HS-11): bundle.sh emite bin/arnesia + modo --daemon-only (self-update sin rustc)`
2. `feat(HS-11): puerto SelfUpdater + adapter selfupdate — build/verificar/instalar atómico/re-exec, tests -race`
3. `feat(HS-11): SelfUpdateService + GET /api/version + POST /api/self-update (409/503, re-exec post-respuesta) + OpenAPI 0.3.0 + arch-lint`
4. `feat(HS-11): vista Ajustes real + tarjeta «Versión y actualización» (RF-100..103·105) + stories=test`
5. `docs(HS-11): boton-actualizar — validación real + PARIDAD ✅ + cierre`

## Retomar aquí

- **Último hecho (2026-07-07, v2):** el operador preguntó «¿dónde estará el botón?» →
  **decisión #5 PROPUESTA**: vista global **Ajustes** (⚙ al pie del rail; hoy ComingSoon
  `global-view.tsx:14`) — la tarjeta la estrena; futuras tarjetas (marketplaces · daemon)
  punteadas. Mockup **v2** publicado (misma URL) con el caso 00 UBICACIÓN: frame del
  shell (rail + pie ⌂/⟳/⚙ con Ajustes activo → vista con la tarjeta). Click-through
  ojo-UI hecho con Playwright headless (el browser MCP seguía tomado): 6 casos renderizan,
  interactividad 00/01 recorre pasos→éxito, consola limpia; screenshot en scratchpad.
- **Base v1:** 5 casuísticas de estados sobre datos reales: binario HOY en
  `/usr/bin/arnesia` (root, .deb) — el diseño migra a `~/.local/bin` (una migración
  inicial, después cero sudo); huella `1c7443f`; el binario NO embebe versión aún
  (pendiente `-ldflags`, se decide en spec).
- **Mockup v2 FIRMADO (2026-07-07)** → specs escritos el mismo turno: `spec.md`
  (RF-100..107: vista Ajustes nace · tarjeta identidad honesta vía GET /api/version ·
  estados no-actualizables · POST /api/self-update con 5 pasos y corte al primer fallo ·
  reinicio re-exec + polling · seguridad cementada: withAuth, cero params, 409) +
  `design.md` (tokens de tarjeta/checklist/pills · máquina de estados · FSD
  features/self-update UI-pura + página transporte · hexagonal ports.SelfUpdater +
  adapter selfupdate + usecase · flag --repo). **Nota clave de honestidad:** la
  checklist «en vivo» del mockup era demo — la app pinta la checklist con los
  veredictos REALES de la respuesta (jamás pasos animados inventados); dicho en spec
  RF-104 y design §A11y.
- **PAQUETE FIRMADO (2026-07-07).** Prompt riguroso de arranque persistido en
  [`PROMPT-implementacion.md`](./PROMPT-implementacion.md) (contexto → diseño técnico
  sellado → implementación Go-primero → validación REAL: el daemon actualizándose a sí
  mismo desde ~/.local/bin). PARIDAD.md esqueleto (8 filas).
- **FASE 0+1 hechas (2026-07-07, conversación de implementación):** paquete + código
  existente leídos; diseño técnico contrastado RF↔existente y SELLADO arriba (tabla de
  piezas + mini-plan de 5 commits); decisiones #6–#9 de FASE 1 registradas en
  decisiones.md (bundle.sh --daemon-only+bin/arnesia · huella +sucio · exePath capturado
  al construir · lock retenido tras actualizado / polling 1.5s solo /api/version).
- **Próximo paso:** FASE 2 — implementar Go primero (commits 1–3), luego FE (commit 4),
  gates verdes; después FASE 3 validación real. Al cerrar: PARIDAD ✅, «Retomar aquí»
  actualizado.
- **Firmas pendientes:** ninguna hasta la paridad (gate final tras implementar).
- **Contexto caliente (candidatos a decisión en la iteración):** ① la ruta del repo
  NO viaja en el request — el daemon la conoce por flag/env/registro explícito
  (superficie-local-confinada); ② el self-update requiere toolchain local (go+pnpm) —
  feature de operador-dev, rotulada; ③ reinicio = write-tmp → rename → re-exec (rename
  sobre binario corriendo es legal en Linux); la UI reconecta por /healthz; ④ migración
  inicial /usr/bin→~/.local/bin: una vez, documentada (PATH: ~/.local/bin precede);
  ⑤ versión visible = huella git embebida por ldflags al build.
