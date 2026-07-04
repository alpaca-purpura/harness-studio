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

<!-- Próximas: HS-03, HS-04, … -->

## Log

| Fecha | Decisión | Fichas |
|---|---|---|
| 2026-07-04 | Fundación del repo propio: graduación de P4 (2ª graduación de la incubadora, precedente DevHub/I-78). Repo nace limpio con visión POR FORJAR (mutó — se firma aquí como HS-02); prefijo nuevo HS-NN; célula `products/harness-studio/` E instancia 0 `tooling/harness-studio/` congeladas como fuente del port gradual; kit dev como plugin del marketplace, canal estable. | HS-01 |
| 2026-07-04 | Visión v3 forjada y FIRMADA: **ArnesIA**, fábrica de arneses por rol × proceso. Constitución de 11 principios; Mapa = lienzo único; grafo agnóstico + adaptador CC; captura local-first + Langfuse espejo; git marketplaces + telemetría de efectividad; stack Go/Vite/React Flow/SQLite. Gran plan 6 fases — siguiente: UX (HS-03). | HS-02 |
