# ArnesIA

**La fábrica de arneses** del talento asistido por IA — **crear · mapear · observar · mejorar**.
Arneses por **rol × proceso** de compañía, con mejora continua; la observación es el sensor de esa
mejora. ArnesIA es el medio de producción; el vendible es el arnés.

> Repo graduado (2026-07-04) de la célula P4 del monorepo `prenter-harness`. Ex-«Harness Studio»
> (nombre muerto por colisión con Harness.io); renombre de la carpeta a `arnesia` pendiente.
> **Visión v3 FIRMADA** (HS-02, 2026-07-04).

## Documentos norte

Todo (salvo skills y `CLAUDE.md`) vive en [`docs/`](./docs/README.md), organizado por el método del
kit `harness@prenter-marketplace` en 3 zonas — **product · architecture · process**. El router
«necesito X → leo Y» es [`CLAUDE.md`](./CLAUDE.md); el front-door del proceso es el skill `/pm`.

- **Dónde estamos · fase · cifras vivas** (generadas) — [`docs/product/checkpoint.md`](./docs/product/checkpoint.md)
- **Qué SABE HACER el sistema hoy** (SSoT funcional, YAML por-cap) — [`docs/product/capabilities/INDEX.md`](./docs/product/capabilities/INDEX.md)
- **Lo que viene / abierto** — [`docs/product/BACKLOG.md`](./docs/product/BACKLOG.md)
- **Historia de decisiones** — [`docs/product/LEDGER.md`](./docs/product/LEDGER.md) (índice → `ledger/HS-NN.md`)
- **Visión / constitución** (11 principios + anatomía A1–A7) — [`docs/product/vision.md`](./docs/product/vision.md)
- **UX firmada** (Command Rail + Organigrama) — [`docs/product/ux.md`](./docs/product/ux.md)
- **Metodología · reglas de negocio · disciplina** — [`docs/process/metodologia.md`](./docs/process/metodologia.md)
- **Arquitectura as-code + stack** — [`docs/architecture/INDEX.md`](./docs/architecture/INDEX.md) · [`docs/architecture/stack.md`](./docs/architecture/stack.md)
- **Estándar as-code por elemento** — [`docs/architecture/knowledge/INDEX.md`](./docs/architecture/knowledge/INDEX.md)

## Estado

**Fase 5 (Implementación) en curso** (fases 1–4 ✓: Visión · UX · Arquitectura · Specs). El estado
vivo, la fase y las **cifras generadas** (`arnesia conformance`) viven en
[`docs/product/checkpoint.md`](./docs/product/checkpoint.md); el port del monorepo sigue gobernado
por el gran plan (nada se porta sin pasar su fase).

## Arnés de construcción

Kit dev: plugin `harness@prenter-marketplace` (canal estable).

Privado — © Alpaca Púrpura / Prenter.
