# PARIDAD — Las conversaciones viven en el panel de conversación

> `tipo: paridad` · paquete `2026-07-26-conversaciones-del-panel` · abierta 2026-07-26.
> Par del mockup 🧑‍⚖️ [`mockup-conversaciones-panel.html`](./mockup-conversaciones-panel.html) y del
> [`plan-desarrollo.md`](./plan-desarrollo.md) (33 tickets · 6 tramos).
>
> **ESTADO: EL PAQUETE ESTÁ A MEDIO CONSTRUIR.** Cerrado y verde el **tramo 0** (T1-T6, la línea
> base). Los tramos 1 a 5 **no se empezaron**. Este archivo se abre ahora, y no al final, porque la
> evidencia del tramo 0 es perecedera: las cifras que lo justifican (CI rojo, 3/7 stories, 11 rutas
> sin declarar) dejan de ser observables en cuanto se corrigen.
>
> ⚠ **El gate 🧑‍⚖️ de PARIDAD está SIN FIRMAR y no se puede firmar todavía** — no hay superficie
> nueva que comparar contra el dibujo. Lo que sigue no es una PARIDAD completa: es su primera
> mitad, la de la línea base, más las desviaciones ya declaradas.

---

## 0 · Qué se construyó y qué no

| Tramo | Tickets | Estado | Gate |
|---|---|---|---|
| **0** llegar a verde | T1-T6 | ✅ **completo y verde** | 🧑‍⚖️ parcial — ver §1.6 |
| **1** modelo y disco | T7-T14 | ⬜ **no empezado** | — |
| **2** usecase y API | T15-T20 | ⬜ no empezado | — |
| **3** cromo del dock | T21-T25 | ⬜ no empezado | — |
| **4** lista y buscador | T26-T29 | ⬜ no empezado | — |
| **5** mudanza y cierre | T30-T33 | ⬜ no empezado | — |

**Commits del tramo 0** (locales, `main`, **sin pushear** por decisión del operador):

| commit | qué |
|---|---|
| `bc9d1d4` | los artefactos de spec/diseño/arquitectura/planes + los 2 boundaries nuevos, `proposed` |
| `56fdda1` | T2 (baseline del dock) + T3 (contraste del picker) |
| `eda01f0` | T4 (allowlist del contrato) + T5 (los 4 enforcers, `enforced` 4/5) |

---

## 1 · Tramo 0, ticket por ticket, con su evidencia

### T1 · el gate local

`lefthook install` corrido. `.git/hooks/` tenía sólo los 14 `.sample` y `core.hooksPath` estaba sin
setear: **los 8 jobs pre-commit no habían corrido nunca en este working copy**. Ahora existen
`pre-commit` y `commit-msg`, y **dispararon de verdad** en los 3 commits — no es una afirmación, es
lo que hizo fallar el primer intento de commit de T4/T5 (ver T5).

### T2 · el baseline del dock (H-1)

`web/src/widgets/chat-dock/ui/chat-dock.stories.tsx` — A-01 `VigenteReposo`, A-02
`VigenteConPermiso`, A-03 `VigenteStreaming`. **3/3 verdes contra el código vigente, antes del
primer cambio.** CAP-68 pasa de `vivo·nc` a `vivo` con su `valida:` en el mismo commit (R4).

🔴 **Hallazgo, y es el que justifica el ticket entero.** Darle su primera story al dock destapó una
**violación a11y real y preexistente del producto que se envía hoy**: `SessionLine` pinta el
`◍ <cc-id>` con `text-primary` sobre `bg-secondary` — **2,21:1** medido por axe (`#00b7aa` sobre
`#eef1f1`), contra el mínimo 4,5 de texto. Invisible durante toda la vida del widget porque no
tenía story.

- **No se tapó:** anotada en `BACKLOG.md` y en el cuerpo de CAP-68.
- **No se arregló acá:** tocar `SessionLine` es superficie vigente fuera del ticket. Se corrige al
  construir `IdentidadDetalle` (T23/RF-328), adonde ese cuerpo se muda.
- **Las 3 stories apagan sólo la regla `color-contrast`**, no bajan el gate a `"todo"`: el resto de
  axe sigue en `error`. Las stories de la superficie nueva corren con el gate entero (RF-353 CA-4).

### T3 · CI a verde por la razón correcta

Los dos `text-warn` de `new-session-picker.tsx` eran **la causa medida** del CI rojo: `#c96a2e`
sobre `#ffffff` a 10 px = **3,76:1**. Se les aplicó **C-3**, la regla que este mismo paquete ya
había decidido para su superficie nueva: el motivo en `--foreground` (18,74:1) y la alarma en un
borde izquierdo de 3 px, que al no ser texto sólo necesita 3:1.

    new-session-picker.stories.tsx    3/7  →  7/7
    --project=storybook completo      43 archivos · 386/386 · 0 failed

> ⚠ **El error de lectura más probable de todo el paquete, declarado por adelantado (criterio 21
> del plan de pruebas):** las 4 stories pasan **por contraste corregido**, NO por la desaparición
> del bloque `ConversacionesDelArnes` — eso ocurre recién en T30. Y **la deuda del TOKEN `--warn`
> sigue abierta** (`BACKLOG.md`): se corrigió el uso, no el valor del token.

### T4 · el contrato declarado como ES HOY

`docs/architecture/contracts/api/_sin-declarar.yaml`, **15 entradas, ninguna razón vacía**: 8 de
telemetría (módulo abierto, declara al cerrar), 1 con fecha de retiro
(`GET /api/sessions/cerradas/{id}/historial`, se va en T18), 1 de deuda pura
(`POST /api/portafolio/arneses/{clave}/identificar`) y 5 estructurales (`/v1/logs` y `/v1/metrics`,
contrato OTLP ajeno · `GET /events`, alias sin prefijo · `GET /healthz`, sonda · `/`, el SPA).

### T5 · los 4 enforcers, y la graduación del boundary

`docs/architecture/fitness/openapi_contract_test.go` — 4/4 verdes.
`ruta-servida-esta-declarada` pasa de `proposed` a **`enforced` con 4 de 5**;
`forma-de-respuesta-unica` **difiere honesto** (H-H: sin mecanismo determinista) y su
`enforced_by` lo dice.

**La graduación no se firmó leyendo el test.** Validación por fallo inducido, corrida:

    con `GET /api/ruta-falsa-de-prueba` inyectada en router.go
      → FAIL: TestRutaServidaEstaDeclarada — «1 ruta(s) servidas sin declarar … GET /api/ruta-falsa-de-prueba»
    quitada
      → ok

🔴 **Dos correcciones que la construcción le hizo al propio nodo de arquitectura** — se registran
porque el documento firmado decía otra cosa:

1. **«50 rutas servidas» era 54.** El conteo original sólo miraba `mux.HandleFunc(`; `mux.Handle(`
   aportaba 4. El **saldo de deuda no cambia** (siguen 11 `/api` sin declarar), la cifra sí.
2. **`query-param-declarado` nació VACUO y hubo que rehacerlo.** La primera versión cruzaba «rutas
   registradas en el archivo X» con «queries leídas en el archivo X»; en este árbol `router.go`
   registra todas las rutas sin leer un solo query y `sessions.go` lee dos sin registrar ninguna.
   Intersección vacía ⇒ **verde mentiroso**. Se rehízo por AST (ruta → identificador del
   constructor del handler → cuerpo de esa función, closures incluidas) y al rehacerlo destapó
   inmediatamente `?arnes=` y `?cerradas=` de `GET /api/sessions`, **servidos desde siempre y jamás
   declarados**. Quedan declarados tal como se comportan hoy, bimorfia incluida.

**Decisión explícita que T5 pedía tomar y no omitir:** el enforcer **no estrena capability**. Vive
en `docs/architecture/fitness/`, que R2 no escanea por construcción (`TestCapabilityCoverage`
recorre `cmd`, `internal`, `web/src`, `web/src-tauri/src`), y ningún otro fitness del repo tiene
hoja propia. Los 4 tests entran como `valida:` de **CAP-52**, que además corrige su nombre: decía
«23 endpoints» y son 54.

### T6 · el gate de línea base — ✅ local completo · ⚠ CI **NO VERIFICADO**

Los 9 comandos que corren los 3 jobs de CI, corridos localmente sobre el árbol final:

| # | comando (job de CI) | resultado |
|---|---|---|
| 1 | `go test ./... -race` | **exit 0** |
| 2 | `go build ./...` | **exit 0** |
| 3 | `go-arch-lint check` | **OK — No warnings found** |
| 4 | `bash scripts/estado.sh --check` | **exit 0** — cifras en sync |
| 5 | `pnpm exec biome ci .` | 169 archivos · 0 errores |
| 6 | `pnpm exec tsc --noEmit` | **exit 0** |
| 7 | `pnpm exec vitest --project=unit run` | **122/122** |
| 8 | `pnpm exec vitest --project=storybook run` | **386/386 · 43 archivos · 0 failed** |
| 9 | `pnpm run tokens:build && git diff --exit-code` | **exit 0** |
| — | `pnpm run verify` (depcruise · steiger · stylelint) | verde |
| — | `cargo fmt --check` | **exit 0** (y el diff contra Rust es vacío: no se tocó) |

🔴 **Lo que NO se hizo, y por qué:** T6 exige `git push` + `gh run watch` con **3/3 jobs verdes**.
**No se pusheó.** La instrucción del encargo es explícita: los 79 commits sin pushear son decisión
del operador. Por lo tanto:

> **El estado real de CI es: rojo desde 2026-07-20, y NO SE VERIFICÓ que este trabajo lo ponga en
> verde.** Lo que sí está medido es que **la causa conocida del rojo está corregida**: la última
> corrida (`30140548006`) falló **sólo** en el job `ts`, paso `fitness visual`, con `go` ✓ y `rust`
> ✓; ese paso corre exactamente el comando 8 de la tabla, que hoy da 386/386 local. Es una
> inferencia fuerte, **no una observación**, y se anota como tal. El gate 🧑‍⚖️ de T6 queda
> **abierto** hasta que el operador pushee y la corrida se vea.

⚠ **Hallazgo colateral, declarado:** el `golangci-lint` **local** (v2.12.2) reporta ~25 hallazgos
preexistentes en archivos que este paquete no toca (`internal/adapters/telemetria/**`,
`internal/adapters/logfile`, `internal/adapters/selfupdate`, y `golangci-lint fmt --diff` marca 3
archivos de telemetría). **El job `go` de CI está verde con esos mismos archivos**, así que es
drift de versión/configuración entre el binario local y `golangci-lint-action@v7`, no una
regresión. No se tocó nada fuera de alcance para «arreglarlo». Lo que sí se arregló son **los 5
hallazgos propios** que el hook `pre-commit` cazó en `openapi_contract_test.go` (3 × G304, 1
`parser.ParseDir` deprecado desde Go 1.25, 1 regex muerta) — y se arreglaron de verdad, no con un
`//nolint` de conveniencia: la regex se borró, el `ParseDir` se reemplazó por `ParseFile` archivo a
archivo, y sólo los 3 G304 llevan `//nolint` **con su justificación escrita**, que es lo que
`nolintlint` exige.

---

## 2 · Mockup §panel → producto

**Ninguna fila se puede marcar ✅ todavía: la superficie nueva no existe.** La tabla se deja armada
con su trazabilidad para que el que siga la complete ticket a ticket, y para que quede claro que
está vacía **por no construida**, no por no mirada.

| § del mockup | qué dibuja | componente destino | ticket | veredicto |
|---|---|---|---|---|
| §1 `vigente` | el dock de hoy, 4 filas, `⟩ colapsar` | `chat-dock.tsx` (as-is) | T2 | ✅ **fijado por A-01..A-03** |
| §2A/§2C | cromo de **2 filas** en reposo, **3** con nodo | `ChatDock` + `ScopeRow` | T25 | ❌ no construido |
| §2 fila 2 | `▶ título · chip ctx · 🔍 · ＋` | `ConversacionRow` | T24 | ❌ no construido |
| §2 chip | ctx como chip-disclosure | `CtxChip` | T23 | ❌ no construido |
| §2 detalle | `cc-id · arnés · modelo · cwd` | `IdentidadDetalle` | T23 | ❌ no construido |
| §3A | lista en sitio, 4 conversaciones | `ConversacionesPanel` | T28 | ❌ no construido |
| §3 fila | título · última interacción · turnos · ctx | `ConversacionFila` | T27 | ❌ no construido |
| §3D | una sola conversación, sin buscador | `ConversacionesPanel` | T28 | ❌ no construido |
| §4A | buscador con `<mark>` | `ConversacionesPanel` | T28 | ❌ no construido |
| §4B | sin coincidencias, dice dónde buscó | `ConversacionesPanel` | T28 | ❌ no construido |
| §4C | marca de rotación inline | `Messages` + frame `conversacion` | T16/T29 | ❌ no construido |
| §5 | la mudanza declarada | `new-session-picker.tsx` | T30 | ❌ no construido |
| §6B | error de lectura con Reintentar | `ConversacionesPanel` | T28 | ❌ no construido |
| §glifo | `⟩` → `»` | `chat-dock.tsx` | T25 | ❌ no construido |

---

## 3 · Desviaciones y huecos declarados

### 3.1 Las 3 desviaciones de `spec.md` §Estado — vigentes, sin cambios

| # | desviación | estado |
|---|---|---|
| 1 | CV-D8 midió «90 turnos = 9,4 KB»; hoy son **12,0 KB** | declarada; la conclusión (scan en memoria, sin FTS5) no cambia |
| 2 | el mockup pinta el error de lectura en `--warn`, que rompe el gate a11y | **gana el gate** (C-3). **Ya realizado en T3** sobre el consumidor vigente |
| 3 | CV-D14 dice «4 filas → 2»; con nodo son **3** | declarada (C-2): el «2» es el reposo |

### 3.2 Las 13 contradicciones de `design.md` §2

Las 13 conservan el veredicto del documento. **Realizadas hasta ahora: C-3** (T3, sobre el picker).
Las otras 12 se realizan en los tramos 2-5 y se anotan acá cuando ocurran.

### 3.3 Los 9 huecos de `arquitectura.md` §9 (H-A…H-I)

Los 9 siguen **abiertos y declarados**. Ninguno se resolvió en el tramo 0 y ninguno se tapó.
H-H (`forma-de-respuesta-unica` sin enforcer determinista) es el único que ya tiene su rastro
as-code: figura como celda `(pendiente — juicio)` en el boundary y fuera de su `enforced_by`.

### 3.4 Hallazgos NUEVOS de la construcción, que no estaban en ningún documento

| # | hallazgo | dónde quedó |
|---|---|---|
| N-1 | `SessionLine` viola contraste con `text-primary` (**2,21:1**) — bug a11y **en producción hoy** | `BACKLOG.md` + CAP-68 + §1 T2 |
| N-2 | el nodo decía 50 rutas servidas; **son 54** | corregido en el boundary v1.1 + CAP-52 |
| N-3 | `query-param-declarado` nació **vacuo**; rehecho por AST | boundary v1.1 changelog |
| N-4 | `?arnes=` y `?cerradas=` se servían **sin declarar** | declarados en `openapi.yaml` |
| N-5 | `golangci-lint` local (v2.12.2) ≠ el de CI: ~25 hallazgos preexistentes locales, CI verde | §1 T6 |

---

## 4 · Lo que este documento TODAVÍA NO PUEDE decir

Se listan para que nadie los lea como verdes:

| criterio (plan-pruebas §5) | estado |
|---|---|
| 13 · los **6 guiones E2E** contra el binario instalado | 🔴 **NO CORRIDOS.** Es T31, tramo 5. Validan una superficie que no existe; correrlos hoy sólo mediría el arreglo de contraste, a costa de reemplazar el binario instalado del operador (`make dev-sync`) y matarle el daemon. Se decidió **no hacerlo**: cero valor probatorio, costo real |
| 14 · Modo B (ventana Tauri, gate humano) | 🔴 no corrido — depende del 13 |
| 15 · los 50 escenarios E-01…E-50 | 🔴 0 de 50 verificados |
| 16 · las 56 stories | ⚠ **3 de 56** (las A-01..A-03 del baseline, que no son de la superficie nueva) |
| 22 · la tabla de 5 filas del **dry-run de CV-D16** | 🔴 no corrida — es T11 |
| 23 · el **número** del benchmark de `persistLocked` | 🔴 no medido — es T14 |
| 24 · el rastro de **E-46** (borrado CV-D6) | 🔴 no ejecutado — es T32, **procedimiento manual del operador**. Los 3 ids (`s1b38a066`, `s408bb085`, `s020210e3`) siguen en `~/.arnesia/sesiones-cerradas.json`, **intactos**: este trabajo no tocó un solo byte de `~/.arnesia/` |
| 26 · `TestTransicionDeConversacionEsAtomica` | 🔴 no escrito — es T15. `sesion-viva-consistente` sigue en **4/5 declarado**, que es lo correcto |

---

## 5 · Gate 🧑‍⚖️

**SIN FIRMAR, y no corresponde firmarlo.** El paquete está en el 17 % de sus tickets (6 de 33) y no
tiene superficie que comparar contra el dibujo firmado. La firma de PARIDAD es del operador y se
pide cuando los 26 criterios estén verdes o declarados; hoy hay 7 que ni siquiera son evaluables.

Lo único que sí admite firma parcial es el **gate de línea base de T6**, y también queda abierto:
el local está completo y verde, pero **CI no se observó** porque pushear es decisión del operador.
