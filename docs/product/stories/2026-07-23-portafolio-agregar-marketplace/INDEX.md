# Portafolio · Consolidar la agregación de arneses (proyecto + marketplace) — paquete de trabajo

> Ficha del programa «Portafolio · ciclo de vida del arnés» ([`BACKLOG.md`](../../BACKLOG.md) outcome,
> item 2). Origen: `bloqueo(1)` cae — Slice 0 «Cimientos» (HS-23) y Slice 1 «FE» (HS-25) ambos
> FIRMADOS 🧑‍⚖️. Fuente de specs a nivel de reglas (aún no promovidas a este paquete):
> [`../2026-07-10-spike-carga-arneses/spec-funcional.md`](../2026-07-10-spike-carga-arneses/spec-funcional.md)
> §7.2 + [`decisiones.md`](../2026-07-10-spike-carga-arneses/decisiones.md) S-D2/S-D6.
>
> **RE-ESCOPEADO 2026-07-24 (AG-D1, orden del operador):** el paquete ya NO es solo «agregar de
> marketplace» — es la **consolidación de toda la agregación al Portafolio**, por marketplace *y*
> por proyecto. La rama Proyecto ya construida entra al alcance porque la auditoría visual probó
> que driftó contra su propio mockup firmado.

## Flujo y gates (METODOLOGIA §10)

1. **Mockup** — cubre **las dos ramas** del wizard `Agregar al portafolio`. Se deriva de
   **Storybook en tema dark** (SSoT, ver [`mockups/INDEX.md`](../../../../mockups/INDEX.md)); del
   snapshot `arnesia-portafolio.html` se hereda solo la estructura, **jamás la paleta** (AG-D2).
   Superset estricto: nada firmado en la PARIDAD del Slice 1 se quita.
2. **Decisiones** (`decisiones.md`) — una entrada por decisión conversada, EN EL MISMO TURNO.
3. **Specs** (`spec.md` + `design.md`) — 🧑‍⚖️ firma del paquete.
4. **Implementación** contra el spec firmado.
5. **Paridad** (`PARIDAD.md`) — 🧑‍⚖️ gate final.

## Estado

- [x] **Auditoría visual en vivo** — [`auditoria-storybook-vs-app.md`](./auditoria-storybook-vs-app.md)
  (2026-07-24). Mockup · Storybook · app real levantados y comparados a 1440×900.
- [x] `decisiones.md` — **PENDIENTE-02 CERRADA** (las 5 preguntas contestadas). FIRMADAS:
  PENDIENTE-01 (07-23) · AG-D1 (re-escopeo) · AG-D2 (mockup pre-rebrand) · AG-D3 (densidad: modal
  más grande + lazy loading + buscador + filtro) · AG-D4 (rótulos `VER POR`/`FILTROS` visibles) ·
  **AG-D6** (máquina de estados + Atrás + Cancelar) · **AG-D7** (los hallazgos NO se premarcan).
  En mesa: **AG-D8 PROPUESTA** — plano «Marketplaces», propuesta de UX cerrada con criterio
  delegado y anclada en `vision.md` (supersede el borrador AG-D5).
- [x] **Mockup CONSTRUIDO** (2026-07-25) →
  [`mockup-agregacion-consolidada.html`](./mockup-agregacion-consolidada.html). 6 superficies
  navegables, verificadas en vivo a 1440×900. Derivado de **Storybook dark** (tokens = copia literal
  de `web/src/app/styles/theme.css`, clases calcan `portafolio.css`) — no del snapshot ámbar (AG-D2).
  Badges `vigente`/`propuesta` por superficie: el superset queda explícito.
  **Gate 1 🧑‍⚖️ FIRMADO.** Actualizado 2026-07-25 con 2 casos que el spec destapó y el dibujo no
  tenía: **canal duplicado** (AG-D11 — `harness-beta`, chip «mismo contenido que harness») y la
  **6ta rama `no-comparable`** (BR-9, `sin-senal` + motivo). Vocabulario alineado: se cuentan
  **entradas de catálogo**, no arneses. Verificado: 1 `ul`, 7 `li`, contador coincide.
- [x] **Gate 1 🧑‍⚖️ FIRMADO** (2026-07-25) — el operador firmó el dibujo y ordenó pasar a spec.
- [x] **`spec.md` ESCRITO** (2026-07-25) → [`spec.md`](./spec.md). 7 superficies · tokens (ninguno
  nuevo) · átomos/componentes por capa de `fe-taxonomia-componentes` v1.1 · modelo de datos FE ·
  lógica y flujo de las 6 superficies · **12 reglas de negocio** · **30 escenarios con insumo real** ·
  5 capas de validación. Insumo nuevo: 4 hallazgos de la máquina real (AG-D9..D12) que **cambiaron el
  diseño** — el más grande: `known_marketplaces.json` de Claude Code es un 4º eslabón, así que el
  registro de marketplaces es **collect-all** y el plano **nace poblado**, no vacío.
- [x] **`design.md` ESCRITO** (2026-07-25) → [`design.md`](./design.md). Contratos Go completos
  (`domain/marketplace{,_situacion}.go` · 5 puertos · `MarketplaceService` · 6 archivos de adapter ·
  composition root) · 8 endpoints con su JSON y su semántica de status (**400 ≠ 503 ≠ 409**) ·
  persistencia decidida y justificada (JSON atómico en `~/.arnesia/`, **no** SQLite: el índice es
  desechable por doctrina y la clase `propio`/`referencia` no es derivable de nada) · lectura de
  catálogo local-primero + `gh api` (no clone shallow, con la tabla de por qué) · **tabla de verdad
  de 14 filas** para la situación con su precedencia · árbol FE exacto + los 3 gates de lint que el
  diseño casi rompe · máquina de estados de las 2 ramas del wizard. **15 contradicciones detectadas**
  con opciones y recomendación (§2), ninguna resuelta en silencio.
- [x] **`plan-pruebas.md` ESCRITO** (2026-07-25) → [`plan-pruebas.md`](./plan-pruebas.md). Los 30
  escenarios del spec **+ 45 agregados** (E-31..E-75), cada uno con capa · archivo · nombre de test ·
  insumo real · aserción exacta, y para cada agregado **la respuesta del sistema definida** (JSON
  corrupto, permisos, `plugins` ausente/vacío, `source` a ruta inexistente, symlinks, nombres con
  `@`/espacios/unicode, renombres, colisión de nombre, red que corta, techos de bytes/entradas,
  caché stale/corrupto, concurrencia, timeouts). Matriz BR→prueba: las 12 reglas de negocio tienen
  ≥1 prueba.
- [x] **Arquitectura as-code ESCRITA** (2026-07-25): **9 capabilities** nuevas (CAP-102..110,
  `stub` sin `valida:` ⇒ R4 consistente, con los punteros futuros en bloque comentado para no
  romper R1) · boundary **nuevo** `marketplace-referencia-es-solo-procedencia` (L1 bounded context
  + anti-corruption layer, fuentes **verificadas en vivo**) · `portafolio-identidad-y-deriva-honesta`
  **v1.1→v1.2** (+4 checks: el mismo L1 aplicado al estante) · `fe-taxonomia-componentes`
  **v1.1→v1.2** (+1 check `no-sibling-widget-imports`, **ya aplicado a `web/.dependency-cruiser.js`
  y verificado verde**) · `docs/architecture/INDEX.md` 21→22 boundaries, 92→101 checks · cifras del
  checkpoint **regeneradas** (`estado.sh`) y `INDEX.md` de capabilities regenerado (`cap_doctor.py
  --index`). Verificado: `go test ./docs/architecture/fitness/...` verde (R1/R2/R4) ·
  `conformance --todo` = 275 checks · **fail 0** · `estado.sh --check` en sync.
- [x] **Revisión del diseño + resolución de las 15 contradicciones** (2026-07-25). Los verdes del
  arquitecto **re-verificados a mano** (no por confianza): `conformance --todo` 275 · pass 57 · **fail
  0** · `pnpm run verify` 5 gates verdes (depcruise 122 módulos/314 deps · steiger limpio · stylelint
  limpio) · `go test ./docs/architecture/fitness/...` ok. Resoluciones:
  - **C12 NO existía** — `AG-D11` está FIRMADA en `decisiones.md` (línea 425); el arquitecto leyó una
    copia stale mientras se registraba la firma. Punto de cambio de `design.md` §8.7 **CERRADO**.
  - **C7 confirmada** — el stale era mi `spec.md §4.3`, no el mockup. Corregido, con nota de por qué.
  - **C1 bajada al spec como normativa** — `situacionDe` vive en el **dominio Go**; el selector FE solo
    presenta. Evita el cross-import entity↔entity que `steiger` rompe.
  - **BR-2 v2 + BR-2b** por AG-D14/D13; **AG-D13..D16 RATIFICADAS** (evidencia, no preferencia).
  - **C5 cerrada en el dibujo** — `Publicar`/`Actualizar mi copia`/`Reparar` ahora con el par
    `disabled`+tooltip que BR-10 exige, para que el constructor no copie la affordance equivocada.
  - C3, C4, C6, C9, C10, C15: recomendaciones del arquitecto **aceptadas** tal cual.
- [x] **AG-D17 FIRMADA 🧑‍⚖️** (2026-07-25) — **`↧ Traer canónico` entra COMPLETO**, los dos caminos
  (A local: copia de subcarpeta del checkout que CC ya tiene, sin red · B externo: clone shallow por
  `ref` + verificación de `sha` + extracción de `path`, auth `gh`→PAT). Resuelve C2, que era un hueco
  de **mi** spec: el ítem 2 del `BACKLOG` pedía el checkout literal y el mockup lo dibuja como acción
  principal, pero ningún §4 lo especificaba. Ya en `spec.md`: superficie **S8** · **§4.7** con los dos
  caminos y la secuencia atómica · **BR-13..BR-18** · escenarios **E-76..E-90**.
- [ ] `design.md` §13 + `plan-pruebas.md` de `Traer canónico` — **en curso** (arquitecto retomado con
  el contexto cargado).
- [x] **`design.md` §13 — mecanismo de `↧ Traer canónico`** (2026-07-25, habilitado por **AG-D17
  FIRMADA 🧑‍⚖️**): contratos Go (`domain/traer.go` puro · `ports.Materializador` · `Traer` en el
  usecase · 3 adapters), endpoint `POST /api/marketplaces/{nombre}/traidos` con 7 códigos y su
  razón, **atomicidad de BR-15** con las 6 ramas de falla tabuladas, BR-13 como función pura +
  aserción de path, y la máquina de estados de las dos puertas. **Verificado contra GitHub EN VIVO:**
  el mecanismo de camino B (`git init` + `sparse-checkout --cone` + `fetch --depth 1 --filter=blob:none
  origin <sha>` + `checkout FETCH_HEAD`) da `HEAD == sha` exacto en 664 KB. Tres correcciones a la
  letra del spec salieron de esa verificación (**C16/C17/C18**).
- [ ] **Firma 🧑‍⚖️ del par `spec.md` + `design.md`** — es lo que falta para habilitar la etapa 4.
- [ ] Implementación — no arrancada (código NO se toca hasta specs firmados; los hallazgos
  W1-W6/L1-L3 de la auditoría son insumo del spec, no bugfixes al voleo).
- [ ] `PARIDAD.md` — no arrancado.

## Qué encontró la auditoría (resumen; detalle + evidencia en la hoja)

- **La app SÍ sigue al Storybook.** Story `paso-1-fuente` con `globals=theme:dark` sale
  pixel-idéntico al wizard de la app. Mismo componente, mismo scope CSS, cero juego paralelo.
- **Lo que driftó es el mockup**, en dos direcciones: **adelante** en diseño (tabs descriptivas,
  pasos numerados simultáneos, contador de hallazgos, badges de diferido visibles — nunca bajaron
  a código) y **atrás** en paleta (ámbar pre-rebrand vs `brand-teal-500` vigente).
- **La rama Marketplace del mockup es inalcanzable por construcción** (`display:none` +
  `pointer-events:none`, criterio G3 del Slice 1) — por eso «el mockup no tiene marketplace».
  Forzada visible, es un placeholder honesto de 2 pasos apagados. No sirve como línea base.
- **Bug real en la lista:** «Marketplace» aparece dos veces en la toolbar (lente y filtro) sin
  distinción visual — 6 pills idénticas en fila.
- **Defecto del artefacto, ya corregido:** el `.html` no declaraba charset y el copy salía
  mojibake. Gotcha: `<meta charset>` solo cuenta en los primeros 1024 bytes del documento.

## Retomar aquí

> Se actualiza al cierre de cada turno de trabajo (METODOLOGIA §10).

- **2026-07-24 — auditoría hecha y paquete re-escopeado.** El operador ordenó consolidar toda la
  agregación (AG-D1). Se auditó en vivo (no por lectura de código): la app respeta el Storybook;
  el mockup es lo que está stale, y encima es pre-rebrand (AG-D2). Charset del mockup reparado.
  `mockups/INDEX.md` re-estampado con la fila real. Cero cambios en código de app — a propósito.
- **2026-07-24 (2° turno) — 3 de las 5 preguntas contestadas.** AG-D3 (densidad) y AG-D4 (rótulos)
  FIRMADAS por el operador. Para §5 el operador propuso **«una pestaña a la altura de los arneses»**
  → desarrollado como **AG-D5 (PROPUESTA)**: el Portafolio gana dos planos hermanos, `Arneses` +
  `Marketplaces`; el wizard deja de intentar listar catálogo dentro del modal y se queda con
  **registrar** el marketplace, aterrizando después en el plano. Esto le da casa a la mitad B de
  PENDIENTE-01 (reconciliación) y a los ítems 3-5 del outcome, que hoy no tienen dónde vivir.
- **2026-07-25 — PENDIENTE-02 cerrada; queda una sola firma.** §1 → AG-D6 (máquina de estados con
  Atrás/Cancelar; el mockup NO se calca acá, a propósito). §2 → AG-D7 (sin premarcar; se descartó la
  propuesta intermedia). §5 → el operador delegó el criterio UX pidiendo anclarlo en la visión →
  **AG-D8**, que reencuadra el plano: leyendo `vision.md` («muere operar arneses de terceros» ·
  marketplace = seam de entrega · ArnesIA dueña única del modificar) el plano **no es una tienda**,
  es *el estante de lo que vendemos y el espejo de si el cliente coincide*. De ahí salen 8
  decisiones, incluida la partición **propio / de referencia** que el borrador AG-D5 no tenía y sin
  la cual el plano degeneraba en gestor universal de plugins.
- **2026-07-25 (3° turno) — AG-D8 FIRMADA 🧑‍⚖️ y MOCKUP CONSTRUIDO.** Las 6 superficies dibujadas y
  verificadas en vivo. Se cazaron 2 defectos en el propio mockup mientras se revisaba: prosa partida
  en línea por `column-flex` (celda `.mk-pendientes`) y «Confirmar origen» habilitado sin selección —
  los dos corregidos. Registrado en `mockups/INDEX.md` (DoD de re-derivación cumplido).
  Cómo verlo: `python3 -m http.server 8100 -d docs/product/stories/2026-07-23-portafolio-agregar-marketplace`
  → <http://127.0.0.1:8100/mockup-agregacion-consolidada.html>
- **2026-07-25 (4° turno) — ETAPA 3 COMPLETA: `design.md` + `plan-pruebas.md` + arquitectura as-code.**
  El diseño se escribió **verificando contra la máquina real, no contra el spec**, y eso destapó
  cuatro hechos que corrigen AG-D10 sin invertir su espíritu (propuestos como **AG-D13..AG-D16**,
  `design.md` §1): `plugins[].source` tiene **4 formas** reales (220 de 273 entradas del catálogo
  oficial son OBJETO `git-subdir`/`url`/`github`, no ruta relativa) · **`version` SÍ es campo
  estándar** del formato (259 `null` explícitos + 14 semver reales, y hay `$schema` oficial), así
  que BR-2 pasa de derivación exclusiva a **precedencia con procedencia anotada** · el catálogo
  ajeno trae 13 claves opcionales más `renames`, `owner.url` y `description` top-level · y el
  catálogo oficial son **273 filas / 159 KB**, no 41 — lo que vuelve el lazy un requisito y
  justifica el caché como O(1) por fila del plano.
  Tres decisiones de fondo que el spec no podía tener: **(a)** la situación se calcula en el
  **dominio Go** y viaja en el wire, porque cruzarla en el FE exigiría un cross-import
  entity↔entity que `steiger` rompe (y `fsd` corre dentro de `verify`); **(b)** el cruce estricto
  por `(home,id)` **habría dado `no-lo-tengo` en toda la columna** contra el dato real (la
  instalación dominante no tiene sello ⇒ identidad provisional), así que se agregan las vías
  `faceta-registry` y `rename` con el CÓMO anotado y mostrado; **(c)** el `home` de la
  reconciliación **no puede ir al sello**: la instalación dominante vive bajo `~/.claude`, que
  `validarRootPortafolio` prohíbe escribir — va al store, con eslabón `declarado-por-operador`.
  **15 contradicciones** documentadas (§2). Dos de ellas necesitan al operador: si `↧ Traer
  canónico` entra o no en alcance (C2), y el **estado de firma de AG-D11**, que figura FIRMADA en
  el HTML del mockup y en `mockups/INDEX.md` pero PROPUESTA en `decisiones.md` (C12). Una es una
  corrección al propio spec: **§4.3 afirma que el mockup no dibuja `no-comparable` ni el chip de
  canales, y el mockup SÍ los dibuja** (C7) — el spec quedó de una versión anterior del dibujo.
  As-code aplicado y **verificado en vivo**: `go test ./docs/architecture/fitness/...` verde
  (R1 archivo+símbolo · R2 cobertura · R4 estado+puntero) · `conformance --todo` = **275 checks ·
  fail 0** (los 9 nuevos suben total y `deferred`, que **no** es regresión) · `depcruise` verde con
  la regla nueva (122 módulos / 0 violaciones) · `estado.sh --check` en sync tras regenerar.
- **2026-07-25 (5° turno) — `↧ Traer canónico` bajado a mecanismo (`design.md` §13).** Las tres
  preguntas del turno anterior quedaron cerradas por el operador: **C2 → AG-D17 FIRMADA** (Traer entra
  **completo**, los dos caminos; la asimetría que lo justifica es que Traer escribe **solo** en
  `~/.arnesia/checkouts/`) · **C12 no existía** (era lectura stale mía: AG-D11 está FIRMADA, así que
  el punto de cambio de §8.7 queda **CERRADO** — no se deduplica por `source`) · **AG-D13..D16
  RATIFICADAS**. Y **C7 se corrigió en el spec**: el dibujo sí trae `no comparable` y el chip de
  canales; el stale era el spec.
  Diseñar el mecanismo exigió volver a la máquina y a GitHub, y ahí salieron **tres correcciones a la
  letra del spec** que un implementador habría descubierto en runtime: **(C16)** el pin de camino B es
  el **`sha`**, no el `ref` — en `42Crunch-AI/claude-plugins` el `ref v1.5.5` resuelve a `faf53053…`
  y el `sha` declarado es `30287f5e…`, **commits distintos**: «clonar por `ref` y verificar `sha`»
  abortaría en la primera entrada real, y además `ref` está en 77/220 contra `sha` en 220/220;
  **(C17)** la identidad del canónico va con `Home = CanonicalizarRepo(repo)` mientras el **path** usa
  el nombre legible del marketplace — colapsarlos haría que el canónico recién traído **no cruzara con
  la fila del catálogo de la que vino**, o sea `no-lo-tengo` sobre algo que acabás de traer;
  **(C18)** `source: "github"` trae **dos** hashes válidos (`commit` y `sha`) sin semántica
  documentada ⇒ gana `sha` (el único presente en las 220 formas objeto) y la divergencia se muestra.
  El mecanismo se **probó en vivo** antes de escribirlo: `HEAD == sha` exacto, 664 KB, solo la
  subruta. As-code: **CAP-111** + 5º check del boundary nuevo (`traer-referencia-rechazado`: con una
  acción que materializa en disco, «referencia es read-only» tiene que vivir en el dominio — un POST
  puede llegar sin pasar por el botón). Cifras regeneradas: **276 checks · fail 0** · 111 capabilities.
- **Próximo paso concreto:** **firma 🧑‍⚖️ del par `spec.md` + `design.md`** (etapa 3 → 4). No quedan
  preguntas de producto abiertas; sí tres correcciones editoriales al `spec.md` que salen de C16/C17/C18
  (la redacción de §4.7 sobre `ref`/`sha`, la precisión de «(nombre del marketplace, nombre de la
  entrada)» en el paso 5, y el caso de los dos hashes) — no cambian ninguna decisión, alinean el texto
  con lo verificado. Firmado eso, la etapa 4 arranca por **T1** del orden de construcción
  (`design.md` §12.1: la tabla del dominio primero) y termina en **T21** (E2E + `PARIDAD.md`).
  Insumo para quien implemente: `design.md` (contratos exactos, §13 para Traer) + `plan-pruebas.md`
  (aserciones exactas, §2.5 para S8) + **`design.md §12.2`: los 17 riesgos concretos**, empezando por
  «`shared/ui` no puede importar `entities/*`» y, para Traer, «el temporal DEBE vivir bajo
  `~/.arnesia/`, no en `/tmp`» (un `os.Rename` cross-device rompe la atomicidad en la máquina del
  operador, no en el CI).
