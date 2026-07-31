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
