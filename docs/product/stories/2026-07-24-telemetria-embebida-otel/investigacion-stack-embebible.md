# Investigación — stack de telemetría embebible en el binario Go (2026-07-26)

> Insumo para D6 (cero-instalación cross-platform) y para el `spec.md`.
> **`[V]` = medido/compilado en esta máquina (Go 1.26.4 linux/amd64) · `[V-doc]` = verificado en
> fuente oficial · `[I]` = inferido / fuente secundaria.**

## 0. La pregunta ya estaba contestada por el repo — verificado

| artefacto | contenido |
|---|---|
| `boundaries/indice-desechable-jsonl-es-verdad.md:88` | check **`sin-cgo`** · severidad `error` · «dependencia CGO rompe el binario único» |
| ídem `:78` | *«**DuckDB descartado** (es CGO → mata el binario estático cross-compile). Si años de telemetría lo exigen: DuckDB out-of-process sobre Parquet, nunca el core»* |
| `fitness/arch_test.go:214` | `TestNoDuckDBOrCGOStore` — **enforced**, banea `duckdb` y `mattn/go-sqlite3` en `domain`/`usecase`/`adapters/index` |
| `internal/adapters/index/store.go:1` | ya corre `modernc.org/sqlite` *pure Go, no CGO* |
| `go.mod:20` | `modernc.org/sqlite v1.54.0` presente |

La investigación **ratifica** la decisión y aporta el número que faltaba para la cláusula «si años de
telemetría lo justifican» (§3). **No se relitiga `sin-cgo`.**

## 1. Receptor OTLP — cuatro rutas, medidas

Binarios reales compilados. Bytes. Daemon actual de referencia: **17 604 873 B** `[V]`.

| ruta | binario | `-s -w` | módulos | paquetes |
|---|---:|---:|---:|---:|
| *baseline* `net/http`+`encoding/json` | 8 836 730 | 6 119 689 | 1 | 188 |
| **A** JSON a mano, cero deps | 8 836 730 | 6 119 689 | **1** | 188 |
| **B** `proto/otlp` + protobuf | 18 012 662 | 12 427 529 | 47 | 324 |
| **B'** B + `protojson` | 18 067 546 | 12 464 393 | 47 | 326 |
| **C** ⭐ **`collector/pdata`** (`pmetricotlp`+`plogotlp`) | 19 325 361 | 13 340 937 | 61 | 332 |
| **D** `collector/receiver/otlpreceiver` completo | 27 691 966 | 19 149 065 | **121** | 484 |

Lecturas `[V]`:
1. **`protojson` es gratis** (+54 806 B sobre B). El costo son los descriptores OTLP + reflexión de
   protobuf: **+9,2 MB** sobre A. «Protobuf en vez de JSON para ahorrar peso» **no ahorra nada**.
2. **D cuesta +8,4 MB y +60 módulos sobre C** — pagás gRPC, `configtls`, `configauth`, `confmap`,
   zap. En loopback no se usa ninguno.
3. **C es el punto dulce**: unmarshalers OTLP correctos por **+1,3 MB sobre B'**.

### Estabilidad — el criterio que decide C vs D

| módulo | versión | estabilidad | riesgo |
|---|---|---|---|
| `collector/pdata` | **v1.63.0** `[V]` | **v1.x ⇒ API Go estable** | bajo |
| `collector/receiver/otlpreceiver` | v0.157.0 `[V]` | señales *Stable*, **módulo en 0.x** | **alto** |
| `proto/otlp` | v1.11.0 `[V]` | v1.x | bajo |
| protocolo OTLP/HTTP | spec 1.11.0 | *Stable* | nulo |

La doc del Collector distingue **estabilidad del componente** (config+output) de **estabilidad de la
API Go**, y la segunda solo se garantiza desde 1.x `[V-doc]`. `otlpreceiver` marca sus señales
"Stable" **estando en v0.157.0**: la config es estable, la API que importás **no**. Atarse a eso
sobre 121 módulos = deuda de mantenimiento recurrente.

### ⚠️ Trampa: `protojson` NO cumple la spec OTLP/JSON

OTLP/JSON se desvía del proto3 JSON mapping en dos puntos `[V-doc]` (OTEP-0122):
1. **`traceId`/`spanId` en HEX**, no base64.
2. **Enums solo como enteros**.

`protojson` implementa el mapping **estándar** ⇒ los lee como base64. Como el alfabeto hex es
subconjunto del de base64, **puede decodificar basura en silencio** `[I — derivado de la spec, no
observado; el test empírico se perdió]`. No afecta `/v1/metrics`, **sí afecta `/v1/logs`** — y Claude
Code manda `api_request`/`tool_decision` **como logs**. `pdata` trae unmarshalers OTLP-JSON escritos
a mano que respetan hex+enums-int. **Razón técnica, además del peso, para C sobre B'.**

### Dato operativo: Claude Code exporta por **gRPC** por default

Default `[V-doc]`: `grpc` → `http://localhost:4317`. Alternativas `http/protobuf` y `http/json` →
`:4318/v1/{metrics,logs,traces}`. Intervalos: métricas **60 s**, logs **5 s**.

⇒ **Al spawnear, ArnesIA debe setear `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`** además del
endpoint. Una línea en el spawn **elimina gRPC** y con él la mitad del peso del receptor completo.

### `ocb` no sirve para embeber `[V-doc]`

Lee un `manifest.yaml` y **genera un paquete `main` completo** que compila a un binario standalone.
Es un generador de distribuciones, no un empaquetador de librerías. Produce exactamente la ruta D,
más un paso de codegen en el build.

## 2. Almacenamiento — SQLite pure-Go, medido

`modernc.org/sqlite`, SQLite 3.53.3, 500 000 filas sintéticas, WAL + `synchronous(NORMAL)` `[V]`:

| operación | resultado |
|---|---|
| INSERT 500 k (1 tx, prepared) | **1,191 s → 419 798 filas/s** |
| tamaño en disco | **47,2 MB ≈ 94 B/fila** |
| `group by model` + `sum(cost)` (full scan) | **323 ms** |
| buckets horarios 24 h (con índice) | **8,9 ms** |
| top-20 sesiones por costo | 357 ms |
| **p95 vía `ntile(100)`** | **2 092 ms** ← único punto débil |

Features: window functions ✅ · `json_extract` ✅ · tablas `STRICT` ✅ ·
**`percentile_cont()`/`median()` ❌** (modernc no carga extensiones `.so`; sí permite registrar
funciones Go propias).

### La mitigación que resuelve el problema de raíz — rollup horario `[V]`

| operación | resultado |
|---|---|
| construir rollup horario desde 500 k filas | 1 820 ms → **3 480 filas** |
| `sum(cost) group by model` **sobre el rollup** | **1,7 ms** (vs 323 ms crudo → **190×**) |
| serie diaria sobre el rollup | **1,4 ms** |

**Es la arquitectura correcta e independiente del motor:** eventos crudos + rollup incremental por
(hora × runtime × modelo). Dashboard lee el rollup en ~2 ms; el drill-down toca lo crudo. Con eso,
**la ventaja analítica de DuckDB no tiene a quién ganarle.**

Para p95: histograma de buckets de 10 ms ⇒ **347 ms** (−83 % vs `ntile`). Para latencias de agentes,
un bucket de 10 ms sobra.

### Umbrales donde SQLite dejaría de alcanzar

| señal | umbral |
|---|---|
| full scan `group by` &gt; ~1 s percibido | **&gt; ~1,5 M filas** `[I, extrapolado]` |
| el `.db` incomoda en disco | **&gt; ~10 M filas ≈ 940 MB** `[V, extrapolado]` |
| p99 exactos multi-dimensionales en tiempo real | ya hoy — mitigable con histogramas |

**Con rollups ninguno se toca al volumen declarado.** El primero en tocarse sería el de disco, y la
respuesta correcta es **retención + archivado a Parquet**, no cambiar de motor.

## 3. DuckDB y chDB — el número para la cláusula de reapertura

**DuckDB** (`github.com/duckdb/duckdb-go`, migrado desde `marcboeker`) `[V]`:

| métrica | valor |
|---|---:|
| descarga del módulo | **~240 MB** |
| expandido en disco | **1 082 170 755 B ≈ 1,08 GB** |
| `libduckdb.a` linux_amd64 / windows_amd64 / darwin_arm64 | 85 MB / 76 MB / 61 MB |

Más: **Windows exige MinGW/MSYS2, no MSVC** `[V-doc]` — pero el resto del bundle Tauri en Windows
**sí es MSVC**: dos toolchains en el mismo runner. macOS universal con CGO exige dos builds + `lipo`.
Y rompe `TestNoDuckDBOrCGOStore`.

⚠️ **No verificado:** el tamaño del binario Go final con DuckDB linkeado (el build falló dos veces
por timeouts de `proxy.golang.org`). El «~30 MB agregados» es fuente secundaria `[I]`.

**Veredicto: DuckDB no se descarta por tamaño — se descarta por CGO y por innecesario.** SQLite hace
el `group by` completo sobre 500 k filas en 323 ms y en 1,7 ms contra el rollup.

**chDB queda fuera por tres motivos simultáneos, cualquiera letal** `[V-doc]`: (a) **no soporta
Windows**; (b) requiere `libchdb.so` **dinámica instalada aparte** ⇒ el usuario instala algo, viola
D6; (c) la instalación documentada es `curl -sL https://lib.chdb.io | bash`, que además choca con
«nada se descarga en post-install».

**Prometheus TSDB**: **292 módulos** `[V]` (2,4× el receptor completo del Collector) y su modelo es
*float64 por serie temporal* — **no guarda eventos con atributos arbitrarios**, que es justo lo que
llega del `stream-json`. Descartado. **VictoriaMetrics** se distribuye como servicio, no como
librería embebible `[I]`. **Parquet** no es motor de consulta: sirve para **archivado frío**.

## 4. La ingesta multi-origen refuerza la elección, no la cambia

Si entran **dos orígenes heterogéneos** (OTLP + parseo de `stream-json` por stdout), el esquema
canónico **no puede ser el modelo de datos de OTel** — hay que normalizar a un evento propio de todos
modos. Eso **degrada `pdata` de "modelo de datos" a "decodificador de wire format"**, y refuerza
elegir el decodificador más barato y estable en vez del framework más completo.

**Diseño que se cae de maduro** (encaja con la hexagonal vigente): un único `ports.TelemetrySink` con
dos adaptadores (`otlp/` y `streamjson/`), ambos emitiendo el mismo `domain.EventoTelemetria` → un
solo writer serializado a SQLite. Un esquema relacional con `runtime` como columna absorbe ambos sin
fricción.

## 5. Empaquetado y firma — los riesgos reales

### Sidecars Tauri `[V-doc]`
Sufijo con target triple **obligatorio** (`-x86_64-pc-windows-msvc.exe`, `-aarch64-apple-darwin`…).
`externalBin` es un **array** ⇒ admite varios. **No hace falta sidecar nuevo para telemetría: todo va
dentro del `arnesia-daemon` que ya existe** — que es exactamente el requisito de D6.

### 🔴 Riesgo #1 — `tauri#11992`: la notarización macOS falla con `externalBin`
La notarización devuelve **`The signature of the binary is invalid`** para `Contents/MacOS/<app>`
**en cuanto agregás `externalBin`**; sacándolo, notariza bien. Issue **abierto / needs triage**
`[V-doc]`. **ArnesIA usa `externalBin` ⇒ esto lo va a pegar de frente al abrir el canal macOS.**
Presupuestar `codesign` manual post-bundle como workaround, y **probarlo temprano, no en la release**.

Sin firmar en macOS: desde **Sequoia (15.x) Apple eliminó el bypass por Control-click** `[V-doc]`.
El camino ahora es abrir → falla → System Settings › Privacy & Security › "Open Anyway" → password →
reintentar → **otro diálogo**. Cinco pasos, dos diálogos de miedo. Inaceptable para un vendible.

### 🔴 Riesgo #2 — Azure Artifact Signing probablemente NO aplica a alpacapurpura
Está **geográficamente restringido: orgs en USA/Canadá/UE/UK; individuos solo USA/Canadá** `[V-doc]`.
Operación latinoamericana ⇒ **verificar elegibilidad antes de presupuestarlo**. Plan realista:
**certificado OV, USD 150–300/año**, con **clave privada en HSM cloud** (desde jun-2023 el CA/Browser
Forum lo exige) `[V-doc]`.

Dato que ahorra plata: **EV ya no compra bypass de SmartScreen** — Microsoft lo removió en 2024
`[V-doc]`. Pagar EV para saltear SmartScreen ya no se justifica.

### 🔴 Riesgo #3 — la reputación de SmartScreen se reinicia por release
Ni Azure ni OV dan pase instantáneo: **la reputación se acumula por hash de archivo** `[V-doc]`.
Con releases frecuentes (vamos por v0.2.18) puede no consolidarse nunca. Argumento para **espaciar
releases públicos** o distribuir por winget/Store.

### Post-install: informativo, no decisivo
`.deb` ✅ `postinst` · `.rpm` ✅ · NSIS ✅ 4 hooks · `.msi` ⚠️ solo compilable en Windows, y
`tauri#5970` («custom actions aren't bundled from WiX fragments») · **`.dmg` ❌ no existe** (es
drag-to-Applications; requeriría `.pkg`, que Tauri no bundlea).

⇒ **Bajo la restricción ya tomada (nada se descarga en post-install), el `.dmg` sin post-install es
el mínimo común denominador que gobierna el diseño. La solución correcta es CERO post-install en las
tres plataformas — exactamente lo que da un binario pure-Go embebido.**

### Presupuesto de tamaño — sobra
Medido en el repo `[V]`: daemon **17,6 MB** · `.deb` **6,9 MB** · `.rpm` **6,9 MB** ·
`.AppImage` **79 MB**. Un Electron equivalente ronda 120–200 MB `[I, fuente secundaria]`.
Sumar telemetría por la ruta recomendada (**+~1,7 MB**) es ruido estadístico.

## 6. Veredicto

| capa | elección | peso extra |
|---|---|---:|
| **recibir OTLP** | handler `net/http` propio en `/v1/metrics` + `/v1/logs`, decodificando con **`collector/pdata`** (`pmetricotlp`/`plogotlp`) | **+~1,7 MB** `[V]` |
| **almacenar** | **`modernc.org/sqlite`** (ya en `go.mod`) — tabla de eventos normalizados + rollup horario incremental | **0** |
| **consultar** | SQL sobre el rollup; drill-down contra crudo; **p95 por histograma de buckets** | 0 |
| **ingesta no-OTLP** | parser `stream-json` en Go puro → **mismo `domain.EventoTelemetria`** → mismo writer | ~0 |

**Config clave del spawn:** `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` +
`OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318`. Evita gRPC por completo.

**Presupuesto final:** daemon 17,6 MB → **~19,5 MB** `[V, por diferencia de builds]`; `.deb` 6,9 →
**~7,5 MB** `[I]`.

### Orden de ejecución

1. esquema `evento_telemetria` + writer serializado (reusar patrón de `adapters/index/store.go`) — bajo
2. handler `/v1/metrics` + `/v1/logs` con `pdata`, montado en el router 127.0.0.1:4200 existente — bajo
3. inyectar env OTEL en el spawn de `claude` — bajo
4. rollup horario incremental + consultas del dashboard — bajo
5. **adaptador `stream-json`** para Codex/Amp/Cursor — **medio**: formatos sin spec, cambian sin aviso
6. retención + archivado a Parquet — solo si se toca el umbral de disco

Todo cambio arranca por su **capability YAML** (doctrina R1/R2/R3). Módulo sugerido: `telemetria/`.

### Descartados y por qué

| descartado | motivo decisivo |
|---|---|
| `otlpreceiver` completo | +8,4 MB, +60 módulos, **módulo 0.x** ⇒ churn de API Go, por features que en loopback no se usan |
| `ocb` | genera binario standalone, no librería |
| `protojson` para OTLP-JSON | **no es spec-compliant** (hex vs base64 en trace IDs) y no ahorra peso |
| DuckDB | CGO + `sin-cgo` + 240 MB CI + MinGW en Windows + **innecesario al volumen declarado** |
| chDB | **sin Windows**, `libchdb.so` dinámica, install `curl \| bash` |
| Prometheus TSDB | 292 módulos; floats-por-serie no absorbe eventos con atributos arbitrarios |
| Parquet como motor primario | es un codec, no un engine — sirve para archivado frío |

## 7. Nota de honestidad del investigador

Tres cosas **no verificadas**, no usar como hecho:
1. **Tamaño del binario Go con DuckDB linkeado** — build falló por timeouts de red. El «~30 MB» es
   secundario `[I]`. Sí verificados: 240 MB de descarga, 1,08 GB en disco.
2. **El test `protojson` vs `pdata` sobre OTLP-JSON con trace IDs hex** — derivado de la spec/OTEP,
   **no observado**. Barato de confirmar: ~20 líneas de test. **Hacerlo antes de comprometer B'.**
3. **Tamaños de apps Electron** (Slack/Obsidian) — fuentes secundarias. Los números de ArnesIA sí
   son medidos.
