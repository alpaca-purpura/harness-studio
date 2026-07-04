# Harness Studio — crear · observar · mejorar los arneses del talento asistido por IA

Producto standalone (graduado del monorepo `prenter-harness`, 2026-07-04). Norte =
[`VISION.md`](./VISION.md) (**por forjar AQUÍ** — la visión mutó en la graduación; hasta
firmarla, la herencia vigente manda) · registro = [`LEDGER.md`](./LEDGER.md) (fichas `HS-NN`,
arranca en HS-01 — prefijo NUEVO, firmado en la graduación; la historia OBS-01..OBS-20 vive en
la incubadora `prenter-harness/products/harness-studio/`, congelada).

**Qué es (herencia vigente, sujeta a la visión nueva):** el producto que crea, observa y
mejora los arneses del talento asistido por IA — la observación es el sensor de la mejora
continua, no el producto entero. Vendible a empresas que gestionan su propia adopción de IA.

**Decisiones técnicas heredadas que siguen vigentes (hasta que la visión nueva las confirme o
las mate):**
- App local: binario Go (`studio`, :4200) + UI Next embebida (`go:embed`), Claude Code
  headless por detrás — patrón conductor (I-76/OBS-16/OBS-18).
- Flujo del creador: grill → spec → build headless → beta → evals-gate → promote, por el
  release train del kit (KIT-06). La app OPERA las primitivas de P3, jamás las duplica.
- Shell de lentes = visor ligero cero-dep (D2 cerrada, OBS-13); contrato L0 con `meta.clase`
  (I-75); el producto hospeda solo `clase: arnes`.

**Estado:** repo recién fundado — cero código; TODO vive aún en el monorepo (congelado: célula
`products/harness-studio/` E instancia 0 `tooling/harness-studio/`) y entra por **port
gradual**, pieza por pieza, gobernado por la visión que se firme aquí.

**Arnés de construcción:** kit dev — plugin `harness@prenter-marketplace` canal ESTABLE
(`alpacapurpura/prenter-marketplace`). Evoluciona con el producto — mejoras al arnés se
upstreamean al kit (backflow I-59), jamás fork silencioso.

**Git:** trunk-based — `main` única, commit/push directo, tags semver cuando haya releases.
