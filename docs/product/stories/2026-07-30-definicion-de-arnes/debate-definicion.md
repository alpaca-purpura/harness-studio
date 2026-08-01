# Debate — qué es un arnés (v1 → casos duros → v2) + diseño de la cadena

> Material de trabajo de DEF-D2/DEF-D3. Nada de acá es doctrina hasta la firma 🧑‍⚖️.
> Anclas firmadas que este debate NO puede contradecir: `vision.md` p1/p3/p7 ·
> terreno D0-D20 (en especial D20 multi-pipeline) · HS-12/HS-22 (identidad
> `(home,id)`, nombre canónico) · S1-D25 (término «arnés»).

## §1 · Definición v1 (propuesta 2026-07-30 — NO ratificada)

> **Arnés** = unidad de know-how operativo, instalable y vendible, que encarna **UN
> proceso de negocio para UN rol** (rol × proceso). Se empaqueta como **UN plugin
> CC** — identidad `(home, id)` — y contiene: doctrina (Base/knowledge as-code) ·
> N pipelines, uno por TIPO de paquete de trabajo · cajas de proceso (skills) con
> gates · META de enganche (rol · proceso · reporta-a · empresa) · sello
> (`arnes.l0.json`) · mejora continua medible.

Los 3 ejes que separa:

| Eje | En el modelo |
|---|---|
| Proceso end-to-end (idea→producción) | **cadena de arneses** encadenados por gates; relación en META + Galaxia |
| Rol/puesto | **el corte del arnés** — frontera de instalación, seguridad y permisos (rol/TTL) |
| Tipo de trabajo (feature/bugfix/spike) | **pipeline interno** del mismo arnés (D20; FIRMADA como DEF-D1) |

Por qué 1 arnés = 1 plugin (argumentos duros):
1. CC instala el plugin ENTERO — un plugin multi-rol le da a todos las herramientas
   de todos: muere la partición de seguridad por rol.
2. La telemetría atribuye por `plugin_id` — plugin gordo = no hay medición por
   rol × proceso (la capa Mejora queda ciega).
3. Mejora continua: versionar/publicar todo junto acopla el ciclo de release de
   roles que no se tocan.

## §2 · Lo ya firmado que esto solo APLICA (no re-decide)

- `vision.md` **p3**: «Cada arnés sirve a un rol dentro de un proceso. En empresa
  chica un rol abarca el proceso entero; en empresa grande el mismo proceso se
  parte en varios arneses (PM → dev → DevOps).»
- Terreno **D20**: multi-pipeline por tipo-de-paquete; `proceso` = mapa
  `{tipo-de-paquete → spine}`.
- **DEF-D1** (este paquete): la aplicación dev concreta, firmada.

## §3 · Casos duros para el debate (steelman de las objeciones)

- **A · Empresa chica / solo-founder.** 1 persona opera todo el proceso. ¿1 arnés
  end-to-end o cadena de 3 instalada por la misma persona? p3 dice 1 arnés — pero
  entonces «end-to-end = cadena» NO es siempre cierto, y el MISMO know-how existe
  en dos formas de corte (dev-full-cycle vs PM→dev→DevOps). ¿Son productos
  distintos? ¿Se deriva uno del otro? «Partir/fusionar un arnés» sería una
  operación de la FÁBRICA (ArnesIA) — hoy no existe ni como concepto firmado.
- **B · Unidad de venta ≠ unidad de instalación.** El cliente compra «el proceso de
  desarrollo» (la familia completa + Galaxia + mejora), no un plugin suelto. Si la
  definición no separa vendible (cadena/familia) de instalable (arnés), el
  marketing y el modelo chocan.
- **C · Ver ≠ operar.** «Que una parte la VEAN los usuarios» — visibilidad
  cross-rol no la da el plugin: la da ArnesIA (Mapa/Galaxia, read-only). La
  partición del arnés es de OPERACIÓN. Si esto no se dice explícito, alguien va a
  meter vistas ajenas dentro del plugin para «que se vea».
- **D · Mismo rol en dos procesos.** Un dev que además opera soporte → 2 arneses
  instalados en la misma máquina. Sin conflicto (identidad `(home,id)` ya lo
  soporta), pero la definición debe permitirlo sin retorcerse.

## §4 · Definición v2 (propuesta que absorbe el caso A)

> **Arnés** = paquete instalable de know-how que **posee una o más FASES de un
> proceso** y **lo opera UN rol**. 1 arnés = 1 plugin (identidad `(home,id)`).
> El **proceso** es una entidad aparte y estable (declarada en el home, §5); el
> **corte** fases→arneses lo decide cada empresa: chica = 1 arnés posee todas las
> fases; grande = cadena de arneses. Vendible = el proceso con su familia de
> arneses y su mejora continua; instalable = el arnés.

Qué gana v2 sobre v1:
- El caso A deja de ser excepción: dev-full-cycle y la cadena PM→dev→DevOps son
  **dos cortes del MISMO proceso declarado** — partir un arnés = reasignar fases,
  sin re-modelar nada.
- Separa venta (proceso/familia) de instalación (arnés) — caso B.
- «rol × proceso» de p3 se conserva: el rol sigue siendo el corte; solo se afila
  que lo poseído son FASES de un proceso con identidad propia.

## §4b · v3 — refinada con el modelo del operador (actividad · macroproceso · carpeta, 2026-08-01)

El operador aportó (round 2, `chris-input.md`): plugin POR PROYECTO/carpeta ·
actividades del puesto con procedimiento propio · macroproceso que las agrupa ·
agrupación flexible actividades→plugins según el caso. Mapeo 1:1 con lo firmado:

| El operador dijo | En el modelo | Estado |
|---|---|---|
| actividad del puesto (feature/bugfix/spike/revisar-capability) | **tipo de paquete de trabajo** | DEF-D1 FIRMADA |
| procedimiento de la actividad (flujo de pasos) | pipeline/spine del tipo | D20 |
| «muchos puntos en común, parte de un macroproceso» | base común del arnés (doctrina · knowledge · plantillas · cajas reusadas entre pipelines) | D19/D20 |
| macroproceso «Implementar software» acotado al puesto | el/los tramo(s)/fase(s) del proceso que el puesto posee | v2 §4 |
| plugin por proyecto (carpeta estructurada) | **instalación** del arnés sobre un terreno — canónico + N instalaciones | HS-22 |
| 2 roles → 2 plugins | 2 arneses instalados | v1/v2 |
| agrupar actividades en 1 o N plugins «según el caso» | grado de libertad NUEVO → criterio de corte (abajo) | NUEVO v3 |

**Jerarquía completa (v3):**

```
PROCESO end-to-end (cross-puestos, declarado en el home)   «desarrollo: idea→producción»
 └─ FASE / tramo (handoff con gate_salida verificable)     «implementar software»
     └─ ARNÉS = tramo(s) de UN puesto · 1 plugin           «arnés-dev»
         ├─ se instala POR CARPETA (terreno concreto)      «repo A · repo B» (canónico + N instalaciones)
         └─ ACTIVIDADES (tipos de paquete), cada una con
            su PROCEDIMIENTO (pipeline) sobre base común   «feature · bugfix · spike · revisar-capability»
             └─ CAJAS (skills) = pasos con gates
```

**Definición v3 (propuesta):**

> **Arnés** = el paquete de know-how de **UN puesto** sobre **UN tipo de terreno**:
> agrupa las **ACTIVIDADES** del puesto dentro de su tramo del proceso (su
> macroproceso), cada actividad con su **procedimiento** (pipeline) propio sobre
> una **base común**. 1 arnés = 1 plugin; **se instala POR CARPETA** — cada
> instalación opera un terreno concreto. El proceso end-to-end es entidad aparte
> (§5); el corte proceso→arneses y la agrupación actividades→arnés son decisiones
> de diseño regidas por el criterio de corte. Cumplís 2 roles → instalás 2 arneses.

**Criterio de corte (el «orden» que ArnesIA hace cumplir) — ¿cuándo va en OTRO plugin?**
1. ¿Otro **OPERADOR** (puesto)? → separá SIEMPRE (frontera de seguridad/permisos).
2. ¿Otro **TERRENO** (tipo de carpeta/datos)? → probablemente separá (CEO:
   outreach-CRM vs campañas-ads pueden ser terrenos distintos).
3. ¿Solo otra **actividad**, mismo operador y mismo terreno? → NO separes: es un
   pipeline más del mismo arnés. Fragmentar acá multiplica sellos/versionados/
   telemetrías para la misma persona sin ganar nada.

**Fork de vocabulario (LA decisión pendiente):** el operador dijo «las actividades
serían los "arneses"» — eso pone la palabra a nivel ACTIVIDAD. Todo lo firmado y
TODO lo construido (sello `arnes.l0.json` · identidad `(home,id)` · Portafolio ·
Mapa · `conformance --arnes` · DEF-D1 de ayer) ponen «arnés» a nivel PAQUETE/plugin.
- **Opción A (recomendada): arnés = el paquete del puesto** (nivel plugin);
  el nivel interno se llama **actividad** con su **procedimiento**. Costo cero:
  estructura idéntica a la del operador, solo se nombra el nivel interno.
- **Opción B: arnés = la actividad**; el paquete necesita nombre nuevo («traje»?
  «equipo»?). Costo: re-vocabulario de docs+código+producto+marketing enteros, y
  contradice DEF-D1 firmada ayer.

## §5 · Diseño de la cadena (DEF-D3 — PROPUESTO, sin firma)

`proceso` como entidad de primer orden, declarada en el **home** (el marketplace de
la familia — SSoT por ley anti-drift):

```yaml
# home: proceso/desarrollo-software.yaml  (PROPUESTO)
proceso:
  id: desarrollo-software
  fases:
    - id: descubrimiento
      arnes: negocio-captura      # arnés owner del tramo
      gate_salida: spec-firmada   # criterio VERIFICABLE (artefacto, no promesa)
    - id: construccion
      arnes: dev
      gate_salida: pr-aprobado
    - id: entrega
      arnes: release
      gate_salida: en-produccion
```

Reglas propuestas:
1. **Gate de salida de una fase = contrato de entrada de la siguiente.** El
   artefacto que lo satisface es verificable (mismo espíritu que PARIDAD/conformance).
2. **META se afila, no se rompe:** `proceso` pasa de string a referencia
   `proceso: <id>` + `fases: [...]` que el arnés posee. `reporta_a`/`empresa`
   quedan como están.
3. **Mapeo fases→arnés es N:1-friendly:** un arnés puede poseer TODAS las fases
   (empresa chica). La cadena queda declarada igual; crecer = reasignar.
4. **Galaxia renderiza la cadena desde el home; telemetría joinea por `proceso.id`**
   (habilita medir el proceso completo cruzando arneses — insumo directo de la
   capa Mejora).
5. La visibilidad cross-rol (caso C) vive ACÁ: el Mapa/Galaxia lee el proceso
   declarado, no los plugins ajenos.

Abierto a decidir en la firma: ¿el `proceso/<id>.yaml` vive SOLO en el home o cada
`arnes.yaml` embebe una copia de su tramo (offline-first)? ¿versionado del proceso
independiente del semver de cada arnés?

## §6 · gentle-ai (hallazgo DEF-D4)

github.com/Gentleman-Programming/gentle-ai — analizado 2026-07-30. NO es plugin CC:
configurador de ecosistema para 14 agentes (Claude Code, Cursor, OpenCode…) con
memoria persistente (Engram), SDD opcional (Explore→…→Archive), review por
receipts (RDD), skill registry con scope global/workspace. **Sin modelo de rol ni
proceso de negocio** — outcome-first (pedís resultado, el sistema rutea). No
contiene «múltiples arneses»: no contiene ninguno según nuestra definición.
Compite con el harness crudo, no con el arnés. A evaluar para robar: instalación
modular por scope · ergonomía del receipt único validado por todas las compuertas.
