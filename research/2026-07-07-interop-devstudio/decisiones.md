# Decisiones — interop DevStudio ⟷ ArnesIA (HS-12)

> Registradas EN EL MISMO TURNO de la firma (disciplina METODOLOGIA §10).
> Contexto: DevStudio (repo `~/Proyectos/dev-studio`, ficha DH-18/PB-25) adoptó el
> estándar ArnesIA sin pedir cambios de formato, lo verificó en vivo consumiendo el
> dogfood `dev-full-cycle`, y trajo 5 pedidos aditivos. Todos respondidos; 4 firmas del
> operador (2026-07-07) + 1 reparación de consistencia interna.

## P1 — Publish fase-5 emite el formato prenter-marketplace vigente — FIRMADA ✅

**Decisión:** ratificado. El publish real (hoy stub `internal/adapters/publish/publisher.go`)
emitirá el formato ya en producción: `.claude-plugin/marketplace.json` + `catalogo.json` +
`plugins/<id>/<versión>/` (forma-plugin intacta con `arnes.l0.json` en la raíz).

**Campos contrato ESTABLE para consumidores externos:**
- `marketplace.json` — layout oficial CC (no lo poseemos; estabilidad heredada de CC):
  `name` · `plugins[].name` · `plugins[].source` (`./plugins/<id>/<versión>`) ·
  `plugins[].description`.
- `plugins/<id>/<versión>/` — forma-plugin gobernada por `nomenclatura-arnes.md` (evolución
  SOLO aditiva; el manifiesto `arnes.l0.json` valida contra `graph.l0.schema.json`).
- `catalogo.json` — extensión nuestra: `marketplace` · `canales{canal→versión}` ·
  `versiones[].{version, estado: habilitada|deprecada, fuente, fecha}` (formato ya vivo en
  `prenter-marketplace`; `instancia` es opcional/de la casa).

**Compromiso:** cambios futuros a las tres superficies = solo aditivos. Lo que el publish
fase-5 puede SUMAR (jamás romper): metadata de evals-gate en `catalogo.json`, granularidad
por-arnés. **Why:** el formato ya está en producción (KIT-06) y DevStudio lo verificó en
vivo; romperlo rompería al primer consumidor externo del ecosistema.

## P2 — `spine.categorias`: mapa opcional, enum fijo de 5 — FIRMADA ✅

**Decisión:** aceptado, encoding = **mapa hermano opcional** `spine.categorias`
(estado→categoría), NO estados-como-objetos. Enum FIJO del producto, idéntico a I-77
RN-28 (patrón Azure state categories): `propuesto · en-progreso · completado ·
descartado · pausado`. Terminalidad derivada: categoría ∈ {completado, descartado}.

**Why mapa y no objetos:** 100 % aditivo — `estados` sigue `[]string`, cero impacto en
loader/consumidores existentes; la forma-objeto de I-77 vive en el descriptor de DevHub,
no en nuestro manifiesto. **No viola agnosticismo:** los estados siguen siendo dato
per-arnés; la categoría es capa semántica agnóstica al proceso (mismo estatus que `clase`).

**Cementado (mismo turno):** `graph.l0.schema.json` (spine.categorias + enum) ·
`domain.Categoria` (+`Valid()`/`Terminal()`) · `Spine.Categorias` · 2 checks warn en
`VerificarSpine` (`categoria-estado-existe` · `terminal-categoria-coherente`, diferido
honesto si el mapa falta) · dogfood `dev-full-cycle` con categorías · FE `types.ts` ·
round-trip `arnesia index` → `conformance --arnes` = **15/15 PASS** (13 previos + 2).

## P3 — `nombre`/`descripcion` canónicos — REPARACIÓN (sin firma nueva)

**Hallazgo de DevStudio (correcto):** nomenclatura §2 (v1 FIRMADA) nombra `nombre` en el
manifiesto pero `graph.l0.schema.json` no lo tenía y es `additionalProperties:false` — un
manifiesto CON `nombre` habría fallado el gate G1. Inconsistencia interna nuestra.

**Resolución:** schema reparado al contrato ya firmado — `nombre` y `descripcion`
opcionales en `arnes`. **Campo canónico para pintar = `arnes.l0.nombre`**; cadena de
fallback bendecida: `arnes.l0.nombre` → `plugin.json name` → `id` (ídem `descripcion` →
`plugin.json description`). Es exactamente lo que DevStudio ya hace — bendecido, no
cambiar. Dogfood ahora lo trae (`"Desarrollo full-cycle"`).

## P4 — `.devstudio/arneses.yaml` bendecido como superficie de auditoría — FIRMADA ✅

**Decisión:** bendecido como **detector 3°** de la nomenclatura (v1.1, aditivo): lock
presente → N arneses complementarios, cada uno cargado en **forma-plugin** resuelta vía
caché `~/.dev-studio/arneses/` o rehidratación del marketplace. El lock es puntero de
descubrimiento **read-only** para ArnesIA — no es un arnés; la unidad reconocible sigue
siendo la forma-plugin. Entrada no resoluble → check rojo visible (jamás omitida).

**Pedido recíproco a DevStudio:** declarar contrato estable los campos del lock —
`id` · `versión` · `canal` · `registry` de origen (evolución aditiva).

**Why:** el patrón DevStudio (forma-plugin intacta + lock, JAMÁS fundir en `.claude/`) es
MÁS fiel a METODOLOGIA §9 (② no se escribe en ③) que la forma-instalada; el roster viaja
por git y cualquier máquina rehidrata.

## P5 — spine ⟷ I-77: se PROYECTA, no se declara — FIRMADA ✅

**Postura fichada:** conforme con la lectura de DevStudio. `spine` (+`categorias`) es el
**subconjunto navegable canónico** del proceso del arnés; gates y dueños se **derivan de
los contratos por caja** (frontmatter fusionado: `estado: "de -> a"` · `gate{tipo,…}` ·
`ruta[]` condicional — todos ya en `box.contract.schema.json`).

**El arnés NO shipeará un descriptor I-77 aparte.** Sería segunda fuente de verdad
(contra aditivo-sin-pérdida y el SSoT del contrato fusionado). Si el ecosistema pide un
I-77 materializado, será una **PROYECCIÓN/export generada del arnés** (como
`.graph.json`), jamás una declaración paralela.

**Derivación bendecida:** dueño de un estado = la caja cuya transición LLEGA a él ·
gates = `gate{}` de cada caja · transiciones con dueño-caja = las de rol; el resto =
operador · terminalidad = categoría.

## Cambios materiales (todos en este commit)

- `arch/contracts/schema/graph.l0.schema.json` — `nombre`/`descripcion` + `spine.categorias`.
- `arch/contracts/nomenclatura-arnes.md` — v1.1 (detector lock · nombre canónico · categorias).
- `internal/domain/graph.go` — `Arnes.Nombre/Descripcion` · `Categoria` (5 consts,
  `Valid`/`Terminal`) · `Spine.Categorias`.
- `internal/domain/conformance.go` — checks `categoria-estado-existe` ·
  `terminal-categoria-coherente` (warn, diferido honesto sin mapa).
- `internal/domain/conformance_test.go` — `TestVerificarSpineCategorias`.
- `dogfood/dev-full-cycle/arnes.l0.json` + `dogfood/dev-full-cycle.graph.json` — nombre ·
  descripcion · categorias (en sync, fixture verbatim).
- `web/src/entities/arnes/model/types.ts` — `Categoria` · `Spine.categorias` ·
  `Arnes.nombre/descripcion`.
- NO tocado: la semilla `~/Proyectos/marketplace-arneses` (la versión `0.1.0` publicada no
  se muta in place; los campos nuevos viajan en la próxima versión publicada).

**Verificación:** `go build` ✓ · `go test ./internal/...` ✓ · round-trip
`arnesia index dogfood/dev-full-cycle` → `conformance --arnes` = **15/15 PASS** ·
`tsc --noEmit` ✓ · biome ✓.
