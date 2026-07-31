# BACKLOG — ArnesIA (lo-que-viene)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE (4º… 1er eje futuro del árbol). Solo items ABIERTOS. Item
> cerrado se BORRA (su cierre vive en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate`
> (firma humana) · `deuda` · `bloqueo` (depende de otra cosa).
> SSoT del futuro: reemplaza los `Siguiente`/`Deuda` que vivían enterrados en LEDGER.

## Prioridad actual — Mejorar un arnés conversando (spike de spec, 2026-07-22)

- [ ] **Atar chat-cc-funcional + kit ② inyectado + reindex-en-vivo + Mapa + historial + rotación de
  contexto en UNA experiencia continua** — el operador señaló que hoy son piezas separadas: el chat ya
  edita el árbol real del arnés, pero el Mapa sirve una foto cacheada, el historial se borra al cerrar
  una sesión, y no hay rotación de contexto para conversaciones largas. Decidido esta sesión: alcance =
  chat libre + doctrina de fondo · Mapa refresca tras cada turno (medido `loader.LoadArnes` ~0.4ms, sin
  costo). **Faltan 3 forks por firmar** antes de `spec.md` (cada uno con recomendación en
  `spike-spec.md`): **A** instalación editable contradice la ley anti-drift (HS-25 sin guard, recom.
  bloquear) · **B** `Close()` borra el historial en vez de archivarlo (recom. archivar + listado por
  arnés) · **C** rotación de contexto por umbral de `ctxPct` hacia un proceso `claude` fresco, invisible
  para el usuario (mecanismo recomendado, 4 sub-preguntas abiertas).
  → [`stories/2026-07-22-mejorar-arnes-conversando/INDEX.md`](stories/2026-07-22-mejorar-arnes-conversando/INDEX.md) · `gate`
- Nota: NO es lo mismo que el ítem "Fase 1 forja-ciclo-vivo" de más abajo (ese es un flujo GUIADO
  `init`/`doctor`/`loop-forward`, pausado, fuera de alcance de este spike) ni el re-key
  `(home,id,scope)` de la deuda viva (plomería de identidad del Portafolio, sin relación directa).

## Definición canónica de arnés + cadena de proceso (paquete de doctrina, abierto 2026-07-30)

- [ ] **DEF-D2 — debatir y firmar la definición canónica de arnés.** v1 NO ratificada (el operador
  pidió debatir); v2 propuesta: arnés = posee FASES de un proceso, lo opera UN rol, 1 arnés = 1
  plugin; proceso = entidad aparte declarada en el home; vendible = proceso/familia, instalable =
  arnés. DEF-D1 (feature/bugfix/spike = pipelines internos, aplica D20) YA FIRMADA 🧑‍⚖️ 2026-07-30.
  → [`stories/2026-07-30-definicion-de-arnes/debate-definicion.md`](stories/2026-07-30-definicion-de-arnes/debate-definicion.md) · `gate`
- [ ] **DEF-D3 — firmar el diseño de la cadena** (`proceso/<id>.yaml` en el home, fases con
  arnés-owner + `gate_salida` verificable, N:1 para empresa chica). Directiva «diseñar ahora» dada;
  se firma junto con DEF-D2. Después: hoja canónica en `docs/` + tipos-de-paquete del arnés-dev
  (dogfood) · `gate` · `bloqueo`(DEF-D2)

## Rebrand PRENTER — gate FIRMADO (HS-25), deuda residual de diseño

- [x] ✅ **Chrome de tokens (color/tipografía/radios/sombras) reemplazado por el sistema PRENTER**
  (dark-first, negro+teal `#00b7aa`) — construido + gate humano 🧑‍⚖️ FIRMADO 2026-07-22 (HS-25). `color.kind`/
  `health`/`heat` (Mapa) INTACTOS. → [`stories/2026-07-15-rebrand-prenter-design-system/paridad.md`](stories/2026-07-15-rebrand-prenter-design-system/paridad.md)
- [ ] **Re-derivar `mockups/arnesia-shell-A-galaxia.html` (🔒 firmado) a los tokens PRENTER** + re-estampar
  su fila en `mockups/INDEX.md` — es el acto de cierre real que quedó pendiente de la firma de arriba
  (trabajo de diseño, no de paperwork); decidir alcance de botones/CTA por-módulo que siguen en
  `font.sans` sin QA visual (D8) · `deuda`

## Idea en research — Instalador público + licencias org (2026-07-15)

- [ ] **Instalador `curl\|sh` público + activación por key de organización** (org→N-sub-keys) —
  `state: idea`, research de mercado hecho y verificado en vivo (Keygen.sh CE recomendado).
  Alcance acotado con el operador: interno, NO venta externa de ArnesIA (no toca `vision.md`).
  Bloquea `refining`: confirmar en código real que `Groups` de Keygen vive en CE (no EE) +
  decidir módulo destino en el seam (`self-update` extendido vs. `distribucion` nuevo).
  → [`stories/2026-07-15-instalador-publico-licencias-org/INDEX.md`](stories/2026-07-15-instalador-publico-licencias-org/INDEX.md) · `deuda`

## Deuda del dictado por voz (CONSTRUIDO 2026-07-26, falta el gate en vivo)

- [ ] **T6 — dictar con la voz real del operador contra el binario INSTALADO.** `state: gate-humano`.
  Es lo único que bloquea la firma 🧑‍⚖️ de PARIDAD del paquete de voz, y **no lo puede cerrar un test**:
  RF-215 existe porque `getUserMedia` se cuelga mudo sin el puente Rust, y eso *solo* se reproduce en la
  app instalada (el dev server concede por su cuenta). Pasos: `make installer` → instalar → abrir un
  frente → dictar → confirmar que el composer se puebla.
  → [`stories/2026-07-25-spike-voz-dictado/PARIDAD.md`](stories/2026-07-25-spike-voz-dictado/PARIDAD.md)
- [ ] **T7 — qué motor STT se EMPAQUETA en `.deb`/`.AppImage`/`.rpm`.** `state: deuda`. V-D1 se firmó
  como A2 local `base`, pero la sub-decisión `whisper.cpp` vs `faster-whisper` se resolvió como
  **adaptador por PATH** porque no había evidencia para elegir: se midió `faster-whisper` sobre WAV con
  voz sintética, no `whisper.cpp` (pide `cmake`/sudo) ni el `audio/mp4` real de la app. Hasta cerrarlo
  el instalador **no promete STT** y el estado honesto por defecto en una máquina limpia es «Falta el
  motor de transcripción». Cerrar = medir los dos motores sobre el mp4 real.
- [ ] **Contraste a11y de `--warn` en tema claro.** `state: bug`. Destapado de paso al correr las
  stories: `--warn` (`#c96a2e`) sobre `--card` (`#ffffff`) a 10px da **3.76:1**, bajo el mínimo 4.5 de
  axe. Rompe 4 stories de `session-rail/new-session-picker` (`text-warn` de «historial de cerradas:
  Failed to fetch», `new-session-picker.tsx:273`). **No se arregló dentro del paquete de voz a
  propósito** — toca `web/tokens/base.tokens.json` (o el tamaño/peso de ese span) y merece su propio
  paquete. Ver `stories/2026-07-25-spike-voz-dictado/PARIDAD.md` §Hallazgos.
- [ ] **Barrido del resto de los `text-primary` del árbol.** `state: bug`. **La mitad del dock
  YA SE CERRÓ** (T23 de `2026-07-26-conversaciones-del-panel`): `SessionLine` pintaba el
  `◍ <cc-id>` con `text-primary` sobre `bg-secondary` — **2,21:1** medido por axe (`#00b7aa` sobre
  `#eef1f1`), contra el mínimo 4,5 — y al mudarse su cuerpo a `IdentidadDetalle` (RF-328) el texto
  pasó a `--foreground`. Evidencia: las 3 stories del baseline **dejaron de apagar
  `color-contrast`** y el gate a11y corre entero sobre todo el dock. Lo que queda abierto es el
  **resto del árbol**: nadie barrió los otros `text-primary` usados como color de texto, y son de
  la misma familia que la deuda de `--warn` de arriba (un token de acento leído como tipografía).
- [ ] **El anillo de foco no llega a 3:1 en tema claro.** `state: bug`. Medido en
  `2026-07-26-conversaciones-del-panel` (C-13): `--primary` (`#00b7aa`) sobre `--card` da **2,51:1**
  y `--ring` (`#1fc6b8`) **2,14:1**, contra el 3:1 de SC 2.4.11. **No rompe el gate** porque esa
  regla no está en el ruleset por defecto de axe, y por eso mismo no se ve sola. El paquete conservó
  el precedente del repo (`focus:outline-2 focus:outline-primary`) en vez de inventar un anillo
  propio sólo para el dock: dos vocabularios de foco en la misma app son peores que una deuda
  declarada. Mismo origen que las dos de arriba — **valores de token**, y por eso van juntas.

## Identidad de build en Ajustes (CONSTRUIDO + REFINADO 2026-07-26, falta el gate en vivo)

- [ ] **AC-9b — ver la tarjeta en la VENTANA TAURI instalada.** `state: gate-humano`. Único ítem que
  bloquea la firma 🧑‍⚖️ de PARIDAD de RF-231. **AC-9a ya está cerrado:** los 4 estados se vieron en
  la app real (daemon sellado + SPA embebido, capturas en
  [`shots/`](stories/2026-07-26-identidad-de-build/shots/)) — al día · build sin instalar · «cerrá y
  reabrí» · sin sellar. Lo que falta pide **sudo** y por eso es del operador: `make installer` →
  instalar → **`make dev-sync`** (existe `~/.local/bin/arnesia`, sin eso el shell instalado sigue
  con el binario viejo: `architecture/conventions/versionado.md`) → abrir Ajustes en la ventana
  nativa.
  → [`stories/2026-07-26-identidad-de-build/PARIDAD.md`](stories/2026-07-26-identidad-de-build/PARIDAD.md)
- [ ] **El SPA embebido hardcodea `127.0.0.1:4200`.** `state: bug`. Destapado verificando RF-231 en
  vivo: con el daemon en otro puerto, la UI entera queda en `Failed to fetch` (`GET /api/version
  falló — Failed to fetch`) aunque el daemon responda perfecto por `curl` desde ese puerto. Impide
  correr una segunda instancia o mover el puerto sin recompilar el FE. **No se arregló al voleo** —
  es ajeno al paquete de identidad de build y merece decidir si el origen sale de `window.location`
  o de una config inyectada al boot.
- [ ] **`arnesia --version` en la CLI.** `state: deuda`. La identidad del build hoy solo se ve por
  `GET /api/version`/Ajustes. El pedido era Ajustes y ahí está; la CLI queda anotada para no
  parecer una promesa incumplida (B-D6). Barato: `VersionCompleta()` ya existe.
- [ ] **El sello a través del botón «Actualizar», de punta a punta.** `state: deuda`. El self-update
  corre `bundle.sh --daemon-only` (camino ya verificado a mano), pero no se disparó el botón
  completo en esta ronda. Se cierra junto con AC-9, en la misma sentada.

## Versionado + changelog metodológicos (CONSTRUIDO 2026-07-26, falta el ciclo real)

- [ ] **AC-9 — correr un ciclo de publicación real y firmar.** `state: gate-humano`. Todo el
  mecanismo está probado E2E sobre el repo (bump → promoción → bump abortado por changelog vacío →
  `bump-minor` → revertido), falta hacerlo de verdad cuando toque publicar: `python3
  scripts/changelog.py sin-publicar` → `make bump-minor` (esta tanda trae superficie nueva) →
  `make installer`. → [`stories/2026-07-26-versionado-y-changelog-metodologicos/PARIDAD.md`](stories/2026-07-26-versionado-y-changelog-metodologicos/PARIDAD.md)
- [ ] **Volcar al changelog lo construido por los otros paquetes abiertos.** `state: deuda`. Las
  entradas iniciales de `[Sin publicar]` se derivaron de las capabilities nuevas y del paquete de
  RF-231 — lo verificable hoy. Cada paquete que cierre agrega la suya en su propio turno
  (`changelog.py add`), que es justamente la regla nueva.
- [ ] **Tag git automático al publicar.** `state: idea`. Hoy el bump **no toca git** a propósito
  (queda revisable en el working tree). Si se quiere `git tag vX.Y.Z` en el mismo acto, es una
  decisión aparte — no se metió al voleo.

## Gates humanos pendientes (código listo, falta firma 🧑‍⚖️ PARIDAD)

- **Abiertos hoy: 3** — dictado por voz (T6) · identidad de build (AC-9) · versionado+changelog (AC-9). Los 4 históricos (chat-cc-funcional · franja-artefactos · boton-actualizar ·
  inspector-drawer) quedaron FIRMADOS 2026-07-09 (HS-20) — sus 7+7+6+5 desviaciones aceptadas;
  cierre en `ledger/HS-20.md`.

## Outcome ACTIVO — Fase 1 · Ciclo de forja de arneses vivo (2026-07-10)

> Modelo de terreno FIRMADO 🧑‍⚖️ (`stories/2026-07-10-terreno-conocimiento/`, D0-D20 + `arnes.yaml`).
> Ejecución en `stories/2026-07-10-forja-ciclo-vivo/`. Loop meta: `chat → arnes.yaml → gate → scaffold → Mapa`.
> Agnóstico al rubro (dev = un ejemplo). Slice fino primero (Shape-Up).

- [ ] **1. Dogfood scaffold** — `docs/terreno/{proposito,producto,organizacion}/` + `docs/wip/` derivados del
  `arnes.yaml` (INDEX/dim + hojas atómicas D9 + `knowledge/`), migrando `docs/architecture/`→`terreno/producto/`;
  reconciliar 11 dims ↔ 82 caps sin romper R1-R4 (P6) · `gate`
  - [ ] **1a** dimensión piloto `forma-trabajo` end-to-end (golden/fixture) + raíz terreno + esqueleto wip
  - [ ] **1b** migración `docs/architecture/`→`terreno/producto/` + reconciliación P6 (82 caps)
- [ ] **2. Forjador mínimo + gate de completitud** — motor que LEE `arnes.yaml`, corre el gate (D19, extiende
  `arnesia conformance`) y reproduce el golden determinísticamente (rellena PLANTILLAS). Dogfood: forjar arnés-dev · `gate`
- [ ] **3. Chat forja/edita** (corazón) — chat CC in-app: crear arnés por conversación → yaml→gate→scaffold; editar
  vía init/doctor/loop-forward (D8 sello/deriva/cosecha-back) · `gate`
- [ ] **4. Mapa DESTINO** — superset ESTRICTO del baseline renderizando terreno real (salud · WIP estados→done ·
  overlays calidad+economía · sello/deriva · receta en inspector) → portar a Storybook · `gate`
- [ ] **5. Después** — forjar 2° orden (`/po /architect /dev-team /auditor`) + upstream schema `arnes.yaml`+terreno al kit · `deuda`

## Outcome — Programa «Portafolio · ciclo de vida del arnés» (2026-07-13)

> Del SPIKE [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md)
> (S-D5/S-D6, forks resueltos). Portafolio = **front-door del ciclo de vida** (crear·mapear·observar·mejorar,
> VISION). Reencuadrado spike→programa (F5). Agrupa por `empresa`; **`empresa` ⊥ `marketplace`** (ambos YA en
> `arnes.l0.json` + `domain.Arnes`). Procedencia = **lock `.devstudio/arneses.yaml`** (HS-12, detector 3°).
> Auth GitHub = **`gh` si existe / PAT fallback** (user-owned, auth-terms firmado — app = conductor). Update =
> **híbrido** (notify propio semver-vs-tag + re-materializa plugin-native CC). Cierre del spike = firmar el modelo de entidades.
> **Unidad = arnés por identidad `(home, id)`** (S-D10/F-A; home = marketplace autor-declarado): 1 = **canónico** (única
> copia editable) + N **instalaciones** (espejos read-only). Arnés↔empresa↔registry-de-adquisición = **N:M:M** (facetas).
> **Ley anti-drift (descriptiva de ArnesIA):** ArnesIA nunca introduce una copia editable divergente nueva; observa +
> re-materializa (reparar) + backportea; el marketplace `home` = SSoT (pull-antes-de-push). **`deriva`** (ex-«drift») =
> hash vs `home/plugins/<id>/<versión>/`. **Recut post-revisión adversaria (S-D9/S-D10):** el FE no es construible sin cimientos.

- [x] ✅ **0. Cimientos del Portafolio** (dominio/backend, SIN FE) — **CONSTRUIDO Y FIRMADO 🧑‍⚖️** (Sonnet 5,
  T1-T9 · gate firmado 2026-07-13, HS-23): store de portafolio **separado** de `arneses.json` (degrada honesto) ·
  **walker** `.claude/plugins/<id>/` + **fallback `plugin.json`** · campo `version` · `empresas[]` (`registries[]`
  vive en `EntradaPortafolio`, S0-D3) · identidad `(home,id)` + canonicalización · cadena de origen **collect-all**
  (lock≙cc-plugins>git-plugin>manifiesto) · **`deriva`** = hash vs `home/plugins/<id>/<v>/` (o `deriva-no-evaluable`) ·
  detector lock · HTTP+CLI · capabilities+boundary as-code. 27 tests Go nuevos + E2E vivo contra la máquina real
  (cazó y corrigió 2 bugs reales, S0-D14/D15).
  → [`stories/2026-07-13-portafolio-slice0-cimientos/paridad.md`](stories/2026-07-13-portafolio-slice0-cimientos/paridad.md)
  - [x] ✅ **eslabón CC investigado** (Fable 5, 2026-07-13, S0-D1) + **corregido con evidencia real** (Sonnet 5,
    S0-D14): `~/.claude/plugins/installed_plugins.json` (v2) anida las entradas bajo una clave `plugins` —
    `{"version":2,"plugins":{"id@marketplace":[...]}}`, no en el nivel raíz como se había asumido en S0-D1.
  - **Deuda** (registrada en `paridad.md`, no bloquea la firma): ~~re-key del índice in-memory a
    `(home,id,scope)` calificado~~ **CERRADA 2026-07-23** — `IndexPort.Upsert` gana la clave explícita,
    dos arneses del mismo id pelado ya no se pisan (ver `docs/product/capabilities/portafolio/
    observar-en-mapa.yaml`). Sigue abierta: política definitiva de conservación de entradas corruptas
    del store más allá de un ciclo save (hoy: sobrevive si es JSON sintácticamente válido, se pierde
    si el archivo entero está roto — ver `paridad.md` desviación #4).
- [x] ✅ **1. FE Portafolio** — construido (Sonnet 5, T1-T8, 2026-07-14) + gate humano 🧑‍⚖️ FIRMADO
  2026-07-22 (HS-25). 3 superficies (Lista/lente empresa · Wizard-Proyecto/carpeta-local · Drawer READ) +
  Observar (Abrir en Mapa, cierra GAP-1) + Desvincular con confirmación; 34 stories `play()` +
  `selectors.test.ts` · 4 capabilities `fe-portafolio/*` `vivo` (R4) · E2E vivo real.
  → [`stories/2026-07-13-portafolio-slice1-fe/paridad.md`](stories/2026-07-13-portafolio-slice1-fe/paridad.md)
- [ ] **2. Consolidar la agregación al Portafolio (proyecto + marketplace)** — **re-escopeado 2026-07-24**
  (AG-D1, orden del operador): ya no es solo la rama nueva. (a) **marketplace**: git url → **validar
  `marketplace.json`** → listar → elegir → checkout `<checkouts>/<home-slug>/<id>/` (`gh`/PAT) → chat/mejorar;
  (b) **proyecto**: la rama ya construida en Slice 1 entra al alcance — **auditoría visual en vivo 2026-07-24**
  probó que la app respeta el Storybook pero el mockup firmado tiene affordances que nunca bajaron a código
  (tabs descriptivas · pasos simultáneos · contador de hallazgos · badges de diferido visibles) + 1 bug real
  («Marketplace» dos veces en la toolbar, lente y filtro sin distinción) · `gate` · `bloqueo`(1). Mockup aún
  no arrancado; PENDIENTE-01 FIRMADA, PENDIENTE-02 (5 preguntas de flujo) abierta →
  [`stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md`](stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md)
  · [`auditoria-storybook-vs-app.md`](stories/2026-07-23-portafolio-agregar-marketplace/auditoria-storybook-vs-app.md)
- [ ] **3. Publicar** — pull/rebase → conformance-verde → bump semver → changelog obligatorio → push → **tag-tras-push** · `gate` · `bloqueo`(2)
- [ ] **4. Update-check + notify** — versión vs último tag de `home` + changelog + **alias de rename** (no cegar el aviso) · `gate` · `bloqueo`(3)
- [ ] **5. Reparar + Backport** — reparar overwrite SOLO dir privado (superficies compartidas = merge) + bloqueo si lock DevStudio +
  `conformance --arnes`; backport instalación→canónico (alineación de versión) · `gate` · `bloqueo`(1)
- Asociación de negocio (`reporta_a`, arneses que se nutren mutuamente) → materializa en **Galaxia** (Mapa), fuera de alcance inmediato · `deuda`

## Deuda viva (registrada, no bloquea la línea principal)

- [forja/§11] **El daemon quedó al 92 % del techo absoluto de peso (23,04 de 25 MB).**
  `text/template`, linkeado por primera vez por el forjador de la semilla (A-D2), costó
  **+3,28 MB** medidos (diff de binarios limpios b2d2200→bdc1a1b). `PresupuestoBaselineMB`
  se movió con la medición en el commit (camino que el propio test sanciona), pero el
  headroom al techo es ~2 MB: la próxima feature con dependencia pesada lo revienta.
  Cerrarlo = expander propio para los 3 placeholders + `range` de la semilla (el uso real
  no justifica el motor entero) o build del daemon con `-ldflags "-s -w"` midiendo qué
  recupera · `deuda`

- [conversaciones/N-21] 🔴 **La marca de rotación se pierde EN SILENCIO cuando hay dos vistas.**
  Repro (E2E-3, contra el binario instalado): dock abierto → mandar un turno desde OTRO cliente
  (`POST /turn` o una segunda ventana) → colapsar el dock → ocurre una rotación → reabrir ⇒ la
  pantalla muestra **1** marca y el servidor tiene **3**; un `reload` completo muestra las 3. El
  frame de rotación se appendea sólo si la copia local del transcript mide exactamente `TurnoIdx`
  (`internal/usecase/session_rotacion.go:81-84`, guard de idempotencia pensado para el replay por
  `Last-Event-ID`), pero **el turno mandado desde otro cliente nunca entra en la copia local** —sólo
  el que lo envía lo agrega, optimista—, así que la longitud queda corta **para siempre** y toda
  marca posterior se descarta. Es **pérdida silenciosa**, justo lo que el boundary
  `sesion-viva-consistente`/`sin-perdida-silenciosa` prohíbe. Cerrarlo = que el frame de rotación
  no dependa de la longitud local (id de turno estable, o reconciliar contra el servidor al
  re-montar el dock) · evidencia en `stories/2026-07-26-conversaciones-del-panel/verificacion-e2e/INFORME.md`
  · `deuda`
- [conversaciones/N-22] **RF-348 CA-1 no está construido**: no hay marca `⟳ hilo reiniciado ·
  checkpoint` cuando `tryHealResume` respawnea fresh. El literal tiene **0 ocurrencias en el árbol**
  y el heal está documentado como *«silent on success»* (`internal/usecase/session_service.go:755`),
  así que el `cc-id` cambia por debajo sin que el operador se entere. Verificado en vivo (E2E-6).
  Necesita ticket propio · `deuda`
- [conversaciones/N-23] **El re-key de CV-D16 no corre solo ni avisa tras migrar.** `main.go:202`
  pasa `clave = nil` y `migracion.go:224` sólo recalibra `if inf.Migro && clave != nil` — deliberado
  (no atar el arranque a que el Portafolio responda), pero la consecuencia es que **3 de las 5
  sesiones del operador siguen invisibles** hasta que corra a mano
  `arnesia sesiones recalibrar-llaves --aplicar`, y nada se lo dice. Cerrarlo = avisarlo en el log
  del arranque y/o en la UI, no automatizarlo en silencio · `deuda`
- [conversaciones/N-18] **El dry-run de CV-D16 contesta «no hay sesiones» si corre antes del primer
  arranque.** `cmd/arnesia/sesiones.go:131-142` abre sólo `~/.arnesia/sesiones.json` (el v2), que
  todavía no existe, y `NewRegistry` sobre una ruta ausente devuelve un registro vacío: el operador
  que quiere previsualizar **antes** de actualizar lee «tu registro está vacío». Cerrarlo = detectar
  el legado y decir «arrancá el daemon primero», o migrar en memoria para el dry-run · `deuda`
- [conversaciones/N-19] **El `＋` bloqueado no ofrece la salida.** `spec.md:347`/E-07 fijan «esperá a
  que termine el turno **(■ para interrumpir)**»; el código dice «…el turno **en vuelo**»
  (`web/src/widgets/chat-dock/ui/conversacion-row.tsx:19`). Una línea de copy. **Esperando decisión
  del operador**, porque el gate 2 del paquete está *autorizado por directiva, no por lectura* ·
  `deuda`
- [conversaciones/N-20] **El 409 del servidor no distingue `await` de `streaming`**:
  `internal/usecase/session_conversaciones.go:262` devuelve `ErrBusy` para los dos, así que el cuerpo
  habla de un turno en vuelo aunque lo que bloquee sea un permiso (E-08 nombra otro motivo). Impacto
  bajo: la UI ya diferencia bien (`motivoBloqueo`) · `deuda`
- [telemetria/A2] **El rollup horario está construido y NO está en el camino de lectura.** `Resumen`,
  `PorCaja`, `Turnos`, `cobertura` y `escenario` van todas a la tabla cruda; `rollup_hora` solo se
  escribe. Consecuencias hoy: (a) el presupuesto de «tablero en milisegundos» **no está realizado**
  —cada consulta escanea el detalle—; (b) la promesa de «detalle purgado, resumen conservado» era
  falsa y se **retiró** (la purga por plazo ahora se lleva el agregado, para que nada afirme una
  retención que ninguna pantalla puede devolver). Cerrarlo = mover los agregados de dinero y tokens
  a `rollup_hora`, resolviendo antes qué pasa con `sesiones`/`corridas`, que el agregado no modela ·
  `deuda`
- [telemetria/D20] **`puesto` como faceta propia de la instalación del Portafolio** — hoy se deriva
  del `rol` del **sello** (`graph.l0` META, que es del arnés), no del **lugar** donde se usa. Dos
  instalaciones del mismo arnés en dos puestos distintos resuelven al mismo `rol` y solo se
  distinguen por instalación. `domain.EntradaPortafolio` no tiene el campo y este paquete **no** lo
  agrega (D20: no se toca el wire de `GET /api/portafolio`) · `deuda`
- [capabilities/G14] **`domain_modules` de `project.config.yaml` está incompleto** — 10 módulos con
  hojas de capability vivas no figuran en la lista (`http-sse`, `cli-daemon`, `fe-mapa`,
  `fe-portafolio`, `fe-shell`, `handoff`, `tauri`, `usecases`, `dominio-l0`, `indice-persistencia`),
  pese a que el template de capability dice «uno de `project.config.yaml domain_modules`». O el
  template miente o la lista está incompleta. **Drift preexistente**, destapado por
  `capabilities-a-crear.md` §0.1; el paquete de telemetría solo agregó `telemetria` · `deuda`
- [telemetria/J-6 · P2] **TTL de retención sin número firmado** — D15.3 firmó «TTL por default»
  **sin número**; el `90` del mockup es **PROPUESTO**. Implementado como flag
  (`--telemetria-retencion`, default 90) y **rotulado `propuesto: true`** en la config y en
  `arnesia telemetria salud`. **El número lo pone el operador** · `decisión de producto`
- [a11y] **`.text-warn` no cumple contraste mínimo** (axe `color-contrast`, ratio 3.76 vs 4.5
  requerido — `#c96a2e` sobre `#ffffff`): destapado 2026-07-23 corriendo
  `new-session-picker.stories.tsx` (4 stories fallan en la a11y gate cuando
  `ConversacionesDelArnes` renderiza su error de fetch, `.text-warn`). Preexistente, sin
  relación con el re-key — fix de token de color, no tocado en este cierre · `deuda`
- [HS-27/HS-28/telemetria-de-nacimiento] **capa «Mejora» del Mapa (ex «capa Tokens») — telemetría
  embebida** — **arquitectura RESUELTA y VERIFICADA EN VIVO** contra `claude 2.1.220` (2026-07-26,
  3 corridas, USD 0,044). Receptor OTLP embebido loopback-only en el daemon; **canal primario
  `/v1/logs` (`api_request`)**, que trae por request los 4 buckets de tokens + `cost_usd_micros` +
  `duration_ms` + atribución; `OTEL_RESOURCE_ATTRIBUTES` inyecta nuestra llave de terreno y **viaja
  copiada en cada punto**. Decodificador **OTLP/JSON con la stdlib (+0,49 MB)** — `pdata` se midió
  en **+10,79 MB** y se descartó. **El MVP es el JOIN** dinero × proceso (D12.2), no un badge de
  tokens: es el hueco que ccusage/tokscale/Dynatrace no llenan. Alcance **S1 + S2**, cero egreso.
  ⚠️ **Hallazgo abierto: la telemetría arrastra PII** (email + ids de cuenta en cada punto) ⇒ la
  ingesta va por allowlist y el forward opcional filtra en el borde (D15).
  Detalle en [`docs/architecture/boundaries/telemetria-de-nacimiento.md`](../architecture/boundaries/telemetria-de-nacimiento.md)
  v2.2 → [`stories/2026-07-24-telemetria-embebida-otel/INDEX.md`](stories/2026-07-24-telemetria-embebida-otel/INDEX.md).
  Falta: **🧑‍⚖️ firma de D9/D11/D13/D14/D15/D16** → mockup → spec+design → build → PARIDAD · `deuda`
- [HS-09/11] **capas Desempeño/Proceso del Mapa** — **reclasificadas 2026-07-26** (ya no son
  `bloqueo`; la verificación en vivo destapó que la señal existe y D12.2 se llevó una parte):
  · **Proceso entra PARCIALMENTE** por la capa Mejora — el detector P1 («caja que consume y se
  rechaza en el gate») es parte del MVP del join. Lo que queda afuera es el mapeo completo de
  eventos a fases del arnés · `deuda de diseño`
  · **Desempeño** sigue afuera, pero **no por falta de señal**: `api_request.duration_ms` y
  `hook_execution_complete.total_duration_ms` ya llegan hoy (verificado). Falta el **diseño** de
  qué es «desempeño» a nivel Mapa · `deuda de diseño`
- [HS-11] **FE del botón «Correr» de una caja** — el backend YA es async con gate post-run
  (`POST .../boxes/{id}/run` → 202+run_id · `GET .../runs/{runId}` · construido 2026-07-23);
  `inspector.tsx` («Corridas donde actuó») solo tiene prosa «al implementar…», sin botón, sin
  fetch, sin spinner. **Paquete completo abierto y PAUSADO 2026-07-23** (mockup publicado +
  decisiones D1-D4 firmes: mecanismo SSE `event: run` no polling, error 409 inline sin toast) —
  el operador pidió la explicación funcional del botón antes de firmar el mockup, se le dio
  (grounded en vision.md A1-A3 + `run_service.go` T3 real), y eligió pausar sin rechazar ni
  pedir cambios → [`stories/2026-07-23-boton-correr-caja/INDEX.md`](stories/2026-07-23-boton-correr-caja/INDEX.md) · `deuda`
- [doctrina-una-fuente-dos-targets] `kit/doctrine.md` es prosa mantenida a mano, sin drift-check
  contra `docs/architecture/knowledge/`/METODOLOGIA — construir el freshness-check (candidato:
  verificar que las secciones citadas por número siguen existiendo en esa forma) · `deuda`
- [chat] **fase presentación (decisión #5) re-scopeada 2026-07-24** — la nota original
  ("assistant-ui + CodeMirror merge + widgets ricos") agrupaba 6 cosas bajo un rótulo de 2026-07-08;
  HS-26 (2026-07-22) ya resolvió 2 a mano sin librerías (markdown rico, tarjetas de tool-use vía
  `ActivityCard`), y el swap completo a `@assistant-ui/react` queda **descartado** (no cierra
  ninguna brecha ya cerrada a mano, costo real de bundle sin beneficio claro). Quedan 4 sub-ítems
  reales e independientes, cada uno con su investigación ya hecha (file:line reales, tamaño
  estimado, qué falta decidir antes de codear) →
  [`stories/2026-07-24-chat-siguientes-fase-presentacion/INDEX.md`](stories/2026-07-24-chat-siguientes-fase-presentacion/INDEX.md):
  widgets ricos para turnos `sys` · slash-menu del composer · cola de turnos (reemplazar el 409) ·
  diff editable accept/reject por chunk (`@codemirror/merge`, el único que pediría librería nueva) · `deuda`
- [HS-09] **checks `deferred` del ruleset `--todo` — investigado a fondo 2026-07-24, re-scopeado.**
  Cifra real hoy: `266 checks · pass 54 · deferred 212` (era 215 antes del fix de abajo). Desglose
  numérico completo (no al ojo): 134 son catálogo de doctrina sobre primitivas de Claude Code en
  general (`docs/architecture/knowledge/elements/`) — **fuera de alcance real**, ningún target
  implementado los valida hoy (ni `--arnes` ni `--todo`); requeriría una capability nueva entera
  (validar el `.claude/` generado de un arnés contra el catálogo), no un enforcer suelto · 44 son
  `nl-judge` que YA están enforced de verdad en otro job/hook de CI (dependency-cruiser · biome ·
  stylelint · tsc · golangci-lint · lefthook · vitest/storybook) — el motor Go simplemente no
  sabe reportarlos como "enforced-elsewhere" en vez de `deferred`; decisión de nomenclatura del
  motor, no un hueco real, **diferida** · ~24 son deuda de diseño ya admitida en otro lado
  (`gate-honesto`/`agencia-dentro-del-frame`/6 bloqueados SQLite-fase5-OTel-modo-por-fase, ver
  ítems HS-16 abajo) · **9 eran genuinamente accionables — 3 YA ARREGLADOS** (bug del adapter
  `GoArchLint`: probaba `exec.LookPath("go-arch-lint")`, que nunca está en el PATH ni en dev ni en
  CI — ambos invocan `go run pkg@latest` — así que el motor difería localmente algo que CI corre y
  hace cumplir de verdad; corregido para invocar exactamente lo mismo que `ci.yml`, saca 3 falsos
  `deferred` reales) · **5 investigados y descartados como "fix chico"**: `SchemaAdapter`/
  `StaticScan` son stubs que SIEMPRE difieren sin importar el target (nunca delegan a
  `SchemaSet`/`firewallScan` reales) — implementarlos de verdad es posible (`SchemaSet.ValidateJSON`
  ya existe, reusable) pero con un problema real de granularidad: 4 de los 5 checks
  (`perfil-tipo-declarado`/`arquetipo-declarado`/`skill-caja-contract`/`meta-de-enganche-completa`)
  apuntan al MISMO schema completo (`box.contract.schema.json`/`graph.l0.schema.json`), no a un
  campo específico — si el schema falla por CUALQUIER motivo, los 4 mostrarían `fail` aunque el
  campo que a cada uno le importa esté bien. Deuda de diseño propia (resolver granularidad
  campo-específico vs schema-completo), no un enforcer de 30 minutos · `deuda`
  - **Hallazgo colateral (no relacionado a lo de arriba, encontrado investigando):** CI en `main`
    estaba roto desde 2026-07-23 por 2 causas nuevas, además del bug a11y `.text-warn` ya conocido
    — un `gosec` G703 falso positivo (`provisioner.go`, `sessionID` es `newID()` interno, nunca
    input de cliente) y `internal/adapters/history/**` sin registrar en `.go-arch-lint.yml`
    (`cmd` lo importa directo, componente nunca declarado) — **ambos arreglados**, más 10 issues
    más de `golangci-lint` completo (gosec/govet/nilerr/noctx/nolintlint/perfsprint) en archivos
    no relacionados, también arreglados. `golangci-lint run` + `go-arch-lint check` + `go test
    ./... -race` limpios de punta a punta.
- [HS-16] loader reconocedores `deferred`: subagent · plugin-nodo-raíz · edges-de-librería (necesitan diseño) · `deuda`
- [HS-16 Grupo A] **6 checks composición `deferred` — re-verificado 2026-07-24, SIGUEN
  bloqueados tal cual (sin cambios desde HS-16, 2026-07-08).** `index-reconstruible` ·
  `sin-migracion-incremental` · `writer-serializado` (`indice-desechable-jsonl-es-verdad.md`)
  esperan SQLite fase 5 real (hoy: `internal/adapters/index/store.go` sigue
  `// TODO(fase 5)`, CAP-21/22/23 siguen vivo·in-memory/parcial/stub — plan de ataque en
  [`stories/2026-07-24-indice-sqlite-watcher-fase5/INDEX.md`](stories/2026-07-24-indice-sqlite-watcher-fase5/INDEX.md))
  · `modo-por-fase`
  (`permisos-gui-human-in-the-loop.md`) espera permission-mode por fase (hoy:
  `--permission-mode` hardcodeado a `"default"` en `conductor.go`) · `hooks-desde-otel`
  espera el canal OTel de Desempeño — **desbloqueado parcialmente 2026-07-26**: los eventos
  `hook_execution_start`/`_complete` llegan hoy con `total_duration_ms`, `num_blocking` y
  `num_success` (verificado en vivo). Falta el diseño, no la señal.
  Ninguno recortable hoy · `bloqueo`
- [HS-16 Grupo C] `gate-honesto` — **re-verificado 2026-07-24, sin cambios**:
  `domain.VerificarGateHonesto` sigue sin existir (grep confirma), deuda de diseño pura (definir
  la semántica de "gate honesto" vs fantasma), sin dependencia de SQLite/OTel · `deuda`
- [conductor-no-parsea-jsonl] **bugfix candidato — `ctx-derivado-etiquetado` probablemente stale**:
  HS-16 lo agrupó junto a `hooks-desde-otel` como "esperan OTel", pero `ctxPct`
  (`conductor.go:611-639`, ya con test `TestCtxPctUsaUltimoUsage`, commit `b91a061` —
  POSTERIOR a HS-16) calcula `% contexto` desde stream-json puro, sin OTel. Mismo patrón que el
  Grupo B de HS-16 (enforcer real ya vive, solo falta citarlo en el `.md` + confirmar que cubre
  el "etiquetado como derivada", no solo el cálculo) — investigación completa + pasos exactos en
  [`stories/2026-07-24-ctx-derivado-etiquetado/INDEX.md`](stories/2026-07-24-ctx-derivado-etiquetado/INDEX.md) · `deuda`

## Capabilities / doctrina (reorg 2026-07-09, FIRMADO)

- [x] ✅ validador R1/R2 construido + boundary `codigo-traza-a-capability` **enforced** (arch-test real, 0 colgantes / 0 huérfanos / cobertura 100%) + job lefthook `capabilities` — HECHO 2026-07-09
- [x] ✅ **R4 aterrizado (HS-20):** `cap-estado-consistente` + `cap-puntero-estable` = arch-test determinista (`TestCapabilityStatusConsistent`/`TestCapabilityPointersStable`), `--todo` pass 40→42 — HECHO 2026-07-09
- [x] ✅ **`scripts/estado.sh` genera TODO el bloque de cifras** (ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities) — mata la deuda «tecleadas»; cazó la línea stale «24 vivo/46 sin-check/6 STUB» → real `38 vivo · 40 vivo·nc · 1 parcial · 3 stub` — HECHO 2026-07-09 (HS-20)
- [x] ✅ **dead-code decidido (HS-20): NADA es dead real** — `Banda`/`Canal`/`Procedencia` vivos (loader/index) · `Origen` vivo en schema+FE (el tipo Go es espejo del dominio) · `UnidadDeTrabajo` = vocabulario del Spine (motor pendiente) · `toggleTheme` único «sin caller» → todos **seed-futuro**, cero borrado (no se destruye vocabulario de schema firmado)
- [capabilities] **derivación LIVE del estado** (`vivo ⟺ check verde` corriendo cada test, no solo consistencia) → cablear a CI · `deuda`
- [x] ✅ **estado.sh → CI drift-gate + auto-cura (HS-21, eje N2·Paquete A — FIRMADO 2026-07-10):**
  `estado.sh --check` en el job `go` rompe el merge si las cifras del checkpoint quedan stale
  (neutraliza la fecha; `ci-estado-drift`), Y el hook `pre-commit estado-cifras` las **regenera solas**
  en el commit (`--check`→regen→`git add`; `lefthook-precommit-estado`). Cifras stale ya NO se cuelan.
  Cierre en `ledger/HS-21.md`; PARIDAD 10 filas firmada
- [capabilities] validar los ~40 `vivo·nc` (sin-check): construir el test que falta por-cap; parte de FE sin tests (solo stories) — es un paquete propio · `deuda`
- [capabilities] `toggleTheme` seed-futuro: cablear el toggle a la vista Ajustes (RF-100) o cortar · `deuda`
- [capabilities] **índice SQLite real (CAP-21) + watcher fsnotify (CAP-23) + Rebuild real
  (CAP-22)** — sale la fase-5 del stack. Investigado a fondo + plan de ataque 2026-07-24 (diseño
  ya firmado en `indice-desechable-jsonl-es-verdad.md`, solo falta construir) →
  [`stories/2026-07-24-indice-sqlite-watcher-fase5/INDEX.md`](stories/2026-07-24-indice-sqlite-watcher-fase5/INDEX.md) · `bloqueo`

## Homologación de metodología — continuaciones (HS-19 cerrada, estos son los siguientes)

- [homologacion] **upstream del método al kit** `harness@prenter-marketplace` (nueva versión): `/pm` +
  `cap_doctor.py` + scaffolder + capability schema + soporte plugin-mode del seam (desviación #1) · `deuda`
- [homologacion] **forjar los 4 arneses secundarios** `/po` · `/architect` · `/dev-team` · `/auditor` (hoy `/pm` los referencia por auto-chain pero no existen — desviación #4) · `deuda`

## Backlog de UX (extraído de `ux.md` al sanear el eje historia/vigente, 2026-07-10)

- [ux] Vista A/B real (hoy «Evaluar A/B» solo salta al tren) · `deuda`
- [ux] ¿Flujo canónico/ideal por skill como concepto aparte del replay real? · `deuda`
- [ux] Detalle de evals del gate («Ver evals») + telemetría post-deploy por proyecto · `deuda`
- [ux] Historia: mapa por versión (exige snapshot del índice — decidir en fase 3) · `deuda`
- [ux] Onboarding/captura de base a fondo (hoy solo dock guionado) · `deuda`
- [ux] Multi-proyecto: ¿vista por proyecto instalado? · `deuda`
- [ux] Búsqueda global (componentes, corridas, hallazgos) · `deuda`
- [ux] Accesibilidad teclado completa (hoy parcial) · estados vacíos restantes · `deuda`
- [ux] Taxonomía: ¿subtipos de regla / clase L0 visible en el nodo? · `deuda`
- [ux] Leyenda del mapa: falta filtro para el tipo `command` (se renderiza, no se puede filtrar) · `deuda`
- [ux] Organigrama: posición ¿100% libre vs auto-layout+ajuste fino? ¿persistir posiciones como
  metadato? · marketplace por-arnés ¿override o hereda de la empresa? · reporta-a ¿cross-empresa o
  solo intra? · «＋ crear arnés para un puesto» desde el organigrama · `deuda`
- [ux] Badge de conformidad por-nodo sobre el mapa: conectar los checks del árbol de conocimiento a
  Diagnóstico real por arnés (hoy «ver arneses afectados» es demo) · `deuda`
- [ux] ⌘K quick-switch de sesión (palette P4) como añadido al rail, no reemplazo · `deuda`
- [ux] Persistencia de la lista de sesiones: índice SQLite desechable vs sidecar propio (JSONL
  sigue siendo fuente de verdad) · `deuda`
- [ux] Unificar al portar: shell-A-galaxia (it.13) + shell-A-sessions (it.14) + detalle v3 en un
  solo shell · `deuda`
- [ux] Empleados-IA del producto (Valeria·Lisa…) vs roster de dev — ¿dos vistas separadas? · `deuda`
