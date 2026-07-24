# Telemetría embebida vía OTel nativo — capa Tokens del Mapa (paquete de trabajo)

> Ficha HS-27 (deuda viva, barrido 2026-07-23/24) · Origen: dos ítems del BACKLOG que resultaron
> ser el mismo trabajo ("telemetría JSONL → indexer real" + "telemetria-de-nacimiento").
> Disciplina METODOLOGIA §10: mockup → decisiones → spec → 🧑‍⚖️ → código → PARIDAD.

## Rumbo firmado (2026-07-24, arquitectura del operador)

**Arquitectura RESUELTA** (ver `decisiones.md` D1-D5): receptor OTLP embebido loopback-only en
el daemon Go · scaffold inyecta env vars de Claude Code (`CLAUDE_CODE_ENABLE_TELEMETRY` +
`OTEL_EXPORTER_OTLP_ENDPOINT`), no un hook custom · JSONL sigue SOLO enumerar/replay
(`conductor-no-parsea-jsonl.md` intacto) · Langfuse 100% opcional, nunca dependencia del
producto. Detalle en `docs/architecture/boundaries/telemetria-de-nacimiento.md` v2.0.

## Por qué es un paquete propio y no se codeó ya

Toca UI **nueva** del Mapa (capa Tokens — hoy ni siquiera el mockup "destino" la diseña,
`mockups/arnesia-mapa-destino.html:295` la deja en "Fase 2" sin dibujar) + un componente Go
nuevo (receptor OTLP). Disciplina de paquete de trabajo: feature nueva = mockup→spec→PARIDAD,
nunca saltar el proceso aunque el diseño de arquitectura ya esté resuelto.

## Flujo y gates

1. **Mockup** — falta por completo. Preguntas de diseño abiertas: ¿granularidad visual (badge
   por nodo con `skill.name` matcheado, vs. panel agregado del inspector)? ¿unidades (tokens
   crudos vs. USD, tabla de precios por modelo)? ¿qué pasa con nodos no-skill (hooks/reglas/
   knowledge) que el atributo `skill.name` no cubre — "sin dato atribuible" honesto, mismo
   patrón que Hallazgos/Contenido hoy?
2. **Decisiones** (`decisiones.md`) — D1-D5 de la arquitectura YA escritas; faltan las de diseño
   visual cuando arranque el mockup.
3. **Spec** (`spec.md`) → 🧑‍⚖️ firma del paquete → implementación autorizada. Cubre: receptor
   OTLP (schema mínimo a decodificar, wire format elegido HTTP/JSON vs protobuf), esquema del
   índice local, contrato del scaffold (env vars), wiring FE de la prop `capa` (hoy
   `map-canvas.tsx` NO la recibe en absoluto).
4. **Implementación** — Go (receptor + índice) + FE (capa Tokens real) + stories/tests.
5. **PARIDAD** → gate final.

## Estado

- [x] investigación real (Explore + claude-code-guide, 2026-07-24): backend/FE actuales, prior
      art legacy (KIT-03/`emit.py`), Langfuse corriendo en la máquina, verificación oficial OTel
      nativo de Claude Code
- [x] arquitectura RESUELTA y documentada (boundary v2.0 + `decisiones.md` D1-D5)
- [ ] mockup de la capa Tokens (granularidad, unidades, nodos sin dato)
- [ ] spec.md
- [ ] implementación
- [ ] PARIDAD

## Retomar aquí

Arquitectura cerrada — lo que falta es 100% diseño visual + implementación, ninguna pregunta de
"¿es esto siquiera viable?" queda abierta. Arrancar releyendo `mockups/INDEX.md` (línea base del
Mapa) antes de proponer el mockup de la capa Tokens, como manda la disciplina.

Fuera de alcance de este paquete (siguen genuinamente bloqueadas, no recortables): capas
**Desempeño** (necesita latencia/reintentos que solo OTel ve vía `hook_*`/`api_retry`, más
diseño de qué es "desempeño" a nivel Mapa) y **Proceso** (mapear eventos a fases del arnés,
diseño no trivial) — quedan en BACKLOG como `bloqueo` genuino.
