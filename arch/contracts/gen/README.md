# contracts/gen/ — tipos generados (NO editar a mano)

Los tipos de este directorio se **generan** desde el single source of truth en
[`../schema/`](../schema/) y [`../api/`](../api/). Escribirlos a mano viola
[`../../boundaries/dominio-independiente-de-transporte.md`](../../boundaries/dominio-independiente-de-transporte.md)
(check `dtos-generados`) y
[`../../boundaries/contrato-de-caja-es-fitness-function.md`](../../boundaries/contrato-de-caja-es-fitness-function.md)
(check `tipos-generados`).

Vacío por ahora: no hay código Go/TS todavía (fase 5). Cuando aterrice el módulo, un
`go:generate` / target de build corre:

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
