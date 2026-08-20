# Cockpit · SDD

Visualizador + editor del workflow **Spec-Driven Development** del dev-process kit. Standalone Next.js 16, **filesystem-as-DB**: lee los `.md`/`.yaml` del workspace del adopter directo del disco (sin Docker, sin Postgres) y permite editar metadata vía forms + el panel `operator-input` conversacional.

## Qué es

- **Board** kanban de stories (10 estados del paradigm v4: idea → done) leído de `{brand}/docs/product/stories/*/checkpoint.md`.
- **Roadmap** por releases (`{brand}/docs/product/releases/*.yaml`).
- **Mapa Implementado** de capabilities (`{brand}/docs/product/capabilities/**.yaml`) sobre el esqueleto de `{brand}/docs/architecture/SYSTEM-MAP.yaml`, con 2 lentes: trabajadores (zonas) y proceso (value-stream).
- **Arquitectura** (SYSTEM-MAP global), **Drift** (caps no verified-live · carril L4), **Learnings** (carril L2) y **Harness · CIL** (board transversal L1+L3 de `docs/process/harness-backlog.md` + `docs/process/tech-debt.md`).
- **Story Drawer** (checkpoint + operator-input + spec/diseño/arq/audit/files) y **Cap Drawer** (ledger + scenarios + acceso + reglas + validación bidireccional).
- Hot-reload vía chokidar + SSE: cuando un skill escribe un checkpoint/cap/release, la UI se refresca sola.

El cockpit **no inventa datos**: renderiza el read-schema del proceso (estados, fases, `chris_verify`, `dod_evidence`, `repro_evidence`, etc. — ver `lib/types.ts`) tal como lo producen los skills del kit.

## Cómo correr

Requisitos: Node 20+ y pnpm 9+.

```bash
cd cockpit
pnpm install
WORKSPACE_ROOT=/ruta/al/workspace/adopter pnpm dev
# → http://localhost:4000
```

Si el cockpit vive DENTRO del workspace del adopter, `WORKSPACE_ROOT` es opcional (autodetect vía `git rev-parse --show-toplevel`). Config opcional en `.env.local` (ver `.env.local.template`): `WORKSPACE_ROOT`, `DEFAULT_BRAND`, `EDITOR_BIN`, `PORT`.

```bash
pnpm dev          # dev server :4000 con hot reload + chokidar SSE
pnpm build        # Next.js build standalone
pnpm start        # producción :4000 (post-build)
pnpm test         # vitest run
pnpm typecheck    # tsc --noEmit
```

## Qué lee del seam (`project.config.yaml`)

El cockpit es genérico: **todo lo configurable sale del seam** del workspace (lector server-only: `lib/project-config.ts`). Slots que consume:

| Slot | Uso en el cockpit | Fallback si está `__FILL_ME__` |
|---|---|---|
| `brands.active[].slug` | Marcas del selector (+ discovery en disco de todo `{slug}/docs/product/`) | discovery en disco · default UI `main` |
| `agent_roster` (per-brand o flat) | Nombre/emoji/color de los agentes del board/mapa | metadata determinística por slug (`lib/agent-meta.ts`: color por hash, emoji por rol, name capitalizado) |
| `value_stream` (per-brand o flat) | Etapas de la lente "proceso" del mapa (vía `/api/value-stream`) | 4 etapas genéricas: descubrir · construir · operar · mejorar (sin agentes asignados) |
| `toolchain` / `brands[].ports` | Referenciados en el prompt de verificación del merge-release (no se hardcodean comandos ni puertos) | el prompt instruye leer el seam |

Workspace esperado (lo produce `installer/new-project.sh`):

```
{workspace}/
├── project.config.yaml          # el seam
├── docs/                        # platform-level (pseudo-marca «⬡ Platform · core»)
│   └── process/                 # harness-backlog.md · tech-debt.md (CIL)
└── {brand}/docs/
    ├── product/{stories,capabilities,releases}/
    ├── architecture/SYSTEM-MAP.yaml
    ├── archive/{year}/stories/
    └── learnings/
```

## Contratos con el kit (literales legacy · F-1)

El kit `core-harness/` es read-only; estas keys/filenames **serializados** se mantienen tal cual (los nombres TS sí se genericizaron a `Operator*`):

- `chris_verify` (checkpoint frontmatter · fase G del story-closure-gate) → tipo TS `OperatorVerify`.
- `ratified_by_chris` (frontmatter de spec/checkpoint).
- `phase: AWAIT_CHRIS_VERIFY` (named-phase v5).
- Filename `chris-input.md` (template del kit) — el parser (`lib/operator-input-parser.ts`) y las rutas aceptan también `operator-input.md`; el autor se serializa siempre como `operador` (acepta `chris` legacy al parsear).

## Permisos de edición

`lib/edit-permissions.ts` es la SSoT: `checkpoint.md` siempre editable; `operator-input` y capabilities append-only vía UI dedicada; spec/design/arch/validators/tickets se lockean desde `developing`; archive y `07-merge.md` read-only. Transiciones de estado desde la UI: solo el whitelist `OPERATOR_ALLOWED_TRANSITIONS` (idea ↔ refining + parked/dropped con razón); el resto las hacen los skills del proceso.

## API (resumen)

Todas en `app/api/` · JSON + Zod + whitelist anti-traversal (`app/api/_lib/responses.ts`):

| Path | Función |
|---|---|
| `/api/stories` · `/api/stories/[id]` | Aggregate active+archive · edit de fields editables por el operador |
| `/api/releases` · `/api/merge-release` | CRUD releases · cierre a shipped con gate de verificación |
| `/api/operator-input/[storyId]` | Append notas/refs/conversación (acepta filename legacy `chris-input.md`) |
| `/api/capabilities*` | Caps + status computado + code-index + validación bidireccional |
| `/api/system-map` · `/api/value-stream` | Esqueleto del mapa · etapas del seam |
| `/api/transition` · `/api/extend-cap` · `/api/from-done` | Whitelist transitions · spawn de stories |
| `/api/file` · `/api/open` · `/api/refs/upload` · `/api/watch` · `/api/sessions` | Read/write whitelisted · editor externo · uploads · SSE · build-claims |

## Tests

```bash
npx tsc --noEmit && npx vitest run
```

Suite self-contained (fixtures genéricas en tmpdir — no requiere workspace real). Los smokes contra `docs/process/*.md` del adopter se saltan si no hay workspace presente.

## Stack

Next.js 16 App Router · React 19 · Tailwind v4 · @dnd-kit (drag-drop) · gray-matter + react-markdown (frontmatter/render) · CodeMirror YAML · chokidar + SSE · simple-git · zod · vitest.
