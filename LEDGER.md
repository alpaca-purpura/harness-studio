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

<!-- Próximas: HS-02, HS-03, … -->

## Log

| Fecha | Decisión | Fichas |
|---|---|---|
| 2026-07-04 | Fundación del repo propio: graduación de P4 (2ª graduación de la incubadora, precedente DevHub/I-78). Repo nace limpio con visión POR FORJAR (mutó — se firma aquí como HS-02); prefijo nuevo HS-NN; célula `products/harness-studio/` E instancia 0 `tooling/harness-studio/` congeladas como fuente del port gradual; kit dev como plugin del marketplace, canal estable. | HS-01 |
