package domain

import (
	"errors"
	"time"
)

// telemetria.go es el evento canónico del módulo de telemetría embebida (paquete
// 2026-07-24-telemetria-embebida-otel, arquitectura-modulo.md §2.1). Los CUATRO emisores
// —receptor OTLP, adaptador de agente (stream-json), hook del arnés y el propio daemon—
// producen este tipo y nada más; el almacén tiene una sola puerta de escritura.
//
// Lo que este archivo NO tiene, y no va a tener: identidad de cuenta ni contenido de
// conversación. Ver docs/architecture/boundaries/ingesta-por-allowlist-declarada.md.

// Emisor dice POR QUÉ CANAL entró el evento. No es `origen` (facet L0 estandar|del-puesto
// del provisioner) ni `canal` (beta|estable|propuesto|deprecado) ni `via` (cómo cruzó una
// coincidencia de marketplace): los tres están tomados con otro significado — ver
// mockups/INDEX.md regla 4 y arquitectura-modulo.md §2.2 (decisión A4).
type Emisor string

const (
	// EmisorOTLP — receptor OTLP embebido: dinero + latencia.
	EmisorOTLP Emisor = "otlp"
	// EmisorStreamJSON — frame `result` del subproceso: split 5m/1h del cache.
	EmisorStreamJSON Emisor = "streamjson"
	// EmisorHook — hook del arnés: proceso. JAMÁS dinero (verificado, ANEXO H2).
	EmisorHook Emisor = "hook"
	// EmisorDaemon — decisión propia: rotación, corrida, gate.
	EmisorDaemon Emisor = "daemon"
)

// Confianza dice CÓMO se atribuyó el evento a una unidad de trabajo, en orden de
// preferencia decreciente (ANEXO H5). Ningún número se muestra sin ella
// (docs/architecture/boundaries/cifra-viaja-con-su-confianza.md).
type Confianza string

const (
	// ConfianzaExacta — los resource attributes `arnesia.*` que inyectamos al spawn (S1).
	ConfianzaExacta Confianza = "exacta"
	// ConfianzaPorHash — `plugin_id_hash` resolvió en la tabla local. Da el arnés, no la caja.
	ConfianzaPorHash Confianza = "por-hash"
	// ConfianzaPorProceso — el `cwd` del hook mapeó a una instalación conocida del Portafolio.
	ConfianzaPorProceso Confianza = "por-proceso"
	// ConfianzaSinDato — nada resolvió. NUNCA se suma al total (decisión A15).
	ConfianzaSinDato Confianza = "sin-dato"
)

// rankConfianza ordena las cuatro de mejor a peor. Se usa para que la confianza de un
// agregado sea la MÍNIMA de sus partes, nunca la máxima ni la moda: un total que mezcla
// una atribución exacta con una adivinada vale lo que la adivinada.
var rankConfianza = map[Confianza]int{
	ConfianzaExacta:     0,
	ConfianzaPorHash:    1,
	ConfianzaPorProceso: 2,
	ConfianzaSinDato:    3,
}

// PeorConfianza devuelve la menos confiable de las dos. Una confianza desconocida se trata
// como sin-dato: no se le concede el beneficio de la duda a un valor que no reconocemos.
func PeorConfianza(a, b Confianza) Confianza {
	ra, oka := rankConfianza[a]
	if !oka {
		ra = rankConfianza[ConfianzaSinDato]
		a = ConfianzaSinDato
	}
	rb, okb := rankConfianza[b]
	if !okb {
		rb = rankConfianza[ConfianzaSinDato]
		b = ConfianzaSinDato
	}
	if ra >= rb {
		return a
	}
	return b
}

// Aritmetica es propiedad del ADAPTADOR (D9.5/D-7): sin este flag el error de conteo es
// ~100 % en un sentido o ~28 % en el otro.
type Aritmetica string

const (
	// AritmeticaDisjunta — Anthropic: los 4 buckets no se solapan, se suman tal cual.
	AritmeticaDisjunta Aritmetica = "disjoint"
	// AritmeticaInclusiva — OTel semconv / OpenAI: `cache_lectura ⊂ entrada`, hay que restar.
	AritmeticaInclusiva Aritmetica = "inclusive"
)

// Acumulacion es la regla del adaptador para no contar dos veces (D9.5/D-6).
type Acumulacion string

const (
	// AcumulacionPorRequest — cada evento es un delta propio.
	AcumulacionPorRequest Acumulacion = "por-request"
	// AcumulacionAcumulada — el último evento contiene a los previos.
	AcumulacionAcumulada Acumulacion = "acumulada"
	// AcumulacionUltima — el runtime re-emite el total.
	AcumulacionUltima Acumulacion = "ultima"
)

// TipoEvento es qué ocurrió. Las filas de dinero y las de proceso viven en la misma tabla
// a propósito: el join es un self-join por (sesion_id, turno_id), no dos tablas.
type TipoEvento string

const (
	EventoAPIRequest  TipoEvento = "api_request"  // OTLP: dinero.
	EventoTurnoInicio TipoEvento = "turno_inicio" // hook UserPromptSubmit.
	EventoHerramienta TipoEvento = "herramienta"  // hook PostToolUse · OTel tool_decision/tool_result.
	EventoTurnoFin    TipoEvento = "turno_fin"    // hook Stop.
	EventoSesionFin   TipoEvento = "sesion_fin"   // hook SessionEnd.
	EventoRotacion    TipoEvento = "rotacion"     // daemon: rotación de contexto (B2).
	EventoCorrida     TipoEvento = "corrida"      // daemon: run T3 de una caja.
	EventoGate        TipoEvento = "gate"         // daemon: veredicto del gate (P1)
	// EventoMetrica — punto del canal SECUNDARIO (`/v1/metrics`).
	//
	// 🔴 Existe como tipo PROPIO por una sola razón, y es la más importante del módulo:
	// `claude_code.cost.usage` y `api_request.cost_usd_micros` son **el mismo gasto de la
	// misma llamada**. Si los dos entraran como `api_request`, `SUM(costo)` los sumaría y el
	// tablero reportaría **el doble de lo que el operador gastó** — y el dedupe no los cruza,
	// porque el punto de métrica no trae `prompt.id` y su `ts_emisor` difiere en ~490 ms.
	//
	// La regla, y no se negocia: **una unidad de gasto se cuenta UNA sola vez.** D14.1 ya
	// había declarado que el canal primario es `/v1/logs`; esto lo IMPLEMENTA. El dato del
	// canal secundario **se guarda** —no se tira— y es visible en el detalle y en la salud,
	// pero **ninguna agregación de dinero ni de tokens lo toca**.
	EventoMetrica TipoEvento = "metrica"
)

// Resultado es el desenlace de un evento de proceso; vacío en los de dinero.
type Resultado string

const (
	ResultadoOK        Resultado = "ok"
	ResultadoRechazado Resultado = "rechazado"
	ResultadoReintento Resultado = "reintento"
	ResultadoCancelado Resultado = "cancelado"
)

// Escenario distingue QUÉ SEÑAL EXISTE, no quién spawneó (ANEXO H9 corrigió la premisa:
// fuera de ArnesIA también puede haber señal completa, si el arnés lleva bloque `env`).
// Son tres niveles de dato distintos y la UI los distingue (decisión A19).
//
// El escenario se DERIVA de lo que llegó, jamás lo declara el emisor: un arnés no puede
// mentir sobre su propio nivel de instrumentación.
type Escenario string

const (
	// EscenarioS1 — spawn nuestro: env inyectado + stream-json + hook. Señal completa.
	EscenarioS1 Escenario = "s1"
	// EscenarioS2Instrumentado — fuera de ArnesIA, PERO el arnés lleva el bloque `env` en
	// sus settings ⇒ llega OTel con dinero. Sin stream-json (no somos el proceso padre).
	EscenarioS2Instrumentado Escenario = "s2-instrumentado"
	// EscenarioS2Degradado — fuera de ArnesIA y sin bloque `env`: solo el hook ⇒ proceso,
	// jamás dinero (ANEXO H2). Los detectores de dinero se apagan CON MOTIVO, nunca en 0.
	EscenarioS2Degradado Escenario = "s2-degradado"
)

// Tokens es el superset multi-proveedor. TODOS los campos son punteros a propósito
// (D-4 · boundary no-aplica-no-es-cero): nil = «este runtime no tiene el concepto», que NO
// es lo mismo que 0. Un 0 donde el concepto no existe es una mentira.
type Tokens struct {
	Entrada          *int64 `json:"entrada,omitempty"`
	Salida           *int64 `json:"salida,omitempty"`
	CacheLectura     *int64 `json:"cache_lectura,omitempty"`
	CacheEscritura5m *int64 `json:"cache_escritura_5m,omitempty"`
	CacheEscritura1h *int64 `json:"cache_escritura_1h,omitempty"`
	Razonamiento     *int64 `json:"razonamiento,omitempty"`
	// CacheEscrituraSinTier son tokens de escritura de cache que llegaron **sin decir a qué
	// vencimiento se escribieron**. Es el caso del canal OTLP, que manda un solo
	// `cache_creation_tokens` agregado.
	//
	// 🔴 Tiene bucket PROPIO y no se pliega al de 5 minutos, y esa es toda la razón de que
	// exista: la tarifa de 1 h cuesta 1,6× la de 5 min, así que asumir el tramo barato
	// SUBESTIMA. Medido en la corrida real: asumir 5 min da 12 280 micros contra los 18 473
	// que el runtime reportó — un 33 % por debajo, que es exactamente el bug `phoenix#14314`
	// que este módulo existe para no reproducir.
	//
	// Un tier desconocido **no se asume: se declara**. `CalcularCosto` lo deja sin cobrar y
	// lo nombra en `SinTarifa`, y el costo sale marcado incompleto.
	CacheEscrituraSinTier *int64 `json:"cache_escritura_sin_tier,omitempty"`
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

// EventoTelemetria es el ÚNICO evento canónico (D9.2).
//
// Lo que NO tiene, y no va a tener: identidad de cuenta (user.email, user.account_uuid,
// user.account_id, organization.id, user.id) ni contenido (prompt, respuesta,
// tool_response). TestEventoNoTieneCamposDeIdentidad lo enforcea por reflexión — no por
// buena voluntad del que agregue el próximo campo.
type EventoTelemetria struct {
	LlaveJoin

	Emisor           Emisor `json:"emisor"`
	Runtime          string `json:"runtime"` // "claude-code" | "codex" | …
	RuntimeVersion   string `json:"runtime_version,omitempty"`
	AdaptadorVersion string `json:"adaptador_version"` // versiona el mapeo (D7.5)

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
	// Reportado es lo que dijo el runtime («Estimated cost», ANEXO H4 — nunca facturación);
	// Calculado es lo que dice NUESTRO catálogo. Si divergen, el test de paridad está
	// corriendo en producción y gratis. nil ≠ 0 en los dos.
	CostoReportadoMicros *int64 `json:"costo_reportado_micros,omitempty"`
	CostoCalculadoMicros *int64 `json:"costo_calculado_micros,omitempty"`
	CostoCompleto        *bool  `json:"costo_completo,omitempty"` // false = algún bucket sin tarifa
	CatalogoVersion      string `json:"catalogo_version,omitempty"`

	// Escenario es columna de primera clase (A19), no un derivado de consulta: sin él
	// `s2-instrumentado` y `s2-degradado` se leerían como el mismo «S2» y el usuario no
	// entendería por qué uno tiene plata y el otro no.
	Escenario   Escenario  `json:"escenario"`
	TipoEvento  TipoEvento `json:"tipo_evento"`
	Resultado   Resultado  `json:"resultado,omitempty"`
	Gate        string     `json:"gate,omitempty"`
	Motivo      string     `json:"motivo,omitempty"`
	Herramienta string     `json:"herramienta,omitempty"`
	Decision    string     `json:"decision,omitempty"` // accept|reject… de tool_decision (ANEXO H8)

	// A21 — los DOS tamaños se persisten desde el día 1 aunque B11 esté fuera del MVP.
	// Son BYTES, no contenido (verificado), y un dato que no se guarda hoy no se recupera
	// mañana. Punteros: un evento que no es tool_result no tiene el concepto.
	ToolInputBytes  *int64 `json:"tool_input_bytes,omitempty"`
	ToolResultBytes *int64 `json:"tool_result_bytes,omitempty"`

	PluginIDHash string    `json:"plugin_id_hash,omitempty"`
	CWDHuella    string    `json:"cwd_huella,omitempty"` // NUNCA la ruta cruda (A14)
	Atribucion   Confianza `json:"atribucion"`
}

// UsoDelTurno es lo que el frame `result` del stream-json YA trae y hoy se tira
// (arquitectura-modulo.md §4.2, decisión A5). No existe `adapters/telemetria/streamjson/`:
// el `result` lo sigue decodificando `agent/claudecode`, que es su dueño único — un segundo
// parser del mismo frame driftaría por separado.
//
// El split `Ephemeral5m`/`Ephemeral1h` es LA señal que habilita el detector B1 (re-warm por
// TTL) y no viaja por OTel: solo se ve cuando ArnesIA es el proceso padre.
type UsoDelTurno struct {
	Entrada      int64 `json:"entrada"`
	Salida       int64 `json:"salida"`
	CacheLectura int64 `json:"cache_lectura"`
	// CacheEscritura es el total; Ephemeral5m + Ephemeral1h es su split por TTL. Un 0 acá
	// es un DATO (el runtime lo dijo), no una ausencia — por eso no son punteros.
	CacheEscritura int64 `json:"cache_escritura"`
	Ephemeral5m    int64 `json:"ephemeral_5m"`
	Ephemeral1h    int64 `json:"ephemeral_1h"`

	ServiceTier string `json:"service_tier,omitempty"`
	Speed       string `json:"speed,omitempty"`

	// CostoUSD es lo que el runtime reportó para el turno (`total_cost_usd` /
	// `modelUsage[].costUSD`). Sigue siendo una ESTIMACIÓN suya, no facturación.
	CostoUSD       float64 `json:"costo_usd"`
	Modelo         string  `json:"modelo,omitempty"`
	ModeloCanonico string  `json:"modelo_canonico,omitempty"`
	Proveedor      string  `json:"proveedor,omitempty"`
	ContextWindow  int     `json:"context_window,omitempty"`
}

// PerfilRuntime lo declara el ADAPTADOR, no el agregador (decisión A18 · D-6/D-7). No hay
// una tabla central que un agregador consulte: cada adaptador estampa su perfil en cada
// evento y el dominio solo tipa el valor.
type PerfilRuntime struct {
	Runtime          string
	AdaptadorVersion string
	Aritmetica       Aritmetica
	Acumulacion      Acumulacion
	// Soporta lista los detectores que este runtime PUEDE sostener; el resto se reporta
	// «no aplica» con motivo (D-5), jamás en cero.
	Soporta []DetectorID
}

// FichaDaemon es el archivo de descubrimiento que el daemon publica para que un hook que
// no conoce nuestras convenciones lo encuentre en las tres plataformas (decisión A8).
// Vive en os.UserConfigDir()/arnesia/daemon.json, `0600`, y se escribe DESPUÉS de que el
// listener acepta: una ficha que nombra un puerto muerto es una mentira que el hook cobra
// en timeouts.
type FichaDaemon struct {
	Version     int    `json:"version"`
	PID         int    `json:"pid"`
	Endpoint    string `json:"endpoint"`
	RutaProceso string `json:"ruta_proceso"`
	OTLP        struct {
		Logs    string `json:"logs"`
		Metrics string `json:"metrics"`
	} `json:"otlp"`
	// TokenIngesta es el token ACOTADO (A6), jamás el de la API: filtrarlo concede
	// «escribime telemetría», nunca «conducí un agente con acceso al filesystem».
	TokenIngesta string    `json:"token_ingesta"`
	Binario      string    `json:"binario"`
	Desde        time.Time `json:"desde"`
}

var (
	// ErrPayloadInvalido — el cuerpo no es el wire format esperado. Se rechaza el LOTE
	// entero, jamás se acepta a medias en silencio.
	ErrPayloadInvalido = errors.New("telemetria: payload inválido")
	// ErrEventoSinSesion — un evento sin sesion_id no es atribuible ni deduplicable.
	ErrEventoSinSesion = errors.New("telemetria: evento sin sesion_id")
	// ErrCampoProhibido — un campo de la denylist dura (PII/contenido) llegó al
	// constructor. Es un BUG del adaptador, no un dato del usuario: se loguea con el
	// nombre del campo y nunca con su valor.
	ErrCampoProhibido = errors.New("telemetria: campo prohibido en el evento canónico")
)

// runtimesSoportados son los runtimes cuya señal sabemos leer HOY. La lista es corta a
// propósito: `investigacion-runtimes.md` relevó 13 y el canal universal —el stream-json por
// turno— todavía no está construido para los otros.
//
// Un runtime fuera de la lista **no se mide en cero**: se DICE que no lo medimos (D26.4,
// estado «runtime no soportado»). Un cero se leería como «este arnés no gasta».
var runtimesSoportados = map[string]bool{"claude-code": true}

// RuntimeSoportado dice si sabemos medir ese runtime. El vacío es false: sin saber qué runtime
// es, no podemos afirmar que lo medimos.
func RuntimeSoportado(runtime string) bool { return runtimesSoportados[runtime] }
