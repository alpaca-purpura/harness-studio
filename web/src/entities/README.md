# entities/ (FSD)

Modelos de dominio del frontend: **arnés**, **caja** (componente de proceso), **corrida**,
**empresa/puesto** (organigrama), **hallazgo**. Cada entity = carpeta con Public API + segmentos
`ui/model`. Cross-imports entre entities SOLO vía la Public API `@x` (nunca import lateral directo).

Vacío por ahora. Regla FSD: importa solo `shared`. Los tipos generados del contrato L0/box viven en
`shared/api` (generados desde `docs/architecture/contracts/`); las entities los envuelven en su modelo de UI.
