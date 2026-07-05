# web/ — SPA de ArnesIA (scaffold vacío)

Primera versión de la aplicación: **Vite + React + TypeScript**, envuelta por **Tauri 2**
(`src-tauri/`), embebible por `go:embed` en el daemon `arnesia`. Topología **FSD-lite**
(arch/boundaries/fe-topologia-fsd.md). **Vacía a propósito**: el shell (Command Rail · lienzo del
Mapa · dock de conversación) se construye en la próxima sesión — aquí queda todo el cableado listo.

## Correr (requiere `npm install` primero — sin red en el scaffold)

```bash
cd web
npm install
npm run dev            # Vite en :5173 (placeholder del shell)
npm run storybook      # Storybook 10 (story = test)
npm run verify         # typecheck + biome + depcruise + steiger + stylelint
npm run tokens:build   # regenera src/app/styles/theme.css + src/shared/config/tokens.ts
# app de escritorio (con toolchain Rust):  cd src-tauri && cargo tauri dev
```

> Versiones de `package.json` = rango razonable a jul-2026; **fijar/lock al primer `npm install`**.

## Estructura FSD (import direccional, enforced por dependency-cruiser + steiger)

```
src/
├─ app/        composition root: main.tsx, App.tsx (placeholder), styles/ (Tailwind v4 + theme.css)
├─ pages/      composition-roots por hash-state (SIN router) — vacío
├─ widgets/    Command Rail · Mapa · dock · Portafolio — vacío
├─ features/   conversar · crear/editar caja · evaluar A/B · publicar — vacío
├─ entities/   arnés · caja · corrida · empresa/puesto — vacío
└─ shared/     ui/ (primitivos shadcn sobre tokens) · lib/ · store/ (Zustand+hash) · config/ (tokens)
```

Dirección permitida: `app → pages → widgets → features → entities → shared` (solo hacia abajo).

## Tokens (contrato mockup↔código)

SSOT = `tokens/base.tokens.json` (DTCG 2025.10; valores = mockup firmado shell-A). Pipeline
`npm run tokens:build` (Style Dictionary v5, `style-dictionary.config.mjs`) → `src/app/styles/theme.css`
(`:root` + `:root[data-theme=dark]`) + `src/shared/config/tokens.ts`. Esas salidas están
**committeadas** para runnabilidad pre-install; se regeneran con el comando. Tailwind v4 las consume
vía `@theme inline` en `src/app/styles/index.css`. Dark por `data-theme` (no `.dark`).

## Stack

React 19 · Vite 6 · Tailwind v4 · shadcn sobre Base UI · React Flow 12 (`@xyflow/react`) ·
Zustand 5 · Storybook 10 (+ addon-vitest, addon-a11y) · Biome 2.4 · TS strictest · Tauri 2.
