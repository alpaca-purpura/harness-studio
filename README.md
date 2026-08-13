# ArnesIA

**La fábrica de arneses** del talento asistido por IA — **crear · mapear · observar · mejorar**.
Arneses por **rol × proceso** de compañía, con mejora continua; la observación es el sensor de esa
mejora. ArnesIA es el medio de producción; el vendible es el arnés.

> Repo graduado (2026-07-04) de la célula P4 del monorepo `prenter-harness`. Ex-«Harness Studio»
> (nombre muerto por colisión con Harness.io); renombre de la carpeta a `arnesia` pendiente.
> **Visión v3 FIRMADA** (HS-02, 2026-07-04).

## Documentos norte

Todo (salvo skills y `CLAUDE.md`) vive en [`docs/`](./docs/README.md), organizado por el método del
kit `harness@prenter-marketplace` en 3 zonas — **product · architecture · process**. El router
«necesito X → leo Y» es [`CLAUDE.md`](./CLAUDE.md); el front-door del proceso es el skill `/pm`.

- **Dónde estamos · fase · cifras vivas** (generadas) — [`docs/product/checkpoint.md`](./docs/product/checkpoint.md)
- **Qué SABE HACER el sistema hoy** (SSoT funcional, YAML por-cap) — [`docs/product/capabilities/INDEX.md`](./docs/product/capabilities/INDEX.md)
- **Lo que viene / abierto** — [`docs/product/BACKLOG.md`](./docs/product/BACKLOG.md)
- **Historia de decisiones** — [`docs/product/LEDGER.md`](./docs/product/LEDGER.md) (índice → `ledger/HS-NN.md`)
- **Visión / constitución** (11 principios + anatomía A1–A7) — [`docs/product/vision.md`](./docs/product/vision.md)
- **UX firmada** (Command Rail + Organigrama) — [`docs/product/ux.md`](./docs/product/ux.md)
- **Metodología · reglas de negocio · disciplina** — [`docs/process/metodologia.md`](./docs/process/metodologia.md)
- **Arquitectura as-code + stack** — [`docs/architecture/INDEX.md`](./docs/architecture/INDEX.md) · [`docs/architecture/stack.md`](./docs/architecture/stack.md)
- **Estándar as-code por elemento** — [`docs/architecture/knowledge/INDEX.md`](./docs/architecture/knowledge/INDEX.md)

## Estado

**Fase 5 (Implementación) en curso** (fases 1–4 ✓: Visión · UX · Arquitectura · Specs). El estado
vivo, la fase y las **cifras generadas** (`arnesia conformance`) viven en
[`docs/product/checkpoint.md`](./docs/product/checkpoint.md); el port del monorepo sigue gobernado
por el gran plan (nada se porta sin pasar su fase).

## Arnés de construcción

Kit dev: plugin `harness@prenter-marketplace` (canal estable).

## Desarrollo en Windows

El daemon compila y corre nativo en Windows (`go build -o bin\arnesia.exe .\cmd\arnesia`; el
self-update degrada honesto: deja el binario nuevo y pide reiniciar — no hay swap en caliente de
un `.exe` en ejecución). Notas de entorno:

- **Prereqs**: Go ≥1.25 · Python 3 (`python`; el `python3` de WindowsApps es un stub falso — el
  Makefile y lefthook ya sondean ejecutando) · Node ≥20.16 + pnpm 9 · **Git Bash** (los `.sh` se
  corren con Git Bash, NUNCA con el `bash` de WSL del PATH de sistema) · lefthook + golangci-lint
  para los hooks locales.
- **Instalador de escritorio** (`.msi`/`.exe`, Tauri): requiere además Rust toolchain
  `x86_64-pc-windows-msvc` + Visual Studio Build Tools (MSVC + Windows SDK) + WebView2 (ya viene
  en Win11). Tauri no cross-compila: el bundle Windows se produce EN Windows. Sin certificado de
  firma, SmartScreen muestra «Windows protegió tu PC» → «Ejecutar de todos modos» (esperado).
- **Line endings**: `.gitattributes` fija LF; no activar `core.autocrlf=true` (el repo lo
  neutraliza con atributos, pero no lo peleés).
- **Empaquetado portable**: `python scripts/bundle.py --solo-daemon` (Windows) es el espejo de
  `bash scripts/bundle.sh --daemon-only` (Linux) — mismo sello de identidad `X.Y.Z.AAMMDDHHMM`.

### Mapa de puertos (local)

| Puerto | Servicio |
| --- | --- |
| **4200** | daemon `arnesia serve` (API + SPA + OTLP) — hardcodeado en SPA/Tauri/auth por diseño actual |
| **4300** | reservado: cockpit de harness-studio (migración futura — ver plan cockpit-y-windows) |
| 4002 | cockpit de vitalia-app (repo hermano en esta máquina) — no usar |

Privado — © Alpaca Púrpura / Prenter.
