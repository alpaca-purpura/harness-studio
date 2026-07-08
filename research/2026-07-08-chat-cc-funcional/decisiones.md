# Decisiones — chat CC completamente funcional (modificar arneses)

> Una entrada por decisión conversada. Estados: PROPUESTA → FIRMADA (🧑‍⚖️ operador).
> Regla METODOLOGIA §10: toda decisión conversada se escribe AQUÍ en el mismo turno.

## #1 — Modelo de contexto: sesión=arnés · chip=nodo (PROPUESTA)

**Qué:** el contexto del chat deriva de DOS niveles ya firmados + uno nuevo:

1. **Sesión ↔ arnés (N:1, firmado it.14):** cada sesión del rail vive atada a UN arnés;
   su conversación CC corre con `cwd` = raíz del arnés (S2, HS-06) y la doctrina
   inyectada al spawn (puente ② HS-11). Cambiar de arnés en el Mapa ⇒ cambiar de
   sesión (o crear una) — jamás se "re-apunta" una conversación viva a otro arnés.
2. **Nodo seleccionado ⇒ chip de alcance (NUEVO):** seleccionar un nodo (skill, hook,
   caja…) en el Mapa ofrece adjuntarlo como **chip de alcance** en el composer; el
   mensaje viaja con la referencia resuelta por nomenclatura v1.1
   (clase→ubicación ⇒ ruta real del archivo). El chip es removible; no fuerza nada
   (guía sin bloqueo).
3. **Doctrina = arneses propios (ya construido):** el kit arnesia-kit embebido se
   inyecta a TODO spawn (`--plugin-dir` + `--append-system-prompt-file`); el chat
   "sabe" crear/modificar arneses porque su arnés de fábrica viaja con él. ② ↛ ③.

**Por qué:** reusa 100% lo firmado (multisesión it.14 · S2 · 3 cuerpos) y responde
exactamente al pedido: "si cambio de arnés, el contexto es de ese arnés; si selecciono
un skill, es para ese skill".

## #2 — Transporte FE: assistant-ui como librería de UI, daemon como única verdad (PROPUESTA)

**Qué:** el Dock usa `@assistant-ui/react` con **ExternalStoreRuntime** — los mensajes
viven en el store Zustand alimentado por el SSE multiplexado del daemon (taxonomía
AG-UI, HS-04); assistant-ui solo pinta. Nada de LocalRuntime/adapters que dupliquen
estado.

**Por qué:** HS-04 firmó component-selection (assistant-ui + AG-UI + CodeMirror
merge); ExternalStoreRuntime es el único modo donde el daemon queda dueño de la
verdad (multisesión, replay, persistencia JSONL).

## #3 — Permisos en el chat: control_request → tarjeta inline con grants TTL (PROPUESTA)

**Qué:** cuando CC pide permiso (`control_request` type `can_use_tool`), el Dock
pinta una tarjeta inline bloqueante de la sesión (estado `await`): Permitir una vez ·
Permitir esta sesión (grant TTL, Fase E HS-11) · Denegar con nota. Los permisos
ofrecidos derivan del rol (boundary `permisos-derivan-del-rol`).

**Enmienda (investigación ①, patrón claudecodeui):** la tarjeta muestra la **regla
derivada** que el grant recordaría (herramienta · alcance · TTL · muere con la sesión)
— «recordar» debe ser una decisión informada. NO persistimos a
`.claude/settings.local.json` como hacen otros: los grants viven en el daemon
(más estricto, y respeta ② ↛ ③).

## #4 — Gate de conformance post-edición visible en el chat (PROPUESTA)

**Qué:** tras cada tanda de ediciones al arnés, el Dock muestra tarjeta de gate:
`conformance --arnes` del arnés de la sesión (endpoint ya existente
`GET /api/harnesses/{id}/conformance`) con veredicto por check; nada se da por
"terminado" en la conversación sin gate verde o desviación explícita (regla de
honestidad + nada sin eval).

## #5 — v1 funcional hand-rolled; assistant-ui/CodeMirror = fase de presentación (FIRMADA de facto por /goal)

**Qué:** la funcionalidad (permisos, alcance, gate, stop) aterriza SOBRE el Dock
hand-rolled existente. La migración a assistant-ui + CodeMirror merge + taxonomía AG-UI
(decisión #2) es un refactor de PRESENTACIÓN aditivo, paquete siguiente.

**Por qué:** el /goal ordena funcionamiento probado E2E; la librería no cambia qué
funciona, cambia cómo se pinta. Aditivo sin pérdida: nada de lo que se construye ahora
(frames, store, endpoints) se tira — assistant-ui consumirá el mismo store.

## #6 — El rol viaja del arnés, jamás del FE (FIRMADA de facto por /goal)

**Qué:** el daemon resuelve el rol desde `graph.l0` del arnés de la sesión (META de
enganche) tanto al spawnear (flags) como al resolver un permiso (si el body no trae
rol). El FE no elige autoridad.

**Por qué:** boundary `permisos-derivan-del-rol` — la autoridad se impone fuera del
razonamiento Y fuera del cliente.

---

*(pendientes de conversar: geometría exacta del dock vs mockup it.14 · qué tan
automático corre el gate · edición con diff aceptar/rechazar vs edición directa)*
