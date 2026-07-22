# PARIDAD — Portafolio · Slice 1 «FE Portafolio» (spec ↔ implementación ↔ verificación)

> Construido por Sonnet 5 (constructor), plan de Fable 5 (arquitecto). Ejecutados T1→T8 en orden.
> T1-T7 ya están en `main` (`7d27a7d`,`cefccb7`,`6ed3fa8`,`d4c9fd6`,`f8fe840`,`7413390`,`dfa82b5`), un
> commit Conventional por ticket, gate local (§M del plan) verde antes de cada commit — ver `git log
> --oneline`. **T8 (esta hoja + capabilities + mockup + cierre documental) queda SIN commitear al
> cerrar esta sesión** por instrucción explícita del orquestador de este turno («NO COMMITEES NI HAGAS
> PUSH») — el commit `feat(portafolio): S1-T8 — …` lo hace el operador junto con (o después de) la
> firma del gate humano. **Firma 🧑‍⚖️ del operador PENDIENTE** — esta hoja NO la simula.
>
> **Verificación final (T8, 2026-07-14):** `go build/vet/test -race` ✓ (0 fail, todos los paquetes OK)
> · `golangci-lint run` **3 issues preexistentes** en `internal/adapters/selfupdate/{updater.go,
> repo_store_test.go}` (gosec G304 + 2 govet shadow) — confirmado con `git log` que datan del commit
> `4e058b6 fix(self-update): repo configurable…` (2026-07-13, AJENO a este paquete; Slice 1 no tocó
> ningún archivo de `selfupdate/`) — el diente no se silencia, se deja visible, no se «arregla» fuera de
> alcance · `go-arch-lint check` OK (0 warnings) · `go test ./docs/architecture/fitness/...` OK (incl.
> los 4 checks R1/R2/R4 de capabilities sobre las hojas recién graduadas) · `conformance --todo`
> **253 checks · pass 42 · fail 0 · error 0 · deferred 211** (SIN regresión: idéntico al baseline
> post-Slice-0) · `conformance --arnes dogfood/dev-full-cycle.graph.json` **21·20·1** (warn
> `art-es-path` preexistente, sin regresión) · `pnpm run verify` (web) verde · `vitest --project=unit
> run` **21/21** · `vitest --project=storybook run` **127/127** (20 archivos, EN FOREGROUND — gotcha
> HS-20; los `console.error` que aparecen en el log son las stories `error-boundary` disparando errores
> a propósito para probar la recuperación, no fallas) · `cap_doctor.py` **92 capabilities válidas** ·
> `estado.sh --check` detectó drift esperado (44→48 vivo, 7→3 stub tras graduar las 4 `fe-portafolio/*`)
> — regenerado como parte del cierre (Paso 5) · **E2E vivo contra la máquina real** (ver §E2E abajo).

## E2E vivo (máquina real, browser + curl, 2026-07-14)

Daemon aislado (`go build -o /tmp/arnesia-e2e-t8 ./cmd/arnesia`, `serve --addr 127.0.0.1:4299
--sessions <tmp> --arneses <tmp> --arnes-root <tmp>`) + `pnpm dev --port 5173` (puerto allowlisted por
el gate CORS/Origin del daemon, `auth.go:AuthConfigFor` — solo `:5173`/`:{addr-port}` están permitidos
en dev, `:5199` NO, se corrigió en el momento) — **el daemon de producción en `:4200` (`arnesia serve
--repo …`, PID 1645682) NO se tocó en ningún momento**, verificado `healthz` antes/durante/después.
`~/.arnesia/portafolio.json` **compartido de verdad** (no hay flag `--portafolio` en `serve`, hallazgo
confirmado leyendo `cmd/arnesia/main.go:newPortafolioService` — el store SIEMPRE resuelve
`~/.arnesia/portafolio.json`, por eso este E2E lo respalda/restaura igual que hizo T7/Slice 0).

1. **Estado inicial honesto:** `#/portafolio` con el store real vacío → "Tu portafolio está vacío." +
   CTA ＋ Agregar (estado `Vacia`, G5). `~/.arnesia/portafolio.json` confirmado AUSENTE antes de tocar
   nada (`ls` → `No existe el archivo o el directorio`).
2. **Wizard → escanear `~/Proyectos/luana-vitalia`** (proyecto real, monorepo — mismo que usó T7 y el
   E2E de Slice 0): 16 candidatos REALES en el paso 2, incluido `harness · Prenter Harness (dev-process
   kit) · v0.5.2 · alpacapurpura/prenter-marketplace · referenciada-cc · en-deriva` (idéntico al hallazgo
   real de Slice 0) y `frontend-design · v unknown · anthropics/claude-plugins-official ·
   referenciada-cc · deriva-no-evaluable`, más 14 `(sin id)` honestos (`v?` · «origen desconocido» ·
   `proyecto-instalado`/`deriva-no-evaluable`, dos de ellos con aviso real de `enabledPlugins` sin
   record). CERO dato fabricado — ningún `drift:N`, ningún `✓` de validación.
3. **Agregar 2 candidatos** (`harness` real + un `(sin id)` proyecto-instalado del root) → `POST
   /api/portafolio/proyectos` real → la Lista se refetchea: **"2 arneses · 0 empresas · lente:
   empresa"** con grupo **"SIN EMPRESA"** (G4 — ninguna empresa inferida del path) conteniendo ambas
   filas; la fila `harness` muestra el chip `en-deriva` (G1) y **CERO** flag `⬆` (G2); dot de salud
   "atención" en ambas (G9, computado — ∃ instalación en-deriva/aviso). Nota real cazada en vivo: los
   dos candidatos `(sin id)` que marqué en el wizard resolvían la MISMA `clave` calificada (mismo
   scope/id vacío del monorepo) — el store los fusionó en una sola entrada con 2 instalaciones
   (`▣ 2 instalac.`), demostrando `mergeInstalaciones`/`TestStoreUpsertMergePorIdentidad` con datos
   reales, no solo con el test unitario.
4. **Drawer del `harness`:** "identidad provisional (sin home) · scope: …" (G6) · facetas EMPRESAS
   "desconocida" / MARKETPLACES "github.com/alpacapurpura/prenter-marketplace" · ley anti-drift ·
   **"update: no-verificado"** literal (G2/S1-D5) · zona Canónico ausente con CTA disabled+tooltip S2 ·
   zona Instalaciones: `referenciada-cc` · `v0.5.2` · chip `en-deriva` · "ORIGEN DE LA COPIA:
   alpacapurpura/prenter-marketplace" · `<details>` trazabilidad expandido a mano → lista real
   `cc-plugins · version · 0.5.2` + `cc-plugins · registry · alpacapurpura/prenter-marketplace` (BR-3) ·
   "⚒ Reparar"/"↩ Backport" disabled+tooltip "próximo · S5" (S1-D11) · línea de frescura "agregado
   2026-07-14".
5. **Abrir en Mapa de la instalación `referenciada-cc` real** (GAP-1, repetición con evidencia fresca):
   click "◉ Observar en Mapa" → navega a `#/mapa` con el picker "Elegir arnés" en `harness` → el grafo
   REAL del plugin renderiza (GUARDIA·hooks, PROCESO, BASE·Knowledge, skill `harness-bootstrap`) desde
   `~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2/` → tab "Contenido" del inspector degrada
   honesto: **"el arnés no tiene directorio registrado — carga la carpeta (PUT /api/arneses/{id}) para
   leer su fuente"**. `curl GET http://127.0.0.1:4299/api/arneses` **ANTES y DESPUÉS**: idéntico,
   `[{"arnes":"e2e-anchor","path":".../harness-studio"}]` — **CERO cwd nuevo registrado** (confirmado,
   no solo asumido).
6. **Corromper `~/.arnesia/portafolio.json` a mano** (`python3`, agregó `{"identidad": "esto rompe el
   shape…"}` al array `entradas`) — el store cachea en memoria al boot (`Store.Listar()` no relee disco,
   confirmado leyendo `store.go`), así que el test real es de REINICIO, igual que hizo Slice 0: reboté
   el daemon aislado CON el archivo corrupto en disco → **booteó igual** (log sin error) → `GET
   /api/portafolio` devolvió `entradas: 2, corruptas: [{"motivo": "json: cannot unmarshal string into Go
   struct field EntradaPortafolio.identidad of type domain.IdentidadArnes"}]}` → recargué `#/portafolio`
   en el browser → banner real **"1 entrada(s) corrupta(s) en el registro — json: cannot unmarshal
   string…"** visible + el resto de la lista sigue viva (BR-11).
7. **Desvincular con confirmación:** click "⊘ Desvincular" del `harness` → dialog real con el copy
   EXACTO "Desvincular harness: solo lo saca del portafolio — no desinstala ni borra nada del disco." +
   checkbox "…y borrar clon local" disabled+tooltip "próximo · S2" (G7) → "Confirmar" → `DELETE
   /api/portafolio/arneses/{clave}` real → la fila desaparece, "1 arneses" en el contador. Segundo
   `DELETE` de la misma clave por `curl` → **`HTTP 404 {"error":"clave no encontrada en el
   Portafolio"}`**.
8. **Recorrido de teclado sin mouse** (G8 manual — el axe automático ya corrió en las 25 stories):
   `Tab` desde el buscador → `Empresa` → `Plano` → (los 4 botones disabled se saltan solos, foco nativo)
   → la fila `(sin id)` restante queda focused → `Enter` → drawer abre con foco inicial en "Cerrar" →
   `Escape` → drawer cierra, foco vuelve a la lista. Cero mouse en todo el recorrido.
9. **Restauración de la máquina:** `~/.arnesia/portafolio.json` **borrado** (estaba AUSENTE antes del
   E2E, confirmado con `ls` al inicio y al final) — no se «restauró contenido», se restauró la
   AUSENCIA, que era el estado real previo. Procesos: daemon aislado (`:4299`) y `pnpm dev` (`:5173`)
   matados explícitamente + un proceso hijo `vite` residual (`kill` del padre `pnpm` no bajó al hijo
   `node .../vite.js`, cazado con `ps`/`ss` y matado aparte) + un `vite --port 5199` residual de mi
   primer intento (puerto no-allowlisted, abandonado) — **verificado con `ps aux`/`ss -ltnp` que
   `:4299`/`:5173`/`:5199` quedaron libres** y que el ÚNICO daemon+dev-server vivos al cierre son los
   del operador (`arnesia serve --repo … :4200` PID 1645682, `pnpm dev --port 3002` PID 3805,
   ninguno de los dos tocado).

## Goal §0 del plan (10 puntos) ↔ realidad

| # | Punto del goal | Estado | Evidencia |
|---|---|---|---|
| 1 | `#/portafolio` renderiza la vista REAL (lente empresa+«sin empresa», lente plano, buscar, `GET /api/portafolio` real, corruptas visibles BR-11, estados vacío/cargando/error/filtro-sin-resultados G5) | ✅ | `global-view.tsx#GlobalView` rutea `portafolio`→`PortafolioView` (T7, `dfa82b5`); `portafolio-list.tsx` (T4); E2E §1/3/6 arriba |
| 2 | Wizard Proyecto/carpeta-local E2E: `POST /escaneos` (spinner+cancelar) → candidatos honestos → `POST /proyectos` → Lista refresca; GitHub/Marketplace disabled+tooltip sin validación fingida | ✅ | `portafolio-wizard.tsx` (T6, `7413390`); E2E §2/3 arriba — 16 candidatos reales, tab Marketplace `aria-disabled`, cero «✓» mostrado |
| 3 | Drawer READ: identidad calificada, facetas honestas, callout anti-drift, zona Canónico, zona Instalaciones (deriva real+detalle+aviso+origen+trazabilidad), «update: no-verificado», acciones staged con tooltip, frescura | ✅ | `portafolio-drawer.tsx` (T5, `f8fe840`); E2E §4 arriba |
| 4 | Desvincular E2E con confirmación (G7): dialog + checkbox disabled S2 → `DELETE` → 404 en repetición | ✅ | confirm interno en `portafolio-drawer.tsx` (T5); `onDesvincular` real en `portafolio-view.tsx` (T7); E2E §7 arriba |
| 5 | Abrir en Mapa E2E incluida instalación `referenciada-cc` real (GAP-1): `POST …/mapa` indexa sin cwd → Mapa renderiza el grafo real → tab fuente degrada honesto; colisión bare-id → confirm (S1-D2); sin sesión → disabled (S1-D13) | ✅ | `usecase.PortafolioService.ObservarEnMapa` (T1, `7d27a7d`) + `onObservarInstalacion`/`ColisionConfirmDialog` (T7); E2E §5 arriba — GET /api/arneses confirma 0 cwd nuevo |
| 6 | Los 9 fixes G1-G9 aplicados y evidenciados (tabla + story/test nombrado, a11y CON test) | ✅ | ver tabla G1-G9 abajo |
| 7 | Storybook = SSoT: stories `portafolio-list`/`portafolio-drawer`/`portafolio-wizard`/chips con `play()`; `vitest --project=storybook run` verde FOREGROUND; `--project=unit run` verde | ✅ | 34 stories (8 chips + 7 list + 11 drawer + 8 wizard — ver conteo real abajo) + `selectors.test.ts`; `127/127` storybook · `21/21` unit (medido T8) |
| 8 | Gates verdes: go build/test -race · golangci-lint · go-arch-lint · fitness · `estado.sh --check` · `conformance --todo` sin regresión (pass ≥42·fail 0) · `pnpm run verify` | ⚠ casi — ver nota | go/arch-lint/fitness/conformance/verify TODOS verdes; golangci-lint tiene 3 issues **preexistentes** ajenos a este paquete (selfupdate, `4e058b6`); `estado.sh --check` detectó drift esperado (regenerado en Paso 5, ver abajo) |
| 9 | Capabilities registradas (T8): 4 `fe-portafolio/*` + extends `navegacion-global`/`desvincular`; mockup corregido G1-G9; `INDEX.md` re-estampado; `paridad.md` con evidencia E2E, firma PENDIENTE | ✅ | ver Paso 1/2 abajo; esta misma hoja |
| 10 | Un commit Conventional por ticket en `main`, gate §M verde ANTES de cada commit | ⚠ parcial — ver nota | T1-T7 commiteados (`7d27a7d`…`dfa82b5`); **T8 queda SIN commitear** por instrucción explícita del orquestador de esta sesión («NO COMMITEES NI HAGAS PUSH») — el commit `S1-T8` lo hace el operador junto a la firma |

**Los 2 puntos con ⚠ NO son gaps de construcción** — son un lint preexistente fuera de alcance (punto 8)
y una restricción explícita de esta sesión de no commitear (punto 10, el trabajo está completo y
verificado, solo falta el acto de commitear que le corresponde al operador). Los 8 puntos restantes
cierran 100 % contra evidencia viva, no solo contra stories.

### Conteo real de stories (Storybook = SSoT, §7)

- `entities/portafolio/ui/chips.stories.tsx` — **8**: `TresEstadosDeriva` · `DerivaConDetalle` ·
  `TresTiposInstalacion` · `AvisoPresente` · `AvisoAusente` · `DotSaludTresRamas` ·
  `EmblemaColorDeterminista` · `EmblemaSinTexto`.
- `widgets/portafolio/ui/portafolio-list.stories.tsx` — **7**: `Vacia` · `Cargando` · `ErrorDeCarga` ·
  `ConDatos` · `LentePlano` · `CorruptasVisibles` · `FiltroSinResultados`.
- `widgets/portafolio/ui/portafolio-drawer.stories.tsx` — **11**: `IdentidadResuelta` ·
  `IdentidadProvisional` · `InstalacionEnDerivaReal` · `UpdateNoVerificado` · `ConDiscrepancias` ·
  `TrazabilidadEslabones` · `ConCanonico` · `SinInstalaciones` · `ObservarSinSesion` ·
  `ConfirmarDesvincular` · `FocoYTeclado`.
- `widgets/portafolio/ui/portafolio-wizard.stories.tsx` — **8**: `Paso1Fuente` ·
  `Paso1FuenteConElegirCarpeta` · `Escaneando` · `CandidatosReales` · `SinHallazgos` · `ErrorDePath` ·
  `Agregando` · `A11yModal`.

Total: **34 stories** con `play()` en las 4 superficies del Portafolio (chips 8 + list 7 + drawer 11 +
wizard 8), todas corriendo dentro del `Frame` `.arnesia-portafolio` (tokens/clases reales) y con el
addon-a11y activo.

## Tabla G1-G9 ↔ evidencia REAL

| G | Fix | Ticket(s) | Evidencia (story/test REAL) | Reproducido en vivo (E2E T8) |
|---|---|---|---|---|
| **G1** deriva fabricada | Chips SOLO desde `EstadoDeriva` real + `deriva_detalle`; muere `drift:N` | T4·T5·T6 | `ConDatos` (list) · `InstalacionEnDerivaReal` (drawer) · `CandidatosReales` (wizard) | ✅ §2/3/4 — `harness` real `en-deriva` con detalle exacto del hash |
| **G2** update fabricado | Fila SIN flag; drawer «update: no-verificado» + tooltip S4 | T4·T5 | `ConDatos` (assert `queryByText(/⬆/) === null`) · `UpdateNoVerificado` (drawer) | ✅ §3/4 — cero `⬆` en toda la sesión; drawer dice literal "update: no-verificado" |
| **G3** marketplace ✓ hardcodeado | Tab+GitHub disabled+tooltip S2, cero resultado fingido | T6 | `Paso1Fuente` (assert disabled+title, cero «✓ marketplace válido» en el DOM) | ✅ §2 — tab Marketplace `role="tab" aria-disabled` en el wizard real, jamás alcanzable |
| **G4** empresa del path | Empresa SOLO del manifiesto; «sin empresa»/«desconocida» | T2·T4·T5 | `agruparPorEmpresa` (selectors.test.ts, unit) · `ConDatos` grupo «sin empresa» · `IdentidadProvisional` (drawer) | ✅ §3/4 — grupo real "SIN EMPRESA" con las 2 entradas agregadas (ninguna inferida de `luana-vitalia`) |
| **G5** estados faltantes | vacío/cargando/error/0-hallazgos/corruptas/sin-instalaciones/sin-resultados | T4·T5·T6 | `Vacia`·`Cargando`·`ErrorDeCarga`·`CorruptasVisibles`·`FiltroSinResultados`·`SinInstalaciones`·`SinHallazgos`·`ErrorDePath` | ✅ §1 (vacío real) · §6 (corruptas reales tras corromper a mano) |
| **G6** keyeado por id desnudo | key = `clave` en todo el FE; identidad provisional VISIBLE | T2·T4·T5 | `types.ts#EntradaPortafolio.clave` + `IdentidadProvisional` (drawer) · `onAbrir(clave)` (ConDatos) | ✅ §4/8 — drawer real "identidad provisional (sin home) · scope: …"; navegación de teclado usa `clave`, no id |
| **G7** borrar-clon sin confirmación | confirm dialog; checkbox borrar-clon disabled S2 | T5 | `ConfirmarDesvincular` | ✅ §7 — dialog real con copy exacto, checkbox disabled, DELETE real solo tras Confirmar |
| **G8** a11y | drawer dialog+trap+Esc · wizard aria-modal · filas focuseables · reduced-motion · axe=CI | T3-T6 | `FocoYTeclado` (drawer) · `A11yModal` (wizard) · asserts de teclado en `ConDatos` · addon-a11y en las 34 stories | ✅ §8 — Tab→fila→Enter→drawer(foco en Cerrar)→Esc→foco devuelto, CERO mouse |
| **G9** dot de salud sin regla | regla S1-D4 definida + testeada + `aria-label` | T2·T3·T4 | `saludDe` (selectors.test.ts, 6+2 ramas) · `DotSaludTresRamas` (chips.stories.tsx) · `ConDatos` | ✅ §3 — dot "atención" real en ambas filas agregadas (en-deriva/aviso), computado, no tecleado |

## Decisiones S1-D1..D22 — estado

Todas resueltas; ninguna quedó abierta.

| # | Resuelta cómo |
|---|---|
| S1-D1 | Endpoint `POST /api/portafolio/arneses/{clave}/mapa` (T1) — cierra GAP-1, verificado en vivo (E2E §5) |
| S1-D2 | `idsColisionados` (selector) + `ColisionConfirmDialog` en la página (T7, S1-D20 formaliza el CÓMO) — deuda de re-key sigue viva, colisión ahora VISIBLE |
| S1-D3 | `AgregarProyecto` puebla `Registries` + `store.Upsert` merge por unión (T1); `TestUpsertMergeUneFacetas` + reproducido en vivo (merge real de 2 candidatos con la misma clave, §3) |
| S1-D4 | `saludDe` (T2, selectors.ts) con 6 ramas testeadas + 2 de grounding real; pintado por `DotSaludPortafolio` (T3) |
| S1-D5 | Línea fija "update: no-verificado" en el drawer (T5); fila sin flag (T4) |
| S1-D6 | Placement FSD-lite tal cual el plan — `entities/portafolio` · `widgets/portafolio` · `pages/shell/ui/portafolio-view.tsx` (T2-T7) |
| S1-D7 | Proyecto vitest `unit` agregado (T2, `vitest.config.ts`); `21/21` verde |
| S1-D8 | Toolbar mínima: buscar + lentes empresa/plano; proyecto/marketplace/filtros disabled+tooltip (T4) |
| S1-D9 | Input path siempre + «Elegir carpeta…» solo Tauri; `AbortController` real del escaneo (T6/T7) |
| S1-D10 | Badge «ya en el portafolio» checkbox habilitado; copy propio de `es_canonico` (T6) |
| S1-D11 | Reparar/Backport disabled+tooltip "próximo · S5" (T5/T6) — reproducido en vivo §4 |
| S1-D12 | Mockup corregido en sitio + `mockups/INDEX.md` re-estampado (T8, Paso 2 de esta hoja) |
| S1-D13 | `onObservar` undefined sin sesión activa, tooltip exacto (T7); demostrado con `ObservarSinSesion` story — en el E2E siempre había sesión activa (seed), así que la rama disabled se probó por story, no en vivo |
| S1-D14 | «origen de la copia» en la UI, `deriva` intacto, vocabulario L0 no pisado (T5) |
| S1-D15 | Línea de frescura "datos del último escaneo — agregado <fecha>" (T5) — reproducida en vivo §4 |
| S1-D16 *(nueva, T1)* | `capability_trace_test.go#capMetas` no parsea YAML real — scenarios BDD sin campo `status:` por-scenario en `observar-en-mapa.yaml` (deuda para BACKLOG: arreglar el parser) |
| S1-D17 *(nueva, T2)* | `fsd/inconsistent-naming` de steiger es falso-positivo sobre nombres ES — apagado global con razón inline en `steiger.config.ts` |
| S1-D18 *(nueva, T5)* | Primitivo de foco propio `shared/lib/focus-trap.ts` en vez de `@base-ui-components/react/dialog` (conflicto con `within(canvasElement)` del patrón de test del repo) |
| S1-D19 *(nueva, T6)* | Esc/✕ del wizard NO cierran durante `estado==="agregando"` (criterio delegado, decisión tomada y documentada) |
| S1-D20 *(nueva, T7)* | Confirm de colisión de id vive en la página (`ColisionConfirmDialog`), no en el contrato cerrado del drawer |
| S1-D21 *(nueva, T7)* | Error de `onObservar` = banner propio de la página (`observarError`), sin slot en el contrato |
| S1-D22 *(nueva, T7)* | Error de `agregarProyecto` reusa el slot `error` del wizard (vuelve a `"candidatos"`) |

## Desviaciones visibles (a firmar en gate humano 🧑‍⚖️)

Mismo concepto que las 5 de Slice 0 — documentadas EN EL MISMO TURNO por cada ticket, listadas acá
consolidadas (son las S1-D16..D22 de la tabla arriba, más 2 encontradas en T8 mismo):

1. **S1-D16** — el enforcer Go de capabilities (`capMetas`) es un parser línea-a-línea que NO entiende
   YAML anidado: un `status:` por-scenario (forma BDD completa del template) le pisa el `status:` real
   del root. Ningún capability existente lo usaba; `observar-en-mapa.yaml` (T1) fue el primero en usar
   la forma BDD completa y tuvo que omitir el campo `status:` por-scenario para no romper el gate.
   **Deuda para BACKLOG**, no se tocó el enforcer (fuera de alcance de T1).
2. **S1-D17** — steiger (`fsd/inconsistent-naming`) clasifica mal los nombres de slice en español
   (heurística de pluralización EN). Apagado global, mismo patrón que el apagado ya existente de
   `fsd/insignificant-slice`.
3. **S1-D18** — no se usó `@base-ui-components/react/dialog` (ya dependencia) para el drawer modal
   porque su `Dialog.Portal` por-default monta fuera de `canvasElement`, rompiendo el patrón de test de
   TODO el Storybook del repo. Primitivo propio chico en su lugar, sin deps nuevas.
4. **S1-D19** — criterio delegado por el plan (Esc durante "agregando"): se decidió BLOQUEAR el cierre
   mientras el POST está en vuelo (mismo espíritu que `UpdateCard`).
5. **S1-D20/D21/D22** (T7) — 3 casos del mismo patrón: el contrato de los widgets (§2.6, cerrado en
   T5/T6) no tenía slot para colisión-de-id / error-de-observar / error-de-agregar. Resueltos SIEMPRE
   por fuera del contrato (en la página), nunca ampliando un contrato ya cerrado a mitad de build.
6. **Gate golangci-lint con 3 issues preexistentes** (nueva, T8): `internal/adapters/selfupdate/
   {updater.go,repo_store_test.go}` traen 1 gosec (G304) + 2 govet (shadow) que datan de `4e058b6`
   (2026-07-13, AJENO a este paquete — Slice 1 no tocó ningún archivo de `selfupdate/`). Se deja
   visible, NO se corrige fuera de alcance de este ticket ni se silencia con `nolint`.
7. **`estado.sh --check` detectó drift esperado** (nueva, T8): graduar las 4 `fe-portafolio/*` de
   `stub`→`vivo` cambió la distribución de capabilities (44→48 vivo, 7→3 stub) — regenerado como parte
   del cierre documental (Paso 5), no es un bug, es la consecuencia directa y esperada de R4.
8. **T8 queda sin commitear** (nueva, T8): por instrucción explícita de esta sesión («NO COMMITEES NI
   HAGAS PUSH»), todo el trabajo de T8 (capabilities completadas, mockup corregido, esta hoja, cierre
   documental) vive en el working tree pero NO en un commit — a diferencia de Slice 0 (T9 sí se
   commiteó antes de la firma). El operador decide cuándo commitear (antes o junto con la firma).

## Amendment pre-firma (2026-07-15) — S1-D23

El operador revisó este `paridad.md` en vivo (gate humano AÚN pendiente) y pidió 2 cambios sobre el
paso Fuente del wizard antes de firmar: modo explícito Escribir-ruta/Elegir-carpeta (input pasa a
solo-lectura en modo elegir, «Escanear» gated por ruta no-vacía) + look ButtonGroup para los radios.
Implementado y documentado en `decisiones.md` **S1-D23** (supersede S1-D9). Re-verificado tras el
cambio: `pnpm run verify` (typecheck·lint·depcruise·fsd·stylelint) verde · `vitest --project=storybook
run` **127/127** (foreground) · `vitest --project=unit run` **21/21** · `cap_doctor.py` **92
capabilities válidas** · `go test ./docs/architecture/fitness/... -run TestCapability` **4/4 PASS**
(R1/R2/R4 sin regresión). Capability `fe-portafolio/wizard-agregar-proyecto.yaml` extendida (nuevo
`change_log` + scenario `wizard-fuente-modo-escribir-o-elegir`), sin tocar entradas históricas. Deuda
diferida: `mockups/arnesia-portafolio.html` queda sin corregir para este modo (registrado en S1-D23).
La tabla del goal §0 y G1-G9 arriba sigue vigente sin cambios de fondo — este amendment no reabre
ningún punto ya cerrado, solo extiende el paso Fuente descrito en el punto 2.

## Firma

- [x] 🧑‍⚖️ **Gate humano — FIRMADO 2026-07-22** (operador): *"ya lo vi, firma vos nomás"* — confirma
  haber revisado en vivo el resultado (click-through previo a esta sesión de cierre). Respaldo: goal §0
  del plan (8/10 ✅, 2 ⚠ explicados arriba, ninguno es gap real de construcción) + tabla G1-G9 + el E2E
  vivo de esta hoja + las 8 desviaciones S1-D16..D22 + golangci-lint preexistente — todas aceptadas sin
  objeción. Nota: T8 ya estaba commiteado (`b35ff6c`) al momento de esta firma, igual que el amendment
  S1-D23..D26 (`4a35479`). Cierre: `checkpoint.md`/`BACKLOG.md` actualizados, `ledger/HS-25.md`.
