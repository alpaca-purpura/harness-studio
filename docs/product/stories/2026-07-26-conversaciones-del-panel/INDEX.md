# Paquete — Las conversaciones viven en el panel de conversación

> Origen: el operador (2026-07-26) pidió ver el historial de una conversación y crear una nueva.
> No encontró ninguna de las dos en el panel. La verificación en vivo le dio la razón: el
> historial existe pero está escondido en el picker del rail, y ahí muestra 0 por un bug de
> llave. Ver [`decisiones.md`](./decisiones.md).

## Estado

| Etapa | Estado |
|---|---|
| decisiones | ✅ **completas** — CV-D1..**CV-D16** 🧑‍⚖️, cero puntos abiertos (F-1..F-4 son correcciones de hecho) |
| relevamiento as-is | ✅ [`relevamiento-as-is.md`](./relevamiento-as-is.md) — 1453 líneas, 5 hallazgos críticos, todo con `archivo:línea` |
| mockup | 🧑‍⚖️ **FIRMADO 2026-07-26** (iteración 2) — [`mockup-conversaciones-panel.html`](./mockup-conversaciones-panel.html), 7 secciones / 16 paneles, registrado en [`mockups/INDEX.md`](../../../../mockups/INDEX.md) |
| spec + diseño | ✅ [`spec.md`](./spec.md) (RF-300…RF-357 · E-01…E-50 · H-1..H-9) + [`design.md`](./design.md) (C-1…C-13) + [`plan-storybook.md`](./plan-storybook.md) (56 stories) |
| **arquitectura** | ✅ [`arquitectura.md`](./arquitectura.md) — modelo · persistencia versionada · ciclo de vida · concurrencia · API + diff OpenAPI · FE · boundaries · **los 50 escenarios con su respuesta** · 9 huecos declarados |
| **plan de tickets** | ✅ [`plan-desarrollo.md`](./plan-desarrollo.md) — **33 tickets · 6 tramos** + cobertura E-01…E-50 → ticket |
| **plan de pruebas** | ✅ [`plan-pruebas.md`](./plan-pruebas.md) — pirámide · **circuito E2E contra la app instalada** (6 guiones) · datos de prueba aislados · 26 criterios de salida |
| GATE 2 🧑‍⚖️ (specs+arquitectura) | ✅ **AUTORIZADO POR DIRECTIVA 2026-07-26** — ver nota abajo |
| **implementar** | 🚧 **tramos 0 y 1 CERRADOS y verdes** (T1-T14) · tramos 2-5 sin empezar — ver [`PARIDAD.md`](./PARIDAD.md) §0, §1.8 y §1.9 |
| PARIDAD | 🚧 [`PARIDAD.md`](./PARIDAD.md) con la evidencia de los tramos 0 y 1; **gate 🧑‍⚖️ SIN FIRMAR** (sigue sin haber superficie visible que comparar) |

## GATE 1 🧑‍⚖️ — FIRMADO 2026-07-26

El operador firmó el mockup en su iteración 2 (la que baja el cromo a 2 filas, CV-D14/D15).
Transcripción de la firma, no auto-verificación: la firma es del operador, el ejecutor solo la
registra tras confirmar que el artefacto existe en el repo. Desbloquea spec + diseño +
arquitectura + build.

## GATE 2 🧑‍⚖️ — AUTORIZADO POR DIRECTIVA, no por lectura

Distinción honesta, porque no es lo mismo y el que audite tiene que saberlo: el operador **no leyó**
`spec.md` + `design.md` + `arquitectura.md` + `plan-desarrollo.md` antes de que empezara el build.
Lo que hizo fue dar, en el mismo turno en que firmó el Gate 1, la instrucción explícita de encadenar
**spec → arquitectura → desarrollo → auditoría** sin volver a consultarlo. Esa instrucción es la
autorización para tocar código; **no es una firma de contenido**.

Consecuencia práctica: la revisión de contenido de estos cuatro documentos **se corre hacia la
auditoría final y hacia el gate de PARIDAD**. Si el operador, al leerlos, rechaza algo, lo
construido sobre esa parte se rehace. El ejecutor no puede presentar esto como «spec firmado».

## Lo firmado hasta acá

- **CV-D1** — `Hist` = historial del arnés, no de la conversación. Fuera de alcance.
- **CV-D2** — crear · listar · buscar · abrir transcript: todo en el dock. Sale del picker del rail.
- **CV-D3** — **sesión CONTIENE N conversaciones**. Parte `Session` (`internal/domain/session.go:68`)
  en dos entidades.
- **CV-D4** — el dock lista SOLO las conversaciones de la sesión activa.
- **CV-D5** — la conversación cuelga de un `session_id`, no de un string de arnés.
- **CV-D6** — las 3 conversaciones **cerradas** de hoy se eliminan (procedimiento manual, T32).
- **CV-D7** — una conversación viva por sesión; crear cierra la anterior.
- **CV-D8** — el buscador busca el texto del transcript (`Conv`), que ahora **se persiste al cerrar**.
- **CV-D9** — título auto-derivado del primer mensaje, editable.
- **CV-D10** — la rotación por contexto es invisible: misma conversación, marca inline.
- **CV-D11** — seleccionar una inactiva la **retoma** (`--resume`); siempre una activa por sesión.
- **CV-D12** — vocabulario **activa/inactiva**, no «cerrada». El registro `sesiones-cerradas.json` pierde sentido.
- **CV-D13** — la fila muestra última interacción · nº turnos · ctx final. Exige timestamp nuevo por turno.
- **CV-D14** — el cromo del dock baja de **4 filas a 2**: ctx = chip-disclosure, identidad técnica
  detrás de un clic, alcance solo con nodo elegido.
- **CV-D15** — el glifo de colapsar pasa de `⟩` a `»`, el que el rail ya usa.
- **CV-D16** — las sesiones **vivas** con id pelado se **re-key** a clave calificada, no se borran.
  Paso 4 de la migración versionada, con respaldo y dos caminos de reversión.

## Radio de impacto (relevado, no estimado)

| Capa | Qué se toca |
|---|---|
| Dominio | `internal/domain/session.go:68` — partir `Session`; `ClaudeSessionID`/`Conv`/`CadenaCC`/`Checkpoint`/`CtxPct`/`CtxHist`/`RotacionPendiente` migran a `Conversacion` |
| Persistencia | `~/.arnesia/sessions.json` + `sesiones-cerradas.json` (registro `store.NewRegistry`, `cmd/arnesia/main.go:325`) — formato nuevo + migración de lo existente |
| API | `/api/sessions?arnes=…&cerradas=1` y `/api/sessions/cerradas/{id}/historial` (`web/src/shared/api/client.ts:214-225`) → endpoints por sesión |
| FE store | `web/src/shared/store/sessions-store.ts` + `web/src/widgets/session-rail/model/conversaciones-store.ts` (se muda al dock) |
| FE UI | `web/src/widgets/chat-dock/ui/chat-dock.tsx` (gana lista+buscador+＋); `new-session-picker.tsx:263` (pierde `ConversacionesDelArnes`) |
| Capabilities | ningún cambio de código sin capability (doctrina `codigo-traza-a-capability`) — CAP nuevas por definir en spec |

## Arquitectura as-code que este paquete deja en el árbol

Escrito en `docs/architecture/` porque **trasciende al paquete** (`arquitectura.md` §7.5 explica por
qué eso y no más, con los 3 candidatos descartados y su razón):

- **`boundaries/archivo-durable-declara-su-esquema.md`** (nuevo, `proposed`, 6 checks) — envelope
  versionado, migración forward-only con respaldo, cuarentena del corrupto, solo-lectura ante
  esquema futuro, y la **prohibición del wipe-and-rebuild sobre lo durable** (la política del índice
  es legítima para lo derivado e ilegítima acá).
- **`boundaries/ruta-servida-esta-declarada.md`** (nuevo, `proposed`, 5 checks) — ratchet
  router ⟷ OpenAPI con exención declarada. Drift medido hoy: **50 servidas · 39 declaradas · 11 sin
  declarar · 0 fantasma**.
- **`boundaries/sesion-viva-consistente.md`** v1.1 → **v1.2** (+1 check
  `transicion-de-conversacion-atomica`) — re-enuncia el sujeto: «la sesión viva» = la conversación
  activa. **Conserva `enforced`**: sus 4 checks siguen verdes; el quinto se declara pendiente.
- **`INDEX.md`** — 2 filas nuevas, la de `sesion-viva-consistente` corregida (estaba stale en 1.0/4),
  totales 26/136 → **28/148**.

⚠ Los dos nodos nuevos **nacen `proposed` y con TODOS sus enforcers sin escribir**: el código llega
con el paquete (T5, T9-T11, T15). Declararlos `enforced` antes sería el pass fabricado.

## Retomar aquí

### Dónde vive el trabajo del tramo 1 — LEELO PRIMERO

**El tramo 1 NO está en `main`.** Se construyó en un **worktree propio** para salir del bloqueo N-6
(dos constructores sobre el mismo working copy), y ahí sigue:

- **worktree:** `/home/chalreme/Proyectos/harness-studio/.claude/worktrees/agent-a12f73cab8f2b13b1`
- **rama:** `worktree-agent-a12f73cab8f2b13b1` (parte de `657054d`, que es `main`)
- **6 commits**, sin pushear y **sin integrar**

Integrarlo a `main` es un `merge --ff-only` mientras `main` no avance. Si avanzó, rebase.

### Lo hecho — TRAMOS 0 y 1 CERRADOS Y VERDES

| commit | ticket | qué |
|---|---|---|
| `bc9d1d4` … `657054d` | T1-T7 | tramo 0 completo + `domain.Conversacion` (ya en `main`) |
| `cd7571f` | **T8** | `Session` se parte: los 9 campos bajan · `Instantanea` (carrera real) · CAP-14 reescrita |
| `4cb4999` | **T9** | sobre versionado + cadena de migradores · CAP-141 · los primeros tests de `store` |
| `4949a9b` | **T10** | cuarentena · esquema futuro · `persistLocked` que devuelve el error y revierte |
| `a62208f` | **T11** | CV-D16, el re-key + `arnesia sesiones recalibrar-llaves` · CAP-142 |
| `d67200a` | **T12** | la ley se invierte: `Conv` y `Checkpoint` sobreviven · CAP-98 + `ledger/HS-29.md` |
| `edd17c4` | **T13/T14** | `AbrirRegistro` cableado + `Informe` en el log · el benchmark de H-7 |

**Gate del tramo 1: los 9 comandos verdes, y las cifras son PROPIAS** (tabla en `PARIDAD.md` §1.9)
— es la diferencia con la corrida anterior, que medía dos sesiones a la vez. ⚠ **CI sigue sin
observarse**: no se pushea, es decisión del operador.

### Cuatro cosas que cambian el punto de partida de T15

1. **`Session.Instantanea()` existe y hay que usarla.** Toda sesión que cruza el candado sale como
   instantánea. Si T15 agrega un método que devuelve una sesión, tiene que hacer lo mismo — devolver
   `*r.meta` reintroduce la carrera N-8 que `-race` ya cazó una vez.
2. **`persistLocked` DEVUELVE error.** El camino del operador (crear/retomar/renombrar) propaga y
   **revierte**; el camino del stream usa `persistOSeguir()`, que loguea y sigue. T15 está del lado
   del operador: su transición tiene que revertir el snapshot si el disco falla (`arquitectura.md`
   §4.2 paso 7).
3. **`usecase.ErrSoloLectura` existe y `SessionService.SoloLectura()` es el guard.** `Create` ya lo
   consulta antes de tocar memoria. Las 3 transiciones de T15 tienen que hacer lo mismo, y **T18 es
   el que mapea ese error a 503** — hoy no está mapeado, y está declarado.
4. **`Turnos` ya no existe en `Session`.** La cuenta es `Conversacion.NumTurnos()`. El DTO del wire
   la manda calculada.

### Lo siguiente, exacto

**Arrancá por T15** (`plan-desarrollo.md` línea 475) — la transición atómica.

- **Ticket:** T15 · crear · retomar · renombrar, con `transicionLocked`.
- **Archivo nuevo:** `internal/usecase/session_conversaciones.go`.
- **El orden exacto de los 8 pasos** está en `arquitectura.md` §4.2 y **no se improvisa**: guard →
  snapshot → operación pura del dominio → denegar los permisos pendientes de la que se desactiva →
  resetear el estado de vuelo (`runSeq` NO se toca) → mover `live` → persistir (con rollback del
  snapshot si falla) → armar frames. Y **fuera del lock**: `viejo.Close()` y después `publish`.
- **El precedente literal del paso 4** es `Interrupt` (`session_service.go`, los `permission_result`
  con `decision: deny`).
- **Capability:** el bloque de la arquitectura §7.3 nombra CAP-141/142/143 para estas hojas, pero
  **esos números YA ESTÁN USADOS** (141 = migración de esquema, 142 = recalibración de llaves). La
  próxima libre se verifica contra `docs/product/capabilities/` real —hoy **CAP-143**—, nunca contra
  la prosa de los documentos.
- **Cierra el 5.º check de `sesion-viva-consistente`** (`transicion-de-conversacion-atomica`), que
  hoy está declarado pendiente y por eso el boundary sigue en **4/5**.

### Cosas del entorno que ya no hace falta re-descubrir

- **El worktree resuelve N-6 de verdad.** El `pre-commit` rechazó **4 commits** por lint propio (12,
  1, 8 y 1 hallazgos) y todos se arreglaron: **cero `--no-verify`**. En el árbol compartido esos
  rechazos habrían sido de archivos ajenos y no habrían significado nada.
- **N-10 · el symlink de `node_modules` NO alcanza para el runner de navegador.** Sirve para `tsc`,
  `biome` y `vitest --project=unit`; con storybook, los 43 archivos fallan con «Failed to fetch
  dynamically imported module» porque los módulos resuelven fuera de la raíz que Vite sirve. La
  salida es `pnpm install --frozen-lockfile` real en el worktree: **1,2 s**, store compartido.
- **`vitest --project=storybook` corre headless en ~11 s** (43 archivos, 386 tests).
- **N-5 sigue vigente:** `golangci-lint` local (v2.12.2) es más ruidoso que el de CI sobre archivos
  preexistentes. Lo que el hook marca en TUS archivos sí es tuyo — y en este tramo lo fue las 4
  veces.
- **`~/.arnesia/` NO se tocó.** `sessions.json` conserva su md5 `b1689d15…` tras seis corridas del
  comando de re-key, todas contra copias en el scratchpad. Las 3 cerradas de CV-D6 siguen enteras:
  el borrado es T32 y es **procedimiento manual del operador**.

### Lo que sigue abierto del propio paquete

- **GATE 2** sigue siendo *autorizado por directiva, no por lectura*. Si el operador lee los specs y
  rechaza algo, lo construido sobre esa parte se rehace.
- **El gate de T6** (CI 3/3 verde) sigue **abierto** hasta que se pushee.
- **`archivo-durable-declara-su-esquema` va 5 de 6** y sigue `proposed`. El que falta es
  `durable-no-se-wipea` (`TestDurableNuncaSeWipea`), y su celda lo dice. Gradúa con los seis.
- **`sesion-viva-consistente` sigue `enforced` 4/5** — el quinto es T15.
- **N-1 vigente:** el dock viola contraste en producción hoy (`SessionLine`, 2,21:1). Se corrige en
  **T23**, no antes.
- **11 hallazgos** declarados en `PARIDAD.md` §3.4. Los que condicionan lo que sigue: **N-8** (la
  copia que salía del candado no era una copia) y **N-9** (leer un registro viejo tiraba lo que
  había cambiado de lugar) — los dos ya arreglados, pero los dos son el tipo de error que T15 puede
  reintroducir si devuelve el agregado vivo o si persiste sin mirar el error.
