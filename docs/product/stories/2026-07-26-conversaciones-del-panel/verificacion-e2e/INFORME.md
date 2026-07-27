# Verificación E2E contra la app INSTALADA — `conversaciones-del-panel`

> `tipo: verificacion-e2e` · paquete `2026-07-26-conversaciones-del-panel` · **2026-07-26/27**
> Guion: [`plan-pruebas.md`](../plan-pruebas.md) §3. Escenarios: [`spec.md`](../spec.md) E-01…E-50.
>
> **Qué contesta este documento, y nada más:** *«cuando yo instale el app, ¿voy a ver lo que
> construiste?»*. Todo lo de acá se midió contra `~/.local/bin/arnesia` —el binario que la
> ventana Tauri del operador ejecuta— sirviendo su SPA embebida. **Nada se midió contra
> `pnpm dev`, contra Storybook ni contra un mock del daemon.**
>
> **Contrato de honestidad:** un paso que no se pudo correr figura como `n/c` **con su
> motivo**, jamás como verde. Un paso que falló figura como falla **con su evidencia cruda**.

---

## 0 · Lo primero: los datos del operador

| Qué | Antes de empezar | Al terminar |
|---|---|---|
| **md5 de `~/.arnesia/sessions.json`** | `b1689d1513e8d3c3fa21692c511ed793` | **`b1689d1513e8d3c3fa21692c511ed793`** ✅ **idéntico** |

Verificado además que el `~/.arnesia` real **no ganó ni un archivo** del circuito: no hay
`sesiones.json` (el formato nuevo), no hay `sessions.json.v1-*.bak`, no hay `mock-*`. La
migración corrió **exclusivamente sobre la copia**.

**Cómo se logró el aislamiento.** No por disciplina al tipear: por construcción.

- El daemon corrió con **`HOME` apuntando a un sandbox** (`env -i HOME=$SB …`). El daemon
  resuelve `~/.arnesia` con `os.UserHomeDir()` (`registry.go`), que en Linux es `$HOME`; y
  `os.UserConfigDir()` (la ficha de telemetría) cae en `$SB/.config`. **No queda nadie
  apuntando al home real.**
- ⚠ **`--sessions` NO sirve para probar la migración**, y conviene saberlo antes de intentarlo:
  `rutasDelRegistro` (`cmd/arnesia/main.go:814-824`) devuelve `legado = ""` cuando el flag
  viene puesto, así que el camino v1→v2 **queda desactivado**. El aislamiento por `HOME` es
  el único que ejercita la migración de verdad. (El relevo anterior tenía razón en que el
  flag existe; lo que no se ve desde afuera es que apagar el legado es su efecto.)
- El **arnés de prueba vive dentro del sandbox** (`$SB/arneses-e2e/vitalia`). E2E-5 mueve
  esa carpeta. **El registro real apuntaba a `/home/chalreme/Proyectos/luana-vitalia/vitalia`
  y a `…/harness-studio/dogfood/…`: mover el proyecto real del operador para probar un
  mensaje de error habría sido el daño que todo esto evita.** Se repuntó `arneses.json` y el
  `cwd` persistido, los dos **sólo en la copia**.

### El binario instalado: respaldo y restauración

```
respaldo:  ~/.local/bin/arnesia.respaldo-20260726-231513
           md5 090bb5d611531e68be36f4f92fcba797  ·  18 538 761 B  ·  mtime jul 26 17:24
```

Es el estado que **otra sesión dejó hoy a las 17:24**, no el original de fábrica.
**Para restaurarlo:**

```bash
pkill -f "$HOME/.local/bin/[a]rnesia serve"          # bajar el daemon vivo
cp -p ~/.local/bin/arnesia.respaldo-20260726-231513 ~/.local/bin/arnesia
md5sum ~/.local/bin/arnesia                          # debe dar 090bb5d6…
```

Y volver a abrir la app: el shell Tauri relanza el daemon solo.

---

## 1 · El circuito que se corrió

```
respaldo del binario  →  make dev-sync  →  sandbox con COPIA de los datos  →  6 guiones
```

**`make installer` NO se corrió, a propósito.** Depende de `bump-patch`
(`Makefile:70`), y el encargo prohíbe `make bump-*`. Para el **Modo A** no hace falta:
`make dev-sync` corre `bundle.sh --daemon-only`, que compila **la misma SPA** (`vite build`
→ `web/dist`) y **el mismo daemon** con esa SPA embebida y el mismo sello `-ldflags`. Lo
único que `make installer` agrega es el empaquetado `.deb/.rpm/.AppImage` — y el shell
instalado **prefiere siempre `~/.local/bin/arnesia` sobre el sidecar** (`Makefile:27-32`),
así que lo que el operador ejecutaría es exactamente el binario que acá se midió.

**Reproducirlo:** `bash verificacion-e2e/correr.sh [1..6]` (arma su propio sandbox, verifica
el sello antes de medir, y al terminar imprime el md5 del registro del operador).

### Aserción de identidad — obligatoria y primera

```json
GET /api/version
{"version":"0.2.24.2607262315","compilado":"2026-07-26 23:15",
 "huella":"657054d+sucio","instalado_en":"/home/chalreme/.local/bin/arnesia"}
```

**Sello del binario probado: `0.2.24.2607262315`** (0.2.24 + build 2026-07-26 23:15), el
minuto en que `make dev-sync` lo compiló. **Si el sello hubiera sido viejo se estaba midiendo
otro binario y el circuito abortaba.**

### El SPA servido ES el de esta rama — probado, no supuesto

| Evidencia | Resultado |
|---|---|
| `GET /` → `<script src="/assets/index-CUTbg9Li.js">` | el bundle que el binario sirve |
| md5 del JS **servido por el daemon** vs md5 de `web/dist/assets/index-CUTbg9Li.js` recién construido | **`af7104b5d4a9d188bcba9666c1381c1e` en los dos — byte a byte idéntico** |

Y adentro de ese bundle servido, la superficie nueva:

| String propio del cromo nuevo | En el JS servido |
|---|---|
| `esperá a que termine el turno` · `esperá tu decisión de permiso` | ✅ |
| `Retomando la conversación` | ✅ |
| `Ninguna conversación de esta sesión menciona` · `Se buscó en el título y en el texto` · `Limpiar búsqueda` | ✅ |
| `hay un turno en vuelo — esperá a que termine` (la traducción del 409) | ✅ |
| `»` (glifo de colapsar, CV-D15) | ✅ |
| **`cerrada`/`cerradas` en copy visible** (RF-302 CA-1, criterio 18) | **0 ocurrencias** ✅ |

`nueva conversación` y `— contexto rotado, seguimos —` **no** están en el JS y eso es lo
correcto: son constantes de Go (`domain/conversacion.go:35`, `session_rotacion.go:12`) —
verificadas **dentro del binario instalado**.

---

## 2 · Los 6 guiones

**Total: 98 aserciones · 92 ok · 3 fallas · 3 no corridas.**

| Guion | ok | falla | n/c | Veredicto |
|---|---|---|---|---|
| **E2E-0** migración del disco del operador | 12 | 0 | 0 | ✅ |
| **E2E-1** crear · listar · buscar · retomar | 31 | 0 | 0 | ✅ |
| **E2E-2** turno en vuelo · 409 | 17 | 1 | 0 | ⚠ 1 desviación de copy |
| **E2E-3** rotación EN VIVO | 21 | 1 | 0 | ⚠ **cierra C-6/H-8** + 1 bug nuevo |
| **E2E-4** dos vistas | 8 | 0 | 1 | ✅ (1 no ejercitable) |
| **E2E-5** fallos del entorno | 13 | 0 | 0 | ✅ |
| **E2E-6** `--resume` no reconocido | 8 | 1 | 1 | ⚠ RF-348 CA-1 sin construir |

---

### E2E-0 · La migración del disco del operador — ✅ *el guion que más importa*

Sobre la **copia del registro real**: 5 sesiones, 17 202 B, formato v1.

| # | Paso | Aserción | Resultado |
|---|---|---|---|
| 1 | `sesiones recalibrar-llaves` (dry-run) **antes** de arrancar | 5 filas | ⚠ **el plan tiene el paso en el orden equivocado** — ver abajo |
| 2 | log del arranque | dice qué migró y dónde quedó el respaldo | ✅ `desde_version=1 hasta_version=2 respaldo=…/sessions.json.v1-2607262315.bak` |
| 3 | `ls .arnesia/` | están `sesiones.json`, `sessions.json` y el `.bak` | ✅ los 3 |
| 4 | `jq .schema_version sesiones.json` | `2` | ✅ `2` (+ `escrito_por`, `escrito_en`) |
| 5 | `diff` del v1 contra la copia previa | vacío | ✅ **idéntico**, md5 `b1689d15…` — **volver a un binario anterior es gratis** |
| 6 | `GET /api/sessions \| length` | 5 | ✅ `5` — **ninguna sesión se perdió** |
| 7 | una activa por sesión | 5 | ✅ `5` |
| 8 | `s6165ac75 → .activa.conv \| length` | 90 | ✅ **`90` — el transcript de 90 turnos sobrevivió entero** |
| 9 | `.activa.ultima_interaccion` | ausente | ✅ `null` — **no se inventó una fecha** (H-4/H-B) |
| 10 | `s0fec7798 → .activa.conv` | `[]`, no `null` | ✅ `array`, largo 0 (`no-aplica-no-es-cero`) |
| 11 | llaves de arnés | ninguna perdida ni fusionada | ✅ las 5 intactas |
| 12 | reiniciar y repetir 4-11 | idéntico | ✅ **byte a byte** (`f2d42d4b…` antes y después), 0 re-migraciones, **1 solo `.bak`** |

**El transcript de 90 turnos, antes y después: 90 → 90.** Conteo total de sesiones: 5 → 5.

#### ⚠ Hallazgo N-18 · el dry-run de CV-D16 no puede correr *antes* del daemon

`plan-pruebas.md` §3.3 paso 1 pide el dry-run **antes de arrancar**. **No es posible**, y el
comando no lo dice: `cmdSesiones` (`cmd/arnesia/sesiones.go:131-142`) abre **sólo**
`~/.arnesia/sesiones.json` con `store.NewRegistry` —el archivo **v2**, que antes del primer
arranque **no existe**— y `NewRegistry` sobre una ruta ausente devuelve un registro vacío.
Resultado observado, literal:

```
DRY-RUN — no se escribió nada. Corré con --aplicar para hacerlo.
no hay sesiones en el registro.
```

Al operador que quiere **previsualizar antes de actualizar** le contesta «no hay sesiones»,
que se lee como *«tu registro está vacío»* y no como *«todavía no hay registro nuevo, arrancá
el daemon primero»*. **El orden correcto es: daemon (migra) → dry-run → `--aplicar`.**

#### La tabla de las 5 filas de CV-D16 (criterio de salida 22) — corrida sobre la copia

```
sesión        antes                         después                       motivo
s25123a2c     vitalia                      →sin-home~vitalia~vitalia      resuelta-por-cwd
s6165ac75     sin-home~vitalia~vitalia      sin-home~vitalia~vitalia      ya-calificada
s0fec7798     vitalia                      →sin-home~vitalia~vitalia      resuelta-por-id
sfc512b15     vitalia                      →sin-home~vitalia~vitalia      resuelta-por-id
s78b3aeeb     arnesia                       arnesia                       sin-candidata

3 de 5 sesiones cambian de llave. Nada se borra y nada se fusiona.
```

**Las 5 salen nombradas, incluidas las 2 que no se movieron** — que es lo que CV-D16 pedía
hacer observable. `s78b3aeeb` queda `sin-candidata`: **no se resolvió, y se ve.**

#### CV-D16 en vivo: el re-key es del OPERADOR, no del arranque

El arranque **no re-keya**, y es deliberado: `main.go:202` llama
`store.AbrirRegistro(…, nil)` con la clave en `nil`, y `migracion.go:224` sólo recalibra
`if inf.Migro && clave != nil`. El comentario de `main.go:198-200` lo explica — atar el
arranque a que el Portafolio responda es justo lo que el paso separado evita.

**Consecuencia práctica para el operador, y hay que decírsela:** tras instalar, las 3
sesiones con llave a medias **siguen invisibles** hasta que corra a mano
`arnesia sesiones recalibrar-llaves --aplicar`. El paso **no está automatizado y no avisa solo**.

---

### E2E-1 · Crear, listar, buscar y retomar — ✅ 31/31

*Capturas: `e2e1-01`…`e2e1-07`.*

| # | Aserción | Resultado |
|---|---|---|
| 1 | **exactamente 2 filas de cromo** (CV-D14) | ✅ **2** medidas en el DOM |
| 1 | `queryByText("Alcance:") === null` · ningún `◍` · chip de ctx presente | ✅ · ✅ · ✅ |
| 1 | el glifo es **`»`** y no `⟩` (CV-D15) | ✅ `» colapsar` |
| 2 | mandar un turno hace crecer el transcript | ✅ 4 → 7 turnos, y el texto se pinta |
| 3 | `＋` → `nueva conversación`, 0 turnos, **chip a 0% VISIBLE** | ✅ · ✅ · ✅ `chip="0%"` |
| 3 | sigue habiendo **una** activa (CV-D7) | ✅ |
| 4 | `getByRole("listbox")` con 2 `option` | ✅ (contrato ARIA sobre `<div>`, N-17) |
| 4 | **el composer sigue visible** · **0 `<dialog>`** · **0 backdrop** | ✅ · ✅ · ✅ **abre en sitio, no es un modal** |
| 5 | «N de 2 coinciden» **sin apretar Enter** · `<mark>` | ✅ `1 de 2 coinciden` · ✅ 1 `<mark>` |
| 6 | vacío que **nombra** lo buscado, dice dónde buscó, ofrece salida | ✅ los 3 · ✅ **nunca «0 conversaciones»** |
| 7 | franja `Retomando la conversación… --resume <8 chars>` | ✅ `--resume cc-conv-` (**8** caracteres exactos) |
| 7 | el transcript se repinta con los turnos viejos · el detalle se abre solo | ✅ 7 turnos · ✅ `aria-expanded=true` |
| 8 | exactamente **una** con `activa:true` | ✅ |

> La franja de retoma es **efímera**: hay que cazarla con un `MutationObserver` armado
> **antes** del clic. Mirar el DOM después es una carrera perdida y da un falso negativo —
> así falló la primera corrida de este mismo guion.

**Consola del navegador: 0 errores.**

---

### E2E-2 · Turno en vuelo: el 409 del servidor — ⚠ 17 ok · 1 falla

*Capturas: `e2e2-01`…`e2e2-03`.*

| # | Aserción | Resultado |
|---|---|---|
| 1 | la conversación queda en `streaming` | ✅ |
| 2 | el `＋` **disabled** … | ✅ |
| 2 | … **con `title`**, nunca apagado y mudo | ✅ `«esperá a que termine el turno en vuelo»` |
| 2 | … y el `title` ofrece la salida **`(■ para interrumpir)`** | ❌ **FALLA — ver N-19** |
| 3 | filas inactivas `aria-disabled="true"` con su `title` | ✅ |
| 3 | **el buscador sigue habilitado y filtra** | ✅ ✅ (buscar es lectura, no transición) |
| 4 | saltear la UI: `POST .../conversaciones` → **409**, no 500 | ✅ **409** |
| 4 | el 409 nombra la causa | ✅ `session is streaming — a turn is already in flight` |
| 4 | la copy que ve el operador está en español | ✅ `hay un turno en vuelo — esperá a que termine`, **en el bundle servido** |
| 5 | tras el 409 **nada cambió** (misma activa, misma cantidad) | ✅ 2 → 2 · misma activa (**la transición es atómica**) |
| 6 | con permiso pendiente: 409 + el `＋` dice **«esperá tu decisión de permiso»** | ✅ **409** · ✅ el `title` correcto |
| 6 | *(declarado)* el 409 del servidor **no** distingue `await` de `streaming` | ⚠ reusa `ErrBusy` — ver N-20 |

#### ❌ N-19 · el `＋` apagado no ofrece la salida que el spec le manda ofrecer

- **spec.md:347 y E-07** fijan el literal: `esperá a que termine el turno (■ para interrumpir)`.
- **El código dice** `esperá a que termine el turno en vuelo`
  (`web/src/widgets/chat-dock/ui/conversacion-row.tsx:19`).

El paréntesis **no es adorno**: es la mitad que convierte «apagado» en «apagado con salida».
Tal como está, el operador con un turno largo ve el `＋` muerto y **no se entera de que `■`
existe**. Es una línea de copy. **No se arregló acá a propósito**: el gate 2 está *autorizado
por directiva, no por lectura*, y cambiar el copy antes de que el operador lea el spec sería
decidir por él cuál de los dos textos gana.

#### ⚠ N-20 · el 409 no distingue el permiso del turno (servidor)

`session_conversaciones.go:262` devuelve `ErrBusy` para `streaming` **y** para `await`, así
que el cuerpo del 409 habla de un turno en vuelo aun cuando lo que bloquea es un permiso.
**Impacto real: bajo** — la UI ya diferencia bien (`motivoBloqueo`, verificado arriba). Queda
declarado porque E-08 nombra otro motivo.

---

### E2E-3 · La rotación llega EN VIVO — ⚠ 21 ok · 1 falla · **cierra C-6/H-8**

*Capturas: `e2e3-01`…`e2e3-04`. Daemon con `--rotacion-umbral 5`.*

> **Sobre `claude` real:** el plan pedía un modelo real. **No hace falta, y por eso no se usó.**
> La rotación la dispara el `ctx_pct`, que el daemon calcula del `usage` reportado
> (`conductor.go:773`); un mock que reporta tokens ejercita **exactamente** el mismo
> `rotarLocked` + el `publish` que se quiere probar, y encima lo hace determinista. Un modelo
> real acá agrega varianza, no cobertura.

| # | Aserción | Resultado |
|---|---|---|
| 2 | superado el umbral, `rotacion_pendiente` = true | ✅ `ctx_pct=15 · pendiente=true` |
| 2 | **C-3 · la BARRA se pinta caliente** | ✅ `class="… bg-warn"` |
| 2 | **C-3 · el NÚMERO no se pinta caliente** | ✅ número en `rgba(255,255,255,.92)`, barra en `rgb(224,144,79)` |
| 3 | la marca `— contexto rotado, seguimos —` aparece | ✅ |
| 3 | **…SIN RECARGAR, por SSE** | ✅ **5 mutaciones del DOM observadas en vivo** |
| 3 | centrada · punteada · sin cola de burbuja | ✅ `align-self:center` · ✅ `dashed` · ✅ radios inferiores iguales |
| 4 | la lista sigue mostrando **N**, no N+1 | ✅ 1 → 1 (**la rotación es invisible**, CV-D10) |
| 4 | el título no cambió · el ctx bajó | ✅ · ✅ **15% → 2%** |
| 5 | el `cc-id` es NUEVO · el detalle se abrió solo | ✅ `cc-mock-0001 → cc-mock-0002` · ✅ |
| 6 | con el dock colapsado: **no se abre solo** · **cero toasts** | ✅ · ✅ |
| 6 | al reabrir, **todas** las marcas en su lugar | ❌ **FALLA — ver N-21** |
| 7 | **cero** conversaciones nuevas · `cadena_cc` creció | ✅ total=1 · ✅ `["4b046945…","cc-mock-0001","cc-mock-0002"]` |
| 7 | una marca por rotación en el transcript persistido | ✅ 3 |

**🟢 Lo que este guion cierra:** `plan-pruebas.md` lo declaraba *«hoy NO realizable»* porque
`session_rotacion.go` no tenía un solo `s.publish`. **Ahora sí llega en vivo**, medido en el
DOM sin recargar. **C-6/H-8 quedan cerrados contra la app real.**

#### ❌ N-21 · la marca de rotación se PIERDE si la copia local del transcript divergió

**Repro:** dock abierto → mandar un turno **desde otro cliente** (`POST /turn`, o una segunda
ventana) → colapsar el dock → ocurre una rotación → reabrir.
**Observado:** pantalla **1** marca, servidor **3**. Un `reload` completo muestra las 3.

**Mecanismo.** El frame de rotación viaja con `TurnoIdx` y el FE **appendea sólo si su copia
del transcript mide exactamente esa longitud** (`session_rotacion.go:81-84`) — un guard de
idempotencia pensado para el replay por `Last-Event-ID`. Pero el turno mandado desde otro
cliente **no se agrega a la copia local** (sólo el que lo envía lo agrega, optimista), así que
la longitud local queda corta **para siempre** y **todas** las marcas siguientes se descartan
en silencio. Se ve en el volcado del transcript en vivo: aparece `recibido.` (la respuesta)
pero **no** el turno del usuario que lo provocó.

**Por qué importa:** es exactamente el escenario de dos ventanas de E2E-4, y el síntoma es
*pérdida silenciosa* — la clase de fallo que el boundary `sesion-viva-consistente`
(`sin-perdida-silenciosa`) existe para prohibir. **No se arregló acá**: tocar el guard de
idempotencia no es un arreglo chico, y hacerlo a las apuradas para dejar el informe lindo es
justo lo que no corresponde. **Va al BACKLOG con este repro.**

---

### E2E-4 · Dos vistas sobre la misma sesión — ✅ 8 ok · 1 n/c

*Capturas: `e2e4-01`…`e2e4-03`. Dos contextos de navegador independientes.*

| # | Aserción | Resultado |
|---|---|---|
| 1 | las dos vistas muestran la misma sesión | ✅ |
| 2 | crear en **A** → la lista de **B** se actualiza **sola**, sin recargar | ✅ B: 1 → 2 (frame `conversacion`) |
| 3 | renombrar en **B** → el título cambia en **A** | ✅ en el servidor **y** en la pantalla de A |
| 4 | **B reconecta sola tras un corte y recupera lo perdido** | ⏸ **n/c — no ejercitable con este driver** |
| 4 | el servidor **replaya** por `Last-Event-ID` | ✅ `Last-Event-ID: 6` → replaya `7, 8, 9` |
| 5 | dos turnos simultáneos: uno **409**, uno aceptado | ✅ `[202, 409]` |
| 5 | el transcript no se intercaló ni se perdió | ✅ |

#### ⏸ Por qué el paso 4 se declara y no se finge

`BrowserContext.setOffline(true)` de Playwright bloquea peticiones **nuevas** pero **no
derriba un stream SSE ya establecido**. Medido: el `readyState` del `EventSource` se queda en
**1 (OPEN)** durante todo el corte y después. Sin evento `error` no hay reconexión, y sin
reconexión no hay `Last-Event-ID` que mandar. Lo que se observaría (B sin el turno) mediría
**la emulación, no el producto** — sería una falla fabricada. **El mecanismo del servidor sí
se probó, a nivel HTTP.** La mitad del navegador queda para el **Modo B / gate humano**
(desconectar el wifi de verdad).

---

### E2E-5 · Los fallos reales del entorno — ✅ 13/13

*Capturas: `e2e5-01`…`e2e5-03`. El arnés movido vive **dentro del sandbox**.*

| # | Aserción | Resultado |
|---|---|---|
| 2 | sin el arnés en disco, **la lista sigue funcionando** | ✅ 2 de 2 (BR-CV-2: cuelgan de `session_id`) |
| 2 | …y **el buscador también** · el endpoint responde 200 | ✅ · ✅ |
| 3 | mandar un turno **falla**, no se acepta en silencio | ✅ HTTP 500 |
| 3 | el fallo **nombra la ruta** | ✅ `session service: spawn s25123a2c: claudecode: start …` |
| 3 | la conversación vuelve a **`idle`**, no queda colgada en `streaming` | ✅ `idle` |
| 4 | restaurada la carpeta, el turno anda | ✅ HTTP 202 → `idle` |
| 5 | desvincular el arnés del Portafolio **no** se lleva la sesión | ✅ la sesión sigue (HTTP 200) |
| 6 | **las N conversaciones siguen ahí — nada en cascada** | ✅ 2 → 2 |

> Nota menor: el turno sin arnés devuelve **500**, no un 4xx. Nombra la causa y deja la
> conversación en `idle`, así que el criterio del guion se cumple; el código es más brusco de
> lo que el caso amerita (es una condición del entorno, no un fallo del servidor).

---

### E2E-6 · `--resume` que el CLI ya no reconoce — ⚠ 8 ok · 1 falla · 1 n/c

*Captura: `e2e6-01`. El `cc-id` se falseó con el daemon detenido; el mock **muere antes del
`init`** al recibir ese `--resume`, que es lo que hace el CLI real con un corpus GC'd.*

| # | Aserción | Resultado |
|---|---|---|
| 2 | `tryHealResume` respawnea **fresh** una vez | ✅ `cc-gc-eliminado-9999 → cc-mock-0001` |
| 2 | …y **reenvía el turno en vuelo** | ✅ el turno está en el transcript (4 → 7) |
| 3 | los turnos viejos **siguen** (vienen de `Conv`, no del proceso) | ✅ |
| 4 | marca `⟳ hilo reiniciado · checkpoint` — el cambio **no es silencioso** | ❌ **FALLA — ver N-22** |
| 5 | el detalle muestra el `cc-id` **nuevo** y ya no el muerto | ✅ `◍ cc-mock- · vitalia · claude-mock · cwd …` |
| 6 | un **segundo** fallo da `error` visible y no un loop | ⏸ n/c — ver abajo |

#### ❌ N-22 · RF-348 CA-1 no está construido

El literal **`hilo reiniciado` tiene 0 ocurrencias en todo el árbol** (Go, TSX, stories). Y no
es un olvido de copy: `tryHealResume` está **documentado como silencioso** —
*«The heal is silent on success»* (`session_service.go:755`)—. El `cc-id` cambia por debajo y
**la única forma de enterarse es abrir el detalle de identidad y acordarse del anterior**.

`spec.md` RF-348 CA-1 pide lo contrario: que el cambio de `cc-id` **no** sea silencioso.
**Es trabajo sin hacer, no un bug**: necesita su ticket. **No se improvisó acá.**

#### ⏸ El paso 6, declarado

El heal es de **una sola vez por proceso** (`resumeRetried`, `session_service.go:760`). Forzar
un segundo fallo exige que el proceso **fresco** también muera — que es **otro modo de fallo**
(spawn roto, cubierto por E2E-5 paso 3), no el de este escenario. Queda para el gate humano.

---

## 3 · Lo que el operador va a ver cuando instale — la respuesta corta

**Sí, va a ver lo construido.** Con tres asteriscos, en orden de importancia:

1. **Sus 5 sesiones y sus 90 turnos sobreviven la migración**, y el archivo viejo queda
   intacto: si algo sale mal, volver atrás es cambiar el binario y nada más.
2. **3 de sus 5 sesiones siguen con la llave a medias hasta que corra un comando a mano**
   (`arnesia sesiones recalibrar-llaves --aplicar`). El arranque **no lo hace ni lo avisa**.
3. **Tres cosas del spec no están como el spec dice**: la salida `(■ para interrumpir)`
   (N-19), la marca de hilo reiniciado (N-22) y la pérdida silenciosa de la marca de rotación
   con dos vistas (N-21).

Y el cromo del dibujo está en la app real: **2 filas**, chip de contexto con el `cwd`, lista
en sitio con su buscador, `»` para colapsar.

## 4 · Lo que este informe NO puede afirmar

- **El Modo B** (la ventana Tauri del `.deb` instalado) **no se corrió.** Exige `make installer`
  → `bump-patch`, prohibido en este encargo. Lo que sí se probó es el **mismo binario** que esa
  ventana ejecuta (`Makefile:27-32`).
- **La reconexión del navegador tras un corte real de red** (E2E-4 paso 4).
- **El segundo fallo consecutivo del `--resume`** (E2E-6 paso 6).
- **Nada se corrió con `claude` real.** Para E2E-3 se justifica arriba; para E2E-6 el modo de
  fallo (proceso que muere antes del `init`) se reprodujo fielmente.
- **T32 (el borrado de `sesiones-cerradas.json`, CV-D6) NO se ejecutó** — es procedimiento
  manual del operador. Los 3 ids siguen intactos en `~/.arnesia/sesiones-cerradas.json`.

## 5 · Hallazgos nuevos, para el BACKLOG

| # | Qué | Dónde | Gravedad |
|---|---|---|---|
| **N-18** | el dry-run de CV-D16 contesta «no hay sesiones» si corre antes del daemon | `cmd/arnesia/sesiones.go:131-142` | media — engaña justo al que quiere previsualizar |
| **N-19** | el `＋` bloqueado no ofrece `(■ para interrumpir)` como manda E-07 | `conversacion-row.tsx:19` | baja — 1 línea de copy |
| **N-20** | el 409 del servidor no distingue `await` de `streaming` | `session_conversaciones.go:262` | baja — la UI ya diferencia |
| **N-21** | **la marca de rotación se pierde en silencio si la copia local divergió** | `session_rotacion.go:81-84` + store del FE | **alta — pérdida silenciosa** |
| **N-22** | RF-348 CA-1 (`⟳ hilo reiniciado`) no está construido | `session_service.go:755` | media — cambio de `cc-id` invisible |
| **N-23** | el re-key de CV-D16 no corre solo ni avisa tras migrar | `main.go:202` (`clave = nil`) | media — sesiones invisibles hasta correr el comando |
