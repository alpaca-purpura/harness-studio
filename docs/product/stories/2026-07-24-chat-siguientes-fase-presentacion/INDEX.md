# Chat dock — 4 sub-ítems reales de la vieja "fase presentación" (paquete de arranque)

> Origen: BACKLOG "decisión #5" (2026-07-08, `stories/2026-07-08-chat-cc-funcional/decisiones.md:59-67`)
> re-scopeado 2026-07-24 (barrido de deuda viva HS-27, ver `docs/product/ledger/HS-27.md`).
> **No es un paquete mockup→spec→PARIDAD todavía** — es el punto de arranque con la
> investigación ya hecha, para que la próxima sesión no vuelva a investigar desde cero. Cada
> sub-ítem se separa a su PROPIO paquete (mockup si toca UI nueva) cuando se ataque.

## Contexto (no relitigar)

HS-26 (2026-07-22, `stories/2026-07-22-chat-dock-ux/`) ya cerró 2 de los 6 sub-ítems originales
de la decisión #5 (markdown rico vía `react-markdown`, tarjetas de tool-use vía `ActivityCard`) —
**sin necesitar `@assistant-ui/react` ni `@codemirror/*`**, validando la premisa original ("la
librería no cambia qué funciona, cambia cómo se pinta"). El swap completo a assistant-ui quedó
**descartado** (no cierra ninguna brecha ya cerrada a mano; costo real de bundle +154kB,
`investigacion.md` del paquete original). Quedan 4 sub-ítems, independientes entre sí.

## 1. Widgets ricos para turnos `sys` (permiso resuelto / gate de conformance)

**Estado hoy:** texto plano en un pill punteado — `web/src/widgets/chat-dock/ui/chat-dock.tsx:230-236`
(`Bubble()`, rama `t.rol === "sys"`), alimentado por strings armados a mano en
`web/src/shared/store/sessions-store.ts:336-358` (p.ej. `"🛡 gate de conformance ${arnes}: ..."`).
Es la desviación #1 que `stories/2026-07-08-chat-cc-funcional/PARIDAD.md:35-37,62` documentó como
pendiente desde el inicio — HS-26 mejoró turnos `act`/`assistant`, nunca tocó `sys`.

**Tamaño:** chico. Mismo patrón hand-rolled que `ActivityCard` (`chat-dock.tsx:180-228`) — un
componente que en vez de un string plano reciba la estructura real (veredicto del gate:
bloqueantes/warns/pass; permiso: qué se pidió/quién lo resolvió) y la pinte con jerarquía visual
(color por severidad, lista de checks). Sin dependencias nuevas.

**Para arrancar:** mirar qué datos YA tiene `sessions-store.ts` en el momento de armar el string
`sys` (no lo pierdas parseando el string después — pasá el objeto estructurado directo al turno).

## 2. Slash-menu funcional en el composer

**Estado hoy:** no existe. `Composer()` (`chat-dock.tsx:265-340`) es solo `<textarea>` + botón
enviar/stop — confirmado leyendo la función completa, cero manejo de `/`.

**Tamaño:** chico-mediano. Necesita: detectar `/` al inicio del input, lista de comandos
disponibles (¿cuáles? — pregunta abierta, no resuelta en esta investigación), popover de
selección con teclado (↑/↓/Enter/Esc). Sin librería nueva necesaria salvo que se decida un menú
más rico.

## 3. Cola de turnos (reemplazar el 409 por encolado)

**Estado hoy:** un 2º turno mientras el 1º sigue en streaming devuelve `409` —
`internal/usecase/session_service.go:29-30` (`ErrBusy`, mapeado a HTTP 409) + comentario línea
330: *"One turn at a time: a second turn while streaming would interleave stdin + assembling."*
Sin cambios desde el diseño original (spec.md del paquete `chat-cc-funcional`).

**Tamaño:** chico-mediano, toca FE + backend. Diseño a decidir ANTES de codear: ¿el turno
encolado se manda automático apenas el anterior cierra, o el usuario debe reenviarlo? ¿el
composer se deshabilita mientras encola, o queda editable y se manda lo que esté cuando le toque
el turno? Necesita una decisión del operador antes de spec.

## 4. Diff editable accept/reject por chunk (`@codemirror/merge`)

**Estado hoy:** `DiffLines` (`web/src/widgets/chat-dock/ui/permission-card.tsx:162-183`) — diff de
solo-lectura, líneas completas −/+, sin edición ni granularidad de chunk. Es el ÚNICO sub-ítem
de los 4 que genuinamente pediría una librería nueva.

**Tamaño:** mediano — es el más grande de los 4. Solo vale la pena si el producto quiere que el
operador pueda EDITAR el diff antes de aprobar (hoy alcanza con aceptar/rechazar en bloque
completo, que ya funciona). El research original (`stories/2026-07-08-chat-cc-funcional/
investigacion.md:216-220`) ya evaluó `@codemirror/merge` (`unifiedMergeView`, `acceptChunk`/
`rejectChunk`) y concluyó que para el caso simple (aceptar/rechazar en bloque) el patrón
hand-rolled actual YA es lo correcto — este sub-ítem solo se justifica si aparece un pedido real
de edición chunk-a-chunk.

**Para arrancar:** antes de instalar la librería, confirmar con el operador que el caso de uso
(editar antes de aprobar) es real y querido — no asumirlo del research viejo.

## Retomar aquí

Ningún sub-ítem tiene mockup ni spec todavía. Al atacar cualquiera de los 4: si toca UI nueva
(1, 2, 4) arrancar releyendo `mockups/INDEX.md` (línea base = Storybook) antes de proponer
diseño — superset estricto del chat dock ya firmado (HS-26). El 3 (cola de turnos) es más
backend/lógica que UI, pero igual necesita la decisión de diseño abierta arriba antes de spec.
