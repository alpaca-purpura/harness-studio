package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// llegaron registra qué rutas llegaron al handler de abajo. Es el instrumento del control
// positivo: sin él, «respondió 401» no distingue «lo frenó el gate» de «el handler no existe».
type espia struct{ rutas []string }

func (e *espia) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.rutas = append(e.rutas, r.URL.Path)
	w.WriteHeader(http.StatusOK)
}

func pedir(t *testing.T, h http.Handler, metodo, ruta, host, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), metodo, "http://"+host+ruta, nil)
	req.Host = host
	if token != "" {
		req.Header.Set("X-Arnesia-Token", token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// TestTokenDeIngestaNoAbreLaAPI — el corazón de A6. El token acotado da 200 en `/v1/logs` y
// **401 en `/api/sessions`**. Filtrarlo concede «podés escribirme telemetría», nunca «podés
// dirigir un agente con acceso al filesystem».
func TestTokenDeIngestaNoAbreLaAPI(t *testing.T) {
	e := &espia{}
	cfg := AuthConfigConIngesta("127.0.0.1:4200", "TOKEN-API", "TOKEN-INGESTA", false)
	h := withAuth(cfg, e)
	const host = "127.0.0.1:4200"

	// Con el token de ingesta: entra a la ingesta…
	if w := pedir(t, h, http.MethodPost, "/api/telemetria/proceso", host, "TOKEN-INGESTA"); w.Code != http.StatusOK {
		t.Errorf("el token de ingesta debe abrir /api/telemetria/proceso: %d", w.Code)
	}
	// …y NO a la API.
	if w := pedir(t, h, http.MethodGet, "/api/sessions", host, "TOKEN-INGESTA"); w.Code != http.StatusUnauthorized {
		t.Fatalf("el token de ingesta NO puede abrir /api/sessions: %d", w.Code)
	}
	// ── control positivo: el token de la API sí abre /api/sessions ──
	if w := pedir(t, h, http.MethodGet, "/api/sessions", host, "TOKEN-API"); w.Code != http.StatusOK {
		t.Fatalf("control positivo: el token de la API debe abrir /api/sessions: %d", w.Code)
	}
	// Y el handler de abajo NUNCA vio la request frenada.
	for _, r := range e.rutas {
		_ = r
	}
	vistas := strings.Join(e.rutas, ",")
	if strings.Count(vistas, "/api/sessions") != 1 {
		t.Errorf("solo UNA de las dos peticiones a /api/sessions debió pasar el gate: %q", vistas)
	}
}

// TestOTLPAceptaSinTokenBajoLoopback — A22. Sin escotilla, `/v1/*` acepta sin token bajo el
// Host gate loopback, porque el runtime no puede recibirlo por indirección y ponerlo literal
// sería publicar un secreto.
func TestOTLPAceptaSinTokenBajoLoopback(t *testing.T) {
	e := &espia{}
	h := withAuth(AuthConfigConIngesta("127.0.0.1:4200", "TOKEN-API", "TOKEN-INGESTA", false), e)
	if w := pedir(t, h, http.MethodPost, "/v1/logs", "127.0.0.1:4200", ""); w.Code != http.StatusOK {
		t.Fatalf("/v1/logs sin token debe pasar bajo loopback: %d", w.Code)
	}
	// ── control positivo del contraste: /api/telemetria/proceso SÍ exige token siempre ──
	if w := pedir(t, h, http.MethodPost, "/api/telemetria/proceso", "127.0.0.1:4200", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("el endpoint del hook exige token SIEMPRE: %d", w.Code)
	}
}

// TestOTLPBajoLosTresGates — la excepción de token no relaja el Host gate: un Host ajeno da
// 403 aunque la ruta sea de ingesta. Es la barrera anti DNS-rebinding, y es la que queda.
func TestOTLPBajoLosTresGates(t *testing.T) {
	e := &espia{}
	h := withAuth(AuthConfigConIngesta("127.0.0.1:4200", "TOKEN-API", "TOKEN-INGESTA", false), e)

	// Host ajeno ⇒ 403, con marcador A.
	if w := pedir(t, h, http.MethodPost, "/v1/logs", "malicioso.example.com", "TOKEN-INGESTA"); w.Code != http.StatusForbidden {
		t.Fatalf("un Host ajeno debe dar 403 aunque la ruta sea de ingesta: %d", w.Code)
	}
	// ── control positivo: el mismo POST con Host loopback entra ──
	if w := pedir(t, h, http.MethodPost, "/v1/logs", "127.0.0.1:4200", ""); w.Code != http.StatusOK {
		t.Fatalf("control positivo: con Host loopback debe entrar: %d", w.Code)
	}
	if len(e.rutas) != 1 {
		t.Errorf("solo la petición con Host loopback debió llegar al handler: %v", e.rutas)
	}
}

// TestModoEstrictoApagaS2Instrumentado — la escotilla del operador, con su consecuencia
// honesta: encendida, `/v1/logs` sin token da 401 y `s2-instrumentado` deja de reportar.
func TestModoEstrictoApagaS2Instrumentado(t *testing.T) {
	e := &espia{}
	h := withAuth(AuthConfigConIngesta("127.0.0.1:4200", "TOKEN-API", "TOKEN-INGESTA", true), e)
	const host = "127.0.0.1:4200"

	if w := pedir(t, h, http.MethodPost, "/v1/logs", host, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("con la escotilla encendida, /v1/logs sin token da 401: %d", w.Code)
	}
	// ── control positivo: CON token sigue entrando (S1, que sí puede llevarlo) ──
	if w := pedir(t, h, http.MethodPost, "/v1/logs", host, "TOKEN-INGESTA"); w.Code != http.StatusOK {
		t.Fatalf("control positivo: con token debe entrar aun en modo estricto: %d", w.Code)
	}
	if len(e.rutas) != 1 {
		t.Errorf("solo la petición con token debió llegar: %v", e.rutas)
	}
}

// TestLasRutasDeIngestaSonTres — la lista es cerrada. Agregar una cuarta exige tocar el
// archivo, que es el punto.
func TestLasRutasDeIngestaSonTres(t *testing.T) {
	esperadas := []string{"/v1/logs", "/v1/metrics", "/api/telemetria/proceso"}
	for _, r := range esperadas {
		if !isRutaIngesta(r) {
			t.Errorf("%s debe ser ruta de ingesta", r)
		}
	}
	if len(rutasIngesta) != len(esperadas) {
		t.Errorf("las rutas de ingesta son %d, hay %d — ampliar la lista es una decisión, no un descuido",
			len(esperadas), len(rutasIngesta))
	}
	// Control positivo del contraste: las rutas de la API NO son de ingesta.
	for _, r := range []string{"/api/sessions", "/api/harnesses", "/events", "/api/telemetria/resumen"} {
		if isRutaIngesta(r) {
			t.Errorf("%s NO puede ser ruta de ingesta: el token acotado la abriría", r)
		}
	}
}

// TestV1QuedaDentroDelConfinamiento — `/v1/` cuenta como superficie con capability: si no
// entrara a `isAPIPath`, quedaría fuera del gate y cualquier proceso local podría envenenar
// el almacén sin pasar por ningún control.
func TestV1QuedaDentroDelConfinamiento(t *testing.T) {
	if !isAPIPath("/v1/logs") || !isAPIPath("/v1/metrics") {
		t.Fatal("/v1/* tiene que estar dentro del confinamiento")
	}
	// Control positivo: la SPA estática NO lo está — un browser debe poder cargarla antes
	// de tener token.
	if isAPIPath("/") || isAPIPath("/assets/app.js") {
		t.Error("la UI estática queda fuera del token gate a propósito")
	}
}
