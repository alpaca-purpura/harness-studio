package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// postDiagnosticoReq arma el request tal como lo manda el FE.
func postDiagnosticoReq(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/api/diagnostico", strings.NewReader(body))
}

// conLogCapturado corre f con el logger por default apuntando a un buffer, y devuelve lo que
// se escribió. Es lo único que importa de este endpoint: **que el detalle llegue al log**.
func conLogCapturado(t *testing.T, f func()) string {
	t.Helper()
	var buf bytes.Buffer
	previo := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(previo)
	f()
	return buf.String()
}

func TestDiagnosticoEscribeElDetalleAlLog(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"origen":"fe","evento":"dictado.sin-bloques","mensaje":"El micrófono no entregó nada",
	          "detalle":{"bloques":0,"muestras":0,"pico":0,"hzEntrada":44100,"mime":"audio/wav"}}`

	log := conLogCapturado(t, func() { postDiagnostico()(w, postDiagnosticoReq(body)) })

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, quiero 204", w.Code)
	}
	// Lo que se está probando es que el fallo quede reconstruible SIN reproducirlo: si el
	// detalle no viaja al log, este endpoint no sirve para nada.
	for _, quiero := range []string{"dictado.sin-bloques", "hzEntrada", "44100", "audio/wav", "El micrófono no entregó nada"} {
		if !strings.Contains(log, quiero) {
			t.Errorf("el log no trae %q:\n%s", quiero, log)
		}
	}
}

// Un evento sin nombre no se puede grepear después, que es lo único para lo que existe.
func TestDiagnosticoSinEventoEsRechazado(t *testing.T) {
	w := httptest.NewRecorder()
	postDiagnostico()(w, postDiagnosticoReq(`{"origen":"fe","mensaje":"algo"}`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400", w.Code)
	}
}

func TestDiagnosticoIlegibleEsRechazado(t *testing.T) {
	w := httptest.NewRecorder()
	postDiagnostico()(w, postDiagnosticoReq(`{no soy json`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400", w.Code)
	}
}

// El origen ausente no bloquea el reporte: perder un fallo por un campo cosmético sería
// exactamente el silencio que este endpoint vino a sacar.
func TestDiagnosticoSinOrigenSeAceptaYSeMarca(t *testing.T) {
	w := httptest.NewRecorder()
	log := conLogCapturado(t, func() {
		postDiagnostico()(w, postDiagnosticoReq(`{"evento":"fe.error"}`))
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, quiero 204", w.Code)
	}
	if !strings.Contains(log, "desconocido") {
		t.Errorf("el origen ausente no quedó marcado:\n%s", log)
	}
}

// Las claves son para grepear, no para prosa: sin tope, un evento gigante haría ilegible el
// archivo que este endpoint existe para hacer legible.
func TestDiagnosticoRecortaLasClaves(t *testing.T) {
	largo := strings.Repeat("x", maxClave+50)
	w := httptest.NewRecorder()
	log := conLogCapturado(t, func() {
		body, _ := json.Marshal(map[string]string{"origen": "fe", "evento": largo})
		postDiagnostico()(w, postDiagnosticoReq(string(body)))
	})
	if strings.Contains(log, largo) {
		t.Error("el evento entró entero al log: no se recortó")
	}
	if !strings.Contains(log, strings.Repeat("x", maxClave)) {
		t.Errorf("no quedó el prefijo recortado:\n%s", log)
	}
}

// El cuerpo está acotado: este endpoint no puede ser la vía para hacerle crecer el disco al
// operador desde la superficie local.
func TestDiagnosticoAcotaElCuerpo(t *testing.T) {
	enorme, _ := json.Marshal(map[string]string{
		"origen": "fe", "evento": "fe.error", "mensaje": strings.Repeat("y", maxDiagnostico+1024),
	})
	w := httptest.NewRecorder()
	req := postDiagnosticoReq(string(enorme))
	conLogCapturado(t, func() { postDiagnostico()(w, req) })
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400 (cuerpo cortado por MaxBytesReader)", w.Code)
	}
}

// El endpoint NO cuelga de ningún servicio: el primer fallo que hay que poder diagnosticar es
// el del arranque, cuando el resto del daemon todavía puede no estar cableado.
func TestDiagnosticoVivoSinDictadoCableado(t *testing.T) {
	vacio := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	h := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, vacio, vacio,
		AuthConfig{AllowedHosts: []string{"127.0.0.1:4200"}})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/diagnostico", strings.NewReader(`{"evento":"fe.error"}`))
	r.Host = "127.0.0.1:4200"
	conLogCapturado(t, func() { h.ServeHTTP(w, r.WithContext(context.Background())) })
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, quiero 204", w.Code)
	}
}
