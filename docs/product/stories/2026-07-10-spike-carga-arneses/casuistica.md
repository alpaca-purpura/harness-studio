# Casuística — Portafolio · escenarios adversos + restricciones

> `tipo: casuistica` · complementa `spec-funcional.md`. Matriz de **escenarios adversos**: qué puede salir mal en cada
> flujo, comportamiento esperado, restricción, y **fallback honesto** (jamás pass fabricado). Esta hoja es el blanco
> de la revisión adversaria (subagentes): buscan huecos, comportamiento indefinido o incorrecto, y restricciones faltantes.
> Convención: `C-XX-n` id de caso · ⛔ = restricción dura · 🟡 = fallback honesto · 🔒 = toca boundary firmado.
>
> **Terminología post-revisión (S-D9/S-D10):** «drift» → **`deriva`** (referencia = hash vs `home/plugins/<id>/<versión>/`
> inmutable; sin referencia → `deriva-no-evaluable`). Identidad = **`(home, id)`** (home = marketplace autor-declarado);
> registry-de-adquisición = faceta. Cadena de origen = **collect-all + reconcilia** (lock > manifiesto para procedencia de la
> copia). Las tablas de abajo siguen válidas como escenarios; su resolución fina se rige por `spec-funcional.md` (rev) + §I nueva.

## A · Agregar PROYECTO (fuente carpeta local | repo GitHub)

| id | Escenario adverso | Comportamiento esperado |
|---|---|---|
| C-P-1 | Carpeta local **no es repo git** (sin `.git`) | Agregar OK. Drift por comparación de contenido (no git). Flag «sin control de versiones». 🟡 |
| C-P-2 | Carpeta local **ES repo git** → leer `.git/config` remote | Registrar remote como hogar del proyecto (probablemente clonado). Habilita drift por `git status`. **NO** confundir con marketplace del arnés (BR-5). ⛔ |
| C-P-3 | Path no existe / sin permiso de lectura | Error honesto con el motivo; no se agrega. 🟡 |
| C-P-4 | Carpeta existe pero **sin `.claude/` ni plugins** | «No encontré arneses instalados aquí» (no es error). Ofrecer elegir otra carpeta. 🟡 |
| C-P-5 | `.claude/plugins/<id>/` **malformado** (sin `plugin.json`) | Mostrar como **no-reconocible visible** (loader `ClaseNoReconocido`), no crashear, resto de arneses siguen. 🟡 |
| C-P-6 | Fuente GitHub, repo **privado**, sin auth | «Repo privado — autenticá (`gh`/PAT)» antes de clonar. 🔒 |
| C-P-7 | GitHub url inválida / inalcanzable / no es repo | Error honesto; no se agrega. 🟡 |
| C-P-8 | Proyecto **ya agregado** (mismo path/remote) | No duplicar. Ofrecer **re-escanear** (detecta nuevas/removidas instalaciones). BR-9 |
| C-P-9 | Proyecto movido/borrado **después** de agregar | Instalación `no-encontrada` visible; ofrecer re-apuntar path o desvincular. 🟡 |
| C-P-10 | Mismo `id` instalado en **varios proyectos** | Cada uno = instalación separada bajo la MISMA identidad (dedup por `(registry,id)`). ✓ |
| C-P-11 | Monorepo: **varios `.claude/`** en subcarpetas | Walker recorre recursivo acotado; cada `.claude/` = un ámbito de instalación. Decidir profundidad máx (evitar escaneo infinito). ⛔ |
| C-P-12 | Symlinks / carpetas cíclicas en el árbol | Walker no sigue symlinks fuera del root; corta ciclos. ⛔ |
| C-P-13 | Proyecto enorme (miles de archivos) | Escaneo acotado a `.claude/` + lock; no leer todo el árbol. Progreso/cancelable. |
| C-P-14 | `.claude/settings.json` lista un plugin **enabled** pero el dir no existe | Instalación declarada-pero-ausente: visible honesto, no fabricar nodo. 🟡 |

## B · Resolución de ORIGEN (marketplace/empresa de una instalación)

| id | Escenario | Comportamiento |
|---|---|---|
| C-OR-1 | `arnes.l0.json` trae `marketplace`+`empresa` | Resolución eslabón 1 (autor). Anotar fuente = «declarado». |
| C-OR-2 | Sin `arnes.l0.json`, hay lock `.devstudio/arneses.yaml` | Eslabón 2. Anotar fuente = «lock». |
| C-OR-3 | Sin manifiesto ni lock, pero `.claude/plugins/<id>/.git` existe | Eslabón 3: remote del plugin = origen del arnés. Anotar «git-plugin». |
| C-OR-4 | Solo `.git/config` del proyecto (insight usuario) | Eslabón 5: es el hogar del **proyecto**, NO el marketplace. Exponer como proyecto, marketplace sigue `desconocido`. ⛔ BR-5 |
| C-OR-5 | Nada resuelve | `origen desconocido` visible. Acción «Resolver origen» (manual/investigación). 🟡 |
| C-OR-6 | Manifiesto dice marketplace **A**, lock dice **B** (conflicto) | Mostrar **conflicto visible**, no elegir en silencio. Preferir declarado (manifiesto) + marcar discrepancia. 🟡 |
| C-OR-7 | `registry` resuelto pero el marketplace ya no existe/movió | Origen «resuelto pero inalcanzable»; distinto de desconocido. 🟡 |
| C-OR-8 | Versión instalada no parseable / ausente | `versión: ?`; drift-no-evaluable si no hay referencia. BR-4 |

## C · Identidad y colisiones (N:M:M)

| id | Escenario | Comportamiento |
|---|---|---|
| C-ID-1 | Mismo `id` en **dos marketplaces** distintos | Entradas **distintas** por clave `(registry,id)`. No fusionar. RN-IDENT |
| C-ID-2 | Instalación con origen **desconocido** + existe canónico del mismo `id` | **NO** fusionar automáticamente (podrían ser distintos). Ofrecer «vincular a esta identidad» explícito. ⛔ |
| C-ID-3 | Arnés base sirve a **N empresas** | 1 sola card; aparece en N grupos bajo la lente empresa (esperado). Desvincular lo quita de todas (es 1 entrada). |
| C-ID-4 | `id` cambia entre versiones (rename) | Tratar como identidad nueva; ofrecer «es el mismo que X» manual. 🟡 |
| C-ID-5 | Dos instalaciones misma identidad, **versiones distintas** | Ambas bajo la card; cada una con su versión/estado. Update-check por instalación. |

## D · Agregar de MARKETPLACE + clonar canónico

| id | Escenario | Comportamiento |
|---|---|---|
| C-M-1 | Url es repo git pero **sin `marketplace.json`** | Error «no tiene estructura de marketplace». No listar. RN-MKT-1 🟡 |
| C-M-2 | Url inválida/inalcanzable | Error honesto. 🟡 |
| C-M-3 | Marketplace **privado** | Auth (`gh`/PAT). 🔒 |
| C-M-4 | Marketplace válido pero **0 plugins** | «Válido, vacío». No falso-listado. 🟡 |
| C-M-5 | Un plugin del marketplace tiene manifiesto **corrupto** | Se muestra como «no válido», no ofrecible; resto sí. 🟡 |
| C-M-6 | Marketplace **ya agregado** | No duplicar; refrescar índice. BR-9 |
| C-M-7 | Red se corta a mitad del clone | Estado parcial → error + reintento; no dejar checkout corrupto silencioso. 🟡 |
| C-C-1 | Canónico **ya existe** (`~/.arnesia/checkouts/<id>/`) | No re-clonar; abrir detalle («ya lo tenés»). BR-9 |
| C-C-2 | Dir del checkout existe pero **sucio/corrupto** | Honesto; ofrecer re-clonar o reparar checkout. 🟡 |
| C-C-3 | Disco lleno / sin permiso de escritura | Error honesto; no dejar a medias. 🟡 |
| C-C-4 | Versión pedida del marketplace no existe | Error; ofrecer versiones disponibles. 🟡 |

## E · Observar / Reparar / Backport (ciclo instalación↔canónico) — INV anti-drift

| id | Escenario | Comportamiento |
|---|---|---|
| C-OB-1 | Observar instalación cuyo path ya no existe | `no-encontrada`; no abrir Mapa vacío. 🟡 |
| C-OB-2 | Drift **no evaluable** (sin canónico ni marketplace accesible) | Mostrar `drift-no-evaluable`, no «sin drift». BR-4 🟡 |
| C-OB-3 | Instalación **adelantada** (versión > canónico) | Estado honesto; sugerir backport/actualizar canónico, no reparar a la baja sin avisar. |
| C-REP-1 | Reparar instalación **con drift** | ⛔ Destructivo: confirmar + ofrecer «backport primero». RN-REP-1 |
| C-REP-2 | Reparar **sin canónico** | ⛔ Bloqueado; CTA traer canónico. RN-REP-2 |
| C-REP-3 | Reparar, proyecto es **repo git** con cambios sin commitear | No auto-commitear; dejar en árbol, avisar. RN-REP-3 |
| C-REP-4 | Sin permiso de escritura en el proyecto | Error honesto; no reparar parcial. 🟡 |
| C-REP-5 | Reparar re-cablea settings/hooks/mcp y **pisa** ajustes locales del cliente | Advertir qué se recablea; **jamás fundir en `.claude/` fuera de la forma-plugin** (§9). ⛔ |
| C-BK-1 | Backport **sin canónico** | ⛔ Requiere traer canónico. RN-BACK-1 |
| C-BK-2 | Backport con **conflicto** de merge | Mostrar conflicto, no auto-resolver. RN-BACK-2 |
| C-BK-3 | Editar instalación directo (intento) | ⛔ **No existe** ese camino en la app (INV-1). La UI no lo ofrece. |

## F · Publicar / Update / Desvincular

| id | Escenario | Comportamiento |
|---|---|---|
| C-PUB-1 | Publicar con conformance **rojo** | ⛔ Bloqueado. RN-PUB-1 |
| C-PUB-2 | Publicar **sin changelog** | ⛔ Bloqueado (changelog obligatorio). RN-PUB-3 |
| C-PUB-3 | Push **rechazado** (alguien publicó antes / versión existe) | Error, no sobrescribe; pedir rebase/bump. RN-PUB-5 🟡 |
| C-PUB-4 | Publicar sin auth GitHub | Auth (`gh`/PAT). 🔒 |
| C-UPD-1 | Update-check **offline** / marketplace inaccesible | «No verificado» (jamás «al día» fabricado). RN-UPD-1 🟡 |
| C-UPD-2 | Marketplace fue **borrado/movido** | Origen resuelto-inalcanzable; no crashear. 🟡 |
| C-UPD-3 | Versión del marketplace **menor** que la instalada (downgrade) | No sugerir «actualizar»; mostrar honesto. |
| C-UNL-1 | Desvincular con canónico **sucio** (cambios sin publicar) | ⛔ Advertir antes. RN-UNL-3 |
| C-UNL-2 | Desvincular + **borrar clon** | Confirmación destructiva aparte. RN-UNL-2 |
| C-UNL-3 | Desvincular arnés con instalaciones activas | Solo saca del Portafolio; instalaciones siguen en los proyectos (no desinstala). RN-UNL-1 |

## G · Transversales (git · auth · offline · concurrencia · filesystem)

| id | Escenario | Comportamiento |
|---|---|---|
| C-GIT-1 | `git`/`gh` **no instalado** en la máquina | Detectar; degradar honesto (carpeta local sí, clone GitHub no). 🟡 🔒 |
| C-GIT-2 | `gh` presente pero **no autenticado** | Ofrecer PAT fallback. 🔒 BR-10 |
| C-GIT-3 | PAT **inválido/expirado** | Error honesto de auth; no loop silencioso. 🔒 |
| C-AUTH-1 | Guardar PAT | User-owned, local `~/.arnesia`, permisos restringidos; nunca al log/telemetría. 🔒 |
| C-CON-1 | Dos operaciones sobre el **mismo canónico** a la vez (publicar + mejorar) | Serializar / lock; una a la vez (patrón selfupdate «un update a la vez»). ⛔ |
| C-CON-2 | App cerrada a mitad de un clone/escaneo | Reanudar limpio o marcar entrada parcial; no registro corrupto. 🟡 |
| C-FS-1 | `~/.arnesia/` no escribible | Error honesto al inicio de cualquier op que escriba. 🟡 |
| C-FS-2 | Path con espacios / unicode / muy largo | Manejo correcto (quoting, normalización). |
| C-FS-3 | Ruta relativa donde se espera absoluta (registry ya valida abs) | Rechazo con motivo (reusa validación `arneses.go`). 🟡 |

## H · Casos que EXPONEN huecos de diseño (para la revisión adversaria)
- ¿Qué pasa si el **canónico y una instalación tienen la misma versión pero contenido distinto** sin que la instalación
  «driftee» según git (fue editada antes de commitear)? → definir referencia de drift. (candidato: hash de contenido de la forma-plugin.)
- ¿Un arnés puede ser **canónico Y estar instalado en el mismo repo** (dogfood)? → sí; la card lo muestra como ambos.
- ¿El lock lista un arnés que **no está físicamente** en `.claude/plugins` (cache/rehidratación)? → entrada del lock
  no resoluble = check rojo visible (HS-12), no omitir.
- **Nombre del «drift»** — RESUELTO: `deriva` (S-D10/F-C). Referencia de deriva — RESUELTO: hash vs `home/plugins/<id>/<versión>/` (S-D9 D1).

## I · Nuevos escenarios adversos confirmados (revisión de 4 subagentes) + resolución

| id | Escenario | Resolución |
|---|---|---|
| C-N-1 | Proyecto es **git worktree / submódulo / bare repo** (`.git` es archivo `gitdir:`, o config en otro lado) | Resolver `gitdir:`; submódulo hereda remote del contenedor solo si se declara; bare (sin working tree) → no es proyecto escaneable. ⛔ |
| C-N-2 | **Multi-remote** (`origin` fork ≠ `upstream` canónico) | `origin` gana como hogar del proyecto; `upstream` = dato secundario; jamás elegir en silencio ambiguo. ⛔ |
| C-N-3 | Dos «proyectos» que **canonicalizan al mismo path** (bind-mount, symlink, FS case-insensitive macOS `~/Proj`≡`~/proj`) | Dedup por path canónico + normalización de case; C-P-8 no basta con igualdad textual. |
| C-N-4 | **Store del portafolio corrupto** / editado a mano | **Degrada honesto** (entrada corrupta = visible + saltada); **jamás** impedir el arranque (hoy `arneses.json` corrupto brickea `main.go`). ⛔ BR-11 |
| C-N-5 | **Dos checkouts canónicos** de la misma `(home,id)` (clonó el marketplace dos veces) | Viola «canónico 0..1»; detectar y forzar elegir uno. ⛔ |
| C-N-6 | Canónico en `~/.arnesia/checkouts/` = **ubicación PROTEGIDA** (`checkProtected` rechaza registrar) | Canónico **fuera de `~/.arnesia`** o excepción explícita; sin esto «Abrir en Mapa» del canónico falla. ⛔ E3 |
| C-N-7 | Reparar **pisa superficies compartidas** (`settings.json`/`.mcp.json` del proyecto) | Overwrite SOLO `.claude/plugins/<id>/`; compartidas = merge aditivo con diff. ⛔ E4 |
| C-N-8 | Reparar sobre proyecto **con lock DevStudio** (read-only para ArnesIA) | **Bloqueado/delegado**; no romper la propiedad del lock. ⛔ E5 boundary HS-12 |
| C-N-9 | Publicar: **tag antes de push confirmado** → dos «v2» distintas | Tag SOLO tras push OK (o rollback); pull-antes-de-push. ⛔ E7 |
| C-N-10 | «Traer canónico» trae **HEAD**, no la versión observada → backport cross-versión | Elegir la versión que empata la instalación; backport exige alineación de versión base o rebase explícito. ⛔ E6 |
| C-N-11 | **Rename de `id`** entre versiones → update-check busca el id viejo → aviso ciego | Tabla de alias `id_previo→id_actual` consultada por update-check. F4 |
| C-N-12 | **checkout detectado como instalación** (dogfood: el proyecto ES el checkout) | path == checkout conocido → clasificar CANÓNICO, jamás instalación (cierra deriva auto-referencial). ⛔ RN-IDENT-4 |
| C-N-13 | **Monorepo**: profundidad de walk / `.claude/plugins/<id>/.claude-plugin/plugin.json` por plugin | Descender + `LoadArnes` por dir; profundidad acotada + cancelable; no seguir symlinks fuera del root. ⛔ C-P-11/12 |
| C-N-14 | **Plugin de marketplace normal** (con `plugin.json`, SIN `arnes.l0.json`) | Fallback `plugin.json` (id=name/version); hoy da `Arnes=nil` y ni el `id` sale — **bloqueante de cimientos**. ⛔ D-DOM-4 |
| C-N-15 | **`empresa` conocida pero origen desconocido** (mockup metió «Vitalia» del path) | `empresa` sale de la MISMA cadena §6 (manifiesto); si no resuelve → empresa **desconocida**, jamás del path. ⛔ BR-3/BR-5 |
| C-N-16 | Instalación-espejo que **ES un clone git con commits locales** (eslabón 3) | Sigue read-only en la app; sus commits locales cuentan como `deriva`, no como autoría. INV-1 |
| C-N-17 | TOCTOU: instalación mutada por el cliente **entre** diagnóstico y reparar | Re-evaluar deriva justo antes del overwrite; re-confirmar si cambió. RN-REP |
