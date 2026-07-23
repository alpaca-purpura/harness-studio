# Decisiones — mejorar-arnes-conversando

> Cada decisión conversada se escribe acá EN EL MISMO TURNO (METODOLOGIA §10). Prefijo `MC-D`.

## Sesión PM 2026-07-22 (nacimiento del paquete)

- **MC-D1 · Alcance de «mejorar» = chat libre + doctrina de fondo** 🧑‍⚖️ — sin menú guiado; el kit ②
  ya inyectado es la doctrina. (Registrada también en `INDEX.md`.)
- **MC-D2 · Refresh del Mapa = después de cada turno** 🧑‍⚖️ — costo medido despreciable (~0.4 ms por
  `loader.LoadArnes`).

## Sesión de refinamiento 2026-07-22 (arquitectura verificada contra código, mapa file:line)

- **MC-D3 · El reindex-tras-turno vive en `SessionService.consume` (rama `EventResult`), NO en el
  conductor.** Corrige el §4.2 original del spike: el conductor es adapter de protocolo puro (boundary
  `adaptadores-de-agente-intercambiables`) y no conoce `IndexPort`. Punto exacto:
  `internal/usecase/session_service.go:385-404`, donde ya se guarda `CtxPct` y se appendea `Conv`.
- **MC-D4 · Push al FE por `event: map`, no un evento nuevo.** `internal/adapters/transport/http/broker.go:18`
  ya reserva `EventMap="map"` sin uso (diseño SSE original `map|dock|run`). FE: agregar listener en
  `sse.ts` (hoy solo escucha `dock`).
- **MC-D5 · Fork C FIRMADO 🧑‍⚖️ — rotación de contexto invisible.** Mecanismo completo: umbral sobre
  `ctxPct` **default 40 % configurable** (centro del rango 38-45 pedido) → al `EventResult` que lo
  cruza se cierra el proceso `claude` y se vacía `ClaudeSessionID` → checkpoint **mecánico** (últimos
  N turnos + delta del grafo desde la última rotación + paso pendiente; pasada LLM solo si el mecánico
  demuestra no alcanzar) → spawn **fresco lazy al próximo `Turn`** con el checkpoint vía
  system-prompt-file → `Session.ID` no cambia, `Conv` sin cortes, breadcrumb `RolSys`. Sub-preguntas
  resueltas: el checkpoint vive en **`~/.arnesia/sessions/<id>/`** (lado store — firewall ②↛③; en el
  árbol del arnés lo vería el reindex del Mapa, lo evaluaría conformance y lo arrastraría Publicar) y
  **«rotar a mitad de edición» no existe por construcción** (se rota entre turnos: cerrar al terminar,
  spawnear al empezar; `spawnLocked` ya es lazy y hoy SIEMPRE pasa `Resume=ClaudeSessionID` — rotar es
  vaciar ese campo + inyectar el checkpoint).
- **MC-D6 · Grounding FIRMADO 🧑‍⚖️ — tarjeta de identidad por sesión.** Hoy el kit ② es doctrina
  genérica compartida (`~/.arnesia/doctrine.md`, un archivo para todas las sesiones) y el Claude de la
  sesión solo conoce el arnés por su cwd. Se agrega un system-prompt-file POR SESIÓN: doctrina +
  identidad `(home,id)` + canónico-vs-instalación + estado `deriva`/degradado. Misma infra que el
  checkpoint de rotación (MC-D5) — un mecanismo, dos usos. Es el corazón de «que la conversación tenga
  completo sentido».
- **MC-D7 · Fork B FIRMADO 🧑‍⚖️ = B2, indexer del JSONL nativo.** El operador eligió el camino
  «correcto» (la JSONL de CC es la fuente de verdad completa) sabiendo que es más grande que archivar.
  Alcance mínimo viable para este paquete: (a) al `Close()` se persiste **metadata liviana** de la
  sesión (sin `Conv`): `Session.ID`, `Arnes`, cwd, fechas y la **cadena de `ClaudeSessionID`s** — con
  la rotación MC-D5 una conversación lógica = N JSONLs; sin ese join no hay cosido; (b) un lector que
  camina `~/.claude/projects/<hash-del-cwd>/*.jsonl` del arnés y lista/reconstruye conversaciones. El
  indexer completo fase-5 (SQLite, watcher) sigue diferido. ⚠ Constraint honesto: los turnos `RolSys`
  propios de ArnesIA (breadcrumbs de rotación, chips) NO están en la JSONL nativa — la metadata al
  `Close()` conserva la cadena de rotación; si en la práctica hace falta más, se guarda también un
  `Conv` liviano (decisión de `spec.md`).
- **MC-D8 · Fork A FIRMADO 🧑‍⚖️ = A4 «sesión de reparación»** (dirección del operador + redacción
  confirmada y firmada el mismo día 2026-07-22). Ni A1 (bloquear) ni A2 (banner). Modelo del operador: la instalación ES el banco de
  pruebas («¿cómo pretendes crear un arnés sin probarlo?»); toda instalación fresca pasa por una etapa
  de MAPEO del proyecto («recablear»); un arnés que anda mal = instalación/mapeo malo (se corrige IN
  SITU, como re-instalar y recablear) o base mal diseñada (el fix se LEVANTA al canónico del
  marketplace y LUEGO se prueba sobre la instalación ya hecha) — siempre detectando POR QUÉ falló para
  corregir la base para nuevas instalaciones. Síntesis firmada: sesión sobre instalación = sesión de
  REPARACIÓN legal; la ley anti-drift se PRECISA — lo prohibido es la deriva SILENCIOSA/huérfana
  (enmienda a la letra de INV-1; INV-2 «reconciliación, no prevención» ya contenía el modelo).
  Condiciones: la sesión sabe que edita una instalación (MC-D6) · deriva visible tras cada turno
  (MC-D3 + recalcular deriva de esa entrada) · aprendizaje marcado pendiente-de-backport cuando la
  causa es de base (cola formal = Reparar/Backport S5). Fuera de alcance: la etapa de MAPEO completa de
  instalación fresca (inicializador universal, paquete futuro). Detalle en `spike-spec.md` §3.
