# contracts/gen/ — tipos generados (NO editar a mano)

Los tipos de este directorio se **generan** desde el single source of truth en
[`../schema/`](../schema/) y [`../api/`](../api/). Escribirlos a mano viola
[`../../boundaries/dominio-independiente-de-transporte.md`](../../boundaries/dominio-independiente-de-transporte.md)
(check `dtos-generados`) y
[`../../boundaries/contrato-de-caja-es-fitness-function.md`](../../boundaries/contrato-de-caja-es-fitness-function.md)
(check `tipos-generados`).

Vacío por ahora — pero NO porque falte código: el código Go/TS existe desde HS-06/HS-08
(`internal/`+`cmd/`, `web/src/`). Lo que AÚN no corrió es el **codegen**: `gen/go` y `gen/ts`
siguen vacíos, `domain.Contract` y los tipos TS del grafo están escritos **a mano**, y los checks
`tipos-generados`/`dtos-generados` están en `warn` (deuda registrada; el step
`openapi-gen-check` de CI se auto-activa cuando el directorio generado exista). Cuando se cablee,
un `go:generate` / target de build corre:

```
# dominio (Go + TS de una sola fuente)
quicktype -s schema arch/contracts/schema/graph.l0.schema.json   -o arch/contracts/gen/go/graph_l0.go   --lang go   --package contracts
quicktype -s schema arch/contracts/schema/box.contract.schema.json -o arch/contracts/gen/ts/boxContract.ts --lang ts

# API (superficie HTTP/SSE)
oapi-codegen        -generate types,server -package api arch/contracts/api/openapi.yaml > arch/contracts/gen/go/api.go
openapi-typescript  arch/contracts/api/openapi.yaml -o arch/contracts/gen/ts/api.ts
```

Runtime validation en Go = `google/jsonschema-go` contra los `schema/*.json` (mismos archivos),
para validar cada `contract:` de caja al indexar (la fitness function del dominio).
