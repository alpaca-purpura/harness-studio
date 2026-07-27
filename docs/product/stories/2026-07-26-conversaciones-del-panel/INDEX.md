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
| **implementar** | 🚧 **tramos 0, 1 y 2 CERRADOS y verdes** (T1-T20, 20 de 33) · tramos 3-5 sin empezar — ver [`PARIDAD.md`](./PARIDAD.md) §0, §1.8, §1.10 y §1.11 |
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

### Dónde vive el trabajo — LEELO PRIMERO

**Nada de esto está en `main`.** Hay **dos** worktrees encadenados y ninguno se pusheó:

| tramo | worktree | rama | parte de | commits |
|---|---|---|---|---|
| 1 (T8-T14) | `.claude/worktrees/agent-a12f73cab8f2b13b1` | `worktree-agent-a12f73cab8f2b13b1` | `657054d` | 6 |
| **2 (T15-T20)** | **`.claude/worktrees/tramo2`** | **`tramo2-conversaciones`** | `68515ec` (cierre del tramo 1) | **4** |

`tramo2-conversaciones` **ya contiene** todo el tramo 1: parte de su commit de cierre. Integrar a
`main` es un `merge --ff-only` de `tramo2-conversaciones` mientras `main` no avance; si avanzó,
rebase.

### Lo hecho — TRAMOS 0, 1 y 2 CERRADOS Y VERDES (20 de 33 tickets)

| commit | ticket | qué |
|---|---|---|
| `bc9d1d4` … `657054d` | T1-T7 | tramo 0 completo + `domain.Conversacion` (ya en `main`) |
| `cd7571f` … `edd17c4` | T8-T14 | el modelo partido, el sobre versionado, la cuarentena, CV-D16, la ley invertida, el cableado y el benchmark |
| `4bbe40e` | **T20/T15** | la sesión nace con su hilo · la transición atómica · el 5.º check con enforcer real · CAP-143 |
| `a08451c` | **T16** | la rotación emite su frame · CAP-144 |
| `b8c5c2a` | **T17** | el buscador entra al texto, fixture real de 90 turnos · CAP-145 |
| `e041aa7` | **T18/T19** | las 4 rutas + el contrato que las declara · el parser del enforcer reparado |

**Gate del tramo 2: los 11 comandos verdes y las cifras son PROPIAS** (tabla en `PARIDAD.md` §1.11).
`conformance --todo` → `323 · pass 91 · fail 0 · **error 0** · deferred 232`. ⚠ **CI sigue sin
observarse**: no se pushea, es decisión del operador.

### 🔴 Lo primero que tenés que saber: el FE está roto a propósito

Con estos 4 commits **la app no funcionaría**, y no es un descuido — es la forma del plan (backend
en el tramo 2, FE en el 3). `PARIDAD.md` §1.12 lo detalla. En corto:

- `GET /api/sessions` ya **no** trae `claude_session_id`, `model`, `ctx_pct`, `conv`, `turnos` ni
  `cadena_cc` en la raíz: bajaron a `activa`. `sessions-store.ts` los lee de la raíz ⇒ el dock
  pintaría **transcript vacío y ctx 0**.
- `client.ts` conserva `conversacionesDeArnes` (ahora **400**) e `historialCerrada` (ahora **404**).
- **`tsc` pasa igual**, y eso es el problema: `types.ts` no cambió, así que el compilador no ve nada.
  T21 existe justamente para que el compilador se vuelva el inventario de call-sites rotos.

Nada llegó al operador: no se pusheó, no se corrió `make dev-sync`, el binario instalado es el de
antes.

### Lo siguiente, exacto

**Arrancá por T21** (`plan-desarrollo.md` línea 673) — los tipos del wire y los 5 métodos del
cliente. Es el ticket que destapa la cascada; los cuatro que le siguen (T22-T25) la cierran.

- El backend YA responde con la forma de `arquitectura.md` §6.1. Comprobalo antes de escribir el
  tipo: `curl` contra el daemon, o leé `sessionWire` en
  `internal/adapters/transport/http/sessions.go` — es la proyección real, campo por campo.
- **`Session.activa` NO es opcional** y `conv` va **sin `?`**: el tipo tiene que hacer imposible el
  `if (!activa)` defensivo. El backend lo garantiza (el `activa` del wire siempre está presente, y
  si faltara el daemon loguea `error` en vez de inventar una).
- `client.ts` **pierde** `conversacionesDeArnes` y `historialCerrada` (sus rutas ya no existen) y
  **gana** `conversaciones(id, q?)` **con `signal`** (teclear rápido pisa peticiones),
  `crearConversacion`, `activarConversacion` y `renombrarConversacion` **sin `signal`**.
- El frame nuevo: `kind: "conversacion"` con `conversacion_id`, `conversacion_evento`
  (`creada|activada|renombrada|rotada`), `turno_idx?` (sólo en `rotada`) y `conversacion?` (el
  estado post-transición, para repintar sin una segunda vuelta). **Sin `run_id`**, a propósito.
- ⚠ `cadena_cc` sigue **sin consumidor**. Si T23 no lo pinta, **se borra**.

### Cuatro cosas del tramo 2 que cambian el punto de partida

1. **La respuesta de `POST …/conversaciones` y de `…/activar` trae `desactivada` SIEMPRE**, con
   `null` cuando no había ninguna. No lo deduzcas de la lista anterior.
2. **`total` de la lista es el total de la SESIÓN**, no el de coincidencias: es el denominador del
   «N de M coinciden». Hay un test que prohíbe confundirlos.
3. **El fragmento llega en texto plano.** El `<mark>` es del FE (T27). El daemon no manda marcado.
4. **`?cerradas=` responde 400 con puntero**, no 200 vacío. Si algo del FE lo manda, lo vas a ver.

### Cosas del entorno que ya no hace falta re-descubrir

- **El worktree resuelve N-6.** El `pre-commit` corrió limpio en los 4 commits del tramo 2 (`go-lint`
  0 issues, `capabilities`, `capabilities-index`, `estado-cifras` regenerando solo): **cero
  `--no-verify`**.
- **`pnpm install --frozen-lockfile` real en el worktree, 1,2 s.** El symlink de `node_modules` NO
  sirve para `vitest --project=storybook` (N-10): 43 rojos por «Failed to fetch dynamically imported
  module». Con install real: **386/386 en 12,3 s**, headless.
- **`estado-cifras` del `pre-commit` cuesta ~2 min** por commit (3× `go run`). Es el precio de que
  las cifras se regeneren solas; no lo saltees.
- **N-5 sigue vigente:** `golangci-lint` local (v2.12.2) es más ruidoso que el de CI sobre archivos
  preexistentes. `--new-from-rev=HEAD` sobre lo tuyo: **0 issues** en los 4 commits.
- **`~/.arnesia/` NO se tocó.** `sessions.json` conserva su md5 `b1689d15…`, verificado al abrir y
  al cerrar el tramo. La única lectura fue la extracción del fixture de T17.

### Lo que sigue abierto del propio paquete

- **GATE 2** sigue siendo *autorizado por directiva, no por lectura*.
- **El gate de T6** (CI 3/3 verde) sigue **abierto** hasta que se pushee.
- **`archivo-durable-declara-su-esquema` va 5 de 6** y sigue `proposed`. Falta
  `durable-no-se-wipea` (`TestDurableNuncaSeWipea`). Gradúa con los seis.
- **`sesion-viva-consistente` pasa a `enforced` 5/5** — cerrado en T15.
- **N-1 vigente:** el dock viola contraste en producción hoy (`SessionLine`, 2,21:1). Se corrige en
  **T23**, no antes.
- **14 hallazgos** declarados en `PARIDAD.md` §3.4. Los del tramo 2: **N-12** (el enforcer del
  contrato tenía código muerto), **N-13** (los `publish` de `consume` están fuera del guard —
  preexistente, declarado, no tapado) y **N-14** (`sessionWire` no estaba en ningún ticket).
- **El E2E contra el binario instalado es T31 (tramo 5).** No lo adelantes: `make dev-sync`
  reemplaza el binario del operador y le mata el daemon.
