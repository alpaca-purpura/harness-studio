# Plan de desarrollo — capa «Mejora» (telemetría embebida)

> `tipo: plan-desarrollo` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> **Cero código.** Este documento dice **qué escribir**, en qué archivo, con qué firma, qué RF
> cierra, qué capability lo reclama, cómo se verifica y con qué comando se demuestra. Otro agente
> construye leyendo **solo** esto y los documentos que se citan.
>
> **Insumos vinculantes** (no se relitigan): [`decisiones.md`](decisiones.md) D1..D17 + el bloque
> 🧑‍⚖️ FIRMADO 2026-07-26 · [`arquitectura-modulo.md`](arquitectura-modulo.md) A1..A22 ·
> [`escenarios.md`](escenarios.md) · [`spec.md`](spec.md) RF-232…RF-286 ·
> [`design.md`](design.md) · [`plan-storybook.md`](plan-storybook.md) ·
> [`capabilities-a-crear.md`](capabilities-a-crear.md) CAP-118…CAP-139 ·
> [`verificacion-2026-07-26/INFORME.md`](verificacion-2026-07-26/INFORME.md) V1..V7 ·
> [`verificacion-2026-07-26/ANEXO-hooks.md`](verificacion-2026-07-26/ANEXO-hooks.md) H1..H10.
>
> **39 tickets · 4 tramos.** Tramo 0 (precondiciones, sin código de producto) · Tramo A (los RF sin
> superficie, Go) · Tramo B (la superficie, bloqueada por el 🧑‍⚖️ del mockup) · Tramo C (cierre).

## 0 · Reglas que valen para TODOS los tickets

1. **Ningún cambio de código sin capability** (R1·R2·R3·R4,
   [`codigo-traza-a-capability.md`](../../../architecture/boundaries/codigo-traza-a-capability.md)).
   La hoja YAML se crea **en el mismo commit que crea el último de sus símbolos** — nunca antes
   (R1 rompería con punteros colgantes) y nunca después (R2/R3 lo cazan).
2. **`null` ≠ `0` en toda la superficie.** Convertir un `null` en `0` es el pass fabricado.
3. **Control positivo obligatorio** en todo escenario cuyo éxito sea «no llegó nada» — §4.
4. **Cero mock donde hay dato real.** Los payloads de `verificacion-2026-07-26/evidencia/` son
   golden files; los fixtures FE se rotulan `MEDIDO` / `ILUSTRATIVO` (plan-storybook §3).
5. **Cada ticket cierra verde**: su comando de demostración pasa **antes** de abrir el siguiente.
6. **Un ticket = un commit** (Conventional Commits, scope del paquete), y el commit incluye su
   hoja de capability. `lefthook` pre-commit regenera las cifras del checkpoint solo.

---

## 1 · Orden de construcción, y por qué

`spec.md` §«Orden obligatorio»: **los RF sin superficie (§H, RF-282…RF-286) van antes que cualquier
píxel.** Si la UI se construye primero, se persiste PII antes de tener dónde borrarla.

| tramo | tickets | por qué va acá |
|---|---|---|
| **0 · Precondiciones** | T1–T3 | Tres bloqueantes verificados y dos contrastes rotos se deciden y se escriben **antes** de tocar código: un constructor que los descubre a mitad de camino los resuelve mal o baja un gate. Cero código de producto ⇒ cero riesgo. |
| **A.1 · Dominio y puertos** | T4–T5 | El evento canónico es el contrato de los cuatro emisores. Todo lo demás lo importa; nada lo importa a él. |
| **A.2 · Cálculo puro** | T6–T7 | Catálogo y costeo son funciones puras sin I/O: se testean con tabla y son donde viven los tres bugs ajenos (D9.6). Van antes que la ingesta porque la ingesta ya costea al escribir. |
| **A.3 · Los dos caminos de ingesta** | T8–T10 | Decodificar → mapear con allowlist (OTLP) → proyectar con allowlist (hook). **Este es RF-282**, el RF que gobierna el orden entero. |
| **A.4 · Almacén** | T11–T12 | Ahora hay evento válido y filtrado que guardar. Antes de esto no hay nada legal que escribir a disco. |
| **A.5 · Servicio** | T13–T15 | Atribución, join, conciliación, retención y detectores: la lógica que convierte filas en la frase del producto. |
| **A.6 · Superficie de red** | T16–T19 | Recién con el servicio armado se abre un puerto. Abrirlo antes sería aceptar tráfico que no se sabe guardar. |
| **A.7 · Emisores** | T20–T23 | Hook (S2), spawn y `result` (S1), eventos del daemon, forward. Se cablean **después** del receptor: un emisor sin receptor es telemetría perdida en silencio, que es exactamente lo que el paquete existe para no hacer. |
| **A.8 · Cierre del backend** | T24–T27 | CLI de verificación, wiring+presupuestos, fitness+arch-lint, checks de conformance del arnés. El fitness va al final porque `go-arch-lint` falla con componentes cuyo `in:` no matchea archivos (§10.4). |
| **B · Superficie** | T28–T37 | **Bloqueado por el 🧑‍⚖️ del mockup** (iteración 1). Orden interno: `shared/ui` → `entities` → nodos → chrome → tarjetas → inspector → diálogo → Portafolio → transporte. Cada capa consume solo la anterior; así ninguna story espera a un componente que no existe. |
| **C · Cierre** | T38–T39 | E2E contra el binario instalado, capabilities/cifras, PARIDAD y el gate humano. |

**Puerta entre A y B:** el Tramo B **no arranca** sin la firma 🧑‍⚖️ del mockup (`spec.md`
§Estado: *«ningún RF 🎨 se construye antes de su 🧑‍⚖️»*). El Tramo A no la necesita: los RF sin
superficie ya están autorizados por el gate del bloque D9·D11·D13·D14·D15·D16.

---

## 2 · Los cinco bloqueantes, resueltos acá

Estas cinco resoluciones se escriben en `decisiones.md` como **D18…D22** en el T1 y a partir de ahí
son vinculantes. No son propuestas: son la decisión que el constructor ejecuta.

### D18 · Cross-import `entities/arnes → entities/telemetria`: **props primitivas + story-candado**

`design.md` §1.2 le da a `arnes-node` (en `entities/arnes`) una prop `mejora?: CifraCaja` tipada
desde `entities/telemetria`. **`steiger fsd/no-cross-imports` lo rechaza**, corre en
`pnpm --dir web run fsd`, y `fsd` está dentro de `verify` — o sea que rompe CI.

**Se adopta la opción A de `plan-storybook.md` §1.3, y se le agrega una segunda pata.** La regla,
literal, para que nadie la reinterprete:

> **`entities/arnes` no importa `entities/telemetria`. Nunca, por ningún escape.**
> `ArnesNode` recibe **primitivos**: `cifraUSD?: string` · `participacionPct?: number` ·
> `confianza?: "exacta" | "por-hash" | "por-proceso" | "sin-dato"` · `marcaFuga?: string` ·
> `motivoSinDato?: string`. El widget `map-canvas` —que **sí** puede importar las dos entities,
> porque widgets→entities es la dirección legal— compone `CifraCaja → props`.

Costo declarado: el literal del chip de confianza y el formato de dinero quedan en dos sitios. Se
ata con **dos** candados, no uno:

1. **`CopyConfianzaEsUnaSola`** (story de `map-canvas.stories.tsx`, §2.13 de plan-storybook):
   renderiza el nodo `por-hash` **y** `<MarcaConfianza confianza="por-hash">` en la misma story y
   asserta igualdad literal de `textContent` y de `title`. Vive en el widget porque es el único
   lugar que puede importar las dos entities.
2. **`entities/telemetria/model/selectors.test.ts`** exporta las constantes de copy
   (`ETIQUETA_CONFIANZA`, `TITULO_CONFIANZA`) y las asserta contra `design.md` §7.3. El nodo las
   recibe por prop desde el widget, que las lee de ahí. **No se duplica el string en `arnes-node`**:
   se duplica el *tipo* (una unión de 4 literales), que es lo que `tsc` sí puede cuidar.

Se rechaza B (`entities/telemetria/@x/arnes.ts`) porque `steiger.config.ts` declara su enforcement
**incierto**, y se rechaza C (overlay del widget sobre el nodo) porque reintroduce el
posicionamiento absoluto que J-9 prohibió.

### D19 · `knowledge` no es una clase: es una **banda**

Verificado en `internal/domain/box.go:27`: el enum `Clase` tiene **diez** primitivas
(`skill · subagent · hook · rule · command · mcp · plugin · settings · output-style · statusline`)
más el marcador `no-reconocido`, y el comentario dice textual que *«el valor legacy `knowledge`
pliega a `rule`»*. En `web/src/entities/arnes/model/types.ts` pasa lo mismo. «Conocimiento» en este
árbol es una **banda** (`selectBase`) y un token (`--c-knowledge`), no una clase de nodo.

**Consecuencia, y es de vocabulario, no de código:**

1. **`design.md` §7.3 pierde la fila `sin dato · conocimiento`.** Un nodo de la Base es `rule`, y su
   copy es el de `rule`. Lo hace T2.
2. **`spec.md` RF-238 corrige su enumeración.** Donde dice *«un nodo de clase rule, mcp, knowledge,
   subagent, hook, command o settings»* pasa a decir *«un nodo que no es caja»* — que es lo
   asertable: `isCaja(box)` ya existe en `entities/arnes` y es el predicado real que usa
   `arnes-node.tsx:39`. Lo hace T2.
3. **La story `SinDatoResto`** (§2.5) cubre el hueco con el copy genérico
   `sin dato atribuible — esta capa mide cajas`.
4. **No se agrega `knowledge` a `Clase`.** Sería tocar el contrato L0 (`graph.l0.schema.json`) y no
   es alcance de este paquete.

### D20 · `puesto` no existe en `EntradaPortafolio`: sale del **`rol` del arnés**, resuelto en el backend

Verificado: `web/src/entities/portafolio/model/types.ts:76` y `internal/domain/portafolio.go:311`
tienen `Empresas []string` e `Instalaciones []Instalacion`, y **ningún campo `puesto`**.
`Instalacion` (`portafolio.go:265`) tiene `ProyectoPath`/`InstallPath`/`Tipo`/`Origen`/`Deriva`.

Pero el dato **sí existe en otro lado**: `internal/domain/graph.go:36` declara
`Rol string \`json:"rol,omitempty"\`` con el comentario *«canonical (was `puesto`); the role the
arnés serves»*, y el FE ya lo tipa (`entities/arnes/model/types.ts:75 rol?: string`).

**Decisión — ni se recorta el RF ni se inventa un campo en el wire del Portafolio:**

> La fila de RF-265 se agrupa por **`(identidad, instalacion_id)`** — que es la unidad real que el
> Portafolio modela. La etiqueta de «puesto» sale del **`rol` del arnés indexado**, resuelto
> **server-side** en `GET /api/telemetria/portafolio` (que ya tiene el índice a mano) y viajando
> como campo `puesto *string` del DTO `FilaPortafolio`. Si el arnés no declara `rol`, o no está
> indexado, viaja `null` y la UI dice `puesto sin declarar` — el copy que `design.md` §3.5 ya
> preveía. **`domain.EntradaPortafolio` y el wire de `GET /api/portafolio` no se tocan.**

Consecuencias que hay que escribir y no olvidar:

- **RF-265 «un arnés en dos puestos» se cumple por instalación**, no por puesto: dos instalaciones
  del mismo arnés dan dos filas. Si las dos resuelven al mismo `rol`, las dos filas dicen el mismo
  puesto y se distinguen por instalación. La story `UnArnesDosPuestos` (§2.15) se mantiene con ese
  fixture (`acme-cli`, 2 instalaciones en 2 proyectos, ya real en `entradasDemo`).
- **`SinPuestoDeclarado`** deja de ser un caso de borde: es el caso **normal** de hoy, porque
  ningún arnés del dogfood declara `rol`. La story lo cementa para que nadie fabrique un puesto.
- **Deuda al `BACKLOG.md`** (T1): «`puesto` como faceta propia de la instalación del Portafolio
  — hoy se deriva del `rol` del sello, que es del arnés, no del lugar donde se usa».
- **J-4 sigue vigente**: no se renombra el `MetaChip` `reporta a` de `map-bar.tsx:79`.

### D21 · Los dos contrastes que rompen el gate: fix **por tema**, sin tokens nuevos

`plan-storybook.md` §2.18 reporta 3,26:1 y 4,03:1 sin decir en qué tema. Recalculado con
composición alpha correcta sobre `--card` de cada tema:

| combinación | tema claro | tema oscuro |
|---|---|---|
| `--warn` sobre `--warn-soft` (disclaimer) | **3,24 : 1 ❌** | 6,11 : 1 ✅ |
| `--crit` sobre `--crit-soft` (marca de fuga grave) | **4,04 : 1 ❌** | 4,66 : 1 ✅ (raspando) |
| `--foreground` sobre `--warn-soft` | 13,14 : 1 ✅ | ✅ (por construcción del token) |
| `--crit` sobre `--card` puro | 4,75 : 1 ✅ | 5,33 : 1 ✅ |
| `--heat-4` vs `--heat-3` contiguos | 1,37 : 1 | — |

**Los dos fallos son del tema CLARO únicamente.** El oscuro pasa los dos, pero `--crit`/`--crit-soft`
pasa a 4,66 contra un mínimo de 4,5: cualquier retoque de esos tokens lo tumba. Por eso el arreglo
**no puede ser un color global**: tiene que ser por combinación, y las dos salidas ya están escritas
en el propio `design.md`.

1. **Disclaimer «estimado por el runtime, no es facturación»** — se aplica la regla que
   `design.md` §4 fin ya declara y que su propia tabla §4.2 contradice:
   **texto `--foreground`** sobre fondo `--warn-soft`, borde
   `color-mix(in srgb, var(--warn) 35%, transparent)`. En claro da 13,14:1 y en oscuro pasa por
   construcción. `--warn` queda **solo** para borde (requisito 3:1). **La tabla §4.2 se corrige**
   (T2); el ticket que lo pinta es **T32**.
2. **Marca de fuga grave** — se le quita el fondo al **texto**:
   `color: var(--crit)` sobre **`background: var(--card)`**, con
   `border: 1px solid color-mix(in srgb, var(--crit) 45%, transparent)`. 4,75:1 claro / 5,33:1
   oscuro: pasa en los dos sin tokens nuevos. `--crit-soft` sobrevive como relleno de superficies
   **sin texto encima** (el punto, el riel). Tickets que lo pintan: **T31** (chip del nodo) y
   **T36** (chip de la fila del Portafolio).
3. **El gate mira los dos temas o mira la mitad.** `.storybook/preview.ts` renderiza en `light` por
   default. Los tres archivos que pintan texto sobre tono bajo un gate `a11y: { test: "error" }`
   llevan **una story espejo en oscuro** con `globals: { theme: "dark" }`:
   `ReposoDark` (T32) · `CompletaB1AtencionDark` (T33) · **`ConDatoDark` (T36, nueva respecto de
   plan-storybook)**. ⇒ el total de stories pasa de **124 a 125**.
4. **`arnes-node.stories.tsx` hereda `a11y: { test: "todo" }`** y **no se amplía** (regla vigente).
   Como ahí axe no corre, `MejoraConMarcaDeFuga` lleva un assert **computado** en su lugar:
   `getComputedStyle(el).backgroundColor` de `.mej-fuga.grave` es el de `--card`, **no** el de
   `--crit-soft`. Es la única forma de que el fix no se pierda en el archivo sin gate.
5. **`--heat-4` vs `--heat-3` es 1,37:1, no 1,20.** No cambia la conclusión —axe no mira contraste
   de no-texto, así que el gate no lo caza— y por eso el assert real de la barra es el `aria-label`
   completo (`CuatroSegmentos`, §2.3) y el separador de 1 px de `--card`. Se corrige el número.

### D22 · CAP-139 apunta a símbolos que el diseño renombró

`capabilities-a-crear.md` CAP-139 lista punteros a
`entities/telemetria/ui/valor-o-sin-dato.tsx#ValorOSinDato`,
`entities/telemetria/ui/chip-confianza.tsx#ChipConfianza` y
`widgets/mejora/ui/tarjeta-mejora.tsx#TarjetaMejora`. **Ninguno de los tres nombres sobrevive** a
`design.md` §1.3-1.4 y `plan-storybook.md` §1.1, que fijan `cifra-usd.tsx#CifraUSD`,
`marca-confianza.tsx#MarcaConfianza` y `punto-mejora-card.tsx#PuntoMejoraCard` — y que **eliminan
la slice `widgets/mejora/`** (todo va en `widgets/map-canvas` para no chocar con
`no-sibling-widget-imports`, que está en `error`). Con los punteros viejos, `TestCapabilityPointerSymbolsResolve` (R1 a nivel símbolo, v1.5) falla.

**Gana `design.md`/`plan-storybook.md`.** CAP-139 se crea (en T33) con estos cuatro punteros:

```
web/src/widgets/map-canvas/model/layers.ts#LAYERS
web/src/entities/telemetria/ui/cifra-usd.tsx#CifraUSD
web/src/entities/telemetria/ui/marca-confianza.tsx#MarcaConfianza
web/src/widgets/map-canvas/ui/punto-mejora-card.tsx#PuntoMejoraCard
```

Es la **contradicción #13** del listado de `arquitectura-modulo.md` §12 (que llega hasta la 12), y
se anota ahí en T2. Misma clase de problema, resuelto igual: `arquitectura-modulo.md` §1 dibuja
`web/src/widgets/mejora/` — **no se crea**.

---

## 3 · Los tickets

### Tramo 0 — Precondiciones (cero código de producto)

---

#### T1 · Firmar las decisiones de construcción y registrar la deuda

**Depende de:** nada.
**Objetivo:** que las cinco resoluciones de §2 dejen de ser propuestas de este plan y pasen a ser
decisiones del paquete, y que las deudas que destapan queden visibles fuera de él.

**Archivos:**

- `docs/product/stories/2026-07-24-telemetria-embebida-otel/decisiones.md` — agregar **D18** (props
  primitivas + story-candado), **D19** (`knowledge` es banda), **D20** (`puesto` sale del `rol`),
  **D21** (contrastes por tema), **D22** (punteros de CAP-139), cada una con el texto de §2 y
  `estado: FIRMADA (plan de desarrollo, 2026-07-26)`.
- `docs/product/BACKLOG.md` — tres ítems nuevos: (a) «`puesto` como faceta de la instalación del
  Portafolio» (D20); (b) «`domain_modules` de `project.config.yaml` está incompleto: 10 módulos con
  hojas de capability no figuran» (drift preexistente, `capabilities-a-crear.md` §0.1); (c) «TTL de
  retención sin número firmado» (J-6 — el `90` del mockup es PROPUESTO).
- `project.config.yaml` — agregar `telemetria` a `domain_modules` con el comentario
  `# telemetría embebida: ingesta OTLP+hook, costeo, detección, capa Mejora (HS-27/28)`.
- `docs/product/stories/2026-07-24-telemetria-embebida-otel/INDEX.md` — «Retomar aquí» apunta a
  este plan y al ticket en curso.

**RF que cierra:** ninguno (habilita todos).
**Escenarios:** ninguno.
**Capabilities:** ninguna (no toca `cmd/`, `internal/`, `web/src`).

**Criterios de aceptación:**

```gherkin
Escenario: las cinco decisiones están escritas y firmadas
  Dado decisiones.md
  Entonces existen los bloques D18, D19, D20, D21 y D22
  Y cada uno dice "FIRMADA"
  Y ninguno contradice el bloque 🧑‍⚖️ del 2026-07-26

Escenario: telemetria es un módulo declarado
  Cuando se lee project.config.yaml
  Entonces domain_modules contiene "telemetria"
```

**V&V:** revisión del operador (es documental). `grep -c "^## D1[89]\|^## D2[012]" decisiones.md` = 5.
**Demostración:** `grep -n "telemetria" project.config.yaml && grep -n "D18\|D19\|D20\|D21\|D22" docs/product/stories/2026-07-24-telemetria-embebida-otel/decisiones.md`
→ 5 bloques + la línea del módulo.

---

#### T2 · Corregir los documentos que hoy dicen algo falso

**Depende de:** T1.
**Objetivo:** que ningún documento del paquete instruya al constructor a hacer algo que rompe un
gate o que describe mal el entorno.

**Archivos y cambio exacto:**

| archivo | qué se corrige |
|---|---|
| `design.md` §8, fila «Storybook + `vitest-browser`» | Sale *«⚠️ no corre en background (Chromium no headless)»*. Entra: *«corre headless — `npx vitest run --project=storybook <archivo>` verificado 2026-07-26, `web/vitest.config.ts` tiene `headless: true` explícito»*. |
| `spec.md` §3, fila **T-21** | Ídem: la nota vieja se deroga y se reemplaza por el comando real. |
| `design.md` §4.2, fila «disclaimer «estimado»» | `texto --warn` → **`texto --foreground`**; `--warn` queda solo en el borde. Nota con las cifras de D21 (3,24:1 claro / 6,11:1 oscuro → 13,14:1). |
| `design.md` §4.2, fila «marca de fuga · grave» | `fondo --crit-soft` → **`fondo --card`**; borde `color-mix(--crit 45%)`. Nota con 4,04:1 → 4,75/5,33:1. |
| `design.md` §4 fin, «Deuda a11y heredada» | Se amplía: la regla «todo texto nuevo sobre `--warn-soft` usa `--foreground`» pasa a ser **normativa de la tabla**, no una nota al pie que la tabla contradice. |
| `design.md` §7.3 | Se **borra** la fila `sin dato · conocimiento` (D19). |
| `spec.md` RF-238, segundo escenario | *«un nodo de clase rule, mcp, knowledge, subagent, hook, command o settings»* → *«un nodo que no es caja (`isCaja(box) === false`)»* (D19). |
| `spec.md` §E RF-265 + §4 | Nota de D20: la fila se agrupa por `(identidad, instalación)` y el puesto sale del `rol` del arnés, `null` ⇒ `puesto sin declarar`. |
| `plan-storybook.md` §2.18 | A11Y-1 y A11Y-2 pasan de «riesgo» a «resuelto en D21», con las cifras por tema. A11Y-3 corrige 1,20 → **1,37 : 1**. §0 y §6 pasan el total 124 → **125** (`ConDatoDark`). |
| `arquitectura-modulo.md` §12 | Se agrega la **contradicción 13** (punteros de CAP-139 y `widgets/mejora/`, D22). |
| `capabilities-a-crear.md` CAP-139 | Punteros corregidos a los cuatro de D22. |
| `docs/architecture/INDEX.md`, fila `telemetria-de-nacimiento` | «🌱 vivo · 1.0 · 2 checks» → el estado real del boundary (v2.3 · 10 checks). Es la contradicción 5 de §12, que el propio documento dice corregir «en este mismo entregable». |

**RF que cierra:** ninguno (desbloquea RF-236, RF-241, RF-238, RF-265 y T-21).
**Escenarios:** ninguno.
**Capabilities:** ninguna.

**Criterios de aceptación:**

```gherkin
Escenario: no queda ninguna nota que diga que vitest-browser no corre headless
  Cuando se busca "no corre en background" o "Chromium no headless" en el paquete
  Entonces no hay ninguna ocurrencia fuera de un bloque marcado como DEROGADO

Escenario: la tabla de tokens no manda pintar un contraste que falla
  Cuando se lee design.md §4.2
  Entonces el disclaimer usa --foreground como color de texto
  Y la marca de fuga grave usa --card como fondo del texto
  Y ninguna fila pone texto --warn sobre --warn-soft ni texto --crit sobre --crit-soft
```

**V&V:** revisión + `grep`. Es la única forma: son documentos.
**Demostración:**

```bash
cd /home/chalreme/Proyectos/harness-studio/docs/product/stories/2026-07-24-telemetria-embebida-otel
grep -rn "Chromium no headless\|no corre en background" . | grep -v DEROGADA   # → vacío
grep -n "warn-soft\|crit-soft" design.md                                        # → ninguna con texto encima
grep -n "conocimiento" design.md                                                # → sin la fila de §7.3
```

---

#### T3 · Cerrar el residuo D8/D14.4: `TestNoJSONLSchemaParsing`

**Depende de:** nada (paralelizable con T1/T2).
**Objetivo:** que el árbol deje de simular una protección que no corre. Hoy
`docs/architecture/fitness/arch_test.go:267` es un `t.Skip` con un TODO — un pass fabricado con otro
nombre, y este paquete lo toca de refilón (agrega dos caminos de ingesta nuevos que **no** leen el
JSONL, y eso hay que poder afirmarlo).

**Archivos:** `docs/architecture/fitness/arch_test.go` (reemplazar el cuerpo de
`TestNoJSONLSchemaParsing`, líneas 267-270).

**Diseño — se enforcea, no se borra.** El test pasa a ser un **source-scan** con exención declarada:

- Recorre `internal/usecase/**`, `internal/domain/**` y `internal/adapters/telemetria/**`.
- Falla si encuentra `json.Unmarshal` / `json.NewDecoder` / `json.RawMessage` aplicado a algo cuyo
  identificador o comentario contenga `transcript`, `jsonl` o `~/.claude/projects`.
- **Exención única y con razón escrita en el propio test:** `internal/adapters/history/**` — es el
  lector de replay del historial B2, ya declarado en `.go-arch-lint.yml`
  (`history: mayDependOn: [domain]`) y en `conductor-no-parsea-jsonl.md`. La exención se lista por
  ruta, no por patrón: agregar un segundo paquete exige tocar el test.
- **Control positivo obligatorio** (§4): el test incluye un fixture en memoria con una llamada que
  **sí** debe detectar y asserta que el detector la marca. Un scanner roto que no encuentra nada
  daría verde con `len(hallazgos) == 0`.

**RF que cierra:** ninguno. **Escenarios:** ninguno de la matriz (es deuda heredada).
**Capabilities:** ninguna (`docs/architecture/fitness/` está fuera de la cobertura R2).

**Criterios de aceptación:**

```gherkin
Escenario: el check corre de verdad
  Cuando se corre TestNoJSONLSchemaParsing
  Entonces NO llama a t.Skip
  Y falla si un caso de uso decodifica el schema del transcript
  Y su control positivo demuestra que el scanner encuentra lo que busca
```

**V&V:** `TestNoJSONLSchemaParsing` (fitness, `docs/architecture/fitness/arch_test.go`).
**Demostración:**

```bash
cd /home/chalreme/Proyectos/harness-studio
go test ./docs/architecture/fitness/ -run TestNoJSONLSchemaParsing -v
# éxito = "--- PASS" y NINGUNA línea "--- SKIP"
```

---

### Tramo A — Los RF sin superficie (Go)

> **Autorizado por el 🧑‍⚖️ del bloque D9·D11·D13·D14·D15·D16.** No espera al mockup.

---

#### T4 · Dominio: el evento canónico y las vistas

**Depende de:** T1.
**Objetivo:** existe **un** tipo que los cuatro emisores producen, con la llave del join, los
buckets como punteros y los dos costos — y no existe ninguna ruta de código que le meta PII.

**Archivos (nuevos):**

- `internal/domain/telemetria.go`
- `internal/domain/telemetria_vistas.go`

**Diseño** — copiar **literal** de `arquitectura-modulo.md` §2.1 (líneas 110-268): los enums
`Emisor` · `Confianza` · `Aritmetica` · `Acumulacion` · `TipoEvento` · `Resultado`, los structs
`Tokens` (los 6 campos **punteros**, `omitempty`) · `LlaveJoin` · `EventoTelemetria`, y los tres
centinelas `ErrPayloadInvalido` · `ErrEventoSinSesion` · `ErrCampoProhibido`. Se **agregan** en el
mismo archivo, porque otros punteros de capability los nombran:

- `Escenario` + `EscenarioS1` / `EscenarioS2Instrumentado` / `EscenarioS2Degradado`
  (§2.4 líneas 325-339) — vive acá y no en `telemetria_deteccion.go` porque CAP-127 lo apunta como
  `internal/domain/telemetria.go#Escenario`.
- `FichaDaemon` (la forma del JSON de §7.2 líneas 1327-1338) — CAP-132 lo apunta acá.
- `UsoDelTurno` — lo que el `result` del stream-json ya trae y hoy se tira (§4.2 líneas 722-726):
  los 4 buckets, el split `ephemeral_5m`/`ephemeral_1h`, `service_tier`, `speed`,
  `modelUsage[].costUSD`, `canonicalModel`, `provider`, `contextWindow`. CAP-136 lo apunta acá.
- `PerfilRuntime` (§4.5 líneas 869-878).

`telemetria_vistas.go`: `ResumenTelemetria` · `Cobertura` · `GastoCaja` · `MarcaDeFuga` ·
`DetalleCaja` · `TurnoUnido` · `SaludTelemetria` · `RespuestaMejoras` · `FilaPortafolio` ·
`VersionCatalogo` — shapes de §3.2 (líneas 621-657) y §8.2 (1522-1529). **`FilaPortafolio` lleva
`Puesto *string` (D20)**, no un `string`.

**Nombres prohibidos, verificados tomados (A4, §2.2):** `origen` · `canal` · `via` · `procedencia`.
Se usa **`emisor`** y **`atribucion`**.

**RF que cierra:** RF-286 (la mitad de dominio) · RF-285 (la forma).
**Escenarios:** C13 · D5 (todo UTC RFC3339).
**Capabilities:** **crea CAP-120** `telemetria/evento-canonico-de-telemetria` — pero **solo cuando
T5 exista**, porque su puntero incluye `internal/ports/telemetria.go#TelemetriaSink`. ⇒ **la hoja se
crea en el commit de T5**; T4 va con la hoja pendiente y T5 la trae. (R2: `telemetria.go` y
`telemetria_vistas.go` quedan huérfanos un commit — **no es aceptable**. ⇒ **T4 y T5 se commitean
juntos**, ver «Demostración».)

**Criterios de aceptación:**

```gherkin
Escenario: un bucket ausente no serializa como cero
  Dado un EventoTelemetria cuyo Tokens.Razonamiento es nil
  Cuando se serializa a JSON
  Entonces la clave "razonamiento" NO aparece en la salida
  Y no aparece con valor 0

Escenario: el evento no tiene dónde poner identidad
  Cuando se inspeccionan por reflexión los campos de EventoTelemetria
  Entonces no existe ningún campo cuyo nombre contenga email, account, organization ni user
  Y no existe ningún campo que contenga contenido de conversación
```

**V&V:**
`TestNoAplicaNoEsCeroEnElWire` (fitness · `docs/architecture/fitness/telemetria_test.go`, se escribe
en T26; su versión mínima sobre el dominio se coloca ya en `internal/domain/telemetria_test.go`) ·
`TestCifraLlevaConfianza` (fitness, T26) · `TestEventoNoTieneCamposDeIdentidad` (colocado,
`internal/domain/telemetria_test.go`, reflexión sobre los campos).

**Demostración:**
```bash
go test ./internal/domain/ -run 'TestNoAplica|TestEventoNoTieneCamposDeIdentidad' -v && go build ./...
```

---

#### T5 · Puertos del módulo

**Depende de:** T4 (se commitea junto con él).
**Objetivo:** los casos de uso pueden hablar de telemetría sin saber que del otro lado hay OTLP,
SQLite o HTTP.

**Archivos (nuevo):** `internal/ports/telemetria.go`.

**Diseño** — literal de `arquitectura-modulo.md` §2.5 (líneas 450-571). Las seis interfaces:
`TelemetriaSink` · `TelemetriaStore` (extiende el sink) · `AtribucionRegistry` ·
`CatalogoPrecios` · `DescubrimientoDaemon` · `ForwardOTLP`; los dos structs de consulta
`ConsultaTelemetria` · `PurgaTelemetria`; los tres centinelas `ErrColaLlena` ·
`ErrAlmacenNoDisponible` · `ErrSinFicha`.

Dos cosas que un implementador «prolijo» rompería y no debe:

- `Ingerir` devuelve `(aceptados int, err error)`. **La diferencia con `len(evs)` son descartes
  contados, no silencio.**
- `Resumen` devuelve punteros `nil`, no ceros, cuando no hubo dato que agregar.

**RF que cierra:** RF-285 (contrato).
**Escenarios:** H13 (la forma del descarte contado).
**Capabilities:** **crea CAP-120** (`docs/product/capabilities/telemetria/evento-canonico-de-telemetria.yaml`),
punteros: `internal/domain/telemetria.go#EventoTelemetria` · `#LlaveJoin` · `#Tokens` ·
`internal/ports/telemetria.go#TelemetriaSink`.
Este es el **primer** archivo del módulo nuevo ⇒ el commit crea también el directorio
`docs/product/capabilities/telemetria/`.

**Criterios de aceptación:**

```gherkin
Escenario: ninguna interfaz nombra su transporte
  Cuando se lee internal/ports/telemetria.go
  Entonces no aparecen las palabras OTLP, sqlite, http ni json en ninguna firma
```

**V&V:** `go build` + `TestCapabilityPointersResolve` + `TestCapabilityPointerSymbolsResolve` +
`TestCapabilityCoverage` (R1/R2, ya existentes en `capability_trace_test.go`).
**Demostración:**
```bash
go build ./... && go test ./docs/architecture/fitness/ -run 'TestCapability(PointersResolve|PointerSymbolsResolve|Coverage)' -v
```

---

#### T6 · Catálogo de precios embebido

**Depende de:** T4, T5.
**Objetivo:** el binario sabe cuánto cuesta un modelo sin pedirle nada a internet, y **dice** con qué
catálogo lo supo.

**Archivos (nuevos):**

- `internal/adapters/telemetria/catalogo/catalogo.go` — implementa `ports.CatalogoPrecios`:
  `Embebido() *Catalogo` · `(*Catalogo).Precio(modeloCanonico) (domain.PrecioModelo, bool)` ·
  `(*Catalogo).Canonizar(modelo) string` · `(*Catalogo).Version() domain.VersionCatalogo`.
- `internal/adapters/telemetria/catalogo/precios.json` — **se commitea** (§8.1: el build es offline
  y reproducible).
- `internal/adapters/telemetria/catalogo/gen.go` — `//go:generate go run ./cmd/bajar-catalogo -rev <sha> -out precios.json`.
- `internal/adapters/telemetria/catalogo/cmd/bajar-catalogo/main.go` — baja
  `model_prices_and_context_window.json` de un **rev fijado** de LiteLLM (MIT, raíz), filtra a los
  proveedores en uso y compacta.
- `internal/adapters/telemetria/catalogo/LICENSE-litellm.md` — MIT + aviso de atribución.
- `internal/domain/telemetria_costo.go` — solo `PrecioModelo` acá (el resto en T7), porque
  `catalogo.go` lo necesita.

**Diseño** — `arquitectura-modulo.md` §8 (líneas 1499-1552). Puntos que no son opcionales:

- `PrecioModelo` tiene **punteros** y un `SobreUmbral *PrecioModelo`: **los tiers no se aplanan**
  (es el bug `phoenix#14314`). Copiar `cache_creation_input_token_cost_above_1hr` de LiteLLM — el
  campo que ccusage descarta y que habilita B1/B12 (E3).
- Alias Bedrock/Vertex: se porta el **regex** de Langfuse (MIT, E6) para normalizar
  `(eu.|us.|apac.)?anthropic.claude-…`. **Un modelo que no matchea se devuelve tal cual** —
  `Canonizar` no inventa una canonización, y `Precio` devuelve `ok=false`.
- `VersionCatalogo` con `Version` (fecha del rev), `Rev` (sha), `SHA256` del archivo embebido,
  `Modelos` y `Refrescado *time.Time` — **`nil` = nunca, y se muestra como tal** (§8.3).
- **Presupuesto:** el `go:embed` ≤ **256 KB** (182 KB medidos filtrado).

**RF que cierra:** RF-272 · RF-273 (el dato que las alimenta).
**Escenarios:** F1 (modelo desconocido) · F2 (tarifa parcial) · F3 (catálogo viejo) · F9 (tiers).
**Capabilities:** **crea CAP-125** `telemetria/catalogo-de-precios-embebido`.

**Criterios de aceptación:**

```gherkin
Escenario: un modelo desconocido no vale cero
  Dado un catálogo embebido
  Cuando se pide el precio de "modelo-que-no-existe"
  Entonces devuelve ok=false
  Y NO devuelve un PrecioModelo con todos los campos en cero

Escenario: el tier largo sobrevive al embebido
  Dado un modelo de LiteLLM con cache_creation_input_token_cost_above_1hr
  Cuando se lee del catálogo embebido
  Entonces SobreUmbral no es nil
```

**V&V:** `TestModeloDesconocido` · `TestTarifaParcial` · `TestAliasBedrockVertex` ·
`TestCatalogoEmbebidoBajoPresupuesto` (todos colocados en
`internal/adapters/telemetria/catalogo/catalogo_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/catalogo/... -v
ls -l internal/adapters/telemetria/catalogo/precios.json   # ≤ 262144 bytes
```

---

#### T7 · Costeo de doble fuente (los tres bugs ajenos)

**Depende de:** T6.
**Objetivo:** cotizar un uso con nuestro catálogo sin reproducir ninguno de los tres modos de fallo
que **toda** la industria tiene (D9.6), y conservar lo que dijo el runtime al lado.

**Archivos:** `internal/domain/telemetria_costo.go` (completar).

**Diseño** — §2.3 (líneas 288-320). `CostoCalculado{Micros, Completo, SinTarifa []string, Version}`
y `CalcularCosto(t Tokens, p PrecioModelo, a Aritmetica) CostoCalculado`, **pura**.

Las tres reglas que los tests dictan, escritas como invariantes del código:

1. **El cache WRITE se cobra** (`langfuse#14249` lo olvida → −28 %).
2. **No se suman buckets que se solapan** — de ahí el parámetro `a Aritmetica`: con
   `AritmeticaInclusiva`, `cache_lectura ⊂ entrada` y restar es obligatorio (`langfuse#12306` → 2×).
3. **Los tiers no se aplanan**: si hay `SobreUmbral` y el prompt pasa `UmbralContextoTok`, se usa
   ese (`phoenix#14314`).

Y la cuarta, que es de honestidad: **un bucket con tokens y sin tarifa NO se cobra a 0** — sale en
`SinTarifa` y `Completo=false`.

**RF que cierra:** RF-261 · RF-285 (el costo calculado).
**Escenarios:** F2 · F6 · F7 · F8 · F9 · F10 · F11.
**Capabilities:** **crea CAP-126** `telemetria/costeo-de-doble-fuente`.

**Criterios de aceptación:**

```gherkin
Escenario: el cache write se cobra
  Dado tokens con cache_escritura_1h > 0 y un precio con tarifa para ese bucket
  Entonces el costo incluye ese bucket
  Y el costo es estrictamente mayor que el mismo cálculo sin él

Escenario: aritmética inclusiva no cuenta dos veces
  Dado un adaptador con Aritmetica = inclusive
  Y tokens donde cache_lectura está contenido en entrada
  Entonces el costo NO suma cache_lectura como si fuera adicional

Escenario: sin tarifa no es gratis
  Dado tokens con razonamiento > 0 y un precio sin tarifa de razonamiento
  Entonces Completo es false
  Y SinTarifa contiene "razonamiento"
  Y el bucket NO se cobró a 0 en silencio
```

**V&V:** `TestCosteoCobraElCacheWrite` · `TestCosteoNoSumaBucketsQueSeSolapan` ·
`TestCosteoNoAplanaLosTiers` · `TestSinNingunCosto` (colocados,
`internal/domain/telemetria_costo_test.go`).
**Demostración:**
```bash
go test ./internal/domain/ -run TestCosteo -v
```

---

#### T8 · Decodificador OTLP/JSON

**Depende de:** T4.
**Objetivo:** entender el wire de OTLP/JSON con **+0,49 MB** de binario y sin reventar con el
`intValue` off-spec que Claude Code emite.

**Archivos (nuevos):**

- `internal/adapters/telemetria/otlp/decodifica.go` — `DecodificarLogs([]byte) ([]RegistroLog, error)`
  y `DecodificarMetricas([]byte) ([]PuntoMetrica, error)`.
- `internal/adapters/telemetria/otlp/testdata/` — **enlaces/copias** de
  `verificacion-2026-07-26/evidencia/logs-run1.json`, `logs-run2-con-skill.json`,
  `logs-run4-con-tools.json`, `metrics-run1.json`: son los **golden files** (8 métricas · 14 puntos ·
  57 log records).

**Diseño** — §4.3 (líneas 737-834), literal. Los structs `valorAtributo` · `atributo` ·
`peticionLogs` · `peticionMetricas` · `puntoDato`, tal cual.

Cuatro reglas que son la diferencia entre que funcione y que no:

1. **`json.Number` en `IntValue` y `AsInt` no es estilo, es obligatorio.** Claude Code emite
   `intValue` como número JSON, off-spec; un runtime que cumple la spec lo emite como string.
   `json.Number` acepta las dos y convierte con `.Int64()`. **Verificado en vivo, V5.**
2. **Lo desconocido se ignora, no es error.** `encoding/json` lo descarta y el evento entra. Es lo
   que hace que una versión nueva del runtime no rompa el receptor.
3. **Temporalidad:** las 4 métricas llegan `Sum · Delta · monotonic` (V5.1) ⇒ **se suma y listo**.
   Si llega `AggregationTemporality == 2` (Cumulative) **no se adivina**: el punto se marca
   `sin-dato` y suma a `salud.temporalidad_no_soportada`. Fabricar un delta sin estado previo es
   inventar.
4. **Lo que se ignora explícitamente:** `Gauge`/`Histogram`/`ExponentialHistogram` ·
   `traceId`/`spanId` · `severityNumber`/`severityText` · `droppedAttributesCount` · `schemaUrl` ·
   `arrayValue`/`kvlistValue`/`bytesValue` · **`body.stringValue`** (los eventos útiles viven en
   atributos; el cuerpo puede traer texto).

**RF que cierra:** RF-282 (la puerta por donde entra).
**Escenarios:** C1 · C3 · C4 · C5 · C6 · C8 · C9 · C10 · C14 · C15.
**Capabilities:** **crea CAP-119** `telemetria/decodificador-otlp-json`.

**Criterios de aceptación:**

```gherkin
Escenario: el mismo payload con las dos formas de intValue
  Dado el payload real de evidencia/logs-run1.json
  Y una copia con cada "intValue": N reescrito como "intValue": "N"
  Cuando se decodifican los dos
  Entonces producen exactamente el mismo resultado

Escenario: un lote malo no se guarda a medias
  Dado un JSON sintácticamente roto
  Entonces DecodificarLogs devuelve ErrPayloadInvalido
  Y no devuelve ningún registro parcial

Escenario: Cumulative no se adivina
  Dado un punto con aggregationTemporality = 2
  Entonces el punto queda marcado sin-dato
  Y NO se convierte a delta
```

**V&V:** `TestIntValueComoNumeroYComoString` · `TestCamposDesconocidosNoRompen` ·
`TestPayloadMalformado` · `TestAtributoNumericoIlegible` · `TestSumaDelta` ·
`TestCumulativeNoSeAdivina` · `TestReinicioDelEmisorNoDuplica` · **`FuzzDecodificarLogs`** (sembrado
con los 3 payloads reales) — todos colocados en
`internal/adapters/telemetria/otlp/decodifica_test.go`.
**Demostración:**
```bash
go test ./internal/adapters/telemetria/otlp/ -v
go test ./internal/adapters/telemetria/otlp/ -run FuzzDecodificarLogs -fuzz FuzzDecodificarLogs -fuzztime 60s
```

---

#### T9 · Mapeo `claude_code.*` → evento canónico, con allowlist (camino OTLP)

**Depende de:** T8.
**Objetivo:** que del wire salga un `EventoTelemetria` **sin una sola de las cinco claves de
identidad**, campo por campo, sin ninguna ruta que copie un mapa entero.

**Archivos (nuevo):** `internal/adapters/telemetria/otlp/mapa_cc.go` —
`MapearLogRecord(RegistroLog, PerfilRuntime) (domain.EventoTelemetria, error)` y
`MapearPuntoMetrica(...)`.

**Diseño** — la tabla de §4.4 (líneas 836-859) **entera**, y la allowlist de §6.1 (1137-1146).

- `session.id`→`SesionID` · `prompt.id`→**`TurnoID`** (la otra mitad de la llave, H1) ·
  `arnesia.arnes|.instalacion|.caja|.corrida`→`LlaveJoin` (⇒ `Atribucion=exacta`) · `model` ·
  los 4 buckets · `cost_usd_micros`→`CostoReportadoMicros` (entero, sin coma flotante) ·
  `duration_ms` · `speed` · `service_tier`.
- `CacheEscritura5m`/`1h` quedan **nil** acá: el split no viene por OTel.
- `plugin_loaded` → `PluginIDHash` + `plugin.scope`.
- `tool_decision` → `TipoEvento=EventoHerramienta`, `Resultado`, `Herramienta`, `Decision` (H8).
- `tool_result` → `+ ToolInputBytes`/`ToolResultBytes` — **A21: se persisten desde el día 1** aunque
  B11 esté fuera del MVP. Son **bytes**, no contenido, y un dato que no se guarda hoy no se
  recupera mañana.
- `hook_execution_start`/`_complete` → **ignorados en el MVP** (los cubre el hook propio).
- **Todo el resto → descartado en la puerta. Default-deny.**
- `AdaptadorVersion = "cc-otlp/1"`. **Cambiar el mapeo obliga a bumpear ese string** (D7.5).

**Forma del enforcement (§6.1 líneas 1173-1178):** `domain.NuevoEvento(campos map[string]any)`
**no existe**. El evento se arma campo por campo. No hay ruta donde un mapa entero se copie.

**RF que cierra:** **RF-282** (camino OTLP).
**Escenarios:** C16 · E6 · I10.
**Capabilities:** CAP-121 se crea en **T10** (su puntero incluye `hooks/proyecta.go#Proyectar`).

**Criterios de aceptación:**

```gherkin
Escenario: la identidad no cruza la puerta
  Dado un payload OTLP con user.email, user.account_uuid, user.account_id, user.id y organization.id
  Cuando se mapea
  Entonces el evento resultante no contiene ninguno de esos valores en ningún campo

Escenario: un atributo nuevo cae afuera y se cuenta
  Dado un payload con un atributo "cosa.nueva"
  Entonces el evento entra con lo declarado
  Y el descarte incrementa un contador de salud
```

**V&V:** `TestAllowlistNoPersistePII` (fitness, T26 — la parte de subcadena sobre el `.db` llega en
T11) · `TestAllowlistEsListaNoSugerencia` (fitness, source-scan) · `TestMapeoCCCompleto` (colocado,
`mapa_cc_test.go`, contra los 3 golden files) · `TestToolResultBytesSePersiste` (fitness, T26).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/otlp/ -run TestMapeo -v
```

---

#### T10 · Proyección del hook, con allowlist (camino 2)

**Depende de:** T4.
**Objetivo:** que el payload del hook —que trae **el prompt, la respuesta y la salida de las
herramientas en claro** (H4)— se reduzca a un puñado de campos declarados **antes** de escribir a
ningún lado, loopback incluido.

**Archivos (nuevo):** `internal/adapters/telemetria/hooks/proyecta.go` —
`Proyectar(stdin []byte, resolverCWD func(string) (arnesID, instalacionID string, ok bool)) (domain.EventoTelemetria, error)`.

**Diseño** — §6.1 (1148-1171) y §7.1. Se persiste **solo**: `session_id` · `prompt_id` ·
`hook_event_name` · `tool_name` · `duration_ms` · `permission_mode` · `reason` (de `SessionEnd`) ·
`cwd` **→ normalizado a `instalacion_id` o a huella (A14: la ruta cruda NUNCA se guarda)**.

Se descarta, nombre por nombre: `prompt` · `last_assistant_message` · `tool_response` ·
`tool_input` · `transcript_path` (**y que nadie asuma que es un archivo** — en la corrida apuntó a
un directorio, H7) · `cwd` crudo · `background_tasks` · `session_crons` · **cualquier campo no
listado**.

Mapeo de `hook_event_name` → `TipoEvento`: `UserPromptSubmit`→`turno_inicio` ·
`PostToolUse`→`herramienta` · `Stop`→`turno_fin` · `SessionEnd`→`sesion_fin`. `PreToolUse` y
`SessionStart` **no se instrumentan en el MVP**: no aportan al join y sí superficie.

**RF que cierra:** **RF-282** (camino hook).
**Escenarios:** E7 · E8 · I9 · I10.
**Capabilities:** **crea CAP-121** `telemetria/allowlist-de-ingesta` — plantilla completa ya escrita
en `capabilities-a-crear.md` §4, se copia tal cual con las fechas reales.

**Criterios de aceptación:**

```gherkin
Escenario: el contenido no sobrevive a la proyección
  Dado el payload real de UserPromptSubmit con un prompt que contiene el marcador MARCA-XYZ-123
  Cuando se proyecta
  Entonces el evento serializado no contiene la subcadena MARCA-XYZ-123
  Y sí contiene session_id y prompt_id  ← control positivo, misma corrida

Escenario: el cwd se usa y se tira
  Dado un payload con cwd = /home/<usuario>/Proyectos/vitalia que mapea a una instalación conocida
  Entonces el evento lleva instalacion_id y Atribucion = por-proceso
  Y no lleva la ruta cruda en ningún campo
```

**⚠ Control positivo (§4):** los dos escenarios de arriba son «no aparece X». **Ninguno se asserta
solo con la ausencia**: en la misma corrida se asserta que un campo de la allowlist **sí** está.

**V&V:** `TestHookNoReenviaContenido` (fitness, T26, con los 6 payloads reales del ANEXO) ·
`TestCwdDesconocidoNoSeGuardaCrudo` (fitness) · `TestProyeccionPorEvento` (colocado,
`hooks/proyecta_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/hooks/ -v
go test ./docs/architecture/fitness/ -run 'TestCapability(PointersResolve|PointerSymbolsResolve)'
```

---

#### T11 · Almacén SQLite: DDL, migración aditiva y archivado

**Depende de:** T5, T9, T10.
**Objetivo:** una base **propia**, que no se borra nunca en silencio, con el índice del join.

**Archivos (nuevos):**

- `internal/adapters/telemetria/store/store.go` — `Store` + `New(ruta string, o Opciones) (*Store, error)`,
  implementa `ports.TelemetriaStore`.
- `internal/adapters/telemetria/store/migracion.go` — `Migracion{Version int; SQL []string}`,
  `Aplicar(db *sql.DB) error`, `const generacionActual = 1`.

**Diseño** — §5 completo (líneas 890-1099).

- **A2: `~/.arnesia/telemetria.db`, jamás dentro de `index.db`.** El índice es **desechable** (un
  bump de esquema lo borra entero); esto **no es reconstruible**. Compartir archivo significaría que
  un bump del índice borra la historia de telemetría.
- **DDL literal de §5.2** (líneas 907-1045): `schema_meta` · `evento` · los 4 índices
  (`idx_evento_join` sobre `(sesion_id, turno_id)` es **EL** índice del módulo) · `rollup_hora`
  (`STRICT, WITHOUT ROWID`) · `rollup_cursor` · `atribucion_hash` · `turno_esperado` · `salud`.
  **Los tokens son `INTEGER` nullable. Jamás `DEFAULT 0`.**
- **Concurrencia (§5.3):** el patrón de `internal/adapters/index/store.go`, **calcado** — writer con
  `SetMaxOpenConns(1)` y reader pooled bajo WAL; DSN idénticos
  (`?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate`
  para el writer). Una goroutine drena la cola y escribe en lotes (**256 eventos o 250 ms**).
- **Dedupe** por `(sesion_id, turno_id, tipo_evento, ts_emisor)` en el writer (C11).
- **A3 · dos números:** `version` crece por migraciones **aditivas** (`CREATE TABLE`, `CREATE INDEX`,
  `ALTER TABLE … ADD COLUMN`); `generacion` cambia solo ante una ruptura no expresable así, y
  entonces el store **archiva** (`telemetria-gen<N>-<fecha>.db.archivada` + `-wal`/`-shm`), arranca
  vacío **y lo dice** en `salud.historia_archivada`. **Nunca borra.**
- **DB corrupta (H1):** mismo trato — `…-corrupta-<fecha>.db` y una nueva vacía.
- **A10 · dos relojes:** `ts_recibido` (el del daemon, **manda**) y `ts_emisor`. Un `ts_emisor`
  fuera de `[ts_recibido − 24 h, ts_recibido + 5 min]` marca `reloj_sospechoso=1`.
- **A19:** `escenario` es **columna**, no un derivado de consulta.

**RF que cierra:** RF-286 (la persistencia) · RF-285 (los dos costos en la fila).
**Escenarios:** C11 · C12 · D1 · D2 · D3 · H1 · H2 · H3 · H6 · H7 · H8 · H9 · B10 · B11.
**Capabilities:** **crea CAP-122** `telemetria/almacen-de-telemetria`.

**Criterios de aceptación:**

```gherkin
Escenario: un wipe del índice no toca la telemetría
  Dado un telemetria.db con eventos y un index.db
  Cuando el índice se borra por mismatch de schema_version
  Entonces telemetria.db sigue existiendo con los mismos eventos

Escenario: una ruptura de generación archiva
  Dado un telemetria.db de generación 1 y un binario de generación 2
  Cuando se abre el store
  Entonces el archivo viejo sigue en disco, renombrado a .archivada
  Y salud reporta su nombre
  Y la base nueva arranca vacía

Escenario: la migración destructiva no compila el merge
  Dado una migración cuyo SQL contiene DROP TABLE
  Entonces TestMigracionesSonAditivas falla
```

**V&V:** `TestTelemetriaDBNoEsElIndice` (fitness) · `TestMigracionesSonAditivas` ·
`TestGeneracionArchivaNoBorra` · `TestDBCorruptaSeArchiva` · `TestDosEscritoresSobreElMismoDB` ·
`TestDiscoLlenoNoTumbaElDaemon` · `TestRelojHaciaAtrasNoRompeLaVentana` · `TestTodoEnUTC` ·
`TestEventoDuplicadoNoSeCuentaDosVeces` · `TestEventoSinSesionSeRechaza` · `TestRutasPorHome`
(colocados en `store/store_test.go` y `store/migracion_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/store/... -race -v
```

---

#### T12 · Rollup horario incremental

**Depende de:** T11.
**Objetivo:** que el tablero cueste ~2 ms en vez de ~323 ms (**190×** medido) **sin que la
agregación convierta un `NULL` en un `0`**.

**Archivos (nuevo):** `internal/adapters/telemetria/store/rollup.go` — `Actualizar(ctx) error` ·
`Recomputar(ctx, horas []string) error`.

**Diseño** — §5.5 (líneas 1101-1126). El `INSERT … SELECT … ON CONFLICT DO UPDATE` literal, con la
suma que **preserva el `NULL`**:

```
tok_entrada = CASE WHEN rollup_hora.tok_entrada IS NULL AND excluded.tok_entrada IS NULL
                   THEN NULL ELSE COALESCE(…,0)+COALESCE(…,0) END
```

- Corre **tras cada lote escrito con debounce de 2 s** y **al boot**.
- `rollup_cursor` guarda el último `evento.id` agregado ⇒ reanudable tras una caída.
- **Recomputo** solo tras borrado por arnés o archivado.
- **Sesgo declarado y en contra (A3):** `COUNT(DISTINCT turno_id)` sobreestima si un turno cruza
  la frontera horaria — **infla el denominador, o sea baja el costo-por-turno que mostramos**. Va
  escrito en el código y en la tarjeta.
- **Cardinalidad (§11):** ≤ **50 000** filas/mes. Al superarlo, `caja_id` colapsa a `(otros)` y la
  fila se marca `cardinalidad_colapsada` — degradación **declarada**, no silenciosa.

**RF que cierra:** RF-286 (que el «no aplica» sobreviva a la agregación).
**Escenarios:** D4 · H10 · H12.
**Capabilities:** **crea CAP-123** `telemetria/rollup-horario-incremental`.

**Criterios de aceptación:**

```gherkin
Escenario: el NULL sobrevive a SUM
  Dado una hora donde ningún evento tuvo tok_cache_1h
  Cuando se actualiza el rollup
  Entonces rollup_hora.tok_cache_1h de esa fila es NULL
  Y NO es 0

Escenario: el rollup reanuda
  Dado un rollup interrumpido a mitad
  Cuando se vuelve a correr
  Entonces agrega solo desde rollup_cursor.ultimo_evento
  Y ninguna fila se cuenta dos veces
```

**V&V:** `TestNoAplicaSobreviveAlRollup` (fitness) · `TestRollupReanudaDesdeElCursor` ·
`TestRollupColapsaCardinalidad` · `TestTurnoCruzaHora` (aserta **la dirección** del sesgo) —
colocados en `store/rollup_test.go`.
**Demostración:**
```bash
go test ./internal/adapters/telemetria/store/ -run TestRollup -v
go test ./internal/adapters/telemetria/store/ -bench BenchmarkRollup -benchtime 3x
```

---

#### T13 · El servicio: ingesta, atribución, escenario, join y conciliación

**Depende de:** T7, T11, T12.
**Objetivo:** que las filas se conviertan en la frase del producto — *«este arnés, en este puesto,
quema $X»* — con cada número cargando **cómo** se atribuyó.

**Archivos (nuevo):** `internal/usecase/telemetria_service.go` —
`NewTelemetriaService(store, catalogo, atribucion, arnesReg, pfStore, detectores, reloj) *TelemetriaService`
con métodos `Ingerir` (implementa `ports.TelemetriaSink`) · `Atribuir` · `Conciliar` · `Resumen` ·
`PorCaja` · `DetalleCaja` · `Turnos` · `Portafolio` · `Salud` · `Forward`.

**Diseño** — §6.3, §6.4, §7.0.2 y §3.2.

- **Orden de atribución (§6.3), el primero que acierta fija `atribucion`:**
  `exacta` (los `arnesia.*` que inyectamos, S1) → `por-hash` (`plugin_id_hash` en
  `atribucion_hash`; da el **arnés**, no la caja) → `por-proceso` (`cwd` → instalación conocida del
  Portafolio, identidad `(home,id)`; da arnés + instalación, no la caja) → `sin-dato`.
- **A15: un evento `sin-dato` se guarda y NO se suma al total.** Aparece en `Cobertura.SinDato` y en
  el drill-down. Sumarlo sería atribuir por adivinanza.
- **El escenario se DERIVA, no se declara** (§7.0.2): hay `arnesia.corrida` o el evento vino por
  `streamjson` ⇒ `s1` · hay `api_request` sin corrida nuestra ⇒ `s2-instrumentado` · solo hay
  eventos de hook ⇒ `s2-degradado`. **Un arnés no puede mentir sobre su propio nivel de
  instrumentación.**
- **El join** es `WHERE sesion_id = ? AND turno_id = ?` — igualdad de dos campos, **sin heurística
  de tiempo ni de orden** (H1).
- **Conciliación (A9):** el denominador sale de `turno_esperado`, que escribe el daemon (T22). En
  **S2 no hay denominador independiente**: la cobertura reporta `esperados: null` y la UI dice
  «cobertura desconocida fuera de ArnesIA», que es la verdad.
- **`Portafolio()` resuelve el `puesto` desde el `rol` del arnés indexado (D20)**; `nil` cuando no
  lo hay.
- **La confianza de un agregado es la MÍNIMA de sus partes**, nunca la máxima ni la moda.
- `AtribucionRegistry` se implementa acá sobre la tabla `atribucion_hash`, y **aprende del spawn
  controlado** (`como_se_aprendio="spawn-controlado"`). **No se asume determinismo entre máquinas**
  (V7.2 sin verificar).

**RF que cierra:** RF-285 · RF-237 (el dato de los 4 segmentos) · RF-274 (por huella) · RF-265 (el
`puesto`).
**Escenarios:** A6 · E1 · E2 · E3 · E5 · E6 · E7 · E8 · E9 · E10 · E11 · A7 · B7 · G1.
**Capabilities:** **crea CAP-127** `telemetria/atribucion-y-confianza` y **CAP-128**
`telemetria/conciliacion-de-cobertura`.

**Criterios de aceptación:**

```gherkin
Escenario: el join es un par
  Dado un api_request y un Stop con el mismo (sesion_id, prompt_id)
  Entonces producen UN TurnoUnido con dinero y proceso
  Y el join no usa ninguna heurística de tiempo

Escenario: lo no atribuido no infla el total
  Dado 3 eventos con atribucion = sin-dato y 5 exactos
  Entonces el costo del resumen suma solo los 5
  Y Cobertura.SinDato es 3
  Y Cobertura.Esperados no es la suma de los medidos

Escenario: el escenario se deriva de la señal
  Dado tres lotes: con arnesia.corrida · api_request sin corrida · solo hook
  Entonces producen s1, s2-instrumentado y s2-degradado respectivamente
  Y ningún campo del emisor pudo elegirlo
```

**V&V:** `TestJoinPorSesionYTurno` · `TestConciliacionCuentaLosNoLlegados` ·
`TestSinDatoNoSumaAlTotal` · `TestAtribucionPorHash` · `TestAtribucionPorProceso` ·
`TestHashDesconocido` · `TestConfianzaDeAgregadoEsLaMinima` · `TestEscenarioSeDerivaDeLaSenal` ·
`TestTurnoSoloConProceso` · `TestTurnoSoloConDinero` · `TestDosHomesNoSeFusionan` ·
`TestInstalacionesNoSeSumanSolas` · `TestGastoSinCaja` · `TestDobleCostoSePersisteEntero` ·
`TestPuestoSaleDelRolDelArnes` (D20) — colocados en `internal/usecase/telemetria_service_test.go`;
los marcados en §10.5 de arquitectura como fitness se duplican allá en T26.
**Demostración:**
```bash
go test ./internal/usecase/ -run TestTelemetria -race -v
go test ./internal/usecase/ -run 'TestJoin|TestConciliacion|TestSinDato|TestAtribucion|TestEscenario' -v
```

---

#### T14 · Retención y borrado

**Depende de:** T13.
**Objetivo:** que la telemetría caduque sola y que el botón «borrar la telemetría de este arnés»
borre **también el agregado** — o quedaría una cifra huérfana alimentándose de nada.

**Archivos (nuevo):** `internal/usecase/telemetria_retencion.go` — `Purgar(ctx, p ports.PurgaTelemetria) (int64, error)`
y `BorrarArnes(ctx, arnesID string) (int64, error)`.

**Diseño** — §5.5 «Recomputo», §11 y escenarios H4/H5/H11.

- **TTL default: 90 días** de `evento`, **24 meses** de `rollup_hora`. ⚠ **El `90` es un valor
  PROPUESTO, no firmado** (J-6): es un flag (`--telemetria-retencion`, T25), la UI lo lee de la
  config y **nunca lo hardcodea**.
- Purga **al boot y cada 6 h**.
- **Borrado por arnés:** `DELETE FROM evento WHERE arnes_id=?` + `DELETE FROM rollup_hora WHERE arnes_id=?`
  + re-agregado de las horas afectadas, **todo en una transacción**.
- Tras purgar, el drill-down de un turno viejo dice «detalle purgado, resumen conservado» — el
  rollup sobrevive.
- Aviso de tamaño a partir de **500 MB** (`salud.tamano_bytes`).

**RF que cierra:** **RF-283**.
**Escenarios:** H4 · H5 · H11.
**Capabilities:** **crea CAP-124** `telemetria/retencion-y-borrado`.

**Criterios de aceptación:**

```gherkin
Escenario: borrar borra también el agregado
  Dado telemetría de dos arneses, con rollup poblado
  Cuando se borra la del arnés A
  Entonces no quedan eventos ni filas de rollup_hora de A
  Y las de B siguen intactas   ← control positivo
  Y el resumen de A vuelve al estado 1

Escenario: el TTL no está hardcodeado
  Dado retencion = 45 días
  Cuando se consulta salud
  Entonces reporta 45, no 90
```

**V&V:** `TestPurgaRespetaTTL` · `TestBorradoPorArnesTambienLimpiaRollup` · `TestAvisoDeTamano` ·
`TestRetencionNoEsUnaConstante` (colocados, `internal/usecase/telemetria_retencion_test.go`).
**Demostración:**
```bash
go test ./internal/usecase/ -run 'TestPurga|TestBorrado|TestAviso|TestRetencion' -v
```

---

#### T15 · Motor de detectores + los seis del MVP

**Depende de:** T13.
**Objetivo:** que cada detector **declare si puede correr** según la señal real de la ventana, y que
el «no aplica» viaje con motivo obligatorio junto a los «no medidos todavía».

**Archivos (nuevos):**

- `internal/domain/telemetria_deteccion.go` — `Detector` (interfaz) · `ContextoDeteccion` ·
  `Aplicabilidad` · `PuntoDeMejora` · `Ventana` · `DetectorID` + los 6 ids · `DetectoresMVP()` y
  las seis implementaciones.
- `internal/usecase/telemetria_mejoras.go` — `(*TelemetriaService).Mejoras(ctx, q) (domain.RespuestaMejoras, error)`.

**Diseño** — §2.4 (líneas 324-443) y §9 (1558-1590), literales.

- `Aplica(c ContextoDeteccion)` se consulta **SIEMPRE** antes que `Evaluar(v Ventana)`.
  **`Motivo` es obligatorio cuando `Aplica` es false.**
- La tabla de aplicabilidad de los seis (§2.4 líneas 422-429) es normativa, incluidos los dos
  matices: **B1 no aplica en `s2-instrumentado` pero por otra razón** (falta el `result`, no falta
  telemetría) — la regla es `TieneSplitTTL`, un hecho de la **ventana**, así que el día que Claude
  Code mande el split por OTel el detector se enciende solo, sin tocar código; y **P1 corre parcial
  en `s2-instrumentado`** (detecta reintentos y fracasos de herramienta vía `tool_result`, no
  rechazos de gate) ⇒ viaja `cobertura_parcial: true` + motivo, **no un ✅ liso**.
- **`PuntoDeMejora` sin las cinco cosas no se muestra** (regla A4): número, contrafactual, umbral
  citado, sesgo declarado **en contra** y **UN** fix.
- **B1** cita la desigualdad entera: `(2−1,25)/(2−0,1) = 39,47 %`. Es **álgebra**, no medición.
- `Mejoras()` devuelve **tres listas siempre** (A16): `puntos[]` · `no_aplican[]` con motivo ·
  `no_medidos[]` con los otros 7 y `motivo: "no medido todavía"`. Omitir `no_aplican` obligaría al
  FE a elegir entre no mostrar nada (gap escondido) o mostrar 0 (mentira).
- `ScoreVersion` en cada punto (A7/RF-254).

**RF que cierra:** RF-246 · RF-249…RF-253 (el dato) · RF-263 · RF-271.
**Escenarios:** G2 · G3 · G4 · G7 · G8 · G9 · G10 · G11 · G12 · G13 · G14.
**Capabilities:** **crea CAP-129** `telemetria/motor-de-detectores` y **CAP-130**
`telemetria/los-seis-detectores-del-mvp`.

**Criterios de aceptación:**

```gherkin
Escenario: ningún detector se apaga sin decir por qué
  Dado un ContextoDeteccion vacío
  Cuando se consulta Aplica en los seis
  Entonces los seis devuelven Aplica=false
  Y los seis tienen Motivo distinto de cadena vacía

Escenario: s2-degradado apaga el dinero con motivo, no con cero
  Dado una ventana con solo eventos de hook
  Entonces B4, B6 y B3 salen en no_aplican con motivo
  Y el resumen trae costo_reportado y costo_calculado en null
  Y ninguno vale 0

Escenario: s2-instrumentado tiene dinero y no tiene split
  Dado una ventana con api_request y sin result del stream-json
  Entonces B4, B6 y B3 aplican
  Y B1 no aplica, con el motivo del split
  Y ese motivo NO es "corrió fuera de ArnesIA"
```

**V&V:** `TestDetectorQueNoAplicaTraeMotivo` · `TestNoMedidosSeDeclaran` · `TestB1BreakEvenTTL` ·
`TestB2CostoDeLaRotacion` · `TestB3CambioDeModelo` · `TestB6SesionAbandonada` ·
`TestP1CajaQueSeRechaza` · `TestP1ParcialSinGate` · `TestS2DegradadoApagaLosDetectoresDeDinero` ·
`TestS2InstrumentadoTieneDineroYNoTieneSplit` · `TestCorridaCancelada` · `TestReintentosSuman` ·
`TestSesgoTieneDireccion` (RF-252: `direccion_sesgo` existe y **no es neutro**).
**Demostración:**
```bash
go test ./internal/domain/ -run 'TestDetector|TestB[1236]|TestP1|TestSesgo' -v
go test ./internal/usecase/ -run 'TestS2|TestNoMedidos' -v
```

---

#### T16 · El receptor OTLP

**Depende de:** T9, T13.
**Objetivo:** un `http.Handler` que reciba OTLP/JSON y **jamás bloquee al emisor** — un receptor que
hace esperar al agente degradaría el trabajo que mide.

**Archivos (nuevo):** `internal/adapters/telemetria/otlp/receptor.go` — `Receptor` ·
`NewReceptor(sink ports.TelemetriaSink, o Opciones) *Receptor` · `(*Receptor).ServeHTTP`.

**Diseño** — §4.1 (líneas 675-709), punto por punto:

1. Solo `POST /v1/logs` y `POST /v1/metrics`; cualquier otra cosa **405**.
2. Content-Type solo `application/json`. `application/x-protobuf` ⇒ **415 con un cuerpo que nombra
   la causa y el fix** (*«este receptor habla OTLP/JSON; poné `OTEL_EXPORTER_OTLP_PROTOCOL=http/json`»*).
   **Jamás un 200 que finge haber guardado.**
3. `http.MaxBytesReader` con `maxBody = 4 MiB` ⇒ **413 sin leer el cuerpo entero a memoria**.
4. Error de forma ⇒ **400** + `ErrPayloadInvalido`; **el lote entero se rechaza**, nunca a medias.
5. **Encolado no bloqueante** (cola 4096). Llena ⇒ **descarta y cuenta**, y lo dice:
   `{"partialSuccess":{"rejectedLogRecords":N,"errorMessage":"cola llena"}}` con **200**.
6. **Nunca `5xx` por un evento malo**: un 5xx hace reintentar al exportador y multiplica el daño.
   `recover()` en el handler solo como red de seguridad (500 + `slog.Error`, **el daemon sigue vivo**).

Presupuestos (§11): **p99 ≤ 5 ms** para decode+encolar, **cero I/O de disco en el handler**.

**RF que cierra:** RF-282 (la puerta HTTP).
**Escenarios:** C1 · C2 · C7 · C14 · C15 · H13.
**Capabilities:** **crea CAP-118** `telemetria/receptor-otlp-embebido` (su tercer puntero,
`router.go#NewHandler`, ya existe hoy ⇒ la hoja se puede crear en este commit).

**Criterios de aceptación:**

```gherkin
Escenario: la cola llena no bloquea
  Dado el receptor con la cola saturada
  Cuando llega un POST /v1/logs
  Entonces responde en menos de 5 ms
  Y responde 200 con partialSuccess.rejectedLogRecords > 0
  Y el contador de descartes de salud subió

Escenario: protobuf se rechaza diciendo por qué
  Dado un POST con Content-Type application/x-protobuf
  Entonces responde 415
  Y el cuerpo nombra OTEL_EXPORTER_OTLP_PROTOCOL=http/json
  Y no se guardó nada
```

**⚠ Control positivo (§4):** los escenarios «no se guardó nada» (415, 413, 400) se assertan contra
un `httptest.Server` de **puerto efímero** y, en la misma corrida, un POST válido con **marcador
distinto** que **sí** tiene que llegar al sink. Nunca `len(recibidos) == 0` a secas.

**V&V:** `TestReceptorNoBloqueaAlEmisor` (fitness) · `TestPayloadGiganteSeRechazaSinLeerlo`
(fitness) · `TestProtobufSeRechazaConMotivo` · `BenchmarkReceptorLogs` (colocados,
`otlp/receptor_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/otlp/ -run TestReceptor -race -v
go test ./internal/adapters/telemetria/otlp/ -bench BenchmarkReceptorLogs -benchtime 200x
```

---

#### T17 · Token de ingesta acotado y montaje en el mux

**Depende de:** T16.
**Objetivo:** que el receptor viva en `:4200` bajo los tres gates, con **su propio token** — porque
filtrar el de la API concedería «conducir un agente con acceso al filesystem».

**Archivos (tocados):**

- `internal/adapters/transport/http/auth.go` — `isAPIPath` pasa a
  `strings.HasPrefix(p, "/api") || p == "/events" || strings.HasPrefix(p, "/v1/")`. Se agregan
  `isRutaIngesta(p string) bool` (las tres rutas de ingesta) y `validaIngesta(r, cfg) bool`.
  `AuthConfig` gana `TokenIngesta string` y `IngestaTokenObligatorio bool`.
- `internal/adapters/transport/http/router.go` — `mux.Handle("POST /v1/logs", otlp)` y
  `mux.Handle("POST /v1/metrics", otlp)`. El handler entra **inyectado como `http.Handler`**, igual
  que `events` hoy: `transport-http` **no importa** `telemetria/otlp` (go-arch-lint lo prohíbe y no
  hace falta).
- `cmd/arnesia/main.go` — mint del segundo token + `AuthConfigFor` recibe los dos.

**Diseño** — §3.1, §7.3 y **A22**.

| token | quién lo tiene | qué abre |
|---|---|---|
| `ARNESIA_AUTH_TOKEN` (el de hoy) | el shell Tauri, por env | **todo** `/api` + `/events` |
| **`token_ingesta`** (nuevo) | la ficha `0600` + el env del spawn | **solo** `POST /v1/logs`, `POST /v1/metrics`, `POST /api/telemetria/proceso` |

**A22 — y es la parte que hay que leer dos veces:** `/v1/logs` y `/v1/metrics` **aceptan sin token**
cuando la request pasa el Host gate loopback. No es un descuido: H10.2 verificó que Claude Code
**no expande `${VAR}` en el bloque `env`**, así que en `s2-instrumentado` el token no puede llegar
por indirección, y ponerlo literal metería un secreto vivo en un archivo versionable. Lo que se
pierde está acotado —un local del mismo usuario puede inyectar ruido contable— y **el ruido no
contamina ningún número**: entra `atribucion=sin-dato`, no suma a ningún total (A15) y es **visible**
en `Cobertura.SinDato` y en `salud`. Quedan tres barreras: Host gate, tope de 4 MiB y rate limit.

**Escotilla:** `--telemetria-ingesta-token-obligatorio` exige token también en `/v1/*`.
**Consecuencia honesta que la UI dice: `s2-instrumentado` deja de reportar.**
`POST /api/telemetria/proceso` exige token **siempre**, con o sin escotilla.

**RF que cierra:** RF-284 (parcial: el confinamiento).
**Escenarios:** B7 · B8 · B9 · B14 · B15.
**Capabilities:** **crea CAP-133** `telemetria/token-de-ingesta-acotado`; **modifica** la hoja de
`auth.go` que ya lo reclama (CAP-48) y la de `router.go` (CAP-49) con `change_log type: extend`.

**Criterios de aceptación:**

```gherkin
Escenario: el token de ingesta no abre la API
  Dado el token de ingesta
  Cuando se usa en POST /v1/logs
  Entonces responde 200
  Cuando se usa en GET /api/sessions
  Entonces responde 401

Escenario: /v1 sigue bajo Host gate
  Dado un POST /v1/logs con Host ajeno
  Entonces responde 403

Escenario: el modo estricto apaga s2-instrumentado y lo dice
  Dado --telemetria-ingesta-token-obligatorio
  Cuando llega un POST /v1/logs sin token
  Entonces responde 401
  Y salud reporta el modo estricto encendido
```

**V&V:** `TestTokenDeIngestaNoAbreLaAPI` · `TestOTLPBajoLosTresGates` ·
`TestOTLPAceptaSinTokenBajoLoopback` · `TestModoEstrictoApagaS2Instrumentado` (fitness, T26 + versión
colocada en `internal/adapters/transport/http/auth_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/transport/http/ -run 'TestToken|TestOTLP|TestModoEstricto' -v
```

---

#### T18 · Ficha de descubrimiento del daemon

**Depende de:** T5, T17.
**Objetivo:** que un hook que no conoce nuestras convenciones encuentre el daemon en las tres
plataformas — y que **nunca** encuentre una ficha que nombra un puerto muerto.

**Archivos:**

- `internal/adapters/telemetria/descubrimiento/ficha.go` (nuevo) — `Publicar(ctx, f domain.FichaDaemon) error` ·
  `Leer() (domain.FichaDaemon, error)` · `Retirar() error`; implementa `ports.DescubrimientoDaemon`.
- `cmd/arnesia/main.go` (tocado) — `runServe` crea un **`net.Listener` explícito** antes de
  `srv.Serve(ln)` y publica la ficha **después** de que acepta; la retira en el shutdown ordenado.

**Diseño** — §7.2 (A8). Ruta: **`os.UserConfigDir()/arnesia/daemon.json`** — no `~/.arnesia`,
porque es el único archivo que tiene que encontrar un proceso ajeno en Linux/Windows/macOS (D6.7).
Contenido literal del JSON de §7.2 (líneas 1327-1338). Escritura **atómica** (temp en el mismo dir
+ `os.Rename`), permisos **`0600`**, directorio **`0700`**.

- **Ficha huérfana:** no se detecta con el pid. El hook falla el POST en 250 ms y sale fail-open —
  más simple y más robusto que un liveness check.
- **Dos daemons:** último que arranca gana la ficha. Los spawns de S1 no se ven afectados (cada uno
  lleva su endpoint en el env del proceso). Se documenta y se acepta.
- **Directorio no escribible (B13):** el daemon **arranca igual**, loguea `warn` y
  `salud.descubrimiento` dice «no publicada». S1 y `s2-instrumentado` siguen funcionando: van por
  env/settings, no por ficha.

**RF que cierra:** ninguno directo (habilita RF-284 y el hook).
**Escenarios:** B3 · B4 · B10 · B12 · B13.
**Capabilities:** **crea CAP-132** `telemetria/ficha-de-descubrimiento-del-daemon`.

**Criterios de aceptación:**

```gherkin
Escenario: la ficha no existe antes de que el listener acepte
  Dado el daemon arrancando
  Cuando el listener todavía no acepta
  Entonces la ficha no existe
  Cuando el listener acepta
  Entonces la ficha existe y su endpoint responde /healthz  ← control positivo
  Cuando el daemon hace shutdown ordenado
  Entonces la ficha ya no existe

Escenario: permisos
  Entonces el archivo es 0600 y su directorio 0700
```

**V&V:** `TestFichaSoloTrasEscuchar` · `TestFichaPermisos0600` · `TestFichaNoEscribibleDegradaHonesto` ·
`TestFichaSeRelePorInvocacion` (fitness T26 + colocados en `descubrimiento/ficha_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/descubrimiento/ -v
XDG_CONFIG_HOME=$(mktemp -d) go test ./cmd/arnesia/ -run TestFicha -v
```

---

#### T19 · API HTTP `/api/telemetria/*` + contrato OpenAPI

**Depende de:** T13, T14, T15, T17.
**Objetivo:** las 8 rutas que el FE consume, con el contrato duro de que **`null` y `0` no son lo
mismo en ninguna**.

**Archivos:**

- `internal/adapters/transport/http/telemetria.go` (nuevo) — `getResumen` · `getCajas` ·
  `getDetalleCaja` · `getMejoras` · `getPortafolio` · `deleteTelemetriaArnes` · `getSalud` ·
  `postProceso`.
- `internal/adapters/transport/http/router.go` (tocado) — las 8 rutas.
- `docs/architecture/contracts/api/openapi.yaml` (tocado) — sube a **`0.8.0-telemetria`** con las 9
  rutas (las 8 + `/v1/*` documentadas como fuera de `/api`) y los schemas `ResumenTelemetria` ·
  `Cobertura` · `GastoCaja` · `DetalleCaja` · `PuntoDeMejora` · `RespuestaMejoras` ·
  `FilaPortafolio` · `SaludTelemetria` · `EventoProceso`.

**Diseño** — la tabla de §3.2 (líneas 607-618), literal. Tres contratos que un implementador
«prolijo» rompería:

- **`GET …/cajas` devuelve también las cajas SIN dato**, con `atribuible:false` + `motivo` y
  `costo_micros: null`. **Nunca omitidas** — omitir obliga al FE a inventar.
- **`Cobertura.NoLlegaron` viaja explícito**, no se calcula por resta en el FE.
- **`Estimado` es siempre `true`** mientras la fuente sea `cost_usd_micros` (H4: está documentado
  como *"Estimated cost"*, no facturación).
- `GET /api/telemetria/portafolio` devuelve `puesto: null` cuando el arnés no declara `rol` (D20).

**RF que cierra:** RF-283 (el endpoint del borrado) · da el dato de RF-235…RF-237, RF-259…RF-268.
**Escenarios:** los de A-I que el FE observa.
**Capabilities:** **crea CAP-137** `http-sse/superficie-telemetria`.

**Criterios de aceptación:**

```gherkin
Escenario: null no se convierte en 0 en el wire
  Dado un evento con tokens de razonamiento ausentes
  Cuando se pide GET /api/telemetria/arneses/{clave}/cajas/{cajaId}
  Entonces el JSON no trae la clave razonamiento
  Y no la trae con valor 0

Escenario: una caja sin dato viaja, no se omite
  Dado un arnés con 4 cajas y costo atribuido a 2
  Entonces GET .../cajas devuelve 4 elementos
  Y los 2 sin dato traen atribuible:false y motivo no vacío
```

**V&V:** `TestNoAplicaNoEsCeroEnElWire` (fitness) · `TestCifraLlevaConfianza` (fitness) ·
`TestCajasSinDatoNoSeOmiten` · `TestPortafolioPuestoNullSinRol` (colocados,
`internal/adapters/transport/http/telemetria_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/transport/http/ -run TestTelemetria -v
npx --yes @redocly/cli lint docs/architecture/contracts/api/openapi.yaml
```

---

#### T20 · `arnesia hook proceso` — el hook es el propio binario

**Depende de:** T10, T18, T19.
**Objetivo:** que el arnés pueda emitir señal de proceso desde S2 **sin shipear un runtime nuevo** y
**sin poder romper nunca el turno del usuario**.

**Archivos (nuevo):** `cmd/arnesia/hook.go` — `runHookProceso(args []string) error`; registrado en
el dispatch de `main.go` como subcomando `hook proceso`.

**Diseño** — §7.1 (A7). El contrato, que es la tabla de líneas 1286-1295 y no admite matices:

| aspecto | contrato |
|---|---|
| entrada | el payload del hook por **stdin**, tal cual |
| stdout | **vacía**. Un hook que imprime inyecta texto al contexto del agente |
| **exit code** | **siempre 0. Sin excepciones.** Un `UserPromptSubmit` que sale ≠0 bloquea el turno |
| duración | tope duro de **250 ms**; pasado eso abandona y sale 0 |
| destino | lee la ficha (T18) y `POST /api/telemetria/proceso` a loopback. **Jamás red externa** |
| sin daemon | **fail-open silencioso**: no reintenta, no escribe a disco, no avisa |
| qué manda | el `EventoTelemetria` **ya proyectado** (T10). **Nunca su stdin** |
| registro | `slog` a `~/.arnesia/logs/arnesia.log` al fallar, **solo si el archivo ya existe** (no lo crea: sería efecto colateral de un hook) |

- **La ficha se relee en CADA invocación**, no se cachea (B4).
- **«Fail-open silencioso» es silencioso hacia el usuario, no invisible**: el gap aparece después
  como `Cobertura.NoLlegaron` y como «sin dato». Nunca como 0.
- **H10.3 garantiza que el binario FALTE, no que FALLE.** Que Claude Code absorba un comando de hook
  inexistente (`exit 0`, `is_error:false`, stderr vacío) es del runtime; que **este** binario salga
  0 en todas sus ramas de error **sigue siendo obligación nuestra** y tiene su test.
- `arnesia hook proceso` queda **fuera** de la familia `arnesia telemetria` a propósito: no es una
  vía de inspección, y su contrato (stdout vacío) es incompatible con un subcomando que imprime.

**RF que cierra:** **RF-284** (el arnés no egresa y no guarda de más).
**Escenarios:** A4 · B1 · B12 · I5 · I6 · I7 · I12.
**Capabilities:** **crea CAP-131** `telemetria/hook-de-proceso-fail-open`.

**Criterios de aceptación:**

```gherkin
Escenario: sale 0 en todas las ramas
  Dado el hook invocado con stdin vacío, con JSON roto, sin daemon, y con daemon devolviendo 500
  Entonces en los cuatro casos el exit code es 0
  Y en los cuatro casos stdout está vacío

Escenario: fail-open sin daemon, con control positivo
  Dado el daemon apagado
  Cuando el hook corre con el marcador A
  Entonces sale 0 en menos de 250 ms y no llegó nada
  Cuando el daemon se levanta en un puerto efímero y el hook corre con el marcador B
  Entonces el receptor recibió exactamente 1 evento, y es el B   ← control positivo
```

**⚠ Control positivo (§4):** el segundo escenario es el arquetipo. Dos marcadores **distintos**, un
`httptest.Server` en `:0`, y el assert es `len(conA)==0 && len(conB)==1` — nunca «no llegó nada».

**V&V:** `TestHookNoTardaNiFalla` · `TestHookStdoutVacio` · `TestHookFailOpenSinDaemon` ·
`TestHookFichaHuerfana` · `TestHookNoReenviaContenido` (colocados en `cmd/arnesia/hook_test.go` +
fitness en T26).
**Demostración:**
```bash
go test ./cmd/arnesia/ -run TestHook -race -v
echo '{"session_id":"s","prompt_id":"p","hook_event_name":"Stop","last_assistant_message":"MARCA"}' \
  | go run ./cmd/arnesia hook proceso; echo "exit=$?"   # exit=0, stdout vacío
```

---

#### T21 · Spawn de S1 + `UsoDelTurno` del `result`

**Depende de:** T13, T17.
**Objetivo:** que toda sesión que ArnesIA spawnea nazca instrumentada, y que el split 5m/1h que el
`result` **ya trae y hoy se tira** llegue al almacén **sin un segundo parser de stream-json**.

**Archivos (tocados):**

- `internal/adapters/agent/claudecode/conductor.go` — (a) `Conductor` gana `EnvExtra []string` y
  `Spawn` aplica `cmd.Env = append(os.Environ(), c.EnvExtra...)`; **hoy `cmd.Env` es nil** (hereda el
  del daemon) ⇒ **es un cambio de código, no de configuración**. (b) nueva función exportada
  `SpawnEnv(opts ports.SpawnOpts, tokenIngesta, endpoint string) []string`. (c) el struct `usage`
  (línea 435) gana `cache_creation.{ephemeral_5m_input_tokens, ephemeral_1h_input_tokens}`,
  `service_tier`, `speed`, y `rawFrame.ModelUsage` gana `costUSD` · `canonicalModel` · `provider`.
  (d) **se extrae una función con nombre `parseResult(f rawFrame, last *usage) *domain.UsoDelTurno`**
  desde el `case "result"` de `translate()` — **hoy no existe** y **CAP-136 la apunta por nombre**,
  así que sin extraerla R1 falla a nivel símbolo.
- `internal/ports/agent.go` — `AgentEvent` gana `Uso *domain.UsoDelTurno`, poblado **solo** en
  `EventResult`.
- `internal/usecase/session_service.go` y `run_service.go` — al recibir `EventResult` con `Uso`,
  construyen el `domain.EventoTelemetria` con `Emisor=streamjson` y lo mandan al sink.

**Diseño** — §4.2 (A5) y §10.2. **No se crea `adapters/telemetria/streamjson/`**: sería un segundo
decodificador del mismo frame, driftando por separado, y violaría
`adaptadores-de-agente-intercambiables` y `conductor-no-parsea-jsonl`. **Un adaptador de agente por
runtime, no dos.**

Env exacto del spawn (§10.2 líneas 1634-1643), las 9 líneas, **literal**. Lo que no es opcional:

- **`OTEL_EXPORTER_OTLP_PROTOCOL=http/json`**, no `http/protobuf` — invierte F4 y es lo que hace
  barato al decodificador (+0,49 MB vs +10,79 MB, V5). **El default de Claude Code es gRPC :4317:
  sin esta línea no llega nada.**
- **`OTEL_EXPORTER_OTLP_ENDPOINT` sin `/v1/...`**: el exportador concatena la ruta.
- **`OTEL_RESOURCE_ATTRIBUTES`** viaja copiado a cada punto y a cada log record (V4).
  `arnesia.caja` solo en el spawn del conductor T3; en el Dock queda vacío y la atribución es a
  nivel sesión.
- El token va por `OTEL_EXPORTER_OTLP_HEADERS`, **nunca** por query string, y es el de **ingesta**,
  **nunca** `ARNESIA_AUTH_TOKEN`.
- ⚠️ **`OTEL_LOGS_EXPORTER` / `OTEL_LOGS_EXPORT_INTERVAL` no están verificados** contra
  `claude 2.1.220` (§13.2 ítem 2). Ver **§6 · parada P3**.

**RF que cierra:** RF-285 (el split que alimenta B1).
**Escenarios:** A1 · B5 · G12 (la señal).
**Capabilities:** **crea CAP-135** `conductor/spawn-inyecta-telemetria` y **CAP-136**
`conductor/uso-del-turno-en-el-result`.

**Criterios de aceptación:**

```gherkin
Escenario: el env del spawn está completo y es el correcto
  Cuando se llama SpawnEnv
  Entonces contiene las 9 variables de §10.2
  Y OTEL_EXPORTER_OTLP_PROTOCOL vale http/json
  Y OTEL_EXPORTER_OTLP_ENDPOINT apunta a 127.0.0.1
  Y NO contiene ARNESIA_AUTH_TOKEN ni su valor

Escenario: el split llega sin tocar el JSONL
  Dado el frame result de evidencia/result-envelope.json
  Cuando el adaptador lo traduce
  Entonces AgentEvent.Uso trae ephemeral_1h_input_tokens = 4099 y ephemeral_5m_input_tokens = 0
  Y el 0 es un dato, no una ausencia
```

**V&V:** `TestSpawnInyectaTelemetria` · `TestSpawnNoFiltraElTokenDeAPI` (fitness) ·
`TestResultTraeUsoDelTurno` · `TestSplitTTLLlegaDelStreamJSON` (colocados,
`internal/adapters/agent/claudecode/conductor_test.go`, contra `result-envelope.json`).
**Demostración:**
```bash
go test ./internal/adapters/agent/claudecode/ -run 'TestSpawn|TestResult|TestSplit' -v
```

---

#### T22 · Eventos de proceso del daemon: rotación, corrida, gate y `turno_esperado`

**Depende de:** T13, T21.
**Objetivo:** que B2 y P1 tengan la señal que **solo el daemon** puede dar, y que la conciliación de
cobertura tenga denominador.

**Archivos (tocados):**

- `internal/usecase/session_service.go` — al rotar contexto (RF-195) emite `EventoRotacion`; al
  cerrar un turno del stream-json escribe la fila de `turno_esperado` **por cada turno que ocurrió,
  se sepa medir o no** (A9).
- `internal/usecase/run_service.go` — emite `EventoCorrida` por iteración de caja (con
  `Resultado ∈ {ok, rechazado, reintento, cancelado}`) y `EventoGate` con el veredicto.

**Diseño** — §6.4 y la tabla de `TipoEvento` de §2.1. `turno_esperado` se marca `medido=1` cuando
llega el `api_request` de ese `(sesion_id, turno_id)`; `Cobertura.NoLlegaron` cuenta los `medido=0`
con **más de 5 minutos** de antigüedad. Eso convierte el agujero de WSL2/devcontainer de «un 0 que
parece un dato» en **«2 de 7 turnos no reportaron telemetría»**.

**En S2 no hay denominador independiente**: `esperados: null` y la UI dice «cobertura desconocida
fuera de ArnesIA».

**RF que cierra:** RF-262 (la mitad de proceso del join).
**Escenarios:** A7 · B6 · G3 · G7 · G9 · G10.
**Capabilities:** **no crea hoja nueva.** Los dos archivos ya están reclamados por capabilities del
módulo `usecases`; se les agrega el puntero y un `change_log` con `type: extend` y el resumen del
cambio (R2 satisfecho, R3 satisfecho).

**Criterios de aceptación:**

```gherkin
Escenario: los turnos que ocurrieron se cuentan aunque no se midan
  Dado 7 turnos cerrados por el stream-json y 5 api_request recibidos
  Entonces Cobertura.Esperados es 7
  Y Cobertura.NoLlegaron es 2
  Y el total NO se presenta como completo

Escenario: el rechazo de gate se une al costo del turno
  Dado una caja con 4 corridas, 3 rechazadas en el gate
  Entonces P1 devuelve un punto con el costo de las 3 en USD
```

**V&V:** `TestConciliacionCuentaLosNoLlegados` (ya en T13, ahora con el productor real) ·
`TestP1CajaQueSeRechaza` · `TestB2CostoDeLaRotacion` (T15, ahora contra eventos reales del daemon).
**Demostración:**
```bash
go test ./internal/usecase/ -run 'TestConciliacion|TestP1|TestB2' -race -v
```

---

#### T23 · Forward externo filtrado (apagado por default)

**Depende de:** T13.
**Objetivo:** el escape hatch del **operador**, que **nunca** exporta el email de quien corra el
arnés y que **ningún arnés puede tocar**.

**Archivos:**

- `internal/usecase/telemetria_service.go` (tocado) — `(*TelemetriaService).Forward`.
- `internal/adapters/telemetria/forward/forward.go` (nuevo) — implementa `ports.ForwardOTLP`.

**Diseño** — §6.2 y D13.

- **Apagado por default.** Se enciende con `--telemetria-forward <endpoint>` /
  `ARNESIA_TELEMETRIA_FORWARD` — **nunca desde la API ni desde un arnés** (D13, invariante 2).
- **Reenvía el `EventoTelemetria` ya proyectado**, que por construcción no tiene PII. **Jamás el
  payload OTLP crudo**: reenviar crudo exportaría el email y los ids de cuenta (V6.1).
- **Indicador visible** en la UI mientras esté encendido (`SaludTelemetria.forward`).
- Los dos niveles de egreso de D13 son distintos y hay que no confundirlos:
  **`telemetria-no-egresa` ata al ARNÉS, no al daemon.**

**RF que cierra:** **RF-284** (la otra mitad).
**Escenarios:** F5 · I1 · I2.
**Capabilities:** **crea CAP-134** `telemetria/forward-externo-filtrado`.

**Criterios de aceptación:**

```gherkin
Escenario: apagado no abre ni un socket
  Dado un daemon recién arrancado sin el flag
  Cuando se ingieren 50 eventos
  Entonces el dialer fake registró 0 conexiones salientes
  Cuando se enciende el flag y se ingiere 1 evento
  Entonces registró exactamente 1   ← control positivo

Escenario: lo que sale es el evento proyectado
  Dado el forward encendido y un payload OTLP con user.email
  Entonces el cuerpo enviado no contiene ese email
  Y sí contiene sesion_id   ← control positivo
```

**⚠ Control positivo (§4):** los dos escenarios son «no salió nada» / «no contiene X». Ninguno se
asserta solo con la ausencia.

**V&V:** `TestForwardApagadoPorDefault` · `TestForwardNoReenviaCrudo` ·
`TestForwardSoloPorFlagDelOperador` (fitness T26 + colocados en `forward/forward_test.go`).
**Demostración:**
```bash
go test ./internal/adapters/telemetria/forward/ -v
go test ./docs/architecture/fitness/ -run TestForward -v
```

---

#### T24 · CLI `arnesia telemetria`

**Depende de:** T13, T14, T15.
**Objetivo:** verificar el módulo entero **sin FE**, con el mismo usecase que sirve el HTTP.

**Archivos (nuevo):** `cmd/arnesia/telemetria.go` — `runTelemetria(args []string) error` con los
subcomandos `resumen` · `mejoras` · `salud` · `purgar` · `catalogo` (§10.3).

**Diseño** — A17, mismo patrón que `arnesia portafolio` (S0-D9): **reusa el MISMO usecase que el
HTTP, cero lógica propia.** Salida JSON.

**RF que cierra:** ninguno directo. Es el instrumento de verificación de casi todos.
**Escenarios:** I11 (auditar qué guardamos).
**Capabilities:** **crea CAP-138** `cli-daemon/telemetria-cli`.

**Criterios de aceptación:**

```gherkin
Escenario: el CLI y el HTTP dicen lo mismo
  Dado la misma base de telemetría
  Cuando se corre `arnesia telemetria resumen --arnes X`
  Y se pide GET /api/telemetria/resumen?arnes=X
  Entonces los dos JSON son equivalentes campo a campo
```

**V&V:** `TestTelemetriaCLIReusaElUsecase` (colocado, `cmd/arnesia/telemetria_test.go`) — compara la
salida del CLI contra la del handler sobre el mismo store.
**Demostración:**
```bash
go test ./cmd/arnesia/ -run TestTelemetriaCLI -v
HOME=$(mktemp -d) go run ./cmd/arnesia telemetria salud | jq .
```

---

#### T25 · Composition root, flags y presupuestos

**Depende de:** T16–T24.
**Objetivo:** que todo esto exista **en el daemon real** y que el binario no engorde más de lo
presupuestado.

**Archivos (tocado):** `cmd/arnesia/main.go`.

**Diseño** — §10.1 (líneas 1610-1628). El wiring, en `runServe`, después del índice y antes del
handler; y **la ficha se publica con el `net.Listener` explícito** (T18), que es la variante limpia
que §10.1 recomienda.

Flags nuevos: `--telemetria-retencion` (default 90 d, **PROPUESTO**, J-6) ·
`--telemetria-forward <endpoint>` (off) · `--telemetria-catalogo-refresco` (off, **A11**: es egreso
del daemon y D13 lo dejó como decisión deliberada del operador) ·
`--telemetria-ingesta-token-obligatorio` (off, A22).

**Presupuestos (§11), todos con su verificación:**

| presupuesto | valor | verificación |
|---|---|---|
| peso del binario | **+1,5 MB** máx. sobre el release anterior | `TestPresupuestoDeBinario` |
| peso absoluto | daemon ≤ **25 MB** sin `-s -w` (hoy 22,71) | ídem |
| latencia del handler OTLP | p99 ≤ **5 ms** | `BenchmarkReceptorLogs` + `TestReceptorNoBloqueaAlEmisor` |
| payload máximo | **4 MiB** | `TestPayloadGiganteSeRechazaSinLeerlo` |
| cola del receptor | **4096** eventos, exceso descartado y **contado** | `TestReceptorNoBloqueaAlEmisor` |
| retención default | **90 d** evento · **24 m** rollup | `TestPurgaRespetaTTL` |
| cardinalidad del rollup | ≤ **50 000** filas/mes | `TestRollupColapsaCardinalidad` |
| latencia del tablero | p95 ≤ **50 ms** sobre el rollup | benchmark |
| hook | ≤ **250 ms**, exit 0, stdout vacío | `TestHookNoTardaNiFalla` |
| catálogo embebido | ≤ **256 KB** | test de tamaño del `go:embed` |

**RF que cierra:** ninguno directo.
**Escenarios:** I4 (self-update con el daemon corriendo).
**Capabilities:** **modifica** la hoja de `cmd/arnesia/main.go` (`cli-daemon`) con `type: extend`.

**Criterios de aceptación:**

```gherkin
Escenario: el binario no engorda más de lo presupuestado
  Cuando se compila el daemon
  Entonces su tamaño no supera en más de 1,5 MB al baseline del release anterior
  Y no supera 25 MB en absoluto

Escenario: el daemon arranca aunque la telemetría falle
  Dado un ~/.arnesia de solo lectura
  Cuando arranca el daemon
  Entonces sirve /healthz y la API
  Y salud reporta el almacén no disponible con motivo
```

**V&V:** `TestPresupuestoDeBinario` (fitness; `t.Skip` bajo `-short`, corre en CI) ·
`TestDaemonArrancaSinAlmacen` (colocado, `cmd/arnesia/main_test.go`).
**Demostración:**
```bash
go build -o /tmp/arnesia-nuevo ./cmd/arnesia && ls -l /tmp/arnesia-nuevo
go test ./docs/architecture/fitness/ -run TestPresupuestoDeBinario -v
go test ./... -race
```

---

#### T26 · Fitness, `.go-arch-lint.yml` y boundaries

**Depende de:** T25.
**Objetivo:** que la doctrina del módulo la enforcee una máquina, no un comentario.

**Archivos:**

- `docs/architecture/fitness/telemetria_test.go` (nuevo) — **los 33 tests de §10.5** (líneas
  1728-1764), tal cual están descritos. Los marcados «(colocado)» ya viven en su paquete; acá van
  los de boundary.
- `docs/architecture/fitness/.go-arch-lint.yml` (tocado) — el bloque de §10.4 (líneas 1682-1716),
  **en este commit y no antes**: `go-arch-lint` falla con componentes cuyo `in:` no matchea ningún
  archivo. Los 5 componentes + sus `deps` + `cmd` extendido.
- `docs/architecture/boundaries/telemetria-de-nacimiento.md` — sus 10 checks pasan de `(pendiente)`
  a nombrar su enforcer; `status: proposed` → **`enforced`**; `enforced_by:` poblado.
- `docs/architecture/boundaries/{ingesta-por-allowlist-declarada,no-aplica-no-es-cero,cifra-viaja-con-su-confianza,peso-del-binario-es-presupuesto}.md`
  — ídem: sus checks nombran los tests de este paquete.
- `docs/architecture/INDEX.md` — cifras regeneradas.

**Diseño — lo que el bloque de arch-lint prohíbe por omisión, y es intencional:**
`transport-http` **no** puede importar `telemetria-*` (el receptor entra inyectado como
`http.Handler`, igual que el broker SSE) · los adaptadores de telemetría **no se importan entre sí**
· `index` y `telemetria-store` **no se conocen**.

**RF que cierra:** los §H, por enforcement.
**Escenarios:** todos los marcados `F` en la matriz.
**Capabilities:** ninguna (`docs/architecture/fitness/` está fuera de R2), pero **todas** las hojas
del módulo dependen de que sus `valida:` existan ⇒ **este ticket es el que hace que R4
(`cap-estado-consistente`) sea verdad** para las 22.

**Criterios de aceptación:**

```gherkin
Escenario: los 33 tests existen y pasan
  Cuando se corre el paquete de fitness
  Entonces los 33 tests de §10.5 aparecen en la salida
  Y ninguno hace t.Skip sin razón escrita

Escenario: el grafo de imports es el declarado
  Cuando corre go-arch-lint
  Entonces transport-http no importa ningún paquete de telemetria
  Y ningún adaptador de telemetria importa a otro
```

**V&V:** los 33 + `go-arch-lint check`.
**Demostración:**
```bash
go test ./docs/architecture/fitness/ -v 2>&1 | grep -c '^=== RUN'
go run github.com/fe3dback/go-arch-lint@latest check --project-path . \
  --arch-file docs/architecture/fitness/.go-arch-lint.yml
```

---

#### T27 · Los cinco checks de conformance del arnés

**Depende de:** T20, T26.
**Objetivo:** el **mecanismo de obligación**: un arnés que no porta telemetría honesta **no sella**.

**Archivos:**

- `docs/architecture/boundaries/telemetria-de-nacimiento.md` — las filas de los 5 checks del arnés
  (§7.4): `arnes-declara-telemetria` · `arnes-porta-hook-proceso` · `hook-es-fail-open` ·
  `telemetria-no-egresa` · **`hook-proyecta-campos`** (el hermano nuevo que pide H4).
- `internal/adapters/conformance/mechanism/adapters.go` (tocado) — la rama `static-scan` de estos
  checks sobre el paquete del arnés.

**Diseño** — §7.4. Lo que exige cada uno:

| check | qué exige |
|---|---|
| `arnes-declara-telemetria` | `arnes.l0.json` tiene bloque `telemetria: {version, eventos[], atribucion}` |
| `arnes-porta-hook-proceso` | el paquete shipea la config de hooks del kit, con los 4 eventos |
| `hook-es-fail-open` | el comando del hook tiene `timeout` declarado y no hay ruta que salga ≠0 |
| `telemetria-no-egresa` | el hook solo apunta a loopback o a disco local |
| **`hook-proyecta-campos`** | el hook **no reenvía su stdin**. Un hook que postea el payload entero a `127.0.0.1` **cumple `telemetria-no-egresa` y aun así filtra la conversación al almacén local** — por eso son dos checks, no uno |

**RF que cierra:** **RF-284** (los dos checks duros).
**Escenarios:** E4 · I8 · I9.
**Capabilities:** **modifica** la hoja de `conformance` con `type: extend`.

**Criterios de aceptación:**

```gherkin
Escenario: no alcanza con no egresar
  Dado un arnés cuyo hook postea su stdin entero a 127.0.0.1
  Cuando corre arnesia conformance --arnes <ese arnés>
  Entonces telemetria-no-egresa pasa
  Y hook-proyecta-campos FALLA
  Y el motivo nombra el campo fuera de la allowlist

Escenario: sin bloque telemetria no sella
  Dado un arnés sin bloque telemetria en su arnes.l0.json
  Entonces arnes-declara-telemetria falla con el check nombrado
```

**V&V:** `TestConformanceArnesSinBloqueTelemetria` · `TestConformanceArnesSinHook` ·
`TestHookNoReenviaContenido` (fitness) — colocados en
`internal/adapters/conformance/mechanism/adapters_test.go`.
**Demostración:**
```bash
go test ./internal/adapters/conformance/... -v
go run ./cmd/arnesia conformance --arnes dogfood/dev-full-cycle.graph.json | grep telemetria
go run ./cmd/arnesia conformance --todo | tail -3   # sin regresión vs. el checkpoint
```

---

### Tramo B — La superficie

> 🚦 **PUERTA: este tramo no arranca sin el 🧑‍⚖️ del mockup** (`mockup-capa-mejora.html` va por
> iteración 1). `spec.md` §Estado: *«los RF 🎨 no se construyen hasta que el mockup tenga su
> 🧑‍⚖️»*. El Tramo A no lo necesita; este sí.
>
> **Reglas del tramo, para los diez tickets:**
> 1. **La story ES el test** (`fe-visual-fitness.md`, job `visual-fitness` de CI). Cada estado de
>    `design.md` §5 es una story con `play()`.
> 2. **`npx vitest run --project=storybook <archivo>` corre headless** (~4 s/archivo). Verificado
>    2026-07-26; la nota vieja quedó derogada en T2.
> 3. **Ninguna story nace con `a11y: { test: "todo" }`.** Si falla axe, se arregla el color o el
>    marcado. El único `todo` heredado (`arnes-node`, `map-canvas`) **no se amplía**.
> 4. **Superset estricto (BR-M16):** nada firmado se quita ni se mueve. Las props nuevas son
>    opcionales y sin ellas el DOM es **idéntico** al de hoy.
> 5. **Fixtures:** `verificacion-2026-07-26/evidencia/` para la **forma**; los números del copy son
>    **constantes de diseño** y van en el bloque `ILUSTRATIVO` con el comentario literal de
>    `plan-storybook.md` §3.3.
> 6. **`entities/arnes` no importa `entities/telemetria`** (D18). Jamás.

---

#### T28 · Promoción de `Skeleton` y `ErrorBody` a `shared/ui`

**Depende de:** 🧑‍⚖️ del mockup.
**Objetivo:** que el Mapa pueda tener los mismos estados de carga y error que el Portafolio ya tiene
firmados (**H-6**), sin copiarlos.

**Archivos:**

- `web/src/shared/ui/estado-carga.tsx` (nuevo) — `Skeleton({label})` y
  `ErrorBody({motivo, onReintentar})`, movidos tal cual de `portafolio-list.tsx:280-307`.
- `web/src/shared/ui/estado-carga.stories.tsx` (nuevo) — 2 stories (§2.17).
- `web/src/widgets/portafolio/ui/portafolio-list.tsx` (tocado) — importa de `shared/ui`; **cero
  cambio de conducta**.
- `web/src/shared/index.ts` (tocado) — barrel.

**Diseño:** refactor de movimiento. Legal en `shared/ui` porque son **puramente presentacionales**
(props `label`, `motivo`, `onReintentar`): cero dominio ⇒ no violan `ui-not-domain` (depcruise,
`error`). Misma jugada que `FiltroDisclosure` en el paquete de marketplace.

**RF:** ninguno directo (H-6). **Escenarios:** los de transporte del Mapa.
**Capabilities:** **modifica** `fe-shell`/`fe-portafolio` con `type: extend` (movimiento de símbolo).

**Aceptación:** las **12 stories firmadas** de `portafolio-list.stories.tsx` pasan sin tocarlas.
**V&V:** `SkeletonHonesto` · `ErrorConMotivoYReintentar` (§2.17) + la suite vigente.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/shared/ui/estado-carga.stories.tsx \
  src/widgets/portafolio/ui/portafolio-list.stories.tsx
```

---

#### T29 · `entities/telemetria`: tipos, selectores, fixtures y los cuatro componentes

**Depende de:** T19 (el wire), T28.
**Objetivo:** el vocabulario visual del dominio, en un solo lugar: **un** formateador de dinero,
**una** marca de confianza, **una** barra de cobertura, **un** sparkline.

**Archivos (nuevos, salvo donde se diga):**

```
web/src/entities/telemetria/index.ts
web/src/entities/telemetria/model/types.ts         Confianza · Ventana · CifraCaja · Cobertura ·
                                                    PuntoMejora · Detector · EstadoDetector ·
                                                    BucketToken · ParidadCosto · FilaPortafolio
web/src/entities/telemetria/model/selectors.ts     usd() · etiquetaConfianza() · tonoConfianza() ·
                                                    etiquetaDetector() · direccionTendencia() +
                                                    ETIQUETA_CONFIANZA · TITULO_CONFIANZA (D18)
web/src/entities/telemetria/model/selectors.test.ts
web/src/entities/telemetria/ui/cifra-usd.tsx        <CifraUSD>       (4 stories, §2.1)
web/src/entities/telemetria/ui/marca-confianza.tsx  <MarcaConfianza> (5 stories, §2.2)
web/src/entities/telemetria/ui/barra-cobertura.tsx  <BarraCobertura> (5 stories, §2.3)
web/src/entities/telemetria/ui/sparkline.tsx        <Sparkline>      (4 stories, §2.4)
web/src/entities/telemetria/testing/telemetria.ts   bloques MEDIDO / ILUSTRATIVO
```

Y `web/src/shared/api/types.ts` + `client.ts` (tocados): los tipos y las 8 llamadas del wire.

**Diseño** — `design.md` §1.3 + `plan-storybook.md` §2.1-2.4. Lo que gobierna:

- **`CifraUSD` es el ÚNICO lugar donde se formatea dinero** (RF-281): `USD` antepuesto, coma
  decimal, dos decimales, `tabular-nums`. **`0,004` no se redondea a `0,00`.** Sin monto ⇒
  `sin dato`, **nunca** `0,00` ni `—` a secas.
- **`MarcaConfianza`: `exacta` no renderiza nada.** *La ausencia de marca **es** la señal.* Los
  otros tres tienen texto propio y `title` propio; **ninguno se pinta como «atribución aproximada»**,
  que taparía tres cosas distintas bajo una.
- **`BarraCobertura`: 4 segmentos**, separador de 1 px de `--card` entre ellos (`--heat-4` vs
  `--heat-3` son 1,37:1 — indistinguibles sin él), rótulo `cobertura` **visible siempre** (H-11), y
  el assert real es el `aria-label` completo, no el color.
- **Una categoría en cero no se dibuja ni se nombra.**
- **`entities` presenta; no calcula.** El join, los detectores y el contrafactual llegan resueltos
  del dominio Go.
- **Fixtures:** el bloque `MEDIDO` sale byte por byte de `evidencia/` (plan-storybook §3.2, incluidos
  **el cero legítimo** `ephemeral_5m = 0` y **la ausencia real** de `razonamiento`); el bloque
  `ILUSTRATIVO` lleva **el comentario literal** de §3.3.

**RF que cierra:** RF-281 · RF-242 · RF-237 · RF-266 · RF-279 · RF-280 (parcial) · RF-274.
**Escenarios:** E6 · E8.
**Capabilities:** CAP-139 se crea en **T33** (necesita los 4 punteros de D22).

**Aceptación:** las 18 stories de §2.1-2.4 pasan con axe en `error`, incluida `LosCuatroJuntos`
(`queryByText(/aproximad/i)` es `null`).
**V&V:** §2.1 (4) · §2.2 (5) · §2.3 (5) · §2.4 (4) + `selectors.test.ts` (project `unit`).
**Demostración:**
```bash
cd web
npx vitest run --project=storybook src/entities/telemetria/ui/
npx vitest run --project=unit src/entities/telemetria/model/selectors.test.ts
pnpm run fsd    # steiger: entities/telemetria no importa widgets ni otra entity
```

---

#### T30 · `layers.ts` + `map-bar.tsx`: el slot «Mejora» y los motivos honestos

**Depende de:** T29.
**Objetivo:** que la capa exista en el conmutador y que los dos slots apagados digan **la verdad de
hoy**, no la de hace tres meses.

**Archivos (tocados):** `web/src/widgets/map-canvas/model/layers.ts` ·
`web/src/widgets/map-canvas/ui/map-bar.tsx` · `web/src/widgets/map-canvas/ui/map-bar.stories.tsx`.

**Diseño:**

- `Capa` pasa de `"estructura" | "tokens" | "perf" | "proceso"` a
  `"estructura" | "mejora" | "perf" | "proceso"`. **Es un rename de literal de unión: lo agarra
  `tsc`** y no queda ningún consumidor huérfano.
- `{ id: "tokens", label: "Tokens", disabled: true }` → `{ id: "mejora", label: "Mejora" }`.
  **Los cuatro slots se conservan, en el mismo orden.** 🔴 Es **desviación declarada de un baseline
  firmado** (`mockups/INDEX.md` regla 3 · D17.1) y va al gate como tal.
- `LayerDef` gana `motivo?: string`. El `title` fijo `"Necesita telemetría (indexer JSONL)"`
  (`map-bar.tsx:123`) **sale** — **eso ya no es verdad**: la señal llega (V1), lo que falta es el
  diseño. Los dos motivos nuevos son el copy literal de `design.md` §7.1.
- **El motivo no puede vivir solo en `title`** (RF-276): `aria-describedby` a un
  `<span class="sr-only">` con el **mismo** texto, y la story asserta la igualdad para que no
  driften.

**RF que cierra:** **RF-232** · **RF-233** · RF-276.
**Escenarios:** J2 (el slot Desempeño con tooltip honesto).
**Capabilities:** puntero `layers.ts#LAYERS` de CAP-139 (hoja creada en T33).

**Aceptación:**

```gherkin
Escenario: el conmutador ofrece Mejora y ya no ofrece Tokens
  Entonces hay 4 tabs con nombres Estructura, Mejora, Desempeño, Proceso en ese orden
  Y no existe ninguna tab llamada Tokens
  Y Mejora no está deshabilitada
  Y Desempeño y Proceso siguen deshabilitadas (la story vigente `Default` no cambia)
```

**V&V:** `CapaMejoraDisponible` · `CapaMejoraActiva` · `MotivosHonestos` · `MotivoAccesible` (§2.6).
**Demostración:**
```bash
cd web && pnpm exec tsc --noEmit && npx vitest run --project=storybook src/widgets/map-canvas/ui/map-bar.stories.tsx
```

---

#### T31 · `arnes-node.tsx` + `lane.tsx`: las marcas dentro del nodo

**Depende de:** T30.
**Objetivo:** que **solo las cajas** lleven cifra, que todo lo demás diga **por qué no**, y que la
capa apagada no deje rastro.

**Archivos (tocados):** `web/src/entities/arnes/ui/arnes-node.tsx` ·
`web/src/widgets/map-canvas/ui/lane.tsx` · `web/src/app/styles/map.css` ·
`arnes-node.stories.tsx` (+13) · `lane.stories.tsx` (+2).

**Diseño** — `design.md` §2.3 + **D18** + **D21**.

- **Props primitivas (D18):** `cifraUSD?: string` · `participacionPct?: number` ·
  `confianza?: "exacta"|"por-hash"|"por-proceso"|"sin-dato"` · `marcaFuga?: string` ·
  `marcaFugaGrave?: boolean` · `motivoSinDato?: string` · `tituloConfianza?: string`.
  **Cero import de `entities/telemetria`.**
- **Orden de hijos**: los 5 de hoy sin mover, y **debajo** `.mej-cifra` · `.mej-fuga` · `.mej-share`
  (§2.3 líneas 186-196). **Ninguna marca nueva es `position: absolute`**: la esquina superior
  derecha ya está ocupada dos veces (`.caja-badge` y `.prop-badge` comparten `top:8px; right:8px`,
  `map.css:310`/`:329`) y un tercer badge se superpondría (J-9).
- **El nodo sigue siendo un `<button>`** (`arnes-node.tsx:52`) ⇒ `.mej-*` son **`<span>`**, jamás
  controles. **Anidar un botón en un botón es DOM inválido.** Abrir una tarjeta no se hace desde el
  nodo: seleccionar la caja lleva el foco a su tarjeta en la lista de abajo.
- **D21 · marca de fuga grave:** `color: var(--crit)` sobre **`background: var(--card)`** +
  `border: 1px solid color-mix(in srgb, var(--crit) 45%, transparent)`. **No** `--crit-soft` de
  fondo (4,04:1 en claro, falla). Como este archivo hereda `a11y: { test: "todo" }` y ahí axe no
  corre, la story lleva un assert **computado** del `backgroundColor`.
- **Nodos sin dato atribuible:** `opacity: .62` + borde izquierdo punteado + línea de motivo. **La
  opacidad sola no alcanza (RF-280): el texto es el portador.** **Cinco** motivos —no seis— porque
  **D19 borró el de `conocimiento`**.
- **`lane-hd` pasa a tres hijos**: `<h3>` · **`<span class="count">` que se conserva** (J-8) ·
  `<span class="lane-usd">`. Carril sin cajas con dato ⇒ `sin dato`, **no `USD 0,00`**.

**RF que cierra:** **RF-238** · **RF-239** · **RF-240** · **RF-241** · **RF-243** · **RF-244** ·
RF-242 (en el nodo) · RF-245 (el guardián `EstructuraIntacta`).
**Escenarios:** G15 · E6 · E8.
**Capabilities:** puntero de CAP-139 (hoja en T33); **modifica** la hoja de `entities/arnes` con
`type: extend`.

**Aceptación:**

```gherkin
Escenario: la capa apagada no deja rastro
  Dado ArnesNode sin ninguna prop de mejora
  Entonces no existen .mej-cifra, .mej-fuga ni .mej-share
  Y no aparece ningún texto "sin dato atribuible"

Escenario: un nodo que no es caja nunca muestra un cero
  Dado un nodo con isCaja(box) === false
  Entonces no hay cifra, no hay barra, y no aparecen "0", "USD 0,00" ni "—" a secas
```

**V&V:** las 13 stories de §2.5 (incluidas `EstructuraIntacta`, `MejoraDesbordamiento`,
`SinDatoResto`) + las 2 de §2.7.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/entities/arnes/ui/arnes-node.stories.tsx \
  src/widgets/map-canvas/ui/lane.stories.tsx
pnpm exec stylelint "src/app/styles/map.css"
```

---

#### T32 · `FranjaMejora` — la franja de contexto y los estados 1, 1b, 2, 3, 3b, 4, 5

**Depende de:** T29, T30.
**Objetivo:** que ningún número aparezca sin su ventana, su denominador, su disclaimer y su
cobertura — y que los siete estados honestos tengan dónde vivir.

**Archivos (nuevos):** `web/src/widgets/map-canvas/ui/franja-mejora.tsx` ·
`franja-mejora.stories.tsx` (20) · `web/src/app/styles/mejora.css` (scope `.arnesia-mejora`,
importado desde `index.css`).

**Diseño** — `design.md` §2.2 (el orden de izquierda a derecha **es el orden de lectura del dato**,
no estética), §5.2 (los 7 estados de transporte y de dato) y §7.2/§7.7 (**copy literal**).

- **Orden:** ventana → total+denominador → **disclaimer pegado al total** (si se va al pie, se lee
  después de haber creído el número) → chip de reenvío *(condicional)* → cobertura
  (`margin-inline-start: auto`) → «qué guardamos».
- **Espaciado por token**, no los `10px 14px` hand-typed del mockup: `var(--space-3) var(--space-4)`
  / `gap: var(--space-3) var(--space-5)`. **Gana el token** — la disciplina vale más que 2 px de
  fidelidad a un `.html`.
- **D21 · disclaimer:** texto **`--foreground`** sobre `--warn-soft`, borde
  `color-mix(in srgb, var(--warn) 35%, transparent)`. **Nunca texto `--warn` sobre `--warn-soft`**
  (3,24:1 en claro).
- **`ReposoDark`** (`globals: { theme: "dark" }`) es obligatoria: sin ella el gate mira solo el tema
  claro.
- **Estado 1 vs 1b son cosas distintas**: «nunca corrió» ≠ «no corrió en estos 7 días» (H-9).
- **Estado 3 vs 3b son dos niveles de dato, no el mismo con otro nombre** (J-10/H9): los rótulos son
  strings **distintos** y la story asserta la desigualdad.
- **`Cargando` deja el `select` de ventana usable**; `Error` **no hace desaparecer la franja**
  (desaparecer se leería como «no hay capa»); `DaemonCaido` **conserva las cifras previas marcadas
  como posiblemente viejas** y es distinguible de «no hay datos».

**RF que cierra:** **RF-234** · **RF-235** · **RF-236** · **RF-237** · **RF-269** · **RF-270** ·
**RF-271** · **RF-272** · **RF-273** · RF-275 (el resumen) · H-6 · H-7 · H-9 · H-11.
**Escenarios:** A5 · A6 · A8 · F3 · F12 · B6.
**Capabilities:** puntero de CAP-139 (hoja en T33).

**Aceptación:**

```gherkin
Escenario: nunca un tablero en cero
  Dado un arnés sin ninguna corrida instrumentada
  Entonces el texto completo de la franja NO contiene "0,00"
  Y dice "Este arnés nunca corrió con telemetría."
  Y dice qué hacer

Escenario: el disclaimer se lee sin hover
  Entonces "estimado por el runtime, no es facturación" es visible
  Y no vive en un title, ni dentro de un [hidden], ni en un details cerrado
```

**V&V:** las 20 stories de §2.8 (19 + `ReposoDark`), con `Estado1SinDatos` asertando
`expect(canvasElement.textContent).not.toContain("0,00")`.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/map-canvas/ui/franja-mejora.stories.tsx
pnpm exec stylelint "src/app/styles/mejora.css"    # cero color literal (strict-value)
```

---

#### T33 · `PuntoMejoraCard` + `PuntosMejoraList` — el corazón del entregable

**Depende de:** T29, T31.
**Objetivo:** una tarjeta = una caja × un detector, con las cinco cosas que A4 exige, y una puerta
de entrada (**H-1**, que el mockup no dibuja).

**Archivos (nuevos):** `punto-mejora-card.tsx` (+ 15 stories) · `puntos-mejora-list.tsx`
(+ 7 stories), los cuatro en `web/src/widgets/map-canvas/ui/`.

**Diseño** — `design.md` §2.4, §2.5, §5.4, §7.4.

- **Anatomía en orden estricto** (§2.5 líneas 234-247): el orden **es el argumento** — qué pasa →
  cuánto → cuánto se ahorra → por qué lo creemos → contra qué → qué hacer.
- **La lista vive debajo del canvas**, ordenada por **ahorro contrafactual descendente**. No es un
  panel flotante ni un modal: es contenido de la página. **No puede ser un botón dentro del nodo**:
  el nodo YA es un `<button>` (H-1/J-9).
- **Relación bidireccional con el canvas:** seleccionar una caja resalta su tarjeta y la trae a la
  vista; enfocar una tarjeta resalta su caja. **La señal de «resaltada» lleva texto**
  (`↔ caja seleccionada`), no solo un borde de 6 px.
- **`Sesgo` es una fila que NUNCA se omite**: si no se identificó ninguno, **lo dice**.
- **`Umbral` no se pinta cuando el detector no se decide por un umbral** (P1) — en su lugar va
  `Patrón`.
- **«ver el cálculo» es un despliegue en línea, no un modal** (H-4): `aria-expanded` + `aria-controls`.
- **Chip S1-only** (J-2): sin él, en una instalación mayormente S2 la tarjeta insignia desaparece
  sin explicación.
- **Severidad como texto** (`atención` / `crítico`) — el `⚠` es `aria-hidden` y el borde de color es
  refuerzo (RF-257/RF-280).
- **`[Proponerlo en el chat]` no escribe archivos** (D17.3/BR-M12): **no existe ninguna prop de
  escritura**, y la story lo asserta por ausencia de `onAplicar`/`onEscribir` en `meta.args`.
- **`CompletaB1AtencionDark`** obligatoria (D21).
- **Una tarjeta sin contrafactual no existe** (A4). No se pinta degradada: la lista la filtra
  (`DescartaSinContrafactual`) y el inspector la lista como `sin fix propuesto` (T34).

**RF que cierra:** **RF-246…RF-257**.
**Escenarios:** G12 · G13.
**Capabilities:** **crea CAP-139** `fe-mapa/capa-mejora`, con los cuatro punteros de **D22**
(`layers.ts#LAYERS` · `cifra-usd.tsx#CifraUSD` · `marca-confianza.tsx#MarcaConfianza` ·
`punto-mejora-card.tsx#PuntoMejoraCard`) y los 10 story-tests de `capabilities-a-crear.md` como
`valida:`.

**Aceptación:**

```gherkin
Escenario: el titular no habla en jerga
  Entonces no contiene "B1" ni "ephemeral" ni "cache_creation"

Escenario: proponer no escribe
  Cuando se usa "Proponerlo en el chat"
  Entonces se llama onProponer una vez con { puntoId, textoPropuesto }
  Y no existe ninguna prop de escritura en el componente
```

**V&V:** las 15 de §2.9 + las 7 de §2.10.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/map-canvas/ui/punto-mejora-card.stories.tsx \
  src/widgets/map-canvas/ui/puntos-mejora-list.stories.tsx
cd .. && go test ./docs/architecture/fitness/ -run 'TestCapability(PointersResolve|PointerSymbolsResolve|Coverage)'
```

---

#### T34 · `InspectorMejora` + la cuarta tab

**Depende de:** T29, T33.
**Objetivo:** que el número se pueda **auditar** — bucket por bucket, contra el catálogo, contra el
proceso, y contra los detectores que no corrieron.

**Archivos:** `inspector-mejora.tsx` (nuevo, 17 stories) · `inspector.tsx` (tocado, +3 stories) ·
`web/src/app/styles/inspector.css` (extensión `.arnesia-inspector .mej-*`).

**Diseño** — `design.md` §2.6, §3.4, §5.5, §7.5.

- **Las tres tabs vigentes no se tocan**; la cuarta reusa `<Section>` (`inspector.tsx:71`).
- **Ancho 340 px** (`inspector.css:13`), **no los 380 del mockup**. Las cuatro etiquetas entran; si
  aparece scroll horizontal en la tira de tabs, **se corrige la etiqueta, no el contenedor**.
- **Orden de secciones:** de lo auditable a lo accionable — buckets → paridad → join → detectores.
- **La fila «no aplica» ocupa las dos columnas numéricas con `colspan=2`**, en itálica — **no** un
  guion en cada celda, que se leería como cero. Y **`0` es un dato válido** cuando el runtime tiene
  el concepto: el fixture usa el **cero real medido** (`ephemeral_5m = 0`).
- **La paridad muestra los dos números y NO elige** cuál es el bueno; el veredicto es **texto**
  (`✓ coinciden` / `⚠ difieren en USD 0,18`), no solo el tono.
- **Detector `off` = `--border`, nunca `--ok`**: «no sé» ≠ «sano» (regla ya enforced en
  `.pf-dot-salud.sin-senal`).
- **Los 7 fuera del MVP dicen `no medido todavía`**, jamás `sin hallazgos`.
- **La ventana es la de la capa** (J-3): el inspector la **hereda**; las «14 corridas» son el
  **denominador**, no la ventana.
- 🔴 **La story firmada `Tabs` (`inspector.stories.tsx:92`) asserta `toHaveLength(3)` y pasa a `4`
  en este mismo commit, o CI se pone roja.** Es la **única** modificación permitida a una story
  firmada en este paquete y va declarada en `PARIDAD.md`.

**RF que cierra:** **RF-258…RF-264** · RF-277.
**Escenarios:** F1 · F2 · F6 · F7 · F8 · G8 · G13 · G14.
**Capabilities:** amplía CAP-139 (`change_log type: extend`).

**Aceptación:**

```gherkin
Escenario: la suma cierra
  Dado una caja con USD 1,92
  Entonces la suma de los USD de las 6 filas de bucket es 1,92  (assert aritmético, no de texto)

Escenario: el panel inactivo está hidden, no oculto por CSS
  Entonces hay exactamente 3 [role=tabpanel][hidden]
```

**V&V:** las 17 de §2.11 + las 3 de §2.12.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/map-canvas/ui/inspector-mejora.stories.tsx \
  src/widgets/map-canvas/ui/inspector.stories.tsx
```

---

#### T35 · `PoliticaDatosDialog` — qué guardamos y el borrado

**Depende de:** T19 (el `DELETE`), T32.
**Objetivo:** que la promesa *«nada de tu cuenta, nada de la conversación»* sea **contrastable**, no
una frase — y que el borrado borre.

**Archivos (nuevos):** `politica-datos-dialog.tsx` + 8 stories, en `widgets/map-canvas/ui/`.

**Diseño** — `design.md` §5.7 y §7.8. **Vive detrás del enlace «qué guardamos» de la franja**
(**H-5**: el mockup lo dibuja como texto dentro de un panel didáctico, que no es una ubicación).

- Tres bloques: qué NO se guarda · **qué SÍ se guarda, con la lista de campos por nombre** ·
  retención + acción.
- **La lista de campos viene por prop (`camposPersistidos`), no hardcodeada**: RF-275 exige que sea
  *«la misma allowlist que aplica la ingesta, no una redacción aparte»*. Si el componente la
  inventa, la promesa deja de ser verificable.
- **El `{N}` de retención sale de la config.** ⚠ El `90` es **PROPUESTO, no firmado** (J-6): la
  story `RetencionDesdeConfig` asserta con `45` y que `queryByText(/90 días/)` es `null`.
- **El botón destructivo nace `disabled`** hasta que el operador confirma que entendió el alcance;
  mientras borra, **el diálogo no se puede cerrar** (mismo patrón que el wizard con POST en vuelo,
  S1-D19).
- Focus trap con `trapTabKeyDown`, `role="dialog"` + `aria-modal="true"`, Esc cierra en reposo.
- **H-14:** el copy suma la segunda línea que dice qué pasa con lo que **sí** llega (el contenido
  por el canal de hooks) y que **se descarta antes de escribirse**.

**RF que cierra:** **RF-275** · RF-278 · RF-283 (su superficie).
**Escenarios:** I10 · I11 · H5.
**Capabilities:** amplía CAP-139 (`type: extend`).

**Aceptación:**

```gherkin
Escenario: la lista de campos no la inventa la UI
  Dado camposPersistidos = [session_id, prompt_id, hook_event_name, tool_name, duration_ms, cwd]
  Entonces el DOM refleja exactamente ese array, ni uno más ni uno menos
```

**V&V:** las 8 de §2.14.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/map-canvas/ui/politica-datos-dialog.stories.tsx
```

---

#### T36 · Portafolio: las tres columnas nuevas

**Depende de:** T29, T19.
**Objetivo:** poder decir **«en este puesto»** — el eje diferencial del producto (D9.8) — sin
fabricar un puesto que el dato no tiene.

**Archivos:** `tabla-mejora-portafolio.tsx` (nuevo, 9+1 stories) · `portafolio-list.tsx` (tocado,
+2 stories) · `web/src/app/styles/portafolio.css` (extensión `.pf-mej-*`) ·
`web/src/pages/shell/ui/portafolio-view.tsx` (tocado: transporte lazy, mismo patrón que
`GET /api/marketplaces`).

**Diseño** — `design.md` §2.7, §3.5, §5.6, §7.6 + **D20**.

- Las 3 celdas se insertan **entre chips y dot de salud**: el dot sigue cerrando la fila, que es el
  ancla visual que la PARIDAD del Slice 1 firmó.
- **Son opcionales** ⇒ las **12 stories firmadas** del Slice 1 no cambian (`SinMejoraDOMIntacto` es
  el guardián).
- **D20 · el puesto:** la fila se agrupa por `(identidad, instalación)` y la etiqueta sale del
  campo `puesto` que el backend resolvió del `rol` del arnés. **`null` ⇒ `puesto sin declarar`**, que
  es el **caso normal de hoy**. `SinPuestoDeclarado` lo cementa: **nadie fabrica un puesto**.
- **D21 · el chip de punto de mejora** (`⚠ re-warm de cache · USD 0,53/corrida`) usa la marca de
  fuga grave corregida: texto `--crit` sobre **`--card`**, no sobre `--crit-soft`. Este archivo **sí**
  tiene el gate axe en `error` ⇒ **es acá donde el fix es load-bearing**. Se agrega **`ConDatoDark`**
  (story nueva respecto de plan-storybook, total 124 → **125**).
- **Al ordenar por costo, los «sin dato» se agrupan al final**, con separador rotulado. **Nunca
  intercalados como si valieran 0.**
- **Pie de tabla con el disclaimer** (H-12): un USD pelado en una tabla se lee como facturación.
- El scroll vive en `.pf-mej-scroll`, con `tabindex="0"` + `role="region"` + `aria-label` (requisito
  de axe para contenedor scrolleable). **El `<body>` jamás scrollea horizontal.**

**RF que cierra:** **RF-265** · **RF-266** · **RF-267** · **RF-268** · H-12.
**Escenarios:** A5 · A8 · E1 · E2.
**Capabilities:** **extiende** `fe-portafolio/lista-del-portafolio.yaml` con `change_log type: extend`
— **no es hoja nueva**, es una columna nueva de un capability que ya existe
(`capabilities-a-crear.md` §2, nota final).

**Aceptación:**

```gherkin
Escenario: la fila sin dato no desaparece ni ordena como cero
  Dado 3 arneses con dato y 2 sin
  Cuando se ordena por costo
  Entonces las 5 filas siguen en la tabla
  Y hay un separador rotulado "Sin datos de telemetría"
  Y el índice de la primera fila sin dato es mayor que el de la última con dato

Escenario: no se fabrica un puesto
  Dado una instalación cuyo arnés no declara rol
  Entonces la celda dice "puesto sin declarar"
  Y no aparece ningún nombre de puesto inventado
```

**V&V:** las 10 de §2.15 (9 + `ConDatoDark`) + las 2 de §2.16.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/portafolio/ui/
pnpm exec stylelint "src/app/styles/portafolio.css"
```

---

#### T37 · Transporte y composición: `workspace-stage` + `map-canvas` + el puente H-3

**Depende de:** T30–T36.
**Objetivo:** que las piezas se hablen, que la geografía **no se mueva**, y que la frase del
producto cruce las dos superficies.

**Archivos (tocados):** `web/src/pages/shell/ui/workspace-stage.tsx` ·
`web/src/widgets/map-canvas/ui/map-canvas.tsx` · `map-canvas.stories.tsx` (+4).

**Diseño** — `design.md` §1.2, §6.2 + H-3.

- **`workspace-stage.tsx` es el dueño del estado `capa` (línea 44) y el ÚNICO que hace transporte**
  (`fe-transporte-independiente`). Gana el estado de ventana y la carga de la telemetría; los
  widgets reciben **props puras**.
- **`map-canvas.tsx` gana `capa` y `mejora` (mapa `nodeId → CifraCaja`) y compone
  `CifraCaja → props primitivas` del nodo (D18). Cero rama de layout.**
- **§6.2 — corrección al mockup:** el mockup usa `.procrow { overflow-x: auto }`; **en el producto
  esa caja no existe**. Los carriles viven dentro del `stage` con el `transform` de `useViewport`:
  el Mapa se recorre con **pan y zoom**, no con scrollbar. ⇒ **no se agrega ningún contenedor con
  `overflow` dentro del canvas** — uno solo rompería `useEdgePaths`, que mide posiciones absolutas.
  Con la capa activa los nodos son ~26 px más altos y `fitView` elegirá un zoom menor: **es
  correcto**, es el overview-first ya firmado (RF-50).
- **H-3 · el puente:** la fila del Portafolio abre el Mapa de ese arnés **con la capa Mejora ya
  activa** y la caja del punto de mejora seleccionada. Reusa «Abrir en Mapa» (GAP-1, ya existe) + un
  parámetro de capa.
- **`CopyConfianzaEsUnaSola`** vive acá: es **el candado de D18** y es el único lugar que puede
  importar las dos entities.

**RF que cierra:** **RF-245** · RF-234 (el cableado) · H-3.
**Escenarios:** T-16 (superset del canvas).
**Capabilities:** amplía CAP-139 (`type: extend`); `pages/` no lleva story (hoy hay 0 en el repo).

**Aceptación:**

```gherkin
Escenario: la geografía es la misma
  Dado dev-full-cycle en capa estructura y en capa mejora
  Entonces el conteo de .node, de .lane y de path del EdgeLayer es idéntico
  Y los 6 asserts de la story firmada `Dogfood` se repiten y pasan

Escenario: volver no pierde nada
  Cuando se conmuta estructura → mejora → estructura con un nodo seleccionado
  Entonces el nodo sigue con aria-pressed="true"
  Y no queda ninguna .mej-cifra en el DOM
```

**V&V:** las 4 de §2.13.
**Demostración:**
```bash
cd web && npx vitest run --project=storybook src/widgets/map-canvas/ui/map-canvas.stories.tsx
pnpm --dir . run verify      # typecheck + lint + depcruise + fsd + stylelint
npx vitest run --project=storybook      # las 125 + las firmadas
```

---

### Tramo C — Cierre

---

#### T38 · E2E contra el daemon vivo y contra la app instalada

**Depende de:** T37 (o T27 si el Tramo B se difiere).
**Objetivo:** que el operador **instale la app y vea funcionando lo que se construyó** — no un dev
server, no un mock. La receta completa está en **§5**; este ticket la ejecuta y versiona su
evidencia.

**Archivos (nuevos):**

- `docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-e2e/RECETA.md` — la corrida
  real, con salida observada (el formato de `verificacion-2026-07-26/CADENA-E2E.md`).
- `.../verificacion-e2e/evidencia/` — capturas `despues-*.png` + `salud-antes.json` /
  `salud-despues.json`.

**Diseño:** ejecutar **§5** de este plan, en sus cinco pasadas. Dos cosas que ya están probadas y
**no se rehacen**: la cadena `bundle.sh → :4200 → navegador` (CADENA-E2E, corrida el 2026-07-26) y
la captura **«antes»** `verificacion-2026-07-26/evidencia/baseline-mapa-antes.png`, que es el
término de comparación del gate. **Se compara contra ella; no se reconstruye.**

**RF que cierra:** RF-236 (legibilidad), RF-247/248 (copy), RF-255 (que el chat se abra de verdad),
T-16 (geografía), T-22 (escala de grises) — los que `plan-storybook.md` §4.2/§4.3 declara **no
cerrables con una story**.
**Escenarios:** B5 · B6 · I3 · I4 · I11 (los marcados `M`, receta escrita) · A1 · A2 (`E2E`).
**Capabilities:** ninguna nueva; el `dev_preview.e2e_test` de CAP-139 apunta a la receta.

**Aceptación:**

```gherkin
Escenario: el binario instalado sirve exactamente lo construido
  Dado make installer + make dev-sync recién corridos
  Entonces GET /api/version del daemon devuelve la versión nueva
  Y sha256(~/.local/bin/arnesia) == sha256(bin/arnesia) del build
  Y el index.html servido en :4200 es byte-idéntico al de web/dist
  Y el puerto 5173 (vite dev) NO está escuchando

Escenario: la telemetría se mueve con una corrida real
  Dado un HOME de prueba y un daemon en puerto efímero
  Cuando se corre una sesión real desde el Dock
  Entonces `arnesia telemetria salud` pasa de 0 recibidos a N > 0
  Y `arnesia telemetria resumen` muestra costo con atribucion = exacta
```

**V&V:** click-through humano + los comandos de §5. No hay test que lo cubra: es la definición de
E2E contra lo instalado.
**Demostración:** §5, pasadas 1 a 5, con la evidencia versionada.

---

#### T39 · Cierre del paquete: capabilities, cifras y PARIDAD

**Depende de:** T38.
**Objetivo:** que el paquete quede cerrable — SSoT funcional al día, cifras generadas (no
tecleadas) y la matriz de paridad lista para el gate.

**Archivos:**

- `docs/product/capabilities/INDEX.md` — regenerado.
- `docs/product/checkpoint.md` — regenerado por `scripts/estado.sh` (lo hace el hook `pre-commit`).
- `docs/product/BACKLOG.md` — cerrar «telemetría JSONL → indexer real» y «telemetria-de-nacimiento»;
  reclasificar «capa Desempeño» de `bloqueo` a **`deuda de diseño`** (la señal llega, falta el
  diseño — J2/D12.3); dejar abiertos D9.9, G4, `tauri#11992` y los huecos de §7.
- `docs/product/LEDGER.md` + `docs/product/ledger/HS-27.md` — la ficha del paquete.
- `.../2026-07-24-telemetria-embebida-otel/PARIDAD.md` (nuevo) — la matriz
  **mockup ↔ componente ↔ story=test ↔ RF**, fila por fila, con la sección **«Desviaciones
  registradas»**.
- `.../INDEX.md` — estado final y «Retomar aquí».

**Desviaciones que van declaradas en `PARIDAD.md`, no coladas** (ya identificadas):

1. Slot `Tokens` → `Mejora` — **desviación de un baseline firmado en PARIDAD** (D17.1, `mockups/INDEX.md` regla 3).
2. `[Aplicar]` → `[Proponerlo en el chat]` — cambio de etiqueta sobre una decisión firmada (J-5).
3. `inspector.stories.tsx:92` `toHaveLength(3)` → `4` — **la única modificación a una story firmada**.
4. Espaciado por token en vez de los `10px/14px` del mockup (design §2.2).
5. Inspector a **340 px**, no 380 (design §2.6).
6. Sin `overflow-x` en el canvas: pan y zoom (design §6.2).
7. Copy del nodo MCP corregido (H-10) y **fila `conocimiento` eliminada** (D19).
8. `arnés × puesto` se realiza como `(identidad, instalación)` con el puesto derivado del `rol` (D20).
9. Contrastes: disclaimer en `--foreground`, marca grave sobre `--card` (D21).

**RF que cierra:** ninguno; los cierra todos formalmente.
**Capabilities:** verifica las 22 (R1·R2·R3·R4) — no crea.

**Aceptación:**

```gherkin
Escenario: el SSoT funcional está completo y consistente
  Entonces existen las 22 hojas CAP-118…CAP-139 salvo las declaradas como "extiende una existente"
  Y ningún puntero cuelga (archivo y símbolo)
  Y ningún archivo nuevo quedó huérfano
  Y ningún estado contradice su evidencia
  Y las cifras del checkpoint no están stale
```

**V&V:** `TestCapabilityPointersResolve` · `TestCapabilityPointerSymbolsResolve` ·
`TestCapabilityCoverage` · `TestCapabilityStatusConsistent` · `TestCapabilityPointersStable` ·
`scripts/estado.sh --check`.
**Demostración:**
```bash
cd /home/chalreme/Proyectos/harness-studio
go test ./docs/architecture/fitness/ -run TestCapability -v
bash scripts/estado.sh --check          # exit 0 = las cifras del checkpoint no están stale
go run ./cmd/arnesia conformance --todo | tail -3
cd web && pnpm run verify && npx vitest run --project=storybook
```

---

## 4 · La regla del control positivo

`arquitectura-modulo.md` §14 la fija, y **se paga sola**: los tres primeros intentos de H9 dieron
negativo y **el negativo era falso** — un receptor viejo seguía ocupando el puerto. Sin control, dos
de los tres resultados de H10 (que son negativos) serían indistinguibles de un receptor caído.

> **Regla, y es requisito de revisión de cada ticket que la toque:**
> **ningún test de este módulo aserta solo `len(recibidos) == 0`.**
> Todo escenario cuyo éxito sea «no llegó nada» necesita, **en la misma corrida**:
> 1. un **control positivo** — un evento con marcador único que **sí** tiene que llegar;
> 2. **un receptor de puerto efímero** (`httptest.Server` en `:0`, el puerto se le pasa al sujeto) —
>    no un puerto fijo que otro proceso puede estar ocupando;
> 3. **marcadores distintos por variante** — dos variantes con el mismo marcador no se distinguen si
>    una filtra a la otra.
>
> La forma canónica del assert es
> `len(conMarcadorA) == 0 && len(conMarcadorB) == 1`, no `len(recibidos) == 0`.

**Los tickets que la tocan, y qué escenario de cada uno:**

| ticket | escenario «no llegó nada» | control positivo obligatorio |
|---|---|---|
| **T3** | el source-scan no encuentra parsing de JSONL | fixture con una llamada que **sí** debe detectar |
| **T9** | la PII no aparece en el evento | un campo de la allowlist que **sí** está |
| **T10** | el marcador del prompt no aparece en lo emitido | `session_id`/`prompt_id` que **sí** están |
| **T11** | la PII no aparece **como subcadena del `.db`** | un `sesion_id` con marcador que **sí** aparece |
| **T14** | tras borrar el arnés A no quedan filas de A | las filas de **B** siguen intactas |
| **T16** | 415 / 413 / 400 no guardaron nada | un POST válido con marcador distinto que **sí** entra |
| **T17** | el token de ingesta no abre `/api` (401) | el mismo token da **200** en `/v1/logs` |
| **T18** | la ficha no existe antes de escuchar | después de escuchar, existe **y su endpoint responde** |
| **T20** | sin daemon no llega nada (marcador A) | con daemon en `:0`, llega **exactamente 1** (marcador B) |
| **T23** | forward apagado ⇒ 0 conexiones salientes | encendido ⇒ **exactamente 1** |
| **T23** | el email no está en lo reenviado | `sesion_id` **sí** está |
| **T27** | `hook-proyecta-campos` falla con un hook que reenvía stdin | `telemetria-no-egresa` **pasa** en el mismo arnés (son dos checks, y hay que ver los dos veredictos) |
| **T31 · T32 · T36** | `queryByText("0,00")` es `null` | el texto que **sí** debe estar (`sin dato`, el copy del estado) |
| **T38** | WSL2/devcontainer no alcanza el loopback del host | una corrida desde el **host** que **sí** llega, misma sesión |

**Revisión:** un ticket de esta lista no se da por cerrado si su test asserta solo una ausencia. Es
el criterio explícito de revisión, no una recomendación.

---

## 5 · Verificación E2E y contra la app instalada

El operador exige poder **instalar la app y ver funcionando lo que se construyó**. La cadena ya se
probó de punta a punta **antes de que exista una línea del módulo** →
[`verificacion-2026-07-26/CADENA-E2E.md`](verificacion-2026-07-26/CADENA-E2E.md). Esta receta **se
construye sobre lo verificado**, no propone nada nuevo.

### 5.0 · Lo que ya está probado, y no se rehace

| paso | comando | resultado observado |
|---|---|---|
| build | `bash scripts/bundle.sh --daemon-only` | OK · identidad `0.2.21.2607261116` · `bin/arnesia` = **17,0 MB** |
| arranque | `./bin/arnesia serve` | escucha en `127.0.0.1:4200` |
| salud | `curl /healthz` | **200** |
| UI embebida | `curl /` | sirve el `index.html` real |
| navegador | Playwright → `http://127.0.0.1:4200/` | título **ArnesIA**, app montada |
| captura «antes» | screenshot | [`evidencia/baseline-mapa-antes.png`](verificacion-2026-07-26/evidencia/baseline-mapa-antes.png) |

**Por qué esto ya prueba que es la UI del binario y no el dev server:** `embeddedUI()`
(`cmd/arnesia/main.go:462`) sirve `web/dist` desde el `go:embed` y **devuelve `nil` si el build no
trae dist** — en ese caso `/` responde un 404 honesto y no habría app. **Que la página monte prueba
que la UI viaja dentro del binario.** Un E2E contra `vite dev` (`:5173`) **no demuestra nada sobre
lo instalado**; contra `:4200` del binario, sí.

**`baseline-mapa-antes.png` es el «antes» del gate de PARIDAD** — muestra el conmutador con
`Estructura | Tokens | Desempeño | Proceso` y los tres últimos apagados. **Se compara contra ella;
no se manda a reconstruirla al final.**

### 5.1 · Dos reglas duras de la corrida, verificadas en la prueba

1. 🔴 **HOME de prueba y puerto efímero, obligatorio.** El daemon de la prueba corrió contra el
   `~/.arnesia` **real del operador** (se vieron sus sesiones reales de `vitalia`). **Ningún ticket
   puede levantar el daemon de un E2E de telemetría contra el HOME real:** escribir eventos de
   prueba ahí **contamina los totales que la propia feature muestra**, que es exactamente el pecado
   que el paquete existe para no cometer. Siempre:
   `HOME=$(mktemp -d) XDG_CONFIG_HOME=$HOME/.config ./bin/arnesia serve --addr 127.0.0.1:0`.
2. ⚠️ **Un navegador de automatización puede quedar colgado reteniendo su perfil** y bloquear la
   corrida con `Browser is already in use`. **No es un fallo de la app** — se cierra por PID
   (`pkill -f 'chrome.*--user-data-dir'`) y se reintenta. Que esté escrito acá es para que nadie lo
   diagnostique como un bug del módulo.

### 5.2 · Pasada 1 — el backend solo, sin navegador (HOME de prueba)

```bash
cd /home/chalreme/Proyectos/harness-studio
export HOME_T=$(mktemp -d) && export XDG_CONFIG_HOME="$HOME_T/.config"
bash scripts/bundle.sh --daemon-only                      # → bin/arnesia
HOME="$HOME_T" ./bin/arnesia serve --addr 127.0.0.1:4299 &   # puerto propio, HOME propio
sleep 1
curl -s -H 'Host: 127.0.0.1:4299' http://127.0.0.1:4299/healthz            # 200
HOME="$HOME_T" ./bin/arnesia telemetria salud | jq .                       # recibidos: 0

# ── control positivo + negativo, en la misma corrida ──
curl -s -X POST http://127.0.0.1:4299/v1/logs -H 'Content-Type: application/json' \
  --data-binary @docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/evidencia/logs-run1.json
curl -s -X POST http://127.0.0.1:4299/v1/logs -H 'Content-Type: application/x-protobuf' --data 'x'   # 415, nombra el fix
HOME="$HOME_T" ./bin/arnesia telemetria salud | jq .    # recibidos > 0, rechazados_formato = 1
HOME="$HOME_T" ./bin/arnesia telemetria resumen | jq .  # costo con su confianza
```

**Éxito:** `recibidos > 0` **y** `rechazados_formato == 1` en la misma salud — el positivo y el
negativo juntos. Y **cero** ocurrencias de PII en el `.db`:

```bash
grep -c 'user.email\|account_uuid\|organization.id' "$HOME_T/.arnesia/telemetria.db" || echo "0 — OK"
grep -c 'sesion' "$HOME_T/.arnesia/telemetria.db"    # > 0 — control positivo: el archivo SÍ tiene datos
```

### 5.3 · Pasada 2 — el hook, con daemon y sin daemon

```bash
# con daemon (marcador B): tiene que llegar
echo '{"session_id":"e2e-B","prompt_id":"p1","hook_event_name":"Stop","last_assistant_message":"SECRETO-B"}' \
  | HOME="$HOME_T" ./bin/arnesia hook proceso; echo "exit=$?"        # exit=0, stdout vacío
HOME="$HOME_T" ./bin/arnesia telemetria resumen | jq '.turnos'        # subió

kill %1   # daemon abajo
# sin daemon (marcador A): NO tiene que llegar, y no puede tardar ni fallar
time (echo '{"session_id":"e2e-A","prompt_id":"p2","hook_event_name":"Stop"}' \
  | HOME="$HOME_T" ./bin/arnesia hook proceso); echo "exit=$?"        # exit=0, < 250 ms
```

**Éxito:** `exit=0` en los dos, `< 250 ms` en el segundo, `e2e-B` en el `.db` y `e2e-A` **no** —
control positivo y negativo con **marcadores distintos**, como manda §4. Y `SECRETO-B` **no** aparece
en el archivo.

### 5.4 · Pasada 3 — la UI real del bundle, en el navegador

```bash
bash scripts/bundle.sh                                   # build completo: web/dist ANTES del daemon
HOME="$HOME_T" ./bin/arnesia serve --addr 127.0.0.1:4299 &
```

Con **Playwright MCP** o **Chrome DevTools MCP**, navegar a `http://127.0.0.1:4299/` y ejercitar:

1. Conmutar a **Mejora** → el conmutador dice `Estructura | Mejora | Desempeño | Proceso`.
2. La franja muestra ventana, total, **el disclaimer visible sin hover** y la barra de cobertura con
   su rótulo.
3. Con el HOME de prueba vacío: **estado 1**, y **el texto de la página no contiene `0,00`**.
4. Tras la pasada 1: cifras en las cajas, **y ningún nodo no-caja con número**.
5. Abrir el inspector → 4ª tab → la suma de los buckets cierra con el total.
6. `qué guardamos` → el diálogo lista los campos **por nombre**.
7. Captura `despues-mapa-mejora.png`, **misma resolución que el «antes» (1440×900)**.

```bash
# el body no scrollea horizontal (design §6.3) — se evalúa en la página:
#   document.body.scrollWidth <= document.body.clientWidth
```

### 5.5 · Pasada 4 — la corrida real de `claude` (S1, la única que prueba el spawn)

```bash
# con la app abierta contra el HOME de prueba, abrir una sesión en el Dock y mandar un turno
HOME="$HOME_T" ./bin/arnesia telemetria resumen --arnes <id> | jq '{turnos, costo_reportado_micros, cobertura}'
HOME="$HOME_T" ./bin/arnesia telemetria mejoras --arnes <id> | jq '{puntos: (.puntos|length), no_aplican, no_medidos: (.no_medidos|length)}'
```

**Éxito:** `turnos > 0` · `costo_reportado_micros` **no nulo** · `cobertura.exacta > 0`
(los `arnesia.*` llegaron) · `no_medidos` tiene **7** entradas · cada `no_aplican` trae **motivo no
vacío**. Y el control de cobertura: `cobertura.esperados >= turnos`.

### 5.6 · Pasada 5 — contra la app INSTALADA (el último paso del paquete)

⚠️ **`make installer` bumpea la versión.** No se corre antes de que la feature exista: va **al final
del paquete**, en T38.

```bash
cd /home/chalreme/Proyectos/harness-studio
make installer          # bumpea Cargo.toml + tauri.conf.json + package.json → instaladores/vX.Y.Z/
make dev-sync           # ⚠ OBLIGATORIO: existe ~/.local/bin/arnesia
```

**Por qué `dev-sync` no es opcional:** el shell Tauri instalado (`web/src-tauri/src/lib.rs`,
`override_local`) **prefiere SIEMPRE `~/.local/bin/arnesia`** por sobre el sidecar del paquete —así
el self-update sin sudo puede pisarse a sí mismo. Sin `dev-sync`, **el operador probaría el binario
viejo creyendo que probó el nuevo**. El propio `make installer` lo avisa por consola.

**Cómo se prueba que el binario instalado sirve exactamente lo construido:**

```bash
V=$(make version)
pgrep -af 'arnesia serve'                                  # el que responde es el instalado, no `go run`
sha256sum ~/.local/bin/arnesia bin/arnesia                 # los dos hashes IGUALES
curl -s -H 'Host: 127.0.0.1:4200' http://127.0.0.1:4200/api/version | jq -r .version   # == $V
diff <(curl -s http://127.0.0.1:4200/) web/dist/index.html # sin diferencias: el HTML es el del bundle
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:5173 || echo "5173 cerrado — no hay dev server"
```

**Los cinco juntos son la prueba.** Cada uno solo se puede falsear: la versión podría venir de un
binario viejo con el mismo número, el HTML podría ser de un dev server en otro puerto, el proceso
podría ser un `go run`. Los cinco a la vez, no.

Después: abrir la app instalada (ventana Tauri, el shell mintea el token y spawnea el sidecar) y
hacer el **click-through humano** del gate de PARIDAD contra `baseline-mapa-antes.png`, con captura
a 1440×900 y **una en escala de grises** (T-22).

### 5.7 · Lo que esta cadena NO cubre, y hay que decirlo

| qué | por qué | dónde queda |
|---|---|---|
| Windows y macOS | no hay runner de esas plataformas | abierto: **G4** (cert de firma) y **`tauri#11992`** (notarización con `externalBin`) |
| WSL2 / devcontainer (B6) | requiere levantar un contenedor | receta `M` en T38, con control positivo desde el host |
| Determinismo de `plugin_id_hash` entre máquinas (E12/V7.2) | una sola máquina | **hueco declarado**, §7 |
| `/compact` (G5) y subagentes (G6) | no ejercitados | **huecos declarados**, §7 |

---

## 6 · Matriz de trazabilidad

**Los 55 RF (RF-232…RF-286) tienen ticket, escenario, capability y mecanismo. Ninguna fila queda a
medias.** Donde falta algo, no está acá: está en **§7 · huecos declarados**.

Leyenda del mecanismo: `S` story-test (`vitest --project=storybook`, headless) · `U` test unitario
colocado · `F` fitness (`docs/architecture/fitness/`) · `C` check de conformance · `E2E` corrida real
· `H` click-through humano del gate.

| RF | qué | ticket | escenario (`escenarios.md`) | capability | mecanismo · nombre exacto |
|---|---|---|---|---|---|
| RF-232 | slot `Tokens` → `Mejora`, encendido | T30 | — | CAP-139 | `S CapaMejoraDisponible` · `S CapaMejoraActiva` |
| RF-233 | Desempeño/Proceso apagados con el motivo REAL | T30 | J2 | CAP-139 | `S MotivosHonestos` |
| RF-234 | ventana temporal explícita | T32 · T37 | — | CAP-139 | `S Reposo` · `S VentanaCambia` |
| RF-235 | total con denominador | T32 | — | CAP-139 | `S TotalConDenominador` · `U TestGastoSinCaja` |
| RF-236 | disclaimer «estimado» en la superficie | T32 | F12 | CAP-139 | `S DisclaimerEnSuperficie` **+ `H` captura 1440×900** (§5.6) |
| RF-237 | cobertura, **4** segmentos | T29 · T32 | A7 | CAP-139 · CAP-128 | `S CuatroSegmentos` · `S CoberturaCuatroNiveles` · `U TestConciliacionCuentaLosNoLlegados` |
| RF-238 | solo las cajas llevan cifra | T31 · T37 | G15 | CAP-139 | `S CapaMejoraSoloCajasLlevanCifra` · `S SinDato*` (5) |
| RF-239 | cifra + participación | T31 | — | CAP-139 | `S MejoraCifraExacta` |
| RF-240 | barra de participación | T31 | — | CAP-139 | `S MejoraCifraExacta` · `S MejoraCajaSinCorridas` |
| RF-241 | la marca **nombra** el detector | T31 | — | CAP-139 | `S MejoraConMarcaDeFuga` · `S MejoraDosDetectores` |
| RF-242 | 4 confianzas, 4 copys | T29 · T31 | E6 · E8 | CAP-139 · CAP-127 | `S LosCuatroJuntos` · `S MejoraPorHuella` · `S MejoraPorProceso` · `U TestAtribucionPorHash` · `U TestAtribucionPorProceso` |
| RF-243 | «sin dato atribuible» con motivo | T31 | G15 | CAP-139 | `S SinDatoSubagente` · `SinDatoRegla` · `SinDatoMcp` · `SinDatoHook` · `SinDatoResto` |
| RF-244 | total por carril | T31 | — | CAP-139 | `S LaneConTotalMejora` · `S LaneSinDato` |
| RF-245 | overlay: nada se mueve | T31 · T37 | — | CAP-139 | `S CapaMejoraSupersetGeografia` · `S ConmutarNoPierdeSeleccion` · `S EstructuraIntacta` · **`H`** («se siente igual») |
| RF-246 | tarjeta solo si pasa A4 | T33 · T34 | G13 | CAP-139 · CAP-130 | `S DescartaSinContrafactual` · `S DetectorSinFixPropuesto` |
| RF-247 | titular en lenguaje del usuario | T33 | — | CAP-139 | `S CompletaB1Atencion` · `S TitularSinJerga` · **`H`** (calidad del copy) |
| RF-248 | lede: cuánto, sobre qué base | T33 | — | CAP-139 | `S CompletaB1Atencion` · **`H`** |
| RF-249 | contrafactual, no «gastaste X» | T33 · T15 | G12 | CAP-130 | `S CompletaB1Atencion` · `U TestB1BreakEvenTTL` |
| RF-250 | umbral algebraico citado | T33 · T15 | G12 | CAP-130 | `S CompletaB1Atencion` · `S CriticaP1SinUmbral` · `S CalculoAbierto` · `U TestB1BreakEvenTTL` |
| RF-251 | confianza con denominador | T33 | — | CAP-139 · CAP-127 | `S CompletaB1Atencion` · `U TestConfianzaDeAgregadoEsLaMinima` |
| RF-252 | sesgo declarado y **en contra** | T33 · T15 | D4 | CAP-130 | `S SesgoSubestima` · `S SesgoSobreestima` · `S SinSesgoIdentificado` · `U TestSesgoTieneDireccion` · `U TestTurnoCruzaHora` |
| RF-253 | **un** fix, concreto | T33 | — | CAP-139 | `S CompletaB1Atencion` |
| RF-254 | score versionado y visible | T33 · T27 | — | CAP-139 · CAP-130 | `S CompletaB1Atencion` · `C` (cambia la fórmula sin bumpear ⇒ falla) |
| RF-255 | `[Proponerlo]` **no escribe archivos** | T33 · T38 | — | CAP-139 | `S ProponerAbreChatNoEscribe` · `S ProponerDeshabilitadoFueraDeAlcance` · **`E2E`** §5.6 |
| RF-256 | descartar sin borrar el dato | T33 | — | CAP-139 | `S DescartarLlamaHandler` |
| RF-257 | severidad no depende del color | T33 | — | CAP-139 | `S SeveridadSinColor` · **`H`** captura en escala de grises (T-22) |
| RF-258 | 4ª tab, las tres intactas | T34 | — | CAP-139 | `S CuatroTabs` · `S NodoNoCaja` |
| RF-259 | desglose por bucket | T34 | — | CAP-139 | `S Completo` (assert **aritmético**: la suma cierra) |
| RF-260 | «no aplica» ≠ 0 | T34 | C16 | CAP-139 · CAP-120 | `S NoAplicaNoEsCero` · `S CeroLegitimo` · `F TestNoAplicaNoEsCeroEnElWire` |
| RF-261 | reportado vs. calculado | T34 · T7 | F6 · F7 · F8 | CAP-126 | `S ParidadCoinciden` · `S ParidadDivergen` · `S SinCostoDelRuntime` · `F TestDivergenciaDeCostoEsVisible` |
| RF-262 | el join a nivel nodo | T34 · T22 | G7 · G8 | CAP-128 | `S JoinCompleto` · `S SinSenalDeGate` · `U TestJoinPorSesionYTurno` |
| RF-263 | detectores, incluidos los que no aplican | T34 · T15 | G13 · G14 | CAP-129 | `S SeisDetectoresConEstado` · `S SieteNoMedidos` · `S DetectorB2Parcial` · `F TestDetectorQueNoAplicaTraeMotivo` |
| RF-264 | la tab declara su ventana | T34 | — | CAP-139 | `S VentanaHeredada` (y `últimas 14 corridas` **no** está en el DOM, J-3) |
| RF-265 | fila por arnés × puesto | T36 · T19 · T13 | E1 · E2 | `fe-portafolio` (extend) · CAP-137 | `S ConDato` · `S UnArnesDosPuestos` · `S SinPuestoDeclarado` · `U TestPuestoSaleDelRolDelArnes` |
| RF-266 | tendencia con texto equivalente | T29 · T36 | — | CAP-139 | `S EnAlza` · `S Estable` · `S ALaBaja` · `S PocasCorridas` |
| RF-267 | punto de mejora por fila | T36 | — | `fe-portafolio` (extend) | `S SinFugas` |
| RF-268 | «sin dato» honesto, no ordena como 0 | T36 | A8 | `fe-portafolio` (extend) | `S SinDato` · `S OrdenadaSinDatoAlFinal` |
| RF-269 | estado 1 · sin datos | T32 | A8 | CAP-139 | `S Estado1SinDatos` (`textContent` **no contiene** `0,00`) · `S SinCorridas` |
| RF-270 | estado 2 · cobertura parcial | T32 | A7 | CAP-139 · CAP-128 | `S Estado2CoberturaParcial` |
| RF-271 | estado 3 / 3b · fuera de ArnesIA | T32 · T15 | A2 · A3 | CAP-130 | `S Estado3S2SinInstrumentar` · `S Estado3bS2Instrumentado` · `S DetectorB1ApagadoEnS2` · `F TestS2DegradadoApagaLosDetectoresDeDinero` · `F TestS2InstrumentadoTieneDineroYNoTieneSplit` |
| RF-272 | estado 4 · otro runtime | T32 · T6 | A5 · A6 | CAP-125 | `S Estado4OtroRuntime` · `S RuntimeNoSoportado` · `U TestCostoSoloCalculado` |
| RF-273 | estado 5 · catálogo viejo | T32 · T6 | F3 · F4 · F5 | CAP-125 | `S Estado5CatalogoViejo` · `U TestRefrescoFallidoNoRompe` |
| RF-274 | estado 6 · por huella | T29 · T31 · T13 | E5 · E6 | CAP-127 | `S PorHash` · `S MejoraPorHuella` · `U TestAtribucionPorHash` · `U TestHashDesconocido` — ⚠ el **determinismo entre máquinas** es hueco §7 |
| RF-275 | estado 7 · qué guardamos + borrado | T35 · T14 | I10 · I11 | CAP-139 · CAP-124 | `S Reposo` · `S CamposDesplegados` · `S Confirmacion` · `S Exito` · `F TestAllowlistNoPersistePII` |
| RF-276 | ARIA del conmutador + motivo accesible | T30 | — | CAP-139 | `S MotivoAccesible` (igualdad `title` ⟷ `sr-only`) |
| RF-277 | contrato de tabs vigente | T34 | — | CAP-139 | `S TabMejoraContratoAria` · `S CambioDeNodoVuelveAResumen` |
| RF-278 | foco visible en todo control nuevo | T35 (+ todos) | — | CAP-139 | `S FocoAtrapado` + axe en `error` en cada archivo |
| RF-279 | texto equivalente en barras y sparklines | T29 · T31 | — | CAP-139 | `S CuatroSegmentos` · `S EnAlza` · `S MejoraCifraExacta` (`aria-label` de participación) |
| RF-280 | ningún estado depende solo del color | T29 · T31 · T33 · T34 · T36 | G13 | CAP-139 | `S LosCuatroJuntos` · `S SeveridadSinColor` · `S SeisDetectoresConEstado` · `S SinFugas` · **`H`** escala de grises |
| RF-281 | un solo formato de dinero | T29 | — | CAP-139 | `S Estandar` · `S SeparadorYDosDecimales` · `S MenorAlCentavo` · `S Ausente` |
| **RF-282** | **allowlist en los DOS caminos** | **T9 · T10** (+T11) | C3 · C16 · E7 · I9 · I10 | **CAP-121** | `F TestAllowlistNoPersistePII` (**subcadena sobre el `.db`**) · `F TestHookNoReenviaContenido` · `F TestAllowlistEsListaNoSugerencia` · `F TestCwdDesconocidoNoSeGuardaCrudo` · `F TestToolResultBytesSePersiste` |
| **RF-283** | **retención con TTL y borrado que borra** | **T14** (+T19, T35) | H4 · H5 · H11 | **CAP-124** | `U TestPurgaRespetaTTL` · `U TestBorradoPorArnesTambienLimpiaRollup` · `U TestAvisoDeTamano` · `S RetencionDesdeConfig` |
| **RF-284** | **dos niveles de egreso; el arnés ni egresa ni guarda de más** | **T23 · T27** (+T17, T20) | I1 · I2 · I8 · I9 · F5 | **CAP-134 · CAP-131 · CAP-133** | `F TestForwardApagadoPorDefault` · `F TestForwardNoReenviaCrudo` · `F TestForwardSoloPorFlagDelOperador` · `C telemetria-no-egresa` · `C hook-proyecta-campos` |
| **RF-285** | **confianza + los DOS costos persistidos** | **T13 · T7** (+T21) | E9 · E10 · E11 · F6 | **CAP-126 · CAP-127 · CAP-128** | `F TestDobleCostoSePersisteEntero` · `F TestCifraLlevaConfianza` · `U TestJoinPorSesionYTurno` · `U TestConfianzaDeAgregadoEsLaMinima` |
| **RF-286** | **«no aplica» se persiste como ausencia** | **T4 · T11 · T12** | C16 · D4 | **CAP-120 · CAP-123** | `F TestNoAplicaNoEsCeroEnElWire` · `F TestNoAplicaSobreviveAlRollup` · `U TestEventoNoTieneCamposDeIdentidad` |

### 6.1 · Escenarios de `escenarios.md` → ticket (los que no aparecen arriba)

| familia | escenarios | ticket | mecanismo |
|---|---|---|---|
| A · captura | A1 · A2 | T21 · T13 | `F TestSpawnInyectaTelemetria` · `F TestEscenarioSeDerivaDeLaSenal` · `E2E` §5.5 |
| A | A4 | T20 | `U TestHookFailOpenSinDaemon` (+ control positivo) |
| B · red | B1 · B12 | T20 | `U TestHookFailOpenSinDaemon` · `U TestHookFichaHuerfana` |
| B | B2 | T38 | `E2E` con daemon apagado a mitad de corrida |
| B | B3 · B4 · B13 | T18 | `F TestFichaSoloTrasEscuchar` · `U TestFichaSeRelePorInvocacion` · `U TestFichaNoEscribibleDegradaHonesto` |
| B | B5 · B6 | T38 | `M` receta escrita (reinicio en otro puerto · devcontainer, **con control positivo desde el host**) |
| B | B7 · B8 · B9 · B14 · B15 | T17 | `F TestOTLPBajoLosTresGates` · `F TestTokenDeIngestaNoAbreLaAPI` · `F TestOTLPAceptaSinTokenBajoLoopback` · `U TestModoEstrictoApagaS2Instrumentado` |
| B | B10 · B11 | T11 | `U TestDosEscritoresSobreElMismoDB` · `U TestRutasPorHome` |
| C · wire | C1…C15 | T8 · T16 | ver T8/T16 (10 tests + `FuzzDecodificarLogs`) |
| D · tiempo | D1…D5 | T11 · T12 | `U TestRelojHaciaAtrasNoRompeLaVentana` · `U TestSinTimestampDelEmisor` · `U TestTurnoCruzaHora` · `U TestTodoEnUTC` |
| E · identidad | E3 · E4 | T13 · T27 | `U TestArnesSinSello` · `U TestConformanceArnesSinBloqueTelemetria` |
| F · dinero | F9 · F10 · F11 | T7 | los tres tests que los bugs ajenos dictaron |
| G · proceso | G1 · G2 · G3 · G9 · G10 · G11 | T15 · T22 | `U TestGastoSinCaja` · `TestB6SesionAbandonada` · `TestB2CostoDeLaRotacion` · `TestCorridaCancelada` · `TestReintentosSuman` · `TestB3CambioDeModelo` |
| H · almacén | H1…H13 | T11 · T12 · T14 · T16 | ver T11/T12/T14/T16 |
| I · operación | I3 · I4 | T38 | `M` receta de desinstalación · `M` self-update con el daemon corriendo |
| I | I5 | — | ✅ **ya verificado** (ANEXO H10.3, con hook de control). No se re-corre |
| I | I6 · I7 · I12 | T20 | `U TestHookNoTardaNiFalla` · `U TestHookStdoutVacio` |
| I | I11 | T24 · T38 | `M` abrir el `.db` y listar columnas |

---

## 7 · Huecos declarados

No se fabrican filas para taparlos. Cada uno dice **qué falta**, **por qué** y **qué se ve mientras
tanto**.

| # | hueco | por qué está abierto | qué se ve mientras tanto | dueño |
|---|---|---|---|---|
| **G1** | **A20 — dónde vive el bloque `env`** que instrumenta `s2-instrumentado` | Es decisión de producto: la opción B **escribe archivos de un tercero** y eso tiene doctrina firmada en contra (A8 + guardrail del chat embebido). ⚡ **Ya no hay verificación que la evite**: H10.1 probó que un plugin **no** puede aportarlo (0 payloads vs 2 del control positivo) | `s2-degradado`: proceso medido, **dinero no disponible con motivo**. Nunca un cero | **operador** — §8 · **parada P1** |
| **G2** | E12 / V7.2 · determinismo de `plugin_id_hash` **entre máquinas** | una sola máquina | El diseño **no lo asume**: la tabla se aprende local (`como_se_aprendio="spawn-controlado"`) y la UI dice «por huella». Si resultara determinista, se sembraría al instalar — **mejora, no cambio de diseño** | verificación en 2ª máquina |
| **G3** | G5 · `/compact` — cuantificar su costo (B7) | no ejercitado (V7.4) | `query_source="compact"` etiqueta el turno y aparece en el desglose; **no hay detector**, y B7 sale en `no_medidos[]` | corrida dirigida |
| **G4** | G6 · subagentes (`isSidechain`) — overhead (B8) | no ejercitado (V7.4) | los `api_request` del sidechain comparten `session.id` ⇒ **suman al total**; no se separan. B8 en `no_medidos[]` | corrida dirigida |
| **G5** | §10.2 · nombres reales de `OTEL_LOGS_EXPORTER` / `OTEL_LOGS_EXPORT_INTERVAL` | no se verificaron contra `claude 2.1.220` (se usó el default) | **No bloquea**: hay default. Si no existen, se aceptan los del runtime y **se documenta** | T21 — §8 · **parada P3** |
| **G6** | J-6 · el número del TTL | D15.3 firmó «TTL por default» **sin número**; el `90` del mockup es PROPUESTO | La UI lo lee de la config y la story asserta con `45` que **no está hardcodeado** | **operador** — §8 · **parada P2** |
| **G7** | H-8 · el `aria-label` de la cobertura (17) no cierra con el denominador (61) | el mockup no lo resolvió | Este plan usa **18** en la cobertura y **declara la unidad** en el texto; los `61` de la franja son de **otra fixture, a propósito**, para no fingir una coherencia que el mockup no tiene | iteración 2 del mockup |
| **G8** | Los **otros 7 detectores** (B5·B7·B8·B9·B10·B11·B12·B13) | no pasan A4 (falta el fix concreto o la cotización) | Lista `no_medidos[]` **visible**, con `no medido todavía`. **Jamás «sin hallazgos»** | posterior al MVP |
| **G9** | **Capa Desempeño** a nivel Mapa | **ya no falta señal** (`duration_ms` por request **y por herramienta**, H6/H8): falta el **diseño** | Slot `disabled` con el tooltip honesto de RF-233. Se reclasifica en el BACKLOG de `bloqueo` a **deuda de diseño** | T39 lo reclasifica |
| **G10** | Adaptadores de **otros runtimes** | H1: no competimos en cobertura; el eje es arnés × empresa × puesto | «todavía no medimos ese runtime» (A5) — no una fila en cero | fuera de alcance |
| **G11** | **S3** (arnés en máquina de un cliente sin ArnesIA) | D12.1 lo descartó: choca con `superficie-local-confinada` y abre consentimiento/GDPR | nada; no hay canal | paquete propio si se reabre |
| **G12** | **Windows / macOS**: firma de código (G4) y `tauri#11992` (notarización con `externalBin`) | no hay runner de esas plataformas; el cert es USD 150-300/año + HSM | Linux funciona; los otros canales no se abren | **operador** + runner |
| **G13** | **D9.9** — ¿parser propio o shell-out a `ccusage`? | posterior al MVP; ccusage **descarta** `cache_creation_input_token_cost_above_1hr`, justo el campo que habilita B1/B12 | el catálogo embebido propio cubre el MVP | **operador** |
| **G14** | `docs/product/capabilities/` tiene 10 módulos que **no** están en `project.config.yaml domain_modules` | drift preexistente, no lo causa este paquete | T1 lo registra en el BACKLOG; **no se resuelve acá** | otro paquete |
| **G15** | El **mockup sigue en iteración 1** | su gate 🧑‍⚖️ no está firmado | **El Tramo B entero está bloqueado.** Los 14 huecos (`spec.md` §I) y las 10 contradicciones (§J) se resuelven en la iteración 2 | **operador** — §8 · **parada P0** |

---

## 8 · Riesgos y dónde el constructor tiene que **parar a preguntar**

> Las **paradas** no son sugerencias. En cada una, el constructor **detiene el ticket, escribe la
> pregunta en `decisiones.md` y espera**. Decidir solo cualquiera de estas es exceder el mandato.

### 🛑 Paradas obligatorias

| # | ticket | qué se pregunta | por qué no lo puede decidir el constructor |
|---|---|---|---|
| **P0** | **antes de T28** | **¿Está firmado el 🧑‍⚖️ del mockup?** | `spec.md` §Estado es explícito: *«ningún RF 🎨 se construye antes de su 🧑‍⚖️»*. Construir píxeles sin la firma rompe la disciplina §10 y se paga en retrabajo. **Si no está firmado, el Tramo B no arranca — se cierra el Tramo A y se para.** |
| **P1** | **T20 / T27** | **A20 — ¿dónde vive el bloque `env`?** Opción **A** (repo del propio arnés: cero fricción, cobertura **parcial** —cubre el desarrollo del arnés, no su uso—, es **nuestro** archivo) · Opción **B** (proyecto del usuario: cobertura **completa**, pero **escribe archivos de un tercero** ⇒ choca con A8 y con el guardrail del chat embebido, y exige consentimiento explícito + backup + botón inverso). **Recomendación de `arquitectura-modulo.md` §7.5: A ahora, B solo por pedido explícito.** ⚡ **No hay tercera vía** (H10.1: 0 payloads con el plugin instalado, 2 en el control positivo) | Una de las dos ramas **escribe en el árbol de un tercero**, y la doctrina de la casa dice que eso **se pide, no se hace**. **Sigue ABIERTA y es del operador.** Sin ella, el mecanismo de obligación de `s2-instrumentado` no se puede terminar; **el resto del paquete sí** (todo corre en `s2-degradado`, que muestra proceso y dice por qué no hay dinero) |
| **P2** | **T14 / T25** | **¿Cuál es el TTL de retención por default?** El `90` del mockup es **PROPUESTO, no firmado** (J-6: D15.3 firmó «TTL por default» **sin número**) | Es una decisión de producto sobre datos del usuario. El constructor implementa el **flag** y la UI lo **lee de la config**; el número lo pone el operador. Hasta entonces el default es 90 y **se rotula como propuesto en la UI y en el `PARIDAD.md`** |
| **P3** | **T21** | **¿Existen `OTEL_LOGS_EXPORTER` y `OTEL_LOGS_EXPORT_INTERVAL` en `claude 2.1.220`?** No se verificaron (§13.2 ítem 2) | Es una verificación de 5 minutos, no una decisión — pero **si no existen y se dejan en el env, el spawn puede fallar en silencio**, que es el peor modo de falla de este módulo. **Se verifica en vivo antes de fijarlas**; si no existen, se aceptan los intervalos por default **y se documenta**. Regla de §4: la verificación va con **control positivo** |
| **P4** | **T31 / T33** | **¿La desviación «`Tokens` → `Mejora`» sigue aceptada?** Es un cambio de etiqueta sobre un **baseline firmado en PARIDAD** (`mockups/INDEX.md` regla 3) | Ya está firmada como D17.1, pero **va al gate declarada, no colada**. Si el operador la retira, T30 cambia y T32-T37 con él. Se confirma **antes** de escribir la primera story del Tramo B |

### ⚠️ Riesgos, con su mitigación ya escrita

| # | riesgo | probabilidad · impacto | mitigación en el plan |
|---|---|---|---|
| **R1** | **El presupuesto de binario se pasa.** `pdata` costaría +10,79 MB medidos; el decodificador propio, +0,49 MB | baja · alta | `TestPresupuestoDeBinario` en CI (T25): **+1,5 MB máx.** y **≤ 25 MB absoluto**. Si se pasa, el ticket no cierra |
| **R2** | **El spawn no llega: el default de Claude Code es gRPC :4317.** Sin `OTEL_EXPORTER_OTLP_PROTOCOL=http/json` **no llega nada, en silencio** | media · alta | `TestSpawnInyectaTelemetria` asserta las 9 vars y el valor `http/json`. Y la conciliación (A9) convierte el silencio en «N turnos no reportaron» |
| **R3** | **Un negativo falso.** Ya pasó: tres intentos de H9 dieron negativo por un receptor viejo pegado al puerto | **alta** · alta | **§4 entera.** Puerto efímero + control positivo + marcadores distintos, en cada uno de los 14 puntos listados |
| **R4** | **Contaminar el `~/.arnesia` real del operador** con eventos de prueba | **alta** · alta | §5.1 regla 1: **HOME de prueba y puerto efímero, obligatorio** en todo E2E. Escribir en la base real contamina los totales que la feature muestra |
| **R5** | **Probar el binario viejo creyendo que se probó el nuevo.** El shell instalado prefiere `~/.local/bin/arnesia` sobre el sidecar | **alta** · alta | §5.6: **`make installer` + `make dev-sync`**, y los **cinco** comandos de verificación cruzada (versión · sha256 · index.html · pgrep · 5173 cerrado) |
| **R6** | **`steiger` rompe CI** por el cross-import de entities | media · media | **D18**: props primitivas, cero import. `pnpm run fsd` en el criterio de cierre de T29, T31 y T37 |
| **R7** | **El gate de a11y solo mira el tema claro** y el fallo del otro tema se escapa | media · media | **D21 ítem 3**: `ReposoDark` · `CompletaB1AtencionDark` · `ConDatoDark`. Total 124 → **125** |
| **R8** | **R1 falla por un símbolo que no existe** (`conductor.go#parseResult`, los punteros viejos de CAP-139) | **alta** · media | **T21** extrae `parseResult` con ese nombre exacto; **D22** corrige los cuatro punteros de CAP-139. `TestCapabilityPointerSymbolsResolve` en el criterio de cierre |
| **R9** | **`go-arch-lint` falla** con componentes cuyo `in:` no matchea archivos | media · baja | §10.4: el bloque se agrega **en el mismo commit que crea los paquetes** (T26), no antes |
| **R10** | **La story firmada `Tabs` rompe CI** al agregar la 4ª tab (`toHaveLength(3)`) | **alta** · baja | T34 la cambia a `4` **en el mismo commit**, declarada en `PARIDAD.md` como la única modificación a una story firmada |
| **R11** | **Se persiste PII** por un camino que nadie miró | baja · **crítica** | Allowlist **default-deny** en los dos caminos, aplicada **dos veces** (A13), + `TestAllowlistNoPersistePII` que busca **como subcadena del archivo `.db`**, no en structs. Es la única forma de asertar una ausencia sin confiar en la forma del código |
| **R12** | **El hook rompe un turno del usuario** | baja · **crítica** | Contrato §7.1: **exit 0 siempre**, stdout vacío, 250 ms, fail-open. `TestHookNoTardaNiFalla` asserta **todas** las ramas, incluidas las de error. H10.3 cubre que el binario *falte*; que *falle* es obligación nuestra |
| **R13** | **Un navegador de automatización colgado** se diagnostica como bug de la app | media · baja | §5.1 regla 2: `Browser is already in use` **no es un fallo de la app**; se cierra por PID y se reintenta |
| **R14** | **El mockup no se firma** y el Tramo B queda parado indefinidamente | media · media | El Tramo A **entrega valor solo**: `arnesia telemetria resumen/mejoras/salud` (T24) es superficie observable sin FE, mismo patrón que `arnesia portafolio` en Slice 0. El paquete puede cerrar una etapa sin píxeles |

---

## 9 · Objeciones

Cosas firmadas que este plan **ejecuta igual**, con la objeción escrita para que quede registro.

1. **`arquitectura-modulo.md` §1 dibuja `web/src/widgets/mejora/` y CAP-139 apunta a
   `tarjeta-mejora.tsx#TarjetaMejora`.** No se construye así (D22): una slice hermana chocaría con
   `no-sibling-widget-imports`, que está en `error`. La objeción no es al destino sino al proceso —
   **tres documentos del mismo paquete nombraron el mismo componente de tres maneras**, y eso se
   caza solo cuando R1 rompe. Vale la pena una pasada de nombres antes de firmar el próximo par
   spec+design.

2. **La regla A4 («una tarjeta necesita UN fix concreto») deja B4 en un lugar raro.** B4 es *«gasto
   por arnés × empresa × puesto»* y su fix es *«dónde mirar»* — que no es un ajuste concreto como
   `cache_ttl: 1h`. Está firmado en D16.1 y entra, pero **B4 es el único de los seis cuya tarjeta no
   propone una acción**; si en la construcción resulta que no pasa su propio filtro, el lugar
   honesto es la **franja** (el total con su cobertura), no una tarjeta.

3. **`Estimado: true` es siempre verdad mientras la fuente sea `cost_usd_micros`, y también lo será
   cuando el costo lo calcule nuestro catálogo.** El contrato de §3.2 lo ata solo al primer caso. Un
   costo derivado de un catálogo de precios embebido y posiblemente viejo **también es una
   estimación**, y la UI debería decirlo igual. El plan lo implementa como está firmado; la
   objeción queda.

4. **El TTL de 90 días viaja como default sin estar firmado** (J-6). Implementarlo como default —
   aunque sea configurable y esté rotulado— **crea el hecho consumado**: cuando el operador decida,
   ya va a haber bases con 90 días de retención. Sería más limpio arrancar sin default y **exigir**
   el flag hasta que haya número. No se hace porque rompería la instalación silenciosa que D6 exige.

5. **`escenarios.md` I5 se da por verificado y no se re-corre.** Es correcto (H10.3 lo probó con
   control positivo), pero la garantía es **de una versión de Claude Code** (`2.1.220`). Un cambio
   de runtime la puede retirar sin avisar, y no hay ningún check que lo detecte. **El contrato
   «exit 0 siempre» nos cubre igual** — por eso se ejecuta así — pero conviene no leer H10.3 como
   una propiedad permanente del mundo.
