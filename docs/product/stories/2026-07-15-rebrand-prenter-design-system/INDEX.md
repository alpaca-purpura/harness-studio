# Rebrand — sistema de diseño PRENTER (paquete de trabajo)

> `tipo: rebrand-ui` · Origen: el operador pidió llevar el look & feel de TODA la app al sistema de
> marca de PRENTER (Prenter Group), extraído en vivo de Claude Design (`mcp__claude_design`, proyecto
> «PRENTER Design System», `a98c2e0d-db82-43f2-8fd7-e7e05c40fd51`). Disciplina METODOLOGIA §10:
> mockup→decisiones→spec→implementar→PARIDAD, firmas 🧑‍⚖️ entre etapas. **Esto REEMPLAZA valores de una
> base ya firmada** (HS-03 shell + HS-05/HS-09 tokens DTCG, `arnesia-shell-A-galaxia.html` = 🔒 firmado
> como fuente de VALORES en `mockups/INDEX.md`) — no es un superset incremental, por eso el proceso
> completo con gate humano es obligatorio antes de tocar `web/tokens/base.tokens.json` de producción.

## Qué se extrajo (fuente: Claude Design, proyecto PRENTER Design System)

`tokens/{colors,typography,spacing,fonts}.css` + `readme.md` (brandbook resumido). Identidad
**dark-first**: negro puro `#000000` + teal `#00b7aa` como único acento saturado. Tipografía Jost
(display, sustituto de Coco Gothic con licencia) + Mulish (cuerpo, sustituto de Sansation) + JetBrains
Mono (motivo terminal/binario de marca). Radios más generosos (sm4/md8/lg14/xl22/pill) y sombras suaves
de bajo contraste + `--shadow-brand` teal. Voz de marca: trato de usted, sentence case, cero emoji.

## 3 decisiones ya firmadas por el operador (mismo turno, ver `decisiones.md`)

1. **La paleta funcional del Mapa queda intacta** — `color.kind` (10 tipos de componente),
   `color.health` (ok/warn/crit), `color.heat` NO se recolorean a teal; son codificación de datos, no
   identidad de marca.
2. **Modo por defecto de la app pasa a dark** (fiel a la identidad real de PRENTER) — light queda como
   alterno, no al revés.
3. **Proceso: paquete formal completo** — este mismo paquete, con gate humano antes de aplicar a
   producción.

## Retomar aquí

> **Estado (2026-07-22): IMPLEMENTADO en `web/`, gate humano 🧑‍⚖️ FIRMADO** — ver
> [`paridad.md`](./paridad.md). Queda deuda visible (mockup `arnesia-shell-A-galaxia.html` sin
> re-derivar, botones/CTA en `font.sans`) → `BACKLOG.md`. Decisiones D1-D9
> resueltas (D1-D3 firmadas por el operador vía preguntas, D4-D5 hallazgos técnicos, D6-D8 resolvieron
> las 3 preguntas abiertas de `propuesta-tokens.md`, D9 documenta 2 correcciones WCAG reales cazadas
> por el suite de a11y — ver `decisiones.md`). Aplicado: `web/tokens/base.tokens.json` (colores light+
> dark · `font.display` nuevo · grupo `shadow` nuevo · radios) → `theme.css`/`tokens.ts` regenerados →
> `index.css` sincronizado (bloque `@theme` literal + regla global `h1-h4{font-display}`) → 3 fuentes
> self-hosted (`web/public/fonts/*.woff2`, variable, subset `latin`) → `app-store.ts` dark-first
> default. **Verificado en vivo, no solo tipeado:** `pnpm run verify` verde · `vitest --project=
> storybook` 127/127 (incluye 24 fallos de contraste reales cazados y corregidos, D9) · `--project=unit`
> 21/21 · Storybook corrido en `:6006` y capturado con screenshot real (Portafolio Lista + Wizard, light
> y dark) — Jost confirmado vía `getComputedStyle`. `color.kind`/`health`/`heat` (Mapa) intactos,
> verificado en el `theme.css` generado.
>
> **NO tocado (a propósito, D8/D3):** botones/CTA por-módulo (`.pf-btn-*` etc) siguen en `font.sans`
> — deuda explícita, sin QA visual por superficie. `mockups/arnesia-shell-A-galaxia.html` (🔒 firmado)
> y su fila en `mockups/INDEX.md` NO se re-derivaron — ese es el acto de cierre del gate, no se simula.
>
> **Gate FIRMADO 2026-07-22** (ver `paridad.md`). Pendiente como deuda separada, NO bloqueante:
> re-derivar `arnesia-shell-A-galaxia.html` + re-estampar `mockups/INDEX.md` + decidir alcance de
> botones/CTA (D8) — ver `BACKLOG.md`.

## Archivos

- [`decisiones.md`](./decisiones.md) — D1-D5, con el porqué de cada una.
- [`propuesta-tokens.md`](./propuesta-tokens.md) — mapeo completo PRENTER→`base.tokens.json`
  (light+dark), tabla revisable, preguntas abiertas. **Gate humano PENDIENTE antes de implementar.**
