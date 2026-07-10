# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-17** (aislamiento de superficie de config). Índice de historia →
`LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **`docs/product/stories/2026-07-09-homologacion-metodologia/`** — homologación al método del plugin
  `harness@prenter-marketplace` (3 repos): `docs/{product,architecture,process}` + seam
  `project.config.yaml` + skill `/pm` + capabilities-as-YAML. → estado/próximo-paso en su `INDEX.md`.
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
