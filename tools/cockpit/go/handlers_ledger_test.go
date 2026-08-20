package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Fase 4 · vista "Evolución" (I-45): /api/ledger lee docs/product/ledger.yaml del
// sistema resuelto (cross-repo vía workspace externo). Con archivo → ledger poblado;
// sin archivo → ledger:null 200 (empty-state honesto, no rompe sistemas SDD).
func TestHandleLedger(t *testing.T) {
	// (1) workspace externo estilo fábrica: board en products/ + ledger.yaml generado.
	ext := mkBoard(t, "products")
	ledgerYAML := "sistema: prenter-harness\nversion: v0.4.0\nfichas:\n- id: I-45\n  titulo: Vista Evolución\n  estado: decidida\n"
	if err := os.WriteFile(filepath.Join(ext, "products", "docs", "product", "ledger.yaml"), []byte(ledgerYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	pFactory := registryProject{
		Name: "prenter", Path: t.TempDir(), Active: false,
		Directorio: directorioMeta{
			Kind:     "factory",
			Sistemas: []sistemaMeta{{Slug: "prenter-harness", Procedencia: "propio", Workspace: ext}},
		},
	}
	// (2) sistema con board pero SIN ledger.yaml (caso SDD).
	bare := mkBoard(t, "product")
	pSDD := registryProject{
		Name: "perusaas", Path: t.TempDir(), Active: false,
		Directorio: directorioMeta{
			Kind:     "own",
			Sistemas: []sistemaMeta{{Slug: "product", Procedencia: "propio", Workspace: bare}},
		},
	}
	// (3) empresa con board en RAÍZ = pseudo-sistema `platform` (caso vitalia/nicolify/comunify).
	// Regresión: validar con getSelectableSistemas (incluye platform), NO getSistemas (lo omite)
	// → debe dar ledger:null, NO "sistema desconocido".
	platDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(platDir, "docs", "product"), 0o755); err != nil {
		t.Fatal(err)
	}
	pPlatform := registryProject{Name: "vitalia", Path: platDir, Active: true, Directorio: directorioMeta{Kind: "own"}}
	enableMultiMode([]registryProject{pFactory, pSDD, pPlatform})
	enableSystemWorkspaces([]registryProject{pFactory, pSDD, pPlatform})
	defer func() { multiProjects = nil; systemWorkspaces = map[string]sysWorkspace{} }()

	// con ledger.yaml → ledger no-null con sus campos
	rec := httptest.NewRecorder()
	handleLedger(rec, httptest.NewRequest(http.MethodGet, "/api/ledger?sistema=prenter/prenter-harness", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Ledger map[string]any `json:"ledger"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Ledger == nil {
		t.Fatalf("ledger no debe ser null con ledger.yaml presente: %s", rec.Body.String())
	}
	if resp.Ledger["version"] != "v0.4.0" {
		t.Errorf("version mal proyectada: %v", resp.Ledger["version"])
	}
	if fichas, ok := resp.Ledger["fichas"].([]any); !ok || len(fichas) != 1 {
		t.Fatalf("fichas mal proyectadas: %v", resp.Ledger["fichas"])
	}

	// sin ledger.yaml → ledger:null, 200 (no rompe los sistemas que corren SDD)
	rec2 := httptest.NewRecorder()
	handleLedger(rec2, httptest.NewRequest(http.MethodGet, "/api/ledger?sistema=perusaas/product", nil))
	if rec2.Code != 200 {
		t.Fatalf("status = %d, want 200 (body=%s)", rec2.Code, rec2.Body.String())
	}
	var resp2 struct {
		Ledger any `json:"ledger"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp2); err != nil {
		t.Fatal(err)
	}
	if resp2.Ledger != nil {
		t.Errorf("sin ledger.yaml, ledger debe ser null: %s", rec2.Body.String())
	}

	// platform-pseudo (vitalia/platform) → 200 ledger:null, NO "sistema desconocido" (el bug reportado)
	rec3 := httptest.NewRecorder()
	handleLedger(rec3, httptest.NewRequest(http.MethodGet, "/api/ledger?sistema=vitalia/platform", nil))
	if rec3.Code != 200 {
		t.Fatalf("platform-pseudo: status = %d, want 200 (body=%s)", rec3.Code, rec3.Body.String())
	}
	var resp3 struct {
		Ledger any `json:"ledger"`
	}
	if err := json.Unmarshal(rec3.Body.Bytes(), &resp3); err != nil {
		t.Fatal(err)
	}
	if resp3.Ledger != nil {
		t.Errorf("platform-pseudo sin ledger.yaml debe ser null: %s", rec3.Body.String())
	}
}
