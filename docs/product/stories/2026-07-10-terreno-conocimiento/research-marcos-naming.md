# Research — Marcos metodológicos: naming + set completo de dimensiones (Wave 1 · 2026-07-10)

> Insumo de co-diseño. Objetivo: pegar el modelo de "terrenos/dimensiones" a terminología
> ESTABLECIDA (no inventar), descubrir dimensiones faltantes, y resolver el agrupador
> rubro-agnóstico. 5 investigaciones web. Fuentes al pie de cada bloque.

---

## A · ISO/IEC/IEEE 42010 + Rozanski & Woods + 4+1 + arc42 (viewpoints/concerns)

**Vocabulario 42010 (normativo):**
- **Concern** = interés de un stakeholder (performance, seguridad, costo, evolución). Es el *interés*, no el artefacto.
- **Viewpoint** = convención/plantilla para construir UNA vista; *enmarca (frames)* concerns + define model kinds. Molde reusable.
- **View** = representación del sistema gobernada por exactamente un viewpoint; *atiende (addresses)* los concerns. **= la instancia por-proyecto.**
- **Model kind** · **Architecture Description** (colección de vistas).
- Mnemónico: *viewpoint FRAMES concerns → view (por-sistema) ADDRESSES concerns.*
- "Perspective" NO es 42010 — es extensión de Rozanski & Woods (calidad transversal).

**Rozanski & Woods — 7 VIEWPOINTS (estructura):** Context · Functional · Information · Concurrency · Development · Deployment · Operational.
**R&W — 10 PERSPECTIVES (calidad transversal):** Security · Performance&Scalability · Availability&Resilience · Evolution · Accessibility · Development-Resource · Internationalization · Location · Regulation · Usability.
**Kruchten 4+1:** Logical · Process · Development · Physical · +Scenarios.
**arc42 (12):** intro/metas · restricciones · contexto&alcance · estrategia · building-blocks · runtime · deployment · conceptos-transversales · decisiones · req-calidad · riesgos/deuda · glosario.

**Faltan en el borrador (presentes en el estándar):** Context/Scope · **Concurrency** (crítico agentes) · Evolution · Location · Internationalization · Accessibility · Development-Resource.
**Recomendación del agente:** partir el set en **viewpoints (estructurales, producen artefacto propio)** vs **perspectives (transversales, permean todas las vistas)** — es la distinción "partition vs integrate" de R&W.
Fuentes: quality.arc42.org/standards/iso-42010 · en.wikipedia.org/wiki/ISO/IEC_42010 · informit.com Viewpoint-Catalog · viewpoints-and-perspectives.info/home/perspectives · arxiv.org/pdf/2006.04975 (4+1) · arc42.org/overview.

---

## B · SWEBOK v4 (18 KAs) + Essence/SEMAT (alphas + areas of concern)

**SWEBOK v4 (2024, 18 KAs):** Requirements · **Architecture**(nueva) · Design · Construction · Testing · **Operations**(nueva) · Maintenance · Config-Mgmt · Engineering-Mgmt · Process · Models&Methods · Quality · **Security**(nueva) · Professional-Practice · Economics · Computing/Mathematical/Engineering-Foundations.

**Essence / OMG (SEMAT) — 7 ALPHAS × 3 AREAS OF CONCERN** (dimensiones UNIVERSALES de cualquier emprendimiento de software):
- **Customer:** Opportunity (el porqué/circunstancias) · Stakeholders (afectan/afectados).
- **Solution:** Requirements (qué debe hacer) · Software System (el producto).
- **Endeavour:** Work (la actividad) · Team (el grupo + capacidades) · Way-of-Working (prácticas+herramientas).
- **Alpha** = cosa esencial que se *rastrea*, con **estados de salud**. **Area of concern** = agrupación de alphas.

**Faltan en el borrador (era ~100% "Solución"):** **Team** · **Way-of-Working** (Proceso lo mezcla) · **Opportunity/Stakeholders** (Customer, ausentes) · **Requirements** (Dominio roza pero ≠) · **Economics/valor** · **Maintenance**.
**Recomendación:** usar término propio para el eje, pero **agrupar con las 3 areas-of-concern de Essence (Cliente·Solución·Emprendimiento)** y tratar cada eje como "alpha" con estados de salud (= tu vacío/parcial/lleno). Nombres SWEBOK para ejes de Solución.
Fuentes: en.wikipedia.org/wiki/Software_Engineering_Body_of_Knowledge · sebokwiki.org · queue.acm.org/detail.cfm?id=2389616 · semat.org/web/book/part-i/chapter-6.

---

## C · TOGAF + Zachman + ArchiMate (agrupador rubro-agnóstico)

**TOGAF — BDAT:** Business · Data · Application · Technology. `Business` universal, pero `Data/App/Tech` sesgados a TI → **generaliza mal fuera de software** (en contabilidad el vendible no es una app). BDAT sirve solo como refinamiento dentro de "How/What".

**Zachman — 6 INTERROGATIVAS (rubro-agnósticas por construcción, lenguaje natural):**
| Interrogativa | Zachman | Mapea a |
|---|---|---|
| **What** | datos/inventario | dominio, entidades, datos, artefactos |
| **How** | función/proceso | proceso, receta, transición de caja |
| **Where** | red/geografía | deploy, entorno, geo, canal |
| **Who** | personas/roles | roles, RACI, agentes, permisos |
| **When** | tiempo | fase, release, cadencia, disparadores |
| **Why** | motivación | visión, valor, principios, mejora |
+ **6 perspectivas** (Ejecutivo→Negocio→Arquitecto→Ingeniero→Técnico→Operación) = eje ortogonal de **zoom/abstracción**.
**ArchiMate:** 3 capas (Business/App/Tech) × 3 aspectos (Active=quién / Behavior=qué hace / Passive=sobre qué) ≈ subconjunto Who/How/What.

**Recomendación:** raíz = **6 interrogativas de Zachman** (agnóstico, citable, cubre Who=roles·How=proceso·What=dominio/datos·When=fase·Where=deploy·Why=valor); zoom = perspectivas Zachman; refinamiento TI = BDAT. Essence areas-of-concern = buen complemento de salud/estado.
Fuentes: zachman.com/about-the-zachman-framework · en.wikipedia.org/wiki/Zachman_Framework · topictrick.com TOGAF-domains · en.wikipedia.org/wiki/ArchiMate.

---

## D · ISO 25010:2023 + FURPS+ + AOP (calidad = overlay)

**ISO/IEC 25010:2023 — 9 características de calidad de producto:** Functional-Suitability · Performance-Efficiency · **Compatibility**(co-existence/interoperability) · **Interaction-Capability**(ex-Usability; +inclusivity, +self-descriptiveness) · Reliability(maturity/availability/fault-tolerance/recoverability) · Security(+resistance) · **Maintainability**(modularity/reusability/analyzability/modifiability/**testability**) · **Flexibility**(ex-Portability; adaptability/**scalability**/installability/replaceability) · **Safety**(NUEVA).
Complementos SQuaRE: ISO 25019 (quality-in-use) · ISO 25012 (data quality, 15 características).

**3 niveles distintos:**
- **Concern estructural:** descompone *qué es* por localización (módulos, capas, datos, UX). Se modulariza → tiene "su carpeta".
- **Atributo de calidad (-ility):** propiedad *emergente* que se predica de TODAS las partes (seguridad, performance). NO tiene carpeta: *califica* al resto.
- **Cross-cutting concern (AOP):** cuando ese atributo, al implementarse, se dispersa/enreda (scattering+tangling) por muchos módulos (logging, security, i18n, caching). AOP lo reencapsula en *aspects*.

**Faltan (transversales):** Compatibility · Maintainability (aquí viven Testing+Observabilidad) · Flexibility(scalability) · Safety · Usability-como-calidad.
**Recomendación:** modelar calidad como **OVERLAY transversal** (no dimensiones peer). Los ejes estructurales *contienen* código; los -ilities *evalúan* código across. Modelo: **ejes estructurales (N) × overlay de calidad ISO 25010 (9)** + cross-cutting técnicos (logging/i18n/caching) = manifestación-en-código del overlay. Precedente: AOP (no crear módulo "logging" peer; tejer el aspecto).
Fuentes: iso25000.com/index.php/en/iso-25000-standards/iso-25010 · iso.org/standard/78176.html · en.wikipedia.org/wiki/FURPS · foojay.io AOP.

---

## E · Naming (dimensión vs terreno vs landscape)

| Término | Fuente | Nombra | eje | mapa | inst. |
|---|---|---|---|---|---|
| **dimension** | OLAP/dimensional-modeling | eje independiente de categorización (jerarquía) | ✓✓ | ✗ | ✗ |
| concern | 42010 | interés de stakeholder | ✓ | ✗ | ✗ |
| viewpoint | 42010 | plantilla para construir una vista | ✓(tipo) | ✗ | ✗ |
| view | 42010 | viewpoint aplicado a un sistema dado | ✗ | ~ | ✓✓ |
| perspective | R&W | cualidad transversal a las vistas | ✓(transv) | ✗ | ✗ |
| subdomain | DDD (problem space) | área a construir; core/supporting/generic | ✓✓ | ✗ | ✗ |
| alpha | Essence | cosa esencial que se rastrea, con **estados** | ✗ | ✗ | ✓(estado) |
| area-of-concern | Essence | agrupación de alphas | ✓(grupo) | ✗ | ✗ |
| **landscape** | Wardley/TOGAF-Architecture-Landscape | **el mapa/terreno COMPLETO** que la org ocupa | ✗ | ✓✓ | ~ |
| terrain/terreno | estrategia militar (metáfora) | **sin concepto formal** en SW | ✗ | metáfora | ✗ |

**Resolución:** "dimensión" y "terreno" NO son sinónimos.
- **"dimensión"** = respaldo fuerte (OLAP) = el EJE. Palabra correcta para (a).
- **"terreno"** = metáfora sin respaldo; su traducción formal es **"landscape"** = el **mapa COMPLETO** (b), no un eje. El borrador usó terreno para el eje → mismatch.
- **Recomendación:** (a) eje = **dimensión** · (b) mapa completo = **terreno/landscape** (aquí SÍ cabe "terreno", para el TODO) · (c) instancia por-proyecto = **view** (42010). Alt: (a) = **subdominio** DDD (core/supporting/generic).
Fuentes: iso-architecture.org/42010/cm · olap.com/learn-bi-olap/olap-bi-definitions/dimension · medium.com/nick-tune-tech-strategy-blog (DDD problem/solution space) · learnwardleymapping.com/landscape · guides.visual-paradigm.com/understanding-the-togaf-ea-landscape.
