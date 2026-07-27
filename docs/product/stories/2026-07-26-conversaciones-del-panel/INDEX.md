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
| **implementar** | 🚧 **tramo 0 de 6 CERRADO y verde** (T1-T6) · tramos 1-5 sin empezar — ver [`PARIDAD.md`](./PARIDAD.md) §0 |
| PARIDAD | 🚧 [`PARIDAD.md`](./PARIDAD.md) abierta con la evidencia del tramo 0; **gate 🧑‍⚖️ SIN FIRMAR** (no hay superficie nueva que comparar) |

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

### Lo hecho (2026-07-26) — TRAMO 0 CERRADO Y VERDE

3 commits locales en `main`, **sin pushear** (los 79 previos tampoco: es decisión del operador):

| commit | ticket | qué |
|---|---|---|
| `bc9d1d4` | — | los artefactos de spec/diseño/arquitectura/planes + los 2 boundaries nuevos, `proposed` con `enforced_by: []` |
| `56fdda1` | T2 · T3 | baseline del dock (3 stories) + contraste del picker ⇒ `--project=storybook` **386/386** |
| `eda01f0` | T4 · T5 | `_sin-declarar.yaml` (15 entradas con razón) + los 4 enforcers ⇒ `ruta-servida-esta-declarada` **`enforced` 4/5** |

**Gate del tramo 0: local completo y verde** (los 9 comandos de los 3 jobs de CI, tabla en
`PARIDAD.md` §1 T6). ⚠ **CI NO SE OBSERVÓ**: T6 exige `git push` + `gh run watch`, y no se pusheó.
El rojo conocido (job `ts`, paso `fitness visual`) tiene su causa corregida y medida, pero eso es
una **inferencia, no una corrida vista**.

### Lo siguiente, exacto

**Arrancá por T7** (`plan-desarrollo.md` línea 193). Es el primer ticket del tramo 1 y es
**aditivo y sin riesgo**: crea archivos que todavía no importa nadie.

- **Ticket:** T7 · `domain.Conversacion` + las 4 operaciones puras + la invariante.
- **Archivos a crear:** `internal/domain/conversacion.go` · `internal/domain/conversacion_test.go`.
- **Diseño literal:** `arquitectura.md` §1.2 (los 14 campos, con `Conv []Turn` **sin `omitempty`** y
  `Turnos` que **no existe**: se deriva con `NumTurnos()`) y §1.4 (las firmas de `Activa`,
  `CrearConversacion`, `ActivarConversacion`, `RenombrarConversacion`, `VerificarUnaActiva`,
  `NormalizarConversaciones`).
- **Capability:** **nueva CAP-140** `dominio-l0/conversacion-como-entidad.yaml`. **Verificado
  2026-07-26: CAP-139 es el máximo real en `docs/product/capabilities/`, así que CAP-140 sigue
  libre** (⚠ un `grep` sobre `docs/` devuelve CAP-146 porque los documentos de ESTE paquete ya
  nombran el bloque 140-146 que todavía no existe — no te confundas).
- **Los 6 tests, por nombre:** `TestInvarianteUnaActivaTrasCadaTransicion` ·
  `TestCrearDesactivaLaAnterior` · `TestActivarConversacionAjenaEs404` ·
  `TestRenombrarVacioNoCambiaElTitulo` · `TestNormalizarReparaYLoDice` ·
  `TestTituloEditadoNoSeReDeriva`.
- **Las 2 trampas mecánicas, leelas antes de escribir** (`arquitectura.md` §7.6): (1) ningún
  comentario con `transcript`/`jsonl`/`.claude/projects` **adyacente** a un decoder, o
  `TestNoJSONLSchemaParsing` da falso positivo — y **está prohibido** sumar el prefijo a
  `jsonlExento`; (2) `domain.Turn` **NO se toca** (es el incidente HS-26,
  `conductor_test.go:TestUserTurnWireSinCamposExtra`).

**Después T8, y es el ticket más riesgoso del paquete** (partir `Session`: el árbol deja de
compilar hasta arreglar todos los call-sites). No lo arranques sin margen para terminarlo: un T8 a
medias deja `main` sin compilar. Los 5 tests del boundary
(`TestOneTurnAtATime`/`TestFramesCarryRunID`/`TestResumeAutoSana`/`TestNoSilentEventDrop`/`TestSessionSpawnsInArnesPath`)
tienen que seguir verdes **sin tocarles el cuerpo**: es el canario de que el reparto no cambió el pipe.

### Cosas del entorno que ya no hace falta re-descubrir

- **`lefthook` YA está instalado** en este working copy (T1). Los 8 jobs `pre-commit` disparan de
  verdad — cazaron 5 hallazgos propios en el commit de T5. `estado-cifras` tarda ~55 s por commit y
  regenera + `git add` del checkpoint solo.
- **`golangci-lint` local (v2.12.2) es MÁS RUIDOSO que el de CI**: ~25 hallazgos preexistentes en
  telemetría/logfile/selfupdate que el job `go` de CI **acepta** (última corrida: `go` ✓). No los
  persigas; sí atendé lo que el hook marque en TUS archivos.
- **El browser runner de vitest corre headless y rápido** (~4 s por archivo, 43 archivos en ~12 s).
- **`~/.arnesia/` NO se tocó**: las 3 cerradas de CV-D6 (`s1b38a066`, `s408bb085`, `s020210e3`)
  siguen enteras. El borrado es T32 y es **procedimiento manual del operador**.

### Lo que sigue abierto del propio paquete

- **GATE 2** sigue siendo *autorizado por directiva, no por lectura* (ver arriba). Si el operador
  lee los specs y rechaza algo, lo construido sobre esa parte se rehace.
- **El gate de T6** (CI 3/3 verde) queda **abierto** hasta que se pushee.
- 5 hallazgos nuevos que ningún documento preveía, todos declarados en `PARIDAD.md` §3.4 — el más
  importante es **N-1**: el dock tiene hoy, en producción, una violación de contraste (2,21:1) que
  nadie había visto porque el widget no tenía story.
