# spec — chat CC funcional v1 (modificar arneses)

> Autorización: /goal del operador 2026-07-08 («implementa la solución de la conversación
> asegurándote el buen funcionamiento probándolo tú mismo…») = orden de implementar sobre
> mockup v1 + decisiones #1–#6. RF trazados a `mockup-chat.html` (commit del paquete).

## RF (numeración continúa el paquete)

- **RF-110 · Alcance sesión=arnés.** El chat de una sesión corre SIEMPRE con cwd del
  arnés de la sesión + doctrina inyectada (ya existente). Chip fijo del arnés visible
  en el Dock. — mockup `scope` (chip fixed).
- **RF-111 · Chip de nodo.** Seleccionar un nodo en el Mapa lo adjunta como chip
  removible; el turno viaja con línea de alcance `[alcance: <clase> «<id>» — archivo
  <fuente_path>]` antepuesta al texto. Quitar el chip = sin línea. — mockup `#nodeChip`.
- **RF-112 · Spawn con permisos del rol.** El spawn del Dock pasa `Permisos` resueltos
  del ROL del arnés (server-side, jamás del FE): `--permission-mode default` +
  `--allowedTools` read-only + `--permission-prompt-tool stdio`. Rol desconocido ⇒ base
  deny-by-default (degradación honesta).
- **RF-113 · Tarjeta de permiso.** Frame `permission` ⇒ tarjeta inline en el transcript
  (estado `await`, pip ámbar): herramienta + input legible (Edit/MultiEdit ⇒ diff
  −/+ · Write ⇒ contenido · Bash ⇒ comando · resto ⇒ JSON) + 3 acciones: **Permitir
  esta sesión** (grant TTL del rol) · **Permitir una vez** (TTL acotado a 1 s) ·
  **Denegar**. La tarjeta muestra la regla que recordaría el grant. — mockup `#permCard`.
- **RF-114 · Resolución server-side.** `POST /sessions/{id}/permission` sin `role` en el
  body ⇒ el daemon usa el rol del arnés de la sesión. Deny del rol gana sobre el click
  (ya existente). `permission_result` cierra la tarjeta y deja rastro en el transcript.
- **RF-115 · Grant vigente no re-pregunta.** Segunda petición del mismo tool con grant
  vivo ⇒ auto-allow + rastro «grant vigente» (ya existente en Go; el FE lo muestra).
- **RF-116 · Stop real.** Botón Stop mientras `streaming|await` ⇒
  `POST /sessions/{id}/interrupt` ⇒ control_request `interrupt` in-band; pendientes de
  permiso se deniegan con motivo «interrumpido». — mockup `#sendBtn.stop`.
- **RF-117 · Gate visible.** Si el turno aprobó ≥1 escritura, al `result` el Dock corre
  `GET /harnesses/{arnes}/conformance` y pinta el veredicto en el transcript
  (`🛡 N/N PASS` o los checks en fallo). — mockup `#gateBlock`.
- **RF-118 · Cambiar de arnés = cambiar de sesión.** El chip/contexto sigue a la sesión
  activa; el nodo en alcance pertenece a UNA sesión (se limpia al cambiar).

## No-goals v1 (decisión #5)

assistant-ui/CodeMirror/AG-UI (migración de presentación, aditiva, paquete siguiente) ·
markdown rico · tarjetas de tool-use no-permiso (Read/Bash del stream) · imágenes ·
slash-menu funcional · cola de turnos.

## Gherkin (casuística E2E)

- Dado sesión sobre `dev-full-cycle` registrado, Cuando pido una edición, Entonces
  aparece tarjeta con diff y la sesión queda `await`.
- Cuando apruebo «esta sesión», Entonces el edit corre, el mismo tool NO re-pregunta
  (grant), y al terminar veo el gate del arnés.
- Cuando apruebo «una vez», Entonces la siguiente petición del mismo tool VUELVE a
  preguntar.
- Cuando deniego, Entonces el edit no corre, el archivo no cambia y la conversación
  sigue con mi motivo.
- Cuando el rol deniega el tool (p.ej. reviewer→Edit), Entonces mi click «permitir» NO
  alcanza: efectiva=deny con motivo del rol.
- Cuando pulso Stop en streaming, Entonces el turno muere pronto y la sesión vuelve a idle.
- Cuando selecciono el skill builder en el Mapa, Entonces el chip aparece y el turno
  lleva la línea de alcance; al quitarlo, no.
- Cuando cambio a otra sesión/arnés, Entonces chips, permisos pendientes y transcript
  son los de ESA sesión.
