# ArnesIA — la fábrica de arneses (crear · mapear · observar · mejorar)

> **ESTE ARCHIVO ES UN ROUTER.** No guarda historia, backlog ni cifras — solo apunta. ¿Qué
> querés? → abrí la hoja. **Sesión nueva:** corré **`/pm`** (front-door: dónde estamos + qué
> sigue) o leé `docs/product/checkpoint.md` (presente) + `docs/product/BACKLOG.md` (abierto).
> Homologación al método del plugin `harness@prenter-marketplace` firmada 2026-07-09 (paquete
> `docs/product/stories/2026-07-09-homologacion-metodologia/`): toda doc no-skill/no-CLAUDE vive en
> [`docs/`](./docs/README.md) (`product/` · `architecture/` · `process/`); el seam es
> [`project.config.yaml`](./project.config.yaml). Migración en curso (ver INDEX del paquete).

**Qué es (1 línea):** fábrica de los arneses que alpacapurpura crea y vende por **rol ×
proceso**; el vendible es el ARNÉS con mejora continua, ArnesIA = medio de producción. Solo
arneses propios. Constitución de 11 principios + visión → [`docs/product/vision.md`](docs/product/vision.md) (v3 FIRMADA).

## Necesito X → leo Y

| Necesito… | Hoja |
|---|---|
| **Orientarme en el proceso (dónde estamos · qué sigue)** | skill **`/pm`** (front-door) |
| **Qué SABE HACER el sistema hoy (validado, SSoT funcional)** | [`docs/product/capabilities/`](./docs/product/capabilities/INDEX.md) (82 hojas YAML por-cap + BDD) |
| **En qué fase estamos · cifras vivas · paquete activo** | [`docs/product/checkpoint.md`](./docs/product/checkpoint.md) |
| **Qué está abierto / pendiente / por hacer** | [`docs/product/BACKLOG.md`](./docs/product/BACKLOG.md) |
| **Qué se decidió y cuándo** (historia) | [`docs/product/LEDGER.md`](./docs/product/LEDGER.md) índice → `docs/product/ledger/HS-NN.md` |
| **El norte / visión / 11 principios** | [`docs/product/vision.md`](docs/product/vision.md) |
| **Reglas de negocio · metodología · contrato de caja · disciplina §10** | [`docs/process/metodologia.md`](docs/process/metodologia.md) |
| **Stack / tecnologías** | [`docs/architecture/stack.md`](./docs/architecture/stack.md) → [`docs/architecture/INDEX.md`](./docs/architecture/INDEX.md) |
| **Arquitectura as-code · boundaries · fitness** | [`docs/architecture/INDEX.md`](./docs/architecture/INDEX.md) |
| **Estándar as-code por elemento** (skill/hook/rule/…) | [`docs/architecture/knowledge/INDEX.md`](./docs/architecture/knowledge/INDEX.md) |
| **UX firmada · inventario · mockups** | [`docs/product/ux.md`](docs/product/ux.md) · `mockups/` |
| **Diseñar / mockupear UI (LEER antes de forkear)** | [`mockups/INDEX.md`](./mockups/INDEX.md) — línea base vigente + disciplina superset (SSoT UI = Storybook) |
| **Trabajar una feature** (mockup→spec→PARIDAD) | `docs/product/stories/<pkg>/INDEX.md` |

## Doctrina de desarrollo (obligatoria)

- **Disciplina de paquete de trabajo** (toda feature nueva; canónico = METODOLOGIA §10): cada
  feature vive en `docs/product/stories/AAAA-MM-DD-<slug>/`; mockup→decisiones→spec→implementar→PARIDAD con
  firmas 🧑‍⚖️ entre etapas. **Toda etapa de mockup/UI arranca leyendo [`mockups/INDEX.md`](./mockups/INDEX.md)**
  (línea base = superficie vigente; propuesta = superset, jamás reinventar lo firmado; SSoT del UI = Storybook). **Toda decisión conversada se escribe en `decisiones.md` del paquete
  EN EL MISMO TURNO.** Sesión nueva arranca leyendo el `INDEX.md`/«Retomar aquí» del paquete activo
  (puntero en `docs/product/checkpoint.md`). Cada iteración firmada se commitea a main.
- **CAPABILITIES = SSoT funcional (docs-as-code):** **ningún cambio de código existe sin construir
  o modificar un capability** (`docs/product/capabilities/{module}/{slug}.yaml`). Doctrina + enforcement (R1 integridad ·
  R2 cobertura · R3 gate-commit · R4 estado-generado; rollout warn→block) =
  [`docs/architecture/boundaries/codigo-traza-a-capability.md`](./docs/architecture/boundaries/codigo-traza-a-capability.md).
- **Honestidad del repo:** nada se da por «listo» sin verificar en vivo; gaps quedan VISIBLES
  (`deferred`/`sin-check`/`no-reconocido`), jamás pass fabricado. Cifras se GENERAN, no se teclean.

## Reglas duras

- **Git:** trunk-based — `main` única, commit/push directo, tags semver en releases.
- **Arnés de construcción:** kit dev = plugin `harness@prenter-marketplace` canal ESTABLE; mejoras
  se upstreamean al kit (backflow), jamás fork silencioso.
- **Estándar propio:** solo arneses propios — nosotros seteamos el estándar (VISION §Constitución).
