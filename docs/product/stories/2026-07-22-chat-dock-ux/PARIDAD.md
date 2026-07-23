# PARIDAD — Chat dock · legibilidad y ergonomía

> paquete `2026-07-22-chat-dock-ux` · decisiones CH-D1..D6 + D4b firmadas · Gate 1 del mockup
> firmado · build completo en `main` (`39a31cd` tramo A · `a8e0e99` tramo B).

## Contraste decisión → entregado

| CH-D | decidido | entregado | evidencia |
|---|---|---|---|
| D1 | composer auto-crece hasta 3 líneas | ✅ `fitComposer` + `max-h-[66px]` + ResizeObserver (re-wrap al estirar) | E2E playwright: 36→45→60→clamp 66 con scroll→reset 36; bug real del placeholder cazado y reparado |
| D2 | actividad visible + desplegable estilo Cursor | ✅ `EventActivity` (conductor deja de descartar thinking/tool_use) → `RolAct` + frame `act` → `ActivityCard` (viva=herramienta en curso · cerrada=«N pasos») | E2E claude real: tarjetas «pensó» y «Bash cd …» intercaladas; screenshot `chat-dock-actividad.png` |
| D3 | una burbuja por paso | ✅ `EventMessage` cierra burbuja; result solo remanente/fallback (`msgFlushed`) | Conv real: 2 burbujas + 2 tarjetas en orden; `TestTurnoSeParteEnBurbujasPorActividad` + fallback |
| D4 | markdown renderizado | ✅ react-markdown + remark-gfm + rehype-highlight, estilos por tokens (`chat.css`) | Screenshot `chat-dock-tramoB.png`: tabla + listas + código renderizados |
| D4b | «cariño»: archivo clickeable · código lindo | ✅ chips ⧉ copian ruta (feedback ✓) · bloque con rótulo de lenguaje + botón copiar · sintaxis con paleta vigente | 43 chips en el turno real; screenshots |
| D5 | estirar el dock por el borde izquierdo | ✅ drag + teclado ←/→, clamp 300px–60 %, persistencia localStorage | E2E: 360→392, `aria-valuenow` + localStorage al día |
| D6 | app instalada NO modifica arneses de arnesia | ✅ deny por ruta en `onControlRequest` ANTES de grants, sin tarjeta (CAP-99) | `TestControlRequestSobrePaqueteCerradoSeDeniegaSinTarjeta` (incluye grant-no-abre) |

## Verificación corrida

Suite Go completa + fitness (0 FAIL) · `pnpm run verify` verde · E2E vivo con **claude real**
contra vitalia (turno partido, ctx 16 % real) · regresión de wire cazada EN VIVO (API 400 por
campos sin `omitempty` en `contentText` compartido parse/send) reparada con test
(`TestUserTurnWireSinCamposExtra`) + historial CC de vitalia curado (backup `.bak-chdockux`).

## Desviaciones (aceptadas en la firma)

1. La tarjeta de actividad no muestra duración en segundos — solo «N pasos» (cosmético; se
   agrega si se pide).
2. Detección CH-D6 = substring sobre el input crudo (blunt adrede, deny es el lado seguro);
   upgrade documentado: parsear path por tool.
3. Detección de paralelo: con tool_use paralelos, solo el último paso pinta «en curso».
4. La app instalada sigue con el binario anterior — falta `make installer` + self-update
   (deuda en `BACKLOG.md`).

## Firma

🧑‍⚖️ **FIRMADO — operador, 2026-07-22.** Cruda: «DE momento todo bien, si encuentro un error
lanzaré un bugfix, firmo». Contexto: el operador revisó las entregas del turno (evidencia E2E +
screenshots + app dev corriendo en `:5173`); el asistente transcribe la confirmación, no la
fabrica, y verificó antes que el código existe en `main` (`39a31cd` + `a8e0e99`, árbol pusheado).
Modalidad acordada: un error posterior se trabaja como **bugfix nuevo**, no reabre este gate.
