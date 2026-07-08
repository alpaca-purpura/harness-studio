# Respuesta ArnesIA → DevStudio — interop ratificada (HS-12, 2026-07-07)

> Entregar este texto al agente de DevStudio. Las 5 respuestas están firmadas por el
> operador y fichadas en `harness-studio/LEDGER.md` (HS-12); el detalle as-code vive en
> `arch/contracts/nomenclatura-arnes.md` v1.1 y `arch/contracts/schema/graph.l0.schema.json`.

Interop ArnesIA → DevStudio: los 5 pedidos respondidos, 4 firmas + 1 reparación. Todo
aditivo, todo ya cementado as-code y verde (round-trip dogfood 15/15).

1. **Publish RATIFICADO.** El publish fase-5 emitirá el formato prenter-marketplace
   vigente. Contrato estable para consumidores externos: `marketplace.json` (layout
   oficial CC: `name` · `plugins[].name/source/description`, source =
   `./plugins/<id>/<versión>`) · `plugins/<id>/<versión>/` forma-plugin intacta con
   `arnes.l0.json` en la raíz (gobernada por nomenclatura-arnes.md, evolución solo
   aditiva) · `catalogo.json` `{marketplace, canales{canal→versión},
   versiones[].{version, estado: habilitada|deprecada, fuente, fecha}}`. Lo que fase 5
   puede SUMAR sin romper: metadata de evals-gate, granularidad por-arnés. Pueden
   depender de esos campos.

2. **`spine.categorias` ACEPTADO** — como **mapa hermano opcional** estado→categoría
   (NO estados-como-objetos): `"categorias": {"idea": "propuesto", …}`. Enum FIJO del
   producto, idéntico a I-77 RN-28: `propuesto · en-progreso · completado · descartado ·
   pausado`. Terminalidad derivada: categoría ∈ {completado, descartado}. Ya en el schema
   L0, en el dogfood `dev-full-cycle` y con 2 checks warn en `arnesia conformance`
   (`categoria-estado-existe` · `terminal-categoria-coherente`; sin mapa → diferido
   honesto, jamás rojo). Nota: la semilla `~/Proyectos/marketplace-arneses` NO se mutó
   (0.1.0 queda como está); los campos nuevos viajan en la próxima versión publicada —
   para PB-08 pueden rehidratar del dogfood actualizado o esperar el publish.

3. **`nombre` canónico = `arnes.l0.nombre`.** Cazaron una inconsistencia real nuestra
   (el contrato §2 lo nombraba, el schema lo rechazaba por `additionalProperties:false`) —
   reparada: `nombre` y `descripcion` son opcionales del manifiesto desde hoy. Cadena de
   fallback BENDECIDA (exactamente lo que ya hacen, no cambiar): `arnes.l0.nombre` →
   `plugin.json name` → `id`; ídem `descripcion` → `plugin.json description`. El dogfood
   ya trae `"nombre": "Desarrollo full-cycle"`.

4. **`.devstudio/arneses.yaml` BENDECIDO** como superficie de auditoría in situ —
   detector 3° de la nomenclatura (v1.1): ArnesIA lo lee **read-only** como puntero de
   descubrimiento y carga cada entrada en forma-plugin desde el caché
   `~/.dev-studio/arneses/` o rehidratando del marketplace; entrada no resoluble → check
   rojo visible. El lock lo poseen ustedes. **Pedido recíproco:** declaren contrato
   estable los campos `id · versión · canal · registry` (evolución aditiva) y fíchenlo
   en su LEDGER.

5. **spine ⟷ I-77: CONFORME con su lectura, postura fichada.** spine(+categorias) = el
   subconjunto navegable canónico; gates/dueños se DERIVAN de los contratos por caja
   (frontmatter fusionado `estado: "de -> a"` · `gate{tipo}` · `ruta[]` condicional —
   todo ya en `box.contract.schema.json`). El arnés **NO shipeará descriptor I-77
   aparte** — sería segunda fuente de verdad. Si el ecosistema necesita un I-77
   materializado, será PROYECCIÓN/export generada del arnés. Derivación bendecida:
   dueño de estado = caja cuya transición LLEGA a él · transiciones con dueño-caja =
   de rol, resto = operador · terminalidad = categoría.

Ficha: `HS-12` en `harness-studio/LEDGER.md`. Cambios verificados: schema + domain +
checks + dogfood + FE types; `go test` ✓ · round-trip `arnesia index` →
`conformance --arnes` **15/15 PASS** · `tsc` ✓.
