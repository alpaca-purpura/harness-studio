# Editar con Claude Code — Fase E completa gobernada

> Ficha HS-09 (fase 5) · plan Hito 2 · 2026-07-06
> Decisión firmada: **Fase E completa** (no el slice mínimo). Se cablean TODAS las piezas del patrón conductor al daemon.

## Estado de partida (verificado por subagente)

Todas las piezas existen **construidas y test-cubiertas**, pero **desconectadas** entre sí y del daemon. El único camino vivo a «modificar con CC» hoy es el Dock interactivo (`POST /api/sessions/{id}/turn` → spawn `claude` confinado al cwd del arnés) **con los permisos propios de CC, sin conductor, sin permisos por rol, sin gate de conformance**.

| Pieza | Estado | Ubicación |
|---|---|---|
| Adapter stream-json | **cableado** (solo al Dock/`SessionService`) | `internal/adapters/agent/claudecode/conductor.go`, `main.go:98` |
| `BoxConductor` (loop T3 determinista) | construido, **NO cableado** (solo tests) | `internal/usecase/box_conductor.go`; `arch_test.go:359` |
| `ports.ArtifactReader` | **interfaz sin adapter** (gap clave) | `internal/ports/conductor.go:13`; no existe `adapters/artifact/` |
| `KitProvisioner` (rol→permisos) | construido, **NO cableado**; sin seam en `SpawnOpts` | `internal/adapters/permission/provisioner.go`; `ports/agent.go:47` |
| `control_request` (role/ttl) HTTP | **stub 501** | `router.go:128`; `openapi.yaml:176` |
| Motor `arnesia conformance` | **callable solo por CLI**; sin ruta HTTP; fuera del camino de edición | `cmd/arnesia/conformance.go`, `usecase/conformance_service.go` |
| Mutación del grafo en el daemon | **ausente** — read-only para arnés/box | `router.go:35-52` |

## Boundaries que este trabajo materializa (as code)

Cablear la Fase E **realiza** boundaries que hoy están enforced-por-fitness pero no-vivos, o proposed. No inventamos doctrina — la ejecutamos:

- `orquestacion-determinista-entre-cajas` (**enforced**) — el loop lo posee el conductor Go, lee `status` de artefacto, **nunca** texto de chat.
- `permisos-derivan-del-rol` (**enforced**) — permission-set = f(rol), deny-by-default, grants con TTL.
- `permisos-gui-human-in-the-loop` (proposed → **realizar**) — cada escritura aprobada por GUI con diff; `--max-turns` siempre; sesión aislada por cwd.
- `conductor-no-parsea-jsonl` (proposed) — el conductor consume stream-json/OTel, no `json.Unmarshal` del transcript.
- `indice-desechable-jsonl-es-verdad` (proposed) — el artefacto en disco es la verdad; el índice se reconstruye.
- `dominio-independiente-de-transporte` / `core-no-importa-shell` / `adaptadores-de-agente-intercambiables` (proposed) — se destraban al instalar **go-arch-lint** (deuda §go-arch-lint).

## Arquitectura objetivo (hexagonal — respeta la topología as-code)

```
                 HTTP (transport)                         use-case                         adapters (outbound)
POST /api/harnesses/{id}/boxes/{boxId}/run  ─►  BoxConductor.Run  ─┬─►  AgentPort (claudecode) ── spawn claude -p --max-turns
POST /api/sessions/{id}/permission (role/ttl) ─► PermissionPort   ├─►  ArtifactReader (NUEVO) ── lee `status:` frontmatter
GET/POST /api/harnesses/{id}/conform        ─►  ConformancePort   └─►  (gate) ConformanceService.Run(TargetArnes)
```

### Piezas nuevas a construir

1. **`internal/adapters/artifact/reader.go`** — implementa `ports.ArtifactReader`. Lee SOLO el `status:` del frontmatter del artefacto de la caja (nunca el chat). Confinado al cwd del arnés (mismo `resolver.Resolve` que la sesión). Test con fixtures de frontmatter.
2. **Seam de permisos en `ports.SpawnOpts`** — añadir el `PermissionSet` resuelto para que el spawn de `claude` lo materialice en `PreToolUse`/`control_request` (nunca en prosa — lo exige el boundary). El adapter `claudecode` debe **reenviar** los frames `control_request`/`canUseTool` de CC (hoy los descarta, `conductor.go:284`).
3. **`POST /api/harnesses/{id}/boxes/{boxId}/run`** — entra al `BoxConductor.Run`: resuelve rol→`PermissionSet` (`KitProvisioner`), spawnea CC con `--max-turns`, corre el loop leyendo `result.subtype` + `ArtifactReader.Status`, rutea por `contract.ruta`, emite eventos SSE (`EventRun`, hoy constante sin publisher, `sse/broker.go:20`).
4. **`POST /api/sessions/{id}/permission` real (role/ttl)** — reemplaza el 501: recibe `{requestId, decision, ttl?, role}`, resuelve vía `PermissionSet.Decide`, minta `Grant` efímero (`domain.NuevoGrant` con TTL), responde por el canal ask→UI.
5. **Gate de conformance en el camino de edición** — tras un `run`/edición, invocar `ConformanceService.Run(TargetArnes)` in-process; el veredicto (G1 schema + G2 spine + escritor único + firewall) **bloquea la promoción** del cambio y se pinta en el Mapa (diagnóstico unificado, VISION §196). Ruta HTTP `GET/POST /api/harnesses/{id}/conform` documentada en OpenAPI.

### Piezas a cablear (ya existen)

- `NewBoxConductor(...)` en `runServe` (`main.go`), con `repairCap`/`maxTurns` de config.
- `NewKitProvisioner(...)` en `runServe`; pasar el set por el nuevo seam de `SpawnOpts`.
- Publisher de `EventRun` en el broker SSE.

## Contratos as-code a actualizar (arch/)

- `docs/architecture/contracts/api/openapi.yaml` — 3 rutas nuevas + esquemas de request/response. Gen-check en CI (`ci.yml:contracts`).
- `docs/architecture/fitness/arch_test.go` — flip de `t.Skip("TODO fase 5")` a real: `TestWriteRequiresApproval`, `TestLiveEventsFromStreamJSON`. Los enforced (`TestConductorOwnsBoxRouting`, `TestPermissionSetParametrizedByRole`, `TestMaxTurnsAlways`, `TestNoBypassPermissions`) deben seguir verdes con el cableado vivo.
- Changelogs de los 5 boundaries tocados (`docs/architecture/boundaries/*.md`) — registrar «realizado en vivo».
- `docs/architecture/fitness/.go-arch-lint.yml` — mapear el nuevo componente `adapters/artifact`.

## Deuda técnica que cierra esta fase

- **go-arch-lint operativo** — instalar el binario (hoy ausente; `docs/architecture/INDEX.md:85`). Destraba 3 boundaries proposed. Fijar el module-path provisional `github.com/alpacapurpura/arnesia`.
- **`deferred` → real en CI** — cablear linters externos para que `arnesia conformance --todo` deje de reportar 211 deferred / 24 real. Honestidad de conteo (memoria: «235 checks 0-fail» era engañoso).

## Riesgos / preguntas abiertas para el operador (Gate 0)

1. **Superficie del run T3** — ¿`POST /boxes/{id}/run` dedicado (recomendado, el conductor es su propio recurso) vs plegar en la sesión? El conductor NO es conversacional; separarlo evita mezclar el loop determinista con el turn humano.
2. **Canal ask→UI del `control_request`** — ¿Dock derecho (reusa el stream vivo) vs panel de permisos dedicado? Recomendado: Dock, con una tarjeta de aprobación de diff inline (assistant-ui + CodeMirror 6/merge, ya en el stack HS-04).
3. **worktree vs main para el track backend** — aislar la Fase E en worktree (mutaciones de archivos en paralelo con el track FE) vs main directo. Recomendado: worktree si los dos tracks corren a la vez.
