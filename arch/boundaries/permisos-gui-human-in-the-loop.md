---
regla: permisos-gui-human-in-the-loop
version: 1.0
updated: 2026-07-05
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
  - fitness/arch_test.go:TestWriteRequiresApproval
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
| no-bypass | ningún path de código pasa `--dangerously-skip-permissions`/`bypassPermissions` en el conductor de cara al usuario | error | banda Guardia «automatización con permisos saltados» | arch_test.go:TestNoBypassPermissions |
| write-requiere-aprobacion | Write/Edit no están en `--allowedTools`; pasan por el diff-approval del GUI | error | «escritura auto-aprobada sin diff» | arch_test.go:TestWriteRequiresApproval |
| modo-por-fase | evals-gate/promote corren `dontAsk`; grill/spec corren `plan` | warn | «fase sin humano corriendo en modo interactivo (o viceversa)» | arch_test.go |
| allowedtools-readonly | `--allowedTools` solo lista tools read-only | warn | «allowlist incluye tools que escriben» | arch_test.go |
| max-turns-siempre | toda corrida del conductor fija `--max-turns` | error | «sin turn-cap — loop/costo runaway» | arch_test.go |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-04). L1 = human-in-the-loop deny-by-default. L2:
  conductor en `default`, diff-approval en el GUI vía `control_request`, modos por fase, bypass
  jamás, `--max-turns` siempre; spike pendiente del shape JSON-RPC. 5 checks.
