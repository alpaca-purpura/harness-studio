# SPEC — Torre de Control read-only (v1) — CONGELADA

> **Congelada 2026-07-03** (épica Torre de Control, F1 · ficha `DH-04`). F2 implementa ESTO
> sin retocar diseño; cambio de diseño = nueva ficha DH-NN + edición explícita aquí.
> Norte de la épica: `../epicas/torre-de-control/NORTE.md` (v3 firmado). Decisiones D1-D4 de
> esta SPEC: checkpoint expirado → recomendadas aplicadas (mecánica de la casa), registradas
> en `../epicas/torre-de-control/ESTADO.md` y en `DH-04`.

## 0 · Qué es

La **primera vista** de la consola DevHub: un tablero read-only que responde, por sistema en
construcción, las 5 preguntas del caso motivador — ¿gate de fábrica verde/rojo? ¿en vuelo sin
commitear? ¿último tag? ¿actividad del ledger? ¿estado de la rama? — para la fábrica
(prenter-harness, 4 células), la EMPRESA (prenter) y los engagements/portfolio que el registro
declare. **Cero escritura sobre los repos monitoreados.**

## 1 · Modelo de datos — 2 ejes (D1)

"Sistema a monitorear" ≠ "repo" 1:1 (hallazgo F0 §11). Dos ejes:

```
sistema                      # unidad del registro (fila de la torre)
├── repo                     # eje GIT — UN veredicto por sistema
│   ├── rama                 # rama activa + upstream + ahead/behind
│   ├── en_vuelo             # dirty files (porcelain)
│   ├── ultimo_tag           # último tag alcanzable + commits desde
│   └── gate_fabrica         # veredicto del check anti-drift declarado
└── proyectos[]              # eje CÉLULA/PROYECTO — N por repo (0..N)
    ├── slug · tipo          # tipo ∈ {celula, sdd}
    ├── ledger               # actividad del LEDGER de célula (nullable)
    └── board                # conteos de stories/releases SDD (nullable)
```

- **RN-01** · Los hechos git se computan **en el path del workspace** (`git -C <workspace>`),
  no en un root asumido: un workspace puede ser **worktree** de otro repo (caso real:
  `luana-vitalia` es worktree de `luana-platform` — rama y dirty son del worktree; el tag es
  del repo común). Se reporta también `repo_root` real (`git rev-parse --show-toplevel`).
- **RN-02** · Todo el eje PROYECTO es **nullable con empty-state honesto** (caso real:
  `prenter` no tiene LEDGER.md raíz ni board ni tags — la fila muestra "—", jamás inventa).
  `marketing/ledger.yaml` de prenter es ledger COMERCIAL de deals: otro dominio, la torre NO
  lo lee.
- **RN-03** · Los proyectos se **descubren por convención** en runtime (no se curan en v1):
  `products/<dir>/LEDGER.md` presente → proyecto `tipo: celula` (slug = dirname) ·
  `docs/product/` propio o `<sistema>/docs/product/` → proyecto `tipo: sdd` (la convención
  que `sistemasIn` ya implementa). Repo sin nada → `proyectos: []`. Declarar proyectos
  esperados en el curado (discrepancia curado-vs-descubierto, estilo `huerfano-emitiendo` de
  flota) = extensión v2, fuera de esta SPEC.

## 2 · Registro de sistemas — 3 bandas, espejo de OBS-15 (D2)

El registro vivo `~/.cockpit/cockpit.yaml` hoy lo genera un script huérfano pre-I-39
(chris-corp). Se reemplaza la FUENTE espejando el corte de flota (OBS-15), no el wire format:

| Banda | Qué | Dónde |
|---|---|---|
| **CURADO** | qué sistemas monitorear: slug · nombre · tipo · `gate_check` esperado · ref al CRM si es engagement | `prenter/sistemas/sistemas.yaml` (la EMPRESA — I-39; reemplaza la prosa de `SISTEMAS.md` que aún apunta a chris-corp) |
| **RUNTIME** | el join curado × config del operador → registro vivo | adapter `products/devhub/scripts/gen_registro.py` emite `~/.cockpit/cockpit.yaml` **conforme al wire format actual** (`registry.go` no cambia para leerlo; red: `TestRegistryWireFormatRoundTrip`) |
| **FÁBRICA** | solo contrato + check + fixture | schema del curado + `check_registro.py` + fixture; el adapter importa el MISMO `validate()` del check y se niega a emitir lo que el gate rechazaría (patrón SC-30) |

- **RN-04** · El curado lleva **slugs/refs, jamás rutas** (I-39). Las rutas locales viven en
  la config del operador: `~/.config/prenter/devhub.yaml` → bloque `torre:` con
  `registro:` (ruta al curado) y `workspaces: {slug: path}` (espejo exacto de
  `harness-studio.yaml` de OBS-15).
- **RN-05** · `gate_check` es el **comando de solo-verificación declarado por sistema** en el
  curado (prenter-harness: `python3 tooling/scripts/gen_all.py --check`; prenter: su
  `harnesses/scripts/check.py`; engagement: el check de su kit). Sin declaración → categoría
  `sin-gate`. El comando DEBE ser side-effect-free (responsabilidad del curador); la torre lo
  corre con timeout y `cwd = workspace`, nunca comandos no declarados.
- **RN-06** · `gate_check` viaja al runtime como **campo aditivo** del wire format
  (`cockpit.yaml`), modelado en los structs de `registry.go` para sobrevivir el round-trip de
  `saveRegistry` (que reescribe el archivo entero). Cambio aditivo + test de round-trip.
- **RN-07** · Adapter tolerante a fuente ausente con error claro (espejo `gen_flota.py`):
  el gate de prenter está rojo por reestructura ajena → la banda curada puede tardar en
  commitearse; F2 arranca contra el **fixture** de la fábrica sin bloquearse.
- **RN-08** · El huérfano `chris-corp/harnesses/scripts/gen_cockpit_registry.py` queda
  obsoleto de facto — NO se arregla ni se retira de rebote (ficha aparte si amerita).

## 3 · Contrato de lectura: ledger de célula (D3)

**UN parser del formato, en la fábrica.** `gen_ledger.py` (que ya parsea el LEDGER.md global
dentro de `gen_all`) se extiende para emitir un espejo máquina por célula; DevHub solo lee
YAML (el reader de `handleLedger` ya existe). Cero parser Markdown en Go.

- **RN-09** · Ubicación del espejo: `products/<célula>/ledger.yaml`, header `# GENERADO`,
  cubierto por el gate de fábrica → los LEDGER.md de célula ganan gate (ficha mal formada =
  gate rojo; mata el hallazgo F0 §8 como consecuencia arquitectónica, no de rebote).
- **RN-10** · Shape (mismas llaves que el global + `log` aditivo):

  ```yaml
  sistema: devhub            # slug de la célula
  fuente: products/devhub/LEDGER.md
  fichas:
    - {id: DH-03, titulo: "...", estado: decidida, vigencia: vigente, nota: ""}
  log:                       # espejo de la tabla § Log — da la FECHA de actividad
    - {fecha: 2026-07-03, decision: "...", fichas: [DH-03]}
  ```

- **RN-11** · Contrato de la columna `ledger` de la torre: `ultima_ficha` (id·titulo·estado·
  vigencia) + `ultima_fecha` (= max fecha del `log`) + `total_fichas`. Célula sin
  `ledger.yaml` → `sin-ledger` (empty-state honesto, igual que `handleLedger` hoy).
- **RN-12** · Tocar `gen_ledger.py` es cambio de FÁBRICA al servicio de las células: se
  implementa en F2; si el operador lo eleva a governance de ecosistema (gate para ledgers de
  todas las células), ficha I-NN en ese momento.

## 4 · Veredictos — categorías semánticas fijas + dato nativo (D4, principio 1 del NORTE)

Patrón DevLake: **la categoría se guarda JUNTO al dato nativo, nunca lo reemplaza.** Todo
comportamiento de UI se ata a la categoría, jamás al detalle. Los nombres de categoría son
**inmutables desde v1** (regla Azure DevOps) — F4 los mapeará como statusCategories del
descriptor sin romperlos.

**Nombre del veredicto anti-drift (D4): UI = «gate de fábrica» · API = `gate_fabrica`.**
Nunca «gate» pelado dentro de DevHub (colisiona con G1-G8 del board SDD, `gates.go`).

| Columna | Categorías (enum FIJO) | Dato nativo al lado |
|---|---|---|
| `gate_fabrica` | `verde · rojo · sin-gate · no-medido` | `exit_code`, `resumen` (primeras líneas de salida), `comando`, `medido_en` |
| `en_vuelo` | `limpio · en-vuelo · no-medido` | `archivos[]` (path + estado XY porcelain), `total`, `medido_en` |
| `rama` | `sincronizada · adelante · atras · divergida · sin-upstream · no-medido` | `rama`, `upstream`, `ahead`, `behind` |
| `ultimo_tag` | `taggeado · sin-tag · no-medido` | `tag`, `fecha`, `commits_desde` |
| `ledger` (proyecto) | `con-ledger · sin-ledger` | `ultima_ficha`, `ultima_fecha`, `total_fichas` (RN-11) |
| `board` (proyecto) | `con-board · sin-board` | conteos de stories por estado + release en curso |

- **RN-13** · `no-medido` cubre: aún no corrido, timeout, error de ejecución — con el error
  nativo al lado. Jamás se disfraza de `rojo` (rojo = el check corrió y FALLÓ).
- **RN-14** · Juicios de valor tipo "ledger quieto hace N días" NO son categorías v1: la
  fecha nativa es visible y el juicio queda para la UI futura (evita hornear umbrales).

## 5 · API y refresh/staleness

- **RN-15** · `GET /api/torre` → `{ sistemas: [ { slug, nombre, repo: {rama, en_vuelo,
  ultimo_tag, gate_fabrica}, proyectos: [ {slug, tipo, ledger, board} ] } ], generado_en }`.
  Handler nuevo; ninguna de las 27 rutas existentes cambia de contrato.
- **RN-16** · Hechos baratos (git status/branch/tag, lectura de YAML) se computan por request
  con cache corto (TTL ≤ 5 s). `gate_fabrica` es CARO: se mide al boot del daemon y por
  re-medición explícita (`GET /api/torre?medir=gate_fabrica`), TTL 10 min; jamás en un loop.
- **RN-17** · Todo veredicto viaja con `medido_en`; la UI muestra la EDAD del dato (staleness
  visible, no silenciosa). Sistema cuyo workspace no existe en disco → fila con categoría
  `no-medido` en todo, no desaparece (el registro dice que debería existir — eso ES señal).

## 6 · UI — principio omnigent

- **RN-18** · Cada pixel = objeto del dominio operable: click en `gate_fabrica` rojo → salida
  completa del check; `en_vuelo` → lista de archivos (file-API existente); `ultimo_tag` →
  historial; `ledger` → la ficha en su LEDGER.md; `rama` → detalle ahead/behind. Veredictos
  accionables (texto de qué hacer), NO gráficas/tendencias.
- **RN-19** · Layout v1: tabla sistemas × columnas del eje REPO; fila expandible al eje
  PROYECTO. Nada más se congela del visual.

## 7 · Fuera de alcance (v1)

Escritura sobre repos monitoreados · PRs/remotos (más allá de ahead/behind del upstream ya
fetcheado — la torre NO hace `git fetch`) · descriptor de proceso (F4) · arquitectura por
sistema / `meta.clase` (F3 — esta SPEC no lo contradice: el eje PROYECTO es donde F3 colgará
su renderizador) · telemetría/evals (frontera dura #1: eso es Harness Studio) · discrepancias
curado-vs-descubierto (v2, RN-03) · umbrales de actividad (RN-14).
