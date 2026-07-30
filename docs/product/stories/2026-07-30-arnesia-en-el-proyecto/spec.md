# Spec-lite — arnesia en el proyecto

> Sancionada por la firma del corte (decisiones.md).

**RF-A.1 (A-T1)** Contrato `docs/architecture/contracts/semilla-arnesia.md` v0: árbol
exacto A-D1, semántica de `semilla.lock.json`, idempotencia, estados de salud, gaps
declarados (deriva de semilla, `--arnes-yaml` del proyecto, nodos en Mapa).
**RF-A.2 (A-T2)** `semilla/` en raíz del repo (`arnes.yaml` graduado + plantillas
text/template: 8 de proceso + INDEX/backlog/roadmap) + `//go:embed all:semilla` en
`embed_doctrina.go` + `internal/adapters/forja/{parser,scaffolder,doctor}.go`:
- parser: subset `territorios` + `gestion_trabajo.tipos_paquete` + `proceso.spines`
  (yaml.v3; ilegible ⇒ error honesto);
- scaffolder: render determinista, jamás pisa (informe `Creados`/`YaExistian`),
  escribe `semilla.lock.json` al final;
- doctor: `Chequear(dir)` ⇒ `sana|ausente|incompleta` + faltantes (sin hashes en v0).
**RF-A.3 (A-T3)** `internal/domain/forja.go` (`InformeSemilla`, `SaludSemilla`) +
`internal/ports/forja.go` (`ForjaPort{Sembrar,Chequear}`) +
`internal/usecase/forja_service.go` (valida dir con la política de
`validarRootPortafolio`; usecase no importa adapters — wiring en cmd) +
`cmd/arnesia/init.go`: `arnesia init [dir] [--check] [--json]`; sin `--check` siembra
Y chequea; **exit≠0 si la instalación queda insana**. HTTP: `POST /api/forja/semillas`
`{path}` → informe · `POST /api/forja/semillas/chequeos` `{path}` → salud (+openapi).
<!-- corrección A-T3 (2026-07-30): el campo del body es `path`, no `dir` — el MISMO
nombre que POST /api/portafolio/escaneos; un sinónimo nuevo por ruta sería drift. -->
**RF-A.4 (A-T5)** Enmienda doctrine A-D3 + test `refiereAlguno` con
`/proyecto/.arnesia/x` NO matchea marcas del kit. Efecto colateral: fingerprint del
kit cambia ⇒ re-materialización automática a `~/.arnesia` al próximo arranque.

Capabilities: módulo nuevo `docs/product/capabilities/forja/` — `sembrar-semilla.yaml`
+ `chequear-semilla.yaml` (punteros a parser/scaffolder/doctor/init/endpoints).

AC: (a) `arnesia init /tmp/x` crea el árbol del contrato; re-run ⇒ todo `ya-existia`;
(b) `--check` exit 0 sana / exit≠0 con faltantes listados al romper un archivo;
(c) unit: parser golden contra `semilla/arnes.yaml`, scaffolder idempotente en
`t.TempDir`, doctor 3 estados; (d) `refiereAlguno` test verde; (e) E2E laptop:
`arnesia init` sobre proyecto real + `curl POST /api/forja/semillas`.

Recortes (orden): endpoints HTTP (queda CLI, ya cumple A) → stubs de terreno a un solo
INDEX → eslabón scanner (ya stretch).
