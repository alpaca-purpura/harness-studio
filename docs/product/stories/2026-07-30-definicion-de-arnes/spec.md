# Spec — crear · visualizar · mantener arneses multi-actividad (esquema v3) + dogfood `developer-vitalia`

> vig: escrita 2026-08-01, **esperando firma 🧑‍⚖️** · deriva de DEF-D1..D3 FIRMADAS
> (`decisiones.md`) + directivas DEF-D5/D6 del operador (round 3, `chris-input.md`).
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

## §1 · Modelo de datos (el seam)

- **Fuente de actividades:** `arnes.yaml` → `gestion_trabajo.tipos_paquete`
  (id · label · estados · criterio_cierre) + `proceso.spines[tipo]` (pasos → qué
  cajas ejecuta cada paso). La semilla ya lo materializa (`semilla/arnes.yaml`
  L146/L170; `.arnesia/proceso/{historia,spike}`).
- **Derivación (runtime, NO sello):** el daemon deriva y sirve en el payload del
  Mapa: `actividades[]` (catálogo) + por caja la faceta **`actividades: [ids]`**
  (qué spines la referencian). Derivada como `ctxPct`: se calcula, no se estampa;
  llevarla al sello `arnes.l0.json` = decisión futura, NO de este corte.
- **Vocabulario L0 verificado 2026-08-01:** «actividad»/«procedimiento» NO chocan
  con lo tomado (procedencia · origen · canal · insumos · banda; `RolAct` de
  session.go es otro plano — paso de actividad del TURNO de chat, no colisiona
  como campo de nodo).
- **Cajas no referenciadas por ningún spine** = grupo **`sin-actividad`**,
  VISIBLE (honestidad; insumo de mantener §6, jamás se oculta).

## §2 · Leyes (MA-L1..L7)

1. **MA-L1 · El-todo-primero.** Vista default = panorama agregado. Con >1
   actividad JAMÁS se renderizan todos los procedimientos expandidos a la vez.
2. **MA-L2 · La base no se colapsa.** Las bandas transversales (Guardia · Base)
   se ven en TODO nivel de detalle — ahí viven las normas de proyecto
   (arquitectura, seguridad, calidad) que aplican a toda actividad.
3. **MA-L3 · Superset estricto.** Geografía firmada intacta: 7 bandas · fases ·
   glifos · facetas · capas (estructura/Mejora) · las 14 stories vigentes de
   `map-canvas.stories.tsx`. «Actividad» = lente/agrupación NUEVA, no reemplaza
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
   arneses = Galaxia (fuera de este corte).

## §3 · Superficie (niveles de detalle)

**N0 · Panorama** (default): la geografía vigente ENTERA + fila de **chips de
actividad** en `map-bar` (uno por tipo: label · nº cajas · salud agregada =
worst-of). Hover chip → pre-resalta sus cajas. Sin tipos declarados → la fila no
existe (MA-L5).

**N1 · Foco de actividad** (click chip): cajas de la actividad a opacidad plena +
**secuencia del procedimiento dibujada** (SVG punteado caja→caja en orden del
spine, pasos numerados). Resto atenuado (MA-L6). Guardia/Base quedan plenas
(MA-L2). Fases donde la actividad no tiene pasos → se atenúan enteras: **se VE**
que el spike no llega a producción. Breadcrumb `arnés ▸ actividad` + «ver todo».
Badge ×N en cajas compartidas, con las otras actividades listadas al tocar.

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

## §5 · Crear — forja + dogfood `developer-vitalia` (DEF-D6)

**Flujo de forja (esquema v3):** chat/`arnesia init` declara tipos en `arnes.yaml`
+ siembra `.arnesia/proceso/<tipo>/` (idempotente, ya construido carril A) → gate
de completitud D19 por actividad → sello → publicar (write-side B2, ya construido
carril B) → aparece en Portafolio/Mapa.

**Dogfood real (directiva 2026-08-01):** extraer de
`/home/chalreme/Proyectos/vitalia-app` (material CRUDO: 42 skills · 10 agents ·
10 commands · 1 workflow · 45 rules · 4 hooks, sin sello en raíz) el primer arnés
**`developer-vitalia`**:

1. **Sanear la ficción de marketplace:** vitalia-app NO tiene home. Donde el
   Portafolio/índice hoy la muestre como-si-de-marketplace, queda como material
   crudo `sin-home` (mecanismo existente, fix S1-D27..29). Nada se inventa.
2. **Aplicar el criterio de corte (primera vez en real):** el `.claude/` de
   vitalia mezcla VARIOS puestos. `developer-vitalia` toma SOLO el puesto
   developer: dev-team · frontend-expert · backend-expert · clerk-* (12) ·
   playwright-expert · tessl-context · debugging/tdd/e2e… Los skills de OTROS
   operadores (pm, pm-vitalia, po, po-ux, sales-agent-expert, brand-*,
   content-hunter, data-storyteller, metrics/offer-*) quedan FUERA — material de
   futuros arneses hermanos (criterio #1: otro operador → separá SIEMPRE).
3. **Clasificar por las 3 caras (D18):** rules transversales (backend-ddd ·
   frontend-fsd · architectural-fitness · tenant-isolation · hipaa-lite ·
   pii-sanitisation · git-safety…) → **base común** (banda Base). Rules de ruta
   (tdd-mandatory · hotfix-repro-mandatory · story-closure-gate…) → procedimiento
   de su actividad.
4. **Declarar las 4 actividades del ejemplo del operador** en su `arnes.yaml`:
   `historia` (dado un SPEC) · `bugfix` · `spike` · `revisar-capability`, cada una
   con estados + criterio_cierre + spine (la semilla trae historia+spike;
   bugfix+revisar-capability se agregan — misma deuda ya anotada en BACKLOG).
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
| MA-T1 | Seam de datos: derivar `actividades[]` + faceta por caja desde `arnes.yaml` (tipos+spines) y servirlo en el payload del Mapa | Go: loader/daemon |
| MA-T2 | Saneo vitalia: material crudo `sin-home`, sin marketplace ficticio | Go/estado Portafolio |
| MA-T3 | Forja `developer-vitalia` (§5.2-5.5): corte por puesto + 3 caras + 4 actividades + sello + publicar B2 | vitalia-app + marketplace |
| MA-T4 | `map-bar`: chips de actividad (N0) | FE |
| MA-T5 | Foco N1: atenuación + secuencia SVG + badge ×N + fila en inspector | FE |
| MA-T6 | **Mockup superset** (forkear `arnesia-mapa-baseline.html`, fila nueva en `mockups/INDEX.md`) — **GATE 🧑‍⚖️ ANTES de MA-T4/T5** | mockups/ |
| MA-T7 | Estados degradados E1/E7/E8 + stories nuevas con gate a11y (superset de las 14) | FE/Storybook |

**Orden:** T1→T2→T3 (backend+dogfood, sin UI) ∥ T6 (mockup+gate) → T4→T5→T7.
**Diferidos explícitos (no silenciosos):** WIP por actividad (E11) · costo por
actividad (E12) · «Identificar actividades» (E8) · render de cadena/Galaxia
(MA-L7 solo puertos) · agrupación por macroproceso (E2) · faceta al sello.

## §8 · Verificación / PARIDAD

- Las 14 stories vigentes del canvas siguen verdes SIN tocar (MA-L3 medible).
- Stories nuevas: N0 chips · N1 foco (secuencia + atenuación + badge) · E1/E5/E7
  degradados — gate a11y entero.
- Dogfood doble: el kit propio (historia/spike de la semilla) + `developer-vitalia`
  real servido por el daemon.
- E2E: `developer-vitalia` publicado → Portafolio lo lista canónico → Mapa lo
  abre → foco `bugfix` muestra su procedimiento y el spike termina antes (E5).
- PARIDAD.md del paquete con evidencia por ticket; gate 🧑‍⚖️ del operador.
