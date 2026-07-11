# Research — Terrenos del arnés dev (síntesis de 5 investigaciones web · 2026-07-10)

> Insumo de co-diseño (realineación core ciclo-arnés). NO firmado. Base para la estructura de
> terrenos + organización en carpetas. Cada sección = un subagente de investigación; fuentes al pie.
>
> **Marco:** TERRENO = dimensión que un proyecto debe DEFINIR y CONSTRUIR como grafo navegable
> as-code (el arnés la PREGUNTA/OBLIGA). CONOCIMIENTO = rules + skills que IMPLEMENTAN las
> decisiones del terreno (nacen de ellas). Requisito: hojas escuetas, 3-4 saltos, punteros
> auto-derivados del código → mínimo token.

---

## 1 · Terreno ARQUITECTURA

Espina = **C4 model** (zoom Context→Container→Component→Code) mapeado a los ejes de **arc42** (12 secciones).

```
arquitectura/
├─ estilo/         monolito | modular | microservicios | distribuido | serverless | event-driven  (1 hoja + ADR)
├─ contexto/       C4 L1: sistema + actores + sistemas externos
├─ niveles/        (C4 L2 Container; un nodo por eje, hoja escueta + puntero a código)
│   ├─ frontend/ backend/ integracion/ datos/ despliegue-devops/ seguridad/ observabilidad/
├─ componentes/    C4 L3: bloques dentro de cada contenedor → 1 archivo/módulo
├─ decisiones/     ADRs (MADR/Nygard: contexto·decisión·consecuencias·estado)
├─ reglas/         fitness functions (boundaries permitidos/prohibidos)
└─ calidad/        atributos (perf/escala/resiliencia) + riesgos/deuda
```

**Auto-puntero / living-docs:** versiones←manifiestos · estructura←Structurizr DSL o AST · boundaries reales←dependency-cruiser/ArchUnit/go-arch-lint · contratos←OpenAPI/proto/GraphQL SDL · ADR-drift: cada ADR "aceptada" mapea a ≥1 fitness function, CI valida el diff (patrón idéntico a `estado.sh --check`).
**Rules nacen de:** estilo→rule anti-deriva · niveles/boundaries→fitness functions · decisiones→rule "toda ADR aceptada ⇒ fitness function" · datos→rule ownership (1 servicio=1 store). **Regla de Rosetta: definir un nodo del terreno GENERA la rule/skill que lo implementa.**
**Carpetas:** `docs/architecture/{workspace.dsl, INDEX.md, niveles/*.md, decisions/NNNN-*.md, fitness/*, .generated/}`.
Fuentes: c4model.com · docs.structurizr.com/dsl · arc42.org · github.com/sverweij/dependency-cruiser · platformtoolsmith.com/blog/operationalizing-adrs-fitness-functions.

---

## 2 · Terreno PROCESO & ROLES

Espina FIJA fases→cajas (cada caja = paso con 1 rol cubridor):
`F1 Idea/Discovery · F2 Definición · F3 Spec&Diseño · F4 Build · F5 Test/QA · F6 Review · F7 Release · F8 Operate · F9 Improve(loop→F1)`. Dos value-streams DORA: happy-path (feature) + recovery-path (incidente).

**Rol por caja = RACI:** cada caja nombra 1 **Accountable** (dueño, regla dura: uno solo) + 1..n Responsible. Roles-slot (arquetipos, no personas): PO/PM · BA · UX · Architect · Developer · QA/Test · Reviewer · DevOps/Release · SRE/Ops · Security.
**Partir el proceso:** value-stream con hand-offs explícitos (DoD de caja saliente = entrada de la siguiente) + WIP limits + pull. Se corta por tramos de fases → cada ejecutor recibe su sub-arnés (cajas+rol+contratos), no el arnés entero.
**FIJO (arnés):** secuencia fases→cajas, catálogo de roles-slot, "toda caja nombra su rol", gates/transiciones legales, contratos de hand-off. **PROYECTO:** quién llena cada rol, cajas opcionales, WIP caps, tooling, módulos, cómo se trocea.
**Rules nacen de:** cada gate/contrato/invariante de caja → rule verificable. **Skills:** el know-how que ese rol corre en esa caja (p.ej. `/dev` en Build, `/auditor` en Review).
**Carpetas:** `terreno/proceso-roles/{spine.yaml, roles/<slot>.yaml, raci.yaml, handoffs/<caja>.yaml, tramos/<slug>.yaml}` + `conocimiento/{rules,skills}`.
Marcos: Scrum · Kanban(WIP/pull) · XP · SAFe · DevOps/CALMS+SRE · Platform Eng · Team Topologies · VSM/DORA.
Fuentes: dora.dev/guides/value-stream-management · atlassian.com/agile/software-development/sdlc · atlassian.com/agile/scrum/roles · scalac.io RACI · wrike.com/kanban-guide/kanban-wip-limits.

---

## 3 · Terreno UX / FRONTEND

```
terreno/ux-frontend/
├─ arquitectura.yaml   capas FSD + regla-de-import (grafo de dependencias)
├─ sistema-visual/
│  ├─ tokens.yaml      índice DTCG (auto-derivado de tokens.json) → {token}: $value·$type
│  └─ storybook.yaml   SSoT UI; índice de stories (auto-derivado de index.json) → {componente}: storyId·importPath·framework·args
└─ escala-atomica.yaml atoms→molecules→organisms→templates→pages (clasifica cada story)
```

**Storybook = SSoT + puntero agnóstico:** el arnés OBLIGA — no hay componente en dev sin story primero (CDD bottom-up). Puntero = registro estable `{id, storyId, importPath, framework∈react/angular/vue, tokens[]}`; en diseño se referencia el `id` neutral, en dev se resuelve al `importPath` real.
**Auto-derivado:** Storybook `index.json` (stories+args) · `tokens.json` DTCG · Public-API FSD (`index.ts` por slice) · escala←tags de story.
**Métodos:** FSD (capas app>pages>widgets>features>entities>shared → slices por dominio → segments ui·api·model·lib·config; import solo hacia abajo, hermanas no) · Atomic Design (vocabulario de composición) · CDD (componente-aislado-primero) · Design Tokens DTCG (→ style-dictionary a CSS/Tailwind/Swift/Kotlin).
**Rules nacen de:** "diseñar solo sobre componentes existentes" (referencia debe resolver a storyId vivo; si no, crea la story antes de aprobar — falla honesto) · "sin literales visuales" (solo tokens) · "import legal FSD". **Skills:** forjar-componente · diseñar-sobre-storybook (lista existentes → superset, nunca reinventa).
**Carpetas:** `terreno/ux-frontend/*.yaml` + `conocimiento/ux-frontend/{rules,skills}` + `web/tokens/*.tokens.json` + `web/src/{shared,entities,features,widgets,pages,app}/` + `web/**/*.stories.*`.
Fuentes: feature-sliced.design · chromatic.com/blog/component-driven-development · storybook.js.org/docs/get-started/why-storybook · atomicdesign.bradfrost.com · designtokens.org/tr/drafts/format · styledictionary.com/info/dtcg.

---

## 4 · Terrenos METODOLOGÍA BASE + TECNOLOGÍAS

**Metodología** = un terreno, dos **carriles pares** (backend/frontend) con el MISMO invariante: capas con dependencia unidireccional ("deps apuntan al centro").

```
terreno/metodologia-base/
├─ INDEX.yaml    carriles + regla-madre
├─ backend/  estilo(DDD táctico+Hexagonal) · capas(domain→application→infrastructure) ·
│            building-blocks(entity·VO·aggregate·repo·domain-event·domain-service) ·
│            contextos(bounded-contexts+context-map) · scaffold/*.hbs(plop/hygen/cookiecutter)
└─ frontend/ estilo(FSD) · capas(app→…→shared) · slice-segment · scaffold/*.hbs
```
Estilo AGNÓSTICO: `estilo.yaml` nombra el *patrón* (DDD/Hexagonal/FSD), no el framework.

**Tecnologías** = cero mantenimiento manual, hojas GENERADAS:
```
terreno/tecnologias/
├─ INDEX.yaml       GENERADO: qué-se-usa-en-qué (módulo × dep)
├─ sbom.cdx.json    GENERADO por CI (CycloneDX): SSoT de dependencias
├─ por-modulo/*.yaml GENERADO: dep → módulo/capa que la importa
└─ reglas-uso.yaml  a mano (poco): "solo lib X para HTTP"; lo demás deriva
```
**Puntero vivo:** manifiestos (`package.json`/`go.mod`/`Cargo.toml`/`pyproject.toml`) → extractor `cdxgen`/`cyclonedx-gomod` → **SBOM CycloneDX** (nombre·versión·licencia·purl·hash). Grafo módulo→tech: dependency-cruiser / `go mod graph`. Navegación `INDEX→por-modulo/<x>.yaml→purl→sbom` (3 saltos). Regenerado por commit; drift-gate rompe merge si stale.
**Rules nacen de:** `backend/capas`→"domain no importa infrastructure" (ArchUnit/dep-cruiser) · `frontend/capas`→FSD import-legal · `scaffold/*`→skill "forjar aggregate/feature" (capas correctas por construcción) · `reglas-uso`→verificada vs SBOM. Cada rule = fitness function ejecutable = capability (SSoT funcional).
Métodos: DDD(strategic+tactical) · Hexagonal(ports&adapters) · Clean/Onion(DIP) · FSD · Plop/Hygen/Yeoman/cookiecutter/Nx-generators · CycloneDX/SPDX · ArchUnit/import-linter/Tach.
Fuentes: martinfowler.com/bliki/BoundedContext.html · learn.microsoft.com tactical-DDD · feature-sliced.design/docs/reference/layers · cyclonedx.org/tool-center · github.com/CycloneDX/cdxgen · infoq.com/articles/fitness-functions-architecture.

---

## 5 · SET COMPLETO de terrenos + grafo navegable + carpetas (integrador)

**[F]** = fijo-del-arnés (forma estable, el kit lo impone/pregunta en todo rubro) · **[P]** = contenido del proyecto. Casi todos son **[F]-forma / [P]-contenido**.

| Terreno | 1 línea | F/P |
|---|---|---|
| Dominio/negocio | lenguaje ubicuo, bounded contexts, glosario, reglas (arc42 §12) | [P] contenido |
| Arquitectura | building blocks, boundaries, vistas C4, ADRs (arc42 §4-9) | [F] método / [P] decisiones |
| Tecnologías/stack | lenguajes, frameworks, toolchain verbatim de build-files | [P] auto-derivado |
| Proceso/metodología | value-stream, WIP, gates, disciplina de paquete | [F] |
| UX/diseño | flujos, inventario, tokens, SSoT Storybook | [F] disciplina / [P] superficie |
| Testing/calidad | ISO 25010 functional-suitability, niveles, coverage, fitness | [F] modelo / [P] umbrales |
| Datos/persistencia | modelo, esquema, migraciones, retención, ownership | [P] |
| Seguridad | ISO 25010 confidentiality/integrity, authz, secrets, threat-model | [F] checklist / [P] postura |
| Observabilidad/SRE | logs/metrics/traces, SLO/SLI, alerting, runbooks | [F] señales / [P] SLOs |
| CI-CD/release | pipeline, gates, versionado, rollback, trunk-based | [F] |
| Performance | ISO 25010 time-behaviour, capacity, budgets, hot-paths | [F] atributo / [P] budgets |
| Compliance/legal | licencias, privacidad (GDPR), auth-terms, regulatorio | [P] |
| Documentación | Diátaxis (tutorial/how-to/reference/explanation), docs-as-code | [F] |
| Reliability/resiliencia | ISO 25010 availability, fault-tolerance, recoverability | [F] atributo / [P] targets |
| Operación/deploy | topología, entornos, IaC, provisioning (arc42 §7) | [P] |
| Compatibilidad/interop | ISO 25010 interoperability, contratos, versionado API | [F] método / [P] contratos |

**Grafo navegable bajo-token:**
- **Hoja atómica** (Zettelkasten): frontmatter YAML mínimo (`id, terreno, estado, owner, punteros[]`) + cuerpo ≤~15 líneas. Un dato = una hoja. Punteros auto-derivados del código (`provenance:`), nunca tecleados.
- **Index-First / progressive disclosure:** cada carpeta tiene `INDEX.md` que lista hijos con una línea "cuándo es relevante". El agente carga solo el INDEX; baja a la hoja si matchea (−17% a −80% tokens). Doctrina de context-engineering / Agent Skills.
- **3-4 saltos:** `CLAUDE.md router → docs/<eje>/INDEX → <terreno>/INDEX → <slug>.yaml`.
- **Frameworks de encuadre:** Diátaxis (tipos de contenido, no mezclar) · arc42 (12 slots, vacíos permitidos) · C4 (zoom) · ISO 25010 (9 atributos de calidad transversales).

**Carpetas — terreno y conocimiento ENTRELAZADOS (co-locados por terreno):** recomendación fuerte: NO una carpeta global `conocimiento/` separada; el knowledge vive junto a la dimensión que aplica (co-locación reduce saltos y drift). Anclado en `project.config.yaml` (seam/DIP).
```
project.config.yaml        # SEAM: slots abstractos → valores del proyecto
CLAUDE.md                  # router puro
docs/
  terreno/                 # dimensiones as-code (una carpeta por terreno)
    INDEX.md               # nivel-1: lista terrenos + "cuándo"
    dominio/  arquitectura/(model/ boundaries/ contracts/ fitness/ adr/)  testing/  datos/
    seguridad/(threat-model/ knowledge/reglas.yaml)  observabilidad/(slo/ runbooks/)
    ci-cd/  performance/  compliance/  ux/  proceso/  ...
    <cada terreno>/knowledge/   # rules/skills que NACEN de las decisiones de ESE terreno
  capabilities/            # SSoT FUNCIONAL: qué SABE hacer (deriva del código)
```
**Terreno↔Conocimiento:** terreno = grafo declarativo (qué dimensiones + estado). Conocimiento = ejecutable que lo aplica. De una decisión (`seguridad/authz.yaml`) NACE una rule (`seguridad/knowledge/authz.rule.yaml`) y/o skill. Puntero **bidireccional** (hoja→su knowledge; rule/skill→su terreno-origen; trazabilidad = `codigo-traza-a-capability`). Rules/skills OPERATIVAS (harness/gates) siguen en `.claude/` (kit); las DE CONOCIMIENTO DEL PROYECTO se co-locan bajo su terreno → un salto del dato a su enforcement.
Fuentes: iso.org/standard/78176 (ISO 25010:2023) · quality.arc42.org/standards/iso-25010 · arc42.org · infoq.com/articles/C4-architecture-model · diataxis.fr · writethedocs.org/guide/docs-as-code · sierra.ai/blog/context-engineering-the-key-to-great-agents · nx.dev folder-structure · zettelkasten.de/atomicity.
