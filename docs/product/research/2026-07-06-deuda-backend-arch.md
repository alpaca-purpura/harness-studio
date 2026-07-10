# Deuda paralela — diseño de arquitectura del track backend (HS-09)

> **Scope y gate:** el track del **Mapa está PARADO en el gate 1 (mockup, sin firmar)** — no se toca su
> diseño ni su código. Este documento avanza el track de **Deuda paralela**, que el goal declara
> *«independiente del Mapa, puede correr en paralelo»*. Es **análisis de arquitectura (read-only) + diseño**,
> NO programación de producto ni edición de `arch/` boundaries (eso se ratifica en Fase D con el operador).
> Grounded en el código real a `a8721e9`. Sirve para arrancar Fase E con arquitectura ya pensada.

## Estado verificado (leído del código)

- `ports.ArtifactReader` = `Status(ctx, artifactRef) (status string, exists bool, err error)` — lee el
  `status:` (document-as-cache) del artefacto, **nunca el chat** (`conductor-no-parsea-jsonl`). **Sin adapter
  concreto.**
- `usecase.BoxConductor` (loop T3) YA consume `ArtifactReader` + `AgentPort`; owns el control flow
  (`repairCap`, `--max-turns`, `blocked→handoff`, `RutaSiguiente`). Solo instanciado en tests, **no en `main.go`**.
- `permission.KitProvisioner` impl `ports.PermissionPort.ResolveForRole(rol) → PermissionSet`; política spike
  por rol (backend-dev / reviewer / base deny-by-default) con TTL. `domain.PermissionSet.Decide(tool)` =
  Deny>Ask>Allow, unlisted→Ask (deny-by-default). `Grant.Vigente(now)` = TTL. **No cableado; `resolvePermission`
  HTTP = 501.**

## Item 1 · Cablear conductor + permisos (el adapter `ArtifactReader`)

**Adapter concreto** `internal/adapters/artifact/reader.go`:
- `Status(ctx, artifactRef)`: resolver `artifactRef` (p.ej. `spec.md`, de `contract.entrega[0].art`) a un path
  **dentro del cwd confinado del arnés** vía `ports.WorkdirResolver` (S2, `store.ArnesRegistry`) → leer el
  frontmatter YAML del archivo → extraer `status:`. `exists=false` si el archivo aún no está (caja no ha
  entregado). Cero parseo de JSONL/chat. Reusa `yaml.v3` (ya en go.mod).
- **Boundary:** cae bajo `conductor-no-parsea-jsonl` (lee artefacto, no traza) + `orquestacion-determinista-entre-cajas`.
  Test de fitness: el adapter lee `status` de un fixture y NUNCA abre el `.jsonl` de sesión.

**Superficie de invocación (decisión de diseño abierta):** el `BoxConductor` es ejecución **autónoma T3** de una
caja — distinta del `SessionService` interactivo (conversación por turno). Cablearlo «al daemon» necesita una
puerta de entrada:
- **Opción A (recomendada):** endpoint `POST /api/harnesses/{id}/boxes/{boxId}/run` → arranca el loop T3 de esa
  caja, emite progreso por SSE `run` (el `EventRun` hoy sin publisher), devuelve `BoxOutcome`. Cablea
  `NewBoxConductor(agent, artifactReader, repairCap, maxTurns)` en `runServe`.
- Opción B: el conductor corre dentro de una sesión existente (fold en `SessionService`) — más acoplado, menos claro.
- **A** mantiene el conductor T3 como ciudadano propio (coherente con el modelo de cajas) y estrena el canal SSE `run`.

**Permisos en el spawn:** `KitProvisioner.ResolveForRole(arnes.rol)` → `PermissionSet` → pasar a
`ports.SpawnOpts` (nuevo campo `Permission`) → el conductor CC lo traduce a `--permission-mode`/`--settings` al
spawnear. Así «el mismo tool decide distinto por rol» se materializa en el proceso `claude`.

## Item 2 · Endpoint HTTP `control_request` (role/ttl)

`POST /api/sessions/{id}/permission` (hoy `resolvePermission`→501) = el **control_request** del protocolo
stream-json de CC (human-in-the-loop). Diseño:
- **Request** (lo que CC emite por el control channel): `{ tool_name, input, ... }`. El adapter del agente
  (`claudecode`) reenvía el `control_request` de CC a este puerto en vez de auto-resolverlo.
- **Resolución:** `PermissionSet` del rol de la sesión → `Decide(tool)`:
  - `allow` → responde permitir (sin UI).
  - `deny` → responde denegar (hook exit 2 / respuesta deny).
  - `ask` → **human-in-the-loop**: emite un frame SSE (`dock`/`run`) a la UI, espera la decisión del operador,
    y al aprobar **mintea `Grant` con TTL** (`NuevoGrant`); mientras `Vigente`, futuras llamadas del mismo tool
    en la sesión no re-preguntan.
- **Response** al control channel: `{ behavior: "allow"|"deny", updatedInput?, message? }` (shape del control
  protocol de CC — verificar el exacto al implementar contra el binario instalado; era el spike que HS-04 reservó).
- **Boundary:** `permisos-derivan-del-rol` + `permisos-gui-human-in-the-loop`. TTL = `permisos-derivan-del-rol`
  check «least temporal privilege».

## Item 3 · go-arch-lint operativo

- Añadir `docs/architecture/fitness/.go-arch-lint.yml` (mapa de componentes: `domain`, `ports`, `usecase`, `adapters/*`,
  `cmd`) + reglas de dependencia (domain no importa nada interno; usecase no importa adapters; etc.).
- Asegurar el binario (`go install github.com/fe3dback/go-arch-lint@latest` o pin) + step en CI.
- Efecto: los 3 boundaries que hoy defieren por `go-arch-lint` ausente pasan a enforced.

## Item 4 · `deferred`→real (CI) + enforcers rotos

- **`.github/workflows/ci.yml`:** correr `go build/vet/test -race`, y en `web/`: `biome check`, `tsc --noEmit`,
  `dependency-cruiser`, `stylelint`, `storybook` story-tests. Cablea los 187 nl-judge + linters externos que hoy
  cuentan `deferred` en `conformance --todo`.
- **Enforcers rotos a arreglar (destapados en Fase B):**
  - `web/.dependency-cruiser.js`: usa `module.exports` en repo `type:module` (ESM) → **"module is not defined"**;
    renombrar a `.cjs` o convertir a ESM export. ADEMÁS globs `canvas-not-chrome`/`chrome-not-canvas-internals`
    apuntan a rutas **stale** (`widgets/command-rail|dock`, `app/shell` inexistentes) → corregir a
    `widgets/session-rail|chat-dock|topbar|view-strip` + `pages/shell` (y sumar la nueva `widgets/map-canvas` +
    `shared/canvas` como lado canvas).
  - `vitest.workspace.ts`: vitest 4 dropeó `defineWorkspace`/workspace file → migrar a `test.projects` en un
    `vitest.config.ts` + `.storybook/vitest.setup.ts` (`setProjectAnnotations`) + añadir `@storybook/addon-vitest`
    a `.storybook/main.ts` addons. Rompe la referencia `vitest.workspace.ts#storybook` de
    `fe-visual-fitness.md` → actualizar el `enforced_by`.

## Boundaries tocados (para ratificar en Fase D)

| boundary | cambio |
|---|---|
| `conductor-no-parsea-jsonl` | nuevo adapter `ArtifactReader` real → check pasa de declarado a enforced |
| `orquestacion-determinista-entre-cajas` | `BoxConductor` cableado en `runServe` (endpoint run) |
| `permisos-derivan-del-rol` | `KitProvisioner` en el spawn + `control_request` con TTL |
| `permisos-gui-human-in-the-loop` | endpoint `control_request` ask→UI→grant |
| `adaptadores-de-agente-intercambiables` + 2 | go-arch-lint operativo → enforced |
| `fe-visual-fitness` | `enforced_by` de story-tests migrado a `vitest.config.ts` |

## Preguntas abiertas para el operador (deuda)

1. **Superficie del conductor T3:** ¿endpoint `POST /boxes/{id}/run` (Opción A) o fold en sesión (B)?
2. **`control_request` ask→UI:** ¿el frame de aprobación va por el dock existente o una superficie propia de permisos?
3. ¿La deuda entra por **worktree** (aislada) o directo en `main`? (el goal menciona worktrees para el track paralelo).

*No se ha escrito código ni editado `arch/` — esto es el diseño previo. La ejecución (Fase E) espera el orden del
plan (tras los gates + Fase D) o tu visto bueno para arrancar la deuda en paralelo ya.*
