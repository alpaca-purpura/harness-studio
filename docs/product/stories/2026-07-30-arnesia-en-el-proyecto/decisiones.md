# Decisiones — arnesia en el proyecto (semilla `.arnesia/`)

> **FIRMADAS 🧑‍⚖️ 2026-07-30 (Chris)** — vía aprobación del plan de sesión del MVP.

- **A-D1 — contrato as-code en `docs/architecture/contracts/semilla-arnesia.md`**
  (contracts, no conventions: es un contrato de reconocimiento como
  `nomenclatura-arnes.md`, enforced por los tests Go del sembrador). Árbol mínimo
  derivado del modelo FIRMADO, cero dimensiones inventadas:
  `.arnesia/{semilla.lock.json, terreno/{INDEX + proposito|producto|organizacion
  INDEX-stubs}, product/{backlog.md, roadmap.md, stories/INDEX.md},
  wip/{INDEX.md DO-NOT-EDIT, activo/, done/}, proceso/{historia/00-research..
  05-paridad, spike/00-investigar,01-decidir}}`. Plantillas con frontmatter derivado
  de `proceso.spines` + estados/wip_caps del tipo (D20 multi-pipeline · D12 el
  forjador RELLENA plantillas, jamás crea de cero). **`arnes.l0.json` SE QUEDA en la
  raíz del proyecto** — moverlo rompe el contrato firmado `nomenclatura-arnes.md`,
  loader y scanner; `.arnesia/` es puramente aditivo.
- **A-D2 — scaffolder = comando Go determinista (`arnesia init`), NO skill LLM.**
  D12 exige determinismo byte a byte; funciona sin daemon/rol/sesión. Fuente =
  `semilla/` embebido nuevo en la raíz del repo (NO dentro de `kit/`: el kit viaja
  como `--plugin-dir` a cada spawn y no se infla con archivos inertes; F-D5 pide
  «embebido en el binario» y un embed hermano lo cumple). `semilla/arnes.yaml` = copia
  GRADUADA del v0 firmado de terreno-conocimiento (la copia es la graduación del
  draft). Parser yaml.v3 (ya es dep; patrón `lockFile` del scanner) — deja el seam
  para leer el `arnes.yaml` del PROYECTO en Fase 2 (flag `--arnes-yaml` declarado como
  TODO en la hoja). Siembra idempotente: archivo existente JAMÁS se pisa (reporta
  `ya-existia`). `semilla.lock.json` = baseline D8 mínimo (hashes por archivo; la
  evaluación de deriva de semilla = Fase 2, gap declarado).
- **A-D3 — enmienda doctrine ADITIVA, runtime intacto.** `kit/doctrine.md`
  §Frontera de cuerpos + `forjar-caja/SKILL.md` + nota con fecha en metodología (§
  «Regla dura», supersede con nota, no reescribe): `.arnesia/` del proyecto = zona de
  escritura sancionada del PROCESO (contrato `semilla-arnesia.md`); sigue PROHIBIDO
  copiar kit/doctrina/knowhow dentro del arnés. El runtime NO bloquea `.arnesia/` de
  proyecto (verificado: `ProtegerPaqueteCerrado` matchea `~/.arnesia`, no
  `<proyecto>/.arnesia`) — se FIJA con test de `refiereAlguno`.
- **A-D4 — salud mínima honesta, sin tocar loader ni conformance.** El doctor vive en
  `arnesia init --check` + `POST /api/forja/semillas/chequeos`: estados VISIBLES
  `sana|ausente|incompleta` + faltantes; exit≠0 si insana («no avanzamos si no está
  sana»). Nodos `.arnesia/` en el Mapa = Fase 2 (la `Clase` de nodo es enum firmado de
  10 primitivas; enmendar ese contrato no entra en 1 día). Stretch declarado: eslabón
  `arnesia-semilla` en el scanner.
