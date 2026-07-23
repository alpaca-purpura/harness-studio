# Spec — Chat dock · legibilidad y ergonomía · tramo A (mecánicos)

> `tipo: spec` · paquete `2026-07-22-chat-dock-ux`. Tramo A = CH-D1/D5/D6 (sin superficie
> visual nueva; decisiones firmadas 🧑‍⚖️ 2026-07-22) — **AS-BUILT**. Tramo B (CH-D2/D3/D4) se
> especifica DESPUÉS de la firma del mockup (Gate 1, `mockups/arnesia-chat-dock-ux.html`).

## CH-D1 — composer auto-grow hasta 3 líneas ✅

- `fitComposer(ta)` (`web/src/widgets/chat-dock/ui/chat-dock.tsx`): `height=auto` →
  `height=scrollHeight` SOLO con contenido; el clamp visual es CSS `max-h-[66px]`
  (3 líneas de `text-xs` + padding + borde) con `min-h-[36px]`; de ahí scroll interno.
- Vacío NO se dimensiona por JS: el placeholder envuelto infla `scrollHeight` (bug real
  cazado en vivo: el composer nacía en 66px). Un `ResizeObserver` re-ajusta el wrap cuando
  el ancho del dock cambia (interacción con CH-D5).
- WebKitGTK (Tauri Linux) no soporta `field-sizing: content` → JS, no CSS.
- **Verificado vivo** (vite+daemon+playwright): vacío 36px → 2 líneas 45px → 3 líneas 60px →
  6 líneas clampa 66px con scroll → reset 36px.

## CH-D5 — dock redimensionable por el borde izquierdo ✅

- `ShellPage` (`web/src/pages/shell/ui/shell-page.tsx`): `aside` pasa de `w-[360px]` fijo a
  `style.width` con estado; handle `role="separator"` en el borde izquierdo (drag por
  pointer-capture + teclado ←/→ de a 16px, accesible con `aria-value*`).
- Clamp `[300px, 60% de la ventana]`; persistencia en `localStorage["arnesia.dock.w"]`
  (efecto al soltar). La transición de apertura/colapso se desactiva mientras se arrastra.
- **Verificado vivo**: 360 → 2×ArrowLeft → 392, `localStorage` y `aria-valuenow` al día.

## CH-D6 — paquete cerrado: el gate deniega el árbol propio ✅ (CAP-99)

- `SessionService.ProtegerPaqueteCerrado(dir)` (patrón `Set*` del composition root):
  registra el árbol del kit materializado + su forma `~/…` (atrapa comandos Bash).
  Cableado en `cmd/arnesia/main.go` con `injector.BaseDir()` (getter nuevo del
  provisioner) → `~/.arnesia`.
- En `onControlRequest`, ANTES del auto-allow por grant: si el input crudo del tool_use
  menciona una marca → `RespondControl` deny inmediato (motivo «paquete cerrado…»), frame
  `permission_result` deny al FE (rastro `sys`), jamás tarjeta ni `await`. Ni grants
  vigentes ni el click humano lo abren. Tools read-only no pasan por el gate (van
  pre-aprobados) → LEER el kit sigue posible.
- Detección = substring sobre el JSON crudo del input — blunt adrede (deny es el lado
  seguro; cubre `file_path`/`command`/notebook de una). Upgrade si molesta un falso
  positivo: parsear el path por tool.
- **Test**: `TestControlRequestSobrePaqueteCerradoSeDeniegaSinTarjeta` (RED→GREEN) — deny
  sin await + un grant vigente de `Edit` NO abre el paquete. Suite Go + fitness verdes.

## Tramo B (pendiente de Gate 1)

CH-D2 (frames de actividad: el daemon deja de descartar thinking/tool_use en
`conductor.go:515-520` y el FE pinta tarjeta desplegable) · CH-D3 (cortar burbuja por
boundary de actividad en `session_service.go` y `sessions-store.ts`) · CH-D4 (renderer
markdown — dependencia nueva a elegir en esta spec cuando se firme el mockup).
