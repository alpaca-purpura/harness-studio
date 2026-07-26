# mockups/ — INDEX · línea base del diseño (LÉEME ANTES DE DISEÑAR)

> vig: activo · dueño: cualquiera que **invoque creación/edición de diseño** (mockup, UI nueva,
> propuesta visual). **Leer esto ANTES de forkear cualquier `.html`.** Este registro existe para que
> una propuesta de diseño no vuelva a borrar/reinventar cosas ya construidas (incidente 2026-07-10).

## Regla dura (SSoT del UI)

1. **El SSoT del UI vigente es Storybook**, no estos `.html`. La superficie completa del Mapa vive en
   `web/src/widgets/map-canvas/ui/map-canvas.stories.tsx` («the full map surface — fitness fixture of
   record; the signed mockup is derived from what renders here») + `inspector.tsx` · `handoff-gutter.tsx`
   · `map-bar.tsx`. Se ve con `pnpm --dir web storybook`.
2. **Estos `.html` son SNAPSHOTS DERIVADOS y fechados** — pueden estar stale. Nunca forkear uno sin
   mirar su `estado` abajo. El baseline vigente se re-deriva de las stories.
3. **Toda propuesta de UI = superset estricto** del vigente. Jamás quitar un artefacto firmado (PARIDAD).
4. **No pisar vocabulario L0.** Antes de nombrar un concepto nuevo, verificar `internal/domain/box.go`
   + `web/src/entities/arnes/`. Ya tomados (con OTRO significado): `procedencia` (honestidad del dato:
   medido·estimado·declarado·inferido·no-declarado) · `origen` (estandar·del-puesto, lo estampa el
   provisioner) · `canal` (beta·estable·propuesto·deprecado) · `insumos` (inputs del contrato `Necesita`) ·
   `banda` (7 bandas). El drift/autoría por-archivo y el conocimiento-del-proyecto **necesitan nombre propio.**
5. **DoD:** todo paquete que toque UI **re-deriva** el baseline `.html` y actualiza su fila aquí
   (fecha + commit de sincronía). Ancla: la norma «pegarse al Storybook / tokens DTCG reales».

## Baseline VIGENTE

- **`arnesia-mapa-baseline.html`** — Mapa, superficie completa. **Derivado de las stories @`a01d845`
  (2026-07-10).** Incluye: barra (picker · META · toggle Artefactos off/auto/todos · conmutador de
  capas estructura/tokens·perf·proceso) · Guardia/Proceso/Base+bandas · edges invoca/lee/escribe ·
  **franja de artefactos** (chips hand-off + refs ↖) · **inspector 3 tabs** (Resumen·Contenido·Corridas)
  con contrato fusionado + botonera staged. **← Forkear DESDE AQUÍ para trabajar el Mapa.**

## Registro de mockups

| mockup | fecha (alta→últ) | rol | spec / paquete | estado |
|---|---|---|---|---|
| `stories/2026-07-24-telemetria-embebida-otel/mockup-capa-mejora.html` | 2026-07-26 | **Capa «Mejora» del Mapa** (ex «capa Tokens»): el join dinero × proceso. 8 secciones: **lo vigente calcado** (barra con las 3 capas disabled) · **la barra con la capa activa** (ventana temporal · total con disclaimer «estimado, no facturación» · barra de cobertura exacta/por-huella/sin-dato) · **el canvas** con las 3 marcas nuevas (participación · cifra+% · marca de fuga con el detector nombrado) y **«sin dato atribuible»** en los nodos no atribuibles · **2 tarjetas de punto de mejora** (contrafactual · umbral algebraico citado · **sesgo declarado en contra** · un fix · score v1) · **4ª tab del inspector** (buckets con «no aplica»≠0 · reportado vs. calculado = el oracle A6 a la vista · el join a nivel nodo · detectores con los que NO aplican y por qué) · **tarjeta del Portafolio** (arnés × puesto) · **los 7 estados honestos** · **tabla de traza** contra la señal medida. Tokens = copia literal de `web/src/app/styles/theme.css` (PRENTER) — **el baseline del Mapa quedó en la paleta ámbar pre-rebrand**, desviación declarada en la cabecera. Cifras: forma real, montos ilustrativos anclados en la medición propia (USD 0,0185/turno con haiku) | `stories/2026-07-24-telemetria-embebida-otel/INDEX.md` (`decisiones.md` D14-D17 · `verificacion-2026-07-26/INFORME.md`) | ⏳ **iteración 1, sin firmar.** 2 desviaciones declaradas: renombra el slot `Tokens`→`Mejora` del conmutador firmado (D17.1) y usa tokens PRENTER en vez de los del baseline. Gaps dibujados como tales: el hook de gate no existe (§8) y el catálogo de precios está sin construir |
| `arnesia-voz-dictado.html` | 2026-07-25 | **Dictado por voz en el composer del chat dock.** 6 secciones: **lo vigente** calcado y etiquetado (reposo · turno en vuelo) · **el botón** (reposo con mic · escuchando con medidor de nivel + contador 0:24/3:00 · cerca del tope en `warn` · cortado por el tope) · **las etapas con nombre** (transcribiendo ~2 s · ordenando 13-17 s, con «usar el crudo» en vivo) · **el resultado** (limpio · **crudo marcado en `warn`** · falló-una-etapa sin pisar lo tecleado) · **4 motivos distintos de deshabilitado** (permiso denegado · sin dispositivo · **sin motor de STT en `$PATH`** con qué instalar · formato no soportado) · **tabla de traza RF→superficie**. El baseline NO es un `.html`: es `web/src/widgets/chat-dock/ui/chat-dock.tsx#Composer` (SSoT = Storybook) — se calca literal (textarea auto-grow tope 3 líneas, botón ↑, botón ■, los 3 placeholders exactos) y el mic entra como **un botón más**, sin cambiar el layout de la fila. Tokens = copia literal de `web/src/app/styles/theme.css` (dark-first post-rebrand PRENTER). Badges `vigente`/`propuesta` por panel para que el superset sea explícito. Vocabulario nuevo verificado contra L0: «dictado» · «crudo» · «limpieza» no chocan con procedencia/origen/canal/insumos/banda | `stories/2026-07-25-spike-voz-dictado/INDEX.md` (`spec.md` RF-215…RF-228 · `decisiones.md` V-D1..V-D8) | 🧑‍⚖️ **decisiones V-D1..V-D8 FIRMADAS 2026-07-25** (2ª ronda). Desbloquea los RF 🎨 (216 · 219 · 226). **RF-215 no tiene superficie a propósito** (puente Rust, invisible, va primero). Gaps dibujados como tales: **T6** (voz real del operador contra el binario instalado) y **T7** (qué motor STT se empaqueta) |
| `stories/2026-07-23-portafolio-agregar-marketplace/mockup-agregacion-consolidada.html` | 2026-07-25 | **Agregación consolidada al Portafolio** (proyecto + marketplace). 6 superficies navegables: plano **Arneses** · plano **Marketplaces** (4 filas: propio-legible · propio-sin-acceso · de-referencia · no-leído) · **catálogo propio** (**6** situaciones → acción por situación: Traer / nada / Publicar / Actualizar / Reparar / **no-comparable**) · **catálogo de referencia** (read-only, Traer disabled) · **wizard** 3 estados con Atrás+Cancelar · **diálogo de reconciliación**. Actualizado 2026-07-25 tras firmar AG-D11: fila = **entrada de índice** (canal) — `harness-beta` con chip «mismo contenido que harness», y la 6ta rama `no-comparable` con `sin-senal`+motivo (BR-9, obligatoria: sin ella el sistema tendría que mentir o callarse). Derivado de **Storybook dark** (tokens = copia literal de `web/src/app/styles/theme.css`; clases calcan `portafolio.css`) — NO del snapshot ámbar pre-rebrand. Lleva badges `vigente`/`propuesta` por superficie para que el superset sea explícito | `stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md` (`decisiones.md` AG-D1..D8 + PENDIENTE-01/02; auditoría en `auditoria-storybook-vs-app.md`) | 🧑‍⚖️ **Gate 1 FIRMADO 2026-07-25** (`decisiones.md` §GATE 1); AG-D8 (modelo del plano) y AG-D11 (fila = entrada de índice) también FIRMADAS. **Etapa 3 escrita** (`spec.md` + `design.md` + `plan-pruebas.md`) — falta la firma del par spec+design. Verificado en vivo a 1440×900, las 6 superficies |
| ↑ `mockup-agregacion-consolidada.html` · **nota de la etapa 3 (2026-07-25)** | — | **Estado real: Gate 1 🧑‍⚖️ FIRMADO** (`decisiones.md` §GATE 1) — la celda de estado de arriba decía «esperando Gate 1» y quedó stale; etapa 3 (`spec.md` + `design.md` + `plan-pruebas.md`) **escrita**. Tres cosas verificadas por lectura directa del HTML que hay que saber antes de forkear: **(1)** el dibujo **SÍ** trae la 6ta rama `no comparable` (línea 670, con `sin-senal`) y **SÍ** trae el chip de canales `mismo contenido que harness` (línea 618) — el que quedó stale en este punto es `spec.md §4.3`, que afirma lo contrario (`design.md` §2 C7). **(2)** Los botones `Publicar`/`Actualizar mi copia`/`Reparar` y el `↧ Traer canónico` del catálogo **propio** están **sin `disabled` y sin `title`**; BR-10 exige el par, y el spec gana — la única fila con el par correcto es la de `referencia` (`disabled title="no aplica: solo arneses propios"`, línea 712), cuyo literal se conserva tal cual (`design.md` §2 C5, §6.3). **(3)** El dibujo **no** tiene la pantalla de éxito de `Validar` (`name`+`owner`+nº de `plugins[]`) que `spec.md §4.5.2` y E-15 exigen: superset a construir (`design.md` §2 C6, §9.2). ✅ **C12 RESUELTA (no existía):** `AG-D11` está **FIRMADA 🧑‍⚖️** en `decisiones.md` (línea 425, firmada por el operador tras comparar las dos opciones renderizadas) — el arquitecto leyó una copia stale del archivo mientras se registraba la firma. No hay superposición y el punto de cambio de `design.md` §8.7 queda **cerrado**: NO se deduplica por `source`. Las cifras del dibujo (6/8/41 entradas) son ilustrativas: las reales son 2 y **273**, salen del wire y el FE no teclea ninguna (`design.md` §2 C14) | `stories/2026-07-23-portafolio-agregar-marketplace/design.md` §2 «Contradicciones detectadas» | 📐 **etapa 3 escrita — esperando firma 🧑‍⚖️ del par spec+design** |
| `stories/2026-07-23-boton-correr-caja/mockup-boton-correr.html` | 2026-07-23 | Inspector · tab Corridas: botón real que dispara `POST …/boxes/{id}/run` — 6 estados (idle interactivo · corriendo · éxito · éxito+advertencias · error 409 · error 500 · no-caja), superset estricto (prosa placeholder Hito 3 intacta) | `stories/2026-07-23-boton-correr-caja/INDEX.md` (`decisiones.md` D1-D4) | ⏸ **PAUSADO** — operador pidió explicación funcional antes de firmar, dada, eligió pausar sin rechazar |
| `arnesia-chat-dock-ux.html` | 2026-07-22 | Chat dock legible: tarjeta de actividad desplegable + burbuja por paso + markdown renderizado — superset estricto del dock vigente (header/sesión/Alcance/permiso/composer intactos) | `stories/2026-07-22-chat-dock-ux/` (CH-D1..D6 + D4b) | ✅ **construido + PARIDAD FIRMADA 🧑‍⚖️** (HS-26, 2026-07-22; SSoT = código vivo `chat-dock/ui/`; desviación aceptada: tarjeta sin segundos) |
| `arnesia-shell-topbar-selector-arnes.html` | 2026-07-20 | Topbar sin empresa hardcodeada + selector de arnés al abrir sesión (lee el Portafolio real, empresas N:M, canónico/instalaciones, deriva) | `stories/2026-07-20-shell-topbar-selector-arnes/` (Gate 1 firmado 2026-07-20; `decisiones.md` TS-D1..D9) | 🧑‍⚖️ **Gate 1 firmado** — spec/build siguen |
| `arnesia-mapa-baseline.html` | 2026-07-10 | **BASELINE** del Mapa (superficie completa) | este INDEX + `map-canvas.stories.tsx` | **✅ vigente** |
| `arnesia-portafolio.html` | 2026-07-13→14 (audit 07-24) | Portafolio (front-door del ciclo de vida del arnés) — snapshot derivado | `stories/2026-07-13-portafolio-slice1-fe/` (SSoT = stories `portafolio-*` @`dfa82b5`) · auditoría 2026-07-24 en `stories/2026-07-23-portafolio-agregar-marketplace/auditoria-storybook-vs-app.md` | ⚠️ **STALE en paleta, ADELANTADO en estructura — NO forkear los colores.** `--primary` ámbar `#d9a35b`/`#a8742c` = **pre-rebrand PRENTER** (mockup 07-14, rebrand 07-15); token vigente = `color.semantic.primary` → `brand-teal-500`. Auditado en vivo: la app **sí** respeta el Storybook (story `paso-1-fuente` dark ≡ app, pixel), el que driftó es este `.html`. Dibuja affordances que el código nunca bajó (tabs descriptivas · pasos numerados simultáneos · contador «encontrados N» · badges `Slice 1`/`S2` visibles) → insumo estructural del paquete `2026-07-23-portafolio-agregar-marketplace`. Rama Marketplace presente pero **inalcanzable por construcción** (`display:none` + `pointer-events:none`, G3). Charset arreglado 07-24 (faltaba `<meta charset>`; ojo: solo cuenta en los primeros 1024 B). Gate humano 🧑‍⚖️ de `paridad.md` del Slice 1 PENDIENTE |
| `arnesia-mapa-mvp.html` | 2026-07-06 | Mapa MVP Hito 1 (canvas bandas/carriles + SVG) | `stories/2026-07-06-mapa-mvp/{00-BRIEF,spec,design}.md` (Gate 1 ✓ `0736d2c`) | ⚠️ **superado** — le faltan franja-artefactos + inspector 3-tabs. NO forkear |
| `arnesia-mockup-v3.html` | 2026-07-04→05 | Shell + detalle (capas/inspector) it.13 | `ux.md` Inventario final §I · `stories/2026-07-07-inspector-drawer/analisis-drawer-v3.md` | 📎 referencia (detalle inspector) |
| `arnesia-shell-A-galaxia.html` | 2026-07-05 | Shell A «galaxia» it.13 — **fuente de VALORES de tokens** | `ux.md` Inventario final · `architecture/boundaries/fe-tokens-contrato.md` | 🔒 firmado (tokens) |
| `arnesia-shell-A-sessions.html` | 2026-07-05 | Multisesión it.14 | `stories/2026-07-08-chat-cc-funcional/INDEX.md` · `ux.md` | 📎 referencia (shell/sesiones) |
| `arnesia-shell-lab.html` | 2026-07-05 | Laboratorio de shell (4 paradigmas; A firmado) | `ux.md` Inventario final (it.13) | 🗄️ histórico (decisión de shell) |
| `arnesia-session-lab.html` | 2026-07-05 | Laboratorio de sesiones | — (exploración) | 🗄️ histórico |
| `arnesia-mockup-v2.html` | 2026-07-04 | Shell v2 (superado por it.13/14) | `ux.md` | 🗄️ histórico |
| `arnesia-arch-inyeccion-knowhow.html` | 2026-07-06→08 | Diagrama de **arquitectura** (inyección know-how HS-07/10) — no es UI de producto | `architecture/` · memoria HS-07/10 | 📎 referencia (arquitectura) |

> **Shell — Topbar sin empresa + selector de arnés (Gate 1 firmado 2026-07-20):**
> `arnesia-shell-topbar-selector-arnes.html` — quita `active.empresa` fijo del breadcrumb (Topbar),
> baja Conversar a línea propia, y mueve la elección de arnés al momento de abrir sesión («＋ Nueva
> sesión»), leyendo el **Portafolio real** (`GET /api/portafolio`, no un catálogo aparte) con buscador
> y selección en 2 niveles (identidad → copia, cuando hay 2+ instalaciones). Toca CAP-72
> (`rail-de-sesiones`) y CAP-74 (`topbar-breadcrumb-k-dock`). Ver
> [`stories/2026-07-20-shell-topbar-selector-arnes/decisiones.md`](../docs/product/stories/2026-07-20-shell-topbar-selector-arnes/decisiones.md)
> (TS-D1..D9). Spec + build siguen.
>
> **Co-diseño en curso (2026-07-10):** «Mapa DESTINO» (sello · llenado/deriva · slots con receta ·
> conocimiento-del-proyecto) se está diseñando como **superset del baseline vigente**. Mientras no
> esté firmado NO reemplaza el baseline; vive como propuesta. Al firmar → se funde y se re-estampa aquí.
>
> **Rebrand PRENTER en curso (2026-07-15):** propuesta de reemplazo de VALORES de tokens (chrome
> estructural: colores/tipografía/radios/sombras, dark-first) extraída de Claude Design — **NO toca**
> `color.kind`/`health`/`heat` (paleta funcional del Mapa, intacta). Afecta directo el estado
> `🔒 firmado (tokens)` de `arnesia-shell-A-galaxia.html` abajo — mientras no haya gate humano, ese
> renglón sigue vigente tal cual. Ver
> [`stories/2026-07-15-rebrand-prenter-design-system/INDEX.md`](../docs/product/stories/2026-07-15-rebrand-prenter-design-system/INDEX.md).
>
> **Portafolio — CONSTRUIDO (Slice 1 FE, 2026-07-13→14):** «**Portafolio**» (`arnesia-portafolio.html`) —
> front-door del ciclo de vida. **3 superficies navegables:** Lista (por empresa, fila compacta) → Drawer
> detalle → Wizard agregar. Modelo FIRMADO S-D8: unidad = **arnés por identidad** (1 card), con **canónico**
> (única copia editable) + **instalaciones** (espejos read-only); **ley anti-drift** visible; relación
> arnés–empresa–marketplace = **N:M:M** (empresa/marketplace = lentes). Reusa tokens + familia de tarjeta del
> `.node`. **Naming RESUELTO (S-D10/F-C):** «drift» → **`deriva`** (`en-deriva`/`al-hilo`/`deriva-no-evaluable`).
> Los 9 hallazgos G1-G9 de `revision-adversaria.md` §G quedaron RESUELTOS por el build (Slice 0 cimientos →
> Slice 1 FE, `stories/2026-07-13-portafolio-slice1-fe/`): el mockup `.html` se corrigió en sitio (T8, S1-D12) y
> la SSoT real pasó a Storybook (25 stories `play()` en `widgets/portafolio/ui/*.stories.tsx` +
> `entities/portafolio/ui/chips.stories.tsx`). Gate humano 🧑‍⚖️ de `paridad.md` PENDIENTE.
>
> ⚠ **AUDITADO EN VIVO 2026-07-24** (`stories/2026-07-23-portafolio-agregar-marketplace/auditoria-storybook-vs-app.md`):
> el `.html` **quedó stale un día después** de firmarse — es del 07-14, el rebrand PRENTER es del 07-15, así
> que su paleta ámbar ya no es la del producto. Y en el otro sentido quedó **adelantado**: dibuja affordances
> que el componente nunca implementó (tabs con descripción · pasos numerados y simultáneos · «encontrados N
> arneses» · badges de diferido visibles). La app **sí** respeta el Storybook — se verificó pixel a pixel con
> la story en dark. Este es el caso testigo de por qué existe el **DoD de re-derivar** (regla dura 5): el
> rebrand tocó UI y no re-estampó este archivo, y nueve días después se leyó el mockup como si fuera vigente.
