# Decisiones — Chat dock · legibilidad y ergonomía (6 quejas del operador)

> `tipo: decisiones` · paquete `2026-07-22-chat-dock-ux` · conversadas con el operador el
> 2026-07-22 (sus 6 quejas textuales = el requerimiento). El ejecutor NO relitiga; si una
> resulta inviable en código, documenta el porqué AQUÍ (nueva CH-D) y elige la alternativa más
> cercana al espíritu. Estado del código verificado contra `main` antes de escribir cada una.

## CH-D1 · El composer auto-crece hasta 3 líneas

Hoy el `<textarea>` es `rows={1}` fijo con `max-h-32` pero SIN auto-grow
(`web/src/widgets/chat-dock/ui/chat-dock.tsx:182-200`): escribís largo y «te quedás a medias»
scrolleando dentro de una línea. Decisión: crece con el contenido hasta un máximo de **3
líneas**; de ahí en adelante scroll interno. Enter envía / Shift+Enter salto siguen igual.

## CH-D2 · Actividad visible mientras claude trabaja + detalle desplegable (estilo Cursor)

Hoy, mientras la sesión ejecuta, solo hay un pip pulsante en el header y 3 puntos en la burbuja
vacía (`chat-dock.tsx:21,122-126`); el **thinking y los tool_use sin permiso se DESCARTAN en el
daemon** (`internal/adapters/agent/claudecode/conductor.go:515-520`) — el operador no sabe qué
está pasando. Decisión: mientras claude piensa/ejecuta se ve **qué** hace (línea de actividad:
herramienta en curso · «pensando»), con un **desplegable opcional** para el detalle de los
pasos. Requiere que el daemon deje de descartar esos eventos y los publique como frames al FE.

## CH-D3 · Una burbuja por paso — nunca un párrafo inmenso

Hoy TODOS los `text_delta` de un turno van a un único buffer — daemon
(`internal/usecase/session_service.go:428-448`) y FE (`web/src/shared/store/sessions-store.ts:282-315`)
— así que los textos intermedios entre tool_use se concatenan en UNA sola burbuja gigante.
Decisión: cada bloque de texto del asistente (separado por actividad/tool_use) = **burbuja
propia**, intercalada con las tarjetas de actividad (CH-D2) y de permiso en el orden real de
llegada.

## CH-D4 · Markdown renderizado, no crudo

Hoy el texto assistant se pinta plano con `whitespace-pre-wrap` (`chat-dock.tsx:142-162`) y no
existe NINGÚN renderer de markdown en `web/package.json`. Decisión: la respuesta se ve
**renderizada** (como Notion/Obsidian): headings, listas, código con bloque, tablas, links.
Implica dependencia nueva (react-markdown o similar — elegir en spec); los estilos salen de los
tokens DTCG reales, jamás inventados (norma «pegarse al Storybook»).

## CH-D5 · El dock se agranda arrastrando el borde izquierdo

Hoy el ancho es fijo `w-[360px]` (`web/src/pages/shell/ui/shell-page.tsx:38-45`) y no hay
ningún splitter en `web/src`. Decisión: agarrar el **borde izquierdo** del dock y estirar hasta
el ancho cómodo, con mín/máx sensatos y **persistencia** del ancho elegido entre sesiones.

## CH-D6 · GUARDRAIL: el chat de la app instalada NO modifica los arneses de arnesia

Incidente reportado: en la app instalada, el chat confundió los arneses que el operador atendía
(arneses-producto) con los arneses propios de arnesia. **Aclaración del operador el mismo día:
aplica al RUNTIME de la app instalada — en el repo de desarrollo sí se trabajan (acá se fabrica
arnesia).** Decisión: el alcance/permisos de las sesiones del chat (gate de permisos de
`stories/2026-07-08-chat-cc-funcional/`) excluye los paths del paquete propio de arnesia
(paquete cerrado) — escritura ahí se **deniega siempre**, sin tarjeta de permiso.
