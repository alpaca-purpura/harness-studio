# Spec de usabilidad — Portafolio

> `tipo: spec-usabilidad` · complementa `spec-funcional.md` + `casuistica.md`. Superficie firmada = mockup
> `mockups/arnesia-portafolio.html` v2. SSoT del UI = Storybook (al construir se porta a stories). Tokens = DTCG
> reales (`web/tokens/base.tokens.json`). Superset ESTRICTO del baseline; no pisa vocabulario L0. NO firmado.

## 1 · Principios de experiencia
- **Una tarea por superficie.** Lista = encontrar/entrar. Wizard = adquirir. Drawer = entender + actuar sobre UNO.
  Las acciones NO viven en la fila de lista (fue el error de v1).
- **Honestidad visible.** Lo no construido = deshabilitado + tooltip con el slice; los datos no resueltos = «desconocido/
  no evaluable», jamás fabricados. (BR-3, BR-4, BR-8.)
- **Defaults honestos del FE-Slice-1** (S-D9/S-D10 — corrige el mockup v2): sin `home` accesible, **`deriva` = `deriva-no-evaluable`**
  (NO `◐ N archivos`); **update = `no-verificado`** (NO `⬆ vX`, que es Slice 4); la rama **Marketplace del wizard = disabled + tooltip**
  (la validación exitosa hardcodeada viola BR-8); `empresa` **jamás del path** (si no resuelve → desconocida). Fixes G1-G9 = ítems de build de este slice.
- **Terminología:** «drift» → **`deriva`** (`en-deriva` / `al-hilo` / `deriva-no-evaluable`) en toda etiqueta.
- **La UI hace cumplir la ley anti-drift** (§5): no existe superficie de edición sobre una instalación.
- **Pegado al Storybook/tokens** (norma anti-drift UI): nada de estilos inventados.

## 2 · Superficie 1 — LISTA (home del Portafolio)

**Layout:** shell (rail izq, Portafolio activo) + stage = topbar + toolbar + lista scroll + leyenda-pie.
- **Topbar:** título · contadores (`N arneses · M empresas · lente`) · botón `＋ Agregar`.
- **Toolbar:** Filtros (empresa ▾ · marketplace ▾ · estado ▾ · buscar) + lente **«Ver por»** (empresa · proyecto ·
  marketplace · plano) con caption de propósito. **Filtro** acota el set; **lente** re-agrupa. Son distintos.
- **Fila (compacta, clickable):** emblema (inicial) · id (mono) + descripción · **presencia** (`◆ canónico vX` |
  `◇ sin canónico` · `▣ N instalac.`) · **flags** (`⬆ vX` update · `◐ drift` · `origen?`) · dot de salud.
- **Grupo (por lente):** header con inicial+nombre+count. Un arnés base aparece en varios grupos (esperado, N:M).

**Estados:**
- **Vacío (primera vez):** ilustración + «Tu portafolio está vacío» + CTA `＋ Agregar` con las 2 vías explicadas.
- **Cargando escaneo:** skeleton de filas / spinner por grupo.
- **Filtro sin resultados:** «Ningún arnés coincide» + limpiar filtros.
- **Fila con problema:** `origen?` / `drift` / `no-encontrada` como chips, nunca ocultos.

**Interacción:** clic en fila → abre Drawer (superficie 3). Hover = resalte. Teclado: filas focuseables, Enter abre.

## 3 · Superficie 2 — AGREGAR (wizard, modal)

**Paso 0 — elegir vía:** dos tabs-card: **Proyecto** (carpeta/repo con arneses instalados → observar/reparar) ·
**Marketplace** (git url → validar → elegir → trabajar). Cada uno con una línea de propósito.

### Rama Proyecto
- **Paso 1 · Fuente:** radio **Carpeta local** (`Slice 1`) con input path + `Escanear`; radio **Repo GitHub**
  (`Slice 2`, deshabilitado) con input url + `Clonar`.
- **Paso 2 · Escaneo:** lista de arneses detectados con checkbox; por cada uno id + versión + marketplace resuelto
  (o «origen desconocido (sin lock)»). Estados: escaneando (spinner) · 0 encontrados (mensaje C-P-4) · malformado
  (fila «no reconocible») · error de path (C-P-3).
- **Paso 3:** botón `Agregar N al portafolio` + nota «entran como espejos read-only; el canónico se trae aparte».

### Rama Marketplace (`Slice 2`, visible-diferido)
- **Paso 1 · Apuntar:** input git url + `Validar`. Resultado: ✓ «válido · estructura OK · N arneses» | ✗ «no tiene
  estructura de marketplace» (C-M-1) | error red (C-M-2) | privado→auth (C-M-3).
- **Paso 2 · Elegir:** lista de arneses del marketplace con checkbox (id·versión·empresa) → `Agregar seleccionados`
  (clona canónico). Nota de acceso `gh`/PAT (auth-terms).

**Reglas UI del wizard:** cancelar/cerrar sin efectos. Idempotencia visible: si ya está agregado → «ya lo tenés»
(BR-9). Deshabilitados con tooltip de slice.

## 4 · Superficie 3 — DETALLE (drawer)

**Header:** emblema · id · descripción · cerrar.
**Facetas:** `empresas` (chips, puede ser varias — N:M) · `marketplaces` (chips o «desconocido»).
**Callout ley anti-drift:** una línea fija que explica canónico-editable / instalaciones-observación.

**Zona Canónico** (badge «única copia editable»):
- Si **presente:** versión + path + (si update) «⬆ hay vX». Acciones: `◉ Abrir en Mapa` (Slice 1) · `✎ Mejorar (chat)`
  (S2) · `▲ Publicar` (S3).
- Si **ausente:** «No tenés el canónico local — necesario para autorear» + CTA `↧ Traer canónico del marketplace` (S2).
  (INV-4: sin canónico, autoría bloqueada.)

**Zona Instalaciones** (badge «observación · read-only»):
- Por cada instalación: dot salud · proyecto · versión · path · estado de drift (`◐ N archivos` | `sin drift` |
  `drift-no-evaluable` | `origen sin resolver` | `no-encontrada`). Acciones: `◉ Observar en Mapa` (Slice 1) ·
  `⚒ Reparar` (S5) · `↩ Backport al canónico` (S2). **Reparar** muestra confirmación destructiva si hay drift (RN-REP-1).
- Si **0 instalaciones:** «sin instalaciones registradas».

**Footer:** `⊘ Desvincular del portafolio` + nota «solo lo saca de la vista — no desinstala ni borra el clon» +
checkbox secundario «…y borrar clon local» (destructivo, RN-UNL-2). Si canónico sucio → advertencia (RN-UNL-3).

## 5 · Cómo la UI hace cumplir la ley anti-drift (afordancias)
- **No hay botón «Editar» sobre una instalación.** Solo Observar (read) · Reparar (re-materializa) · Backport (→ canónico).
- **Mejorar/chat** solo aparece habilitado en la **zona Canónico**, nunca en Instalaciones.
- **Reparar** es visualmente un flujo de *re-materialización* (con confirmación de sobrescritura), no un editor.
- Sin canónico, la zona Canónico muestra CTA de traer, y Mejorar/Publicar no existen como acción disponible.

## 6 · Accesibilidad + interacción transversal
- Roles ARIA: lista = filas botón; drawer = `complementary`/dialog con foco atrapado + Esc cierra; wizard = dialog modal.
- Teclado: Tab por filas/acciones; Enter/Espacio activan; Esc cierra drawer/wizard.
- Contraste: tokens DTCG (light+dark) ya validados; estados de error usan `--crit`/`--warn` + texto (no solo color).
- Reduced-motion: transiciones desactivadas (media query).
- Copy: español rioplatense, honesto y corto; los tooltips de deshabilitado nombran el **slice** («próximo · S2»).

## 7 · Responsive
- Rail colapsable a gutter (patrón existente). Drawer full-width en < 640px. Grilla de filtros wrapea. Lista siempre
  scroll vertical; nada de scroll horizontal del body (contenedores anchos con overflow propio).

## 8 · Alcance del Slice 1 (usabilidad construida ahora)
Lista + toolbar (lente empresa + filtros visuales) · Wizard rama Proyecto/carpeta-local (pasos 1-3) · Drawer READ
completo (facetas, zonas, anti-drift, Observar, Desvincular). Todo lo demás = deshabilitado + tooltip. Porte a Storybook
(stories nuevas: `portafolio-list`, `portafolio-drawer`, `portafolio-wizard`) como SSoT.
