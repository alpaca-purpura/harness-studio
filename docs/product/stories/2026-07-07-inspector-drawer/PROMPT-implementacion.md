# PROMPT de implementación — paquete inspector-drawer (copiar a una conversación nueva)

> Persistido por disciplina §10: la conversación jamás es el único registro.
> Paquete FIRMADO por el operador el 2026-07-07.

---

Implementa el paquete FIRMADO `historias/2026-07-07-inspector-drawer/` (drawer/inspector
del Mapa) con rigurosidad y en fases. El código no se toca hasta cerrar la Fase 1.

**FASE 0 — Contexto (leer, en este orden):**
1. `historias/2026-07-07-inspector-drawer/INDEX.md` (flujo, estado, «Retomar aquí»)
2. `decisiones.md` (6 decisiones firmadas — son LEY) · `spec.md` (RF-80..96 + Gherkin +
   trazabilidad) · `design.md` (UI al pixel) · `PARIDAD.md` (16 filas a llenar) ·
   `spike-medicion-contrato.md` (contexto del debate #6) · `analisis-drawer-v3.md`
3. `mockup-drawer.html` v6 = la verdad visual (ábrelo en el browser; artifact en INDEX)
4. Código existente: `web/src/widgets/map-canvas/ui/inspector.tsx` (punto de partida) ·
   `web/src/entities/arnes/` (model: types/kind/node-view/proposals/selectors · ui) ·
   `web/src/widgets/map-canvas/ui/map-canvas.tsx` (onSelect/selección) ·
   `web/src/shared/api/client.ts` · Go: `internal/adapters/transport/http/router.go` ·
   `internal/usecase/` · `internal/adapters/loader/` · `internal/domain/box.go`
5. Arquitectura que RIGE: `docs/architecture/INDEX.md` + boundaries `fe-taxonomia-componentes`
   (canvas⊥chrome) · `fe-transporte-independiente` · `fe-topologia-fsd` ·
   `superficie-local-confinada` (el endpoint nuevo NO puede salir del dir del arnés) ·
   `docs/architecture/conventions/` (estilo Go/TS, commits) · `docs/architecture/contracts/` (OpenAPI + schemas)

**FASE 1 — Revisión + diseño técnico (sin código):** contrasta cada RF del spec contra
lo existente; decide dónde vive cada pieza cumpliendo la arquitectura: diccionario
doctrinal → `entities/arnes/model/doctrina.ts` (la entity es dueña; el widget consume);
edges inversos → selector en `entities/arnes/model/selectors.ts` sobre `graph.edges`;
tabs/expandido/chips/hallazgos → `inspector.tsx` + CSS con tokens (`var(--…)`, stylelint
anti-magic-value); navegación de chips → el `onSelect` que el canvas ya expone (jamás
estado global nuevo sin necesidad); endpoint fuente → puerto/usecase/adapter hexagonal +
ruta en router.go + **changelog OpenAPI** (`docs/architecture/contracts/`), lectura CONFINADA al dir
registrado del arnés (S2, valida traversal), 404 sin `fuente_path`; conformance por nodo
→ filtra el endpoint existente `GET /api/harnesses/{id}/conformance` en FE o añade query
param (elige y justifica). Emite un mini-plan de commits (orden, qué toca cada uno) y
séllalo en el INDEX del paquete antes de codear.

**FASE 2 — Implementación por RF (orden del spec):** RF-80..84 shell → RF-85..88
tooltips → RF-89..92 Resumen → RF-93..95 Contenido (incluye Go) → RF-96 Corridas.
Story=test por cada marca/comportamiento nuevo (Storybook 10, `vitest --project=storybook`;
sigue las convenciones de `inspector.stories.tsx`). Gates que DEBEN quedar verdes:
`tsc --noEmit` · `biome ci src` · `depcruise` · `steiger` · `stylelint` · story-tests ·
`go build/vet/test -race ./...` · golangci-lint · go-arch-lint. Commits al estilo del
repo (`feat(HS-09): …`), trunk-based a main, uno por bloque coherente de RFs.

**FASE 3 — Validación REAL (no negociable, funcional Y visual):** levanta daemon
(`go run ./cmd/arnesia serve`, :4200) + vite (`pnpm dev` en web/, :5173); con **Chrome
DevTools MCP** recorre CADA RF en la app real contra el showcase (picker «Editorial ·
Content Lead») Y el dogfood dev-full-cycle: click nodo→drawer, ⤢/⤡/✕/Esc (cerrar ≠
colapsar), las 3 tabs, tooltips por hover Y teclado, chips que navegan y seleccionan el
nodo, hallazgo gate:none en draft-caja, Contenido sirviendo el ARCHIVO REAL del dogfood
vía el endpoint nuevo, Corridas honesta, expandido (tabs sticky tras scroll, botonera
fila, columna 920). Compara LADO A LADO contra el mockup v6. Screenshots al scratchpad,
**consola limpia** (0 errores/warnings). Marca cada fila de `PARIDAD.md` → ✅ solo tras
verificarla así; si algo no se puede cumplir tal cual, NO lo maquilles: regístralo en
PARIDAD como desviación y consúltame.

**Reglas permanentes:** disciplina METODOLOGIA §10 (todo hallazgo/decisión → archivo del
paquete EN EL MISMO TURNO; «Retomar aquí» al día al cerrar) · lo staged queda disabled y
rotulado, jamás finge funcionar · tokens DTCG only · contraste AA en ambos temas ·
alcance = el spec, ni un RF más (los no-goals están escritos) · honestidad: lo que no
hay, se dice.

Al terminar: PARIDAD completa, gates verdes listados con su salida, resumen honesto de
desviaciones, commit/push, INDEX «Retomar aquí» actualizado.
