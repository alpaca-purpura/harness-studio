# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-22** (spike CARGA → **Programa «Portafolio · ciclo de vida del arnés»**; modelo
FIRMADO 🧑‍⚖️ 2026-07-13 tras revisión adversaria de 4 subagentes). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Programa «Portafolio · ciclo de vida del arnés» — modelo FIRMADO 🧑‍⚖️ (HS-22, 2026-07-13), en cola de build.**
  El spike de CARGA se reencuadró en el front-door del ciclo de vida (agregar de marketplace/proyecto · observar · mejorar ·
  publicar · actualizar · reparar). Modelo firmado: identidad **`(home,id)`** · N:M:M · **canónico** (editable) + N **instalaciones**
  (read-only) · ley anti-drift descriptiva · **`deriva`**. Revisión adversaria (4 subagentes) probó que el FE no es construible sin
  cimientos → recut **Slice 0 «Cimientos» (dominio) → Slice 1 «FE»** (épica en `BACKLOG.md`). Paquete
  [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md) (specs + casuística + revisión + mockup v2).
  **Hand-off:** el operador divide tickets con **Fable 5** (arquitecto) → build con **Sonnet 5**. **Sin código aún.**
- **Fase 1 «forja-ciclo-vivo» — PAUSADA** ([`stories/2026-07-10-forja-ciclo-vivo/`](stories/2026-07-10-forja-ciclo-vivo/INDEX.md)):
  Slice 1a (golden scaffold `docs/terreno/` dim piloto `forma-trabajo` + `docs/wip/`) **construido, pendiente firma
  🧑‍⚖️**; verificado verde (F-D4). **Aclaración F-D5:** la expertise de forja se encarna en el **kit ② embebido**
  (cuerpos ①②, HS-10) — se inyecta al chat in-app; NO es `docs/terreno/` top-level (eso es un ③ dogfood mal-ubicado,
  migrará, no revertir). Firmado «docs/terreno absorbe architecture» → bajo revisión (P9 en terreno-conocimiento).
- **Modelo de TERRENO FIRMADO 🧑‍⚖️ 2026-07-10** ([`stories/2026-07-10-terreno-conocimiento/`](stories/2026-07-10-terreno-conocimiento/INDEX.md):
  D0-D20 + `arnes.yaml` + `estructura-terreno.html` v4). Sigue válido como **schema que el forjador aplica**.
- **Retomar aquí:** **plan de Slice 0 LISTO (Fable 5, 2026-07-13)** en
  [`stories/2026-07-13-portafolio-slice0-cimientos/`](stories/2026-07-13-portafolio-slice0-cimientos/INDEX.md):
  tickets T1-T9 + decisiones S0-D1..D11 + gate local + goal verificable. **Eslabón CC (F5) investigado con
  evidencia real** (CC registra procedencia en `~/.claude/plugins/{installed_plugins,known_marketplaces}.json` +
  `enabledPlugins`; instalación CC = referenciada al cache global). Próximo = **Sonnet 5 ejecuta T1-T9**
  (un commit por ticket, gates verdes); al final: `paridad.md` + gate humano del operador.
  Pausadas: **Fase 1 forja** (Slice 1a pendiente firma) · **terreno** (schema firmado). Contexto: memorias
  `hs-forja-fase1-carga-spike` · `hs-terreno-modelo-firmado` · `hs-linea-base-ui-storybook-ssot` · `hs-chat-cc-funcional`.
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `249 checks · pass 42 · fail 0 · error 0 · deferred 207 · n/a 0` (medido 2026-07-13, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-13
- **arch/:** 17 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 87 — 38 vivo · 40 vivo·nc · 1 parcial · 8 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
