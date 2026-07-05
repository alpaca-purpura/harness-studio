# ArnesIA — la fábrica de arneses (crear · mapear · observar · mejorar)

Producto standalone (graduado del monorepo `prenter-harness`, 2026-07-04; ex "Harness
Studio" — renombre del repo pendiente, debate abierto). Norte = [`VISION.md`](./VISION.md)
(**v3 FIRMADA**, ficha HS-02, 2026-07-04; + sección aditiva «Anatomía del arnés — fábrica
de cajas de proceso», reglas A1–A7, HS-03 it.9) · registro = [`LEDGER.md`](./LEDGER.md) (fichas
`HS-NN`; la historia OBS-01..OBS-20 vive en la incubadora
`prenter-harness/products/harness-studio/`, congelada).

**Qué es:** la fábrica de los arneses que alpacapurpura crea y vende por **rol × proceso**
de compañía. El vendible es el ARNÉS con mejora continua; ArnesIA es el medio de
producción. Solo arneses propios — nosotros seteamos el estándar. Constitución de 11
principios en VISION.md (proceso implícito · base antes de acción · rol×proceso · aditivo
sin pérdida · autodocumentación como efecto · guía sin bloqueo · agnóstico a rubro/tech ·
estándar propio · telemetría de nacimiento · nada sin eval · economía de contexto medible).
Ecosistema: ArnesIA es dueña única de observar y modificar; DevHub y apps de rol solo
ejecutan; marketplace git elegible por proyecto.

**Decisiones técnicas vigentes (HS-02):**
- Binario Go único `arnesia`: `serve` (watcher + indexer JSONL + API HTTP/SSE + UI
  embebida, :4200) · `open` · `index` · `publish`. Topología Syncthing/opencode.
- UI: **Vite + React SPA** vía `go:embed` (Next muere) · Mapa: **React Flow 12** + layout
  de carriles custom · **SQLite puro-Go** (modernc, WAL) como índice desechable; los JSONL
  de `~/.claude` son la fuente de verdad; Langfuse = espejo opcional, jamás dependencia dura.
- Mapa = lienzo único: banda Guardia (hooks) · carriles por fase del proceso · banda Base
  (knowledge); capas Estructura/Tokens/Desempeño/Proceso; crear/editar = acciones sobre el
  mapa.
- Creación conversacional: Claude Code headless por detrás — patrón conductor
  (I-76/OBS-16/OBS-18). Flujo: grill → spec → build headless → beta → evals-gate → promote
  por el release train del kit (KIT-06). La app OPERA las primitivas de P3, jamás las
  duplica. Contrato L0 `meta.clase` (I-75) — el grafo agnóstico es su evolución.
- Tauri 2 = milestone futuro "app vendible" (wrapper del mismo daemon, no rewrite). Wails
  v3 en watchlist.

**Estado:** fase 1 (Visión) ✓ · **fase 2 UX (HS-03) EN CURSO — iteración 10 de ~11-13**.
Norte de la fase = [`UX.md`](./UX.md) (decisiones firmadas · inventario de funcionalidades
al corte · backlog · registro iteración por iteración). **Reglas de negocio / metodología
cementadas** = [`METODOLOGIA.md`](./METODOLOGIA.md) (ArnesIA dueño de crear Y mantener ·
fábrica de cajas · qué debe tener cada componente · contrato de caja · reglas de honestidad ·
proceso de conformación — doc vivo, crece con las iteraciones UX) · mockup vigente = `mockups/arnesia-mockup-v3.html`
(navegable, artifact único — URL en memoria auto) · estándares mapeados =
`research/2026-07-04-salud-trazas-edicion.md`. Disciplina de iteración: mockups viven en
el repo · publicar siempre al MISMO artifact (parámetro `url`) · nada se entrega sin
click-through con asserts + screenshots revisados + consola limpia. Cero código de
producto aún; port del monorepo gobernado por la regla de VISION.md.

**Arnés de construcción:** kit dev — plugin `harness@prenter-marketplace` canal ESTABLE
(`alpacapurpura/prenter-marketplace`). Evoluciona con el producto — mejoras al arnés se
upstreamean al kit (backflow I-59), jamás fork silencioso.

**Git:** trunk-based — `main` única, commit/push directo, tags semver cuando haya releases.
