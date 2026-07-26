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
- [~] investigación multi-runtime (2026-07-26, PARCIAL) → [`investigacion-runtimes.md`](investigacion-runtimes.md):
      6 runtimes verificados (Claude Code · Gemini · Qwen · Codex · Amp · Cursor). Destapó que el
      canal universal es el **stream-json por turno**, no OTel (5/6 vs 3/6), y 3 restricciones
      duras del esquema normalizado (aritmética de tokens no uniforme entre proveedores · una regla
      de acumulación distinta por runtime · nadie adopta `gen_ai.*` puro). Faltan OpenCode/Cline/
      Goose/Aider/Crush + estado 2026 de semconv GenAI + licencias de plataformas OSS
- [x] investigación plataformas OSS + licencias (2026-07-26) →
      [`investigacion-plataformas.md`](investigacion-plataformas.md): **contiene una CORRECCIÓN a
      `telemetria-de-nacimiento.md` v2.0** (la redacción «third-party» de `skill.name`/`plugin.name`
      NO se evita con `OTEL_LOG_TOOL_DETAILS` — amenaza la atribución por-componente). Veredicto: no
      existe backend lite adoptable; lo vendorizable es dato+esquema (catálogo LiteLLM MIT ·
      `ModelUsage` de Helicone Apache-2.0 · semconv Go de OpenInference) — Phoenix ELv2 descartado,
      Lunary borrado del mapa
- [x] investigación stack embebible (2026-07-26, **con mediciones propias**) →
      [`investigacion-stack-embebible.md`](investigacion-stack-embebible.md): stack recomendado =
      handler propio + `collector/pdata` (**+1,7 MB**, API v1.x estable) + `modernc.org/sqlite` ya
      presente + rollup horario (**190× más rápido**). Ratifica el `sin-cgo` ya enforced. Destapa
      2 riesgos operativos altos: `tauri#11992` (notarización macOS falla **con `externalBin`**, que
      ya usamos) y Azure Artifact Signing geo-restringido (probablemente no aplica a LATAM)
- [ ] mockup de la capa Tokens (granularidad, unidades, nodos sin dato)
- [ ] spec.md
- [ ] implementación
- [ ] PARIDAD

- [x] **arquitectura consolidada** (2026-07-26) →
      [`arquitectura-telemetria.md`](arquitectura-telemetria.md): las ~50 recomendaciones agrupadas
      A-H + los 13 detectores · la arquitectura en 8 capas (L0 contrato → L7 superficie) · **el
      mecanismo de obligación** (se logra en el SELLO vía `arnesia conformance`, no en runtime) ·
      alcance S1+S2 y MVP=join cerrados por D12

## Retomar aquí

**Investigación SOTA CERRADA (2026-07-26, 3 carriles)** → D9 (arquitectura consolidada), D10
(corrección del boundary a v2.1), D11 (empaquetado A-o-B-nunca-C) en `decisiones.md`.
**D9 está SIN FIRMAR** — es recomendación de la investigación, no decisión. Primer paso al retomar:
🧑‍⚖️ sobre D9, y resolver las dos abiertas (**D8** ¿leer 4 campos del JSONL? · **D9.9** ¿parser propio
o shell-out a `ccusage`?).

Después, el mockup. Arrancar releyendo `mockups/INDEX.md` (línea base del Mapa) como manda la
disciplina. **Dos preguntas de diseño que la investigación agregó y bloquean el mockup:**
1. **Granularidad de atribución** — `skill.name` se redacta a `"third-party"` para nuestros arneses
   (D10); `OTEL_RESOURCE_ATTRIBUTES` mitiga pero solo **por proceso**. ¿El número vive por nodo del
   Mapa (⇒ hace falta un proceso por unidad) o por caja/arnés?
2. **Llave canónica** — no puede ser `skill.name` (concepto exclusivo de Claude Code, D7.2). Debe
   ser del terreno propio (arnés × caja/paso × sesión) o el mockup no sobrevive al segundo runtime.

Fuera de alcance de este paquete (siguen genuinamente bloqueadas, no recortables): capas
**Desempeño** (necesita latencia/reintentos que solo OTel ve vía `hook_*`/`api_retry`, más
diseño de qué es "desempeño" a nivel Mapa) y **Proceso** (mapear eventos a fases del arnés,
diseño no trivial) — quedan en BACKLOG como `bloqueo` genuino.
