// Package otlp es el ÚNICO lugar del árbol que conoce el wire format OTLP (mismo patrón que
// `agent/claudecode` con stream-json). Decodifica el subconjunto que Claude Code emite —con
// la stdlib, +0,49 MB de binario medidos, contra +10,79 MB que costaría `collector/pdata`—
// y lo mapea al evento canónico.
//
// Lo que este paquete NO conoce: el almacén, el catálogo de precios ni el HTTP de la API.
// Habla `ports.TelemetriaSink` y nada más (.go-arch-lint.yml: `telemetria-otlp` solo puede
// depender de `domain` y `ports`).
package otlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// valorAtributo es UN valor de `AnyValue` del wire OTLP. Solo los cuatro tipos escalares que
// Claude Code emite; `arrayValue`/`kvlistValue`/`bytesValue` se IGNORAN — no los usa, y
// decodificarlos sería superficie de ataque gratis.
type valorAtributo struct {
	StringValue *string      `json:"stringValue"`
	IntValue    *json.Number `json:"intValue"`
	DoubleValue *float64     `json:"doubleValue"`
	BoolValue   *bool        `json:"boolValue"`
}

type atributo struct {
	Key   string        `json:"key"`
	Value valorAtributo `json:"value"`
}

// peticionLogs es `/v1/logs`, el canal PRIMARIO (V1): es por donde viaja `api_request`, o
// sea el dinero.
type peticionLogs struct {
	ResourceLogs []struct {
		Resource struct {
			Attributes []atributo `json:"attributes"`
		} `json:"resource"`
		ScopeLogs []struct {
			Scope struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"scope"`
			LogRecords []struct {
				TimeUnixNano numeroTolerante `json:"timeUnixNano"`
				Attributes   []atributo      `json:"attributes"`
				// Body se decodifica para no romper el parseo, pero NO se usa: los eventos
				// útiles viven en los atributos y el cuerpo puede traer texto (§4.3).
				Body struct {
					StringValue *string `json:"stringValue"`
				} `json:"body"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	} `json:"resourceLogs"`
}

// peticionMetricas es `/v1/metrics`, el canal SECUNDARIO.
type peticionMetricas struct {
	ResourceMetrics []struct {
		Resource struct {
			Attributes []atributo `json:"attributes"`
		} `json:"resource"`
		ScopeMetrics []struct {
			Scope struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"scope"`
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

// numeroTolerante es un `json.Number` que NO invalida el lote cuando el emisor manda basura
// donde iba un número. Se usa SOLO en `timeUnixNano`: un timestamp ilegible deja el evento
// sin reloj de emisor —y manda el del daemon, A10—, pero rechazar el lote entero por eso
// tiraría el dinero de todos los demás records del mismo cuerpo.
type numeroTolerante struct{ n json.Number }

func (t *numeroTolerante) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}
	if len(b) >= 2 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return nil //nolint:nilerr // ilegible = sin reloj de emisor, no lote inválido.
		}
		t.n = json.Number(s)
		return nil
	}
	t.n = json.Number(b)
	return nil
}

type puntoDato struct {
	Attributes   []atributo      `json:"attributes"`
	AsInt        *json.Number    `json:"asInt"`
	AsDouble     *float64        `json:"asDouble"`
	TimeUnixNano numeroTolerante `json:"timeUnixNano"`
}

// Atributos es un lote de atributos ya aplanado, con acceso tipado. El mapeo lee de acá
// clave por clave — NUNCA copia el mapa entero a un evento (§6.1: `NuevoEvento(map)` no
// existe, y no debe existir).
type Atributos map[string]valorAtributo

// Texto devuelve el valor de una clave como cadena, o "" si no está o no es texto.
func (a Atributos) Texto(k string) string {
	v, ok := a[k]
	if !ok || v.StringValue == nil {
		return ""
	}
	return *v.StringValue
}

// Entero devuelve el valor de una clave como int64.
//
// Acepta las TRES formas que se ven en el wire real: `intValue` numérico (lo que Claude Code
// emite, off-spec), `intValue` como string (lo que manda la spec protobuf→JSON), y
// `stringValue` con un número adentro — que es como llegan `duration_ms` y
// `tool_result_size_bytes` en los `tool_result` medidos (ANEXO H8).
//
// `ok=false` cuando la clave no está o no es legible. **No devuelve 0 con ok=true**: un 0
// fabricado por un atributo ilegible se sumaría a un total como si fuera una medición.
func (a Atributos) Entero(k string) (int64, bool) {
	v, ok := a[k]
	if !ok {
		return 0, false
	}
	if v.IntValue != nil {
		n, err := v.IntValue.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	}
	if v.DoubleValue != nil {
		return int64(*v.DoubleValue), true
	}
	if v.StringValue != nil {
		var n json.Number = json.Number(*v.StringValue)
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	}
	return 0, false
}

// Decimal devuelve el valor de una clave como float64.
func (a Atributos) Decimal(k string) (float64, bool) {
	v, ok := a[k]
	if !ok {
		return 0, false
	}
	switch {
	case v.DoubleValue != nil:
		return *v.DoubleValue, true
	case v.IntValue != nil:
		f, err := v.IntValue.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case v.StringValue != nil:
		f, err := json.Number(*v.StringValue).Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	}
	return 0, false
}

// RegistroLog es un log record ya aplanado: los atributos del recurso y los del record en un
// solo mapa (los del record ganan), más el scope y el timestamp.
type RegistroLog struct {
	Scope        string
	ScopeVersion string
	TS           *time.Time
	Attrs        Atributos
}

// PuntoMetrica es un data point ya aplanado.
type PuntoMetrica struct {
	Metrica string
	Unidad  string
	Scope   string
	TS      *time.Time
	Attrs   Atributos
	// Valor y ValorDecimal: solo uno está poblado. `Legible` false = el punto llegó con un
	// valor que no se pudo leer, y entonces NO vale 0: vale «sin dato».
	Valor        int64
	ValorDecimal float64
	EsDecimal    bool
	Legible      bool
	// Temporalidad 1 = Delta (lo que Claude Code manda, V5.1), 2 = Cumulative.
	// TemporalidadSoportada es false para Cumulative: **no se adivina un delta sin estado
	// previo**. El punto entra marcado y suma a `salud.temporalidad_no_soportada`.
	Temporalidad          int
	TemporalidadSoportada bool
	Monotonica            bool
}

// DecodificarLogs decodifica un cuerpo `/v1/logs`.
//
// Un JSON sintácticamente roto devuelve `ErrPayloadInvalido` y **cero registros**: el lote
// entero se rechaza, jamás se guarda a medias. Un campo desconocido NO es error —
// `encoding/json` lo descarta y el evento entra igual, que es lo que hace que una versión
// nueva del runtime no rompa el receptor.
func DecodificarLogs(cuerpo []byte) ([]RegistroLog, error) {
	var p peticionLogs
	dec := json.NewDecoder(bytes.NewReader(cuerpo))
	dec.UseNumber()
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPayloadInvalido, err)
	}

	var out []RegistroLog
	for _, rl := range p.ResourceLogs {
		base := aplanar(nil, rl.Resource.Attributes)
		for _, sl := range rl.ScopeLogs {
			for _, lr := range sl.LogRecords {
				out = append(out, RegistroLog{
					Scope:        sl.Scope.Name,
					ScopeVersion: sl.Scope.Version,
					TS:           deUnixNano(lr.TimeUnixNano),
					Attrs:        aplanar(base, lr.Attributes),
				})
			}
		}
	}
	return out, nil
}

// DecodificarMetricas decodifica un cuerpo `/v1/metrics`. Solo se leen métricas `Sum`:
// `Gauge`, `Histogram` y `ExponentialHistogram` se ignoran explícitamente (Claude Code manda
// `Sum` en las cuatro métricas medidas).
func DecodificarMetricas(cuerpo []byte) ([]PuntoMetrica, error) {
	var p peticionMetricas
	dec := json.NewDecoder(bytes.NewReader(cuerpo))
	dec.UseNumber()
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPayloadInvalido, err)
	}

	var out []PuntoMetrica
	for _, rm := range p.ResourceMetrics {
		base := aplanar(nil, rm.Resource.Attributes)
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				if m.Sum == nil {
					continue // Gauge/Histogram: ignorados a propósito (§4.3).
				}
				for _, dp := range m.Sum.DataPoints {
					// ⚠️ Claude Code manda los CONTEOS DE TOKENS como `asDouble`, no como
					// `asInt` (medido: `claude_code.token.usage` = 10 / 39 / 17536 / 8257,
					// todos `asDouble`). Es la segunda desviación del wire, hermana del
					// `intValue` numérico. Por eso el punto lleva los dos campos y el flag
					// `EsDecimal`: convertir a int64 a ciegas perdería el costo, que sí es
					// decimal de verdad (`claude_code.cost.usage` = 0,0184726).
					pt := PuntoMetrica{
						Metrica: m.Name,
						Unidad:  m.Unit,
						Scope:   sm.Scope.Name,
						TS:      deUnixNano(dp.TimeUnixNano),
						Attrs:   aplanar(base, dp.Attributes),
						// Temporalidad 1 = Delta ⇒ se suma y listo. Cualquier otra cosa NO
						// se adivina: fabricar un delta a partir de un acumulado sin estado
						// previo es inventar.
						Temporalidad:          m.Sum.AggregationTemporality,
						TemporalidadSoportada: m.Sum.AggregationTemporality == 1,
						Monotonica:            m.Sum.IsMonotonic,
					}
					switch {
					case dp.AsInt != nil:
						if n, err := dp.AsInt.Int64(); err == nil {
							pt.Valor, pt.Legible = n, true
						}
					case dp.AsDouble != nil:
						pt.ValorDecimal, pt.EsDecimal, pt.Legible = *dp.AsDouble, true, true
					}
					out = append(out, pt)
				}
			}
		}
	}
	return out, nil
}

// aplanar mezcla los atributos del recurso con los del record. Los del record GANAN: son más
// específicos. `base` no se muta — se copia — porque el mismo recurso alimenta varios records.
func aplanar(base Atributos, extra []atributo) Atributos {
	out := make(Atributos, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for _, a := range extra {
		out[a.Key] = a.Value
	}
	return out
}

// deUnixNano convierte el `timeUnixNano` del wire. Un valor no numérico devuelve **nil**, no
// el epoch: el evento entra con `TSEmisor` vacío y manda el reloj del daemon (A10). Un
// 1970-01-01 fabricado caería fuera de toda ventana y el evento desaparecería en silencio.
func deUnixNano(nt numeroTolerante) *time.Time {
	n := nt.n
	if n == "" {
		return nil
	}
	i, err := n.Int64()
	if err != nil || i <= 0 {
		return nil
	}
	t := time.Unix(0, i).UTC()
	return &t
}
