# Plan de pruebas · Agregación consolidada al Portafolio (proyecto + marketplace)

> `tipo: plan-pruebas` · paquete `2026-07-23-portafolio-agregar-marketplace` · 2026-07-25.
> Complemento ejecutable de [`spec.md`](./spec.md) §6/§7 y [`design.md`](./design.md).
>
> **Ejecutable, no aspiracional.** Cada fila trae: capa · archivo exacto · nombre de test ·
> insumo real concreto · aserción exacta. Los escenarios `E-01..E-30` son los del spec; los
> `E-31..E-59` los agrega este plan por orden del operador («todos los casos habidos y por
> haber») y **cada uno define la RESPUESTA DEL SISTEMA**, que siempre degrada visible: nunca
> inventa, nunca calla.

## 0 · Las 5 capas y qué prueba cada una

| capa | dónde | comando | qué NO prueba |
|---|---|---|---|
| **D · Dominio Go (tabla)** | `internal/domain/marketplace{,_situacion}_test.go` | `go test ./internal/domain/...` | I/O, red, orden de puertos |
| **A · Adaptadores Go** | `internal/adapters/marketplace/*_test.go` | `go test ./internal/adapters/marketplace/...` | política (qué se lee primero) |
| **U · Usecase Go** | `internal/usecase/{marketplace,portafolio}_test.go` | `go test ./internal/usecase/...` | shape del JSON del cable |
| **H · HTTP Go** | `internal/adapters/transport/http/marketplace_test.go` | `go test ./internal/adapters/transport/http/...` | reglas de negocio (viven en D/U) |
| **F · Selectores FE** | `web/src/entities/marketplace/model/selectors.test.ts` | `pnpm --dir web test --project=unit` | render |
| **S · Stories (story = test)** | `*.stories.tsx` con `play()` | `pnpm --dir web test --project=storybook` ⚠ interactivo | transporte real |
| **E · E2E vivo** | manual, contra esta máquina | `curl` + app levantada | nada mockeado |

**Prohibido mockear donde hay dato real disponible** (spec §7.4). Las fixtures de las capas D/A/F
son **copias literales** de los archivos de esta laptop, no invenciones:

```
internal/adapters/marketplace/testdata/
  known_marketplaces.json                     ← copia de ~/.claude/plugins/known_marketplaces.json (5 entradas)
  prenter/.claude-plugin/marketplace.json     ← copia literal (2 entradas, mismo source)
  prenter/catalogo.json                       ← copia literal (canales + 4 versiones, 0.5.1 deprecada)
  prenter/plugins/harness/{0.5.0,0.5.1,0.5.2,0.5.3}/.keep   ← el árbol real conserva las 4
  oficial/.claude-plugin/marketplace.json     ← MUESTRA de claude-plugins-official: 25 filas que cubren
                                                 las 4 formas de source + version null + version semver
                                                 + `renames` + `$schema` + description top-level
  caveman/.claude-plugin/marketplace.json     ← copia literal (source "./", owner con `url`, 1 entrada)
  dañados/…                                   ← los fixtures de daño de E-31..E-49 (§3)
```

> **Por qué una muestra de 25 y no las 273 del oficial:** el archivo real son 159 KB y meterlo al
> repo infla el árbol sin agregar una sola rama de comportamiento nueva. Las 273 se ejercitan en la
> capa **E** (vivo) y en la story de lazy (E-28) con un generador. Las 25 elegidas cubren **todas**
> las formas reales de `source` y de `version`. Esto queda anotado como decisión, no como omisión.

---

## 1 · Los 30 escenarios del spec §6

| id | escenario | capa | archivo | test / story | insumo real | aserción exacta |
|---|---|---|---|---|---|---|
| E-01 | plano nace poblado por CC | A + E | `adapters/marketplace/detector_test.go` | `TestDetectorCCCincoEntradasReales` | `testdata/known_marketplaces.json` (copia literal) | `len(out)==5`; nombres exactos `{claude-plugins-official, caveman, claude-code-warp, ponytail, prenter-marketplace}`; cada uno `Eslabones==[EslabonCCKnown]`, `InstallLocation!=""`, `Repo` canonicalizado a `github.com/<owner>/<repo>`, `CCActualizado` = el `lastUpdated` crudo |
| E-02 | catálogo propio legible offline | A | `adapters/marketplace/catalogo_local_test.go` | `TestLectorLocalPrenterDosEntradas` | `testdata/prenter/` | `Lectura.Tipo==LecturaLeida`, `Lectura.Fuente=="local"`, `len(Entradas)==2`, `Entradas[*].Nombre=={"harness","harness-beta"}`, `OwnerNombre=="Prenter"`, `OwnerEmail=="hola@alpacapurpura.lat"` |
| E-03 | dos entradas, mismo `source` | D | `domain/marketplace_test.go` | `TestAgruparCanalesMismoSource` | las 2 filas reales de prenter (`./plugins/harness/0.5.3`) | `Entradas[0].ComparteSourceCon==["harness-beta"]` ∧ `Entradas[1].ComparteSourceCon==["harness"]`; una fila con source único ⇒ `ComparteSourceCon==nil` |
| E-04 | versión derivada de la ruta | D | `domain/marketplace_test.go` | `TestVersionDeEntradaDerivaDeSource` | `SourceCatalogo{Tipo:RutaRelativa, Ruta:"./plugins/harness/0.5.3"}`, `declarada==""` | `("0.5.3", VersionDeSource)` |
| E-05 | enriquecimiento `catalogo.json` | A | `adapters/marketplace/catalogo_local_test.go` | `TestEnriquecimientoCatalogoJSON` | `testdata/prenter/catalogo.json` real | `Canales=={"estable":"0.5.3","beta":"0.5.3"}`; `len(Versiones)==4`; `Versiones` contiene `{Version:"0.5.1", Estado:"deprecada"}`; la fila `harness` (v0.5.3) ⇒ `EstadoCanal=="habilitada"`. **Corrección del spec:** `0.5.1` es deprecada en `versiones[]`, no en una fila del catálogo (ninguna fila apunta a 0.5.1) |
| E-06 | marketplace sin `catalogo.json` | A | `adapters/marketplace/catalogo_local_test.go` | `TestSinCatalogoJSONDegradaSinRuido` | `testdata/oficial/` (no tiene `catalogo.json`) | `Canales==nil` ∧ `Versiones==nil` ∧ todas las filas `EstadoCanal==""` ∧ **`Lectura.Motivo==""`** (degradado SIN ruido) ∧ `Lectura.Tipo==LecturaLeida` |
| E-07 | clase referencia sin Traer | D + S | `domain/marketplace_situacion_test.go` · `widgets/marketplace/ui/marketplace-catalogo.stories.tsx` | `TestReferenciaNuncaHabilitaAccion` · story `CatalogoReferenciaReadOnly` | las 6 situaciones × `ClaseReferencia`; fixture `mkOficialReferencia` | **Go:** para las 6 ramas, `Habilitada==false` ∧ `Motivo=="no aplica: solo arneses propios"`. **Story:** el botón `↧ Traer canónico` tiene `disabled` ∧ `title=="no aplica: solo arneses propios"`; `queryByRole("button",{name:/Publicar|Actualizar|Reparar/})===null` |
| E-08 | 0 hallazgos honesto | E + S | vivo · `portafolio-wizard.stories.tsx` | `WizardCeroHallazgos` (extender la existente) | escanear `/home/chalreme/Proyectos/vitalia` (tiene `.claude/settings.json` + `.claude/skills/`, **sin** `.claude/plugins/`) | **Vivo:** `POST /api/portafolio/escaneos` devuelve exactamente 1 candidato (`proyecto-instalado` del root) o 0 según `enabledPlugins`; la UI muestra el mensaje honesto completo. **Story:** `getByText(/No encontré arneses instalados aquí/)` presente ∧ `PasoFuente` re-montado (`getByLabelText("Ruta del proyecto")` visible) ∧ **cero** filas de candidato |
| E-09 | hallazgos reales positivos | E | vivo | — | escanear `/home/chalreme/Proyectos/luana-platform` | ≥5 candidatos. **Composición real corregida (C10 de `design.md`):** `enabledPlugins` tiene 6 claves, **4 en `true`** (`frontend-design`, `claude-md-management`, `commit-commands`, `harness@prenter-marketplace`); 3 tienen record para ese `projectPath`, `harness@prenter-marketplace` **no** (su record es de `luana-vitalia`) ⇒ 3 `referenciada-cc` + 1 hallazgo-aviso + 1 `proyecto-instalado`. **Nada premarcado**, botón `Agregar 0 al portafolio` `disabled` |
| E-10 | aviso real de eslabón | E | vivo | — | escanear `/home/chalreme/Proyectos/harness-studio` | aparece un candidato con `instalacion.aviso` que contiene `enabledPlugins declara "harness@prenter-marketplace" pero sin record de instalación para /home/chalreme/Proyectos/harness-studio`; se pinta con `AvisoChip` **completo, sin truncar** |
| E-11 | `Atrás` desde candidatos | S | `portafolio-wizard.stories.tsx` | `WizardAtrasDesdeCandidatos` | 4 candidatos fixture, `pathEscaneado="/home/chalreme/Proyectos/luana-platform"`, 2 tildados | click en `← Atrás` ⇒ `onAtras` llamado 1 vez; con `estado="fuente"` el input conserva el valor exacto de la ruta ∧ el set de elegidos sobrevive (al volver a `candidatos` los 2 siguen tildados) |
| E-12 | `cambiar ruta` | S | idem | `WizardCambiarRuta` | idem | click en `cambiar ruta` ⇒ **el mismo `onAtras`**, mismo resultado que E-11 (una transición, dos puertas) |
| E-13 | Cancelar = cero efectos | S | idem | `WizardCancelarCeroEfectos` | idem, 2 tildados | click en `Cancelar` ⇒ `onClose` llamado ∧ `onAgregar` **nunca** llamado ∧ `onEscanear` **nunca** llamado |
| E-14 | cierre bloqueado en `agregando` | S | idem | `WizardCierreBloqueadoAgregando` (extender) | `estado="agregando"` | `✕` tiene `disabled` ∧ `title=="agregando en curso — esperá a que termine"`; `Atrás` y `Cancelar` también `disabled`; `userEvent.keyboard("{Escape}")` ⇒ `onClose` **no** llamado |
| E-15 | validar url real | E + A | vivo · `adapters/marketplace/catalogo_remoto_test.go` | `TestLectorRemotoGHShimPrenter` | vivo: `https://github.com/alpacapurpura/prenter-marketplace` con `gh` autenticado como `alpacapurpura`. Test: shim `gh` (script en `testdata/bin/gh` que emite el base64 del fixture) vía `GHBin` | **Vivo:** 200 con `{nombre:"prenter-marketplace", owner_nombre:"Prenter", entradas:2, fuente:"local"}` (usa el checkout: ese repo YA es conocido por CC). Forzando `install_location:""`: `fuente:"remoto"`, mismos 3 datos. **Test:** idem con el shim, y el `gh` real nunca se invoca |
| E-16 | validar url inexistente | U + H | `usecase/marketplace_test.go` · `transport/http/marketplace_test.go` | `TestValidarURLInexistente` · `TestPostValidacionesRepoInexistente400` | `https://github.com/alpacapurpura/no-existe-xyz`; lector fake que devuelve `ErrNoEsMarketplace` con stderr `HTTP 404` | `errors.Is(err, ErrNoEsMarketplace)`; HTTP **400** ∧ body `{"error":...}` que contiene `404`; **body NO contiene `"valido"`** (no hay forma de pintar ✓, BR-5) |
| E-17 | validar repo sin `marketplace.json` | U + H | idem | `TestValidarRepoSinMarketplaceJSON` · `TestPostValidacionesNoEsMarketplace400` | `https://github.com/alpacapurpura/vitalia` (existe, no es marketplace) | 400 ∧ el motivo contiene literalmente `no expone .claude-plugin/marketplace.json legible` — **distinguible** del 404 de E-16 |
| E-18 | registrar duplicado | U + H + S | `usecase/marketplace_test.go` · `transport/http/marketplace_test.go` · `portafolio-wizard.stories.tsx` | `TestRegistrarDuplicadoNoPisaNiDuplica` · `TestPostMarketplaces409ConNombre` · `WizardRegistrarDuplicado` | store con `prenter-marketplace` ya declarado, clase `propio`; registrar de nuevo con clase `referencia` | `errors.Is(err, ErrMarketplaceYaRegistrado)`; **el store queda intacto** (`Listar()` sigue con clase `propio`, 1 sola fila); HTTP **409** con `{"error":…,"nombre":"prenter-marketplace"}`; **Story:** aparece el copy «ya está registrado» + botón `ir a él` ⇒ `onIrAMarketplace("prenter-marketplace")` |
| E-19 | registrar y aterrizar | S + E | `portafolio-wizard.stories.tsx` · vivo | `WizardRegistrarYAterriza` | `mkEstado="validado"`, `mkValidacion` poblada | click en `Registrar y ver catálogo` ⇒ `onRegistrarMarketplace(url,"propio")` con los valores exactos; **Vivo:** el modal cierra, `plano==="marketplaces"`, el catálogo de la fila nueva queda abierto (`#/portafolio?plano=marketplaces`) |
| E-20 | situación `mi-copia-adelantada` | D + S | `domain/marketplace_situacion_test.go` · `marketplace-catalogo.stories.tsx` | `TestSituacionMiCopiaAdelantada` · story `CatalogoLas6Situaciones` | canónico `0.5.4` vs estante `0.5.3`, todas las instalaciones `al-hilo` | `{Tipo:mi-copia-adelantada, Mia:"0.5.4", Estante:"0.5.3"}` ∧ `Accion=={publicar,false,"Publicar se construye en su propio paquete (ítem 3 del outcome)"}`. **Story:** botón `Publicar` con `disabled` ∧ ese `title` exacto (**cierra C5**: el mockup no lo tenía) |
| E-21 | situación `estante-adelantado` | D + S | idem | `TestSituacionEstanteAdelantado` · idem | canónico `0.5.2` vs estante `0.5.3` | `{estante-adelantado, "0.5.2","0.5.3"}` ∧ `Accion=={actualizar-mi-copia,false,"…(ítem 4 del outcome)"}` |
| E-22 | situación `instalaciones-en-deriva` | D + S | idem | `TestSituacionEnDerivaGanaAVersion` · idem | 2 instalaciones, 1 `en-deriva`; **y a la vez** canónico `0.5.2` vs estante `0.5.3` | `{instalaciones-en-deriva, Cuantas:1}` — **la deriva gana a la divergencia de versión** (precedencia §6.1) ∧ `Accion=={reparar,false,"…(ítem 5 del outcome)"}` |
| E-23 | situación `no-comparable` | D + S | idem | `TestSituacionNoComparableSinVersionEstante` · idem | fila con `Version==""` (p.ej. `source` objeto `git-subdir` + `version:null`) y entrada presente | `{Tipo:no-comparable, Motivo:"el catálogo no declara versión de esta entrada ni se puede derivar de su source"}` ∧ `Accion.Verbo==""`. **Story:** `DotSaludPortafolio` con clase `sin-senal` (**nunca** la de `ok`) + el motivo textual completo visible |
| E-24 | discrepancia entre eslabones | D + S | `domain/marketplace_test.go` · `marketplace-list.stories.tsx` | `TestMergeMarketplacesAnotaDiscrepancia` · story `MarketplaceConDiscrepancia` | detectado `{nombre:"x", repo:"github.com/a/b"}` + declarado `{nombre:"x", repo:"github.com/c/d"}` | 1 sola fila; `Repo=="github.com/c/d"` (declarado manda); `len(Discrepancias)==1` con **los dos** valores en el texto; `Eslabones==[cc-known-marketplaces, declarado-por-operador]`. **Story:** las dos señales visibles, sin botón de «elegir una» |
| E-25 | reconciliar con confirmación | U + S | `usecase/portafolio_test.go` · `resolver-origen-dialog.stories.tsx` | `TestAsignarOrigenReKeyeaYAnotaEslabon` · story `ResolverOrigenConfirma` | entrada provisional `sin-home~legal-administrativo~…`; `home="github.com/vitalia/arneses"` | **Go:** la clave nueva es `github.com-vitalia-arneses~legal-administrativo~`; la vieja ya no está en `Listar()`; cada instalación tiene el eslabón `{Fuente:"declarado-por-operador",Campo:"home",Valor:"github.com/vitalia/arneses"}`; **cero archivos escritos fuera del store** (fixture con el dir de instalación en modo `0555` ⇒ el test pasa igual). **Story:** `Confirmar origen` nace `disabled` con `title=="elegí un destino primero"`; al elegir una opción se habilita |
| E-26 | reconciliar «ninguno» | U + S | idem | `TestAsignarOrigenNingunoNoInventaHome` · story `ResolverOrigenNinguno` | `home=""` / `{"sin_origen":true}` | **la identidad NO cambia** (sigue `sin-home~…`) ∧ `OrigenSinResolverDesde` != `""` (RFC3339) ∧ `SinOrigenResuelto()` bajó en 1. **Story:** la opción «ninguno — dejarlo sin origen» es la **última**, con la señal `honesto, no se inventa un home`, y habilita `Confirmar` |
| E-27 | toolbar sin ambigüedad | S | `portafolio-list.stories.tsx` | `ToolbarRotulosVisibles` | `entradasDemo` | `getByText("Ver por")` ∧ `getByText("Filtros")` **visibles en el DOM** (no solo `aria-label`); los `role="group"` con sus `aria-label` existentes **siguen** presentes (no se degrada a11y); las dos «Marketplace» quedan dentro de grupos rotulados distintos |
| E-28 | lazy loading | S | `shared/ui/lista-lazy.stories.tsx` · `marketplace-catalogo.stories.tsx` | `ListaLazyNoTruncaEnSilencio` · `CatalogoDoscientasSetentaYTres` | 273 filas generadas (la cifra real del oficial, AG-D16) | montaje inicial ≤ 20 filas; el pie dice `… 253 más (se cargan al bajar)` — **el número exacto de faltantes, nunca un truncado mudo**; click en el botón de fallback monta el paso siguiente; tras N pasos, `getAllByRole("listitem").length===273` |
| E-29 | sin red, catálogo cacheado | U | `usecase/marketplace_test.go` | `TestCatalogoSinRedUsaCacheConCuando` | caché con `leido_en` de hace 2 días; `local` falla (`installLocation` ausente) y `remoto` devuelve error de red | `len(cat.Entradas)>0` (las del caché) ∧ `cat.Lectura.Tipo==LecturaSinAcceso` ∧ `cat.Lectura.Cuando` == el del caché (2 días atrás) ∧ `cat.Lectura.Motivo` contiene el error de red real. **Nunca** `Entradas==nil` habiendo caché, **nunca** `Tipo==leido` sin lectura fresca |
| E-30 | a11y de modales | S | `portafolio-wizard.stories.tsx` · `resolver-origen-dialog.stories.tsx` | `WizardMarketplaceA11y` · `ResolverOrigenA11y` | — | `role="dialog"` ∧ `aria-modal="true"` ∧ `aria-labelledby` apuntando al `<h2>`; `Tab` cicla dentro del diálogo (`trapTabKeyDown`); `Escape` cierra (salvo estado bloqueado); el gate a11y de `.storybook/preview.ts` está en `error` ⇒ **cualquier violación axe rompe el test sin aserción extra** |

---

## 2 · Escenarios agregados — y la respuesta del sistema

Todos definen **qué hace el sistema**, y esa respuesta es honesta por construcción: degrada
visible, jamás inventa ni calla.

### 2.1 · Integridad del `marketplace.json` ajeno

| id | caso | capa · archivo · test | insumo | **respuesta del sistema** (la aserción) |
|---|---|---|---|---|
| E-31 | `marketplace.json` **sin** la clave `plugins` | A · `catalogo_local_test.go` · `TestSinClavePlugins` | `testdata/dañados/sin-plugins.json` (`{name,owner}` y nada más) | `Entradas==nil` (⇒ `null` en el cable) ∧ `Lectura.Tipo==LecturaLeida` ∧ `Lectura.Motivo=="marketplace.json sin la clave plugins"`. **Es «leí y el archivo está incompleto», no «no tiene arneses»** |
| E-32 | `plugins: []` de verdad | A · idem · `TestPluginsVacioEsAfirmacion` | `testdata/dañados/plugins-vacio.json` | `Entradas==[]EntradaCatalogo{}` (⇒ `[]`, **NO** `null`) ∧ `Lectura.Tipo==LecturaLeida` ∧ `Lectura.Entradas==0` ∧ `Motivo==""`. La UI dice «leído hace X · no declara ningún arnés» — afirmación evidenciada |
| E-33 | JSON sintácticamente corrupto | A · idem · `TestJSONCorruptoNoCrashea` | archivo truncado a la mitad | `err != nil` mapeado a `Lectura.Tipo==LecturaURLNoResuelve` ∧ `Motivo` contiene el offset del error de parseo ∧ `Entradas==nil`. **El proceso no paniquea** (`recover` innecesario: parseo tolerante) |
| E-34 | `source` apunta a una ruta inexistente | A · idem · `TestSourceRutaAusenteAvisaSinBorrarFila` | `testdata/prenter/` con `plugins/harness/0.5.3/` **borrado** | la fila **se muestra igual** ∧ `Aviso==["el catálogo apunta a ./plugins/harness/0.5.3 pero ese dir no existe en el checkout"]` ∧ `Version=="0.5.3"` (la derivación no depende de que el dir exista) ∧ `Lectura.Tipo==LecturaLeida`. En lectura **remota** no se verifica (§5.2.6: N requests sería hacer de proxy) |
| E-35 | `installLocation` ya no existe en disco | A + U · `catalogo_local_test.go` + `marketplace_test.go` · `TestInstallLocationAusenteCaeARemoto` | fila con `InstallLocation` a un dir borrado | el local devuelve `ErrNoEsMarketplace` con motivo `installLocation ya no existe en disco: <ruta>` ∧ el usecase **cae al remoto**; sin remoto ⇒ `Lectura.Tipo==LecturaSinAcceso` con ese motivo. **Nunca** un `[]` |
| E-36 | permisos: `marketplace.json` en modo `0000` | A · idem · `TestSinPermisoDeLectura` | `os.Chmod(0o000)` en el fixture (skip si el test corre como root) | `Lectura.Tipo==LecturaSinAcceso` ∧ `Motivo` contiene `permission denied` y la ruta ∧ `Entradas==nil` |
| E-37 | dos filas con el **mismo `name`** | A + D · `catalogo_local_test.go` + `marketplace_situacion_test.go` · `TestNombreDuplicadoAmbasVisibles` | `testdata/dañados/nombre-duplicado.json` | **las dos filas se conservan** ∧ las dos con `Aviso==["nombre duplicado en el catálogo: harness"]` ∧ las dos con `Situacion.Tipo==no-comparable`. El sistema **no elige una** ni dedupea el error ajeno |
| E-38 | el catálogo renombró un plugin (`renames`) | D · `marketplace_test.go` · `TestRenamesCruzaPorNombreAnterior` | `renames: {"convex-backend":"convex"}` (entrada REAL del oficial) + una entrada del Portafolio con `identidad.ID=="convex-backend"` | la fila `convex` tiene `NombreAnterior==["convex-backend"]` ∧ `CruzarConPortafolio` la encuentra con `Via=="rename"` ∧ la UI muestra `ViaChip` `renombrado: convex-backend → convex`. **Sin `renames` la misma entrada daría `no-lo-tengo`** — el test asserta las dos ramas |
| E-39 | forma de `source` no reconocida | A + D · `parse` + situación · `TestSourceDesconocidoEsVisible` | `"source": 42` (ni string ni objeto) | `Source.Tipo==SourceDesconocido` ∧ `Source.Crudo=="42"` (crudo recortado, visible) ∧ `Aviso` con el motivo ∧ `Situacion.Tipo==no-comparable` con motivo `el catálogo declara un source que no reconozco: 42`. **La fila no se descarta** |
| E-40 | fila sin `name` | A · `catalogo_local_test.go` · `TestFilaSinNombreSeDescartaVisible` | `plugins: [{"source":"./x"}]` | la fila **se descarta** (sin `name` no hay nada instalable que nombrar) ∧ `Truncado==1` ∧ `Lectura.Motivo` contiene `1 fila(s) sin name`. **Descarte VISIBLE**, no silencioso |
| E-41 | `owner` con `url` en vez de `email` | A · idem · `TestOwnerConURL` | `testdata/caveman/` real (`{"name":"Julius Brussee","url":"https://github.com/JuliusBrussee"}`) | `OwnerNombre=="Julius Brussee"` ∧ `OwnerURL=="https://github.com/JuliusBrussee"` ∧ `OwnerEmail==""`. **Ningún campo inventado** (AG-D15) |
| E-42 | `description` top-level vs `metadata.description` | A · idem · `TestDescripcionTopLevelGana` | oficial (top-level) y prenter (`metadata`) | oficial ⇒ la top-level; prenter ⇒ la de `metadata`; un archivo con **las dos** ⇒ gana la top-level; sin ninguna ⇒ `""` |
| E-43 | versión no-semver en el catálogo | D · `marketplace_test.go` + `marketplace_situacion_test.go` · `TestVersionNoSemverEsNoComparable` | `version:"latest"` · `ref:"v1.5.5"` · `ruta:"./plugins/x/main"` | `VersionDeEntrada("latest", …)` ⇒ `("latest", VersionDeCampo)` (**el dato se conserva**, no se descarta) ∧ al comparar contra un canónico semver ⇒ `no-comparable` con motivo `versiones no comparables (no-semver): «0.5.2» vs «latest»`. **Jamás una comparación de strings** |

### 2.2 · Persistencia y corrupción

| id | caso | capa · archivo · test | insumo | **respuesta del sistema** |
|---|---|---|---|---|
| E-44 | `~/.arnesia/marketplaces.json` con envelope ilegible | A · `store_test.go` · `TestStoreEnvelopeIlegibleDegradaHonesto` | archivo con JSON roto | `NewStore` **no falla** ∧ `Listar()` ⇒ `(nil, [1]EntradaCorrupta{Motivo:"envelope ilegible: …", Raw: todo})` ∧ el plano sigue **poblado por el detector CC** + banner de corrupta |
| E-45 | una fila del registro ilegible | A · idem · `TestStoreFilaCorruptaNoContagia` | 3 filas, la 2ª con `clase` de tipo equivocado | 2 sanas + 1 corrupta; un `Upsert` posterior **re-serializa la corrupta cruda** junto a las sanas (misma decisión y misma limitación honesta que `portafolio.Store.saveLocked`) |
| E-46 | fila del registro sin `nombre` | A · idem · `TestStoreFilaSinNombreEsCorrupta` | `{"repo":"github.com/a/b","clase":"propio"}` | `EntradaCorrupta{Motivo:"fila sin nombre: no hay clave de merge"}`. **No se le inventa un nombre** derivándolo del repo |
| E-47 | clase desconocida en el registro | D + A · `marketplace_test.go` · `TestClaseDesconocidaDegradaAReferencia` | `"clase":"tienda"` | `ClaseSegura("tienda")==ClaseReferencia` ∧ discrepancia visible `clase desconocida "tienda": se trata como de referencia`. **Fail-safe hacia el lado que NO habilita operar** (boundary §11.3) |
| E-48 | caché de catálogo corrupto | A · `cache_test.go` · `TestCacheCorruptoNoEsCeroEntradas` | `~/.arnesia/catalogos/<slug>.json` con JSON roto | `Leer` ⇒ `(_, "caché de catálogo ilegible: …", false)` ∧ la fila del plano dice `no-leido` **con el motivo visible** ∧ el próximo refresco lo sobreescribe sin intervención |
| E-49 | caché de versión desconocida | A · idem · `TestCacheVersionDesconocidaSeDescarta` | `{"version":99,…}` | `ok=false` con motivo `caché de versión 99 desconocida (se re-leerá)` ∧ **se descarta y se reconstruye** — acá sí es legítimo: el dato ES derivable (a diferencia del registro, donde la clase declarada no lo es) |
| E-50 | `Guardar` de un catálogo sin lectura | A · idem · `TestCacheRechazaEntradasNil` | `Catalogo{Entradas:nil}` | `errors.Is(err, ErrCacheSinEntradas)` ∧ **no se escribe ningún archivo**. Cachear un «no sé» convertiría un fallo transitorio en un estado persistente |
| E-51 | no se puede escribir el caché (dir `0555` / disco lleno) | U · `marketplace_test.go` · `TestCatalogoSeDevuelveAunqueNoSePuedaCachear` | `Cache` fake que falla en `Guardar` | el catálogo **se devuelve completo** ∧ `Lectura.Motivo` gana el sufijo `(no se pudo cachear: …)`. **La lectura no se pierde por no poder guardarla** |
| E-52 | nombre con `@`, espacios o unicode | A · `cache_test.go` · `TestCacheSlugSeguroYSinColision` | nombres `mi mkt`, `mi-mkt`, `ñandú@v2`, `../escape` | el archivo cae **siempre dentro** de `~/.arnesia/catalogos/` (ningún `..` sobrevive a `domain.Slug`); dos nombres que colapsan al mismo slug ⇒ el segundo usa `<slug>~<huella>.json` ∧ **ninguno pisa al otro** ∧ el `nombre` se muestra **crudo** en la UI (el slug es de disco, no de presentación) |
| E-53 | `symlink` en `installLocation` | A · `catalogo_local_test.go` · `TestInstallLocationSymlink` | symlink válido → dir con catálogo; y symlink **roto** | válido ⇒ lee normal (`EvalSymlinks` primero); roto ⇒ igual que E-35 (`ErrNoEsMarketplace` con motivo, cae a remoto). **Nunca se sigue un symlink a un dir que no existe** |

### 2.3 · Red, `gh` y credenciales

| id | caso | capa · archivo · test | insumo | **respuesta del sistema** |
|---|---|---|---|---|
| E-54 | `gh` ausente o no autenticado, sin PAT | A + H · `catalogo_remoto_test.go` · `TestSinViaDeLectura503` | `GHBin="/no/existe"`, `Token=""` | `errors.Is(err, ErrSinViaDeLectura)`; en `POST /validaciones` ⇒ **503** (no 400) con motivo `sin vía de lectura (ni checkout local ni gh/PAT)`. **503 ≠ 400: el sistema dice «no puedo mirar», nunca «tu url está mal»** |
| E-55 | selector de credencial | A · idem · `TestElegirCredencial` | tabla: `(gh sí/no) × (auth sí/no) × (pat ""/"x")` | `gh+auth` ⇒ `CredencialGH`; `gh` sin auth + PAT ⇒ `CredencialPAT`; sin gh + PAT ⇒ `CredencialPAT`; nada ⇒ `CredencialNinguna` + motivo. **Función pura: se testea sin ningún login real** (spec §8) |
| E-56 | timeout / red que corta a mitad | A · idem · `TestTimeoutEsSinAccesoConMotivo` | shim `gh` que duerme 30 s; `Timeout=200ms` | `context.DeadlineExceeded` mapeado a `Lectura.Tipo==LecturaSinAcceso` ∧ `Motivo` contiene `timeout` y los segundos ∧ **el caché viejo se conserva intacto** (`Guardar` no se llama) ∧ el daemon sigue respondiendo (`/healthz` 200 durante el test) |
| E-57 | respuesta gigante (> techo de bytes) | A · idem · `TestTechoDeBytes` | shim que emite 9 MiB; `MaxBytes=8<<20` | `Lectura.Tipo==LecturaURLNoResuelve` ∧ `Motivo` contiene `excede el techo de 8 MiB` ∧ **el proceso no crece sin control** (`io.LimitReader`, no `ReadAll` a pelo) |
| E-58 | catálogo con más filas que el techo | A · idem · `TestTechoDeEntradasEsVisible` | 6 000 filas generadas; techo 5 000 | `len(Entradas)==5000` ∧ `Truncado==1000` ∧ el FE **muestra** `1000 entradas no cargadas (techo de seguridad)`. **Recorte visible, jamás silencioso** |
| E-59 | host que no es `github.com` | A + U · idem · `TestHostNoGitHub` | `https://gitlab.com/a/b` | `ErrSinViaDeLectura` con motivo `solo se sabe leer catálogos de github.com por ahora` ⇒ 503. **Límite declarado, no un fallo genérico** |
| E-60 | `gh` devuelve 401/403 (repo privado sin permiso) | A · idem · `TestGH401EsSinAcceso` | shim con exit 1 + stderr `HTTP 403` | `Lectura.Tipo==LecturaSinAcceso` (no `url-no-resuelve`) ∧ el **stderr real** viaja como `Motivo` — distinguible de E-16 (404 ⇒ `ErrNoEsMarketplace` ⇒ 400) |

### 2.4 · Concurrencia, caché stale e identidad

| id | caso | capa · archivo · test | insumo | **respuesta del sistema** |
|---|---|---|---|---|
| E-61 | dos lecturas simultáneas del mismo catálogo | A · `cache_test.go` · `TestCacheDosEscriturasConcurrentes` | 2 goroutines × `Guardar` + 2 × `Leer`, `-race` | ninguna respuesta partida (temp+rename es atómico: se ve el viejo o el nuevo, nunca mitad y mitad) ∧ el archivo final parsea ∧ **`go test -race` limpio**. El último gana: `singleflight` queda como deuda con razón (`design.md §4.3`) |
| E-62 | dos registros simultáneos del mismo nombre | U · `marketplace_test.go` · `TestRegistrarConcurrenteNoDuplica` | 2 goroutines × `Registrar` misma url | exactamente 1 éxito ∧ 1 `ErrMarketplaceYaRegistrado` ∧ `Listar()` con **1 sola fila** (el `Upsert` es bajo mutex del store) |
| E-63 | caché más viejo que el remoto | U · idem · `TestCacheStaleNoAfirmaEstarAlDia` | caché de hace 2 días con `harness@0.5.2`; remoto ya en `0.5.3`; `refrescar=false` | se devuelve **el caché** ∧ `Lectura.Cuando` = hace 2 días (**la UI dice «leído hace 2 días», no «al día»**) ∧ con `refrescar=true` la respuesta pasa a `0.5.3` y `Cuando` a ahora. El sistema **nunca afirma frescura que no tiene** (BR-3) |
| E-64 | Portafolio vacío + catálogo de 273 | D · `marketplace_situacion_test.go` · `TestPortafolioVacioTodoNoLoTengo` | 0 entradas persistidas | **las 273** filas ⇒ `no-lo-tengo`; **cero** `al-hilo`. El `al-hilo` requiere una entrada real: no es el default |
| E-65 | N entradas del Portafolio cruzan con la misma fila | D · idem · `TestNCoincidenciasEsNoComparable` | 2 entradas provisionales, mismo `id`, ambas con el registry del marketplace | `Situacion.Tipo==no-comparable` ∧ `Motivo` cita **las 2 claves** ∧ **no se elige ninguna** (C-ID-2). La salida es reconciliar (S7) |
| E-66 | cruce débil por faceta `registry` | D · idem · `TestCruceFacetaRegistry` | entrada `sin-home~harness~…` con `Registries==["github.com/alpacapurpura/prenter-marketplace"]` (el caso REAL dominante) | **encuentra** la entrada ∧ `Via=="faceta-registry"` ∧ la UI pinta `ViaChip` `origen no declarado — cruzado por el registry de la copia` ∧ la fila **sigue contando** en `sin_origen_resuelto`. Sin esta vía, toda la columna diría `no-lo-tengo` (§6.2) |
| E-67 | dos marketplaces distintos con el mismo `repo` | D · `marketplace_test.go` · `TestMismoRepoNombresDistintosNoMergean` | `{nombre:"a", repo:"github.com/x/y"}` + `{nombre:"b", repo:"github.com/x/y"}` | **2 filas** (la clave de merge es el `nombre`, el que CC usa para keyear `installed_plugins.json`) ∧ cada una con la discrepancia informativa `otro marketplace conocido apunta al mismo repo: <otro>`. No se fusionan por coincidencia |
| E-68 | `known_marketplaces.json` ilegible | U + H · `marketplace_test.go` · `TestDetectorIlegibleNoOcultaNiAborta` | archivo corrupto | `GET /api/marketplaces` ⇒ **200** con `aviso_detector` poblado ∧ `marketplaces` = solo los declarados. **Nunca un 500, nunca una lista vacía muda** |
| E-69 | colisión al asignar origen | U + H + S · `portafolio_test.go` · `TestAsignarOrigenColisionaNoFusiona` | ya existe `github.com-vitalia-arneses~legal-administrativo~`; se asigna el mismo home a otra provisional con el mismo id | `errors.Is(err, ErrAsignarOrigenColisiona)` ∧ **las dos entradas siguen existiendo intactas** ∧ HTTP **409**. **Story:** el copy dice «ya existe una entrada con esa identidad — fusionar es otra operación» |
| E-70 | asignar origen habilita la deriva | U + E · `portafolio_test.go` + vivo · `TestAsignarOrigenReevaluaDeriva` | instalación con `deriva-no-evaluable` («sin referencia local accesible») cuyo `home` asignado SÍ tiene checkout con esa versión | tras `AsignarOrigen`, la instalación pasa a `al-hilo` o `en-deriva` **por hash real** (nunca por el string de versión) ∧ el cambio se persiste. **Es el efecto útil de reconciliar, y aparece sin un botón nuevo** |
| E-71 | `Olvidar` un marketplace que CC sigue conociendo | U + H · `marketplace_test.go` · `TestOlvidarSigueDetectado` | `prenter-marketplace` declarado + detectado; `DELETE` | 200 `{olvidado:true, sigue_detectado:true}` ∧ `Listar()` **lo sigue mostrando** con `Eslabones==[cc-known-marketplaces]` y `Clase==referencia`. **No podemos hacer que Claude Code deje de conocerlo, y lo decimos** |
| E-72 | `catalogo.json` presente pero corrupto | A · `catalogo_local_test.go` · `TestCatalogoJSONCorruptoAvisaSinRomper` | `testdata/prenter/catalogo.json` truncado | `Entradas` completas ∧ `Canales==nil` ∧ `Versiones==nil` ∧ `Lectura.Motivo` con el sufijo `(catalogo.json ignorado: …)`. **Degradar sin ruido ≠ ocultar un archivo roto que el operador puso a propósito** |
| E-73 | `catalogo.json` de OTRO marketplace | A · idem · `TestCatalogoJSONDeOtroMarketplaceSeIgnora` | `catalogo.json` con `"marketplace":"otro"` | se ignora ∧ mismo aviso que E-72. El enriquecimiento no se aplica a ciegas |
| E-74 | `plugins-vacio` + caché previo con 2 entradas | U · `marketplace_test.go` · `TestVacioRealPisaElCache` | caché con 2; lectura fresca devuelve `[]` legítimo | el resultado es `[]` con `Entradas:0` y `Cuando` de ahora — **el vacío REAL pisa el caché** (es una lectura exitosa). Distinto de E-29, donde no hubo lectura |
| E-75 | corrupta del registro + detector OK, a la vez | U · idem · `TestCorruptaYDetectorConviven` | registro con 1 fila corrupta + 5 detectadas | 200 con **5** `marketplaces` ∧ **1** `corruptas` ∧ ninguna de las dos oculta a la otra |


### 2.5 · `↧ Traer canónico` (S8) — E-76..E-90 afinados

Fixtures nuevas en `internal/adapters/traer/testdata/`: `mkt-prenter/` (copia del checkout real, con
`plugins/harness/0.5.3/` poblado, un **symlink interno**, un **symlink que escapa**, un archivo
ejecutable y un `.git/` con contenido) · `bin/git` (shim que emite salidas reales capturadas) ·
`bin/gh` (idem). Los tests de red van con `-tags red` y **no corren en CI** (opt-in), salvo E-81b.

| id | escenario | capa · archivo · test | insumo real | aserción exacta |
|---|---|---|---|---|
| E-76 | camino A feliz | A+U · `adapters/traer/local_test.go` · `TestCopiadorLocalPrenter` · `usecase/traer_test.go` · `TestTraerCaminoALocal` | `testdata/mkt-prenter`, fila `harness`, `source ./plugins/harness/0.5.3` | destino `<raiz>/checkouts/prenter-marketplace/harness` con el árbol completo; `camino=="local"`; `sha_efectivo==""`; entrada registrada con `Canonico.Path==destino` y `Canonico.Version=="0.5.3"`; **cero llamadas de red** (el fake de red falla el test si se invoca) |
| E-77 | deriva tras traer | U · `usecase/traer_test.go` · `TestTraerEvaluaDerivaDeInmediato` | E-76 + `Referencias` apuntando al mismo checkout | `res.Deriva=="al-hilo"` **y persistida** en la entrada; con la referencia ausente ⇒ `deriva-no-evaluable` **con motivo** y **200 igual** (BR-17: se muestra lo que salga) |
| E-78 | canal comparte `source` | U · idem · `TestTraerCanalesDestinoPropio` | traer `harness` y después `harness-beta` (mismo `source`) | dos destinos distintos (`…/harness` y `…/harness-beta`), **ninguno pisa al otro**, dos entradas con `Identidad.ID` distinto y el **mismo** `Home` |
| E-79 | destino ya poblado | U+H · idem + `transport/http/marketplace_test.go` · `TestTraerDestinoPobladoAborta` · `TestPostTraidos409ConDestino` | repetir E-76 | `errors.Is(err, ErrTraerDestinoPoblado)`; **el contenido previo es byte-idéntico** (hash del árbol antes/después); **cero temporales creados** (`<raiz>/tmp` vacío); HTTP **409** con `{"error":…,"destino":…}` |
| E-79b | destino existe pero VACÍO | U · idem · `TestTraerDestinoVacioSePuedeUsar` | `mkdir -p` del destino, sin archivos | **no** es "poblado": el Traer procede. `os.ReadDir` (no `os.Stat`) es la comprobación correcta |
| E-79c | destino es un ARCHIVO | U · idem · `TestTraerDestinoEsArchivo` | tocar un archivo en la ruta del destino | tratado como poblado ⇒ 409, sin tocarlo |
| E-80 | clase `referencia` | D+H · `domain/traer_test.go` · `TestPlanificarTraerRechazaReferencia` · `TestPostTraidosReferencia400` | `frontend-design@claude-plugins-official` | `errors.Is(err, ErrTraerClaseReferencia)` con el literal `no aplica: solo arneses propios`; HTTP **400**; **cero I/O** (el fs fake asserta cero syscalls de escritura) |
| E-81 | camino B feliz (shim) | A · `adapters/traer/externo_test.go` · `TestClonadorExternoShim` | shim `git` que reproduce la salida real capturada de `42Crunch-AI/claude-plugins` | los comandos invocados son, en orden: `init` · `remote add` · `sparse-checkout init --cone` · `sparse-checkout set plugins/api-security-testing` · `fetch --depth 1 --filter=blob:none origin 30287f5e…` · `checkout FETCH_HEAD` · `rev-parse HEAD`; el pin es **el `sha`, NO el `ref`** (C16); staging queda con el contenido de la subruta y **sin `.git`** |
| E-81b | camino B feliz (RED REAL, opt-in) | A · idem · `TestClonadorExternoRedReal` (`-tags red`) | `42Crunch-AI/claude-plugins`, sha `30287f5e…`, path `plugins/api-security-testing` | `HEAD == 30287f5e3f122a646d1ac5ca3ab96e130c52a3ad` exacto; el staging trae `skills/`+`references/` y **no** el `README.md` de la raíz (§13.1 hecho 7); tamaño < 5 MiB. **Verificado a mano el 2026-07-25 antes de escribir el diseño** |
| E-82 | `sha` que no coincide | A+U · idem · `TestTraerSHANoCoincideAborta` | shim cuyo `rev-parse` devuelve otro sha | `errors.Is(err, ErrTraerSHANoCoincide)`; **`<raiz>/tmp` vacío** tras el error; **destino inexistente**; **cero entradas** en el store; HTTP **502** |
| E-83 | `ref`/`sha` inexistente | A+U · idem · `TestTraerRemotoNoTiene` | shim con exit 128 + stderr `couldn't find remote ref` | `errors.Is(err, ErrTraerRemotoNoTiene)`; el **stderr real** viaja en el motivo; HTTP **502**; nada registrado, temporal limpio |
| E-84 | sin `gh` ni PAT | A+U+H · idem · `TestTraerSinAuth503` | `GHBin="/no/existe"`, `Token=""`, shim con stderr `could not read Username` | `errors.Is(err, ErrTraerSinAuth)`; HTTP **503**; el motivo **no** insinúa que el repo no exista; **ArnesIA nunca escribe ni pide una credencial** (aserción: `<raiz>` no contiene ningún archivo con el token, y el token nunca aparece en el `argv` capturado por el shim) |
| E-85 | 401/403 vs 404 | A+H · idem · `TestTraerClasificaAuthVsNoExiste` | dos shims: stderr `403` y stderr `Repository not found` | **códigos distintos** (503 vs 502) **y motivos distintos**, ninguno genérico. Tabla de clasificación de §13.7 cubierta entera (4 filas) |
| E-86 | interrupción a mitad | U · `usecase/traer_test.go` · `TestTraerCancelacionNoDejaResiduo` | `Materializador` fake que bloquea; `ctx` cancelado a mitad | `ctx.Err()` propagado; **`<raiz>/tmp` vacío** (el `defer` corrió); **destino inexistente**; **cero entradas**. Segundo caso: fake que paniquea ⇒ el test asserta que el `defer` limpió igual |
| E-86b | huérfano de un SIGKILL previo | U · idem · `TestBarridoDeTemporalesAlArrancar` | `<raiz>/tmp/traer-viejo` con mtime de hace 2 h + `traer-nuevo` de hace 1 min | al construir el servicio: el viejo se borra, **el reciente NO** (podría ser de otra instancia viva) |
| E-87 | destino que escapa por `..` | D · `domain/traer_test.go` · `TestRutaCanonicoNoEscapa` | tabla: `"../../.claude"`, `"..%2F.."`, `"/etc"`, `"a/../../b"`, `"…"` (unicode puro), `""`, `"mi mkt"`, `"ñandú@v2"` | **todos** los resultados caen dentro de `RaizCheckouts`; `Slug` colapsa los separadores y los `..`; el caso unicode-puro y el vacío usan `HuellaPath` en vez de un segmento vacío; **dos nombres distintos nunca dan el mismo destino** |
| E-88 | jamás escribe en `~/.claude` | U · `usecase/traer_test.go` · `TestTraerJamasEscribeEnClaude` | `HOME` falso con `.claude/` poblado (3 archivos + mtimes registrados); se corren **las 6 ramas** de §13.5 | tras cada rama: el árbol `<HOME>/.claude` es **idéntico byte-a-byte y mtime-a-mtime**; **todo** path escrito (capturado por un `fs` instrumentado) cae bajo la raíz inyectada; un `Materializador` fake que intenta escribir en `<HOME>/.claude` hace **fallar** la operación sin dejar rastro. **Aserción de ruta y bytes, no de confianza** |
| E-89 | disco lleno / sin permiso | U · idem · `TestTraerFalloLocalLimpia` | destino padre en modo `0o555` (skip si root); y un `Rename` fake que devuelve `ENOSPC` | `errors.Is(err, ErrTraerLocal)` con el motivo real; **temporal limpio**; **nada registrado**; HTTP **500** |
| E-90 | `source` no reconocido | D · `domain/traer_test.go` · `TestPlanificarTraerSourceNoMaterializable` | `SourceDesconocido` con crudo `42`; objeto sin `sha` **ni** `ref`; ruta relativa **sin** `installLocation` en disco | los tres ⇒ `ErrTraerSourceNoMaterializable` con el **crudo visible** en el motivo; **no se intenta adivinar**; el 3º explica que una ruta relativa no dice a qué repo pertenece |
| **E-91** | `source: "./"` (raíz del marketplace) | A · `adapters/traer/local_test.go` · `TestCopiaRaizDeMarketplaceExcluyeGit` | `testdata/mkt-caveman/` con `source "./"` y un `.git/` poblado | se copia el árbol **sin `.git`**; `plan.RaizDeMarketplace==true`; `Avisos` trae el literal de §13.2. **Verificado que el caso existe** (caveman, ponytail lo usan) |
| **E-92** | symlink interno vs symlink que escapa | A · `adapters/traer/copia_test.go` · `TestCopiaSymlinks` | árbol con `dentro -> ./sub/x`, `fuera -> /etc/hostname`, `roto -> ./nada` | `dentro` se copia **como symlink relativo**; `fuera` **no se copia** y deja `Aviso` con la ruta y el target; `roto` **no se copia** y deja `Aviso`; **`/etc/hostname` no aparece en el staging** |
| **E-93** | permisos y bit de ejecución | A · idem · `TestCopiaPreservaEjecutable` | un `hooks/pre.sh` en `0o755` y un `README.md` en `0o644`; un archivo con setuid | dirs `0o750`, archivos `0o640`; el `.sh` conserva `0o111`; el **setuid/setgid/sticky se descarta** |
| **E-94** | el set de exclusión es EL de `HashFormaPlugin` | A · idem · `TestExclusionCoincideConHashFormaPlugin` | árbol con `.git/`, `.in_use`, `.orphaned_at` y un `.gitignore` | los 3 primeros **no** se copian; `.gitignore` **sí** (no está en el set). Aserción cruzada: `HashFormaPlugin(origen) == HashFormaPlugin(staging)` — **es la garantía de que BR-17 dé `al-hilo`** |
| **E-95** | `commit` ≠ `sha` en `source: github` | D · `domain/traer_test.go` · `TestGitHubDosHashesGanaSha` | la entrada real de `fullstorydev/fullstory-skills` (`commit 1ec5865e…`, `sha b20614e2…`) | el pin es **`sha`** (`b20614e2…`); `Avisos` trae el literal que cita **los dos** hashes; **no se elige en silencio** (C18) |
| **E-96** | `ref` que no coincide con `sha` | D · idem · `TestRefDivergenteEsAvisoNoAborto` | la entrada real de `42Crunch-AI` (`ref v1.5.5` → `faf53053…` vs `sha 30287f5e…`) | el pin es el `sha`; **`Avisos`** trae el literal de divergencia; **NO** aborta (C16 — abortar rechazaría una entrada sana) |
| **E-97** | fetch por sha rechazado por el remoto | A · `adapters/traer/externo_test.go` · `TestFetchPorShaFallaReintentaConRef` | shim: 1er `fetch <sha>` con stderr `not our ref`; 2º `fetch --branch <ref>` OK, `rev-parse` = el sha | **un solo** reintento, con `--branch <ref>`; BR-16 **sigue aplicando** (si el `rev-parse` no diera el sha ⇒ 502). Sin `ref` ⇒ aborta con el stderr real |
| **E-98** | techo de tamaño del clone | A · idem · `TestTechoDeBytesEnClone` | shim que materializa 70 MiB; `MaxBytes=64<<20` | aborta con motivo `excede el techo de 64 MiB`; **temporal limpio**; nada registrado |
| **E-99** | timeout del clone | A+U · idem · `TestTimeoutDeClone` | shim que duerme 30 s; `Timeout=200ms` | `ErrTraerRemotoNoTiene` con motivo de timeout; **temporal limpio**; el daemon sigue respondiendo (`/healthz` 200 durante el test) |
| **E-100** | el `Upsert` falla tras materializar | U · `usecase/traer_test.go` · `TestUpsertFallidoDejaDirVisible` | store fake que falla en `Upsert` | HTTP **500** con **el destino en el motivo**; el dir **existe** (no se borra: costó red) ; **cero entradas**; y el Traer siguiente da **409 destino-poblado** — el parcial es VISIBLE, no un silencio |
| **E-101** | `Rename` cross-device | U · idem · `TestStagingEnMismoFilesystemQueDestino` | raíz inyectada; se asserta que el temporal se crea **bajo** `<raiz>/tmp`, no en `os.TempDir()` | el path del temporal tiene el prefijo `<raiz>`; test de regresión del riesgo `EXDEV` (§12.2 #11) |
| **E-102** | dos Traer concurrentes del mismo destino | U · idem · `TestTraerConcurrenteMismoDestino` (`-race`) | 2 goroutines, mismo `nombre`+`entrada` | **exactamente 1** éxito y 1 `ErrTraerDestinoPoblado`; **una sola** entrada en el store; el árbol del destino parsea y está completo (no mezclado) |
| **E-103** | dos Traer concurrentes de canales distintos | U · idem · `TestTraerConcurrenteCanalesDistintos` (`-race`) | `harness` y `harness-beta` a la vez | los **dos** tienen éxito, dos destinos, dos entradas; `-race` limpio |
| **E-104** | el catálogo no está cacheado | U · idem · `TestTraerSinCacheLeePrimero` | store sin caché de catálogo | el Traer **lee el catálogo primero** (una lectura) y sigue; si la lectura falla ⇒ el error de lectura, **no** un «entrada no encontrada» engañoso |
| **E-105** | traer algo que ya está en el Portafolio como instalación | U · idem · `TestTraerConInstalacionesPrevias` | entrada `(home,id)` ya registrada con 2 instalaciones y **sin** canónico | el `Upsert` **agrega** el `Canonico` y **conserva** las 2 instalaciones (`mergeInstalaciones` + `unionDedup`, S1-D3); la situación pasa de `no-lo-tengo`/`no-comparable` a comparable |
| **E-106** | el canónico traído cruza con su fila | D+U · `domain/marketplace_situacion_test.go` · `TestCanonicoTraidoCruzaPorHome` | tras E-76, recalcular la situación de la fila `harness` | la fila pasa a `al-hilo` con `via=="home-declarado"` — **la prueba de que C17 se resolvió bien**: con `Home` = nombre del marketplace daría `no-lo-tengo` sobre algo recién traído |
| **E-107** | a11y y estado de la acción en el FE | S · `widgets/marketplace/ui/marketplace-catalogo.stories.tsx` · `CatalogoTrayendo` · `CatalogoTraerFallo` · `portafolio-drawer.stories.tsx` · `DrawerTraerHabilitado` | fixtures con `accion.habilitada=true`, y estados `trayendo`/`falló`/409 | `trayendo` ⇒ botón `disabled` + `aria-busy`, la fila no-interactiva; `falló` ⇒ motivo textual + `Reintentar`; 409 ⇒ además `abrir el canónico que ya tenés`; el botón del drawer ya **no** tiene `TOOLTIP_S2`; gate axe en `error` |

**Comandos E2E vivos de S8** (capa E, contra esta máquina):

```bash
# E-76/E-77 · camino A real (sin red)
curl -s -X POST localhost:4200/api/marketplaces/prenter-marketplace/traidos \
  -H 'Content-Type: application/json' -d '{"entrada":"harness"}' | python3 -m json.tool
ls -a ~/.arnesia/checkouts/prenter-marketplace/harness/ | head   # y NO debe haber .git
# E-79 · repetir ⇒ 409 sin tocar nada
curl -s -o /tmp/r -w '%{http_code}\n' -X POST localhost:4200/api/marketplaces/prenter-marketplace/traidos \
  -H 'Content-Type: application/json' -d '{"entrada":"harness"}'; cat /tmp/r
# E-80 · clase referencia ⇒ 400 con el literal
curl -s -o /tmp/r -w '%{http_code}\n' -X POST localhost:4200/api/marketplaces/claude-plugins-official/traidos \
  -H 'Content-Type: application/json' -d '{"entrada":"frontend-design"}'; cat /tmp/r
# E-88 · aserción de ruta: NADA cambió en ~/.claude
find ~/.claude -newermt '-2 minutes' -type f | head   # debe salir VACÍO
# E-86b · cero residuo
ls -la ~/.arnesia/tmp/ 2>/dev/null || echo "sin temporales ✓"
```

---

## 3 · Comandos exactos de verificación y su gate

### 3.1 · Backend Go

```bash
# unidad + tabla + adaptadores + usecase + HTTP (todo el árbol Go)
go test ./...
# carrera (obligatorio por E-61/E-62)
go test -race ./internal/adapters/marketplace/... ./internal/usecase/...
# grafo de imports: el componente `marketplace` nuevo tiene que estar declarado
go run github.com/fe3dback/go-arch-lint@latest check --project-path . \
  --arch-file docs/architecture/fitness/.go-arch-lint.yml
# estilo
golangci-lint run
```

**Gate:** `go test ./...` verde · `go-arch-lint` sin violaciones · `golangci-lint` sin issues nuevos.

### 3.2 · Frontend

```bash
# typecheck + biome + depcruise + steiger + stylelint (los 5 gates estáticos)
pnpm --dir web run verify
# selectores puros (Node, corre en background sin problema)
pnpm --dir web test --project=unit
# story = test (Chromium/Playwright) — ⚠ NO EN BACKGROUND
pnpm --dir web test --project=storybook
```

> ⚠ **Gotcha conocido y vigente:** `vitest-browser` **no corre en background** (Chromium no
> headless en este entorno). El proyecto `storybook` se corre **en sesión interactiva**; en
> background solo `verify` y `--project=unit`. No «arreglar» esto agregando `--browser.headless`
> a ciegas: la config ya declara `headless: true` y el bloqueo es del entorno, no del flag.

**Gate:** `verify` verde (los 5) · `--project=unit` verde · `--project=storybook` verde con el gate
a11y en `error` (cualquier violación axe **es** un fallo, no un warning).

### 3.3 · Arquitectura as-code y capabilities

```bash
# R1 (archivo + símbolo) · R2 (cobertura) · R4 (estado consistente + puntero estable)
go test ./docs/architecture/fitness/...
# motor de fitness sobre arch/ + knowledge/
go run ./cmd/arnesia conformance --todo
# doctor rápido local de capabilities
python3 scripts/cap_doctor.py
# cifras del checkpoint: el drift-gate de CI
bash scripts/estado.sh --check
```

**Gate:**

- `go test ./docs/architecture/fitness/...` **verde**. Los dos modos de romperlo en este paquete:
  (a) un `pointers:` que apunta a un archivo/símbolo que todavía no existe (**R1**), (b) un archivo
  fuente nuevo sin ningún capability que lo reclame (**R2**). Los dos se evitan con la misma regla:
  **código y capability en el MISMO commit** (`design.md §11.1`).
- `conformance --todo`: **el criterio es «cero FAIL nuevo», NO «el total no cambió»**. Este paquete
  agrega 9 checks declarados (4 a `portafolio-identidad-y-deriva-honesta` v1.2, 4 al boundary nuevo,
  1 a `fe-taxonomia-componentes` v1.2) ⇒ total y `deferred` **suben a propósito**. Un `fail` nuevo
  sí es regresión.
- `estado.sh --check`: si el checkpoint queda stale, **rompe CI**. Hay hook `pre-commit` que
  regenera las cifras y hay que `git add` lo regenerado (gotcha lefthook 1.13.6: `stage_fixed` no
  stagea cross-file). **Ninguna cifra se teclea a mano en ningún `.md`.**

### 3.4 · E2E vivo (capa E) — cómo se reproduce

```bash
go run ./cmd/arnesia serve                    # :4200
pnpm --dir web dev                            # :5173
# o, mejor, contra el binario instalado (memoria: validar SIEMPRE contra el .deb instalado)
```

```bash
# E-01
curl -s localhost:4200/api/marketplaces | python3 -m json.tool
# E-02/E-05
curl -s localhost:4200/api/marketplaces/prenter-marketplace/catalogo | python3 -m json.tool
# E-06 + AG-D16 (273 filas, catalogo.json ausente)
curl -s localhost:4200/api/marketplaces/claude-plugins-official/catalogo \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print(len(d["entradas"]), d["lectura"], d.get("canales"))'
# E-15 / E-16 / E-17 — los tres mensajes tienen que ser DISTINGUIBLES
for u in alpacapurpura/prenter-marketplace alpacapurpura/no-existe-xyz alpacapurpura/vitalia; do
  echo "--- $u"; curl -s -o /tmp/r -w '%{http_code}\n' -X POST localhost:4200/api/marketplaces/validaciones \
    -H 'Content-Type: application/json' -d "{\"url\":\"https://github.com/$u\"}"; cat /tmp/r; echo
done
# E-29 (sin red): cortar la red y pedir el catálogo ya cacheado — tiene que responder con `leído hace`
# E-08 / E-09 / E-10
for p in /home/chalreme/Proyectos/vitalia /home/chalreme/Proyectos/luana-platform /home/chalreme/Proyectos/harness-studio; do
  echo "--- $p"; curl -s -X POST localhost:4200/api/portafolio/escaneos \
    -H 'Content-Type: application/json' -d "{\"path\":\"$p\"}" \
    | python3 -c 'import json,sys; c=json.load(sys.stdin); print(len(c),"candidatos"); [print(" ",x["clave"],"|",x["instalacion"].get("aviso","")) for x in c]'
done
```

**Gate del E2E:** los 30 + 45 escenarios recorridos con su resultado anotado en `PARIDAD.md`, cada
uno con evidencia (salida de `curl` o captura). **Un escenario sin verificar se anota `sin-check`,
jamás `pass`.**

### 3.5 · Orden de corrida recomendado (falla rápido y barato)

```
1. go test ./internal/domain/...                       ← la tabla: si esto falla, nada más importa
2. go test ./internal/adapters/marketplace/... ./internal/adapters/traer/...
3. go test -race ./internal/adapters/{marketplace,traer}/... ./internal/usecase/...
4. go test ./...  +  go-arch-lint  +  golangci-lint
5. pnpm --dir web run verify                            ← los 5 gates estáticos (background OK)
6. pnpm --dir web test --project=unit                   ← selectores (background OK)
7. pnpm --dir web test --project=storybook              ← ⚠ interactivo
8. go test ./docs/architecture/fitness/...  +  conformance --todo  +  estado.sh --check
9. E2E vivo (§3.4) → PARIDAD.md
```

---

## 4 · Matriz de cobertura: cada regla de negocio tiene ≥1 prueba

| BR | regla | probada por |
|---|---|---|
| BR-1 | `referencia` jamás habilita `Traer` | E-07 (D + S), enforcer del boundary nuevo |
| BR-2 | `version` derivada, nunca inventada (con AG-D14: precedencia) | E-04, E-23, E-43 |
| BR-3 | catálogo cacheado, la fila siempre dice **cuándo** | E-29, E-63 |
| BR-4 | inalcanzable ⇒ motivo textual, **nunca** lista vacía | E-31, E-32, E-33, E-35, E-36, E-54, E-68 |
| BR-5 | `Validar` no pinta ✓ sin archivo real leído | E-15, E-16, E-17, E-54 (y la máquina de estados §9.2: sin `validado` no existe `Registrar`) |
| BR-6 | los hallazgos no se premarcan | E-09, E-13 |
| BR-7 | registrar existente no pisa ni duplica | E-18, E-62 |
| BR-8 | discrepancias se muestran, no se resuelven solas | E-24, E-37, E-65, E-67 |
| BR-9 | `no-comparable` es rama de primera clase con motivo | E-23, E-37, E-39, E-43, E-65 |
| BR-10 | Publicar/Actualizar/Reparar `disabled` + tooltip | E-20, E-21, E-22 (los 3 asertan el `title` literal) |
| BR-11 | reconciliar escribe `home`; no clona, no instala | E-25, E-26, E-69, E-70 |
| BR-12 | toda superficie nueva es superset; nada firmado se quita | las 2 stories existentes del disclosure (`FiltroEstadoAcota`/`FiltroMarketplaceAcota`) **deben pasar sin tocarlas** tras la promoción a `shared/ui`; ninguna prop del contrato de los widgets se quita (§9.1); las props nuevas del drawer (`onTraerCanonico`/`trayendo`/`traerError`) son **opcionales** ⇒ las stories firmadas del Slice 1 pasan sin tocarlas |
| BR-13 | destino en `~/.arnesia/checkouts/…`; nada fuera de `~/.arnesia`, jamás en `~/.claude` | **E-87** (tabla de escapes, función pura) · **E-88** (aserción de ruta y bytes sobre las 6 ramas) · E-101 (el temporal también cae bajo la raíz) |
| BR-14 | destino poblado ⇒ aborta sin tocar nada | **E-79** (contenido previo byte-idéntico + cero temporales) · E-79b (dir vacío SÍ se usa) · E-79c (archivo cuenta como poblado) · E-102 (concurrencia: 1 éxito, 1 conflicto) |
| BR-15 | materialización atómica; ni canónico parcial ni entrada registrada | **E-86** (cancelación y pánico) · E-86b (barrido de huérfanos al arrancar) · E-82/E-83/E-89/E-98/E-99 (cada rama deja `tmp` vacío y cero entradas) · **E-100** (el único parcial posible, y es visible) · E-101 (mismo filesystem ⇒ el `Rename` es atómico de verdad) |
| BR-16 | `sha` declarado que no coincide ⇒ aborta | **E-82** (aborta, 502, nada registrado) · E-96 (`ref` divergente **no** aborta: el `sha` es la autoridad) · E-97 (el chequeo sigue aplicando en el reintento por `ref`) · E-95 (con dos hashes, gana `sha`) |
| BR-17 | tras traer, `deriva` se evalúa de inmediato y se muestra tal cual salga | **E-77** (`al-hilo` persistido; y `no-evaluable` con motivo ⇒ **200 igual**) · **E-94** (`HashFormaPlugin(origen)==HashFormaPlugin(staging)`: la garantía estructural de que dé `al-hilo`) |
| BR-18 | auth `gh` → PAT, user-owned; ArnesIA no pide, guarda ni proxya credenciales | **E-84** (503 + el token no aparece en ningún archivo de `<raiz>` ni en el `argv` capturado) · E-85 (auth vs no-existe distinguibles) · E-55 (selector de credencial puro, sin login real) |

## 5 · Lo que este plan NO cubre (deuda honesta, declarada)

| hueco | por qué | dónde queda |
|---|---|---|
| las 273 filas reales del oficial como fixture del repo | 159 KB de árbol sin una rama de comportamiento nueva; se cubre con muestra de 25 (todas las formas) + generador en la story + capa E vivo | decisión de este plan, §0 |
| camino PAT contra un login real | auth-terms: la app es conductor, **nunca** proxya un login. Se cubre con el test de unidad del selector de credencial (E-55) | `spec.md §8`, ya declarado |
| `singleflight` en refrescos concurrentes | último-gana no corrompe (E-61); el ahorro no justifica la maquinaria hoy | deuda con razón → `PARIDAD.md` |
| regresión visual pixel (Chromatic) | el repo hace fitness visual **local** (`fe-visual-fitness`), sin servicio externo | ya declarado en el boundary |
| ejecutar Publicar / Actualizar / Reparar | fuera de alcance (`spec.md §0`); acá solo se prueba que quedan `disabled` con su motivo literal (E-20/E-21/E-22) | ítems 3-5 del outcome, `BACKLOG.md` |
| `Traer` contra un repo **privado** real | exigiría un repo privado de prueba y un login; el auth-terms firmado dice que ArnesIA no proxya credenciales. Se cubre con E-84/E-85 (shims con el stderr real de `git`) + E-55 (selector de credencial puro) | `spec.md §8`, ya declarado |
| camino B contra los 220 `source` objeto | se cubren las **3 formas** (`git-subdir`/`url`/`github`) con fixtures reales + 1 test de red real opt-in (E-81b); correr 220 clones en CI es absurdo | decisión de este plan, §2.5 |
| `singleflight` en Traer concurrente | no hace falta: el chequeo de destino-poblado (BR-14) **es** el candado natural, y E-102 lo prueba (1 éxito / 1 conflicto, `-race` limpio) | — |
| deuda a11y `.text-warn` (contraste 4.5:1) | abierta desde antes; este paquete **no la agrava** (§1 del spec: texto nuevo sobre `--warn-soft` usa `--foreground`) | `BACKLOG.md` |
