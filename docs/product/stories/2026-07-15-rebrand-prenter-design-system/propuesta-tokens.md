# Propuesta de tokens — PRENTER → `web/tokens/base.tokens.json`

> `tipo: propuesta` (spec-en-borrador) · paquete `2026-07-15-rebrand-prenter-design-system` · **NADA de
> esto está aplicado todavía.** Es la traducción exacta de los valores extraídos de Claude Design
> (`tokens/{colors,typography,spacing,fonts}.css` del proyecto «PRENTER Design System») a los slots
> semánticos que YA existen en `base.tokens.json` (D5: cero cambio de forma, solo de valores + 2
> adiciones aditivas). `color.kind`/`health`/`heat` NO aparecen acá — quedan intactos (D1).
>
> **Gate humano PENDIENTE.** Revisar la tabla, responder las 3 preguntas abiertas al final, y recién
> ahí esto pasa a `spec.md` (edición real del `.tokens.json` + regeneración de `theme.css` + re-derivar
> `arnesia-shell-A-galaxia.html`).

## Semántico — `color.semantic` (chrome estructural, D2: dark = default runtime)

| slot | hoy (light, sand) | **PRENTER light** (alterno) | hoy (dark) | **PRENTER dark** (default) |
|---|---|---|---|---|
| `background` | `#f2f4f7` | `neutral-50` `#f6f8f8` | `#14171c` | `dark-bg` `#000000` |
| `foreground` | `#212832` | `text-strong`(neutral-900) `#0e1312` | `#e8ecf1` | `text-on-dark` `rgba(255,255,255,.92)` |
| `card` | `#ffffff` | `surface-card` `#ffffff` (sin cambio) | `#1b2027` | `dark-surface` `#0c1110` |
| `card-foreground` | `#212832` | `text-strong` | `#e8ecf1` | `text-on-dark` |
| `popover` | `#ffffff` | `#ffffff` (sin cambio) | `#1b2027` | `dark-raised` `#141a19` |
| `popover-foreground` | `#212832` | `text-strong` | `#e8ecf1` | `text-on-dark` |
| `primary` | `#a8742c` (sand-600) | `teal-500` `#00b7aa` | `#d9a35b` (sand-400) | `teal-400` `#1fc6b8` (más claro sobre negro, mismo patrón sand-600→400 que ya usa el schema) |
| `primary-foreground` | `#ffffff` | `brand-contrast` `#000000` (PRENTER lo pide así: teal es claro, negro lee mejor que blanco encima) | `#ffffff` | `#000000` (mismo motivo) |
| `secondary` | `#e9edf2` | `neutral-100` `#eef1f1` | `#242b34` | `dark-raised` `#141a19` |
| `secondary-foreground` | `#212832` | `text-strong` | `#e8ecf1` | `text-on-dark` |
| `muted` | `#e9edf2` | `neutral-100` | `#242b34` | `dark-raised` |
| `muted-foreground` | `#5d6a79` | `text-muted`(neutral-500) `#6b7676` | `#94a1af` | `text-on-dark-muted` `rgba(255,255,255,.58)` |
| `accent` | `#e9edf2` | `neutral-100` | `#242b34` | `dark-raised` |
| `accent-foreground` | `#212832` | `text-strong` | `#e8ecf1` | `text-on-dark` |
| `accent-soft` | `rgba(168,116,44,.14)` | `rgba(0,183,170,.14)` (teal, mismo patrón de opacidad) | `rgba(217,163,91,.14)` | `rgba(0,183,170,.14)` (misma, PRENTER no da 2 tonos de teal-soft) |
| `destructive` | `#c94545` (red-600) | `danger` `#d0483c` (PRENTER) | `#e25b5b` (red-400) | `#d0483c` — ⚠ **pregunta abierta #1**: PRENTER no da un `danger` distinto para dark |
| `destructive-foreground` | `#ffffff` | `#ffffff` (sin cambio) | `#ffffff` | `#ffffff` (sin cambio) |
| `border` | `#d5dbe3` | `border-subtle`(neutral-200) `#e1e6e6` | `#2d3743` | `dark-border` `#1f2826` |
| `input` | `#c3cbd6` | `border-default`(neutral-300) `#c9d1d1` | `#3a4655` | `dark-border-teal` `rgba(0,183,170,.35)` (PRENTER lo define para exactamente este uso — borde con más presencia que `border` en superficies dark) |
| `ring` | `#a8742c` | `focus-ring`(teal-400) `#1fc6b8` | `#d9a35b` | `focus-ring` `#1fc6b8` (mismo valor ambos modos, PRENTER solo define un `--focus-ring`) |

`sidebar/*` (Command Rail) — mismo criterio, espeja `card`/`border`/`primary` de cada modo (no
duplico la tabla: `sidebar-background=card`, `sidebar-primary=primary`, `sidebar-border=border`, etc.,
igual que hoy).

## Tipografía — `font` (D5: agrega `font.display`, `font.sans`/`font.mono` cambian de pila)

| slot | hoy | PRENTER |
|---|---|---|
| `font.sans` (cuerpo) | `system-ui, -apple-system, Segoe UI, Roboto, sans-serif` | `Mulish, system-ui, sans-serif` (sustituto de Sansation) |
| `font.display` **(nuevo)** | — no existe | `Jost, system-ui, sans-serif` (sustituto de Coco Gothic) — para títulos/headings, no todo el body |
| `font.mono` | `ui-monospace, SFMono-Regular, Menlo, Consolas, Liberation Mono, monospace` | `JetBrains Mono, ui-monospace, SF Mono, Menlo, monospace` |

## Radios — `radius` (valores SÍ cambian, no solo chrome)

| slot | hoy | PRENTER |
|---|---|---|
| `sm` | `6px` | `4px` |
| `md` | `8px` | `8px` (sin cambio) |
| `lg` | `9px` | `14px` |
| `xl` | `12px` | `22px` |
| `full` | `9999px` | `999px` (equivalente visual, ajusto a `9999px` para no perder precisión de pill) |

## Sombras — `shadow` **(grupo nuevo, D5: cierra deuda declarada en `fe-tokens-contrato.md` L87)**

| slot | valor PRENTER |
|---|---|
| `shadow.xs` | `0 1px 2px rgba(11,15,14,.05)` |
| `shadow.sm` | `0 1px 3px rgba(11,15,14,.07), 0 1px 2px rgba(11,15,14,.04)` |
| `shadow.md` | `0 4px 14px rgba(11,15,14,.08), 0 2px 4px rgba(11,15,14,.04)` |
| `shadow.lg` | `0 14px 40px rgba(11,15,14,.10), 0 4px 10px rgba(11,15,14,.05)` |
| `shadow.brand` | `0 8px 24px rgba(0,183,170,.22)` (para el CTA/acento principal) |

## Espaciado — `space` (sin cambio propuesto)

La escala 4px-base de PRENTER (`4·8·12·16·24·32·48·64·96·128`) es un superset de la actual
(`4·8·12·16·20·24·32·40·48`) salvo por los pasos `20`/`40` que hoy existen y PRENTER no define. **No
se tocan** — se mantiene la escala actual completa, PRENTER no obliga a nada acá (la app ya cumple
"4px base grid").

## Preguntas — RESUELTAS (respuesta del operador, 2026-07-15)

1. **`destructive` en dark → mismo hex `#d0483c`, verificado con WCAG, no inventado.** El operador pidió
   "mejor práctica entre devs": la práctica estándar (Radix/Material/GitHub Primer) es generar variante
   solo si la existente FALLA contraste — no inventar un tono que la marca no definió. Calculado
   (fórmula de luminancia relativa WCAG 2.1): `#d0483c` sobre `#000000` (el fondo más oscuro posible de
   la superficie) da **contraste 4.66:1** — pasa AA texto normal (≥4.5:1) con margen. Sobre `--dark-
   surface`/`--dark-raised` (más claros que negro puro) el margen crece. No hace falta clarear: se
   mantiene UN SOLO valor en ambos modos, honesto con lo que la marca realmente definió (a diferencia de
   `primary`, donde PRENTER SÍ da explícitamente teal-500/teal-400 — ahí se respeta esa intención,
   acá no se inventa una que no existe).
2. **Self-host, confirmado.** 13 archivos `.woff2` (Jost 300/400/500/600/700 · Mulish 300/400/500/600/
   700 · JetBrains Mono 400/500/700), subset **`latin`** únicamente (`unicode-range: U+0000-00FF` —
   cubre Latin-1 Supplement completo: á/é/í/ó/ú/ñ/ü/¿/¡, todo el español; `latin-ext` es para
   centroeuropeo, no hace falta). Van a `web/public/fonts/{jost,mulish,jetbrains-mono}/`, `@font-face`
   propio en `fonts.css` — cero request a `fonts.googleapis.com` en runtime.
3. **`font.display` (Jost) — alcance real verificado contra el brandbook vivo** (`Brandbook
   PRENTER.dc.html`, sección 08 «Lenguaje digital», no solo el resumen de `readme.md`): el markup real
   usa `font-display` en headings (h1-h4) Y TAMBIÉN en los botones/CTA en píldora (`"Cotizar proyecto"`,
   `"Ver el producto →"`, ambos con `font-family: var(--font-display); font-weight: 600`) — mi lectura
   inicial (solo headings) estaba INCOMPLETA, corregida con evidencia del propio archivo. **Aplicado en
   esta pasada: headings (h1-h4) vía regla global** (`font-display` cascada automática, cero riesgo).
   **DEFERRED: botones/CTA por-módulo** — la app NO tiene un primitivo `Button` central en uso real (0
   componentes lo importan; cada módulo tiene sus propias clases `.pf-btn-*`/`.map-btn-*` etc, grep
   verificado) — aplicarlo a ciegas con un selector `button` global arriesga tabs/iconos/botones
   secundarios que el brandbook NO muestra en Jost. Queda como deuda explícita en `decisiones.md` D8,
   no aplicado sin QA visual por módulo.
