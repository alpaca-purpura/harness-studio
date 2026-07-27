# PARIDAD — Tramo B (la superficie de la capa «Mejora»)

> `tipo: paridad` · paquete `2026-07-24-telemetria-embebida-otel` · 2026-07-26.
> Tickets **T28–T37** de [`plan-desarrollo.md`](plan-desarrollo.md), habilitados por el
> 🧑‍⚖️ del mockup (iteración 2) y por **D23**.
> Par de [`mockup-capa-mejora.html`](mockup-capa-mejora.html) (iteración 2, FIRMADA),
> [`design.md`](design.md) y [`plan-storybook.md`](plan-storybook.md).
>
> **Esta hoja NO firma nada.** Es la matriz que el gate humano recorre.
>
> 🔴 **Leé §9 antes que la matriz.** La auditoría independiente del Tramo B
> ([`auditoria-tramo-b.md`](auditoria-tramo-b.md)) encontró **4 críticos** de la misma clase que
> D24 y una objeción de proceso a esta hoja: *«nueve filas acreditan “story verde” y el gate
> humano las lee como “la superficie hace esto”»*. Los cuatro críticos están corregidos; la
> objeción está atendida con la **columna «¿alcanzable?»** que ahora lleva cada fila de §1.

## 0-bis · Ejecución de D26 (2026-07-27) — las 6 decisiones, aplicadas

El gate se firmó **junto con** las seis decisiones (§7). Esto es lo que cambió al ejecutarlas, y
**la columna «¿alcanzable?» de §1 ya no tiene ningún `no`**.

| # | decisión | qué cambió |
|---|---|---|
| 1-A | B1 y B3 | los buckets de cache se **cotizan** (antes eran conteos de tokens en campos `micros`, defecto M2) y el contrafactual de B1 compara contra **una sola escritura a 1 h**, que es de donde sale el ahorro. `ScoreVersionMVP` 1 → **2** (A7). Sin catálogo, los dos pasan a `no_aplica` con motivo |
| 2-A | A20 | el bloque `env` viaja en `dogfood/dev-full-cycle/.claude/settings.json`, y `VariablesTelemetria` es la **fuente única** de los dos caminos (spawn y settings). No se escribe en árbol ajeno |
| 3 | TTL | **90 días, firmado**: `retencion_propuesta` salió del wire, del CLI y de la UI |
| 4-A | las 10 filas | cableadas. Hizo falta wire nuevo: `ultima_corrida`, `runtime`/`runtime_soportado`, `cajas`, `descartados`, y **dos endpoints de descarte** que no existían |
| 5-B | A-4 | el `DELETE` acepta ventana y **borra lo que la confirmación declara**; una ventana ilegible da 400 |
| 6 | firma | transcrita en §7, con alcance acotado |

Y el contrato dejó de ser deuda: **las 10 rutas de `/api/telemetria/*` están declaradas en
`openapi.yaml` (0.8.0-telemetria)** y sus exenciones salieron del allowlist — que es exactamente
lo que la exención prometía, «declara su contrato al cerrar».

**Dos huecos nuevos que la ejecución destapó** —ninguno estaba en la lista, los dos corregidos—:

1. 🔴 **`cajas` no existía en el wire.** El denominador de la franja («… · 4 cajas»), que es copy
   firmado, se pintaba desde un campo que el FE tipaba y el Go nunca mandaba. **Misma clase exacta
   que V-5**, y lo cazó el candado de contrato al extenderse a `ResumenTelemetria`.
2. **El gate de arquitectura se ponía rojo por tener trabajo en curso al lado**: go-arch-lint medía
   los `.go` de las copias del árbol en `.claude/worktrees/` como si fueran el producto. Excluido —
   un gate que se pone rojo por algo que no es el código que se mide entrena a ignorarlo.

**Gates al cerrar (2026-07-27):** `go test ./... -race` verde · fitness verde ·
`conformance --todo` **323 checks, fail 0** · `npm run verify` verde · `npx vitest run`
**524 / 524**. ⚠️ La flakiness de V-7 sigue viva y se manifestó: dos corridas de la suite completa
dieron 5 y 8 rojos dispersos —en archivos que este trabajo no tocó— y las mismas pasan en
aislamiento y en la corrida siguiente. **No está arreglada; está declarada.**

## 0 · Cifras, generadas

| | |
|---|---|
| Tickets cerrados | **10 de 10** (T28…T37), un commit por ticket, cada uno verde antes de abrir el siguiente |
| Stories nuevas | **142** en **18 archivos** (12 nuevos · 6 supersets de archivos firmados) — el plan pedía 125; las 13 de más están declaradas en §3 y §8 |
| Story-tests totales | `380` (antes del paquete: `238`) |
| Tests de tabla (`unit`) | `125` (antes: `85`) — 22 en `entities/telemetria` + 18 en `widgets/map-canvas/model/capa-mejora.test.ts` |
| Suite completa | **501 / 505 pasando.** Los **4 rojos son preexistentes y ajenos**: el bug de contraste de `.text-warn` en `new-session-picker.stories.tsx`, que ya estaba en el `BACKLOG.md` |
| `npm run verify` | verde (tsc · biome · depcruise · steiger · stylelint) |
| `go test ./docs/architecture/fitness/...` | verde (incluye R1 · R1-símbolo · R2 · R4) |

> ⚠️ **La línea base que el encargo declaraba (320/323, «3 rojos») estaba desactualizada: eran
> 319/323 y CUATRO rojos**, los cuatro en `new-session-picker.stories.tsx`. Verificado antes de
> tocar una línea. No se arreglaron ni se contaron como propios.

---

## 1 · Matriz mockup ↔ componente ↔ story = test ↔ RF

> **Cómo leer la columna «¿alcanzable?»** — la objeción de proceso de la auditoría. Una story
> verde prueba que **el componente hace lo que dice con esas props**; NO prueba que la app llegue
> a pasárselas. Los valores son:
> **`sí`** = un composition-root la alimenta con dato real ·
> **`no`** = la prop existe, está storiada y **ningún composition-root la pasa** (estado muerto) ·
> **`n/a`** = pieza interna que no depende del transporte.

### §2 del mockup — La barra con la capa activa y la franja de contexto

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| slot `Mejora` encendido, 4 tabs | `widgets/map-canvas/model/layers.ts#LAYERS` · `ui/map-bar.tsx` | `CapaMejoraDisponible` · `CapaMejoraActiva` | RF-232 | sí |
| tooltips honestos de los slots apagados | `layers.ts#LayerDef.motivo` | `MotivosHonestos` · `MotivoAccesible` | RF-233 · RF-276 | sí |
| `select` de ventana, primero en la línea | `ui/franja-mejora.tsx` | `Reposo` · `VentanaCambia` | RF-234 | sí |
| total + denominador que cierra (61/58/12/4) | ídem | `TotalConDenominador` | RF-235 | sí |
| disclaimer «estimado…», pegado al total | ídem | `DisclaimerEnSuperficie` · `ReposoDark` | RF-236 | sí |
| barra de cobertura de 4 niveles + rótulo | `entities/telemetria/ui/barra-cobertura.tsx#BarraCobertura` | `CuatroSegmentos` · `CategoriaEnCeroNoOcupaLugar` · `CoberturaCompleta` · `SinCorridas` · `RotuloVisibleSiempre` · `CoberturaCuatroNiveles` | RF-237 · RF-279 · H-8 · H-11 | sí |
| chip de reenvío externo encendido | `franja-mejora.tsx` | `ReenvioEncendido` · `ReenvioApagadoNoSeDibuja` | H-7 · D13 | sí · **desde este paquete** (A-2) |
| «Nada de tu cuenta…» + enlace | ídem | `ResumenQueGuardamos` | RF-275 · H-5 | sí |

### §3 del mockup — El canvas: misma geografía, marcas nuevas en el flujo

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| cifra + % + barra en la caja | `entities/arnes/ui/arnes-node.tsx` (props primitivas, D18) | `MejoraCifraExacta` | RF-238 · RF-239 · RF-240 | sí |
| marca de confianza en el nodo | ídem, alimentado desde `entities/telemetria` por el widget | `MejoraPorHuella` · `MejoraPorProceso` · **`CopyConfianzaEsUnaSola`** | RF-242 · RF-274 · H-13 | sí |
| los 4 casos de confianza, standalone | `entities/telemetria/ui/marca-confianza.tsx` | `Exacta` · `PorHash` · `PorProceso` · `SinDato` · `LosCuatroJuntos` | RF-242 · RF-280 | n/a |
| marca de fuga nombrada | `arnes-node.tsx` `.mej-fuga` | `MejoraConMarcaDeFuga` · `MejoraDosDetectores` | RF-241 | sí |
| «sin dato atribuible» por clase (5 motivos) | `entities/telemetria/model/selectors.ts#MOTIVO_SIN_DATO` | `SinDatoSubagente` · `SinDatoRegla` · `SinDatoMcp` · `SinDatoHook` · `SinDatoResto` | RF-243 · D19 · H-10 | sí |
| caja sin corridas en la ventana | `arnes-node.tsx` | `MejoraCajaSinCorridas` | RF-240 · RF-243 | sí |
| total de la fase en el encabezado | `widgets/map-canvas/ui/lane.tsx` | `LaneConTotalMejora` · `LaneSinDato` | RF-244 · J-8 | sí |
| la geografía NO se mueve | `ui/map-canvas.tsx` | `CapaMejoraSupersetGeografia` · `CapaMejoraSoloCajasLlevanCifra` · `ConmutarNoPierdeSeleccion` · `EstructuraIntacta` | RF-245 · T-16 · T-17 · BR-M16 | sí |
| el dinero, formateado en un solo lugar | `entities/telemetria/ui/cifra-usd.tsx#CifraUsd` | `Estandar` · `SeparadorDosDecimales` · `MenorAlCentavo` · `Ausente` | RF-281 | n/a |

### §4 del mockup — La lista de puntos de mejora, debajo del canvas

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| la lista con su encabezado y contador | `ui/puntos-mejora-list.tsx#PuntosMejoraList` | `DosTarjetasOrdenadas` | H-1 · RF-246 | sí |
| la tarjeta insignia B1, completa | `ui/punto-mejora-card.tsx#PuntoMejoraCard` | `CompletaB1Atencion` · `CompletaB1AtencionDark` | RF-247…254 | sí |
| P1 crítica, con `Patrón` en vez de `Umbral` | ídem | `CriticaP1SinUmbral` | RF-250 · RF-257 | sí |
| la fila `Sesgo`, con su dirección | ídem | `SesgoSubestima` · `SesgoSobreestima` · `SinSesgoIdentificado` | RF-252 | sí |
| chip S1-only | ídem | `S1Only` | J-2 | sí |
| «ver el cálculo» desplegado en línea | ídem | `CalculoCerrado` · `CalculoAbierto` | H-4 | sí |
| tarjeta resaltada por su caja | ídem | `Resaltada` | design §5.4 · RF-280 | sí |
| `Descartar` con vuelta atrás | ídem | `DescartarLlamaHandler` | RF-256 | **sí** · **desde D26.4**: `POST\|DELETE …/mejoras/{puntoId}/descartar` + tabla `punto_descartado` |
| `Proponerlo en el chat` **no escribe** | ídem | `ProponerAbreChatNoEscribe` · `ProponerDeshabilitadoFueraDeAlcance` · `NotaAlPie` | RF-255 · BR-M12 · D17.3 | **sí** · **desde D26.4**: buzón `propuestaChat` en `app-store`, el Dock lo consume UNA vez y **puebla el composer sin enviar** |
| severidad legible sin color | ídem | `SeveridadSinColor` · `TitularSinJerga` | RF-247 · RF-257 · RF-280 | sí |
| estado vacío con los 6 detectores | `puntos-mejora-list.tsx` | `VaciaConDatos` | H-2 | sí |
| sin contrafactual no hay tarjeta | ídem | `DescartaSinContrafactual` | RF-246 · A4 | sí |
| carga y error de la lista | ídem | `Cargando` · `ErrorDeConsulta` · `MuchasTarjetas` | design §5.4 · §6 | sí |

### §5 del mockup — Inspector, cuarta tab «Mejora»

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| la 4ª tab con su contrato ARIA | `ui/inspector.tsx` (superset) | `CuatroTabs` · `TabMejoraContratoAria` · `CambioDeNodoVuelveAlResumen` · **`Tabs`** (ver §3) | RF-258 · RF-277 | sí |
| tabla de 6 buckets, la suma cierra | `ui/inspector-mejora.tsx#InspectorMejora` | `Completo` · `CorridaRealMedida` | RF-259 | sí |
| «no aplica» ≠ «midió 0» | ídem | `NoAplicaNoEsCero` · `CeroLegitimo` | RF-260 · RF-286 | sí |
| paridad reportado vs. calculado | ídem | `ParidadCoinciden` · `ParidadDivergen` · `SinCostoDelRuntime` · `CatalogoSinConstruir` | RF-261 · RF-272 | sí |
| el join de la caja | ídem | `JoinCompleto` · `SinSenalDeGate` | RF-262 | parcial · las 3 filas de proceso llegan `null` (T22 abierto) |
| los 6 detectores + los 7 no medidos | ídem | `SeisDetectoresConEstado` · `DetectorB1ApagadoEnS2` · `DetectorB2Parcial` · `SieteNoMedidos` · `DetectorSinFixPropuesto` | RF-263 · RF-271 · T-09 | sí |
| la ventana es la de la capa | ídem | `VentanaHeredada` | RF-264 · J-3 | sí |
| nodo que no es caja | ídem | `NodoNoCaja` | RF-258 | sí |
| carga y error dentro de la tab | ídem | `Cargando` · `ErrorDeConsulta` | design §5.5 | sí · **desde C-3** |

### §6 del mockup — La tarjeta del Portafolio

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| `USD/corrida` · `tendencia` · `punto de mejora` | `widgets/portafolio/ui/tabla-mejora-portafolio.tsx` | `ConDato` · `ConDatoDark` | RF-265 · D21 | sí |
| un arnés en dos instalaciones | `celdas-mejora-fila.tsx` (la tabla se borró en C-4) | `ConMejoraDosInstalaciones` · `ConMejoraUnaSolaInstalacionNoAvisa` | RF-265 · D20 | **sí** · **desde D26.4**: la fila DICE «1 de N instalac.» — la cifra sigue siendo de una, pero ya no se lee como la del arnés |
| «puesto sin declarar» | ídem | `SinPuestoDeclarado` | RF-265 · D20 | sí |
| sparkline + texto equivalente | `entities/telemetria/ui/sparkline.tsx` | `EnAlza` · `Estable` · `HaciaLaBaja` · `PocasCorridas` (×2: entity y tabla) | RF-266 · RF-279 | sí (siempre el copy de ausencia: el wire no manda serie) |
| «✓ sin fugas» ≠ «sin dato» | `tabla-mejora-portafolio.tsx` | `SinFugas` · `SinDato` | RF-267 · RF-268 · RF-280 | sí · **desde C-4** |
| orden por costo, sin-dato al final | — | — (story borrada con C-4) | RF-268 | **n/a** · la superficie que ordenaba **ya no existe**: las celdas viven en la fila del Portafolio, que tiene su propio orden. Reordenar el Portafolio entero por costo es una decisión de producto que nadie tomó — no se hace de oficio |
| pie de tabla + confianza por fila | ídem | `PieConDisclaimerDeEstimacion` | H-12 | sí · **desde C-4** |
| desbordamiento y scroll accesible | ídem | `NombresLargosNumerosGrandes` | design §6.3 | parcial · M-5 abierto |
| H-3 · el puente al Mapa | `portafolio-drawer.tsx` + `app-store#mapaPeek` | (afordancia vigente «Abrir en Mapa») | H-3 | **sí** · el puente existe y está cableado desde Slice 1; lo que se había borrado con C-4 era una SEGUNDA entrada al mismo puente |
| las 3 celdas en la fila de la lista | `ui/portafolio-list.tsx` (superset) | `ConMejora` · `SinMejoraDomIntacto` | RF-265 · BR-M16 | sí |

### §7 del mockup — Los estados honestos

| # | estado | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| 1 | nunca corrió | `Estado1SinDatos` | RF-269 | sí |
| 1b | sin corridas en la ventana | `Estado1bSinCorridasEnVentana` + `estadosDeLaFranja` (4 tests de tabla) | H-9 · T-15 | **sí** · **desde D26.4**: el wire trae `ultima_corrida`, que mira TODO el historial e ignora la ventana |
| 2 | cobertura parcial | `Estado2CoberturaParcial` | RF-270 | sí |
| 3 | S2 sin instrumentar | `Estado3S2SinInstrumentar` | RF-271 · J-10 | sí · **desde C-2** |
| 3b | S2 instrumentado | `Estado3bS2Instrumentado` | RF-271 · ANEXO H9 | sí · **desde C-2** (era inalcanzable) |
| 4 | otro runtime | `Estado4OtroRuntime` + test de tabla con control negativo | RF-272 | **sí** · **desde D26.4**: se deriva de «el runtime no reportó costo y el catálogo sí» |
| 5 | catálogo viejo | `Estado5CatalogoViejo` + test de tabla con control negativo | RF-273 | **sí** · **desde D26.4**: sale de `/salud`, que la página ya consume |
| 6 | por huella | `PorHash` · `MejoraPorHuella` | RF-274 | sí |
| — | runtime no soportado | `RuntimeNoSoportado` + test de tabla (doble control negativo) | escenarios A5 | **sí** · **desde D26.4**: el wire trae `runtime` + `runtime_soportado`, derivados del dato real |
| B1 | cargando | `Cargando` (franja · lista · inspector) | H-6 | sí |
| B2 | error de consulta | `ErrorDeConsulta` | H-6 | sí |
| B3 | daemon caído, cifras viejas | `DaemonCaido` | H-6 | **sí** · **desde D26.4**: la página distingue «no contestó» (`TypeError`, sin status) de «contestó un error», y solo la primera conserva las cifras viejas marcadas |
| — | promoción de los estados a `shared/ui` | `SkeletonHonesto` · `SkeletonConMedidaPropia` · `ErrorConMotivoReintentable` | H-6 | n/a |

### §8 del mockup — Qué guardamos y cómo se borra

| qué dibuja el mockup | componente real | story = test | RF | ¿alcanzable? |
|---|---|---|---|---|
| los 3 bloques del diálogo | `ui/politica-datos-dialog.tsx` | `Reposo` | RF-275 · H-5 · H-14 | sí |
| la lista de campos, **por prop** | ídem | `CamposDesplegados` | RF-275 · RF-282 | sí (allowlist fija del hook, no del wire) |
| retención desde la config | ídem | `RetencionDesdeConfig` | RF-283 · J-6 | sí · **desde A-2** |
| confirmación con alcance | ídem | `Confirmacion` | RF-275 | parcial · A-4: el conteo es de la ventana, el borrado es de todo |
| borrando / error / éxito | ídem | `Borrando` · `ErrorDeBorrado` · `Exito` | design §5.7 | sí |
| focus trap y teclado | ídem | `FocoAtrapado` | RF-278 | sí |

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
| **D-16** | 138 stories, no 125 | +6 de cobertura de contrato (§3) y +7 del defecto de composición D24 (§8) | — |

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
| ~~7~~ | ~~M2 — B1 y B3 ponen conteos de tokens en campos `micros`~~ | ✅ **CERRADO por D26.1**: los dos cotizan con el catálogo. El candado es `TestElMontoDelCacheSaleDeLaTarifaNoDelConteo` — duplica la tarifa con los mismos tokens y exige que el monto se duplique; si no se mueve, el número son tokens disfrazados de dinero |
| 8 | **`RespuestaMejoras` no trae la lista de detectores que corrieron y no encontraron nada.** Trae `no_aplican` (no pudieron correr) y `no_medidos` (fuera del MVP); un detector que aplicó y salió limpio **no está en ninguna** | H-2 pide esa lista. **Parcialmente resuelto en D24.4**: el CONTEO sí es derivable (`6 − no_aplican.length`), así que el vacío dice cuántos corrieron y enumera los que no pudieron con su motivo. Lo que sigue faltando es poder **nombrar** a los que salieron limpios |
| ~~10~~ | ~~B1 no puede emitir contrafactual~~ | ✅ **CERRADO por D26.1**: el contrafactual compara contra **una sola escritura a 1 h** —de ahí sale el ahorro— y el monto se cotiza con el catálogo. `ScoreVersionMVP` 1 → 2. Con una sola escritura en la ventana no hay tarjeta, y es correcto: no hay re-warm que evitar |
| 9 | 🔴 **`DetalleCaja` no trae el costo POR BUCKET.** Trae `Tokens` (los seis punteros) y `Paridad` (los dos totales), pero no un desglose de dinero por bucket | La columna `USD` de la tabla del inspector viaja **`null` — «no aplica»** contra el wire real. Calcularla en el FE sería costear en la UI, que es justo lo que `design.md` §1.3 prohíbe («entities presenta; no calcula»). **El assert de «la suma cierra» (`Completo`) corre contra fixture, no contra el wire**, y no puede correr contra el wire hasta que este campo exista |

---

## 5 · Lo que quedó SIN verificar

| # | qué | por qué |
|---|---|---|
| V-1 | **La cadena E2E contra el daemon vivo** (plan §5.2-5.6: HOME de prueba, puerto efímero, `claude` real, app instalada) | Es el **Tramo C (T38-T39)**, fuera del encargo de este tramo. Todo lo construido acá se verificó contra fixtures rotuladas y contra los shapes reales del wire leídos en `internal/domain/telemetria_vistas.go`, **no** contra respuestas HTTP en vivo |
| V-2 | **El aspecto real de la capa en el navegador** | No se levantó el daemon ni se sacó captura. Las stories corren en Chromium headless y assertan DOM, contraste (axe) y `getComputedStyle`, pero **nadie miró la pantalla** |
| V-3 | **`fitView` con los nodos ~26 px más altos** (design §6.2) | Es comportamiento de layout con el canvas montado a tamaño real; las stories lo montan a 620 px de alto. El razonamiento está escrito (overview-first, RF-50) pero no medido |
| V-4 | **El tema oscuro más allá de las 3 stories espejo** | `ReposoDark`, `CompletaB1AtencionDark` y `ConDatoDark` cubren franja, tarjeta y Portafolio. El inspector, la lista y el diálogo **solo tienen gate en claro** |
| ~~V-5~~ | ~~**El wire real de `puntos`**~~ | ✅ **CERRADA por D25** (2026-07-26). La prosa se arma en el dominio (`telemetria_prosa.go#Redactar`), los 7 campos que el FE tipaba sin productor ahora existen o se borraron (`caja_nombre`), y **el JSON real está atado al tipo del FE** por `telemetria_contrato_fe_test.go`, con control positivo. Lo que la decisión destapó está abajo, en el ítem 10 de §4 |
| V-7 | **La suite tiene flakiness intermitente bajo carga** — no introducida por este paquete | En 7 corridas completas seguidas, 5 cerraron **exactamente** en los 4 rojos preexistentes y 2 sumaron fallos intermitentes. Los archivos afectados incluyen `shared/ui/buscador-filtro.stories.tsx` y `entities/marketplace/ui/chips.stories.tsx`, **que este paquete nunca tocó** — así que es del runner (browser mode en paralelo), no del diff. Ninguno de los archivos nuevos falló dos veces seguidas, y los 4 archivos nuevos corridos en aislamiento pasan 3/3. **Vale la pena abrirlo en el BACKLOG**: una suite que falla 2 de cada 7 veces por razones de infraestructura entrena a re-correr, y eso es exactamente lo que hace que un rojo real pase desapercibido |
| V-6 | **La 4ª tab contra el wire real** | Está cableada (`workspace-stage.tsx` pide `GET …/cajas/{cajaId}` al seleccionar una caja con la capa encendida), pero la columna USD por bucket llega `null` (§4 ítem 9) y el join llega con `null` en las tres filas de proceso (T22 abierto). **Lo que se ve contra el wire real es una tabla de tokens y una paridad**, no el desglose completo del mockup |

> ✅ **V-5 está cerrada (D25).** No era una decisión abierta: `design.md` §1.3 —firmado— ya decía
> *«el contrafactual se resuelve en el dominio Go y viaja resuelto en el wire»*, y la
> implementación se lo salteó. El dominio arma la frase; el FE dejó de tipar campos sin
> productor; un fitness test ata los dos lados y falla si vuelven a separarse.
>
> 🔴 **Lo que cerrar V-5 destapó, y el gate humano tiene que decidir:** al redactar la frase de
> los seis detectores, **B1 —la tarjeta insignia— no la pudo armar**. Su contrafactual cotiza las
> mismas escrituras de cache a la tarifa larga, que es más cara, así que su «ahorro» es negativo
> por construcción. Hoy sale declarado `sin_fix` con su motivo en vez de dibujar una tarjeta que
> dice que el arreglo cuesta más. Detalle en `decisiones.md` §D25.4.

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

## 7 · 🧑‍⚖️ Gate humano — **FIRMADO 2026-07-27**

**Firmado por el operador**, con **D-1 ok** y **D-2 ok**, junto con las seis decisiones de
[`decisiones.md`](decisiones.md) §D26. El alcance de la firma está acotado ahí y se repite acá
porque importa: cubre el Tramo B construido **más la ejecución de D26**, y **nada más** — lo que
se destape al ejecutar vuelve al gate, no se cuela bajo esta firma.

Los seis ítems que el gate recorrió, con lo que el operador resolvió en cada uno:

| # | ítem | resuelto |
|---|---|---|
| 1 | D-1 y D-2, las dos desviaciones 🔴 sobre superficie firmada | **ok las dos** |
| 2 | §4 — los huecos del backend que la UI declara en pantalla | los que D26 decidió arreglar se arreglan; el resto sigue declarado |
| 3 | V-5, el contrato `PuntoDeMejora` Go ↔ FE | cerrado por **D25**; su consecuencia (B1) **se arregla** (D26.1), no se declara |
| 4 | El TTL de retención | **90 días, firmado** (D26.3). Sale el rótulo «propuesto» |
| 5 | Las 10 filas «no alcanzable» de §1 | **se cablean las diez** (D26.4) |
| 6 | A-4, la única acción irreversible | **el `DELETE` acepta ventana** (D26.5): se corrige el poder, no el copy |

<details><summary>Lo que esta sección decía antes de la firma</summary>

Este tramo **no está firmado**, y la auditoría independiente del Tramo B dijo **no firmar** (§9).
Los 4 críticos están corregidos; lo que queda abierto está en §9 y en §4-§5.
Para firmarlo hay que mirar, como mínimo:

1. **D-1 y D-2**, las dos desviaciones marcadas 🔴 sobre superficie firmada.
2. **§4 completo**: siete huecos del backend que la UI declara en pantalla. Si alguno no es
   aceptable como estado visible, cambia el diseño, no el código.
3. ~~**V-5**: el contrato de `PuntoDeMejora` entre Go y el FE.~~ **Cerrado (D25).** Lo que queda
   para el gate es su consecuencia: **§4 ítem 10 — B1, la tarjeta insignia, no se dibuja** porque
   su contrafactual da un ahorro negativo. Decidir si se corrige la fórmula (bump de
   `ScoreVersionMVP`) o si el mockup deja de tener a B1 como insignia.
4. **El TTL de retención** sigue **PROPUESTO** (J-6 · parada P2). La UI **ahora sí** lo lee de
   `GET /api/telemetria/salud` (antes estaba hardcodeado en 90 — la afirmación anterior de esta
   hoja era falsa, A-2) y lo rotula según `retencion_propuesta`; el número lo pone el operador.
5. **La columna «¿alcanzable?» de §1**: 10 filas dicen `no`. Son superficies construidas y
   testeadas que la app no alcanza — hay que decidir si se cablean o se declaran fuera de alcance.
6. **A-4**, la única acción irreversible: el conteo que la confirmación declara es de la ventana y
   el borrado es de todo el historial. Es decisión de producto.

</details>

---

## 8 · 🔴 Defecto de COMPOSICIÓN, encontrado en la app instalada (D24)

**No lo cazó ninguna de las 131 stories**, y no podía: cada una renderiza un componente aislado,
y la mentira no estaba en ninguno — estaba en la composición. Lo cazó el operador mirando el
binario instalado contra sus datos reales
(`verificacion-2026-07-26/evidencia/instalada-v0222-capa-mejora.png`).

**Lo observado** — arnés `vitalia`, que nunca corrió:

| bloque | qué decía | ¿verdad? |
|---|---|---|
| franja | `— —` · «Este arnés nunca corrió con telemetría.» | ✅ |
| lista, 12 cm más abajo | «**Hay datos** y ningún punto de mejora que pase el corte.» | ❌ no hay ninguno |
| ídem | «**Los seis detectores corrieron** sobre 0 corridas.» | ❌ correr sobre cero no es correr |

Las dos falsas contradicen a la de arriba, y juntas hacen leer «los detectores buscaron y no
encontraron nada» donde la verdad es «todavía no medimos nada». Es la formulación exacta que
RF-263 prohíbe: **presentar la ausencia de búsqueda como resultado de una búsqueda.**

### Qué se corrigió

| # | fix | dónde |
|---|---|---|
| 1 | Sin medición atribuible, **la sección no se dibuja** — que es lo que `design.md` §7.4 ya decía y la implementación ignoró. `hayDatos` pasa a prop **obligatoria** (`tsc` cazó los 6 call-sites): un default que asume datos miente cuando no los hay | `puntos-mejora-list.tsx` |
| 2 | La condición vive en **un** lugar y la usan la página **y** las stories. Duplicarla en el composition-root es cómo nació el defecto | `model/capa-mejora.ts#hayDatosAtribuibles` |
| 3 | El vacío dice **cuántos** detectores corrieron (`6 − no_aplican`) y lista los que no pudieron con su motivo. «Los seis corrieron» era falso también en `s2-degradado`, donde B1 no puede correr. El literal firmado se conserva para el caso que describe | ídem |
| 4 | El ✓ `sin fugas detectadas` del Portafolio sale de `puntos_de_mejora === 0` **como dato del wire**, no de la ausencia del campo `punto`. Con hallazgos sin punto principal se dice cuántos hay: un ✓ miente en la dirección más cara, «acá no hay nada que mirar» | `tabla-mejora-portafolio.tsx` |
| 5 | Una sección «Detectores» sin filas se leía como «no hay detectores». Ahora lo dice | `inspector-mejora.tsx` |
| 6 | El umbral del estado 2 era un `corridas <= 5` **escondido en el JSX**. Pasa a `coberturaEsParcial()`: sigue siendo decisión de producto (más de un tercio sin atribuir), pero **declarada, con nombre y testeada por los dos bordes** | `model/capa-mejora.ts` |

### ⚠️ La condición obvia habría creado el mismo defecto al lado

La formulación natural —`resumen.confianza === "sin-dato"`— **es incorrecta**: la confianza del
agregado es `PeorConfianza` de la ventana (`telemetria_service.go:334`), así que **una sola**
corrida sin atribuir la deja en `sin-dato` con 60 perfectas al lado. Habría escondido la lista en
el **estado 2 (cobertura parcial), que sí tiene datos**. La condición correcta se deriva de la
cobertura: `corridas > 0` **y** `exacta + por_hash + por_proceso > 0`.
Cementado en `CoberturaParcialConservaLaLista`.

### El candado, verificado como test y no asumido

`widgets/map-canvas/ui/capa-mejora-coherencia.stories.tsx` — **6 stories** que renderizan los dos
bloques desde **un solo `resumen`**, con las mismas funciones que usa la página (patrón de
`CopyConfianzaEsUnaSola`, candado de D18).

**Control positivo del propio candado**, corrido a mano: revirtiendo el fix, las 3 stories de
ausencia se ponen rojas y el DOM reproduce el texto del screenshot palabra por palabra
(`…nunca corrió con telemetría…Puntos de mejora0Hay datos y ningún punto…sobre 0 corridas…`);
`ConDatosLaListaSiAfirmaBusqueda` queda **verde**, así que «esconder siempre la lista» **no**
pasa el candado.

### La regla que queda

> **Toda superficie que componga dos bloques que hablan del mismo hecho lleva una story de
> coherencia sobre el DOM completo.** Un assert por componente no puede ver una mentira por
> composición — y este repo tiene 0 stories en `pages/`, que es justo donde se compone.

Queda como deuda declarada: **`pages/` sigue sin cobertura**. El candado cubre esta composición
concreta porque la story la reproduce; una composición futura que nadie espeje vuelve a quedar
ciega.

---

## 9 · 🔴 Auditoría independiente del Tramo B — 4 críticos, corregidos

[`auditoria-tramo-b.md`](auditoria-tramo-b.md), 2026-07-26. **Veredicto: no firmar.** Los cuatro
son de la misma clase que D24 —bloques defendibles por separado que juntos afirman algo falso— y
**el candado de D24 no cubría ninguno**: montaba `FranjaMejora` + `PuntosMejoraList`, y los
bloques que hablan del mismo hecho en esta capa son **cinco** (franja · carril · nodo · tarjeta ·
inspector) más la fila del Portafolio.

### La causa común, y el refactor que la ataca

Los cuatro vivían en **`pages/shell/ui/workspace-stage.tsx`**: el único lugar del repo donde se
componen las cinco superficies, y el único **sin stories**. La conclusión de la auditoría es la
correcta y la adopto: *«`pages/` con 0 stories no es deuda de cobertura — es el único lugar donde
vive esta clase de defecto»*.

**La composición sale de `pages/`.** Nace `widgets/map-canvas/ui/capa-mejora-stage.tsx`, un widget
con stories que recibe el wire crudo y deriva **todo** con `vistaCapaMejora()` — pura,
unit-testeada, único consumidor. La página quedó con transporte y estado, y ya no puede
contradecirse a sí misma porque no toma ninguna de estas decisiones.

| # | qué | fix | test que falla si vuelve |
|---|---|---|---|
| **C-1** | Encender la capa **destruía el Mapa**: canvas 739 px → 257 con una tarjeta → **0 px con cuatro**; 0 px con UNA a 200 % de zoom | el canvas lleva piso, la lista techo con scroll adentro. En una vista que se llama «Mapa», el mapa gana el reparto | `CapaEncendidaConservaElMapa` · `CuatroTarjetasNoAplastanElMapa` · `ZoomAltoConservaElMapa` — **miden el alto real** |
| **C-2** | En `s2-instrumentado` (el escenario que esta hoja declara **mayoritario**) la franja decía «nunca corrió» con USD 1,08 en el carril, la rama de s2 era **inalcanzable** y **el candado de D24 escondía el entregable** | `corridas` no es proxy de «hay datos»: es 0 fuera de S1 por construcción. La condición es la cobertura atribuida | 15 tests de tabla + `S2InstrumentadoNoDiceNuncaCorrio` |
| **C-3** | Un `GET` de detalle fallido **se pintaba como dato**: ocho afirmaciones falsas, con la nota «"No aplica" no es 0» **defendiendo la mentira** | el cuerpo de la 4ª tab lo arma el widget, no la página | `DetalleRotoNoAfirmaSobreElRuntime` + control positivo |
| **C-4** | `TablaMejoraPortafolio` era **código muerto**; la superficie real era una **segunda implementación sin el guard**, y el ✓ mentía con 3 hallazgos | se borra la tabla (175 líneas); una sola implementación en el widget; la página pasa wire crudo | `ConHallazgosNoPintaElTilde` + `SinFugasPintaElTilde` |

**Los cuatro fixes se verificaron revirtiéndolos**, no asumiéndolos: cada lock se pone rojo sin su
fix y los controles positivos quedan verdes.

### Dos lecciones que me llevo, y una que duele

1. **Un componente con su rama de error y su story puede tener el defecto igual.** C-3 lo prueba:
   `InspectorMejora` tenía las dos, y el bug estaba en quién decidía. Lo verifiqué revirtiendo el
   cableado — **la story del componente siguió verde**. El test tiene que vivir donde vive la
   decisión.
2. **Un fix aplicado a un componente que nadie monta no es un fix.** C-4: D24.4 estaba escrito,
   testeado e inalcanzable. Antes de declarar un fix hay que verificar que la superficie lo usa.
3. La que duele: **mi candado de D24 se convirtió en la causa de C-2.** Escribí una condición
   para tapar un agujero y la basé en el campo equivocado, así que escondió el entregable central
   en el caso normal. Un candado mal fundado es peor que ninguno.

### Altas atendidas en esta pasada

| # | qué | estado |
|---|---|---|
| **A-1** | `Descartar` y `Proponerlo en el chat` no hacían nada y lo disimulaban con un refetch | **corregido**: sin handler nacen `disabled` con su motivo en texto (patrón `BotoneraStaged`). El anuncio ya no promete una afordancia que no existe |
| **A-2** | `GET /api/telemetria/salud` construido y **nunca consumido**: retención hardcodeada en 90 y **el chip de reenvío externo no se dibujaba nunca** | **corregido**: la página lo consume; retención y reenvío salen de ahí |
| **A-5** | «nunca corrió con telemetría» sobre un arnés con 47 corridas | **corregido**: sale de `corridas`, no de `costo_por_corrida === null` |
| **A-6** | La superficie real había perdido el pie H-12, la `MarcaConfianza` y la unidad de la celda | **corregido** con C-4 |
| **A-7** | El denominador mezclaba dos magnitudes (`corridas` vs turnos) prometiendo que cerraban | **corregido**: nombra las dos unidades. Desviación del literal de design §7.2, declarada |
| **M-7** | `catálogo v` colgante | **corregido** |
| — | la caja de paridad se pintaba ámbar «difieren» con `divergencia === null`, sin veredicto | **corregido**: sin veredicto no hay tono |

### Lo que sigue ABIERTO de la auditoría

| # | qué | por qué no se hizo acá |
|---|---|---|
| ~~**A-3**~~ | ~~Estado 1b inalcanzable~~ | ✅ **CERRADO por D26.4**: el resumen trae `ultima_corrida`, que mira todo el historial e ignora la ventana |
| ~~**A-4**~~ | ~~«Se borran 61 corridas» — el conteo es de la ventana, el `DELETE` es de todo~~ | ✅ **CERRADO por D26.5**: el `DELETE` acepta ventana. Se eligió arreglar el poder, no el copy |
| ~~**M-1**~~ | ~~Nueve estados storiados e inalcanzables~~ | ✅ **CERRADO por D26.4**: se cablearon las diez. La transparencia (la columna «¿alcanzable?») fue el paso intermedio, no la solución — el operador eligió arreglar en vez de declarar |
| **M-2** | Un error de consulta deja el dinero viejo en el canvas | Relacionado con B3, que la página no tipa |
| **M-3** | «pocas corridas para una tendencia» nombra una causa que no es la causa | El wire no manda serie (§4 ítem 1); el copy honesto sería «todavía no calculamos la serie» |
| **M-4** | `por-proceso` avisa de la mezcla **solo por `title`** | Cambia copy firmado de `design.md` §7.3 |
| **M-5** | El Portafolio desborda 21 px en horizontal y recorta chips a 900 px | No medido tras C-4, que cambió el marcado de las celdas |
| **M-6** | Tipografía de 9 px en los rótulos que hacen el argumento | Cambia `design.md` §2.5 |
