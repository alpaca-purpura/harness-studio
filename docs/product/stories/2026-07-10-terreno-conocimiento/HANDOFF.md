# HANDOFF — Modelo de Dimensiones/Terreno de ArnesIA (continuidad de co-diseño)

> **Este archivo es la red de seguridad para compactar la conversación.** Léelo COMPLETO al retomar,
> junto con `decisiones.md` (**D0-D19**). Estado: **co-diseño en curso, NADA firmado, nada commiteado.**
> Fecha: 2026-07-10. Paquete: `docs/product/stories/2026-07-10-terreno-conocimiento/`.
>
> ⚠ **MODELO ACTUALIZADO A D18-D19 — este HANDOFF (§1-§7) puede tener texto stale.** La fuente de verdad
> del modelo es `decisiones.md` D18 (**regla de las 3 caras**: NORMA→dimensión · PASO→proceso · ARTEFACTO→WIP)
> y **D19** (set consolidado: **11 canónicas + 2 overlays**). Cambios clave que abajo NO están reflejados:
> `requisitos` migró Propósito→Producto · nació `necesidad` en Propósito · `economia`→overlay · NUEVA
> `gestion-trabajo` (schema de WIP) en Organización. Visual al día = `estructura-terreno.html` (v4).
> Las "observaciones pendientes" del user YA se cocearon (dieron D18-D19). Siguiente: refresco del mapa-UI
> DESTINO + diseñar `arnes.yaml`.

---

## 0 · El norte (por qué estamos acá)

Realineación estratégica: todo el esfuerzo va al **core del ciclo de vida del arnés** (medición/observabilidad
→ Fase 2, congelado). **ArnesIA = fábrica/FORJA de "arneses-as-process" (harnesses de proceso para agentes de
IA), AGNÓSTICA AL RUBRO.** Desarrollo de software es UN ejemplo, jamás la norma (contabilidad, finanzas, legal
= mismo motor, otro contenido). Detonante: hay un cliente real esperando; forjar/instalar/arreglar arneses YA.

Estábamos por co-diseñar el mapa-UI "DESTINO", pero el user abrió una pregunta más profunda: **¿el modelo tiene
cimientos claros para que, en cualquier rubro, se sepa intuitivamente dónde va cada cosa?** Esa pregunta nos
llevó a definir el MODELO metodológico completo (abajo), fundado en research. El mapa-UI se rehace DESPUÉS.

---

## 1 · Vocabulario canónico (naming — con respaldo, D1/D11/D14)

| palabra | qué es | respaldo |
|---|---|---|
| **Terreno** | el landscape COMPLETO (los 4 territorios juntos) | Wardley/TOGAF *Architecture Landscape* |
| **Dimensión** | cada eje que el proyecto define y puebla | OLAP/dimensional-modeling |
| **Conocimiento** | rules+skills co-locadas que IMPLEMENTAN una dimensión (nacen de sus decisiones) | — |
| **Estado / salud** | vacío/parcial/lleno de una dimensión o paquete | ISO 42010 *view* + Essence *alpha states* |
| **WIP** | el trabajo vivo (instancias con estados) — 4º territorio | Essence `Work` alpha |
| **capability** | detalle de lo implementado (unidad); CIERRA el paquete | (ya existe en el repo, SSoT funcional) |
| **paquete de trabajo** | la unidad que VIAJA por el proceso (historia/outcome/engagement) | value-stream flow item |

Regla dura **(E, D11):** los nombres canónicos son **internos del motor**; **cada arnés DEBE renombrarlos al
lenguaje de su rubro** (`{id_canónico, label_rubro}`). El motor traza por `id`; el usuario ve el `label`.

---

## 2 · ArnesIA = FORJA (no marketplace de packs ajenos) — D10

**NOSOTROS forjamos cada arnés** con nuestra metodología. El dimension-set le **pertenece al arnés**
(inversión de propiedad); como lo creamos determinísticamente, **lo sabemos leer/mantener/cosechar**. El
marketplace distribuye NUESTROS arneses (solo arneses propios — seteamos el estándar). NO es un pack opinado de
terceros que el usuario compra.

**3 capas:** `Motor (frame fijo + forjador)` → `Arnés (forjado por nosotros, dimension-set del rol)` →
`Proyecto (del usuario: instancias, sello/deriva vivo, cosecha-back)`.

**Forjador (D12, parte del motor + dogfood):** forja con NUESTRA metodología (fases→cajas→gates→PARIDAD),
**rellenando PLANTILLAS, NUNCA de cero, DETERMINISTA**. **Gates de forja** validan que el arnés cumpla todo
(ver gate de completitud). ArnesIA se forja con ArnesIA.

---

## 3 · Los 4 TERRITORIOS + Calidad (el corazón — D14/D15/D16)

Agrupados por las 3 areas-of-concern de Essence + un 4º de ejecución. Cada dimensión lleva tag **Zachman**
(What/How/Where/Who/When/Why) = rubro-agnóstico.

```
PROPÓSITO      (Essence·Customer)     por qué · para quién · qué se pide          [DEFINICIÓN]
   encargo [Why] · stakeholders [Who] · requisitos [What]
PRODUCTO       (Essence·Solution)     lo construido durable = Σ capabilities       [DEFINICIÓN]
   producto [How/What] (el rubro EXPANDE o AGREGA) · datos [What] · contexto [Where]
ORGANIZACIÓN   (Essence·Endeavour)    cómo trabajamos: MÉTODO + PLANTILLAS         [DEFINICIÓN]
   proceso [When/How] (paso-a-paso + plantillas de artefactos, FIJO) · equipo [Who] ·
   forma-trabajo [How] · economia [Why]
WIP            (Essence·Work alpha)   el trabajo VIVO (instancias con estados)     [INSTANCIA]
   paquetes (outcome→paquete→sub) con estados · planificación · artefactos LLENADOS · foto auto-gen → done/
CALIDAD (OVERLAY transversal)         ISO 25010 / catálogo por-rubro → produce capabilities NO-funcionales
```

**Notas clave:**
- **10 dimensiones canónicas** (D13): las de Propósito(3) + Producto(3) + Organización(4). El motor las EXIGE
  = **gate de completitud**. `economia` y `contexto` incluidas (máximo rigor).
- **`producto`** (D13): el rubro ELIGE expandir en sub-nodos (dev: arquitectura→FE/BE) o agregar hermanas
  (dev suma ux·tecnologías·concurrencia). El gate solo exige las canónicas presentes.
- **Calidad = overlay, NO dimensiones peer** (D3): permea todo; se proyecta como checks. Excepción: Seguridad ·
  Testing · Compliance conservan carpeta propia (peso operativo). Al ratificarse → **capabilities no-funcionales**.
- **WIP (D15):** 4º territorio de naturaleza **instancia** (no definición). **Regla anti-cajón:** toda carpeta de
  trabajo lleva **estado + criterio de cierre (ratifica capability) + pasa a `done/`**. Las **plantillas** de
  artefactos viven en Organización/proceso; los **artefactos llenados** viven en WIP.
- **capability = cierre + constitución (D16):** un paquete CIERRA al ratificar su(s) capability(es). **Producto =
  Σ capabilities (funcionales + no-funcionales).** Frontera WIP↔Producto = la capability.

**Fix del concepto "épica" (D17, lección de Vitalia):** UNA jerarquía limpia outcome→paquete→sub (framing por
OUTCOME, Opportunity-Solution-Tree, no épica-cajón-de-output) · release/fase = **TAG ortogonal**, jamás en el
slug/id · snapshot-vivo(auto-gen, DO-NOT-EDIT) ≠ bitácora(LEDGER), nada de prosa-log en YAML · pipeline que
ESCALA al appetite (Shape-Up), no forzar un pipeline fijo en todo paquete.

---

## 4 · Mecánica del arnés (robada + endurecida — D5-D9)

- **Conocimiento co-locado (D5):** cada dimensión tiene `<dimensión>/knowledge/` (rules+skills que nacen de sus
  decisiones). Traza bidireccional (`codigo-traza-a-capability`). NO es región/territorio aparte — es capa.
- **Nav anti-token (D6/D9):** Index-First + hojas atómicas + **punteros auto-derivados del código** (SBOM ·
  storybook `index.json` · manifiestos · tree-sitter/PageRank estilo aider repo-map) verificados en CI. 3-4
  saltos: `CLAUDE.md → docs/terreno/INDEX → <dimensión>/INDEX → hoja → puntero-a-código`.
- **Schema hoja de conocimiento (D9):** frontmatter `{id, tipo(Diátaxis), dimension, area, zachman, label_rubro,
  resumen(L1→INDEX), activacion(always|glob|demand|manual), globs, punteros_auto(GENERADOS), budget_tok}` +
  cuerpo atómico ≤~400 tok. Sin historia (→LEDGER), sin cifras a mano (se generan).
- **Packaging (D7):** cada dimensión = un `SKILL.md` (frontmatter `description` selector + cuerpo + refs lazy)
  → co-locación instalable. Publicar `AGENTS.md` como fachada multi-herramienta.
- **Sello/deriva/init/doctor (D8, robado de cruft/copier/projen):**
  - **Sello** = `.copier-answers.yml`/`.cruft.json` endurecido: fuente + versión canónica pineada + inputs +
    **firma de fábrica** + **hash por-archivo** (baseline). Vive en la raíz del arnés instalado.
  - **Deriva** = three-way merge de `copier update` → clasifica cada archivo: `original` · `modificado-usuario`
    · `generado`(marker) · `no-reconocido`(doctor jamás lo toca).
  - **Doctor** = re-síntesis idempotente (projen) por marker de propiedad; **propiedad PARCIAL** (solo lo que el
    proceso exige, NO el proyecto entero).
  - **Drift-gate** = anti-tamper + `cruft check` en CI = honestidad automática (N2, HS-21).
  - **★ Loop-forward (aporte genuino, nadie lo tiene):** invierte cruft — la fábrica **COSECHA** el diff local,
    lo publica canónico, las instancias corren `update`. "La deriva es la señal."

---

## 5 · Carpetas (propuesta as-code)

```
project.config.yaml · CLAUDE.md · AGENTS.md · arnes.yaml     # seam · router · fachada · dimension-set (a diseñar)
docs/terreno/                    # DEFINICIÓN (3 territorios)
  proposito/{encargo,stakeholders,requisitos}/
  producto/{producto,datos,contexto}/     # producto/ ← MIGRA docs/architecture/ (dogfood)
  organizacion/{proceso(spine.yaml,raci.yaml,handoffs/,plantillas/),equipo,forma-trabajo,economia}/
  calidad/  (+ seguridad/ testing/ compliance/ con hojas)
  # cada dimensión: INDEX.md + hojas atómicas + knowledge/{rules,skills}
docs/wip/                        # EJECUCIÓN (4º territorio) — instancias
  activo/<outcome>/<paquete>/    # artefactos llenados + estado
  done/<outcome>/                # cerrados (ratificaron capability)
  INDEX.md                       # foto viva AUTO-GENERADA (DO-NOT-EDIT)
docs/product/capabilities/       # PRODUCTO durable: Σ capabilities func+no-func — SSoT (QUEDA)
.claude/                         # rules/skills OPERATIVAS del harness (gates) — kit
sello.yaml + .baseline-hashes    # manifiesto firmado + hash por-archivo
```
Decisión tomada: `docs/terreno/<dim>/` **absorbe** `docs/architecture/`; ArnesIA se **dogfoodea**;
`capabilities/` queda como SSoT.

---

## 6 · Fundamento metodológico (marcos que respaldan — detalle en los research-*.md)

- **ISO/IEC/IEEE 42010** (viewpoint/perspective/concern) + **Rozanski&Woods** + **arc42** + **C4** → estructura
  vs calidad; viewpoints.
- **SWEBOK v4** (18 KAs) + **SEMAT/Essence** (7 alphas × 3 areas-of-concern: Customer/Solution/Endeavour; Work
  alpha con estados) → las áreas + el 4º territorio WIP + estados/salud.
- **Zachman** (6 interrogativas What/How/Where/Who/When/Why) → agrupador rubro-agnóstico (los tags).
- **ISO 25010:2023** (9 atributos) + FURPS+ + AOP → calidad = overlay transversal.
- **Naming:** OLAP(dimension) · Wardley/TOGAF(landscape) · DDD(subdominio) → terreno≠dimensión.
- **Arneses-as-code:** Kiro steering (activación always|glob|demand|manual) · OpenSpec (delta-first=sello/deriva)
  · BMAD (sharded + module-help.csv; EVITAR su 80%-tokens-re-inyectando-standards) · Claude Code plugins/skills
  (progressive disclosure) · **cruft/copier/projen** (sello/deriva/doctor) · aider repo-map + llms.txt (punteros
  auto-derivados) · Backstage/golden-path.
- **PM/planning:** jerarquía initiative→epic→story→task; crítica de épica → **Opportunity Solution Tree** +
  **Shape Up** (appetite/bet/no-gos). Universal: legal=matter · consultoría=engagement · finanzas=deal.
- **Vitalia** (`/home/chalreme/Proyectos/luana-vitalia/vitalia`): PRUEBA el modelo (método externalizado al kit,
  story con 10-estados+WIP-caps, capability durable=puente, foto auto-generada DO-NOT-EDIT). Muestra el modo de
  falla a EVITAR: la "épica" = 4 agrupadores compitiendo (outcome-muerto+release+fase-en-slug+areas/modules);
  estado-vivo mezclado con bitácora en YAML; lifecycle desigual (2 stories 48 archivos, resto 2).

---

## 7 · Estado: decidido vs abierto

**DECIDIDO (D0-D17, ver `decisiones.md`):** distinción Terreno/Dimensión/Conocimiento · naming triada + WIP ·
ArnesIA=forja · forjador determinista+gates+dogfood · 4 territorios (Propósito·Producto·Organización·WIP) + 10
canónicas + calidad-overlay · capability=cierre + Producto=Σcapabilities · fix-épica · sello/deriva/doctor/
loop-forward · conocimiento co-locado · nav low-token · carpetas (`docs/terreno/` absorbe architecture).

**ABIERTO / PRÓXIMO:**
1. ⚠ **Observaciones pendientes del user** (dijo que tiene más) — PREGUNTAR primero.
2. **Detalle menor:** en dev, Organización se renombró "Ejecución" pero choca con WIP (la ejecución real).
   Propuesta a validar: Organización dev = "Método/Proceso", WIP dev = "Ejecución". Sin resolver.
3. **Diseñar `arnes.yaml`** (linchpin as-code) — escribirlo COMO instancia del arnés-dev (dogfood): declara el
   dimension-set {id,label,área,zachman,receta,activación} + proceso + roles + calidad + WIP-estados +
   conocimiento. Fuerza el determinismo. Luego estresarlo con un `arnes-finanzas.yaml` mínimo (agnosticismo).
4. **Dogfood:** forjar el arnés-dev (Vitalia) para validar 10 canónicas + WIP + gate end-to-end.
5. **Rebuild** `mockups/arnesia-mapa-destino.html` como superset del baseline (SSoT=Storybook) con el modelo v3.
   ⚠ El draft actual de ese mockup está STALE (modelo viejo conflacionado).
6. **Scaffold real** de `docs/terreno/` + `docs/wip/` + migrar `docs/architecture/`.
7. Firmar → registrar en `mockups/INDEX.md` → repriorizar BACKLOG Fase-1/Fase-2.

---

## 8 · Método de trabajo a preservar (cómo venimos co-diseñando)

- **Disciplina de paquete (§10 METODOLOGIA):** todo vive en este paquete `stories/2026-07-10-terreno-conocimiento/`.
  Toda decisión conversada → `decisiones.md` EN EL MISMO TURNO. Retomar sesión = leer este HANDOFF + INDEX.
- **Patrón que funcionó:** para preguntas de fondo → lanzar **subagentes de research** (web + local) en paralelo
  (background) → sintetizar "qué robar/qué evitar" → **decidir con el user vía preguntas concretas** → loguear →
  actualizar el visual → iterar. NO re-hacer la research (ya está en los 4 `research-*.md`).
- **Honestidad del repo:** nada "listo" sin verificar; gaps VISIBLES; cifras se GENERAN, no se teclean.
- **Anti-drift UI:** SSoT del UI = Storybook; `mockups/*.html` = snapshots derivados; toda propuesta = superset
  estricto del baseline, no pisar vocabulario L0. (Este paquete es co-diseño de MODELO, aún no toca el mapa-UI.)
- **No inventar nombres:** pegarse a marcos (42010/Essence/Zachman/ISO25010/DDD). Todo determinista + gated.

---

## 9 · Archivos del paquete

- `HANDOFF.md` — **este archivo** (leer primero al retomar).
- `INDEX.md` — «retomar aquí» corto.
- `decisiones.md` — **D0-D17** (la fuente de verdad de las decisiones).
- `estructura-terreno.html` — **mapa visual v3** (4 territorios + toggle rubro Canónico/Dev/Finanzas + WIP/
  estados/done + flujo + carpetas). Abrir en browser para revisar.
- `research-terrenos.md` · `research-marcos-naming.md` · `research-arneses-as-code.md` ·
  `research-proyecto-ejecucion.md` — el detalle con fuentes (NO re-investigar).
