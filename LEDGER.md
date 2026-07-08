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

> **Nota (2026-07-07, sync HS-10):** al corte de la firma eran **121 checks**; el check 122 llegó
> después, con la corrección `--bare` de HS-04 (nodo `headless-sdk` v1.0→v1.1). El «122 al corte»
> de arriba se conserva como historia; la fila del Log de esta ficha arrastra el mismo matiz.

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

### HS-06 · Endurecimiento multisesión + confinamiento de la superficie local — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-05):* "Quiero que revises la aplicación en tauri desarrollada y me
comentes si técnicamente ya separaste por completo el contexto y manejo de claude code entre tabs
de sesión… Audita a profundidad" → "Si, investiga cada punto a profundidad y soluciónalos, debemos
estar en todos en Alta. Eso sí, todo diséñalo primero, ve que haya coherencia total en el diseño y
arquitectura técnica".

*Desarrollo:* auditoría a fondo del shell Tauri v1 (`a0e4fe2`). **Veredicto: el aislamiento de
contexto CC por tab SÍ es real** a nivel de proceso — 1 ventana → 1 daemon → N subprocesos `claude`,
cada uno con su `ClaudeSessionID`; el FE rutea por `session_id` con buffer por-sesión. Pero la
auditoría destapó **7 huecos**. Coherencia primero (design-mode + forks firmados por AskUserQuestion),
luego implementación en las 4 capas + materialización as-code. Hallazgo clave: **5 de los 7 no eran
arquitectura nueva** — caían sobre checks ya declarados o sobre el scope multisesión ya cementado en
«Siguiente» de HS-05: S2 = la prosa `permisos-gui` «aislar sesiones por cwd/worktree» (que el
conductor shippeado violaba, corriendo todo en un cwd global); c.1 run_id = S5; c.3 max-turns/gobierno
= S3 + fold-in; c.4 contención `~/.claude` = S2. **Solo S1 (auth de la superficie) era un boundary
genuinamente ausente** — ninguna regla cubría «quién le habla al daemon» (con CORS `*` + sin token,
cualquier web abierta en el navegador podía POSTear turnos y conducir un agente con acceso al
filesystem).

**Forks ratificados (AskUserQuestion):** (1) **S2 = registro explícito arnés→ruta** (puerto
`WorkdirResolver` + `store/arnes_registry.go`; fallback aislado por arnés si no registrado — jamás el
cwd global; worktree-por-sesión = upgrade futuro del mismo puerto). (2) **S1 = el shell emite el
token** (Tauri mint por lanzamiento → env `ARNESIA_AUTH_TOKEN` al spawnear el sidecar → WebView por
`invoke('auth_token')`; dev sin Tauri → daemon bajo Host+Origin; shell = raíz de confianza,
coherente con core⊥shell). (3) **ficha = HS-06** (las specs completas corren a HS-07).

**Materializado (código + arch as-code, nacen enforced):** **2 boundaries nuevos** —
[`superficie-local-confinada`](./arch/boundaries/superficie-local-confinada.md) (middleware `withAuth`
en 3 gates: Host anti-rebinding · Origin allowlist reflejado —adiós CORS `*`— · token constant-time;
6 checks) + [`sesion-viva-consistente`](./arch/boundaries/sesion-viva-consistente.md) (un-turno-a-la-vez
`ErrBusy`→409 · emit bloqueante + broker shed-on-lag→replay `Last-Event-ID` · `run_id` en todo frame +
dedup FE `finalizedRun` · `tryHealResume` para `--resume` stale; 4 checks) — **+ 2 checks a
`permisos-gui`** (`sesion-aislada-por-cwd`, `cwd-path-contenido`) y `--max-turns` ahora satisfecho por
el conductor. **Fitness tests reales que PASAN** (`go test ./arch/fitness/...`, `-race` limpio):
8 tests HS-06. **Contrato OpenAPI reconciliado** (drift `/dock/*`→`/sessions/*` real + `/arneses` +
`securitySchemes` bearer/`?token=` + `DockFrame` con `run_id` + 409). **go-arch-lint**: mapeado el
componente `store` faltante. **arch/ = 14 boundaries · 63 checks + 26 conventions = 89.**

**Honestidad:** el shell Rust (`lib.rs` + `Cargo.toml` `getrandom`) se ESCRIBIÓ pero **no se compiló
aquí** (sin toolchain Rust ni red — coherente con «verificar al instalar» del Cargo.toml); se valida
con `cargo build` en la máquina provisionada. [*Actualización HS-09:* confirmado **compilado +
corriendo** — `web/src-tauri/target/debug/arnesia` + sidecar `arnesia-daemon`, commit `a0e4fe2`; la nota
describe el delta Rust de HS-06 sin recompilar en esa ola.] **Explícitamente fuera de scope → HS-07:** diff-approval
/ `--permission-mode` / protocolo `control_request` (el spike de permisos que HS-04 ya reservaba) ·
OTel + `system/api_retry` (warn) · worktree-git-por-sesión · UX completa de registro de ruta + gobierno
de presupuesto (c.2/c.3 restantes). El backend Go compila + vetea limpio; el FE typa (tsc strictest) +
lintea (Biome) limpio.

*Conecta:* la auditoría del shell (`a0e4fe2` HS-03 it.14) · HS-04 (arch backend as code + patrón
conductor + `permisos-gui` cuya prosa esto vuelve ejecutable) · HS-05 (scope multisesión c.1/c.3/c.4
que esto ejecuta; arch FE) · `knowledge/headless-sdk` (`--max-turns`, session-pin — dogfood) · I-76/
OBS-18 (patrón conductor).

*Siguiente:* **HS-07 = fase 4, specs** (heredada): cementar schemas del dominio (fase/estado spine,
`meta.clase` 7→11 · label unify), el **spike del protocolo de permisos** (`control_request` contra el
binario instalado) + diff-approval, las specs del MVP (Mapa primero), y los restos del scope
multisesión (worktree por sesión · gobierno de presupuesto costo/`%contexto` · OTel/api_retry). Debates
de VISION heredados: segundo cerebro · config de marketplace por proyecto · renombre repo → `arnesia`.

### HS-07 · Doctrina propia v1 — operacionalizamos Agentic BPM (cruce DAOP + barrido externo) — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-05):* "necesito de una vez setear la doctrina de cómo construiremos
nuestros arneses" → "vamos a replantear toda nuestra doctrina creando nuestro propio estándar, aquí
una investigación que he hecho [DAOP v0.2] que necesito que adaptemos a nuestra visión… lanza
subagentes para revisar cada una de las capas" → "necesito… me digas si nos ayuda… para no ser un
[B]MAD clone sino nuestra propia doctrina basada en proceso e independiente al rubro" [+ 7 fuentes
externas] → "si Ratifico y ejecuta todo as code… dando **prioridad total a que todo esto sea nuestra
doctrina**".

*Desarrollo:* tres pasos. (1) Se cementó METODOLOGIA §2 (estructura obligatoria de subagente·hook·
rule/conocimiento, derivada de `knowledge/`). (2) **Cruce de DAOP v0.2** (investigación del operador,
derivada de BMAD v6 + Agent SDK) contra las 5 capas vía estudio 5-subagentes → propuesta
`research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`. Regla de Rosetta: **DAOP-«Arnés» ≈ nuestra
CAJA · DAOP-«manifiesto de rol» ≈ nuestro ARNÉS** (VISION intacta). Firewall CC-native: los BMAD-ismos
(`persistent_facts`, `customize.toml`, sanctum) que CC ignora en silencio se RECHAZAN/traducen. **3
decisiones firmadas:** producto-puro + META de enganche (L1/organigrama = sistema externo futuro) ·
dogfood-first (arnés real antes del Mapa) · frontera P6/Guardia por scope. (3) **Barrido de 7 fuentes
externas** (4 subagentes) → el reencuadre clave: el **manifiesto Agentic BPM** («A Research Manifesto»,
18 autores BPM; *Information Systems* 2026 / arXiv 2603.18916) da el **linaje que nos saca de clon de
BMAD** — **operacionalizamos Agentic Business Process Management** (disciplina de proceso, independiente
de dominio). Vocabulario propio: *framed autonomy*, *autonomy≠automation*, *adaptation/evolution*, 4+1
capacidades. Sierra ADLC valida el ciclo-de-vida-de-producción; Salesforce Agentforce valida el eje
despliegue-en-orgs. Único gap conceptual nuevo = **explainability**.

*Materializado (as-code, prioridad = doctrina PROPIA):* **VISION** — nota aditiva «Linaje doctrinal y
precisiones» (APM, framed autonomy, precisión P6 ratificada, seam producto-puro, dogfood-first) sin
tocar los 11 principios. **METODOLOGIA** — §3 contrato de caja FUSIONADO (intención SPEC-kernel
why/capabilities/constraints/non-goals + cableado it.10 intacto + aceptación Gherkin + `arquetipo` +
`perfil_harness` + `escritor_unico` + `handoff`) · **§8 doctrina de proceso** (arquetipos · perfil de
harness T1–T3 · document-as-cache · gate de fidelidad · frontera P6/Guardia · firewall CC-native).
**knowledge/** — nace el **nodo 12 [`harness-profile`](./knowledge/elements/harness-profile.md)** (11
checks; L1 = APM+Sierra+Salesforce+CC) + skills/rules/subagents bumpeados (+5 checks del firewall) →
**12 nodos · 138 checks**. **arch/** — **2 boundaries draft** ([`orquestacion-determinista-entre-cajas`](./arch/boundaries/orquestacion-determinista-entre-cajas.md)
· [`permisos-derivan-del-rol`](./arch/boundaries/permisos-derivan-del-rol.md); 8 checks) → **16
boundaries · 97 checks**. Bibliografía embebida en cada doc («Para profundizar»).

*Conecta:* HS-02 (VISION — los 11 principios que esto ancla, no altera) · HS-04/05 (arch as code que
esto extiende; el `contract:` schema) · HS-06 (multisesión/conductor sobre el que corre el loop T3;
`permisos-gui` que `permisos-derivan-del-rol` extiende de fase→rol) · `knowledge/` (los 12 nodos) ·
`research/…inyeccion-knowhow` (el `KitProvisioner` que parametriza permisos por rol) · fuentes externas
(manifiesto Agentic BPM · Sierra ADLC · Salesforce Agentforce · DAOP v0.2 filtrado).

*Siguiente:* **HS-08 = fase 4, specs** (heredada de HS-06, que reservaba «HS-07 specs»; la doctrina v1
tomó HS-07 por ser fundacional a las specs): schemas del dominio (fase/estado spine · `meta.clase` +
`arquetipo`/`perfil_harness` en el contrato) · **spike `control_request`** (adoptando permisos-por-rol
+ TTL) · specs del MVP con **dogfood-first** (forjar el arnés dev-full-cycle real antes del Mapa) ·
restos multisesión (worktree · gobierno de presupuesto · OTel). Abiertas de doctrina: persona-state/
DevHub · nombre de la doctrina · detalle A2 (quién ejecuta `ruta`).

### HS-08 · Fase 4 specs (dogfood-first) — doctrina BAJADA A EJECUTABLE — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-05):* "la doctrina v1 quedó as-code pero SIN diente — el schema rechaza el
contrato fusionado, no hay spine de estados, los `enforced_by:` cuelgan de una herramienta que no existe;
antes del Mapa, forja el arnés `dev-full-cycle` real y hazlo verificable de verdad" (dogfood-first — la
§Linaje de VISION exige el arnés real ANTES del Mapa).

*Desarrollo:* la auditoría (`research/2026-07-05-auditoria-doctrina-v1-aplicabilidad.md`) confirmó el hueco:
doctrina declarada, no ejecutable. HS-08 la aterriza. (1) **Contrato de caja fusionado con diente** — bajado
a `box.contract.schema.json` + `domain.Contract` (3 ejes: intención + cableado + aceptación), con
`TestBoxContractValidatesAgainstSchema` activo (el schema ya no rechaza el contrato de METODOLOGIA §3). (2)
**`Clase` canónica de 10 primitivas** (enum único, se acabó el doble enum). (3) **Manifiesto del arnés en
`graph.l0`** — carga `fases` · `spine` (estados DECLARADOS por el arnés, no fijos) · META `rol`·`proceso`·
`reporta_a`·`empresa`. (4) **Motor `arnesia conformance` CONSTRUIDO** (hexagonal): `RulesetPort` parsea
`knowledge/` + `arch/` = **235 checks a datos**; `ConformancePort` + adapters por mecanismo (arch-test ·
schema-validation · go-arch-lint · static-scan · nl-judge); los checks de consistencia de spine se
parametrizan por el **spine declarado** del arnés (agnóstico). Los `enforced_by:` ya no cuelgan.

*Materializado (as-code):* **motor** `arnesia conformance` (hexagonal, RulesetPort + ConformancePort + 5
adapters de mecanismo) · **fixture dogfood** `dogfood/dev-full-cycle.graph.json` valida **verde 13/13**
(4 contratos fusionados + spine-auto-consistente + estado-en-spine + transicion-legal + una-transicion-por-caja
+ spine-cobertura + fase-en-fases + escritor-unico + firewall) · gate **G1** (contrato de caja real verde con
los 3 ejes) · gate **G2** (`conformance <elemento>` da veredictos reales) · gate **G3** (`--todo` = 235 checks,
**0 fail, 0 error** = cero `enforced_by` colgante). **Loop conductor T3 REAL** (`usecase.BoxConductor` +
`domain.AvanzarCaja`/`RutaSiguiente` + FSM `caja_fsm.go`): el conductor Go dueña el loop, lee `result`+`status`
(NUNCA el texto del chat — garantía anti-scrape testeada), cap de reparación, `blocked→handoff`, ejecuta
`contract.ruta`. **Spike permisos por rol/TTL** (`domain.PermissionSet` + `permission.KitProvisioner`): el
MISMO tool decide distinto por rol, grants efímeros (`Grant.Vigente`). **3 boundaries SUBIDOS a `enforced`**
(orquestacion-determinista + permisos-derivan-del-rol + contrato-de-caja-es-fitness-function) con
`TestConductorOwnsBoxRouting` · `TestPermissionSetParametrizedByRole` · `TestBoxContractValidatesAgainstSchema`
pasando → **5 boundaries enforced** (2 HS-06 + 3 HS-08). Checks nuevos ejecutables: `escritor-unico`,
`spine-auto-consistente`, `no-arnesar⇒no-caja` (schema). **Auditoría final 2-frentes limpia** (código VERDE
9/9 blockers + docs VERDE tras cerrar residuos); **arXiv 2603.18916 verificado en web** (existe · 4 capacidades
textuales). Deps nuevas: `google/jsonschema-go`, `yaml.v3`. Doctrina intacta: **knowledge 12 nodos · 138 checks ·
arch 16 boundaries · 97 checks**. Sincronización mecánica de docs de prosa (VISION §gran plan · METODOLOGIA
§3/§5/§8 · CLAUDE Estado · INDEX de knowledge/arch · re-atribución DAOP-A2/A7 · 4 capacidades) en la misma ola.

*Conecta:* HS-07 (la doctrina v1 que esto hace ejecutable — contrato fusionado, nodo `harness-profile`, 2
boundaries) · HS-04/05 (`arch/` as code que el motor parsea) · HS-06 (conductor multisesión sobre el que
corre el routing de cajas T3) · la auditoría en `research/…auditoria-doctrina-v1-aplicabilidad.md`.

*Siguiente:* MVP del Mapa (ahora con el arnés dogfood real como base) · superficie HTTP `control_request` con
`role`/`ttl` (el modelo de permisos ya existe; falta el endpoint) · gate de fidelidad §8.4 ejecutable (hoy
`deferred`, ficha posterior) · 6 patrones de subagente ya detallados · restos multisesión (worktree · presupuesto · OTel).

### HS-09 · Fase 5 arrancada — MVP del Mapa (dogfood real), tras auditoría 5-frentes del estado — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-06):* "Puedes revisar con subagentes todo el estado actual del proyecto, hasta el
último detalle, y con ello, actualizar el plan? en cuanto a funcionalidades primero implementaría el mapa."

*Desarrollo:* apertura de la **fase 5 (Implementación)** del gran plan (fase 4 specs cerrada en HS-08).
Precedida por una **auditoría de 5 subagentes en paralelo, VERIFICADA con build/test real** (no lectura de
docs): (A) backend Go · (B) motor conformance + dogfood · (C) FE + shell Tauri · (D) doctrina as-code ·
(E) readiness del Mapa. **Veredictos clave:** backend **compila + vetea limpio, todos los tests PASS
(`-race`)**; el **shell Tauri v1 SÍ compiló y corre** (multisesión + CC real — corrige la nota stale de HS-06);
**React Flow 12 ya instalado** y `GET /api/harnesses/{id}/graph` **vivo**, pero el Mapa es **0% código**
(`workspace-stage.tsx` = `<ComingSoon>`) y el endpoint sirve un **demo de 2 nodos**, no el dogfood. Honestidad
del motor destapada: `conformance --todo` = **235 checks pero 211 `deferred`** (187 nl-judge + 3 go-arch-lint
sin binario/config + 5 schema/scan que solo corren en ruta `--arnes`) → **solo 24 pass real** (vía ~16 arch-tests);
la potencia determinista real vive **fuera de los 235**, en la ruta `--arnes` (**13/13 verde** en el dogfood, con
test adversarial). Real-pero-**no-cableado** al daemon: `BoxConductor` (loop T3) + `KitProvisioner` (permisos) —
falta adapter concreto `ArtifactReader`.

*Materializado (sincronización as-code, misma ola):* **commit de HS-08** (estaba íntegro en working tree, verde,
sin commitear sobre `1d6880d`). **3 drifts muertos:** CLAUDE.md ya no dice «cero código de producto aún» (HS-08
aterrizó ~1500 LOC) · METODOLOGIA §5 deja de enmarcar el linter como «fase 5 futuro» (ya construido, con la
salvedad de los `deferred`) · nota del shell actualizada. **Conteos intactos** (knowledge 12·138 · arch 16·97 ·
5 boundaries enforced). Sin código de producto nuevo aún — esta ficha **abre el plan**, no lo ejecuta.

*Plan (Mapa primero, dogfood-first):*
- **Hito 1 — Mapa read-only navegable del dogfood.** Backend: loader `dogfood/dev-full-cycle.graph.json` →
  índice bajo id `dev-full-cycle` (~30 LOC, **único cambio backend imprescindible**). FE: tipos TS del grafo
  (claves español exactas) + `api.getGraph(id)` · slice `entities/arnes` (store + selectores por banda/fase) ·
  `widgets/map-canvas` (`<ReactFlow>` + **layout de carriles custom** determinista: Guardia arriba · 1 carril
  por `arnes.fases[]` · Base abajo — React Flow NO da esto) · nodos custom por `clase` (tokens `var(--c-*)`,
  regla **canvas⊥chrome**) · swap del `ComingSoon` · story=test con el fixture dogfood.
- **Hito 2 — Inspector + picker.** Backend `getNode` (Box+contract) + `listHarnesses` (hoy 501/`[]`). FE:
  panel inspector S3 (el contrato fusionado ya lo alimenta) + click nodo→inspector. Solo capa **Estructura**;
  Tokens/Desempeño/Proceso esperan telemetría (indexer JSONL real).
- **Hito 3 (diferible):** realtime SSE `event: map` + indexer JSONL real (reemplaza el índice stub in-memory) +
  crear/editar sobre el lienzo (dock + nodos punteados).

> **Nota (2026-07-07, sync HS-10):** el `<ReactFlow>` del plan del Hito 1 quedó SUPERADO al
> ejecutarse: el sustrato firmado del Mapa es **HTML+SVG** (bandas/carriles + overlay SVG de
> edges — Gate 1 / Fase D de HS-09, commit `9cd8e77`); **React Flow 12 queda reservado al
> Organigrama** (lienzo libre 2D). El texto del plan de arriba se conserva como historia.

*Deuda paralela registrada (no bloquea el Mapa):* cablear `BoxConductor`+`KitProvisioner` al daemon (falta
`ArtifactReader`) · endpoint HTTP `control_request` con `role`/`ttl` · go-arch-lint (binario + `.go-arch-lint.yml`;
declarado en 3 boundaries, hoy inoperable) · convertir `deferred`→real cableando linters externos (biome/golangci/
tsc/dependency-cruiser) a CI · gate de fidelidad §8.4 · restos multisesión (worktree · presupuesto · OTel).

*Conecta:* HS-08 (el motor + el arnés dogfood real que el Mapa renderiza) · HS-05 (arch FE FSD-lite + canvas⊥chrome
+ tokens que el lienzo respeta) · HS-03 (UX firmada S2 lienzo/carriles/capas + S3 inspector) · mockups
`arnesia-mockup-v3.html` (superficie del Mapa) + `arnesia-shell-A-galaxia.html` (`mapView()` a portar).

*Siguiente:* ejecutar el **Hito 1** (loader Go → tipos+`getGraph` → `entities/arnes` → `widgets/map-canvas` con
carriles → swap del placeholder → story=test).

### HS-10 · Auditoría integral doctrina⇄app + enforcement reparado (CI verde de raíz) + nomenclatura de reconocimiento — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-07):* "quiero que revises y audites archivo por archivo de nuestra doctrina y me
indiques inconsistencias, puntos flacos […] revisa si está siendo bien implementada por nuestra aplicación y
esta la 'contiene' […] La aplicación debe ser instalada en cualquier computadora y siempre que tenga el usuario
instalado claude code, debería funcionar sin que el usuario tenga que hacer nada más." Y al recibir el informe:
"Ficha Propia YA" (nomenclatura) · "Arreglarlo ahora para probar cada cosa que avancemos" (CI).

*Desarrollo:* **auditoría 4-frentes en paralelo, VERIFICADA en vivo** (knowledge/ · arch/ · docs-norte ·
implementación; conformance ejecutado, schemas validados con el motor real, `gh run list` revisado).
**Veredictos:** doctrina estructuralmente sana y aritméticamente honesta (138+97=235 ✓ · 8 enforced ✓ · ambos
dogfoods validan 0 violaciones) pero con deriva de sincronización (CADENCE 2 fichas atrás · CLAUDE.md 2–3 olas ·
ficha HS-09 aún decía React Flow · METODOLOGIA §3 5 orígenes vs schema 7 · UX «11 elementos»). **La app NO
contiene la doctrina — 3 puentes bloqueantes:** (1) sin loader real de arneses (índice = demo + 2 `go:embed`
a mano), (2) doctrina no viaja a las sesiones CC (spawn sin `--plugin-dir`/`--append-system-prompt`), (3)
conformance atado al repo fuente + toolchain Go y sin endpoint. **Hueco doctrinal mayor:** la nomenclatura de
reconocimiento archivo→grafo NO estaba escrita en ningún doc (dogfoods armados A MANO; «nomenclatura» = 0 hits).
**Hallazgo estrella:** CI de main llevaba **≥5 pushes en rojo** (npm ci vs pnpm-lock → los enforcers FE JAMÁS
corrieron en CI; golangci con 66 hallazgos reales; go-arch-lint roto por construcción; step `openapi:gen`
fantasma; lefthook sin instalar) — TODO el enforcement era disciplina local.

**Ejecutado (mismo día):** ① `ci.yml` reescrito — pnpm (+`packageManager` en web/package.json) · go-arch-lint
con invocación correcta (`--project-path . --arch-file arch/fitness/.go-arch-lint.yml`) · guard honesto de
openapi-gen sobre el dir generado · **job rust ACTIVADO** (clippy verde verificado local + deps webkit del
runner + rust-cache) · playwright install para story-tests. ② `.go-arch-lint.yml` saneado a sintaxis v3 REAL
(era inválido: `cannotDependOn` no existe; allow-list + default-deny) + componentes nuevos (conformance ·
permission · dogfood) + `deepScan: false` justificado (flaggeaba al composition root inyectando concretos —
el patrón que la doctrina ORDENA). **Al arrancar cazó una violación real que arch_test.go no veía:** `usecase`
importaba el adapter concreto `conformance/mechanism` → fix: puerto `ports.SchemaValidator` + inyección desde
cmd (hexagonal restaurado). ③ golangci **66→0** (causa raíz; 8 `nolint:gosec` con razón concreta local-first;
godoc reales; nilerr ya no traga errores de WalkDir; contextcheck con `context.WithoutCancel`). ④ lefthook
instalado (hooks commit-msg + pre-commit vivos). ⑤ `arch/contracts/nomenclatura-arnes.md` **draft v0**: unidad
reconocible (plugin CC | arnés instalado) · manifiesto `arnes.l0.json` · tabla clase→ubicación (10
reconocedores) · derivación archivo→grafo · reconciliación honesta (`no-reconocido` visible, jamás crash) —
**PENDIENTE DE FIRMA** (D-a manifiesto · D-b instalado-primera-clase · D-c no-reconocido).

*FIRMADO (operador, 2026-07-07 — "Firmo las 4"):* **① estrategia de empaquetado (c)** — `go:embed` del
ruleset (conformance portable SIN contexto LLM) + doctrina como **plugin CC propio** inyectado por flags al
spawn con progressive disclosure (diseño 3-cuerpos de `research/2026-07-05-arquitectura-inyeccion-knowhow.md`,
que pasa de propuesta a VIGENTE; regla dura: ② jamás se escribe en el árbol de ③) · **② D-a** manifiesto
`arnes.l0.json` en la raíz del arnés · **③ D-b** arnés instalado (`.claude/` sin plugin.json) = ciudadano de
primera · **④ D-c** elemento no reconocido = nodo `no-reconocido` VISIBLE con warn. `nomenclatura-arnes.md`
v0→**v1 FIRMADA**. Decisión de scope tomada en el mismo acto: conformance en dos scopes — `fabrica`
(arch-tests, solo CI del repo) vs `arnes` (portable, nativo Go en el binario); los arch-tests NO se reescriben
para cliente.

*Siguiente:* ola de sync mecánico (CLAUDE.md · CADENCE ×2 · UX 11→12 · nota React Flow en ficha HS-09 ·
METODOLOGIA 5→7 orígenes · debate 3 VISION · tensión `alw` en proposals.ts) · los 3 puentes (= Hito 3+:
loader real por nomenclatura · inyección al conductor · conformance embebido + endpoint) · ErrorBoundary +
fallback `no-reconocido` en FE · migrar `dogfood/skills/` al layout L1.

### HS-11 · Los 3 puentes — la app CONTIENE la doctrina (ola de sync + loader + inyección + conformance portable) — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-07):* "Haz la ola de sync mecánico y HS-11 y todo lo que falte resolver de la
auditoría que has hecho." (Sobre las 4 firmas de HS-10 del mismo día.)

*Desarrollo:* **Ola de sync mecánico** (3 agentes en paralelo, verificada con el motor: 235 intacto):
knowledge/ (CADENCE des-staleado: 12 elementos + `académica` + runner HS-08 + filename-manda · rules.md
v1.2 retroactivo por L2.6 · anclas skills.md reparadas · harness-profile v1.1 con 2 checks CABLEADOS a
`box.contract.schema.json` → reparto real nl-judge 185 · arch-test 42 · schema 4 · go-arch-lint 3 ·
static-scan 1) · arch/ (97/138/235 en CADENCE · gen/README y conventions/INDEX sin "fase 5" · shell =
`web/src-tauri/` · go-arch-lint vivo en CI) · docs-norte (CLAUDE.md al día · METODOLOGIA 7 orígenes de
`necesita.de` + **§9 «Los 3 cuerpos» canónica** + supersesión luana · VISION debate 3 CERRADO · UX 12
elementos/10 clases · 2 notas aditivas en fichas HS-03/HS-09).

**Los 3 puentes (código, todo verificado E2E):**
**① Loader real por nomenclatura** — `internal/adapters/loader` implementa `nomenclatura-arnes.md` v1:
detección plugin/`.claude/` · manifiesto `arnes.l0.json` · frontmatter fusionado → `domain.Contract` ·
edges DERIVADOS (R1 `base:` → lee lector→conocimiento · R2 `caja:` → invoca; **`ruta[].a` NO genera
edges** — decisión doctrinal: el rework vive en contrato+spine, no como cableado del Mapa) ·
`no-reconocido` VISIBLE (nueva clase-marcador en schema+domain, NO 11ª primitiva) · CLI `arnesia index
[-o out] <dir>`. **El dogfood se volvió arnés REAL en disco**: `dogfood/dev-full-cycle/` (plugin.json ·
arnes.l0.json · 4 `skills/<id>/SKILL.md` con contrato fusionado completo · CLAUDE.md `std-spec`);
migrado el layout plano viejo; `fuente_path` ×5 estampados en el fixture. **Round-trip verificado:
dir → loader → graph.l0 → `conformance --arnes` = 13/13 PASS** (el firewall escanea los 5 fuentes reales).
**② Inyección de doctrina al conductor** — `kit/` REAL en el repo (plugin `arnesia-kit`: `doctrine.md`
overlay + skills `forjar-caja` y `auditar-arnes`) embebido (`all:kit`); `provision.Provisioner`
materializa `~/.arnesia/{kit,doctrine.md,knowhow}` idempotente por huella sha256 del contenido;
`SpawnOpts.Injection` → el conductor suma `--plugin-dir` / `--append-system-prompt-file` / `--add-dir`
(SIN `--bare`; suscripción intacta); fallo de provisión = spawn sin doctrina + warn (guía sin bloqueo);
**② jamás se escribe en ③** (METODOLOGIA §9). **③ Conformance portable** — paquete raíz `doctrina` con
`go:embed` de knowledge/+arch/(md+schemas); parser y SchemaSet refactorizados a `fs.FS` (una lógica, dos
fuentes); **scope `fabrica` vs `arnes`**: arch-test/go-arch-lint difieren honesto sin repo fuente; nuevo
`ConformancePort.RunGraph` + **endpoint `GET /api/harnesses/{id}/conformance`** (el botón de auditoría
del Mapa ya tiene sustrato). **Verificado fuera del repo**: binario en dir ajeno → 235 checks embebidos ·
`--arnes` 12/13 (firewall diferido honesto sin baseDir).

**Resto de la auditoría, resuelto:** ErrorBoundary FE (`shared/ui/error-boundary`) + banda desconocida →
región Base VISIBLE + 2 stories de crash/fallback (**suite 47/47 verde**) · descubrimiento de `claude`
multi-ruta para lanzamientos GUI (PATH de .desktop sin ~/.local/bin) · usage del daemon sin el "embedded
UI" fantasma · `proposals.ts` des-staleado (`origen` = campo L0 real que estampa el provisioner; `alw`
DERIVADO por decisión firmada, jamás campo) · openapi + path conformance · go-arch-lint con componentes
nuevos (doctrina · provision · loader) · conteo de story-tests des-fragilizado en fe-visual-fitness.

*Deuda que HS-11 deja registrada (no bloquea):* cablear el loader al índice del daemon (colisiona con el
WIP del Hito 2 en `index/store.go`; hoy es CLI + los fixtures embebidos siguen) · BoxConductor +
`control_request` (Fase E del plan Hito 2 del operador) · go:embed de la SPA en el daemon · bundle
instalador (GoReleaser + sidecar automático) · bajar los 3 boundaries del research de inyección
(`maquinaria-no-contamina-arnes` · `doctrina-una-fuente-dos-targets` · `telemetria-de-nacimiento`) a
nodos formales de arch/ en la próxima cadencia.

*Conecta:* HS-10 (las 4 firmas que habilitaron esto) · HS-08 (motor que ③ vuelve portable) · HS-07 (§8/
harness-profile que el kit encarna) · research inyección 2026-07-05 (FIRMADO, ahora implementado) ·
`arch/contracts/nomenclatura-arnes.md` v1 (spec de ①).

*Siguiente:* Hito 2 del Mapa (working tree del operador, gates 2..6) · loader→índice del daemon ·
instalador.

> **Nota (2026-07-07, mismo día — la deuda de arriba se PAGÓ):** el operador ordenó «termina lo que
> queda» y HS-11 cerró TODO su *Siguiente* técnico: **loader→índice VIVO** (`IndexPort.Upsert`; PUT
> /api/arneses/{id} reconoce el dir y lo indexa — «Cargar carpeta» funciona; smoke E2E verde) ·
> **Fase E COMPLETA según el plan firmado** (D2/D3/D4): `adapters/artifact` (lector de `status:`
> confinado) · `SpawnOpts.Permisos` → flags CC-native + reenvío/respuesta de `control_request` ·
> `POST /api/harnesses/{id}/boxes/{boxId}/run` (rol→set→BoxConductor.RunWith, SSE `event: run`) ·
> `POST /api/sessions/{id}/permission` REAL (deny>ask>allow · grants TTL · tarjeta al Dock) ·
> `TestWriteRequiresApproval`+`TestLiveEventsFromStreamJSON` flipados a reales · **instalador REAL**:
> SPA `go:embed` servida por el daemon (token solo /api|/events; Host+Origin en todo) ·
> `scripts/bundle.sh` · bundles producidos y el binario DEL PAQUETE probado E2E
> (`ArnesIA_0.1.0_amd64.deb` 7.6M con sidecar fresco de hoy · `.AppImage` 80M · `.rpm`). El WIP del
> Hito 2 (Fase 1 showcase + Fase 2 parcial) se commiteó a main por orden «pon todo en main». Deuda
> honesta restante: spike del wire format `control_response` contra claude real · run síncrono
> (async+202 futuro) · gate de conformance no auto-invocado tras run · telemetría/indexer JSONL ·
> codegen gen/ · 3 boundaries del research a arch/ formal.
>
> **G-instalación VERIFICADO EN MÁQUINA REAL (2026-07-07):** `sudo dpkg -i ArnesIA_0.1.0_amd64.deb`
> por el operador → app GUI lanzada (shell Tauri + sidecar `/usr/bin/arnesia-daemon` spawneado con
> token) → `healthz` ok · **UI servida por el daemon del paquete** · API confinada (401 sin token) →
> stamp de doctrina BORRADO a propósito y **re-provisionado por el binario instalado**
> (`~/.arnesia/{kit,doctrine.md,knowhow}` frescos) → sesión CC REAL sobre el arnés dogfood con la
> doctrina inyectada por flags: turno «Responde únicamente: OK» → **respuesta `OK` de
> `claude-fable-5`**, `claude_session_id` persistido, sesión `idle` resumible. El constraint de UX
> («el usuario solo necesita Claude Code instalado y logueado») quedó DEMOSTRADO end-to-end.

### HS-12 · Interop DevStudio ⟷ ArnesIA — el eslabón de consumo del ecosistema, ratificado — `decidida` · `vig:vigente`

*Cruda (operador, 2026-07-07):* entrega el prompt de interop del agente de DevStudio (repo
`~/Proyectos/dev-studio`, ficha DH-18/PB-25): "Interop DevStudio ⟷ ArnesIA — ratificar el eslabón del
ecosistema y 4 pedidos aditivos… responder y, si se firma algo, fichar en el LEDGER de acá."

*Contexto:* DevStudio (la "app de rol" del diagrama de VISION) entregó su registry de arneses y ADOPTÓ el
estándar ArnesIA sin pedir cambios de formato — verificado EN VIVO consumiendo el dogfood real
`dev-full-cycle` (forma-plugin de nomenclatura v1 tal cual · marketplace git formato prenter-marketplace ·
inyección por `--plugin-dir` + `--append-system-prompt` replicando METODOLOGIA §9 · lock as-code
`.devstudio/arneses.yaml` como único artefacto committeado al repo del cliente). Primer consumidor externo
real del estándar. Dato de campo aportado: la CLI rechaza `--append-system-prompt` y
`--append-system-prompt-file` juntos («use only one»).

*Desarrollo — 4 firmas (operador, 2026-07-07) + 1 reparación:*
**P1 publish RATIFICADO** — fase 5 emitirá el formato prenter-marketplace vigente; contrato estable:
`marketplace.json` (layout CC oficial) · `plugins/<id>/<versión>/` forma-plugin (nomenclatura, solo
aditivo) · `catalogo.json` `{canales, versiones[].{version,estado,fuente,fecha}}`; fase 5 solo SUMA
(metadata evals-gate, granularidad por-arnés). **P2 `spine.categorias` ACEPTADO** — mapa hermano opcional
estado→categoría, enum FIJO de 5 idéntico a I-77 RN-28 (propuesto·en-progreso·completado·descartado·
pausado, patrón Azure); terminalidad DERIVADA (∈{completado,descartado}); no viola agnosticismo: los
estados siguen dato per-arnés, la categoría es capa semántica del producto (estatus de `clase`).
**P3 `nombre`/`descripcion` — REPARACIÓN**: DevStudio cazó inconsistencia nuestra (nomenclatura §2 lo
nombraba, el schema lo rechazaba por `additionalProperties:false`); reparado + cadena de fallback
BENDECIDA: `arnes.l0.nombre` → `plugin.json name` → `id`. **P4 lock BENDECIDO** — `.devstudio/arneses.yaml`
= detector 3° de nomenclatura (v1.1): puntero de descubrimiento read-only → N arneses en forma-plugin vía
caché/marketplace; entrada no resoluble = check rojo visible; pedido recíproco: DevStudio declara estables
`id·versión·canal·registry`. **P5 postura I-77 FICHADA** — spine(+categorias) = subconjunto navegable
canónico; gates/dueños se DERIVAN de contratos por caja (`estado`/`gate`/`ruta` del frontmatter fusionado);
el arnés **NO shipeará descriptor I-77 aparte** (segunda fuente de verdad); un I-77 materializado será
PROYECCIÓN/export generada del arnés.

*Cementado (mismo turno, todo verde):* `graph.l0.schema.json` (+`nombre`/`descripcion`/`spine.categorias`)
· nomenclatura-arnes.md **v1.1** · `domain.Categoria` (Valid/Terminal) + `Spine.Categorias` + 2 checks
warn `categoria-estado-existe`/`terminal-categoria-coherente` (diferido honesto sin mapa) + test · dogfood
manifiesto+fixture en sync · FE `types.ts`. Round-trip `arnesia index` → `conformance --arnes` =
**15/15 PASS** (13 + 2 nuevos) · suite Go ✓ · tsc/biome ✓. La semilla `~/Proyectos/marketplace-arneses`
NO se mutó (0.1.0 inmutable; los campos viajan en la próxima versión publicada). Paquete:
`research/2026-07-07-interop-devstudio/` (decisiones + respuesta entregable a DevStudio).

*Deuda registrada (no bloquea):* el loader aún no implementa el detector 3° (leer lock → resolver
caché/marketplace) — entra cuando exista un proyecto DevStudio real que auditar · ficha recíproca de
DevStudio declarando estables los campos del lock.

*Conecta:* HS-10 (nomenclatura v1 que DevStudio adoptó) · HS-11 (dogfood real + inyección §9 que DevStudio
replicó) · I-77 (contrato de ecosistema, monorepo `tooling/strategy`) · DH-18/PB-25 (la ficha gemela en
dev-studio) · KIT-06 (release train del publish).

*Siguiente:* entregar `respuesta-devstudio.md` al agente de DevStudio · publish real fase 5.

<!-- Próximas: HS-13, … -->

## Log

| Fecha | Decisión | Fichas |
|---|---|---|
| 2026-07-04 | Fundación del repo propio: graduación de P4 (2ª graduación de la incubadora, precedente DevHub/I-78). Repo nace limpio con visión POR FORJAR (mutó — se firma aquí como HS-02); prefijo nuevo HS-NN; célula `products/harness-studio/` E instancia 0 `tooling/harness-studio/` congeladas como fuente del port gradual; kit dev como plugin del marketplace, canal estable. | HS-01 |
| 2026-07-04 | Visión v3 forjada y FIRMADA: **ArnesIA**, fábrica de arneses por rol × proceso. Constitución de 11 principios; Mapa = lienzo único; grafo agnóstico + adaptador CC; captura local-first + Langfuse espejo; git marketplaces + telemetría de efectividad; stack Go/Vite/React Flow/SQLite. Gran plan 6 fases — siguiente: UX (HS-03). | HS-02 |
| 2026-07-05 | Fase 2 (UX) FIRMADA en it.13: **shell = Command Rail (A)** (rail iconos izq. + visual casi-fullscreen + chat dock ⌘K) · Portafolio 2 lentes (Organigrama ↔ Cuadrícula) · 13 iteraciones sobre el lienzo · retiro andamiaje REAL/DEMO. Cementa `UX.md` + `METODOLOGIA.md` como **docs VIVOS** (excepción declarada a firmado=congelado). Base de evidencia = `knowledge/` (11 nodos · 122 checks). | HS-03 |
| 2026-07-05 | Fase 3 (arquitectura) arrancada: investigación 5-frentes verificada → stack completo (Tauri 2, subproceso-conductor stream-json, modernc SQLite, SSE, React Flow 12+Zustand, **AG-UI+assistant-ui+CodeMirror6**, schema-first Go+TS, go-arch-lint). Fork firmado: **shell = Tauri 2 desde v1** (divergencia declarada vs browser-first del research). **Arquitectura as code materializada:** árbol `arch/` (7 boundaries · 29 checks · model/ · contracts/ · fitness/) espejando `knowledge/`. Corrección `--bare`/auth propagada a knowledge (→122 checks). | HS-04 |
| 2026-07-05 | Fase 3 completada al FE: investigación 5-frentes (FSD · atomic · storybook · convenciones · tokens) verificada → **arquitectura FE as code**. Veredictos: **FSD-lite** + dependency-cruiser · 6 capas direccionales (canvas⊥chrome) + **shadcn/Base UI** · **Storybook 10** story=test (Chromatic descartado) · **golangci-lint v2 + Biome v2.4 + tsc strictest + lefthook** · **DTCG→Style Dictionary v5→Tailwind v4**. Materializado: **5 boundary nodes FE (22 checks) + `arch/conventions/` (8 nodes · 26 checks)** + config files declarados → **arch/ = 77 checks**. Specs corren a HS-06. | HS-05 |
| 2026-07-05 | Auditoría del shell Tauri v1: aislamiento CC por-tab CONFIRMADO real; 7 huecos → **endurecimiento multisesión + confinamiento local**. Forks firmados: **S2 registro explícito arnés→ruta** (WorkdirResolver, cwd por sesión, nunca global) · **S1 shell emite el token** (Host+Origin+token, adiós CORS `*`). Materializado (nace **enforced**, fitness tests PASAN `-race`): **2 boundaries** (superficie-local-confinada · sesion-viva-consistente) + 2 checks a permisos-gui + `--max-turns` + OpenAPI reconciliado + `store` mapeado en go-arch-lint → **arch/ = 14 boundaries · 89 checks**. Shell Rust escrito, no compilado aquí. Specs + spike de permisos → HS-07. | HS-06 |
| 2026-07-05 | **Doctrina propia v1**: cruce de DAOP v0.2 (BMAD+Agent SDK, 5 subagentes) + barrido de 7 fuentes externas (4 subagentes). Reencuadre clave: **operacionalizamos Agentic BPM** (manifiesto *Information Systems* 2026) — doctrina PROPIA basada en proceso e independiente de rubro, **no clon de BMAD**. Regla de Rosetta (DAOP-Arnés→CAJA), firewall CC-native (`no-phantom-frontmatter`), vocabulario framed-autonomy. 3 decisiones firmadas: producto-puro+META · dogfood-first · P6-por-scope. **As-code:** VISION §Linaje · METODOLOGIA §3 contrato fusionado + §8 doctrina de proceso · **nodo 12 `harness-profile`** → knowledge 12 nodos·138 checks · 2 boundaries → arch 16·97. Specs → HS-08. | HS-07 |
| 2026-07-05 | **Fase 4 specs (dogfood-first): doctrina BAJADA A EJECUTABLE.** Contrato de caja fusionado con diente (`box.contract.schema.json` + `domain.Contract`, 3 ejes, `TestBoxContractValidatesAgainstSchema` verde) · `Clase` canónica de 10 primitivas · manifiesto del arnés en `graph.l0` (`fases`·`spine` declarado·META rol·proceso·reporta_a·empresa). **Motor `arnesia conformance` construido** (hexagonal: RulesetPort parsea knowledge/+arch/ = 235 checks a datos; ConformancePort + adapters arch-test/schema-validation/go-arch-lint/static-scan/nl-judge; consistencia de spine parametrizada por el spine declarado). Fixture dogfood `dev-full-cycle.graph.json` verde; gates G1 (schema) + G2 (spine); stubs conductor/permisos aterrizan los `enforced_by:`. Conteos intactos (knowledge 12·138 · arch 16·97). Deps: google/jsonschema-go, yaml.v3. | HS-08 |
| 2026-07-06 | **Fase 5 (Implementación) arrancada: MVP del Mapa, dogfood-first.** Auditoría de 5 subagentes VERIFICADA con build/test real: backend compila+tests PASS `-race`; **shell Tauri v1 confirmado compilado+corriendo** (corrige nota HS-06); React Flow 12 instalado + endpoint del grafo vivo, pero **Mapa 0% código** y sirve un demo (no el dogfood). Honestidad del motor: `--todo` = 235 checks pero **211 `deferred`** (24 pass real); el enforcement determinista vive en la ruta `--arnes` (13/13 verde). **HS-08 commiteado** (estaba verde sin commit) + 3 drifts muertos (cero-código, linter-fase-5, shell-no-compiló). **Plan Mapa:** Hito 1 read-only navegable (loader Go dogfood→índice + `entities/arnes` + `widgets/map-canvas` carriles custom) · Hito 2 inspector+picker · Hito 3 realtime/edición diferido. Deuda paralela registrada (cablear conductor/permisos, `control_request`, go-arch-lint, deferred→CI). | HS-09 |
| 2026-07-06 | **Gate 1 (mockup) + Gate 2 (spec) del Mapa FIRMADOS; auditoría destapa inversión de orden → retro-ajuste.** Gate 1 (mockup `arnesia-mapa-mvp.html`, commit `0736d2c`) firmado con 6 decisiones doctrinales (Base canónica · activación por-nodo · facet `origen` · knowledge as-code+semántico-opcional · discovery=data · paquete/spine). **Paquete Fase C** redactado (`research/2026-07-06-mapa-mvp/`: `spec`=QUÉ/RF+Gherkin trazado a `mockup:línea`+shot · `design`=UI al pixel · `architecture`=CÓMO hexagonal+FSD + secuencia de PRs + cementado · `PARIDAD`=round-trip) y **Gate 2 FIRMADO**. Corrige dato stale: seed = 3 nodos/1 edge keyed «demo» (no 2) → `dev-full-cycle` da **404** hoy. **Auditoría 5-subagentes (build/test real) revela INVERSIÓN DE ORDEN:** se programó adelantándose a los gates — Fase F (Hito 1: loader dogfood + `<MapCanvas>` montado + inspector/picker de Hito 2) YA en working-tree y **verde** (go/tsc/biome/depcruise/steiger/stylelint), SIN cerrar Fase D (arch as-code) ni Fase E (deuda), Fase C sin firmar. Código **doctrinalmente limpio** (backend agnóstico, canvas⊥chrome, tokens, PROPUESTAs aisladas+etiquetadas) PERO doctrina as-code NO refleja las 6 decisiones + `fe-visual-fitness` flipado a `enforced` apuntando a enforcer **BORRADO** (`vitest.workspace.ts`) = **pass fabricado**. **Operador ordena RETRO-AJUSTAR:** Gate 2 ✓ → **Fase D real** (cementar 6 decisiones + arreglar `enforced` falso + cuerpos boundaries + re-sync conteos + C4/contratos + graduar `origen`/`alw` a L0) → **recién ahí** commitear código como Fase F. Deuda E **diferida honesta**. | HS-09 |
| 2026-07-06 | **Retro-ajuste EJECUTADO: Fase D (arq as-code) + Fase F (código) en orden — Hito 1 del Mapa VIVO.** **Fase D** (commit `9cd8e77`): 🚩 corregido el pass FABRICADO (`fe-visual-fitness` enforced apuntaba a `vitest.workspace.ts` BORRADO → migrado a `vitest.config.ts`; verificado **40 story-tests verdes**); 3 boundaries FE proposed→**enforced** con enforcers verificados corriendo (`fe-taxonomia-componentes`/canvas⊥chrome = depcruise 76 mód/0 viol · `fe-tokens-contrato`/stylelint + kind 6→10 regenerado · `fe-visual-fitness`) → **8 enforced** (2 HS-06 + 3 HS-08 + 3 HS-09; sin sumar checks, 97 intacto). Las 6 decisiones del Gate 1 cementadas as-code: **#3 `origen`** = campo nuevo en `graph.l0.schema.json` (`$defs.nodo`, `estandar\|del-puesto`, ESTAMPADO al provisionar la instancia ③ — no maquinaria filtrándose, 3 cuerpos); **#2 `alw`** = DERIVADO de `fuente_path` en `rules.md` (sin campo nuevo); #1/#4/#5/#6 ya doctrina. C4 (`container.d2`) + propagación de la reversión firmada **HTML+SVG** del Mapa a VISION.md/CLAUDE.md (supera «React Flow para el Mapa» → RF = Organigrama) + tabla del gran plan al día (fase 4 ✓, fase 5 en curso). **Fase F** (commits `31a7532` backend + `c013dc3` frontend): loader del dogfood real (`dev-full-cycle`, 5 nodos/4 edges) al índice + `getNode`/`listHarnesses` cableados + `<MapCanvas>` HTML+SVG montado (swap del `ComingSoon`) + seed de la 1ª sesión al Mapa real. **Todo verde**: go build/vet/test `-race` · tsc strictest · biome · depcruise · steiger · stylelint · 40 story-tests (Playwright Chromium). Capas Tokens/Desempeño/Proceso staged (telemetría JSONL → Hito 3); `origen`/`alw` se dibujan PROPUESTA hasta que provisioning/telemetría los pueble. **Gate `G-hito1` VERIFICADO EN VIVO** (daemon `arnesia serve` :4200 + vite :5173 → browser Chromium): `GET /api/harnesses/dev-full-cycle/graph` sirve el dogfood real (5 nodos/4 edges, 200) y la vista Mapa lo RENDERIZA — Guardia «— sin hooks —» · 4 carriles spec/build/review/release con 1 caja c/u (handle `/spec-writer`… + transición del spine ◇ idea→spec…review→released) + edges invoca (backbone rojo) · Base con la regla `std-spec` (0 siempre/1 condicional) · capas Tokens/Desempeño/Proceso disabled «necesita telemetría» · **consola 0 errores/warnings**; screenshot revisado. Nota: el daemon NO sirve la SPA embebida aún (skeleton TODO); en dev la carga vite, en prod la cargará Tauri. **Deuda Fase E (conductor/permisos/`control_request`/go-arch-lint/`deferred`→CI) sigue diferida honesta.** | HS-09 |
| 2026-07-07 | **Auditoría integral doctrina⇄app (4 frentes, verificada en vivo) + enforcement REPARADO DE RAÍZ.** Doctrina sana y aritmética honesta (138+97=235 ✓ · 8 enforced ✓ · dogfoods validan 0 viol.) pero **la app NO la contiene** (3 puentes bloqueantes: loader real · inyección a CC · conformance portable+endpoint) y **CI de main llevaba ≥5 pushes ROJO** sin bloquear nada (npm vs pnpm → enforcers FE jamás corrieron en CI · 66 hallazgos golangci · go-arch-lint roto por construcción · `openapi:gen` fantasma · lefthook sin instalar). **Reparado el mismo día:** ci.yml → pnpm + job rust activado (clippy verde) + guard honesto openapi · `.go-arch-lint.yml` a sintaxis v3 real (`cannotDependOn` no existía; deepScan off justificado) — **al arrancar cazó violación real**: `usecase` importaba `conformance/mechanism` → puerto `ports.SchemaValidator` + inyección desde cmd · golangci **66→0** (causa raíz, 8 nolint:gosec razonados) · lefthook vivo. **Hueco doctrinal mayor destapado:** la nomenclatura de reconocimiento archivo→grafo no estaba escrita → `arch/contracts/nomenclatura-arnes.md` **draft v0 PENDIENTE DE FIRMA** (plugin CC \| instalado · `arnes.l0.json` · 10 reconocedores · `no-reconocido` visible). Pendiente firma: D-a/D-b/D-c + empaquetado doctrina (rec.: embed ruleset + plugin propio, 3 cuerpos). | HS-10 |
| 2026-07-07 | **4 firmas del operador ("Firmo las 4") — la ruta de los puentes queda decidida.** ① **Empaquetado (c) embed+plugin** (3 cuerpos VIGENTE: `go:embed` ruleset → conformance portable sin contexto LLM · kit/doctrina materializados en `~/.arnesia/` e inyectados por flags al spawn · ② jamás se escribe en ③) · ② **D-a** manifiesto `arnes.l0.json` en la raíz · ③ **D-b** arnés instalado = ciudadano de primera · ④ **D-c** `no-reconocido` visible con warn. `arch/contracts/nomenclatura-arnes.md` **v1 FIRMADA**; research de inyección estampado VIGENTE. Scope conformance: `fabrica` (CI del repo) vs `arnes` (portable). *Siguiente:* ola de sync mecánico + **HS-11 = los 3 puentes** (loader por nomenclatura · inyección al conductor · conformance embebido + endpoint). | HS-10 |
| 2026-07-07 | **HS-11 EJECUTADO — la app CONTIENE la doctrina: ola de sync + los 3 puentes, todo E2E.** Sync (3 agentes, 235 intacto): CADENCE×2 des-staleados · rules v1.2 retroactivo · harness-profile v1.1 (2 checks→schema-validation) · METODOLOGIA 7 orígenes + **§9 «3 cuerpos»** · VISION debate 3 cerrado · UX 12/10 · CLAUDE.md al día. **① Loader real** (`internal/adapters/loader`, nomenclatura v1; edges derivados R1-lee/R2-invoca, `ruta` NO cablea; `no-reconocido` visible; `arnesia index <dir>`) + **dogfood = arnés REAL** (`dogfood/dev-full-cycle/` plugin-form) → **round-trip dir→grafo→conformance 13/13 PASS**. **② Inyección** (kit `arnesia-kit` embebido → `~/.arnesia` por huella → `--plugin-dir`/`--append-system-prompt-file`/`--add-dir`; ②↛③). **③ Conformance portable** (paquete raíz `doctrina` go:embed · parser/schemas a fs.FS · scope fabrica\|arnes · `RunGraph` + **endpoint `GET /api/harnesses/{id}/conformance`**) → verificado fuera del repo: 235 embebidos · --arnes 12/13 honesto. Extra: ErrorBoundary+banda-fallback FE (47/47) · claude multi-PATH GUI · proposals.ts honesto · clase-marcador en schema. Deuda registrada: loader→índice daemon (WIP Hito 2) · BoxConductor/control_request (Fase E operador) · SPA embed · instalador · 3 boundaries research→arch/. | HS-11 |
| 2026-07-07 | **HS-11 cierre total («termina lo que queda»): loader→índice VIVO + Fase E COMPLETA + instalador REAL.** WIP Hito 2 del operador a main («pon todo en main»). `IndexPort.Upsert` + carga al registrar y al boot — «Cargar carpeta» E2E verde. Fase E según plan firmado: `adapters/artifact` · `SpawnOpts.Permisos`→flags CC-native · `control_request` reenviado/respondido (Dock, D3) · `POST …/boxes/{boxId}/run` (D2) · permission REAL con grants TTL (deny>ask>allow) · 2 arch-tests flipados de skip a reales · 4 changelogs «realizado en vivo». Instalador: SPA go:embed servida por el daemon (token solo API) · `scripts/bundle.sh` · `.goreleaser.yaml` · **bundles producidos con el daemon del día**: `.deb` 7.6M (binario del paquete probado E2E: UI + conformance embebidas) · `.AppImage` 80M (fix bundle.icon) · `.rpm`. Todo verde: race+lint+arch-lint+fmt. Deuda honesta: spike control_response vs claude real · run async · gate post-run · telemetría · codegen · 3 boundaries research. | HS-11 |
