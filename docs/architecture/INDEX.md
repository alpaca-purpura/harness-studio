# arch/ — arquitectura y diseño técnico as code de ArnesIA

> **La arquitectura vive aquí, no en un doc que se pudre.** Cada regla arquitectónica
> estructural (boundary) = un nodo con **L1** (principio con fuente) ↔ **L2** (realización en
> este árbol Go) + una **tabla de checks ejecutables**. La parte «por qué» vive en el
> [`../LEDGER.md`](../product/LEDGER.md) (fichas firmadas, sin duplicar); la parte «prueba» vive en
> [`fitness/`](./fitness/); el «cómo se ve» en [`model/`](./model/); el «contrato de datos» en
> [`contracts/`](./contracts/). Espeja el patrón de [`../docs/architecture/knowledge/`](knowledge/INDEX.md).
> Norte: [`../product/vision.md`](../product/vision.md) · evidencia L1:
> [`../historias/2026-07-05-arquitectura-fase3.md`](../product/research/2026-07-05-arquitectura-fase3.md).

## Cómo se relaciona con el resto del repo

- [`../product/vision.md`](../product/vision.md) — constitución (11 principios + anatomía A1–A7) y **decisiones
  técnicas fundacionales** (HS-02). El **norte**; esta capa las aterriza y las enforça.
- [`../process/metodologia.md`](../process/metodologia.md) — reglas de negocio. El `contract:` de caja (§3) es
  **el mismo schema** que valida [`contracts/schema/box.contract.schema.json`](./contracts/schema/box.contract.schema.json).
- [`../docs/architecture/knowledge/`](knowledge/INDEX.md) — estándar as code por **elemento de arnés** (138
  checks · 12 nodos, incl. `harness-profile`). `arch/` es el gemelo: estándar as code de **la app
  ArnesIA misma** (dogfood). Un solo
  runner (`arnesia conformance`) corre ambos árboles — **construido en HS-08**: el `RulesetPort` parsea
  `docs/architecture/knowledge/` + `arch/` = 247 checks a datos (`--todo`, medido 2026-07-09) y el `ConformancePort` los corre con adapters por mecanismo
  (arch-test/schema-validation/go-arch-lint/static-scan/nl-judge). El **contrato de check común** (mecanismo
  por check) ya existe; los `enforced_by:` de `arch/` y los checks de `docs/architecture/knowledge/` corren por el mismo motor.
- [`../LEDGER.md`](../product/LEDGER.md) — **índice del diario firmado** → `ledger/HS-NN.md` (una ficha por
  archivo). Cada boundary cita su ficha en `ledger:`. El **backlog y el estado NO viven aquí**:
  pendientes en [`../BACKLOG.md`](../product/BACKLOG.md), cifras (generadas) en [`../checkpoint.md`](../product/checkpoint.md);
  [`../CLAUDE.md`](../../CLAUDE.md) = **router de punteros**, no repositorio de estado/historia/stack.
- [`../CAPABILITIES.md`](../product/capabilities/INDEX.md) — **SSoT de lo funcional** (qué HACE la app, por
  capability). El boundary [`codigo-traza-a-capability`](./boundaries/codigo-traza-a-capability.md)
  lo enforça: todo archivo fuente traza a ≥1 capability, cero huérfanos.

## Las dos capas (en cada boundary node)

1. **L1 · Principio** — el patrón de arquitectura como estándar de industria (hexagonal,
   ports&adapters, local-first, event-sourcing…), fechado y con fuente. Evidencia, no opinión.
2. **L2 · Realización** — cómo se mapea en ESTE árbol Go+React: qué paquetes son core/shell/
   dominio/transporte/adaptador, y cómo la regla se encarna aquí.

Y cada nodo emite un **`Checklist evaluable`**: la rúbrica as code (checks con severidad + señal)
que un linter (`go-arch-lint`/`depguard`/validación de schema/`arch_test.go`) corre para probar
que el código NO viola la arquitectura. Cada check declara su `enforced_by:` — el link 1:1 a la
check ejecutable que lo guarda (esto ya rige en `arch/`; en `docs/architecture/knowledge/` el `enforced_by:` por
check llega con HS-08, ver runner unificado arriba).

## El árbol — boundary nodes

| Nodo | Regla | Estado | Versión | Checks | Enforcer |
|------|-------|--------|---------|--------|----------|
| [`boundaries/core-no-importa-shell.md`](./boundaries/core-no-importa-shell.md) | El daemon-core no depende del shell (Tauri) | 🌱 vivo | 1.2 | 5 | go-arch-lint · arch_test.go · lib.rs (revisión) |
| [`boundaries/dominio-independiente-de-transporte.md`](./boundaries/dominio-independiente-de-transporte.md) | El dominio no depende de HTTP/SSE/store | 🌱 vivo | 1.1 | 4 | go-arch-lint · depguard |
| [`boundaries/adaptadores-de-agente-intercambiables.md`](./boundaries/adaptadores-de-agente-intercambiables.md) | Claude Code = un adaptador tras `AgentPort` | 🌱 vivo | 1.0 | 4 | go-arch-lint · arch_test.go |
| [`boundaries/indice-desechable-jsonl-es-verdad.md`](./boundaries/indice-desechable-jsonl-es-verdad.md) | Árbol de arneses = verdad; el índice (SQLite real desde fase 5) es reconstruible, nunca migrado | 🌱 vivo | 1.3 | 4 | arch_test.go · schema |
| [`boundaries/conductor-no-parsea-jsonl.md`](./boundaries/conductor-no-parsea-jsonl.md) | El conductor consume stream-json/OTel, no parsea JSONL | 🌱 vivo | 1.0 | 4 | depguard · arch_test.go |
| [`boundaries/permisos-gui-human-in-the-loop.md`](./boundaries/permisos-gui-human-in-the-loop.md) | Deny-by-default; el GUI aprueba cada write vía diff · **+ sesión aislada por cwd** | 🌱 vivo | 1.1 | 7 | arch_test.go |
| [`boundaries/superficie-local-confinada.md`](./boundaries/superficie-local-confinada.md) | La API local está confinada (Host+Origin) y autenticada (token del shell) · + eje config-source del spawn | 🌳 enforced | 1.3 | 9 | arch_test.go · auth_test.go |
| [`boundaries/sesion-viva-consistente.md`](./boundaries/sesion-viva-consistente.md) | El pipe conductor↔dock: guardado · sin pérdida · idempotente · auto-sana · **+ la transición de conversación es atómica** | 🌳 enforced | 1.2 | 5 | arch_test.go (**5/5 verdes**; `TestTransicionDeConversacionEsAtomica` escrito en T15) |
| [`boundaries/contrato-de-caja-es-fitness-function.md`](./boundaries/contrato-de-caja-es-fitness-function.md) | Validar el `contract:` de caja contra su schema + composición del cableado (huérfanos·dead-ends·rutas·refina) | 🌳 enforced | 1.3 | 7 | schema · arch_test.go:TestBoxContractValidatesAgainstSchema · domain.Verificar{SinHuerfanos,DeadEnds,RutaExiste,RefinaCoherente} (ruta `--arnes`) |
| [`boundaries/fe-topologia-fsd.md`](./boundaries/fe-topologia-fsd.md) | La SPA se estructura por FSD (import direccional) | 🌱 vivo | 1.0 | 5 | dependency-cruiser · steiger |
| [`boundaries/fe-taxonomia-componentes.md`](./boundaries/fe-taxonomia-componentes.md) | Taxonomía por dirección/pureza, no ladder atómico (canvas⊥chrome · widget⊥widget) | 🌳 enforced | 1.2 | 5 | dependency-cruiser (verde: 122 módulos / 0 violaciones) |
| [`boundaries/fe-transporte-independiente.md`](./boundaries/fe-transporte-independiente.md) | Dominio FE ⊥ transporte; SSE singleton en `app` | 🌱 vivo | 1.1 | 4 | dependency-cruiser |
| [`boundaries/fe-tokens-contrato.md`](./boundaries/fe-tokens-contrato.md) | Tokens DTCG = contrato mockup↔código, cero magic-value | 🌳 enforced | 1.1 | 4 | stylelint · tokens-sync |
| [`boundaries/fe-visual-fitness.md`](./boundaries/fe-visual-fitness.md) | Story = test que rompe CI (fitness visual local) | 🌳 enforced | 1.1 | 5 | vitest.config.ts + Storybook 10 (story=test; conteo crece con cada story — ≥45 verdes HS-11) |
| [`boundaries/orquestacion-determinista-entre-cajas.md`](./boundaries/orquestacion-determinista-entre-cajas.md) | La secuencia entre cajas es código; la agencia vive dentro (framed autonomy) | 🌳 enforced | 1.1 | 4 | arch_test.go:TestConductorOwnsBoxRouting |
| [`boundaries/permisos-derivan-del-rol.md`](./boundaries/permisos-derivan-del-rol.md) | El permission-set se parametriza por el rol que hidrata (autoridad externa) | 🌳 enforced | 1.1 | 4 | arch_test.go:TestPermissionSetParametrizedByRole |
| [`boundaries/codigo-traza-a-capability.md`](./boundaries/codigo-traza-a-capability.md) | Todo código fuente traza a un capability (`docs/product/capabilities/` = SSoT funcional), incl. el `#Símbolo` del puntero | 🌳 enforced | 1.5 | 6 | capability_trace_test.go:TestCapabilityPointersResolve · TestCapabilityPointerSymbolsResolve · TestCapabilityCoverage · TestCapabilityStatusConsistent · TestCapabilityPointersStable |
| [`boundaries/portafolio-identidad-y-deriva-honesta.md`](./boundaries/portafolio-identidad-y-deriva-honesta.md) | Identidad `(home,id)` calificada (nunca fusión por coincidencia) + deriva por hash de contenido (nunca semver-string) + el estante no miente (`null`≠`[]`, `no-comparable` de 1ª clase) | 🌳 enforced | 1.2 | 8 | portafolio/store_test.go · portafolio/deriva_test.go · domain/portafolio_test.go · domain/marketplace_test.go · domain/marketplace_situacion_test.go · usecase/marketplace_test.go |
| [`boundaries/marketplace-referencia-es-solo-procedencia.md`](./boundaries/marketplace-referencia-es-solo-procedencia.md) | Un marketplace de referencia es dato read-only en el DOMINIO, no un `disabled` de la UI (bounded context + anti-corruption layer) | 🌳 enforced | 1.2 | 5 | domain/marketplace_situacion_test.go:TestReferenciaNuncaHabilitaAccion · domain/traer_test.go:TestPlanificarTraerRechazaReferencia · fitness/marketplace_shape_test.go:TestDominioNoAdoptaShapeAjeno |
| [`boundaries/maquinaria-no-contamina-arnes.md`](./boundaries/maquinaria-no-contamina-arnes.md) | La doctrina ①② entra por flags de sesión, jamás escrita en el árbol del arnés (③) | 🌳 enforced | 1.0 | 2 | arch_test.go:TestMaquinariaNoContaminaArnes |
| [`boundaries/doctrina-una-fuente-dos-targets.md`](./boundaries/doctrina-una-fuente-dos-targets.md) | El ruleset ejecutable y `kit/doctrine.md` no deben divergir | 🌱 vivo | 1.0 | 2 | (pendiente — `kit/doctrine.md` es prosa a mano, sin drift-check) |
| [`boundaries/telemetria-de-nacimiento.md`](./boundaries/telemetria-de-nacimiento.md) | Todo arnés nace observable (VISION p9) · S2 en dos modos (instrumentado/degradado) · hook fail-open que proyecta campos | 🌱 vivo | 2.3 | 10 | (pendiente — el módulo `telemetria/` no existe; diseño en `stories/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md`) |
| [`boundaries/ingesta-por-allowlist-declarada.md`](./boundaries/ingesta-por-allowlist-declarada.md) | Lo que entra de afuera se persiste por allowlist declarada, nunca por denylist (ni identidad de cuenta ni contenido ni rutas del usuario) | 🌱 vivo | 1.0 | 7 | (pendiente — nace con el módulo `telemetria/`) |
| [`boundaries/no-aplica-no-es-cero.md`](./boundaries/no-aplica-no-es-cero.md) | «No aplica» y «no lo sé» son valores; `0` y `[]` son afirmaciones | 🌱 vivo | 1.0 | 6 | 3/6 REALES ya verdes: usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio · domain/marketplace_situacion_test.go:TestSituacionNoComparableSinVersionEstante · portafolio/deriva_test.go:TestEvaluarDerivaNoEvaluable |
| [`boundaries/cifra-viaja-con-su-confianza.md`](./boundaries/cifra-viaja-con-su-confianza.md) | Ninguna cifra viaja sola: lleva cómo se atribuyó y, si hay dos fuentes del mismo número, las dos | 🌱 vivo | 1.0 | 6 | (pendiente — nace con el módulo `telemetria/`) |
| [`boundaries/peso-del-binario-es-presupuesto.md`](./boundaries/peso-del-binario-es-presupuesto.md) | El peso del binario es un presupuesto; una dependencia se MIDE antes de adoptarse y un formato ajeno se decodifica con subconjunto propio | 🌱 vivo | 1.0 | 5 | (pendiente — job de CI + tests del decodificador) |
| [`boundaries/archivo-durable-declara-su-esquema.md`](./boundaries/archivo-durable-declara-su-esquema.md) | Un archivo durable declara su esquema y se migra con respaldo; uno derivado se puede tirar, uno durable jamás | 🌱 vivo | 1.0 | 6 | (pendiente — `internal/adapters/store/` no tiene un solo test hoy) |
| [`boundaries/ruta-servida-esta-declarada.md`](./boundaries/ruta-servida-esta-declarada.md) | Una ruta que el daemon sirve está declarada en el contrato, o exenta con razón escrita (ratchet) | 🌳 enforced | 1.1 | 5 | `openapi_contract_test.go` (**4/5 verdes**, validados por fallo inducido; `forma-de-respuesta-unica` difiere honesto por falta de mecanismo) |

Leyenda de estado: ⏳ en forja · 🌱 vivo (nace, se enforça cuando el código llegue) · 🌳 enforced
(código + check corriendo) · 🔍 en-revisión. **Total boundaries: 28 · 148 checks**

> **+2 boundaries · +12 checks el 2026-07-26** (paquete
> [`stories/2026-07-26-conversaciones-del-panel/`](../product/stories/2026-07-26-conversaciones-del-panel/arquitectura.md),
> etapa de arquitectura): nacen `archivo-durable-declara-su-esquema` (6) y
> `ruta-servida-esta-declarada` (5), y `sesion-viva-consistente` sube **v1.1→v1.2** con
> `transicion-de-conversacion-atomica` (+1). **Los dos nodos nuevos nacen `proposed` con TODOS sus
> checks sin enforcer escrito** — el código llega con el paquete; declararlos `enforced` sería el
> pass fabricado (mismo criterio que `indice-desechable-jsonl-es-verdad` v1.3). Los dos salen de
> hallazgos **medidos**, no de intuición: `store.Registry` hace `json.Unmarshal` desnudo sobre
> `[]domain.Session` sin versión ni envelope y `internal/adapters/store` reporta `[no test files]`
> (un cambio de forma pierde el registro entero del operador); y `grep -c cerradas openapi.yaml` →
> **0** con dos endpoints en producción, sin nada en CI que lo cace porque el paso
> `openapi-gen-check` es condicional sobre un directorio generado que no existe (`ci.yml:73`). El
> conteo exacto del drift de contrato — **50 rutas servidas · 39 declaradas · 11 sin declarar · 0
> fantasma** — se generó antes de escribir la fila. `sesion-viva-consistente` **conserva
> `enforced`** (sus 4 checks originales siguen verdes) y su quinto queda declarado pendiente. Un
> tercer candidato se **descartó como nodo propio, con razón escrita**: «el buscador entra al texto
> del transcript» es producto, no arquitectura — vive en un capability (`buscar-en-el-transcript`),
> y generalizado sin sujeto se leería como una licencia para indexar contenido del operador, que es
> justo lo que `ingesta-por-allowlist-declarada` acota.

> **+4 boundaries · +24 checks el 2026-07-26** (paquete
> [`stories/2026-07-24-telemetria-embebida-otel/`](../product/stories/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md),
> etapa de arquitectura del módulo `telemetria/`): nacen `ingesta-por-allowlist-declarada` (7),
> `no-aplica-no-es-cero` (6), `cifra-viaja-con-su-confianza` (6) y
> `peso-del-binario-es-presupuesto` (5). **Los cuatro nacen `proposed`, y tres de ellos con los
> checks difiriendo honesto** — el módulo no existe: declararlos `enforced` sería el pass
> fabricado. La excepción es `no-aplica-no-es-cero`, que **no inventa doctrina sino que consolida
> la que ya estaba resuelta suelta en tres lugares del árbol** y por eso nace con **3 de 6 checks
> con enforcer real y verde** (corridos antes de escribir la fila). Dos candidatos evaluados se
> **descartaron como nodo propio, con razón escrita**: «fail-open de la instrumentación» se
> absorbió en `telemetria-de-nacimiento` v2.3 (su único sujeto es el hook que ese nodo gobierna, y
> generalizado sin contexto se leería como «tragarse los errores»), y «doble costo como oracle» se
> **fusionó** en `cifra-viaja-con-su-confianza` (son las dos mitades de la misma oración y
> comparten superficie de enforcement). En la misma pasada, `telemetria-de-nacimiento` sube
> **v1.0→v2.3** (la fila decía 1.0/2 checks y el archivo iba por 2.2/7: drift de índice, corregido)
> y se corrigen dos filas más que estaban stale respecto de su propio archivo —
> `superficie-local-confinada` 1.2/7 → **1.3/9** e `indice-desechable-jsonl-es-verdad` 1.2 →
> **1.3** — sin tocar el contenido de esos nodos.

**Historia del conteo (pasadas anteriores).**
(+1 boundary · +10 checks el 2026-07-25, paquete `docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`
etapa 3 «diseño técnico»: nace `marketplace-referencia-es-solo-procedencia` **`proposed`** con 5
checks — L1 bounded context + anti-corruption layer, fuentes verificadas en vivo ese día; el 5º
(`traer-referencia-rechazado`) se sumó el mismo día al firmarse `AG-D17`, que habilita
`↧ Traer canónico` completo: con una acción que **materializa en disco** en juego, el invariante
tiene que vivir en el dominio y no en un `disabled` de la UI — un `POST` puede llegar sin pasar
nunca por el botón. Ese es el nodo entero en una línea: el plano Marketplaces necesitaba una regla
que la UI no puede sostener, la clase `referencia` es read-only **en el dominio**, no de render.
`portafolio-identidad-y-deriva-honesta` v1.1→v1.2 suma 4 checks aplicando su mismo L1 al estante
(registro collect-all · catálogo ilegible viaja `null` y no `[]` · versión nunca inventada ·
`no-comparable` de primera clase) y **sigue `enforced`** por sus 4 originales;
`fe-taxonomia-componentes` v1.1→v1.2 suma `no-sibling-widget-imports`, el único cross-import de
capa que quedaba sin gate, **verificado verde antes de agregarlo**. **Los 10 corren**: la etapa 4
(implementación) llegó el MISMO día y cableó los 9 que habían nacido declarados —
`marketplace-referencia-es-solo-procedencia` pasó `proposed → enforced` (v1.2) con sus 5 enforcers
(4 tests colocados + 1 source-scan en `fitness/` para la forma del dominio) y
`portafolio-identidad-y-deriva-honesta` cableó sus 4 nuevos. NUNCA pass fabricado: ninguno se
declaró verde antes de existir.) (+3 boundaries ·
+7 checks en HS-11 2026-07-23: los 3 boundaries que el research de inyección de know-how había
dejado como draft — `maquinaria-no-contamina-arnes` **enforced** de una, `doctrina-una-fuente-dos-targets`
y `telemetria-de-nacimiento` nacen `proposed` honesto, sin código detrás todavía — + 1 check nuevo
en `codigo-traza-a-capability` v1.5, R1 a nivel símbolo) — fundacional HS-04
(backend, 7 boundaries · 29 checks; **+3 en franja-artefactos Fase 1** (2026-07-08): `dead-end` ·
`ruta-a-existe` · `refina-coherente` en `contrato-de-caja-es-fitness-function` v1.3, y `sin-huerfanos`
pasó de promesa a enforcer vivo `domain.VerificarSinHuerfanos` — los 5 verificadores de composición
corren en la ruta `--arnes`; **+1 en HS-14** (2026-07-08): `single-instance-reenfoca` en
`core-no-importa-shell` v1.2 → 33 checks) + HS-05 (frontend, 5 boundaries · 22 checks) + **HS-06 (2 boundaries
· 10 checks + 2 checks nuevos a permisos-gui + 1 en HS-14** (2026-07-08, `healthz-refleja-cors` en
`superficie-local-confinada` v1.2 — REGRESIÓN real detectada en producción por el operador el
mismo día del fix ②, no mejora cosmética; ver su changelog) **= 13 checks)** + **HS-07/HS-08 doctrina v1 (2 boundaries ·
8 checks: `orquestacion-determinista-entre-cajas` + `permisos-derivan-del-rol` — nacieron draft en HS-07,
**enforced en HS-08** con `TestConductorOwnsBoxRouting` + `TestPermissionSetParametrizedByRole`)** + **HS-18
(2026-07-09, doctrina de trazado): 1 boundary · 5 checks — `codigo-traza-a-capability` **enforced**
(`TestCapabilityPointersResolve` + `TestCapabilityCoverage`; R1/R2 pasan sobre el árbol `docs/product/capabilities/`)**. **+
HS-22** (2026-07-14, Portafolio Slice 0 «Cimientos»): 1 boundary · 4 checks —
`portafolio-identidad-y-deriva-honesta` **enforced** (identidad calificada única · deriva nunca
semver · store degrada honesto · procedencia anotada; los 4 con test Go real ya verde). **+
[`conventions/`](./conventions/INDEX.md): 9 convention nodes · 31 checks** (HS-05 + nodo
`versionado.md` 2026-07-15, +1 check `dev-daemon-sincronizado` 2026-07-25 — installer vs.
self-update). **Gran total `arch/`:
123 checks** (92 boundary + 31 convention; nota: el ruleset `--todo` cuenta un total propio a partir de
arch+knowledge — el motor y la suma de docs difieren en un par por deuda menor, cifra viva en `checkpoint.md`).

> **Honestidad (heredada de METODOLOGIA §4 / CADENCE):** la mayoría de los checks están **declarados,
> no corriendo** (`status: proposed`) — se activan cuando cada superficie aterrice. **Excepción HS-06:**
> los 2 boundaries `superficie-local-confinada` + `sesion-viva-consistente` y los 2 checks S2 de
> `permisos-gui` **nacen `enforced`**: su código (daemon Go + shell) y sus fitness tests
> (`fitness/arch_test.go`: `TestNoWildcardCORS`, `TestLocalSurfaceConfined`, `TestOneTurnAtATime`,
> `TestFramesCarryRunID`, `TestNoSilentEventDrop`, `TestMaxTurnsAlways`, `TestSessionSpawnsInArnesPath`,
> `TestArnesPathContainment`) shippean juntos y **PASAN** (`go test ./docs/architecture/fitness/...`). **HS-08 sumó 2
> boundaries `enforced`** (`orquestacion-determinista-entre-cajas` + `permisos-derivan-del-rol`, con
> `TestConductorOwnsBoxRouting` + `TestPermissionSetParametrizedByRole` pasando) + subió
> `contrato-de-caja-es-fitness-function` a enforced (`TestBoxContractValidatesAgainstSchema`). **HS-09
> (Fase D del Mapa) sumó 3 boundaries FE `enforced`** — `fe-taxonomia-componentes` (canvas⊥chrome,
> `pnpm depcruise` verde: 76 módulos/0 violaciones), `fe-tokens-contrato` (stylelint strict-value verde +
> tokens 6→10 regenerados) y `fe-visual-fitness` (`pnpm test` = story-tests verdes en Playwright — el conteo crece con cada story; ≥45 al corte HS-11
> Chromium; migrado `vitest.workspace.ts`→`vitest.config.ts`) — **sin sumar checks** (solo cambio de
> `status`). → **11 boundaries enforced en total** (2 HS-06 + 3 HS-08 + 3 HS-09 + 1 HS-18
> [`codigo-traza-a-capability`: 5 checks incl. R1 símbolo v1.5] + 1 HS-22 [`portafolio-identidad-y-deriva-honesta`]
> + 1 HS-11 2026-07-23 [`maquinaria-no-contamina-arnes`, `TestMaquinariaNoContaminaArnes`]). El resto
> sigue `proposed`; `fe-topologia-fsd` y `fe-transporte-independiente` quedan proposed (el 2º hasta que exista
> `app/realtime/` SSE en Hito 3). El `go-arch-lint check` **corre en CI desde HS-10**
> (`go-arch-lint check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml`, deepScan off —
> el linter de imports es el enforcement). Estado del enforcer por nodo en su frontmatter.

## Subdirectorios

- [`model/`](./model/) — **diagramas as code** (render on-demand, nunca hand-sync). C4 nivel
  contexto (`system-context.mmd`, Mermaid) + container (`container.d2`, D2 = binario Go).
- [`contracts/`](./contracts/) — **schema-first, single source of truth**. `schema/` = las 2
  schemas del dominio (L0 `meta.clase` + `contract:` de caja); `api/openapi.yaml` = superficie
  HTTP/SSE. `gen/` = tipos generados Go+TS (quicktype/oapi-codegen; **aún vacío**: el codegen no
  corrió — `domain.Contract` y los tipos TS son a mano; checks `tipos-generados`/`dtos-generados`
  en warn, deuda registrada).
- [`fitness/`](./fitness/) — **las checks ejecutables** (fallan CI): `.go-arch-lint.yml` (grafo
  de imports Go), `arch_test.go` (lo que el linter no expresa). Los enforcers FE viven junto a la SPA
  (`web/.dependency-cruiser.js`, `web/steiger.config.ts`, `web/.stylelintrc.json`) y los de estilo en
  la raíz (`.golangci.yml`, `web/biome.json`, `lefthook.yml`, `.github/workflows/ci.yml`).
- [`conventions/`](./conventions/INDEX.md) — **convenciones de código as code** (HS-05): 8 nodes · 26
  checks de estilo/lint/format/naming/commits/hooks/CI (Go + TS/React + Rust). Hermano de `fitness/`
  (`fitness/` = forma arquitectónica; `conventions/` = estilo de código). Cada nodo referencia el config
  real vía `enforced_by:`.
- `decisions/` — MADR completos, **solo cuando hace falta** el tratamiento de opciones que una
  ficha no aguanta. Vacío por ahora (el LEDGER cubre el «por qué»).

## El stack (resumen; detalle y fuentes en el research doc y los nodos)

| Capa | Pick |
|---|---|
| Shell v1 | **Tauri 2** (daemon = sidecar `externalBin`); ruta a «app vendible» aditiva |
| Backend | binario Go único `arnesia` (serve/open/index/publish) |
| Conexión CC | subproceso `claude` + **stream-json** por stdin/stdout (patrón conductor) |
| Índice | **map in-memory + store JSON atómico** (reconstruible, desechable) · JSONL = verdad · SQLite modernc (WAL) = **fase 5** (escala, aún no) |
| Transporte | **SSE** multiplexado (`event: map\|dock\|run`) |
| Frontend | Vite+React SPA `go:embed` · Mapa = **HTML+SVG** bandas/carriles (React Flow 12 → Organigrama, HS-09) · **Zustand** + hash-state |
| Dock | **AG-UI** (taxonomía, emisor Go propio) · **assistant-ui** · **CodeMirror 6** + merge |
| Arch as code | JSON Schema 2020-12 → quicktype (Go+TS) · go-arch-lint + depguard · D2+Mermaid · MADR-proyección del LEDGER |
| FE topología (HS-05) | **FSD-lite** (`web/src/{app,pages,widgets,features,entities,shared}`; `pages`=composition-roots por hash-state, sin router) · **dependency-cruiser** (gate) + steiger |
| FE componentes (HS-05) | **shadcn/ui sobre Base UI** (copy-in, lintable) · 6 capas de UI direccionales (**canvas⊥chrome**) · atomic = vocabulario en `shared/ui` |
| Tokens + estilo (HS-05) | **DTCG 2025.10** `.tokens.json` → **Style Dictionary v5** → **Tailwind v4** `@theme`+TS · dark `data-theme` · stylelint anti-magic-value |
| Fitness/convenciones FE (HS-05) | **Storybook 10** (story=test) + a11y axe + regresión visual **local** (sin Chromatic) · **Biome v2.4** + `tsc` strictest · **golangci-lint v2** · **lefthook** |

## Cómo crece

La arquitectura se revisa **al cambiar** (no en cadencia fija como `docs/architecture/knowledge/`): cada decisión
estructural nueva = una ficha `HS-NN` + (si crea/cambia un boundary) un nodo aquí con su check.
Detalle en [`CADENCE.md`](./CADENCE.md).
