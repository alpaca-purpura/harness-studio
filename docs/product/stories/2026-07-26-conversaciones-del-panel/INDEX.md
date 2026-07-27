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
| **implementar** | 🚧 **tramos 0-4 CERRADOS + T30/T31 del tramo 5** (**31 de 33**) · faltan **T32** (manual del operador) y **T33** (cierre) — ver [`PARIDAD.md`](./PARIDAD.md) §1.13, §1.14 y §2.1 |
| **verificación E2E contra el binario INSTALADO (T31)** | ✅ [`verificacion-e2e/INFORME.md`](./verificacion-e2e/INFORME.md) — los **6 guiones** contra `~/.local/bin/arnesia` sello **`0.2.24.2607262315`** · **98 aserciones: 92 ok · 3 fallas · 3 n/c** · 21 capturas · reproducible con `bash verificacion-e2e/correr.sh`. **`~/.arnesia` del operador INTACTO** (md5 `b1689d15…` antes y después) |
| PARIDAD | 🚧 [`PARIDAD.md`](./PARIDAD.md) — **§2 ya compara el dibujo contra el producto: 30 ✅ · 5 ⚠️ · 0 ❌**, con 26 capturas en `verificacion-tramo3/` (13 escenas × 2 temas). **Gate 🧑‍⚖️ SIN FIRMAR**: es del operador |

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
- **CV-D17** — la mudanza del picker (RF-333/RF-334) **se adelanta del tramo 5 al 3**: retirar los
  dos métodos del cliente deja a su único consumidor sin compilar, y conservarlo sólo garantizaba
  un error a la vista del operador. 🧑‍⚖️ **pendiente de lectura**.

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

**Nada de esto está en `main`.** Hay **tres** worktrees encadenados y ninguno se pusheó:

| tramo | worktree | rama | parte de | commits |
|---|---|---|---|---|
| 1 (T8-T14) | `.claude/worktrees/agent-a12f73cab8f2b13b1` | `worktree-agent-a12f73cab8f2b13b1` | `657054d` | 6 |
| 2 (T15-T20) | `.claude/worktrees/tramo2` | `tramo2-conversaciones` | `68515ec` | 4 |
| **3 + 4 (T21-T29)** | **`.claude/worktrees/tramo3`** | **`tramo3-conversaciones`** | `4c32a4d` (cierre del tramo 2) | **5** |

`tramo3-conversaciones` **ya contiene** los tramos 1 y 2: parte del commit de cierre del 2.
Integrar a `main` es un `merge --ff-only` mientras `main` no avance; si avanzó, rebase.

### Lo hecho — TRAMOS 0-3 CERRADOS + EL 4 CASI ENTERO (29 de 33 tickets)

| commit | ticket | qué |
|---|---|---|
| `bc9d1d4` … `657054d` | T1-T7 | tramo 0 + `domain.Conversacion` (ya en `main`) |
| `cd7571f` … `edd17c4` | T8-T14 | el modelo partido, el sobre versionado, la cuarentena, CV-D16, la ley invertida |
| `4bbe40e` … `e041aa7` | T15-T20 | la transición atómica, la rotación que emite, el buscador, las rutas y el contrato |
| `d36417e` | **T21** | los tipos del wire · el cliente · **RF-333/RF-334 adelantado (CV-D17)** |
| `f96453c` | **T22** | el store: `activa`, `convRev`, la rama del frame, el `default:` · 7 tests |
| `0073fec` | **T23/T24/T26/T27/T28** | las 4 piezas del panel + el transporte del widget · 44 stories · 10 tests |
| `55f5612` | **T25/T29** | el dock recompuesto · 17 stories · **la app vuelve a funcionar** |
| (este) | — | PARIDAD, CV-D17, capabilities, cifras, changelog |

**Gate de los tramos 3 y 4: 11 de 12 comandos verdes, cifras PROPIAS** (tabla en `PARIDAD.md`
§1.14). `conformance --todo` → `323 · pass 91 · fail 0 · **error 0** · deferred 232`.
⚠ **`go-arch-lint` NO se corrió**: el binario no está en el `PATH` de este entorno y no hay target
en el `Makefile`. ⚠ **CI sigue sin observarse**: no se pushea, es decisión del operador.

### 🟢 La app volvió a funcionar, y ahora tiene la superficie del dibujo

Lo que §1.12 declaraba roto está reparado: `tsc` pasó a **18 errores** con T21 —que era el objetivo
del ticket: volver al compilador el inventario— y volvió a **0** con T22-T25. El cromo tiene 2
filas, el chip de contexto abre la identidad con el `cwd`, la lista abre en sitio con su buscador,
y crear · retomar · renombrar están cableados contra el daemon.

**26 capturas** en `verificacion-tramo3/`, 13 escenas × 2 temas, revisadas a ojo.

### Lo siguiente, exacto

**T30 y T31 están CERRADOS** (tramo 5, worktree `.claude/worktrees/tramo5`, rama
`tramo5-conversaciones`, partida de `508b9b4`). Lo que queda:

1. **Leer** [`verificacion-e2e/INFORME.md`](./verificacion-e2e/INFORME.md) §3 («lo que el
   operador va a ver cuando instale») y §5 (los 6 hallazgos nuevos, N-18…N-23).
2. **Decidir N-19** — es el único que espera una decisión de contenido: el `＋` bloqueado dice
   «esperá a que termine el turno **en vuelo**» y el spec manda «…el turno **(■ para
   interrumpir)**». No se tocó porque el gate 2 sigue *autorizado por directiva, no por lectura*.
3. **T32 · el borrado de CV-D6 (E-46)** — **procedimiento manual del operador, con el daemon
   detenido y copia previa.** Los 3 ids (`s1b38a066`, `s408bb085`, `s020210e3`) siguen
   **intactos** en `~/.arnesia/sesiones-cerradas.json`. NO se ejecutó acá, a propósito.
4. **N-21 al BACKLOG** (pérdida silenciosa de la marca de rotación con dos vistas) — es el
   hallazgo grave y su arreglo NO es chico: toca el guard de idempotencia por `turno_idx`.
5. **T33 · cierre** + gate 🧑‍⚖️ de PARIDAD.

### ⚠ Antes de tocar nada en esta máquina

- **El binario instalado del operador FUE REEMPLAZADO** por `make dev-sync` (sello
  `0.2.24.2607262315`). Respaldo del estado previo:
  `~/.local/bin/arnesia.respaldo-20260726-231513` (md5 `090bb5d6…`). Para volver:
  `pkill -f "$HOME/.local/bin/[a]rnesia serve" && cp -p ~/.local/bin/arnesia.respaldo-20260726-231513 ~/.local/bin/arnesia`
- **`~/.arnesia/` NO se tocó.** md5 `b1689d1513e8d3c3fa21692c511ed793` verificado al abrir y al
  cerrar el tramo. Todo el E2E corrió con `HOME` apuntando a un sandbox.
- **`make installer` NO se corrió** (depende de `bump-patch`) y **no se pusheó nada**.

### Lo que sigue abierto del propio paquete

- **GATE 2** sigue siendo *autorizado por directiva, no por lectura*. **CV-D17 tampoco fue leída.**
- **El gate de T6** (CI 3/3 verde) sigue **abierto** hasta que se pushee.
- **`archivo-durable-declara-su-esquema` va 5 de 6** y sigue `proposed`. Falta `durable-no-se-wipea`.
- **N-1 CERRADO** en T23: el `◍ <cc-id>` pasó a `--foreground` al mudarse al detalle. Las 3 stories
  del baseline **dejaron de apagar `color-contrast`**.
- **C-13 sigue abierta**: el anillo de foco mide 2,51:1 en claro. Se conservó el precedente del
  repo, como manda su propio veredicto; la deuda del token vive en `BACKLOG.md`.
- **23 hallazgos** en `PARIDAD.md` §3.4. Los del tramo 5 (todos del E2E contra el binario
  instalado): **N-18** (el dry-run de CV-D16 miente si corre antes del daemon), **N-19** (el `＋`
  bloqueado no ofrece `(■ para interrumpir)`), **N-20** (el 409 no distingue `await` de
  `streaming`), **N-21** 🔴 (**la marca de rotación se pierde en silencio con dos vistas** — el
  grave), **N-22** (RF-348 CA-1 sin construir) y **N-23** (el re-key no corre solo ni avisa).
- **Lo que el E2E CERRÓ:** la rotación **llega en vivo por SSE** (5 mutaciones del DOM medidas,
  sin recargar) — **C-6/H-8 dejan de estar abiertos**; y C-3 se confirmó en el producto (barra
  caliente, número no).
