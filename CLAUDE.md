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
- Tauri 2 = **shell v1 desde el nacimiento** (fork firmado HS-04 — supera la lectura previa
  "milestone futuro"; ver Decisiones de arquitectura abajo). La ruta a "app vendible" es aditiva
  (wrap del mismo daemon, no rewrite). Wails v3 en watchlist.

**Decisiones de arquitectura (HS-04, fase 3 — 2026-07-05):** investigación 5-frentes verificada
(`research/2026-07-05-arquitectura-fase3.md`). **Shell v1 = Tauri 2 DESDE v1** (fork firmado; el
milestone «futuro» se adelantó — daemon como sidecar `externalBin`, el WebView consume la misma
API; mitigaciones Mint load-bearing: webkit 4.1, `WEBKIT_DISABLE_DMABUF_RENDERER=1`,
single-instance). **Conexión a CC** = Go spawnea el `claude` local, habla **stream-json** por
stdin/stdout (subproceso-conductor; SDK-sidecar y Managed Agents descartados). **Event sourcing**:
stream-json (live) + OTel (hooks/skills) + JSONL (enumerar/replay, jamás parsear). **Dock** =
taxonomía **AG-UI** sobre SSE (emisor Go propio) + **assistant-ui** + **CodeMirror 6/merge**;
component-selection (format-authoring = trampa). **SSE** multiplexado 1 conexión. **Estado FE** =
Zustand + hash-state. **Arquitectura as code** = árbol [`arch/`](./arch/INDEX.md) (16 boundaries +
`conventions/` = **97 checks** · schemas L0+contrato · go-arch-lint) espejando `knowledge/`; runner
futuro `arnesia conformance` corre ambos. Corrección propagada: `--bare` rompe auth de suscripción
(knowledge headless v1.1). **HS-06** endureció el shell v1 (2 boundaries `enforced`: superficie-local-
confinada + sesion-viva-consistente; fitness tests corriendo).

**Decisiones de arquitectura FE (HS-05, fase 3 — 2026-07-05):** investigación 5-frentes verificada
(`research/2026-07-05-fe-arch-atomic-storybook-convenciones.md`) — cierra el hueco FE de HS-04. **Topología =
FSD-lite** (`web/src/{app,pages,widgets,features,entities,shared}`; `pages`=composition-roots por hash-state
—app de escritorio SIN router—; enforcer = **dependency-cruiser** gate + steiger, NO eslint-plugin-boundaries).
**Componentes** = 6 capas de UI direccionales (**canvas⊥chrome** = regla estrella), NO atomic de 5 tiers como
carpetas; primitivos = **shadcn sobre Base UI** (copy-in lintable). **Tokens** = **DTCG 2025.10** `.tokens.json`
(SSOT, valores del mockup firmado, nombres shadcn) → **Style Dictionary v5** → **Tailwind v4** `@theme`; dark
`data-theme`; stylelint anti-magic-value. **Fitness visual** = **Storybook 10** story=test (Chromatic descartado
= local-first). **Convenciones** ([`arch/conventions/`](./arch/conventions/INDEX.md)): **golangci-lint v2** +
**Biome v2.4** (>ESLint) + `tsc` strictest + **lefthook** (binario Go). Config files declarados (raíz + `web/`).

**Doctrina propia v1 (HS-07, fase 4 — 2026-07-05):** replanteo cruzando DAOP v0.2 (BMAD+Agent SDK, 5
subagentes) + barrido de 7 fuentes externas (4 subagentes) — `research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`.
**Reencuadre clave: ArnesIA OPERACIONALIZA Agentic BPM** (manifiesto *Information Systems* 2026 /
arXiv 2603.18916) — doctrina PROPIA basada en proceso e independiente de rubro, **NO clon de BMAD**.
Regla de Rosetta: DAOP-«Arnés»→nuestra **CAJA** · DAOP-«manifiesto de rol»→nuestro **ARNÉS** (VISION
intacta). **Firewall CC-native** (`no-phantom-frontmatter`: prohibido `persistent_facts`/`customize.toml`/
sanctum — CC los ignora). Vocabulario propio: framed autonomy · autonomy≠automation · adaptation/evolution ·
4+1 capacidades (nueva: **explainability**). **3 decisiones firmadas:** producto-puro + **META de enganche**
(L1/organigrama = sistema externo futuro; cada arnés carga rol·proceso·reporta-a·empresa) · **dogfood-first**
(arnés real antes del Mapa) · **P6/Guardia por scope**. As-code: VISION §Linaje · METODOLOGIA §3 contrato
FUSIONADO (intención+cableado+Gherkin+arquetipo+perfil_harness) + §8 doctrina de proceso · **nodo 12
`harness-profile`** · 2 boundaries (`orquestacion-determinista-entre-cajas` · `permisos-derivan-del-rol`).

**Estado:** fase 1 (Visión) ✓ · fase 2 UX (HS-03) ✓ firmada it.13 · **fase 3 arquitectura (HS-04
backend + HS-05 frontend) — as code COMPLETA** · **HS-06 endurecimiento multisesión + confinamiento local
(aislamiento CC por-tab confirmado real; S1 auth Host+Origin+token · S2 cwd por arnés · S3–S6 stream
confiable; 2 boundaries nacen enforced)** · **HS-07 doctrina propia v1 as-code (operacionalizamos Agentic
BPM; contrato fusionado + nodo `harness-profile` + 2 boundaries → knowledge 12 nodos·138 checks · arch
16 boundaries·97 checks) — siguiente HS-08 = specs (dogfood-first)**. Shell UX = **Command Rail (A)** + **multisesión (it.14)**: el borde izquierdo es
un **rail de sesiones** (tabs paralelas tipo WARP, colapsable a gutter; sesión = **frente de trabajo**
N:1 con arnés, con su conversación CC viva; estado CC vivo `streaming/await/idle`; persisten) · vistas =
tira slim por sesión · visual a pantalla casi completa · chat invocado (⌘K) como dock derecho colapsable. Portafolio con
2 lentes: **Organigrama** (arneses por empresa/puesto, «reporta a», libre 2D) ↔ Cuadrícula.
App = fábrica de arneses (crear + mantener), NO cockpit de empresa; metadata puesto·empresa·
reporta-a + marketplace por empresa/arnés. Mockups de shell: `mockups/arnesia-shell-lab.html`
(compara 4 paradigmas) · `mockups/arnesia-shell-A-galaxia.html` (dirección firmada). **Próximo:
fase 4 specs (HS-08, dogfood-first: forjar el arnés dev-full-cycle real end-to-end ANTES del Mapa) →
schemas del dominio (fase/estado · `meta.clase`+arquetipo+perfil_harness) + spike de permisos
(control_request, adoptando permisos-por-rol+TTL) + MVP Mapa; restos multisesión (worktree, gobierno
de presupuesto, OTel).**
Norte de la fase = [`UX.md`](./UX.md) (decisiones firmadas · inventario de funcionalidades
al corte · backlog · registro iteración por iteración). **Reglas de negocio / metodología
cementadas** = [`METODOLOGIA.md`](./METODOLOGIA.md) (ArnesIA dueño de crear Y mantener ·
fábrica de cajas · qué debe tener cada componente · contrato de caja · reglas de honestidad ·
proceso de conformación — doc vivo, crece con las iteraciones UX) · **estándar as code por
elemento** = [`knowledge/`](./knowledge/INDEX.md) (árbol de conocimiento VIVO: **12 nodos** —skill·
hook·rule·subagent·command·mcp·plugin·settings·output-style·statusline·headless· **perfil-de-harness**
(HS-07)—, cada uno L1 oficial+expertos ↔ L2 nuestra adaptación + checklist evaluable = **138 checks**; se
actualiza CADA SEMANA vía [`knowledge/CADENCE.md`](./knowledge/CADENCE.md), NO se congela al firmar HS-03) ·
**arquitectura y diseño técnico as code** = [`arch/`](./arch/INDEX.md) (árbol gemelo de knowledge:
**16 boundary nodes** [9 backend + 5 FE; HS-06 sumó superficie-local-confinada + sesion-viva-consistente;
HS-07 sumó orquestacion-determinista-entre-cajas + permisos-derivan-del-rol]
L1↔L2 + [`conventions/`](./arch/conventions/INDEX.md) [8 nodes:
go/ts-style·types·naming·commits·hooks·editor·ci] = **97 checks** `enforced_by:` · `model/` C4 ·
`contracts/` schemas L0+contrato · `fitness/` go-arch-lint [Go] + enforcers FE en `web/`
[dependency-cruiser·steiger·stylelint·biome] + configs de estilo en raíz; se revisa al cambiar, mecanismo
en `arch/CADENCE.md` y `arch/conventions/CADENCE.md`) ·
mockups vigentes (se unifican al portar) = `mockups/arnesia-shell-A-galaxia.html` (shell
firmado it.13) + `mockups/arnesia-shell-A-sessions.html` (**multisesión firmado it.14**: rail de
sesiones tipo WARP colapsable a gutter · sesión = frente N:1 con arnés · chat invocado colapsable ·
persisten) + `mockups/arnesia-session-lab.html` (lab de 4 paradigmas de navegación multisesión) +
`mockups/arnesia-mockup-v3.html` (detalle de superficies; navegable, URL en memoria auto) ·
estándares mapeados =
`research/2026-07-04-salud-trazas-edicion.md`. Disciplina de iteración: mockups viven en
el repo · publicar siempre al MISMO artifact (parámetro `url`) · nada se entrega sin
click-through con asserts + screenshots revisados + consola limpia. Cero código de
producto aún; port del monorepo gobernado por la regla de VISION.md.

**Arnés de construcción:** kit dev — plugin `harness@prenter-marketplace` canal ESTABLE
(`alpacapurpura/prenter-marketplace`). Evoluciona con el producto — mejoras al arnés se
upstreamean al kit (backflow I-59), jamás fork silencioso.

**Git:** trunk-based — `main` única, commit/push directo, tags semver cuando haya releases.
