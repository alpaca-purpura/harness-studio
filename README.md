# ArnesIA

**La fábrica de arneses** del talento asistido por IA — **crear · mapear · observar · mejorar**.
Arneses por **rol × proceso** de compañía, con mejora continua; la observación es el sensor de esa
mejora. ArnesIA es el medio de producción; el vendible es el arnés.

> Repo graduado (2026-07-04) de la célula P4 del monorepo `prenter-harness`. Ex-«Harness Studio»
> (nombre muerto por colisión con Harness.io); renombre de la carpeta a `arnesia` pendiente.
> **Visión v3 FIRMADA** (HS-02, 2026-07-04).

## Documentos norte

Los docs se organizan en un **árbol de 4 ejes** (BACKLOG · ESTADO · CAPABILITIES · LEDGER); el
router de todo es [`CLAUDE.md`](./CLAUDE.md) — «necesito X → leo Y».

- **Qué SABE HACER el sistema hoy** (SSoT funcional, con punteros al código) — [`CAPABILITIES.md`](./CAPABILITIES.md)
- **Estado · fase · cifras vivas** (generadas, no tecleadas) — [`ESTADO.md`](./ESTADO.md)
- **Lo que viene / abierto** — [`BACKLOG.md`](./BACKLOG.md)
- **Historia de decisiones** — [`LEDGER.md`](./LEDGER.md) (índice → `ledger/HS-NN.md`; OBS-01..OBS-20 en la incubadora `prenter-harness`)
- **Visión** — [`VISION.md`](./VISION.md) (v3 firmada: constitución de 11 principios + anatomía A1–A7)
- **Reglas de negocio / metodología** — [`METODOLOGIA.md`](./METODOLOGIA.md) (doc vivo: fábrica de cajas, contrato de caja, honestidad)
- **UX del producto** — [`UX.md`](./UX.md) (fase 2 firmada it.13 · Command Rail + Organigrama)
- **Stack / tecnologías** — [`STACK.md`](./STACK.md)
- **Arquitectura as code** — [`arch/INDEX.md`](./arch/INDEX.md) (boundaries + convenciones `enforced_by:`; cifras vivas en [`ESTADO.md`](./ESTADO.md))
- **Estándar as code por elemento** — [`knowledge/INDEX.md`](./knowledge/INDEX.md) (nodos skill/hook/rule/…, se actualiza cada semana)

## Estado

**Fase 5 (Implementación) en curso** (fases 1–4 ✓: Visión · UX · Arquitectura · Specs). El
estado vivo, la fase y las **cifras generadas** (`arnesia conformance`) viven en
[`ESTADO.md`](./ESTADO.md); el port del monorepo sigue gobernado por el gran plan (nada se porta
sin pasar su fase).

## Arnés de construcción

Kit dev: plugin `harness@prenter-marketplace` (canal estable).

Privado — © Alpaca Púrpura / Prenter.
