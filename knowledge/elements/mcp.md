---
elemento: mcp
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/mcp
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/managed-mcp
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/security
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://modelcontextprotocol.io/introduction
    autoridad: estándar-abierto
    revisado: 2026-07-04
  - url: https://cheatsheetseries.owasp.org/cheatsheets/MCP_Security_Cheat_Sheet.html
    autoridad: experto
    revisado: 2026-07-04
---

# MCP (Model Context Protocol servers)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Estándar abierto (origen Anthropic) para conectar el agente a tools/datos/
workflows externos — «un puerto USB-C para IA». Un server expone **tools** (funciones),
**resources** (`@server:protocol://path`) y **prompts** (slash commands `/mcp__server__prompt`).
*(oficial: modelcontextprotocol.io)*

**Alta y scopes (L1.2).** `claude mcp add --transport http|sse|stdio ...`; `.mcp.json` en root =
scope proyecto (git-shared, requiere aprobación interactiva). Scopes: local (default, `~/.claude.json`)
> proyecto (`.mcp.json`) > user > plugin > connector. `${VAR}`/`${VAR:-default}` en command/args/
env/url/headers. *(oficial: code.claude.com/docs/en/mcp)*

**Transportes (L1.3):** stdio (local, filesystem) · HTTP/`streamable-http` (cloud, OAuth,
reconnect) · **SSE deprecado** (migrar a HTTP) · WebSocket (solo `add-json`, push de eventos). *(oficial)*

**Auth (L1.4):** OAuth 2.0 (`/mcp` o `claude mcp login`), DCR por defecto; `oauth.scopes` para
least-privilege; `headersHelper` para esquemas no-OAuth. Enterprise: `managed-mcp.json` +
`allowedMcpServers`/`deniedMcpServers` por `serverUrl`/`serverCommand` (match por **nombre NO es
control de seguridad** — es spoofeable). *(oficial: managed-mcp)*

**Naming y permisos (L1.5):** `mcp__<server>__<tool>`; reglas `mcp__server` (todo el server),
`mcp__server__*` (wildcard), `mcp__server__tool` (uno), `mcp__*` (todo MCP). *(oficial: permissions)*

**Costo de contexto (L1.6, clave):** **Tool Search default-ON** en Claude Code — tools
`defer_loading`, solo nombres + instrucciones cargan al inicio; reduce ~85% el costo de defs
(ej. 55k→~1k tok para GitHub MCP lazy). Mantener solo 3–5 tools críticas non-deferred
(`alwaysLoad`). Cargar todo upfront = ~55–77k tok para 5–10 servers antes de trabajar; agregar
30–50+ tools sin search **degrada la selección** del modelo. `MAX_MCP_OUTPUT_TOKENS` default 25k.
*(oficial: tool-search-tool + mcp)*

**Seguridad (L1.7):** Anthropic **no audita** servers MCP (solo revisa el Directory). Superficie:
tool poisoning (instrucciones ocultas en descripciones/schemas/returns), rug-pull (server cambia
tras aprobación), confused-deputy, prompt injection vía outputs. Scan Equixly 2025: **43%** de
servers populares con command-injection, 22% path traversal, 30% SSRF. Marcar tools destructivas
`_meta["anthropic/requiresUserInteraction"]:true` (prompt no bypasseable, v2.1.199+). *(oficial
security + OWASP + Checkmarx)*

**Novedades (L1.8):** Tool Search default-on (anunciado 2026-01-14) · `defer_loading`/`alwaysLoad`
first-class · SSE deprecado · connector API `mcp-client-2025-11-20` (reemplaza 2025-04-04),
tool allow/deny movido a `mcp_toolset` · WebSocket · **Channels** (server pushea a la sesión) ·
Elicitation (input estructurado mid-task) · auto-flatten de root `anyOf/oneOf` (v2.1.195) ·
`requiresUserInteraction` (v2.1.199). *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Los MCP suelen ser **terceros** en un arnés nuestro (banda Terceros del mapa, ya modelada en
it.8 con `mcp-tessl`). No protagonizan excepciones de proceso (atenuados con «—» en capa
Proceso). Su métrica dominante es **costo de contexto vs uso**.

1. **Deferred-loading por defecto:** un arnés nuestro con ≥10 tools MCP corre Tool Search;
   `alwaysLoad` solo para las 3–5 críticas. ⇐ L1.6 + principio 11. Un MCP que fuerza carga
   upfront sin justificarlo = hallazgo de costo.
2. **MCP frío = candidato a desconectar:** server con 0 invocaciones en 30d paga contexto/
   superficie y no rinde. ⇐ L1.6/L1.7. Ya es hallazgo real en luana (`mcp-tessl` sin uso).
3. **Cero secretos en `.mcp.json` versionado:** `${VAR}`/`headersHelper`/OAuth, jamás literal.
   ⇐ L1.4/L1.7. Check de Guardia.
4. **Solo servers de fuente confiable/allowlist:** al importar un arnés (conformación §6) se
   verifica procedencia del server; untrusted = superficie de prompt injection. ⇐ L1.7.
5. **stdio con versión fijada:** `npx -y pkg` sin pin = riesgo rug-pull/supply-chain. ⇐ L1.7.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| mcp-unused | server con 0 invocaciones en 30d | warn | badge «MCP sin uso (paga contexto, no rinde)» | L1.6 · L2.2 |
| mcp-toolsearch-off | ≥10 tools o >10k tok de defs con Tool Search off / `alwaysLoad` no-crítico | warn | «carga tools sin deferred — costo evitable» | L1.6 · L2.1 |
| mcp-secret-in-config | `.mcp.json` versionado con token/key literal en env/headers | error | banda Guardia «secreto en config versionada» | L1.4 · L2.3 |
| mcp-sse-deprecated | server con `type:"sse"` | info | «transporte SSE deprecado — migrar a HTTP» | L1.3 |
| mcp-untrusted | `serverUrl`/`serverCommand` fuera de allowlist / Directory | warn | «server no verificado — superficie prompt injection» | L1.7 · L2.4 |
| mcp-unpinned-stdio | comando stdio `npx -y pkg` sin versión | warn | «comando MCP sin pin — rug-pull/supply chain» | L1.7 · L2.5 |
| mcp-broad-oauth | `oauth.scopes` sin fijar con auth server de scopes amplios | warn | «alcance OAuth amplio sin restringir» | L1.4 |
| mcp-destructive-no-confirm | tool con verbo destructivo sin `requiresUserInteraction` | warn | «tool destructiva sin confirmación forzada» | L1.7 |
| mcp-name-reserved | server `workspace` (reservado) o duplicado con def divergente | error | «nombre inválido/reservado — se ignora» | L1.5 |
| mcp-output-unbounded | tools disparan `MAX_MCP_OUTPUT_TOKENS` repetido sin `maxResultSizeChars` | info | «salidas grandes sin acotar — infla contexto» | L1.6 |
| mcp-no-governance | escala enterprise sin `managed-mcp.json`/`allowedMcpServers` | info | «sin gobierno central de MCP» | L1.4 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (mcp, tool-search-tool, managed-mcp,
  security) + modelcontextprotocol.io + OWASP/Checkmarx. L2 amarra MCP = tercero, métrica = costo
  vs uso, deferred-loading por defecto, frío = desconectar (conecta hallazgo real mcp-tessl).
  11 checks. Novedades: Tool Search default-on (2026-01-14), SSE deprecado, Channels, Elicitation,
  `requiresUserInteraction` v2.1.199. · pasada fundacional.
