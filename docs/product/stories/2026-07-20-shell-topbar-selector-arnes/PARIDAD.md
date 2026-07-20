# PARIDAD — Shell · Topbar sin empresa + selector de arnés

> Paquete `2026-07-20-shell-topbar-selector-arnes`. Base: `spec.md` RF-1..17 + `design.md`, sobre
> `mockups/arnesia-shell-topbar-selector-arnes.html` (FIRMADO 🧑‍⚖️ 2026-07-20, commit `5ca97c2`).
> Implementación 100% FE (cero backend Go — el dato ya viaja por `GET /api/portafolio`).
>
> **Método de verificación:** Storybook = SSoT del UI (story = test, corre en Chromium real vía
> `vitest --project=storybook`, con el addon a11y de axe activo). Los `.stories.tsx` renderizan el
> componente con los tokens reales de `theme.css` y ejercitan el comportamiento con `play()`. El
> store se cubre con vitest `unit` (Node). El click-through en vivo de la app instalada + la firma
> quedan como **gate humano** (abajo).

## Estado de la verificación automática

| Suite | Comando | Resultado |
|---|---|---|
| Typecheck | `tsc --noEmit` | ✅ 0 errores |
| Lint / boundaries / FSD / stylelint | `npm run verify` | ✅ verde (depcruise 0 violaciones, fsd sin problemas) |
| Build de producción | `npm run build` | ✅ `built` (98 módulos) |
| Stories del picker | `vitest --project=storybook new-session-picker` | ✅ 7/7 (incl. a11y axe) |
| Store del picker | `vitest --project=unit` | ✅ 5/5 (32/32 en el proyecto unit) |
| Suite completa | `vitest run` | ✅ 168/168, 23 files |
| Capability fitness R1/R2/R4 | `go test ./docs/architecture/fitness -run TestCapability` | ✅ ok (cobertura 100%, punteros resuelven) |

## Trazabilidad RF → verificación

| RF | Qué | Verificación | Estado |
|---|---|---|---|
| RF-1 | empresa fuera del breadcrumb | `topbar.tsx` ya no pinta `active.empresa`; breadcrumb = arnés / vista | ✅ código + typecheck · click-through humano |
| RF-2 | chip de arnés → etiqueta plana (dashed, sin ▾/onClick) | `topbar.tsx` `<span>` borde punteado, `background:none`, sin flecha | ✅ código · click-through humano |
| RF-3 | Conversar a su propia línea (`self-end`) | contenedor `flex-col gap-1.5`; botón `self-end` (ya no `ml-auto`) | ✅ código · click-through humano |
| RF-4 | `quedaste en:` fuera de la Topbar | bloque `active.parked` eliminado | ✅ código · click-through humano |
| RF-5 | picker inline, el rail ensancha a 360px | `session-rail.tsx` `pickerOpen` → `w-[360px]`, oculta lista+pie; cero popover | ✅ código · click-through humano |
| RF-6 | buscador en vivo (id/nombre/empresa) | story **Buscar** (filtra "acme", sin-resultados, limpiar) | ✅ story |
| RF-7 | fuente = Portafolio real (`GET /api/portafolio` vía store) | `portafolio-picker-store.test.ts` (cargando→datos/error) | ✅ unit |
| RF-8 | fila reusa chips de `entities/portafolio` | stories renderizan `EmblemaInicial`/`DerivaChip`/`DotSaludPortafolio` sin drift (scope `.arnesia-portafolio`) | ✅ story (a11y verde) |
| RF-9 | colisión de id → chip distintivo, no bloquea | story **Colision** (2 filas id `dup`, chips home distintos, Crear igual se habilita) | ✅ story |
| RF-10 | selección 2 niveles identidad→copia (0/1/2+) | stories **CasoSimple** (1 copia auto) + **CasoAmbiguo** (3 copias, sub-lista, Crear disabled hasta elegir) | ✅ story |
| RF-11 | `usará <path>` + Crear condicionado | **CasoSimple**/**CasoAmbiguo** asertan resolved + `disabled`→`enabled` | ✅ story |
| RF-12 | Cancelar = cero efectos | story **Cancelar** (dispara `onCancelar`) | ✅ story |
| RF-13 | portafolio vacío → «Ir a Portafolio» | story **PortafolioVacio** (copy + botón, sin buscador) | ✅ story |
| RF-14 | cargando / error + Reintentar | story **ErrorDeCarga** (alert + Reintentar); estado cargando = skeleton | ✅ story (error) · skeleton visual pendiente humano |
| RF-15 | refetch en cada apertura | `portafolio-picker-store.test.ts` (2 `cargar()` → 2 GET) + `session-rail` llama `cargar()` al abrir | ✅ unit + código |
| RF-16 | payload real (arnes/empresa/path, sin puesto) | **CasoSimple** (`{arnes:"harness", path}`) + **CasoAmbiguo** (`{arnes:"acme-cli", empresa:"alpacapurpura", path}`); wiring `onCrear`→`create({...,salud:"info",view:"Mapa"})` en `session-rail` | ✅ story + código |
| RF-17 | reset al cerrar/crear | `session-rail` desmonta el picker + `store.reset()`; story **Cancelar** documenta que el reset local es por unmount | ✅ código · story nota |

## Desviaciones respecto del mockup (documentadas, dentro del espíritu firmado)

1. **Layout del host (TS-D10 ya firmado):** el mockup demuestra el picker en una 2ª columna «canvas»;
   la app lo monta ENSANCHANDO el propio `<aside>` del rail (224→360px, ocultando lista+pie) — más
   inline aún que el mockup (la MISMA caja creciendo, no una vecina). Ya decidido en TS-D10.
2. **Cero CSS nuevo (design §4):** el mockup usa clases `.pf-row`/`.picker-*`; la app las reconstruye
   con utilidades Tailwind sobre los MISMOS tokens (mismo patrón que `topbar.tsx`/`session-rail.tsx`,
   que hoy no tienen hoja `.css` propia). Los chips reusados sí traen sus `.pf-*` — por eso la raíz
   del picker lleva la clase `arnesia-portafolio` (scope de esos estilos).
3. **Contraste AA (design §A11y):** los botones «Reintentar»/«Ir a Portafolio» usan `bg-primary` sólido
   (no `bg-accent-soft` + `text-primary`, que axe rechazó por contraste 2.19 en tema claro); «Limpiar
   búsqueda» pasó a botón bordeado. Cazado por el suite a11y, corregido, ahora verde en ambos temas.
4. **Sub-lista de copias = `<div role="group">`** (no `<ul>`): con `role="group"` el `<ul>` pierde su
   rol de lista y sus `<li>` quedan huérfanos (axe `listitem`). Los botones son los items directos.

## Efecto secundario conocido (fuera de alcance, ya documentado en TS-D16)

`workspace-stage.tsx:235` pinta `{empresa} · {puesto}`; con `puesto` sin escribir por el picker nuevo,
el separador `·` queda colgando para sesiones creadas por el picker. Ese archivo NO está en el alcance
firmado del paquete — queda BACKLOG si molesta en uso real, no se toca de rebote.

## Retomar aquí — GATE HUMANO PENDIENTE 🧑‍⚖️

Falta lo que sólo el operador puede firmar: **correr la app instalada y hacer el click-through lado a
lado contra el mockup**, en AMBOS temas, verificando en vivo:

- [ ] Topbar: sin empresa · chip plano (sin ▾) · Conversar en 2ª línea · sin «quedaste en» — consola limpia.
- [ ] «＋ Nueva sesión» ensancha el rail a 360px inline (sin popup), buscador con foco, lista = tu Portafolio real.
- [ ] Caso simple (1 copia) resuelve solo; caso ambiguo (2+) exige elegir copia; Crear escribe la sesión con arnés+ruta reales.
- [ ] Portafolio vacío → «Ir a Portafolio»; fetch caído → Reintentar.
- [ ] Cancelar restaura el ancho sin crear nada; reabrir re-fetchea.

**Firma del operador (Gate final del paquete):** ☐ (sin marcar — no simulada).

Todo lo automático arriba está verde. La implementación NO se declara «lista» hasta esta firma.
