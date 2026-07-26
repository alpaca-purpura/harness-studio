# PARIDAD — Dictado por voz en el composer (RF-215…RF-230)

> Paquete `2026-07-25-spike-voz-dictado`. Round-trip: **cada RF del `spec.md` ⇒ código real + su
> verificación**; **cada verificación ⇒ corrida de verdad, con su salida**. Lo que no se pudo probar
> queda `sin-check` con el motivo — jamás un pass fabricado (doctrina de honestidad del repo).
>
> **Fecha de la corrida: 2026-07-26.** Firma 🧑‍⚖️ del gate final: **PENDIENTE**.

## Cómo se verificó (comandos reales)

| Suite | Comando | Resultado |
|---|---|---|
| Rust (shell) | `cargo test --lib` (en `web/src-tauri`) | **4 passed · 0 failed** |
| Go (todo el árbol) | `go test ./...` | **verde** (0 fail, 0 error) |
| Go vet | `go vet ./...` | limpio |
| Fitness de arquitectura | `go test ./docs/architecture/fitness/` | verde (incl. go-arch-lint sobre el árbol propio) |
| Conformance | `go run ./cmd/arnesia conformance --todo` | `279 checks · pass 66 · fail 0 · error 0 · deferred 213` |
| Verify FE | `npm run verify` (typecheck · biome · depcruise · steiger · stylelint) | **verde** |
| Stories del dictado | `npx vitest run --project=storybook src/widgets/chat-dock` | **19 passed** (2 archivos: dictado + permission-card) |
| Stories — todo el árbol | idem sobre `src/widgets` + `src/entities src/features src/shared` | **30/30 archivos · 254 tests · 250 pass · 4 fail PREEXISTENTES** (ver §Hallazgos) |

## ✅ CADENA COMPLETA PROBADA CON VOZ REAL (2026-07-26)

Se instaló un motor STT de verdad (venv aislado sin sudo, `whisper-ctranslate2` + `piper` para
sintetizar la voz — el patrón del spike) y se corrió **el pipeline entero por el daemon real**.
Esto es lo que el spike §1.8 nunca hizo: ahí se midió un script de Python, acá corre **el
adaptador Go de este paquete**.

**Dictado hablado** (dicho enredado a propósito):

> «Este eh... quiero que le cambies el color al pip de la tarjeta esa de permisos, ponele el de
> warning, y de paso fijate la lógica del daemon que arranca solo, porque la app sola no
> levanta bien.»

**Transcripto CRUDO que devolvió el motor** (con sus errores reales):

> «Este quiero que le cambies el color al pip de la tarjeta de esa de permisos, ponele el de
> **warming** y de paso fijate la lógica del **danon** que arranca solo porque la app sola no
> levanta bien.»

**Respuesta del endpoint** (`POST /api/sessions/{id}/dictado`, **13.4 s** de punta a punta):

```json
{
  "texto": "Cambiar el color del pip de la PermissionCard al de warning. Revisar la lógica del daemon que arranca automáticamente porque la app no se inicia bien por sí sola.",
  "estado": "limpio",
  "motor": "whisper-ctranslate2"
}
```

**Qué queda probado con esto, y no por un test con fakes:**

| Claim del spike | Evidencia |
|---|---|
| el contexto **repara errores del STT** | «warming» → `warning` · «danon» → `daemon` (`daemon` está literal en `glosarioGlobal`) |
| el contexto **resuelve referencias ambiguas** | «la tarjeta de esa de permisos» → **`PermissionCard`** |
| RF-222 marca el estado | `"estado": "limpio"` |
| RF-224 la limpieza recibe contexto útil | sin glosario ninguna de las 3 correcciones era posible |
| el modelo chico alcanza (V-D1) | todo lo anterior con `base` |

Latencia STT del adaptador, corridas en caliente: **1.5 s – 3.7 s** para ~14 s de audio —
consistente con los 2.2 s / 47 s de §1.8. Confirma que el cuello es la limpieza.

**Reproducible:** `internal/adapters/stt/local/local_e2e_test.go` quedó versionado. Skipea por
defecto (exige motor + modelo, que CI no tiene) y se corre con
`ARNESIA_STT_E2E_AUDIO=<wav> go test ./internal/adapters/stt/local/ -run E2E -v`. Que skipee es
honesto; fingir que pasó no lo sería.

### 3 bugs REALES que solo aparecieron al correr los binarios

La primera versión del adaptador estaba escrita contra lo que yo suponía que hacían los CLI.
Los tres con fakes pasaban en verde:

1. **Se leía `stdout` en vez del archivo.** `whisper-ctranslate2` imprime en stdout
   `Detected language 'Spanish'…` / `Transcription results written to '<dir>' directory` y deja
   el transcripto en un `.txt`. **El composer se habría poblado con la charla del CLI en vez de
   con el dictado del operador.** Fix: `dondeSale` (`enStdout` | `enArchivoTxt`) +
   `leerTranscripto`. Test de regresión: `TestTranscribirLeeElArchivoNoStdout`, cuyo stub imita
   ese ruido literal.
2. **`--output_dir -` no existe.** No es «stdout»: habría creado un directorio llamado `-`.
3. **Sin `--device cpu` el CLI intenta CUDA y EXPLOTA:**
   `RuntimeError: Library libcublas.so.12 is not found or cannot be loaded` en cualquier máquina
   sin toolkit de NVIDIA. Fix: `--device cpu --compute_type int8` (que además es la config
   medida en §1.8). Test: `TestTranscribirFijaCpuYInt8`.

Y uno más de la misma familia: **`faster-whisper` no expone ejecutable** (es librería —
verificado instalándola). Mi tabla lo listaba como binario a buscar y el hint le decía al
operador instalar algo que no habría servido. El CLI real es `whisper-ctranslate2`. Test:
`TestElHintDeInstalarSoloOfreceBinariosQueExisten`.

### El bug que habría matado T6 en la app instalada

`exec.LookPath` usa el `$PATH` **del proceso**, y **un proceso lanzado desde el `.desktop` no
hereda `~/.profile` ni `~/.bashrc`**. El motor que el operador instala —y que cualquier terminal
ve— habría quedado **invisible para la app instalada**: botón gris diciendo «falta el motor» con
el motor ahí puesto.

Es el **mismo root cause** que el repo ya había arreglado para la toolchain del self-update
(`selfupdate.pathAumentado`). Se replicó el patrón en el adaptador: `candidatosPATH()`
(`~/.local/bin` · `~/.local/share/arnesia-stt/bin` · `/usr/local/bin`) + `lookPathAumentado()`,
y **todo se ejecuta por ruta absoluta** — correr por nombre pelado volvería a depender del PATH
del proceso.

**Probado de las dos formas:** el E2E pasa con el venv en `$PATH` **y con `$PATH` limpio**
(escenario del launcher gráfico). Tests: `TestPathAumentadoEncuentraElMotorFueraDelPathHeredado`
· `TestPathAumentadoNoInventaLoQueNoEsta`.

### Footgun del `Makefile` (arreglado)

`make dev-sync` reportaba **`make: *** [Makefile:76] Terminado` en cada corrida** aunque el sync
hubiera terminado bien: el `pkill -f "<path> serve"` del final **se auto-mataba**, porque
`pkill -f` matchea líneas de comando completas — incluida la del shell que corre esa misma
receta. El `|| true` no ayudaba: el shell recibe SIGTERM, no sale con código. Fix con el truco
del corchete (`[a]rnesia`), verificado: ahora sale `EXIT=0` con su OK.

Importa más de lo que parece: un paso documentado que **siempre miente sobre su resultado**
enmascara la próxima falla de verdad.

## Verificación EN VIVO del backend (daemon real, no un test)

Los tests usan fakes, así que la superficie se sondeó además contra el binario compilado
(`go build ./cmd/arnesia` → `serve --addr 127.0.0.1:4272`). Salidas literales:

| Sonda | Respuesta real | Lectura |
|---|---|---|
| `GET /healthz` | `{"status":"ok"}` | el daemon levantó |
| `GET /api/dictado/disponibilidad` **sin motor** | `{"disponible":false,"motivo":"no hay motor de transcripción instalado","instalar":["whisper-cli","whisper-ctranslate2"]}` | **la degradación honesta funciona de punta a punta**; era el estado real de la máquina antes de instalar el motor |
| `GET /api/dictado/disponibilidad` **con motor** | `{"disponible":true,"motor":"whisper-ctranslate2"}` | el mismo endpoint, después de instalar el venv — sin reiniciar nada más |
| `POST /api/sessions/no-existe/dictado` | `404` | sesión inexistente |
| `POST /api/sessions/{real}/dictado` (sin motor) | `503` · `{"error":"dictado: no disponible: no hay motor de transcripción instalado"}` | capacidad ausente, no falla interna |
| `POST …/dictado` cuerpo vacío | `400` · `{"error":"la grabación vino vacía"}` | 0 bytes no es un dictado |
| `POST …/dictado` 25 MB (tope 24) | `413` | la cota de transporte muerde |

**Un arreglo salió de esta corrida.** El primer sondeo devolvió **500** para «no hay motor». Es
impreciso: «no puedo transcribir» es una **capacidad ausente**, no un error interno, y un 500 mandaría
al operador a buscar un bug donde falta un `apt install`. Se agregó el sentinel
`usecase.ErrDictadoNoDisponible` → **503**, con su test (`TestPostDictadoSinMotorEs503NoEs500`), y se
re-verificó en vivo. Mismo criterio que el 503 ya existente de `validarMarketplace` («no puedo mirar»
≠ «tu url está mal»).

## Tabla RF → código → verificación

| RF | Qué exige | Código | Verificación | Estado |
|---|---|---|---|---|
| **RF-215** | El shell concede el permiso de mic del WebView, acotado a audio | `web/src-tauri/src/lib.rs#conceder_permiso_de_microfono` · `#decidir_permiso` · `#concede_captura` · dep `webkit2gtk` feature `v2_8` en `Cargo.toml` | `concede_solo_captura_de_audio` · `rechaza_permisos_que_no_son_de_medios` · `rechaza_la_camara_aunque_venga_junto_con_el_microfono` · `rechaza_media_que_no_pide_audio` — **4/4 pass** | ✅ regla probada · ⚠ **loop vivo = T6, ABIERTO** |
| **RF-216** | Botón de mic en el composer + estado «escuchando» visible | `dictado-button.tsx#DictadoButton` · `#VoiceBar` · cableado en `chat-dock.tsx#Composer` | stories `Reposo` · `Escuchando` | ✅ |
| **RF-217** | Toggle: el mismo botón corta; el mic se libera | `dictado-store.ts#cortar` · `#soltar` (para todos los tracks) | story `Escuchando` (el botón `Dictar` desaparece y aparece `Cortar el dictado`) · `CortarDisparaElFlujo` | ✅ superficie · ⚠ liberación real del track = T6 |
| **RF-218** | Tope duro; para sola, CONSERVA lo grabado, avisa | `dictado-store.ts` `TOPE_MS` + `R.tope` · `cortadoPorTope` · `DictadoAviso` | stories `CercaDelTope` (contador en `text-warn`) · `CortadoPorTope` (avisa + «no se descarta nada») | ✅ |
| **RF-219** | Cancelar sin mandar; composer queda como estaba | `dictado-store.ts#cancelar` (aborta el POST y el recorder) | story `Escuchando` (botón `cancelar` presente y distinto de cortar) | ✅ superficie · ⚠ descarte real del blob = T6 |
| **RF-220** | mimeType explícito; sin soporte, deshabilitado con motivo | `dictado-store.ts` `MIME = "audio/mp4"` + guarda `MediaRecorder.isTypeSupported` | `TestPostDictadoAsumeMp4SinContentType` (el default vacío del motor real) · `TestTranscribirAceptaElMimeConParametros` | ✅ |
| **RF-221** | El audio no se persiste | `dictado-store.ts` (sin `createObjectURL`, sin reproductor) · `local.go#escribirTemporal` con `defer limpiar()` | **`TestAdapterLocalBorraElTemporal`** (lista el dir tras transcribir: 0 archivos) · `TestTranscribirTranscodificaParaMotoresQueSoloComenWav` (ni el mp4 ni el wav sobreviven) | ✅ |
| **RF-222** | Endpoint de dictado; la respuesta dice LIMPIO o CRUDO | `http/dictado.go#postDictado` · `router.go#NewHandler` | `TestPostDictadoDevuelveLimpioOCrudo` (los dos casos) · `…RechazaAudioDemasiadoGrande` (413) · `…RechazaCuerpoVacio` (400) · `…404SiLaSesionNoExiste` · `…NoSeEntendioNadaNoEs500` (422) | ✅ |
| **RF-223** | El motor detrás de un puerto; sin motor, no-disponible con motivo | `ports/dictado.go#TranscriptionPort` · `stt/local/local.go#Adapter` `#Detectar` `#lookPathAumentado` `#leerTranscripto` · `http/dictado.go#getDisponibilidad` | `TestAdapterLocalEligeElPrimerMotorDelPath` (3 casos) · `TestDisponibleSinMotorDiceMotivoYQueInstalar` · `TestDisponibleAvisaQueFaltaFfmpeg` · `TestWhisperCppSinModeloNoEsUsableYLoDice` · `TestMotorInstaladoPeroIncompletoNoDiceQueNoHayNada` · `TestElHintDeInstalarSoloOfreceBinariosQueExisten` · `TestTranscribirLeeElArchivoNoStdout` · `TestTranscribirFijaCpuYInt8` · `TestPathAumentado*` (2) · `TestGetDisponibilidadDiceElMotivo` · **`TestE2ETranscribeConMotorReal` (motor REAL)** · story `SinMotorDeTranscripcion` | ✅ **+ E2E con voz real** |
| **RF-224** | La limpieza recibe glosario global + últimos 2-3 turnos | `usecase/dictado_service.go#contextoDeLimpieza` · `#turnosRelevantes` · `glosarioGlobal` | `TestContextoDeLimpiezaRecortaAUltimosTurnos` (y que el 4º NO entra, y el orden es cronológico) · `…LlevaElGlosario` · `…IgnoraLaActividad` · `…RecortaTurnosGigantes` | ✅ |
| **RF-225** | El spawn de limpieza va endurecido | `claudecode/limpieza.go#LimpiezaArgs` · `#PromptLimpieza` | **fitness** `docs/architecture/fitness/dictado_spawn_test.go`: `TestSpawnDeLimpiezaVaEndurecido` · `TestLimpiezaNiegaTodaHerramienta` (11 tools) · `TestLimpiezaNoPreAprubaNadaNiEscapaElSandbox` · `TestLimpiezaCorreHeadless` · `TestPromptDeLimpiezaNoContestaElPedido` — **5/5 pass** | ✅ |
| **RF-226** | Composer poblado, editable, foco al final, NO auto-envía | `chat-dock.tsx#Composer` → `recibirDictado` (`setValue` + foco al final; **jamás `submit()`**) | story `Limpio` (el enviar queda habilitado pero no se disparó; el texto está en el campo) | ✅ |
| **RF-227** | Escalera de degradación con motivo visible | `dictado-store.ts#motivoDeGetUserMedia` · `DictadoAviso` · `usecase` escape a crudo | stories `Crudo` · `FalloDeEtapa` (no pisa lo tecleado) · `PermisoDenegado` · `SinMotorDeTranscripcion` · `NoSeSabeTodavia` · Go: `TestDictarCaeACrudoSiLaLimpiezaFalla` · `TestSinLimpiadorDevuelveCrudoHonesto` | ✅ |
| **RF-228** | Nada tapa un cuelgue: sale de la carga y nombra la etapa | `dictado-store.ts` `ETAPA_LABEL` (3 etapas con nombre) · `Fallo.etapa` · tope en el usecase | stories `Transcribiendo` · `Ordenando` · `FalloDeEtapa` (nombra la etapa) · Go: `TestDictarMarcaElCorteDeTiempoDistintoDeLaFalla` | ✅ |

## Trazabilidad de capabilities (doctrina R1/R2/R4)

4 capabilities nuevas — **una más que las 3 que pedía el `spec.md`**: RF-215 es código en `web/src-tauri`
y R2 exige que TODO archivo de código trace a un capability, así que el puente Rust se reclama aparte.

| Capability | cap_num | Estado (GENERADO por R4) |
|---|---|---|
| `arnesia.tauri.permiso-de-microfono` | CAP-112 | `vivo` |
| `arnesia.usecases.dictado-transcribir-y-ordenar` | CAP-113 | `vivo` |
| `arnesia.http-sse.superficie-dictado` | CAP-114 | `vivo` |
| `arnesia.fe-chat.dictar-en-el-composer` | CAP-115 | `vivo` |

Cifras: capabilities **111 → 115** · `arch/` **21 → 22 boundaries** (componente `stt` nuevo en
`.go-arch-lint.yml`) · conformance `--todo` **266 → 279 checks**, `pass 57 → 66`, **fail 0**.

## Lo que los fitness tests cazaron durante la construcción

Se registran porque son la prueba de que el arnés muerde, no adorno:

1. **go-arch-lint** — `cmd` importaba `internal/adapters/stt/local` sin que el componente existiera en
   el grafo permitido. Se declaró `stt: mayDependOn: [ports]` y se lo agregó a la allow-list de `cmd`.
2. **R1 punteros** — un capability apuntaba a `internal/usecase/ports/transcription.go`, ruta que
   nunca existió (el puerto vive en `internal/ports/dictado.go`).
3. **R1 símbolo** — `router.go#Router` no resuelve: el símbolo real es `NewHandler`.
4. **R2 cobertura** — `internal/adapters/agent/claudecode/limpieza.go` quedó sin capability que lo
   reclamara.

## Hallazgos abiertos (VISIBLES, no tapados)

### 4 fallos preexistentes en `session-rail` — NO son del dictado

`src/widgets/session-rail/ui/new-session-picker.stories.tsx` falla 4 stories (`Caso Simple`,
`Caso Ambiguo`, `Colision`, `Cancelar`) con:

```
Elements must meet minimum color contrast ratio thresholds (color-contrast)
Element has insufficient color contrast of 3.76 (foreground color: #c96a2e,
background color: #ffffff, font size: 7.5pt (10px), font weight: normal). Expected 4.5:1
<span class="text-warn">historial de cerradas: Failed to fetch</span>
```

- **Qué pasa:** `new-session-picker.tsx:273` pinta ese `text-warn` cuando falla el fetch del historial
  de cerradas. Sin daemon en `:4200` el fetch falla, el span aparece, y axe marca `--warn` (`#c96a2e`)
  sobre `--card` (`#ffffff`) en tema claro a **3.76:1**, bajo el mínimo 4.5.
- **Por qué NO es de este paquete:** ni ese componente ni esa story importan nada del dictado
  (verificado por grep). El token y el componente son anteriores.
- **Es un hallazgo real igual:** `--warn` a 10px sobre `--card` no pasa contraste en tema claro. Eso
  toca `web/tokens/base.tokens.json`, o el tamaño/peso de ese span. **Fuera de alcance acá** — se
  registra para el BACKLOG en vez de arreglarse al voleo dentro de un paquete de voz.

### T6 — el tramo del MICRÓFONO, ABIERTO (lo demás ya cerró)

Después de la corrida del 2026-07-26, T6 se **encogió** a un solo tramo. Ya está probado con voz
real y por el daemon real: `audio → STT → limpieza → respuesta del endpoint`.

**Lo que sigue sin probar, y ningún test reemplaza:** el tramo
`micrófono real → MediaRecorder → POST`, contra el **binario instalado**.

- Los tests de Rust prueban **la regla** de concesión (`concede_captura`), no el enganche de la
  señal GTK — eso exige un WebView vivo con display.
- Las stories prueban **la superficie** con el store fijado a mano; no abren un micrófono.
- El E2E de Go entra por `[]byte` de un `.wav`: no pasa por `getUserMedia` ni por `audio/mp4`.

**Por qué tiene que ser contra lo instalado:** RF-215 existe porque `getUserMedia` **no rechaza —
queda pendiente para siempre** sin el puente, y eso *solo* se reproduce ahí; el dev server concede
por su cuenta, así que un pass en `pnpm dev` no probaría nada (norma de la casa).

**Pasos, con todo ya preparado** — instalador en `instaladores/v0.2.20/`, `~/.local/bin/arnesia`
sincronizado (`make dev-sync`), y motor STT instalado en esta máquina: instalar el `.deb` → abrir
la app → abrir un frente → botón de mic → hablar → cortar → confirmar que el composer se puebla y
**no se auto-envía**.

**Sub-gap propio de este tramo:** el **`audio/mp4` real nunca se transcribió**. Todo lo medido usó
WAV. Si `whisper-ctranslate2` no digiere el mp4 de WebKitGTK, el adaptador lo va a decir como
falla de etapa (RF-228) — pero es una incógnita ABIERTA, no algo verificado.

### T7 — qué motor STT se empaqueta, ABIERTO (con evidencia nueva)

**Lo que cambió el 2026-07-26:** ya hay medición propia de `whisper-ctranslate2` corrido por **el
adaptador de este repo** (1.5–3.7 s para ~14 s de audio, modelo `base`, CPU int8), y quedó
documentado que **exige `--device cpu`** o explota sin CUDA.

**Lo que sigue faltando** para cerrar T7: medir `whisper.cpp` (pide `cmake`/sudo) y medir ambos
sobre el **`audio/mp4` real**. Hasta entonces el instalador **no promete STT**, y en una máquina
limpia el estado honesto por defecto es «Falta el motor de transcripción».

**En ESTA máquina el motor ya está instalado** — venv en `~/.local/share/arnesia-stt`, sin sudo,
borrable con `rm -rf` — así que el botón va a aparecer habilitado.

## Firma

- [ ] 🧑‍⚖️ **Gate de PARIDAD** — el operador verifica T6 en vivo contra el binario instalado y firma
      acá. Sin eso, este paquete queda «construido y verificado por tests», **no** «probado en vivo».


---

# Ronda 3 — RF-229 / RF-230 (2026-07-26)

> Detonante: el operador probó el dictado en su instalación **v0.2.20** y salió
> `«Falló escuchando: No se grabó nada.»`. Se reprodujo contra el motor real y se diagnosticó con
> `GST_DEBUG`. Firma 🧑‍⚖️ de esta ronda: **PENDIENTE** (falta el tramo E2E contra el binario
> instalado — el mismo gate que ya estaba abierto para el micrófono).

## Evidencia del diagnóstico (corridas reales, no razonamiento)

| Qué se probó | Comando | Salida |
|---|---|---|
| `MediaRecorder` sin timeslice (**lo de v0.2.20**) | harness GTK propio, WebKitGTK 2.52.3 | `onstart:true · state:recording · eventos_data:1 · tam_por_evento:[0] · blob_bytes:0 · onerror: nunca` |
| idem + `start(250)` | idem | `eventos_data:1 · tam_por_evento:[0] · blob_bytes:0` |
| idem + `audioBitsPerSecond:128000` | idem | `{"n":1,"bytes":0}` |
| **WebAudio (`ScriptProcessor`)** | idem | `sampleRate:44100 · bloques:37 · muestras:151552 · segundos:3.44 · pico:0.0787` ✅ |
| ¿el mic entrega audio? | `parec … --rate=16000` 4 s + medición | `frames 64000 · RMS 5716 · MAX 25131` ✅ |
| ¿están los encoders? | `gst-launch-1.0 pulsesrc ! voaacenc ! mp4mux ! filesink` | `gst.mp4` de **16905 bytes** ✅ |
| causa raíz | `GST_DEBUG="*:2,webkit*:6" …` | `restriction caps audio/x-raw, rate=(int)0` → `encodebasebin: Couldn't find a compatible stream profile` → `basesrc: streaming stopped, reason not-linked (-1)` → `Transfering 0 encoded bytes` |

**Veredicto:** falla del motor, no del código. `isTypeSupported('audio/mp4')` devuelve `true` y miente;
el `error` que el pipeline postea en el bus de GStreamer WebKit no lo propaga a JS, así que
`MediaRecorder.onerror` nunca dispara.

## Fila por fila

| RF | Qué exige | Código | Verificación | Estado |
|---|---|---|---|---|
| **RF-229** | captura sin `MediaRecorder`, WAV a 16 kHz armado en el FE | `web/src/shared/lib/wav.ts` · `web/src/shared/store/dictado-store.ts#empezar/cortar` | `vitest run src/shared/lib/wav.test.ts` | ✅ **12 passed** |
| RF-229 | el daemon no necesita `ffmpeg` para leerlo | `internal/adapters/stt/local/local.go` (`.wav` saltea el transcodificado) | `ARNESIA_STT_E2E_AUDIO=mic.wav go test ./internal/adapters/stt/local/ -run E2E -v` | ✅ el motor **corrió** sobre el WAV (`whisper-ctranslate2`, 3.63 s) sin pedir `ffmpeg`. Transcripto vacío **porque el WAV era ambiente sin voz**, no por el formato |
| RF-229 | «no entregó ni un bloque» ≠ «entregó silencio» | `dictado-store.ts#cortar` | story `CortarDisparaElFlujo` | ✅ **13 passed** en `dictado-button.stories.tsx` |
| **RF-230** | el daemon escribe a `~/.arnesia/logs/arnesia.log` | `internal/adapters/logfile/` · `cmd/arnesia/main.go#activarLog` | `go test ./internal/adapters/logfile/` | ✅ **6 passed** (rota · appendea · concurrente · post-Close) |
| RF-230 | stderr texto + archivo JSON | `cmd/arnesia/main.go#dosDestinos` | daemon real en `:4271` | ✅ archivo: `{"level":"ERROR","msg":"diagnostico",…,"detalle":{"bloques":37,…}}` · stderr: `level=ERROR msg=diagnostico …` |
| RF-230 | el FE puede escribir en ese log | `internal/adapters/transport/http/diagnostico.go` · `web/src/shared/lib/diagnostico.ts` | `curl -X POST /api/diagnostico` contra el daemon real | ✅ **HTTP 204** y el `detalle` anidado completo en el archivo |
| RF-230 | cada dictado deja su línea | `internal/adapters/transport/http/dictado.go` | `curl` de un dictado a sesión inexistente | ✅ `{"msg":"dictado","sesion":"no-existe","mime":"audio/wav","bytes":128044,"ms":0,"err":…}` |
| RF-230 | acotado y no rompe nada | idem | `go test ./internal/adapters/transport/http/ -run Diagnostico` | ✅ **7 passed** (tope de cuerpo · claves recortadas · sin evento → 400 · vivo sin dictado cableado) |

## Suites tras el cambio

| Suite | Resultado |
|---|---|
| `go build ./...` · `go test ./...` | **verde** |
| `go test ./docs/architecture/fitness/` | **verde** (incl. go-arch-lint con el componente `logfile` nuevo y R1/R2/R3/R4 de capabilities) |
| `pnpm run verify` (typecheck · biome · depcruise · steiger · stylelint) | **verde** |
| stories del dictado | **13 passed** |
| stories — todo el árbol | **5 fail**, de los cuales **4 son los PREEXISTENTES** de `new-session-picker` (contraste a11y en `.text-warn`, archivo sin cambios locales) y el 5º era `CortarDisparaElFlujo`, **actualizado** a la conducta nueva |

## ⛔ Lo que NO se probó (gap visible, no pass fabricado)

- **La cadena entera en la app instalada.** El diagnóstico y la captura se reprodujeron en un harness
  GTK propio contra **la misma versión de WebKitGTK** (2.52.3), mismo origin y mismo mime — no contra
  el `.deb` corriendo. Es exactamente el tramo que el gate del paquete ya tenía abierto.
- **Transcripción de una captura hecha por el código nuevo.** Se probó que el motor come el WAV y que
  el encoder produce WAV válido; falta una grabación real de punta a punta con voz del operador.
