# Semilla `.arnesia/` — process-as-code en el proyecto del usuario

> **Estado: v0 (2026-07-30) — nace de las decisiones FIRMADAS 🧑‍⚖️ A-D1..A-D4** (paquete
> [`stories/2026-07-30-arnesia-en-el-proyecto/`](../../product/stories/2026-07-30-arnesia-en-el-proyecto/decisiones.md),
> firmadas vía aprobación del plan de sesión del MVP). Contrato de reconocimiento hermano de
> [`nomenclatura-arnes.md`](./nomenclatura-arnes.md): define QUÉ siembra `arnesia init` en un
> proyecto, de dónde se deriva cada archivo y cómo se juzga la salud de la instalación.
> Enforcement = tests Go del sembrador — **`pendiente-de-construir`** (llegan con A-T2/A-T3/A-T5,
> ver §7; declararlos verdes hoy sería pass fabricado).
>
> Norte: [`kit/doctrine.md`](../../../kit/doctrine.md) §Frontera de cuerpos (§9) · modelo de
> terreno FIRMADO D0-D20
> ([`stories/2026-07-10-terreno-conocimiento/decisiones.md`](../../product/stories/2026-07-10-terreno-conocimiento/decisiones.md)) ·
> fuente de derivación = [`semilla/arnes.yaml`](../../../semilla/arnes.yaml) (graduación del v0
> firmado).

## 1. Qué es `.arnesia/` (y qué NO es)

`.arnesia/` es la **zona de escritura sancionada del PROCESO** dentro del proyecto del
usuario: el process-as-code que ArnesIA siembra — stubs de terreno, product (backlog ·
roadmap · índice de paquetes), esqueleto de WIP y plantillas de proceso **por tipo de
paquete** — derivado determinísticamente del modelo de terreno firmado (D0-D20).

- **Excepción EXPLÍCITA a la frontera de cuerpos ②/③.** La regla dura de
  [`kit/doctrine.md`](../../../kit/doctrine.md) §Frontera de cuerpos (§9) prohíbe que la
  maquinaria escriba archivos del kit/doctrina dentro del arnés (② no contamina ③). La
  **enmienda A-D3** (aditiva, runtime intacto) sanciona `<proyecto>/.arnesia/` como LA
  excepción: es process-as-code DEL PROYECTO, no doctrina copiada. **Sigue PROHIBIDO** copiar
  kit, doctrina o knowhow dentro del arnés.
- **Runtime intacto:** el guardrail `ProtegerPaqueteCerrado` matchea `~/.arnesia` (el paquete
  propio de la app), NO `<proyecto>/.arnesia` — se FIJA con test de `refiereAlguno`
  (`pendiente-de-construir`, A-T5).
- **Aditivo puro:** `arnes.l0.json` **SE QUEDA en la raíz del proyecto**. Moverlo dentro de
  `.arnesia/` rompería el contrato firmado [`nomenclatura-arnes.md`](./nomenclatura-arnes.md)
  (D-a: manifiesto en la raíz), el loader y el scanner. `.arnesia/` no reubica NADA existente.
- **NO es** el arnés (③ se reconoce por `nomenclatura-arnes.md`), NO es `~/.arnesia` (el
  paquete cerrado de la app instalada), NO es un cache desechable: es contenido del proyecto,
  versionable por el usuario.

## 2. El árbol v0 (exacto — A-D1)

Cero dimensiones inventadas; todo deriva del modelo firmado (D14 territorios · D19 canónicas ·
D20 multi-pipeline).

```
<proyecto>/
├── arnes.l0.json                    ← NO lo toca la siembra (se queda donde está)
└── .arnesia/
    ├── semilla.lock.json            ← baseline D8 mínimo (§3); lo escribe el sembrador AL FINAL
    ├── terreno/
    │   ├── INDEX.md                 ← mapa: 4 territorios + 11 dimensiones pendientes
    │   ├── proposito/INDEX.md       ← stub de territorio (definición · Customer)
    │   ├── producto/INDEX.md        ← stub de territorio (definición · Solution)
    │   └── organizacion/INDEX.md    ← stub de territorio (definición · Endeavour)
    ├── product/
    │   ├── backlog.md
    │   ├── roadmap.md
    │   └── stories/INDEX.md
    ├── wip/
    │   ├── INDEX.md                 ← foto viva AUTO-GENERADA (DO-NOT-EDIT)
    │   ├── activo/                  ← vacío (`.gitkeep` para sobrevivir a git)
    │   └── done/                    ← vacío (`.gitkeep`)
    └── proceso/
        ├── historia/                ← spine `historia` (6 pasos, D20)
        │   ├── 00-research.md
        │   ├── 01-spec.md
        │   ├── 02-ux.md             ← cond: tiene_ui
        │   ├── 03-arch.md           ← cond: toca_arquitectura
        │   ├── 04-build.md
        │   └── 05-paridad.md
        └── spike/                   ← spine `spike` (2 pasos)
            ├── 00-investigar.md
            └── 01-decidir.md
```

- **WIP es el 4º territorio (D15) pero vive en `.arnesia/wip/`, no bajo `terreno/`:** es
  naturaleza INSTANCIA (trabajo vivo), no definición — `terreno/` solo aloja los 3 territorios
  de definición.
- **`proceso/` guarda las plantillas VACÍAS** (cara-PASO, D18); el artefacto LLENADO vive
  dentro del paquete en `wip/activo/` (cara-ARTEFACTO, D20). Las plantillas sembradas no se
  editan para llenarse: se COPIAN al paquete.

## 3. `semilla.lock.json` — schema 0 (baseline D8 mínimo)

Lo escribe el sembrador **al final** de la siembra (si la siembra falla a medias, no hay lock
que mienta «completa»). Es el sello mínimo de D8: fuente + versión pineada + hash por archivo.

```json
{
  "schema": 0,
  "arnes_id": "dev",
  "version_pineada": 0,
  "fecha": "2026-07-30",
  "archivos": { "<ruta relativa a .arnesia/>": "<sha256 del contenido sembrado>" }
}
```

| Campo | Semántica |
|---|---|
| `schema` | versión del schema del lock (0 = este contrato) |
| `arnes_id` | `arnes.id` del `arnes.yaml` fuente de la siembra |
| `version_pineada` | `schema` del `arnes.yaml` fuente (versión canónica pineada, D8) |
| `fecha` | fecha de siembra `AAAA-MM-DD` |
| `archivos` | mapa ruta→sha256 del contenido TAL COMO SE SEMBRÓ (baseline por archivo) |

**GAP declarado (Fase 2):** la **evaluación de deriva de semilla** (comparar working-tree
contra el baseline, clasificar `original`/`modificado-usuario`/… al estilo D8) NO existe en
v0. El lock se ESCRIBE pero no se evalúa; el doctor v0 juzga por presencia, sin hashes (§4).

## 4. Reglas de siembra y salud

1. **Idempotente — lo existente JAMÁS se pisa.** Archivo ya presente en el destino → se
   respeta byte a byte y se reporta `ya-existia`; solo lo ausente se crea (informe
   `Creados`/`YaExistian`). Re-correr `arnesia init` sobre una instalación sana ⇒ todo
   `ya-existia`, cero escrituras.
2. **Determinista byte a byte** (D12): el sembrador RELLENA plantillas, jamás improvisa ni
   crea de cero. Misma fuente + mismos inputs ⇒ mismo árbol.
3. **Aditivo puro:** nada del proyecto se mueve, borra ni reescribe (§1).
4. **Estados de salud** (A-D4, visibles, jamás fabricados):

   | Estado | Significa |
   |---|---|
   | `sana` | `.arnesia/` presente con TODOS los archivos del árbol §2 |
   | `ausente` | no existe `.arnesia/` en el proyecto |
   | `incompleta` | existe pero faltan archivos — el veredicto LISTA los faltantes |

   El doctor v0 (`arnesia init --check` · `POST /api/forja/semillas/chequeos`) juzga por
   **presencia de archivos, sin hashes**. `exit ≠ 0` si insana — «no avanzamos si no está
   sana».

## 5. Derivación — de dónde sale todo

**Fuente única = [`semilla/arnes.yaml`](../../../semilla/arnes.yaml)** (raíz del repo,
embebido en el binario vía `//go:embed all:semilla`, A-T2): la copia GRADUADA del `arnes.yaml`
v0 firmado en `stories/2026-07-10-terreno-conocimiento/` (la copia ES la graduación del
draft — A-D2). El parser lee el subset `territorios` + `gestion_trabajo.tipos_paquete` +
`proceso.spines` (yaml.v3; ilegible ⇒ error honesto).

| Destino | Plantilla fuente (`semilla/plantillas/`) | De qué parte del `arnes.yaml` deriva |
|---|---|---|
| `terreno/INDEX.md` | `terreno-INDEX.md` | `territorios` + las 11 `dimensiones` (id·territorio·zachman) |
| `terreno/<territorio>/INDEX.md` (×3) | `territorio-INDEX.md` (genérica, se renderiza por territorio) | `territorios.<id>` + sus `dimensiones` |
| `product/backlog.md` | `product-backlog.md` | — (esqueleto) |
| `product/roadmap.md` | `product-roadmap.md` | — (esqueleto; regla D17 release=TAG) |
| `product/stories/INDEX.md` | `product-stories-INDEX.md` | `gestion_trabajo` (tipos de paquete) |
| `wip/INDEX.md` | `wip-INDEX.md` | `gestion_trabajo.carpetas.snapshot` (AUTO-GEN/DO-NOT-EDIT) |
| `proceso/historia/0N-*.md` (×6) | `proceso/historia/0N-*.md` | `proceso.spines.historia` (paso·rol·artefacto·cond) + `gestion_trabajo.tipos_paquete.historia` |
| `proceso/spike/0N-*.md` (×2) | `proceso/spike/0N-*.md` | `proceso.spines.spike` + `gestion_trabajo.tipos_paquete.spike` |

`semilla.lock.json` y los `.gitkeep` NO tienen plantilla: los produce el sembrador.

**Fase 2 — GAPs declarados (no prometidos como hechos):**

- leer el `arnes.yaml` DEL PROYECTO (flag `--arnes-yaml`; el seam del parser ya queda);
- las **11 dimensiones con hojas** (hoy: listadas `pendiente` en los INDEX, sin hoja);
- **nodos `.arnesia/` en el Mapa** (la `Clase` de nodo es enum firmado de 10 primitivas;
  enmendar ese contrato no entra en v0);
- evaluación de **deriva de semilla** contra el lock (§3);
- *(stretch)* eslabón `arnesia-semilla` en el scanner.

## 6. Set de placeholders (contrato de plantillas)

Las plantillas son Go `text/template`. Set mínimo, usado consistente:

| Placeholder | Tipo | Semántica | Dónde aparece |
|---|---|---|---|
| `{{.ProyectoNombre}}` | string | nombre del proyecto sembrado (v0: basename del dir destino) | título de `terreno-INDEX` · `territorio-INDEX` · `product-*` · `wip-INDEX` |
| `{{.ArnesID}}` | string | `arnes.id` del `arnes.yaml` fuente | frontmatter `arnes:` de TODAS las plantillas |
| `{{.Fecha}}` | string `AAAA-MM-DD` | fecha de siembra | frontmatter `sembrado:` de TODAS las plantillas |
| `{{.Territorio}}` | string | id canónico del territorio | solo `territorio-INDEX.md` |
| `{{.Label}}` | string | label del territorio (regla E, del `arnes.yaml`) | solo `territorio-INDEX.md` |
| `{{.Dimensiones}}` | lista `{ID, Zachman, Label}` | dimensiones del territorio (se recorre con `range`) | solo `territorio-INDEX.md` |

Los datos fijos del spine (paso · rol · artefacto · cond · estados · wip_caps) van **horneados**
en cada plantilla de proceso — son la graduación del v0 firmado, no inputs (determinismo D12).

## 7. Enforcement — `pendiente-de-construir` (HONESTO)

Ningún check de este contrato corre hoy. Llegan con los tickets del paquete; hasta entonces el
estado es `pendiente-de-construir`, visible, jamás verde fabricado:

| Check | Qué fija | Llega con |
|---|---|---|
| parser golden contra `semilla/arnes.yaml` | el subset parseado coincide con el firmado | A-T2 |
| scaffolder idempotente (`t.TempDir`) | re-siembra ⇒ todo `ya-existia`, cero re-escritura | A-T2 |
| árbol sembrado = árbol §2 | `arnesia init` crea EXACTAMENTE el contrato | A-T2 |
| lock al final, schema 0 | `semilla.lock.json` con los 5 campos §3 + sha256 reales | A-T2 |
| doctor 3 estados | `sana`/`ausente`/`incompleta` + faltantes listados · exit≠0 insana | A-T2/A-T3 |
| `refiereAlguno` no matchea `/proyecto/.arnesia/x` | el guardrail runtime no bloquea la semilla del proyecto | A-T5 |
| E2E: `arnesia init` sobre proyecto real + `POST /api/forja/semillas` | AC del spec (PARIDAD Installer 3) | A-T3 |

## Changelog

- 2026-07-30 · v0 — nace de A-D1..A-D4 (paquete `2026-07-30-arnesia-en-el-proyecto`, A-T1):
  árbol exacto, lock schema 0, reglas de siembra/salud, derivación desde `semilla/` embebido,
  set de placeholders, enforcement declarado `pendiente-de-construir`.
