# Propuesta de mockup — capa «Mejora» del Mapa (2026-07-26)

> **Estado: PROPUESTA, sin dibujar.** Se presenta antes de forkear el `.html` porque tres
> elecciones necesitan tu visto: el slot del conmutador de capas, la segunda superficie
> (Portafolio) y qué se hace con el botón «Aplicar».
> Línea base leída como manda `mockups/INDEX.md`: SSoT = Storybook
> (`map-canvas.stories.tsx` · `map-bar.tsx` · `inspector.tsx`), baseline derivado =
> `arnesia-mapa-baseline.html` @`a01d845`.

## El objetivo, y qué implica para el dibujo

La frase que el MVP tiene que poder decir (D12.2):

> *«este arnés, en este puesto, quema $X — y el 60 % se va en la caja Y, que falla el gate 3 de
> cada 4 veces»*

Eso **no es un badge de tokens por nodo**. Tres consecuencias de diseño, todas forzadas por la
verificación en vivo:

1. **La frase vive en DOS superficies.** «este arnés **en este puesto**» es comparación entre
   arneses ⇒ vive en el **Portafolio**. «el 60 % se va en la caja Y» es dentro de un arnés ⇒ vive
   en el **Mapa**. El Mapa solo no puede decir la frase.
2. **Solo las CAJAS pueden llevar número.** La granularidad honesta es `arnés × caja × sesión ×
   turno` (V3). Poner una cifra en un nodo de conocimiento, una regla o un hook sería inventarla.
   Esos nodos se dibujan **«sin dato atribuible»**, mismo patrón que Hallazgos/Contenido hoy.
3. **Un número sin fix es opinión** (regla A4). Cada cifra que se muestra viene con **un** fix
   concreto y su contrafactual, o no se muestra.

## Superficie 1 — el Mapa, capa «Mejora»

### 1.1 · El conmutador de capas ⚠️ *necesita tu decisión*

Hoy: `Estructura` (activa) · `Tokens` · `Desempeño` · `Proceso` — las tres últimas `disabled`
(`web/src/widgets/map-canvas/model/layers.ts`).

Propuesta: **renombrar el slot `tokens` a «Mejora» y encenderlo**; `Desempeño` y `Proceso` siguen
`disabled` pero con tooltip honesto que diga qué falta (diseño, no señal — la señal ya llega).

⚠️ **Es una desviación del baseline firmado**: la regla 3 de `mockups/INDEX.md` prohíbe quitar
artefactos firmados. No se quita ninguno —los 4 slots siguen— pero **una etiqueta firmada cambia**.
Va al gate como desviación explícita, no colada.

### 1.2 · La barra, cuando la capa está activa

Una franja de contexto (no reemplaza la barra, se suma):

```
[ventana: 7 días ▾]   USD 4,82 estimado · 61 corridas · 12 sesiones
                      ⓘ estimado por el runtime, no es facturación
                      cobertura: 12 corridas con atribución exacta · 3 por huella · 2 sin dato
```

Tres cosas deliberadas: la **ventana temporal** (sin ella «cuánto gasta» no significa nada), el
disclaimer de H4 (`cost.usage` es *"Estimated cost"*), y la **cobertura** — que es lo que impide
leer el total como si fuera completo.

### 1.3 · El canvas: misma geografía, tres marcas nuevas

**No se mueve nada de sitio.** Guardia · carriles por fase · Base · franja de artefactos · edges,
todo idéntico. Encima:

| marca | dónde | qué dice |
|---|---|---|
| **barra de participación** | borde inferior de cada **caja** | qué proporción del gasto del arnés se va ahí |
| **cifra + %** | esquina de la caja | `USD 1,92 · 40 %` |
| **marca de fuga** | esquina opuesta | el detector concreto: `re-warm TTL` / `modelo cambiado` / `rechazo en gate` — **nunca un ⚠ genérico** |
| **atenuado + «sin dato atribuible»** | conocimiento · reglas · hooks · artefactos | honestidad, no cero |
| **subrayado punteado** | cualquier cifra con `atribucion_confianza ≠ exacta` | al hover: «atribuido por proceso» / «por huella del arnés» |

Ese último es el que hace que la pantalla no mienta: **cada número carga cómo se atribuyó**.

### 1.4 · La tarjeta de punto de mejora — el corazón del entregable

Al tocar una marca de fuga (o desde una lista lateral):

```
⚠  La caja «Diseñar spec» reescribe el cache en cada corrida

   Gasta USD 0,84 de los USD 1,92 de la caja (44 %) escribiendo cache que se vence
   antes de volver a leerse.

   Contrafactual   con TTL de 1 h, las mismas 14 corridas costaban USD 0,31
                   → diferencia USD 0,53 por corrida
   Umbral          relectura 61 % > break-even 39,47 %        [ver el cálculo]
   Confianza       exacta · 14 de 14 corridas con atribución
   Sesgo           el cálculo asume 0 lecturas fuera de ventana → subestima el ahorro

   → Fix:  fijar `cache_ttl: 1h` en esta caja            [Aplicar]   [Descartar]
                                                          score v1
```

Eso es A1 (umbral algebraico citado, auditable), A2 (contrafactual simétrico, no «gastaste X»),
A3 (el sesgo declarado y **en contra** de la recomendación), A4 (un fix), A7 (versión del score).

⚠️ **`[Aplicar]` necesita tu decisión.** Escribir la config del arnés es tocar archivos de un
paquete de terceros — A8 pide backup + confirmación, y el guardrail vigente dice que el chat
embebido no toca el paquete propio. Tres salidas: **(a)** el botón solo copia el cambio al
portapapeles · **(b)** abre el chat con el cambio propuesto y que el flujo normal lo aplique ·
**(c)** lo aplica con diff + confirmación. **Recomiendo (b)** — reusa el camino ya construido y
firmado, y no inventa una segunda vía de escritura.

### 1.5 · Inspector: cuarta tab «Mejora»

Hoy `Resumen | Contenido | Corridas`. Se suma una, para la caja seleccionada:

- **desglose de tokens** — entrada · salida · cache lectura · cache escritura 5 m · 1 h ·
  razonamiento, con **«no aplica»** donde el runtime no tiene el concepto (jamás 0).
- **costo reportado vs. costo calculado** — lo que dijo el runtime contra lo que dice nuestro
  catálogo. Si divergen, la fila lo dice. *Es el test de paridad A6 corriendo a la vista.*
- **el join, a nivel nodo** — corridas de esta caja · cuántas se rechazaron en el gate · **cuánto
  costaron las rechazadas**. Esta fila sola es la mitad de la frase objetivo.
- **detectores** — los que aplican, y los que **no**, con el motivo: *«re-warm por TTL: no
  disponible, este arnés corrió fuera de ArnesIA»*.

### 1.6 · Los siete estados honestos (paneles del mockup)

Un mockup que solo dibuja el caso feliz no sirve para firmar. Se dibujan los 7:

1. **sin datos** — el arnés nunca corrió con telemetría. Qué hacer, no un tablero en cero.
2. **parcial** — «3 de 5 corridas sin atribución».
3. **S2 degradado** — corrió fuera de ArnesIA: B1 no disponible, marcado como tal.
4. **otro runtime** — sin costo del runtime; calculado con catálogo `v2026-07-20`.
5. **catálogo viejo** — «precios del release, sin refrescar desde el 20/07».
6. **atribución por huella** — identificado por `plugin_id_hash`, no por nombre.
7. **qué guardamos** — enlace visible a la política de datos (D15): nada de cuenta, nada de
   contenido, TTL y botón de borrado.

## Superficie 2 — la tarjeta del Portafolio ⚠️ *necesita tu decisión de alcance*

Sin esto la frase objetivo **no se puede decir**: «en este puesto» es comparación entre arneses, y
el Mapa mira uno solo. Propuesta mínima (una fila por arnés × puesto):

```
vitalia · Product Owner        USD 0,31/corrida   ▁▂▃▅▃  ⚠ 1 punto de mejora (USD 0,53/corrida)
harness · Dev                  USD 0,12/corrida   ▁▁▂▁▁  ✓ sin fugas detectadas
prenter · Dev                  sin dato — nunca corrió con telemetría
```

Es poco dibujo y es **el eje diferencial** (D9.8/H2: nadie correlaciona por unidad de trabajo).
Si preferís acotar el paquete al Mapa, sale — pero entonces el MVP no dice la frase completa, y
eso hay que escribirlo, no dejarlo implícito.

## Lo que este mockup NO dibuja (y por qué)

- **Capa Desempeño** — la señal ya llega (`duration_ms`, `hook_execution_complete`), falta el
  diseño de qué es «desempeño» a nivel Mapa. Fuera.
- **Los otros 7 detectores** — no pasan la regla A4 todavía. Se muestran como no medidos.
- **Un tablero de series temporales** — es lo que ya venden ccusage/Dynatrace. Competir ahí es
  perder (H1).

## Las tres decisiones que necesito de vos antes de dibujar

| # | decisión | mi recomendación |
|---|---|---|
| 1 | ¿el slot `Tokens` del conmutador pasa a llamarse «Mejora»? | **sí** — «Tokens» nombra el insumo, no el producto (D16.3) |
| 2 | ¿entra la tarjeta del Portafolio en este paquete? | **sí, mínima** — sin ella el MVP no dice la frase objetivo |
| 3 | ¿qué hace `[Aplicar]` en la tarjeta de mejora? | **(b) abrir el chat con el cambio propuesto** — reusa el camino firmado |
