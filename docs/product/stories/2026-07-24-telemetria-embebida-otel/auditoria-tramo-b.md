# Auditoría independiente — Tramo B (la UI de la capa «Mejora»)

> `tipo: auditoría` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26 · auditor
> independiente, **adversario**. Par de [`PARIDAD.md`](PARIDAD.md), [`design.md`](design.md),
> [`spec.md`](spec.md) y [`auditoria-tramo-a.md`](auditoria-tramo-a.md).
>
> **Esta hoja no firma nada y no toca código.** Reporta.

## Veredicto

**No firmar.** Los componentes están bien construidos y son honestos de a uno; **la composición no
lo es** — se reprodujeron **cuatro defectos críticos de la misma clase que D24**, más grave el peor
de ellos que el que motivó D24, y el candado que D24 dejó **no cubre ninguno de los cuatro**.

---

## 0 · Cómo se auditó (para que se pueda repetir)

El bundle apunta a `http://127.0.0.1:4200` fijo (`shared/api/client.ts:21`), así que **no** se tocó
el `~/.arnesia` real: se levantó un **proxy de fixtures de solo lectura** en `:4300` que reenvía al
daemon real **solo GET** (405 duro para POST/PUT/DELETE) y sustituye `/api/telemetria/*` y
`/api/sessions` por payloads con la forma exacta del wire de Go. Encima, el dev server de vite con
`VITE_ARNESIA_API` apuntando al proxy — **el mismo composition-root que corre en el binario**.

```bash
# 1 · proxy de fixtures (script versionado junto a esta hoja)
python3 docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/proxy-fixtures-tramo-b.py &
echo "rico" > escenario.txt          # rico | vacio | s2 | ventana | parcial | muchos | detalle-roto

# 2 · la app real contra el proxy
cd web && VITE_ARNESIA_API=http://127.0.0.1:4300 npx vite --port 5199

# 3 · navegador real → http://127.0.0.1:5199/ → tab «Mejora» del conmutador
```

El daemon real quedó intacto (se le leyó `/api/telemetria/salud`, `/api/harnesses`,
`/api/harnesses/vitalia/graph`; **cero escrituras**). Capturas en
[`verificacion-2026-07-26/evidencia/tramo-b*.png`](verificacion-2026-07-26/evidencia/).

---

# CONFIRMADO

## 🔴 CRÍTICOS

### C-1 · Encender la capa «Mejora» destruye el Mapa

**Qué está mal.** La lista de puntos de mejora es un hermano flex del canvas dentro de la misma
columna (`workspace-stage.tsx:462` canvas `min-h-0 flex-1` · `:538` lista sin altura acotada). La
lista no tiene scroll propio (`overflow-y: visible`), así que **le come la altura al canvas hasta
dejarlo en 0**.

**Cómo se reprodujo.** Medido en el DOM, viewport 1136×833:

| caso | alto del canvas | alto de la lista |
|---|---|---|
| capa `Estructura` | **739 px** | — |
| capa `Mejora`, **1** tarjeta | **257 px** (−65 %) | 369 px |
| capa `Mejora`, **4** tarjetas | **0 px** | 1 231 px |
| capa `Mejora`, 1 tarjeta, viewport 700×450 (≈200 % de zoom) | **0 px** | — |

```js
// consola del navegador, con la capa encendida
document.querySelector('.relative.min-h-0.flex-1').getBoundingClientRect().height  // → 0
```

Capturas: [`tramo-b2-canvas-colapsado-4-tarjetas.png`](verificacion-2026-07-26/evidencia/tramo-b2-canvas-colapsado-4-tarjetas.png)
y [`tramo-b7-mapa-zoom200.png`](verificacion-2026-07-26/evidencia/tramo-b7-mapa-zoom200.png):
la vista **Mapa** no muestra mapa. Ni un carril, ni una caja.

**Qué debería pasar.** La capa es un *superset* del Mapa (RF-245 · BR-M16): el canvas conserva su
altura y la lista scrollea dentro de sí misma o cuelga de un panel con altura propia. Cuatro
tarjetas no es un borde — el mockup firmado dibuja dos, y `MuchasTarjetas` existe como story.

**Dónde.** `web/src/pages/shell/ui/workspace-stage.tsx:462` y `:538`.

**Gravedad: crítica.** Rompe la superficie firmada del Slice anterior en el caso normal, y a 200 %
de zoom rompe con **una** tarjeta. Ninguna story podía verlo: `CapaMejoraSupersetGeografia` monta
el canvas solo, `MuchasTarjetas` monta la lista sola.

---

### C-2 · En `s2-instrumentado` la franja dice «nunca corrió» mientras el canvas cobra dinero — y el candado de D24 esconde el entregable

**Qué está mal.** Tres bloques de la misma pantalla afirman cosas incompatibles, y el fix de D24 es
uno de los causantes.

El wire produce `corridas: 0` fuera de S1 — no es una hipótesis, es aritmética de SQL:
`out.Corridas = COUNT(DISTINCT corrida_id)` y `corrida_id` es NULL cuando la corrida no la lanzó
ArnesIA (`internal/adapters/telemetria/store/consultas.go:95` · el propio store usa
`corrida_id IS NOT NULL` para *derivar* S1 en `:249`). La auditoría del Tramo A ya lo declaró
(§UX ítem 7) y `PARIDAD.md` §4 ítem 2 dice que ese caso «es la mayoría».

Con esa forma de wire —`corridas: 0`, pero `costo_reportado_micros`, `cobertura` y `/cajas`
poblados— la pantalla queda así:

| bloque | qué dice |
|---|---|
| franja | `— —` · **«Este arnés nunca corrió con telemetría.»** · «Abrí una sesión desde ArnesIA y la medición arranca sola.» |
| carril, 250 px abajo | `verificacion-de-marca · 1 · USD 1,08` |
| caja en el canvas | `USD 1,08 · 56 % · ⚠ re-warm TTL` |
| lista de puntos | **no se dibuja** |

**Cómo se reprodujo.** `escenario.txt = s2` (resumen `corridas:0`, `costo_reportado_micros:1920000`,
`cobertura 58/0/0/3`, `escenario:"s2-instrumentado"`; `/cajas` con `costo_micros:1080000`).
Captura: [`tramo-b6-s2-franja-nunca-corrio-canvas-con-plata.png`](verificacion-2026-07-26/evidencia/tramo-b6-s2-franja-nunca-corrio-canvas-con-plata.png).
Texto literal del DOM:

```
franja : ventana | 7 días | — — | Este arnés nunca corrió con telemetría. |
         Abrí una sesión desde ArnesIA y la medición arranca sola.
nodo   : CAJA | hipaa-check | USD | 1,08 | 56 % | ⚠ re-warm TTL
lista  : (ausente)
```

**Tres cosas mal, no una:**

1. `workspace-stage.tsx:302` — `setResumen(r && r.corridas > 0 ? r : null)` colapsa **todo** el
   estado del resumen a «nunca corrió» mirando un solo campo que en S2 es siempre 0. El costo, la
   cobertura, las sesiones y las cajas están ahí y se tiran.
2. La rama `escenario === "s2-instrumentado"` de la franja (`franja-mejora.tsx:204`), que existe
   justamente para este caso, **nunca se alcanza**: la rama `resumen == null` (`:143`) gana antes.
3. `hayDatosAtribuibles()` —el candado de D24— usa `resumen.corridas <= 0`
   (`model/capa-mejora.ts:37`), así que **esconde la lista de puntos de mejora**, que es el
   entregable del paquete entero, justo donde el backend sí los calculó.

**Qué debería pasar.** El estado 1 es «no hay medición», no «`corridas` vino en 0». La condición
tiene que derivarse de la señal que existe (costo, cobertura, cajas), igual que `hayDatosAtribuibles`
ya derivó de la cobertura en vez de la confianza. Y si el estado 1 se dispara, el canvas no puede
seguir pintando dinero.

**Dónde.** `workspace-stage.tsx:302` · `widgets/map-canvas/model/capa-mejora.ts:36-40` ·
`widgets/map-canvas/ui/franja-mejora.tsx:143`.

**Gravedad: crítica.** Es **el mismo defecto que D24**, movido un bloque más allá, en el escenario
que la propia PARIDAD declara mayoritario. El candado `capa-mejora-coherencia.stories.tsx` no lo ve
porque **solo monta `FranjaMejora` + `PuntosMejoraList`** (verificado: el archivo importa esos dos
y nada más) — el canvas se alimenta de otra consulta y quedó fuera del espejo.

---

### C-3 · Un `GET` de detalle que falla se pinta como dato, y contradice al canvas de al lado

**Qué está mal.** `workspace-stage.tsx:335-339` traga el error del detalle de caja
(`.catch(() => setDetalle(null))`) y **nunca setea `estado="error"`**. `InspectorMejora` tiene la
prop `estado` con su rama de error (`inspector-mejora.tsx:101`) y **la página no se la pasa nunca**.
Los defaults hacen el resto: `corridas: 0`, seis buckets `null`, `paridad` de relleno con
`catalogo_sin_construir: true`.

**Cómo se reprodujo.** `escenario.txt = detalle-roto` (500 en
`/api/telemetria/arneses/{id}/cajas/{cajaId}`, todo lo demás 200). Con la caja `hipaa-check`
seleccionada, la 4ª tab dice, literal:

```
TOKENS · 7 DÍAS (0 CORRIDAS)
entrada                 no aplica en este runtime
salida                  no aplica en este runtime
cache · lectura         no aplica en este runtime
cache · escritura 5 m   no aplica en este runtime
cache · escritura 1 h   no aplica en este runtime
razonamiento            no aplica en este runtime
«No aplica» no es 0. Un cero donde el concepto no existe sería mentira.

COSTO — REPORTADO VS. CALCULADO
runtime            este runtime no reporta costo — calculado con el catálogo v
nuestro catálogo   catálogo sin construir
```

Al mismo tiempo, **en la misma pantalla**: el nodo dice `USD 1,08 · 56 %`, la tarjeta de abajo dice
`exacta · 14 de 14 corridas con atribución`, y `/api/telemetria/salud` del daemon real devuelve
`catalogo: {version: "2026-07-26", modelos: 694}`.
Captura: [`tramo-b3-inspector-error-como-dato.png`](verificacion-2026-07-26/evidencia/tramo-b3-inspector-error-como-dato.png).

**Ocho afirmaciones falsas de un fetch que falló**, y la nota de honestidad («No aplica no es 0»)
desplegada **en defensa** de la mentira. Es la inversión exacta de RF-260/RF-286: una ausencia
etiquetada como «no aplica».

**Bonus del mismo bloque:** la caja de paridad toma la clase `difieren` cuando `divergencia === null`
(`inspector-mejora.tsx:188` — `divergencia === 0 ? "coinciden" : "difieren"`), así que se pinta
**ámbar con borde ámbar y sin una sola palabra que lo explique** (el veredicto está guardado por
`divergencia !== null`). Verificado en vivo: `class="mej-paridad difieren"`,
`background: rgba(201,106,46,.13)`, `.mej-veredicto` ausente. El encabezado del propio archivo dice
«el veredicto es TEXTO, no un tono».

Y `catálogo v` queda colgando: `paridad.catalogo_version` es `undefined` en el objeto de relleno
(`workspace-stage.tsx:508-514`), y React no imprime nada.

**Qué debería pasar.** Un fetch fallido es estado de transporte, no dato:
`estado="error"` + el motivo real + `Reintentar` (la rama ya existe y está storiada como
`ErrorDeConsulta`). Y `detalle == null` ≠ «este runtime no tiene estos conceptos».

**Dónde.** `workspace-stage.tsx:322-343` y `:498-529` · `inspector-mejora.tsx:158-171, 188, 191-198`.

**Gravedad: crítica.** El inspector existe para hacer **auditable** el número; en este estado
inventa seis hechos sobre el runtime y uno sobre el catálogo. Y sucede también con la caja de un
arnés que nunca corrió (`escenario = vacio`, respuesta 200 con `{turnos_totales: 0}`): la lista se
esconde por D24 y **el inspector dice «no aplica en este runtime» seis veces** — D24 arreglado en un
bloque, intacto en el de al lado.

---

### C-4 · El ✓ «sin fugas detectadas» del Portafolio miente — D24.4 se aplicó a un componente que la app no monta

**Qué está mal.** `PARIDAD.md` §8 fix 4 declara: *«El ✓ `sin fugas detectadas` del Portafolio sale
de `puntos_de_mejora === 0` **como dato del wire**, no de la ausencia del campo `punto`»*. Ese fix
está en `widgets/portafolio/ui/tabla-mejora-portafolio.tsx:165-179`.

**`TablaMejoraPortafolio` no se renderiza en ninguna parte de la aplicación.** Fuera de sus stories,
el único lugar donde aparece es el `export` de `widgets/portafolio/index.ts:17`. La superficie que
el operador ve es `portafolio-view.tsx` → `portafolio-list.tsx`, que tiene una **segunda
implementación de las mismas tres celdas** (`portafolio-view.tsx:170-192`) — y esa **no tiene el
guard**:

```tsx
// portafolio-view.tsx:178-188  (la que sí se renderiza)
f.costo_por_corrida === null ? <chip "nunca corrió con telemetría">
: f.punto              ? <chip con el punto>
:                        <chip ✓ "sin fugas detectadas">     // ← sin mirar puntos_de_mejora
```

**Cómo se reprodujo.** Fila del wire `{arnes_id:"vitalia", costo_por_corrida:31000,
puntos_de_mejora:3, punto:undefined}` → la fila pinta **`✓ sin fugas detectadas`**, en verde, como
el elemento más saliente del renglón.
Captura: [`tramo-b4-portafolio-check-mentiroso.png`](verificacion-2026-07-26/evidencia/tramo-b4-portafolio-check-mentiroso.png).

**Qué debería pasar.** Lo que dice D24.4 y lo que la tabla muerta ya hace: con
`puntos_de_mejora > 0` y sin `punto`, decir cuántos hay. La rama existe, escrita, testeada y
**inalcanzable**.

**Dónde.** `web/src/pages/shell/ui/portafolio-view.tsx:170-192` (real) vs
`web/src/widgets/portafolio/ui/tabla-mejora-portafolio.tsx:165-179` (muerto).

**Gravedad: crítica.** Es literalmente el defecto que el paquete dice haber corregido, en la
superficie que el operador abre primero, mintiendo *«acá no hay nada que mirar»* — la dirección que
D24 nombra como la más cara. Y arrastra a todo T36 (ver A-6).

---

## 🟠 ALTAS

### A-1 · `Descartar` y `Proponerlo en el chat` no hacen nada

- `workspace-stage.tsx:553-554`: los dos handlers son `() => setNonce(n => n + 1)`. Solo recargan.
- **No existe endpoint de descarte**: nada en `internal/adapters/transport/http/router.go`, nada en
  `shared/api/client.ts`. El descarte no se persiste.
- El refetch **desmonta y remonta** la tarjeta, así que el `descartado` local se resetea y el
  anuncio `aria-live` queda **vacío**. Medido: 1 tarjeta antes, 1 después, `.mej-live` = `""`.
- El anuncio prometía *«Se puede volver a mostrar desde la tab Mejora de esa caja»*. **Esa
  afordancia no existe**: `grep -rn "volver a mostrar"` devuelve solo el literal que la promete.
- `Proponerlo en el chat` no abre ningún chat: sin dock, sin `<textarea>`, nada — mientras la nota
  al pie, 40 px más abajo, dice *«abre el chat con el cambio propuesto»*.

**Repro.** Escenario `rico`, clic en cada botón, medir antes/después.
**Dónde.** `workspace-stage.tsx:553-554` · `punto-mejora-card.tsx:145-172`.
**Por qué duele.** RF-255 y RF-256 son las dos únicas acciones de la superficie. Las stories
`DescartarLlamaHandler` y `ProponerAbreChatNoEscribe` assertan que *se llama al handler* — y el
handler que la página conecta no hace lo que la pantalla promete.

### A-2 · `GET /api/telemetria/salud` está construido y **nunca se consume**

`api.telemetriaSalud` existe (`client.ts:341`) y el daemon real responde 200 con
`retencion_dias`, `retencion_propuesta`, `forward`, `forward_destino`, `almacen_disponible`,
`almacen_motivo`, `catalogo`, `ultima_recepcion`. **Ninguna página lo llama.** Consecuencias
verificadas contra un `/salud` que devolvía `{retencion_dias: 400, retencion_propuesta: false,
forward: true, forward_destino: "https://otlp.datadoghq.com", almacen_disponible: false}`:

| lo que /salud dice | lo que la pantalla dice |
|---|---|
| retención **400 días, firmada** | «Retención **90 días** *(propuesto)*» en la franja **y** «Se borra solo a los **90 días**. Valor propuesto, sin firmar.» en el diálogo |
| **reenvío externo ENCENDIDO** → `otlp.datadoghq.com` | **nada**. El chip `.fm-forward` no está en el DOM |
| almacén **caído** (`disco lleno`) | nada |

`PARIDAD.md` §7 ítem 4 afirma *«La UI lo lee de la config y lo rotula como tal; el número lo pone el
operador»*. **Es falso**: `retencionDias={90}` está hardcodeado en `workspace-stage.tsx:455` y
`:564`.
**Gravedad alta**, y el chip de reenvío es lo más grave del bloque: D13 · H-7 dicen *«un estado
peligroso no se esconde a la derecha»*. Hoy no se esconde a la derecha — **no se dibuja**. Una
promesa de privacidad que la pantalla repite («Nada de tu cuenta. Nada de la conversación.») sin
poder saber si los datos se están reenviando afuera.

### A-3 · El estado 1b no existe en la app: «nunca corrió» sobre un arnés con 340 corridas

`ultimaCorridaFuera` **no se pasa desde ningún lado** (`grep` sobre todo `src/` sin stories: solo
aparece dentro de `franja-mejora.tsx`). Con la ventana en 7 días sobre un arnés cuyo historial está
fuera de esa ventana, la franja dice:

```
7 días : — —  Este arnés nunca corrió con telemetría.
              Abrí una sesión desde ArnesIA y la medición arranca sola.
todo   : USD 1,92 — de 340 corridas, 58 con atribución · 88 sesiones · 4 cajas
```

**Repro:** `escenario.txt = ventana` (el fixture devuelve `corridas: 0` si la query trae `desde`, y
340 si no). Cambiar el `<select>` de ventana.

Dos daños: la afirmación es falsa, y el remedio que ofrece es el equivocado («abrí una sesión»
cuando lo que hace falta es ampliar la ventana). El botón `Ver todo` —que la franja tiene escrito en
`:159`— no aparece nunca. `Estado1bSinCorridasEnVentana` (H-9 · T-15) pasa en aislamiento.

### A-4 · «Se borran 61 corridas» — el conteo es de la ventana; el borrado es de todo

`workspace-stage.tsx:566` pasa `corridasPorBorrar={resumen?.corridas ?? 0}`, que es el resumen **de
la ventana activa**. `api.telemetriaBorrarArnes(clave)` es `DELETE /api/telemetria/arneses/{clave}`
**sin ventana** (`client.ts:359` · `router.go:48`). Con la ventana en «7 días» sobre dos años de
historial, la confirmación subdeclara la destrucción.

**Repro:** diálogo abierto con `corridas: 61` en 7 d → *«Se borran 61 corridas medidas y los puntos
de mejora que salieron de ellas. No se puede deshacer.»* Cambiando la ventana a «todo» el número
cambia; la llamada, no.

**Gravedad alta:** es la única acción irreversible de la superficie y su alcance declarado no es su
alcance real.

### A-5 · El Portafolio dice «nunca corrió con telemetría» sobre un arnés con 47 corridas

Tanto la superficie real (`portafolio-view.tsx:178`) como la tabla muerta
(`tabla-mejora-portafolio.tsx:113`) derivan «nunca corrió» de **`costo_por_corrida === null`**.
`FilaPortafolio.corridas` está en el payload y **no se mira**. Con
`{corridas: 47, costo_por_corrida: null}` —que es exactamente lo que el backend produce fuera de S1
(PARIDAD §4 ítem 2)— la fila afirma que nunca corrió. El `aria-label` de la celda de tendencia
repite la mentira: `"sin tendencia: nunca corrió con telemetría"`.

**Repro:** fila `harness` del fixture. Verificado en el DOM y en
[`tramo-b4-portafolio-check-mentiroso.png`](verificacion-2026-07-26/evidencia/tramo-b4-portafolio-check-mentiroso.png).

### A-6 · T36 completo está en un componente que la app no monta; la superficie real perdió tres garantías

`TablaMejoraPortafolio` es código muerto (ver C-4). La lista que sí se renderiza **no tiene**:

| lo que la tabla muerta garantiza | en la superficie real |
|---|---|
| pie **H-12** «Costo estimado por el runtime, no es facturación» | **ausente** (`.pf-mej-pie` no está en el DOM) |
| `MarcaConfianza` por fila (RF-242 · H-12) | **ausente** — 0 nodos de confianza en las filas |
| encabezado `USD/corrida` (el prefijo «vive en el encabezado», dice el propio CSS) | **no hay encabezado**: la celda muestra `0,03` pelado |
| separador «Sin datos de telemetría» + orden por costo | ausentes |
| H-3, el puente Portafolio → Mapa (`onAbrirEnMapa`) | ausente |

Resultado en pantalla: un **`0,03` sin moneda, sin unidad, sin período y sin disclaimer**, entre un
chip `origen?` y un ✓ verde. H-12 dice, textual, *«sin este pie, un USD pelado en una tabla se lee
como facturación»*. Es exactamente lo que quedó.

### A-7 · La franja mezcla dos unidades en un denominador que promete cerrar

`franja-mejora.tsx:186` imprime `de ${resumen.corridas} corridas, ${exacta+por_hash+por_proceso} con
atribución`, y `BarraCobertura` imprime `— sobre ${suma de los 4 cubos} corridas`. En el backend
esas cifras **no son la misma magnitud**:

```sql
-- consultas.go:95  (resumen.corridas)
COUNT(DISTINCT corrida_id)  FROM evento WHERE … AND atribucion <> 'sin-dato'
                                              AND tipo_evento <> 'metrica'
-- consultas.go:143 (cobertura.*)
COUNT(DISTINCT COALESCE(turno_id,'sin-turno:'||id)) FROM evento WHERE …   -- sin ambos filtros
GROUP BY atribucion
```

El numerador cuenta **turnos** (de toda la población, incluidos los eventos métrica), el denominador
cuenta **corridas atribuidas**. Nada los ata. Cuando divergen, la franja dice cosas como
`de 340 corridas, 58 con atribución` y dos centímetros abajo `sobre 61 corridas` — **reproducido en
vivo**, texto del DOM en A-3. Además, `coberturaEsParcial()` divide sobre el total de la cobertura,
así que una divergencia grande **desactiva** el aviso de cobertura parcial en vez de dispararlo.

Ninguna story lo puede ver: `PARIDAD.md` D-9 y D-10 documentan explícitamente que las fixtures se
eligieron **para que cierren** (44/9/5/3 sobre 61) y que la entity usa otro juego. Se testeó que el
copy sabe formatear; no que las dos cifras vengan del mismo lugar.

**Gravedad alta**, y es el corazón del RF-235 («total + denominador que cierra»).

---

## 🟡 MEDIAS

### M-1 · Nueve estados honestos storiados e inalcanzables

Verificado con `grep` sobre todo `src/` excluyendo `*.stories.tsx`: la prop se define, se usa dentro
del componente, y **ningún composition-root la pasa**.

| estado | story que `PARIDAD.md` §1/§7 acredita | ¿lo pasa la app? |
|---|---|---|
| 1b · sin corridas en la ventana | `Estado1bSinCorridasEnVentana` | **no** (A-3) |
| 4 · otro runtime / costo del catálogo | `Estado4OtroRuntime` | **no** |
| 5 · catálogo viejo | `Estado5CatalogoViejo` | **no** |
| runtime no soportado | `RuntimeNoSoportado` | **no** |
| B3 · daemon caído, cifras viejas | `DaemonCaido` | **no** (el estado de la página ni siquiera tipa `"daemon-caido"`, `:110`) |
| chip de reenvío externo | `ReenvioEncendido` | **no** (A-2) |
| retención desde la config | `RetencionDesdeConfig` | **no** (A-2) |
| proponer deshabilitado fuera de alcance | `ProponerDeshabilitadoFueraDeAlcance` | **no** |
| H-3 · puente Portafolio → Mapa | `AbreElMapaDeEsaInstalacion` | **no** (A-6) |

La matriz de `PARIDAD.md` §1 mapea «qué dibuja el mockup → componente → story = test». El gate
humano recorre esa matriz y encuentra una story verde por cada fila. **La matriz no distingue
“implementado” de “alcanzable”**, y nueve filas están del lado equivocado.

El guardrail de alcance del chat (memoria `hs-arneses-usuario-vs-arneses-arnesia`, CH-D6) entra por
esta vía: `proponerDeshabilitado` nunca se pasa, así que el botón está habilitado también sobre el
arnés propio.

### M-2 · Un error de consulta deja el dinero viejo en el canvas

En el `.catch` del efecto de la capa (`workspace-stage.tsx:310-314`) se setea `mejEstado="error"`
pero **no se limpian `cajas` ni `puntos`**. Observado en vivo durante la auditoría: la franja decía
*«No se pudo leer la telemetría — Failed to fetch»* y **el nodo seguía mostrando `USD 1,08 · 56 % ·
⚠ re-warm TTL`**. Una cifra que el sistema acaba de declarar no-leíble sigue pintada como dato
vigente. Es el estado B3 (`DaemonCaido`) sin su rótulo — el rótulo existe y no se usa (M-1).

### M-3 · «pocas corridas para una tendencia» nombra una causa que no es la causa

`sparkline.tsx:30` devuelve ese literal cuando `puntos` es `undefined`. El wire **nunca** manda la
serie (PARIDAD §4 ítem 1). Sobre una fila con `corridas: 61` la pantalla dice «pocas corridas», que
es falso: la verdad es «todavía no calculamos la serie». `PARIDAD.md` §4 lo declara como «su
ausencia se DICE… es honesto» — **el impacto declarado no es el real**: no es una ausencia dicha, es
una causa inventada, en la única columna cuyo propósito es afirmar una dirección.

### M-4 · `por-proceso` avisa que puede mezclar arneses **solo por `title`**

`marca-confianza.tsx:31` pone la diferencia material en un atributo `title` sobre un `<span>` sin
`tabindex`. El texto visible es «por huella» / «por proceso»; la advertencia *«Si ahí corre más de
un arnés, este número los mezcla»* **no llega por teclado, no llega por touch y no llega en modo
navegación de un lector de pantalla**. El encabezado del propio archivo afirma que «el texto es el
portador de la diferencia» — el texto visible no la porta.
Y en el Portafolio la marca directamente no se renderiza (A-6): la fila `vitalia` con
`confianza: "por-proceso"` no lleva ninguna señal.

### M-5 · El Portafolio desborda en horizontal y recorta los chips

Con las tres celdas de mejora, el contenedor del Portafolio pasa a scroll horizontal
(`scrollWidth − clientWidth = 21 px` a 1136 px de ancho) y los chips escapan de su propia tarjeta.
A 900 px de ancho el chip «✓ sin fugas detect…» y «nunca corrió con…» quedan **cortados por el
viewport** y tapan el dot de salud, que es el ancla visual que la PARIDAD del Slice 1 firmó.
Captura: [`tramo-b5-portafolio-900px.png`](verificacion-2026-07-26/evidencia/tramo-b5-portafolio-900px.png).

### M-6 · Tipografía de 9 px en las etiquetas que hacen el argumento

`.mej-dl dt` (CONTRAFACTUAL · UMBRAL · CONFIANZA · SESGO) y `.mej-score` renderizan a **9 px**.
Contraste medido 7,2:1 — el problema no es el color, es el tamaño. Son los rótulos de la anatomía
que el design.md llama «el argumento» de la tarjeta.

### M-7 · `catálogo v` colgante

`inspector-mejora.tsx:193-194` imprime `catálogo v{paridad.catalogo_version}` sin comprobar que
exista. Con el objeto de relleno de `workspace-stage.tsx:508-514` (que no lleva el campo) la
pantalla muestra literalmente `calculado con el catálogo v`.

---

## 🔵 BAJAS

- **B-1.** `franja-mejora.tsx:157`: `Sin corridas en los últimos ${ventana === "30d" ? "30" : "7"}
  días` — con `ventana === "todo"` imprime «los últimos 7 días». Hoy es inalcanzable (A-3), pero es
  la bomba que queda armada para cuando se cablee el 1b.
- **B-2.** En estado 2 (cobertura parcial) el denominador **pierde** las sesiones y las cajas: la
  línea `de N corridas, M con atribución · S sesiones · C cajas` se reemplaza por
  `${sin_dato} corridas quedaron sin atribución` (`franja-mejora.tsx:182-188`). El estado que más
  necesita el denominador completo es el que lo recorta.
- **B-3.** `BarraCobertura` duplica sus cifras: el mismo texto en el `aria-label` del `role="img"`
  y en el `<span>` visible de al lado. Un lector de pantalla las lee dos veces.
- **B-4.** Las tabs `Desempeño` y `Proceso` están `disabled`, así que salen del orden de foco; sus
  motivos honestos sí están en el DOM como texto, pero el elemento que los explica no es
  alcanzable por teclado (RF-233 · RF-276 · `MotivoAccesible`).

---

# SOSPECHAS (no verificadas)

| # | qué | por qué no se cerró |
|---|---|---|
| S-1 | **Cifras del arnés anterior sobre el grafo nuevo.** Al cambiar `viewedId` con la capa encendida, el efecto del grafo y el de la telemetría corren en paralelo y **`cajas`/`totalesPorFase` no se limpian** (`workspace-stage.tsx:283-318`). Si el grafo resuelve primero y algún `caja_id` coincide entre arneses, el canvas pinta plata ajena | No se pudo forzar la carrera con dos arneses de ids de caja colisionantes en el tiempo disponible. El código no tiene guard |
| S-2 | **`DELETE /api/telemetria/arneses/{clave}` recibe `viewedId`** (`vitalia`), mientras el Portafolio identifica por `clave` (`sin-home~vitalia~vitalia`). Si el store indexa por clave, el borrado podría no borrar nada y el diálogo declarar éxito igual (`.finally` cierra y recarga sin mirar el resultado, `:567-575`) | Es un contrato de backend; no se probó contra el daemon real por la regla de solo lectura |
| S-3 | La magnitud real de A-7 en producción (¿cuántos turnos por corrida?) depende de los datos. La **divergencia estructural** está probada por código; el tamaño del error visible, no | Requiere una base con corridas reales |

---

# UX — lo que un usuario real va a tropezar, por cuánto duele

1. **Abre el Mapa, enciende «Mejora» y el mapa desaparece.** (C-1) Con dos o tres hallazgos —el caso
   que el mockup dibuja— la vista Mapa deja de tener mapa. La primera reacción es «se rompió».
2. **La pantalla se contradice consigo misma a dos centímetros de distancia.** (C-2, C-3, A-7) Un
   producto cuyo argumento entero es la honestidad de la cifra pierde el argumento cuando el usuario
   ve dos números incompatibles en una sola captura de pantalla. No hace falta que sepa cuál es el
   correcto: le alcanza con saber que uno miente.
3. **La frase objetivo no se puede leer de la pantalla.** *«Este arnés, en este puesto, quema $X y el
   60 % se va en la caja Y, que falla el gate 3 de cada 4 veces»* — el puesto está en el Portafolio,
   el $X en la franja, el 60 % en el nodo, y el «3 de cada 4» **no existe** (T22 abierto). El Mapa
   no muestra el puesto en ningún lado, y el Portafolio no muestra la caja. **Hay que cruzar dos
   vistas y una no tiene la mitad de la oración.** La tarjeta de mejora es lo más cerca que se
   llega, y le falta el puesto.
4. **Los dos botones de la tarjeta no hacen nada.** (A-1) `Descartar` deja la tarjeta ahí; el
   usuario lo aprieta de nuevo. `Proponerlo en el chat` no abre el chat, contra lo que dice la nota
   que está justo debajo. Es la clase de silencio que se lee como bug, no como decisión.
5. **En el Portafolio el número no dice qué es.** (A-6) `0,03` pelado, sin moneda, sin `/corrida`,
   sin «estimado». Con el pie ausente, un `USD` en una tabla se lee como factura — que es
   exactamente lo que H-12 dice que no puede pasar.
6. **El ✓ verde es la señal más fuerte del renglón y es la falsa.** (C-4) El ojo va al check antes
   que a nada; dice «no hay nada que mirar» sobre un arnés con tres hallazgos.
7. **«Nunca corrió» cuando lo que pasa es que la ventana es corta.** (A-3) El usuario cree que la
   telemetría no funciona y va a buscar el problema donde no está. No hay ninguna pista de que
   cambiar el `<select>` de arriba resuelve.
8. **Los estados de error dicen qué pasó, no qué hacer** — cuando aparecen. `ErrorBody` lleva
   `Reintentar`, que está bien; pero los tres estados de degradación *que sí explican por qué*
   (catálogo viejo, otro runtime, runtime no soportado) están escritos y no se alcanzan (M-1). Lo
   que el usuario ve en su lugar es el estado limpio, que no explica nada.
9. **A 200 % de zoom la franja ocupa la pantalla entera.** (C-1) Seis líneas envueltas y cero canvas.
10. **9 px** en los rótulos que estructuran el argumento de la tarjeta. (M-6)

---

# Lo revisado que está bien

- **Contraste, medido en los dos temas** sobre la superficie compuesta: 17 selectores nuevos, todos
  por encima de 4,5:1 en claro **y** en oscuro (`.fm-disclaimer` 15,5 / 15,3 · `.fm-mut` 7,4 / 6,9 ·
  `.mej-btn-primary` 15,6 / 15,4). Las decisiones D21 y D-7 se sostienen fuera del gate de stories.
  El tema oscuro no destapó nada de color pese a tener solo 3 stories espejo.
- **`entities/telemetria` es correcta y disciplinada.** `usd()` sube decimales antes que redondear a
  cero, `pct()`/`entero()` devuelven `null` en vez de `0`, la cobertura no dibuja categorías vacías,
  `MOTIVO_SIN_DATO` es una sola fuente. El nivel de los componentes aislados es alto; **el problema
  está exclusivamente en quien los compone**.
- **El candado D24 funciona y es honesto en su alcance.** Corrido en aislamiento: 6/6 verde en 4 s.
  El control positivo descrito en §8 es un buen patrón.
- **La 4ª tab del inspector, con datos completos, es la mejor parte del tramo**: la tabla de buckets,
  el «no aplica» con `colspan`, la paridad con los dos números, el join con motivos por fila y los
  detectores con `no medido todavía` — todo se lee y todo dice de dónde sale.
- **Las secciones no se caen cuando les falta un dato**: los seis motivos por clase de nodo, el
  «sin señal de gate», el «El detalle de esta caja no trajo el estado de los detectores».
- El diálogo de política tiene `role="dialog"` + `aria-modal`, foco al abrir y confirmación en dos
  pasos.
- La tabla de buckets y el scroll del Portafolio llevan `tabIndex`+`role`+`aria-label`.
- **Los formatos son consistentes**: coma decimal, espacio fino de millar, `USD` atenuado aparte.

---

# Lo que NO se pudo auditar, y por qué

| # | qué | por qué |
|---|---|---|
| N-1 | **El binario instalado v0.2.23** | El bundle fija `BASE = http://127.0.0.1:4200` y el único daemon en ese puerto es el del `~/.arnesia` real del operador, que la consigna declara de solo lectura. Todo lo de arriba se verificó sobre el **mismo código fuente** servido por vite. La diferencia entre dev server y bundle es el minificador, no la composición |
| N-2 | **El wire real con datos** | El `~/.arnesia` del operador tiene `corridas: 0` en las cuatro filas y `ultima_recepcion: null`: nunca se ingirió telemetría. Todos los estados con dato se ejercitaron con fixtures **de la forma exacta del wire**, leída de `internal/domain/telemetria_vistas.go` y de las consultas SQL. La forma está verificada; los valores reales, no |
| N-3 | **axe** | `axe-core` no está en `node_modules` (solo lo trae el runner de stories). El contraste se midió a mano con la fórmula WCAG sobre `getComputedStyle` compuesto; el resto de reglas de axe sobre la **página compuesta** queda sin correr |
| N-4 | **La suite completa** | No se re-corrieron los 483 tests: `PARIDAD.md` §5 V-7 ya declara flakiness de 2 de cada 7 corridas por el runner. Se corrió el candado de coherencia (6/6) y se leyeron los archivos de stories relevantes |
| N-5 | **V-5, el contrato de `PuntoDeMejora`** | Confirmo el hueco tal como `PARIDAD.md` lo declara: el FE tipa `contrafactual`/`patron`/`calculo`/`fix_codigo` como prosa y el Go entrega `ContrafactualMicros`/`DiferenciaMicros`. Mis fixtures usan el tipo del FE, así que **la tarjeta que se ve en las capturas es la del contrato del FE, no la del wire**. Sigue siendo el hueco más grande y bloquea el E2E real de la tarjeta |
| N-6 | **Shell Tauri, Windows, macOS** | Sin runner |
| N-7 | **`fitView` con los nodos más altos** (V-3) | Tapado por C-1: con el canvas en 0 px no hay layout que medir |

---

# Sobre lo que `PARIDAD.md` declara abierto — ¿el impacto es el real?

| declaración | veredicto |
|---|---|
| §4.1 · sin serie ⇒ «pocas corridas para una tendencia», *«es honesto»* | **Peor.** Nombra una causa falsa (M-3) |
| §4.2 · `costo_por_corrida` null fuera de S1 ⇒ «la celda dirá sin dato» | **Peor.** Dice «nunca corrió con telemetría» sobre 47 corridas (A-5), y en el Mapa colapsa la franja entera (C-2) |
| §4.9 · sin costo por bucket, *«la columna USD viaja null — no aplica»* | **Peor y distinto.** Viaja `null` y se renderiza `sin dato` en las 5 filas **con un total de `USD 1,08` debajo**: la sección que existe para auditar el número muestra un número sin origen |
| §5 V-2 · *«nadie miró la pantalla»* | **Exacto, y es la causa raíz de los cuatro críticos.** Los cuatro se ven a simple vista y ninguno es detectable desde el DOM de un componente aislado |
| §7.4 · *«El TTL … la UI lo lee de la config»* | **Falso.** Hardcodeado en dos lugares (A-2) |
| §8 · *«el candado cubre esta composición… una composición futura que nadie espeje vuelve a quedar ciega»* | **Subdeclarado.** No es una composición futura: **tres composiciones existentes ya están ciegas hoy** — franja↔canvas (C-2), canvas↔inspector (C-3) y la fila del Portafolio (C-4) |

---

# Objeciones a lo firmado (no relitigo, dejo constancia)

1. **La matriz de `PARIDAD.md` §1 no puede ser el instrumento del gate tal como está.** Acredita
   «componente + story verde» y el gate lo lee como «la superficie hace esto». Nueve filas tienen
   story verde y no se alcanzan desde la app (M-1). Sugiero una columna más: *quién le pasa la prop*.
2. **La regla que D24 dejó escrita es correcta pero está mal acotada.** Dice «toda superficie que
   componga dos bloques que hablan del mismo hecho lleva una story de coherencia sobre el DOM
   completo», y el candado que la implementa monta dos de los cinco bloques. Los bloques que hablan
   del mismo hecho en esta capa son **cinco**: franja, carril, nodo, tarjeta e inspector — más la
   fila del Portafolio, que habla del mismo hecho en otra vista.
3. **`0 stories en `pages/`» está declarado como deuda y es la causa de los cuatro críticos.** No es
   una deuda de cobertura: es el único lugar donde vive el defecto que este paquete existe para no
   cometer.
4. **D18 («entities presenta; no calcula») está bien y no está en discusión** — pero su corolario no
   se cumplió: si el FE no calcula, alguien tiene que pasarle el dato calculado, y `salud`,
   `ultima_corrida`, `forward` y el costo por bucket no se los pasa nadie. La prohibición de
   calcular no exime de cablear.
5. **El bloque D-10 de `PARIDAD.md`** —dos fixtures distintos para la misma barra, uno elegido «para
   que cierre con el denominador»— es, leído hoy, la firma del defecto A-7: se eligió una fixture
   coherente en vez de probar que las dos cifras vengan de la misma fuente.
