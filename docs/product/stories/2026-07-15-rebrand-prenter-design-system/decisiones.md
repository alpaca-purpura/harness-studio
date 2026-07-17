# Decisiones — Rebrand sistema de diseño PRENTER

> `tipo: decisiones` · paquete `2026-07-15-rebrand-prenter-design-system`. D1-D3 conversadas y firmadas
> por el operador en vivo (`AskUserQuestion`, mismo turno). D4-D5 son hallazgos técnicos del arquitecto
> al verificar la fuente real (Claude Design) contra el pipeline real (`fe-tokens-contrato.md`,
> `theme.css` generado) — mismo criterio que los paquetes anteriores: documentadas AQUÍ, no relitigadas
> en silencio.

## D1 · La paleta funcional del Mapa queda intacta — NO se recolorea a teal

`color.kind` (10 tipos de componente de arnés: skill/agent/hook/knowledge/mcp/rule/command/plugin/
settings/output-style), `color.health` (ok/warn/crit + soft) y `color.heat` (escala 1-4) son
codificación categórica de DATOS reales del dominio (`domain.Clase`, salud de trazas, desempeño) —
existen para que 10 tipos distintos se puedan diferenciar de un vistazo en el grafo del Mapa. Migrarlos
todos a variantes de un único acento teal destruiría esa distinción (mismo error que G1 de Slice 1 del
Portafolio advertía para OTRO eje: no fabricar/aplanar dato real). → Estos 3 grupos NO se tocan en este
paquete. Lo que SÍ cambia es el **chrome estructural** que los rodea: `background/foreground/card/
popover/primary/secondary/muted/accent/destructive/border/input/ring` + `sidebar/*` + tipografía +
radios + sombras.

## D2 · Modo por defecto pasa a dark (fiel a la identidad real de la marca)

PRENTER es dark-first por diseño de marca (`readme.md`: "Identidad dark-first: negro puro + teal");
un modo por defecto light-con-acento-teal sería una identidad distinta, no la de PRENTER. El operador
eligió fidelidad de marca sobre menor disrupción. **Cómo se logra sin romper el pipeline firmado
(HS-05/HS-09, ver D5):** el archivo `.tokens.json` SIGUE con la convención `$value` = light /
`$extensions.mode.dark` = dark (Style Dictionary v5 ya lo procesa así, `:root` + `:root[data-theme=
dark]`) — lo que cambia es el estado INICIAL de `data-theme` que la app setea al arrancar (hoy
"light"/preferencia del SO; pasa a "dark" por defecto, el toggle de tema existente lo overridea igual
que hoy). No se reescribe la convención del pipeline, se cambia el default runtime — cirugía mucho más
chica y de menor riesgo que invertir qué mode vive en `$value`.

## D3 · Proceso: paquete formal completo, gate humano antes de tocar producción

Esto reemplaza VALORES de una base con firma explícita: `arnesia-shell-A-galaxia.html` está marcado
`🔒 firmado (tokens)` en `mockups/INDEX.md`, y es la fuente de valores citada por
`fe-tokens-contrato.md` L2. No es un superset (regla 3 de `mockups/INDEX.md`) en el sentido estricto de
"agregar sin quitar" — los VALORES cambian de raíz (sand→teal, light-default→dark-default). Por eso:
paquete propio, con su propia carpeta, y el `.tokens.json`/`theme.css`/mockup baseline NO se tocan hasta
que este paquete tenga su propio gate humano 🧑‍⚖️ (documentado en un `paridad.md` cuando llegue esa
etapa). Al firmar, `mockups/INDEX.md` se re-estampa (nueva fila o nota de re-firma sobre
`arnesia-shell-A-galaxia.html`) — mismo patrón que S1-D12 hizo con el mockup del Portafolio.

## D4 · Fuentes: sustitutas por licencia, self-host pendiente de decidir (hallazgo técnico)

La marca real usa Coco Gothic (Zetafonts, comercial) y Sansation (free, no está en Google Fonts);
Claude Design ya las sustituyó por Jost/Mulish/JetBrains Mono (Google Fonts, vía `@import` a
`fonts.googleapis.com`). **Pregunta abierta para `propuesta-tokens.md`:** ArnesIA es una app de
escritorio (Tauri) pensada para poder correr sin depender de una CDN externa en cada arranque — cargar
tipografía vía `@import` remoto en cada boot choca con eso. Alternativas: (a) self-hostear los `.woff2`
de Jost/Mulish/JetBrains Mono dentro de `web/public/fonts/` (build-time, cero red en runtime) — mismo
criterio que ya se sigue con el resto del bundle (`go:embed`, cero CDN); (b) mantener el `@import`
remoto como está en Claude Design, aceptando que la app pierde tipografía de marca offline (degrada a
`system-ui` del stack, que ya es el fallback declarado). Sin decidir aún — se resuelve en la etapa spec.

## D5 · El sistema PRENTER no toca el schema DTCG, solo puebla valores + agrega 1 token nuevo

Verificado contra el `.tokens.json` real: los 21 slots semánticos que PRENTER puede alimentar
(`background…ring` + `sidebar/*`) YA EXISTEN en el schema — esto es un cambio de VALORES, no de forma.
Dos gaps reales encontrados: (a) el schema no tiene un slot `font.display` separado de `font.sans` —
PRENTER trae display (Jost) y body (Mulish) como familias DISTINTAS, hoy `font.sans` es una sola pila
system-ui; se agrega `font.display` (nuevo token, aditivo, no rompe consumidores existentes de
`font.sans`); (b) `fe-tokens-contrato.md` L87 ya declaraba `--shadow-*` como **deuda honesta** (fuera
del SSOT DTCG, hardcodeado donde aparece) — PRENTER trae una escala `shadow` completa (`xs/sm/md/lg` +
`shadow-brand`), así que este rebrand de paso CIERRA esa deuda preexistente agregando `shadow` como
grupo DTCG real. Ambos son adiciones aditivas al schema (cero breaking change de consumidores actuales),
se detallan en `propuesta-tokens.md`.

## D6 · `destructive` en dark: mismo hex, verificado WCAG (respuesta operador: "mejor práctica")

Contraste `#d0483c` sobre `#000000` = **4.66:1** (fórmula de luminancia relativa WCAG 2.1, calculado a
mano) — pasa AA texto normal (≥4.5:1). No se inventa una variante que PRENTER no definió (a diferencia
de `primary`, donde SÍ hay teal-500/teal-400 explícitos). Detalle en `propuesta-tokens.md` pregunta 1.

## D7 · Self-host de fuentes confirmado — subset `latin` alcanza para español

`unicode-range: U+0000-00FF` (subset `latin` de Google Fonts) cubre Latin-1 Supplement completo →
á/é/í/ó/ú/ñ/ü/¿/¡ incluidos. NO hace falta `latin-ext` (centroeuropeo). 13 archivos `.woff2` descargados
a `web/public/fonts/`, `@font-face` propio — cero request a `fonts.googleapis.com` en runtime, la app
queda tipográficamente offline-capaz (consistente con el resto del bundle `go:embed`).

## D8 · `font.display` aplicado a headings; botones/CTA por-módulo QUEDA DEFERRED (deuda explícita)

Brandbook real (`Brandbook PRENTER.dc.html` §08, no solo `readme.md`) muestra `font-display` en
headings Y en botones/CTA en píldora. Esta pasada aplica **headings (h1-h4) vía regla global** —
cascada automática, cero riesgo, cero archivos por módulo que tocar. Los botones NO se tocan: la app no
tiene un primitivo `Button` compartido en uso real (`shared/ui/button.tsx` existe pero 0 componentes lo
importan — verificado por grep), cada módulo (Portafolio, Mapa, Ajustes) trae sus propias clases
`.pf-btn-*`/etc. Aplicar `font-display` a un selector `button` global sin ver cada superficie en vivo
(tabs, iconos, botones secundarios que el brandbook NO muestra en Jost) es un riesgo visual real que no
se puede verificar sin QA por módulo. → **Deuda registrada en BACKLOG**, no se adivina a ciegas.

## D9 · 2 valores PRENTER corregidos por WCAG real — cazado por el suite de a11y, no a mano

Al aplicar la propuesta y correr `vitest --project=storybook` (127 stories, addon-a11y activo), **24
tests fallaron** por `color-contrast` — evidencia real, no hipotética. Axe midió exacto:

1. **`neutral-500` (#6b7676) como `muted-foreground` sobre `neutral-100`/`neutral-50`** → 4.12:1 /
   4.39:1 (falla AA 4.5:1). Es el mismo par que D1/propuesta asumía seguro por analogía con el `#5d6a79`
   anterior (que sí pasaba, 4.69:1) — PRENTER define un neutral-500 más CLARO que el que reemplazó, y
   eso rompe el contraste en chips/labels pequeños (9-14px) del Portafolio/Mapa. → oscurecido a
   **`#4d5555`** (mismo matiz gris-frío H≈180° S≈5%, interpolado entre los pasos 500/700 de PRENTER
   que sí definen — PRENTER no tiene un 600) → **5.89:1** contra `neutral-100`. Único consumidor:
   `muted-foreground` + `sidebar-foreground` (mismo rol: texto subordinado sobre superficie clara).
2. **`danger` (#d0483c) como `destructive-foreground` (blanco) de fondo** → 4.49:1 (falla por 0.01 —
   D6 había verificado la pareja EQUIVOCADA: danger-como-texto-sobre-negro, no danger-como-fondo-con-
   texto-blanco, que es como realmente se usa vía `bg-destructive text-destructive-foreground`).
   Oscurecido a **`#bd3a2e`** (mismo matiz H≈5°, L −10pp) → **5.51:1**.

Ambos re-verificados: `pnpm run verify` verde · `vitest --project=storybook` **127/127** · `--project=
unit` **21/21**. Lección para el gate: la "mejor práctica" de D6 (no inventar sin necesidad) sigue en
pie — acá SÍ hacía falta, y el suite real lo demostró antes de llegar a producción, no un cálculo manual
aislado.
