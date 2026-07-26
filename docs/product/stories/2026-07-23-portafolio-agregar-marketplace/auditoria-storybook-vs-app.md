# Auditoría · ¿la app sigue al Storybook? — superficie «agregar al portafolio»

> `tipo: auditoria` · paquete `2026-07-23-portafolio-agregar-marketplace` · 2026-07-24.
> Método: **inspección visual en vivo**, no lectura de código sola. Tres superficies levantadas
> al mismo tiempo y comparadas a 1440×900: mockup snapshot (`:8099`) · Storybook (`:6006`) ·
> app real Vite+daemon (`:5173` + `:4200`). Evidencia = capturas en `.playwright-mcp/`
> (untracked; se re-generan con los pasos de §0).

## 0 · Cómo se reprodujo

```
pnpm --dir web storybook                      # :6006
go run ./cmd/arnesia serve                    # :4200
pnpm --dir web dev                            # :5173
python3 -m http.server 8099 -d mockups        # snapshot del mockup
```

Recorrido en la app: rail → **Portafolio** → **+ Agregar** → tab Proyecto → ruta
`/home/chalreme/Proyectos/harness-studio` → **Escanear** → estado `candidatos` (4 hallazgos reales).

## 1 · Veredicto (invierte la hipótesis de entrada)

**La app SÍ sigue al Storybook.** El story `paso-1-fuente` con `globals=theme:dark` sale
**pixel-idéntico** al wizard de la app: mismo header, mismos dos radiogroups («Origen» /
«Cómo cargar la ruta»), mismo input, mismo botón. Es el mismo componente
(`widgets/portafolio/ui/portafolio-wizard.tsx`) y el mismo scope CSS — el decorador `Frame` de las
stories envuelve en `.arnesia-portafolio`, que es exactamente lo que la página real monta, y
**todas** las reglas de `app/styles/portafolio.css` están calificadas con ese ancestro. No hay
un segundo juego de componentes ni estilos paralelos.

**Lo que sí driftó es el mockup contra el Storybook.** El `.html` es un snapshot hecho a mano y
quedó *adelante* en diseño y *atrás* en paleta:

- **adelante** — el mockup dibuja affordances que el componente nunca implementó (tabs
  descriptivas, pasos numerados y simultáneos, contador de hallazgos, badges de diferido
  visibles). Por eso se ve mejor: nadie lo bajó a código, se firmó el código sin esa capa.
- **atrás** — su paleta es **pre-rebrand PRENTER** (`--primary` ámbar `#d9a35b`/`#a8742c`),
  mientras el token vigente es `color.semantic.primary` = `brand-teal-500`
  (`web/tokens/base.tokens.json`). El mockup es del 2026-07-14; el rebrand, del 2026-07-15.

Corolario: la queja «no estamos usando lo de Storybook» apunta a un problema real, pero el
culpable no es la implementación — es que **el mockup nunca se volvió a derivar** y la disciplina
del `mockups/INDEX.md` §DoD («todo paquete que toque UI re-deriva el baseline») no se cumplió en
el rebrand.

## 2 · Por qué «el mockup no tiene la parte del marketplace»

Confirmado leyendo y ejecutando el archivo: la rama Marketplace **existe en la fuente**
(`#branch-mkt`, 2 pasos maquetados) pero es **inalcanzable por construcción**:

```html
<div id="branch-mkt" style="display:none">
<div class="wz-tab" id="tab-mkt" aria-disabled="true"
     style="cursor:not-allowed;opacity:.55;pointer-events:none">
```

`pointer-events:none` en la tab + `display:none` en la rama. Fue deliberado (criterio G3 del
Slice 1: nunca pintar un «✓ marketplace válido» sin backend detrás). Forzándola visible por
consola, lo que hay es un **placeholder honesto**: input de git url `disabled`, 3 arneses de
ejemplo con checkbox `disabled`, botón `disabled`, y la nota de acceso `gh`/PAT. Ninguna
pantalla de validación, listado real, ni destino de clon.

**Cómo se vería hoy si se habilitara:** exactamente ese placeholder — dos pasos apagados. No
sirve como línea base de este paquete; hay que diseñarla.

## 3 · Divergencias mockup ↔ Storybook+app (wizard)

Ambas columnas verificadas en vivo. «mockup» = `mockups/arnesia-portafolio.html`;
«código» = story + app (idénticos entre sí).

| # | mockup | código (Storybook = app) | lectura |
|---|---|---|---|
| W1 | tabs = 2 **tarjetas** con título + descripción («Carpeta o repo con arneses ya instalados. Los detectamos para observar / reparar.») | tabs = 2 **links de texto** sin descripción | el operador pierde la explicación de qué hace cada rama justo donde decide |
| W2 | badges de estado **visibles** por opción (`Slice 1` verde, `S2` ámbar) | solo `title=` (tooltip invisible sin hover) | la honestidad existe pero no se ve; contra el espíritu «diferido VISIBLE» |
| W3 | paso 1 y paso 2 numerados y **simultáneos** | máquina de 4 estados, un paso a la vez — y `PasoFuente` **desaparece** en la rama feliz de `candidatos` | tras escanear no podés ver ni corregir la ruta ni re-escanear sin cerrar el wizard. `PasoFuente` solo se re-monta en las ramas `error` y `0 hallazgos` (`portafolio-wizard.tsx:240-275`) |
| W4 | «paso 2 · escaneo — **encontrados 2 arneses**» | sin contador de hallazgos | no sabés cuántos encontró; el botón solo cuenta los *elegidos* |
| W5 | checkboxes **premarcados** → «Agregar 2 al portafolio» | todos **desmarcados** → «Agregar 0 al portafolio» (disabled) | N clicks obligatorios; el caso normal («quiero todo lo que encontraste») cuesta más |
| W6 | lista plana, sin scroll interno | `max-height: 280px` + `overflow-y:auto` dentro de un modal de `max-width: 480px` (`portafolio.css:652,814`) | con 4 hallazgos reales entran 3½ filas y aparece scrollbar interno; las filas son altas porque llevan chips + avisos multilínea |
| W7 | nota de acceso `gh` / PAT fallback en el footer | ausente | esperado: es de la rama no construida |
| W8 | rama Marketplace maquetada (inalcanzable) | tab `disabled` + tooltip | los dos honestos; ver §2 |
| W9 | paleta **ámbar** pre-rebrand | **teal** `brand-teal-500`, dark-first | mockup stale, ver §1 |

## 4 · Divergencias en la lista (toolbar) — incluye un bug real

| # | hallazgo | evidencia |
|---|---|---|
| L1 | **«Marketplace» aparece DOS VECES en la toolbar**, con dos significados distintos y cero distinción visual: una es *lente* (agrupar por marketplace), la otra es *filtro* (acotar por marketplace). Quedan como 6 pills iguales en fila: `Empresa · Plano · Proyecto · Marketplace · Estado · Marketplace` | `portafolio-list.tsx:199-255` — dos `role="group"` con `aria-label` («Lente del Portafolio» / «Filtros del Portafolio») pero **sin rótulo visible**. El mockup sí rotula: `FILTROS […]` y `VER POR […]` |
| L2 | el mockup tiene 3 filtros (`empresa: todas` · `marketplace: todos` · `estado: todos`); el código tiene 2 (Estado, Marketplace) — falta el de empresa | `portafolio-list.tsx:233-254` |
| L3 | el mockup lleva un hint a la derecha («lente para encontrar un arnés cuando hay muchos»); ausente en el código | — |

L1 es defecto de usabilidad, no cosmética: dos controles con la misma etiqueta y distinto efecto.
Un lector de pantalla lo distingue por el `aria-label` del grupo; un ojo, no.

## 5 · Defecto del propio mockup — CORREGIDO en este turno

El archivo **no declaraba charset** y el copy en español salía mojibake (`arnÃ©s`,
`selecciÃ³n`, `â€”`) en cuanto se sirve sin `charset` en el header HTTP. Primer intento de fix
falló por una razón que vale registrar: **`<meta charset>` solo cuenta si cae en los primeros
1024 bytes del documento**, y el comentario de cabecera del archivo pasa de 2 KB. Quedó movido
arriba del comentario, verificado en vivo (`<title>` ahora rinde `ArnesIA · Portafolio v2`).

No es un defecto del diseño; es del artefacto. Se arregló porque bloquea leerlo.

## 6 · Lo que NO se tocó y por qué

Cero cambios en código de app. W1-W6 y L1-L3 son **insumo del spec de este paquete**, no
bugfixes al voleo: la disciplina de paquete (METODOLOGIA §10) manda mockup → decisiones → spec
firmado → implementación. Meter los fixes ahora sería exactamente el orden que ya se corrigió por
retro-ajuste en el Hito 2 del Mapa. Excepción tomada: el charset del mockup (§5), que es
reparación de un artefacto ilegible, no diseño nuevo.

## 7 · Preguntas que el mockup nuevo tiene que contestar

1. **W3** — ¿el wizard consolidado muestra los pasos simultáneos (como el mockup) o sigue siendo
   máquina de estados? Si sigue siendo máquina, ¿cómo se vuelve a paso 1 sin cerrar?
2. **W5** — ¿premarcar los hallazgos por default? Choca de frente con el principio anti-drift
   «el operador siempre confirma» (S1-D2 / PENDIENTE-01 §3). Decisión, no detalle.
3. **W6** — ¿el modal crece, o la lista pagina/virtualiza? Con marketplace en la ecuación un
   listado de 30+ arneses es el caso normal, no el extremo.
4. **L1** — ¿rótulos de grupo visibles, o se renombra una de las dos «Marketplace»?
5. **Reconciliación (PENDIENTE-01 mitad B)** — ¿vive dentro de este wizard como tercer paso, o
   es superficie aparte? Registrar `marketplaces_conocidos` no es «agregar algo al portafolio».
