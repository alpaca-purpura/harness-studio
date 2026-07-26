# Design — Identidad de build en la tarjeta de Ajustes (RF-231)

> Paquete `2026-07-26-identidad-de-build`. **El UI al pixel.** Escrito en el refinamiento
> (2026-07-26, 2ª pasada): el paquete se construyó sin este archivo y §10 lo pide.
> **Es descriptivo de lo construido**, no una propuesta nueva — cada valor sale de leer
> `web/src/app/styles/ajustes.css` y `update-card.tsx`, no de inventar.
> SSoT del UI = Storybook (`update-card.stories.tsx`). Snapshot derivado:
> `stories/2026-07-07-boton-actualizar/mockup-actualizar.html`.

## 1. Dónde vive

Vista **Ajustes** → tarjeta `<section class="uc-card" aria-label="Versión y actualización">`,
encabezado `Versión y actualización`. Ancho 460 px, `--card` sobre `--border`, `--radius-md`,
padding 16, `gap: 14`. **La tarjeta no cambia de tamaño, de sitio ni de encabezado** — RF-231 solo
reescribe las dos primeras filas y agrega una caja condicional.

## 2. Orden de las filas (lo que cambió)

| # | Antes (RF-107) | Ahora (RF-231) | Por qué |
|---|---|---|---|
| 1 | `daemon · arnesia · huella 1c7443f · 2026-07-07` | `daemon · **arnesia v0.2.21.2607260225**` | El build es lo que contesta «¿corro lo último que compilé?» (B-D3) |
| 2 | — | `compilado · 2026-07-26 02:25 · commit 1c7443f · 2026-07-07` | El commit no desaparece: baja un escalón (B-D3) |
| 3 | — | *(condicional)* caja `warn` con el aviso | Superficie, no `title` (B-D4) |
| 4 | `instalado en …` + pill | **sin cambios** | — |
| 5 | `origen …` | **sin cambios** | — |
| 6 | botón **Actualizar** + checklist de pasos | **sin cambios** | — |

Superset estricto: no se quitó ningún dato firmado en PARIDAD del paquete `boton-actualizar`
(huella, fecha, `instalado en`, pill de escribibilidad, origen, botón, checklist).

## 3. Tabla de campos

| Campo | Fuente (wire) | Tipografía | Color | Ausente cuando |
|---|---|---|---|---|
| identificador del build | `version.version` | `--font-mono`, `<b>` | `--foreground` | nunca (cae a texto `dev`) |
| «build sin sellar (dev)» | `version.version` vacío o `"dev"` | `--font-sans` (`.uc-mut`) | `--muted-foreground` | hay sello |
| fecha de compilación | `version.compilado` | `--font-mono` | `--foreground` | build sin sellar → la fila muta a `commit` |
| etiqueta `commit` | literal | `--font-sans` (`.uc-mut`) | `--muted-foreground` | `huella === "dev"` |
| huella | `version.huella` | `--font-mono` | `--foreground` | — |
| fecha del commit | `version.fecha` | `--font-mono` | `--foreground` | el build no la trae |
| aviso de build viejo | `version.aviso_build` | `--text-xs` | `--foreground` sobre `--warn-soft` | está al día (`omitempty`) |

Llaves (`.uc-k`): 110 px, `--muted-foreground`, `--text-xs`. Valores (`.uc-v`): `--font-mono`,
`overflow-wrap: anywhere` (un identificador de 18 caracteres nunca desborda la tarjeta).

## 4. Estados (los 3 que hay, todos con story)

| Estado | Fila 1 | Fila 2 | Caja warn | Story |
|---|---|---|---|---|
| **sellado y al día** | `arnesia v0.2.21.2607260225` en negrita | `compilado 2026-07-26 02:25 · commit 1c7443f · 2026-07-07` | ausente | `IdentidadDelBuild` |
| **sellado, hay uno más nuevo** | idem | idem | `▲ el binario instalado es más nuevo (…) — cerrá y reabrí la app` · o · `▲ hay un build más nuevo sin instalar (… en …/bin/arnesia) — corré \`make dev-sync\`` | `BuildViejoCorriendo` |
| **sin sellar (dev)** | `arnesia · build sin sellar (dev)` | fila `commit` (o `versión no embebida (dev)` si tampoco hay huella) | ausente **siempre** | `BuildSinSellar` |

Los estados heredados de RF-104 (`leyendo /api/version…`, `uc-critbox` de error, actualizando,
reiniciando, éxito, ya-al-día) quedan **intactos** y siguen cubiertos por sus 12 stories previas.

## 5. Tokens (DTCG reales, cero valor inventado)

Todos salen de `web/src/app/styles/theme.css` (PRENTER, dark-first) vía `ajustes.css`:
`--card` · `--border` · `--foreground` · `--muted-foreground` · `--warn-soft` · `--radius-md` ·
`--text-xs` · `--font-mono` · `--font-sans`.

**La caja del aviso reusa `.uc-warnbox`**, la misma clase que ya usa la advertencia de
`no escribible` del paquete `boton-actualizar`: `--warn-soft` de fondo con texto `--foreground`
(nunca texto `--warn` — es justo el par que da 3.76:1 y rompe la gate de a11y, deuda abierta en
BACKLOG). **Cero CSS nuevo en este RF.**

## 6. Microcopy (literal, es contrato — las stories lo assertan)

- `arnesia v{version}` — la `v` la pone el FE, el wire manda el número pelado.
- `build sin sellar (dev)`
- `▲ el binario instalado es más nuevo (2026-07-26 12:07) que el que está corriendo — cerrá y reabrí la app`
- `` ▲ hay un build más nuevo sin instalar (2026-07-26 11:06 en /home/…/bin/arnesia) — corré `make dev-sync` ``

Regla: **el aviso termina en la acción**, no en el diagnóstico (B-D4/RN-6).

## 7. A11y

- El aviso es un `<p>` dentro de la `<section aria-label="Versión y actualización">`: se lee en
  orden, sin requerir foco.
- Contraste: `--foreground` sobre `--warn-soft` (el par ya validado por la gate de axe de las
  stories de esta tarjeta — las 15 pasan).
- **Deuda declarada:** el aviso no tiene `role="status"`/`aria-live`. Aparece en el primer render de
  la tarjeta, no como cambio dinámico, así que un live-region anunciaría de más. Si alguna vez se
  refetchea en vivo (ver «límites» del spec), hay que revisarlo.
