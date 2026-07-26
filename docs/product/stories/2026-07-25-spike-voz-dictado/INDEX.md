# Spike · Dictado por voz en el composer (STT + limpieza con contexto)

> `tipo: spike` — investigación acotada. Estados: abierto · investigando · cerrado. **Vig: activo ·
> DECISIONES FIRMADAS 🧑‍⚖️ 2026-07-25 (2ª ronda) · en construcción.** Detonante: el operador quiere
> hablarle a ArnesIA en vez de escribir — grabar, transcribir, y que un paso "entienda" el pedido (se
> enreda al hablar, pero el contexto de dominio lo desenreda).

## Pregunta del spike

¿Con qué armamos un botón de voz en el composer, dado el motor real de la app (Tauri + WebKitGTK en
Linux), y qué contexto necesita el paso de limpieza para de verdad ordenar un dictado enredado en vez
de solo sacarle las muletillas?

## Hallazgo (probado hoy, con evidencia real — ver `spike-spec.md` §1)

- ❌ **WebKitGTK nativo (`SpeechRecognition`):** probado en vivo contra el motor real (`libwebkit2gtk-4.1
  2.52.3`, WebView real sobre `DISPLAY :0`) — ausente. Hipótesis de la research anterior (integración
  nativa Whisper.cpp) NO está en esta versión empaquetada. Plan A descartado.
- ✅ **Captura de audio — PROBADA de verdad** (§1.6, 2ª vuelta): `getUserMedia` abre el **micrófono
  real** (`state: live`) en la WebView real, y `MediaRecorder` existe. `http://127.0.0.1:4200` (el origin
  del daemon) es secure context → **sin HTTPS ni certificados**.
- ⛔ **…pero con un bloqueante adelante:** `wry 0.55.1` **no maneja `permission-request` en Linux** (sí en
  macOS/Android) → sin un puente Rust, `getUserMedia` **no falla: se cuelga mudo para siempre**. El botón
  andaría en `pnpm dev` y moriría en la app instalada. ⇒ **ticket T0, antes que el FE.**
- ⚠️ **El formato lo fija el motor:** `MediaRecorder` solo soporta **`audio/mp4`** (webm/opus/ogg/wav =
  false) y su mime por defecto viene vacío. Mueve el costo del Fork A: A1 (cloud) lo come tal cual; A2
  (local) suma **ffmpeg** para transcodificar.
- 📊 **Fork A2 (local) MEDIDO** (§1.8, por orden del operador "medir antes de decidir"): transcribir
  **47 s de voz española real** cuesta **2.2 s** con `base` (16 CPUs, int8). **El STT no es el cuello de
  botella** — la limpieza sí (13-17 s). Y el hallazgo mayor: corrida la cadena completa con los
  **errores reales** del STT ("locita"→`lógica`, "demon"→`daemon`, "la absola"→`la app sola`), **la
  limpieza con contexto los corrigió todos** ⇒ **no hace falta un modelo grande**, el contexto repara
  el STT. Da vuelta la recomendación: **A2 local**, y lo único en su contra queda el empaquetado.
- ✅ **Prior art real:** `luana-core-copilot` (producto hermano del operador) ya tiene un
  `WhisperTranscriber` en producción (OpenAI Whisper cloud) — el patrón captura→backend→STT ya está
  validado en un sistema real, no hay que inventarlo.
- ⚠️ **Gap honesto:** el loop audio-real→transcripción NO se probó end-to-end (sin motor STT local
  instalado, sin credencial cloud, sin voz del operador grabada). Queda abierto como **T6**, no
  fabricado como "andaba".
- ✅ **El núcleo del pedido — probado end-to-end:** `claude -p --model claude-haiku-4-5-20251001`
  headless (sin sesión, "se limpia cada vez que grabás") sobre el mismo dictado enredado, con y sin
  contexto. Con un bloque corto de contexto (glosario + qué se tocó en la sesión), resolvió referencias
  ambiguas reales ("la tarjeta esa" → `PermissionCard`, "el nodo ese" → chip "(no reconocido)" del Mapa,
  "el daemon que arranca solo" → el override `~/.local/bin/arnesia`) que sin contexto quedaron
  genéricas o sin resolver. Confirma la intuición del operador: lo que desenreda no es más transcripción,
  es contexto de dominio compacto.

## Decisiones — TODAS FIRMADAS 🧑‍⚖️ → [`decisiones.md`](./decisiones.md)

| # | Qué se decide | Estado |
|---|---|---|
| **V-D1** | Motor STT (A1 cloud · A2 local · A3 nativo ❌ · A0 Claude ❌) | ✅ **FIRMADA** — **A2 local, modelo `base`**; sub-decisión resuelta como **adaptador por PATH** (detecta el motor instalado, degrada visible si no hay). Deuda **T7** = qué motor se bundlea |
| **V-D2** | Shape del contexto de la limpieza | ✅ **FIRMADA** — glosario **global** corto + últimos 2-3 turnos de `Session.Conv`; NO la conversación entera |
| **V-D3** | Gesto: toggle vs push-to-talk | ✅ **FIRMADA** (1ª ronda) — toggle + tope duro de duración |
| **V-D4** | ¿Limpieza obligatoria u opcional? | ✅ **FIRMADA** — por defecto, con escape a crudo (mismo camino que RF-227) |
| **V-D5** | ¿Se persiste el audio? | ✅ **FIRMADA** — no; solo temporal de vida acotada si el motor por PATH lo exige |
| **V-D6/D7** | T0 primero · spawn de limpieza endurecido | ✅ **FIRMADAS por no-objeción** |

## Estado

- [x] Investigación en vivo contra el motor real (WebKitGTK 2.52.3) — 2 vueltas
- [x] Núcleo (limpieza con contexto) probado end-to-end con evidencia
- [x] **T0bis ejecutado** — Fork A2 medido con voz real sintetizada (§1.8)
- [x] Probes + banco de medición reproducibles versionados en el paquete
- [x] `decisiones.md` con V-D0..V-D8 — **las 7 decisiones FIRMADAS 🧑‍⚖️**
- [x] `spec.md` — **RF-215…RF-228** con Gherkin
- [x] Paquete versionado en `main` (estaba untracked entero)
- [x] 🎨 **Mockup** `mockups/arnesia-voz-dictado.html` forkeado del baseline + registrado en `mockups/INDEX.md`
- [x] Capabilities — **4** (CAP-112 tauri · CAP-113 usecases · CAP-114 http-sse · CAP-115 fe-chat), las 4 `vivo` por R4
- [x] **RF-215 puente Rust** — `webkit2gtk` `permission-request`, solo audio · 4/4 tests
- [x] **Backend Go** — `TranscriptionPort` + adaptador por PATH + endpoint + limpieza endurecida
- [x] **FE** — botón, franja de etapas con nombre, composer poblado sin auto-envío · 13 stories
- [x] **Test de fitness del spawn endurecido** (RF-225) · 5/5
- [x] `PARIDAD.md` — RF-215…RF-228 fila por fila, con los comandos y sus salidas
- [x] **CADENA COMPLETA probada con VOZ REAL** por el daemon real (2026-07-26): el STT erró
      «warming»/«danon» y la limpieza con contexto los reparó + resolvió «la tarjeta esa de
      permisos» → `PermissionCard`. Destapó **4 bugs reales** del adaptador + **1 que habría
      matado T6 en la app instalada** (PATH del launcher gráfico)
- [x] **Instalador v0.2.20** en `instaladores/v0.2.20/` (.deb/.rpm/.AppImage) + `make dev-sync`
- [x] **Ronda 3 (2026-07-26) — el bug del micrófono, CAZADO y REPARADO.** El operador probó el dictado
      en su instalación v0.2.20: `«Falló escuchando: No se grabó nada»`. **No era el mic ni el permiso
      ni el motor STT** — es `MediaRecorder` de WebKitGTK, que dice soportar `audio/mp4`, arranca y
      entrega **0 bytes sin emitir `error`** (perfil de encoding con `rate=0` → `encodebin` nunca
      linkea el pad de audio). ⇒ **RF-229** (captura por WebAudio + WAV armado en el FE; de paso
      desaparece la dependencia de `ffmpeg`) + **RF-230** (log del daemon a archivo + endpoint de
      diagnóstico del FE, porque **no había ningún log que mirar**). V-D9..V-D12 en `decisiones.md`
- [ ] 🧑‍⚖️ **Gate final de PARIDAD** — falta SOLO el tramo del **micrófono** contra el binario instalado
      (ahora incluye RF-229/RF-230)

## Retomar aquí

**CONSTRUIDO · CADENA COMPLETA PROBADA CON VOZ REAL · falta el tramo del micrófono.** Rust 4/4 ·
Go entero verde · fitness verde · `npm run verify` verde · 13 stories del dictado · conformance
`279 checks · pass 66 · fail 0` · capabilities 111 → 115. Y sobre todo: el pipeline
`audio → STT → limpieza → endpoint` corrió con voz real por el daemon real y **reparó los errores
del STT** como el spike prometía. Detalle y evidencia literal en [`PARIDAD.md`](./PARIDAD.md).

**Lo único que queda para firmar el gate — y no lo puede cerrar un test:** el tramo
`micrófono real → MediaRecorder → POST` contra el **binario instalado**.

Todo está preparado: instalador en **`instaladores/v0.2.20/`**, `~/.local/bin/arnesia`
sincronizado, y el motor STT instalado en esta máquina (venv en `~/.local/share/arnesia-stt`, sin
sudo). Pasos: instalar el `.deb` → abrir la app → abrir un frente → botón de mic → hablar → cortar
→ confirmar que el composer se puebla y **no se auto-envía**.

**Tiene que ser contra lo instalado**: RF-215 existe porque `getUserMedia` no rechaza —queda
pendiente para siempre— sin el puente, y eso *solo* se reproduce ahí; el dev server concede por su
cuenta, así que un pass en `pnpm dev` no probaría nada.

**Lo demás abierto:**

- **T7** — qué motor STT se **empaqueta** en `.deb`/`.AppImage`/`.rpm`. Ya hay medición propia de
  `whisper-ctranslate2` corrido por el adaptador de este repo (1.5–3.7 s para ~14 s, `base`, CPU
  int8) y quedó documentado que **exige `--device cpu`** o explota sin CUDA. Falta medir
  `whisper.cpp` (necesita `cmake`/sudo) y medir ambos sobre el **`audio/mp4` real**.
- **El `audio/mp4` real nunca se transcribió** — todo lo medido usó WAV. Sub-gap del tramo del
  micrófono, no verificado.
- **Hallazgo ajeno destapado de paso:** 4 stories de `session-rail` fallan por contraste a11y de
  `--warn` sobre `--card` en tema claro (3.76:1 < 4.5). **No es de este paquete** — se registró en el
  BACKLOG en vez de arreglarse al voleo. Ver `PARIDAD.md` §Hallazgos.

## Archivos

- `spike-spec.md` — el documento completo: evidencia, comandos+outputs reales, forks, plan técnico.
- `decisiones.md` — V-D0..V-D8 (firmadas) + **V-D9..V-D12** (ronda 3: captura sin `MediaRecorder`,
  log del daemon, diagnóstico del FE, los dos motivos de «sin audio»); estas últimas esperan firma 🧑‍⚖️.
- `spec.md` — **RF-215…RF-230** con Gherkin; escrito contra `TranscriptionPort` para no depender del
  motor. Marca qué RF necesitan mockup (🎨) y qué decisiones asume como default.
- `medir_stt.py` — banco de medición del Fork A2 (§1.8): venv aislado, sin `sudo`, voz sintetizada con
  `piper`. Re-correr si cambia la máquina o se quiere comparar otro motor.
- `webkit_speech_test.py` — probe de `SpeechRecognition` (Fork A3): re-correr tras un upgrade del
  paquete WebKitGTK del sistema.
- `webkit_grabacion_test.py` — **la sonda de la ronda 3: ¿el motor GRABA, o solo dice que puede?**
  Mide bytes reales en tres caminos (MediaRecorder sin timeslice · con timeslice · WebAudio) en una
  sola corrida. Corrida 2026-07-26: `0 bytes · 0 bytes · 176128 muestras (pico 0.0777)`.
  **Re-correr tras un upgrade de WebKitGTK:** si el primer camino vuelve con bytes > 0, el bug del
  motor se arregló y RF-229 puede revisarse.
- ⚠️ **`webkit_captura_test.py` NO alcanza** y por eso existe el de arriba: prueba que
  `MediaRecorder` EXISTE y qué mimes DICE soportar, no que grabe. Pasó en verde mientras el dictado
  estaba roto toda la v0.2.20 — misma lección que V-D8, una vuelta más abajo.
- `webkit_captura_test.py` — probe de **captura real** (§1.6): `getUserMedia` + `MediaRecorder` + mimes
  + el cuelgue sin permiso. Re-correr tras un upgrade de WebKitGTK **o de wry** — si wry llega a manejar
  `permission-request` en `webkitgtk`, **T0 se borra**. Abre el mic y lo cierra en el acto; no graba nada.
