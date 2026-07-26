# Plan de pruebas · Las conversaciones viven en el panel de conversación

> `tipo: plan-pruebas` · paquete `2026-07-26-conversaciones-del-panel` · 2026-07-26.
> Hermano de [`plan-storybook.md`](./plan-storybook.md) (**qué asserta cada story**, 56) y de
> [`plan-desarrollo.md`](./plan-desarrollo.md) (**qué ticket trae cada test**).
> Este documento define **la pirámide**, el **circuito E2E contra la app INSTALADA**, los **datos de
> prueba** y el **criterio de salida** de PARIDAD.
>
> **Contrato de honestidad:** ningún check se declara verde sin correrse. Cifras generadas. Lo que no
> se puede automatizar se cubre con rastro escrito y **se declara**, jamás con un pass.

---

## 1 · La pirámide

| Nivel | Qué se prueba acá **y no en otro lado** | Dónde | Cuántos | Comando |
|---|---|---|---|---|
| **Go · dominio** | **La invariante** (una activa), las 4 transiciones puras, el título, la normalización. Sin un solo fake — si un test necesita `fakeAgent`, la invariante quedó en el lugar equivocado | `internal/domain/conversacion_test.go` | 6 | `go test ./internal/domain/ -run TestInvariante -v` |
| **Go · persistencia** | Envelope, detección, cadena de migradores, idempotencia, cuarentena, esquema futuro, re-key. **`internal/adapters/store` tiene 0 tests hoy** — este paquete le da los primeros | `internal/adapters/store/*_test.go` | 17 | `go test ./internal/adapters/store/ -v` |
| **Go · usecase** | La transición atómica bajo lock, el 409, la denegación de permisos pendientes, el rollback, la rotación que emite, el matcher de búsqueda | `internal/usecase/session_{conversaciones,busqueda,rotacion,historial}_test.go` | 18 | `go test ./internal/usecase/ -race -v` |
| **Go · transporte** | Los códigos y la forma de las 5 rutas. **Cierra el gap más grande del lado Go**: `sessions_test.go` cubre hoy **1 de 10** endpoints | `internal/adapters/transport/http/sessions_conversaciones_test.go` | 6 | `go test ./internal/adapters/transport/http/ -v` |
| **Go · fitness** | Los boundaries: la transición atómica + los 4 del contrato OpenAPI + los 5 de capabilities | `docs/architecture/fitness/` | +5 | `go test ./docs/architecture/fitness/ -v` |
| **FE · unit (Node, sin Chromium)** | El **transporte** del panel y el **ruteo de frames**: refetch, descarte, errores tal cual, 409, timeout, idempotencia por `turno_idx` | `web/src/widgets/chat-dock/model/conversaciones-store.test.ts` · `web/src/shared/store/sessions-store.test.ts` (**el store de 478 líneas no tiene test hoy**) | 16 | `pnpm --dir web exec vitest --project=unit run` |
| **Story `play()` (Chromium headless)** | **La superficie**: estados visuales, roles ARIA, teclado, foco, y el gate **axe en `error`** en light y dark | 5 archivos `*.stories.tsx` del `chat-dock` + el superset del picker | 56 | `pnpm --dir web exec vitest --project=storybook run` |
| **E2E vivo · app instalada** | Lo que **ningún fake ve**: la migración real del disco del operador, el `--resume` real, la rotación real llegando por SSE, dos vistas, el `cwd` que se mueve | `e2e/` del paquete | 6 | §3 |

**Por qué esta repartición y no otra**

- **La invariante va en `domain`, no en un test de UI.** Es la regla que no se puede violar; probarla
  en una story la ataría al render.
- **El matcher de búsqueda va en Go, no en el FE.** RF-341 lo pone en el servidor (`normalizar()`
  existe **una sola vez**): probarlo en el FE probaría una reimplementación que no existe.
- **La accesibilidad va en `play()`, no en E2E.** El gate axe global es `error`
  (`web/.storybook/preview.ts:7`) y corre en los dos temas; un E2E lo mediría una sola vez y en un
  solo tema.
- **La migración va en Go *y* en E2E.** El test Go usa el fixture real de 17 202 B; el E2E usa el
  **archivo del operador copiado a un sandbox**. El primero prueba el algoritmo, el segundo prueba
  que el binario instalado lo corre. Son cosas distintas y las dos hacen falta.
- 🕳 **No hay regresión visual por píxel.** `fe-visual-fitness.md` menciona `storybook-addon-vis`,
  pero **no está instalado** (0 ocurrencias). «Story = test» significa hoy **render + `play` + axe**.
  Todos los asserts del plan son **estructurales o de rol**; el único de estilo computado es
  `ctx-chip.stories.tsx:C-02` (el número caliente **no** en `--warn`), y está justificado porque C-3
  es la contradicción más cara del paquete.

---

## 2 · Lo que ya existe y no se toca

Estos siguen midiendo lo suyo y **que sigan verdes es el canario** de que partir `Session` no cambió
el pipe conductor↔dock:

| Test | Archivo:línea | Por qué importa acá |
|---|---|---|
| `TestOneTurnAtATime` | `arch_test.go:1911` | El guard de un-turno-a-la-vez es el mismo que protege la transición |
| `TestFramesCarryRunID` | `arch_test.go:1924` | `runSeq` se queda por sesión (§4.4 de `arquitectura.md`) |
| `TestResumeAutoSana` | `arch_test.go:1957` | El heal ahora limpia el `ClaudeSessionID` **de la conversación** |
| `TestNoSilentEventDrop` | `arch_test.go:1699` | — |
| `TestSessionSpawnsInArnesPath` | `arch_test.go:1891` | — |
| `TestRotacionInvisible` · `TestCtxHistYUmbralRotacion` | `session_rotacion_test.go:67`, `:16` | `rotarLocked` cambia de **firma**, no de lógica |
| `TestUserTurnWireSinCamposExtra` | `conductor_test.go:54` | La regresión de HS-26. **`domain.Turn` no se toca** justamente por esto |
| `TestHistorialCerradaCoseCadena` | `session_historial_test.go:82` | El lector JSONL sobrevive como fallback pre-migración |

⚠ **Los 5 primeros hacen `svc.List()[0].ID`** y dependen de que `newTestService` deje exactamente una
sesión sembrada en la posición 0. T20 ajusta **el helper**, no los tests: si alguno necesita que le
cambien una aserción, el reparto cambió el comportamiento y eso es un hallazgo, no un ajuste.

---

## 3 · E2E contra la app INSTALADA

**El requisito del operador:** cuando él instale, tiene que ver exactamente lo construido. Por eso el
E2E no corre contra `pnpm dev` ni contra un mock del daemon.

### 3.1 El circuito, paso a paso

```
┌─ PREPARAR ────────────────────────────────────────────────────────────────────┐
│ 0. CHANGELOG primero. `make installer` depende de `bump-patch`, que FALLA si   │
│    [Sin publicar] está vacía (versionado.md §changelog-y-bump). Si cada ticket │
│    anotó su línea en el mismo turno (METODOLOGIA §10 regla 1), esto ya está.   │
│    python3 scripts/changelog.py add Agregado "…"                              │
│                                                                                │
│ 1. Sandbox de datos (NO se toca el ~/.arnesia real — §4):                      │
│    export ARNESIA_E2E=/tmp/arnesia-e2e-$(date +%s)                             │
│    mkdir -p "$ARNESIA_E2E/.arnesia" "$ARNESIA_E2E/.claude/projects"            │
│                                                                                │
│ 2. Semilla: la copia del registro REAL en formato v1 (para probar la migración)│
│    cp ~/.arnesia/sessions.json "$ARNESIA_E2E/.arnesia/sessions.json"           │
│    ⚠ `cp`, jamás `mv`. Y NUNCA se copia `sesiones-cerradas.json`: CV-D6 lo     │
│      borra y el E2E no debe resucitarlo.                                       │
└────────────────────────────────────────────────────────────────────────────────┘
┌─ CONSTRUIR E INSTALAR ────────────────────────────────────────────────────────┐
│ 3. make dev-sync                                                              │
│    → bundle.sh --daemon-only: paso 1 SPA (vite build → web/dist) + paso 2      │
│      go build con la SPA EMBEBIDA y el sello -ldflags (bundle.sh:21-48)        │
│    → install -m755 bin/arnesia ~/.local/bin/arnesia.new && mv -f (atómico)     │
│    → pkill del daemon vivo (truco del corchete `[a]rnesia`, Makefile:96-101)   │
│    ⚠ `make dev-sync` sale con exit 0 SILENCIOSO si ~/.local/bin/arnesia no     │
│      existe (Makefile:88-91). Verificar que el archivo existe ANTES.           │
│                                                                                │
│ 4. SOLO para el Modo B (ventana Tauri): make installer                        │
│    → bump-patch + bundle.sh completo → instaladores/vX.Y.Z/ + checksums        │
│    → e instalar el .deb. ⚠ Y CORRER `make dev-sync` DESPUÉS: el shell instalado│
│      PREFIERE SIEMPRE ~/.local/bin/arnesia sobre el sidecar (Makefile:27-32).  │
│      Sin eso se mide el daemon viejo con el formato viejo y se concluye mal.   │
└────────────────────────────────────────────────────────────────────────────────┘
┌─ LEVANTAR ────────────────────────────────────────────────────────────────────┐
│ 5. pkill -f '[a]rnesia serve'   # el daemon del operador, si estaba            │
│ 6. HOME="$ARNESIA_E2E" ~/.local/bin/arnesia serve --addr 127.0.0.1:4200 &      │
│    ⚠ :4200 NO es negociable: el SPA lo tiene HARDCODEADO                       │
│      (client.ts:21, `import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"`)│
│      y el .env no llega al bundle. BACKLOG.md:77-82 lo declara como bug.       │
│    ⚠ HOME sandbox: el daemon resuelve ~/.arnesia con os.UserHomeDir()          │
│      (registry.go:34) y ~/.claude/projects con history.New("") (reader.go:31). │
│ 7. Esperar: until curl -sf http://127.0.0.1:4200/api/version; do sleep 0.2; done│
│ 8. ASERCIÓN DE IDENTIDAD (obligatoria, primera):                              │
│    curl -s /api/version | jq -r .version  ==  X.Y.Z.<sello de hace un minuto>  │
│    Si el sello es viejo, se está midiendo OTRO binario. Se aborta.            │
└────────────────────────────────────────────────────────────────────────────────┘
┌─ EJERCITAR ───────────────────────────────────────────────────────────────────┐
│ 9. node e2e/conversaciones.mjs      # Playwright headless contra :4200         │
│ 10. Modo B (gate humano): abrir la ventana Tauri instalada y repetir E2E-1..3  │
└────────────────────────────────────────────────────────────────────────────────┘
┌─ LIMPIAR ─────────────────────────────────────────────────────────────────────┐
│ 11. pkill -f '[a]rnesia serve' ; rm -rf "$ARNESIA_E2E"                         │
│ 12. Relanzar el daemon del operador con SU HOME.                              │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Modo A (automatable) y Modo B (gate humano)

| | Modo A | Modo B |
|---|---|---|
| **Qué se ejercita** | El **binario instalado** sirviendo la **SPA construida** (`bundle.sh` paso 1 la embebe en el daemon, `router.go:57` la sirve en `/`) | La **ventana Tauri** del `.deb` instalado |
| **Datos** | `HOME` sandbox — el `~/.arnesia` real **no se toca** | `HOME` real ⇒ **se corre último, después de T32**, y sólo se mira |
| **Driver** | `playwright` **crudo** de `web/node_modules` (v1.61.1). **`@playwright/test` NO está instalado** y no se usa | ojos del operador |
| **Depende de MCP** | **NO.** Los MCP de Chrome pueden estar tomados por otra sesión: el circuito no los toca | no |
| **Cuándo** | T31, y en cada re-corrida | una vez, en el gate de PARIDAD |
| **Precedente** | `stories/2026-07-08-chat-cc-funcional/e2e/{casuistica,real}.mjs` — se calca la forma y **se corrige** su import por ruta absoluta hardcodeada (`casuistica.mjs:6`) por `fileURLToPath` + `resolve` | — |

```js
// e2e/conversaciones.mjs — cabecera
import { fileURLToPath } from "node:url"
import { dirname, resolve } from "node:path"
const aqui = dirname(fileURLToPath(import.meta.url))
const { chromium } = await import(
  resolve(aqui, "../../../../../web/node_modules/playwright/index.mjs")
)   // ← relativo al archivo, NO la ruta absoluta del precedente
const browser = await chromium.launch({ headless: true })
```

`mock-claude.sh` se reusa del paquete de chat (`stories/2026-07-08-chat-cc-funcional/e2e/`) para los
escenarios que no necesitan un modelo real; **E2E-3 y E2E-6 corren con `claude` de verdad** porque
lo que prueban es justamente el `--resume` y la rotación.

### 3.3 Los 6 guiones, paso a paso, con su aserción

---

#### **E2E-0 · La migración del disco del operador** (E-43, E-01, CV-D16) — *el más importante*

Es el primer minuto del operador con la versión nueva. Si esto falla, nada más importa.

| # | Paso | Aserción |
|---|---|---|
| 1 | Antes de arrancar: `HOME=$ARNESIA_E2E ~/.local/bin/arnesia sesiones recalibrar-llaves` (dry-run) | Imprime **5 filas**, una por sesión, cada una con su motivo. **Se transcribe a `PARIDAD.md`** |
| 2 | Arrancar el daemon y leer el log | Dice `migró desde v1`, la **ruta del respaldo**, y una línea por `Recalibracion` |
| 3 | `ls $ARNESIA_E2E/.arnesia/` | Existen `sesiones.json` **y** `sessions.json` **y** `sessions.json.bak-<sello>` |
| 4 | `jq -r .schema_version sesiones.json` | `2` |
| 5 | `diff <(jq -S . sessions.json) <(jq -S . <copia previa>)` | **vacío** — el v1 **no se tocó** |
| 6 | `curl /api/sessions \| jq length` | `5` |
| 7 | `curl /api/sessions \| jq '[.[].activa] \| length'` | `5` — **una activa por sesión** |
| 8 | `curl /api/sessions \| jq '.[] \| select(.id=="s6165ac75") \| .activa.conv \| length'` | `90` — el transcript sobrevivió |
| 9 | `... \| .activa.ultima_interaccion` | `null`/ausente — **no se inventó** (H-4/H-B) |
| 10 | `... \| select(.id \| test("s0fec7798")) \| .activa.conv` | `[]` y **no** `null` (`no-aplica-no-es-cero`) |
| 11 | `curl /api/sessions \| jq -r '.[].arnes'` | Las que el dry-run marcó `resuelta-*` salen **calificadas**; las `sin-candidata`/`ambigua` salen **igual que antes** — **ninguna perdida, ninguna fusionada** |
| 12 | Reiniciar el daemon y repetir 4-11 | **Idéntico** — la migración no volvió a correr (idempotencia) |

---

#### **E2E-1 · Crear, listar, buscar y retomar** (E-02, E-03, E-06, E-10, E-22, E-24)

| # | Paso | Aserción |
|---|---|---|
| 1 | Navegar a `http://127.0.0.1:4200`, esperar el dock | **Exactamente 2 filas de cromo** entre el borde del dock y el primer mensaje: `queryByText("Alcance:") === null` ∧ ningún `◍` visible ∧ el botón de ctx presente ∧ `» colapsar` (**no** `⟩`) |
| 2 | Mandar un turno con `mock-claude` | El transcript crece; el título de la conversación **se deriva** del texto |
| 3 | Clic en `＋` | Nace `nueva conversación`, `0 turnos`, **chip a `0%` visible** (no escondido); el vacío **nombra** la desactivada |
| 4 | Clic en `▶` | `getByRole("listbox")` con **2** `option`; **el composer sigue visible**; **cero** `<dialog>` y cero backdrop en el DOM |
| 5 | Escribir `manifiesto` en el buscador | El rótulo pasa a `N de 2 coinciden` **sin apretar Enter**; las filas que coinciden traen `<mark>` |
| 6 | Escribir `telemetría` (sin coincidencias) | «Ninguna conversación de esta sesión menciona «telemetría»» + «Se buscó en el título y en el texto de **las 2**» + `Limpiar búsqueda`. **Nunca «0 conversaciones»** |
| 7 | Limpiar y clic en la fila inactiva | La franja `Retomando la conversación… --resume <8 chars>` aparece; el transcript **se repinta con los turnos viejos**; **el detalle de identidad se abre solo** |
| 8 | `curl /api/sessions/<id>/conversaciones` | Exactamente **una** con `"activa": true` |

---

#### **E2E-2 · Turno en vuelo: el 409 del servidor** (E-07, E-08, E-11, E-28, E-39)

| # | Paso | Aserción |
|---|---|---|
| 1 | Mandar un turno largo (`mock-claude` con delay) | `status` = `streaming` |
| 2 | Mirar el `＋` | `disabled` **con `title`** «esperá a que termine el turno (■ para interrumpir)» — nunca apagado y mudo |
| 3 | Abrir la lista | Las filas inactivas con `aria-disabled="true"` y su `title`; **el buscador sigue habilitado** y filtra |
| 4 | Saltear la UI: `curl -X POST .../conversaciones` | **409** con motivo legible, no un 500 ni un genérico |
| 5 | Tras el 409, `curl .../conversaciones` | **Nada cambió**: la misma activa, la misma cantidad |
| 6 | Con una tarjeta de permiso abierta (`await`), repetir 2-5 | Idéntico, con el motivo «esperá tu decisión de permiso» |

---

#### **E2E-3 · La rotación llega EN VIVO** (E-18, E-19 · cierra C-6/H-8) — *con `claude` real*

Es el escenario que hoy **no es realizable** (`session_rotacion.go` no tiene un solo `s.publish`).

| # | Paso | Aserción |
|---|---|---|
| 1 | Arrancar con `-rotacion-umbral 5` | — |
| 2 | Mandar turnos hasta que `rotacion_pendiente` sea `true` (`curl` lo confirma) | El chip de ctx se pinta **caliente**: la **barra** en `--warn`, el **número NO** (C-3) |
| 3 | Mandar el turno siguiente, **con el dock abierto** | La marca `— contexto rotado, seguimos —` (el literal de `session_rotacion.go:12`, **C-5: gana el código**) aparece **SIN RECARGAR**, centrada, punteada, sin cola de burbuja |
| 4 | Mirar la lista | Sigue mostrando **N**, no N+1; el título de la activa **no cambió**; su `ctx` bajó |
| 5 | Mirar el detalle | Se abrió solo, con el `cc-id` **NUEVO** |
| 6 | Repetir con el dock **colapsado** (`⌘K`), reabrir | La marca está **en su lugar cronológico**; el dock **no se abrió solo**; **cero toasts** |
| 7 | `curl .../conversaciones` | **Cero entradas nuevas**; la `cadena_cc` de la activa creció en 1 |

---

#### **E2E-4 · Dos vistas sobre la misma sesión** (E-40, E-41)

| # | Paso | Aserción |
|---|---|---|
| 1 | Abrir **dos** contextos de Playwright sobre `:4200` | Las dos muestran la misma sesión |
| 2 | Crear una conversación en la vista **A** | La lista de la vista **B** se actualiza **sola** (el frame `conversacion`), sin recargar |
| 3 | Renombrar en **B** | El título cambia en **A** |
| 4 | Cortar la red de **B** 3 s y reconectar (SSE replay por `Last-Event-ID`) | El transcript de **B** **no duplica** el breadcrumb de rotación (idempotencia por `turno_idx`) |
| 5 | Mandar un turno desde **A** y otro desde **B** a la vez | Uno responde **409**; el transcript **no se intercala** |

---

#### **E2E-5 · Los fallos reales del entorno** (E-14, E-49)

| # | Paso | Aserción |
|---|---|---|
| 1 | Con una sesión abierta, `mv` de la carpeta del arnés | — |
| 2 | Abrir la lista y buscar | **Funcionan**: son datos nuestros, dependen de `session_id` (BR-CV-2) |
| 3 | Mandar un turno | Falla con el motivo del resolver **nombrando la ruta**; la conversación vuelve a `idle`, **no queda colgada en streaming** |
| 4 | Restaurar la carpeta, mandar otro turno | Anda |
| 5 | Desvincular el arnés desde la vista Portafolio | La sesión **sigue en el rail con sus N conversaciones** — **nada en cascada** |
| 6 | `curl .../conversaciones` | Las N siguen ahí |

---

#### **E2E-6 · `--resume` que el CLI ya no reconoce** (E-13) — *con `claude` real*

| # | Paso | Aserción |
|---|---|---|
| 1 | Retomar una conversación cuyo `claude_session_id` fue GC'd (o falsearlo en el registro con el daemon detenido) | — |
| 2 | Mandar un turno | El spawn con `--resume` muere antes del `init`; `tryHealResume` respawnea **fresh una vez** y **reenvía el turno** |
| 3 | Mirar el transcript | Los turnos viejos **siguen** (vienen de `Conv`, no del proceso) |
| 4 | Mirar la marca | `⟳ hilo reiniciado · checkpoint` inline — el cambio de `cc-id` **no es silencioso** (RF-348 CA-1) |
| 5 | Mirar el detalle | `cc-id` **NUEVO** |
| 6 | Forzar un **segundo** fallo | `error` visible, **no un loop** |

---

## 4 · Datos de prueba, y cómo no ensuciar el `~/.arnesia` del operador

**La regla:** el E2E **nunca** escribe en el `~/.arnesia` real. El aislamiento es por `HOME`, no por
flags — el daemon resuelve `~/.arnesia` con `os.UserHomeDir()` (`registry.go:34`) y
`~/.claude/projects` con `history.New("")` (`reader.go:31`), y los dos leen `$HOME` en Linux.

| Dato | De dónde sale | Cómo se siembra |
|---|---|---|
| `sessions.json` v1 (5 sesiones, 17 202 B) | **copia del archivo real del operador** | `cp` (jamás `mv`) a `$ARNESIA_E2E/.arnesia/` |
| El transcript de 90 turnos (12 281 B) | `s6165ac75` del archivo real | Viaja dentro del anterior; y `internal/usecase/testdata/conv-90-turnos.json` para el test Go de E-25 |
| `sesiones-cerradas.json` | **NO se siembra** | CV-D6 lo borra. Resucitarlo en el E2E sería probar algo que se decidió eliminar |
| Turnos nuevos | `mock-claude.sh` (`stories/2026-07-08-chat-cc-funcional/e2e/`) | `PATH` con el mock adelante |
| E2E-3 y E2E-6 | `claude` **real** | Lo que prueban es el `--resume` y la rotación reales |
| Arnés de trabajo | `dogfood/dev-full-cycle` del repo | Registrado en el Portafolio del sandbox al arrancar |

**Fixtures Go: prohibido inventar un transcript sintético cuando hay uno real.** El de 90 turnos se
copia a `testdata/`, no se genera. Es la disciplina que el paquete de marketplace ya fijó.

⚠ **Antes de agregar cualquier fixture, revisar `.gitignore`.** HS-28 destapó que `main` no compilaba
porque `.gitignore` se comía un fixture de tests y cuatro paquetes nunca se habían commiteado.
**Verificar contra un worktree limpio** (`git status --ignored` sobre `testdata/`).

**Riesgo residual, declarado:** el Modo B corre con el `HOME` real porque la ventana Tauri se lanza
desde el `.desktop` y no hereda un `HOME` inyectado. Por eso el Modo B va **último, después de T32**,
es **de sólo mirar**, y su lista de pasos se limita a E2E-1..3. Todo lo destructivo (E2E-0, E2E-5)
corre **exclusivamente** en Modo A.

---

## 5 · Criterio de salida — qué tiene que estar verde para firmar PARIDAD

### 5.1 Los gates técnicos (todos, sin excepción)

| # | Check | Comando | Criterio |
|---|---|---|---|
| 1 | Go completo con carreras | `go test ./... -race` | **0 FAIL**. Hoy: 26 ok · 8 sin tests · 0 FAIL — y `internal/adapters/store` deja de estar en «sin tests» |
| 2 | Grafo de imports | `go run github.com/fe3dback/go-arch-lint@latest check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml` | 0 violaciones. Es **allow-list default-deny**: un adapter nuevo sin entrada se rechaza por ausencia |
| 3 | Estáticos FE | `pnpm --dir web run verify` | verde (tsc · biome · depcruise · steiger · stylelint). ⚠ **no corre tests** |
| 4 | Unit FE | `pnpm --dir web exec vitest --project=unit run` | **todos**, incluidos los 16 nuevos |
| 5 | **Fitness visual** | `pnpm --dir web exec vitest --project=storybook run` | **0 failed**. Es el step que rompe CI desde 2026-07-20 |
| 6 | a11y sin escapes | `grep -rn 'a11y' web/src/widgets/chat-dock/ui/*.stories.tsx` | **ninguna** baja a `"todo"` (RF-353 CA-4) |
| 7 | Capabilities R1-R4 | `go test ./docs/architecture/fitness/ -run TestCapability` | los **5** verdes |
| 8 | Contrato OpenAPI | `go test ./docs/architecture/fitness/ -run 'TestRuta\|TestQueryParam\|TestExencion'` | los **4** verdes · `_sin-declarar.yaml` en **10** entradas `/api` (de 11) |
| 9 | Boundary de sesión | `go test ./docs/architecture/fitness/ -run 'TestOneTurn\|TestFrames\|TestResume\|TestNoSilent\|TestTransicion'` | **5/5** — los 4 originales **con el cuerpo intacto** |
| 10 | Drift de cifras | `bash scripts/estado.sh --check` | exit 0 |
| 11 | Changelog | `go test ./docs/architecture/fitness/ -run TestChangelog` | verde |
| 12 | **CI** | `gh run list --limit 1` | **3/3 jobs verdes** (`go`, `ts`, `rust`) en la corrida del commit final |

### 5.2 Los gates de comportamiento

| # | Check | Criterio |
|---|---|---|
| 13 | Los **6 E2E** de §3.3 | Pasan contra el binario **instalado**, con la aserción de identidad del sello (paso 8 del circuito) verde |
| 14 | **Modo B** | La ventana Tauri instalada muestra lo construido; E2E-1..3 repetidos a ojo |
| 15 | Los **50 escenarios** | Cada uno con su ticket y su mecanismo (`plan-desarrollo.md` §Cobertura). **E-46 declarado sin test**, con su rastro |
| 16 | Las **56 stories** | Los 42 estados visuales de `design.md` §4 con al menos una story cada uno |
| 17 | **Superset estricto** | El diff de `chat-dock.tsx` no toca `Composer` (`:268-375`), `fitComposer` (`:263-266`), `Bubble` (`:233-258`), `ActivityCard` (`:183-231`), `agrupar` (`:160-172`), `rotulo` (`:176-179`). Verificable con `git diff` |
| 18 | Vocabulario | `grep -ri "cerrada" web/src/widgets/chat-dock/` → **0 copy visible** (RF-302 CA-1) |
| 19 | Cromo | **2 filas** en reposo, **3** con nodo (C-2). Medido en la app, no en una story |

### 5.3 Los gates de honestidad — lo que hay que **declarar**, no arreglar

| # | Qué se declara en `PARIDAD.md` | Por qué |
|---|---|---|
| 20 | Las **3 desviaciones** de `spec.md` §Estado y las **13 contradicciones** de `design.md` §2, con su veredicto | Ninguna se resolvió en silencio |
| 21 | Que las 4 stories del picker se pusieron verdes en **T3 por contraste corregido**, **no** por la desaparición del bloque (T30), y que **la deuda del token `--warn` sigue abierta** (`BACKLOG.md:59-64`) | Es el error de lectura más probable de todo el paquete |
| 22 | La tabla de **5 filas del dry-run de CV-D16**, con los motivos, incluidas las que **no se pudieron resolver** | CV-D16 dice «nada se borra»; la tabla lo hace observable |
| 23 | El **número** del benchmark de `persistLocked` (T14) | Sin la medición, «alcanza» es un pass fabricado |
| 24 | El rastro de **E-46**: fecha, los 3 ids (`s1b38a066`, `s408bb085`, `s020210e3`), la ruta del `.bak`, el resultado del `GET` | Es el único escenario sin test |
| 25 | Los **9 huecos** de `arquitectura.md` §9 (H-A…H-I), con los que van al BACKLOG | Gaps VISIBLES |
| 26 | Que `TestTransicionDeConversacionEsAtomica` **nació con el paquete**: el boundary v1.2 se declaró con 4/5 enforcers y gradúa a 5/5 en T15 | El nodo no se declaró `enforced` antes de tener el código |

### 5.4 La firma

**PARIDAD se firma cuando los 26 están verdes o declarados.** La firma es **del operador**: el
ejecutor la registra tras verificar que el artefacto existe en `main`, y **jamás** fabrica un
click-through. Un check que no se corrió se anota como *no corrido*, no como *verde*.

**Y una condición que no está en ninguna tabla:** el operador tiene que poder abrir la app instalada,
apretar `＋`, ver la conversación anterior en la lista, buscar una palabra que dijo hace tres días y
volver a ese hilo. Si eso no pasa, los 26 checks verdes no significan nada.
