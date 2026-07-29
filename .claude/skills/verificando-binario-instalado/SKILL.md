---
name: verificando-binario-instalado
description: "Verifica un cambio de backend Go de arnesia CONTRA EL BINARIO REAL (build + daemon + CLI + HTTP), sin arriesgar el ~/.arnesia real del operador. Usar cuando se tocó código del daemon/CLI (internal/, cmd/arnesia/) y hay que confirmar con evidencia — no con lectura de código ni solo `go test` — que el cambio funciona instalado. No es para verificar UI/FE (eso es Storybook + el navegador)."
contract:
  caja: false
  rol: apoyo-testing
---

# Verificando contra el binario instalado (skill de apoyo)

> Nace del paquete `docs/product/stories/2026-07-28-persistencia-escalabilidad/`: el operador pidió
> "primero asegurate de cómo vas a probar antes de arrancar" para un cambio de persistencia
> (concurrencia cross-proceso). Este es el mecanismo que se armó, se probó fase a fase, y se
> promueve acá para no reinventarlo la próxima vez.

## Cuándo

Un cambio toca `internal/adapters/**`, `internal/usecase/**` o `cmd/arnesia/**` (backend Go del
daemon/CLI) y hace falta MÁS que `go test` — específicamente cuando el riesgo es de
**comportamiento en runtime real** (concurrencia entre procesos, arranque, HTTP, archivos bajo
`~/.arnesia`), no de lógica pura que un test unitario ya cubre entero.

No usar para cambios de FE/UI (`web/`) — ese camino es Storybook + navegador, no este.

## Por qué NO tocar el `~/.arnesia` real directo

`~/.arnesia` es el estado del OPERADOR (portafolio, sesiones, índice). Un bug en el código que se
está probando puede escribir basura ahí. La regla: **todo el ciclo de iteración corre contra una
COPIA en un HOME sandbox**; el binario real instalado (`~/.local/bin/arnesia`, si existe — self-update
migrado) solo se toca al final, para una verificación de cierre, y con las mismas precauciones.

## Pasos

1. **Build.** Si el cambio es solo Go (no toca `web/`), `go build -o <scratch>/arnesia-test
   ./cmd/arnesia` alcanza — el SPA embebida (`web/dist`) ya está compilada en el árbol y no hace
   falta rehacerla para probar backend. Si el cambio SÍ toca el FE, usar
   `bash scripts/bundle.sh --daemon-only` (recompila SPA + daemon, sin sidecar/instaladores).

2. **Sandbox HOME.** Copiar el `~/.arnesia` real a un HOME de prueba **respetando el anidado**:
   ```
   mkdir -p <scratch>/home/.arnesia
   cp -r ~/.arnesia/* <scratch>/home/.arnesia/
   ```
   ⚠️ Gotcha real (se pisó una vez en este mismo paquete): copiar al HOME sin el subdirectorio
   `.arnesia/` produce un store vacío silencioso — el binario resuelve `$HOME/.arnesia/portafolio.json`,
   no `$HOME/portafolio.json`, y no hay ningún error, solo `{"entradas":[]}` donde no correspondía.

3. **Levantar el daemon sandboxed**, puerto libre distinto del real (4200):
   ```
   HOME=<scratch>/home <scratch>/arnesia-test serve --addr 127.0.0.1:<puerto> \
     --log <scratch>/home/.arnesia/logs/test.log > <scratch>/serve.out 2>&1 &
   disown
   ```

4. **Verificar con AMBAS vías**, nunca solo una — HTTP y CLI usan instancias de store
   independientes y eso es exactamente lo que a veces hay que probar (dos procesos, dos cachés):
   - `curl http://127.0.0.1:<puerto>/healthz` + los endpoints `/api/*` relevantes.
   - `HOME=<scratch>/home <scratch>/arnesia-test <subcomando>` (mismo HOME, proceso NUEVO).

5. **Matar por ruta exacta ANTES de confiar en un resultado**, y entre corridas:
   ```
   pkill -f "<scratch>/arnesia-test serve"
   ```
   ⚠️ Gotcha real: un daemon sandbox de una corrida ANTERIOR que quedó vivo en el mismo puerto
   responde curl con datos VIEJOS y ninguna señal de error — un falso positivo silencioso. Pasó
   en este mismo paquete. Confirmar con `ps aux | grep <nombre-del-binario-de-prueba>` si hay
   dudas de qué proceso contestó.

6. **Concurrencia cross-proceso — nunca goroutines.** Si el cambio promete una garantía entre
   PROCESOS (no solo goroutines de un mismo binario), el test tiene que spawnear procesos del SO
   de verdad. Patrón Go: el test re-invoca `os.Args[0]` con una env var que dispara un
   `TestHelperProcess` que hace el trabajo real (ver
   `internal/adapters/filelock/filelock_test.go` para el ejemplo completo, incluida la prueba
   negativa: deshabilitar el fix a propósito y confirmar que el test SÍ lo detecta).

7. **Cierre — contra el binario REAL instalado** (`~/.local/bin/arnesia`, si existe):
   `make dev-sync` lo reconstruye y reemplaza atómicamente. Antes de correrlo, confirmar que
   ningún daemon sandbox de prueba sigue vivo en el puerto real (4200) para no confundir salidas.
   Después: solo operaciones de LECTURA contra el `~/.arnesia` REAL (p.ej. `arnesia portafolio
   listar`) para confirmar compatibilidad — las de escritura se probaron ya contra la copia.

## Guardarraíles

- Nunca lanzar el binario de prueba con `HOME` apuntando al `$HOME` real.
- Nunca asumir que un `curl`/CLI que respondió 200 probó lo que se quería probar sin confirmar
  qué PROCESO contestó (paso 5).
- Un test de concurrencia sobre goroutines de un mismo proceso NO prueba una garantía
  cross-proceso — son dos cosas distintas (flock, por ejemplo, es advisory y por-proceso).
- Limpiar los sandboxes/procesos de prueba al terminar (`pkill` + `rm -rf` del scratch home).
