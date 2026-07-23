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

## E2/E3/E4 — (se completa al cerrar cada etapa)
