# GENERICIZE-NOTES — cockpit · estado: F-4 EJECUTADO (2026-06-10)

> Origen: copia 1:1 de `tools/luana-cockpit` (upstream @ 968857a5) sin node_modules/.next. Standalone Next.js 16 + filesystem-as-DB (lee `.md`/`.yaml` del workspace) · sin Docker · sin Postgres. `pnpm install && WORKSPACE_ROOT=<repo> pnpm dev`.

## Resuelto en F-4 (2026-06-10)

| Pieza | Fix aplicado |
|---|---|
| `lib/agent-meta.ts` | roster hardcodeado eliminado → `AgentId = string` + `agentMetaOf()` determinístico (color por hash del slug, emoji por rol builder/auditor/supervisor, name capitalizado). `agentOf()` sigue funcionando (agent_owner → owner → prefijo cap_target; el fallback por regex de story_id se retiró — requería roster). |
| `lib/map-zones.ts` | `VALUE_STREAM_STAGES` hardcodeado → `buildProcessLens(tree, stages?)` parametrizado + `DEFAULT_VALUE_STREAM_STAGES` genérico (descubrir/construir/operar/mejorar, sin agentes). Etapas reales: slot `value_stream` del seam vía `/api/value-stream` (nuevo) + `getValueStreamStages()` en `project-config.ts` (normaliza objetos/string[]/flat, filtra boxIds contra `agent_roster`). |
| `components/map/MapView.tsx` | `VITALIA_ROLES` eliminado → roles derivados del data (`access.entry_points[].requires_role` de las caps; sin roles → el select no se muestra). `FALLBACK_AGENTS_BY_BRAND` eliminado → la vista legacy deriva agentes de los `agent_owner` presentes en el data + metadata de agent-meta. |
| `lib/chris-input-parser.ts` | renombrado a `lib/operator-input-parser.ts`. Serializa autor `operador`; parsea `operador\|chris` (legacy F-1). Filenames: `operator-input.md` + legacy `chris-input.md`. Ruta API `app/api/chris-input/` → `app/api/operator-input/`; `ChrisInputTab` → `OperatorInputTab`; api-client renombrado. |
| `lib/types.ts` | `CHRIS_ALLOWED_TRANSITIONS`→`OPERATOR_ALLOWED_TRANSITIONS`, `ChrisVerify/ChrisSignoff/ChrisInput*`→`Operator*`, `isChrisAllowed`→`isOperatorAllowed`, `SpecialistAgentId`/`AgentOwner` vitalia → `string`. Keys serializadas del contrato kit conservadas con comentario `// legacy F-1`. |
| `lib/workspace.ts` | `KNOWN_BRAND_SLUGS` eliminado → brands = `brands.active` del seam + discovery en disco (`{slug}/docs/product/`). `brandPath` valida slug sano + bootstrapeado (sin lista). Igual en `app/api/_lib/responses.ts` (whitelist /api/file). `lib/story-paths.ts`: heurística ancla en `{brand}/docs` con brand del caller. |
| `components/providers/BrandProvider.tsx` | `FALLBACK_BRAND` 'vitalia' → 'main'. `app/layout.tsx` fallback igual. |
| `app/layout.tsx` | metadata → "Cockpit · SDD". Sidebar/Header/package.json/.env.local.template genericizados. |
| `components/modals/MergeReleaseModal.tsx` | mapa de puertos FE hardcodeado eliminado → prompt de verificación genérico que apunta al slot `toolchain` del seam. |
| Tests `lib/__tests__/` | fixtures genéricas ('main'/'acme'/agentes alfa-eco); `chris-input-parser.test.ts` → `operator-input-parser.test.ts` (fixture inline, sin dependencia del workspace luana); `project-config.test.ts` self-contained (seam fixture en tmpdir); smokes contra `docs/process/*.md` del adopter → `it.skipIf` cuando no hay workspace. |
| Textos/labels | Chris → operador en UI/comentarios; `/pm-luana` → `/pm`; ADR-vitalia-NNN → genérico. |

**Verificación 2026-06-10:** `npx tsc --noEmit` → 0 errores · `npx vitest run` → 17 files / 148 passed, 2 skipped (smokes sin workspace) · grep `luana|vitalia|nicolify|comunify|lupulo|chris` en código/UI → solo literales legacy documentados (abajo).

## Literales legacy que QUEDAN (contrato con el kit read-only · F-1)

- `chris_verify` key (lib/types.ts · story-closure-gate del kit) — tipo TS `OperatorVerify`.
- `ratified_by_chris` key (lib/types.ts + story-templates · 01-spec-template del kit).
- `phase: AWAIT_CHRIS_VERIFY` (valor del named-phase v5 · comentarios en types.ts).
- Filename `chris-input.md`: aceptado al leer (parser/rutas/edit-permissions/watcher) y usado al CREAR stories nuevas (story-templates espeja `00-chris-input-template.md` del kit, incl. heading y pointer a `chris-input-protocol.md`).
- Alias autor `'chris'` aceptado al parsear conversación / PATCH API (se serializa siempre `operador`).

Cada uno marcado en código con `// legacy F-1`. Cuando F-8 resuelva el SSoT y el kit renombre template/keys, retirar estos alias en un solo barrido.

## Pendientes reales (no bloquean F-4)

- **Detección de marca por worktree + mapa de puertos 4000-4004** (`scripts/cockpit-up.sh` upstream): el script NO vino en la copia — el cockpit ya arranca genérico (`pnpm dev` + env). Si el kit quiere `cockpit-up` genérico, derivar puerto de `brands[].ports.cockpit` del seam (slot ya tipado en project-config.ts).
- **Vistas cross-brand (portfolio consolidado :4000):** siguen funcionando como "multi-brand selector"; con 1 brand degradan solas (selector de 1). No hay vista portfolio dedicada en esta copia — nada que degradar.
- **Smoke en adopter ficticio** (criterio done F-4, patrón W8): apuntar `WORKSPACE_ROOT` a un adopter generado por el installer en `/tmp` y verificar board/stories/caps/harness renderizando. Hecho a nivel unit (fixtures tmpdir en tests); falta el pase visual end-to-end con `pnpm dev`.
- `app/api/debug/*` (si existe tras NEXT_PUBLIC_DEV_MODE) no auditado a fondo — sin tokens según grep.
