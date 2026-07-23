# spec — Mejorar un arnés conversando (QUÉ + protocolo de ejecución)

> Paquete `2026-07-22-mejorar-arnes-conversando` · base firmada: `decisiones.md` MC-D1..MC-D8 +
> `spike-spec.md` (spike CERRADO, 4 gates del enfoque firmados 🧑‍⚖️). Ejecución ordenada por el
> operador vía `/goal` 2026-07-22: autónoma, ticket a ticket, hasta cerrar el paquete; el gate
> humano final de PARIDAD queda abierto para el operador. RF arrancan en **RF-183** (RF-182 = último
> usado en el repo).

## Protocolo de ejecución por ticket (pedido explícito del operador)

Cada ticket se ejecuta en 4 fases, en orden, sin saltos:

1. **Investigación + diseño técnico.** Releer el código REAL que se va a tocar (no confiar en
   memoria ni en este spec — el repo manda) y la arquitectura as-code que lo gobierna
   (`docs/architecture/boundaries/`, `docs/architecture/knowledge/`). Si la arquitectura as-code está
   stale respecto de lo que el ticket necesita, SE ACTUALIZA en el mismo ticket (nodo/check nuevo o
   versión del existente). El diseño resultante se anota en `aprendizajes.md` ANTES de codear.
2. **Implementación TDD.** Test primero (RED) → código (GREEN) → refactor. Go: test por capa
   (domain → adapter → usecase). FE: vitest/story. Sin test no hay código.
3. **Verificación.** `go test ./...` + `go vet` + (si tocó FE) `pnpm run verify` — y la medida de
   validación REAL del ticket (ver «Validación transversal» abajo). Gaps quedan visibles, jamás
   pass fabricado.
4. **Aprendizajes.** Cerrar la entrada del ticket en `aprendizajes.md`: qué se descubrió del código,
   qué gotchas, qué decisión de diseño tomó forma distinta a lo previsto y por qué. **El ticket
   siguiente ARRANCA leyendo `aprendizajes.md`** — es el mecanismo para que cada ticket gaste menos
   tokens que el anterior (no re-descubrir lo ya aprendido).

Toda duda de producto que no resuelva la base firmada → se le consulta al operador
(AskUserQuestion), no se inventa.

## Validación transversal — el caso Vitalia (real, no fixture)

El arnés `vitalia` registrado en el Portafolio (`~/.arnesia/portafolio.json`, path
`/home/chalreme/Proyectos/luana-vitalia/vitalia`) es el caso de validación E2E de este paquete.
Diagnóstico real (2026-07-22, contra el árbol vivo):

- **Es una instalación sin canónico ni home** (`identidad {id:vitalia, scope:"."}`, `origen: {}`,
  `deriva-no-evaluable "sin version instalada conocida"`). Es un brand overlay nacido in-project en
  el worktree `luana-vitalia` (cuyo harness raíz sí viene de `harness@prenter-marketplace`); NO
  existe copia en ningún marketplace (verificado: prenter-marketplace, marketplace-arneses,
  marketplace-metodo).
- **El Mapa lo ve casi vacío por bug NUESTRO:** el loader reconoce `rule` SOLO como `CLAUDE.md` raíz
  (`loader.go reconocerRegla`); las 4 rules de `.claude/rules/*.md` (el corazón del arnés) son
  invisibles — ni siquiera `no-reconocido`. Grafo actual: 2 nodos / 0 edges.
- **El arnés en sí está mal:** `skills/` solo tiene un README (cero SKILL.md), una rule está
  marcada SUPERSEDED en su propio README (`shell-mockup-per-component.md`), y el sello no declara
  procedencia.

Medida de éxito del paquete: una sesión de chat REAL contra Vitalia (sesión de reparación A4)
mejora ese arnés conversando, el Mapa refleja cada cambio sin recargar, la deriva queda honesta, y
el historial + rotación funcionan sobre esa misma sesión.

## Etapas y tickets (orden de ejecución, dependencias primero)

### E1 — El Mapa se entera (T-L → T2 → T3 → T4)

**T-L · Loader reconoce el directorio de rules** *(nuevo — lo destapó el diagnóstico Vitalia)*
- **RF-183.** `LoadArnesInfo` reconoce `rules/*.md` bajo el dir de elementos (= `.claude/rules/` en
  forma instalada, `rules/` en forma plugin) como nodos `clase: rule`, un nodo por archivo, además
  del `CLAUDE.md` raíz ya reconocido. **Ajuste de investigación (T-L fase 1):** el runtime carga
  TODO `.md` del dir — README incluido — así que el grafo lo refleja igual (fidelidad al runtime >
  suposición de «índice aparte»); lo no-.md es no-reconocido visible (§4.5). El estándar as-code
  (`docs/architecture/knowledge/elements/rules.md` L1.4) YA cubría el dir de rules — el que estaba
  atrás era el loader, no el estándar (sin bump).
- Verificación: `arnesia index` contra Vitalia pasa de 2 nodos a ≥6 (CLAUDE.md + 3 rules + README
  índice + skills/README), test Go con fixture.

**T2 · Reindex-tras-turno (MC-D3)**
- **RF-184.** Al consumir `EventResult` en `SessionService.consume` (NO en el conductor — boundary
  `adaptadores-de-agente-intercambiables`), se corre `loader.LoadArnes(cwd)` de la sesión y se hace
  `IndexPort.Upsert` del grafo nuevo. Si `LoadArnes` falla, el Upsert refleja el estado DEGRADADO
  real (patrón `manifiesto-ausente`), nunca la foto vieja ni silencio.
- **RF-185.** El reindex es best-effort respecto del turno: un fallo de reindex NO rompe el turno ni
  la sesión (se loguea + estado degradado), el chat sigue.
- Verificación: test Go — editar un archivo del arnés entre turnos simulados → `IndexPort.Query`
  devuelve el grafo nuevo; caso degradado (manifiesto roto) → `Degradado: true`.

**T3 · Push al FE por `event: map` (MC-D4)**
- **RF-186.** Tras un Upsert exitoso (o degradado) disparado por turno, el daemon publica por el
  broker SSE el evento YA reservado `event: map` con payload `{harnessId}`.
- **RF-187.** El FE (`sse.ts`) agrega listener de `map`; si el Mapa está montado viendo ese
  `harnessId`, refetchea `GET /api/harnesses/{id}/graph` (patrón existente de
  `workspace-stage.tsx`). Si no está montado, no hace nada (el fetch on-mount ya trae lo último).
- Verificación: test FE del store/listener + `pnpm run verify`.

**T4 · E2E vivo sobre Vitalia (primera pasada de reparación)**
- **RF-188.** Sesión real de chat contra `luana-vitalia/vitalia`: pedirle a Claude un arreglo
  concreto del arnés (p.ej. archivar la rule SUPERSEDED y/o crear una skill real) → confirmar en
  vivo que el Mapa muestra el cambio sin recargar a mano. Documentar la corrida (comandos, evidencia)
  en `PARIDAD.md`.

### E2 — La sesión sabe qué arnés es (T10 → T1)

**T10 · Tarjeta de identidad por sesión (MC-D6)**
- **RF-189.** El provisioner materializa un system-prompt-file POR SESIÓN (p.ej.
  `~/.arnesia/sessions/<id>/system.md`): doctrina ② + tarjeta de identidad del arnés — identidad
  `(home,id)` si el cwd matchea una entrada del Portafolio, canónico-vs-instalación, estado
  `deriva`, degradado si aplica. `Injection.SystemPromptFile` de esa sesión apunta ahí. Cwd fuera
  del Portafolio (carpeta suelta) → tarjeta mínima honesta («arnés sin registrar en el
  Portafolio»).
- **RF-190.** La doctrina ② gana una sección corta de «sesión de reparación» (el loop A4:
  diagnosticar → arreglar in situ → identificar causa instalación-vs-base → backport si es base).
  Se inyecta SOLO no-vacía cuando la tarjeta dice instalación.
- Verificación: test Go del provisioner (tarjeta correcta por caso: canónico / instalación / suelto)
  + inspección manual del archivo materializado de una sesión Vitalia real.

**T1 · Picker rotula + deriva viva (Fork A/A4)**
- **RF-191.** `new-session-picker.tsx` distingue visualmente canónico vs instalación en la
  sub-lista de copias (rotulado, NO bloqueado): elegir una instalación anota la sesión como
  «reparación». El payload de `POST /api/sessions` lleva esa marca (campo nuevo opcional).
- **RF-192.** La sesión en el rail/dock muestra el rótulo «reparación» cuando aplica (chip chico,
  sin banner intrusivo).
- **RF-193.** Tras cada turno cuyo cwd pertenece a una entrada del Portafolio, se re-evalúa la
  `deriva` de esa instalación (mismo mecanismo que ya usa el Portafolio para evaluarla) — el estado
  del Portafolio nunca finge `al-hilo` tras una edición.
- Verificación: test FE del picker (rotulado) + test Go de deriva re-evaluada tras turno.

### E3 — Conversación infinita (T7 → T8 → T9 → T6)

**T7 · `ctxPct` historizado + umbral (MC-D5)**
- **RF-194.** `Session` guarda el histórico de `ctxPct` por turno (slice liviano, no solo el último
  valor). Expuesto en `GET /api/sessions/:id`.
- **RF-195.** Umbral de rotación configurable (default **40 %**); al cruzarlo en `EventResult`, la
  sesión queda marcada `rotacion-pendiente` (estado interno; no rota a mitad de nada).
- Verificación: test Go (histórico crece por turno; umbral marca pendiente; nunca marca en medio de
  un turno streaming).

**T8 · Checkpoint mecánico + rotación (MC-D5)**
- **RF-196.** Con `rotacion-pendiente`, ANTES del próximo spawn: se escribe checkpoint mecánico en
  `~/.arnesia/sessions/<id>/checkpoint.md` (últimos N turnos del `Conv` + delta del grafo desde la
  última rotación + paso pendiente declarado por el último turno) — determinístico, sin LLM.
- **RF-197.** El próximo `Turn` spawnea FRESCO: `ClaudeSessionID` vaciado (sin `--resume`), el
  system-prompt-file de la sesión (T10) incorpora el checkpoint. `Session.ID` no cambia; `Conv`
  sigue creciendo sin cortes.
- **RF-198.** Breadcrumb `RolSys` en `Conv` («— contexto rotado, seguimos —») + la cadena de
  `ClaudeSessionID`s previa queda registrada en la sesión (insumo de T6).
- Verificación: test Go — forzar umbral bajo, simular turnos, confirmar: checkpoint escrito, spawn
  sin resume, `Conv` íntegro, cadena registrada.

**T9 · E2E vivo de rotación**
- **RF-199.** Sesión real (Vitalia u otro arnés) con umbral temporalmente bajo: confirmar que el
  usuario ve UNA conversación continua, el `ClaudeSessionID` rotó por detrás, el checkpoint existe
  y el turno post-rotación retoma el hilo. Evidencia en `PARIDAD.md`.

**T6 · Historial B2 (MC-D7)**
- **RF-200.** `SessionService.Close` deja de borrar sin rastro: persiste metadata liviana de la
  sesión cerrada (ID, `Arnes`, cwd, fechas, cantidad de turnos, cadena de `ClaudeSessionID`s) en
  un registro de cerradas (`~/.arnesia/sesiones-cerradas.json`, patrón JSON atómico existente).
  `Conv` NO se persiste ahí (B2: la JSONL nativa es la verdad; el breadcrumb `RolSys` vive en la
  metadata como parte de la cadena).
- **RF-201.** Lector del corpus nativo: dado el cwd de un arnés, resolver
  `~/.claude/projects/<hash-del-cwd>/` y listar/leer las JSONL de sus `ClaudeSessionID`s conocidos
  (de sesiones vivas + cerradas). Adapter nuevo chico, detrás de un port (respeta
  `indice-desechable-jsonl-es-verdad`: JSONL = verdad, solo lectura).
- **RF-202.** `GET /api/sessions?arnes=<id>&cerradas=1` lista conversaciones del arnés
  (abiertas + cerradas, metadata liviana). `GET /api/sessions/cerradas/{id}/historial` reconstruye
  los turnos user/assistant desde la JSONL nativa (best-effort honesto: si la JSONL ya no está, se
  dice, no se inventa).
- **RF-203.** FE: lista de conversaciones del arnés (desde el picker o el drawer — decidir en
  investigación del ticket con el patrón UI existente), con acceso al historial de una cerrada.
- Verificación: test Go (Close persiste metadata; lector resuelve hash y lee JSONL fixture) + test
  FE + `pnpm run verify`.

### E4 — Cierre (T5)

**T5 · Capabilities + cifras + PARIDAD**
- **RF-204.** Cada superficie nueva/cambiada traza a capability
  (`docs/product/capabilities/…`): nuevas o extendidas para reindex-vivo, tarjeta de identidad,
  sesión de reparación, rotación de contexto, historial. Cero huérfanos (R1/R2/R4 verdes).
- **RF-205.** `scripts/estado.sh` regenera cifras; `arnesia conformance --todo` y `--arnes` sin
  regresión (el warn honesto preexistente `art-es-path` puede seguir).
- **RF-206.** `PARIDAD.md` del paquete: tabla RF→evidencia (test/corrida E2E), desviaciones
  honestas, gate 🧑‍⚖️ ABIERTO para el operador.

## No-goals (fuera de este paquete)

- Reparar/Backport formal (cola, botones) — sigue S5; acá solo se alimenta (deriva visible + marca).
- Etapa de MAPEO de instalación fresca (inicializador universal) — paquete futuro.
- Indexer completo fase-5 (SQLite + watcher + walk total del corpus).
- Conformance automático tras cada turno (extensión futura; MC del spike §4.4).
- Resumen de checkpoint por LLM (solo si el mecánico demuestra no alcanzar — fuera de V1).

## Riesgos conocidos (de la investigación del spike)

- `vitest-browser` no corre headless en background — verificación FE = `pnpm run verify` (gotcha
  HS-20).
- `initialize` handshake obligatorio antes de `can_use_tool` (conductor ya lo hace — no tocar).
- El control-channel de CC está semi-documentado (issue #24594) — cambios al spawn se validan
  contra el binario real, no contra docs.
- `permission/provisioner.go` es spike (rol hardcodeado) — T10 NO debe acoplarse a esa policy.
