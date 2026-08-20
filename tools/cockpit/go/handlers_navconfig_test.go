package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Per-board nav (I-51): /api/nav-config?sistema=<key> lee el cockpit.config.yaml de
// la raíz del repo del board RESUELTO (resolveSistema, incl. cross-repo vía
// systemWorkspaces). Cada board cura su nav; board sin config → default (nav:null);
// key que no resuelve → default (200, NO error: el sidebar debe renderizar igual).
func TestHandleNavConfigPerBoard(t *testing.T) {
	// (1) board estilo fábrica con workspace externo + cockpit.config.yaml propio.
	extA := mkBoard(t, "products")
	cfg := "nav:\n  - board\n  - evolucion\n"
	if err := os.WriteFile(filepath.Join(extA, "cockpit.config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	pFactory := registryProject{
		Name: "prenter", Path: t.TempDir(), Active: false,
		Directorio: directorioMeta{
			Kind:     "factory",
			Sistemas: []sistemaMeta{{Slug: "prenter-harness", Procedencia: "propio", Workspace: extA}},
		},
	}
	// (2) board platform-pseudo (vitalia) en su Path, SIN cockpit.config.yaml.
	platDir := t.TempDir()
	pPlatform := registryProject{Name: "vitalia", Path: platDir, Active: true, Directorio: directorioMeta{Kind: "own"}}

	enableMultiMode([]registryProject{pFactory, pPlatform})
	enableSystemWorkspaces([]registryProject{pFactory, pPlatform})
	defer func() { multiProjects = nil; systemWorkspaces = map[string]sysWorkspace{} }()

	type navResp struct {
		Nav    []string `json:"nav"`
		Exists bool     `json:"exists"`
	}
	get := func(sistema string) (int, navResp) {
		rec := httptest.NewRecorder()
		url := "/api/nav-config"
		if sistema != "" {
			url += "?sistema=" + sistema
		}
		handleNavConfig(rec, httptest.NewRequest(http.MethodGet, url, nil))
		var resp navResp
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		return rec.Code, resp
	}

	// board CON config → su nav curado, en orden (cross-repo vía workspace externo).
	code, resp := get("prenter/prenter-harness")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	if !resp.Exists {
		t.Errorf("exists debe ser true con cockpit.config.yaml presente")
	}
	if len(resp.Nav) != 2 || resp.Nav[0] != "board" || resp.Nav[1] != "evolucion" {
		t.Errorf("nav per-board mal leído: %v", resp.Nav)
	}

	// board SIN config → default (nav:null, exists:false), 200.
	code, resp = get("vitalia/platform")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	if resp.Exists || resp.Nav != nil {
		t.Errorf("board sin config debe dar default (nav:null, exists:false): nav=%v exists=%v", resp.Nav, resp.Exists)
	}

	// key que NO resuelve → default 200 (NO error: el sidebar debe renderizar igual).
	code, resp = get("fantasma/nope")
	if code != 200 {
		t.Fatalf("key desconocida: status = %d, want 200 (no error)", code)
	}
	if resp.Exists || resp.Nav != nil {
		t.Errorf("key desconocida debe dar default: nav=%v exists=%v", resp.Nav, resp.Exists)
	}
}
