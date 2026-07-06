# ArnesIA — Visión de producto (v3)

> **FIRMADA por el operador — 2026-07-04** (ficha HS-02, forjada y firmada en la primera
> sesión de trabajo del repo, como mandó HS-01). Registro: [`LEDGER.md`](./LEDGER.md).
> Historia y visión anterior: `prenter-harness/products/harness-studio/` (incubadora, congelada).

## Tesis

**ArnesIA es la fábrica de arneses.** Crea conversacionalmente, mapea como grafo vivo,
observa en producción y mejora continuamente los arneses que nosotros (alpacapurpura)
construimos y vendemos a empresas. Un arnés no es un skill suelto: es un **framework
file-based definido por rol × proceso de la compañía** — el arnés del developer full-cycle,
el del product manager, el del DevOps, el del analista de finanzas. ArnesIA es el único
lugar donde los arneses nacen, se ven, se miden y se corrigen; las apps de rol (DevHub y
sucesoras) son donde los trabajadores los ejecutan.

Claude Code headless (stdin/stdout, patrón conductor) es el cerebro detrás de la creación;
el mercado puede mutar a otros agentes (OpenCode, Codex, pi, Antigravity) porque el núcleo
del producto es un **grafo de componentes agnóstico** y cada agente es solo un adaptador.

## El arnés — constitución (lo que TODO arnés nuestro ES)

1. **Proceso implícito en el core.** El proceso al que sirve el arnés es su columna: define
   qué skills, agentes, hooks y knowledge actúan, cuándo y para qué. No se declara como
   burocracia; se encarna en los componentes.
2. **Base antes de acción.** El arnés es framework, no herramienta suelta: exige la base del
   proyecto capturada as code (arquitectura, flujos, capabilities, tecnologías, visión) en su
   instalación, ANTES de producir nada nuevo. Sin base, no opera.
3. **Rol × proceso.** Cada arnés sirve a un rol dentro de un proceso de la compañía. En
   empresa chica un rol abarca el proceso entero; en empresa grande el mismo proceso se parte
   en varios arneses (PM → dev → DevOps). Agnóstico a rubro; moldeado al proceso.
4. **Aditivo y sin pérdida.** El trabajo real cruza muchas conversaciones; cada paso suma
   sobre el anterior y nada se pierde entre sesiones. El ledger del proyecto es memoria, no
   adorno.
5. **Autodocumentación como efecto.** Specs as code, arquitectura as code, capabilities —
   emergen del trabajo mismo, nunca como tarea aparte. Son artefactos vivos que otras apps
   (DevHub) consumen.
6. **Guía sin bloqueo.** El arnés nunca detiene al trabajador: advierte, mitiga llenando
   vacíos y registra la excepción. Las excepciones son datos de mejora, no castigos.
7. **Agnóstico a rubro y tecnología, moldeable a proceso.** Mismo estándar constitucional
   para software, finanzas o lo que venga; solo cambia el proceso que encarna.

## Gobierno (cómo ArnesIA opera los arneses)

8. **Estándar propio.** Solo operamos arneses nuestros. Nosotros seteamos el estándar; el
   Mapa puede ASUMIR la constitución en vez de adivinarla. Arneses ajenos no nos interesan.
9. **Telemetría de nacimiento.** Todo arnés nace conectado al monitoreador. No es opt-in.
10. **Nada se promueve sin eval.** Beta → eval gate → promote (release train KIT-06). El
    canary de skills no existe como producto en el mercado; nosotros lo hacemos norma.
11. **Economía de contexto medible.** Cada componente rinde cuentas de sus tokens (carga e
    invocación). Toda edición conoce su costo de invalidación de caché.

## Anatomía del arnés — la fábrica de cajas de proceso

> Añadido en fase UX (HS-03, iteración 9, 2026-07-04) — operacionaliza el principio 1
> («proceso implícito») y aterriza el 10 («nada sin eval»). Aditivo: no altera los 11
> principios firmados.

Un arnés es una **fábrica**: el trabajo entra por un extremo, cruza una línea de **cajas de
proceso** y sale transformado. Reglas de cómo se arma todo arnés nuestro:

- **A1 · Jerarquía fase › caja › maquinaria.** El arnés se organiza en **fases**; cada fase
  agrupa una o más **cajas de proceso**. Una caja = **una skill** (el frente que define el
  trabajo de esa etapa) que orquesta su **maquinaria dedicada** (agentes, sub-skills) y se
  apoya en infraestructura compartida (rules, hooks, knowledge).
- **A2 · Contrato input → output.** Toda caja declara qué **entra** y qué **sale**, como
  artefactos as code (spec, arquitectura, código, review). La **salida de una caja es la
  entrada de la siguiente**: el hand-off ES el contrato, no una convención.
- **A3 · Dos estados, no los confundas.** El **trabajo** (la unidad que fluye: historia,
  ticket, caso) lleva su propio estado a lo largo de la línea. Cada **caja es dueña de UNA
  transición** de ese estado (entra en X, sale en Y). La caja además tiene su **estado
  operativo** (corriendo/ociosa/bloqueada, éxito, throughput) — lo que mide la telemetría.
  Estado del trabajo ≠ estado de la caja.
- **A4 · El eval-gate vive en el contrato de salida.** El principio 10 se aterriza aquí: el
  gate de una caja evalúa **si su output honra el contrato dado el input**. Sin contrato de
  salida declarado no hay dónde poner el eval — y la caja no puede prometer calidad.
- **A5 · La fábrica no es una recta.** Hay **retrabajo** (una caja de control devuelve el
  trabajo a una caja anterior: el bucle detectar→corregir) y **cajas en paralelo** dentro de
  una fase (una línea por variante: por marca, por tipo de trabajo). El mapa honra ambos:
  aristas de retorno y cajas apiladas.
- **A6 · La infraestructura compartida no es una caja.** Los hooks transversales (banda
  Guardia) y el knowledge/reglas (banda Base) actúan sobre TODAS las cajas; viven en
  **bandas**, no dentro de una fase. Solo la maquinaria *dedicada* de una caja vive en ella.
- **A7 · Crear un arnés = definir sus fases y los contratos entre cajas.** La fábrica
  conversacional (grill → spec → build) produce ante todo esta línea: las fases, la caja
  (skill) de cada una, su contrato input→output y la transición de estado que posee. Agentes,
  rules y hooks se cuelgan después como maquinaria y apoyo.

## Linaje doctrinal y precisiones (aditivo · 2026-07-05)

> Nota aditiva firmada por el operador (2026-07-05). **No altera los 11 principios ni A1–A7:** los
> ancla a una disciplina y precisa una frontera. Detalle, evidencia y bibliografía:
> [`research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md`](./research/2026-07-05-doctrina-propia-v1-adaptacion-daop.md).

**Qué operacionalizamos — no clonamos un framework.** ArnesIA **operacionaliza Agentic Business
Process Management (APM)**: una disciplina de PROCESO, independiente de dominio, con ~30 años de
linaje BPM. Un arnés ES un *framing mechanism* instanciado (frame normativo + operacional +
conocimiento + tools) que da **framed autonomy** a Claude Code por rol×proceso. Esto blinda dos
principios firmados: p1 «proceso implícito» (el frame ES el proceso) y p7 «agnóstico a rubro» (APM
deja la implementación abierta por diseño). BMAD/DAOP, Sierra y Salesforce son **insumos filtrados**,
no el padre de la doctrina.

**Distinción rectora — autonomía ≠ automatización.** Una caja *pipeline* es automatización (ejecuta
lo especificado); una caja *abierta / excepción* es framed autonomy (percibe, razona y elige DENTRO
del frame). Es el eje de nuestros **arquetipos de trabajo** (METODOLOGIA §8).

**Vocabulario propio (renombra mejor lo que ya teníamos):** banda Guardia + permisos + rules =
**frame normativo (deóntico) + operacional** · loop de mejora = **adaptation** (instancia, efímero)
vs **evolution** (modelo, persistente) · creación conversacional + dock = **conversational
actionability**. Las **4+1 capacidades** que todo arnés debe proveer: framed autonomy ·
**explainability** · conversational actionability · self-modification.

**Precisión del principio 6 (frontera guía / Guardia — RATIFICADA).** «Guía sin bloqueo» (p6) rige
la **guía de proceso/calidad**: el arnés nunca detiene al trabajador por incumplir el proceso — la
excepción es dato de mejora. Es un plano DISTINTO de la **banda Guardia**: los hooks SÍ bloquean
(exit 2) efectos externos destructivos / de-secreto / de alto riesgo (HITL). Dos planos, sin
contradicción — p6 no autoriza efectos peligrosos, y la Guardia no bloquea el flujo de trabajo.

**Seam organizacional (producto puro — RATIFICADO).** ArnesIA NO modela el proceso de la empresa
(MOF / organigrama): eso vive en OTRO sistema, futuro; su salida será el input para *crear* arneses.
Nuestro seam = la **META de enganche** que ya carga cada arnés (rol · proceso · reporta-a · empresa).
No construimos el organigrama; exponemos la superficie para engancharlo.

**Roadmap (dogfood-first — RATIFICADO).** El primer entregable de producto es UN arnés real
end-to-end (el dev-full-cycle) sobre nuestro propio proceso, ANTES del Mapa/compilador — para que el
Mapa renderice datos reales, no mocks.

## Ecosistema y fronteras

```
ArnesIA (nuestra fábrica: crea · mapea · observa · MODIFICA)
   └─ publica → marketplace git (elegible POR PROYECTO, formato marketplace.json + semver/SHA)
        └─ instala → arnés en el proyecto del cliente
             └─ ejecuta ← DevHub y apps de rol (la herramienta del trabajador:
             │            interfaz amigable, restringe el trabajo a su rol×proceso,
             │            Claude Code por detrás)
             └─ telemetría (estrato seguro: métricas · scores · hashes) → ArnesIA
```

- **ArnesIA es dueña única del observar y el modificar.** DevHub y pares jamás editan
  arneses; solo los usan y emiten telemetría.
- **El análisis de datos de cliente vive fuera de los repos de fábrica** (I-53/I-39):
  solo métricas, scores y hashes cruzan el borde. Langfuse actúa de espejo/backend de
  equipo, nunca de dependencia dura.
- **P3/Kit:** ArnesIA OPERA sus primitivas (marketplace, sensor, eval-gate); no las duplica.
  Cross-repo por contrato/ID/slug, nunca por ruta frágil.

## Buyer y job-to-be-done

El producto **vendible es el arnés con su mejora continua** — empresas que gestionan su
adopción de IA compran arneses por rol×proceso, las apps de rol para ejecutarlos, y el
servicio de que mejoren solos con el uso. **ArnesIA es el medio de producción** que hace ese
negocio escalable: interno primero, dogfood siempre. (Si ArnesIA misma se vuelve producto,
será ficha futura — no es esta visión.)

## Qué mutó (respuesta a HS-01)

- **Queda:** crear · observar · mejorar; la observación como sensor, no producto entero; el
  loop detectar→forjar→propagar que OBS-18 ya cerró una vez; patrón conductor; KIT-06.
- **Entra:** arnés = framework por rol×proceso con constitución de 11 principios; el Mapa
  como lienzo único del producto; grafo agnóstico como modelo núcleo; ecosistema explícito
  con apps de rol; telemetría de nacimiento; nombre **ArnesIA**.
- **Muere:** "Harness Studio" como nombre (colisión Harness.io); UI Next embebida (→ Vite);
  la idea de operar/cargar arneses de terceros.

## Por qué ahora (evidencia 2026, investigada 2026-07-04)

- El arnés importa casi tanto como el modelo: Claw-SWE-Bench — arnés fijo mueve Pass@1
  **27.4 pts**, cambiar modelo 29.4. "Harness engineering" ya es disciplina nombrada.
- **Nadie combina** creación + mapa-grafo + monitoreo por trazas + publicación. Nadie dibuja
  el mapa estático de un arnés; nadie une trazas al inventario de componentes; todo
  marketplace muere en install counts.
- Estándares ganadores adoptados: **SKILL.md** (agentskills.io, estándar abierto, 26+
  plataformas), **AGENTS.md** (Linux Foundation), marketplace.json git + semver/SHA.
- Anthropic ya hace eval-driven authoring de skills (skill-creator Eval/Improve/Benchmark) —
  valida el loop y es a la vez la amenaza de ensamblaje first-party. Nuestra defensa: mapa
  visual, atribución traza→componente, multi-agente, y que el negocio es el arnés, no la tool.

## El producto — funcionalidades núcleo

- **Mapa = lienzo único.** Geografía fija de 3 niveles: banda **Guardia** (hooks
  transversales) · **carriles por fase del proceso** · banda **Base** (knowledge as code).
  Cuatro capas conmutables sobre la misma geografía: Estructura · Tokens · Desempeño ·
  Proceso. Crear y editar son acciones sobre el mapa, no vistas aparte. (Esencia validada
  en mockup 2026-07-04; diseño fino en fase UX.)
- **Sensor local-first.** Los JSONL de `~/.claude` son la fuente de verdad (Claude Code ya
  es el colector); el sensor indexa — daemon caído = cero pérdida. Índice SQLite desechable;
  hooks + traces OTel beta complementan; export a Langfuse como espejo.
- **Fábrica conversacional.** El arnés se crea conversando (grill → spec → build headless →
  beta), con Claude Code como cerebro vía patrón conductor.
- **Tren de release.** beta → eval gate → promote → publish al marketplace git del proyecto;
  telemetría post-deploy de vuelta al mapa.
- **Diagnóstico unificado.** Hallazgos estáticos (lint, bloat, costo de carga) y runtime
  (fallos, componentes sin uso, presión de contexto) — misma bandeja, pintados sobre el mapa.

## Decisiones técnicas fundacionales

- **Binario Go único** `arnesia` — modos `serve` (watcher + indexer + API HTTP/SSE + UI
  embebida), `open`, `index`, `publish`. Topología Syncthing/opencode, validada 2026.
- **UI: Vite + React SPA** embebida vía go:embed (Next.js muere: pelea con export estático).
- **Mapa: React Flow 12** + layout de carriles custom (~200 líneas; elkjs solo si hace falta).
- **Storage: SQLite puro-Go** (modernc, WAL, rollups) — cross-compile sin CGO. DuckDB solo
  si años de telemetría lo exigen.
- **Empaque:** GoReleaser → brew/scoop/deb para el daemon. **Shell v1 = Tauri 2 desde el
  nacimiento** (fork firmado HS-04, 2026-07-05): app de escritorio nativa con el daemon Go como
  sidecar `externalBin`; el WebView consume la misma API HTTP/SSE. El boundary `core⊥shell` no
  cambia — «servable headless» sigue gratis y la ruta a la «app vendible» es aditiva (wrap, no
  rewrite). Wails v3 en watchlist. *(Supera la lectura previa «Tauri = milestone futuro»: nacemos
  con Tauri. Ver HS-04 en el LEDGER.)*

## Lo que NO es

- No opera arneses ajenos ni es "gestor universal de configs".
- No es marketplace hosted propio — publica a marketplaces git estándar.
- No es la app del trabajador — eso es DevHub y las apps de rol.
- No es observabilidad LLM genérica — no compite con Langfuse; la usa.
- No duplica primitivas P3 — las opera.

## El gran plan

| # | Fase | Estado |
|---|------|--------|
| 1 | Visión del producto | **esta ficha (HS-02)** ✓ |
| 2 | UX del producto (mapa a fondo, inspector, flujos) | ✓ firmada (HS-03, it.13) |
| 3 | Arquitectura del software y diseño técnico | ✓ **as code** (HS-04 backend + HS-05 frontend) — stack cerrado + `arch/` = 77 checks |
| 4 | Definición de specs | **siguiente (HS-06)** |
| 5 | Implementación y pruebas (MVP = Mapa primero) | pendiente |
| 6 | Instalación y dogfood | pendiente |

**Port del monorepo (regla):** nada se porta sin pasar por la fase del gran plan que le
corresponde. Sobrevive en principio: patrón conductor (I-76/OBS-16/OBS-18), flujo creador
sobre KIT-06, contrato L0 `meta.clase` (I-75 — el grafo agnóstico es su evolución). Muere:
UI Next, nombre. En evaluación (fase UX): shell de lentes (D2/OBS-13) vs mapa React Flow
como visor único. *[Resuelto en HS-03 it.13 (2026-07-05): el shell es **Command Rail** — rail
de iconos izq. + chat como dock derecho + Portafolio con lente Organigrama; ni «shell de
lentes» puro ni «visor único». Ver UX.md «Iteración 13». Nota aditiva; no altera los 11
principios ni A1–A7.]*

## Debates abiertos (no bloquean la firma)

1. **Segundo cerebro** del arnés (vector DB / SQLite obligatorio en la constitución) — fase 3.
2. Diseño fino del mapa, inspector y acciones — fase 2.
3. Config de marketplace por proyecto (¿dónde se declara el destino?) — fase 3/4.
4. Renombrar el repo `harness-studio` → `arnesia` — operativo, decide el operador.
