# ArnesIA — la fábrica de arneses (crear · mapear · observar · mejorar)

> **ESTE ARCHIVO ES UN ROUTER.** No guarda historia, backlog ni cifras — solo apunta. ¿Qué
> querés? → abrí la hoja. **Sesión nueva:** leé `ESTADO.md` (fase + paquete activo) + `BACKLOG.md`
> (lo abierto) y seguí como si fuera la misma conversación. Reorg de 4 ejes firmada 2026-07-09
> (ledger HS-18); estructura de docs = espejo del sistema `memory/` (índice + hojas atómicas + TTL).

**Qué es (1 línea):** fábrica de los arneses que alpacapurpura crea y vende por **rol ×
proceso**; el vendible es el ARNÉS con mejora continua, ArnesIA = medio de producción. Solo
arneses propios. Constitución de 11 principios + visión → [`VISION.md`](./VISION.md) (v3 FIRMADA).

## Necesito X → leo Y

| Necesito… | Hoja |
|---|---|
| **Qué SABE HACER el sistema hoy (validado, SSoT funcional)** | [`CAPABILITIES.md`](./CAPABILITIES.md) |
| **En qué fase estamos · cifras vivas · paquete activo** | [`ESTADO.md`](./ESTADO.md) |
| **Qué está abierto / pendiente / por hacer** | [`BACKLOG.md`](./BACKLOG.md) |
| **Qué se decidió y cuándo** (historia) | [`LEDGER.md`](./LEDGER.md) índice → `ledger/HS-NN.md` |
| **El norte / visión / 11 principios** | [`VISION.md`](./VISION.md) |
| **Reglas de negocio · metodología · contrato de caja · disciplina §10** | [`METODOLOGIA.md`](./METODOLOGIA.md) |
| **Stack / tecnologías** | [`STACK.md`](./STACK.md) → [`arch/INDEX.md`](./arch/INDEX.md) |
| **Arquitectura as-code · boundaries · fitness** | [`arch/INDEX.md`](./arch/INDEX.md) |
| **Estándar as-code por elemento** (skill/hook/rule/…) | [`knowledge/INDEX.md`](./knowledge/INDEX.md) |
| **UX firmada · inventario · mockups** | [`UX.md`](./UX.md) · `mockups/` |
| **Trabajar una feature** (mockup→spec→PARIDAD) | `historias/<pkg>/INDEX.md` |

## Doctrina de desarrollo (obligatoria)

- **Disciplina de paquete de trabajo** (toda feature nueva; canónico = METODOLOGIA §10): cada
  feature vive en `historias/AAAA-MM-DD-<slug>/`; mockup→decisiones→spec→implementar→PARIDAD con
  firmas 🧑‍⚖️ entre etapas. **Toda decisión conversada se escribe en `decisiones.md` del paquete
  EN EL MISMO TURNO.** Sesión nueva arranca leyendo el `INDEX.md`/«Retomar aquí» del paquete activo
  (puntero en `ESTADO.md`). Cada iteración firmada se commitea a main.
- **CAPABILITIES = SSoT funcional (docs-as-code):** **ningún cambio de código existe sin construir
  o modificar un capability y actualizar `CAPABILITIES.md`.** Doctrina + enforcement (R1 integridad ·
  R2 cobertura · R3 gate-commit · R4 estado-generado; rollout warn→block) =
  [`arch/boundaries/codigo-traza-a-capability.md`](./arch/boundaries/codigo-traza-a-capability.md).
- **Honestidad del repo:** nada se da por «listo» sin verificar en vivo; gaps quedan VISIBLES
  (`deferred`/`sin-check`/`no-reconocido`), jamás pass fabricado. Cifras se GENERAN, no se teclean.

## Reglas duras

- **Git:** trunk-based — `main` única, commit/push directo, tags semver en releases.
- **Arnés de construcción:** kit dev = plugin `harness@prenter-marketplace` canal ESTABLE; mejoras
  se upstreamean al kit (backflow), jamás fork silencioso.
- **Estándar propio:** solo arneses propios — nosotros seteamos el estándar (VISION §Constitución).
