# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-20** (cierre de los 4 gates humanos de PARIDAD + R4 de capabilities
aterrizado + `estado.sh` genera todo el bloque de cifras). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Ninguno activo.** Los 4 gates humanos de PARIDAD (chat-cc-funcional · franja-artefactos ·
  boton-actualizar · inspector-drawer) quedaron **FIRMADOS** (HS-20) — ya no hay código-listo-sin-firma.
- Continuaciones abiertas (homologación 2° orden · cola de capabilities · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `247 checks · pass 42 · fail 0 · error 0 · deferred 205 · n/a 0` (medido 2026-07-09, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-09
- **arch/:** 17 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 82 — 38 vivo · 40 vivo·nc · 1 parcial · 3 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. Falta solo cablearlo a un hook/CI (hoy es manual). Drift histórico ya corregido.
