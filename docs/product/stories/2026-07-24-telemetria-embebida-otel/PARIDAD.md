# PARIDAD — Tramo B (la superficie de la capa «Mejora»)

> `tipo: paridad` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> Tickets **T28–T37** de [`plan-desarrollo.md`](plan-desarrollo.md), habilitados por el
> 🧑‍⚖️ del mockup (iteración 2) y por **D23**.
> Par de [`mockup-capa-mejora.html`](mockup-capa-mejora.html) (iteración 2, FIRMADA),
> [`design.md`](design.md) y [`plan-storybook.md`](plan-storybook.md).
>
> **Esta hoja NO firma nada.** Es la matriz que el gate humano recorre.

## 0 · Cifras, generadas

| | |
|---|---|
| Tickets cerrados | **10 de 10** (T28…T37), un commit por ticket, cada uno verde antes de abrir el siguiente |
| Stories nuevas | **131** en **17 archivos** (11 nuevos · 6 supersets de archivos firmados) — el plan pedía 125; las 6 de más están declaradas en §3 |
| Story-tests totales | `369` (antes del paquete: `238`) |
| Tests de tabla (`unit`) | `107` (antes: `85`) — 22 nuevos en `entities/telemetria/model/selectors.test.ts` |
| Suite completa | **472 / 476 pasando.** Los **4 rojos son preexistentes y ajenos**: el bug de contraste de `.text-warn` en `new-session-picker.stories.tsx`, que ya estaba en el `BACKLOG.md` |
| `npm run verify` | verde (tsc · biome · depcruise · steiger · stylelint) |
| `go test ./docs/architecture/fitness/...` | verde (incluye R1 · R1-símbolo · R2 · R4) |

> ⚠️ **La línea base que el encargo declaraba (320/323, «3 rojos») estaba desactualizada: eran
> 319/323 y CUATRO rojos**, los cuatro en `new-session-picker.stories.tsx`. Verificado antes de
> tocar una línea. No se arreglaron ni se contaron como propios.

---

## 1 · Matriz mockup ↔ componente ↔ story = test ↔ RF

### §2 del mockup — La barra con la capa activa y la franja de contexto

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| slot `Mejora` encendido, 4 tabs | `widgets/map-canvas/model/layers.ts#LAYERS` · `ui/map-bar.tsx` | `CapaMejoraDisponible` · `CapaMejoraActiva` | RF-232 |
| tooltips honestos de los slots apagados | `layers.ts#LayerDef.motivo` | `MotivosHonestos` · `MotivoAccesible` | RF-233 · RF-276 |
| `select` de ventana, primero en la línea | `ui/franja-mejora.tsx` | `Reposo` · `VentanaCambia` | RF-234 |
| total + denominador que cierra (61/58/12/4) | ídem | `TotalConDenominador` | RF-235 |
| disclaimer «estimado…», pegado al total | ídem | `DisclaimerEnSuperficie` · `ReposoDark` | RF-236 |
| barra de cobertura de 4 niveles + rótulo | `entities/telemetria/ui/barra-cobertura.tsx#BarraCobertura` | `CuatroSegmentos` · `CategoriaEnCeroNoOcupaLugar` · `CoberturaCompleta` · `SinCorridas` · `RotuloVisibleSiempre` · `CoberturaCuatroNiveles` | RF-237 · RF-279 · H-8 · H-11 |
| chip de reenvío externo encendido | `franja-mejora.tsx` | `ReenvioEncendido` · `ReenvioApagadoNoSeDibuja` | H-7 · D13 |
| «Nada de tu cuenta…» + enlace | ídem | `ResumenQueGuardamos` | RF-275 · H-5 |

### §3 del mockup — El canvas: misma geografía, marcas nuevas en el flujo

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| cifra + % + barra en la caja | `entities/arnes/ui/arnes-node.tsx` (props primitivas, D18) | `MejoraCifraExacta` | RF-238 · RF-239 · RF-240 |
| marca de confianza en el nodo | ídem, alimentado desde `entities/telemetria` por el widget | `MejoraPorHuella` · `MejoraPorProceso` · **`CopyConfianzaEsUnaSola`** | RF-242 · RF-274 · H-13 |
| los 4 casos de confianza, standalone | `entities/telemetria/ui/marca-confianza.tsx` | `Exacta` · `PorHash` · `PorProceso` · `SinDato` · `LosCuatroJuntos` | RF-242 · RF-280 |
| marca de fuga nombrada | `arnes-node.tsx` `.mej-fuga` | `MejoraConMarcaDeFuga` · `MejoraDosDetectores` | RF-241 |
| «sin dato atribuible» por clase (5 motivos) | `entities/telemetria/model/selectors.ts#MOTIVO_SIN_DATO` | `SinDatoSubagente` · `SinDatoRegla` · `SinDatoMcp` · `SinDatoHook` · `SinDatoResto` | RF-243 · D19 · H-10 |
| caja sin corridas en la ventana | `arnes-node.tsx` | `MejoraCajaSinCorridas` | RF-240 · RF-243 |
| total de la fase en el encabezado | `widgets/map-canvas/ui/lane.tsx` | `LaneConTotalMejora` · `LaneSinDato` | RF-244 · J-8 |
| la geografía NO se mueve | `ui/map-canvas.tsx` | `CapaMejoraSupersetGeografia` · `CapaMejoraSoloCajasLlevanCifra` · `ConmutarNoPierdeSeleccion` · `EstructuraIntacta` | RF-245 · T-16 · T-17 · BR-M16 |
| el dinero, formateado en un solo lugar | `entities/telemetria/ui/cifra-usd.tsx#CifraUsd` | `Estandar` · `SeparadorDosDecimales` · `MenorAlCentavo` · `Ausente` | RF-281 |

### §4 del mockup — La lista de puntos de mejora, debajo del canvas

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| la lista con su encabezado y contador | `ui/puntos-mejora-list.tsx#PuntosMejoraList` | `DosTarjetasOrdenadas` | H-1 · RF-246 |
| la tarjeta insignia B1, completa | `ui/punto-mejora-card.tsx#PuntoMejoraCard` | `CompletaB1Atencion` · `CompletaB1AtencionDark` | RF-247…254 |
| P1 crítica, con `Patrón` en vez de `Umbral` | ídem | `CriticaP1SinUmbral` | RF-250 · RF-257 |
| la fila `Sesgo`, con su dirección | ídem | `SesgoSubestima` · `SesgoSobreestima` · `SinSesgoIdentificado` | RF-252 |
| chip S1-only | ídem | `S1Only` | J-2 |
| «ver el cálculo» desplegado en línea | ídem | `CalculoCerrado` · `CalculoAbierto` | H-4 |
| tarjeta resaltada por su caja | ídem | `Resaltada` | design §5.4 · RF-280 |
| `Descartar` con vuelta atrás | ídem | `DescartarLlamaHandler` | RF-256 |
| `Proponerlo en el chat` **no escribe** | ídem | `ProponerAbreChatNoEscribe` · `ProponerDeshabilitadoFueraDeAlcance` · `NotaAlPie` | RF-255 · BR-M12 · D17.3 |
| severidad legible sin color | ídem | `SeveridadSinColor` · `TitularSinJerga` | RF-247 · RF-257 · RF-280 |
| estado vacío con los 6 detectores | `puntos-mejora-list.tsx` | `VaciaConDatos` | H-2 |
| sin contrafactual no hay tarjeta | ídem | `DescartaSinContrafactual` | RF-246 · A4 |
| carga y error de la lista | ídem | `Cargando` · `ErrorDeConsulta` · `MuchasTarjetas` | design §5.4 · §6 |

### §5 del mockup — Inspector, cuarta tab «Mejora»

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| la 4ª tab con su contrato ARIA | `ui/inspector.tsx` (superset) | `CuatroTabs` · `TabMejoraContratoAria` · `CambioDeNodoVuelveAlResumen` · **`Tabs`** (ver §3) | RF-258 · RF-277 |
| tabla de 6 buckets, la suma cierra | `ui/inspector-mejora.tsx#InspectorMejora` | `Completo` · `CorridaRealMedida` | RF-259 |
| «no aplica» ≠ «midió 0» | ídem | `NoAplicaNoEsCero` · `CeroLegitimo` | RF-260 · RF-286 |
| paridad reportado vs. calculado | ídem | `ParidadCoinciden` · `ParidadDivergen` · `SinCostoDelRuntime` · `CatalogoSinConstruir` | RF-261 · RF-272 |
| el join de la caja | ídem | `JoinCompleto` · `SinSenalDeGate` | RF-262 |
| los 6 detectores + los 7 no medidos | ídem | `SeisDetectoresConEstado` · `DetectorB1ApagadoEnS2` · `DetectorB2Parcial` · `SieteNoMedidos` · `DetectorSinFixPropuesto` | RF-263 · RF-271 · T-09 |
| la ventana es la de la capa | ídem | `VentanaHeredada` | RF-264 · J-3 |
| nodo que no es caja | ídem | `NodoNoCaja` | RF-258 |
| carga y error dentro de la tab | ídem | `Cargando` · `ErrorDeConsulta` | design §5.5 |

### §6 del mockup — La tarjeta del Portafolio

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| `USD/corrida` · `tendencia` · `punto de mejora` | `widgets/portafolio/ui/tabla-mejora-portafolio.tsx` | `ConDato` · `ConDatoDark` | RF-265 · D21 |
| un arnés en dos instalaciones | ídem | `UnArnesDosPuestos` | RF-265 · D20 |
| «puesto sin declarar» | ídem | `SinPuestoDeclarado` | RF-265 · D20 |
| sparkline + texto equivalente | `entities/telemetria/ui/sparkline.tsx` | `EnAlza` · `Estable` · `HaciaLaBaja` · `PocasCorridas` (×2: entity y tabla) | RF-266 · RF-279 |
| «✓ sin fugas» ≠ «sin dato» | `tabla-mejora-portafolio.tsx` | `SinFugas` · `SinDato` | RF-267 · RF-268 · RF-280 |
| orden por costo, sin-dato al final | ídem | `OrdenadaSinDatoAlFinal` | RF-268 |
| pie de tabla + confianza por fila | ídem | `PieConDisclaimerDeEstimacion` | H-12 |
| desbordamiento y scroll accesible | ídem | `NombresLargosNumerosGrandes` | design §6.3 |
| H-3 · el puente al Mapa | ídem + `portafolio-view.tsx` | `AbreElMapaDeEsaInstalacion` | H-3 |
| las 3 celdas en la fila de la lista | `ui/portafolio-list.tsx` (superset) | `ConMejora` · `SinMejoraDomIntacto` | RF-265 · BR-M16 |

### §7 del mockup — Los estados honestos

| # | estado | story = test | RF |
|---|---|---|---|
| 1 | nunca corrió | `Estado1SinDatos` | RF-269 |
| 1b | sin corridas en la ventana | `Estado1bSinCorridasEnVentana` | H-9 · T-15 |
| 2 | cobertura parcial | `Estado2CoberturaParcial` | RF-270 |
| 3 | S2 sin instrumentar | `Estado3S2SinInstrumentar` | RF-271 · J-10 |
| 3b | S2 instrumentado | `Estado3bS2Instrumentado` | RF-271 · ANEXO H9 |
| 4 | otro runtime | `Estado4OtroRuntime` | RF-272 |
| 5 | catálogo viejo | `Estado5CatalogoViejo` | RF-273 |
| 6 | por huella | `PorHash` · `MejoraPorHuella` | RF-274 |
| — | runtime no soportado | `RuntimeNoSoportado` | escenarios A5 |
| B1 | cargando | `Cargando` (franja · lista · inspector) | H-6 |
| B2 | error de consulta | `ErrorDeConsulta` | H-6 |
| B3 | daemon caído, cifras viejas | `DaemonCaido` | H-6 |
| — | promoción de los estados a `shared/ui` | `SkeletonHonesto` · `SkeletonConMedidaPropia` · `ErrorConMotivoReintentable` | H-6 |

### §8 del mockup — Qué guardamos y cómo se borra

| qué dibuja el mockup | componente real | story = test | RF |
|---|---|---|---|
| los 3 bloques del diálogo | `ui/politica-datos-dialog.tsx` | `Reposo` | RF-275 · H-5 · H-14 |
| la lista de campos, **por prop** | ídem | `CamposDesplegados` | RF-275 · RF-282 |
| retención desde la config | ídem | `RetencionDesdeConfig` | RF-283 · J-6 |
| confirmación con alcance | ídem | `Confirmacion` | RF-275 |
| borrando / error / éxito | ídem | `Borrando` · `ErrorDeBorrado` · `Exito` | design §5.7 |
| focus trap y teclado | ídem | `FocoAtrapado` | RF-278 |

---

## 2 · Desviaciones del mockup firmado y de los planes

Ninguna se resolvió en silencio. Cada una tiene su motivo verificado.

| # | qué | por qué | dónde |
|---|---|---|---|
| **D-1** 🔴 | El slot `Tokens` del conmutador pasa a `Mejora` | Desviación de un **baseline firmado** (`mockups/INDEX.md` regla 3). Autorizada por **D17.1**; se declara igual porque cambia superficie vigente | `layers.ts` |
| **D-2** 🔴 | La story firmada `Tabs` (`inspector.stories.tsx`) pasa de `toHaveLength(3)` a `4` | **Única modificación permitida a una story firmada en este paquete** (plan T34). Se cambió en el mismo commit que agrega la tab. Nada más de esa story cambia | `inspector.stories.tsx` |
| **D-3** | `CifraUsd`, no `CifraUSD` (design §1.3 · **D22**) | `biome lint/style/useNamingConvention` con `strictCase: true` (severidad `error`, dentro de `verify` y de CI) rechaza dos mayúsculas consecutivas en PascalCase. **Se renombra el símbolo, no se baja el gate** — la misma lógica con la que D18 rechazó apagar `fsd/no-cross-imports`. El archivo sigue siendo `cifra-usd.tsx` y el puntero de CAP-139 está corregido | contradicción **#14** |
| **D-4** | Props `cifraUsd` / `totalUsd`, no `cifraUSD` / `totalUSD` (**D18**) | Misma regla, también en camelCase | `arnes-node.tsx` · `lane.tsx` |
| **D-5** | Stories renombradas: `ErrorConMotivoYReintentar`→`ErrorConMotivoReintentable` · `ALaBaja`→`HaciaLaBaja` · `CambioDeNodoVuelveAResumen`→`…AlResumen` · `PieConDisclaimerYConfianza`→`…DeEstimacion` · `NombresLargosYNumerosGrandes`→`NombresLargosNumerosGrandes` · `SinMejoraDOMIntacto`→`SinMejoraDomIntacto` | Todas caían en `strictCase` (dos mayúsculas consecutivas; en varias, una `Y` seguida de mayúscula) | varios |
| **D-6** | Story `Error`→`ErrorDeConsulta` (franja · lista · inspector) | `biome lint/suspicious/noShadowRestrictedNames`: `Error` sombrea el global | varios |
| **D-7** | El enlace «qué guardamos» **no usa `--primary` como color de texto** | **Fallo de contraste MEDIDO al construir, no previsto por D21**: axe reportó **2,41 : 1** de `#00b7aa` sobre el fondo de la franja. La afordancia de enlace la carga el subrayado; el acento queda para el foco. Es la misma regla de D21, aplicada a un tercer caso | `mejora.css` |
| **D-8** | El contado de «cobertura parcial» resalta en `--foreground` **negrita**, no con tono `warn` | `plan-storybook.md` §2.8 pedía `toHaveClass` de tono warn; `design.md` §5.2 y D21 dicen `--foreground` porque `--warn` sobre `--card` mide 3,76 : 1 y es la deuda que el BACKLOG ya tiene abierta. **Gana `design.md`**: este paquete no puede agravarla | `franja-mejora.tsx` |
| **D-9** | El denominador de la franja usa el literal de `design.md` §7.2 (`de 61 corridas, 58 con atribución · 12 sesiones · 4 cajas`), no el de `plan-storybook.md` §2.8 (`61 corridas · 12 sesiones · 4 cajas`) | Los dos documentos difieren. `design.md` es el SSoT del copy literal, es posterior (12:07 vs 11:48) y su cabecera declara que la iteración 2 **cerró H-8** con un juego que cierra entre sí | `franja-mejora.stories.tsx` |
| **D-10** | `BarraCobertura` conserva el fixture 12/3/2/1 sobre 18 de `plan-storybook.md` §2.3; la franja usa el juego coherente 44/9/5/3 sobre 61 | El texto lo produce **el mismo componente** a partir de sus props: no hay copy duplicado, hay dos fixtures. El de la entity prueba el contrato con cuatro segmentos no nulos; el de la franja es el que tiene que cerrar con el denominador | dos archivos |
| **D-11** | El bloque `MEDIDO` usa `evidencia/result-envelope.json` (`input 10 · cache_read 17536 · cache_creation 8257 · output 39 · 18 473 micros`), **no** los `21695/4099/35` que cita `plan-storybook.md` §3.2 | Esos números son de **OTRA** de las tres corridas (`INFORME.md` V2, líneas 61-63). El archivo golden versionado y el `api_request` de `logs-run1.json` coinciden **entre sí**. Gana el archivo (regla 4 del plan: los payloads de `evidencia/` son golden files). Los dos hechos que el plan necesita sobreviven intactos: el **cero legítimo** (`ephemeral_5m = 0`) y la **ausencia real** de `razonamiento` | `testing/telemetria.ts` |
| **D-12** | Fixture nuevo `BUCKETS_ILUSTRATIVOS` para el assert aritmético de `Completo` | A dos decimales —que es como se muestra el dinero (RF-281)— los 18 473 micros reales **no suman nada**: cada bucket redondea. Bajar la precisión del formateador para que el assert cierre sería arreglar el termómetro. La **FORMA** sale del bloque MEDIDO (el 0 real, la ausencia real); los montos suman exactamente los `1,92` del copy. `CorridaRealMedida` prueba el inspector contra el payload medido, con asserts de forma | `testing/telemetria.ts` |
| **D-13** | CAP-139 se crea en **T29** y no en T33 | La regla 1 del propio plan (§0) obliga a crear la hoja con el ÚLTIMO símbolo de cada tanda: crearla en T33 dejaría **7 archivos huérfanos** en el commit de T29 (R2), y crearla completa dejaría un puntero colgante a `PuntoMejoraCard` (R1). La hoja crece ticket a ticket | `capa-mejora.yaml` |
| **D-14** | `Skeleton`/`ErrorBody` conservan sus clases `pf-*` al promoverse a `shared/ui` | Es lo que pide «movidos tal cual» y es el precedente sancionado (`FiltroDisclosure` conservó `pf-filtro`). Renombrarlas obligaría a reescribir CSS firmado del Portafolio, que es el cambio de conducta que un refactor de movimiento no puede permitirse. La capa Mejora las estiliza bajo su propio scope | `estado-carga.tsx` · `mejora.css` |
| **D-15** | El CSS de las cuatro piezas de `entities/telemetria` va **sin scope de superficie** (`telemetria.css`) | **D23**: son vocabulario de la plataforma. Scoparlas a `.arnesia-mejora` obligaría a la próxima superficie a reinventarlas. Las clases son distintivas y no colisionan (verificado) | `telemetria.css` |
| **D-16** | 131 stories, no 125 | +6, todas de cobertura de contrato (D23) o de un assert que una sola story no puede hacer — detalle en §3 | — |

---

## 3 · Las 6 stories de más

| story | archivo | por qué existe |
|---|---|---|
| `SkeletonConMedidaPropia` | `shared/ui/estado-carga.stories.tsx` | `filas`/`altura`/`data` son API nueva de la pieza promovida, y `Cargando` de la lista de puntos depende de ella (`[data-skeleton='mejora']`). D23: la pieza se documenta por su contrato, no por un caso |
| `ReenvioApagadoNoSeDibuja` | `franja-mejora.stories.tsx` | `plan-storybook.md` §2.8 pedía «assert espejo dentro de esta story sobre el default». **Una story renderiza UN set de args**: el espejo no se puede hacer sin una segunda story |
| `CoberturaCompletaEnLaFranja` | ídem | La integración del caso «todo exacto» en la franja, con el denominador real de 61 |
| `CorridaRealMedida` | `inspector-mejora.stories.tsx` | Prueba el inspector contra el payload **MEDIDO** (18 473 micros), y no solo contra el juego ilustrativo. Assert de FORMA, no de suma (ver D-12) |
| `AbreElMapaDeEsaInstalacion` | `tabla-mejora-portafolio.stories.tsx` | **H-3**, el puente Portafolio → Mapa, que el plan enumera en T37 pero no tenía story propia |
| `ConDatoDark` | ídem | Ya estaba prevista en D21 ítem 3 (la que llevaba el total de 124 a 125); se cuenta acá por completitud |

---

## 4 · Lo que este tramo NO puede mostrar honestamente, y lo dice

Sale de [`auditoria-tramo-a.md`](auditoria-tramo-a.md) §UX. **La UI está construida; el dato que
la alimenta tiene huecos declarados.** Ninguno se tapó.

| # | hueco del backend | cómo se comporta la UI |
|---|---|---|
| 1 | **`FilaPortafolio` no tiene ningún campo de serie** (auditoría §UX ítem 6 · M10: `Tendencia` existe en el contrato y **nunca se calcula**) | La columna `tendencia` dirá **«pocas corridas para una tendencia»** contra el wire real, siempre. `Sparkline` acepta `puntos` opcional y su ausencia se DICE; no se dibuja una línea inventada. Documentado en el tipo (`types.ts#FilaPortafolio.serie`) y en el componente |
| 2 | **`Corridas` es 0 fuera de S1** ⇒ `costo_por_corrida` es **siempre `null`** en `s2-instrumentado` (auditoría §UX ítem 7) | La celda dirá `sin dato` para todo lo que no corra dentro de ArnesIA — que es la mayoría. Es honesto, y es lo que el operador tiene que decidir qué hacer |
| 3 | **T22 abierto**: no hay eventos de gate del daemon | La cláusula «falla el gate 3 de cada 4 veces» **no se puede decir con dato real**. `SinSenalDeGate` es el estado que la UI muestra: la fila dice el motivo, no un `0` |
| 4 | **`GastoCaja.Nombre = CajaID`** (auditoría §UX ítem 3) | El canvas pinta el nombre del NODO (que sí es humano) y no el del wire, así que este hueco no se ve en el Mapa; sí se vería en cualquier superficie que liste cajas por el wire |
| 5 | **A2 — «detalle purgado, resumen conservado» no existe** | **No se dibujó ese estado.** El diálogo de borrado dice «Este arnés vuelve a estar sin datos de telemetría», que es lo que el backend realmente produce |
| 6 | **M1 — B6 dispara siempre en `s2-instrumentado`** | La lista de mejoras nunca estará vacía contra ese escenario, y su primer ítem será un falso positivo estructural. `VaciaConDatos` existe y es correcta, pero **hoy no se va a alcanzar en S2** |
| 7 | **M2 — B1 y B3 ponen conteos de tokens en campos `micros`** | Cualquier cifra de dinero de esas dos tarjetas está mal **en el backend**. La UI la formatea correctamente; el número es el que el dominio manda |
| 8 | 🔴 **`RespuestaMejoras` no trae la lista de detectores que corrieron y no encontraron nada.** Trae `no_aplican` (no pudieron correr) y `no_medidos` (fuera del MVP); un detector que aplicó y salió limpio **no está en ninguna** | H-2 pide exactamente esa lista. Alimentar el vacío con `no_medidos` los pintaría como «sin hallazgos», que es **la mentira que RF-263 prohíbe por su nombre**. Encontrado al cablear el transporte: el vacío muestra su copy **sin enumerar a nadie** hasta que el wire mande la lista. Declarado en el código (`workspace-stage.tsx`) |
| 9 | 🔴 **`DetalleCaja` no trae el costo POR BUCKET.** Trae `Tokens` (los seis punteros) y `Paridad` (los dos totales), pero no un desglose de dinero por bucket | La columna `USD` de la tabla del inspector viaja **`null` — «no aplica»** contra el wire real. Calcularla en el FE sería costear en la UI, que es justo lo que `design.md` §1.3 prohíbe («entities presenta; no calcula»). **El assert de «la suma cierra» (`Completo`) corre contra fixture, no contra el wire**, y no puede correr contra el wire hasta que este campo exista |

---

## 5 · Lo que quedó SIN verificar

| # | qué | por qué |
|---|---|---|
| V-1 | **La cadena E2E contra el daemon vivo** (plan §5.2-5.6: HOME de prueba, puerto efímero, `claude` real, app instalada) | Es el **Tramo C (T38-T39)**, fuera del encargo de este tramo. Todo lo construido acá se verificó contra fixtures rotuladas y contra los shapes reales del wire leídos en `internal/domain/telemetria_vistas.go`, **no** contra respuestas HTTP en vivo |
| V-2 | **El aspecto real de la capa en el navegador** | No se levantó el daemon ni se sacó captura. Las stories corren en Chromium headless y assertan DOM, contraste (axe) y `getComputedStyle`, pero **nadie miró la pantalla** |
| V-3 | **`fitView` con los nodos ~26 px más altos** (design §6.2) | Es comportamiento de layout con el canvas montado a tamaño real; las stories lo montan a 620 px de alto. El razonamiento está escrito (overview-first, RF-50) pero no medido |
| V-4 | **El tema oscuro más allá de las 3 stories espejo** | `ReposoDark`, `CompletaB1AtencionDark` y `ConDatoDark` cubren franja, tarjeta y Portafolio. El inspector, la lista y el diálogo **solo tienen gate en claro** |
| V-5 | **El wire real de `puntos`** | El FE tipa `PuntoMejora` con campos que el Go **no manda con esos nombres** (`contrafactual` como prosa, `patron`, `fix_codigo`, `calculo`, `caja_nombre`, `id`, `solo_s1`, `confianza_detalle`). El dominio Go tiene `ContrafactualMicros`/`DiferenciaMicros`/`Umbral`/`Sesgo`/`Fix`. **Falta el adaptador que arme la prosa** — hoy la composición está escrita contra el tipo del FE y ningún test la ata al JSON real |
| V-7 | **La suite tiene flakiness intermitente bajo carga** — no introducida por este paquete | En 7 corridas completas seguidas, 5 cerraron **exactamente** en los 4 rojos preexistentes y 2 sumaron fallos intermitentes. Los archivos afectados incluyen `shared/ui/buscador-filtro.stories.tsx` y `entities/marketplace/ui/chips.stories.tsx`, **que este paquete nunca tocó** — así que es del runner (browser mode en paralelo), no del diff. Ninguno de los archivos nuevos falló dos veces seguidas, y los 4 archivos nuevos corridos en aislamiento pasan 3/3. **Vale la pena abrirlo en el BACKLOG**: una suite que falla 2 de cada 7 veces por razones de infraestructura entrena a re-correr, y eso es exactamente lo que hace que un rojo real pase desapercibido |
| V-6 | **La 4ª tab contra el wire real** | Está cableada (`workspace-stage.tsx` pide `GET …/cajas/{cajaId}` al seleccionar una caja con la capa encendida), pero la columna USD por bucket llega `null` (§4 ítem 9) y el join llega con `null` en las tres filas de proceso (T22 abierto). **Lo que se ve contra el wire real es una tabla de tokens y una paridad**, no el desglose completo del mockup |

> 🔴 **V-5 es el hueco más grande de este tramo y no se puede cerrar desde el FE.** El
> `PuntoDeMejora` de Go entrega piezas (`ContrafactualMicros`, `DiferenciaMicros`) y la tarjeta
> necesita **frases con su unidad declarada** (RF-249). O el backend arma la prosa —donde están
> los números y el sesgo— o hay que decidir que la arma el FE, que es exactamente el cálculo que
> `design.md` §1.3 prohíbe en `entities`. **Es una decisión de contrato, no de píxeles.**

---

## 6 · Comandos de la demostración

```bash
cd web && npm run verify                 # tsc · biome · depcruise · steiger · stylelint → verde
cd web && npx vitest run                 # 472 / 476 (4 rojos preexistentes, new-session-picker)
cd web && npx vitest run --project=storybook   # 365 / 369
cd web && npx vitest run --project=unit        # 107 / 107
go test ./docs/architecture/fitness/...  # verde (R1 · R1-símbolo · R2 · R4)
```

---

## 7 · 🧑‍⚖️ Gate humano — PENDIENTE

Este tramo **no está firmado**. Para firmarlo hay que mirar, como mínimo:

1. **D-1 y D-2**, las dos desviaciones marcadas 🔴 sobre superficie firmada.
2. **§4 completo**: siete huecos del backend que la UI declara en pantalla. Si alguno no es
   aceptable como estado visible, cambia el diseño, no el código.
3. **V-5**: el contrato de `PuntoDeMejora` entre Go y el FE. Bloquea el E2E real de la tarjeta.
4. **El TTL de 90 días** sigue **PROPUESTO** (J-6 · parada P2). La UI lo lee de la config y lo
   rotula como tal; el número lo pone el operador.
