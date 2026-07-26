# Auditoría independiente — Tramo A (backend de `telemetria/`)

> **Veredicto:** el módulo está bien construido en su esqueleto y en su privacidad —la allowlist
> aguanta las dos puertas, probado con canarios— pero **la cifra de dinero que el producto existe
> para mostrar sale mal por dos caminos distintos y simultáneos** (se cuenta dos veces, y un
> agregado se presenta como completo cuando no lo es), hay **truncado silencioso a 500 turnos** que
> hace que dos pantallas del mismo dato no coincidan, y **el rollup entero es peso muerto: se
> escribe y no lo lee ninguna consulta**. No es firmable como está.

**Auditor:** agente independiente · **fecha:** 2026-07-26 · **rango:** `b058396..HEAD`
**Método:** todo se corrió con `HOME` efímero (`mktemp -d`) y puerto efímero. El
`~/.arnesia` del operador **no se tocó** (verificado: `~/.arnesia/telemetria.db` no existe).
**No se modificó una sola línea de código de producción.**

| gravedad | confirmados | sospechas |
|---|---|---|
| 🔴 crítico | 4 | 0 |
| 🟠 alto | 7 | 1 |
| 🟡 medio | 10 | 1 |
| ⚪ bajo | 7 | 0 |

---

## Entorno de reproducción

```bash
SP=<scratchpad>;  go build -o $SP/arnesia ./cmd/arnesia
H2=$(mktemp -d);  P2=<puerto efímero>
env -i PATH=/usr/bin:/bin HOME=$H2 XDG_CONFIG_HOME=$H2/.config \
  $SP/arnesia serve --addr 127.0.0.1:$P2 --log $SP/d2.log &
```

Payloads: `verificacion-2026-07-26/evidencia/*.json` (los originales, y copias con canarios
únicos sustituyendo los cinco campos de identidad).

`go test ./... -race -count=1` → **exit 0**, 26 paquetes `ok`. Los defectos de abajo **no** los
caza la suite; están reproducidos contra el binario corriendo.

---

# CONFIRMADO

## 🔴 C1 · El dinero se cuenta DOS VECES cuando están los dos exportadores — y los dos los enciende el propio daemon

**Qué está mal.** `claude_code.cost.usage` (canal `/v1/metrics`) y `api_request.cost_usd_micros`
(canal `/v1/logs`) son **el mismo gasto de la misma llamada**. Los dos se mapean a
`TipoEvento = api_request` con `CostoReportadoMicros`, y `SUM(costo_reportado_micros)` los suma a
los dos. El dedupe no los cruza porque la llave única incluye `turno_id` (el punto de métrica no
trae `prompt.id`) y `ts_emisor` (difieren en ~490 ms).

Y no es un caso de laboratorio: **`SpawnEnv` pone `OTEL_LOGS_EXPORTER=otlp` *y*
`OTEL_METRICS_EXPORTER=otlp`** en todo spawn S1, y el bloque `env` recomendado para S2 hace lo
mismo. La configuración que el módulo prescribe es exactamente la que produce el doble conteo.

**Cómo lo reproduje.** Base vacía; primero solo logs, después el `/v1/metrics` de **la misma
corrida**:

```
### PASO 1 — solo /v1/logs (run1)
costo_reportado_micros = 18473 | turnos = 1 | sesiones = 1
### PASO 2 — MISMA corrida, ahora también /v1/metrics (run1)
costo_reportado_micros = 36946 | costo_calculado_micros = 1959 | turnos = 1 | sesiones = 1
```

Los dos valores de origen, extraídos de la evidencia:

```
metrics-run1.json : claude_code.cost.usage = 0.018472600000000002 USD  (sess 89c8bdde)
logs-run1.json    : api_request.cost_usd_micros = 18473               (sess 89c8bdde)
```

Y en la tabla, las dos filas convivientes:

```
id | emisor | tipo_evento | turno_id  | costo_reportado_micros
 1 | otlp   | api_request | 8e957894… | 18473
 2 | otlp   | api_request | NULL      | 18473
```

**Qué debería pasar.** O el canal secundario no aporta dinero cuando el primario ya lo trajo para
esa sesión/ventana (D14.1 ya declara que *el canal primario es `/v1/logs`* — pero solo lo declara,
no lo implementa), o la llave de dedupe se hace insensible al `turno_id` ausente y al `ts_emisor`
para los eventos de costo.

**Dónde.**
- `internal/adapters/telemetria/otlp/mapa_cc.go:310-313` — `claude_code.cost.usage` → `CostoReportadoMicros`.
- `internal/adapters/telemetria/otlp/mapa_cc.go:273` — `TipoEvento: domain.EventoAPIRequest`.
- `internal/adapters/telemetria/store/consultas.go:80` — `SUM(costo_reportado_micros)`.
- `internal/adapters/agent/claudecode/conductor.go:178-179` — los dos exportadores.
- `internal/adapters/telemetria/store/migracion.go:101-102` — la llave de dedupe.

**Gravedad.** Crítica. Es **la** cifra del producto («este arnés quema $X»), y todo lo que cuelga
de ella: `costo_por_corrida`, `GastoCaja.Parte`, `Ventana.TotalMicros` de los seis detectores,
`ParidadCosto`, `DivergenciaPct`. Un tablero de costos que sobre-reporta 2× es peor que no tener
tablero: dispara decisiones de gasto sobre un número inventado hacia arriba.

**Sobre el ítem que el constructor dejó declarado abierto** («el dedupe colapsa puntos de
`/v1/metrics`»): el impacto real es **peor** que el declarado. No es solo que los 4 datapoints de
`claude_code.token.usage` colapsen entre sí (`eventos_duplicados: 4` por lote) — es que el que
sobrevive es el de **costo**, que llega sin tokens, y duplica el dinero. Además, al colapsar los
de token, **todo el dato de tokens del canal de métricas se pierde sin que nada lo diga** (ver A6).

---

## 🔴 C2 · Cero de dinero fabricado: `costo_calculado_micros: 0` con `costo_completo: true`

**Qué está mal.** `CalcularCosto` con **cero tokens en todos los buckets** devuelve
`{Micros: 0, Completo: true, SinNingunaTarifa: false}` — porque `SinNingunaTarifa` exige
`cobrados == 0 **&& len(sinTarifa) > 0**`, y sin tokens no hay nada que nombrar en `sinTarifa`.
`costear()` entonces guarda el puntero a **0** y marca **completo**. Es exactamente el «no aplica
convertido en 0» que el boundary prohíbe, en el campo de dinero.

**Cómo lo reproduje.** El evento de métrica del paso 2 de C1, en la base y en el wire:

```
-- tabla evento
id | modelo                    | tok_entrada | costo_reportado_micros | costo_calculado_micros | costo_completo
 2 | claude-haiku-4-5-20251001 | NULL        | 18473                  | 0                      | 1
```

```json
// GET /api/telemetria/arneses/vitalia/cajas/paso-3  → turnos[1]
{ "turno_id": "", "modelo": "claude-haiku-4-5-20251001",
  "tokens": {},
  "costo_reportado_micros": 18473,
  "costo_calculado_micros": 0 }
```

Un turno que costó $0,018473 reportados se muestra como **$0,000000 calculado**, sin ninguna marca.

**Qué debería pasar.** Sin ningún bucket con tokens no hay nada que cotizar: `CostoCalculadoMicros`
debe viajar **nil**, no 0, y `CostoCompleto` debe quedar nil (no `true`).

**Dónde.** `internal/domain/telemetria_costo.go:120-127` (la condición de `SinNingunaTarifa`) ·
`internal/usecase/telemetria_service.go:184-192`.

**Gravedad.** Crítica. Es el primo hermano exacto del defecto de tier que ya se corrigió, en la
misma función. Y el 0 no queda encerrado: entra al `SUM(costo_calculado_micros)` del resumen y al
`MIN(COALESCE(costo_completo,1))` **como un voto de «completo»**.

---

## 🔴 C3 · Dos endpoints, el mismo dato, veredictos de completitud contradictorios

**Qué está mal.** `ParidadCosto.Completo` se calcula como `hayRep && hayCalc` — o sea *«existen los
dos números»*, no *«el costo cotizó todos los buckets»*, que es lo que la palabra significa en
`ResumenTelemetria.CostoCompleto` y en la doc del DTO. Y `ParidadCosto.SinTarifa` **nunca se llena**
en ningún lado del árbol.

**Cómo lo reproduje.** Mismos datos, dos rutas:

```
resumen.costo_completo = False  sin_tarifa= ['cache_escritura_sin_tier']
detalle.paridad        = {'reportado_micros': 36946, 'calculado_micros': 1959,
                          'divergencia_pct': -94.69766686515455, 'completo': True}
```

La 4.ª tab del inspector —la que el mockup dibuja— declara **completo** un número que está 94 %
por debajo del reportado justamente porque **no** está completo.

**Qué debería pasar.** `ParidadCosto.Completo` tiene que ser la conjunción de los `costo_completo`
de sus partes (o directamente `MIN`), y `SinTarifa` tiene que traer el motivo, igual que el resumen.
Si el campo quiere decir «hay dos números», el nombre está mal y hay que renombrarlo.

**Dónde.** `internal/usecase/telemetria_service.go:315` ·
`internal/domain/telemetria_vistas.go:140-148`.

**Gravedad.** Crítica. Es un pass fabricado en el wire, y el que lo va a leer es el FE del Tramo B.

---

## 🔴 C4 · Truncado silencioso a 500 turnos: dos pantallas del mismo dato no coinciden

**Qué está mal.** `Store.Turnos` **no lleva `LIMIT` en el SQL**: escanea la ventana entera, agrupa
en memoria y recorta la salida a `limite` (default 500) **sin marcar que recortó**. `DetalleCaja`
suma tokens y costos **sobre esa lista ya recortada** y presenta el resultado como total de la caja.
`Mejoras` hace lo mismo, pero saca `TotalMicros` del `Resumen`, que **no** está recortado.

**Cómo lo reproduje.** 600 `api_request` de 1 000 micros cada uno, mismo arnés, misma caja:

```
=== resumen acotado al arnés masivo ===
  resumen.costo_reportado_micros = 600000   turnos = 600
=== detalle de la caja (mismo dato) ===
  detalle.paridad.reportado_micros = 500000
  detalle.turnos devueltos          = 500
  ¿avisa de truncado?               = ['caja_id','nombre','atribuible','desde','hasta','tokens',
                                       'paridad','confianza','turnos','detectores','catalogo']
```

Y el efecto sobre los detectores — numerador truncado, denominador completo:

```
b4-gasto-por-arnes-empresa-puesto  gasto_micros= 500000  parte_del_total= 0.8333
                                   corridas_usadas= 500  corridas_totales= 500
```

La caja se lleva el **100 %** del gasto. El detector dice **83,33 %**, y declara «500 corridas
totales» donde hubo 600.

**Qué debería pasar.** O el agregado se calcula en SQL (no sobre la página), o la respuesta lleva un
campo `truncado`/`total_turnos` explícito. La doctrina del paquete es que una degradación se
DECLARA; ésta no se declara en ningún campo.

**Dónde.** `internal/adapters/telemetria/store/consultas.go:358-363, 453-459` ·
`internal/usecase/telemetria_service.go:274-315` · `internal/usecase/telemetria_mejoras.go:53, 67-69`.

**Gravedad.** Crítica. Un total parcial presentado como total, con dos superficies del producto
mostrando cifras distintas del mismo dato y ninguna diciendo por qué.

---

## 🟠 A1 · `DetalleCaja.confianza` está clavada en `sin-dato`, siempre

**Qué está mal.** El acumulador arranca en `ConfianzaSinDato` y después se combina con
`PeorConfianza`, que devuelve la peor. Como `sin-dato` es el piso, **el resultado no puede ser
nunca otra cosa**.

**Cómo lo reproduje.**

```
detalle.confianza  = sin-dato  | atribucion de sus turnos = ['exacta', 'exacta']
```

(y en la misma corrida `resumen.confianza = exacta`, que sí funciona).

**Qué debería pasar.** Sembrar con `""` y tomar el primero, como hace `cobertura()`
(`consultas.go:135`).

**Dónde.** `internal/usecase/telemetria_service.go:279`.

**Gravedad.** Alta. El boundary `cifra-viaja-con-su-confianza` se cumple en la forma (el campo
viaja) y se rompe en el fondo (el valor es constante). Una insignia que siempre dice lo mismo no
informa; peor, entrena a ignorarla.

---

## 🟠 A2 · El rollup se escribe y **no lo lee ninguna consulta de producción**

**Qué está mal.** `rollup_hora` solo aparece en escrituras (`rollup.go`, `Purgar`, `PurgarRollup`)
y en tests. `Resumen`, `PorCaja`, `Turnos`, `cobertura`, `escenario` — **todas** van a la tabla
cruda. Consecuencias, en cadena:

1. La promesa documentada *«el rollup sobrevive más que el detalle: tras purgar, el drill-down de
   un turno viejo dice “detalle purgado, resumen conservado”»* (`store.go:569-571`) **es falsa**.
   La purga por TTL de 90 días borra la historia a efectos de todo lo que el producto muestra.
2. El presupuesto de performance («el tablero cuesta milisegundos en vez de cientos, 190× medido»)
   **no está realizado**: cada consulta escanea la cruda.
3. La columna `tok_cache_sin_tier` que la migración v2 agrega a `rollup_hora` **nunca se llena**:
   no está en `columnasSumables` ni en el `SELECT` del insert.

**Cómo lo reproduje.**

```
  rollup_hora del arnés masivo ANTES: [(600, 600000)]
  eventos borrados (simulando purga TTL): 600
  rollup_hora del arnés masivo DESPUES: [(600, 600000)]
### despues: el 'resumen conservado' que la doc promete
  costo_reportado_micros = None  turnos = 0  cobertura = {...todo en 0...}
```

Y el grep que lo cierra:

```
$ grep -rn "rollup_hora" --include=*.go | grep -v _test.go
# → solo INSERT / DELETE / COUNT(*) de cardinalidad. Cero SELECT de lectura.
```

**Dónde.** `internal/adapters/telemetria/store/rollup.go:91-95, 120-148` ·
`internal/adapters/telemetria/store/store.go:562-574` · `consultas.go` (entero).

**Gravedad.** Alta. Un componente completo con su cursor, su debounce, su tope de cardinalidad y su
suite de tests que no participa del producto. Y la capability CAP-123 lo declara `status: vivo` con
`api_endpoints: ["GET /api/telemetria/resumen"]` — un endpoint que no lo toca.

---

## 🟠 A3 · El bucket que explica el 94 % de la divergencia es invisible en toda la superficie de lectura

**Qué está mal.** `cache_escritura_sin_tier` —el bucket que existe para no reproducir
`phoenix#14314`, el que hace que el costo calculado sea una cota inferior— **se persiste y no se
lee nunca**:

- `Store.Turnos` no lo incluye en su `SELECT` (`consultas.go:366-370`);
- `TurnoUnido.Tokens.CacheEscrituraSinTier` queda por lo tanto siempre nil;
- `DetalleCaja` no lo suma (`telemetria_service.go:285-290`);
- el rollup no lo agrega (A2).

**Cómo lo reproduje.** La base tiene `tok_cache_sin_tier = 8257` para el evento id 1. La respuesta
de `GET /api/telemetria/arneses/vitalia/cajas/paso-3` para esa misma caja:

```json
"tokens": { "entrada": 10, "salida": 39, "cache_lectura": 17536 }
```

Sin rastro de los 8 257 tokens que son la causa de `divergencia_pct: -94.7`.

**Qué debería pasar.** El resumen nombra el motivo (`sin_tarifa: ["cache_escritura_sin_tier"]`) —
bien— pero el drill-down no puede mostrar **cuánto**. Quien dibuje el inspector no tiene con qué
justificar la cifra incompleta.

**Gravedad.** Alta. La honestidad se declara en el flag y se pierde en el dato.

---

## 🟠 A4 · El hook **cuelga sin tope** leyendo stdin — el fail-open no cubre su propia entrada

**Qué está mal.** El contrato de `hook.go` dice *«duración: tope duro de 250 ms; pasado eso abandona
y sale 0»*. El `context.WithTimeout(250ms)` **solo protege la petición HTTP**. La lectura de stdin
es `io.ReadAll(io.LimitReader(entrada, 1<<20))`: acotada en **bytes**, no en **tiempo**. Si el
escritor no cierra, el hook espera para siempre.

**Cómo lo reproduje.**

```
--- stdin lento (el escritor tarda 3 s en cerrar) ---
stdin lento: rc=0  duracion=3003ms  (tope contractual: 250ms)

--- fifo abierto sin escritor ---
fifo sin escritor: rc=124 (124=matado por timeout) duracion=8002ms — el tope es 250ms
```

En el segundo caso el hook **no terminó**: lo mató `timeout 8`. Y entonces tampoco salió 0.

Todas las demás ramas del fail-open **sí** cumplen (verificadas en la misma corrida, todas
`rc=0`, stdout 0 bytes, ≤6 ms): sin ficha · ficha corrupta · JSON roto · puerto muerto ·
subcomando desconocido. Y el daemon lento sí se corta a 256 ms.

**Por qué no lo caza la suite.** `TestHookNoTardaNiFalla` (`cmd/arnesia/telemetria_test.go`) alimenta
el hook con `strings.NewReader(...)`, que devuelve EOF de inmediato: **estructuralmente no puede
observar un cuelgue**, aunque su comentario diga *«lo que se prueba es que no cuelga»*. Y el techo
que asserta es 3 s, no 250 ms.

**Dónde.** `cmd/arnesia/hook.go:59-66`.

**Gravedad.** Alta. Un `UserPromptSubmit` colgado bloquea el turno del usuario — la única cosa que
este módulo tiene prohibido hacer. Que en la práctica el runtime cierre stdin rápido lo hace
latente, no inexistente.

---

## 🟠 A5 · El guardarraíl de tamaño mide el `.db` e ignora el `-wal`

**Qué está mal.** `os.Stat(s.ruta)` mira el archivo principal. Bajo WAL, los datos recién escritos
viven en `telemetria.db-wal` hasta el checkpoint. El aviso de 500 MB nunca se enciende a tiempo.

**Cómo lo reproduje.**

```
telemetria.db      4096 bytes
telemetria.db-shm  32768 bytes
telemetria.db-wal  1161872 bytes
salud.tamano_bytes reportado = 4096   aviso_tamano = False
```

El reporte ve el **0,34 %** de la huella real en disco.

**Dónde.** `internal/adapters/telemetria/store/store.go:529-534`.

**Gravedad.** Alta. Combinado con M3 (sin rate limit en la ingesta sin token), es el camino a llenar
el disco del usuario sin que la salud lo diga.

---

## 🟠 A6 · Contadores de pérdida que se persisten y **nunca se muestran**

**Qué está mal.** El store escribe `eventos_duplicados` y `errores_escritura` en la tabla `salud`,
pero `Store.Salud()` no los mapea y **`domain.SaludTelemetria` no tiene campo para ellos**. Ni la
API ni el CLI los exponen.

**Cómo lo reproduje.**

```
  salud persistida: {'recibidos': 618, 'atributos_fuera_de_lista': 1461,
                     'aceptados': 606, 'eventos_duplicados': 12}
```

```json
// GET /api/telemetria/salud  —  y arnesia telemetria salud da lo mismo
{"recibidos":618,"aceptados":606,"descartados_cola_llena":0,"rechazados_formato":0,
 "rechazados_tamano":0,"atributos_fuera_de_lista":1461,"temporalidad_no_soportada":0, ...}
```

`618 − 606 = 12` sin ninguna línea que lo explique. Y esos 12 son justamente el mecanismo por el que
**se pierden todos los datapoints de token de `/v1/metrics`** (C1).

**Dónde.** `internal/adapters/telemetria/store/store.go:45-46, 501-516` ·
`internal/domain/telemetria_vistas.go:218-242`.

**Gravedad.** Alta. «Cuántos descarté» es dato de honestidad por doctrina explícita del propio
archivo (`store.go:160-161`), y es el único dato que no sale.

---

## 🟠 A7 · La re-validación de A13 no está cableada: `postProceso(telemetria, nil)`

**Qué está mal.** El handler documenta *«el daemon lo re-valida igual (A13: la allowlist se aplica
dos veces, y no se confía en que el emisor haya filtrado aunque el emisor sea nuestro propio
binario)»*. El router lo monta con `revalidar = nil`.

**Cómo lo reproduje.** Con el token de ingesta (legible por cualquier proceso del mismo usuario en
la ficha `0600`), inyecté texto arbitrario en campos libres:

```bash
curl -X POST -H "X-Arnesia-Token: $TOK" -d '{"sesion_id":"s-canario2", ...,
  "motivo":"CANARIOMOTIVO-inyectado","gate":"CANARIOGATE"}' .../api/telemetria/proceso
→ 202
```

y quedaron en disco:

```
telemetria.db-wal → CANARIOGATECANARIOMOTIVO
```

**Qué debería pasar.** Lo que dice el comentario: una segunda pasada de proyección/allowlist antes
de `Ingerir`. Hoy la segunda barrera es una intención.

**Dónde.** `internal/adapters/transport/http/router.go` (la línea
`mux.HandleFunc("POST /api/telemetria/proceso", postProceso(telemetria, nil))`) ·
`internal/adapters/transport/http/telemetria.go:188-213`.

**Gravedad.** Alta como control, media como riesgo: el struct canónico no tiene campos de identidad
ni de contenido, así que lo inyectable es texto libre en `motivo`/`gate`/`herramienta` — que además
**egresa por el forward** si el operador lo enciende.

---

## 🟡 M1 · B6 dispara siempre en `s2-instrumentado`: reclama el 100 % del gasto como «sesiones abandonadas»

`detB6.Aplica` solo pide `TieneCosto`. Fuera de ArnesIA sin hook **nunca** hay señal de proceso, así
que ninguna sesión «cierra» y todas quedan abandonadas.

```
b6-sesion-abandonada  gasto_micros= 500000  parte_del_total= 0.8333
                      corridas_usadas= 1  corridas_totales= 1  confianza= exacta
```

Es el 100 % del gasto de un arnés perfectamente sano marcado como desperdicio. P1 sí exige
`TieneSenalProceso`; B6 debería exigirlo igual (o declararse `Parcial` con motivo, como P1).

⚠️ **`TestS2InstrumentadoTieneDineroYNoTieneSplit` (fitness, línea 505-509) cementa el
comportamiento**: asserta que B6 *tiene que aplicar* con dinero y sin proceso. No es un descuido del
test — es una decisión que el test protege, y creo que es la decisión equivocada.

`internal/domain/telemetria_deteccion.go:323-373`.

---

## 🟡 M2 · B1 y B3 ponen **conteos de tokens** en campos que se llaman `micros` de dólar

- **B1:** `hoy := int64(float64(escrito5m) * 1.25)` — `escrito5m` son tokens y `1.25` es un
  multiplicador **relativo al precio de entrada**, no una tarifa. El resultado va a `GastoMicros`.
- **B3:** `micros += *t.Tokens.CacheEscritura5m` — suma tokens directamente a `GastoMicros`.

En los dos, `ParteDelTotal = parteDe(gasto, v.TotalMicros)` divide **tokens sobre micros**: una
fracción sin significado, que puede pasar de 1. B3 lo insinúa en su `Sesgo` («se cuentan tokens de
re-escritura, no su costo exacto por modelo»), pero el campo se llama `gasto_micros` y la UI lo va a
formatear como dinero.

`internal/domain/telemetria_deteccion.go:392-431` (B3) · `:459-502` (B1).

---

## 🟡 M3 · No existe el rate limit que `auth.go` cita como tercera barrera

`auth.go:191` justifica que `/v1/logs` y `/v1/metrics` acepten **sin token**: *«Quedan tres
barreras: Host gate, tope de cuerpo y rate limit»*. Y `auth.go:32` lo repite.

```
$ grep -rln "ratelimit\|RateLimit\|rateLimit\|limitador" --include=*.go .
(sin resultados)
```

Es un control declarado que no existe. Con A5 (el aviso de tamaño ciego al WAL) queda un endpoint
local sin autenticar, sin límite de tasa, cuya única señal de crecimiento está rota.

---

## 🟡 M4 · La allowlist declarada **no es** el conjunto de lo que se lee: falta `type`

`MapearPuntoMetrica` lee `p.Attrs.Texto("type")` para desagregar `claude_code.token.usage`, y
`type` **no está en `allowlistOTLP`**.

```
  leídos y NO declarados: ['type']
```

Efectos: (a) cada datapoint de token suma +1 a `atributos_fuera_de_lista` por un atributo que sí
usamos, ensuciando el contador de salud; (b) se rompe el invariante que el propio archivo enuncia
(*«se usa … para que el test de source-scan pueda contrastar la lista contra lo que el mapeo
realmente lee»*, `mapa_cc.go:36-37`) — y ningún test lo contrasta.

`internal/adapters/telemetria/otlp/mapa_cc.go:38-57, 297`.

---

## 🟡 M5 · `recibidos` no es «recibidos»

`r.contar(&r.salud.Recibidos, "recibidos", int64(len(evs)))` cuenta los eventos **mapeados con
éxito**, después del default-deny. Quedan fuera: los `event.name` descartados a propósito
(`assistant_response`, `user_prompt`, `hook_execution_*`), los sin `session.id`, y **todo el camino
del hook** (`/api/telemetria/proceso` no pasa por el receptor). `aceptados`, en cambio, cuenta
inserciones reales de los cuatro emisores. Son poblaciones distintas presentadas como el numerador y
el denominador de la misma barra.

`internal/adapters/telemetria/otlp/receptor.go:187`.

---

## 🟡 M6 · `Resumen.Confianza` la fija cualquier fila de la ventana, aporte o no al número

`cobertura()` toma la peor confianza de **todas** las filas, incluidas las `sin-dato`, que por A15
están excluidas de los totales de dinero. En la práctica basta un evento no atribuido en 30 días
para que la insignia quede en `sin-dato` de forma permanente.

Reproducido: con 3 eventos `exacta` (todo el dinero) + 1 evento de hook `sin-dato` (sin dinero),
`confianza = "sin-dato"`; borrando el evento de hook, `confianza = "exacta"`.

`internal/adapters/telemetria/store/consultas.go:135-163`.

---

## 🟡 M7 · T19 declarado cerrado sin la mitad de su entregable, y el cierre honesto no lo menciona

El plan pide para T19 «las 8 rutas **+ contrato OpenAPI**: `openapi.yaml` sube a
**`0.8.0-telemetria`** con las 9 rutas y los schemas `ResumenTelemetria` · `Cobertura` ·
`GastoCaja` · `DetalleCaja` · `PuntoDeMejora` · `RespuestaMejoras` · `FilaPortafolio` ·
`SaludTelemetria` · `EventoProceso`».

```
$ grep -n "telemetria" docs/architecture/contracts/api/openapi.yaml
(sin resultados)
```

Cero rutas, cero schemas, sin bump de versión. El `INDEX.md` («Retomar aquí», tabla *«Lo que NO se
construyó, y por qué»*) lista T22, T27 y el Tramo B — **no lista esto**, y declara T19 dentro de
«T1-T21 construidos y verdes». El documento de estado honesto tiene un hueco.

---

## 🟡 M8 · `TestPresupuestoDeBinario` no verifica el presupuesto que dicen que verifica

`arquitectura-modulo.md:1758` afirma: *«compila el daemon y compara contra el baseline del release
anterior: delta ≤ 1,5 MB»*. El test **solo** compara contra el techo absoluto de 25 MB; su propio
comentario lo admite («se verifica el techo absoluto y se declara el hueco»). El presupuesto
principal del boundary `peso-del-binario-es-presupuesto` no está enforced en ningún lado.

Lo medí yo, porque nadie lo mide: mismo árbol (`git archive HEAD`), mismo toolchain, misma
invocación; la única diferencia es que `cmd/arnesia` deja de importar `internal/adapters/telemetria/*`:

```
arnesia-head    19,41 MB
arnesia-notel   18,95 MB
delta atribuible al módulo telemetría = 0.45 MB   (presupuesto +1,5 MB)  ✔ DENTRO
```

**El módulo está dentro de presupuesto.** El defecto es que eso es un accidente feliz, no un hecho
verificado en CI.

Un matiz que sí importa para el test: **su resultado depende de si `web/dist` está construido.**
Árbol limpio → 19,41 MB; árbol de trabajo con la SPA embebida → **23,38 MB**, o sea el 93,5 % del
techo de 25 MB. El mismo test da dos respuestas muy distintas según el estado del working tree.

*(Nota metodológica: mi primera medición fue contra `b058396` y daba +5,63 MB. Es falsa — ese
commit no tenía versionados `marketplace`, `stt/local` ni `traer`, y mi build local llevaba la SPA
embebida. La corregí antes de escribirla. La regla del control positivo aplica también al auditor.)*

---

## 🟡 M9 · El oráculo de divergencia está saturado: dispara siempre

`UmbralDivergenciaPct = 5.0` y `divergencia_sospechosa` se enciende arriba de eso. En el camino
normal —OTLP, que nunca dice el tier del cache write— la divergencia medida es:

```
resumen  : divergencia_pct = -88.59   divergencia_sospechosa = true
detalle  : divergencia_pct = -94.70
```

Una alarma que suena en el 100 % de las corridas no distingue «el catálogo está viejo» de «así
funciona el canal». El oráculo cazó el defecto del tier una vez y desde entonces está pegado en
rojo; hay que separar la divergencia **estructural conocida** (bucket sin tier) de la
**inesperada**, o el próximo error de costeo va a llegar tapado por el ruido.

---

## 🟡 M10 · `Tendencia` existe en el contrato y nunca se calcula

`FilaPortafolio.Tendencia string json:"tendencia,omitempty"` con los valores documentados
`"alza" | "estable" | "baja"`. No hay una sola asignación en todo el árbol:

```
$ grep -rn "Tendencia" --include=*.go internal/ cmd/ | grep -v _test
internal/domain/telemetria_vistas.go:209:  Tendencia string `json:"tendencia,omitempty"`
```

Un campo del contrato que nunca llega. Quien dibuje la columna de tendencia va a tener que
inventarla o dejarla vacía sin saber si es «no hay datos» o «no está construido».

---

## ⚪ Bajos (confirmados, sin reproducción extensa)

| # | qué | dónde |
|---|---|---|
| B1 | `Store.Turnos` carga **toda** la ventana en memoria y recorta después; sin `LIMIT` en SQL, sin cursor, sin conteo total. Con retención de 90 días eso puede ser cientos de miles de filas por consulta. | `consultas.go:358-461` |
| B2 | `ts_recibido` se compara **lexicográficamente** en formato RFC3339Nano, que omite la fracción cuando es 0. Un `--desde 2026-07-26T00:00:00Z` (sin fracción) excluye los eventos de `00:00:00.5` porque `'.'(0x2E) < 'Z'(0x5A)`. Pierde hasta 1 s de datos en el borde de la ventana. | `consultas.go:41-48` |
| B3 | `leerTodo` devuelve `buf, nil` para cualquier error que no sea EOF ni «too large»: un cuerpo truncado por reset de conexión se acepta como completo y se intenta decodificar. | `receptor.go:236-244` |
| B4 | `Store.Ingerir` descarta con `continue` los eventos sin `sesion_id` **sin contarlos** en ningún contador de salud; el caller solo ve un `n` menor. | `store.go:200-206` |
| B5 | `aceptados` es monótono creciente y no se ajusta al purgar: tras borrar 600 filas la salud sigue diciendo `aceptados: 606` con 6 eventos en la tabla. | `store.go:377` |
| B6 | En la misma respuesta, `turnos: 2` (distinct `turno_id`, excluye `sin-dato`) conviven con `cobertura` sumando 4 (distinct con clave sintética `sin-turno:<id>`, incluye `sin-dato`). Dos denominadores distintos, nada que lo diga. | `consultas.go:81, 129` |
| B7 | `Rollup.Recomputar` borra las horas en una transacción y las reinserta en **otra**: entre las dos, un lector ve el agregado vacío. (Hoy es inocuo porque nadie lee el rollup — ver A2.) | `rollup.go:231-271` |

---

## Auditoría de tests: la regla del control positivo (§14)

**El paquete cumple la regla mejor que el promedio del repo.** Revisé los 15 tests de
`docs/architecture/fitness/telemetria_test.go` y los colocados de los 8 paquetes del módulo. Los
controles positivos están puestos donde importan y son de los buenos —marcadores distintos en la
misma corrida, no un «y además pasó algo»:

- `TestHookFailOpenSinDaemon` — `MARCADOR-A-SIN-DAEMON` (debe ser 0) vs `MARCADOR-B-CON-DAEMON`
  (debe ser 1). Es el arquetipo, y está bien hecho.
- `TestAllowlistEsListaNoSugerencia` — le da al detector AST un archivo sintético con el patrón
  prohibido para probar que el detector no está roto. Correcto.
- `TestAllowlistNoPersistePII`, `TestHookNoReenviaContenido`, `mapa_cc_test.go`, `proyecta_test.go`,
  `ficha_test.go`, `rollup_test.go`, `catalogo_test.go` — todos con control positivo en la misma
  corrida.
- `TestCifraLlevaConfianza` y `TestAllowlistEsListaNoSugerencia` fallan explícitamente si el scan no
  revisó nada («verde por vacío»). Bien.

**Tres huecos reales:**

1. **`TestHookNoTardaNiFalla` no puede fallar por lo que dice probar.** Su comentario declara «lo
   que se prueba es que no cuelga», y alimenta el hook con `strings.NewReader`, que devuelve EOF de
   inmediato. Es un assert que no puede observar el modo de falla. Es el test que debía cazar A4 y
   no puede. Su techo, además, es 3 s, no los 250 ms del contrato.
2. **`TestNoAplicaSobreviveAlRollup` no lee el rollup.** Ingiere, llama a `roll.Actualizar` y
   después verifica con `svc.Resumen(...)` — que va a la tabla cruda. El nombre y la fila del
   boundary `no-aplica-sobrevive-al-agregado`
   (`no-aplica-no-es-cero.md:94`, marcada «pendiente — TestNoAplicaSobreviveAlRollup») prometen una
   verificación del agregado que este test no ejerce. La verificación fina sí existe, colocada en
   `store/rollup_test.go`; el problema es que el boundary apunta al test que no la hace.
3. **Ningún test contrasta la allowlist declarada contra los atributos que el mapeo realmente lee**,
   aunque `mapa_cc.go:36-37` dice que ése es su propósito. Por eso M4 (`type`) pasó.

`go test ./... -race` verde · `conformance --todo`: **311 checks · pass 80 · fail 0 · deferred 231**.
El «fail 0» es cierto; el 74 % deferred también.

---

# SOSPECHA (no reproducido)

## S1 · 🟠 La ruta cruda del proyecto se persiste como `instalacion_id` y egresa por el forward

`instalacionesDelPortafolio` construye `FilaInstalacion{InstalacionID: inst.ProyectoPath}` y
`resolverHuella` devuelve `inst.ProyectoPath` como `instalacionID`. Ese valor lo asigna
`TelemetriaService.Atribuir` a `e.InstalacionID`, que **se persiste crudo** en la columna
`evento.instalacion_id` y viaja en el JSON del forward.

Eso contradice el comentario que está tres líneas arriba de la función:

> *«El hook manda la HUELLA, no la ruta (A14) … **La ruta del usuario nunca entra al almacén, ni
> siquiera acá.**»* — `cmd/arnesia/telemetria_wiring.go:206-208`

**Por qué queda como sospecha:** para reproducirlo hace falta un Portafolio poblado con
instalaciones reales en el `HOME` de prueba, y no lo monté. El camino de código es directo y
verificable por lectura (`telemetria_wiring.go:196, 220` → `telemetria_service.go:144, 153` →
`store.go:343`), pero **no lo vi en disco**, así que no lo cuento como confirmado.

Nota: en S1 el mismo valor viaja por `OTEL_RESOURCE_ATTRIBUTES=arnesia.instalacion=<path>`, o sea
que si `AtribucionSpawn.InstalacionID` es un path, también llega por ahí. En la evidencia medida el
valor es `home-local` (un id, no una ruta), lo que sugiere que en el spawn se usa un identificador
— pero por el camino del Portafolio es el `ProyectoPath`.

## S2 · 🟡 Disco lleno / base corrupta a mitad de escritura

No lo probé: requiere un filesystem lleno controlado o corromper un `.db` en vuelo. Por lectura, el
camino parece cubierto (`escribirLote` no tumba el daemon, cuenta en `errores_escritura`, `New`
archiva en vez de borrar y `Salud().HistoriaArchivada` lo dice). El agujero conocido es que ese
contador **no se muestra** (A6): un disco lleno se traduce hoy en telemetría que desaparece sin que
la salud lo reporte.

---

# Lo que revisé y está bien

- **Privacidad del canal OTLP — probado con canarios, no leído.** Sustituí los cinco campos de
  identidad de los payloads reales por marcadores únicos (`CANARIOEMAIL@example.invalid`,
  `CANARIOUUID-1111`, `CANARIOACCT-2222`, `CANARIOUID-3333`, `CANARIOORG-4444`), los posteé a
  `/v1/logs` y `/v1/metrics`, y busqué las subcadenas en `telemetria.db`, **`telemetria.db-wal`**,
  `telemetria.db-shm`, el log del daemon, stdout y stderr. **Cero coincidencias en los cinco.**
  La allowlist default-deny aguanta, y el evento se arma campo por campo (no hay constructor por
  mapa, verificado por AST).
- **Privacidad del camino del hook.** `Proyectar` decodifica a un struct que no tiene dónde poner
  `prompt`, `last_assistant_message` ni `tool_response`. El `cwd` se hashea (`HuellaCWD`, sha256
  truncado a 16 hex, con `filepath.Clean` previo). `TestHookNoReenviaContenido` lo enforcea con
  control positivo.
- **Fail-open del hook, 5 de 6 ramas.** Sin ficha · ficha corrupta · JSON roto · puerto muerto ·
  subcomando desconocido: todas `rc=0`, stdout **0 bytes**, ≤6 ms. Daemon lento: cortado a 256 ms.
  (La sexta —stdin— es A4.)
- **El gate de ingesta.** `POST /api/telemetria/proceso` **sin token → 401**, con token → 202.
  `/v1/*` sin token → aceptado, que es A22 firmado y documentado.
- **La ficha del daemon.** `0600`, en `os.UserConfigDir()`, publicada **después** de que el listener
  acepta, con el token acotado (24 bytes aleatorios, distinto del de la API) y el endpoint real.
- **Migración v1 → v2 con datos viejos.** Fabriqué una base sintética en versión 1 con un evento de
  777 micros; tras abrirla: `schema_meta = (1, 2)`, la columna `tok_cache_sin_tier` existe, el
  evento sigue ahí (`costo_reportado_micros = 777`), y **no** se archivó nada. Correcto.
- **Dos escritores sobre el mismo `.db`.** Dos daemons + el CLI contra el mismo archivo, 6 POST
  concurrentes: **cero `SQLITE_BUSY`, cero «lote no escrito»**, 606 eventos aceptados. El patrón
  writer con `SetMaxOpenConns(1)` + `_txlock=immediate` + `busy_timeout(5000)` + WAL funciona.
- **`CalcularCosto`, los tres bugs ajenos.** El cache write se cobra; la aritmética inclusiva
  descuenta `cache_lectura` de `entrada` sin bajar de cero; los tiers no se aplanan (`SobreUmbral`);
  y el tier desconocido no se asume barato — la corrección de `3144db9` está bien hecha y bien
  atada por `TestTierDesconocidoNoSeAsumeBarato` y `TestB1NoDetectaReWarmSobreUnTierFabricado`.
- **`sumaPreservandoNULL`.** El `CASE WHEN los dos lados son NULL THEN NULL` está correcto — el
  problema del rollup es que nadie lo lee (A2), no que la aritmética esté mal.
- **`PeorConfianza`** trata la confianza desconocida como `sin-dato`: no concede beneficio de la
  duda. Correcto. (El bug de A1 es del caller, no de esta función.)
- **`escenario` se deriva de la señal** y no lo declara el emisor. `corrida_id` se escribe como
  `NULL` cuando está vacío (`textoOpcional`), así que el `IS NOT NULL` funciona.
- **`Purgar` por arnés** borra detalle, agregado y turnos esperados en **una** transacción.
- **`arnesia telemetria`** reusa el mismo caso de uso, cero lógica de agregación propia. `catalogo`
  no abre la base. La consecuencia de `portafolio` vacío sin daemon está declarada en el código.
- **El catálogo embebido.** 694 modelos, 64 KB (presupuesto 256 KB), con `rev`, `sha256` y
  `refrescado: null` cuando nunca se refrescó. El flag `--telemetria-catalogo-refresco` **avisa que
  la descarga no está construida** en vez de fingir.
- **El forward está apagado por default** y no abre socket cuando lo está.
- **`TestNoJSONLSchemaParsing`** dejó de ser `t.Skip` y enforcea (commit `929ad1b`).

---

# Lo que NO pude auditar

| qué | por qué |
|---|---|
| **La UI** | No se construyó a propósito (mockup sin firma, P0). No la audité como faltante. Sí evalué si el contrato la va a poder sostener → sección siguiente. |
| **Disco lleno / corrupción a mitad de escritura** | Requiere un filesystem lleno controlado. Ver S2. |
| **Reloj hacia atrás en vivo** | `relojSospechoso` marca `ts_emisor` fuera de `[−24 h, +5 min]` y el reloj del daemon manda igual — verificado por lectura, no reproducido con un reloj movido. |
| **Concurrencia bajo carga real** | Probé contención ligera (6 POST simultáneos, 2 procesos). No probé miles de eventos/s ni el comportamiento del `Rollup` con debounce bajo escritura sostenida. |
| **T22 y T27** | Declarados abiertos por el constructor, correctamente. No los audito como defectos. |
| **Delta de binario contra el release anterior** | No existe el binario del release anterior a mano; medí el delta del módulo contra el mismo árbol sin telemetría (M8). |
| **Windows / macOS** | Solo Linux. `os.UserConfigDir()` y el archivado de sidecars `-wal`/`-shm` no se probaron en las otras dos plataformas. |
| **La ruta cruda como `instalacion_id`** | Ver S1: falta montar un Portafolio poblado. |

---

# UX y producto: qué le va a faltar a quien dibuje encima

## La frase objetivo, cláusula por cláusula

> *«este arnés, en este puesto, quema $X y el 60 % se va en la caja Y, que falla el gate 3 de cada
> 4 veces»*

| cláusula | ¿se puede dibujar hoy? |
|---|---|
| «este arnés» | ✅ `FilaPortafolio.arnes_id`/`clave`/`nombre`, una fila por (arnés, instalación). |
| «en este puesto» | ⚠️ `puesto` sale del `rol` del arnés indexado y **hoy ningún arnés declara `rol`** — está documentado como el caso normal. La cláusula se renderiza «puesto sin declarar» en el 100 % de los casos. |
| «quema $X» | ❌ **duplicado** (C1). Y la insignia de confianza que lo acompaña está clavada en `sin-dato` en el detalle (A1) y tiende a `sin-dato` en el resumen (M6). |
| «el 60 % se va en la caja Y» | ⚠️ `GastoCaja.Parte` existe y es correcta en forma, pero el numerador puede venir truncado (C4) y el denominador duplicado (C1). |
| «que falla el gate 3 de cada 4 veces» | ❌ **no construido**: los eventos de gate son T22, declarado abierto. `GastoCaja` no tiene ningún campo de tasa de rechazo; solo P1 lleva `corridas_usadas`/`corridas_totales`, y P1 no aplica sin señal de proceso. |

**Dos de las cuatro cláusulas no se pueden decir hoy, y una de las que sí se puede está mal.**

## Lo que le falta al contrato para que la UI no tenga que inventar

1. **Un marcador de truncado.** Ningún DTO dice «esto está cortado». Con >500 turnos, el inspector
   y la franja muestran cifras distintas y la UI no tiene cómo saber cuál creer (C4). Hace falta
   `truncado: bool` + `total_turnos: int`, o agregación en SQL.
2. **El bucket `cache_escritura_sin_tier` en el wire.** El resumen dice «incompleto por
   `cache_escritura_sin_tier`» pero el drill-down no puede mostrar **cuántos tokens** son (A3). La
   UI puede nombrar la causa y no puede cuantificarla.
3. **Un `nombre` real de caja.** `GastoCaja.Nombre = CajaID` y `DetalleCaja.Nombre = q.CajaID`.
   La UI va a pintar `paso-11` como etiqueta humana.
4. **`eventos_duplicados` en la salud.** Sin él, la pantalla de salud muestra `recibidos 618 /
   aceptados 606` y el usuario no tiene con qué explicar los 12 (A6, M5).
5. **Un desglose por modelo.** `rollup_hora` lo tiene (`modelo_canonico`) y ninguna lectura lo
   expone. El resumen no permite responder «¿qué modelo me está costando?».
6. **`Tendencia`** está en el contrato y nunca se llena (M10). La columna de tendencia del
   Portafolio queda vacía sin poder distinguir «sin datos» de «sin construir».
7. **`Corridas` es 0 fuera de S1.** Viene de `COUNT(DISTINCT corrida_id)` y `corrida_id` solo lo
   pone nuestro propio spawn. Consecuencia encadenada: `costo_por_corrida` es **siempre null** en
   `s2-instrumentado` (el gate es `Corridas > 0`). La columna «costo por corrida» del Portafolio
   está vacía para todo lo que no corra dentro de ArnesIA. Es honesto, pero es la mayoría de los
   casos y conviene decidir qué se dibuja ahí.
8. **`EstadoDetector.Hallazgos` se pone en 1 por punto**, no se cuenta
   (`telemetria_service.go:331`). Con dos puntos del mismo detector, la UI verá dos entradas con
   `hallazgos: 1`, no una con `hallazgos: 2`.

## Lo que la UI **no va a poder** mostrar honestamente

- **El total de dinero**, hasta que C1 se arregle. Cualquier gráfico de gasto va a estar al doble,
  y la divergencia entre reportado y calculado (M9) va a estar siempre en rojo, así que ni siquiera
  el oráculo la va a delatar como anomalía.
- **«Sin puntos de mejora»**, porque B6 va a devolver siempre un punto en `s2-instrumentado`
  reclamando el 100 % del gasto (M1). La lista de mejoras nunca va a estar vacía, y su primer ítem
  va a ser un falso positivo estructural.
- **Cualquier cifra de dinero de B1 o B3**, que son conteos de tokens en un campo llamado `micros`
  (M2).
- **«Se conservó el resumen»** tras la retención. No se conserva nada (A2). Si la UI del Tramo B
  dibuja el estado «detalle purgado, resumen conservado» que la doc describe, va a estar mostrando
  un estado que el backend no produce.

---

# Objeciones a decisiones firmadas

No relitigo lo firmado; lo dejo anotado.

1. **A22 (`/v1/*` acepta sin token) + la ausencia del rate limit.** La decisión está firmada y su
   fundamento medido (H10.2: `${VAR}` no se expande en el bloque `env`). Mi objeción no es a la
   decisión sino a su compensación: se justificó con tres barreras y **una de las tres no existe**
   (M3), y la única señal de crecimiento del almacén está rota (A5). Si A22 se sostiene, el rate
   limit deja de ser opcional.
2. **B4 entra al MVP aunque su recomendación sea «dónde mirar» y no un ajuste.** La objeción ya está
   registrada en el plan §9.2 y firmada en D16.1. Al verlo correr coincido con la objeción original:
   B4 dice «una caja se lleva la mayor parte» y el fix es «mirala». Eso es la franja, no una tarjeta.
3. **`turnos = COUNT(DISTINCT turno_id)` con el sesgo declarado del cruce de hora.** El sesgo está
   declarado y va en contra de la propia recomendación, que es la dirección correcta. Pero convive
   en la misma respuesta con `cobertura`, que usa otro denominador (B6). Dos denominadores en una
   pantalla es una discusión que la UI va a tener que ganar sola.

---

## Orden de ataque sugerido

| # | qué | por qué primero |
|---|---|---|
| 1 | **C1** — el doble conteo | Es la cifra del producto. Todo lo demás cuelga de ella. |
| 2 | **C2 + C3** — el 0 fabricado y el `completo` contradictorio | Son dos líneas cada uno y son pass fabricado en el wire. |
| 3 | **C4** — el truncado silencioso | Bloquea el Tramo B: la UI no puede dibujar sobre un total que a veces es parcial. |
| 4 | **A1** — la confianza clavada | Una línea. |
| 5 | **A2** — decidir el rollup: cablearlo a las lecturas o borrarlo | Hoy es código muerto con presupuesto de mantenimiento y una promesa falsa en la doc de retención. |
| 6 | **A4, A6, A5, A7** | El fail-open, los contadores mudos, el WAL invisible y la segunda barrera que no existe. |
| 7 | **M1, M2** | Los detectores mienten en dirección «hay un problema», que es la peor. |
| 8 | **M7** — el `openapi.yaml` y corregir el «Retomar aquí» | El cierre honesto tiene un hueco; corregirlo es parte del cierre. |
