# Revisión visual/UX — índice SQLite real + watcher fsnotify (fase 5)

> Fase 4 (agente 4, 2026-07-24). Disciplina §10 regla 6 «nada se firma sin verse»: la app de
> verdad, levantada en `127.0.0.1:4200` con el binario compilado desde el working tree (`go build
> ./cmd/arnesia`), navegada con `claude-in-chrome`. Entorno REAL del operador (`~/.arnesia/arneses.json`
> con 2 entradas: `dev-full-cycle` → `dogfood/dev-full-cycle`, `vitalia` →
> `/home/chalreme/Proyectos/luana-vitalia/vitalia`) — no una instalación fresca, no editado.
> Los 5 escenarios de la consigna, todos con evidencia (screenshots en el scratchpad de la sesión +
> respuestas de API citadas literal). **Ningún bug encontrado — nada que arreglar.**

## Escenario 1 — Boot muestra datos REALES, no demo

`GET /api/harnesses` (el endpoint que `listHarnesses` sirve directo desde el índice, RF-207)
devuelve exactamente los 2 arneses reales de `ArnesRegistry`, nada más:

```json
[
  {"id": "dev-full-cycle", "rol": "Ingeniería · Desarrollo full-cycle", "proceso": "desarrollo de software end-to-end (idea → released)", "empresas": ["alpacapurpura"]},
  {"id": "vitalia", "rol": "Ingeniería · Desarrollo full-cycle", "proceso": "desarrollo del producto Vitalia (overlay de brand: HIPAA-lite + shell-feature)", "empresas": ["vitalia"]}
]
```

Cero rastro de `content-studio-full` ni de ningún otro seed hardcodeado — confirma en vivo lo que
`TestIndexRebuildsFromJSONL`/`TestRebuildDesdeArnesRegistry` ya probaban por test.

Visualmente, la sesión ilustrativa de primer uso (`seedSessions()`) abre directo en el Mapa de
`dev-full-cycle` con contenido real: las 4 cajas del ciclo (`escribir el spec` → `/spec-writer`,
`construir contra el spec` → `/builder`, `revisar el build` → `/reviewer`, `promover a released` →
`/releaser`) — coinciden 1:1 con los `SKILL.md` reales bajo `dogfood/dev-full-cycle/skills/`.
Screenshot: `screenshots/1-mapa-dev-full-cycle-boot.jpg`.

`vitalia` no tiene una superficie de FE que liste "todos los arneses indexados" como cards — la
decisión TS-D21 sacó ese picker de la MapBar (ver `web/src/pages/shell/ui/workspace-stage.tsx:132-134`,
comentario explícito: «ya no alimenta ningún picker»); el `Nueva sesión` picker lee del Portafolio
(otro store, workspace-scan), no de `ArnesRegistry`, así que tampoco lo ofrece ahí — comportamiento
esperado, no un gap de este paquete. Se confirmó `vitalia` por API (arriba) — el mismo mecanismo de
`Rebuild()` que expone `dev-full-cycle` en el Mapa expone `vitalia` en `/api/harnesses`, no hay
tratamiento especial de ninguno de los dos.

**Resultado: OK.**

## Escenario 2 — Consola limpia

`read_console_messages` (patrón `.`, sin filtro) en cada punto de control (boot, tras el evento del
watcher, tras kill+restart) — **"No console messages found"** las 3 veces; con `onlyErrors: true`
al final de toda la sesión — **"No console errors or exceptions found"**. Cero warnings nuevos ni
preexistentes que se hicieran visibles.

**Resultado: OK — consola limpia en las 3 pasadas.**

## Escenario 3 — Watcher fsnotify en vivo (RF-210)

Con el daemon corriendo y la pestaña abierta en el Mapa de `dev-full-cycle`, se creó
`dogfood/dev-full-cycle/skills/_visual-check-tmp.md` (archivo suelto, no directorio — revisado
`internal/adapters/loader/loader.go:282-325` `reconocerSkills`: un archivo suelto bajo `skills/`
sin ser un dir con `SKILL.md` se reconoce como nodo `no-reconocido`, visible, nunca descartado en
silencio — §4.5 del loader).

- `GET /api/harnesses/dev-full-cycle/graph` tras el `touch`: aparece el nodo nuevo
  `_visual-check-tmp.md` (`clase: no-reconocido`) junto a los 7 nodos preexistentes — confirma que
  fsnotify emitió `Create`, `ownerOf` resolvió el dueño, `turnReindex` recargó y `Upsert`-eó.
- **La FE se actualizó SOLA, sin recargar la página**: apareció una sección nueva «Knowledge &
  servicios · 1» con la tarjeta `_visual-check-tmp.md (no…)` — el push `event: map` por SSE +
  refetch automático de `MapService` funcionando de punta a punta. Screenshot:
  `screenshots/2-watcher-nodo-nuevo-vivo.jpg`. `read_network_requests` confirma 3 llamados a
  `GET /api/harnesses/dev-full-cycle/graph` disparados solos tras el evento (sin interacción del
  usuario) — consistente con fsnotify emitiendo más de un evento por el mismo `touch`/escritura
  (`Create`+`Write`), cada uno reindexando (comportamiento ya documentado como deuda aceptada en
  `arquitectura.md` §6.2 "debounce no implementado" — no es un bug, es la decisión D1 ya firmada).
- Se borró el archivo temporal. `GET /api/harnesses/dev-full-cycle/graph` volvió a los 7 nodos
  originales, y la FE volvió sola a mostrar el Mapa sin la sección «Knowledge & servicios»
  (screenshot: `screenshots/3-watcher-nodo-removido-vivo.jpg`) — la remoción también viaja por el
  mismo camino (fsnotify `Remove` → reindex → `event: map`).
- `git status --porcelain -- dogfood` devuelve **vacío** — cero rastro del archivo temporal.

**Resultado: OK — watcher confirmado en vivo por partida doble (alta y baja), visible en el grafo
Y en la red, sin recargar la página.**

## Escenario 4 — Reinicio del daemon (RF-207/211)

`kill -9` al proceso (confirmado por `ps aux`/`ss -ltnp`: puerto libre, sin respuesta de
`/healthz`), reinicio con el mismo comando (`go run`-equivalente, el binario compilado del working
tree), `curl /healthz` → `{"status":"ok"}` inmediato. `GET /api/harnesses` tras el reinicio: **los
mismos 2 arneses, idéntico contenido** (`dev-full-cycle` + `vitalia`, mismos rol/proceso/empresas).
Recarga de la pestaña: el Mapa vuelve a mostrar exactamente el mismo grafo de `dev-full-cycle` que
antes del kill (screenshot: `screenshots/4-mapa-tras-restart-daemon.jpg`), la sesión ilustrativa
persistió también (se guarda aparte, en el store de sesiones — no es parte de este paquete pero es
buena señal de que nada se perdió). Consola limpia tras el reinicio (ver Escenario 2).

Nota honesta sobre la red durante la ventana de caída: `read_network_requests` muestra varios
`GET /events` con `503` mientras el daemon estaba abajo (reintentos normales del cliente SSE del
navegador) y un `200` apenas el daemon volvió a responder — comportamiento esperado de una
reconexión, no generó ningún error de consola ni rompió la UI.

**Resultado: OK — cero pérdida, cero contenido demo tras el reinicio.**

## Escenario 5 — Gap de `seedSessions()`/`dev-full-cycle` (desviación 2 de PARIDAD.md)

Como anticipaba la consigna: en ESTE entorno `dev-full-cycle` SÍ está registrado en
`ArnesRegistry`, así que la sesión ilustrativa de primer uso cargó su Mapa sin ningún error 404 en
los 3 boots que se hicieron durante esta revisión (inicial, tras el evento del watcher, tras el
reinicio del daemon) — el gap documentado en `PARIDAD.md` (desviación 2) y `arquitectura.md` §6.3
**no se manifestó acá**, exactamente como se esperaba. No se simuló una instalación fresca
(no se tocó `~/.arnesia/arneses.json`, es estado real del operador). El gap de UX de
primer-instalación sigue siendo un hallazgo documentado, a criterio del operador si amerita ticket
propio — esta revisión no agrega evidencia nueva sobre ese caso porque no es reproducible sin
destruir estado real, tal como la consigna anticipaba.

**Resultado: confirmado no-reproducible en este entorno (esperado); gap sigue documentado, sin
cambios.**

## Bugs encontrados

**Ninguno.** Los 5 escenarios pasaron limpio en el primer intento — no hizo falta ningún fix de
código en esta fase.

## Screenshots (scratchpad de la sesión, no versionados)

1. `1-mapa-dev-full-cycle-boot.jpg` — boot con datos reales, sesión ilustrativa.
2. `2-watcher-nodo-nuevo-vivo.jpg` — nodo `_visual-check-tmp.md (no reconocido)` apareciendo solo
   tras el evento fsnotify, sin recargar.
3. `3-watcher-nodo-removido-vivo.jpg` — el mismo Mapa, solo, tras borrar el archivo temporal.
4. `4-mapa-tras-restart-daemon.jpg` — mismo estado tras `kill -9` + reinicio del daemon.

## Cierre

- Daemon apagado al terminar (`kill` limpio, confirmado por `ps aux`/`ss -ltnp` sin proceso ni
  puerto abierto).
- `git status` del repo: solo los archivos que Fase 1-3 ya habían modificado/creado (ver diff de
  arriba) + este paquete de historias — `dogfood/` (donde se hizo la prueba del watcher) quedó
  100% limpio.
- No se tocó código de producción en esta fase — no hizo falta.
