# Anexo — qué trae de verdad el payload de un hook (2026-07-26)

> Segunda tanda de verificación en vivo, mismo día. Cierra el lado **S2 / señal de proceso**, que el
> informe principal no había ejercitado. Método: `settings.json` temporal con un hook que vuelca su
> stdin tal cual, sobre los 6 eventos; una corrida de `claude -p` con uso real de herramienta
> (`Read`). Payloads crudos observados, no documentación.

## H1 · 🎯 La llave del join es `session_id` + **`prompt_id`**, y está en LOS DOS lados

Es el hallazgo que ordena toda la arquitectura del join:

| lado | de dónde sale | llave |
|---|---|---|
| **dinero** | log event OTel `claude_code.api_request` | `session.id` + **`prompt.id`** |
| **proceso** | payload del hook (`UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`, `SessionEnd`) | `session_id` + **`prompt_id`** |

El mismo turno tiene el mismo identificador en el canal de telemetría y en el canal de hooks. **El
join no necesita heurística de tiempo ni de orden**: es una igualdad de dos campos.

`arquitectura-telemetria.md` decía que la llave era «`session.id` + los resource attributes». Es
incompleto: con solo `session.id` el join es a nivel sesión y no se puede decir *«el 60 % se va en la
caja Y»*. Con `prompt_id` el join es **a nivel turno**, que es la granularidad que D14.3 fijó.

## H2 · Confirmado: los hooks NO traen dinero

Ninguno de los 6 payloads trae tokens, costo ni datos de cache. La separación de canales que
proponía la arquitectura es correcta y ahora está verificada, no supuesta:
**hook = proceso · OTel = dinero · daemon = decisión.**

## H3 · Campos por evento (observados)

Comunes a todos: `session_id` · `transcript_path` · `cwd` · `hook_event_name`.

| evento | campos propios |
|---|---|
| `SessionStart` | `source` (`startup`) |
| `UserPromptSubmit` | `prompt_id` · `permission_mode` · **`prompt`** (texto completo) |
| `PreToolUse` | `prompt_id` · `permission_mode` · `tool_name` · `tool_input` · `tool_use_id` |
| `PostToolUse` | ídem + `tool_response` · **`duration_ms`** |
| `Stop` | `prompt_id` · `permission_mode` · `stop_hook_active` · **`last_assistant_message`** · `background_tasks` · `session_crons` |
| `SessionEnd` | `prompt_id` · `reason` |

## H4 · 🔴 El payload del hook trae CONTENIDO — la allowlist también aplica acá

Distinto del canal OTel, donde `prompt` y `response` llegan `<REDACTED>` por default:

- `UserPromptSubmit.prompt` = **el prompt completo del usuario, en claro**.
- `Stop.last_assistant_message` = **la respuesta del asistente, en claro**.
- `PostToolUse.tool_response` = **el contenido leído/escrito por la herramienta**.

⇒ **El hook de S2 no puede reenviar su payload.** Tiene que proyectar a un puñado de campos
declarados (`session_id`, `prompt_id`, `hook_event_name`, `tool_name`, `duration_ms`, `cwd`
normalizado) y descartar todo lo demás **antes** de escribir a ningún lado. La allowlist de D15 deja
de ser una regla del receptor OTLP y pasa a ser **una regla de los dos caminos de ingesta**.

Corolario para el check `telemetria-no-egresa`: no alcanza con verificar que el hook escriba solo a
loopback. Hay que verificar **qué campos** escribe. Un hook que postea su stdin entero a
`127.0.0.1` cumple «no egresa» y aun así filtra la conversación al almacén local.

## H5 · `cwd` en todos los payloads — atribución de respaldo para S2

Todos los eventos traen `cwd`. En S2 (el usuario corre el arnés fuera de ArnesIA, sin inyección de
env vars) eso da una tercera vía de atribución, además de `plugin_id_hash`: **mapear `cwd` →
instalación del Portafolio**, que el índice ya conoce por su identidad `(home, id)`.

Orden de preferencia sugerido para `atribucion_confianza`:
`exacta` (resource attributes inyectados por nosotros) → `por-hash` (`plugin_id_hash`) →
`por-proceso` (`cwd` mapeado a una instalación conocida) → `sin-dato`.

## H6 · `PostToolUse.duration_ms` — la capa Desempeño tiene más señal de la que se creía

Duración **por herramienta**, no solo por request. Sigue fuera de alcance por falta de diseño, pero
el argumento «no hay señal» queda definitivamente muerto.

## H7 · ⚠️ `transcript_path` no apunta a un archivo

En esta corrida apuntó al **directorio** del proyecto
(`~/.claude/projects/<slug-del-cwd>`), no a un `.jsonl`. No se investigó más porque no lo usamos
—&nbsp;el JSONL no es fuente de nada acá—&nbsp;pero **cualquier diseño que asuma que
`transcript_path` es un archivo está equivocado**. Queda anotado para que nadie lo asuma después.

## H8 · OTel emite eventos POR HERRAMIENTA, y sin contenido

Una corrida con uso real de `Read` trajo dos eventos que no se habían visto:

```
tool_decision : prompt.id · decision=accept · source=config · tool_name=Read
                tool_use_id · tool_source=builtin
tool_result   : prompt.id · tool_name=Read · success=true · duration_ms=1
                tool_input_size_bytes=143 · tool_result_size_bytes=9
```

Dos consecuencias:

1. **El detector B11 («tool results obesos») queda desbloqueado** y es barato:
   `tool_result_size_bytes` es exactamente la señal, y llega **sin el contenido**. No entra al MVP
   (no pasó la regla A4 en D16.1), pero ya no está bloqueado por falta de dato.
2. **Parte de la señal de proceso existe también en OTel**, no solo en los hooks: decisión de
   permiso, éxito/fracaso y duración por herramienta. Los hooks siguen haciendo falta para lo que
   OTel no da (gate humano, rotación, corrida de caja), pero la dependencia es menor de lo asumido.

## H9 · 🎯 El bloque `env` de settings ENCIENDE la telemetría — S2 puede medirse completo

Probado con marcador distinto por variante, un solo receptor, y **con las env vars del shell
explícitamente desarmadas** (`env -u`):

| variante | ¿llegó telemetría? |
|---|---|
| **A** — `--settings <archivo>` con bloque `env` | ✅ sí |
| **B** — `.claude/settings.json` del proyecto, sin ningún flag | ✅ sí |
| **C** — ídem + `--setting-sources project,local` | ✅ sí |

**Esto cambia el diseño de S2.** La arquitectura asumía que fuera de ArnesIA no controlamos el
spawn ⇒ solo quedaba el hook, con señal de proceso y **sin dinero**. Falso: si el arnés lleva un
bloque `env` en los settings de su proyecto, en S2 llega **exactamente la misma señal que en S1**,
dinero incluido. Corolario: **B1 (re-warm por TTL) deja de ser «no aplica en S2»** siempre que el
canal esté encendido por esta vía.

⚠️ **Dónde vive ese bloque es una decisión de producto, no técnica**, y hay que tomarla:

- En el repo **del propio arnés** (nuestro) — sin fricción, es nuestro archivo.
- En el proyecto **del usuario** donde corre el arnés — es escribir settings de un tercero, y eso
  choca de frente con A8 (nunca sin backup + confirmación) y con el guardrail vigente. **Requiere
  consentimiento explícito, no puede hacerse en silencio.**

⚠️ **No verificado:** si un **plugin** puede aportar un bloque `env` (lo probado es el settings del
**proyecto**). Si pudiera, el arnés se instrumentaría solo al instalarse, sin tocar nada del usuario
— vale la pena verificarlo antes de diseñar el mecanismo de obligación.

> **Nota de método, para que no se repita:** los tres primeros intentos de esta prueba dieron
> negativo y el negativo era **falso** — un receptor de una prueba anterior seguía ocupando el
> puerto 4318 y se quedaba con el tráfico. Se detectó con un control (`pgrep` + una corrida de
> referencia) antes de escribir ninguna conclusión. **Un negativo sin control no es un resultado.**

---

## Qué cambia en el diseño

1. **La llave del join se escribe `(session_id, prompt_id)`** en el evento canónico y en el esquema
   SQLite, con índice por ese par. Es la fila que une dinero y proceso.
2. **La allowlist se aplica en los DOS adaptadores de ingesta**, no solo en el OTLP.
3. **`telemetria-no-egresa` se acompaña de un check hermano** que verifique la proyección de campos
   del hook, no solo su destino.
4. **`cwd` entra al evento** (normalizado, sin ruta de usuario) como vía de atribución de respaldo.
5. **S2 se rediseña con dos modos, no uno** (H9): *instrumentado* (el arnés lleva el bloque `env` ⇒
   misma señal que S1, dinero incluido, B1 aplica) y *degradado* (solo hook ⇒ proceso sin dinero).
   La UI tiene que distinguirlos: son dos niveles de dato distintos, no el mismo con otro nombre.
6. **`tool_result_size_bytes` se persiste** (H8): habilita B11 más adelante y no cuesta nada ahora.
