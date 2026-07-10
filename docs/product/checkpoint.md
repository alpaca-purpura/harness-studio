# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-19** (homologación de metodología firmada + cierre coherente: fix del
drift de 65 links vivos). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Ninguno activo.** La homologación (`stories/2026-07-09-homologacion-metodologia/`) quedó
  `done` + firmada (HS-19). Sus continuaciones (upstream del método al kit · replicar a
  cockpit/dev-studio · forjar `/po`·`/architect`·`/dev-team`·`/auditor`) están en `docs/product/BACKLOG.md`.
- Paquetes con **gate humano de PARIDAD pendiente** (código listo, falta firma) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera la línea del ruleset; no editar a mano -->
- **ruleset `--todo`:** `247 checks · pass 40 · fail 0 · error 0 · deferred 207 · n/a 0` (medido 2026-07-09, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · 20 pass · 1 fail` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-09
- **arch/:** 17 boundaries (nuevo `codigo-traza-a-capability` **enforced**: R1/R2 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 82 — ~24 vivo · ~46 sin-check · 6 STUB · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera la línea del ruleset desde conformance (RF-178). Las cifras
> de arch/knowledge aún se teclean (deuda menor en BACKLOG). Drift histórico ya corregido.
