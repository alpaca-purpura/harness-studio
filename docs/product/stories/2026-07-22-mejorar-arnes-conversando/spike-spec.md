# Spike — Mejorar un arnés conversando (chat + Mapa en vivo)

> Paquete `2026-07-22-mejorar-arnes-conversando`. Este documento es autocontenido: no asume que quien
> lo lee vio la conversación original. Etapa: spike de spec (resolver el Fork A → firma → `spec.md`).

## 0. Norte

El operador quiere UNA experiencia, no dos historias separadas: conversás con Claude Code para
mejorar/reparar/agregar algo al arnés de tu sesión, y mientras lo hacés el Mapa (la forma visual de
ArnesIA de ver arneses) va reflejando esos cambios — click en un nodo, ves el contenido actualizado.
Hoy eso NO pasa: el chat edita el árbol real, pero el Mapa sirve una foto vieja cacheada.

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

## 2. Decisiones ya firmadas esta sesión (2026-07-22, con el operador)

1. **Alcance de "mejorar un arnés" = chat libre + doctrina de fondo.** Le pedís lo que quieras (agregar
   función, arreglar bug, refactor); Claude Code edita directo guiado por el kit ② (§1.2, ya inyectado
   hoy). Descartado: menú guiado de comandos fijos (eso sería revivir forja-ciclo-vivo item 3, fuera de
   alcance de este paquete).
2. **Frecuencia de refresh del Mapa = después de cada turno de Claude.** Validado en vivo: el costo de
   `loader.LoadArnes` es despreciable (§1.3).

## 3. Fork A — Instalación editable durante el chat (ABIERTO, a firmar antes de `spec.md`)

**El problema:** el picker ya firmado deja abrir sesión de escritura contra una instalación (copia
read-only por doctrina). No hay guard. Contradice la ley anti-drift firmada.

**Opción A1 (recomendada) — Bloquear edición de instalaciones en el chat.** El chat de escritura solo
abre contra el CANÓNICO. Si el usuario elige una instalación en el picker, la sesión abre en un modo
sin escritura real (observar/auditar), o el picker directamente no ofrece instalaciones como destino de
"nueva sesión de chat" (solo las ofrece "Abrir en Mapa", que ya es observación read-only, cierra GAP-1
del Portafolio). Mantiene la ley anti-drift tal como se firmó — no la reabre. Costo: hay que decidir qué
ve el usuario si intenta elegir una instalación (¿la oculta la lista? ¿la muestra deshabilitada con
tooltip, mismo patrón que Reparar/Backport S1-D11?).

**Opción A2 — Permitir con aviso explícito.** Se deja editar, pero la sesión muestra un banner
permanente: "estás editando una copia local (`<home>/<id>`), no vas a poder Publicar sin sincronizar
primero con el canónico". No bloquea, delega la responsabilidad. Riesgo: el usuario edita y el cambio
queda huérfano — sin Reparar/Backport construido, ese trabajo no tiene cómo volver al canónico ni al
`home`, se pierde conceptualmente (el `deriva` de esa instalación queda divergiendo para siempre a menos
que alguien lo note).

**Opción A3 — Descartada por el operador esta sesión** (dejar sin resolver indefinidamente): ya se
decidió que el fork se resuelve DENTRO de este paquete, antes de `spec.md` — no se pospone otra vez.

**Recomendación de este spike: A1.** Razón: A2 crea trabajo huérfano real (peor que bloquear) mientras
Reparar/Backport no exista; construir A1 es más chico (es un filtro en el picker + un chequeo server-side
en `ArnesRegistry`, reusa el patrón ya existente de "protected paths") que diseñar un banner+flujo de
recuperación para trabajo que puede perderse.

## 4. Plan técnico (una vez firmado el Fork A)

1. **Guard de escritura canónico-only** (resuelve Fork A si se firma A1): `ArnesRegistry`/
   `PortafolioService` exponen si un `path` es canónico o instalación (el dominio YA tiene
   `EntradaPortafolio`/`IdentidadArnes` con esa distinción, S0-D3/S1-D1); `createSession` rechaza (o
   redirige) si el path es una instalación. FE: el picker filtra o deshabilita instalaciones como
   destino de "nueva sesión" con tooltip, mismo patrón visual que Reparar/Backport (S1-D11).
2. **Reindex-tras-turno:** al final de cada turno del conductor (`internal/adapters/agent/claudecode/
   conductor.go`, donde hoy se resuelve `control_response`/Stop), disparar `loader.LoadArnes(cwd de la
   sesión)` → `MapService`/`IndexPort.Upsert(ctx, g)`. Si `LoadArnes` falla (el chat rompió el
   manifiesto), el Upsert debe reflejar el estado DEGRADADO real (mismo patrón `manifiesto-ausente` de
   Slice 2), nunca ocultar el error ni servir la foto vieja como si nada.
3. **Push al FE:** publicar un evento nuevo por `/events` (reusa el broker SSE existente, ver
   `EventPublisher` en `internal/usecase/run_service.go`) — p.ej. `event: reindexado` con el `harnessId`.
   El Mapa (si está montado, viendo ese arnés) refetchea `GET /api/harnesses/{id}/graph` al recibirlo.
4. **Feedback de "mejoró de verdad":** evaluar si conviene correr conformance automáticamente tras el
   reindex (no fue parte de las 2 decisiones firmadas hoy — el operador eligió chat libre, no el
   híbrido con auditoría automática) — dejar CHICO en V1: solo reindex+push, sin conformance automático;
   si el operador lo quiere después, es una extensión, no un rediseño.

## 5. Plan de tickets (tentativo, ajustar al firmar Fork A + `spec.md`)

- **T1** — Guard canónico-only en `ArnesRegistry`/`createSession` + filtro/disabled en el picker
  (resuelve Fork A si A1).
- **T2** — Reindex-tras-turno en el conductor (Go, con test que reproduce edición real → grafo
  actualizado, incl. caso degradado).
- **T3** — Evento SSE `reindexado` por `/events` + consumo en el Mapa (FE refetch on event).
- **T4** — E2E vivo: sesión real contra un arnés real, pedirle a Claude un cambio concreto (agregar un
  campo a un manifiesto, por ejemplo), confirmar que el Mapa lo muestra sin recargar la página a mano.
- **T5** — Capabilities: nueva o extendida (`fe-shell`/`portafolio`, a decidir en `spec.md`) + cifras
  regeneradas.

## 6. Lo que sigue

**Falta SOLO** la firma del Fork A (recomendación A1 arriba) para pasar a `spec.md` con RF numerados.
Sin eso, no se escribe código — el operador decide en la próxima sesión o ahora mismo si quiere seguir.
