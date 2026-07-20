# Shell — Topbar sin empresa + selector de arnés al abrir sesión

> `tipo: paquete-de-trabajo` · abierto 2026-07-20 · disparado por observación directa del operador
> («hay una barra que tiene esa información hardcodeada alpacapurpura») usando la app. Etapa:
> **mockup → decisiones (Gate 1 FIRMADO 🧑‍⚖️ 2026-07-20) → spec → implementar → PARIDAD**. Spec y
> build siguen abiertos.

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

- [`decisiones.md`](./decisiones.md) — TS-D1..D9, las decisiones de diseño cerradas en la conversación
  (con el porqué de cada una — varias corrigen un error de un intento anterior, documentado).
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
> **Retomar aquí:** siguiente etapa = **spec**. Ver `decisiones.md` §«Para la spec» — deuda visible
> pendiente de resolver ahí (estado vacío del Portafolio, colisión de `id` entre entradas, qué pasa
> si `GET /api/portafolio` falla al abrir el picker). Luego implementar (Storybook = SSoT real, este
> `.html` es solo referencia de paridad visual) y cerrar con PARIDAD + gate humano.
