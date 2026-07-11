# Research — Arneses-as-code: cómo otros lo resolvieron (Wave 2 · 2026-07-10)

> Insumo de co-diseño. Objetivo: robar patrones probados para init/doctor · sello/deriva · loop
> forward · conocimiento co-locado low-token · packaging instalable. 5 investigaciones web.

---

## A · BMAD-METHOD + agent-teams (CrewAI/MetaGPT/Swarm)
- **BMAD v6 sharded:** cada workflow = dir con `SKILL.md` (entrypoint/L2) + `step-01..NN.md` (camino
  determinista) + frontmatter que persiste variables. Catálogo **`module-help.csv`** (menu-code, phase
  1-4, preceded-by/followed-by, output-location, required) → skill `bmad-help` recomienda la siguiente
  transición legal según estado del repo. Config = **merge 3-capas** (`customize.toml` base →
  `{skill}.toml` team → `.user.toml`). **Expansion packs = módulos** instalables (`marketplace.json` +
  `PluginResolver`). Agente = persona+commands+**dependencies** lazy-load.
- **Patrones agent-team:** CrewAI (roles YAML: role/goal/backstory + tasks) · MetaGPT (SOP por rol en
  assembly-line + **shared message pool** con structured outputs) · Swarm (routines + **handoff = función
  que devuelve otro agente**) · AutoGen (group-chat emergente).
- **ROBAR:** espina fases→cajas→ROL ≈ phases + preceded/followed-by (catálogo CSV navegable low-token);
  handoff tipado (caja posee 1 transición); merge 3-capas = init/doctor; expansion-pack = unidad vendible.
- **EVITAR (confirma "NO clon BMAD"):** 80% token-spend re-inyectando standards cada invocación
  (issue #511, ~230M tok/sem) → nuestro grafo low-token es la anti-tesis; 6-7 personas hard-coded (curva
  ~2 meses); PRD monolítico. Fuentes: github.com/bmad-code-org/bmad-method · issues/511 · arxiv 2308.00352.

## B · Spec-driven / AI-native SDLC
- **GitHub Spec Kit:** `.specify/{memory/constitution.md, templates/, scripts/}` + `specs/NNN/{spec,plan,
  tasks}.md`; flujo `/constitution→/specify→/clarify→/plan→/tasks→/implement` + `/analyze` (consistencia)
  + `/converge` (gap vs código).
- **Amazon Kiro:** specs (requirements/design/tasks) + **steering** (`.kiro/steering/*.md`: product/tech/
  structure) con activación por front-matter: `inclusion: always | fileMatch(+glob) | manual(#ref) | auto
  (por description)`. Hooks por evento.
- **OpenSpec:** `specs/` (verdad/SSoT) vs `changes/` (propuestas aisladas); merge de **deltas semánticos**
  (ADDED/MODIFIED/REMOVED) solo al `archive`; agente consulta runtime (`status --json`).
- **Tessl:** spec-as-source + **Spec Registry** ("NPM del conocimiento", specs versionadas).
- **ROBAR:** front-matter Kiro `always|fileMatch|manual|auto` = **nuestra activación low-token exacta** ·
  OpenSpec delta-first (verdad vs propuesta, merge=firma) = **sello+deriva + PARIDAD** (mata
  código-listo-sin-firma) · Spec Kit templates-por-caja · constitution.md = conocimiento de gobierno ·
  `/converge`+capabilities=SSoT casi idéntico · Tessl Registry = marketplace. Fuentes: github.com/github/
  spec-kit · kiro.dev/docs/steering · github.com/Fission-AI/OpenSpec · tessl.io.

## C · Packaging Claude Code + estándares
- **Plugin** = `.claude-plugin/plugin.json`; **marketplace** = `.claude-plugin/marketplace.json`. Primitivos
  en raíz: `skills/<n>/SKILL.md` (frontmatter `description` + cuerpo; model-invoked) · `agents/*.md`
  (subagent: frontmatter+system-prompt+tools-allowlist) · `hooks/hooks.json` (Pre/PostToolUse, Stop,
  SessionStart…) · `commands/*.md` · `.mcp.json` · `.lsp.json` · `monitors/monitors.json` · `settings.json`
  (puede activar un agent como main-thread).
- **Progressive disclosure 3 capas:** metadata name+desc (~500 tok arranque) → cuerpo SKILL.md al activar
  (<500 líneas) → referencias bajo demanda. 8 skills = 500 tok vs 70k.
- **AGENTS.md** (Linux Foundation, 28+ tools): markdown plano, anidable, **closest-file-wins**.
  **Cursor rules** (`.cursor/rules/*.mdc`: description/globs/alwaysApply) 4 modos: always / auto-attached
  (glob) / agent-requested (description) / manual. **Windsurf**: mismos 4 vía `trigger`.
- **Harnesses comunitarios:** SuperClaude (solo .md; comandos×personas) · **Agent OS** (spec-driven;
  `discover-standards`/`inject-standards` extraen convenciones del código y las inyectan; role-based con
  verification-chains; "implementar→emerge patrón→standard→informa próximo spec") · claude-flow (SPARC +
  hive-mind).
- **ROBAR:** cada DIMENSIÓN = un `SKILL.md` con `description` selector (co-locación física + Index-First) ·
  activación por glob/description · Agent OS discover/inject-standards = **init** · `plugin validate` +
  `--plugin-dir` + `/reload-plugins` = **doctor** · AGENTS.md = fachada multi-herramienta del CLAUDE.md.
- **DIFERENCIAL confirmado:** nadie en la comunidad tiene **sello+deriva/drift-gate/honestidad
  automática** — es nuestro. Fuentes: code.claude.com/docs/en/plugins · agents.md · cursor.com/docs/rules ·
  buildermethods.com/agent-os.

## D · Knowledge / context-as-code (auto-derivado low-token)
- **aider repo-map:** grafo generado en runtime — tree-sitter (tag-queries `.scm`) extrae `def`(superficie
  pública)+`ref`(usos) → grafo símbolo→símbolo → **PageRank** rankea → búsqueda binaria mete cuantos tags
  quepan en `map_tokens` (default 1024). Mapa de *paths+firmas*, no cuerpos.
- **llms.txt:** índice curado (H1+blockquote+H2 links a .md) → fetch **just-in-time** de la hoja. `llms-
  full.txt` = corpus opuesto.
- **Casa con capabilities=SSoT:** los YAML de capability ya son hojas atómicas derivadas del código; el
  knowledge-graph = capa de navegación (INDEX + punteros auto-derivados como repo-map deriva de AST).
- **Schema de "hoja de conocimiento" recomendado (fundado):**
```yaml
---
id: tecnologias/tauri
tipo: reference            # Diátaxis: reference|how-to|explanation|tutorial
dimension: tecnologias
resumen: "Tauri v1 = shell desktop; Rust+webview."   # L1 → va al INDEX (~1 línea)
activacion: glob           # always|glob|demand|manual  (Cursor/Kiro)
globs: ["src-tauri/**"]
punteros_auto:             # GENERADOS (tree-sitter/deps), no tecleados
  - src-tauri/src/lib.rs#run
  - capability: desktop/self-update
budget_tok: 400
---
# cuerpo escueto (L2): hechos atómicos, 1 concepto. Sin historia (→LEDGER), sin cifras a mano.
## Ver también (L3, fetch on-demand)
```
Reglas: INDEX-first (un `INDEX.md`/llms.txt por dimensión = solo resumen+link) · hoja atómica (≤~400 tok) ·
punteros auto-derivados+rankeados+verificados en CI (link vivo o falla) · carga 3-4 saltos. Fuentes:
aider.chat/docs/repomap.html · llmstxt.org · anthropic.com/engineering (skills, context-engineering) · diataxis.fr.

## E · Scaffolding / golden-path + drift/update (init/doctor/sello/loop-forward) ★ la mina
- **cruft** (`.cruft.json`: template URL + **commit-hash** + vars) · **copier** (`.copier-answers.yml`:
  `_src`+`_commit` + **respuestas resueltas**). `update` = **three-way merge git**: (1) re-render template
  versión-VIEJA con respuestas guardadas → "old-generated"; (2) diff old-generated vs working-tree =
  **captura tu deriva local**; (3) re-render versión-NUEVA; (4) aplica el parche. Conflictos → markers/`.rej`
  (gap VISIBLE). `skip_if_exists` = seed-once (asegurar presencia, jamás pisar); archivos borrados se excluyen.
- **projen:** re-síntesis idempotente desde `.projenrc.ts`; **`PROJEN_MARKER`** marca propiedad → solo
  limpia lo que él generó, **nunca toca código de usuario**; anti-tamper CI.
- **Backstage:** scaffolder = pipeline de `actions` + `catalog:register` (registro/descubrimiento).
- **Netflix Wall-E / paved-road:** init te pone on-road, doctor te mantiene; off-road permitido pero
  **visible y sin soporte** = deriva honesta.
- **ROBAR (mecanismo directo):**
  - **Sello** ← `.copier-answers.yml`/`.cruft.json` endurecido: fuente + **versión canónica pineada** +
    inputs resueltos + **firma** + **hash por-archivo** del render canónico (baseline).
  - **Deriva por-archivo** ← copier old-render + projen marker: `modificado-usuario` (difiere de
    old-generated) · `generado`/`original` (idéntico) · `no-reconocido` (sin match) · seed-once (`skip_if_exists`).
  - **Doctor** ← projen re-síntesis idempotente + limpieza-por-marker + copier `update` 3-way para lo
    modificado. Sobreescribe `generado`, no toca `no-reconocido`.
  - **Drift-gate** ← projen anti-tamper + `cruft check` en CI = tu honestidad automática N2.
  - **Migrations** ← copier/Nx `migrations.json`: transforms versionados que evolucionan archivos **Y el
    esquema del sello**.
  - **Marketplace** ← Backstage catalog:register.
- **APORTE GENUINO:** el **loop-forward invierte cruft** (template→instancias): la fábrica **cosecha** el
  diff local, lo promueve a canónico y re-publica → las instancias corren `update` y su artificio se
  reemplaza. **Nadie automatiza la cosecha-back** — es nuestro. El sustrato (answers+pin+3-way) sí se roba.
- **EVITAR:** Yeoman (sin link ni pin) · editar answers a mano (frágil → firmar/hashear) · projen posee el
  proyecto ENTERO (nosotros = **propiedad parcial**, solo lo que el proceso exige) · migrations forward-only
  sin cosecha-back · Backstage fire-and-forget (mantener el sello VIVO tras instalar = nuestro diferencial).
Fuentes: cruft.github.io/cruft · copier.readthedocs.io/en/stable/updating · projen.io · nx.dev · backstage.io/
docs/features/software-templates · infoq.com/news/2021/03/spotify-paved-paths.
