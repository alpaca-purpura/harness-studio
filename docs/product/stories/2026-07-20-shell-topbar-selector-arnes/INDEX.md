# Shell — Topbar sin empresa + selector de arnés al abrir sesión

> `tipo: paquete-de-trabajo` · abierto 2026-07-20 · disparado por observación directa del operador
> («hay una barra que tiene esa información hardcodeada alpacapurpura») usando la app. Etapa:
> **mockup → decisiones (Gate 1 FIRMADO 🧑‍⚖️ 2026-07-20) → spec+design (ESCRITAS) → implementar
> (CONSTRUIDO 2026-07-20, verde) → PARIDAD (escrita, gate humano final PENDIENTE 🧑‍⚖️)**. El
> operador ordenó «Desarrolla» → se tomó como autorización de Gate 2 (implementar contra el spec).

## Qué resuelve

Tres problemas reales de la Topbar y del flujo «＋ Nueva sesión» del rail
(`web/src/widgets/topbar/`, `web/src/widgets/session-rail/`), encontrados usando la app instalada:

1. **Empresa fija hardcodeada en el breadcrumb.** `Topbar` pinta `active.empresa` como si fuera un
   dato estable de la sesión; en el dogfood siempre resuelve a `alpacapurpura` sea o no la empresa
   real del arnés activo — y con el modelo del Portafolio (N:M:M arnés↔empresa) un arnés puede no
   tener una única empresa. Se saca del breadcrumb.
2. **Falso control de cambiar arnés a mitad de conversación.** El chip de arnés dibuja un `▾` sin
   `onClick` — invita a un cambio que la sesión no permite (1 arnés fijo por sesión) ni debería
   sugerir. Pasa a etiqueta plana.
3. **La elección real de arnés no existe.** «＋ Nueva sesión» (`session-rail.tsx`) crea la sesión con
   `arnes: "nuevo-arnes"` hardcodeado y solo pregunta la ruta vía `window.prompt()` nativo — ningún
   punto de la UI deja elegir un arnés real. Se corrige moviendo la elección a ESE momento (abrir
   sesión), y la fuente es el **Portafolio real del operador** (`GET /api/portafolio`), no un catálogo
   aparte — con buscador (escala cuando el portafolio crece) y selección en 2 niveles: identidad →
   copia, cuando la identidad tiene 2+ instalaciones.

De paso: `quedaste en: …` (`Session.parked`) se saca de la Topbar — es un campo diseñado pero nunca
cableado en vivo (solo seed hardcodeado en `session_service.go`, ver `decisiones.md` TS-D4). No se
diseña sobre un dato que no existe.

## Documentos

- [`decisiones.md`](./decisiones.md) — TS-D1..D9 (mockup, firmadas 🧑‍⚖️) + TS-D10..D17 (deuda de
  TS-D9 resuelta al escribir la spec: host del picker, estados vacío/error/colisión, refetch,
  cómputo de copias, payload real de sesión, dónde vive el fetch).
- [`spec.md`](./spec.md) — RF-1..17 + Gherkin, trazado a `mockup:línea`.
- [`design.md`](./design.md) — anatomía al pixel, tabla de campos por fila, estados, tokens
  (reuso total, cero CSS nuevo).
- [`PARIDAD.md`](./PARIDAD.md) — RF por RF → verificación (7 stories + 5 unit + build + R1/R2/R4),
  desviaciones documentadas, y el checklist del gate humano final (sin marcar).
- Mockup firmado: [`mockups/arnesia-shell-topbar-selector-arnes.html`](../../../../mockups/arnesia-shell-topbar-selector-arnes.html)
  (snapshot derivado; ver `mockups/INDEX.md` antes de tocarlo).

## Capabilities que toca (spec debe actualizarlas, no crear nuevas)

- `docs/product/capabilities/fe-shell/topbar-breadcrumb-k-dock.yaml` (CAP-74)
- `docs/product/capabilities/fe-shell/rail-de-sesiones.yaml` (CAP-72)

## Retomar aquí

> **Estado (2026-07-20): Gate 1 (mockup/diseño) FIRMADO 🧑‍⚖️ por el operador.** El mockup vivió 3
> iteraciones dentro de la misma conversación — cada una corrigió un error real señalado por el
> operador (ver `decisiones.md` para el detalle de qué se descartó y por qué):
> 1. primera pasada: quitar empresa + bajar Conversar + popover de arnés con datos de fixture
>    (`entities/arnes/testing/*` — dogfood, NO el Portafolio real) → **rechazado en concepto**.
> 2. segunda pasada: mismo picker pero inline (sin popover) + buscador real, siguiendo el patrón
>    `.pf-buscar`/`.pf-fila` de Portafolio → **rechazado en origen de datos** («yo debo poder elegir
>    de los arneses que ya agregué a mi portafolio», no un catálogo inventado).
> 3. tercera pasada (LA FIRMADA): lista sourced del modelo real (`EntradaPortafolio`: `empresas[]`,
>    `canonico`, `instalaciones[]`, `deriva`) con selección en 2 niveles cuando hay ambigüedad de
>    copia. Verificado visualmente con Chrome headless (screenshots reales, no solo lectura de código)
>    en ambos temas — 2 bugs reales de layout cazados y corregidos ahí mismo (texto sin truncar
>    reventando el ancho: falta de `min-width:0` en cadenas flex anidadas, y un track `1fr` de CSS
>    Grid sin `minmax(0, …)`).
>
> **Estado (2026-07-20, mismo día): spec + design ESCRITAS, luego IMPLEMENTADAS.** `spec.md`
> (RF-1..17 + Gherkin) y `design.md` (anatomía/tokens/estados) resuelven la deuda TS-D9 completa vía
> TS-D10..D17 — layout real del picker (rail 224→360px, TS-D10), fetch en store propio del widget
> (`portafolio-picker-store.ts`, TS-D17), estados vacío/error/colisión/refetch (TS-D11-14), copias
> 0/1/2+ (TS-D15), payload real (TS-D16).
>
> **Estado (2026-07-20): CONSTRUIDO — todo verde.** El operador ordenó «Desarrolla» (autorización de
> Gate 2). Implementado:
> - `topbar.tsx` (RF-1..4): empresa fuera · chip plano sin ▾ · Conversar 2ª línea · `parked` fuera.
> - `session-rail.tsx` (RF-5/16): `pickerOpen` ensancha el `<aside>` a 360px + oculta lista/pie;
>   `NewSessionButton` ya no usa `window.prompt`/hardcode — abre el picker y `onCrear`→`create(...)`.
> - **NUEVO** `new-session-picker.tsx` (RF-6..17, props-puras) + `portafolio-picker-store.ts` (TS-D17,
>   Zustand envuelve `api.listPortafolio`).
> - Tests: `new-session-picker.stories.tsx` (7 `play()` — simple·ambiguo·vacío·error·colisión·buscar·
>   cancelar, a11y axe verde ambos temas) + `portafolio-picker-store.test.ts` (5 unit).
> - `verify` verde · `build` OK · suite `vitest run` 168/168 · fitness R1/R2/R4 ok.
> - Capabilities CAP-72 (3 punteros nuevos + scenario) y CAP-74 (2 scenarios) actualizadas — R2
>   cobertura sigue 100%. `status` NO se tecleó (R4 generado, sigue `vivo·nc`, consistente).
>
> **Retomar aquí: gate humano final del paquete 🧑‍⚖️.** Ver checklist en [`PARIDAD.md`](./PARIDAD.md):
> el operador corre la app instalada, hace el click-through lado a lado contra el mockup en AMBOS
> temas, y firma. Nada se declara «listo» hasta esa firma (checkbox sin marcar, no simulada). El
> `.html` queda como referencia de paridad visual; el SSoT del UI es Storybook.
