# design — Shell: Topbar sin empresa + selector de arnés (UI al pixel)

> Compañero de `spec.md` (mismo corte: mockup firmado 2026-07-20, commit `5ca97c2`). Todos los
> valores = tokens DTCG de `web/tokens/base.tokens.json` → `theme.css` (`var(--…)`), disciplina
> stylelint anti-magic-value. Nada se inventa: lo que no está aquí, se lee del mockup.

## 1 · Topbar

```
┌──────────────────────────────────────────────────────────┐
│ ┌dev-full-cycle┐ / Mapa                                   │  fila 1: breadcrumb (sin empresa)
├──────────────────────────────────────────────────────────┤
│                                          ✦ Conversar ⌘K   │  fila 2: acción, alineada derecha
└──────────────────────────────────────────────────────────┘
```

- Contenedor pasa de 1 fila (`min-h-[44px]`, `items-center`) a 2 filas (`flex-col`, `gap-1.5` aprox
  — mismo espaciado vertical que el mockup `.topbar.row2` `gap:7px`).
- **Chip de arnés (etiqueta plana):** `border` `--border` **dashed** (no sólido — distingue
  visualmente de un control real), `background: none` (no `--secondary` — un chip con fondo relleno
  lee como botón), `radius-md`, `padding: 2px 8px`, mono `11.5px`. Sin `▾`, sin `onClick`, sin
  `hover` state (no es interactivo).
- **Botón Conversar:** sin cambios de estilo propio (`--primary` border, `--accent-soft` bg) — solo
  cambia su posición: `self-end` en vez de `ml-auto` dentro de la misma fila.
- Breadcrumb: `arnés / vista` únicamente. Si `multi` (2+ frentes del mismo arnés), el sufijo
  `· frente «…»` que ya existe se mantiene sin cambios (no está en el alcance de las decisiones).

## 2 · Rail — picker de nueva sesión

### Anatomía (rail ensanchado a 360px)

```
┌────────────────────────────────────┐
│ A  ArnesIA                      «  │  header (sin cambios)
├────────────────────────────────────┤
│ Elegí el arnés de esta sesión      │  picker-head · 13px 700
│ No se cambia después…              │  picker-sub · 11px muted
│ 🔍 Buscar por id, rol o empresa…   │  input, mismos tokens que .pf-buscar
├────────────────────────────────────┤
│ ○ D  dev-full-cycle   Desarrollo…  │  pf-row (caso simple)
│      [alpacapurpura] ◆ canónico    │
│                                  ●  │  ← DotSaludPortafolio
│ ○ L  luana-feature-cycle …         │  pf-row.ambiguous
│      [luana][alpacapurpura] ▣2 [en-deriva]
│   └─ ○ proyecto-instalado  ~/proy… │  copy-option (visible solo si .selected)
│      ○ referenciada-cc     ~/.cla… │
├────────────────────────────────────┤
│ usará ~/proyectos/harness-studio   │  picker-resolved (solo si hay path)
│                    Cancelar  Crear │  picker-actions
└────────────────────────────────────┘
```

- Ancho fijo `360px` (no fluido — mismo valor que `ChatDock`, TS-D10), `flex-col`, `gap-2.5`
  aproximado (mockup `gap:10px`).
- El `<ul>` de filas tiene `max-height` con scroll propio (no crece sin límite con un portafolio
  grande) — `overflow-y:auto`, `max-height` ~ el alto disponible del rail menos header/buscador/
  acciones (usar `flex-1 min-h-0 overflow-y-auto`, no un px mágico).

### Tabla de campos por fila (RF-8)

| Campo visual | Fuente (wire) | Componente a reusar |
|---|---|---|
| Emblema (inicial+color) | `identidad.id` | `EmblemaInicial` (`entities/portafolio/ui/chips.tsx`) |
| id | `identificadorDe(identidad)` | selector puro, ya existe |
| nombre | `entrada.nombre` (opcional) | — |
| chips de empresa | `entrada.empresas[]` — TODAS | chip simple `pf-chip`-like (Tailwind, no CSS nuevo) |
| presencia | `canonico ? "◆ canónico vX.X" : ""` + `"▣ N instalac."` si `instalaciones.length>0` | — |
| deriva | instalación con `deriva==="en-deriva"` | `DerivaChip` |
| colisión | `idsColisionados` marca la fila | chip nuevo: `identidad.home ?? identidad.scope` truncado |
| salud | `saludDe(entrada)` | `DotSaludPortafolio` |
| radio | estado de selección local del picker | nuevo (no existe en entities/portafolio — es propio de ESTE flujo de selección, no de la Lista que usa click simple) |

Ningún campo se fabrica: sin `empresas` → sin chips (no "sin empresa" inventado); sin `canonico` ni
`instalaciones` → fila deshabilitada (TS-D15).

### Sub-lista de copias (caso ambiguo, RF-10)

- `border-top` `--border` dashed, `padding-left` alineado bajo el emblema (mismo indent que el
  mockup `padding: 6px 10px 8px 41px`).
- Cada `copy-option`: radio propio (independiente del radio de la fila) + `TipoInstalacionChip` (o
  literal `canónico` si es la copia canónica — no hay chip de tipo para eso hoy, texto mono simple)
  + path truncado (`overflow:hidden;text-overflow:ellipsis;white-space:nowrap`, `min-width:0` en
  TODA la cadena flex — el bug real cazado en la verificación del mockup, ver decisiones.md
  «Verificación del mockup») + `DerivaChip` si esa copia está en deriva.

### Estados (RF-13/14/15)

- **cargando** — skeleton compacto: 3 barras `pf-skeleton-fila`-like (mismo espíritu que
  `portafolio-list.tsx` `Skeleton`), reemplaza la lista, buscador queda deshabilitado.
- **error** — mensaje `role="alert"` + botón «Reintentar» (mismo copy que `ErrorBody`), buscador
  oculto (no hay nada que filtrar).
- **vacío** (`entradas.length === 0`, estado `datos`) — copy «Tu portafolio está vacío.» + botón
  «Ir a Portafolio» (TS-D11).
- **sin-resultados** (búsqueda no matchea nada, `entradas.length > 0`) — «Ningún arnés coincide con
  la búsqueda.» + botón «Limpiar búsqueda» (mockup `#pickerEmpty`).
- **resuelto** — línea `usará <path>` en mono, solo visible cuando hay `path` (simple o copia
  elegida).

### A11y

- Buscador con `aria-label="Buscar arnés"` (mockup ya lo tiene).
- Filas y copy-options operables por teclado (`role="radio"`/`radiogroup` o `button` + `aria-pressed`
  — elegir UNA convención y aplicarla consistente en ambos niveles).
- Foco inicial al abrir el picker: el input de búsqueda (mismo patrón que `search.focus()` del
  mockup, `mockup:587`).
- Estado deshabilitado (0 copias) con `aria-disabled` + `title` — nunca solo opacidad visual.
- Contraste AA en ambos temas (light/dark de `theme.css`), mismo criterio que el resto del Shell.

## 3 · Tokens (sin novedad — reuso total)

Ningún token nuevo: `--primary`, `--border`, `--input`, `--secondary`, `--accent-soft`,
`--muted-foreground`, `--warn`/`--warn-soft`, `--radius-{sm,md,lg}`, `--font-mono`. Los chips de
`entities/portafolio` (`DerivaChip`, `TipoInstalacionChip`, `DotSaludPortafolio`,
`EmblemaInicial`) se importan tal cual — cero reimplementación, cero drift visual contra la Lista
del Portafolio real.

## 4 · Fuera de este paquete (explícitamente)

- Estilo/contenido de `SessionCard` (sin cambios).
- `workspace-stage.tsx:235` (`{empresa} · {puesto}`) — efecto secundario conocido de TS-D16, no se
  toca acá.
- Cualquier CSS nuevo en `app/styles/*.css` — este picker es 100% Tailwind utilitario, mismo
  patrón que `topbar.tsx`/`session-rail.tsx` ya usan hoy (ninguno de los dos tiene hoja `.css`
  propia).
