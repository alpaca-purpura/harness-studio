---
regla: permisos-gui-human-in-the-loop
version: 1.3
updated: 2026-08-01
status: proposed
ledger: HS-04
sources:
  - url: https://code.claude.com/docs/en/permission-modes
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/agent-sdk/permissions
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://code.claude.com/docs/en/agent-sdk/user-input
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestNoBypassPermissions
  - adapters/agent/claudecode/conductor_test.go:TestPermissionArgsMaterialization
  - fitness/arch_test.go:TestWriteRequiresApproval
  - fitness/arch_test.go:TestSessionSpawnsInArnesPath
  - fitness/arch_test.go:TestArnesPathContainment
severity: critical
---

# Deny-by-default; el GUI aprueba cada escritura vía diff

## L1 · Principio (estándar de industria)

**Human-in-the-loop con deny-by-default.** En una automatización que edita archivos, el modo
`bypassPermissions` no protege contra prompt injection y no debe usarse en un conductor de cara al
usuario. El patrón seguro: correr en `default` (ahora «Manual»), deny-by-default, y cuando el
agente pide un `Write`/`Edit`, recibir el input, **mostrar el diff y dejar que el humano apruebe**
antes de escribir. Orden de evaluación fijo: PreToolUse hook → deny → ask → mode → allow →
`canUseTool`/prompt → PostToolUse (los auto-aprobados **no** llegan al callback). *(oficial:
permission-modes, agent-sdk/permissions, user-input)*

## L2 · Realización (este árbol Go+React)

Aterriza los journeys J2/J4 del UX (diff-before-save → beta → tren) y el principio 10 (nada sin
control):

- **El conductor corre en `default`, deny-by-default.** `--allowedTools` = solo read-only genuino;
  **Write/Edit nunca pre-aprobados**. ⇐ L1: deny-by-default.
- **El canal va SIEMPRE cableado (DD-1, 2026-08-01).** Un `PermissionSet` de valor cero (sesión
  sobre material sin sello → sin rol) emite igual `--permission-mode default
  --permission-prompt-tool stdio`: nada pre-aprobado, TODO pasa por la tarjeta del panel. El
  comportamiento previo (set vacío = cero flags) dejaba esa sesión headless sin canal: CC
  auto-negaba Write y el permiso jamás llegaba al Dock — read-only de facto.
- **El GUI es el human-in-the-loop.** Claude pide edit → `control_request:can_use_tool` con el
  input (`file_path`/`old_string`/`new_string`) → el dock **pinta el diff (CodeMirror merge)** →
  el usuario aprueba → `control_response` allow (echo `updatedInput`) / deny (con `message`). ⇐ L1.
- **Modo por fase:** `plan` en grill/spec (explora, rutea edits a aprobación) · `default`/Manual en
  edición interactiva · `dontAsk` en evals-gate/promote (sin humano presente) ·
  **`bypassPermissions` NUNCA** en un conductor de cara al usuario. `acceptEdits` solo dentro de un
  worktree aislado que se revisa después.
- **Guardrails gratis:** los paths protegidos (`.git`, `.claude`, `.mcp.json`, dotfiles) nunca se
  auto-aprueban salvo bypass — y bypass no se usa. Aislar sesiones por cwd/worktree.
- **⚠ verificar antes de construir (frente B):** el shape JSON-RPC de `control_request` está
  semi-documentado (oficial solo para el SDK, issue #24594) → primer spike de fase 4 contra el
  binario instalado.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| canal-siempre-cableado | todo spawn del Dock emite `--permission-mode default --permission-prompt-tool stdio`, incluso con set de valor cero (sin rol) | error | «sesión sin canal de permisos — escritura auto-negada sin HITL» | conductor_test.go:TestPermissionArgsMaterialization |
| no-bypass | ningún path de código pasa `--dangerously-skip-permissions`/`bypassPermissions` en el conductor de cara al usuario | error | banda Guardia «automatización con permisos saltados» | arch_test.go:TestNoBypassPermissions |
| write-requiere-aprobacion | Write/Edit no están en `--allowedTools`; pasan por el diff-approval del GUI | error | «escritura auto-aprobada sin diff» | arch_test.go:TestWriteRequiresApproval |
| modo-por-fase | evals-gate/promote corren `dontAsk`; grill/spec corren `plan` | warn | «fase sin humano corriendo en modo interactivo (o viceversa)» | arch_test.go |
| allowedtools-readonly | `--allowedTools` solo lista tools read-only | warn | «allowlist incluye tools que escriben» | arch_test.go:TestWriteRequiresApproval |
| max-turns-siempre | toda corrida del conductor fija `--max-turns` | error | «sin turn-cap — loop/costo runaway» | arch_test.go:TestMaxTurnsAlways |
| sesion-aislada-por-cwd | cada sesión spawnea `claude` en la ruta de SU arnés (WorkdirResolver), nunca un cwd global compartido | error | «N sesiones en un cwd → se pisan archivos + contención ~/.claude» | arch_test.go:TestSessionSpawnsInArnesPath |
| cwd-path-contenido | la ruta registrada de un arnés pasa validación (absoluta, dir existente, no `/`/`$HOME`/`~/.claude`/`~/.ssh`) | error | banda Guardia «cwd de sesión sobre ubicación protegida» | arch_test.go:TestArnesPathContainment |

## Changelog

- 2026-08-01 · v1.3 · **DD-1 (deuda D del dogfood developer-vitalia): el canal de permisos va
  SIEMPRE cableado.** `permissionArgs` con set de valor cero emitía cero flags («Dock unchanged»,
  deliberado en Fase E) — pero una sesión sobre material sin sello (sin rol) quedaba headless SIN
  `--permission-prompt-tool`: CC auto-negaba Write/mkdir y el prompt jamás llegaba al panel; la
  forja conversacional era read-only de facto. Ahora el set vacío emite `--permission-mode default
  --permission-prompt-tool stdio` (nada pre-aprobado, todo por la tarjeta). +1 check
  (`canal-siempre-cableado`). El «rol de forja» (perfil explícito para material sin sello) queda
  como evolución del modelo de roles. 7 → **8 checks**.
- 2026-07-07 · v1.2 · **realizado en vivo (HS-11/Fase E).** El diff-approval reservado desde v1.0
  aterriza: el adapter claudecode REENVÍA `control_request:can_use_tool` (antes lo descartaba) como
  evento normalizado; el daemon lo pinta como tarjeta `permission` del Dock por SSE (D3) y `POST
  /api/sessions/{id}/permission` (reemplaza el 501) responde `control_response` por stdin — allow
  eco del input / deny con message, deny del rol gana al click. `write-requiere-aprobacion` pasa de
  t.Skip a test REAL (`TestWriteRequiresApproval`: Write/Edit filtrados de `--allowedTools` aunque
  el rol los permita + loop humano con fakes + grant efímero sin re-pregunta), que también cubre
  `allowedtools-readonly` (celda actualizada, sin-test → con-test). El shape del control channel
  sigue semi-documentado (issue #24594): implementación best-effort contra el protocolo del Agent
  SDK, incertidumbre anotada en el adapter. `status` sigue `proposed` honesto: `modo-por-fase`
  (plan/dontAsk por fase) aún no tiene enforcement real.
- 2026-07-05 · v1.1 · **HS-06 — la prosa L2 «aislar sesiones por cwd/worktree» se vuelve
  ejecutable.** La auditoría del shell Tauri v1 encontró que el conductor corría TODO en un cwd
  global (arnés = metadata cosmética). Fix: puerto `WorkdirResolver` + registro explícito arnés→ruta
  (`internal/adapters/store/arnes_registry.go`), con fallback aislado por arnés (nunca cwd global) y
  denylist de paths protegidos. **+2 checks** (`sesion-aislada-por-cwd`, `cwd-path-contenido`) +
  `max-turns-siempre` ahora satisfecho por el conductor (`--max-turns`). El diff-approval /
  `--permission-mode` sigue reservado al spike de permisos (HS-07). 5 → **7 checks**.
- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = human-in-the-loop deny-by-default. L2:
  conductor en `default`, diff-approval en el GUI vía `control_request`, modos por fase, bypass
  jamás, `--max-turns` siempre; spike pendiente del shape JSON-RPC. 5 checks.
