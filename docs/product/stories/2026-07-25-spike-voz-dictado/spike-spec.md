# Spike — Dictado por voz en el composer (STT + limpieza con contexto)

> Paquete `2026-07-25-spike-voz-dictado`. Autocontenido: no asume que quien lo lee vio la conversación
> original. Etapa: investigación **CERRADA**, [`spec.md`](./spec.md) **escrito** (RF-215…RF-228);
> **V-D1 · V-D2 · V-D4 · V-D5** ([`decisiones.md`](./decisiones.md)) esperan firma 🧑‍⚖️ antes del
> *build* (ya no del papel: el spec está redactado contra `TranscriptionPort` para no depender del
> motor). Todo lo marcado "probado hoy" tiene comando + output real pegado abajo (nada fabricado —
> doctrina `docs/architecture/boundaries/codigo-traza-a-capability.md` §honestidad).
>
> **Rev 2 (2026-07-25):** segunda vuelta de pruebas. La captura pasó de "asumida disponible" a
> **probada, con un bloqueante Rust adelante** (§1.6); se corrigió un salto de honestidad de la v1
> (§1.2); apareció el ticket **T0**; y el Fork C se abrió en 4 decisiones que no estaban.
>
> **Rev 3 (2026-07-25):** el operador firmó **V-D3** y ordenó **medir A2 antes de decidir V-D1** →
> §1.8: el STT local cuesta **2.2 s**, no es el cuello de botella, y **la limpieza con contexto corrige
> los errores del STT** — se da vuelta la recomendación hacia A2 local. `spec.md` escrito.

## 0. Norte

El operador quiere hablarle a ArnesIA en vez de escribir: un botón de grabar en el composer del chat
que capture voz, la transcriba, y **antes de mandarla** la pase por un paso que la entienda y la ponga
en orden — porque habla mucho y se enreda, pero "alguien que conoce del tema" lo entiende bien. Pidió
explícitamente: (a) que se pruebe en vivo, no solo se diseñe en papel, y (b) que se investigue cómo
darle **contexto** a ese paso de limpieza para que resuelva el enredo, no solo que le saque las
muletillas.

## 1. Estado real (probado hoy, 2026-07-25)

### 1.1 Sin prior art propio en ArnesIA

Grep de `micrófono|speech|whisper|voice|dictado|transcrib` sobre `docs/`, `web/src`, `web/src-tauri`,
`internal/` → nada relacionado a voz. Feature nueva de punta a punta en este repo.

### 1.2 WebKitGTK nativo — PROBADO, negativo (mata la hipótesis inicial)

La sesión anterior (research web) encontró que WebKitGTK viene integrando un motor de reconocimiento
nativo basado en Whisper.cpp detrás del propio `SpeechRecognition` estándar — pero esa fuente era de
2023 y no confirmaba el estado 2026. Se probó **en vivo, contra el motor real instalado en esta
máquina** (`libwebkit2gtk-4.1 2.52.3`, el mismo que usa el shell Tauri de ArnesIA): se levantó un
`WebKit2.WebView` real (PyGObject, GObject introspection `WebKit2-4.1.typelib`) sobre el `DISPLAY :0`
real de la máquina y se evaluó JS adentro.

```python
# webkit_speech_test.py — WebView real, HTML inline que chequea las APIs y reporta por window.title
out = {
  SpeechRecognition: 'SpeechRecognition' in window,
  webkitSpeechRecognition: 'webkitSpeechRecognition' in window,
  mediaDevices: !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia),
}
```

**Resultado real:**

```json
{"SpeechRecognition":false,"webkitSpeechRecognition":false,"mediaDevices":true,
 "userAgent":"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/605.1.15 ... Safari/605.1.15"}
```

`SpeechRecognition`/`webkitSpeechRecognition`: **ausentes**. La integración nativa de la research
anterior sigue sin llegar al paquete `2.52.3` que corre esta app hoy — se descarta como plan A. Si
Ubuntu empaqueta una versión más nueva de WebKitGTK en el futuro, vale re-correr este mismo test antes
de invertir en una alternativa (queda barato: el script vive en este paquete, ver §5).

`mediaDevices.getUserMedia`: **presente**, en la misma WebView real. ⚠️ **Ojo — presencia de API ≠
captura funcionando**, y la primera redacción de este documento dio ese salto («sí está disponible sin
nada adicional»). Se re-probó a fondo (§1.6): la API existe, pero llamarla sin plomería adicional
**cuelga para siempre**. Corregido acá para no dejar un pass fabricado en el papel.

### 1.3 Prior art real: ya existe en un producto hermano del operador

`luana-core-copilot` (`/home/chalreme/Proyectos/luana-vitalia/core/luana-core-copilot/src/
luana_core_copilot/infrastructure/voice/whisper_transcriber.py`) ya tiene un `WhisperTranscriber` en
producción: recibe `(audio: bytes, mime_type)`, llama `openai.audio.transcriptions.create(model=
"whisper-1", language="es", response_format="verbose_json")`, detrás de un `TranscriptionPort`
(Protocol) en `domain/voice.py`. Es cloud (OpenAI, no local), y **no** tiene ningún paso de limpieza
con contexto — transcribe y listo. Prueba que el patrón captura-navegador → backend → Whisper YA está
validado y corriendo en un sistema real del operador; no hay que inventar esa mecánica desde cero, solo
decidir si ArnesIA la replica (cloud) o va local (§3, Fork A).

### 1.4 Gap honesto — lo que NO se pudo probar hoy

*(Escrito en la 1ª vuelta; **§1.8 cerró buena parte de esto** — se deja el registro de qué faltaba y
cuándo, en vez de reescribir la historia.)*

En la 1ª vuelta no había motor STT local instalado ni `OPENAI_API_KEY` ni forma de sintetizar voz de
prueba, y no se instaló nada sin confirmar antes. **Después el operador ordenó medir** → §1.8 montó un
venv aislado (sin `sudo`, sin tocar el sistema) con `faster-whisper` + `piper`, y corrió la cadena
**voz sintética → STT → limpieza con contexto**, completa.

**Lo que SIGUE abierto (ticket T6):** el extremo real de los dos puntas — **la voz del operador, por el
micrófono real, en formato `audio/mp4`, contra el binario instalado**. §1.8 usa voz sintética y entrada
WAV; §1.6 prueba que el mic abre pero no graba. **La prueba conjunta de verdad no está hecha** — queda
visible, no fabricada como "andaba".

### 1.5 La parte núcleo del pedido — PROBADA end-to-end, con evidencia real

Esto sí se corrió completo, dos veces, mismo transcripto, cambiando solo si había contexto. Transcripto
de prueba (simula un dictado real, con las mismas referencias ambiguas de un dictado sin editar):

```
a ver o sea lo que pasa es que ayer estuvimos viendo lo de la tarjeta esa cuando pregunta algo
y bueno funcionó pero yo creo que también deberíamos hacer que cuando uno está eligiendo
la opción se vea de una vez cuál es la que ya está mandada digamos que si ya se envió no se pueda
tocar más porque si no uno hace doble click sin querer y bueno también quería decirte
lo otro que es que en el mapa cuando aparece el nodo ese que decía no reconocido
capaz que ahí también convendría poner algo parecido no sé si es la misma lógica de tarjeta
o es otra cosa aparte, y ah me acordé también lo del daemon que a veces lo dejo corriendo
y después no sé si el que está corriendo es el bueno o el que arranca la app sola, eso me
genera dudas cada vez que pruebo algo
```

Corrido real vía `claude -p "<instrucción+transcripto>" --model claude-haiku-4-5-20251001` (headless,
one-shot — sin sesión, sale y listo):

**A — sin contexto** (13.0s reales):
> Desactiva visualmente las opciones de tarjeta ya enviadas para evitar doble-click accidental. Aplica
> lógica similar al nodo "no reconocido" del mapa [¿idéntica a tarjeta o distinta?]. Clarifica cuál
> daemon está activo cuando hay uno en background vs el que arranca la app automáticamente—genera
> confusión al probar.

**B — con contexto** (17.2s reales) — se le agregó un bloque corto (glosario de 4 líneas: qué es
`PermissionCard`, qué significa el chip "(no reconocido)" del Mapa, y los 2 modos en que corre el
daemon — el mismo tipo de dato que ya vive en `Session.Conv`/el estado real de la sesión, no algo
inventado para la prueba):

> PermissionCard: mostrar qué opción fue enviada y bloquear clics posteriores; evaluar aplicar lógica
> similar a chips '(no reconocido)' del Panel Base del Mapa; aclarar cuál daemon corre cuando hay
> override en ~/.local/bin/arnesia.

**Comparación:** sin contexto, el modelo entendió los 3 pedidos pero se quedó en genérico ("tarjeta",
"nodo", "daemon en background") y dejó una ambigüedad sin resolver. Con contexto, nombró el componente
real (`PermissionCard`), el elemento real (chip "(no reconocido)" del Panel Base) y el mecanismo real
(`~/.local/bin/arnesia` override) — quedó **directamente accionable**, sin necesitar una vuelta de
aclaración. Esto confirma la intuición del operador ("alguien que conoce del tema me entiende bien"):
lo que resuelve el enredo no es más audio ni más transcripción — es contexto de dominio compacto, no
el historial completo de la conversación ni el `CLAUDE.md` entero.

Latencia real (una sola corrida, orientativa): ~13-17s por limpieza vía CLI headless. Aceptable para
"grabás, esperás un toque, revisás el texto en el composer" — no para respuesta en tiempo real mientras
hablás.

### 1.6 La captura — PROBADA de verdad (segunda vuelta, 2026-07-25)

La primera vuelta solo preguntó «¿existe la API?». No alcanza: **existir no es funcionar**. Se re-probó
con el mismo aparato (WebView WebKitGTK 2.52.3 real sobre `DISPLAY :0`), esta vez llamando a
`getUserMedia` de verdad y mirando `MediaRecorder` (script: `webkit_captura_test.py`, en este paquete).
Cuatro hallazgos, todos con output real:

**(a) `MediaRecorder` existe — pero solo sabe grabar `audio/mp4`.**

```json
{"MediaRecorder":true,
 "mimes":{"audio/webm":false,"audio/webm;codecs=opus":false,"audio/ogg":false,
          "audio/ogg;codecs=opus":false,"audio/mp4":true,"audio/wav":false}}
```

El `mimeType` por defecto del `MediaRecorder` viene **vacío** (`""`), así que el FE tiene que pasar
`{mimeType:'audio/mp4'}` explícito. Esto NO es cosmético: el formato de salida decide el costo del
Fork A — mp4/m4a lo come la API de Whisper cloud tal cual (A1), mientras que `whisper.cpp` quiere
WAV 16 kHz mono → A2 arrastra **un transcodificador (ffmpeg) además del motor y el modelo** (§1.7).
El plan anterior asumía webm/opus, que en este motor **no existe**.

**(b) Sin manejar el permiso, `getUserMedia` no falla: se CUELGA.** Corrida sin handler de
`permission-request`: la promesa quedó pendiente, sin resolver ni rechazar, hasta el timeout del script
— el estado se quedó clavado en `calling-gum` (watchdog de 3 s incluido, nunca avanzó). **No hay
excepción que catchear.** Es el peor modo de falla posible para un botón de mic: no tira error, no
muestra nada, se queda mudo para siempre.

**(c) Con la ventana visible + el permiso concedido, funciona y abre el mic REAL:**

```json
{"stage":"gum-ok-stream-stopped","visibility":"visible","secureContext":true,
 "getUserMedia":"OK",
 "tracks":[{"label":"Family 17h/19h HD Audio Controller Digital Microphone","state":"live"}]}
```

con la señal del lado GTK:

```
permission-request FIRED: UserMediaPermissionRequest
```

Device real, `state: live`. *(El stream se cortó en el acto —`track.stop()`— dentro del mismo tick: la
prueba confirma que el mic ABRE, no graba nada ni escribe audio a disco.)* Dos detalles que caen de acá:
`isSecureContext: true` sobre **`http://127.0.0.1:4200`** — el origin real que sirve el daemon, probado
literal, no por analogía con `localhost` — o sea **no hace falta HTTPS ni certificados** para esto; y la
`permission-request` solo se emite con la página **visible** (en `Gtk.OffscreenWindow` ni siquiera
llegaba a emitirse), consistente con que WebKit difiere la captura mientras el documento está oculto.

**(d) ⛔ El bloqueante de verdad: `wry` NO maneja ese permiso en Linux.** Inspección del código de la
dependencia real del shell (`wry 0.55.1`, la que fija `web/src-tauri/Cargo.lock:4962`):

```
grep -rn "permission-request|permission_request|UserMediaPermission" wry-0.55.1/src/  → (sin resultados)
grep -rn "permission" wry-0.55.1/src/
  → src/wkwebview/class/wry_web_view_ui_delegate.rs:127  fn request_media_capture_permission(   # macOS
  → src/android/kotlin/RustWebChromeClient.kt:44         permissionLauncher                     # Android
```

wry resuelve el permiso de captura en **macOS y Android**, y **no** en el backend `webkitgtk` — la
única plataforma donde ArnesIA corre hoy. Combinado con (b): **si se construye solo el FE, el botón de
mic funciona en `pnpm dev` (Chromium/Firefox) y se cuelga mudo en la app instalada.** Eso convierte el
puente Rust en el **primer** ticket (T0, §5), no en un detalle de integración: `with_webview` →
`webkit2gtk::WebView` → conectar `permission-request` → conceder **solo** `UserMediaPermissionRequest`
de audio del origin del daemon (encuadre `superficie-local-confinada`: el shell es la raíz de confianza,
concede el mic acotado — jamás un allow-all de permisos del WebView).

### 1.7 Inventario real de la máquina para el Fork A (qué hay y qué falta hoy)

Chequeado en esta máquina, hoy:

| Pieza | Estado real |
|---|---|
| GStreamer captura+encode (`pipewiresrc`·`pulsesrc`·`alsasrc`·`voaacenc`·`avenc_aac`·`mp4mux`) | ✅ presentes (por eso (c) funciona) |
| Hardware de mic (`ALC245 Analog` + DMIC `acp63`, PipeWire) | ✅ presente |
| `ffmpeg` | ❌ ausente — lo pide A2 para mp4 → WAV 16 kHz |
| `whisper` / `whisper-cli` / `faster_whisper` | ❌ ausentes |
| Modelo ggml/whisper descargado | ❌ ninguno |
| `OPENAI_API_KEY` | ❌ sin setear |

Ninguna de las dos ramas del Fork A es gratis: **A1** necesita credencial nueva (hoy no existe acá),
**A2** necesita empaquetar motor + modelo + transcodificador en los instaladores (.deb/.rpm/.AppImage),
o sea: peso de bundle. Esta tabla es el costo honesto que faltaba para poder firmar el Fork A con datos
en vez de con intuición. → **La latencia de A2 se midió después, en §1.8** (venv aislado con `uv`, sin
`sudo`, sin instalar nada global).

### 1.8 T0bis — el Fork A2 MEDIDO (2026-07-25, por orden del operador: "medir antes de decidir")

El operador no quiso firmar V-D1 a ciegas. Se midió. **Y el resultado da vuelta la recomendación.**

**Aparato:** venv aislado (`uv`, sin `sudo`, sin tocar el sistema) + `faster-whisper` CPU `int8` sobre
esta máquina (16 CPUs). Audio de prueba: **el mismo dictado enredado de §1.5**, sintetizado como voz
española real con `piper` (`es_ES-davefx-medium`) → **47.4 s de habla**. Nada de silencio ni ruido: voz
de verdad, con las mismas referencias ambiguas.

| Modelo | Carga (caché tibia) | Inferencia | Factor tiempo real | Calidad sobre este dictado |
|---|---|---|---|---|
| `tiny` | 0.5 s | **1.3 s** | 36.9× | pobre: "tajeta", "diamond", frases rotas |
| `base` | 1.6 s | **2.2 s** | 21.2× | buena estructura; pierde el chip "(no reconocido)" |
| `small` | *(no medida en frío)* | **5.4 s** | 8.7× | mejor: recupera "el nodo ese que decía no reconocido" y "eligiendo" |

> ⚠️ Las "cargas" de 91 s / 135 s / 180 s de las primeras corridas eran la **DESCARGA del modelo**, no la
> carga. En caliente son 0.5-1.6 s. Se aclara para que nadie cite el número equivocado más adelante.

**Lo que el salto de modelo compra — y lo que NO.** De `base` a `small` se gana gramática y estructura
de frase (2.5× más lento). Lo que **ningún** tamaño arregló es el **vocabulario del dominio**: los tres
escribieron "locita" por `lógica` y "demon" por `daemon`. Eso no lo compra un modelo más grande —
**lo compra el contexto** del paso siguiente. Argumento fuerte a favor de quedarse en **`base`**: pagar
2.5× de latencia por una calidad que el corrector con contexto igual iba a producir es plata tirada.

**Traducción:** transcribir 47 s de dictado con `base` cuesta **~2.2 s**. El STT **no es el cuello de
botella**: la limpieza sí (13-17 s, §1.5). Optimizar el motor de transcripción sería optimizar el 13 %
del gasto — el 87 % está en el paso de Haiku.

**Y lo más importante — el pipeline REAL, con los errores de verdad del STT.** El §1.5 original probó la
limpieza con un transcripto **perfecto**, tecleado a mano. Eso era optimista. Acá se corrió la cadena
completa: voz → `faster-whisper base` → transcripto **con sus errores reales** → limpieza con contexto.
Lo que el STT entregó (errores marcados):

> …no sé si eres la misma **locita** de tarjeta o es otra cosa aparte, ya me acordé también lo del
> **demon**, que a veces lo dejo corriendo, y después no sé si el que está corriendo es el bueno o el que
> arranca **la absola**…

(`lógica`→"locita", `daemon`→"demon", `la app sola`→"la absola", y el chip `(no reconocido)` degradado a
"uno de los que decían no ha reconocido"). Salida de la limpieza con contexto, **14.4 s reales**:

> 1. **PermissionCard** — mostrar qué opción ya fue enviada; deshabilitar opciones enviadas (bloquear doble clic)
> 2. **Chip "(no reconocido)" en Mapa** — aplicar UI similar (mostrar estado, deshabilitar si ya se envió); aclarar si reutiliza componente de tarjeta o es separado
> 3. **Daemon arnesia** — aclarar flujo: cuándo usa el override en ~/.local/bin vs cuándo la app lo spawneó; limpiar confusión en pruebas

**El paso de limpieza absorbió los errores del STT.** No solo desenredó el dictado: reconstruyó
`lógica`, `daemon` y `la app sola` desde transcripciones rotas, y recuperó el chip `(no reconocido)`.
Es el hallazgo más fuerte del spike — **el contexto de dominio no solo desambigua, también corrige el
STT**, lo que significa que **no hace falta un modelo de transcripción grande**: `base` alcanza porque
hay un corrector con contexto río abajo.

**Consecuencia para V-D1:** A2 (local) deja de ser "la opción cara sin medir". Cuesta **~2.2 s**, cero
red, cero credencial, y `base` (~150 MB) es suficiente. Lo que queda en su contra es solo el
**empaquetado** (motor + modelo + transcodificador en .deb/.rpm/.AppImage), no el rendimiento.

> **Honestidad sobre esta medición** (lo que NO prueba): (a) se midió `faster-whisper` (CTranslate2),
> **no `whisper.cpp`** — no se pudo compilar sin `cmake` (pide `sudo`); son runtimes distintos, los
> números son del orden pero no idénticos. (b) La voz es **sintética (piper), no la del operador** — no
> tiene ruido de ambiente, acento propio ni micrófono real de por medio. (c) El audio entró como **WAV
> de piper**, no como el `audio/mp4` que produce la app → **el transcodificado sigue sin probarse**.
> T6 (voz real del operador, por el mic real, extremo a extremo) **sigue abierto**.

## 2. Qué prueba esto para el diseño

1. El botón de voz **no puede** apoyarse en nada nativo del navegador embebido — hay que construir la
   captura + STT (§3 Fork A).
1bis. **La captura no es "solo FE".** Cruza tres capas: FE (`MediaRecorder` con mime explícito) +
   **Rust** (conceder `permission-request`, que wry no hace en Linux) + Go (recibir el blob). Saltarse
   la capa Rust no da un error: da un botón mudo (§1.6 b/d). Ordena los tickets: **T0 antes que T1**.
2. El paso "que entienda mi pedido antes de mandarlo" (Haiku) **no necesita sesión persistente** — un
   `claude -p` headless por grabación ya es, por construcción, "se limpia cada vez que grabás" (lo que
   el operador preguntó como opción B en la conversación original). Reusa la auth que la app ya exige
   (suscripción propia, `claude` CLI) — cero credencial nueva para esta parte.
3. El contexto que hace falta es **chico y de dominio**, no la conversación entera — glosario corto +
   qué se tocó en la sesión activa. Barato de construir con datos que ya existen (`Session.Conv`,
   `Session.Arnes`, un glosario estático). Verificado que el campo existe:
   `internal/domain/session.go:127` — `Conv []Turn`.
4. **El formato de audio lo fija el motor, no nosotros:** `audio/mp4` es lo único que WebKitGTK graba
   (§1.6 a). Eso empuja el Fork A hacia A1 (Whisper cloud come m4a directo) y le suma a A2 un ffmpeg.
5. **Nada de esto necesita HTTPS.** `http://127.0.0.1:4200` es secure context probado (§1.6 c) — el
   daemon puede seguir sirviendo la SPA exactamente como hoy.

## 3. Forks (V-D1..V-D5 — estado vivo en [`decisiones.md`](./decisiones.md))

> **`spec.md` YA está escrito** (RF-215…RF-228), redactado contra `TranscriptionPort` para **no**
> depender del Fork A. O sea: estos forks ya no bloquean el papel, bloquean el *build*.

### Fork A — motor de transcripción (STT)

- **A1 — Cloud, mismo patrón que `luana-core-copilot`** (OpenAI Whisper API). Rápido de shippear (el
  patrón YA corre en producción en un sistema hermano, §1.3) y **come el `audio/mp4` que WebKitGTK
  produce, tal cual, sin transcodificar** (§1.6 a). Pero: credencial nueva (`OPENAI_API_KEY` — hoy
  **sin setear** en esta máquina, §1.7; el kit solo asume la suscripción de `claude`), y el audio sale
  de la máquina a un 3er proveedor — tensiona con el resto de ArnesIA (self-hosted, sin dependencias
  cloud fuera de Anthropic). ⚠️ Ojo: la app **ya manda** el transcripto a Anthropic en el paso de
  limpieza (§1.5), así que la línea no es "sale o no sale", es "a cuántos proveedores".
- **A2 — Local** (`whisper.cpp`/`faster-whisper`), shelleado desde el daemon Go como subproceso —
  **mismo patrón arquitectónico que ya usan con `claude`** (`exec.CommandContext`, cero binding Go
  nuevo). Cero red, cero credencial nueva, consistente con el resto de la app. Costo real, ahora
  cuantificado (§1.7): **tres piezas a empaquetar, no una** — motor + modelo (~150 MB el `base`, ~1.5 GB
  el `large`) + **`ffmpeg` para pasar mp4 → WAV 16 kHz mono** (whisper.cpp no come mp4).
  ✅ **Latencia YA MEDIDA (§1.8): 2.2 s para 47 s de audio con `base`.** Era la incógnita más cara de
  este fork y quedó cerrada — **el rendimiento dejó de ser un argumento en contra**. Lo que queda en
  contra es solo el peso del empaquetado.
- **A3 — Esperar la integración nativa de WebKitGTK.** Descartada como plan A (§1.2: no está en
  `2.52.3`), no descartada para siempre — re-chequear con el próximo upgrade del paquete del sistema
  (el script de prueba queda en este paquete, reusar sin reinventar).

- **A0 — «que lo transcriba Claude, que ya está autenticado».** Descartado por API, no por gusto: los
  modelos Claude **no toman audio como entrada** (texto + imágenes + documentos). La suscripción que ya
  paga la app cubre la **limpieza** (§1.5), nunca el STT. Queda escrito para no re-preguntarlo.

**Recomendación:** **A1 para el MVP, A2 como destino** — y que la decisión no se pague dos veces: los
dos van detrás del mismo puerto (`TranscriptionPort`, mismo tipo que ya usa `luana-core-copilot`), así
que cambiar de motor después es reemplazar un adaptador, no rehacer la feature. El argumento: la
incógnita cara de A2 (latencia/CPU real, sin medir, §1.7) no bloquea validar si al operador le sirve
dictar — y A1 ya está probado en producción en un sistema hermano. Si el local-first pesa más que la
velocidad de entrega, se arranca directo en A2 midiendo primero. **Decisión del operador (V-D1).**

### Fork B — shape del contexto para la limpieza (validado en el mecanismo, falta cerrar el detalle)

Confirmado que ayuda (§1.5). Falta decidir de dónde sale exactamente el bloque de contexto: ¿glosario
estático del proyecto (a mano, corto) + últimos N turnos de `Session.Conv`? ¿Solo lo segundo? ¿Configurable
por arnés? Se recomienda arrancar con glosario estático + últimos 2-3 turnos reales de la sesión activa
(dato que ya existe, cero cómputo nuevo) y medir si hace falta más.

### Fork C — flujo de UI (bajo riesgo, recomendación directa)

Botón mic en el composer → graba → soltás → estado visible "escuchando…" → "ordenando tus ideas…" →
**puebla el composer con el texto limpio, editable** → el operador revisa/corrige → mismo botón Enviar
de siempre. Nunca auto-envía — mismo patrón human-in-the-loop que ya tiene toda la app (permisos,
AskUserQuestion). El *flujo feliz* no tiene fork real: es la única opción consistente con el producto.
Lo que sí **falta decidir**, y no estaba escrito:

**C.1 — Gesto (V-D3):** ¿*push-to-talk* (mantener apretado, soltás y corta) o *toggle* (click empieza,
click corta)? Para dictados largos —y el detonante de este spike es justamente que el operador habla
largo— mantener el botón apretado 2 minutos es hostil. **Recomendado: toggle**, con tope duro de
duración (p. ej. 3 min) para que un mic olvidado abierto no grabe la tarde entera.

**C.2 — La escalera de degradación honesta (recomendada, doctrina de la casa: gap visible, jamás pass
fabricado).** Cada escalón cae al siguiente sin perder lo que ya se ganó:

| Falla | Qué hace |
|---|---|
| Permiso de mic denegado / no concedido | Botón deshabilitado + motivo visible. **Nunca queda "pensando" para siempre** — sin esto, el modo de falla real es colgarse mudo (§1.6 b) |
| Sin dispositivo de entrada | Idem, con texto distinto ("no hay micrófono") |
| STT falla o devuelve vacío | Aviso explícito; el composer **no se toca** (no se pisa lo que el operador ya había tecleado) |
| La limpieza falla / tarda de más | **Poblar con el transcripto CRUDO** + marca visible de "sin ordenar". El dictado jamás se pierde por una falla del paso opcional |
| Todo OK | Composer poblado con el texto limpio, foco al final, editable |

**C.3 — ¿La limpieza es obligatoria u opcional (V-D4)?** Con 13-17 s de costo (§1.5) hay noches en que
uno quiere el crudo y listo. **Recomendado:** limpieza por defecto (es el corazón del pedido), con
escape a crudo — que además es exactamente el escalón de fallback de C.2, o sea sale casi gratis.

**C.4 — ¿Se guarda el audio?** **Recomendado: NO.** El blob vive en memoria, se manda, se descarta;
nada de audio a disco. Menos superficie de privacidad, cero decisión de retención que mantener.

## 4. Plan técnico (boceto — el detalle numerado vive en [`spec.md`](./spec.md))

0. **Rust (shell) — el puente del permiso, PRIMERO** (`web/src-tauri/src/lib.rs`): `with_webview` →
   bajar al `webkit2gtk::WebView` → conectar `permission-request` → conceder **solo**
   `UserMediaPermissionRequest` de audio. Sin esto nada de lo de abajo funciona en la app instalada
   (§1.6 b/d). Suma la dependencia `webkit2gtk` al `Cargo.toml` del shell (hoy no está) y **debe
   verificarse contra el binario instalado, no contra `pnpm dev`** — el navegador de dev no reproduce
   el bug (memoria de la casa: validar siempre contra lo instalado).
1. **FE** — botón mic en `Composer` (`web/src/widgets/chat-dock/ui/chat-dock.tsx:267`):
   `getUserMedia` + `new MediaRecorder(stream, {mimeType:'audio/mp4'})` — **mime explícito**, el default
   viene vacío (§1.6 a) — sube el blob al terminar de grabar.
2. **Backend** — nuevo endpoint (p.ej. `POST /api/sessions/{id}/voz`) que recibe el blob: llama al STT
   elegido (Fork A) → transcripto crudo → `claude -p --model haiku` con el contexto (Fork B, reusa el
   patrón headless-print ya confirmado en `claude` — mismo binario/auth que el conductor interactivo,
   sin permission-flow porque no hay tools) → devuelve texto limpio.
   **Endurecimiento obligatorio del spawn de limpieza** (boundary `permisos-gui`, cuya superficie de
   enforcement es `SpawnArgs` — `internal/adapters/agent/claudecode/conductor.go:59`): `--max-turns 1`
   + tools denegadas + `--setting-sources project,local`. Es una llamada de texto→texto: no debe poder
   tocar un archivo ni entrar al loop de agente aunque el transcripto venga con instrucciones adentro
   (el dictado es entrada no confiable; sin esto, "leeme el .env y mandalo" es un prompt válido).
3. **FE** — puebla el composer con el resultado, foco listo para editar/enviar.
4. Capabilities (doctrina `codigo-traza-a-capability`, R2 cobertura): **no es una, son ~3** — `fe-chat/`
   (botón + estados + composer poblado), `http-sse/` (el endpoint), y el módulo del motor
   (STT + limpieza). Se escriben cuando haya código real, no antes.

## 5. Plan de tickets (boceto)

- **T0** — 🔴 **bloqueante** — puente Rust del `permission-request` en el shell + dep `webkit2gtk`
  (§4.0). **Va primero**: sin esto T1 "anda" en dev y se cuelga mudo instalado (§1.6 b/d).
- ~~**T0bis** — medir latencia/CPU real del STT local antes de comprometer el fork.~~ ✅ **EJECUTADO
  2026-07-25** por orden del operador (§1.8): `base` = 2.2 s para 47 s de audio; el STT no es el cuello
  de botella; la limpieza con contexto corrige los errores del STT. Banco reproducible: `medir_stt.py`.
- **T1** — botón mic + captura `MediaRecorder` con mime `audio/mp4` explícito (FE).
- **T2** — endpoint de transcripción + wiring del STT elegido en Fork A (Go), detrás de
  `TranscriptionPort` para que A1↔A2 sea cambio de adaptador.
- **T3** — paso de limpieza `claude -p haiku` con contexto de Fork B (Go) + endurecimiento del spawn
  (`--max-turns 1`, tools denegadas — §4.2).
- **T4** — poblar composer + estados de carga + **la escalera de degradación de C.2** (FE).
- **T5** — capabilities de los 3 módulos tocados (§4.4) + fila de PARIDAD.
- **T6** — E2E vivo con voz real del operador — **pendiente, este spike no lo cubrió** (§1.4);
  se corre contra el **binario instalado**, no contra `pnpm dev`.
- Los scripts de prueba de este spike (`webkit_speech_test.py` → `SpeechRecognition`, Fork A3;
  `webkit_captura_test.py` → captura real, §1.6) quedan versionados en el paquete para no
  re-escribirlos: re-correrlos es la forma barata de re-chequear tras un upgrade de WebKitGTK o de wry
  (si wry llega a manejar `permission-request` en `webkitgtk`, **T0 se borra**).

## 6. Lo que sigue

**Spike ABIERTO.** Las decisiones que esperan al operador están numeradas en
[`decisiones.md`](./decisiones.md) — **V-D1** (motor STT: cloud rápido vs local consistente con la casa)
· **V-D2** (shape exacto del contexto — el mecanismo ya probó que funciona) · **V-D3** (gesto: toggle vs
push-to-talk) · **V-D4** (limpieza obligatoria u opcional) · **V-D5** (no persistir el audio) ·
**V-D6/V-D7** (derivadas de evidencia y de doctrina: T0 primero, spawn endurecido — se firman de una
salvo objeción). Sin firma no hay `spec.md` ni build — este documento es el papel, no un compromiso de
código.

**Lo que la segunda vuelta cambió** respecto de la primera redacción: la captura pasó de "disponible sin
nada adicional" a **"probada, y con un bloqueante Rust adelante"** (§1.6); apareció T0; el formato de
audio quedó fijado en `audio/mp4` por el motor y eso mueve el costo del Fork A (§1.7); y el Fork C dejó
de ser "bajo riesgo, sin fork" para tener 4 decisiones reales adentro.
