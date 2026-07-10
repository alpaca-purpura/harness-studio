# Casuística de artefactos — catálogo completo (D9–D11)

Fecha: 2026-07-08 · Disparador: 4 escenarios del operador (largo alcance · mejora ·
cantidad · sin plantilla). Cada caso resuelto en 4 planos: **doctrina** (VISION/
METODOLOGIA) · **contrato** (necesita/entrega) · **motor** (conductor/conformance) ·
**mapa** (vista A firmada). Demo = `mockup-artefactos.html` v2 (ejemplo que lo ejercita
entre paréntesis).

## A. Por origen del input

| # | Caso | Resolución |
|---|------|------------|
| C1 | **Hand-off adyacente** (N → N+1) | Base firmada D2: chip en el gutter entre carriles. *(dogfood: spec.md)* |
| C2 | **Salto de fases** (N → N+k, k>1) | El chip vive UNA vez donde nace; consumidor lejano = edge `lee` largo + referencia ↖ en su panel de entrada al seleccionarlo (D11a/b). Jamás chip duplicado (doble fuente visual). *(Luana: spec.md → construir Y promover)* |
| C3 | **Externo de usuario** (idea, aprobación de gerencia) | `necesita.de: usuario` → chip punteado «externo» en el gutter de entrada del consumidor; sin plantilla; admisión D10. *(cobranza: aprobación de gerencia)* |
| C4 | **Externo de terceros/sistemas** (factura, orden de compra) | `necesita.de: terceros` → ídem C3 con origen «terceros». El caso factura del operador. *(cobranza: factura del proveedor)* |
| C5 | **De la Base** (estándar, glosario) | NO es chip: es componente, no trabajo — edge `lee` al nodo de la Base (ya existe). *(dogfood: std-spec)* |
| C6 | **De librería/maquinaria/marcas** | Ídem C5 — componentes. Sin chip. |

## B. Por consumo

| # | Caso | Resolución |
|---|------|------------|
| C7 | **Fan-out** (1 art → varias cajas) | Un chip, varios edges `lee`. *(Luana: spec.md)* |
| C8 | **Fan-in** (1 caja ← arts de varias cajas) | La selección lo resuelve: panel de entrada lista TODO; chips reales se iluminan donde viven (D11b). *(Luana: promover ← 4 inputs)* |
| C9 | **Cantidad (8+ inputs / gutter saturado)** | Tope 3 chips/gutter + «+N más» expansor; relacionados con la selección saltan el tope (D11c). *(cobranza: gutter de pago)* |
| C10 | **Opcional** (`requerido: false`) | Edge tenue + tag «opcional» en la referencia; el conductor NO bloquea el spawn por su ausencia (D7). *(Luana: promover ← diseño.md opcional)* |

## C. Por ciclo de vida

| # | Caso | Resolución |
|---|------|------------|
| C11 | **Mejora/refina** (output = mismo art mejorado, OTRO escritor) | `entrega[].refina` D9: cadena lineal de revisiones, versión derivada, chip repetido con `↻ v2` en el gutter del refinador. *(cobranza: factura.pdf → validar → factura.pdf v2)* |
| C12 | **Rework** (ruta de vuelta: hallazgos → el MISMO escritor itera) | NO es refina: el escritor único reescribe su propia entrega. Vive en `ruta[]` (estático, Inspector) y en el tablero Flujo (instancias). El mapa de estructura no lo pinta como chip nuevo. *(dogfood: reviewer → builder)* |
| C13 | **Acumulativo transversal** (bitácora/log que crece todo el proceso) | Anti-patrón como art multi-escritor (rompe escritor_unico). Dos salidas doctrinales: telemetría de la Guardia (hook escribe, no contrato de caja) o cadena refina explícita si es trabajo real. |
| C14 | **Efímero intra-caja** (borradores internos) | NO se declara: el contrato solo declara lo que CRUZA fronteras (hand-off o gate). Sin chip, sin check. |
| C15 | **Entrega final** (caja terminal) | Sin consumidor ≠ dead-end: terminalidad derivada del spine (HS-12) → badge verde «salida del proceso». *(cobranza: comprobante de pago)* |
| C16 | **Dead-end real** (output que nadie consume, caja no terminal) | Chip warn «sin consumidor» + check `dead-end` (D8). *(Luana: notas de build)* |

## D. Por naturaleza y validación

| # | Caso | Resolución |
|---|------|------------|
| C17 | **Documento con plantilla** (spec.md, asiento.md) | D4/D5/D6 completo: reference + scripts + hooks + gate Gherkin. Badge «plantilla». |
| C18 | **Documento sin plantilla** (memo libre, factura recibida) | Sin badge (silencio honesto). Si es entrega propia: gate valida lo que el Gherkin declare o `tipo: none` honesto. Si es input: admisión D10. |
| C19 | **Etiqueta sin path** («código + tests», «veredicto») | Cursiva en el chip (D3); check warn `art-es-path` solo exige path a `entrega[0]` de cajas pipeline/excepcion. |
| C20 | **Opaco/binario** (pdf, imagen, dataset) | `path` sí, plantilla no; icono relleno; gate valida forma por script (mime/campos extraíbles); document-as-cache NO aplica al binario (D10) — status en el doc de la caja; digest sidecar = deuda declarada. *(cobranza: factura.pdf, comprobante.pdf)* |

## E. Bordes de coherencia (motor)

| # | Caso | Resolución |
|---|------|------------|
| C21 | **Mismatch de nombres** (entrega «spec.md» vs necesita «espec.md») | Chip huérfano-visible + check `art-identidad-coherente` (D8) — jamás fusión silenciosa. |
| C22 | **Multi-entrega por caja** | Soportado (D3: `path`/`plantilla` por entrega; el conductor hoy solo lee `entrega[0]` — deuda conocida). *(Luana: construir → 2 entregas)* |
| C23 | **Caja `abierto` sin document-as-cache estricto** | Sus chips existen si declara necesita/entrega; plantilla/llenado NO se le exige (§8.1, D5). |
| C24 | **Dos escritores del mismo art SIN refina** | Sigue siendo hallazgo `escritor-unico` (error). `refina` es la ÚNICA puerta legal a la multi-escritura, y es en cadena, jamás en paralelo (D9b). |

## Reglas transversales que cierran el catálogo

1. **Un chip = una revisión de un artefacto con productor único** (o un input externo).
   Todo lo demás se deriva.
2. **Plantilla ⊂ entregas propias.** Lo que llega del mundo se ADMITE (existencia +
   forma), no se plantilla (D10).
3. **El mapa pinta TIPOS; el Flujo pinta INSTANCIAS** (frontera A3, decisión #6 Gate 1) —
   rework, bucles vivos y estados de llenado pertenecen al tablero Flujo telemetry-gated.
4. **Nada se duplica visualmente**: chip una vez; referencias ↖ son índice, no segundo
   chip (D11).
