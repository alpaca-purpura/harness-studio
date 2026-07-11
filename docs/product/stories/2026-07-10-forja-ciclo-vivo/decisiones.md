# Decisiones — Forja · Ciclo de vida del arnés VIVO (Fase 1, 2026-07-10)

> Paquete de EJECUCIÓN de la Fase 1. El **modelo** (D0-D20 + `arnes.yaml`) ya está FIRMADO 🧑‍⚖️ en
> `../2026-07-10-terreno-conocimiento/`; acá se decide **CÓMO se ejecuta** (no se re-litiga el modelo).
> Cada decisión conversada se escribe EN EL MISMO TURNO (disciplina §10). Nada firmado aún.

## F-D0 · Slice de arranque = dimensión piloto `forma-trabajo` end-to-end — DECIDIDA
- El slice más fino que ya es una **dimensión real**: se construye UNA (no las 11) end-to-end como
  ejemplo vivo, y se convierte en el **fixture** que el forjador (paso 2) debe reproducir determinísticamente.
- **Elegida `forma-trabajo`** (territorio Organización, `activacion: always`) porque:
  1. Tiene NORMA durable REAL ya en el repo → `docs/architecture/conventions/` (10 hojas: commits · go-style ·
     ts-style · naming · ci · git-hooks · editor · CADENCE · ts-types) + PARIDAD (metodologia §10/§3). No se fabrica.
  2. **NO posee capabilities** (Producto = Σ capabilities; Organización no aporta caps) → **cero colisión R1-R4**.
     La reconciliación 11-dims ↔ 82-caps (P6) es sobre todo un problema de `producto`; se difiere a Slice 1b.
  3. Demuestra las 3 caras (D18): NORMA (la dimensión) · PASO (`forma-trabajo` toca el proceso build/paridad) ·
     ARTEFACTO (queda en WIP) — sin arrastrar la migración pesada de `docs/architecture/`.
- Territorios `proposito` y `producto` se dejan como INDEX **stub** (shape visible) hasta su slice.

## F-D1 · Layout del scaffold (mapping `arnes.yaml` → `docs/terreno/`) — DECIDIDA
Derivación determinista `arnes.yaml` → árbol de disco:
```
docs/terreno/
  INDEX.md                                  # raíz: 4 territorios (de `territorios:`), links, leyenda salud
  <territorio>/INDEX.md                      # por territorio: dimensiones que le pertenecen (de `dimensiones[].territorio`)
  <territorio>/<dimension>/INDEX.md          # L1: resumen (de receta) + link a la hoja + a knowledge/
  <territorio>/<dimension>/<dimension>.md    # L2: la hoja atómica (NORMA), ≤~400 tok, punteros_auto a código real
  <territorio>/<dimension>/knowledge/INDEX.md# D5/D9: rules/skills co-locadas que nacen de las decisiones de la dim
docs/wip/
  INDEX.md                                   # snapshot vivo AUTO-GEN (DO-NOT-EDIT) desde gestion_trabajo
  activo/  done/                             # carpetas del lifecycle (D15: estado + cierre → done/)
```
- **3 caras (D18) en disco:** NORMA→`docs/terreno/<t>/<dim>/` · PASO→plantillas de `proceso/` (paso 2) ·
  ARTEFACTO→`docs/wip/activo/<outcome>/<paquete>/`.
- La hoja atómica sigue el **schema D9** (frontmatter `id/tipo/dimension/resumen/activacion/globs/punteros_auto/budget_tok`).
- `receta.plantilla` del `arnes.yaml` = la PLANTILLA VACÍA que el forjador rellena. En Slice 1a la hoja se
  **autora a mano** (golden output); la plantilla-fuente y su ubicación (kit) se formalizan en el paso 2.

## F-D2 · Orden golden-primero: fixture ANTES del forjador — DECIDIDA
- Slice 1a autora el scaffold **a mano** (golden). El forjador del paso 2 debe **reproducirlo byte-relevante**
  (PARIDAD determinista). Sin golden no hay contra-qué verificar el generador. Robado de projen/snapshot-testing.
- Regla: el golden vive en `docs/terreno/`+`docs/wip/` del propio repo (dogfood) y es a la vez producto y fixture.

## F-D3 · Migración `docs/architecture/` → `terreno/producto/` DIFERIDA a Slice 1b — DECIDIDA
- Es la parte más pesada (P6: reconciliar con 82 capabilities sin romper R1-R4; `arch/` es fuente de checks
  del conformance). Mover ahora arriesga romper el ruleset. Se hace en su propio slice, con red de PARIDAD.
- Slice 1a NO toca `docs/architecture/`; solo lo **referencia** desde la hoja `forma-trabajo` (punteros vivos).

## F-D4 · `docs/terreno/` y `docs/wip/` NO rompen las cifras generadas — DECIDIDA
- `scripts/estado.sh` genera el bloque de cifras desde conformance + árbol de capabilities. El nuevo árbol
  `docs/terreno/` es DEFINICIÓN (capa encima), no capabilities → no debe alterar el conteo de 82. Verificar
  `estado.sh --check` verde tras el scaffold (honestidad automática N2, HS-21).

## F-D5 · El terreno/forja se ENCARNA en el kit ② que ships, NO en `docs/terreno/` top-level — DECIDIDA (revisa firmado)
Aclaración conceptual del user (2026-07-10, esta sesión). Separa 3 niveles:
- **A · el SOFTWARE ArnesIA (motor/fábrica):** `cmd/ internal/ web/`. Sus features = 82 caps. = "nuestro proyecto".
- **B · los ARNESES-producto:** paquetes as-code que el motor LEE/CONDUCE (agnóstico al rubro). Se distribuyen.
- **C · proyecto-DESTINO:** donde se instala un arnés; el terreno se puebla AHÍ en runtime.
- **Los 3 cuerpos (HS-10, YA firmado):** ① doctrina (`kit/doctrine.md` + árboles como DATA embebida) · ② kit
  (`kit/` = skills `forjar-caja`/`auditar-arnes` que hacen a CC **experto forjando arneses**) · ③ el arnés forjado.
  La app **embebe ①② (`embed_doctrina.go`), los materializa a `~/.arnesia/` (provisioner) e inyecta por
  `--plugin-dir` a cada sesión de chat** → el chat in-app actúa como experto. Cuerpo ② JAMÁS se escribe en ③.
- **Corrección:** mi `docs/terreno/`+`docs/wip/` (Slice 1a) = un **③ dogfoodeado sobre el motor (A), mal-ubicado
  top-level**, mezclado con los docs del proyecto. La **ubicación canónica de la expertise de forja es el kit ②
  (embebido)**, no `docs/terreno/`. El OUTPUT de forjar = un ③ en su propio tree, no top-level acá.
- **Decisión operativa:** NO revertir `docs/terreno/`+`docs/wip/` — el user dijo que **migrará** al modelo (queda
  como fixture del Slice 1a; el modelo D0-D20 sigue válido como **schema que el forjador aplica** para producir un ③).
  Pero **no es la ubicación canónica** del arnés-producto.
- **Revisa lo firmado:** terreno-conocimiento («`docs/terreno` absorbe `docs/architecture` in-repo», INDEX L22)
  crea justo la mezcla A/B que el user rechaza → queda **bajo revisión** (nueva `D` + firma cuando toque). Ver **P9** allí.

## F-D6 · Historia PAUSADA — antes del scaffold/forjador va el spike de CARGA — DECIDIDA
- Prioridad del user: **usar la app y mejorar arneses YA** (tiene trabajo retrasado). La app corre
  (`bin/arnesia serve` → `127.0.0.1:4200`, kit inyectado, 3 arneses de ejercicio cargados) pero le falta la
  **agencia**: poder elegir desde la app qué arnés/proyecto cargar. → se abre spike aparte.
- **Esta historia (forja-ciclo-vivo) queda PAUSADA** en Slice 1a (built, pending firma 🧑‍⚖️) hasta que el spike
  de carga entregue su decisión. Paquete del spike: [`../2026-07-10-spike-carga-arneses/`](../2026-07-10-spike-carga-arneses/INDEX.md).
- Retomar esta historia = firmar Slice 1a → paso 2 (forjador+gate) reencuadrado por F-D5 (la forja vive en ②).

## Por profundizar (heredado de terreno-conocimiento P1-P8)
Se cocinan cuando su etapa llegue; no bloquean el slice: P2 `expande` (sub-dims de producto) · P3 `cond` de
spines · P4 `receta.scraping` · P5 knowledge-hoja ↔ `arnes.yaml` · P6 reconciliar 82 caps (Slice 1b) ·
P7 refresco `HANDOFF.md` · P8 overlays carve-out.

## Abiertas (siguiente)
- Autorar el golden de `forma-trabajo` + raíz terreno + esqueleto wip (Slice 1a).
- Firmar Slice 1a → capability de scaffold → paso 2 (forjador + gate).
