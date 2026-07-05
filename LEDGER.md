# Ledger — Harness Studio (fichas HS-NN)

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
+ señal, `enforced_by:` 1:1) — el gemelo-arquitectura de los 123 de knowledge; **`model/`** (C4
contexto Mermaid + container D2); **`contracts/`** (schemas L0 `meta.clase` + `contract:` de caja —
el mismo de METODOLOGIA §3 — + OpenAPI 3.1, con `gen/` para tipos Go+TS); **`fitness/`**
(`.go-arch-lint.yml` + `arch_test.go`). Un solo runner futuro `arnesia conformance` corre arch +
knowledge. **Corrección load-bearing propagada** al nodo `knowledge/headless-sdk` (v1.0→v1.1):
`--bare` **rompe el auth de suscripción** (no default-earlo en el conductor) + `%contexto` = métrica
derivada (no existe en stream-json/OTel) → 123 checks totales.

**Honestidad:** cero código de producto aún (fase 5); los 29 checks están **declarados, no
corriendo** (`status: proposed`) — se activan cuando el módulo Go aterrice, igual que los 123 de
knowledge son el linter futuro. Flags de verificación registrados en el research doc (shape
JSON-RPC de `control_request` = primer spike de fase 4; contención `~/.claude` = load-test;
macOS+Keychain; ToS si se distribuye).

*Conecta:* HS-02 (decisiones técnicas fundacionales que esto aterriza) · HS-03 (UX firmada que la
arquitectura debe servir) · I-75 (contrato L0 `meta.clase` = el schema `graph.l0`) · I-76/OBS-16/
OBS-18 (patrón conductor, hecho subproceso-stream-json) · KIT-03 (`telemetry-emit` OTel = el
sidechannel del conductor) · `knowledge/` (el árbol gemelo; `arch/` copia su patrón L1/L2/checks) ·
METODOLOGIA §3 (`contract:` de caja = `box.contract.schema.json`).

*Siguiente:* **HS-05 = fase 4, specs** — cementar los schemas del dominio (fase/estado spine,
`meta.clase` por-formalizar), el spike del protocolo de permisos contra el binario instalado, y las
specs del MVP (Mapa primero). Debates de VISION que fase 4 hereda: segundo cerebro (vector DB) ·
config de marketplace por proyecto · renombre repo → `arnesia`.

<!-- Próximas: HS-05, … -->

## Log

| Fecha | Decisión | Fichas |
|---|---|---|
| 2026-07-04 | Fundación del repo propio: graduación de P4 (2ª graduación de la incubadora, precedente DevHub/I-78). Repo nace limpio con visión POR FORJAR (mutó — se firma aquí como HS-02); prefijo nuevo HS-NN; célula `products/harness-studio/` E instancia 0 `tooling/harness-studio/` congeladas como fuente del port gradual; kit dev como plugin del marketplace, canal estable. | HS-01 |
| 2026-07-04 | Visión v3 forjada y FIRMADA: **ArnesIA**, fábrica de arneses por rol × proceso. Constitución de 11 principios; Mapa = lienzo único; grafo agnóstico + adaptador CC; captura local-first + Langfuse espejo; git marketplaces + telemetría de efectividad; stack Go/Vite/React Flow/SQLite. Gran plan 6 fases — siguiente: UX (HS-03). | HS-02 |
| 2026-07-05 | Fase 3 (arquitectura) arrancada: investigación 5-frentes verificada → stack completo (Tauri 2, subproceso-conductor stream-json, modernc SQLite, SSE, React Flow 12+Zustand, **AG-UI+assistant-ui+CodeMirror6**, schema-first Go+TS, go-arch-lint). Fork firmado: **shell = Tauri 2 desde v1** (divergencia declarada vs browser-first del research). **Arquitectura as code materializada:** árbol `arch/` (7 boundaries · 29 checks · model/ · contracts/ · fitness/) espejando `knowledge/`. Corrección `--bare`/auth propagada a knowledge (→123 checks). | HS-04 |
