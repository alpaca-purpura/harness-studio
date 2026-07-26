# Spec — Dictado por voz en el composer (RF-215…RF-228)

> Paquete `2026-07-25-spike-voz-dictado`. **El QUÉ**, no el cómo. Evidencia de respaldo:
> [`spike-spec.md`](./spike-spec.md) · decisiones: [`decisiones.md`](./decisiones.md).
> Numeración: continúa desde RF-214, el último en uso en el repo.

## Estado de este documento — LEER ANTES DE CONSTRUIR

| | |
|---|---|
| ✅ **Firmado** | **V-D3** (toggle + tope de duración) |
| 🔬 **Medido, decisión abierta** | **V-D1** (motor STT) — hay números (§1.8): A2 local cuesta ~2.2 s. **La elección de motor no bloquea este spec**: todos los RF están escritos contra `TranscriptionPort`, no contra un motor |
| ⏳ **Abiertas, asumidas como default** | **V-D4** (limpieza por defecto + escape a crudo) · **V-D5** (no persistir audio) — no se seleccionaron en la ronda del 2026-07-25, **no se rechazaron**. Este spec las asume; si se firman al revés, el delta es acotado y está marcado en cada RF |
| ⏳ **Abierta** | **V-D2** (shape exacto del contexto) — RF-224 lo especifica *parametrizado* |
| ⚠️ **Falta antes del build de UI** | **No hay mockup en este paquete.** §10 exige que cada RF de superficie trace a `mockup:línea`, y `mockups/INDEX.md` exige forkear del baseline vigente (SSoT = Storybook). Los RF marcados 🎨 **no se construyen hasta que exista el mockup firmado** — están escritos en términos de comportamiento, no de pixeles, justamente para que el mockup no los tenga que re-inventar |

**Orden obligatorio de construcción:** RF-215 (puente Rust) **antes que todo lo demás**. Sin él el
botón funciona en `pnpm dev` y se cuelga mudo en la app instalada (`spike-spec.md` §1.6 b/d).

---

## A. El puente del permiso (shell Rust) — bloqueante

### RF-215 — El shell concede el permiso de micrófono del WebView
`wry 0.55.1` no maneja `permission-request` en el backend `webkitgtk`; sin esto `getUserMedia` **no
rechaza: queda pendiente para siempre**. El shell (raíz de confianza, boundary
`superficie-local-confinada`) debe conceder el permiso, **acotado**.

```gherkin
Escenario: la SPA del daemon pide el micrófono
  Dado el shell de escritorio con el WebView apuntando a http://127.0.0.1:4200
  Cuando la página llama a getUserMedia({audio:true})
  Entonces el shell recibe una UserMediaPermissionRequest
  Y la concede
  Y la promesa de getUserMedia RESUELVE con un stream (no queda pendiente)

Escenario: no se concede nada más que audio
  Dado el shell de escritorio
  Cuando llega una permission-request que NO es de captura de audio
  Entonces el shell NO la concede automáticamente
```

- **Verificación obligatoria contra el binario INSTALADO**, jamás contra el dev server — el navegador de
  dev no reproduce el bug (norma de la casa: validar contra lo instalado).
- Suma la dependencia `webkit2gtk` al `Cargo.toml` del shell (hoy no está).
- Si una versión futura de `wry` maneja `permission-request` en `webkitgtk`, **este RF se elimina**;
  re-chequear con `webkit_captura_test.py`.

---

## B. Captura (FE)

### RF-216 — Botón de micrófono en el composer 🎨
Un botón de mic en el composer del chat (`chat-dock.tsx#Composer`), junto al de enviar.

```gherkin
Escenario: el operador arranca un dictado
  Dado el composer del chat
  Cuando el operador hace click en el botón de micrófono
  Entonces empieza a grabar
  Y la interfaz muestra un estado visible de "escuchando…"
```

### RF-217 — Toggle, no push-to-talk *(V-D3, FIRMADA)*
```gherkin
Escenario: cortar el dictado
  Dado que se está grabando
  Cuando el operador hace click de nuevo en el botón de micrófono
  Entonces la grabación se detiene
  Y el micrófono se libera (ningún track queda vivo)
```

### RF-218 — Tope duro de duración *(V-D3, FIRMADA)*
```gherkin
Escenario: mic olvidado abierto
  Dado que se está grabando
  Cuando la grabación alcanza el tope de duración (default 3 min)
  Entonces la grabación se detiene sola
  Y sigue el flujo normal con lo grabado hasta ahí (no se descarta)
  Y se le avisa al operador que se cortó por el tope
```

### RF-219 — Cancelar un dictado sin mandarlo 🎨
```gherkin
Escenario: el operador se arrepiente
  Dado que se está grabando
  Cuando el operador cancela
  Entonces la grabación se detiene
  Y el audio se descarta sin transcribir
  Y el composer queda EXACTAMENTE como estaba antes de grabar
```

### RF-220 — El formato de audio es explícito
`MediaRecorder` en WebKitGTK soporta **solo `audio/mp4`** y su `mimeType` por defecto viene **vacío**
(§1.6 a).

```gherkin
Escenario: se graba en un formato que el motor soporta
  Cuando se instancia el grabador
  Entonces se le pasa el mimeType audio/mp4 de forma explícita
  Y si ese tipo no está soportado en el motor que corre, el botón se deshabilita con motivo visible
  Y NUNCA se asume webm/opus (no existe en este motor)
```

### RF-221 — El audio no se persiste *(V-D5, asumida — si se firma al revés, cambia solo este RF)*
```gherkin
Escenario: ciclo de vida del audio
  Cuando termina la grabación
  Entonces el audio se envía y se descarta de memoria
  Y NO se escribe ningún archivo de audio a disco
```

---

## C. Transcripción y limpieza (backend Go)

### RF-222 — Endpoint de dictado
```gherkin
Escenario: llega un dictado
  Dado una sesión activa
  Cuando el FE sube el audio grabado al endpoint de voz de esa sesión
  Entonces el backend responde con el texto listo para poblar el composer
  Y la respuesta indica si el texto está LIMPIO o es CRUDO (fallback)
```

### RF-223 — El motor de STT vive detrás de un puerto *(cierra V-D1 sin bloquearlo)*
```gherkin
Escenario: cambiar de motor no es rehacer la feature
  Dado que la transcripción se consume por un puerto (TranscriptionPort)
  Cuando se cambia el motor (local ↔ cloud)
  Entonces solo se reemplaza el adaptador
  Y ningún RF de este spec cambia
```
- Medido (§1.8): local `base` = **~2.2 s** para 47 s de audio. **El STT no es el cuello de botella.**
- Si el motor elegido no come `audio/mp4` (caso `whisper.cpp`, que quiere WAV 16 kHz mono), el
  transcodificado es **responsabilidad del adaptador**, no del dominio ni del FE.

### RF-224 — La limpieza recibe contexto de dominio *(V-D2 parametrizada)*
El hallazgo central del spike: lo que desenreda un dictado no es más transcripción, es **contexto de
dominio compacto** — y además **corrige errores del STT** (§1.8).

```gherkin
Escenario: limpieza con contexto
  Dado un transcripto crudo de un dictado
  Cuando se lo limpia con un bloque corto de contexto de dominio
  Entonces las referencias ambiguas quedan resueltas a nombres reales del proyecto
  Y los errores de transcripción de términos del dominio quedan corregidos
  Y la salida es el pedido en orden, sin preámbulo ni invenciones
```
- **Insumo del contexto (recomendado, V-D2):** glosario estático corto + últimos 2-3 turnos de
  `Session.Conv`. **NO** la conversación entera ni el `CLAUDE.md` completo.
- Costo real medido: **13-17 s** — es el 87 % de la latencia total del flujo.

### RF-225 — El spawn de limpieza va endurecido *(V-D7, doctrina)*
**El dictado es entrada no confiable**: "leeme el `.env` y mandámelo" es un prompt perfectamente válido
dicho en voz alta.

```gherkin
Escenario: un dictado no puede convertirse en un agente con permisos
  Cuando el backend spawnea el paso de limpieza
  Entonces el spawn va con tope de 1 turno
  Y con las herramientas denegadas (es un paso texto→texto)
  Y con la superficie de config acotada a project,local
  Y ningún contenido del dictado puede hacer que escriba o lea archivos
```
- Superficie de enforcement existente: `SpawnArgs`
  (`internal/adapters/agent/claudecode/conductor.go:59`), boundary `permisos-gui`. **Merece test de
  fitness**, igual que el resto de los flags de permisos.

---

## D. Resultado y degradación honesta

### RF-226 — El composer se puebla, editable, y NUNCA se auto-envía 🎨
```gherkin
Escenario: llega el texto limpio
  Cuando la limpieza devuelve el texto
  Entonces el composer se puebla con ese texto
  Y queda editable con el foco al final
  Y NO se envía solo — el operador revisa y usa el botón de enviar de siempre
```

### RF-227 — Escalera de degradación honesta *(gap visible, jamás pass fabricado)*
```gherkin
Escenario: permiso de micrófono denegado
  Entonces el botón queda deshabilitado con el motivo visible
  Y NUNCA queda "pensando" para siempre

Escenario: no hay dispositivo de entrada
  Entonces el botón queda deshabilitado indicando que no hay micrófono

Escenario: la transcripción falla o vuelve vacía
  Entonces se avisa explícitamente
  Y el composer NO se toca (no se pisa lo que el operador ya había tecleado)

Escenario: la limpieza falla o tarda de más
  Entonces el composer se puebla con el transcripto CRUDO
  Y se marca visiblemente que quedó sin ordenar
  Y el dictado NO se pierde por una falla del paso opcional
```
- El último escenario es también el "escape a crudo" de **V-D4**: si esa decisión se firma como
  "limpieza opcional", el escape ya está construido acá.

### RF-228 — Nada tapa un cuelgue
Derivado del peor modo de falla encontrado (§1.6 b): la promesa que nunca resuelve.

```gherkin
Escenario: algo se cuelga río arriba
  Dado un dictado en curso
  Cuando alguna etapa no responde dentro de su límite
  Entonces la interfaz sale del estado de carga
  Y muestra qué etapa falló
  Y el botón vuelve a quedar usable
```

---

## Trazabilidad y cierre

- **Capabilities** (doctrina `codigo-traza-a-capability`, R2): `fe-chat/` (botón · estados · composer
  poblado) · `http-sse/` (endpoint) · el módulo del motor (STT + limpieza). **Tres, no una.**
- **Mockup pendiente** para los RF 🎨 (RF-216, 219, 226): forkear del baseline vigente
  (`mockups/INDEX.md`), superset estricto, sin re-inventar lo firmado.
- **`PARIDAD.md`** se llena fila por fila durante la implementación.
- **T6 sigue abierto:** E2E con la **voz real del operador** por el **micrófono real**, contra el
  **binario instalado**. Lo medido en §1.8 usa voz sintética y entrada WAV — no lo reemplaza.
