# Revisión adversaria — consolidado (4 subagentes) + resolución

> `tipo: revision` · 2026-07-13. 4 subagentes adversarios (adquisición/git · ley anti-drift · identidad/N:M:M ·
> coherencia/UX) atacaron `spec-funcional.md` + `casuistica.md` + `spec-usabilidad.md` + el mockup, anclando contra
> código real. Acá el consolidado deduplicado, por tema, con **resolución**: ✅ adopto (→ aplico a specs) · 🔀 fork a
> decidir con el user · 📋 spec-fix mecánico · 🎭 fix del mockup. Severidad: 🔴 CRIT/ALTA · 🟠 MEDIA · ⚪ BAJA.

## 0 · Las 4 convergencias (los 4 coinciden, señal máxima)
1. **Cimientos de dominio faltan antes del FE.** El modelo (identidad calificada, 1:N instalaciones, versión, walker
   multi-arnés, store propio) **no tiene sustrato** hoy. El "Slice 1 reusa el loader" es engañoso: ~60% es código nuevo.
2. **El "drift" está indefinido** (3 referencias contradictorias, `git status` mide contra el proyecto ≠ origen) **y el
   mockup lo fabrica** (`drift:2/3`, `⬆ v1.2.0`) — viola BR-4/BR-8/RN-UPD-1 y el propio header "jamás falso".
3. **La identidad `(registry,id)` se contradice sola:** *fractura* un arnés multi-marketplace (2 cards) Y *separa* dos
   arneses distintos con id colisionado. El discriminador real es **identidad de autor**, no el registry.
4. **La ley anti-drift, como estaba, es falsa por construcción:** las instalaciones se editan FUERA de ArnesIA (por eso
   existe el backport). El invariante "jamás dos copias divergen" no es garantizable; es descriptivo de ArnesIA.

## A · Persistencia + identidad 🔴
- **A1 [código] El registro real es `map[string]string` id→UN path, `Register` REEMPLAZA** (`store/arnes_registry.go:28,78`).
  No hay `(registry,id)` ni 1:N. Y ES el confinamiento de cwd de sesión (`Resolve` hot-path, `WorkdirResolver`), NO el
  portafolio. → ✅ **store de portafolio SEPARADO** (no extender `arneses.json`). 📋 corregir §10.
- **A2 [código] `index/store.go` indexa `s.graphs[id]` desnudo; `Upsert` acepta cualquier `id`** → auto-fusiona
  resuelto↔desconocido (viola C-ID-2). → ✅ clave de índice = identidad calificada; prohibir fusión bare-id.
- **A3 [modelo] `domain.Arnes.Empresa`/`Marketplace` son STRINGS singulares** (`graph.go:36,39`; `types.ts`) → N:M
  irrepresentable; la lente "por empresa" no tiene de dónde sacar N empresas. `store.go:102` hardcodea `"alpacapurpura"`.
  → ✅ **`empresas[]` / `marketplaces[]`** (colección) en dominio + FE. 🔀 requiere tocar dominio (ver recut).
- **A4 [modelo] Identidad = ¿`(registry,id)` o identidad de autor?** `(registry,id)` fractura mirrors y no dedup el mismo
  arnés en 2 markets. → 🔀 **F-A**: propongo **identidad = `(home, id)`** donde `home` = marketplace canónico declarado
  (`arnes.l0.marketplace`, "dónde vivo yo autor"); el **registry-de-adquisición** (de dónde salió ESTA copia) = faceta
  N:M aparte. Mismo `home+id` = mismo arnés aunque venga de mirrors distintos; `home` distinto = arnés distinto. A ratificar.
- **A5 [código] Startup frágil:** `arneses.json` corrupto → `NewArnesRegistry` propaga → `main.go` **brickea el arranque**.
  → ✅ el store del portafolio **degrada honesto** (entrada corrupta = visible+saltada), jamás impide boot.

## B · Detección física 🔴
- **B1 [código] El loader NO reconoce `.claude/plugins/<id>/`** — `detectarElementos` (`loader.go:129-137`) solo ve
  `.claude-plugin/plugin.json` (raíz) o `.claude/` (escanea skills/commands, NO plugins/). → ✅ **walker nuevo** que
  desciende a cada `.claude/plugins/<id>/.claude-plugin/plugin.json` y llama `LoadArnes` por dir. 📋 §7.1 estaba indefinido.
- **B2 [código] La cadena de nombre/id `arnes.l0.nombre→plugin.json name→id` NO está implementada** — `leerManifiesto`
  lee SOLO `arnes.l0.json`; `plugin.json` nunca se parsea (`loader.go:143`, solo `os.Stat`). **Un plugin de marketplace
  normal (con `plugin.json`, sin `arnes.l0.json`) → `Arnes=nil` → ni el `id` se computa.** Es el artefacto MÁS común.
  → ✅ **parsear `plugin.json` como fallback** (id=name, version, description) antes de rendirse. Bloqueante.
- **B3 [código] El detector 3° (lock `.devstudio/arneses.yaml`), FIRMADO v1.1, NO está en el loader.** → 🔀 construirlo
  en el slice de cimientos, o sacar el lock del corte (afecta qué eslabón resuelve origen).

## C · Versión 🔴
- **C1 [modelo] No existe campo `Version` en `domain.Arnes`** (`graph.go:30-42`); el loader no lee versión. Todo lo
  version-based (drift, update-check, estados `desactualizada`/`adelantada`) **no tiene fuente**; el mockup inventó "1.1.0".
  → ✅ agregar `version` (parsear `plugin.json.version` / catálogo). Hasta entonces, honesto = `versión: ?`.

## D · Drift — indefinido + fabricado 🔴 (raíz de H-1/H-2/R-2)
- **D1 [modelo] Referencia de drift indefinida** (canónico local sucio vs marketplace vs `git status` del proyecto =
  3 veredictos contradictorios). → ✅ **referencia = contenido inmutable de `marketplace/plugins/<id>/<versiónInstalada>/`**
  (dir por-versión, siempre reproducible) comparado por **hash de contenido** (árbol forma-plugin, sha256 por path, excluye
  `.git/`); NUNCA por string semver ni por `git status` (eso solo responde "¿el user ya commiteó?"). Sin referencia
  accesible → **`deriva-no-evaluable`**, jamás drift positivo.
- **D2 [nombre] "drift" sin nombre propio** (regla 4 mockups; `origen`/`procedencia` L0 tomados). → 🔀 **F-C**: candidato
  **`deriva`** (estados `en-deriva` / `al-hilo` / `deriva-no-evaluable`). A ratificar.
- **D3 [honestidad] Slice 1 NO puede computar drift** (sin canónico ni marketplace) → 🎭 **default honesto en S1 =
  `deriva-no-evaluable`**; el mockup debe dejar de pintar `drift:N`.

## E · Ley anti-drift — reframe + restricciones 🔴
- **E1 [scope] INV-1..4 son locales a un `~/.arnesia`; el canónico NO es único global.** Máquina A y B clonan, editan,
  publican → last-writer-error (`publisher.go` solo `push`). → ✅ **reformular INV-2/3 como DESCRIPTIVOS de ArnesIA**
  ("ArnesIA nunca INTRODUCE una copia editable divergente nueva; reconcilia las que el mundo ya produjo"); **marketplace =
  SSoT compartida**; **publicar exige pull/rebase-antes-de-push**.
- **E2 [código] checkout por `id` desnudo** (`~/.arnesia/checkouts/<id>/`) → traer `homeB/foo` **pisa** `homeA/foo`.
  → ✅ checkout llaveado por identidad calificada (`checkouts/<home-slug>/<id>/`).
- **E3 [seguridad] `~/.arnesia` es ubicación PROTEGIDA** (`checkProtected` rechaza) → el canónico ahí **no es registrable**
  para "Abrir en Mapa". → 🔀 canónico fuera de `~/.arnesia` (p.ej. `~/.arnesia-checkouts/` o dir elegible) **o** excepción
  explícita. A decidir.
- **E4 [dominio] Reparar pisa superficies COMPARTIDAS** (`settings.json`/`.mcp.json` = del proyecto, compartidas entre
  plugins). → ✅ reparar overwrite SOLO `.claude/plugins/<id>/` (dir privado); superficies compartidas = **merge aditivo
  con diff**, jamás overwrite (respeta §9 "② no se escribe en ③").
- **E5 [boundary] Reparar sobre proyecto con lock DevStudio** (read-only para ArnesIA) → deja el lock mintiendo. → ✅
  **bloqueado/delegado** si hay lock presente.
- **E6 [robustez] "Traer canónico" trae HEAD, no la versión observada** → backport cross-versión silencioso. → ✅ traer
  canónico permite elegir la versión que empata la instalación; backport exige alineación de versión base o rebase explícito.
- **E7 [robustez] Publicar: tag antes de push confirmado → dos "v2" distintas.** → ✅ tag SOLO tras push OK (o rollback).
- **E8 [concurrencia] Lock in-process no cubre cross-proceso/máquina; backport no está en C-CON-1.** → ✅ **lock de
  filesystem** sobre el checkout cubriendo {mejorar, backport, publicar}.
- **E9 [aliasing] canónico∩instalación no prohibido** (symlink/anidamiento → editar canónico edita instalación).
  → ✅ paths disjuntos; detector rechaza/avisa si coinciden. **checkout ≠ instalación** (cierra el doble-conteo dogfood).
- **E10 [estados] Estados del canónico ortogonales** (limpieza × actualidad; "sucio Y desactualizado" no representable).
  → ✅ estado = tupla (limpio/sucio) × (al-día/atrás/adelante) + ruta "publicar estando atrás".

## F · Resolución de origen 🔴
- **F1 [algoritmo] "Para en primer hit" invierte autoridad + hace imposible detectar conflicto (C-OR-6).** El manifiesto
  viaja en los bytes (igual en toda copia) → mala fuente de *procedencia de la copia*; el **lock** (install-time) es el
  record real de dónde salió ESTA copia. → ✅ la cadena **recolecta TODOS los eslabones disponibles y reconcilia/marca
  discrepancia**, no corta. Para *procedencia* → **lock > manifiesto**; el manifiesto da `home` (autor-declarado), no origen-de-copia.
- **F2 [normalización] `registry`/remotes sin canonicalizar** (slug `owner/repo` vs URL `https://…`/`git@…:` vs `.git`)
  → mismo marketplace, claves distintas → cards duplicadas; el canónico traído por URL no fusiona con instalaciones por slug.
  → ✅ **canonicalización** (host+owner+repo, sin esquema/`.git`/trailing) ANTES de comparar/keyear.
- **F3 [clave provisional] `(⟂desconocido, id, projectPath)` colapsa instalaciones distintas** (monorepo: 2 `.claude/`
  bajo un path → se fusionan) **y duplica** (git-clone re-hecho → path efímero nuevo). → ✅ clave = `(⟂, id, install-path-
  relativo-al-scope)` local / `(⟂, id, project-remote-normalizado)` git.
- **F4 [rename] `id` cambia entre versiones → update-check busca el id viejo, no lo halla → notificación ciega** (mata el
  valor multi-máquina). → 🔀 tabla de alias `id_previo→id_actual` (slice de update); por ahora documentar el hueco.
- **F5 [eslabón 4] "¿dónde registra CC la procedencia de un plugin?" es un unknown load-bearing** — si CC la registra,
  sería más autoritativo que el manifiesto y reordena la cadena. → ✅ investigación embebida ANTES de firmar el orden.

## G · Honestidad del mockup + estados UI 🟠 (todos → 🎭 fix)
- **G1** `drift:2/3` fabricado (H-1) · **G2** `⬆ v1.2.0` update fabricado, es Slice 4 (H-2) · **G3** rama Marketplace
  con validación ✓ **exitosa hardcodeada + no deshabilitada** (H-3, viola BR-8 + su propio header) · **G4**
  `legal-administrativo` con `empresa:"Vitalia"` sobre origen desconocido (sale del path = BR-5 prohibido) + `origen?` y
  `drift` a la vez (H-4).
- **G5** faltan estados: vacío/cargando/error en Lista y Wizard; `presente-sucio` (drawer no lee `dirty`);
  `no-encontrada`/`deriva-no-evaluable`/`desactualizada`/`adelantada` en instalación; advertencia canónico-sucio al desvincular.
- **G6** mockup keyea por `id` desnudo → no puede representar la identidad calificada (D-2).
- **G7** "…y borrar clon local" se ofrece sin canónico + sin confirmación aparte (RN-UNL-2).
- **G8** a11y: drawer sin `role=dialog`/foco atrapado/Esc; wizard sin `aria-modal`; sin reduced-motion.
- **G9** dot de salud sin regla de cómputo (mock incoherente: drift:2→ok, drift:1→warn). → ✅ definir regla o quitar de S1.

## H · Coherencia / vocabulario 🟠
- **H1** Colisión de nombre: **CAP-58** ya es "Mapa/**portafolio**/inspector" + `GET /api/harnesses` rotulado "portafolio".
  → 🔀 desambiguar el nuevo front-door (¿"Portafolio de arneses" vs el portafolio-grafo del Mapa?).
- **H2** Slice de **backport sin asignar** (funcional dice S2/5, usabilidad S2, decisiones-5-slices no lo lista). → 📋 asignar.
- **H3** "Qué eslabones de origen resuelve Slice 1" tiene **4 respuestas distintas** entre docs. → 📋 unificar tras F5.
- **H4** Trazabilidad §9 no cita `conformance` (CAP-30/31) ni map-service (CAP-58/61) que Observar reusa. → 📋 completar.

## Recut propuesto (consecuencia de A-C-E)
El FE honesto de Slice 1 **no es construible** sin cimientos. Propongo partir en dos:
- **Slice 0 · Cimientos del Portafolio (dominio/backend):** store de portafolio separado (degrada honesto) · walker
  `.claude/plugins/*` · fallback `plugin.json` · campo `version` · `empresas[]`/`marketplaces[]` · identidad calificada
  `(home,id)` + canonicalización · cadena de origen collect-all (lock>manifiesto) · `deriva` = hash vs marketplace-por-versión
  (o `deriva-no-evaluable`). Todo con capabilities + tests. Investigar eslabón CC (F5).
- **Slice 1 · FE Portafolio (sobre los cimientos):** Lista + Wizard-Proyecto/carpeta + Drawer READ + Observar + Desvincular,
  con **defaults honestos** (deriva-no-evaluable, update no-verificado, marketplace disabled). Mockup corregido (G1-G9).

→ 🔀 **F-B**: ¿partimos así (cimientos → FE) o encogemos Slice 1 a un mínimo distinto? (decisión del user).
