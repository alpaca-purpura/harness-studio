---
regla: archivo-durable-declara-su-esquema
version: 1.1
updated: 2026-07-26
status: proposed
ledger: HS-29
sources:
  - url: https://www.sqlite.org/pragma.html
    autoridad: oficial
    revisado: 2026-07-26
  - url: https://docs.confluent.io/platform/current/schema-registry/fundamentals/schema-evolution.html
    autoridad: oficial
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-26-conversaciones-del-panel/relevamiento-as-is.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by:      # 3 de 6. Los que faltan siguen `(pendiente — Test…)` en la Checklist y
                  # FUERA de esta lista: declarar acá un enforcer que no existe hace que el
                  # motor lo invoque, no lo encuentre y reporte `error`.
  - internal/adapters/store/migracion_test.go:TestSaveEstampaEsquemaActual
  - internal/adapters/store/migracion_test.go:TestMigracionRespaldaAntesDeEscribir
  - internal/adapters/store/migracion_test.go:TestMigracionEsIdempotente
severity: error
---

# Un archivo durable declara su esquema; uno derivado se puede tirar, uno durable jamás

## L1 · Principio (estándar de industria)

**La versión del esquema vive en el archivo, no en el código que lo lee.** Un lector que asume la
forma del dato que encuentra está adivinando, y cuando se equivoca no falla: interpreta mal.

- **El sello va adentro del archivo.** SQLite reserva 4 bytes del header del `.db` para que la
  aplicación estampe su propia versión de esquema, y explícitamente no los usa para nada: son del
  que escribe, para el que lee. *(oficial: SQLite — PRAGMA, textual: «The user_version pragma will
  get or set the value of the user-version integer at offset 60 in the database header. The
  user-version is an integer that is available to applications to use however they want. SQLite
  makes no use of the user-version itself.»)*
- **Cada versión es una entidad, y la compatibilidad se chequea antes de aceptar.** Un registro de
  esquemas versiona cada forma y valida la transición contra la anterior en vez de confiar en que
  el consumidor «se las arregle». *(oficial: Confluent Schema Registry — schema evolution, textual:
  «Each schema version gets a unique ID and incremented version number» · «Schema Registry checks
  compatibility before accepting the new version»)*

Corolario propio, y es la mitad que ninguna de las dos fuentes cubre porque las dos hablan de datos
reconstruibles: **la política legítima para un archivo DERIVADO es ilegítima para uno DURABLE.**
Tirar y reconstruir es correcto para un índice (el árbol lo regenera) y es pérdida de datos del
operador para un registro de sesiones (nadie lo regenera). El mismo repo ya tiene las dos clases y
las trataba igual.

## L2 · Realización (este árbol Go+React)

`~/.arnesia/` tiene **dos clases de archivo** y hasta hoy sólo una tenía disciplina:

| clase | ejemplos | política legítima | quién la enforça |
|---|---|---|---|
| **derivada** (se reconstruye del árbol) | `index.db` | `schema_version` + **wipe-and-rebuild** (RF-208) | [`indice-desechable-jsonl-es-verdad`](./indice-desechable-jsonl-es-verdad.md), `arch_test.go:TestSchemaVersionTriggersRebuild` |
| **durable** (nadie la regenera) | `sesiones.json`, `sesiones-archivadas.json`, el registro de arneses | `schema_version` + **migración forward-only con respaldo** · el wipe está **prohibido** | este nodo |

Realización concreta en `internal/adapters/store/`:

- **Envelope.** Todo archivo durable se escribe como
  `{"schema_version": N, "escrito_por": "<sello de build>", "escrito_en": "<RFC3339>", "<payload>": …}`.
  Un archivo cuyo primer byte no-blanco es `[` es un **v1 implícito** (la forma que `registry.go:56-59`
  escribía antes de existir este nodo): el bootstrap se detecta, no se adivina.
- **Cadena de migradores declarada**, `map[int]func(json.RawMessage) (json.RawMessage, error)`,
  aplicada de la versión leída a la actual paso a paso. Nunca un `Unmarshal` tolerante que «se lleve
  lo que entre»: un campo que no se supo migrar es un error, no un zero-value.
- **Migración de FORMA ≠ migración de DATO, y van en pasos separados.** Cambiar la estructura
  (`[]Session` → `{sesiones:[…]}`) es determinista, puro y se testea con un fixture. Recalibrar un
  valor contra una autoridad externa (p. ej. re-key de una llave contra el Portafolio) hace IO,
  **puede no poder decidir** y depende de datos que no están en el archivo. Fusionarlos hace
  intestable la parte determinista y ata el arranque del daemon a que la autoridad responda. El paso
  de dato emite **una fila de reporte por registro, se haya movido o no**, con su motivo — un valor
  que no se pudo decidir **se deja como está**, jamás se elige al azar (⇐
  [`no-aplica-no-es-cero`](./no-aplica-no-es-cero.md)).
- **Respaldo antes de escribir.** El original se copia a `<path>.v<N>-<AAMMDDHHMM>.bak` **antes** del
  primer `Save` del formato nuevo. Mismo sello `AAMMDDHHMM` que
  [`versionado.md`](../conventions/versionado.md) usa para el build: un segundo arranque no pisa el
  respaldo del primero.
- **Escritura atómica heredada.** El `Save` ya es temp-en-el-mismo-dir + `os.Rename`
  (`registry.go:77-95`); la migración no inventa otra: un crash deja **el viejo entero o el nuevo
  entero**, nunca medio archivo.
- **Idempotencia por construcción.** Con el `schema_version` actual en disco, la cadena de
  migradores está vacía y no se toca nada. Correrla dos veces es correrla cero veces.
- **Ilegible ⇒ cuarentena, jamás sobreescritura.** Un JSON inválido se renombra a
  `<path>.corrupto-<AAMMDDHHMM>`, el daemon arranca con registro vacío **y lo dice en el log con la
  ruta del respaldo**. Ninguna semilla ilustrativa (`seedSessions`) tapa una corrupción: la semilla
  sólo aplica a registro **ausente**, que es otra cosa.
- **Esquema futuro ⇒ solo-lectura, jamás degradación.** Un archivo con `schema_version` **mayor** que
  la del binario (el operador corrió una build nueva y volvió a la vieja) **no se lee a medias ni se
  pisa**: el registro entra en modo `solo-lectura` y toda escritura falla con el motivo. Degradar a
  «registro vacío» y después persistir sería destruir el archivo nuevo con el binario viejo — el
  peor modo de fallo posible y el único que no se puede deshacer.
- **`Save` propaga su error.** Hoy `persistLocked` se lo traga con un `slog.Error`
  (`session_service.go:909-911`); bajo este nodo, toda mutación **que el operador acaba de pedir**
  devuelve el fallo del filesystem y revierte su cambio en memoria. Estado en memoria y estado en
  disco no divergen en silencio. ⇐ hereda [`no-aplica-no-es-cero`](./no-aplica-no-es-cero.md): «no
  persistió» no es «persistió».

**Lo que este nodo NO dice.** No obliga a migrar hacia atrás (forward-only es la política) ni a
mantener lectores de todas las versiones vivas: obliga a **saber** con qué versión se escribió y a
**no perder** lo que no se supo leer.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| envelope-versionado | todo archivo durable que el daemon escribe lleva `schema_version` dentro del archivo; ningún lector asume la forma que encuentra | error | «el formato cambió y el lector lo interpretó mal en silencio» | internal/adapters/store/migracion_test.go:TestSaveEstampaEsquemaActual |
| migracion-forward-only-con-respaldo | una versión anterior se respalda con sello ANTES de escribir la nueva, y se migra por cadena declarada, nunca por unmarshal tolerante | error | «migración sin red: el archivo viejo ya no existe» | internal/adapters/store/migracion_test.go:TestMigracionRespaldaAntesDeEscribir |
| migracion-idempotente | correr la migración dos veces da el mismo resultado; con el esquema actual en disco no corre | error | «el segundo arranque volvió a migrar lo ya migrado» | internal/adapters/store/migracion_test.go:TestMigracionEsIdempotente |
| corrupto-se-preserva | un archivo ilegible se renombra a `.corrupto-<sello>` y el arranque lo DICE; jamás se sobreescribe ni se tapa con semilla | error | «un archivo roto se pisó con datos vacíos» | (pendiente — TestArchivoCorruptoSePreservaYSeDice) |
| esquema-futuro-no-se-degrada | un `schema_version` mayor que la del binario deja la superficie en solo-lectura con motivo; nunca se lee a medias ni se pisa | error | «el binario viejo destruyó el archivo nuevo» | (pendiente — TestEsquemaFuturoNoSePisa) |
| durable-no-se-wipea | la política wipe-and-rebuild del índice (RF-208) no se aplica a un archivo durable: ningún camino lo borra para «arreglarlo» | error | «se reconstruyó lo que nadie puede reconstruir» | (pendiente — TestDurableNuncaSeWipea) |

## Changelog

- 2026-07-26 · v1.1 · **3 de los 6 enforcers existen y corren verdes** (T9 del paquete):
  `envelope-versionado`, `migracion-forward-only-con-respaldo` y `migracion-idempotente`, sobre los
  primeros tests que `internal/adapters/store` haya tenido, con el registro **real** del operador
  (17 202 B, 5 sesiones) como fixture. Sigue `proposed`: los otros 3 —cuarentena del corrupto,
  esquema futuro y prohibición del wipe— no tienen código todavía y su celda lo dice. Gradúa a
  `enforced` con los 6, no antes.
- 2026-07-26 · v1.0 · Nodo fundacional (paquete
  `stories/2026-07-26-conversaciones-del-panel/`, etapa de arquitectura). **No inventa doctrina
  nueva: nombra la que faltaba.** El relevamiento del paquete midió el hueco: `store.Registry` hace
  `json.Unmarshal` desnudo sobre `[]domain.Session` (`registry.go:56-59`) — sin `schema_version`,
  sin envelope, sin ruta de upgrade — y `internal/adapters/store` reporta `[no test files]`. Un
  cambio de forma hace fallar el `Load` entero y **el daemon pierde el registro completo**. El
  árbol ya tenía la mitad derivada resuelta (`index/store.go#schemaVersion` + wipe-and-rebuild,
  enforced por `TestSchemaVersionTriggersRebuild`) y **este nodo declara que esa política es
  ilegal sobre la mitad durable**, que es la distinción que el repo trataba por omisión. L1 =
  sello-dentro-del-archivo (SQLite `user_version`) + versión-por-esquema-con-chequeo-previo
  (Schema Registry); el aporte propio es la partición derivada⊥durable y la prohibición de
  degradar ante un esquema futuro. **Nace `proposed` y no `enforced`**: los 6 enforcers están
  nombrados y **ninguno existe todavía** — el código llega con el paquete. Gradúa a `enforced`
  cuando los 6 corran verdes, mismo criterio que `indice-desechable-jsonl-es-verdad` v1.3. NUNCA
  pass fabricado.
