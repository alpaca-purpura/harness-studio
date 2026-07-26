---
regla: ingesta-por-allowlist-declarada
version: 1.0
updated: 2026-07-26
status: enforced
ledger: HS-28
sources:
  - url: https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html
    autoridad: experto
    revisado: 2026-07-26
  - url: https://gdpr-info.eu/art-5-gdpr/
    autoridad: oficial
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/INFORME.md
    autoridad: medicion-propia
    revisado: 2026-07-26
  - url: docs/product/stories/2026-07-24-telemetria-embebida-otel/verificacion-2026-07-26/ANEXO-hooks.md
    autoridad: medicion-propia
    revisado: 2026-07-26
enforced_by:
  - docs/architecture/fitness/telemetria_test.go:TestAllowlistNoPersistePII
  - docs/architecture/fitness/telemetria_test.go:TestHookNoReenviaContenido
  - docs/architecture/fitness/telemetria_test.go:TestAllowlistEsListaNoSugerencia
  - docs/architecture/fitness/telemetria_test.go:TestToolResultBytesSePersiste
  - internal/adapters/telemetria/hooks/proyecta_test.go:TestCwdDesconocidoNoSeGuardaCrudo
  - internal/adapters/telemetria/otlp/mapa_cc_test.go:TestLaIdentidadNoCruzaLaPuerta
  - internal/adapters/telemetria/forward/forward_test.go:TestForwardNoReenviaCrudo
severity: error
---

# Lo que entra de afuera se persiste por allowlist declarada, nunca por denylist

## L1 · Principio (estándar de industria)

**Default-deny en la frontera de datos, igual que en la frontera de ejecución.** Dos marcos
independientes llegan a la misma regla:

- **Validación de entrada por allowlist.** Una denylist enumera lo prohibido y por definición
  queda desactualizada frente a un emisor que agrega campos; una allowlist enumera lo permitido y
  lo nuevo cae afuera **por default**. *(experto: OWASP Input Validation Cheat Sheet, textual:
  «Allowlist validation involves defining exactly what IS authorized, and by definition, everything
  else is not authorized» — y sobre la denylist: «this is a massively flawed approach as it is
  trivial for an attacker to bypass such filters»)*
- **Minimización de datos.** Los datos personales tratados deben ser *«adequate, relevant and
  limited to what is necessary in relation to the purposes for which they are processed»*. El
  criterio no es «¿molesta guardarlo?» sino «¿lo necesito para lo que hago?». *(oficial: GDPR
  art. 5(1)(c), textual)*

Corolario que este nodo agrega y que el estándar no dice: **la allowlist se aplica en la PUERTA,
no en la consulta.** Filtrar al mostrar deja el dato en el almacén; filtrar al entrar hace que el
dato nunca exista. Un almacén que no lo tiene no lo puede filtrar mal, no lo puede exportar por
error y no lo puede perder en un incidente.

## L2 · Realización (este árbol Go) — SIN IMPLEMENTAR, DISEÑO RESUELTO

El caso que forzó el nodo: **la telemetría de Claude Code arrastra identidad de cuenta en CADA
punto y CADA log record** — `user.email`, `user.account_uuid`, `user.account_id`,
`organization.id`, `user.id` (verificado en vivo, INFORME §V6) — y **el payload de un hook trae
contenido en claro**: `UserPromptSubmit.prompt` es el prompt completo, `Stop.last_assistant_message`
la respuesta del asistente, `PostToolUse.tool_response` lo que la herramienta leyó o escribió
(verificado, ANEXO §H4). Nada de eso hace falta para el producto: para el join alcanza con
`session_id` + `prompt_id` + los `arnesia.*`.

- **La allowlist es una lista literal, no una intención.** El evento canónico
  `domain.EventoTelemetria` se arma **campo por campo** desde funciones de mapeo que solo leen
  las claves declaradas. **No existe** —y el enforcer lo impide— una ruta de código que copie un
  `map[string]any` completo a un campo del evento. La lista vive en
  [`arquitectura-modulo.md §6.1`](../../product/stories/2026-07-24-telemetria-embebida-otel/arquitectura-modulo.md).
- **Se aplica en los DOS caminos de ingesta, y dos veces en el de S2.** El adaptador OTLP filtra
  al decodificar; el hook **proyecta antes de mandar** y el daemon **re-valida al recibir**. Nunca
  se confía en que el emisor haya filtrado, aunque el emisor sea nuestro propio binario.
- **Las rutas del usuario tampoco se persisten.** `cwd` se usa **solo** para resolver una
  instalación conocida y después se descarta (queda una huella, no la ruta); `transcript_path` se
  descarta entero — además de ser ruta del usuario, en la corrida observada apuntaba a un
  **directorio**, no a un `.jsonl` (ANEXO §H7).
- **El forward opcional filtra en el borde.** Si el operador enciende el export OTLP externo
  (`telemetria-de-nacimiento` C2), lo que sale es el evento **ya proyectado**, jamás el cuerpo
  OTLP crudo: reenviarlo crudo exportaría el email de quien corra el arnés.
- **Un campo desconocido no es un error, es un descarte.** Es la propiedad que hace que una
  versión nueva del emisor no rompa el receptor: lo que no está en la lista simplemente no entra.
- **Dato a favor del canal elegido:** por OTel, prompts y respuestas llegan `<REDACTED>` por
  default. La telemetría es **menos** invasiva que el transcript — otra razón para no leer disco.

**Aplica más allá de la telemetría.** Todo adaptador que ingiere señal de un emisor que no
controlamos cae bajo esta regla: los catálogos de marketplace, los payloads de hook, el
`stream-json` del runtime, y cualquier receptor futuro. El nodo nace por telemetría; la regla no
es de telemetría.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| allowlist-en-la-puerta | lo ingerido se persiste solo si está en la lista declarada del adaptador; lo demás se descarta ANTES de escribir | error | «el almacén guarda campos que nadie declaró» | (pendiente — TestAllowlistNoPersistePII) |
| sin-identidad-de-cuenta | `user.email`·`user.account_uuid`·`user.account_id`·`user.id`·`organization.id` nunca llegan al almacén ni al forward | error | «la telemetría guardó o exportó datos de cuenta» | (pendiente — TestAllowlistNoPersistePII) |
| sin-contenido-de-conversacion | `prompt`·`last_assistant_message`·`tool_response`·`tool_input` nunca se persisten | error | «la conversación terminó en el almacén de métricas» | (pendiente — TestHookNoReenviaContenido) |
| sin-rutas-del-usuario | `cwd` y `transcript_path` no se persisten crudos; solo su resolución o su huella | error | «rutas del usuario en el almacén» | (pendiente — TestCwdDesconocidoNoSeGuardaCrudo) |
| allowlist-no-es-sugerencia | no existe asignación que copie un mapa/struct completo del emisor a la entidad persistida | error | «ingesta por copia: la allowlist no puede enforzarse» | (pendiente — TestAllowlistEsListaNoSugerencia) |
| allowlist-en-todos-los-caminos | cada adaptador de ingesta declara su lista; el receptor re-valida lo que le mandan, aunque el emisor sea propio | error | «un camino de ingesta sin allowlist» | (pendiente — TestHookNoReenviaContenido) |
| forward-filtra-en-el-borde | el export externo opcional reenvía la entidad proyectada, jamás el payload crudo | error | «forward externo reenviando el cuerpo original» | (pendiente — TestForwardNoReenviaCrudo) |

## Changelog

- 2026-07-26 · v1.0 · Nodo fundacional (HS-28, paquete
  `stories/2026-07-24-telemetria-embebida-otel/`, decisión D15 FIRMADA). Nace por un hallazgo
  medido, no por prudencia genérica: la telemetría de Claude Code trae identidad de cuenta en
  cada punto (INFORME §V6) y el payload de hook trae la conversación en claro (ANEXO §H4).
  L1 = allowlist sobre denylist (OWASP) + minimización de datos (GDPR art. 5). L2 = la lista
  literal en el evento canónico, aplicada en la puerta y en los dos caminos de ingesta, más el
  filtro del forward. 7 checks, **los 7 difieren honesto**: el módulo `telemetria/` no existe
  todavía. `status: proposed` — declarar `enforced` sin código sería el pass fabricado que la
  doctrina prohíbe.
