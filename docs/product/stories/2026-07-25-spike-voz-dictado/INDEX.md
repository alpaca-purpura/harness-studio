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
- [ ] 🎨 **Mockup** forkeado del baseline (`mockups/INDEX.md`) — bloquea los RF de superficie
- [ ] Capabilities (doctrina R2: `fe-chat/` · `http-sse/` · motor STT)
- [ ] Implementación (RF-215 primero, siempre)
- [ ] `PARIDAD.md` + 🧑‍⚖️ gate final

## Retomar aquí

**Investigación CERRADA · decisiones FIRMADAS · construcción DESBLOQUEADA.** El spec está redactado
contra `TranscriptionPort`, así que la elección de motor no lo toca. Lo que falta, en orden:

1. **Mockup** del botón + estados, forkeado del baseline vigente, antes de construir los RF 🎨.
2. **Capabilities** — tres (`fe-chat/`, `http-sse/`, motor STT). Sin ellas el gate-commit R3 bloquea.
3. **Construir empezando por RF-215** (puente Rust). Nunca al revés: sin él, el resto "anda" en dev y se
   cuelga mudo instalado.
4. **Test de fitness** del spawn endurecido (RF-225) — un enforcement sin test es una promesa.

**Lo que sigue abierto y no se puede cerrar solo:**

- **T6** — dictado con la **voz real del operador** por el **micrófono real**, contra el **binario
  instalado**. Lo medido usa voz sintética y entrada WAV.
- **T7** — qué motor STT se **empaqueta** en `.deb`/`.AppImage`/`.rpm`. Necesita medir `whisper.cpp`
  (pide `cmake`/sudo) y medir ambos sobre el `audio/mp4` real, no sobre WAV.

## Archivos

- `spike-spec.md` — el documento completo: evidencia, comandos+outputs reales, forks, plan técnico.
- `decisiones.md` — V-D0..V-D8; lo que espera firma 🧑‍⚖️.
- `spec.md` — **RF-215…RF-228** con Gherkin; escrito contra `TranscriptionPort` para no depender del
  motor. Marca qué RF necesitan mockup (🎨) y qué decisiones asume como default.
- `medir_stt.py` — banco de medición del Fork A2 (§1.8): venv aislado, sin `sudo`, voz sintetizada con
  `piper`. Re-correr si cambia la máquina o se quiere comparar otro motor.
- `webkit_speech_test.py` — probe de `SpeechRecognition` (Fork A3): re-correr tras un upgrade del
  paquete WebKitGTK del sistema.
- `webkit_captura_test.py` — probe de **captura real** (§1.6): `getUserMedia` + `MediaRecorder` + mimes
  + el cuelgue sin permiso. Re-correr tras un upgrade de WebKitGTK **o de wry** — si wry llega a manejar
  `permission-request` en `webkitgtk`, **T0 se borra**. Abre el mic y lo cierra en el acto; no graba nada.
