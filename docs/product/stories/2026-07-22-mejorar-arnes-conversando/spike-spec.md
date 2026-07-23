# Spike — Mejorar un arnés conversando (chat + Mapa en vivo)

> Paquete `2026-07-22-mejorar-arnes-conversando`. Este documento es autocontenido: no asume que quien
> lo lee vio la conversación original. Etapa: spike de spec **CERRADO** — refinado 2026-07-22 contra
> el código real (mapa file:line); **Forks A/B/C + grounding FIRMADOS 🧑‍⚖️** — ver `decisiones.md`
> (MC-D3..MC-D8). Sigue: `spec.md` con RF numerados.

## 0. Norte

El operador quiere UNA experiencia, no historias separadas: conversás con Claude Code para
mejorar/reparar/agregar algo al arnés de tu sesión, y mientras lo hacés el Mapa (la forma visual de
ArnesIA de ver arneses) va reflejando esos cambios — click en un nodo, ves el contenido actualizado.
Hoy eso NO pasa: el chat edita el árbol real, pero el Mapa sirve una foto vieja cacheada. **Ampliación
del norte (misma sesión, 2026-07-22):** además, la conversación con un arnés tiene que sentirse
INFINITA — poder recuperar el historial de conversaciones pasadas de ese arnés, poder arrancar una
conversación nueva sobre el mismo arnés, y que el sistema rote de contexto por detrás (silencioso, sin
perder continuidad) cuando el uso de tokens se acerca a un umbral — en vez de que la sesión se rompa o
el usuario tenga que gestionar eso a mano.

## 1. Estado real (verificado contra código, 2026-07-22)

### 1.1 El chat ya escribe sobre el arnés real

`chat-cc-funcional` (`docs/product/stories/2026-07-08-chat-cc-funcional/`, firmado HS-20) ya resuelve
sesión↔cwd real, permisos/scope/gate/Stop, y su propia historia incluye el escenario "modificación de
arneses existentes". El cwd de cada sesión es el `installPath` real que el usuario eligió (no una
copia) — `internal/ports/workdir.go` + `internal/adapters/store/arnes_registry.go`.

### 1.2 El kit ② YA enseña "qué es un buen arnés" — no hay que inventar la doctrina

Cada sesión de chat recibe, vía `internal/adapters/provision/provisioner.go` (materializa a
`~/.arnesia/{kit,doctrine.md,knowhow,mcp.json}`) + `internal/adapters/agent/claudecode/conductor.go`
(`SpawnArgs` → `--plugin-dir`/`--append-system-prompt-file`/`--add-dir`):

- `kit/doctrine.md` (embebido vía `embed_doctrina.go`, `go:embed all:kit`): anatomía A1-A7, el spine que
  el arnés declara (`arnes.l0.json`), el contrato de caja fusionado, la clase de 10 primitivas
  CC-native, la nomenclatura.
- Skills `kit/skills/auditar-arnes/SKILL.md` y `kit/skills/forjar-caja/SKILL.md` — operacionalizan esa
  doctrina en acciones concretas.

Esto **NO es el kit `harness@prenter-marketplace`** (ese es un plugin externo, instalado en el
`.claude/settings.json` de ESTE repo solo para desarrollar ArnesIA — vive fuera del árbol, en
`prenter-marketplace/`, y no viaja a las sesiones de chat de los arneses de los usuarios).

El motor de conformance (kit ①, `go:embed` del ruleset) NO viaja como contenido de sesión — corre
server-side: CLI `arnesia conformance` (`cmd/arnesia/conformance.go`) o
`GET /api/harnesses/{id}/conformance` (`router.go`). La skill `auditar-arnes` sabe invocarlo por `Bash`,
pero `Bash` está gateado por permiso de rol (`Ask` o `Deny` según `internal/adapters/permission/
provisioner.go`) — no es un tool/MCP directo, así que "correr conformance" durante el chat depende de
que el rol lo permita y de que Claude decida invocarlo.

### 1.3 El Mapa no se entera de nada (el gap real #1)

`IndexPort` (`internal/ports/index.go`) es un índice **desechable**: `Query`/`List` leen lo que haya
cacheado, `Upsert` lo reemplaza, `Rebuild` reconstruye TODO desde cero. Nada dispara un `Upsert` del
arnés de la sesión activa después de que el chat le edita archivos. Resultado: aunque volvés a navegar
al Mapa, seguís viendo el grafo de cuando se cargó por última vez (boot / "Cargar carpeta" / Abrir en
Mapa desde el Portafolio).

`loader.LoadArnes(dir)` (el mismo loader que arma el grafo) es barato — medido en vivo esta sesión:
**~440µs** contra el fixture dogfood (`dogfood/dev-full-cycle`, 5 nodos/4 edges) y **~75µs** contra un
plugin real completo (`~/.claude/plugins/cache/prenter-marketplace/harness/0.5.3`, con skills+agents+
hooks+rules). Reindexar tras cada turno no tiene costo de performance relevante.

`/events` (`router.go`) ya es un broker SSE multiplexado — el Dock lo usa hoy para streamear cada turno
(`event: run`). Es el canal natural para empujar "este arnés se reindexó, aquí el grafo nuevo".

### 1.4 Instalación editable — contradicción real entre 2 paquetes ya firmados (el gap real #2)

- **`2026-07-10-spike-carga-arneses`** (modelo del Portafolio, firmado 2026-07-13): la ley anti-drift —
  "ArnesIA nunca introduce una copia editable divergente nueva" — está en `BACKLOG.md` y en
  `spec-funcional.md` (INV-1). `casuistica.md` de ese paquete dice explícito: *"Editar instalación
  directo ⛔ No existe ese camino... la UI no lo ofrece."* Canónico = única copia editable;
  instalaciones = espejos READ-ONLY que se sincronizan por Reparar/Backport (S1-D11, **sin construir
  todavía** — solo botones disabled+tooltip "próximo · S5", cero lógica de dominio, grep negativo en
  `internal/`).
- **`2026-07-20-shell-topbar-selector-arnes`** (picker de sesión, firmado HOY mismo, HS-25): TS-D15/
  TS-D16 (`decisiones.md`) formalizan que el picker deja elegir CUALQUIER copia del arnés —
  canónico O instalación — como origen de una sesión nueva, **sin mencionar INV-1**.
- **Verificado en código:** `new-session-picker.tsx` (`copiasDe()`) mezcla canónico + instalaciones sin
  distinguir; `crear()` manda el `path` de la copia elegida tal cual. Backend: `sessions.go#createSession`
  → `ArnesRegistry.Register(arnes, path)` → `validate()`/`checkProtected()`
  (`internal/adapters/store/arnes_registry.go`) solo chequea que el path sea absoluto/exista/no esté en
  una carpeta protegida (`~/.claude`, `~/.ssh`, etc.) — **cero chequeo de canónico-vs-instalación.**
  `WorkdirResolver.Resolve` (`internal/ports/workdir.go`) confía ciegamente en ese registro.

**Esto es un hueco real, no una decisión tomada.** Si el operador abre chat contra una instalación hoy
y le pide a Claude que "mejore" algo, el chat va a escribir ahí — violando la ley anti-drift firmada,
sin ningún aviso. El operador decidió (2026-07-22) **dejarlo abierto a propósito** en vez de resolverlo
a ciegas en esta conversación — ver Fork A abajo.

### 1.5 Historial de conversación — YA existe, pero se BORRA al cerrar (el gap real #3)

ArnesIA sí persiste un transcript propio por sesión: `domain.Session.Conv []Turn`
(`internal/domain/session.go:59-95`, comentario explícito: "NOT the source of truth — Claude Code's own
JSONL is") se llena turno a turno (`session_service.go:275,393`) y el `Session` completo (con `Conv`) se
serializa a `~/.arnesia/sessions.json` (`internal/adapters/store/registry.go`, escritura atómica). Esto
sobrevive reinicios del daemon y se expone real por `GET /api/sessions/:id` — **no es solo memoria del
navegador**, el FE (`sessions-store.ts`) se rehidrata de ahí.

`Session.Arnes` (`session.go:73`, comentario: *"N sessions may share one [arnés]"*) YA agrupa sesiones
por arnés — o sea, "todas las conversaciones de este arnés" es un filtro trivial sobre `svc.List()`
(`GET /api/sessions`), no hace falta inventar un modelo de agrupación nuevo.

**El hueco real: `SessionService.Close(id)` (`session_service.go:229-248`) BORRA la entrada del registro
por completo** ("ends a session: ... drops the registry entry") y la persiste así (sin `Conv`) al
archivo. Hoy, cerrar una sesión desde el rail = perder su historial para siempre en ArnesIA. La JSONL
nativa de Claude Code (`~/.claude/projects/<hash>/<id>.jsonl`) técnicamente sigue en disco (fuera del
control de ArnesIA), pero el indexer que la leería está sin construir — `internal/adapters/index/
store.go` tiene el TODO literal ("walk the ~/.claude JSONL corpus and rebuild from it") y
`internal/adapters/watch/watcher.go` es un stub no-op (fase 5, mismo bloqueo que el reindex del Mapa).
Hoy, entonces, **"obtener el historial" solo funciona mientras la sesión sigue abierta.**

### 1.6 Uso de contexto por turno — YA se calcula, solo no se historiza ni dispara nada

Cada frame `result` que devuelve el proceso `claude` trae `usage` real (`input_tokens`,
`cache_read_input_tokens`, `cache_creation_input_tokens`) + `modelUsage[model].contextWindow`
(`internal/adapters/agent/claudecode/conductor.go:396-430`). ArnesIA YA calcula `ctxPct` desde eso
(`conductor.go:552-566`, ventana default 200k) y lo guarda en `Session.CtxPct` — **pero solo el último
valor, se pisa cada turno** (`session_service.go:395-396`); no hay histórico ni umbral configurable ni
acción disparada. Grep negativo total (código+docs) de "38%"/"45%"/compactación-por-umbral — la NOCIÓN
de contexto existe (`CAP-42`, `docs/architecture/knowledge/elements/statusline.md`), la FEATURE de
alerta/rotación no.

El conductor usa `--resume <ClaudeSessionID>` (`conductor.go:72-73`) SOLO para reconectar tras un
reinicio del daemon (mismo contexto, mismo proceso lógico) — nunca para "arrancar en limpio con un
resumen". `--continue` del CLI real: grep negativo, no se usa. El check `resume-auto-sana`
(`docs/architecture/boundaries/sesion-viva-consistente.md`) resuelve un proceso que muere ANTES de
emitir `init` tras un `--resume` — es recuperación de CRASH, no rotación de contexto deliberada.

### 1.7 Precedente real en la casa para "archivos chicos en vez de contexto gigante"

**Franja Artefactos** (`docs/product/stories/2026-07-07-franja-artefactos/`, firmado, capability
`fe-mapa/franja-de-artefactos.yaml`) ya resolvió — para el hand-off entre cajas del pipeline
spec→build→review — el MISMO problema de fondo que este Fork C: en vez de pasar el documento entero
entre etapas, pasa un **digest chico** (`internal/adapters/artifact/reader.go`) — medido en su momento:
**−90% de contexto por insumo** (75 tok de digest vs 750 del spec completo). Es la filosofía exacta que
el operador describe ("muchos pequeños archivos intermedios") — ya validada en este repo para OTRO
punto de la tubería. No hace falta inventar el patrón, hace falta aplicarlo acá.

## 2. Decisiones ya firmadas esta sesión (2026-07-22, con el operador)

1. **Alcance de "mejorar un arnés" = chat libre + doctrina de fondo.** Le pedís lo que quieras (agregar
   función, arreglar bug, refactor); Claude Code edita directo guiado por el kit ② (§1.2, ya inyectado
   hoy). Descartado: menú guiado de comandos fijos (eso sería revivir forja-ciclo-vivo item 3, fuera de
   alcance de este paquete).
2. **Frecuencia de refresh del Mapa = después de cada turno de Claude.** Validado en vivo: el costo de
   `loader.LoadArnes` es despreciable (§1.3).

## 3. Fork A — Instalación editable durante el chat (RESUELTO 2026-07-22: A4 FIRMADO 🧑‍⚖️, MC-D8)

**El problema:** el picker ya firmado deja abrir sesión de escritura contra una instalación (copia
read-only por la letra de INV-1). No hay guard.

**Descartadas por el operador (2026-07-22):** A1 (bloquear: chat de escritura solo contra el canónico)
y A2 (permitir con banner de advertencia). Ninguna de las dos captura el modelo real.

**Dirección del operador (condensada de sus palabras):** al elegir una instalación, ArnesIA detecta
inmediatamente de qué arnés y de qué marketplace viene. La instalación ES el banco de pruebas — «¿cómo
pretendes crear un arnés sin probarlo?». Toda instalación fresca pasa por una etapa de MAPEO del
proyecto (sea de código o una carpeta de trabajo normal): crear los elementos necesarios para que el
arnés funcione ahí correctamente («recablear»). Un arnés que anda mal en un proyecto tiene dos causas
posibles: (a) la instalación/mapeo fue mala → se corrige IN SITU, como si se reinstalara y recableara;
(b) el arnés está mal diseñado de base → el fix se LEVANTA al canónico del marketplace (backport) y
LUEGO se prueba sobre la instalación ya hecha para ponerlo en claro. En ambos casos se edita sobre el
arnés del proyecto; lo que cambia es a dónde viaja el aprendizaje — siempre detectando POR QUÉ falló
para corregir la base para nuevas instalaciones.

**Síntesis A4 (FIRMADA 🧑‍⚖️ 2026-07-22, MC-D8):** sesión sobre instalación = **sesión de REPARACIÓN**,
legal. La ley anti-drift se PRECISA, no se tumba: lo prohibido es la deriva **silenciosa/huérfana**, no
la edición. INV-2 («reconciliación, no prevención» + backport) ya contenía este modelo; lo que se
enmienda es la frase absoluta de INV-1 «ninguna instalación se edita en sitio desde la app» →
**«ninguna instalación deriva en silencio»**. Condiciones que hacen legal la edición in situ:

1. **La sesión SABE que edita una instalación** (tarjeta de identidad MC-D6): procedencia `(home,id)`,
   quién es su canónico, estado `deriva`; la doctrina ② enseña el loop diagnosticar → arreglar in situ
   → identificar la causa (¿instalación mala o base mala?) → backport si es base.
2. **La deriva queda VISIBLE tras cada turno:** el reindex-tras-turno recalcula (grafo + `deriva` de
   esa entrada) y el Portafolio muestra `en-deriva` honesto — jamás finge que la copia sigue al-hilo.
3. **El aprendizaje no queda huérfano:** cuando la causa es de base, el hallazgo queda marcado
   pendiente-de-backport. V1 de este paquete: deriva visible + grounding de reparación; la cola formal
   de backport es Reparar/Backport (S5), que este modelo ALIMENTA en vez de contradecir.

**Fuera de alcance de este paquete:** la etapa de MAPEO completa de una instalación fresca (eso es el
inicializador universal de la realineación core, paquete propio futuro).

## 3b. Fork B — Historial de conversación que sobrevive el cierre (RESUELTO 2026-07-22: B2 FIRMADO 🧑‍⚖️)

**El problema (§1.5):** `Close()` borra el `Conv` de la sesión del registro. El operador pidió poder
"obtener el historial de las conversaciones por cada sesión" — hoy eso deja de ser cierto en cuanto
cerrás la pestaña.

**Opción B1 (recomendada) — Archivar en vez de borrar.** `Close()` mueve la sesión (con su `Conv`
completo) a un archivo append-only separado (p.ej. `~/.arnesia/sesiones-archivadas.jsonl`, una línea
por sesión cerrada) ANTES de sacarla del registro activo. Nuevo endpoint `GET /api/arneses/{id}/
historial` (o query param sobre `/api/sessions`) lista TODAS las conversaciones de un arnés — abiertas
+ archivadas — con metadata liviana (fecha, `Frente`, cantidad de turnos) para no traer el `Conv`
completo de todas de una. "Agregar una nueva conversación" ya existe hoy (el mismo `+ Nueva sesión` del
picker, creando otra `Session` con el mismo `Arnes`) — solo falta la vista que las lista junto a las
archivadas.

**Opción B2 — Confiar en el JSONL nativo de Claude Code.** No archivar nada propio; en su lugar
construir el indexer que hoy es TODO (`internal/adapters/index/store.go`) para leer
`~/.claude/projects/<hash>/*.jsonl` directamente. Más "correcto" a largo plazo (esa SÍ es la fuente de
verdad completa, con tool-calls y todo) pero es un proyecto mucho más grande (fase 5, ya bloqueado por
motivos ajenos a este spike) — no es proporcional a lo que este paquete necesita resolver ahora.

**Resolución (2026-07-22, MC-D7): el operador FIRMÓ B2** — el camino correcto (la JSONL nativa es la
fuente de verdad completa), sabiendo que es más grande que archivar. Alcance mínimo viable acotado en
`decisiones.md` MC-D7: metadata liviana al `Close()` (incluida la **cadena de `ClaudeSessionID`s** —
con la rotación del Fork C una conversación lógica = N JSONLs, sin ese join no hay cosido) + lector del
corpus `~/.claude/projects/<hash-del-cwd>/*.jsonl` por arnés. El indexer completo fase-5 (SQLite,
watcher) sigue diferido. ⚠ Los turnos `RolSys` propios de ArnesIA no están en la JSONL — ver el
constraint en MC-D7.

## 3c. Fork C — Rotación de contexto invisible por umbral de tokens (RESUELTO 2026-07-22: FIRMADO 🧑‍⚖️, MC-D5)

**El pedido del operador:** cuando el uso de contexto de la sesión llega a ~38-45%, el sistema tiene que
—por detrás, sin que el usuario lo note ni tenga que actuar— dejar un rastro chico, arrancar un proceso
`claude` NUEVO (contexto limpio) y seguir la conversación como si fuera la misma, sin perder continuidad.
Es, en espíritu, el mismo mecanismo que ESTE propio Claude Code usa para compactar contexto largo
(mencionado en las instrucciones de sistema de esta sesión) — pero el operador lo quiere MÁS proactivo
(38-45%, muy por debajo del ~92% típico de auto-compact) y con artefactos explícitos, no un resumen
opaco.

**Mecanismo propuesto (usa infra que YA existe, ver §1.6/§1.7 — nada de esto es una pieza nueva rara):**

1. **Disparador:** cada turno ya calcula `ctxPct` real (§1.6). Agregar un umbral configurable
   (default a decidir en `spec.md`, rango 38-45% pedido) — cuando `ctxPct` lo cruza, se dispara la
   rotación DESPUÉS de responder el turno en curso (nunca a mitad de una respuesta).
2. **Checkpoint chico:** antes de rotar, escribir un artefacto liviano (mismo espíritu que Franja
   Artefactos §1.7 — mecánico primero, no un resumen por LLM si un digest estructurado alcanza:
   últimos N turnos + qué se tocó en el arnés desde el último checkpoint + próximo paso pendiente).
   Vive junto al resto de artefactos de la sesión (a definir en `spec.md` — ¿en `~/.arnesia/` o dentro
   del propio árbol del arnés?).
3. **Rotar el proceso, no el `Session`:** spawnear un `claude` **fresco** (SIN `--resume` — contexto
   nuevo de verdad) inyectando el checkpoint vía `--append-system-prompt-file` (la MISMA vía que ya
   inyecta el kit ②, §1.2 — infra reusada, no nueva). El `domain.Session.ID` (el de ArnesIA, lo que ve
   el usuario) **no cambia** — solo rota el `ClaudeSessionID` interno. `Conv` (§1.5) sigue creciendo
   sin cortes — a nivel ArnesIA nunca se pierde nada, aunque el proceso `claude` de abajo haya arrancado
   de cero.
4. **Transparencia mínima:** un breadcrumb `RolSys` (`domain.Turn`, ya existe el rol para esto — se usa
   hoy para "skill_activated"/hand-offs) en el `Conv`, tipo "— contexto rotado, seguimos —", para que si
   el usuario mira el historial entienda que pasó algo, sin que sea una interrupción real de la
   conversación.

**Resolución (2026-07-22, MC-D5) — las 4 sub-preguntas cerradas:**
- **Umbral: default 40 %, configurable** (centro del rango 38-45 pedido).
- **Formato: mecánico primero** (últimos N turnos + delta del grafo desde la última rotación + paso
  pendiente); pasada LLM solo si el mecánico demuestra no alcanzar en la práctica.
- **Vive en `~/.arnesia/sessions/<id>/`** (lado store): firewall ②↛③ — la inyección jamás escribe en
  el árbol del arnés; si viviera en el árbol, el reindex-tras-turno lo mostraría como nodo del Mapa,
  conformance lo evaluaría y Publicar lo arrastraría al marketplace.
- **«Rotar a mitad de edición» no existe por construcción:** al `EventResult` que cruza el umbral se
  cierra el proceso y se vacía `ClaudeSessionID` (la sesión ya quedó `idle`); el spawn fresco ocurre
  recién al próximo `Turn` (`spawnLocked` ya es lazy). Detalle: hoy `spawnLocked` SIEMPRE pasa
  `Resume=ClaudeSessionID` — rotar es vaciar ese campo + inyectar el checkpoint.

## 4. Plan técnico (refinado 2026-07-22 contra el código; decisiones en `decisiones.md`)

1. **Sesión de reparación sobre instalación** (Fork A/A4 firmado): el picker deja de bloquear y
   pasa a ROTULAR — elegir una instalación abre la sesión etiquetada «reparación sobre instalación de
   `<home>`»; la tarjeta de identidad (punto 7) le dice a Claude que edita una instalación, quién es
   su canónico y su `deriva`; el reindex (punto 2) recalcula la deriva de esa entrada tras cada turno
   y el Portafolio la muestra honesta. Sin guard server-side bloqueante; la cola formal de backport
   llega con Reparar/Backport (S5).
2. **Reindex-tras-turno (MC-D3):** en `SessionService.consume`, rama `EventResult`
   (`internal/usecase/session_service.go:385-404` — donde ya se guarda `CtxPct` y se appendea `Conv`):
   `loader.LoadArnes(cwd)` → `IndexPort.Upsert`. NO en el conductor (boundary
   `adaptadores-de-agente-intercambiables` — el adapter es protocolo puro, no conoce `IndexPort`). Si
   `LoadArnes` falla (el chat rompió el manifiesto), el Upsert refleja el estado DEGRADADO real
   (patrón `manifiesto-ausente` de Slice 2 / `ObservarEnMapa`), nunca la foto vieja.
3. **Push al FE (MC-D4):** publicar por el canal YA reservado `event: map` (`broker.go:18`,
   `EventMap` sin uso desde el diseño original `map|dock|run`) con el `harnessId`; `sse.ts` agrega el
   listener (hoy solo escucha `dock`) y el Mapa refetchea `GET /api/harnesses/{id}/graph`.
4. **Feedback de "mejoró de verdad":** V1 chico — solo reindex+push, sin conformance automático
   (extensión futura, no rediseño).
5. **Historial (MC-D7, B2):** `Close()` persiste metadata liviana (sin `Conv`): identidad de la
   sesión, `Arnes`, cwd, fechas, cadena de `ClaudeSessionID`s. Lector del corpus
   `~/.claude/projects/<hash-del-cwd>/*.jsonl` por arnés + listado de conversaciones
   (abiertas+cerradas) filtrando por `Session.Arnes`.
6. **Rotación de contexto (MC-D5):** historizar `ctxPct` + umbral 40 % configurable → al cruzarlo en
   `EventResult`: checkpoint mecánico en `~/.arnesia/sessions/<id>/` + cerrar proceso + vaciar
   `ClaudeSessionID` + breadcrumb `RolSys`; el próximo `Turn` spawnea fresco con el checkpoint vía
   system-prompt-file.
7. **Tarjeta de identidad por sesión (MC-D6):** el provisioner materializa un system-prompt-file POR
   SESIÓN (doctrina ② + identidad `(home,id)` + canónico-vs-instalación + `deriva` + degradado). Hoy
   `Injection.SystemPromptFile` apunta al `doctrine.md` compartido
   (`internal/adapters/provision/provisioner.go`). Misma infra que inyecta el checkpoint de rotación —
   un mecanismo, dos usos.

## 5. Plan de tickets (refinado; a cerrar en `spec.md`)

- **T1** — Sesión de reparación (Fork A/A4): picker rotula instalaciones (sin bloquear) + etiqueta de
  sesión «reparación» + deriva recalculada visible en el Portafolio tras editar.
- **T2** — Reindex-tras-turno en `SessionService.consume` (Go, con test que reproduce edición real →
  grafo actualizado, incl. caso degradado).
- **T3** — Evento `map` por `/events` + listener en `sse.ts` + refetch del Mapa.
- **T4** — E2E vivo: sesión real contra un arnés real, pedirle a Claude un cambio concreto (agregar un
  campo a un manifiesto, por ejemplo), confirmar que el Mapa lo muestra sin recargar la página a mano.
- **T5** — Capabilities: nueva o extendida (`fe-shell`/`portafolio`, a decidir en `spec.md`) + cifras
  regeneradas.
- **T6** — Historial B2 (MC-D7): metadata al `Close()` (cadena de `ClaudeSessionID`s incluida) +
  lector JSONL por arnés + FE: lista de conversaciones del arnés (abiertas+cerradas) desde el picker o
  el drawer.
- **T7** — Historización mínima de `ctxPct` (no solo el último valor) + umbral configurable (default
  40 %) + el disparador de rotación (MC-D5).
- **T8** — Checkpoint mecánico en `~/.arnesia/sessions/<id>/` (mismo patrón que Franja Artefactos
  §1.7) + spawn fresco vía system-prompt-file + breadcrumb `RolSys` en `Conv`.
- **T9** — E2E vivo de la rotación: sesión real, forzar el umbral (o bajarlo temporalmente para el test),
  confirmar que el usuario sigue viendo UNA conversación continua mientras el `ClaudeSessionID` rotó por
  detrás, y que `Conv` no perdió ningún turno.
- **T10** — Tarjeta de identidad por sesión (MC-D6): provisioner materializa system-prompt-file
  per-session (doctrina + identidad + estado portafolio); comparte infra con T8.

## 6. Lo que sigue

**Spike CERRADO — los 4 gates del enfoque FIRMADOS 🧑‍⚖️ (2026-07-22):** **Fork A = A4** (sesión de
reparación) · **Fork B = B2** (indexer JSONL nativo, alcance mínimo) · **Fork C** (rotación, mecanismo
completo + umbral 40 %) · **grounding MC-D6** (tarjeta de identidad por sesión). Sigue: escribir
`spec.md` con RF numerados y arrancar la implementación (T1-T10, §5) — sin decisiones de producto
pendientes.
