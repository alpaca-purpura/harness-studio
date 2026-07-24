# Decisiones — índice SQLite real + watcher fsnotify (fase 5)

> Paquete 100 % backend, sin gate de mockup (disciplina §10, caso «backend puro» — ver
> precedente `2026-07-13-portafolio-slice0-cimientos/decisiones.md`). Ejecución **autónoma por
> directiva `/goal` del operador (2026-07-24)**, sin diálogo en vivo: cada decisión de acá queda
> marcada **DECIDIDA (agente, autónoma por /goal 2026-07-24)** — NO es una firma 🧑‍⚖️ de
> operador. El gate humano de PARIDAD sigue abierto y separado (RF-214 en `spec.md`).
>
> Numeración: **D1..D6**. D1-D5 resuelven las 5 preguntas que la consigna del agente 1
> (`hallazgos.md`) dejó abiertas explícitamente para el agente 2. D6 es un hallazgo propio,
> descubierto investigando D1/D5, que no estaba en la lista original pero bloquea poder
> escribir el RF de `Rebuild()` sin ambigüedad.

## D1 · Cursores de reanudación `{path,inode,size,offset}` — FUERA de este paquete

**Qué:** los cursores de reanudación por-archivo que el boundary v1.1/v1.2 diseña bajo
«Rebuild rápido» **no se construyen en Fase 3**. `Rebuild()` recarga cada arnés completo
(`loader.LoadArnes(path)`, lectura íntegra del árbol) en cada corrida, sin trackear offsets.

**Por qué:** el diseño de cursores tiene sentido para UN archivo JSONL grande que crece por
líneas apendeadas (offset = «hasta dónde ya leí»). El corpus real (T0, `hallazgos.md` §2) es el
**árbol de arneses** — un puñado de archivos chicos (`SKILL.md`, `arnes.l0.json`, hooks/rules
`.md`/`.json`) que **no crecen por apéndice**; `loader.LoadArnes` ya los relee íntegros y es
barato: hoy el índice tiene 3 grafos sembrados (`store.go:93-163`) y el caso real de uso (sesión
de chat contra un arnés) es de **1 arnés a la vez**. Aunque el Portafolio llegue a tener decenas
de entradas, releer archivos de texto de un `.claude/` completo en cada `Rebuild()` es
microsegundos, no el problema de I/O que los cursores resuelven en la industria (logs de GB).
Construir el mecanismo de offsets hoy sería trabajo sin caller real (regla 1 del ladder: ¿esto
necesita existir?). Deuda futura explícita: **si** el corpus indexado alguna vez incluye archivos
que crecen por apéndice (p.ej. si el índice empieza a indexar historial de chat, hoy
explícitamente fuera de alcance — ver `INDEX.md` «Fuera de alcance explícito»), este ticket
revive.

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).

## D2 · `reconstruccion-del-indice.yaml` (CAP-22), entrada `~/.claude` en `valida:` — SE LIMPIA

**Qué:** el agente 3, al implementar `Rebuild()` real, retira la entrada `~/.claude` de
`valida:` (no es un nombre de test Go — placeholder heredado del barrido HS-18) y la reemplaza
por los tests reales que cubran el nuevo `Rebuild()` (p.ej. el que active
`TestIndexRebuildsFromJSONL`, más los que el agente 3 escriba para el camino
`ArnesRegistry→loader→Upsert`, ver D6). `status` sube de `parcial` a `vivo` cuando `Rebuild` deja
de ser el mismo `seed()` (hoy resuelve a código real pero un `seed()` disfrazado — no es
`parcial` por capricho, es la descripción honesta del gap).

**Por qué:** R4 (`codigo-traza-a-capability.md`) exige que `valida:` liste tests reales — un
placeholder que no es un `^func Test` es directamente falso una vez que alguien intente
correrlo como enforcer. Es limpieza gratis en el mismo commit que ya toca ese símbolo; no
amerita un ticket aparte.

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).

## D3 · Nombre del boundary `indice-desechable-jsonl-es-verdad` — NO SE RENOMBRA

**Qué:** el slug/archivo/título se mantienen tal cual (`indice-desechable-jsonl-es-verdad.md`,
regla `indice-desechable-jsonl-es-verdad`). NO se toca `arch_test.go`, `BACKLOG.md`,
`ledger/HS-16.md`, ni los 3 paquetes de `stories/` que lo citan por nombre.

**Por qué:** el rename es **costo real (≥6 archivos a tocar, riesgo de referencia rota) por
beneficio cero de comportamiento** — puro cosmético. El propio doc ya resuelve la confusión que
el nombre podría generar: la v1.2 (`indice-desechable-jsonl-es-verdad.md:45-59`) explica in
extenso que «JSONL» es la analogía L1 de industria (Claude Code transcripts como ejemplo
genérico de «índice desechable sobre una fuente durable»), y que el corpus L2 concreto de este
árbol es el árbol de arneses, no JSONL de conversación. Es el mismo patrón que YA existe en el
repo para otros nombres heredados que sobreviven como analogía documentada en vez de
renombrarse (p.ej. wording de «indexer JSONL» en el mensaje de CLI de `runIndex`,
`hallazgos.md` línea 58-63, dejado deliberadamente). Ladder regla 1: un rename sin caller que lo
necesite es trabajo especulativo. Si en el futuro el boundary gana un quinto principio o cambia
de fondo, ahí sí se justifica una migración de nombre con su propio ticket — no acá.

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).

## D4 · Schema del store SQLite — CONFIRMADO: 1 tabla JSON-blob por clave

**Qué:** el store nuevo usa **una** tabla `graphs(clave TEXT PRIMARY KEY, graph_json TEXT,
updated_at TEXT)` (más la tabla meta `schema_version`) — NO tablas normalizadas por nodo/edge.
Serializa/deserializa con `encoding/json`, el mismo formato que `decodeGraph`
(`internal/adapters/index/store.go:167-173`) ya usa para los grafos embebidos.

**Por qué:** los 3 métodos de `IndexPort` que escriben/leen (`Upsert`, `Query`, `List`) operan
**siempre sobre el `Graph` entero** — ningún caller real hace `WHERE` sobre un nodo o edge
individual. Confirmado por los callers reales (`hallazgos.md` §1): `map_service.go`,
`run_service.go:113`, `fuente_service.go:45` hacen `Query`/`List` de solo lectura del grafo
completo; `portafolio.go:316` y `session_reindex.go:57` hacen `Upsert` del grafo completo. Cero
caller pide «dame los nodos de clase `hook` de todos los arneses» o similar. Normalizar hoy es
la regla 1 del ladder violada al revés (construir para un `WHERE` que nadie pidió). Si el Mapa
alguna vez necesita búsqueda cross-arnés granular (p.ej. «qué arneses usan la skill X»), se
normaliza entonces, con ese caller real como spec.

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).

## D5 · Watcher fsnotify — directorio(s) a observar + reindex incremental (no `Rebuild()` completo)

**Qué:** dos sub-decisiones:

1. **Directorio(s):** el watcher observa el **working-dir de cada arnés conocido por
   `ports.ArnesRegistry`** (`internal/adapters/store/arnes_registry.go`, CAP-24 — el mismo
   registro `id→path` que YA usa el boot de hoy en `main.go:148-154` para poblar el índice). NO
   observa `~/.claude/projects` (eso es el corpus de transcripts, adapter `history` aparte,
   confirmado fuera de alcance por `hallazgos.md` §1). Ver D6 para por qué es `ArnesRegistry` y
   no el store del Portafolio.
2. **Qué dispara un evento:** el TODO literal de `main.go:300-305` (`idx.Rebuild(ctx)` +
   `broker.Publish(sse.EventMap, delta)`) **NO es el comportamiento correcto** — se reemplaza
   por **reindex incremental de SOLO el arnés dueño del path que cambió**, reusando el patrón ya
   vivo de `internal/usecase/session_reindex.go` (`NewTurnReindexer`): resolver qué entrada de
   `ArnesRegistry` es prefijo del `WatchEvent.Path` → `loader.LoadArnes(esa-path)` →
   `idx.Upsert(ctx, esa-id, g)` (degradado si `LoadArnes` falla) → `broker.Publish(sse.EventMap,
   {harness_id: esa-id})`. Un evento bajo un path que no pertenece a ningún arnés registrado se
   ignora (no Upsert, no publish).

**Por qué:** `Rebuild()` completo en cada evento fsnotify es correcto pero desperdiciado — el
corpus es finito y chico (D1), pero un `git checkout`/`pnpm install`/edición masiva dispara
docenas de eventos fsnotify en ráfaga, y un `Rebuild()` por cada uno recargaría TODOS los
arneses registrados por CADA archivo tocado de UNO solo, además de publicar `event: map` sin
`harnessId` (el TODO deja `delta` sin definir — ¿de qué shape?). El patrón de
`session_reindex.go` ya resuelve exactamente este problema (reindex de un arnés, publish
scoped con `harness_id`) y el FE YA sabe consumir ese evento (`RF-186/187`,
`2026-07-22-mejorar-arnes-conversando/spec.md`) — reusar la forma existente es la opción más
lazy que sigue siendo correcta (ladder regla 2: patrón que ya vive en el codebase).

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).

## D6 · `Rebuild()` reconstruye desde `ArnesRegistry`, NO desde el store del Portafolio — hallazgo propio

**Qué:** no estaba en la lista de 5 preguntas de la consigna, pero investigar D5 (qué observa el
watcher) obliga a resolverlo primero: `hallazgos.md` §2 sugiere «`Rebuild()` real =
`Listar()`/`Escanear()` del Portafolio → por cada entrada, `Load(dir)` → `Upsert`» como insumo,
sin resolver una ambigüedad real que ese diseño tiene. Decido que **`Rebuild()` reconstruye
desde `ports.ArnesRegistry.List()`** (`arneses.json`, CAP-24), el registro simple `id→path`
1:1 — el MISMO mecanismo que `main.go:148-154` ya prueba en producción hoy (boot repuebla el
índice iterando `arnesReg.List()`). El store del Portafolio (`portafolio.json`, CAP-25, modelo
N:M:M de HS-22/23) sigue alimentando el índice **solo** vía el mecanismo YA existente y
explícito `PortafolioService.ObservarEnMapa` (`internal/usecase/portafolio.go:257-320`,
per-instalación, disparado por el usuario/FE) — sin cambios, fuera de este paquete.

**Por qué:**

- **Es lo que YA funciona.** `ArnesRegistry` es exactamente el shape que `Rebuild()` necesita
  (1 id → 1 path, sin ambigüedad); mover ESE loop de `main.go` adentro de `Store.Rebuild()` es
  la mecanización mínima de un comportamiento ya probado — no un rediseño.
- **El Portafolio tiene una ambigüedad real que `ArnesRegistry` no tiene.** Una
  `EntradaPortafolio` puede tener **N `Instalaciones`** para la misma identidad (multi-copia,
  spec de Slice 0). `Rebuild()` indexando «todas las entradas conocidas» tendría que inventar
  una política no pedida por nadie («¿cuál instalación gana? ¿el canónico si existe? ¿todas bajo
  la misma clave, pisándose?») — el propio código de `ObservarEnMapa` deja esa elección
  EXPLÍCITA al caller (`clave, installPath` como parámetros), precisamente para no inventarla en
  silencio. Automatizar eso dentro de `Rebuild()` violaría el mismo espíritu anti-silencio que
  `ObservarEnMapa` ya documenta en su comentario (`portafolio.go:257-268`).
- **Separación de fronteras ya escrita en el código.** El comentario de `ObservarEnMapa` dice
  explícitamente «jamás toca `arneses.json`/`ArnesRegistry`» — los dos registros son fronteras
  deliberadamente separadas (confinamiento de sesión vs. inventario de identidad/deriva); hacer
  que `Rebuild()` mezcle el Portafolio ahí adentro difumina esa frontera que HS-22/23 ya
  construyó a propósito.
- **Consecuencia aceptada, documentada:** una entrada del Portafolio «observada» en el Mapa
  (vía `ObservarEnMapa`) **no sobrevive un reinicio del daemon** hoy ni con este paquete —
  el índice es desechable y `Rebuild()` no la re-observa automáticamente. Es un gap conocido,
  no silenciado: si producto quiere que el Portafolio persista su propia presencia en el índice
  entre reinicios, es un paquete futuro (extender `Rebuild()` o encadenar `ObservarEnMapa` a un
  evento de boot) — no se resuelve por goteo acá.

**Estado:** DECIDIDA (agente, autónoma por /goal 2026-07-24).
