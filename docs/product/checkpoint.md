# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-21** (estado.sh → CI drift-gate + auto-cura `pre-commit`; eje N2
«honestidad automática», Paquete A — FIRMADA 2026-07-10). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Co-diseño ABIERTO (2026-07-10) — Realineación al core del ciclo de vida del arnés.** Foco =
  levantar·auditar·editar·publicar arneses de un marketplace; medición·observabilidad·mejora-continua
  → **Fase 2**. Arnesia = fábrica **agnóstica al rubro**. Aún SIN paquete formal (co-diseño visual, sin código).
- **Retomar aquí:** línea base del Mapa fijada → [`mockups/INDEX.md`](../../mockups/INDEX.md) +
  `mockups/arnesia-mapa-baseline.html` (baseline fiel del UI actual @`a01d845`; disciplina anti-drift
  enganchada en router/§10/ux). **Próximo:** el operador da comentarios de diseño → rehacer «Mapa DESTINO»
  como **superset del baseline** (capas Llenado+Deriva · sello · inicializador/doctor · receta en el
  inspector · conocimiento-del-proyecto, nombres a decidir) → 🧑‍⚖️ firma → paquete `stories/2026-07-10-<slug>/`
  + repriorizar BACKLOG Fase-1/Fase-2. Contexto: memorias `hs-repriorizacion-core-ciclo-arnes` +
  `hs-linea-base-ui-storybook-ssot`.
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
