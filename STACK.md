# STACK — decisiones técnicas vigentes de ArnesIA

> vig: activo · revisar: 2026-08-01
> Puntero al stack. El **por qué** de cada decisión → `VISION.md` (HS-02) + `ledger/`; la
> **arquitectura ENFORZADA as-code** → [`arch/INDEX.md`](./arch/INDEX.md) (boundaries +
> convenciones, go-arch-lint/dependency-cruiser). Las **cifras vivas** → `ESTADO.md`.

## Backend
- Binario Go único `arnesia`: `serve` · `open`(STUB) · `index` · `publish`(STUB) · `conformance`. :4200.
- Hexagonal (`internal/{domain,usecase,ports,adapters}`). Conexión CC = **subproceso-conductor
  stream-json** por stdin/stdout (SDK-sidecar y Managed Agents descartados).
- ⚠ **Índice = map in-memory desechable** (`adapters/index`) · **persistencia = archivos JSON
  atómicos** en `~/.arnesia` (`adapters/store`). El **índice SQLite puro-Go (modernc, WAL) es
  FASE 5 FUTURA, NO implementado hoy** — corrige la lectura previa de CLAUDE.md («SQLite» como
  vigente). Ver `ledger/HS-18.md` (hallazgo de drift, barrido de capabilities 2026-07-09).
- JSONL de `~/.claude` = fuente de verdad; Langfuse = espejo opcional, jamás dependencia dura.
- Doctrina embebida (`go:embed`) → conformance portable; kit inyectado por flags al spawn (②↛③).

## Frontend
- Vite + React SPA vía `go:embed`. **FSD-lite** (`web/src/{app,pages,widgets,features,entities,shared}`).
- Mapa = **sustrato HTML+SVG** (bandas/carriles + overlay SVG), NO React Flow (RF → Organigrama futuro).
- Estado = **Zustand + hash-state** (app de escritorio sin router). Dock = AG-UI/SSE + assistant-ui +
  CodeMirror 6/merge (assistant-ui staged).
- Tokens **DTCG** → Style Dictionary v5 → **Tailwind v4** `@theme`. **Storybook 10 story=test**
  (única validación FE hoy — no hay `*.test.*`/e2e en `web/`).

## Shell
- **Tauri 2 desde v1** (bin `arnesia-app`); daemon Go = **sidecar externalBin**; single-instance
  reenfoca; **token minted por lanzamiento** = raíz de confianza. `WEBKIT_DISABLE_DMABUF_RENDERER=1`
  (Linux/WebKitGTK 4.1). Ruta a «app vendible» = aditiva (wrap del mismo daemon).

## As-code (enforcement)
- `arch/` = 16 boundaries + `conventions/` = ~101 checks `enforced_by:` · `fitness/` go-arch-lint.
- `knowledge/` = estándar por elemento (12 nodos · 138 checks). Motor único `arnesia conformance`.
- `CAPABILITIES.md` = SSoT funcional (qué hace el sistema, punteros al código).
- Cifras vivas GENERADAS → `ESTADO.md`. Historia de decisiones de stack → `ledger/`.
