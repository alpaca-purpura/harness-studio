# Arquitectura del módulo `telemetria/` — plan de construcción

> **Qué es esto:** el plan arquitectónico a detalle de la capa «Mejora». Está escrito para que
> otro agente construya **sin volver a decidir nada**. Todo lo que acá aparece como firma Go, DDL o
> nombre de archivo es la forma que se espera; lo que quede sin decidir está marcado
> **`ABIERTO`** con su dueño.
>
> **Insumos vinculantes:** `decisiones.md` D1..D17 + el bloque 🧑‍⚖️ FIRMADO 2026-07-26 ·
> [`verificacion-2026-07-26/INFORME.md`](verificacion-2026-07-26/INFORME.md) (V1..V7) ·
> [`verificacion-2026-07-26/ANEXO-hooks.md`](verificacion-2026-07-26/ANEXO-hooks.md) (H1..H7) ·
> boundary [`telemetria-de-nacimiento.md`](../../../architecture/boundaries/telemetria-de-nacimiento.md) v2.3.
>
> **Estado: PROPUESTA de arquitectura. CERO código.** El par `spec.md`+`design.md` y su firma 🧑‍⚖️
> siguen siendo requisito antes de construir (METODOLOGIA §10).

---

## 0 · Las 18 decisiones de arquitectura que toma este documento

Ninguna estaba en `decisiones.md`. Cada una tiene su justificación en la sección que se cita.

| id | decisión | §                        |
|---|---|---|
| **A1** | Un solo listener y un solo puerto: `/v1/logs` y `/v1/metrics` se montan en el mux de `:4200`, no en un `:4318` propio | [§3.1](#31-un-solo-puerto) |
| **A2** | **DB separada** `~/.arnesia/telemetria.db`; jamás dentro de `index.db` | [§5.1](#51-por-qué-una-db-aparte) |
| **A3** | Migración **aditiva versionada**; una ruptura **archiva** el archivo, nunca lo borra | [§5.4](#54-versionado-y-migración) |
| **A4** | El campo se llama **`emisor`**, no `origen` — D16.2 pisaba vocabulario L0 | [§2.2](#22-colisiones-de-vocabulario-l0-resueltas) |
| **A5** | **No existe `adapters/telemetria/streamjson/`**: el `result` lo sigue decodificando `agent/claudecode`, que es su dueño | [§4.2](#42-el-canal-stream-json-no-tiene-adaptador-propio) |
| **A6** | **Token de ingesta separado** del token de la API: filtrarlo concede «escribir telemetría», no «conducir el agente» | [§7.3](#73-el-token-de-ingesta-es-otro-token) |
| **A7** | El hook de S2 **es el propio binario** `arnesia hook proceso`; cero runtime nuevo | [§7.1](#71-el-hook-es-el-propio-binario) |
| **A8** | Ficha de descubrimiento en `os.UserConfigDir()/arnesia/daemon.json`, escrita **después** de que el listener acepta | [§7.2](#72-la-ficha-del-daemon) |
| **A9** | **Conciliación de cobertura**: el `result` del stream-json declara los turnos que hubo; la diferencia con los medidos ES el «sin dato» | [§6.4](#64-conciliación-de-cobertura-de-dónde-sale-la-barra) |
| **A10** | **Dos relojes**: `ts_emisor` y `ts_recibido`; la ventana usa el nuestro | [§5.2](#52-ddl-completo) |
| **A11** | El refresco del catálogo va **apagado por default** (es egreso del daemon, D13) | [§8.3](#83-refresco-degradación-y-honestidad) |
| **A12** | El receptor **nunca bloquea al emisor**: cola acotada + descarte contado + `partialSuccess` en la respuesta OTLP | [§4.1](#41-el-receptor-otlp) |
| **A13** | La allowlist se aplica **dos veces**: en el emisor (hook) y en la puerta (daemon) | [§6.1](#61-allowlist-la-lista-explícita) |
| **A14** | `cwd` nunca se persiste crudo: se resuelve a `instalacion_id` o se guarda como huella | [§6.1](#61-allowlist-la-lista-explícita) |
| **A15** | Un evento que no resuelve a un arnés conocido **se guarda aparte** con `atribucion=sin-dato`; jamás se suma al total | [§6.3](#63-atribución-el-orden-de-preferencia) |
| **A16** | El «no aplica» de un detector **viaja al wire** como fila con `aplica:false`+`motivo`; omitirlo obliga al FE a inventar | [§9.2](#92-cómo-viaja-el-no-aplica) |
| **A17** | Subcomando `arnesia telemetria` — verificación E2E sin FE, mismo patrón que `arnesia portafolio` (S0-D9) | [§10.3](#103-cli-de-verificación) |
| **A18** | La tabla `runtime → aritmética/acumulación` la **declara cada adaptador** y el dominio solo la tipa | [§4.5](#45-normalización-l3-quién-es-dueño-de-la-tabla) |
| **A19** | **S2 tiene DOS modos**, no uno: `s2-instrumentado` (el arnés lleva bloque `env` ⇒ misma señal que S1) y `s2-degradado` (solo hook). El `escenario` es columna del almacén y campo del wire | [§7.0](#70-s2-tiene-dos-modos-h9) |
| **A20** | **Dónde vive el bloque `env` NO se resuelve acá**: dos opciones con sus consecuencias, recomendación marcada, **elige el operador**. ⚡ **Confirmada necesaria** (H10.1): un plugin **no** puede aportar el bloque, así que no hay tercera vía que evite la decisión | [§7.5](#75-dónde-vive-el-bloque-env--decisión-de-producto-abierta) |
| **A21** | Se persiste `tool_result_size_bytes` / `tool_input_size_bytes` desde el día 1 aunque B11 esté fuera del MVP: no cuesta nada ahora y el dato no se puede recuperar después | [§4.4](#44-mapeo-claude_code--evento-canónico) |
| **A22** | ⚡ **El token de ingesta NO viaja en el bloque `env`.** Claude Code no expande `${VAR}` ahí (H10.2) ⇒ `/v1/*` acepta **sin token** bajo loopback+Host allowlisted; el hook (que lee la ficha `0600` en runtime) **sí** lo exige siempre | [§7.3](#73-el-token-de-ingesta-es-otro-token) |

---

## 1 · Mapa de componentes

```
cmd/arnesia/
  main.go                     ← + wiring del módulo (composition root)
  telemetria.go        NUEVO  ← subcomando `arnesia telemetria` (A17)
  hook.go              NUEVO  ← subcomando `arnesia hook proceso` (A7)

internal/domain/
  telemetria.go        NUEVO  ← EventoTelemetria + enums + Tokens + LlaveJoin
  telemetria_costo.go  NUEVO  ← CalcularCosto (puro) + PrecioModelo
  telemetria_deteccion.go NUEVO ← Detector, ContextoDeteccion, PuntoDeMejora, los 6 detectores
  telemetria_vistas.go NUEVO  ← ResumenTelemetria, GastoCaja, TurnoUnido, SaludTelemetria

internal/ports/
  telemetria.go        NUEVO  ← TelemetriaSink · TelemetriaStore · CatalogoPrecios
                                 AtribucionRegistry · DescubrimientoDaemon · ForwardOTLP

internal/usecase/
  telemetria_service.go   NUEVO ← ingesta → costeo → persistencia; consultas del FE
  telemetria_mejoras.go   NUEVO ← corre los detectores y arma los puntos de mejora
  telemetria_retencion.go NUEVO ← purga por TTL + borrado por arnés + recomputo de rollup

internal/adapters/telemetria/
  otlp/receptor.go     NUEVO  ← http.Handler /v1/logs + /v1/metrics
  otlp/decodifica.go   NUEVO  ← el subconjunto OTLP/JSON, con json.Number
  otlp/mapa_cc.go      NUEVO  ← claude_code.* → EventoTelemetria (perfil del runtime)
  hooks/proyecta.go    NUEVO  ← payload de hook → EventoTelemetria (allowlist)
  store/store.go       NUEVO  ← modernc.org/sqlite (patrón de adapters/index/store.go)
  store/rollup.go      NUEVO  ← rollup horario incremental
  store/migracion.go   NUEVO  ← versionado aditivo + archivado
  catalogo/catalogo.go NUEVO  ← go:embed del catálogo de precios
  catalogo/precios.json NUEVO ← generado por `go generate` (LiteLLM filtrado, MIT)
  descubrimiento/ficha.go NUEVO ← escribe/lee la ficha del daemon

internal/adapters/agent/claudecode/
  conductor.go        TOCADO  ← Env del spawn (S1) + usage del `result` en AgentEvent (A5)

internal/adapters/transport/http/
  router.go           TOCADO  ← monta el receptor OTLP + las rutas /api/telemetria
  auth.go             TOCADO  ← isAPIPath cubre /v1/ ; token de ingesta (A6)
  telemetria.go        NUEVO  ← handlers de /api/telemetria/*

web/src/
  entities/telemetria/         NUEVO ← tipos + selectores + chips (confianza, no-aplica)
  widgets/map-canvas/model/layers.ts TOCADO ← slot `tokens` → «Mejora», enabled (D17.1)
  widgets/map-canvas/ui/…      TOCADO ← marcas del canvas + 4ª tab del inspector
  widgets/mejora/              NUEVO ← tarjeta de punto de mejora
  widgets/portafolio/          TOCADO ← fila arnés × puesto (D17.2)
```

**Lo que NO se crea, a propósito:** `adapters/telemetria/streamjson/` (A5, §4.2) ·
un segundo binario/sidecar (G7) · `otlpreceiver`/`pdata` (V5) · cualquier tabla dentro de
`index.db` (A2).

---

## 2 · El dominio

### 2.1 · `internal/domain/telemetria.go`

```go
package domain

import "time"

// Emisor dice POR QUÉ CANAL entró el evento. No es `origen` (facet L0 estandar|del-puesto
// del provisioner) ni `canal` (beta|estable|propuesto|deprecado) ni `via` (cómo cruzó una
// coincidencia de marketplace): los tres están tomados con otro significado — ver
// mockups/INDEX.md regla 4 y §2.2 de arquitectura-modulo.md.
type Emisor string

const (
	EmisorOTLP       Emisor = "otlp"        // receptor OTLP: dinero + latencia
	EmisorStreamJSON Emisor = "streamjson"  // frame `result` del subproceso: split 5m/1h
	EmisorHook       Emisor = "hook"        // hook del arnés: proceso (jamás dinero, H2)
	EmisorDaemon     Emisor = "daemon"      // decisión propia: rotación, corrida, gate
)

// Confianza dice CÓMO se atribuyó el evento a una unidad de trabajo, en orden de
// preferencia decreciente (ANEXO H5). Ningún número se muestra sin ella.
type Confianza string

const (
	ConfianzaExacta     Confianza = "exacta"      // resource attributes que inyectamos (S1)
	ConfianzaPorHash    Confianza = "por-hash"    // plugin_id_hash → tabla local
	ConfianzaPorProceso Confianza = "por-proceso" // cwd → instalación conocida
	ConfianzaSinDato    Confianza = "sin-dato"    // no se pudo atribuir. NUNCA se suma al total
)

// Aritmetica es propiedad del ADAPTADOR (D9.5/D-7): sin este flag el error de conteo es
// ~100 % en un sentido o ~28 % en el otro.
type Aritmetica string

const (
	AritmeticaDisjunta  Aritmetica = "disjoint"  // Anthropic: los 4 buckets no se solapan
	AritmeticaInclusiva Aritmetica = "inclusive" // OTel semconv / OpenAI: cached ⊂ input
)

// Acumulacion es la regla del adaptador para no contar dos veces (D9.5/D-6).
type Acumulacion string

const (
	AcumulacionPorRequest Acumulacion = "por-request" // cada evento es un delta propio
	AcumulacionAcumulada  Acumulacion = "acumulada"   // el último evento contiene a los previos
	AcumulacionUltima     Acumulacion = "ultima"      // el runtime re-emite el total
)

type TipoEvento string

const (
	EventoAPIRequest  TipoEvento = "api_request"  // OTLP: dinero
	EventoTurnoInicio TipoEvento = "turno_inicio" // hook UserPromptSubmit
	EventoHerramienta TipoEvento = "herramienta"  // hook PostToolUse
	EventoTurnoFin    TipoEvento = "turno_fin"    // hook Stop
	EventoSesionFin   TipoEvento = "sesion_fin"   // hook SessionEnd
	EventoRotacion    TipoEvento = "rotacion"     // daemon: rotación de contexto (B2)
	EventoCorrida     TipoEvento = "corrida"      // daemon: run T3 de una caja
	EventoGate        TipoEvento = "gate"         // daemon: veredicto del gate (P1)
)

// Resultado es el desenlace de un evento de proceso; vacío en los de dinero.
type Resultado string

const (
	ResultadoOK        Resultado = "ok"
	ResultadoRechazado Resultado = "rechazado"
	ResultadoReintento Resultado = "reintento"
	ResultadoCancelado Resultado = "cancelado"
)

// Tokens es el superset multi-proveedor. TODOS los campos son punteros a propósito
// (D-4 · boundary no-aplica-no-es-cero): nil = «este runtime no tiene el concepto»,
// que NO es lo mismo que 0. Un 0 donde el concepto no existe es una mentira.
type Tokens struct {
	Entrada          *int64 `json:"entrada,omitempty"`
	Salida           *int64 `json:"salida,omitempty"`
	CacheLectura     *int64 `json:"cache_lectura,omitempty"`
	CacheEscritura5m *int64 `json:"cache_escritura_5m,omitempty"`
	CacheEscritura1h *int64 `json:"cache_escritura_1h,omitempty"`
	Razonamiento     *int64 `json:"razonamiento,omitempty"`
}

// LlaveJoin es la unidad de trabajo. El par (SesionID, TurnoID) es LA llave del join
// dinero×proceso: el mismo turno lleva el mismo identificador en OTel (`session.id` +
// `prompt.id`) y en el payload del hook (`session_id` + `prompt_id`) — verificado, ANEXO H1.
// Sin TurnoID el join sería a nivel sesión y no alcanza para decir «el 60 % se va en la caja Y».
type LlaveJoin struct {
	ArnesID       string `json:"arnes_id,omitempty"`
	InstalacionID string `json:"instalacion_id,omitempty"`
	CajaID        string `json:"caja_id,omitempty"`
	SesionID      string `json:"sesion_id"`
	TurnoID       string `json:"turno_id,omitempty"`
	CorridaID     string `json:"corrida_id,omitempty"`
}

// EventoTelemetria es el ÚNICO evento canónico (D9.2). Los cuatro emisores producen este
// tipo y nada más; el almacén tiene una sola puerta de escritura.
//
// Lo que NO tiene, y no va a tener: identidad de cuenta (user.email, user.account_uuid,
// user.account_id, organization.id, user.id) ni contenido (prompt, respuesta, tool_response).
// Ver boundary ingesta-por-allowlist-declarada.md.
type EventoTelemetria struct {
	LlaveJoin

	Emisor           Emisor `json:"emisor"`
	Runtime          string `json:"runtime"`                     // "claude-code" | "codex" | …
	RuntimeVersion   string `json:"runtime_version,omitempty"`
	AdaptadorVersion string `json:"adaptador_version"`           // versiona el mapeo (D7.5)

	// TSRecibido es el reloj del daemon al aceptar el evento — es el que manda para
	// ventanas y rollups. TSEmisor es lo que dijo el emisor y puede ser basura (A10).
	TSRecibido      time.Time  `json:"ts_recibido"`
	TSEmisor        *time.Time `json:"ts_emisor,omitempty"`
	RelojSospechoso bool       `json:"reloj_sospechoso,omitempty"`
	DuracionMs      *int64     `json:"duracion_ms,omitempty"`

	Modelo         string `json:"modelo,omitempty"`
	ModeloCanonico string `json:"modelo_canonico,omitempty"` // tras el alias Bedrock/Vertex (E6)
	Proveedor      string `json:"proveedor,omitempty"`
	Speed          string `json:"speed,omitempty"`
	ServiceTier    string `json:"service_tier,omitempty"`

	Tokens      Tokens      `json:"tokens"`
	Aritmetica  Aritmetica  `json:"aritmetica,omitempty"`
	Acumulacion Acumulacion `json:"acumulacion,omitempty"`

	// Los DOS costos viajan siempre (D16.2 · boundary cifra-viaja-con-su-confianza):
	// Reportado es lo que dijo el runtime («Estimated cost», H4 — nunca facturación);
	// Calculado es lo que dice NUESTRO catálogo. Si divergen, el test de paridad A6 está
	// corriendo en producción y gratis. nil ≠ 0 en los dos.
	CostoReportadoMicros *int64 `json:"costo_reportado_micros,omitempty"`
	CostoCalculadoMicros *int64 `json:"costo_calculado_micros,omitempty"`
	CatalogoVersion      string `json:"catalogo_version,omitempty"`

	TipoEvento  TipoEvento `json:"tipo_evento"`
	Resultado   Resultado  `json:"resultado,omitempty"`
	Gate        string     `json:"gate,omitempty"`
	Motivo      string     `json:"motivo,omitempty"`
	Herramienta string     `json:"herramienta,omitempty"`

	PluginIDHash string    `json:"plugin_id_hash,omitempty"`
	CWDHuella    string    `json:"cwd_huella,omitempty"` // NUNCA la ruta cruda (A14)
	Atribucion   Confianza `json:"atribucion"`
}
```

Errores centinela del dominio (`telemetria.go`):

```go
var (
	// ErrPayloadInvalido — el cuerpo no es el wire format esperado. Se rechaza el LOTE,
	// jamás se acepta a medias en silencio.
	ErrPayloadInvalido = errors.New("telemetria: payload inválido")
	// ErrEventoSinSesion — un evento sin sesion_id no es atribuible ni deduplicable.
	ErrEventoSinSesion = errors.New("telemetria: evento sin sesion_id")
	// ErrCampoProhibido — un campo de la denylist dura (PII/contenido) llegó al constructor.
	// Es un BUG del adaptador, no un dato del usuario: se loguea con el nombre del campo.
	ErrCampoProhibido = errors.New("telemetria: campo prohibido en el evento canónico")
)
```

### 2.2 · Colisiones de vocabulario L0 resueltas

`mockups/INDEX.md` regla 4 obliga a verificar antes de nombrar. Tres palabras que el diseño
previo usaba **ya están tomadas con otro significado**:

| palabra | significado ya tomado | dónde | qué se usa en su lugar |
|---|---|---|---|
| `origen` | facet L0 `estandar` \| `del-puesto`, lo estampa el provisioner | `graph.l0.schema.json`, `web/src/entities/arnes/model/proposals.ts` | **`emisor`** (verificado libre en Go y TS) |
| `canal` | `beta` \| `estable` \| `propuesto` \| `deprecado` | `domain.CanalBeta` | — |
| `via` | cómo se cruzó una coincidencia de marketplace (`home-declarado`\|`faceta-registry`\|`rename`) | `domain.CoincidenciaPortafolio.Via` | — |
| `procedencia` | honestidad del dato: `medido`·`estimado`·`declarado`·`inferido`·`no-declarado` | `domain.Procedencia` | **`atribucion`** — eje distinto: `procedencia` dice *cómo se obtuvo el valor*, `atribucion` dice *a qué unidad de trabajo se lo asignó*. Un `api_request` es `procedencia: medido` y puede ser `atribucion: por-hash` a la vez |

Palabras nuevas del módulo, verificadas libres: **`emisor`** · **`atribucion`** ·
**`punto de mejora`** · **`contrafactual`** · **`cobertura`** · **`conciliación`**.

### 2.3 · Costeo — `internal/domain/telemetria_costo.go`

```go
// PrecioModelo es una fila del catálogo, en USD por token. Punteros: un modelo que no
// publica tarifa de cache 1h tiene nil, no 0 (aplanar los tiers es el bug phoenix#14314).
type PrecioModelo struct {
	ModeloCanonico    string   `json:"modelo"`
	Proveedor         string   `json:"proveedor"`
	Entrada           *float64 `json:"entrada,omitempty"`
	Salida            *float64 `json:"salida,omitempty"`
	CacheLectura      *float64 `json:"cache_lectura,omitempty"`
	CacheEscritura5m  *float64 `json:"cache_escritura_5m,omitempty"`
	CacheEscritura1h  *float64 `json:"cache_escritura_1h,omitempty"`
	Razonamiento      *float64 `json:"razonamiento,omitempty"`
	UmbralContextoTok *int64   `json:"umbral_contexto_tokens,omitempty"`
	SobreUmbral       *PrecioModelo `json:"sobre_umbral,omitempty"` // tier largo, NO aplanado
}

// CostoCalculado es el resultado del costeo con nuestro catálogo. `Completo:false` marca
// que ALGÚN bucket no tenía tarifa: el número es parcial y la UI lo dice, no lo redondea.
type CostoCalculado struct {
	Micros     int64    `json:"micros"`
	Completo   bool     `json:"completo"`
	SinTarifa  []string `json:"sin_tarifa,omitempty"` // buckets que no se pudieron cotizar
	Version    string   `json:"catalogo_version"`
}

// CalcularCosto cotiza un uso con un precio. Es pura y es donde viven los tres tests que
// los bugs ajenos dictaron (E8):
//   1. el cache WRITE se cobra (langfuse#14249 lo olvida → −28 %),
//   2. no se suman buckets que se solapan — de ahí el parámetro `a` (langfuse#12306 → 2×),
//   3. los tiers NO se aplanan: si hay `SobreUmbral` y el prompt lo pasa, se usa (phoenix#14314).
// Un bucket con tokens y sin tarifa NO se cobra a 0: sale en SinTarifa y Completo=false.
func CalcularCosto(t Tokens, p PrecioModelo, a Aritmetica) CostoCalculado
```

### 2.4 · Detección — `internal/domain/telemetria_deteccion.go`

```go
// Escenario distingue QUÉ SEÑAL EXISTE, no quién spawneó (ANEXO H9 corrigió la premisa:
// fuera de ArnesIA también puede haber señal completa, si el arnés lleva bloque `env`).
// Son tres niveles de dato distintos y la UI los distingue.
type Escenario string

const (
	// EscenarioS1 — spawn nuestro: env inyectado + stream-json + hook. Señal completa.
	EscenarioS1 Escenario = "s1"
	// EscenarioS2Instrumentado — fuera de ArnesIA, PERO el arnés lleva el bloque `env` en
	// sus settings ⇒ llega OTel con dinero. Sin stream-json (no somos el padre del proceso).
	EscenarioS2Instrumentado Escenario = "s2-instrumentado"
	// EscenarioS2Degradado — fuera de ArnesIA y sin bloque `env`: solo el hook ⇒ proceso,
	// jamás dinero (ANEXO H2). Los detectores de dinero se apagan CON MOTIVO.
	EscenarioS2Degradado Escenario = "s2-degradado"
)

// ContextoDeteccion es lo que un detector necesita para decidir si PUEDE correr. Se arma
// del dato real de la ventana, jamás se asume.
type ContextoDeteccion struct {
	Runtime   string
	Escenario Escenario
	// TieneCosto: llegó ALGÚN evento con dinero en la ventana. En s2 esto depende de si el
	// arnés lleva el bloque `env` (H9), no del escenario — por eso es un hecho de la ventana.
	TieneCosto          bool
	TieneSplitTTL       bool // el `result` del stream-json llegó (B1 lo exige)
	CatalogoDisponible  bool
	TieneSenalProceso   bool // llegaron eventos de hook/daemon (P1 los exige)
	TieneGateHumano     bool // hubo eventos de gate del daemon (P1 completo vs parcial)
	TieneEventoRotacion bool // el daemon registró rotaciones (B2)
}

// Aplicabilidad es la respuesta honesta de un detector a «¿podés correr acá?».
// Motivo es OBLIGATORIO cuando Aplica es false: un detector apagado sin razón es un gap
// escondido (boundary no-aplica-no-es-cero).
type Aplicabilidad struct {
	Aplica bool   `json:"aplica"`
	Motivo string `json:"motivo,omitempty"`
}

// PuntoDeMejora es la unidad del entregable: un número, su contrafactual, su umbral
// citado, su sesgo declarado y UN fix (regla A4). Sin las cinco cosas no se muestra.
type PuntoDeMejora struct {
	Detector      DetectorID `json:"detector"`
	ScoreVersion  int        `json:"score_version"` // A7
	Titulo        string     `json:"titulo"`
	CajaID        string     `json:"caja_id,omitempty"`
	GastoMicros   int64      `json:"gasto_micros"`
	ParteDelTotal float64    `json:"parte_del_total"` // 0..1
	// Contrafactual: qué habría costado el mundo alternativo, con las MISMAS corridas (A2).
	ContrafactualMicros int64   `json:"contrafactual_micros"`
	DiferenciaMicros    int64   `json:"diferencia_micros"`
	Umbral              string  `json:"umbral"`  // la desigualdad algebraica citada (A1)
	Sesgo               string  `json:"sesgo"`   // declarado y EN CONTRA de la recomendación (A3)
	Fix                 string  `json:"fix"`     // UNO, concreto
	Confianza           Confianza `json:"confianza"`
	CorridasUsadas      int     `json:"corridas_usadas"`
	CorridasTotales     int     `json:"corridas_totales"`
}

// Detector es el contrato común de los 6 del MVP (D16.1). Aplica() se consulta SIEMPRE
// antes que Evaluar(): un detector que no aplica devuelve su motivo y la UI lo pinta
// apagado con la razón, nunca en 0.
type Detector interface {
	ID() DetectorID
	Nombre() string
	Aplica(c ContextoDeteccion) Aplicabilidad
	Evaluar(v Ventana) []PuntoDeMejora
}

// Ventana es el corte de datos sobre el que corre un detector: los turnos unidos
// (dinero×proceso) de un arnés en un rango, ya costeados.
type Ventana struct {
	Desde, Hasta time.Time
	ArnesID      string
	Turnos       []TurnoUnido
	TotalMicros  int64
	Contexto     ContextoDeteccion
}

type DetectorID string

const (
	DetB4 DetectorID = "b4-gasto-por-arnes-empresa-puesto"
	DetP1 DetectorID = "p1-caja-que-consume-y-se-rechaza"
	DetB2 DetectorID = "b2-costo-de-la-rotacion"
	DetB6 DetectorID = "b6-sesion-abandonada"
	DetB3 DetectorID = "b3-cambio-de-modelo-invalida-cache"
	DetB1 DetectorID = "b1-rewarm-por-ttl"
)

// DetectoresMVP devuelve los 6 en orden estable. Es la lista completa: los otros 7 de la
// familia B se reportan como «no medidos todavía» por el servicio, no por un detector vacío.
func DetectoresMVP() []Detector
```

Aplicabilidad de los seis, tal como la declara cada uno (D16.1 + ANEXO):

| detector | aplica cuando | motivo cuando no | s1 | s2-instr. | s2-degr. |
|---|---|---|:--:|:--:|:--:|
| **B4** | `TieneCosto` y ≥1 turno atribuido | «ningún turno de esta ventana pudo atribuirse a una caja» · «este arnés no reporta costo: no lleva el bloque `env` de telemetría» | ✅ | ✅ | ❌ |
| **P1** | `TieneSenalProceso` | «este arnés no porta el hook de proceso — no hay veredicto de gate que leer» | ✅ | ⚠️ | ✅ |
| **B2** | `TieneEventoRotacion` (evento del daemon) | «la rotación es decisión de ArnesIA: corriendo fuera, no hay rotaciones que medir» | ✅ | ❌ | ❌ |
| **B6** | `TieneCosto` | «sin señal de dinero no se puede decir que una sesión se desperdició» | ✅ | ✅ | ❌ |
| **B3** | `TieneCosto` y ≥2 turnos con `Modelo` no vacío | «un solo modelo en la ventana» | ✅ | ✅ | ❌ |
| **B1** | `TieneSplitTTL` **y** `Runtime=="claude-code"` | «re-warm por TTL: el split 5m/1h viaja en el `result` del stream-json, que solo se ve cuando ArnesIA es el proceso padre» · otro runtime: «el break-even de TTL depende de la estructura 1,25×/2×/0,1× de Anthropic (D-5)» | ✅ | ⚠️ | ❌ |

⚠️ **Los dos casos que la corrección de H9 dejó matizados, y hay que escribirlos bien:**

- **B1 en `s2-instrumentado`.** D16.1 lo daba por «no aplica en S2» y H9 lo corrigió a medias:
  con bloque `env` llega el `api_request` (dinero) pero **no el `result` del stream-json**, que es
  donde vive el split `ephemeral_5m`/`ephemeral_1h` (V2). ⇒ B1 sigue **sin aplicar** en
  `s2-instrumentado`, pero **por otra razón** y con otro motivo en pantalla. La regla es
  `TieneSplitTTL`, que es un hecho de la ventana, no del escenario — así el día que Claude Code
  mande el split por OTel el detector se enciende solo, sin tocar código.
- **P1 en `s2-instrumentado`.** OTel trae parte de la señal de proceso (`tool_decision`,
  `tool_result` con `success` y `duration_ms` — ANEXO H8), pero **no** el veredicto del gate humano
  ni la corrida de caja, que son eventos del daemon. ⇒ P1 corre **parcial**: detecta reintentos y
  fracasos de herramienta, no rechazos de gate. Ese matiz viaja como `cobertura_parcial: true` +
  motivo, no como un ✅ liso.

### 2.5 · Los puertos — `internal/ports/telemetria.go`

Seis interfaces. Ninguna menciona OTLP, SQLite ni HTTP: los casos de uso no saben qué hay del
otro lado (`dominio-independiente-de-transporte` + `.go-arch-lint.yml`).

```go
package ports

import (
	"context"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TelemetriaSink es la ÚNICA puerta de escritura del módulo. Los cuatro emisores
// (receptor OTLP, adaptador de agente, hook, daemon) escriben por acá y por ningún otro lado.
//
// Ingerir NUNCA bloquea al llamador más allá de encolar: un receptor que hace esperar al
// agente degradaría el trabajo que mide (boundary telemetria-de-nacimiento, fail-open).
// Devuelve cuántos aceptó — la diferencia con len(evs) son descartes CONTADOS, no silencio.
type TelemetriaSink interface {
	Ingerir(ctx context.Context, evs []domain.EventoTelemetria) (aceptados int, err error)
}

// ConsultaTelemetria acota una lectura. Desde/Hasta en cero = «toda la historia retenida».
// Limite en cero = el default del adaptador; jamás «sin límite».
type ConsultaTelemetria struct {
	ArnesID       string
	InstalacionID string
	CajaID        string
	SesionID      string
	Desde, Hasta  time.Time
	Limite        int
}

// PurgaTelemetria parametriza un borrado. Los dos modos son excluyentes: TTL aplica la
// retención, ArnesID borra todo lo de un arnés (D15.3, el botón de la UI).
type PurgaTelemetria struct {
	AntesDe time.Time
	ArnesID string
}

// TelemetriaStore es el almacén. Extiende el sink porque el mismo adaptador escribe y lee,
// pero los casos de uso de ingesta piden SOLO el sink (interfaz mínima en el consumidor).
type TelemetriaStore interface {
	TelemetriaSink

	// Resumen agrega sobre el rollup horario. Devuelve punteros nil —no ceros— cuando no
	// hubo dato que agregar (boundary no-aplica-no-es-cero).
	Resumen(ctx context.Context, q ConsultaTelemetria) (domain.ResumenTelemetria, error)
	// PorCaja devuelve el gasto por caja, INCLUIDAS las cajas sin dato atribuible: la
	// ausencia viaja explícita con motivo, nunca omitida.
	PorCaja(ctx context.Context, q ConsultaTelemetria) ([]domain.GastoCaja, error)
	// Turnos hace el join dinero×proceso por (SesionID, TurnoID) — la llave verificada
	// (ANEXO H1). Es el drill-down: toca la tabla cruda, no el rollup.
	Turnos(ctx context.Context, q ConsultaTelemetria) ([]domain.TurnoUnido, error)
	// EsperarTurno registra que un turno OCURRIÓ, lo sepamos medir o no. Es el denominador
	// de la cobertura (§6.4); sin él, «cuánto medimos» se leería como «cuánto hubo».
	EsperarTurno(ctx context.Context, sesionID, turnoID, arnesID, cajaID string) error
	// Purgar aplica retención o borra un arnés; devuelve cuántas filas se fueron.
	// Recomputa el rollup de las horas afectadas dentro de la misma transacción.
	Purgar(ctx context.Context, p PurgaTelemetria) (int64, error)
	// Salud son los contadores del receptor + estado del almacén. Persistidos, no en
	// memoria: «cuántos descarté» es dato de honestidad y tiene que sobrevivir al reinicio.
	Salud(ctx context.Context) (domain.SaludTelemetria, error)
	// Close libera los dos handles. A diferencia del índice, este .db NO es desechable:
	// cerrar bien importa.
	Close() error
}

// AtribucionRegistry es la tabla aprendida `plugin_id_hash → arnés` (INFORME §V3). Se
// aprende del spawn controlado: nosotros instalamos el arnés, así que podemos observar qué
// hash le corresponde. NO se asume determinismo entre máquinas (V7.2, sin verificar).
type AtribucionRegistry interface {
	PorHash(hash string) (arnesID, instalacionID string, ok bool)
	// PorCWD resuelve el cwd de un hook a una instalación conocida del Portafolio
	// (identidad (home,id)). La ruta se usa y se descarta: nunca se persiste cruda.
	PorCWD(cwd string) (arnesID, instalacionID string, ok bool)
	// Aprender registra un par observado. `como` documenta de dónde salió, para poder
	// explicar después una atribución rara.
	Aprender(hash, arnesID, instalacionID, como string) error
}

// CatalogoPrecios cotiza un modelo. Un modelo desconocido devuelve ok=false — jamás un
// PrecioModelo en cero, que costearía todo gratis en silencio.
type CatalogoPrecios interface {
	Precio(modeloCanonico string) (domain.PrecioModelo, bool)
	// Canonizar aplica los alias de cloud (Bedrock/Vertex) al nombre crudo del runtime.
	// Un nombre que no matchea se devuelve tal cual: no se inventa una canonización.
	Canonizar(modelo string) string
	Version() domain.VersionCatalogo
}

// DescubrimientoDaemon publica dónde está el daemon para que el hook lo encuentre sin que
// nadie le pase nada (D6.5). Publicar se llama DESPUÉS de que el listener acepta: una ficha
// que nombra un puerto muerto es una mentira que el hook cobra en timeouts.
type DescubrimientoDaemon interface {
	Publicar(ctx context.Context, f domain.FichaDaemon) error
	Leer() (domain.FichaDaemon, error)
	Retirar() error
}

// ForwardOTLP es el escape hatch del OPERADOR (C2/D13), apagado por default. Recibe el
// evento YA PROYECTADO — jamás el cuerpo OTLP crudo, que llevaría el email de quien corra
// el arnés (INFORME §V6). Un arnés no puede alcanzar ni configurar esto.
type ForwardOTLP interface {
	Enviar(ctx context.Context, evs []domain.EventoTelemetria) error
	Activo() bool
	Destino() string
}
```

Errores que cruzan la frontera de puerto (además de los centinelas del dominio, §2.1):

```go
var (
	// ErrColaLlena — el sink descartó por saturación. El caller lo CUENTA y sigue; jamás
	// bloquea ni reintenta: la instrumentación no degrada al sistema instrumentado.
	ErrColaLlena = errors.New("telemetria: cola llena")
	// ErrAlmacenNoDisponible — el .db no se pudo abrir/escribir (disco lleno, corrupción).
	// El daemon SIGUE: se pierde telemetría, no la sesión del usuario.
	ErrAlmacenNoDisponible = errors.New("telemetria: almacén no disponible")
	// ErrSinFicha — no hay ficha de daemon publicada. Es la señal de fail-open del hook.
	ErrSinFicha = errors.New("telemetria: no hay daemon publicado")
)
```

---

## 3 · Transporte

### 3.1 · Un solo puerto

**Decisión A1.** El receptor OTLP se monta en el mux que ya existe (`:4200`), no en un
listener `:4318` propio. Se le pasa al runtime `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4200`
y el exportador OTLP/HTTP le concatena `/v1/logs` y `/v1/metrics` por spec.

Por qué:

1. **Un solo puerto que descubrir.** D6.5 ya advierte que el puerto entra en la config de un
   tercero; dos puertos son dos formas de perder telemetría en silencio.
2. **Un solo diálogo de firewall.** D6.3: bindear un segundo socket es una segunda chance de que
   Windows pregunte.
3. **Un solo confinamiento.** `superficie-local-confinada` ya cubre este mux con los tres gates;
   un listener aparte los duplicaría (y la duplicación es donde se cuelan los agujeros).
4. **Un solo shutdown.** El `srv.Shutdown` existente cubre todo.

Cambios exactos en el transporte:

- `router.go`: `mux.Handle("POST /v1/logs", otlp)` y `mux.Handle("POST /v1/metrics", otlp)`.
  El handler entra **inyectado como `http.Handler`**, igual que `events` hoy — `transport-http`
  no importa el paquete `telemetria/otlp` (go-arch-lint lo prohíbe y no hace falta).
- `auth.go`: `isAPIPath` pasa a
  `strings.HasPrefix(p, "/api") || p == "/events" || strings.HasPrefix(p, "/v1/")`.
  Sin esto el receptor quedaría **fuera del token gate** — un proceso local cualquiera podría
  envenenar el almacén (D6.4).

### 3.2 · API HTTP nueva

Base `/api` (server de `openapi.yaml`). Todas bajo los tres gates.

| método · ruta | qué devuelve | notas de contrato |
|---|---|---|
| `GET /api/telemetria/resumen` | `ResumenTelemetria` | query: `arnes`, `instalacion`, `desde`, `hasta` (RFC3339). Alimenta la franja de la barra (§1.2 del mockup) |
| `GET /api/telemetria/arneses/{clave}/cajas` | `[]GastoCaja` | una fila por caja **más** las cajas sin dato (`atribuible:false`), nunca omitidas |
| `GET /api/telemetria/arneses/{clave}/cajas/{cajaId}` | `DetalleCaja` | la 4ª tab del inspector: buckets con `null`, reportado vs calculado, join a nivel nodo, detectores |
| `GET /api/telemetria/arneses/{clave}/mejoras` | `RespuestaMejoras` | `puntos[]` + `no_aplican[]` con motivo + `no_medidos[]` (los otros 7) |
| `GET /api/telemetria/portafolio` | `[]FilaPortafolio` | una fila por **arnés × puesto** (D17.2); `costo_por_corrida: null` para el que nunca corrió |
| `DELETE /api/telemetria/arneses/{clave}` | `{borrados: N}` | D15.3, el botón «borrar la telemetría de este arnés» |
| `GET /api/telemetria/salud` | `SaludTelemetria` | recibidos · descartados · última recepción · retención · versión del catálogo · forward on/off |
| `POST /api/telemetria/proceso` | `202` | ingesta del hook de S2 (§7). Cuerpo = evento **ya proyectado**; el daemon lo re-valida |
| `POST /v1/logs` · `POST /v1/metrics` | `200` + `partialSuccess` | OTLP. Fuera de `/api` porque la ruta la fija la spec del protocolo |

Shapes (los que no son obvios):

```go
type ResumenTelemetria struct {
	Desde, Hasta time.Time `json:"desde","hasta"`
	// Estimado es SIEMPRE true mientras la fuente sea `cost.usage`/`cost_usd_micros`:
	// está documentado como "Estimated cost", no facturación (H4). La UI lo dice.
	Estimado             bool   `json:"estimado"`
	CostoReportadoMicros *int64 `json:"costo_reportado_micros"` // null = ningún turno lo trajo
	CostoCalculadoMicros *int64 `json:"costo_calculado_micros"` // null = sin catálogo aplicable
	Corridas             int    `json:"corridas"`
	Sesiones             int    `json:"sesiones"`
	Turnos               int    `json:"turnos"`
	Cobertura            Cobertura `json:"cobertura"`
	Catalogo             VersionCatalogo `json:"catalogo"`
}

// Cobertura es la barra de §1.2 del mockup. `Esperados` sale de la conciliación (A9):
// los turnos que SABEMOS que ocurrieron. La diferencia con la suma de los cuatro es
// el «sin dato» honesto — no se calcula por resta en el FE, viaja explícito.
type Cobertura struct {
	Esperados  int `json:"esperados"`
	Exacta     int `json:"exacta"`
	PorHash    int `json:"por_hash"`
	PorProceso int `json:"por_proceso"`
	SinDato    int `json:"sin_dato"`
	NoLlegaron int `json:"no_llegaron"` // esperados − medidos: el agujero WSL2/devcontainer
}

type GastoCaja struct {
	CajaID       string  `json:"caja_id"`
	Nombre       string  `json:"nombre"`
	Atribuible   bool    `json:"atribuible"`            // false ⇒ el FE pinta «sin dato atribuible»
	Motivo       string  `json:"motivo,omitempty"`      // obligatorio si Atribuible=false
	CostoMicros  *int64  `json:"costo_micros"`          // null, no 0, cuando no es atribuible
	Parte        *float64 `json:"parte,omitempty"`
	Confianza    Confianza `json:"confianza"`
	Marcas       []MarcaDeFuga `json:"marcas,omitempty"` // el detector NOMBRADO, nunca un ⚠ genérico
}
```

Regla de contrato transversal, que un implementador «prolijo» rompería (mismo aviso que
`openapi.yaml` ya lleva para `entradas: null`): **`null` y `0` significan cosas distintas en toda
esta superficie.** Convertir un `null` en `0` es el pass fabricado. Enforced por
`TestNoAplicaNoEsCeroEnElWire`.

`openapi.yaml` sube a `0.8.0-telemetria` con las 9 rutas y los schemas
`ResumenTelemetria`, `Cobertura`, `GastoCaja`, `DetalleCaja`, `PuntoDeMejora`,
`RespuestaMejoras`, `FilaPortafolio`, `SaludTelemetria`, `EventoProceso`.

---

## 4 · Ingesta

### 4.1 · El receptor OTLP

`internal/adapters/telemetria/otlp/receptor.go`:

```go
// Receptor es el http.Handler de /v1/logs y /v1/metrics. Es el ÚNICO lugar del árbol que
// conoce el wire format OTLP (mismo patrón que claudecode con stream-json).
type Receptor struct {
	sink     ports.TelemetriaSink
	perfiles map[string]domain.PerfilRuntime // por scope/namespace del payload
	maxBody  int64
	reloj    func() time.Time
	salud    *Contadores
}

func NewReceptor(sink ports.TelemetriaSink, opts Opciones) *Receptor
func (r *Receptor) ServeHTTP(w http.ResponseWriter, req *http.Request)
```

Comportamiento del handler, punto por punto:

1. **Método y ruta.** Solo `POST /v1/logs` y `POST /v1/metrics`; cualquier otra cosa `405`.
2. **Content-Type.** Solo `application/json`. Un `application/x-protobuf` responde
   **`415` con un cuerpo que nombra la causa** («este receptor habla OTLP/JSON; poné
   `OTEL_EXPORTER_OTLP_PROTOCOL=http/json`») — jamás un 200 que finge haber guardado.
   `pdata` sigue siendo el plan B documentado si algún runtime no deja elegir (V5).
3. **Tope de cuerpo.** `http.MaxBytesReader(w, req.Body, maxBody)` con `maxBody = 4 MiB`
   (§11). Superarlo es `413`, sin leer el cuerpo entero a memoria.
4. **Decodificación** (§4.3). Un error de forma es `400` + `ErrPayloadInvalido`; **el lote
   entero se rechaza**, nunca se guarda a medias.
5. **Encolado no bloqueante.** El handler empuja al canal del writer (cap. 4096). Si está
   lleno, **descarta y cuenta** — y lo dice en la respuesta OTLP:
   `{"partialSuccess":{"rejectedLogRecords":N,"errorMessage":"cola llena"}}` con `200`.
   Es lo que la spec OTLP pide y además es honesto: el emisor se entera. **El emisor jamás
   se bloquea** (A12) — un receptor que hace esperar al agente degradaría el trabajo que mide.
6. **Respuesta.** `200` + `{}` (o `partialSuccess`). Nunca `5xx` por un evento malo: un
   `5xx` hace reintentar al exportador y multiplica el daño.

### 4.2 · El canal stream-json NO tiene adaptador propio

**Decisión A5, y es una corrección a `arquitectura-telemetria.md` L2**, que dibuja
`adapters/telemetria/streamjson/ (parser por runtime)`.

Construirlo violaría `adaptadores-de-agente-intercambiables` («nada fuera del adaptador
conoce stream-json») y `conductor-no-parsea-jsonl` («el adaptador conductor es el dueño del
protocolo»), y crearía **dos decodificadores del mismo frame** que driftan por separado.

En su lugar:

- `ports.AgentEvent` gana, **solo en `EventResult`**, un puntero
  `Uso *domain.UsoDelTurno` con lo que el `result` ya trae y hoy se tira:
  los 4 buckets, el split `ephemeral_5m`/`ephemeral_1h`, `service_tier`, `speed`,
  `modelUsage[].costUSD`, `canonicalModel`, `provider`, `contextWindow`.
  El adaptador `claudecode` ya parsea ese frame para `ctxPct` — es una extensión de lo que hace.
- `SessionService` y `RunService`, que ya reciben esos eventos, construyen el
  `domain.EventoTelemetria` con `Emisor=streamjson` y lo mandan al sink.
- El día que entre un segundo runtime, su adaptador de agente hace lo mismo con su propio
  formato. **Un adaptador de agente por runtime, no dos.**

Consecuencia visible: **B1 (re-warm por TTL) solo existe en S1**, porque el split viaja en el
stream-json del subproceso que nosotros spawneamos. Ya está declarado así en D16.1 y en el
estado 3 del mockup. No es una limitación nueva: es la misma, ahora con la razón estructural
escrita.

### 4.3 · El decodificador OTLP/JSON

`internal/adapters/telemetria/otlp/decodifica.go`. Probado contra los payloads reales de
[`verificacion-2026-07-26/evidencia/`](verificacion-2026-07-26/evidencia/), que quedan como
**golden files del test** (8 métricas · 14 puntos · 57 log records).

**Qué subconjunto se parsea** (y nada más):

```go
// valorAtributo es UN valor de AnyValue del wire OTLP. Solo los cuatro tipos escalares que
// Claude Code emite; arrayValue/kvlistValue/bytesValue se IGNORAN (no los usa y decodificarlos
// sería superficie de ataque gratis).
type valorAtributo struct {
	StringValue *string      `json:"stringValue"`
	IntValue    *json.Number `json:"intValue"` // ⚠️ json.Number: ver abajo
	DoubleValue *float64     `json:"doubleValue"`
	BoolValue   *bool        `json:"boolValue"`
}

type atributo struct {
	Key   string        `json:"key"`
	Value valorAtributo `json:"value"`
}

// --- /v1/logs (canal PRIMARIO, V1) ---
type peticionLogs struct {
	ResourceLogs []struct {
		Resource  struct{ Attributes []atributo `json:"attributes"` } `json:"resource"`
		ScopeLogs []struct {
			Scope      struct{ Name string `json:"name"` } `json:"scope"`
			LogRecords []struct {
				TimeUnixNano json.Number `json:"timeUnixNano"`
				Attributes   []atributo  `json:"attributes"`
				Body         struct {
					StringValue *string `json:"stringValue"`
				} `json:"body"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	} `json:"resourceLogs"`
}

// --- /v1/metrics (SECUNDARIO) ---
type peticionMetricas struct {
	ResourceMetrics []struct {
		Resource     struct{ Attributes []atributo `json:"attributes"` } `json:"resource"`
		ScopeMetrics []struct {
			Metrics []struct {
				Name string `json:"name"`
				Unit string `json:"unit"`
				Sum  *struct {
					DataPoints             []puntoDato `json:"dataPoints"`
					AggregationTemporality int         `json:"aggregationTemporality"`
					IsMonotonic            bool        `json:"isMonotonic"`
				} `json:"sum"`
			} `json:"metrics"`
		} `json:"scopeMetrics"`
	} `json:"resourceMetrics"`
}

type puntoDato struct {
	Attributes   []atributo   `json:"attributes"`
	AsInt        *json.Number `json:"asInt"`
	AsDouble     *float64     `json:"asDouble"`
	TimeUnixNano json.Number  `json:"timeUnixNano"`
}
```

**`json.Number` no es una preferencia de estilo, es obligatorio.** Claude Code emite
`intValue` como **número JSON**, off-spec (el mapeo protobuf→JSON manda string para int64).
Un `*string` revienta con `cannot unmarshal number into Go struct field … of type string`;
un `*int64` reventaría con el runtime que sí cumple la spec. `json.Number` acepta las dos
formas y se convierte con `.Int64()`. **Verificado en vivo, V5.** Test:
`TestIntValueComoNumeroYComoString` con las dos variantes del mismo payload.

**Lo que se ignora, explícitamente:** `Gauge`/`Histogram`/`ExponentialHistogram`
(Claude Code manda `Sum` en las 4 métricas medidas) · `traceId`/`spanId` (no los usamos; y son
justamente donde `protojson` se rompe, F5) · `severityNumber`/`severityText` ·
`droppedAttributesCount` · `schemaUrl` · `arrayValue`/`kvlistValue`/`bytesValue`.
Un campo desconocido **no es un error**: `encoding/json` lo descarta y el evento entra igual.
Eso es lo que hace que una versión nueva del runtime no rompa el receptor.

**Temporalidad.** Las 4 métricas llegan `Sum · Delta · monotonic=true` (V5.1) ⇒ **se suma y
listo**, no hay que diferenciar contadores acumulados. Si algún día llega
`AggregationTemporality == 2` (Cumulative), el decodificador **no adivina**: marca el punto
`sin-dato` y lo cuenta en `salud.temporalidad_no_soportada`. Fabricar un delta a partir de un
acumulado sin estado previo es inventar.

**Cómo se rechaza lo malo sin tirar el daemon:**

| falla | respuesta |
|---|---|
| JSON sintácticamente roto | `400`, lote descartado, contador `payload_invalido++`, `slog.Warn` con los primeros 200 bytes **sin el cuerpo** |
| cuerpo > 4 MiB | `413`, no se lee entero |
| `timeUnixNano` no numérico | el evento entra con `TSEmisor=nil`; el reloj del daemon manda |
| lote sin `resourceLogs`/`resourceMetrics` | `200` + `partialSuccess` con 0 aceptados |
| pánico en el decodificador | `recover()` en el handler → `500` + `slog.Error`; **el daemon sigue vivo**. Es la red de seguridad, no la estrategia: el decodificador no debe entrar por ahí y el test de fuzz existe para eso |

Fuzz obligatorio: `FuzzDecodificarLogs` sembrado con los tres payloads de `evidencia/`.

### 4.4 · Mapeo `claude_code.*` → evento canónico

`internal/adapters/telemetria/otlp/mapa_cc.go`. Un log record con
`event.name == "claude_code.api_request"` produce un `EventoTelemetria` con
`TipoEvento=EventoAPIRequest`. Tabla de mapeo:

| atributo del wire | campo del evento | nota |
|---|---|---|
| `session.id` | `SesionID` | mitad de la llave del join |
| `prompt.id` | `TurnoID` | **la otra mitad** (ANEXO H1) |
| `arnesia.arnes` / `.instalacion` / `.caja` / `.corrida` | `ArnesID` / `InstalacionID` / `CajaID` / `CorridaID` | inyectados por nosotros; viajan **copiados en cada log record** (V4) ⇒ `Atribucion=exacta` |
| `model` | `Modelo`, y `ModeloCanonico` tras el alias | |
| `input_tokens` / `output_tokens` / `cache_read_tokens` / `cache_creation_tokens` | los 4 buckets | `CacheEscritura5m`/`1h` quedan **nil** acá: el split no viene por OTel, lo aporta el stream-json |
| `cost_usd_micros` | `CostoReportadoMicros` | entero, sin coma flotante |
| `duration_ms` | `DuracionMs` | |
| `speed`, `service_tier`, `query_source` | `Speed`, `ServiceTier`, y `query_source` alimenta B7 (fuera del MVP) | |
| `event.name == "plugin_loaded"` → `plugin_id_hash`, `plugin.scope` | `PluginIDHash` + alimenta `atribucion_hash` | la tabla `hash → arnés` (V3) |
| `event.name == "tool_decision"` → `decision`, `source`, `tool_name`, `tool_source` | `TipoEvento=EventoHerramienta`, `Resultado`, `Herramienta`, `Decision` | ANEXO H8. Es señal de **proceso** que llega por OTel, no por hook |
| `event.name == "tool_result"` → `tool_name`, `success`, `duration_ms`, `tool_input_size_bytes`, `tool_result_size_bytes` | `Herramienta`, `Resultado`, `DuracionMs`, `ToolInputBytes`, `ToolResultBytes` | **A21**: los dos tamaños se persisten desde el día 1 aunque B11 esté fuera del MVP. Son bytes, **no contenido** (verificado) — y un dato que no se guarda hoy no se recupera mañana |
| `event.name ∈ {hook_execution_start, hook_execution_complete}` | ignorados en el MVP | los cubre el hook propio con más contexto; se anota como reserva |
| **todo el resto** | **descartado en la puerta** | §6.1 |

`AdaptadorVersion = "cc-otlp/1"`. Cambiar el mapeo **obliga a bumpear ese string**: es lo que
permite explicar una divergencia histórica sin adivinar (D7.5).

### 4.5 · Normalización (L3): quién es dueño de la tabla

**Decisión A18.** `aritmetica` y `acumulacion` son **propiedad del adaptador** (D-6/D-7): cada
adaptador de ingesta declara su perfil y lo **estampa en cada evento**. El dominio solo tipa
el valor y el almacén lo guarda. No hay una tabla central que un agregador consulte —
justamente el error que D-6 prohíbe.

```go
// PerfilRuntime lo declara el adaptador, no el agregador (D-6/D-7).
type PerfilRuntime struct {
	Runtime          string
	AdaptadorVersion string
	Aritmetica       Aritmetica
	Acumulacion      Acumulacion
	// Detectores que este runtime PUEDE sostener; el resto se reporta «no aplica» (D-5).
	Soporta []DetectorID
}
```

Perfil de Claude Code, con dato verificado:
`{Runtime:"claude-code", Aritmetica: AritmeticaDisjunta, Acumulacion: AcumulacionPorRequest,
Soporta: [B4,P1,B2,B6,B3,B1]}`. En S2 el mismo perfil, sin B1 y sin B2 (el contexto de
detección lo apaga con motivo, no el perfil: el perfil es del runtime, la aplicabilidad es de
la corrida).

---

## 5 · Almacén

### 5.1 · Por qué una DB aparte

**Decisión A2: `~/.arnesia/telemetria.db`, separada de `index.db`.**

| | `index.db` | `telemetria.db` |
|---|---|---|
| doctrina | **desechable** — mismatch de `schema_version` ⇒ se borra el archivo entero y se reconstruye (`indice-desechable-jsonl-es-verdad` RF-208) | **no reconstruible** — el runtime no guarda esto en ningún lado; si se pierde, se perdió |
| escritura | esporádica, lotes grandes (`Rebuild`) | continua, eventos chicos |
| retención | ninguna (siempre refleja el árbol) | TTL + borrado por arnés (D15.3) |
| tamaño | decenas de KB | decenas de MB con retención |

Compartir archivo significaría que **un bump del esquema del índice borra la historia de
telemetría**. Eso es inaceptable y además no hay ganancia: no hay una sola consulta que haga
join SQL entre el grafo y la telemetría (el join lo hace el usecase por `arnes_id`).

### 5.2 · DDL completo

```sql
-- schema_meta: versión del esquema Y generación. Ver §5.4 para la diferencia.
CREATE TABLE IF NOT EXISTS schema_meta (
  generacion INTEGER NOT NULL,
  version    INTEGER NOT NULL,
  aplicado   TEXT    NOT NULL
) STRICT;

-- evento: la fila cruda. UNA tabla para los cuatro emisores, a propósito:
--   · el join dinero×proceso es un self-join por (sesion_id, turno_id) — un índice, no dos tablas;
--   · la purga y el borrado por arnés son una sentencia, no dos que se pueden desincronizar;
--   · las columnas de dinero quedan NULL en las filas de proceso, que es exactamente
--     «no aplica» a nivel de almacenamiento (los hooks NO traen dinero — ANEXO H2).
CREATE TABLE IF NOT EXISTS evento (
  id                     INTEGER PRIMARY KEY AUTOINCREMENT,
  ts_recibido            TEXT    NOT NULL,          -- RFC3339 UTC, reloj del daemon (manda)
  ts_emisor              TEXT,                      -- lo que dijo el emisor; NULL = no lo trajo
  reloj_sospechoso       INTEGER NOT NULL DEFAULT 0,
  emisor                 TEXT    NOT NULL,          -- otlp | streamjson | hook | daemon
  runtime                TEXT    NOT NULL,
  runtime_version        TEXT,
  adaptador_version      TEXT    NOT NULL,
  -- llave del join
  sesion_id              TEXT    NOT NULL,
  turno_id               TEXT,
  arnes_id               TEXT,
  instalacion_id         TEXT,
  caja_id                TEXT,
  corrida_id             TEXT,
  atribucion             TEXT    NOT NULL,          -- exacta | por-hash | por-proceso | sin-dato
  plugin_id_hash         TEXT,
  cwd_huella             TEXT,                      -- huella, NUNCA la ruta (A14)
  -- modelo
  modelo                 TEXT,
  modelo_canonico        TEXT,
  proveedor              TEXT,
  speed                  TEXT,
  service_tier           TEXT,
  -- tokens: NULL = no aplica. Jamás DEFAULT 0.
  tok_entrada            INTEGER,
  tok_salida             INTEGER,
  tok_cache_lectura      INTEGER,
  tok_cache_5m           INTEGER,
  tok_cache_1h           INTEGER,
  tok_razonamiento       INTEGER,
  aritmetica             TEXT,
  acumulacion            TEXT,
  -- dinero: los DOS costos, siempre
  costo_reportado_micros INTEGER,
  costo_calculado_micros INTEGER,
  costo_completo         INTEGER,                   -- 0 = algún bucket sin tarifa
  catalogo_version       TEXT,
  duracion_ms            INTEGER,
  -- proceso
  escenario              TEXT    NOT NULL,          -- s1 | s2-instrumentado | s2-degradado (A19)
  tipo_evento            TEXT    NOT NULL,
  resultado              TEXT,
  gate                   TEXT,
  motivo                 TEXT,
  herramienta            TEXT,
  decision               TEXT,                      -- accept | reject … (tool_decision, H8)
  tool_input_bytes       INTEGER,                   -- A21: habilita B11 después, cuesta 0 hoy
  tool_result_bytes      INTEGER                    -- A21. Son BYTES, no contenido
) STRICT;

-- El índice del JOIN. Es EL índice del módulo: la consulta que une dinero y proceso
-- (ANEXO H1) es `WHERE sesion_id = ? AND turno_id = ?`.
CREATE INDEX IF NOT EXISTS idx_evento_join    ON evento (sesion_id, turno_id);
CREATE INDEX IF NOT EXISTS idx_evento_ventana ON evento (ts_recibido);
CREATE INDEX IF NOT EXISTS idx_evento_arnes   ON evento (arnes_id, instalacion_id, ts_recibido);
CREATE INDEX IF NOT EXISTS idx_evento_caja    ON evento (arnes_id, caja_id, ts_recibido);

-- rollup_hora: la proyección que hace que el tablero cueste ~2 ms en vez de ~323 ms
-- (medido, F2: 190×). Incremental — nunca se recalcula entero salvo tras un borrado.
-- Las sumas son NULLABLE a propósito: SUM() sobre un grupo donde nadie tenía el bucket
-- devuelve NULL, y así el «no aplica» sobrevive a la agregación.
CREATE TABLE IF NOT EXISTS rollup_hora (
  hora                   TEXT    NOT NULL,          -- "2026-07-26T14" UTC
  arnes_id               TEXT    NOT NULL DEFAULT '',
  instalacion_id         TEXT    NOT NULL DEFAULT '',
  caja_id                TEXT    NOT NULL DEFAULT '',
  runtime                TEXT    NOT NULL,
  modelo_canonico        TEXT    NOT NULL DEFAULT '',
  emisor                 TEXT    NOT NULL,
  atribucion             TEXT    NOT NULL,
  eventos                INTEGER NOT NULL,
  turnos                 INTEGER NOT NULL,
  tok_entrada            INTEGER,
  tok_salida             INTEGER,
  tok_cache_lectura      INTEGER,
  tok_cache_5m           INTEGER,
  tok_cache_1h           INTEGER,
  tok_razonamiento       INTEGER,
  costo_reportado_micros INTEGER,
  costo_calculado_micros INTEGER,
  duracion_ms_suma       INTEGER,
  duracion_ms_cuenta     INTEGER,
  rechazados             INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion)
) STRICT, WITHOUT ROWID;

-- Cursor del rollup: hasta qué evento.id se agregó. Reanudable tras una caída.
CREATE TABLE IF NOT EXISTS rollup_cursor (
  unico         INTEGER PRIMARY KEY CHECK (unico = 1),
  ultimo_evento INTEGER NOT NULL,
  actualizado   TEXT    NOT NULL
) STRICT;

-- atribucion_hash: la tabla `plugin_id_hash → arnés` que rescata la atribución cuando el
-- runtime redacta el nombre a "third-party" (V3). La aprendemos porque NOSOTROS instalamos.
CREATE TABLE IF NOT EXISTS atribucion_hash (
  plugin_id_hash  TEXT PRIMARY KEY,
  arnes_id        TEXT NOT NULL,
  instalacion_id  TEXT NOT NULL DEFAULT '',
  visto           TEXT NOT NULL,
  como_se_aprendio TEXT NOT NULL   -- spawn-controlado | correlacion-en-sesion | manual
) STRICT;

-- turno_esperado: la conciliación de cobertura (A9). Lo escribe el daemon cuando el
-- stream-json declara un turno; `medido` se marca cuando llega su api_request.
-- La diferencia es el «no llegaron» honesto (WSL2, devcontainer, telemetría apagada).
CREATE TABLE IF NOT EXISTS turno_esperado (
  sesion_id TEXT NOT NULL,
  turno_id  TEXT NOT NULL,
  arnes_id  TEXT,
  caja_id   TEXT,
  ts        TEXT NOT NULL,
  medido    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (sesion_id, turno_id)
) STRICT, WITHOUT ROWID;

-- salud: contadores del receptor. Se persisten para que sobrevivan a un reinicio —
-- «cuántos descarté» es dato de honestidad, no un gauge en memoria que se pierde.
CREATE TABLE IF NOT EXISTS salud (
  clave       TEXT PRIMARY KEY,
  valor       INTEGER NOT NULL,
  actualizado TEXT NOT NULL
) STRICT;
```

`STRICT` está disponible en `modernc.org/sqlite` (verificado en la investigación §2).
`WITHOUT ROWID` en las dos tablas con PK compuesta natural.

### 5.3 · Concurrencia — el patrón de `index/store.go`, calcado

```go
// Store implementa ports.TelemetriaStore sobre modernc.org/sqlite (pure Go, sin CGO —
// check sin-cgo). Dos handles como en internal/adapters/index/store.go: writer con
// SetMaxOpenConns(1) (database/sql serializa toda escritura) y reader pooled bajo WAL.
type Store struct {
	writer *sql.DB
	reader *sql.DB
	cola   chan []domain.EventoTelemetria
	// …
}
```

DSN idénticos a los del índice:
`?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate`
para el writer; `?_pragma=busy_timeout(5000)` para el reader.

Una goroutine drena `cola` y escribe en lotes (hasta 256 eventos o 250 ms, lo que llegue
primero) dentro de una transacción con `INSERT` preparado. Medido en la investigación:
419 798 filas/s con ese patrón — la escritura no es el cuello de nada.

### 5.4 · Versionado y migración

**Decisión A3.** Dos números, no uno:

- **`version`** — el esquema crece por **migraciones aditivas** aplicadas en orden dentro de
  una transacción. Aditivo significa exactamente: `CREATE TABLE`, `CREATE INDEX`,
  `ALTER TABLE … ADD COLUMN` (con default o nullable). Un test escanea la lista de migraciones
  y **falla si aparece `DROP`, `RENAME` o `ALTER … DROP COLUMN`**.
- **`generacion`** — cambia solo cuando una ruptura no se puede expresar de forma aditiva.
  Al detectar mismatch de generación el store **archiva** el archivo
  (`telemetria.db` → `telemetria-gen<N>-<fecha>.db.archivada`, más sus `-wal`/`-shm`), arranca
  vacío **y lo dice**: `GET /api/telemetria/salud` devuelve
  `{"historia_archivada":"telemetria-gen1-2026-07-26.db.archivada"}` y la UI lo muestra.

**Por qué NO se calca el wipe del índice:** el índice se puede reconstruir del árbol; esto no.
Borrar telemetría en silencio para simplificar una migración es destruir el único registro que
existe. Archivar cuesta un `os.Rename` y conserva la posibilidad de recuperarla a mano.

```go
type Migracion struct {
	Version int
	SQL     []string // solo aditivo — enforced por TestMigracionesSonAditivas
}

const generacionActual = 1

var migraciones = []Migracion{ {Version: 1, SQL: []string{ /* el DDL de §5.2 */ }} }
```

### 5.5 · Rollup incremental

`store/rollup.go`. Corre (a) tras cada lote escrito, con un debounce de 2 s, y (b) al boot.

```sql
INSERT INTO rollup_hora (hora, arnes_id, …, eventos, turnos, tok_entrada, …)
SELECT strftime('%Y-%m-%dT%H', ts_recibido), COALESCE(arnes_id,''), …,
       COUNT(*), COUNT(DISTINCT turno_id), SUM(tok_entrada), …
  FROM evento WHERE id > :cursor
 GROUP BY 1,2,3,4,5,6,7,8
ON CONFLICT DO UPDATE SET
  eventos = eventos + excluded.eventos,
  turnos  = turnos  + excluded.turnos,
  -- La suma que preserva el NULL: si los dos lados son NULL, sigue NULL.
  tok_entrada = CASE WHEN rollup_hora.tok_entrada IS NULL AND excluded.tok_entrada IS NULL
                     THEN NULL ELSE COALESCE(rollup_hora.tok_entrada,0)+COALESCE(excluded.tok_entrada,0) END,
  …;
```

`COUNT(DISTINCT turno_id)` sobreestima si un turno cruza una frontera horaria; se acepta y se
documenta como **sesgo declarado, en contra de la recomendación** (A3): infla el denominador,
o sea baja el costo-por-turno que mostramos.

**Recomputo:** solo el borrado por arnés y el archivado disparan
`DELETE FROM rollup_hora WHERE arnes_id=?` + re-agregado de las horas afectadas.

---

## 6 · Privacidad, allowlist y atribución

### 6.1 · Allowlist: la lista explícita

**D15.1 + ANEXO H4: la allowlist rige los DOS caminos de ingesta**, y se aplica **dos veces**
(A13) — el hook proyecta antes de mandar, y el daemon re-valida al recibir. Nunca se confía en
que el emisor haya filtrado.

**Se persiste solo esto, del canal OTLP:**

`session.id` · `prompt.id` · `arnesia.arnes` · `arnesia.instalacion` · `arnesia.caja` ·
`arnesia.corrida` · `event.name` · `model` · `input_tokens` ·
`output_tokens` · `cache_read_tokens` · `cache_creation_tokens` · `cost_usd_micros` ·
`duration_ms` · `speed` · `service_tier` · `query_source` · `plugin_id_hash` · `plugin.scope` ·
`timeUnixNano` · `terminal.type` *(solo como `runtime_version` si viene versionado)* ·
`tool_name` · `tool_source` · `decision` · `source` · `success` ·
`tool_input_size_bytes` · `tool_result_size_bytes` *(los cinco últimos, de `tool_decision`/
`tool_result` — ANEXO H8; **tamaños en bytes, jamás el contenido**)*.

**Se persiste solo esto, del payload del hook:**

`session_id` · `prompt_id` · `hook_event_name` · `tool_name` · `duration_ms` ·
`permission_mode` · `reason` (de `SessionEnd`) · `cwd` **→ normalizado a `instalacion_id` o a
huella** (A14).

**Se descarta en la puerta, nombre por nombre:**

| campo | canal | por qué |
|---|---|---|
| **`user.email`** | OTLP (cada punto y cada log record) | 🔴 PII. Es el email real de la cuenta |
| **`user.account_uuid`** | OTLP | 🔴 identificador de cuenta |
| **`user.account_id`** | OTLP | 🔴 identificador de cuenta |
| **`user.id`** | OTLP | 🔴 hash sha256 de identidad — sigue siendo identidad |
| **`organization.id`** | OTLP | 🔴 identidad de la organización |
| `prompt` | hook `UserPromptSubmit` | 🔴 **el prompt completo, en claro** (ANEXO H4) |
| `last_assistant_message` | hook `Stop` | 🔴 **la respuesta del asistente, en claro** |
| `tool_response` | hook `PostToolUse` | 🔴 **lo que la herramienta leyó/escribió** |
| `tool_input` | hook `PreToolUse`/`PostToolUse` | contiene rutas y contenido |
| `transcript_path` | hook (los 6) | ruta del usuario. Y **no es un archivo**: en la corrida apuntó a un directorio (ANEXO H7) — que nadie asuma `.jsonl` |
| `cwd` crudo | hook (los 6) | ruta del usuario; se usa **solo** para resolver la instalación y después se tira |
| `background_tasks`, `session_crons` | hook `Stop` | sin uso, superficie gratis |
| `body.stringValue` del log record | OTLP | los eventos útiles viven en atributos; el cuerpo no aporta y puede traer texto |
| cualquier atributo **no listado arriba** | los dos | **default-deny**: la allowlist es la lista, no una sugerencia |

Forma del enforcement: el constructor `domain.NuevoEvento(campos map[string]any)` **no existe**.
El evento se arma campo por campo desde funciones de mapeo que solo leen las claves de la
allowlist — no hay una ruta de código donde un mapa entero se copie al evento. El test
`TestAllowlistNoPersistePII` corre los payloads reales de `evidencia/` (que traen la PII
redactada, más una fixture con la PII simulada) y verifica que **ninguna** de las cinco claves
aparece ni en el evento ni en el `.db` resultante (búsqueda de subcadena sobre el archivo).

### 6.2 · Forward opcional del operador (C2 / D13)

- **Apagado por default.** Se enciende con `--telemetria-forward <endpoint>` /
  `ARNESIA_TELEMETRIA_FORWARD`, nunca desde la API ni desde un arnés (D13, invariante 2).
- **Filtra en el borde**: reenvía el `EventoTelemetria` ya proyectado (que por construcción no
  tiene PII), **jamás el payload OTLP crudo**. Reenviar crudo exportaría el email de quien
  corra el arnés (V6.1).
- **Indicador visible** en la UI mientras esté encendido (`SaludTelemetria.forward`).
- Test: `TestForwardApagadoPorDefault` + `TestForwardNoReenviaCrudo`.

### 6.3 · Atribución: el orden de preferencia

Se resuelve en este orden (ANEXO H5), y el primero que acierta fija `atribucion`:

1. **`exacta`** — `arnesia.arnes`/`.instalacion`/`.caja` presentes. Solo en S1: los inyectamos
   nosotros y viajan copiados en cada punto (V4).
2. **`por-hash`** — `plugin_id_hash` resuelve en `atribucion_hash`. Da el **arnés**, no la caja.
3. **`por-proceso`** — `cwd` del hook mapea a una instalación conocida del Portafolio
   (identidad `(home,id)`). Da arnés + instalación, no la caja.
4. **`sin-dato`** — nada resolvió.

**A15: un evento `sin-dato` se guarda, pero no se suma al total.** Aparece en
`Cobertura.SinDato` y en el drill-down. Sumarlo al gasto de un arnés sería atribuir por
adivinanza — exactamente lo que `atribucion_confianza` existe para impedir.

⚠️ **No verificado (V7.2):** si `plugin_id_hash` es determinista **entre máquinas**. El diseño
no lo asume: la tabla se aprende **localmente**, del spawn controlado de S1
(`como_se_aprendio="spawn-controlado"`). Si resultara determinista, se podría sembrar al
instalar — es una mejora, no un cambio de diseño.

### 6.4 · Conciliación de cobertura: de dónde sale la barra

**Decisión A9.** La barra de cobertura del mockup (§1.2) no se puede calcular con lo que llegó
— eso solo dice cuánto medimos, no cuánto hubo. El denominador sale de otro lado:

- En **S1**, el `result` del stream-json cierra cada turno y trae su `prompt_id`. El daemon
  escribe una fila en `turno_esperado` **por cada turno que ocurrió**, la sepa medir o no.
- Cuando llega el `api_request` de ese `(sesion_id, turno_id)`, se marca `medido=1`.
- `Cobertura.NoLlegaron = COUNT(*) WHERE medido=0` con más de 5 minutos de antigüedad.

Eso convierte el agujero de WSL2/devcontainer (D6.6) de «un 0 que parece un dato» en **«2 de
7 turnos no reportaron telemetría»**, visible y explicable. En **S2** no hay denominador
independiente: la cobertura se reporta `esperados: null` y la UI dice «cobertura desconocida
fuera de ArnesIA», que es la verdad.

---

## 7 · S2 — el arnés que corre fuera de ArnesIA

### 7.0 · S2 tiene DOS modos (H9)

**La premisa original de S2 era falsa y hay que decirlo así.** El paquete asumía: *fuera de
ArnesIA no controlamos el spawn ⇒ solo queda el hook ⇒ proceso sin dinero*. La verificación H9
la rompió: **el bloque `env` de un `settings.json` enciende la telemetría OTel**, probado con
marcador distinto por variante, un solo receptor y las env vars del shell desarmadas
(`env -u`), en tres formas — `--settings <archivo>`, `.claude/settings.json` del proyecto sin
flags, y lo mismo con `--setting-sources project,local`.

⇒ **S2 se parte en dos escenarios que son dos niveles de dato distintos** (A19):

| | **s1** | **s2-instrumentado** | **s2-degradado** |
|---|---|---|---|
| quién spawnea | ArnesIA | el usuario (terminal/IDE) | el usuario |
| **cómo se enciende OTel** | env inyectado al spawn (§10.2) | **bloque `env` en los settings del arnés** (H9) | no se enciende |
| dinero (`api_request`) | ✅ | ✅ **misma señal que S1** | ❌ |
| split TTL 5m/1h (`result`) | ✅ | ❌ — no somos el proceso padre | ❌ |
| proceso por hook | ✅ | ✅ | ✅ |
| proceso por OTel (`tool_decision`/`tool_result`) | ✅ | ✅ | ❌ |
| gate humano / rotación / corrida (daemon) | ✅ | ❌ | ❌ |
| atribución típica | `exacta` (resource attributes) | `exacta` si el bloque `env` trae `OTEL_RESOURCE_ATTRIBUTES`; si no, `por-hash` | `por-proceso` (cwd) o `por-hash` |
| detectores del MVP que corren | 6 | 4 (+P1 parcial) | 1 (P1) |

**Consecuencias de diseño, todas obligatorias:**

1. **`escenario` es un campo de primera clase**: columna del almacén (§5.2), campo del wire y
   dimensión de la UI. Sin él, `s2-instrumentado` y `s2-degradado` se leerían como el mismo
   «S2» y el usuario no entendería por qué uno tiene plata y el otro no.
2. **El escenario se DERIVA, no se declara.** Lo decide el receptor con lo que llegó:
   hay `arnesia.corrida` o el evento vino por `streamjson` ⇒ `s1` · hay `api_request` sin
   corrida nuestra ⇒ `s2-instrumentado` · solo hay eventos de hook ⇒ `s2-degradado`.
   Un arnés no puede mentir sobre su propio nivel de instrumentación.
3. **B1 sigue sin aplicar en `s2-instrumentado`, pero por otra razón** (§2.4): falta el
   `result` del stream-json, no falta la telemetría. El motivo en pantalla cambia.
4. **El estado 3 del mockup («corrió fuera de ArnesIA») se parte en dos paneles.** Hoy dibuja
   un solo caso degradado; tiene que dibujar el instrumentado también, que es el bueno.
5. **El mecanismo de obligación gana una pata**: `arnes-declara-telemetria` deja de ser solo
   «declaralo en el sello» y pasa a poder exigir que el paquete **shipee el bloque `env`**
   — pero dónde lo shipea es §7.5, y no lo decide este documento.

### 7.1 · El hook es el propio binario

**Decisión A7.** El hook que viaja dentro del arnés **no es un script**: es
`arnesia hook proceso`.

Por qué:

- S2 está definido como *«el arnés en la MISMA máquina, fuera de ArnesIA»* (D12.1) ⇒ **el
  binario `arnesia` está instalado**. No hay runtime nuevo que shipear (D11 nivel A) ni
  intérprete que asumir: Python no está garantizado en Windows, Node tampoco (Claude Code 2.x
  es binario nativo), y un `.sh` no corre en `cmd.exe`.
- La proyección de campos (§6.1) es **el mismo código Go** que el daemon usa y testea. Dos
  implementaciones de la misma allowlist es la forma segura de que una se olvide de un campo.
- Firma de código: no agrega un binario más al problema G3/`tauri#11992`.

Contrato de `arnesia hook proceso`:

| aspecto | contrato |
|---|---|
| entrada | el payload del hook por **stdin**, tal cual lo manda Claude Code |
| salida stdout | **vacía**. Un hook que imprime inyecta texto al contexto del agente |
| **exit code** | **siempre 0**. Sin excepciones. Un `UserPromptSubmit` que sale ≠0 bloquea el turno del usuario |
| duración | tope duro de **250 ms**; pasado eso, abandona y sale 0 |
| destino | lee la ficha (§7.2) y hace `POST /api/telemetria/proceso` a loopback. **Jamás a red externa** (check `telemetria-no-egresa`) |
| sin daemon | **fail-open silencioso**: no reintenta, no escribe a disco, no avisa. El trabajo del usuario nunca se degrada por la instrumentación |
| qué manda | el `EventoTelemetria` **ya proyectado** (§6.1). Nunca su stdin |
| registro | un `slog` local a `~/.arnesia/logs/arnesia.log` cuando falla, **solo si el archivo ya existe** (no lo crea: sería efecto colateral de un hook) |

«Fail-open silencioso» significa **silencioso hacia el usuario**, no invisible: el gap aparece
después como `Cobertura.NoLlegaron` y como «sin dato» en la UI. Nunca como 0.

Eventos que el arnés declara y el scaffold estampa:
`UserPromptSubmit` (turno_inicio) · `PostToolUse` (herramienta) · `Stop` (turno_fin) ·
`SessionEnd` (sesion_fin). `PreToolUse` y `SessionStart` **no** se instrumentan en el MVP:
no aportan al join y sí superficie.

✅ **VERIFICADO (H10.3), ya no es un supuesto:** con un hook apuntando a un comando inexistente,
Claude Code cierra el turno con `exit 0`, `is_error: false`, stderr vacío, duración normal — y el
hook válido de control **sí** se ejecutó en la misma corrida. ⇒ **la propiedad fail-open que A7
necesita está garantizada por el runtime**, no hay que envolver el comando ni apoyarse en el
`timeout`. El caso cubierto es el arnés copiado a una máquina sin ArnesIA (S3, fuera de alcance,
pero pasa).

⚠️ **Ojo con el alcance de esa garantía: cubre que el binario FALTE, no que FALLE.** Un
`arnesia hook proceso` que existe y sale ≠0 sí puede romper un `UserPromptSubmit`. El contrato de
la tabla de arriba —**exit 0 siempre**— sigue siendo obligación nuestra y sigue teniendo su test
(`TestHookNoTardaNiFalla`).

### 7.2 · La ficha del daemon

**Decisión A8.** Archivo de descubrimiento en
**`os.UserConfigDir()/arnesia/daemon.json`** — no en `~/.arnesia`.

Por qué esa ruta y no la convención del repo: es el **único** archivo que tiene que encontrar
un proceso que no conoce nuestras convenciones, en las tres plataformas (D6.7). En Windows es
`%APPDATA%\arnesia\daemon.json`, en macOS `~/Library/Application Support/arnesia/daemon.json`.
El resto del estado sigue en `~/.arnesia` — migrarlo entero es otra deuda, registrada aparte.

```jsonc
{
  "version": 1,
  "pid": 48213,
  "endpoint": "http://127.0.0.1:4200",
  "ruta_proceso": "/api/telemetria/proceso",
  "otlp": { "logs": "/v1/logs", "metrics": "/v1/metrics" },
  "token_ingesta": "…",          // el token ACOTADO, no el de la API (§7.3)
  "binario": "/home/u/.local/bin/arnesia",
  "desde": "2026-07-26T14:03:11Z"
}
```

- **Quién la escribe:** el daemon, en `runServe`, **después** de que el listener acepta
  conexiones — no antes. Una ficha que nombra un puerto donde nadie escucha es una mentira que
  el hook cobra en timeouts.
- **Cómo:** escritura atómica (temp en el mismo dir + `os.Rename`), permisos **`0600`**,
  directorio `0700`.
- **Cuándo se borra:** en el shutdown ordenado. Una ficha huérfana (daemon muerto sin
  shutdown) **no se detecta con el pid** — el hook simplemente falla el POST en 250 ms y sale
  fail-open. Es más simple y más robusto que un liveness check.
- **Dos daemons:** último que arranca gana la ficha. Los eventos de S1 no se ven afectados
  (cada spawn lleva su endpoint en el env del proceso). Se documenta y se acepta.

### 7.3 · El token de ingesta es OTRO token

**Decisión A6, y es de seguridad.**

La ficha tiene que llevar un token o el hook no pasa el gate (§3.1). Pero poner ahí el token
de la API **destruiría** la garantía de `superficie-local-confinada`: ese token existe
precisamente para que un proceso local cualquiera no pueda conducir el agente, y un archivo
`0600` legible por cualquier proceso **del mismo usuario** lo regala.

Solución: el daemon mint **dos** tokens.

| token | quién lo tiene | qué abre |
|---|---|---|
| `ARNESIA_AUTH_TOKEN` (el de hoy) | el shell Tauri, por env | **todo** `/api` + `/events` |
| **`token_ingesta`** (nuevo) | la ficha `0600` + el env del spawn | **solo** `POST /v1/logs`, `POST /v1/metrics`, `POST /api/telemetria/proceso` |

Filtrar el token de ingesta concede «podés escribirme telemetría basura» (que ya se mitiga con
la atribución y el rate limit), no «podés dirigir un agente con acceso al filesystem».

Enforcement: `validToken` gana una variante por ruta; test
`TestTokenDeIngestaNoAbreLaAPI` — 200 en `/v1/logs`, **401 en `/api/sessions`** con el mismo token.

#### ⚡ Cómo llega el token a cada camino — y por qué a uno NO llega (A22)

La versión anterior de esta sección asumía que el token llegaba solo. **No llega.** H10.2 verificó
que Claude Code **no expande `${VAR}` dentro del bloque `env`**: con
`OTEL_RESOURCE_ATTRIBUTES=arnesia.expand=${MARCA_PROPIA}` llegó el literal `${MARCA_PROPIA}`, con
control positivo presente en la misma corrida. *(Matiz que importa y que conviene no
sobre-generalizar: en los **comandos de hook** la expansión **sí** funciona — `${CLAUDE_PLUGIN_ROOT}`
es de uso corriente. La limitación es específica del bloque `env`.)*

| camino | cómo llega el token | ¿exige token? |
|---|---|---|
| **S1** — spawn nuestro | lo ponemos nosotros al armar el env del subproceso (§10.2). No hay interpolación de por medio: construimos la cadena en Go | **sí** |
| **hook** (`POST /api/telemetria/proceso`) | el hook lee la ficha `0600` **en runtime**, en cada invocación. Nada versionable | **sí, siempre** |
| **`s2-instrumentado`** (`/v1/*` desde un bloque `env`) | **no llega**: el archivo es estático y no interpola | **no** — ver decisión |

**Decisión A22: `/v1/logs` y `/v1/metrics` aceptan sin token cuando la request pasa el Host gate
loopback.** El token sigue siendo obligatorio en `POST /api/telemetria/proceso` y en todo `/api`.

Por qué es aceptable, explícito para que nadie lo lea como un descuido:

- **Lo que se pierde está acotado por diseño previo.** El token de ingesta ya estaba definido
  (A6) como «podés escribirme telemetría», nunca como «podés conducir el agente». Sin él, un
  proceso local del mismo usuario puede inyectar ruido en `/v1/*` — y nada más.
- **El ruido inyectado no contamina ningún número.** Un evento que no resuelve a un arnés conocido
  entra `atribucion=sin-dato` y **no suma a ningún total** (A15); aparece en `Cobertura.SinDato` y
  en `salud`, o sea que la inyección es **visible**, no silenciosa.
- **La alternativa es peor.** Poner el token literal en un `settings.json` mete un secreto vivo en
  un archivo que —en la opción A de §7.5— **está versionado en el repo del arnés**: iría a git, se
  publicaría con el paquete y no tendría rotación posible. Cambiar «un local puede mandarme ruido
  contable» por «el token de esta máquina viaja en un release» es un mal negocio.
- **Sigue habiendo tres barreras** en ese camino: Host gate loopback (anti DNS-rebinding), tope de
  cuerpo de 4 MiB, y rate limit del receptor.

**Escotilla para el operador que quiera lo contrario:** `--telemetria-ingesta-token-obligatorio`
exige token también en `/v1/*`. Consecuencia honesta, y la UI la dice: **`s2-instrumentado` deja de
funcionar** salvo que el operador acepte poner el token literal en un settings **no versionado**
(`0600`, en `.gitignore`, con rotación manual). No es el default y no se recomienda.

### 7.4 · Los cuatro checks de conformance del arnés

Ya declarados en `arquitectura-telemetria.md` §Parte 3 y en el boundary v2.3. Se construyen
como checks del motor `arnesia conformance` (mecanismo `static-scan` sobre el paquete):

| check | qué exige |
|---|---|
| `arnes-declara-telemetria` | `arnes.l0.json` tiene bloque `telemetria: {version, eventos[], atribucion}` |
| `arnes-porta-hook-proceso` | el paquete shipea la config de hooks del kit, con los 4 eventos |
| `hook-es-fail-open` | el comando del hook tiene `timeout` declarado y no hay ruta de código que salga ≠0 |
| `telemetria-no-egresa` | el hook solo apunta a loopback o a disco local |
| **`hook-proyecta-campos` (nuevo, ANEXO H4)** | el hook **no reenvía su stdin**: solo emite los campos declarados. Un hook que postea el payload entero a `127.0.0.1` cumple `telemetria-no-egresa` y aun así filtra la conversación al almacén local |

### 7.5 · Dónde vive el bloque `env` — decisión de producto, ABIERTA y ahora CONFIRMADA NECESARIA

**A20: este documento plantea las opciones y recomienda. NO decide. La elección es del
operador**, porque una de las dos ramas escribe archivos de un tercero y eso ya tiene doctrina
firmada en contra (A8 + el guardrail vigente de que el alcance del chat embebido excluye el
paquete propio).

⚡ **La escapatoria que habría eliminado esta decisión NO existe (H10.1, verificado).** Se armó un
marketplace local con un plugin cuyo `plugin.json` declara el bloque `env` completo, se instaló y
se corrió con las env vars del shell desarmadas: **0 payloads**. Control positivo acto seguido,
mismo receptor: **2 payloads** — el receptor estaba vivo. **El bloque `env` de `plugin.json` se
ignora.** ⇒ **no hay tercera vía**: no existe forma de que el arnés se instrumente solo al
instalarse. La decisión A/B de abajo hay que tomarla.

El bloque a colocar es, textualmente:

```jsonc
// settings.json — bloque `env`, la vía verificada (H9)
{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "http/json",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4200",
    "OTEL_RESOURCE_ATTRIBUTES": "arnesia.arnes=<id>,arnesia.instalacion=<clave>"
  }
}
```

⚠️ Nótese lo que **no** lleva: el `token_ingesta`. No es una omisión — **es imposible**:
Claude Code no expande `${VAR}` dentro del bloque `env` (H10.2), y meter el valor literal en un
archivo versionable sería publicar un secreto. **Resuelto en A22 (§7.3): `/v1/*` acepta sin token
bajo Host loopback.** Este bloque no necesita más nada.

| | **Opción A — en el repo del propio arnés** | **Opción B — en el proyecto del usuario** |
|---|---|---|
| dónde | `<arnés>/.claude/settings.json`, shipeado en el paquete | `<proyecto-del-usuario>/.claude/settings.json` |
| ¿archivo de quién? | **nuestro** | **de un tercero** |
| fricción | cero: viaja con el paquete, se instala solo | requiere escribir en el árbol del usuario |
| ¿choca con A8? | no | **sí** — A8 exige backup + confirmación explícita |
| ¿choca con el guardrail del chat embebido? | no | sí, es la misma clase de escritura |
| ¿cuándo aplica? | cuando el **cwd de la corrida es el árbol del arnés** — o sea, cuando el usuario está *trabajando el arnés* | cuando el arnés se *usa* sobre otro proyecto, que es el caso normal |
| cobertura real | **parcial** — cubre el desarrollo del arnés, no su uso | **completa** |
| reversibilidad | trivial (es nuestro archivo) | requiere deshacer una escritura ajena |

**Recomendación: A ahora, B solo con consentimiento explícito y por pedido del usuario.**
Razones: (1) A no toca nada de nadie y ya entrega el caso `s2-instrumentado` completo para el
escenario que más nos importa hoy —el operador iterando su propio arnés—; (2) B es una escritura
en árbol ajeno y la doctrina de la casa dice que eso se pide, no se hace; (3) el degradado de B
es honesto: sin el bloque queda `s2-degradado`, que **muestra proceso y dice por qué no hay
dinero**, en vez de un cero.

Si se elige B, el diseño mínimo es: un botón «instrumentar este proyecto» en la tarjeta del
Portafolio → diff a la vista → backup del `settings.json` previo → escritura → y un botón
inverso que revierte. Nada de eso se construye sin la firma.

**El problema del token: cerrado, no pendiente.** Las dos salidas que se habían planteado como
hipótesis quedaron descartadas por medición: la interpolación **no funciona** en el bloque `env`
(H10.2) y un token estable embebido en un archivo versionable es publicar un secreto. **La salida
que queda es la que se toma (A22, §7.3): `/v1/*` acepta sin token bajo Host loopback**, con la
escotilla `--telemetria-ingesta-token-obligatorio` para el operador que prefiera lo contrario y
acepte perder `s2-instrumentado`.

**Dos huecos que esta verificación cerró y por qué importa dejarlo escrito:**

| hipótesis del plan | resultado (H10) | consecuencia |
|---|---|---|
| «un plugin podría aportar el bloque `env` ⇒ A20 desaparece» | ❌ **0 payloads** con el plugin instalado; **2** en el control positivo | A20 **hay que decidirla**; no hay auto-instrumentación al instalar |
| «`${VAR}` se expande en el bloque `env` ⇒ el token entra por indirección» | ❌ llegan literales (`arnesia.expand = ${MARCA_PROPIA}`) | A22: el canal OTLP de `s2-instrumentado` va sin token |

Las dos con **control positivo en la misma corrida** — que es la única razón por la que se pueden
escribir como resultado y no como sospecha (§14).

---

## 8 · Catálogo de precios

### 8.1 · Cómo se embebe

```
internal/adapters/telemetria/catalogo/
  gen.go        //go:generate go run ./cmd/bajar-catalogo -rev <sha> -out precios.json
  precios.json  //go:embed — LiteLLM filtrado, ~182 KB (medido)
  catalogo.go   // implementa ports.CatalogoPrecios
  LICENSE-litellm.md  // MIT, aviso de atribución
```

`go generate` baja `model_prices_and_context_window.json` de un **rev fijado** de LiteLLM
(MIT, raíz — el carve-out `enterprise/` no toca esto), lo filtra a los proveedores en uso y lo
compacta. Patrón tomado del `build.rs` de ccusage (E2), **sin heredar su bug**: ccusage colapsa
todo a un solo `cache_create` y pierde
`cache_creation_input_token_cost_above_1hr` — justo el campo que habilita B1/B12 (E3).

El JSON generado **se commitea**. El build es offline y reproducible; `go generate` es un acto
deliberado del que actualiza precios, no un paso del build.

### 8.2 · Formato y versionado

```go
type VersionCatalogo struct {
	Version   string    `json:"version"`    // "2026-07-20" — fecha del rev de LiteLLM
	Rev       string    `json:"rev"`        // sha del commit de origen
	SHA256    string    `json:"sha256"`     // del archivo embebido
	Modelos   int       `json:"modelos"`
	Refrescado *time.Time `json:"refrescado,omitempty"` // nil = nunca; se muestra como tal
}
```

Alias Bedrock/Vertex: se porta el **regex** de Langfuse (MIT, E6) para normalizar
`(eu.|us.|apac.)?anthropic.claude-…` a `modelo_canonico`. Un modelo que no matchea **no se
inventa**: `Precio()` devuelve `ok=false`, el costo calculado queda `nil` y la UI lo dice.

### 8.3 · Refresco, degradación y honestidad

**Decisión A11: el refresco va APAGADO por default.** Es tráfico saliente del daemon, y D13
dejó firmado que el egreso del daemon es decisión deliberada del operador. Se enciende con
`--telemetria-catalogo-refresco`; corre en background, con timeout, y **si falla no pasa nada**:
se sigue con el embebido.

Los tres estados, todos visibles (G2):

| estado | qué muestra la UI |
|---|---|
| embebido, nunca refrescado (**default**) | «precios del release · catálogo `v2026-07-20`» |
| refrescado hoy | «catálogo `v2026-07-26`, refrescado hace 2 h» |
| refresco encendido pero fallando | ⚠️ «precios del release, sin refrescar desde el 20/07» (estado 5 del mockup) |

Modelo sin tarifa ⇒ `costo_calculado_micros = null` y `sin_tarifa: ["cache_escritura_1h"]`.
**Nunca 0.** Un 0 en dinero se lee como «gratis».

---

## 9 · Detectores y superficie

### 9.1 · Cómo corre el motor

`usecase/telemetria_mejoras.go`:

```go
func (s *TelemetriaService) Mejoras(ctx context.Context, q ports.ConsultaTelemetria) (domain.RespuestaMejoras, error)
```

1. arma la `Ventana` (turnos unidos por `(sesion_id, turno_id)`, ya costeados),
2. arma el `ContextoDeteccion` **del dato real** de esa ventana (no de una config),
3. por cada detector de `domain.DetectoresMVP()`: `Aplica(c)` primero. Si no aplica, se agrega
   a `no_aplican[]` con su motivo. Si aplica, `Evaluar(v)` y sus puntos van a `puntos[]`,
4. los 7 detectores fuera del MVP se emiten en `no_medidos[]` con
   `motivo: "no medido todavía"` — patrón `sin-check`, jamás cero fabricado.

### 9.2 · Cómo viaja el «no aplica»

**Decisión A16.** La respuesta lleva las tres listas, siempre:

```json
{
  "puntos": [ { "detector": "b2-costo-de-la-rotacion", "…": "…" } ],
  "no_aplican": [
    { "detector": "b1-rewarm-por-ttl",
      "nombre": "re-warm por TTL",
      "motivo": "no disponible: este arnés corrió fuera de ArnesIA y el hook no ve el stream-json" }
  ],
  "no_medidos": [ { "detector": "b5-cache-escrito-y-nunca-leido", "motivo": "no medido todavía" } ]
}
```

Omitir `no_aplican` obligaría al FE a decidir entre no mostrar nada (gap escondido) o mostrar
0 (mentira). Por eso viaja. Es el boundary `no-aplica-no-es-cero` aplicado al wire.

### 9.3 · Frontend

- `web/src/widgets/map-canvas/model/layers.ts`: el slot `tokens` pasa a
  `{ id: "mejora", label: "Mejora" }` y se enciende (D17.1). **Es desviación declarada de un
  baseline firmado** y va al gate como tal; los 4 slots se conservan.
- `web/src/entities/telemetria/`: tipos + `ChipConfianza` (subrayado punteado para todo lo que
  no sea `exacta`) + `ValorOSinDato` (el componente que renderiza `null` como «sin dato» y
  **nunca** como 0). Es el punto único donde se decide cómo se ve un `null`.
- `web/src/widgets/mejora/`: la tarjeta de punto de mejora (contrafactual · umbral · sesgo ·
  un fix · `score v1`). `[Aplicar]` **abre el chat con el cambio propuesto** (D17.3), no
  escribe archivos.
- Verificación: **story = test**. Cada uno de los 7 estados honestos es una story con `play()`
  bajo `vitest --project=storybook`, que corre headless en este entorno (verificado 2026-07-26).

---

## 10 · Wiring, CLI y fitness

### 10.1 · Composition root

En `runServe`, después del índice y antes del handler:

```go
telStore, err := telstore.New("", telstore.Opciones{Retencion: *telRetencion})
catalogo := catalogo.Embebido()
telSvc := usecase.NewTelemetriaService(telStore, catalogo, arnesReg, pfStore, domain.DetectoresMVP(), time.Now)
receptorOTLP := otlp.NewReceptor(telSvc, otlp.Opciones{MaxBody: 4 << 20})
ficha := descubrimiento.New("") // os.UserConfigDir()/arnesia/daemon.json
```

`ficha.Publicar(...)` se llama **después** de que `ListenAndServe` está aceptando (en una
goroutine que espera a que `/healthz` responda, o con un `net.Listener` explícito creado antes
de `srv.Serve(ln)` — la segunda es la limpia y es la que se recomienda).

El spawn (S1) inyecta el env: `claudecode.Conductor` gana un campo `EnvExtra []string` que
`Spawn` aplica con `cmd.Env = append(os.Environ(), …)`. **Hoy `cmd.Env` es nil** (hereda el del
daemon) — es un cambio real, no una configuración.

### 10.2 · Contrato del spawn en S1

Env exacto que el daemon inyecta en cada `claude` que spawnea:

```
CLAUDE_CODE_ENABLE_TELEMETRY=1
OTEL_LOGS_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_PROTOCOL=http/json
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4200
OTEL_EXPORTER_OTLP_HEADERS=x-arnesia-token=<token_ingesta>
OTEL_METRIC_EXPORT_INTERVAL=10000
OTEL_LOGS_EXPORT_INTERVAL=5000
OTEL_RESOURCE_ATTRIBUTES=arnesia.arnes=<id>,arnesia.instalacion=<clave>,arnesia.caja=<cajaId>,arnesia.corrida=<runId>
```

Notas que no son opcionales:

- **`http/json` y no `http/protobuf`.** Invierte F4. Es lo que hace barato al decodificador
  (+0,49 MB vs +10,79 MB, V5). El default de Claude Code es **gRPC :4317**: sin esta línea no
  llega nada.
- **`OTEL_EXPORTER_OTLP_ENDPOINT` sin `/v1/...`**: el exportador concatena la ruta.
- **`OTEL_RESOURCE_ATTRIBUTES`** es el vector de atribución y **viaja copiado a cada punto y a
  cada log record** (V4). `arnesia.caja` solo se pone en el spawn del conductor T3 (una corrida
  de caja); en el Dock queda vacío y la atribución es a nivel sesión.
- El token va por `OTEL_EXPORTER_OTLP_HEADERS` (D6.4), no por query string.
- ⚠️ **`ABIERTO`:** los nombres `OTEL_LOGS_EXPORT_INTERVAL` / `OTEL_LOGS_EXPORTER` no se
  verificaron contra `claude 2.1.220` en la corrida del 26/07 (se usó el default). Confirmar
  antes de fijarlos; si no existen, se acepta el intervalo por default y se documenta.

### 10.3 · CLI de verificación

`arnesia telemetria <subcomando>` — mismo patrón que `arnesia portafolio` (S0-D9): reusa el
MISMO usecase que el HTTP, cero lógica propia.

```
resumen [--arnes X] [--desde …]   emite el ResumenTelemetria (JSON)
mejoras --arnes X                  emite puntos + no_aplican + no_medidos (JSON)
salud                              contadores del receptor, retención, catálogo, forward
purgar [--arnes X]                 aplica el TTL, o borra todo lo de un arnés
catalogo                           versión, rev, sha256, nº de modelos, último refresco
```

`arnesia hook proceso` queda **fuera** de esta familia a propósito: no es una vía de
inspección, es la instrumentación, y su contrato (stdout vacío, exit 0 siempre) es incompatible
con un subcomando que imprime.

### 10.4 · Bloque propuesto para `.go-arch-lint.yml`

**No se agrega al archivo todavía**: `go-arch-lint` falla con componentes cuyo `in:` no matchea
ningún archivo, y el módulo no existe. Se agrega en el mismo commit que crea los paquetes.

```yaml
components:
  telemetria-otlp:
    in: internal/adapters/telemetria/otlp/**
  telemetria-hooks:
    in: internal/adapters/telemetria/hooks/**
  telemetria-store:
    in: internal/adapters/telemetria/store/**
  telemetria-catalogo:
    in: internal/adapters/telemetria/catalogo/**
  telemetria-descubrimiento:
    in: internal/adapters/telemetria/descubrimiento/**

deps:
  # Decodifica el wire OTLP y emite el evento canónico. NO conoce el almacén ni el catálogo:
  # habla con ports.TelemetriaSink y nada más.
  telemetria-otlp:
    mayDependOn: [domain, ports]
  # Proyecta el payload del hook a los campos declarados (allowlist). Lo usan el daemon
  # (endpoint) y el propio binario (subcomando hook) — por eso vive acá y no en cmd.
  telemetria-hooks:
    mayDependOn: [domain, ports]
  telemetria-store:
    mayDependOn: [domain, ports]
    canUse: [sqlite]        # telemetria.db — DB PROPIA, jamás index.db (A2)
  telemetria-catalogo:
    mayDependOn: [domain, ports]
  # Escribe la ficha del daemon a os.UserConfigDir(). No conoce el dominio: habla `FichaDaemon`.
  telemetria-descubrimiento:
    mayDependOn: [domain, ports]

  cmd:
    mayDependOn: [ …lo de hoy…, telemetria-otlp, telemetria-hooks, telemetria-store,
                   telemetria-catalogo, telemetria-descubrimiento ]
```

**Lo que el bloque prohíbe por omisión (default-deny) y es intencional:** `transport-http` **no**
puede importar `telemetria-*` (el receptor entra inyectado como `http.Handler`, igual que el
broker SSE hoy) · los adaptadores de telemetría **no** se importan entre sí · `index` y
`telemetria-store` no se conocen.

### 10.5 · Tests de fitness nuevos (descritos, no implementados)

En `docs/architecture/fitness/telemetria_test.go` salvo donde se indique otra cosa.

| test | qué aserta | boundary |
|---|---|---|
| `TestAllowlistNoPersistePII` | corre los payloads de `evidencia/` + una fixture con PII simulada por el receptor; ninguna de las 5 claves (`user.email`, `user.account_uuid`, `user.account_id`, `user.id`, `organization.id`) aparece en el evento **ni como subcadena del `.db`** | ingesta-por-allowlist-declarada |
| `TestHookNoReenviaContenido` | la proyección del hook, alimentada con los 6 payloads reales, jamás emite `prompt`, `last_assistant_message`, `tool_response`, `tool_input`, `transcript_path` ni `cwd` crudo | ingesta-por-allowlist-declarada |
| `TestAllowlistEsListaNoSugerencia` | source-scan: no existe ninguna asignación que copie un `map[string]any` completo a un campo del evento | ingesta-por-allowlist-declarada |
| `TestNoAplicaNoEsCeroEnElWire` | un evento con buckets `nil` serializa a JSON **sin las claves**, no con `0`; y un `GastoCaja` no atribuible lleva `costo_micros: null` + `motivo` no vacío | no-aplica-no-es-cero |
| `TestNoAplicaSobreviveAlRollup` | un grupo del rollup donde ningún evento tenía `tok_cache_1h` queda `NULL`, no `0` | no-aplica-no-es-cero |
| `TestDetectorQueNoAplicaTraeMotivo` | los 6 detectores, con un `ContextoDeteccion` vacío: todos los `Aplica:false` tienen `Motivo != ""` | no-aplica-no-es-cero |
| `TestCifraLlevaConfianza` | reflexión sobre los DTOs de salida: todo campo de dinero convive con un campo de confianza en la misma struct | cifra-viaja-con-su-confianza |
| `TestDobleCostoSePersisteEntero` | un `api_request` con `cost_usd_micros` y catálogo disponible guarda **los dos** costos; ninguno pisa al otro | cifra-viaja-con-su-confianza |
| `TestDivergenciaDeCostoEsVisible` | si reportado y calculado difieren > 1 %, `DetalleCaja.divergencia` sale poblado, no se elige uno en silencio | cifra-viaja-con-su-confianza |
| `TestCosteoCobraElCacheWrite` | el bug de langfuse#14249 no se reproduce: el costo incluye la escritura de cache | (colocado, `domain/telemetria_costo_test.go`) |
| `TestCosteoNoSumaBucketsQueSeSolapan` | con `AritmeticaInclusiva` no se cuenta dos veces (langfuse#12306) | ídem |
| `TestCosteoNoAplanaLosTiers` | con `SobreUmbral` poblado y un prompt largo, se usa el tier (phoenix#14314) | ídem |
| `TestIntValueComoNumeroYComoString` | el mismo payload con `"intValue": 5` y `"intValue": "5"` decodifica igual | (colocado, `otlp/decodifica_test.go`) |
| `FuzzDecodificarLogs` | sembrado con los 3 payloads reales; ningún input hace panic | ídem |
| `TestPayloadGiganteSeRechazaSinLeerlo` | 8 MiB ⇒ 413 y el proceso no crece en memoria | ídem |
| `TestReceptorNoBloqueaAlEmisor` | con la cola llena, `ServeHTTP` responde en < 5 ms con `partialSuccess` | peso-del-binario-es-presupuesto (§ presupuestos) |
| `TestTokenDeIngestaNoAbreLaAPI` | 200 en `/v1/logs`, **401 en `/api/sessions`** con el mismo token | superficie-local-confinada |
| `TestOTLPBajoLosTresGates` | `/v1/logs` con Host ajeno ⇒ 403; sin token ⇒ 401 | superficie-local-confinada |
| `TestFichaSoloTrasEscuchar` | la ficha no existe antes de que el listener acepte; existe después; se borra en el shutdown | telemetria-de-nacimiento |
| `TestFichaPermisos0600` | el archivo es `0600` y su dir `0700` | telemetria-de-nacimiento |
| `TestTelemetriaDBNoEsElIndice` | source-scan: `telemetria/store` nunca resuelve a `index.db`; y un wipe del índice no toca `telemetria.db` | indice-desechable-jsonl-es-verdad |
| `TestMigracionesSonAditivas` | la lista de migraciones no contiene `DROP`, `RENAME` ni `ALTER … DROP COLUMN` | (colocado, `store/migracion_test.go`) |
| `TestGeneracionArchivaNoBorra` | con mismatch de generación, el `.db` viejo queda en disco renombrado y `salud` lo reporta | ídem |
| `TestJoinPorSesionYTurno` | un `api_request` y un `Stop` del mismo `(sesion_id, prompt_id)` producen **un** `TurnoUnido` con dinero y proceso | (colocado, `usecase/telemetria_service_test.go`) |
| `TestConciliacionCuentaLosNoLlegados` | 7 turnos esperados, 5 medidos ⇒ `Cobertura.NoLlegaron == 2` | ídem |
| `TestSinDatoNoSumaAlTotal` | eventos `atribucion=sin-dato` no entran en `costo_*` del resumen y sí en `Cobertura.SinDato` | ídem |
| `TestRelojHaciaAtrasNoRompeLaVentana` | un `ts_emisor` 3 días en el pasado marca `reloj_sospechoso` y el rollup usa `ts_recibido` | ídem |
| `TestForwardApagadoPorDefault` | sin flag, cero conexiones salientes (fake dialer) | telemetria-de-nacimiento |
| `TestForwardNoReenviaCrudo` | con forward encendido, lo que sale es el evento proyectado, no el body OTLP | ingesta-por-allowlist-declarada |
| `TestPresupuestoDeBinario` | compila el daemon y compara contra el baseline del release anterior: delta ≤ 1,5 MB. `t.Skip` bajo `-short`; corre en CI | peso-del-binario-es-presupuesto |
| `TestSpawnInyectaTelemetria` | `SpawnEnv(opts)` contiene las 8 vars, con `http/json` y el endpoint loopback | telemetria-de-nacimiento |
| `TestSpawnNoFiltraElTokenDeAPI` | el env del spawn lleva `token_ingesta`, **nunca** `ARNESIA_AUTH_TOKEN` | superficie-local-confinada |
| `TestEscenarioSeDerivaDeLaSenal` | tres lotes sintéticos (con corrida nuestra · `api_request` sin corrida · solo hook) producen `s1` / `s2-instrumentado` / `s2-degradado`. Un arnés **no puede declarar su propio escenario** | telemetria-de-nacimiento |
| `TestS2DegradadoApagaLosDetectoresDeDinero` | con solo eventos de hook, B4/B6/B3 salen en `no_aplican` **con motivo**, y el resumen trae `costo_*: null` — nunca 0 | no-aplica-no-es-cero |
| `TestS2InstrumentadoTieneDineroYNoTieneSplit` | con `api_request` y sin `result`, B4/B6/B3 aplican y **B1 no**, con el motivo del split (no con el motivo «corrió fuera») | no-aplica-no-es-cero |
| `TestToolResultBytesSePersiste` | un `tool_result` guarda `tool_input_bytes`/`tool_result_bytes` y **ningún** campo de contenido | ingesta-por-allowlist-declarada |

FE (`vitest --project=storybook`, story = test): `capa-mejora.stories.tsx` con los 7 estados
honestos, cada uno con `play()` que aserta el texto del estado — en particular que el estado 1
**no** renderiza «USD 0,00».

---

## 11 · Presupuestos no funcionales

Todos medibles, todos con su verificación.

| presupuesto | valor | de dónde sale | cómo se verifica |
|---|---|---|---|
| **peso del binario** | **+1,5 MB** máximo sobre el release anterior | decodificador medido +0,49 MB (V5) + catálogo 182 KB + margen | `TestPresupuestoDeBinario` en CI |
| **peso absoluto** | daemon ≤ **25 MB** sin `-s -w` | hoy 22,71 MB | ídem |
| **latencia del handler OTLP** | p99 ≤ **5 ms** (decode + encolar; **cero I/O de disco**) | el disco lo hace el writer, asíncrono | benchmark `BenchmarkReceptorLogs` + `TestReceptorNoBloqueaAlEmisor` |
| **payload máximo** | **4 MiB** | el mayor observado es 24 KB (`evidencia/logs-run1.json`); 170× de margen | `TestPayloadGiganteSeRechazaSinLeerlo` |
| **cola del receptor** | **4096** eventos; el exceso se descarta y se **cuenta** | absorbe ~7 min de una sesión intensa a 10 ev/s | `TestReceptorNoBloqueaAlEmisor` |
| **lote de escritura** | 256 eventos o 250 ms | 419 798 filas/s medidos: sobra | benchmark |
| **retención default** | **90 días** de `evento`; **24 meses** de `rollup_hora` | 90 días ≈ 3 % del umbral de disco (~10 M filas / 940 MB) | `TestPurgaRespetaTTL` |
| **tamaño del `.db`** | aviso en la UI a partir de **500 MB** | umbral de incomodidad medido ~940 MB | `salud.tamano_bytes` |
| **cardinalidad del rollup** | ≤ **50 000** filas/mes. Al superarlo, `caja_id` colapsa a `(otros)` y se marca `cardinalidad_colapsada` | la clave tiene 8 dimensiones; sin tope, un arnés con cajas generadas la explota | `TestRollupColapsaCardinalidad` |
| **latencia del tablero** | p95 ≤ **50 ms** sobre el rollup | medido 1,7 ms para `sum group by` sobre rollup | benchmark |
| **hook** | ≤ **250 ms** de pared, exit 0 siempre, stdout vacío | tope duro del contrato §7.1 | `TestHookNoTardaNiFalla` |
| **catálogo embebido** | ≤ **256 KB** | 182 KB medidos filtrado | test de tamaño del `go:embed` |

---

## 12 · Contradicciones encontradas en la documentación existente

Se listan acá, no se resuelven en silencio.

| # | dónde | qué dice | qué corresponde |
|---|---|---|---|
| 1 | `arquitectura-telemetria.md` L2 y §L1 | dibuja `adapters/telemetria/streamjson/ (parser por runtime)` | **no se construye** (A5): duplicaría el decodificador de stream-json y violaría `adaptadores-de-agente-intercambiables`. El `result` lo sigue decodificando `agent/claudecode` |
| 2 | `arquitectura-telemetria.md` §Parte 3 y `decisiones.md` D12.2 | «la llave del join es `session.id` + los resource attributes» | **incompleto**: la llave es **`(session_id, prompt_id)`** (ANEXO H1). Con solo `session.id` el join es a nivel sesión y no dice «el 60 % se va en la caja Y» |
| 3 | `decisiones.md` D16.2, grupo «procedencia», campo `origen` | usa `origen` y `procedencia`, **las dos tomadas en L0** con otro significado | renombrado a `emisor` y `atribucion` (A4, §2.2). `via` tampoco estaba libre |
| 4 | `arquitectura-telemetria.md` L2 «L2 INGESTA … (handler net/http + pdata)» y F1 | `pdata` como decodificador | **refutado y superseded** por V5/D14.2 (+10,79 MB medidos). El texto de la Parte 2 quedó stale respecto de su propio §V5 |
| 5 | `docs/architecture/INDEX.md`, fila `telemetria-de-nacimiento` | «🌱 vivo · 1.0 · 2 checks» | el boundary va por **v2.2 con 7 checks** (v2.3 con 9 tras este paquete). La fila está stale — se corrige en este mismo entregable |
| 6 | `docs/architecture/INDEX.md`, fila `indice-desechable-jsonl-es-verdad` | «(in-memory, sin persistencia hoy) · 1.2» | el boundary va por **v1.3** y el store **es SQLite real** desde la fase 5. Fila stale (no se corrige acá: es de otro paquete, queda anotado) |
| 7 | `decisiones.md` D8 / `arch_test.go:267` | `TestNoJSONLSchemaParsing` sigue `t.Skip`eado | residuo vivo declarado por D14.4. **Este paquete lo toca de refilón** (no agrega parsing de JSONL) pero no lo cierra: sigue abierto |
| 8 | `arquitectura-telemetria.md` §L7 y `INDEX.md` del paquete | «capa Tokens del Mapa» | renombrada a **capa «Mejora»** por D16.3/D17.1. Quedan referencias al nombre viejo en la Parte 2 |
| 9 | `decisiones.md` D9.3 / F4 | `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` | **invertido** por D14.2: `http/json`. El texto de D9.3 conserva el valor viejo dentro de un bloque que el gate declaró FIRMADO |
| 10 | `decisiones.md` D12.1, `arquitectura-telemetria.md` §Alcance, `INDEX.md` del paquete | «S2 = no controlamos el spawn ⇒ la instrumentación tiene que viajar dentro del arnés (hook)» | **premisa falsa** (ANEXO H9): el bloque `env` de un `settings.json` enciende OTel. S2 se parte en `s2-instrumentado` y `s2-degradado` (A19, §7.0). El hook sigue haciendo falta —para gate, rotación y corrida— pero **ya no es la única señal de S2** |
| 11 | `decisiones.md` D16.1, tabla de detectores, columna S2 | «B1 ❌ no aplica en S2» y «B2 ⚠️» | el veredicto de B1 no cambia pero **el motivo sí**: no es «corrió fuera de ArnesIA», es «falta el `result` del stream-json». Escrito así, el detector se enciende solo el día que el split llegue por OTel |
| 12 | `arquitectura-telemetria.md` §Parte 3, tabla «las tres señales» | «proceso ⟵ hooks del propio arnés» como única fuente | **parcialmente superado** (ANEXO H8): `tool_decision` y `tool_result` traen decisión de permiso, éxito y duración **por herramienta** vía OTel. Los hooks siguen siendo la única fuente de gate humano, rotación y corrida |

---

## 13 · Lo que este plan deja `ABIERTO`

### 13.1 · Cerrados el 2026-07-26 (tercera tanda en vivo, ANEXO §H10)

Los tres se corrieron **con control positivo en la misma corrida**. Ninguno queda como supuesto.

| # | qué | resultado | qué cambió en el plan |
|---|---|---|---|
| ~~1~~ | Comportamiento de CC cuando el **comando del hook no existe** | ✅ **fail-open confirmado**: `exit 0`, `is_error:false`, stderr vacío, duración normal; el hook válido de control sí corrió | §7.1: **la propiedad que A7 necesita la garantiza el runtime**, no hay que construirla. Aclarado que cubre que el binario *falte*, no que *falle*: el contrato «exit 0 siempre» sigue siendo obligación nuestra |
| ~~9~~ | ¿Un **plugin** puede aportar bloque `env`? | ❌ **NO** — 0 payloads con el plugin instalado, 2 en el control positivo | §7.5: **A20 no desaparece**; pasa de «pendiente de verificar» a **confirmada necesaria**. No existe auto-instrumentación al instalar |
| ~~10~~ | ¿CC expande `${VAR}` dentro del bloque `env`? | ❌ **NO** — llegan literales (`arnesia.expand = ${MARCA_PROPIA}`). *Matiz: en los comandos de hook la expansión sí funciona* | **A22 nueva** (§7.3): el token no puede llegar por indirección ⇒ `/v1/*` acepta sin token bajo Host loopback, con escotilla `--telemetria-ingesta-token-obligatorio` |

### 13.2 · Sigue abierto

| # | qué | dueño | bloquea |
|---|---|---|---|
| 1 | **A20 — dónde vive el bloque `env`** (opción A repo del arnés · opción B proyecto del usuario con consentimiento) | **operador** | el diseño de `s2-instrumentado` y el mecanismo de obligación. **Ya no hay verificación que la evite** (H10.1) |
| 2 | Nombres reales de `OTEL_LOGS_EXPORTER` / `OTEL_LOGS_EXPORT_INTERVAL` (§10.2) | verificación en vivo | nada — hay default |
| 3 | Determinismo de `plugin_id_hash` **entre máquinas** (V7.2) | verificación en 2ª máquina | nada: el diseño aprende local |
| 4 | Comportamiento con subagentes / `/compact` (V7.4) | corrida dirigida | B7/B8, fuera del MVP. *(Los eventos `tool_*` de V7.3 ya se cerraron en H8.)* |
| 5 | **D9.9** — ¿parser propio o shell-out a `ccusage`? | operador | nada: posterior al MVP |
| 6 | **G4** — cert de firma de código Windows/macOS | operador | el canal Windows/macOS, no el módulo |
| 7 | `tauri#11992` (notarización con `externalBin`) | runner macOS | el canal macOS |

---

## 14 · Nota de método: un negativo sin control positivo no es un resultado

Los tres primeros intentos de la prueba H9 dieron **negativo y el negativo era falso**: un
receptor de una prueba anterior seguía ocupando el puerto y se quedaba con el tráfico. Se
detectó con un control antes de escribir ninguna conclusión.

**La regla ya se pagó y ya rindió.** Las tres verificaciones de H10 —las que cerraron §13 ítems
1, 9 y 10— se corrieron con control positivo en la misma corrida, y **dos de ellas son negativos**
(el plugin no aporta `env`, `${VAR}` no se expande). Sin el control, esos dos negativos serían
indistinguibles de un receptor caído, y el plan habría quedado apoyado en dos «no funciona» que
podrían haber sido «no probé bien».

**Regla que hereda todo el módulo, y en particular la matriz de
[`escenarios.md`](escenarios.md):** cualquier escenario cuyo criterio de éxito sea *«no llegó
nada»* — telemetría apagada, hook fail-open, forward apagado, WSL2 inalcanzable, payload
rechazado — **no se valida solo con la ausencia**. Necesita, en la misma corrida y contra el
mismo receptor:

1. un **control positivo**: un evento con marcador único que **sí** tiene que llegar;
2. verificación de que el receptor de la prueba es el que está escuchando (`pgrep` del puerto,
   o mejor: el receptor de test bindea `:0` y el puerto se le pasa al sujeto);
3. **marcadores distintos por variante** — dos variantes con el mismo marcador no se distinguen
   si una filtra a la otra.

En los tests Go esto se materializa así, y es requisito de revisión: **ningún test de este
módulo aserta solo `len(recibidos) == 0`.** Aserta `len(conMarcadorA) == 0 && len(conMarcadorB) == 1`
sobre un `httptest.Server` de puerto efímero. Un test que solo mira el vacío puede estar
midiendo su propia rotura.
