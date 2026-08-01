# Spec — crear · visualizar · mantener arneses multi-actividad (esquema v3) + dogfood `developer-vitalia`

> vig: escrita 2026-08-01, **AUDITADA Y CORREGIDA contra el sistema real 2026-08-01**
> (v2; hallazgos AUD-1..AUD-9 en `decisiones.md`) · **FIRMADA 🧑‍⚖️ 2026-08-01 junto
> al gate del mockup (DEF-D8)** · deriva de DEF-D1..D3 FIRMADAS (`decisiones.md`)
> + directivas DEF-D5/D6 del operador (round 3, `chris-input.md`).
> **Nota §10:** el orden mockup→spec se invirtió por directiva («hazlo en un spec e
> intégralo al MVP ahora mismo»); el mockup (MA-T6) queda como **gate 🧑‍⚖️ previo a
> construir FE** — la disciplina se conserva, solo cambió el orden de escritura.
> Prefijo propio `MA-` (no colisiona con RF-NNN globales).

## §0 · Objeto

Con v3 firmada, un arnés agrupa **N actividades** (cada una con su procedimiento)
sobre una **base común** (normas transversales: arquitectura, seguridad, calidad —
lo que «afecta a nivel proyecto más allá de un paso»). Esta spec define cómo eso se
**crea** (forja), se **visualiza** (Mapa sin saturación: el todo → foco por
actividad) y se **mantiene** (mejora continua con radio de impacto), y fija el
dogfood real: **`developer-vitalia`**, primer arnés forjado con el esquema v3.

## §1 · Modelo de datos (el seam) — CORREGIDO contra el código real (AUD-1..AUD-4)

- **Fuente de actividades — estructura REAL de `arnes.yaml`** (verificada en
  `semilla/arnes.yaml` L145-156/L169-180 y `internal/adapters/forja/parser.go:36-64`):
  `gestion_trabajo.tipos_paquete.<id>` = `{jerarquia, estados, wip_caps?, cierre}`
  — el **id es la CLAVE** del mapa (no hay campo `id` ni `label`; el campo de
  cierre se llama **`cierre`**, no `criterio_cierre`) — y `proceso.spines.<tipo>`
  = lista de pasos `{paso, rol, plantilla, artefacto, cond?}`. Las plantillas
  viven en `semilla/plantillas/proceso/{historia,spike}/` y se materializan a
  `.arnesia/proceso/<tipo>/` al sembrar (`scaffolder.go:88-113`).
- **AUD-1 · La cadena paso→caja NO existe como dato hoy.** Ningún paso de spine
  referencia una caja del grafo (campo inexistente en YAML y en Go). Este corte
  la crea: **campo opcional `caja: <box-id>`** por paso (aditivo a `PasoSpine`,
  retro-compatible). Paso sin `caja` = «paso sin caja aún», VISIBLE en el foco
  (E13). La semilla y `developer-vitalia` lo declaran donde exista la caja.
- **AUD-2 · El seam no está cableado.** El daemon HOY no lee el `arnes.yaml` de
  ningún proyecto: el parser vive solo en el carril de forja (gap Fase 2
  declarado en `parser.go:6-7`); el Mapa se sirve del índice
  (`GET /api/harnesses/{id}/graph` = passthrough de `domain.Graph`,
  `router.go:87/221`). MA-T1 incluye cablearlo: el loader lee `arnes.yaml` del
  proyecto y **deriva AL INDEXAR** (patrón vigente de los edges,
  `loader/edges.go:27-52`) — no en request, no en el sello. Llevarla al sello
  `arnes.l0.json` = decisión futura, NO de este corte.
- **AUD-3 · Dónde viaja en el wire.** El schema L0 cierra la raíz y el bloque
  `arnes` (`additionalProperties:false` — `graph.l0.schema.json`); un
  `actividades[]` sin enmienda VIOLA contrato firmado. Este corte hace la
  **enmienda ADITIVA** al schema: `arnes.actividades[]` (catálogo:
  `{id, estados, cierre, pasos:[{paso, rol, artefacto, caja?}]}`) + faceta por
  caja `actividades: [ids]` (legal ya: `$defs.nodo` es `additionalProperties:
  true`). De paso se **regulariza el drift destapado**: `degradado` (S1-D27) está
  en el wire Go (`graph.go:159`) y NO en el schema — entra en la misma enmienda.
- **AUD-4 · «spine» ya está tomado en el wire** con OTRO significado:
  `arnes.spine` (singular) = máquina de estados del paquete de trabajo
  (`graph.go:71-77`, consumida por `VerificarSpine` y `transLabel`/`esTerminal`
  en FE). Queda INTACTO (compat mono-actividad). La secuencia por actividad se
  llama **`pasos`** en el wire y **procedimiento** en producto — jamás «spine».
  En código FE tampoco usar `act`: tomado 2 veces (`ActKind` activación de
  banda · `act` paso de turno del chat) — siempre `actividad` completo.
- **Vocabulario L0 verificado 2026-08-01:** «actividad»/«procedimiento» NO chocan
  con lo tomado (procedencia · origen · canal · insumos · banda; `RolAct` de
  session.go es otro plano — paso de actividad del TURNO de chat, no colisiona
  como campo de nodo).
- **Cajas no referenciadas por ningún procedimiento** = grupo **`sin-actividad`**,
  VISIBLE (honestidad; insumo de mantener §6, jamás se oculta).

## §2 · Leyes (MA-L1..L7)

1. **MA-L1 · El-todo-primero.** Vista default = panorama agregado. Con >1
   actividad JAMÁS se renderizan todos los procedimientos expandidos a la vez.
2. **MA-L2 · La base no se colapsa.** Las bandas transversales (Guardia · Base)
   se ven en TODO nivel de detalle — ahí viven las normas de proyecto
   (arquitectura, seguridad, calidad) que aplican a toda actividad.
3. **MA-L3 · Superset estricto.** Geografía firmada intacta: 7 bandas · fases ·
   glifos · facetas · capas (estructura/Mejora) · las **13** stories vigentes de
   `map-canvas.stories.tsx` (+5 de `map-bar.stories.tsx`; la cifra «14» de la v1
   era errónea — AUD-5) · inspector de **4 tabs** (Mejora existe desde T34/RF-258)
   · `.node-meta` de cajas (arquetipo·perfil·gate) · picker ELIMINADO (TS-D21:
   id read-only + peek). «Actividad» = lente/agrupación NUEVA, no reemplaza
   ningún vocabulario.
4. **MA-L4 · Compartido visible.** Caja usada por N actividades lo declara
   (badge ×N). Editarla avisa el radio de impacto («usada por N actividades»).
5. **MA-L5 · Honestidad.** Actividad sin procedimiento → «sin procedimiento aún»
   (degradado + CTA al chat). Caja sin actividad → `sin-actividad`. Arnés sin
   tipos declarados → Mapa exactamente como hoy (cero selector, cero invento).
6. **MA-L6 · El foco atenúa, no borra.** Focar una actividad baja la opacidad del
   resto — el todo sigue presente y clickeable. Nunca filtra-y-esconde.
7. **MA-L7 · La cadena se asoma, la Galaxia la muestra.** El Mapa de UN arnés
   muestra sus puertos de entrada/salida (gates DEF-D3); la cadena completa entre
   arneses = Galaxia (fuera de este corte). **En este MVP la ley la satisface la
   franja de artefactos VIGENTE** (chips `externo` = puerto de entrada · tag
   `salida del proceso` = puerto de salida — ya derivados de necesita/entrega,
   `artefactos.ts`): cero render nuevo; `proceso/<id>.yaml` aún no existe como
   dato (AUD-6).

## §3 · Superficie (niveles de detalle)

**N0 · Panorama** (default): la geografía vigente ENTERA + fila de **chips de
actividad** en `map-bar` (uno por tipo: **id del tipo** — no hay `label` en el
dato, AUD-5 — · nº cajas · salud agregada = worst-of). **Fuente de la salud
(honesta, sin dato nuevo):** worst-of de los hallazgos FE ya existentes de sus
cajas (`gate:none` crit · `no-reconocido` warn · checks rojos de
`/conformance`) — la misma fuente del inspector, jamás una cifra tecleada.
Hover chip → pre-resalta sus cajas. Sin tipos declarados → la fila no existe
(MA-L5). Chip final **`sin-actividad · N`** (muted) si el grupo no está vacío.

**N1 · Foco de actividad** (click chip): cajas de la actividad a opacidad plena +
**secuencia del procedimiento dibujada** (SVG caja→caja en orden de `pasos`,
numerados). **Estilo NUEVO sin colisión:** trazo sólido `--primary` con número de
paso — los dash ya significan otra cosa (`3 3` escribe · `4 4` lee · `2 6`
opcional). **Reusa la maquinaria de atenuación vigente** (`focus`/`related` set +
`.node.dim` opacity .4, `map-canvas.tsx:101-171` + `map.css:265`): solo cambia el
origen del set. Resto atenuado (MA-L6). Guardia/Base quedan plenas (MA-L2).
Fases donde la actividad no tiene pasos → se atenúan enteras: **se VE** que el
spike no llega a producción. Breadcrumb `arnés ▸ actividad` + «ver todo» en la
`map-bar` (junto al id read-only — el picker ya no existe, TS-D21). Badge ×N en
cajas compartidas, con las otras actividades listadas al tocar. Paso con
`caja` ausente → marcador «paso sin caja aún» EN la secuencia (E13).

**Inspector:** tab Resumen gana fila «Actividades: historia · bugfix» (click →
foco). **Capas:** foco × capa Mejora componen — MVP: el foco solo atenúa; las
cifras siguen siendo del todo, con disclaimer («cifras del arnés completo»);
filtrar costo por actividad = diferido (§7).

## §4 · Escenarios complejos (MA-E1..E12)

| # | Escenario | Resolución |
|---|---|---|
| E1 | Arnés mono-actividad | Selector ausente; Mapa = hoy. Cero regresión |
| E2 | 8+ actividades (caso CEO) | Chips con overflow → menú; agrupar por macroproceso = futuro |
| E3 | Caja compartida ×N | Badge + pertenencia visible; editar avisa radio de impacto (MA-L4) |
| E4 | Cambia una norma Base (p.ej. arquitectura) | Impacto = todas las actividades; visible porque Base nunca se atenúa (MA-L2) |
| E5 | Actividades con alcance de fases distinto | Fases vacías atenuadas en foco — el spike TERMINA ANTES y se ve |
| E6 | Caja `sin-actividad` | Grupo visible; candidata a declarar o podar (insumo §6) |
| E7 | Actividad recién forjada sin cajas | Chip degradado «sin procedimiento aún» + CTA chat (MA-L5) |
| E8 | Arnés legacy sin `arnes.yaml`/tipos | Mapa vigente tal cual; «Identificar actividades» = futuro (extiende «Identificar» de Slice 2) |
| E9 | Deriva en una instalación | Mecanismo vigente intacto; el foco AYUDA a localizarla por actividad |
| E10 | Operador con 2 roles | 2 arneses (v3); relación = cadena/Galaxia, no este Mapa |
| E11 | WIP en vuelo por actividad | Diferido (Mapa DESTINO item 4); el modelo lo soporta (paquete lleva tipo) |
| E12 | ¿Qué actividad quema más plata? | Diferido; el join D12.2 gana `actividad` como dimensión cuando la llave de terreno lleve el tipo |
| E13 | Paso de procedimiento sin `caja` declarada (AUD-1: la cadena paso→caja nace opcional) | Marcador «paso sin caja aún» EN la secuencia del foco — hueco visible, jamás se salta en silencio |

## §5 · Crear — forja + dogfood `developer-vitalia` (DEF-D6)

**Flujo de forja (esquema v3):** chat/`arnesia init` declara tipos en `arnes.yaml`
+ siembra `.arnesia/proceso/<tipo>/` (idempotente, ya construido carril A) → gate
de completitud D19 por actividad → sello → publicar (write-side B2, ya construido
carril B) → aparece en Portafolio/Mapa.

**Dogfood real (directiva 2026-08-01):** extraer de
`/home/chalreme/Proyectos/vitalia-app` (material CRUDO: 42 skills · 10 agents ·
10 commands · 1 workflow · 45 rules · 4 hooks, sin sello en raíz) el primer arnés
**`developer-vitalia`**:

1. **Sanear — estado OBSERVADO 2026-08-01 (AUD-7), no el asumido:** lo indexado
   como «vitalia» es `~/Proyectos/luana-vitalia/vitalia` y **YA está `sin-home`**
   (`portafolio.json`: identidad sin home, `origen: {}`; índice:
   `sin-home~vitalia~vitalia`) — la ficción de marketplace en el store no existe.
   Lo que SÍ hay que sanear: **residuo del re-key** (clave legacy `vitalia`
   duplicada en `index.db` y `arneses.json` junto a `sin-home~vitalia~vitalia`).
   Y `~/Proyectos/vitalia-app` **jamás fue escaneado**: MA-T2 = limpiar el
   duplicado legacy + agregar vitalia-app como material crudo `sin-home`
   (mecanismo existente, S1-D27..29). Nada se inventa.
2. **Aplicar el criterio de corte (primera vez en real):** el `.claude/` de
   vitalia-app mezcla VARIOS puestos (inventario verificado: 42 skills · 10
   agents · 10 commands · 45 rules · 4 hooks · 1 workflow). `developer-vitalia`
   toma SOLO el puesto developer —
   **DENTRO:** dev-team · frontend-expert · backend-expert · clerk-* (12) ·
   playwright-expert · chrome-devtools-verify · tessl-context · copilot-expert ·
   commit-push · hipaa-check · vitalia-design-system (knowledge) + agents
   builder-*/gate-runner/context-* + rules de ruta (debugging · tdd-mandatory ·
   e2e-testing · hotfix-repro-mandatory · story-closure-gate…).
   **FUERA (otro operador → separá SIEMPRE):** pm · pm-vitalia · po · po-ux ·
   ux-agentico · sales-agent-expert · brand-* · content-hunter ·
   data-storyteller · metrics/offer-* · **pase-produccion** (ejemplo 1 del
   operador: el pase lo hace OTRA persona) + agents auditor-* si el veredicto es
   de otro puesto.
   **AMBIGUOS — decisión del operador EN el gate de MA-T3** (no se resuelven
   solos): architect · auditor · manychat-expert · handoff · harnesses-improvement
   · harness-issue. El entregable de MA-T3 incluye la **tabla de clasificación
   completa 42+10+10+45** con su lado del corte y el porqué.
3. **Clasificar por las 3 caras (D18):** rules transversales (backend-ddd ·
   frontend-fsd · architectural-fitness · tenant-isolation · hipaa-lite ·
   pii-sanitisation · git-safety…) → **base común** (banda Base). Rules de ruta
   (tdd-mandatory · hotfix-repro-mandatory · story-closure-gate…) → procedimiento
   de su actividad.
4. **Declarar las 4 actividades del ejemplo del operador** en su `arnes.yaml`:
   `historia` (dado un SPEC) · `bugfix` · `spike` · `revisar-capability`, cada una
   con `estados` + `cierre` + `pasos` (vocabulario real, AUD-5) y `caja:` donde la
   caja exista (AUD-1). La semilla trae historia+spike; bugfix+revisar-capability
   se agregan — misma deuda ya anotada en BACKLOG.
5. **Sellar + publicar** al marketplace propio como plugin **`developer-vitalia`**
   (nombre canónico manifiesto→plugin.json→id, HS-12) → verificar en Portafolio
   (canónico) y en el Mapa multi-actividad (§3).

## §6 · Mantener

- **Edición conversacional** (vivo: tarjeta de identidad + reindex-tras-turno):
  el refetch del Mapa refresca las facetas `actividades` solo.
- **Radio de impacto (MA-L4):** tocar caja compartida o norma Base → el chat
  antepone «usada por N actividades» / «norma transversal: afecta todas».
- **Deriva por instalación** intacta; **versionado:** cambio en cualquier
  actividad = semver del ARNÉS (1 arnés = 1 plugin, un solo ciclo de release).
- **Insumos de poda:** `sin-actividad` (E6) + telemetría por actividad (E12,
  diferido) alimentan la mejora continua.

## §7 · Corte MVP (integrado al MVP 2026-07-30 como carril D) — tickets

| # | Qué | Dónde |
|---|---|---|
| MA-T1a | **Dato:** campo opcional `caja:` en `PasoSpine` (parser + semilla lo declara donde aplique) — AUD-1 | Go: forja/parser + `semilla/arnes.yaml` |
| MA-T1b | **Seam:** loader lee `arnes.yaml` del proyecto (gap Fase 2, `parser.go:6-7`) + deriva `arnes.actividades[]` + faceta por caja **AL INDEXAR** (patrón edges) + **enmienda ADITIVA** a `graph.l0.schema.json` (actividades + regularizar `degradado`) + espejo TS en `entities/arnes` — AUD-2/AUD-3 | Go: loader/índice + schema + FE types |
| MA-T2 | Saneo vitalia REAL (AUD-7): limpiar clave legacy `vitalia` duplicada (residuo re-key) + agregar `vitalia-app` como material crudo `sin-home` | Go/estado Portafolio |
| MA-T3 | Forja `developer-vitalia` (§5.2-5.5): tabla de clasificación completa (ambiguos = decisión del operador en el gate) + 3 caras + 4 actividades + sello + publicar B2 | vitalia-app + marketplace |
| MA-T4 | `map-bar`: chips de actividad (N0, salud = worst-of hallazgos FE) + chip `sin-actividad` | FE |
| MA-T5 | Foco N1: reusar `related`/`.dim` + secuencia sólida `--primary` numerada + badge ×N + fila «Actividades» en tab Resumen (entre Clasificación y Fuente) + E13 | FE |
| MA-T6 | **Mockup superset** — fork del baseline **RE-DERIVADO a la superficie vigente** (el `.html` baseline quedó stale: sin picker TS-D21 · 4 capas con Mejora · inspector 4 tabs · node-meta · franja artefactos completa · tokens PRENTER · `<meta charset>`), fila nueva en `mockups/INDEX.md` — **GATE 🧑‍⚖️ ANTES de MA-T4/T5** | mockups/ + paquete |
| MA-T7 | Estados degradados E1/E7/E8/E13 + stories nuevas con gate a11y (superset de las 13) | FE/Storybook |

**Orden:** T1a→T1b→T2→T3 (backend+dogfood, sin UI) ∥ T6 (mockup+gate) → T4→T5→T7.
**Diferidos explícitos (no silenciosos):** WIP por actividad (E11) · costo por
actividad (E12) · «Identificar actividades» (E8) · render de cadena/Galaxia
(MA-L7 solo puertos) · agrupación por macroproceso (E2) · faceta al sello.

## §8 · Verificación / PARIDAD

- Las **13** stories vigentes del canvas (+5 de map-bar) siguen verdes SIN tocar
  (MA-L3 medible).
- La enmienda del schema valida: grafo vigente (sin actividades) Y grafo nuevo
  (con `arnes.actividades[]` + `degradado`) pasan ambos contra
  `graph.l0.schema.json`; `VerificarSpine` intacto sobre `arnes.spine` (AUD-4).
- Stories nuevas: N0 chips · N1 foco (secuencia + atenuación + badge) ·
  E1/E5/E7/E13 degradados — gate a11y entero.
- Dogfood doble: el kit propio (historia/spike de la semilla) + `developer-vitalia`
  real servido por el daemon.
- E2E: `developer-vitalia` publicado → Portafolio lo lista canónico → Mapa lo
  abre → foco `bugfix` muestra su procedimiento y el spike termina antes (E5).
- PARIDAD.md del paquete con evidencia por ticket; gate 🧑‍⚖️ del operador.
