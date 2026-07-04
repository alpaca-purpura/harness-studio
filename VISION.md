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
- **Empaque:** GoReleaser → brew/scoop/deb. Tauri 2 = milestone futuro "app vendible"
  (wrapper sobre el mismo daemon, no rewrite). Wails v3 en watchlist (alpha aún).

## Lo que NO es

- No opera arneses ajenos ni es "gestor universal de configs".
- No es marketplace hosted propio — publica a marketplaces git estándar.
- No es la app del trabajador — eso es DevHub y las apps de rol.
- No es observabilidad LLM genérica — no compite con Langfuse; la usa.
- No duplica primitivas P3 — las opera.

## El gran plan

| # | Fase | Estado |
|---|------|--------|
| 1 | Visión del producto | **esta ficha (HS-02)** |
| 2 | UX del producto (mapa a fondo, inspector, flujos) | siguiente |
| 3 | Arquitectura del software y diseño técnico | pendiente |
| 4 | Definición de specs | pendiente |
| 5 | Implementación y pruebas (MVP = Mapa primero) | pendiente |
| 6 | Instalación y dogfood | pendiente |

**Port del monorepo (regla):** nada se porta sin pasar por la fase del gran plan que le
corresponde. Sobrevive en principio: patrón conductor (I-76/OBS-16/OBS-18), flujo creador
sobre KIT-06, contrato L0 `meta.clase` (I-75 — el grafo agnóstico es su evolución). Muere:
UI Next, nombre. En evaluación (fase UX): shell de lentes (D2/OBS-13) vs mapa React Flow
como visor único.

## Debates abiertos (no bloquean la firma)

1. **Segundo cerebro** del arnés (vector DB / SQLite obligatorio en la constitución) — fase 3.
2. Diseño fino del mapa, inspector y acciones — fase 2.
3. Config de marketplace por proyecto (¿dónde se declara el destino?) — fase 3/4.
4. Renombrar el repo `harness-studio` → `arnesia` — operativo, decide el operador.
