# Decisiones — Spike carga de arneses/proyectos (2026-07-10)

> `tipo: spike`. Cada decisión conversada se escribe EN EL MISMO TURNO (§10). Cierre = decisión documentada.
> Nada firmado aún.

## S-D0 · Alcance del spike — DECIDIDA
- Investigar + **decidir el diseño** del mecanismo de carga (cara A arnés-único + cara B proyecto-multi).
- Entregar un **thin proof** de la cara A (dialog FE → PUT existente) para dar agencia inmediata.
- La cara B (detección proyecto-con-arneses) se **decide en papel**; su build es slice posterior.

## S-D1 · Estado hallado (no reconstruir) — HALLAZGO
- Backend de carga single-arnés **HECHO** (PUT valida+registra+indexa, honesto; loader CAP-15-20; FE client).
- Gaps: (1) **dialog FE** (diferido HS-07) · (2) **detector proyecto-multi** lock `.devstudio/arneses.yaml` (diferido HS-12).
- ⇒ el spike NO toca el motor de carga; ataca la afordancia FE + decide la detección de proyecto.

## S-D2 · Path input vs folder-picker nativo — ABIERTA (decidir en el proof)
- Browser (dev) no tiene folder-picker seguro; el shell de escritorio (Tauri) sí (dialog nativo).
- Hipótesis: el thin proof usa **input de path de texto** (funciona en browser Y desktop) + más tarde el
  picker nativo de Tauri como azúcar. Confirmar al construir el dialog.

## S-D3 · Cómo se deriva el `id` del arnés al cargar — ABIERTA
- El PUT exige `{id}` en la ruta. ¿El usuario lo teclea, o se **deriva** del manifiesto/nombre de carpeta
  (nomenclatura v1: manifiesto→plugin.json→id, HS-12)? Preferencia: derivar + permitir override. Decidir en el proof.

## S-D4 · Detección proyecto-con-arneses (cara B) — ABIERTA (decisión de papel)
- Un proyecto real (Vitalia) = repo con arnés(es) instalados en `.claude/` (+ posible lock `.devstudio/arneses.yaml`).
- Cargar el proyecto = detectar TODOS los arneses adentro + registrar cada uno. Requiere el detector 3° (HS-12).
- Salida esperada del spike: decidir si (i) se reusa el lock DevStudio, (ii) se camina `.claude/plugins`, o (iii) ambos;
  y si «cargar proyecto» es un endpoint nuevo o el PUT actual iterado. → documenta_decision al cerrar.

## S-D5 · Reencuadre: «Portafolio» = front-door del CICLO DE VIDA del arnés — EN DISCUSIÓN (no firmada)
> El user amplió el alcance (2026-07-13): «cargar carpeta» era la punta. Lo que quiere es una **vista
> Portafolio** desde donde se **agrega** y **gestiona el ciclo de vida completo** (VISION crear·mapear·observar·mejorar).

- **Dos tipos de «agregar»:**
  - (a) **Arnés de marketplace** → gestionar acceso GitHub → «descargar» (clonar) el origen a un lugar de la
    máquina → editar/mejorar → **actualizar y publicar** siguiendo un ciclo de vida óptimo para Claude.
  - (b) **Proyecto (carpeta) con arnés YA instalado y ejecutado** → al cargarlo, **detectar de inmediato**
    qué arnés es y **de qué marketplace vino** (resolver origen).
- **Por qué cargar un proyecto (valor):**
  1. **Revisar** qué cambios hizo / cómo está → mejorar el arnés ORIGEN según lo que veamos que funciona (backport).
  2. **Reparar** (reinstalar plugin + recablear). Lo puede hacer el user; si lo hacemos nosotros, detectamos problemas.
- **Distribución multi-máquina:** un arnés vive instalado en muchos lugares. Al publicar una versión, los users en
  sus máquinas deben **ver que hay versión nueva → actualizar → saber qué hay de nuevo** (changelog).

### Modelo de entidades (propuesto, para firmar)
- **Arnés canónico (origen):** repo fuente autoral en un marketplace git (GitHub). Identidad = `id` (nomenclatura v1).
  Tiene versiones semver + changelog.
- **Marketplace:** repo GitHub (formato `prenter-marketplace`, HS-12) que indexa arneses + versiones. Fuente de
  distribución y de chequeo-de-actualización.
- **Checkout local:** el origen clonado a disco para editar/publicar (③ editable).
- **Instalación:** arnés materializado en el `.claude/` de un proyecto. Tiene `installedVersion` + ref-de-origen
  (procedencia). Puede driftar del origen.
- **Proyecto:** carpeta/repo con ≥1 instalación.
- **Portafolio:** el registro home de la app de todo lo anterior que el user agregó.
- **Enlace keystone:** `instalación --resolver--> origen`. Es lo que desbloquea revisar/backport/reparar/actualizar.

### Mapa de huecos (anclado en código, 2026-07-13)
- REUSA: loader (detecta forma+manifiesto, 1 arnés/dir) · registry+PUT/list · selfupdate (plantilla identidad+update
  git+apply atómico) · publisher (seam del push, STUB) · conformance `--arnes` (diagnóstico = base de «reparar»).
- NET-NEW: (1) **procedencia** — `plugin.json` NO tiene campo de origen; nada liga ③→canónico [keystone].
  (2) **detección multi-arnés/proyecto** (loader es 1 dir=1 arnés). (3) **acceso GitHub + clone/pull**.
  (4) **versionado+changelog+publish real** (publisher es stub). (5) **update-check + notify** (ver versión nueva).
  (6) **reparar/recablear** (re-materializar plugin en el `.claude/` de un proyecto arbitrario).

### Forks — RESUELTOS 2026-07-13 (con el user)
- **F5 · Encuadre → REENCUADRAR spike→PROGRAMA.** El spike decide modelo de entidades + F1 + elige 1er slice;
  el resto = épica(s) en BACKLOG. No disfrazar programa de spike.
- **F1 · Procedencia → LOCK SIDECAR `.devstudio/arneses.yaml`** (detector 3° HS-12). No forkear `plugin.json` (schema CC ajeno).
- **F2 · Acceso GitHub → AMBAS: `gh` si está autenticado, PAT (user-owned, local `~/.arnesia`) como fallback.**
  App = conductor (shell-out), nunca proxya login hosteado — respeta auth-terms firmado.
- **F3 · Update → HÍBRIDO:** capa propia detecta+notifica (semver instalado vs último tag del marketplace + changelog);
  re-materialización delegada a plugin-native de CC cuando aplique.
- **F4 · Gate de publish (no preguntado, propuesto):** editar → conformance-verde → bump semver → changelog OBLIGATORIO
  → tag + `git push` → índice del marketplace. El changelog ES lo que el user remoto lee. (confirmar al llegar a S5.)

### Programa «Portafolio · ciclo de vida del arnés» — slices (para BACKLOG)
- **Slice 1 (keystone, 1° · thin) — Portafolio + agregar PROYECTO → detectar + resolver origen.** Vista home Portafolio;
  «agregar carpeta» → walker (S2) enumera `.claude/plugins/*` + settings + lock → resuelve origen (S1, lock) →
  muestra `{id, versión, origen|desconocido, drift}`. Read-side end-to-end. Reusa loader+registry; net-new = walker + lector de lock.
  ⮑ investiga: ¿dónde registra CC ya la procedencia de un plugin instalado (settings/plugins config)? — fallback honesto a «desconocido».
- **Slice 2 — agregar ARNÉS de marketplace → clonar (S3).** `gh`/PAT → clone a `~/.arnesia/checkouts/<id>/` → aparece en
  Portafolio → chat/mejorar (S4 ya existe).
- **Slice 3 — publicar (S5).** publisher real: gate conformance + semver + changelog + push al marketplace.
- **Slice 4 — update-check + notify (S6).** badge «hay vX» + changelog + actualizar (reusa patrón selfupdate).
- **Slice 5 — reparar/recablear (S7).** re-materializar plugin en `.claude/` + `conformance --arnes`.

## S-D6 · Facetas `empresa` + `marketplace` (ortogonales) + asociación de negocio — DECIDIDA 2026-07-13
> Acotación del user: los arneses NO son ajenos entre sí a nivel de negocio (ej. RRHH: reclutamiento/selección ·
> clima/cultura · legal-administrativo — asociados por proceso/personas). Técnicamente = acoplamiento mínimo.
> **No cambia el plan — lo VALIDA:** el modelo firmado YA lo anticipa.

- **YA en el modelo (reusar, no reinventar):** `arnes.l0.json` canónico carga `empresa` + `marketplace` + `reporta_a`
  (verificado en `dogfood/dev-full-cycle/arnes.l0.json`); `domain.Arnes` tiene `Empresa`/`Marketplace` (graph.go:36,39);
  META de enganche = `rol·proceso·reporta-a·empresa` (graph.go:24, VISION); sesiones llevan `empresa`/`puesto` (Galaxia HS-03).
- **REGLA (Portafolio):** el Portafolio **agrupa por `empresa`**. `empresa` y `marketplace` son **ORTOGONALES** —
  un marketplace puede contener varias empresas (arranque: alpacapurpura hostea todos los plugins porque la empresa no
  maneja GitHub; futuro: quizá 1 marketplace por empresa). Cada arnés declara AMBOS.
- **Procedencia (S1) refinada:** origen se resuelve de `manifiesto.marketplace` + `manifiesto.empresa` (declarados) +
  lock `.devstudio/arneses.yaml` (ref/tag/`installedVersion` de install-time). Cadena honesta: manifiesto → lock → «desconocido».
- **Asociación de negocio (`reporta_a` / procesos que nutren a otros):** hook YA existe (`reporta_a` nullable). Se
  materializa a futuro en Mapa/**Galaxia** (relación entre arneses). **FUERA de Slice 1** — noted, no se construye ahora.
- **Deuda menor detectada:** `internal/adapters/index/store.go:102` hardcodea `Empresa: "alpacapurpura"` (seed/fallback) —
  cuando el walker (Slice 1) lea empresa real del manifiesto, revisar que no pise el valor declarado.

## S-D7 · Correcciones UX + lógica de negocio (feedback mockup v1) — 2026-07-13
> El mockup v1 (`arnesia-portafolio.html`) puso TODO en una pantalla → no entendible. El user pide pensar el
> flujo de experiencia + de uso, con la lógica de negocio bien clara. Correcciones CONFIRMADAS:

- **C1 · Relación arnés–empresa–marketplace = MUCHOS A MUCHOS** (corrige el «⊥ ortogonal, 1 c/u» de S-D6).
  Una misma empresa puede tener arneses en marketplaces distintos; un arnés (base/compartido) puede servir a
  varias empresas y estar publicado en varios marketplaces. ⇒ el Portafolio NO es un árbol empresa→arnés; empresa
  y marketplace son **lentes/facetas**, no dueños exclusivos. Un arnés puede aparecer bajo 2 empresas.
- **C2 · Multi-superficie, no una pantalla.** IA propuesta: (1) **Lista** (browse, limpia, fila compacta) ·
  (2) **Agregar** (wizard con pasos) · (3) **Detalle** (drawer con toda la info + acciones de ciclo de vida).
  Las 5+ acciones NO van en la card de lista — viven en el detalle.
- **C3 · Flujos de adquisición (git-aware, con validación):**
  - **Proyecto:** fuente = **carpeta local** O **repo GitHub** (clonar). Luego escanear → detectar arneses
    instalados → resolver empresa/marketplace → elegir cuáles agregar.
  - **Marketplace:** apuntar con **ruta git** → la app **valida** que tenga estructura de marketplace real →
    si válido, **lista los plugins** que ofrece → el user **elige** cuáles agregar (clona checkout para trabajar).
    Si inválido → error honesto.
- **C4 · Acción `desvincular` por arnés:** quitarlo del portafolio (deja de verse). Semántica a confirmar
  (solo saca de la vista; NO desinstala del proyecto ni borra el clon — Q3).
- **C5 · Propósito del «Agrupar»** no era claro. Reencuadre propuesto: **filtros** (acotar el set) +
  **lente «Ver por»** (empresa | proyecto | marketplace | plano) para ENCONTRAR un arnés cuando hay muchos.
  A confirmar si se quiere en v1 (Q2).

### Preguntas abiertas al user (espacio para opinión escrita) — Q1-Q4 en el mensaje. Al responder → S-D8.
### Re-corte Slice 1 (tras C1-C5): Lista(por empresa) + Agregar-Proyecto(carpeta local 1°)→escanear→detectar+resolver
  + Detalle-drawer(read) + Desvincular. Marketplace-flow + GitHub-clone-proyecto = Slice 2. Mockup se rehace tras Q1-Q4.

## S-D8 · Modelo CERRADO (Q1-Q4 respondidas) — 2026-07-13
> Respuestas del user a Q1-Q4. El modelo de entidades queda cerrado para firmar; de acá sale el mockup v2.

- **Q1 · Unidad = el ARNÉS por identidad → 1 card** (dedup, no duplica con N:M:M). Adentro, DOS presencias distintas:
  - **Canónico (checkout)** = la **ÚNICA copia editable** — único lugar de autorear/mejorar/publicar. 0 o 1 por arnés.
  - **Instalaciones[]** = **espejos de OBSERVACIÓN, read-only** (en proyectos de cliente). N por arnés. Se diagnostican, no se editan.
- **LEY ANTI-DRIFT (principio de dominio, cementar):** todo authoring **converge en el canónico**. Una instalación se
  **observa** (diagnosticar cómo funciona en real) y se **re-materializa** (reparar = push canónico→instalación); **nunca se
  edita en sitio**. **Backport** = traer el *aprendizaje* de una instalación AL canónico → bump versión → publicar → reinstalar.
  ⇒ jamás dos copias editables de la misma versión divergen. Si NO hay canónico (solo se observa una instalación de cliente),
  CTA **«traer canónico del marketplace»** antes de habilitar autorear.
- **Q2 · Lente + filtros** aprobado (empresa | proyecto | marketplace | plano, como ayuda de navegación).
- **Q3 · Desvincular** = quitar del portafolio (no desinstala, no borra clon) + **opción secundaria «…y borrar clon local»**.
- **Q4 · Proyecto = SIEMPRE procedencia** de una instalación (no entidad de 1er nivel). La lente «por proyecto» es un filtro,
  la unidad sigue siendo el arnés.

### IA final (3 superficies)
1. **Lista** — 1 fila compacta por arnés; presencia = `◆ canónico vX` + `▣ N instalaciones` + flags (⬆ update · ◐ drift). Sin botones de acción.
2. **Agregar (wizard)** — Proyecto (carpeta local | repo GitHub → escanear → detectar → elegir) · Marketplace (git url → validar → listar → elegir → clonar).
3. **Detalle (drawer)** — zona **Canónico** (editable: Abrir Mapa · Mejorar · Publicar | o CTA traer-canónico) + zona **Instalaciones**
   (read-only: Observar/Diagnosticar · Reparar · Backport) + facets empresas/marketplaces + Desvincular. Ley anti-drift visible.

### Re-corte Slice 1 (final): Lista(por empresa) + Agregar-Proyecto(carpeta local) → escanear → detectar/resolver + Detalle-drawer
  (zonas canónico/instalaciones, READ) + Observar(Abrir en Mapa) + Desvincular. Editar/Reparar/Publicar/Backport/Marketplace-flow = slices sig.
- **Próximo:** mockup v2 (3 superficies, navegable) → si aprueba → build Slice 1 (capability + PARIDAD + firma). El spike CIERRA al firmar este modelo.

## S-D9 · Revisión adversaria (4 subagentes) — hallazgos + resolución — 2026-07-13
> Consolidado completo → `revision-adversaria.md`. Los 4 convergen: (1) faltan cimientos de dominio antes del FE;
> (2) el «drift» está indefinido y el mockup lo fabrica; (3) identidad `(registry,id)` se contradice sola; (4) la
> ley anti-drift era falsa por construcción (las instalaciones se editan fuera de ArnesIA). El «Slice 1 = reuso» es
> engañoso: ~60% es código nuevo/reestructuración (anclado en `loader.go`, `arnes_registry.go`, `graph.go`).

**ADOPTO (✅, aplico a specs):**
- Store de portafolio **separado** de `arneses.json` (que es cwd de sesión), y **degrada honesto** (no brickea boot).
- **Walker** `.claude/plugins/<id>/` + **fallback `plugin.json`** (id=name/version) — hoy un plugin sin `arnes.l0.json` da `Arnes=nil`.
- Campo **`version`** en dominio; `empresas[]`/`marketplaces[]` (N:M representable).
- **Drift = hash de contenido vs `marketplace/plugins/<id>/<versión>/` (inmutable)**; sin referencia → `deriva-no-evaluable`; nunca semver-string ni `git status`.
- **Ley anti-drift reformulada DESCRIPTIVA de ArnesIA** + marketplace=SSoT + pull-antes-de-push + tag-tras-push + reparar solo dir privado (superficies compartidas=merge) + bloqueo si lock DevStudio + fs-lock + checkout≠instalación (paths disjuntos).
- **Cadena de origen collect-all + reconcilia** (no «para en primer hit»); para *procedencia de la copia* **lock > manifiesto**; **canonicalizar** registry/remotes; clave provisional con scope.
- Fixes de honestidad del mockup (G1-G9): defaults `deriva-no-evaluable`/update-no-verificado, rama Marketplace disabled, estados faltantes, a11y.

**FORKS ABIERTOS (🔀, decide el user) — F-A/F-B/F-C en el mensaje:**
- **F-A · Identidad:** propongo `(home, id)` (home = `arnes.l0.marketplace` autor-declarado); registry-de-adquisición = faceta N:M. Ratificar.
- **F-B · Recut:** ¿**Slice 0 cimientos (dominio) → Slice 1 FE** o encoger Slice 1 distinto? (el FE honesto no es construible sin cimientos).
- **F-C · Nombre del drift:** candidato `deriva` (`en-deriva`/`al-hilo`/`deriva-no-evaluable`). Ratificar.
- Menores a decidir: ubicación del canónico (fuera de `~/.arnesia` vs excepción a `checkProtected`, E3); desambiguar «Portafolio» vs CAP-58 (H1).

**Estado:** el modelo NO se firma hasta cerrar F-A/F-B/F-C + aplicar los ✅ a las specs. NO se codea nada aún.

## S-D10 · Forks cerrados (F-A/F-B/F-C) — 2026-07-13
- **F-B → RECUT: Slice 0 «Cimientos del Portafolio» (dominio/backend) PRIMERO, luego Slice 1 «FE Portafolio».** El FE
  honesto no es construible sin cimientos (S-D9). Slice 0 = capabilities + tests sin FE; Slice 1 = FE sobre cimientos sanos.
- **F-A → Identidad = `(home, id)`**, donde `home` = `arnes.l0.marketplace` (marketplace canónico **autor-declarado**, «dónde
  vivo yo»). Mismo `home+id` = mismo arnés aunque venga de mirrors distintos; `home` distinto = arnés distinto. El
  **registry-de-adquisición** (de dónde salió ESTA copia) = **faceta N:M** aparte, resuelta por la cadena de origen.
- **F-C → «deriva»** (reemplaza «drift»). Estados: `en-deriva` · `al-hilo` · `deriva-no-evaluable`. Referencia =
  hash de contenido vs `home/plugins/<id>/<versión>/` inmutable (S-D9 D1).
- Los ✅ de S-D9 se aplican a `spec-funcional`/`casuistica`/`spec-usabilidad`; BACKLOG re-cortado (Slice 0 + Slice 1);
  mockup: los fixes de honestidad (G1-G9) se aplican en la etapa **FE (Slice 1)**, no ahora (cimientos-first).
- **Siguiente:** aplicar a specs → firmar modelo (cierra spike, → LEDGER) → build **Slice 0**.

## S-D11 · FIRMA del modelo 🧑‍⚖️ — 2026-07-13 (cierra el spike)
> Orden del operador: «ok, firmalo». El **modelo de entidades del Portafolio queda FIRMADO** tras: mockup v2 aprobado →
> specs escritas → **revisión adversaria de 4 subagentes** (integrada) → forks cerrados (S-D10). El spike se CIERRA
> como decisión documentada (→ `ledger/HS-22.md`). El programa continúa en `BACKLOG.md` (Slice 0 → 1 → …).

- **Firmado:** identidad `(home, id)` · N:M:M (facetas empresa/registry) · ley anti-drift **descriptiva** (marketplace=SSoT,
  pull-antes-de-push, reparar solo dir privado, checkout≠instalación, fs-lock) · **`deriva`** = hash vs `home/plugins/<id>/<v>/` ·
  cadena de origen **collect-all** (lock>manifiesto) · 7 deltas de dominio (D-DOM-1..7) · recut **Slice 0 cimientos → Slice 1 FE**.
- **NO firmado / difiere al build:** eslabón CC de procedencia (F5, investigar ANTES de firmar el orden fino) · ubicación del
  canónico (E3) · tabla de alias de rename (F4, slice update) · fixes de honestidad del mockup G1-G9 (etapa FE).
- **Hand-off:** el operador pasará el modelo por **Fable 5 (arquitecto/líder técnico)** para dividir tickets de trabajo
  (a construir con Sonnet 5). Superficie de descomposición lista: `spec-funcional.md` §9 (Slice 0) + §11 (deltas D-DOM-1..7) + casuística §I.
- **NO se codea desde ArnesIA-Opus:** el build lo hará el equipo de tickets. Este turno cierra en: modelo firmado + spike cerrado + commit a main.

## Abiertas (siguiente)
- Cerrar F-A/F-B/F-C → aplicar ✅ a `spec-funcional`/`casuistica`/`spec-usabilidad` + corregir mockup → firmar modelo (cierra spike, → LEDGER) → build.
- Registrar la épica + slices en `docs/product/BACKLOG.md`; el spike se CIERRA al firmar el modelo de entidades (→ LEDGER).
- **Slice 1** elegido como 1° → siguiente etapa = **mockup Portafolio** (story-first, superset del baseline anti-drift; leer `mockups/INDEX.md`).
- Confirmar F4 (gate publish) al llegar a Slice 3.
