# ArnesIA — la fábrica de arneses (crear · mapear · observar · mejorar)

> ESTE ARCHIVO ES UN ROUTER. No guarda historia, backlog ni cifras — solo apunta.
> ¿Qué querés? → abrí la hoja. Sesión nueva: leé `ESTADO.md` + `BACKLOG.md` y seguí.

**Qué es (1 línea):** fábrica de los arneses que alpacapurpura crea y vende por **rol ×
proceso**; el vendible es el ARNÉS con mejora continua, ArnesIA = medio de producción. Solo
arneses propios. Detalle y constitución de 11 principios → [`VISION.md`](./VISION.md).

## Necesito X → leo Y

| Necesito… | Hoja |
|---|---|
| **Qué sabe hacer el sistema HOY (validado)** | [`CAPABILITIES.md`](./CAPABILITIES.md) |
| **En qué fase estamos · cifras vivas · paquete activo** | [`ESTADO.md`](./ESTADO.md) |
| **Qué está abierto / pendiente / por hacer** | [`BACKLOG.md`](./BACKLOG.md) |
| **Qué se decidió y cuándo** (historia) | [`LEDGER.md`](./LEDGER.md) (índice) → `ledger/HS-NN.md` |
| **El norte / visión / principios** | [`VISION.md`](./VISION.md) |
| **Reglas de negocio · metodología · contrato de caja** | [`METODOLOGIA.md`](./METODOLOGIA.md) |
| **Estándar as-code por elemento** (skill/hook/rule/…) | [`knowledge/INDEX.md`](./knowledge/INDEX.md) |
| **Arquitectura/diseño técnico as-code · boundaries · fitness** | [`arch/INDEX.md`](./arch/INDEX.md) |
| **UX firmada · inventario · mockups** | [`UX.md`](./UX.md) · `mockups/` |
| **Trabajar una feature** (mockup→spec→PARIDAD) | `historias/<pkg>/INDEX.md` → `RETOMAR.md` |

## Reglas duras (siempre)

- **Disciplina de paquete de trabajo** (obligatoria, toda feature nueva): canónico en
  METODOLOGIA §10. mockup→decisiones→spec→implementar→PARIDAD, con firmas 🧑‍⚖️ entre etapas.
  Toda decisión conversada se escribe en `decisiones.md` del paquete EN EL MISMO TURNO.
  Sesión nueva arranca leyendo el `RETOMAR.md` del paquete activo (puntero en `ESTADO.md`).
- **Git:** trunk-based — `main` única, commit/push directo, tags semver en releases.
- **Arnés de construcción:** kit dev = plugin `harness@prenter-marketplace` canal ESTABLE;
  mejoras se upstreamean al kit (backflow), jamás fork silencioso.
- **Estándar propio:** solo arneses propios — nosotros seteamos el estándar (VISION §Constitución).

## Stack (referencia rápida — detalle en `arch/`)

Binario Go `arnesia` (serve/open/index/publish, :4200) · UI Vite+React SPA `go:embed` ·
Mapa = sustrato HTML+SVG (bandas/carriles) · SQLite puro-Go índice desechable, JSONL de
`~/.claude` = fuente de verdad · shell Tauri 2 desde v1 · conexión CC = subproceso-conductor
stream-json. Decisiones y su historia → `ledger/` (grep por tema).
