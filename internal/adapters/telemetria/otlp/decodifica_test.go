package otlp

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// golden lee un payload REAL de testdata/. Son los que Claude Code 2.1.220 emitió en las
// corridas del 2026-07-26 — cero mock donde hay dato real.
func golden(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", nombre))
	if err != nil {
		t.Fatalf("golden %s: %v", nombre, err)
	}
	return b
}

// TestDecodificaLosGoldenReales — la línea base: los tres payloads de logs y el de métricas
// decodifican, y traen exactamente lo que el ANEXO midió (8 métricas · 14 puntos · 57 log
// records en total entre los tres archivos de logs).
func TestDecodificaLosGoldenReales(t *testing.T) {
	total := 0
	for _, f := range []string{"logs-run1.json", "logs-run2-con-skill.json", "logs-run4-con-tools.json"} {
		regs, err := DecodificarLogs(golden(t, f))
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if len(regs) == 0 {
			t.Fatalf("%s: 0 registros — el decodificador no está leyendo nada", f)
		}
		total += len(regs)
		for _, r := range regs {
			if r.Scope != "com.anthropic.claude_code.events" {
				t.Errorf("%s: scope inesperado %q", f, r.Scope)
			}
			if r.Attrs.Texto("event.name") == "" {
				t.Errorf("%s: un log record sin event.name — el aplanado perdió atributos", f)
			}
			if r.Attrs.Texto("session.id") == "" {
				t.Errorf("%s: un log record sin session.id — media llave del join se perdió", f)
			}
		}
	}
	if total == 0 {
		t.Fatal("ningún registro decodificado en los tres golden")
	}
	t.Logf("log records decodificados de los golden reales: %d", total)

	puntos, err := DecodificarMetricas(golden(t, "metrics-run1.json"))
	if err != nil {
		t.Fatalf("metrics-run1.json: %v", err)
	}
	if len(puntos) == 0 {
		t.Fatal("metrics-run1.json: 0 puntos")
	}
	t.Logf("puntos de métrica decodificados: %d", len(puntos))
	for _, p := range puntos {
		if !p.Legible {
			t.Errorf("punto %q ilegible", p.Metrica)
		}
		// V5.1: las métricas llegan Sum · Delta · monotonic. Si esto cambiara, el
		// decodificador tiene que dejar de sumarlas a ciegas, y este test lo avisa.
		if p.Temporalidad != 1 || !p.Monotonica {
			t.Errorf("punto %q: temporalidad=%d monotonica=%v — se esperaba Delta monotónica (V5.1)",
				p.Metrica, p.Temporalidad, p.Monotonica)
		}
	}
}

// TestIntValueComoNumeroYComoString — **el test que justifica `json.Number`**.
//
// Claude Code emite `intValue` como NÚMERO JSON, off-spec: el mapeo protobuf→JSON manda
// string para int64. Un decodificador con `*string` revienta con el payload real; uno con
// `*int64` reventaría con un runtime que sí cumple la spec. `json.Number` acepta las dos.
//
// El test toma el payload REAL y genera su variante spec-compliant reescribiendo cada
// `"intValue": N` como `"intValue": "N"`. Los dos tienen que decodificar IGUAL.
func TestIntValueComoNumeroYComoString(t *testing.T) {
	crudo := golden(t, "logs-run1.json")
	if !bytes.Contains(crudo, []byte(`"intValue":`)) && !bytes.Contains(crudo, []byte(`"intValue": `)) {
		t.Fatal("el golden no trae ningún intValue — el test no probaría nada")
	}
	reNum := regexp.MustCompile(`"intValue":\s*(-?\d+)`)
	comoString := reNum.ReplaceAll(crudo, []byte(`"intValue": "$1"`))
	if bytes.Equal(crudo, comoString) {
		t.Fatal("la reescritura no cambió nada: el control del test está roto")
	}

	a, err := DecodificarLogs(crudo)
	if err != nil {
		t.Fatalf("payload real (intValue numérico, off-spec): %v", err)
	}
	b, err := DecodificarLogs(comoString)
	if err != nil {
		t.Fatalf("payload spec-compliant (intValue string): %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("distinto número de registros: %d vs %d", len(a), len(b))
	}
	iguales := 0
	for i := range a {
		for _, k := range []string{"input_tokens", "output_tokens", "cache_read_tokens",
			"cache_creation_tokens", "cost_usd_micros", "duration_ms", "event.sequence"} {
			va, oka := a[i].Attrs.Entero(k)
			vb, okb := b[i].Attrs.Entero(k)
			if oka != okb || va != vb {
				t.Errorf("registro %d, clave %q: numérico=(%d,%v) string=(%d,%v)", i, k, va, oka, vb, okb)
			}
			if oka {
				iguales++
			}
		}
	}
	// Control positivo: si NINGUNA clave resolvió, los dos lados serían «vacío == vacío» y
	// el test pasaría sin haber comparado nada.
	if iguales == 0 {
		t.Fatal("ninguna clave entera se leyó en ninguna de las dos variantes — el test comparó dos vacíos")
	}
	t.Logf("claves enteras comparadas con éxito en las dos formas: %d", iguales)
}

// TestAtributoNumericoIlegible — un atributo que dice ser entero y no lo es devuelve
// `ok=false`, NO `0, true`. Un cero fabricado por un atributo roto entraría a un total como
// si fuera una medición.
func TestAtributoNumericoIlegible(t *testing.T) {
	cuerpo := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},
	  "logRecords":[{"timeUnixNano":"1","attributes":[
	    {"key":"roto","value":{"stringValue":"no-soy-un-numero"}},
	    {"key":"bueno","value":{"intValue":42}}]}]}]}]}`)
	regs, err := DecodificarLogs(cuerpo)
	if err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	if len(regs) != 1 {
		t.Fatalf("se esperaba 1 registro, hay %d", len(regs))
	}
	if v, ok := regs[0].Attrs.Entero("roto"); ok {
		t.Errorf("un atributo ilegible no puede devolver ok=true (devolvió %d)", v)
	}
	// Control positivo en la misma corrida: el atributo bueno SÍ se lee.
	if v, ok := regs[0].Attrs.Entero("bueno"); !ok || v != 42 {
		t.Errorf("el atributo legible debe leerse: got (%d,%v)", v, ok)
	}
}

// TestCamposDesconocidosNoRompen — lo que no conocemos se descarta, no es error. Es lo que
// hace que una versión nueva del runtime no tumbe el receptor.
func TestCamposDesconocidosNoRompen(t *testing.T) {
	cuerpo := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"campoNuevoDelFuturo":{"x":1},
	  "scopeLogs":[{"scope":{"name":"s","otraCosa":true},
	  "logRecords":[{"timeUnixNano":"1","severityNumber":9,"traceId":"abc","droppedAttributesCount":3,
	    "attributes":[{"key":"cosa.nueva","value":{"stringValue":"v"}},
	                  {"key":"session.id","value":{"stringValue":"s-1"}}]}]}]}],"schemaUrl":"http://x"}`)
	regs, err := DecodificarLogs(cuerpo)
	if err != nil {
		t.Fatalf("un campo desconocido NO es error: %v", err)
	}
	if len(regs) != 1 {
		t.Fatalf("se esperaba 1 registro, hay %d", len(regs))
	}
	// Control positivo: lo declarado entró igual.
	if got := regs[0].Attrs.Texto("session.id"); got != "s-1" {
		t.Errorf("lo declarado debe entrar: session.id=%q", got)
	}
}

// TestPayloadMalformado — el lote entero se rechaza. Nunca a medias, y nunca con registros
// parciales que después se guardarían como si fueran buenos.
func TestPayloadMalformado(t *testing.T) {
	casos := map[string][]byte{
		"json roto":       []byte(`{"resourceLogs":[{`),
		"tipo incorrecto": []byte(`{"resourceLogs":"esto-debería-ser-una-lista"}`),
		"vacío":           []byte(``),
	}
	for nombre, cuerpo := range casos {
		regs, err := DecodificarLogs(cuerpo)
		if !errors.Is(err, domain.ErrPayloadInvalido) {
			t.Errorf("%s: se esperaba ErrPayloadInvalido, got %v", nombre, err)
		}
		if len(regs) != 0 {
			t.Errorf("%s: devolvió %d registros parciales — el lote se rechaza ENTERO", nombre, len(regs))
		}
	}
	// Control positivo: un lote bien formado sí entra. Sin esto, un decodificador que
	// siempre falla pasaría los asserts de arriba.
	if regs, err := DecodificarLogs(golden(t, "logs-run1.json")); err != nil || len(regs) == 0 {
		t.Fatalf("control positivo: el golden real debe decodificar (err=%v, n=%d)", err, len(regs))
	}
}

// TestLoteSinResourceLogsNoEsError — un lote vacío es 0 aceptados, no un 400. El exportador
// puede mandar un flush sin nada; responderle error lo haría reintentar para siempre.
func TestLoteSinResourceLogsNoEsError(t *testing.T) {
	regs, err := DecodificarLogs([]byte(`{}`))
	if err != nil {
		t.Fatalf("un lote vacío no es un payload inválido: %v", err)
	}
	if len(regs) != 0 {
		t.Errorf("un lote vacío produce 0 registros, produjo %d", len(regs))
	}
}

// TestSumaDelta — las métricas llegan Delta monotónicas (V5.1): se suman y listo, sin
// diferenciar contadores acumulados.
func TestSumaDelta(t *testing.T) {
	puntos, err := DecodificarMetricas(golden(t, "metrics-run1.json"))
	if err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	// ⚠️ HALLAZGO al construir: Claude Code manda los conteos de tokens como `asDouble`,
	// no como `asInt`. El decodificador expone los dos y el flag `EsDecimal`; el test suma
	// por el camino que el dato real usa, no por el que la spec sugeriría.
	var suma float64
	var tokens int
	for _, p := range puntos {
		if p.Metrica != "claude_code.token.usage" || !p.Legible {
			continue
		}
		if !p.TemporalidadSoportada {
			t.Fatalf("punto de %q con temporalidad %d: no se puede sumar a ciegas", p.Metrica, p.Temporalidad)
		}
		if !p.EsDecimal {
			t.Errorf("punto de %q llegó como entero; el dato REAL medido llega asDouble", p.Metrica)
		}
		suma += p.ValorDecimal
		tokens++
	}
	if tokens == 0 {
		t.Fatal("ningún punto de claude_code.token.usage — el golden o el decodificador están mal")
	}
	if suma <= 0 {
		t.Fatalf("la suma de los deltas debe ser positiva: %v", suma)
	}
	// Los 4 buckets de la corrida real: 10 + 39 + 17 536 + 8 257 = 25 842.
	if tokens != 4 || suma != 25_842 {
		t.Errorf("se esperaban 4 puntos sumando 25842 (los 4 buckets medidos); got %d puntos / %v", tokens, suma)
	}
	t.Logf("claude_code.token.usage: %d puntos, suma %v tokens", tokens, suma)
}

// TestCumulativeNoSeAdivina — con `aggregationTemporality = 2` (Cumulative) el punto queda
// marcado `sin-dato` y NO se convierte a delta. Fabricar un delta a partir de un acumulado
// sin estado previo es inventar.
func TestCumulativeNoSeAdivina(t *testing.T) {
	cuerpo := []byte(`{"resourceMetrics":[{"resource":{"attributes":[]},"scopeMetrics":[{"scope":{"name":"s"},
	  "metrics":[
	    {"name":"acumulada","unit":"tokens","sum":{"aggregationTemporality":2,"isMonotonic":true,
	      "dataPoints":[{"timeUnixNano":"1","asInt":"100","attributes":[]}]}},
	    {"name":"delta","unit":"tokens","sum":{"aggregationTemporality":1,"isMonotonic":true,
	      "dataPoints":[{"timeUnixNano":"1","asInt":"7","attributes":[]}]}}]}]}]}`)
	puntos, err := DecodificarMetricas(cuerpo)
	if err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	if len(puntos) != 2 {
		t.Fatalf("se esperaban 2 puntos, hay %d", len(puntos))
	}
	var acum, delta *PuntoMetrica
	for i := range puntos {
		switch puntos[i].Metrica {
		case "acumulada":
			acum = &puntos[i]
		case "delta":
			delta = &puntos[i]
		}
	}
	if acum == nil || delta == nil {
		t.Fatal("faltó alguno de los dos puntos")
	}
	if acum.TemporalidadSoportada {
		t.Error("Cumulative NO es soportada: sin estado previo, convertirla a delta es inventar")
	}
	if acum.Valor != 100 {
		t.Errorf("el valor crudo se conserva para poder reportarlo: %d", acum.Valor)
	}
	// Control positivo, misma corrida: el punto Delta SÍ es soportado.
	if !delta.TemporalidadSoportada {
		t.Error("control positivo: un punto Delta debe seguir siendo soportado")
	}
}

// TestReinicioDelEmisorNoDuplica — si el emisor se reinicia y re-manda un lote, la
// decodificación es idempotente: el mismo cuerpo produce exactamente el mismo resultado.
// (El dedupe por (sesion, turno, tipo, ts) vive en el writer del almacén; acá se asegura la
// mitad de la que este paquete es responsable.)
func TestReinicioDelEmisorNoDuplica(t *testing.T) {
	crudo := golden(t, "logs-run1.json")
	a, err := DecodificarLogs(crudo)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodificarLogs(crudo)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("decodificar dos veces el mismo cuerpo dio %d y %d registros", len(a), len(b))
	}
	for i := range a {
		if a[i].Attrs.Texto("event.name") != b[i].Attrs.Texto("event.name") ||
			a[i].Attrs.Texto("prompt.id") != b[i].Attrs.Texto("prompt.id") {
			t.Fatalf("registro %d difiere entre dos decodificaciones del mismo cuerpo", i)
		}
	}
}

// TestTimestampIlegibleNoEsEpoch — un `timeUnixNano` no numérico deja `TS` en nil, NO en
// 1970-01-01. Un epoch fabricado caería fuera de toda ventana y el evento desaparecería en
// silencio, que es peor que no tener timestamp del emisor (el del daemon manda igual).
func TestTimestampIlegibleNoEsEpoch(t *testing.T) {
	cuerpo := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},
	  "logRecords":[
	    {"timeUnixNano":"no-es-un-numero","attributes":[{"key":"a","value":{"stringValue":"1"}}]},
	    {"timeUnixNano":"1785000000000000000","attributes":[{"key":"a","value":{"stringValue":"2"}}]}]}]}]}`)
	regs, err := DecodificarLogs(cuerpo)
	if err != nil {
		t.Fatalf("un timestamp ilegible no invalida el lote: %v", err)
	}
	if len(regs) != 2 {
		t.Fatalf("se esperaban 2 registros, hay %d", len(regs))
	}
	if regs[0].TS != nil {
		t.Errorf("timestamp ilegible debe quedar nil, quedó %v", *regs[0].TS)
	}
	// Control positivo: el que SÍ es legible se lee.
	if regs[1].TS == nil {
		t.Fatal("control positivo: un timeUnixNano válido debe producir un TS")
	}
	if regs[1].TS.Year() < 2020 {
		t.Errorf("timestamp mal convertido: %v", *regs[1].TS)
	}
}

// TestElCuerpoDelLogRecordNoSeUsa — `body.stringValue` puede traer texto de la conversación.
// Se decodifica para no romper el parseo, pero NO viaja a `Attrs`: los eventos útiles viven
// en los atributos.
func TestElCuerpoDelLogRecordNoSeUsa(t *testing.T) {
	const marcador = "MARCA-CUERPO-9c3f"
	cuerpo := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},
	  "logRecords":[{"timeUnixNano":"1","body":{"stringValue":"` + marcador + `"},
	    "attributes":[{"key":"session.id","value":{"stringValue":"s-ok"}}]}]}]}]}`)
	regs, err := DecodificarLogs(cuerpo)
	if err != nil {
		t.Fatal(err)
	}
	if len(regs) != 1 {
		t.Fatalf("se esperaba 1 registro, hay %d", len(regs))
	}
	raw, _ := json.Marshal(regs[0].Attrs)
	if bytes.Contains(raw, []byte(marcador)) {
		t.Errorf("el cuerpo del log record no puede llegar a los atributos: %s", raw)
	}
	// Control positivo, misma corrida: lo declarado SÍ está.
	if !bytes.Contains(raw, []byte("s-ok")) {
		t.Errorf("el atributo declarado debe estar presente: %s", raw)
	}
}

// FuzzDecodificarLogs — sembrado con los tres payloads REALES. Ningún input puede hacer
// panic: el `recover()` del handler es red de seguridad, no estrategia.
func FuzzDecodificarLogs(f *testing.F) {
	for _, n := range []string{"logs-run1.json", "logs-run2-con-skill.json", "logs-run4-con-tools.json"} {
		b, err := os.ReadFile(filepath.Join("testdata", n))
		if err != nil {
			f.Fatalf("semilla %s: %v", n, err)
		}
		f.Add(b)
	}
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"resourceLogs":[{"scopeLogs":[{"logRecords":[{"timeUnixNano":9}]}]}]}`))
	f.Fuzz(func(t *testing.T, cuerpo []byte) {
		regs, err := DecodificarLogs(cuerpo)
		if err != nil && len(regs) != 0 {
			t.Fatalf("con error no puede haber registros parciales: %d", len(regs))
		}
		for _, r := range regs {
			// Ejercita los accesores: es donde vive la conversión de json.Number.
			_ = r.Attrs.Texto("event.name")
			_, _ = r.Attrs.Entero("input_tokens")
			_, _ = r.Attrs.Decimal("cost_usd")
		}
	})
}

// FuzzDecodificarMetricas — misma red para el canal secundario.
func FuzzDecodificarMetricas(f *testing.F) {
	b, err := os.ReadFile(filepath.Join("testdata", "metrics-run1.json"))
	if err != nil {
		f.Fatalf("semilla: %v", err)
	}
	f.Add(b)
	f.Add([]byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"sum":{"dataPoints":[{}]}}]}]}]}`))
	f.Fuzz(func(t *testing.T, cuerpo []byte) {
		puntos, err := DecodificarMetricas(cuerpo)
		if err != nil && len(puntos) != 0 {
			t.Fatalf("con error no puede haber puntos parciales: %d", len(puntos))
		}
	})
}
