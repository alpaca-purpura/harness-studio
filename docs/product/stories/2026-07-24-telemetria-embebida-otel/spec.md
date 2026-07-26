# Spec funcional · Capa «Mejora» del Mapa + tarjeta del Portafolio (RF-232…RF-286)

> `tipo: spec` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> Etapa 3 del flujo (METODOLOGIA §10), habilitada por el **🧑‍⚖️ GATE del bloque
> D9·D11·D13·D14·D15·D16** firmado el 2026-07-26 (`decisiones.md`, línea 613) — que en su última
> línea dice, textual: *«queda autorizado el spec»*.
> Numeración: continúa desde **RF-231** (`stories/2026-07-26-identidad-de-build/spec.md`), el último
> en uso en el repo.
>
> **El QUÉ, no el cómo.** El UI al pixel, los tokens y los estados por componente van en
> [`design.md`](./design.md). La arquitectura del receptor, el evento canónico y el catálogo de
> precios ya están resueltos en [`decisiones.md`](./decisiones.md) D9/D14/D16 y
> [`arquitectura-telemetria.md`](arquitectura-telemetria.md); este documento **no los relitiga**.
>
> **Trazabilidad:** cada RF cita `mockup-capa-mejora.html:<línea>` — línea real, verificada contra
> el archivo. Los RF sin superficie (§H) lo dicen y citan su decisión firmada.

## Estado de este documento — LEER ANTES DE CONSTRUIR

| | |
|---|---|
| ✅ **Decisiones firmadas que este spec ejecuta** | D9 · D11 · D13 · D14 · D15 · D16 (gate del 2026-07-26) · **D17.1/D17.2/D17.3** (firmadas en el mismo hilo) |
| ⚠️ **El mockup NO está firmado** | `mockup-capa-mejora.html` va por **iteración 1**. Este spec se escribe contra él porque el gate de decisiones lo autorizó, pero **los RF 🎨 no se construyen hasta que el mockup tenga su 🧑‍⚖️** |
| 🔴 **Desviación de un baseline firmado** | **RF-232** renombra el slot `Tokens`→`Mejora` del conmutador de capas (`mockups/INDEX.md` regla 3 · D17.1). No se quita ningún slot: cambia **una etiqueta ya firmada en PARIDAD**. Va al gate declarada, no colada |
| 📌 **Deuda visible, no bloqueante** | **D9.9** (parser propio vs. shell-out a `ccusage`) sigue abierta — posterior al MVP. **G4** (cert de firma de código) no bloquea Linux |
| 🆕 **Anexo de hooks (2026-07-26, misma fecha)** | [`verificacion-2026-07-26/ANEXO-hooks.md`](verificacion-2026-07-26/ANEXO-hooks.md) llegó durante la escritura de este spec y cambió **dos** cosas: (1) el payload del hook trae **contenido en claro** ⇒ la allowlist ata a **los dos caminos de ingesta** (RF-282, BR-M14); (2) la atribución tiene **cuatro** niveles y `por-proceso` es real vía `cwd` ⇒ cuatro casos distinguibles en la UI (RF-242, RF-237). No cambia la forma del mockup; cambia lo que promete y cuántos estados dibuja |
| 🕳 **14 huecos del mockup** | §I al final. Ninguno se inventa acá: se listan con la resolución propuesta para la iteración 2 |
| ⚔️ **10 contradicciones mockup ↔ decisiones** | §J al final, con el veredicto de cuál gana |

**Orden obligatorio de construcción:** los RF sin superficie (§H, RF-282…RF-286) **antes** que
cualquier píxel. Son la ingesta, la allowlist de PII y la retención: si la UI se construye primero,
se persiste PII antes de tener dónde borrarla.

---

## 0 · Alcance y no-alcance

**Dentro — dos superficies (D17.2):**

| # | Superficie | Estado hoy |
|---|---|---|
| S-MAPA | Capa **Mejora** del Mapa: barra + franja de contexto + canvas + tarjetas de punto de mejora + 4ª tab del inspector | slot `tokens` existe `disabled` (`widgets/map-canvas/model/layers.ts:13`); todo lo demás es nuevo |
| S-PORT | **Tarjeta del Portafolio**: una fila por arnés × puesto con costo por corrida, tendencia y punto de mejora | nuevo |

**Detectores en alcance — seis, y solo seis (D16.1):**

| id | qué muestra | S1 | S2 |
|---|---|:--:|:--:|
| **B4** | gasto por arnés × empresa × puesto (el espinazo del join) | ✅ | ✅ |
| **P1** | caja que consume y se rechaza en el gate | ✅ | ✅ |
| **B2** | costo de la rotación de contexto | ✅ | ⚠️ |
| **B6** | sesión abandonada (escribe cache, nunca lo lee) | ✅ | ✅ |
| **B3** | cambio de modelo que invalida el cache | ✅ | ✅ |
| **B1** | re-warm por TTL vencido (break-even 39,47 %) | ✅ | ❌ **no aplica** |

**Fuera de alcance (explícito):**

- **Los otros 7 detectores** (B5 · B7 · B8 · B9 · B10 · B11 · B12 · B13). No pasan A4 todavía. Se
  declaran **no medidos** en la UI, con el patrón `sin-check`; jamás en cero.
- **Capa Desempeño** — la señal llega (`api_request.duration_ms`), falta el diseño de qué es
  «desempeño» a nivel Mapa. El slot sigue `disabled`, con motivo nuevo (RF-233).
- **Capa Proceso completa** — entra parcialmente por P1; el mapeo completo de eventos a fases queda
  fuera. El slot sigue `disabled` (RF-233).
- **S3** (arnés en máquina de un cliente sin ArnesIA) — descartado por D12.1.
- **Series temporales / tablero de tendencia por día.** Es lo que ya venden ccusage y Dynatrace
  (D9.8/H1). La única tendencia que entra es el sparkline de 5 puntos del Portafolio (RF-266).
- **Ejecutar el fix.** `[Proponerlo en el chat]` abre el chat; el cambio lo aplica el camino ya
  firmado con sus permisos y su gate (RF-255).
- **Forward OTLP a un backend externo.** Existe como escape hatch del operador, apagado por default
  (D13/C2); su superficie de configuración **no** se construye en este paquete (hueco H-7).

## 1 · Reglas de negocio (BR)

| id | regla | por qué |
|---|---|---|
| **BR-M1** | **Solo las cajas llevan cifra.** La granularidad real es `arnés × caja × sesión × turno` (D14.3) | Un número sobre una regla, un MCP, un subagente o un nodo de conocimiento sería inventado |
| **BR-M2** | Todo lo que no es caja va **«sin dato atribuible» con motivo**, jamás en 0 | `skill_activated` no trae hash (V3): por-skill no se recupera. Un 0 es una afirmación, y sería falsa |
| **BR-M3** | **Cada cifra carga su `atribucion_confianza`** y la UI muestra **los cuatro** niveles como casos distinguibles, cada uno con su copy: `exacta` → `por-hash` → `por-proceso` → `sin-dato`, en ese orden de preferencia | D16.2 + ANEXO H5. `por-proceso` deja de ser teórico: los hooks traen `cwd` en todos los payloads y el índice ya conoce la instalación por `(home, id)`. Un genérico «atribución aproximada» tapa tres cosas distintas |
| **BR-M4** | El costo es **estimado**, no facturación, y la UI lo dice en la superficie | `claude_code.cost.usage` está documentado como *"Estimated cost"* (D10) |
| **BR-M5** | **A4:** una cifra se muestra solo si (1) sale de dato propio, (2) está cotizada en dinero y (3) viene con **UN** fix concreto | Sin las tres es opinión, no producto |
| **BR-M6** | **A2:** el veredicto es contrafactual (*«el mundo alternativo costaba $Y»*), nunca «gastaste $X» | Convierte un tablero en una decisión |
| **BR-M7** | **A3:** el sesgo se declara, con su dirección, y va **en contra** de la recomendación | Si igual conviene, conviene de verdad |
| **BR-M8** | **A1:** todo umbral se cita como desigualdad algebraica con sus constantes | Se audita a mano, no se cree |
| **BR-M9** | **A5:** una fila en `0,00` es un dato honesto **cuando el concepto existe** en ese runtime. Cuando **no** existe, la fila dice «no aplica» — nunca 0 | D-4: un 0 donde el concepto no existe es mentira |
| **BR-M10** | **A6:** se guardan los dos costos (el que reportó el runtime y el que dice nuestro catálogo) y la UI los enfrenta | El oracle independiente sin construir un segundo sistema |
| **BR-M11** | **A7:** el score se versiona y la versión es visible | La gente los va a comparar entre releases |
| **BR-M12** | **A8/D17.3:** ninguna acción de esta capa escribe archivos de un arnés. `[Proponerlo en el chat]` **abre el chat** | No se abre una segunda vía de escritura hacia el paquete de un tercero |
| **BR-M13** | Un detector que **no aplica** se muestra **apagado con el motivo**, nunca oculto | Un detector ausente sin explicación se lee como «todo bien» |
| **BR-M14** | **Ni identidad ni contenido se persisten.** Allowlist —no denylist— **en los DOS caminos de ingesta**: el receptor OTLP y el hook | D15.1 + ANEXO H4. Por OTLP llega identidad (`user.email`, `user.account_uuid`, `organization.id`, V6) con el contenido ya `<REDACTED>`. **Por el hook llega el contenido en claro**: `UserPromptSubmit.prompt`, `Stop.last_assistant_message`, `PostToolUse.tool_response`. Son dos fugas distintas y hacen falta dos allowlists |
| **BR-M15** | El arnés **nunca** egresa: loopback o disco local. **Y proyecta antes de escribir**: un hook que postea su stdin entero a `127.0.0.1` cumple «no egresa» y aun así vuelca la conversación al almacén local. El forward del daemon es del operador, va apagado y filtra en el borde | D13 + ANEXO H4: «no egresa» y «no guarda de más» son dos checks, no uno |
| **BR-M16** | **Superset estricto:** nada firmado en la PARIDAD del Mapa se quita ni se mueve de sitio | Disciplina de paquete · `mockups/INDEX.md` regla 3 |

---

## A · La barra y la franja de contexto (6 RF)

### RF-232 — El slot `Tokens` del conmutador pasa a llamarse «Mejora» y se enciende 🎨 🔴
`mockup-capa-mejora.html:303` (propuesta) · `:272` (lo vigente, calcado) ·
`widgets/map-canvas/model/layers.ts:13` (el código que cambia).

**Desviación declarada de un baseline firmado.** Los cuatro slots se conservan; cambia una
etiqueta. «Tokens» nombra el insumo, no el producto (D16.3).

```gherkin
Escenario: el conmutador ofrece la capa Mejora
  Dado el Mapa con un arnés cargado
  Cuando se pinta el conmutador de capas
  Entonces los cuatro slots siguen presentes en el mismo orden
  Y el segundo se llama "Mejora", no "Tokens"
  Y "Mejora" NO está deshabilitado

Escenario: conmutar a la capa Mejora
  Dado el Mapa en la capa Estructura
  Cuando el operador elige "Mejora"
  Entonces la capa Mejora queda seleccionada
  Y la capa Estructura queda deseleccionada
  Y la geografía del canvas no cambia de sitio
```

### RF-233 — `Desempeño` y `Proceso` siguen apagados, con el motivo REAL 🎨
`mockup-capa-mejora.html:304-305`. Hoy el tooltip dice «Necesita telemetría (indexer JSONL)»
(`map-bar.tsx:123`). **Eso deja de ser verdad**: la señal llega (V1). Lo que falta es el diseño.

```gherkin
Escenario: el motivo del slot Desempeño
  Cuando el operador apunta al slot "Desempeño"
  Entonces el motivo dice que la señal ya llega y que lo que falta es el diseño
  Y NO dice "necesita telemetría"

Escenario: el motivo del slot Proceso
  Cuando el operador apunta al slot "Proceso"
  Entonces el motivo dice que entra parcialmente por la capa Mejora
  Y nombra lo que falta: el mapeo completo de eventos a fases
```

### RF-234 — Ventana temporal explícita 🎨
`mockup-capa-mejora.html:308-311`. Sin ventana, «cuánto gasta» no significa nada.

```gherkin
Escenario: la ventana nace en 7 días
  Cuando se activa la capa Mejora
  Entonces la franja de contexto ofrece elegir la ventana: 7 días, 30 días o todo
  Y arranca en 7 días

Escenario: cambiar la ventana recalcula todo
  Dado la capa Mejora con ventana de 7 días
  Cuando el operador elige 30 días
  Entonces el total, la cobertura, las cifras por caja y los puntos de mejora se recalculan sobre esa ventana
  Y ninguna cifra visible queda calculada sobre una ventana distinta de la elegida
```

### RF-235 — El total del arnés en la ventana, con su denominador 🎨
`mockup-capa-mejora.html:312-315`.

```gherkin
Escenario: el total y sobre qué se calculó
  Dado un arnés con 61 corridas en 12 sesiones sobre 4 cajas en la ventana
  Cuando se pinta la franja de contexto
  Entonces muestra "USD 4,82"
  Y debajo muestra "61 corridas · 12 sesiones · 4 cajas"
  Y el total es exactamente la suma de las cifras de las cajas visibles
```

### RF-236 — El disclaimer «estimado, no facturación» va en la superficie 🎨
`mockup-capa-mejora.html:316`. **No** en un `title`, **no** en un pie de página.

```gherkin
Escenario: la cifra dice qué es
  Cuando se pinta el total de la capa Mejora
  Entonces junto al total se lee "estimado por el runtime, no es facturación"
  Y ese texto es visible sin hover, sin foco y sin desplegar nada
```

### RF-237 — Barra de cobertura por calidad de atribución — CUATRO segmentos 🎨
`mockup-capa-mejora.html:317-322` (el mockup dibuja **tres**; el cuarto es el hueco **H-13**).
Es lo que impide leer el total como si fuera completo.

```gherkin
Escenario: cobertura con las cuatro calidades
  Dado 12 corridas exactas, 3 por huella, 2 por proceso y 1 sin dato
  Cuando se pinta la franja de contexto
  Entonces la barra tiene cuatro segmentos proporcionales a 12, 3, 2 y 1
  Y al lado se lee "12 exactas · 3 por huella · 2 por proceso · 1 sin dato"
  Y el texto declara la UNIDAD que cuenta (corridas), no la deja implícita
  Y los cuatro segmentos son distinguibles entre sí sin depender del color

Escenario: una calidad que no aparece no ocupa lugar
  Dado que ninguna corrida se atribuyó por proceso
  Entonces ese segmento no se dibuja
  Y el texto no menciona la categoría en cero

Escenario: cobertura completa
  Dado que todas las corridas tienen atribución exacta
  Entonces la barra es un solo segmento
  Y el texto dice que la atribución es exacta en todas
```

---

## B · El canvas (8 RF)

### RF-238 — Solo las cajas llevan cifra 🎨
`mockup-capa-mejora.html:331-334` (la regla) · `:341-348` (una caja con cifra).

```gherkin
Escenario: una caja lleva su cifra
  Dado la capa Mejora activa
  Y una caja con costo atribuido en la ventana
  Entonces la caja muestra su cifra en USD dentro de su tarjeta

Escenario: un nodo que no es caja jamás lleva cifra
  Dado la capa Mejora activa
  Cuando se pinta un nodo que no es caja (`isCaja(box) === false`)   # D19: `knowledge` no es una clase, es una banda
  Entonces ese nodo NO muestra ninguna cifra
  Y NO muestra 0
```

### RF-239 — Cifra + participación porcentual sobre el arnés 🎨
`mockup-capa-mejora.html:345` · `:359` · `:376` · `:390`.

```gherkin
Escenario: la caja dice cuánto y qué proporción
  Dado un arnés que gastó USD 4,82 en la ventana
  Y una caja que gastó USD 1,92
  Entonces la caja muestra "USD 1,92" y "40 %"
  Y la suma de los porcentajes de las cajas con dato no supera 100 %
```

### RF-240 — Barra de participación al pie de la caja 🎨
`mockup-capa-mejora.html:347` · `:360` · `:378` · `:392`.

```gherkin
Escenario: la proporción se lee sin leer el número
  Dado una caja con 40 % de participación
  Entonces al pie de su tarjeta hay una barra que ocupa el 40 % del ancho
  Y la barra tiene texto equivalente para lector de pantalla

Escenario: caja sin dato
  Dado una caja sin costo atribuido en la ventana
  Entonces NO se dibuja barra de participación
  Y la caja dice por qué no tiene dato
```

### RF-241 — La marca de fuga nombra el detector, nunca un ⚠ genérico 🎨
`mockup-capa-mejora.html:346` (`re-warm TTL`) · `:377` (`3 de 4 rechazos en gate`) · `:391`
(`por huella`).

```gherkin
Escenario: el detector se nombra
  Dado una caja donde el detector B1 encontró re-warm por TTL
  Entonces la caja muestra una marca que dice "re-warm TTL"
  Y NO muestra un icono de advertencia sin texto

Escenario: dos detectores en la misma caja
  Dado una caja con hallazgos de dos detectores
  Entonces se muestra la marca del de mayor ahorro contrafactual
  Y la tarjeta de punto de mejora de esa caja lista los dos
```

### RF-242 — Una cifra marcada según CÓMO se atribuyó — cuatro casos, cuatro copys 🎨
`mockup-capa-mejora.html:390` (subrayado punteado + motivo) · `:407-411` (la regla) · `:150` (la
marca) · ANEXO H5 (el orden de preferencia). El mockup solo distingue **exacta** y **por huella** —
los otros dos son el hueco **H-13**.

Orden de preferencia al atribuir: `exacta` → `por-hash` → `por-proceso` → `sin-dato`. Se usa el
primero que resuelva; la UI muestra **cuál** se usó, nunca «aproximada» a secas.

```gherkin
Escenario: atribución exacta
  Dado una cifra atribuida por los resource attributes que inyectamos al spawnear
  Entonces se pinta sin marca de duda
  Y no lleva motivo

Escenario: atribución por huella del arnés
  Dado una cifra atribuida por plugin_id_hash contra la tabla local
  Entonces va con marca de duda
  Y el motivo dice que se identificó por la huella del arnés, no por su nombre, porque corrió fuera de ArnesIA
  Y la caja muestra la marca corta "por huella"

Escenario: atribución por proceso
  Dado una corrida sin resource attributes y con un plugin_id_hash desconocido
  Y un cwd que mapea a una instalación conocida del Portafolio
  Entonces la cifra se atribuye a esa instalación
  Y va con marca de duda
  Y el motivo dice que se dedujo por el directorio donde corrió, no por el arnés en sí
  Y la caja muestra la marca corta "por proceso"
  Y el motivo aclara que si en ese directorio corre más de un arnés, el número los mezcla

Escenario: sin dato
  Dado una corrida que ninguna de las tres vías pudo atribuir
  Entonces NO se suma a ninguna caja
  Y suma al segmento "sin dato" de la barra de cobertura
  Y NO desaparece del total: el total dice sobre cuántas corridas se calculó

Escenario: los cuatro se distinguen entre sí
  Cuando se ven las cuatro marcas juntas
  Entonces cada una tiene texto propio
  Y ninguna se pinta como "atribución aproximada" genérica
```

### RF-243 — «Sin dato atribuible» con motivo, en cada nodo que no puede llevar cifra 🎨
`mockup-capa-mejora.html:362-365` (subagente) · `:398-401` (regla) · `:402-405` (MCP).

Tres motivos distintos, no un texto genérico.

```gherkin
Escenario: subagente
  Dado un nodo de clase subagent en la capa Mejora
  Entonces dice "sin dato atribuible — el subagente no se distingue en el turno"

Escenario: regla
  Dado un nodo de clase rule
  Entonces dice "sin dato atribuible — una regla no consume por sí misma"

Escenario: MCP
  Dado un nodo de clase mcp
  Entonces dice "sin dato atribuible — el costo del MCP va en la caja que lo llama"

Escenario: nunca un cero
  Cuando se pinta cualquier nodo sin dato atribuible
  Entonces NO aparece "0", "USD 0,00" ni "—" a secas
```

### RF-244 — Cada carril muestra el total de su fase 🎨
`mockup-capa-mejora.html:339` · `:353` · `:370` · `:384`.

```gherkin
Escenario: total por fase
  Dado la capa Mejora activa
  Cuando se pinta el encabezado del carril "spec"
  Entonces muestra "USD 1,92"
  Y CONSERVA el contador de nodos que ya mostraba en la capa Estructura
  Y el total del carril es la suma de las cajas de ese carril

Escenario: carril sin cajas con dato
  Dado un carril cuyas cajas no tienen costo atribuido
  Entonces el encabezado dice "sin dato" en vez de USD 0,00
```

### RF-245 — La capa es un overlay: nada se mueve de sitio 🎨
`mockup-capa-mejora.html:331` (*«Nada se mueve de sitio»*) · BR-M16.

```gherkin
Escenario: la geografía es la misma en las dos capas
  Dado un arnés pintado en la capa Estructura
  Cuando el operador conmuta a la capa Mejora
  Entonces la Guardia, los carriles por fase, la Base y la franja de artefactos siguen en la misma posición
  Y los mismos nodos siguen presentes, con el mismo glifo, nombre y transición
  Y los edges no cambian

Escenario: volver no pierde nada
  Cuando el operador conmuta de Mejora a Estructura
  Entonces la superficie queda idéntica a como estaba antes de conmutar
  Y la selección de nodo se conserva
```

---

## C · La tarjeta de punto de mejora (12 RF)

> El corazón del entregable. Una tarjeta = un punto de mejora = una caja × un detector.

### RF-246 — Una tarjeta existe solo si la cifra pasa A4 🎨
`mockup-capa-mejora.html:419-421`.

```gherkin
Escenario: la regla de aceptación
  Dado un hallazgo de un detector
  Cuando el hallazgo sale de dato propio, está cotizado en dinero y tiene UN fix concreto
  Entonces se pinta su tarjeta de punto de mejora

Escenario: hallazgo sin fix
  Dado un hallazgo cotizado pero sin un fix concreto asociado
  Entonces NO se pinta tarjeta
  Y el detector queda listado como "sin fix propuesto" en el inspector, no oculto
```

### RF-247 — Titular: qué caja y qué le pasa, en lenguaje del usuario 🎨
`mockup-capa-mejora.html:425` · `:447`.

```gherkin
Escenario: el titular nombra la caja y el problema
  Dado un punto de mejora del detector B1 sobre la caja "escribir el spec"
  Entonces el titular dice: La caja «escribir el spec» reescribe el cache en cada corrida
  Y NO nombra el identificador interno del detector
  Y NO usa jerga del runtime en el titular
```

### RF-248 — Lede: cuánto, sobre qué base y qué proporción 🎨
`mockup-capa-mejora.html:426-427` · `:448-450`.

```gherkin
Escenario: el monto siempre trae su base
  Dado un punto de mejora que cuesta USD 0,84 sobre una caja de USD 1,92
  Entonces el lede dice "USD 0,84 de los 1,92 de la caja (44 %)"
  Y explica en una frase qué está pasando
  Y NUNCA muestra un monto sin su base
```

### RF-249 — Contrafactual, no «gastaste X» (A2) 🎨
`mockup-capa-mejora.html:429-430` · `:452-453`.

```gherkin
Escenario: el mundo alternativo
  Dado un punto de mejora con fix "TTL de 1 h"
  Entonces la tarjeta dice qué habrían costado las MISMAS corridas con el fix aplicado
  Y dice la diferencia
  Y la diferencia declara su unidad (por corrida, o en la ventana) sin ambigüedad
```

### RF-250 — Umbral algebraico citado, auditable a mano (A1) 🎨
`mockup-capa-mejora.html:431-432`.

```gherkin
Escenario: la desigualdad se muestra entera
  Dado el detector B1 con break-even de re-warm
  Entonces la tarjeta muestra la desigualdad con sus constantes: (2−1,25)/(2−0,1) = 39,47 %
  Y muestra el valor medido que la cruza: relectura 61 %
  Y ofrece "ver el cálculo" para el detalle completo

Escenario: un detector sin umbral algebraico
  Dado un detector que no se decide por un umbral (P1, rechazo en el gate)
  Entonces la fila "Umbral" no se pinta
  Y en su lugar se pinta el patrón observado
```

### RF-251 — Confianza con denominador explícito 🎨
`mockup-capa-mejora.html:433-434` · `:456-457`.

```gherkin
Escenario: la confianza dice sobre cuántas corridas
  Dado un punto de mejora calculado sobre 14 corridas, todas con atribución exacta
  Entonces la tarjeta dice "exacta · 14 de 14 corridas con atribución"

Escenario: confianza degradada
  Dado un punto de mejora calculado sobre 14 corridas, 3 de ellas atribuidas por huella
  Entonces la tarjeta dice que la atribución no es exacta en todas
  Y dice cuántas y por qué mecanismo
```

### RF-252 — El sesgo se declara y va EN CONTRA de la recomendación (A3) 🎨
`mockup-capa-mejora.html:435-436` (subestima) · `:458-459` (sobreestima).

```gherkin
Escenario: sesgo que subestima el ahorro
  Dado un cálculo que asume 0 lecturas fuera de la ventana
  Entonces la tarjeta dice el supuesto
  Y dice su dirección: subestima el ahorro

Escenario: sesgo que sobreestima el desperdicio
  Dado un cálculo que no descuenta lo que la revisión aporta aunque rechace
  Entonces la tarjeta dice el supuesto
  Y dice su dirección: sobreestima el desperdicio

Escenario: no hay tarjeta sin sesgo declarado
  Cuando se pinta cualquier tarjeta de punto de mejora
  Entonces la fila "Sesgo" está siempre presente
  Y si no se identificó ningún sesgo, lo dice explícitamente en vez de omitir la fila
```

### RF-253 — Un fix, concreto, y dónde se aplica 🎨
`mockup-capa-mejora.html:439` · `:462`.

```gherkin
Escenario: el fix es una acción, no un consejo
  Dado un punto de mejora del detector B1
  Entonces el fix dice: Fijar cache_ttl: 1h en esta caja
  Y nombra el ajuste concreto y el lugar

Escenario: un solo fix por tarjeta
  Cuando se pinta una tarjeta
  Entonces se propone exactamente UN fix
  Y si hay dos caminos posibles, son dos tarjetas o una decisión previa, nunca dos botones
```

### RF-254 — La versión del score es visible (A7) 🎨
`mockup-capa-mejora.html:440` · `:463`.

```gherkin
Escenario: el score se versiona
  Cuando se pinta una tarjeta de punto de mejora
  Entonces muestra la versión del score con que se calculó
  Y la versión es legible sin hover
```

### RF-255 — `[Proponerlo en el chat]` abre el chat: NO escribe archivos 🎨
`mockup-capa-mejora.html:442` · `:468-472` (la regla) · D17.3.

```gherkin
Escenario: proponer el fix
  Dado una tarjeta de punto de mejora
  Cuando el operador usa "Proponerlo en el chat"
  Entonces se abre el chat con el cambio propuesto ya redactado
  Y NINGÚN archivo del arnés se modifica por esta acción
  Y el cambio, si se aplica, pasa por los permisos, el alcance y el gate del camino ya firmado

Escenario: el alcance del chat sigue vigente
  Dado que el arnés propuesto es el paquete propio de ArnesIA
  Cuando el operador usa "Proponerlo en el chat"
  Entonces se respeta el guardrail vigente de alcance del chat embebido
  Y si el cambio queda fuera de alcance, se dice antes de abrir el chat
```

### RF-256 — `[Descartar]` saca la tarjeta sin borrar el dato 🎨
`mockup-capa-mejora.html:441` · `:464`.

```gherkin
Escenario: descartar un punto de mejora
  Dado una tarjeta de punto de mejora
  Cuando el operador la descarta
  Entonces la tarjeta deja de mostrarse en esta superficie
  Y la cifra de la caja NO cambia
  Y el descarte es reversible desde el inspector de esa caja
```

### RF-257 — La severidad no depende solo del color 🎨
`mockup-capa-mejora.html:424` (tono `warn`) vs `:446` (tono `crit`).

```gherkin
Escenario: dos severidades distinguibles en escala de grises
  Dado dos tarjetas, una de severidad atención y otra crítica
  Cuando se ven sin color
  Entonces siguen siendo distinguibles por otro atributo además del color
  Y la severidad tiene texto equivalente para lector de pantalla
```

---

## D · Inspector — cuarta tab «Mejora» (7 RF)

### RF-258 — Cuarta tab, y las tres vigentes intactas 🎨
`mockup-capa-mejora.html:480-481` (la regla) · `:492-497` (las cuatro tabs) ·
`widgets/map-canvas/ui/inspector.tsx:111-115` (el arreglo `TABS` que se extiende).

```gherkin
Escenario: la cuarta tab existe
  Dado el inspector de una caja
  Entonces hay cuatro tabs: Resumen, Contenido, Corridas, Mejora
  Y Resumen, Contenido y Corridas se comportan exactamente igual que antes
  Y la tab por defecto sigue siendo Resumen

Escenario: nodo que no es caja
  Dado el inspector de un nodo que no es caja
  Cuando se abre la tab Mejora
  Entonces dice por qué no hay cifra para ese nodo, con el mismo motivo que muestra el canvas
  Y NO muestra tablas vacías ni ceros
```

### RF-259 — Desglose por bucket de tokens con su costo 🎨
`mockup-capa-mejora.html:499-511`.

```gherkin
Escenario: el número se puede auditar
  Dado una caja con USD 1,92 en la ventana
  Cuando se abre la tab Mejora
  Entonces hay una fila por bucket: entrada, salida, cache lectura, cache escritura 5 m, cache escritura 1 h, razonamiento
  Y cada fila muestra tokens y USD
  Y la suma de los USD de las filas es el total de la caja
```

### RF-260 — «No aplica» ≠ 0 🎨
`mockup-capa-mejora.html:509` (razonamiento) · `:512-513` (la regla) · `:508` (un 0 legítimo).

```gherkin
Escenario: concepto que el runtime no tiene
  Dado un runtime sin tokens de razonamiento
  Entonces la fila "razonamiento" dice "no aplica en este runtime"
  Y NO dice 0

Escenario: concepto que el runtime sí tiene, con valor cero
  Dado un runtime con cache de 1 h, que en esta ventana no escribió ninguno
  Entonces la fila "cache · escritura 1 h" dice 0 y 0,00
  Y ese cero es un dato, no un hueco
```

### RF-261 — Costo reportado vs. calculado, a la vista (A6) 🎨
`mockup-capa-mejora.html:515-524`.

```gherkin
Escenario: los dos costos coinciden
  Dado que el runtime reportó USD 1,92 y nuestro catálogo calcula USD 1,92
  Entonces la sección muestra los dos números y dice que coinciden

Escenario: los dos costos divergen
  Dado que el runtime reportó USD 1,92 y nuestro catálogo calcula USD 2,10
  Entonces la sección muestra los dos números y marca la divergencia
  Y dice las dos causas posibles: catálogo viejo o tarifa cambiada
  Y NO elige una de las dos por su cuenta

Escenario: el runtime no reporta costo
  Dado un runtime que no emite costo en dinero
  Entonces se muestra solo el calculado, rotulado como calculado por nosotros
  Y se nombra la versión del catálogo usada
```

### RF-262 — El join, a nivel nodo 🎨
`mockup-capa-mejora.html:525-535`. Esta tabla sola es la mitad de la frase objetivo.

```gherkin
Escenario: dinero y proceso en la misma tabla
  Dado una caja con 14 corridas en la ventana
  Entonces la tab muestra: corridas, rechazadas en el gate, costo de las rechazadas, rotaciones de contexto
  Y el costo de las rechazadas está en USD

Escenario: la señal de proceso no está disponible
  Dado una caja cuyas corridas no dejaron evento de gate
  Entonces las filas de proceso dicen que no hay señal y por qué
  Y las filas de dinero se muestran igual
```

### RF-263 — Detectores, incluidos los que NO aplican, con motivo 🎨
`mockup-capa-mejora.html:536-543` · BR-M13.

```gherkin
Escenario: los seis detectores del MVP están listados
  Dado la tab Mejora de una caja
  Entonces se listan los seis detectores en alcance
  Y cada uno está en uno de tres estados: activo con hallazgo, sin hallazgos, o no disponible con motivo

Escenario: detector no disponible
  Dado el detector B2 en una caja donde 2 de 3 rotaciones ocurrieron fuera de ArnesIA
  Entonces se muestra apagado
  Y el motivo dice: no disponible: 2 de 3 rotaciones ocurrieron fuera de ArnesIA

Escenario: los detectores fuera del MVP
  Cuando se lista la sección de detectores
  Entonces los siete detectores fuera de alcance se declaran como no medidos todavía
  Y NO se muestran como "sin hallazgos"
```

### RF-264 — La tab declara sobre qué ventana calculó 🎨
`mockup-capa-mejora.html:500` (`Tokens · últimas 14 corridas`).

```gherkin
Escenario: la ventana de la tab es la de la capa
  Dado la capa Mejora con ventana de 7 días y 14 corridas de esta caja en esa ventana
  Cuando se abre la tab Mejora
  Entonces el encabezado dice sobre qué ventana calculó, en la misma unidad que la franja de contexto
  Y si la ventana de la capa cambia, el encabezado y las cifras cambian con ella
```

---

## E · La tarjeta del Portafolio (4 RF)

> D17.2 · sin esta superficie el MVP no puede decir «en este puesto», que es el eje diferencial.

### RF-265 — Una fila por arnés × puesto, con costo por corrida 🎨

> **D20 (FIRMADA 2026-07-26):** no existe `puesto` en `EntradaPortafolio`. La fila se agrupa por
> **`(identidad, instalacion_id)`** —la unidad que el Portafolio sí modela— y la etiqueta de puesto
> sale del **`rol` del arnés indexado**, resuelto **server-side** en `GET /api/telemetria/portafolio`
> y viajando como `puesto *string`. Sin `rol` viaja **`null`** y la UI dice `puesto sin declarar`.
> «Un arnés en dos puestos» se cumple **por instalación**: dos instalaciones ⇒ dos filas.
> `domain.EntradaPortafolio` y el wire de `GET /api/portafolio` **no se tocan**.
`mockup-capa-mejora.html:571-578`.

```gherkin
Escenario: la fila del Portafolio
  Dado un arnés "vitalia" en el puesto "Producto · Product Owner" con 0,31 USD por corrida
  Cuando se pinta la tarjeta del Portafolio
  Entonces la fila muestra el arnés, el puesto y "0,31"
  Y la columna dice explícitamente que es USD por corrida

Escenario: un arnés en dos puestos
  Dado un arnés instalado en dos puestos distintos
  Entonces hay dos filas, una por puesto
  Y cada una con su propio costo por corrida
```

### RF-266 — Tendencia, con texto equivalente 🎨
`mockup-capa-mejora.html:577` (en alza) · `:583` (estable).

```gherkin
Escenario: la tendencia se lee sin ver el dibujo
  Dado una fila con tendencia en alza
  Entonces el sparkline tiene texto equivalente que dice la dirección
  Y el último punto se distingue de los anteriores

Escenario: menos puntos que el mínimo
  Dado un arnés con una sola corrida en la ventana
  Entonces no se dibuja sparkline
  Y se dice que no hay suficientes corridas para una tendencia
```

### RF-267 — Punto de mejora principal por fila, con su monto 🎨
`mockup-capa-mejora.html:578` (con fuga) · `:584` (sin fugas).

```gherkin
Escenario: la fila lleva su punto de mejora
  Dado un arnés con un punto de mejora de USD 0,53 por corrida
  Entonces la fila muestra el detector nombrado y el monto: re-warm de cache · USD 0,53/corrida

Escenario: sin fugas detectadas
  Dado un arnés con datos y sin ningún hallazgo de los seis detectores
  Entonces la fila dice "sin fugas detectadas"
  Y ese estado se distingue visualmente de "sin datos"
```

### RF-268 — «Sin dato» honesto: la fila no desaparece ni va en cero 🎨
`mockup-capa-mejora.html:586-591` · `:594-595`.

```gherkin
Escenario: arnés que nunca corrió con telemetría
  Dado un arnés del Portafolio sin ninguna corrida instrumentada
  Entonces la fila SIGUE en la tabla
  Y la columna de costo dice "sin dato", no 0,00
  Y la columna de tendencia queda vacía marcada como tal
  Y la columna de punto de mejora dice "nunca corrió con telemetría"

Escenario: no se puede ordenar mintiendo
  Dado una tabla con arneses con dato y sin dato
  Cuando se ordena por costo
  Entonces los "sin dato" quedan agrupados aparte
  Y NO se ordenan como si valieran 0
```

---

## F · Los siete estados honestos (7 RF)

> Son RF de primera clase, no notas al pie. Un mockup que solo dibuja el caso feliz no alcanza
> para firmar, y un build que solo construye el caso feliz tampoco.

### RF-269 — Estado 1 · Sin datos todavía 🎨
`mockup-capa-mejora.html:608-612`.

```gherkin
Escenario: el arnés nunca corrió con telemetría
  Dado un arnés sin ninguna corrida instrumentada
  Cuando se activa la capa Mejora
  Entonces NO se pinta un tablero en cero
  Y se dice que este arnés nunca corrió con telemetría
  Y se dice qué hacer para que empiece a haber datos
```

### RF-270 — Estado 2 · Cobertura parcial 🎨
`mockup-capa-mejora.html:613-617`.

```gherkin
Escenario: el total se calculó sobre parte de las corridas
  Dado 5 corridas en la ventana, 2 sin atribución
  Entonces el total dice sobre cuántas corridas se calculó: de 5 corridas, 3
  Y dice cuántas quedaron sin atribución
  Y el total NO se presenta como completo
```

### RF-271 — Estado 3 · Corrió fuera de ArnesIA (S2 degradado) 🎨
`mockup-capa-mejora.html:618-622` · `:666` (la traza) · D16.1 (B1 no aplica en S2).

```gherkin
Escenario: B1 no aplica porque el hook no ve el stream-json
  Dado un arnés que corrió fuera de ArnesIA
  Cuando se listan los detectores
  Entonces "re-warm por TTL" se muestra APAGADO
  Y el motivo dice: no disponible — sin el result del stream-json
  Y el detector NO se esconde
  Y los otros cinco detectores siguen funcionando

Escenario: mezcla S1 + S2 en la misma ventana
  Dado corridas dentro y fuera de ArnesIA en la misma ventana
  Entonces B1 se calcula solo sobre las que tienen la señal
  Y dice sobre cuántas de cuántas se calculó

Escenario: los dos modos de S2 se distinguen (ANEXO H9, ver J-10)
  Dado un arnés que corrió fuera de ArnesIA SIN telemetría instrumentada
  Entonces se dice que corrió fuera y sin instrumentar, y qué detectores no aplican

  Dado un arnés que corrió fuera de ArnesIA CON telemetría instrumentada
  Entonces se dice que corrió fuera pero instrumentado
  Y las cifras de dinero se muestran igual que en S1
  Y los dos casos NO se rotulan con la misma frase
```

### RF-272 — Estado 4 · Otro runtime, costo puesto por nosotros 🎨
`mockup-capa-mejora.html:623-627`.

```gherkin
Escenario: el runtime no reporta costo en dinero
  Dado un runtime que emite tokens pero no costo
  Entonces la cifra se muestra igual
  Y se dice que fue calculada con nuestro catálogo, nombrando su versión
  Y se dice que este runtime no reporta costo
```

### RF-273 — Estado 5 · Catálogo de precios viejo 🎨
`mockup-capa-mejora.html:628-632` · D11 (embeber + refresco opcional con degradación honesta).

```gherkin
Escenario: el catálogo nunca se refrescó
  Dado un catálogo embebido en el release, sin refresco desde el 20/07
  Entonces se avisa que los precios son los del release y desde cuándo
  Y la app sigue funcionando sin internet
  Y el aviso está en la superficie, no en un tooltip
```

### RF-274 — Estado 6 · Atribución por huella 🎨
`mockup-capa-mejora.html:633-637` · `:660` (la traza) · V3.

```gherkin
Escenario: el runtime redactó el nombre del plugin
  Dado una corrida donde plugin.name llegó como "third-party"
  Y un plugin_id_hash conocido en la tabla local
  Entonces la cifra se atribuye a ese arnés
  Y se marca como atribuida por huella, no por nombre
  Y se dice cuál de las dos vías se usó

Escenario: huella desconocida
  Dado un plugin_id_hash que no está en la tabla local
  Entonces la cifra NO se atribuye a ningún arnés
  Y suma al segmento "sin dato" de la cobertura
```

### RF-275 — Estado 7 · Qué guardamos, y el botón de borrado 🎨
`mockup-capa-mejora.html:638-642` · D15 · **ANEXO H4**.

El mockup promete *«Nada de tu cuenta. Nada del contenido.»* El anexo verificó que por el canal de
hooks el contenido llega **en claro** (`UserPromptSubmit.prompt`, `Stop.last_assistant_message`,
`PostToolUse.tool_response`). La promesa **se mantiene**, pero deja de ser una afirmación sobre el
canal y pasa a ser una afirmación sobre **lo que se guarda** — y por eso tiene que ser verificable
sobre los dos caminos (RF-282), no solo escrita.

**El copy no cambia de tono.** Sigue siendo corto y en primera persona del producto; lo que se
agrega es una segunda línea que dice qué pasa con lo que sí llega. Un tratado legal no es el
entregable: el entregable es que un humano entienda en cinco segundos qué se guardó de él.

```gherkin
Escenario: la política de datos es visible desde la capa
  Cuando el operador consulta qué guarda la telemetría
  Entonces se dice que no se guarda nada de su cuenta
  Y se dice que no se guarda nada del contenido de la conversación ni de lo que leyeron las herramientas
  Y se dice que lo que llega con contenido se descarta antes de escribirse
  Y se dice la retención vigente
  Y hay una acción para borrar la telemetría de este arnés

Escenario: la promesa se puede contrastar
  Cuando el operador quiere ver qué campos SÍ se guardan
  Entonces la superficie lista los campos persistidos, por nombre
  Y esa lista es la misma allowlist que aplica la ingesta, no una redacción aparte

Escenario: borrar de verdad
  Dado telemetría acumulada de un arnés
  Cuando el operador confirma el borrado
  Entonces los datos de telemetría de ese arnés dejan de existir en el almacenamiento local
  Y las superficies que los mostraban pasan al estado 1 (sin datos)
  Y el borrado pide confirmación antes de ejecutarse
```

---

## G · Accesibilidad (6 RF)

> El patrón lo fijan los componentes vigentes: `map-bar.tsx:108-133` (tablist de capas con
> `role="tab"` + `aria-selected` + `disabled`) e `inspector.tsx:205-250` (tablist con
> `id`/`aria-controls`/`aria-labelledby` y `tabpanel` con `hidden`). La capa nueva **no inventa
> patrón**: extiende el que ya está firmado.

### RF-276 — El conmutador de capas mantiene su semántica ARIA, y el motivo es accesible
`mockup-capa-mejora.html:301-306` · `map-bar.tsx:108-133`.

```gherkin
Escenario: semántica del tablist de capas
  Cuando se pinta el conmutador con la capa Mejora encendida
  Entonces el contenedor sigue con role tablist y su aria-label
  Y cada slot sigue con role tab y su aria-selected
  Y "Mejora" tiene aria-selected true cuando está activa

Escenario: el motivo de un slot deshabilitado es accesible
  Dado los slots Desempeño y Proceso deshabilitados
  Entonces el motivo NO vive solo en el atributo title
  Y un lector de pantalla puede anunciar por qué están deshabilitados
```

### RF-277 — La cuarta tab del inspector respeta el contrato de tabs vigente
`mockup-capa-mejora.html:492-497` · `inspector.tsx:205-250`.

```gherkin
Escenario: la tab nueva se cablea igual que las tres vigentes
  Cuando se pinta la tab Mejora
  Entonces tiene role tab, su id propio y aria-controls apuntando a su panel
  Y su panel tiene role tabpanel y aria-labelledby apuntando a la tab
  Y el panel inactivo queda hidden, no montado-y-oculto por CSS

Escenario: cambiar de nodo
  Dado el inspector abierto en la tab Mejora de una caja
  Cuando el operador selecciona otro nodo
  Entonces el inspector vuelve a la tab Resumen, igual que hoy
```

### RF-278 — Foco visible en todo control nuevo
`mockup-capa-mejora.html:105` (`.seg button:focus-visible`) · `:176` (`.btn:focus-visible`).

```gherkin
Escenario: recorrer la capa con el teclado
  Dado la capa Mejora activa
  Cuando el operador tabula por la superficie
  Entonces cada control nuevo (ventana, botones de la tarjeta, tab Mejora, borrado) muestra anillo de foco visible
  Y el anillo usa el token de foco del sistema, no un outline del navegador por defecto
  Y ningún control queda inalcanzable por teclado
```

### RF-279 — Barras y sparklines tienen texto equivalente
`mockup-capa-mejora.html:318` (`role="img"` + `aria-label` de cobertura) · `:577` · `:583`
(sparklines) · `:347` (barra de participación).

```gherkin
Escenario: la barra de cobertura se puede leer
  Dado la barra de cobertura con 12 exactas, 3 por huella y 2 sin dato
  Entonces expone un texto equivalente con esos tres números y sus etiquetas

Escenario: la barra de participación se puede leer
  Dado la barra de participación de una caja al 40 %
  Entonces expone un texto equivalente con el porcentaje y a qué total se refiere

Escenario: el sparkline se puede leer
  Dado un sparkline de tendencia
  Entonces expone un texto equivalente con la dirección de la tendencia
  Y NO es la única forma de saberla
```

### RF-280 — Ningún estado depende solo del color
`mockup-capa-mejora.html:346` (fuga con texto) · `:391` (`por huella` con texto) · `:538-542`
(detectores con texto además del punto) · `:588-590` (sin dato con texto).

```gherkin
Escenario: los tres estados de un detector en escala de grises
  Dado un detector activo, uno sin hallazgos y uno no disponible
  Cuando se ven sin color
  Entonces los tres siguen siendo distinguibles por texto
  Y el punto de color es refuerzo, nunca el único portador

Escenario: la severidad de una tarjeta
  Dado una tarjeta de severidad crítica y otra de atención
  Entonces la diferencia se lee sin color

Escenario: sin dato vs. sin fugas
  Dado una fila del Portafolio sin datos y otra sin fugas detectadas
  Entonces los dos estados son distinguibles por texto, no solo por el tono del chip
```

### RF-281 — La cifra monetaria se escribe igual en las cinco superficies
`mockup-capa-mejora.html:77` (`.num` con `tabular-nums`) · `:313` · `:345` · `:504-508` · `:576`.

```gherkin
Escenario: un solo formato de dinero
  Cuando se muestra un monto en la franja de contexto, en una caja, en una tarjeta, en el inspector o en el Portafolio
  Entonces la moneda se escribe "USD" antepuesta
  Y el separador decimal es coma
  Y se usan dos decimales
  Y los dígitos se alinean con cifras tabulares

Escenario: un monto por debajo del centavo
  Dado un monto de USD 0,004
  Entonces NO se redondea a 0,00
  Y se muestra con la precisión necesaria o se agrega en un total que sí sea legible
```

---

## H · RF sin superficie (5 RF)

> Requisitos invisibles: no tienen píxel, y sin ellos la superficie miente o es un riesgo.
> **Se construyen ANTES que la UI.**

### RF-282 — Allowlist en los DOS caminos de ingesta: ni identidad ni contenido se persisten
Sin superficie; **es la condición de que RF-275 no sea una promesa vacía.**
D15.1 · V6 (identidad por OTLP) · **ANEXO H4** (contenido por el hook).

Son dos fugas distintas y hacen falta dos allowlists:

| camino | qué llega de más | qué se persiste |
|---|---|---|
| receptor **OTLP** | `user.email` · `user.account_uuid` · `user.account_id` · `organization.id` · `user.id`, en **cada** data point y **cada** log record | solo los campos del esquema canónico (D16.2). El contenido ya llega `<REDACTED>` |
| **hook** (S2) | `UserPromptSubmit.prompt` · `Stop.last_assistant_message` · `PostToolUse.tool_response`, **en claro** | proyección a un puñado de campos declarados: `session_id` · `prompt_id` · `hook_event_name` · `tool_name` · `duration_ms` · `cwd` normalizado |

```gherkin
Escenario: OTLP — solo se guarda lo declarado
  Dado un payload OTLP con atributos de usuario, cuenta y organización
  Cuando el receptor lo procesa
  Entonces se persisten únicamente los campos declarados en el esquema canónico
  Y ningún campo de identidad de usuario, cuenta u organización llega al almacenamiento

Escenario: hook — el contenido se descarta ANTES de escribir
  Dado un payload de hook con el prompt completo del usuario, la respuesta del asistente y el resultado de una herramienta
  Cuando el hook lo procesa
  Entonces proyecta a los campos declarados y descarta el resto
  Y el descarte ocurre antes de escribir a cualquier destino, incluido loopback
  Y en el almacenamiento local NO queda ninguna subcadena del prompt, de la respuesta ni del resultado de la herramienta

Escenario: un campo nuevo no se cuela
  Dado una versión futura del runtime que agrega un campo desconocido al payload
  Cuando cualquiera de los dos caminos lo procesa
  Entonces ese campo NO se persiste
  Y el hecho de haberlo descartado queda contable, para saber que el esquema quedó corto

Escenario: cwd normalizado
  Dado un payload con cwd = /home/<usuario>/Proyectos/vitalia
  Cuando se persiste
  Entonces se guarda la forma normalizada, sin el segmento de usuario
  Y alcanza para mapear a la instalación del Portafolio (RF-242, por-proceso)

Escenario: la llave del join no necesita nada de eso
  Dado un evento de telemetría y un evento de proceso del mismo turno
  Entonces el join se resuelve con el par (session_id, prompt_id)
  Y no se usa ningún identificador de cuenta ni ningún contenido para atribuir
```

### RF-283 — Retención con TTL, y borrado que borra
Sin superficie propia; su botón es RF-275. D15.3.

```gherkin
Escenario: la telemetría caduca sola
  Dado telemetría más vieja que el TTL de retención
  Cuando corre la purga
  Entonces esos registros se eliminan del almacenamiento local
  Y el TTL vigente es consultable desde la UI

Escenario: borrar por arnés
  Cuando se pide borrar la telemetría de un arnés
  Entonces se eliminan sus eventos, sus agregados y sus hallazgos de detectores
  Y no queda un agregado huérfano que siga alimentando una cifra
```

### RF-284 — Dos niveles de egreso, y el arnés ni egresa ni guarda de más
Sin superficie en este paquete (el indicador del forward es el hueco H-7). D13 · BR-M15 ·
**ANEXO H4** (el check hermano).

```gherkin
Escenario: el arnés escribe solo a loopback o a disco local
  Dado el hook de telemetría que viaja dentro de un arnés (escenario S2)
  Cuando emite un evento
  Entonces el destino es 127.0.0.1 o un archivo local
  Y jamás una dirección de red externa

Escenario: no alcanza con no egresar
  Dado un hook que postea su stdin entero a 127.0.0.1
  Cuando corre el check de conformance
  Entonces el check de destino pasa
  Y el check hermano de PROYECCIÓN DE CAMPOS falla
  Y el motivo dice que el hook escribe campos fuera de la allowlist

Escenario: el forward del daemon nace apagado
  Dado un daemon recién instalado
  Entonces el forward OTLP externo está apagado
  Y encenderlo es una acción deliberada del operador

Escenario: un arnés no puede tocar el forward
  Dado un arnés instalado
  Cuando intenta leer o configurar el forward del daemon
  Entonces no existe capacidad expuesta que se lo permita
```

### RF-285 — Cada cifra persiste su confianza y sus DOS costos
Sin superficie directa; alimenta RF-237, RF-242, RF-251 y RF-261. D16.2.

```gherkin
Escenario: la llave del join es un par, no un campo
  Cuando se persiste un evento de dinero y un evento de proceso del mismo turno
  Entonces los dos llevan session_id y prompt_id
  Y el join se resuelve por igualdad de ese par, sin heurística de tiempo ni de orden
  Y con session_id solo el join sería a nivel sesión, que NO alcanza para decir «el 60 % se va en la caja Y»

Escenario: la confianza viaja con el dato
  Cuando se persiste un evento de telemetría
  Entonces lleva su atribucion_confianza: exacta, por-hash, por-proceso o sin-dato
  Y ninguna cifra puede llegar a la UI sin ese campo

Escenario: los dos costos
  Cuando se persiste un evento
  Entonces se guarda el costo que reportó el runtime y el que calcula nuestro catálogo
  Y se guarda la versión del catálogo usada
  Y si el runtime no reporta costo, el campo queda ausente, no en cero
```

### RF-286 — «No aplica» se persiste como ausencia, no como cero
Sin superficie directa; es la condición de que RF-260 pueda ser verdad. D16.2 · D-4.

```gherkin
Escenario: un bucket que el runtime no tiene
  Dado un runtime sin tokens de razonamiento
  Cuando se persiste el evento
  Entonces el campo de razonamiento queda ausente
  Y NO se persiste como 0

Escenario: un bucket que el runtime tiene, en cero
  Dado un runtime con cache de 1 h que no escribió ninguno en ese turno
  Entonces el campo se persiste con valor 0
  Y la UI puede distinguir ese 0 de la ausencia
```

---

## 2 · Escenarios de verificación

Con datos reales: la evidencia cruda de `verificacion-2026-07-26/evidencia/` (3 corridas de
`claude 2.1.220`, 57 log records, 8 métricas, 14 puntos) y el dogfood `dev-full-cycle` de las
stories. **Prohibido mock donde hay dato real disponible.**

| id | escenario | insumo real | resultado esperado |
|---|---|---|---|
| T-01 | canal primario | payload `/v1/logs` de la evidencia | `api_request` decodificado con los 4 buckets + `cost_usd_micros` |
| T-02 | `intValue` off-spec | payload real (número JSON, no string) | decodifica sin error (V5) |
| T-03 | temporalidad Delta | 4 métricas `Sum · Delta · monotonic` | el receptor suma, no diferencia (V5.1) |
| T-04 | identidad descartada (OTLP) | payload con `user.email` real | ningún campo de cuenta en el almacenamiento (RF-282) |
| T-04b | **contenido descartado (hook)** | payload real de `UserPromptSubmit` + `Stop` + `PostToolUse` del anexo | ninguna subcadena del `prompt`, de `last_assistant_message` ni de `tool_response` aparece en el almacén; **búsqueda de texto sobre el archivo SQLite**, no inspección de structs (RF-282) |
| T-04c | campo nuevo desconocido | payload con un campo inventado | no se persiste, y el descarte queda contable (RF-282) |
| T-05 | atribución exacta | corrida con `OTEL_RESOURCE_ATTRIBUTES` inyectados | `atribucion_confianza = exacta` |
| T-06 | atribución por huella | corrida sin env vars, con `plugin_id_hash` conocido | `por-hash`, cifra subrayada punteada (RF-242) |
| T-06b | **atribución por proceso** | corrida sin env vars, hash desconocido, `cwd` que mapea a una instalación del Portafolio | `por-proceso`, con su copy propio — distinto del de huella (RF-242) |
| T-06c | **los cuatro niveles a la vez** | ventana con corridas de las cuatro clases | 4 segmentos en la cobertura, 4 copys distintos, ninguno «aproximada» (RF-237/RF-242) |
| T-07 | huella y `cwd` desconocidos | `plugin_id_hash` fuera de la tabla, `cwd` no mapeado | `sin-dato`, suma al segmento sin dato (RF-274) |
| T-07b | llave del join | evento OTel + evento de hook del mismo turno | unen por `(session_id, prompt_id)`, sin heurística (RF-285) |
| T-08 | split 5m/1h | `result` del stream-json de la corrida 3 | B1 calculable sin tocar el JSONL (V2) |
| T-09 | B1 en S2 | corrida sin stream-json | detector apagado con motivo (RF-271) |
| T-10 | «no aplica» ≠ 0 | evento sin razonamiento | fila «no aplica» (RF-260) |
| T-11 | cero legítimo | evento con `ephemeral_1h = 0` | fila con 0 y 0,00 (RF-260) |
| T-12 | paridad de costos | `cost_usd_micros` vs. catálogo | coinciden, o la divergencia se muestra (RF-261) |
| T-13 | total = suma | 4 cajas del dogfood | total de la franja = suma de las cajas (RF-235) |
| T-14 | cambio de ventana | 7 → 30 días | todas las cifras recalculan juntas (RF-234) |
| T-15 | ventana sin corridas | ventana de 7 días sobre un arnés que corrió hace un mes | estado propio, distinto de «nunca corrió» (hueco H-9) |
| T-16 | superset del canvas | story de `map-canvas` en Estructura vs. Mejora | misma geografía, mismos nodos, mismos edges (RF-245) |
| T-17 | nodo no-caja | regla, MCP, subagente del dogfood | los tres motivos distintos, cero cifras (RF-243) |
| T-18 | tarjeta sin fix | hallazgo cotizado sin fix | no se pinta tarjeta (RF-246) |
| T-19 | proponer en el chat | tarjeta B1 | se abre el chat, cero escrituras en disco (RF-255) |
| T-20 | borrado | telemetría de un arnés | vuelve al estado 1, sin agregados huérfanos (RF-283) |
| T-21 | a11y de la capa | axe sobre la superficie completa | sin violaciones nuevas; gate a11y en `error` ya configurado |
| T-22 | sin color | captura en escala de grises | los tres estados de detector y las dos severidades distinguibles (RF-280) |
| T-23 | Portafolio sin dato | arnés nunca instrumentado | fila presente, «sin dato», no ordena como 0 (RF-268) |
| T-24 | dos puestos | un arnés en dos puestos | dos filas con costos propios (RF-265) |
| T-25 | forward apagado | daemon recién instalado | forward off; el arnés no puede tocarlo (RF-284) |

## 3 · Qué NO se puede verificar con un test, y cómo se verifica igual

| RF | por qué no hay test | cómo se verifica |
|---|---|---|
| **RF-236** («estimado, no facturación» en la superficie) | un test asserta que el texto existe, no que se **lee** | click-through del gate PARIDAD con captura a 1440×900; el texto tiene que ser legible sin hover en la captura |
| **RF-247 / RF-248** (el copy) | la calidad de una frase no es asertable | revisión del operador en el gate del mockup; el spec fija el copy exacto en `design.md` §Copy y el test asserta el literal |
| **RF-252** (el sesgo va **en contra**) | «en contra» es un juicio sobre el cálculo, no sobre el DOM | revisión por par de cada detector nuevo: para cada fórmula se escribe el supuesto y su dirección en el código, y el test asserta que el campo `direccion_sesgo` existe y no es neutro |
| **RF-254** (score versionado) | testeable el literal, no que se haya **bumpeado** | check de conformance: si cambia la fórmula de un detector y no cambia `score_version`, falla |
| **RF-274** (determinismo de `plugin_id_hash` entre máquinas) | **no verificado** (V7.2), una sola máquina | correr el mismo plugin en otra máquina/usuario. Hasta entonces, el mapeo se aprende localmente y la UI dice «por huella» — que es honesto aunque el hash no sea portable |
| **RF-282** (el contenido no se persiste) | probar «no hay contenido» es probar una ausencia, y un test de structs solo prueba que **ese** camino no lo escribe | **Test de texto sobre el almacén real**: se corre el hook con un prompt que contiene un marcador único y se busca ese marcador en el archivo SQLite. Si aparece, falla. Es la única forma de asertar la ausencia sin confiar en la forma del código |
| **RF-283** (borrado) | el test asserta que la consulta no devuelve nada; no que el byte se fue | test de integración sobre el archivo SQLite real + verificación de que no queda agregado huérfano. La destrucción física del dato no se promete en la UI |
| **RF-284** (el arnés ni egresa ni guarda de más) | no se puede probar la ausencia por muestreo | **dos checks duros de conformance sobre el arnés** (ya en alcance por D12.1): `telemetria-no-egresa` (destino) + su hermano de **proyección de campos** (qué escribe). El primero solo no alcanza — ANEXO H4 |
| **T-21** (a11y) | ~~vitest-browser no corre en background~~ **DEROGADA 2026-07-26** | Corre **headless**: `npx vitest run --project=storybook <archivo>` (~4 s/archivo). `web/vitest.config.ts` declara `headless: true` |

## 4 · Trazabilidad y cierre

- **Capabilities** (doctrina `codigo-traza-a-capability`, R2): el módulo `telemetria/` (ingesta ·
  costeo · detectores · retención/borrado) · `fe-mapa/` (capa Mejora · 4ª tab del inspector) ·
  `fe-portafolio/` (tarjeta) · `http-sse/` (endpoints de consulta). **Cuatro, no una.**
- **Boundaries a tocar:** `telemetria-de-nacimiento.md` → **v2.1** (la corrección de D10 que el
  propio bloque firmado pide) · `codigo-traza-a-capability.md` (capabilities nuevas) ·
  `superficie-local-confinada.md` (sin cambio: se cita, no se enmienda).
- **Checks de conformance del arnés** (D12.1 §Parte 3): `arnes-declara-telemetria` ·
  `arnes-porta-hook-proceso` · `hook-es-fail-open` · `telemetria-no-egresa` · **`hook-proyecta-campos`**
  (el hermano nuevo que pide ANEXO H4 — «no egresa» no cubre «no guarda de más»).
- **Residuo abierto de D8/D14.4:** `fitness/arch_test.go:267` (`TestNoJSONLSchemaParsing`) sigue
  `t.Skip`eado. **Hay que enforcearlo o borrarlo** — hoy simula una protección que no corre. No
  bloquea este spec, pero se cierra en el mismo paquete.
- **No asumir que `transcript_path` es un archivo** (ANEXO H7): en la corrida verificada apuntó al
  **directorio** del proyecto. No lo usamos, pero queda anotado para que nadie lo asuma después.
- **Decisión de producto pendiente, destapada por ANEXO H9:** dónde vive el bloque `env` que
  instrumenta S2. No bloquea las superficies; bloquea el mecanismo de obligación.
- **`PARIDAD.md`** se llena fila por fila durante la implementación.
- **El mockup sigue en iteración 1.** Ningún RF 🎨 se construye antes de su 🧑‍⚖️.

---

## I · Huecos del mockup (12) — para la iteración 2

> No se inventa nada acá. Cada hueco es algo que el spec necesita y el dibujo no resuelve, con la
> resolución propuesta.

| # | hueco | por qué importa | resolución propuesta |
|---|---|---|---|
| **H-1** | **Cómo se llega a la tarjeta de punto de mejora.** §4 la dibuja como panel suelto. `propuesta-mockup.md:74` decía «al tocar una marca de fuga (o desde una lista lateral)»; el dibujo no muestra ni el gesto ni dónde vive la lista | Sin esto, la pieza central del entregable no tiene puerta de entrada | Lista de puntos de mejora **debajo del canvas**, ordenada por ahorro contrafactual descendente. La marca de fuga de la caja lleva el foco a su tarjeta. **No** puede ser un botón dentro del nodo: el nodo YA es un `<button>` (`arnes-node.tsx:52`) y anidar controles es ilegal |
| **H-2** | **Estado vacío de la sección de puntos de mejora en el Mapa.** El Portafolio tiene `✓ sin fugas detectadas` (`:584`); el Mapa no tiene equivalente | Sin él, «no hay tarjetas» se lee como «la sección está rota» | Copy propio: «Hay datos y ningún punto de mejora que pase el corte. Los seis detectores corrieron; ninguno encontró fuga cotizable.» + la lista de detectores que corrieron |
| **H-3** | **Puente Portafolio → Mapa.** La tabla (`:571-593`) no dibuja qué hace la fila al activarse | La frase objetivo cruza las dos superficies; sin puente son dos pantallas sueltas | La fila abre el Mapa de ese arnés **con la capa Mejora ya activa** y la caja del punto de mejora seleccionada. Reusa «Abrir en Mapa» del Portafolio (GAP-1, ya existe) + un parámetro de capa |
| **H-4** | **Destino de «ver el cálculo»** (`:432`) | Es el mecanismo de auditabilidad de A1; hoy es un botón sin destino | Despliegue en línea dentro de la misma tarjeta (no modal): la fórmula con los números reemplazados, la ventana y las corridas contadas |
| **H-5** | **Dónde vive «borrar la telemetría de este arnés»** (`:640`) | Está dibujado como texto dentro del panel didáctico de estados, no como superficie real. Sin ubicación no es construible | En la **franja de contexto**, detrás de un enlace «qué guardamos» que abre el detalle con la retención y la acción de borrado, con confirmación |
| **H-6** | **Estados de transporte.** El mockup dibuja 7 estados de *dato* y ninguno de *transporte*: daemon caído, consulta lenta, error | El Portafolio ya los tiene (`portafolio-list.tsx:280-307`); el Mapa quedaría sin ellos | Tres estados en la franja y en el canvas: cargando (skeleton en las cifras, la geografía se pinta igual), error con motivo y reintento, y «el daemon no respondió» distinguible de «no hay datos» |
| **H-7** | **Indicador del forward del daemon.** D13 lo exige *«con indicador visible en la UI»* y no está dibujado en ninguna superficie | Sin indicador, un forward encendido es invisible — que es exactamente el riesgo de PII de D15.2 | Chip en la franja de contexto **solo cuando está encendido**, nombrando el destino. Fuera de alcance su pantalla de configuración; el indicador entra |
| **H-8** | **Unidad de la barra de cobertura.** El `aria-label` (`:318`) dice «12 exactas, 3 por huella, 2 sin dato» = 17 unidades, y el contador de al lado dice «61 corridas · 12 sesiones · 4 cajas». 17 no es ninguno de los tres | Una barra de cobertura cuya unidad no se sabe no cubre nada | Declarar la unidad en el texto: «12 de 17 corridas con atribución exacta». Y que los números del mockup cierren en la iteración 2 |
| **H-9** | **Ventana sin corridas.** El selector ofrece 7/30/todo (`:310`) pero no se dibuja qué pasa cuando la ventana elegida no contiene ninguna corrida | «Nunca corrió» y «no corrió en estos 7 días» son cosas distintas y el estado 1 solo cubre la primera | Estado propio: «Sin corridas en los últimos 7 días. La última fue el <fecha>.» + atajo para ampliar la ventana |
| **H-10** | **Desglose por MCP dentro de la caja.** El nodo MCP dice «el costo va en la caja que lo llama» (`:404`), y la caja no muestra ese desglose | Se promete una atribución que no se puede ver | Declarar explícitamente que **no** se desglosa en el MVP (es B13, fuera de alcance) y ajustar el copy del nodo a «el costo del MCP está incluido en la caja que lo llama — todavía no se desglosa» |
| **H-11** | **Franja de contexto en pantalla angosta.** `.cov` usa `margin-left:auto` (`:115`): al envolver, la barra de cobertura salta de línea **sin rótulo propio** | Queda una barra tricolor huérfana | Rótulo «cobertura» explícito, visible siempre, no solo cuando hay espacio |
| **H-12** | **El Portafolio no lleva el disclaimer ni la cobertura.** El Mapa sí (`:316`, `:317-322`); la tabla del Portafolio muestra USD pelado | Un número sin la etiqueta «estimado» se lee como facturación — justo lo que D10 prohíbe | Una línea de pie en la tabla con el disclaimer, y la calidad de atribución por fila con la misma marca de duda del canvas (RF-242) |
| **H-13** | **El mockup distingue DOS niveles de confianza, y son cuatro.** Dibuja `exacta` (`:345`) y `por huella` (`:390-391`); la barra de cobertura tiene tres segmentos (`:319`) y el texto nombra tres categorías (`:321`). `por-proceso` no está dibujado en ninguna parte, y el anexo lo hizo real vía `cwd` | Cuatro cosas distintas que se pintan como dos: un número deducido del **directorio** no vale lo mismo que uno deducido de la **huella del arnés**, y el operador tiene que poder distinguirlos antes de actuar sobre la cifra | Cuatro segmentos en la barra (RF-237) y cuatro copys en la marca de duda (RF-242), con la advertencia propia de `por-proceso`: si en ese directorio corre más de un arnés, el número los mezcla. Lo resuelve `design.md` §Estados |
| **H-14** | **El copy del estado 7 promete sobre el canal, no sobre lo guardado.** «Nada de tu cuenta. Nada del contenido.» (`:640`) era exacto cuando el único canal era OTel, que redacta prompts y respuestas. Con el canal de hooks **el contenido llega en claro** (ANEXO H4) | La promesa sigue siendo cierta —descartamos— pero la frase, tal cual, describe lo que *llega*, no lo que *se guarda*. La diferencia importa el día que alguien audite | Segunda línea en el copy que diga que lo que llega con contenido se descarta antes de escribirse, y una lista de los campos que sí se guardan, por nombre (RF-275). Copy exacto en `design.md` §Copy |

---

## J · Contradicciones entre el mockup y las decisiones firmadas (9)

> Cada una con su veredicto. Donde gana el spec, el mockup se corrige en la iteración 2.

| # | contradicción | veredicto |
|---|---|---|
| **J-1** | **P1 dibujado como disponible, declarado parcial en la misma página.** El canvas marca `3 de 4 rechazos en gate` (`:377`) y la tarjeta §4 #2 lo usa como el caso insignia del join (`:447-459`), pero la propia tabla de traza dice «**parcial — el hook de gate no existe todavía**» (`:663`) | **Gana la traza.** P1 es uno de los 6 firmados (D16.1) y es la mitad del join, así que **entra** — pero su hook es trabajo del paquete, no señal existente. El spec lo asume construible; si no se construye, P1 nace en estado degradado con motivo (RF-263), nunca dibujado como si tuviera dato |
| **J-2** | **La tarjeta insignia usa el único detector que no aplica en S2.** §4 #1 (`:424-443`) es B1, y B1 está `❌` para S2 en D16.1 | **Gana D16.1.** La tarjeta tiene que declararse **S1-only**: en una instalación mayormente S2 el ejemplo insignia no existe. El estado 3 (`:618-622`) lo dice a nivel *detector*; falta decirlo a nivel *tarjeta* |
| **J-3** | **Dos ventanas conviviendo.** La franja fija la ventana en **tiempo** (7/30/todo, `:310`); el inspector la fija en **corridas** («últimas 14 corridas», `:500`) | **Gana la franja.** Una sola ventana, la de la capa, y el inspector la hereda (RF-264). La cuenta de corridas se muestra como *denominador*, no como ventana |
| **J-4** | **`rol` vs `puesto` vs `reporta a`.** El mockup dice `rol` en la barra vigente (`:263`), `puesto` en el Portafolio (`:575`); el componente real emite `reporta a` (`map-bar.tsx:79`); el vocabulario firmado del paquete es «arnés × empresa × **puesto**» (D9.8/D12.2) | **Gana «puesto»** en las superficies nuevas, por ser el término de la decisión firmada. **No se renombra** el `MetaChip` vigente en este paquete (sería una segunda desviación de baseline); queda anotado como deuda de vocabulario |
| **J-5** | **D17.3 firmó `[Aplicar]`; el mockup dibuja `[Proponerlo en el chat]`** (`:442`) | **Gana el mockup** — el copy nuevo describe mejor lo que hace, y el comportamiento firmado es idéntico. Pero **es un cambio de etiqueta sobre una decisión firmada** y se declara acá, igual que se declaró el renombre del slot |
| **J-6** | **El estado 7 fija «Retención 90 días»** (`:640`); D15.3 firmó «TTL por default» **sin número** | **Ninguna gana: falta la decisión.** El número 90 no está firmado. Se marca como valor **PROPUESTO** hasta que el operador lo fije; el spec exige que el TTL vigente sea consultable (RF-283), no que valga 90 |
| **J-7** | **El estado 4 promete «otro runtime»** (`:623-627`) cuando el único adaptador verificado es Claude Code (V7.5: «los otros 5 runtimes — solo se probó Claude Code») | **Se sostiene como estado previsto, declarado.** El esquema es multi-runtime por D7/D9.2, pero el MVP entrega **un** adaptador. La UI tiene que poder mostrar el estado; el paquete **no** promete un segundo runtime |
| **J-8** | **El encabezado del carril muestra dinero donde el componente vigente muestra el contador de nodos.** El mockup pone `USD 1,92` en `.lane-hd` (`:339`); `lane.tsx:29` pinta ahí `<span className="count">` | **Superset, no sustitución.** El contador se conserva y el total de la fase se suma al lado (RF-244). Quitarlo sería violar BR-M16 |
| **J-9** | **La esquina de badges del nodo ya está ocupada dos veces.** `.caja-badge` y `.prop-badge` comparten `top:8px; right:8px` en `map.css:310` y `:329`; la capa Mejora agrega marcas al mismo nodo | **No se agrega un tercer badge absoluto.** Las marcas nuevas (cifra, fuga, participación) van **en el flujo** de la tarjeta y al pie, como las dibuja el mockup (`:345-347`). Se resuelve en `design.md` §2 |
| **J-10** | **S2 tiene DOS modos, y D16.1 firmó suponiendo uno solo.** El bloque `env` de los settings del proyecto enciende la telemetría completa (ANEXO H9, verificado en tres variantes con control de puerto) ⇒ en S2 *instrumentado* llega la misma señal que en S1, **dinero incluido, y B1 aplica**. D16.1 marca B1 con `❌` para S2, y el estado 3 del mockup (`:618-622`) dibuja un único S2 degradado | **No se relitiga la firma: se declara la precisión que falta.** El spec **mantiene** RF-271 tal cual —B1 apagado con motivo cuando no hay `result` del stream-json— porque ese es el S2 que D16.1 firmó. Lo que se agrega es que la UI debe **distinguir los dos modos**: «corrió fuera de ArnesIA, sin instrumentar» ≠ «corrió fuera de ArnesIA, instrumentado». Son dos niveles de dato, no el mismo con otro nombre. **Dónde vive el bloque `env` es decisión de producto pendiente** (repo del arnés = sin fricción · proyecto del usuario = escribir settings de un tercero, choca con A8 y con el guardrail vigente ⇒ consentimiento explícito). Y sigue **sin verificar** si un plugin puede aportar `env` — si pudiera, el arnés se instrumenta solo al instalarse |
