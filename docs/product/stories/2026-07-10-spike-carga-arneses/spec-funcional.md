# Spec funcional — Portafolio · ciclo de vida del arnés

> `tipo: spec` · paquete `2026-07-10-spike-carga-arneses` · deriva de `decisiones.md` S-D5…S-D10 (modelo cerrado) +
> mockup `mockups/arnesia-portafolio.html` v2 + revisión adversaria (`revision-adversaria.md`, 4 subagentes).
> Anclajes de código: `internal/adapters/loader/` · `internal/adapters/store/arnes_registry.go` · `internal/domain/graph.go`
> · `internal/adapters/publish/` · `internal/adapters/selfupdate/` · contrato `docs/architecture/contracts/nomenclatura-arnes.md` (v1.1).
> Vig: borrador post-revisión, listo para firmar el modelo. NO firmado.

## 0 · Nombre y desambiguación
**«Portafolio (de arneses)»** = este front-door del ciclo de vida (registro de identidades de arnés que el usuario
agregó). **NO** es el «portafolio» de CAP-58 (ensamblado Mapa/inspector de UN arnés). Cuando haya ambigüedad, este es
el **Portafolio de arneses**; aquél es el **lienzo del Mapa**. (Hallazgo H1.)

## 1 · Propósito y alcance
El Portafolio de arneses es el front-door del ciclo de vida (VISION crear·mapear·observar·mejorar): desde la app el
usuario **agrega**, **observa**, **mejora**, **publica**, **actualiza** y **repara** arneses que viven en muchos lugares.
Este documento fija el **modelo funcional del programa** y el **recut en dos primeros slices** (S-D10):
- **Slice 0 · Cimientos (dominio/backend):** el modelo construible con capabilities + tests, SIN FE. §9.
- **Slice 1 · FE Portafolio:** las 3 superficies sobre los cimientos, con defaults honestos. §10.
Slices 2-5 (marketplace/clonar · publicar · update · reparar) se especifican a nivel de reglas; se construyen después.

## 2 · Entidades (glosario canónico)

| Entidad | Definición | Fuente física |
|---|---|---|
| **Arnés (identidad)** | Un harness. **Clave = `(home, id)`** (§3). Unidad del Portafolio: **1 card por identidad**. | `arnes.l0.json` / fallback `plugin.json` |
| **`home`** | Marketplace canónico **autor-declarado** (`arnes.l0.marketplace`). «Dónde vivo yo». Parte de la identidad. | `arnes.l0.json`.marketplace |
| **Canónico (checkout)** | La **única copia editable en ESTE install**. Autorear/mejorar/publicar. 0 o 1 por identidad. | `<checkouts>/<home-slug>/<id>/` (clon git de `home`) |
| **Instalación** | Materialización del arnés en un proyecto. **Espejo read-only de observación.** N por identidad. | `.claude/plugins/<id>/` (forma-plugin) + lock |
| **Proyecto** | Carpeta/repo con ≥1 instalación. **Siempre = procedencia** de la instalación (no entidad de 1er nivel). | carpeta local o repo GitHub clonado |
| **Marketplace / registry-de-adquisición** | Repo GitHub prenter (`marketplace.json` + `plugins/<id>/<versión>/`). De dónde salió una copia = **faceta N:M**, no la identidad. | repo git (ej. `~/Proyectos/marketplace-arneses`) |
| **Lock** | `.devstudio/arneses.yaml` — puntero de descubrimiento **read-only** (detector 3°, `id·versión·canal·registry`). | raíz del proyecto |
| **Registro del Portafolio** | Store **nuevo y separado** (NO `arneses.json`, que es cwd de sesión): identidades `(home,id)` + facetas + presencias. Degrada honesto. | nuevo store en `~/.arnesia` |

## 3 · Identidad, relaciones y clave (F-A cerrada)

- **Clave de identidad = `(home, id)`** (S-D10/F-A). `home` = marketplace autor-declarado (`arnes.l0.marketplace`).
  Mismo `home+id` = **mismo arnés** aunque se haya adquirido de mirrors distintos → **1 card** (resuelve el fracture del
  `(registry,id)`). `home` distinto = arnés distinto → cards separadas (resuelve la colisión de `id`).
- **RN-IDENT-1 · Canonicalización.** `home` y todo git-remote se **canonicalizan** (host+owner+repo, sin esquema/`.git`/
  trailing/case) ANTES de comparar o keyear. Slug `owner/repo` y URL `https://…`/`git@…:` del mismo repo colapsan a una clave.
- **RN-IDENT-2 · Sin `home` resoluble** → identidad provisional `(⟂sin-home, id, scope)` donde `scope` = install-path
  relativo-al-proyecto (local) o project-remote-canonicalizado (git). **NO** se fusiona con ninguna identidad resuelta
  automáticamente (podrían ser distintas); fusión = acción explícita del usuario (C-ID-2).
- **RN-IDENT-3 · Precedencia intra-artefacto** para el `id`: `arnes.l0.json`.id → `plugin.json`.name → nombre del dir;
  discrepancia entre ellos = check visible, no elección silenciosa.
- **Relaciones:** Arnés ↔ Empresa = **N:M** · Arnés ↔ registry-de-adquisición = **N:M** · Arnés → Instalación = **1:N** ·
  Arnés → Canónico = **1:0..1**. `empresa` y `registry` son **facetas/lentes**, no dueños. ⇒ el dominio guarda
  `empresas[]` y `registries[]` (colecciones), no escalares (delta D-DOM-2).
- **RN-IDENT-4 · checkout ≠ instalación:** un path detectado que **es** un checkout conocido se clasifica CANÓNICO, jamás
  instalación (cierra el doble-conteo y la deriva auto-referencial del dogfood). Paths de canónico e instalación **disjuntos**
  (sin symlink/anidamiento); el detector rechaza/avisa si coinciden.

## 4 · Ley anti-drift (reformulada DESCRIPTIVA — E1..E10)

> Las instalaciones **se editan fuera de ArnesIA** (usuario/DevStudio) — es la premisa, no un caso evitado. Por eso los
> invariantes son **descriptivos de ArnesIA**, no del mundo:
> **INV-1 · ArnesIA nunca introduce una copia editable divergente nueva.** El único artefacto que ArnesIA deja editar es
> el **canónico**; ninguna instalación se edita en sitio desde la app.
> **INV-2 · Reconciliación, no prevención.** ArnesIA **observa** (read) y **re-materializa** (reparar) instalaciones, y
> **backportea** su aprendizaje al canónico. No pretende que el mundo no haya divergido; lo reconcilia.
> **INV-3 · `home` (marketplace) = SSoT compartida.** Cross-máquina, la verdad vive en el marketplace git. Publicar exige
> **pull/rebase-antes-de-push**; el tag se crea **solo tras push OK** (o se revierte). Nunca «dos v2 distintas».
> **INV-4 · Sin canónico no hay autoría.** Solo observar una instalación de cliente → autorear/mejorar/publicar bloqueados
> hasta «traer el canónico de `home`».

## 5 · Estados (presencia derivada, honesta)

- **Canónico = tupla ortogonal** (limpieza ∈ {limpio, sucio}) × (actualidad ∈ {al-día, atrás, adelante}) ∪ {ausente}.
  «sucio Y atrás» es representable; incluye ruta «publicar estando atrás» (pull→merge→conflicto antes de push).
- **Instalación:** `alineada` · `en-deriva` · `desactualizada` · `adelantada` · `origen-sin-resolver` · `no-encontrada`.
- **Deriva (F-C):** `en-deriva` / `al-hilo` / `deriva-no-evaluable`. **Referencia = contenido inmutable de
  `home/plugins/<id>/<versiónInstalada>/`** comparado por **hash de contenido** (árbol forma-plugin, sha256 por path,
  excluye `.git/`; declara si excluye superficies compartidas). **NUNCA** semver-string ni `git status` (eso solo responde
  «¿el usuario ya commiteó?»). Sin referencia accesible (ni marketplace, ni canónico-limpio-en-esa-versión) → `deriva-no-evaluable`.

## 6 · Resolución de origen (cadena collect-all + reconcilia — F1..F5)

Al detectar una instalación se **recolectan TODOS los eslabones disponibles** (no «para en el primer hit») y se reconcilia,
anotando la fuente de cada dato (trazabilidad honesta). Autoridad para la **procedencia de la copia** = **lock > manifiesto**
(el manifiesto viaja en los bytes → declara `home` autor, no de-dónde-salió-la-copia).

1. **`arnes.l0.json`.marketplace** → `home` + `empresa(s)` (autor-declarado). [define la identidad]
2. **Lock `.devstudio/arneses.yaml`** → `registry`(de-adquisición) · `versión` · `canal`. [procedencia de la copia — MANDA]
3. **`.claude/plugins/<id>/.git`** remote → registry-de-adquisición si la copia se instaló por clon.
4. **Metadata de plugins de CC** (dónde CC registre el marketplace del plugin). **UNKNOWN load-bearing** — si CC lo registra,
   puede ser más autoritativo que 2/3 y reordena; **se investiga en Slice 0 ANTES de firmar el orden** (F5).
5. **`.git/config` del PROYECTO** → hogar del proyecto (para deriva-por-commit y correlación), **NO** el `home`/registry del
   arnés (RN-GIT-1). Resolver `gitdir:` si es worktree; `origin` gana, `upstream` = dato secundario; normalizar SSH↔HTTPS.
6. Conflicto entre eslabones (ej. manifiesto.home ≠ lock.registry) → **discrepancia visible**, no elección silenciosa (C-OR-6).
7. Nada resuelve `home` → identidad provisional (RN-IDENT-2), origen `desconocido` visible. `empresa` **no** se infiere del path.

## 7 · Operaciones (pre/post + invariantes)
*(Igual que la versión previa en intención; cambios clave marcados. Detalle adverso → `casuistica.md`.)*
- **7.1 Agregar PROYECTO** — fuente carpeta local | repo GitHub (clonar). **Walker nuevo** (D-DOM-3) desciende a cada
  `.claude/plugins/<id>/.claude-plugin/plugin.json` + lee lock + settings-enabled → `LoadArnes` por dir (con **fallback
  `plugin.json`**, D-DOM-4) → resuelve origen (§6) → el usuario elige → instalaciones read-only bajo su identidad. Profundidad
  de walk acotada + cancelable (C-P-11/13). Detección git (§6.5) = **código nuevo** (go-git o parseo `.git/config`).
- **7.2 Agregar de MARKETPLACE** (S2) — git url → validar `marketplace.json` → listar → elegir → clonar canónico a
  `<checkouts>/<home-slug>/<id>/`. Fusiona con instalaciones `(⟂,id,*)` del mismo id **solo por vinculación explícita** (C-ID-2).
- **7.3 Observar/Diagnosticar** (Slice 1 READ) — Abrir instalación en Mapa (reusa CAP-58/61) + `conformance --arnes` (CAP-30/31). No edita.
- **7.4 Reparar** (S5) — re-materializar desde canónico: **overwrite SOLO `.claude/plugins/<id>/`** (dir privado);
  superficies compartidas (`settings.json`/`.mcp.json`) = **merge aditivo con diff**, jamás overwrite (E4). **Bloqueado si hay
  lock DevStudio** (read-only, E5). Destructivo del `en-deriva` → confirmar + ofrecer backport; re-evaluar deriva justo antes (TOCTOU).
- **7.5 Backport** (S2) — aprendizaje instalación→canónico. Exige canónico **en la versión de la instalación fuente** (o rebase
  explícito, E6). Conflictos se muestran. Único puente instalación→canónico.
- **7.6 Publicar** (S3) — pull/rebase → conformance-verde sobre árbol commiteado → bump semver → **changelog obligatorio** →
  push → **tag solo tras push OK** (E7). Push rechazado → sin tag huérfano.
- **7.7 Update-check** (S4) — comparar versión canónico/instalación vs último tag de `home`. Offline → `no-verificado`. Alias
  de rename `id_previo→id_actual` para no cegar el aviso (F4).
- **7.8 Desvincular** — quita del registro. No desinstala, no borra clon. Opción «…y borrar clon» = destructiva, confirmación
  aparte, deshabilitada si no hay canónico. Advertir si canónico `sucio` (RN-UNL-3).

## 8 · Reglas de negocio (verificables)
- **BR-1** Unidad = identidad `(home, id)` canonicalizada; nunca card por instalación.
- **BR-2** Solo el canónico es editable; la app no expone editor sobre una instalación; paths canónico∩instalación = ∅.
- **BR-3** Resolución de origen **honesta y trazable** (anota fuente por dato); sin `home` → provisional/`desconocido`; `empresa` jamás del path.
- **BR-4** `deriva(inst)` = `∃ hash(inst) ≠ hash(home/plugins/<id>/<versiónInst>/)`; sin referencia accesible → `deriva-no-evaluable`. Prohibido semver-string o `git status` como veredicto.
- **BR-5** El remote git del **proyecto** ≠ el `home`/registry del arnés.
- **BR-6** Reparar overwrite SOLO el dir privado; superficies compartidas = merge; deriva re-evaluada justo antes; confirmación si `en-deriva`.
- **BR-7** Publicar: pull→conformance-verde→bump→push→tag-tras-push; changelog obligatorio; sin tag huérfano.
- **BR-8** Lo no construido = deshabilitado + tooltip, jamás simulado.
- **BR-9** Duplicado: proyecto ⇔ mismo remote canonicalizado (o path canónico); canónico/identidad ⇔ misma `(home,id)`.
- **BR-10** GitHub: app = conductor (`gh`/PAT user-owned); nunca proxya login hosteado.
- **BR-11** El registro del Portafolio **degrada honesto** (entrada corrupta = visible+saltada); jamás impide el arranque.

## 9 · Slice 0 — Cimientos (build inmediato, dominio/backend, con capabilities + tests)
**Entra (sin FE):**
- **Store de portafolio nuevo** (separado de `arneses.json`; degrada honesto, BR-11).
- **Walker** `.claude/plugins/<id>/` multi-arnés + **fallback `plugin.json`** en el loader (id=name/version/description).
- **Campo `version`** en el dominio (parseo `plugin.json.version`).
- **`empresas[]` / `registries[]`** (colecciones) en `domain.Arnes` + reconciliar seed hardcodeado (`store.go:102`).
- **Identidad `(home,id)`** + canonicalización de registry/remotes + clave provisional con scope.
- **Cadena de origen collect-all** (§6, eslabones 1-2-3-5; investigar **eslabón 4** ANTES de firmar el orden, F5).
- **Deriva** = hash vs `home/plugins/<id>/<versión>/` (o `deriva-no-evaluable` cuando no hay referencia).
- **Detector 3° del lock** `.devstudio/arneses.yaml` (o decisión explícita de diferirlo).
- Ubicación del canónico **fuera de `~/.arnesia`** (o excepción explícita a `checkProtected`) — E3, a decidir en build.

**Capabilities (nuevas):** `portafolio.registrar-identidad` · `portafolio.escanear-proyecto` (walker) · `portafolio.resolver-origen`
· `portafolio.evaluar-deriva` · `portafolio.desvincular`. **Reusa/extiende:** `loader.reconocer-forma-fisica` (CAP-15, +plugin.json)
· `loader.leer-manifiesto` (+fallback) · registry (nuevo store, NO extiende el de cwd). Ningún código sin capability (R1-R4).

## 10 · Slice 1 — FE Portafolio (sobre los cimientos)
Las 3 superficies del mockup v2 con **defaults honestos**: deriva `deriva-no-evaluable` mientras no haya `home` accesible;
update `no-verificado`; rama Marketplace **deshabilitada + tooltip** (S2); estados vacío/cargando/error; a11y (§spec-usabilidad).
Reusa Mapa (CAP-58/61) + conformance (CAP-30/31) para «Observar». Porte a Storybook (SSoT). **Fix de honestidad del mockup
(G1-G9) se aplica aquí.**

## 11 · Deltas de dominio requeridos (D-DOM)
- **D-DOM-1** `domain.Arnes.Version` (nuevo). **D-DOM-2** `Empresa string`→`Empresas []string`, `Marketplace string`→
  `Home string`+`Registries []string`. **D-DOM-3** walker multi-arnés. **D-DOM-4** loader parsea `plugin.json` (fallback).
  **D-DOM-5** store de portafolio separado. **D-DOM-6** detector lock. **D-DOM-7** canonicalización de registry/remote.
