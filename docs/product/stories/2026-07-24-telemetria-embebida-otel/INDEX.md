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
- [x] **2ª tanda de verificación en vivo (hooks + bloque `env`)** →
      [`verificacion-2026-07-26/ANEXO-hooks.md`](verificacion-2026-07-26/ANEXO-hooks.md): 9 hallazgos.
      Los dos que cambian el diseño: la llave del join es **`(session_id, prompt_id)`** y **la premisa
      de S2 era falsa** — el bloque `env` de un `settings.json` enciende OTel ⇒ S2 se parte en
      `s2-instrumentado` y `s2-degradado`
- [x] **arquitectura del módulo** → [`arquitectura-modulo.md`](arquitectura-modulo.md): mapa de
      componentes · firmas Go de puertos/dominio · DDL completo · decodificador OTLP/JSON ·
      allowlist campo por campo · contrato del spawn · S2 en dos modos · catálogo · motor de
      detectores · API HTTP · presupuestos no funcionales. **21 decisiones de arquitectura nuevas
      (A1..A21) y 12 contradicciones registradas.** 3 preguntas quedan `ABIERTO` para el operador
- [x] **matriz de escenarios** → [`escenarios.md`](escenarios.md): 10 familias, ~70 escenarios, con
      qué hace el sistema · qué ve el usuario · cómo se verifica
- [x] **doctrina persistida** → 4 boundaries nuevos en `docs/architecture/boundaries/`
      (`ingesta-por-allowlist-declarada` · `no-aplica-no-es-cero` · `cifra-viaja-con-su-confianza` ·
      `peso-del-binario-es-presupuesto`) + `telemetria-de-nacimiento` **v2.3** + `INDEX.md` de
      arquitectura al día (22 → **26 boundaries · 136 checks**)
- [x] **capabilities planificadas** → [`capabilities-a-crear.md`](capabilities-a-crear.md):
      22 hojas (CAP-118…CAP-139), con puntero y check. **No se crean todavía**: R1 exige que el
      símbolo exista
- [x] `spec.md` + `design.md` escritos (RF-232…RF-286 · UI al pixel) → **falta 🧑‍⚖️ firma del par**
- [x] **plan de stories** → [`plan-storybook.md`](plan-storybook.md): 125 stories en 17 archivos
- [x] **plan de desarrollo** → [`plan-desarrollo.md`](plan-desarrollo.md): **39 tickets en 4 tramos**,
      con archivos, firmas, RF, capabilities, criterios y comando de demostración por ticket.
      Resuelve los 5 bloqueantes (cross-import de entities · `knowledge` no es clase · `puesto` no
      existe en el Portafolio · los 2 contrastes que rompían el gate a11y · los punteros de CAP-139)
      y deja 4 paradas obligatorias para el operador (P0 firma del mockup · P1 **A20** · P2 TTL ·
      P4 desviación del slot)
- [x] **cadena E2E probada antes de codear** →
      [`verificacion-2026-07-26/CADENA-E2E.md`](verificacion-2026-07-26/CADENA-E2E.md) +
      `evidencia/baseline-mapa-antes.png` (el «antes» del gate)
- [ ] 🧑‍⚖️ firma del mockup ← **bloquea el Tramo B entero**
- [ ] implementación
- [ ] PARIDAD

## Decisiones abiertas

| # | qué | ¿bloquea el MVP? |
|---|---|---|
| **D9.9** | ¿parser propio o shell-out a `ccusage`? | no — posterior al MVP |
| **D15** | privacidad/retención: TTL, borrado, filtrado en el forward | **sí** — antes de persistir nada |
| **G4** | cert de firma de código (USD 150-300/año + HSM) | no para Linux; **sí** para Windows/macOS |
| — | ~~**D8**~~ | **cerrada por inexistencia del dilema** (V2) |

## Retomar aquí (2026-07-26, EN CONSTRUCCIÓN — Tramo 0 + Tramo A)

**El constructor está ejecutando [`plan-desarrollo.md`](plan-desarrollo.md) T1→T27** (Tramo 0 +
Tramo A: los RF sin superficie). El **Tramo B (T28-T37) NO se construye**: su parada **P0** exige
la firma 🧑‍⚖️ del mockup, que sigue en iteración 1.

Las tres paradas, al día de hoy:

- **P0 · mockup sin firmar** ⇒ Tramo B bloqueado. El Tramo A entrega igual: `arnesia telemetria
  resumen|mejoras|salud|purgar|catalogo` es superficie observable sin FE (patrón `arnesia portafolio`).
- **P1 · A20 (dónde vive el bloque `env`)** ⇒ **ABIERTA, del operador**. Escrita en
  [`decisiones.md`](decisiones.md) §🛑. Se construyó todo lo que no depende de ella.
- **P2 · el TTL** ⇒ flag con default 90 **rotulado PROPUESTO**; el número lo pone el operador.
- **P3 · `OTEL_LOGS_EXPORTER`** ⇒ **CERRADA** en vivo (ANEXO H10.4): es obligatoria.

<details><summary>Retomar aquí anterior (tras el plan de desarrollo)</summary>

## Retomar aquí (2026-07-26, tras el plan de desarrollo)

**El plan de construcción está escrito y es ejecutable sin volver a decidir nada** →
[`plan-desarrollo.md`](plan-desarrollo.md). El constructor arranca por **T1** (firmar D18-D22 en
`decisiones.md` y registrar la deuda) y sigue el orden del §1. El **Tramo A** (T4-T27, los RF sin
superficie) **ya está autorizado** por el gate del bloque D9·D11·D13·D14·D15·D16 y no espera al
mockup; el **Tramo B** (T28-T37, los píxeles) **sí lo espera**.

Lo que el operador tiene que responder, y hasta dónde llega el trabajo sin eso:

1. **🧑‍⚖️ firma del mockup** (parada P0) — sin ella el Tramo A entrega igual: `arnesia telemetria
   resumen|mejoras|salud` es superficie observable sin FE, mismo patrón que `arnesia portafolio`.
2. **A20 — dónde vive el bloque `env`** (parada P1, §7.5 de `arquitectura-modulo.md`): opción A
   (repo del arnés, cobertura parcial, cero fricción) u opción B (proyecto del usuario, cobertura
   completa, escribe archivos de un tercero ⇒ consentimiento explícito). ⚡ H10.1 cerró la tercera
   vía: un plugin **no** puede aportar el bloque. Sin la respuesta, todo corre en `s2-degradado`.
3. **El número del TTL de retención** (parada P2, J-6): el `90` del mockup no está firmado.
4. **Confirmar la desviación del slot `Tokens` → `Mejora`** (parada P4) antes de la primera story.

</details>

<details><summary>Retomar aquí anterior (tras la etapa de arquitectura)</summary>

**La arquitectura del módulo está resuelta a detalle** →
[`arquitectura-modulo.md`](arquitectura-modulo.md) + [`escenarios.md`](escenarios.md) +
[`capabilities-a-crear.md`](capabilities-a-crear.md). La doctrina reusable ya está persistida como
boundaries. **Lo que falta para poder construir son tres cosas, en este orden:**

1. **Decidir A20 — dónde vive el bloque `env`** (opción A: repo del propio arnés · opción B:
   proyecto del usuario, con consentimiento explícito). Es del operador porque B escribe archivos
   de un tercero. Detalle y recomendación en `arquitectura-modulo.md` §7.5. ⚡ **Ya no hay
   verificación que la evite:** H10.1 probó que un **plugin no puede aportar el bloque `env`**
   (0 payloads vs 2 del control positivo) ⇒ no existe auto-instrumentación al instalar.
2. ~~Cerrar tres verificaciones~~ ✅ **HECHAS** (ANEXO §H10, con control positivo en cada corrida):
   plugin `env` **no** · `${VAR}` **no se expande** en el bloque `env` (sí en los comandos de hook)
   · comando de hook inexistente = **fail-open confirmado por el runtime**. Las dos primeras
   cambiaron el diseño → **A22** (el token de ingesta no viaja en el bloque `env`: `/v1/*` acepta
   sin token bajo Host loopback) y A20 confirmada necesaria.
3. **Iterar y firmar el mockup**, que ahora tiene que dibujar un panel más: el estado 3 se parte en
   `s2-instrumentado` (con dinero) y `s2-degradado` (sin dinero, con motivo).

Después: `spec.md` + `design.md` → firma del par → implementación.

</details>

## Contexto previo (pre-arquitectura)

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
