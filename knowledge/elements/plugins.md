---
elemento: plugin
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/plugins
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/plugin-marketplaces
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/plugins-reference
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/discover-plugins
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/plugin-dependencies
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://claudefa.st/blog/tools/mcp-extensions/plugins-distribution
    autoridad: experto
    revisado: 2026-07-04
---

# Plugins & marketplaces (empaque y distribución)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Un **plugin** = directorio autocontenido que empaqueta skills/commands/agents/
hooks/MCP/LSP/monitors/themes con manifiesto `.claude-plugin/plugin.json`. Un **marketplace** =
repo git con `.claude-plugin/marketplace.json` que cataloga plugins. Es lo que vuelve el `.claude/`
personal un **producto versionado, instalable y namespaced** (`/plugin-name:skill`). *(oficial: plugins)*

**Layout (L1.2):** solo `plugin.json` va en `.claude-plugin/`; el resto en root: `skills/`,
`commands/`, `agents/`, `hooks/hooks.json`, `.mcp.json`, `bin/` (a PATH). *(oficial: plugins-reference)*

**`plugin.json` (L1.3):** requerido solo `name`. Relevantes: `version` (semver; omitir → cae a
SHA del commit), `description`, `author`, `homepage`, `repository`, `license`, `keywords`,
`dependencies[]` (semver o `{name,version,marketplace}`), `userConfig{}` (`sensitive:true` →
keychain OS), `displayName` (v2.1.143, cosmético). Campos no reconocidos se ignoran (warn con
`--strict`). *(oficial)*

**`marketplace.json` (L1.4):** requeridos `name` (kebab-case; reservados como
`claude-plugins-official` bloqueados), `owner{name}`, `plugins[]`. Cada entry: `name` + `source`
(path `./…` sin `..`, o `{github{repo,ref?,sha?}}`, `url`, `git-subdir`, `npm`). **Con `ref` y
`sha`, gana `sha`.** `renames{}` (v2.1.193) migra plugins renombrados. *(oficial: plugin-marketplaces)*

**Instalación (L1.5):** `/plugin marketplace add owner/repo` → `/plugin install name@marketplace`
(scope user/project/local/managed) → `/reload-plugins`. Equivalentes CLI `claude plugin …`. *(oficial)*

**Prácticas (L1.6):** **subir `version` en cada release** (no confiar en drift silencioso de SHA)
· producción/compartido → pin `sha`, no solo `ref` (branch/tag es force-pusheable) · **dos
marketplaces en refs distintos = canales estable vs latest** · plugin **single-purpose** (una
skill + su hook/MCP, no grab-bag) · documentar `README`/`CHANGELOG` · backflow: dependientes
declaran rango semver, upstream tagea `{name}--v{version}` con `claude plugin tag` · **nunca
secretos** en `.mcp.json`/hooks → `userConfig{sensitive:true}` · validar con `--strict` en CI ·
`renames` al cambiar `name` estable · repos privados para plugins internos. *(oficial + claudefa.st)*

**Anti-patrones (L1.7):** `version` fijo que nunca sube = updates que nunca llegan (misma
version → cache) · `version` en plugin.json Y marketplace = plugin.json gana silencioso · pin
solo `ref` en prod · grab-bag inflando contexto always-on · secretos en config · `source` con
`..` (traversal) · nombres no kebab-case (los rechaza claude.ai) · renombrar sin `renames`
(rompe installs) · instalar de fuente no confiable («plugins... ejecutan código arbitrario con
tus privilegios»). *(oficial)*

**Novedades (L1.8):** grafo de dependencias con rangos semver + cross-marketplace gating (v2.1.110+)
· `renames` (v2.1.193) · `displayName` (v2.1.143) · **transparencia de costo de contexto en `/plugin`**
(token estimate v2.1.143, «Not used recently» v2.1.187) · dos-tier oficial: `claude-plugins-official`
(curado) vs `claude-community` (screening + SHA-pinned) · nuevos componentes: LSP, Monitors, Channels
· sources `npm`/`git-subdir` · `--plugin-dir` acepta `.zip` (v2.1.128) · `claude plugin validate --strict`. *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

El plugin/marketplace es **el vehículo de distribución de un arnés** — VISION lo cementa:
«publica → marketplace git (elegible POR PROYECTO, marketplace.json + semver/SHA)». El arnés de
construcción propio (`harness@prenter-marketplace`) ya vive así. Esto conecta con el **Tren de
release** (S6) y con la regla de backflow I-59.

1. **Semver + SHA pin obligatorio en canal estable:** un arnés nuestro publicado nunca flota en
   HEAD. ⇐ L1.6. El Tren declara el destino con su pin.
2. **Canales estable/latest = dos refs:** mapea 1:1 a nuestro beta→estable (KIT-06). ⇐ L1.6.
3. **Backflow, jamás fork silencioso:** mejoras al arnés se upstreamean al kit; el plugin declara
   su rango semver sobre upstream. ⇐ L1.6 + I-59 (ya doctrina en CLAUDE.md).
4. **Single-purpose y costo de contexto acotado:** un plugin-arnés no es grab-bag; su costo
   always-on se mide antes de rollout (el `/plugin details` token estimate = insumo del mapa).
   ⇐ L1.6/L1.8 + principio 11.
5. **Cero secretos, fuente confiable:** al importar/adoptar (conformación §6) se valida manifiesto
   (`--strict`), pin, y procedencia del marketplace. ⇐ L1.6/L1.7.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| plugin-version-pinned | plugin con `version` semver explícita (o source SHA-pinned), no flotando | warn | badge «plugin sin pin (drift silencioso)» | L1.6 · L2.1 |
| marketplace-valid | `marketplace.json` pasa `claude plugin validate` sin errores | error | validator exit / lista de errores | L1.4 |
| plugin-manifest-fields | declara `name`, `description`, `version`, `author`, `license` | warn | diff de campos faltantes | L1.3 |
| plugin-no-secrets | `.mcp.json`/hooks sin token/key literal; secretos vía `userConfig{sensitive}` | error | banda Guardia «secreto en plugin» | L1.6 · L2.5 |
| plugin-sha-pin | entry `source` con `sha` (no solo `ref`) para canal estable | warn | «canal estable sin SHA — no reproducible» | L1.6 · L2.1 |
| plugin-version-bumped | `version` sube cuando cambió el contenido | warn | «version estancada — update silencioso no llega» | L1.7 |
| plugin-single-purpose | inventario de componentes coherente, no grab-bag | info | «plugin grab-bag — infla contexto» | L1.6 · L2.4 |
| marketplace-trusted | marketplace añadido de fuente allowlisted / conocida | error | «marketplace no confiable — ejecuta código» | L1.7 · L2.5 |
| plugin-renames-tracked | `renames` presente cuando un `name` cambió o se quitó | warn | «rename sin migración — rompe installs» | L1.7 |
| plugin-dep-range | dependencia declara rango semver, no nombre pelado | warn | «dependencia sin rango (implícito latest)» | L1.6 |
| plugin-context-cost | costo always-on (`/plugin details`) bajo el techo pre-rollout | info | «costo de contexto sobre umbral» | L1.8 · L2.4 |
| marketplace-name-ok | `name` kebab-case, no reservado/impersonación | error | «nombre reservado/inválido» | L1.4 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (plugins, plugin-marketplaces,
  plugins-reference, discover, dependencies) + claudefa.st. L2 amarra plugin = vehículo de
  distribución del arnés (VISION marketplace git), semver+SHA, canales beta/estable = KIT-06,
  backflow I-59, single-purpose + costo acotado. 12 checks. Novedades: grafo de dependencias,
  `renames`, transparencia de costo `/plugin`, dos-tier oficial/community, LSP/Monitors/Channels.
  · pasada fundacional.
