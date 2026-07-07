---
regla: fe-tokens-contrato
version: 1.1
updated: 2026-07-06
status: enforced
ledger: HS-09
sources:
  - url: https://www.designtokens.org/tr/2025.10/format/
    autoridad: estándar
    revisado: 2026-07-05
  - url: https://styledictionary.com/reference/utils/dtcg/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://tailwindcss.com/blog/tailwindcss-v4
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://github.com/AndyOGo/stylelint-declaration-strict-value
    autoridad: experto
    revisado: 2026-07-05
enforced_by:
  - web/.stylelintrc.json#strict-value
  - .github/workflows/ci.yml:tokens-sync
severity: high
---

# Los design tokens son el contrato mockup↔código (DTCG SSOT, cero valores mágicos)

## L1 · Principio (estándar de industria)

**W3C Design Tokens (DTCG), formato estable 2025.10** (primera versión estable 28-oct-2025; media type
`application/design-tokens+json`). Los tokens como single-source-of-truth versionable, modelo 3-capas:
primitive (raw) → **semantic (intent, lo que consumen los componentes)** → component. *(estándar: DTCG
2025.10)*

**Pipeline con un build-step:** `.tokens.json` DTCG → **Style Dictionary v5** → CSS (`@theme` de
Tailwind) + tipos TS. Corre en build, no viaja en el bundle. *(oficial: Style Dictionary DTCG)*

**Tailwind v4 CSS-first (`@theme`):** los tokens SON CSS custom properties nativas; Oxide (Rust) da
bundle 5-20 KB (crítico bajo `go:embed`), zero-runtime. *(oficial: Tailwind v4)*

**Enforcement anti-magic-value:** el código no debe usar hex/valores crudos; stylelint
`declaration-strict-value` obliga `var(--…)`. *(experto: stylelint-declaration-strict-value)*

## L2 · Realización (este árbol Go+React)

- **SSOT = `web/tokens/*.tokens.json` (DTCG, dual-theme).** Verificado: los valores extraídos calzan
  exacto con `mockups/arnesia-shell-A-galaxia.html` (el mockup firmado = fuente de VALORES). Semánticos
  base: `--bg --surface --surface2 --line --line2 --text --muted --accent --accent-soft`. Dominio: tipo
  de componente (`--c-{skill,agent,hook,rule,command,mcp,plugin,settings,output-style,statusline}` = las
  **10 clases** de `domain.Clase`, + `--c-knowledge` para el eje conocimiento; mapea a `knowledge/`), salud de trazas
  (`--ok/warn/crit` + variantes `-soft` de fill, valores del mockup shell-A L7-8/L18-19), desempeño
  (**escala `--heat-1..4` de v3**, no el `--heat` colapsado de shell-A). ⇐ L1: DTCG 3-capas.
- **Convención de nombres = shadcn** (`--background/--foreground/--border` como capa semántica
  canónica) para no mantener dos vocabularios con shadcn/assistant-ui; los mockups aportan valores, no
  nombres. **Divergencia declarada** vs los nombres del mockup (`--bg/--surface`): se renombra a la
  convención shadcn en el `.tokens.json`. **Vocabulario shadcn COMPLETO** (curado del mockup firmado,
  2026-07-05): `background/foreground · card(+fg) · popover(+fg) · primary(+fg) · secondary(+fg) ·
  muted(+fg) · accent(+fg)+accent-soft · destructive(+fg) · border · input · ring` + grupo `sidebar/*`
  (Command Rail) + `chart-1..5` (categórico) + dominio `kind/health(+soft)/heat` + escalas
  `radius/space/text/font`. Cero magic-value: los componentes consumen solo estos tokens.
- **Pipeline:** `web/tokens/*.tokens.json` → Style Dictionary v5 (`web/style-dictionary.config.mjs`,
  formato custom `css/theme-modes` = `:root` + `:root[data-theme=dark]` desde `$extensions.mode.dark`) →
  `web/src/app/styles/theme.css` + `web/src/shared/config/tokens.ts`; Tailwind v4 los consume vía
  `@theme inline` en `index.css`. ⇐ L1: pipeline. **Materializado (scaffold, 2026-07-05):** config, script
  `tokens:build` y `package.json` existen; `theme.css`/`tokens.ts` se commitean (runnabilidad pre-install)
  y se regeneran con el comando. `tokens-sync` de `ci.yml` sigue `if: hashFiles(...)` hasta el primer
  `npm install`. **HS-09 (Fase D):** regeneradas las 5 clases config que faltaban
  (`command/plugin/settings/output-style/statusline`) → `theme.css` emite las **10 clases** en light+dark
  (gap 6→10 resuelto); antes caían a `--muted-foreground` (grises indistinguibles).
- **Styling = Tailwind v4** (`@tailwindcss/vite`); descartados vanilla-extract/CSS-Modules (sin
  enforcement de tokens, reescriben el CSS firmado) y CSS-in-JS (runtime cost). React Flow 12 (ya en
  CSS-vars) y CodeMirror consumen los mismos tokens. ⇐ L1: Tailwind v4.
- **Dark mode:** `@media (prefers-color-scheme: dark)` como señal + `:root[data-theme="light|dark"]`
  que siempre gana; Zustand+hash-state guarda la preferencia. **Las 4 capas del Mapa
  (Estructura/Tokens/Desempeño/Proceso) NO son 4 paletas** — un `data-capa` reasigna un token indirecto
  (`--node-fill`/`--node-accent`); salud/tipo/heat quedan como tokens semánticos reutilizables.
- **Enforcement doble:** (a) Tailwind solo expone el tema en utility-land; (b) stylelint
  `declaration-strict-value` prohíbe valores mágicos en CSS a mano (overrides RF/nodos/CodeMirror) +
  check de **sync `.tokens.json` ⟷ `@theme` generado** (diff del output de SD5 vs committed) → detecta
  divergencia mockup↔código. ⇐ L1: anti-magic-value.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| tokens-dtcg-ssot | existe `web/tokens/*.tokens.json` DTCG y es la fuente (no CSS vars sueltas) | warn | «token fuera del SSOT DTCG» | tokens-sync |
| no-magic-value | ningún `color/background/border/fill` en CSS a mano usa hex/valor crudo, solo `var(--…)` (box-shadow queda FUERA del strict-value — deuda `--shadow-*` declarada honesta, no es color) | error | banda Base «valor mágico (rompe el contrato de tokens)» | stylelint#strict-value |
| tokens-sync | el `@theme` generado está en sync con el `.tokens.json` (diff vacío) | error | «tokens divergen del mockup firmado» | ci.yml:tokens-sync |
| heat-escala-v3 | el token de desempeño usa la escala `--heat-1..4`, no `--heat` colapsado | warn | capa Desempeño «heatmap de una sola parada» | revisión |

## Changelog

- 2026-07-06 · v1.1 · `proposed → enforced` (HS-09, Fase D). El Mapa MVP consume solo `var(--…)`;
  stylelint `strict-value` verde sobre el CSS del canvas. Gap **6→10 clases** resuelto: regeneradas
  `command/plugin/settings/output-style/statusline` en `theme.css` (light+dark). Reconciliado el check
  `no-magic-value` a **solo-color** (box-shadow queda como deuda `--shadow-*` honesta, no color). Sin
  cambio de checks (4).
- 2026-07-05 · v1.0 · Nodo fundacional FE (HS-05). L1 = DTCG 2025.10 + Style Dictionary v5 + Tailwind v4
  + stylelint strict-value. L2 = SSOT `web/tokens/*.tokens.json` (valores del mockup firmado, nombres
  shadcn — divergencia declarada), pipeline SD5→`@theme`+TS, dark mode `data-theme`, capas del Mapa como
  recoloreo, doble enforcement (stylelint + sync). 4 checks.
