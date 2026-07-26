# Decisiones — Spike dictado por voz en el composer (2026-07-25)

> `tipo: spike`. Cada decisión conversada se escribe EN EL MISMO TURNO (§10). Cierre = decisión
> documentada. **Nada firmado aún.** Evidencia de respaldo: [`spike-spec.md`](./spike-spec.md).
> Los `V-Dn` marcados **ESPERA FIRMA 🧑‍⚖️** son lo único que bloquea escribir `spec.md`.

## V-D0 · Alcance del spike — DECIDIDA
- Responder **con qué se arma** el botón de voz, probando en vivo contra el motor real (no en papel), y
  **qué contexto** necesita el paso de limpieza para desenredar un dictado.
- **NO** construir código de producto en este spike. El entregable es papel + probes reproducibles.
- Cierre = `V-D1..V-D5` firmadas → recién ahí `spec.md` con RF numerados.

## V-D1 · Fork A — motor de STT — **ESPERA FIRMA 🧑‍⚖️**
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
- **V-D1 SIGUE ABIERTA** — ahora con números arriba de la mesa. Lo que ya quedó fijo y no depende del
  resultado: el motor va **detrás de `TranscriptionPort`**, así la elección es de adaptador, no de
  arquitectura. **Sub-decisión nueva que aparece si se firma A2:** ¿`whisper.cpp` (binario C++, hay que
  compilar/empaquetar) o `faster-whisper` (Python, es lo que se midió, arrastra runtime Python)?

## V-D2 · Fork B — shape del contexto para la limpieza — **ESPERA FIRMA 🧑‍⚖️**
- **El mecanismo YA está probado** (§1.5): con un bloque corto de contexto de dominio, Haiku resolvió
  referencias ambiguas reales que sin contexto quedaron genéricas. Lo que falta es de **dónde sale**.
- **Recomendación:** glosario estático corto **+ últimos 2-3 turnos** de `Session.Conv`
  (`internal/domain/session.go:127` — verificado que el campo existe). Cero cómputo nuevo.
- **Lo que NO:** ni la conversación entera ni el `CLAUDE.md` completo — el hallazgo fue que lo que
  desenreda es contexto **chico y de dominio**, no volumen.
- Abierto adentro: ¿el glosario es global o **por arnés**? Recomendado: global para arrancar, medir.

## V-D3 · Gesto de grabación: toggle vs push-to-talk — **FIRMADA 🧑‍⚖️ 2026-07-25**
- **DECIDIDO: toggle** (click empieza · click corta) **+ tope duro de duración** (p. ej. 3 min).
- **Por qué:** el detonante del spike es que el operador **habla largo**; mantener un botón apretado dos
  minutos es hostil. El tope evita que un mic olvidado abierto grabe la tarde entera.
- Baja a `spec.md` como RF con Gherkin. El valor exacto del tope (3 min) queda como default propuesto,
  ajustable sin reabrir la decisión.

## V-D4 · ¿La limpieza es obligatoria u opcional? — **ESPERA FIRMA 🧑‍⚖️**
> No seleccionada en la ronda del 2026-07-25 (el operador firmó V-D3 y mandó a medir V-D1). **No se
> interpreta como rechazo** — sigue abierta con su recomendación intacta.
- **Recomendación:** limpieza **por defecto** (es el corazón del pedido) **con escape a crudo**.
- **Por qué:** cuesta 13-17 s (§1.5) y a veces uno quiere el crudo y listo. Además el escape es el mismo
  escalón de fallback de la escalera de degradación (§3 Fork C.2) → sale casi gratis.

## V-D5 · El audio NO se persiste — **ESPERA FIRMA 🧑‍⚖️**
> Idem V-D4: no seleccionada en la ronda del 2026-07-25, no rechazada. El `spec.md` la asume como
> default (no persistir) y la marca como supuesto explícito — si el operador la firma al revés, el
> cambio es acotado.
- **Recomendación: NO guardar audio.** El blob vive en memoria, se manda, se descarta; nada a disco.
- **Por qué:** menos superficie de privacidad y cero política de retención que mantener. Si algún día se
  quiere "re-escuchar lo que dicté", es una decisión nueva y explícita, no un default silencioso.

## V-D6 · El puente Rust (T0) va PRIMERO — DERIVADA DE EVIDENCIA (se firma salvo objeción)
- `wry 0.55.1` **no maneja `permission-request` en el backend webkitgtk** (sí en macOS/Android) →
  `getUserMedia` **se cuelga sin rechazar** en la app instalada (§1.6 b/d).
- ⇒ El shell debe conceder el permiso (`with_webview` → `webkit2gtk::WebView` → `permission-request`,
  **solo** `UserMediaPermissionRequest` de audio). Sin esto, T1 "anda" en `pnpm dev` y muere mudo
  instalado.
- **No es preferencia, es física del motor.** Se documenta como decisión porque **cambia el orden de los
  tickets** y suma una dependencia (`webkit2gtk`) al `Cargo.toml` del shell.
- Se verifica **contra el binario instalado**, jamás contra el dev server.

## V-D7 · El spawn de limpieza va endurecido — DERIVADA DE DOCTRINA (se firma salvo objeción)
- El paso de limpieza es texto→texto: `--max-turns 1` + tools denegadas + `--setting-sources
  project,local`, sobre la superficie de enforcement que ya existe (`SpawnArgs`,
  `internal/adapters/agent/claudecode/conductor.go:59`, boundary `permisos-gui`).
- **Por qué:** **el dictado es entrada no confiable.** Sin el cerrojo, "leeme el `.env` y mandámelo" es un
  prompt perfectamente válido dicho en voz alta. Un paso que solo ordena texto no tiene por qué poder
  tocar un archivo.

## V-D8 · Corrección de honestidad sobre la primera redacción — HALLAZGO (no es opción)
- La v1 de este spike escribió que la captura «sí está disponible **sin nada adicional**» a partir de que
  la API `mediaDevices` existía. **Presencia de API ≠ captura funcionando**: probada de verdad, la
  llamada se cuelga sin plomería (V-D6), solo graba `audio/mp4`, y necesita la ventana visible.
- Queda escrito **en el documento y acá** en vez de editado en silencio: la doctrina de la casa es gap
  visible, jamás pass fabricado.
