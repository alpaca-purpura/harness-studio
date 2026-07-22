# Mejorar un arnés conversando — chat + Mapa en vivo (paquete de trabajo)

> `tipo: paquete-de-trabajo` · abierto 2026-07-22 · nace de una duda del operador en sesión de PM:
> *"todo está asociado. Si converso con Claude a través del chat, es para modificar/reparar/mejorar
> el arnés de la sesión en la que me encuentro... conforme voy hablando, tengo que poder ir viendo
> cómo estos cambios se van dando en el Mapa."* Etapa: **spike de spec** (resolver el fork abierto →
> firma del enfoque → recién ahí código). El gate 🧑‍⚖️ del enfoque está ABIERTO.

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

## Decisiones ya tomadas esta sesión (con el operador, 2026-07-22)

1. **Alcance de "mejorar":** **chat libre + doctrina de fondo.** Le pedís lo que quieras (agregar
   función, arreglar bug, refactor); Claude Code edita directo guiado por el kit ② ya inyectado. NO es
   un menú guiado de comandos — eso quedaría para una capa futura sobre forja-ciclo-vivo, fuera de
   alcance acá.
2. **Frecuencia de refresh del Mapa:** **después de cada turno de Claude.** Validado que el costo es
   despreciable (~0.4ms por carga real).
3. **Instalación editable (el fork del punto 2 de arriba): QUEDA ABIERTO A PROPÓSITO** — el operador
   decidió no resolverlo a ciegas en esta conversación; el spike (`spike-spec.md` Fork A) lo deja
   planteado con una recomendación, a firmar antes de pasar a `spec.md`.

## Retomar aquí (para una sesión nueva sin este contexto)

Este paquete puede empezar de cero: leé `spike-spec.md` completo (contiene TODO el contexto — no
asumas que quien lo abre vio esta conversación). Lo único que falta antes de escribir `spec.md` es
**firmar el Fork A** (instalación editable durante el chat, ver `spike-spec.md` §Fork A) con el
operador. Una vez firmado ese fork, el resto (reindex-tras-turno + push por `/events` + el Mapa
consume) es implementación directa, sin decisiones de producto pendientes.

## Archivos

- [`spike-spec.md`](./spike-spec.md) — el spike completo: contexto, el fork abierto con recomendación,
  plan técnico, plan de tickets.
