# Capa «Mejora» del Mapa — telemetría embebida (paquete de trabajo)

> Ficha HS-27/HS-28 · Origen: dos ítems del BACKLOG que resultaron ser el mismo trabajo
> ("telemetría JSONL → indexer real" + "telemetria-de-nacimiento").
> Disciplina METODOLOGIA §10: mockup → decisiones → spec+design → 🧑‍⚖️ → código → PARIDAD.
>
> **El paquete se abrió como «capa Tokens» y ya no se llama así** (D12.3 lo pidió, D16.3 lo
> resolvió): el entregable junta **dinero + proceso** y propone un fix — «Tokens» nombraba el
> insumo, no el producto. El slug de la carpeta se conserva por estabilidad de links.

## Qué se entrega (D12.2 — el MVP es el JOIN)

> *«este arnés, en este puesto, quema $X — y el 60 % se va en la caja Y, que falla el gate 3 de
> cada 4 veces»*

No es tokens solos ni proceso solo: es la **correlación**. Es el hueco de mercado confirmado (D9.8):
ccusage, tokscale, Dynatrace y Azure agregan por herramienta, modelo, proyecto y día —
**ninguno por unidad de trabajo**. El eje **arnés × empresa × puesto** es terreno libre.

**Alcance (D12.1): S1 + S2, máquina propia. S3 fuera.** Todo en `127.0.0.1`.

## Arquitectura — resuelta y VERIFICADA EN VIVO

Receptor OTLP embebido loopback-only en el daemon Go · el scaffold/spawn inyecta env vars de
telemetría · el JSONL **no** se toca · Langfuse 100 % opcional, nunca dependencia del producto.
Detalle en [`arquitectura-telemetria.md`](arquitectura-telemetria.md) (8 capas L0→L7) y en el
boundary [`telemetria-de-nacimiento.md`](../../../architecture/boundaries/telemetria-de-nacimiento.md) v2.2.

**Lo que la verificación del 26/07 corrigió** ([`verificacion-2026-07-26/INFORME.md`](verificacion-2026-07-26/INFORME.md)):

| antes | después | evidencia |
|---|---|---|
| canal = `/v1/metrics` (D4.1) | **canal primario = `/v1/logs` (`api_request`)** | V1 |
| D8: leer el JSONL **o** resignar B1/B12 | **el dilema no existía** — el split 5m/1h viene en el `result` del stream-json | V2 |
| `pdata`, +1,7 MB (F1) | **+10,79 MB medido** → OTLP/JSON + stdlib, **+0,49 MB** | V5 |
| forzar `http/protobuf` (F4) | **forzar `http/json`** | V5 |
| atribución por-skill amenazada (D10) | confirmada la amenaza, **rescatada por `plugin_id_hash`** hasta nivel de arnés | V3 |
| — | **nuevo:** PII en cada punto · temporalidad Delta · `intValue` off-spec | V6, V5.1, V5 |

## Estado

- [x] investigación real (2026-07-24): backend/FE actuales, prior art legacy (KIT-03/`emit.py`),
      Langfuse corriendo en la máquina, verificación oficial de OTel nativo
- [x] arquitectura resuelta y documentada (boundary + `decisiones.md` D1-D5)
- [x] investigación multi-runtime → [`investigacion-runtimes.md`](investigacion-runtimes.md):
      **13 runtimes** en dos tandas (Claude Code · Gemini · Qwen · Codex · Amp · Cursor · Copilot
      CLI · OpenCode · Cline · Goose · Aider · Crush · Factory Droid) + estado 2026 de la semconv
      GenAI (**todo en `Development`**). El canal universal es el **stream-json por turno**, no OTel.
      *Genuinamente sin verificar:* Factory Droid y la lista de §«NO verificado» del propio informe
- [x] investigación plataformas OSS + licencias → [`investigacion-plataformas.md`](investigacion-plataformas.md)
- [x] investigación stack embebible → [`investigacion-stack-embebible.md`](investigacion-stack-embebible.md)
      — ⚠️ su recomendación de `pdata` quedó **refutada por medición propia** (V5)
- [x] arquitectura consolidada → [`arquitectura-telemetria.md`](arquitectura-telemetria.md)
- [x] **verificación EN VIVO contra `claude 2.1.220`** (2026-07-26) →
      [`verificacion-2026-07-26/`](verificacion-2026-07-26/INFORME.md): 7 hallazgos, evidencia cruda
      versionada (PII redactada), USD 0,044 de costo
- [x] refinamiento pre-mockup: detectores del MVP · evento canónico · renombre (D16)
- [ ] **🧑‍⚖️ firma del bloque D9 + D11 + D13 + D14 + D15 + D16** ← *acá estamos*
- [~] **propuesta de mockup escrita** → [`propuesta-mockup.md`](propuesta-mockup.md) — 2 superficies
      (capa Mejora del Mapa + tarjeta del Portafolio), tarjeta de punto de mejora con contrafactual,
      4ª tab del inspector, **7 estados honestos**. 3 decisiones pendientes del operador antes de dibujar
- [~] **mockup dibujado** → [`mockup-capa-mejora.html`](mockup-capa-mejora.html) · publicado en
      https://claude.ai/code/artifact/c161bc8b-c612-4a0a-aaac-ffd6ff65b5e2 — 8 secciones, 2 desviaciones
      declaradas (renombre del slot · tokens PRENTER vs. baseline ámbar). **Iteración 1: falta iterar y firmar 🧑‍⚖️**
- [ ] `spec.md` + `design.md` → 🧑‍⚖️ firma del par
- [ ] implementación
- [ ] PARIDAD

## Decisiones abiertas

| # | qué | ¿bloquea el MVP? |
|---|---|---|
| **D9.9** | ¿parser propio o shell-out a `ccusage`? | no — posterior al MVP |
| **D15** | privacidad/retención: TTL, borrado, filtrado en el forward | **sí** — antes de persistir nada |
| **G4** | cert de firma de código (USD 150-300/año + HSM) | no para Linux; **sí** para Windows/macOS |
| — | ~~**D8**~~ | **cerrada por inexistencia del dilema** (V2) |

## Retomar aquí

**Investigación CERRADA y VERIFICADA EN VIVO.** El refinamiento pre-mockup está hecho: la lista
corta de detectores (D16.1), los campos del evento canónico (D16.2) y el renombre (D16.3).

**Próximo paso: 🧑‍⚖️ sobre D9/D11/D13/D14/D15/D16, y después el mockup.** D14 requiere firma
explícita porque **corrige a D4, que ya estaba firmada**.

Arrancar el mockup releyendo [`mockups/INDEX.md`](../../../../mockups/INDEX.md) como manda la
disciplina — la línea base es el Storybook, no un `.html` suelto. **Las dos preguntas de diseño que
bloqueaban el mockup ya tienen respuesta empírica:**

1. **Granularidad** — resuelta: `arnés × caja × sesión × turno` (V3/D14.3). Por-skill no se puede y
   no hace falta; cada número carga su `atribucion_confianza` (D16.2).
2. **Llave canónica** — resuelta: la del terreno propio, inyectada por `OTEL_RESOURCE_ATTRIBUTES`,
   que **viaja copiada en cada punto** (V4). No depende del vocabulario de ningún runtime.

## Fuera de alcance de este paquete

- **Capa Desempeño** (latencia/reintentos a nivel Mapa) — sigue afuera **pero ya no por falta de
  señal**: `api_request.duration_ms` y `hook_execution_complete.total_duration_ms` llegan hoy
  (V1). Lo que falta es el **diseño** de qué es «desempeño» a nivel Mapa. Reclasificar en el
  BACKLOG: era `bloqueo`, es `deuda de diseño`.
- **Capa Proceso completa** — **entra PARCIALMENTE por D12.2**: el join necesita la señal de
  proceso (detector P1: caja que consume y se rechaza en el gate). Lo que queda afuera es el mapeo
  completo de eventos a fases del arnés.
- **S3** (arnés en máquina de un cliente sin ArnesIA) — descartado por D12.1: choca con
  `superficie-local-confinada.md` y abre consentimiento/GDPR. Paquete propio si se reabre.
