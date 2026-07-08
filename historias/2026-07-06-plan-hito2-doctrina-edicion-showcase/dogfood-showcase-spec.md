# Dogfood showcase — `content-studio-full` (kitchen-sink)

> Ficha HS-09 (fase 5) · plan Hito 2 · 2026-07-06
> Artefacto: [`showcase.graph.json`](./showcase.graph.json) — **BORRADOR**, se valida con `arnesia conformance --arnes` en Fase 1 antes de promover a `dogfood/`.

## Por qué un segundo arnés (decisión firmada)

`dev-full-cycle` es el arnés **real y honesto** de nuestro propio proceso de desarrollo — debe seguir siéndolo (5 nodos, 2 clases, lo que de verdad usamos hoy). Meterle casuística falsa lo volvería demostrativo y mentiroso (viola VISION P5 «autodocumentación como efecto» y la regla de honestidad de METODOLOGIA §4).

Por eso el showcase es **un arnés aparte**: `content-studio-full`. El picker del Mapa alterna entre ambos. El showcase existe para **ver el Mapa en su esplendor** — ejercita cada valor de cada enum del L0.

## Agnóstico a rubro — a propósito

El rol es **Editorial · Content Lead** (NO ingeniería). Prueba en vivo el principio VISION P7 (agnóstico a rubro): las primitivas son CC-native, el proceso es editorial. Si el Mapa lee bien un pipeline de contenidos, lee bien cualquier rubro.

## Matriz de cobertura (verificada por script)

| Eje | Cardinalidad | Valores ejercidos | Nodos testigo |
|---|---|---|---|
| `clase` | **10/10** | skill · subagent · hook · rule · command · mcp · plugin · settings · output-style · statusline | cajas+expertos · fact/tone · guardia · base · publish-now/legacy-cms · db/store/index/search/analytics · content-kit · role-permissions · product-style · arnes-health |
| `banda` | **7/7** | guardia · fase · base · libreria-expertos · meta-harness · terceros · marcas-dormidas | hooks · cajas · reglas+knowledge · seo/legal · config · web/analytics · legacy-cms |
| `canal` | **4/4** | beta · estable · propuesto · deprecado | research/draft/measure · brief/edit/review/publish · semantic-index · legal-checklist/legacy-cms |
| `procedencia` | **5/5** | medido · estimado · declarado · inferido · no-declarado | brief/edit/publish · research/draft · review/tone · measure · legal-checklist/legacy-cms |
| `arquetipo` | **4/4** | pipeline · excepcion · abierto · no-arnesar | review/publish/measure · brief/edit · research/draft · tone-guardian (`caja:false`) |
| `perfil_harness` | **3/3** | T1 · T2 · T3 | review/publish/measure · brief/edit · research/draft |
| `gate.tipo` | **4/4** | auto · manual · parcial · **none** | edit/publish · brief/review · research · **draft/measure (huecos honestos)** |
| `origen` | **2/2** | estandar · del-puesto | mayoría · publish-caja / tone / publish-now (borde punteado) |
| `edge.tipo` | **3/3** | invoca · lee · escribe | spine backbone + retrabajo · lee reglas/experto · escribe a mcp/terceros |
| `necesita.de` | **7/7** | usuario · base: · caja: · libreria: · maquinaria: · terceros: · marcas-dormidas: | brief · brief · research · research · edit (lint-runner) · research/measure · review (legacy) |
| `escritor_unico` | **true + false** | 1 escritor · escritor compartido | mayoría · measure→`metrics.log` (append) |
| spine | forward + **retrabajo** | 7 transiciones forward + 1 loop `editado→borrador` | review→draft y edit→draft (return edges) |

## Casuísticas doctrinales que el showcase pone a prueba (VISION Anatomía A1–A7)

- **A2 contrato input→output** — cada caja declara `necesita`/`entrega`; el hand-off ES el contrato.
- **A3 dos estados** — cada caja posee UNA transición del spine (`estado`), visible en el tag `◇`.
- **A4 eval-gate en el contrato de salida** — `gate.tipo:none` en `draft-caja` y `measure-caja` = **hueco honesto visible** (no se pinta verde lo que es gris).
- **A5 la fábrica no es recta** — retrabajo (`edit→draft`, `review→draft`) como return edges + `ruta.si`.
- **A6 infra compartida en bandas** — Guardia (3 hooks) y Base (reglas + knowledge) fuera de las fases.
- **Honestidad de superficie fría** (METODOLOGIA §5) — `legal-checklist` (deprecado, no-declarado) y `legacy-cms` (marcas-dormidas, 0 corridas) = hallazgos, no ruido.
- **`del-puesto` vs `estandar`** — `publish-caja` es el paso a medida de este puesto (borde punteado); el resto es estándar del arnés.

## Cierre de Fase 1 (2026-07-06) ✅

- ✅ Validado: `arnesia conformance --arnes dogfood/content-studio-full.graph.json` → **16 checks · 15 pass · 1 deferred** (firewall diferido honesto). Schema L0 + 7 contratos de caja + spine + escritor-único, todo verde.
- ✅ Promovido a `dogfood/` + embed + seed; render live verificado ojo-UI (`scratchpad/showcase-live.png`).
- ✅ Limpieza: quitado `_meta` (root `additionalProperties:false`) y `alw` (dato muerto — no es campo L0; el FE lo deriva de `proposals.ts`, ruling Fase D).
- ⏳ **Diferido a Fase 2:** fixture FE `content-studio-full.ts` + stories (las marcas doctrina-visibles necesitan story=test; el render live ya cubre la verificación del dato). Y cablear `origen`/split-reglas desde el dato (hoy FE-fixture-backed).
