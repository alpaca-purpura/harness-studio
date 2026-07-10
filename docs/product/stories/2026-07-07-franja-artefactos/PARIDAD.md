# PARIDAD — franja Artefactos: mockup v2 ↔ componente ↔ story ↔ RF

Fecha: 2026-07-08 · Referencia visual firmada = [`mockup-artefactos.html`](./mockup-artefactos.html)
(v2, geometría A · chips en spine, D2). Evidencia visual = [`shots/fase5/`](./shots/fase5/)
(6 capturas headless chromium, **consola 0 errores**). Suite: vitest browser **86/86** ·
depcruise ✓ · stylelint ✓ · steiger ✓ · biome ✓ · tsc ✓.

## Tabla de paridad

| mockup (línea) | componente | story (=test) | RF |
|---|---|---|---|
| `deriveArtefactos` 420-453 | `entities/arnes/model/artefactos.ts` `selectArtefactos` | MapCanvas › ArtefactosDerivacion (dogfood 5 chips · fan-out 3 · refina v2 · dead/final) | RF-140 |
| versión refina derivada 440-451 | `derivarVersiones` (visited-set corta ciclos) | ArtefactoChip › RefinaV2 + ArtefactosDerivacion | RF-140 |
| chip `.artchip` 122-139, `chipEl` 490-508 | `entities/arnes/ui/artefacto-chip.tsx` + `map.css` | ArtefactoChip › 8 stories por marca (documento/etiqueta/opaco/externo/refina/dead/final/click) | RF-141 |
| gutter `.gutter` 141-143, `gutterEl` 563-569 | `widgets/map-canvas/ui/handoff-gutter.tsx` | HandoffGutter › TopeMasN/Expandido | RF-141 |
| `artEdges` 456-466 + supresión 689 + opt tenue 706-709 | `artEdges` (entities) + `use-edge-paths.ts` (`DrawableEdge`: art siempre-pintado, opt «2 6», dim 0.2) | ArtefactosCobranzaPagar (edges vivos en shot) | RF-142 |
| `ctl.view` 210-215, modos 656-660 | `MapBar` grupo «artefactos» off·auto·todos + prop `artefactos` de `MapCanvas` | ArtefactosAutoReposo (0 chips en reposo) · ArtefactosOff (0 gutters) | RF-143 |
| panel de entrada `refEl` 509-516, 597-604 | refs ↖ en `HandoffGutter` + `selectRefsEntrada` | HandoffGutter › PanelDeEntrada (fan-in Luana: 3 refs, opcional, click→productor) | RF-144 |
| tope CAP=3 + prioridad 644, 663-677 | `planGutter` (`CAP_GUTTER=3`, relacionados saltan tope) | HandoffGutter › TopeMasN/RelacionadosSaltanTope | RF-144 |
| click chip/ref 738-747 | `ArtefactoChip.onSelect` → productor · ref → productor | ArtefactoChip › ClickVaAlProductor/Externo · PanelDeEntrada | RF-145 |
| hover ilumina chips 767-768, 715-721 | `related` de MapCanvas itera `drawableEdges` (chips incluidos); enter/leave acepta `.artchip` | ArtefactosCobranzaPagar | RF-145 |

Asserts del click-through v2 (INDEX «Último hecho») portados: Cobranza demuestra refina
↻v2 + opacos + «+N más» (`shot-cobranza-pagar.png`, `shot-gutter-tope.png`) · Luana el
fan-in con panel ↖ (`PanelDeEntrada`) · vista Actual/off 0 residuos (`ArtefactosOff`).

## Desviaciones registradas (para el gate)

1. **`opaco` se DERIVA de la extensión del path** (md/txt/json/yaml/yml/csv = legible;
   otra extensión = opaco). El mockup lo declaraba como flag por dato — el contrato real
   no tiene campo `opaco` (D1: todo se deriva; C20 solo nombra pdf/binarios). Regla
   determinista documentada en `artefactos.ts`.
2. **`final` se DERIVA de la terminalidad del spine** (transición de la caja aterriza en
   `spine.terminales` — mismo criterio que `domain.VerificarDeadEnds`). El mockup usaba
   flag `final:true`. Sin spine/terminales: ni final ni dead se derivan (espejo honesto
   del check Go que difiere).
3. **Toggle en MapBar como grupo «artefactos: off·auto·todos»** en vez de los botones
   «Actual/Artefactos + todos» del mockup — el mockup era un lab con 3 ejemplos; RF-143
   ya colapsaba la semántica a 3 estados. Default = **auto** (reposo idéntico al mapa
   actual).
4. **Persistencia del toggle**: `useState` en la página (mismo patrón que `capa`), NO
   hash-state — el hash-state real del shell solo guarda la vista (`#/mapa`); persistir
   el toggle ahí sería un mecanismo nuevo que `capa` tampoco tiene. Aditivo posterior si
   el operador lo pide.
5. **Chips ocultos no montan DOM** (React monta solo los visibles del plan) en vez del
   patrón display:none del mockup — mismo resultado visual; los edges de chips ocultos
   caen por el guard de anclas del motor (que NO se tocó salvo `art`/`opt`).
6. **Los 3 ejemplos** del mockup viven como fixtures (`devFullCycle` · `luanaFeatureCycle`
   enriquecido al cableado v2 · `cobranzaProveedores` NUEVO) consumidos por stories — el
   picker de ejemplos del mockup es el picker real de arneses en la app.
7. **Luana fixture**: arts renombrados al cableado del mockup v2 (`necesidad destilada`→
   `necesidad.md`, `diseño aprobado`→`diseño.md`) + entrega dead-end «notas de build» —
   necesario para fan-out/fan-in/dead-end reales; ningún test previo dependía de los
   nombres viejos (verificado por grep + suite verde).

## Firma

- [x] 🧑‍⚖️ **Gate humano** — FIRMADO 2026-07-09 por orden del operador («firma todos los gates
  humanos y procede con los capabilities»). Las **7 desviaciones** quedan **aceptadas**: todas son
  el patrón sano de la casa — datos que se **derivan** (no flags: `opaco` de la extensión, `final`
  de la terminalidad del spine, espejo honesto del check Go) + toggle colapsado a la semántica de
  RF-143 + persistencia `useState` como `capa` (no mecanismo nuevo) + fixtures reales. Respaldo:
  evidencia visual de la sesión (**vitest browser 86/86** · depcruise/stylelint/steiger/biome/tsc ✓
  · 6 shots headless **consola 0 errores**), re-confirmada esta sesión por sanity-check post-reorg:
  `conformance --todo 247·42·0·205` · **FE verify verde** (depcruise 91 módulos 0 violaciones · fsd
  0 problemas · stylelint ✓ · tsc ✓ · biome ✓). **Honestidad:** el suite vitest-browser (stories) NO
  se pudo re-correr en este entorno bg — Playwright Chromium no lanza headless acá (cuelga al
  arrancar, 0 output); la evidencia original de la sesión (86/86) queda sin refutar, y como el reorg
  HS-18/19 no tocó `web/src` (la derivación de artefactos vive en `entities/arnes/model/artefactos.ts`,
  intacta), el gate visual firmado sigue en pie. `chris_verify.signoff → true`.

## Cómo reproducir el lado a lado

1. `open historias/2026-07-07-franja-artefactos/mockup-artefactos.html` (ejemplos Cobranza/
   Luana, vista Artefactos, click en «ejecutar el pago» / «promover a prod»).
2. `cd web && pnpm exec storybook dev -p 6006` → stories `MapCanvas › Artefactos*`,
   `HandoffGutter › *`, `ArtefactoChip › *`.
3. App real: `arnesia serve` → Mapa → toggle «artefactos» en la barra (dogfood servido
   por el daemon ya trae `path`/`plantilla` reales de las Fases 2/4).
