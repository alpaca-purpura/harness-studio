# Ledger — ArnesIA (fichas HS-NN)

> Registro de decisiones de ESTE producto. Mismo formato/disciplina que la casa prenter-harness.
> **Continuidad:** la historia OBS-01..OBS-20 vive en
> `prenter-harness/products/harness-studio/LEDGER.md` (la incubadora, congelada en la
> graduación). Este repo arranca con prefijo **NUEVO `HS-NN`** — fork firmado: la visión mutó,
> identidad limpia (a diferencia de DevHub, que mantuvo DH-NN al graduarse).

## Fichas

### HS-01 · Fundación del repo propio — graduación de P4 con visión POR FORJAR — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-04):* "sabes, la visión ha mutado y necesito que tengamos nuestro
propio repositorio y trabajemos allí, vamos a ir migrando todo de a pocos, el repo se llamará
'Harness Studio', allí crearemos nuestra propia visión y usaremos los harneses del plugin de
nuestro marketplace."

*Desarrollo:* segunda graduación del monorepo-incubadora (precedente: DevHub/I-78, mismo día).
Forks firmados por AskUserQuestion: **prefijo `HS-NN` nuevo** (arranca HS-01; rompe
deliberadamente con el precedente DevHub — visión nueva, identidad limpia; OBS queda como
historia congelada) · **instancia 0 TAMBIÉN congelada** (`tooling/harness-studio/` — el
dogfood del operador entra al perímetro congelado; sigue SIRVIENDO tal cual está — gates y
eval-suites del monorepo operativos — pero nada nuevo se desarrolla ahí; su reemplazo lo
decide el port) · **canal ESTABLE** del plugin (`harness@prenter-marketplace`, como DevHub y
Vitalia). El repo nace limpio: cero código; VISION.md = marco + preguntas (la visión v3 se
forja aquí, ficha HS-02); el código (app `studio` Go+Next, shell de lentes, adapters,
eval-suites) entra por port gradual gobernado por esa visión. A diferencia de DevHub (que
graduó con visión ampliada YA firmada), aquí la graduación PRECEDE a la visión — decisión
consciente del operador.

*Conecta:* OBS-01..OBS-20 (la historia en la incubadora; OBS-20 = ficha espejo de esta
graduación) · I-NN de ecosistema (graduación — en `prenter-harness/tooling/strategy/LEDGER.md`)
· I-78 (precedente DevHub, mecánica replicada) · KIT-06 (el marketplace del que este repo
consume su arnés) · I-76/OBS-18 (la app y el loop cerrado — herencia técnica vigente).

*Siguiente:* **HS-02 = forjar y firmar la VISIÓN v3** (qué mutó · buyer/job vendible · orden
del port · fronteras con DevHub y P3). Hasta entonces, ninguna pieza se porta.

### HS-02 · Visión v3 forjada y FIRMADA — ArnesIA, la fábrica de arneses — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-04):* "Lo que yo deseo es tener una aplicación de escritorio que me
permita monitorear cualquier arnés… crear arneses, cargar arneses existentes que nosotros
hayamos creado y mejorarlos, y conectar con arneses que hayan sido desplegados desde un
marketplace… Debo poder ver los arneses como un mapa" · "Nombre: ArnesIA" · "firmo".

*Desarrollo:* primera sesión del repo cumple el mandato de HS-01. Proceso: investigación
4-frentes del state of the art 2026 (estándares SKILL.md/AGENTS.md · observabilidad
Langfuse/OTel · competidores · loops de mejora) → 3 forks estructurales firmados
(**grafo agnóstico + adaptador Claude Code** · **captura local-first + Langfuse espejo** ·
**git marketplaces + telemetría de efectividad propia**) → grill al operador (proceso as
code, principios, mapa) → mockup del Mapa validado en esencia → VISION.md v3 firmada.
Mutaciones clave vs herencia: arnés = **framework por rol × proceso** con constitución de
11 principios (7 constitución + 4 gobierno, incl. telemetría de nacimiento); **Mapa =
lienzo único** (Guardia · carriles de fase · Base; capas Estructura/Tokens/Desempeño/
Proceso); el vendible es el ARNÉS con mejora continua, ArnesIA = medio de producción; solo
arneses propios; nombre **ArnesIA** (muere "Harness Studio", colisión Harness.io); UI Next
muere (→ Vite+React SPA go:embed); stack decidido por investigación: binario Go `arnesia`
(serve/open/index/publish), React Flow 12 + carriles custom, SQLite puro-Go, JSONL de
~/.claude = fuente de verdad. Gran plan 6 fases: Visión ✓ → UX → Arquitectura/diseño
técnico → Specs → Implementación (MVP = Mapa) → Dogfood. Evidencia de mercado: el arnés
mueve Pass@1 casi tanto como el modelo (Claw-SWE-Bench 27.4 vs 29.4 pts); nadie combina
crear+mapa+monitoreo+publicar; ningún marketplace mide efectividad post-install.

*Conecta:* HS-01 (mandato) · KIT-06 (release train que el Tren opera) · I-76/OBS-16/OBS-18
(patrón conductor, sobrevive) · I-75 (contrato L0 — el grafo agnóstico es su evolución) ·
I-53/I-39 (frontera de datos: solo estrato seguro cruza) · DevHub (app de rol hermana:
ejecuta arneses, jamás los modifica).

*Siguiente:* **HS-03 = fase 2 del gran plan: UX del producto** (mapa a fondo, inspector,
flujos de creación/edición sobre el lienzo). Debates abiertos en VISION.md: segundo
cerebro · diseño fino del mapa · config de marketplace por proyecto · renombre del repo.

### HS-03 · Fase 2 — UX del producto FIRMADA (it.13: Command Rail + Organigrama) — `firmada` · `vig:vigente`

*Cruda (operador, 2026-07-04→05):* grill iterativo de 13 vueltas sobre el mockup del Mapa; los
insumos crudos vuelta-por-vuelta viven en [`UX.md`](./UX.md) (registro iteración por iteración).
Mandato heredado de HS-02: "Debo poder ver los arneses como un mapa… crear, cargar y mejorar".

*Desarrollo:* fase 2 del gran plan. **13 iteraciones sobre el lienzo** (mapa a fondo · inspector ·
flujos de creación/edición · capas Estructura/Tokens/Desempeño/Proceso · runbook de cadencia
semanal · S9 estándar pintado en el mapa) hasta el **shell firmado it.13 = Command Rail (A)**:
rail de iconos izq. + visual (mapa/portafolio) a pantalla casi completa + chat invocado (⌘K)
como dock derecho; Portafolio con dos lentes (**Organigrama** ↔ Cuadrícula). Retirado el
andamiaje REAL/DEMO del mockup. La fase cementó dos entregables: [`UX.md`](./UX.md) (decisiones
firmadas + inventario + backlog) y [`METODOLOGIA.md`](./METODOLOGIA.md) (reglas de negocio:
ArnesIA dueño de crear Y mantener · fábrica de cajas · contrato de caja §3 · reglas de honestidad).
Base de evidencia = árbol [`knowledge/`](./knowledge/INDEX.md) (11 nodos · **122 checks** al corte).

*Excepción declarada (2026-07-05):* HS-03 está **firmada** (regla Rust: firmado = congelado), PERO
`UX.md` y `METODOLOGIA.md` quedan **docs VIVOS** por decisión del operador — siguen creciendo con
nuevas iteraciones; no se congelan. (El árbol `knowledge/` también vive por diseño.) La firma
cierra la **fase**, no los documentos; ambos son la base de las specs de fase 4 y siguen mutando.

*Conecta:* HS-02 (mandato: mapa + crear/mejorar) · `knowledge/` (evidencia de la que METODOLOGIA
§2–3 deriva) · mockups `arnesia-shell-A-galaxia.html` (shell firmado) + `arnesia-mockup-v3.html`
(detalle de superficies) + `arnesia-shell-lab.html` (laboratorio, A firmado) · HS-04/HS-05
(arquitectura as code que sirve esta UX) · VISION §Anatomía A1–A7 (bajada a METODOLOGIA).

*Siguiente:* **HS-04 = fase 3 del gran plan: arquitectura y diseño técnico as code.**

### HS-04 · Fase 3 arrancada — stack + arquitectura as code (shell = Tauri 2 desde v1) — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-05):* "necesito que arranquemos la fase de arquitectura y diseño
técnico haciendo una investigación profunda de cómo hacer aplicaciones web multiplataforma y
conectándome a claude code instalado en la computadora, la primera instalación será en linux
(mint)… trabajemos bajo el concepto de arquitectura y diseño técnico as code para que, en vez de
tener una documentación pesada para mantener mi aplicación, la 'documentación' viva aquí mismo…
dame el stack completo que usaríamos, así como los patrones de arquitectura y diseño técnico."

*Desarrollo:* apertura de la fase 3 del gran plan (HS-03 UX firmado en it.13). **Investigación de
5 frentes en paralelo (subagentes), verificada julio-2026**, condensada en
[`research/2026-07-05-arquitectura-fase3.md`](./research/2026-07-05-arquitectura-fase3.md): (A)
shell/empaque multiplataforma · (B) conexión al Claude Code local · (C) contrato del dock (UI
agéntica de streaming, respondiendo el state-of-the-art que trajo el operador) · (D) local-first
(watch/índice/transporte/React Flow) · (E) arquitectura as code. **Stack** (confirma/refina
HS-02): binario Go único · Tauri 2 · subproceso-conductor stream-json · modernc SQLite desechable
(DuckDB descartado por CGO) · SSE multiplexado · React Flow 12 + Zustand+hash · **AG-UI
(taxonomía) + assistant-ui + CodeMirror6/merge** (nuevo, resuelve el dock) · JSON Schema 2020-12 →
quicktype (Go+TS) · go-arch-lint + depguard · D2+Mermaid.

**Forks firmados por el operador (AskUserQuestion):** (1) **shell v1 = Tauri 2 desde v1** (app de
escritorio nativa), NO browser-first — se acepta el riesgo WebKitGTK de Mint y sus mitigaciones
(webkit 4.1, `WEBKIT_DISABLE_DMABUF_RENDERER=1`, single-instance) pasan a load-bearing. Es una
**divergencia declarada** frente a la recomendación del research (browser-first por menor riesgo);
el boundary `core⊥shell` no cambia: el daemon sirve HTTP/SSE y Tauri es shell tonto encima →
«servable headless» sigue gratis y la ruta a la app vendible es aditiva (wrap, no rewrite). (2)
**materializar la arquitectura as code YA**.

**Materializado (arquitectura as code, espeja `knowledge/`):** nuevo árbol
[`arch/`](./arch/INDEX.md) — INDEX + CADENCE + **7 boundary nodes** (core⊥shell · dominio⊥transporte
· adaptadores-de-agente-intercambiables · índice-desechable-JSONL-es-verdad · conductor-no-parsea-
JSONL · permisos-GUI-human-in-the-loop · contrato-de-caja-es-fitness-function), cada uno L1
(principio con fuente) ↔ L2 (realización en el árbol Go) + tabla de checks (**29 checks**, severidad
+ señal, `enforced_by:` 1:1) — el gemelo-arquitectura de los 122 de knowledge; **`model/`** (C4
contexto Mermaid + container D2); **`contracts/`** (schemas L0 `meta.clase` + `contract:` de caja —
el mismo de METODOLOGIA §3 — + OpenAPI 3.1, con `gen/` para tipos Go+TS); **`fitness/`**
(`.go-arch-lint.yml` + `arch_test.go`). Un solo runner futuro `arnesia conformance` corre arch +
knowledge. **Corrección load-bearing propagada** al nodo `knowledge/headless-sdk` (v1.0→v1.1):
`--bare` **rompe el auth de suscripción** (no default-earlo en el conductor) + `%contexto` = métrica
derivada (no existe en stream-json/OTel) → 122 checks totales (headless v1.0=10 → v1.1=11; 1 alta
neta `ctx-derivado`, `bare-ci` fue corrección).

**Honestidad:** cero código de producto aún (fase 5); los 29 checks están **declarados, no
corriendo** (`status: proposed`) — se activan cuando el módulo Go aterrice, igual que los 122 de
knowledge son el linter futuro. Flags de verificación registrados en el research doc (shape
JSON-RPC de `control_request` = primer spike de fase 4; contención `~/.claude` = load-test;
macOS+Keychain; ToS si se distribuye).

*Conecta:* HS-02 (decisiones técnicas fundacionales que esto aterriza) · HS-03 (UX firmada que la
arquitectura debe servir) · I-75 (contrato L0 `meta.clase` = el schema `graph.l0`) · I-76/OBS-16/
OBS-18 (patrón conductor, hecho subproceso-stream-json) · KIT-03 (`telemetry-emit` OTel = el
sidechannel del conductor) · `knowledge/` (el árbol gemelo; `arch/` copia su patrón L1/L2/checks) ·
METODOLOGIA §3 (`contract:` de caja = `box.contract.schema.json`).

*Siguiente:* **HS-05 = fase 4, specs** *[corregido: HS-05 resultó ser **arquitectura FE as code**
—el operador detectó un hueco de FE antes de specs—; las specs corrieron a **HS-06**]* — cementar
los schemas del dominio (fase/estado spine, `meta.clase` por-formalizar), el spike del protocolo de
permisos contra el binario instalado, y las specs del MVP (Mapa primero). Debates de VISION que fase 4 hereda: segundo cerebro (vector DB) ·
config de marketplace por proyecto · renombre repo → `arnesia`.

### HS-05 · Arquitectura FE as code — FSD, taxonomía, tokens, storybook, convenciones — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-05):* "quiero saber si en la fase de arquitectura definiste también la
estructura de carpetas, convenciones de programación, atomic design claro con un storybook,
arquitectura de la aplicación a nivel de manejo de módulos, carpetas, etc. Quiero que todo lo tengamos
muy claro as code antes de comenzar con las especificaciones" · "ojo que es una aplicación instalable,
la idea es que sea un app de escritorio" · "Investiga 5-frentes".

*Desarrollo:* el operador detectó un **hueco en fase 3**: HS-04 aterrizó la arquitectura del **backend**
(7 boundaries hexagonales) pero NO el frontend as-code (solo picks de stack). HS-05 cierra el hueco antes
de specs. **Investigación de 5 frentes en paralelo (subagentes), verificada julio-2026**, condensada en
[`research/2026-07-05-fe-arch-atomic-storybook-convenciones.md`](./research/2026-07-05-fe-arch-atomic-storybook-convenciones.md):
(A) arquitectura de módulos · (B) atomic design/taxonomía · (C) Storybook como fitness visual · (D)
convenciones as code · (E) design tokens + styling. **Veredictos:** **FSD-lite** (topología; `pages`=
composition-roots por hash-state — app de escritorio sin router) con **dependency-cruiser** como CI-gate
(espeja go-arch-lint) + steiger secundario · **taxonomía plana de 6 capas direccionales**, NO atomic de
5 tiers como carpetas (canvas⊥chrome = frontera estrella); **shadcn sobre Base UI** (copy-in lintable) ·
**Storybook 10** + addon-vitest (**story = test que rompe CI**), regresión visual **local** (Chromatic
descartado por chocar con local-first) · **golangci-lint v2** + **Biome v2.4** (>ESLint+Prettier;
react-hooks ya stable) + `tsc` strictest + **lefthook** (binario Go) · **DTCG 2025.10** `.tokens.json` →
Style Dictionary v5 → **Tailwind v4** `@theme`, tokens verificados contra el mockup firmado (hex calzan).

**Reconciliaciones firmadas** (tensiones entre frentes): un solo enforcer de grafo (dependency-cruiser,
NO eslint-plugin-boundaries → evita drift) · FSD=topología + atomic=taxonomía (no compiten) · convención
de vars = **adoptar shadcn** (`--background/--foreground`; los mockups aportan valores, no nombres —
divergencia declarada vs `--bg/--surface`) · Biome y dependency-cruiser = concerns distintos, conviven.

**Materializado (arquitectura FE as code, gemelo de los 7 backend):** **5 boundary nodes** en
[`arch/boundaries/`](./arch/INDEX.md) (fe-topologia-fsd · fe-taxonomia-componentes · fe-transporte-
independiente · fe-tokens-contrato · fe-visual-fitness — **22 checks**) + nuevo árbol
[`arch/conventions/`](./arch/conventions/INDEX.md) (INDEX + CADENCE + **8 nodes**: go-style · ts-style ·
ts-types · naming · commits · hooks · editor · ci — **26 checks**), cada uno L1↔L2 + `enforced_by:`. **Config
files declarados** (raíz: `.golangci.yml`, `.editorconfig`, `lefthook.yml`, `.github/workflows/ci.yml`;
`web/`: `biome.json`, `tsconfig.json`, `.dependency-cruiser.js`, `.stylelintrc.json`, `steiger.config.ts`,
`tokens/base.tokens.json` [seed DTCG con valores del mockup]). **Gran total `arch/`: 12 boundaries + 8
conventions = 77 checks** (backend 29 + FE 22 + conventions 26). El runner futuro `arnesia conformance`
corre knowledge + arch (boundaries+fitness+conventions).

**Honestidad:** cero código aún (fase 5); los 77 checks y los config files están **declarados, no
corriendo** (`status: proposed`; `if: hashFiles(...)` en CI). Paths (`web/src/**`, module path Go) y la
ubicación de la SPA (`web/`) son **provisionales** hasta el scaffold Vite / `go mod init`. Flags de
verificación en el research doc (steiger beta no-extensible · determinismo de snapshots de canvas React
Flow · convención de vars de assistant-ui · composites DTCG con bugs en SD5 · Node ≥20.16/22.19/24 por SB10).

*Conecta:* HS-04 (backend as code que esto completa; mismo patrón L1/L2/checks) · HS-03 (UX firmada —
Command Rail A — que la topología FE sirve) · HS-02 (stack fundacional: Vite+React, React Flow 12, Zustand,
go:embed) · `knowledge/` (los tokens `--c-*` mapean a los 11 tipos de elemento; el runner unifica) ·
METODOLOGIA §3 (`contract:` de caja alimenta los tipos generados de `shared/api`).

*Siguiente:* **HS-06 = fase 4, specs** (heredada de la HS-04 «siguiente», corrida un número por esta
inserción): cementar los schemas del dominio (fase/estado spine, `meta.clase`), el spike del protocolo de
permisos contra el binario instalado, y las specs del MVP (Mapa primero). **Scope adicional cementado por
la auditoría de coherencia (2026-07-05):** (a) **contrato de check común** (schema unificado con
`enforced_by`/mecanismo por check) que vuelva ejecutable el runner único `arnesia conformance` sobre
knowledge + arch — hoy los checks de knowledge no llevan `enforced_by`; (b) **converger el enum L0
`meta.clase`** de 7 a los **11 tipos** de `knowledge/` + unificar labels (agente→subagent, regla→rule);
(c) **multi-sesión concurrente** — la app corre múltiples conversaciones/sesiones CC en paralelo
(central a la tesis fábrica; ya en el estándar: `headless-session-pin`, dock↔sesión S4, cajas en
paralelo). Base lista (conductor por-sesión `AgentPort.Spawn→Session`, Go concurrente, mapa=monitor),
faltan specs de: **(c.1)** `session_id`/`run_id` en cada frame SSE para rutear N streams (hoy multiplexa
por tipo `map|dock|run`, no por sesión); **(c.2)** superficie UX de N conversaciones paralelas (tabs/lista
de sesiones activas/progreso en el rail — el mockup firmado tiene un solo dock en foco); **(c.3)**
gobernanza (límite de sesiones concurrentes, presupuesto costo/`%contexto`/`max-turns` por sesión);
**(c.4)** contención de `~/.claude` bajo N procesos `claude` (el flag load-test que HS-04 dejó abierto).
Debates de VISION que fase 4 hereda: segundo cerebro · config de marketplace por proyecto · renombre
repo → `arnesia`.

<!-- Próximas: HS-06, … -->

## Log

| Fecha | Decisión | Fichas |
|---|---|---|
| 2026-07-04 | Fundación del repo propio: graduación de P4 (2ª graduación de la incubadora, precedente DevHub/I-78). Repo nace limpio con visión POR FORJAR (mutó — se firma aquí como HS-02); prefijo nuevo HS-NN; célula `products/harness-studio/` E instancia 0 `tooling/harness-studio/` congeladas como fuente del port gradual; kit dev como plugin del marketplace, canal estable. | HS-01 |
| 2026-07-04 | Visión v3 forjada y FIRMADA: **ArnesIA**, fábrica de arneses por rol × proceso. Constitución de 11 principios; Mapa = lienzo único; grafo agnóstico + adaptador CC; captura local-first + Langfuse espejo; git marketplaces + telemetría de efectividad; stack Go/Vite/React Flow/SQLite. Gran plan 6 fases — siguiente: UX (HS-03). | HS-02 |
| 2026-07-05 | Fase 2 (UX) FIRMADA en it.13: **shell = Command Rail (A)** (rail iconos izq. + visual casi-fullscreen + chat dock ⌘K) · Portafolio 2 lentes (Organigrama ↔ Cuadrícula) · 13 iteraciones sobre el lienzo · retiro andamiaje REAL/DEMO. Cementa `UX.md` + `METODOLOGIA.md` como **docs VIVOS** (excepción declarada a firmado=congelado). Base de evidencia = `knowledge/` (11 nodos · 122 checks). | HS-03 |
| 2026-07-05 | Fase 3 (arquitectura) arrancada: investigación 5-frentes verificada → stack completo (Tauri 2, subproceso-conductor stream-json, modernc SQLite, SSE, React Flow 12+Zustand, **AG-UI+assistant-ui+CodeMirror6**, schema-first Go+TS, go-arch-lint). Fork firmado: **shell = Tauri 2 desde v1** (divergencia declarada vs browser-first del research). **Arquitectura as code materializada:** árbol `arch/` (7 boundaries · 29 checks · model/ · contracts/ · fitness/) espejando `knowledge/`. Corrección `--bare`/auth propagada a knowledge (→122 checks). | HS-04 |
| 2026-07-05 | Fase 3 completada al FE: investigación 5-frentes (FSD · atomic · storybook · convenciones · tokens) verificada → **arquitectura FE as code**. Veredictos: **FSD-lite** + dependency-cruiser · 6 capas direccionales (canvas⊥chrome) + **shadcn/Base UI** · **Storybook 10** story=test (Chromatic descartado) · **golangci-lint v2 + Biome v2.4 + tsc strictest + lefthook** · **DTCG→Style Dictionary v5→Tailwind v4**. Materializado: **5 boundary nodes FE (22 checks) + `arch/conventions/` (8 nodes · 26 checks)** + config files declarados → **arch/ = 77 checks**. Specs corren a HS-06. | HS-05 |
