# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-23** (Portafolio Slice 0 «Cimientos» — gate humano 🧑‍⚖️ FIRMADO 2026-07-13,
8 puntos del goal + 5 desviaciones aceptadas). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Programa «Portafolio · ciclo de vida del arnés» — modelo FIRMADO 🧑‍⚖️ (HS-22); Slice 0
  CONSTRUIDO + FIRMADO 🧑‍⚖️ (HS-23); Slice 1 «FE» CONSTRUIDO, gate humano 🧑‍⚖️ PENDIENTE (2026-07-14).**
  El spike de CARGA se reencuadró en el front-door del ciclo de vida (agregar de marketplace/proyecto · observar · mejorar ·
  publicar · actualizar · reparar). Modelo firmado: identidad **`(home,id)`** · N:M:M · **canónico** (editable) + N **instalaciones**
  (read-only) · ley anti-drift descriptiva · **`deriva`**. Revisión adversaria (4 subagentes) probó que el FE no es construible sin
  cimientos → recut **Slice 0 «Cimientos» (dominio) → Slice 1 «FE»** (épica en `BACKLOG.md`). Paquete
  [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md) (specs + casuística + revisión + mockup v2).
  **Hand-off:** Fable 5 (arquitecto) planificó → Sonnet 5 (constructor) ejecutó Slice 0 T1-T9 y Slice 1
  T1-T8 — los 8 tickets YA en `main` (T8 = `b35ff6c`, commiteado por el operador). **Auditoría post-build
  de la carga (Fable 5, 2026-07-14) EJECUTADA:** build sano; CI-rojo por lint preexistente REPARADO,
  motor conformance extendido a tests colocados (pass 42→46), docs stale sincronizadas →
  [`stories/2026-07-14-auditoria-portafolio-carga/`](stories/2026-07-14-auditoria-portafolio-carga/informe.md).
  Ver «Retomar aquí».
- **Fase 1 «forja-ciclo-vivo» — PAUSADA** ([`stories/2026-07-10-forja-ciclo-vivo/`](stories/2026-07-10-forja-ciclo-vivo/INDEX.md)):
  Slice 1a (golden scaffold `docs/terreno/` dim piloto `forma-trabajo` + `docs/wip/`) **construido, pendiente firma
  🧑‍⚖️**; verificado verde (F-D4). **Aclaración F-D5:** la expertise de forja se encarna en el **kit ② embebido**
  (cuerpos ①②, HS-10) — se inyecta al chat in-app; NO es `docs/terreno/` top-level (eso es un ③ dogfood mal-ubicado,
  migrará, no revertir). Firmado «docs/terreno absorbe architecture» → bajo revisión (P9 en terreno-conocimiento).
- **Modelo de TERRENO FIRMADO 🧑‍⚖️ 2026-07-10** ([`stories/2026-07-10-terreno-conocimiento/`](stories/2026-07-10-terreno-conocimiento/INDEX.md):
  D0-D20 + `arnes.yaml` + `estructura-terreno.html` v4). Sigue válido como **schema que el forjador aplica**.
- **Retomar aquí:** **Slice 1 «FE Portafolio» CONSTRUIDO (T1-T8), gate humano 🧑‍⚖️ PENDIENTE
  (2026-07-14)** en
  [`stories/2026-07-13-portafolio-slice1-fe/`](stories/2026-07-13-portafolio-slice1-fe/INDEX.md):
  3 superficies (Lista/lente-empresa · Wizard-Proyecto/carpeta-local · Drawer READ) cableadas al HTTP
  real de Slice 0, «Abrir en Mapa» (observación read-only, cierra GAP-1) + Desvincular con
  confirmación, los 9 fixes G1-G9 de `revision-adversaria.md` aplicados y evidenciados (34 stories
  `play()` + `selectors.test.ts`, Storybook = SSoT), mockup `arnesia-portafolio.html` corregido en
  sitio + `mockups/INDEX.md` re-estampado, 4 capabilities `fe-portafolio/*` graduadas a `vivo` (R4).
  **E2E vivo contra la máquina real** (browser+curl, daemon aislado sin tocar el de producción):
  escaneo real de `luana-vitalia` → candidatos reales (`harness@0.5.2 en-deriva`) → agregar → Lista
  con grupo «sin empresa» real → Drawer con trazabilidad de eslabones real → **Abrir en Mapa de una
  instalación `referenciada-cc` real confirma GAP-1 cerrado** (grafo real, CERO cwd registrado) →
  corrupción a mano + reboot → banner honesto → Desvincular real + 404 en repetición → recorrido de
  teclado sin mouse. Ver evidencia completa en
  [`paridad.md`](stories/2026-07-13-portafolio-slice1-fe/paridad.md) — **firma 🧑‍⚖️ del operador
  PENDIENTE** (checkbox sin marcar, no simulada). Slice 0 (`stories/2026-07-13-portafolio-slice0-
  cimientos/`) sigue FIRMADO (`ledger/HS-23.md`), sin cambios.
  Pausadas: **Fase 1 forja** (Slice 1a pendiente firma) · **terreno** (schema firmado). Contexto: memorias
  `hs-forja-fase1-carga-spike` · `hs-terreno-modelo-firmado` · `hs-linea-base-ui-storybook-ssot` · `hs-chat-cc-funcional`.
- **Nuevo (2026-07-15): Rebrand sistema de diseño PRENTER — IMPLEMENTADO, gate humano PENDIENTE.**
  `docs/product/stories/2026-07-15-rebrand-prenter-design-system/`: decisiones D1-D9 (paleta funcional
  del Mapa intacta · dark-first default · fuentes self-hosted · 2 correcciones WCAG reales cazadas por
  el suite de a11y, D9). Aplicado a `web/`: tokens+`theme.css`+`index.css`+`app-store.ts`+3 `.woff2`.
  Verificado en vivo: `pnpm run verify` + `vitest` 127/127+21/21 + screenshot real (Storybook `:6006`).
  Falta: el operador corre la app y firma antes de re-derivar `arnesia-shell-A-galaxia.html`
  (🔒 firmado, HS-03/HS-05) y cerrar el paquete (`paridad.md` sin escribir aún).
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `253 checks · pass 46 · fail 0 · error 0 · deferred 207 · n/a 0` (medido 2026-07-14, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-14
- **arch/:** 18 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 92 — 49 vivo · 39 vivo·nc · 1 parcial · 3 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
