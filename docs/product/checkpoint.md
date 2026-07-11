# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-21** (estado.sh → CI drift-gate + auto-cura `pre-commit`; eje N2
«honestidad automática», Paquete A — FIRMADA 2026-07-10). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **SPIKE de CARGA de arneses (2026-07-10) — ACTIVO.** El user quiere **usar la app y mejorar arneses YA**
  (trabajo retrasado). La app CORRE (`bin/arnesia serve` → `127.0.0.1:4200`; kit ①② inyectado desde `~/.arnesia`;
  3 arneses de ejercicio cargados) pero falta **agencia**: elegir desde la app qué arnés/proyecto cargar. Paquete
  [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md): backend de carga
  single-arnés HECHO (`PUT /api/arneses/{id}`+loader+índice); gaps = **dialog FE** (dif. HS-07) + **detección
  proyecto-multi** (lock `.devstudio/arneses.yaml`, dif. HS-12). Flujo pedido: revisar estado → ideas → mockup → build.
- **Fase 1 «forja-ciclo-vivo» — PAUSADA** ([`stories/2026-07-10-forja-ciclo-vivo/`](stories/2026-07-10-forja-ciclo-vivo/INDEX.md)):
  Slice 1a (golden scaffold `docs/terreno/` dim piloto `forma-trabajo` + `docs/wip/`) **construido, pendiente firma
  🧑‍⚖️**; verificado verde (F-D4). **Aclaración F-D5:** la expertise de forja se encarna en el **kit ② embebido**
  (cuerpos ①②, HS-10) — se inyecta al chat in-app; NO es `docs/terreno/` top-level (eso es un ③ dogfood mal-ubicado,
  migrará, no revertir). Firmado «docs/terreno absorbe architecture» → bajo revisión (P9 en terreno-conocimiento).
- **Modelo de TERRENO FIRMADO 🧑‍⚖️ 2026-07-10** ([`stories/2026-07-10-terreno-conocimiento/`](stories/2026-07-10-terreno-conocimiento/INDEX.md):
  D0-D20 + `arnes.yaml` + `estructura-terreno.html` v4). Sigue válido como **schema que el forjador aplica**.
- **Retomar aquí:** correr el **spike de carga** en conversación nueva (revisar → ideas → mockup story-first
  superset anti-drift → build). Al cerrar el spike → firmar Slice 1a → paso 2 forjador (reencuadrado por F-D5).
  Contexto: memorias `hs-forja-fase1-carga-spike` · `hs-terreno-modelo-firmado` · `hs-linea-base-ui-storybook-ssot` · `hs-chat-cc-funcional`.
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `249 checks · pass 42 · fail 0 · error 0 · deferred 207 · n/a 0` (medido 2026-07-09, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-09
- **arch/:** 17 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 82 — 38 vivo · 40 vivo·nc · 1 parcial · 3 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
