# Bugfix — `ctx-derivado-etiquetado` probablemente stale (HS-16 lo dio por bloqueado en OTel)

> Origen: hallazgo colateral del barrido de deuda viva HS-27 (2026-07-24), investigando el
> ítem 6 (confirmar HS-16 Grupo A/C). NO investigado a fondo ni construido en esa sesión — este
> doc es el punto de arranque para la próxima.

## El hallazgo (con cita real, no de memoria)

`docs/architecture/boundaries/conductor-no-parsea-jsonl.md:60` — checklist:

```
| ctx-derivado-etiquetado | `% contexto` se computa y se marca como métrica derivada | info | «% contexto presentado como si fuera medido» | arch_test.go |
```

Enforcer = `arch_test.go` genérico (sin nombre de test) → el parser (`inferMechanism`,
`internal/adapters/conformance/ruleset/parser.go`) lo resuelve a mecanismo `arch-test` pero SIN
test nombrado → el adapter `ArchTest.Run` devuelve `deferred` con el mensaje "enforcer genérico
`arch_test.go` sin función nombrada" (`internal/adapters/conformance/mechanism/adapters.go:53-55`).

HS-16 (2026-07-08) agrupó este check junto a `hooks-desde-otel` bajo "esperan OTel" — pero **la
función que este check describe YA EXISTE, y no usa OTel para nada**:

- `internal/adapters/agent/claudecode/conductor.go:611-639`, función `ctxPct` — calcula
  `% contexto = (input+cacheRead+cacheCreation)/ventana` **desde `stream-json` puro** (el campo
  `usage` de los frames `assistant`/`result` del subproceso `claude`), exactamente la fórmula que
  describe `conductor-no-parsea-jsonl.md:46-47` como "métrica derivada nuestra".
- Ya tiene un test real y específico: `internal/adapters/agent/claudecode/conductor_test.go:216`,
  `func TestCtxPctUsaUltimoUsage`.
- Commit `b91a061` ("fix(conductor): ctxPct real...", 2026-07-22) — **posterior** a HS-16
  (2026-07-08). Es plausible que HS-16 haya bloqueado esto correctamente EN SU MOMENTO (la
  función quizás no existía o estaba rota) y el bloqueo simplemente quedó sin re-revisar tras el
  fix del 22/07.

## Lo que falta (si se confirma que ya no está bloqueado)

**NO requiere código nuevo, solo el mismo patrón "wrapper" que HS-16 Grupo B / el fix del ítem 5
de HS-27** (`docs/architecture/fitness/arch_test.go:TestDogfoodComposicionFabricaConforma` /
`TestGoArchLintAdapterCorreDeVerdad` son el precedente exacto a copiar):

1. Verificar que `TestCtxPctUsaUltimoUsage` cubre de verdad la afirmación del check ("se computa
   Y se marca como derivada" — el test de hoy puede que solo cubra el cálculo, no el
   etiquetado/marcado; leer el test antes de asumir que cubre las DOS partes del check).
2. Actualizar el enforcer de `conductor-no-parsea-jsonl.md:60` para citar el test real,
   siguiendo el patrón repo-relativo ya soportado por el parser (`pkgOf`,
   `internal/adapters/conformance/ruleset/parser.go:88-101`): algo como
   `internal/adapters/agent/claudecode/conductor_test.go:TestCtxPctUsaUltimoUsage` (o un test
   nuevo si el existente no cubre el "etiquetado").
3. Correr `go run ./cmd/arnesia conformance --todo` y confirmar que pasa de `deferred` a `pass` —
   si no pasa, el check describe algo que el test actual NO prueba (hay que escribir la pieza
   que falte, no forzar el enforcer).

## Riesgo de NO hacerlo bien

No asumir que esto es un "cierre gratis" sin leer el check con cuidado: "se marca como métrica
derivada" podría exigir algo más que el cálculo correcto — por ejemplo, que el campo expuesto al
FE (`ctx_pct` en algún frame/DTO) tenga un nombre o metadato que lo distinga de una métrica
medida (no inventada). Verificar esa parte del criterio antes de dar el check por resuelto.
