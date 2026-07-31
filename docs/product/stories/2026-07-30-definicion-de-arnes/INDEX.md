# Definición canónica de arnés + cadena de proceso — paquete de doctrina

> Nace 2026-07-30 de la pregunta fundacional del operador: «¿qué es un arnés? ¿un
> plugin puede tener múltiples arneses? ¿feature/bugfix/spike son arneses distintos?»
> (transcripta en `chris-input.md`). Es un paquete de DOCTRINA: el entregable es la
> definición firmada + su bajada as-code, no código de app.

**Qué es:** cementar QUÉ ES UN ARNÉS por definición para ArnesIA — el corte
(rol × proceso vs end-to-end vs tipo-de-trabajo), la relación arnés↔plugin, y la
**cadena** (el proceso end-to-end que atraviesa varios arneses) como entidad
declarada. Primero dev como dogfood, luego generaliza a otros rubros (p7).

## Etapas §10

| Etapa | Estado |
|---|---|
| Mockup | n/a (doctrina; la superficie visual es la Galaxia, ya firmada HS-03) |
| Decisiones | 🟡 EN DEBATE — DEF-D1 FIRMADA 🧑‍⚖️ · DEF-D2 ABIERTA (operador pidió debatir) · DEF-D3 propuesta lista, sin firma |
| Spec | ⬜ (tras firmar DEF-D2/D3: hoja canónica en `docs/` + schema de cadena) |
| Implementación | ⬜ (glosario canónico · afilar META `proceso` a referencia · tipos-de-paquete del arnés-dev) |
| PARIDAD | ⬜ |

## Retomar aquí

1. **Debate DEF-D2 abierto** — el operador NO ratificó la definición v1; leer
   [`debate-definicion.md`](./debate-definicion.md) (v1 · casos duros A-D · v2
   propuesta) y preguntar QUÉ caso le hace ruido antes de proponer de nuevo.
2. DEF-D3 (cadena) tiene directiva «diseñar ahora» — propuesta en
   `debate-definicion.md` §5, espera la firma junto con DEF-D2 (v2 y cadena se
   sostienen mutuamente).
3. Firmadas D2+D3 → spec: hoja canónica (candidato: `docs/product/` glosario o
   `docs/architecture/contracts/`) + declarar tipos-de-paquete del arnés-dev
   (dogfood de DEF-D1, aterriza en `semilla/arnes.yaml` ya graduado — carril A del MVP).
