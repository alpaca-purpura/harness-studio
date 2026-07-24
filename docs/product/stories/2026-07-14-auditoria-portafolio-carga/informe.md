# Informe de auditoría — carga de arneses (Portafolio Slices 0+1)

> Auditor: Fable 5 · 2026-07-14. Método: 3 barridos paralelos (código · arch as-code ·
> capabilities, con verificación puntero-por-puntero) + ejecución REAL de todos los gates
> (`go test ./... -race` · golangci-lint · go-arch-lint · fitness · `conformance --todo`/
> `--arnes` · `estado.sh --check` · `pnpm run verify` · vitest unit+storybook) + revisión de
> los 12 últimos runs de CI en GitHub.

## Veredicto global

**Build sano y por encima del estándar de la casa.** Hexagonal limpio (usecase solo
domain+ports; adapter portafolio sin imports entre adapters; puertos puente S0-D13 correctos),
~40 tests Go + 26 unit FE + 38 stories `play()`, honestidad aplicada de punta a punta
(degradación visible, `deriva-no-evaluable` default, discrepancias nunca silenciosas), y las
desviaciones del ejecutor TODAS documentadas en el mismo turno (S0-D12..15, S1-D16..22 —
incluidos 2 bugs reales cazados por el E2E que el plan mandaba, exactamente para eso estaba).
Los 10 capabilities: punteros 100 % resolubles, 0 huérfanos R2, R4 consistente.

## Hallazgos (por severidad) y qué se hizo

| # | Sev | Hallazgo | Acción (este turno) |
|---|-----|----------|---------------------|
| H1 | **alta** | **CI de main ROJO 12 runs seguidos** (todo Slice 1 se construyó sobre CI roto): 3 issues golangci-lint introducidos por `4e058b6` (`fix(self-update)`, ANTERIOR al slice — 1 gosec G304 + 2 govet shadow). Como lint es el 1er step del job `go`, en CI **jamás corrieron** `go test -race`, go-arch-lint ni `estado-drift` para esos 12 commits. La paridad de Slice 1 lo declaró honesto (desviación #6) pero nadie reparó. | **REPARADO**: shadows corregidos de verdad; G304 con `//nolint` justificado (`validarRepo` ES la validación del path del operador — mismo patrón que los 2 G204 del archivo). `golangci-lint run` = 0 issues. |
| H2 | **alta** | **El motor `arnesia conformance` era CIEGO a los enforcers del boundary nuevo**: `portafolio-identidad-y-deriva-honesta` declara `status: enforced` con 4 tests COLOCADOS junto al código (`internal/...*_test.go`), pero el parser solo reconocía `arch_test.go:` y el runner solo el paquete `fitness/` → los 4 checks salían `deferred` en `--todo` pese a correr en CI. La construcción (test colocado) es MEJOR que el patrón test-proxy de HS-14 → se actualizó el as-code, no el build. | **MOTOR EXTENDIDO**: parser `reGoTest` (ruta repo-relativa `dir/foo_test.go:TestX` ⇒ arch-test) + `mechanism.pkgOf` (corre en el paquete del archivo; archivo inexistente ⇒ `VeredictoError`, jamás pass). Celdas de la tabla del boundary a ruta completa (el path ES el wiring). **`--todo`: pass 42→46, deferred 211→207, 0 fail/error.** Boundary v1.1 · CAP-27/CAP-28 actualizados (CAP-28 gradúa `vivo` con `adapters_test.go`). |
| H3 | media | **R1 overclaim**: la doctrina decía «cada puntero resuelve a archivo/símbolo real» y `cap_doctor.py` afirmaba «go test resuelve el símbolo» — FALSO: ambos enforcers solo stat-ean el archivo (parte antes de `#`); un símbolo renombrado pasa en silencio. (Los 51 símbolos del Portafolio SÍ existen — verificado a mano en esta auditoría.) | Textos corregidos a lo que ES (`codigo-traza-a-capability` v1.4 + comentarios de `cap_doctor.py`); **deuda «R1 a nivel símbolo» registrada en BACKLOG**. |
| H4 | media | `enforced_by:` de `codigo-traza-a-capability` apuntaba a `fitness/arch_test.go` pero los 4 tests viven en `fitness/capability_trace_test.go` (corría igual por ser paquete-nivel; la referencia mentía). | Refs corregidas a la ruta real (v1.4). |
| H5 | media | **Huecos de trazabilidad a nivel símbolo**: `postEscanear`/`postAgregar`/`deleteDesvincular` (3 de los 5 handlers HTTP) y el subcomando CLI `runPortafolio` no estaban reclamados por ningún capability (el archivo quedaba cubierto solo incidentalmente — R2 es file-level). | Punteros agregados a CAP-84/CAP-87 con change_log. |
| H6 | media | Deuda S1-D16 (`capMetas` del enforcer R4 no parsea YAML real: cualquier línea `status:` anidada pisa el root) estaba en `decisiones.md` pero NO en BACKLOG. | **Registrada en BACKLOG** (Deuda viva). |
| H7 | baja | `stack.md` stale: «`arch/` = 16 boundaries ~101 checks» — real: `docs/architecture/` 18 boundaries · 111 checks (INDEX.md correcto); además nombraba la ruta pre-homologación `arch/`. | Corregido con puntero al conteo canónico vivo. |
| H8 | baja | Prosa de `capabilities/INDEX.md` stale: «82 capabilities» (real 92) · «FE sin tests» (falso desde S1-D7: proyecto vitest `unit` + 26 tests) · apéndice «23 endpoints» sin los 5 del Portafolio. | Corregido (la distribución por estado se remite a `estado.sh`, no se teclea). |
| H9 | baja | `cap_doctor.py` GROUP_ORDER sin `portafolio`/`fe-portafolio` (caían al fallback alfabético). | Agregados. |
| H10 | baja | CAP-90 decía «10 stories» — el archivo tiene 11 (todas con `play()`; 2 las reclama CAP-92). `checkpoint.md` decía «T8 sin commitear» — ya está en main (`b35ff6c`). | Ambos textos corregidos. |

## Cierre de deuda H3 (2026-07-23)

`TestCapabilityPointerSymbolsResolve` construido (`capability_symbol_resolve_test.go`): Go exacto
vía `go/parser`, TS/TSX/Rust por regex de declaración+miembro+import (sin parser TS/Rust en la
stdlib de Go). Corrida contra las 101 hojas reales: 17 fallos iniciales — 15 eran símbolos reales
que la primera versión del regex TS (solo top-level) no reconocía (propiedades de store Zustand,
miembros de interface, imports re-usados; regex ampliado) y **2 eran los bugs de datos genuinos
que H3 predijo podían colarse**: `fe-shell/rail-de-sesiones.yaml` tenía un puntero
`#portafolio-picker-store` (nombre del archivo, no un identificador — corregido a
`#usePortafolioPicker`, el símbolo real que ese test importa y ejercita) y un puntero
`sessions-store.ts#create` sin relación con el widget (eliminado). Doctrina actualizada
(`codigo-traza-a-capability` v1.5). `go build ./...`, `go vet`, suite `fitness` completa y
`cap_doctor.py` verdes.

## Lo que NO se tocó (y por qué)

- **`paridad.md` de ambos slices**: son evidencia de las sesiones de build, pendiente/objeto de
  firma humana — la auditoría no reescribe evidencia. La desviación #6 de Slice 1 (lint
  preexistente) queda RESUELTA por H1; el operador lo ve aquí.
- **GAP-2 (re-key índice bare-id)**: sigue deuda viva legítima en BACKLOG — la mitigación FE
  (`idsColisionados` + confirm) es correcta para el alcance del slice.
- **Firma del gate de Slice 1**: sigue PENDIENTE del operador; nada de esta auditoría la simula.

## Prácticas del build que quedan SANCIONADAS como estándar

1. **Tests colocados junto al código como enforcers de boundary** (en vez de test-proxy en
   `fitness/`) — ahora el motor los corre de verdad (H2); patrón preferido en adelante.
2. **Desviaciones del ejecutor como S{N}-D numeradas en el mismo turno**, ancladas en código
   real y con veredicto del porqué (S0-D12..15, S1-D16..22) — es la disciplina §10 funcionando.
3. **E2E vivo contra la máquina real como ticket final** — cazó 2 bugs reales (S0-D14/D15) que
   ningún fixture habría encontrado.
