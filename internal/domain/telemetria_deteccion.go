package domain

import "time"

// telemetria_deteccion.go es el CONTRATO de los detectores de puntos de mejora
// (arquitectura-modulo.md §2.4). Las seis implementaciones del MVP y `DetectoresMVP()`
// llegan en su propio ticket; acá vive lo que el resto del módulo necesita para hablar de
// ellos —incluidas las vistas de lectura, que devuelven `PuntoDeMejora`.
//
// La regla que gobierna el archivo entero: **`Aplica()` se consulta SIEMPRE antes que
// `Evaluar()`**, y **`Motivo` es obligatorio cuando `Aplica` es false**. Un detector
// apagado sin razón es un gap escondido, y un detector que devuelve 0 porque no pudo
// correr es una mentira (boundary no-aplica-no-es-cero).

// DetectorID identifica un detector de forma estable. El id viaja al wire y a la UI: es lo
// que hace que una marca de fuga NOMBRE al detector en vez de ser un ⚠ genérico.
type DetectorID string

const (
	DetB4 DetectorID = "b4-gasto-por-arnes-empresa-puesto"
	DetP1 DetectorID = "p1-caja-que-consume-y-se-rechaza"
	DetB2 DetectorID = "b2-costo-de-la-rotacion"
	DetB6 DetectorID = "b6-sesion-abandonada"
	DetB3 DetectorID = "b3-cambio-de-modelo-invalida-cache"
	DetB1 DetectorID = "b1-rewarm-por-ttl"
)

// ContextoDeteccion es lo que un detector necesita para decidir si PUEDE correr. Se arma
// del dato real de la ventana, jamás se asume ni se configura: si se declarara, un arnés
// podría decir que está mejor medido de lo que está.
type ContextoDeteccion struct {
	Runtime   string
	Escenario Escenario
	// TieneCosto: llegó ALGÚN evento con dinero en la ventana. En s2 esto depende de si el
	// arnés lleva el bloque `env` (ANEXO H9), no del escenario — por eso es un hecho de la
	// ventana y no un derivado del modo.
	TieneCosto bool
	// TieneSplitTTL: llegó el `result` del stream-json, que es donde vive el split
	// ephemeral_5m/1h. Es la regla de B1 — así, el día que el split llegue por OTel, el
	// detector se enciende SOLO, sin tocar código.
	TieneSplitTTL       bool
	CatalogoDisponible  bool
	TieneSenalProceso   bool // llegaron eventos de hook/daemon (P1 los exige)
	TieneGateHumano     bool // hubo eventos de gate del daemon (P1 completo vs. parcial)
	TieneEventoRotacion bool // el daemon registró rotaciones (B2)
	// ModelosDistintos cuenta los modelos no vacíos de la ventana (B3 exige ≥2).
	ModelosDistintos int
}

// Aplicabilidad es la respuesta honesta de un detector a «¿podés correr acá?».
type Aplicabilidad struct {
	Aplica bool   `json:"aplica"`
	Motivo string `json:"motivo,omitempty"`
	// Parcial marca que el detector corre pero no ve todo (el caso de P1 en
	// s2-instrumentado: ve reintentos y fracasos de herramienta, no rechazos de gate).
	// Viaja como matiz explícito con su motivo, jamás como un ✅ liso.
	Parcial bool `json:"parcial,omitempty"`
}

// PuntoDeMejora es la unidad del entregable: un número, su contrafactual, su umbral
// citado, su sesgo declarado EN CONTRA y UN fix (regla A4). Sin las cinco cosas no se
// muestra — una tarjeta sin contrafactual es un reproche, no una recomendación.
type PuntoDeMejora struct {
	Detector     DetectorID `json:"detector"`
	ScoreVersion int        `json:"score_version"` // A7: cambiar la fórmula obliga a bumpearlo
	Titulo       string     `json:"titulo"`
	Lede         string     `json:"lede,omitempty"`
	CajaID       string     `json:"caja_id,omitempty"`
	GastoMicros  int64      `json:"gasto_micros"`
	// ParteDelTotal va en 0..1.
	ParteDelTotal float64 `json:"parte_del_total"`
	// ContrafactualMicros es qué habría costado el mundo alternativo, con las MISMAS
	// corridas (A2). No es «gastaste X»: es «con el cambio habrías gastado Y».
	ContrafactualMicros int64 `json:"contrafactual_micros"`
	DiferenciaMicros    int64 `json:"diferencia_micros"`
	// Umbral es la desigualdad algebraica CITADA (A1), no un número mágico.
	Umbral string `json:"umbral"`
	// Sesgo se declara y va EN CONTRA de la recomendación (A3): si el sesgo favoreciera
	// al fix que proponemos, la recomendación se estaría auto-justificando.
	Sesgo string `json:"sesgo"`
	// DireccionSesgo es "subestima" | "sobreestima". Nunca vacío ni neutro (RF-252).
	DireccionSesgo  string    `json:"direccion_sesgo"`
	Fix             string    `json:"fix"` // UNO, concreto
	Confianza       Confianza `json:"confianza"`
	CorridasUsadas  int       `json:"corridas_usadas"`
	CorridasTotales int       `json:"corridas_totales"`
	Grave           bool      `json:"grave,omitempty"`
}

// Ventana es el corte de datos sobre el que corre un detector: los turnos unidos
// (dinero×proceso) de un arnés en un rango, YA costeados. Un detector no consulta el
// almacén: recibe su ventana y calcula.
type Ventana struct {
	Desde       time.Time
	Hasta       time.Time
	ArnesID     string
	Turnos      []TurnoUnido
	TotalMicros int64
	Contexto    ContextoDeteccion
}

// Detector es el contrato común de los seis del MVP (D16.1).
type Detector interface {
	ID() DetectorID
	Nombre() string
	Aplica(c ContextoDeteccion) Aplicabilidad
	Evaluar(v Ventana) []PuntoDeMejora
}
