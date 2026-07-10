# PROMPT de implementación — paquete boton-actualizar (copiar a una conversación nueva)

> Persistido por disciplina §10: la conversación jamás es el único registro.
> Paquete FIRMADO por el operador el 2026-07-07 (mockup v2 + spec/design).

---

Implementa el paquete FIRMADO `historias/2026-07-07-boton-actualizar/` (botón
«Actualizar» — self-update sin sudo) con rigurosidad y en fases. El código no se toca
hasta cerrar la Fase 1.

**FASE 0 — Contexto (leer, en este orden):**
1. `historias/2026-07-07-boton-actualizar/INDEX.md` (flujo, estado, «Retomar aquí»)
2. `decisiones.md` (5 firmadas + #4 cementada = LEY) · `spec.md` (RF-100..107 + Gherkin
   + trazabilidad) · `design.md` (UI al pixel + hexagonal) · `PARIDAD.md` (8 filas) ·
   `mockup-actualizar.html` v2 = la verdad visual (ábrelo; artifact en INDEX)
3. Código existente: `web/src/pages/shell/ui/global-view.tsx` (la vista a estrenar) ·
   `web/src/shared/api/client.ts` · `web/src/app/styles/inspector.css` (patrón CSS
   scoped + tokens) · Go: `cmd/arnesia/main.go` (runServe: flags, wiring) ·
   `internal/adapters/transport/http/router.go` · `internal/adapters/artifact/`
   (patrón de adapter confinado + tests) · `internal/usecase/fuente_service.go`
   (patrón sentinels+servicio) · `scripts/bundle.sh` · `docs/architecture/contracts/api/openapi.yaml`
   · `docs/architecture/fitness/.go-arch-lint.yml` (añadir componente `selfupdate`)
4. Arquitectura que RIGE: `superficie-local-confinada` (endpoint bajo withAuth, CERO
   parámetros en el request) · `dominio-independiente-de-transporte` · FSD
   (`fe-topologia-fsd`: feature UI-pura, transporte en la página — patrón Inspector) ·
   `docs/architecture/conventions/` (estilo Go/TS, commits `feat(HS-11): …`)

**FASE 1 — Revisión + diseño técnico (sin código):** contrasta cada RF contra lo
existente; decide dónde vive cada pieza cumpliendo design.md: puerto `ports.SelfUpdater`
+ `GET /api/version` (buildinfo VCS: `debug.ReadBuildInfo()` del proceso; el binario
NUEVO se inspecciona sin ejecutarlo) · adapter `internal/adapters/selfupdate/` (build
con cwd=repo cancelable · verificación del binario · instalar write-tmp→rename atómico
EN EL MISMO dir de `os.Executable()`) · `usecase.SelfUpdateService` (orquesta, corta al
primer fallo, reporte por paso, mutex→409, `Reiniciar()` re-exec post-respuesta) ·
flag `--repo`/env `ARNESIA_REPO` · rutas + OpenAPI (bump) · FE: `features/self-update/
ui/update-card.tsx` UI PURA (props+callbacks) · página = transporte + polling
healthz/version del reinicio · CSS tokens `var(--…)`. Cuidados especiales: el re-exec
NO debe matar la respuesta HTTP en vuelo (responder → flush → agendar); «ya al día»
(huella igual) termina SIN reinstalar; binario no-escribible / sin-repo = estados
honestos con botón disabled (RF-102/103). Emite mini-plan de commits y séllalo en el
INDEX antes de codear.

**FASE 2 — Implementación por RF (orden):** Go primero (RF-107 version → RF-104/106
self-update + tests -race con repo fake y binario temporal) → FE (RF-100 vista ·
RF-101..103 tarjeta · RF-105 polling) con story=test por estado (idle · actualizando ·
error · no-escribible · sin-repo · éxito). Gates que DEBEN quedar verdes: `tsc` ·
`biome ci src` · `depcruise` · `steiger` · `stylelint` · story-tests (vitest
--project=storybook) · `go build/vet/test -race ./...` · golangci-lint · go-arch-lint
(con el componente `selfupdate` añadido). Commits `feat(HS-11): …` a main, uno por
bloque coherente.

**FASE 3 — Validación REAL (no negociable):** la paridad de ESTE paquete es el daemon
ACTUALIZÁNDOSE A SÍ MISMO de verdad: ① `scripts/bundle.sh` → `install -m755
bin/arnesia ~/.local/bin/` (la migración única del operador — documentarla en el
paquete); ② arrancar ESE binario (`~/.local/bin/arnesia serve --repo <repo>`) + vite;
③ con Chrome DevTools: ⚙ Ajustes → tarjeta con huella/ruta/repo REALES → hacer un
commit trivial en el repo (p.ej. touch de un comentario) → click «Actualizar desde el
repo» → checklist real → reinicio → la tarjeta muestra la huella NUEVA. Verificar
también: error de build real (romper algo a propósito y restaurar) · «ya al día» ·
409 doble click · estado no-escribible (arrancar el binario de /usr/bin si sigue) ·
sin `--repo`. Lado a lado con el mockup v2, screenshots al scratchpad, consola limpia.
PARIDAD fila → ✅ solo tras verificar; desviación = se registra y se consulta, jamás
se maquilla. OJO: valida contra el binario INSTALADO, no `go run` (el self-update
reemplaza `os.Executable()`).

**Reglas permanentes:** METODOLOGIA §10 (todo al archivo del paquete en el MISMO
turno; «Retomar aquí» al día) · lo staged disabled y rotulado, jamás finge · tokens
DTCG only · contraste AA ambos temas · checklist SOLO con veredictos reales (la
animación del mockup era demo — RF-104) · alcance = el spec exacto (no-goals escritos)
· seguridad de decisión #4/RF-106 al pie de la letra.

Al terminar: PARIDAD completa, gates verdes listados con su salida, resumen honesto de
desviaciones, commit/push, INDEX «Retomar aquí» actualizado.

El éxito se da cuando todos los valores de las pruebas de validación y verificación
sean exitosos.
