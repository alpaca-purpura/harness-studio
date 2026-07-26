# La cadena E2E, probada de punta a punta ANTES de escribir código (2026-07-26)

> **Para qué existe este documento:** el operador puso una condición dura — *«que se asegure que
> cuando yo instale el app realmente vea lo que tú has construido y validado»*. Esa cadena se probó
> **antes** de que exista una línea del módulo, para que el constructor no la descubra rota al final.
> Lo de acá está **corrido**, no propuesto.

## Qué se probó

```
web/ (vite build)  →  dist embebido en el binario Go  →  ./bin/arnesia serve
                   →  http://127.0.0.1:4200  →  navegador real (Playwright)
                   →  captura de la superficie vigente
```

| paso | comando | resultado observado |
|---|---|---|
| 1 · build | `bash scripts/bundle.sh --daemon-only` | OK · identidad `0.2.21.2607261116` · **`bin/arnesia` = 17,0 MB** |
| 2 · arranque | `./bin/arnesia serve` | escucha en `127.0.0.1:4200` |
| 3 · salud | `curl /healthz` | **200** |
| 4 · UI embebida | `curl /` | sirve `index.html` real (`<html lang="es" data-theme="light">`) |
| 5 · navegador | Playwright → `http://127.0.0.1:4200/` | título **ArnesIA**, app montada |
| 6 · captura | screenshot | [`evidencia/baseline-mapa-antes.png`](evidencia/baseline-mapa-antes.png) |

**Es la UI del binario, no el dev server de vite.** `embeddedUI()` (`cmd/arnesia/main.go:462`) sirve
`web/dist` desde el `go:embed`; si el build no trae dist, devuelve `nil` y no habría UI. Que la
página cargue **prueba** que el binario lleva la UI adentro.

## Lo que muestra la captura — la línea base del gate

La superficie vigente del Mapa, con datos reales del operador (arnés `vitalia`), y en particular
**el conmutador de capas en su estado de hoy**:

```
ARTEFACTOS  off | auto | todos        Estructura | Tokens | Desempeño | Proceso
                                       ─────────   ▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓
                                       (activa)    (los tres, apagados)
```

Esta imagen es el **«antes»** del gate de PARIDAD: contra ella se compara el «después» con la capa
Mejora encendida. Guardarla ahora, y no reconstruirla al final, es lo que hace que la comparación
sea honesta.

## Nota de método

⚠️ El daemon de esta prueba corre contra el `~/.arnesia` **real del operador**, así que se levantó
para la captura y se bajó. Cualquier E2E del módulo de telemetría debe apuntar a un **HOME de
prueba** (`HOME` o `--addr` propios) y a un **puerto efímero**: escribir eventos de prueba en la
base real contaminaría los totales que la propia feature muestra.

⚠️ Un navegador de automatización quedó colgado reteniendo su perfil y bloqueó el primer intento
(`Browser is already in use`). Si vuelve a pasar, se cierra por PID; no es un fallo de la app.

## Lo que esta prueba NO cubre

| qué | por qué | cómo se cierra |
|---|---|---|
| El **instalador** (`.deb`/`.AppImage`/`.rpm`) | `make installer` **bumpea la versión** y versiona en `instaladores/vX.Y.Z/`; no corresponde hacerlo antes de que exista la feature | al cerrar el paquete: `make installer` **+ `make dev-sync`** (existe `~/.local/bin/arnesia`, y sin eso el shell instalado ignora el sidecar nuevo) |
| El **shell Tauri** | acá se manejó la UI servida por el daemon, que es la misma que embebe el bundle; la ventana nativa no se ejercitó | click-through humano sobre la app instalada, gate de PARIDAD |
| Windows y macOS | no hay runner de esas plataformas | sigue abierto (G4, `tauri#11992`) |

---

## Cierre del Tramo A contra la app INSTALADA (2026-07-26)

> La condición del operador era literal: *«que se asegure que cuando yo instale el app realmente
> vea lo que tú has construido y validado»*. Esto es esa verificación, corrida de verdad.

```
make installer   → v0.2.22 · .deb 10,6 MB · .rpm 10,6 MB · .AppImage 85,9 MB + checksums.txt
make dev-sync    → ~/.local/bin/arnesia reemplazado (identidad 0.2.22.2607261433)
```

`make dev-sync` **no es opcional en esta máquina**: existe el override local de self-update, y el
shell instalado lo prefiere **siempre** sobre el sidecar del `.deb`. Sin ese paso se probaría el
binario viejo creyendo haber probado el nuevo — el modo de falla más caro de esta cadena.

### La prueba de que lo instalado ES lo construido

```
sha256  ~/.local/bin/arnesia  = e699bd73a0f939b9…
sha256  ./bin/arnesia         = e699bd73a0f939b9…      ⇒ byte por byte, el mismo binario
```

### El módulo, vivo en el binario instalado

| qué | resultado |
|---|---|
| `~/.local/bin/arnesia telemetria` | lista los 7 subcomandos |
| `GET /healthz` | **200** |
| `GET /api/telemetria/resumen` · `/salud` · `/mejoras` | **200** los tres |
| `POST /v1/logs` | **200** — la ingesta OTLP está viva |
| UI en `http://127.0.0.1:4200/` con navegador real | monta, título ArnesIA, consola sin errores |

Captura: [`evidencia/instalada-v0222-tramo-a.png`](evidencia/instalada-v0222-tramo-a.png).

### Lo que esta captura NO muestra, y es correcto que no lo muestre

**La capa «Mejora» todavía no está dibujada.** El Tramo B se autorizó recién con la firma del
mockup y se está construyendo. Lo que esta verificación prueba es que **el backend construido viaja
dentro del binario que el operador instala** y responde ahí — no en un dev server.

La verificación visual de la capa se repite al cerrar el Tramo B, contra un instalador nuevo, y va
al `PARIDAD.md` junto a [`evidencia/baseline-mapa-antes.png`](evidencia/baseline-mapa-antes.png),
que es el «antes».

---

## Cierre del Tramo B — la capa, vista en la app instalada (2026-07-26)

```
make installer  → v0.2.23        make dev-sync → ~/.local/bin/arnesia
sha256 instalado = sha256 compilado = 70a19f8446fe08a0…
```

Navegador real contra el binario instalado, arnés `vitalia` (datos reales del operador, que **nunca
corrió con telemetría**): el conmutador muestra **`Estructura · Mejora · Desempeño · Proceso`**, la
capa Mejora enciende, y los dos slots que siguen apagados ya no dicen «necesita telemetría» sino el
motivo real —falta el diseño, no la señal—.

### Lo que la pantalla destapó, y ninguna story podía cazar

**Primera pasada** ([`evidencia/instalada-v0222-capa-mejora.png`](evidencia/instalada-v0222-capa-mejora.png)):
la franja decía bien *«— — · Este arnés nunca corrió con telemetría»* y doce centímetros más abajo
la lista afirmaba **«Hay datos y ningún punto de mejora que pase el corte · Los seis detectores
corrieron sobre 0 corridas»**.

**Dos afirmaciones falsas, cada bloque defendible por separado, contradiciendo al de arriba.** Las
131 stories no podían verlo: verifican componentes **aislados**, y las dos frases nunca se
renderizaron juntas.

**Segunda pasada, tras el fix**
([`evidencia/instalada-v0223-capa-mejora-corregida.png`](evidencia/instalada-v0223-capa-mejora-corregida.png)):
la sección **no se dibuja** —que es lo que `design.md` §7.4 ya mandaba y la implementación había
ignorado— y el DOM ya no contiene «Hay datos». La franja manda.

### Lo que deja como norma

1. **El fix fue en la composición, no en el copy**: `hayDatos` pasó de opcional-con-default a
   **prop obligatoria**, y `tsc` cazó los 6 call-sites. Un default que asume que hay datos miente
   justo cuando no los hay.
2. **La condición se verificó revirtiendo el fix**: el candado se puso rojo y reprodujo la captura
   palabra por palabra, y el **control positivo** quedó verde — así que «esconder siempre la lista»
   tampoco pasa.
3. **El mismo patrón estaba en cuatro sitios más** (S2 degradado, el `✓` del Portafolio, la sección
   vacía del inspector y un umbral `corridas <= 5` escondido en el JSX, que era una decisión de
   producto tomada a escondidas).

### El hueco que queda abierto, y es estructural

El candado cubre **esta** composición porque la story la reproduce a mano. **`pages/` sigue con 0
stories en el repo**, y es justo donde se compone. Una composición futura que nadie espeje queda
ciega igual. Está declarado en `PARIDAD.md`.
