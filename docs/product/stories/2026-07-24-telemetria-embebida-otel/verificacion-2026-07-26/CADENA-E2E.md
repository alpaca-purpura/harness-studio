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
