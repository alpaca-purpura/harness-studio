# Decisiones — Identidad de build en Ajustes (2026-07-26)

> Paquete chico. Cada decisión conversada se escribe EN EL MISMO TURNO (§10).
> Detonante: *«¿puedes hacer que en ajustes aparezca la última versión incluyendo un último
> número que sea el build, para así saber si estoy usando la última versión compilada e
> instalada o una anterior?»*

## El agujero (medido, no supuesto)

Ajustes mostraba `arnesia · huella 1c7443f · 2026-07-07`. Eso **no responde** la pregunta, por dos
motivos que se comprobaron el mismo día:

1. **El semver no estaba en el binario.** `0.2.21` vive en `Cargo.toml`/`package.json`; el daemon Go
   no lo conocía. Ajustes no podía decir qué versión corría.
2. **La huella es el COMMIT, no el BUILD.** Ese día se bundleó v0.2.21 **dos veces** desde el mismo
   árbol (una para el instalador, otra para incluir un fix posterior): dos binarios distintos,
   huella idéntica. Justo el caso que había que distinguir.

## B-D1 · El número de build es un sello de compilación `AAMMDDHHMM` — **ELEGIDA POR EL OPERADOR**

Se ofrecieron tres opciones; el operador eligió **timestamp compacto**.

| Opción | Por qué no |
|---|---|
| Contador incremental (`0.2.21.34`) | Necesita un archivo de estado versionado que sube en cada compilación → ruido en git, y se pisa si se compila en otra máquina |
| Fecha-hora legible sin número | No hay «número de build» que comparar de un vistazo — era el pedido literal |
| **Sello `AAMMDDHHMM`** ✅ | **Cambia SIEMPRE** (dos builds del mismo commit dan números distintos) · **monótono** (más grande = más nuevo, sin excepciones) · **cero estado** en el repo |

⇒ La identidad completa es **`0.2.21.2607260225`**.

**Hora LOCAL, no UTC.** El número se compara contra «cuándo compilé», y esa referencia es el reloj
que el operador tiene delante; UTC lo mostraría corrido cinco horas. El costo aceptado: un cambio de
huso o el fin del horario de verano puede repetir una hora **una vez al año**.

**Sin identidad inyectada, dice `dev`.** Un `go build` pelado (CI, `go run`) no sella nada y la
tarjeta escribe «build sin sellar (dev)» — decir «no sé qué build soy» es honesto; inventar un
número sería la mentira que este paquete vino a sacar.

## B-D2 · El sello se inyecta en `bundle.sh`, NO en el Makefile — **DECIDIDA**

El self-update corre **este mismo script** (`scripts/bundle.sh --daemon-only`). Si los `-ldflags`
vivieran en el Makefile, un binario producido por el botón **Actualizar** se reportaría como `dev`:
la app se actualizaría a sí misma y perdería su identidad en el acto. Un solo punto de inyección
cubre el build a mano, el `make dev-sync` y el self-update.

## B-D3 · El sello es ADITIVO a la huella, no su reemplazo — **DECIDIDA**

Son dos preguntas distintas y las dos siguen en la tarjeta: la **huella** dice *qué commit*, el
**sello** dice *qué compilación*. Se reordena la jerarquía visual —el build arriba, el commit
debajo— porque el build es el que contesta lo que el operador abre Ajustes para averiguar.

## B-D4 · Ajustes avisa cuando corre un build viejo — **ELEGIDA POR EL OPERADOR**

Se ofreció «solo mostrar el número» vs «avisar»; eligió **avisar**. Mira dos lugares, los dos casos
reales de esta casa:

1. **El propio ejecutable instalado.** `make dev-sync` reemplaza `~/.local/bin/arnesia` con la app
   abierta: el archivo ya es nuevo y el proceso sigue siendo el viejo. → «cerrá y reabrí la app».
2. **El binario compilado en el repo** (`<repo>/bin/arnesia`). Compilaste y no instalaste. →
   «corré `make dev-sync`».

**Compara mtime contra el sello, no dos sellos entre sí.** Leer el sello del OTRO binario obligaría
a ejecutarlo (arbitrario, lento) o a rastrear sus bytes (frágil); un `stat` lo responde.

**Margen de 5 minutos** entre el sello y el mtime: el linker escribe el archivo DESPUÉS de que el
script sella, así que sin margen el binario recién instalado se denunciaría a sí mismo como viejo en
cada arranque — el falso positivo que volvería inútil el aviso. Un aviso que aparece siempre no se
lee nunca.

## B-D5 · Este paquete NO lleva mockup nuevo — la superficie se especifica contra las stories — **DECIDIDA (refinamiento 2026-07-26)**

§10 pide mockup→decisiones→spec, y la norma «pegarse al Storybook» dice que el **SSoT del UI es
Storybook**, no los `.html`. Acá la superficie **ya existe** (`UpdateCard`, firmada en el paquete
`boton-actualizar`) y RF-231 reescribe dos filas y agrega una caja condicional: forkear un `.html`
de 276 líneas para dibujar dos filas habría producido un snapshot nuevo que driftea, no
entendimiento.

Lo que **sí** se hace, para no dejar la deuda escondida:

1. El QUÉ al pixel vive en [`design.md`](./design.md), escrito contra el CSS real.
2. Las tres stories nuevas (`IdentidadDelBuild` · `BuildViejoCorriendo` · `BuildSinSellar`) son la
   especificación ejecutable de la superficie.
3. Se **re-deriva el bloque de versión** del snapshot `stories/2026-07-07-boton-actualizar/mockup-actualizar.html`
   —que además traía la línea «el binario aún NO embebe versión (pendiente `-ldflags`)», o sea que
   este RF cierra una deuda que ese mockup ya tenía anotada— y se registra su fila en
   `mockups/INDEX.md` (DoD §5).

**Lo que NO se toca del snapshot:** su paleta ámbar pre-rebrand. Es deuda del rebrand PRENTER, ya
registrada en BACKLOG, y arreglarla acá sería meter un paquete adentro de otro.

## B-D6 · Los límites del aviso se declaran, no se tapan — **DECIDIDA (refinamiento 2026-07-26)**

El aviso de B-D4 tiene tres bordes conocidos. Ninguno es bug; los tres se escriben para que nadie
los descubra creyendo que encontró uno:

| Límite | Qué pasa | Por qué se acepta |
|---|---|---|
| **Ventana ciega de 5 min** | Un rebuild dentro del margen de B-D4 **no** dispara aviso | Es el mismo margen que evita el falso positivo. Falso negativo corto y auto-corregido: el siguiente build sí avisa. Bajarlo reintroduce la denuncia a sí mismo |
| **Se evalúa al montar la tarjeta** | El aviso aparece cuando Ajustes pide `/api/version` — o sea, cada vez que se ENTRA a Ajustes. **Medido en vivo 2026-07-26:** con la app abierta se tocó el binario y el aviso apareció al volver a Ajustes, sin recargar. Lo que no hay es polling estando ya parado ahí | Un watcher sobre dos rutas para un dato que se mira una vez por sesión no se paga |
| **CI compila sin sello** | `ci.yml` usa `go build` pelado ⇒ esos binarios dicen `dev` | **Nunca se distribuyen**: no hay workflow de release; los instaladores salen de `make installer` → `bundle.sh`. Y si uno corriera, decir `dev` es exactamente lo correcto (B-D1) |

**Además, lo que este paquete NO hizo:** no hay `arnesia --version` en la CLI. El pedido era Ajustes;
la CLI queda como deuda registrada en BACKLOG, no como promesa incumplida.

## ⏳ Pendiente de firma 🧑‍⚖️

B-D1..B-D6 quedan **DECIDIDAS y construidas** (B-D5 y B-D6 nacieron en el refinamiento y documentan
lo ya construido: no cambian una línea de código). Falta la firma del gate de PARIDAD.
