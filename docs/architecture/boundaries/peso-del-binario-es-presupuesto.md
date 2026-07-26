---
regla: peso-del-binario-es-presupuesto
version: 1.0
updated: 2026-07-26
status: proposed
ledger: HS-28
sources:
  - url: https://go-proverbs.github.io/
    autoridad: experto
    revisado: 2026-07-26
  - url: https://research.swtch.com/deps
    autoridad: experto
    revisado: 2026-07-26
    # Reemplaza a go.dev/doc/security/best-practices, que se abrió y NO sostiene la afirmación:
    # esa página habla de ACTUALIZAR dependencias, no de evaluarlas antes de adoptarlas.
    # «Our Software Dependency Problem» (Russ Cox) sí es exactamente sobre eso.
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/INFORME.md
    autoridad: medicion-propia
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/investigacion-stack-embebible.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by: []
severity: error
---

# El peso del binario es un presupuesto, y una dependencia se mide antes de adoptarse

## L1 · Principio (estándar de industria)

**En software instalable, el binario ES el producto.** Cada MB tiene un costo de distribución, de
descarga, de CI, de superficie de auditoría y de tiempo de arranque; y cada módulo agregado es
superficie de cadena de suministro que hay que sostener. Dos formulaciones conocidas:

- *«A little copying is better than a little dependency.»* Copiar lo mínimo que se necesita de un
  formato o algoritmo es preferible a arrastrar el framework que lo implementa entero.
  *(experto: Go Proverbs, textual — es el proverbio nº 8)*
- **Una dependencia se inspecciona ANTES de adoptarse, no después.** *«Before you depend on a
  package you found on the internet, it is similarly prudent to learn a bit about it first»*; el
  antipatrón nombrado es que *«most of the time the entirety of the decision is "let's see what
  happens"»*, con el resultado de que *«we are trusting more code with less justification for doing
  so»*. Y el criterio es contextual: *«the context where a dependency will be used determines the
  cost of a bad outcome»* — un producto instalable que se distribuye firmado es contexto caro.
  *(experto: Russ Cox, «Our Software Dependency Problem»)*

Corolario propio, y es el que le da filo al nodo: **el peso se MIDE, no se estima.** Una
estimación de un informe externo no es un dato del árbol.

## L2 · Realización (este árbol Go)

El caso testigo, medido el 2026-07-26 sobre la base real (`net/http` + `modernc.org/sqlite`):

| opción para decodificar OTLP | delta de binario |
|---|---:|
| decodificador propio con `encoding/json` (~120 líneas, probado contra los payloads reales) | **+0,49 MB** |
| `collector/pdata` (`pmetricotlp` + `plogotlp`) | **+10,79 MB** |

**La investigación previa (F1) había estimado +1,7 MB para `pdata`. No reproduce: es 6×.** Sobre un
daemon de 22,71 MB, adoptarlo sería **+48 %** en un producto cuyo argumento de venta es «se
instala y ya». La medición cambió la decisión; la estimación la habría dejado pasar.

No es un caso aislado — el árbol ya rechazó por peso/superficie, cada vez con número:

| descartado | número que lo descartó |
|---|---|
| **DuckDB** | CGO (mata el binario estático) + 240 MB de descarga en CI + 1,08 GB en disco + MinGW-no-MSVC en Windows → check `sin-cgo`, ya enforced en `indice-desechable-jsonl-es-verdad` |
| **chDB** | sin Windows · `libchdb.so` dinámica (el usuario instalaría algo) · install `curl \| bash` |
| **`otlpreceiver` completo** | +8,4 MB · +60 módulos · API Go 0.x, por features que en loopback no se usan |
| **Prometheus TSDB** | 292 módulos, y su modelo (floats por serie) no absorbe eventos con atributos arbitrarios |
| **`ocb`** | genera un binario standalone, no una librería |

Las reglas que salen de ahí:

- **Todo formato de cable ajeno se decodifica con un subconjunto PROPIO y mínimo**, no con el
  framework del vendor. El subconjunto se declara: qué se parsea, qué se ignora, qué se rechaza.
  Lo que no se necesita no se decodifica — y lo que no se decodifica no es superficie de ataque.
  Precedente vivo: `adapters/agent/claudecode` es el dueño único del stream-json y traduce a
  eventos normalizados, sin librería de por medio.
- **El golden del decodificador es payload REAL capturado**, no un ejemplo escrito a mano. Es lo
  que atrapó que Claude Code emite `intValue` como número JSON, **off-spec** — un decodificador
  escrito contra la especificación y probado contra un fixture inventado habría pasado los tests
  y fallado en producción.
- **Todo decodificador de bytes ajenos tiene fuzz**, sembrado con esos mismos payloads. Es barato
  y es la única forma de saber que un cuerpo hostil no tumba el daemon.
- **Un módulo nuevo que suma más de 1 MB necesita su medición registrada** en el paquete de
  trabajo, con las dos builds comparadas. «Parece liviano» no es un dato.
- **El presupuesto se declara por paquete y se chequea en CI** contra el binario del release
  anterior. Para el módulo `telemetria/`: **+1,5 MB máximo**, daemon ≤ 25 MB sin `-s -w`.
- **Plan B declarado, no descartado.** `pdata` queda documentado como la salida si algún runtime
  futuro no deja elegir el protocolo. Rechazar por peso no es prohibir para siempre: es exigir que
  el que lo traiga vuelva con el número y la razón.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| dependencia-pesada-se-mide | un módulo nuevo que suma > 1 MB al binario entra con su medición registrada (dos builds comparadas), nunca con una estimación | error | «dependencia adoptada por estimación ajena» | (pendiente — job de CI + revisión del paquete) |
| presupuesto-de-binario | el delta contra el release anterior no supera el presupuesto declarado del paquete | error | «el binario creció fuera de presupuesto» | (pendiente — TestPresupuestoDeBinario) |
| decodificador-propio-y-minimo | un formato de cable ajeno se decodifica con un subconjunto propio y declarado, no con el framework del vendor | error | «framework de terceros arrastrado para leer 20 campos» | (pendiente — TestDecodificadorEsPropio) |
| payload-real-es-el-golden | los tests del decodificador usan payload capturado en vivo y versionado, no fixtures escritos a mano | error | «decodificador probado contra la spec, no contra el emisor real» | (pendiente — TestIntValueComoNumeroYComoString) |
| formato-externo-tiene-fuzz | todo decodificador de bytes ajenos tiene fuzz sembrado con el payload real | error | «cuerpo hostil sin cobertura de fuzz» | (pendiente — FuzzDecodificarLogs) |

## Changelog

- 2026-07-26 · v1.0 · Nodo fundacional (HS-28, paquete
  `stories/2026-07-24-telemetria-embebida-otel/`, decisiones D9.3/D11/D14.2 FIRMADAS). Nace de una
  medición que **refutó a la investigación** que lo precedía: `collector/pdata` cuesta +10,79 MB
  medidos contra +1,7 MB estimados (INFORME §V5), y el decodificador propio con la stdlib cuesta
  +0,49 MB. L1 = «a little copying is better than a little dependency» + superficie de
  dependencias como riesgo. L2 = el presupuesto por paquete, el decodificador propio y mínimo, el
  payload real como golden (fue lo que atrapó el `intValue` off-spec) y el fuzz obligatorio;
  consolida además los cinco descartes por peso que el árbol ya había hecho sueltos (DuckDB,
  chDB, `otlpreceiver`, Prometheus TSDB, `ocb`). El check `sin-cgo` sigue viviendo en
  `indice-desechable-jsonl-es-verdad` y no se re-declara acá. 5 checks, los 5 difieren honesto.
  `status: proposed`.
