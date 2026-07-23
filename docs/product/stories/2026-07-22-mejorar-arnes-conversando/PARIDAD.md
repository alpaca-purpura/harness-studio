# PARIDAD — mejorar-arnes-conversando

> Tabla RF → evidencia. Gate 🧑‍⚖️ final ABIERTO para el operador. Corridas E2E ejecutadas por la
> sesión autónoma (goal 2026-07-22); las aprobaciones de permisos del E2E las hizo el aprobador
> delegado vía API (evidencia en logs, decisiones `allow` explícitas por tarjeta).

## E1 — El Mapa se entera

| RF | Evidencia | Estado |
|---|---|---|
| RF-183 loader rules/ | `TestReconocerReglasDir` verde; Vitalia real 2→6 nodos (`arnesia index`) | ✅ |
| RF-184 reindex-tras-turno | `TestTurnReindexer*` + `TestReindexTrasTurno` verdes; **E2E vivo:** índice 0→6 nodos tras turno 1 real, 6→7 tras turno 2 (skill nueva visible sin recarga) | ✅ |
| RF-185 best-effort | Reindexer sin error-return; carga rota → `TestTurnReindexerCargaRota` (grafo vacío Degradado) | ✅ |
| RF-186 event: map | `TestTurnReindexerUpsertaGrafo` (frame `{harness_id}`) + **E2E vivo:** 1 `event: map` por turno en `/events` (capturado con curl -N) | ✅ |
| RF-187 FE refetch | `map-live-store.test.ts` (3/3) + listener `map` en `connectDock` + efecto por `mapRev` en `workspace-stage` (`pnpm run verify` verde) | ✅ (unit; check visual en E2E final) |
| RF-188 E2E Vitalia | **Corrida real 2026-07-22** (daemon aislado :4213, claude real): sesión `s6ade9dfb` de reparación contra `/home/chalreme/Proyectos/luana-vitalia/vitalia`; Claude leyó las rules del arnés y creó `.claude/skills/hipaa-check/SKILL.md` (operacionaliza `hipaa-lite`); tarjeta de permiso Write → `allow` (200) → archivo REAL en disco; grafo del Mapa lo muestra (`hipaa-check:skill`) sin recarga | ✅ |

### Reparación de Vitalia — qué se arregló (validación transversal)

1. **Sello sin rol** → `arnes.l0.json` ahora declara `rol: "Ingeniería · Desarrollo full-cycle"` +
   `proceso` (sin rol, el spawn no materializa flags de permisos y el chat no puede escribir).
2. **Rules invisibles** → RF-183: las 4 rules de `.claude/rules/` ahora son nodos del Mapa.
3. **skills/ vacío** → primera skill real `hipaa-check` creada VÍA CHAT (el circuito completo del
   paquete: conversar → tarjeta → write → Mapa refleja).
4. **Gap de wiring destapado y reparado:** `createSession` con `path` no indexaba el árbol → el
   primer spawn no resolvía rol (sin tarjetas). Fix: `onRegistered` en `createSession`
   (`TestCreateSessionIndexaAlRegistrar`) — turno 1 de la corrida lo sufrió, turno 2 (tras fix +
   restart) fluyó completo.
5. Pendiente (anotado, no bloquea): rule `shell-mockup-per-component.md` sigue SUPERSEDED en el
   árbol (candidata a archivar en una sesión de reparación futura); `deriva` de la instalación
   sigue `no-evaluable` (sin canónico registrado — el overlay no tiene home todavía).

### Hallazgo crítico para Fork C (a reparar en T7)

`ctx_pct` reportó **100** en el primer turno real: `ctxPct` suma el `usage` ACUMULADO del turno
(`input+cache_read+cache_creation` de TODOS los API calls — `cache_read` se re-cuenta por cada
tool-call), no la ocupación real de la ventana. El umbral de rotación (40 %) sobre esa métrica
rotaría de inmediato. T7 corrige la métrica (usage del ÚLTIMO llamado) ANTES de historizar/disparar.

## E2 — La sesión sabe qué arnés es

| RF | Evidencia | Estado |
|---|---|---|
| RF-189 tarjeta por sesión | `TestTarjeta*` + `TestSpawnUsaTarjetaPorSesion` + `TestProvisionSessionEscribeTarjeta`; **E2E vivo:** `~/.arnesia/sessions/s6ade9dfb/system.md` real con doctrina+tarjeta | ✅ |
| RF-190 loop A4 en la tarjeta | Solo con copia=instalación (`TestTarjetaInstalacionEsReparacion`); **E2E vivo:** Claude cerró su respuesta con «Causa diagnosticada: instalación (la skill se creó IN SITU…)» — el loop A4 operando SIN pedirlo en el turno | ✅ |
| RF-191 picker rotula + payload | Hint «→ reparación» + `ArnesElegido.reparacion` → `Session.Reparacion` (verify verde) | ✅ (visual al gate) |
| RF-192 chip en la card | `SessionCard` pinta chip `reparación` con tooltip A4 | ✅ (visual al gate) |
| RF-193 deriva re-evaluada | `TestReevaluarDerivaTrasEdicion`; encadenada al reindexer en main | ✅ |

## E3 — Conversación infinita (rotación)

| RF | Evidencia | Estado |
|---|---|---|
| RF-194 ctxPct real + histórico | `TestCtxPctUsaUltimoUsage`; **E2E vivo:** turno con tool-calls dio **14 %** donde el código viejo daba 100 % espurio; `ctx_hist=[14,68]` | ✅ |
| RF-195 umbral configurable | Flag `-rotacion-umbral` (default 40, 0 apaga); `TestCtxHistYUmbralRotacion`; **E2E vivo** con umbral 1 % | ✅ |
| RF-196 checkpoint mecánico | `CheckpointMecanico` (últimos 6 turnos + instrucción, sin LLM); en `Session.Checkpoint` + `system.md` de la sesión (grep «Checkpoint de rotación» = 1) | ✅ |
| RF-197 spawn fresco invisible | `TestRotacionInvisible` (spawn 2 SIN resume); **E2E vivo:** `claude_session_id` fd4632f8→3c67bca3, `Session.ID` intacto, y el proceso FRESCO respondió «la rule hipaa-lite respalda esa skill» — continuidad semántica pura por checkpoint | ✅ |
| RF-198 breadcrumb + cadena | Conv real: `[user,assistant,user,assistant,user,assistant,sys,user,assistant]` + `cadena_cc=[fd4632f8]` | ✅ |
| RF-199 E2E rotación | Corrida completa 2026-07-22 (arriba) — el usuario ve UNA conversación | ✅ |

**Desviación honesta (RF-196):** el checkpoint mecánico V1 lleva últimos-N-turnos + instrucción;
el «delta del grafo desde la última rotación» del spike NO viaja aún (el Conv no registra
tool-calls y el grafo vivo ya está en el Mapa/árbol) — anotado como extensión si la práctica lo
pide. El checkpoint materializa DENTRO de `system.md` (mismo dir `~/.arnesia/sessions/<id>/` que
manda MC-D5), no como archivo aparte.

## E3b — Historial B2

| RF | Evidencia | Estado |
|---|---|---|
| RF-200 Close archiva metadata | `TestCloseArchivaMetadata`; **E2E vivo:** DELETE de la sesión real → `sesiones-cerradas.json` con `cadena_cc=[fd4632f8, 3c67bca3]`, cwd, fecha, 9 turnos, sin Conv | ✅ |
| RF-201 lector JSONL nativo | `TestDirParaCwd` (encoding verificado contra el CLI real) + `TestTurnosLeeCorpus` | ✅ |
| RF-202 endpoints | `GET /api/sessions?arnes=&cerradas=1` + `/sessions/cerradas/{id}/historial`; **E2E vivo:** conversación cerrada reconstruida COMPLETA (9 turnos) cosiendo las 2 JSONLs de la cadena de rotación, 0 faltantes | ✅ |
| RF-203 FE conversaciones | `conversaciones-store` (3 tests unit) + sección en el picker (vivas+cerradas+expand) | ✅ (visual al gate) |

Fix colateral: sesiones archivadas sin `Cwd` (pre-T6) → backfill vía resolver en `archivarLocked`.

## E4 — Cierre

| RF | Evidencia | Estado |
|---|---|---|
| RF-204 capabilities | 5 nuevas: CAP-94 reindex-tras-turno · CAP-95 fe-mapa/reindex-en-vivo · CAP-96 tarjeta-identidad-por-sesion · CAP-97 rotacion-de-contexto · CAP-98 historial-de-conversaciones; cobertura **100 %** (0 huérfanos) | ✅ |
| RF-205 cifras sin regresión | `estado.sh` regenerado: `--todo` 257 checks · 48 pass · 0 fail; dogfood `--arnes` 20/21 (warn honesto preexistente); capabilities 93→98 | ✅ |
| RF-206 PARIDAD | este documento — **gate 🧑‍⚖️ ABIERTO para el operador** | ⏳ |

## Estado global

**T-L, T1, T2, T3, T4, T6, T7, T8, T9, T10 implementados, testeados y con E2E vivo. T5 cerrado
salvo la firma.** Pendiente ÚNICAMENTE del operador:

1. **Gate 🧑‍⚖️ de PARIDAD** — verificación visual en vivo (chips del picker/rail, Mapa
   refrescando en pantalla, sección Conversaciones) — todo el circuito de datos ya está probado
   E2E a nivel API; lo visual se firma viéndolo.
2. Deuda honesta anotada: story-tests visuales de los chips nuevos (vitest-browser no corre
   headless en bg) · umbral 40 % puede quedar corto con harnesses pesados (ctx inicial ~68 % en
   el worktree luana-vitalia — si molesta, subir umbral o excluir `cache_creation`) · delta de
   grafo no viaja en el checkpoint V1 · Vitalia: rule SUPERSEDED sigue en el árbol (siguiente
   sesión de reparación) y el overlay no tiene home/canónico en ningún marketplace (candidato a
   Publicar cuando exista).
