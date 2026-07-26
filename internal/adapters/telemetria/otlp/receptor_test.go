package otlp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/otlp"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// sinkEspia guarda lo que llega. Es el instrumento del CONTROL POSITIVO: sin él, «no llegó
// nada» sería indistinguible de «el receptor está roto».
type sinkEspia struct {
	mu        sync.Mutex
	recibidos []domain.EventoTelemetria
	// tope simula una cola acotada: pasado ese número, acepta 0 y devuelve ErrColaLlena.
	tope int
}

func (s *sinkEspia) Ingerir(_ context.Context, evs []domain.EventoTelemetria) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tope > 0 && len(s.recibidos) >= s.tope {
		return 0, ports.ErrColaLlena
	}
	s.recibidos = append(s.recibidos, evs...)
	return len(evs), nil
}

// conMarcador cuenta los eventos cuya sesión contiene el marcador. **La forma canónica del
// assert** de §4: `len(conMarcadorA) == 0 && len(conMarcadorB) == 1`, jamás
// `len(recibidos) == 0`.
func (s *sinkEspia) conMarcador(m string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, e := range s.recibidos {
		if strings.Contains(e.SesionID, m) {
			n++
		}
	}
	return n
}

// servidorEfimero levanta el receptor en un `httptest.Server` de PUERTO EFÍMERO. Nunca un
// puerto fijo: los tres primeros intentos de la verificación H9 dieron negativo y el negativo
// era FALSO — un receptor viejo seguía pegado al puerto.
func servidorEfimero(t *testing.T, sink ports.TelemetriaSink) (*httptest.Server, *otlp.Receptor) {
	t.Helper()
	r := otlp.NewReceptor(sink, otlp.Opciones{
		Reloj: func() time.Time { return time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC) },
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, r
}

func golden(t *testing.T, n string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", n))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// conMarca reescribe el session.id del golden con un marcador único, para poder distinguir
// una variante de otra. **Marcadores distintos por variante** (§4 regla 3): dos variantes con
// el mismo marcador no se distinguen si una filtra a la otra.
func conMarca(t *testing.T, archivo, marca string) []byte {
	t.Helper()
	crudo := golden(t, archivo)
	var doc map[string]any
	if err := json.Unmarshal(crudo, &doc); err != nil {
		t.Fatal(err)
	}
	reescribir(doc, marca)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func reescribir(v any, marca string) {
	switch x := v.(type) {
	case map[string]any:
		if k, ok := x["key"].(string); ok && k == "session.id" {
			if val, ok := x["value"].(map[string]any); ok {
				val["stringValue"] = marca
			}
		}
		for _, sub := range x {
			reescribir(sub, marca)
		}
	case []any:
		for _, sub := range x {
			reescribir(sub, marca)
		}
	}
}

func postear(t *testing.T, url, ct string, cuerpo []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(cuerpo))
	if err != nil {
		t.Fatal(err)
	}
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// TestProtobufSeRechazaConMotivo — 415 nombrando la causa Y el arreglo. **Jamás un 200 que
// finge haber guardado**, que es la forma de perder telemetría sin que nadie se entere.
//
// ⚠ CONTROL POSITIVO en la misma corrida y contra el MISMO servidor: un POST válido con
// marcador DISTINTO que sí tiene que llegar.
func TestProtobufSeRechazaConMotivo(t *testing.T) {
	sink := &sinkEspia{}
	srv, _ := servidorEfimero(t, sink)

	// Variante A · protobuf: no tiene que llegar.
	respA := postear(t, srv.URL+"/v1/logs", "application/x-protobuf", conMarca(t, "logs-run1.json", "MARCA-A-PROTOBUF"))
	if respA.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("protobuf debe dar 415, dio %d", respA.StatusCode)
	}
	cuerpo := make([]byte, 512)
	n, _ := respA.Body.Read(cuerpo)
	texto := string(cuerpo[:n])
	if !strings.Contains(texto, "OTEL_EXPORTER_OTLP_PROTOCOL=http/json") {
		t.Errorf("el 415 tiene que nombrar el arreglo, no solo la causa: %q", texto)
	}

	// Variante B · JSON válido: SÍ tiene que llegar.
	respB := postear(t, srv.URL+"/v1/logs", "application/json", conMarca(t, "logs-run1.json", "MARCA-B-VALIDO"))
	if respB.StatusCode != http.StatusOK {
		t.Fatalf("un lote válido debe dar 200, dio %d", respB.StatusCode)
	}

	// ── el assert canónico de §4 ──
	if a, b := sink.conMarcador("MARCA-A-PROTOBUF"), sink.conMarcador("MARCA-B-VALIDO"); a != 0 || b == 0 {
		t.Fatalf("control positivo/negativo: conA=%d (debe ser 0) conB=%d (debe ser >0)", a, b)
	}
}

// TestPayloadGiganteSeRechazaSinLeerlo — 413, y el cuerpo no se lee entero a memoria.
func TestPayloadGiganteSeRechazaSinLeerlo(t *testing.T) {
	sink := &sinkEspia{}
	r := otlp.NewReceptor(sink, otlp.Opciones{MaxBody: 64 << 10}) // tope chico para el test
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Variante A · 8 MiB de relleno con su marcador: no tiene que llegar.
	gigante := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},"logRecords":[{"attributes":[{"key":"session.id","value":{"stringValue":"MARCA-A-GIGANTE"}},{"key":"relleno","value":{"stringValue":"`)
	gigante = append(gigante, bytes.Repeat([]byte("x"), 8<<20)...)
	gigante = append(gigante, []byte(`"}}]}]}]}`)...)
	respA := postear(t, srv.URL+"/v1/logs", "application/json", gigante)
	if respA.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("un cuerpo de 8 MiB debe dar 413, dio %d", respA.StatusCode)
	}

	// Variante B · lote chico con OTRO marcador: sí llega.
	chico := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},
	  "logRecords":[{"attributes":[
	    {"key":"session.id","value":{"stringValue":"MARCA-B-CHICO"}},
	    {"key":"event.name","value":{"stringValue":"api_request"}},
	    {"key":"model","value":{"stringValue":"claude-haiku-4-5"}}]}]}]}]}`)
	respB := postear(t, srv.URL+"/v1/logs", "application/json", chico)
	if respB.StatusCode != http.StatusOK {
		t.Fatalf("un lote chico debe dar 200, dio %d", respB.StatusCode)
	}
	if a, b := sink.conMarcador("MARCA-A-GIGANTE"), sink.conMarcador("MARCA-B-CHICO"); a != 0 || b != 1 {
		t.Fatalf("conA=%d (debe ser 0) conB=%d (debe ser 1)", a, b)
	}
}

// TestLoteMalformadoSeRechazaEntero — 400 y **nada a medias**.
func TestLoteMalformadoSeRechazaEntero(t *testing.T) {
	sink := &sinkEspia{}
	srv, _ := servidorEfimero(t, sink)

	// El JSON se rompe DESPUÉS del primer registro: un decodificador que guardara lo leído
	// hasta el error dejaría medio lote adentro.
	roto := []byte(`{"resourceLogs":[{"resource":{"attributes":[]},"scopeLogs":[{"scope":{"name":"s"},
	  "logRecords":[{"attributes":[{"key":"session.id","value":{"stringValue":"MARCA-A-ROTO"}}]},`)
	respA := postear(t, srv.URL+"/v1/logs", "application/json", roto)
	if respA.StatusCode != http.StatusBadRequest {
		t.Fatalf("un lote roto debe dar 400, dio %d", respA.StatusCode)
	}

	respB := postear(t, srv.URL+"/v1/logs", "application/json", conMarca(t, "logs-run1.json", "MARCA-B-SANO"))
	if respB.StatusCode != http.StatusOK {
		t.Fatalf("el lote sano debe dar 200, dio %d", respB.StatusCode)
	}
	if a, b := sink.conMarcador("MARCA-A-ROTO"), sink.conMarcador("MARCA-B-SANO"); a != 0 || b == 0 {
		t.Fatalf("conA=%d (debe ser 0) conB=%d (debe ser >0)", a, b)
	}
}

// TestReceptorNoBloqueaAlEmisor — con la cola saturada, `ServeHTTP` responde **200 con
// partialSuccess** en menos de 5 ms. Un 5xx haría reintentar al exportador y multiplicaría el
// daño; un bloqueo degradaría el trabajo que se está midiendo.
func TestReceptorNoBloqueaAlEmisor(t *testing.T) {
	sink := &sinkEspia{tope: 1} // se satura al segundo lote
	srv, r := servidorEfimero(t, sink)

	// Primer lote: entra.
	if resp := postear(t, srv.URL+"/v1/logs", "application/json",
		conMarca(t, "logs-run1.json", "MARCA-B-ENTRA")); resp.StatusCode != http.StatusOK {
		t.Fatalf("el primer lote debe entrar: %d", resp.StatusCode)
	}

	// Segundo lote: la cola está llena.
	inicio := time.Now()
	resp := postear(t, srv.URL+"/v1/logs", "application/json", conMarca(t, "logs-run1.json", "MARCA-A-DESCARTADO"))
	transcurrido := time.Since(inicio)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("la cola llena responde 200 con partialSuccess, no %d — un 5xx haría reintentar", resp.StatusCode)
	}
	var cuerpo struct {
		PartialSuccess *struct {
			RejectedLogRecords int64  `json:"rejectedLogRecords"`
			ErrorMessage       string `json:"errorMessage"`
		} `json:"partialSuccess"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
		t.Fatal(err)
	}
	if cuerpo.PartialSuccess == nil || cuerpo.PartialSuccess.RejectedLogRecords == 0 {
		t.Fatal("el descarte tiene que viajar en partialSuccess: el emisor se entera o la pérdida es silenciosa")
	}
	if cuerpo.PartialSuccess.ErrorMessage == "" {
		t.Error("el descarte se explica")
	}
	// El presupuesto es p99 ≤ 5 ms para decode+encolar; el round-trip HTTP local suma algo,
	// así que se verifica un techo generoso — lo que se está probando es que NO BLOQUEA.
	if transcurrido > 250*time.Millisecond {
		t.Errorf("el receptor bloqueó al emisor: %v", transcurrido)
	}
	// El descarte se CONTÓ.
	if r.Salud().ColaLlena.Load() == 0 {
		t.Error("el descarte por cola llena tiene que contarse en la salud")
	}
	// Y el assert canónico.
	if a, b := sink.conMarcador("MARCA-A-DESCARTADO"), sink.conMarcador("MARCA-B-ENTRA"); a != 0 || b == 0 {
		t.Fatalf("conA=%d (debe ser 0) conB=%d (debe ser >0)", a, b)
	}
}

// TestMetodoYRutaSeAcotan — cualquier cosa que no sea POST a las dos rutas se rechaza.
func TestMetodoYRutaSeAcotan(t *testing.T) {
	sink := &sinkEspia{}
	srv, _ := servidorEfimero(t, sink)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/v1/logs", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/logs debe dar 405, dio %d", resp.StatusCode)
	}
	if r2 := postear(t, srv.URL+"/v1/otra-cosa", "application/json", []byte(`{}`)); r2.StatusCode != http.StatusNotFound {
		t.Errorf("una ruta desconocida debe dar 404, dio %d", r2.StatusCode)
	}
	// Control positivo: la ruta buena sí responde.
	if r3 := postear(t, srv.URL+"/v1/logs", "application/json", []byte(`{}`)); r3.StatusCode != http.StatusOK {
		t.Errorf("POST /v1/logs debe dar 200, dio %d", r3.StatusCode)
	}
}

// TestMetricasEntranPorSuRuta — el canal secundario funciona y usa su propio contador de
// rechazos en `partialSuccess`.
func TestMetricasEntranPorSuRuta(t *testing.T) {
	sink := &sinkEspia{}
	srv, _ := servidorEfimero(t, sink)
	if resp := postear(t, srv.URL+"/v1/metrics", "application/json",
		conMarca(t, "metrics-run1.json", "MARCA-METRICAS")); resp.StatusCode != http.StatusOK {
		t.Fatalf("las métricas deben entrar por /v1/metrics: %d", resp.StatusCode)
	}
	if n := sink.conMarcador("MARCA-METRICAS"); n == 0 {
		t.Fatal("ningún punto de métrica llegó al sink")
	}
}

// BenchmarkReceptorLogs mide decode + mapeo + encolar sobre el payload REAL. Presupuesto
// §11: p99 ≤ 5 ms, **cero I/O de disco en el handler** (el disco lo hace el writer, asíncrono).
func BenchmarkReceptorLogs(b *testing.B) {
	cuerpo, err := os.ReadFile(filepath.Join("testdata", "logs-run1.json"))
	if err != nil {
		b.Fatal(err)
	}
	r := otlp.NewReceptor(&sinkEspia{}, otlp.Opciones{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("código %d", w.Code)
		}
	}
}
