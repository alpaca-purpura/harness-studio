package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthzReflectsCORSForAllowlistedOrigin es la prueba de regresión de HS-14: /healthz
// respondía 200 pero sin cabecera CORS, así que un fetch() del WebView (origin propio,
// distinto de 127.0.0.1:4200) lo veía como error de red pese a que el daemon respondía bien —
// indistinguible de "el daemon no arrancó" para conectando.html.
func TestHealthzReflectsCORSForAllowlistedOrigin(t *testing.T) {
	cfg := AuthConfigFor("127.0.0.1:4200", "sometoken")
	h := withAuth(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "tauri://localhost")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "tauri://localhost" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want the reflected WebView origin", got)
	}
}

// TestHealthzNoTokenOrHostRequired: /healthz sigue siendo liveness pura — ni token ni Host
// gate, incluso con uno configurado.
func TestHealthzNoTokenOrHostRequired(t *testing.T) {
	cfg := AuthConfigFor("127.0.0.1:4200", "sometoken")
	h := withAuth(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	req.Host = "attacker.example"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, /healthz must stay reachable without a token or matching Host", w.Code)
	}
}
