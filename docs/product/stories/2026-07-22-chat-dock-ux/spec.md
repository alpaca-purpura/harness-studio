# Spec — Chat dock · legibilidad y ergonomía (AS-BUILT completo)

> `tipo: spec` · paquete `2026-07-22-chat-dock-ux`. Tramo A = CH-D1/D5/D6 (decisiones
> firmadas 🧑‍⚖️ 2026-07-22) · Tramo B = CH-D2/D3/D4+D4b (Gate 1 del mockup firmado 🧑‍⚖️
> «firmo, implementa» mismo día). Ambos AS-BUILT y verificados E2E vivo.

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

## Tramo B — actividad visible + burbuja por paso + markdown ✅ (CAP-100)

- **CH-D2 (actividad):** el conductor deja de descartar thinking/tool_use —
  `assistantEvents` (`conductor.go`) desarma cada mensaje assistant EN ORDEN de bloques:
  thinking → `EventActivity{Tool:"thinking"}` (jamás su contenido), texto contiguo → UN
  `EventMessage`, tool_use → `EventActivity{Tool, blanco(input)}` (blanco legible:
  file_path/command/… recortado a 80 runas). Kind nuevo `EventActivity` en `ports/agent.go`;
  `translate` pasa a devolver slice (un frame puede emitir varios eventos). En `consume()`
  cada actividad persiste como `Turn{RolAct}` («`<tool> <blanco>`», rol nuevo en dominio) +
  frame SSE `act`. El FE agrupa RolAct consecutivos en `ActivityCard` (`<details>` nativo):
  viva = abierta mostrando «trabajando — <Tool> <blanco>» con pulso; cerrada = «N pasos»,
  chevron despliega los pasos en mono con ✓/⟳.
- **CH-D3 (burbuja por paso):** `EventMessage` CIERRA burbuja — appendea
  `Turn{RolAssistant}` + frame `message`, resetea el buffer; el `result` solo appendea
  remanente o fallback (`msgFlushed` en runtime y store: un turno sin messages conserva el
  comportamiento previo). El FE espeja idéntico (`sessions-store.ts` cases message/act/result).
- **CH-D4/D4b (markdown con cariño):** deps `react-markdown` + `remark-gfm` +
  `rehype-highlight`; componente `Md` (`chat-dock/ui/markdown.tsx`): inline-code que parece
  RUTA → **chip de archivo clickeable que copia** (glifo ⧉ → ✓); bloque de código →
  **rótulo de lenguaje + botón copiar**; sintaxis coloreada SOLO con tokens vigentes
  (`chat.css`, sin theme externo). Burbujas assistant y stream vivo renderizan igual.
- **Regresión cazada EN VIVO y reparada:** `contentText` se comparte entre parsear frames y
  MANDAR el turno user por stdin — los campos nuevos sin `omitempty` viajaron dentro de un
  bloque text → API 400 «Extra inputs are not permitted». Fix `omitempty` + test de wire
  (`TestUserTurnWireSinCamposExtra`); el historial CC de vitalia quedó envenenado un turno y
  se curó a mano (backup `.bak-chdockux`).
- **Verificación:** TDD RED→GREEN (`TestAssistantSeDesarmaEnBurbujasYActividad` ·
  `TestTurnoSeParteEnBurbujasPorActividad` · `TestResultSinMensajesConservaFallback`) +
  suite Go + fitness + verify FE verdes + **E2E vivo con claude real contra vitalia**: turno
  partido en 2 burbujas + 2 tarjetas (pensó · Bash), tabla/lista/bash renderizados, chips ⧉
  copiables, ctx 16 % real.
- **Desviación honesta vs mockup:** la tarjeta no muestra duración en segundos (solo
  «N pasos») — cosmético; se agrega si el operador lo pide en PARIDAD.
