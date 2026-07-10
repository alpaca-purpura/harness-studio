# mockups/ — INDEX · línea base del diseño (LÉEME ANTES DE DISEÑAR)

> vig: activo · dueño: cualquiera que **invoque creación/edición de diseño** (mockup, UI nueva,
> propuesta visual). **Leer esto ANTES de forkear cualquier `.html`.** Este registro existe para que
> una propuesta de diseño no vuelva a borrar/reinventar cosas ya construidas (incidente 2026-07-10).

## Regla dura (SSoT del UI)

1. **El SSoT del UI vigente es Storybook**, no estos `.html`. La superficie completa del Mapa vive en
   `web/src/widgets/map-canvas/ui/map-canvas.stories.tsx` («the full map surface — fitness fixture of
   record; the signed mockup is derived from what renders here») + `inspector.tsx` · `handoff-gutter.tsx`
   · `map-bar.tsx`. Se ve con `pnpm --dir web storybook`.
2. **Estos `.html` son SNAPSHOTS DERIVADOS y fechados** — pueden estar stale. Nunca forkear uno sin
   mirar su `estado` abajo. El baseline vigente se re-deriva de las stories.
3. **Toda propuesta de UI = superset estricto** del vigente. Jamás quitar un artefacto firmado (PARIDAD).
4. **No pisar vocabulario L0.** Antes de nombrar un concepto nuevo, verificar `internal/domain/box.go`
   + `web/src/entities/arnes/`. Ya tomados (con OTRO significado): `procedencia` (honestidad del dato:
   medido·estimado·declarado·inferido·no-declarado) · `origen` (estandar·del-puesto, lo estampa el
   provisioner) · `canal` (beta·estable·propuesto·deprecado) · `insumos` (inputs del contrato `Necesita`) ·
   `banda` (7 bandas). El drift/autoría por-archivo y el conocimiento-del-proyecto **necesitan nombre propio.**
5. **DoD:** todo paquete que toque UI **re-deriva** el baseline `.html` y actualiza su fila aquí
   (fecha + commit de sincronía). Ancla: la norma «pegarse al Storybook / tokens DTCG reales».

## Baseline VIGENTE

- **`arnesia-mapa-baseline.html`** — Mapa, superficie completa. **Derivado de las stories @`a01d845`
  (2026-07-10).** Incluye: barra (picker · META · toggle Artefactos off/auto/todos · conmutador de
  capas estructura/tokens·perf·proceso) · Guardia/Proceso/Base+bandas · edges invoca/lee/escribe ·
  **franja de artefactos** (chips hand-off + refs ↖) · **inspector 3 tabs** (Resumen·Contenido·Corridas)
  con contrato fusionado + botonera staged. **← Forkear DESDE AQUÍ para trabajar el Mapa.**

## Registro de mockups

| mockup | fecha (alta→últ) | rol | spec / paquete | estado |
|---|---|---|---|---|
| `arnesia-mapa-baseline.html` | 2026-07-10 | **BASELINE** del Mapa (superficie completa) | este INDEX + `map-canvas.stories.tsx` | **✅ vigente** |
| `arnesia-mapa-mvp.html` | 2026-07-06 | Mapa MVP Hito 1 (canvas bandas/carriles + SVG) | `stories/2026-07-06-mapa-mvp/{00-BRIEF,spec,design}.md` (Gate 1 ✓ `0736d2c`) | ⚠️ **superado** — le faltan franja-artefactos + inspector 3-tabs. NO forkear |
| `arnesia-mockup-v3.html` | 2026-07-04→05 | Shell + detalle (capas/inspector) it.13 | `ux.md` Inventario final §I · `stories/2026-07-07-inspector-drawer/analisis-drawer-v3.md` | 📎 referencia (detalle inspector) |
| `arnesia-shell-A-galaxia.html` | 2026-07-05 | Shell A «galaxia» it.13 — **fuente de VALORES de tokens** | `ux.md` Inventario final · `architecture/boundaries/fe-tokens-contrato.md` | 🔒 firmado (tokens) |
| `arnesia-shell-A-sessions.html` | 2026-07-05 | Multisesión it.14 | `stories/2026-07-08-chat-cc-funcional/INDEX.md` · `ux.md` | 📎 referencia (shell/sesiones) |
| `arnesia-shell-lab.html` | 2026-07-05 | Laboratorio de shell (4 paradigmas; A firmado) | `ux.md` Inventario final (it.13) | 🗄️ histórico (decisión de shell) |
| `arnesia-session-lab.html` | 2026-07-05 | Laboratorio de sesiones | — (exploración) | 🗄️ histórico |
| `arnesia-mockup-v2.html` | 2026-07-04 | Shell v2 (superado por it.13/14) | `ux.md` | 🗄️ histórico |
| `arnesia-arch-inyeccion-knowhow.html` | 2026-07-06→08 | Diagrama de **arquitectura** (inyección know-how HS-07/10) — no es UI de producto | `architecture/` · memoria HS-07/10 | 📎 referencia (arquitectura) |

> **Co-diseño en curso (2026-07-10):** «Mapa DESTINO» (sello · llenado/deriva · slots con receta ·
> conocimiento-del-proyecto) se está diseñando como **superset del baseline vigente**. Mientras no
> esté firmado NO reemplaza el baseline; vive como propuesta. Al firmar → se funde y se re-estampa aquí.
