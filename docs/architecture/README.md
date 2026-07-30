# docs/architecture/ — arquitectura-as-code + tecnologías

> Owner: `/architect` (ratifica) · `/pm`. "La arquitectura vive aquí, no en un doc que se pudre."
> Cada nodo = L1 (principio industria, fechado + fuente) ↔ L2 (realización en este árbol) +
> `Checklist evaluable` con `enforced_by:` 1:1 a un check ejecutable.

## Contenido

| Path | Qué |
|---|---|
| `INDEX.md` | tabla de boundaries + estado + enforcer |
| `boundaries/` | nodos de frontera (core-no-importa-shell, dominio-independiente-de-transporte, …) |
| `fitness/` | checks ejecutables que fallan CI (`.go-arch-lint.yml`, `arch_test.go`, `capability_trace_test.go`) |
| `conventions/` | estilo/lint/format/naming/commits/hooks/CI (Go+TS+Rust) |
| `contracts/` | schema-first: `schema/*.json`, `api/openapi.yaml`, `nomenclatura-arnes.md`, `semilla-arnesia.md` (siembra `.arnesia/` en el proyecto, v0) |
| `model/` | diagramas as-code (C4): `system-context.mmd`, `container.d2` |
| `stack.md` | tecnologías (← STACK.md) |
| `knowledge/` | metodología-as-code sobre elementos de arnés (dominio-producto ArnesIA) |

## Ruleset embebido — dónde lo lee el motor

`boundaries/` + `conventions/` + `knowledge/elements/` + `contracts/schema/` son el **ruleset que
consume el motor de conformance** (CAP-26). Viven acá y el binario los **embebe** (`embed_doctrina.go`
→ `//go:embed docs/architecture/knowledge docs/architecture/boundaries docs/architecture/conventions
docs/architecture/contracts/schema`). El parser (`internal/adapters/conformance/ruleset/parser.go`)
los lee por ruta `docs/architecture/{knowledge/elements,boundaries,conventions}` — misma ruta en disco
(dev) y en el FS embebido (instalación de cliente), cero divergencia.

**Relocación completada 2026-07-09** (homologación, Task 7): los árboles vivían en `arch/` +
`knowledge/` a la raíz; se movieron a `docs/architecture/` con `git mv` + reescritura coordinada de
`go:embed` + parser + `go-arch-lint.yml` + schema-loader + `arch_test.go` + ~115 referencias.
**Verificado verde:** `go build`, `go test ./...` (11 ok / 0 FAIL), `go-arch-lint` (OK, sin warnings),
`conformance --todo` (247 checks, idéntico al baseline) y `--arnes` (21·20·1, idéntico).

## Cadencia

Crece event-driven ("al cambiar"), no por calendario — ver `CADENCE.md`.
