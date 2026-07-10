---
elemento: settings
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/settings
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/permissions
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/permission-modes
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/security
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://www.backslash.security/blog/claude-code-security-best-practices
    autoridad: experto
    revisado: 2026-07-04
  - url: https://www.mintmcp.com/blog/claude-code-security
    autoridad: experto
    revisado: 2026-07-04
---

# Settings & permisos (settings.json · permissions · modes)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** `settings.json` es la config en capas: permisos (allow/deny/ask), **modos**
de permiso, env, registro de hooks, modelo, etc. A diferencia de CLAUDE.md/skills/rules (que
moldean lo que Claude *intenta*), **los permisos los enforca Claude Code, no el modelo** — un
hook ni un prompt pueden pisar un deny. Es **el límite de seguridad real** de la superficie.
*(oficial: settings + permissions)*

**Scopes y precedencia (L1.2, alto→bajo; las reglas de permiso MERGEAN — deny > ask > allow
desde cualquier scope):** managed/enterprise (no-overridable) > CLI args > local
`.claude/settings.local.json` (gitignore) > proyecto `.claude/settings.json` (git) > usuario
`~/.claude/settings.json`. *(oficial)*

**Sintaxis de reglas (L1.3):** `Tool` o `Tool(specifier)`. `Bash(npm run test:*)` prefijo;
bare `Bash`/`Bash(*)` = todo. `Read(./.env)`, `Read(./secrets/**)` gitignore-style;
`WebFetch(domain:*.example.com)`; `mcp__server__*`; `Agent(model:opus)`; `Cd(~/code/*)`. deny/ask
se evalúan antes que allow, primer match gana. *(oficial: permissions)*

**Modos de permiso (L1.4):** `default`/`manual` (solo lectura) · `acceptEdits` (auto-aprueba
edits in-scope) · `plan` (lectura + plan) · `auto` (research preview; clasificador bloquea
`curl|bash`, force-push, escritura a secret-manager) · `dontAsk` (auto-DENIEGA lo no
pre-allowed, para CI) · `bypassPermissions` (salta casi todo; requiere flag, se niega como
root fuera de sandbox). **Paths protegidos** (`.git`, `.claude`, `.env`, shell rc…) nunca
auto-aprobados salvo bypass. *(oficial: permission-modes)*

**Prácticas (L1.5):** allow **solo** comandos 100% inofensivos (nunca `git push`/`docker run`/
installs → esos a `ask`) · **deny agresivo de secretos**: `Read(./.env)`, `Read(./secrets/**)`,
`Read(~/.ssh/**)`, `Bash(curl *)` o red solo por `WebFetch(domain:trusted)` · managed settings
para piso org non-overridable (`allowManagedPermissionRulesOnly`) · **nunca secretos en
settings.json versionado** (usar `.local.json` o `apiKeyHelper`) · bypass solo en contenedor
aislado sin internet (`disableBypassPermissionsMode: "disable"` para prohibirlo) · sandbox OS
(`/sandbox`) como defensa en profundidad · retención corta (`cleanupPeriodDays` 7–14 en repos
sensibles) · auditar edición de settings en sesión con hook `ConfigChange`. *(oficial + Backslash + MintMCP)*

**Trampas (L1.6):** patrones de argumento frágiles (`Bash(curl http://x/ *)` es bypasseable por
reorden de flags/redirects — usar deny+WebFetch) · prefijos NO son transitivos sobre `&&`/`;`/`|`
(`Bash(safe *)` no cubre `safe && rm -rf /`) · confiar en CLAUDE.md para prohibir en vez de un
deny (advisory ≠ enforced) · `enableAllProjectMcpServers: true` auto-confía servers. *(oficial + Backslash)*

**Novedades (L1.7):** modo `auto` (v2.1.83+, clasificador, block-lists creciendo hasta v2.1.200)
· reglas `Cd` (v2.1.169) · deny/ask por parámetro `Tool(param:value)` · glob `mcp__*`/`"*"` ·
`manual` alias de `default` · `dontAsk` formalizado CI · sandbox `filesystem`/`network` mergea con
deny rules · **workspace-trust gatea allow-rules de proyecto** (endurecido tras CVE, fix 2.1.200)
· lockdown managed: `strictPluginOnlyCustomization`, `disableSideloadFlags`, `forceRemoteSettingsRefresh`. *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Settings/permisos = la **banda Guardia** del mapa junto con [[hooks]]: la infraestructura
transversal que protege TODAS las cajas (VISION A6). Es donde el arnés declara su postura de
seguridad, y donde ArnesIA es más estricto porque es el único límite real.

1. **Least-privilege enforced, no advisory:** un arnés nuestro nunca lleva `Bash(*)` en allow
   de scope compartido; ops riesgosas van a `ask`, secretos a `deny`. ⇐ L1.5. Check de error.
2. **Deny de secretos obligatorio:** `.env`/`secrets/**`/`~/.ssh/**` bloqueados de arranque.
   ⇐ L1.5. Sin ellos = crítico de Guardia.
3. **Bypass jamás en repo compartido:** `bypassPermissions` solo local/contenedor; en
   `.claude/settings.json` versionado = error. ⇐ L1.4/L1.5.
4. **La regla la enuncia CLAUDE.md, el permiso la aplica:** si una regla de banda Base dice
   «nunca X» sobre acción peligrosa, debe existir el deny/hook que la enforce. ⇐ L1.6 +
   [[rules]] L2.2. Cross-check entre capas.
5. **Cero secretos en config versionada:** aplica a settings.json Y `.mcp.json` ([[mcp]] L2.3).
   ⇐ L1.5.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| perm-no-blanket-bash | bare `Bash`/`Bash(*)` en allow (proyecto/usuario) | error | banda Guardia «permiso Bash total (riesgo)» | L1.3 · L2.1 |
| perm-secrets-deny | deny cubre `.env`, `.env.*`, `secrets/**`, `~/.ssh/**` | error | banda Guardia «sin bloqueo de secretos» | L1.5 · L2.2 |
| perm-bypass-shared | `bypassPermissions` en `.claude/settings.json` versionado | error | banda Guardia «bypass compartido en repo» | L1.4 · L2.3 |
| perm-network-open | `curl`/`wget` permitidos sin deny ni `WebFetch(domain:)` allowlist | warn | banda Guardia «salida de red sin acotar» | L1.5 |
| perm-secret-in-env | env de settings.json versionado con valor tipo `_KEY`/`_TOKEN`/`_SECRET` | error | banda Guardia «secreto en texto plano» | L1.5 · L2.5 |
| perm-mcp-unscoped | `enableAllProjectMcpServers: true` o `mcp__*` sin acotar | warn | banda Guardia «MCP sin allowlist» | L1.6 |
| perm-fragile-arg | allow intenta acotar URL/dominio inline en vez de deny+WebFetch | info | banda Guardia «patrón de permiso frágil» | L1.6 |
| perm-no-managed | org sin managed settings / `allowManagedPermissionRulesOnly` | warn | banda Guardia «sin política gerenciada» | L1.5 |
| perm-retention | `cleanupPeriodDays` sin fijar o >30 en repo sensible | info | banda Guardia «retención de transcripts larga» | L1.5 |
| perm-advisory-only | CLAUDE.md prohíbe una acción sin deny/hook que la aplique | warn | banda Base «regla solo advisory, no aplicada» | L1.6 · L2.4 |
| perm-local-conflict | `.local.json` allow lo que `.json` compartido deniega para el mismo tool | info | banda Guardia «conflicto local vs compartido» | L1.2 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (settings, permissions,
  permission-modes, security) + Backslash + MintMCP. L2 amarra settings/permisos = banda Guardia,
  límite de seguridad real, least-privilege enforced, cross-check regla-advisory↔deny.
  11 checks. Novedades: modo `auto` v2.1.83, reglas `Cd`, deny por parámetro, workspace-trust
  gatea allow, lockdown managed. · pasada fundacional.
