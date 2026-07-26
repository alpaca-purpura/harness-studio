---
regla: cifra-viaja-con-su-confianza
version: 1.0
updated: 2026-07-26
status: proposed
ledger: HS-28
sources:
  - url: https://www.w3.org/TR/prov-dm/
    autoridad: oficial
    revisado: 2026-07-26
  - url: https://www.bipm.org/en/committees/jc/jcgm/publications
    autoridad: oficial
    revisado: 2026-07-26
    # Alcance HONESTO de lo verificado: la página es el índice de publicaciones del JCGM y sí
    # lista `JCGM 100:2008` con su título («Evaluation of measurement data — Guide to the
    # expression of uncertainty in measurement»). El PDF NO se abrió ⇒ el principio se apoya en
    # el título y el alcance declarado del documento, no en una cita textual de su cuerpo.
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/INFORME.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by: []
severity: error
---

# Ninguna cifra viaja sola: lleva cómo se obtuvo y, si hay dos fuentes, las dos

## L1 · Principio (estándar de industria)

Dos disciplinas dicen lo mismo desde lados opuestos:

- **Procedencia del dato.** Un valor sin su linaje —de dónde salió, qué actividad lo produjo, en
  qué se basó— no es auditable: se puede creer o no creer, pero no verificar. *(oficial: W3C
  PROV-DM, textual: «Provenance is information about entities, activities, and people involved in
  producing a piece of data or thing, **which can be used to form assessments about its quality,
  reliability or trustworthiness**»)*
- **Incertidumbre de la medición.** La metrología tiene un documento internacional dedicado a
  **cómo se expresa la incertidumbre de un resultado**: la GUM (`JCGM 100:2008`, *Evaluation of
  measurement data — Guide to the expression of uncertainty in measurement*). Que exista y sea
  normativo es el punto: la incertidumbre no es un adorno del reporte, es parte de lo que se
  reporta. *(oficial: JCGM/BIPM — índice de publicaciones. ⚠️ Se verificó el listado y el título de
  la edición; el PDF no se abrió, así que acá no se cita texto de su cuerpo.)*

Y una tercera, de ingeniería, para el caso en que hay dos formas de obtener el mismo número:
**testing diferencial**. Dos implementaciones independientes del mismo cálculo son un oracle
mutuo — la divergencia entre ellas es una señal de primera clase, no un empate que alguien
resuelve eligiendo su favorita. *(experto: differential / N-version testing)*

## L2 · Realización (este árbol Go+React)

El eje ya existe en el árbol para el dato de un nodo: **`domain.Procedencia`**
(`medido`·`estimado`·`declarado`·`inferido`·`no-declarado`) es exactamente esta doctrina aplicada
al grafo, y `% contexto` se computa y se **etiqueta como métrica derivada**
(`conductor-no-parsea-jsonl` check `ctx-derivado-etiquetado`, que sigue viviendo en su nodo — acá
no se re-declara). Lo que faltaba era el segundo eje y el contraste.

**Los dos ejes son distintos y conviven.** `procedencia` dice *cómo se obtuvo el valor*;
**`atribucion`** dice *a qué unidad de trabajo se le asignó*. Un `api_request` es
`procedencia: medido` y a la vez puede ser `atribucion: por-hash`. Confundirlos es lo que hace que
una pantalla muestre con seguridad algo que se dedujo.

- **Toda cifra de dinero de la capa «Mejora» lleva su `atribucion`**
  (`exacta`·`por-hash`·`por-proceso`·`sin-dato`), y la UI lo hace visible: subrayado punteado en
  todo lo que no sea `exacta`, con el motivo al hover. Verificado por qué hace falta: el runtime
  **redacta** los nombres de plugin de terceros a `third-party` y nuestros arneses caen ahí
  (INFORME §V3) — la atribución se recupera por huella, y eso hay que decirlo.
- **La confianza de un agregado es la MÍNIMA de sus partes, nunca el promedio.** Sumar tres
  turnos `exacta` y uno `por-hash` da un total `por-hash`. Promediar confianzas produce un número
  que se presenta mejor de lo que es.
- **Lo que no se pudo atribuir no entra al total.** Un evento `sin-dato` se guarda y se cuenta en
  la cobertura, pero **no suma** al gasto de ningún arnés: sumarlo sería atribuir por adivinanza.
- **Los DOS costos se persisten**: `costo_reportado_micros` (lo que dijo el runtime) y
  `costo_calculado_micros` (lo que dice nuestro catálogo). Eso convierte el test de paridad en algo
  que **corre en producción y gratis**: si divergen, o el catálogo está viejo o el runtime cambió
  su tarifa, y las dos cosas son información. No se elige uno en silencio.
- **Lo estimado se declara.** `cost.usage`/`cost_usd_micros` está documentado como *"Estimated
  cost"*, no facturación: la UI no puede presentarlo como plata gastada sin decirlo. El campo
  `estimado` del resumen es `true` por construcción mientras la fuente sea esa.
- **Precedente hermano en el árbol:** la **deriva** se calcula por hash de contenido y *nunca* por
  el `version` declarado (`portafolio-identidad-y-deriva-honesta`). Es la misma familia: cuando
  un tercero declara un número que nosotros podemos recomputar, el declarado no gana por default.

**Por qué este nodo fusiona «la cifra lleva su confianza» y «el doble costo es un oracle»**, que
se evaluaron como dos boundaries: son las dos mitades de una misma oración —*el número no viaja
solo*— y comparten la superficie de enforcement (la struct del número y su DTO). Como nodos
separados, uno diría «lleva su confianza» y el otro «lleva su contraste», y el segundo quedaría
sin L1 propio.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| cifra-lleva-confianza | todo campo de una cifra derivada/atribuida convive con su campo de confianza en la misma entidad y llega así al cliente | error | «un número en pantalla sin decir cómo se obtuvo» | (pendiente — TestCifraLlevaConfianza) |
| confianza-no-se-promedia | la confianza de un agregado es la MÍNIMA de sus partes | error | «un total `exacta` compuesto de partes deducidas» | (pendiente — TestConfianzaDeAgregadoEsLaMinima) |
| sin-dato-no-suma | lo que no se pudo atribuir se cuenta aparte y jamás suma al total de una unidad de trabajo | error | «gasto atribuido por adivinanza» | (pendiente — TestSinDatoNoSumaAlTotal) |
| dos-fuentes-dos-valores | cuando un tercero reporta un número que también podemos calcular, se persisten los dos; ninguno pisa al otro | error | «se eligió una fuente en silencio» | (pendiente — TestDobleCostoSePersisteEntero) |
| divergencia-es-visible | si las dos fuentes difieren más allá del umbral declarado, la divergencia se muestra; no se promedia ni se oculta | error | «divergencia entre fuentes escondida al usuario» | (pendiente — TestDivergenciaDeCostoEsVisible) |
| estimado-se-declara | una cifra que el emisor documenta como estimada se presenta como estimada, nunca como hecho | error | «un estimado presentado como facturación» | (pendiente — story-test `barra-con-disclaimer`) |

## Changelog

- 2026-07-26 · v1.0 · Nodo fundacional (HS-28, paquete
  `stories/2026-07-24-telemetria-embebida-otel/`, decisión D16.2 FIRMADA — el campo
  `atribucion_confianza`, que la investigación registra como algo que **ninguna plataforma del
  estado del arte propone**). L1 = procedencia (W3C PROV) + incertidumbre de medición (GUM) +
  testing diferencial. L2 = el segundo eje (`atribucion`) junto al que ya existía
  (`domain.Procedencia`), la confianza mínima en los agregados, el doble costo como oracle de
  paridad corriendo en producción, y el «estimado» declarado (INFORME §H4). **Fusiona dos
  candidatos** que se evaluaron por separado; la justificación está en L2. 6 checks, los 6
  difieren honesto: el módulo `telemetria/` no existe. `status: proposed`.
