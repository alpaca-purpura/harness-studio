// delivery_test.go — fija las derivaciones del cockpit de delivery (F5 · DH-08 · SPEC
// delivery-cockpit.md): tramo desde el descriptor (cero literales de estado — RN-39),
// gates del tramo, contexto as-code por convención (D14) y honestidad del proxy sin
// sidecar (RN-40). Incluye la prueba "otro descriptor, cero cambio de motor" sobre el
// alterno de proceso_test.go.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSiguienteTramoRol(t *testing.T) {
	cases := []struct {
		estado string
		de, a  string
		ok     bool
	}{
		{"idea", "refining", "refined", true}, // inicial: solo salidas de operador → profundidad 2
		{"refining", "refining", "refined", true},
		{"developing", "developing", "developed", true},
		{"reviewing", "reviewing", "done", true},
		{"done", "", "", false},    // terminal
		{"dropped", "", "", false}, // terminal
		{"parked", "", "", false},  // pausado: su única salida (→idea) no alcanza un tramo de rol
	}
	for _, c := range cases {
		de, a, ok := siguienteTramoRol(proceso, c.estado)
		if ok != c.ok || de != c.de || a != c.a {
			t.Errorf("siguienteTramoRol(%q) = (%q, %q, %v), want (%q, %q, %v)",
				c.estado, de, a, ok, c.de, c.a, c.ok)
		}
	}
}

// otra empresa, otro descriptor, cero cambio del motor de delivery (tesis del corte).
func TestSiguienteTramoRolDescriptorAlterno(t *testing.T) {
	p, err := parseProceso([]byte(descriptorAlterno))
	if err != nil {
		t.Fatal(err)
	}
	if de, a, ok := siguienteTramoRol(p, "triage"); !ok || de != "build" || a != "staging" {
		t.Errorf("triage → (%q, %q, %v), want (build, staging, true)", de, a, ok)
	}
	if de, a, ok := siguienteTramoRol(p, "build"); !ok || de != "build" || a != "staging" {
		t.Errorf("build → (%q, %q, %v), want (build, staging, true)", de, a, ok)
	}
	if _, _, ok := siguienteTramoRol(p, "frozen"); ok {
		t.Error("frozen no debería alcanzar tramo de rol")
	}
	gates := gatesDelTramo(p, "build", "staging")
	if len(gates) != 1 || gates[0].ID != "go-live" {
		t.Errorf("gatesDelTramo(build→staging) = %v, want [go-live]", gates)
	}
}

func TestGatesDelTramo(t *testing.T) {
	cases := []struct {
		de, a string
		want  []string
	}{
		{"developing", "developed", []string{"verificacion-operador"}}, // momento = estado
		{"reviewing", "done", []string{"merge-gate"}},
		{"refining", "refined", nil},
		{"idea", "parked", []string{"razon-de-cierre"}}, // momento = transición de→a
	}
	for _, c := range cases {
		got := gatesDelTramo(proceso, c.de, c.a)
		var ids []string
		for _, g := range got {
			ids = append(ids, g.ID)
		}
		if len(ids) != len(c.want) {
			t.Errorf("gatesDelTramo(%s→%s) = %v, want %v", c.de, c.a, ids, c.want)
			continue
		}
		for i := range ids {
			if ids[i] != c.want[i] {
				t.Errorf("gatesDelTramo(%s→%s)[%d] = %s, want %s", c.de, c.a, i, ids[i], c.want[i])
			}
		}
	}
}

func TestContextosAsCode(t *testing.T) {
	dir := t.TempDir()
	escribir := func(rel, contenido string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	escribir("arquitectura.yaml", "meta: {}")
	escribir("celda/arquitectura.yaml", "meta: {}")
	escribir("celda/VISION.md", "# visión")
	escribir("otra/sin-arquitectura.txt", "x") // subdir sin candidato → no aparece

	ctxs := contextosAsCode(dir)
	if len(ctxs) != 2 {
		t.Fatalf("contextos = %d, want 2 (%v)", len(ctxs), ctxs)
	}
	if ctxs[0].ID != filepath.Base(dir) || len(ctxs[0].Fuentes) != 1 {
		t.Errorf("contexto raíz = %+v", ctxs[0])
	}
	if ctxs[1].ID != "celda" || len(ctxs[1].Fuentes) != 2 {
		t.Errorf("contexto celda = %+v (want arquitectura+VISION)", ctxs[1])
	}
}

// workspace de fixture con una story en `estado`, en modo multi (patrón de la suite).
func setupDeliveryWorkspace(t *testing.T, estado string) (sistemaKey string) {
	t.Helper()
	dir := mkBoard(t, "web")
	storyDir := filepath.Join(dir, "web", "docs", "product", "stories", "st-demo")
	if err := os.MkdirAll(storyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ckpt := "---\nstory_id: st-demo\nstate: " + estado + "\ngoal: Probar el cockpit de delivery end-to-end\n---\ncuerpo de la story\n"
	if err := os.WriteFile(filepath.Join(storyDir, "checkpoint.md"), []byte(ckpt), 0o644); err != nil {
		t.Fatal(err)
	}
	// as-code de la "célula" (D14): arquitectura + visión en el root del sistema
	if err := os.WriteFile(filepath.Join(dir, "web", "arquitectura.yaml"), []byte("meta: {clase: modelo}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "web", "VISION.md"), []byte("# Visión\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	enableMultiMode([]registryProject{{Name: "demo", Path: dir, Active: true}})
	t.Cleanup(func() { multiProjects = nil; systemWorkspaces = map[string]sysWorkspace{} })
	return "demo/web"
}

func TestPlantillaPara(t *testing.T) {
	sistema := setupDeliveryWorkspace(t, "idea")

	pl, status, motivo := plantillaPara(sistema, "st-demo", "")
	if status != 200 {
		t.Fatalf("plantilla: status %d (%s)", status, motivo)
	}
	// el tramo derivado del descriptor: idea (inicial) → refining→refined
	if pl.Tramo["de"] != "refining" || pl.Tramo["a"] != "refined" {
		t.Errorf("tramo = %v, want refining→refined", pl.Tramo)
	}
	// el prompt nombra el arnés dueño del destino (binding, no hardcode) y el mandato
	for _, frag := range []string{"/architect", "st-demo", "«refined»", "salida/", "NO transiciones"} {
		if !strings.Contains(pl.Prompt, frag) {
			t.Errorf("prompt no contiene %q\n---\n%s", frag, pl.Prompt)
		}
	}
	// fuentes: checkpoint SIEMPRE primero + as-code de la célula descubierta
	if len(pl.Fuentes) != 3 || pl.Fuentes[0].Nombre != "checkpoint.md" {
		t.Errorf("fuentes = %+v, want [checkpoint.md, arquitectura.yaml, VISION.md]", pl.Fuentes)
	}
	if pl.Contexto.ID != "web" || len(pl.ContextosDisponibles) != 1 {
		t.Errorf("contexto = %+v disponibles=%v", pl.Contexto, pl.ContextosDisponibles)
	}
}

func TestPlantillaParaNoLanzable(t *testing.T) {
	cases := map[string]string{
		"done":    "terminal",
		"dropped": "terminal",
		"parked":  "pausado",
	}
	for estado, fragmento := range cases {
		t.Run(estado, func(t *testing.T) {
			sistema := setupDeliveryWorkspace(t, estado)
			_, status, motivo := plantillaPara(sistema, "st-demo", "")
			if status != 422 || !strings.Contains(motivo, fragmento) {
				t.Errorf("%s: status=%d motivo=%q, want 422 con %q", estado, status, motivo, fragmento)
			}
		})
	}
}

func TestPlantillaParaErrores(t *testing.T) {
	sistema := setupDeliveryWorkspace(t, "idea")
	if _, status, _ := plantillaPara("no/existe", "st-demo", ""); status != 400 {
		t.Errorf("sistema desconocido: status=%d, want 400", status)
	}
	if _, status, _ := plantillaPara(sistema, "fantasma", ""); status != 404 {
		t.Errorf("story desconocida: status=%d, want 404", status)
	}
}

// RN-40: sin sidecar el proxy responde honesto — jamás datos inventados.
func TestDeliveryProxyHonestoSinSidecar(t *testing.T) {
	port, motivo := deliverySidecar.estado()
	defer deliverySidecar.set(port, motivo)
	deliverySidecar.set(0, "sidecar no arrancado")

	for _, tc := range []struct {
		nombre  string
		handler http.HandlerFunc
		req     *http.Request
	}{
		{"salud", handleDeliverySalud, httptest.NewRequest("GET", "/api/delivery/salud", nil)},
		{"sesiones", handleDeliverySesionesList, httptest.NewRequest("GET", "/api/delivery/sesiones", nil)},
	} {
		rec := httptest.NewRecorder()
		tc.handler(rec, tc.req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: status=%d, want 503", tc.nombre, rec.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: body ilegible: %v", tc.nombre, err)
		}
		if disponible, _ := body["disponible"].(bool); disponible {
			t.Errorf("%s: disponible=true con sidecar caído", tc.nombre)
		}
	}
}

func TestDeliverySesionCrearValida(t *testing.T) {
	sistema := setupDeliveryWorkspace(t, "idea")
	port, motivo := deliverySidecar.estado()
	defer deliverySidecar.set(port, motivo)
	deliverySidecar.set(0, "sidecar no arrancado")

	post := func(payload string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/delivery/sesiones", strings.NewReader(payload))
		handleDeliverySesionCrear(rec, req)
		return rec
	}
	if rec := post(`{"sistema":"` + sistema + `","story":"st-demo","prompt":""}`); rec.Code != 400 {
		t.Errorf("prompt vacío: status=%d, want 400", rec.Code)
	}
	if rec := post(`{"sistema":"` + sistema + `","story":"fantasma","prompt":"hola"}`); rec.Code != 404 {
		t.Errorf("story fantasma: status=%d, want 404", rec.Code)
	}
	// story válida + prompt válido pero sidecar caído → 503 honesto (la validación pasó)
	if rec := post(`{"sistema":"` + sistema + `","story":"st-demo","prompt":"hola mundo"}`); rec.Code != 503 {
		t.Errorf("sidecar caído: status=%d, want 503", rec.Code)
	}
}
