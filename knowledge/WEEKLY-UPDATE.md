# WEEKLY-UPDATE — runbook ejecutable de la cadencia semanal

> **El procedimiento exacto que corre cada semana** para mantener el árbol
> [`knowledge/`](./INDEX.md) al día sin romper la forma en que trabajamos. Las *reglas* del
> árbol viven en [`CADENCE.md`](./CADENCE.md); esto es el *cómo* paso a paso. Lo corre un
> **routine programado** (cron cloud, semanal) o el operador a mano. Tier de autonomía vigente:
> **HÍBRIDO** (ver §Aplicar vs Proponer).

## El prompt (lo que recibe la corrida)

```
ACTUALIZACIÓN SEMANAL — árbol de conocimiento de ArnesIA (knowledge/)

Corres la cadencia que mantiene knowledge/ (el estándar vivo) al día con el ecosistema
Claude Code / arneses. LEE PRIMERO knowledge/CADENCE.md (reglas) y knowledge/WEEKLY-UPDATE.md
(este runbook). Síguelos al pie.

Meta: capturar lo NUEVO o CAMBIADO desde la última corrida — comandos nuevos, campos de
frontmatter nuevos, eventos de hook nuevos, ELEMENTOS nuevos de la superficie, deprecaciones,
mejores prácticas nuevas de Anthropic + expertos — y evolucionar el árbol SIN romper la forma
en que trabajamos.

Reglas duras (innegociables):
- Aditivo sin pérdida: append a L1; lo obsoleto se marca "(deprecado <fecha>)", NUNCA se borra.
- Toda afirmación L1 lleva URL de fuente + fecha. Sin fuente re-chequeable, no entra.
- L2 sigue derivando de L1; si un cambio la vuelve divergente, la MARCAS, no la reescribes.
- Prioridad de fuentes: docs oficiales Anthropic > estándares abiertos > expertos.
- Bump de version + changelog en cada nodo tocado.
- Ante duda de si algo es nuevo/correcto → va a PROPUESTAS, no se aplica.

Ejecuta las 8 fases de este runbook. Tier HÍBRIDO: auto-aplica solo evidencia L1; todo lo que
toca el mapa (L2, checks, elementos nuevos) va al reporte de propuestas. Entrega:
knowledge/_inbox/<AÑO>-W<semana>.md + UN commit.
```

## Las 8 fases

### 0 · Baseline
Leer `CADENCE.md`, `INDEX.md` y, por nodo, su frontmatter `updated` + el último anillo del
`## Changelog`. Fijar la **marca de agua** = fecha de la última corrida (la más reciente entre
los `updated` de los nodos, o el último archivo de `_inbox/`). Todo lo que se busca es *delta
desde esa fecha* — no se re-deriva el árbol entero (eso fue la pasada fundacional).

### 1 · Barrido delta (ligero, proporcional)
Research SOLO de cambios desde la marca de agua. Fuentes en orden:
1. **Changelog oficial de Claude Code** (`code.claude.com/docs` → release notes / `/release-notes`).
2. **claude.com/blog** + **anthropic.com/engineering** (posts nuevos de dirección/agentes/skills).
3. **Estándares abiertos:** agentskills.io, agents.md, modelcontextprotocol.io (cambios de spec).
4. **Expertos reputados** (blogs, GitHub, HN/Reddit) — solo para prácticas y señales de seguridad.

Un frente por familia de elementos (se puede fan-out con subagentes, pero **proporcional al
delta**, no 12-deep siempre). **Pregunta obligatoria extra:** «¿apareció una primitiva/superficie
de dirección NUEVA sin nodo?» (ej. un tipo de componente que no existía). Si sí → propuesta de
**nodo nuevo**.

### 2 · Triage + verificación adversarial
Por hallazgo: (a) ¿a qué nodo pertenece? (b) ¿es NUEVO o ya está en L1? (c) **verificar contra
una fuente oficial** antes de aceptarlo. Ningún claim sin verificar entra a L1 — si solo lo dice
un experto y no hay confirmación oficial, se marca `[experto, no confirmado oficial]` y va a
propuestas, no a L1 auto.

### 3 · Clasificar por riesgo
- **Evidencia (bajo riesgo → auto en tier híbrido):** hecho L1 nuevo con fuente+fecha · marca de
  deprecación en L1 · re-fecha de una fuente ya listada · bump de version + changelog.
- **Doctrina / afecta-mapa (→ SIEMPRE propuesta, nunca auto):** cualquier cambio de **L2** ·
  cualquier check **nuevo/modificado/retirado** (los checks son badges del mapa) · cualquier
  **nodo/elemento nuevo** · cualquier cambio de severidad/señal de un check existente.

### 4 · Aplicar vs Proponer (tier HÍBRIDO)
- **Auto-aplica** (edita el nodo + commitea): SOLO lo de la categoría «Evidencia» de §3.
  Cada append cita fuente+fecha y es git-reversible.
- **Propone** (escribe en el reporte, NO edita el nodo): todo lo «Doctrina/afecta-mapa» de §3.
  Con: qué, por qué, fuente, **diff sugerido**, y —si es check— su `señal en el mapa`.

### 5 · Version + changelog
En cada nodo con evidencia auto-aplicada: bump `version` (minor: +0.1 típico; major si algo se
deprecó fuerte), `updated` a hoy, y anillo nuevo:
`YYYY-MM-DD · vX.Y · qué se añadió/deprecó · fuente`.

### 6 · Propagar (como propuestas salvo que sea evidencia pura)
- Cambio que toca una regla de negocio → nota de propuesta apuntando a `METODOLOGIA.md`.
- Check nuevo que el mapa debería pintar → nota de propuesta en el **backlog de `UX.md`**.
- Estas notas van en el reporte; NO se editan METODOLOGIA/UX en automático (afectan cómo trabajamos).

### 7 · Reporte + commit
Escribir `knowledge/_inbox/<AÑO>-W<ww>.md` (plantilla abajo). **Un solo commit** por corrida,
mensaje claro (`chore(knowledge): update semanal <AÑO>-W<ww> — N evidencias, M propuestas`).
`main` siempre verde. Emitir el resumen en el output de la corrida.

## Plantilla del reporte (`_inbox/<AÑO>-W<ww>.md`)

```markdown
# Update semanal — <AÑO>-W<ww> (<fecha>)

## Resumen
- Evidencia L1 auto-aplicada: N · Propuestas pendientes: M · Deprecaciones: K
- Nodos tocados: <lista> · Elementos nuevos propuestos: <lista o —>
- Marca de agua: <fecha última corrida> → <hoy>

## Aplicado automáticamente (L1 + bump + changelog)
- [<nodo>] <hecho nuevo> — fuente <url> (<fecha>)
- [<nodo>] (deprecado) <lo viejo> — fuente <url>

## Propuestas para revisión (NO aplicado — el operador decide)
### L2 · doctrina
- [<nodo>] <cambio propuesto> · por qué · fuente · diff sugerido
### Checks · afectan el mapa
- [<nodo>] <check nuevo/cambiado> · id · severidad · señal en el mapa · fuente
### Elementos / nodos nuevos
- <elemento> · qué es · por qué merece nodo · fuente
### Propagación
- METODOLOGIA: <nota> · UX backlog: <nota>

## Fuentes revisadas esta corrida
- <url> — <qué se miró> (<fecha>)
```

## Guardarraíles (nunca romper)
- **Nunca borrar** L1/L2/checks → deprecar.
- **Nunca** cambiar severidad/señal de un check en automático (el mapa depende) → propuesta.
- **Nunca** meter un claim sin verificar en L1.
- **Nunca** editar METODOLOGIA/UX/mockup en automático → propuesta.
- Ante duda de si algo es nuevo/correcto → propuestas, no auto-aplica.
- **Un commit por corrida**; si no hubo nada nuevo, se escribe un reporte «sin cambios» y no se
  toca ningún nodo (la ausencia de novedad también es dato).

## Cómo el operador cierra el loop
En una sesión normal: leer el `_inbox/<...>.md` más reciente → aprobar/rechazar propuestas →
el asistente aplica las aprobadas (edita nodos + checks + propaga a METODOLOGIA/UX) → mueve el
reporte a `_inbox/procesados/` o lo marca `revisado: <fecha>`. Así el árbol crece verificado y el
mapa nunca recibe un check que el operador no vio.

## Mecanismo
Routine programado (cron cloud) **semanal**. Corre este runbook contra el repo, commitea a `main`
(solo evidencia + reporte), y notifica el resumen. Cambiar cadencia/tier = editar el routine
(ver `CADENCE.md` §ritual) y este archivo.
