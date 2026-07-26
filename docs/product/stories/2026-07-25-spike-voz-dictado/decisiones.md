# Decisiones — Spike dictado por voz en el composer (2026-07-25)

> `tipo: spike`. Cada decisión conversada se escribe EN EL MISMO TURNO (§10). Cierre = decisión
> documentada. Evidencia de respaldo: [`spike-spec.md`](./spike-spec.md).
>
> ## 🧑‍⚖️ RONDA DE FIRMA — 2026-07-25 (2ª ronda) · **TODAS LAS DECISIONES CERRADAS**
>
> El operador firmó en bloque **V-D1 · V-D2 · V-D4 · V-D5** con la recomendación de cada una, y
> resolvió la **sub-decisión** que abría V-D1 al firmar A2. **V-D6 y V-D7** quedan firmadas por
> **no-objeción** (eran «se firman salvo objeción» y no hubo objeción en esta ronda). **V-D3** ya
> estaba firmada en la 1ª ronda. **No queda ninguna decisión abierta en este paquete** — la
> construcción está desbloqueada.
>
> | # | Firma | Qué quedó |
> |---|---|---|
> | V-D1 | 🧑‍⚖️ **FIRMADA** | A2 local, modelo `base` · sub-decisión → **adaptador por PATH** |
> | V-D2 | 🧑‍⚖️ **FIRMADA** | glosario corto + últimos 2-3 turnos de `Session.Conv`; glosario global |
> | V-D3 | 🧑‍⚖️ **FIRMADA** (1ª ronda) | toggle + tope duro |
> | V-D4 | 🧑‍⚖️ **FIRMADA** | limpieza por defecto + escape a crudo |
> | V-D5 | 🧑‍⚖️ **FIRMADA** | el audio NO se persiste |
> | V-D6 | 🧑‍⚖️ **FIRMADA** (no-objeción) | el puente Rust va PRIMERO |
> | V-D7 | 🧑‍⚖️ **FIRMADA** (no-objeción) | spawn de limpieza endurecido |
> | V-D8 | — | hallazgo de honestidad, no es opción |

## V-D0 · Alcance del spike — DECIDIDA
- Responder **con qué se arma** el botón de voz, probando en vivo contra el motor real (no en papel), y
  **qué contexto** necesita el paso de limpieza para desenredar un dictado.
- **NO** construir código de producto en este spike. El entregable es papel + probes reproducibles.
- Cierre = `V-D1..V-D5` firmadas → recién ahí `spec.md` con RF numerados.

## V-D1 · Fork A — motor de STT — **FIRMADA 🧑‍⚖️ 2026-07-25 (2ª ronda)**
Opciones y su costo **real** (medido, no estimado — `spike-spec.md` §1.7):

| | Qué es | Costo honesto |
|---|---|---|
| **A1 cloud** | OpenAI Whisper API, mismo patrón que `luana-core-copilot` (ya en producción) | `OPENAI_API_KEY` **hoy sin setear**; 2° proveedor externo; **come `audio/mp4` tal cual** |
| **A2 local** | `whisper.cpp` shelleado desde Go (`exec.CommandContext`, patrón ya usado con `claude`) | **3 piezas a empaquetar**: motor + modelo (~150 MB base) + **`ffmpeg`** (mp4→WAV 16 k); latencia/CPU **sin medir** |
| **A3 nativo** | Esperar `SpeechRecognition` en WebKitGTK | ❌ ausente en 2.52.3 (probado) — no es opción hoy |
| **A0 "que lo haga Claude"** | — | ❌ los modelos Claude **no toman audio**; la suscripción cubre la limpieza, nunca el STT |

- **Recomendación (del análisis):** A1 para el MVP, A2 como destino, los dos detrás de
  `TranscriptionPort` → cambiar de motor es reemplazar un adaptador, no rehacer la feature.
- **🧑‍⚖️ RESPUESTA DEL OPERADOR (2026-07-25): "medir A2 antes de decidir".** La decisión NO se toma a
  ciegas: primero se cierra la incógnita cara (latencia/CPU real en esta máquina).
  ⇒ **`T0bis` EJECUTADO el mismo día** (§1.8).

### 📊 Medición hecha — y **da vuelta la recomendación**

- **`faster-whisper`, CPU int8, 16 CPUs, sobre 47.4 s de voz española real** (el dictado de §1.5
  sintetizado con `piper`): `tiny` **1.3 s** · `base` **2.2 s** (21× tiempo real, 1.6 s de carga en
  caliente) · `small` **5.4 s**. *(Los 91/135/180 s de las 1ª corridas eran la DESCARGA, no la carga.)*
- **Modelo recomendado: `base`.** Subir a `small` compra gramática pero **no** vocabulario de dominio
  (los tres escribieron "locita" por `lógica`, "demon" por `daemon`) — eso lo compra el contexto del
  paso siguiente, no el tamaño del modelo. Pagar 2.5× de latencia por eso es plata tirada.
- **El STT no es el cuello de botella:** 2.2 s contra 13-17 s de la limpieza. Optimizar el motor de
  transcripción sería pelear por el 13 % del gasto.
- **Hallazgo mayor:** se corrió la cadena completa con los **errores reales** del STT (`lógica`→"locita",
  `daemon`→"demon", `la app sola`→"la absola") y **la limpieza con contexto los corrigió todos**. El
  contexto de dominio no solo desambigua: **también repara el STT** ⇒ **no hace falta un modelo grande**.
- **Nueva recomendación: A2 (local) directo.** Lo único que queda en su contra es el **empaquetado**
  (motor + modelo ~150 MB + transcodificador en .deb/.rpm/.AppImage), ya no el rendimiento — que era el
  argumento entero a favor de A1.
- **Honestidad:** se midió `faster-whisper`, **no `whisper.cpp`** (sin `cmake`, pide sudo); voz
  **sintética**, no la del operador; entrada **WAV**, no el `audio/mp4` de la app. Ver §1.8.
### 🧑‍⚖️ FIRMA (2026-07-25, 2ª ronda) — **A2 local, modelo `base`, adaptador por PATH**

**Lo firmado:**

1. **Motor = A2 local** con modelo **`base`**, detrás de `TranscriptionPort`. Se descartan A1 (cloud),
   A3 (nativo, ausente) y A0 (Claude no toma audio).
2. **Sub-decisión `whisper.cpp` vs `faster-whisper` → resuelta como NINGUNA DE LAS DOS, todavía:** el
   adaptador **detecta en `$PATH`** qué motor hay (`whisper-cli`/`main` de whisper.cpp, o
   `faster-whisper`/`whisper-ctranslate2`) y usa el primero que encuentre. Si no hay ninguno,
   **degrada VISIBLE** (RF-227) en vez de romper o fingir.

**Por qué así y no eligiendo un binario ahora:**

- El argumento entero que separaba las dos opciones era **el empaquetado** (§1.8 dejó el rendimiento
  fuera de discusión), y el empaquetado **no se puede decidir con lo medido**: se midió
  `faster-whisper` sobre WAV con voz sintética, no `whisper.cpp` y no el `audio/mp4` real de la app.
  Elegir el binario hoy sería **fabricar una decisión sobre evidencia que no existe** — exactamente lo
  que la doctrina de honestidad del repo prohíbe.
- El puerto ya estaba fijo, así que la elección era **de adaptador**: postergarla no cuesta
  arquitectura. Un adaptador que detecta es el mismo trabajo que uno que hardcodea, más un `LookPath`.
- **La deuda queda VISIBLE, no silenciosa:** «qué motor STT se bundlea en el `.deb`/`.AppImage`/`.rpm`»
  se registra como **T7** (abajo), con el criterio de cierre escrito. No se cierra por olvido.
- **Lo que NO cambia por esta firma:** ningún RF del `spec.md`. Ese era el punto de `TranscriptionPort`.

**Consecuencia técnica que sí baja hoy:** el transcodificado `audio/mp4` → WAV 16 kHz mono es
**responsabilidad del adaptador** (RF-223), porque depende de qué motor se encontró. `ffmpeg` entra
por el mismo mecanismo: si no está, es un motivo de degradación visible, no un panic.

### T7 (deuda ABIERTA, derivada de esta firma) — qué motor se empaqueta

- **Pregunta:** ¿`whisper.cpp` (binario chico, hay que compilar, exige WAV) o `faster-whisper` (lo
  medido, arrastra runtime Python) se bundlea en los instaladores?
- **Qué falta para poder cerrarla:** medir `whisper.cpp` en esta máquina (necesita `cmake`, que pedía
  `sudo` el 2026-07-25) **y** medir ambos sobre el `audio/mp4` real que produce WebKitGTK, no sobre WAV.
- **Hasta entonces:** el adaptador por PATH funciona con lo que el operador tenga instalado; el
  instalador **no promete STT** — si no hay motor, el botón se deshabilita con motivo visible.

## V-D2 · Fork B — shape del contexto para la limpieza — **FIRMADA 🧑‍⚖️ 2026-07-25 (2ª ronda)**
- **El mecanismo YA está probado** (§1.5): con un bloque corto de contexto de dominio, Haiku resolvió
  referencias ambiguas reales que sin contexto quedaron genéricas. Lo que falta es de **dónde sale**.
- **Recomendación:** glosario estático corto **+ últimos 2-3 turnos** de `Session.Conv`
  (`internal/domain/session.go:127` — verificado que el campo existe). Cero cómputo nuevo.
- **Lo que NO:** ni la conversación entera ni el `CLAUDE.md` completo — el hallazgo fue que lo que
  desenreda es contexto **chico y de dominio**, no volumen.
- Abierto adentro: ¿el glosario es global o **por arnés**? Recomendado: global para arrancar, medir.

### 🧑‍⚖️ FIRMA (2026-07-25, 2ª ronda) — la recomendación tal cual

- **Insumo = glosario estático corto + últimos 2-3 turnos de `Session.Conv`.** Cero cómputo nuevo.
- **Glosario GLOBAL** para arrancar (la sub-pregunta «¿por arnés?» se cierra como *global*, con la
  puerta abierta a medir después si aparece evidencia de que hace falta).
- **Explícitamente NO:** la conversación entera ni el `CLAUDE.md` completo. El hallazgo del spike fue
  que lo que desenreda es contexto **chico y de dominio**, no volumen — meter más lo empeora y encima
  paga latencia sobre el paso que YA es el 87 % del costo.
- Baja a **RF-224**, que estaba escrito parametrizado justamente para absorber esta firma sin cambiar.

## V-D3 · Gesto de grabación: toggle vs push-to-talk — **FIRMADA 🧑‍⚖️ 2026-07-25**
- **DECIDIDO: toggle** (click empieza · click corta) **+ tope duro de duración** (p. ej. 3 min).
- **Por qué:** el detonante del spike es que el operador **habla largo**; mantener un botón apretado dos
  minutos es hostil. El tope evita que un mic olvidado abierto grabe la tarde entera.
- Baja a `spec.md` como RF con Gherkin. El valor exacto del tope (3 min) queda como default propuesto,
  ajustable sin reabrir la decisión.

## V-D4 · ¿La limpieza es obligatoria u opcional? — **FIRMADA 🧑‍⚖️ 2026-07-25 (2ª ronda)**
> No seleccionada en la 1ª ronda (el operador firmó V-D3 y mandó a medir V-D1). **No se interpretó
> como rechazo** — llegó a la 2ª ronda con su recomendación intacta y se firmó ahí.
- **Recomendación:** limpieza **por defecto** (es el corazón del pedido) **con escape a crudo**.
- **Por qué:** cuesta 13-17 s (§1.5) y a veces uno quiere el crudo y listo. Además el escape es el mismo
  escalón de fallback de la escalera de degradación (§3 Fork C.2) → sale casi gratis.

### 🧑‍⚖️ FIRMA — limpieza por defecto, con escape a crudo

- El `spec.md` ya la asumía como default: **la firma la confirma, no cambia ningún RF**.
- El escape a crudo **no es una feature aparte**: es el mismo escalón que RF-227 construye para el caso
  «la limpieza falla o tarda de más». Un solo camino de código, dos motivos para entrar.
- El texto crudo **se marca visiblemente** como sin ordenar — nunca se lo pasa por limpio.

## V-D5 · El audio NO se persiste — **FIRMADA 🧑‍⚖️ 2026-07-25 (2ª ronda)**
> Idem V-D4: no seleccionada en la 1ª ronda, no rechazada. El `spec.md` la asumía como default (no
> persistir) y la marcaba como supuesto explícito; la 2ª ronda la firma en ese sentido.
- **Recomendación: NO guardar audio.** El blob vive en memoria, se manda, se descarta; nada a disco.
- **Por qué:** menos superficie de privacidad y cero política de retención que mantener. Si algún día se
  quiere "re-escuchar lo que dicté", es una decisión nueva y explícita, no un default silencioso.

### 🧑‍⚖️ FIRMA — nada de audio a disco

- Confirma **RF-221** tal como estaba escrito. Ningún RF cambia.
- **Alcance de la firma:** cubre el audio. El **transcripto** sí entra a `Session.Conv` como cualquier
  mensaje tecleado — es lo mismo que si el operador lo hubiera escrito, no un dato nuevo.
- **Consecuencia sobre el adaptador STT (V-D1):** un motor por PATH necesita un archivo para leer. Se
  permite **un temporal de vida acotada** (crear → transcribir → borrar en `defer`), nunca un archivo
  persistente ni en el árbol del arnés. Esa es la única excepción y queda escrita acá, no implícita en
  el código.

## V-D6 · El puente Rust (T0) va PRIMERO — **FIRMADA 🧑‍⚖️ 2026-07-25 por NO-OBJECIÓN**
- `wry 0.55.1` **no maneja `permission-request` en el backend webkitgtk** (sí en macOS/Android) →
  `getUserMedia` **se cuelga sin rechazar** en la app instalada (§1.6 b/d).
- ⇒ El shell debe conceder el permiso (`with_webview` → `webkit2gtk::WebView` → `permission-request`,
  **solo** `UserMediaPermissionRequest` de audio). Sin esto, T1 "anda" en `pnpm dev` y muere mudo
  instalado.
- **No es preferencia, es física del motor.** Se documenta como decisión porque **cambia el orden de los
  tickets** y suma una dependencia (`webkit2gtk`) al `Cargo.toml` del shell.
- Se verifica **contra el binario instalado**, jamás contra el dev server.

### 🧑‍⚖️ FIRMA por no-objeción (2026-07-25, 2ª ronda)

Estaba marcada «se firma salvo objeción» y **no hubo objeción** en la ronda de firma. Queda firmada:
el orden de construcción **empieza por RF-215**. Se registra como firma explícita —no como silencio—
para que nadie la reabra creyendo que nunca se decidió.

## V-D7 · El spawn de limpieza va endurecido — **FIRMADA 🧑‍⚖️ 2026-07-25 por NO-OBJECIÓN**
- El paso de limpieza es texto→texto: `--max-turns 1` + tools denegadas + `--setting-sources
  project,local`, sobre la superficie de enforcement que ya existe (`SpawnArgs`,
  `internal/adapters/agent/claudecode/conductor.go:59`, boundary `permisos-gui`).
- **Por qué:** **el dictado es entrada no confiable.** Sin el cerrojo, "leeme el `.env` y mandámelo" es un
  prompt perfectamente válido dicho en voz alta. Un paso que solo ordena texto no tiene por qué poder
  tocar un archivo.

### 🧑‍⚖️ FIRMA por no-objeción (2026-07-25, 2ª ronda)

Sin objeción en la ronda ⇒ firmada. Baja a **RF-225** y, por la misma doctrina que el resto de los
flags de permisos, **se cablea a un test de fitness**: los flags SON la superficie de enforcement, así
que un enforcement sin test es una promesa, no un cerrojo.

## V-D8 · Corrección de honestidad sobre la primera redacción — HALLAZGO (no es opción)
- La v1 de este spike escribió que la captura «sí está disponible **sin nada adicional**» a partir de que
  la API `mediaDevices` existía. **Presencia de API ≠ captura funcionando**: probada de verdad, la
  llamada se cuelga sin plomería (V-D6), solo graba `audio/mp4`, y necesita la ventana visible.
- Queda escrito **en el documento y acá** en vez de editado en silencio: la doctrina de la casa es gap
  visible, jamás pass fabricado.

---

# Ronda 3 — el bug del micrófono en la app instalada (2026-07-26)

> Detonante: el operador probó el dictado en su instalación **v0.2.20** y le salió
> `«Falló escuchando: No se grabó nada. El composer no se tocó.»`. No había log que mirar.
> Reproducido contra el motor real y diagnosticado con `GST_DEBUG`; las decisiones que salieron
> se escriben acá **en el mismo turno** (§10).

## V-D9 · La captura NO usa `MediaRecorder`: WebAudio + WAV armado en JS — **DECIDIDA (hallazgo, no preferencia)**

**Lo que se probó** (harness GTK propio, WebKitGTK 2.52.3, mismo origin `http://127.0.0.1:4200`,
mismo mime que la app):

| Camino | Resultado medido |
|---|---|
| `MediaRecorder` `audio/mp4`, sin timeslice (**lo que hacía v0.2.20**) | `onstart` ✅ · `state: recording` ✅ · 1 evento de **0 bytes** · blob **0 bytes** · `onerror` **nunca** |
| idem + `start(250)` | **0 bytes** |
| idem + `audioBitsPerSecond: 128000` | **0 bytes** |
| **WebAudio (`ScriptProcessor`)** | **151552 muestras en 3.44 s, pico 0.0787** ✅ |

**Causa raíz**, del log de WebKit:

```
MediaRecorderPrivateGStreamer.cpp:412: Setting audio restriction caps to audio/x-raw, rate=(int)0
gstencodebasebin.c:718: <encodebin2-0> Couldn't find a compatible stream profile
gstbasesrc.c:3177: <capture-audiosrc0> error: streaming stopped, reason not-linked (-1)
MediaRecorderPrivateGStreamer.cpp:272: Transfering 0 encoded bytes
```

WebKit arma el perfil de encoding con `rate=0`, `encodebin` no casa ningún profile, el pad de audio
**nunca se linkea** y el micrófono empuja a la nada. Contribuye que faltan los plugins
`isobmff`/`fmp4` (cae al fallback `mp4mux`), pero el `rate=0` es el que mata.

**Descartado explícitamente:** el mic anda (`parec` directo: RMS 5716, pico 25131/32767), el permiso
se concede (el puente Rust de RF-215 funciona), y los encoders están (`pulsesrc ! voaacenc ! mp4mux`
a mano produjo 16905 bytes).

- ⇒ **RF-229**: el FE captura PCM por `AudioContext` + `ScriptProcessor`, remuestrea a 16 kHz y
  arma el WAV él mismo. `MIME` pasa de `audio/mp4` a `audio/wav`.
- **`ScriptProcessor` (deprecado) a conciencia:** `AudioWorklet` necesita cargar un módulo por URL y
  la app se sirve bajo CSP desde el daemon — un `blob:` de worklet es justo lo que esa política
  bloquea. El deprecado está **medido andando** en el motor real; el moderno habría que probarlo
  antes de confiarle la única vía de captura que queda.
- **Efecto lateral que conviene:** el WAV 16 kHz es lo que come `whisper.cpp` ⇒ **desaparece el
  transcodificado y con él la dependencia de `ffmpeg`**, que `spike-spec.md` §1.7 encontró AUSENTE
  en esta máquina.
- **Corrige V-D8 y la BR `mime-explicito`:** «solo graba `audio/mp4`» era optimista. Lo exacto es
  que `isTypeSupported('audio/mp4')` **devuelve `true` y miente**: el único formato que este motor
  declara soportar es también el único que no puede producir.

## V-D10 · El daemon escribe a un archivo de log — **DECIDIDA**

El operador preguntó «¿hay algún log que puedas revisar?». **No lo había**: el daemon logueaba solo a
stderr y la app instalada lo lanza desde el `.desktop` del `.deb`, que no tiene terminal. Un bug vivió
una versión entera y del incidente solo quedó la frase que el operador leyó en pantalla.

- ⇒ **RF-230**: `~/.arnesia/logs/arnesia.log`, JSON por línea, rota a 8 MB conservando una
  generación. Se apaga con `--log -`, se muda con `--log <ruta>` o `$ARNESIA_LOG`.
- **stderr sigue en texto y el archivo va en JSON.** No es inconsistencia: en la terminal lee una
  persona, y el archivo lo lee un `grep` después de un incidente sobre un `detalle` anidado que el
  handler de texto aplastaría. Un `io.MultiWriter` forzaría un solo formato — por eso el fan-out.
- **Un log que no se puede abrir no tumba el daemon:** avisa y sigue con stderr. Degradación honesta.
- El endpoint de dictado además loguea **cada** dictado (formato, bytes, ms, estado, motor), no solo
  los que fallan: sin la línea del camino feliz no hay con qué comparar cuando algo empieza a fallar.

## V-D11 · El FE puede escribir en ese log (`POST /api/diagnostico`) — **DECIDIDA**

El WebView **no tiene devtools** y su `console.error` no va a ningún lado. El daemon es el único
proceso de la app que escribe a disco.

- ⇒ **RF-230**: endpoint que recibe `{origen, evento, mensaje, detalle}` y lo escribe al log.
  El store del dictado reporta en el mismo acto en que marca el fallo visible, y un cazador global
  (`window.onerror` + `unhandledrejection`) cubre lo que ningún `try/catch` nuestro ve.
- **`detalle` es libre y nadie parsea su forma como contrato** (mismo criterio que
  `conductor-no-parsea-jsonl.md`): cada fallo tiene sus propias variables, y un schema fijo haría que
  la próxima falla desconocida no tuviera dónde contarse.
- **Acotado:** 64 KB por cuerpo, 30 eventos por minuto en el FE, claves recortadas a 120 chars. Un
  diagnóstico que rompe el flujo que estaba diagnosticando no sirve.
- **NO es telemetría de producto.** Destino: loopback → archivo local. No sale de la máquina. HS-27
  (OTel) es otro paquete, otra decisión y otro destino — esto no lo adelanta ni lo reemplaza.

## V-D12 · «No se grabó nada» se reemplaza por dos motivos distintos — **DECIDIDA**

El mensaje viejo describía el síntoma y escondía la causa. Se parte en dos, porque se arreglan
distinto:

- **sin bloques** → «El micrófono no entregó ni un bloque de audio» (grafo/device roto).
- **silencio digital** (pico < 0.001, medido: hablar normal da ≈ 0.08) → «El micrófono entregó Ns de
  silencio (nivel 0)» (mic muteado o tomado por otra app).

Cortar en el FE evita mandarle 3 minutos de ceros a un STT que devolvería vacío, haciendo parecer que
el bug es de la transcripción.

## ⏳ Pendiente de firma 🧑‍⚖️

V-D9..V-D12 quedan **DECIDIDAS y construidas**, sin firma del operador todavía. Lo que falta para
firmar es el tramo E2E contra el **binario instalado** (mismo gate que ya tenía abierto el paquete
para el micrófono), no más discusión de diseño.
