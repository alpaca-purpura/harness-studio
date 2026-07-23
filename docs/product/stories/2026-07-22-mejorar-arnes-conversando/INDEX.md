# Mejorar un arnés conversando — chat + Mapa en vivo (paquete de trabajo)

> `tipo: paquete-de-trabajo` · abierto 2026-07-22 · nace de una duda del operador en sesión de PM:
> *"todo está asociado. Si converso con Claude a través del chat, es para modificar/reparar/mejorar
> el arnés de la sesión en la que me encuentro... conforme voy hablando, tengo que poder ir viendo
> cómo estos cambios se van dando en el Mapa."* Ampliado el mismo día con: historial de conversaciones
> recuperable por arnés + rotación de contexto invisible por umbral de tokens ("debemos seguir
> conversando de forma infinita sobre un arnés... nunca perdamos el contexto"). Etapa: **spike de spec
> CERRADO** (refinado 2026-07-22 contra código file:line): **los 4 gates del enfoque FIRMADOS 🧑‍⚖️
> (Forks A4/B2/C + grounding)** — ver `decisiones.md`. Sigue: `spec.md` con RF numerados.

## Qué resuelve

Hoy el chat (`chat-cc-funcional`, firmado HS-20) YA escribe sobre el árbol real del arnés de la
sesión — eso funciona. Pero está desconectado de todo lo demás:

1. **El Mapa no se entera.** El índice que sirve el Mapa (`IndexPort`) es una foto cacheada; nadie
   re-corre el loader sobre el cwd de la sesión después de un turno. Editás por chat, volvés al Mapa,
   ves la foto vieja.
2. **Hay una contradicción sin resolver entre dos paquetes YA firmados.** La ley anti-drift
   (`2026-07-10-spike-carga-arneses`, INV-1, firmada 2026-07-13) dice que una **instalación** es un
   espejo read-only y "no existe ese camino" para editarla. El picker de sesión que firmamos HOY mismo
   (`2026-07-20-shell-topbar-selector-arnes`, HS-25) deja elegir CUALQUIER copia — incluida una
   instalación — como `installPath` de una sesión de chat con permisos de escritura reales. Nadie lo
   bloquea (`ArnesRegistry.validate` solo chequea path absoluto/no-protegido, nunca canónico-vs-
   instalación). Este paquete tiene que RESOLVER esto, no ignorarlo.
3. **"Mejorar un arnés" no estaba definido.** Ya está — ver decisiones de esta sesión abajo.
4. **El historial de conversación se BORRA al cerrar una sesión.** `SessionService.Close()` dropea el
   registro completo (`Conv` incluido) — hoy "obtener el historial" solo funciona mientras la pestaña
   sigue abierta. El operador lo pidió explícito: "un historial de las conversaciones... que debo poder
   obtener".
5. **No hay rotación de contexto — la conversación con un arnés no puede sentirse infinita todavía.**
   ArnesIA ya calcula el % de contexto usado por turno (`ctxPct`) pero lo tira (solo guarda el último
   valor) y no dispara nada. El operador quiere que, cerca de 38-45% de uso, el sistema rote a un
   proceso `claude` fresco por detrás, sin que el usuario note el corte ni pierda continuidad.

## Prior-art (verificado contra código real esta sesión, no supuesto)

- **`chat-cc-funcional`** (`stories/2026-07-08-chat-cc-funcional/`, firmado HS-20): sesión↔cwd real,
  permisos/scope/gate/Stop. El escenario "modificación de arneses existentes" YA está en su historia —
  no faltaba, lo que faltaba era el resto de las piezas de este paquete.
- **Kit ② ya inyectado a CADA sesión de chat** (`internal/adapters/provision/provisioner.go`,
  `embed_doctrina.go`, `kit/doctrine.md` + skills `auditar-arnes`/`forjar-caja`): enseña anatomía A1-A7,
  `arnes.l0.json`, contrato de caja, clase de 10 primitivas, nomenclatura. **Esto YA es "cómo mejorar un
  arnés ajeno" en forma de doctrina** — no hay que inventarlo, hay que darle feedback en vivo (Mapa +
  conformance). El kit ① (motor conformance embebido) NO viaja como contenido — corre server-side
  (`arnesia conformance` CLI / `GET /api/harnesses/{id}/conformance`); la skill `auditar-arnes` lo invoca
  vía `Bash`, gateado por permiso rol (Ask/Deny).
- **`IndexPort`** (`internal/ports/index.go`) es desechable por diseño ("puede reconstruirse siempre") —
  el mecanismo de reindex-en-vivo que falta NO pelea contra la arquitectura, la completa.
- **`/events`** (`router.go`): ya existe un broker SSE multiplexado que el Dock usa para streamear cada
  turno (`event: run`). Empujar un evento "grafo reindexado" es extender ese canal, no crear uno nuevo.
- **Timing real medido esta sesión** (`loader.LoadArnes` contra `dogfood/dev-full-cycle` y el plugin real
  `harness@0.5.3`): **~0.44ms y ~0.08ms respectivamente** por carga. Reindexar tras CADA turno es barato,
  sin riesgo de performance.
- **Reparar/Backport** (`S1-D11`): siguen sin ninguna lógica de dominio, solo botones disabled+tooltip
  "próximo · S5". Relevante para el fork de instalaciones (abajo).
- **Forja-ciclo-vivo item 3** ("Chat forja/edita", `stories/2026-07-10-forja-ciclo-vivo/`): es un flujo
  GUIADO distinto (`init`/`doctor`/`loop-forward`, cero código, PAUSADO desde F-D6). **Este paquete NO
  lo revive** — es más chico y más inmediato: hacer que el chat libre que YA existe se vea reflejado en
  vivo. Si más adelante se quiere el flujo guiado, es una capa aparte sobre esto, no un reemplazo.
- **`Session.Arnes` ya agrupa sesiones por arnés** ("N sessions may share one", `session.go`) — listar
  el historial de conversaciones de un arnés es un filtro, no un modelo nuevo. El hueco real es que
  `Close()` borra en vez de archivar (§4 arriba).
- **`ctxPct` ya se calcula con datos reales de uso** (`conductor.go`, frame `result.usage`) — la
  rotación por umbral no necesita instrumentar nada nuevo, solo historizarlo + un disparador.
- **Franja Artefactos** (firmado, `stories/2026-07-07-franja-artefactos/`): la casa YA resolvió "pasar
  un digest chico en vez del documento entero" para el hand-off entre cajas (−90% contexto/insumo
  medido). Es el mismo patrón que sirve para el checkpoint de rotación — no hay que inventarlo.

## Decisiones ya tomadas esta sesión (con el operador, 2026-07-22)

1. **Alcance de "mejorar":** **chat libre + doctrina de fondo.** Le pedís lo que quieras (agregar
   función, arreglar bug, refactor); Claude Code edita directo guiado por el kit ② ya inyectado. NO es
   un menú guiado de comandos — eso quedaría para una capa futura sobre forja-ciclo-vivo, fuera de
   alcance acá.
2. **Frecuencia de refresh del Mapa:** **después de cada turno de Claude.** Validado que el costo es
   despreciable (~0.4ms por carga real).
3. **Instalación editable (Fork A): QUEDA ABIERTO A PROPÓSITO** — el operador decidió no resolverlo a
   ciegas en esta conversación; el spike (`spike-spec.md` §3) lo deja planteado con recomendación A1
   (bloquear edición de instalaciones), a firmar antes de pasar a `spec.md`.
4. **Historial que sobrevive el cierre (Fork B): QUEDA ABIERTO**, recomendación B1 (archivar en
   `Close()` + endpoint de listado por arnés, `spike-spec.md` §3b).
5. **Rotación de contexto (Fork C): QUEDA ABIERTO**, mecanismo recomendado en `spike-spec.md` §3c
   (umbral sobre `ctxPct` → checkpoint chico tipo Franja Artefactos → spawn fresco con
   `--append-system-prompt-file` → el `Session.ID` del usuario no cambia, solo el `ClaudeSessionID` de
   abajo) — pero con 4 sub-preguntas todavía sin decidir (umbral exacto, formato del checkpoint, dónde
   vive, riesgo de rotar a mitad de una edición).

## Refinamiento 2026-07-22 (sesión posterior, arquitectura verificada contra código)

Nace `decisiones.md` (MC-D1..MC-D8). Resultado:

- **Fork B FIRMADO 🧑‍⚖️ = B2** (indexer del JSONL nativo, alcance mínimo: metadata al `Close()` con
  cadena de `ClaudeSessionID`s + lector del corpus por arnés).
- **Fork C FIRMADO 🧑‍⚖️** (umbral default 40 % configurable · checkpoint mecánico en
  `~/.arnesia/sessions/<id>/` · rotación lazy entre turnos — las 4 sub-preguntas cerradas, MC-D5).
- **Grounding FIRMADO 🧑‍⚖️ (MC-D6, gap nuevo):** tarjeta de identidad por sesión — hoy el Claude del
  chat no sabe qué arnés edita (solo ve el cwd); se inyecta identidad `(home,id)` +
  canónico-vs-instalación + `deriva` vía system-prompt-file per-session (misma infra que el checkpoint).
- **Correcciones de arquitectura:** reindex va en `SessionService.consume`, no en el conductor (MC-D3,
  boundary `adaptadores-de-agente-intercambiables`) · push por `event: map` ya reservado en el broker
  (MC-D4).
- **Fork A FIRMADO 🧑‍⚖️ = A4 «sesión de reparación»** (reparación in situ + backport al canónico
  según causa — instalación mala vs base mala; enmienda la letra de INV-1 → «ninguna instalación
  deriva en silencio») (MC-D8, `spike-spec.md` §3).

## Retomar aquí (para una sesión nueva sin este contexto)

Este paquete puede empezar de cero: leé `spike-spec.md` + `decisiones.md` (MC-D1..MC-D8) — contienen
TODO el contexto, no asumas que quien los abre vio las conversaciones. Estado: **spike CERRADO, los 4
gates del enfoque FIRMADOS 🧑‍⚖️** (Fork A = A4 sesión de reparación · Fork B = B2 indexer JSONL ·
Fork C rotación 40 % · grounding MC-D6). **Próximo paso: escribir `spec.md` con RF numerados** y de
ahí implementación directa (T1-T10 en `spike-spec.md` §5), sin decisiones de producto pendientes.

## Archivos

- [`spike-spec.md`](./spike-spec.md) — el spike completo: contexto, forks (B/C resueltos, A en
  redacción), plan técnico, plan de tickets T1-T10.
- [`decisiones.md`](./decisiones.md) — MC-D1..MC-D8 (las firmas y la dirección del Fork A).
