# 01-spec.md — Template (PO)

> Owner: `/po`. **Spec ejecutable Gherkin AI-resistant.** Esta es la fuente de verdad de QUÉ debe construirse.
> Los architect, dev y auditor consumen ESTE archivo + el story YAML.
> Los scenarios de aquí se reflejan 1:1 en el story YAML (`docs/product/stories/{m}/{story-id}.yaml`).

---
story_id: STORY_ID_KEBAB
type: ui-story | agentic-story | service-story | bugfix
module: MODULE_NAME
capability: CAPABILITY_ID
po_version: 1                                     # bump cuando cambies post-handoff
last_modified: 2026-05-04T14:30Z
ratified_by_chris: false                          # /po pide ratificación antes pasar a UX/architect
links:
  story_yaml: "../../../../../product/stories/{module}/{story-id}.yaml"
  story_md: "00-story.md"
---

## Resumen ejecutivo

[1 párrafo: qué se construye, para quién, outcome esperado.]

> **Convención de comentarios (doc vivo · /po-ux).** Este `01-spec.md` se edita en vivo en el cockpit durante
> el refinamiento. Chris deja sus notas como blockquote `> 🗨️ CHRIS: ...` (borrá esto / mejorá aquello / agregá
> esto) para distinguirlas del cuerpo; el refiner las reconcilia a un doc limpio y cada pedido va a
> `chris-input.md`. (REQ-TAKING-DETAIL §5 · marcador ratificado 2026-06-08.)

<!-- ═══ RONDA 1 · input-spec (intención) · ✍ FIRMA 1 — solo /po-ux UI · ver spec-mapa-funcional.md § Dos rondas ═══ -->

## § Dónde vive (RONDA 1 · cement 2026-06-03 · solo /po-ux)

> Zona/caja del sistema + shell + ruta donde el user aterriza. Derivada del árbol de `.claude/rules/paradigm-arquitectura.md`. (Para service/agentic stories esta sección es opcional.)

- **Zona/caja:** [Agentes {agente} | Plataforma {acceso/onboarding/configuración} | Infraestructura {...}] — derivada del `SYSTEM-MAP.yaml`
- **Shell:** [qué shell aplica · del `{sistema}/docs/architecture/SHELL-DESIGN-CONTRACT.md` · si no existe → generar con el design-system actual]
- **Ruta del user:** [`/[tenantId]/(shell-organism)/{agent}/{subtab}/...` donde el user aterriza]

> **Sin mockup en RONDA 1 (W0.5-bis · funcional-primero).** Primero se cementa lo funcional (mapa + pantallas/campos);
> el mockup creativo nace DESPUÉS de FIRMA 1, en `§ Mockup FINAL` (RONDA 2). No se dibuja la forma antes de cerrar el qué.

## § Mapa funcional (capa humana — ratifica Chris ANTES de UX/architect)

> **v5 cement 2026-05-31 (Opción A).** Esta sección es el **panorama en lenguaje humano**: lo que Chris
> lee para validar QUÉ se va a construir, sin tener que reconstruir el flujo desde el Gherkin.
> NO compite con el Gherkin — vive a otra altitud. El Gherkin de abajo es la *formalización* de esto;
> la `§ Matriz de cobertura` los liga (cada bifurcación/RN/AC → ≥1 SC). Se renderiza nativo en el cockpit.
> **Profundidad proporcional al tipo de story** (bugfix: happy path opcional, foco en repro+branch+RN).

### 1. Happy path (el camino dorado, narrado)

[Prosa numerada del flujo exitoso end-to-end. 3-8 pasos. Lenguaje humano, no Gherkin.]

1. El usuario [entra a / abre] ___ y ve ___.
2. [Acción] ___ → el sistema ___.
3. [Confirmación / efecto observable] ___.

### 2. Bifurcaciones (árbol de decisión — TODOS los branch points)

> Árbol, no lista plana. Cada nodo: **condición → resultado → [SC que lo cubre]**.
> Acá Chris valida COMPLETITUD: un branch sin SC = hueco visible.

```
Happy path
├─ Bif-1 · ¿[condición]?
│   ├─ sí → [resultado]                         → SC-1
│   └─ no → [resultado alterno]                  → SC-2
├─ Bif-2 · ¿[condición de borde/error]?         → [resultado]  → SC-3
└─ Bif-3 · ¿[condición adversarial/duplicado]?  → [resultado]  → SC-4
```

### 3. Reglas de negocio (RN — invariantes en lenguaje humano)

> Constraints transversales del dominio. Numeradas. Se reflejan en la cap (`capability.business_rules`).

- **RN-1** — [invariante en una frase. Ej: "No se puede desactivar un doctor con citas futuras confirmadas."]
- **RN-2** — [...]

### 4. Criterios de aceptación (AC — checklist "listo cuando…")

> Nivel feature-done (NO son los scenarios; son el checklist de cierre que Chris tilda).

- [ ] **AC-1** — [condición observable de que la feature está completa]
- [ ] **AC-2** — [...]

## § Pantallas (RONDA 1 · campos NUEVOS vs EXISTENTES — SIN mockup todavía)

> ★ W0.5-bis (REQ-TAKING-DETAIL §4/§10). Enumera las vistas + sus campos diciendo cuáles son **NUEVOS** y
> cuáles ya **EXISTEN** (en la entidad/cap actual) — en viñetas humanas, NO Gherkin, ANTES de cualquier mockup.
> (Para shell stories el equivalente es Atomic Design + Reuse Map; ver el override de marca.)

| Vista / pantalla | Campo | Nuevo o existente | Origen (entidad/cap) | Validación |
|---|---|---|---|---|
| [pantalla] | [campo] | nuevo \| existente | [entidad/cap de donde sale] | [regla] |

## § Dudas abiertas (RONDA 1 · Chris las cierra ANTES del mockup)

- [ ] [duda funcional 1 — bloquea FIRMA 1 si no se resuelve · detalle en § Open questions]

<!-- ═══ RONDA 2 · spec ejecutable · GENERADA al firmar · ✍ FIRMA 2 → refining→refined (mockup FINAL + Gherkin + matriz + business-rules + design-spec) ═══ -->

## § Mockup FINAL (RONDA 2 · compone del design-system-canon · ANTES del Gherkin)

> ★ W0.5-bis (REQ-TAKING-DETAIL §4 · spec-mapa-funcional.md § Dos rondas · `design-system-canon.md` binding HARD).
> Recién acá — con lo funcional ya cerrado (FIRMA 1) — nace el mockup creativo: **shell completo + la hoja
> correspondiente + TODOS los campos conversados + TODOS los átomos**, **compuesto partiendo de Storybook**
> (`{design_system_ref.package}` = SSoT visual · `design-system-canon.md §5` · átomos + layout-primitives + archetypes REALES,
> tokens de la fuente única — spacing/radius/tipografía/color **NUNCA arbitrary**). **Partí de las stories de Storybook
> (`build-storybook` / `:{design_system_ref.storybook_port}` / `{design_system_ref.showcase_route}`) para reusar lo que ya existe ANTES de crear**;
> un átomo/primitiva faltante se **PROPONE + PROMUEVE a `{design_system_ref.package}` + su story** (vía `/pm-{platform}`, queda reusable).
> NO se maqueta a mano con `<div>` + clases sueltas, NO se copia CSS compartido a mano, NO se reinventa una primitiva
> existente. El refiner es creativo y **puede MEJORAR lo escrito** (actualiza el § Mapa funcional si la forma cambió
> algo). Lo que se ve en Storybook = lo que se programa. Iterá hasta que Chris lo apruebe. **✍ FIRMA 2 = la firma FINAL única**
> (`mockup_final_signed: true`) → dispara la GENERACIÓN de la RONDA 2.

- **Mockup:** [compuesto de Storybook — shell (stories `Shell/*`) + hoja + átomos reales nombrados de `{design_system_ref.package}` (revisá Storybook `build-storybook` / `:{design_system_ref.storybook_port}` o `{design_system_ref.showcase_route}`); link opcional al render]
- **Estados:** default / hover / loading / empty / error / success
- **Microcopy:** [Spanish neutro · ver `.claude/rules/spanish-text.md`]

## Acceptance Criteria (Gherkin AI-resistant)

> **★ GENERADO en FIRMA 2 (W0.5-bis · funcional-primero).** El Gherkin de abajo + la `§ Matriz de cobertura`
> + las business-rules + el design-spec se **GENERAN** desde el `§ Mapa funcional` + el `§ Mockup FINAL` al
> firmar FIRMA 2 — **NO se redactan a mano durante el refinamiento** (REQ-TAKING-DETAIL §4/§10).
> **v4.1 cement 2026-05-19:** Mínimo 4 scenarios base + sub-categorías mandatory aplicables.
> Cada scenario es testeable + tiene grader explícito + `playwright_required` flag.
> /po-ux REFUSE ratificar refined si falta cobertura de sub-categorías aplicables.
> **v5 cement 2026-05-31:** cada scenario lleva `Covers:` con los IDs del `§ Mapa funcional`
> que formaliza (`Bif-N` / `RN-N` / `AC-N`). La `§ Matriz de cobertura` valida que no quede
> ningún branch/RN huérfano ni ningún SC sin hogar en el mapa.

### Scenario 1 — `happy-path` (`type: happy`)

**Covers:** [Bif-1, AC-1]                          # IDs del § Mapa funcional que este SC formaliza

**Given:**
- [precondición concreta y verificable]

**When:**
- [acción exacta del actor]

**Then:**
- [efecto 1 medible]
- [efecto 2 medible]
- [efecto 3 medible]

**playwright_required:** true | false
**Graders:**
- [Tipo grader] — [target/path]

---

### Scenario 2 — `[id-negative]` (`type: negative`)

**Given:** ...
**When:** [input/estado inválido]
**Then:**
- [error visible/respuesta clara]
- [estado NO se modifica]
- [audit log entry si aplica]

**playwright_required:** true | false
**Graders:** ...

---

### Scenario 3 — `[id-edge]` (`type: edge`)

**Given:** [estado de borde: concurrencia, límite, race condition]
**When:** ...
**Then:** ...

**playwright_required:** true | false
**Graders:** ...

---

### Scenario 4 — `[id-adversarial]` (`type: adversarial`)

> AI-resistant: usuario hostil, tenant cross-leak, prompt injection, datos sensibles.

**Given:** ...
**When:** [acción adversarial]
**Then:**
- [security/safety check explícito]
- [no leak]
- [audit/alerting]

**playwright_required:** true | false
**Graders:** ...

---

## ★ Sub-categorías mandatory v4.1 (cement 2026-05-19)

> Cada sub-categoría aplicable a la story DEBE tener ≥1 scenario adicional o `not_applicable_reason` ratificado por Chris.
> /po-ux gate refuse refined sin cobertura.

### Scenario 5 — `race-condition` (`type: edge`, sub: race_condition)

> Aplica cuando: story incluye create/update con unique constraint (slug, key único).

**Given:** [2 actores intentan crear/modificar mismo recurso simultáneamente]
**When:** [requests A + B llegan en window <100ms]
**Then:**
- [solo uno gana — el otro recibe 409 Conflict o 422 con mensaje claro]
- [DB consistency: 1 row con el slug/key]
- [no estado intermedio leaked]

**playwright_required:** true (testear via Promise.all 2 requests)
**Graders:**
- { type: e2e, path: "{sistema}/frontend/e2e/regression/{story-id}/{m}-edge.spec.ts", function: "test_concurrent_create" }
- { type: state_check, target: db, query: "SELECT count(*) FROM {table} WHERE slug='X'", expect: 1 }

`not_applicable_reason: <razón si NO aplica>`

---

### Scenario 6 — `concurrent-users` (`type: edge`, sub: concurrent_users)

> Aplica cuando: list/detail filterable consumido por 2+ tenants/users mismo momento.

**Given:** [tenant A y tenant B logged simultáneamente]
**When:** [tenant A lista resources + tenant B lista resources]
**Then:**
- [cada uno ve SOLO sus resources (tenant isolation)]
- [no cross-leak en queries]
- [performance: p95 < N ms ambos]

**playwright_required:** true (2 contextos e2e paralelos)
**Graders:** ...

`not_applicable_reason: <razón si NO aplica>`

---

### Scenario 7 — `network-failure` (`type: edge`, sub: network_failure)

> Aplica cuando: surface FE hace fetch API.

**Given:** [usuario en pantalla X]
**When:** [API request falla con 500 / 503 / timeout / connectivity drop]
**Then:**
- [error UI visible con mensaje claro Spanish neutro ("No pudimos cargar...")]
- [retry button presente]
- [no white screen, no infinite loading]
- [data en memoria NO se pierde (form drafts)]

**playwright_required:** true (mock `page.route` con 500)
**Graders:**
- { type: e2e, function: "test_network_failure_retry" }
- { type: visual_state, screen: "error", element: "[role=alert]", expect: "visible" }

`not_applicable_reason: <razón si NO aplica>`

---

### Scenario 8 — `empty-state` (`type: edge`, sub: empty_state)

> Aplica cuando: list / dashboard / search.

**Given:** [tenant nuevo sin data, o filtro retorna 0 items]
**When:** [usuario carga pantalla]
**Then:**
- [empty state illustration + heading + CTA (no white screen)]
- [microcopy Spanish neutro ("Aún no tienes..." / "No encontramos resultados")]
- [CTA dispara create flow o clear filters]

**playwright_required:** true
**Graders:** ...

`not_applicable_reason: <razón si NO aplica>`

---

### Scenario 9 — `large-dataset` (`type: edge`, sub: large_dataset)

> Aplica cuando: list con pagination.

**Given:** [tenant con ≥1000 items en {table}]
**When:** [usuario navega list]
**Then:**
- [pagination renderiza correctamente (no carga 1000 en DOM)]
- [scroll smooth, p95 render < 200ms per page]
- [filtros funcionan vs 1000 items]
- [no memory leak después N páginas]

**playwright_required:** true (seed DB con 1000 + navigate)
**Graders:** ...

`not_applicable_reason: <razón si NO aplica>`

---

### Scenario 10 — `accessibility` (`type: edge`, sub: accessibility)

> Aplica cuando: TODO surface FE user-facing.

**Given:** [pantalla cualquier estado]
**When:** [scan a11y automatizado + keyboard nav + screen reader]
**Then:**
- [0 violaciones critical/serious WCAG AA]
- [Tab order lógico]
- [ARIA labels en inputs/buttons sin texto visible]
- [Contrast ratio ≥ 4.5:1 (text), ≥ 3:1 (UI)]
- [Focus visible en TODOS interactivos]

**playwright_required:** true (scan a11y automatizado del runner e2e)
**Graders:**
- { type: axe, ruleset: "wcag2aa", paths: ["all-screens"] }

`not_applicable_reason: <razón si NO aplica — service-only stories>`

---

### Scenario 11 — `i18n` (`type: edge`, sub: i18n)

> Aplica cuando: copy user-facing o currency display.

**Given:** [3 tenants distintos locale (AR/MX/CL/PE/CO)]
**When:** [render screens con currency, dates, microcopy]
**Then:**
- [currency tenant_locale respetada (no hardcoded 'USD')]
- [dates formato es-LATAM (DD/MM/YYYY)]
- [copy Spanish neutro (no voseo regional excepto sales_agent voice tenant)]
- [tildes, ñ, ¿ ¡ renderizan correcto]

**playwright_required:** true (3 fixtures tenants distintos)
**Graders:** ...

`not_applicable_reason: <razón si NO aplica>`

---

## § Matriz de cobertura (el puente humano ↔ verificación)

> **v5 cement 2026-05-31 (Opción A).** Cierra el loop: cada **bifurcación** y cada **regla de negocio**
> del `§ Mapa funcional` mapea a ≥1 scenario Gherkin, y cada scenario tiene una **verificación REAL**
> (acción real ejercida + efecto observado — NUNCA "GET 200"). Es la mitad delantera del
> `gherkin-matrix.md` que el `/auditor` completa en Phase D.
>
> **Gate /po-ux:** REFUSE cerrar `refined` si hay un `Bif-N` o `RN-N` sin SC (hueco), o un SC sin
> ítem del mapa (scope creep). Ver `.claude/rules/test-design-doctrine.md` § Verificación REAL.
>
> **★ LEDGER DE COBERTURA VIVO (proceso v5 §5.2).** La columna `estado` hace esta matriz VIVA:
> po-ux la **siembra** en `refined` (todo `⬜ pendiente`); **dev-team la mantiene viva** durante
> developing (`✅ construido` con su test/ruta); en **G** lo no-construido se decide ítem-por-ítem
> (`⏳ ahora` corto+necesario · `→ historia {id}` spawn); **R** la congela; el **auditor** la verifica
> en Phase D. La columna `→ historia` ES la lista visible de lo NO construido (mata el gap invisible).
> **PISO HARD:** si `cap_change_type: new`, los ítems del **happy path** DEBEN estar `✅` antes de `done`
> (no se difiere el core). Estados: `✅ construido` · `⏳ ahora` · `→ historia {id}` · `⬜ pendiente`.

| Ítem (Mapa funcional) | Tipo | estado | Cubierto por | Verificación REAL (acción + efecto, no HTTP 200) |
|---|---|---|---|---|
| Bif-1 · [condición] | branch | ⬜ pendiente | SC-1 | [POST/PATCH real → efecto en DB/UI + log] |
| Bif-3 · [duplicado/adversarial] | branch | ⬜ pendiente | SC-4 | [acción hostil → 409/403 + row intacta] |
| RN-1 · [invariante] | rule | ⬜ pendiente | SC-3, SC-3b | [write que viola la regla → 422 + estado sin cambio] |
| AC-1 · [criterio cierre] | accept | ⬜ pendiente | SC-1, SC-8 | [flujo real + empty state] |

**Huecos detectados:** [ninguno | lista de Bif/RN sin SC — bloquea refined]
**SC huérfanos (sin ítem del mapa):** [ninguno | lista — revisar scope creep]
**Diferido (ledger · lo NO construido, visible):** [ninguno | `Bif-N → historia {sistema}-{slug}` con razón]

## Non-functional requirements

| Categoría | Requisito | Verificador |
|---|---|---|
| Latencia | p95 < N ms | métrica + load test |
| Cost | <= $X/session (agentic) | copilot_llm_call |
| Mobile | viewport >= 375px (ui) | e2e viewport resize |
| Accesibilidad | WCAG AA (ui) | scan a11y automatizado |
| i18n | Spanish neutro (no voseo, salvo sales_agent voz tenant) | Lint regex |
| PII | Response no expone PII sin mask | validación del response model |
| Tenant isolation | Tenant cross → 403/404 | adversarial scenario |

## Constraints técnicos heredados

- [De `.claude/rules/*` que aplican: backend-ddd, tenant-isolation, etc.]
- [Canonical docs relevantes: el framework/lib canónico relevante del stack — WebFetch the canonical docs URL (canonical-docs skill si está instalado)]

## Cross-module impact

- **Lee de:** [módulos cuyas tablas/eventos consume]
- **Es leído por:** [módulos que dependen]
- **Eventos emitidos:** [event_name v1]
- **Eventos consumidos:** [event_name v1]

## Open questions (para resolver con Chris ANTES de UX/architect)

- [ ] [Pregunta 1]
- [ ] [Pregunta 2]

## Próximo paso

- Si `type=ui-story` → ya producido por `/po-ux` (01-spec unificado); `/architect` consume este doc directo
- Si `type=agentic-story` → `/ux-agentico` lee `01-spec.md` → produce `02-design-agentic.md`
- Si `type=service-story` → skip UX → `/architect` directo

## Changelog

- v1 2026-05-04 — /po draft inicial
- v2 2026-05-04 — Chris ratificó scenarios 2 y 3, ajusté wording scenario 4
