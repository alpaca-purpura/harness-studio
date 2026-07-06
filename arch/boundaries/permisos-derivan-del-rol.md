---
regla: permisos-derivan-del-rol
version: 1.1
updated: 2026-07-05
status: enforced
ledger: HS-08
sources:
  - url: https://architect.salesforce.com/docs/architect/fundamentals/guide/enterprise-agentic-architecture.html
    autoridad: experto
    nota: "Permisos task-based que EXPIRAN; guardrails enforced-at-reasoning-layer (no prompt); identidad verificable."
    revisado: 2026-07-05
  - url: https://arxiv.org/abs/2603.18916
    autoridad: académica
    nota: "Agentic BPM — frame normativo (deóntico): el permiso es autoridad del rol, no default."
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestPermissionSetParametrizedByRole
severity: high
---

# El permission-set del arnés se parametriza por el ROL que lo hidrata

## L1 · Principio (estándar de industria)

**Permisos = autoridad del rol, impuestos fuera del razonamiento (Salesforce + frame normativo APM).**
El conjunto de permisos de un agente no es un default fijo: es **lo que la autoridad del rol
autoriza**, y el **mismo** artefacto lleva permisos distintos según el rol que lo ejecuta. Se imponen
en una capa que el modelo **no puede razonar para esquivar** (hooks / `PreToolUse` / `canUseTool`), y
pueden ser **task-based y expirar** (mínimo privilegio temporal), no acceso perpetuo. *(experto:
Salesforce enterprise-agentic-architecture; académica: arXiv 2603.18916 — frame normativo deóntico)*

## L2 · Realización (este árbol Go+React)

Nuestros permisos hoy son **safety-céntricos** (`permisos-gui-human-in-the-loop`: deny-by-default, el
GUI aprueba writes vía diff). Este boundary añade el eje **rol-céntrico**, coherente con «producto
puro + META de enganche» (VISION §Linaje): la autoridad del rol viene de la **META** del arnés
(rol·proceso·reporta-a·empresa, ya en `graph.l0`), NO de una consultoría MOF nuestra.

- **Parametrización por hidratación:** el `KitProvisioner` (research inyección §8.1) resuelve el
  permission-set según el rol que hidrata y lo pasa como `--permission-mode`/`--settings` al spawn.
  Extiende el `--permission-mode <según modo>` de fase→rol. ⇐ L1 (permisos = f(rol)).
- **Impuesto en hook, no en prompt:** el límite duro vive en `PreToolUse`/`control_request` (exit 2 /
  deny), nunca en la prosa del arnés. ⇐ L1 + `permisos-gui` L2 + frontera P6/Guardia (METODOLOGIA §8.5).
- **Permiso efímero (spike HS-07):** el `control_request:can_use_tool` puede conceder un permiso con
  TTL/por-tarea, no perpetuo. ⇐ L1 (task-based expiring). *(depende del spike de permisos.)*
- **La autoridad es externa:** el rol y su autoridad los provee el futuro sistema L1 (organigrama);
  ArnesIA solo consume la META de enganche. `iam-por-agente` (identidad verificable por-instancia con
  credenciales reales) = concern **DevHub**, fuera de la fábrica. ⇐ VISION §Linaje (seam).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| permisos-parametrizados-por-rol | el permission-set lo resuelve el `KitProvisioner` según el rol que hidrata, no un default fijo | warn | banda Guardia «permisos no derivados del rol» | arch_test.go:TestPermissionSetParametrizedByRole |
| guardrail-en-hook-no-prompt | el límite duro vive en `PreToolUse`/`control_request` (exit 2/deny), no en prosa | error | banda Guardia «límite solo advisory — no aplicado» | arch_test.go:TestNoBypassPermissions |
| permiso-efimero-ttl | los grants (`Grant.Vigente`) expiran / son por-tarea, no perpetuos | info | banda Guardia «permiso perpetuo — mínimo privilegio temporal ausente» | arch_test.go:TestPermissionSetParametrizedByRole |
| meta-de-enganche-completa | el arnés declara META completa (rol·proceso·reporta-a·empresa) — el seam del sistema L1 externo | warn | «META incompleta — no enganchable» | box.contract/graph.l0.schema.json |

## Changelog

- 2026-07-05 · v1.1 · **status → `enforced` (HS-08, spike aterrizado).** `domain.PermissionSet` +
  `permission.KitProvisioner` resuelven el permission-set por rol (el MISMO tool decide distinto según
  el rol) y `domain.Grant.Vigente` hace los grants efímeros (TTL/por-tarea, no perpetuos).
  `TestPermissionSetParametrizedByRole` lo enforça. `meta-de-enganche-completa` se valida contra el
  schema graph.l0 (arnés declara rol·proceso·reporta_a·empresa, required). Falta la superficie HTTP
  `control_request` con role/ttl (endpoint) — anotado en openapi como pendiente.
- 2026-07-05 · v1.0 · Nodo draft (HS-07, doctrina v1). L1 = permisos = autoridad del rol impuesta
  fuera del razonamiento (Salesforce + frame normativo APM). L2: `KitProvisioner` parametriza por rol;
  autoridad externa (META de enganche / sistema L1 futuro); permiso efímero = spike `control_request`;
  IAM-por-agente = DevHub. 4 checks · `status: proposed`. · doctrina v1 (VISION §Linaje).
