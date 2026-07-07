# design — drawer/inspector del Mapa (UI al pixel)

> Compañero de `spec.md` (mismo corte: mockup v6). Todos los valores = tokens DTCG de
> `web/tokens/base.tokens.json` → `theme.css` (`var(--…)`), disciplina stylelint
> anti-magic-value. Nada se inventa: lo que no está aquí, se lee del mockup v6.

## Anatomía (drawer normal, 340px)

```
┌──────────────────────────────────────┐
│ [glyph] nombre (mono sm 600)   ⤢  ✕ │  header · p-16 · gap-10 · border-b
│         Clase · handle (xs muted)    │
├──────────────────────────────────────┤
│  Resumen   Contenido   Corridas      │  tabs · flex-1 c/u · underline 2px --primary
├──────────────────────────────────────┤
│ pane activo (p-16, col, gap-16)      │
└──────────────────────────────────────┘
```

- Ancho fijo `340px`, `bg-card`, `border-border`, dockeado a la derecha del canvas.
- Glyph 17px (forma+char+color por clase = canal redundante, `KIND`); texto del glyph
  = `--background`.
- Botones header 24×24, `radius-md`, borde `--border`, hover→`--foreground`.

## Modo expandido (decisiones #1/#5)

- Cubre el área del mapa (`inset:0` del contenedor del canvas), header sticky top:0,
  **tabs sticky top:57px** (alto real del header), mismas capas `bg-card`.
- Tabs pasan a `flex:0 0 auto; padding:8px 16px`, alineadas a la izquierda.
- Resumen: rejilla `repeat(auto-fit, minmax(280px,1fr))`, `gap:20px 36px`,
  `max-width:1280px`, `padding:20px 24px`; **botonera `grid-column:1/-1`** al pie con
  `border-top --border`, botones en fila (`flex:0 0 auto`) + nota al lado.
- Contenido/Corridas: `max-width:920px` (columna de lectura), alineada a la izquierda.
- Semántica: ⤢→⤡ conmuta (aria-pressed); ✕ CIERRA (drawer desaparece, mapa queda);
  Esc colapsa. Tooltips nativos en ambos.

## Tipografía y jerarquía

| Rol | Token | Uso |
|---|---|---|
| nombre del nodo | `--text-sm` mono 600 | header |
| handle/clase | `--text-xs` muted; handle mono | header línea 2 |
| título de sección | `--text-xs` 600 uppercase tracking .05em muted | h4 + «i» |
| campo k/v | `--text-xs`; k muted (punteado si tooltip), v mono foreground | Field |
| prosa (why, roles) | `--text-xs` foreground/muted | Section p |
| viewer de fuente | mono 10.5px lh 1.55 | tab Contenido |
| notas staged/pendiente | 10px muted | act-note / src-note |

## Marcas y componentes nuevos

- **«i» de sección**: círculo 13px, borde+texto `--muted-foreground`, char mono `i`,
  `margin-left:2px`, cursor help, `tabindex=0`.
- **Campo con tooltip**: `border-bottom:1px dotted var(--muted-foreground)` + cursor
  help en la K (jamás en el valor).
- **Tooltip**: caja `bg-card` borde `--border` radius-md, padding 7×9, `--text-xs`
  sans lh 1.45, sombra suave, `max-width:250px`, arriba-izquierda del ancla; visible
  en `:hover` y `:focus-visible`; transición opacity .12s solo sin reduced-motion.
- **Chip (necesita/ruta/viene-de)**: borde `--border` radius-sm, padding 2×7, mono
  `--text-xs`; origen dentro en 10px muted (`← tipo: id`); navegable → hover borde
  `--muted-foreground`; inerte → `opacity:.85` + title.
- **Badge condición de ruta**: `border:1px dashed var(--warn)`, texto `--warn` 10px
  mono, pill 999px, `si <condición>`.
- **Hallazgo**: caja radius-md padding 7×9, texto `--foreground`; fondo `--crit-soft`
  (gate:none) o `--warn-soft` (no-reconocido). NUNCA texto warn/crit a 11px sobre
  fondo plano (contraste AA — lección v3→v5).
- **Botonera staged**: botones full-width (columna) en normal; `--accent-soft` +
  borde `--primary` la primaria; disabled `opacity:.68` + `cursor:not-allowed` +
  `title` con la fase que los cablea.
- **Chip «versiona con el arnés»**: pill 9px 700, borde+texto `--primary`, fondo
  `--accent-soft`.
- **Viewer de fuente**: fondo `--muted` radius-md, filas `nº | código` (nº muted .6
  min-width 2ch right), `overflow-x:auto`.

## Tabla clase → contenido del drawer (consolidada)

| clase | handle | Resumen añade | Contenido | Corridas |
|---|---|---|---|---|
| skill caja | `/id` | contrato completo + Viene de + Hallazgos | SKILL.md real | honesto (+nota D2/sesión) |
| skill apoyo | `/id` | rol + pendiente reconocedor | SKILL.md real | honesto |
| subagent | `@id` | contract si lo trae; rol si no | agents/<id>.md | honesto |
| command | `id` | rol + pendiente | commands/<id>.md | honesto |
| hook | `evento` | rol Guardia + pendiente | hooks.json (fragmento) | honesto |
| rule | always-on/condicional | Activación (alw) + rol Base | CLAUDE.md | honesto |
| mcp | `id` | rol + pendiente | .mcp.json | honesto |
| plugin/settings/output-style/statusline | `id` | rol + pendiente | su celda canónica | honesto |
| no-reconocido | `id` | Reconciliación warn + hallazgo D-c | RAW del artefacto | honesto |
| (vacío) | — | línea de affordance | — | — |

## Estados

- **Con dato** → se dibuja; **sin dato** → se DICE (pendiente del reconocedor /
  sin corridas indexadas / fuente no disponible) — gris ≠ verde, jamás blanco mudo.
- **PROPUESTA** (origen/alw de fixture) → sufijo · nota en tooltip.
- **staged** → disabled + rotulado con la fase que lo cablea.

## A11y

Focus visible en todo interactivo · tooltips operables por teclado · `aria-pressed`
(⤢) / `aria-selected` (tabs) / `aria-label` (inspector, cerrar, ampliar) · contraste
AA en ambos temas (light/dark de theme.css) · reduced-motion respetado.
