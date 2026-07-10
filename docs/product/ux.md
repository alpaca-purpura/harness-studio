# ArnesIA — UX del producto (fase 2 del gran plan, ficha HS-03)

> **FIRMADA** (HS-03, it.13, 2026-07-05). Por **excepción declarada** (ver METODOLOGIA §7 y ficha
> HS-03 del LEDGER), la firma cierra la **fase** pero este doc queda **VIVO**: sigue creciendo con
> nuevas iteraciones (no se congela). Norte: [`vision.md`](vision.md) v3. Mockup v1 del Mapa =
> herencia validada en esencia (HS-02).
>
> **Sync 2026-07-07 (HS-10, edición mecánica — doc vivo):** S9 pasa de «11 elementos/nodos» a
> **12** (nodo `harness-profile`, HS-07) y el inventario S2 de «7 tipos» a **10 clases** (enum
> canónico L0 de HS-08; tokens de color 6→10 resueltos en HS-09 Fase D). Sin cambio de diseño.

## Bifurcaciones firmadas (grill 2026-07-04)

1. **Alcance: todas las superficies parejas.** Diseño fino firmable de Mapa, fábrica
   conversacional, tren de release y diagnóstico — cero deuda UX para fases 3-5.
2. **Fábrica conversacional = dock en el lienzo.** Panel acoplado al mapa; los componentes
   nacen y mutan EN el mapa mientras se conversa (propuesto punteado → beta → estable).
   Honra "crear y editar = acciones sobre el mapa, no vistas aparte".
3. **Entrada = portafolio de arneses.** Home con la producción de la fábrica de un vistazo
   (salud + telemetría resumida por arnés) → clic entra al mapa. Picker en toolbar para
   saltar sin volver.
4. **Herencia shell de lentes: capas + vistas hermanas.** Lentes espaciales
   (Estructura/Tokens/Desempeño/Proceso) = capas sobre la geografía del mapa. Lentes no
   espaciales (Historia, Diagnóstico) = pestañas hermanas. El **contrato-de-lente
   sobrevive como principio de arquitectura de capas** (se diseña en fase 3); la UI del
   shell cero-dep muere. Deuda heredada que esta UX SÍ resuelve: contexto inter-vista
   (OBS-13: el deep-link "con ese nodo" se perdía entre lentes).
   **› Superado por it.13/it.14** (ver «Inventario final» abajo): la chrome que aloja capas/
   pestañas dejó de ser el shell-de-lentes original — es el Command Rail (it.13) mutado a
   rail-de-sesiones (it.14); el principio de capas/pestañas sobrevive, la UI que lo hospeda no.

## Journeys del operador (fábrica: crear · mapear · observar · mejorar)

- **J1 · Observar la producción.** Abrir app → portafolio (salud de N arneses) → entrar a
  un arnés → Mapa capa Estructura → conmutar capas → clic en nodo → inspector (métricas,
  hallazgos, acciones).
- **J2 · Diagnosticar y corregir.** Pestaña Diagnóstico (hallazgos estáticos + runtime,
  misma bandeja) → hallazgo → "ver en el mapa" (foco nodo + capa relevante) → «Editar» →
  dock fábrica abre CON contexto del hallazgo → conversación → cambio queda en beta →
  eval A/B → promote → publish. El loop detectar→forjar→propagar (OBS-18) en una sola
  superficie.
- **J3 · Crear arnés desde cero.** Portafolio → «Nuevo arnés» → grill conversacional
  (rol × proceso de la compañía → fases del proceso = carriles del mapa) → spec → build
  headless con progreso visible (los nodos aparecen punteados en el mapa mientras nacen)
  → beta → gate → promote → publish al marketplace git del proyecto. Telemetría conectada
  de nacimiento (principio 9) — no hay paso "conectar".
- **J4 · Crear/editar componente suelto.** Acción sobre el mapa («+» en carril o banda, o
  «Editar» en inspector) → dock con contexto de fase/tipo → nodo propuesto punteado →
  mismo tren que J3.
- **J5 · Tren de release.** Pestaña Tren: cola de candidatos beta, estado del eval-gate
  por candidato, promotes históricos, publish por proyecto destino. Nada se promueve sin
  eval (principio 10).
- **J6 · Historia.** Pestaña Historia: línea de versiones del arnés, diff de componentes
  entre versiones (+/−/~), hitos ligados a fichas del ledger del arnés.
- **J7 · Base incompleta.** El arnés exige base as code antes de operar (principio 2):
  banda Base en alerta con faltantes explícitos → «Capturar base» abre el dock en modo
  captura. Guía sin bloqueo (principio 6): advierte y registra excepción, no detiene.

## Superficies

| # | Superficie | Rol |
|---|-----------|-----|
| S1 | **Portafolio** (home) | Producción de la fábrica: cards por arnés — salud, versión/canal, corridas 30d, hallazgos abiertos, proyecto(s) donde está instalado |
| S2 | **Mapa** (lienzo único) | Geografía fija: banda Guardia · carriles por fase · banda Base. 4 capas conmutables. Edges invocación/lectura/escritura |
| S3 | **Inspector** (panel derecho) | Identidad + métricas + hallazgos + acciones del nodo (Editar · Evaluar A/B · Promover · Ver historia) |
| S4 | **Dock fábrica** (panel conversacional del lienzo) | Modos: crear arnés · crear componente · editar componente · capturar base. Muta el mapa en vivo; Claude Code headless detrás (patrón conductor) |
| S5 | **Diagnóstico** (pestaña hermana) | Bandeja unificada estático+runtime; severidad; cada hallazgo → foco en mapa |
| S6 | **Tren** (pestaña hermana) | beta → eval-gate → promote → publish (release train KIT-06 operado, no duplicado) |
| S7 | **Historia** (pestaña hermana) | Versiones, diffs, hitos del arnés |
| S8 | **Corridas** (pestaña hermana) | Corridas reales JSONL: Conversación/Árbol/Waterfall + attribution paso→mapa + replay |
| S9 | **Estándar** (screen global, cross-arnés) | El árbol de conocimiento vivo (`docs/architecture/knowledge/`): 12 elementos con L1/L2/checks (12º = `harness-profile`, HS-07) + botón «Actualizar estándar» → drawer de novedades. Chip global en el header |

**Navegación (detalle completo → «Inventario final» §I abajo, es la misma chrome, no se repite
aquí):** shell = Command Rail (it.13) mutado a rail-de-sesiones (it.14); picker de arnés en el
breadcrumb; conmutador de capas en la barra del mapa; deep-link por hash de todo.

## Estados que la UX debe cubrir

- Arnés recién nacido, sin corridas: mapa completo pero métricas "—"; badge «sin
  telemetría aún» (nunca inventar datos; procedencia obligatoria, herencia OBS-08).
- Base incompleta: banda Base en alerta, faltantes nombrados (J7).
- Build headless en curso: nodos naciendo punteados + stream de progreso en dock.
- Daemon caído / reconexión: los JSONL son la fuente de verdad — aviso «índice
  desactualizado», jamás pérdida.
- Candidato beta con gate fallado: visible en Tren + pintado en mapa (capa Desempeño).

> **Nota de numeración:** el registro guarda las iteraciones donde se **fijó un acuerdo** (por eso
> hay saltos: it.1 = grill/bifurcaciones sin numerar; it.4/it.5 se disolvieron en las subversiones
> v3.1–v3.2 de it.3 sin acuerdo propio que cementar). No es un hueco: falta mucho trabajo de UX y el
> doc sigue vivo (crece con nuevas iteraciones).

## Historia de las iteraciones 2–14 (condensada en el LEDGER)

> Las decisiones y verdictos vuelta-por-vuelta del grill (it.2 a it.14 — feedback sobre
> mockups auditados, hallazgos con luana real, doctrina de fábrica de cajas, S9 Estándar,
> Session Rail) viven condensados en [`ledger/HS-03.md`](ledger/HS-03.md) — eje historia,
> separado de este documento (eje vigente) para no volver a mezclar los dos ejes. El texto
> verbatim de cada iteración queda en el historial git de este archivo (commits pre
> 2026-07-10). Lo que sigue es el estado VIGENTE que esas iteraciones produjeron.

## Inventario final (firma) — corte iteración 13 (2026-07-05)

> Baseline **firmado** de HS-03 (it.13, 2026-07-05). El snapshot «corte iteración 6» de abajo se conserva como
> historia. **Dos mockups, dos roles** (se unifican al portar): SHELL/navegación/organigrama =
> `arnesia-shell-A-galaxia.html` (artifact 682f3890) · detalle profundo de superficies =
> `arnesia-mockup-v3.html` (artifact 6a63cdf3) · laboratorio de shell (4 paradigmas, A firmado) =
> `arnesia-shell-lab.html` (artifact 9bcef775).

**I · Shell & navegación (shell-A, chrome al día con it.14 — única fuente de este detalle, ver
nota en «Superficies» arriba).** El rail izquierdo es **rail-de-sesiones** (P1 Session Rail,
~224px, colapsa a gutter ~52px): tarjeta por sesión (frente·arnés·empresa·salud·estado CC vivo),
«Nueva sesión» la lanza desde Portafolio/Organigrama (lente que el Portafolio gana en it.13:
arneses por empresa/puesto, «reporta a»). Las **vistas** (Portafolio·Mapa·Corridas·Diagnóstico·
Tren·Historia) bajan a **tira slim por sesión** (a la derecha del rail — eran ítems del rail en
it.13, antes de mutar a sesiones en it.14); el **global** (Portafolio·Estándar·Ajustes) baja al
pie del rail. Topbar con breadcrumb (empresa / picker-arnés ▾ / vista / `· frente «…»` si N>1) +
chip global ⟳ Estándar + botón «Conversar ⌘K» · **chat dock invocado** (derecho, contextual,
colapsable): fila de sesión CC (id·modelo·tokens·barra ctx·%) + mensajes user/sys/asistente + diff
+ CTAs «Aplicar→beta»/«Ver en el mapa» + composer; Claude Code headless detrás. Muere el backlog
#1 (el dock ya no come el mapa). **Contexto inter-vista + sesiones persistentes (paga la deuda
OBS-13, un solo mecanismo):** toda entidad (nodo, hallazgo, candidato, versión) navega a cualquier
vista conservando foco — estado serializado en hash (arnés·vista·capa·selección) ⇒ deep-link
compartible; al reabrir la app las sesiones se restauran completas (arnés+vista+selección+
conversación) desde ese mismo estado. Light/dark · reduced-motion · focus · toasts honestos.

**II · S1 Portafolio, 2 lentes (shell-A).**
- **Organigrama (hero):** tablero libre 2D **arrastrable** por empresa; nodo = arnés, **línea =
  «reporta a»** (bezier SVG), color = salud, badge = hallazgos, badge «raíz»; drag real (la línea
  sigue, clic-sin-mover = entra al mapa). CTA «＋ Conectar empresa».
- **Cuadrícula:** triage peor-primero; tarjeta con empresa·puesto·**reporta-a**·**marketplace**,
  estado (sano/atención/crítico/naciendo/**señales incompletas**/inactivo), señales + hallazgos.
- **Metadato de arnés de primera clase:** puesto · empresa · reporta-a (en tarjeta, breadcrumb y
  barra del mapa). **Marketplace por empresa:** repo + tag (mío / repo propio) + gestionar +
  ⇢ desplegar. Alcance declarado: aquí SOLO se crean y mantienen arneses; los puestos sin arnés no
  aparecen (lo demás vive en el cockpit, futuro).

**III · Superficies de detalle (v3).**
- **S2 Mapa:** Guardia / carriles-fase / Base + **bandas transversales** (Librería de expertos ·
  Meta-harness·Git·Utilidades·Comandos · Marcas dormidas · Terceros·Clerk). **10 clases**
  color+forma+etiqueta (enum canónico L0 de HS-08: skill·subagent·hook·rule·command·mcp·plugin·
  settings·output-style·statusline — al corte it.13 eran 7 tipos; gap de tokens de color 6→10
  resuelto en HS-09 Fase D). **4 capas con
  contrato de datos:** Estructura · Tokens (carga+30d, fórmula única, heat percentil, Σ carril,
  budget+Pareto, solape declarado) · Desempeño (inv×·éxito%·p95/invocación·sin-uso·gate-fallado·
  grosor de edge) · **Proceso = pipeline de contratos por caja** (entra→cajas→eval-gate chip→sale,
  rework ↩, cajas paralelas, gate auto/manual/parcial/**sin-gate**). Edges direccionales. Leyenda
  = filtro por tipo (pendiente: falta el filtro de `command` — backlog). Replay de corrida sobre
  el mapa (#1..#n, camino, deep-link). Base en alerta + CTA capturar.
- **S3 Inspector (3 pestañas):** Resumen (identidad, métricas inv/éxito/p95/carga/excepciones/
  consumo, **Contrato: necesita↑·entrega·ruta↓ condicional·viene-de** [inferido/derivado/no-
  declarado], hallazgos, acciones) · Contenido (fuente + editar → diff → nota → beta → tren;
  CodeMirror 6) · Corridas.
- **S8 Corridas:** lista JSONL (disparo·fase·fecha·modelo·turnos·in/out·ctx·costo·duración·✓/✕·
  ⚠exc); detalle 3 vistas Conversación / Árbol (contabilidad de contexto por paso) / Waterfall +
  attribution paso→mapa + ▶ reproducir en el mapa.
- **S5 Diagnóstico:** bandeja unificada estático+runtime, severidad, capa sugerida, «ver en el
  mapa», badge de count. **S6 Tren:** horno / eval-gate / listo / publicado; gate-fallado →
  «Reabrir en fábrica»; promote bloqueado por gate; destinos publish. **S7 Historia:** versiones +
  deltas + hitos. **S9 Estándar (global):** 12 nodos (12º = `harness-profile`, HS-07) + rail L1/L2/checks (severidad) + botón
  Actualizar → drawer de novedades (tier híbrido auto vs requiere-OK, arneses afectados, cómo-
  aplicarlo con diff); deep-link `#/estandar/<nodo>`. **S4 Dock fábrica** (equivalente al chat
  dock): 5 modos (editar · crear-componente · crear-arnés · capturar-base · **importar arnés** →
  chequeo constitucional → adopción) + fila de sesión CC + eventos de sistema.

**IV · Global.** ⚙ Marketplaces · deep-link por hash de todo (paga OBS-13) · **honestidad de
producto intacta** (nada se inventa · procedencia · «—» sin medir · gris ≠ verde) · daltonismo
(color+forma) · toasts honestos. Estados: naciendo, sin telemetría, señales incompletas, base
incompleta, gate fallado, sin uso, propuesto, corrida fallida, bandejas vacías.

**Decisiones de arquitectura firmadas:** shell = Command Rail (A) · D Canvas-native = milestone
futuro · contrato de datos por capa · frontera Proceso↔Diagnóstico · p95 por invocación (jamás
promedios) · atribución de tokens solapada (Σ ≠ total) · fórmula de consumo única mapa=inspector ·
heat percentil · edición directa siempre pare beta→tren · Contrato por caja como forcing function ·
**andamiaje REAL/DEMO retirado del mockup, honestidad de producto conservada.**

## Inventario de funcionalidades — corte iteración 6 (snapshot histórico)

> **Snapshot histórico de UX** (nota 2026-07-09, HS-18; recondensado 2026-07-10). El **SSoT
> funcional VIVO** de qué sabe hacer el sistema es [`capabilities/INDEX.md`](capabilities/INDEX.md)
> (82 capabilities derivados del código, con puntero a la implementación de cada uno). El
> detalle completo de este corte (mockup v3.3, iteración 6) vive en
> [`ledger/HS-03.md`](ledger/HS-03.md) junto al resto de la historia — no se repite aquí.

## Backlog de UX

> El backlog de profundización (vista A/B real, búsqueda global, accesibilidad, organigrama,
> badge de conformidad conectado a Diagnóstico, etc.) vive en [`BACKLOG.md`](BACKLOG.md) —
> ítems abiertos con tag `[ux]`. Los ya resueltos (dock a 1280px → it.13, contrato de datos
> por capa → it.6) quedan como historia en `ledger/HS-03.md`; el ítem de Artefactos se dio de
> baja de este backlog porque ya fue EJECUTADO (HS-13).

<!-- Al cerrar cada iteración: registrar en `ledger/HS-03.md` (o la ficha vigente) + actualizar el backlog. -->
