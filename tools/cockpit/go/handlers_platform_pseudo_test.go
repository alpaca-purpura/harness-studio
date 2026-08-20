package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Bug latente platform-pseudo (anclado en I-40, fuente única del selector): los
// handlers de LECTURA validaban el `sistema`
// con getSistemas() (omite el pseudo `platform`) en vez de getSelectableSistemas()
// (lo incluye) → vitalia/nicolify/comunify (board en la raíz = platform) daban
// "sistema desconocido: <empresa>/platform" en vez del empty-state honesto.
// Espejo del caso platform de TestHandleLedger. Los handlers de ESCRITURA (PATCH/POST)
// SIGUEN rechazando platform (getSistemas) — no se edita/crea contra el pseudo-sistema.
func setupPlatformEmpresa(t *testing.T) string {
	t.Helper()
	// empresa con board en la RAÍZ (docs/product) = pseudo-sistema `platform`.
	platDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(platDir, "docs", "product"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := registryProject{Name: "vitalia", Path: platDir, Active: true, Directorio: directorioMeta{Kind: "own"}}
	enableMultiMode([]registryProject{p})
	enableSystemWorkspaces([]registryProject{p})
	t.Cleanup(func() { multiProjects = nil; systemWorkspaces = map[string]sysWorkspace{} })
	return platDir
}

const platformKey = "vitalia/platform"

// notUnknownSistema falla si el handler de lectura rechazó platform como desconocido.
func notUnknownSistema(t *testing.T, rec *httptest.ResponseRecorder, name string) {
	t.Helper()
	if rec.Code == 400 && strings.Contains(rec.Body.String(), "sistema desconocido") {
		t.Fatalf("%s: rechazó platform como desconocido (bug platform-pseudo): %s", name, rec.Body.String())
	}
}

func TestReadHandlersAcceptPlatformPseudo(t *testing.T) {
	platDir := setupPlatformEmpresa(t)

	t.Run("system-map", func(t *testing.T) {
		// SYSTEM-MAP.yaml debe existir para no 500ear por archivo faltante (eso es
		// comportamiento pre-existente, no del bug). Con archivo → 200.
		mapDir := filepath.Join(platDir, "docs", "architecture")
		if err := os.MkdirAll(mapDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(mapDir, "SYSTEM-MAP.yaml"), []byte("zones: []\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		handleSystemMap(rec, httptest.NewRequest(http.MethodGet, "/api/system-map?sistema="+platformKey, nil))
		notUnknownSistema(t, rec, "system-map")
		if rec.Code != 200 {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("value-stream", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleValueStream(rec, httptest.NewRequest(http.MethodGet, "/api/value-stream?sistema="+platformKey, nil))
		notUnknownSistema(t, rec, "value-stream")
		if rec.Code != 200 {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("capabilities", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleCapabilitiesList(rec, httptest.NewRequest(http.MethodGet, "/api/capabilities?sistema="+platformKey, nil))
		notUnknownSistema(t, rec, "capabilities")
		if rec.Code != 200 {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("capabilities/status (capArtifactHandler)", func(t *testing.T) {
		h := capArtifactHandler("_status-computed.json", "status", "")
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodGet, "/api/capabilities/status?sistema="+platformKey, nil))
		notUnknownSistema(t, rec, "capabilities/status")
		if rec.Code != 200 {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("capabilities/doctor", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleCapDoctor(rec, httptest.NewRequest(http.MethodGet, "/api/capabilities/doctor?sistema="+platformKey, nil))
		notUnknownSistema(t, rec, "capabilities/doctor")
		if rec.Code != 200 {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("capabilities/single GET acepta platform (404, no 400 desconocido)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/capabilities/m/c?sistema="+platformKey, nil)
		req.SetPathValue("module", "m")
		req.SetPathValue("cap", "c")
		rec := httptest.NewRecorder()
		handleCapabilitySingle(rec, req)
		notUnknownSistema(t, rec, "capabilities/single GET")
		// platform no tiene caps individuales → 404 (no 400 desconocido).
		if rec.Code != 404 {
			t.Fatalf("status = %d, want 404 (capability no encontrada) (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

// El lado ESCRITURA del handler mixto SIGUE rechazando platform (no se edita un cap
// del pseudo-sistema): PATCH platform → 400 "sistema desconocido".
func TestCapabilitySinglePatchRejectsPlatform(t *testing.T) {
	setupPlatformEmpresa(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/capabilities/m/c?sistema="+platformKey,
		strings.NewReader(`{"status":"deprecated","reason":"x"}`))
	req.SetPathValue("module", "m")
	req.SetPathValue("cap", "c")
	rec := httptest.NewRecorder()
	handleCapabilitySingle(rec, req)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "sistema desconocido") {
		t.Fatalf("PATCH platform debe rechazarse 400 sistema desconocido, got %d: %s", rec.Code, rec.Body.String())
	}
}
