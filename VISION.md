# Harness Studio — Visión de producto

> **⚠ VISIÓN v3 — POR FORJAR.** El operador declaró en la graduación (2026-07-04): *"la visión
> ha mutado… allí crearemos nuestra propia visión"*. Este archivo es el marco de arranque, NO
> la visión firmada. Primera sesión de trabajo de este repo = forjar y firmar la visión (ficha
> HS-02). Registro: [`LEDGER.md`](./LEDGER.md).
> Historia y visión anterior: `prenter-harness/products/harness-studio/` (incubadora, congelada).

## Lo que se sabe al fundar (herencia vigente)

**Identidad heredada (OBS-17/OBS-19):** Harness Studio crea, observa y mejora los arneses del
talento asistido por IA. La observación = sensor de la mejora continua. Vendible a empresas
que gestionan su propia adopción de IA.

**El loop que ya cerró una vez (OBS-18, precedente):** la flota detectó un gap → la app forjó
la mejora con Claude Code headless → el release train la propagó (beta → smoke → promote →
update de installs). Detectar → forjar → propagar, con evidencia en el backend.

**Restricciones de ecosistema que viajan:**
- El motor de distribución/medición es P3 (el Kit): marketplace KIT-06, sensor de telemetría
  embebido (I-53), eval-gate. El Studio OPERA esas primitivas, no las duplica.
- El análisis de datos de cliente vive FUERA de los repos de fábrica (I-53/I-39); solo el
  estrato seguro (métricas · scores · hashes) cruza el borde.
- Cross-repo por contrato/ID/slug, nunca por ruta frágil (I-39).

## Preguntas que la visión v3 debe responder

1. ¿Qué mutó? — qué del "crear·observar·mejorar arneses" queda, qué muere, qué entra.
2. ¿Quién es el buyer y cuál es el job-to-be-done vendible (más allá del dogfood)?
3. ¿Qué se porta del monorepo, en qué orden, y qué NO se porta jamás?
4. ¿Qué relación tiene con DevHub (repo hermano) y con el marketplace P3 — fronteras y seams?
